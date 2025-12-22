#!/bin/bash
set -e

# Script to reorganize flat migrations into namespace-based structure
# migrations/ → migrations/{core,posts,profiles}/

MIGRATIONS_DIR="migrations"
BACKUP_DIR="migrations.backup"

echo "🔄 Reorganizing migrations into namespace structure..."

# Create backup
if [ -d "$BACKUP_DIR" ]; then
    echo "  Backup directory already exists, removing..."
    rm -rf "$BACKUP_DIR"
fi
cp -r "$MIGRATIONS_DIR" "$BACKUP_DIR"
echo " Created backup in $BACKUP_DIR"

# Define migration mapping
# Format: "original_number:namespace:new_number:description"
declare -a MAPPING=(
    "000001:core:000001:init_schema_deps"
    "000002:core:000002:create_auth_schema"
    "000009:core:000003:create_rbac_tables"
    "000014:core:000004:create_timezones_table"
    "000015:core:000005:create_languages_table"
    "000003:core:000006:create_countries_currencies"
    "000004:profiles:000001:create_user_contacts"
    "000005:profiles:000002:create_user_profiles"
    "000006:posts:000001:create_user_posts"
    "000007:posts:000002:create_post_comments"
    "000008:posts:000003:create_comment_likes_table"
)

# Create namespace directories
mkdir -p "$MIGRATIONS_DIR/core"
mkdir -p "$MIGRATIONS_DIR/posts"
mkdir -p "$MIGRATIONS_DIR/profiles"
echo " Created namespace directories"

# Move and rename migrations
for entry in "${MAPPING[@]}"; do
    IFS=':' read -r old_num namespace new_num desc <<< "$entry"
    
    # Find files matching the pattern
    up_file="${MIGRATIONS_DIR}/${old_num}_${desc}.up.sql"
    down_file="${MIGRATIONS_DIR}/${old_num}_${desc}.down.sql"
    
    if [ -f "$up_file" ] && [ -f "$down_file" ]; then
        # Move with new numbering
        new_up="${MIGRATIONS_DIR}/${namespace}/${new_num}_${desc}.up.sql"
        new_down="${MIGRATIONS_DIR}/${namespace}/${new_num}_${desc}.down.sql"
        
        mv "$up_file" "$new_up"
        mv "$down_file" "$new_down"
        
        echo "   Moved ${old_num}_${desc} → ${namespace}/${new_num}_${desc}"
    else
        echo "    Files not found: ${old_num}_${desc}"
    fi
done

# Check for any remaining files
remaining=$(find "$MIGRATIONS_DIR" -maxdepth 1 -name "*.sql" 2>/dev/null || true)
if [ -n "$remaining" ]; then
    echo ""
    echo "  Warning: Some migration files were not moved:"
    echo "$remaining"
    echo ""
    echo "Please review and manually organize these files."
else
    echo ""
    echo " All migrations successfully reorganized!"
fi

# Show final structure
echo ""
echo "📁 New structure:"
echo "================"
for ns in core posts profiles; do
    count=$(ls -1 "$MIGRATIONS_DIR/$ns"/*.sql 2>/dev/null | wc -l || echo 0)
    echo "  $ns: $((count / 2)) migrations"
    ls -1 "$MIGRATIONS_DIR/$ns"/*.up.sql 2>/dev/null | sed 's|.*/||' | sed 's/.up.sql//' | sed 's/^/    - /' || true
done

echo ""
echo " Done! Backup saved in: $BACKUP_DIR"
echo ""
echo "Next steps:"
echo "  1. Review the new structure in migrations/{core,posts,profiles}/"
echo "  2. Update cmd/api/main.go to use migration.Manager"
echo "  3. Test migrations: ./bin/migrate -cmd=status"
echo "  4. If everything works, remove backup: rm -rf $BACKUP_DIR"
