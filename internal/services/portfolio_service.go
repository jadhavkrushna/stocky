package services

import (
	"database/sql"
	"fmt"
	"time"

	"stocky-backend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type PortfolioService struct {
	db               *sql.DB
	stockPriceService *StockPriceService
	logger           *logrus.Logger
}

func NewPortfolioService(db *sql.DB, stockPriceService *StockPriceService, logger *logrus.Logger) *PortfolioService {
	// Set the DB reference in stock price service if not already set
	if stockPriceService != nil {
		stockPriceService.SetDB(db)
	}
	
	return &PortfolioService{
		db:               db,
		stockPriceService: stockPriceService,
		logger:           logger,
	}
}

func (s *PortfolioService) GetTodayStocks(userID uuid.UUID) (*models.TodayStocksResponse, error) {
	today := time.Now().Format("2006-01-02")
	
	rows, err := s.db.Query(`
		SELECT r.id, r.stock_id, r.quantity, r.market_price_inr, r.total_value_inr, 
		       r.reward_type, r.created_at, st.symbol
		FROM rewards r
		JOIN stocks st ON r.stock_id = st.id
		WHERE r.user_id = $1 AND DATE(r.created_at) = $2
		ORDER BY r.created_at DESC`, userID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's rewards: %w", err)
	}
	defer rows.Close()
	
	var rewards []models.RewardSummary
	totalValue := decimal.Zero
	
	for rows.Next() {
		var reward models.RewardSummary
		var stockID uuid.UUID
		var quantity, marketPrice, rewardValue decimal.Decimal
		
		if err := rows.Scan(&reward.RewardID, &stockID, &quantity, &marketPrice, 
			&rewardValue, &reward.RewardType, &reward.Timestamp, &reward.StockSymbol); err != nil {
			return nil, fmt.Errorf("failed to scan reward: %w", err)
		}
		
		reward.Quantity = quantity.String()
		reward.MarketPriceINR = marketPrice.String()
		reward.TotalValueINR = rewardValue.String()
		
		rewards = append(rewards, reward)
		totalValue = totalValue.Add(rewardValue)
	}
	
	return &models.TodayStocksResponse{
		UserID:        userID,
		Date:          today,
		Rewards:       rewards,
		TotalValueINR: totalValue.String(),
	}, nil
}

func (s *PortfolioService) GetHistoricalINR(userID uuid.UUID) (*models.HistoricalINRResponse, error) {
	// Get portfolio snapshots for the last 30 days (excluding today)
	rows, err := s.db.Query(`
		SELECT ps.snapshot_date, ps.stock_id, ps.quantity, ps.price_inr, ps.value_inr, st.symbol
		FROM portfolio_snapshots ps
		JOIN stocks st ON ps.stock_id = st.id
		WHERE ps.user_id = $1 AND ps.snapshot_date < CURRENT_DATE
		ORDER BY ps.snapshot_date DESC, st.symbol`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical portfolio data: %w", err)
	}
	defer rows.Close()
	
	// Group by date
	dateMap := make(map[string][]models.StockHolding)
	dateTotals := make(map[string]decimal.Decimal)
	
	for rows.Next() {
		var snapshotDate time.Time
		var stockID uuid.UUID
		var quantity, price, value decimal.Decimal
		var symbol string
		
		if err := rows.Scan(&snapshotDate, &stockID, &quantity, &price, &value, &symbol); err != nil {
			return nil, fmt.Errorf("failed to scan portfolio snapshot: %w", err)
		}
		
		dateStr := snapshotDate.Format("2006-01-02")
		
		holding := models.StockHolding{
			StockSymbol: symbol,
			Quantity:    quantity.String(),
			PriceINR:    price.String(),
			ValueINR:    value.String(),
		}
		
		dateMap[dateStr] = append(dateMap[dateStr], holding)
		dateTotals[dateStr] = dateTotals[dateStr].Add(value)
	}
	
	// Convert to response format
	var historicalValues []models.HistoricalDayValue
	for dateStr, holdings := range dateMap {
		historicalValues = append(historicalValues, models.HistoricalDayValue{
			Date:          dateStr,
			TotalValueINR: dateTotals[dateStr].String(),
			Stocks:        holdings,
		})
	}
	
	return &models.HistoricalINRResponse{
		UserID:           userID,
		HistoricalValues: historicalValues,
	}, nil
}

