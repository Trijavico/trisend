package server

import (
	"net/http"
	"trisend/db"
	"trisend/public"
	"trisend/server/middleware"
	"trisend/views"

	"github.com/a-h/templ"
)

func (wb *WebServer) AddRoutes(userStore db.UserStore, sessStore db.SessionStore) {
	handler := http.NewServeMux()
	wb.server.Handler = handler

	handler.HandleFunc("/", handleHome)
	handler.Handle("/assets/", http.FileServer(http.FS(public.Files)))
	handler.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))

	handler.Handle("POST /logout", middleware.WithAuth(handleLogout()))
	handler.Handle("GET /login", templ.Handler(views.Login()))
	handler.Handle("GET /login/create", templ.Handler(views.FillProfile()))
	handler.Handle("POST /login/create", handleLoginCreate(userStore))
	handler.Handle("POST /login/send-code", handleAuthCode(sessStore))
	handler.Handle("POST /login/verify-code", handleVerification(sessStore, userStore))
	handler.Handle("GET /auth/{action}", handleOAuth(userStore))

	handler.Handle("GET /keys", middleware.WithAuth(handleKeysView(userStore)))
	handler.Handle("POST /keys", middleware.WithAuth(handleCreateKey(userStore)))
	handler.Handle("GET /keys/create", middleware.WithAuth(handleCreateKeyView()))
	handler.Handle("DELETE /keys/{id}", middleware.WithAuth(handleDeleteKey(userStore)))

	// app.Group(func(onboarding chi.Router) {
	// 	onboarding.Use()
	// 	onboarding.Use()
	//
	// 	onboarding.Get()
	// 	onboarding.Post()
	// 	onboarding.Get()
	// })

	handler.HandleFunc("GET /download/{id}", handleDownloadPage)
	handler.HandleFunc("GET /download/direct/{id}", handleTransferFiles)
}
