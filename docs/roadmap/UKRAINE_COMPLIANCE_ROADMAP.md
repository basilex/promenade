# Ukrainian Compliance Integration Roadmap

**Date**: January 16, 2026  \
**Status**: In progress (PRRO core)  \
**Priority**: Critical for the Ukrainian market  \
**Timeline**: Q1-Q2 2026 (February - June)

---

## Rationale

The ban on Russian software (1C, Bitrix24, AmoCRM) created a major market opportunity:
- **150,000+** SMBs seeking replacements
- **$500M+** CRM/ERP market in Ukraine
- **70%** of 1C users without alternatives

**Without Ukrainian compliance, Promenade cannot compete in this market.**

---

## Phase 5: Ukrainian Compliance Foundation (Q1 2026)

### Priority 1: Fiscal Integration (February 2026, 3-4 weeks)

**Criticality**: Blocker - 70% target market  \
**Timeline**: Week 7-10 (Feb 10 - Mar 3)

#### Week 7-8: Fiscal Core

**Architecture**:
- New context: `internal/contexts/fiscal/`
- Aggregates: CashRegister, Receipt, Shift, Report
- Integrations: Checkbox API client, Vchasno.Kasa API client

**Event Flow**:
Order.Confirm() → order.confirmed event → fiscal receipt creation → fiscal.receipt.printed

**Milestones**:
- Receipt end-to-end flow (create → print → cancel).
- Order event wiring and idempotent handler logic.
- Security-safe error mapping in handlers.

**Testing**:
- Unit tests: 30+
- Integration tests: 15+
- Smoke tests: 10+

#### Week 9-10: Providers and Polish

- Vchasno.Kasa integration
- Error handling and retry logic
- Admin configuration UI
- Documentation and examples
- 10 beta testers (retail)

---

### Priority 2: Tax Invoices (March 2026, 2-3 weeks)

**Criticality**: Blocker - 50% target market (accounting firms)  \
**Timeline**: Week 11-13 (Mar 10 - Mar 30)

**Architecture**:
- New context: `internal/contexts/accounting/`
- Aggregates: TaxInvoice, TaxDeclaration
- Value objects: EDRPOU, TaxNumber, TaxRate
- Integrations: M.E.Doc API client, Cabinet API client

**API**:
- Create, list, get, register, cancel, download XML

---

### Priority 3: Bank Statements (March 2026, 1-2 weeks)

**Criticality**: High (automation)  \
**Timeline**: Week 13-14 (Mar 24 - Apr 6)

**Supported Banks**:
- Monobank
- Privat24 for Business
- PUMB

**Architecture**:
- New context: `internal/contexts/banking/`
- Aggregates: BankAccount, BankStatement, BankTransaction

---

## Phase 6: Ukrainian HRM (Q2 2026)

**Timeline**: Week 14-17 (April 2026, 3-4 weeks)  \
**Priority**: Medium (competitive advantage)

**Features**:
- Employees, contracts, positions
- Time tracking
- Payroll and tax calculations
- Reporting (1DF, Unified Social Contribution)

---

## Phase 7: Delivery Integration (Q2 2026)

**Timeline**: Week 17 (April 2026, 1 week)

**Integrations**:
- Nova Poshta API
- Ukrposhta API

---

## Summary Timeline

Q1 2026:
- Fiscal integration (Checkbox + Vchasno.Kasa)
- Tax invoices (XML generation)
- Bank integrations

Q2 2026:
- HRM (payroll + Ukrainian taxes)
- Delivery integrations
- NATS gateway
- Mobile app (Flutter)

---

## Success Metrics

**Q1 2026**:
- Fiscal integration (2 providers)
- Tax invoices (XML generation)
- Bank integrations (2 banks)
- 20 beta testers

**Q2 2026**:
- HRM with Ukrainian taxes
- Nova Poshta integration
- 100 beta testers
- 10 paying customers
- $1K MRR

---

## Next Steps

