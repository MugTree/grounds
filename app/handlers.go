package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/starfederation/datastar-go/datastar"
)

type HomepageVm struct {
	SelectedCustomer int
	SelectedLocation int
	ShowLocations    bool
	Customers        []Customer
	Locations        []Location
	IsValid          bool
}

type homePageSignals struct {
	CustomerId int `json:"customerId"`
	LocationId int `json:"locationId"`
}

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
				vm.Locations = getCustomerLocations(locations, signals.CustomerId)
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
				sse.ExecuteScript(fmt.Sprintf(`window.location = "/new-visit/%v/"`, signals.LocationId))
			}
		}

	}
}

func getCustomerLocations(locations []Location, customerId int) []Location {
	customerLocations := []Location{}
	for _, loc := range locations {
		if loc.CustomerId == customerId {
			customerLocations = append(customerLocations, loc)
		}
	}
	return customerLocations
}

func getHomepageData(db *sqlx.DB, w http.ResponseWriter, r *http.Request) (bool, []Customer, []Location) {

	var customers []Customer
	var locations []Location

	selectData := func(data any, name string, sql string) bool {
		if err := db.SelectContext(r.Context(), data, sql); err != nil {
			renderServerError(w, r, fmt.Sprintf("db: error selecting from table '%s' - %v", name, err))
			return false
		}
		return true
	}

	ok := selectData(&customers, "customers", SelectCustomersSql)
	if !ok {
		return false, customers, locations
	}

	ok = selectData(&locations, "locations", SelectLocationsSql)
	if !ok {
		return false, customers, locations
	}

	return true, customers, locations

}

// func handlePatchLocation(db *sqlx.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		signals := getLocSignals{}
// 		datastar.ReadSignals(r, &signals)
// 		lbc := []locationByCustomer{}

// 		if err := db.Select(&lbc, SelectLocationsByCustomerIdSql, signals.CustomerId); err != nil {
// 			renderServerError(w, r, fmt.Sprintf("sql: error getting locations by customer - %v", err))
// 			return
// 		}

// 		sse := datastar.NewSSE(w, r)
// 		sse.PatchElementTempl(LocationsByCustomerForm(lbc))
// 		sse.ExecuteScript(`
// 			document.getElementById("location-form").addEventListener("submit", function(e) {
// 				 e.preventDefault();
// 				 const value = document.getElementById("location-id").value;
// 				 if (value === "") {

// 				 	alert("Please select a location");
// 				 } else {
// 				 	window.location.assign(e.target.action  + value + "/");
// 				 }
// 			});`,
// 		)
// 	}
// }

func handleNewVisit(db *sqlx.DB) http.HandlerFunc {
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

		NewVisitCreate(vm).Render(r.Context(), w)
	}
}

func handleCreateVisit(db *sqlx.DB) http.HandlerFunc {
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
				renderServerError(w, r, fmt.Sprintf("sql: error updating visit table - %v", err))
				return
			}

			rows, _ := res.RowsAffected()
			if rows != 1 {
				renderServerError(w, r, fmt.Sprintf("sql rows: weird number of rows effected on visit table - %v", rows))
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

	SelectLocationsSql = `SELECT * FROM location;`

	// ----------------------------------------

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
