#!/bin/bash

# Development Environment Setup Script

set -e

echo "🔧 Setting up Tanzanite Go Backend development environment..."

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

echo "✅ Go version: $(go version)"

# Check Docker installation
if ! command -v docker &> /dev/null; then
    echo "⚠️  Docker is not installed. Docker is optional but recommended."
else
    echo "✅ Docker version: $(docker --version)"
fi

# Install development tools
echo "📦 Installing development tools..."

# Air for hot reload
if ! command -v air &> /dev/null; then
    echo "Installing Air (hot reload)..."
    go install github.com/cosmtrek/air@latest
fi

# golangci-lint for linting
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# Create config file
if [ ! -f "config/config.yaml" ]; then
    echo "Creating config file..."
    cp config/config.example.yaml config/config.yaml
fi

# Create .env file
if [ ! -f ".env" ]; then
    echo "Creating .env file..."
    cp .env.example .env
fi

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod download
go mod tidy

# Start Docker services (PostgreSQL + Redis)
if command -v docker-compose &> /dev/null; then
    echo "🐳 Starting Docker services..."
    docker-compose up -d postgres redis
    echo "⏳ Waiting for services to be ready..."
    sleep 5
fi

echo ""
echo "✅ Development environment setup complete!"
echo ""
echo "📝 Next steps:"
echo "  1. Update config/config.yaml with your settings"
echo "  2. Update .env with your secrets"
echo "  3. Run 'make run' or 'make dev' to start the server"
echo "  4. Visit http://localhost:9000/health to check if it's running"
echo ""
echo "🔥 For hot reload development: make dev"
echo "🧪 To run tests: make test"
echo "🐳 To start all services with Docker: make docker-up"
