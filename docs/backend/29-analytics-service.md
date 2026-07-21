# AeroXe Nexus AI — Analytics Module

## Conversation Analytics, AI Performance Metrics, Business Intelligence & Dashboards

> **Modular Monolith Module:** This document describes the `nexus-analytics` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces and consumes NATS events.

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-analytics` |
| Crate | `nexus-analytics` (workspace member) |
| Bounded Context | Analytics & Intelligence |
| Domain Type | Supporting Domain |
| Language | Rust (edition 2024) |
| Schema | `analytics_` (in shared PostgreSQL) |
| Dependencies | NATS (event consumption), PostgreSQL (storage), Elasticsearch (search) |

---

## 2. Purpose

The Analytics module provides business intelligence and operational insights:

- Real-time conversation metrics
- AI agent performance tracking
- Customer satisfaction analytics
- Voice call analytics (call center metrics)
- Cost and usage tracking
- Anomaly detection
- Custom report generation
- Dashboard data APIs

---

## 3. Aggregate Design

### AnalyticsSnapshot Aggregate

```
AnalyticsSnapshot (Aggregate Root)
├── TimeRange
│   ├── Start
│   ├── End
│   └── Granularity (Minute | Hour | Day | Week | Month)
├── Metrics
│   ├── ConversationMetrics
│   ├── VoiceCallMetrics
│   ├── AgentMetrics
│   ├── CustomerMetrics
│   └── CostMetrics
├── Dimensions
│   ├── TenantId
│   ├── AgentId
│   ├── Channel
│   └── TimeBucket
└── Trends
    ├── Comparison (vs previous period)
    └── ChangePercent
```

---

## 4. Public API Trait

```rust
// nexus-analytics/src/interfaces/api.rs
#[async_trait]
pub trait AnalyticsService: Send + Sync {
    // Real-time metrics
    async fn get_dashboard(&self, tenant_id: TenantId, time_range: TimeRange) -> Result<Dashboard, AnalyticsError>;
    async fn get_realtime_metrics(&self, tenant_id: TenantId) -> Result<RealtimeMetrics, AnalyticsError>;
    
    // Conversation analytics
    async fn get_conversation_metrics(&self, req: ConversationMetricsRequest) -> Result<ConversationMetrics, AnalyticsError>;
    async fn get_csat_distribution(&self, tenant_id: TenantId, time_range: TimeRange) -> Result<CSATDistribution, AnalyticsError>;
    async fn get_resolution_rate(&self, tenant_id: TenantId, time_range: TimeRange) -> Result<RateMetric, AnalyticsError>;
    
    // Voice call analytics
    async fn get_call_metrics(&self, req: CallMetricsRequest) -> Result<CallMetrics, AnalyticsError>;
    async fn get_queue_metrics(&self, tenant_id: TenantId) -> Result<QueueMetrics, AnalyticsError>;
    async fn get_agent_utilization(&self, tenant_id: TenantId, time_range: TimeRange) -> Result<Vec<AgentUtilization>, AnalyticsError>;
    
    // Agent performance
    async fn get_agent_performance(&self, agent_id: AgentId, time_range: TimeRange) -> Result<AgentPerformance, AnalyticsError>;
    async fn get_agent_leaderboard(&self, tenant_id: TenantId, time_range: TimeRange, metric: LeaderboardMetric) -> Result<Vec<AgentScore>, AnalyticsError>;
    
    // Cost analytics
    async fn get_cost_breakdown(&self, tenant_id: TenantId, time_range: TimeRange) -> Result<CostBreakdown, AnalyticsError>;
    async fn get_token_usage(&self, tenant_id: TenantId, time_range: TimeRange) -> Result<TokenUsageReport, AnalyticsError>;
    
    // Custom reports
    async fn generate_report(&self, req: ReportRequest) -> Result<Report, AnalyticsError>;
    async fn list_reports(&self, tenant_id: TenantId) -> Result<Vec<ReportSummary>, AnalyticsError>;
}

pub struct TimeRange {
    pub start: DateTime,
    pub end: DateTime,
    pub granularity: Granularity,
}

pub enum Granularity {
    Minute,
    Hour,
    Day,
    Week,
    Month,
}

