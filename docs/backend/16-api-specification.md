# AeroXe Nexus AI — API Specification

## REST Protobuf — POST Only — No Path Variables — No Query Strings

---

## 1. API Design Principles

| Principle | Implementation |
|---|---|
| Protocol | REST with Protobuf request/response bodies |
| HTTP Method | **POST only** (no GET, no PUT, no DELETE) |
| Path Variables | **NOT ALLOWED** — resource IDs in request body |
| Query Strings | **NOT ALLOWED** — all parameters in request body |
| Versioning | `/api/v{version}/` prefix |
| Authentication | JWT Bearer (`Authorization` header) |
| Business Status | Every response includes `status` field |
| Pagination | Server-side: `limit` (default 10) + `offset` in request body |

---

## 2. HTTP Methods

| Method | Usage | Allowed |
|---|---|---|
| `GET` | Read resource(s) | **NOT ALLOWED** |
| `POST` | All operations | **REQUIRED** |
| `PUT` | Full replace | **NOT ALLOWED** |
| `DELETE` | Remove resource | **NOT ALLOWED** |
| `PATCH` | Partial update | **NOT ALLOWED** |

**ALL operations use POST.** Resource IDs, parameters, and filters go in the request body.

---

## 3. Request/Response Format

### 3.1 Request Envelope (Protobuf)

Every request is a POST with a JSON body (Protobuf-serialized to JSON):

```json
{
  "operation": "GetCustomer",
  "request_id": "uuid",
  "tenant_id": 1,
  "user_id": 123,
  "data": {
    "customer_id": 456
  }
}
```

### 3.2 Success Response Envelope (Protobuf)

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

### 3.3 List Response Envelope (Protobuf)

```json
{
  "status": "SUCCESS",
  "operation": "ListCustomers",
  "request_id": "uuid",
  "data": [
    {"id": 1, "name": "Item 1", "status": "active"},
    {"id": 2, "name": "Item 2", "status": "inactive"}
  ],
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

### 3.4 Error Response Envelope (Protobuf)

```json
{
  "status": "VALIDATION_ERROR",
  "operation": "CreateCustomer",
  "request_id": "uuid",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email is required",
    "details": [{"field": "email", "message": "is required"}]
  },
  "meta": {
    "timestamp": "2026-07-21T12:00:00Z"
  }
}
```

---

## 4. Business Status Codes

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

## 5. API Endpoint Structure

All endpoints follow: `POST /api/v{version}/{resource}/{operation}`

**No path variables.** Resource IDs go in the request body.

**No query strings.** All parameters go in the request body.

---

## 6. Authentication APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/auth/login` | Login | `SUCCESS` |
| `POST /api/v1/auth/refresh` | Refresh Token | `SUCCESS` |
| `POST /api/v1/auth/register` | Register | `CREATED` |
| `POST /api/v1/auth/me` | Get Current User | `SUCCESS` |
| `POST /api/v1/auth/change-password` | Change Password | `SUCCESS` |

---

## 7. Customer APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/customers/create` | Create Customer | `CREATED` |
| `POST /api/v1/customers/get` | Get Customer | `SUCCESS` |
| `POST /api/v1/customers/list` | List Customers | `SUCCESS` |
| `POST /api/v1/customers/update` | Update Customer | `UPDATED` |
| `POST /api/v1/customers/delete` | Delete Customer | `DELETED` |
| `POST /api/v1/customers/suspend` | Suspend Customer | `UPDATED` |
| `POST /api/v1/customers/activate` | Activate Customer | `UPDATED` |

---

## 8. AI Chat APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/ai/chat` | Submit Chat Request | `SUCCESS` |
| `WS /ws/v1/chat` | Streaming Chat | — |

---

## 9. Agent APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/agents/execute` | Execute Agent | `CREATED` |
| `POST /api/v1/agents/get-execution` | Get Execution Status | `SUCCESS` |
| `POST /api/v1/agents/list-executions` | List Executions | `SUCCESS` |
| `POST /api/v1/agents/bind-document-sets` | Bind Document Sets | `UPDATED` |
| `POST /api/v1/agents/unbind-document-sets` | Unbind Document Sets | `UPDATED` |
| `POST /api/v1/agents/list-document-sets` | List Bound Document Sets | `SUCCESS` |
| `POST /api/v1/agents/test-sql-connection` | Test SQL Connection | `SUCCESS` |
| `POST /api/v1/agents/discover-schema` | Discover Database Schema | `SUCCESS` |
| `POST /api/v1/agents/bind-tables` | Bind Tables | `UPDATED` |

