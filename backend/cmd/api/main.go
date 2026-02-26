package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	db := sync.NewPostgresDB(pool)
	syncer := sync.NewSyncer(fetcher, db)

	dataReader := database.NewPostgresDataReader(pool)
	userStore := auth.NewPostgresUserStore(pool)
	sessionStore := auth.NewPostgresSessionStore(pool)
	invCodeStore := auth.NewPostgresInvitationCodeStore(pool)
	resetStore := auth.NewPostgresPasswordResetStore(pool)
	authService := auth.NewService(userStore, sessionStore, invCodeStore, resetStore)
	server := api.NewServer(syncer, dataReader, authService, cfg.Server.SecureCookies)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	httpServer := &http.Server{Addr: addr, Handler: server.Routes()}

	go func() {
		slog.Info("starting server", "address", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	slog.Info("server stopped")
}
