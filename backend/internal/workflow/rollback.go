package workflow

import (
	"context"

	"sponsor-tracker/internal/sync"
)

// RollbackWorkflow rolls back the most recent sync run.
type RollbackWorkflow struct {
	syncer *sync.Syncer
}

// NewRollbackWorkflow creates a RollbackWorkflow with the given syncer.
func NewRollbackWorkflow(syncer *sync.Syncer) *RollbackWorkflow {
	return &RollbackWorkflow{syncer: syncer}
}

func (w *RollbackWorkflow) Name() string           { return "rollback" }
func (w *RollbackWorkflow) Description() string     { return "Roll back the most recent sync run" }
func (w *RollbackWorkflow) Parameters() []Parameter { return nil }

func (w *RollbackWorkflow) Run(ctx context.Context, _ map[string]any) ([]Record, error) {
	result, err := w.syncer.Rollback(ctx)
	if err != nil {
		return nil, err
	}

	return []Record{{
		"rolled_back_run_id":     result.RolledBackRunID,
		"organisations_deleted":  result.OrganisationsDeleted,
		"organisations_restored": result.OrganisationsRestored,
		"licences_deleted":       result.LicencesDeleted,
		"licences_restored":      result.LicencesRestored,
	}}, nil
}
