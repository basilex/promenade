# Customer Analytics - CQRS Read Models

**Business Intelligence & Reporting** for Customer Management context with real-time dashboards and metrics.

**Status**:  Production-ready (December 2025)  
**Endpoints**: 8 HTTP GET routes  
**Architecture**: CQRS pattern with denormalized queries

---

## Overview

Customer Analytics implements **CQRS read models** optimized for analytical queries. Instead of aggregating data from multiple aggregates at query time, analytics use **denormalized queries** directly against the database for maximum performance.

**Key Features**:
- 9 analytical query methods
- 8 HTTP GET endpoints
- Denormalized SQL queries (joins across aggregates)
- No repository pattern (direct DB access for reads)
- Optimized for reporting and dashboards
- Real-time metrics without pre-aggregation

---

## Architecture

### CQRS Pattern

```

   Write Model (Commands)                
   - Customer CRUD                       
   - Deal CRUD                           
   - Interaction CRUD                    
   - Repository Pattern                  

                  
                   Database
                  

   Read Model (Queries)                  
   - Analytics (this module)             
   - Direct SQL (no repository)          
   - Denormalized views                  
   - Optimized for reporting             

```

**Why CQRS for Analytics?**
- Write models optimize for consistency (aggregates, transactions)
- Read models optimize for query performance (joins, denormalization)
- Analytics queries span multiple aggregates (Customer + Deal + Interaction)
- No need for repository abstraction (read-only, no business logic)

---

## Available Analytics

### 1. Customer Analytics

#### Customer Overview
**GET** `/api/v1/customer-mgmt/analytics/customers/overview`

High-level customer metrics and top sales reps.

**Metrics**:
- Total customers by status (lead, prospect, customer, churned)
- Distribution by tier (free, basic, pro, enterprise)
- Average customer lifetime (weeks)
- Top sales reps by active customers and revenue

#### Customer Lifecycle
**GET** `/api/v1/customer-mgmt/analytics/customers/lifecycle?period=month`

Customer lifecycle funnel with conversion rates.

**Metrics**:
- Lead → Prospect conversion rate
- Prospect → Customer conversion rate
- Customer → Churned churn rate

#### Customer Segmentation
**GET** `/api/v1/customer-mgmt/analytics/customers/segmentation`

Customer distribution by status, tier, and lifetime value.

**Metrics**:
- Total customers
- Distribution by status and tier
- Average lifetime weeks
- Active deals count
- Total revenue across all customers

---

### 2. Deal Analytics

#### Deal Pipeline
**GET** `/api/v1/customer-mgmt/analytics/deals/pipeline`

Sales pipeline with deals by stage and win/loss analysis.

**Metrics**:
- Total deals and value
- Average deal value
- Deals by stage (lead, qualified, proposal, negotiation, closed)
- Win rate (percentage)
- Average days to close

#### Deal Conversions
**GET** `/api/v1/customer-mgmt/analytics/deals/conversions`

Deal stage conversion rates and bottleneck analysis.

**Metrics**:
- Lead → Qualified conversion rate
- Qualified → Proposal conversion rate
- Proposal → Negotiation conversion rate
- Negotiation → Closed conversion rate

---

### 3. Sales Rep Analytics

#### Sales Rep Performance
**GET** `/api/v1/customer-mgmt/analytics/sales-reps/performance?top_n=10`

Top sales reps by revenue and deal count.

**Metrics** (per sales rep):
- Active customers
- Deals won / deals lost
- Total revenue
- Average deal value
- Win rate (percentage)
- Average days to close

---

### 4. Revenue Analytics

#### Revenue Time Series
**GET** `/api/v1/customer-mgmt/analytics/revenue/time-series?start_date=2025-01-01&end_date=2025-12-31&granularity=month`

Revenue over time with new customer tracking.

**Parameters**:
- `start_date`: Start date (YYYY-MM-DD)
- `end_date`: End date (YYYY-MM-DD)
- `granularity`: day, week, or month (default: month)

**Metrics**:
- Revenue per period
- New customers per period
- Growth trends visualization

---

### 5. Interaction Analytics

#### Interaction Insights
**GET** `/api/v1/customer-mgmt/analytics/interactions/insights`

Customer interaction statistics and engagement metrics.

**Metrics**:
- Total interactions
- Distribution by type (call, email, meeting, note)
- Distribution by outcome (successful, not_interested, no_answer, etc.)
- Average duration (minutes)
- Pending follow-ups count

---

## API Examples

### Get Customer Overview

```bash
curl http://localhost:8081/api/v1/customer-mgmt/analytics/customers/overview
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "total_customers": 450,
    "by_status": {
      "lead": 120,
      "prospect": 80,
      "customer": 230,
      "churned": 20
    },
    "by_tier": {
      "free": 200,
      "basic": 150,
      "pro": 80,
      "enterprise": 20
    },
    "avg_lifetime_weeks": 48.5,
    "top_sales_reps": [
      {
        "sales_rep_id": "uuid",
        "active_customers": 45,
        "deals_won": 12,
        "total_revenue": {"cents": 50000000, "currency": "USD"}
      }
    ]
  }
}
```

