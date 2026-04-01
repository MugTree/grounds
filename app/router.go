package app

import (
	"embed"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

//go:embed public/css/*.css
//go:embed public/js/*.js
//go:embed public/js/manifest.json
//go:embed public/img/*.png
var staticFS embed.FS

func ServerSetup(db *sqlx.DB, uploadsDir string, sessions *scs.SessionManager) chi.Router {

	r := chi.NewRouter()
	r.Use(sessions.LoadAndSave)
	r.Handle("/public/*", neuterDirectoryHandler(http.FileServer(http.FS(staticFS))))
	r.Group(func(site chi.Router) {

		//site.Use(basicAuthHandler("matt", "test"))

		site.HandleFunc("/", indexPageHandler(sessions))
		site.Route("/visits", func(r chi.Router) {
			r.Get("/choose-customer", stepOneHandler(db))
			r.Post("/choose-customer", stepOneSubmitHandler(db, sessions))
			r.Get("/choose-location", stepTwoHandler(db, sessions))
			r.Post("/choose-location", stepTwoSubmitHandler(db, sessions))
			r.Route("/log-visit", func(r chi.Router) {
				r.Get("/", stepThreeHandler(db, sessions))
				r.Post("/", stepThreeSubmitHandler(db, uploadsDir, sessions))
				r.Get("/complete", confirmationHandler(db, sessions))
				r.Post("/validate-date", validateVisitDateHandler)
				r.Post("/validate-notes", validateVisitNotesHandler)
				r.Post("/validate-time", validateVisitTimeHandler)
			})

		})
	})

	return r
}
