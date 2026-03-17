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

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/visit/", 303)
	})

	r.Route("/visit", func(r chi.Router) {
		r.Get("/", handleHomepageGet(db))
		r.Post("/", handleHomepagePost(db))
		r.Route("/add", func(r chi.Router) {
			r.Get("/{location_id}/", handleVisitGet(db))
			r.Post("/{location_id}/", handleVisitPost(db))
		})
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

func LogInfo(msg string)  { log.Println("INFO: " + msg) }
func LogError(msg string) { log.Println("ERROR: " + msg) }
