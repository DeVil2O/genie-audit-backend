package middleware

import (
	"time"

	"genie-audit-backend/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestLogger emits structured logs for each HTTP request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		traceID, _ := c.Get(string(logger.TraceIDKey))
		entry := logger.Log.WithFields(logrus.Fields{
			"method":  c.Request.Method,
			"path":    c.FullPath(),
			"status":  c.Writer.Status(),
			"latency": latency.String(),
			"traceID": traceID,
		})

		if len(c.Errors) > 0 {
			entry.WithField("errors", c.Errors.String()).Warn("request completed with errors")
			return
		}
		entry.Info("request completed")
	}
}
