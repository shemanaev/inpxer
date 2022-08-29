package server

import (
	"fmt"
	"net/http"
)

func internalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "Internal server error")
}

func notFound(w http.ResponseWriter, id string) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "File with id %s not found\n", id)
}
