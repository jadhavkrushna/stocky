# API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
Most endpoints are public for this demo. Admin endpoints require the `X-Admin-API-Key` header.

## Response Format
All responses follow this format:
```json
{
  "success": true|false,
  "data": {...},
  "error": "error message",
  "message": "additional info"
}
```

## Endpoints

### POST /reward
Create a stock reward for a user.

**Request Body:**
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

### POST /adjust-reward
Adjust a previously created reward (refund/correction).

**Request Body:**
```json
{
  "original_reward_id": "456e7890-e89b-12d3-a456-426614174000",
  "adjustment_type": "REFUND",
  "quantity_adjustment": "-2.5",
  "reason": "Partial refund due to user request",
  "approved_by": "admin@stocky.com"
}
```

## Admin Endpoints

### POST /admin/stock-price
Manually set stock price (requires admin auth).

**Headers:**
```
X-Admin-API-Key: admin-secret-key-123
```

**Request Body:**
```json
{
  "stock_symbol": "RELIANCE",
  "price_inr": "2500.75"
}
```

## Error Codes

- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing/invalid admin key)
- `404` - Not Found (user/stock not found)
- `500` - Internal Server Error

## Rate Limits
No rate limits implemented in this demo version.