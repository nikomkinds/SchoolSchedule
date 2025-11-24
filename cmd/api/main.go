package main

import (
	"github.com/gin-gonic/gin"
	"github.com/nikomkinds/SchoolSchedule/internal/config"
	"github.com/nikomkinds/SchoolSchedule/internal/handlers"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories/postgres"
	"github.com/nikomkinds/SchoolSchedule/internal/services"
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

	// ========== Repository layer ==========
	authRepo := repositories.NewAuthRepository(db)

	// ========== Service layer ==========
	authService := services.NewAuthService(authRepo, db, cfg.JWTSecret)

	// ========== Handler layer ==========
	authHandler := handlers.NewAuthHandler(authService)

	// ========== Gin router ==========
	router := gin.Default()

	api := router.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	router.Run(":8080")
	slog.Info("Server started on port 8080")
}
