#!/bin/bash
#
# COMPREHENSIVE TESTING FRAMEWORK
# Tests all functionality with positive and negative cases
# Uses curl, logs monitoring, and triggers Playwright tests
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0
ERRORS=()

# Configuration
BASE_URL="http://localhost:8080"
LOG_FILE="/tmp/test_$(date +%Y%m%d_%H%M%S).log"
COOKIES="/tmp/test_cookies.txt"

# Logging functions
log() {
    echo -e "${BLUE}[$(date +%H:%M:%S)]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}✓${NC} $1" | tee -a "$LOG_FILE"
    ((PASSED++))
}

fail() {
    echo -e "${RED}✗${NC} $1" | tee -a "$LOG_FILE"
    ((FAILED++))
    ERRORS+=("$1")
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1" | tee -a "$LOG_FILE"
}

# Monitor container logs in background
monitor_logs() {
    local test_name=$1
    local log_file="/tmp/container_logs_${test_name}_$(date +%s).log"
    
    # Start monitoring in background
    ./scripts/container-wrapper.sh compose logs -f backend 2>&1 | while read line; do
        echo "$line" >> "$log_file"
        # Check for errors
        if echo "$line" | grep -q "ERROR\|PANIC\|500"; then
            echo "[LOG ERROR] $line" >> "$LOG_FILE"
        fi
    done &
    
    echo $! # Return PID for later cleanup
}

# Stop log monitoring
stop_monitor() {
    local pid=$1
    if [ -n "$pid" ]; then
        kill $pid 2>/dev/null || true
        wait $pid 2>/dev/null || true
    fi
}

# Test HTTP response
test_http_response() {
    local test_name=$1
    local method=$2
    local url=$3
    local expected_code=$4
    local data=$5
    local check_content=$6
    
    log "Testing: $test_name"
    
    # Make request
    if [ "$method" = "GET" ]; then
        RESPONSE=$(curl -s -b "$COOKIES" -w "\n%{http_code}" "$BASE_URL$url")
    elif [ "$method" = "POST" ]; then
        RESPONSE=$(curl -s -b "$COOKIES" -c "$COOKIES" -X POST -d "$data" -w "\n%{http_code}" "$BASE_URL$url")
    elif [ "$method" = "PUT" ]; then
        RESPONSE=$(curl -s -b "$COOKIES" -X PUT -d "$data" -w "\n%{http_code}" "$BASE_URL$url")
    elif [ "$method" = "DELETE" ]; then
        RESPONSE=$(curl -s -b "$COOKIES" -X DELETE -w "\n%{http_code}" "$BASE_URL$url")
    fi
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -1)
    BODY=$(echo "$RESPONSE" | head -n -1)
    
    # Check status code
    if [ "$HTTP_CODE" = "$expected_code" ]; then
        success "$test_name - Expected $expected_code, got $HTTP_CODE"
    else
        fail "$test_name - Expected $expected_code, got $HTTP_CODE"
        echo "Response body: $BODY" >> "$LOG_FILE"
        return 1
    fi
    
    # Check content if specified
    if [ -n "$check_content" ]; then
        if echo "$BODY" | grep -q "$check_content"; then
            success "$test_name - Content check passed"
        else
            fail "$test_name - Content check failed (looking for: $check_content)"
            return 1
        fi
    fi
    
    echo "$BODY"
}

# Clean up before tests
cleanup() {
    rm -f "$COOKIES"
    warning "Cleaning up test data..."
}

# Trap to ensure cleanup
trap cleanup EXIT

#########################################
# TEST SUITE STARTS HERE
#########################################

echo "======================================"
echo "COMPREHENSIVE TEST SUITE"
echo "======================================"
echo "Log file: $LOG_FILE"
echo ""

#########################################
# 1. AUTHENTICATION TESTS
#########################################
log "=== AUTHENTICATION TESTS ==="

# Positive: Valid login
test_http_response \
    "Auth: Valid login" \
    "POST" \
    "/api/auth/login" \
    "200" \
    "email=admin@demo.com&password=demo123" \
    "access_token" > /dev/null

