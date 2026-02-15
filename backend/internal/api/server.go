package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"sponsor-tracker/internal/database"
	"sponsor-tracker/internal/sync"
)

// DataReader provides read-only access to the current application state.
type DataReader interface {
	GetAll(ctx context.Context) (*database.DataResponse, error)
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
	data, err := s.data.GetAll(r.Context())
	writeJSON(w, data, err)
}

func writeJSON(w http.ResponseWriter, data any, err error) {
	if err != nil {
		slog.Error("request failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
