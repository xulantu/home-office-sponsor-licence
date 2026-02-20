package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"sponsor-tracker/internal/api"
	"sponsor-tracker/internal/auth"
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
	runs := sync.NewPostgresSyncRunRepository(pool)
	syncer := sync.NewSyncer(fetcher, orgs, licences, cfgRepo, runs)

	dataReader := database.NewPostgresDataReader(pool)
	userStore := auth.NewPostgresUserStore(pool)
	sessionStore := auth.NewPostgresSessionStore(pool)
	authService := auth.NewService(userStore, sessionStore)
	server := api.NewServer(syncer, dataReader, authService)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	slog.Info("starting server", "address", addr)
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
