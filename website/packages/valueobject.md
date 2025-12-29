# Value Objects Package

**Immutable Domain-Driven Design value objects**

---

## Overview

The `valueobject` package provides **immutable value objects** following Domain-Driven Design (DDD) principles. Value objects encapsulate domain concepts that are defined by their attributes rather than identity.

**Status**: Production-ready  
**Tests**: 25 tests, 95% coverage  
**Location**: `pkg/valueobject/`

---

## Available Value Objects

| Value Object | Purpose | Validation |
|--------------|---------|------------|
| **Email** | Email addresses | RFC 5322 format |
| **Phone** | Phone numbers | E.164 format |
| **Money** | Currency amounts | Stored in cents |
| **Address** | Postal addresses | Multi-field validation |

---

## Email Value Object

### Overview

Immutable email address with validation and normalization.

### Features

- RFC 5322 format validation
- Case normalization (lowercase)
- Domain extraction
- Local part extraction
- Immutable

### Usage

#### Create Email

```go
import "github.com/basilex/promenade/pkg/valueobject"

// Create valid email
email, err := valueobject.NewEmail("John.Doe@Example.COM")
if err != nil {
    log.Fatal(err)
}

fmt.Println(email.Value()) // Output: john.doe@example.com (normalized)
```

#### Extract Parts

```go
email, _ := valueobject.NewEmail("john.doe@example.com")

// Get local part (before @)
fmt.Println(email.LocalPart()) // Output: john.doe

// Get domain (after @)
fmt.Println(email.Domain()) // Output: example.com
```

#### Validation

```go
// Invalid emails return error
_, err := valueobject.NewEmail("invalid-email")
fmt.Println(err) // Output: invalid email format

_, err = valueobject.NewEmail("")
fmt.Println(err) // Output: email cannot be empty

_, err = valueobject.NewEmail("no-domain@")
fmt.Println(err) // Output: invalid email format
```

#### Comparison

```go
email1, _ := valueobject.NewEmail("john@example.com")
email2, _ := valueobject.NewEmail("JOHN@EXAMPLE.COM") // Normalized

// Equality check
fmt.Println(email1.Equals(email2)) // Output: true (case-insensitive)

// Check if empty
fmt.Println(email1.IsEmpty()) // Output: false
```

### API Reference

```go
// NewEmail creates a new Email value object
func NewEmail(email string) (Email, error)

// Value returns the email address
func (e Email) Value() string

// Domain returns the domain part
func (e Email) Domain() string

// LocalPart returns the local part (before @)
func (e Email) LocalPart() string

// Equals checks if two emails are equal (case-insensitive)
func (e Email) Equals(other Email) bool

// IsEmpty checks if email is empty
func (e Email) IsEmpty() bool
```

---

## Phone Value Object

### Overview

E.164 formatted international phone numbers with validation.

### Features

- E.164 format validation (+[country code][number])
- Country code extraction
- Human-readable formatting
- Separator removal (+1-555-123-4567 → +15551234567)
- Immutable

### Usage

#### Create Phone

```go
// Create with E.164 format
phone, err := valueobject.NewPhone("+380501234567")
if err != nil {
    log.Fatal(err)
}

fmt.Println(phone.Value()) // Output: +380501234567
```

#### Auto-Format Input

```go
// Removes separators automatically
phone, _ := valueobject.NewPhone("+1-555-123-4567")
fmt.Println(phone.Value()) // Output: +15551234567

phone, _ = valueobject.NewPhone("+38 (050) 123-45-67")
fmt.Println(phone.Value()) // Output: +380501234567
```

#### Extract Country Code

```go
phone, _ := valueobject.NewPhone("+380501234567")

// Get country code
code, _ := phone.CountryCode()
fmt.Println(code) // Output: 380
```

#### Formatted Display

```go
phone, _ := valueobject.NewPhone("+15551234567")

// Human-readable format
fmt.Println(phone.Formatted()) // Output: +1 555 123 4567
```

#### Validation

```go
// Missing + sign
_, err := valueobject.NewPhone("380501234567")
fmt.Println(err) // Output: phone must start with +

// Too short
_, err = valueobject.NewPhone("+123")
fmt.Println(err) // Output: phone must be at least 8 characters

// Invalid characters
_, err = valueobject.NewPhone("+38050abc1234")
fmt.Println(err) // Output: phone can only contain digits
```

### API Reference

```go
// NewPhone creates a new Phone value object
func NewPhone(phone string) (Phone, error)

// Value returns the E.164 formatted phone number
func (p Phone) Value() string

// CountryCode extracts country code
func (p Phone) CountryCode() (int, error)

// Formatted returns human-readable format
func (p Phone) Formatted() string

// IsEmpty checks if phone is empty
func (p Phone) IsEmpty() bool
```

---

## Money Value Object

### Overview

Monetary amounts stored in cents/smallest currency unit to avoid floating-point precision issues.

### Features

- Stored as cents (integer)
- Currency code (ISO 4217)
- Arithmetic operations
- Formatting for display
- Immutable

