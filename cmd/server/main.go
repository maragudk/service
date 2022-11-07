package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maragudk/env"
	"github.com/maragudk/errors"
	"golang.org/x/sync/errgroup"

	"github.com/maragudk/service/http"
	"github.com/maragudk/service/sql"
)

func main() {
	os.Exit(start())
}

func start() int {
	log := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	log.Println("Starting")

	_ = env.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	db := sql.NewDatabase(sql.NewDatabaseOptions{
		Log:                   log,
		URL:                   env.GetStringOrDefault("DATABASE_URL", "file:app.db"),
		MaxOpenConnections:    env.GetIntOrDefault("DATABASE_MAX_OPEN_CONNS", 5),
		MaxIdleConnections:    env.GetIntOrDefault("DATABASE_MAX_IDLE_CONNS", 5),
		ConnectionMaxLifetime: env.GetDurationOrDefault("DATABASE_CONN_MAX_LIFETIME", time.Hour),
		ConnectionMaxIdleTime: env.GetDurationOrDefault("DATABASE_CONN_MAX_IDLE_TIME", time.Hour),
	})

	if err := db.Connect(); err != nil {
		log.Println("Error connecting to database:", err)
		return 1
	}

	s := http.NewServer(http.NewServerOptions{
		Database: db,
		Host:     env.GetStringOrDefault("HOST", ""),
		Log:      log,
		Port:     env.GetIntOrDefault("PORT", 8080),
	})

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := s.Start(); err != nil {
			return errors.Wrap(err, "error starting server")
		}
		return nil
	})

	<-ctx.Done()
	log.Println("Stopping")

	eg.Go(func() error {
		if err := s.Stop(); err != nil {
			return errors.Wrap(err, "error stopping server")
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Println("Error:", err)
		return 1
	}

	log.Println("Stopped")

	return 0
}
