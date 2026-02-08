package database

import (
	"testing"
)

func TestConnect(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()

	t.Log("Successfully connected to database")
}
