package app

import (
	"embed"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

//go:embed public/css/*.css
//go:embed public/js/*.js
//go:embed public/js/manifest.json
//go:embed public/img/*.png
var staticFS embed.FS

func ServerSetup(db *sqlx.DB, uploadsDir string, cookieKey []byte) chi.Router {

	r := chi.NewRouter()
	r.Handle("/public/*", neuterDirectoryHandler(http.FileServer(http.FS(staticFS))))

	r.Group(func(site chi.Router) {

		//site.Use(basicAuthHandler("matt", "test"))

		site.HandleFunc("/", indexPageHandler())

		site.Route("/visits", func(r chi.Router) {
			r.Get("/choose-customer", stepOneHandler(db))
			r.Post("/choose-customer", stepOneSubmitHandler(db, cookieKey))
			r.Get("/choose-location", stepTwoHandler(db, cookieKey))
			r.Post("/choose-location", stepTwoSubmitHandler(db, cookieKey))
			r.Route("/log-visit", func(r chi.Router) {
				r.Get("/", stepThreeHandler(db, cookieKey))
				r.Post("/", stepThreeSubmitHandler(db, uploadsDir, cookieKey))
				r.Get("/complete", confirmationHandler(db, cookieKey))
				r.Post("/validate-date", validateVisitDateHandler)
				r.Post("/validate-notes", validateVisitNotesHandler)
				r.Post("/validate-time", validateVisitTimeHandler)
			})

		})
	})

	return r
}
