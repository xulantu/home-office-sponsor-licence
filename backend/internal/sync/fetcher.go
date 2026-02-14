package sync

import "sponsor-tracker/internal/csvfetch"

// GovUKFetcher implements CSVFetcher by downloading from gov.uk.
type GovUKFetcher struct {
	url string
}

func NewGovUKFetcher(url string) *GovUKFetcher {
	return &GovUKFetcher{url: url}
}

func (f *GovUKFetcher) FetchRecords() ([]csvfetch.Record, error) {
	return csvfetch.FetchAndParse(f.url)
}
