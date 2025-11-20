package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a placeholder for authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// -----------------------------
		// Example: Check for Authorization header
		// -----------------------------
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing Authorization header"})
			c.Abort()
			return
		}

		// TODO: Validate token or session here
		// Example:
		// token := strings.TrimPrefix(authHeader, "Bearer ")
		// validateToken(token)

		// Continue to the handler
		c.Next()
	}
}
