package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
)

type ScheduleRepository interface {
	// GetScheduleForTeacher loads the schedule for a specific teacher
	GetSchedule(ctx context.Context, teacherID uuid.UUID) ([]models.ScheduleDay, error)
	// GetScheduleByID loads a specific named schedule by its ID
	GetScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.Schedule, error)
	// GetAllSchedules loads all named schedules
	GetAllSchedules(ctx context.Context) ([]models.Schedule, error)
	// CreateSchedule creates a new named schedule and its associated slots/lessons
	CreateSchedule(ctx context.Context, schedule models.Schedule, slots []models.ScheduleSlotInput) (*models.Schedule, error)
	// UpdateSchedule updates the main schedule table and replaces its slots/lessons
	UpdateSchedule(ctx context.Context, scheduleID uuid.UUID, name *string, slots []models.ScheduleSlotInput) error
	// DeleteSchedule deletes a schedule and all its associated data
	DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error
	// GenerateSchedule generates a schedule based on study plans and workload (stub for now)
	GenerateSchedule(ctx context.Context, req models.GenerateScheduleRequest) ([]models.ScheduleDay, error)
}

type scheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

// GetSchedule loads the schedule for a specific teacher from the *active* named schedule
func (r *scheduleRepository) GetSchedule(ctx context.Context, teacherID uuid.UUID) ([]models.ScheduleDay, error) {
	// First, find the active schedule ID
	var activeScheduleID uuid.UUID
	err := r.db.QueryRowContext(ctx, `SELECT id FROM schedules WHERE is_active = true LIMIT 1`).Scan(&activeScheduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no active schedule exists, return empty schedule
			return []models.ScheduleDay{}, nil
		}
		return nil, fmt.Errorf("failed to find active schedule: %w", err)
	}

	// Now, query lessons for this teacher within the active schedule
	const q = `
		SELECT 
			ss.day_of_week,
			ss.lesson_number,
			sl.id as lesson_id,
			s.id as subject_id,
			s.name as subject_name,
			t.id as teacher_id,
			t.first_name,
			t.last_name,
			t.patronymic,
			cr.id as classroom_id,
			cr.name as classroom_name,
			c.id as class_id,
			c.name as class_name,
			cg.id as group_id,
			cg.name as group_name
		FROM schedule_slots ss
		JOIN schedule_lessons sl ON sl.slot_id = ss.id
		JOIN lesson_teachers lt ON lt.lesson_id = sl.id
		JOIN teachers t ON t.id = lt.teacher_id
		JOIN subjects s ON s.id = sl.subject_id
		JOIN lesson_participants lp ON lp.lesson_id = sl.id
		JOIN classes c ON c.id = lp.class_id
		LEFT JOIN lesson_rooms lr ON lr.lesson_id = sl.id
		LEFT JOIN classrooms cr ON cr.id = lr.classroom_id
		LEFT JOIN lesson_participant_groups lpg ON lp.id = lpg.participant_id
		LEFT JOIN class_groups cg ON lpg.group_id = cg.id
		WHERE ss.schedule_id = $1 AND t.id = $2
		ORDER BY ss.day_of_week, ss.lesson_number, sl.id, t.last_name, t.first_name, c.name, cg.name
	`

	rows, err := r.db.QueryContext(ctx, q, activeScheduleID, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// dayLessonKey → map[lessonNumber][]ScheduleLesson
	scheduleMap := make(map[string]map[int][]models.ScheduleLesson)

	for rows.Next() {
		var (
			dayOfWeek, lessonNumber int

			lessonID    uuid.UUID
			subjectID   uuid.UUID
			subjectName string

			teacherUUID                                          uuid.UUID
			teacherFirstName, teacherLastName, teacherPatronymic sql.NullString

			classroomID   sql.NullString
			classroomName sql.NullString

			classID   sql.NullString
			className sql.NullString

			groupID   sql.NullString
			groupName sql.NullString
		)

		if err := rows.Scan(
			&dayOfWeek, &lessonNumber,
			&lessonID, &subjectID, &subjectName,
			&teacherUUID, &teacherFirstName, &teacherLastName, &teacherPatronymic,
			&classroomID, &classroomName,
			&classID, &className,
			&groupID, &groupName,
		); err != nil {
			return nil, err
		}

		dayLessonKey := fmt.Sprintf("%d-%d", dayOfWeek, lessonNumber)

		dayMap, ok := scheduleMap[dayLessonKey]
		if !ok {
			dayMap = make(map[int][]models.ScheduleLesson)
			scheduleMap[dayLessonKey] = dayMap
		}

		lessons := dayMap[lessonNumber]

		// Find lesson in slice
		lessonIdx := -1
		for i, l := range lessons {
			if l.ID == lessonID {
				lessonIdx = i
				break
			}
		}

		if lessonIdx == -1 {
			lessons = append(lessons, models.ScheduleLesson{
				ID: lessonID,
				Subject: &models.Subject{
					ID:   subjectID,
					Name: subjectName,
				},
				Teachers:     []models.Teacher{},
				Rooms:        []models.Classroom{},
				Participants: []models.LessonParticipant{},
			})
			lessonIdx = len(lessons) - 1
		}

		lesson := &lessons[lessonIdx]

		// -------- Add teacher --------
		var patronymic *string
		if teacherPatronymic.Valid {
			p := teacherPatronymic.String
			patronymic = &p
		}

		lesson.Teachers = append(lesson.Teachers, models.Teacher{
			ID:         teacherUUID,
			FirstName:  teacherFirstName.String,
			LastName:   teacherLastName.String,
			Patronymic: patronymic,
		})

		// -------- Add room (with uuid.Parse) --------
		if classroomID.Valid {
			rid, err := uuid.Parse(classroomID.String)
			if err == nil {
				// avoid duplicates
				found := false
				for _, r := range lesson.Rooms {
					if r.ID == rid {
						found = true
						break
					}
				}
				if !found {
					lesson.Rooms = append(lesson.Rooms, models.Classroom{
						ID:   rid,
						Name: classroomName.String,
					})
				}
			}
		}

		// -------- Add participant (class + group) --------
		var participantClassUUID uuid.UUID
		if classID.Valid {
			participantClassUUID, _ = uuid.Parse(classID.String)
		}

		participantIdx := -1
		for i, p := range lesson.Participants {
			if p.ClassID == participantClassUUID {
				participantIdx = i
				break
			}
		}

		// If participant does not exist — create
		if participantIdx == -1 {
			lesson.Participants = append(lesson.Participants, models.LessonParticipant{
				ClassID: participantClassUUID,
				Class: &models.Class{
					ID:   participantClassUUID,
					Name: className.String,
				},
				GroupIDs: []uuid.UUID{},
			})
			participantIdx = len(lesson.Participants) - 1
		}

		// Add group if exists
		if groupID.Valid {
			gid, err := uuid.Parse(groupID.String)
			if err == nil {
				found := false
				for _, existing := range lesson.Participants[participantIdx].GroupIDs {
					if existing == gid {
						found = true
						break
					}
				}
				if !found {
					lesson.Participants[participantIdx].GroupIDs =
						append(lesson.Participants[participantIdx].GroupIDs, gid)
				}
			}
		}

		// Write back lesson slice
		dayMap[lessonNumber] = lessons
	}

	// -------- Now convert scheduleMap into []ScheduleDay --------
	// Secondary query to detect slots *only within the active schedule for this teacher*
	const lessonIDsQuery = `
		SELECT 
			ss.day_of_week,
			ss.lesson_number,
			sl.id
		FROM schedule_slots ss
		JOIN schedule_lessons sl ON sl.slot_id = ss.id
		JOIN lesson_teachers lt ON lt.lesson_id = sl.id
		WHERE ss.schedule_id = $1 AND lt.teacher_id = $2
		ORDER BY ss.day_of_week, ss.lesson_number
	`

	lessonRows, err := r.db.QueryContext(ctx, lessonIDsQuery, activeScheduleID, teacherID)
	if err != nil {
		return nil, err
	}
	defer lessonRows.Close()

	type slotLessons struct {
		dayOfWeek    int
		lessonNumber int
		lessons      []uuid.UUID
	}

	var slotData []slotLessons

	for lessonRows.Next() {
		var (
			day, num int
			lid      uuid.UUID
		)

		if err := lessonRows.Scan(&day, &num, &lid); err != nil {
			return nil, err
		}

		found := false
		for i := range slotData {
			if slotData[i].dayOfWeek == day && slotData[i].lessonNumber == num {
				slotData[i].lessons = append(slotData[i].lessons, lid)
				found = true
				break
			}
		}

		if !found {
			slotData = append(slotData, slotLessons{
				dayOfWeek:    day,
				lessonNumber: num,
				lessons:      []uuid.UUID{lid},
			})
		}
	}

	// Final result
	result := make([]models.ScheduleDay, len(slotData))

	for i, slot := range slotData {
		result[i].DayOfWeek = dayOfWeekToString(slot.dayOfWeek)
		result[i].LessonNumber = slot.lessonNumber

		// Load lessons by IDs
		lessons, err := r.loadLessonsForIDs(ctx, slot.lessons)
		if err != nil {
			return nil, fmt.Errorf("load lessons for slot %d-%d: %w",
				slot.dayOfWeek, slot.lessonNumber, err)
		}

		result[i].Lessons = lessons
	}

	return result, nil
}

