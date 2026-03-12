package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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

// GetRecentSyncRuns returns the most recent sync runs, ordered by start time descending.
func GetRecentSyncRuns(ctx context.Context, q Querier, limit int) ([]SyncRun, error) {
	rows, err := q.Query(ctx,
		`SELECT id, start_time, end_time, new_organisations, new_licences, changed_licences, closed_organisations, closed_licences, error_count
		 FROM sync_runs
		 ORDER BY start_time DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get recent sync runs: %w", err)
	}
	defer rows.Close()

	var runs []SyncRun
	for rows.Next() {
		var r SyncRun
		if err := rows.Scan(&r.ID, &r.StartTime, &r.EndTime, &r.NewOrganisations, &r.NewLicences, &r.ChangedLicences, &r.ClosedOrganisations, &r.ClosedLicences, &r.ErrorCount); err != nil {
			return nil, fmt.Errorf("scan sync run: %w", err)
		}
		runs = append(runs, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sync runs: %w", err)
	}
	return runs, nil
}

// GetLatestSyncRun returns the most recent sync run, or false if none exist.
func GetLatestSyncRun(ctx context.Context, q Querier) (SyncRun, bool, error) {
	var r SyncRun
	err := q.QueryRow(ctx,
		`SELECT id, start_time, end_time, new_organisations, new_licences, changed_licences, closed_organisations, closed_licences, error_count
		 FROM sync_runs
		 ORDER BY start_time DESC
		 LIMIT 1`,
	).Scan(&r.ID, &r.StartTime, &r.EndTime, &r.NewOrganisations, &r.NewLicences, &r.ChangedLicences, &r.ClosedOrganisations, &r.ClosedLicences, &r.ErrorCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SyncRun{}, false, nil
		}
		return SyncRun{}, false, fmt.Errorf("get latest sync run: %w", err)
	}
	return r, true, nil
}

// DeleteSyncRun removes a sync run by ID.
func DeleteSyncRun(ctx context.Context, q Querier, id int) error {
	_, err := q.Exec(ctx, `DELETE FROM sync_runs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete sync run: %w", err)
	}
	return nil
}

// InsertSyncRun records a completed sync run and returns its ID.
func InsertSyncRun(ctx context.Context, q Querier, run SyncRun) (int, error) {
	var id int
	err := q.QueryRow(ctx,
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
