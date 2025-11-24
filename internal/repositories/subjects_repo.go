package repositories

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
)

type SubjectRepository interface {
	GetAll() ([]models.Subject, error)
	Create(name string) (models.Subject, error)
	Delete(id uuid.UUID) error
}

type subjectRepository struct {
	db *sql.DB
}

func NewSubjectRepository(db *sql.DB) SubjectRepository {
	return &subjectRepository{db: db}
}

func (r *subjectRepository) GetAll() ([]models.Subject, error) {
	rows, err := r.db.Query(`SELECT id, name FROM subjects ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []models.Subject

	for rows.Next() {
		var s models.Subject
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		subjects = append(subjects, s)
	}

	return subjects, nil
}

func (r *subjectRepository) Create(name string) (models.Subject, error) {
	var s models.Subject

	err := r.db.QueryRow(
		`INSERT INTO subjects (name) 
         VALUES ($1) 
         RETURNING id, name`,
		name,
	).Scan(&s.ID, &s.Name)

	return s, err
}

func (r *subjectRepository) Delete(id uuid.UUID) error {
	res, err := r.db.Exec(`DELETE FROM subjects WHERE id = $1`, id)
	if err != nil {
		return err
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("subject not found")
	}

	return nil
}
