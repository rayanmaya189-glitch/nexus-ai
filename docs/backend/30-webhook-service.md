# AeroXe Nexus AI — Webhook Module

## Outbound Webhooks, Event Delivery, Webhook Management & Retry Logic

> **Modular Monolith Module:** This document describes the `nexus-webhook` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via NATS event consumption.

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-webhook` |
| Crate | `nexus-webhook` (workspace member) |
| Bounded Context | Integration |
| Domain Type | Supporting Domain |
| Language | Rust (edition 2024) |
| Schema | `webhook_` (in shared PostgreSQL) |
| Dependencies | NATS (event consumption), Redis (delivery queue), PostgreSQL (config) |

---

## 2. Purpose

The Webhook module enables external systems to subscribe to AeroXe events:

- Tenant-configurable outbound webhooks
- Event subscription management
- Guaranteed delivery with retry
- Webhook signature verification (HMAC)
- Delivery logging and debugging
- Rate limiting per webhook
- Event filtering (subscribe to specific event types)

---

## 3. Aggregate Design

### WebhookSubscription Aggregate

```
WebhookSubscription (Aggregate Root)
├── SubscriptionMetadata
│   ├── SubscriptionId
│   ├── TenantId
│   ├── Name
│   ├── Url
│   ├── Status (Active | Paused | Failed)
│   └── CreatedBy
├── EventFilter
│   ├── EventTypes[]     // e.g., ["conversation.ended", "call.completed"]
│   ├── AgentFilter[]    // Only events from specific agents
│   └── ChannelFilter[]  // Only events from specific channels
├── DeliveryConfig
│   ├── Secret (HMAC signing)
│   ├── RetryPolicy
│   │   ├── MaxRetries (default: 5)
│   │   ├── InitialDelay (1s)
│   │   ├── BackoffMultiplier (2x)
│   │   └── MaxDelay (5min)
│   ├── Timeout (30s)
│   └── RateLimit (100/min)
└── DeliveryLog[]
    ├── DeliveryId
    ├── EventId
    ├── Status (Delivered | Failed | Pending)
    ├── ResponseCode
    ├── ResponseBody
    ├── LatencyMs
    └── AttemptCount
```

---

## 4. Public API Trait

```rust
// nexus-webhook/src/interfaces/api.rs
#[async_trait]
pub trait WebhookService: Send + Sync {
    // Subscription management
    async fn create_subscription(&self, req: CreateWebhookRequest) -> Result<WebhookSubscription, WebhookError>;
    async fn get_subscription(&self, id: SubscriptionId) -> Result<Option<WebhookSubscription>, WebhookError>;
    async fn update_subscription(&self, req: UpdateWebhookRequest) -> Result<WebhookSubscription, WebhookError>;
    async fn delete_subscription(&self, id: SubscriptionId) -> Result<(), WebhookError>;
    async fn list_subscriptions(&self, tenant_id: TenantId) -> Result<Vec<WebhookSubscription>, WebhookError>;
    
    // Testing
    async fn test_webhook(&self, id: SubscriptionId) -> Result<WebhookTestResult, WebhookError>;
    
    // Delivery logs
    async fn get_delivery_logs(&self, id: SubscriptionId, limit: u32) -> Result<Vec<DeliveryLog>, WebhookError>;
    
    // Retry management
    async fn retry_delivery(&self, delivery_id: DeliveryId) -> Result<(), WebhookError>;
}

pub struct CreateWebhookRequest {
    pub tenant_id: TenantId,
    pub name: String,
    pub url: String,
    pub secret: String,
    pub event_types: Vec<String>,
    pub agent_filter: Option<Vec<AgentId>>,
    pub channel_filter: Option<Vec<String>>,
    pub user_id: UserId,
}

pub struct WebhookSubscription {
    pub subscription_id: SubscriptionId,
    pub tenant_id: TenantId,
    pub name: String,
    pub url: String,
    pub status: WebhookStatus,
    pub event_types: Vec<String>,
    pub secret: String,
    pub retry_policy: RetryPolicy,
    pub created_at: DateTime,
    pub updated_at: DateTime,
}

