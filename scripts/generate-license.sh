#!/bin/bash

# License Generator Script for Promenade Modules
# Usage: ./scripts/generate-license.sh [module] [tier] [days]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
LICENSE_SECRET="${LICENSE_SECRET:-default-secret-change-in-production}"

# Help message
usage() {
    echo "Usage: $0 [module] [tier] [days]"
    echo ""
    echo "Arguments:"
    echo "  module    Module name (analytics, warehouse, auditlog, etc.)"
    echo "  tier      License tier (BASIC, PRO, ENTERPRISE)"
    echo "  days      Valid for N days from today"
    echo ""
    echo "Example:"
    echo "  $0 analytics PRO 365"
    echo ""
    echo "Environment Variables:"
    echo "  LICENSE_SECRET    Secret key for signing (default: test key)"
    exit 1
}

# Check arguments
if [ $# -ne 3 ]; then
    usage
fi

MODULE=$1
TIER=$2
DAYS=$3

# Validate tier
if [[ ! "$TIER" =~ ^(BASIC|PRO|ENTERPRISE)$ ]]; then
    echo -e "${RED}Error: Invalid tier. Must be BASIC, PRO, or ENTERPRISE${NC}"
    exit 1
fi

# Validate days
if ! [[ "$DAYS" =~ ^[0-9]+$ ]]; then
    echo -e "${RED}Error: Days must be a positive number${NC}"
    exit 1
fi

# Calculate expiry date
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    EXPIRY_DATE=$(date -v+${DAYS}d +%Y%m%d)
else
    # Linux
    EXPIRY_DATE=$(date -d "+${DAYS} days" +%Y%m%d)
fi

echo -e "${YELLOW}Generating license...${NC}"
echo "  Module: $MODULE"
echo "  Tier: $TIER"
echo "  Valid for: $DAYS days"
echo "  Expires: $EXPIRY_DATE"
echo ""

# Build the license generator if not exists
if [ ! -f "./bin/license-generator" ]; then
    echo -e "${YELLOW}Building license generator...${NC}"
    go build -o ./bin/license-generator ./cmd/license-generator/main.go
fi

# Generate license
LICENSE_KEY=$(./bin/license-generator -module="$MODULE" -tier="$TIER" -expiry="$EXPIRY_DATE" -secret="$LICENSE_SECRET")

echo -e "${GREEN}✓ License generated successfully!${NC}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}$LICENSE_KEY${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Set environment variable:"
echo "  export $(echo $MODULE | tr '[:lower:]' '[:upper:]')_LICENSE_KEY=\"$LICENSE_KEY\""
echo ""
echo "Or add to config/modules.yaml:"
echo "  modules:"
echo "    config:"
echo "      $MODULE:"
echo "        license_key: \"$LICENSE_KEY\""
echo ""
