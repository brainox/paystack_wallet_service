# Quick Start Guide

## 🚀 Fast Setup (< 5 minutes)

### 1. Prerequisites
- Go 1.21.5+
- PostgreSQL (or Docker)
- Google OAuth credentials
- Paystack API keys

### 2. Quick Setup
```bash
# Run the setup script
./setup.sh

# OR manually:
cp .env.example .env
# Edit .env with your credentials
docker-compose up -d  # Start PostgreSQL
go mod tidy           # Install dependencies
make migrate-up       # Run migrations
go run main.go        # Start server
```

### 3. First API Call
```bash
# Test health endpoint
curl http://localhost:8080/health

# Should return: {"status":"ok"}
```

## 📌 Essential Commands

```bash
# Development
make run              # Run the application
make build            # Build binary
make test             # Run tests

# Database
make migrate-up       # Apply migrations
make migrate-down     # Rollback migrations
make migrate-create NAME=add_users  # Create new migration

# Docker
docker-compose up -d  # Start database
docker-compose down   # Stop database
docker-compose logs   # View logs
```

## 🔑 Get Your First JWT Token

1. Open browser: `http://localhost:8080/auth/google`
2. Sign in with Google
3. Copy the JWT token from response
4. Save it: `export JWT_TOKEN="your_token"`

## 💰 Make Your First Deposit

```bash
# Initiate deposit
curl -X POST http://localhost:8080/wallet/deposit \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000}'

# You'll get a Paystack payment link
# Complete payment on Paystack
# Webhook will credit your wallet automatically
```

## 🔐 Create Your First API Key

```bash
curl -X POST http://localhost:8080/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-service",
    "permissions": ["deposit", "transfer", "read"],
    "expiry": "1D"
  }'

# Save the API key
export API_KEY="sk_live_..."
```

## 📊 Check Balance

```bash
# Using JWT
curl -X GET http://localhost:8080/wallet/balance \
  -H "Authorization: Bearer $JWT_TOKEN"

# Using API Key
curl -X GET http://localhost:8080/wallet/balance \
  -H "x-api-key: $API_KEY"
```

## 💸 Transfer Money

```bash
curl -X POST http://localhost:8080/wallet/transfer \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_number": "1234567890123",
    "amount": 1000
  }'
```

## 📜 View Transactions

```bash
curl -X GET http://localhost:8080/wallet/transactions \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## 🎯 Key Features

✅ Google OAuth Login → JWT Token  
✅ API Keys with Permissions  
✅ Paystack Deposit Integration  
✅ Wallet-to-Wallet Transfer  
✅ Transaction History  
✅ Webhook Verification  
✅ Balance Inquiry  

## 📚 Full Documentation

- `README.md` - Complete documentation
- `API_TESTS.md` - All API test cases
- `.env.example` - Configuration template

## ⚡ Production Checklist

- [ ] Use strong JWT secret (32+ chars)
- [ ] Enable HTTPS/TLS
- [ ] Use Paystack live keys
- [ ] Set up proper CORS
- [ ] Add rate limiting
- [ ] Configure logging
- [ ] Set up monitoring
- [ ] Backup database regularly
- [ ] Test webhook in production
- [ ] Secure environment variables

## 🆘 Common Issues

### Database connection failed
```bash
# Check PostgreSQL is running
docker-compose ps

# Or restart it
docker-compose restart postgres
```

### Migration errors
```bash
# Check migration status
migrate -path db/migrations -database "$DB_URL" version

# Force version (if stuck)
migrate -path db/migrations -database "$DB_URL" force VERSION
```

### Port already in use
```bash
# Change port in .env
SERVER_PORT=8081

# Or kill process on port 8080
lsof -ti:8080 | xargs kill -9
```

## 🔧 Environment Variables

Essential variables in `.env`:
```bash
# Server
SERVER_PORT=8080

# Database
DB_PASSWORD=your_password

# Auth
JWT_SECRET=your_secret_key
GOOGLE_CLIENT_ID=your_client_id
GOOGLE_CLIENT_SECRET=your_client_secret

# Payment
PAYSTACK_SECRET_KEY=sk_test_xxx
PAYSTACK_PUBLIC_KEY=pk_test_xxx
```

## 📞 Support

For issues or questions:
- Check `README.md` for detailed docs
- Review `API_TESTS.md` for examples
- Check error logs for details
