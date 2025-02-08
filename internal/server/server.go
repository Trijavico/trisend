package server

import (
	_ "embed"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"
	"trisend/internal/config"
	"trisend/internal/db"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

//go:embed banner.txt
var banner string

type Server struct {
	httpServer *http.Server
	sshServer  *ssh.Server
}

func NewWebServer() *Server {
	httpServer := &http.Server{
		Addr:         net.JoinHostPort("0.0.0.0", config.SERVER_PORT),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}

	var sshport string
	if config.SSH_PORT != "" {
		sshport = net.JoinHostPort("0.0.0.0", config.SSH_PORT)
	} else {
		sshport = net.JoinHostPort("0.0.0.0", "22")
	}

	sshServer := &ssh.Server{
		Addr:    sshport,
		Handler: handleSSH,
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return false
		},
	}

	return &Server{
		httpServer: httpServer,
		sshServer:  sshServer,
	}
}

func (server *Server) SetupConfig(router *http.ServeMux, privKey gossh.Signer, userStore db.UserStore) {
	configCallback := func(ctx ssh.Context) *gossh.ServerConfig {
		conf := &gossh.ServerConfig{}
		conf.AddHostKey(privKey)
		return conf
	}

	server.httpServer.Handler = router
	server.sshServer.Banner = banner
	server.sshServer.PublicKeyHandler = handlePublicKey(userStore)
	server.sshServer.ServerConfigCallback = configCallback
	server.sshServer.SubsystemHandlers = map[string]ssh.SubsystemHandler{
		"sftp": handleSFTP(userStore),
	}
}

func (server *Server) ListenAndServe() {
	go func() {
		slog.Info("SSH Server running")
		if err := server.sshServer.ListenAndServe(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()

	slog.Info(fmt.Sprintf("HTTP Server running on PORT: %s", config.SERVER_PORT))
	if err := server.httpServer.ListenAndServe(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

}
