# License System Architecture

Promenade uses a signature-based licensing system for commercial modules. This document describes the architecture, implementation, and usage patterns.

---

## Overview

### Design Principles

1. **Module Independence**: Each module manages its own licensing
2. **Signature-Based Security**: HMAC-SHA256 prevents tampering
3. **Graceful Degradation**: Grace periods for expired licenses
4. **Environment-Specific**: Different validation rules per environment
5. **Human-Readable**: License keys are structured and parseable

### License Format

```
PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}
```

Example:

```
PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Components:

- **Prefix**: Always `PROMENADE`
- **Module**: Module name in UPPERCASE (ANALYTICS, WAREHOUSE, AUDITLOG)
- **Tier**: License tier (BASIC, PRO, ENTERPRISE)
- **Expiry**: Date in YYYYMMDD format
- **Signature**: HMAC-SHA256 of first 4 parts, base64 URL-encoded

---

## Architecture

### System Components

```
─────────────────────────────────────────────────────────────
│                      Application Startup                     │
└────────────────────────────────────────────────────────────
                       │
                       ▼
         ─────────────────────────
         │  Module Registry        │
         │  (pkg/module)           │
         └────────────────────────
                  │
                  │ For each enabled module
                  ▼
         ─────────────────────────
         │  Module.Initialize()    │
         └────────────────────────
                  │
                  ▼
         ─────────────────────────
         │  License Validator      │
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
         │  License Status         │
         │  - Valid                │
         │  - Expired (grace)      │
         │  - Invalid              │
         └─────────────────────────
```

### Validation Flow

1. **Parse**: Extract components from license string
2. **Verify Module**: Ensure license matches module name
3. **Verify Signature**: HMAC-SHA256 validation
4. **Check Expiry**: Validate date + grace period
5. **Return Status**: Valid, expired (grace), or invalid

### Storage

Licenses are stored as environment variables:

- `{MODULE}_LICENSE_KEY`: The license key
- `LICENSE_SECRET`: Secret for signature verification (production)

Example:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret-key"
```

---

## Implementation

### Module Integration

Each commercial module implements license validation in its `Initialize()` method:

```go
func (m *AnalyticsModule) Initialize(cfg interface{}) error {
    // Load module config
    config, ok := cfg.(*Config)
    if !ok {
        return ErrInvalidConfig
    }

    // Validate license if required
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

### License Validation

The license package provides core validation logic:

```go
// Parse extracts components from license string
func ParseLicense(licenseKey string) (*License, error)

// Validate checks license validity
func (l *License) Validate(
    expectedModule string,
    secret string,
    options ValidationOptions,
) error

// Generate creates new license (for testing/tooling)
func GenerateLicense(
    module, tier, expiry, secret string,
) (string, error)
```

### Health Checks

Each module's health check includes license status:

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

## Configuration

### Environment-Specific Settings

#### Development (`config.dev.yaml`)

```yaml
license_required: false # Optional in development
validate_expiry: false # Ignore expiration
validate_signature: true # Still verify signatures
grace_period_days: 30 # Long grace period
```

#### Test (`config.test.yaml`)

```yaml
license_required: false # No license in CI/CD
validate_expiry: false
validate_signature: false
grace_period_days: 0
```

#### Production (`config.prod.yaml`)

```yaml
license_required: true # Mandatory
validate_expiry: true # Strict expiration
validate_signature: true # Full security
grace_period_days: 7 # Limited grace
validate_on_request: true # Optional per-request validation
```

### Module Configuration

In `config/modules.yaml`:

```yaml
modules:
  analytics:
    enabled: true
    version: "1.0.0"
    description: "Advanced analytics and reporting"
    license_key: "" # Set via environment variable
    settings:
      metrics_retention_days: 90
      max_reports_per_user: 10
      max_dashboards_per_user: 5
