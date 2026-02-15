package sync

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
	NewOrganisations    int
	NewLicences         int
	ChangedLicences     int
	ClosedOrganisations int
	ClosedLicences      int
	Errors              []error
}

// CSVFetcher fetches sponsor licence records
type CSVFetcher interface {
	FetchRecords() ([]csvfetch.Record, error)
}

// OrgRepository handles organisation database operations
type OrgRepository interface {
	Find(ctx context.Context, name, townCity, county string) (database.Organisation, bool, error)
	Insert(ctx context.Context, org database.Organisation, initialRun bool) (int, error)
	Close(ctx context.Context, orgID int) error
	GetAllActive(ctx context.Context) ([]database.Organisation, error)
}

// LicenceRepository handles licence database operations
type LicenceRepository interface {
	FindActive(ctx context.Context, orgID int, licenceType, route string) (database.Licence, bool, error)
	Insert(ctx context.Context, lic database.Licence, initialRun bool) (int, error)
	Close(ctx context.Context, licenceID int) error
	GetAllActive(ctx context.Context) ([]database.Licence, error)
}

// ConfigRepository handles application config database operations
type ConfigRepository interface {
	GetValue(ctx context.Context, name, key string) (string, bool, error)
	SetValue(ctx context.Context, name, key, value string) error
	GetInitialRunTime(ctx context.Context) (string, bool, error)
}

// Syncer synchronises the database with gov.uk data
type Syncer struct {
	fetcher  CSVFetcher
	orgs     OrgRepository
	licences LicenceRepository
	config   ConfigRepository
}

// NewSyncer creates a Syncer with the given dependencies.
func NewSyncer(fetcher CSVFetcher, orgs OrgRepository, licences LicenceRepository, config ConfigRepository) *Syncer {
	return &Syncer{
		fetcher:  fetcher,
		orgs:     orgs,
		licences: licences,
		config:   config,
	}
}

// Run syncs the database with the current gov.uk CSV.
// It checks the config table to determine if this is the initial run.
func (s *Syncer) Run(ctx context.Context) (*Result, error) {
	result := &Result{}

	_, initialRunTimeHasValue, err := s.config.GetInitialRunTime(ctx)
	initialRun := !initialRunTimeHasValue
	if err != nil {
		return nil, fmt.Errorf("check initial run: %w", err)
	}
	slog.Info("sync starting", "initial_run", initialRun)

	records, err := s.fetcher.FetchRecords()
	if err != nil {
		return nil, fmt.Errorf("fetch CSV: %w", err)
	}
	slog.Info("fetched sponsor list", "count", len(records))

	seenOrgs := make(map[int]bool)
	seenLicences := make(map[int]bool)
	for _, rec := range records {
		orgID, licID, err := s.processRecord(ctx, rec, initialRun, result)
		if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}
		seenOrgs[orgID] = true
		seenLicences[licID] = true
	}

	if !initialRun {
		s.closeStale(ctx, seenOrgs, seenLicences, result)
	}

	if initialRun {
		now := time.Now().UTC().Format(time.RFC3339)
		if err := s.config.SetValue(ctx, "InitialRunDateTime", "Default", now); err != nil {
			return result, fmt.Errorf("set initial run time: %w", err)
		}
	}

	slog.Info("sync complete",
		"new_organisations", result.NewOrganisations,
		"new_licences", result.NewLicences,
		"changed_licences", result.ChangedLicences,
		"closed_organisations", result.ClosedOrganisations,
		"closed_licences", result.ClosedLicences,
		"errors", len(result.Errors),
	)

	return result, nil
}

// processRecord syncs a single CSV record. Returns the active orgID and licenceID
// for stale record detection, or an error.
func (s *Syncer) processRecord(ctx context.Context, rec csvfetch.Record, initialRun bool, result *Result) (int, int, error) {
	orgID, isNew, err := s.processOrg(ctx, rec, initialRun)
	if err != nil {
		return 0, 0, err
	}
	if isNew {
		result.NewOrganisations++
	}

	licID, outcome, err := s.processLicence(ctx, orgID, rec, initialRun)
	if err != nil {
		return 0, 0, err
	}
	switch outcome {
	case LicenceNew:
		result.NewLicences++
	case LicenceChanged:
		result.ChangedLicences++
	}
	return orgID, licID, nil
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

// processLicence syncs a single licence record. Returns the active licence ID,
// what happened (new/changed/unchanged), and any error.
func (s *Syncer) processLicence(ctx context.Context, orgID int, rec csvfetch.Record, initialRun bool) (int, LicenceResult, error) {
	lic, found, err := s.licences.FindActive(ctx, orgID, rec.LicenceType, rec.Route)
	if err != nil {
		return 0, LicenceUnchanged, fmt.Errorf("find licence: %w", err)
	}
	if !found {
		newLic := database.Licence{OrganisationID: orgID, LicenceType: rec.LicenceType, Rating: rec.Rating, Route: rec.Route}
		id, err := s.licences.Insert(ctx, newLic, initialRun)
		if err != nil {
			return 0, LicenceUnchanged, fmt.Errorf("insert licence: %w", err)
		}
		return id, LicenceNew, nil
	}
	if lic.Rating != rec.Rating {
		if err := s.licences.Close(ctx, lic.ID); err != nil {
			return 0, LicenceUnchanged, fmt.Errorf("close licence: %w", err)
		}
		newLic := database.Licence{OrganisationID: orgID, LicenceType: rec.LicenceType, Rating: rec.Rating, Route: rec.Route}
		id, err := s.licences.Insert(ctx, newLic, false)
		if err != nil {
			return 0, LicenceUnchanged, fmt.Errorf("insert updated licence: %w", err)
		}
		return id, LicenceChanged, nil
	}
	return lic.ID, LicenceUnchanged, nil
}

// closeStale closes organisations and licences that are active in the database
// but were not present in the CSV (i.e. removed by gov.uk).
func (s *Syncer) closeStale(ctx context.Context, seenOrgs, seenLicences map[int]bool, result *Result) {
	activeOrgs, err := s.orgs.GetAllActive(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("get active orgs: %w", err))
		return
	}
	for _, org := range activeOrgs {
		if !seenOrgs[org.ID] {
			if err := s.orgs.Close(ctx, org.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("close org %q: %w", org.Name, err))
				continue
			}
			result.ClosedOrganisations++
		}
	}

	activeLicences, err := s.licences.GetAllActive(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("get active licences: %w", err))
		return
	}
	for _, lic := range activeLicences {
		if !seenLicences[lic.ID] {
			if err := s.licences.Close(ctx, lic.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("close licence %d: %w", lic.ID, err))
				continue
			}
			result.ClosedLicences++
		}
	}
}
