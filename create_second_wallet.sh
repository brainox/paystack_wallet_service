#!/bin/bash

# Script to create a second wallet for transfer testing

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

BASE_URL="https://pure-plateau-79480-6fc7adb7399c.herokuapp.com"

echo -e "${YELLOW}🔐 Create Second Wallet for Transfer Testing${NC}"
echo "=============================================="
echo ""
echo "Option 1: Sign in with a different Google account"
echo "   - Open: $BASE_URL/auth/google"
echo "   - Sign in with a DIFFERENT Google account"
echo "   - This will create a new user and wallet"
echo ""
echo -e "${YELLOW}Opening browser...${NC}"
open "$BASE_URL/auth/google" 2>/dev/null

echo ""
echo "After signing in, copy the JWT token and run:"
echo ""
echo -e "${GREEN}export JWT_TOKEN_2='paste_second_token_here'${NC}"
echo ""
echo "Then get the second wallet number:"
echo -e "${GREEN}curl -s -X GET $BASE_URL/wallet/balance -H \"Authorization: Bearer \$JWT_TOKEN_2\" | jq -r '.wallet_number'${NC}"
echo ""
echo "Use that wallet number in the transfer test!"
echo ""