```

---

## Tiers & Features

### BASIC Tier

- Core metrics collection (100 metrics/day)
- Basic reports (JSON, CSV)
- 5 dashboards per user
- 30-day data retention
- Email support

### PRO Tier

- Unlimited metrics
- Advanced reports (PDF, XLSX)
- Unlimited dashboards
- 90-day data retention
- Scheduled reports
- Priority support

### ENTERPRISE Tier

- All PRO features
- Custom retention (1+ years)
- Multi-tenant support
- Custom branding
- API access
- Dedicated support
- On-premise deployment

---

## Usage

### Generating Licenses

Use the provided script:

```bash
# Generate a PRO license valid for 365 days
./scripts/generate-license.sh analytics PRO 365

# Output:
# PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Set environment variable:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret"
```

### Testing License Validation

The analytics module includes comprehensive tests:

```bash
# Run license validation tests
go test ./internal/modules/analytics/license/... -v

# Test scenarios:
# - Valid license parsing
# - Invalid format detection
# - Signature verification
# - Expiration handling
# - Grace period logic
# - Module mismatch detection
```

### License Generation Tool

Build the license generator:

```bash
make build-license-generator

# Or manually:
go build -o ./bin/license-generator ./cmd/license-generator/main.go
```

Generate license programmatically:

```bash
./bin/license-generator \
  -module=ANALYTICS \
  -tier=PRO \
  -expiry=20261231 \
  -secret="your-secret-key"
```

---

## Security

### HMAC-SHA256 Signatures

- **Algorithm**: HMAC with SHA-256
- **Key**: Stored in `LICENSE_SECRET` environment variable
- **Encoding**: Base64 URL encoding (URL-safe, no padding)
- **Verification**: Constant-time comparison to prevent timing attacks

### Secret Management

**Development**:

```bash
# Default test secret (acceptable for local dev)
LICENSE_SECRET="default-dev-secret-change-in-production"
```

**Production**:

```bash
# Generate strong secret (32+ bytes)
openssl rand -base64 32

# Store in secure vault (AWS Secrets Manager, HashiCorp Vault, etc.)
# Set via environment variable (never commit to git)
```

### Rotation Strategy

1. Generate new secret
2. Sign new licenses with new secret
3. Support both old and new secrets during transition period
4. Deprecate old secret after grace period

---

## Error Handling

### Error Types

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

### Graceful Degradation

1. **Missing License** (dev/test): Module loads with reduced features
2. **Expired License**: Grace period allows continued operation with warnings
3. **Invalid License**: Module fails to initialize in production, logs in dev

### User-Facing Messages

```go
// Production error (strict)
return fmt.Errorf("analytics module requires valid license: %w", err)

// Development warning (permissive)
logger.Warn("Analytics license validation failed, continuing in dev mode",
    "error", err)
```

---

## Monetization Strategy

### Current State

| Module        | Status            | Tier               | Reason                           |
| ------------- | ----------------- | ------------------ | -------------------------------- |
| posts         |  Free           | N/A                | Core social features             |
| profiles      |  Free           | N/A                | Core user features               |
| **analytics** | ** Commercial** | **PRO/ENTERPRISE** | **Active: Analytics as premium** |
| warehouse     | 🔮 Planned        | TBD                | Future: Inventory management     |

### Roadmap

**Phase 1 (Current - Active)**:

- Analytics module is **commercial** (PRO/ENTERPRISE tiers)
- Focus on business metrics, reports, dashboards
- Target: SMBs, enterprises needing data insights
- Status:  Implemented and enabled

**Phase 2 (Q2 2026 - Planned)**:

- Release **Audit Log module** (commercial)
- Features: Compliance, GDPR, detailed audit trails
- Target: Regulated industries, enterprises

**Phase 3 (Post Audit Log)**:

- Move **Analytics to free tier**
- Analytics becomes accessible to all users
- Audit Log remains commercial for compliance needs

### Rationale

1. **Analytics First**:  Works with existing data, immediate value
2. **Audit Log Premium**: Compliance is enterprise requirement
3. **Free Analytics**: Broader adoption, upsell to Audit Log

---

## Monitoring & Observability

### License Metrics

Track license usage via logs:

```go
logger.Info("License validated",
    "module", "analytics",
    "tier", license.Tier,
    "expiry", license.ExpiryDate,
    "days_remaining", daysRemaining)
```

### Health Endpoint

License status available via health check:

```bash
curl http://localhost:8080/api/v1/analytics/health

# Response:
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

### Alerts

Set up monitoring for:

