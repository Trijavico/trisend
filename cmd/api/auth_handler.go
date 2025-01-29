package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"trisend/internal/config"
	"trisend/internal/mailer"
	"trisend/internal/types"
	"trisend/internal/util"
	"trisend/internal/views"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
)

const (
	SESSION_COOKIE = "sess"
	AUTH_COOKIE    = "auth"

	auth_code_error      = "Unable to sent authentication code"
	verify_auth_error    = "Unable to verify authentication code"
	create_account_error = "Unable to create account"
)

func handleLoginCreateView(w http.ResponseWriter, r *http.Request) {
	views.FillProfile().Render(r.Context(), w)
	return
}

func handleOAuth(app App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("action") {
		case "login":
			gothic.BeginAuthHandler(w, r)

		case "callback":
			gothUser, err := gothic.CompleteUserAuth(w, r)
			if err != nil {
				return
			}

			err = app.Auth.OAuthAuthenticate(w, gothUser)
			if err != nil {
				slog.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		}
	}
}

func handleLogout(app App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.Auth.Logout(w)
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func handleAuthCode(app App) http.HandlerFunc {
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

		sessionID := uuid.NewString()
		code := util.GetRandomID(8)

		err := app.SessionStore.CreateTransitSess(r.Context(), sessionID, code)
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

		http.SetCookie(w, &http.Cookie{
			Name:     AUTH_COOKIE,
			Value:    sessionID,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Secure:   config.IsAppEnvProd(),
			MaxAge:   60 * 2,
		})
		w.WriteHeader(http.StatusAccepted)
		views.EmailForm(email, nil).Render(r.Context(), w)
	}
}

func handleVerification(app App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		email := r.FormValue("email")

		cookie, err := r.Cookie(AUTH_COOKIE)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ID := cookie.Value

		value, err := app.SessionStore.GetTransitSessByID(r.Context(), ID)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, verify_auth_error, http.StatusInternalServerError)
			return
		}

		if code != value {
			views.AuthCodeForm(code, email, fmt.Errorf("Invalid code")).Render(r.Context(), w)
			return
		}

		user, err := app.UserStore.FindByEmail(r.Context(), email)
		if user == nil {
			token, err := util.CreateSignupToken(ID, email, 60)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     AUTH_COOKIE,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   config.IsAppEnvProd(),
				MaxAge:   60 * 60,
			})
			w.Header().Set("HX-Redirect", "/login/create")

			return
		} else if err != nil {
			slog.Error(err.Error())
			http.Error(w, verify_auth_error, http.StatusInternalServerError)
			return
		}

		app.Auth.Login(w, *user)
		http.SetCookie(w, &http.Cookie{
			Name:     AUTH_COOKIE,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   config.IsAppEnvProd(),
			MaxAge:   -1,
		})

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}

func handleLoginCreate(app App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AUTH_COOKIE)
		if err != nil {
			slog.Error(err.Error())
			w.Header().Set("HX-Redirect", "/login")
			return
		}

		token, err := util.ParseToken(cookie.Value)
		if err != nil {
			slog.Error(err.Error())
			w.Header().Set("HX-Redirect", "/login")
			return
		}
		claims := token.Claims.(jwt.MapClaims)

		limit := int64(5 << 20)
		if err := r.ParseMultipartForm(limit); err != nil {
			slog.Error(err.Error())
			http.Error(w, create_account_error, http.StatusInternalServerError)
			return
		}
		defer r.MultipartForm.RemoveAll()

		email := claims["email"].(string)
		username := r.FormValue("username")
		validationErrors := make([]error, 2)

		user, err := app.UserStore.FindByEmail(r.Context(), email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if user != nil {
			w.Header().Set("HX-Redirect", "/login")
			return
		}

		usernameValidator := regexp.MustCompile(`^(?:.*[a-zA-Z]){4,}`)
		if username == "" {
			validationErrors[0] = fmt.Errorf("Invalid username")
			views.CreateUserForm(username, validationErrors).Render(r.Context(), w)
			return
		} else if !usernameValidator.MatchString(username) {
			validationErrors[0] = fmt.Errorf("Username must contain at least 4 letters")
			views.CreateUserForm(username, validationErrors).Render(r.Context(), w)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			validationErrors[1] = fmt.Errorf("Provide an image file")
			views.CreateUserForm(username, validationErrors).Render(r.Context(), w)
			return
		}

		if header.Size == 0 {
			validationErrors[1] = fmt.Errorf("Uploaded file is empty")
			views.CreateUserForm(username, validationErrors).Render(r.Context(), w)
			return
		}

		_, _, err = image.DecodeConfig(file)
		if err != nil {
			validationErrors[1] = fmt.Errorf("Invalid image format")
			views.CreateUserForm(username, validationErrors).Render(r.Context(), w)
			return
		}
		file.Seek(0, 0)

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
		filename := uuid.NewString() + parsedFilename
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
		createUser := types.CreateUser{
			Email:    email,
			Username: username,
			Pfp:      pfp,
		}

		app.Auth.Register(w, createUser)
		http.SetCookie(w, &http.Cookie{
			Name:     AUTH_COOKIE,
			Value:    "",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   config.IsAppEnvProd(),
		})
		w.Header().Set("HX-Redirect", "/")
	}
}
