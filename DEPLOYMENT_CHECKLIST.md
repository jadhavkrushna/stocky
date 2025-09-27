# 🚀 Deployment Checklist for Stocky Backend

## ✅ Pre-Deployment Checklist

### 1. GitHub Repository Setup

- [ ] Create GitHub repository
- [ ] Add collaborator "ashutosh@021.trade"
- [ ] Push all code to repository
- [ ] Verify CI/CD workflow runs successfully

### 2. GitHub Secrets Configuration ✅

- [x] `DOCKER_USERNAME` - Your Docker Hub username
- [x] `DOCKER_PASSWORD` - Your Docker Hub access token/password

### 3. Local Development Setup

- [ ] Install Go 1.21+
- [ ] Install PostgreSQL 12+
- [ ] Copy `.env.example` to `.env`
- [ ] Update `.env` with your database credentials
- [ ] Run `go mod tidy` to install dependencies

### 4. Database Setup

- [ ] Create PostgreSQL database named "assignment"
- [ ] Run migrations: `go run cmd/migrate/main.go`
- [ ] Seed sample data: `go run cmd/seed/main.go`

### 5. Testing

- [ ] Run server: `go run cmd/server/main.go`
- [ ] Test health endpoint: `curl http://localhost:8080/health`
- [ ] Import Postman collection and test all endpoints
- [ ] Verify all API responses are correct

## 🧪 API Testing Checklist

### Core Endpoints

- [ ] `POST /api/v1/reward` - Create stock reward
- [ ] `GET /api/v1/today-stocks/{userId}` - Get today's rewards
- [ ] `GET /api/v1/historical-inr/{userId}` - Get historical values
- [ ] `GET /api/v1/stats/{userId}` - Get user statistics
- [ ] `GET /api/v1/portfolio/{userId}` - Get portfolio details

### Sample Test Data (from seed)

Use these User IDs for testing:

- John Doe: `123e4567-e89b-12d3-a456-426614174000`
- Jane Smith: `234e5678-f12c-23e4-b567-537625184111`
- Mike Wilson: `345e6789-012d-34f5-c678-648736295222`

### Test Stock Symbols

- RELIANCE, TCS, INFOSYS, HDFCBANK, ICICIBANK, ITC, HINDUNILVR, SBIN

## 📋 Production Deployment Steps

### 1. Environment Configuration

```bash
# Production .env
DB_HOST=your-prod-db-host
DB_PORT=5432
DB_USER=your-prod-user
DB_PASSWORD=your-secure-password
DB_NAME=assignment
GIN_MODE=release
LOG_LEVEL=info
```

### 2. Docker Deployment

```bash
# Build and run with Docker
docker build -t stocky-backend .
docker run -p 8080:8080 --env-file .env stocky-backend

# Or use Docker Compose
docker-compose up --build -d
```

### 3. Database Migration in Production

```bash
# Run migrations
./bin/migrate

# Optionally seed initial stock data (not user data)
./bin/seed
```

### 4. Health Checks

```bash
# Verify application is running
curl http://your-domain:8080/health

# Check database connectivity
curl http://your-domain:8080/health | jq '.data.checks.database.status'
```

## 🔍 Monitoring Setup

### Application Monitoring

- [ ] Set up log aggregation (ELK stack, CloudWatch, etc.)
- [ ] Configure error tracking (Sentry, Bugsnag, etc.)
- [ ] Set up metrics collection (Prometheus, DataDog, etc.)
- [ ] Configure uptime monitoring (Pingdom, UptimeRobot, etc.)

### Database Monitoring

- [ ] Monitor database performance
- [ ] Set up automated backups
- [ ] Configure connection pool monitoring
- [ ] Track slow query logs

### Security

- [ ] Enable HTTPS/TLS
- [ ] Set up proper firewall rules
- [ ] Configure rate limiting in reverse proxy
- [ ] Regular security updates

## 📊 Performance Optimization

### Database

- [ ] Optimize queries with EXPLAIN ANALYZE
- [ ] Set up read replicas if needed
- [ ] Configure connection pooling
- [ ] Implement query caching

### Application

- [ ] Enable gzip compression
- [ ] Set up Redis for caching (optional)
- [ ] Configure proper logging levels
- [ ] Optimize Docker image size

## 🚨 Troubleshooting Guide

### Common Issues

**Database Connection Failed**

```bash
# Check database is running
pg_isready -h localhost -p 5432

# Verify credentials
psql -h localhost -U postgres -d assignment
```

**Port Already in Use**

```bash
# Find process using port 8080
netstat -tulpn | grep :8080
# Kill the process
kill -9 <process_id>
```

**Docker Build Failed**

```bash
# Clean Docker cache
docker system prune -f

# Rebuild without cache
docker build --no-cache -t stocky-backend .
```

## 📝 Final Verification

### Code Quality

- [ ] All tests pass: `go test ./...`
- [ ] Linter passes: `golangci-lint run`
- [ ] Code coverage > 80%
- [ ] No security vulnerabilities

### Documentation

- [ ] README.md is complete and accurate
- [ ] API documentation matches implementation
- [ ] Database schema is documented
- [ ] All environment variables documented

### CI/CD Pipeline

- [ ] All GitHub Actions workflows pass
- [ ] Docker images build successfully
- [ ] Automated tests run on PR/push
- [ ] Deployment workflow is tested

## 🎯 Success Criteria

- ✅ All API endpoints respond correctly
- ✅ Database operations work without errors
- ✅ Stock price updates run automatically
- ✅ Ledger entries balance correctly (debits = credits)
- ✅ Error handling works properly
- ✅ Performance meets requirements
- ✅ Security measures are in place
- ✅ Monitoring and logging configured

## 📞 Support

For deployment issues:

1. Check the logs: `docker logs <container-id>`
2. Verify environment variables
3. Test database connectivity
4. Review the troubleshooting guide
5. Check GitHub Actions logs for CI/CD issues

---

**🎉 Once all items are checked, your Stocky backend is ready for production!**
