# Arquitetura de Configuração de Módulos

🇬🇧 [English](../MODULE_CONFIG_ARCHITECTURE.md) | [🇺🇦 Українська](../uk/MODULE_CONFIG_ARCHITECTURE.uk.md) | [🇩🇪 Deutsch](../de/MODULE_CONFIG_ARCHITECTURE.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/MODULE_CONFIG_ARCHITECTURE.es.md)

## Visão Geral

Os módulos no Promenade são **fatias verticais totalmente autônomas** que gerenciam sua própria configuração. Cada módulo carrega sua configuração de sua própria árvore de diretórios, garantindo independência completa do sistema central.

## Princípios de Arquitetura

### 1. Autonomia do Módulo

- Cada módulo é responsável por carregar sua própria configuração
- Módulos armazenam configurações em seu próprio diretório: `internal/modules/{nome_modulo}/config/`
- O núcleo não carrega nem gerencia configurações de módulos
- O núcleo fornece apenas infraestrutura (DB, EventBus, JWT) e contexto de ambiente

### 2. Suporte a Ambientes

- Cada módulo possui três arquivos de configuração específicos de ambiente:
  - `config.dev.yaml` - Configurações de desenvolvimento
  - `config.test.yaml` - Configurações de teste
  - `config.prod.yaml` - Configurações de produção
- O SDK do módulo carrega automaticamente a configuração correta com base na variável `ENVIRONMENT`

### 3. Fluxo de Carregamento de Configuração

```
Inicialização da Aplicação
    ↓
Núcleo carrega app.{env}.yaml
    ↓
Núcleo inicializa infraestrutura (DB, EventBus etc.)
    ↓
Registro de módulos descobre módulos
    ↓
Para cada módulo habilitado:
    Module.Initialize(ctx, core) é chamado
        ↓
    Módulo carrega internal/modules/{nome}/config/config.{env}.yaml
        ↓
    Módulo valida configurações (licença, recursos etc.)
        ↓
    Módulo inicializa repositórios, casos de uso, manipuladores
        ↓
    Module.RegisterRoutes() registra endpoints HTTP
```

## Estrutura de Diretórios

```
config/
├── app.dev.yaml          # Configuração do núcleo - desenvolvimento
├── app.test.yaml         # Configuração do núcleo - teste
├── app.prod.yaml         # Configuração do núcleo - produção
└── modules.yaml          # Registro de módulos (habilitado/desabilitado)

internal/modules/
├── posts/
│   ├── config/
│   │   ├── config.dev.yaml   # Configuração do módulo posts para dev
│   │   ├── config.test.yaml  # Configuração do módulo posts para teste
│   │   └── config.prod.yaml  # Configuração do módulo posts para prod
│   ├── entity/
│   ├── repository/
│   ├── usecase/
│   ├── handler/
│   └── module.go            # Carrega própria configuração em Initialize()
│
├── warehouse/
│   ├── config/
│   │   ├── config.dev.yaml   # Configuração do módulo warehouse para dev
│   │   ├── config.test.yaml  # Configuração do módulo warehouse para teste
│   │   └── config.prod.yaml  # Configuração do módulo warehouse para prod
│   └── module.go            # Carrega própria configuração em Initialize()
│
└── profiles/
    ├── config/
    │   ├── config.dev.yaml   # Configuração do módulo profiles para dev
    │   ├── config.test.yaml  # Configuração do módulo profiles para teste
    │   └── config.prod.yaml  # Configuração do módulo profiles para prod
    └── ...
```

## Escopos de Configuração

### Configuração do Núcleo (`config/app.{env}.yaml`)

**Gerenciado por**: `internal/infrastructure/config/yaml_config.go`

Contém:

- Metadados da aplicação (nome, versão, ambiente)
- Configurações do servidor (host, porta, timeouts)
- Conexão com banco de dados (host, porta, credenciais)
- Configurações JWT (segredo, expiração)
- Configuração de logging
- Configurações CORS
- Adaptador de barramento de eventos (memory/redis)
- Limitação de taxa
- Serviço de e-mail

**Nunca contém**: Configurações de lógica de negócios específicas de módulos

### Configuração do Módulo (`internal/modules/{nome}/config/config.{env}.yaml`)

**Gerenciado por**: Cada módulo usando `pkg/module/config`

Contém:

- Metadados do módulo (nome, versão, flag de habilitação)
- Configurações específicas do módulo
- Políticas de limpeza (dias de retenção, tamanhos de lote)
- Definições de permissões
- Flags de recursos
- Chaves de licença (para módulos comerciais)

**Nunca contém**: Configurações de infraestrutura central

## Exemplo de Implementação

### Carregamento de Configuração do Módulo Posts

```go
// internal/modules/posts/module.go
package posts

import (
    "context"
    "github.com/basilex/promenade/pkg/module"
    moduleconfig "github.com/basilex/promenade/pkg/module/config"
)

type PostsModule struct {
    *module.BaseModule
    config *moduleconfig.Config
    // ... outros campos
}

func (m *PostsModule) Initialize(ctx context.Context, core *module.Core) error {
    // Chamar implementação base
    if err := m.BaseModule.Initialize(ctx, core); err != nil {
        return err
    }

    // Carregar própria configuração do módulo de seu próprio diretório
    config, err := moduleconfig.Load("internal/modules/posts/config", core.Config.AppName)
    if err != nil {
        return fmt.Errorf("failed to load posts config: %w", err)
    }
    m.config = config

    // Usar configurações
    retentionDays := m.config.GetRetentionDays("posts", 30)
    batchSize := m.config.GetBatchSize("posts", 100)

    // ... inicializar repositórios, casos de uso, manipuladores

    return nil
}
```

