#!/usr/bin/env bash

set -e

echo "🚀 Initializing project..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go first."
    exit 1
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed. Please install Docker first."
    exit 1
fi

echo "✅ Go and Docker are installed"

# Install tools
echo "Installing tools..."
make install

# Copy environment file
if [ !  -f .env.development ]; then
    cp .env.example .env.development
    echo "Created .env.development"
fi

# Start Docker services
echo "Starting Docker services..."
make docker-up

# Wait for database
echo "⏳ Waiting for database..."
sleep 5

# Run migrations
echo "Running migrations..."
make migrate-up

echo ""
echo "Project initialized successfully!"
echo ""
echo "Next steps:"
echo "  1. Review .env.development and adjust if needed"
echo "  2. Run 'make dev' to start development server"
echo "  3. Open http://localhost:8081 in your browser"
echo ""
