package main

import (
	"context"
	"fmt"
	"log"

	"sponsor-tracker/internal/config"
	"sponsor-tracker/internal/database"
	"sponsor-tracker/internal/sync"
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
	orgs := sync.NewPostgresOrgRepository(pool)
	licences := sync.NewPostgresLicenceRepository(pool)
	cfgRepo := sync.NewPostgresConfigRepository(pool)
	syncer := sync.NewSyncer(fetcher, orgs, licences, cfgRepo)

	result, err := syncer.Run(context.Background())
	if err != nil {
		log.Fatalf("sync failed: %v", err)
	}

	fmt.Printf("Sync complete:\n")
	fmt.Printf("  New organisations:    %d\n", result.NewOrganisations)
	fmt.Printf("  New licences:         %d\n", result.NewLicences)
	fmt.Printf("  Changed licences:     %d\n", result.ChangedLicences)
	fmt.Printf("  Closed organisations: %d\n", result.ClosedOrganisations)
	fmt.Printf("  Closed licences:      %d\n", result.ClosedLicences)
	fmt.Printf("  Errors:               %d\n", len(result.Errors))

	for i, e := range result.Errors {
		fmt.Printf("  error %d: %v\n", i+1, e)
	}
}
