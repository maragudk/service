package http

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/maragudk/service/sql"
)

type Server struct {
	address  string
	database *sql.Database
	log      *log.Logger
	mux      chi.Router
	server   *http.Server
	metrics  *prometheus.Registry
}

type NewServerOptions struct {
	Database *sql.Database
	Host     string
	Log      *log.Logger
	Metrics  *prometheus.Registry
	Port     int
}

// NewServer returns an initialized, but unstarted Server.
// If no logger is provided, logs are discarded.
func NewServer(opts NewServerOptions) *Server {
	if opts.Log == nil {
		opts.Log = log.New(io.Discard, "", 0)
	}

	if opts.Metrics == nil {
		opts.Metrics = prometheus.NewRegistry()
	}

	address := net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port))
	mux := chi.NewMux()

	return &Server{
		address:  address,
		database: opts.Database,
		log:      opts.Log,
		metrics:  opts.Metrics,
		mux:      mux,
		server: &http.Server{
			Addr:              address,
			Handler:           mux,
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      5 * time.Second,
			IdleTimeout:       5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	s.log.Println("Starting")

	s.setupRoutes()

	s.log.Println("Listening on http://" + s.address)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	s.log.Println("Stopping")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	s.log.Println("Stopped")
	return nil
}
