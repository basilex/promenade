# Verificação de Independência de Módulos

[🇬🇧 English](MODULE_INDEPENDENCE.pt.md) | [🇺🇦 Українська](../uk/MODULE_INDEPENDENCE.uk.md) | [🇩🇪 Deutsch](../de/MODULE_INDEPENDENCE.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/MODULE_INDEPENDENCE.es.md)

## Estrutura do Módulo Posts

O módulo `posts` demonstra independência completa do sistema principal, implementando sua própria pilha completa de Clean Architecture:

```
internal/modules/posts/
├── domain/
│   ├── entity/          # Entidade Post + erros de domínio
│   └── repository/      # Interface de repositório
├── usecase/             # Camada de lógica de negócios
├── adapter/
│   ├── http/           # Handlers HTTP e DTOs
│   └── repository/     # Implementação de banco de dados
├── tests/              # Testes específicos do módulo
├── module.go           # Integração do módulo
└── register.go         # Auto-registro
```

## Análise de Independência

### Sem Dependências do Core

Verificado por busca: `grep -r "github.com/basilex/promenade/internal/(domain|usecase|adapter)" internal/modules/posts/`

**Resultado**: ZERO correspondências - módulo não importa nenhum pacote interno do core.

### Apenas Dependências Compartilhadas

O módulo importa apenas:

- `pkg/*` - Utilitários compartilhados (logger, uuidv7, pagination, bus, response, module SDK)
- `github.com/gin-gonic/gin` - Framework HTTP
- `github.com/jmoiron/sqlx` - Biblioteca de banco de dados
- Pacotes da biblioteca padrão

### Fatia Vertical Completa

Cada camada implementada dentro do módulo:

| Camada                     | Localização                            | Dependências                |
| -------------------------- | -------------------------------------- | --------------------------- |
| **Entidade de Domínio**    | `domain/entity/post.go`                | Apenas `pkg/uuidv7`         |
| **Repositório de Domínio** | `domain/repository/post_repository.go` | Entidade de domínio, pkg    |
| **Caso de Uso**            | `usecase/post_usecase.go`              | Apenas Domain               |
| **Impl. Repositório**      | `adapter/repository/postgres/`         | Domain, pkg/database        |
| **Handler HTTP**           | `adapter/http/handler/`                | Use case, DTO, pkg/response |
| **DTOs**                   | `adapter/http/dto/`                    | Entidade de domínio         |

### Benefícios do Isolamento do Módulo

1. **Desenvolvimento Independente**: Pode ser desenvolvido/testado isoladamente
2. **Reutilizabilidade**: Pode ser copiado para outro projeto com pkg/
3. **Sem Mudanças Críticas**: Refatoração do core não afeta o módulo
4. **Limites Claros**: Todas as dependências explícitas e mínimas
5. **Testes Fáceis**: Mockar apenas interfaces de domínio, não serviços core

## Construindo Novos Módulos

Para criar um novo módulo independente:

1. Criar estrutura de diretórios:

   ```
   internal/modules/{name}/
   ├── domain/entity/
   ├── domain/repository/
   ├── usecase/
   ├── adapter/http/handler/
   ├── adapter/http/dto/
   ├── adapter/repository/postgres/
   ├── tests/
   ├── module.go
   └── register.go
   ```

2. Copiar implementação do core ou escrever do zero
3. Atualizar todos os imports para apontar para caminhos do módulo
4. Definir erros específicos do módulo em `domain/entity/errors.go`
5. Implementar interface `module.Module` em `module.go`
6. Auto-registrar em `register.go` usando `init()`

## Comando de Verificação

```bash
# Verificar qualquer dependência interna do core
grep -r "github.com/basilex/promenade/internal/\(domain\|usecase\|adapter\)" \
  internal/modules/posts/ || echo "✅ Módulo é independente"
```

## Status Atual

- **posts**: Completamente independente, Clean Architecture completa
- **comments**: Precisa de refatoração (atualmente wrapper)
- **warehouse**: Precisa de refatoração (atualmente wrapper)

## Próximos Passos

1. Refatorar módulo `comments` seguindo padrão `posts`
2. Refatorar módulo `warehouse`
3. Adicionar testes de integração aos diretórios `tests/`
4. Documentar APIs específicas do módulo
5. Criar guia de desenvolvimento de módulos com templates