---

## 10. RAG APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/rag/upload-document` | Upload Document | `CREATED` |
| `POST /api/v1/rag/get-document-status` | Get Document Status | `SUCCESS` |
| `POST /api/v1/rag/list-documents` | List Documents | `SUCCESS` |
| `POST /api/v1/rag/delete-document` | Delete Document | `DELETED` |
| `POST /api/v1/rag/search` | Search Knowledge | `SUCCESS` |
| `POST /api/v1/rag/create-document-set` | Create Document Set | `CREATED` |
| `POST /api/v1/rag/list-document-sets` | List Document Sets | `SUCCESS` |
| `POST /api/v1/rag/update-document-set` | Update Document Set | `UPDATED` |
| `POST /api/v1/rag/delete-document-set` | Delete Document Set | `DELETED` |

---

## 11. Vision APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/vision/analyze` | Analyze Image | `SUCCESS` |
| `POST /api/v1/vision/ocr` | Extract Text | `SUCCESS` |
| `POST /api/v1/vision/batch` | Batch Analysis | `SUCCESS` |

---

## 12. SQL Intelligence APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/sql/query` | Execute SQL Query | `SUCCESS` |
| `POST /api/v1/sql/generate` | Generate SQL | `SUCCESS` |

---

## 13. Memory APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/memory/store` | Store Memory | `CREATED` |
| `POST /api/v1/memory/search` | Search Memory | `SUCCESS` |
| `POST /api/v1/memory/get-context` | Get Conversation Context | `SUCCESS` |
| `POST /api/v1/memory/delete` | Delete Memory | `DELETED` |

---

## 14. Workflow APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/workflows/start` | Start Workflow | `CREATED` |
| `POST /api/v1/workflows/get` | Get Workflow | `SUCCESS` |
| `POST /api/v1/workflows/list` | List Workflows | `SUCCESS` |
| `POST /api/v1/workflows/approve-step` | Approve Step | `UPDATED` |
| `POST /api/v1/workflows/cancel` | Cancel Workflow | `UPDATED` |

---

## 15. Model Management APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/models/list` | List Models | `SUCCESS` |
| `POST /api/v1/models/get` | Get Model | `SUCCESS` |
| `POST /api/v1/models/pull` | Pull Model | `CREATED` |
| `POST /api/v1/models/delete` | Delete Model | `DELETED` |
| `POST /api/v1/models/usage` | Get Usage Stats | `SUCCESS` |

---

## 16. KYC APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/kyc/get-status` | Get KYC Status | `SUCCESS` |
| `POST /api/v1/kyc/upload-document` | Upload KYC Document | `CREATED` |
| `POST /api/v1/kyc/list-documents` | List KYC Documents | `SUCCESS` |
| `POST /api/v1/kyc/delete-document` | Delete KYC Document | `DELETED` |
| `POST /api/v1/kyc/submit` | Submit for Review | `UPDATED` |
| `POST /api/v1/kyc/review` | Review KYC | `UPDATED` |

---

## 17. Security APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/security/scan` | Security Scan | `CREATED` |
| `POST /api/v1/security/review` | Code Review | `SUCCESS` |

---

