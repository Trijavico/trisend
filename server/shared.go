package server

import (
	"log/slog"
	"net/http"
	"trisend/auth"
	"trisend/types"
)

func handleError(w http.ResponseWriter, msg string, err error, code int) {
	slog.Error(err.Error())
	http.Error(w, msg, code)
}

func getUserFromCookie(r *http.Request) *types.Session {
	cookie, err := r.Cookie(SESSION_COOKIE)
	if err != nil {
		return nil
	}

	claims, err := auth.ParseToken(cookie.Value)
	if err != nil {
		slog.Error("Failed to parse jwt token", "error", err)
		return nil
	}

	return &types.Session{
		ID:       claims["id"].(string),
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
		Pfp:      claims["pfp"].(string),
	}
}
