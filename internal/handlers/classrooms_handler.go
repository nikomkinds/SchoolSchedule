package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/services"
)

type ClassroomHandler struct {
	service *services.ClassroomService
}

func NewClassroomHandler(service *services.ClassroomService) *ClassroomHandler {
	return &ClassroomHandler{service: service}
}

// GetAll implements ep: GET /classrooms
func (h *ClassroomHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	classrooms, err := h.service.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch classrooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": classrooms})
}

// Create implements ep: POST /classrooms
func (h *ClassroomHandler) Create(c *gin.Context) {
	var req models.CreateClassroomRequest

	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.service.Create(ctx, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create classroom"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": resp.ID, "name": resp.Name})
}

// Delete implements ep: DELETE /classrooms/:id
func (h *ClassroomHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	err = h.service.Delete(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete classroom"})
		return
	}

	c.Status(http.StatusNoContent)
}
