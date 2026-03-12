package workflow

import (
	"context"

	"sponsor-tracker/internal/sync"
)

// SyncWorkflow runs the gov.uk sponsor licence synchronisation.
type SyncWorkflow struct {
	syncer *sync.Syncer
}

// NewSyncWorkflow creates a SyncWorkflow with the given syncer.
func NewSyncWorkflow(syncer *sync.Syncer) *SyncWorkflow {
	return &SyncWorkflow{syncer: syncer}
}

func (w *SyncWorkflow) Name() string        { return "sync" }
func (w *SyncWorkflow) Description() string  { return "Synchronise the database with the current gov.uk sponsor licence CSV" }
func (w *SyncWorkflow) Parameters() []Parameter { return nil }

func (w *SyncWorkflow) Run(ctx context.Context, _ map[string]any) ([]Record, error) {
	result, err := w.syncer.Run(ctx)
	if err != nil {
		return nil, err
	}

	return []Record{{
		"new_organisations":    result.NewOrganisations,
		"new_licences":         result.NewLicences,
		"changed_licences":     result.ChangedLicences,
		"closed_organisations": result.ClosedOrganisations,
		"closed_licences":      result.ClosedLicences,
		"errors":               len(result.Errors),
	}}, nil
}
