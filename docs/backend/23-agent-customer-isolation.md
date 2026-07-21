# AeroXe Nexus AI — Agent-Customer Isolation

## Customer Data Isolation + Cross-Platform Boundaries + Sensitive Info Protection

> **Modular Monolith Context:** Customer isolation is enforced at multiple layers within the `aeroxe-nexus` binary. The `identity` module (`src/modules/identity/`) handles auth, `agent` module (`src/modules/agent/`) enforces scope, `customer` module (`src/modules/customer/`) manages customer data, and `audit` module (`src/modules/audit/`) logs all access. All cross-module isolation checks use trait interfaces, not network calls. All database access uses SeaORM — no raw SQL. See [Security Architecture](14-security-architecture.md).

---

## 1. Overview

This document defines how agents isolate customer data across platforms and tenants, with real-world scenarios showing correct agent behavior.

### Core Principles

| Principle | Description |
|---|---|
| Guest Access | Unregistered users can ask general purpose inquiries |
| Customer Isolation | Authenticated customers only access their own data |
| Platform Isolation | Agent only accesses data for the current platform/tenant |
| Identity Verification | Sensitive data requires identity validation before disclosure |
| Sensitive Data Refusal | Passwords, secrets, tokens must NEVER be disclosed |
| Cross-Platform Redirect | Wrong platform → redirect to correct platform, not block |

---

## 2. Isolation Architecture

### 2.1 Four-Layer Isolation

```
Layer 0: Guest/Unregistered Access
  → Unregistered users can ask general purpose inquiries
  → No personal data, no account info, no sensitive data
  → FAQs, pricing, plans, general product info only

Layer 1: Platform/Tenant Isolation
  → Agent only operates within its assigned platform

Layer 2: Customer Isolation
  → Within a platform, agent only accesses current customer's data

Layer 3: Data Sensitivity Classification
  → Some data requires additional identity verification before disclosure
```

### 2.2 Isolation Enforcement Flow

```
Customer Request
  → Extract context:
      - platform_id (from session/JWT)
      - customer_id (from session/JWT, may be NULL for guests)
      - user_id (from session/JWT, may be NULL for guests)
        → Is user authenticated?
          → NO (Guest):
              → Is request general purpose? (FAQ, pricing, plans, product info)
                → YES: Respond with public information
                → NO (personal/sensitive): "Please log in to access this information"
          → YES (Authenticated):
              → Validate platform membership
                → Is customer registered on this platform?
                  → YES: Proceed with customer scope
                  → NO: "You are not registered on this platform"
                    → Validate request scope
                      - Does request match customer's own data?
                      - Is customer trying to access another customer's data?
                        → Validate data sensitivity
                          - Public data: Return directly
                          - Private data: Return after login verification
                          - Sensitive data: Refuse or require MFA
                            → Return response within scope
```

---

## 3. Platform Isolation Rules

### 3.1 Rule: Agent Only Responds for Its Platform

Each agent is bound to a specific platform/tenant. It has NO access to other platforms.

| Scenario | Correct Agent Response |
|---|---|
| User on Platform A asks about Platform B | "You need to check with Platform B. I can only help you with Platform A." |
| User on Platform A asks about Platform B's data | "I don't have access to Platform B's information. Please contact Platform B support." |
| Unregistered user asks general question | Answer directly (plans, pricing, FAQs, product info) |
| Unregistered user asks personal data | "Please log in to access your account information." |

### 3.2 Example: Cross-Platform Request

```
Context:
  - User: Ali (user_id: 101)
  - Current Platform: AeroXe Broadband
  - Ali is also registered on: AeroXe ERP

User: "What is my invoice amount?"

Agent thinks:
  1. User is on AeroXe Broadbound platform
  2. Query invoices for this customer on THIS platform only
  3. Do NOT query AeroXe ERP

Agent response:
  "Your current outstanding balance on AeroXe Broadband is ₹2,450.
   Your last payment was ₹1,200 on January 10, 2026.
   Would you like me to help with anything else?"

NOT:
  "Your AeroXe ERP invoice is ₹15,000" ← WRONG, different platform
```

---

## 4. Customer Isolation Rules

### 4.1 Rule: Agent Only Responds for the Authenticated Customer

| Scenario | Correct Agent Response |
|---|---|
| Customer A asks about Customer B | "I can only provide information about your own account." |
| Customer A asks "What is Customer B's balance?" | "I cannot share other customers' information. This is private data." |
| Customer A asks "Show me all customers' data" | "I can only show your own account information." |

### 4.2 Example: Customer Data Request

```
Context:
  - User: Priya (customer_id: 201)
  - Platform: AeroXe Broadband
  - Priya is a residential customer

User: "What is Rahul's account status?"

Agent thinks:
  1. Priya is asking about another customer (Rahul)
  2. Customer isolation: Priya can only see HER data
  3. Reject request

Agent response:
  "I can only provide information about your own account.
   For other accounts, please contact our support team."
```

