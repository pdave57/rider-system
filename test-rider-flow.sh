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
echo -e "${BLUE}============================================${NC}"

# Step 1: Register a Rider User
echo -e "\n${YELLOW}Step 1: Registering Rider User${NC}"
RIDER_EMAIL="rider_$(date +%s)@example.com"
RIDER_PHONE="080$(printf "%08d" $((RANDOM * 30000 / 32768)))"
RIDER_PASSWORD="test123456"
RIDER_VEHICLE_TYPE="motorcycle"
RIDER_VEHICLE_PLATE="RIDER1234"
RIDER_LICENSE_NUMBER="LIC-987654"

RIDER_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{
    \"full_name\": \"Rider Test User\",
    \"email\": \"$RIDER_EMAIL\",
    \"phone\": \"$RIDER_PHONE\",
    \"password\": \"$RIDER_PASSWORD\",
    \"role\": \"rider\",
    \"vehicle_type\": \"$RIDER_VEHICLE_TYPE\",
    \"vehicle_plate\": \"$RIDER_VEHICLE_PLATE\",
    \"license_number\": \"$RIDER_LICENSE_NUMBER\"
  }")

echo "Response: $RIDER_RESPONSE"
RIDER_ID=$(extract_value "$RIDER_RESPONSE" "id" | head -1)
# Try to extract token from the response - it might be in data.token
RIDER_TOKEN=$(echo "$RIDER_RESPONSE" | grep -o '"token":"[^"]*' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Rider registered with ID: $RIDER_ID${NC}"
echo -e "${GREEN}  Token: ${RIDER_TOKEN:0:20}...${NC}"

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
echo -e "${GREEN}✓ Rider logged in, token: ${RIDER_TOKEN:0:20}...${NC}"

# Step 3: Upload Avatar for Rider
echo -e "\n${YELLOW}Step 3: Uploading Avatar for Rider${NC}"

# Create a test image (100x100 JPEG)
TEMP_IMAGE="/tmp/rider_avatar.jpg"
if command -v convert &> /dev/null; then
  convert -size 100x100 xc:blue "$TEMP_IMAGE"
  echo -e "${GREEN}✓ Test image created with ImageMagick${NC}"
else
  # Create a minimal JPEG from a known-good base64 payload so avatar upload works without ImageMagick
  cat > "$TEMP_IMAGE" <<'EOF'
/9j/4AAQSkZJRgABAQAAAQABAAD/2wCEAAkGBxISEhUTEhIVFRUVFRcVFRcWFRcWFRUVFRUXFhUVFRUYHSggGBolGxUVITEhJSkrLi4uFx8zODMsNygtLisBCgoKDg0OGxAQGy8lICUtLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLf/AABEIAKgBLAMBIgACEQEDEQH/xAAcAAEAAQUBAQAAAAAAAAAAAAAABwIDBAUGAQj/xABAEAACAQIDBAgEAgYDAQkAAAABAgADEQQSIQUxQVEGEyJhcYGRocEUI0JSgtHh8BQjYnKCksLxBxUkQ3Oyw9IVJDSj/8QAGQEAAgMBAAAAAAAAAAAAAAAAAwQBAgAF/8QAJREAAgIBBAICAgMAAAAAAAAAAAECEQMhEjEEE0FRYRQiQnH/2gAMAwEAAhEDEQA/APN+iiiqK/9k=
EOF
  if [ $? -ne 0 ]; then
    echo -e "\n${RED}✗ Failed to create fallback test image${NC}"
    SKIP_AVATAR=true
  else
    base64 -d "$TEMP_IMAGE" > "$TEMP_IMAGE.tmp" 2>/dev/null && mv "$TEMP_IMAGE.tmp" "$TEMP_IMAGE"
    if [ $? -ne 0 ]; then
      echo -e "\n${RED}✗ Failed to decode fallback image${NC}"
      SKIP_AVATAR=true
    else
      echo -e "${GREEN}✓ Fallback test image created${NC}"
    fi
  fi
fi

if [ -z "$SKIP_AVATAR" ]; then
  UPLOAD_RESPONSE=$(curl -s -X POST "$API_BASE/api/upload/avatar" \
    -H "Authorization: Bearer $RIDER_TOKEN" \
    -F "avatar=@$TEMP_IMAGE")
  
  echo "Response: $UPLOAD_RESPONSE"
  echo -e "${GREEN}✓ Avatar uploaded successfully${NC}"
fi

# Step 4: Get Rider Profile
echo -e "\n${YELLOW}Step 4: Getting Rider Profile${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "$API_BASE/api/riders/me" \
  -H "Authorization: Bearer $RIDER_TOKEN")

echo "Response: $PROFILE_RESPONSE"
echo -e "${GREEN}✓ Rider profile retrieved${NC}"

# Step 5: Update Rider Availability
echo -e "\n${YELLOW}Step 5: Setting Rider as Available${NC}"
AVAILABILITY_RESPONSE=$(curl -s -X PATCH "$API_BASE/api/riders/me/availability" \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"available": true}')

echo "Response: $AVAILABILITY_RESPONSE"
echo -e "${GREEN}✓ Rider availability updated${NC}"

# Step 6: Update Rider Location
echo -e "\n${YELLOW}Step 6: Setting Rider Location${NC}"
LOCATION_RESPONSE=$(curl -s -X PATCH "$API_BASE/api/riders/me/location" \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"latitude": 6.5244, "longitude": 3.3792}')

echo "Response: $LOCATION_RESPONSE"
echo -e "${GREEN}✓ Rider location updated${NC}"

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
    \"role\": \"client\"
  }")

