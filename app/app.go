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

var hashFS = hashfs.NewFS(staticFS)

func AppSetup(db *sqlx.DB) chi.Router {

	r := chi.NewRouter()
	r.Handle("/public/*", hashfs.FileServer(hashFS))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		buildings := []Building{}

		if err := db.Select(&buildings, `SELECT * FROM buildings;`); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting buildings - %v", err))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte("hey there" + fmt.Sprintf("building: %v", buildings[0].Name)))

	})

	return r

}

func renderServerError(w http.ResponseWriter, _ *http.Request, msg string) {
	log.Printf("Error: %v", msg)
	w.WriteHeader(500)
	w.Write([]byte("some error"))
	//ErrorPage().Render(r.Context(), w)
}
