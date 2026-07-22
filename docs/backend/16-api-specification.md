# AeroXe Nexus AI — API Specification

## REST + Protobuf — Microservices in Modular Monolith

---

## 1. API Design Principles

| Principle | Implementation |
|---|---|
| Protocol | REST + Protobuf (proto3) serialized as JSON over HTTP |
| HTTP Methods | PATCH (partial update), POST (create + read + actions), DELETE (remove) |
| Serialization | All request/response bodies are Protobuf messages encoded as JSON |
| Path Variables | Resource IDs in URL path |
| Versioning | `/api/v{version}/` prefix |
| Authentication | JWT Bearer (`Authorization` header) |
| Business Status | Every response includes `status` field |
| Pagination | Included in Protobuf response message (`Pagination` field) |

---

## 2. HTTP Methods

| Method | Usage | Example |
|---|---|---|
| `POST` | Read, create, or trigger action | `POST /api/v1/customers` (body specifies operation) |
| `PATCH` | Partial update of resource | `PATCH /api/v1/customers/{id}` |
| `DELETE` | Remove resource | `DELETE /api/v1/customers/{id}` |

> **Note:** No `GET` or `PUT` methods are used. Read operations use `POST` with an operation-specific Protobuf request body.

---

## 3. Request/Response Format

### 3.1 Request Envelope

Every request includes a Protobuf message (serialized as JSON) with structured envelopes:

```json
{
  "operation": "GetCustomer",
  "request_id": "uuid",
  "tenant_id": 1,
  "user_id": 123,
  "data": "<base64-encoded Protobuf request>"
}
```

> All request bodies are Protobuf messages. The `data` field contains the serialized operation-specific Protobuf request message.

### 3.2 Success Response Envelope

```json
{
  "status": "SUCCESS",
  "operation": "GetCustomer",
  "request_id": "uuid",
  "data": {
    "id": 456,
    "name": "Acme Corp",
    "status": "active"
  },
  "meta": {
    "timestamp": "2026-07-21T12:00:00Z"
  }
}
```

### 3.3 List Response Envelope

```json
{
  "status": "SUCCESS",
  "operation": "ListCustomers",
  "request_id": "uuid",
  "data": "<base64-encoded Protobuf repeated message>",
  "summary": {
    "total_items": 1234,
    "active_items": 980,
    "inactive_items": 254
  },
  "pagination": {
    "total": 1234,
    "limit": 10,
    "offset": 0,
    "has_more": true
  },
  "meta": {
    "timestamp": "2026-07-21T12:00:00Z"
  }
}
```

### 3.4 Error Response Envelope

All error responses are also Protobuf messages (`ErrorInfo` + `ErrorDetail`):

```json
{
  "status": "ERROR",
  "code": "VALIDATION_ERROR",
  "message": "Email is required",
  "details": [
    {"field": "email", "message": "is required"}
  ],
  "request_id": "uuid"
}
```

---

## 4. Error Response Standard

All error responses follow a consistent Protobuf-based format:

```json
{
  "status": "ERROR",
  "code": "ERROR_CODE",
  "message": "Human-readable error description",
  "details": [],
  "request_id": "uuid"
}
```

| Field | Type | Description |
|---|---|---|
| `status` | string | Always `"ERROR"` for error responses |
| `code` | string | Machine-readable error code (e.g., `VALIDATION_ERROR`, `NOT_FOUND`) |
| `message` | string | Human-readable error description |
| `details` | array | Additional error context (field-level errors, validation details) |
| `request_id` | string | UUID correlating request to server-side trace |

---

## 5. Business Status Codes

| Status | Description |
|---|---|
| `SUCCESS` | Operation completed successfully |
| `CREATED` | Resource created successfully |
| `UPDATED` | Resource updated successfully |
| `DELETED` | Resource deleted successfully |
| `VALIDATION_ERROR` | Request body validation failed |
| `UNAUTHORIZED` | Missing or invalid JWT |
| `TOKEN_EXPIRED` | JWT has expired |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource does not exist |
| `CONFLICT` | Resource already exists |
| `UNPROCESSABLE_ENTITY` | Business rule violation |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `INTERNAL_ERROR` | Unexpected server error |
| `SERVICE_UNAVAILABLE` | Service temporarily down |

