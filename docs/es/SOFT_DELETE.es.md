# Guía de Implementación de Eliminación Lógica

[🇬🇧 English](../SOFT_DELETE.md) | [🇺🇦 Українська](../uk/SOFT_DELETE.uk.md) | [🇩🇪 Deutsch](../de/SOFT_DELETE.de.md) | [🇵🇹 Português](../pt/SOFT_DELETE.pt.md) | 🇪🇸 **Español**

---

## Visión General

Promenade implementa **eliminación lógica** (soft delete) para las tablas `user_posts` y `post_comments`. La eliminación lógica marca registros como eliminados estableciendo una marca de tiempo en la columna `deleted_at` en lugar de eliminarlos físicamente de la base de datos.

## Beneficios

- **Recuperación de Datos**: El contenido eliminado puede ser restaurado
- **Pista de Auditoría**: Seguimiento de cuándo se eliminó el contenido
- **Integridad Referencial**: Las relaciones de claves foráneas permanecen intactas
- **Analítica**: Los datos históricos permanecen disponibles para análisis
- **Cumplimiento**: Cumplir con requisitos de retención de datos

## Implementación

### Esquema de Base de Datos

Ambas tablas `user_posts` y `post_comments` incluyen:

```sql
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
```

#### Índices

Los índices parciales excluyen registros eliminados lógicamente para mejor rendimiento:

```sql
-- user_posts
CREATE INDEX idx_user_posts_user_id ON user_posts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_posts_published ON user_posts(published_at DESC)
    WHERE status = 'published' AND is_public = true AND deleted_at IS NULL;

-- post_comments
CREATE INDEX idx_post_comments_post_id ON post_comments(post_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_comments_user_id ON post_comments(user_id) WHERE deleted_at IS NULL;
```

Índice para registros eliminados lógicamente (para operaciones de administración/restauración):

```sql
CREATE INDEX idx_user_posts_deleted_at ON user_posts(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_post_comments_deleted_at ON post_comments(deleted_at) WHERE deleted_at IS NOT NULL;
```

### Entidades de Dominio

#### Entidad UserPost

```go
type UserPost struct {
    // ... otros campos
    DeletedAt *time.Time `db:"deleted_at" validate:"omitempty"`
    // ... marcas de tiempo
}

func (p *UserPost) SoftDelete() {
    now := time.Now()
    p.DeletedAt = &now
    p.UpdatedAt = now
}

func (p *UserPost) Restore() {
    p.DeletedAt = nil
    p.UpdatedAt = time.Now()
}
```

#### Entidad PostComment

```go
type PostComment struct {
    // ... otros campos
    DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at" validate:"omitempty"`
    // ... marcas de tiempo
}

func (c *PostComment) SoftDelete() {
    now := time.Now()
    c.DeletedAt = &now
    c.UpdatedAt = now
}

func (c *PostComment) Restore() {
    c.DeletedAt = nil
    c.UpdatedAt = time.Now()
}
```

### Interfaz del Repositorio

```go
type UserPostRepository interface {
    // ... métodos CRUD

    // Operaciones de eliminación lógica
    SoftDelete(ctx context.Context, id uuidv7.UUID) error
    Restore(ctx context.Context, id uuidv7.UUID) error
}

type PostCommentRepository interface {
    // ... métodos CRUD

    // Operaciones de eliminación lógica
    SoftDelete(ctx context.Context, id uuidv7.UUID) error
    Restore(ctx context.Context, id uuidv7.UUID) error
}
```

### Implementación del Repositorio

#### Eliminación Lógica

```go
func (r *userPostRepository) SoftDelete(ctx context.Context, id uuidv7.UUID) error {
    query := `UPDATE user_posts SET deleted_at = NOW(), updated_at = NOW()
              WHERE id = $1 AND deleted_at IS NULL`

    executor := r.getExecutor(ctx)
    result, err := executor.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to soft delete post: %w", err)
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return entity.ErrNotFound
    }

    return nil
}
```

#### Restauración

```go
func (r *userPostRepository) Restore(ctx context.Context, id uuidv7.UUID) error {
    query := `UPDATE user_posts SET deleted_at = NULL, updated_at = NOW()
              WHERE id = $1 AND deleted_at IS NOT NULL`

    executor := r.getExecutor(ctx)
    result, err := executor.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to restore post: %w", err)
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return entity.ErrNotFound
    }

    return nil
}
```

#### Filtrado de Consultas

**CRÍTICO**: Todas las consultas SELECT deben filtrar registros eliminados lógicamente:

```go
func (r *userPostRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
    query := `
        SELECT id, user_id, title, slug, content, ...
        FROM user_posts
        WHERE id = $1 AND deleted_at IS NULL  -- CRÍTICO: Filtrar eliminados lógicamente
    `
    return r.scanPost(ctx, query, id)
}

