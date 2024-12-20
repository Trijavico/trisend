package server

import (
	"net/http"
	"trisend/db"
	"trisend/public"
	"trisend/views"

	"github.com/a-h/templ"
)

func (wb *WebServer) AddRoutes(userStore db.UserStore, sessStore db.SessionStore) {
	handler := http.NewServeMux()
	wb.server.Handler = handler

	handler.HandleFunc("/", handleIndex)
	handler.Handle("/assets/", http.FileServer(http.FS(public.Files)))

	handler.Handle("GET /login", templ.Handler(views.Login()))
	handler.HandleFunc("POST /login/send-code", handleAuthCode(sessStore))
	handler.HandleFunc("POST /login/verify-code", handleVerification(sessStore))
	handler.HandleFunc("GET /auth/{action}", handleOAuth)

	// app.Group(func(onboarding chi.Router) {
	// 	onboarding.Use()
	// 	onboarding.Use()
	//
	// 	onboarding.Get()
	// 	onboarding.Post()
	// 	onboarding.Get()
	// })

	// TODO: userKeysView, C.R.U.D operations
	// TODO: profileView C.R.U.D operations
	// onboardingView

	handler.HandleFunc("GET /download/{id}", handleDownloadPage) // TODO: protect
	handler.HandleFunc("GET /stream-data", handleTransferFiles)  // TODO: protect
}
