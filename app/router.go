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
//go:embed public/js/manifest.json
//go:embed public/img/*.png
var staticFS embed.FS

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
		r.Get("/choose-customer", chooseCustomer(db))
		r.Post("/choose-customer", chooseCustomerSubmit(db))
		r.Get("/{customer_id}/choose-location", chooseLocation(db))
		r.Post("/{customer_id}/choose-location", choosteLocationSubmit(db))
		r.Get("/{location_id}/log-visit", logVisit(db))
		r.Post("/log-visit", logVisitSubmit(db, uploadsDir))
	})

	r.Get("/testing", func(w http.ResponseWriter, r *http.Request) {
		TestPage().Render(r.Context(), w)
	})
	return r
}

func renderServerError(w http.ResponseWriter, r *http.Request, msg string, statusCode ...int) {
	status := 500
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	_, file, line, ok := runtime.Caller(1) // 1 = caller of this function
	if ok {
		msg = fmt.Sprintf("%s (at %s:%d)", msg, file, line)
	}
	LogError(msg)
	w.WriteHeader(status)
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
