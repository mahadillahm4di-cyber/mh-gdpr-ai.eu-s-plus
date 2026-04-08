package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/auth"
)

// AuthRequired validates the JWT token from the Authorization header.
// SECURITY: In production (API_SECRET_KEY set), rejects requests without valid tokens.
// In local/dev mode (no API_SECRET_KEY), allows anonymous access with a default user ID.
func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		// Local/dev mode: no secret key configured → allow anonymous access
		if header == "" && secret == "dev-secret-key-not-for-production-use" {
			c.Set("user_id", "local-user")
			c.Next()
			return
		}

		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization format",
			})
			return
		}

		claims, err := auth.ValidateToken(parts[1], secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		// Set user ID in context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
