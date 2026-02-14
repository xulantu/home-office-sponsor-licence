package database

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"sponsor-tracker/internal/config"
)

// getTestPool creates a connection pool for testing.
// Loads configuration from config.yaml and .env in the backend directory.
func getTestPool(t *testing.T) *pgxpool.Pool {
	// Find the backend directory (where config.yaml lives)
	_, thisFile, _, _ := runtime.Caller(0)
	backendDir := filepath.Join(filepath.Dir(thisFile), "..", "..")

	configPath := filepath.Join(backendDir, "config.yaml")
	envPath := filepath.Join(backendDir, ".env")

	cfg, err := config.Load(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	connString := cfg.TestDatabase.ConnectionString()
	pool, err := Connect(connString)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	return pool
}
