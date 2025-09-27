# Database Schema Documentation

## Overview

The Stocky database uses PostgreSQL with a double-entry ledger system to track stock rewards and company expenses. The schema is designed to handle:

- Stock reward events
- User portfolio tracking
- Company expense tracking
- Historical price data
- Corporate actions (splits, mergers)
- Audit trail for adjustments

## Tables

### users

Stores user information.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### stocks

Master data for all stocks.

```sql
CREATE TABLE stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    exchange VARCHAR(10) NOT NULL DEFAULT 'NSE',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### rewards

Records of stock rewards given to users.

```sql
CREATE TABLE rewards (
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

    UNIQUE(user_id, idempotency_key)
);
```

### ledger_entries

Double-entry bookkeeping system for all transactions.

```sql
CREATE TABLE ledger_entries (
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
);
```

### stock_prices

Historical stock price data.

```sql
CREATE TABLE stock_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_id UUID NOT NULL REFERENCES stocks(id),
    price_inr NUMERIC(18, 4) NOT NULL CHECK (price_inr > 0),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    source VARCHAR(50) DEFAULT 'API',

    UNIQUE(stock_id, timestamp)
);
```

### portfolio_snapshots

Daily portfolio snapshots for historical tracking.

```sql
CREATE TABLE portfolio_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    stock_id UUID NOT NULL REFERENCES stocks(id),
    quantity NUMERIC(18, 6) NOT NULL,
    price_inr NUMERIC(18, 4) NOT NULL,
    value_inr NUMERIC(18, 4) NOT NULL,
    snapshot_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, stock_id, snapshot_date)
);
```

### corporate_actions

Track stock splits, mergers, and other corporate actions.

```sql
CREATE TABLE corporate_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_id UUID NOT NULL REFERENCES stocks(id),
    action_type VARCHAR(20) NOT NULL CHECK (action_type IN ('SPLIT', 'MERGER', 'DELISTING')),
    ratio NUMERIC(10, 6),
    effective_date DATE NOT NULL,
    description TEXT,
    processed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### adjustments

Track adjustments and refunds to rewards.

```sql
CREATE TABLE adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_reward_id UUID NOT NULL REFERENCES rewards(id),
    adjustment_type VARCHAR(20) NOT NULL CHECK (adjustment_type IN ('REFUND', 'CORRECTION')),
    quantity_adjustment NUMERIC(18, 6) NOT NULL,
    reason TEXT NOT NULL,
    approved_by VARCHAR(255),
    processed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Account Types in Ledger

The `ledger_entries` table uses these account types:

### Asset Accounts

- `USER_STOCK_HOLDINGS` - User's stock holdings
- `COMPANY_CASH` - Company cash account

### Expense Accounts

- `COMPANY_STOCK_EXPENSE` - Cost of stocks given as rewards
- `COMPANY_TRADING_FEES` - Brokerage and regulatory fees

### Liability Accounts

- `ACCOUNTS_PAYABLE` - Outstanding payments to brokers/exchanges

## Sample Double-Entry for Reward

When a user receives 10 shares of RELIANCE at ₹2,450 per share:

```sql
-- Credit stock to user (Asset)
INSERT INTO ledger_entries (
    entry_type, account_type, account_identifier,
    amount_inr, quantity, stock_id, description
) VALUES (
    'CREDIT', 'USER_STOCK_HOLDINGS', 'user-123',
    24500.00, 10.0, 'reliance-stock-id', 'Stock reward'
);

-- Debit from company stock expense (Expense)
INSERT INTO ledger_entries (
    entry_type, account_type, account_identifier,
    amount_inr, quantity, stock_id, description
) VALUES (
    'DEBIT', 'COMPANY_STOCK_EXPENSE', 'STOCK_REWARDS',
    24500.00, 10.0, 'reliance-stock-id', 'Stock reward expense'
);

-- Debit trading fees (Expense) - 0.5% of value
INSERT INTO ledger_entries (
    entry_type, account_type, account_identifier,
    amount_inr, description
) VALUES (
    'DEBIT', 'COMPANY_TRADING_FEES', 'BROKERAGE_AND_TAXES',
    122.50, 'Trading fees for reward'
);

-- Credit accounts payable (Liability)
INSERT INTO ledger_entries (
    entry_type, account_type, account_identifier,
    amount_inr, description
) VALUES (
    'CREDIT', 'ACCOUNTS_PAYABLE', 'TRADING_FEES_PAYABLE',
    122.50, 'Accrued trading fees'
);
```

## Indexes

The schema includes performance indexes:

```sql
-- Rewards table
CREATE INDEX idx_rewards_user_id ON rewards(user_id);
CREATE INDEX idx_rewards_created_at ON rewards(created_at);
CREATE INDEX idx_rewards_user_date ON rewards(user_id, DATE(created_at));

-- Stock prices
CREATE INDEX idx_stock_prices_stock_time ON stock_prices(stock_id, timestamp DESC);

-- Portfolio snapshots
CREATE INDEX idx_portfolio_snapshots_user_date ON portfolio_snapshots(user_id, snapshot_date);

-- Ledger entries
CREATE INDEX idx_ledger_entries_account ON ledger_entries(account_type, account_identifier);
```

## Data Types

### Precision Guidelines

- **Stock quantities**: `NUMERIC(18, 6)` - Supports fractional shares up to 6 decimal places
- **INR amounts**: `NUMERIC(18, 4)` - Supports amounts up to ₹99,999,999,999,999 with 4 decimal precision
- **Ratios**: `NUMERIC(10, 6)` - For stock splits and mergers

### UUID Usage

All primary keys use UUID v4 for:

- Better distribution in sharded databases
- No sequential enumeration attacks
- Easier cross-system integration

## Backup and Recovery

### Daily Snapshots

- Full database backup daily at 2 AM IST
- Transaction log backup every 15 minutes
- Point-in-time recovery available

### Data Retention

- Rewards and ledger entries: Permanent
- Stock prices: 7 years
- Portfolio snapshots: 7 years
- Log files: 90 days
