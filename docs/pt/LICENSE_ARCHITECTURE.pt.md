# Arquitetura do Sistema de Licenças

Promenade usa um sistema de licenciamento baseado em assinatura para módulos comerciais. Este documento descreve a arquitetura, implementação e padrões de uso.

---

## Visão Geral

### Princípios de Design

1. **Independência de Módulo**: Cada módulo gerencia seu próprio licenciamento
2. **Segurança Baseada em Assinatura**: HMAC-SHA256 previne adulteração
3. **Degradação Gradual**: Período de graça para licenças expiradas
4. **Específico por Ambiente**: Regras de validação diferentes por ambiente
5. **Legível para Humanos**: Chaves de licença são estruturadas e parseáveis

### Formato de Licença

```
PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}
```

Exemplo:

```
PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Componentes:

- **Prefixo**: Sempre `PROMENADE`
- **Módulo**: Nome do módulo em MAIÚSCULAS (ANALYTICS, WAREHOUSE, AUDITLOG)
- **Nível**: Nível da licença (BASIC, PRO, ENTERPRISE)
- **Expiração**: Data no formato YYYYMMDD
- **Assinatura**: HMAC-SHA256 das primeiras 4 partes, codificada em base64 URL

---

## Arquitetura

### Componentes do Sistema

```
─────────────────────────────────────────────────────────────
│                    Inicialização da Aplicação                │
└────────────────────────────────────────────────────────────
                       │
                       ▼
         ─────────────────────────
         │  Registro de Módulos    │
         │  (pkg/module)           │
         └────────────────────────
                  │
                  │ Para cada módulo habilitado
                  ▼
         ─────────────────────────
         │  Module.Initialize()    │
         └────────────────────────
                  │
                  ▼
         ─────────────────────────
         │  Validador de Licença   │
         │  (module/license/)      │
         └────────────────────────
                  │
    ──────────────────────────
    │             │             │
    ▼             ▼             ▼
Parse()       Validate()    HealthCheck()
    │             │             │
    └──────────────────────────
                  │
                  ▼
         ─────────────────────────
         │  Status da Licença      │
         │  - Válida               │
         │  - Expirada (Graça)     │
         │  - Inválida             │
         └─────────────────────────
```

### Fluxo de Validação

1. **Parsear**: Extrair componentes da string de licença
2. **Verificar Módulo**: Garantir que a licença corresponde ao nome do módulo
3. **Verificar Assinatura**: Validação HMAC-SHA256
4. **Verificar Expiração**: Validar data + período de graça
5. **Retornar Status**: Válida, expirada (graça) ou inválida

### Armazenamento

Licenças são armazenadas como variáveis de ambiente:

- `{MODULE}_LICENSE_KEY`: A chave de licença
- `LICENSE_SECRET`: Segredo para verificação de assinatura (produção)

Exemplo:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret-key"
```

---

## Implementação

### Integração de Módulo

Cada módulo comercial implementa validação de licença em seu método `Initialize()`:

```go
func (m *AnalyticsModule) Initialize(cfg interface{}) error {
    // Carregar configuração do módulo
    config, ok := cfg.(*Config)
    if !ok {
        return ErrInvalidConfig
    }

    // Validar licença se necessário
    if config.LicenseRequired {
        if err := m.validateLicense(config); err != nil {
            return fmt.Errorf("license validation failed: %w", err)
        }
    }

    return nil
}

func (m *AnalyticsModule) validateLicense(config *Config) error {
    licenseKey := config.LicenseKey
    if licenseKey == "" {
        licenseKey = os.Getenv("ANALYTICS_LICENSE_KEY")
    }

    if licenseKey == "" {
        return ErrLicenseRequired
    }

    secret := os.Getenv("LICENSE_SECRET")
    if secret == "" {
        secret = "default-dev-secret"
    }

    license, err := ParseLicense(licenseKey)
    if err != nil {
        return err
    }

    validationOptions := ValidationOptions{
        ValidateExpiry:    config.ValidateExpiry,
        ValidateSignature: config.ValidateSignature,
        GracePeriodDays:   config.GracePeriodDays,
    }

    return license.Validate("ANALYTICS", secret, validationOptions)
}
```

### Validação de Licença

O pacote license fornece a lógica central de validação:

