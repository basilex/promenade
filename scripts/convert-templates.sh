#!/bin/bash

echo "[CONVERT] Converting templates to __VAR__ syntax..."

# Конвертируем {{VAR}} в __VAR__
# Конвертируем @VAR@ в __VAR__ (если уже были конвертированы)

for template in scripts/templates/*.tmpl; do
    if [ -f "$template" ]; then
        echo "Processing: $(basename $template)"
        
        # Создаем backup
        cp "$template" "${template}.bak"
        
        # Конвертируем оба варианта
        cat "${template}. bak" | \
            sed 's/{{ENTITY}}/__ENTITY__/g' | \
            sed 's/{{entity}}/__entity__/g' | \
            sed 's/{{entity_snake}}/__entity_snake__/g' | \
            sed 's/{{entity_plural}}/__entity_plural__/g' | \
            sed 's/{{MODULE}}/__MODULE__/g' | \
            sed 's/@ENTITY@/__ENTITY__/g' | \
            sed 's/@entity@/__entity__/g' | \
            sed 's/@entity_snake@/__entity_snake__/g' | \
            sed 's/@entity_plural@/__entity_plural__/g' | \
            sed 's|@MODULE@|__MODULE__|g' \
            > "$template"
        
        rm "${template}.bak"
        echo "  [OK] Converted"
    fi
done

echo ""
echo "[OK] All templates converted to __VAR__ syntax!"
echo ""
echo "Example:"
echo "  {{ENTITY}} or @ENTITY@ -> __ENTITY__"
echo "  {{entity}} or @entity@ -> __entity__"
