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

type locationByCustomer struct {
	LocationName string `db:"location_name"`
	CustomerName string `db:"customer_name"`
	LocationId   string `db:"location_id"`
}

func handleGetLocation(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		signals := getLocSignals{}
		datastar.ReadSignals(r, &signals)
		lbc := []locationByCustomer{}

		locationsByCustomerSql :=
			`SELECT
    			l.name AS location_name,
   		 		c.name AS customer_name,
    			l.id AS location_id
			FROM location l
    		INNER JOIN customer c
    		ON l.customer_id = c.id
			WHERE c.id = $1;
			`

		if err := db.Select(&lbc, locationsByCustomerSql, signals.CustomerId); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting locations - %v", err))
			return
		}

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(GetLocationsByCustomer(lbc))
	}
}
