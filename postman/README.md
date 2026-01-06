# Promenade API - Postman Collection

**Complete Postman collection** for Promenade Platform API with 160+ endpoints across 6 bounded contexts.

---

## Quick Start

### 1. Import Collection

**Import the Postman collection**:

1. Open Postman
2. Click **Import** button (top-left)
3. Select **Promenade_API.postman_collection.json**
4. Click **Import**

**Or use URL import**:
```
https://raw.githubusercontent.com/basilex/promenade/dev/postman/Promenade_API.postman_collection.json
```

### 2. Import Environment

**Import environment** (choose one):

- **Development**: `Development.postman_environment.json` (localhost:8081)
- **Staging**: `Staging.postman_environment.json` (staging server)
- **Production**: `Production.postman_environment.json` (production server)

**Steps**:
1. Click **Environments** (left sidebar)
2. Click **Import**
3. Select environment file
4. Click **Import**
5. Select environment from dropdown (top-right)

---

## Authentication Flow

### Step 1: Register User

**Request**: `POST /api/v1/identity/users/register`

**Body**:
```json
{
  "email": "{{user_email}}",
  "name": "Test User",
  "password": "{{user_password}}"
}
```

**Response**: User object with ID

### Step 2: Login

**Request**: `POST /api/v1/identity/users/login`

**Body**:
```json
{
  "email": "{{user_email}}",
  "password": "{{user_password}}"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "expires_at": "2026-01-05T15:30:00Z",
    "token_type": "Bearer",
    "user": { ... }
  }
}
```

**Auto-save tokens** (add to Tests tab):
```javascript
// Save tokens to environment
const response = pm.response.json();
if (response.status === "success" && response.data.access_token) {
    pm.environment.set("access_token", response.data.access_token);
    pm.environment.set("refresh_token", response.data.refresh_token);
    console.log(" Tokens saved to environment");
}
```

### Step 3: Use Protected Endpoints

**All protected endpoints** require `Authorization` header:

```
Authorization: Bearer {{access_token}}
```

Postman auto-injects `{{access_token}}` from environment variables.

---

## Environment Variables

| Variable          | Description                      | Example                          |
| ----------------- | -------------------------------- | -------------------------------- |
| `base_url`        | API base URL                     | `http://localhost:8081`          |
| `api_version`     | API version                      | `v1`                             |
| `access_token`    | JWT access token (auto-saved)    | `eyJhbGciOiJIUzI1NiIsInR5cCI...` |
| `refresh_token`   | JWT refresh token (auto-saved)   | `eyJhbGciOiJIUzI1NiIsInR5cCI...` |
| `user_email`      | Test user email                  | `dev@example.com`                |
| `user_password`   | Test user password               | `DevPassword123`                 |
| `environment`     | Environment name                 | `development`                    |

---

## Collection Structure

### Identity Context (35 endpoints)

**Users** (8 endpoints):
- Register, Login, Logout, Refresh token
- Get by ID, List, Update, Suspend/Ban/Delete

**Contacts** (8 endpoints):
- Create (email/phone/address), Get by ID, List, Update, Delete
- Verify contact, Set as primary

**Profiles** (9 endpoints):
- Create, Get by ID, List, Update, Delete
- List public profiles, Update visibility

**Roles** (5 endpoints):
- Create, Get by ID, List, Update, Delete

**Permissions** (5 endpoints):
- Create, Get by ID, List, Update, Delete

### Customer Management Context (48 endpoints)

**Customers** (14 endpoints):
- Create B2C/B2B, Get by ID, List, Update, Delete
- Lifecycle transitions (Qualify, Convert, Churn)
- Tier management, Tags, Sales rep assignment

**Companies** (14 endpoints):
- Create, Get by ID, List, Update, Delete
- Industry/size management, Parent company, Legal info

**Deals** (12 endpoints):
- Create, Get by ID, List, Update, Delete
- Stage transitions, Mark Won/Lost, Pipeline stats

**Interactions** (8 endpoints):
- Create, Get by ID, List, Update, Delete
- Start/End interaction, Mark completed

### Order Management Context (14 endpoints)

**Orders** (14 endpoints):
- Create, Get by ID, List, Update, Delete
- Add/Remove/Update line items
- State transitions (Confirm, Start Processing, Mark Fulfilled, Cancel)

### Billing Context (22 endpoints)

