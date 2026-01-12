package middleware

import (
	"context"
	"time"

	"genie-audit-backend/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DefaultTimeout applies to inbound requests to avoid hangs.
const DefaultTimeout = 10 * time.Second

// RequestContext injects a trace ID into the request context and response headers.
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := uuid.NewString()
		ctx := context.WithValue(c.Request.Context(), logger.TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Set(string(logger.TraceIDKey), traceID)
		c.Writer.Header().Set("X-Trace-ID", traceID)
		c.Next()
	}
}

// Timeout protects handlers from hanging indefinitely.
func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
