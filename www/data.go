package www

var Queries = struct {
	LocationsByCustomerId string
	LocationById          string
}{
	// --------------------------------------
	LocationsByCustomerId: `
		SELECT 
			l.name AS location_name,
			c.name AS customer_name,
			l.id AS location_id
		FROM location l
		INNER JOIN customer c
		ON l.customer_id = c.id
		WHERE c.id = $1;`,
	// --------------------------------------
	LocationById: `
 		SELECT
			l.name AS location_name,
			c.name AS customer_name,
			l.id AS location_id
		FROM location l
		INNER JOIN customer c
		ON l.customer_id = c.id
		WHERE l.id = $1;`,
}
