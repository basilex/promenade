#!/bin/bash

# Create a new migration file for a namespace
# Usage: ./scripts/create-migration.sh <namespace> <name>

set -e

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <namespace> <name>"
    echo "Example: $0 posts add_post_views"
    exit 1
fi

NAMESPACE=$1
NAME=$2
MIGRATIONS_DIR="migrations/$NAMESPACE"

# Create namespace directory if it doesn't exist
mkdir -p "$MIGRATIONS_DIR"

# Get next version number
NEXT_VERSION=1
if [ -d "$MIGRATIONS_DIR" ]; then
    # Find highest version number
    LAST_FILE=$(ls "$MIGRATIONS_DIR" 2>/dev/null | grep "^[0-9]" | sort -V | tail -n 1)
    if [ -n "$LAST_FILE" ]; then
        LAST_VERSION=$(echo "$LAST_FILE" | cut -d'_' -f1 | sed 's/^0*//')
        NEXT_VERSION=$((LAST_VERSION + 1))
    fi
fi

# Format version with leading zeros (6 digits)
VERSION=$(printf "%06d" "$NEXT_VERSION")

# Create migration files
UP_FILE="$MIGRATIONS_DIR/${VERSION}_${NAME}.up.sql"
DOWN_FILE="$MIGRATIONS_DIR/${VERSION}_${NAME}.down.sql"

# Create UP migration template
cat > "$UP_FILE" << EOF
-- Migration: ${NAME}
-- Namespace: ${NAMESPACE}
-- Version: ${VERSION}
-- Created: $(date '+%Y-%m-%d %H:%M:%S')

-- Add your UP migration here

EOF

# Create DOWN migration template
cat > "$DOWN_FILE" << EOF
-- Migration: ${NAME} (rollback)
-- Namespace: ${NAMESPACE}
-- Version: ${VERSION}
-- Created: $(date '+%Y-%m-%d %H:%M:%S')

-- Add your DOWN migration here (must undo the UP migration)

EOF

echo " Created migration files:"
echo "  → $UP_FILE"
echo "  → $DOWN_FILE"
echo ""
echo "Next steps:"
echo "  1. Edit the migration files to add your SQL"
echo "  2. Run: make migrate-module MODULE=$NAMESPACE"