### Usage

#### Create Money

```go
// Create from cents
money := valueobject.NewMoney(9999, "USD")
fmt.Println(money.Amount()) // Output: 9999 (cents)
fmt.Println(money.Currency()) // Output: USD

// Create from dollars
money := valueobject.NewMoneyFromFloat(99.99, "USD")
fmt.Println(money.Amount()) // Output: 9999 (cents)
```

#### Arithmetic Operations

```go
price := valueobject.NewMoney(5000, "USD") // $50.00
tax := valueobject.NewMoney(500, "USD")    // $5.00

// Addition
total := price.Add(tax)
fmt.Println(total.Amount()) // Output: 5500 ($55.00)

// Subtraction
discount := valueobject.NewMoney(1000, "USD") // $10.00
final := total.Subtract(discount)
fmt.Println(final.Amount()) // Output: 4500 ($45.00)

// Multiplication
quantity := 3
totalPrice := price.Multiply(quantity)
fmt.Println(totalPrice.Amount()) // Output: 15000 ($150.00)
```

#### Formatting

```go
money := valueobject.NewMoney(12345, "USD")

// Display format
fmt.Println(money.String()) // Output: $123.45

money = valueobject.NewMoney(5000, "EUR")
fmt.Println(money.String()) // Output: €50.00
```

#### Currency Validation

```go
// Must use same currency
usd := valueobject.NewMoney(100, "USD")
eur := valueobject.NewMoney(100, "EUR")

_, err := usd.Add(eur)
fmt.Println(err) // Output: cannot add different currencies
```

### API Reference

```go
// NewMoney creates Money from cents
func NewMoney(amount int64, currency string) Money

// NewMoneyFromFloat creates Money from float
func NewMoneyFromFloat(amount float64, currency string) Money

// Amount returns amount in cents
func (m Money) Amount() int64

// Currency returns currency code
func (m Money) Currency() string

// Add adds two Money values
func (m Money) Add(other Money) (Money, error)

// Subtract subtracts two Money values
func (m Money) Subtract(other Money) (Money, error)

// Multiply multiplies Money by quantity
func (m Money) Multiply(quantity int) Money

// String returns formatted display string
func (m Money) String() string
```

---

## Address Value Object

### Overview

Postal address with multi-field validation.

### Features

- Street, city, state, country, postal code
- Country validation (ISO 3166-1)
- Postal code format validation
- Full address formatting
- Immutable

### Usage

#### Create Address

```go
address, err := valueobject.NewAddress(
    "123 Main St",      // street
    "Kyiv",             // city
    "Kyiv Oblast",      // state/region
    "UA",               // country (ISO 3166-1)
    "01001",            // postal code
)
if err != nil {
    log.Fatal(err)
}
```

#### Access Fields

```go
fmt.Println(address.Street())     // Output: 123 Main St
fmt.Println(address.City())       // Output: Kyiv
fmt.Println(address.State())      // Output: Kyiv Oblast
fmt.Println(address.Country())    // Output: UA
fmt.Println(address.PostalCode()) // Output: 01001
```

#### Formatted Address

```go
// Full address (multi-line)
fmt.Println(address.String())
// Output:
// 123 Main St
// Kyiv, Kyiv Oblast 01001
// Ukraine
```

#### Validation

```go
// Empty street
_, err := valueobject.NewAddress("", "Kyiv", "Kyiv Oblast", "UA", "01001")
fmt.Println(err) // Output: street is required

// Invalid country code
_, err = valueobject.NewAddress("123 Main St", "Kyiv", "", "INVALID", "01001")
fmt.Println(err) // Output: invalid country code
```

### API Reference

```go
// NewAddress creates a new Address value object
func NewAddress(street, city, state, country, postalCode string) (Address, error)

// Street returns street address
func (a Address) Street() string

// City returns city name
func (a Address) City() string

// State returns state/region
func (a Address) State() string

// Country returns ISO 3166-1 country code
func (a Address) Country() string

// PostalCode returns postal code
func (a Address) PostalCode() string

// String returns formatted multi-line address
func (a Address) String() string

// IsEmpty checks if address is empty
func (a Address) IsEmpty() bool
```

---

## Usage in Entities

### Contact Entity Example

```go
package contact

import (
    "github.com/basilex/promenade/pkg/valueobject"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type Contact struct {
    ID      uuidv7.UUID
    UserID  uuidv7.UUID
    Type    ContactType
    
    // Value Objects
    Email   *valueobject.Email   // Optional
    Phone   *valueobject.Phone   // Optional
    Address *valueobject.Address // Optional
    
    IsPrimary bool
}

// Factory method with validation
func NewEmailContact(userID uuidv7.UUID, emailStr, label string) (*Contact, error) {
    // Create Email value object
    email, err := valueobject.NewEmail(emailStr)
    if err != nil {
        return nil, fmt.Errorf("invalid email: %w", err)
    }
    
    return &Contact{
        ID:     uuidv7.New(),
        UserID: userID,
        Type:   ContactTypeEmail,
        Email:  &email,
    }, nil
}

func NewPhoneContact(userID uuidv7.UUID, phoneStr, label string) (*Contact, error) {
    // Create Phone value object
    phone, err := valueobject.NewPhone(phoneStr)
    if err != nil {
        return nil, fmt.Errorf("invalid phone: %w", err)
    }
    
    return &Contact{
        ID:     uuidv7.New(),
        UserID: userID,
        Type:   ContactTypePhone,
        Phone:  &phone,
    }, nil
}
```

