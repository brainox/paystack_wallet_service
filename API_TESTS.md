# API Test Cases for Wallet Service

## Prerequisites
- Server running on `http://localhost:8080`
- PostgreSQL database configured and migrated
- Google OAuth credentials configured
- Paystack API keys configured

## 1. Health Check

```bash
curl -X GET http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "ok"
}
```

## 2. Google OAuth Authentication

### Step 1: Initiate Google Login
```bash
# Open in browser
http://localhost:8080/auth/google
```

### Step 2: After Google callback
You'll receive a JWT token. Save it for subsequent requests.

**Example Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login successful"
}
```

**Save the token:**
```bash
export JWT_TOKEN="your_jwt_token_here"
```

## 3. API Key Management

### Create API Key
```bash
curl -X POST http://localhost:8080/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "wallet-service-test",
    "permissions": ["deposit", "transfer", "read"],
    "expiry": "1D"
  }'
```

**Expected Response:**
```json
{
  "api_key": "sk_live_a1b2c3d4e5f6...",
  "expires_at": "2025-12-10T16:00:00Z"
}
```

**Save the API key:**
```bash
export API_KEY="sk_live_a1b2c3d4e5f6..."
```

### Test Different Expiry Options
```bash
# 1 Hour
curl -X POST http://localhost:8080/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "hourly-key", "permissions": ["read"], "expiry": "1H"}'

# 1 Month
curl -X POST http://localhost:8080/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "monthly-key", "permissions": ["read"], "expiry": "1M"}'

# 1 Year
curl -X POST http://localhost:8080/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "yearly-key", "permissions": ["read"], "expiry": "1Y"}'
```

### Rollover Expired API Key
```bash
curl -X POST http://localhost:8080/keys/rollover \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "expired_key_id": "uuid-of-expired-key",
    "expiry": "1M"
  }'
```

## 4. Wallet Operations

### Get Wallet Balance (JWT Auth)
```bash
curl -X GET http://localhost:8080/wallet/balance \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
{
  "balance": 0
}
```

### Get Wallet Balance (API Key Auth)
```bash
curl -X GET http://localhost:8080/wallet/balance \
  -H "x-api-key: $API_KEY"
```

## 5. Wallet Deposit with Paystack

### Initiate Deposit (JWT)
```bash
curl -X POST http://localhost:8080/wallet/deposit \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 5000
  }'
```

**Expected Response:**
```json
{
  "reference": "DEP_12abc345_1733760000",
  "authorization_url": "https://checkout.paystack.com/abc123def456"
}
```

**Save the reference:**
```bash
export DEPOSIT_REF="DEP_12abc345_1733760000"
```

### Initiate Deposit (API Key)
```bash
curl -X POST http://localhost:8080/wallet/deposit \
  -H "x-api-key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 10000
  }'
```

### Check Deposit Status
```bash
curl -X GET http://localhost:8080/wallet/deposit/$DEPOSIT_REF/status \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
{
  "reference": "DEP_12abc345_1733760000",
  "status": "pending",
  "amount": 5000
}
```

After payment on Paystack, status will change to "success".

## 6. Wallet-to-Wallet Transfer

### Get Your Wallet Number First
Note: You'll need another user's wallet number for transfer. Create a second user via Google OAuth or use a test wallet number.

### Transfer Funds (JWT)
```bash
curl -X POST http://localhost:8080/wallet/transfer \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_number": "1234567890123",
    "amount": 1000
  }'
```

**Expected Response:**
```json
{
  "status": "success",
  "message": "Transfer completed"
}
```

### Transfer Funds (API Key)
```bash
curl -X POST http://localhost:8080/wallet/transfer \
  -H "x-api-key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_number": "1234567890123",
    "amount": 500
  }'
```

## 7. Transaction History

### Get All Transactions (Default)
```bash
curl -X GET http://localhost:8080/wallet/transactions \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Get Transactions with Pagination
```bash
# Get first 10 transactions
curl -X GET "http://localhost:8080/wallet/transactions?limit=10&offset=0" \
  -H "Authorization: Bearer $JWT_TOKEN"

# Get next 10 transactions
curl -X GET "http://localhost:8080/wallet/transactions?limit=10&offset=10" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Expected Response:**
```json
[
  {
    "type": "deposit",
    "amount": 5000,
    "status": "success"
  },
  {
    "type": "debit",
    "amount": 1000,
    "status": "success"
  },
  {
    "type": "credit",
    "amount": 500,
    "status": "success"
  }
]
```

## 8. Error Cases

### Insufficient Balance
```bash
curl -X POST http://localhost:8080/wallet/transfer \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_number": "1234567890123",
    "amount": 999999999
  }'
