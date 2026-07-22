# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 19 — Billing & Subscription Architecture

## Subscription Plans, Usage Metering, Invoice Generation, Payment Processing

---

## 1. Billing Module Overview

The Billing module manages subscription and usage-based billing:

- Subscription plan management (Free, Pro, Enterprise)
- Usage metering (tokens, calls, minutes)
- Invoice generation
- Payment processing (Stripe/Razorpay)
- Dunning management (failed payment retry)
- Credit/debit notes
- Financial reporting
- Multi-currency support

---

## 2. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-billing` |
| Bounded Context | Financial |
| Schema | `billing_` (shared PostgreSQL) |

---

## 3. Subscription Plans

| Plan | Price | AI Requests | Voice Minutes | Documents | Agents |
|---|---|---|---|---|---|
| Free | ₹0/month | 100 | 10 | 10 | 2 |
| Starter | ₹999/month | 5,000 | 100 | 100 | 5 |
| Professional | ₹4,999/month | 50,000 | 1,000 | 1,000 | 20 |
| Enterprise | Custom | Unlimited | Unlimited | Unlimited | Unlimited |

---

## 4. Usage Metering

```rust
pub struct UsageRecord {
    pub tenant_id: TenantId,
    pub service: String,        // llm | stt | tts | telephony | storage
    pub operation: String,      // chat | call | transcription | synthesis
    pub quantity: f64,          // tokens | minutes | characters | bytes
    pub unit_price: Decimal,
    pub total_cost: Decimal,
    pub timestamp: DateTime,
}
```

---

## 5. Invoice Generation

```rust
pub struct Invoice {
    pub invoice_id: InvoiceId,
    pub tenant_id: TenantId,
    pub billing_period: DateRange,
    pub line_items: Vec<LineItem>,
    pub subtotal: Decimal,
    pub tax: Decimal,
    pub total: Decimal,
    pub status: InvoiceStatus,  // draft | sent | paid | overdue | void
    pub due_date: NaiveDate,
    pub paid_at: Option<DateTime>,
}

pub struct LineItem {
    pub description: String,
    pub quantity: f64,
    pub unit_price: Decimal,
    pub amount: Decimal,
}
```

---

## 6. Payment Processing

| Provider | Support | Integration |
|---|---|---|
| Stripe | International | API + Webhooks |
| Razorpay | India | API + Webhooks |
| UPI | India | Via Razorpay |
| Bank Transfer | Manual | Invoice-based |

---

## 7. Dunning Management

```
Payment Failed → Retry (3 attempts, 1/3/7 days)
  → Notify Customer → Suspend Service → Final Notice
  → Account Suspension
```

---

## 8. REST API Endpoints

| Method | Endpoint | Business Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/billing/plans` | `SUCCESS` | `200` |
| `POST` | `/api/v1/billing/subscriptions` | `CREATED` | `201` |
| `POST` | `/api/v1/billing/subscriptions/{id}` | `SUCCESS` | `200` |
| `PATCH` | `/api/v1/billing/subscriptions/{id}` | `UPDATED` | `200` |
| `DELETE` | `/api/v1/billing/subscriptions/{id}` | `UPDATED` | `200` |
| `POST` | `/api/v1/billing/usage?start=...&end=...` | `SUCCESS` | `200` |
| `POST` | `/api/v1/billing/invoices?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/billing/invoices/{id}` | `SUCCESS` | `200` |
| `POST` | `/api/v1/billing/invoices/{id}/pay` | `SUCCESS` | `200` |
| `POST` | `/api/v1/billing/payments?limit=10&offset=0` | `SUCCESS` | `200` |

---

## 9. Database Tables

| Table | Purpose |
|---|---|
| `billing_.plans` | Subscription plans |
| `billing_.subscriptions` | Tenant subscriptions |
| `billing_.usage_records` | Usage metering |
| `billing_.invoices` | Generated invoices |
| `billing_.invoice_line_items` | Invoice line items |
| `billing_.payments` | Payment records |
| `billing_.payment_methods` | Stored payment methods |
| `billing_.dunning_attempts` | Dunning retry attempts |

---

## 10. Ledger Integration

Every financial transaction posts to the double entry ledger:

```sql
-- Invoice Created
INSERT INTO ledger.entries (transaction_id, account_id, entry_type, amount)
VALUES
  (txn_id, accounts_receivable_id, 'debit', 4999.00),
  (txn_id, subscription_revenue_id, 'credit', 4999.00);

-- Payment Received
INSERT INTO ledger.entries (transaction_id, account_id, entry_type, amount)
VALUES
  (txn_id, cash_id, 'debit', 4999.00),
  (txn_id, accounts_receivable_id, 'credit', 4999.00);
```

---

# End of Part 19
