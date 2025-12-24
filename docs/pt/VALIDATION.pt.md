[🇬🇧 English](../VALIDATION.md) | [🇺🇦 Українська](../uk/VALIDATION.uk.md) | [🇩🇪 Deutsch](../de/VALIDATION.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/VALIDATION.es.md)

# Estratégia de Validação

## Abordagem de Validação em Múltiplas Camadas

Promenade implementa **defesa em profundidade** através de validação em múltiplas camadas para garantir integridade e segurança dos dados.

### 1. Camada DTO (Entrada HTTP)

**Localização**: `internal/adapter/http/v*/dto/*.go`

**Propósito**: Primeira linha de defesa - valida dados de requisição HTTP antes de chegarem à lógica de negócio.

**Tecnologias**: Gin binding tags + go-playground/validator

**Exemplo**:

```go
type CreateCountryRequest struct {
    Name string `json:"name" binding:"required,min=2,max=100"`
    Code string `json:"code" binding:"required,min=2,max=10,uppercase"`
    ISO2 string `json:"iso2" binding:"required,len=2,alpha,uppercase"`
    ISO3 string `json:"iso3" binding:"required,len=3,alpha,uppercase"`
}
```

**Custom Validators** (`pkg/validator/custom_validators.go`):

- `uppercase` - Garante que string está em maiúsculas
- `alpha` - Apenas letras permitidas
- `alphanum` - Apenas letras e números
- `no_special` - Sem caracteres especiais

### 2. Camada Entity (Domain)

**Localização**: `internal/domain/entity/*.go`

**Propósito**: Validação de regras de negócio - garante que objetos de domínio estão sempre em estado válido.

**Método**: `Validate() error` em cada entity

**Exemplo**:

```go
func (c *Country) Validate() error {
    var errs []string

    // Name validation
    if strings.TrimSpace(c.Name) == "" {
        errs = append(errs, "name is required")
    } else if len(c.Name) < 2 || len(c.Name) > 100 {
        errs = append(errs, "name must be between 2 and 100 characters")
    }

    // Auto-normalize
    c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
    c.ISO2 = strings.ToUpper(strings.TrimSpace(c.ISO2))
    c.ISO3 = strings.ToUpper(strings.TrimSpace(c.ISO3))

    // ... more validation

    if len(errs) > 0 {
        return errors.New(strings.Join(errs, "; "))
    }
    return nil
}
```

**Recursos**:

- Auto-normalização (remover espaços, códigos em maiúsculas)
- Coleta de múltiplos erros
- Aplicação de regras de negócio

### 3. Camada Use Case

**Localização**: `internal/usecase/*.go`

**Propósito**: Validação de orquestração - garante que validação de entity é chamada antes da persistência.

**Exemplo**:

```go
func (uc *CountryUseCase) Create(ctx context.Context, country *entity.Country) error {
    // Validate entity
    if err := country.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    return uc.countryRepo.Create(ctx, country)
}
```

## Regras de Validação por Entity

### Country

| Campo | Regras DTO                         | Regras Entity                       | Notas           |
| ----- | ---------------------------------- | ----------------------------------- | --------------- |
| Name  | required, min=2, max=100           | required, 2-100 caracteres          | Aparado         |
| Code  | required, min=2, max=10, uppercase | required, 2-10 caracteres, alphanum | Auto-maiúsculas |
| ISO2  | required, len=2, alpha, uppercase  | required, 2 caracteres, alpha       | Auto-maiúsculas |
| ISO3  | required, len=3, alpha, uppercase  | required, 3 caracteres, alpha       | Auto-maiúsculas |

### Currency

| Campo  | Regras DTO                                   | Regras Entity                       | Notas                     |
| ------ | -------------------------------------------- | ----------------------------------- | ------------------------- |
| Name   | required, min=2, max=100                     | required, 2-100 caracteres          | Aparado                   |
| Code   | required, min=3, max=10, alphanum, uppercase | required, 3-10 caracteres, alphanum | Auto-maiúsculas, ISO 4217 |
| Symbol | omitempty, max=10                            | opcional, max 10 caracteres         | Opcional                  |

## Tratamento de Erros

### Erros de Validação DTO

**HTTP 400 Bad Request** com erros detalhados de campos:

```json
{
  "success": false,
  "message": "invalid request",
  "error": "Key: 'CreateCountryRequest.ISO2' Error:Field validation for 'ISO2' failed on the 'len' tag"
}
```

### Erros de Validação Entity

**HTTP 500 Internal Server Error** (deve ser capturado antes pela validação DTO):

```json
{
  "success": false,
  "message": "validation failed",
  "error": "name is required; iso2 must be exactly 2 characters"
}
```

## Testes

Executar testes de validação:

```bash
# Entity validation tests
go test ./internal/domain/entity -run TestCountry_Validate
go test ./internal/domain/entity -run TestCurrency_Validate

# All tests
make test
```

## Melhores Práticas

1. **Sempre valide primeiro na camada DTO** - impede que dados ruins entrem no sistema
2. **Use validação de entity como rede de segurança** - regras de negócio aplicadas mesmo se DTO for contornado
3. **Auto-normalização na camada entity** - códigos em maiúsculas, remover espaços
4. **Colete múltiplos erros** - melhor UX do que falhar no primeiro erro
5. **Teste validação minuciosamente** - testes unitários para casos válidos e inválidos

## Adicionando Novos Validadores

### Custom Gin Validator

1. Adicione a `pkg/validator/custom_validators.go`:

```go
func validateMyRule(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    // Your validation logic
    return true
}
```

2. Registre em `RegisterCustomValidators()`:

```go
if err := v.RegisterValidation("myrule", validateMyRule); err != nil {
    return err
}
```

3. Use em DTOs:

```go
Field string `json:"field" binding:"myrule"`
```

### Validação Entity

Adicione ao método `Validate()` da entity:

```go
if !isValidMyRule(c.Field) {
    errs = append(errs, "field must meet my rule")
}
```

## Considerações de Segurança

- **SQL Injection**: Prevenido por queries parametrizadas (sqlx)
- **XSS**: Prevenido por codificação JSON + Content-Type headers
- **Integridade de Dados**: Validação em múltiplas camadas garante ausência de dados incorretos
- **Code Injection**: Validadores alpha/alphanum previnem caracteres especiais
- **Length Attacks**: Comprimento máximo aplicado em todas as camadas

## Performance

- Validação DTO: ~1-5μs por requisição (insignificante)
- Validação Entity: ~100-500ns por entity (insignificante)
- Overhead total: <0.1% do tempo típico de requisição

## Referências

- [go-playground/validator docs](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [Gin validation](https://gin-gonic.com/docs/examples/binding-and-validation/)
- [Clean Architecture validation patterns](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
