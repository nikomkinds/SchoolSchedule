package utils

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

// AuthMiddlewareWithTeacher extends AuthMiddleware to also fetch teacherID
func AuthMiddlewareWithTeacher(secret string, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Validate access token
		accessToken, err := c.Cookie("access-token")
		if err != nil || accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
			return
		}

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// 2) Save user info
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// 3) Teacher restriction: user MUST be a teacher
		var teacherID uuid.UUID
		err = db.QueryRow(`SELECT id FROM teachers WHERE user_id = $1`, claims.UserID).Scan(&teacherID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access restricted to teachers only"})
			return
		}

		c.Set("teacherID", teacherID)
		c.Next()
	}
}
