# AeroXe Nexus AI — AI Gateway Module

## Central AI Request Processing, Routing & Management

> **Modular Monolith Module:** This document describes the `nexus-ai-gateway` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces (see [Communication Architecture](12-communication-architecture.md)), not gRPC.

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-ai-gateway` |
| Crate | `nexus-ai-gateway` (workspace member) |
| Bounded Context | AI Gateway |
| Domain Type | Core Domain |
| Language | Rust |
| Schema | `ai_` (in shared PostgreSQL) |
| Ports | Internal trait dispatch (no gRPC port — optional: 50051 for external gRPC) |

---

## 2. Purpose

The AI Gateway module is the entry point for all AI requests within the AeroXe Nexus AI monolith. It handles:

- Request reception and validation (from `nexus-gateway` via trait call)
- Authentication and authorization enforcement (delegates to `nexus-identity` trait)
- Rate limiting enforcement
- Request routing to appropriate AI agents (`nexus-agent` trait)
- Response aggregation and streaming (via tokio channels)
- Session management (Redis + PostgreSQL)
- Audit event publishing (NATS)

---

## 3. Aggregate Design

### AIRequest Aggregate

```
AIRequest (Aggregate Root)
├── RequestContext
│   ├── RequestId (UUID)
│   ├── TenantId (i64)
│   ├── UserId (i64)
│   └── TraceId (string)
├── SecurityContext
│   ├── JWT Token
│   ├── Permissions
│   └── Tenant Scope
└── ExecutionPlan
    ├── TargetAgent
    ├── ModelPreference
    └── Priority
```

### Entities

| Entity | Description |
|---|---|
| AISession | Represents a conversation session between user and AI |
| AIRequest | Individual request within a session |
| RequestContext | Request metadata (ID, tenant, user, trace) |

### Value Objects

| Value Object | Type | Validation |
|---|---|---|
| `Prompt` | string | Max 32KB, sanitized |
| `ModelName` | string | Must exist in model registry |
| `RequestId` | UUID | Auto-generated |
| `SessionId` | UUID | Auto-generated per session |
| `TenantId` | i64 | Required, validated against JWT |

---

## 4. Public API Trait (Replaces gRPC)

```rust
// nexus-ai-gateway/src/interfaces/api.rs
#[async_trait]
pub trait AIGatewayService: Send + Sync {
    async fn submit_request(&self, req: AIRequest) -> Result<AIResponse, AIGatewayError>;
    async fn stream_response(&self, req: AIRequest) -> Result<Receiver<AIChunk>, AIGatewayError>;
    async fn get_session_status(&self, id: SessionId) -> Result<SessionStatus, AIGatewayError>;
    async fn cancel_request(&self, id: RequestId) -> Result<(), AIGatewayError>;
}

pub struct AIRequest {
    pub session_id: SessionId,
    pub prompt: Prompt,
    pub agent: AgentType,
    pub metadata: HashMap<String, String>,
    pub tenant_id: TenantId,
    pub user_id: UserId,
}

pub struct AIResponse {
    pub response: String,
    pub model: String,
    pub execution_id: ExecutionId,
    pub latency_ms: f64,
}

pub struct AIChunk {
    pub token: String,
    pub is_final: bool,
    pub chunk_type: ChunkType, // Token, ToolCall, Thinking, Completed
}
```

> **Note:** The optional external gRPC contract is defined in `proto/ai_gateway.proto` for partner/SDK integrations. It wraps the same trait methods via tonic-grpc.

---

## 5. REST API Endpoints

### Submit AI Chat Request

```
POST /api/v1/ai/chat
```

**Request:**
```json
{
  "message": "Explain my customer complaint",
  "agent": "customer-agent",
  "conversation_id": "uuid"
}
```

**Response:**
```json
{
  "conversation_id": "123",
  "agent": "customer-agent",
  "model": "command-r7b",
  "answer": "Customer has network issue..."
}
```

### Streaming AI Response (WebSocket)

**Connection:**
```
wss://api.aeroxenexus.com/ws/chat
```

**Message:**
```json
{
  "type": "message",
  "content": "Analyze my broadband issue"
}
```

**Streaming Response:**
```json
{ "type": "token", "content": "Customer" }
{ "type": "token", "content": " network" }
{ "type": "tool_call", "content": "customer.lookup()" }
{ "type": "completed" }
```

---

## 6. Request Processing Pipeline

```
External Request
    |
    v
