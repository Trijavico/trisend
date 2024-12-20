package middleware

import (
	"context"
	"net/http"
	"trisend/auth"
	"trisend/server"
	"trisend/types"
)

func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(server.Session_key)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusInternalServerError)
			return
		}

		claims, err := auth.ParseToken(cookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var session types.User
		session.ID = claims["id"].(string)
		session.Email = claims["email"].(string)
		session.Username = claims["username"].(string)
		session.Pfp = claims["pfp"].(string)

		ctx := r.Context()
		ctxWithUser := context.WithValue(ctx, "session", session)
		r.WithContext(ctxWithUser)

		next(w, r)
	}
}
