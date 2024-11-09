package server

import (
	"fmt"
	"io"
	"net/http"
	"time"
	"trisend/util"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type Stream struct {
	Writer io.Writer
	Done   chan struct{}
}

var streamings = map[string]chan Stream{}

func NewServer() *http.Server {
	router := http.NewServeMux()

	router.HandleFunc("/download/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		done := make(chan struct{})
		stream, ok := streamings[id]
		if !ok {
			http.Error(w, "stream not found", http.StatusInternalServerError)
			return
		}

		stream <- Stream{
			Writer: w,
			Done:   done,
		}

		<-done
		delete(streamings, id)
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
		Handler: handleSSH,
		Banner:  banner,
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return false
		},
		ServerConfigCallback: func(ctx ssh.Context) *gossh.ServerConfig {
			conf := &gossh.ServerConfig{}
			conf.AddHostKey(privKey)
			return conf
		},
	}
}

func handleSSH(session ssh.Session) {
	id := util.GetRandomID()
	fmt.Fprintf(session, "LINK: http://localhost:3000/download/%s\n", id)
	streamings[id] = make(chan Stream)

	stream := <-streamings[id]

	_, err := io.Copy(stream.Writer, session)
	if err != nil {
		session.Write([]byte("an error occurred while streaming data"))
	}

	close(stream.Done)
}
