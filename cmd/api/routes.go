package main

import (
	"net/http"
	"trisend/internal/views"
	"trisend/public"

	"github.com/a-h/templ"
)

func AddRoutes(app App) *http.ServeMux {
	handler := http.NewServeMux()

	handler.HandleFunc("/", handleHome)
	handler.Handle("/assets/", http.FileServer(http.FS(public.Files)))
	handler.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))

	handler.Handle("POST /logout", handleLogout(app))
	handler.Handle("GET /login", templ.Handler(views.Login()))
	handler.Handle("GET /login/create", templ.Handler(views.FillProfile()))
	handler.Handle("POST /login/create", handleLoginCreate(app))
	handler.Handle("POST /login/send-code", handleAuthCode(app))
	handler.Handle("POST /login/verify-code", handleVerification(app))
	handler.Handle("GET /auth/{action}", handleOAuth(app))

	handler.Handle("GET /keys", WithAuth(handleKeysView(app)))
	handler.Handle("POST /keys", WithAuth(handleCreateKey(app)))
	handler.Handle("GET /keys/create", WithAuth(handleCreateKeyView()))
	handler.Handle("DELETE /keys/{id}", WithAuth(handleDeleteKey(app)))

	handler.HandleFunc("GET /download/{id}", WithAuth(handleDownloadPage))
	handler.HandleFunc("GET /download/direct/{id}", WithAuth(handleTransferFiles))

	return handler
}
