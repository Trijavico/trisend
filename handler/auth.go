package handler

import (
	"net/http"
	"trisend/auth"

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
