# Soft Delete Implementierungsleitfaden

[ English](../SOFT_DELETE.md) | [ Українська](../uk/SOFT_DELETE.uk.md) |  **Deutsch** | [ Português](../pt/SOFT_DELETE.pt.md) | [ Español](../es/SOFT_DELETE.es.md)

## Übersicht

Promenade implementiert **Soft Delete** für die Tabellen `user_posts` und `post_comments`. Soft Delete markiert Datensätze als gelöscht, indem ein Zeitstempel in der Spalte `deleted_at` gesetzt wird, anstatt sie physisch aus der Datenbank zu entfernen.

## Vorteile

- **Datenwiederherstellung**: Gelöschte Inhalte können wiederhergestellt werden
- **Audit Trail**: Nachverfolgung, wann Inhalte gelöscht wurden
- **Referenzielle Integrität**: Foreign Key Beziehungen bleiben intakt
- **Analytics**: Historische Daten bleiben für Analysen verfügbar
- **Compliance**: Erfüllung von Anforderungen zur Datenaufbewahrung

## Implementierung

### Datenbankschema

Sowohl `user_posts` als auch `post_comments` Tabellen enthalten:

```sql
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
```

#### Indizes

Partielle Indizes schließen soft-gelöschte Datensätze für bessere Performance aus:

```sql
-- user_posts
CREATE INDEX idx_user_posts_user_id ON user_posts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_posts_published ON user_posts(published_at DESC)
    WHERE status = 'published' AND is_public = true AND deleted_at IS NULL;

-- post_comments
CREATE INDEX idx_post_comments_post_id ON post_comments(post_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_comments_user_id ON post_comments(user_id) WHERE deleted_at IS NULL;
```

Index für soft-gelöschte Datensätze (für Admin/Wiederherstellungsoperationen):

```sql
CREATE INDEX idx_user_posts_deleted_at ON user_posts(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_post_comments_deleted_at ON post_comments(deleted_at) WHERE deleted_at IS NOT NULL;
```

### Domain Entities

#### UserPost Entity

