package app

import (
	"database/sql"
	"embed"
	"log"
	"main/app/db"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

//go:embed public/css/*.css
//go:embed public/js/*.js
//go:embed public/js/manifest.json
//go:embed public/img/*.png
var staticFS embed.FS

const SessionCookieName string = "session"

func RouterSetup(queries *db.Queries, sqldb *sql.DB, uploadsDir string, sessions *scs.SessionManager) chi.Router {

	r := chi.NewRouter()
	r.Use(sessions.LoadAndSave)
	r.Handle("/public/*", neuterDirectoryHandler(http.FileServer(http.FS(staticFS))))
	r.Group(func(site chi.Router) {

		// site.Use(basicAuthHandler("matt", "test"))

		site.HandleFunc("/", indexPageHandler(queries, sessions))
		site.HandleFunc("/login", loginHandler(queries, sessions))

		site.Route("/visit", func(r chi.Router) {
			r.Get("/{visit_id}", visitPageHandler(queries, sessions))
			r.Get("/step-1/", visitStepOneHandler(queries))
			r.Post("/step-1/", visitStepOneSubmitHandler(queries, sessions))
			r.Get("/step-2/", visitStepTwoHandler(queries, sessions))
			r.Post("/step-2/", visitStepTwoSubmitHandler(queries, sessions))
			r.Route("/step-3", func(r chi.Router) {
				r.Get("/", visitStepThreeHandler(queries, sessions))
				r.Post("/", visitStepThreeSubmitHandler(queries, sqldb, uploadsDir, sessions))
				//r.Get("/confirm", visitConfirmationHandler(db, sessions))
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
