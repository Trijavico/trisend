package server

import (
	"fmt"
	"net/http"
	"trisend/tunnel"
	"trisend/views"
)

func handleDownloadPage(w http.ResponseWriter, r *http.Request) {
	fullURL := fmt.Sprintf("%s/download/direct/%s",
		r.URL.Hostname(),
		r.PathValue("id"),
	)
	views.Download(fullURL).Render(r.Context(), w)
}

func handleTransferFiles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	details, err := tunnel.GetStreamDetails(id)
	if err != nil {
		handleError(w, "an error occurred", err, http.StatusInternalServerError)
		return
	}
	done := make(chan struct{})

	channel, ok := tunnel.GetStream(id)
	defer tunnel.DeleteStream(id)
	if !ok {
		http.Error(w, "stream not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", details.Filename))
	w.Header().Set("Content-Type", "application/zip")

	channel <- tunnel.Stream{
		Writer: w,
		Done:   done,
	}

	<-done
}
