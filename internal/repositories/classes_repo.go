package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
)

type ClassRepository interface {
	GetAll(ctx context.Context) ([]models.Class, error)
	Create(ctx context.Context, name string) (*models.Class, error)
	Delete(ctx context.Context, id uuid.UUID) error
	BulkUpdate(ctx context.Context, items []models.Class) (int, error)
}

type classRepository struct {
	db *sql.DB
}

func NewClassRepository(db *sql.DB) ClassRepository {
	return &classRepository{db: db}
}

func (r *classRepository) GetAll(ctx context.Context) ([]models.Class, error) {
	// Base classes
	const q = `
		SELECT c.id, c.name, t.id, t.first_name, t.last_name, t.patronymic
		FROM classes c
		LEFT JOIN teachers t ON t.id = c.homeroom_teacher_id
		ORDER BY c.name
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.Class

	for rows.Next() {
		var c models.Class
		var teacherID sql.NullString
		var firstName, lastName sql.NullString
		var patronymic sql.NullString

		if err := rows.Scan(
			&c.ID, &c.Name,
			&teacherID, &firstName, &lastName, &patronymic,
		); err != nil {
			return nil, err
		}

		// Attach homeroom teacher
		if teacherID.Valid {
			tid, err := uuid.Parse(teacherID.String)
			if err == nil {
				c.HomeroomTeacher = &models.Teacher{
					ID:        tid,
					FirstName: firstName.String,
					LastName:  lastName.String,
				}
				if patronymic.Valid {
					c.HomeroomTeacher.Patronymic = &patronymic.String
				}
			}
		}

		// Load subjects
		subjs, err := r.loadClassSubjects(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		c.Subjects = subjs

		// Load groups
		gr, err := r.loadClassGroups(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		c.Groups = gr

		res = append(res, c)
	}

	return res, nil
}

func (r *classRepository) loadClassSubjects(ctx context.Context, classID uuid.UUID) ([]models.ClassSubjectAssignment, error) {
	const q = `
		SELECT cs.subject_id, s.name, cs.hours_per_week,
		       cs.split_groups_count, cs.cross_class_allowed
		FROM class_subjects cs
		JOIN subjects s ON s.id = cs.subject_id
		WHERE cs.class_id = $1
		ORDER BY s.name
	`

	rows, err := r.db.QueryContext(ctx, q, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.ClassSubjectAssignment
	for rows.Next() {
		var subjID uuid.UUID
		var subjName string
		var hours int
		var groupsCount sql.NullInt32
		var crossClass sql.NullBool

		if err := rows.Scan(&subjID, &subjName, &hours, &groupsCount, &crossClass); err != nil {
			return nil, err
		}

		item := models.ClassSubjectAssignment{
			Subject: models.Subject{
				ID:   subjID,
				Name: subjName,
			},
			HoursPerWeek: hours,
		}

		if groupsCount.Valid {
			split := models.ClassSubjectSplit{
				GroupsCount: int(groupsCount.Int32),
			}
			if crossClass.Valid {
				split.CrossClassAllowed = &crossClass.Bool
			}
			item.Split = &split
		}

		out = append(out, item)
	}

	return out, nil
}

func (r *classRepository) loadClassGroups(ctx context.Context, classID uuid.UUID) ([]models.ClassGroup, error) {
	const q = `
		SELECT id, name, size
		FROM class_groups
		WHERE class_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, q, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.ClassGroup
	for rows.Next() {
		var g models.ClassGroup
		var size sql.NullInt32

		if err := rows.Scan(&g.ID, &g.Name, &size); err != nil {
			return nil, err
		}
		if size.Valid {
			n := int(size.Int32)
			g.Size = &n
		}
		out = append(out, g)
	}

	return out, nil
}

func (r *classRepository) Create(ctx context.Context, name string) (*models.Class, error) {
	const q = `
		INSERT INTO classes (name)
		VALUES ($1)
		RETURNING id
	`

	var id uuid.UUID
	if err := r.db.QueryRowContext(ctx, q, name).Scan(&id); err != nil {
		return nil, err
	}

	return &models.Class{
		ID:              id,
		Name:            name,
		Subjects:        []models.ClassSubjectAssignment{},
		Groups:          []models.ClassGroup{},
		HomeroomTeacher: nil,
	}, nil
}

func (r *classRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM classes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("class not found")
	}
	return nil
}

func (r *classRepository) BulkUpdate(ctx context.Context, items []models.Class) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	updated := 0

	for _, c := range items {
		// Update base class
		var teacherID interface{}
		if c.HomeroomTeacher != nil {
			teacherID = c.HomeroomTeacher.ID
		} else {
			teacherID = nil
		}

		res, err := tx.ExecContext(ctx, `
			UPDATE classes SET name = $1, homeroom_teacher_id = $2
			WHERE id = $3
		`, c.Name, teacherID, c.ID)
		if err != nil {
			tx.Rollback()
			return updated, err
		}

		n, _ := res.RowsAffected()
		if n > 0 {
			updated++
		}

		// Replace subjects
		if _, err := tx.ExecContext(ctx, `DELETE FROM class_subjects WHERE class_id = $1`, c.ID); err != nil {
			tx.Rollback()
			return updated, err
		}

		for _, subj := range c.Subjects {
			var groupsCount interface{}
			var crossClass interface{}

			if subj.Split != nil {
				groupsCount = subj.Split.GroupsCount
				crossClass = subj.Split.CrossClassAllowed
			}

			_, err := tx.ExecContext(ctx, `
				INSERT INTO class_subjects (class_id, subject_id, hours_per_week,
				                            split_groups_count, cross_class_allowed)
				VALUES ($1, $2, $3, $4, $5)
			`, c.ID, subj.Subject.ID, subj.HoursPerWeek, groupsCount, crossClass)
			if err != nil {
				tx.Rollback()
				return updated, err
			}
		}

		// Replace groups
		if _, err := tx.ExecContext(ctx, `DELETE FROM class_groups WHERE class_id = $1`, c.ID); err != nil {
			tx.Rollback()
			return updated, err
		}

		for _, gr := range c.Groups {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO class_groups (id, class_id, name, size)
				VALUES ($1, $2, $3, $4)
			`, gr.ID, c.ID, gr.Name, gr.Size)
			if err != nil {
				tx.Rollback()
				return updated, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return updated, err
	}

	return updated, nil
}
