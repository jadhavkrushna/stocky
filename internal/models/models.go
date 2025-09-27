package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Stock struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Symbol    string    `json:"symbol" db:"symbol"`
	Name      string    `json:"name" db:"name"`
	Exchange  string    `json:"exchange" db:"exchange"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Reward struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	StockID         uuid.UUID       `json:"stock_id" db:"stock_id"`
	Quantity        decimal.Decimal `json:"quantity" db:"quantity"`
	MarketPriceINR  decimal.Decimal `json:"market_price_inr" db:"market_price_inr"`
	TotalValueINR   decimal.Decimal `json:"total_value_inr" db:"total_value_inr"`
	RewardType      string          `json:"reward_type" db:"reward_type"`
	IdempotencyKey  string          `json:"idempotency_key" db:"idempotency_key"`
	Metadata        interface{}     `json:"metadata" db:"metadata"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	
	// Joined fields
	StockSymbol string `json:"stock_symbol,omitempty" db:"stock_symbol"`
	UserEmail   string `json:"user_email,omitempty" db:"user_email"`
}

type LedgerEntry struct {
	ID                uuid.UUID        `json:"id" db:"id"`
	RewardID          *uuid.UUID       `json:"reward_id" db:"reward_id"`
	AdjustmentID      *uuid.UUID       `json:"adjustment_id" db:"adjustment_id"`
	EntryType         string           `json:"entry_type" db:"entry_type"` // DEBIT, CREDIT
	AccountType       string           `json:"account_type" db:"account_type"`
	AccountIdentifier string           `json:"account_identifier" db:"account_identifier"`
	AmountINR         decimal.Decimal  `json:"amount_inr" db:"amount_inr"`
	Quantity          *decimal.Decimal `json:"quantity" db:"quantity"`
	StockID           *uuid.UUID       `json:"stock_id" db:"stock_id"`
	Description       string           `json:"description" db:"description"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
}

type StockPrice struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	StockID   uuid.UUID       `json:"stock_id" db:"stock_id"`
	PriceINR  decimal.Decimal `json:"price_inr" db:"price_inr"`
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
	Source    string          `json:"source" db:"source"`
	
	// Joined fields
	StockSymbol string `json:"stock_symbol,omitempty" db:"stock_symbol"`
}

type PortfolioSnapshot struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	UserID       uuid.UUID       `json:"user_id" db:"user_id"`
	StockID      uuid.UUID       `json:"stock_id" db:"stock_id"`
	Quantity     decimal.Decimal `json:"quantity" db:"quantity"`
	PriceINR     decimal.Decimal `json:"price_inr" db:"price_inr"`
	ValueINR     decimal.Decimal `json:"value_inr" db:"value_inr"`
	SnapshotDate time.Time       `json:"snapshot_date" db:"snapshot_date"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	
	// Joined fields
	StockSymbol string `json:"stock_symbol,omitempty" db:"stock_symbol"`
}

type CorporateAction struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	StockID       uuid.UUID        `json:"stock_id" db:"stock_id"`
	ActionType    string           `json:"action_type" db:"action_type"` // SPLIT, MERGER, DELISTING
	Ratio         *decimal.Decimal `json:"ratio" db:"ratio"`
	EffectiveDate time.Time        `json:"effective_date" db:"effective_date"`
	Description   string           `json:"description" db:"description"`
	Processed     bool             `json:"processed" db:"processed"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
}

type Adjustment struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	OriginalRewardID   uuid.UUID       `json:"original_reward_id" db:"original_reward_id"`
	AdjustmentType     string          `json:"adjustment_type" db:"adjustment_type"` // REFUND, CORRECTION
	QuantityAdjustment decimal.Decimal `json:"quantity_adjustment" db:"quantity_adjustment"`
	Reason             string          `json:"reason" db:"reason"`
	ApprovedBy         *string         `json:"approved_by" db:"approved_by"`
	Processed          bool            `json:"processed" db:"processed"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
}

