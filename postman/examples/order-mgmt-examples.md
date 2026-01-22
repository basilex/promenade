# Order Management API Examples

Examples for **Order Management Context** (`/api/v1/orders`).

---

## 1. Create Order

**Use Case**: Create new order with line items (triggers fulfillment saga)

### Request

```http
POST /api/v1/orders HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
  "status": "pending",
  "currency_code": "USD",
  "items": [
    {
      "product_id": "01932e9a-5678-7abc-9def-0123456789cd",
      "quantity": 10,
      "unit_price": 99.99,
      "discount_percent": 10
    },
    {
      "product_id": "01932e9a-5679-7abc-9def-0123456789ce",
      "quantity": 5,
      "unit_price": 199.99,
      "discount_percent": 0
    }
  ],
  "shipping_address": {
    "street": "456 Oak Ave",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country_code": "US"
  }
}
```

### curl

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
    "status": "pending",
    "currency_code": "USD",
    "items": [
      {
        "product_id": "01932e9a-5678-7abc-9def-0123456789cd",
        "quantity": 10,
        "unit_price": 99.99,
        "discount_percent": 10
      }
    ],
    "shipping_address": {
      "street": "456 Oak Ave",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country_code": "US"
    }
  }'
```

### Response (201 Created)

```json
{
  "success": true,
  "data": {
    "id": "01932e9b-1234-7abc-9def-0123456789ef",
    "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
    "status": "pending",
    "currency_code": "USD",
    "subtotal": 1899.81,
    "discount": 99.99,
    "tax": 180.0,
    "total": 1979.82,
    "items": [
      {
        "id": "01932e9b-2345-7abc-9def-0123456789f0",
        "product_id": "01932e9a-5678-7abc-9def-0123456789cd",
        "quantity": 10,
        "unit_price": 99.99,
        "discount_percent": 10,
        "line_total": 899.91
      },
      {
        "id": "01932e9b-2346-7abc-9def-0123456789f1",
        "product_id": "01932e9a-5679-7abc-9def-0123456789ce",
        "quantity": 5,
        "unit_price": 199.99,
        "discount_percent": 0,
        "line_total": 999.95
      }
    ],
    "fulfillment_status": "pending",
    "created_at": "2026-01-22T10:30:00Z"
  }
}
```

---

## 2. Get Order by ID

**Use Case**: Retrieve order details with fulfillment status

### Request

```http
GET /api/v1/orders/01932e9b-1234-7abc-9def-0123456789ef HTTP/1.1
Host: localhost:8080
```

### curl

```bash
curl http://localhost:8080/api/v1/orders/01932e9b-1234-7abc-9def-0123456789ef
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "01932e9b-1234-7abc-9def-0123456789ef",
    "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
    "status": "confirmed",
    "fulfillment_status": "reserved",
    "total": 1979.82,
    "items": [...],
    "created_at": "2026-01-22T10:30:00Z",
    "updated_at": "2026-01-22T10:35:00Z"
  }
}
```

---

## 3. List Orders

**Use Case**: Get orders with pagination and filtering

### Request

```http
GET /api/v1/orders?customer_id=01932e8f-1234-7abc-9def-0123456789ab&status=confirmed&page=1&page_size=20 HTTP/1.1
Host: localhost:8080
```

### curl

```bash
curl "http://localhost:8080/api/v1/orders?customer_id=01932e8f-1234-7abc-9def-0123456789ab&status=confirmed&page=1&page_size=20"
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "01932e9b-1234-7abc-9def-0123456789ef",
        "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
        "status": "confirmed",
        "total": 1979.82,
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

## 4. Confirm Order (Fulfillment Saga)

**Use Case**: Confirm order (triggers fulfillment saga: inventory reservation → payment → shipping)

### Request

```http
POST /api/v1/orders/01932e9b-1234-7abc-9def-0123456789ef/confirm HTTP/1.1
Host: localhost:8080
```

### curl

```bash
curl -X POST http://localhost:8080/api/v1/orders/01932e9b-1234-7abc-9def-0123456789ef/confirm
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "01932e9b-1234-7abc-9def-0123456789ef",
    "status": "confirmed",
    "fulfillment_status": "reserved",
    "saga_id": "01932e9c-3456-7abc-9def-0123456789f2",
    "updated_at": "2026-01-22T10:35:00Z"
  }
}
```

**Fulfillment Saga Steps**:

1. **Reserve Inventory** (warehouse context)
2. **Process Payment** (billing context)
3. **Ship Order** (warehouse context)

---

## 5. Cancel Order

**Use Case**: Cancel order (triggers saga compensation if already reserved)

### Request

```http
POST /api/v1/orders/01932e9b-1234-7abc-9def-0123456789ef/cancel HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "reason": "Customer requested cancellation"
}
```

### curl

```bash
curl -X POST http://localhost:8080/api/v1/orders/01932e9b-1234-7abc-9def-0123456789ef/cancel \
  -H "Content-Type: application/json" \
  -d '{"reason": "Customer requested cancellation"}'
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "01932e9b-1234-7abc-9def-0123456789ef",
    "status": "cancelled",
    "fulfillment_status": "cancelled",
    "cancellation_reason": "Customer requested cancellation",
    "updated_at": "2026-01-22T11:00:00Z"
  }
}
```

---

## 6. Get Order Fulfillment Status

**Use Case**: Check saga execution progress

### Request

```http
GET /api/v1/orders/01932e9b-1234-7abc-9def-0123456789ef/fulfillment HTTP/1.1
Host: localhost:8080
```

### curl

```bash
curl http://localhost:8080/api/v1/orders/01932e9b-1234-7abc-9def-0123456789ef/fulfillment
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "order_id": "01932e9b-1234-7abc-9def-0123456789ef",
    "saga_id": "01932e9c-3456-7abc-9def-0123456789f2",
    "status": "in_progress",
    "steps": [
      {
        "name": "reserve_inventory",
        "status": "completed",
        "completed_at": "2026-01-22T10:35:30Z"
      },
      {
        "name": "process_payment",
        "status": "in_progress",
        "started_at": "2026-01-22T10:35:35Z"
      },
      {
        "name": "ship_order",
        "status": "pending"
      }
    ]
  }
}
```

---

## Related Contexts

- [Customer Management](customer-mgmt-examples.md) - Customer creation
- [Warehouse](../../internal/contexts/warehouse/README.md) - Inventory reservation
- [Billing](billing-examples.md) - Payment processing
- [Fulfillment Saga](../../docs/concepts/fulfillment-saga.md) - Saga pattern documentation

---

## API Documentation

See [Swagger UI](http://localhost:8080/swagger/index.html) for full API reference.
