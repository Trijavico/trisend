package main

import (
	"net/http"
	"trisend/internal/types"
	"trisend/internal/util"

	"github.com/golang-jwt/jwt/v5"
)

func getUserFromCookie(r *http.Request) *types.Session {
	cookie, err := r.Cookie("sess")
	if err != nil {
		return nil
	}

	token, err := util.ParseToken(cookie.Value)
	if err != nil {
		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	return &types.Session{
		ID:       claims["id"].(string),
		Username: claims["username"].(string),
		Email:    claims["email"].(string),
		Pfp:      claims["pfp"].(string),
	}
}
