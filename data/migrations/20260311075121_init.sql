-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = ON;

-- =========================
-- CUSTOMER
-- =========================
CREATE TABLE IF NOT EXISTS customer (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- =========================
-- EMPLOYEE
-- =========================
CREATE TABLE IF NOT EXISTS employee (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- =========================
-- LOCATION
-- A customer can have many locations
-- =========================
CREATE TABLE IF NOT EXISTS location (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    customer_id INTEGER NOT NULL,
    FOREIGN KEY (customer_id)
        REFERENCES customer(id)
        ON DELETE CASCADE,
    UNIQUE(customer_id, name)
);

-- =========================
-- VISITS
-- Each visit has one location and one employee
-- =========================
CREATE TABLE IF NOT EXISTS visits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_id INTEGER NOT NULL,
    employee_id INTEGER NOT NULL,
    FOREIGN KEY (location_id)
        REFERENCES location(id)
        ON DELETE CASCADE,
    FOREIGN KEY (employee_id)
        REFERENCES employee(id)
        ON DELETE RESTRICT
);

-- +goose StatementEnd


