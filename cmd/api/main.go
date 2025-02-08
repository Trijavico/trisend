package main

import (
	_ "embed"
	"log/slog"
	"os"
	"trisend/internal/config"
	"trisend/internal/db"
	"trisend/internal/server"
	"trisend/internal/services"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	gossh "golang.org/x/crypto/ssh"
)

func SetupOAuth() {
	cookieStore := sessions.NewCookieStore([]byte(config.SESSION_SECRET))
	cookieStore.Options.HttpOnly = true
	gothic.Store = cookieStore

	goth.UseProviders(github.New(config.CLIENT_ID, config.CLIENT_SECRET, "", "user:email"))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error(err.Error())
	}
	config.LoadConfig()
	SetupOAuth()

	keyBytes, err := os.ReadFile("./keys/host")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	privateKey, err := gossh.ParsePrivateKey(keyBytes)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	redisDB, err := db.NewRedisDB()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	userStore := db.NewUserRedisStore(redisDB)
	app := App{
		Auth:         services.NewAuthService(userStore),
		UserStore:    userStore,
		SessionStore: db.NewRedisSessionStore(redisDB),
	}

	server := server.NewWebServer()
	router := AddRoutes(app)
	server.SetupConfig(router, privateKey, userStore)

	server.ListenAndServe()
}
