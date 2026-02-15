-- Schema for UK Sponsor Licence Tracker
-- PostgreSQL with Temporal Approach

-- Application configuration / metadata
CREATE TABLE config (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value VARCHAR(500) NOT NULL,
    UNIQUE(name, key)
);

-- Organisations (sponsors)
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

-- Licences with temporal versioning
CREATE TABLE licences (
    id SERIAL PRIMARY KEY,
    organisation_id INTEGER NOT NULL REFERENCES organisations(id),
    licence_type VARCHAR(50) NOT NULL,      -- 'Worker' or 'Temporary Worker'
    rating VARCHAR(20) NOT NULL,            -- 'A rating', 'B rating'
    route VARCHAR(100) NOT NULL,            -- 'Skilled Worker', etc.
    valid_from TIMESTAMPTZ DEFAULT NOW(),   -- NULL = existed before tracking began
    valid_to TIMESTAMPTZ                    -- NULL = still active
);

CREATE INDEX idx_licences_organisation ON licences(organisation_id);
CREATE INDEX idx_licences_valid_range ON licences(valid_from, valid_to);
CREATE INDEX idx_licences_route ON licences(route);

CREATE INDEX idx_licences_current ON licences(organisation_id, licence_type, route)
    WHERE valid_to IS NULL;

-- ============================================
-- Example Queries
-- ============================================

-- Current active licences:
-- SELECT * FROM licences WHERE valid_to IS NULL;

-- State on a specific date:
-- SELECT * FROM licences
-- WHERE (valid_from IS NULL OR valid_from <= '2024-06-01')
--   AND (valid_to IS NULL OR valid_to > '2024-06-01');

-- Full history for an organisation:
-- SELECT * FROM licences
-- WHERE organisation_id = 123
-- ORDER BY valid_from NULLS FIRST;

-- Licences added on a specific date:
-- SELECT * FROM licences WHERE valid_from = '2024-06-01';

-- Licences removed on a specific date:
-- SELECT * FROM licences WHERE valid_to = '2024-06-01';
