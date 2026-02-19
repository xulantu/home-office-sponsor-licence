package sync

import (
	"context"
	"testing"
	"time"

	"sponsor-tracker/internal/csvfetch"
)

// switchableFetcher is a mock CSVFetcher whose records can be changed between runs.
type switchableFetcher struct {
	records []csvfetch.Record
}

func (f *switchableFetcher) FetchRecords() ([]csvfetch.Record, error) {
	return f.records, nil
}

func TestIntegration_OrgMovesAndReturns(t *testing.T) {
	pool := getTestPool(t)
	defer pool.Close()
	truncateAll(t, pool)

	ctx := context.Background()
	orgs := NewPostgresOrgRepository(pool)
	licences := NewPostgresLicenceRepository(pool)
	cfg := NewPostgresConfigRepository(pool)

	fetcher := &switchableFetcher{}
	s := NewSyncer(fetcher, orgs, licences, cfg)

	// Day 1: initial run — StaffCo in Leeds
	fetcher.records = []csvfetch.Record{
		{OrganisationName: "StaffCo", TownCity: "Leeds", County: "", LicenceType: "Worker", Rating: "A rating", Route: "Skilled Worker"},
	}

	result, err := s.Run(ctx)
	if err != nil {
		t.Fatalf("day 1: %v", err)
	}
	if result.NewOrganisations != 1 {
		t.Errorf("day 1: got %d new orgs, want 1", result.NewOrganisations)
	}
	if result.NewLicences != 1 {
		t.Errorf("day 1: got %d new licences, want 1", result.NewLicences)
	}

	// Day 2: subsequent run — StaffCo moves to Newcastle
	fetcher.records = []csvfetch.Record{
		{OrganisationName: "StaffCo", TownCity: "Newcastle", County: "", LicenceType: "Worker", Rating: "A rating", Route: "Skilled Worker"},
	}

	result, err = s.Run(ctx)
	if err != nil {
		t.Fatalf("day 2: %v", err)
	}
	if result.NewOrganisations != 1 {
		t.Errorf("day 2: got %d new orgs, want 1", result.NewOrganisations)
	}
	if result.NewLicences != 1 {
		t.Errorf("day 2: got %d new licences, want 1", result.NewLicences)
	}
	if result.ClosedOrganisations != 1 {
		t.Errorf("day 2: got %d closed orgs, want 1", result.ClosedOrganisations)
	}
	if result.ClosedLicences != 1 {
		t.Errorf("day 2: got %d closed licences, want 1", result.ClosedLicences)
	}

	// Day 3: subsequent run — StaffCo returns to Leeds
	fetcher.records = []csvfetch.Record{
		{OrganisationName: "StaffCo", TownCity: "Leeds", County: "", LicenceType: "Worker", Rating: "A rating", Route: "Skilled Worker"},
	}

	result, err = s.Run(ctx)
	if err != nil {
		t.Fatalf("day 3: %v", err)
	}
	if result.NewOrganisations != 1 {
		t.Errorf("day 3: got %d new orgs, want 1", result.NewOrganisations)
	}
	if result.NewLicences != 1 {
		t.Errorf("day 3: got %d new licences, want 1", result.NewLicences)
	}
	if result.ClosedOrganisations != 1 {
		t.Errorf("day 3: got %d closed orgs, want 1", result.ClosedOrganisations)
	}
	if result.ClosedLicences != 1 {
		t.Errorf("day 3: got %d closed licences, want 1", result.ClosedLicences)
	}

	// Verify: 3 organisation rows total
	var orgCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM organisations").Scan(&orgCount)
	if err != nil {
		t.Fatalf("count orgs: %v", err)
	}
	if orgCount != 3 {
		t.Fatalf("got %d org rows, want 3", orgCount)
	}

	// Verify each org's temporal columns
	rows, err := pool.Query(ctx,
		"SELECT town_city, created_at, deleted_at FROM organisations ORDER BY id")
	if err != nil {
		t.Fatalf("query orgs: %v", err)
	}
	defer rows.Close()

	type orgRow struct {
		townCity  string
		createdAt *time.Time
		deletedAt *time.Time
	}
	var orgRows []orgRow
	for rows.Next() {
		var r orgRow
		if err := rows.Scan(&r.townCity, &r.createdAt, &r.deletedAt); err != nil {
			t.Fatalf("scan org: %v", err)
		}
		orgRows = append(orgRows, r)
	}

	// Org 1: Leeds (day 1) — created_at NULL (initial run), deleted_at set (closed day 2)
	if orgRows[0].townCity != "Leeds" {
		t.Errorf("org 1: got town %q, want Leeds", orgRows[0].townCity)
	}
	if orgRows[0].createdAt != nil {
		t.Error("org 1: created_at should be NULL (initial run)")
	}
	if orgRows[0].deletedAt == nil {
		t.Error("org 1: deleted_at should be set (closed day 2)")
	}

	// Org 2: Newcastle (day 2) — created_at set, deleted_at set (closed day 3)
	if orgRows[1].townCity != "Newcastle" {
		t.Errorf("org 2: got town %q, want Newcastle", orgRows[1].townCity)
	}
	if orgRows[1].createdAt == nil {
		t.Error("org 2: created_at should be set (subsequent run)")
	}
	if orgRows[1].deletedAt == nil {
		t.Error("org 2: deleted_at should be set (closed day 3)")
	}

	// Org 3: Leeds (day 3) — created_at set, deleted_at NULL (still active)
	if orgRows[2].townCity != "Leeds" {
		t.Errorf("org 3: got town %q, want Leeds", orgRows[2].townCity)
	}
	if orgRows[2].createdAt == nil {
		t.Error("org 3: created_at should be set (subsequent run)")
	}
	if orgRows[2].deletedAt != nil {
		t.Error("org 3: deleted_at should be NULL (still active)")
	}

	// Verify: 3 licence rows total
	var licCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM licences").Scan(&licCount)
	if err != nil {
		t.Fatalf("count licences: %v", err)
	}
	if licCount != 3 {
		t.Fatalf("got %d licence rows, want 3", licCount)
	}

	// Verify each licence's temporal columns
	licRows, err := pool.Query(ctx,
		"SELECT valid_from, valid_to FROM licences ORDER BY id")
	if err != nil {
		t.Fatalf("query licences: %v", err)
	}
	defer licRows.Close()

	type licRow struct {
		validFrom *time.Time
		validTo   *time.Time
	}
	var lics []licRow
	for licRows.Next() {
		var r licRow
		if err := licRows.Scan(&r.validFrom, &r.validTo); err != nil {
			t.Fatalf("scan licence: %v", err)
		}
		lics = append(lics, r)
	}

	// Licence 1: day 1 — valid_from NULL (initial run), valid_to set (closed day 2)
	if lics[0].validFrom != nil {
		t.Error("licence 1: valid_from should be NULL (initial run)")
	}
	if lics[0].validTo == nil {
		t.Error("licence 1: valid_to should be set (closed day 2)")
	}

	// Licence 2: day 2 — valid_from set, valid_to set (closed day 3)
	if lics[1].validFrom == nil {
		t.Error("licence 2: valid_from should be set (subsequent run)")
	}
	if lics[1].validTo == nil {
		t.Error("licence 2: valid_to should be set (closed day 3)")
	}

	// Licence 3: day 3 — valid_from set, valid_to NULL (still active)
	if lics[2].validFrom == nil {
		t.Error("licence 3: valid_from should be set (subsequent run)")
	}
	if lics[2].validTo != nil {
		t.Error("licence 3: valid_to should be NULL (still active)")
	}
}
