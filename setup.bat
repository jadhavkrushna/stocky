@echo off
REM Stocky Backend Setup Script for Windows
REM This script helps set up the development environment on Windows

setlocal enabledelayedexpansion

echo.
echo 🚀 Setting up Stocky Backend Development Environment...
echo =================================================
echo.

REM Function to print colored output (simplified for Windows)
goto :main

:print_status
echo [INFO] %~1
goto :eof

:print_success
echo [SUCCESS] %~1
goto :eof

:print_warning
echo [WARNING] %~1
goto :eof

:print_error
echo [ERROR] %~1
goto :eof

:check_go
call :print_status "Checking Go installation..."
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    call :print_error "Go is not installed. Please install Go 1.21 or later."
    echo Download from: https://golang.org/dl/
    exit /b 1
)
for /f "tokens=3" %%i in ('go version 2^>nul') do set GO_VERSION=%%i
call :print_success "Go !GO_VERSION! is installed"
goto :eof

:check_postgresql
call :print_status "Checking PostgreSQL..."
where psql >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    call :print_warning "PostgreSQL client not found. Make sure PostgreSQL is installed."
    call :print_warning "You can continue setup and install PostgreSQL later."
) else (
    call :print_success "PostgreSQL client is available"
)
goto :eof

:setup_env
call :print_status "Setting up environment file..."
if not exist .env (
    copy .env.example .env >nul
    call :print_success "Created .env file"
    call :print_warning "Please edit .env file with your database credentials before running the server"
) else (
    call :print_success ".env file already exists"
)
goto :eof

:install_deps
call :print_status "Installing Go dependencies..."
go mod download
go mod tidy
if %ERRORLEVEL% EQU 0 (
    call :print_success "Dependencies installed"
) else (
    call :print_error "Failed to install dependencies"
    exit /b 1
)
goto :eof

:build_apps
call :print_status "Building applications..."
if not exist bin mkdir bin

call :print_status "Building server..."
go build -o bin\server.exe .\cmd\server
if %ERRORLEVEL% NEQ 0 (
    call :print_error "Failed to build server"
    exit /b 1
)

call :print_status "Building migration tool..."
go build -o bin\migrate.exe .\cmd\migrate
if %ERRORLEVEL% NEQ 0 (
    call :print_error "Failed to build migrate tool"
    exit /b 1
)

call :print_status "Building seed tool..."
go build -o bin\seed.exe .\cmd\seed
if %ERRORLEVEL% NEQ 0 (
    call :print_error "Failed to build seed tool"
    exit /b 1
)

call :print_success "All applications built successfully"
goto :eof

:setup_database
where createdb >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    call :print_status "Setting up database..."
    
    REM Check if database exists (simplified check)
    psql -lqt 2>nul | findstr /C:"assignment" >nul
    if !ERRORLEVEL! EQU 0 (
        call :print_success "Database 'assignment' already exists"
    ) else (
        createdb assignment 2>nul
        if !ERRORLEVEL! EQU 0 (
            call :print_success "Database 'assignment' created"
        ) else (
            call :print_warning "Could not create database. Please create it manually: createdb assignment"
        )
    )
    
    REM Run migrations
    if exist bin\migrate.exe (
        call :print_status "Running database migrations..."
        bin\migrate.exe
        if !ERRORLEVEL! EQU 0 (
            call :print_success "Migrations completed"
        ) else (
            call :print_warning "Migrations failed. Check your database configuration."
        )
    )
    
    REM Seed data
    if exist bin\seed.exe (
        call :print_status "Seeding database with sample data..."
        bin\seed.exe
        if !ERRORLEVEL! EQU 0 (
            call :print_success "Database seeded with sample data"
        ) else (
            call :print_warning "Seeding failed. You can run it later with: bin\seed.exe"
        )
    )
) else (
    call :print_warning "PostgreSQL not available. Skipping database setup."
    call :print_warning "Please install PostgreSQL and run: createdb assignment"
)
goto :eof

:run_tests
call :print_status "Running tests..."
go test -v .\...
if %ERRORLEVEL% EQU 0 (
    call :print_success "All tests passed!"
) else (
    call :print_warning "Some tests failed. Check the output above."
)
goto :eof

:main
echo.
call :print_status "Step 1: Checking prerequisites..."
call :check_go
call :check_postgresql

echo.
call :print_status "Step 2: Setting up environment..."
call :setup_env

echo.
call :print_status "Step 3: Installing dependencies..."
call :install_deps

echo.
call :print_status "Step 4: Building applications..."
call :build_apps

echo.
call :print_status "Step 5: Setting up database..."
call :setup_database

echo.
call :print_status "Step 6: Running tests..."
call :run_tests

echo.
echo 🎉 Setup Complete!
echo ==================
echo.
echo Next steps:
echo 1. Edit .env file with your database credentials (if needed)
echo 2. Start the server: bin\server.exe
echo 3. Test the API: curl http://localhost:8080/health
echo 4. Import Postman collection: Stocky_API.postman_collection.json
echo.
echo Sample user IDs for testing:
echo - John Doe: 123e4567-e89b-12d3-a456-426614174000
echo - Jane Smith: 234e5678-f12c-23e4-b567-537625184111
echo - Mike Wilson: 345e6789-012d-34f5-c678-648736295222
echo.
echo Available stock symbols: RELIANCE, TCS, INFOSYS, HDFCBANK, ICICIBANK
echo.
echo For detailed documentation, see:
echo - README.md - Complete project documentation
echo - QUICKSTART.md - Quick start guide
echo - docs\API.md - API documentation
echo - DEPLOYMENT_CHECKLIST.md - Deployment guide
echo.
call :print_success "Happy coding! 🚀"

pause