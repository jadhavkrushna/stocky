package handlers

import (
	"net/http"

	"stocky-backend/internal/models"
	"stocky-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PortfolioHandler struct {
	portfolioService *services.PortfolioService
	logger           *logrus.Logger
}

func NewPortfolioHandler(portfolioService *services.PortfolioService, logger *logrus.Logger) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: portfolioService,
		logger:           logger,
	}
}

func (h *PortfolioHandler) GetTodayStocks(c *gin.Context) {
	userID, err := parseUUIDParam(c, "userId")
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"user_id":    c.Param("userId"),
			"error":      err.Error(),
		}).Error("Invalid user ID format")
		
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}
	
	response, err := h.portfolioService.GetTodayStocks(userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get today's stocks")
		
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve today's stocks",
			Message: "An internal error occurred while fetching today's stock rewards",
		})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"request_id":    c.GetString("request_id"),
		"user_id":       userID,
		"rewards_count": len(response.Rewards),
	}).Info("Today's stocks retrieved successfully")
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

func (h *PortfolioHandler) GetHistoricalINR(c *gin.Context) {
	userID, err := parseUUIDParam(c, "userId")
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"user_id":    c.Param("userId"),
			"error":      err.Error(),
		}).Error("Invalid user ID format")
		
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}
	
	response, err := h.portfolioService.GetHistoricalINR(userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get historical INR values")
		
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve historical INR values",
			Message: "An internal error occurred while fetching historical portfolio values",
		})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"request_id":   c.GetString("request_id"),
		"user_id":      userID,
		"history_days": len(response.HistoricalValues),
	}).Info("Historical INR values retrieved successfully")
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

func (h *PortfolioHandler) GetUserStats(c *gin.Context) {
	userID, err := parseUUIDParam(c, "userId")
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"user_id":    c.Param("userId"),
			"error":      err.Error(),
		}).Error("Invalid user ID format")
		
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}
	
	response, err := h.portfolioService.GetUserStats(userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get user stats")
		
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve user statistics",
			Message: "An internal error occurred while fetching user portfolio statistics",
		})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"request_id":            c.GetString("request_id"),
		"user_id":               userID,
		"portfolio_value":       response.CurrentPortfolioValueINR,
		"today_rewards_count":   len(response.TodayRewards),
	}).Info("User stats retrieved successfully")
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

func (h *PortfolioHandler) GetUserPortfolio(c *gin.Context) {
	userID, err := parseUUIDParam(c, "userId")
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"user_id":    c.Param("userId"),
			"error":      err.Error(),
		}).Error("Invalid user ID format")
		
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID format",
		})
		return
	}
	
	response, err := h.portfolioService.GetUserPortfolio(userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get user portfolio")
		
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve user portfolio",
			Message: "An internal error occurred while fetching user portfolio details",
		})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"request_id":       c.GetString("request_id"),
		"user_id":          userID,
		"holdings_count":   len(response.Holdings),
		"portfolio_value":  response.TotalPortfolioValueINR,
	}).Info("User portfolio retrieved successfully")
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// Helper function to parse UUID from URL parameter
func parseUUIDParam(c *gin.Context, paramName string) (uuid.UUID, error) {
	paramValue := c.Param(paramName)
	return uuid.Parse(paramValue)
}