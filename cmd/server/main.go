// Command waqti runs the offline-first mosque signage server:
// the /display kiosk view, the password-protected /admin panel, and the
// REST/SSE API backing both.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/heyubaidullah/waqti/internal/api"
	"github.com/heyubaidullah/waqti/internal/auth"
	"github.com/heyubaidullah/waqti/internal/config"
	"github.com/heyubaidullah/waqti/internal/db"
)

const backupTickInterval = 6 * time.Hour

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config.Load: %v", err)
	}
	log.Printf("%s starting; data dir: %s", config.AppName, cfg.DataDir)

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("db.Open: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("db.Migrate: %v", err)
	}
	if err := api.SeedDefaultSettings(database); err != nil {
		log.Fatalf("api.SeedDefaultSettings: %v", err)
	}

	authManager := auth.NewManager(database)
	if err := authManager.Bootstrap(); err != nil {
		log.Fatalf("auth.Bootstrap: %v", err)
	}

	deps := &api.Deps{
		DB:          database,
		Auth:        authManager,
		Cfg:         cfg,
		Broadcaster: api.NewBroadcaster(),
	}
	mux := api.NewRouter(deps)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stopBackupTicker := startBackupTicker(ctx, database, cfg.BackupsDir)
	defer stopBackupTicker()

	go func() {
		log.Printf("%s listening on %s", config.AppName, cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

// startBackupTicker runs a periodic safety-net backup (independent of the
// per-admin-write backups triggered in internal/api) and returns a function
// that stops it.
func startBackupTicker(ctx context.Context, database *sql.DB, backupsDir string) func() {
	ticker := time.NewTicker(backupTickInterval)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-ticker.C:
				if _, err := db.Backup(database, backupsDir); err != nil {
					log.Printf("periodic backup failed: %v", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	return func() { <-done }
}
