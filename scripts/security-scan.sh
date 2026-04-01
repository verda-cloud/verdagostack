#!/bin/bash
#
# security-scan.sh
# Run security scans locally (same tools as CI workflow)
#
# Tools:
# - gitleaks: Secret scanning
# - trivy: Vulnerability and misconfiguration scanning
# - osv-scanner: Open Source Vulnerability scanning
#
# Usage:
#   ./scripts/security-scan.sh
#   make security-scan

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$REPO_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

RESULT=0

echo -e "${BLUE}====================================================${NC}"
echo -e "${BLUE}  Security Scan${NC}"
echo -e "${BLUE}====================================================${NC}"
echo ""

# Check for required tools
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}  $1 is not installed${NC}"
        return 1
    fi
    return 0
}

echo -e "${YELLOW}Checking tools...${NC}"
MISSING=0
check_tool gitleaks || MISSING=1
check_tool trivy || MISSING=1
check_tool osv-scanner || MISSING=1

if [ $MISSING -eq 1 ]; then
    echo ""
    echo -e "${YELLOW}Install:${NC}"
    echo "  gitleaks:    brew install gitleaks"
    echo "  trivy:       brew install trivy"
    echo "  osv-scanner: brew install osv-scanner"
    exit 1
fi
echo -e "${GREEN}  All tools installed${NC}"
echo ""

# Step 1: Gitleaks
echo -e "${BLUE}Step 1: gitleaks (secret scanning)${NC}"
GITLEAKS_RESULT=0
if [ -f .gitleaks.toml ]; then
    gitleaks detect --config .gitleaks.toml --source . --no-git || GITLEAKS_RESULT=$?
else
    gitleaks detect --source . --no-git || GITLEAKS_RESULT=$?
fi

if [ $GITLEAKS_RESULT -eq 0 ]; then
    echo -e "${GREEN}  No secrets found${NC}"
else
    echo -e "${RED}  Secrets detected!${NC}"
    RESULT=1
fi
echo ""

# Step 2: Trivy
echo -e "${BLUE}Step 2: trivy (vulnerability & misconfiguration)${NC}"
TRIVY_RESULT=0
trivy fs --scanners vuln,misconfig,secret --severity CRITICAL,HIGH --skip-dirs "vendor,.git,.cache" . || TRIVY_RESULT=$?

if [ $TRIVY_RESULT -eq 0 ]; then
    echo -e "${GREEN}  No critical/high vulnerabilities${NC}"
else
    echo -e "${YELLOW}  Vulnerabilities found (review output above)${NC}"
fi
echo ""

# Step 3: OSV Scanner
echo -e "${BLUE}Step 3: osv-scanner (dependency vulnerabilities)${NC}"
OSV_RESULT=0
osv-scanner -r . || OSV_RESULT=$?

if [ $OSV_RESULT -eq 0 ]; then
    echo -e "${GREEN}  No known vulnerabilities${NC}"
else
    echo -e "${YELLOW}  Vulnerabilities found (review output above)${NC}"
fi
echo ""

# Summary
echo -e "${BLUE}====================================================${NC}"
if [ $RESULT -eq 0 ]; then
    echo -e "${GREEN}  Security scan completed${NC}"
else
    echo -e "${RED}  Security issues found${NC}"
    exit 1
fi
echo -e "${BLUE}====================================================${NC}"
