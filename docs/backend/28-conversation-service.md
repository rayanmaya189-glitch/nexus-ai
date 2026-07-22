# AeroXe Nexus AI — Conversation Module

## Conversation State Machine, Context Management, Turn-Taking & Flow Control

> **Modular Monolith Module:** This document describes the `nexus-conversation` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces.

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-conversation` |
| Crate | `nexus-conversation` (workspace member) |
| Bounded Context | Conversation Management |
| Domain Type | Core Domain |
| Language | Rust (edition 2024) |
| Schema | `conversation_` (in shared PostgreSQL) |
| Dependencies | `nexus-memory` (context), `nexus-agent` (processing), `nexus-telephony` (voice channel), `nexus-audit` (logging) |

---

## 2. Purpose

The Conversation module manages the lifecycle and state of every interaction:

- Conversation state machine (greeting → intent → gathering → processing → confirming → closing)
- Multi-turn context management
- Turn-taking model (who speaks when)
- Token budget management per conversation
- Conversation branching and correction
- Multi-channel conversation (chat + voice + hybrid)
- Conversation handoff between channels

---

## 3. Aggregate Design

### Conversation Aggregate

```
Conversation (Aggregate Root)
├── ConversationMetadata
│   ├── ConversationId (UUID)
│   ├── TenantId
│   ├── CustomerId
│   ├── Channel (Chat | Voice | Hybrid)
│   ├── StartedAt
│   ├── EndedAt
│   └── Duration
├── ConversationState
│   ├── CurrentState (ConversationState enum)
│   ├── StateHistory[]
│   ├── TurnCount
│   └── LastActivityAt
├── Context
│   ├── SystemPrompt
│   ├── ConversationSummary
│   ├── Entities[]
│   ├── Entities[]
│   ├── IntentStack[]
│   └── TokenBudget
├── Participants
│   ├── Customer
│   │   ├── CustomerId
│   │   ├── Channel
│   │   └── AuthLevel
│   └── Agent
│       ├── AgentId
│       ├── Model
│       └── Personality
├── Messages[]
│   ├── MessageId
│   ├── Role (User | Assistant | System | Tool)
│   ├── Content
│   ├── Tokens
│   ├── Timestamp
│   └── Metadata
└── Outcome
    ├── Resolution (Resolved | Escalated | Abandoned | Transferred)
    ├── SatisfactionScore
    └── Summary
```

---

## 4. Conversation State Machine

### 4.1 States

```
                    ┌─────────────┐
                    │   START     │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  GREETING   │
                    └──────┬──────┘
                           │ User responds
                    ┌──────▼──────┐
                    │   INTENT    │
                    │  CAPTURED   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
     ┌────────▼──────┐ ┌──▼──────┐ ┌──▼──────────┐
     │  DATA GATHER  │ │PROCESS  │ │ ESCALATING  │
     │   (questions) │ │         │ │             │
     └────────┬──────┘ └──┬──────┘ └──┬──────────┘
              │            │            │
              └────────────┼────────────┘
                           │
                    ┌──────▼──────┐
                    │ CONFIRMING  │
                    │  (verify)   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │                         │
     ┌────────▼──────┐          ┌───────▼──────┐
     │   CLOSING     │          │   ERROR      │
     │ (resolution)  │          │  (retry)     │
     └────────┬──────┘          └───────┬──────┘
              │                         │
              └─────────────┬───────────┘
                           │
                    ┌──────▼──────┐
                    │    END      │
                    └─────────────┘
```

### 4.2 State Definitions

| State | Description | Allowed Transitions |
|---|---|---|
| `Greeting` | Initial greeting, waiting for user | `IntentCaptured`, `DataGathering` |
| `IntentCaptured` | User's intent identified | `DataGathering`, `Processing`, `Escalating` |
| `DataGathering` | Collecting needed information | `Processing`, `Confirming`, `Error` |
| `Processing` | Agent is working (tool calls, RAG, etc.) | `Confirming`, `DataGathering`, `Escalating` |
| `Confirming` | Verifying answer with user | `Closing`, `DataGathering`, `Processing` |
| `Escalating` | Transferring to human/specialist | `Closing`, `DataGathering` |
| `Error` | Something went wrong, retrying | `DataGathering`, `Processing`, `Closing` |
| `Closing` | Wrapping up conversation | `End` |
| `OnHold` | Paused (user away) | Previous state (resume) |
| `End` | Conversation complete | (terminal) |

### 4.3 State Transition Rules

```rust
pub enum ConversationState {
    Greeting,
    IntentCaptured,
    DataGathering,
    Processing,
    Confirming,
    Escalating,
    Error,
    Closing,
    OnHold,
    End,
}

