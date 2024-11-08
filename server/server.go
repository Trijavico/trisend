package server

import (
	"fmt"
	"net/http"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func NewServer() *http.Server {
	router := http.NewServeMux()

	router.HandleFunc("/download/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("here to download"))
	})

	return &http.Server{
		Addr:         "0.0.0.0:3000",
		Handler:      router,
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  0,
	}
}

func NewSSHServer(privKey gossh.Signer, banner string) *ssh.Server {
	return &ssh.Server{
		Addr: "0.0.0.0:2222",
		Handler: func(s ssh.Session) {
			fmt.Fprintln(s, "Hello world from server")
		},
		Version: "trisend-0.1",
		Banner:  "banner",
		// PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {},
		ServerConfigCallback: func(ctx ssh.Context) *gossh.ServerConfig {
			conf := &gossh.ServerConfig{}
			conf.AddHostKey(privKey)
			return conf
		},
	}
}
