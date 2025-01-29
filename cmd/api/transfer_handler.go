package main

import (
	"fmt"
	"net/http"
	"trisend/internal/tunnel"
	"trisend/internal/types"
	"trisend/internal/views"
	"trisend/internal/views/components"
)

func handleDownloadPage(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(SESSION_COOKIE)
	user := value.(*types.Session)

	id := r.PathValue("id")
	details, ok := tunnel.GetStreamDetails(id)
	if !ok {
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

	channel, ok := tunnel.GetStream(id)
	if !ok {
		views.NotFound(user).Render(r.Context(), w)
		return
	}
	defer tunnel.DeleteStream(id)

	done := make(chan struct{})
	Error := make(chan struct{})

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
