[🇬🇧 English](../VALIDATION.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](../de/VALIDATION.de.md) | [🇵🇹 Português](../pt/VALIDATION.pt.md) | [🇪🇸 Español](../es/VALIDATION.es.md)

# Стратегія валідації

## Багаторівнева валідація

Promenade реалізує **багаторівневий захист** через валідацію на декількох рівнях для забезпечення цілісності та безпеки даних.

### 1. Рівень DTO (HTTP вхід)

**Розташування**: `internal/adapter/http/v*/dto/*.go`

**Призначення**: Перша лінія захисту - валідує дані HTTP-запиту перед їх надходженням до бізнес-логіки.

**Технології**: Gin binding tags + go-playground/validator

**Приклад**:

```go
type CreateCountryRequest struct {
    Name string `json:"name" binding:"required,min=2,max=100"`
    Code string `json:"code" binding:"required,min=2,max=10,uppercase"`
    ISO2 string `json:"iso2" binding:"required,len=2,alpha,uppercase"`
    ISO3 string `json:"iso3" binding:"required,len=3,alpha,uppercase"`
}
```

**Кастомні валідатори** (`pkg/validator/custom_validators.go`):

- `uppercase` - Забезпечує верхній регістр рядка
- `alpha` - Дозволяє тільки літери
- `alphanum` - Тільки літери та цифри
- `no_special` - Без спецсимволів

### 2. Рівень Entity (Domain)

**Розташування**: `internal/domain/entity/*.go`

**Призначення**: Валідація бізнес-правил - забезпечує, що доменні об'єкти завжди у валідному стані.

**Метод**: `Validate() error` на кожній entity

**Приклад**:

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

**Можливості**:

- Авто-нормалізація (видалення пробілів, верхній регістр кодів)
- Збір множинних помилок
- Застосування бізнес-правил

### 3. Рівень Use Case

**Розташування**: `internal/usecase/*.go`

**Призначення**: Оркестраційна валідація - забезпечує виклик валідації entity перед збереженням.

**Приклад**:

```go
func (uc *CountryUseCase) Create(ctx context.Context, country *entity.Country) error {
    // Validate entity
    if err := country.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    return uc.countryRepo.Create(ctx, country)
}
```

## Правила валідації за Entity

### Country

| Поле | Правила DTO                        | Правила Entity                    | Примітки                    |
| ---- | ---------------------------------- | --------------------------------- | --------------------------- |
| Name | required, min=2, max=100           | required, 2-100 символів          | Обрізається                 |
| Code | required, min=2, max=10, uppercase | required, 2-10 символів, alphanum | Верхній регістр автоматично |
| ISO2 | required, len=2, alpha, uppercase  | required, 2 символи, alpha        | Верхній регістр автоматично |
| ISO3 | required, len=3, alpha, uppercase  | required, 3 символи, alpha        | Верхній регістр автоматично |

### Currency

| Поле   | Правила DTO                                  | Правила Entity                    | Примітки                              |
| ------ | -------------------------------------------- | --------------------------------- | ------------------------------------- |
| Name   | required, min=2, max=100                     | required, 2-100 символів          | Обрізається                           |
| Code   | required, min=3, max=10, alphanum, uppercase | required, 3-10 символів, alphanum | Верхній регістр автоматично, ISO 4217 |
| Symbol | omitempty, max=10                            | необов'язково, макс 10 символів   | Опціонально                           |

## Обробка помилок

### Помилки валідації DTO

**HTTP 400 Bad Request** з детальними помилками полів:

```json
{
  "success": false,
  "message": "invalid request",
  "error": "Key: 'CreateCountryRequest.ISO2' Error:Field validation for 'ISO2' failed on the 'len' tag"
}
```

### Помилки валідації Entity

**HTTP 500 Internal Server Error** (повинна бути спіймана раніше валідацією DTO):

```json
{
  "success": false,
  "message": "validation failed",
  "error": "name is required; iso2 must be exactly 2 characters"
}
```

## Тестування

Запуск тестів валідації:

```bash
# Entity validation tests
go test ./internal/domain/entity -run TestCountry_Validate
go test ./internal/domain/entity -run TestCurrency_Validate

# All tests
make test
```

## Найкращі практики

1. **Завжди валідуйте спочатку на рівні DTO** - запобігає потраплянню некоректних даних у систему
2. **Використовуйте валідацію entity як страховку** - бізнес-правила застосовуються навіть якщо DTO обійшли
3. **Авто-нормалізація на рівні entity** - верхній регістр кодів, видалення пробілів
4. **Збирайте множинні помилки** - краще UX ніж провал на першій помилці
5. **Ретельно тестуйте валідацію** - юніт-тести для валідних та невалідних кейсів

## Додавання нових валідаторів

### Кастомний Gin валідатор

1. Додайте до `pkg/validator/custom_validators.go`:

```go
func validateMyRule(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    // Your validation logic
    return true
}
```

2. Зареєструйте у `RegisterCustomValidators()`:

```go
if err := v.RegisterValidation("myrule", validateMyRule); err != nil {
    return err
}
```

3. Використовуйте у DTO:

```go
Field string `json:"field" binding:"myrule"`
```

### Валідація Entity

Додайте до методу `Validate()` entity:

```go
if !isValidMyRule(c.Field) {
    errs = append(errs, "field must meet my rule")
}
```

## Міркування безпеки

- **SQL Injection**: Запобігається параметризованими запитами (sqlx)
- **XSS**: Запобігається JSON-кодуванням + Content-Type headers
- **Цілісність даних**: Багаторівнева валідація забезпечує відсутність некоректних даних
- **Code Injection**: Alpha/alphanum валідатори запобігають спецсимволам
- **Length Attacks**: Максимальна довжина застосовується на всіх рівнях

## Продуктивність

- Валідація DTO: ~1-5μs на запит (незначна)
- Валідація Entity: ~100-500ns на entity (незначна)
- Загальні витрати: <0.1% типового часу запиту

## Посилання

- [go-playground/validator docs](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [Gin validation](https://gin-gonic.com/docs/examples/binding-and-validation/)
- [Clean Architecture validation patterns](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
