# AeroXe Nexus AI — Outbound Module

## Proactive AI Calls, Campaign Management, DNC Compliance & Scheduling

> **Modular Monolith Module:** This document describes the `nexus-outbound` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces and NATS.

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-outbound` |
| Crate | `nexus-outbound` (workspace member) |
| Bounded Context | Outbound Communication |
| Domain Type | Supporting Domain |
| Language | Rust (edition 2024) |
| Schema | `outbound_` (in shared PostgreSQL) |
| Dependencies | `nexus-telephony` (calls), `nexus-conversation` (state), `nexus-agent` (AI), `nexus-notification` (messages), `nexus-audit` (logging) |

---

## 2. Purpose

The Outbound module enables AI-initiated communications:

- Outbound voice calls (AI calling customers)
- Proactive chat messages (AI initiating conversations)
- Campaign management (bulk outbound)
- Callback scheduling
- Do-Not-Call (DNC) compliance
- Call scheduling and timezone management
- Outbound rate limiting and throttling

---

## 3. Aggregate Design

### Campaign Aggregate

```
Campaign (Aggregate Root)
├── CampaignMetadata
│   ├── CampaignId
│   ├── TenantId
│   ├── Name
│   ├── Description
│   ├── Type (Voice | Chat | Email | SMS)
│   ├── Status (Draft | Active | Paused | Completed | Cancelled)
│   └── Schedule
├── TargetAudience
│   ├── CustomerFilter (segment definition)
│   ├── CustomerIds[] (explicit list)
│   └── totalCount
├── CampaignContent
│   ├── Script / Message template
│   ├── AgentId
│   ├── Variables[]
│   └── FallbackMessage
├── Execution
│   ├── TotalTargets
│   ├── Completed
│   ├── Successful
│   ├── Failed
│   ├── Skipped (DNC)
│   └── Pending
└── Results
    ├── ContactRate
    ├── SuccessRate
    ├── ConversionRate
    └── Cost
```

---

## 4. Public API Trait

```rust
// nexus-outbound/src/interfaces/api.rs
#[async_trait]
pub trait OutboundService: Send + Sync {
    // Campaign management
    async fn create_campaign(&self, req: CreateCampaignRequest) -> Result<Campaign, OutboundError>;
    async fn start_campaign(&self, id: CampaignId) -> Result<(), OutboundError>;
    async fn pause_campaign(&self, id: CampaignId) -> Result<(), OutboundError>;
    async fn cancel_campaign(&self, id: CampaignId) -> Result<(), OutboundError>;
    async fn get_campaign(&self, id: CampaignId) -> Result<Campaign, OutboundError>;
    async fn list_campaigns(&self, tenant_id: TenantId) -> Result<Vec<Campaign>, OutboundError>;
    async fn get_campaign_stats(&self, id: CampaignId) -> Result<CampaignStats, OutboundError>;
    
    // One-off outbound
    async fn make_outbound_call(&self, req: OutboundCallRequest) -> Result<OutboundCallResult, OutboundError>;
    async fn send_proactive_message(&self, req: ProactiveMessageRequest) -> Result<(), OutboundError>;
    
    // Scheduling
    async fn schedule_callback(&self, req: ScheduleCallbackRequest) -> Result<ScheduledCallback, OutboundError>;
    async fn get_callbacks(&self, tenant_id: TenantId, date: NaiveDate) -> Result<Vec<ScheduledCallback>, OutboundError>;
    async fn cancel_callback(&self, id: CallbackId) -> Result<(), OutboundError>;
    
    // DNC management
    async fn add_to_dnc(&self, req: DNCRequest) -> Result<(), OutboundError>;
    async fn remove_from_dnc(&self, req: DNCRequest) -> Result<(), OutboundError>;
    async fn check_dnc(&self, phone: PhoneNumber, tenant_id: TenantId) -> Result<bool, OutboundError>;
    async fn get_dnc_list(&self, tenant_id: TenantId) -> Result<Vec<DNCEntry>, OutboundError>;
}

pub struct CreateCampaignRequest {
    pub tenant_id: TenantId,
    pub name: String,
    pub description: String,
    pub campaign_type: CampaignType,
    pub agent_id: AgentId,
    pub script: String,
    pub target_filter: Option<CustomerFilter>,
    pub target_ids: Option<Vec<CustomerId>>,
    pub schedule: CampaignSchedule,
    pub rate_limit: u32,           // Max concurrent calls
    pub user_id: UserId,
}

pub enum CampaignType {
    Voice,
    Chat,
    Email,
    SMS,
}

pub struct OutboundCallRequest {
    pub tenant_id: TenantId,
    pub callee_number: PhoneNumber,
    pub agent_id: AgentId,
    pub context: HashMap<String, String>,
    pub schedule: Option<DateTime>,   // Immediate if None
    pub user_id: UserId,
}

pub struct ProactiveMessageRequest {
    pub tenant_id: TenantId,
    pub customer_id: CustomerId,
    pub channel: String,              // chat | email | sms | whatsapp
    pub agent_id: AgentId,
    pub message: String,
    pub context: HashMap<String, String>,
    pub user_id: UserId,
}

pub struct ScheduleCallbackRequest {
    pub tenant_id: TenantId,
    pub customer_id: CustomerId,
    pub phone_number: PhoneNumber,
    pub agent_id: Option<AgentId>,
    pub scheduled_at: DateTime,
    pub reason: String,
    pub context: HashMap<String, String>,
    pub user_id: UserId,
}

pub struct DNCRequest {
    pub tenant_id: TenantId,
    pub phone_number: PhoneNumber,
    pub reason: String,
    pub user_id: UserId,
}

