-- +goose Up
-- +goose StatementBegin

-- =========================
-- EMPLOYEES
-- =========================
INSERT INTO employee (id, name) VALUES
(1, 'matt'),
(2, 'joe')
ON CONFLICT (name) DO NOTHING;

-- =========================
-- CUSTOMERS
-- =========================
INSERT INTO customer (id, name) VALUES
(1, 'Acme Corp'),
(2, 'Globex Ltd'),
(3, 'Soylent Industries'),
(4, 'Initech')
ON CONFLICT (name) DO NOTHING;

-- =========================
-- LOCATIONS 
-- =========================
INSERT INTO location (name, customer_id)
SELECT 'HQ', id FROM customer WHERE name = 'Acme Corp'
ON CONFLICT (customer_id, name) DO NOTHING;

-- =========================
-- VISITS 
-- =========================
INSERT INTO visits (location_id, employee_id)
SELECT l.id, e.id
FROM location l
JOIN employee e ON e.name = 'matt'
WHERE l.name = 'HQ';

-- +goose StatementEnd