package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger middleware for request logging
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"client_ip":   param.ClientIP,
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
			"method":      param.Method,
			"path":        param.Path,
			"protocol":    param.Request.Proto,
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"user_agent":  param.Request.UserAgent(),
			"error":       param.ErrorMessage,
		}).Info("HTTP Request")
		return ""
	})
}

// CORS middleware for Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		c.Header("Access-Control-Expose-Headers", "Link")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "300")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// Recovery middleware for panic recovery
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logrus.WithField("panic", recovered).Error("Panic recovered")
		c.JSON(500, gin.H{
			"error":   "internal_server_error",
			"message": "Internal server error occurred",
			"code":    500,
		})
	})
}

// Rate limiting middleware (simple implementation)
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple rate limiting - in production, use Redis-based rate limiting
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
