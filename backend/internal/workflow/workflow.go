package workflow

import (
	"context"
	"fmt"
)

// ParamType indicates the data type of a workflow parameter.
type ParamType int

const (
	ParamString ParamType = iota
	ParamInt
	ParamFloat
	ParamBool
	ParamJSON
)

// Record is a single row of workflow output.
type Record map[string]any

// Parameter describes a single input to a workflow.
type Parameter struct {
	Name        string
	Type        ParamType
	Required    bool
	Default     any
	Description string
}

// Workflow is a named, self-describing unit of work that can be invoked
// from any entry point (CLI, HTTP, scheduled job).
type Workflow interface {
	Name() string
	Description() string
	Parameters() []Parameter
	Run(ctx context.Context, args map[string]any) ([]Record, error)
}

// Registry holds all registered workflows for discovery and invocation.
type Registry struct {
	workflows map[string]Workflow
}

// NewRegistry creates an empty workflow registry.
func NewRegistry() *Registry {
	return &Registry{workflows: make(map[string]Workflow)}
}

// Register adds a workflow to the registry. It panics if a workflow
// with the same name is already registered, since duplicate names
// indicate a wiring bug that must be caught at startup.
func (r *Registry) Register(w Workflow) {
	name := w.Name()
	if _, exists := r.workflows[name]; exists {
		panic(fmt.Sprintf("workflow %q already registered", name))
	}
	r.workflows[name] = w
}

// Get returns the workflow with the given name, or nil if not found.
func (r *Registry) Get(name string) Workflow {
	return r.workflows[name]
}

// All returns every registered workflow in no guaranteed order.
func (r *Registry) All() []Workflow {
	out := make([]Workflow, 0, len(r.workflows))
	for _, w := range r.workflows {
		out = append(out, w)
	}
	return out
}
