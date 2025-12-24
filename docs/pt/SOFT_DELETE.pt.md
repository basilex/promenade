# Guia de Implementação de Soft Delete

[🇬🇧 English](../SOFT_DELETE.md) | [🇺🇦 Українська](SOFT_DELETE.uk.md) | [🇩🇪 Deutsch](SOFT_DELETE.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](SOFT_DELETE.es.md)

## Visão Geral

Promenade implementa **soft delete** para as tabelas `user_posts` e `post_comments`. Soft delete marca registros como excluídos definindo um timestamp na coluna `deleted_at` em vez de removê-los fisicamente do banco de dados.

## Benefícios

- **Recuperação de Dados**: Conteúdo excluído pode ser restaurado
- **Trilha de Auditoria**: Rastrear quando o conteúdo foi excluído
- **Integridade Referencial**: Relacionamentos de chave estrangeira permanecem intactos
- **Analytics**: Dados históricos permanecem disponíveis para análise
- **Conformidade**: Atender requisitos de retenção de dados

## Implementação

### Schema do Banco de Dados

Ambas as tabelas `user_posts` e `post_comments` incluem:

```sql
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
```

#### Índices

Índices parciais excluem registros soft-deleted para melhor desempenho:

```sql
-- user_posts
CREATE INDEX idx_user_posts_user_id ON user_posts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_posts_published ON user_posts(published_at DESC)
    WHERE status = 'published' AND is_public = true AND deleted_at IS NULL;

-- post_comments
CREATE INDEX idx_post_comments_post_id ON post_comments(post_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_comments_user_id ON post_comments(user_id) WHERE deleted_at IS NULL;
```

Índice para registros soft-deleted (para operações de admin/restauração):

```sql
CREATE INDEX idx_user_posts_deleted_at ON user_posts(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_post_comments_deleted_at ON post_comments(deleted_at) WHERE deleted_at IS NOT NULL;
```

### Entidades de Domínio

#### Entidade UserPost

