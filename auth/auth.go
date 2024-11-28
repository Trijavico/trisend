package auth

import (
	"net/http"
	"trisend/util"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

func SetupOAuth() {
	sessionKey := []byte(util.GetEnvStr("SESSION_SECRET", ""))
	cookieStore := sessions.NewCookieStore(sessionKey)
	gothic.Store = cookieStore

	clientSecret := util.GetEnvStr("CLIENT_SECRET", "")
	clientID := util.GetEnvStr("CLIENT_ID", "")
	goth.UseProviders(github.New(clientID, clientSecret, ""))
}

func CreateSessionCookie(w http.ResponseWriter, claims jwt.MapClaims) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := util.GetEnvStr("JWT_SECRET", "")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     "auth",
		Value:    tokenString,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	return nil
}