// Helper to convert internal day number to string
func dayOfWeekToString(day int) string {
	switch day {
	case 1:
		return "monday"
	case 2:
		return "tuesday"
	case 3:
		return "wednesday"
	case 4:
		return "thursday"
	case 5:
		return "friday"
	// Add more if Saturday/Sunday are needed
	default:
		return "unknown"
	}
}

// loadLessonsForIDs fetches detailed lesson information for a list of lesson IDs
func (r *scheduleRepository) loadLessonsForIDs(ctx context.Context, lessonIDs []uuid.UUID) ([]models.ScheduleLesson, error) {
	if len(lessonIDs) == 0 {
		return []models.ScheduleLesson{}, nil
	}

	// Use a placeholder query builder for IN clause
	placeholders := make([]string, len(lessonIDs))
	args := make([]interface{}, len(lessonIDs))
	for i, id := range lessonIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	inClause := "(" + fmt.Sprintf("%s", placeholders[0])
	for i := 1; i < len(placeholders); i++ {
		inClause += ", " + placeholders[i]
	}
	inClause += ")"

	q := fmt.Sprintf(`
		SELECT 
			sl.id, sl.subject_id, s.name as subject_name
		FROM schedule_lessons sl
		JOIN subjects s ON s.id = sl.subject_id
		WHERE sl.id IN %s
	`, inClause)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessonsMap := make(map[uuid.UUID]*models.ScheduleLesson)
	var lessonOrder []uuid.UUID // To maintain order based on input IDs

	for rows.Next() {
		var lessonID, subjectID uuid.UUID
		var subjectName string
		if err := rows.Scan(&lessonID, &subjectID, &subjectName); err != nil {
			return nil, err
		}

		lesson := &models.ScheduleLesson{
			ID: lessonID,
			Subject: &models.Subject{
				ID:   subjectID,
				Name: subjectName,
			},
			Teachers:     []models.Teacher{},
			Rooms:        []models.Classroom{},
			Participants: []models.LessonParticipant{},
		}
		lessonsMap[lessonID] = lesson
		lessonOrder = append(lessonOrder, lessonID)
	}

	// Load teachers for these lessons
	teachers, err := r.loadTeachersForLessons(ctx, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("load teachers for lessons: %w", err)
	}
	for lessonID, tList := range teachers {
		if lesson, ok := lessonsMap[lessonID]; ok {
			lesson.Teachers = tList
		}
	}

	// Load rooms for these lessons
	rooms, err := r.loadRoomsForLessons(ctx, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("load rooms for lessons: %w", err)
	}
	for lessonID, rList := range rooms {
		if lesson, ok := lessonsMap[lessonID]; ok {
			lesson.Rooms = rList
		}
	}

	// Load participants for these lessons
	participants, err := r.loadParticipantsForLessons(ctx, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("load participants for lessons: %w", err)
	}
	for lessonID, pList := range participants {
		if lesson, ok := lessonsMap[lessonID]; ok {
			lesson.Participants = pList
		}
	}

	// Return lessons in the order they were requested
	var result []models.ScheduleLesson
	for _, id := range lessonOrder {
		if lesson, ok := lessonsMap[id]; ok {
			result = append(result, *lesson)
		}
	}

	return result, nil
}

