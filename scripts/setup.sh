#!/bin/bash
#
# setup.sh
# Set up the development environment for verdagostack.
#
# Installs required tools and configures pre-commit hooks.
#
# Usage:
#   ./scripts/setup.sh
#   make setup

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "Setting up development environment..."
echo ""

# --- Go tools ---

install_go_tool() {
    local name="$1"
    local pkg="$2"
    if command -v "$name" >/dev/null 2>&1; then
        echo -e "${GREEN}  $name${NC} (installed)"
    else
        echo -e "${YELLOW}  Installing $name...${NC}"
        go install "$pkg"
        echo -e "${GREEN}  $name${NC} (installed)"
    fi
}

echo "Go tools:"
install_go_tool golangci-lint github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
install_go_tool goimports     golang.org/x/tools/cmd/goimports@latest
install_go_tool addlicense   github.com/google/addlicense@v1.2.0
install_go_tool govulncheck   golang.org/x/vuln/cmd/govulncheck@latest
echo ""

# --- Security tools (optional) ---

echo "Security tools (optional):"
for tool in gitleaks trivy osv-scanner; do
    if command -v "$tool" >/dev/null 2>&1; then
        echo -e "${GREEN}  $tool${NC} (installed)"
    else
        echo -e "${YELLOW}  $tool${NC} (not installed — brew install $tool)"
    fi
done
echo ""

# --- Pre-commit hooks ---

echo "Pre-commit hooks:"
if command -v pre-commit >/dev/null 2>&1; then
    if [ ! -f .git/hooks/pre-commit ]; then
        pre-commit install >/dev/null 2>&1
        pre-commit install --hook-type commit-msg >/dev/null 2>&1
        echo -e "${GREEN}  Hooks installed${NC}"
    else
        echo -e "${GREEN}  Hooks already installed${NC}"
    fi
else
    echo -e "${YELLOW}  pre-commit not found (install: brew install pre-commit)${NC}"
fi
echo ""

echo "Done!"
