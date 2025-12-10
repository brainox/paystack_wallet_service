# Wallet Service with Paystack, JWT & API Keys

A production-grade backend wallet service built with Go that enables users to deposit money using Paystack, manage wallet balances, view transaction history, and transfer funds to other users. The service supports both JWT authentication (via Google sign-in) and API key-based service-to-service access.

## Features

- ✅ Google OAuth authentication with JWT token generation
- ✅ Wallet creation per user with unique wallet numbers
- ✅ Paystack integration for deposits
- ✅ Mandatory webhook handling for transaction verification
- ✅ Wallet-to-wallet transfers with ACID compliance
- ✅ API key management with permissions and expiry
- ✅ Maximum 5 active API keys per user
- ✅ API key rollover for expired keys
- ✅ Transaction history with pagination
- ✅ Balance checking with proper authentication

## Tech Stack

- **Language**: Go 1.21.5
- **Web Framework**: Gin
- **Database**: PostgreSQL with sqlx
- **Authentication**: JWT (golang-jwt/jwt) & Google OAuth2
- **Payment Gateway**: Paystack
- **Database Migrations**: SQL migrations

## Project Structure

```
.
├── db/
│   └── migrations/          # Database migration files
├── external/
│   └── external_models/     # External API models (Paystack)
├── internal/
│   ├── config/             # Configuration management
│   └── models/             # Domain models
├── pkg/
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # Authentication middleware
│   └── router/             # Route definitions
├── services/
│   ├── auth/               # JWT & API key services
│   ├── database/           # Database connection
│   ├── paystack/           # Paystack integration
│   ├── repository/         # Data access layer
│   └── wallet/             # Wallet business logic
├── main.go                 # Application entry point
├── go.mod                  # Go modules
└── .env.example            # Environment variables template
```

## Prerequisites

- Go 1.21.5 or higher
- PostgreSQL 12 or higher
- Paystack account (for test/live keys)
- Google OAuth credentials

## Setup Instructions

### 1. Clone the repository

```bash
cd /Users/aguwa/Developer/HNG/paystack_wallet_service
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Set up PostgreSQL database

```bash
createdb wallet_service
```

### 4. Run migrations

Install golang-migrate:

```bash
# macOS
brew install golang-migrate

