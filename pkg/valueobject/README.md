# Value Objects Package

**Purpose:** Immutable Domain-Driven Design value objects  
**Status:** Production-ready  
**Tests:** 25 tests, 95% coverage

---

## Overview

The `valueobject` package provides **immutable value objects** following Domain-Driven Design principles. Value objects encapsulate domain concepts that are defined by their attributes rather than identity.

## Available Value Objects

- **Email** - Email addresses with validation
- **Phone** - E.164 formatted phone numbers
- **Money** - Currency amounts (stored in cents)
- **Address** - Postal addresses with validation

---

## Installation

```go
import "github.com/basilex/promenade/pkg/valueobject"
```

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

### Usage Examples

#### Create Email

```go
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

### Usage Examples

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
fmt.Println(err) // Output: phone can only contain digits, spaces, and hyphens
```

### API Reference

```go
// NewPhone creates a new Phone value object
func NewPhone(phone string) (Phone, error)

// Value returns the E.164 formatted phone
func (p Phone) Value() string

// CountryCode extracts the country calling code
func (p Phone) CountryCode() (string, error)

// Formatted returns human-readable format (+1 555 123 4567)
func (p Phone) Formatted() string

// Equals checks if two phones are equal
func (p Phone) Equals(other Phone) bool

// IsEmpty checks if phone is empty
func (p Phone) IsEmpty() bool
```

---

## Money Value Object

### Overview

Currency amount stored in cents to avoid floating-point errors.

### Features

- Stored as **int64 cents** (no floating-point errors)
- Currency validation (ISO 4217)
- Arithmetic operations (Add, Subtract, Multiply, Divide)
- Comparison operations
- Currency conversion helpers
- Immutable

### Usage Examples

#### Create Money

```go
// From cents (recommended)
price := valueobject.NewMoney(1050, "USD") // $10.50
fmt.Println(price.ToFloat()) // Output: 10.50

// From float (convenience method)
price, _ := valueobject.NewMoneyFromFloat(10.50, "USD")
fmt.Println(price.Amount) // Output: 1050 (cents)
```

#### Arithmetic Operations

```go
price := valueobject.NewMoney(1000, "USD") // $10.00
tax := valueobject.NewMoney(150, "USD")    // $1.50

// Addition
total, _ := price.Add(tax)
fmt.Println(total.ToFloat()) // Output: 11.50

// Subtraction
discount := valueobject.NewMoney(200, "USD") // $2.00
final, _ := total.Subtract(discount)
fmt.Println(final.ToFloat()) // Output: 9.50

// Multiplication (by scalar)
double, _ := price.Multiply(2)
fmt.Println(double.ToFloat()) // Output: 20.00

// Division (by scalar)
half, _ := price.Divide(2)
fmt.Println(half.ToFloat()) // Output: 5.00
```

#### Currency Validation

```go
price := valueobject.NewMoney(1000, "USD")
vat := valueobject.NewMoney(200, "EUR")

// Operations require same currency
_, err := price.Add(vat)
fmt.Println(err) // Output: cannot add different currencies: USD and EUR
```

#### Comparison

```go
price1 := valueobject.NewMoney(1000, "USD")
price2 := valueobject.NewMoney(1500, "USD")

// Less than
less, _ := price1.LessThan(price2)
fmt.Println(less) // Output: true

// Greater than
greater, _ := price1.GreaterThan(price2)
fmt.Println(greater) // Output: false

// Equality
equal := price1.Equals(price2)
fmt.Println(equal) // Output: false
```

#### Formatting

```go
price := valueobject.NewMoney(1050, "USD")

// String format
fmt.Println(price.String()) // Output: USD 10.50

// Symbol format
fmt.Println(price.Format()) // Output: $10.50

// Negative amounts
debt := valueobject.NewMoney(-5000, "EUR")
fmt.Println(debt.Format()) // Output: -€50.00
```

#### Type Checks

```go
zero := valueobject.NewMoney(0, "USD")
positive := valueobject.NewMoney(100, "USD")
negative := valueobject.NewMoney(-100, "USD")

fmt.Println(zero.IsZero())         // Output: true
fmt.Println(positive.IsPositive()) // Output: true
fmt.Println(negative.IsNegative()) // Output: true

// Absolute value
abs := negative.Abs()
fmt.Println(abs.ToFloat()) // Output: 1.00

// Negate
negated := positive.Negate()
fmt.Println(negated.ToFloat()) // Output: -1.00
```

### API Reference

