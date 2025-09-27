# Stocky Backend - Stock Rewards System

A comprehensive Golang backend system for managing stock rewards, built with Gin framework and PostgreSQL.

## Overview

Stocky is a platform where users can earn shares of Indian stocks (e.g., Reliance, TCS, Infosys) as incentives for various actions like onboarding, referrals, or trading milestones. The system handles:

- Recording stock rewards for users
- Tracking company expenses (brokerage, taxes, regulatory fees)
- Real-time stock price updates and portfolio valuation
- Double-entry ledger system for financial accuracy
- Comprehensive API endpoints for rewards management

## Features

### Core APIs
- **POST /reward** - Record stock rewards for users
- **GET /today-stocks/{userId}** - Get user's today stock rewards
- **GET /historical-inr/{userId}** - Get historical INR portfolio values
- **GET /stats/{userId}** - Get user portfolio statistics
- **GET /portfolio/{userId}** - Get detailed portfolio holdings

### System Features
- Double-entry ledger system
- Hourly stock price updates
- Fractional share support
- Comprehensive error handling
- Replay attack prevention
- Audit logging

## Database Schema

### Core Tables
- `users` - User management
- `stocks` - Stock master data
- `rewards` - Stock reward events
- `ledger_entries` - Double-entry bookkeeping
- `stock_prices` - Historical stock prices
- `portfolio_snapshots` - Daily portfolio valuations

## Technology Stack

- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL
- **Logging**: Logrus
- **Environment**: godotenv
- **Decimal Math**: shopspring/decimal
- **Cron Jobs**: robfig/cron
- **Testing**: testify

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 12+
- Git

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd stocky/backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Setup PostgreSQL database**
```sql
CREATE DATABASE assignment;
```

4. **Configure environment variables**
```bash
cp .env.example .env
# Edit .env with your database credentials
```

5. **Run database migrations**
```bash
go run cmd/migrate/main.go
```

6. **Start the server**
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Documentation

### POST /reward
Record a stock reward for a user.

**Request:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "stock_symbol": "RELIANCE",
  "quantity": "10.5",
  "reward_type": "onboarding",
  "metadata": {
    "campaign_id": "welcome2024"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "reward_id": "456e7890-e89b-12d3-a456-426614174000",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "stock_symbol": "RELIANCE",
    "quantity": "10.5",
    "market_price_inr": "2450.75",
    "total_value_inr": "25733.88",
    "timestamp": "2025-09-27T10:30:00Z"
  }
}
```

### GET /today-stocks/{userId}
Get all stock rewards for a user for today.

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "date": "2025-09-27",
    "rewards": [
      {
        "reward_id": "456e7890-e89b-12d3-a456-426614174000",
        "stock_symbol": "RELIANCE",
        "quantity": "10.5",
        "market_price_inr": "2450.75",
        "total_value_inr": "25733.88",
        "reward_type": "onboarding",
        "timestamp": "2025-09-27T10:30:00Z"
      }
    ],
    "total_value_inr": "25733.88"
  }
}
```

### GET /historical-inr/{userId}
Get historical INR portfolio values for a user.

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "historical_values": [
      {
        "date": "2025-09-26",
        "total_value_inr": "48500.25",
        "stocks": [
          {
            "stock_symbol": "RELIANCE",
            "quantity": "20.0",
            "price_inr": "2425.00",
            "value_inr": "48500.00"
          }
        ]
      }
    ]
  }
}
```

### GET /stats/{userId}
Get user portfolio statistics.

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "today_rewards": {
      "RELIANCE": "10.5",
      "TCS": "5.0"
    },
    "current_portfolio_value_inr": "74234.13",
    "total_stocks_rewarded": "15.5",
    "portfolio_performance": {
      "today_change_inr": "+1234.56",
      "today_change_percent": "+1.68%"
    }
  }
}
```

### GET /portfolio/{userId}
Get detailed portfolio holdings.

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "holdings": [
      {
        "stock_symbol": "RELIANCE",
        "total_quantity": "30.5",
        "current_price_inr": "2450.75",
        "current_value_inr": "74747.88",
        "average_cost_inr": "2400.00",
        "unrealized_pnl_inr": "+1547.88",
        "unrealized_pnl_percent": "+2.11%"
      }
    ],
    "total_portfolio_value_inr": "74747.88",
    "total_cost_inr": "73200.00",
    "total_unrealized_pnl_inr": "+1547.88",
    "last_updated": "2025-09-27T15:00:00Z"
  }
}
```

## Edge Cases Handled

### 1. Duplicate Reward Events / Replay Attacks
- **Idempotency Key**: Each reward request includes a unique idempotency key
- **Database Constraints**: Unique constraint on (user_id, idempotency_key)
- **Response**: Returns existing reward if duplicate detected

### 2. Stock Splits, Mergers, or Delisting
- **Event Tracking**: Corporate actions table tracks splits/mergers
- **Automatic Adjustment**: Background job adjusts quantities and prices
- **Delisting Handling**: Marks stocks as delisted, preserves historical data

### 3. Rounding Errors in INR Valuation
- **Decimal Precision**: Uses `shopspring/decimal` for precise calculations
- **Configurable Precision**: 6 decimal places for quantities, 4 for INR amounts
- **Rounding Rules**: Consistent banker's rounding throughout

### 4. Price API Downtime or Stale Data
- **Fallback Strategy**: Uses last known good price with staleness indicator
- **Circuit Breaker**: Prevents cascading failures
- **Retry Logic**: Exponential backoff for API calls
- **Manual Override**: Admin API to set prices during outages

### 5. Adjustments/Refunds of Previously Given Rewards
- **Adjustment API**: POST /adjust-reward endpoint
- **Audit Trail**: Complete history of all adjustments
- **Double-Entry**: Proper accounting entries for adjustments
- **Approval Workflow**: Multi-level approval for large adjustments

## Scaling Considerations

### Database
- **Partitioning**: Time-based partitioning for rewards and ledger tables
- **Indexing**: Optimized indexes for common query patterns
- **Read Replicas**: Separate read replicas for analytics queries

### Caching
- **Redis Integration**: Cache frequently accessed stock prices
- **Portfolio Caching**: Cache calculated portfolio values
- **Cache Invalidation**: Smart invalidation on price updates

### Background Processing
- **Queue System**: Async processing for heavy operations
- **Batch Processing**: Bulk portfolio value calculations
- **Job Monitoring**: Health checks for background jobs

### API Performance
- **Rate Limiting**: Per-user and global rate limits
- **Response Compression**: Gzip compression for large responses
- **Connection Pooling**: Optimized database connections

## Testing

Run tests with:
```bash
go test ./...
```

Run with coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Monitoring

- **Health Checks**: `/health` endpoint for service monitoring
- **Metrics**: Prometheus-compatible metrics
- **Logging**: Structured logging with correlation IDs
- **Alerts**: Critical error notifications

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.