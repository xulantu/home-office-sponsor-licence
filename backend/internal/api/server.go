package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"sponsor-tracker/internal/sync"
)

// Server is the HTTP server handling API requests.
type Server struct {
	syncer *sync.Syncer
}

// NewServer creates a Server with the given dependencies.
func NewServer(syncer *sync.Syncer) *Server {
	return &Server{syncer: syncer}
}

// Routes registers all HTTP handlers and returns the root handler.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/sync", s.handleSync)
	return mux
}

// handleSync triggers a full sync with the gov.uk sponsor list.
func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	result, err := s.syncer.Run(r.Context())
	if err != nil {
		slog.Error("sync failed", "error", err)
		http.Error(w, "sync failed", http.StatusInternalServerError)
		return
	}

	slog.Info("sync complete",
		"new_organisations", result.NewOrganisations,
		"new_licences", result.NewLicences,
		"changed_licences", result.ChangedLicences,
		"errors", len(result.Errors),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
