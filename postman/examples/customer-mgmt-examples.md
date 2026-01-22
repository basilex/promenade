# Customer Management API Examples

Examples for **Customer Management Context** (`/api/v1/customers`).

---

## 1. Create Customer

**Use Case**: Create a new B2B customer for CRM/invoicing

### Request

```http
POST /api/v1/customers HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "name": "Acme Corporation",
  "email": "contact@acme.com",
  "phone": "+1-555-123-4567",
  "status": "active",
  "address": {
    "street": "123 Main St",
    "city": "San Francisco",
    "state": "CA",
    "postal_code": "94105",
    "country_code": "US"
  },
  "tax_id": "12-3456789",
  "payment_terms_days": 30
}
```

### curl

```bash
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "email": "contact@acme.com",
    "phone": "+1-555-123-4567",
    "status": "active",
    "address": {
      "street": "123 Main St",
      "city": "San Francisco",
      "state": "CA",
      "postal_code": "94105",
      "country_code": "US"
    },
    "tax_id": "12-3456789",
    "payment_terms_days": 30
  }'
```

### Response (201 Created)

```json
{
  "success": true,
  "data": {
    "id": "01932e8f-1234-7abc-9def-0123456789ab",
    "name": "Acme Corporation",
    "email": "contact@acme.com",
    "phone": "+1-555-123-4567",
    "status": "active",
    "address": {
      "street": "123 Main St",
      "city": "San Francisco",
      "state": "CA",
      "postal_code": "94105",
      "country_code": "US"
    },
    "tax_id": "12-3456789",
    "payment_terms_days": 30,
    "created_at": "2026-01-22T10:30:00Z",
    "updated_at": "2026-01-22T10:30:00Z"
  }
}
```

---

## 2. List Customers

**Use Case**: Get paginated list of customers with filtering

### Request

```http
GET /api/v1/customers?page=1&page_size=20&status=active&search=acme HTTP/1.1
Host: localhost:8080
```

### curl

```bash
curl "http://localhost:8080/api/v1/customers?page=1&page_size=20&status=active&search=acme"
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "01932e8f-1234-7abc-9def-0123456789ab",
        "name": "Acme Corporation",
        "email": "contact@acme.com",
        "status": "active",
        "created_at": "2026-01-22T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1,
      "total_pages": 1
    }
  }
}
```

---

## 3. Get Customer by ID

**Use Case**: Retrieve single customer details

### Request

```http
GET /api/v1/customers/01932e8f-1234-7abc-9def-0123456789ab HTTP/1.1
Host: localhost:8080
```

### curl

```bash
curl http://localhost:8080/api/v1/customers/01932e8f-1234-7abc-9def-0123456789ab
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "01932e8f-1234-7abc-9def-0123456789ab",
    "name": "Acme Corporation",
    "email": "contact@acme.com",
    "phone": "+1-555-123-4567",
    "status": "active",
    "address": {
      "street": "123 Main St",
      "city": "San Francisco",
      "state": "CA",
      "postal_code": "94105",
      "country_code": "US"
    },
    "tax_id": "12-3456789",
    "payment_terms_days": 30,
    "created_at": "2026-01-22T10:30:00Z",
    "updated_at": "2026-01-22T10:30:00Z"
  }
}
```

---

## 4. Update Customer

**Use Case**: Update customer details (name, email, status)

### Request

```http
PUT /api/v1/customers/01932e8f-1234-7abc-9def-0123456789ab HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "name": "Acme Corp (Updated)",
  "email": "new-contact@acme.com",
  "status": "active"
}
```

### curl

```bash
curl -X PUT http://localhost:8080/api/v1/customers/01932e8f-1234-7abc-9def-0123456789ab \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp (Updated)",
    "email": "new-contact@acme.com",
    "status": "active"
  }'
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "01932e8f-1234-7abc-9def-0123456789ab",
    "name": "Acme Corp (Updated)",
    "email": "new-contact@acme.com",
    "status": "active",
    "updated_at": "2026-01-22T11:00:00Z"
  }
}
```

---

## 5. Deactivate Customer

**Use Case**: Soft-delete customer (set status to inactive)

### Request

```http
PATCH /api/v1/customers/01932e8f-1234-7abc-9def-0123456789ab/deactivate HTTP/1.1
Host: localhost:8080
```

### curl

```bash
curl -X PATCH http://localhost:8080/api/v1/customers/01932e8f-1234-7abc-9def-0123456789ab/deactivate
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "01932e8f-1234-7abc-9def-0123456789ab",
    "status": "inactive",
    "updated_at": "2026-01-22T11:15:00Z"
  }
}
```

---

## 6. Error Responses

### Customer Not Found (404)

```bash
curl http://localhost:8080/api/v1/customers/invalid-uuid
```

```json
{
  "success": false,
  "error": "customer not found"
}
```

### Validation Error (400)

```bash
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "",
    "email": "invalid-email"
  }'
```

```json
{
  "success": false,
  "error": "validation failed: name is required; email is invalid"
}
```

---

## Related Contexts

- [Interactions](../../internal/contexts/customer-mgmt/interaction/README.md) - Track customer interactions (calls, meetings, notes)
- [Deals](../../internal/contexts/customer-mgmt/deal/README.md) - Sales pipeline management
- [Billing](billing-examples.md) - Create invoices for customers

---

## API Documentation

See [Swagger UI](http://localhost:8080/swagger/index.html) for full API reference.
