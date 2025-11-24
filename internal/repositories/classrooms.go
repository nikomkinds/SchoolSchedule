package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
)

type ClassroomRepository interface {
	GetAll(ctx context.Context) ([]*models.Classroom, error)
	Create(ctx context.Context, name string) (*models.Classroom, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type classroomRepository struct {
	db *sql.DB
}

func NewClassroomRepository(db *sql.DB) ClassroomRepository {
	return &classroomRepository{db: db}
}

func (r *classroomRepository) GetAll(ctx context.Context) ([]*models.Classroom, error) {
	const query = `
		SELECT id, name
		FROM classrooms
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*models.Classroom, 0)
	for rows.Next() {
		var c models.Classroom
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}

	return list, nil
}

func (r *classroomRepository) Create(ctx context.Context, name string) (*models.Classroom, error) {
	const query = `
		INSERT INTO classrooms (name)
		VALUES ($1)
		RETURNING id, name
	`

	var c models.Classroom
	err := r.db.QueryRowContext(ctx, query, name).Scan(&c.ID, &c.Name)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *classroomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM classrooms
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
