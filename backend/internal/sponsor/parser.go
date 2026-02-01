package sponsor

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenStream opens an HTTP connection and returns a stream to read from.
// The caller is responsible for closing the stream.
func OpenStream(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// FetchAndParse downloads and parses the CSV in one call.
// The connection is closed before returning.
func FetchAndParse(url string) ([]Record, error) {
	stream, err := OpenStream(url)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	return Parse(stream)
}

// Parse reads the CSV and returns a slice of Records
func Parse(r io.Reader) ([]Record, error) {
	reader := csv.NewReader(r)

	// Skip header row
	_, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var records []Record
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		record, err := parseRow(row)
		if err != nil {
			// Log and skip malformed rows
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// parseRow converts a CSV row into a Record
func parseRow(row []string) (Record, error) {
	if len(row) < 5 {
		return Record{}, fmt.Errorf("row has %d columns, expected 5", len(row))
	}

	licenceType, rating := parseTypeAndRating(row[3])

	return Record{
		OrganisationName: strings.TrimSpace(row[0]),
		TownCity:         strings.TrimSpace(row[1]),
		County:           strings.TrimSpace(row[2]),
		LicenceType:      licenceType,
		Rating:           rating,
		Route:            strings.TrimSpace(row[4]),
	}, nil
}

// parseTypeAndRating splits "Worker (A rating)" into ("Worker", "A rating")
func parseTypeAndRating(s string) (licenceType, rating string) {
	s = strings.TrimSpace(s)

	// Find the opening parenthesis
	parenIndex := strings.Index(s, "(")
	if parenIndex == -1 {
		return s, ""
	}

	licenceType = strings.TrimSpace(s[:parenIndex])

	// Extract rating from between parentheses
	rating = strings.TrimSpace(s[parenIndex+1:])
	rating = strings.TrimSuffix(rating, ")")

	return licenceType, rating
}
