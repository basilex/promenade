# Warehouse IModule - Commercial

Enterprise-grade warehouse and inventory management system.

## Features

 **Inventory Management**

- Item tracking with SKU, barcode, location
- Stock levels, min/max quantities
- Categories and suppliers
- Real-time inventory updates

 **Stock Operations**

- Stock in (receiving)
- Stock out (issuing)
- Stock transfers between locations
- Inventory adjustments

 **Reporting**

- Inventory reports
- Movement history
- Low stock alerts
- Stock valuation

 **Barcode Integration**

- Barcode scanner support
- Generate/print barcodes
- Quick stock operations

 **Automation**

- Low stock monitoring
- Auto-reorder suggestions
- Email notifications

---

## Installation

### 1. Purchase License

Visit [https://promenade.dev/modules/warehouse](https://promenade.dev) to purchase a license.

### 2. Enable IModule

Add to `config/modules.yaml`:

```yaml
modules:
  enabled:
    - warehouse

  config:
    warehouse:
      version: "1.2.0"
      license_key: "WH-YOUR-LICENSE-KEY-HERE"
      settings:
        max_items: 10000
        enable_barcode_scanner: true
        low_stock_threshold: 10
```

### 3. Import IModule

Uncomment in `cmd/api/main.go`:

```go
import (
    _ "github.com/basilex/promenade/internal/modules/warehouse"
)
```

### 4. Apply Migrations

```bash
make migrate-up
```

---

## API Endpoints

### Items

```
GET    /api/v1/warehouse/items         # List all items
POST   /api/v1/warehouse/items         # Create item
GET    /api/v1/warehouse/items/:id     # Get item
PUT    /api/v1/warehouse/items/:id     # Update item
DELETE /api/v1/warehouse/items/:id     # Delete item
```

### Stock Operations

```
POST   /api/v1/warehouse/stock/in       # Receive stock
POST   /api/v1/warehouse/stock/out      # Issue stock
POST   /api/v1/warehouse/stock/transfer # Transfer between locations
```

### Reports

```
GET    /api/v1/warehouse/reports/inventory  # Inventory report
GET    /api/v1/warehouse/reports/movements  # Movement history
GET    /api/v1/warehouse/reports/low-stock  # Low stock items
```

### Barcode (if enabled)

```
POST   /api/v1/warehouse/scan           # Scan barcode
```

---

## Permissions

```
warehouse:items:read     # View items
warehouse:items:create   # Create items
warehouse:items:update   # Update items
warehouse:items:delete   # Delete items
warehouse:stock:in       # Receive stock
warehouse:stock:out      # Issue stock
warehouse:stock:transfer # Transfer stock
warehouse:reports:read   # View reports
```

---

## Configuration

```yaml
warehouse:
  license_key: "WH-XXX" # Required
  settings:
    max_items: 10000 # Maximum items (default: 10000)
    enable_barcode_scanner: true # Enable barcode features
    low_stock_threshold: 10 # Low stock alert threshold
    auto_reorder: false # Auto-generate purchase orders
```

---

## Database Schema

### warehouse_items

```sql
id              UUID PRIMARY KEY
sku             VARCHAR(100) UNIQUE NOT NULL
name            VARCHAR(255) NOT NULL
description     TEXT
category        VARCHAR(100)
quantity        INTEGER DEFAULT 0
min_quantity    INTEGER DEFAULT 0
max_quantity    INTEGER
unit_price      DECIMAL(10,2)
barcode         VARCHAR(50)
location        VARCHAR(100)
supplier        VARCHAR(255)
created_at      TIMESTAMP
updated_at      TIMESTAMP
```

### warehouse_movements

```sql
id              UUID PRIMARY KEY
item_id         UUID REFERENCES warehouse_items
type            VARCHAR(20)  -- in, out, transfer, adjustment
quantity        INTEGER
from_location   VARCHAR(100)
to_location     VARCHAR(100)
reason          TEXT
user_id         UUID REFERENCES users
created_at      TIMESTAMP
```

---

## Support

- Email: support@promenade.dev
- Docs: https://docs.promenade.dev/modules/warehouse
- License: Commercial (1 year)
- Price: $99/month or $999/year

---

## Changelog

### v1.2.0 (Current)

-  Barcode scanner integration
-  Low stock monitoring
- 🐛 Fixed stock transfer validation
- 📚 Improved documentation

### v1.1.0

-  Multi-location support
-  Stock movement reports
- 🐛 Fixed quantity calculations

### v1.0.0

-  Initial release
-  Basic CRUD operations
-  Stock in/out operations
