package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const apiKeyHeader = "X-API-Key"

// Auth is a lightweight API-key guard for MVP.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(apiKeyHeader)
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "missing API key",
			})
			return
		}
		// In production, validate apiKey and set project scope in context.
		c.Set("apiKey", apiKey)
		c.Next()
	}
}
