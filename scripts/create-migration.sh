#!/bin/bash
# Create new migration files with proper naming and namespace

set -e

if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Usage: $0 CONTEXT NAME"
    echo "Example: $0 identity add_email_verification"
    exit 1
fi

CONTEXT=$1
NAME=$2
MIGRATIONS_DIR="migrations/${CONTEXT}"

# Create migrations directory if it doesn't exist
mkdir -p "$MIGRATIONS_DIR"

# Find the next migration number
LAST_NUM=$(ls -1 "$MIGRATIONS_DIR" 2>/dev/null | grep -E '^[0-9]+_' | sed 's/_.*//' | sort -n | tail -1)
if [ -z "$LAST_NUM" ]; then
    NEXT_NUM="000001"
else
    NEXT_NUM=$(printf "%06d" $((10#$LAST_NUM + 1)))
fi

# Create migration files
UP_FILE="${MIGRATIONS_DIR}/${NEXT_NUM}_${NAME}.up.sql"
DOWN_FILE="${MIGRATIONS_DIR}/${NEXT_NUM}_${NAME}.down.sql"

# Create UP migration
cat > "$UP_FILE" <<EOF
-- Migration: ${NAME}
-- Context: ${CONTEXT}
-- Created: $(date +"%Y-%m-%d %H:%M:%S")

BEGIN;

-- Add your UP migration SQL here

COMMIT;
EOF

# Create DOWN migration
cat > "$DOWN_FILE" <<EOF
-- Migration: ${NAME}
-- Context: ${CONTEXT}
-- Created: $(date +"%Y-%m-%d %H:%M:%S")

BEGIN;

-- Add your DOWN migration SQL here (reverse of UP)

COMMIT;
EOF

echo "✅ Created migration files:"
echo "   - $UP_FILE"
echo "   - $DOWN_FILE"
