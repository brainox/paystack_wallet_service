#!/bin/bash

# Wallet Service Setup Script
# This script helps you set up the wallet service quickly

set -e

echo "🚀 Wallet Service Setup Script"
echo "==============================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Go is installed
echo "📦 Checking prerequisites..."
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go 1.21.5 or higher.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Go is installed: $(go version)${NC}"

# Check if PostgreSQL is available
if ! command -v psql &> /dev/null; then
    echo -e "${YELLOW}⚠️  PostgreSQL client not found. You'll need to install it or use Docker.${NC}"
    USE_DOCKER="yes"
else
    echo -e "${GREEN}✅ PostgreSQL client is available${NC}"
    USE_DOCKER="no"
fi

# Check if Docker is installed (for optional database)
if command -v docker &> /dev/null; then
    echo -e "${GREEN}✅ Docker is installed${NC}"
fi

echo ""
echo "📋 Setup Steps:"
echo "1. Create .env file"
echo "2. Start database"
echo "3. Download Go dependencies"
echo "4. Run database migrations"
echo "5. Build the application"
echo ""

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo -e "${GREEN}✅ .env file created${NC}"
    echo -e "${YELLOW}⚠️  Please update .env with your actual credentials:${NC}"
    echo "   - JWT_SECRET"
    echo "   - GOOGLE_CLIENT_ID"
    echo "   - GOOGLE_CLIENT_SECRET"
    echo "   - PAYSTACK_SECRET_KEY"
    echo "   - PAYSTACK_PUBLIC_KEY"
    echo "   - DB_PASSWORD"
    echo ""
    read -p "Press Enter to continue after updating .env..."
else
    echo -e "${GREEN}✅ .env file already exists${NC}"
fi

# Ask if user wants to use Docker for database
if [ "$USE_DOCKER" = "yes" ]; then
    read -p "🐳 Would you like to start PostgreSQL using Docker? (y/n): " START_DOCKER
    if [ "$START_DOCKER" = "y" ] || [ "$START_DOCKER" = "Y" ]; then
        echo "🐳 Starting PostgreSQL with Docker..."
        docker-compose up -d postgres
        echo -e "${GREEN}✅ PostgreSQL started${NC}"
        echo "⏳ Waiting for PostgreSQL to be ready..."
        sleep 5
    fi
else
    echo "📊 Make sure PostgreSQL is running..."
    read -p "Is your PostgreSQL running? (y/n): " PG_RUNNING
    if [ "$PG_RUNNING" != "y" ] && [ "$PG_RUNNING" != "Y" ]; then
        echo -e "${RED}❌ Please start PostgreSQL and run this script again.${NC}"
        exit 1
    fi
fi

# Download Go dependencies
echo ""
echo "📦 Downloading Go dependencies..."
go mod download
go mod tidy
echo -e "${GREEN}✅ Dependencies downloaded${NC}"

# Check if golang-migrate is installed
if ! command -v migrate &> /dev/null; then
    echo -e "${YELLOW}⚠️  golang-migrate is not installed.${NC}"
    echo "Please install it:"
    echo "  macOS: brew install golang-migrate"
    echo "  Or visit: https://github.com/golang-migrate/migrate"
    echo ""
    read -p "Skip migrations? (y/n): " SKIP_MIGRATIONS
    if [ "$SKIP_MIGRATIONS" != "y" ] && [ "$SKIP_MIGRATIONS" != "Y" ]; then
        exit 1
    fi
else
    # Run migrations
    echo ""
    echo "🗄️  Running database migrations..."
    
    # Load DB config from .env
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | xargs)
    fi
    
    DB_URL="postgresql://${DB_USER:-postgres}:${DB_PASSWORD}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-wallet_service}?sslmode=${DB_SSL_MODE:-disable}"
    
    migrate -path db/migrations -database "$DB_URL" up
    echo -e "${GREEN}✅ Migrations completed${NC}"
fi

# Build the application
echo ""
echo "🔨 Building the application..."
go build -o bin/wallet-service main.go
echo -e "${GREEN}✅ Build successful${NC}"

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo -e "${RED}🚨 CRITICAL: Configure Webhook Public URL${NC}"
echo "Webhooks MUST be publicly accessible. Paystack cannot reach localhost."
echo ""
echo "For Development:"
echo "  1. Install ngrok: brew install ngrok"
echo "  2. Run: ngrok http 8080"
echo "  3. Copy HTTPS URL (e.g., https://abc123.ngrok.io)"
echo "  4. Add to Paystack Dashboard → Settings → Webhooks:"
echo "     https://abc123.ngrok.io/wallet/paystack/webhook"
echo ""
echo "📖 See WEBHOOK_SETUP.md for complete guide"
echo ""
echo "📚 Next steps:"
echo "1. Make sure your .env file has all required credentials"
echo "2. Configure webhook URL (see above - MANDATORY)"
echo "3. Start the server: ./bin/wallet-service or go run main.go"
echo "4. The server will run on http://localhost:8080"
echo "5. Check API_TESTS.md for testing endpoints"
echo "6. Visit http://localhost:8080/health to verify the server is running"
echo ""
echo "📖 Documentation:"
echo "- README.md - Complete project documentation"
echo "- API_TESTS.md - API testing guide with curl examples"
echo "- .env.example - Environment variables template"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"
