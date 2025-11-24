package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories"
)

type ClassroomService struct {
	repo repositories.ClassroomRepository
}

func NewClassroomService(repo repositories.ClassroomRepository) *ClassroomService {
	return &ClassroomService{repo: repo}
}

func (s *ClassroomService) GetAll(ctx context.Context) ([]*models.Classroom, error) {
	return s.repo.GetAll(ctx)
}

func (s *ClassroomService) Create(ctx context.Context, name string) (*models.CreateClassroomResponse, error) {
	classroom, err := s.repo.Create(ctx, name)
	if err != nil {
		return nil, err
	}

	return &models.CreateClassroomResponse{
		ID:   classroom.ID,
		Name: classroom.Name,
	}, nil
}

func (s *ClassroomService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
