package database

import (
	"context"
	"testing"
)

func TestEscapeLike(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "hello"},
		{"100%", `100\%`},
		{"a_b", `a\_b`},
		{`back\slash`, `back\\slash`},
		{`%_\`, `\%\_\\`},
	}
	for _, tt := range tests {
		got := escapeLike(tt.input)
		if got != tt.want {
			t.Errorf("escapeLike(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

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
	found, ok, err := FindActiveOrganisation(ctx, pool, org.Name, org.TownCity, org.County)
	if err != nil {
		t.Fatalf("FindActiveOrganisation failed: %v", err)
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

func TestGetAllActiveOrganisations_ReturnsSortedByName(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()
	ctx := context.Background()

	pool.Exec(ctx, `DELETE FROM licences`)
	pool.Exec(ctx, `DELETE FROM organisations`)

	// Insert in wrong order, plus one deleted org
	idZ, _ := InsertOrganisation(ctx, pool, Organisation{Name: "Zebra Ltd", TownCity: "London", County: ""}, false)
	idA, _ := InsertOrganisation(ctx, pool, Organisation{Name: "Acme Corp", TownCity: "Manchester", County: ""}, false)
	idD, _ := InsertOrganisation(ctx, pool, Organisation{Name: "Deleted Inc", TownCity: "Leeds", County: ""}, false)
	CloseOrganisation(ctx, pool, idD)

	orgs, err := GetAllActiveOrganisations(ctx, pool, 1, 0, "")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(orgs) != 2 { t.Fatalf("got %d orgs, want 2", len(orgs)) }
	if orgs[0].ID != idA || orgs[1].ID != idZ { t.Error("orgs not sorted by name") }

	pool.Exec(ctx, `DELETE FROM licences`)
	pool.Exec(ctx, `DELETE FROM organisations`)
}

func TestCountAllActiveOrganisations_WithSearch(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()
	ctx := context.Background()

	pool.Exec(ctx, `DELETE FROM licences`)
	pool.Exec(ctx, `DELETE FROM organisations`)

	InsertOrganisation(ctx, pool, Organisation{Name: "Acme Corp", TownCity: "London", County: ""}, false)
	InsertOrganisation(ctx, pool, Organisation{Name: "Beta Ltd", TownCity: "Acme Town", County: ""}, false)
	InsertOrganisation(ctx, pool, Organisation{Name: "Gamma Inc", TownCity: "Leeds", County: ""}, false)

	count, err := CountAllActiveOrganisations(ctx, pool, "acme")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if count != 2 { t.Errorf("got %d, want 2 (matches name and town)", count) }

	count, err = CountAllActiveOrganisations(ctx, pool, "")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if count != 3 { t.Errorf("got %d, want 3 (no filter)", count) }

	pool.Exec(ctx, `DELETE FROM licences`)
	pool.Exec(ctx, `DELETE FROM organisations`)
}

func TestGetAllActiveOrganisations_WithSearch(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()
	ctx := context.Background()

	pool.Exec(ctx, `DELETE FROM licences`)
	pool.Exec(ctx, `DELETE FROM organisations`)

	InsertOrganisation(ctx, pool, Organisation{Name: "Acme Corp", TownCity: "London", County: ""}, false)
	InsertOrganisation(ctx, pool, Organisation{Name: "Beta Ltd", TownCity: "Acme Town", County: ""}, false)
	InsertOrganisation(ctx, pool, Organisation{Name: "Gamma Inc", TownCity: "Leeds", County: ""}, false)

	orgs, err := GetAllActiveOrganisations(ctx, pool, 1, 20, "acme")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(orgs) != 2 { t.Fatalf("got %d orgs, want 2", len(orgs)) }
	if orgs[0].Name != "Acme Corp" { t.Errorf("first org = %q, want Acme Corp", orgs[0].Name) }
	if orgs[1].Name != "Beta Ltd" { t.Errorf("second org = %q, want Beta Ltd", orgs[1].Name) }

	orgs, err = GetAllActiveOrganisations(ctx, pool, 2, 2, "acme")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(orgs) != 1 { t.Fatalf("got %d orgs, want 1", len(orgs)) }
	if orgs[0].Name != "Beta Ltd" { t.Errorf("org = %q, want Beta Ltd", orgs[0].Name) }

	orgs, err = GetAllActiveOrganisations(ctx, pool, 21, 40, "acme")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(orgs) != 0 { t.Errorf("got %d orgs, want 0 (past end)", len(orgs)) }

	pool.Exec(ctx, `DELETE FROM licences`)
	pool.Exec(ctx, `DELETE FROM organisations`)
}
