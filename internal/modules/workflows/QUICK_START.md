# Workflows Module - Quick Start Guide

**Time to First Workflow:** 5 minutes ⏱️  
**Difficulty:** Beginner 🟢

---

## What You'll Build

A simple **document approval workflow** with 3 states:

1. `draft` - Document is being prepared
2. `pending_approval` - Waiting for manager review
3. `approved` / `rejected` - Final state

---

## Prerequisites

✅ Promenade server running (`make dev`)  
✅ Valid authentication token  
✅ Workflows module enabled and licensed

---

## Step 1: Get Authentication Token (30 seconds)

```bash
# Login as system admin
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "system@promenade.com",
    "password": "passw0rd"
  }' | jq -r '.data.token')

echo "Token: $TOKEN"
```

**Expected output:**

```
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Step 2: Create Workflow Definition (1 minute)

```bash
curl -X POST http://localhost:8081/api/v1/workflows/definitions \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "simple_approval",
    "display_name": "Simple Document Approval",
    "description": "Basic approval workflow for documents",
    "category": "approval",
    "schema": {
      "states": [
        "draft",
        "pending_approval",
        "approved",
        "rejected"
      ],
      "initial_state": "draft",
      "transitions": [
        {
          "from": "draft",
          "to": "pending_approval",
          "event": "submit"
        },
        {
          "from": "pending_approval",
          "to": "approved",
          "event": "approve"
        },
        {
          "from": "pending_approval",
          "to": "rejected",
          "event": "reject"
        }
      ]
    }
  }' | jq '.'
```

**Expected output:**

```json
{
  "status": "success",
  "data": {
    "id": "019b4bbf-86b5-7a4d-a943-7a3ea3c322f9",
    "name": "simple_approval",
    "display_name": "Simple Document Approval",
    "version": "1.0.0",
    "status": "DRAFT",
    "schema": { ... }
  }
}
```

**📝 Save the workflow ID for next steps!**

```bash
WORKFLOW_ID="019b4bbf-86b5-7a4d-a943-7a3ea3c322f9"  # Use your actual ID
```

---

## Step 3: Activate Workflow (30 seconds)

```bash
curl -X POST http://localhost:8081/api/v1/workflows/definitions/$WORKFLOW_ID/activate \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

**Expected output:**

```json
{
  "status": "success",
  "data": {
    "id": "019b4bbf-86b5-7a4d-a943-7a3ea3c322f9",
    "status": "ACTIVE",
    ...
  }
}
```

**✅ Your workflow is now ready to use!**

---

## Step 4: Start Workflow Instance (1 minute)

```bash
curl -X POST http://localhost:8081/api/v1/workflows/instances \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "definition_id": "'$WORKFLOW_ID'",
    "priority": 5,
    "external_reference": "DOC-2025-001",
    "metadata": {
      "document_title": "Q4 Budget Report",
      "author": "John Doe",
      "department": "Finance"
    }
  }' | jq '.'
```

**Expected output:**

```json
{
  "status": "success",
  "data": {
    "id": "019b4bc0-1111-7111-9111-111111111111",
    "definition_id": "019b4bbf-86b5-7a4d-a943-7a3ea3c322f9",
    "current_state": "draft",
    "status": "RUNNING",
    "external_reference": "DOC-2025-001",
    "state_history": [
      {
        "from_state": "",
        "to_state": "draft",
        "event": "start",
        "timestamp": "2025-12-26T10:00:00Z"
      }
    ]
  }
}
```

**📝 Save the instance ID!**

```bash
INSTANCE_ID="019b4bc0-1111-7111-9111-111111111111"  # Use your actual ID
```

---

## Step 5: Test Workflow Transitions (2 minutes)

### Transition 1: Submit for Approval

```bash
curl -X POST http://localhost:8081/api/v1/workflows/instances/$INSTANCE_ID/transition \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "event": "submit"
  }' | jq '.data | {current_state, status}'
```

**Expected output:**

```json
{
  "current_state": "pending_approval",
  "status": "RUNNING"
}
```

### Transition 2: Approve Document

```bash
curl -X POST http://localhost:8081/api/v1/workflows/instances/$INSTANCE_ID/transition \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "event": "approve"
  }' | jq '.data | {current_state, status}'
```

**Expected output:**

```json
{
  "current_state": "approved",
  "status": "COMPLETED"
}
```

**🎉 Congratulations! You've completed your first workflow!**

