package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Organisation represents a sponsor organisation
type Organisation struct {
	ID        int
	Name      string
	TownCity  string
	County    string
	CreatedAt *time.Time // nil = existed before tracking began
	DeletedAt *time.Time // nil = not deleted
}

// InsertOrganisation adds a new organisation and returns its ID.
// If initialRun is true, created_at is set to NULL (existed before tracking).
// If initialRun is false, created_at uses the database default (NOW()).
func InsertOrganisation(ctx context.Context, pool *pgxpool.Pool, org Organisation, initialRun bool) (int, error) {
	var id int
	var err error

	if initialRun {
		err = pool.QueryRow(ctx,
			`INSERT INTO organisations (name, town_city, county, created_at)
			 VALUES ($1, $2, $3, NULL)
			 RETURNING id`,
			org.Name, org.TownCity, org.County,
		).Scan(&id)
	} else {
		err = pool.QueryRow(ctx,
			`INSERT INTO organisations (name, town_city, county)
			 VALUES ($1, $2, $3)
			 RETURNING id`,
			org.Name, org.TownCity, org.County,
		).Scan(&id)
	}

	if err != nil {
		return 0, fmt.Errorf("insert organisation: %w", err)
	}
	return id, nil
}

// FindActiveOrganisation looks up an organisation by name, town, and county
// Returns the organisation and true if found, or empty and false if not found
func FindActiveOrganisation(ctx context.Context, pool *pgxpool.Pool, name, townCity, county string) (Organisation, bool, error) {
	var org Organisation
	err := pool.QueryRow(ctx,
		`SELECT id, name, town_city, county, created_at, deleted_at
		 FROM organisations
		 WHERE name = $1
		   AND (town_city = $2 OR (town_city IS NULL AND $2 = ''))
		   AND (county = $3 OR (county IS NULL AND $3 = ''))
		   AND deleted_at IS NULL`,
		name, townCity, county,
	).Scan(&org.ID, &org.Name, &org.TownCity, &org.County, &org.CreatedAt, &org.DeletedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Organisation{}, false, nil
	}
	if err != nil {
		return Organisation{}, false, fmt.Errorf("find organisation: %w", err)
	}
	return org, true, nil
}

// GetOrganisationByID retrieves an organisation by its ID
func GetOrganisationByID(ctx context.Context, pool *pgxpool.Pool, id int) (Organisation, error) {
	var org Organisation
	err := pool.QueryRow(ctx,
		`SELECT id, name, town_city, county, created_at, deleted_at
		 FROM organisations
		 WHERE id = $1`,
		id,
	).Scan(&org.ID, &org.Name, &org.TownCity, &org.County, &org.CreatedAt, &org.DeletedAt)

	if err != nil {
		return Organisation{}, fmt.Errorf("get organisation by id: %w", err)
	}
	return org, nil
}

// CloseOrganisation sets deleted_at to NOW() on an organisation (marks it as removed).
func CloseOrganisation(ctx context.Context, pool *pgxpool.Pool, orgID int) error {
	_, err := pool.Exec(ctx,
		`UPDATE organisations SET deleted_at = NOW() WHERE id = $1`,
		orgID,
	)
	if err != nil {
		return fmt.Errorf("close organisation: %w", err)
	}
	return nil
}

// CountAllActiveOrganisations returns the total number of active organisations.
func CountAllActiveOrganisations(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var count int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM organisations WHERE deleted_at IS NULL`,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active organisations: %w", err)
	}
	return count, nil
}

// GetAllActiveOrganisations retrieves active organisations, optionally paginated.
// from and to are 1-based order numbers. If to == 0, all organisations are returned.
func GetAllActiveOrganisations(ctx context.Context, pool *pgxpool.Pool, from, to int) ([]Organisation, error) {
	query := `SELECT id, name, town_city, county, created_at
		 FROM organisations
		 WHERE deleted_at IS NULL
		 ORDER BY name`

	var rows pgx.Rows
	var err error
	if to == 0 {
		rows, err = pool.Query(ctx, query)
	} else {
		query += ` OFFSET $1 LIMIT $2`
		rows, err = pool.Query(ctx, query, from-1, to-from+1)
	}
	if err != nil {
		return nil, fmt.Errorf("get all active organisations: %w", err)
	}
	defer rows.Close()

	var orgs []Organisation
	for rows.Next() {
		var org Organisation
		err := rows.Scan(&org.ID, &org.Name, &org.TownCity, &org.County, &org.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("get all active organisations: scan row: %w", err)
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}
