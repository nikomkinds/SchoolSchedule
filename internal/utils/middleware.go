package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates access-token cookie and extracts user claims
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Read cookie
		accessToken, err := c.Cookie("access-token")
		if err != nil || accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
			c.Abort()
			return
		}

		// Parse JWT
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Save user info into context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