func (r *userPostRepository) List(ctx context.Context, params ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error) {
    conditions := []string{"deleted_at IS NULL"}  // CRÍTICO: Condición base
    // ... filtros adicionales
}
```

### Capa de Caso de Uso

```go
func (uc *userPostUseCase) SoftDeletePost(ctx context.Context, userID, postID uuidv7.UUID) error {
    // Obtener publicación para verificar propiedad
    post, err := uc.postRepo.GetByID(ctx, postID)
    if err != nil {
        return fmt.Errorf("failed to get post: %w", err)
    }

    // Verificar autorización
    if post.UserID != userID {
        return ErrUnauthorized
    }

    // Realizar eliminación lógica
    if err := uc.postRepo.SoftDelete(ctx, postID); err != nil {
        return fmt.Errorf("failed to soft delete post: %w", err)
    }

    return nil
}
```

#### Eliminación Lógica de Comentario con Efectos Secundarios

```go
func (uc *postCommentUseCase) DeleteComment(ctx context.Context, id, userID uuidv7.UUID) error {
    comment, err := uc.commentRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, entity.ErrNotFound) {
            return ErrCommentNotFound
        }
        return err
    }

    // Verificar propiedad
    if comment.UserID != userID {
        return ErrUnauthorizedComment
    }

    // Eliminación lógica
    if err := uc.commentRepo.SoftDelete(ctx, id); err != nil {
        return err
    }

    // Decrementar contador de respuestas en el padre si es una respuesta
    if comment.ParentID != nil {
        _ = uc.commentRepo.DecrementReplies(ctx, *comment.ParentID)
    }

    // Decrementar contador de comentarios de la publicación
    _ = uc.postRepo.DecrementComments(ctx, comment.PostID)

    return nil
}
```

## Pruebas

### Pruebas Unitarias (Entidad)

```go
func TestUserPost_SoftDelete(t *testing.T) {
    post, _ := NewUserPost(userID, "Test", "test", "Content")
    assert.Nil(t, post.DeletedAt)

    post.SoftDelete()

    assert.NotNil(t, post.DeletedAt)
}

