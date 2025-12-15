#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

echo ""
echo "╔════════════════════════════════════════╗"
echo "║   Interactive Entity Generator         ║"
echo "╚════════════════════════════════════════╝"
echo ""

# Prompt for entity name
read -p "Enter entity name (e.g., Product, Order): " ENTITY_NAME

if [ -z "$ENTITY_NAME" ]; then
    error "Entity name is required"
    exit 1
fi

# Prompt for fields
echo ""
info "Define entity fields (press Enter with empty name to finish):"
echo "Format: field_name field_type (e.g., 'title string', 'price float64', 'quantity int')"
echo ""

FIELDS=()
while true; do
    read -p "Field (name type): " FIELD_INPUT
    if [ -z "$FIELD_INPUT" ]; then
        break
    fi
    FIELDS+=("$FIELD_INPUT")
done

# Prompt for options
echo ""
read -p "Generate migration? (Y/n): " GEN_MIGRATION
GEN_MIGRATION=${GEN_MIGRATION:-Y}

read -p "Generate v1 API? (Y/n): " GEN_V1
GEN_V1=${GEN_V1:-Y}

read -p "Generate v2 API? (y/N): " GEN_V2
GEN_V2=${GEN_V2:-N}

# Build command
CMD="$SCRIPT_DIR/generate.sh entity $ENTITY_NAME"

if [[ !  $GEN_MIGRATION =~ ^[Yy]$ ]]; then
    CMD="$CMD --skip-migration"
fi

if [[ ! $GEN_V1 =~ ^[Yy]$ ]]; then
    CMD="$CMD --skip-handler"
fi

if [[ $GEN_V2 =~ ^[Yy]$ ]]; then
    CMD="$CMD --v2-only"
fi

# Execute
echo ""
info "Executing: $CMD"
echo ""
eval "$CMD"

# Show custom fields reminder
if [ ${#FIELDS[@]} -gt 0 ]; then
    echo ""
    warning "Remember to add these custom fields to your entity:"
    for field in "${FIELDS[@]}"; do
        echo "  - $field"
    done
fi

echo ""
success "Done! 🎉"