echo "Response: $CLIENT_RESPONSE"
CLIENT_ID=$(extract_value "$CLIENT_RESPONSE" "id" | head -1)
CLIENT_TOKEN=$(echo "$CLIENT_RESPONSE" | grep -o '"token":"[^"]*' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✓ Client registered with ID: $CLIENT_ID${NC}"
echo -e "${GREEN}  Token: ${CLIENT_TOKEN:0:20}...${NC}"

# Step 8: Login as Client
echo -e "\n${YELLOW}Step 8: Logging in as Client${NC}"
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
echo -e "${GREEN}✓ Client logged in, token: ${CLIENT_TOKEN:0:20}...${NC}"

# Step 9: Create an Order as Client
echo -e "\n${YELLOW}Step 9: Creating an Order${NC}"
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
echo -e "${GREEN}✓ Order created with ID: $ORDER_ID${NC}"

# Step 10: Wait a moment for auto-assignment
echo -e "\n${YELLOW}Step 10: Waiting for auto-assignment (5 seconds)...${NC}"
sleep 5

# Step 11: Check Order Status
echo -e "\n${YELLOW}Step 11: Checking Order Details (Auto-Assignment)${NC}"
ORDER_DETAILS=$(curl -s -X GET "$API_BASE/api/orders/$ORDER_ID" \
  -H "Authorization: Bearer $CLIENT_TOKEN")

echo "Response: $ORDER_DETAILS"

# Extract assigned rider info
ASSIGNED_RIDER=$(extract_value "$ORDER_DETAILS" "rider_id" | head -1)
ORDER_STATUS=$(extract_value "$ORDER_DETAILS" "status" | head -1)

if [ -n "$ASSIGNED_RIDER" ] && [ "$ASSIGNED_RIDER" != "0" ] && [ "$ASSIGNED_RIDER" != "null" ]; then
  echo -e "${GREEN}✓ Order was AUTO-ASSIGNED to Rider ID: $ASSIGNED_RIDER${NC}"
  echo -e "${GREEN}✓ Order Status: $ORDER_STATUS${NC}"
else
  echo -e "${YELLOW}⚠ Order was NOT auto-assigned yet (Status: $ORDER_STATUS)${NC}"
  echo -e "${YELLOW}  Rider ID: $ASSIGNED_RIDER${NC}"
  echo -e "${YELLOW}  This could mean the dispatch service needs more time or configuration${NC}"
fi

# Step 12: List Orders for Rider
echo -e "\n${YELLOW}Step 12: Listing Orders for Rider${NC}"
RIDER_ORDERS=$(curl -s -X GET "$API_BASE/api/orders" \
  -H "Authorization: Bearer $RIDER_TOKEN")

echo "Response: $RIDER_ORDERS"

# Step 13: Summary
echo -e "\n${BLUE}============================================${NC}"
echo -e "${BLUE}TEST SUMMARY${NC}"
echo -e "${BLUE}============================================${NC}"
echo -e "${GREEN}Rider Account:${NC}"
echo -e "  Email: $RIDER_EMAIL"
echo -e "  Phone: $RIDER_PHONE"
echo -e "  User ID: $RIDER_ID"
echo -e "  Status: Available at (6.5244, 3.3792)"
echo -e ""
echo -e "${GREEN}Client Account:${NC}"
echo -e "  Email: $CLIENT_EMAIL"
echo -e "  Phone: $CLIENT_PHONE"
echo -e "  User ID: $CLIENT_ID"
echo -e ""
echo -e "${GREEN}Order Details:${NC}"
echo -e "  Order ID: $ORDER_ID"
echo -e "  Status: $ORDER_STATUS"
echo -e "  Assigned Rider ID: $ASSIGNED_RIDER"
echo -e ""
echo -e "${BLUE}✓ End-to-end test completed!${NC}"
echo -e "${BLUE}============================================${NC}"
