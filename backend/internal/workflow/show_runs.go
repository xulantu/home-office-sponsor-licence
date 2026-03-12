package workflow

import (
	"context"
	"fmt"

	"sponsor-tracker/internal/sync"
)

// ShowRunsWorkflow displays recent sync run history.
type ShowRunsWorkflow struct {
	syncer *sync.Syncer
}

// NewShowRunsWorkflow creates a ShowRunsWorkflow with the given syncer.
func NewShowRunsWorkflow(syncer *sync.Syncer) *ShowRunsWorkflow {
	return &ShowRunsWorkflow{syncer: syncer}
}

func (w *ShowRunsWorkflow) Name() string        { return "show-runs" }
func (w *ShowRunsWorkflow) Description() string  { return "Show recent sync run history" }
func (w *ShowRunsWorkflow) Parameters() []Parameter {
	return []Parameter{
		{Name: "count", Type: ParamInt, Required: false, Default: 5, Description: "Number of recent runs to show"},
	}
}

func (w *ShowRunsWorkflow) Run(ctx context.Context, args map[string]any) ([]Record, error) {
	count := 5
	if v, ok := args["count"]; ok {
		n, ok := v.(int)
		if !ok {
			return nil, fmt.Errorf("count must be an integer")
		}
		count = n
	}

	runs, err := w.syncer.ShowRuns(ctx, count)
	if err != nil {
		return nil, err
	}

	records := make([]Record, len(runs))
	for i, r := range runs {
		records[i] = Record{
			"id":                   r.ID,
			"start_time":           r.StartTime,
			"end_time":             r.EndTime,
			"new_organisations":    r.NewOrganisations,
			"new_licences":         r.NewLicences,
			"changed_licences":     r.ChangedLicences,
			"closed_organisations": r.ClosedOrganisations,
			"closed_licences":      r.ClosedLicences,
			"error_count":          r.ErrorCount,
		}
	}
	return records, nil
}