[1] Request ID Generation
    |
    v
[2] Authentication (JWT Validation)
    |
    v
[3] Tenant Validation
    |
    v
[4] Rate Limiting (Token Bucket / Redis)
    |
    v
[5] Authorization (RBAC + ABAC)
    |
    v
[6] Sensitive Words Filter (prompt scan)
    |
    v
[7] Prompt Injection Detection
    |
    v
[8] Request Validation
    |
    v
[9] Agent Routing Decision
    |
    v
[10] gRPC Call to Agent Orchestrator
    |
    v
[11] Response Aggregation / Streaming
    |
    v
[12] Sensitive Words Filter (response scan)
    |
    v
[13] Audit Event Publishing
    |
    v
[14] Response to Client
```

---

## 7. Sensitive Words Filter

Every prompt and every response passes through a sensitive words filter before reaching the AI model or the customer.

### 7.1 Filter Architecture

```
User Prompt
  → [Prompt Filter] Scan for banned words
    → MATCH FOUND? → BLOCK + return safe message + audit log
    → NO MATCH → Continue to AI model

AI Response
  → [Response Filter] Scan for leaked sensitive data
    → MATCH FOUND? → BLOCK + redact + audit log
    → NO MATCH → Deliver to customer
```

### 7.2 Word Categories

| Category | Action | Examples |
|---|---|---|
| **Profanity** | BLOCK prompt | Abusive/vulgar words |
| **Hate Speech** | BLOCK prompt | Discriminatory language |
| **Violence** | BLOCK prompt | Threats, harm instructions |
| **Self-Harm** | BLOCK prompt | Suicide, self-injury content |
| **Sexual** | BLOCK prompt | Explicit sexual content |
| **Illegal Activities** | BLOCK prompt | Drug manufacturing, hacking instructions |
| **PII Leakage** | BLOCK response | Aadhaar numbers, PAN, credit cards |
| **Credential Leakage** | BLOCK response | Passwords, API keys, tokens |
| **Internal System Info** | BLOCK response | DB passwords, internal IPs, secret keys |
| **Competitor Names** | FLAG (log only) | Competitor product names |

### 7.3 Filter Configuration (Per Tenant)

```json
{
  "tenant_id": 1,
  "filter_config": {
    "enabled": true,
    "block profanity": true,
    "block hate_speech": true,
    "block violence": true,
    "block self_harm": true,
    "block sexual": true,
    "block illegal_activities": true,
    "block pii_leakage": true,
    "block credential_leakage": true,
    "block internal_info": true,
    "flag_competitor_names": true,
    "custom_blocked_words": ["word1", "word2"],
    "custom_allowed_words": ["allowed_word1"],
    "max_prompt_length": 10000,
    "log_all_matches": true
  }
}
```

### 7.4 Filter Implementation

```rust
// Rust implementation in nexus-ai-gateway::domain
use aho_corasick::AhoCorasick;

struct SensitiveWordFilter {
    profanity: AhoCorasick,
    hate_speech: AhoCorasick,
    violence: AhoCorasick,
    self_harm: AhoCorasick,
    sexual: AhoCorasick,
    illegal: AhoCorasick,
    pii_patterns: Vec<Regex>,
    credential_patterns: Vec<Regex>,
    internal_patterns: Vec<Regex>,
    custom_words: AhoCorasick,
}

impl SensitiveWordFilter {
    fn scan_prompt(&self, prompt: &str) -> FilterResult {
        let lower = prompt.to_lowercase();

        // Check all word categories
        if self.profanity.is_match(&lower) { return FilterResult::Blocked("profanity"); }
        if self.hate_speech.is_match(&lower) { return FilterResult::Blocked("hate_speech"); }
        if self.violence.is_match(&lower) { return FilterResult::Blocked("violence"); }
        if self.self_harm.is_match(&lower) { return FilterResult::Blocked("self_harm"); }
        if self.sexual.is_match(&lower) { return FilterResult::Blocked("sexual"); }
        if self.illegal.is_match(&lower) { return FilterResult::Blocked("illegal"); }
        if self.custom_words.is_match(&lower) { return FilterResult::Blocked("custom"); }

        FilterResult::Clean
    }

