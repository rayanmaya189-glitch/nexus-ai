# AeroXe Nexus AI — API Specification

## REST, WebSocket, gRPC Gateway & Streaming

---

## 1. API Architecture

```
Client Apps
    |
    v
Nexus API Gateway
    |
  ================================================
  REST APIs | WebSocket Streaming | gRPC Gateway
  ================================================
    |
    v
Internal Services (gRPC + NATS JetStream)
```

---

## 2. API Design Principles

All APIs must support:
- Versioning (`/api/v1/`)
- Authentication (JWT)
- Tenant isolation (`tenant_id` in every request)
- Rate limiting
- Request tracing (`X-Request-ID`, `X-Trace-ID`)
- Audit logging

---

## 3. Authentication APIs

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

**Response:**
```json
{
  "access_token": "jwt-token",
  "refresh_token": "refresh-token",
  "expires_in": 3600
}
```

### Refresh Token

```
POST /api/v1/auth/refresh
```

---

## 4. AI Chat API

### Submit Chat Request

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

### Streaming (WebSocket)

**Connection:** `wss://api.aeroxenexus.com/ws/chat`

**Message:**
```json
{ "type": "message", "content": "Analyze my broadband issue" }
```

**Streaming Response:**
```json
{ "type": "token", "content": "Customer" }
{ "type": "token", "content": " network" }
{ "type": "tool_call", "content": "customer.lookup()" }
{ "type": "tool_result", "content": "{ ... }" }
{ "type": "completed" }
```

---

## 5. Agent APIs

### Execute Agent

```
POST /api/v1/agents/execute
```

**Request:**
```json
{
  "agent": "developer-agent",
  "task": "Review this Rust code",
  "context": { "repository": "backend" }
}
```

**Response:**
```json
{
  "execution_id": "abc123",
  "status": "started"
}
```

### Get Execution Status

```
GET /api/v1/agents/execution/{id}
```

**Response:**
```json
{
  "id": "abc123",
  "status": "completed",
  "steps": [
    { "step": "analyze", "status": "done" },
    { "step": "review", "status": "done" }
  ],
  "result": "Code review complete. 3 issues found."
}
```

---

## 6. RAG Knowledge APIs

### Upload Document

```
POST /api/v1/rag/documents
Content-Type: multipart/form-data
```

**Response:**
```json
{
  "document_id": "uuid",
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
      "title": "ONU Guide",
      "score": 0.91,
      "content": "Configuration steps...",
      "source": "network-guide.pdf"
    }
  ]
}
```

### Get Document Status

```
GET /api/v1/rag/documents/{id}/status
```

---

## 7. Vision APIs

### Analyze Image

```
POST /api/v1/vision/analyze
Content-Type: multipart/form-data
```

**Response:**
```json
{
  "description": "Router LED is showing red",
  "confidence": 0.94
}
```

### OCR

```
POST /api/v1/vision/ocr
Content-Type: multipart/form-data
```

---

## 8. SQL Intelligence APIs

### Query

```
POST /api/v1/sql/query
```

**Request:**
```json
{
  "question": "Show monthly revenue",
  "database": "aeroxe_billing_db"
}
```

**Response:**
```json
{
  "sql": "SELECT SUM(amount)...",
  "data": [...],
  "row_count": 7
}
```

---

## 9. Memory APIs

### Store Memory

```
POST /api/v1/memory
```

**Request:**
```json
{
  "user_id": "123",
  "memory": "Customer prefers Hindi support",
  "type": "preference"
}
```

### Search Memory

```
GET /api/v1/memory/search?q=customer&limit=5
```

---

## 10. Workflow APIs

### Start Workflow

```
POST /api/v1/workflows/start
```

**Request:**
```json
{
  "workflow": "customer-support-flow",
  "context": { "ticket_id": "tkt_123" }
}
```

**Response:**
```json
{
  "workflow_id": "123",
  "status": "running"
}
```

### Get Workflow Status

```
GET /api/v1/workflows/{id}
```

---

## 11. Model Management APIs

### List Models

```
GET /api/v1/models
```

**Response:**
```json
[
  {
    "name": "qwen3-vl:4b",
    "type": "vision",
    "status": "available"
  }
]
```

---

## 12. API Error Standard

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

## 13. API Gateway Middleware Pipeline

Every request passes through:

```
Request ID -> Authentication -> Tenant Validation
    -> Rate Limit -> Authorization -> Logging -> Routing
```

---

## 14. API Performance Requirements

| API | Target |
|---|---|
| Authentication | < 200ms |
| Chat First Token | < 2s |
| RAG Search | < 500ms |
| Agent Start | < 300ms |
| Vision Request | < 5s |
| SQL Query | < 3s |
| Memory Search | < 200ms |
| Workflow Start | < 300ms |
