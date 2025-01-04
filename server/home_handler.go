package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"trisend/auth"
	"trisend/types"
	"trisend/views"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.String() != "/" {
		message := fmt.Sprintf("url: %s", r.URL)
		http.Error(w, message, http.StatusNotFound)
		return
	}

	cookie, err := r.Cookie(SESSION_COOKIE)
	if err != nil {
		views.Home(nil).Render(r.Context(), w)
		return
	}

	claims, err := auth.ParseToken(cookie.Value)
	if err != nil {
		slog.Error("Failed to parse jwt token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		views.Home(nil).Render(r.Context(), w)
		return
	}

	user := &types.Session{
		ID:       claims["id"].(string),
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
		Pfp:      claims["pfp"].(string),
	}

	views.Home(user).Render(r.Context(), w)
}
