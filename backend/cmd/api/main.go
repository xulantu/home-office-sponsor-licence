package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"sponsor-tracker/internal/api"
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

	server := api.NewServer(syncer)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	slog.Info("starting server", "address", addr)
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
