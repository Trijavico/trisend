package auth

import (
	"errors"
	"fmt"
	"net/http"
	"trisend/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

func SetupOAuth() {
	cookieStore := sessions.NewCookieStore([]byte(config.SESSION_SECRET))
	gothic.Store = cookieStore

	goth.UseProviders(github.New(config.CLIENT_ID, config.CLIENT_SECRET, ""))
}

func CreateCookie(name, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Secure:   config.IsAppEnvProd(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}
}

func CreateToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(config.JWT_SECRET))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWT_SECRET), nil
	})

	if err != nil {
		return nil, errors.New("unauthorized")
	}
	if !token.Valid {
		return nil, errors.New("unauthorized")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	return claims, nil
}
