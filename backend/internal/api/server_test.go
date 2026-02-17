package api

import (
	"net/http"
	"testing"
)

func TestParseGetDataInput(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantFrom   int
		wantTo     int
		wantSearch string
		wantErr    bool
	}{
		{"valid with search", "from=1&to=20&search=acme", 1, 20, "acme", false},
		{"valid without search", "from=1&to=20", 1, 20, "", false},
		{"page size over 100", "from=1&to=200", 0, 0, "", true},
		{"page size exactly 100", "from=1&to=100", 1, 100, "", false},
		{"missing from", "to=20", 0, 0, "", true},
		{"missing to", "from=1", 0, 0, "", true},
		{"to less than from", "from=10&to=5", 0, 0, "", true},
		{"from is zero", "from=0&to=20", 0, 0, "", true},
		{"non-integer from", "from=abc&to=20", 0, 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/api/data?"+tt.query, nil)
			from, to, search, err := parseGetDataInput(r)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil { return }
			if from != tt.wantFrom { t.Errorf("from = %d, want %d", from, tt.wantFrom) }
			if to != tt.wantTo { t.Errorf("to = %d, want %d", to, tt.wantTo) }
			if search != tt.wantSearch { t.Errorf("search = %q, want %q", search, tt.wantSearch) }
		})
	}
}
