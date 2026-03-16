package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfederation/datastar-go/datastar"
)

func handleHomepageGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, customers, _ := getHomepageData(db, w, r)
		if !ok {
			return
		}

		vm := HomepageVm{Customers: customers}
		HomePage(vm).Render(r.Context(), w)
	}
}

func handleHomepagePost(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ok, customers, locations := getHomepageData(db, w, r)
		if !ok {
			return
		}

		signals := homePageSignals{}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			renderServerError(w, r, fmt.Sprintf("http: error reading signals - %v", err))
			return
		}

		sse := datastar.NewSSE(w, r)
		vm := HomepageVm{Customers: customers}

		switch signals.CustomerId {
		case 0:
			sse.PatchElementTempl(HomePage(vm))
			return
		default:

			switch signals.LocationId {
			case 0:

				getLocations := func(locations []Location, customerId int) []Location {
					customerLocations := []Location{}
					for _, loc := range locations {
						if loc.CustomerId == customerId {
							customerLocations = append(customerLocations, loc)
						}
					}
					return customerLocations
				}

				vm.Locations = getLocations(locations, signals.CustomerId)
				vm.ShowLocations = true
				sse.PatchElementTempl(HomePage(vm))
			default:

				var location Location
				err := db.GetContext(r.Context(), &location,
					"SELECT * FROM location WHERE id = ? AND customer_id = ?",
					signals.LocationId, signals.CustomerId,
				)

				if err != nil {
					msg := ""
					if err == sql.ErrNoRows {
						msg = fmt.Sprintf("sql: error selecting location - check inputs - %v - %v",
							signals.LocationId, signals.CustomerId)
					} else {
						msg = fmt.Sprintf("http: error selecting location - check inputs - %v - %v",
							signals.LocationId, signals.CustomerId)
					}
					renderServerError(w, r, msg)
					return
				}

				LogInfo(fmt.Sprintf("All looking good: %v", location))
				sse.ExecuteScript(fmt.Sprintf(`window.location = "/visit/%v/"`, signals.LocationId))
			}
		}

	}
}

func handleVisitGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		locationId := r.PathValue("location_id")

		loc := locationByCustomer{}

		if err := db.Get(&loc, SelectLocationById, locationId); err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error getting location - %v", err))
			return
		}

		vm := visitVM{
			Date:         time.Now().Format("2006-01-02"),
			Duration:     "60",
			CustomerName: loc.CustomerName,
			LocationName: loc.LocationName,
			LocationId:   loc.LocationId,
		}

		LogInfo("starting stage 1")

		VisitStage1(vm).Render(r.Context(), w)
	}
}

func handleVisitPost(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		date := r.FormValue("visit-date")
		duration := r.FormValue("visit-duration")
		notes := r.FormValue("visit-notes")
		locationId := r.FormValue("location-id")
		locationName := r.FormValue("location-name")
		customerName := r.FormValue("customer-name")
		validatedConfirmed := r.FormValue("validated-confirmed")

		var isValid = true

		sse := datastar.NewSSE(w, r)
		if isValid && validatedConfirmed == "true" {
			res, err := db.Exec(InsertVisitSql, locationId, 1)

			if err != nil {
				renderServerError(w, r, fmt.Sprintf("sql: error updating visit table - %v", err))
				return
			}

			rows, _ := res.RowsAffected()
			if rows != 1 {
				renderServerError(w, r, fmt.Sprintf("sql rows: weird number of rows effected on visit table - %v", rows))
				return
			}

			LogInfo("stage 2 finished and redirecting")

			sse.ExecuteScript(`window.location = "/"`)
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

		LogInfo("stage 2 ...")

		sse.PatchSignals([]byte(`{"stage2": true}`))
		sse.PatchElementTempl(VisitStage2(vm))

	}
}
