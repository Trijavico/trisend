package server

import (
	"archive/zip"
	"fmt"
	"net/http"
	"time"
	"trisend/handler"
	"trisend/tunnel"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func NewServer() *http.Server {
	router := http.NewServeMux()

	router.HandleFunc("/download/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		filename := r.URL.Query().Get("filename")
		if filename == "" {
			filename = "unknown"
		}

		done := make(chan struct{})
		streamChan, ok := tunnel.GetStream(id)
		defer tunnel.DeleteStream(id)
		if !ok {
			http.Error(w, "stream not found", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", "trisend"))
		w.Header().Set("Content-Type", "application/zip")

		zipWriter := zip.NewWriter(w)
		defer zipWriter.Close()
		fileWriter, err := zipWriter.Create(filename)
		if err != nil {
			http.Error(w, "stream not found", http.StatusInternalServerError)
			return
		}

		streamChan <- tunnel.Stream{
			Writer: fileWriter,
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
