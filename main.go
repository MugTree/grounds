package main

import (
	"context"
	"fmt"
	"log"
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(ctx, *logger); err != nil {
		fmt.Println(err)
	}
}

func run(parent context.Context, logger slog.Logger) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	fmt.Println("reading .env")

	dbPath := mustEnv("VT_APP_DB")
	appPort := mustEnv("VT_APP_PORT")

	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("sql: error opening DB - %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("sql: error pinging DB - %v", err)
	}

	appRouterSetup := func() func() chi.Router {
		return func() chi.Router {
			return www.AppSetup(db, logger)
		}
	}

	app := appRouterSetup()
	return runBlocking(ctx, appPort, app())
}

func runBlocking(ctx context.Context, host string, app http.Handler) error {

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", host),
		Handler: app,
	}

	fmt.Println("starting server")

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fmt.Println("Shutting down server...")

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	fmt.Printf("Server running at http://localhost%s", server.Addr)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	fmt.Println("Server stopped.")
	return nil
}

func mustEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("Missing required env var: %s", key)
	}
	return val
}