```go
// Parse extrai componentes da string de licença
func ParseLicense(licenseKey string) (*License, error)

// Validate verifica a validade da licença
func (l *License) Validate(
    expectedModule string,
    secret string,
    options ValidationOptions,
) error

// Generate cria uma nova licença (para testes/ferramentas)
func GenerateLicense(
    module, tier, expiry, secret string,
) (string, error)
```

### Verificações de Saúde

A verificação de saúde de cada módulo inclui o status da licença:

```go
func (m *AnalyticsModule) HealthCheck(ctx context.Context) error {
    if m.license != nil {
        if m.license.IsExpired() && !m.license.InGracePeriod(7) {
            return ErrLicenseExpired
        }
    }
    return nil
}
```

---

## Configuração

### Configurações Específicas por Ambiente

#### Desenvolvimento (`config.dev.yaml`)

```yaml
license_required: false # Opcional em desenvolvimento
validate_expiry: false # Ignorar expiração
validate_signature: true # Ainda verificar assinaturas
grace_period_days: 30 # Período de graça longo
```

#### Teste (`config.test.yaml`)

```yaml
license_required: false # Sem licença no CI/CD
validate_expiry: false
validate_signature: false
grace_period_days: 0
```

#### Produção (`config.prod.yaml`)

```yaml
license_required: true # Obrigatória
validate_expiry: true # Expiração rigorosa
validate_signature: true # Segurança completa
grace_period_days: 7 # Graça limitada
validate_on_request: true # Validação opcional por requisição
```

### Configuração de Módulo

Em `config/modules.yaml`:

```yaml
modules:
  analytics:
    enabled: true
    version: "1.0.0"
    description: "Advanced analytics and reporting"
    license_key: "" # Definir via variável de ambiente
    settings:
      metrics_retention_days: 90
      max_reports_per_user: 10
      max_dashboards_per_user: 5
```

---

## Níveis e Recursos

### Nível BASIC

- Coleta básica de métricas (100 métricas/dia)
- Relatórios básicos (JSON, CSV)
- 5 dashboards por usuário
- 30 dias de retenção de dados
- Suporte por email

### Nível PRO

- Métricas ilimitadas
- Relatórios avançados (PDF, XLSX)
- Dashboards ilimitados
- 90 dias de retenção de dados
- Relatórios agendados
- Suporte prioritário

### Nível ENTERPRISE

- Todos os recursos PRO
- Retenção personalizada (1+ anos)
- Suporte multi-tenant
- Marca personalizada
- Acesso à API
- Suporte dedicado
- Implantação on-premise

---

## Uso

### Gerando Licenças

Use o script fornecido:

```bash
# Gerar licença PRO válida por 365 dias
./scripts/generate-license.sh analytics PRO 365

# Saída:
# PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Definir variável de ambiente:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret"
```

### Testando Validação de Licença

O módulo analytics inclui testes abrangentes:

```bash
# Executar testes de validação de licença
go test ./internal/modules/analytics/license/... -v

# Cenários de teste:
# - Parsear licença válida
# - Detectar formato inválido
# - Verificação de assinatura
# - Tratamento de expiração
# - Lógica de período de graça
# - Detecção de incompatibilidade de módulo
```

### Ferramenta de Geração de Licença

Construir o gerador de licenças:

```bash
make build-license-generator

# Ou manualmente:
go build -o ./bin/license-generator ./cmd/license-generator/main.go
```

Gerar licença programaticamente:

```bash
./bin/license-generator \
  -module=ANALYTICS \
  -tier=PRO \
  -expiry=20261231 \
  -secret="your-secret-key"
```

---

## Segurança

### Assinaturas HMAC-SHA256

- **Algoritmo**: HMAC com SHA-256
- **Chave**: Armazenada na variável de ambiente `LICENSE_SECRET`
- **Codificação**: Codificação base64 URL (URL-safe, sem padding)
- **Verificação**: Comparação de tempo constante para prevenir ataques de temporização

### Gerenciamento de Segredos

**Desenvolvimento**:

```bash
# Segredo de teste padrão (aceitável para desenvolvimento local)
LICENSE_SECRET="default-dev-secret-change-in-production"
```

**Produção**:

```bash
# Gerar segredo forte (32+ bytes)
openssl rand -base64 32

# Armazenar em vault seguro (AWS Secrets Manager, HashiCorp Vault, etc.)
# Definir via variável de ambiente (nunca commitar no git)
```

### Estratégia de Rotação

