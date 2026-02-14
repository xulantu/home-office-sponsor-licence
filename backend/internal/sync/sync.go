package sync

import (
	"context"
	"fmt"
	"log/slog"

	"sponsor-tracker/internal/csvfetch"
	"sponsor-tracker/internal/database"
)

// LicenceResult indicates what happened when syncing a licence
type LicenceResult int

const (
	LicenceUnchanged LicenceResult = iota
	LicenceNew
	LicenceChanged
)

// Result holds statistics from a sync operation.
type Result struct {
	NewOrganisations int
	NewLicences      int
	ChangedLicences  int
	Errors           []error
}

// CSVFetcher fetches sponsor licence records
type CSVFetcher interface {
	FetchRecords() ([]csvfetch.Record, error)
}

// OrgRepository handles organisation database operations
type OrgRepository interface {
	Find(ctx context.Context, name, townCity, county string) (database.Organisation, bool, error)
	Insert(ctx context.Context, org database.Organisation, initialRun bool) (int, error)
}

// LicenceRepository handles licence database operations
type LicenceRepository interface {
	FindActive(ctx context.Context, orgID int, route string) (database.Licence, bool, error)
	Insert(ctx context.Context, lic database.Licence, initialRun bool) (int, error)
	Close(ctx context.Context, licenceID int) error
}

// Syncer synchronises the database with gov.uk data
type Syncer struct {
	fetcher  CSVFetcher
	orgs     OrgRepository
	licences LicenceRepository
}

// NewSyncer creates a Syncer with the given dependencies.
func NewSyncer(fetcher CSVFetcher, orgs OrgRepository, licences LicenceRepository) *Syncer {
	return &Syncer{
		fetcher:  fetcher,
		orgs:     orgs,
		licences: licences,
	}
}

// Run syncs the database with the current gov.uk CSV.
// If initialRun is true, created_at and valid_from will be NULL.
func (s *Syncer) Run(ctx context.Context, initialRun bool) (*Result, error) {
	result := &Result{}

	records, err := s.fetcher.FetchRecords()
	if err != nil {
		return nil, fmt.Errorf("fetch CSV: %w", err)
	}
	slog.Info("fetched sponsor list", "count", len(records))

	for _, rec := range records {
		s.processRecord(ctx, rec, initialRun, result)
	}
	return result, nil
}

func (s *Syncer) processRecord(ctx context.Context, rec csvfetch.Record, initialRun bool, result *Result) {
	orgID, isNew, err := s.processOrg(ctx, rec, initialRun)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return
	}
	if isNew {
		result.NewOrganisations++
	}

	outcome, err := s.processLicence(ctx, orgID, rec, initialRun)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return
	}
	switch outcome {
	case LicenceNew:
		result.NewLicences++
	case LicenceChanged:
		result.ChangedLicences++
	}
}

func (s *Syncer) processOrg(ctx context.Context, rec csvfetch.Record, initialRun bool) (int, bool, error) {
	org, found, err := s.orgs.Find(ctx, rec.OrganisationName, rec.TownCity, rec.County)
	if err != nil {
		return 0, false, fmt.Errorf("find org %q: %w", rec.OrganisationName, err)
	}
	if found {
		return org.ID, false, nil
	}
	newOrg := database.Organisation{Name: rec.OrganisationName, TownCity: rec.TownCity, County: rec.County}
	id, err := s.orgs.Insert(ctx, newOrg, initialRun)
	if err != nil {
		return 0, false, fmt.Errorf("insert org %q: %w", rec.OrganisationName, err)
	}
	return id, true, nil
}

func (s *Syncer) processLicence(ctx context.Context, orgID int, rec csvfetch.Record, initialRun bool) (LicenceResult, error) {
	lic, found, err := s.licences.FindActive(ctx, orgID, rec.Route)
	if err != nil {
		return LicenceUnchanged, fmt.Errorf("find licence: %w", err)
	}
	if !found {
		newLic := database.Licence{OrganisationID: orgID, LicenceType: rec.LicenceType, Rating: rec.Rating, Route: rec.Route}
		_, err = s.licences.Insert(ctx, newLic, initialRun)
		if err != nil {
			return LicenceUnchanged, fmt.Errorf("insert licence: %w", err)
		}
		return LicenceNew, nil
	}
	if lic.Rating != rec.Rating || lic.LicenceType != rec.LicenceType {
		if err := s.licences.Close(ctx, lic.ID); err != nil {
			return LicenceUnchanged, fmt.Errorf("close licence: %w", err)
		}
		newLic := database.Licence{OrganisationID: orgID, LicenceType: rec.LicenceType, Rating: rec.Rating, Route: rec.Route}
		if _, err := s.licences.Insert(ctx, newLic, false); err != nil {
			return LicenceUnchanged, fmt.Errorf("insert updated licence: %w", err)
		}
		return LicenceChanged, nil
	}
	return LicenceUnchanged, nil
}
