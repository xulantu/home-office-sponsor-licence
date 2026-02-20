package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncRun records the result of a single sync operation.
type SyncRun struct {
	ID                  int
	StartTime           time.Time
	EndTime             time.Time
	NewOrganisations    int
	NewLicences         int
	ChangedLicences     int
	ClosedOrganisations int
	ClosedLicences      int
	ErrorCount          int
}

// InsertSyncRun records a completed sync run and returns its ID.
func InsertSyncRun(ctx context.Context, pool *pgxpool.Pool, run SyncRun) (int, error) {
	var id int
	err := pool.QueryRow(ctx,
		`INSERT INTO sync_runs (start_time, end_time, new_organisations, new_licences, changed_licences, closed_organisations, closed_licences, error_count)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		run.StartTime, run.EndTime, run.NewOrganisations, run.NewLicences, run.ChangedLicences, run.ClosedOrganisations, run.ClosedLicences, run.ErrorCount,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert sync run: %w", err)
	}
	return id, nil
}
