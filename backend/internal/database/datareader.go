package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DataResponse holds a paginated view of the current application state.
type DataResponse struct {
	InitialRunTime     string         `json:"initial_run_time"`
	TotalOrganisations int            `json:"total_organisations"`
	From               int            `json:"from"`
	To                 int            `json:"to"`
	Organisations      []Organisation `json:"organisations"`
	Licences           []Licence      `json:"licences"`
}

// PostgresDataReader provides read-only access to the current application state.
type PostgresDataReader struct {
	pool *pgxpool.Pool
}

func NewPostgresDataReader(pool *pgxpool.Pool) *PostgresDataReader {
	return &PostgresDataReader{pool: pool}
}

// GetAll returns a paginated view of the data. from and to are 1-based org
// order numbers. If to == 0, all organisations and licences are returned.
func (r *PostgresDataReader) GetAll(ctx context.Context, from, to int, search string) (*DataResponse, error) {
	initialRunTime, _, err := GetInitialRunTime(ctx, r.pool)
	if err != nil {
		return nil, fmt.Errorf("get all data: %w", err)
	}

	total, err := CountAllActiveOrganisations(ctx, r.pool, search)
	if err != nil {
		return nil, fmt.Errorf("get all data: %w", err)
	}

	orgs, err := GetAllActiveOrganisations(ctx, r.pool, from, to, search)
	if err != nil {
		return nil, fmt.Errorf("get all data: %w", err)
	}

	var licences []Licence
	if len(orgs) > 0 {
		orgIDs := make([]int, len(orgs))
		for i, org := range orgs {
			orgIDs[i] = org.ID
		}
		licences, err = GetActiveLicencesByOrgIDs(ctx, r.pool, orgIDs)
		if err != nil {
			return nil, fmt.Errorf("get all data: %w", err)
		}
	}

	return &DataResponse{
		InitialRunTime:     initialRunTime,
		TotalOrganisations: total,
		From:               from,
		To:                 to,
		Organisations:      orgs,
		Licences:           licences,
	}, nil
}
