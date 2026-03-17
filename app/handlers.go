package app

import (
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

		if signals.CustomerId == 0 {
			sse.PatchElementTempl(HomePage(vm))
			return
		}

		if signals.LocationId == 0 {
			vm.ShowLocations = true
			vm.Locations = filteredLocations(locations, signals.CustomerId)
			sse.PatchElementTempl(HomePage(vm))
			return
		}

		location, err := getLocation(r.Context(), db, signals.LocationId, signals.CustomerId)
		if err != nil {
			handleLocationError(w, r, err, signals)
			return
		}

		LogInfo(fmt.Sprintf("All looking good: %v", location))
		sse.ExecuteScript(fmt.Sprintf(`window.location = "/visit/add/%v/"`, signals.LocationId))

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