**Invoices** (8 endpoints):
- Create, Get by ID, List, Update, Delete
- Finalize, Void, Mark Paid/Cancelled

**Payments** (8 endpoints):
- Create, Get by ID, List, Update, Delete
- Process, Mark Failed, Refund

**Subscriptions** (6 endpoints):
- Create, Get by ID, List, Update, Delete
- Activate, Suspend, Resume, Cancel

### Shared Context (9 endpoints)

**Reference Data**:
- Countries (3 endpoints): List, Get by code, Get by ID
- Currencies (3 endpoints): List, Get by code, Get by ID
- Languages (3 endpoints): List, Get by code, Get by ID
- Timezones (3 endpoints): List, Get by name, Get by ID

### Infrastructure (4 endpoints)

**Health Checks**:
- `/health` - Overall system health
- `/health/db` - Database health
- `/health/redis` - Redis health
- `/health/bus` - Event Bus health

---

## Pre-request Scripts

### Auto-refresh Token

**Add to Collection Pre-request Script** to auto-refresh expired tokens:

```javascript
// Check if access token is expired
const accessToken = pm.environment.get("access_token");
if (!accessToken) {
    console.log(" No access token - please login first");
    return;
}

// Decode JWT to check expiration (simple parsing, not validation)
try {
    const payload = JSON.parse(atob(accessToken.split('.')[1]));
    const expiresAt = payload.exp * 1000; // Convert to milliseconds
    const now = Date.now();
    
    // Refresh if token expires in < 1 minute
    if (expiresAt - now < 60000) {
        console.log(" Token expiring soon, refreshing...");
        
        const refreshToken = pm.environment.get("refresh_token");
        const baseUrl = pm.environment.get("base_url");
        const apiVersion = pm.environment.get("api_version");
        
        pm.sendRequest({
            url: `${baseUrl}/api/${apiVersion}/identity/auth/refresh`,
            method: 'POST',
            header: {
                'Content-Type': 'application/json'
            },
            body: {
                mode: 'raw',
                raw: JSON.stringify({
                    refresh_token: refreshToken
                })
            }
        }, (err, res) => {
            if (err) {
                console.error(" Token refresh failed:", err);
            } else {
                const data = res.json();
                if (data.status === "success") {
                    pm.environment.set("access_token", data.data.access_token);
                    pm.environment.set("refresh_token", data.data.refresh_token);
                    console.log(" Token refreshed successfully");
                }
            }
        });
    }
} catch (e) {
    console.error(" Failed to parse token:", e);
}
```

---

## Test Scripts

### Validate Responses

**Add to Collection Tests** for automatic validation:

```javascript
// Check response time
pm.test("Response time < 500ms", () => {
    pm.expect(pm.response.responseTime).to.be.below(500);
});

// Check status code for success endpoints
if (pm.request.method !== "DELETE") {
    pm.test("Status code is 200 or 201", () => {
        pm.expect(pm.response.code).to.be.oneOf([200, 201]);
    });
}

// Validate response structure
pm.test("Response has status field", () => {
    const response = pm.response.json();
    pm.expect(response).to.have.property("status");
    pm.expect(response.status).to.be.oneOf(["success", "error"]);
});

// For success responses, validate data field
if (pm.response.code === 200 || pm.response.code === 201) {
    pm.test("Success response has data field", () => {
        const response = pm.response.json();
        if (response.status === "success") {
            pm.expect(response).to.have.property("data");
        }
    });
}

// For error responses, validate error field
if (pm.response.code >= 400) {
    pm.test("Error response has error field", () => {
        const response = pm.response.json();
        pm.expect(response).to.have.property("error");
        pm.expect(response.error).to.have.property("code");
        pm.expect(response.error).to.have.property("message");
    });
}
```

---

## Common Use Cases

### 1. Complete Authentication Flow

```
1. Register → POST /identity/users/register
2. Login → POST /identity/users/login (saves tokens)
3. Access protected endpoint → GET /identity/users (uses token)
4. Logout → POST /identity/users/logout
```

### 2. Create Customer with Contact

```
1. Login (get token)
2. Create Customer → POST /customer-mgmt/customers
3. Add Contact → POST /identity/contacts (with customer_id)
4. Qualify as Prospect → POST /customer-mgmt/customers/{id}/qualify
```