```

**Expected Response:**
```json
{
  "error": "insufficient balance"
}
```

### Invalid API Key
```bash
curl -X GET http://localhost:8080/wallet/balance \
  -H "x-api-key: invalid_key_12345"
```

**Expected Response:**
```json
{
  "error": "Invalid or expired API key"
}
```

### Missing Permission
Create an API key with only "read" permission, then try to transfer:
```bash
curl -X POST http://localhost:8080/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "read-only-key",
    "permissions": ["read"],
    "expiry": "1D"
  }'

# Save the read-only API key
export READ_ONLY_KEY="sk_live_..."

# Try to transfer (should fail)
curl -X POST http://localhost:8080/wallet/transfer \
  -H "x-api-key: $READ_ONLY_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_number": "1234567890123",
    "amount": 100
  }'
```

**Expected Response:**
```json
{
  "error": "Insufficient permissions"
}
```

### Invalid Wallet Number
```bash
curl -X POST http://localhost:8080/wallet/transfer \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_number": "0000000000000",
    "amount": 100
  }'
```

**Expected Response:**
```json
{
  "error": "recipient wallet not found: wallet not found"
}
```

### Maximum API Keys Exceeded
Create 6 API keys (should fail on the 6th):
```bash
for i in {1..6}; do
  curl -X POST http://localhost:8080/keys/create \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"key-$i\", \"permissions\": [\"read\"], \"expiry\": \"1D\"}"
  echo ""
done
```

**Expected Response on 6th request:**
```json
{
  "error": "maximum of 5 active API keys allowed"
}
```

## 9. Paystack Webhook Testing

**Note:** This endpoint is called by Paystack. For local testing, you can use ngrok or similar tools to expose your local server.

### Using ngrok
```bash
ngrok http 8080
```

Then configure the ngrok URL in Paystack dashboard:
```
https://your-ngrok-url.ngrok.io/wallet/paystack/webhook
```

### Manual Webhook Simulation (For Testing)
```bash
# This is for testing only - In production, Paystack sends this
curl -X POST http://localhost:8080/wallet/paystack/webhook \
  -H "Content-Type: application/json" \
  -H "x-paystack-signature: valid_signature_from_paystack" \
  -d '{
    "event": "charge.success",
    "data": {
      "reference": "'"$DEPOSIT_REF"'",
      "status": "success",
      "amount": 500000
    }
  }'
```

## 10. Complete Workflow Test

```bash
# 1. Login and get JWT
# (Do this manually in browser)

# 2. Create API key
API_KEY_RESPONSE=$(curl -s -X POST http://localhost:8080/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "test-key", "permissions": ["deposit", "transfer", "read"], "expiry": "1D"}')
echo "API Key Response: $API_KEY_RESPONSE"

# 3. Check balance
curl -X GET http://localhost:8080/wallet/balance \
  -H "Authorization: Bearer $JWT_TOKEN"

# 4. Initiate deposit
DEPOSIT_RESPONSE=$(curl -s -X POST http://localhost:8080/wallet/deposit \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 10000}')
echo "Deposit Response: $DEPOSIT_RESPONSE"

# 5. (Complete payment on Paystack)

# 6. Check deposit status
DEPOSIT_REF=$(echo $DEPOSIT_RESPONSE | jq -r '.reference')
curl -X GET http://localhost:8080/wallet/deposit/$DEPOSIT_REF/status \
  -H "Authorization: Bearer $JWT_TOKEN"

# 7. Check new balance
curl -X GET http://localhost:8080/wallet/balance \
  -H "Authorization: Bearer $JWT_TOKEN"

# 8. View transaction history
curl -X GET http://localhost:8080/wallet/transactions \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Notes

- Replace `$JWT_TOKEN`, `$API_KEY`, and `$DEPOSIT_REF` with actual values
- Paystack webhook requires a publicly accessible URL
- All amounts are in Naira (NGN)
- Wallet numbers are 13-digit unique identifiers
- API keys can have permissions: `deposit`, `transfer`, `read`
- JWT tokens provide full access to all endpoints
- Maximum 5 active API keys per user