1. Complete Phase 3 (LUA + UI Metadata).
2. Finalize PRRO receipt flow (E2E + order events + tests).
3. Integrate scheduler for fiscal retries, shifts, and daily reports.
4. Start tax invoices after PRRO stabilization.
5. Recruit 10 beta testers for fiscal integration.

<!--
*** End Patch
    z_report_number INTEGER,
    z_report_data JSONB,
    
    -- Підсумки
    receipts_count INTEGER DEFAULT 0,
    receipts_total INTEGER DEFAULT 0,           -- копійки
    
    status VARCHAR(20),                         -- 'open', 'closed'
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_receipts_order ON fiscal_receipts(order_id);
CREATE INDEX idx_receipts_shift ON fiscal_receipts(fiscal_shift_id);
CREATE INDEX idx_receipts_created ON fiscal_receipts(created_at DESC);
CREATE INDEX idx_shifts_register ON fiscal_shifts(cash_register_id, opened_at DESC);
```

-->

**API Implementation**:
```go
// POST /api/v1/fiscal/cash-registers
func (h *FiscalHandler) RegisterCashRegister(c *gin.Context) {
    var req struct {
        LocationID uuidv7.UUID `json:"location_id"`
        Name       string      `json:"name"`
        Provider   string      `json:"provider"` // "checkbox", "vchadno"
        APIKey     string      `json:"api_key"`
        APISecret  string      `json:"api_secret"`
    }
    
    // Валідація через провайдера
    if err := h.fiscalService.ValidateCredentials(req.Provider, req.APIKey, req.APISecret); err != nil {
        response.BadRequest(c, "Invalid API credentials")
        return
    }
    
    // Реєстрація ПРРО
    register, err := h.fiscalUC.RegisterCashRegister(ctx, &fiscal.CashRegisterParams{
        LocationID: req.LocationID,
        Name:       req.Name,
        Provider:   req.Provider,
        APIKey:     req.APIKey,
        APISecret:  req.APISecret,
    })
    
    response.Created(c, register)
}

// POST /api/v1/fiscal/receipts
func (h *FiscalHandler) PrintReceipt(c *gin.Context) {
    var req struct {
        OrderID         uuidv7.UUID `json:"order_id"`
        CashRegisterID  uuidv7.UUID `json:"cash_register_id"`
        PaymentType     string      `json:"payment_type"` // "cash", "card"
    }
    
    // Отримати замовлення
    order, err := h.orderUC.GetOrder(ctx, req.OrderID)
    if err != nil {
        response.NotFound(c, "Order not found")
        return
    }
    
    // Створити фіскальний чек
    receipt, err := h.fiscalUC.CreateReceipt(ctx, &fiscal.ReceiptParams{
        OrderID:        req.OrderID,
        CashRegisterID: req.CashRegisterID,
        PaymentType:    req.PaymentType,
        Lines:          convertOrderLinesToFiscalLines(order.Lines),
        TotalAmount:    order.Total.Amount,
    })
    
    // Відправити в ПРРО
    fiscalData, err := h.fiscalService.PrintReceipt(ctx, receipt)
    if err != nil {
        response.InternalError(c, "Failed to print receipt")
        return
    }
    
    // Оновити чек фіскальними даними
    receipt.SetFiscalData(fiscalData.FiscalNumber, fiscalData.FiscalURL, fiscalData.QRCode)
    h.fiscalUC.UpdateReceipt(ctx, receipt)
    
    response.Success(c, receipt)
}

// POST /api/v1/fiscal/shifts/open
func (h *FiscalHandler) OpenShift(c *gin.Context) {
    var req struct {
        CashRegisterID uuidv7.UUID `json:"cash_register_id"`
    }
    
    userID := jwt.GetUserID(c)
    
    shift, err := h.fiscalUC.OpenShift(ctx, req.CashRegisterID, userID)
    if err != nil {
        response.InternalError(c, "Failed to open shift")
        return
    }
    
    response.Created(c, shift)
}

