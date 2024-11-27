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
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	gossh "golang.org/x/crypto/ssh"
)

func NewAuth() {
	sessionKey := []byte(util.GetEnvStr("SESSION_SECRET", ""))
	cookieStore := sessions.NewCookieStore(sessionKey)
	gothic.Store = cookieStore

	goth.UseProviders(github.New(util.GetEnvStr("CLIENT_ID", ""), util.GetEnvStr("CLIENT_SECRET", ""), ""))
}

func NewServer(files embed.FS) *http.Server {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/" {
			message := fmt.Sprintf("url: %s", r.URL)
			http.Error(w, message, http.StatusNotFound)
			return
		}

		cookie, err := r.Cookie("session_token")
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			views.Home().Render(r.Context(), w)
			return
		}

		fmt.Printf("COOKIE VALUE: %s\n", cookie.Value)
		views.Home().Render(r.Context(), w)
	})

	router.Handle("/assets/", http.FileServer(http.FS(files)))

	router.Handle("/login", templ.Handler(views.Login()))
	router.HandleFunc("/auth/{action}", func(w http.ResponseWriter, r *http.Request) {
		action := r.PathValue("action")

		switch action {
		case "login":
			gothic.BeginAuthHandler(w, r)

		case "callback":
			user, err := gothic.CompleteUserAuth(w, r)
			if err != nil {
				return
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"name":     user.Name,
				"email":    user.Email,
				"username": user.NickName,
			})

			tokenString, err := token.SignedString([]byte(util.GetEnvStr("JWT_SECRET", "TESTSecret")))
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}

			cookie := &http.Cookie{
				Name:     "session_token",
				Value:    tokenString,
				Path:     "/",
				Expires:  time.Time{},
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			}

			http.SetCookie(w, cookie)
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
