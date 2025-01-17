package main

import (
	_ "embed"
	"log/slog"
	"os"
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

	slog.Info("HTTP Server running...")
	if err := webserver.ListenAndServe(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
