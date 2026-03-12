package csvfetch

import (
	"context"
	"strings"
	"testing"
)

func TestParse6Column(t *testing.T) {
	csv := `Organisation Name,Town/City,County,Tier,Rating,Route
Google UK,London,,Worker,A rating,Skilled Worker
Microsoft Ltd,Reading,Berkshire,Worker,B rating,Skilled Worker
Acme Corp,Manchester,,Temporary Worker,A rating,Creative Worker`

	// strings.NewReader creates an io.Reader from a string
	reader := strings.NewReader(csv)

	records, err := Parse(context.Background(), reader)
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

func TestParse5Column(t *testing.T) {
	csv := `"Organisation Name","Town/City","County","Type & Rating","Route"
"Google UK","London",,"Worker (A rating)","Skilled Worker"
"Microsoft Ltd","Reading","Berkshire","Worker (B rating)","Skilled Worker"
"Acme Corp","Manchester",,"Temporary Worker (A rating)","Creative Worker"`

	records, err := Parse(context.Background(), strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
	if records[0].LicenceType != "Worker" {
		t.Errorf("expected 'Worker', got '%s'", records[0].LicenceType)
	}
	if records[0].Rating != "A rating" {
		t.Errorf("expected 'A rating', got '%s'", records[0].Rating)
	}
	if records[1].Rating != "B rating" {
		t.Errorf("expected 'B rating', got '%s'", records[1].Rating)
	}
	if records[2].LicenceType != "Temporary Worker" {
		t.Errorf("expected 'Temporary Worker', got '%s'", records[2].LicenceType)
	}
}

func TestParseTypeAndRating(t *testing.T) {
	tests := []struct {
		input      string
		wantType   string
		wantRating string
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

