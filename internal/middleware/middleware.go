package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// Process request
		c.Next()
		
		// Calculate latency
		latency := time.Since(start)
		
		// Get request ID
		requestID := c.GetString("request_id")
		
		// Build log entry
		entry := logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"status":     c.Writer.Status(),
			"latency":    latency,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})
		
		// Add error if exists
		if len(c.Errors) > 0 {
			entry = entry.WithField("errors", c.Errors.String())
		}
		
		// Log based on status code
		if c.Writer.Status() >= 500 {
			entry.Error("Request completed with server error")
		} else if c.Writer.Status() >= 400 {
			entry.Warn("Request completed with client error")
		} else {
			entry.Info("Request completed")
		}
	}
}

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real application, you would validate JWT token or API key
		// For this example, we'll use a simple API key
		apiKey := c.GetHeader("X-Admin-API-Key")
		
		// This should be loaded from environment variables in production
		expectedKey := "admin-secret-key-123"
		
		if apiKey != expectedKey {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Unauthorized: Invalid admin credentials",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func RateLimiting() gin.HandlerFunc {
	// In a production environment, you would implement proper rate limiting
	// using Redis or similar. For this example, we'll just pass through.
	return func(c *gin.Context) {
		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Log the error
			if logger, exists := c.Get("logger"); exists {
				if log, ok := logger.(*logrus.Logger); ok {
					log.WithFields(logrus.Fields{
						"request_id": c.GetString("request_id"),
						"error":      err.Error(),
					}).Error("Request processing error")
				}
			}
			
			// Return error response if not already handled
			if !c.Writer.Written() {
				c.JSON(500, gin.H{
					"success": false,
					"error":   "Internal server error",
					"message": "An unexpected error occurred",
				})
			}
		}
	}
}