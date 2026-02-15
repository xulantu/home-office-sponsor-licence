package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetConfigValue retrieves a value from the config table by name and key.
// Returns the value and true if found, or empty string and false if not.
func GetConfigValue(ctx context.Context, pool *pgxpool.Pool, name, key string) (string, bool, error) {
	var value string
	err := pool.QueryRow(ctx,
		`SELECT value FROM config WHERE name = $1 AND key = $2`,
		name, key,
	).Scan(&value)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("get config value: %w", err)
	}
	return value, true, nil
}

// GetInitialRunTime checks whether the initial sync has been performed.
// Returns the timestamp and true if it has, or empty string and false if it hasn't.
func GetInitialRunTime(ctx context.Context, pool *pgxpool.Pool) (string, bool, error) {
	value, found, err := GetConfigValue(ctx, pool, "InitialRunDateTime", "Default")
	if err != nil {
		return "", false, fmt.Errorf("get initial run time: %w", err)
	}
	if !found {
		return "", false, nil
	}
	return value, true, nil
}

// SetConfigValue inserts a new row into the config table.
func SetConfigValue(ctx context.Context, pool *pgxpool.Pool, name, key, value string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO config (name, key, value) VALUES ($1, $2, $3)`,
		name, key, value,
	)
	if err != nil {
		return fmt.Errorf("set config value: %w", err)
	}
	return nil
}
