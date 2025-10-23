#!/bin/bash

# Quick Test Script for Orders API
# Make this file executable: chmod +x quick-test-orders.sh

set -e  # Exit on error

BASE_URL="http://localhost:8080"
echo "🚀 Testing Orders API at $BASE_URL"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Login
echo "📝 Step 1: Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_buyer",
    "password": "Test@1234"
  }')

# Extract token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo -e "${RED}❌ Login failed. Response:${NC}"
  echo "$LOGIN_RESPONSE" | jq
  echo ""
  echo -e "${YELLOW}💡 Make sure you've registered the user first:${NC}"
  echo "curl -X POST $BASE_URL/api/v1/auth/register \\"
  echo "  -H 'Content-Type: application/json' \\"
  echo "  -d '{\"username\":\"test_buyer\",\"email\":\"buyer@test.com\",\"password\":\"Test@1234\",\"phone_number\":\"+919876543210\"}'"
  exit 1
fi

echo -e "${GREEN}✅ Login successful${NC}"
echo "Token: ${TOKEN:0:50}..."
echo ""

# Decode token to check organization (requires jq and base64)
echo "🔍 Checking token for organization ID..."
PAYLOAD=$(echo "$TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null || echo "{}")
ORG_ID=$(echo "$PAYLOAD" | jq -r '.organization_id // .OrganizationID // empty')

if [ -z "$ORG_ID" ] || [ "$ORG_ID" = "null" ]; then
  echo -e "${RED}❌ No organization_id in token!${NC}"
  echo "Token payload:"
  echo "$PAYLOAD" | jq
  echo ""
  echo -e "${YELLOW}⚠️  You need to assign this user to an organization in the AAA service.${NC}"
  echo "The orders API will return 401 without a valid organization_id in the token."
  exit 1
fi

echo -e "${GREEN}✅ Found organization ID: $ORG_ID${NC}"
echo ""

# Step 2: List existing orders
echo "📋 Step 2: Listing existing orders..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/orders?page=1&limit=5" \
  -H "Authorization: Bearer $TOKEN")

echo "$LIST_RESPONSE" | jq

# Check if it failed
if echo "$LIST_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
  echo -e "${RED}❌ Failed to list orders${NC}"
  echo ""
  echo -e "${YELLOW}Error details:${NC}"
  echo "$LIST_RESPONSE" | jq '.error'
  exit 1
fi

TOTAL_ORDERS=$(echo "$LIST_RESPONSE" | jq -r '.meta.pagination.total // 0')
echo -e "${GREEN}✅ Found $TOTAL_ORDERS existing orders${NC}"
echo ""

# Step 3: Create a new order
echo "🛒 Step 3: Creating a new order..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"buyer_organization_id\": \"$ORG_ID\",
    \"seller_organization_id\": \"org-seller-123\",
    \"items\": [
      {
        \"catalog_item_id\": \"prod-test-001\",
        \"catalog_item_type\": \"product\",
        \"quantity\": 5,
        \"unit_price\": 100.00,
        \"currency\": \"INR\",
        \"tax_percentage\": 18
      }
    ],
    \"delivery_address\": {
      \"street\": \"123 Test Street\",
      \"city\": \"Bangalore\",
      \"state\": \"Karnataka\",
      \"postal_code\": \"560001\",
      \"country\": \"India\"
    },
    \"notes\": \"Test order from script\"
  }")

echo "$CREATE_RESPONSE" | jq

# Check if creation failed
if echo "$CREATE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
  echo -e "${RED}❌ Failed to create order${NC}"
  echo ""
  echo -e "${YELLOW}Error details:${NC}"
  echo "$CREATE_RESPONSE" | jq '.error'
  exit 1
fi

ORDER_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id // empty')
ORDER_NUMBER=$(echo "$CREATE_RESPONSE" | jq -r '.data.order_number // empty')

if [ -z "$ORDER_ID" ]; then
  echo -e "${RED}❌ No order ID in response${NC}"
  exit 1
fi

echo -e "${GREEN}✅ Created order: $ORDER_NUMBER ($ORDER_ID)${NC}"
echo ""

# Step 4: Get the order details
echo "🔍 Step 4: Getting order details..."
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/orders/$ORDER_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "$GET_RESPONSE" | jq

if echo "$GET_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
  echo -e "${RED}❌ Failed to get order${NC}"
else
  echo -e "${GREEN}✅ Retrieved order successfully${NC}"
fi
echo ""

# Step 5: Update order status
echo "📝 Step 5: Updating order status to 'confirmed'..."
STATUS_RESPONSE=$(curl -s -X PATCH "$BASE_URL/api/v1/orders/$ORDER_ID/status" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "status": "confirmed",
    "notes": "Status updated by test script"
  }')

echo "$STATUS_RESPONSE" | jq

if echo "$STATUS_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
  echo -e "${RED}❌ Failed to update status${NC}"
else
  echo -e "${GREEN}✅ Status updated successfully${NC}"
fi
echo ""

# Step 6: List orders again
echo "📋 Step 6: Listing orders again..."
LIST_RESPONSE2=$(curl -s -X GET "$BASE_URL/api/v1/orders?page=1&limit=5" \
  -H "Authorization: Bearer $TOKEN")

TOTAL_ORDERS2=$(echo "$LIST_RESPONSE2" | jq -r '.meta.pagination.total // 0')
echo -e "${GREEN}✅ Now showing $TOTAL_ORDERS2 total orders${NC}"
echo "$LIST_RESPONSE2" | jq '.data[] | {id, order_number, status, total_amount}'
echo ""

# Summary
echo "======================================"
echo -e "${GREEN}🎉 Test completed successfully!${NC}"
echo "======================================"
echo ""
echo "Summary:"
echo "  - Organization ID: $ORG_ID"
echo "  - Created Order: $ORDER_NUMBER"
echo "  - Order ID: $ORDER_ID"
echo "  - Total Orders: $TOTAL_ORDERS2"
echo ""
echo "Next steps:"
echo "  - Update delivery address: curl -X PUT $BASE_URL/api/v1/orders/$ORDER_ID ..."
echo "  - Process payment: curl -X POST $BASE_URL/api/v1/orders/$ORDER_ID/payment ..."
echo "  - Cancel order: curl -X POST $BASE_URL/api/v1/orders/$ORDER_ID/cancel ..."
echo ""
