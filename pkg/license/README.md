# License Package

Shared licensing system for Promenade commercial modules.

## Overview

This package provides signature-based license validation for commercial modules. It uses HMAC-SHA256 for cryptographic verification and supports multiple tiers with grace periods.

## License Format

```
PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}
```

**Example:**

```
PROMENADE-AUDIT-PRO-20261231-AbC1234XyZ...
```

## Usage

### Generating License Keys

```go
import "github.com/basilex/promenade/pkg/license"

key := license.Generate(
    "your-secret-key",
    "audit",
    license.TierPro,
    time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
)
// Returns: PROMENADE-AUDIT-PRO-20261231-{signature}
```

### Validating Licenses

```go
// Parse license key
lic, err := license.Parse(licenseKey)
if err != nil {
    return err
}

// Validate
err = lic.Validate(secret, "audit", 3) // 3 days grace period
if err != nil {
    return err
}

// Check expiry
if lic.IsExpired() {
    fmt.Println("License expired", lic.DaysUntilExpiry(), "days ago")
}
```

## License Tiers

| Tier       | Constant                 | Description                           |
| ---------- | ------------------------ | ------------------------------------- |
| BASIC      | `license.TierBasic`      | Basic features, limited retention     |
| PRO        | `license.TierPro`        | Advanced features, extended retention |
| ENTERPRISE | `license.TierEnterprise` | Full features, unlimited retention    |

## Security

- **Algorithm**: HMAC-SHA256
- **Encoding**: Base64 URL-safe (no padding)
- **Secret Storage**: Environment variables only
- **Grace Period**: Configurable days after expiration

## Environment Variables

```bash
# IModule-specific license key
export AUDIT_LICENSE_KEY="PROMENADE-AUDIT-PRO-20261231-..."

# License secret (for validation)
export LICENSE_SECRET="your-secret-key"
```

## Error Handling

```go
import "github.com/basilex/promenade/pkg/license"

err := lic.Validate(secret, "audit", 3)
switch {
case errors.Is(err, license.ErrInvalidFormat):
    // Malformed license key
case errors.Is(err, license.ErrInvalidSignature):
    // Signature verification failed
case errors.Is(err, license.ErrExpired):
    // License expired beyond grace period
case errors.Is(err, license.ErrModuleMismatch):
    // License for different module
}
```

## Testing

```bash
go test ./pkg/license/...
```

## Used By

- `internal/modules/audit` - Audit logging and compliance (Commercial)
- Future commercial modules

---

**Security Note**: Never commit license secrets to version control. Use environment variables or secure secret management systems.
