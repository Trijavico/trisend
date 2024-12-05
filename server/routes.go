package server

import (
	"fmt"
	"net/http"
	"trisend/handler"
	"trisend/public"
	"trisend/views"

	"github.com/a-h/templ"
)

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/" {
			message := fmt.Sprintf("url: %s", r.URL)
			http.Error(w, message, http.StatusNotFound)
			return
		}

		views.Home().Render(r.Context(), w)
	})

	mux.Handle("/assets/", http.FileServer(http.FS(public.Files)))

	mux.Handle("/login", templ.Handler(views.Login()))
	mux.HandleFunc("POST /login/send-code", handler.HandleLoginCode)
	mux.HandleFunc("/login/verify-code", handler.HandleVerification)

	mux.HandleFunc("/download/{id}", handler.HandleDownloadPage)

	// REST Enpoints
	mux.HandleFunc("/stream-data", handler.HandleTransferFiles)
	mux.HandleFunc("/auth/{action}", handler.HandleAuthentication)
}
