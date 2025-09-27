package handlers

import (
	"net/http"
	"strings"

	"stocky-backend/internal/models"
	"stocky-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type RewardHandler struct {
	rewardService *services.RewardService
	logger        *logrus.Logger
}

func NewRewardHandler(rewardService *services.RewardService, logger *logrus.Logger) *RewardHandler {
	return &RewardHandler{
		rewardService: rewardService,
		logger:        logger,
	}
}

func (h *RewardHandler) CreateReward(c *gin.Context) {
	var req models.CreateRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"error":      err.Error(),
		}).Error("Invalid request payload")
		
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}
	
	// Validate stock symbol format
	req.StockSymbol = strings.ToUpper(strings.TrimSpace(req.StockSymbol))
	if req.StockSymbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Stock symbol is required",
		})
		return
	}
	
	// Validate reward type
	validRewardTypes := map[string]bool{
		"onboarding":      true,
		"referral":        true,
		"trading":         true,
		"milestone":       true,
		"bonus":           true,
	}
	
	if !validRewardTypes[req.RewardType] {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid reward type",
		})
		return
	}
	
	// Create reward
	response, err := h.rewardService.CreateReward(&req)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"user_id":    req.UserID,
			"error":      err.Error(),
		}).Error("Failed to create reward")
		
		if strings.Contains(err.Error(), "invalid quantity") ||
			strings.Contains(err.Error(), "quantity must be positive") {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to create reward",
			Message: "An internal error occurred while processing the reward",
		})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"request_id": c.GetString("request_id"),
		"reward_id":  response.RewardID,
		"user_id":    response.UserID,
	}).Info("Reward created successfully")
	
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

func (h *RewardHandler) AdjustReward(c *gin.Context) {
	var req models.AdjustRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"error":      err.Error(),
		}).Error("Invalid adjustment request payload")
		
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}
	
	// Validate adjustment type
	validAdjustmentTypes := map[string]bool{
		"REFUND":     true,
		"CORRECTION": true,
	}
	
	if !validAdjustmentTypes[req.AdjustmentType] {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid adjustment type",
		})
		return
	}
	
	// Create adjustment
	err := h.rewardService.AdjustReward(&req)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id":         c.GetString("request_id"),
			"original_reward_id": req.OriginalRewardID,
			"error":              err.Error(),
		}).Error("Failed to create reward adjustment")
		
		if strings.Contains(err.Error(), "invalid quantity") ||
			strings.Contains(err.Error(), "failed to get original reward") {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to create adjustment",
			Message: "An internal error occurred while processing the adjustment",
		})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"request_id":         c.GetString("request_id"),
		"original_reward_id": req.OriginalRewardID,
		"adjustment_type":    req.AdjustmentType,
	}).Info("Reward adjustment created successfully")
	
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Reward adjustment created successfully",
	})
}

func (h *RewardHandler) SetStockPrice(c *gin.Context) {
	var req models.SetStockPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": c.GetString("request_id"),
			"error":      err.Error(),
		}).Error("Invalid set price request payload")
		
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request payload",
			Message: err.Error(),
		})
		return
	}
	
	// Validate stock symbol format
	req.StockSymbol = strings.ToUpper(strings.TrimSpace(req.StockSymbol))
	if req.StockSymbol == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Stock symbol is required",
		})
		return
	}
	
	// Set stock price
	err := h.rewardService.SetStockPrice(&req)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id":   c.GetString("request_id"),
			"stock_symbol": req.StockSymbol,
			"error":        err.Error(),
		}).Error("Failed to set stock price")
		
		if strings.Contains(err.Error(), "invalid price") ||
			strings.Contains(err.Error(), "price must be positive") ||
			strings.Contains(err.Error(), "stock not found") {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to set stock price",
			Message: "An internal error occurred while setting the stock price",
		})
		return
	}
	
	h.logger.WithFields(logrus.Fields{
		"request_id":   c.GetString("request_id"),
		"stock_symbol": req.StockSymbol,
		"price_inr":    req.PriceINR,
	}).Info("Stock price set successfully")
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Stock price set successfully",
	})
}

// Helper function to parse UUID from URL parameter
func parseUUIDParam(c *gin.Context, paramName string) (uuid.UUID, error) {
	paramValue := c.Param(paramName)
	return uuid.Parse(paramValue)
}