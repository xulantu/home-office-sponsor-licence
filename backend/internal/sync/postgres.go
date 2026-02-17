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
	return database.FindActiveOrganisation(ctx, r.pool, name, townCity, county)
}

func (r *PostgresOrgRepository) Insert(ctx context.Context, org database.Organisation, initialRun bool) (int, error) {
	return database.InsertOrganisation(ctx, r.pool, org, initialRun)
}

func (r *PostgresOrgRepository) Close(ctx context.Context, orgID int) error {
	return database.CloseOrganisation(ctx, r.pool, orgID)
}

func (r *PostgresOrgRepository) GetAllActive(ctx context.Context) ([]database.Organisation, error) {
	return database.GetAllActiveOrganisations(ctx, r.pool, 1, 0, "")
}

// PostgresLicenceRepository implements LicenceRepository using PostgreSQL.
type PostgresLicenceRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresLicenceRepository(pool *pgxpool.Pool) *PostgresLicenceRepository {
	return &PostgresLicenceRepository{pool: pool}
}

func (r *PostgresLicenceRepository) FindActive(ctx context.Context, orgID int, licenceType, route string) (database.Licence, bool, error) {
	return database.FindActiveLicence(ctx, r.pool, orgID, licenceType, route)
}

func (r *PostgresLicenceRepository) Insert(ctx context.Context, lic database.Licence, initialRun bool) (int, error) {
	return database.InsertLicence(ctx, r.pool, lic, initialRun)
}

func (r *PostgresLicenceRepository) Close(ctx context.Context, licenceID int) error {
	return database.CloseLicence(ctx, r.pool, licenceID)
}

func (r *PostgresLicenceRepository) GetAllActive(ctx context.Context) ([]database.Licence, error) {
	return database.GetAllActiveLicences(ctx, r.pool)
}

// PostgresConfigRepository implements ConfigRepository using PostgreSQL.
type PostgresConfigRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresConfigRepository(pool *pgxpool.Pool) *PostgresConfigRepository {
	return &PostgresConfigRepository{pool: pool}
}

func (r *PostgresConfigRepository) GetValue(ctx context.Context, name, key string) (string, bool, error) {
	return database.GetConfigValue(ctx, r.pool, name, key)
}

func (r *PostgresConfigRepository) SetValue(ctx context.Context, name, key, value string) error {
	return database.SetConfigValue(ctx, r.pool, name, key, value)
}

func (r *PostgresConfigRepository) GetInitialRunTime(ctx context.Context) (string, bool, error) {
	return database.GetInitialRunTime(ctx, r.pool)
}

