package handlers

import (
	"database/sql"
	"net/http"
	"strings" // Добавлен импорт strings

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/services"
)

type ScheduleHandler struct {
	service services.ScheduleService
}

func NewScheduleHandler(service services.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

// GetSchedule implements ep: GET /schedule
func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx := c.Request.Context()
	schedule, err := h.service.GetSchedule(ctx, uuid.MustParse(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schedule", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": schedule})
}

// UpdateScheduleForTeacher implements ep: PUT /schedule
func (h *ScheduleHandler) UpdateScheduleForTeacher(c *gin.Context) {
	var payload struct {
		Data []models.ScheduleSlotInput `json:"data"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload structure, expected {  [...] }"})
		return
	}

	// Карта для сопоставления дня недели из строки в число
	dayMap := map[string]int{
		"monday":    1,
		"tuesday":   2,
		"wednesday": 3,
		"thursday":  4,
		"friday":    5,
		"saturday":  6,
		// "sunday": 7, // Если нужно, добавьте
	}

	// Проходим по всем слотам в payload и устанавливаем DayOfWeekInt
	// И приводим DayOfWeek к нижнему регистру для совместимости с репозиторием
	for i := range payload.Data {
		// Приводим строку к нижнему регистру для сопоставления
		day := strings.ToLower(payload.Data[i].DayOfWeek)

		val, ok := dayMap[day]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid day_of_week",
				"value":   payload.Data[i].DayOfWeek,
				"allowed": []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday"},
			})
			return
		}

		// Устанавливаем внутреннее числовое значение
		payload.Data[i].DayOfWeekInt = val
		// Приводим строковое поле к нижнему регистру для передачи в репозиторий
		payload.Data[i].DayOfWeek = day
	}

	ctx := c.Request.Context()

	// Get userID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userUUID := uuid.MustParse(userID)

	allSchedules, err := h.service.GetAllSchedules(ctx, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to identify schedule to update", "details": err.Error()})
		return
	}

	var activeScheduleID uuid.UUID
	for _, s := range allSchedules {
		if s.IsActive {
			activeScheduleID = s.ID
			break
		}
	}

	if activeScheduleID == uuid.Nil {
		newSchedule := models.Schedule{Name: "Расписание", IsActive: true}
		created, err := h.service.CreateSchedule(ctx, userUUID, newSchedule, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create initial schedule", "details": err.Error()})
			return
		}
		activeScheduleID = created.ID
	}

	// Передаем в сервис уже обновленный payload.Data, где DayOfWeekInt заполнен и DayOfWeek в нижнем регистре
	err = h.service.UpdateSchedule(ctx, activeScheduleID, nil, payload.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update schedule", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schedule successfully saved"})
}

// GenerateSchedule implements ep: POST /schedule/generate
func (h *ScheduleHandler) GenerateSchedule(c *gin.Context) {
	var req models.GenerateScheduleRequest
	// The spec shows the body can be optional, so we bind with ShouldBindJSON which handles empty bodies
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := c.Request.Context()
	schedule, err := h.service.GenerateSchedule(ctx, req)
	if err != nil {
		// Could return a 400 with conflict details if generation fails due to constraints
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Unable to generate schedule",
			"reason": err.Error(), // Or a more generic message
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": schedule})
}

// GetScheduleByID implements ep: GET /schedule/:id
func (h *ScheduleHandler) GetScheduleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	schedule, err := h.service.GetScheduleByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schedule", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": schedule})
}

// CreateSchedule implements ep: POST /schedule
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var req struct {
		Name          string                     `json:"name"`
		ScheduleSlots []models.ScheduleSlotInput `json:"scheduleSlots"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Get userID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userUUID := uuid.MustParse(userID)

	ctx := c.Request.Context()
	newSchedule := models.Schedule{
		Name: req.Name,
		// Set other fields as necessary, e.g., IsActive, AcademicYear from context/body
		IsActive: false, // Usually not active when created
	}
	created, err := h.service.CreateSchedule(ctx, userUUID, newSchedule, req.ScheduleSlots)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create schedule", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": created})
}

// DeleteSchedule implements ep: DELETE /schedule/:id
func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.DeleteSchedule(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete schedule", "details": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