- Expired licenses (grace period ending)
- Invalid license attempts
- License verification failures

---

## Testing

### Unit Tests

Each module includes comprehensive license tests:

```bash
# Run all license tests
go test ./internal/modules/*/license/... -v

# Run specific module tests
go test ./internal/modules/analytics/license/... -v
```

Test coverage:

-  Valid license parsing
-  Invalid format detection
-  Signature verification
-  Expiration handling
-  Grace period logic
-  Module mismatch
-  Tier validation

### Integration Tests

Test module initialization with various license states:

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

## Troubleshooting

### Common Issues

**1. License Required Error**

```
Error: license validation failed: license key is required
```

Solution:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
```

**2. Invalid Signature**

```
Error: invalid license signature
```

Causes:

- Wrong `LICENSE_SECRET`
- Tampered license key
- Key copied incorrectly

Solution: Regenerate license with correct secret.

**3. Expired License**

```
Warning: License expired 3 days ago, grace period active
```

Solution: Renew license before grace period ends (7 days default).

**4. Module Mismatch**

```
Error: license module mismatch: expected ANALYTICS, got WAREHOUSE
```

Solution: Use correct license for each module.

### Debug Mode

Enable verbose license logging:

```yaml
# config/app.dev.yaml
log:
  level: debug

modules:
  analytics:
    settings:
      debug_license: true # Log all validation steps
```

### License Validation Tool

Test license validity manually:

```bash
# Validate license
go run ./cmd/license-generator/main.go \
  -validate \
  -key="PROMENADE-ANALYTICS-PRO-20261231-..." \
  -secret="your-secret-key"

# Output:
#  License valid
# Module: ANALYTICS
# Tier: PRO
# Expiry: 2026-12-31
# Days remaining: 365
```

---

## Best Practices

### For Developers

1. **Never commit secrets**: Use `.env` files (gitignored)
2. **Test with invalid licenses**: Ensure error handling works
3. **Document tier features**: Clear feature matrix per tier
4. **Graceful degradation**: Don't crash on license issues in dev
5. **Structured logging**: Log license events for monitoring

### For Operators

1. **Rotate secrets regularly**: Update `LICENSE_SECRET` quarterly
2. **Monitor expiration dates**: Alert 30 days before expiry
3. **Secure secret storage**: Use vault solutions (AWS, HashiCorp)
4. **Track license usage**: Monitor which tiers are active
5. **Plan renewals**: Renew before grace period expires

### For Module Authors

1. **Follow patterns**: Use existing analytics module as template
2. **Document features**: Clear tier-based feature documentation
3. **Test thoroughly**: Comprehensive license validation tests
4. **Health checks**: Include license status in health endpoints
5. **Error messages**: User-friendly, actionable error messages

---

## References

- [Analytics Module README](../internal/modules/analytics/README.md)
- [License Package](../internal/modules/analytics/license/)
- [Module Development Guide](./MODULE_DEVELOPMENT.md)
- [Configuration Architecture](./MODULE_CONFIG_ARCHITECTURE.md)

---

## Appendix

### License Format Specification

```
Format: PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}

Constraints:
- MODULE: [A-Z0-9]+ (uppercase, no spaces)
- TIER: BASIC|PRO|ENTERPRISE
- EXPIRY: YYYYMMDD (valid date)
- SIGNATURE: [A-Za-z0-9_-]+ (base64 URL-encoded)

Max length: 256 characters
Min length: 50 characters
```

### HMAC-SHA256 Implementation

```go
func generateSignature(data, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(data))
    signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
    return strings.TrimRight(signature, "=")  // Remove padding
}

func verifySignature(data, signature, secret string) bool {
    expected := generateSignature(data, secret)
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Testing Checklist

- [ ] Parse valid license
- [ ] Reject invalid format
- [ ] Verify signature
- [ ] Check expiration
- [ ] Test grace period
- [ ] Detect module mismatch
- [ ] Validate tiers
- [ ] Test missing license
- [ ] Test tampered license
- [ ] Integration with module
- [ ] Health check includes status
- [ ] Error messages are clear

---

**Last Updated**: 2024-12-19
**Version**: 1.0.0
**Status**: Production Ready
