package server

import (
	"net"
	"net/http"
	"time"
	"trisend/config"
	"trisend/db"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type WebServer struct {
	server *http.Server
}

func NewWebServer() *WebServer {
	address := net.JoinHostPort("0.0.0.0", config.SERVER_PORT)
	server := &http.Server{
		Addr:         address,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}

	return &WebServer{
		server: server,
	}
}

func (wbserver *WebServer) ListenAndServe() error {
	return wbserver.server.ListenAndServe()
}

func NewSSHServer(privKey gossh.Signer, banner string, userStore db.UserStore) *ssh.Server {
	address := net.JoinHostPort("0.0.0.0", config.SSH_PORT)
	subSysHandlers := map[string]ssh.SubsystemHandler{
		"sftp": handleSFTP,
	}

	configCallback := func(ctx ssh.Context) *gossh.ServerConfig {
		conf := &gossh.ServerConfig{}
		conf.AddHostKey(privKey)
		return conf
	}

	return &ssh.Server{
		Addr:                 address,
		Banner:               banner,
		Handler:              handleSSH,
		PublicKeyHandler:     handlePublicKey(userStore),
		ServerConfigCallback: configCallback,
		SubsystemHandlers:    subSysHandlers,
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return false
		},
	}
}
