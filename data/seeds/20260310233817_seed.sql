-- -- +goose Up
-- -- +goose StatementBegin

-- =========================
-- CUSTOMERS (5 gardening companies)
-- =========================
INSERT INTO customer (name) VALUES
('Jarrolds'),
('The Forum'),
('Expresso Cafe'),
('City Hall');

-- =========================
-- EMPLOYEE
-- =========================
INSERT INTO employee (name) VALUES
('Matt');

-- GreenRoot Gardens
INSERT INTO location (name, customer_id) VALUES
('Upstairs Cafe',
 (SELECT id FROM customer WHERE name = 'Jarrolds')),
('Makeup floor',
 (SELECT id FROM customer WHERE name = 'Jarrolds')),
('Art department',
 (SELECT id FROM customer WHERE name = 'Jarrolds')),
('Catering floor',
 (SELECT id FROM customer WHERE name = 'Jarrolds'));

-- Bloom & Branch Ltd
INSERT INTO location (name, customer_id) VALUES
('Marzanno',
 (SELECT id FROM customer WHERE name = 'The Forum')),
('Pizza Express',
 (SELECT id FROM customer WHERE name = 'The Forum')),
('Library',
 (SELECT id FROM customer WHERE name = 'The Forum')),
('BBC lobby',
 (SELECT id FROM customer WHERE name = 'The Forum'));

-- MeadowLeaf Landscaping
INSERT INTO location (name, customer_id) VALUES
('Window display',
 (SELECT id FROM customer WHERE name = 'Expresso Cafe')),
('Planters at front',
 (SELECT id FROM customer WHERE name = 'Expresso Cafe'));

-- Urban Sprout Co
INSERT INTO location (name, customer_id) VALUES
('Atrium',
 (SELECT id FROM customer WHERE name = 'City Hall')),
('Customer lobby',
 (SELECT id FROM customer WHERE name = 'City Hall'));




-- +goose StatementEnd