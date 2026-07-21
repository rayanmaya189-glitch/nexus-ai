# AeroXe Nexus AI — API Specification

## REST + WebSocket + Streaming (Modular Monolith Interface) — Versioned

---

## 1. API Architecture

```
Client Apps
    |
    v
+-----------------------------------------+
|         gateway Module                  |
|     (axum HTTP/WS Server, port 8080)    |
|      Middleware: Auth → Tenant →        |
|      Rate-Limit → Version Check →       |
|      Schema Validation → Authz → Audit  |
+-----------------------------------------+
    |
    v
+-----------------------------------------+
|     API Version Router                  |
|     /api/v1/<resource>                  |
|     /api/v2/<resource>  (future)        |
+-----------------------------------------+
    |
    v
+-----------------------------------------+
|     Trait-based Module Dispatch          |
|     (Rust trait method calls, in-proc)  |
+-----------------------------------------+
    |
    v
All Modules (identity, customer, agent, ...)
```

The API Gateway does NOT proxy to separate services. It calls module trait methods directly — zero network overhead.

---

## 2. API Design Principles

| Principle | Implementation |
|---|---|
| Versioning | All routes prefixed with `/api/v{version}/` |
| Current Version | `v1` |
| Authentication | JWT Bearer token (`Authorization` header) |
| Tenant Isolation | `tenant_id` extracted from JWT, enforced by gateway |
| Rate Limiting | Token Bucket (Redis), per-tenant + per-endpoint + per-version |
| Request Tracing | `X-Request-ID` (auto-generated UUID v4) |
| Schema Validation | Versioned request schemas in `request_validator/schemas/` |
| Audit | All sensitive actions logged via `audit` trait call |
| Deprecation | Deprecated endpoints return `Sunset` header |
| Error Format | Standard JSON error envelope (see below) |
| CORS | Configurable per-tenant allowed origins |

---

## 3. Customer APIs (NEW)

Handled by `customer` module (`src/modules/customer/`).

### Create Customer

```
POST /api/v1/customers
Authorization: Bearer <jwt>
```

**Request:**
```json
{
  "name": "Acme Corp",
  "email": "contact@acme.com",
  "phone": "+1234567890",
  "addresses": [
    {
      "type": "billing",
      "line1": "123 Main St",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "US",
      "is_default": true
    }
  ],
  "tags": ["enterprise", "isp"],
  "custom_fields": {
    "industry": "telecom",
    "account_manager": "john.doe@aeroxe.com"
  }
}
```

**Response (201):**
```json
{
  "id": 1,
  "tenant_id": 1,
  "name": "Acme Corp",
  "email": "contact@acme.com",
  "phone": "+1234567890",
  "status": "active",
  "addresses": [
    {
      "id": 1,
      "type": "billing",
      "line1": "123 Main St",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "US",
      "is_default": true
    }
  ],
  "tags": ["enterprise", "isp"],
  "created_at": "2026-07-21T12:00:00Z"
}
```

### Get Customer

```
GET /api/v1/customers/{id}
Authorization: Bearer <jwt>
```

**Response (200):**
```json
{
  "id": 1,
  "tenant_id": 1,
  "name": "Acme Corp",
  "email": "contact@acme.com",
  "status": "active",
  "addresses": [...],
  "created_at": "2026-07-21T12:00:00Z"
}
```

### List Customers

```
GET /api/v1/customers?page=1&per_page=20&status=active
Authorization: Bearer <jwt>
```

### Suspend Customer

```
POST /api/v1/customers/{id}/suspend
Authorization: Bearer <jwt>
```

**Response (200):**
```json
{
  "id": 1,
  "status": "suspended",
  "suspended_at": "2026-07-21T12:00:00Z"
}
```

### Activate Customer

```
POST /api/v1/customers/{id}/activate
Authorization: Bearer <jwt>
```

### Update Customer

```
PUT /api/v1/customers/{id}
Authorization: Bearer <jwt>
```

### Delete Customer (Soft)

```
DELETE /api/v1/customers/{id}
Authorization: Bearer <jwt>
```

---

## 4. Authentication APIs

All handled by `identity` module via trait calls.

### Login

```
POST /api/v1/auth/login
```

**Request:**
```json
{
  "email": "admin@company.com",
  "password": "password"
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJl...",
  "expires_in": 3600,
  "user": {
    "id": 1,
    "email": "admin@company.com",
    "roles": ["admin"]
  }
}
```

### Refresh Token

```
POST /api/v1/auth/refresh
```

**Request:**
```json
{
  "refresh_token": "<token>"
}
```

### Register (Admin Only)

```
POST /api/v1/auth/register
```

