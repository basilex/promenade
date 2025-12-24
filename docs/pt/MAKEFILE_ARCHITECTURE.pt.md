[🇬🇧 English](../MAKEFILE_ARCHITECTURE.md) | [🇺🇦 Українська](../uk/MAKEFILE_ARCHITECTURE.uk.md) | [🇩🇪 Deutsch](../de/MAKEFILE_ARCHITECTURE.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](MAKEFILE_ARCHITECTURE.es.md)

---

# Arquitetura do Makefile

Sistema modular de Makefile para separação limpa de responsabilidades e escalabilidade.

## Estrutura

```
Makefile             (63 linhas)  - Arquivo principal: variáveis, carregamento env, help
Makefile.dev.mk      (64 linhas)  - Fluxo de trabalho de desenvolvimento
Makefile.test.mk     (57 linhas)  - Infraestrutura de testes
Makefile.prod.mk     (90 linhas)  - Operações de produção/DevOps

Total:              274 linhas
```

## Filosofia

**Makefile Principal** - Apenas comum:

- Carregamento de variáveis de ambiente (`.env.development`)
- Variáveis compartilhadas (`APP_NAME`, `VERSION`, `DB_URL`, `MIGRATE`)
- Inclusão de módulos (`include Makefile.*.mk`)
- Comando help agrupado (mostra todos os módulos)

**Makefile.dev.mk** - Fluxo de trabalho do desenvolvedor:

```bash
make install             # Instalar ferramentas (swag, migrate, golangci-lint)
make dev                 # Iniciar servidor dev (postgres + migrations + app)
make build               # Construir binário (com geração de swagger)
make run                 # Executar binário compilado
make lint                # Executar golangci-lint
make fmt                 # Formatar código (go fmt + gofmt -s)
make deps-update         # Atualizar dependências
make config-show         # Mostrar configuração env atual
```

**Makefile.test.mk** - Testes:

```bash
make test                # Testes unit + integration
make test-unit           # Apenas testes unit (domain + usecase)
make test-integration    # Testes integration (DB real na porta 5433)
make test-smoke          # Testes smoke (fluxos críticos end-to-end)
make test-coverage       # Gerar relatório HTML de cobertura
make test-db-start       # Iniciar banco de dados de teste
make test-db-stop        # Parar banco de dados de teste
```

**Makefile.prod.mk** - Operações DevOps:

```bash
# Docker
make docker-build        # Construir imagem (VERSION=0.1.0 ENV=dev)
make docker-run          # Construir + executar containers
make docker-up           # Iniciar serviços
make docker-down         # Parar serviços
make docker-logs         # Ver logs
make docker-restart      # Reiniciar containers
make docker-ps           # Mostrar containers em execução
make docker-clean        # Remover containers + volumes

# Migrations
make migrate-create      # Criar migration (NAME=xxx)
make migrate-up          # Aplicar migrations
make migrate-down        # Reverter última migration
make migrate-force       # Forçar versão (VERSION=N)
make migrate-version     # Mostrar versão atual
make migrate-status      # Mostrar status

# Documentação
make swagger-all         # Gerar documentação Swagger v1 + v2

# Limpeza
make clean               # Remover artefatos (bin/, docs/, coverage)
```

## Benefícios

1. **Modularidade** - Cada arquivo tem responsabilidade única
2. **Escalabilidade** - Fácil adicionar `Makefile.{stage,ci,deploy}.mk`
3. **Legibilidade** - Separação clara por contexto
4. **Manutenibilidade** - Arquivos pequenos focados vs monolito de 188 linhas
5. **Amigável para equipe** - Desenvolvedores/QA/DevOps veem apenas comandos relevantes

## Exemplos de Uso

**Desenvolvimento:**

```bash
make help          # Ver todos os comandos disponíveis
make dev           # Iniciar desenvolvimento (mais comum)
make build         # Construir para testes locais
make fmt lint      # Formatar e lint antes do commit
```

**Testes:**

```bash
make test          # Executar suite completa de testes antes do PR
make test-unit     # Feedback rápido durante desenvolvimento
make test-smoke    # Verificar fluxos críticos após mudanças
```

**DevOps:**

```bash
make docker-run    # Implantar no Docker local
make migrate-up    # Aplicar migrations do banco de dados
make swagger-all   # Regerar documentação da API
make clean         # Limpar antes de implantação nova
```

## Adicionando Novos Comandos

1. Identificar contexto: dev/test/prod
2. Editar `Makefile.{context}.mk` apropriado
3. Adicionar `## Comentário` para exibição no help
4. Executar `make help` para verificar

Exemplo:

```makefile
# Em Makefile.dev.mk
watch: ## Observar e recarregar em mudanças de arquivo
	air -c .air.toml
```

## Variáveis

Todas as variáveis compartilhadas estão no `Makefile` principal:

- `APP_NAME` - Nome da aplicação
- `VERSION` - Versão da build (padrão: 0.1.0)
- `ENV` - Ambiente (dev/test/prod)
- `DB_URL` - String de conexão PostgreSQL
- `MIGRATE` - Comando migrate com DB URL
- `DOCKER_COMPOSE` - Comando Docker Compose

Sobrescrever com:

```bash
make docker-build VERSION=1.2.3 ENV=prod
make migrate-up DB_NAME=promenade_staging
```

## Migração da Estrutura Antiga

Antes (monolito de 188 linhas):

```
Makefile  ← Tudo misturado junto
```

Depois (274 linhas modular):

```
Makefile          ← Comum (63 linhas)
Makefile.dev.mk   ← Desenvolvimento (64 linhas)
Makefile.test.mk  ← Testes (57 linhas)
Makefile.prod.mk  ← Produção (90 linhas)
```

**Sem breaking changes** - Todos os comandos funcionam exatamente como antes!
