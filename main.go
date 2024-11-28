package main

import (
	"log"
	"os"
	"trisend/server"

	"github.com/joho/godotenv"
	gossh "golang.org/x/crypto/ssh"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error: %s\n", err)
	}

	keyBytes, err := os.ReadFile("./keys/host")
	if err != nil {
		log.Println("key not found")
		os.Exit(1)
	}
	privateKey, err := gossh.ParsePrivateKey(keyBytes)
	if err != nil {
		log.Println("error parsing key bytes")
		os.Exit(1)
	}

	bannerBytes, err := os.ReadFile("banner.txt")
	if err != nil {
		log.Println("banner not found")
		os.Exit(1)
	}

	httpserver := server.NewServer()
	sshserver := server.NewSSHServer(privateKey, string(bannerBytes))

	log.Println("HTTP Server running...")
	log.Println("SSH Server running...")

	go func() {
		if err := sshserver.ListenAndServe(); err != nil {
			log.Fatalf("SSH server failed %v", err)
		}
	}()

	if err := httpserver.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server failed %v", err)
	}
}
