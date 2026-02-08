package csvfetch

import "testing"

func TestExtractCSVURL(t *testing.T) {
	// Simulated HTML containing a CSV link
	html := `
		<html>
		<body>
			<a href="https://assets.publishing.service.gov.uk/media/abc123/2026-01-29_-_Worker_and_Temporary_Worker.csv">Download CSV</a>
		</body>
		</html>
	`

	url, err := extractCSVURL(html)
	if err != nil {
		t.Fatalf("extractCSVURL failed: %v", err)
	}

	expected := "https://assets.publishing.service.gov.uk/media/abc123/2026-01-29_-_Worker_and_Temporary_Worker.csv"
	if url != expected {
		t.Errorf("expected %q, got %q", expected, url)
	}
}

func TestExtractCSVURL_NotFound(t *testing.T) {
	html := `<html><body>No CSV here</body></html>`

	_, err := extractCSVURL(html)
	if err == nil {
		t.Error("expected error when CSV URL not found")
	}
}
