package server

import (
	"net/http"
	"trisend/views"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCookie(r)

	if r.URL.String() != "/" {
		w.WriteHeader(http.StatusNotFound)
		views.NotFound(user).Render(r.Context(), w)
		return
	}

	views.Home(user).Render(r.Context(), w)
}
