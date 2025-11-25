package handlers

import (
	"database/sql"
	"net/http"

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

// GetScheduleForTeacher implements ep: GET /schedule
// Assumes teacher ID is available in the context from middleware.
func (h *ScheduleHandler) GetScheduleForTeacher(c *gin.Context) {
	teacherID, exists := c.Get("teacherID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx := c.Request.Context()
	schedule, err := h.service.GetScheduleForTeacher(ctx, teacherID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": schedule})
}

// UpdateScheduleForTeacher implements ep: PUT /schedule
// Assumes the ID of the schedule to update is determined by the service (e.g., active schedule).
func (h *ScheduleHandler) UpdateScheduleForTeacher(c *gin.Context) {
	var payload struct {
		Data []models.ScheduleSlotInput `json:"data"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload structure, expected { data: [...] }"})
		return
	}

	ctx := c.Request.Context()

	// Find the active schedule ID to update
	allSchedules, err := h.service.GetAllSchedules(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to identify schedule to update"})
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
		c.JSON(http.StatusConflict, gin.H{"error": "no active schedule found to update"})
		return
	}

	err = h.service.UpdateSchedule(ctx, activeScheduleID, nil, payload.Data) // nil name means don't update name
	if err != nil {
		// Check for specific conflict errors if needed, based on repo logic
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Расписание успешно сохранено"})
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
			"error":  "Невозможно сгенерировать расписание",
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schedule"})
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

	ctx := c.Request.Context()
	newSchedule := models.Schedule{
		Name: req.Name,
		// Set other fields as necessary, e.g., IsActive, AcademicYear from context/body
		IsActive: false, // Usually not active when created
	}
	created, err := h.service.CreateSchedule(ctx, newSchedule, req.ScheduleSlots)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create schedule"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete schedule"})
		return
	}
	c.Status(http.StatusNoContent)
}
