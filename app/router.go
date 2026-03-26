package app

import (
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"fmt"
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

/*
	add customer
	add location
	add employee
	show visits
	show locations
	show employees
*/

func ServerSetup(db *sqlx.DB, uploadsDir string) chi.Router {
	r := chi.NewRouter()
	r.Handle("/public/*", neuter(http.FileServer(http.FS(staticFS))))

	r.Group(func(site chi.Router) {
		site.Use(basicAuthMiddleware("matt", "test"))

		site.HandleFunc("/", indexPage())

		site.Route("/visits", func(r chi.Router) {
			r.Get("/choose-customer", chooseCustomer(db))
			r.Post("/choose-customer", chooseCustomerSubmit(db))
			r.Get("/{customer_id}/choose-location", chooseLocation(db))
			r.Post("/{customer_id}/choose-location", choosteLocationSubmit(db))
			r.Get("/{location_id}/log-visit", logVisit(db))
			r.Post("/log-visit", logVisitSubmit(db, uploadsDir))
			r.Post("/log-visit/confirm", logVisitConfirm(db))
		})
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

func basicAuthMiddleware(user string, user_password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if ok {
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))
				expectedUsernameHash := sha256.Sum256([]byte(user))
				expectedPasswordHash := sha256.Sum256([]byte(user_password))

				usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
				passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

				if usernameMatch && passwordMatch {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}
