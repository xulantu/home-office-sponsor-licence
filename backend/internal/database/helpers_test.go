package database

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// getTestPool creates a connection pool for testing.
// Fails the test if TEST_DATABASE_URL is not set.
func getTestPool(t *testing.T) *pgxpool.Pool {
	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		t.Fatal("TEST_DATABASE_URL not set")
	}

	pool, err := Connect(connString)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	return pool
}
