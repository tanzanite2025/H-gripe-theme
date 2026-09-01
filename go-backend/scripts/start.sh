#!/bin/bash

# Storefront Go Backend Startup Script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"
BIN_BASE="$BIN_DIR/server"

echo "🚀 Starting Storefront Go Backend..."

cd "$ROOT_DIR"

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

# Build the application into the shared bin directory
echo "🔨 Building application..."
mkdir -p "$BIN_DIR"
go build -o "$BIN_BASE" ./cmd/server

# Run the application
echo "✅ Starting server..."
if [ -f "${BIN_BASE}.exe" ]; then
    exec "${BIN_BASE}.exe"
fi
exec "$BIN_BASE"
