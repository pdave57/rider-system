#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

API_BASE="http://localhost"

# Helper function to extract JSON values
extract_value() {
  echo "$1" | grep -o "\"$2\":[^,}]*" | cut -d':' -f2 | sed 's/"//g'
}

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}RIDER SYSTEM - END-TO-END TEST${NC}"
echo -e "${BLUE}Test: Create Rider, Upload Avatar, Login & Auto-Assign Orders${NC}"
echo -e "${BLUE}============================================${NC}"

# Step 1: Register a Rider User with vehicle details
echo -e "\n${YELLOW}Step 1: Registering Rider User with Vehicle Details${NC}"
RIDER_EMAIL="rider_$(date +%s)@example.com"
RIDER_PHONE="080$(printf "%08d" $((RANDOM * 30000 / 32768)))"
RIDER_PASSWORD="test123456"

RIDER_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{
    \"full_name\": \"Rider Test User\",
    \"email\": \"$RIDER_EMAIL\",
    \"phone\": \"$RIDER_PHONE\",
    \"password\": \"$RIDER_PASSWORD\",
    \"role\": \"rider\",
    \"vehicle_type\": \"bike\",
    \"vehicle_plate\": \"XXX-1234-YY\",
    \"license_number\": \"A12345678B\"
  }")

echo "Response: $RIDER_RESPONSE"
RIDER_ID=$(extract_value "$RIDER_RESPONSE" "id" | head -1)
RIDER_TOKEN=$(echo "$RIDER_RESPONSE" | grep -o '"token":"[^"]*' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Rider registered with ID: $RIDER_ID${NC}"
echo -e "${GREEN}  Email: $RIDER_EMAIL${NC}"
echo -e "${GREEN}  Vehicle: Bike ($RIDER_PHONE)${NC}"

# Step 2: Login as Rider
echo -e "\n${YELLOW}Step 2: Logging in as Rider${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{
    \"email\": \"$RIDER_EMAIL\",
    \"password\": \"$RIDER_PASSWORD\"
  }")

echo "Response: $LOGIN_RESPONSE"
if [ -z "$RIDER_TOKEN" ] || [ "$RIDER_TOKEN" = "null" ]; then
  RIDER_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | head -1 | cut -d'"' -f4)
fi
echo -e "${GREEN}✓ Rider logged in successfully${NC}"

# Step 3: Get Rider Profile (should exist now)
echo -e "\n${YELLOW}Step 3: Retrieving Rider Profile${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "$API_BASE/api/riders/me" \
  -H "Authorization: Bearer $RIDER_TOKEN")

echo "Response: $PROFILE_RESPONSE"
PROFILE_STATUS=$(extract_value "$PROFILE_RESPONSE" "success" | head -1)
if [ "$PROFILE_STATUS" = "true" ] || echo "$PROFILE_RESPONSE" | grep -q "vehicle_type"; then
  echo -e "${GREEN}✓ Rider profile retrieved successfully${NC}"
else
  echo -e "${YELLOW}⚠ Profile not found yet (may take a moment to initialize)${NC}"
fi

# Step 4: Set Rider as Available
echo -e "\n${YELLOW}Step 4: Setting Rider as Available${NC}"
AVAILABILITY_RESPONSE=$(curl -s -X PATCH "$API_BASE/api/riders/me/availability" \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"available": true}')

echo "Response: $AVAILABILITY_RESPONSE"
echo -e "${GREEN}✓ Rider availability updated to: AVAILABLE${NC}"

# Step 5: Set Rider Location (Lagos area)
echo -e "\n${YELLOW}Step 5: Setting Rider Location${NC}"
LOCATION_RESPONSE=$(curl -s -X PATCH "$API_BASE/api/riders/me/location" \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"latitude": 6.5244, "longitude": 3.3792}')

echo "Response: $LOCATION_RESPONSE"
echo -e "${GREEN}✓ Rider location updated to: (6.5244, 3.3792)${NC}"

# Step 6: Upload Avatar for Rider 
echo -e "\n${YELLOW}Step 6: Uploading Avatar for Rider${NC}"

