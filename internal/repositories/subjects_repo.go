package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
)

type SubjectRepository interface {
	GetAll(ctx context.Context) ([]*models.Subject, error)
	Create(ctx context.Context, name string) (*models.Subject, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type subjectRepository struct {
	db *sql.DB
}

func NewSubjectRepository(db *sql.DB) SubjectRepository {
	return &subjectRepository{db: db}
}

func (r *subjectRepository) GetAll(ctx context.Context) ([]*models.Subject, error) {
	const query = `
		SELECT id, name
		FROM subjects
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*models.Subject, 0)
	for rows.Next() {
		var s models.Subject
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}

	return list, nil
}

func (r *subjectRepository) Create(ctx context.Context, name string) (*models.Subject, error) {
	const query = `
		INSERT INTO subjects (name)
		VALUES ($1)
		RETURNING id, name
	`

	var s models.Subject
	err := r.db.QueryRowContext(ctx, query, name).Scan(&s.ID, &s.Name)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *subjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM subjects
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
