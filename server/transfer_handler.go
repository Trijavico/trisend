package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"trisend/tunnel"
	"trisend/types"
	"trisend/views"
	"trisend/views/components"
)

func handleDownloadPage(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(SESSION_COOKIE)
	user := value.(*types.Session)

	id := r.PathValue("id")
	details, err := tunnel.GetStreamDetails(id)
	if err != nil {
		slog.Error(err.Error())
		views.NotFound(user).Render(r.Context(), w)
		return
	}

	url := fmt.Sprintf("%s/download/direct/%s", r.URL.Hostname(), id)

	profileBtn := components.ProfileButton(user)
	views.Download(details, url, profileBtn).Render(r.Context(), w)
}

func handleTransferFiles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	value := r.Context().Value(SESSION_COOKIE)
	user := value.(*types.Session)

	done := make(chan struct{})
	Error := make(chan struct{})

	channel, ok := tunnel.GetStream(id)
	defer tunnel.DeleteStream(id)
	if !ok {
		views.NotFound(user).Render(r.Context(), w)
		return
	}

	channel <- tunnel.Stream{
		Writer: w,
		Done:   done,
		Error:  Error,
	}

	select {
	case <-done:
	case <-Error:
		views.NotFound(user).Render(r.Context(), w)
	}
}
