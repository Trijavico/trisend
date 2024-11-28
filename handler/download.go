package handler

import (
	"fmt"
	"net/http"
	"trisend/tunnel"
	"trisend/views"
)

func HandleDownloadPage(w http.ResponseWriter, r *http.Request) {
	fullURL := fmt.Sprintf("%s/stream/%s?zip=%s",
		r.URL.Hostname(),
		r.PathValue("id"),
		r.URL.Query().Get("zip"),
	)
	views.Download(fullURL).Render(r.Context(), w)
}

func HandleTransferFiles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	zipParam := r.URL.Query().Get("zip")

	done := make(chan struct{})
	streamChan, ok := tunnel.GetStream(id)
	defer tunnel.DeleteStream(id)
	if !ok {
		http.Error(w, "stream not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", zipParam))
	w.Header().Set("Content-Type", "application/zip")

	streamChan <- tunnel.Stream{
		Writer: w,
		Done:   done,
	}

	<-done
}