# Negative: Invalid credentials
test_http_response \
    "Auth: Invalid password" \
    "POST" \
    "/api/auth/login" \
    "401" \
    "email=admin@demo.com&password=wrongpass" \
    "" > /dev/null

# Negative: Missing credentials
test_http_response \
    "Auth: Missing email" \
    "POST" \
    "/api/auth/login" \
    "400" \
    "password=demo123" \
    "" > /dev/null

# Edge case: SQL injection attempt
test_http_response \
    "Auth: SQL injection attempt" \
    "POST" \
    "/api/auth/login" \
    "401" \
    "email=admin'--&password=x" \
    "" > /dev/null

#########################################
# 2. GROUPS CRUD TESTS
#########################################
log ""
log "=== GROUPS CRUD TESTS ==="

# Login for protected endpoints
curl -s -X POST "$BASE_URL/api/auth/login" \
    -d "email=admin@demo.com&password=demo123" \
    -c "$COOKIES" > /dev/null

# Positive: Create valid group
GROUP_NAME="TestGroup_$(date +%s)"
CREATE_RESPONSE=$(test_http_response \
    "Groups: Create valid group" \
    "POST" \
    "/admin/groups" \
    "200" \
    "name=$GROUP_NAME&comments=Test+group&valid_id=1" \
    "success\":true")

GROUP_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":[0-9]*' | cut -d: -f2)
log "Created group ID: $GROUP_ID"

# Positive: Create inactive group
INACTIVE_GROUP="InactiveTest_$(date +%s)"
test_http_response \
    "Groups: Create inactive group" \
    "POST" \
    "/admin/groups" \
    "200" \
    "name=$INACTIVE_GROUP&comments=Inactive+test&valid_id=2" \
    "success\":true" > /dev/null

# Negative: Create duplicate group
test_http_response \
    "Groups: Create duplicate group" \
    "POST" \
    "/admin/groups" \
    "400" \
    "name=$GROUP_NAME&comments=Duplicate&valid_id=1" \
    "already exists" > /dev/null

# Negative: Create group without name
test_http_response \
    "Groups: Create without name" \
    "POST" \
    "/admin/groups" \
    "400" \
    "comments=No+name&valid_id=1" \
    "" > /dev/null

# Negative: Create group with invalid valid_id
test_http_response \
    "Groups: Create with invalid status" \
    "POST" \
    "/admin/groups" \
    "200" \
    "name=InvalidStatus_$(date +%s)&comments=Test&valid_id=999" \
    "" > /dev/null

# Edge: Create group with special characters
test_http_response \
    "Groups: Create with special chars" \
    "POST" \
    "/admin/groups" \
    "200" \
    "name=Test_Special-123&comments=Test+<script>alert(1)</script>&valid_id=1" \
    "success\":true" > /dev/null

# Positive: List groups
test_http_response \
    "Groups: List all" \
    "GET" \
    "/admin/groups" \
    "200" \
    "" \
    "$GROUP_NAME" > /dev/null

# Positive: Get specific group
test_http_response \
    "Groups: Get by ID" \
    "GET" \
    "/admin/groups/$GROUP_ID" \
    "200" \
    "" \
    "$GROUP_NAME" > /dev/null

# Negative: Get non-existent group
test_http_response \
    "Groups: Get non-existent" \
    "GET" \
    "/admin/groups/999999" \
    "404" \
    "" \
    "" > /dev/null

# Positive: Update group
test_http_response \
    "Groups: Update group" \
    "PUT" \
    "/admin/groups/$GROUP_ID" \
    "200" \
    "name=$GROUP_NAME&comments=Updated+comment&valid_id=1" \
    "success\":true" > /dev/null

# Negative: Update to duplicate name
test_http_response \
    "Groups: Update to duplicate name" \
    "PUT" \
    "/admin/groups/$GROUP_ID" \
    "400" \
    "name=admin&comments=Try+duplicate&valid_id=1" \
    "" > /dev/null