pub struct Dashboard {
    pub realtime: RealtimeMetrics,
    pub conversations: ConversationMetrics,
    pub calls: Option<CallMetrics>,
    pub agents: AgentSummary,
    pub costs: CostSummary,
    pub trends: Vec<Trend>,
}

pub struct RealtimeMetrics {
    pub active_conversations: u32,
    pub active_calls: u32,
    pub queue_size: u32,
    pub avg_response_time_ms: f64,
    pub messages_per_minute: f64,
    pub error_rate: f64,
}

pub struct ConversationMetrics {
    pub total_conversations: u64,
    pub avg_duration_seconds: f64,
    pub avg_turns: f64,
    pub resolution_rate: f64,
    pub escalation_rate: f64,
    pub abandonment_rate: f64,
    pub avg_first_response_ms: f64,
    pub avg_csat_score: f64,
    pub channel_breakdown: HashMap<String, u64>,
}

pub struct CallMetrics {
    pub total_calls: u64,
    pub answered_calls: u64,
    pub abandoned_calls: u64,
    pub avg_handle_time_seconds: f64,
    pub avg_wait_time_seconds: f64,
    pub avg_talk_time_seconds: f64,
    pub service_level: f64,            // % answered within threshold
    pub first_call_resolution: f64,
    pub transfer_rate: f64,
    pub recording_rate: f64,
}

pub struct AgentPerformance {
    pub agent_id: AgentId,
    pub conversations_handled: u64,
    pub avg_response_time_ms: f64,
    pub avg_csat_score: f64,
    pub resolution_rate: f64,
    pub tokens_used: u64,
    pub cost: f64,
    pub uptime_percentage: f64,
}

pub struct CostBreakdown {
    pub total_cost: f64,
    pub llm_cost: f64,
    pub stt_cost: f64,
    pub tts_cost: f64,
    pub telephony_cost: f64,
    pub storage_cost: f64,
    pub cost_per_conversation: f64,
    pub cost_per_message: f64,
}

pub struct CSATDistribution {
    pub scores: Vec<CSATBucket>,
    pub average: f64,
    pub total_responses: u64,
}

pub struct CSATBucket {
    pub score: u32,
    pub count: u64,
    pub percentage: f64,
}
```

---

## 5. Metrics Collection

### 5.1 Event-Driven Collection

The analytics module consumes NATS events from all other modules:

| Source Event | Analytics Update |
|---|---|
| `conversation.created` | Increment conversation count |
| `conversation.ended` | Calculate duration, outcome |
| `conversation.state.changed` | Track state transitions |
| `message.added` | Track message volume, response time |
| `telephony.call.answered` | Update call metrics |
| `telephony.call.ended` | Calculate call duration, outcome |
| `telephony.call.transferred` | Track transfer rate |
| `agent.execution.completed` | Track agent performance |
| `agent.tool.executed` | Track tool usage |
| `security.threat.detected` | Track security events |
| `workflow.completed` | Track workflow metrics |

### 5.2 Aggregation Pipeline

```
Raw Events (NATS)
    |
    v
[1] Event Ingestion
    |  - Parse event
    |  - Validate schema
    |
    v
[2] Real-time Update
    |  - Update Redis counters
    |  - Update active metrics
    |
    v
[3] Time-bucket Aggregation
    |  - Minute buckets → Hour buckets → Day buckets
    |  - Pre-compute common queries
    |
    v
[4] Persistence
    |  - Write to PostgreSQL (analytics schema)
    |  - Index to Elasticsearch (for search)
    |
    v
[5] Alert Evaluation
    |  - Check anomaly thresholds
    |  - Trigger alerts if needed
