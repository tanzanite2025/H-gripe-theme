#!/bin/bash

# Tanzanite Go Backend Startup Script

set -e

echo "🚀 Starting Tanzanite Go Backend..."

# Check if config file exists
if [ ! -f "config/config.yaml" ]; then
    echo "⚠️  Config file not found. Copying from example..."
    cp config/config.example.yaml config/config.yaml
    echo "✅ Config file created. Please update config/config.yaml with your settings."
fi

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "⚠️  .env file not found. Copying from example..."
    cp .env.example .env
    echo "✅ .env file created. Please update .env with your settings."
fi

# Download dependencies
echo "📦 Downloading dependencies..."
go mod download

# Build the application
echo "🔨 Building application..."
go build -o tanzanite-api ./cmd/server

# Run the application
echo "✅ Starting server..."
./tanzanite-api
