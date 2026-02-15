package sync

import (
	"fmt"

	"sponsor-tracker/internal/csvfetch"
)

// GovUKFetcher implements CSVFetcher by discovering and downloading from gov.uk.
type GovUKFetcher struct{}

func NewGovUKFetcher() *GovUKFetcher {
	return &GovUKFetcher{}
}

func (f *GovUKFetcher) FetchRecords() ([]csvfetch.Record, error) {
	url, err := csvfetch.DiscoverCSVURL()
	if err != nil {
		return nil, fmt.Errorf("discover CSV URL: %w", err)
	}
	return csvfetch.FetchAndParse(url)
}
