package sponsor

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

const (
	// HomeOfficePageURL is the page listing sponsor licence data
	HomeOfficePageURL = "https://www.gov.uk/government/publications/register-of-licensed-sponsors-workers"
)

// DiscoverCSVURL fetches the Home Office page and extracts the CSV download link
func DiscoverCSVURL() (string, error) {
	resp, err := http.Get(HomeOfficePageURL)
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
	// Pattern matches URLs ending in .csv on the assets domain
	pattern := `https://assets\.publishing\.service\.gov\.uk/[^"]+\.csv`
	re := regexp.MustCompile(pattern)

	match := re.FindString(html)
	if match == "" {
		return "", fmt.Errorf("CSV URL not found in page")
	}

	return match, nil
}
