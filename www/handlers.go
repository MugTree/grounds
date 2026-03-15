package www

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfederation/datastar-go/datastar"
)

func handleHomepage(db *sqlx.DB, logger slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		customers := []Customer{}

		if err := db.Select(&customers, SelectCustomersSql); err != nil {
			renderServerError(logger, w, r, fmt.Sprintf("sql: error getting customers - %v", err))
			return
		}

		HomePage(customers).Render(r.Context(), w)
	}
}

func handlePatchLocation(db *sqlx.DB, logger slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		signals := getLocSignals{}
		datastar.ReadSignals(r, &signals)
		lbc := []locationByCustomer{}

		if err := db.Select(&lbc, SelectLocationsByCustomerIdSql, signals.CustomerId); err != nil {
			renderServerError(logger, w, r, fmt.Sprintf("sql: error getting locations by customer - %v", err))
			return
		}

		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(LocationsByCustomerForm(lbc))
	}
}

func handleNewVisit(db *sqlx.DB, logger slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		locationId := r.FormValue("location_id")

		loc := locationByCustomer{}

		if err := db.Get(&loc, SelectLocationById, locationId); err != nil {
			renderServerError(logger, w, r, fmt.Sprintf("sql: error getting location - %v", err))
			return
		}

		vm := visitVM{
			Date:         time.Now().Format("2006-01-02"),
			Duration:     "60",
			CustomerName: loc.CustomerName,
			LocationName: loc.LocationName,
			LocationId:   loc.LocationId,
		}

		NewVisitCreate(vm).Render(r.Context(), w)
	}
}

func handleCreateVisit(db *sqlx.DB, logger slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		date := r.FormValue("visit-date")
		duration := r.FormValue("visit-duration")
		notes := r.FormValue("visit-notes")
		locationId := r.FormValue("location-id")
		locationName := r.FormValue("location-name")
		customerName := r.FormValue("customer-name")
		validatedConfirmed := r.FormValue("validated-confirmed")

		var isValid = true

		if isValid && validatedConfirmed == "true" {
			res, err := db.Exec(InsertVisitSql, locationId, 1)

			if err != nil {
				renderServerError(logger, w, r, fmt.Sprintf("sql: error updating visit table - %v", err))
				return
			}

			rows, _ := res.RowsAffected()
			if rows != 1 {
				renderServerError(logger, w, r, fmt.Sprintf("sql rows: weird number of rows effected on visit table - %v", rows))
				return
			}

			http.Redirect(w, r, "/", 303)
			return
		}

		vm := visitVM{
			Date:         date,
			Duration:     duration,
			Notes:        notes,
			LocationId:   locationId,
			LocationName: locationName,
			CustomerName: customerName,
		}

		sse := datastar.NewSSE(w, r)
		sse.PatchSignals([]byte(`{"stage2": true}`))
		sse.PatchElementTempl(NewVisitConfirm(vm))

	}
}

const (
	// --------------------------------------

	InsertVisitSql = `INSERT INTO visits (location_id, employee_id) VALUES ($1, $2);`

	// --------------------------------------

	SelectCustomersSql = `SELECT * FROM customer;`

	// --------------------------------------

	SelectLocationsByCustomerIdSql = `
		SELECT 
			l.name AS location_name,
			c.name AS customer_name,
			l.id AS location_id
		FROM location l
		INNER JOIN customer c
		ON l.customer_id = c.id
		WHERE c.id = $1;`

	// --------------------------------------

	SelectLocationById = `
 		SELECT
			l.name AS location_name,
			c.name AS customer_name,
			l.id AS location_id
		FROM location l
		INNER JOIN customer c
		ON l.customer_id = c.id
		WHERE l.id = $1;`
	//----------------------------------

)

type Customer struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type Employee struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type Location struct {
	Id         int    `db:"id"`
	Name       string `db:"name"`
	CustomerId int    `db:"customer_id"`
}

type Visit struct {
	Id         int `db:"id"`
	EmployeeId int `db:"employee_id"`
	LocationId int `db:"location_id"`
}

type locationByCustomer struct {
	LocationName string `db:"location_name"`
	CustomerName string `db:"customer_name"`
	LocationId   string `db:"location_id"`
}

type getLocSignals struct {
	CustomerId string `json:"customerId"`
}

type visitVM struct {
	Date         string
	Duration     string
	Notes        string
	CustomerName string
	LocationName string
	LocationId   string
}
