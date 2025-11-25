package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories"
)

type TeacherService interface {
	GetAllFull(ctx context.Context) ([]models.Teacher, error)
	GetAllLight(ctx context.Context) ([]models.LightTeacher, error)
	Create(ctx context.Context, firstName, lastName string, patronymic *string) (*models.Teacher, error)
	Delete(ctx context.Context, id uuid.UUID) error
	BulkUpdate(ctx context.Context, items []models.Teacher) (int, error)
}

type teacherService struct {
	repo repositories.TeacherRepository
}

func NewTeacherService(repo repositories.TeacherRepository) TeacherService {
	return &teacherService{repo: repo}
}

func (s *teacherService) GetAllFull(ctx context.Context) ([]models.Teacher, error) {
	return s.repo.GetAllFull(ctx)
}

func (s *teacherService) GetAllLight(ctx context.Context) ([]models.LightTeacher, error) {
	return s.repo.GetAllLight(ctx)
}

func (s *teacherService) Create(ctx context.Context, firstName, lastName string, patronymic *string) (*models.Teacher, error) {
	return s.repo.Create(ctx, firstName, lastName, patronymic)
}

func (s *teacherService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *teacherService) BulkUpdate(ctx context.Context, items []models.Teacher) (int, error) {
	return s.repo.BulkUpdate(ctx, items)
}
