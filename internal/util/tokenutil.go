package util

import (
	"fmt"
	"time"
	"trisend/internal/config"
	"trisend/internal/types"

	"github.com/golang-jwt/jwt/v5"
)

var invalidToken = fmt.Errorf("invalid token")

func CreateAccessToken(user types.Session, expiry int) (string, error) {
	exp := time.Now().Add(time.Hour * time.Duration(expiry))
	claims := &types.JwtSessClaims{
		Session: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	return createToken(claims)
}

func CreateSignupToken(id string, email string, expiry int) (string, error) {
	exp := time.Now().Add(time.Minute * time.Duration(expiry))
	claims := &types.JwtSignupClaims{
		ID:    id,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	return createToken(claims)
}

func createToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWT_SECRET))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWT_SECRET), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, invalidToken
	}

	return token, nil
}