// Request/Response DTOs
type CreateRewardRequest struct {
	UserID      uuid.UUID   `json:"user_id" binding:"required"`
	StockSymbol string      `json:"stock_symbol" binding:"required"`
	Quantity    string      `json:"quantity" binding:"required"`
	RewardType  string      `json:"reward_type" binding:"required"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

type CreateRewardResponse struct {
	RewardID       uuid.UUID `json:"reward_id"`
	UserID         uuid.UUID `json:"user_id"`
	StockSymbol    string    `json:"stock_symbol"`
	Quantity       string    `json:"quantity"`
	MarketPriceINR string    `json:"market_price_inr"`
	TotalValueINR  string    `json:"total_value_inr"`
	Timestamp      time.Time `json:"timestamp"`
}

type TodayStocksResponse struct {
	UserID        uuid.UUID            `json:"user_id"`
	Date          string               `json:"date"`
	Rewards       []RewardSummary      `json:"rewards"`
	TotalValueINR string               `json:"total_value_inr"`
}

type RewardSummary struct {
	RewardID       uuid.UUID `json:"reward_id"`
	StockSymbol    string    `json:"stock_symbol"`
	Quantity       string    `json:"quantity"`
	MarketPriceINR string    `json:"market_price_inr"`
	TotalValueINR  string    `json:"total_value_inr"`
	RewardType     string    `json:"reward_type"`
	Timestamp      time.Time `json:"timestamp"`
}

type HistoricalINRResponse struct {
	UserID           uuid.UUID              `json:"user_id"`
	HistoricalValues []HistoricalDayValue   `json:"historical_values"`
}

type HistoricalDayValue struct {
	Date          string                    `json:"date"`
	TotalValueINR string                    `json:"total_value_inr"`
	Stocks        []StockHolding           `json:"stocks"`
}

type StockHolding struct {
	StockSymbol string `json:"stock_symbol"`
	Quantity    string `json:"quantity"`
	PriceINR    string `json:"price_inr"`
	ValueINR    string `json:"value_inr"`
}

type UserStatsResponse struct {
	UserID                    uuid.UUID                `json:"user_id"`
	TodayRewards              map[string]string        `json:"today_rewards"`
	CurrentPortfolioValueINR  string                   `json:"current_portfolio_value_inr"`
	TotalStocksRewarded       string                   `json:"total_stocks_rewarded"`
	PortfolioPerformance      PortfolioPerformance     `json:"portfolio_performance"`
}

type PortfolioPerformance struct {
	TodayChangeINR     string `json:"today_change_inr"`
	TodayChangePercent string `json:"today_change_percent"`
}

type UserPortfolioResponse struct {
	UserID                  uuid.UUID         `json:"user_id"`
	Holdings                []PortfolioHolding `json:"holdings"`
	TotalPortfolioValueINR  string            `json:"total_portfolio_value_inr"`
	TotalCostINR            string            `json:"total_cost_inr"`
	TotalUnrealizedPnLINR   string            `json:"total_unrealized_pnl_inr"`
	LastUpdated             time.Time         `json:"last_updated"`
}

type PortfolioHolding struct {
	StockSymbol           string `json:"stock_symbol"`
	TotalQuantity         string `json:"total_quantity"`
	CurrentPriceINR       string `json:"current_price_inr"`
	CurrentValueINR       string `json:"current_value_inr"`
	AverageCostINR        string `json:"average_cost_inr"`
	UnrealizedPnLINR      string `json:"unrealized_pnl_inr"`
	UnrealizedPnLPercent  string `json:"unrealized_pnl_percent"`
}

type AdjustRewardRequest struct {
	OriginalRewardID   uuid.UUID `json:"original_reward_id" binding:"required"`
	AdjustmentType     string    `json:"adjustment_type" binding:"required"`
	QuantityAdjustment string    `json:"quantity_adjustment" binding:"required"`
	Reason             string    `json:"reason" binding:"required"`
	ApprovedBy         string    `json:"approved_by" binding:"required"`
}

type SetStockPriceRequest struct {
	StockSymbol string `json:"stock_symbol" binding:"required"`
	PriceINR    string `json:"price_inr" binding:"required"`
}

// API Response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}