    fn scan_response(&self, response: &str) -> FilterResult {
        // Check PII patterns
        for pattern in &self.pii_patterns {
            if pattern.is_match(response) {
                return FilterResult::Redacted("pii_leakage");
            }
        }
        // Check credential patterns
        for pattern in &self.credential_patterns {
            if pattern.is_match(response) {
                return FilterResult::Redacted("credential_leakage");
            }
        }
        // Check internal info patterns
        for pattern in &self.internal_patterns {
            if pattern.is_match(response) {
                return FilterResult::Redacted("internal_info");
            }
        }
        FilterResult::Clean
    }
}
```

### 7.5 PII Detection Patterns

| Pattern | Regex | Action |
|---|---|---|
| Aadhaar Number | `\b\d{4}\s?\d{4}\s?\d{4}\b` | Redact: `XXXX XXXX 1234` |
| PAN Card | `\b[A-Z]{5}\d{4}[A-Z]\b` | Redact: `XXXXX1234X` |
| Credit Card | `\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b` | Redact: `XXXX XXXX XXXX 1234` |
| Phone Number (IN) | `\b(\+91|91|0)?[6-9]\d{9}\b` | Redact: `+91 XXXXX1234` |
| Email | `\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b` | Redact: `a***@gmail.com` |
| Passport | `\b[A-Z]\d{8}\b` | Redact: `X1234XXXX` |

### 7.6 Credential Detection Patterns

| Pattern | Regex | Action |
|---|---|---|
| API Key | `\b(sk|ak|pk|key|token)[_-]?[a-zA-Z0-9]{20,}\b` | Redact: `[REDACTED_API_KEY]` |
| Password in text | `password\s*[:=]\s*\S+` | Redact: `password: [REDACTED]` |
| JWT Token | `eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+` | Redact: `[REDACTED_JWT]` |
| Private Key | `-----BEGIN (RSA |EC )?PRIVATE KEY-----` | Redact: `[REDACTED_PRIVATE_KEY]` |
| AWS Key | `\b(AKIA|ABIA|ACCA|ASIA)[A-Z0-9]{16}\b` | Redact: `[REDACTED_AWS_KEY]` |

### 7.7 Filter Response Messages

| Block Reason | User Message |
|---|---|
| Profanity | "Your message contains inappropriate language. Please rephrase." |
| Hate Speech | "Your message contains discriminatory language. Please rephrase." |
| Violence | "Your message contains violent content. Please rephrase." |
| Self-Harm | "If you're in crisis, please contact a helpline. I'm here to help with platform-related questions." |
| Sexual | "Your message contains inappropriate content. Please rephrase." |
| Illegal | "I can't assist with illegal activities. I'm here to help with platform-related questions." |
| PII Leakage (response) | Response redacted, PII replaced with masked values |
| Credential Leakage (response) | Response redacted, credentials removed |

### 7.8 Audit Logging

Every filter match is logged:

```json
{
  "event_type": "sensitive_word_filter",
  "filter_type": "prompt",
  "category": "profanity",
  "matched_words": ["***"],
  "action": "blocked",
  "tenant_id": 1,
  "customer_id": 12345,
  "session_id": "sess_abc",
  "prompt_length": 150,
  "created_at": "2026-07-16T10:30:00Z"
}
```

---

## 8. Rate Limiting

### Configuration

| Tier | Limit | Window |
|---|---|---|
| Free | 10 AI requests/min | 60s |
| Customer | 100 AI requests/min | 60s |
| Enterprise | 10,000 AI requests/min | 60s |

### Implementation

- Algorithm: Token Bucket
- Storage: Redis
- Key: `rate_limit:{tenant_id}:{user_id}`
- Response on limit: `429 Too Many Requests` with `Retry-After` header

---

## 8. Session Management

### Session Lifecycle

```
User connects -> Create Session -> Set TTL (24h)
    |
    v
