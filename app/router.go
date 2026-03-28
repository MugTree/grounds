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

func ServerSetup(db *sqlx.DB, uploadsDir string) chi.Router {

	r := chi.NewRouter()
	r.Handle("/public/*", neuterDirectoryHandler(http.FileServer(http.FS(staticFS))))

	r.Group(func(site chi.Router) {

		//site.Use(basicAuthHandler("matt", "test"))

		site.HandleFunc("/", indexPageHandler())

		site.Route("/visits", func(r chi.Router) {
			r.Get("/choose-customer", chooseCustomerHandler(db))
			r.Post("/choose-customer", chooseCustomerSubmitHandler(db))
			r.Get("/choose-location", chooseLocationHandler(db))
			r.Post("/choose-location", chooseLocationSubmitHandler(db))
			r.Get("/log-visit", logVisitHandler(db))
			r.Post("/log-visit/confirm", logVisitSubmitHandler(db, uploadsDir))
			r.Post("/log-visit/validate-date", validateVisitDateHandler)
			r.Post("/log-visit/validate-notes", validateVisitNotesHandler)
			r.Post("/log-visit/validate-time", validateVisitTimeHandler)
		})
	})

	return r
}
