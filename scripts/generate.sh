#!/usr/bin/env bash

set -e

# Load helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

# Usage
usage() {
    cat << USAGE
Usage: ./scripts/generate.sh <command> <entity-name> [options]

Commands:
    entity      Generate complete entity with all layers
    migration   Generate only database migration
    handler     Generate only HTTP handler (v1 and v2)
    dto         Generate only DTOs
    usecase     Generate only use case
    repository  Generate only repository

Options:
    --skip-migration    Skip migration generation
    --skip-handler      Skip handler generation
    --v1-only           Generate only v1 endpoints
    --v2-only           Generate only v2 endpoints

Examples:
    ./scripts/generate.sh entity Product
    ./scripts/generate.sh entity Order --skip-migration
    ./scripts/generate.sh handler Product --v1-only
    ./scripts/generate.sh migration Product

USAGE
    exit 1
}

# Check arguments
if [ $# -lt 2 ]; then
    usage
fi

COMMAND=$1
ENTITY_NAME=$2
shift 2

# Parse options
SKIP_MIGRATION=false
SKIP_HANDLER=false
V1_ONLY=false
V2_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-migration)
            SKIP_MIGRATION=true
            shift
            ;;
        --skip-handler)
            SKIP_HANDLER=true
            shift
            ;;
        --v1-only)
            V1_ONLY=true
            shift
            ;;
        --v2-only)
            V2_ONLY=true
            shift
            ;;
        *)
            error "Unknown option: $1"
            usage
            ;;
    esac
done

# Calculate variations of entity name
ENTITY_LOWER=$(to_lower "$ENTITY_NAME")
ENTITY_SNAKE=$(to_snake_case "$ENTITY_NAME")
ENTITY_PLURAL=$(to_plural "$ENTITY_LOWER")
MODULE_NAME=$(get_module_name)

info "Generating files for entity: $ENTITY_NAME"
info "Module: $MODULE_NAME"
info "Snake case: $ENTITY_SNAKE"
info "Plural: $ENTITY_PLURAL"
echo ""

