package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikomkinds/SchoolSchedule/internal/services"
)

type SubjectHandler struct {
	service services.SubjectService
}

func NewSubjectHandler(service services.SubjectService) *SubjectHandler {
	return &SubjectHandler{service: service}
}

type createSubjectRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *SubjectHandler) GetAll(c *gin.Context) {
	subjects, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load subjects"})
		return
	}

	c.JSON(http.StatusOK, subjects)
}

func (h *SubjectHandler) Create(c *gin.Context) {
	var req createSubjectRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	subject, err := h.service.Create(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subject"})
		return
	}

	c.JSON(http.StatusCreated, subject)
}

func (h *SubjectHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subject id"})
		return
	}

	err = h.service.Delete(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subject not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
