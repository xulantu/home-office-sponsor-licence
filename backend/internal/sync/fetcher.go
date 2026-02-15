package sync

import (
	"sponsor-tracker/internal/csvfetch"
)

// GovUKFetcher implements CSVFetcher by discovering and downloading from gov.uk.
type GovUKFetcher struct{}

func NewGovUKFetcher() *GovUKFetcher {
	return &GovUKFetcher{}
}

func (f *GovUKFetcher) FetchRecords() ([]csvfetch.Record, error) {
	return csvfetch.DiscoverAndFetch()
}
