package server

import (
	"fmt"
	"net/http"
)

func internalServerError(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func notFound(w http.ResponseWriter, id string) {
	msg := fmt.Sprintf("File with id %s not found", id)
	http.Error(w, msg, http.StatusNotFound)
}
