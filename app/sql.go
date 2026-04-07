package app

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const SaveImageSql string = `INSERT INTO image (visit_id, filename, original_name, mimetype, size, created_at) VALUES($1, $2, $3, $4, $5, CURRENT_TIMESTAMP);`

// --------------------------------------

const InsertVisitSql string = `INSERT INTO visit (location_id, employee_id, notes, datetime, duration) VALUES ($1, $2, $3, $4, $5);`

// --------------------------------------

type customer struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

const SelectCustomersSql string = `SELECT * FROM customer;`

// --------------------------------------

// --------------------------------------
type location struct {
	Id         string `db:"id"`
	Name       string `db:"name"`
	CustomerId string `db:"customer_id"`
}

const SelectLocationsSql string = `SELECT * FROM location;`
const SelectionLocationByCustomerIdSql = `SELECT * FROM location WHERE id = ? AND customer_id = ?`

// ----------------------------------------

const SelectCustomerByIdSql string = `SELECT * FROM customer WHERE id = $1`

// --------------------------------------

// --------------------------------------

type locationData struct {
	LocationName string `db:"location_name"`
	CustomerName string `db:"customer_name"`
	CustomerId   string `db:"customer_id"`
	LocationId   string `db:"location_id"`
}

func selectLocationData(ctx context.Context, db *sqlx.DB, id string) (locationData, error) {

	ld := locationData{}

	err := db.GetContext(ctx, &ld, `
 		SELECT
			l.name AS location_name,
			c.name AS customer_name,
			c.Id AS customer_id,
			l.id AS location_id
		FROM location l
		INNER JOIN customer c
		ON l.customer_id = c.id
		WHERE l.id = $1;`, id)
	if err != nil {
		err = fmt.Errorf("sql: error getting location - %v", err)
	}

	return ld, err
}

// --------------------------------------

const SelectImagePathsSql = `SELECT filename from image where visit_id = $1;`

type visitData struct {
	CustomerName  string `db:"customer_name"`
	LocationName  string `db:"location_name"`
	EmployeeName  string `db:"employee_name"`
	VisitDuration string `db:"visit_duration"`
	VisitDateTime string `db:"visit_datetime"`
}

func selectVisitData(ctx context.Context, db *sqlx.DB, id string) (visitData, error) {
	d := visitData{}
	err := db.GetContext(ctx, &d, `SELECT 
						v.datetime AS visit_datetime,  
						v.duration AS visit_duration, 
						c.name customer_name, 
						l.name location_name, 
						e.name employee_name 
			FROM visit v
         		INNER JOIN location l ON v.location_id = l.id
         		INNER JOIN employee e ON e.id = v.employee_id
         		INNER JOIN customer c ON c.id = l.customer_id
			WHERE v.id = $1`, id)

	if err != nil {
		err = fmt.Errorf("sql: error selecting visit data: %v", err)
	}

	return d, err
}

type visitByEmployee struct {
	VisitId      string `db:"visit_id"`
	EmployeeName string `db:"employee_name"`
	LocationName string `db:"location_name"`
}

func selectVisitsByEmployee(ctx context.Context, db *sqlx.DB, id string) ([]visitByEmployee, error) {
	ve := []visitByEmployee{}

	err := db.SelectContext(ctx, &ve,
		`SELECT 
				v.id AS visit_id,  
				e.name AS employee_name, 
				l.name AS location_name
			FROM visit v
         		INNER JOIN employee e ON v.employee_id = e.id
         		INNER JOIN location l ON v.location_id = l.id
			WHERE e.id = $1;`, id)
	if err != nil {
		err = fmt.Errorf("sql: error selecting visit by employee data: %v", err)
	}
	return ve, err
}
