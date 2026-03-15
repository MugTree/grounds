package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"main/www"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(ctx, *log); err != nil {
		log.Error(err.Error())
	}
}

func run(parent context.Context, logger slog.Logger) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	logger.Info("reading .env")

	mustEnv := func(key string) string {
		val, ok := os.LookupEnv(key)
		if !ok {
			logger.Error(fmt.Sprintf("Missing required env var: %s", key))
			os.Exit(1)
		}
		return val
	}

	dbPath := mustEnv("VT_APP_DB")
	appPort := mustEnv("VT_APP_PORT")

	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("sql: error opening DB - %v", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("sql: error pinging DB - %v", err)
	}

	appRouterSetup := func() func() chi.Router {
		return func() chi.Router {
			return www.AppSetup(db, logger)
		}
	}

	app := appRouterSetup()
	return runBlocking(ctx, appPort, app(), logger)
}

func runBlocking(ctx context.Context, host string, app http.Handler, logger slog.Logger) error {

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", host),
		Handler: app,
	}

	logger.Info("Starting app server")

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logger.Info("Shutting down server...")

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error(fmt.Sprintf("Error during shutdown: %v", err))
		}
	}()

	logger.Info(fmt.Sprintf("Server running at http://localhost%s", server.Addr))

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	logger.Info("Server stopped.")
	return nil
}
