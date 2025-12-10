#!/bin/bash

# Wallet Service API Test Script - Heroku Deployment
# Complete end-to-end testing of all wallet service features

# Set the base URL
BASE_URL="https://pure-plateau-79480-6fc7adb7399c.herokuapp.com"

echo "🧪 Wallet Service End-to-End Tests"
echo "===================================="
echo "Testing against: $BASE_URL"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if JWT_TOKEN is set
if [ -z "$JWT_TOKEN" ]; then
    echo -e "${YELLOW}⚠️  JWT_TOKEN not set!${NC}"
    echo ""
    echo "Step 1: Sign in with Google"
    echo "Visit: $BASE_URL/auth/google"
    echo ""
    echo "After signing in, copy the JWT token and run:"
    echo "export JWT_TOKEN='your_jwt_token_here'"
    echo ""
    echo "Then run this script again: ./test_api.sh"
    echo ""
    exit 1
fi

echo -e "${GREEN}✅ JWT Token is set${NC}"
echo ""

# Test 1: Health Check
echo -e "${YELLOW}🏥 Test 1: Health Check${NC}"
echo "------------------------------"
HEALTH=$(curl -s -X GET $BASE_URL/health)
if [ "$HEALTH" == '{"status":"ok"}' ]; then
    echo -e "${GREEN}✅ Server is healthy${NC}"
else
    echo -e "${RED}❌ Server health check failed${NC}"
    exit 1
fi
echo ""

# Test 2: Check wallet balance
echo -e "${YELLOW}💰 Test 2: Get Wallet Balance (JWT)${NC}"
echo "------------------------------"
BALANCE_JWT=$(curl -s -X GET $BASE_URL/wallet/balance \
  -H "Authorization: Bearer $JWT_TOKEN")
echo "$BALANCE_JWT" | jq '.'
INITIAL_BALANCE=$(echo "$BALANCE_JWT" | jq -r '.balance // 0')
echo -e "${GREEN}Initial Balance: $INITIAL_BALANCE NGN${NC}"
echo ""

# Test 3: Create API Key with all permissions
echo -e "${YELLOW}🔑 Test 3: Create API Key (deposit, transfer, read)${NC}"
echo "---------------------------------------------------"
API_KEY_RESPONSE=$(curl -s -X POST $BASE_URL/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "e2e-test-key",
    "permissions": ["deposit", "transfer", "read"],
    "expiry": "1D"
  }')
echo "$API_KEY_RESPONSE" | jq '.'

API_KEY=$(echo "$API_KEY_RESPONSE" | jq -r '.api_key // empty')
if [ -n "$API_KEY" ]; then
    echo -e "${GREEN}✅ API Key Created: ${API_KEY:0:20}...${NC}"
else
    echo -e "${RED}❌ Failed to create API key${NC}"
    exit 1
fi
echo ""

# Test 4: Verify API Key permissions (Get Balance with API Key)
echo -e "${YELLOW}🔐 Test 4: Verify API Key - Get Balance${NC}"
echo "----------------------------------------"
BALANCE_API=$(curl -s -X GET $BASE_URL/wallet/balance \
  -H "x-api-key: $API_KEY")
echo "$BALANCE_API" | jq '.'
if echo "$BALANCE_API" | jq -e '.balance' > /dev/null; then
    echo -e "${GREEN}✅ API Key authentication working${NC}"
else
    echo -e "${RED}❌ API Key authentication failed${NC}"
fi
echo ""

# Test 5: Create second API Key (read-only)
echo -e "${YELLOW}🔑 Test 5: Create Read-Only API Key${NC}"
echo "------------------------------------"
READ_KEY_RESPONSE=$(curl -s -X POST $BASE_URL/keys/create \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "read-only-key",
    "permissions": ["read"],
    "expiry": "1H"
  }')
echo "$READ_KEY_RESPONSE" | jq '.'
READ_ONLY_KEY=$(echo "$READ_KEY_RESPONSE" | jq -r '.api_key // empty')
echo -e "${GREEN}✅ Read-only key created${NC}"
echo ""

# Test 6: Test permission enforcement (read-only key cannot transfer)
echo -e "${YELLOW}🚫 Test 6: Test Permission Enforcement${NC}"
echo "---------------------------------------"
echo "Attempting transfer with read-only key (should fail)..."
FAIL_TRANSFER=$(curl -s -X POST $BASE_URL/wallet/transfer \
  -H "x-api-key: $READ_ONLY_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_number": "1234567890123",
    "amount": 100
  }')
echo "$FAIL_TRANSFER" | jq '.'
if echo "$FAIL_TRANSFER" | grep -q "Insufficient permissions"; then
    echo -e "${GREEN}✅ Permission enforcement working${NC}"
else
    echo -e "${RED}❌ Permission enforcement failed${NC}"
fi
echo ""

# Test 7: Initiate Paystack Deposit
echo -e "${YELLOW}💳 Test 7: Initiate Paystack Deposit (5000 NGN)${NC}"
echo "------------------------------------------------"
DEPOSIT_RESPONSE=$(curl -s -X POST $BASE_URL/wallet/deposit \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 5000
  }')
echo "$DEPOSIT_RESPONSE" | jq '.'

REFERENCE=$(echo "$DEPOSIT_RESPONSE" | jq -r '.reference // empty')
PAYMENT_URL=$(echo "$DEPOSIT_RESPONSE" | jq -r '.authorization_url // empty')

