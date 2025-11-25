package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories"
)

type ClassService interface {
	GetAll(ctx context.Context) ([]models.Class, error)
	Create(ctx context.Context, name string) (*models.Class, error)
	Delete(ctx context.Context, id uuid.UUID) error
	BulkUpdate(ctx context.Context, items []models.Class) (int, error)
}

type classService struct {
	repo repositories.ClassRepository
}

func NewClassService(repo repositories.ClassRepository) ClassService {
	return &classService{repo: repo}
}

func (s *classService) GetAll(ctx context.Context) ([]models.Class, error) {
	return s.repo.GetAll(ctx)
}

func (s *classService) Create(ctx context.Context, name string) (*models.Class, error) {
	return s.repo.Create(ctx, name)
}

func (s *classService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *classService) BulkUpdate(ctx context.Context, items []models.Class) (int, error) {
	return s.repo.BulkUpdate(ctx, items)
}
