package server

import (
	"log/slog"
	"net/http"
)

func handleError(w http.ResponseWriter, msg string, err error, code int) {
	slog.Error(err.Error())
	http.Error(w, msg, code)
}
