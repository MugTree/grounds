package www

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func handleHomepage(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		buildings := []Building{}

		if err := db.Select(&buildings, `SELECT * FROM buildings;`); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting buildings - %v", err))
			return
		}

		HomePage(buildings).Render(r.Context(), w)
	}
}

func handleJobsPage(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		bid := r.PathValue("id")

		jobs := []Jobs{}

		if err := db.Select(&jobs, `SELECT * FROM jobs WHERE building_id = $1;`, bid); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting jobs - %v", err))
			return
		}

		JobsPage(jobs).Render(r.Context(), w)

	}
}
