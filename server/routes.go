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

	handler.HandleFunc("/", handleHome)
	handler.Handle("/assets/", http.FileServer(http.FS(public.Files)))
	handler.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))

	handler.Handle("GET /login", templ.Handler(views.Login()))
	handler.Handle("GET /login/create", templ.Handler(views.FillProfile()))
	handler.HandleFunc("POST /login/create", handleLoginCreate(userStore))
	handler.HandleFunc("POST /login/send-code", handleAuthCode(sessStore))
	handler.HandleFunc("POST /login/verify-code", handleVerification(sessStore, userStore))
	handler.HandleFunc("GET /auth/{action}", handleOAuth(userStore))

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

	handler.HandleFunc("GET /download/{id}", handleDownloadPage) // TODO: protect
	handler.HandleFunc("GET /stream-data", handleTransferFiles)  // TODO: protect
}
