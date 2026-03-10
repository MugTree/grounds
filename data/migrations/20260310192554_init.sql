-- +goose Up
-- +goose StatementBegin
CREATE TABLE buildings (
  id SERIAL PRIMARY KEY,
  name VARCHAR (50) UNIQUE NOT NULL
);
-- +goose StatementEnd

