package app

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
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
