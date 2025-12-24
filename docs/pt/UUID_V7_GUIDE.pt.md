# Guia de Implementação UUID v7

## Visão Geral

Este projeto foi migrado de UUID v4 para UUID v7 para geração de chaves primárias. UUID v7 fornece benefícios significativos de desempenho enquanto mantém a unicidade global e capacidades de geração distribuída do UUID v4.

## Por Que UUID v7?

### Benefícios sobre UUID v4

1. **Ordenado por tempo**: UUIDs são naturalmente ordenados por tempo de criação
2. **Melhor desempenho B-tree**: Reduz divisões de página em índices PostgreSQL (INSERTs 20-50% mais rápidos em benchmarks)
3. **Fragmentação reduzida**: IDs sequenciais-ish minimizam fragmentação de índice
4. **Amigável ao cache**: Melhor localidade de referência para registros recém-criados
5. **Timestamp extraível**: Pode recuperar tempo de criação do próprio UUID
6. **Ainda globalmente único**: Mantém garantias de unicidade do UUID v4

### Comparação com outras estratégias

| Estratégia       | Prós                                                   | Contras                                           |
| ---------------- | ------------------------------------------------------ | ------------------------------------------------- |
| **UUID v7**      | [+] Ordenado por tempo, globalmente único, distribuído | Ligeiramente maior que BIGINT (16 bytes)          |
| UUID v4          | Globalmente único, distribuído                         | [X] Aleatório = localidade de índice ruim         |
| SERIAL/BIGSERIAL | Pequeno, sequencial, rápido                            | [X] Ponto único de falha, problemas de replicação |
| ULID             | Similar a v7, codificado em base32                     | Menos padrão, suporte limitado de bibliotecas     |
| Snowflake ID     | Rápido, ordenado por tempo                             | Requer serviço de coordenação                     |

## Implementação

### Configuração PostgreSQL

O projeto inclui duas migrações:

1. **000005_add_uuidv7_support.sql**: Adiciona a função `uuid_generate_v7()` ao PostgreSQL
2. **000006_switch_tables_to_uuidv7.sql**: Atualiza padrões de tabelas para usar v7

```sql
-- Gerar um UUID v7 no PostgreSQL
SELECT uuid_generate_v7();
-- Exemplo: 018d2f07-0f3c-7000-8000-123456789abc
```

### Código Go

Use o pacote `pkg/uuidv7`:

```go
import "github.com/basilex/promenade/pkg/uuidv7"

// Gerar novo UUID v7
id := uuidv7.New()

// Gerar com timestamp específico (útil para testes)
id := uuidv7.NewWithTime(time.Now())

// Extrair timestamp de criação do UUID
timestamp := uuidv7.ExtractTime(id)

// Verificar se UUID é versão 7
if uuidv7.IsV7(id) {
    // ...
}
```

### Caminho de Migração

Registros existentes com UUID v4 continuarão funcionando. O sistema suporta UUIDs mistos:

| Tipo de Registro  | Versão UUID | Notas        |
| ----------------- | ----------- | ------------ |
| Registros antigos | UUID v4     | Pré-migração |
| Registros novos   | UUID v7     | Pós-migração |

## Considerações de Desempenho

### Quando Usar UUID v7

[+] **Bom para:**

- Sistemas distribuídos onde múltiplos nós geram IDs
- Implantações multi-região
- Arquitetura de microserviços
- APIs onde clientes precisam gerar IDs
- Tabelas com alta taxa de INSERT
- Quando você precisa de ordenação natural baseada em tempo

[X] **Reconsidere se:**

- Sistema de banco de dados único sem necessidades de distribuição
- Armazenamento extremamente restrito (UUID = 16 bytes vs BIGINT = 8 bytes)
- Você precisa de análise de lacunas sequenciais (use SERIAL)

### Otimização de Índice PostgreSQL

UUID v7 funciona melhor com índices B-tree devido à sua natureza ordenada por tempo:

```sql
-- Bom: Chave primária usa B-tree automaticamente
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7()
);

-- Bom: Consultas baseadas em tempo se beneficiam da ordenação UUID v7
CREATE INDEX idx_users_created ON users(created_at DESC);

-- Ótimo: Combine UUID v7 com timestamp para consultas de intervalo
SELECT * FROM users
WHERE id >= uuid_generate_v7_from_timestamp('2024-01-01'::timestamp)
ORDER BY id DESC;
```