pub struct StateTransition {
    pub from: ConversationState,
    pub to: ConversationState,
    pub trigger: TransitionTrigger,
    pub guard: Option<Box<dyn Fn(&Conversation) -> bool>>,
}

pub enum TransitionTrigger {
    UserMessage,
    AgentResponse,
    ToolResult,
    Timeout,
    HumanRequest,
    Error,
    CustomerDisconnected,
}
```

---

## 5. Token Budget Management

### 5.1 Budget Allocation

| Component | Budget | Purpose |
|---|---|---|
| System prompt | ~2000 tokens | Agent personality + rules |
| Conversation history | ~4000 tokens | Recent messages |
| RAG context | ~2000 tokens | Retrieved knowledge |
| Tool results | ~1000 tokens | Tool call outputs |
| Response generation | ~1000 tokens | Agent response |
| **Total budget** | **~10000 tokens** | Per turn |

### 5.2 Context Window Management

```
Full Conversation History
    |
    v
[1] Token counting per message
    |
    v
[2] Budget allocation
    |  - System prompt: fixed
    |  - Recent messages: sliding window
    |  - Important messages: prioritized by relevance
    |  - Summarized old messages: compressed
    |
    v
[3] Context assembly
    |  - System prompt
    |  - Conversation summary (old messages)
    |  - Recent messages (sliding window)
    |  - RAG context
    |  - Tool results
    |
    v
[4] LLM prompt
    |
    v
[5] Response
    |
    v
[6] Update context
    |  - Add new messages
    |  - Re-count tokens
    |  - Summarize if over budget
```

---

## 6. Turn-Taking Model

### 6.1 Turn Types

| Turn | Actor | Action |
|---|---|---|
| `UserTurn` | Customer | Sends message / speaks |
| `AgentTurn` | AI Agent | Processes + responds |
| `ToolTurn` | System | Executes tool call |
| `SystemTurn` | System | System message (timeout, etc.) |
| `TransferTurn` | System | Handoff to human |

### 6.2 Turn Rules

| Rule | Description |
|---|---|
| One speaker at a time | No simultaneous responses |
| Agent waits for user | After response, agent yields |
| Tool calls are internal | Tool calls don't count as agent turns |
| Barge-in handling | In voice, caller can interrupt agent |
| Max consecutive turns | Agent can ask max 3 questions before responding |
| Turn timeout | 30s chat, 60s voice before auto-continuing |

---

## 7. Public API Trait

```rust
// nexus-conversation/src/interfaces/api.rs
#[async_trait]
pub trait ConversationService: Send + Sync {
    // Conversation lifecycle
    async fn create_conversation(&self, req: CreateConversationRequest) -> Result<Conversation, ConversationError>;
    async fn get_conversation(&self, id: ConversationId) -> Result<Option<Conversation>, ConversationError>;
    async fn end_conversation(&self, id: ConversationId, outcome: ConversationOutcome) -> Result<(), ConversationError>;
    
    // State management
    async fn transition_state(&self, id: ConversationId, trigger: TransitionTrigger) -> Result<ConversationState, ConversationError>;
    async fn get_state(&self, id: ConversationId) -> Result<ConversationState, ConversationError>;
    async def pause_conversation(&self, id: ConversationId) -> Result<(), ConversationError>;
    async fn resume_conversation(&self, id: ConversationId) -> Result<(), ConversationError>;
    
    // Message handling
    async fn add_message(&self, id: ConversationId, msg: NewMessage) -> Result<Message, ConversationError>;
    async fn get_messages(&self, id: ConversationId, limit: u32) -> Result<Vec<Message>, ConversationError>;
    async fn get_context(&self, id: ConversationId) -> Result<ConversationContext, ConversationError>;
    
    // Token management
    async fn get_token_usage(&self, id: ConversationId) -> Result<TokenUsage, ConversationError>;
    async fn summarize_old_messages(&self, id: ConversationId) -> Result<(), ConversationError>;
    
    // Branching
    async fn branch_conversation(&self, id: ConversationId, from_message: MessageId) -> Result<Conversation, ConversationError>;
    async fn edit_message(&self, id: ConversationId, msg_id: MessageId, new_content: String) -> Result<Message, ConversationError>;
    
