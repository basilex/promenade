[ English](../VALIDATION.md) | [ Українська](../uk/VALIDATION.uk.md) |  **Deutsch** | [ Português](../pt/VALIDATION.pt.md) | [ Español](../es/VALIDATION.es.md)

# Validierungsstrategie

## Mehrschichtige Validierung

Promenade implementiert **mehrschichtige Verteidigung** durch Validierung auf mehreren Ebenen, um Datenintegrität und Sicherheit zu gewährleisten.

### 1. DTO-Ebene (HTTP-Eingang)

**Ort**: `internal/adapter/http/v*/dto/*.go`

**Zweck**: Erste Verteidigungslinie - validiert HTTP-Request-Daten, bevor sie zur Geschäftslogik gelangen.

**Technologien**: Gin binding tags + go-playground/validator

**Beispiel**:

```go
type CreateCountryRequest struct {
    Name string `json:"name" binding:"required,min=2,max=100"`
    Code string `json:"code" binding:"required,min=2,max=10,uppercase"`
    ISO2 string `json:"iso2" binding:"required,len=2,alpha,uppercase"`
    ISO3 string `json:"iso3" binding:"required,len=3,alpha,uppercase"`
}
```

**Custom Validators** (`pkg/validator/custom_validators.go`):

- `uppercase` - Stellt sicher, dass String in Großbuchstaben ist
- `alpha` - Nur Buchstaben erlaubt
- `alphanum` - Nur Buchstaben und Zahlen
- `no_special` - Keine Sonderzeichen

### 2. Entity-Ebene (Domain)

**Ort**: `internal/domain/entity/*.go`

**Zweck**: Geschäftsregelvalidierung - stellt sicher, dass Domain-Objekte immer in einem gültigen Zustand sind.

**Methode**: `Validate() error` auf jeder entity

**Beispiel**:

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

**Funktionen**:

- Auto-Normalisierung (Leerzeichen entfernen, Codes in Großbuchstaben)
- Sammlung mehrerer Fehler
- Durchsetzung von Geschäftsregeln

### 3. Use Case-Ebene

**Ort**: `internal/usecase/*.go`

**Zweck**: Orchestrierungsvalidierung - stellt sicher, dass Entity-Validierung vor Persistierung aufgerufen wird.

**Beispiel**:

```go
func (uc *ICountryUseCase) Create(ctx context.Context, country *entity.Country) error {
    // Validate entity
    if err := country.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    return uc.countryRepo.Create(ctx, country)
}
```

## Validierungsregeln nach Entity

### Country

| Feld | DTO-Regeln                         | Entity-Regeln                    | Hinweise            |
| ---- | ---------------------------------- | -------------------------------- | ------------------- |
| Name | required, min=2, max=100           | required, 2-100 Zeichen          | Gekürzt             |
| Code | required, min=2, max=10, uppercase | required, 2-10 Zeichen, alphanum | Auto-Großbuchstaben |
| ISO2 | required, len=2, alpha, uppercase  | required, 2 Zeichen, alpha       | Auto-Großbuchstaben |
| ISO3 | required, len=3, alpha, uppercase  | required, 3 Zeichen, alpha       | Auto-Großbuchstaben |

### Currency

| Feld   | DTO-Regeln                                   | Entity-Regeln                    | Hinweise                      |
| ------ | -------------------------------------------- | -------------------------------- | ----------------------------- |
| Name   | required, min=2, max=100                     | required, 2-100 Zeichen          | Gekürzt                       |
| Code   | required, min=3, max=10, alphanum, uppercase | required, 3-10 Zeichen, alphanum | Auto-Großbuchstaben, ISO 4217 |
| Symbol | omitempty, max=10                            | optional, max 10 Zeichen         | Optional                      |

## Fehlerbehandlung

### DTO-Validierungsfehler

**HTTP 400 Bad Request** mit detaillierten Feldfehlern:

```json
{
  "success": false,
  "message": "invalid request",
  "error": "Key: 'CreateCountryRequest.ISO2' Error:Field validation for 'ISO2' failed on the 'len' tag"
}
```

### Entity-Validierungsfehler

**HTTP 500 Internal Server Error** (sollte früher durch DTO-Validierung abgefangen werden):

```json
{
  "success": false,
  "message": "validation failed",
  "error": "name is required; iso2 must be exactly 2 characters"
}
```

## Testen

Validierungstests ausführen:

```bash
# Entity validation tests
go test ./internal/domain/entity -run TestCountry_Validate
go test ./internal/domain/entity -run TestCurrency_Validate

# All tests
make test
```

## Best Practices

1. **Immer zuerst auf DTO-Ebene validieren** - verhindert, dass fehlerhafte Daten ins System gelangen
2. **Entity-Validierung als Sicherheitsnetz verwenden** - Geschäftsregeln werden durchgesetzt, auch wenn DTO umgangen wird
3. **Auto-Normalisierung auf Entity-Ebene** - Codes in Großbuchstaben, Leerzeichen entfernen
4. **Mehrere Fehler sammeln** - bessere UX als beim ersten Fehler abzubrechen
5. **Validierung gründlich testen** - Unit-Tests für gültige und ungültige Fälle

## Neue Validatoren hinzufügen

### Custom Gin Validator

1. Hinzufügen zu `pkg/validator/custom_validators.go`:

```go
func validateMyRule(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    // Your validation logic
    return true
}
```

2. Registrieren in `RegisterCustomValidators()`:

```go
if err := v.RegisterValidation("myrule", validateMyRule); err != nil {
    return err
}
```

3. In DTOs verwenden:

```go
Field string `json:"field" binding:"myrule"`
```

### Entity-Validierung

Zur `Validate()`-Methode der Entity hinzufügen:

```go
if !isValidMyRule(c.Field) {
    errs = append(errs, "field must meet my rule")
}
```

## Sicherheitsüberlegungen

- **SQL Injection**: Verhindert durch parametrisierte Abfragen (sqlx)
- **XSS**: Verhindert durch JSON-Codierung + Content-Type Headers
- **Datenintegrität**: Mehrschichtige Validierung gewährleistet keine fehlerhaften Daten
- **Code Injection**: Alpha/alphanum-Validatoren verhindern Sonderzeichen
- **Length Attacks**: Maximale Länge auf allen Ebenen durchgesetzt

## Performance

- DTO-Validierung: ~1-5μs pro Anfrage (vernachlässigbar)
- Entity-Validierung: ~100-500ns pro Entity (vernachlässigbar)
- Gesamtoverhead: <0.1% der typischen Anfragzeit

## Referenzen

- [go-playground/validator docs](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [Gin validation](https://gin-gonic.com/docs/examples/binding-and-validation/)
- [Clean Architecture validation patterns](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
