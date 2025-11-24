package main

import (
	"github.com/nikomkinds/SchoolSchedule/internal/config"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories/postgres"
	"log/slog"
)

func main() {

	// Loading config (connection params)
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config:", err)
	}
	slog.Info("Load config")

	// Connecting to the database
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("Failed to connect to database:", err)
	}
	defer db.Close()
	slog.Info("Connect to database")

}