func (s *PortfolioService) GetUserStats(userID uuid.UUID) (*models.UserStatsResponse, error) {
	// Get today's rewards by stock symbol
	todayRewards := make(map[string]string)
	rows, err := s.db.Query(`
		SELECT st.symbol, SUM(r.quantity)
		FROM rewards r
		JOIN stocks st ON r.stock_id = st.id
		WHERE r.user_id = $1 AND DATE(r.created_at) = CURRENT_DATE
		GROUP BY st.symbol`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's rewards: %w", err)
	}
	defer rows.Close()
	
	totalTodayQuantity := decimal.Zero
	for rows.Next() {
		var symbol string
		var quantity decimal.Decimal
		if err := rows.Scan(&symbol, &quantity); err != nil {
			return nil, fmt.Errorf("failed to scan today's rewards: %w", err)
		}
		todayRewards[symbol] = quantity.String()
		totalTodayQuantity = totalTodayQuantity.Add(quantity)
	}
	
	// Get current portfolio value
	portfolio, err := s.GetUserPortfolio(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user portfolio: %w", err)
	}
	
	// Calculate today's performance (simplified - would need yesterday's snapshot in real app)
	performance := models.PortfolioPerformance{
		TodayChangeINR:     "+0.00", // Would calculate from yesterday's snapshot
		TodayChangePercent: "+0.00%",
	}
	
	return &models.UserStatsResponse{
		UserID:                   userID,
		TodayRewards:             todayRewards,
		CurrentPortfolioValueINR: portfolio.TotalPortfolioValueINR,
		TotalStocksRewarded:      totalTodayQuantity.String(),
		PortfolioPerformance:     performance,
	}, nil
}

func (s *PortfolioService) GetUserPortfolio(userID uuid.UUID) (*models.UserPortfolioResponse, error) {
	// Get current holdings (sum of all rewards minus any adjustments)
	rows, err := s.db.Query(`
		WITH user_holdings AS (
			SELECT 
				r.stock_id,
				st.symbol,
				SUM(r.quantity) as total_quantity,
				SUM(r.total_value_inr) / SUM(r.quantity) as avg_cost
			FROM rewards r
			JOIN stocks st ON r.stock_id = st.id
			WHERE r.user_id = $1
			GROUP BY r.stock_id, st.symbol
			HAVING SUM(r.quantity) > 0
		)
		SELECT stock_id, symbol, total_quantity, avg_cost
		FROM user_holdings
		ORDER BY symbol`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user holdings: %w", err)
	}
	defer rows.Close()
	
	var holdings []models.PortfolioHolding
	var stockIDs []uuid.UUID
	holdingsMap := make(map[uuid.UUID]*models.PortfolioHolding)
	
	for rows.Next() {
		var stockID uuid.UUID
		var symbol string
		var totalQuantity, avgCost decimal.Decimal
		
		if err := rows.Scan(&stockID, &symbol, &totalQuantity, &avgCost); err != nil {
			return nil, fmt.Errorf("failed to scan holding: %w", err)
		}
		
		holding := &models.PortfolioHolding{
			StockSymbol:     symbol,
			TotalQuantity:   totalQuantity.String(),
			AverageCostINR:  avgCost.String(),
		}
		
		holdings = append(holdings, *holding)
		stockIDs = append(stockIDs, stockID)
		holdingsMap[stockID] = holding
	}
	
	// Get current prices for all stocks
	currentPrices, err := s.stockPriceService.BatchGetCurrentPrices(stockIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get current prices: %w", err)
	}
	
	// Calculate current values and P&L
	totalPortfolioValue := decimal.Zero
	totalCost := decimal.Zero
	
	for i := range holdings {
		stockID := stockIDs[i]
		holding := &holdings[i]
		
		quantity, _ := decimal.NewFromString(holding.TotalQuantity)
		avgCost, _ := decimal.NewFromString(holding.AverageCostINR)
		
		currentPrice := currentPrices[stockID]
		currentValue := quantity.Mul(currentPrice)
		cost := quantity.Mul(avgCost)
		unrealizedPnL := currentValue.Sub(cost)
		unrealizedPnLPercent := decimal.Zero
		if !cost.IsZero() {
			unrealizedPnLPercent = unrealizedPnL.Div(cost).Mul(decimal.NewFromInt(100))
		}
		
		holding.CurrentPriceINR = currentPrice.String()
		holding.CurrentValueINR = currentValue.String()
		holding.UnrealizedPnLINR = formatPnL(unrealizedPnL)
		holding.UnrealizedPnLPercent = formatPnLPercent(unrealizedPnLPercent)
		
		totalPortfolioValue = totalPortfolioValue.Add(currentValue)
		totalCost = totalCost.Add(cost)
	}
	
	totalUnrealizedPnL := totalPortfolioValue.Sub(totalCost)
	
	return &models.UserPortfolioResponse{
		UserID:                 userID,
		Holdings:               holdings,
		TotalPortfolioValueINR: totalPortfolioValue.String(),
		TotalCostINR:           totalCost.String(),
		TotalUnrealizedPnLINR:  formatPnL(totalUnrealizedPnL),
		LastUpdated:            time.Now(),
	}, nil
}