pub struct WebhookTestResult {
    pub success: bool,
    pub status_code: Option<u16>,
    pub response_body: Option<String>,
    pub latency_ms: f64,
    pub error: Option<String>,
}
```

---

## 5. Webhook Payload

### Standard Event Payload

```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "conversation.ended",
  "api_version": "v1",
  "timestamp": "2026-07-21T12:00:00Z",
  "tenant_id": 1,
  "data": {
    "conversation_id": "conv-uuid",
    "channel": "voice",
    "agent_id": "support-agent",
    "customer_id": 123,
    "duration_seconds": 342,
    "turn_count": 12,
    "outcome": "resolved",
    "satisfaction_score": 5,
    "summary": "Resolved billing inquiry"
  },
  "metadata": {
    "source": "aeroxe-nexus",
    "version": "1.0.0"
  }
}
```

### HMAC Signature

```
X-AeroXe-Signature: sha256=<hex-digest>
X-AeroXe-Timestamp: 1721548800
```

Verification:
```rust
let payload = request.body;
let timestamp = request.headers["X-AeroXe-Timestamp"];
let signature = request.headers["X-AeroXe-Signature"];

let expected = hmac_sha256(secret, format!("{}.{}", timestamp, payload));
assert_eq!(signature, format!("sha256={}", hex(expected)));
```

---

## 6. Delivery Pipeline

```
NATS Event Received
    |
    v
[1] Event Filter
    |  - Match event_type to subscriptions
    |  - Apply agent/channel filters
    |  - Skip if no matching subscriptions
    |
    v
[2] Rate Limit Check
    |  - Per-subscription rate limit (Redis)
    |  - Queue if over limit
    |
    v
[3] Payload Construction
    |  - Build JSON payload
    |  - Generate HMAC signature
    |  - Add timestamp
    |
    v
[4] HTTP POST
    |  - POST to webhook URL
    |  - Timeout: 30s
    |  - Content-Type: application/json
    |  - Headers: X-AeroXe-Signature, X-AeroXe-Timestamp
    |
    v
[5] Response Handling
    |  - 2xx: Success, log delivery
    |  - 4xx: Client error, don't retry
    |  - 5xx: Server error, retry with backoff
    |  - Timeout: Retry with backoff
    |
    v
[6] Retry Logic
    |  - Exponential backoff
    |  - Max 5 retries
    |  - Max delay: 5 minutes
    |  - After max retries: Mark as failed, alert tenant
```

---

## 7. Database Schema (webhook_ schema)

### webhook_subscriptions

```sql
CREATE TABLE webhook.subscriptions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    subscription_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    url TEXT NOT NULL,
    secret_hash VARCHAR(100) NOT NULL,
    event_types JSONB NOT NULL,
    agent_filter JSONB,
    channel_filter JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    retry_max INT NOT NULL DEFAULT 5,
    retry_backoff FLOAT NOT NULL DEFAULT 2.0,
    rate_limit INT NOT NULL DEFAULT 100,
    timeout_ms INT NOT NULL DEFAULT 30000,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_tenant ON webhook.subscriptions(tenant_id);
