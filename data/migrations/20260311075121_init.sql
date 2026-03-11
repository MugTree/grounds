-- +goose Up
-- +goose StatementBegin
BEGIN;

-- =========================
-- CUSTOMER
-- =========================
CREATE TABLE IF NOT EXISTS customer (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- =========================
-- EMPLOYEE
-- =========================
CREATE TABLE IF NOT EXISTS employee (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL 
);

-- =========================
-- LOCATION
-- A customer can have many locations
-- =========================
CREATE TABLE IF NOT EXISTS location (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    customer_id BIGINT NOT NULL,
    CONSTRAINT fk_location_customer
        FOREIGN KEY (customer_id)
        REFERENCES customer(id)
        ON DELETE CASCADE
);

-- =========================
-- VISITS
-- Each visit has one location and one employee
-- =========================
CREATE TABLE IF NOT EXISTS visits (
    id BIGSERIAL PRIMARY KEY,
    location_id BIGINT NOT NULL,
    employee_id BIGINT NOT NULL,
    CONSTRAINT fk_visits_location
        FOREIGN KEY (location_id)
        REFERENCES location(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_visits_employee
        FOREIGN KEY (employee_id)
        REFERENCES employee(id)
        ON DELETE RESTRICT
);



ALTER TABLE customer
ADD CONSTRAINT uq_customer_name UNIQUE (name);

ALTER TABLE employee
ADD CONSTRAINT uq_employee_name UNIQUE (name);

ALTER TABLE location
ADD CONSTRAINT uq_location_customer_id_name UNIQUE (customer_id, name);



COMMIT;
-- +goose StatementEnd