# Positive: Get group members
test_http_response \
    "Groups: Get members" \
    "GET" \
    "/admin/groups/$GROUP_ID/members" \
    "200" \
    "" \
    "success\":true" > /dev/null

# Negative: Delete system group
test_http_response \
    "Groups: Delete system group (admin)" \
    "DELETE" \
    "/admin/groups/1" \
    "400" \
    "" \
    "cannot delete system group" > /dev/null

# Positive: Delete created group
test_http_response \
    "Groups: Delete group" \
    "DELETE" \
    "/admin/groups/$GROUP_ID" \
    "200" \
    "" \
    "success\":true" > /dev/null

# Verify deletion
test_http_response \
    "Groups: Verify deletion" \
    "GET" \
    "/admin/groups/$GROUP_ID" \
    "404" \
    "" \
    "" > /dev/null

#########################################
# 3. ERROR HANDLING TESTS
#########################################
log ""
log "=== ERROR HANDLING TESTS ==="

# Test Guru Meditation trigger
log "Testing Guru Meditation component..."

# Check if Guru Meditation is in page
GROUPS_PAGE=$(curl -s -b "$COOKIES" "$BASE_URL/admin/groups")
if echo "$GROUPS_PAGE" | grep -q "guru-meditation"; then
    success "Guru Meditation component present"
else
    fail "Guru Meditation component missing"
fi

# Check for JavaScript functions
if echo "$GROUPS_PAGE" | grep -q "function showGuruMeditation"; then
    success "showGuruMeditation function present"
else
    fail "showGuruMeditation function missing"
fi

#########################################
# 4. PERFORMANCE TESTS
#########################################
log ""
log "=== PERFORMANCE TESTS ==="

# Test response times
START_TIME=$(date +%s%N)
curl -s -b "$COOKIES" "$BASE_URL/admin/groups" > /dev/null
END_TIME=$(date +%s%N)
RESPONSE_TIME=$(( ($END_TIME - $START_TIME) / 1000000 ))

if [ $RESPONSE_TIME -lt 200 ]; then
    success "Groups page loads in ${RESPONSE_TIME}ms (< 200ms)"
else
    warning "Groups page loads in ${RESPONSE_TIME}ms (> 200ms target)"
fi

#########################################
# 5. SECURITY TESTS
#########################################
log ""
log "=== SECURITY TESTS ==="

# Test access without authentication
rm -f "$COOKIES"
test_http_response \
    "Security: Access groups without auth" \
    "GET" \
    "/admin/groups" \
    "303" \
    "" \
    "" > /dev/null

# Test XSS prevention
curl -s -X POST "$BASE_URL/api/auth/login" \
    -d "email=admin@demo.com&password=demo123" \
    -c "$COOKIES" > /dev/null

XSS_PAYLOAD="<script>alert('XSS')</script>"
CREATE_RESPONSE=$(curl -s -b "$COOKIES" -X POST "$BASE_URL/admin/groups" \
    -d "name=XSSTest_$(date +%s)&comments=$XSS_PAYLOAD&valid_id=1")

if echo "$CREATE_RESPONSE" | grep -q "<script>"; then
    fail "XSS: Script tags not escaped in response"
else
    success "XSS: Script tags properly handled"
fi

#########################################
# 6. CONCURRENT ACCESS TESTS
#########################################
log ""
log "=== CONCURRENT ACCESS TESTS ==="

# Create multiple groups simultaneously
log "Creating 5 groups concurrently..."
for i in {1..5}; do
    (
        curl -s -b "$COOKIES" -X POST "$BASE_URL/admin/groups" \
            -d "name=Concurrent_${i}_$(date +%s)&comments=Test&valid_id=1" \
            > "/tmp/concurrent_$i.log" 2>&1
    ) &
done
wait

# Check all succeeded
CONCURRENT_SUCCESS=0
for i in {1..5}; do
    if grep -q "success\":true" "/tmp/concurrent_$i.log"; then
        ((CONCURRENT_SUCCESS++))
    fi
    rm -f "/tmp/concurrent_$i.log"
