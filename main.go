package main

import (
	"embed"
	"log"
	"os"
	"trisend/server"

	gossh "golang.org/x/crypto/ssh"
)

//go:embed "assets"
var Files embed.FS

func main() {
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

	httpserver := server.NewServer(Files)
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