---

## 5. Identity Verification for Sensitive Data

### 5.1 Data Sensitivity Levels

| Level | Examples | Disclosure Rule |
|---|---|---|
| **Public (Guest OK)** | FAQs, pricing, plans, product features, general support | Return directly, no login required |
| **Private** | Own balance, own usage, own tickets | Return after login verification |
| **Sensitive** | Password, payment details, KYC docs, personal ID | Refuse disclosure, offer alternatives |
| **Restricted** | Other customer's data, admin data | Never disclose |

### 5.2 Identity Verification Flow

```
Customer requests sensitive data
  → Check if request is about OWN data
    → IF YES: Request identity verification
      → Ask verification questions:
          - Registered email
          - Last 4 digits of registered phone
          - Account number or customer ID
          - Last payment amount
        → IF verification passes: Provide data (or redirect to secure channel)
        → IF verification fails: "I couldn't verify your identity"
    → IF NO: Refuse immediately
```

---

## 6. Real-World Scenarios

### Scenario 1: Balance Inquiry

```
Platform: AeroXe Broadband
Customer: Ali (authenticated)

User: "What is my current balance?"

Agent thinks:
  1. Ali is asking about HIS OWN balance ← OK
  2. Balance is PRIVATE data ← needs identity verification
  3. Ali is already authenticated via JWT ← sufficient for own data
  4. Query billing_db for Ali's balance

Agent response:
  "Hi Ali, your current balance is ₹2,450.
   Your next billing date is February 1, 2026.
   Would you like to make a payment?"
```

### Scenario 2: Balance Inquiry with Cross-Platform Attempt

```
Platform: AeroXe Broadband
Customer: Ali (authenticated)

User: "What is my balance on AeroXe ERP?"

Agent thinks:
  1. Ali is on AeroXe Broadband platform
  2. Request is for AeroXe ERP ← different platform
  3. Agent has NO access to AeroXe ERP data
  4. Redirect to correct platform

Agent response:
  "I can only help you with AeroXe Broadband.
   For your AeroXe ERP balance, please log in to the AeroXe ERP platform
   or contact their support team."
```

### Scenario 3: Password Request

```
Platform: AeroXe Broadband
Customer: Ali (authenticated)

User: "What is my password?"

Agent thinks:
  1. Password is SENSITIVE data ← must NEVER disclose
  2. Even if it's Ali's own password
  3. Passwords are stored as hashes, never readable
  4. Refuse and offer alternatives

Agent response:
  "For security reasons, I cannot share your password.
   If you've forgotten your password, I can help you reset it.
   Would you like me to send a password reset link to your registered email?"
```

### Scenario 4: API Key / Token Request

```
Platform: AeroXe Broadband
Customer: Ali (authenticated)

User: "Show me my API key"

Agent thinks:
  1. API key is SENSITIVE data ← must NEVER disclose
  2. Even if it's Ali's own API key
  3. API keys are stored hashed
  4. Refuse and offer alternatives

Agent response:
  "For security reasons, I cannot display your API key.
   If you need a new API key, I can generate one for you.
   The new key will only be shown once. Would you like me to proceed?"
```

### Scenario 5: Personal ID / KYC Document

```
Platform: AeroXe Broadband
Customer: Ali (authenticated)

User: "Show me my Aadhaar card"

Agent thinks:
  1. KYC document is RESTRICTED data
  2. Even if it's Ali's own document
  3. Documents stored encrypted, not accessible via chat
  4. Refuse and redirect to secure portal

Agent response:
  "For security and compliance reasons, I cannot share KYC documents
   through this chat. You can view your documents in the
   Document Center on the web portal, or contact support for assistance."
```

### Scenario 6: Another Customer's Balance

```
Platform: AeroXe Broadband
Customer: Ali (authenticated)

User: "What is Priya's balance?"

Agent thinks:
  1. Ali is asking about Priya's data ← NOT his own
  2. Customer isolation violated
  3. Refuse immediately

Agent response:
  "I can only provide information about your own account.
   For other accounts, please contact our support team."
```

### Scenario 7: Unregistered User — General Inquiry (Allowed)

```
Platform: AeroXe Broadband
Customer: Raj (NOT registered on Broadband)

User: "What internet plans do you offer?"

Agent thinks:
  1. Raj is not registered on AeroXe Broadband
  2. This is a GENERAL PURPOSE inquiry (pricing/plans)
  3. No login required for public information
  4. Provide plan details

Agent response:
  "We offer the following internet plans:
   - Basic: 50 Mbps at ₹499/month
   - Standard: 100 Mbps at ₹799/month
   - Premium: 200 Mbps at ₹1,299/month
   Would you like to register for any of these plans?"
```

