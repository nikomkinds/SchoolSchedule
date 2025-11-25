package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/services"
)

type TeacherHandler struct {
	service services.TeacherService
}

func NewTeacherHandler(service services.TeacherService) *TeacherHandler {
	return &TeacherHandler{service: service}
}

// GetAllFull implements ep: GET /users/Teachers
func (h *TeacherHandler) GetAllFull(c *gin.Context) {
	ctx := c.Request.Context()
	teachers, err := h.service.GetAllFull(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load teachers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": teachers})
}

// GetAllLight implements ep: GET /users/LightTeachers
func (h *TeacherHandler) GetAllLight(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.service.GetAllLight(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load teachers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// Create implements ep: POST /users/Teachers
func (h *TeacherHandler) Create(c *gin.Context) {
	var req models.Teacher
	if err := c.ShouldBindJSON(&req); err != nil {
		// Expecting at least firstName and lastName
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := c.Request.Context()
	created, err := h.service.Create(ctx, req.FirstName, req.LastName, req.Patronymic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create teacher"})
		return
	}

	// Response envelope per spec: { data: { ... } }
	if created.Subjects == nil {
		created.Subjects = []models.TeacherSubjectAssignment{}
	}
	if created.ClassHours == nil {
		created.ClassHours = []models.TeacherClassHour{}
	}
	c.JSON(http.StatusCreated, gin.H{"data": created})
}

// Delete implements ep: DELETE /users/Teachers/:id
func (h *TeacherHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.Delete(ctx, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teacher not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

// BulkUpdate implements ep: PATCH /users/Teachers/bulk
func (h *TeacherHandler) BulkUpdate(c *gin.Context) {
	var req models.BulkUpdateTeachersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := c.Request.Context()
	updated, err := h.service.BulkUpdate(ctx, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update teachers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Teachers successfully updated",
		"updated": updated,
	})
}
