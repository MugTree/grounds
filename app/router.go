package app

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

//go:embed public/css/*.css
//go:embed public/js/*.js
var staticFS embed.FS

var (
	StaticSys = hashfs.NewFS(staticFS)
)

// admin features
/*
	add customer
	add location
	add employee
	show visits
	show locations
	show employees
*/

func AppSetup(db *sqlx.DB) chi.Router {
	r := chi.NewRouter()
	r.Handle("/public/*", hashfs.FileServer(StaticSys))
	r.Get("/", handleHomepageGet(db))
	r.Post("/", handleHomepagePost(db))

	r.Get("/new-visit/{location_id}/", handleNewVisit(db))

	r.Post("/create-visit/", handleCreateVisit(db))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK!")
	})

	r.Get("/testing", func(w http.ResponseWriter, r *http.Request) {
		TestPage().Render(r.Context(), w)
	})
	return r
}

func staticPath(format string, args ...any) string {
	return "/" + StaticSys.HashName(fmt.Sprintf("public/"+format, args...))
}

func renderServerError(w http.ResponseWriter, r *http.Request, msg string) {
	LogError(msg)
	ErrorPage().Render(r.Context(), w)
}

func LogInfo(msg string) { log.Println("INFO: " + msg) }

// func Warn(msg string)     { log.Println("WARN: " + msg) }
func LogError(msg string) { log.Println("ERROR: " + msg) }
