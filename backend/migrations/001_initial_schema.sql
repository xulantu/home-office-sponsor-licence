-- +goose Up

CREATE TABLE config (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value VARCHAR(500) NOT NULL,
    UNIQUE(name, key)
);

CREATE TABLE organisations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    town_city VARCHAR(255),
    county VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(name, town_city, county)
);

CREATE INDEX idx_organisations_name ON organisations(name);

CREATE TABLE licences (
    id SERIAL PRIMARY KEY,
    organisation_id INTEGER NOT NULL REFERENCES organisations(id),
    licence_type VARCHAR(50) NOT NULL,
    rating VARCHAR(20) NOT NULL,
    route VARCHAR(100) NOT NULL,
    valid_from TIMESTAMPTZ DEFAULT NOW(),
    valid_to TIMESTAMPTZ
);

CREATE INDEX idx_licences_organisation ON licences(organisation_id);
CREATE INDEX idx_licences_valid_range ON licences(valid_from, valid_to);
CREATE INDEX idx_licences_route ON licences(route);
CREATE INDEX idx_licences_current ON licences(organisation_id, licence_type, route)
    WHERE valid_to IS NULL;

-- +goose Down
DROP INDEX idx_licences_current;
DROP INDEX idx_licences_route;
DROP INDEX idx_licences_valid_range;
DROP INDEX idx_licences_organisation;
DROP TABLE licences;
DROP INDEX idx_organisations_name;
DROP TABLE organisations;
DROP TABLE config;