```go
// NewMoney creates Money from cents
func NewMoney(cents int64, currency string) Money

// NewMoneyFromFloat creates Money from float (convenience)
func NewMoneyFromFloat(amount float64, currency string) (Money, error)

// Add adds two Money values (same currency)
func (m Money) Add(other Money) (Money, error)

// Subtract subtracts two Money values (same currency)
func (m Money) Subtract(other Money) (Money, error)

// Multiply multiplies Money by scalar
func (m Money) Multiply(multiplier float64) (Money, error)

// Divide divides Money by scalar
func (m Money) Divide(divisor float64) (Money, error)

// ToFloat converts cents to float
func (m Money) ToFloat() float64

// IsZero checks if amount is zero
func (m Money) IsZero() bool

// IsPositive checks if amount is positive
func (m Money) IsPositive() bool

// IsNegative checks if amount is negative
func (m Money) IsNegative() bool

// Abs returns absolute value
func (m Money) Abs() Money

// Negate returns negated value
func (m Money) Negate() Money

// Compare compares two Money values
func (m Money) Compare(other Money) (int, error)

// Equals checks equality
func (m Money) Equals(other Money) bool

// String returns formatted string
func (m Money) String() string

// Format returns formatted string with currency symbol
func (m Money) Format() string
```

---

## Address Value Object

### Overview

Postal address with validation and formatting.

### Features

- ISO 3166-1 country codes
- Street validation
- Optional Street2
- City/State/ZIP validation
- Single-line formatting
- Immutable

### Usage Examples

#### Create Address

```go
address, err := valueobject.NewAddress(
    "123 Main Street",
    "",                  // Street2 (optional)
    "Kyiv",
    "Kyiv Oblast",
    "01001",
    "UA",               // ISO 3166-1 country code
)
if err != nil {
    log.Fatal(err)
}
```

#### With Optional Street2

```go
address, _ := valueobject.NewAddress(
    "456 Oak Avenue",
    "Apartment 5B",    // Street2
    "New York",
    "NY",
    "10001",
    "US",
)

fmt.Println(address.Street1) // Output: 456 Oak Avenue
fmt.Println(address.Street2) // Output: Apartment 5B
```

#### Formatting

```go
address, _ := valueobject.NewAddress(
    "123 Main St",
    "",
    "Kyiv",
    "Kyiv Oblast",
    "01001",
    "UA",
)

// Single line format
fmt.Println(address.SingleLine())
// Output: 123 Main St, Kyiv, Kyiv Oblast 01001, Ukraine

// Multi-line format
fmt.Println(address.String())
// Output:
// 123 Main St
// Kyiv, Kyiv Oblast 01001
// Ukraine
```

#### Update Street2

```go
address, _ := valueobject.NewAddress("123 Main St", "", "Kyiv", "Kyiv Oblast", "01001", "UA")

// Add Street2
updated := address.WithStreet2("Suite 100")
fmt.Println(updated.Street2) // Output: Suite 100

// Original unchanged (immutable)
fmt.Println(address.Street2) // Output: (empty)
```

#### Validation

```go
// Missing street
_, err := valueobject.NewAddress("", "", "Kyiv", "Kyiv Oblast", "01001", "UA")
fmt.Println(err) // Output: street is required

// Missing city
_, err = valueobject.NewAddress("123 Main St", "", "", "Kyiv Oblast", "01001", "UA")
fmt.Println(err) // Output: city is required

// Invalid country code
_, err = valueobject.NewAddress("123 Main St", "", "Kyiv", "", "01001", "UKR")
fmt.Println(err) // Output: invalid country code (must be ISO 3166-1)
```

### API Reference

```go
// NewAddress creates a new Address value object
func NewAddress(street1, street2, city, state, zip, country string) (Address, error)

// SingleLine returns single-line formatted address
func (a Address) SingleLine() string

// String returns multi-line formatted address
func (a Address) String() string

// WithStreet2 returns new Address with updated Street2
func (a Address) WithStreet2(street2 string) Address

// Equals checks if two addresses are equal
func (a Address) Equals(other Address) bool

// IsEmpty checks if address is empty
func (a Address) IsEmpty() bool
```

---

## Best Practices

### DO

- **Use value objects for domain concepts** - Not just primitives
  ```go
  // Good: Type-safe email
  type User struct {
      Email valueobject.Email
  }
  
  // Bad: Primitive string
  type User struct {
      Email string // No validation
  }
  ```

- **Validate on creation** - Constructor ensures validity
  ```go
  email, err := valueobject.NewEmail(input)
  if err != nil {
      return err // Invalid email rejected immediately
  }
  ```

- **Use Money for financial calculations** - Avoid floating-point errors
  ```go
  // Good: No rounding errors
  price := valueobject.NewMoney(1050, "USD")
  tax := valueobject.NewMoney(150, "USD")
  total, _ := price.Add(tax)
  
  // Bad: Floating-point errors
  price := 10.50
  tax := 1.50
  total := price + tax // May be 12.000000000000002
  ```

