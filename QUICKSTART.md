# Quick Start Guide

Follow these steps to get Stocky backend running locally.

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Git

## Setup Instructions

### 1. Clone Repository

```bash
git clone <repository-url>
cd stocky/backend
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Setup PostgreSQL Database

```sql
-- Connect to PostgreSQL as superuser
CREATE DATABASE assignment;
CREATE USER stocky_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE assignment TO stocky_user;
```

### 4. Configure Environment

```bash
cp .env.example .env
# Edit .env with your database credentials
```

### 5. Run Database Migrations

```bash
go run cmd/migrate/main.go
```

### 6. Seed Sample Data (Optional)

```bash
go run cmd/seed/main.go
```

### 7. Start the Server

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## Using Docker (Alternative)

### Quick Start with Docker Compose

```bash
# Start PostgreSQL and the application
docker-compose up --build

# The application will be available at http://localhost:8080
```

### Manual Docker Build

```bash
# Build the Docker image
docker build -t stocky-backend .

# Run with environment variables
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PASSWORD=your_password \
  stocky-backend
```

## Testing the API

### 1. Import Postman Collection

Import `Stocky_API.postman_collection.json` into Postman.

### 2. Test Health Check

```bash
curl http://localhost:8080/health
```

### 3. Create a Test User Reward

```bash
curl -X POST http://localhost:8080/api/v1/reward \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "stock_symbol": "RELIANCE",
    "quantity": "10.0",
    "reward_type": "onboarding"
  }'
```

### 4. Check User's Portfolio

```bash
curl http://localhost:8080/api/v1/portfolio/123e4567-e89b-12d3-a456-426614174000
```

## Sample User IDs (from seed data)

- John Doe: `123e4567-e89b-12d3-a456-426614174000`
- Jane Smith: `234e5678-f12c-23e4-b567-537625184111`
- Mike Wilson: `345e6789-012d-34f5-c678-648736295222`

## Available Stock Symbols

- RELIANCE (Reliance Industries Limited)
- TCS (Tata Consultancy Services Limited)
- INFOSYS (Infosys Limited)
- HDFCBANK (HDFC Bank Limited)
- ICICIBANK (ICICI Bank Limited)
- ITC (ITC Limited)
- HINDUNILVR (Hindustan Unilever Limited)
- SBIN (State Bank of India)

## Development Commands

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Build binaries
make build

# Run with hot reload (if you have air installed)
air
```

## Troubleshooting

### Database Connection Issues

- Ensure PostgreSQL is running
- Check database credentials in `.env`
- Verify database exists and user has permissions

### Port Already in Use

```bash
# Find process using port 8080
netstat -ano | findstr :8080
# Kill the process (Windows)
taskkill /PID <process_id> /F
```

### Missing Go Modules

```bash
go mod tidy
go mod download
```

## Next Steps

1. Review the API documentation in `docs/API.md`
2. Understand the database schema in `docs/DATABASE.md`
3. Check the README.md for detailed architecture information
4. Explore the Postman collection for all available endpoints

## Support

For issues or questions, please check:

1. The README.md file for detailed documentation
2. The API documentation for endpoint details
3. Database schema documentation for data structure
