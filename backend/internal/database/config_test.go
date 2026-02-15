package database

import (
	"context"
	"testing"
)

func TestSetAndGetConfigValue(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()
	ctx := context.Background()

	pool.Exec(ctx, `DELETE FROM config WHERE name = 'TestConfig'`)

	// Insert
	err := SetConfigValue(ctx, pool, "TestConfig", "Key1", "Value1")
	if err != nil { t.Fatalf("SetConfigValue failed: %v", err) }

	value, found, err := GetConfigValue(ctx, pool, "TestConfig", "Key1")
	if err != nil { t.Fatalf("GetConfigValue failed: %v", err) }
	if !found { t.Fatal("expected config value to be found") }
	if value != "Value1" { t.Errorf("got %q, want %q", value, "Value1") }

	// Upsert (same name+key, different value)
	err = SetConfigValue(ctx, pool, "TestConfig", "Key1", "Value2")
	if err != nil { t.Fatalf("SetConfigValue upsert failed: %v", err) }

	value, found, err = GetConfigValue(ctx, pool, "TestConfig", "Key1")
	if err != nil { t.Fatalf("GetConfigValue after upsert failed: %v", err) }
	if !found { t.Fatal("expected config value to be found after upsert") }
	if value != "Value2" { t.Errorf("got %q, want %q", value, "Value2") }

	pool.Exec(ctx, `DELETE FROM config WHERE name = 'TestConfig'`)
}

func TestGetInitialRunTime_NotSet_ReturnsFalse(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()
	ctx := context.Background()

	pool.Exec(ctx, `DELETE FROM config WHERE name = 'InitialRunDateTime'`)

	_, found, err := GetInitialRunTime(ctx, pool)
	if err != nil { t.Fatalf("GetInitialRunTime failed: %v", err) }
	if found { t.Error("expected found=false when no initial run time set") }

	pool.Exec(ctx, `DELETE FROM config WHERE name = 'InitialRunDateTime'`)
}
