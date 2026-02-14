package sync

import (
	"context"
	"testing"

	"sponsor-tracker/internal/csvfetch"
	"sponsor-tracker/internal/database"
)

// mockOrgRepo implements OrgRepository for testing.
type mockOrgRepo struct {
	findFn   func(ctx context.Context, name, townCity, county string) (database.Organisation, bool, error)
	insertFn func(ctx context.Context, org database.Organisation, initialRun bool) (int, error)
}

func (m *mockOrgRepo) Find(ctx context.Context, name, townCity, county string) (database.Organisation, bool, error) {
	return m.findFn(ctx, name, townCity, county)
}

func (m *mockOrgRepo) Insert(ctx context.Context, org database.Organisation, initialRun bool) (int, error) {
	return m.insertFn(ctx, org, initialRun)
}

func TestProcessOrg_ExistingOrg_ReturnsIDAndFalse(t *testing.T) {
	orgs := &mockOrgRepo{
		findFn: func(_ context.Context, name, townCity, county string) (database.Organisation, bool, error) {
			return database.Organisation{ID: 42, Name: name, TownCity: townCity, County: county}, true, nil
		},
	}

	s := NewSyncer(nil, orgs, nil)

	rec := csvfetch.Record{
		OrganisationName: "Acme Ltd",
		TownCity:         "London",
		County:           "Greater London",
	}

	id, isNew, err := s.processOrg(context.Background(), rec, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("got id=%d, want 42", id)
	}
	if isNew {
		t.Error("got isNew=true, want false")
	}
}
