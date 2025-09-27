package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"stocky-backend/internal/config"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type StockPriceService struct {
	db     *sql.DB
	config *config.Config
	logger *logrus.Logger
	client *http.Client
}

type StockPriceAPIResponse struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp string  `json:"timestamp"`
}

func NewStockPriceService(cfg *config.Config, logger *logrus.Logger) *StockPriceService {
	return &StockPriceService{
		config: cfg,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *StockPriceService) SetDB(db *sql.DB) {
	s.db = db
}

func (s *StockPriceService) GetCurrentPrice(stockID uuid.UUID) (decimal.Decimal, error) {
	// Try to get the latest price from database first
	var price decimal.Decimal
	var timestamp time.Time
	
	err := s.db.QueryRow(`
		SELECT price_inr, timestamp FROM stock_prices 
		WHERE stock_id = $1 
		ORDER BY timestamp DESC 
		LIMIT 1`, stockID).Scan(&price, &timestamp)
	
	if err != nil && err != sql.ErrNoRows {
		return decimal.Zero, fmt.Errorf("failed to get stock price: %w", err)
	}
	
	// If no price found or price is too old (more than 1 hour), fetch new price
	if err == sql.ErrNoRows || time.Since(timestamp) > time.Hour {
		s.logger.WithField("stock_id", stockID).Info("Fetching fresh stock price")
		
		// Get stock symbol for API call
		var symbol string
		err := s.db.QueryRow("SELECT symbol FROM stocks WHERE id = $1", stockID).Scan(&symbol)
		if err != nil {
			return decimal.Zero, fmt.Errorf("failed to get stock symbol: %w", err)
		}
		
		// Fetch from API
		freshPrice, err := s.fetchPriceFromAPI(symbol)
		if err != nil {
			// If API fails and we have a cached price, use it with warning
			if timestamp.IsZero() {
				return decimal.Zero, fmt.Errorf("no price available and API failed: %w", err)
			}
			
			s.logger.WithFields(logrus.Fields{
				"stock_id": stockID,
				"symbol":   symbol,
				"error":    err.Error(),
			}).Warn("Using stale price due to API failure")
			
			return price, nil
		}
		
		// Store fresh price
		_, err = s.db.Exec(`
			INSERT INTO stock_prices (stock_id, price_inr, source)
			VALUES ($1, $2, 'API')`,
			stockID, freshPrice)
		if err != nil {
			s.logger.WithError(err).Error("Failed to store fresh price")
			// Continue with the fetched price even if storage fails
		}
		
		return freshPrice, nil
	}
	
	return price, nil
}

func (s *StockPriceService) fetchPriceFromAPI(symbol string) (decimal.Decimal, error) {
	// For this example, we'll simulate a stock price API with random prices
	// In a real application, you would call an actual stock price API
	
	// Simulate different base prices for different stocks
	basePrices := map[string]float64{
		"RELIANCE":  2450.00,
		"TCS":       3850.00,
		"INFOSYS":   1750.00,
		"HDFCBANK":  1650.00,
		"ICICIBANK": 1150.00,
	}
	
	basePrice, exists := basePrices[symbol]
	if !exists {
		basePrice = 1000.00 // Default base price
	}
	
	// Add some random variation (-5% to +5%)
	variation := (rand.Float64() - 0.5) * 0.1 // -5% to +5%
	finalPrice := basePrice * (1 + variation)
	
	// Simulate occasional API failures (5% chance)
	if rand.Float64() < 0.05 {
		return decimal.Zero, fmt.Errorf("simulated API failure for %s", symbol)
	}
	
	// Add some latency simulation
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	
	price := decimal.NewFromFloat(finalPrice)
	
	s.logger.WithFields(logrus.Fields{
		"symbol":    symbol,
		"price_inr": price.String(),
		"source":    "API",
	}).Debug("Fetched stock price from API")
	
	return price, nil
}

func (s *StockPriceService) fetchPriceFromRealAPI(symbol string) (decimal.Decimal, error) {
	// This is how you would implement a real API call
	// Example with a hypothetical API
	
	url := fmt.Sprintf("%s/stock/%s/price", s.config.StockAPI.URL, symbol)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add API key if configured
	if s.config.StockAPI.Key != "" {
		req.Header.Add("Authorization", "Bearer "+s.config.StockAPI.Key)
	}
	
	resp, err := s.client.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	var apiResp StockPriceAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return decimal.Zero, fmt.Errorf("failed to decode API response: %w", err)
	}
	
	return decimal.NewFromFloat(apiResp.Price), nil
}

