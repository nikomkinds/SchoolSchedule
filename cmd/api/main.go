package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikomkinds/SchoolSchedule/internal/config"
	"github.com/nikomkinds/SchoolSchedule/internal/handlers"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories/postgres"
	"github.com/nikomkinds/SchoolSchedule/internal/services"
	"github.com/nikomkinds/SchoolSchedule/internal/utils"
)

func main() {

	// ================= LOAD CONFIG =================
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}
	slog.Info("Config loaded")

	// ================= DATABASE ====================
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Database connected")

	// ================= REPOSITORIES ================
	authRepo := repositories.NewAuthRepository(db)
	classroomRepo := repositories.NewClassroomRepository(db)
	subjectRepo := repositories.NewSubjectRepository(db)
	teacherRepo := repositories.NewTeacherRepository(db)
	classRepo := repositories.NewClassRepository(db)
	scheduleRepo := repositories.NewScheduleRepository(db)

	// ================= SERVICES =====================
	authService := services.NewAuthService(authRepo, db, cfg.JWTSecret)
	classroomService := services.NewClassroomService(classroomRepo)
	subjectService := services.NewSubjectService(subjectRepo)
	teacherService := services.NewTeacherService(teacherRepo)
	classService := services.NewClassService(classRepo)
	scheduleService := services.NewScheduleService(scheduleRepo)

	// ================= HANDLERS =====================
	authHandler := handlers.NewAuthHandler(authService)
	classroomHandler := handlers.NewClassroomHandler(classroomService)
	subjectHandler := handlers.NewSubjectHandler(subjectService)
	teacherHandler := handlers.NewTeacherHandler(teacherService)
	classHandler := handlers.NewClassHandler(classService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService)

	// ================= ROUTER (GIN) ================
	router := gin.Default()
	api := router.Group("/api")

	// ---------- AUTH ----------
	auth := api.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	// ---------- PROTECTED ----------
	protected := api.Group("/")
	protected.Use(utils.AuthMiddleware(cfg.JWTSecret))

	// ---------- CLASSROOMS ----------
	classrooms := protected.Group("/classrooms")
	classrooms.GET("", classroomHandler.GetAll)
	classrooms.POST("", classroomHandler.Create)
	classrooms.DELETE("/:id", classroomHandler.Delete)

	// ---------- SUBJECTS ----------
	subjects := protected.Group("/subjects")
	subjects.GET("", subjectHandler.GetAll)
	subjects.POST("", subjectHandler.Create)
	subjects.DELETE("/:id", subjectHandler.Delete)

	// ---------- TEACHERS ----------
	users := protected.Group("/users")
	users.GET("/Teachers", teacherHandler.GetAllFull)
	users.GET("/LightTeachers", teacherHandler.GetAllLight)
	users.POST("/Teachers", teacherHandler.Create)
	users.DELETE("/Teachers/:id", teacherHandler.Delete)
	users.PATCH("/Teachers/bulk", teacherHandler.BulkUpdate)

	// ---------- CLASSES ----------
	classes := protected.Group("/classes")
	classes.GET("", classHandler.GetAll)
	classes.POST("", classHandler.Create)
	classes.DELETE("/:id", classHandler.Delete)
	classes.PUT("/bulk", classHandler.BulkUpdate)

	// ---------- SCHEDULE ----------
	schedule := protected.Group("/schedule")
	schedule.GET("", scheduleHandler.GetScheduleForTeacher)
	schedule.PUT("", scheduleHandler.UpdateScheduleForTeacher)
	schedule.POST("/generate", scheduleHandler.GenerateSchedule)
	schedule.GET("/:id", scheduleHandler.GetScheduleByID)
	schedule.POST("", scheduleHandler.CreateSchedule)
	schedule.DELETE("/:id", scheduleHandler.DeleteSchedule)

	// ================= SERVER ======================
	addr := cfg.ServHost + ":" + cfg.ServPort

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		slog.Info("Server starting...", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
		}
	}()

	// GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	slog.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Forced shutdown", "error", err)
	}

	slog.Info("Server exited cleanly")
}
