package handlers

import (
	"github.com/nikomkinds/SchoolSchedule/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikomkinds/SchoolSchedule/internal/models"
	"github.com/nikomkinds/SchoolSchedule/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login implements ep: POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ctx := c.Request.Context()

	resp, err := h.authService.Login(ctx, &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// === Set httpOnly cookies ===
	c.SetCookie(
		"access-token",
		resp.AccessToken,
		60*10, // 10 minutes
		"/",
		"",
		false,
		true,
	)

	c.SetCookie(
		"refresh-token",
		resp.RefreshToken,
		60*60*24*7, // 7 days
		"/",
		"",
		false,
		true,
	)

	// Return JSON (frontend also stores tokens in cookies)
	c.JSON(http.StatusOK, resp)
}

// Refresh implements ep: POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := utils.ExtractRefreshTokenFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing refresh token"})
		return
	}

	ctx := c.Request.Context()

	resp, err := h.authService.Refresh(ctx, refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// === Update cookies ===
	c.SetCookie(
		"access-token",
		resp.AccessToken,
		60*10,
		"/",
		"",
		false,
		true,
	)

	c.SetCookie(
		"refresh-token",
		resp.RefreshToken,
		60*60*24*7,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, resp)
}
