#!/bin/bash
set -e

USERNAME="worker_test_$(date +%s)"
PASSWORD="password123"
API_URL="http://localhost:8080/api/v1"

echo "--- 1. Registering user: $USERNAME ---"
curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\"}"
echo ""

echo "--- 2. Logging in ---"
LOGIN_RESP=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\"}")

# Extract Token (using grep/sed fallback if jq missing, but assuming jq for now or python)
# Using python for robustness
TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")

if [ -z "$TOKEN" ] || [ "$TOKEN" == "None" ]; then
  echo "Failed to get token. Response: $LOGIN_RESP"
  exit 1
fi
echo "Token received (len: ${#TOKEN})"

echo "--- 3. Creating Dummy Image ---"
echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==" | base64 -d > test_worker.png

echo "--- 4. Uploading Image ---"
UPLOAD_RESP=$(curl -s -X POST "$API_URL/images" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test_worker.png")

IMAGE_ID=$(echo "$UPLOAD_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")

if [ -z "$IMAGE_ID" ] || [ "$IMAGE_ID" == "None" ]; then
  echo "Failed to upload. Response: $UPLOAD_RESP"
  exit 1
fi
echo "Image ID: $IMAGE_ID"

echo "--- 5. Triggering Async Transform ---"
curl -v -X POST "$API_URL/images/$IMAGE_ID/transform" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"resize": {"width": 100, "height": 100}, "format": "png"}'

echo -e "\n\n--- Done! Check your worker terminal for logs ---"
