-- name: CreateVisit :one
INSERT INTO visit (location_id, employee_id, notes, datetime, duration) VALUES (?, ?, ?, ?, ?) RETURNING id;

-- name: CreateImage :exec
INSERT INTO image (visit_id, filename, original_name, mimetype, size, created_at) VALUES(?, ?, ?, ?, ?,CURRENT_TIMESTAMP);

-- name: GetVisitsByEmployee :many
SELECT 
	v.id AS visit_id,  
	e.name AS employee_name, 
	l.name AS location_name
FROM visit v
    INNER JOIN employee e ON v.employee_id = e.id
	INNER JOIN location l ON v.location_id = l.id
WHERE e.id = ?;

-- name: GetVisitById :one
SELECT 
	v.datetime AS visit_datetime,  
	v.duration AS visit_duration, 
	c.name customer_name, 
	l.name location_name, 
	e.name employee_name 
FROM visit v
	INNER JOIN location l ON v.location_id = l.id
	INNER JOIN employee e ON e.id = v.employee_id
	INNER JOIN customer c ON c.id = l.customer_id
WHERE v.id = ?;


-- name: ListCustomers :many
SELECT * FROM customer;

-- name: GetCustomerById :one
SELECT * FROM customer WHERE id = ?;

-- name: ListLocations :many
SELECT * FROM location;

-- name: GetLocationById :one
SELECT
	l.name AS location_name,
	c.name AS customer_name,
	c.Id AS customer_id,
	l.id AS location_id
FROM location l
INNER JOIN customer c
ON l.customer_id = c.id
WHERE l.id = ?;