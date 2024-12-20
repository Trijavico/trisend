package server

import (
	"fmt"
	"net/http"
	"trisend/views"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.String() != "/" {
		message := fmt.Sprintf("url: %s", r.URL)
		http.Error(w, message, http.StatusNotFound)
		return
	}
	_, err := r.Cookie(Session_key)
	isNotLoggedIn := err != nil

	if isNotLoggedIn {
		views.Home().Render(r.Context(), w)
	}

	// TODO: Render with a navigation with the logged in user pfp
	views.Home().Render(r.Context(), w)
}