// loadTeachersForLessons fetches teachers assigned to a list of lessons
func (r *scheduleRepository) loadTeachersForLessons(ctx context.Context, lessonIDs []uuid.UUID) (map[uuid.UUID][]models.Teacher, error) {
	if len(lessonIDs) == 0 {
		return make(map[uuid.UUID][]models.Teacher), nil
	}

	placeholders := make([]string, len(lessonIDs))
	args := make([]interface{}, len(lessonIDs))
	for i, id := range lessonIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	inClause := "(" + fmt.Sprintf("%s", placeholders[0])
	for i := 1; i < len(placeholders); i++ {
		inClause += ", " + placeholders[i]
	}
	inClause += ")"

	q := fmt.Sprintf(`
		SELECT 
			lt.lesson_id, t.id, t.first_name, t.last_name, t.patronymic
		FROM lesson_teachers lt
		JOIN teachers t ON t.id = lt.teacher_id
		WHERE lt.lesson_id IN %s
		ORDER BY lt.lesson_id, t.last_name, t.first_name
	`, inClause)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teachersMap := make(map[uuid.UUID][]models.Teacher)
	for rows.Next() {
		var lessonID, teacherID uuid.UUID
		var firstName, lastName string
		var patronymic sql.NullString

		if err := rows.Scan(&lessonID, &teacherID, &firstName, &lastName, &patronymic); err != nil {
			return nil, err
		}

		var patronymicPtr *string
		if patronymic.Valid {
			patronymicPtr = &patronymic.String
		}

		teacher := models.Teacher{
			ID:         teacherID,
			FirstName:  firstName,
			LastName:   lastName,
			Patronymic: patronymicPtr,
		}

		teachersMap[lessonID] = append(teachersMap[lessonID], teacher)
	}

	return teachersMap, nil
}