### Get Current User

```
GET /api/v1/auth/me
Authorization: Bearer <jwt>
```

### Change Password

```
POST /api/v1/auth/change-password
```

---

## 5. AI Chat APIs

Handled by `ai-gateway` → `agent` → `rag` / `memory` etc. via trait calls.

### Submit Chat Request

```
POST /api/v1/ai/chat
Authorization: Bearer <jwt>
```

**Request:**
```json
{
  "message": "Explain my customer complaint",
  "agent": "customer-agent",
  "conversation_id": "uuid"
}
```

**Response (200):**
```json
{
  "conversation_id": "uuid",
  "agent": "customer-agent",
  "model": "command-r7b",
  "answer": "Based on your customer's account, they are experiencing a network issue...",
  "tokens_used": 450,
  "latency_ms": 3200
}
```

### Streaming Chat (WebSocket)

```
GET /api/v1/ai/stream
Upgrade: websocket
```

**Client sends:**
```json
{ "type": "message", "content": "Analyze my broadband issue", "conversation_id": "uuid" }
```

**Server streams:**
```json
{ "type": "token", "content": "Customer" }
{ "type": "token", "content": " network" }
{ "type": "tool_call", "content": "customer.lookup()" }
{ "type": "tool_result", "content": "{ ... }" }
{ "type": "completed", "usage": { "tokens": 450, "latency_ms": 3200 } }
```

---

## 6. Agent APIs

Handled by `agent` module.

### Execute Agent

```
POST /api/v1/agents/execute
```

**Request:**
```json
{
  "agent": "developer-agent",
  "task": "Review this Rust code for security issues",
  "context": {
    "code": "fn main() { println!(\"hello\"); }",
    "language": "rust"
  }
}
```

**Response (202):**
```json
{
  "execution_id": 12345,
  "status": "started"
}
```

### Get Execution Status

```
GET /api/v1/agents/executions/{id}
```

**Response:**
```json
{
  "id": 12345,
  "status": "completed",
  "agent": "developer-agent",
  "steps": [
    { "step": 1, "action": "analyze", "status": "done" },
    { "step": 2, "action": "review", "status": "done" }
  ],
  "result": "Code review complete. 2 issues found.",
  "tokens_used": 1250,
  "latency_ms": 4500
}
```

### Agent-Document Set Binding

```
POST /api/v1/agents/{id}/document-sets
DELETE /api/v1/agents/{id}/document-sets/{set_id}
GET  /api/v1/agents/{id}/document-sets
```

### Agent-Database Connection

```
POST /api/v1/agents/{id}/sql-connections/test
POST /api/v1/agents/{id}/sql-connections/discover
POST /api/v1/agents/{id}/sql-connections/tables
```

---

## 7. RAG Knowledge APIs

Handled by `rag` module.

### Upload Document

```
POST /api/v1/rag/documents
Content-Type: multipart/form-data
```

**Response (202):**
```json
{
  "document_id": 123,
  "status": "processing"
}
```

### Search Knowledge

```
POST /api/v1/rag/search
```

**Request:**
```json
{
  "query": "How to configure ONU?",
  "limit": 5
}
```

**Response:**
```json
{
  "results": [
    {
      "document_id": 456,
      "title": "ONU Configuration Guide",
      "score": 0.91,
      "content": "Step 1: Connect to ONU via...",
      "source": "network-guide.pdf",
      "metadata": { "department": "network" }
    }
  ],
  "latency_ms": 180
}
```

### Document Set Management

```
POST   /api/v1/document-sets                — Create set
GET    /api/v1/document-sets                 — List sets
GET    /api/v1/document-sets/{id}            — Get set details
PUT    /api/v1/document-sets/{id}            — Update set
DELETE /api/v1/document-sets/{id}            — Delete set
POST   /api/v1/document-sets/{id}/activate   — Activate
POST   /api/v1/document-sets/{id}/documents  — Add documents
DELETE /api/v1/document-sets/{id}/documents/{doc_id} — Remove
```

---

## 8. Vision APIs

Handled by `vision` module.

### Analyze Image

```
POST /api/v1/vision/analyze
Content-Type: multipart/form-data
```

**Response:**
```json
{
  "image_id": 789,
  "description": "Router LED is showing red — indicating PON signal loss",
  "confidence": 0.94,
  "analysis_type": "troubleshoot",
  "recommendations": ["Check fiber connection", "Restart device"],
  "latency_ms": 2800
}
```

### OCR

```
POST /api/v1/vision/ocr
Content-Type: multipart/form-data
```

### Batch Analysis

```
POST /api/v1/vision/batch
Content-Type: multipart/form-data
```

