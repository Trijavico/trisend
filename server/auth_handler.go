package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
)

func handleOAuth(store db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("action") {
		case "login":
			gothic.BeginAuthHandler(w, r)

		case "callback":
			user, err := gothic.CompleteUserAuth(w, r)
			if err != nil {
				return
			}

			userData, err := store.GetByEmail(r.Context(), user.Email)
			if err == redis.Nil || err != nil {
				slog.Error("An error occurred", "error", err)
				http.Error(w, "an error ocurred", http.StatusInternalServerError)
				return
			}

			claims := map[string]interface{}{
				"id":       userData.ID,
				"email":    userData.Email,
				"username": userData.Username,
				"pfp":      userData.Pfp,
			}

			token, err := auth.CreateToken(claims)
			if err != nil {
				http.Error(w, "an error ocurred", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, auth.CreateCookie(SESSION_COOKIE, token, int(time.Hour*5)))
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
	}
}

func handleAuthCode(store db.SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var email string
		if email = r.FormValue("email"); email == "" {
			http.Error(w, "no email provided", http.StatusInternalServerError)
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
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		err = store.CreateTransitSess(r.Context(), sessionID, code)
		if err != nil {
			http.Error(w, "failed to sent mail message", http.StatusInternalServerError)
			return
		}

		body := fmt.Sprintf("CODE: %s", code)
		emailer := mailer.NewMailer("Verfication code", email, body)

		if err := emailer.Send(); err != nil {
			fmt.Println(err)
			http.Error(w, "failed to sent mail message", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, auth.CreateCookie(AUTH_COOKIE, token, int(time.Minute*5)))
		views.ContinueWithCode().Render(r.Context(), w)
	}
}

func handleVerification(sessStore db.SessionStore, usrStore db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AUTH_COOKIE)
		if err != nil {
			slog.Error("Failed to sent message", "error", err)
			http.Error(w, "no code provided", http.StatusInternalServerError)
			return
		}

		claims, err := auth.ParseToken(cookie.Value)
		if err != nil {
			slog.Error("Failed to sent message", ":", err)
			http.Error(w, "no code provided", http.StatusInternalServerError)
			return
		}

		key := claims["sub"].(string)
		email := claims["email"].(string)
		code := r.FormValue("code")

		value, err := sessStore.GetTransitSessByID(r.Context(), key)
		if err != nil {
			slog.Error("failed to sent message", "error", err)
			http.Error(w, "failed to sent message", http.StatusInternalServerError)
			return
		}

		if code != value {
			slog.Error("invalid code")
			http.Error(w, "invalid code", http.StatusInternalServerError)
			return
		}

		user, err := usrStore.GetByEmail(r.Context(), email)
		if err == redis.Nil {
			w.Header().Set("HX-Redirect", "/login/create")
			return
		} else if err != nil {
			slog.Error("Failed to continue with login", "error", err)
			http.Error(w, "failed to continue with login", http.StatusInternalServerError)
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
			slog.Error("Failed to continue with login", "error", err)
			http.Error(w, "failed to continue with login", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, auth.CreateCookie(SESSION_COOKIE, token, int(time.Hour*5)))
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func handleLoginCreate(usrStore db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AUTH_COOKIE)
		if err != nil {
			slog.Error("Failed cookie", "error", err)
			http.Error(w, "an error occurred", http.StatusInternalServerError)
			return
		}

		claims, err := auth.ParseToken(cookie.Value)
		if err != nil {
			slog.Error("Failed parsing JWT", "error", err)
			http.Error(w, "an error occurred", http.StatusInternalServerError)
			return
		}

		limit := int64(5 << 20) // 5MB
		if err := r.ParseMultipartForm(limit); err != nil {
			slog.Error("Failed parsing Multipart Form", "error", err)
			http.Error(w, "an error occurred", http.StatusInternalServerError)
			return
		}
		defer r.MultipartForm.RemoveAll()

		email := claims["email"].(string)
		username := r.FormValue("username")

		file, header, err := r.FormFile("image")
		if err != nil {
			slog.Error("Failed parsing JWT", "error", err)
			http.Error(w, "an error occurred", http.StatusInternalServerError)
			return
		}

		path, err := os.Getwd()
		if err != nil {
			slog.Error("Faile creating dir", "error", err)
			http.Error(w, "an error occurred", http.StatusInternalServerError)
			return
		}

		path = filepath.Join(path, "media")

		var once sync.Once
		once.Do(func() {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				http.Error(w, "an error occurred", http.StatusInternalServerError)
				return
			}
		})

		parsedFilename := strings.ReplaceAll(header.Filename, " ", "")
		filename := util.GetRandomID(12) + parsedFilename
		createdFile, err := os.Create(filepath.Join(path, filename))
		if err != nil {
			slog.Error("Failed creating file", "error", err)
			http.Error(w, "an error occurred", http.StatusInternalServerError)
			return
		}
		defer createdFile.Close()

		_, err = io.Copy(createdFile, file)
		if err != nil {
			slog.Error("Failed saving file", "error", err)
			http.Error(w, "an error occurred", http.StatusInternalServerError)
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
			slog.Error("Failed parsing JWT", "error", err)
			http.Error(w, "an error occurred", http.StatusInternalServerError)
			return
		}

		token, err := auth.CreateToken(map[string]interface{}{
			"id":       userSession.ID,
			"email":    email,
			"username": username,
			"pfp":      pfp,
		})
		if err != nil {
			slog.Error("Failed parsing JWT", "error", err)
			http.Error(w, "an error occurred", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, auth.CreateCookie(SESSION_COOKIE, token, int(time.Hour*5)))
		w.Header().Set("HX-Redirect", "/")
	}
}
