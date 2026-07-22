# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 15 — Conversation State Machine Architecture

## State Management, Context Window, Turn-Taking, SLA, Sentiment, GDPR

---

## 1. Conversation Module Overview

The Conversation module manages the lifecycle and state of every interaction:

- Conversation state machine (greeting → intent → gathering → processing → confirming → closing)
- Multi-turn context management
- Turn-taking model (who speaks when)
- Token budget management per conversation
- Conversation branching and correction
- Multi-channel (chat + voice + hybrid)
- SLA management
- Real-time sentiment tracking
- GDPR deletion/export

---

## 2. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-conversation` |
| Bounded Context | Conversation Management |
| Schema | `conversation_` (shared PostgreSQL) |

---

## 3. Conversation State Machine

### States

```
START → GREETING → INTENT_CAPTURED → DATA_GATHERING → PROCESSING
  → CONFIRMING → CLOSING → END
  ↘ ESCALATING ↗ ERROR ↗ ON_HOLD ↗
```

### State Definitions

| State | Description | Allowed Transitions |
|---|---|---|
| `Greeting` | Initial greeting | IntentCaptured, DataGathering |
| `IntentCaptured` | Intent identified | DataGathering, Processing, Escalating |
| `DataGathering` | Collecting information | Processing, Confirming, Error |
| `Processing` | Agent working | Confirming, DataGathering, Escalating |
| `Confirming` | Verifying answer | Closing, DataGathering, Processing |
| `Escalating` | Transferring to human | Closing, DataGathering |
| `Error` | Something went wrong | DataGathering, Processing, Closing |
| `Closing` | Wrapping up | End |
| `OnHold` | Paused | Previous state |
| `End` | Complete | (terminal) |

---

## 4. Token Budget Management

| Component | Budget | Purpose |
|---|---|---|
| System prompt | ~2000 tokens | Agent personality + rules |
| Conversation history | ~4000 tokens | Recent messages |
| RAG context | ~2000 tokens | Retrieved knowledge |
| Tool results | ~1000 tokens | Tool call outputs |
| Response generation | ~1000 tokens | Agent response |

---

## 5. Turn-Taking Model

| Turn | Actor | Action |
|---|---|---|
| UserTurn | Customer | Sends message / speaks |
| AgentTurn | AI Agent | Processes + responds |
| ToolTurn | System | Executes tool call |
| TransferTurn | System | Handoff to human |

---

## 6. SLA Management

```rust
pub struct ConversationSLA {
    pub sla_id: SLAId,
    pub tenant_id: TenantId,
    pub rules: Vec<SLARule>,
}

pub struct SLARule {
    pub metric: SLAMetric,
    pub target_seconds: u32,
    pub warning_seconds: u32,
    pub breach_action: SLABreachAction,
}

pub enum SLAMetric {
    FirstResponseTime, ResolutionTime, EscalationTime, CustomerWaitTime,
}

pub enum SLABreachAction {
    NotifySupervisor, EscalateToHuman, LogWarning, AutoTransfer,
}
```

---

## 7. Sentiment Tracking

```rust
pub struct SentimentResult {
    pub score: f32,                   // -1.0 to 1.0
    pub label: SentimentLabel,
    pub confidence: f32,
    pub keywords: Vec<String>,
}

pub enum SentimentLabel {
    VeryNegative, Negative, Neutral, Positive, VeryPositive,
}
```

---

## 8. GDPR Deletion/Export

### Export Flow

```
Customer Requests Export → Verify Identity → Collect Data
  → Package (JSON) → Deliver (download link)
```

### Deletion Flow

```
Customer Requests Deletion → Verify Identity → Mark for Deletion
  → Cascading Deletion → Confirm Deletion (certificate)
```

---

## 9. REST API Endpoints

| Method | Endpoint | Business Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/conversations` | `CREATED` | `201` |
| `POST` | `/api/v1/conversations/{id}` | `SUCCESS` | `200` |
| `POST` | `/api/v1/conversations?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/conversations/{id}/messages?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/conversations/{id}/messages` | `CREATED` | `201` |
| `POST` | `/api/v1/conversations/{id}/state` | `SUCCESS` | `200` |
| `POST` | `/api/v1/conversations/{id}/end` | `UPDATED` | `200` |
| `POST` | `/api/v1/conversations/{id}/branch` | `CREATED` | `201` |
| `DELETE` | `/api/v1/conversations/{id}` | `DELETED` | `204` |

---

## 10. Database Tables

| Table | Purpose |
|---|---|
| `conversation_.conversations` | Conversation records |
| `conversation_.messages` | Conversation messages |
| `conversation_.entities` | Extracted entities |
| `conversation_.sentiment` | Per-message sentiment |
| `conversation_.post_call_surveys` | Post-call satisfaction surveys |

---

# End of Part 15
