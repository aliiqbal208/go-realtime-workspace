#!/bin/bash

# Quick Start Script for Organization Hub
# This script sets up the environment and starts the application

set -e  # Exit on error

echo "🚀 Organization Hub - Quick Start"
echo "================================="
echo

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install it first."
    echo "   Visit: https://docs.docker.com/compose/install/"
    exit 1
fi

echo "✅ Docker and docker-compose found"
echo

# Start databases
echo "📦 Starting PostgreSQL and Redis..."
docker-compose up -d

echo "⏳ Waiting for databases to be ready..."
sleep 10

# Check PostgreSQL
echo "🔍 Checking PostgreSQL..."
if docker exec orghub-postgres pg_isready -U postgres &> /dev/null; then
    echo "✅ PostgreSQL is ready"
else
    echo "❌ PostgreSQL is not ready. Check logs with: docker-compose logs postgres"
    exit 1
fi

# Check Redis
echo "🔍 Checking Redis..."
if docker exec orghub-redis redis-cli ping &> /dev/null; then
    echo "✅ Redis is ready"
else
    echo "❌ Redis is not ready. Check logs with: docker-compose logs redis"
    exit 1
fi

# Verify schema
echo "🔍 Verifying database schema..."
TABLES=$(docker exec orghub-postgres psql -U postgres -d orghub -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';")
if [ "$TABLES" -ge 2 ]; then
    echo "✅ Database schema initialized"
else
    echo "⚠️  Schema may not be initialized. Tables found: $TABLES"
    echo "   Initializing schema..."
    docker exec -i orghub-postgres psql -U postgres -d orghub < database/schema.sql
    echo "✅ Schema initialized"
fi

echo
echo "✅ All systems ready!"
echo
echo "📝 Next steps:"
echo "   1. Install Go dependencies: go mod download"
echo "   2. Start the application: go run main.go"
echo "   3. Test the health endpoint: curl http://localhost:8080/health"
echo
echo "📚 Documentation:"
echo "   - Database Setup: DATABASE_SETUP.md"
echo "   - API Reference: API_REFERENCE.md"
echo "   - Integration Guide: DATABASE_INTEGRATION.md"
echo
echo "🛠️  Useful commands:"
echo "   - View logs: docker-compose logs -f"
echo "   - Stop services: docker-compose down"
echo "   - Reset data: docker-compose down -v"
echo "   - PostgreSQL CLI: docker exec -it orghub-postgres psql -U postgres -d orghub"
echo "   - Redis CLI: docker exec -it orghub-redis redis-cli"
echo