// POST /api/v1/fiscal/shifts/{id}/close
func (h *FiscalHandler) CloseShift(c *gin.Context) {
    shiftID := c.Param("id")
    userID := jwt.GetUserID(c)
    
    // Закрити зміну + сформувати Z-звіт
    shift, zReport, err := h.fiscalUC.CloseShift(ctx, shiftID, userID)
    if err != nil {
        response.InternalError(c, "Failed to close shift")
        return
    }
    
    response.Success(c, gin.H{
        "shift":   shift,
        "z_report": zReport,
    })
}
```

**Checkbox API Integration**:
```go
// pkg/fiscal/checkbox/client.go
package checkbox

type Client struct {
    apiURL    string
    apiKey    string
    apiSecret string
    httpClient *http.Client
}

func (c *Client) PrintReceipt(ctx context.Context, receipt *Receipt) (*FiscalData, error) {
    // Підготовка даних для Checkbox
    payload := map[string]interface{}{
        "payment": map[string]interface{}{
            "type":  receipt.PaymentType, // "CASH", "CARD"
            "value": receipt.TotalAmount,
        },
        "goods": convertLinesToCheckboxGoods(receipt.Lines),
    }
    
    // HTTP запит до Checkbox API
    resp, err := c.post(ctx, "/receipts", payload)
    if err != nil {
        return nil, fmt.Errorf("checkbox API error: %w", err)
    }
    
    // Парсинг відповіді
    var result struct {
        ID          string `json:"id"`
        FiscalCode  string `json:"fiscal_code"`
        FiscalURL   string `json:"fiscal_url"`
        QRCode      string `json:"qr_code"`
    }
    if err := json.Unmarshal(resp, &result); err != nil {
        return nil, err
    }
    
    return &FiscalData{
        FiscalNumber: result.FiscalCode,
        FiscalURL:    result.FiscalURL,
        QRCode:       result.QRCode,
    }, nil
}

func (c *Client) OpenShift(ctx context.Context, cashRegisterID string) error {
    payload := map[string]interface{}{
        "cash_register_id": cashRegisterID,
    }
    _, err := c.post(ctx, "/shifts/open", payload)
    return err
}

func (c *Client) CloseShift(ctx context.Context, shiftID string) (*ZReport, error) {
    resp, err := c.post(ctx, fmt.Sprintf("/shifts/%s/close", shiftID), nil)
    if err != nil {
        return nil, err
    }
    
    var zReport ZReport
    if err := json.Unmarshal(resp, &zReport); err != nil {
        return nil, err
    }
    
    return &zReport, nil
}
```

**Event Flow**:
```
Order.Confirm() 
  → order.confirmed event 
  → FiscalHandler.AutoPrint (якщо налаштовано)
  → fiscal.receipt.printed event
  → Order.SetFiscalReceiptID()
```

**Testing**:
- Unit tests: 30+ (entity, usecase)
- Integration tests: 15+ (з mock API)
- Smoke tests: 10+ (handler validation)

#### Week 9-10: ПРРО Providers & Polish (Лютий 24 - Березень 9)

**Tasks**:
- Вчасно.Каса API integration
- Error handling & retry logic
- Admin UI для налаштування ПРРО
- Документація + приклади
- 10 beta-тестерів (роздрібні магазини)

---

### 🔥 Priority 2: Податкові Накладні (Березень 2026, 2-3 тижні)

**Критичність**: 🔴 БЛОКЕР - 50% target market (бухгалтерські фірми)  
**Timeline**: Week 11-13 (Березень 10-30)

#### Week 11-12: Tax Invoices Core

**Технічна Архітектура**:
```
Новий Context: internal/contexts/accounting/

Aggregates:
  - TaxInvoice (Податкова Накладна)
  - TaxDeclaration (Декларація ПДВ)

Value Objects:
  - EDRPOU (код ЄДРПОУ, 8-10 цифр)
  - TaxNumber (ІПН, 10-12 цифр)
  - TaxRate (0%, 7%, 20%)