### Get Deal Pipeline

```bash
curl http://localhost:8081/api/v1/customer-mgmt/analytics/deals/pipeline
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "total_deals": 120,
    "total_value": {"cents": 1500000000, "currency": "USD"},
    "avg_deal_value": {"cents": 12500000, "currency": "USD"},
    "by_stage": {
      "lead": {"count": 30, "value": {"cents": 300000000}},
      "qualified": {"count": 25, "value": {"cents": 400000000}},
      "proposal": {"count": 20, "value": {"cents": 350000000}},
      "negotiation": {"count": 15, "value": {"cents": 250000000}},
      "closed_won": {"count": 20, "value": {"cents": 150000000}},
      "closed_lost": {"count": 10, "value": {"cents": 50000000}}
    },
    "win_rate": 66.7,
    "avg_days_to_close": 45.2
  }
}
```

### Get Revenue Time Series

```bash
curl http://localhost:8081/api/v1/customer-mgmt/analytics/revenue/time-series?start_date=2025-01-01&end_date=2025-12-31&granularity=month
```

**Response**:
```json
{
  "status": "success",
  "data": [
    {
      "period": "2025-01",
      "timestamp": "2025-01-01T00:00:00Z",
      "revenue": {"cents": 125000000, "currency": "USD"},
      "new_customers": 35
    },
    {
      "period": "2025-02",
      "timestamp": "2025-02-01T00:00:00Z",
      "revenue": {"cents": 142000000, "currency": "USD"},
      "new_customers": 42
    }
  ]
}
```

---

## Performance Characteristics

### Query Performance

| Endpoint                   | Avg Time | Complexity | Notes                           |
| -------------------------- | -------- | ---------- | ------------------------------- |
| GetCustomerOverview        | ~15ms    | Medium     | 1 main query + 1 subquery       |
| GetCustomerLifecycle       | ~10ms    | Low        | 1 query with conditional counts |
| GetCustomerSegmentation    | ~20ms    | Medium     | Multiple aggregations           |
| GetDealPipeline            | ~12ms    | Low        | 1 query with stage grouping     |
| GetSalesRepPerformance     | ~25ms    | High       | LEFT JOIN + grouping            |
| GetRevenueTimeSeries       | ~18ms    | Medium     | Time series aggregation         |
| GetInteractionInsights     | ~8ms     | Low        | Simple aggregation              |

**Optimization**:
- Uses indexes on foreign keys (`customer_id`, `assigned_to`)
- Leverages timestamp indexes (`created_at`, `actual_close_date`)
- Denormalized queries avoid N+1 problem
- No repository overhead (direct SQL)

---

## Use Cases

### Executive Dashboard
- Customer Overview: Total customers, status distribution, top sales reps
- Revenue Time Series: Monthly revenue trends, new customer acquisition
- Deal Pipeline: Total pipeline value, win rate, average deal size

### Sales Manager Dashboard
- Sales Rep Performance: Leaderboard by revenue and win rate
- Deal Conversions: Identify bottlenecks in sales funnel
- Customer Lifecycle: Monitor lead → customer conversion

### Marketing Dashboard
- Customer Segmentation: Understand customer distribution by tier
- Interaction Insights: Engagement metrics by type and outcome
- Revenue Time Series: Measure campaign impact on revenue growth

### Customer Success Dashboard
- Customer Lifecycle: Track churn rates and reactivations
- Interaction Insights: Monitor customer touchpoints and follow-ups
- Sales Rep Performance: Account manager effectiveness

---

## Testing

**Test Coverage**: 15 tests (7 unit + 8 integration)

- **Unit Tests**: Structure validation (no DB)
- **Integration Tests**: Full E2E with real database and test data

```bash
# Unit tests only
go test ./internal/contexts/customer-mgmt/analytics -v

# Integration tests
go test ./test/integration/contexts/customer-mgmt/analytics -v
```

---

## Future Enhancements

### Planned Features
- [ ] **Caching**: Redis cache for analytics results (5-15 min TTL)
- [ ] **Date Range Filtering**: Add date filters to all endpoints
- [ ] **Pagination**: Large result sets (e.g., sales rep list)
- [ ] **Export**: CSV/Excel export for reports
- [ ] **Comparative Analysis**: Year-over-year, month-over-month trends
- [ ] **Forecasting**: Predictive analytics based on historical data
- [ ] **Custom Metrics**: User-defined KPIs and calculations
- [ ] **Materialized Views**: Pre-aggregated tables for frequently accessed data
- [ ] **Real-time Updates**: WebSocket push for live dashboard updates

