package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/services"
)

type ClassHandler struct {
	service services.ClassService
}

func NewClassHandler(service services.ClassService) *ClassHandler {
	return &ClassHandler{service: service}
}

// GetAll implements ep: GET /classes
func (h *ClassHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.service.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load classes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Create implements ep: POST /classes
func (h *ClassHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := c.Request.Context()
	newClass, err := h.service.Create(ctx, req.Name)
	if err != nil {
		if errors.Is(err, strconv.ErrSyntax) || errors.Is(err, strconv.ErrRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class grade"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create class"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": newClass})
}

// Delete implements ep: DELETE /classes/:id
func (h *ClassHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.Delete(ctx, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// BulkUpdate implements ep: PUT /classes/bulk
func (h *ClassHandler) BulkUpdate(c *gin.Context) {
	var req models.BulkUpdateClassesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := c.Request.Context()
	updated, err := h.service.BulkUpdate(ctx, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update classes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Classes successfully updated",
		"updated": updated,
	})
}
