#!/bin/bash

# Script to help get JWT token from Heroku deployment

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

BASE_URL="https://pure-plateau-79480-6fc7adb7399c.herokuapp.com"

echo -e "${YELLOW}🔐 Get JWT Token from Heroku${NC}"
echo "================================"
echo ""
echo -e "${YELLOW}Step 1: Open Google Sign-In${NC}"
echo "Opening browser to: $BASE_URL/auth/google"
echo ""

# Open browser
open "$BASE_URL/auth/google" 2>/dev/null || echo "Please open manually: $BASE_URL/auth/google"

echo ""
echo -e "${YELLOW}Step 2: After successful sign-in${NC}"
echo "You will be redirected to a callback URL that looks like:"
echo "https://pure-plateau-79480-6fc7adb7399c.herokuapp.com/auth/google/callback?token=YOUR_JWT_TOKEN_HERE"
echo ""
echo -e "${GREEN}Step 3: Copy ONLY the JWT token from the URL${NC}"
echo "(Everything after 'token=' in the browser address bar)"
echo ""
echo -e "${YELLOW}Step 4: Export the token as environment variable${NC}"
echo "Run this command with your token:"
echo -e "${GREEN}export JWT_TOKEN='your_token_here'${NC}"
echo ""
echo "Then run the tests:"
echo -e "${GREEN}./test_api.sh${NC}"
echo ""
