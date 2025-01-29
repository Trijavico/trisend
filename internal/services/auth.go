package services

import (
	"context"
	"log/slog"
	"net/http"
	"trisend/internal/config"
	"trisend/internal/db"
	"trisend/internal/types"
	"trisend/internal/util"

	"github.com/markbates/goth"
)

const (
	SESSION_COOKIE   = "sess"
	session_duration = 5
)

type AuthService struct {
	userStore db.UserStore
}

func NewAuthService(userStore db.UserStore) AuthService {
	return AuthService{
		userStore: userStore,
	}
}

func (s *AuthService) Login(w http.ResponseWriter, user types.Session) {
	token, err := util.CreateAccessToken(user, session_duration)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SESSION_COOKIE,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   config.IsAppEnvProd(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   session_duration * 3600,
	})
}

func (s *AuthService) Logout(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SESSION_COOKIE,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   config.IsAppEnvProd(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (s *AuthService) Register(w http.ResponseWriter, createUser types.CreateUser) error {
	user, err := s.userStore.CreateUser(context.Background(), createUser)
	if err != nil {
		return err
	}

	s.Login(w, *user)

	return nil
}

func (s *AuthService) OAuthAuthenticate(w http.ResponseWriter, gothUser goth.User) error {
	user, err := s.userStore.FindByEmail(context.Background(), gothUser.Email)
	if err != nil {
		return err
	}
	if user != nil {
		s.Login(w, *user)
		return nil
	}

	createUser := types.CreateUser{
		Email:    gothUser.Email,
		Username: gothUser.NickName,
		Pfp:      gothUser.AvatarURL,
	}

	s.Register(w, createUser)

	return nil
}
