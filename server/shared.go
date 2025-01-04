package server

import (
	"log"
	"net/http"
)

func handleError(w http.ResponseWriter, msg string, err error, code int) {
	log.Println(err)
	http.Error(w, msg, code)
}
