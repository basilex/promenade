# Arquitetura de Migrações de Módulos

[🇬🇧 English](../MIGRATION_ARCHITECTURE.md) | [🇺🇦 Українська](MIGRATION_ARCHITECTURE.uk.md) | [🇩🇪 Deutsch](MIGRATION_ARCHITECTURE.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](MIGRATION_ARCHITECTURE.es.md)

## Problema

Sistema de migração atual viola independência de módulos:

- Todas as migrações em uma única pasta `migrations/`
- Numeração sequencial global (000001, 000002, ...)
- Migrações core misturadas com migrações de módulos
- Sem forma de habilitar/desabilitar migrações de módulos independentemente

## Solução: Migrações Baseadas em Namespace

### 1. Estrutura de Diretórios

```
migrations/
├── core/                          # Migrações de infraestrutura core
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_auth_tables.up.sql
│   ├── 000002_auth_tables.down.sql
│   └── ...
│
├── posts/                         # Migrações do módulo Posts
│   ├── 000001_create_posts.up.sql
│   ├── 000001_create_posts.down.sql
│   ├── 000002_create_comments.up.sql
│   └── ...
│
└── profiles/                      # Migrações do módulo Profiles
    ├── 000001_create_profiles.up.sql
    └── ...
```

**Cada namespace tem versionamento independente:**

- Core: 1, 2, 3, 4, ...
- Posts: 1, 2, 3, ...
- Profiles: 1, 2, ...

---

### 2. Schema da Tabela de Migração

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     BIGINT       NOT NULL,
    namespace   VARCHAR(50)  NOT NULL,  -- 'core', 'posts', 'profiles', etc.
    dirty       BOOLEAN      NOT NULL DEFAULT FALSE,
    applied_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (namespace, version)
);

CREATE INDEX idx_schema_migrations_namespace ON schema_migrations(namespace);
```

**Dados de exemplo:**

```
| version | namespace | dirty | applied_at          |
|---------|-----------|-------|---------------------|
| 1       | core      | false | 2024-12-22 10:00:00 |
| 2       | core      | false | 2024-12-22 10:00:01 |
| 3       | core      | false | 2024-12-22 10:00:02 |
| 1       | posts     | false | 2024-12-22 10:00:03 |
| 2       | posts     | false | 2024-12-22 10:00:04 |
| 1       | profiles  | false | 2024-12-22 10:00:05 |
```

---

### 3. Migration Manager (pkg/migration/)

#### Interface

```go
package migration

type Manager interface {
    // MigrateNamespace aplica todas as migrações pendentes para um namespace
    MigrateNamespace(ctx context.Context, namespace string) error

    // MigrateAll aplica todas as migrações pendentes (core + módulos habilitados)
    MigrateAll(ctx context.Context, enabledModules []string) error

    // Rollback reverte N migrações para um namespace
    Rollback(ctx context.Context, namespace string, steps int) error

    // Version retorna versão atual para um namespace
    Version(ctx context.Context, namespace string) (int, error)

    // Status retorna status de migração para todos os namespaces
    Status(ctx context.Context) (map[string]MigrationStatus, error)
}

type MigrationStatus struct {
    Namespace      string
    CurrentVersion int
    PendingCount   int
    Dirty          bool
}
```

#### Implementação

```go
package migration

type manager struct {
    db            *sqlx.DB
    migrationsDir string  // "migrations/" por padrão
}

func NewManager(db *sqlx.DB, migrationsDir string) Manager {
    return &manager{db: db, migrationsDir: migrationsDir}
}

func (m *manager) MigrateNamespace(ctx context.Context, namespace string) error {
    // 1. Verificar se tabela schema_migrations existe
    // 2. Obter versão atual para namespace
    // 3. Ler arquivos de migração de migrations/{namespace}/
    // 4. Aplicar migrações pendentes em transação
    // 5. Atualizar tabela schema_migrations
}

func (m *manager) MigrateAll(ctx context.Context, enabledModules []string) error {
    // 1. Sempre migrar core primeiro
    if err := m.MigrateNamespace(ctx, "core"); err != nil {
        return err
    }

    // 2. Migrar cada módulo habilitado
    for _, module := range enabledModules {
        if err := m.MigrateNamespace(ctx, module); err != nil {
            return err
        }
    }

    return nil
}
```

---

### 4. Integração de Módulos

Módulos podem opcionalmente fornecer migrações programaticamente:

```go
// pkg/module/module.go
type Module interface {
    // ... métodos existentes ...

    // RegisterMigrations retorna migrações incorporadas para este módulo
    // Retornar nil para usar migrações baseadas em arquivo de migrations/{namespace}/
    RegisterMigrations() []Migration
}

