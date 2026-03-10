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

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	dbHost := mustEnv("VT_DB_HOST")
	dbPort := mustEnv("VT_DB_PORT")
	dbUser := mustEnv("VT_DB_USER")
	dbPassword := mustEnv("VT_DB_PASSWORD")
	dbName := mustEnv("VT_DB_NAME")
	appLog := mustEnv("VT_APP_LOG")
	appPort := mustEnv("VT_APP_PORT")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
	)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("sql: error opening DB - %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("sql: error pinging DB - %v", err)
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   appLog,
		MaxSize:    50, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	})

	appRouterSetup := func() func() chi.Router {
		return func() chi.Router {
			return app.AppSetup(db)
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

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fmt.Println("Shutting down server...")

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	fmt.Printf("Server running at http://localhost%s", host)

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
