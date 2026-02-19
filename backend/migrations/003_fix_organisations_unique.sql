-- +goose Up
ALTER TABLE organisations DROP CONSTRAINT organisations_name_town_city_county_key;

CREATE UNIQUE INDEX idx_organisations_active_unique
    ON organisations (name, town_city, county)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX idx_organisations_active_unique;

ALTER TABLE organisations ADD CONSTRAINT organisations_name_town_city_county_key
    UNIQUE (name, town_city, county);
