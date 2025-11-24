package models

import (
	"github.com/google/uuid"
	"time"
)

// User represents the base user in the system
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Phone        *string   `json:"phone,omitempty" db:"phone"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never send to frontend
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

func (u *User) DisplayName() string {
	return u.Email
}

// Teacher represents a teacher in the system
type Teacher struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	UserID               *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	FirstName            string     `json:"firstName" db:"first_name"`
	LastName             string     `json:"lastName" db:"last_name"`
	Patronymic           *string    `json:"patronymic,omitempty" db:"patronymic"`
	ClassroomID          *uuid.UUID `json:"-" db:"classroom_id"`      // Internal, not for frontend
	Classroom            *Classroom `json:"classRoom,omitempty"`      // Expanded for frontend
	HomeroomClassID      *uuid.UUID `json:"-" db:"homeroom_class_id"` // Internal, not for frontend
	HomeroomClass        *Class     `json:"class,omitempty"`          // Expanded for frontend
	WorkloadHoursPerWeek int        `json:"workloadHoursPerWeek" db:"workload_hours_per_week"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	// Expanded fields for frontend
	Subjects   []TeacherSubjectAssignment `json:"subjects"`
	ClassHours []TeacherClassHour         `json:"classHours"`
}

// LightTeacher is a simplified version of Teacher for dropdowns/lists
type LightTeacher struct {
	ID         uuid.UUID `json:"id"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Patronymic *string   `json:"patronymic,omitempty"`
}

// Classroom represents a classroom
type Classroom struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Capacity  *int      `json:"capacity,omitempty" db:"capacity"`
	Equipment *string   `json:"equipment,omitempty" db:"equipment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Subject represents a school subject
type Subject struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	ShortName *string   `json:"shortName,omitempty" db:"short_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Class represents a school class (e.g., "5A", "11Б")
type Class struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Name              string     `json:"name" db:"name"`
	GradeLevel        int        `json:"gradeLevel" db:"grade_level"`
	TotalStudents     *int       `json:"totalStudents,omitempty" db:"total_students"`
	HomeroomTeacherID *uuid.UUID `json:"-" db:"homeroom_teacher_id"` // Internal, not for frontend
	HomeroomTeacher   *Teacher   `json:"classTeacher,omitempty"`     // Expanded for frontend
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	// Expanded fields for frontend
	Subjects []ClassSubjectAssignment `json:"subjects"`
	Groups   []ClassGroup             `json:"groups"`
}

// ClassGroup represents a subgroup within a class (e.g., "Group 1", "Boys", "Girls")
type ClassGroup struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ClassID     uuid.UUID `json:"-" db:"class_id"`
	Name        string    `json:"name" db:"name"`
	Size        *int      `json:"size,omitempty" db:"size"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// StudyPlan represents the study plan for a class (which subjects and how many hours per week)
type StudyPlan struct {
	ID                uuid.UUID `json:"id" db:"id"`
	ClassID           uuid.UUID `json:"-" db:"class_id"`
	SubjectID         uuid.UUID `json:"-" db:"subject_id"`
	Subject           *Subject  `json:"subject"` // Expanded for frontend
	HoursPerWeek      int       `json:"hoursPerWeek" db:"hours_per_week"`
	SplitEnabled      bool      `json:"splitEnabled" db:"split_enabled"`
	SplitGroupsCount  *int      `json:"splitGroupsCount,omitempty" db:"split_groups_count"`
	CrossClassAllowed bool      `json:"crossClassAllowed,omitempty" db:"cross_class_allowed"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// TeacherSubject represents the qualification of a teacher (which subjects they can teach)
type TeacherSubject struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	TeacherID             uuid.UUID `json:"-" db:"teacher_id"`
	SubjectID             uuid.UUID `json:"-" db:"subject_id"`
	Subject               *Subject  `json:"subject"` // Expanded for frontend
	PreferredHoursPerWeek *int      `json:"hoursPerWeek,omitempty" db:"preferred_hours_per_week"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// TeacherWorkload represents the actual workload distribution
type TeacherWorkload struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	TeacherID    uuid.UUID   `json:"-" db:"teacher_id"`
	ClassID      uuid.UUID   `json:"-" db:"class_id"`
	SubjectID    uuid.UUID   `json:"-" db:"subject_id"`
	GroupID      *uuid.UUID  `json:"-" db:"group_id"`
	Group        *ClassGroup `json:"group,omitempty"` // Expanded for frontend
	HoursPerWeek int         `json:"hours" db:"hours_per_week"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// Schedule represents a named schedule (e.g., "Main Schedule", "Winter Schedule")
type Schedule struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	AcademicYear *string   `json:"academicYear,omitempty" db:"academic_year"`
	IsActive     bool      `json:"isActive" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ScheduleSlot represents a time slot in the schedule (day + lesson number)
type ScheduleSlot struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ScheduleID   uuid.UUID `json:"-" db:"schedule_id"`
	DayOfWeek    int       `json:"dayOfWeek" db:"day_of_week"` // 1=Monday, 2=Tuesday, ..., 5=Friday
	LessonNumber int       `json:"lessonNumber" db:"lesson_number"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ScheduleLesson represents a lesson in the schedule
type ScheduleLesson struct {
	ID        uuid.UUID `json:"id" db:"id"`
	SlotID    uuid.UUID `json:"-" db:"slot_id"`
	SubjectID uuid.UUID `json:"-" db:"subject_id"`
	Subject   *Subject  `json:"subject"` // Expanded for frontend
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	// Expanded fields for frontend
	Teachers     []Teacher           `json:"teachers"`
	Rooms        []Classroom         `json:"rooms"`
	Participants []LessonParticipant `json:"participants"`
}

// LessonParticipant represents which class/group participates in a lesson
type LessonParticipant struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	LessonID  uuid.UUID   `json:"-" db:"lesson_id"`
	ClassID   uuid.UUID   `json:"-" db:"class_id"`
	Class     *Class      `json:"class"`              // Expanded for frontend
	GroupIDs  []uuid.UUID `json:"groupIds,omitempty"` // If empty, means the whole class
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

// TeacherSubjectAssignment is a helper struct for Teacher's subjects list
type TeacherSubjectAssignment struct {
	Subject      Subject `json:"subject"`
	HoursPerWeek *int    `json:"hoursPerWeek"`
}

// TeacherClassHour is a helper struct for Teacher's class hours list
type TeacherClassHour struct {
	Class   Class   `json:"class"`
	Subject Subject `json:"subject"`
	GroupID *string `json:"groupId,omitempty"` // Using string ID for frontend
	Hours   int     `json:"hours"`
}

// ClassSubjectAssignment is a helper struct for Class's subjects list
type ClassSubjectAssignment struct {
	Subject      Subject            `json:"subject"`
	HoursPerWeek int                `json:"hoursPerWeek"`
	Split        *ClassSubjectSplit `json:"split,omitempty"`
}

// ClassSubjectSplit represents splitting settings for a subject
type ClassSubjectSplit struct {
	GroupsCount       int   `json:"groupsCount"`
	CrossClassAllowed *bool `json:"crossClassAllowed,omitempty"`
}

// ScheduleDay represents a day in the schedule response
type ScheduleDay struct {
	DayOfWeek    string           `json:"dayOfWeek"`
	LessonNumber int              `json:"lessonNumber"`
	Lessons      []ScheduleLesson `json:"lessons"`
}

// ScheduleSlotInput represents input for creating/updating schedule
type ScheduleSlotInput struct {
	DayOfWeek    string        `json:"dayOfWeek"`
	LessonNumber int           `json:"lessonNumber"`
	Lessons      []LessonInput `json:"lessons"`
}

// LessonInput represents input for a lesson
type LessonInput struct {
	Subject      SubjectInput       `json:"subject"`
	Teachers     []TeacherInput     `json:"teachers"`
	Rooms        []ClassroomInput   `json:"rooms"`
	Participants []ParticipantInput `json:"participants"`
}

// SubjectInput represents input for a subject
type SubjectInput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TeacherInput represents input for a teacher
type TeacherInput struct {
	ID         string  `json:"id"`
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Patronymic *string `json:"patronymic"`
}

// ClassroomInput represents input for a classroom
type ClassroomInput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ParticipantInput represents input for a participant
type ParticipantInput struct {
	Class    ClassInput `json:"class"`
	GroupIDs []string   `json:"groupIds,omitempty"`
}

// ClassInput represents input for a class
type ClassInput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse represents the login response body
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"user"`
}

// RefreshRequest represents the refresh token request body
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshResponse represents the refresh token response body
type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// ConflictDetail represents a conflict in the schedule
type ConflictDetail struct {
	Type         string `json:"type"`
	Message      string `json:"message"`
	DayOfWeek    string `json:"dayOfWeek"`
	LessonNumber int    `json:"lessonNumber"`
}

// ConflictResponse represents a schedule conflict response
type ConflictResponse struct {
	Error   string           `json:"error"`
	Details []ConflictDetail `json:"details"`
}

// GenerateScheduleRequest represents the request body for schedule generation
type GenerateScheduleRequest struct {
	Algorithm        *string `json:"algorithm,omitempty"`
	MaxLessonsPerDay *int    `json:"maxLessonsPerDay,omitempty"`
	Priorities       *struct {
		BalanceWorkload *bool `json:"balanceWorkload,omitempty"`
		MinimizeGaps    *bool `json:"minimizeGaps,omitempty"`
	} `json:"priorities,omitempty"`
}

// BulkUpdateClassesRequest represents the request body for bulk class update
type BulkUpdateClassesRequest struct {
	Data []Class `json:"data"`
}

// BulkUpdateTeachersRequest represents the request body for bulk teacher update
type BulkUpdateTeachersRequest struct {
	Data []Teacher `json:"data"`
}

// BulkUpdateResponse represents the response for bulk update operations
type BulkUpdateResponse struct {
	Message string `json:"message"`
	Updated int    `json:"updated"`
}
