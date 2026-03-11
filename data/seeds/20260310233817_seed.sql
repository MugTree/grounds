-- +goose Up
-- +goose StatementBegin
DELETE FROM buildings;
INSERT INTO buildings (id,name) VALUES (1,'Red house');
INSERT INTO buildings (id, name) VALUES (2,'Garage');
INSERT INTO buildings (id, name) VALUES (3,'Fish pond cottage');
INSERT INTO buildings (id, name) VALUES (4,'Terrace');

DELETE FROM jobs;
INSERT INTO jobs (description, building_id) VALUES ('Paint required', 3);
INSERT INTO jobs (description, building_id) VALUES ('Fence required', 1);
INSERT INTO jobs (description, building_id) VALUES ('Wall fix', 4);
INSERT INTO jobs (description, building_id) VALUES ('Lawn mowed', 2);
-- +goose StatementEnd