func (s *StockPriceService) UpdateAllStockPrices() error {
	// Get all active stocks
	rows, err := s.db.Query(`
		SELECT id, symbol FROM stocks 
		WHERE is_active = true`)
	if err != nil {
		return fmt.Errorf("failed to get active stocks: %w", err)
	}
	defer rows.Close()
	
	var stocks []struct {
		ID     uuid.UUID
		Symbol string
	}
	
	for rows.Next() {
		var stock struct {
			ID     uuid.UUID
			Symbol string
		}
		if err := rows.Scan(&stock.ID, &stock.Symbol); err != nil {
			s.logger.WithError(err).Error("Failed to scan stock")
			continue
		}
		stocks = append(stocks, stock)
	}
	
	s.logger.WithField("count", len(stocks)).Info("Starting bulk price update")
	
	// Update prices for all stocks
	successCount := 0
	for _, stock := range stocks {
		price, err := s.fetchPriceFromAPI(stock.Symbol)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"stock_id": stock.ID,
				"symbol":   stock.Symbol,
				"error":    err.Error(),
			}).Error("Failed to fetch price for stock")
			continue
		}
		
		// Store the price
		_, err = s.db.Exec(`
			INSERT INTO stock_prices (stock_id, price_inr, source)
			VALUES ($1, $2, 'API')`,
			stock.ID, price)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"stock_id": stock.ID,
				"symbol":   stock.Symbol,
				"error":    err.Error(),
			}).Error("Failed to store price for stock")
			continue
		}
		
		successCount++
	}
	
	s.logger.WithFields(logrus.Fields{
		"total":   len(stocks),
		"success": successCount,
		"failed":  len(stocks) - successCount,
	}).Info("Bulk price update completed")
	
	return nil
}

func (s *StockPriceService) GetPriceHistory(stockID uuid.UUID, days int) ([]struct {
	Price     decimal.Decimal
	Timestamp time.Time
}, error) {
	rows, err := s.db.Query(`
		SELECT price_inr, timestamp FROM stock_prices
		WHERE stock_id = $1 AND timestamp >= NOW() - INTERVAL '%d days'
		ORDER BY timestamp DESC`, stockID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}
	defer rows.Close()
	
	var history []struct {
		Price     decimal.Decimal
		Timestamp time.Time
	}
	
	for rows.Next() {
		var record struct {
			Price     decimal.Decimal
			Timestamp time.Time
		}
		if err := rows.Scan(&record.Price, &record.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan price history: %w", err)
		}
		history = append(history, record)
	}
	
	return history, nil
}

func (s *StockPriceService) GetStockPriceAtTime(stockID uuid.UUID, targetTime time.Time) (decimal.Decimal, error) {
	var price decimal.Decimal
	
	err := s.db.QueryRow(`
		SELECT price_inr FROM stock_prices
		WHERE stock_id = $1 AND timestamp <= $2
		ORDER BY timestamp DESC
		LIMIT 1`, stockID, targetTime).Scan(&price)
	
	if err == sql.ErrNoRows {
		return decimal.Zero, fmt.Errorf("no price data available for stock at requested time")
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get historical price: %w", err)
	}
	
	return price, nil
}

// BatchGetCurrentPrices gets current prices for multiple stocks efficiently
func (s *StockPriceService) BatchGetCurrentPrices(stockIDs []uuid.UUID) (map[uuid.UUID]decimal.Decimal, error) {
	if len(stockIDs) == 0 {
		return make(map[uuid.UUID]decimal.Decimal), nil
	}
	
	// Create placeholders for the IN clause
	placeholders := make([]interface{}, len(stockIDs))
	for i, id := range stockIDs {
		placeholders[i] = id
	}
	
	query := `
		WITH latest_prices AS (
			SELECT DISTINCT ON (stock_id) stock_id, price_inr, timestamp
			FROM stock_prices
			WHERE stock_id = ANY($1)
			ORDER BY stock_id, timestamp DESC
		)
		SELECT stock_id, price_inr FROM latest_prices`
	
	rows, err := s.db.Query(query, stockIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get stock prices: %w", err)
	}
	defer rows.Close()
	
	prices := make(map[uuid.UUID]decimal.Decimal)
	for rows.Next() {
		var stockID uuid.UUID
		var price decimal.Decimal
		if err := rows.Scan(&stockID, &price); err != nil {
			return nil, fmt.Errorf("failed to scan stock price: %w", err)
		}
		prices[stockID] = price
	}
	
	return prices, nil
}