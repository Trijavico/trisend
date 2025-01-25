package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"trisend/auth"
	"trisend/db"
	"trisend/mailer"
	"trisend/types"
	"trisend/util"
	"trisend/views"

	"github.com/markbates/goth/gothic"
	"github.com/redis/go-redis/v9"
)

const (
	SESSION_COOKIE = "sess"
	AUTH_COOKIE    = "auth"

	auth_code_error      = "Unable to sent authentication code"
	verify_auth_error    = "Unable to verify authentication code"
	create_account_error = "Unable to create account"
)

func handleOAuth(store db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("action") {
		case "login":
			gothic.BeginAuthHandler(w, r)

		case "callback":
			gothUser, err := gothic.CompleteUserAuth(w, r)
			if err != nil {
				return
			}

			user, err := store.GetByEmail(r.Context(), gothUser.Email)
			if err != nil && err != redis.Nil {
				slog.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if user == nil {
				createUser := types.CreateUser{
					Email:    gothUser.Email,
					Username: gothUser.NickName,
					Pfp:      gothUser.AvatarURL,
				}

				user, err = store.CreateUser(r.Context(), createUser)
				if err != nil {
					slog.Error(err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			claims := map[string]interface{}{
				"id":       user.ID,
				"email":    user.Email,
				"username": user.Username,
				"pfp":      user.Pfp,
			}

			token, err := auth.CreateToken(claims)
			if err != nil {
				slog.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, auth.CreateCookie(SESSION_COOKIE, token, int(time.Hour*5)))
			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		}
	}
}

func handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth.DeleteCookie(w, SESSION_COOKIE)
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func handleAuthCode(store db.SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var email string
		if email = r.FormValue("email"); email == "" {
			views.EmailForm(email, fmt.Errorf("Invalid email")).Render(r.Context(), w)
			return
		}
		if !mailer.IsValidEmail(email) {
			views.EmailForm(email, fmt.Errorf("Invalid email")).Render(r.Context(), w)
			return
		}

		sessionID := util.GetRandomID(10)
		code := util.GetRandomID(10)

		claims := map[string]interface{}{
			"sub":   sessionID,
			"email": email,
		}

		token, err := auth.CreateToken(claims)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, auth_code_error, http.StatusInternalServerError)
			return
		}

		err = store.CreateTransitSess(r.Context(), sessionID, code)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, auth_code_error, http.StatusInternalServerError)
			return
		}

		body := fmt.Sprintf("CODE: %s", code)
		emailer := mailer.NewMailer("Verfication code", email, body)

		if err := emailer.Send(); err != nil {
			slog.Error(err.Error())
			http.Error(w, auth_code_error, http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, auth.CreateCookie(AUTH_COOKIE, token, int(time.Minute*5)))
		w.WriteHeader(http.StatusAccepted)
		views.EmailForm(email, nil).Render(r.Context(), w)
	}
}

func handleVerification(sessStore db.SessionStore, usrStore db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AUTH_COOKIE)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, verify_auth_error, http.StatusInternalServerError)
			return
		}

		claims, err := auth.ParseToken(cookie.Value)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, verify_auth_error, http.StatusInternalServerError)
			return
		}

		key := claims["sub"].(string)
		email := claims["email"].(string)
		code := r.FormValue("code")

		value, err := sessStore.GetTransitSessByID(r.Context(), key)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, verify_auth_error, http.StatusInternalServerError)
			return
		}

		if code != value {
			views.AuthCodeForm(code, fmt.Errorf("Invalid code")).Render(r.Context(), w)
			return
		}

		user, err := usrStore.GetByEmail(r.Context(), email)
		if err == redis.Nil {
			w.Header().Set("HX-Redirect", "/login/create")
			return
		} else if err != nil {
			slog.Error(err.Error())
			http.Error(w, verify_auth_error, http.StatusInternalServerError)
			return
		}

		mappedUser := map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
			"pfp":      user.Pfp,
		}

		token, err := auth.CreateToken(mappedUser)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, verify_auth_error, http.StatusInternalServerError)
			return
		}

		auth.DeleteCookie(w, AUTH_COOKIE)
		http.SetCookie(w, auth.CreateCookie(SESSION_COOKIE, token, int(time.Hour*5)))

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func handleLoginCreate(usrStore db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AUTH_COOKIE)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}

		claims, err := auth.ParseToken(cookie.Value)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}

		limit := int64(5 << 20) // 5MB
		if err := r.ParseMultipartForm(limit); err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}
		defer r.MultipartForm.RemoveAll()

		email := claims["email"].(string)
		username := r.FormValue("username")
		if username == "" {
			views.CreateUserForm(username, fmt.Errorf("Invalid username")).Render(r.Context(), w)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}

		path, err := os.Getwd()
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}

		path = filepath.Join(path, "media")
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}

		parsedFilename := strings.ReplaceAll(header.Filename, " ", "")
		filename := util.GetRandomID(12) + parsedFilename
		createdFile, err := os.Create(filepath.Join(path, filename))
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}
		defer createdFile.Close()

		_, err = io.Copy(createdFile, file)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}

		pfp := "/media/" + filename
		user := types.CreateUser{
			Email:    email,
			Username: username,
			Pfp:      pfp,
		}

		userSession, err := usrStore.CreateUser(r.Context(), user)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}

		token, err := auth.CreateToken(map[string]interface{}{
			"id":       userSession.ID,
			"email":    email,
			"username": username,
			"pfp":      pfp,
		})
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}

		auth.DeleteCookie(w, AUTH_COOKIE)
		http.SetCookie(w, auth.CreateCookie(SESSION_COOKIE, token, int(time.Hour*5)))
		w.Header().Set("HX-Redirect", "/")
	}
}