---

## 6. API Endpoint Structure

| Pattern | Method | Example |
|---|---|---|
| Read resource(s) | `POST /api/v1/{resource}` | Body: `{"operation": "ListCustomers", "limit": 10}` |
| Read single resource | `POST /api/v1/{resource}` | Body: `{"operation": "GetCustomer", "id": 123}` |
| Create resource | `POST /api/v1/{resource}` | Body: `{"operation": "CreateCustomer", ...}` |
| Update resource (partial) | `PATCH /api/v1/{resource}/{id}` | Body: partial update fields |
| Delete resource | `DELETE /api/v1/{resource}/{id}` | — |
| Trigger action | `POST /api/v1/{resource}/{action}` | `POST /api/v1/customers/123/suspend` |

> **Note:** No `GET` or `PUT` methods. Read operations use `POST` with a Protobuf request body specifying the operation.

---

## 7. Authentication APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/auth/login` | Login | `SUCCESS` |
| `POST /api/v1/auth/refresh` | Refresh Token | `SUCCESS` |
| `POST /api/v1/auth/register` | Register | `CREATED` |
| `POST /api/v1/auth/me` | Get Current User | `SUCCESS` |
| `POST /api/v1/auth/change-password` | Change Password | `SUCCESS` |

---

## 8. Customer APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/customers` | Create Customer | `CREATED` |
| `POST /api/v1/customers` | Get Customer | `SUCCESS` |
| `POST /api/v1/customers` | List Customers | `SUCCESS` |
| `PATCH /api/v1/customers/{id}` | Update Customer | `UPDATED` |
| `DELETE /api/v1/customers/{id}` | Delete Customer | `DELETED` |
| `POST /api/v1/customers/{id}/suspend` | Suspend Customer | `UPDATED` |
| `POST /api/v1/customers/{id}/activate` | Activate Customer | `UPDATED` |

---

## 9. AI Chat APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/ai/chat` | Submit Chat Request | `SUCCESS` |
| `WS /ws/v1/chat` | Streaming Chat | — |

---

## 10. Agent APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/agents` | Execute Agent | `CREATED` |
| `POST /api/v1/agents/{id}/execution` | Get Execution Status | `SUCCESS` |
| `POST /api/v1/agents/{id}/executions` | List Executions | `SUCCESS` |
| `POST /api/v1/agents/{id}/document-sets` | Bind Document Sets | `UPDATED` |
| `DELETE /api/v1/agents/{id}/document-sets/{set_id}` | Unbind Document Sets | `UPDATED` |
| `POST /api/v1/agents/{id}/document-sets` | List Bound Document Sets | `SUCCESS` |
| `POST /api/v1/agents/sql/test-connection` | Test SQL Connection | `SUCCESS` |
| `POST /api/v1/agents/sql/discover-schema` | Discover Database Schema | `SUCCESS` |
| `POST /api/v1/agents/{id}/tables` | Bind Tables | `UPDATED` |

---

## 11. RAG APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/rag/documents` | Upload Document | `CREATED` |
| `POST /api/v1/rag/documents/{id}/status` | Get Document Status | `SUCCESS` |
| `POST /api/v1/rag/documents` | List Documents | `SUCCESS` |
| `DELETE /api/v1/rag/documents/{id}` | Delete Document | `DELETED` |
| `POST /api/v1/rag/search` | Search Knowledge | `SUCCESS` |
| `POST /api/v1/rag/document-sets` | Create Document Set | `CREATED` |
| `POST /api/v1/rag/document-sets` | List Document Sets | `SUCCESS` |
| `PATCH /api/v1/rag/document-sets/{id}` | Update Document Set | `UPDATED` |
| `DELETE /api/v1/rag/document-sets/{id}` | Delete Document Set | `DELETED` |

---

