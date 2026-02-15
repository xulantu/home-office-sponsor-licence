package sync

import (
	"context"
	"testing"

	"sponsor-tracker/internal/csvfetch"
	"sponsor-tracker/internal/database"
)

// mockOrgRepo implements OrgRepository for testing.
type mockOrgRepo struct {
	findFn          func(ctx context.Context, name, townCity, county string) (database.Organisation, bool, error)
	insertFn        func(ctx context.Context, org database.Organisation, initialRun bool) (int, error)
	closeFn         func(ctx context.Context, orgID int) error
	getAllActiveFn   func(ctx context.Context) ([]database.Organisation, error)
}

func (m *mockOrgRepo) Find(ctx context.Context, name, townCity, county string) (database.Organisation, bool, error) {
	return m.findFn(ctx, name, townCity, county)
}

func (m *mockOrgRepo) Insert(ctx context.Context, org database.Organisation, initialRun bool) (int, error) {
	return m.insertFn(ctx, org, initialRun)
}

func (m *mockOrgRepo) Close(ctx context.Context, orgID int) error {
	return m.closeFn(ctx, orgID)
}

func (m *mockOrgRepo) GetAllActive(ctx context.Context) ([]database.Organisation, error) {
	return m.getAllActiveFn(ctx)
}

// mockLicenceRepo implements LicenceRepository for testing.
type mockLicenceRepo struct {
	findActiveFn   func(ctx context.Context, orgID int, licenceType, route string) (database.Licence, bool, error)
	insertFn       func(ctx context.Context, lic database.Licence, initialRun bool) (int, error)
	closeFn        func(ctx context.Context, licenceID int) error
	getAllActiveFn func(ctx context.Context) ([]database.Licence, error)
}

func (m *mockLicenceRepo) FindActive(ctx context.Context, orgID int, licenceType, route string) (database.Licence, bool, error) {
	return m.findActiveFn(ctx, orgID, licenceType, route)
}

func (m *mockLicenceRepo) Insert(ctx context.Context, lic database.Licence, initialRun bool) (int, error) {
	return m.insertFn(ctx, lic, initialRun)
}

func (m *mockLicenceRepo) Close(ctx context.Context, licenceID int) error {
	return m.closeFn(ctx, licenceID)
}

func (m *mockLicenceRepo) GetAllActive(ctx context.Context) ([]database.Licence, error) {
	return m.getAllActiveFn(ctx)
}

// mockCSVFetcher implements CSVFetcher for testing.
type mockCSVFetcher struct {
	fetchFn func() ([]csvfetch.Record, error)
}

func (m *mockCSVFetcher) FetchRecords() ([]csvfetch.Record, error) {
	return m.fetchFn()
}

// mockConfigRepo implements ConfigRepository for testing.
type mockConfigRepo struct {
	getValueFn func(ctx context.Context, name, key string) (string, bool, error)
	setValueFn func(ctx context.Context, name, key, value string) error
}

func (m *mockConfigRepo) GetValue(ctx context.Context, name, key string) (string, bool, error) {
	return m.getValueFn(ctx, name, key)
}

func (m *mockConfigRepo) SetValue(ctx context.Context, name, key, value string) error {
	return m.setValueFn(ctx, name, key, value)
}

func TestProcessOrg_ExistingOrg_ReturnsIDAndFalse(t *testing.T) {
	orgs := &mockOrgRepo{
		findFn: func(_ context.Context, name, townCity, county string) (database.Organisation, bool, error) {
			if name != "Acme Ltd" || townCity != "London" || county != "Greater London" { t.Fatal("find wrong args") }
			return database.Organisation{ID: 42, Name: name, TownCity: townCity, County: county}, true, nil
		},
	}

	s := NewSyncer(nil, orgs, nil, nil)
	rec := csvfetch.Record{OrganisationName: "Acme Ltd", TownCity: "London", County: "Greater London"}

	id, isNew, err := s.processOrg(context.Background(), rec, false)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if id != 42 { t.Errorf("got id=%d, want 42", id) }
	if isNew { t.Error("got isNew=true, want false") }
}

