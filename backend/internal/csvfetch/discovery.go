package csvfetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

const (
	// HomeOfficePageURL is the page listing sponsor licence data
	HomeOfficePageURL = "https://www.gov.uk/government/publications/register-of-licensed-sponsors-workers"
)

// csvURLPattern matches CSV download URLs on the gov.uk assets domain.
// Compiled once at package init for performance.
var csvURLPattern = regexp.MustCompile(`https://assets\.publishing\.service\.gov\.uk/[^"]+\.csv`)

// DiscoverCSVURL fetches the Home Office page and extracts the CSV download link
func DiscoverCSVURL(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, HomeOfficePageURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	resp, err := discoveryClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read page: %w", err)
	}

	return extractCSVURL(string(body))
}

// extractCSVURL finds the CSV link in the HTML page
func extractCSVURL(html string) (string, error) {
	match := csvURLPattern.FindString(html)
	if match == "" {
		return "", fmt.Errorf("CSV URL not found in page")
	}

	return match, nil
}
