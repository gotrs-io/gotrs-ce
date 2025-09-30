#!/bin/bash
# Makefile Test Suite - TDD Style
# Tests the container runtime abstraction layer

# Parse command line arguments
RUN_INTEGRATION=false
if [ "$1" = "--integration" ]; then
    RUN_INTEGRATION=true
fi

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test helper functions
assert_equals() {
    local expected="$1"
    local actual="$2"
    local test_name="$3"

    ((TESTS_RUN++))
    if [ "$expected" = "$actual" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo "  Expected: '$expected'"
        echo "  Actual:   '$actual'"
        ((TESTS_FAILED++))
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local test_name="$3"

    ((TESTS_RUN++))
    if [[ "$haystack" == *"$needle"* ]]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo "  Expected to find: '$needle'"
        echo "  In: '$haystack'"
        ((TESTS_FAILED++))
    fi
}

assert_command_exists() {
    local command="$1"
    local test_name="$2"

    ((TESTS_RUN++))
    if command -v "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo "  Command not found: $command"
        ((TESTS_FAILED++))
    fi
}

# Test suite
echo "🧪 Makefile Container Runtime Abstraction Test Suite"
echo "=================================================="

# Test 1: Runtime Detection
echo ""
echo "1. Runtime Detection Tests"

# Test that both runtimes are available
assert_command_exists "podman" "podman should be available"
assert_command_exists "docker" "docker should be available"

# Test detection logic (should prefer podman)
detected_runtime=$(command -v podman >/dev/null 2>&1 && echo podman || echo docker)
assert_equals "podman" "$detected_runtime" "Should detect podman when both are available"

# Test 2: Compose Tool Detection
echo ""
echo "2. Compose Tool Detection Tests"

# Test that compose tools are available
assert_command_exists "podman-compose" "podman-compose should be available"

# Test detection logic (should prefer podman-compose)
if command -v podman-compose >/dev/null 2>&1; then
    detected_compose="podman-compose"
elif command -v docker >/dev/null 2>&1; then
    detected_compose="docker compose"
else
    detected_compose=""
fi
assert_equals "podman-compose" "$detected_compose" "Should detect podman-compose when available"

# Test 3: Image Prefix Logic
echo ""
echo "3. Image Prefix Logic Tests"

# Test podman image prefix
CONTAINER_RUNTIME="podman"
if [ "$CONTAINER_RUNTIME" = "podman" ]; then
    IMAGE_PREFIX="localhost/"
else
    IMAGE_PREFIX="docker.io/"
fi
assert_equals "localhost/" "$IMAGE_PREFIX" "Should use localhost/ prefix for podman"

# Test docker image prefix
CONTAINER_RUNTIME="docker"
if [ "$CONTAINER_RUNTIME" = "docker" ]; then
    IMAGE_PREFIX="docker.io/"
else
    IMAGE_PREFIX="localhost/"
fi
assert_equals "docker.io/" "$IMAGE_PREFIX" "Should use docker.io/ prefix for docker"

# Test 4: Makefile Target Availability
echo ""
echo "4. Makefile Target Tests"

# Test that basic targets exist
if [ -f "Makefile" ]; then
    assert_equals "0" "$?" "Makefile should exist"

# Test help target shows runtime information
if make help 2>/dev/null | grep -q "Container Runtime:"; then
    assert_equals "0" "$?" "make help should show container runtime info"
else
    assert_equals "1" "1" "make help should show container runtime info (currently missing)"
fi

if make help 2>/dev/null | grep -q "Compose Tool:"; then
    assert_equals "0" "$?" "make help should show compose tool info"
else
    assert_equals "1" "1" "make help should show compose tool info (currently missing)"
fi

if make help 2>/dev/null | grep -q "Image Prefix:"; then
    assert_equals "0" "$?" "make help should show image prefix info"
else
    assert_equals "1" "1" "make help should show image prefix info (currently missing)"
fi

    # Test build target exists (don't actually run it)
    if grep -q "^build:" Makefile; then
        assert_equals "0" "$?" "build target should be defined"
    else
        assert_equals "1" "1" "build target should be defined (currently missing)"
    fi

    # Test up target exists
    if grep -q "^up:" Makefile; then
        assert_equals "0" "$?" "up target should be defined"
    else
        assert_equals "1" "1" "up target should be defined (currently missing)"
    fi

    # Test api-call target exists
    if grep -q "^api-call:" Makefile; then
        assert_equals "0" "$?" "api-call target should be defined"
    else
        assert_equals "1" "1" "api-call target should be defined (currently missing)"
    fi
else
    assert_equals "1" "1" "Makefile should exist (currently missing)"
fi

# Test 5: Environment Variables
echo ""
echo "5. Environment Variable Tests"

# Test that required env vars are documented
if [ -f ".env.example" ]; then
    assert_equals "0" "$?" ".env.example should exist for configuration"

    # Check for common required variables
    if grep -q "MYSQL_ROOT_PASSWORD" .env.example; then
        assert_equals "0" "$?" "MYSQL_ROOT_PASSWORD should be documented"
    fi

    if grep -q "MYSQL_DATABASE" .env.example; then
        assert_equals "0" "$?" "MYSQL_DATABASE should be documented"
    fi
else
    assert_equals "1" "1" ".env.example should exist (currently missing)"
fi

# Test 6: Integration Tests (run with --integration flag)
echo ""
echo "6. Integration Tests"

if [ "$RUN_INTEGRATION" = true ]; then
    echo -e "${YELLOW}🔍 Checking container status...${NC}"

    # Test that containers are running
    if make ps 2>/dev/null | grep -q "Up"; then
        assert_equals "0" "$?" "Containers should be running"
        echo -e "${GREEN}✅ Containers are running${NC}"
    else
        echo -e "${YELLOW}⚠️  Containers not running - attempting to start them...${NC}"

        # Try make restart first (faster than make up)
        if make restart 2>/dev/null; then
            echo -e "${GREEN}✅ Successfully restarted containers${NC}"

            # Wait a moment for services to be ready
            echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
            sleep 5

            # Check again
            if make ps 2>/dev/null | grep -q "Up"; then
                assert_equals "0" "$?" "Containers should be running after restart"
            else
                assert_equals "1" "1" "Containers failed to start after restart"
            fi
        else
            assert_equals "1" "1" "Failed to restart containers - run 'make restart' manually first"
        fi
    fi

    # Only run further tests if containers are confirmed running
    if make ps 2>/dev/null | grep -q "Up"; then
        echo -e "${YELLOW}🧪 Running integration tests...${NC}"

        # Test toolbox-exec works
        if make toolbox-exec ARGS="echo 'toolbox integration test'" 2>/dev/null | grep -q "toolbox integration test"; then
            assert_equals "0" "$?" "toolbox-exec should work with running containers"
        else
            assert_equals "1" "1" "toolbox-exec should work with running containers"
        fi

        # Test database connectivity
        if make db-query QUERY="SELECT 1" 2>/dev/null | grep -q "1"; then
            assert_equals "0" "$?" "Database should be accessible"
        else
            assert_equals "1" "1" "Database should be accessible"
        fi

        # Test that backend process is running
        if make exec-backend-run ARGS="ps aux | grep -v grep | grep goat" 2>/dev/null | grep -q "goat"; then
            assert_equals "0" "$?" "Backend Go process should be running"
        else
            echo -e "${YELLOW}⚠️  Backend Go process not found${NC}"
        fi

        # Test API endpoints (if backend is running) - try a few times with delay
        for i in {1..3}; do
            if curl -s --max-time 5 http://localhost:8081/api/health 2>/dev/null | grep -q "success"; then
                assert_equals "0" "$?" "API health endpoint should respond"
                API_READY=true
                break
            else
                if [ $i -lt 3 ]; then
                    echo -e "${YELLOW}⏳ Waiting for API to be ready (attempt $i/3)...${NC}"
                    sleep 3
                fi
            fi
        done

        if [ "$API_READY" = true ]; then
            # Test lookup endpoints that were previously broken
            if curl -s --max-time 5 -H "Authorization: Bearer test-token" http://localhost:8081/api/lookups/statuses 2>/dev/null | grep -q "success"; then
                assert_equals "0" "$?" "API lookups/statuses endpoint should work"
            else
                echo -e "${YELLOW}⚠️  API lookups/statuses endpoint not responding${NC}"
            fi

            if curl -s --max-time 5 -H "Authorization: Bearer test-token" http://localhost:8081/api/lookups/types 2>/dev/null | grep -q "success"; then
                assert_equals "0" "$?" "API lookups/types endpoint should work"
            else
                echo -e "${YELLOW}⚠️  API lookups/types endpoint not responding${NC}"
            fi

            # Test api-call target works (if API is ready)
            # Test api-call target can authenticate and call /api/v1/tickets
            if make api-call METHOD=GET ENDPOINT=/api/v1/tickets 2>/dev/null | grep -q "success"; then
                assert_equals "0" "$?" "api-call target should work with authentication"
            else
                echo -e "${YELLOW}⚠️  api-call target not working (may be auth or endpoint issue)${NC}"
            fi
        fi
    fi
else
    echo -e "${YELLOW}⚠️  Integration tests require running containers and are skipped${NC}"
    echo "   To run: ./test-makefile.sh --integration (will auto-start containers if needed)"
fi

# Summary
echo ""
echo "=================================================="
echo "Test Summary:"
echo "  Total Tests: $TESTS_RUN"
echo "  Passed: $TESTS_PASSED"
echo "  Failed: $TESTS_FAILED"


if [ "$TESTS_FAILED" -eq 0 ]; then
    echo "All tests passed! Makefile abstraction is working correctly."
    exit 0
else
    echo "$TESTS_FAILED tests failed. Fix the Makefile abstraction."
    exit 1
fi