Integrations:
  - pkg/accounting/medoc/ (M.E.Doc API client)
  - pkg/accounting/cabinet/ (Cabinet ДПС API)
```

**Database Schema**:
```sql
-- migrations/accounting/000001_tax_invoices.up.sql
CREATE TABLE accounting_tax_invoices (
    id UUID PRIMARY KEY,
    invoice_id UUID REFERENCES billing_invoices(id),
    
    -- Податкова інформація
    tax_invoice_number VARCHAR(50) UNIQUE NOT NULL,
    tax_invoice_date DATE NOT NULL,
    tax_invoice_type VARCHAR(20) NOT NULL,      -- 'standard', 'correction', 'consolidated'
    
    -- Постачальник (наша компанія)
    supplier_edrpou VARCHAR(10) NOT NULL,
    supplier_ipn VARCHAR(12),
    supplier_name VARCHAR(255) NOT NULL,
    supplier_address TEXT,
    
    -- Покупець
    customer_edrpou VARCHAR(10) NOT NULL,
    customer_ipn VARCHAR(12),
    customer_name VARCHAR(255) NOT NULL,
    customer_address TEXT,
    
    -- Суми (в копійках)
    amount_without_vat INTEGER NOT NULL,
    vat_rate INTEGER NOT NULL,                  -- 0, 7, 20
    vat_amount INTEGER NOT NULL,
    total_amount INTEGER NOT NULL,
    
    -- Статуси
    status VARCHAR(20) NOT NULL,                -- 'draft', 'registered', 'cancelled'
    registration_number VARCHAR(50),            -- номер реєстрації в ЄРПН
    registered_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,
    
    -- XML для ДПС
    xml_content TEXT,
    xml_hash VARCHAR(64),
    
    -- Metadata
    notes TEXT,
    created_by UUID REFERENCES identity_users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE accounting_tax_invoice_lines (
    id UUID PRIMARY KEY,
    tax_invoice_id UUID REFERENCES accounting_tax_invoices(id),
    
    -- Товар/послуга
    product_code VARCHAR(20),                   -- УКТЗЕД
    product_name VARCHAR(500) NOT NULL,
    unit VARCHAR(10),                           -- 'шт', 'кг', 'л'
    quantity DECIMAL(10,3) NOT NULL,
    
    -- Ціни (копійки)
    unit_price INTEGER NOT NULL,
    amount_without_vat INTEGER NOT NULL,
    vat_amount INTEGER NOT NULL,
    total_amount INTEGER NOT NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tax_invoices_invoice ON accounting_tax_invoices(invoice_id);
CREATE INDEX idx_tax_invoices_date ON accounting_tax_invoices(tax_invoice_date DESC);
CREATE INDEX idx_tax_invoices_status ON accounting_tax_invoices(status);
CREATE INDEX idx_tax_invoice_lines_invoice ON accounting_tax_invoice_lines(tax_invoice_id);
```

**XML Generation** (згідно з форматом ДПС):
```go
// pkg/accounting/xml/generator.go
package xml

func GenerateTaxInvoiceXML(ti *accounting.TaxInvoice) (string, error) {
    // Формат XML згідно з вимогами ДПС України
    doc := &TaxInvoiceXML{
        XMLName: xml.Name{Local: "DECLAR"},
        XmlnsXsi: "http://www.w3.org/2001/XMLSchema-instance",
        
        // Заголовок
        DECLARBODY: DeclarBody{
            // Загальна інформація
            H01: H01{
                HTIN:   ti.SupplierEDRPOU,        // ЄДРПОУ постачальника
                HNAME:  ti.SupplierName,          // Назва постачальника
                HBOS:   ti.CustomerEDRPOU,        // ЄДРПОУ покупця
                HNAM:   ti.CustomerName,          // Назва покупця
                HSTI:   ti.TaxInvoiceNumber,      // Номер ПН
                HFILL:  ti.TaxInvoiceDate,        // Дата ПН
                HTYPR:  getTypeCode(ti.Type),     // Тип ПН
            },
            
            // Рядки ПН
            R01G3S: convertLinesToXML(ti.Lines),
            
            // Підсумки
            R01G7:  R01G7{
                R03G7:  ti.AmountWithoutVAT,      // Сума без ПДВ
                R04G7:  ti.VATAmount,             // ПДВ
                R07G7:  ti.TotalAmount,           // Всього з ПДВ
            },
        },
    }
    
    // Marshal to XML
    output, err := xml.MarshalIndent(doc, "", "  ")
    if err != nil {
        return "", fmt.Errorf("failed to marshal XML: %w", err)
    }
    
    return string(output), nil
}
```

**M.E.Doc Integration**:
```go
// pkg/accounting/medoc/client.go
package medoc

func (c *Client) RegisterTaxInvoice(ctx context.Context, xmlContent string) (string, error) {
    // Завантаження XML до M.E.Doc
    resp, err := c.post(ctx, "/api/v1/documents/tax_invoice", map[string]interface{}{
        "xml_content": xmlContent,
        "action":      "register",
    })
    
    if err != nil {
        return "", fmt.Errorf("M.E.Doc registration error: %w", err)
    }
    
    var result struct {
        RegistrationNumber string `json:"registration_number"`
        Status             string `json:"status"`
    }
    
    if err := json.Unmarshal(resp, &result); err != nil {
        return "", err
    }
    
    return result.RegistrationNumber, nil
}
```

**API Endpoints**:
```go
// POST /api/v1/accounting/tax-invoices
// GET /api/v1/accounting/tax-invoices
// GET /api/v1/accounting/tax-invoices/:id
// POST /api/v1/accounting/tax-invoices/:id/register (реєстрація в ЄРПН)
// POST /api/v1/accounting/tax-invoices/:id/cancel
// GET /api/v1/accounting/tax-invoices/:id/xml (завантажити XML)
```

#### Week 13: Tax Reports & Documentation

**Tasks**:
- Декларація ПДВ (auto-calculation)
- Експорт звітності у форматі ДПС
- Документація + приклади
- 10 beta-тестерів (бухгалтерські фірми)

---

### 🔥 Priority 3: Банківські Виписки (Березень 2026, 1-2 тижні)

**Критичність**: 🟡 HIGH - автоматизація  
**Timeline**: Week 13-14 (Березень 24 - Квітень 6)

**Підтримувані Банки**:
1. Monobank (Mono API)
2. ПриватБанк (Privat24 for Business API)
3. PUMB (ПУМБ Business API)

**Технічна Архітектура**:
```
Новий Context: internal/contexts/banking/

Aggregates:
  - BankAccount (Банківський рахунок)
  - BankStatement (Виписка)
  - BankTransaction (Транзакція)

Integrations:
  - pkg/banking/monobank/
  - pkg/banking/privat24/
  - pkg/banking/pumb/
```

**Database Schema**:
```sql
-- migrations/banking/000001_banking.up.sql
CREATE TABLE banking_accounts (
    id UUID PRIMARY KEY,
    
    -- Рахунок
    iban VARCHAR(34) UNIQUE NOT NULL,
    account_number VARCHAR(20),
    currency VARCHAR(3) NOT NULL,                -- 'UAH', 'USD', 'EUR'
    
    -- Банк
    bank_name VARCHAR(255),
    bank_mfo VARCHAR(6),
    
    -- Інтеграція
    provider VARCHAR(20) NOT NULL,               -- 'monobank', 'privat24', 'pumb', 'manual'
    api_token_encrypted TEXT,
    
    -- Баланс
    balance INTEGER,                             -- копійки
    balance_updated_at TIMESTAMP,
    
    -- Налаштування
    name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    auto_sync_enabled BOOLEAN DEFAULT false,
    sync_interval_minutes INTEGER DEFAULT 60,
    last_sync_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE banking_transactions (
    id UUID PRIMARY KEY,
    bank_account_id UUID REFERENCES banking_accounts(id),
    
    -- Транзакція
    transaction_id VARCHAR(50),                  -- ID від банку
    transaction_date TIMESTAMP NOT NULL,
    
    -- Сума
    amount INTEGER NOT NULL,                     -- копійки (+ дебет, - кредит)
    currency VARCHAR(3) NOT NULL,
    
    -- Контрагент
    counterparty_name VARCHAR(255),
    counterparty_account VARCHAR(34),           -- IBAN контрагента
    counterparty_bank VARCHAR(255),
    counterparty_mfo VARCHAR(6),
    
    -- Призначення платежу
    description TEXT,
    
    -- Зв'язки
    invoice_id UUID REFERENCES billing_invoices(id),  -- auto-matched
    order_id UUID REFERENCES orders(id),
    matched_at TIMESTAMP,
    matched_by UUID REFERENCES identity_users(id),
    
    -- Metadata
    status VARCHAR(20),                          -- 'pending', 'completed', 'reconciled'
    raw_data JSONB,                             -- повні дані від банку
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_account ON banking_transactions(bank_account_id, transaction_date DESC);
CREATE INDEX idx_transactions_invoice ON banking_transactions(invoice_id);
CREATE INDEX idx_transactions_amount ON banking_transactions(amount);
```

**Monobank API Integration**:
```go
// pkg/banking/monobank/client.go
package monobank

func (c *Client) GetStatements(ctx context.Context, accountID string, from, to time.Time) ([]*Transaction, error) {
    url := fmt.Sprintf("/personal/statement/%s/%d/%d", 
        accountID, 
        from.Unix(), 
        to.Unix(),
    )
    
    resp, err := c.get(ctx, url)
    if err != nil {
        return nil, err
    }
    
    var transactions []MonobankTransaction
    if err := json.Unmarshal(resp, &transactions); err != nil {
        return nil, err
    }
    
    return convertToStandardFormat(transactions), nil
}
```

**Auto-Matching Logic**:
```go
// internal/contexts/banking/usecase.go

func (uc *useCase) AutoMatchTransactions(ctx context.Context, accountID uuidv7.UUID) error {
    // Отримати нові транзакції
    transactions, err := uc.repo.GetUnmatchedTransactions(ctx, accountID)
    if err != nil {
        return err
    }
    
    for _, tx := range transactions {
        // Спробувати знайти рахунок по сумі
        invoice, err := uc.invoiceRepo.FindByAmount(ctx, tx.Amount)
        if err == nil && invoice != nil {
            // Match знайдено!
            invoice.MarkAsPaid(tx.ID, tx.TransactionDate)
            tx.LinkToInvoice(invoice.ID)
            
            // Event
            uc.eventBus.Publish(ctx, bus.TopicInvoicePaid, &InvoicePaidEvent{
                InvoiceID:     invoice.ID,
                TransactionID: tx.ID,
                Amount:        tx.Amount,
                PaidAt:        tx.TransactionDate,
            })
        }
    }
    
    return nil
}
```

**API Endpoints**:
```go
// POST /api/v1/banking/accounts (додати банківський рахунок)
// GET /api/v1/banking/accounts
// POST /api/v1/banking/accounts/:id/sync (синхронізація)
// GET /api/v1/banking/transactions
// POST /api/v1/banking/transactions/:id/match (manual matching)
```

---

## Phase 6: Ukrainian HRM (Q2 2026)

**Timeline**: Week 14-17 (Квітень 2026, 3-4 тижні)  
**Priority**: 💼 MEDIUM - конкурентна перевага

**Функціонал**:
1. Співробітники (картки, договори, посади)
2. Табель робочого часу
3. Нарахування зарплати
4. Розрахунок податків:
   - ПДФО (18%)
   - Військовий збір (1.5%)
   - ЄСВ (22%)
5. Відпускні, лікарняні
6. Звітність:
   - Форма 1ДФ
   - Звіт по ЄСВ
   - Довідки 2-ПДФО

**Context**: `internal/contexts/hrm/`

**Timeline Details**: Див. [UKRAINE_MARKET_STRATEGY_2026.md](UKRAINE_MARKET_STRATEGY_2026.md) (секція 2.1)

---

## Phase 7: Delivery Integration (Q2 2026)

**Timeline**: Week 17 (Квітень 2026, 1 тиждень)  
**Priority**: 🚚 MEDIUM - e-commerce

**Інтеграції**:
1. Нова Пошта API (пріоритет)
2. Укрпошта API

**Функціонал**:
- Створення ТТН
- Розрахунок вартості
- Трекінг відправлень
- Друк етикеток

**Integration**: Додати до Order Context

---

## Summary Timeline

```
Q1 2026 (Січень - Березень):
═══════════════════════════════════════════════════════════
Week 1-2  : ✅ Phase 2 Complete (Domain Errors)
Week 3    : 🔄 Phase 3 Week 2 (Script Storage)
Week 4    : 🔄 Phase 3 Week 3 (UI Metadata)
Week 5-6  : ⏱️ Phase 4 (Scheduler)
Week 7-10 : 🇺🇦 ПРРО Integration (Checkbox + Вчасно.Каса)
Week 11-13: 🇺🇦 Податкові Накладні + Банківські Виписки

Q2 2026 (Квітень - Червень):
═══════════════════════════════════════════════════════════
Week 14-17: 🇺🇦 HRM (Зарплата + Податки)
Week 17   : 🇺🇦 Нова Пошта API
Week 18-20: 🏭 Manufacturing Module (optional)
Week 21-24: ⚡ NATS Gateway
Week 25-26: 📱 Mobile App (Flutter)
```

---

## Success Metrics (Ukrainian Market)

### Q1 2026
- ✅ ПРРО Integration (2 провайдери)
- ✅ Податкові Накладні (XML генерація)
- ✅ Банківські інтеграції (2 банки)
- 🎯 20 beta-тестерів (роздріб + бухгалтерія)

### Q2 2026
- ✅ HRM з українськими податками
- ✅ Нова Пошта integration
- 🎯 100 beta-тестерів
- 🎯 10 paying customers
- 🎯 $1K MRR

### Q3 2026 (Launch)
- 🎯 PUBLIC BETA
- 🎯 200 active users
- 🎯 50 paying customers
- 🎯 $5K MRR

---

## Risk Assessment

### Risk 1: ПРРО APIs нестабільні
**Probability**: MEDIUM  
**Impact**: HIGH  
**Mitigation**:
- Підтримка 2 провайдерів (Checkbox + Вчасно.Каса)
- Retry logic + fallback mechanisms
- Моніторинг API uptime

### Risk 2: Зміни податкового законодавства
**Probability**: MEDIUM  
**Impact**: MEDIUM  
**Mitigation**:
- Модульна архітектура
- Швидкі hotfixes (1-2 дні)
- Моніторинг змін ДПС

### Risk 3: Конкуренти швидше впровадять
**Probability**: LOW-MEDIUM  
**Impact**: HIGH  
**Mitigation**:
- Швидка розробка (3 місяці до MVP)
- Технічна перевага (архітектура)
- Краща ціна

---

## Next Steps

1. ✅ Завершити Phase 3 (LUA + UI Metadata) - 2 тижні
2. ✅ Завершити Phase 4 (Scheduler) - 2 тижні
3. 🚀 **Почати ПРРО Integration** - Week 7 (Лютий 10, 2026)
4. 📝 Зареєструвати ТОВ в Україні
5. 🎯 Знайти 10 beta-тестерів для ПРРО

---

**Автор**: Promenade Team  
**Дата**: 14 січня 2026  
**Статус**: Ready for Implementation 🇺🇦  

**Детальна Стратегія**: [UKRAINE_MARKET_STRATEGY_2026.md](UKRAINE_MARKET_STRATEGY_2026.md)
