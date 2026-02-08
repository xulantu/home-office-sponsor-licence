package database

import (
	"context"
	"testing"
)

func TestInsertAndFindOrganisation(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()

	ctx := context.Background()

	// Clean up test data first (in case previous test failed)
	pool.Exec(ctx, `DELETE FROM licences WHERE organisation_id IN
		(SELECT id FROM organisations WHERE name = 'Test Corp Integration')`)
	pool.Exec(ctx, `DELETE FROM organisations WHERE name = 'Test Corp Integration'`)

	// Insert
	org := Organisation{
		Name:     "Test Corp Integration",
		TownCity: "London",
		County:   "Greater London",
	}
	id, err := InsertOrganisation(ctx, pool, org, false)
	if err != nil {
		t.Fatalf("InsertOrganisation failed: %v", err)
	}
	if id == 0 {
		t.Error("Expected non-zero ID")
	}
	t.Logf("Inserted organisation with ID: %d", id)

	// Find
	found, ok, err := FindOrganisation(ctx, pool, org.Name, org.TownCity, org.County)
	if err != nil {
		t.Fatalf("FindOrganisation failed: %v", err)
	}
	if !ok {
		t.Fatal("Organisation not found")
	}
	if found.ID != id {
		t.Errorf("Expected ID %d, got %d", id, found.ID)
	}
	if found.Name != org.Name {
		t.Errorf("Expected name %q, got %q", org.Name, found.Name)
	}

	// Get by ID
	fetched, err := GetOrganisationByID(ctx, pool, id)
	if err != nil {
		t.Fatalf("GetOrganisationByID failed: %v", err)
	}
	if fetched.Name != org.Name {
		t.Errorf("Expected name %q, got %q", org.Name, fetched.Name)
	}

	// Clean up
	pool.Exec(ctx, `DELETE FROM organisations WHERE id = $1`, id)
}