## 12. Vision APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/vision/analyze` | Analyze Image | `SUCCESS` |
| `POST /api/v1/vision/ocr` | Extract Text | `SUCCESS` |
| `POST /api/v1/vision/batch` | Batch Analysis | `SUCCESS` |

---

## 13. SQL Intelligence APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/sql/query` | Execute SQL Query | `SUCCESS` |
| `POST /api/v1/sql/generate` | Generate SQL | `SUCCESS` |

---

## 14. Memory APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/memory` | Store Memory | `CREATED` |
| `POST /api/v1/memory/search` | Search Memory | `SUCCESS` |
| `POST /api/v1/memory/context` | Get Conversation Context | `SUCCESS` |
| `DELETE /api/v1/memory/{id}` | Delete Memory | `DELETED` |

---

## 15. Workflow APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/workflows` | Start Workflow | `CREATED` |
| `POST /api/v1/workflows/{id}` | Get Workflow | `SUCCESS` |
| `POST /api/v1/workflows` | List Workflows | `SUCCESS` |
| `POST /api/v1/workflows/{id}/approve-step` | Approve Step | `UPDATED` |
| `POST /api/v1/workflows/{id}/cancel` | Cancel Workflow | `UPDATED` |

---

## 16. Model Management APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/models` | List Models | `SUCCESS` |
| `POST /api/v1/models/{id}` | Get Model | `SUCCESS` |
| `POST /api/v1/models/pull` | Pull Model | `CREATED` |
| `DELETE /api/v1/models/{id}` | Delete Model | `DELETED` |
| `POST /api/v1/models/usage` | Get Usage Stats | `SUCCESS` |

---

## 17. KYC APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/kyc/status` | Get KYC Status | `SUCCESS` |
| `POST /api/v1/kyc/documents` | Upload KYC Document | `CREATED` |
| `POST /api/v1/kyc/documents` | List KYC Documents | `SUCCESS` |
| `DELETE /api/v1/kyc/documents/{id}` | Delete KYC Document | `DELETED` |
| `POST /api/v1/kyc/submit` | Submit for Review | `UPDATED` |
| `POST /api/v1/kyc/review` | Review KYC | `UPDATED` |

---

## 18. Security APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/security/scan` | Security Scan | `CREATED` |
| `POST /api/v1/security/review` | Code Review | `SUCCESS` |

---

## 19. Telephony APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/telephony/webhook/inbound` | Inbound Call Webhook | `SUCCESS` |
| `POST /api/v1/telephony/calls` | Initiate Outbound Call | `CREATED` |
| `POST /api/v1/telephony/calls/{id}` | Get Call Details | `SUCCESS` |
| `POST /api/v1/telephony/calls` | List Calls | `SUCCESS` |
| `POST /api/v1/telephony/calls/{id}/hold` | Hold Call | `UPDATED` |
| `POST /api/v1/telephony/calls/{id}/resume` | Resume Call | `UPDATED` |
| `POST /api/v1/telephony/calls/{id}/transfer` | Transfer Call | `UPDATED` |
| `POST /api/v1/telephony/calls/{id}/end` | End Call | `UPDATED` |
| `POST /api/v1/telephony/calls/{id}/start-recording` | Start Recording | `UPDATED` |
| `POST /api/v1/telephony/calls/{id}/stop-recording` | Stop Recording | `UPDATED` |
| `POST /api/v1/telephony/calls/{id}/transcript` | Get Transcript | `SUCCESS` |
| `POST /api/v1/telephony/calls/{id}/verify-pin` | Verify PIN | `SUCCESS` |
| `POST /api/v1/telephony/calls/{id}/verify-voice` | Verify Voice | `SUCCESS` |
| `POST /api/v1/telephony/calls/{id}/auth-status` | Get Auth Status | `SUCCESS` |
| `POST /api/v1/telephony/voicemails` | List Voicemails | `SUCCESS` |
| `POST /api/v1/telephony/voicemails/{id}` | Get Voicemail | `SUCCESS` |
| `POST /api/v1/telephony/voicemails/{id}/handle` | Handle Voicemail | `UPDATED` |
| `POST /api/v1/telephony/ivr` | Create IVR Flow | `CREATED` |
| `POST /api/v1/telephony/ivr` | List IVR Flows | `SUCCESS` |
| `PATCH /api/v1/telephony/ivr/{id}` | Update IVR Flow | `UPDATED` |
| `DELETE /api/v1/telephony/ivr/{id}` | Delete IVR Flow | `DELETED` |
| `POST /api/v1/telephony/monitor/listen` | Listen-In | `SUCCESS` |
| `POST /api/v1/telephony/monitor/whisper` | Whisper to Agent | `SUCCESS` |
| `POST /api/v1/telephony/monitor/barge-in` | Barge-In | `SUCCESS` |
| `POST /api/v1/telephony/monitor/stop` | Stop Monitoring | `UPDATED` |
| `WS /ws/v1/telephony` | Audio Stream (Protobuf framing) | — |
| `WS /ws/v1/telephony/monitor` | Live Monitoring (Protobuf framing) | — |

