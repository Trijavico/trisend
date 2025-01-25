package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"trisend/auth"
	"trisend/config"
	"trisend/db"
	"trisend/server"

	"github.com/joho/godotenv"
	gossh "golang.org/x/crypto/ssh"
)

//go:embed banner.txt
var banner string

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	config.LoadConfig()
	auth.SetupOAuth()

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
	sessStore := db.NewRedisSessionStore(redisDB)

	webserver := server.NewWebServer()
	sshserver := server.NewSSHServer(privateKey, banner, userStore)

	webserver.AddRoutes(userStore, sessStore)

	go func() {
		slog.Info("SSH Server running...")
		if err := sshserver.ListenAndServe(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()

	slog.Info(fmt.Sprintf("HTTP Server running on PORT %s...", config.SERVER_PORT))
	if err := webserver.ListenAndServe(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
