package sync

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"sponsor-tracker/internal/database"
)

// PostgresOrgRepository implements OrgRepository using PostgreSQL.
type PostgresOrgRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrgRepository(pool *pgxpool.Pool) *PostgresOrgRepository {
	return &PostgresOrgRepository{pool: pool}
}

func (r *PostgresOrgRepository) Find(ctx context.Context, name, townCity, county string) (database.Organisation, bool, error) {
	return database.FindOrganisation(ctx, r.pool, name, townCity, county)
}

func (r *PostgresOrgRepository) Insert(ctx context.Context, org database.Organisation, initialRun bool) (int, error) {
	return database.InsertOrganisation(ctx, r.pool, org, initialRun)
}

// PostgresLicenceRepository implements LicenceRepository using PostgreSQL.
type PostgresLicenceRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresLicenceRepository(pool *pgxpool.Pool) *PostgresLicenceRepository {
	return &PostgresLicenceRepository{pool: pool}
}

func (r *PostgresLicenceRepository) FindActive(ctx context.Context, orgID int, route string) (database.Licence, bool, error) {
	return database.FindActiveLicence(ctx, r.pool, orgID, route)
}

func (r *PostgresLicenceRepository) Insert(ctx context.Context, lic database.Licence, initialRun bool) (int, error) {
	return database.InsertLicence(ctx, r.pool, lic, initialRun)
}

func (r *PostgresLicenceRepository) Close(ctx context.Context, licenceID int) error {
	return database.CloseLicence(ctx, r.pool, licenceID)
}
