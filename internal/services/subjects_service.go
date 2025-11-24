package services

import (
	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/repositories"
)

type SubjectService interface {
	GetAll() ([]models.Subject, error)
	Create(name string) (models.Subject, error)
	Delete(id uuid.UUID) error
}

type subjectService struct {
	repo repositories.SubjectRepository
}

func NewSubjectService(repo repositories.SubjectRepository) SubjectService {
	return &subjectService{repo: repo}
}

func (s *subjectService) GetAll() ([]models.Subject, error) {
	return s.repo.GetAll()
}

func (s *subjectService) Create(name string) (models.Subject, error) {
	return s.repo.Create(name)
}

func (s *subjectService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}
