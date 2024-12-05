package handler

import (
	"net/http"
	"trisend/auth"
	"trisend/views"

	"github.com/markbates/goth/gothic"
)

func HandleAuthentication(w http.ResponseWriter, r *http.Request) {
	action := r.PathValue("action")

	switch action {
	case "login":
		gothic.BeginAuthHandler(w, r)

	case "callback":
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			return
		}
		claims := map[string]interface{}{
			"name":     user.Name,
			"email":    user.Email,
			"username": user.NickName,
		}

		auth.CreateSessionCookie(w, claims)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func HandleLoginCode(w http.ResponseWriter, r *http.Request) {
	// emailer := mailer.NewMailer("Verfication code", "example8203@gmail.com", "CODE: 323CdsF2#")
	// if err := emailer.Send(); err != nil {
	// 	fmt.Println(err)
	// 	http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
	// 	return
	// }
	views.ContinueWithCode().Render(r.Context(), w)
}

func HandleVerification(w http.ResponseWriter, r *http.Request) {}