---

## 20. Conversation APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/conversations` | Create Conversation | `CREATED` |
| `POST /api/v1/conversations/{id}` | Get Conversation | `SUCCESS` |
| `POST /api/v1/conversations` | List Conversations | `SUCCESS` |
| `POST /api/v1/conversations/{id}/messages` | Get Messages | `SUCCESS` |
| `POST /api/v1/conversations/{id}/messages` | Add Message | `CREATED` |
| `POST /api/v1/conversations/{id}/state` | Get State | `SUCCESS` |
| `POST /api/v1/conversations/{id}/end` | End Conversation | `UPDATED` |
| `POST /api/v1/conversations/{id}/branch` | Branch Conversation | `CREATED` |
| `DELETE /api/v1/conversations/{id}` | Delete Conversation | `DELETED` |

---

## 21. STT APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/stt/sessions` | Start Session | `CREATED` |
| `POST /api/v1/stt/sessions/{id}/audio` | Send Audio Chunk | `SUCCESS` |
| `POST /api/v1/stt/sessions/{id}/end` | End Session | `UPDATED` |
| `POST /api/v1/stt/transcribe` | Batch Transcribe | `SUCCESS` |
| `POST /api/v1/stt/sessions/{id}` | Get Session | `SUCCESS` |

---

## 22. TTS APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/tts/synthesize` | Synthesize Speech | `SUCCESS` |
| `POST /api/v1/tts/synthesize-ssml` | Synthesize SSML | `SUCCESS` |
| `POST /api/v1/tts/voices` | List Voices | `SUCCESS` |
| `POST /api/v1/tts/voices/{id}` | Get Voice | `SUCCESS` |
| `POST /api/v1/tts/voices/{id}/preview` | Preview Voice | `SUCCESS` |
| `POST /api/v1/tts/voices/clone` | Clone Voice | `CREATED` |
| `DELETE /api/v1/tts/voices/clone/{id}` | Revoke Clone | `DELETED` |

---

## 23. Outbound APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/outbound/campaigns` | Create Campaign | `CREATED` |
| `POST /api/v1/outbound/campaigns/{id}` | Get Campaign | `SUCCESS` |
| `POST /api/v1/outbound/campaigns` | List Campaigns | `SUCCESS` |
| `POST /api/v1/outbound/campaigns/{id}/start` | Start Campaign | `UPDATED` |
| `POST /api/v1/outbound/campaigns/{id}/pause` | Pause Campaign | `UPDATED` |
| `POST /api/v1/outbound/campaigns/{id}/cancel` | Cancel Campaign | `UPDATED` |
| `POST /api/v1/outbound/campaigns/{id}/stats` | Get Campaign Stats | `SUCCESS` |
| `POST /api/v1/outbound/callbacks` | Schedule Callback | `CREATED` |
| `POST /api/v1/outbound/callbacks` | List Callbacks | `SUCCESS` |
| `DELETE /api/v1/outbound/callbacks/{id}` | Cancel Callback | `DELETED` |
| `POST /api/v1/outbound/dnc` | Add to DNC | `CREATED` |
| `POST /api/v1/outbound/dnc` | List DNC | `SUCCESS` |
| `DELETE /api/v1/outbound/dnc/{id}` | Remove from DNC | `DELETED` |

