#!/bin/bash
# GOTRS API Testing Script
# Handles authentication and makes API calls with proper TLS

set -euo pipefail

# Load environment variables from .env if it exists
if [[ -f ".env" ]]; then
    # Use a safer method to load .env that handles special characters
    while IFS='=' read -r key value; do
        # Skip empty lines and comments
        [[ -z "$key" || "$key" =~ ^[[:space:]]*# ]] && continue
        # Remove quotes if present
        value=$(echo "$value" | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")
        export "$key=$value"
    done < .env
fi

# Also try to load ADMIN_USER and ADMIN_PASSWORD directly if the above failed
if [[ -f ".env" ]]; then
    ADMIN_USER_FROM_ENV=$(grep '^ADMIN_USER=' .env | cut -d'=' -f2- | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")
    ADMIN_PASSWORD_FROM_ENV=$(grep '^ADMIN_PASSWORD=' .env | cut -d'=' -f2- | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")
    if [[ -n "$ADMIN_USER_FROM_ENV" ]]; then
        export ADMIN_USER="$ADMIN_USER_FROM_ENV"
    fi
    if [[ -n "$ADMIN_PASSWORD_FROM_ENV" ]]; then
        export ADMIN_PASSWORD="$ADMIN_PASSWORD_FROM_ENV"
    fi
fi

echo "DEBUG: BACKEND_URL=$BACKEND_URL"
echo "DEBUG: ADMIN_USER=${ADMIN_USER:-<not set>}"
echo "DEBUG: ADMIN_PASSWORD=${ADMIN_PASSWORD:-<not set>}"
echo "DEBUG: TEST_USERNAME=${TEST_USERNAME:-<not set>}"
echo "DEBUG: TEST_PASSWORD=${TEST_PASSWORD:-<not set>}"

# Default test credentials if not set in .env
# Prefer ADMIN_USER/ADMIN_PASSWORD, fall back to TEST_USERNAME/TEST_PASSWORD, then defaults
TEST_USERNAME="${ADMIN_USER:-${TEST_USERNAME:-root@localhost}}"
TEST_PASSWORD="${ADMIN_PASSWORD:-${TEST_PASSWORD:-admin123}}"
BACKEND_URL="${BACKEND_URL:-https://gotrs-backend:8080}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}ℹ️  $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to authenticate and get JWT token
get_auth_token() {
    echo "DEBUG: Authenticating with $TEST_USERNAME..." >&2

    local response
    echo "DEBUG: Making request to $BACKEND_URL/api/v1/auth/login" >&2
    response=$(curl -s --max-time 30 -X POST "$BACKEND_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"login\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")

    echo "DEBUG: Response: $response" >&2

    # Check if login was successful
    if echo "$response" | jq -e '.success == true' >/dev/null 2>&1; then
        local access_token
        access_token=$(echo "$response" | jq -r '.access_token')
        echo "DEBUG: Authentication successful" >&2
        echo "$access_token"
    else
        echo "DEBUG: Authentication failed. Response: $response" >&2
        return 1
    fi
}

# Function to make API call
make_api_call() {
    local method="$1"
    local endpoint="$2"
    local body="${3:-}"

    echo "DEBUG: Calling get_auth_token" >&2
    local auth_token
    auth_token=$(get_auth_token)

    echo "DEBUG: auth_token=$auth_token" >&2

    echo "DEBUG: Making $method request to $endpoint" >&2

    local curl_cmd=("curl" "-s" "--max-time" "30" "-X" "$method" "$BACKEND_URL$endpoint" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $auth_token")

    if [[ -n "$body" ]]; then
        curl_cmd+=(-d "$body")
    fi

    echo "DEBUG: curl command: ${curl_cmd[*]}" >&2

    local response
    response=$("${curl_cmd[@]}")

    echo "DEBUG: API response: $response" >&2

    # Pretty print JSON response
    if echo "$response" | jq . >/dev/null 2>&1; then
        echo "$response" | jq .
    else
        echo "$response"
    fi
}

# Main script
echo "DEBUG: Starting main script"
if [[ $# -lt 2 ]]; then
    log_error "Usage: $0 <METHOD> <ENDPOINT> [BODY]"
    log_error "Example: $0 GET /api/v1/tickets"
    log_error "Example: $0 POST /api/v1/tickets '{\"title\":\"Test\"}'"
    exit 1
fi

METHOD="$1"
ENDPOINT="$2"
BODY="${3:-}"

echo "DEBUG: METHOD=$METHOD, ENDPOINT=$ENDPOINT, BODY=$BODY"

make_api_call "$METHOD" "$ENDPOINT" "$BODY"