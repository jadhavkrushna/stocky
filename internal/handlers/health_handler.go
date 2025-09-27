package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"stocky-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HealthHandler struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewHealthHandler(db *sql.DB, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]HealthCheck `json:"checks"`
}

type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

var startTime = time.Now()

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	start := time.Now()
	
	health := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // This could come from build info
		Uptime:    time.Since(startTime).String(),
		Checks:    make(map[string]HealthCheck),
	}
	
	// Database health check
	dbStart := time.Now()
	if err := h.checkDatabase(); err != nil {
		health.Status = "unhealthy"
		health.Checks["database"] = HealthCheck{
			Status:  "fail",
			Message: err.Error(),
			Latency: time.Since(dbStart).String(),
		}
	} else {
		health.Checks["database"] = HealthCheck{
			Status:  "pass",
			Latency: time.Since(dbStart).String(),
		}
	}
	
	// System memory check (basic)
	health.Checks["memory"] = HealthCheck{
		Status: "pass", // In production, you'd check actual memory usage
	}
	
	// Overall health status
	if health.Status == "healthy" {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"latency":    time.Since(start).String(),
		}).Debug("Health check completed")
		
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data:    health,
		})
	} else {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"status":     health.Status,
			"latency":    time.Since(start).String(),
		}).Warn("Health check failed")
		
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Data:    health,
			Error:   "Service is unhealthy",
		})
	}
}

func (h *HealthHandler) checkDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Simple ping to check database connectivity
	if err := h.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	// Check if we can perform a simple query
	var count int
	if err := h.db.QueryRowContext(ctx, "SELECT 1").Scan(&count); err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}
	
	return nil
}

// ReadinessCheck is more strict than health check - ensures service is ready to serve traffic
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// This would check if all dependencies are ready
	// For now, just check database
	if err := h.checkDatabase(); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "Service not ready",
			Message: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Service is ready",
	})
}

// LivenessCheck is a simple check to see if the service is running
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Service is alive",
		Data: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime":    time.Since(startTime).String(),
		},
	})
}