package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/honeybadger-io/honeybadger-go"
)

func (s *Server) setupRoutes() {
	s.mux.Use(middleware.Recoverer, honeybadger.Handler)
	s.mux.Use(middleware.Compress(5))
	s.mux.Use(middleware.RealIP)
	s.mux.Use(AddMetrics(s.metrics))

	Health(s.mux, s.database)
	Metrics(s.mux, s.metrics)

	s.mux.Group(func(r chi.Router) {
		r.Use(VersionedAssets)

		Static(r)
	})

	s.mux.Group(func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "text/html; charset=utf-8"))

		Home(r)
	})

	Migrate(s.mux, s.database)
}
