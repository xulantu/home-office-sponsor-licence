package database

import (
	"context"
	"errors"
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
		return 0, err
	}
	return id, nil
}

// FindActiveLicence finds a current (valid_to IS NULL) licence for an org and route
func FindActiveLicence(ctx context.Context, pool *pgxpool.Pool, orgID int, route string) (Licence, bool, error) {
	var lic Licence
	err := pool.QueryRow(ctx,
		`SELECT id, organisation_id, licence_type, rating, route, valid_from, valid_to
		 FROM licences
		 WHERE organisation_id = $1
		   AND route = $2
		   AND valid_to IS NULL`,
		orgID, route,
	).Scan(&lic.ID, &lic.OrganisationID, &lic.LicenceType, &lic.Rating, &lic.Route, &lic.ValidFrom, &lic.ValidTo)

	if errors.Is(err, pgx.ErrNoRows) {
		return Licence{}, false, nil
	}
	if err != nil {
		return Licence{}, false, err
	}
	return lic, true, nil
}

// CloseLicence sets valid_to to NOW() on a licence (marks it as ended)
func CloseLicence(ctx context.Context, pool *pgxpool.Pool, licenceID int) error {
	_, err := pool.Exec(ctx,
		`UPDATE licences SET valid_to = NOW() WHERE id = $1`,
		licenceID,
	)
	return err
}

// GetLicencesForOrg retrieves all licences (including history) for an organisation
func GetLicencesForOrg(ctx context.Context, pool *pgxpool.Pool, orgID int) ([]Licence, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, organisation_id, licence_type, rating, route, valid_from, valid_to
		 FROM licences
		 WHERE organisation_id = $1
		 ORDER BY valid_from NULLS FIRST`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var licences []Licence
	for rows.Next() {
		var lic Licence
		err := rows.Scan(&lic.ID, &lic.OrganisationID, &lic.LicenceType, &lic.Rating, &lic.Route, &lic.ValidFrom, &lic.ValidTo)
		if err != nil {
			return nil, err
		}
		licences = append(licences, lic)
	}

	return licences, rows.Err()
}