### Scenario 7b: Unregistered User — Personal Data Request (Blocked)

```
Platform: AeroXe Broadband
Customer: Raj (NOT registered on Broadband)

User: "What is my current balance?"

Agent thinks:
  1. Raj is not registered on AeroXe Broadband
  2. This is a PERSONAL DATA request (balance)
  3. Requires authentication
  4. Redirect to login/registration

Agent response:
  "You need to be logged in to view your account information.
   Please log in or register for an account to access your balance."
```

### Scenario 7c: Unregistered User — FAQ Query (Allowed)

```
Platform: AeroXe Broadband
Customer: Raj (NOT registered)

User: "How do I reset my router?"

Agent thinks:
  1. Raj is not registered
  2. This is a GENERAL PURPOSE inquiry (technical FAQ)
  3. No login required
  4. Provide troubleshooting steps

Agent response:
  "To reset your router:
   1. Locate the reset button on the back of your router
   2. Press and hold for 10 seconds until lights blink
   3. Wait 2-3 minutes for the router to restart
   4. Reconnect your devices

   If the issue persists, please register for support
   or contact our helpline at 1800-XXX-XXXX."
```

### Scenario 7d: Unregistered User — Pricing Comparison (Allowed)

```
Platform: AeroXe Broadband
Customer: Guest (not logged in via website chat)

User: "What's the difference between Basic and Premium plans?"

Agent thinks:
  1. Guest user
  2. GENERAL PURPOSE inquiry (pricing comparison)
  3. No login required
  4. Provide comparison

Agent response:
  "Here's the comparison:
   - Basic (₹499/month): 50 Mbps, 100 GB data, 1 device
   - Premium (₹1,299/month): 200 Mbps, Unlimited data, 5 devices

   Premium also includes:
   - Free router upgrade
   - Priority support
   - Static IP option

   Would you like to register for a plan?"
```

### Scenario 8: Sensitive Data with Identity Verification

```
Platform: AeroXe Broadband
Customer: Ali (authenticated)

User: "What was my last payment amount?"

Agent thinks:
  1. Payment info is PRIVATE data
  2. Ali is authenticated ← sufficient for own data
  3. Query billing_db for Ali's payment history

Agent response:
  "Your last payment was ₹1,200 on January 10, 2026.
   Payment method: UPI.
   Transaction ID: TXN-2026-001234.
   Would you like the full payment history?"
```

### Scenario 9: Password Reset Request

```
Platform: AeroXe Broadband
Customer: Ali (authenticated)

User: "I forgot my password, can you tell me?"

Agent thinks:
  1. Password is SENSITIVE ← never disclose
  2. Password reset is allowed
  3. Generate reset link

Agent response:
  "No problem! I've sent a password reset link to your
   registered email (a***@gmail.com).
   The link expires in 30 minutes.
   Please check your inbox (and spam folder)."
```

### Scenario 10: Multi-Platform User Balance Check

```
Platform: AeroXe Broadband
Customer: Ali (registered on Broadband + ERP + CRM)

User: "Show me all my balances across platforms"

Agent thinks:
  1. Ali is on AeroXe Broadband
  2. Agent only has access to Broadband data
  3. Cannot query ERP or CRM
  4. Provide Broadband balance, redirect for others

Agent response:
  "Here's your AeroXe Broadband balance: ₹2,450.

   For your other platforms:
   - AeroXe ERP: Please log in to the ERP portal
   - AeroXe CRM: Please log in to the CRM portal

   Each platform has its own support team to help you."
```

---

## 7. SQL Agent Customer Isolation

### 7.1 Rule: SQL Queries Must Filter by Current Customer

Every SQL query generated by the agent MUST include:

```sql
WHERE customer_id = {current_customer_id}
  AND tenant_id = {current_tenant_id}
```

### 7.2 Example: SQL Generation with Customer Isolation

```
User: "Show me my invoices"

Agent generates:
  SELECT id, amount, status, invoice_date
  FROM invoices
  WHERE customer_id = 201          ← Current customer only
    AND tenant_id = 1              ← Current platform only
  ORDER BY invoice_date DESC;

NOT:
  SELECT id, amount, status, invoice_date
  FROM invoices;                   ← MISSING customer filter, BLOCKED
```

### 7.3 Cross-Customer Attempt via SQL

```
User: "Show me all customer invoices"

Agent generates:
  SELECT id, amount, status, invoice_date
  FROM invoices
  WHERE customer_id = 201          ← Forced to current customer only
    AND tenant_id = 1
  ORDER BY invoice_date DESC;

Agent explains:
  "I've shown your invoices. For security reasons,
   I can only access your own account data."
```

---

## 8. Memory Service Customer Isolation