- **Store in cents** - Database integer storage
  ```sql
  CREATE TABLE products (
      price_cents INT NOT NULL,
      price_currency VARCHAR(3) NOT NULL
  );
  ```

### DON'T

- **Don't modify value objects** - They're immutable
  ```go
  // Bad: Trying to modify
  email.value = "new@email.com" // Compiler error (private field)
  
  // Good: Create new instance
  newEmail, _ := valueobject.NewEmail("new@email.com")
  ```

- **Don't bypass validation** - Always use constructors
  ```go
  // Bad: Direct struct initialization
  email := valueobject.Email{value: "invalid"} // Compiler error
  
  // Good: Use constructor
  email, err := valueobject.NewEmail("test@example.com")
  ```

- **Don't use floats for money** - Use Money value object
  ```go
  // Bad: Floating-point arithmetic
  price := 10.50
  quantity := 3
  total := price * float64(quantity) // Rounding errors
  
  // Good: Money arithmetic
  price := valueobject.NewMoney(1050, "USD")
  total, _ := price.Multiply(3) // Exact calculation
  ```

---

## Domain Entity Example

```go
package user

import (
    "github.com/basilex/promenade/pkg/valueobject"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type User struct {
    ID      uuidv7.UUID
    Email   valueobject.Email    // Type-safe email
    Name    string
}

type Contact struct {
    ID      uuidv7.UUID
    UserID  uuidv7.UUID
    Email   *valueobject.Email   // Optional email
    Phone   *valueobject.Phone   // Optional phone
    Address *valueobject.Address // Optional address
}

// Factory method with validation
func NewUser(email, name string) (*User, error) {
    // Email validation happens in constructor
    emailVO, err := valueobject.NewEmail(email)
    if err != nil {
        return nil, fmt.Errorf("invalid email: %w", err)
    }
    
    return &User{
        ID:    uuidv7.New(),
        Email: emailVO,
        Name:  name,
    }, nil
}
```

---

## Database Storage

### Email & Phone

```sql
-- Store as VARCHAR
CREATE TABLE users (
    id    UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) NULL
);
```

```go
// Repository layer
func (r *userRepository) Create(ctx context.Context, user *User) error {
    query := `INSERT INTO users (id, email, phone) VALUES ($1, $2, $3)`
    _, err := r.db.ExecContext(ctx, query, 
        user.ID,
        user.Email.Value(),    // Extract string value
        user.Phone.Value(),
    )
    return err
}
```

### Money

```sql
-- Store cents as INT, currency as VARCHAR
CREATE TABLE products (
    id             UUID PRIMARY KEY,
    price_cents    INT NOT NULL,
    price_currency VARCHAR(3) NOT NULL,
    CHECK (price_cents >= 0)
);
```

```go
// Repository layer
func (r *productRepository) Create(ctx context.Context, product *Product) error {
    query := `INSERT INTO products (id, price_cents, price_currency) VALUES ($1, $2, $3)`
    _, err := r.db.ExecContext(ctx, query, 
        product.ID,
        product.Price.Amount,   // Cents (int64)
        product.Price.Currency, // ISO code
    )
    return err
}
```

### Address

```sql
-- Store as individual columns or JSONB
CREATE TABLE users (
    id          UUID PRIMARY KEY,
    address     JSONB NULL
);
```

```go
// Repository layer (JSONB)
func (r *userRepository) Create(ctx context.Context, user *User) error {
    addressJSON, _ := json.Marshal(user.Address)
    query := `INSERT INTO users (id, address) VALUES ($1, $2)`
    _, err := r.db.ExecContext(ctx, query, user.ID, addressJSON)
    return err
}
```

---

## Testing

```go
package valueobject_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/basilex/promenade/pkg/valueobject"
)

func TestEmail_Validation(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid", "test@example.com", false},
        {"empty", "", true},
        {"no @", "invalid", true},
        {"no domain", "test@", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := valueobject.NewEmail(tt.email)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestMoney_Arithmetic(t *testing.T) {
    price := valueobject.NewMoney(1000, "USD")
    tax := valueobject.NewMoney(150, "USD")
    
    total, err := price.Add(tax)
    assert.NoError(t, err)
    assert.Equal(t, int64(1150), total.Amount)
    assert.Equal(t, 11.50, total.ToFloat())
}
```

---

## Related Packages

- `pkg/aggregate` - Base aggregate with value objects
- `pkg/uuidv7` - UUID generation for entities
- `internal/contexts/identity/contact` - Contact aggregate uses Email, Phone, Address

---

**Last Updated:** 2025-12-28  
**Status:** Production-ready  
**Maintainer:** Promenade Team