CREATE INDEX idx_webhook_status ON webhook.subscriptions(status) WHERE status = 'active';
```

### webhook_deliveries

```sql
CREATE TABLE webhook.deliveries (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES webhook.subscriptions(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    event_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempt_count INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    response_code INT,
    response_body TEXT,
    latency_ms FLOAT,
    next_retry_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP
);

CREATE INDEX idx_webhook_deliveries_sub ON webhook.deliveries(subscription_id, created_at DESC);
CREATE INDEX idx_webhook_deliveries_pending ON webhook.deliveries(status, next_retry_at) WHERE status = 'pending';
```

---

## 8. REST API Endpoints

| Method | Endpoint | Business Status | HTTP | Description |
|---|---|---|---|---|
| `POST` | `/api/v1/webhooks` | `CREATED` | `201` | Create webhook |
| `GET` | `/api/v1/webhooks/{id}` | `SUCCESS` | `200` | Get webhook |
| `GET` | `/api/v1/webhooks?limit=10&offset=0` | `SUCCESS` | `200` | List webhooks |
| `PATCH` | `/api/v1/webhooks/{id}` | `UPDATED` | `200` | Update webhook |
| `DELETE` | `/api/v1/webhooks/{id}` | `DELETED` | `204` | Delete webhook |
| `POST` | `/api/v1/webhooks/{id}/test` | `SUCCESS` | `200` | Test webhook |
| `GET` | `/api/v1/webhooks/{id}/deliveries?limit=10&offset=0` | `SUCCESS` | `200` | List deliveries |
| `POST` | `/api/v1/webhooks/{id}/deliveries/{delivery_id}/retry` | `SUCCESS` | `200` | Retry delivery |

### List Webhooks Response

```json
{
  "status": "SUCCESS",
  "data": [...],
  "summary": {
    "total_items": 25,
    "active_webhooks": 20,
    "paused_webhooks": 3,
    "failed_webhooks": 2,
    "total_deliveries_today": 1500,
    "delivery_success_rate": 0.96
  },
  "pagination": {"total": 25, "limit": 10, "offset": 0, "has_more": false},
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

### List Deliveries Response

```json
{
  "status": "SUCCESS",
  "data": [...],
  "summary": {
    "total_items": 5000,
    "delivered": 4800,
    "failed": 150,
    "pending": 50,
    "avg_latency_ms": 245,
    "recent_activity": {
      "delivered_today": 200,
      "failed_today": 5
    }
  },
  "pagination": {"total": 5000, "limit": 10, "offset": 0, "has_more": true},
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

**Note:** No PUT method. Use PATCH for updates. All list endpoints support `limit` (default 10) and `offset`.

---

## 9. Supported Event Types

| Event Type | Description |
|---|---|
| `conversation.created` | New conversation started |
| `conversation.ended` | Conversation completed |
| `conversation.escalated` | Conversation escalated to human |
| `call.inbound` | Inbound call received |
| `call.outbound` | Outbound call initiated |
| `call.completed` | Call ended |
| `customer.created` | New customer |
| `customer.suspended` | Customer suspended |
| `agent.completed` | Agent execution done |
| `security.threat.detected` | Security event |
| `workflow.completed` | Workflow finished |
| `kyc.approved` | KYC verified |
| `kyc.rejected` | KYC rejected |

---

## 10. Observability

| Metric | Description |
|---|---|
| `webhook_subscriptions_active` | Active subscriptions |
| `webhook_deliveries_total` | Total deliveries |
| `webhook_deliveries_success` | Successful deliveries |
| `webhook_deliveries_failed` | Failed deliveries |
| `webhook_delivery_latency_ms` | Delivery latency |
| `webhook_retry_total` | Retry attempts |
| `webhook_rate_limit_hits` | Rate limit rejections |

---

## 11. Webhook Security (NEW - CRITICAL)

### 11.1 IP Whitelisting

```rust
pub struct WebhookSecurityConfig {
    pub subscription_id: SubscriptionId,
    pub allowed_ips: Vec<String>,        // CIDR ranges allowed
    pub allowed_ip_ranges: Vec<CIDRRange>,
    pub require_tls: bool,               // Reject non-HTTPS
    pub min_tls_version: String,         // "1.2" or "1.3"
    pub timeout_ms: u32,                 // Max response time
    pub max_payload_bytes: u64,          // Max body size
}

pub struct CIDRRange {
    pub start: String,                   // e.g., "192.168.1.0"
    pub prefix_len: u8,                  // e.g., 24
}
```

### 11.2 Source Validation Flow

```
Webhook Request Received
    |
    v
[1] IP Whitelist Check
    |  - Verify source IP is in allowed list
    |  - Reject if not whitelisted
    |
    v
[2] TLS Check
    |  - Verify HTTPS connection
    |  - Check TLS version (>= 1.2)
    |
    v
[3] HMAC Signature Verification
    |  - Verify X-AeroXe-Signature header
    |  - Validate timestamp (prevent replay)
    |
    v
[4] Rate Limit Check
    |  - Per-source-IP rate limit
    |  - Per-subscription rate limit
    |
    v
[5] Payload Validation
    |  - Content-Type check
    |  - Size limit check
    |  - Schema validation
    |
    v
[6] Process Request
```

### 11.3 Security Entities

```sql
ALTER TABLE webhook.subscriptions ADD COLUMN allowed_ips JSONB;
ALTER TABLE webhook.subscriptions ADD COLUMN require_tls BOOLEAN DEFAULT true;
ALTER TABLE webhook.subscriptions ADD COLUMN min_tls_version VARCHAR(5) DEFAULT '1.2';
ALTER TABLE webhook.subscriptions ADD COLUMN max_payload_bytes BIGINT DEFAULT 1048576;
ALTER TABLE webhook.subscriptions ADD COLUMN timeout_ms INT DEFAULT 30000;
```

### 11.4 Webhook Inbound Rate Limiting

| Scope | Limit | Window |
|---|---|---|
| Per source IP | 1000 requests | 1 minute |
| Per subscription | 100 requests | 1 minute |
| Per tenant (total) | 10000 requests | 1 minute |
