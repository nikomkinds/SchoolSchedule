package utils

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// ExtractRefreshTokenFromHeader extracts: Authorization: Bearer xxx
func ExtractRefreshTokenFromHeader(c *gin.Context) (string, error) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid Authorization header format")
	}

	return parts[1], nil
}