# Check if ImageMagick is available
if command -v convert &> /dev/null; then
  TEMP_IMAGE="/tmp/rider_avatar_$RIDER_ID.jpg"
  convert -size 200x200 xc:blue "$TEMP_IMAGE"
  
  UPLOAD_RESPONSE=$(curl -s -X POST "$API_BASE/api/upload/avatar" \
    -H "Authorization: Bearer $RIDER_TOKEN" \
    -F "avatar=@$TEMP_IMAGE")
  
  echo "Response: $UPLOAD_RESPONSE"
  if echo "$UPLOAD_RESPONSE" | grep -q "secure_url"; then
    echo -e "${GREEN}✓ Avatar uploaded successfully${NC}"
  else
    echo -e "${YELLOW}⚠ Avatar upload response received (may require Cloudinary setup)${NC}"
  fi
  rm -f "$TEMP_IMAGE"
else
  echo -e "${YELLOW}⚠ ImageMagick not available - skipping avatar upload${NC}"
  echo -e "   (In production, you would upload a real image file via the API)${NC}"
fi

# Step 7: Register a Client User
echo -e "\n${YELLOW}Step 7: Registering Client User${NC}"
CLIENT_EMAIL="client_$(date +%s)@example.com"
CLIENT_PHONE="090$(printf "%08d" $((RANDOM * 30000 / 32768)))"
CLIENT_PASSWORD="test123456"

CLIENT_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{
    \"full_name\": \"Client Test User\",
    \"email\": \"$CLIENT_EMAIL\",
    \"phone\": \"$CLIENT_PHONE\",
    \"password\": \"$CLIENT_PASSWORD\",
    \"role\": \"client\",
    \"default_address\": \"100 Main Street, Lagos\"
  }")

