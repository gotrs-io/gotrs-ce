#!/bin/bash
# Legal Compliance Check for OTRS Materials
# Ensures all OTRS analysis materials stay in local/ directory

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Checking OTRS legal compliance...${NC}"

# Function to check if we're in a git repository
check_git_repo() {
    if [ ! -d ".git" ]; then
        echo -e "${RED}Error: Not in a git repository root${NC}"
        exit 1
    fi
}

# Function to check if local/ directory exists and is properly gitignored
check_local_directory() {
    echo -e "${BLUE}📁 Checking local/ directory setup...${NC}"
    
    # Check if .gitignore exists and contains local/
    if [ ! -f ".gitignore" ]; then
        echo -e "${YELLOW}⚠️  Warning: No .gitignore file found${NC}"
        return 1
    fi
    
    if ! grep -q "^local/" ".gitignore"; then
        echo -e "${YELLOW}⚠️  Warning: local/ directory not found in .gitignore${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ local/ directory properly configured in .gitignore${NC}"
    return 0
}

# Function to check for any local/ files in git
check_git_tracking() {
    echo -e "${BLUE}🔎 Checking if any local/ files are tracked by git...${NC}"
    
    LOCAL_TRACKED=$(git ls-files | grep "^local/" | wc -l)
    if [ "$LOCAL_TRACKED" -gt 0 ]; then
        echo -e "${RED}❌ ERROR: Found local/ files tracked by git:${NC}"
        git ls-files | grep "^local/" | sed 's/^/   - /'
        echo -e "${YELLOW}💡 These files should be removed from git tracking:${NC}"
        echo -e "   git rm --cached <file>"
        echo -e "   git commit -m \"Remove local/ files for legal compliance\""
        return 1
    fi
    
    echo -e "${GREEN}✅ No local/ files are tracked by git${NC}"
    return 0
}

# Function to check for any local/ files staged for commit
check_staged_files() {
    echo -e "${BLUE}📋 Checking staged files for local/ directory content...${NC}"
    
    if ! git rev-parse --verify HEAD >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Initial commit - no staged file check needed${NC}"
        return 0
    fi
    
    LOCAL_STAGED=$(git diff --cached --name-only | grep "^local/" | wc -l)
    if [ "$LOCAL_STAGED" -gt 0 ]; then
        echo -e "${RED}❌ ERROR: Found local/ files staged for commit:${NC}"
        git diff --cached --name-only | grep "^local/" | sed 's/^/   - /'
        echo -e "${YELLOW}💡 Unstage these files:${NC}"
        echo -e "   git reset HEAD <file>"
        return 1
    fi
    
    echo -e "${GREEN}✅ No local/ files staged for commit${NC}"
    return 0
}

# Function to provide usage guidance
show_usage() {
    echo -e "${BLUE}📖 OTRS Legal Compliance Guidelines:${NC}"
    echo -e ""
    echo -e "${YELLOW}✅ DO:${NC}"
    echo -e "   - Store all OTRS screenshots in local/screenshots/"
    echo -e "   - Keep analysis notes in local/analysis/"
    echo -e "   - Save reference materials in local/reference/"
    echo -e "   - Generate specification documents from analysis"
    echo -e ""
    echo -e "${YELLOW}❌ DON'T:${NC}"
    echo -e "   - Commit any files from local/ directory"
    echo -e "   - Store OTRS copyrighted materials outside local/"
    echo -e "   - Include OTRS screenshots in git repository"
    echo -e ""
    echo -e "${YELLOW}📁 Recommended local/ structure:${NC}"
    echo -e "   local/"
    echo -e "   ├── screenshots/          # OTRS instance screenshots"
    echo -e "   ├── analysis/            # Analysis notes and findings"
    echo -e "   ├── reference/           # Reference documentation"
    echo -e "   └── specifications/      # Generated specs (can be copied out)"
    echo -e ""
}

# Main execution
main() {
    check_git_repo
    
    local errors=0
    
    check_local_directory || errors=$((errors + 1))
    check_git_tracking || errors=$((errors + 1))
    check_staged_files || errors=$((errors + 1))
    
    if [ $errors -eq 0 ]; then
        echo -e ""
        echo -e "${GREEN}🎉 Legal compliance check passed!${NC}"
        echo -e "${GREEN}✅ All OTRS materials properly contained in local/ directory${NC}"
    else
        echo -e ""
        echo -e "${RED}❌ Legal compliance issues found ($errors error(s))${NC}"
        echo -e "${YELLOW}Please resolve the issues above before proceeding${NC}"
        show_usage
        exit 1
    fi
    
    # Show usage even on success for reference
    echo -e ""
    show_usage
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        show_usage
        exit 0
        ;;
    *)
        main
        ;;
esac