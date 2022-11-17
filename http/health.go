package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/maragudk/service/model"
)

type jobCreator interface {
	CreateJob(ctx context.Context, name string, payload model.Map, timeout time.Duration) error
}

func Health(mux chi.Router, db jobCreator) {
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.CreateJob(r.Context(), "health", model.Map{}, time.Second); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("OK"))
	})
}
