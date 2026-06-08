package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mugtree/grounds/app"
	"github.com/mugtree/grounds/app/db"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"

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

	mustEnv := func(key string) string {
		val, ok := os.LookupEnv(key)
		if !ok {
			log.Fatalf("missing .env: %s", key)
		}
		return val
	}

	dbPath := mustEnv("APP_DB")
	appPort := mustEnv("APP_PORT")
	appUser := mustEnv("APP_USER")
	appPassword := mustEnv("APP_PASSWORD")
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

	webserver := &http.Server{
		Addr:    ":" + appPort,
		Handler: app.SetupHttpServer(queries, dbHandle, uploadsDir, sessionManager, appUser, appPassword),
	}

	application := app.NewApp(dbHandle, queries, webserver)

	// bind OS signal context → app shutdown
	go func() {
		<-parent.Done()
		application.Stop()
	}()

	application.Start()

	if err := application.Wait(); err != nil {
		return err
	}

	app.LogInfo("server stopped cleanly")
	return nil

}