    // Query
    async fn list_conversations(&self, query: ConversationQuery) -> Result<Vec<ConversationSummary>, ConversationError>;
    async fn get_active_conversations(&self, tenant_id: TenantId) -> Result<Vec<ConversationSummary>, ConversationError>;
}

pub struct CreateConversationRequest {
    pub tenant_id: TenantId,
    pub customer_id: Option<CustomerId>,
    pub channel: ConversationChannel,
    pub agent_id: Option<AgentId>,
    pub initial_message: Option<String>,
    pub metadata: HashMap<String, String>,
}

pub enum ConversationChannel {
    Chat,
    Voice,
    Hybrid,    // Started as chat, escalated to voice
    Email,
    WhatsApp,
    SMS,
}

pub struct NewMessage {
    pub role: MessageRole,
    pub content: String,
    pub tokens: Option<u32>,
    pub metadata: Option<HashMap<String, String>>,
}

pub enum MessageRole {
    User,
    Assistant,
    System,
    Tool,
}

pub struct ConversationContext {
    pub conversation_id: ConversationId,
    pub state: ConversationState,
    pub messages: Vec<Message>,
    pub summary: Option<String>,
    pub entities: Vec<Entity>,
    pub token_usage: TokenUsage,
    pub system_prompt: String,
}

pub struct TokenUsage {
    pub total_tokens: u32,
    pub prompt_tokens: u32,
    pub completion_tokens: u32,
    pub budget_remaining: u32,
}
```

---

## 8. Database Schema (conversation_ schema)

### conversations

```sql
CREATE TABLE conversation_.conversations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT,
    agent_id BIGINT,
    channel VARCHAR(20) NOT NULL DEFAULT 'chat',
    state VARCHAR(30) NOT NULL DEFAULT 'greeting',
    state_history JSONB,
    turn_count INT NOT NULL DEFAULT 0,
    token_usage JSONB,
    summary TEXT,
    outcome VARCHAR(30),
    satisfaction_score INT,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP,
    last_activity_at TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conv_tenant ON conversation_.conversations(tenant_id, created_at DESC);
