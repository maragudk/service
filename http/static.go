package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Static(mux chi.Router) {
	staticHandler := http.FileServer(http.Dir("public"))
	mux.Get(`/{:[^.]+\.[^.]+}`, staticHandler.ServeHTTP)
	mux.Get(`/{:images|scripts|styles}/*`, staticHandler.ServeHTTP)
}