// loadRoomsForLessons fetches rooms assigned to a list of lessons
func (r *scheduleRepository) loadRoomsForLessons(ctx context.Context, lessonIDs []uuid.UUID) (map[uuid.UUID][]models.Classroom, error) {
	if len(lessonIDs) == 0 {
		return make(map[uuid.UUID][]models.Classroom), nil
	}

	placeholders := make([]string, len(lessonIDs))
	args := make([]interface{}, len(lessonIDs))
	for i, id := range lessonIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	inClause := "(" + fmt.Sprintf("%s", placeholders[0])
	for i := 1; i < len(placeholders); i++ {
		inClause += ", " + placeholders[i]
	}
	inClause += ")"

	q := fmt.Sprintf(`
		SELECT 
			lr.lesson_id, cr.id, cr.name
		FROM lesson_rooms lr
		JOIN classrooms cr ON cr.id = lr.classroom_id
		WHERE lr.lesson_id IN %s
		ORDER BY lr.lesson_id, cr.name
	`, inClause)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roomsMap := make(map[uuid.UUID][]models.Classroom)
	for rows.Next() {
		var lessonID, roomID uuid.UUID
		var roomName string

		if err := rows.Scan(&lessonID, &roomID, &roomName); err != nil {
			return nil, err
		}

		room := models.Classroom{
			ID:   roomID,
			Name: roomName,
		}

		roomsMap[lessonID] = append(roomsMap[lessonID], room)
	}

	return roomsMap, nil
}

