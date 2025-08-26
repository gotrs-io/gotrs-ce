#!/bin/bash
# Install git hooks for GOTRS project
# This script sets up pre-commit hooks without requiring Python/pip

set -e

HOOKS_DIR=".git/hooks"
PRE_COMMIT_HOOK="$HOOKS_DIR/pre-commit"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Installing GOTRS git hooks..."

# Ensure we're in a git repository
if [ ! -d ".git" ]; then
    echo -e "${RED}Error: Not in a git repository root${NC}"
    exit 1
fi

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Create pre-commit hook
cat > "$PRE_COMMIT_HOOK" << 'EOF'
#!/bin/bash
# GOTRS Pre-commit Hook - Secret Scanning
# This hook runs Gitleaks to scan for secrets before each commit

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Running pre-commit security checks...${NC}"

# Determine container runtime (podman or docker)
if command -v podman &> /dev/null; then
    CONTAINER_CMD="podman"
elif command -v docker &> /dev/null; then
    CONTAINER_CMD="docker"
else
    echo -e "${YELLOW}⚠️  Warning: Neither docker nor podman found. Skipping secret scan.${NC}"
    echo -e "${YELLOW}   Install Docker or Podman to enable secret scanning.${NC}"
    exit 0
fi

# Check if gitleaks.toml exists for custom configuration
CONFIG_ARG=""
if [ -f "gitleaks.toml" ]; then
    CONFIG_ARG="--config gitleaks.toml"
fi

# Run Gitleaks on staged files
echo -e "${BLUE}🔐 Scanning staged files for secrets...${NC}"

$CONTAINER_CMD run --rm \
    -v "$(pwd):/workspace" \
    -w /workspace \
    zricethezav/gitleaks:latest \
    protect --staged $CONFIG_ARG --verbose 2>&1

GITLEAKS_EXIT=$?

if [ $GITLEAKS_EXIT -ne 0 ]; then
    echo -e "${RED}❌ Secrets detected in staged files!${NC}"
    echo -e "${YELLOW}📋 Review the output above to identify and remove secrets${NC}"
    echo -e "${YELLOW}💡 Tips:${NC}"
    echo -e "   - Move secrets to environment variables"
    echo -e "   - Use .env files (never commit them)"
    echo -e "   - If it's a false positive, update gitleaks.toml"
    echo -e "${YELLOW}⚠️  To bypass this check (NOT RECOMMENDED):${NC}"
    echo -e "   git commit --no-verify"
    exit 1
fi

# Optional: Check for large files
LARGE_FILES=$(git diff --cached --name-only | xargs -I {} sh -c 'test -f "{}" && stat -f%z "{}" 2>/dev/null || stat -c%s "{}" 2>/dev/null' | awk '$1 > 5242880 {print}' | wc -l)
if [ "$LARGE_FILES" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Warning: Large files (>5MB) detected in commit${NC}"
    echo -e "   Consider using Git LFS for large files"
fi

# OTRS Legal Compliance: Block local/ directory
LOCAL_FILES=$(git diff --cached --name-only | grep "^local/" | wc -l)
if [ "$LOCAL_FILES" -gt 0 ]; then
    echo -e "${RED}❌ BLOCKED: Files from local/ directory detected!${NC}"
    echo -e "${YELLOW}📋 The following files are blocked for legal compliance:${NC}"
    git diff --cached --name-only | grep "^local/" | sed 's/^/   - /'
    echo -e "${YELLOW}💡 Keep all OTRS analysis materials in local/ directory${NC}"
    echo -e "${YELLOW}   This directory is gitignored for legal compliance${NC}"
    exit 1
fi

# Optional: Check for common sensitive file patterns
SENSITIVE_PATTERNS=".env .pem .key .p12 .pfx id_rsa id_dsa"
for pattern in $SENSITIVE_PATTERNS; do
    if git diff --cached --name-only | grep -q "$pattern"; then
        echo -e "${YELLOW}⚠️  Warning: Potentially sensitive file pattern detected: $pattern${NC}"
        echo -e "   Ensure this file should be committed"
    fi
done

echo -e "${GREEN}✅ Pre-commit checks passed!${NC}"
exit 0
EOF

# Make the hook executable
chmod +x "$PRE_COMMIT_HOOK"

echo -e "${GREEN}✅ Git hooks installed successfully!${NC}"
echo ""
echo "Installed hooks:"
echo "  - pre-commit: Scans for secrets using Gitleaks"
echo ""
echo "The hook will run automatically before each commit."
echo "To bypass (use with caution): git commit --no-verify"
echo ""
echo "To test the hook manually:"
echo "  $PRE_COMMIT_HOOK"