# Audit IModule (Commercial)

> **License Required**: This is a commercial module requiring a valid license key for production use.

Immutable audit logging system with cryptographic signatures for tamper detection and compliance.

## Features

-  **Immutable Audit Log**: Records cannot be modified after creation
-  **Cryptographic Signatures**: HMAC-SHA256 signatures for tamper detection
-  **Flexible Filtering**: Search by user, action, entity, date range
-  **Signature Verification**: Verify event integrity at any time
-  **Retention Policies**: Automatic purge based on license tier
-  **Performance Optimized**: Indexed for common query patterns

## License Tiers

| Tier           | Retention | Features                                   | Price      |
| -------------- | --------- | ------------------------------------------ | ---------- |
| **BASIC**      | 30 days   | Core audit logging, signature verification | $99/month  |
| **PRO**        | 1 year    | + Advanced filtering, custom reports       | $299/month |
| **ENTERPRISE** | Unlimited | + Compliance exports, priority support     | Custom     |

## Configuration

### Development (No License Required)

```yaml
# internal/modules/audit/config/config.dev.yaml
module:
  enabled: true
  license_required: false

audit:
  signature_secret: "dev-audit-secret-key"
```

### Production (License Required)

```yaml
# internal/modules/audit/config/config.prod.yaml
module:
  enabled: true
  license_required: true

audit:
  signature_secret: "" # Set via AUDIT_SECRET env var
```

**Environment Variables:**

```bash
export AUDIT_LICENSE_KEY="PROMENADE-AUDIT-PRO-20261231-<signature>"
export AUDIT_SECRET="your-secure-signature-secret"
export LICENSE_SECRET="your-license-validation-secret"
```

## Obtaining a License

### Generate License (For Vendors)

```bash
# Generate PRO license valid for 1 year
./scripts/generate-license.sh audit PRO 365

# Or using the tool
go run ./cmd/license-generator/main.go \
  -module=audit \
  -tier=PRO \
  -expiry=20261231 \
  -secret="$LICENSE_SECRET"
```

### For Customers

Contact sales@example.com or visit https://example.com/pricing

## API Endpoints

All endpoints require authentication (`Authorization: Bearer <token>`).

### Create Audit Event

```http
POST /api/v1/audit/events
Content-Type: application/json

{
  "user_id": "01234567-89ab-cdef-0123-456789abcdef",
  "action": "user.login",
  "entity_type": "user",
  "entity_id": "01234567-89ab-cdef-0123-456789abcdef",
  "changes": {"ip": "192.168.1.1"},
  "ip_address": "192.168.1.1",
  "request_id": "req-12345"
}
```

### List Audit Events

```http
GET /api/v1/audit/events?user_id=<uuid>&action=user.login&page=1&page_size=20
```

### Get Audit Event

```http
GET /api/v1/audit/events/:id
```

### Verify Signature

```http
GET /api/v1/audit/events/:id/verify
```

### Get Events by Entity

```http
GET /api/v1/audit/events/entity/:entity_type/:entity_id
```

## Usage Example

```go
import (
    "github.com/basilex/promenade/internal/modules/audit/domain/entity"
    "github.com/basilex/promenade/internal/modules/audit/usecase"
)

// Create audit event
data := &entity.AuditEventCreate{
    UserID:     userID,
    Action:     "post.create",
    EntityType: "post",
    EntityID:   postID,
    Changes: map[string]interface{}{
        "title": "New Post",
        "status": "published",
    },
    IPAddress:  &ipAddr,
    RequestID:  &reqID,
}

event, err := auditUseCase.CreateAuditEvent(ctx, data)
if err != nil {
    return err
}

// Verify signature later
valid, err := auditUseCase.VerifyAuditEvent(ctx, event.ID)
```

## Common Actions

Standard action naming convention: `{entity}.{operation}`

**Authentication:**

- `user.login` - User login
- `user.logout` - User logout
- `user.password_reset` - Password reset

**User Management:**

- `user.create` - User created
- `user.update` - User updated
- `user.delete` - User deleted
- `user.suspend` - User suspended

**Content:**

- `post.create` - Post created
- `post.update` - Post updated
- `post.delete` - Post deleted
- `post.publish` - Post published

**Permissions:**

- `role.assign` - Role assigned
- `role.revoke` - Role revoked
- `permission.grant` - Permission granted
- `permission.revoke` - Permission revoked

## Database Schema

```sql
CREATE TABLE audit_events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    changes JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    metadata JSONB,
    signature VARCHAR(128) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Optimized indexes for common queries
CREATE INDEX idx_audit_events_user_created ON audit_events(user_id, created_at DESC);
CREATE INDEX idx_audit_events_entity_created ON audit_events(entity_type, entity_id, created_at DESC);
```

## Retention & Purge

Automatic purge runs daily based on license tier:

- **BASIC**: 30 days retention
- **PRO**: 365 days retention
- **ENTERPRISE**: Unlimited (manual purge only)

Purge schedule configured in `config.{env}.yaml`:

```yaml
purge:
  enabled: true
  schedule: "0 3 * * *" # 3 AM daily
```

## Security Considerations

1. **Signature Secret**: Use strong, random secret (32+ characters)
2. **Database Security**: Audit table should have restricted DELETE permissions
3. **Access Control**: Limit audit log access to administrators only
4. **Encryption**: Enable at-rest encryption for audit database
5. **Backup**: Separate backup schedule for audit logs

## Compliance

Meets requirements for:

- GDPR Article 30 (Records of processing activities)
- HIPAA §164.312(b) (Audit controls)
- SOC 2 Type II (Logging and monitoring)
- PCI DSS 10 (Track and monitor access)

## Troubleshooting

### License Errors

```
Error: license validation failed: license expired
```

**Solution**: Renew license or set `license_required: false` in dev config.

### Signature Verification Failed

```
Error: signature verification failed
```

**Possible causes:**

1. `AUDIT_SECRET` changed after event creation
2. Event data was tampered with
3. Database corruption

**Resolution**: Check signature secret and verify database integrity.

### Performance Issues

For high-volume systems (>10K events/day):

1. **Partition table by date**:

```sql
CREATE TABLE audit_events_2024_01 PARTITION OF audit_events
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

2. **Use separate database** for audit logs

3. **Async event creation** via event bus

## Support

- Documentation: https://docs.example.com/modules/audit
- Issues: https://github.com/basilex/promenade/issues
- Email: support@example.com

---

**IModule Version**: 1.0.0
**License**: Commercial (Proprietary)
**Requires**: Promenade Core 1.0.0+