func TestProcessOrg_NewOrg_InsertsAndReturnsID(t *testing.T) {
	orgs := &mockOrgRepo{
		findFn: func(_ context.Context, name, townCity, county string) (database.Organisation, bool, error) {
			if name != "New Corp" || townCity != "Manchester" || county != "Greater Manchester" { t.Fatal("find wrong args") }
			return database.Organisation{}, false, nil
		},
		insertFn: func(_ context.Context, org database.Organisation, _ bool) (int, error) {
			if org.Name != "New Corp" || org.TownCity != "Manchester" { t.Fatal("insert wrong args") }
			return 7, nil
		},
	}

	s := NewSyncer(nil, orgs, nil, nil)
	rec := csvfetch.Record{OrganisationName: "New Corp", TownCity: "Manchester", County: "Greater Manchester"}

	id, isNew, err := s.processOrg(context.Background(), rec, false)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if id != 7 { t.Errorf("got id=%d, want 7", id) }
	if !isNew { t.Error("got isNew=false, want true") }
}

func TestProcessLicence_NewLicence_ReturnsLicenceNew(t *testing.T) {
	licences := &mockLicenceRepo{
		findActiveFn: func(_ context.Context, orgID int, licenceType, route string) (database.Licence, bool, error) {
			if orgID != 42 || licenceType != "Worker" || route != "Skilled Worker" { t.Fatal("findActive wrong args") }
			return database.Licence{}, false, nil
		},
		insertFn: func(_ context.Context, lic database.Licence, _ bool) (int, error) {
			if lic.OrganisationID != 42 || lic.Rating != "A rating" || lic.LicenceType != "Worker" { t.Fatal("insert wrong args") }
			return 1, nil
		},
	}

	s := NewSyncer(nil, nil, licences, nil)
	rec := csvfetch.Record{LicenceType: "Worker", Rating: "A rating", Route: "Skilled Worker"}

	id, result, err := s.processLicence(context.Background(), 42, rec, false)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if id != 1 { t.Errorf("got id=%d, want 1", id) }
	if result != LicenceNew { t.Errorf("got result=%d, want LicenceNew", result) }
}

func TestProcessLicence_Unchanged_ReturnsLicenceUnchanged(t *testing.T) {
	licences := &mockLicenceRepo{
		findActiveFn: func(_ context.Context, orgID int, licenceType, route string) (database.Licence, bool, error) {
			if orgID != 42 || licenceType != "Worker" || route != "Skilled Worker" { t.Fatal("findActive wrong args") }
			return database.Licence{ID: 10, LicenceType: "Worker", Rating: "A rating", Route: "Skilled Worker"}, true, nil
		},
	}

	s := NewSyncer(nil, nil, licences, nil)
	rec := csvfetch.Record{LicenceType: "Worker", Rating: "A rating", Route: "Skilled Worker"}

	id, result, err := s.processLicence(context.Background(), 42, rec, false)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if id != 10 { t.Errorf("got id=%d, want 10", id) }
	if result != LicenceUnchanged { t.Errorf("got result=%d, want LicenceUnchanged", result) }
}

func TestProcessLicence_ChangedRating_ClosesAndInserts(t *testing.T) {
	closed := false
	licences := &mockLicenceRepo{
		findActiveFn: func(_ context.Context, orgID int, licenceType, route string) (database.Licence, bool, error) {
			if orgID != 42 || licenceType != "Worker" || route != "Skilled Worker" { t.Fatal("findActive wrong args") }
			return database.Licence{ID: 10, LicenceType: "Worker", Rating: "A rating", Route: "Skilled Worker"}, true, nil
		},
		closeFn: func(_ context.Context, licenceID int) error {
			if closed { t.Fatal("close called twice") }
			if licenceID != 10 { t.Fatal("close wrong licence ID") }
			closed = true
			return nil
		},
		insertFn: func(_ context.Context, lic database.Licence, _ bool) (int, error) {
			if lic.Rating != "B rating" { t.Fatal("insert wrong rating") }
			return 11, nil
		},
	}

	s := NewSyncer(nil, nil, licences, nil)
	rec := csvfetch.Record{LicenceType: "Worker", Rating: "B rating", Route: "Skilled Worker"}

	id, result, err := s.processLicence(context.Background(), 42, rec, false)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if id != 11 { t.Errorf("got id=%d, want 11", id) }
	if result != LicenceChanged { t.Errorf("got result=%d, want LicenceChanged", result) }
	if !closed { t.Error("expected close to be called") }
}
