package app

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"runtime"
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
		http.Redirect(w, r, "/visits/choose-customer", 303)
	})

	r.Route("/visits", func(r chi.Router) {
		r.Get("/choose-customer", chooseCustomerGet(db))
		r.Post("/choose-customer", chooseCustomerPost())

		r.Get("/choose-location", chooseLocationGet(db))
		r.Post("/choose-location", chooseLocationPost())

		r.Get("/log-visit", logLocationGet(db))
	})

	r.Get("/testing", func(w http.ResponseWriter, r *http.Request) {
		TestPage().Render(r.Context(), w)
	})
	return r
}

func renderServerError(w http.ResponseWriter, r *http.Request, msg string) {
	_, file, line, ok := runtime.Caller(1) // 1 = caller of this function
	if ok {
		msg = fmt.Sprintf("%s (at %s:%d)", msg, file, line)
	}
	LogError(msg)
	w.WriteHeader(500)
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
