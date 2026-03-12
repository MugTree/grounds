package www

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

// admin features
/*
	add customer
	add location
	add employee
	show visits
	show locations
	show employees
*/

func AppSetup(db *sqlx.DB) chi.Router {
	r := chi.NewRouter()
	r.Handle("/public/*", hashfs.FileServer(StaticSys))
	r.Get("/", handleHomepage(db))
	r.Get("/customer-location/", handlePatchLocation(db))
	r.Post("/new-visit/", handleNewVisit(db))
	r.Post("/create-visit/", handleCreateVisit(db))
	return r
}

func staticPath(format string, args ...any) string {
	return "/" + StaticSys.HashName(fmt.Sprintf("public/"+format, args...))
}

func renderServerError(w http.ResponseWriter, r *http.Request, msg string) {
	log.Printf("Error: %v", msg)
	ErrorPage().Render(r.Context(), w)
}