User sends message -> Attach to Session
    |
    v
User disconnects -> Keep Session (TTL active)
    |
    v
Session expires -> Cleanup
```

### Session Storage

| Store | Purpose |
|---|---|
| Redis | Active session context, conversation history (short-term) |
| PostgreSQL | Session metadata, history for long-term |

---

## 9. Agent Routing

The gateway determines which specialized agent handles each request.

### Routing Rules

| Intent | Target Agent | Model |
|---|---|---|
| Customer support | `customer-agent` | Phi-4-Mini:3.8B |
| Code generation | `developer-agent` | Qwen2.5-Coder:3B |
| Document Q&A | `rag-agent` | Command-R:7B |
| Image analysis | `vision-agent` | Qwen3-VL:4B |
| Security review | `security-agent` | WhiteRabbitNeo:7B |
| Business analysis | `business-agent` | Llama3.1:7B |
| General chat | `assistant-agent` | Phi-4-Mini:3.8B |

### Routing Decision Flow

```
User Message
    |
    v
LFM2.5 Thinking Model (Intent Detection)
    |
    v
Classify Intent -> Select Agent -> Execute
```

---

## 10. Multi-Tenancy

Every request must include `tenant_id`. The gateway:

1. Extracts `tenant_id` from JWT claims
2. Validates tenant exists and is active
3. Attaches `tenant_id` to all downstream gRPC calls
4. Enforces tenant-specific rate limits
5. Filters responses by tenant scope

---

## 11. Database Schema (gateway_db)

### ai_sessions

```sql
CREATE TABLE ai_sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB
);
```

### ai_requests

```sql
CREATE TABLE ai_requests (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES ai_sessions(id),
    tenant_id BIGINT NOT NULL,
    prompt TEXT NOT NULL,
    model VARCHAR(100),
    agent VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    latency_ms FLOAT,
    tokens_used INT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);
```

---

## 12. NATS Events (Async Communication)

> **Note:** Unlike the old microservice architecture, module-to-module calls use trait interfaces, not NATS. NATS is used only for async event publishing that other modules may consume for background processing.

### Published

| Subject | Event | Trigger |
|---|---|---|
| `aeroxe.ai.request.created` | `AIRequestReceived` | New request |
| `aeroxe.ai.response.generated` | `AIResponseGenerated` | Response ready |
| `aeroxe.ai.failed` | `AIRequestFailed` | Request error |

### Subscribed

| Subject | Handler |
|---|---|
| `aeroxe.agent.completed` | Update request status |
| `aeroxe.agent.failed` | Handle failure, retry if applicable |

---

## 13. Error Handling

### Error Types

| Code | Description |
|---|---|
| `AIGatewayError::Unknown` | Unexpected error |
| `AIGatewayError::InvalidRequest` | Malformed input |
| `AIGatewayError::ModelError` | Ollama inference failed |
| `AIGatewayError::Timeout` | Processing exceeded deadline |
| `AIGatewayError::RateLimited` | Rate limit exceeded |
| `AIGatewayError::FilterBlocked` | Content filter triggered |

### Standard Error Response

```json
{
  "error": {
    "code": "AI_MODEL_TIMEOUT",
    "message": "Model unavailable",
    "request_id": "uuid"
  }
}
```

---

## 14. Observability

### Metrics

| Metric | Type | Description |
|---|---|---|
| `ai_requests_total` | Counter | Total requests by agent/model |
| `ai_request_latency_ms` | Histogram | Request latency |
| `ai_tokens_generated` | Counter | Tokens produced |
| `ai_stream_connections` | Gauge | Active WebSocket connections |
| `ai_rate_limit_hits` | Counter | Rate limit rejections |

### Tracing

Every request gets a `trace_id` propagated through:
- REST header: `X-Trace-ID`
- gRPC metadata: `trace-id`
- NATS: `trace_id` in event data

---

## 15. Health Checks

```
GET /health/live    -> 200 OK (process alive)
GET /health/ready   -> 200 OK (dependencies available)
```

Readiness checks verify:
- PostgreSQL connection
- Redis connection
- NATS connection
- Agent Orchestrator availability