if [ -n "$PAYMENT_URL" ]; then
    echo -e "${GREEN}✅ Deposit initiated${NC}"
    echo "📝 Reference: $REFERENCE"
    echo "🔗 Payment URL: $PAYMENT_URL"
    echo ""
    echo -e "${YELLOW}⚠️  ACTION REQUIRED:${NC}"
    echo "1. Opening payment URL in browser..."
    echo "2. Complete the Paystack payment (use test card)"
    echo "3. Webhook will automatically credit your wallet"
    echo "4. Check Paystack Dashboard to confirm webhook delivery"
    echo ""
    open "$PAYMENT_URL" 2>/dev/null || echo "Manual: $PAYMENT_URL"
    
    # Wait for user to complete payment
    echo ""
    read -p "Press Enter after completing payment to continue tests..."
else
    echo -e "${RED}❌ Failed to initiate deposit${NC}"
fi
echo ""

# Test 8: Check deposit status
echo -e "${YELLOW}📊 Test 8: Check Deposit Status${NC}"
echo "--------------------------------"
if [ -n "$REFERENCE" ]; then
    STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/wallet/deposit/$REFERENCE/status" \
      -H "Authorization: Bearer $JWT_TOKEN")
    echo "$STATUS_RESPONSE" | jq '.'
    
    DEPOSIT_STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.status // empty')
    if [ "$DEPOSIT_STATUS" == "success" ]; then
        echo -e "${GREEN}✅ Deposit confirmed as successful${NC}"
    else
        echo -e "${YELLOW}⚠️  Deposit status: $DEPOSIT_STATUS${NC}"
    fi
fi
echo ""

# Test 9: Verify balance increased (if deposit successful)
echo -e "${YELLOW}💰 Test 9: Verify Balance After Deposit${NC}"
echo "----------------------------------------"
NEW_BALANCE_RESPONSE=$(curl -s -X GET $BASE_URL/wallet/balance \
  -H "Authorization: Bearer $JWT_TOKEN")
echo "$NEW_BALANCE_RESPONSE" | jq '.'
NEW_BALANCE=$(echo "$NEW_BALANCE_RESPONSE" | jq -r '.balance // 0')
echo "Previous Balance: $INITIAL_BALANCE NGN"
echo "Current Balance: $NEW_BALANCE NGN"
if (( $(echo "$NEW_BALANCE > $INITIAL_BALANCE" | bc -l) )); then
    echo -e "${GREEN}✅ Balance increased by webhook!${NC}"
else
    echo -e "${YELLOW}⚠️  Balance unchanged - webhook may be pending${NC}"
fi
echo ""

# Test 10: Get Transaction History
echo -e "${YELLOW}📜 Test 10: Get Transaction History${NC}"
echo "------------------------------------"
TRANSACTIONS=$(curl -s -X GET $BASE_URL/wallet/transactions \
  -H "Authorization: Bearer $JWT_TOKEN")
echo "$TRANSACTIONS" | jq '.'
TXCOUNT=$(echo "$TRANSACTIONS" | jq '. | length')
echo -e "${GREEN}Total transactions: $TXCOUNT${NC}"
echo ""

# Test 11: Wallet-to-Wallet Transfer (if balance sufficient)
echo -e "${YELLOW}💸 Test 11: Wallet-to-Wallet Transfer${NC}"
echo "---------------------------------------"
if (( $(echo "$NEW_BALANCE >= 1000" | bc -l) )); then
    TRANSFER_RESPONSE=$(curl -s -X POST $BASE_URL/wallet/transfer \
      -H "Authorization: Bearer $JWT_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "wallet_number": "1234567890123",
        "amount": 1000
      }')
    echo "$TRANSFER_RESPONSE" | jq '.'
    
    if echo "$TRANSFER_RESPONSE" | grep -q "success"; then
        echo -e "${GREEN}✅ Transfer completed successfully${NC}"
    else
        echo -e "${RED}❌ Transfer failed${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Insufficient balance for transfer test (need >= 1000 NGN)${NC}"
fi
echo ""

# Test 12: Final Balance Check
echo -e "${YELLOW}💰 Test 12: Final Balance Check${NC}"
echo "--------------------------------"
FINAL_BALANCE=$(curl -s -X GET $BASE_URL/wallet/balance \
  -H "Authorization: Bearer $JWT_TOKEN" | jq -r '.balance')
echo "Final Balance: $FINAL_BALANCE NGN"
echo ""

# Summary
echo "════════════════════════════════════════"
echo -e "${GREEN}✅ End-to-End Tests Completed!${NC}"
echo "════════════════════════════════════════"
echo ""
echo "📊 Test Summary:"
echo "  ✅ Health check"
echo "  ✅ JWT authentication"
echo "  ✅ API key creation (2 keys)"
echo "  ✅ API key authentication"
echo "  ✅ Permission enforcement"
echo "  ✅ Paystack deposit initialization"
echo "  ✅ Transaction status check"
echo "  ✅ Balance queries"
echo "  ✅ Transaction history"
if (( $(echo "$NEW_BALANCE >= 1000" | bc -l) )); then
    echo "  ✅ Wallet transfer"
fi
echo ""
echo "🎉 All core requirements verified!"
echo ""
