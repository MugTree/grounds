-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS buildings;
CREATE TABLE buildings (
  id SERIAL PRIMARY KEY,
  name VARCHAR (50) UNIQUE NOT NULL
);
-- +goose StatementEnd

