[🇬🇧 English](../VALIDATION.md) | [🇺🇦 Українська](../uk/VALIDATION.uk.md) | [🇩🇪 Deutsch](../de/VALIDATION.de.md) | [🇵🇹 Português](../pt/VALIDATION.pt.md) | 🇪🇸 **Español**

# Estrategia de Validación

## Enfoque de Validación Multicapa

Promenade implementa **defensa en profundidad** a través de validación en múltiples capas para garantizar integridad y seguridad de datos.

### 1. Capa DTO (Entrada HTTP)

**Ubicación**: `internal/adapter/http/v*/dto/*.go`

**Propósito**: Primera línea de defensa - valida datos de solicitud HTTP antes de que lleguen a la lógica de negocio.

**Tecnologías**: Gin binding tags + go-playground/validator

**Ejemplo**:

```go
type CreateCountryRequest struct {
    Name string `json:"name" binding:"required,min=2,max=100"`
    Code string `json:"code" binding:"required,min=2,max=10,uppercase"`
    ISO2 string `json:"iso2" binding:"required,len=2,alpha,uppercase"`
    ISO3 string `json:"iso3" binding:"required,len=3,alpha,uppercase"`
}
```

**Custom Validators** (`pkg/validator/custom_validators.go`):

- `uppercase` - Asegura que el string está en mayúsculas
- `alpha` - Solo letras permitidas
- `alphanum` - Solo letras y números
- `no_special` - Sin caracteres especiales

### 2. Capa Entity (Domain)

**Ubicación**: `internal/domain/entity/*.go`

**Propósito**: Validación de reglas de negocio - asegura que los objetos de dominio siempre están en estado válido.

**Método**: `Validate() error` en cada entity

**Ejemplo**:

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

**Características**:

- Auto-normalización (eliminar espacios, códigos en mayúsculas)
- Recopilación de múltiples errores
- Aplicación de reglas de negocio

### 3. Capa Use Case

**Ubicación**: `internal/usecase/*.go`

**Propósito**: Validación de orquestación - asegura que la validación de entity se llama antes de la persistencia.

**Ejemplo**:

```go
func (uc *ICountryUseCase) Create(ctx context.Context, country *entity.Country) error {
    // Validate entity
    if err := country.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    return uc.countryRepo.Create(ctx, country)
}
```

## Reglas de Validación por Entity

### Country

| Campo | Reglas DTO                         | Reglas Entity                       | Notas           |
| ----- | ---------------------------------- | ----------------------------------- | --------------- |
| Name  | required, min=2, max=100           | required, 2-100 caracteres          | Recortado       |
| Code  | required, min=2, max=10, uppercase | required, 2-10 caracteres, alphanum | Auto-mayúsculas |
| ISO2  | required, len=2, alpha, uppercase  | required, 2 caracteres, alpha       | Auto-mayúsculas |
| ISO3  | required, len=3, alpha, uppercase  | required, 3 caracteres, alpha       | Auto-mayúsculas |

### Currency

| Campo  | Reglas DTO                                   | Reglas Entity                       | Notas                     |
| ------ | -------------------------------------------- | ----------------------------------- | ------------------------- |
| Name   | required, min=2, max=100                     | required, 2-100 caracteres          | Recortado                 |
| Code   | required, min=3, max=10, alphanum, uppercase | required, 3-10 caracteres, alphanum | Auto-mayúsculas, ISO 4217 |
| Symbol | omitempty, max=10                            | opcional, max 10 caracteres         | Opcional                  |

## Manejo de Errores

### Errores de Validación DTO

**HTTP 400 Bad Request** con errores detallados de campos:

```json
{
  "success": false,
  "message": "invalid request",
  "error": "Key: 'CreateCountryRequest.ISO2' Error:Field validation for 'ISO2' failed on the 'len' tag"
}
```

### Errores de Validación Entity

**HTTP 500 Internal Server Error** (debe ser capturado antes por validación DTO):

```json
{
  "success": false,
  "message": "validation failed",
  "error": "name is required; iso2 must be exactly 2 characters"
}
```

## Pruebas

Ejecutar pruebas de validación:

```bash
# Entity validation tests
go test ./internal/domain/entity -run TestCountry_Validate
go test ./internal/domain/entity -run TestCurrency_Validate

# All tests
make test
```

## Mejores Prácticas

1. **Siempre valide primero en la capa DTO** - previene que datos incorrectos entren al sistema
2. **Use validación de entity como red de seguridad** - reglas de negocio aplicadas incluso si se evita DTO
3. **Auto-normalización en capa entity** - códigos en mayúsculas, eliminar espacios
4. **Recopile múltiples errores** - mejor UX que fallar en el primer error
5. **Pruebe validación exhaustivamente** - pruebas unitarias para casos válidos e inválidos

## Agregando Nuevos Validadores

### Custom Gin Validator

1. Agregue a `pkg/validator/custom_validators.go`:

```go
func validateMyRule(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    // Your validation logic
    return true
}
```

2. Registre en `RegisterCustomValidators()`:

```go
if err := v.RegisterValidation("myrule", validateMyRule); err != nil {
    return err
}
```

3. Use en DTOs:

```go
Field string `json:"field" binding:"myrule"`
```

### Validación Entity

Agregue al método `Validate()` de la entity:

```go
if !isValidMyRule(c.Field) {
    errs = append(errs, "field must meet my rule")
}
```

## Consideraciones de Seguridad

- **SQL Injection**: Prevenido por consultas parametrizadas (sqlx)
- **XSS**: Prevenido por codificación JSON + Content-Type headers
- **Integridad de Datos**: Validación multicapa asegura ausencia de datos incorrectos
- **Code Injection**: Validadores alpha/alphanum previenen caracteres especiales
- **Length Attacks**: Longitud máxima aplicada en todas las capas

## Rendimiento

- Validación DTO: ~1-5μs por solicitud (insignificante)
- Validación Entity: ~100-500ns por entity (insignificante)
- Overhead total: <0.1% del tiempo típico de solicitud

## Referencias

- [go-playground/validator docs](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [Gin validation](https://gin-gonic.com/docs/examples/binding-and-validation/)
- [Clean Architecture validation patterns](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
