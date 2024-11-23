package server

import (
	"embed"
	"fmt"
	"net/http"
	"time"
	"trisend/handler"
	"trisend/tunnel"
	"trisend/views"

	"github.com/a-h/templ"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func NewServer(files embed.FS) *http.Server {
	router := http.NewServeMux()

	router.Handle("/", templ.Handler(views.Home()))
	router.Handle("/assets/", http.FileServer(http.FS(files)))

	router.HandleFunc("/download/{id}", func(w http.ResponseWriter, r *http.Request) {
		fullURL := fmt.Sprintf("%s/stream/%s?zip=%s", r.URL.Hostname(), r.PathValue("id"), r.URL.Query().Get("zip"))
		views.Download(fullURL).Render(r.Context(), w)
	})

	router.HandleFunc("/stream/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		zipParam := r.URL.Query().Get("zip")

		done := make(chan struct{})
		streamChan, ok := tunnel.GetStream(id)
		defer tunnel.DeleteStream(id)
		if !ok {
			http.Error(w, "stream not found", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", zipParam))
		w.Header().Set("Content-Type", "application/zip")

		streamChan <- tunnel.Stream{
			Writer: w,
			Done:   done,
		}

		<-done
	})

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
