package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
)

type TeacherRepository interface {
	GetAllFull(ctx context.Context) ([]models.Teacher, error)
	GetAllLight(ctx context.Context) ([]models.LightTeacher, error)
	Create(ctx context.Context, firstName, lastName string, patronymic *string) (*models.Teacher, error)
	Delete(ctx context.Context, id uuid.UUID) error
	BulkUpdate(ctx context.Context, items []models.Teacher) (int, error)
}

type teacherRepository struct {
	db *sql.DB
}

func NewTeacherRepository(db *sql.DB) TeacherRepository {
	return &teacherRepository{db: db}
}

// GetAllFull loads teachers with expanded fields (classRoom, class, subjects, classHours)
func (r *teacherRepository) GetAllFull(ctx context.Context) ([]models.Teacher, error) {
	const q = `
		SELECT id, first_name, last_name, patronymic,
		       workload_hours_per_week,
		       classroom_id, classroom_name,
		       homeroom_class_id, homeroom_class_name
		FROM v_teachers_full
		ORDER BY last_name, first_name
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.Teacher
	for rows.Next() {
		var t models.Teacher
		var classroomID sql.NullString
		var classroomName sql.NullString
		var homeroomClassID sql.NullString
		var homeroomClassName sql.NullString

		if err := rows.Scan(
			&t.ID,
			&t.FirstName,
			&t.LastName,
			&t.Patronymic,
			&t.WorkloadHoursPerWeek,
			&classroomID,
			&classroomName,
			&homeroomClassID,
			&homeroomClassName,
		); err != nil {
			return nil, err
		}

		// Classroom
		if classroomID.Valid {
			cid, err := uuid.Parse(classroomID.String)
			if err == nil {
				t.Classroom = &models.Classroom{
					ID:   cid,
					Name: classroomName.String,
				}
			}
		}

		// Homeroom class
		if homeroomClassID.Valid {
			clid, err := uuid.Parse(homeroomClassID.String)
			if err == nil {
				t.HomeroomClass = &models.Class{
					ID:   clid,
					Name: homeroomClassName.String,
				}
			}
		}

		// Load subjects for this teacher
		subjs, err := r.loadTeacherSubjects(ctx, t.ID)
		if err != nil {
			return nil, fmt.Errorf("load subjects for teacher %s: %w", t.ID, err)
		}
		t.Subjects = subjs

		// Load classHours (teacher_workload)
		ch, err := r.loadTeacherClassHours(ctx, t.ID)
		if err != nil {
			return nil, fmt.Errorf("load classHours for teacher %s: %w", t.ID, err)
		}
		t.ClassHours = ch

		res = append(res, t)
	}

	return res, nil
}

func (r *teacherRepository) loadTeacherSubjects(ctx context.Context, teacherID uuid.UUID) ([]models.TeacherSubjectAssignment, error) {
	const q = `
		SELECT subject_id, subject_name, preferred_hours_per_week
		FROM v_teacher_subjects_detailed
		WHERE teacher_id = $1
		ORDER BY subject_name
	`

	rows, err := r.db.QueryContext(ctx, q, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.TeacherSubjectAssignment
	for rows.Next() {
		var subjectID uuid.UUID
		var subjectName string
		var preferred sql.NullInt64

		if err := rows.Scan(&subjectID, &subjectName, &preferred); err != nil {
			return nil, err
		}

		var hoursPtr *int
		if preferred.Valid {
			v := int(preferred.Int64)
			hoursPtr = &v
		}

		out = append(out, models.TeacherSubjectAssignment{
			Subject: models.Subject{
				ID:   subjectID,
				Name: subjectName,
			},
			HoursPerWeek: hoursPtr,
		})
	}
	return out, nil
}

func (r *teacherRepository) loadTeacherClassHours(ctx context.Context, teacherID uuid.UUID) ([]models.TeacherClassHour, error) {
	const q = `
		SELECT class_id, class_name, subject_id, subject_name, group_id, hours_per_week
		FROM v_teacher_workload_detailed
		WHERE teacher_id = $1
		ORDER BY class_name, subject_name
	`

	rows, err := r.db.QueryContext(ctx, q, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.TeacherClassHour
	for rows.Next() {
		var classID uuid.UUID
		var className string
		var subjectID uuid.UUID
		var subjectName string
		var groupID sql.NullString
		var hours int

		if err := rows.Scan(&classID, &className, &subjectID, &subjectName, &groupID, &hours); err != nil {
			return nil, err
		}

		var groupIDPtr *string
		if groupID.Valid {
			tmp := groupID.String
			groupIDPtr = &tmp
		}

		out = append(out, models.TeacherClassHour{
			Class: models.Class{
				ID:   classID,
				Name: className,
			},
			Subject: models.Subject{
				ID:   subjectID,
				Name: subjectName,
			},
			GroupID: groupIDPtr,
			Hours:   hours,
		})
	}

	return out, nil
}

// GetAllLight returns lightweight teacher list (id + names)
func (r *teacherRepository) GetAllLight(ctx context.Context) ([]models.LightTeacher, error) {
	const q = `
		SELECT id, first_name, last_name, patronymic
		FROM teachers
		ORDER BY last_name, first_name
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.LightTeacher
	for rows.Next() {
		var t models.LightTeacher
		var patronymic sql.NullString
		if err := rows.Scan(&t.ID, &t.FirstName, &t.LastName, &patronymic); err != nil {
			return nil, err
		}
		if patronymic.Valid {
			t.Patronymic = &patronymic.String
		}
		out = append(out, t)
	}
	return out, nil
}

