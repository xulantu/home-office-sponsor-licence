package csvfetch

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	// Fake CSV data (no network needed!)
	csv := `"Organisation Name","Town/City","County","Type & Rating","Route"
"Google UK","London","","Worker (A rating)","Skilled Worker"
"Microsoft Ltd","Reading","Berkshire","Worker (B rating)","Skilled Worker"
"Acme Corp","Manchester","","Temporary Worker (A rating)","Creative Worker"`

	// strings.NewReader creates an io.Reader from a string
	reader := strings.NewReader(csv)

	records, err := Parse(reader)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check we got 3 records
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}

	// Check first record
	first := records[0]
	if first.OrganisationName != "Google UK" {
		t.Errorf("expected 'Google UK', got '%s'", first.OrganisationName)
	}
	if first.LicenceType != "Worker" {
		t.Errorf("expected 'Worker', got '%s'", first.LicenceType)
	}
	if first.Rating != "A rating" {
		t.Errorf("expected 'A rating', got '%s'", first.Rating)
	}

	// Check second record has B rating
	second := records[1]
	if second.Rating != "B rating" {
		t.Errorf("expected 'B rating', got '%s'", second.Rating)
	}

	// Check third record is Temporary Worker
	third := records[2]
	if third.LicenceType != "Temporary Worker" {
		t.Errorf("expected 'Temporary Worker', got '%s'", third.LicenceType)
	}
}

func TestParseTypeAndRating(t *testing.T) {
	tests := []struct {
		input       string
		wantType    string
		wantRating  string
	}{
		{"Worker (A rating)", "Worker", "A rating"},
		{"Worker (B rating)", "Worker", "B rating"},
		{"Temporary Worker (A rating)", "Temporary Worker", "A rating"},
		{"Unknown Format", "Unknown Format", ""},
	}

	for _, tt := range tests {
		gotType, gotRating := parseTypeAndRating(tt.input)
		if gotType != tt.wantType {
			t.Errorf("parseTypeAndRating(%q): type = %q, want %q", tt.input, gotType, tt.wantType)
		}
		if gotRating != tt.wantRating {
			t.Errorf("parseTypeAndRating(%q): rating = %q, want %q", tt.input, gotRating, tt.wantRating)
		}
	}
}