1. Gerar novo segredo
2. Assinar novas licenças com novo segredo
3. Suportar ambos os segredos antigo e novo durante transição
4. Depreciar segredo antigo após período de graça

---

## Tratamento de Erros

### Tipos de Erro

```go
var (
    ErrLicenseRequired   = errors.New("license key is required")
    ErrInvalidFormat     = errors.New("invalid license format")
    ErrInvalidSignature  = errors.New("invalid license signature")
    ErrLicenseExpired    = errors.New("license has expired")
    ErrModuleMismatch    = errors.New("license module mismatch")
    ErrInvalidTier       = errors.New("invalid license tier")
)
```

### Degradação Gradual

1. **Licença Ausente** (dev/test): Módulo carrega com recursos reduzidos
2. **Licença Expirada**: Período de graça permite operação contínua com avisos
3. **Licença Inválida**: Inicialização do módulo falha em produção, loga em dev

### Mensagens ao Usuário

```go
// Erro de produção (rigoroso)
return fmt.Errorf("analytics module requires valid license: %w", err)

// Aviso de desenvolvimento (permissivo)
logger.Warn("Analytics license validation failed, continuing in dev mode",
    "error", err)
```

---

## Estratégia de Monetização

### Estado Atual

| Módulo        | Status           | Nível              | Motivo                            |
| ------------- | ---------------- | ------------------ | --------------------------------- |
| posts         |  Gratuito      | N/A                | Recursos sociais principais       |
| profiles      |  Gratuito      | N/A                | Recursos principais de usuário    |
| **analytics** | ** Comercial** | **PRO/ENTERPRISE** | **Ativo: Analytics como premium** |
| warehouse     | 🔮 Planejado     | TBD                | Futuro: Gerenciamento de estoque  |

### Roteiro

**Fase 1 (Atual - Ativo)**:

- Módulo Analytics é **comercial** (níveis PRO/ENTERPRISE)
- Foco em métricas de negócio, relatórios, dashboards
- Alvo: PMEs, empresas com necessidade de insights de dados
- Status:  Implementado e ativado

**Fase 2 (Q2 2026 - Planejado)**:

- Lançamento do **Módulo Audit Log** (comercial)
- Recursos: Conformidade, GDPR, trilhas de auditoria detalhadas
- Alvo: Indústrias regulamentadas, empresas

**Fase 3 (Após Audit Log)**:

- **Migrar Analytics para nível gratuito**
- Analytics torna-se acessível para todos os usuários
- Audit Log permanece comercial para necessidades de conformidade

### Justificativa

1. **Analytics Primeiro**:  Funciona com dados existentes, valor imediato
2. **Audit Log Premium**: Conformidade é requisito empresarial
3. **Analytics Gratuito**: Adoção mais ampla, upsell para Audit Log

---

## Monitoramento e Observabilidade

### Métricas de Licença

Rastrear uso de licença através de logs:

```go
logger.Info("License validated",
    "module", "analytics",
    "tier", license.Tier,
    "expiry", license.ExpiryDate,
    "days_remaining", daysRemaining)
```

### Endpoint de Saúde

Status da licença disponível via verificação de saúde:

```bash
curl http://localhost:8080/api/v1/analytics/health

# Resposta:
{
  "status": "healthy",
  "license": {
    "tier": "PRO",
    "expiry": "2026-12-31",
    "days_remaining": 365,
    "in_grace_period": false
  }
}
```

### Alertas

Configurar monitoramento para:

- Licenças expiradas (período de graça terminando)
- Tentativas de licença inválida
- Erros de verificação de licença

---

## Testes

### Testes Unitários

Cada módulo inclui testes de licença abrangentes:

```bash
# Executar todos os testes de licença
go test ./internal/modules/*/license/... -v

# Executar testes de módulo específico
go test ./internal/modules/analytics/license/... -v
```

Cobertura de testes:

-  Parsear licença válida
-  Detectar formato inválido
-  Verificação de assinatura
-  Tratamento de expiração
-  Lógica de período de graça
-  Incompatibilidade de módulo
-  Validação de nível

### Testes de Integração

Testar inicialização de módulo com diferentes estados de licença:

```go
func TestModuleInitialization(t *testing.T) {
    tests := []struct {
        name        string
        licenseKey  string
        expectError bool
    }{
        {"valid license", validLicense, false},
        {"expired license", expiredLicense, true},
        {"invalid signature", tamperedLicense, true},
        {"missing license", "", true},
    }
    // ...
}
```

---

