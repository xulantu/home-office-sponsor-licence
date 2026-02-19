-- +goose Up
ALTER TABLE licences ALTER COLUMN rating TYPE VARCHAR(100);

-- +goose Down
ALTER TABLE licences ALTER COLUMN rating TYPE VARCHAR(20);