---

## 9. SQL Intelligence APIs

Handled by `sql-agent` module.

### Query (Generate + Execute)

```
POST /api/v1/sql/query
```

**Request:**
```json
{
  "question": "Show monthly revenue for 2026",
  "database_id": 1
}
```

**Response:**
```json
{
  "sql": "SELECT DATE_TRUNC('month', invoice_date) AS month, SUM(amount) AS total FROM invoices WHERE tenant_id = $1 AND invoice_date >= '2026-01-01' GROUP BY 1 ORDER BY 1",
  "explanation": "This query groups invoices by month...",
  "data": [
    { "month": "2026-01-01", "total": 500000 },
    { "month": "2026-02-01", "total": 620000 }
  ],
  "row_count": 7,
  "execution_time_ms": 45.2
}
```

### Generate SQL Only

```
POST /api/v1/sql/generate
```

---

## 10. Memory APIs

Handled by `memory` module.

### Store Memory

```
POST /api/v1/memory
```

**Request:**
```json
{
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

## 11. Workflow APIs

Handled by `workflow` module.

### Start Workflow

```
POST /api/v1/workflows/start
```

**Request:**
```json
{
  "workflow": "customer-support-flow",
  "context": { "ticket_id": "tkt_123", "customer_id": 456 }
}
```

**Response (202):**
```json
{
  "workflow_id": 789,
  "status": "running"
}
```

### Get Workflow Status

```
GET /api/v1/workflows/{id}
```

### Approve Step

```
POST /api/v1/workflows/{id}/steps/{step_id}/approve
```

---

## 12. Model Management APIs

Handled by `model-registry` module.

```
GET    /api/v1/models                    — List available models
GET    /api/v1/models/{name}             — Get model details
POST   /api/v1/models/pull               — Download a model
DELETE /api/v1/models/{name}             — Remove a model
GET    /api/v1/models/usage              — Usage statistics
```

---

## 13. KYC APIs

Handled by `identity` module.

```
GET    /api/v1/kyc/status                — Get KYC status
POST   /api/v1/kyc/documents             — Upload KYC document
GET    /api/v1/kyc/documents             — List submitted docs
POST   /api/v1/kyc/submit                — Submit for review
POST   /api/v1/kyc/review                — Admin: approve/reject
```

---

## 14. Security APIs

Handled by `security` module.

```
POST /api/v1/security/scan               — Security scan
POST /api/v1/security/review             — Code review
```

---

## 15. Telephony & Voice APIs (NEW)

Handled by `telephony`, `conversation`, `stt`, `tts` modules.

### Inbound Call Webhook

```
POST /api/v1/telephony/webhook/inbound
Content-Type: application/json
```

### Initiate Outbound Call

```
POST /api/v1/telephony/calls/outbound
Authorization: Bearer <jwt>
```

### Call Control

```
POST /api/v1/telephony/calls/{call_id}/hold
POST /api/v1/telephony/calls/{call_id}/resume
POST /api/v1/telephony/calls/{call_id}/transfer
POST /api/v1/telephony/calls/{call_id}/end
POST /api/v1/telephony/calls/{call_id}/recording/start
POST /api/v1/telephony/calls/{call_id}/recording/stop
```

### Caller Authentication (NEW)

```
POST /api/v1/telephony/calls/{call_id}/auth/verify-pin
POST /api/v1/telephony/calls/{call_id}/auth/verify-voice
GET  /api/v1/telephony/calls/{call_id}/auth/status
```

### Voicemail (NEW)

```
GET  /api/v1/telephony/voicemails
GET  /api/v1/telephony/voicemails/{id}
POST /api/v1/telephony/voicemails/{id}/listen
POST /api/v1/telephony/voicemails/{id}/handle
GET  /api/v1/telephony/voicemails/{id}/audio
GET  /api/v1/telephony/voicemails/{id}/transcript
```

### IVR Management (NEW)

```
POST   /api/v1/telephony/ivr-flows
GET    /api/v1/telephony/ivr-flows
GET    /api/v1/telephony/ivr-flows/{id}
PUT    /api/v1/telephony/ivr-flows/{id}
DELETE /api/v1/telephony/ivr-flows/{id}
```

### Live Monitoring (NEW)

```
POST /api/v1/telephony/calls/{call_id}/monitor/listen
POST /api/v1/telephony/calls/{call_id}/monitor/whisper
POST /api/v1/telephony/calls/{call_id}/monitor/barge-in
POST /api/v1/telephony/calls/{call_id}/monitor/stop
```

### Call Query

```
GET /api/v1/telephony/calls/{call_id}
GET /api/v1/telephony/calls?status=active
GET /api/v1/telephony/calls/{call_id}/transcript
```

### WebSocket Audio Stream

```
wss://host/ws/v1/telephony/{call_id}
```

### WebSocket Live Monitoring (NEW)

```
wss://host/ws/v1/telephony/monitor/{call_id}
```

---

## 16. Conversation APIs (NEW)

Handled by `conversation` module.

### Create Conversation

```
POST /api/v1/conversations
Authorization: Bearer <jwt>
```

**Request:**
```json
{
  "channel": "voice",
  "customer_id": 123,
  "agent_id": "support-agent",
  "initial_message": "Hello"
}
```

### Send Message

```
POST /api/v1/conversations/{id}/messages
```

### Get Conversation State

```
GET /api/v1/conversations/{id}/state
```

### End Conversation

```
POST /api/v1/conversations/{id}/end
```

---

## 17. STT/TTS APIs (NEW)

Handled by `stt` and `tts` modules.

### Start STT Session

```
POST /api/v1/stt/sessions
```

### Send Audio to STT

```
POST /api/v1/stt/sessions/{session_id}/audio
Content-Type: application/octet-stream
```

### Synthesize Speech (TTS)

```
POST /api/v1/tts/synthesize
```

### List TTS Voices

```
GET /api/v1/tts/voices?language=en
```

---

## 18. Outbound Campaign APIs (NEW)

Handled by `outbound` module.

### Create Campaign

```
POST /api/v1/outbound/campaigns
Authorization: Bearer <jwt>
```

### Start Campaign

```
POST /api/v1/outbound/campaigns/{id}/start
```

### Schedule Callback

```
POST /api/v1/outbound/callbacks
```

### Add to DNC List

```
POST /api/v1/outbound/dnc
```

---

## 19. Webhook APIs (NEW)

Handled by `webhook` module.

### Create Webhook

```
POST /api/v1/webhooks
Authorization: Bearer <jwt>
```

### Test Webhook

```
POST /api/v1/webhooks/{id}/test
```

### Get Delivery Logs

```
GET /api/v1/webhooks/{id}/deliveries
```

---

## 20. Analytics APIs (NEW)

Handled by `analytics` module.

### Get Dashboard

```
GET /api/v1/analytics/dashboard?start=2026-07-01&end=2026-07-21
Authorization: Bearer <jwt>
```

### Get Call Metrics

```
GET /api/v1/analytics/calls?start=2026-07-01&end=2026-07-21
```

### Get Agent Performance

```
GET /api/v1/analytics/agents/{agent_id}/performance
```

---

## 21. Health & Observability

```
GET /health    — Module health + dependency status
GET /metrics   — Prometheus metrics
```

### Health Response

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "api_versions": ["v1"],
  "uptime_seconds": 86400,
  "checks": {
    "postgresql": "healthy",
    "redis": "healthy",
    "nats": "healthy",
    "ollama": "healthy"
  }
}
```