# Generate Entity
generate_entity() {
    info "Generating entity..."
    
    local output_dir="internal/domain/entity"
    ensure_dir "$output_dir"
    
    local output_file="$output_dir/${ENTITY_SNAKE}.go"
    
    if [ -f "$output_file" ]; then
        warning "Entity file already exists: $output_file"
        read -p "Overwrite? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            return
        fi
    fi
    
    create_from_template \
        "$SCRIPT_DIR/templates/entity.tmpl" \
        "$output_file" \
        "$ENTITY_NAME" \
        "$ENTITY_LOWER" \
        "$ENTITY_SNAKE" \
        "$ENTITY_PLURAL" \
        "$MODULE_NAME"
    
    # Add error to errors.go
    local errors_file="$output_dir/errors.go"
    if [ -f "$errors_file" ]; then
        local error_line="    Err${ENTITY_NAME}NotFound = errors.New(\"${ENTITY_LOWER} not found\")"
        
        if !  grep -q "Err${ENTITY_NAME}NotFound" "$errors_file"; then
            add_line_after_pattern "$errors_file" "var (" "$error_line"
            success "Added error to errors.go"
        fi
    else
        # Создаем errors.go если его нет
        cat > "$errors_file" << ERRORFILE
package entity

import "errors"

var (
    ErrUserNotFound       = errors.New("user not found")
    ErrUserAlreadyExists  = errors.New("user already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrInvalidInput       = errors.New("invalid input")
    ErrUnauthorized       = errors.New("unauthorized")
    Err${ENTITY_NAME}NotFound = errors.New("${ENTITY_LOWER} not found")
)
ERRORFILE
        success "Created errors.go"
    fi
}

# Generate Repository Interface
generate_repository_interface() {
    info "Generating repository interface..."
    
    local output_dir="internal/domain/repository"
    ensure_dir "$output_dir"
    
    local output_file="$output_dir/${ENTITY_SNAKE}_repository.go"
    
    create_from_template \
        "$SCRIPT_DIR/templates/repository_interface.tmpl" \
        "$output_file" \
        "$ENTITY_NAME" \
        "$ENTITY_LOWER" \
        "$ENTITY_SNAKE" \
        "$ENTITY_PLURAL" \
        "$MODULE_NAME"
}

# Generate Repository Implementation
generate_repository_impl() {
    info "Generating repository implementation..."
    
    local output_dir="internal/adapter/repository/postgres"
    ensure_dir "$output_dir"
    
    local output_file="$output_dir/${ENTITY_SNAKE}_repository.go"
    
    create_from_template \
        "$SCRIPT_DIR/templates/repository_impl.tmpl" \
        "$output_file" \
        "$ENTITY_NAME" \
        "$ENTITY_LOWER" \
        "$ENTITY_SNAKE" \
        "$ENTITY_PLURAL" \
        "$MODULE_NAME"
}

# Generate UseCase
generate_usecase() {
    info "Generating use case..."
    
    local output_dir="internal/usecase"
    ensure_dir "$output_dir"
    
    local output_file="$output_dir/${ENTITY_SNAKE}_usecase.go"
    
    create_from_template \
        "$SCRIPT_DIR/templates/usecase.tmpl" \
        "$output_file" \
        "$ENTITY_NAME" \
        "$ENTITY_LOWER" \
        "$ENTITY_SNAKE" \
        "$ENTITY_PLURAL" \
        "$MODULE_NAME"
}

# Generate DTO v1
generate_dto_v1() {
    info "Generating DTO v1..."
    
    local output_dir="internal/adapter/http/v1/dto"
    ensure_dir "$output_dir"
    
    local output_file="$output_dir/${ENTITY_SNAKE}_dto.go"
    
    create_from_template \
        "$SCRIPT_DIR/templates/dto_v1.tmpl" \
        "$output_file" \
        "$ENTITY_NAME" \
        "$ENTITY_LOWER" \
        "$ENTITY_SNAKE" \
        "$ENTITY_PLURAL" \
        "$MODULE_NAME"
}

# Generate Handler v1
generate_handler_v1() {
    info "Generating handler v1..."
    
    local output_dir="internal/adapter/http/v1/handler"
    ensure_dir "$output_dir"
    
    local output_file="$output_dir/${ENTITY_SNAKE}_handler.go"
    
    create_from_template \
        "$SCRIPT_DIR/templates/handler_v1.tmpl" \
        "$output_file" \
        "$ENTITY_NAME" \
        "$ENTITY_LOWER" \
        "$ENTITY_SNAKE" \
        "$ENTITY_PLURAL" \
        "$MODULE_NAME"
}

# Generate Migration
generate_migration() {
    info "Generating migration..."
    
    local migration_name="create_${ENTITY_SNAKE}s_table"
    
    # Find next migration number
    local last_migration=$(ls migrations/*.up.sql 2>/dev/null | tail -1 | sed 's/[^0-9]//g' || echo "0")
    local next_number=$(printf "%06d" $((10#$last_migration + 1)))
    
    local up_file="migrations/${next_number}_${migration_name}.up.sql"
    local down_file="migrations/${next_number}_${migration_name}.down.sql"
    
    create_from_template \
        "$SCRIPT_DIR/templates/migration_up.tmpl" \
        "$up_file" \
        "$ENTITY_NAME" \
        "$ENTITY_LOWER" \
        "$ENTITY_SNAKE" \
        "$ENTITY_PLURAL" \
        "$MODULE_NAME"
    
    create_from_template \
        "$SCRIPT_DIR/templates/migration_down.tmpl" \
        "$down_file" \
        "$ENTITY_NAME" \
        "$ENTITY_LOWER" \
        "$ENTITY_SNAKE" \
        "$ENTITY_PLURAL" \
        "$MODULE_NAME"
}

# Generate complete entity
generate_complete_entity() {
    echo ""
    success "========================================="
    success "  Generating Complete Entity:  $ENTITY_NAME"
    success "========================================="
    echo ""
    
    generate_entity
    generate_repository_interface
    generate_repository_impl
    generate_usecase
    
    if [ "$SKIP_HANDLER" = false ]; then
        if [ "$V2_ONLY" = false ]; then
            generate_dto_v1
            generate_handler_v1
        fi
        
        # TODO: Add v2 generation here
    fi
    
    if [ "$SKIP_MIGRATION" = false ]; then
        generate_migration
    fi
    
    echo ""
    success "========================================="
    success "  Generation Complete!"
    success "========================================="
    echo ""
    
    info "Next steps:"
    echo "  1. Review generated files"
    echo "  2.  Customize entity fields in:  internal/domain/entity/${ENTITY_SNAKE}.go"
    echo "  3. Update migration if needed:  migrations/*_create_${ENTITY_SNAKE}s_table.up.sql"
    echo "  4. Run migration: make migrate-up"
    echo "  5. Add routes in router.go (see example below)"
    echo "  6. Generate Swagger:  make swagger-all"
    echo ""
    
    info "Add these routes to your v1 router:"
    cat << ROUTES
    
    ${ENTITY_PLURAL} := v1.Group("/${ENTITY_PLURAL}")
    ${ENTITY_PLURAL}.Use(r.authMiddleware. RequireAuth())
    {
        ${ENTITY_PLURAL}.POST("", r.${ENTITY_SNAKE}Handler.Create)
        ${ENTITY_PLURAL}.GET("/: id", r.${ENTITY_SNAKE}Handler.GetByID)
        ${ENTITY_PLURAL}.GET("", r.${ENTITY_SNAKE}Handler.List)
        ${ENTITY_PLURAL}.PUT("/:id", r.${ENTITY_SNAKE}Handler.Update)
        ${ENTITY_PLURAL}.DELETE("/: id", r.${ENTITY_SNAKE}Handler.Delete)
    }
ROUTES
    echo ""
}

# Main execution
case $COMMAND in
    entity)
        generate_complete_entity
        ;;
    migration)
        generate_migration
        ;;
    handler)
        if [ "$V2_ONLY" = false ]; then
            generate_dto_v1
            generate_handler_v1
        fi
        ;;
    dto)
        generate_dto_v1
        ;;
    usecase)
        generate_usecase
        ;;
    repository)
        generate_repository_interface
        generate_repository_impl
        ;;
    *)
        error "Unknown command: $COMMAND"
        usage
        ;;
esac
