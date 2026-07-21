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
CREATE TABLE conversation.conversations (
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

CREATE INDEX idx_conv_tenant ON conversation.conversations(tenant_id, created_at DESC);
CREATE INDEX idx_conv_customer ON conversation.conversations(customer_id, created_at DESC);
CREATE INDEX idx_conv_state ON conversation.conversations(state) WHERE state != 'end';
CREATE INDEX idx_conv_active ON conversation.conversations(last_activity_at DESC) WHERE state != 'end';
```

### conversation_messages

```sql
CREATE TABLE conversation.messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversation.conversations(id) ON DELETE CASCADE,
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

CREATE INDEX idx_conv_msg_conv ON conversation.messages(conversation_id, created_at);
```

### conversation_entities

```sql
CREATE TABLE conversation.entities (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversation.conversations(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    entity_type VARCHAR(50) NOT NULL,     -- customer_id, product, ticket, etc.
    entity_value TEXT NOT NULL,
    confidence FLOAT,
    extracted_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 9. REST API Endpoints

### Create Conversation

```
POST /api/v1/conversations
Authorization: Bearer <jwt>
```

**Request:**
```json
{
  "channel": "chat",
  "customer_id": 123,
  "agent_id": "support-agent",
  "initial_message": "Hello, I need help with my internet"
}
```

### Send Message

```
POST /api/v1/conversations/{id}/messages
Authorization: Bearer <jwt>
```

**Request:**
```json
{
  "role": "user",
  "content": "My internet has been slow for 3 days"
}
```

### Get Conversation

```
GET /api/v1/conversations/{id}
```

### Get Messages

```
GET /api/v1/conversations/{id}/messages?limit=50
```

### End Conversation

```
POST /api/v1/conversations/{id}/end
```

**Request:**
```json
{
  "outcome": "resolved",
  "satisfaction_score": 5,
  "summary": "Resolved slow internet by resetting ONU"
}
```

### Branch Conversation

```
POST /api/v1/conversations/{id}/branch
```

**Request:**
```json
{
  "from_message_id": "uuid",
  "new_content": "Actually, let me try a different approach"
}
```

---

## 10. NATS Events

| Subject | Event |
|---|---|
| `aeroxe.v1.conversation.created` | `ConversationCreated` |
| `aeroxe.v1.conversation.state.changed` | `ConversationStateChanged` |
| `aeroxe.v1.conversation.message.added` | `MessageAdded` |
| `aeroxe.v1.conversation.ended` | `ConversationEnded` |
| `aeroxe.v1.conversation.timeout` | `ConversationTimeout` |
| `aeroxe.v1.conversation.escalated` | `ConversationEscalated` |

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
