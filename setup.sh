#!/bin/bash

# Stocky Backend Setup Script
# This script helps set up the development environment

set -e  # Exit on any error

echo "🚀 Setting up Stocky Backend Development Environment..."
echo "================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        echo "Download from: https://golang.org/dl/"
        exit 1
    fi
    
    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    print_success "Go $GO_VERSION is installed"
}

# Check if PostgreSQL is available
check_postgresql() {
    if ! command -v psql &> /dev/null; then
        print_warning "PostgreSQL client not found. Make sure PostgreSQL is installed."
        print_warning "You can continue setup and install PostgreSQL later."
    else
        print_success "PostgreSQL client is available"
    fi
}

# Setup environment file
setup_env() {
    if [ ! -f .env ]; then
        print_status "Creating .env file from template..."
        cp .env.example .env
        print_success "Created .env file"
        print_warning "Please edit .env file with your database credentials before running the server"
    else
        print_success ".env file already exists"
    fi
}

# Download Go dependencies
install_deps() {
    print_status "Installing Go dependencies..."
    go mod download
    go mod tidy
    print_success "Dependencies installed"
}

# Build applications
build_apps() {
    print_status "Building applications..."
    mkdir -p bin
    
    print_status "Building server..."
    go build -o bin/server ./cmd/server
    
    print_status "Building migration tool..."
    go build -o bin/migrate ./cmd/migrate
    
    print_status "Building seed tool..."
    go build -o bin/seed ./cmd/seed
    
    print_success "All applications built successfully"
}

# Setup database (if PostgreSQL is available)
setup_database() {
    if command -v createdb &> /dev/null; then
        print_status "Setting up database..."
        
        # Check if database exists
        if psql -lqt | cut -d \| -f 1 | grep -qw assignment; then
            print_success "Database 'assignment' already exists"
        else
            createdb assignment 2>/dev/null && print_success "Database 'assignment' created" || print_warning "Could not create database. Please create it manually: createdb assignment"
        fi
        
        # Run migrations
        if [ -f bin/migrate ]; then
            print_status "Running database migrations..."
            ./bin/migrate && print_success "Migrations completed" || print_warning "Migrations failed. Check your database configuration."
        fi
        
        # Seed data
        if [ -f bin/seed ]; then
            print_status "Seeding database with sample data..."
            ./bin/seed && print_success "Database seeded with sample data" || print_warning "Seeding failed. You can run it later with: ./bin/seed"
        fi
    else
        print_warning "PostgreSQL not available. Skipping database setup."
        print_warning "Please install PostgreSQL and run: createdb assignment"
    fi
}

# Run tests
run_tests() {
    print_status "Running tests..."
    if go test -v ./...; then
        print_success "All tests passed!"
    else
        print_warning "Some tests failed. Check the output above."
    fi
}

# Main setup process
main() {
    echo
    print_status "Step 1: Checking prerequisites..."
    check_go
    check_postgresql
    
    echo
    print_status "Step 2: Setting up environment..."
    setup_env
    
    echo
    print_status "Step 3: Installing dependencies..."
    install_deps
    
    echo
    print_status "Step 4: Building applications..."
    build_apps
    
    echo
    print_status "Step 5: Setting up database..."
    setup_database
    
    echo
    print_status "Step 6: Running tests..."
    run_tests
    
    echo
    echo "🎉 Setup Complete!"
    echo "=================="
    echo
    echo "Next steps:"
    echo "1. Edit .env file with your database credentials (if needed)"
    echo "2. Start the server: ./bin/server"
    echo "3. Test the API: curl http://localhost:8080/health"
    echo "4. Import Postman collection: Stocky_API.postman_collection.json"
    echo
    echo "Sample user IDs for testing:"
    echo "- John Doe: 123e4567-e89b-12d3-a456-426614174000"
    echo "- Jane Smith: 234e5678-f12c-23e4-b567-537625184111"
    echo "- Mike Wilson: 345e6789-012d-34f5-c678-648736295222"
    echo
    echo "Available stock symbols: RELIANCE, TCS, INFOSYS, HDFCBANK, ICICIBANK"
    echo
    echo "For detailed documentation, see:"
    echo "- README.md - Complete project documentation"
    echo "- QUICKSTART.md - Quick start guide"
    echo "- docs/API.md - API documentation"
    echo "- DEPLOYMENT_CHECKLIST.md - Deployment guide"
    echo
    print_success "Happy coding! 🚀"
}

# Run main function
main "$@"