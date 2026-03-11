package www

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/starfederation/datastar-go/datastar"
)

func handleHomepage(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		customers := []Customer{}

		if err := db.Select(&customers, `SELECT * FROM customer;`); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting customers - %v", err))
			return
		}

		HomePage(customers).Render(r.Context(), w)
	}
}

type getLocSignals struct {
	CustomerId string `json:"customerId"`
}

func handleGetLocation(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		signals := getLocSignals{}
		datastar.ReadSignals(r, &signals)
		locations := []Location{}

		if err := db.Select(&locations, `SELECT * FROM location WHERE customer_id = $1;`, signals.CustomerId); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting locations - %v", err))
			return
		}

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(GetLocation(locations))
	}
}