## Solução de Problemas

### Problemas Comuns

**1. Erro de Licença Necessária**

```
Error: license validation failed: license key is required
```

Solução:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
```

**2. Assinatura Inválida**

```
Error: invalid license signature
```

Causas:

- `LICENSE_SECRET` incorreto
- Chave de licença adulterada
- Chave copiada incorretamente

Solução: Regerar licença com segredo correto.

**3. Licença Expirada**

```
Warning: License expired 3 days ago, grace period active
```

Solução: Renovar licença antes do fim do período de graça (padrão 7 dias).

**4. Incompatibilidade de Módulo**

```
Error: license module mismatch: expected ANALYTICS, got WAREHOUSE
```

Solução: Usar licença correta para cada módulo.

### Modo de Depuração

Habilitar log verboso de licença:

```yaml
# config/app.dev.yaml
log:
  level: debug

modules:
  analytics:
    settings:
      debug_license: true # Logar todas as etapas de validação
```

### Ferramenta de Validação de Licença

Testar validade de licença manualmente:

```bash
# Validar licença
go run ./cmd/license-generator/main.go \
  -validate \
  -key="PROMENADE-ANALYTICS-PRO-20261231-..." \
  -secret="your-secret-key"

# Saída:
#  License valid
# Module: ANALYTICS
# Tier: PRO
# Expiry: 2026-12-31
# Days remaining: 365
```

---

## Melhores Práticas

### Para Desenvolvedores

1. **Nunca Commitar Segredos**: Usar arquivos `.env` (gitignored)
2. **Testar com Licenças Inválidas**: Garantir tratamento de erros
3. **Documentar Recursos por Nível**: Matriz clara de recursos por nível
4. **Degradação Gradual**: Não crashar em problemas de licença em dev
5. **Log Estruturado**: Logar eventos de licença para monitoramento

### Para Operadores

1. **Rotacionar Segredos Regularmente**: Atualizar `LICENSE_SECRET` trimestralmente
2. **Monitorar Datas de Expiração**: Alertar 30 dias antes da expiração
3. **Armazenamento Seguro de Segredos**: Usar soluções de vault (AWS, HashiCorp)
4. **Rastrear Uso de Licença**: Monitorar quais níveis estão ativos
5. **Planejar Renovações**: Renovar antes do fim do período de graça

### Para Autores de Módulos

1. **Seguir Padrões**: Usar módulo analytics existente como template
2. **Documentar Recursos**: Documentação clara de recursos baseados em nível
3. **Testar Completamente**: Testes abrangentes de validação de licença
4. **Verificações de Saúde**: Incluir status de licença em endpoints de saúde
5. **Mensagens de Erro**: Mensagens amigáveis e acionáveis ao usuário

---

## Referências

- [README do Módulo Analytics](../internal/modules/analytics/README.pt.md)
- [Pacote License](../internal/modules/analytics/license/)
- [Guia de Desenvolvimento de Módulos](./MODULE_DEVELOPMENT.pt.md)
- [Arquitetura de Configuração](./MODULE_CONFIG_ARCHITECTURE.pt.md)

---

## Apêndice

### Especificação de Formato de Licença

```
Formato: PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}

Restrições:
- MODULE: [A-Z0-9]+ (maiúsculas, sem espaços)
- TIER: BASIC|PRO|ENTERPRISE
- EXPIRY: YYYYMMDD (data válida)
- SIGNATURE: [A-Za-z0-9_-]+ (base64 URL-encoded)

Comprimento máximo: 256 caracteres
Comprimento mínimo: 50 caracteres
```

### Implementação HMAC-SHA256

```go
func generateSignature(data, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(data))
    signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
    return strings.TrimRight(signature, "=")  // Remover padding
}

func verifySignature(data, signature, secret string) bool {
    expected := generateSignature(data, secret)
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Checklist de Testes

- [ ] Parsear licença válida
- [ ] Rejeitar formato inválido
- [ ] Verificar assinatura
- [ ] Verificar expiração
- [ ] Testar período de graça
- [ ] Detectar incompatibilidade de módulo
- [ ] Validar níveis
- [ ] Testar licença ausente
- [ ] Testar licença adulterada
- [ ] Integração com módulo
- [ ] Verificação de saúde inclui status
- [ ] Mensagens de erro claras

---

**Última Atualização**: 2024-12-19
**Versão**: 1.0.0
**Status**: Pronto para Produção