---

## 16. Error Response Standard

All errors follow a consistent format with api_version:

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Retry after 30 seconds.",
    "request_id": "uuid",
    "api_version": "v1",
    "timestamp": "2026-07-21T12:00:00Z"
  }
}
```

### Error Codes

| Code | HTTP | Description |
|---|---|---|
| `UNAUTHORIZED` | 401 | Missing or invalid JWT |
| `TOKEN_EXPIRED` | 401 | JWT has expired |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `TENANT_VIOLATION` | 403 | Cross-tenant access |
| `NOT_FOUND` | 404 | Resource not found |
| `VERSION_NOT_FOUND` | 404 | API version does not exist |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit hit |
| `VERSION_DEPRECATED` | 400 | Deprecated API version used |
| `AI_MODEL_TIMEOUT` | 504 | Model inference timeout |
| `INTERNAL_ERROR` | 500 | Server error |

---

## 17. API Version Deprecation Flow

```
Step 1: New version v2 released
  → v1 enters deprecation period (6 months)
    → v1 endpoints return `Sunset: <date>` header
      → v1 returns `X-Deprecated: true` header
        → After deprecation period: v1 returns 404
          → All clients should migrate to v2
```

---

## 18. Performance Targets

| API | Target |
|---|---|
| Authentication | < 200ms |
| Customer CRUD | < 100ms |
| Chat First Token | < 2s |
| Chat Full Response | < 5s |
| RAG Search | < 500ms |
| Agent Start | < 300ms |
| Vision Request | < 5s |
| SQL Query | < 3s |
| Memory Search | < 200ms |
| Workflow Start | < 300ms |
| Gateway Middleware | < 5ms per layer |
