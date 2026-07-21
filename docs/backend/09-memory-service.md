# AeroXe Nexus AI — Memory Module

## Short-Term, Long-Term & Organizational AI Memory

> **Modular Monolith Module:** This document describes the `nexus-memory` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces (see [Communication Architecture](12-communication-architecture.md)).

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-memory` |
| Crate | `nexus-memory` (workspace member) |
| Bounded Context | Memory |
| Domain Type | Supporting Domain |
| Language | Rust |
| Schema | `memory_` (in shared PostgreSQL + pgvector) + Redis |
| Dependencies | Ollama (embeddings), Redis (short-term), PostgreSQL (long-term) |

---

## 2. Purpose

The Memory module provides persistent AI memory across conversations and sessions within the `aeroxe-nexus` monolith. It stores:

- Current conversation context (short-term)
- User preferences and past interactions (long-term)
- Entity relationships (organizational)

---

## 3. Aggregate Design

### MemoryProfile Aggregate

```
MemoryProfile (Aggregate Root)
├── ShortTermMemory (Redis)
│   ├── CurrentConversation
│   ├── ActiveTasks
│   └── TemporaryContext
├── LongTermMemory (PostgreSQL + pgvector)
│   ├── UserPreferences
│   ├── PastConversations
│   └── ImportantFacts
└── OrganizationalMemory (Apache AGE)
    ├── EntityRelationships
    └── BusinessKnowledge
```

---

## 4. Three-Layer Memory Architecture

### 4.1 Short-Term Memory

| Attribute | Value |
|---|---|
| Technology | Redis |
| Key Pattern | `conversation:{user_id}:{session_id}` |
| TTL | 24 hours |
| Purpose | Current conversation, temporary context, active tasks |

**Redis Data Structure:**
```
conversation:user123:session456
├── messages[] (List)
│   ├── {role: "user", content: "My internet is slow"}
│   └── {role: "assistant", content: "Let me check..."}
├── context{} (Hash)
│   ├── customer_id: "cust_789"
│   ├── ticket_id: "tkt_456"
│   └── network_status: "degraded"
└── active_tasks[] (Set)
    └── "check_network_status"
```

### 4.2 Long-Term Memory

| Attribute | Value |
|---|---|
| Technology | PostgreSQL + pgvector |
| Embedding Dimension | 768 |
| Purpose | User preferences, past conversations, important facts |

**Use Cases:**
- "Customer prefers Hindi support"
- "User often asks about billing"
- "Customer had network issue on 2026-07-10"
- "User is technical administrator"

### 4.3 Organizational Memory

| Attribute | Value |
|---|---|
| Technology | Apache AGE (Knowledge Graph) |
| Purpose | Entity relationships, business knowledge |

**Example Graph:**
```
Customer --uses--> Product --related_to--> Issue
Customer --located_in--> City --has--> NetworkSegment
Customer --has--> Subscription --connected_to--> Device
```

---

## 5. Public API Trait

```rust
// nexus-memory/src/interfaces/api.rs
#[async_trait]
pub trait MemoryService: Send + Sync {
    async fn store(&self, req: StoreMemoryRequest) -> Result<(), MemoryError>;
    async fn search(&self, req: SearchMemoryRequest) -> Result<Vec<MemoryItem>, MemoryError>;
    async fn get_conversation_context(&self, session_id: SessionId) -> Result<Vec<Message>, MemoryError>;
    async fn clear_session(&self, session_id: SessionId) -> Result<(), MemoryError>;
}

pub struct StoreMemoryRequest {
    pub user_id: UserId,
    pub tenant_id: TenantId,
    pub content: String,
    pub memory_type: MemoryType, // Preference, Fact, Conversation, Context
    pub importance: f32,
    pub metadata: HashMap<String, String>,
}

pub struct SearchMemoryRequest {
    pub user_id: UserId,
    pub tenant_id: TenantId,
    pub query: String,
    pub limit: u32,
    pub memory_type: MemoryTypeFilter,
}

pub struct MemoryItem {
    pub id: String,
    pub content: String,
    pub similarity: f32,
    pub memory_type: MemoryType,
    pub created_at: String,
}
```

> **Note:** `MemoryService` is consumed by `nexus-agent` during execution (context retrieval) and `nexus-ai-gateway` (session management) — all via in-process trait dispatch.

---

## 6. Memory Operations

### Store Long-Term Memory

```
Memory to store: "Customer prefers Hindi support"
    |
    v
[1] Generate Embedding (via Ollama)
    |
    v
