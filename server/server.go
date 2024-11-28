package server

import (
	"net/http"
	"time"
	"trisend/handler"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func NewServer() *http.Server {
	router := http.NewServeMux()
	registerRoutes(router)

	return &http.Server{
		Addr:         "0.0.0.0:3000",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}
}

func NewSSHServer(privKey gossh.Signer, banner string) *ssh.Server {
	return &ssh.Server{
		Addr:    "0.0.0.0:2222",
		Handler: handler.HandleSSH,
		Banner:  banner,
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return false
		},
		ServerConfigCallback: func(ctx ssh.Context) *gossh.ServerConfig {
			conf := &gossh.ServerConfig{}
			conf.AddHostKey(privKey)
			return conf
		},
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": handler.HandleSFTP,
		},
	}
}
