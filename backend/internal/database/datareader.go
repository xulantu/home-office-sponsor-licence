package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DataResponse holds the complete current state of the application.
type DataResponse struct {
	InitialRunTime string         `json:"initial_run_time"`
	Organisations  []Organisation `json:"organisations"`
	Licences       []Licence      `json:"licences"`
}

// PostgresDataReader provides read-only access to the current application state.
type PostgresDataReader struct {
	pool *pgxpool.Pool
}

func NewPostgresDataReader(pool *pgxpool.Pool) *PostgresDataReader {
	return &PostgresDataReader{pool: pool}
}

func (r *PostgresDataReader) GetAll(ctx context.Context) (*DataResponse, error) {
	initialRunTime, _, err := GetInitialRunTime(ctx, r.pool)
	if err != nil {
		return nil, fmt.Errorf("get all data: %w", err)
	}

	orgs, err := GetAllActiveOrganisations(ctx, r.pool)
	if err != nil {
		return nil, fmt.Errorf("get all data: %w", err)
	}

	licences, err := GetAllActiveLicences(ctx, r.pool)
	if err != nil {
		return nil, fmt.Errorf("get all data: %w", err)
	}

	return &DataResponse{
		InitialRunTime: initialRunTime,
		Organisations:  orgs,
		Licences:       licences,
	}, nil
}