### Monitorando Desempenho

Compare desempenho de INSERT antes/depois da migração:

```sql
-- Verificar inchaço de índice
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE tablename IN ('users', 'products', 'roles')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Monitorar desempenho de INSERT
EXPLAIN ANALYZE
INSERT INTO users (email, name, password)
VALUES ('test@example.com', 'Test', 'hash');
```

## Recomendações para Chaves Estrangeiras

### Use UUID v7 para Chaves Estrangeiras Também

```go
type Product struct {
    ID        uuidv7.UUID `db:"id"`        // UUID v7 chave primária
    UserID    uuidv7.UUID `db:"user_id"`   // UUID v7 chave estrangeira
    Name      string      `db:"name"`
    CreatedAt time.Time   `db:"created_at"`
}
```

### Indexando Chaves Estrangeiras

Sempre indexe chaves estrangeiras para desempenho de JOIN:

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL
);

-- Crítico: Indexar a chave estrangeira
CREATE INDEX idx_products_user_id ON products(user_id);
```

### Chaves Compostas com UUID v7

Para relacionamentos muitos-para-muitos:

```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Índice para busca reversa
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

## Estratégias Alternativas (Consideração Futura)

### Opção 1: ULID (Universally Unique Lexicographically Sortable Identifier)

Similar a UUID v7 mas usa codificação base32 (representação de string mais curta):

```
01ARZ3NDEKTSV4RRFFQ69G5FAV  (ULID - 26 caracteres)
018d2f07-0f3c-7000-8000-123456789abc  (UUID v7 - 36 caracteres)
```

**Quando usar**: Se você frequentemente exibe IDs para usuários e quer strings mais curtas.

### Opção 2: Snowflake IDs

Snowflake do Twitter: inteiros de 64 bits com timestamp, ID de datacenter e sequência.

```go
// Exemplo de estrutura (não implementado neste projeto)
// 41 bits: timestamp
// 10 bits: datacenter/worker ID
// 12 bits: número de sequência
```

**Quando usar**: Sistemas de throughput extremamente alto (escala Twitter) onde coordenação é aceitável.

### Opção 3: PostgreSQL SERIAL/BIGSERIAL

Inteiros auto-incrementais tradicionais:

```sql
CREATE TABLE simple_table (
    id BIGSERIAL PRIMARY KEY,  -- 1, 2, 3, 4, ...
    name VARCHAR(255)
);
```

**Quando usar**: Aplicações simples de banco de dados único sem necessidades de distribuição.

### Opção 4: Chaves Naturais Compostas

Use dados de negócio significativos como chaves:

```sql
CREATE TABLE country_codes (
    code CHAR(2) PRIMARY KEY,  -- 'US', 'GB', 'FR'
    name VARCHAR(100)
);
```

**Quando usar**: Quando chaves naturais são estáveis, curtas e verdadeiramente únicas.

## Testes

Execute a suíte de testes para verificar implementação UUID v7:

```bash
# Testar pacote UUID v7
go test -v ./pkg/uuidv7/

# Executar com benchmarks
go test -bench=. ./pkg/uuidv7/

# Testar integração com banco de dados
make test-integration
```

Resultados esperados de benchmark (aproximados):

```
BenchmarkNew-8      3000000    450 ns/op    (UUID v7)
BenchmarkNewV4-8    2800000    480 ns/op    (UUID v4)
```

## Lista de Verificação de Migração

- [x] Adicionar função `uuid_generate_v7()` ao PostgreSQL
- [x] Atualizar padrões de tabelas para usar UUID v7
- [x] Atualizar código Go para usar pacote `pkg/uuidv7`
- [x] Atualizar todos os use cases (auth, product, role)
- [ ] Executar migrações: `make migrate-up`
- [ ] Testar geração de UUID: `go test ./pkg/uuidv7/`
- [ ] Monitorar métricas de produção após implantação

## Leitura Adicional

- [RFC 9562 - UUID Version 7 (Draft)](https://datatracker.ietf.org/doc/draft-ietf-uuidrev-rfc4122bis/)
- [PostgreSQL UUID Performance](https://www.2ndquadrant.com/en/blog/sequential-uuid-generators/)
- [UUID v7 in Production (Blog post)](https://buildkite.com/blog/goodbye-integers-hello-uuids)

## Perguntas?

Veja `.github/copilot-instructions.md` para convenções do projeto ou pergunte no chat da equipe.