// loadParticipantsForLessons fetches participants (classes/groups) for a list of lessons
func (r *scheduleRepository) loadParticipantsForLessons(ctx context.Context, lessonIDs []uuid.UUID) (map[uuid.UUID][]models.LessonParticipant, error) {
	if len(lessonIDs) == 0 {
		return make(map[uuid.UUID][]models.LessonParticipant), nil
	}

	placeholders := make([]string, len(lessonIDs))
	args := make([]interface{}, len(lessonIDs))
	for i, id := range lessonIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	inClause := "(" + fmt.Sprintf("%s", placeholders[0])
	for i := 1; i < len(placeholders); i++ {
		inClause += ", " + placeholders[i]
	}
	inClause += ")"

	q := fmt.Sprintf(`
		SELECT 
			lp.lesson_id, lp.class_id, c.name as class_name
		FROM lesson_participants lp
		JOIN classes c ON c.id = lp.class_id
		WHERE lp.lesson_id IN %s
		ORDER BY lp.lesson_id, c.name
	`, inClause)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participantsMap := make(map[uuid.UUID][]models.LessonParticipant)
	for rows.Next() {
		var lessonID, classID uuid.UUID
		var className string

		if err := rows.Scan(&lessonID, &classID, &className); err != nil {
			return nil, err
		}

		participant := models.LessonParticipant{
			LessonID: lessonID,
			ClassID:  classID,
			Class: &models.Class{
				ID:   classID,
				Name: className,
			},
			GroupIDs: []uuid.UUID{}, // Will populate next
		}

		participantsMap[lessonID] = append(participantsMap[lessonID], participant)
	}

	// Now load the group IDs for these participants
	// Get participant IDs first
	var participantIDs []uuid.UUID
	for _, pList := range participantsMap {
		for _, p := range pList {
			participantIDs = append(participantIDs, p.ID) // Assuming p.ID is the participant ID from lesson_participants table
		}
	}

	// If participant IDs exist, load groups
	if len(participantIDs) > 0 {
		placeholdersP := make([]string, len(participantIDs))
		argsP := make([]interface{}, len(participantIDs))
		for i, id := range participantIDs {
			placeholdersP[i] = fmt.Sprintf("$%d", i+1)
			argsP[i] = id
		}
		inClauseP := "(" + fmt.Sprintf("%s", placeholdersP[0])
		for i := 1; i < len(placeholdersP); i++ {
			inClauseP += ", " + placeholdersP[i]
		}
		inClauseP += ")"

		qGroups := fmt.Sprintf(`
			SELECT 
				participant_id, group_id
			FROM lesson_participant_groups
			WHERE participant_id IN %s
			ORDER BY participant_id, group_id
		`, inClauseP)

		rowsGroups, err := r.db.QueryContext(ctx, qGroups, argsP...)
		if err != nil {
			return nil, err
		}
		defer rowsGroups.Close()

		groupsMap := make(map[uuid.UUID][]uuid.UUID)
		for rowsGroups.Next() {
			var participantID, groupID uuid.UUID
			if err := rowsGroups.Scan(&participantID, &groupID); err != nil {
				return nil, err
			}
			groupsMap[participantID] = append(groupsMap[participantID], groupID)
		}

		// Assign groups back to participants
		for lessonID, pList := range participantsMap {
			for j, p := range pList {
				if groupIDs, ok := groupsMap[p.ID]; ok { // Assuming p.ID is the primary key of lesson_participants
					participantsMap[lessonID][j].GroupIDs = groupIDs
				}
			}
		}
	}

	return participantsMap, nil
}

