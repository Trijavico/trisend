package main

import (
	gossh "golang.org/x/crypto/ssh"
	"log"
	"os"
	"trisend/server"
)

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

	httpserver := server.NewServer()
	sshserver := server.NewSSHServer(privateKey, string(bannerBytes))

	log.Println("HTTP Server running...")
	log.Println("SSH Server running...")

	go log.Fatal(sshserver.ListenAndServe())
	log.Fatal(httpserver.ListenAndServe())
}