package sync

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"sponsor-tracker/internal/config"
	"sponsor-tracker/internal/database"
)

// getTestPool creates a connection pool to the test database.
func getTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	_, thisFile, _, _ := runtime.Caller(0)
	backendDir := filepath.Join(filepath.Dir(thisFile), "..", "..")

	cfg, err := config.Load(
		filepath.Join(backendDir, "config.yaml"),
		filepath.Join(backendDir, ".env"),
	)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	pool, err := database.Connect(cfg.TestDatabase.ConnectionString())
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return pool
}

// truncateAll clears all tables so each test starts from a clean state.
func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(),
		"TRUNCATE licences, organisations, config, sync_runs RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}
}
