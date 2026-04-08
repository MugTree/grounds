-- name: CreateVisit :exec
INSERT INTO visit (location_id, employee_id, notes, datetime, duration) VALUES (?, ?, ?, ?, ?);