### 3. Sales Pipeline Flow

```
1. Create Deal → POST /customer-mgmt/deals
2. Move to Qualified → PUT /customer-mgmt/deals/{id}/stage (qualified)
3. Move to Proposal → PUT /customer-mgmt/deals/{id}/stage (proposal)
4. Mark Won → POST /customer-mgmt/deals/{id}/won
```

### 4. Order Processing Flow

```
1. Create Order → POST /order-mgmt/orders
2. Add Line Items → POST /order-mgmt/orders/{id}/lines
3. Confirm Order → POST /order-mgmt/orders/{id}/confirm
4. Start Processing → POST /order-mgmt/orders/{id}/start-processing
5. Mark Fulfilled → POST /order-mgmt/orders/{id}/fulfill
```

### 5. Billing & Subscription Flow

```
1. Create Subscription → POST /billing/subscriptions
2. Activate Subscription → POST /billing/subscriptions/{id}/activate
3. Create Invoice → POST /billing/invoices (for subscription)
4. Create Payment → POST /billing/payments (for invoice)
5. Process Payment → POST /billing/payments/{id}/process
```

---

## Troubleshooting

### Issue 1: "Unauthorized" Error

**Symptoms**: All requests return 401 Unauthorized

**Solution**:
1. Check `{{access_token}}` is set in environment
2. Re-login via `/identity/users/login`
3. Verify token auto-save script in Login request Tests tab
4. Check Authorization header: `Bearer {{access_token}}`

### Issue 2: Token Expired

**Symptoms**: 401 error with message "token expired"

**Solution**:
1. Use Refresh Token endpoint: `POST /identity/auth/refresh`
2. Or re-login: `POST /identity/users/login`
3. Enable auto-refresh in Collection Pre-request Script (see above)

### Issue 3: Environment Variables Not Working

**Symptoms**: `{{variable}}` showing as literal text in requests

**Solution**:
1. Verify environment is selected (dropdown top-right)
2. Check variable names match exactly (case-sensitive)
3. Re-import environment file if needed

### Issue 4: CORS Errors (Browser)

**Symptoms**: Network error, CORS policy blocked

**Solution**:
1. Use Postman Desktop app (not web version)
2. Or enable CORS in server config for dev environment
3. Development server should allow `Access-Control-Allow-Origin: *`

---

## Export & Sharing

### Export Collection

1. Right-click collection name
2. Select **Export**
3. Choose **Collection v2.1** format
4. Save JSON file

### Share with Team

**Option 1: Team Workspace** (Recommended)
- Create Team Workspace in Postman
- Share collection with team members
- Auto-sync changes

**Option 2: Git Repository**
- Commit collection to Git
- Team imports from repository
- Manual sync required

**Option 3: Public Link**
- Publish collection in Postman
- Share read-only link
- View-only access

---

## CI/CD Integration

### Newman CLI

**Install Newman**:
```bash
npm install -g newman
```

**Run collection**:
```bash
newman run Promenade_API.postman_collection.json \
  -e Development.postman_environment.json \
  --reporters cli,json \
  --reporter-json-export results.json
```

**Run in CI pipeline** (.github/workflows/api-tests.yml):
```yaml
- name: Run API Tests
  run: |
    newman run postman/Promenade_API.postman_collection.json \
      -e postman/Development.postman_environment.json \
      --bail
```

---

## Related Documentation

- **API Documentation**: [docs/guides/api-documentation.md](../docs/guides/api-documentation.md)
- **Swagger UI**: http://localhost:8081/api/docs/index.html
- **OpenAPI Spec**: [docs/swagger/swagger.json](../docs/swagger/swagger.json)
- **Authentication**: [pkg/jwt/README.md](../pkg/jwt/README.md)
- **Testing Guide**: [test/README.md](../test/README.md)

---

## External Resources

- [Postman Documentation](https://learning.postman.com/docs/)
- [Postman Variables Guide](https://learning.postman.com/docs/sending-requests/variables/)
- [Newman CLI](https://learning.postman.com/docs/running-collections/using-newman-cli/)
- [Postman Scripts Guide](https://learning.postman.com/docs/writing-scripts/intro-to-scripts/)

---

**Version**: 0.1.0  
**Last Updated**: January 6, 2026  
**Collection**: 160+ endpoints across 6 contexts  
**Maintainer**: Promenade Team
