package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"sponsor-tracker/internal/database"
	"sponsor-tracker/internal/sync"
)

// DataReader provides read-only access to the current application state.
type DataReader interface {
	GetAll(ctx context.Context, from, to int, search string) (*database.DataResponse, error)
}

// Server is the HTTP server handling API requests.
type Server struct {
	syncer *sync.Syncer
	data   DataReader
}

// NewServer creates a Server with the given dependencies.
func NewServer(syncer *sync.Syncer, data DataReader) *Server {
	return &Server{syncer: syncer, data: data}
}

// Routes registers all HTTP handlers and returns the root handler.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/sync", s.handleSync)
	mux.HandleFunc("GET /api/data", s.handleGetData)
	return mux
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	result, err := s.syncer.Run(r.Context())
	writeJSON(w, result, err)
}

func (s *Server) handleGetData(w http.ResponseWriter, r *http.Request) {
	from, to, search, err := parseGetDataInput(r)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	data, dataErr := s.data.GetAll(r.Context(), from, to, search)
	writeJSON(w, data, dataErr)
}

// parseGetDataInput extracts and validates the from/to/search query parameters.
// from and to must be positive integers less than 1 billion, with to >= from.
// search is optional (empty string if absent).
func parseGetDataInput(r *http.Request) (int, int, string, error) {
	const maxVal = 1_000_000_000
	from, err := extractInt(r, "from")
	if err != nil { return 0, 0, "", err }
	to, err := extractInt(r, "to")
	if err != nil { return 0, 0, "", err }
	if from < 1 || from > maxVal {
		return 0, 0, "", fmt.Errorf("from must be between 1 and %d", maxVal)
	}
	if to < 1 || to > maxVal {
		return 0, 0, "", fmt.Errorf("to must be between 1 and %d", maxVal)
	}
	if to < from {
		return 0, 0, "", fmt.Errorf("to (%d) must be >= from (%d)", to, from)
	}
	if to-from+1 > 100 {
		return 0, 0, "", fmt.Errorf("page size must not exceed 100")
	}
	search := r.URL.Query().Get("search")
	if len(search) > 200 {
		return 0, 0, "", fmt.Errorf("search must not exceed 200 characters")
	}
	return from, to, search, nil
}

// extractInt parses a required integer query parameter.
// Returns an error if the parameter is missing or not a valid integer.
func extractInt(r *http.Request, name string) (int, error) {
	s := r.URL.Query().Get(name)
	if s == "" {
		return 0, fmt.Errorf("missing required parameter: %s", name)
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: must be an integer", name)
	}
	return v, nil
}

func writeJSON(w http.ResponseWriter, data any, err error) {
	if err != nil {
		slog.Error("request failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