## 18. Telephony APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/telephony/webhook/inbound` | Inbound Call Webhook | `SUCCESS` |
| `POST /api/v1/telephony/calls/initiate` | Initiate Outbound Call | `CREATED` |
| `POST /api/v1/telephony/calls/get` | Get Call Details | `SUCCESS` |
| `POST /api/v1/telephony/calls/list` | List Calls | `SUCCESS` |
| `POST /api/v1/telephony/calls/hold` | Hold Call | `UPDATED` |
| `POST /api/v1/telephony/calls/resume` | Resume Call | `UPDATED` |
| `POST /api/v1/telephony/calls/transfer` | Transfer Call | `UPDATED` |
| `POST /api/v1/telephony/calls/end` | End Call | `UPDATED` |
| `POST /api/v1/telephony/calls/start-recording` | Start Recording | `UPDATED` |
| `POST /api/v1/telephony/calls/stop-recording` | Stop Recording | `UPDATED` |
| `POST /api/v1/telephony/calls/get-transcript` | Get Transcript | `SUCCESS` |
| `POST /api/v1/telephony/calls/verify-pin` | Verify PIN | `SUCCESS` |
| `POST /api/v1/telephony/calls/verify-voice` | Verify Voice | `SUCCESS` |
| `POST /api/v1/telephony/calls/get-auth-status` | Get Auth Status | `SUCCESS` |
| `POST /api/v1/telephony/voicemails/list` | List Voicemails | `SUCCESS` |
| `POST /api/v1/telephony/voicemails/get` | Get Voicemail | `SUCCESS` |
| `POST /api/v1/telephony/voicemails/handle` | Handle Voicemail | `UPDATED` |
| `POST /api/v1/telephony/ivr/create` | Create IVR Flow | `CREATED` |
| `POST /api/v1/telephony/ivr/list` | List IVR Flows | `SUCCESS` |
| `POST /api/v1/telephony/ivr/update` | Update IVR Flow | `UPDATED` |
| `POST /api/v1/telephony/ivr/delete` | Delete IVR Flow | `DELETED` |
| `POST /api/v1/telephony/monitor/listen` | Listen-In | `SUCCESS` |
| `POST /api/v1/telephony/monitor/whisper` | Whisper to Agent | `SUCCESS` |
| `POST /api/v1/telephony/monitor/barge-in` | Barge-In | `SUCCESS` |
| `POST /api/v1/telephony/monitor/stop` | Stop Monitoring | `UPDATED` |
| `WS /ws/v1/telephony` | Audio Stream | — |
| `WS /ws/v1/telephony/monitor` | Live Monitoring | — |

---

## 19. Conversation APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/conversations/create` | Create Conversation | `CREATED` |
| `POST /api/v1/conversations/get` | Get Conversation | `SUCCESS` |
| `POST /api/v1/conversations/list` | List Conversations | `SUCCESS` |
| `POST /api/v1/conversations/get-messages` | Get Messages | `SUCCESS` |
| `POST /api/v1/conversations/add-message` | Add Message | `CREATED` |
| `POST /api/v1/conversations/get-state` | Get State | `SUCCESS` |
| `POST /api/v1/conversations/end` | End Conversation | `UPDATED` |
| `POST /api/v1/conversations/branch` | Branch Conversation | `CREATED` |
| `POST /api/v1/conversations/delete` | Delete Conversation | `DELETED` |

---

## 20. STT APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/stt/start-session` | Start Session | `CREATED` |
| `POST /api/v1/stt/send-audio` | Send Audio Chunk | `SUCCESS` |
| `POST /api/v1/stt/end-session` | End Session | `UPDATED` |
| `POST /api/v1/stt/transcribe` | Batch Transcribe | `SUCCESS` |
| `POST /api/v1/stt/get-session` | Get Session | `SUCCESS` |

---

## 21. TTS APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/tts/synthesize` | Synthesize Speech | `SUCCESS` |
| `POST /api/v1/tts/synthesize-ssml` | Synthesize SSML | `SUCCESS` |
| `POST /api/v1/tts/list-voices` | List Voices | `SUCCESS` |
| `POST /api/v1/tts/get-voice` | Get Voice | `SUCCESS` |
| `POST /api/v1/tts/preview-voice` | Preview Voice | `SUCCESS` |
| `POST /api/v1/tts/clone-voice` | Clone Voice | `CREATED` |
| `POST /api/v1/tts/revoke-clone` | Revoke Clone | `DELETED` |

---

## 22. Outbound APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/outbound/create-campaign` | Create Campaign | `CREATED` |
| `POST /api/v1/outbound/get-campaign` | Get Campaign | `SUCCESS` |
| `POST /api/v1/outbound/list-campaigns` | List Campaigns | `SUCCESS` |
| `POST /api/v1/outbound/start-campaign` | Start Campaign | `UPDATED` |
| `POST /api/v1/outbound/pause-campaign` | Pause Campaign | `UPDATED` |
| `POST /api/v1/outbound/cancel-campaign` | Cancel Campaign | `UPDATED` |
| `POST /api/v1/outbound/get-campaign-stats` | Get Campaign Stats | `SUCCESS` |
| `POST /api/v1/outbound/schedule-callback` | Schedule Callback | `CREATED` |
| `POST /api/v1/outbound/list-callbacks` | List Callbacks | `SUCCESS` |
| `POST /api/v1/outbound/cancel-callback` | Cancel Callback | `DELETED` |
| `POST /api/v1/outbound/add-dnc` | Add to DNC | `CREATED` |
| `POST /api/v1/outbound/list-dnc` | List DNC | `SUCCESS` |
| `POST /api/v1/outbound/remove-dnc` | Remove from DNC | `DELETED` |

