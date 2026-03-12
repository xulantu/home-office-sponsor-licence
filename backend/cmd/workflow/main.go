package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"sponsor-tracker/internal/config"
	"sponsor-tracker/internal/database"
	"sponsor-tracker/internal/sync"
	"sponsor-tracker/internal/workflow"
)

func main() {
	cfg, err := config.Load("config.yaml", ".env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := database.Connect(cfg.Database.ConnectionString())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	fetcher := sync.NewGovUKFetcher()
	db := sync.NewPostgresDB(pool)
	syncer := sync.NewSyncer(fetcher, db)

	registry := workflow.NewRegistry()
	registry.Register(workflow.NewSyncWorkflow(syncer))
	registry.Register(workflow.NewShowRunsWorkflow(syncer))
	registry.Register(workflow.NewRollbackWorkflow(syncer))

	if len(os.Args) < 2 {
		printUsage(registry)
		os.Exit(0)
	}

	name := os.Args[1]
	w := registry.Get(name)
	if w == nil {
		fmt.Fprintf(os.Stderr, "unknown workflow: %s\n\n", name)
		printUsage(registry)
		os.Exit(1)
	}

	args, err := parseArgs(os.Args[2:], w.Parameters())
	if err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}

	records, err := w.Run(context.Background(), args)
	if err != nil {
		log.Fatalf("workflow failed: %v", err)
	}

	printRecords(records)
}

func printUsage(registry *workflow.Registry) {
	fmt.Println("Usage: workflow <name> [--key=value ...]")
	fmt.Println()
	fmt.Println("Available workflows:")
	for _, w := range registry.All() {
		fmt.Printf("  %-15s %s\n", w.Name(), w.Description())
		for _, p := range w.Parameters() {
			req := "optional"
			if p.Required {
				req = "required"
			}
			fmt.Printf("    --%-12s %s (%s, default: %v)\n", p.Name, p.Description, req, p.Default)
		}
	}
}

func parseArgs(raw []string, params []workflow.Parameter) (map[string]any, error) {
	args := make(map[string]any)

	// Apply defaults first.
	for _, p := range params {
		if p.Default != nil {
			args[p.Name] = p.Default
		}
	}

	// Build lookup for declared parameters.
	paramByName := make(map[string]workflow.Parameter, len(params))
	for _, p := range params {
		paramByName[p.Name] = p
	}

	// Parse --key=value pairs.
	for _, s := range raw {
		if !strings.HasPrefix(s, "--") {
			return nil, fmt.Errorf("unexpected argument: %s (expected --key=value)", s)
		}
		kv := strings.TrimPrefix(s, "--")
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("argument %s must use --key=value format", s)
		}
		key, val := parts[0], parts[1]

		p, ok := paramByName[key]
		if !ok {
			return nil, fmt.Errorf("unknown parameter: %s", key)
		}

		converted, err := convertValue(val, p.Type)
		if err != nil {
			return nil, fmt.Errorf("parameter %s: %w", key, err)
		}
		args[key] = converted
	}

	// Check required parameters.
	for _, p := range params {
		if p.Required {
			if _, ok := args[p.Name]; !ok {
				return nil, fmt.Errorf("missing required parameter: %s", p.Name)
			}
		}
	}

	return args, nil
}

func convertValue(val string, t workflow.ParamType) (any, error) {
	switch t {
	case workflow.ParamString:
		return val, nil
	case workflow.ParamInt:
		return strconv.Atoi(val)
	case workflow.ParamFloat:
		return strconv.ParseFloat(val, 64)
	case workflow.ParamBool:
		return strconv.ParseBool(val)
	case workflow.ParamJSON:
		var v any
		if err := json.Unmarshal([]byte(val), &v); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported parameter type")
	}
}

func printRecords(records []workflow.Record) {
	if len(records) == 0 {
		fmt.Println("(no results)")
		return
	}

	// Collect column names from the first record for consistent ordering.
	var cols []string
	for k := range records[0] {
		cols = append(cols, k)
	}

	// Calculate column widths.
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = len(c)
	}
	rows := make([][]string, len(records))
	for i, rec := range records {
		rows[i] = make([]string, len(cols))
		for j, c := range cols {
			s := fmt.Sprintf("%v", rec[c])
			rows[i][j] = s
			if len(s) > widths[j] {
				widths[j] = len(s)
			}
		}
	}

	// Print header.
	for i, c := range cols {
		fmt.Printf("%-*s", widths[i]+2, c)
	}
	fmt.Println()
	for i := range cols {
		fmt.Print(strings.Repeat("-", widths[i]+2))
	}
	fmt.Println()

	// Print rows.
	for _, row := range rows {
		for i, val := range row {
			fmt.Printf("%-*s", widths[i]+2, val)
		}
		fmt.Println()
	}
}
