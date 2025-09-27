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

type RewardService struct {
	db               *sql.DB
	stockPriceService *StockPriceService
	logger           *logrus.Logger
}

func NewRewardService(db *sql.DB, stockPriceService *StockPriceService, logger *logrus.Logger) *RewardService {
	return &RewardService{
		db:               db,
		stockPriceService: stockPriceService,
		logger:           logger,
	}
}

func (s *RewardService) CreateReward(req *models.CreateRewardRequest) (*models.CreateRewardResponse, error) {
	// Parse quantity
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	// Validate quantity is positive
	if quantity.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("quantity must be positive")
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Ensure user exists (create if not exists)
	var userExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&userExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	// Get or create stock
	stock, err := s.getOrCreateStock(tx, req.StockSymbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	// Get current stock price
	currentPrice, err := s.stockPriceService.GetCurrentPrice(stock.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock price: %w", err)
	}

	// Calculate total value
	totalValue := quantity.Mul(currentPrice)

	// Generate idempotency key if not provided
	idempotencyKey := fmt.Sprintf("%s-%s-%s-%d", 
		req.UserID.String(), req.StockSymbol, req.Quantity, time.Now().Unix())

	// Create reward record
	rewardID := uuid.New()
	_, err = tx.Exec(`
		INSERT INTO rewards (id, user_id, stock_id, quantity, market_price_inr, total_value_inr, reward_type, idempotency_key, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		rewardID, req.UserID, stock.ID, quantity, currentPrice, totalValue, req.RewardType, idempotencyKey, req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create reward: %w", err)
	}

	// Create ledger entries for double-entry bookkeeping
	if err := s.createLedgerEntries(tx, rewardID, req.UserID, stock.ID, quantity, totalValue); err != nil {
		return nil, fmt.Errorf("failed to create ledger entries: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"reward_id":    rewardID,
		"user_id":      req.UserID,
		"stock_symbol": req.StockSymbol,
		"quantity":     quantity.String(),
		"total_value":  totalValue.String(),
	}).Info("Reward created successfully")

	return &models.CreateRewardResponse{
		RewardID:       rewardID,
		UserID:         req.UserID,
		StockSymbol:    req.StockSymbol,
		Quantity:       quantity.String(),
		MarketPriceINR: currentPrice.String(),
		TotalValueINR:  totalValue.String(),
		Timestamp:      time.Now(),
	}, nil
}

func (s *RewardService) getOrCreateStock(tx *sql.Tx, symbol string) (*models.Stock, error) {
	// Try to get existing stock
	var stock models.Stock
	err := tx.QueryRow(`
		SELECT id, symbol, name, exchange, is_active, created_at, updated_at 
		FROM stocks WHERE symbol = $1`, symbol).Scan(
		&stock.ID, &stock.Symbol, &stock.Name, &stock.Exchange, 
		&stock.IsActive, &stock.CreatedAt, &stock.UpdatedAt)

	if err == nil {
		return &stock, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query stock: %w", err)
	}

	// Stock doesn't exist, create it
	stock = models.Stock{
		ID:       uuid.New(),
		Symbol:   symbol,
		Name:     s.getStockName(symbol), // This would typically come from a stock info API
		Exchange: "NSE",
		IsActive: true,
	}

	_, err = tx.Exec(`
		INSERT INTO stocks (id, symbol, name, exchange, is_active)
		VALUES ($1, $2, $3, $4, $5)`,
		stock.ID, stock.Symbol, stock.Name, stock.Exchange, stock.IsActive)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock: %w", err)
	}

	return &stock, nil
}

func (s *RewardService) getStockName(symbol string) string {
	// In a real application, this would query a stock information API
	stockNames := map[string]string{
		"RELIANCE": "Reliance Industries Limited",
		"TCS":      "Tata Consultancy Services Limited",
		"INFOSYS":  "Infosys Limited",
		"HDFCBANK": "HDFC Bank Limited",
		"ICICIBANK": "ICICI Bank Limited",
	}
	
	if name, exists := stockNames[symbol]; exists {
		return name
	}
	return fmt.Sprintf("%s Limited", symbol)
}

func (s *RewardService) createLedgerEntries(tx *sql.Tx, rewardID, userID, stockID uuid.UUID, quantity, totalValue decimal.Decimal) error {
	// Entry 1: Credit stock to user's account
	_, err := tx.Exec(`
		INSERT INTO ledger_entries (reward_id, entry_type, account_type, account_identifier, amount_inr, quantity, stock_id, description)
		VALUES ($1, 'CREDIT', 'USER_STOCK_HOLDINGS', $2, $3, $4, $5, $6)`,
		rewardID, userID.String(), totalValue, quantity, stockID, 
		fmt.Sprintf("Stock reward credited to user %s", userID.String()))
	if err != nil {
		return fmt.Errorf("failed to create user stock credit entry: %w", err)
	}

	// Entry 2: Debit from company's stock expense account
	_, err = tx.Exec(`
		INSERT INTO ledger_entries (reward_id, entry_type, account_type, account_identifier, amount_inr, quantity, stock_id, description)
		VALUES ($1, 'DEBIT', 'COMPANY_STOCK_EXPENSE', 'STOCK_REWARDS', $2, $3, $4, $5)`,
		rewardID, totalValue, quantity, stockID, "Stock reward expense")
	if err != nil {
		return fmt.Errorf("failed to create company expense debit entry: %w", err)
	}

	// Entry 3: Additional company costs (brokerage, taxes, etc.) - 0.5% of total value
	additionalCosts := totalValue.Mul(decimal.NewFromFloat(0.005)) // 0.5% fees
	_, err = tx.Exec(`
		INSERT INTO ledger_entries (reward_id, entry_type, account_type, account_identifier, amount_inr, description)
		VALUES ($1, 'DEBIT', 'COMPANY_TRADING_FEES', 'BROKERAGE_AND_TAXES', $2, $3)`,
		rewardID, additionalCosts, 
		fmt.Sprintf("Trading fees and taxes for reward %s", rewardID.String()))
	if err != nil {
		return fmt.Errorf("failed to create trading fees entry: %w", err)
	}

	// Entry 4: Credit to accounts payable (representing the liability to pay these fees)
	_, err = tx.Exec(`
		INSERT INTO ledger_entries (reward_id, entry_type, account_type, account_identifier, amount_inr, description)
		VALUES ($1, 'CREDIT', 'ACCOUNTS_PAYABLE', 'TRADING_FEES_PAYABLE', $2, $3)`,
		rewardID, additionalCosts, "Accrued trading fees payable")
	if err != nil {
		return fmt.Errorf("failed to create accounts payable entry: %w", err)
	}

	return nil
}

func (s *RewardService) AdjustReward(req *models.AdjustRewardRequest) error {
	// Parse quantity adjustment
	quantityAdj, err := decimal.NewFromString(req.QuantityAdjustment)
	if err != nil {
		return fmt.Errorf("invalid quantity adjustment: %w", err)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get original reward
	var originalReward models.Reward
	err = tx.QueryRow(`
		SELECT r.id, r.user_id, r.stock_id, r.quantity, r.market_price_inr, r.total_value_inr
		FROM rewards r WHERE r.id = $1`, req.OriginalRewardID).Scan(
		&originalReward.ID, &originalReward.UserID, &originalReward.StockID,
		&originalReward.Quantity, &originalReward.MarketPriceINR, &originalReward.TotalValueINR)
	if err != nil {
		return fmt.Errorf("failed to get original reward: %w", err)
	}

	// Create adjustment record
	adjustmentID := uuid.New()
	_, err = tx.Exec(`
		INSERT INTO adjustments (id, original_reward_id, adjustment_type, quantity_adjustment, reason, approved_by)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		adjustmentID, req.OriginalRewardID, req.AdjustmentType, quantityAdj, req.Reason, req.ApprovedBy)
	if err != nil {
		return fmt.Errorf("failed to create adjustment: %w", err)
	}

	// Create reverse ledger entries
	adjustmentValue := quantityAdj.Mul(originalReward.MarketPriceINR)
	if err := s.createAdjustmentLedgerEntries(tx, adjustmentID, originalReward, quantityAdj, adjustmentValue); err != nil {
		return fmt.Errorf("failed to create adjustment ledger entries: %w", err)
	}

	// Mark adjustment as processed
	_, err = tx.Exec("UPDATE adjustments SET processed = true WHERE id = $1", adjustmentID)
	if err != nil {
		return fmt.Errorf("failed to mark adjustment as processed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"adjustment_id":      adjustmentID,
		"original_reward_id": req.OriginalRewardID,
		"adjustment_type":    req.AdjustmentType,
		"quantity_adj":       quantityAdj.String(),
		"approved_by":        req.ApprovedBy,
	}).Info("Reward adjustment created successfully")

	return nil
}

func (s *RewardService) createAdjustmentLedgerEntries(tx *sql.Tx, adjustmentID uuid.UUID, originalReward models.Reward, quantityAdj, adjustmentValue decimal.Decimal) error {
	// For a refund, we need to reverse the original entries
	if quantityAdj.IsNegative() {
		// Debit from user's stock holdings
		_, err := tx.Exec(`
			INSERT INTO ledger_entries (adjustment_id, entry_type, account_type, account_identifier, amount_inr, quantity, stock_id, description)
			VALUES ($1, 'DEBIT', 'USER_STOCK_HOLDINGS', $2, $3, $4, $5, $6)`,
			adjustmentID, originalReward.UserID.String(), adjustmentValue.Abs(), quantityAdj.Abs(), originalReward.StockID,
			fmt.Sprintf("Stock adjustment for user %s", originalReward.UserID.String()))
		if err != nil {
			return fmt.Errorf("failed to create user stock debit entry: %w", err)
		}

		// Credit to company's stock expense account
		_, err = tx.Exec(`
			INSERT INTO ledger_entries (adjustment_id, entry_type, account_type, account_identifier, amount_inr, quantity, stock_id, description)
			VALUES ($1, 'CREDIT', 'COMPANY_STOCK_EXPENSE', 'STOCK_REWARDS', $2, $3, $4, $5)`,
			adjustmentID, adjustmentValue.Abs(), quantityAdj.Abs(), originalReward.StockID, "Stock reward adjustment")
		if err != nil {
			return fmt.Errorf("failed to create company expense credit entry: %w", err)
		}
	}

	return nil
}

func (s *RewardService) SetStockPrice(req *models.SetStockPriceRequest) error {
	// Parse price
	price, err := decimal.NewFromString(req.PriceINR)
	if err != nil {
		return fmt.Errorf("invalid price: %w", err)
	}

	// Validate price is positive
	if price.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("price must be positive")
	}

	// Get stock ID
	var stockID uuid.UUID
	err = s.db.QueryRow("SELECT id FROM stocks WHERE symbol = $1", req.StockSymbol).Scan(&stockID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("stock not found: %s", req.StockSymbol)
	}
	if err != nil {
		return fmt.Errorf("failed to get stock: %w", err)
	}

	// Insert price (with manual source)
	_, err = s.db.Exec(`
		INSERT INTO stock_prices (stock_id, price_inr, source)
		VALUES ($1, $2, 'MANUAL')
		ON CONFLICT (stock_id, timestamp) DO UPDATE SET
		price_inr = EXCLUDED.price_inr, source = EXCLUDED.source`,
		stockID, price)
	if err != nil {
		return fmt.Errorf("failed to set stock price: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"stock_symbol": req.StockSymbol,
		"price_inr":    price.String(),
		"source":       "MANUAL",
	}).Info("Stock price set manually")

	return nil
}