func (s *PortfolioService) CreateDailySnapshots() error {
	s.logger.Info("Creating daily portfolio snapshots")
	
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	
	// Get all users with holdings
	rows, err := s.db.Query(`
		SELECT DISTINCT user_id FROM rewards`)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()
	
	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			s.logger.WithError(err).Error("Failed to scan user ID")
			continue
		}
		userIDs = append(userIDs, userID)
	}
	
	successCount := 0
	for _, userID := range userIDs {
		if err := s.createUserDailySnapshot(userID, yesterday); err != nil {
			s.logger.WithFields(logrus.Fields{
				"user_id": userID,
				"date":    yesterday,
				"error":   err.Error(),
			}).Error("Failed to create user snapshot")
			continue
		}
		successCount++
	}
	
	s.logger.WithFields(logrus.Fields{
		"date":    yesterday,
		"total":   len(userIDs),
		"success": successCount,
		"failed":  len(userIDs) - successCount,
	}).Info("Daily snapshot creation completed")
	
	return nil
}

func (s *PortfolioService) createUserDailySnapshot(userID uuid.UUID, date string) error {
	// Get user's holdings as of end of the given date
	rows, err := s.db.Query(`
		SELECT 
			r.stock_id,
			SUM(r.quantity) as total_quantity
		FROM rewards r
		WHERE r.user_id = $1 AND DATE(r.created_at) <= $2
		GROUP BY r.stock_id
		HAVING SUM(r.quantity) > 0`, userID, date)
	if err != nil {
		return fmt.Errorf("failed to get user holdings: %w", err)
	}
	defer rows.Close()
	
	var holdings []struct {
		StockID  uuid.UUID
		Quantity decimal.Decimal
	}
	
	for rows.Next() {
		var holding struct {
			StockID  uuid.UUID
			Quantity decimal.Decimal
		}
		if err := rows.Scan(&holding.StockID, &holding.Quantity); err != nil {
			return fmt.Errorf("failed to scan holding: %w", err)
		}
		holdings = append(holdings, holding)
	}
	
	// Get stock prices for end of day
	endOfDay, _ := time.Parse("2006-01-02", date)
	endOfDay = endOfDay.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	
	for _, holding := range holdings {
		price, err := s.stockPriceService.GetStockPriceAtTime(holding.StockID, endOfDay)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"user_id":  userID,
				"stock_id": holding.StockID,
				"date":     date,
				"error":    err.Error(),
			}).Warn("Failed to get stock price for snapshot")
			continue
		}
		
		value := holding.Quantity.Mul(price)
		
		// Insert snapshot
		_, err = s.db.Exec(`
			INSERT INTO portfolio_snapshots (user_id, stock_id, quantity, price_inr, value_inr, snapshot_date)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, stock_id, snapshot_date) DO UPDATE SET
			quantity = EXCLUDED.quantity,
			price_inr = EXCLUDED.price_inr,
			value_inr = EXCLUDED.value_inr`,
			userID, holding.StockID, holding.Quantity, price, value, date)
		if err != nil {
			return fmt.Errorf("failed to insert snapshot: %w", err)
		}
	}
	
	return nil
}

func formatPnL(pnl decimal.Decimal) string {
	if pnl.IsPositive() {
		return "+" + pnl.String()
	}
	return pnl.String()
}

func formatPnLPercent(pnlPercent decimal.Decimal) string {
	if pnlPercent.IsPositive() {
		return "+" + pnlPercent.StringFixed(2) + "%"
	}
	return pnlPercent.StringFixed(2) + "%"
}