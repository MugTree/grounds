-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS jobs;
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    description VARCHAR (50) UNIQUE NOT NULL,
    building_id integer NOT NULL
)
-- +goose StatementEnd