---

## 23. Webhook APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/webhooks/create` | Create Webhook | `CREATED` |
| `POST /api/v1/webhooks/get` | Get Webhook | `SUCCESS` |
| `POST /api/v1/webhooks/list` | List Webhooks | `SUCCESS` |
| `POST /api/v1/webhooks/update` | Update Webhook | `UPDATED` |
| `POST /api/v1/webhooks/delete` | Delete Webhook | `DELETED` |
| `POST /api/v1/webhooks/test` | Test Webhook | `SUCCESS` |
| `POST /api/v1/webhooks/list-deliveries` | List Deliveries | `SUCCESS` |
| `POST /api/v1/webhooks/retry-delivery` | Retry Delivery | `SUCCESS` |

---

## 24. Analytics APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/analytics/dashboard` | Get Dashboard | `SUCCESS` |
| `POST /api/v1/analytics/realtime` | Get Real-time Metrics | `SUCCESS` |
| `POST /api/v1/analytics/conversations` | Get Conversation Metrics | `SUCCESS` |
| `POST /api/v1/analytics/calls` | Get Call Metrics | `SUCCESS` |
| `POST /api/v1/analytics/list-agents` | List Agent Metrics | `SUCCESS` |
| `POST /api/v1/analytics/get-agent-performance` | Get Agent Performance | `SUCCESS` |
| `POST /api/v1/analytics/costs` | Get Cost Breakdown | `SUCCESS` |
| `POST /api/v1/analytics/get-customer-cost` | Get Customer Cost | `SUCCESS` |
| `POST /api/v1/analytics/create-report` | Create Report | `CREATED` |
| `POST /api/v1/analytics/list-reports` | List Reports | `SUCCESS` |
| `POST /api/v1/analytics/get-report` | Get Report | `SUCCESS` |

---

## 25. Billing APIs

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/billing/list-plans` | List Plans | `SUCCESS` |
| `POST /api/v1/billing/create-subscription` | Create Subscription | `CREATED` |
| `POST /api/v1/billing/get-subscription` | Get Subscription | `SUCCESS` |
| `POST /api/v1/billing/update-subscription` | Update Subscription | `UPDATED` |
| `POST /api/v1/billing/cancel-subscription` | Cancel Subscription | `UPDATED` |
| `POST /api/v1/billing/get-usage` | Get Usage | `SUCCESS` |
| `POST /api/v1/billing/list-invoices` | List Invoices | `SUCCESS` |
| `POST /api/v1/billing/get-invoice` | Get Invoice | `SUCCESS` |
| `POST /api/v1/billing/pay-invoice` | Pay Invoice | `SUCCESS` |
| `POST /api/v1/billing/list-payments` | List Payments | `SUCCESS` |

---

## 26. Health & Observability

| Endpoint | Operation | Status |
|---|---|---|
| `POST /api/v1/health/check` | Health Check | `SUCCESS` |
| `POST /api/v1/metrics/get` | Get Metrics | `SUCCESS` |

---

## 27. Error Codes

| Code | Business Status | Description |
|---|---|---|
| `VALIDATION_ERROR` | `VALIDATION_ERROR` | Request body validation failed |
| `UNAUTHORIZED` | `UNAUTHORIZED` | Missing or invalid JWT |
| `TOKEN_EXPIRED` | `TOKEN_EXPIRED` | JWT has expired |
| `FORBIDDEN` | `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | `NOT_FOUND` | Resource not found |
| `CONFLICT` | `CONFLICT` | Resource already exists |
| `UNPROCESSABLE_ENTITY` | `UNPROCESSABLE_ENTITY` | Business rule violation |
| `RATE_LIMIT_EXCEEDED` | `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `AI_MODEL_TIMEOUT` | `INTERNAL_ERROR` | Model inference timeout |
| `INTERNAL_ERROR` | `INTERNAL_ERROR` | Server error |
| `SERVICE_UNAVAILABLE` | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

---

## 28. Protobuf Definitions

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

## 29. Performance Targets

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