CREATE INDEX idx_conv_customer ON conversation_.conversations(customer_id, created_at DESC);
CREATE INDEX idx_conv_state ON conversation_.conversations(state) WHERE state != 'end';
CREATE INDEX idx_conv_active ON conversation_.conversations(last_activity_at DESC) WHERE state != 'end';
```

### conversation_messages

```sql
CREATE TABLE conversation_.messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversation_.conversations(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    message_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL,            -- user | assistant | system | tool
    content TEXT NOT NULL,
    tokens INT,
    tool_name VARCHAR(100),
    tool_input JSONB,
    tool_output JSONB,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conv_msg_conv ON conversation_.messages(conversation_id, created_at);
```

### conversation_entities

```sql
CREATE TABLE conversation_.entities (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversation_.conversations(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    entity_type VARCHAR(50) NOT NULL,     -- customer_id, product, ticket, etc.
    entity_value TEXT NOT NULL,
    confidence FLOAT,
    extracted_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 9. REST API Endpoints

| Method | Endpoint | Business Status | HTTP | Description |
|---|---|---|---|---|
| `POST` | `/api/v1/conversations` | `CREATED` | `201` | Create conversation |
| `GET` | `/api/v1/conversations/{id}` | `SUCCESS` | `200` | Get conversation |
| `GET` | `/api/v1/conversations?limit=10&offset=0` | `SUCCESS` | `200` | List conversations |
| `GET` | `/api/v1/conversations/{id}/messages?limit=10&offset=0` | `SUCCESS` | `200` | Get messages |
| `POST` | `/api/v1/conversations/{id}/messages` | `CREATED` | `201` | Send message |
| `GET` | `/api/v1/conversations/{id}/state` | `SUCCESS` | `200` | Get state |
| `POST` | `/api/v1/conversations/{id}/end` | `UPDATED` | `200` | End conversation |
| `POST` | `/api/v1/conversations/{id}/branch` | `CREATED` | `201` | Branch conversation |
| `DELETE` | `/api/v1/conversations/{id}` | `DELETED` | `204` | Delete conversation |

### List Conversations Response

```json
{
  "status": "SUCCESS",
  "data": [...],
  "summary": {
    "total_items": 2000,
    "active_conversations": 42,
    "completed_conversations": 1900,
    "escalated_conversations": 58,
    "avg_duration_seconds": 342,
    "avg_turns": 8.5,
    "avg_csat_score": 4.2,
    "recent_activity": {
      "created_today": 150,
      "completed_today": 140,
      "escalated_today": 10
    }
  },
  "pagination": {"total": 2000, "limit": 10, "offset": 0, "has_more": true},
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

**Note:** No PUT method. Use PATCH for updates. All list endpoints support `limit` (default 10) and `offset`.

---

## 10. NATS Events

| Subject | Event |
|---|---|
| `aeroxe.v1.conversation_.created` | `ConversationCreated` |
| `aeroxe.v1.conversation_.state.changed` | `ConversationStateChanged` |
| `aeroxe.v1.conversation_.message.added` | `MessageAdded` |
| `aeroxe.v1.conversation_.ended` | `ConversationEnded` |
| `aeroxe.v1.conversation_.timeout` | `ConversationTimeout` |
| `aeroxe.v1.conversation_.escalated` | `ConversationEscalated` |

---

## 11. Observability

| Metric | Description |
|---|---|
| `conversation_active` | Active conversations |
| `conversation_created_total` | Conversations created |
| `conversation_ended_total` | Conversations ended |
| `conversation_duration_seconds` | Conversation duration |
| `conversation_turns_total` | Turns per conversation |
| `conversation_tokens_used` | Token consumption |
| `conversation_state_transitions` | State changes |
| `conversation_timeout_total` | Timed-out conversations |
| `conversation_escalation_total` | Escalated conversations |
| `conversation_satisfaction_score` | CSAT distribution |

---

## 12. Performance Targets

| Operation | Target |
|---|---|
| Create conversation | < 50ms |
| Add message | < 20ms |
| Get context | < 100ms |
| State transition | < 10ms |
| Token counting | < 10ms |
| Summarize messages | < 2s |
| Branch conversation | < 100ms |

---

## 13. Conversation SLA Management (NEW)

### 13.1 SLA Definitions

```rust
pub struct ConversationSLA {
    pub sla_id: SLAId,
    pub tenant_id: TenantId,
    pub name: String,
    pub channel: Option<ConversationChannel>,  // NULL = all channels
    pub priority: u32,
    rules: Vec<SLARule>,
}

pub struct SLARule {
    pub metric: SLAMetric,
    pub target_seconds: u32,
    pub warning_seconds: u32,       // Alert before breach
    pub breach_action: SLABreachAction,
}

pub enum SLAMetric {
    FirstResponseTime,              // Time to first agent response
    ResolutionTime,                 // Time to resolve issue
    EscalationTime,                 // Time to escalate to human
    CustomerWaitTime,               // Time customer waits in queue
}

pub enum SLABreachAction {
    NotifySupervisor,
    EscalateToHuman,
    LogWarning,
    AutoTransfer,
}
```

### 13.2 SLA Enforcement

```
Conversation Created
    |
    v
[1] Load SLA Rules
    |  - Match tenant + channel + priority
    |
    v
[2] Start SLA Timer
    |  - Track first response time
    |  - Track resolution time
    |
    v
[3] Monitor During Conversation
    |  - Warning threshold: Alert supervisor
    |  - Breach: Execute breach action
    |
    v
[4] SLA Complete
    |  - Record actual vs target
    |  - Update SLA compliance metrics
```

---

## 14. Real-Time Sentiment Tracking (NEW)

### 14.1 Sentiment Analysis

```rust
pub struct SentimentResult {
    pub score: f32,                   // -1.0 (negative) to 1.0 (positive)
    pub label: SentimentLabel,
    pub confidence: f32,
    pub keywords: Vec<String>,
}

pub enum SentimentLabel {
    VeryNegative,
    Negative,
    Neutral,
    Positive,
    VeryPositive,
}

pub struct SentimentAlert {
    pub conversation_id: ConversationId,
    pub trigger: SentimentTrigger,
    pub severity: AlertSeverity,
    pub action: AlertAction,
}

pub enum SentimentTrigger {
    ScoreBelow(f32),                  // e.g., score < -0.5
    RapidDecline(f32),                // e.g., dropped 0.3 in 2 messages
    ConsecutiveNegative(i32),         // e.g., 3 negative messages in row
}

pub enum AlertAction {
    LogOnly,
    NotifySupervisor,
    AutoEscalate,
    TriggerCoaching,
}
```

### 14.2 Sentiment Tracking Flow

```
Each Message in Conversation
    |
    v
[1] Analyze Sentiment
    |  - LLM-based or ML model
    |  - Score + label + confidence
    |
    v
[2] Update Conversation Sentiment
    |  - Store sentiment per message
    |  - Calculate rolling average
    |  - Detect trends
    |
    v
[3] Evaluate Alerts
    |  - Check against thresholds
    |  - Trigger alerts if needed
    |
    v
[4] Adapt Agent Behavior
    |  - Positive: Maintain approach
    |  - Negative: Switch to empathetic mode
    |  - Very negative: Escalate to human
```

### 14.3 Sentiment Entities

```sql
CREATE TABLE conversation_.sentiment (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversation_.conversations(id) ON DELETE CASCADE,
    message_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    score FLOAT NOT NULL,              -- -1.0 to 1.0
    label VARCHAR(20) NOT NULL,
    confidence FLOAT NOT NULL,
    keywords JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 15. Conversation Deletion & Export (GDPR) (NEW)

### 15.1 GDPR Export Flow

```
Customer Requests Data Export
    |
    v
[1] Verify Identity
    |  - Confirm customer identity
    |  - Log export request
    |
    v
[2] Collect Data
    |  - All conversations
    |  - All messages
    |  - All recordings (if voice)
    |  - All transcripts
    |  - All sentiment data
    |  - All entity extractions
    |
    v
[3] Package Data
    |  - JSON format
    |  - Include metadata
    |  - Remove internal IDs
    |
    v
[4] Deliver
    |  - Download link (time-limited)
    |  - Or email delivery
```

### 15.2 GDPR Deletion Flow

```
Customer Requests Deletion
    |
    v
[1] Verify Identity + Authorization
    |  - Confirm customer owns data
    |  - Check retention requirements (some data must be kept for compliance)
    |
    v
[2] Mark for Deletion
    |  - Anonymize conversation content
    |  - Delete recordings
    |  - Delete transcripts
    |  - Keep anonymized audit trail (compliance requirement)
    |
    v
[3] Cascading Deletion
    |  - Delete from conversation schema
    |  - Delete from memory schema
    |  - Delete from telephony schema
    |  - Delete from analytics (anonymize)
    |
    v
[4] Confirm Deletion
    |  - Generate deletion certificate
    |  - Log deletion event
    |  - Notify customer
```

### 15.3 Data Retention Configuration

```rust
pub struct DataRetentionPolicy {
    pub tenant_id: TenantId,
    pub data_type: DataType,
    pub retention_days: u32,
    pub deletion_method: DeletionMethod,
    pub compliance_hold: bool,         // Prevent deletion during legal hold
}

pub enum DataType {
    Conversation,
    Message,
    Recording,
    Transcript,
    Sentiment,
    Entity,
    AuditLog,     // Never fully deleted, anonymized only
}

pub enum DeletionMethod {
    HardDelete,    // Completely remove
    SoftDelete,    // Mark as deleted, purge later
    Anonymize,     // Replace with anonymized data
}
```

---

## 16. Post-Call Survey

Post-call surveys collect customer satisfaction feedback after conversations end.

### 16.1 Survey Configuration

- survey_enabled: bool
- survey_delay_seconds: u32 (default 5)
- rating_scale: u8 (default 5)
- questions: Vec<SurveyQuestion>

### 16.2 Survey Flow

1. Conversation ends
2. Wait survey_delay_seconds
3. Play/ask survey prompt via TTS
4. Collect DTMF/speech rating (1-5 scale)
5. Optional free-text comment
6. Store in post_call_surveys table

### 16.3 Database Table

```sql
CREATE TABLE conversation_.post_call_surveys (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    channel VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 16.4 NATS Events

- aeroxe.v1.conversation.survey.completed

---

## 17. Conversation Timeout

Conversations are automatically managed when participants become inactive.

### 17.1 Timeout Configuration

```rust
pub struct ConversationTimeoutConfig {
    pub inactivity_timeout_seconds: u32,  // default 300 (5 min)
    pub warning_threshold_seconds: u32,   // default 240 (4 min)
    pub auto_end_on_timeout: bool,        // default true
    pub max_idle_conversations: usize,    // default 10000
}
```

### 17.2 Timeout Flow

1. Timer starts after each message
2. At warning_threshold: send "Are you still there?" prompt
3. At inactivity_timeout: log timeout event, end conversation
4. Archive conversation with outcome=timeout
5. Publish NATS event: aeroxe.v1.conversation.timeout