pub struct CampaignStats {
    pub campaign_id: CampaignId,
    pub total_targets: u64,
    pub attempted: u64,
    pub connected: u64,
    pub successful: u64,
    pub failed: u64,
    pub skipped_dnc: u64,
    pub avg_duration_seconds: f64,
    pub conversion_rate: f64,
    pub total_cost: f64,
}
```

---

## 5. Campaign Execution Flow

```
Campaign Created (Draft)
    |
    v
[1] Target Selection
    |  - Apply customer filter
    |  - Load customer list
    |  - Check DNC list → Remove DNC numbers
    |  - Validate phone numbers
    |
    v
[2] Schedule
    |  - Determine execution time
    |  - Apply timezone rules
    |  - Respect business hours
    |  - Apply rate limits
    |
    v
[3] Execution Loop
    |  - Dequeue target
    |  - Check DNC again (real-time)
    |  - Initiate outbound call (nexus-telephony)
    |  - Assign AI agent (nexus-agent)
    |  - Start conversation (nexus-conversation)
    |
    v
[4] Conversation
    |  - AI agent executes script
    |  - Handles responses
    |  - Collects information
    |  - Records outcome
    |
    v
[5] Post-Call
    |  - Record result
    |  - Update campaign stats
    |  - Log to audit
    |  - Send notification if needed
    |
    v
[6] Repeat until all targets processed
```

---

## 6. DNC (Do-Not-Call) Compliance

### 6.1 DNC Rules

| Rule | Description |
|---|---|
| Universal DNC | Check national DNC registry |
| Tenant DNC | Tenant-specific block list |
| Customer DNC | Customer-level opt-out |
| Time-based DNC | No calls outside business hours |
| Frequency limit | Max N calls per customer per period |
| Opt-out processing | Honor opt-out requests |

### 6.2 Business Hours

```rust
pub struct BusinessHours {
    pub timezone: String,
    pub monday: Option<TimeRange>,
    pub tuesday: Option<TimeRange>,
    pub wednesday: Option<TimeRange>,
    pub thursday: Option<TimeRange>,
    pub friday: Option<TimeRange>,
    pub saturday: Option<TimeRange>,
    pub sunday: Option<TimeRange>,
    pub holidays: Vec<NaiveDate>,
}

pub struct TimeRange {
    pub start: NaiveTime,
    pub end: NaiveTime,
}
```

---

## 7. Database Schema (outbound_ schema)

### campaigns

```sql
CREATE TABLE outbound.campaigns (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    campaign_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    campaign_type VARCHAR(20) NOT NULL,
    agent_id BIGINT NOT NULL,
    script TEXT NOT NULL,
    target_filter JSONB,
    target_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    rate_limit INT NOT NULL DEFAULT 10,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### campaign_targets

```sql
CREATE TABLE outbound.targets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    campaign_id BIGINT NOT NULL REFERENCES outbound.campaigns(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT,
    phone_number VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempt_count INT NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP,
    call_id BIGINT,
    outcome VARCHAR(50),
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbound_targets_campaign ON outbound.targets(campaign_id, status);
CREATE INDEX idx_outbound_targets_pending ON outbound.targets(status, created_at) WHERE status = 'pending';
```

### scheduled_callbacks

```sql
CREATE TABLE outbound.callbacks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    callback_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    agent_id BIGINT,
    scheduled_at TIMESTAMP NOT NULL,
    reason TEXT NOT NULL,
    context JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    call_id BIGINT,
    completed_at TIMESTAMP,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_callbacks_tenant_date ON outbound.callbacks(tenant_id, scheduled_at);
CREATE INDEX idx_callbacks_pending ON outbound.callbacks(status, scheduled_at) WHERE status = 'scheduled';
```

### dnc_list

```sql
CREATE TABLE outbound.dnc_list (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    source VARCHAR(50) NOT NULL,          -- customer_request | regulatory | admin
    reason TEXT,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    removed_at TIMESTAMP,
    added_by BIGINT
);

CREATE UNIQUE INDEX idx_dnc_unique ON outbound.dnc_list(tenant_id, phone_number) WHERE removed_at IS NULL;
```

---

## 8. REST API Endpoints

### Create Campaign

```
POST /api/v1/outbound/campaigns
Authorization: Bearer <jwt>
```

### Start Campaign

```
POST /api/v1/outbound/campaigns/{id}/start
```

### Get Campaign Stats

```
GET /api/v1/outbound/campaigns/{id}/stats
```

### Schedule Callback

```
POST /api/v1/outbound/callbacks
```

### Add to DNC

```
POST /api/v1/outbound/dnc
```

---

## 9. NATS Events

| Subject | Event |
|---|---|
| `aeroxe.v1.outbound.campaign.started` | `CampaignStarted` |
| `aeroxe.v1.outbound.campaign.completed` | `CampaignCompleted` |
| `aeroxe.v1.outbound.call.initiated` | `OutboundCallInitiated` |
| `aeroxe.v1.outbound.call.completed` | `OutboundCallCompleted` |
| `aeroxe.v1.outbound.callback.scheduled` | `CallbackScheduled` |
| `aeroxe.v1.outbound.callback.triggered` | `CallbackTriggered` |

---

## 10. Observability

| Metric | Description |
|---|---|
| `outbound_campaigns_total` | Total campaigns |
| `outbound_calls_initiated` | Calls initiated |
| `outbound_calls_connected` | Calls connected |
| `outbound_callbacks_scheduled` | Callbacks scheduled |
| `outbound_callbacks_completed` | Callbacks completed |
| `outbound_dnc_hits` | DNC blocks |
| `outbound_success_rate` | Campaign success rate |