func TestUserPost_Restore(t *testing.T) {
    post, _ := NewUserPost(userID, "Test", "test", "Content")
    post.SoftDelete()
    assert.NotNil(t, post.DeletedAt)

    post.Restore()

    assert.Nil(t, post.DeletedAt)
}
```

### Pruebas de Integración (Repositorio)

```go
func TestUserPostRepository_SoftDelete(t *testing.T) {
    // ... configuración

    err := repo.SoftDelete(ctx, post.ID)
    require.NoError(t, err)

    // No debería encontrarse después de la eliminación lógica
    _, err = repo.GetByID(ctx, post.ID)
    assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestUserPostRepository_Restore(t *testing.T) {
    // ... configuración
    err := repo.SoftDelete(ctx, post.ID)
    require.NoError(t, err)

    err = repo.Restore(ctx, post.ID)
    require.NoError(t, err)

    // Verificar restauración
    restored, err := repo.GetByID(ctx, post.ID)
    require.NoError(t, err)
    assert.Equal(t, post.ID, restored.ID)
    assert.Nil(t, restored.DeletedAt)
}
```

### Pruebas de Humo (End-to-End)

```go
t.Run("[+] Soft_delete_and_restore", func(t *testing.T) {
    // Eliminación lógica
    require.NoError(t, postRepo.SoftDelete(ctx, postID))

    // Verificar que no es accesible
    _, err := postRepo.GetByID(ctx, postID)
    assert.ErrorIs(t, err, entity.ErrNotFound)

    // Restaurar
    require.NoError(t, postRepo.Restore(ctx, postID))

    // Verificar que es accesible nuevamente
    restored, err := postRepo.GetByID(ctx, postID)
    require.NoError(t, err)
    assert.Equal(t, postID, restored.ID)
})
```

## Patrones Comunes

### 1. Verificar Antes de Eliminar Lógicamente

Siempre verificar propiedad/permisos antes de la eliminación lógica:

```go
post, err := uc.postRepo.GetByID(ctx, postID)
if err != nil {
    return err
}

if post.UserID != userID {
    return ErrUnauthorized
}

return uc.postRepo.SoftDelete(ctx, postID)
```

### 2. Operaciones Idempotentes

La eliminación lógica devuelve `ErrNotFound` si el registro ya está eliminado:

```go
// Primera eliminación - éxito
err := repo.SoftDelete(ctx, postID)  // nil

// Segunda eliminación - error
err = repo.SoftDelete(ctx, postID)   // entity.ErrNotFound
```

### 3. Patrón de Exclusión en Consultas

Todas las consultas de lista/búsqueda deben excluir registros eliminados lógicamente:

```go
conditions := []string{"deleted_at IS NULL"}

// Agregar filtros adicionales
if params.UserID != nil {
    conditions = append(conditions, fmt.Sprintf("user_id = '%s'", *params.UserID))
}

query := fmt.Sprintf("SELECT ... WHERE %s", strings.Join(conditions, " AND "))
```

### 4. Consultas de Administrador (Incluyendo Eliminados)

Para interfaces de administración, crear métodos separados para consultar registros eliminados:

```go
// Implementación futura
func (r *userPostRepository) GetByIDIncludingDeleted(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
    query := `SELECT ... FROM user_posts WHERE id = $1`  // Sin filtro deleted_at
    return r.scanPost(ctx, query, id)
}
```

## Mejores Prácticas

### SÍ [+]

- Siempre filtrar `deleted_at IS NULL` en consultas SELECT
- Usar índices parciales: `WHERE deleted_at IS NULL` en columnas consultadas frecuentemente
- Verificar propiedad antes de la eliminación lógica
- Actualizar contadores relacionados (como reply_count, comment_count) después de la eliminación lógica
- Probar tanto las operaciones de eliminación lógica como las de restauración
- Documentar qué entidades soportan eliminación lógica

### NO [-]

- No olvidar `deleted_at IS NULL` en consultas (¡bug común!)
- No usar eliminación física (`DELETE FROM`) para entidades con eliminación lógica
- No exponer registros eliminados lógicamente en APIs públicas sin autorización
- No restaurar sin verificar permisos (en producción, agregar verificación de propiedad)
- No aplicar cascada de eliminaciones lógicas automáticamente (decisión de diseño por caso de uso)

## Limitaciones Conocidas

### Implementación Actual

1. **Sin verificación de propiedad en restauración**: `RestorePost()` no verifica la propiedad porque `GetByID()` filtra los registros eliminados. En producción, implementar `GetByIDIncludingDeleted()` para esta verificación.

2. **Sin cascada de eliminación lógica**: Eliminar una publicación no elimina lógicamente sus comentarios automáticamente. Esto es intencional - decisión de diseño para preservar el historial de comentarios.

3. **Sin marca de tiempo de restauración**: La implementación no rastrea cuándo se restauró un registro (se podría agregar `restored_at` si es necesario).

## Mejoras Futuras

- [ ] Agregar `GetByIDIncludingDeleted()` para operaciones de administración
- [ ] Agregar `ListDeleted()` para UI de papelera/recuperación de administrador
- [ ] Agregar marca de tiempo `restored_at` para pista de auditoría
- [ ] Agregar opción de cascada de eliminación lógica para publicaciones → comentarios
- [ ] Agregar operaciones de eliminación lógica/restauración en lote
- [ ] Agregar eliminación física automática después de X días (tarea programada)

## Ver También

- [Guía de Migración UUID v7](UUID_V7_MIGRATION.md) - Estrategia de clave primaria
- [Guía de Pruebas](TESTING_GUIDE.md) - Patrones de prueba y helpers
- [Migraciones de Base de Datos](../migrations/) - Definiciones de esquema
- [Base de Repositorio](../internal/adapter/repository/postgres/base_repository.go) - Operaciones comunes de BD