// GetScheduleByID loads a specific named schedule by its ID
func (r *scheduleRepository) GetScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.Schedule, error) {
	// For this endpoint, we might just return the schedule header info
	// and let the frontend call GET /schedule for the actual data if needed.
	// Or load slots/lessons. Let's load the header for now as per spec.
	const q = `SELECT id, name, academic_year, is_active, created_at, updated_at FROM schedules WHERE id = $1`
	row := r.db.QueryRowContext(ctx, q, scheduleID)

	var s models.Schedule
	var academicYear sql.NullString
	err := row.Scan(&s.ID, &s.Name, &academicYear, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	if academicYear.Valid {
		s.AcademicYear = &academicYear.String
	}

	return &s, nil
}

// GetAllSchedules loads all named schedules
func (r *scheduleRepository) GetAllSchedules(ctx context.Context) ([]models.Schedule, error) {
	const q = `SELECT id, name, academic_year, is_active, created_at, updated_at FROM schedules ORDER BY name`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var s models.Schedule
		var academicYear sql.NullString
		if err := rows.Scan(&s.ID, &s.Name, &academicYear, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if academicYear.Valid {
			s.AcademicYear = &academicYear.String
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

// CreateSchedule creates a new named schedule and its associated slots/lessons
func (r *scheduleRepository) CreateSchedule(ctx context.Context, schedule models.Schedule, slots []models.ScheduleSlotInput) (*models.Schedule, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Insert into schedules table
	var newID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO schedules (name, academic_year, is_active)
		VALUES ($1, $2, $3)
		RETURNING id
	`, schedule.Name, schedule.AcademicYear, schedule.IsActive).Scan(&newID)
	if err != nil {
		return nil, err
	}

	// 2. Process slots and lessons within the transaction
	for _, slotInput := range slots {
		dayNum := stringToDayOfWeek(slotInput.DayOfWeek)
		if dayNum == 0 {
			err = fmt.Errorf("invalid day of week: %s", slotInput.DayOfWeek)
			return nil, err
		}

		var slotID uuid.UUID
		err = tx.QueryRowContext(ctx, `
			INSERT INTO schedule_slots (schedule_id, day_of_week, lesson_number)
			VALUES ($1, $2, $3)
			RETURNING id
		`, newID, dayNum, slotInput.LessonNumber).Scan(&slotID)
		if err != nil {
			return nil, err
		}

		for _, lessonInput := range slotInput.Lessons {
			subjectID, err := uuid.Parse(lessonInput.Subject.ID)
			if err != nil {
				return nil, fmt.Errorf("invalid subject ID: %w", err)
			}

			var lessonID uuid.UUID
			err = tx.QueryRowContext(ctx, `
				INSERT INTO schedule_lessons (slot_id, subject_id)
				VALUES ($1, $2)
				RETURNING id
			`, slotID, subjectID).Scan(&lessonID)
			if err != nil {
				return nil, err
			}

			// Insert lesson_teachers
			for _, teacherInput := range lessonInput.Teachers {
				teacherID, err := uuid.Parse(teacherInput.ID)
				if err != nil {
					return nil, fmt.Errorf("invalid teacher ID: %w", err)
				}
				_, err = tx.ExecContext(ctx, `
					INSERT INTO lesson_teachers (lesson_id, teacher_id)
					VALUES ($1, $2)
				`, lessonID, teacherID)
				if err != nil {
					return nil, err
				}
			}

			// Insert lesson_rooms
			for _, roomInput := range lessonInput.Rooms {
				roomID, err := uuid.Parse(roomInput.ID)
				if err != nil {
					return nil, fmt.Errorf("invalid room ID: %w", err)
				}
				_, err = tx.ExecContext(ctx, `
					INSERT INTO lesson_rooms (lesson_id, classroom_id)
					VALUES ($1, $2)
				`, lessonID, roomID)
				if err != nil {
					return nil, err
				}
			}

			// Insert lesson_participants and lesson_participant_groups
			for _, participantInput := range lessonInput.Participants {
				classID, err := uuid.Parse(participantInput.Class.ID)
				if err != nil {
					return nil, fmt.Errorf("invalid class ID: %w", err)
				}

				var participantID uuid.UUID
				err = tx.QueryRowContext(ctx, `
					INSERT INTO lesson_participants (lesson_id, class_id)
					VALUES ($1, $2)
					RETURNING id
				`, lessonID, classID).Scan(&participantID)
				if err != nil {
					return nil, err
				}

				for _, groupIDStr := range participantInput.GroupIDs {
					groupID, err := uuid.Parse(groupIDStr)
					if err != nil {
						return nil, fmt.Errorf("invalid group ID: %w", err)
					}
					_, err = tx.ExecContext(ctx, `
						INSERT INTO lesson_participant_groups (participant_id, group_id)
						VALUES ($1, $2)
					`, participantID, groupID)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return the created schedule
	createdSchedule, err := r.GetScheduleByID(ctx, newID)
	if err != nil {
		// This should not happen if commit was successful
		return nil, err
	}
	return createdSchedule, nil
}

// stringToDayOfWeek converts string day to internal integer (1-6), case-insensitive
func stringToDayOfWeek(day string) int {
	lowerDay := strings.ToLower(day)
	switch lowerDay {
	case "monday":
		return 1
	case "tuesday":
		return 2
	case "wednesday":
		return 3
	case "thursday":
		return 4
	case "friday":
		return 5
	case "saturday":
		return 6
	default:
		return 0
	}
}

// UpdateSchedule updates the main schedule table and replaces its slots/lessons
func (r *scheduleRepository) UpdateSchedule(ctx context.Context, scheduleID uuid.UUID, name *string, slots []models.ScheduleSlotInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Update schedule name if provided
	if name != nil {
		_, err = tx.ExecContext(ctx, `UPDATE schedules SET name = $1 WHERE id = $2`, *name, scheduleID)
		if err != nil {
			return err
		}
	}

	// 2. Delete existing slots/lessons for this schedule (CASCADE should handle children)
	_, err = tx.ExecContext(ctx, `DELETE FROM schedule_slots WHERE schedule_id = $1`, scheduleID)
	if err != nil {
		return err
	}

	// 3. Insert new slots and lessons (same logic as Create)
	for _, slotInput := range slots {
		// Use stringToDayOfWeek which is now case-insensitive
		dayNum := stringToDayOfWeek(slotInput.DayOfWeek)
		if dayNum == 0 {
			err = fmt.Errorf("invalid day of week: %s", slotInput.DayOfWeek)
			return err
		}

		var slotID uuid.UUID
		err = tx.QueryRowContext(ctx, `
			INSERT INTO schedule_slots (schedule_id, day_of_week, lesson_number)
			VALUES ($1, $2, $3)
			RETURNING id
		`, scheduleID, dayNum, slotInput.LessonNumber).Scan(&slotID)
		if err != nil {
			return err
		}

		for _, lessonInput := range slotInput.Lessons {
			subjectID, err := uuid.Parse(lessonInput.Subject.ID)
			if err != nil {
				return fmt.Errorf("invalid subject ID: %w", err)
			}

			var lessonID uuid.UUID
			err = tx.QueryRowContext(ctx, `
				INSERT INTO schedule_lessons (slot_id, subject_id)
				VALUES ($1, $2)
				RETURNING id
			`, slotID, subjectID).Scan(&lessonID)
			if err != nil {
				return err
			}

			// Insert lesson_teachers
			for _, teacherInput := range lessonInput.Teachers {
				teacherID, err := uuid.Parse(teacherInput.ID)
				if err != nil {
					return fmt.Errorf("invalid teacher ID: %w", err)
				}
				_, err = tx.ExecContext(ctx, `
					INSERT INTO lesson_teachers (lesson_id, teacher_id)
					VALUES ($1, $2)
				`, lessonID, teacherID)
				if err != nil {
					return err
				}
			}

			// Insert lesson_rooms
			for _, roomInput := range lessonInput.Rooms {
				roomID, err := uuid.Parse(roomInput.ID)
				if err != nil {
					return fmt.Errorf("invalid room ID: %w", err)
				}
				_, err = tx.ExecContext(ctx, `
					INSERT INTO lesson_rooms (lesson_id, classroom_id)
					VALUES ($1, $2)
				`, lessonID, roomID)
				if err != nil {
					return err
				}
			}

			// Insert lesson_participants and lesson_participant_groups
			for _, participantInput := range lessonInput.Participants {
				classID, err := uuid.Parse(participantInput.Class.ID)
				if err != nil {
					return fmt.Errorf("invalid class ID: %w", err)
				}

				var participantID uuid.UUID
				err = tx.QueryRowContext(ctx, `
					INSERT INTO lesson_participants (lesson_id, class_id)
					VALUES ($1, $2)
					RETURNING id
				`, lessonID, classID).Scan(&participantID)
				if err != nil {
					return err
				}

				for _, groupIDStr := range participantInput.GroupIDs {
					groupID, err := uuid.Parse(groupIDStr)
					if err != nil {
						return fmt.Errorf("invalid group ID: %w", err)
					}
					_, err = tx.ExecContext(ctx, `
						INSERT INTO lesson_participant_groups (participant_id, group_id)
						VALUES ($1, $2)
					`, participantID, groupID)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// DeleteSchedule deletes a schedule and all its associated data
func (r *scheduleRepository) DeleteSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM schedules WHERE id = $1`, scheduleID)
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

// GenerateSchedule generates a schedule based on study plans and workload (stub for now)
func (r *scheduleRepository) GenerateSchedule(ctx context.Context, req models.GenerateScheduleRequest) ([]models.ScheduleDay, error) {
	// This is a complex business logic that would require:
	// 1. Loading study plans (what classes need which subjects)
	// 2. Loading teacher workload (who can teach what)
	// 3. Loading available classrooms
	// 4. Running an algorithm (e.g., greedy, genetic) to assign lessons to slots
	// 5. Checking constraints (no teacher/class/group conflicts)
	// For now, return an empty schedule as a stub.
	return []models.ScheduleDay{}, nil
}