# Or download from https://github.com/golang-migrate/migrate
```

Run migrations:

```bash
migrate -path db/migrations -database "postgresql://postgres:your_password@localhost:5432/wallet_service?sslmode=disable" up
```

### 5. Configure environment variables

Copy the example env file and update with your credentials:

```bash
cp .env.example .env
```

Edit `.env` with your actual values:
- `JWT_SECRET`: A strong random secret for JWT signing
- `GOOGLE_CLIENT_ID` & `GOOGLE_CLIENT_SECRET`: From Google Cloud Console
- `PAYSTACK_SECRET_KEY` & `PAYSTACK_PUBLIC_KEY`: From Paystack Dashboard
- `DB_PASSWORD`: Your PostgreSQL password

### 6. Run the application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Authentication

#### 1. Google Sign-In
```
GET /auth/google
```
Redirects to Google OAuth consent screen.

#### 2. Google Callback
```
GET /auth/google/callback
```
Handles OAuth callback and returns JWT token.

**Response:**
```json
{
  "token": "eyJhbGc...",
  "message": "Login successful"
}
```

### API Key Management

#### 3. Create API Key
```
POST /keys/create
Authorization: Bearer <jwt_token>
```

**Request:**
```json
{
  "name": "wallet-service",
  "permissions": ["deposit", "transfer", "read"],
  "expiry": "1D"
}
```

**Expiry Options:** `1H` (hour), `1D` (day), `1M` (month), `1Y` (year)

**Response:**
```json
{
  "api_key": "sk_live_...",
  "expires_at": "2025-01-01T12:00:00Z"
}
```

#### 4. Rollover Expired API Key
```
POST /keys/rollover
Authorization: Bearer <jwt_token>
```

**Request:**
```json
{
  "expired_key_id": "uuid-of-expired-key",
  "expiry": "1M"
}
```

### Wallet Operations

#### 5. Initiate Deposit
```
POST /wallet/deposit
Authorization: Bearer <jwt_token> OR x-api-key: <api_key>
```

**Request:**
```json
{
  "amount": 5000
}
```

**Response:**
```json
{
  "reference": "DEP_12345678_1234567890",
  "authorization_url": "https://paystack.co/checkout/..."
}
```

#### 6. Paystack Webhook (Mandatory)
```
POST /wallet/paystack/webhook
x-paystack-signature: <signature>
```

**Note:** This endpoint is called by Paystack. Ensure your webhook URL is configured in Paystack dashboard.

#### 7. Get Deposit Status
```
GET /wallet/deposit/{reference}/status
Authorization: Bearer <jwt_token> OR x-api-key: <api_key>
```

**Response:**
```json
{
  "reference": "DEP_12345678_1234567890",
  "status": "success",
  "amount": 5000
}
```

#### 8. Get Wallet Balance
```
GET /wallet/balance
Authorization: Bearer <jwt_token> OR x-api-key: <api_key>
```

**Response:**
```json
{
  "balance": 15000
}
```

#### 9. Transfer Funds
```
POST /wallet/transfer
Authorization: Bearer <jwt_token> OR x-api-key: <api_key>
```

**Request:**
```json
{
  "wallet_number": "4566678954356",
  "amount": 3000
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Transfer completed"
}
```

#### 10. Get Transaction History
```
GET /wallet/transactions?limit=50&offset=0
Authorization: Bearer <jwt_token> OR x-api-key: <api_key>
```

**Response:**
```json
[
  {
    "type": "deposit",
    "amount": 5000,
    "status": "success"
  },
  {
    "type": "transfer",
    "amount": 3000,
    "status": "success"
  }
]
```

## Authentication Methods

### JWT Authentication (Users)
```bash
Authorization: Bearer <jwt_token>
```
- Full access to all wallet operations
- Obtained via Google OAuth

### API Key Authentication (Services)
```bash
x-api-key: <api_key>
```
- Permission-based access
- Must have valid permissions: `deposit`, `transfer`, `read`
- Maximum 5 active keys per user
- Keys expire based on configured duration

## API Key Permissions

- **deposit**: Allows initiating deposits
- **transfer**: Allows wallet-to-wallet transfers
- **read**: Allows viewing balance and transaction history

## Security Features

- ✅ Paystack webhook signature validation
- ✅ JWT token expiry and validation
- ✅ API key hashing (SHA-256)
- ✅ API key expiration enforcement
- ✅ Permission-based access control
- ✅ Database-level balance constraints
- ✅ Transaction locking for ACID compliance
- ✅ Idempotent webhook processing

## Error Handling

The service returns clear error messages:

```json
{
  "error": "insufficient balance"
}
```

Common errors:
- `insufficient balance` - Not enough funds for transfer
- `Invalid or expired API key` - API key is invalid/expired/revoked
- `Insufficient permissions` - API key lacks required permission
- `maximum of 5 active API keys allowed` - API key limit reached
- `wallet not found` - Invalid wallet number

## Database Schema

### Users
- `id` (UUID, PK)
- `email` (unique)
- `google_id` (unique)
- `name`

### Wallets
- `id` (UUID, PK)
- `user_id` (FK to users)
- `wallet_number` (unique, 13 digits)
- `balance` (decimal, ≥ 0)

### Transactions
- `id` (UUID, PK)
- `user_id`, `wallet_id` (FKs)
- `type` (deposit, transfer, credit, debit)
- `amount` (decimal, > 0)
- `status` (pending, success, failed)
- `reference` (unique)
- `paystack_reference`

### API Keys
- `id` (UUID, PK)
- `user_id` (FK)
- `key_hash` (SHA-256)
- `permissions` (array)
- `expires_at` (timestamp)
- `is_active` (boolean)

## Testing

### Test the health endpoint
```bash
curl http://localhost:8080/health
```

### Test Google OAuth flow
1. Visit `http://localhost:8080/auth/google` in browser
2. Complete Google sign-in
3. Copy the JWT token from response

### Test deposit with JWT
```bash
curl -X POST http://localhost:8080/wallet/deposit \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000}'
```

### Test with API key
```bash
curl -X GET http://localhost:8080/wallet/balance \
  -H "x-api-key: YOUR_API_KEY"
```

## Paystack Webhook Setup

1. Go to Paystack Dashboard → Settings → Webhooks
2. Add webhook URL: `https://your-domain.com/wallet/paystack/webhook`
3. The service automatically validates signatures

## Development Notes

- All monetary amounts are in Naira (NGN)
- Paystack uses kobo (100 kobo = 1 Naira)
- Wallet numbers are 13-digit unique identifiers
- Transactions are atomic with database-level locking
- Webhooks are idempotent (no double-crediting)

## Production Considerations

- [ ] Use strong JWT secrets (32+ characters)
- [ ] Enable HTTPS/TLS
- [ ] Set up proper database connection pooling
- [ ] Implement rate limiting
- [ ] Add comprehensive logging
- [ ] Set up monitoring and alerts
- [ ] Use Paystack live keys
- [ ] Configure CORS properly
- [ ] Add request validation
- [ ] Implement audit trails

## License

This project is part of the HNG Internship program.

## Support

For issues or questions, please create an issue in the repository.