type Migration struct {
    Version     int
    Description string
    Up          string  // SQL para aplicar
    Down        string  // SQL para reverter
}
```

Exemplo em módulo:

```go
// internal/modules/posts/module.go
func (m *PostsModule) RegisterMigrations() []Migration {
    // Opção 1: Retornar nil para usar migrações baseadas em arquivo
    return nil

    // Opção 2: Incorporar migrações no código (para bibliotecas)
    return []Migration{
        {
            Version:     1,
            Description: "Create posts table",
            Up:          `CREATE TABLE user_posts (...)`,
            Down:        `DROP TABLE user_posts`,
        },
    }
}
```

---

### 5. Uso em main.go

```go
// cmd/api/main.go
func main() {
    // ... setup ...

    // Inicializar migration manager
    migrationMgr := migration.NewManager(db, "migrations")

    // Obter módulos habilitados da config
    enabledModules := cfg.Modules.Enabled  // ["posts", "profiles"]

    // Executar migrações para core + módulos habilitados
    if err := migrationMgr.MigrateAll(ctx, enabledModules); err != nil {
        log.Fatal("Failed to run migrations", "error", err)
    }

    // ... continuar startup ...
}
```

---

### 6. Comandos CLI

```bash
# Migrar apenas core
make migrate-core

# Migrar módulo específico
make migrate-module MODULE=posts

# Migrar todos (core + módulos habilitados)
make migrate-all

# Reverter migrações de módulo
make migrate-rollback MODULE=posts STEPS=1

# Mostrar status de migração
make migrate-status
```

**Makefile:**

```makefile
migrate-core:
	go run cmd/migrate/main.go up --namespace=core

migrate-module:
	go run cmd/migrate/main.go up --namespace=$(MODULE)

migrate-all:
	go run cmd/migrate/main.go up --all

migrate-rollback:
	go run cmd/migrate/main.go down --namespace=$(MODULE) --steps=$(STEPS)

migrate-status:
	go run cmd/migrate/main.go status
```

---

### 7. Reorganização de Migrações

**Atual (flat - DESATUALIZADO, usar com namespace):**

```
migrations/
├── 000001_init_schema_deps.up.sql         # core
├── 000002_create_auth_schema.up.sql       # core
├── 000003_create_countries_currencies.up.sql  # core
├── 000004_create_user_contacts.up.sql     # módulo profiles
├── 000005_create_user_profiles.up.sql     # módulo profiles
├── 000006_create_user_posts.up.sql        # módulo posts
├── 000007_create_post_comments.up.sql     # módulo posts
├── 000008_create_comment_likes_table.up.sql  # módulo posts
├── 000009_create_rbac_tables.up.sql       # core
├── 000014_create_timezones_table.up.sql   # core
└── 000015_create_languages_table.up.sql   # core
```

**Novo (com namespace e nomes descritivos):**

```
migrations/
├── core/
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000001_core_init_uuid_v7.down.sql
│   ├── 000002_core_auth_full.up.sql
│   ├── 000002_core_auth_full.down.sql
│   ├── 000003_core_rbac_full.up.sql            # era 000009
│   ├── 000003_core_rbac_full.down.sql
│   ├── 000004_core_ref_timezones.up.sql        # era 000014
│   ├── 000004_core_ref_timezones.down.sql
│   ├── 000005_core_ref_languages.up.sql        # era 000015
│   ├── 000005_core_ref_languages.down.sql
│   ├── 000006_core_ref_countries_currencies.up.sql  # era 000003
│   └── 000006_core_ref_countries_currencies.down.sql
│
├── posts/
│   ├── 000001_posts_posts.up.sql               # era 000006_create_user_posts
│   ├── 000001_posts_posts.down.sql
│   ├── 000002_posts_comments.up.sql            # era 000007_create_post_comments
│   ├── 000002_posts_comments.down.sql
│   ├── 000003_create_comment_likes.up.sql    # era 000008
│   └── 000003_create_comment_likes.down.sql
│
└── profiles/
    ├── 000001_create_user_contacts.up.sql    # era 000004
    ├── 000001_create_user_contacts.down.sql
    ├── 000002_create_user_profiles.up.sql    # era 000005
    └── 000002_create_user_profiles.down.sql
