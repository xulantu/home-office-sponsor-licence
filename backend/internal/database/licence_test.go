package database

import (
	"context"
	"testing"
)

func TestInsertAndFindLicence(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()

	ctx := context.Background()

	// Clean up first
	pool.Exec(ctx, `DELETE FROM licences WHERE organisation_id IN
		(SELECT id FROM organisations WHERE name = 'Test Licence Corp')`)
	pool.Exec(ctx, `DELETE FROM organisations WHERE name = 'Test Licence Corp'`)

	// Create an organisation first
	org := Organisation{
		Name:     "Test Licence Corp",
		TownCity: "Manchester",
		County:   "",
	}
	orgID, err := InsertOrganisation(ctx, pool, org, false)
	if err != nil {
		t.Fatalf("InsertOrganisation failed: %v", err)
	}

	// Insert a licence (not initial run, so valid_from = NOW())
	lic := Licence{
		OrganisationID: orgID,
		LicenceType:    "Worker",
		Rating:         "A rating",
		Route:          "Skilled Worker",
	}
	licID, err := InsertLicence(ctx, pool, lic, false)
	if err != nil {
		t.Fatalf("InsertLicence failed: %v", err)
	}
	t.Logf("Inserted licence with ID: %d", licID)

	// Find active licence
	found, ok, err := FindActiveLicence(ctx, pool, orgID, "Worker", "Skilled Worker")
	if err != nil {
		t.Fatalf("FindActiveLicence failed: %v", err)
	}
	if !ok {
		t.Fatal("Licence not found")
	}
	if found.Rating != "A rating" {
		t.Errorf("Expected rating 'A rating', got %q", found.Rating)
	}

	// Close the licence
	err = CloseLicence(ctx, pool, licID)
	if err != nil {
		t.Fatalf("CloseLicence failed: %v", err)
	}

	// Should not find active licence anymore
	_, ok, err = FindActiveLicence(ctx, pool, orgID, "Worker", "Skilled Worker")
	if err != nil {
		t.Fatalf("FindActiveLicence failed: %v", err)
	}
	if ok {
		t.Error("Expected licence to not be found after closing")
	}

	// But should appear in history
	history, err := GetAllLicencesForOrg(ctx, pool, orgID)
	if err != nil {
		t.Fatalf("GetAllLicencesForOrg failed: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("Expected 1 licence in history, got %d", len(history))
	}

	// Clean up
	pool.Exec(ctx, `DELETE FROM licences WHERE organisation_id = $1`, orgID)
	pool.Exec(ctx, `DELETE FROM organisations WHERE id = $1`, orgID)
}

func TestGetAllActiveLicences_ExcludesClosed(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()
	ctx := context.Background()

	pool.Exec(ctx, `DELETE FROM licences`)
	pool.Exec(ctx, `DELETE FROM organisations`)

	orgID, _ := InsertOrganisation(ctx, pool, Organisation{Name: "Test Active Lic", TownCity: "London", County: ""}, false)

	activeID, _ := InsertLicence(ctx, pool, Licence{OrganisationID: orgID, LicenceType: "Worker", Rating: "A rating", Route: "Skilled Worker"}, false)
	closedID, _ := InsertLicence(ctx, pool, Licence{OrganisationID: orgID, LicenceType: "Worker", Rating: "B rating", Route: "Skilled Worker"}, false)
	CloseLicence(ctx, pool, closedID)

	licences, err := GetAllActiveLicences(ctx, pool)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(licences) != 1 { t.Fatalf("got %d licences, want 1", len(licences)) }
	if licences[0].ID != activeID { t.Errorf("got ID=%d, want %d", licences[0].ID, activeID) }

	pool.Exec(ctx, `DELETE FROM licences WHERE organisation_id = $1`, orgID)
	pool.Exec(ctx, `DELETE FROM organisations WHERE id = $1`, orgID)
}