### Estrutura do Arquivo de Configuração do Módulo

```yaml
# internal/modules/posts/config/config.dev.yaml
module:
  enabled: true
  name: "posts"
  version: "1.0.0"

settings:
  max_content_length: 10000
  allow_markdown: true

purge:
  enabled: true
  schedule: "0 2 * * *"
  settings:
    posts:
      retention_days: 30
      batch_size: 100
    comments:
      retention_days: 60
      batch_size: 50

permissions:
  - resource: "posts"
    actions: ["create", "read", "update", "delete"]
  - resource: "comments"
    actions: ["create", "read", "update", "delete"]

features:
  enable_likes: true
  enable_sharing: true
```

## SDK de Configuração do Módulo

### Carregar Configuração

```go
import moduleconfig "github.com/basilex/promenade/pkg/module/config"

// Carregar configuração para ambiente atual
config, err := moduleconfig.Load("internal/modules/{nome}/config", environment)
```

### Acessar Configurações

```go
// Obter dias de retenção para limpeza com fallback padrão
retentionDays := config.GetRetentionDays("nome_entidade", 30)

// Obter tamanho de lote para limpeza com fallback padrão
batchSize := config.GetBatchSize("nome_entidade", 100)

// Verificar se recurso está habilitado
enabled := config.GetFeature("enable_likes", true)

// Obter configuração aninhada
value := config.GetNestedSetting("settings", "max_content_length")
```

## Migração de Configurações Centralizadas

### Arquitetura Antiga (Obsoleta)

```
config/modules/
├── posts.dev.yaml          Centralizado
├── posts.test.yaml         Centralizado
├── posts.prod.yaml         Centralizado
└── ...

Núcleo carrega todas configurações de módulos   Acoplamento forte
Núcleo passa configurações para módulos   Dependência
```

### Nova Arquitetura (Atual)

```
internal/modules/posts/config/
├── config.dev.yaml         Propriedade do módulo
├── config.test.yaml        Propriedade do módulo
└── config.prod.yaml        Propriedade do módulo

Módulo carrega própria configuração   Autônomo
Módulo gerencia próprias configurações   Independente
```

## Benefícios

### 1. Verdadeira Independência de Módulo

- Módulos podem ser desenvolvidos, testados e implantados independentemente
- Não são necessárias alterações no núcleo ao adicionar/modificar configurações de módulos
- Módulos são verdadeiramente plugins autônomos

### 2. Melhor Encapsulamento

- Configuração está junto com o código que ela configura
- Propriedade e responsabilidade claras
- Mais fácil entender as capacidades do módulo

### 3. Núcleo Simplificado

- Núcleo gerencia apenas infraestrutura
- Sem lógica de negócios em configurações do núcleo
- Separação de responsabilidades mais clara

### 4. Testes Mais Fáceis

- Cada módulo pode ter diferentes configurações de teste
- Sem poluição de configuração global
- Testes de módulos são isolados

### 5. Implantação Flexível

- Habilitar/desabilitar módulos sem alterações no núcleo
- Diferentes ambientes podem ter diferentes configurações de módulos
- Módulos comerciais podem ter validação de licença

## Módulos Comerciais

Módulos comerciais (por exemplo, warehouse) requerem chaves de licença:

```yaml
# internal/modules/warehouse/config/config.prod.yaml
module:
  enabled: true
  name: "warehouse"
  version: "1.2.0"
  license_key: "sua-chave-de-licenca-comercial-aqui"

settings:
  max_items: 100000
  enable_barcode_scanner: true
```

O módulo valida a licença durante `Initialize()`:

```go
func (m *WarehouseModule) verifyLicense() error {
    licenseKey := m.config.GetNestedSetting("module", "license_key").(string)
    if licenseKey == "" {
        return fmt.Errorf("warehouse module requires license key")
    }
    // Validar licença...
    return nil
}
```

## Melhores Práticas

### FAÇA

- Armazene configurações de módulos em `internal/modules/{nome}/config/`
- Carregue configuração no método `Initialize()` do módulo
- Use arquivos de configuração específicos de ambiente
- Valide configurações críticas durante a inicialização
- Use métodos auxiliares de `pkg/module/config`
- Forneça valores padrão sensatos para configurações opcionais

### NÃO FAÇA

- Colocar configurações de módulos em `config/modules/` (obsoleto)
- Carregar configurações de módulos no núcleo
- Colocar configurações de módulos na configuração do núcleo
- Codificar valores específicos de ambiente
- Pular validação de configuração
- Acessar configurações de outros módulos

## Depuração

### Verificar qual configuração foi carregada:

```go
logger.FromContext(ctx).Info("Module config loaded",
    "module", m.GetMetadata().Name,
    "config_path", "internal/modules/posts/config",
    "environment", os.Getenv("ENVIRONMENT"),
)
```

### Verificar ambiente:

```bash
echo $ENVIRONMENT  # Deve ser: development, test ou production
```

### Falhas no carregamento de configuração:

- Verifique se o arquivo existe: `internal/modules/{nome}/config/config.{env}.yaml`
- Verifique a sintaxe YAML
- Verifique as permissões do arquivo
- Certifique-se de que o ambiente está configurado corretamente

## Documentação Relacionada

- [Guia de Desenvolvimento de Módulos](MODULE_DEVELOPMENT.pt.md)
- [Independência de Módulos](MODULE_INDEPENDENCE.pt.md)
- [Infraestrutura de Testes](TESTING_INFRASTRUCTURE.pt.md)
- [Migração de Configuração](CONFIG_MIGRATION.pt.md)