### 8.1 Rule: Agent Memory is Scoped to Customer

| Memory Type | Scope |
|---|---|
| Short-term (Redis) | customer_id + session_id |
| Long-term (pgvector) | customer_id + tenant_id |
| Organizational (AGE) | tenant_id only |

### 8.2 Example: Memory Isolation

```
Customer Ali asks: "What did I discuss last time?"

Agent searches memory:
  SELECT * FROM memories
  WHERE customer_id = 201           ← Ali's memories only
    AND tenant_id = 1
  ORDER BY created_at DESC
  LIMIT 5;

Agent response:
  "Last time you asked about your billing cycle.
   I explained that your plan renews on the 1st of each month."

NOT:
  "Last time Priya asked about..." ← WRONG, different customer
```

---

## 9. Agent Response Guard Rules

### 9.1 Pre-Response Validation

Every agent response passes through validation before delivery:

| Check | Action |
|---|---|
| Response contains other customer's data | BLOCK, log violation |
| Response contains password/secret | BLOCK, log violation |
| Response contains PII of another customer | BLOCK, log violation |
| Response contains platform data from wrong platform | BLOCK, redirect |
| Response contains internal system details | BLOCK, generic error |
| Response contains API keys/tokens | BLOCK, log violation |
| Prompt contains banned/sensitive words | BLOCK, safe message |
| Response contains profanity/hate speech | BLOCK, safe message |

### 9.2 Response Validation Flow

```
Agent generates response
  → Response Validator
    → Sensitive Words Filter (profanity, hate, violence, etc.)
    → Extract entities (customer names, IDs, accounts)
    → Check: Do any match OTHER customers?
    → Check: Does response contain sensitive patterns?
      - password, secret, token, key, PIN
      - Credit card numbers
      - SSN/Aadhaar patterns
      - API keys, JWT tokens
        → IF violation detected: BLOCK + audit log
        → IF clean: Deliver to customer
```

---

## 10. Audit Trail

### 10.1 Every Isolation Event is Logged

| Event | Logged |
|---|---|
| Customer data access | customer_id, data_type, result |
| Cross-platform attempt | from_platform, to_platform, blocked |
| Cross-customer attempt | requester_id, target_id, blocked |
| Sensitive data request | customer_id, data_type, action |
| Identity verification | customer_id, method, result |
| Response validation | customer_id, blocked_reason |

### 10.2 Example Audit Event

```json
{
  "event_type": "security.customer.isolation.violation",
  "customer_id": 201,
  "platform_id": 1,
  "action": "cross_customer_data_access",
  "details": {
    "requested_customer": 301,
    "data_type": "balance",
    "result": "blocked"
  },
  "timestamp": "2026-01-15T10:30:00Z"
}
```

---

## 11. API Endpoints

### Customer Data API

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/customers/me/balance` | Get own balance |
| GET | `/api/v1/customers/me/usage` | Get own usage |
| GET | `/api/v1/customers/me/invoices` | Get own invoices |
| GET | `/api/v1/customers/me/tickets` | Get own support tickets |
| POST | `/api/v1/customers/me/verify` | Verify identity for sensitive data |

### Identity Verification API

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/identity/verify` | Verify customer identity |
| POST | `/api/v1/identity/verify/mfa` | MFA verification for restricted data |

### Verification Questions

| Question | Expected Answer | Source |
|---|---|---|
| Registered email | Partial match (a***@gmail.com) | customers.email |
| Last 4 digits of phone | Exact match | customers.phone |
| Account number | Exact match | customers.account_number |
| Last payment amount | Approximate match (±10%) | payments.amount |

---

## 12. Grafana Dashboard

| Panel | Description |
|---|---|
| Customer Isolation Violations | Cross-customer access attempts |
| Cross-Platform Redirects | Users redirected to correct platform |
| Sensitive Data Refusals | Password/secret/key requests blocked |
| Identity Verification Success Rate | Verification pass/fail rates |
| Response Validation Blocks | Pre-response blocks by reason |

---

## 13. Security Checklist

| Check | Status |
|---|---|
| Guest access for general inquiries allowed | Public info (FAQ, pricing, plans) available without login |
| Personal data requires authentication | Balance, usage, invoices blocked for unregistered users |
| Agent only accesses current customer's data | Enforced at query level |
| Agent only operates within bound platform | Enforced at agent scope |
| Passwords never disclosed | Hard rule, no exceptions |
| API keys never disclosed | Hard rule, no exceptions |
| KYC docs never via chat | Redirect to secure portal |
| Cross-customer data blocked | Response validator + audit |
| Cross-platform data blocked | Platform scope enforcement |
| Identity verification for private data | Authentication sufficient |
| Identity verification for sensitive data | Additional verification required |
| All violations logged | Audit trail with alerts |
