package app

import (
	"embed"
	"log"
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

const JourneyCookieName string = "visit_journey"

func RouterSetup(db *sqlx.DB, uploadsDir string, sessions *scs.SessionManager) chi.Router {

	r := chi.NewRouter()
	r.Use(sessions.LoadAndSave)
	r.Handle("/public/*", neuterDirectoryHandler(http.FileServer(http.FS(staticFS))))
	r.Group(func(site chi.Router) {

		//site.Use(basicAuthHandler("matt", "test"))

		site.HandleFunc("/", indexPageHandler(sessions))
		site.Route("/visits", func(r chi.Router) {
			r.Get("/choose-customer", visitStepOneHandler(db))
			r.Post("/choose-customer", visitStepOneSubmitHandler(db, sessions))
			r.Get("/choose-location", visitStepTwoHandler(db, sessions))
			r.Post("/choose-location", visitStepTwoSubmitHandler(db, sessions))
			r.Route("/log-visit", func(r chi.Router) {
				r.Get("/", visitStepThreeHandler(db, sessions))
				r.Post("/", visitStepThreeSubmitHandler(db, uploadsDir, sessions))
				r.Get("/confirm", visitConfirmationHandler(db, sessions))
				r.Post("/validate-date", validateVisitDateHandler)
				r.Post("/validate-notes", validateVisitNotesHandler)
				r.Post("/validate-time", validateVisitTimeHandler)
			})

		})
	})

	return r
}

func LogInfo(msg string)  { log.Println("INFO: " + msg) }
func LogError(msg string) { log.Println("ERROR: " + msg) }