---

## Step 6: View Workflow History (30 seconds)

```bash
curl -X GET http://localhost:8081/api/v1/workflows/instances/$INSTANCE_ID \
  -H "Authorization: Bearer $TOKEN" | jq '.data.state_history'
```

**Expected output:**

```json
[
  {
    "from_state": "",
    "to_state": "draft",
    "event": "start",
    "timestamp": "2025-12-26T10:00:00Z"
  },
  {
    "from_state": "draft",
    "to_state": "pending_approval",
    "event": "submit",
    "timestamp": "2025-12-26T10:01:00Z"
  },
  {
    "from_state": "pending_approval",
    "to_state": "approved",
    "event": "approve",
    "timestamp": "2025-12-26T10:02:00Z"
  }
]
```

---

## Visual Workflow Diagram

```
┌───────┐
│ draft │
└───┬───┘
    │ submit
    ▼
┌─────────────────┐
│ pending_approval│
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
 approve   reject
    │         │
    ▼         ▼
┌─────────┐ ┌──────────┐
│approved │ │ rejected │
└─────────┘ └──────────┘
```

---

## What's Next?

### Try These Examples

1. **[Multi-Stage Approval](examples/multi_stage_approval.md)** - Manager + Finance approval
2. **[Retry Loop](examples/retry_loop.md)** - Automatic retry on failure
3. **[Complex Branching](examples/complex_branching.md)** - Conditional paths

### Learn More

- **[Complete API Reference](README.md#api-reference)** - All endpoints
- **[Design Best Practices](DESIGN_GUIDE.md)** - How to design good workflows
- **[Validation Guide](VALIDATION_ARCHITECTURE.md)** - Understanding validation
- **[Testing Guide](TEST_STATUS.md)** - How to test workflows

---

## Common Issues

### Issue: "Workflow not found"

**Cause:** Wrong workflow ID or workflow not activated

**Solution:**

```bash
# List all workflows
curl -X GET http://localhost:8081/api/v1/workflows/definitions?status=ACTIVE \
  -H "Authorization: Bearer $TOKEN" | jq '.data[].id'
```

### Issue: "Invalid transition"

**Cause:** Event not allowed from current state

**Solution:**

```bash
# Check current state
curl -X GET http://localhost:8081/api/v1/workflows/instances/$INSTANCE_ID \
  -H "Authorization: Bearer $TOKEN" | jq '.data.current_state'

# Check allowed transitions from definition
curl -X GET http://localhost:8081/api/v1/workflows/definitions/$WORKFLOW_ID \
  -H "Authorization: Bearer $TOKEN" | jq '.data.schema.transitions'
```

### Issue: "Validation failed"

**Cause:** Invalid workflow schema

**Solution:** Check validation errors:

- Unreachable states
- Cycles without exits
- Invalid state references

See **[Troubleshooting Guide](README.md#troubleshooting)** for details.

---

## Quick Reference

### Essential Commands

```bash
# List workflows
curl -X GET http://localhost:8081/api/v1/workflows/definitions \
  -H "Authorization: Bearer $TOKEN"

# Get workflow details
curl -X GET http://localhost:8081/api/v1/workflows/definitions/$WORKFLOW_ID \
  -H "Authorization: Bearer $TOKEN"

# List instances
curl -X GET http://localhost:8081/api/v1/workflows/instances?definition_id=$WORKFLOW_ID \
  -H "Authorization: Bearer $TOKEN"

# Get instance details
curl -X GET http://localhost:8081/api/v1/workflows/instances/$INSTANCE_ID \
  -H "Authorization: Bearer $TOKEN"

# Transition instance
curl -X POST http://localhost:8081/api/v1/workflows/instances/$INSTANCE_ID/transition \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"event": "approve"}'

# Cancel instance
curl -X POST http://localhost:8081/api/v1/workflows/instances/$INSTANCE_ID/cancel \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"reason": "Cancelled by user"}'
```

---

## Support

**Questions?** Check the [FAQ](README.md#faq)  
**Found a bug?** See [Troubleshooting](README.md#troubleshooting)  
**Need help?** Contact: alexander.vasilenko@gmail.com

---

**Next:** [Design Best Practices →](DESIGN_GUIDE.md)

**Time spent:** ~5 minutes ✅  
**Workflows created:** 1 ✅  
**Instances executed:** 1 ✅  
**Ready for production:** Let's go! 🚀
