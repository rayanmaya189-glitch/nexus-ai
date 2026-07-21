# AeroXe Nexus AI — API Specification

## REST + WebSocket + Streaming (Modular Monolith Interface)

---

## 1. API Architecture

```
Client Apps
    |
    v
+-----------------------------------------+
|         nexus-gateway Module            |
|     (axum HTTP/WS Server, port 8080)    |
|      Middleware: Auth → Tenant →        |
|      Rate-Limit → Authz → Audit         |
+-----------------------------------------+
    |
    v
+-----------------------------------------+
|     Trait-based Module Dispatch          |
|     (Rust trait method calls, in-proc)  |
+-----------------------------------------+
    |
    v
All Modules (nexus-identity, nexus-agent, ...)
```

The API Gateway does NOT proxy to separate services. It calls module trait methods directly — zero network overhead.

---

## 2. API Design Principles

| Principle | Implementation |
|---|---|
| Versioning | All routes prefixed with `/api/v1/` |
| Authentication | JWT Bearer token (`Authorization` header) |
| Tenant Isolation | `tenant_id` extracted from JWT, enforced by gateway |
| Rate Limiting | Token Bucket (Redis), per-tenant + per-endpoint |
| Request Tracing | `X-Request-ID` (auto-generated UUID v4) |
| Audit | All sensitive actions logged via `nexus-audit` trait call |
| Error Format | Standard JSON error envelope (see below) |
| CORS | Configurable per-tenant allowed origins |

---

## 3. Authentication APIs

All handled by `nexus-identity` module via trait calls.

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

## 4. AI Chat APIs

Handled by `nexus-ai-gateway` → `nexus-agent` → `nexus-rag` / `nexus-memory` etc. via trait calls.

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

## 5. Agent APIs

Handled by `nexus-agent` module.

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

## 6. RAG Knowledge APIs

Handled by `nexus-rag` module.

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

## 7. Vision APIs

Handled by `nexus-vision` module.

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

## 8. SQL Intelligence APIs

Handled by `nexus-sql-agent` module.

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

## 9. Memory APIs

Handled by `nexus-memory` module.

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

## 10. Workflow APIs

Handled by `nexus-workflow` module.

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

## 11. Model Management APIs

Handled by `nexus-model-registry` module.

```
GET    /api/v1/models                    — List available models
GET    /api/v1/models/{name}             — Get model details
POST   /api/v1/models/pull               — Download a model
DELETE /api/v1/models/{name}             — Remove a model
GET    /api/v1/models/usage              — Usage statistics
```

---

## 12. KYC APIs

Handled by `nexus-identity` module.

```
GET    /api/v1/kyc/status                — Get KYC status
POST   /api/v1/kyc/documents             — Upload KYC document
GET    /api/v1/kyc/documents             — List submitted docs
POST   /api/v1/kyc/submit                — Submit for review
POST   /api/v1/kyc/review                — Admin: approve/reject
```

---

## 13. Security APIs

Handled by `nexus-security-ai` module.

```
POST /api/v1/security/scan               — Security scan
POST /api/v1/security/review             — Code review
```

---

## 14. Health & Observability

```
GET /health    — Module health + dependency status
GET /metrics   — Prometheus metrics
```

### Health Response

```json
{
  "status": "healthy",
  "version": "1.0.0",
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

## 15. Error Response Standard

All errors follow a consistent format:

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Retry after 30 seconds.",
    "request_id": "uuid",
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
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit hit |
| `AI_MODEL_TIMEOUT` | 504 | Model inference timeout |
| `INTERNAL_ERROR` | 500 | Server error |

---

## 16. Performance Targets

| API | Target |
|---|---|
| Authentication | < 200ms |
| Chat First Token | < 2s |
| Chat Full Response | < 5s |
| RAG Search | < 500ms |
| Agent Start | < 300ms |
| Vision Request | < 5s |
| SQL Query | < 3s |
| Memory Search | < 200ms |
| Workflow Start | < 300ms |
| Gateway Middleware | < 5ms per layer |