```go
type UserPost struct {
    // ... andere Felder
    DeletedAt *time.Time `db:"deleted_at" validate:"omitempty"`
    // ... timestamps
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

#### PostComment Entity

```go
type PostComment struct {
    // ... andere Felder
    DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at" validate:"omitempty"`
    // ... timestamps
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

### Repository Interface

```go
type IUserPostRepository interface {
    // ... CRUD Methoden

    // Soft delete Operationen
    SoftDelete(ctx context.Context, id uuidv7.UUID) error
    Restore(ctx context.Context, id uuidv7.UUID) error
}

type IPostCommentRepository interface {
    // ... CRUD Methoden

    // Soft delete Operationen
    SoftDelete(ctx context.Context, id uuidv7.UUID) error
    Restore(ctx context.Context, id uuidv7.UUID) error
}
```

### Repository Implementierung

#### Soft Delete

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

#### Restore

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

#### Query Filterung

**KRITISCH**: Alle SELECT Abfragen müssen soft-gelöschte Datensätze ausfiltern:

```go
func (r *userPostRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
    query := `
        SELECT id, user_id, title, slug, content, ...
        FROM user_posts
        WHERE id = $1 AND deleted_at IS NULL  -- KRITISCH: Soft-gelöschte filtern
    `
    return r.scanPost(ctx, query, id)
}

func (r *userPostRepository) List(ctx context.Context, params ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error) {
    conditions := []string{"deleted_at IS NULL"}  // KRITISCH: Basisbedingung
    // ... zusätzliche Filter
}
```

### Use Case Layer

```go
func (uc *userPostUseCase) SoftDeletePost(ctx context.Context, userID, postID uuidv7.UUID) error {
    // Post abrufen um Eigentümerschaft zu prüfen
    post, err := uc.postRepo.GetByID(ctx, postID)
    if err != nil {
        return fmt.Errorf("failed to get post: %w", err)
    }

    // Autorisierung prüfen
    if post.UserID != userID {
        return ErrUnauthorized
    }

    // Soft delete durchführen
    if err := uc.postRepo.SoftDelete(ctx, postID); err != nil {
        return fmt.Errorf("failed to soft delete post: %w", err)
    }

    return nil
}
```

#### Comment Soft Delete mit Seiteneffekten

```go
func (uc *postCommentUseCase) DeleteComment(ctx context.Context, id, userID uuidv7.UUID) error {
    comment, err := uc.commentRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, entity.ErrNotFound) {
            return ErrCommentNotFound
        }
        return err
    }

    // Eigentümerschaft prüfen
    if comment.UserID != userID {
        return ErrUnauthorizedComment
    }

    // Soft delete
    if err := uc.commentRepo.SoftDelete(ctx, id); err != nil {
        return err
    }

    // Reply-Anzahl beim Parent dekrementieren, falls es eine Antwort ist
    if comment.ParentID != nil {
        _ = uc.commentRepo.DecrementReplies(ctx, *comment.ParentID)
    }

    // Kommentaranzahl des Posts dekrementieren
    _ = uc.postRepo.DecrementComments(ctx, comment.PostID)

    return nil
}
```

## Testing

### Unit Tests (Entity)

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

### Integration Tests (Repository)

```go
func TestUserPostRepository_SoftDelete(t *testing.T) {
    // ... Setup

    err := repo.SoftDelete(ctx, post.ID)
    require.NoError(t, err)

    // Sollte nach Soft Delete nicht gefunden werden
    _, err = repo.GetByID(ctx, post.ID)
    assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestUserPostRepository_Restore(t *testing.T) {
    // ... Setup
    err := repo.SoftDelete(ctx, post.ID)
    require.NoError(t, err)

    err = repo.Restore(ctx, post.ID)
    require.NoError(t, err)

    // Wiederherstellung verifizieren
    restored, err := repo.GetByID(ctx, post.ID)
    require.NoError(t, err)
    assert.Equal(t, post.ID, restored.ID)
    assert.Nil(t, restored.DeletedAt)
}
```

### Smoke Tests (End-to-End)

```go
t.Run("[+] Soft_delete_and_restore", func(t *testing.T) {
    // Soft delete
    require.NoError(t, postRepo.SoftDelete(ctx, postID))

    // Verifizieren, dass nicht zugreifbar
    _, err := postRepo.GetByID(ctx, postID)
    assert.ErrorIs(t, err, entity.ErrNotFound)

    // Wiederherstellen
    require.NoError(t, postRepo.Restore(ctx, postID))

    // Verifizieren, dass wieder zugreifbar
    restored, err := postRepo.GetByID(ctx, postID)
    require.NoError(t, err)
    assert.Equal(t, postID, restored.ID)
})
```

## Gängige Patterns

### 1. Prüfung vor Soft Delete

Immer Eigentümerschaft/Berechtigungen vor Soft Delete verifizieren:

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

### 2. Idempotente Operationen

Soft Delete gibt `ErrNotFound` zurück, wenn der Datensatz bereits gelöscht ist:

```go
// Erstes Löschen - Erfolg
err := repo.SoftDelete(ctx, postID)  // nil

// Zweites Löschen - Fehler
err = repo.SoftDelete(ctx, postID)   // entity.ErrNotFound
```

### 3. Query Ausschluss-Pattern

Alle List/Such-Abfragen müssen soft-gelöschte Datensätze ausschließen:

```go
conditions := []string{"deleted_at IS NULL"}

// Zusätzliche Filter hinzufügen
if params.UserID != nil {
    conditions = append(conditions, fmt.Sprintf("user_id = '%s'", *params.UserID))
}

query := fmt.Sprintf("SELECT ... WHERE %s", strings.Join(conditions, " AND "))
```

### 4. Admin Queries (Inklusive Gelöschter)

Für Admin-Interfaces separate Methoden erstellen, um gelöschte Datensätze abzufragen:

```go
// Zukünftige Implementierung
func (r *userPostRepository) GetByIDIncludingDeleted(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
    query := `SELECT ... FROM user_posts WHERE id = $1`  // Kein deleted_at Filter
    return r.scanPost(ctx, query, id)
}
```

## Best Practices

### DO [+]

- Immer `deleted_at IS NULL` in SELECT Abfragen filtern
- Partielle Indizes verwenden: `WHERE deleted_at IS NULL` auf häufig abgefragten Spalten
- Eigentümerschaft vor Soft Delete prüfen
- Zugehörige Zähler (wie reply_count, comment_count) nach Soft Delete aktualisieren
- Sowohl Soft Delete als auch Restore Operationen testen
- Dokumentieren, welche Entities Soft Delete unterstützen

### DON'T [-]

- `deleted_at IS NULL` in Abfragen nicht vergessen (häufiger Bug!)
- Kein Hard Delete (`DELETE FROM`) für soft-deletable Entities verwenden
- Soft-gelöschte Datensätze nicht ohne Autorisierung in öffentlichen APIs exponieren
- Nicht ohne Berechtigungsprüfung wiederherstellen (in Produktion Eigentümerschaftsprüfung hinzufügen)
- Soft Deletes nicht automatisch kaskadieren (Designentscheidung pro Anwendungsfall)

## Bekannte Einschränkungen

### Aktuelle Implementierung

1. **Keine Eigentümerschaftsprüfung bei Wiederherstellung**: `RestorePost()` verifiziert keine Eigentümerschaft, da `GetByID()` gelöschte Datensätze ausfiltert. In Produktion `GetByIDIncludingDeleted()` für diese Prüfung implementieren.

2. **Kein kaskadierendes Soft Delete**: Das Löschen eines Posts löscht nicht automatisch seine Kommentare per Soft Delete. Dies ist beabsichtigt - eine Designentscheidung zur Erhaltung der Kommentarhistorie.

3. **Kein Restore Zeitstempel**: Die Implementierung verfolgt nicht, wann ein Datensatz wiederhergestellt wurde (könnte bei Bedarf `restored_at` hinzufügen).

## Zukünftige Erweiterungen

- [ ] `GetByIDIncludingDeleted()` für Admin-Operationen hinzufügen
- [ ] `ListDeleted()` für Admin Trash/Recovery UI hinzufügen
- [ ] `restored_at` Zeitstempel für Audit Trail hinzufügen
- [ ] Kaskadierende Soft Delete Option für Posts → Comments hinzufügen
- [ ] Bulk Soft Delete/Restore Operationen hinzufügen
- [ ] Automatisches Hard Delete nach X Tagen hinzufügen (geplanter Job)

## Siehe auch

- [UUID v7 Migrationsleitfaden](UUID_V7_GUIDE.de.md) - Primary Key Strategie
- [Testing Guide](TESTING_GUIDE.de.md) - Testing Patterns und Helpers
- [Datenbankmigrationen](../../migrations/) - Schema Definitionen
- [Repository Base](../../internal/adapter/repository/postgres/base_repository.go) - Gemeinsame DB Operationen
