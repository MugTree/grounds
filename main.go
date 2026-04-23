package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"main/app"
	"main/app/db"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if err := run(ctx); err != nil {
		app.LogError(err.Error())
	}
}

func run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	mustEnv := func(key string) string {
		val, ok := os.LookupEnv(key)
		if !ok {
			app.LogError(fmt.Sprintf("env: missing required env var: %s", key))
			os.Exit(1)
		}
		return val
	}

	dbPath := mustEnv("APP_DB")
	appPort := mustEnv("APP_PORT")
	uploadsDir := mustEnv("APP_UPLOADS_DIR")

	dbHandle, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("sql: error opening DB - %v", err)
	}

	err = dbHandle.Ping()
	if err != nil {
		return fmt.Errorf("sql: error pinging DB - %v", err)
	}

	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(dbHandle)

	queries := db.New(dbHandle)

	appRouterSetup := func() func() chi.Router {
		return func() chi.Router {
			return app.RouterSetup(queries, dbHandle, uploadsDir, sessionManager)
		}
	}

	app := appRouterSetup()
	return runBlocking(ctx, appPort, app())
}

func runBlocking(ctx context.Context, host string, router http.Handler) error {

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", host),
		Handler: router,
	}

	app.LogInfo("Starting app server")

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.LogInfo("Shutting down server...")

		if err := server.Shutdown(shutdownCtx); err != nil {
			app.LogError(fmt.Sprintf("shutdown: error during shutdown: %v", err))
		}
	}()

	app.LogInfo(fmt.Sprintf("Server running at http://localhost%s", server.Addr))

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	app.LogInfo("Server stopped.")
	return nil
}