```go
type UserPost struct {
    // ... outros campos
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

#### Entidade PostComment

```go
type PostComment struct {
    // ... outros campos
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

### Interface do Repository

```go
type IUserPostRepository interface {
    // ... métodos CRUD

    // Operações de soft delete
    SoftDelete(ctx context.Context, id uuidv7.UUID) error
    Restore(ctx context.Context, id uuidv7.UUID) error
}

type IPostCommentRepository interface {
    // ... métodos CRUD

    // Operações de soft delete
    SoftDelete(ctx context.Context, id uuidv7.UUID) error
    Restore(ctx context.Context, id uuidv7.UUID) error
}
```

### Implementação do Repository

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

#### Restauração

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

#### Filtragem de Consultas

**CRÍTICO**: Todas as consultas SELECT devem filtrar registros soft-deleted:

```go
func (r *userPostRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
    query := `
        SELECT id, user_id, title, slug, content, ...
        FROM user_posts
        WHERE id = $1 AND deleted_at IS NULL  -- CRÍTICO: Filtrar soft-deleted
    `
    return r.scanPost(ctx, query, id)
}

func (r *userPostRepository) List(ctx context.Context, params ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error) {
    conditions := []string{"deleted_at IS NULL"}  // CRÍTICO: Condição base
    // ... filtros adicionais
}
```

### Camada de Use Case

```go
func (uc *userPostUseCase) SoftDeletePost(ctx context.Context, userID, postID uuidv7.UUID) error {
    // Obter post para verificar propriedade
    post, err := uc.postRepo.GetByID(ctx, postID)
    if err != nil {
        return fmt.Errorf("failed to get post: %w", err)
    }

    // Verificar autorização
    if post.UserID != userID {
        return ErrUnauthorized
    }

    // Executar soft delete
    if err := uc.postRepo.SoftDelete(ctx, postID); err != nil {
        return fmt.Errorf("failed to soft delete post: %w", err)
    }

    return nil
}
```

#### Soft Delete de Comentário com Efeitos Colaterais

```go
func (uc *postCommentUseCase) DeleteComment(ctx context.Context, id, userID uuidv7.UUID) error {
    comment, err := uc.commentRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, entity.ErrNotFound) {
            return ErrCommentNotFound
        }
        return err
    }

    // Verificar propriedade
    if comment.UserID != userID {
        return ErrUnauthorizedComment
    }

    // Soft delete
    if err := uc.commentRepo.SoftDelete(ctx, id); err != nil {
        return err
    }

    // Decrementar contagem de respostas no pai se for uma resposta
    if comment.ParentID != nil {
        _ = uc.commentRepo.DecrementReplies(ctx, *comment.ParentID)
    }

    // Decrementar contagem de comentários do post
    _ = uc.postRepo.DecrementComments(ctx, comment.PostID)

    return nil
}
```

## Testes

### Testes Unitários (Entidade)

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

### Testes de Integração (Repository)

```go
func TestUserPostRepository_SoftDelete(t *testing.T) {
    // ... setup

    err := repo.SoftDelete(ctx, post.ID)
    require.NoError(t, err)

    // Não deve ser encontrado após soft delete
    _, err = repo.GetByID(ctx, post.ID)
    assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestUserPostRepository_Restore(t *testing.T) {
    // ... setup
    err := repo.SoftDelete(ctx, post.ID)
    require.NoError(t, err)

    err = repo.Restore(ctx, post.ID)
    require.NoError(t, err)

    // Verificar restauração
    restored, err := repo.GetByID(ctx, post.ID)
    require.NoError(t, err)
    assert.Equal(t, post.ID, restored.ID)
    assert.Nil(t, restored.DeletedAt)
}
```

### Testes Smoke (End-to-End)

```go
t.Run("[+] Soft_delete_and_restore", func(t *testing.T) {
    // Soft delete
    require.NoError(t, postRepo.SoftDelete(ctx, postID))

    // Verificar não acessível
    _, err := postRepo.GetByID(ctx, postID)
    assert.ErrorIs(t, err, entity.ErrNotFound)

    // Restaurar
    require.NoError(t, postRepo.Restore(ctx, postID))

    // Verificar acessível novamente
    restored, err := postRepo.GetByID(ctx, postID)
    require.NoError(t, err)
    assert.Equal(t, postID, restored.ID)
})
```

## Padrões Comuns

### 1. Verificar Antes de Soft Delete

Sempre verificar propriedade/permissões antes de soft delete:

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

### 2. Operações Idempotentes

Soft delete retorna `ErrNotFound` se o registro já estiver excluído:

```go
// Primeiro delete - sucesso
err := repo.SoftDelete(ctx, postID)  // nil

// Segundo delete - erro
err = repo.SoftDelete(ctx, postID)   // entity.ErrNotFound
```

### 3. Padrão de Exclusão de Consulta

Todas as consultas list/search devem excluir registros soft-deleted:

```go
conditions := []string{"deleted_at IS NULL"}

// Adicionar filtros adicionais
if params.UserID != nil {
    conditions = append(conditions, fmt.Sprintf("user_id = '%s'", *params.UserID))
}

query := fmt.Sprintf("SELECT ... WHERE %s", strings.Join(conditions, " AND "))
```

### 4. Consultas de Admin (Incluindo Excluídos)

Para interfaces administrativas, criar métodos separados para consultar registros excluídos:

```go
// Implementação futura
func (r *userPostRepository) GetByIDIncludingDeleted(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
    query := `SELECT ... FROM user_posts WHERE id = $1`  // Sem filtro deleted_at
    return r.scanPost(ctx, query, id)
}
```

## Melhores Práticas

### FAÇA [+]

- Sempre filtrar `deleted_at IS NULL` em consultas SELECT
- Usar índices parciais: `WHERE deleted_at IS NULL` em colunas consultadas frequentemente
- Verificar propriedade antes de soft delete
- Atualizar contadores relacionados (como reply_count, comment_count) após soft delete
- Testar operações de soft delete e restauração
- Documentar quais entidades suportam soft delete

### NÃO FAÇA [-]

- Não esquecer `deleted_at IS NULL` nas consultas (bug comum!)
- Não usar hard delete (`DELETE FROM`) para entidades soft-deletáveis
- Não expor registros soft-deleted em APIs públicas sem autorização
- Não restaurar sem verificar permissões (em produção, adicionar verificação de propriedade)
- Não fazer cascade de soft deletes automaticamente (decisão de design por caso de uso)

## Limitações Conhecidas

### Implementação Atual

1. **Sem verificação de propriedade na restauração**: `RestorePost()` não verifica propriedade porque `GetByID()` filtra registros excluídos. Em produção, implementar `GetByIDIncludingDeleted()` para essa verificação.

2. **Sem cascade de soft delete**: Excluir um post não exclui automaticamente seus comentários. Isso é intencional - decisão de design para preservar histórico de comentários.

3. **Sem timestamp de restauração**: A implementação não rastreia quando um registro foi restaurado (poderia adicionar `restored_at` se necessário).

## Melhorias Futuras

- [ ] Adicionar `GetByIDIncludingDeleted()` para operações de admin
- [ ] Adicionar `ListDeleted()` para UI de lixeira/recuperação de admin
- [ ] Adicionar timestamp `restored_at` para trilha de auditoria
- [ ] Adicionar opção de cascade soft delete para posts → comentários
- [ ] Adicionar operações de soft delete/restauração em lote
- [ ] Adicionar hard delete automático após X dias (job agendado)

## Veja Também

- [Guia de Migração UUID v7](UUID_V7_MIGRATION.md) - Estratégia de chave primária
- [Guia de Testes](TESTING_GUIDE.md) - Padrões e helpers de testes
- [Migrações de Banco de Dados](../migrations/) - Definições de schema
- [Base Repository](../internal/adapter/repository/postgres/base_repository.go) - Operações comuns de BD
