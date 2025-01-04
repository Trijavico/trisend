package middleware

import (
	"context"
	"net/http"
	"trisend/auth"
	"trisend/types"
)

func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("sess")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusInternalServerError)
			return
		}

		claims, err := auth.ParseToken(cookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session := &types.Session{
			ID:       claims["id"].(string),
			Email:    claims["email"].(string),
			Username: claims["username"].(string),
			Pfp:      claims["pfp"].(string),
		}

		ctx := r.Context()
		ctxWithUser := context.WithValue(ctx, "sess", session)
		r = r.WithContext(ctxWithUser)

		next(w, r)
	}
}