---

## 24. Webhook APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/webhooks` | Create Webhook | `CREATED` |
| `POST /api/v1/webhooks/{id}` | Get Webhook | `SUCCESS` |
| `POST /api/v1/webhooks` | List Webhooks | `SUCCESS` |
| `PATCH /api/v1/webhooks/{id}` | Update Webhook | `UPDATED` |
| `DELETE /api/v1/webhooks/{id}` | Delete Webhook | `DELETED` |
| `POST /api/v1/webhooks/{id}/test` | Test Webhook | `SUCCESS` |
| `POST /api/v1/webhooks/{id}/deliveries` | List Deliveries | `SUCCESS` |
| `POST /api/v1/webhooks/deliveries/{id}/retry` | Retry Delivery | `SUCCESS` |

---

## 25. Analytics APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/analytics/dashboard` | Get Dashboard | `SUCCESS` |
| `POST /api/v1/analytics/realtime` | Get Real-time Metrics | `SUCCESS` |
| `POST /api/v1/analytics/conversations` | Get Conversation Metrics | `SUCCESS` |
| `POST /api/v1/analytics/calls` | Get Call Metrics | `SUCCESS` |
| `POST /api/v1/analytics/agents` | List Agent Metrics | `SUCCESS` |
| `POST /api/v1/analytics/agents/{id}/performance` | Get Agent Performance | `SUCCESS` |
| `POST /api/v1/analytics/costs` | Get Cost Breakdown | `SUCCESS` |
| `POST /api/v1/analytics/customers/{id}/cost` | Get Customer Cost | `SUCCESS` |
| `POST /api/v1/analytics/reports` | Create Report | `CREATED` |
| `POST /api/v1/analytics/reports` | List Reports | `SUCCESS` |
| `POST /api/v1/analytics/reports/{id}` | Get Report | `SUCCESS` |

---

## 26. Billing APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/billing/plans` | List Plans | `SUCCESS` |
| `POST /api/v1/billing/subscriptions` | Create Subscription | `CREATED` |
| `POST /api/v1/billing/subscriptions/{id}` | Get Subscription | `SUCCESS` |
| `PATCH /api/v1/billing/subscriptions/{id}` | Update Subscription | `UPDATED` |
| `POST /api/v1/billing/subscriptions/{id}/cancel` | Cancel Subscription | `UPDATED` |
| `POST /api/v1/billing/usage` | Get Usage | `SUCCESS` |
| `POST /api/v1/billing/invoices` | List Invoices | `SUCCESS` |
| `POST /api/v1/billing/invoices/{id}` | Get Invoice | `SUCCESS` |
| `POST /api/v1/billing/invoices/{id}/pay` | Pay Invoice | `SUCCESS` |
| `POST /api/v1/billing/payments` | List Payments | `SUCCESS` |

---

## 27. Health & Observability

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/health/check` | Health Check | `SUCCESS` |
| `POST /api/v1/metrics` | Get Metrics | `SUCCESS` |

---

## 28. WebSocket Protocol

> **Note:** WebSocket messages use Protobuf framing. The JSON examples below illustrate the logical message structure; actual wire format is Protobuf-serialized binary.

### Connection

```
ws://host/ws/v1/{channel}?token={jwt}
```

Channels: `chat`, `telephony`, `telephony/monitor`

### Message Format

All WebSocket messages use a standard Protobuf envelope:

```json
{
  "type": "message_type",
  "data": {},
  "timestamp": "2026-07-21T12:00:00Z"
}
```

| Field | Type | Description |
|---|---|---|
| `type` | string | Message type (`message`, `ping`, `pong`, `error`, `done`) |
| `data` | object | Message payload (type-specific) |
| `timestamp` | string | ISO 8601 timestamp |

### Ping/Pong Keep-Alive

Server sends `ping` every **30 seconds**. Client must respond with `pong` within **10 seconds**. If no `pong` is received, the connection is closed and the client should reconnect.

```json
// Server → Client
{"type": "ping", "data": {}, "timestamp": "..."}