[2] Store in PostgreSQL
    |  - content
    |  - embedding (vector(768))
    |  - importance score
    |  - user_id, tenant_id
    |
    v
[3] Update Knowledge Graph (if relationship)
```

### Retrieve Relevant Memory

```
Current context: "User asking about billing"
    |
    v
[1] Generate Query Embedding
    |
    v
[2] Vector Search (pgvector)
    |  - Find semantically similar memories
    |  - Filter by tenant_id
    |  - Filter by user_id
    |  - Similarity threshold > 0.7
    |
    v
[3] Combine with Short-Term (Redis)
    |  - Merge current conversation context
    |
    v
[4] Rank by relevance + recency + importance
    |
    v
[5] Return top results as context for LLM
```

---

## 7. Database Schema (memory_db)

### memories

```sql
CREATE TABLE memory.memories (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(768),
    memory_type VARCHAR(50) NOT NULL DEFAULT 'fact',
    importance FLOAT NOT NULL DEFAULT 0.5,
    access_count INT NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMP,
    expires_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_memories_embedding ON memory.memories
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);
```

### conversation_history

```sql
CREATE TABLE memory.conversation_history (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL, -- 'user' or 'assistant'
    content TEXT NOT NULL,
    tokens_used INT,
    model VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversation_session ON memory.conversation_history(session_id);
CREATE INDEX idx_conversation_user ON memory.conversation_history(user_id, created_at DESC);
```

---

## 8. REST API Endpoints

### Store Memory

```
POST /api/v1/memory
```

**Request:**
```json
{
  "user_id": "123",
  "content": "Customer prefers Hindi support",
  "type": "preference",
  "importance": 0.8
}
```

### Search Memory

```
GET /api/v1/memory/search?q=customer+preferences&limit=5
```

### Get Conversation Context

```
GET /api/v1/memory/context/{session_id}
```

---

## 9. Memory Lifecycle

```
Short-Term (Redis)
    |
    ├── TTL: 24 hours
    ├── Auto-cleanup on expiry
    └── Promotes to Long-Term if important
    |
    v
Long-Term (PostgreSQL + pgvector)
    |
    ├── Permanent storage
    ├── Vector-indexed for semantic search
    ├── Importance decay over time
    └── Consolidation (merge similar memories)
    |
    v
Organizational (Apache AGE)
    |
    ├── Relationship-based storage
    ├── Graph traversal for context
    └── Entity-relationship intelligence
```

---

## 10. NATS Events

| Subject | Event |
|---|---|
| `aeroxe.memory.created` | New memory stored |
| `aeroxe.memory.updated` | Memory modified |
| `aeroxe.memory.expired` | Short-term memory expired |
| `aeroxe.memory.consolidated` | Similar memories merged |

---

## 11. Context Window Management (NEW)

### Token Budget Strategy

| Component | Budget | Purpose |
|---|---|---|
| System prompt | ~2000 tokens | Agent personality + rules |
| Conversation history | ~4000 tokens | Recent messages (sliding window) |
| RAG context | ~2000 tokens | Retrieved knowledge |
| Tool results | ~1000 tokens | Tool call outputs |
| Response generation | ~1000 tokens | Agent response |
| **Total per turn** | **~10000 tokens** | Context limit |

### Context Assembly Algorithm

```
Full Conversation History
    |
    v
[1] Count tokens per message
    |
    v
[2] Identify important messages (entity mentions, decisions)
    |
    v
[3] Summarize old messages (beyond sliding window)
    |
    v
[4] Assemble context:
    |  - System prompt (fixed)
    |  - Conversation summary (compressed old messages)
    |  - Recent messages (sliding window, full detail)
    |  - RAG context (relevant knowledge)
    |  - Tool results (recent tool outputs)
    |
    v
[5] Validate total tokens < budget
    |
    v
[6] Send to LLM
```

### Summarization Strategy

| Trigger | Action |
|---|---|
| Token count > 80% budget | Summarize oldest 50% of messages |
| Conversation > 20 turns | Summarize first 10 turns |
| Explicit user request | Generate full summary |
| Conversation end | Store summary in long-term memory |

---

## 12. Performance Targets

| Operation | Target |
|---|---|
| Short-Term Read (Redis) | < 10ms |
| Short-Term Write (Redis) | < 5ms |
| Long-Term Vector Search | < 200ms |
| Long-Term Write | < 100ms |
| Embedding Generation | < 500ms |
| Memory Consolidation | Background job |
| Context Assembly | < 100ms |
| Token Counting | < 10ms |
| Message Summarization | < 2s |
