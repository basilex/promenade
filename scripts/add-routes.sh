#!/usr/bin/env bash

ENTITY=$1
ENTITY_LOWER=$(echo "$ENTITY" | tr '[:upper:]' '[:lower:]')
ENTITY_SNAKE=$(echo "$ENTITY" | sed -E 's/([A-Z])/_\L\1/g' | sed 's/^_//')
ENTITY_PLURAL="${ENTITY_LOWER}s"

ROUTER_FILE="internal/adapter/http/v1/router/router.go"

if [ ! -f "$ROUTER_FILE" ]; then
    echo "Router file not found: $ROUTER_FILE"
    exit 1
fi

echo "[INFO] Adding routes for $ENTITY to $ROUTER_FILE"

# TODO: добавить автоматическую вставку routes
echo "[OK] Routes template generated.  Add manually for now:"
cat << ROUTES

    ${ENTITY_PLURAL} := v1.Group("/${ENTITY_PLURAL}")
    ${ENTITY_PLURAL}.Use(r.authMiddleware.RequireAuth())
    {
        ${ENTITY_PLURAL}.POST("", r.${ENTITY_SNAKE}Handler.Create)
        ${ENTITY_PLURAL}.GET("/:id", r.${ENTITY_SNAKE}Handler.GetByID)
        ${ENTITY_PLURAL}.GET("", r. ${ENTITY_SNAKE}Handler.List)
        ${ENTITY_PLURAL}.PUT("/:id", r.${ENTITY_SNAKE}Handler.Update)
        ${ENTITY_PLURAL}.DELETE("/:id", r.${ENTITY_SNAKE}Handler.Delete)
    }
ROUTES
