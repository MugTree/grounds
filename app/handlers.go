package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/goforj/godump"
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

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, fmt.Sprintf("http: multipart form error - %v", err), http.StatusBadRequest)
			return
		}

		// check the images are ok?
		// ---------------------------------------------
		files := r.MultipartForm.File["original-photos"]
		for _, header := range files {

			fmt.Println("------------------------------------")
			fmt.Println(header.Filename)

			file, err := header.Open()
			if err != nil {
				continue
			}
			defer file.Close()
		}

		date := r.FormValue("visit-date")
		duration := r.FormValue("visit-duration")
		notes := r.FormValue("visit-notes")
		locationId := r.FormValue("location-id")
		locationName := r.FormValue("location-name")
		customerName := r.FormValue("customer-name")

		godump.Dump(date, duration, notes, locationId, locationName, customerName)

		res, err := db.Exec(InsertVisitSql, locationId, 1, notes)

		/*

			1). We'll need a writable directory on the host to write the images to.
			2). Well need a process of tying all the cross table data together

			use the last_insert_id eg.

			This will return a last_insert_id that we use as the visit_id for each image added tothe images table.

			So for each image ...

			    create a filename hash
			    save to disk
			    insert into images storing the hash so we can ref it in the app
			    maybe create a thumbnail at this point as well


		*/

		if err != nil {
			renderServerError(w, r, fmt.Sprintf("sql: error updating visit table - %v", err))
			return
		}

		rows, _ := res.RowsAffected()
		if rows != 1 {
			renderServerError(w, r, fmt.Sprintf("sql rows: weird number of rows effected on visit table - %v", rows))
			return
		}
		sse := datastar.NewSSE(w, r)
		time.Sleep(3 * time.Second)
		sse.PatchElementTempl(Thanks())
		LogInfo("stage 2 finished and redirecting")
	}
}