```

---

## 6. Database Schema (analytics_ schema)

### conversation_metrics (partitioned)

```sql
CREATE TABLE analytics.conversation_metrics (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id BIGINT NOT NULL,
    conversation_id BIGINT,
    agent_id BIGINT,
    channel VARCHAR(20) NOT NULL,
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    duration_seconds INT,
    turn_count INT,
    tokens_input INT,
    tokens_output INT,
    model VARCHAR(100),
    outcome VARCHAR(30),
    satisfaction_score INT,
    first_response_ms INT,
    avg_response_ms INT,
    cost DECIMAL(10,4),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
```

### call_metrics (partitioned)

```sql
CREATE TABLE analytics.call_metrics (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id BIGINT NOT NULL,
    call_id BIGINT,
    agent_id BIGINT,
    direction VARCHAR(10) NOT NULL,
    caller_number VARCHAR(20),
    started_at TIMESTAMP NOT NULL,
    answered_at TIMESTAMP,
    ended_at TIMESTAMP,
    duration_seconds INT,
    wait_time_seconds INT,
    talk_time_seconds INT,
    hold_time_seconds INT,
    hold_count INT,
    transfer_count INT,
    outcome VARCHAR(30),
    hangup_cause VARCHAR(50),
    recording_path TEXT,
    cost DECIMAL(10,4),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
```

### agent_metrics (partitioned)

```sql
CREATE TABLE analytics.agent_metrics (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id BIGINT NOT NULL,
    agent_id BIGINT NOT NULL,
    metric_date DATE NOT NULL,
    conversations_handled INT DEFAULT 0,
    messages_processed INT DEFAULT 0,
    avg_response_time_ms FLOAT,
    tokens_input BIGINT DEFAULT 0,
    tokens_output BIGINT DEFAULT 0,
    tool_calls INT DEFAULT 0,
    errors INT DEFAULT 0,
    escalations INT DEFAULT 0,
    avg_satisfaction FLOAT,
    cost DECIMAL(10,4) DEFAULT 0,
    uptime_minutes INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE UNIQUE INDEX idx_agent_metrics_daily ON analytics.agent_metrics(agent_id, metric_date);
```

### cost_tracking (partitioned)

```sql
CREATE TABLE analytics.cost_tracking (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id BIGINT NOT NULL,
    service VARCHAR(50) NOT NULL,         -- llm | stt | tts | telephony | storage
    model VARCHAR(100),
    operation VARCHAR(100),
    tokens INT,
    duration_seconds INT,
    cost DECIMAL(10,6) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
```

### analytics_snapshots (pre-aggregated)

```sql
CREATE TABLE analytics.snapshots (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id BIGINT NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    dimensions JSONB NOT NULL,
    value DECIMAL(15,4) NOT NULL,
    granularity VARCHAR(20) NOT NULL,     -- minute | hour | day | week | month
    bucket_start TIMESTAMP NOT NULL,
    bucket_end TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, bucket_start)
) PARTITION BY RANGE (bucket_start);

CREATE INDEX idx_snapshots_tenant_metric ON analytics.snapshots(tenant_id, metric_name, bucket_start DESC);
```

---

## 7. REST API Endpoints

| Method | Endpoint | Business Status | HTTP | Description |
|---|---|---|---|---|
| `GET` | `/api/v1/analytics/dashboard?start=...&end=...` | `SUCCESS` | `200` | Get dashboard |
| `GET` | `/api/v1/analytics/realtime` | `SUCCESS` | `200` | Real-time metrics |
| `GET` | `/api/v1/analytics/conversations?start=...&end=...` | `SUCCESS` | `200` | Conversation metrics |
| `GET` | `/api/v1/analytics/calls?start=...&end=...` | `SUCCESS` | `200` | Call metrics |
| `GET` | `/api/v1/analytics/agents?limit=10&offset=0` | `SUCCESS` | `200` | List agent metrics |
| `GET` | `/api/v1/analytics/agents/{agent_id}/performance` | `SUCCESS` | `200` | Agent performance |
| `GET` | `/api/v1/analytics/costs?start=...&end=...` | `SUCCESS` | `200` | Cost breakdown |
| `GET` | `/api/v1/analytics/costs/customers/{customer_id}` | `SUCCESS` | `200` | Customer cost |
| `POST` | `/api/v1/analytics/reports` | `CREATED` | `201` | Generate report |
| `GET` | `/api/v1/analytics/reports?limit=10&offset=0` | `SUCCESS` | `200` | List reports |
| `GET` | `/api/v1/analytics/reports/{id}` | `SUCCESS` | `200` | Get report |

### List Agent Metrics Response

```json
{
  "status": "SUCCESS",
  "data": [...],
  "summary": {
    "total_items": 50,
    "avg_response_time_ms": 1250,
    "avg_csat_score": 4.2,
    "avg_resolution_rate": 0.87,
    "total_tokens_used": 1500000,
    "total_cost": 1250.50,
    "top_performers": 5,
    "underperformers": 2
  },
  "pagination": {"total": 50, "limit": 10, "offset": 0, "has_more": true},
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

**Note:** No PUT method. Use PATCH for updates. All list endpoints support `limit` (default 10) and `offset`.

---

## 8. NATS Events (Subscribed)

| Subject | Handler |
|---|---|
| `aeroxe.v1.conversation.*` | Process conversation events |
| `aeroxe.v1.telephony.call.*` | Process call events |
| `aeroxe.v1.agent.*` | Process agent events |
| `aeroxe.v1.security.*` | Process security events |
| `aeroxe.v1.workflow.*` | Process workflow events |

---

## 9. Observability

| Metric | Description |
|---|---|
| `analytics_events_processed` | Events ingested |
| `analytics_aggregation_latency_ms` | Aggregation time |
| `analytics_query_latency_ms` | Dashboard query time |
| `analytics_snapshots_stored` | Snapshots created |
| `analytics_anomalies_detected` | Anomalies found |

---

## 10. Performance Targets

| Operation | Target |
|---|---|
| Dashboard query | < 200ms |
| Real-time metrics | < 50ms |
| Report generation | < 5s |
| Event ingestion | < 10ms per event |
| Snapshot aggregation | < 1s per bucket |

---

## 11. Call Center Metrics (NEW)

### 11.1 Standard Call Center KPIs

| Metric | Description | Target |
|---|---|---|
| **AHT** (Average Handle Time) | Total talk + hold + wrap-up time | < 5 min |
| **ASA** (Average Speed of Answer) | Time to answer in queue | < 30s |
| **ACW** (After Call Work) | Post-call processing time | < 30s |
| **FCR** (First Call Resolution) | Resolved without callback | > 80% |
| **Abandonment Rate** | Calls abandoned in queue | < 5% |
| **Service Level** | % answered within threshold | > 80% |
| **Occupancy Rate** | Agent busy time / available time | 70-85% |
| **Utilization Rate** | Agent productive time / shift time | > 70% |

### 11.2 Real-Time Call Center Dashboard

```rust
pub struct CallCenterMetrics {
    pub active_calls: u32,
    pub calls_in_queue: u32,
    pub agents_available: u32,
    pub agents_busy: u32,
    pub agents_offline: u32,
    pub avg_wait_time_seconds: f64,
    pub longest_wait_seconds: f64,
    pub avg_handle_time_seconds: f64,
    pub service_level_percent: f64,
    pub abandonment_rate_percent: f64,
    pub calls_today: u64,
    pub avg_talk_time_seconds: f64,
    pub avg_acw_seconds: f64,
}
```

### 11.3 Call Routing Analytics

```rust
pub struct RoutingAnalytics {
    pub route_id: RouteId,
    pub route_name: String,
    pub total_calls: u64,
    pub avg_wait_seconds: f64,
    pub avg_handle_seconds: f64,
    pub success_rate: f64,
    pub transfer_rate: f64,
    pub abandonment_rate: f64,
    pub cost_per_call: f64,
}

pub struct ProviderAnalytics {
    pub provider_name: String,
    pub total_calls: u64,
    pub avg_latency_ms: f64,
    pub packet_loss_percent: f64,
    pub failure_rate: f64,
    pub cost_per_minute: f64,
}
```

---

## 12. Conversation Cost Allocation (NEW)

### 12.1 Cost Tracking

```rust
pub struct ConversationCost {
    pub conversation_id: ConversationId,
    pub tenant_id: TenantId,
    pub llm_cost: f64,
    pub llm_tokens_input: u64,
    pub llm_tokens_output: u64,
    pub stt_cost: f64,
    pub stt_duration_seconds: u32,
    pub tts_cost: f64,
    pub tts_characters: u32,
    pub telephony_cost: f64,
    pub call_duration_seconds: u32,
    pub storage_cost: f64,
    pub total_cost: f64,
    pub cost_per_minute: f64,
}

pub struct CostAlert {
    pub alert_id: AlertId,
    pub tenant_id: TenantId,
    pub alert_type: CostAlertType,
    pub threshold: f64,
    pub current_value: f64,
    pub triggered_at: DateTime,
    pub acknowledged: bool,
}

pub enum CostAlertType {
    DailyLimit,
    MonthlyLimit,
    PerConversationLimit,
    PerCustomerLimit,
    CostPerMinuteThreshold,
}
```

### 12.2 Cost Tracking Entities

```sql
CREATE TABLE analytics.conversation_costs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT,
    llm_cost DECIMAL(10,6) DEFAULT 0,
    llm_tokens_input BIGINT DEFAULT 0,
    llm_tokens_output BIGINT DEFAULT 0,
    stt_cost DECIMAL(10,6) DEFAULT 0,
    stt_duration_seconds INT DEFAULT 0,
    tts_cost DECIMAL(10,6) DEFAULT 0,
    tts_characters INT DEFAULT 0,
    telephony_cost DECIMAL(10,6) DEFAULT 0,
    call_duration_seconds INT DEFAULT 0,
    storage_cost DECIMAL(10,6) DEFAULT 0,
    total_cost DECIMAL(10,6) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conv_costs_tenant ON analytics.conversation_costs(tenant_id, created_at DESC);
```

---

## 13. Agent Performance Scoring (NEW)

### 13.1 Scoring Model

```rust
pub struct AgentPerformanceScore {
    pub agent_id: AgentId,
    pub period: DateRange,
    pub overall_score: f32,          // 0-100
    pub dimensions: Vec<ScoreDimension>,
}

pub struct ScoreDimension {
    pub name: String,                // "response_quality", "resolution_rate", "customer_satisfaction"
    pub score: f32,                  // 0-100
    pub weight: f32,                 // Contribution to overall
    pub trend: TrendDirection,       // Up, Down, Stable
}

pub enum TrendDirection {
    Improving,
    Declining,
    Stable,
}
```

### 13.2 Scoring Factors

| Factor | Weight | Description |
|---|---|---|
| Customer Satisfaction | 30% | CSAT scores |
| Resolution Rate | 25% | Issues resolved without escalation |
| Response Time | 20% | Speed of response |
| First Contact Resolution | 15% | Resolved in first interaction |
| Compliance Score | 10% | Script adherence, policy compliance |

---

## 14. Sentiment Tracking Across Calls (NEW)

### 14.1 Customer Sentiment History

```rust
pub struct CustomerSentimentProfile {
    pub customer_id: CustomerId,
    pub overall_sentiment: f32,      // -1.0 to 1.0
    pub sentiment_trend: TrendDirection,
    pub call_sentiments: Vec<CallSentiment>,
    pub risk_score: f32,             // 0-1.0, risk of churn
    pub last_negative_interaction: Option<DateTime>,
}

pub struct CallSentiment {
    pub call_id: CallId,
    pub start_sentiment: f32,
    pub end_sentiment: f32,
    pub average_sentiment: f32,
    pub lowest_point: f32,
    pub triggers: Vec<String>,       // What caused negative sentiment
}
```

### 14.2 Churn Risk Detection

| Risk Level | Score | Action |
|---|---|---|
| Low | 0.0-0.3 | Monitor |
| Medium | 0.3-0.6 | Proactive outreach |
| High | 0.6-0.8 | Escalate to account manager |
| Critical | 0.8-1.0 | Immediate intervention |

---

## 15. Real-Time Agent Monitoring (NEW)

### 15.1 Live Agent Dashboard

```rust
pub struct LiveAgentStatus {
    pub agent_id: AgentId,
    pub agent_name: String,
    pub status: AgentLiveStatus,
    pub current_call_id: Option<CallId>,
    pub current_conversation_id: Option<ConversationId>,
    pub call_duration_seconds: u32,
    pub calls_today: u32,
    pub avg_csat_today: f32,
    pub current_sentiment: f32,
    pub queue_depth: u32,
}

pub enum AgentLiveStatus {
    Available,
    OnCall,
    InACW,              // After Call Work
    OnBreak,
    Offline,
    Training,
}
```

### 15.2 Supervisor Real-Time View

| View | Data |
|---|---|
| Agent list with status | Who's available, busy, offline |
| Active calls | Live calls with duration and sentiment |
| Queue depth | Calls waiting, longest wait |
| Real-time metrics | Service level, AHT, abandonment |
| Alert feed | Active alerts and escalations |