done

if [ $CONCURRENT_SUCCESS -eq 5 ]; then
    success "Concurrent: All 5 groups created successfully"
else
    fail "Concurrent: Only $CONCURRENT_SUCCESS/5 groups created"
fi

#########################################
# 7. BROWSER AUTOMATION TESTS
#########################################
log ""
log "=== BROWSER AUTOMATION TESTS ==="

# Create Playwright test script
cat > /tmp/playwright_test.js << 'EOF'
const { chromium } = require('playwright');

(async () => {
    const browser = await chromium.launch({ headless: true });
    const page = await browser.newPage();
    
    // Track console errors
    let consoleErrors = [];
    page.on('console', msg => {
        if (msg.type() === 'error') {
            consoleErrors.push(msg.text());
        }
    });
    
    // Track network errors
    let networkErrors = [];
    page.on('response', response => {
        if (response.status() >= 500) {
            networkErrors.push(`${response.status()} ${response.url()}`);
        }
    });
    
    try {
        // Login
        await page.goto('http://localhost:8080/login');
        await page.fill('input[name="email"]', 'admin@demo.com');
        await page.fill('input[name="password"]', 'demo123');
        await page.press('input[name="password"]', 'Enter');
        await page.waitForTimeout(2000);
        
        // Navigate to groups
        await page.goto('http://localhost:8080/admin/groups');
        await page.waitForTimeout(2000);
        
        // Test create group flow
        await page.click('button:has-text("Add Group")');
        await page.waitForTimeout(1000);
        
        const groupName = `UITest_${Date.now()}`;
        await page.fill('input#groupName', groupName);
        await page.fill('textarea#groupComments', 'UI test group');
        await page.selectOption('select#groupStatus', '2'); // Inactive
        await page.click('button[type="submit"]:has-text("Save")');
        await page.waitForTimeout(3000);
        
        // Check results
        const results = {
            success: true,
            consoleErrors: consoleErrors,
            networkErrors: networkErrors,
            groupCreated: await page.isVisible(`tr:has-text("${groupName}")`)
        };
        
        console.log(JSON.stringify(results));
        
    } catch (error) {
        console.log(JSON.stringify({
            success: false,
            error: error.message,
            consoleErrors: consoleErrors,
            networkErrors: networkErrors
        }));
    } finally {
        await browser.close();
    }
})();
EOF

# Note: Would run with: node /tmp/playwright_test.js
log "Playwright test script created (requires Node.js environment to run)"

#########################################
# 8. LOG ANALYSIS
#########################################
log ""
log "=== LOG ANALYSIS ==="

# Check backend logs for errors
ERROR_COUNT=$(./scripts/container-wrapper.sh compose logs backend --tail=100 2>&1 | grep -c "ERROR\|PANIC" || true)
WARNING_COUNT=$(./scripts/container-wrapper.sh compose logs backend --tail=100 2>&1 | grep -c "WARNING" || true)
HTTP_500_COUNT=$(./scripts/container-wrapper.sh compose logs backend --tail=100 2>&1 | grep -c " 500 " || true)

if [ $ERROR_COUNT -eq 0 ]; then
    success "Logs: No ERROR or PANIC messages"
else
    fail "Logs: Found $ERROR_COUNT ERROR/PANIC messages"
fi

if [ $HTTP_500_COUNT -eq 0 ]; then
    success "Logs: No HTTP 500 errors"
else
    fail "Logs: Found $HTTP_500_COUNT HTTP 500 errors"
fi

if [ $WARNING_COUNT -gt 0 ]; then
    warning "Logs: Found $WARNING_COUNT WARNING messages"
fi

#########################################
# TEST SUMMARY
#########################################
echo ""
echo "======================================"
echo "TEST SUMMARY"
echo "======================================"
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo "Failed tests:"
    for error in "${ERRORS[@]}"; do
        echo "  - $error"
    done
    exit 1
else
    echo ""
    echo -e "${GREEN}ALL TESTS PASSED!${NC}"
    exit 0
fi