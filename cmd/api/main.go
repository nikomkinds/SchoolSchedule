package main

import (
	"github.com/nikomkinds/SchoolSchedule/internal/config"
	"log/slog"
)

func main() {

	// Loading config (connection params)
	_, err := config.LoadConfig() // cfg, err !!!
	if err != nil {
		slog.Error("Failed to load config:", err)
	}
	slog.Info("Load config")
}
