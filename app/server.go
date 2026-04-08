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

const JourneyCookieName string = "visit_journey"

func RouterSetup(db *db.Queries, sqldb *sql.DB, uploadsDir string, sessions *scs.SessionManager) chi.Router {

	r := chi.NewRouter()
	r.Use(sessions.LoadAndSave)
	r.Handle("/public/*", neuterDirectoryHandler(http.FileServer(http.FS(staticFS))))
	r.Group(func(site chi.Router) {

		//site.Use(basicAuthHandler("matt", "test"))

		site.HandleFunc("/", indexPageHandler(db, sessions))

		site.Route("/visit", func(r chi.Router) {
			r.Get("/{visit_id}", func(w http.ResponseWriter, r *http.Request) {
				vid, ok := pathValueAsIntOrErr(w, r, "visit_id")
				if !ok {
					return
				}

				d, err := db.GetVisitById(r.Context(), vid) //selectVisitData(r.Context(), queries, vid)
				if err != nil {
					errorHandler(w, r, err.Error(), 500)
					return
				}

				ViewVisitTemplate(d).Render(r.Context(), w)
			})
			r.Get("/step-1/", visitStepOneHandler(db))
			r.Post("/step-1/", visitStepOneSubmitHandler(db, sessions))
			r.Get("/step-2/", visitStepTwoHandler(db, sessions))
			r.Post("/step-2/", visitStepTwoSubmitHandler(db, sessions))
			r.Route("/step-3", func(r chi.Router) {
				r.Get("/", visitStepThreeHandler(db, sessions))
				r.Post("/", visitStepThreeSubmitHandler(db, sqldb, uploadsDir, sessions))
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
