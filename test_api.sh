#!/bin/bash

# Wallet Service API Test Script
# Make sure to set your JWT_TOKEN after Google sign-in

echo "🧪 Wallet Service API Tests"
echo "=========================="
echo ""

# Check if JWT_TOKEN is set
if [ -z "$JWT_TOKEN" ]; then
    echo "⚠️  JWT_TOKEN not set!"
    echo "Please complete Google sign-in and set your token:"
    echo "export JWT_TOKEN='your_jwt_token_here'"
    echo ""
    echo "Visit: http://localhost:8080/auth/google"
    echo ""
    exit 1
fi

echo "✅ JWT Token is set"
echo ""

# Test 1: Check wallet balance
echo "📊 Test 1: Get Wallet Balance"
echo "------------------------------"
curl -s -X GET http://localhost:8080/wallet/balance \
  -H "Authorization: Bearer $JWT_TOKEN" | jq '.'
echo ""
echo ""

# Test 2: Create API Key
echo "🔑 Test 2: Create API Key"
echo "-------------------------"
API_KEY_RESPONSE=$(curl -s -X POST http://localhost:8080/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-key",
    "permissions": ["deposit", "transfer", "read"],
    "expiry": "1D"
  }')
echo "$API_KEY_RESPONSE" | jq '.'

# Extract API key for later use
API_KEY=$(echo "$API_KEY_RESPONSE" | jq -r '.api_key // empty')

if [ -n "$API_KEY" ]; then
    echo ""
    echo "💾 Saved API Key: $API_KEY"
    export API_KEY
fi
echo ""
echo ""

# Test 3: Get balance with API key
if [ -n "$API_KEY" ]; then
    echo "📊 Test 3: Get Balance with API Key"
    echo "------------------------------------"
    curl -s -X GET http://localhost:8080/wallet/balance \
      -H "x-api-key: $API_KEY" | jq '.'
    echo ""
    echo ""
fi

# Test 4: Initiate deposit
echo "💰 Test 4: Initiate Deposit (5000 NGN)"
echo "---------------------------------------"
DEPOSIT_RESPONSE=$(curl -s -X POST http://localhost:8080/wallet/deposit \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 5000
  }')
echo "$DEPOSIT_RESPONSE" | jq '.'

# Extract reference and payment URL
REFERENCE=$(echo "$DEPOSIT_RESPONSE" | jq -r '.reference // empty')
PAYMENT_URL=$(echo "$DEPOSIT_RESPONSE" | jq -r '.authorization_url // empty')

if [ -n "$PAYMENT_URL" ]; then
    echo ""
    echo "🔗 Payment URL: $PAYMENT_URL"
    echo "📝 Reference: $REFERENCE"
    echo ""
    echo "⚠️  To complete the deposit:"
    echo "1. Open the payment URL in your browser"
    echo "2. Complete the Paystack payment (test mode)"
    echo "3. The webhook will automatically credit your wallet"
    echo ""
    echo "Opening payment URL in browser..."
    open "$PAYMENT_URL" 2>/dev/null || echo "Please open manually: $PAYMENT_URL"
fi
echo ""
echo ""

# Test 5: Get transaction history
echo "📜 Test 5: Get Transaction History"
echo "-----------------------------------"
curl -s -X GET http://localhost:8080/wallet/transactions \
  -H "Authorization: Bearer $JWT_TOKEN" | jq '.'
echo ""
echo ""

echo "✅ Tests completed!"
echo ""
echo "📖 Next Steps:"
echo "1. Complete the Paystack payment if you initiated a deposit"
echo "2. Check your balance again: curl -H 'Authorization: Bearer \$JWT_TOKEN' http://localhost:8080/wallet/balance"
echo "3. Try transferring to another wallet"
echo ""
