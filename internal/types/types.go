package types

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Session struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Pfp      string `json:"pfp"`
}

func (sess *Session) ShortEmail() string {
	index := strings.Index(sess.Email, "@")
	return sess.Email[:index]
}

type CreateUser struct {
	Email    string
	Username string
	Pfp      string
}

type TransitSess struct {
	ID    string
	Email string
}

type SSHKey struct {
	ID          string
	Title       string
	Fingerprint string
}

type JwtSignupClaims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type JwtSessClaims struct {
	Session
	jwt.RegisteredClaims
}

type ValidationSSHForm struct {
	Fields map[string]string
	Errors map[string]string
}

func (form *ValidationSSHForm) Validate(title, key string) bool {
	form.Fields["title"] = title
	form.Fields["key"] = key
	isValid := true

	if title == "" {
		form.Errors["title"] = "Invalid title"
		isValid = false
	}
	if key == "" {
		form.Errors["key"] = "Invalid key"
		isValid = false
	}

	return isValid
}
