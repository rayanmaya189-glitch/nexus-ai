# AeroXe Nexus AI — AI Gateway Service

## Central AI Request Processing, Routing & Management

---

## 1. Service Identity

| Attribute | Value |
|---|---|
| Service Name | `ai-gateway-service` |
| Bounded Context | AI Gateway |
| Domain Type | Core Domain |
| Language | Go |
| Database | `gateway_db` (PostgreSQL) |
| gRPC Port | 50051 |
| REST Port | 8080 |

---

## 2. Purpose

The AI Gateway is the single entry point for all AI requests in the AeroXe Nexus AI platform. It handles:

- Request reception and validation
- Authentication and authorization enforcement
- Rate limiting per tenant
- Request routing to appropriate AI agents
- Response aggregation and streaming
- Session management
- Audit logging

---

## 3. Aggregate Design

### AIRequest Aggregate

```
AIRequest (Aggregate Root)
├── RequestContext
│   ├── RequestId (UUID)
│   ├── TenantId (UUID)
│   ├── UserId (UUID)
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
| `TenantId` | UUID | Required, validated against JWT |

---

## 4. gRPC Contract

```protobuf
syntax = "proto3";
package aeroxe.ai_gateway;

service AIGatewayService {
  rpc SubmitRequest(AIRequest) returns (AIResponse);
  rpc StreamResponse(AIRequest) returns (stream AIChunk);
  rpc GetSessionStatus(SessionRequest) returns (SessionStatus);
  rpc CancelRequest(CancelRequest) returns (CancelResponse);
}

message AIRequest {
  string session_id = 1;
  string prompt = 2;
  string agent = 3;
  map<string, string> metadata = 4;
}

message AIResponse {
  string response = 1;
  string model = 2;
  string execution_id = 3;
  float latency_ms = 4;
}

message AIChunk {
  string token = 1;
  bool is_final = 2;
  string chunk_type = 3; // "token", "tool_call", "thinking", "completed"
}
```

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
[6] Input Sanitization (Prompt Injection Check)
    |
    v
[7] Request Validation
    |
    v
[8] Agent Routing Decision
    |
    v
[9] gRPC Call to Agent Orchestrator
    |
    v
[10] Response Aggregation / Streaming
    |
    v
[11] Audit Event Publishing
    |
    v
[12] Response to Client
```

---

## 7. Rate Limiting

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
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB
);
```

### ai_requests

```sql
CREATE TABLE ai_requests (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES ai_sessions(id),
    tenant_id UUID NOT NULL,
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

## 12. NATS Events

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

### gRPC Error Codes

| Code | Name | HTTP Equivalent |
|---|---|---|
| 0 | UNKNOWN | 500 |
| 1 | INVALID_REQUEST | 400 |
| 2 | UNAUTHORIZED | 401 |
| 3 | FORBIDDEN | 403 |
| 4 | NOT_FOUND | 404 |
| 5 | TIMEOUT | 408 |
| 6 | MODEL_ERROR | 502 |
| 7 | DATABASE_ERROR | 503 |

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
