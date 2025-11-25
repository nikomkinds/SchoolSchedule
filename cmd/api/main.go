package main

import (
	"github.com/gin-gonic/gin"
	"github.com/nikomkinds/SchoolSchedule/internal/config"
	"github.com/nikomkinds/SchoolSchedule/internal/handlers"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories/postgres"
	"github.com/nikomkinds/SchoolSchedule/internal/services"
	"github.com/nikomkinds/SchoolSchedule/internal/utils"
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

	// === Repositories ===
	authRepo := repositories.NewAuthRepository(db)
	classroomRepo := repositories.NewClassroomRepository(db)
	subjectRepo := repositories.NewSubjectRepository(db)
	teacherRepo := repositories.NewTeacherRepository(db)

	// === Services ===
	authService := services.NewAuthService(authRepo, db, cfg.JWTSecret)
	classroomService := services.NewClassroomService(classroomRepo)
	subjectService := services.NewSubjectService(subjectRepo)
	teacherService := services.NewTeacherService(teacherRepo)

	// === Handlers ===
	authHandler := handlers.NewAuthHandler(authService)
	classroomHandler := handlers.NewClassroomHandler(classroomService)
	subjectHandler := handlers.NewSubjectHandler(subjectService)
	teacherHandler := handlers.NewTeacherHandler(teacherService)

	// ========== Gin router ==========
	router := gin.Default()

	api := router.Group("/api")

	// ----- AUTH -----
	auth := api.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	// ----- PROTECTED GROUPS -----
	protected := api.Group("/")
	protected.Use(utils.AuthMiddleware(cfg.JWTSecret))

	// ----- CLASSROOMS -----
	classrooms := protected.Group("/classrooms")
	classrooms.GET("", classroomHandler.GetAll)
	classrooms.POST("", classroomHandler.Create)
	classrooms.DELETE("/:id", classroomHandler.Delete)

	// ----- SUBJECTS -----
	subjects := protected.Group("/subjects")
	subjects.GET("", subjectHandler.GetAll)
	subjects.POST("", subjectHandler.Create)
	subjects.DELETE("/:id", subjectHandler.Delete)

	// ----- TEACHERS -----
	users := protected.Group("/users")
	users.GET("/Teachers", teacherHandler.GetAllFull)
	users.GET("/LightTeachers", teacherHandler.GetAllLight)
	users.POST("/Teachers", teacherHandler.Create)
	users.DELETE("/Teachers/:id", teacherHandler.Delete)
	users.PATCH("/Teachers/bulk", teacherHandler.BulkUpdate)

	router.Run(":8080")
	slog.Info("Server started on port 8080")
}
