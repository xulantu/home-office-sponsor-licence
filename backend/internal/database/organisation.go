package database

import (
	"context"
	"errors"
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
		return 0, err
	}
	return id, nil
}

// FindOrganisation looks up an organisation by name, town, and county
// Returns the organisation and true if found, or empty and false if not found
func FindOrganisation(ctx context.Context, pool *pgxpool.Pool, name, townCity, county string) (Organisation, bool, error) {
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
		return Organisation{}, false, err
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
		return Organisation{}, err
	}
	return org, nil
}
