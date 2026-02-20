-- +goose Up
CREATE TABLE sync_runs (
    id                    SERIAL PRIMARY KEY,
    start_time            TIMESTAMPTZ NOT NULL,
    end_time              TIMESTAMPTZ NOT NULL,
    new_organisations     INTEGER NOT NULL,
    new_licences          INTEGER NOT NULL,
    changed_licences      INTEGER NOT NULL,
    closed_organisations  INTEGER NOT NULL,
    closed_licences       INTEGER NOT NULL,
    error_count           INTEGER NOT NULL
);

-- +goose Down
DROP TABLE sync_runs;