---

## Best Practices

### DO

✅ **Use value objects for domain concepts**:
```go
type User struct {
    Email valueobject.Email  // ✅ Type-safe
    Phone valueobject.Phone  // ✅ Validated
}
```

✅ **Validate on creation**:
```go
email, err := valueobject.NewEmail(input)
if err != nil {
    return fmt.Errorf("invalid email: %w", err)
}
```

✅ **Use immutability**:
```go
// Value objects are immutable
email1 := valueobject.MustNewEmail("john@example.com")
email2 := valueobject.MustNewEmail("jane@example.com")

// Create new instance, don't modify
user.Email = email2
```

✅ **Compare by value**:
```go
if user.Email.Equals(other.Email) {
    // Same email address
}
```

### DON'T

❌ **Don't use primitive types**:
```go
// ❌ Bad (no validation)
type User struct {
    Email string
    Phone string
}

// ✅ Good (validated value objects)
type User struct {
    Email valueobject.Email
    Phone valueobject.Phone
}
```

❌ **Don't bypass validation**:
```go
// ❌ Bad
user.Email = "invalid-email"

// ✅ Good
email, err := valueobject.NewEmail("valid@example.com")
if err == nil {
    user.Email = email
}
```

❌ **Don't expose internal state**:
```go
// ❌ Bad
type Email struct {
    Value string  // Public field
}

// ✅ Good
type Email struct {
    value string  // Private field
}

func (e Email) Value() string {
    return e.value
}
```

---

## Testing

### Running Tests

```bash
# Run value object tests
go test ./pkg/valueobject -v

# With coverage
go test ./pkg/valueobject -cover
```

### Test Examples

```go
func TestEmail_NewEmail_Valid(t *testing.T) {
    email, err := valueobject.NewEmail("test@example.com")
    
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", email.Value())
    assert.Equal(t, "example.com", email.Domain())
    assert.Equal(t, "test", email.LocalPart())
}

func TestPhone_NewPhone_InvalidFormat(t *testing.T) {
    _, err := valueobject.NewPhone("invalid")
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "phone must start with +")
}

func TestMoney_Add_DifferentCurrencies(t *testing.T) {
    usd := valueobject.NewMoney(100, "USD")
    eur := valueobject.NewMoney(100, "EUR")
    
    _, err := usd.Add(eur)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "different currencies")
}
```

---

## Database Persistence

### PostgreSQL Example

```go
// Entity with value objects
type Contact struct {
    ID      uuidv7.UUID
    Email   *valueobject.Email
    Phone   *valueobject.Phone
    Address *valueobject.Address
}

// Repository row mapping
type contactRow struct {
    ID         string         `db:"id"`
    Email      sql.NullString `db:"email"`
    Phone      sql.NullString `db:"phone"`
    Street     sql.NullString `db:"street"`
    City       sql.NullString `db:"city"`
    Country    sql.NullString `db:"country"`
    PostalCode sql.NullString `db:"postal_code"`
}

func (r *contactRow) toEntity() (*Contact, error) {
    contact := &Contact{
        ID: uuidv7.MustParse(r.ID),
    }
    
    // Map Email value object
    if r.Email.Valid {
        email, err := valueobject.NewEmail(r.Email.String)
        if err != nil {
            return nil, err
        }
        contact.Email = &email
    }
    
    // Map Phone value object
    if r.Phone.Valid {
        phone, err := valueobject.NewPhone(r.Phone.String)
        if err != nil {
            return nil, err
        }
        contact.Phone = &phone
    }
    
    // Map Address value object
    if r.Street.Valid && r.City.Valid {
        address, err := valueobject.NewAddress(
            r.Street.String,
            r.City.String,
            "",
            r.Country.String,
            r.PostalCode.String,
        )
        if err != nil {
            return nil, err
        }
        contact.Address = &address
    }
    
    return contact, nil
}
```

---

## Related Documentation

- [Architecture Guide](/guide/architecture) - DDD concepts
- [Testing Guide](/guide/testing) - Testing patterns
- [Contributing Guide](/guide/contributing) - Development workflow
- [GitHub Repository](https://github.com/basilex/promenade/tree/dev/pkg/valueobject) - Source code

---

## References

- [Domain-Driven Design](https://martinfowler.com/bliki/ValueObject.html) - Martin Fowler on Value Objects
- [DDD Patterns](https://www.domainlanguage.com/ddd/) - Eric Evans' DDD concepts
- [Effective Go](https://go.dev/doc/effective_go) - Go best practices