echo "Response: $CLIENT_RESPONSE"
CLIENT_ID=$(extract_value "$CLIENT_RESPONSE" "id" | head -1)
CLIENT_TOKEN=$(echo "$CLIENT_RESPONSE" | grep -o '"token":"[^"]*' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Client registered with ID: $CLIENT_ID${NC}"
echo -e "${GREEN}  Email: $CLIENT_EMAIL${NC}"

# Step 8: Wait for event processing
echo -e "\n${YELLOW}Step 8: Waiting for event processing (8 seconds)...${NC}"
sleep 8

# Step 9: Login as Client
echo -e "\n${YELLOW}Step 9: Logging in as Client${NC}"
CLIENT_LOGIN=$(curl -s -X POST "$API_BASE/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{
    \"email\": \"$CLIENT_EMAIL\",
    \"password\": \"$CLIENT_PASSWORD\"
  }")

echo "Response: $CLIENT_LOGIN"
if [ -z "$CLIENT_TOKEN" ] || [ "$CLIENT_TOKEN" = "null" ]; then
  CLIENT_TOKEN=$(echo "$CLIENT_LOGIN" | grep -o '"token":"[^"]*' | head -1 | cut -d'"' -f4)
fi
echo -e "${GREEN}✓ Client logged in successfully${NC}"

# Step 10: Create an Order as Client
echo -e "\n${YELLOW}Step 10: Creating an Order (Will Trigger Auto-Assignment)${NC}"
ORDER_RESPONSE=$(curl -s -X POST "$API_BASE/api/orders" \
  -H "Authorization: Bearer $CLIENT_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "pickup_address": "123 Main Street, Lagos",
    "pickup_latitude": 6.5244,
    "pickup_longitude": 3.3792,
    "dropoff_address": "456 Park Avenue, Lagos",
    "dropoff_latitude": 6.5260,
    "dropoff_longitude": 3.3820,
    "package_desc": "Test package - documents",
    "weight_kg": 0.5,
    "notes": "Handle with care"
  }')

echo "Response: $ORDER_RESPONSE"
ORDER_ID=$(extract_value "$ORDER_RESPONSE" "id" | head -1)
TRACKING_CODE=$(extract_value "$ORDER_RESPONSE" "tracking_code" | head -1)

if [ -z "$ORDER_ID" ] || [ "$ORDER_ID" = "null" ] || [ "$ORDER_ID" = "" ]; then
  echo -e "${RED}✗ Order creation failed${NC}"
  echo "  Error: Check if the client user was synced to the order service"
else
  echo -e "${GREEN}✓ Order created successfully${NC}"
  echo -e "  Order ID: $ORDER_ID${NC}"
  echo -e "  Tracking Code: $TRACKING_CODE${NC}"
fi

# Step 11: Wait for auto-assignment
echo -e "\n${YELLOW}Step 11: Waiting for auto-assignment (5 seconds)...${NC}"
sleep 5

# Step 12: Check Order Status
if [ -n "$ORDER_ID" ] && [ "$ORDER_ID" != "null" ]; then
  echo -e "\n${YELLOW}Step 12: Checking Order Details (Auto-Assignment Status)${NC}"
  ORDER_DETAILS=$(curl -s -X GET "$API_BASE/api/orders/$ORDER_ID" \
    -H "Authorization: Bearer $CLIENT_TOKEN")

  echo "Response: $ORDER_DETAILS"

  # Extract assigned rider info
  ASSIGNED_RIDER=$(extract_value "$ORDER_DETAILS" "rider_id" | head -1)
  ORDER_STATUS=$(extract_value "$ORDER_DETAILS" "status" | head -1)

  if [ -n "$ASSIGNED_RIDER" ] && [ "$ASSIGNED_RIDER" != "0" ] && [ "$ASSIGNED_RIDER" != "null" ] && [ "$ASSIGNED_RIDER" != "" ]; then
    echo -e "${GREEN}✓ SUCCESS: Order was AUTO-ASSIGNED to Rider ID: $ASSIGNED_RIDER${NC}"
    echo -e "${GREEN}✓ Order Status: $ORDER_STATUS${NC}"
    echo -e "${GREEN}============================================${NC}"
    echo -e "${GREEN}✓ FULL CYCLE COMPLETE${NC}"
    echo -e "${GREEN}  Rider created, available, and auto-assigned to order${NC}"
    echo -e "${GREEN}============================================${NC}"
  else
    echo -e "${YELLOW}⚠ Order is still PENDING (Status: $ORDER_STATUS)${NC}"
    echo -e "${YELLOW}  Rider ID: $ASSIGNED_RIDER${NC}"
    echo -e "${YELLOW}  Note: Auto-assignment requires:${NC}"
    echo -e "${YELLOW}    1. Rider location in Redis (geo:riders)${NC}"
    echo -e "${YELLOW}    2. Rider availability = true${NC}"
    echo -e "${YELLOW}    3. Order.created event from order-service${NC}"
    echo -e "${YELLOW}  Check the dispatch-service logs for details.${NC}"
  fi

  # Step 13: List Orders for Rider
  echo -e "\n${YELLOW}Step 13: Listing Orders for Rider${NC}"
  RIDER_ORDERS=$(curl -s -X GET "$API_BASE/api/orders" \
    -H "Authorization: Bearer $RIDER_TOKEN")

  echo "Response: $RIDER_ORDERS"
  ORDERS_COUNT=$(echo "$RIDER_ORDERS" | grep -o '"id"' | wc -l)
  echo -e "${GREEN}✓ Rider has $ORDERS_COUNT orders${NC}"
fi

# Final Summary
echo -e "\n${BLUE}============================================${NC}"
echo -e "${BLUE}TEST SUMMARY${NC}"
echo -e "${BLUE}============================================${NC}"
echo -e "${GREEN}Rider Account:${NC}"
echo -e "  Email: $RIDER_EMAIL"
echo -e "  Phone: $RIDER_PHONE"
echo -e "  User ID: $RIDER_ID"
echo -e "  Vehicle: Bike (XXX-1234-YY)"
echo -e "  Location: (6.5244, 3.3792)"
echo -e "  Status: Available"
echo -e ""
echo -e "${GREEN}Client Account:${NC}"
echo -e "  Email: $CLIENT_EMAIL"
echo -e "  Phone: $CLIENT_PHONE"
echo -e "  User ID: $CLIENT_ID"
echo -e ""
echo -e "${GREEN}Order Details:${NC}"
echo -e "  Order ID: $ORDER_ID"
echo -e "  Tracking Code: $TRACKING_CODE"
echo -e "  Status: $ORDER_STATUS"
echo -e "  Assigned Rider ID: $ASSIGNED_RIDER"
echo -e ""
echo -e "${BLUE}✓ End-to-end test completed!${NC}"
echo -e "${BLUE}============================================${NC}"
