package main

import (
	"fmt"
	"log/slog"
	"os"
	"trisend/config"
	"trisend/db"
	"trisend/server"

	"github.com/joho/godotenv"
	gossh "golang.org/x/crypto/ssh"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error(fmt.Sprintf("Error loading env variables: %s", err))
		os.Exit(1)
	}
	config.LoadConfig()

	keyBytes, err := os.ReadFile("./keys/host")
	if err != nil {
		slog.Error("key not found")
		os.Exit(1)
	}
	privateKey, err := gossh.ParsePrivateKey(keyBytes)
	if err != nil {
		slog.Error("error parsing key bytes")
		os.Exit(1)
	}

	bannerBytes, err := os.ReadFile("banner.txt")
	if err != nil {
		slog.Info("banner not found")
		os.Exit(1)
	}

	redisDB := db.NewRedisDB()
	userStore := db.NewUserRedisStore(redisDB)
	sessStore := db.NewRedisSessionStore(redisDB)

	webserver := server.NewWebServer()
	sshserver := server.NewSSHServer(privateKey, string(bannerBytes))

	webserver.AddRoutes(userStore, sessStore)

	go func() {
		slog.Info("SSH Server running...")
		if err := sshserver.ListenAndServe(); err != nil {
			slog.Error(fmt.Sprintf("SSH server failed: %s", err))
			os.Exit(1)
		}
	}()

	slog.Info("HTTP Server running...")
	if err := webserver.ListenAndServe(); err != nil {
		slog.Error(fmt.Sprintf("HTTP server failed: %s", err))
		os.Exit(1)
	}
}
