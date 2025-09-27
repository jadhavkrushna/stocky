package database

import (
	"database/sql"
	"fmt"

	"stocky-backend/internal/config"

	_ "github.com/lib/pq"
)

func Connect(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
		createStocksTable,
		createRewardsTable,
		createLedgerEntriesTable,
		createStockPricesTable,
		createPortfolioSnapshotsTable,
		createCorporateActionsTable,
		createAdjustmentsTable,
		createIndexes,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createStocksTable = `
CREATE TABLE IF NOT EXISTS stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(10) NOT NULL DEFAULT 'NSE',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createRewardsTable = `
CREATE TABLE IF NOT EXISTS rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    stock_id UUID NOT NULL REFERENCES stocks(id),
    quantity NUMERIC(18, 6) NOT NULL CHECK (quantity > 0),
    market_price_inr NUMERIC(18, 4) NOT NULL CHECK (market_price_inr > 0),
    total_value_inr NUMERIC(18, 4) NOT NULL CHECK (total_value_inr > 0),
    reward_type VARCHAR(50) NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Prevent duplicate rewards
    UNIQUE(user_id, idempotency_key)
);`

const createLedgerEntriesTable = `
CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reward_id UUID REFERENCES rewards(id),
    adjustment_id UUID,
    entry_type VARCHAR(20) NOT NULL CHECK (entry_type IN ('DEBIT', 'CREDIT')),
    account_type VARCHAR(50) NOT NULL,
    account_identifier VARCHAR(255) NOT NULL,
    amount_inr NUMERIC(18, 4) NOT NULL,
    quantity NUMERIC(18, 6),
    stock_id UUID REFERENCES stocks(id),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createStockPricesTable = `
CREATE TABLE IF NOT EXISTS stock_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_id UUID NOT NULL REFERENCES stocks(id),
    price_inr NUMERIC(18, 4) NOT NULL CHECK (price_inr > 0),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    source VARCHAR(50) DEFAULT 'API',
    
    -- Ensure we don't have duplicate prices for the same timestamp
    UNIQUE(stock_id, timestamp)
);`

const createPortfolioSnapshotsTable = `
CREATE TABLE IF NOT EXISTS portfolio_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    stock_id UUID NOT NULL REFERENCES stocks(id),
    quantity NUMERIC(18, 6) NOT NULL,
    price_inr NUMERIC(18, 4) NOT NULL,
    value_inr NUMERIC(18, 4) NOT NULL,
    snapshot_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- One snapshot per user per stock per day
    UNIQUE(user_id, stock_id, snapshot_date)
);`

const createCorporateActionsTable = `
CREATE TABLE IF NOT EXISTS corporate_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_id UUID NOT NULL REFERENCES stocks(id),
    action_type VARCHAR(20) NOT NULL CHECK (action_type IN ('SPLIT', 'MERGER', 'DELISTING')),
    ratio NUMERIC(10, 6), -- For splits (e.g., 2.0 for 2:1 split)
    effective_date DATE NOT NULL,
    description TEXT,
    processed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createAdjustmentsTable = `
CREATE TABLE IF NOT EXISTS adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_reward_id UUID NOT NULL REFERENCES rewards(id),
    adjustment_type VARCHAR(20) NOT NULL CHECK (adjustment_type IN ('REFUND', 'CORRECTION')),
    quantity_adjustment NUMERIC(18, 6) NOT NULL,
    reason TEXT NOT NULL,
    approved_by VARCHAR(255),
    processed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createIndexes = `
-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_rewards_user_id ON rewards(user_id);
CREATE INDEX IF NOT EXISTS idx_rewards_created_at ON rewards(created_at);
CREATE INDEX IF NOT EXISTS idx_rewards_user_date ON rewards(user_id, DATE(created_at));

CREATE INDEX IF NOT EXISTS idx_ledger_entries_reward_id ON ledger_entries(reward_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_account ON ledger_entries(account_type, account_identifier);

CREATE INDEX IF NOT EXISTS idx_stock_prices_stock_id ON stock_prices(stock_id);
CREATE INDEX IF NOT EXISTS idx_stock_prices_timestamp ON stock_prices(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_stock_prices_stock_time ON stock_prices(stock_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_portfolio_snapshots_user_id ON portfolio_snapshots(user_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_snapshots_date ON portfolio_snapshots(snapshot_date);
CREATE INDEX IF NOT EXISTS idx_portfolio_snapshots_user_date ON portfolio_snapshots(user_id, snapshot_date);

-- Partial indexes for active stocks
CREATE INDEX IF NOT EXISTS idx_stocks_active ON stocks(symbol) WHERE is_active = true;
`