package app

const SaveImageSql string = `INSERT INTO images (visit_id, filename, original_name, mimetype, size, created_at) VALUES($1, $2, $3, $4, $5, CURRENT_TIMESTAMP);`

// --------------------------------------

const InsertVisitSql string = `INSERT INTO visits (location_id, employee_id, notes) VALUES ($1, $2, $3);`

// --------------------------------------

type customer struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

const SelectCustomersSql string = `SELECT * FROM customer;`

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

type visitData struct {
	CustomerName string `db:"customer_name"`
	LocationName string `db:"location_name"`
	EmployeeName string `db:"employee_name"`
}

const SelectVisitDataSql string = `SELECT c.name customer_name, l.name location_name, e.name employee_name
			FROM visits v
         		INNER JOIN location l ON v.location_id = l.id
         		INNER JOIN employee e ON e.id = v.employee_id
         		INNER JOIN customer c ON c.id = l.customer_id
			WHERE v.id = $1`

// --------------------------------------

type locationData struct {
	LocationName string `db:"location_name"`
	CustomerName string `db:"customer_name"`
	CustomerId   string `db:"customer_id"`
	LocationId   string `db:"location_id"`
}

const SelectLocationByIdSql string = `
 		SELECT
			l.name AS location_name,
			c.name AS customer_name,
			c.Id AS customer_id,
			l.id AS location_id
		FROM location l
		INNER JOIN customer c
		ON l.customer_id = c.id
		WHERE l.id = $1;`

// --------------------------------------

const SelectImagePathsSql = `SELECT filename from images where visit_id = $1;`
