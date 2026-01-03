#!/bin/bash

# Script to check all markdown links in the repository
# Usage: ./scripts/check-links.sh

set -e

PROJECT_ROOT="/Users/basilex/Workspace/src/promenade"
cd "$PROJECT_ROOT"

echo "🔍 Checking all markdown links in the repository..."
echo ""

# Find all markdown files
MD_FILES=$(find . -name "*.md" -not -path "*/node_modules/*" -not -path "*/.vitepress/*" -not -path "*/vendor/*")

BROKEN_LINKS=0
TOTAL_LINKS=0

# Function to check if a file path exists
check_file_link() {
    local file="$1"
    local link="$2"
    local link_file="$3"
    
    # Resolve relative path
    local dir=$(dirname "$file")
    local full_path="$dir/$link"
    
    # Normalize path
    full_path=$(cd "$dir" && cd "$(dirname "$link")" 2>/dev/null && pwd)/$(basename "$link") 2>/dev/null || echo "$full_path"
    
    if [ ! -f "$full_path" ] && [ ! -d "$full_path" ]; then
        echo "❌ BROKEN: $file"
        echo "   Link: $link_file"
        echo "   Expected: $full_path"
        echo ""
        return 1
    fi
    return 0
}

# Check each markdown file
for file in $MD_FILES; do
    # Extract all markdown links [text](path)
    LINKS=$(grep -oE '\[([^\]]+)\]\(([^)]+)\)' "$file" 2>/dev/null || true)
    
    if [ -n "$LINKS" ]; then
        while IFS= read -r link_match; do
            # Extract URL from [text](url)
            LINK=$(echo "$link_match" | sed -E 's/.*\]\(([^)]+)\).*/\1/')
            
            # Skip external links (http://, https://, mailto:, #anchors)
            if [[ "$LINK" =~ ^(https?://|mailto:|#) ]]; then
                continue
            fi
            
            TOTAL_LINKS=$((TOTAL_LINKS + 1))
            
            # Remove anchor from link
            LINK_FILE="${LINK%%#*}"
            
            # Skip empty links
            if [ -z "$LINK_FILE" ]; then
                continue
            fi
            
            # Check if file exists
            if ! check_file_link "$file" "$LINK_FILE" "$LINK"; then
                BROKEN_LINKS=$((BROKEN_LINKS + 1))
            fi
        done <<< "$LINKS"
    fi
done

echo ""
echo "📊 Summary:"
echo "   Total links checked: $TOTAL_LINKS"
echo "   Broken links: $BROKEN_LINKS"
echo ""

if [ $BROKEN_LINKS -eq 0 ]; then
    echo "✅ All links are valid!"
    exit 0
else
    echo "❌ Found $BROKEN_LINKS broken link(s)"
    exit 1
fi
