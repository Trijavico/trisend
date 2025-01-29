package main

import (
	"context"
	"net/http"
	"trisend/internal/types"
	"trisend/internal/util"

	"github.com/golang-jwt/jwt/v5"
)

func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("sess")
		if err != nil {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		token, err := util.ParseToken(cookie.Value)
		if err != nil {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user := &types.Session{
			ID:       claims["id"].(string),
			Email:    claims["email"].(string),
			Username: claims["username"].(string),
			Pfp:      claims["pfp"].(string),
		}

		ctx := r.Context()
		ctxWithUser := context.WithValue(ctx, "sess", user)
		r = r.WithContext(ctxWithUser)

		next(w, r)
	}
}