// Create inserts a new teacher (minimal fields) and returns created teacher
func (r *teacherRepository) Create(ctx context.Context, firstName, lastName string, patronymic *string) (*models.Teacher, error) {
	const q = `
		INSERT INTO teachers (first_name, last_name, patronymic)
		VALUES ($1, $2, $3)
		RETURNING id, first_name, last_name, patronymic, workload_hours_per_week, classroom_id, homeroom_class_id
	`

	var t models.Teacher
	var classroomID sql.NullString
	var homeroomClassID sql.NullString
	var patron sql.NullString

	if patronymic != nil {
		patron.String = *patronymic
		patron.Valid = true
	}

	err := r.db.QueryRowContext(ctx, q, firstName, lastName, patron).Scan(
		&t.ID, &t.FirstName, &t.LastName, &patron, &t.WorkloadHoursPerWeek, &classroomID, &homeroomClassID,
	)
	if err != nil {
		return nil, err
	}
	if patron.Valid {
		t.Patronymic = &patron.String
	}

	// Classroom and class are nil by default; subjects and classHours empty slices
	t.Subjects = []models.TeacherSubjectAssignment{}
	t.ClassHours = []models.TeacherClassHour{}

	// if classroom/homeroom present - try to parse and set (unlikely just created)
	if classroomID.Valid {
		if cid, err := uuid.Parse(classroomID.String); err == nil {
			t.Classroom = &models.Classroom{ID: cid, Name: ""} // name not returned
		}
	}
	if homeroomClassID.Valid {
		if hid, err := uuid.Parse(homeroomClassID.String); err == nil {
			t.HomeroomClass = &models.Class{ID: hid, Name: ""} // name not returned
		}
	}

	return &t, nil
}

// Delete deletes teacher by id
func (r *teacherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM teachers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// BulkUpdate updates many teachers; performs updates to teachers, teacher_subjects and teacher_workload
// Returns number of teachers updated.
func (r *teacherRepository) BulkUpdate(ctx context.Context, items []models.Teacher) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	updated := 0

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, t := range items {
		// update base teacher info
		var patron sql.NullString
		if t.Patronymic != nil {
			patron.String = *t.Patronymic
			patron.Valid = true
		}
		// classroom_id and homeroom_class_id may be nil (we accept nil)
		var classroomID interface{}
		var homeroomID interface{}
		if t.Classroom != nil {
			classroomID = t.Classroom.ID
		} else {
			classroomID = nil
		}
		if t.HomeroomClass != nil {
			homeroomID = t.HomeroomClass.ID
		} else {
			homeroomID = nil
		}

		res, err := tx.ExecContext(ctx, `
			UPDATE teachers SET first_name = $1, last_name = $2, patronymic = $3,
				classroom_id = $4, homeroom_class_id = $5
			WHERE id = $6
		`, t.FirstName, t.LastName, patron, classroomID, homeroomID, t.ID)
		if err != nil {
			tx.Rollback()
			return updated, err
		}
		n, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return updated, err
		}

		if n > 0 {
			updated++
		}

		// Replace teacher_subjects: delete & insert
		if _, err := tx.ExecContext(ctx, `DELETE FROM teacher_subjects WHERE teacher_id = $1`, t.ID); err != nil {
			tx.Rollback()
			return updated, err
		}
		for _, tsa := range t.Subjects {
			// tsa.Subject.ID must be valid
			var pref interface{}
			if tsa.HoursPerWeek != nil {
				pref = *tsa.HoursPerWeek
			} else {
				pref = nil
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO teacher_subjects (teacher_id, subject_id, preferred_hours_per_week)
				VALUES ($1, $2, $3)
				ON CONFLICT (teacher_id, subject_id) DO UPDATE SET preferred_hours_per_week = EXCLUDED.preferred_hours_per_week
			`, t.ID, tsa.Subject.ID, pref); err != nil {
				tx.Rollback()
				return updated, err
			}
		}

		// Replace teacher_workload (classHours): delete & insert
		if _, err := tx.ExecContext(ctx, `DELETE FROM teacher_workload WHERE teacher_id = $1`, t.ID); err != nil {
			tx.Rollback()
			return updated, err
		}
		for _, ch := range t.ClassHours {
			var groupID interface{}
			if ch.GroupID != nil {
				// stored as uuid in DB: but frontend groupId may be string - here it is *string containing UUID
				if gid, err := uuid.Parse(*ch.GroupID); err == nil {
					groupID = gid
				} else {
					groupID = nil
				}
			} else {
				groupID = nil
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO teacher_workload (teacher_id, class_id, subject_id, group_id, hours_per_week)
				VALUES ($1, $2, $3, $4, $5)
			`, t.ID, ch.Class.ID, ch.Subject.ID, groupID, ch.Hours); err != nil {
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
