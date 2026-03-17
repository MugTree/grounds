-- +goose Up
-- +goose StatementBegin
ALTER TABLE visits ADD COLUMN notes TEXT DEFAULT '';
CREATE TABLE images (
    id INTEGER PRIMARY KEY,
    visit_id INTEGER NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    original_name TEXT NOT NULL,
    mimetype TEXT,
    size INTEGER,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd
