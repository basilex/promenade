# Postman API Examples

This directory contains **context-specific API examples** for common workflows in Promenade Platform.

Each example includes:

- Request details (method, endpoint, body)
- Expected response
- Use case description

---

## Available Examples

| Context             | File                                                   | Use Cases                                           |
| ------------------- | ------------------------------------------------------ | --------------------------------------------------- |
| Customer Management | [customer-mgmt-examples.md](customer-mgmt-examples.md) | Create customer, list customers, update status      |
| Order Management    | [order-mgmt-examples.md](order-mgmt-examples.md)       | Create order, add items, fulfillment saga           |
| Billing             | [billing-examples.md](billing-examples.md)             | Create invoice, record payment, subscription        |
| Accounting          | [accounting-examples.md](accounting-examples.md)       | Create fiscal period, journal entry, reconciliation |
| Warehouse           | [warehouse-examples.md](warehouse-examples.md)         | Create product, stock movement, reservation         |

---

## How to Use

### Option 1: Import Postman Collection

Use the main Postman collection in the parent directory:

```bash
# Import into Postman
postman/Promenade_API.postman_collection.json

# Select environment
- Development: http://localhost:8080
- Staging: https://staging-api.promenade.com
- Production: https://api.promenade.com
```

### Option 2: Use curl Commands

Copy curl commands from example files:

```bash
# Example: Create customer
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "email": "contact@acme.com",
    "status": "active"
  }'
```

### Option 3: Use httpie

```bash
# Install httpie
brew install httpie  # macOS
apt install httpie   # Ubuntu

# Example: Create customer
http POST http://localhost:8080/api/v1/customers \
  name="Acme Corp" \
  email="contact@acme.com" \
  status="active"
```

---

## Related Documentation

- [API Documentation Guide](../../docs/guides/api-documentation.md) - Swagger/OpenAPI
- [API Versioning](../../docs/guides/api-versioning.md) - /api/v1 structure
- [Postman README](../README.md) - Collection setup
