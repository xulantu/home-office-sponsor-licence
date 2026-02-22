package sync

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sponsor-tracker/internal/database"
)

// PostgresDB implements DB using a pgxpool connection pool.
type PostgresDB struct {
	pool *pgxpool.Pool
}

func NewPostgresDB(pool *pgxpool.Pool) *PostgresDB {
	return &PostgresDB{pool: pool}
}

func (db *PostgresDB) Begin(ctx context.Context) (Transaction, error) {
	return db.pool.BeginTx(ctx, pgx.TxOptions{})
}

// txOrgRepo implements OrgRepository within a transaction.
type txOrgRepo struct{ q database.Querier }

func newTxOrgRepo(q database.Querier) *txOrgRepo { return &txOrgRepo{q: q} }

func (r *txOrgRepo) Find(ctx context.Context, name, townCity, county string) (database.Organisation, bool, error) {
	return database.FindActiveOrganisation(ctx, r.q, name, townCity, county)
}
func (r *txOrgRepo) Insert(ctx context.Context, org database.Organisation, initialRun bool) (int, error) {
	return database.InsertOrganisation(ctx, r.q, org, initialRun)
}
func (r *txOrgRepo) Close(ctx context.Context, orgID int) error {
	return database.CloseOrganisation(ctx, r.q, orgID)
}
func (r *txOrgRepo) GetAllActive(ctx context.Context) ([]database.Organisation, error) {
	return database.GetAllActiveOrganisationsUnfiltered(ctx, r.q)
}

// txLicenceRepo implements LicenceRepository within a transaction.
type txLicenceRepo struct{ q database.Querier }

func newTxLicenceRepo(q database.Querier) *txLicenceRepo { return &txLicenceRepo{q: q} }

func (r *txLicenceRepo) FindActive(ctx context.Context, orgID int, licenceType, route string) (database.Licence, bool, error) {
	return database.FindActiveLicence(ctx, r.q, orgID, licenceType, route)
}
func (r *txLicenceRepo) Insert(ctx context.Context, lic database.Licence, initialRun bool) (int, error) {
	return database.InsertLicence(ctx, r.q, lic, initialRun)
}
func (r *txLicenceRepo) Close(ctx context.Context, licenceID int) error {
	return database.CloseLicence(ctx, r.q, licenceID)
}
func (r *txLicenceRepo) GetAllActive(ctx context.Context) ([]database.Licence, error) {
	return database.GetAllActiveLicences(ctx, r.q)
}

// txConfigRepo implements ConfigRepository within a transaction.
type txConfigRepo struct{ q database.Querier }

func newTxConfigRepo(q database.Querier) *txConfigRepo { return &txConfigRepo{q: q} }

func (r *txConfigRepo) GetValue(ctx context.Context, name, key string) (string, bool, error) {
	return database.GetConfigValue(ctx, r.q, name, key)
}
func (r *txConfigRepo) SetValue(ctx context.Context, name, key, value string) error {
	return database.SetConfigValue(ctx, r.q, name, key, value)
}
func (r *txConfigRepo) GetInitialRunTime(ctx context.Context) (string, bool, error) {
	return database.GetInitialRunTime(ctx, r.q)
}

// txSyncRunRepo implements SyncRunRepository within a transaction.
type txSyncRunRepo struct{ q database.Querier }

func newTxSyncRunRepo(q database.Querier) *txSyncRunRepo { return &txSyncRunRepo{q: q} }

func (r *txSyncRunRepo) Insert(ctx context.Context, run database.SyncRun) (int, error) {
	return database.InsertSyncRun(ctx, r.q, run)
}
