package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Licence represents a sponsor licence record
type Licence struct {
	ID             int
	OrganisationID int
	LicenceType    string     // "Worker" or "Temporary Worker"
	Rating         string     // "A rating" or "B rating"
	Route          string     // "Skilled Worker", etc.
	ValidFrom      *time.Time // nil = existed before tracking
	ValidTo        *time.Time // nil = still active
}

// InsertLicence adds a new licence and returns its ID.
// If initialRun is true, valid_from is NULL (existed before tracking).
// If initialRun is false, valid_from uses the database default (NOW()).
// valid_to is always NULL (licence is active when inserted).
func InsertLicence(ctx context.Context, pool *pgxpool.Pool, lic Licence, initialRun bool) (int, error) {
	var id int
	var err error

	if initialRun {
		err = pool.QueryRow(ctx,
			`INSERT INTO licences (organisation_id, licence_type, rating, route, valid_from)
			 VALUES ($1, $2, $3, $4, NULL)
			 RETURNING id`,
			lic.OrganisationID, lic.LicenceType, lic.Rating, lic.Route,
		).Scan(&id)
	} else {
		err = pool.QueryRow(ctx,
			`INSERT INTO licences (organisation_id, licence_type, rating, route)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			lic.OrganisationID, lic.LicenceType, lic.Rating, lic.Route,
		).Scan(&id)
	}

	if err != nil {
		return 0, fmt.Errorf("insert licence: %w", err)
	}
	return id, nil
}

// FindActiveLicence finds a current (valid_to IS NULL) licence for an org, licence type, and route
func FindActiveLicence(ctx context.Context, pool *pgxpool.Pool, orgID int, licenceType, route string) (Licence, bool, error) {
	var lic Licence
	err := pool.QueryRow(ctx,
		`SELECT id, organisation_id, licence_type, rating, route, valid_from, valid_to
		 FROM licences
		 WHERE organisation_id = $1
		   AND licence_type = $2
		   AND route = $3
		   AND valid_to IS NULL`,
		orgID, licenceType, route,
	).Scan(&lic.ID, &lic.OrganisationID, &lic.LicenceType, &lic.Rating, &lic.Route, &lic.ValidFrom, &lic.ValidTo)

	if errors.Is(err, pgx.ErrNoRows) {
		return Licence{}, false, nil
	}
	if err != nil {
		return Licence{}, false, fmt.Errorf("find active licence: %w", err)
	}
	return lic, true, nil
}

// CloseLicence sets valid_to to NOW() on a licence (marks it as ended)
func CloseLicence(ctx context.Context, pool *pgxpool.Pool, licenceID int) error {
	_, err := pool.Exec(ctx,
		`UPDATE licences SET valid_to = NOW() WHERE id = $1`,
		licenceID,
	)
	if err != nil {
		return fmt.Errorf("close licence: %w", err)
	}
	return nil
}

// GetAllLicencesForOrg retrieves all licences (including history) for an organisation
func GetAllLicencesForOrg(ctx context.Context, pool *pgxpool.Pool, orgID int) ([]Licence, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, organisation_id, licence_type, rating, route, valid_from, valid_to
		 FROM licences
		 WHERE organisation_id = $1
		 ORDER BY valid_from NULLS FIRST`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("get licences for org: %w", err)
	}
	defer rows.Close()

	var licences []Licence
	for rows.Next() {
		var lic Licence
		err := rows.Scan(&lic.ID, &lic.OrganisationID, &lic.LicenceType, &lic.Rating, &lic.Route, &lic.ValidFrom, &lic.ValidTo)
		if err != nil {
			return nil, fmt.Errorf("get licences for org: scan row: %w", err)
		}
		licences = append(licences, lic)
	}

	return licences, rows.Err()
}

// GetAllActiveLicences retrieves all licences that are currently active.
func GetAllActiveLicences(ctx context.Context, pool *pgxpool.Pool) ([]Licence, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, organisation_id, licence_type, rating, route, valid_from
		 FROM licences
		 WHERE valid_to IS NULL
		 ORDER BY organisation_id`,
	)
	if err != nil {
		return nil, fmt.Errorf("get all active licences: %w", err)
	}
	defer rows.Close()

	var licences []Licence
	for rows.Next() {
		var lic Licence
		err := rows.Scan(&lic.ID, &lic.OrganisationID, &lic.LicenceType, &lic.Rating, &lic.Route, &lic.ValidFrom)
		if err != nil {
			return nil, fmt.Errorf("get all active licences: scan row: %w", err)
		}
		licences = append(licences, lic)
	}
	return licences, rows.Err()
}

// GetActiveLicencesByOrgIDs retrieves active licences for the given organisation IDs.
func GetActiveLicencesByOrgIDs(ctx context.Context, pool *pgxpool.Pool, orgIDs []int) ([]Licence, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, organisation_id, licence_type, rating, route, valid_from
		 FROM licences
		 WHERE organisation_id = ANY($1)
		   AND valid_to IS NULL
		 ORDER BY organisation_id`,
		orgIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("get active licences by org IDs: %w", err)
	}
	defer rows.Close()

	var licences []Licence
	for rows.Next() {
		var lic Licence
		err := rows.Scan(&lic.ID, &lic.OrganisationID, &lic.LicenceType, &lic.Rating, &lic.Route, &lic.ValidFrom)
		if err != nil {
			return nil, fmt.Errorf("get active licences by org IDs: scan row: %w", err)
		}
		licences = append(licences, lic)
	}
	return licences, rows.Err()
}
