package database

import (
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	// Get connection string from environment variable
	// This avoids hardcoding passwords in code
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		t.Skip("DATABASE_URL not set, skipping database test")
	}

	pool, err := Connect(connString)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer pool.Close()

	t.Log("Successfully connected to database")
}
