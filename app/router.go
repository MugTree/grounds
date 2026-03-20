package app

import (
	"embed"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

//go:embed public/css/*.css
//go:embed public/js/*.js
var staticFS embed.FS

// var (
// 	StaticSys = hashfs.NewFS(staticFS)
// )

// admin features
/*
	add customer
	add location
	add employee
	show visits
	show locations
	show employees
*/

func AppSetup(db *sqlx.DB, uploadsDir string) chi.Router {
	r := chi.NewRouter()
	fs := http.FileServerFS(staticFS)
	r.Handle("/public/*", neuter(fs))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/visit/", 303)
	})

	r.Route("/visit", func(r chi.Router) {
		r.Get("/", handleHomepageGet(db))
		r.Post("/", handleHomepagePost(db))
		r.Route("/add", func(r chi.Router) {
			r.Get("/{location_id}/", handleVisitGet(db))
			r.Post("/{location_id}/", handleVisitPost(db, uploadsDir))
		})
	})

	r.Get("/testing", func(w http.ResponseWriter, r *http.Request) {
		TestPage().Render(r.Context(), w)
	})
	return r
}

// this needs to return an error code as well???
func renderServerError(w http.ResponseWriter, r *http.Request, msg string) {
	LogError(msg)
	ErrorPage().Render(r.Context(), w)
}

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LogInfo(msg string)  { log.Println("INFO: " + msg) }
func LogError(msg string) { log.Println("ERROR: " + msg) }
