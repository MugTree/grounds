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
			renderServerError(w, r, fmt.Sprintf("sql: error getting locations by customer - %v", err))
			return
		}

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(GetLocationsByCustomer(lbc))
	}
}

func handleNewVisitCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		locationId := r.URL.Query().Get("location")

		loc := locationByCustomer{}

		locationSql := `
			SELECT 
				l.name AS location_name, 
				l.id AS location_id, 
				c.name customer_name 
			FROM location l 
			INNER JOIN 
				customer c ON c.id = l.customer_id 
			WHERE l.id = $1;`

		if err := db.Get(&loc, locationSql, locationId); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting location - %v", err))
			return
		}

		NewVisitCreate(loc).Render(r.Context(), w)
	}
}

type visitConfirmSignals struct {
	LocationId string `json:"locationId"`
	Notes      string `json:"notes"`
}

func handleNewVisitConfirm(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vcs := visitConfirmSignals{}
		datastar.ReadSignals(r, &vcs)

		res, err := db.Exec(`INSERT INTO visits (location_id, employee_id) VALUES ($1, $2);`, vcs.LocationId, 1)
		if err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error updating visit table - %v", err))
			return
		}

		rows, _ := res.RowsAffected()
		if rows != 1 {
			renderServerError(w, r, fmt.Sprintf("sql rows: weird number of rows effected on visit table - %v", rows))
			return
		}

		w.WriteHeader(200)
		fmt.Fprintf(w, "location id: %v - notes: %v", vcs.LocationId, vcs.Notes)

	}
}