```

---

### 8. Benefícios

**Independência de Módulos**

- Cada módulo possui suas migrações
- Habilitar/desabilitar módulos sem conflitos de migração
- Limites de propriedade claros

**Controle de Versão**

- Cada namespace tem versionamento independente
- Sem conflitos de numeração global
- Fácil entender em qual versão um módulo está

**Implantação Flexível**

- Implantar com apenas módulos necessários
- Adicionar novos módulos sem tocar em migrações existentes
- Reverter migrações de módulos independentemente

**Experiência do Desenvolvedor**

- Claro onde colocar novas migrações
- Sem adivinhação do próximo número global
- Comandos de migração específicos por módulo

---

### 9. Fluxo de Trabalho de Migração

#### Criando uma Nova Migração

```bash
# Migração core
make migrate-create-core NAME=add_audit_tables

# Migração de módulo
make migrate-create MODULE=posts NAME=add_post_views
```

**Arquivos gerados:**

```
migrations/posts/
├── 000004_add_post_views.up.sql    # Auto-incrementado
└── 000004_add_post_views.down.sql
```

#### Aplicando Migrações

```bash
# Desenvolvimento: Migrar tudo
make migrate-all

# Produção: Migrar core + módulos específicos
MODULES=posts,profiles make migrate-all

# Seletivo: Migrar apenas novo módulo
make migrate-module MODULE=warehouse
```

---

### 10. Compatibilidade Retroativa

Para implantações existentes:

1. **Script de migração único** que:
   - Faz backup da tabela `schema_migrations` atual
   - Cria nova `schema_migrations` com namespace
   - Mapeia versões antigas para versões com namespace
   - Marca todas como aplicadas

```sql
-- Backup
CREATE TABLE schema_migrations_backup AS SELECT * FROM schema_migrations;

-- Remover tabela antiga
DROP TABLE schema_migrations;

-- Criar nova tabela com namespace
CREATE TABLE schema_migrations (...);

-- Inserir versões mapeadas
INSERT INTO schema_migrations (namespace, version, dirty, applied_at)
VALUES
    ('core', 1, false, NOW()),  -- era 000001
    ('core', 2, false, NOW()),  -- era 000002
    ('core', 3, false, NOW()),  -- era 000003
    ('profiles', 1, false, NOW()),  -- era 000004
    ('profiles', 2, false, NOW()),  -- era 000005
    ('posts', 1, false, NOW()),  -- era 000006
    ...
```

2. **Script de reorganização de migração** que move arquivos para pastas de namespace

---

### 11. Plano de Implementação

**Fase 1: Infraestrutura (Semana 1)**

1. Criar pacote `pkg/migration/`
2. Implementar interface `Manager`
3. Adicionar suporte a namespace na tabela schema_migrations
4. Escrever testes

**Fase 2: CLI & Ferramentas (Semana 1)**

1. Criar ferramenta CLI `cmd/migrate/main.go`
2. Adicionar comandos Makefile
3. Atualizar documentação

**Fase 3: Migração (Semana 2)**

1. Reorganizar migrações existentes em namespaces
2. Criar migração de compatibilidade retroativa
3. Atualizar main.go para usar novo manager
4. Testar em staging

**Fase 4: Integração de Módulos (Semana 2)**

1. Adicionar `RegisterMigrations()` à interface de módulo
2. Atualizar módulos existentes
3. Documentação & exemplos

---

### 12. Alternativa: Fork golang-migrate

Se quisermos usar a biblioteca `golang-migrate`:

```go
import "github.com/golang-migrate/migrate/v4"

// Driver de source customizado que lê de pastas de namespace
type NamespaceSource struct {
    namespace string
    basePath  string
}

func (s *NamespaceSource) First() (version uint, err error) {
    // Ler de migrations/{namespace}/
}

// Registrar source customizado
migrate.Register("namespace", &NamespaceSource{})
```

---

## Resumo

**Melhor Solução:** Migration manager customizado com suporte a namespace.

**Por quê?**

- Controle total sobre lógica de namespace
- Versionamento independente de módulos
- Fácil implementar habilitar/desabilitar módulo
- Limites de propriedade claros
- Sem restrições de biblioteca externa

**Caminho de Migração:**

1. Implementar manager `pkg/migration/`
2. Adicionar namespace a schema_migrations
3. Reorganizar migrações existentes
4. Atualizar código de startup
5. Documentar fluxo de trabalho

**Esforço Estimado:** 2-3 dias para implementação completa + testes
