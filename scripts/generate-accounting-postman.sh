#!/bin/bash
# Generate Postman collection entries for Accounting context

cat > /tmp/accounting-postman.json << 'EOF'
{
  "name": "Accounting",
  "item": [
    {
      "name": "Accounts",
      "item": [
        {
          "name": "Create Account",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"code\": \"1000\",\n  \"name\": \"Cash\",\n  \"account_type\": \"asset\",\n  \"currency_code\": \"UAH\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/accounts",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "accounts"]
            }
          }
        },
        {
          "name": "Get Account by ID",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/accounts/:id",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "accounts", ":id"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "List Accounts",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/accounts",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "accounts"]
            }
          }
        },
        {
          "name": "Update Account",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"name\": \"Updated Account Name\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/accounts/:id",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "accounts", ":id"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Activate Account",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/accounts/:id/activate",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "accounts", ":id", "activate"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Deactivate Account",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/accounts/:id/deactivate",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "accounts", ":id", "deactivate"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        }
      ]
    },
    {
      "name": "Budgets",
      "item": [
        {
          "name": "Create Budget",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"name\": \"Annual Budget 2024\",\n  \"fiscal_year\": 2024\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/budgets",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "budgets"]
            }
          }
        },
        {
          "name": "Get Budget by ID",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/budgets/:id",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "budgets", ":id"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "List Budgets",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/budgets",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "budgets"]
            }
          }
        },
        {
          "name": "Add Budget Line",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"account_id\": \"\",\n  \"amount_cents\": 1000000,\n  \"period\": \"Q1\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/budgets/:id/lines",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "budgets", ":id", "lines"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Approve Budget",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/budgets/:id/approve",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "budgets", ":id", "approve"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Activate Budget",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/budgets/:id/activate",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "budgets", ":id", "activate"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        }
      ]
    },
    {
      "name": "Cost Centers",
      "item": [
        {
          "name": "Create Cost Center",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"code\": \"CC-001\",\n  \"name\": \"IT Department\",\n  \"center_type\": \"cost\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/cost-centers",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "cost-centers"]
            }
          }
        },
        {
          "name": "Get Cost Center by ID",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/cost-centers/:id",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "cost-centers", ":id"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "List Cost Centers",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/cost-centers",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "cost-centers"]
            }
          }
        },
        {
          "name": "Set Manager",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"manager_id\": \"\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/cost-centers/:id/manager",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "cost-centers", ":id", "manager"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Activate Cost Center",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/cost-centers/:id/activate",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "cost-centers", ":id", "activate"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Deactivate Cost Center",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/cost-centers/:id/deactivate",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "cost-centers", ":id", "deactivate"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        }
      ]
    },
    {
      "name": "Fiscal Periods",
      "item": [
        {
          "name": "Create Fiscal Period",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"code\": \"2024-Q1\",\n  \"name\": \"Q1 2024\",\n  \"period_type\": \"quarter\",\n  \"start_date\": \"2024-01-01T00:00:00Z\",\n  \"end_date\": \"2024-03-31T23:59:59Z\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/fiscal-periods",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "fiscal-periods"]
            }
          }
        },
        {
          "name": "Get Fiscal Period by ID",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/fiscal-periods/:id",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "fiscal-periods", ":id"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "List Fiscal Periods",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/fiscal-periods",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "fiscal-periods"]
            }
          }
        },
        {
          "name": "Close Period",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/fiscal-periods/:id/close",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "fiscal-periods", ":id", "close"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Reopen Period",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/fiscal-periods/:id/reopen",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "fiscal-periods", ":id", "reopen"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        }
      ]
    },
    {
      "name": "Journal Entries",
      "item": [
        {
          "name": "Create Journal Entry",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"entry_date\": \"2024-01-15T00:00:00Z\",\n  \"description\": \"Monthly salary payment\",\n  \"source_type\": \"manual\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/journal-entries",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "journal-entries"]
            }
          }
        },
        {
          "name": "Get Journal Entry by ID",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/journal-entries/:id",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "journal-entries", ":id"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "List Journal Entries",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/journal-entries",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "journal-entries"]
            }
          }
        },
        {
          "name": "Add Entry Line",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"account_id\": \"\",\n  \"debit_cents\": 50000,\n  \"credit_cents\": 0,\n  \"currency_code\": \"UAH\",\n  \"description\": \"Debit line\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/journal-entries/:id/lines",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "journal-entries", ":id", "lines"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Post Entry",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/journal-entries/:id/post",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "journal-entries", ":id", "post"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Reverse Entry",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"reverse_description\": \"Reversal entry\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/journal-entries/:id/reverse",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "journal-entries", ":id", "reverse"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        }
      ]
    },
    {
      "name": "Bank Reconciliations",
      "item": [
        {
          "name": "Create Reconciliation",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"bank_account_id\": \"\",\n  \"account_id\": \"\",\n  \"reconciliation_date\": \"2024-01-31T00:00:00Z\",\n  \"statement_date\": \"2024-01-31T00:00:00Z\",\n  \"bank_statement_balance_cents\": 100000,\n  \"book_balance_cents\": 98000,\n  \"currency_code\": \"UAH\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/bank-reconciliations",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "bank-reconciliations"]
            }
          }
        },
        {
          "name": "Get Reconciliation by ID",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/bank-reconciliations/:id",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "bank-reconciliations", ":id"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "List Reconciliations",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/bank-reconciliations",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "bank-reconciliations"]
            }
          }
        },
        {
          "name": "Add Reconciliation Item",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"transaction_type\": \"bank_transaction\",\n  \"transaction_date\": \"2024-01-15T00:00:00Z\",\n  \"description\": \"Bank deposit\",\n  \"amount_cents\": 50000,\n  \"notes\": \"Monthly transfer\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/bank-reconciliations/:id/items",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "bank-reconciliations", ":id", "items"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Complete Reconciliation",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/bank-reconciliations/:id/complete",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "bank-reconciliations", ":id", "complete"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        }
      ]
    },
    {
      "name": "Tax Codes",
      "item": [
        {
          "name": "Create Tax Code",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"code\": \"VAT20\",\n  \"name\": \"VAT 20%\",\n  \"tax_type\": \"vat\",\n  \"rate\": 2000\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/tax-codes",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "tax-codes"]
            }
          }
        },
        {
          "name": "Get Tax Code by ID",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/tax-codes/:id",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "tax-codes", ":id"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "List Tax Codes",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/tax-codes",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "tax-codes"]
            }
          }
        },
        {
          "name": "Update Tax Code",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"rate\": 1800\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/tax-codes/:id",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "tax-codes", ":id"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Set Tax Payable Account",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"account_id\": \"\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/tax-codes/:id/payable-account",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "tax-codes", ":id", "payable-account"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Activate Tax Code",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/tax-codes/:id/activate",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "tax-codes", ":id", "activate"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        },
        {
          "name": "Deactivate Tax Code",
          "request": {
            "method": "PUT",
            "header": [
              {"key": "Authorization", "value": "Bearer {{access_token}}"}
            ],
            "url": {
              "raw": "{{base_url}}/api/v1/accounting/tax-codes/:id/deactivate",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "accounting", "tax-codes", ":id", "deactivate"],
              "variable": [{"key": "id", "value": ""}]
            }
          }
        }
      ]
    }
  ]
}
EOF

echo "Accounting Postman collection fragment generated at /tmp/accounting-postman.json"
echo ""
echo "To add to main collection:"
echo "1. Open Promenade_API.postman_collection.json"
echo "2. Find the 'item' array at root level"
echo "3. Add the content from /tmp/accounting-postman.json to the items array"
echo "4. Import updated collection to Postman"
