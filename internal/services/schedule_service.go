package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories"
)

type ScheduleService interface {
	// GetScheduleForTeacher loads the schedule for a specific teacher
	GetScheduleForTeacher(ctx context.Context, teacherID uuid.UUID) ([]models.ScheduleDay, error)
	// GetScheduleByID loads a specific named schedule by its ID
	GetScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.Schedule, error)
	// GetAllSchedules loads all named schedules
	GetAllSchedules(ctx context.Context) ([]models.Schedule, error)
	// CreateSchedule creates a new named schedule
	CreateSchedule(ctx context.Context, schedule models.Schedule, slots []models.ScheduleSlotInput) (*models.Schedule, error)
	// UpdateSchedule updates an existing schedule
	UpdateSchedule(ctx context.Context, scheduleID uuid.UUID, name *string, slots []models.ScheduleSlotInput) error
	// DeleteSchedule deletes a schedule
	DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error
	// GenerateSchedule generates a schedule based on study plans and workload
	GenerateSchedule(ctx context.Context, req models.GenerateScheduleRequest) ([]models.ScheduleDay, error)
}

type scheduleService struct {
	repo repositories.ScheduleRepository
}

func NewScheduleService(repo repositories.ScheduleRepository) ScheduleService {
	return &scheduleService{repo: repo}
}

func (s *scheduleService) GetScheduleForTeacher(ctx context.Context, teacherID uuid.UUID) ([]models.ScheduleDay, error) {
	return s.repo.GetScheduleForTeacher(ctx, teacherID)
}

func (s *scheduleService) GetScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.Schedule, error) {
	return s.repo.GetScheduleByID(ctx, scheduleID)
}

func (s *scheduleService) GetAllSchedules(ctx context.Context) ([]models.Schedule, error) {
	return s.repo.GetAllSchedules(ctx)
}

func (s *scheduleService) CreateSchedule(ctx context.Context, schedule models.Schedule, slots []models.ScheduleSlotInput) (*models.Schedule, error) {
	return s.repo.CreateSchedule(ctx, schedule, slots)
}

func (s *scheduleService) UpdateSchedule(ctx context.Context, scheduleID uuid.UUID, name *string, slots []models.ScheduleSlotInput) error {
	return s.repo.UpdateSchedule(ctx, scheduleID, name, slots)
}

func (s *scheduleService) DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	return s.repo.DeleteSchedule(ctx, scheduleID)
}

func (s *scheduleService) GenerateSchedule(ctx context.Context, req models.GenerateScheduleRequest) ([]models.ScheduleDay, error) {
	return s.repo.GenerateSchedule(ctx, req)
}
