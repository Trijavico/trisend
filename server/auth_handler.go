package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"trisend/auth"
	"trisend/config"
	"trisend/db"
	"trisend/mailer"
	"trisend/util"
	"trisend/views"

	"github.com/markbates/goth/gothic"
)

const (
	Session_key = "sess"
	Auth_key    = "auth"
)

func handleOAuth(w http.ResponseWriter, r *http.Request) {
	switch r.PathValue("action") {
	case "login":
		gothic.BeginAuthHandler(w, r)

	case "callback":
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			return
		}
		claims := map[string]interface{}{
			"email":    user.Email,
			"username": user.NickName,
		}

		token, err := auth.CreateToken(claims)
		if err != nil {
			http.Error(w, "an error ocurred", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     Session_key,
			Value:    token,
			Path:     "/",
			Secure:   config.IsAppEnvProd(),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(time.Hour * 5),
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

		http.SetCookie(w, &http.Cookie{
			Name:     Auth_key,
			Value:    token,
			Path:     "/",
			Secure:   config.IsAppEnvProd(),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(time.Minute * 5),
		})

		body := fmt.Sprintf("CODE: %s", code)
		emailer := mailer.NewMailer("Verfication code", email, body)

		if err := emailer.Send(); err != nil {
			fmt.Println(err)
			http.Error(w, "failed to sent mail message", http.StatusInternalServerError)
			return
		}

		views.ContinueWithCode().Render(r.Context(), w)
	}
}

func handleVerification(store db.SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(Auth_key)
		if err != nil {
			slog.Error("Failed to sent message", ":", err)
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

		var code string
		if code = r.FormValue("code"); code == "" {
			slog.Error("no code provided")
			http.Error(w, "no code provided", http.StatusInternalServerError)
			return
		}

		value, err := store.GetTransitSessByID(r.Context(), key)
		if err != nil {
			slog.Error("failed to sent message", ":", err)
			http.Error(w, "failed to sent message", http.StatusInternalServerError)
			return
		}

		if code != value {
			slog.Error("invalid code")
			http.Error(w, "invalid code", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
	}
}
