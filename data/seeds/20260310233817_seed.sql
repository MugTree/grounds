-- -- +goose Up
-- -- +goose StatementBegin

-- =========================
-- CUSTOMERS (5 gardening companies)
-- =========================
INSERT INTO customer (name) VALUES
('GreenRoot Gardens'),
('Bloom & Branch Ltd'),
('MeadowLeaf Landscaping'),
('Urban Sprout Co'),
('Rosewood Horticulture');

-- =========================
-- EMPLOYEE
-- =========================
INSERT INTO employee (name) VALUES
('Alice Gardener');

-- =========================
-- LOCATIONS (4 per customer)
-- =========================

-- GreenRoot Gardens
INSERT INTO location (name, customer_id) VALUES
('North Greenhouse',
 (SELECT id FROM customer WHERE name = 'GreenRoot Gardens')),
('South Nursery',
 (SELECT id FROM customer WHERE name = 'GreenRoot Gardens')),
('Compost Yard',
 (SELECT id FROM customer WHERE name = 'GreenRoot Gardens')),
('Tool Shed',
 (SELECT id FROM customer WHERE name = 'GreenRoot Gardens'));

-- Bloom & Branch Ltd
INSERT INTO location (name, customer_id) VALUES
('Rose Tunnel',
 (SELECT id FROM customer WHERE name = 'Bloom & Branch Ltd')),
('Lavender Field',
 (SELECT id FROM customer WHERE name = 'Bloom & Branch Ltd')),
('Seed Store',
 (SELECT id FROM customer WHERE name = 'Bloom & Branch Ltd')),
('Irrigation Hub',
 (SELECT id FROM customer WHERE name = 'Bloom & Branch Ltd'));

-- MeadowLeaf Landscaping
INSERT INTO location (name, customer_id) VALUES
('Oak Courtyard',
 (SELECT id FROM customer WHERE name = 'MeadowLeaf Landscaping')),
('Fern House',
 (SELECT id FROM customer WHERE name = 'MeadowLeaf Landscaping')),
('Pond Edge',
 (SELECT id FROM customer WHERE name = 'MeadowLeaf Landscaping')),
('Mulch Depot',
 (SELECT id FROM customer WHERE name = 'MeadowLeaf Landscaping'));

-- Urban Sprout Co
INSERT INTO location (name, customer_id) VALUES
('Roof Garden',
 (SELECT id FROM customer WHERE name = 'Urban Sprout Co')),
('Vertical Planter Bay',
 (SELECT id FROM customer WHERE name = 'Urban Sprout Co')),
('Hydroponics Lab',
 (SELECT id FROM customer WHERE name = 'Urban Sprout Co')),
('Potting Bench',
 (SELECT id FROM customer WHERE name = 'Urban Sprout Co'));

-- Rosewood Horticulture
INSERT INTO location (name, customer_id) VALUES
('Cedar Grove',
 (SELECT id FROM customer WHERE name = 'Rosewood Horticulture')),
('Wildflower Meadow',
 (SELECT id FROM customer WHERE name = 'Rosewood Horticulture')),
('Propagation Room',
 (SELECT id FROM customer WHERE name = 'Rosewood Horticulture')),
('Glasshouse East',
 (SELECT id FROM customer WHERE name = 'Rosewood Horticulture'));

-- =========================
-- ONE VISIT
-- Visit Alice Gardener -> North Greenhouse
-- =========================
INSERT INTO visits (location_id, employee_id)
VALUES (
    (SELECT id FROM location WHERE name = 'North Greenhouse'
        AND customer_id = (SELECT id FROM customer WHERE name = 'GreenRoot Gardens')),
    (SELECT id FROM employee WHERE name = 'Alice Gardener')
);


-- +goose StatementEnd