---

## Related Documentation

- [Customer Management](/concepts/customer-management) - Customer aggregate (base data)
- [Deal Management](/concepts/deal-management) - Deal aggregate (pipeline data)
- [Interaction Management](/concepts/interaction-management) - Interaction aggregate (engagement data)
- [CQRS Pattern](https://martinfowler.com/bliki/CQRS.html) - Martin Fowler's CQRS overview

---

**Version**: 1.0.0  
**Status**: Production-ready  
**Test Coverage**: 15 tests, 100% passing  
**Maintainer**: Promenade Team

###  Custom Reports

**Report Builder**:
- Drag-and-drop report designer
- Custom filters and grouping
- Chart types: line, bar, pie, funnel
- Schedule automated delivery

**Export Formats**:
- Excel (.xlsx)
- PDF with charts
- CSV for raw data
- API access for integrations

###  Revenue Forecasting

**Predictive Analytics**:
- Sales pipeline forecast
- Revenue projection by quarter
- Churn prediction model
- Upsell opportunity scoring

**What-If Scenarios**:
- Model different growth rates
- Test pricing strategy impact
- Simulate market conditions

###  Customer Insights

**Segmentation Analysis**:
- RFM (Recency, Frequency, Monetary) scoring
- Cohort analysis
- Customer journey mapping
- Behavioral patterns

**Retention Analysis**:
- Churn prediction
- At-risk customer identification
- Win-back campaign targeting
- Loyalty program effectiveness

###  Sales Analytics

**Pipeline Analysis**:
- Deal stage duration
- Win/loss reasons tracking
- Sales cycle optimization
- Territory performance

**Team Performance**:
- Individual rep metrics
- Quota attainment
- Activity tracking (calls, meetings, demos)
- Leaderboards and gamification

---

## Technical Architecture

### Data Warehouse

**OLAP Database**:
- Separate analytics database (PostgreSQL replica)
- Optimized for read-heavy queries
- Star schema for dimensional modeling
- Materialized views for performance

**ETL Pipeline**:
- Scheduled data extraction (nightly)
- Transform business rules
- Load into analytics tables
- Incremental updates

### Query Engine

**Fast Queries**:
- Pre-aggregated data cubes
- Column-oriented storage
- Query result caching
- Parallel query execution

**Real-Time Streaming**:
- Event-driven updates
- Redis for hot data
- WebSocket live dashboards

### Integration Options

**BI Tools**:
- Metabase (open-source)
- Tableau connector
- Power BI integration
- Looker/Google Data Studio

**Data Export**:
- REST API for custom integrations
- Webhooks for data sync
- Bulk export for data science teams

---

## Use Cases

### SaaS Company

**Daily Operations**:
- Monitor MRR growth
- Track trial-to-paid conversions
- Identify churn risks
- Forecast quarterly revenue

### E-commerce Platform

**Business Intelligence**:
- Analyze customer purchase patterns
- Optimize inventory levels
- Track marketing campaign ROI
- Predict seasonal demand

### B2B Sales Team

**Sales Optimization**:
- Pipeline health monitoring
- Win rate by deal size
- Sales cycle bottlenecks
- Territory rebalancing decisions

---

## Roadmap

### Phase 1: Foundation (Q3 2026)

- [ ] Analytics database setup
- [ ] Basic dashboards (Revenue, Customers, Orders)
- [ ] Report builder MVP
- [ ] Excel/PDF export

### Phase 2: Advanced Analytics (Q4 2026)

- [ ] Predictive models (churn, LTV)
- [ ] Custom KPI tracking
- [ ] Cohort analysis
- [ ] Funnel visualization

### Phase 3: Enterprise Features (Q1 2027)

- [ ] BI tool integrations (Tableau, Power BI)
- [ ] Advanced forecasting
- [ ] Multi-tenant analytics
- [ ] White-label dashboards

---

## Why Wait for This Module?

You can start using Promenade **today** without analytics:

1. **Export Raw Data**: Use API to extract data to Excel/CSV
2. **Third-Party BI**: Connect Metabase or similar to Promenade database
3. **Custom Queries**: Run SQL queries directly on PostgreSQL
4. **Event Streaming**: Track events in real-time via Event Bus

When Analytics module launches (Q3 2026), you'll get:
- Pre-built dashboards for common use cases
- One-click report generation
- No SQL knowledge required
- Mobile-optimized views

---

## Related Documentation

- [Customer Management](customer-management.md) - Data source for analytics
- [Deal Management](deal-management.md) - Sales pipeline data
- [Order Management](order-management.md) - Transaction data
- [Event-Driven Architecture](event-driven.md) - Real-time data streaming

---

**Questions?** [Open a discussion](https://github.com/basilex/promenade/discussions)

**Want this sooner?** [Sponsor development](https://github.com/sponsors/basilex) to prioritize Analytics module.