// Client → Server
{"type": "pong", "data": {}, "timestamp": "..."}
```

### Auto-Reconnect with Exponential Backoff

Clients must implement automatic reconnection:

| Attempt | Delay | Jitter |
|---|---|---|
| 1 | 1s | ±250ms |
| 2 | 2s | ±500ms |
| 3 | 4s | ±1s |
| 4 | 8s | ±2s |
| 5+ | 16s (cap) | ±2s |

After 5 failed attempts, the client should enter a degraded state and notify the user. Reconnection must resume the previous session using the stored `conversation_id` or `session_id`.

### Streaming Chat Message Flow

```
Client → Server:  {"type": "message", "data": {"content": "Hello", "conversation_id": "123"}}
Server → Client:  {"type": "message", "data": {"content": "Hi", "role": "assistant", "done": false}}
Server → Client:  {"type": "message", "data": {"content": "there!", "role": "assistant", "done": false}}
Server → Client:  {"type": "done", "data": {"tokens_used": 42, "model": "phi4-mini:3.8b"}}
```

---

## 29. Error Codes

| Code | HTTP Status | Description |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Request body validation failed |
| `UNAUTHORIZED` | 401 | Missing or invalid JWT |
| `TOKEN_EXPIRED` | 401 | JWT has expired |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource already exists |
| `UNPROCESSABLE_ENTITY` | 422 | Business rule violation |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit exceeded |
| `AI_MODEL_TIMEOUT` | 504 | Model inference timeout |
| `INTERNAL_ERROR` | 500 | Server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

---

## 30. Protobuf Definitions

> All API request/response bodies are Protobuf messages (proto3) serialized as JSON over HTTP.

### 2.1 Request Envelope

```protobuf
syntax = "proto3";
package aeroxe.v1;

message RequestEnvelope {
  string operation = 1;
  string request_id = 2;
  int64 tenant_id = 3;
  int64 user_id = 4;
  bytes data = 5;  // Serialized operation-specific request
}
```

### 2.2 Response Envelope

```protobuf
syntax = "proto3";
package aeroxe.v1;

message ResponseEnvelope {
  string status = 1;
  string operation = 2;
  string request_id = 3;
  bytes data = 4;           // Serialized operation-specific response
  Summary summary = 5;      // For list operations
  Pagination pagination = 6; // For list operations
  ErrorInfo error = 7;      // For error responses
  ResponseMeta meta = 8;
}

message Summary {
  int32 total_items = 1;
  int32 active_items = 2;
  int32 inactive_items = 3;
  map<string, int32> recent_activity = 4;
}

message Pagination {
  int32 total = 1;
  int32 limit = 2;
  int32 offset = 3;
  bool has_more = 4;
}

message ErrorInfo {
  string code = 1;
  string message = 2;
  repeated ErrorDetail details = 3;
}

message ErrorDetail {
  string field = 1;
  string message = 2;
}

message ResponseMeta {
  string timestamp = 1;
}
```

### 2.3 Example Operation Request/Response

```protobuf
// Customer Operations
message CreateCustomerRequest {
  string name = 1;
  string email = 2;
  string phone = 3;
  repeated Address addresses = 4;
  repeated string tags = 5;
}

message CreateCustomerResponse {
  int64 id = 1;
  string name = 2;
  string email = 3;
  string status = 4;
  string created_at = 5;
}

message GetCustomerRequest {
  int64 customer_id = 1;
}

message ListCustomersRequest {
  int32 limit = 1;
  int32 offset = 2;
  string status = 3;
}
```

---

## 31. Performance Targets

| API | Target |
|---|---|
| Authentication | < 200ms |
| Customer CRUD | < 100ms |
| Chat First Token | < 2s |
| RAG Search | < 500ms |
| Agent Start | < 300ms |
| Vision Request | < 5s |
| SQL Query | < 3s |
| Memory Search | < 200ms |
| Workflow Start | < 300ms |
| Gateway Middleware | < 5ms per layer |
