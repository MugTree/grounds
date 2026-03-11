package app

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

//go:embed public/css/*.css
//go:embed public/js/*.js
var staticFS embed.FS

var (
	StaticSys = hashfs.NewFS(staticFS)
)

func AppSetup(db *sqlx.DB) chi.Router {

	r := chi.NewRouter()
	r.Handle("/public/*", hashfs.FileServer(StaticSys))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		buildings := []Building{}

		if err := db.Select(&buildings, `SELECT * FROM buildings;`); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting buildings - %v", err))
			return
		}

		Home(buildings).Render(r.Context(), w)
	})

	r.Get("/building/{id}", func(w http.ResponseWriter, r *http.Request) {

		bid := r.PathValue("id")

		jobs := []Jobs{}

		if err := db.Select(&jobs, `SELECT * FROM jobs WHERE building_id = $1;`, bid); err != nil {
			renderServerError(w, r, err.Error())
			return
		}

		JobsPage(jobs).Render(r.Context(), w)

	})

	return r

}

func StaticPath(format string, args ...any) string {
	return "/" + StaticSys.HashName(fmt.Sprintf("public/"+format, args...))
}

func renderServerError(w http.ResponseWriter, _ *http.Request, msg string) {
	log.Printf("Error: %v", msg)
	w.WriteHeader(500)
	w.Write([]byte("some error"))
	//ErrorPage().Render(r.Context(), w)
}
