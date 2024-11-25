package server

import (
	"embed"
	"fmt"
	"net/http"
	"time"
	"trisend/handler"
	"trisend/tunnel"
	"trisend/util"
	"trisend/views"

	"github.com/a-h/templ"
	"github.com/gliderlabs/ssh"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	gossh "golang.org/x/crypto/ssh"
)

func NewServer(files embed.FS) *http.Server {
	sessionKey := []byte(util.GetEnvStr("SESSION_SECRET", ""))
	cookieStore := sessions.NewCookieStore(sessionKey)
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // 24 hours
		Secure:   true,
		HttpOnly: true,
	}

	gothic.Store = cookieStore

	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/" {
			message := fmt.Sprintf("url: %s", r.URL)
			http.Error(w, message, http.StatusNotFound)
			return
		}
		views.Home().Render(r.Context(), w)
	})
	router.Handle("/assets/", http.FileServer(http.FS(files)))

	router.Handle("/login", templ.Handler(views.Login()))
	router.HandleFunc("/auth/{action}", func(w http.ResponseWriter, r *http.Request) {
		action := r.PathValue("action")
		provider := r.PathValue("provider")

		switch action {
		case "login":
			gothic.BeginAuthHandler(w, r)

		case "callback":
			user, err := gothic.CompleteUserAuth(w, r)
			if err != nil {
				return
			}
			fmt.Printf("username: %s provider: %s\n", user.Name, provider)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
	})

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
