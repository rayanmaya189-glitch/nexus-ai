# AeroXe Nexus AI — API Specification

## REST + WebSocket + Streaming — Versioned

---

## 1. API Design Principles

| Principle | Implementation |
|---|---|
| Versioning | `/api/v{version}/` prefix |
| Authentication | JWT Bearer (`Authorization` header) |
| Tenant Isolation | `tenant_id` from JWT, enforced by gateway |
| **HTTP Methods** | **POST, GET, PATCH, DELETE only (no PUT)** |
| **Pagination** | **Server-side: limit (default 10) + offset** |
| **Business Status** | **Every response includes business status code** |
| **Response Consistency** | **Same envelope format across all endpoints** |
| **List Summary** | **All list APIs include summary metrics** |

---

## 2. HTTP Methods

| Method | Usage | Body | Idempotent |
|---|---|---|---|
| `GET` | Read resource(s) | No | Yes |
| `POST` | Create / Execute action | Yes | No |
| `PATCH` | Partial update | Yes | Yes |
| `DELETE` | Remove resource | No | Yes |

**PUT is NOT allowed.** Use POST for creation, PATCH for updates.

---

## 3. Business Status Codes

Every response MUST include a `status` field with a business-specific code:

| Status | Description | HTTP Code |
|---|---|---|
| `SUCCESS` | Operation completed successfully | 200, 201, 204 |
| `CREATED` | Resource created successfully | 201 |
| `UPDATED` | Resource updated successfully | 200 |
| `DELETED` | Resource deleted successfully | 204 |
| `VALIDATION_ERROR` | Request body validation failed | 400 |
| `UNAUTHORIZED` | Missing or invalid JWT | 401 |
| `TOKEN_EXPIRED` | JWT has expired | 401 |
| `FORBIDDEN` | Insufficient permissions | 403 |
| `NOT_FOUND` | Resource does not exist | 404 |
| `CONFLICT` | Resource already exists | 409 |
| `UNPROCESSABLE_ENTITY` | Business rule violation | 422 |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded | 429 |
| `INTERNAL_ERROR` | Unexpected server error | 500 |
| `SERVICE_UNAVAILABLE` | Service temporarily down | 503 |

---

## 4. Server-Side Pagination

All list endpoints MUST support:

```
GET /api/v1/resource?limit=10&offset=0
```

| Param | Type | Default | Max | Description |
|---|---|---|---|---|
| `limit` | int | 10 | 100 | Items per page |
| `offset` | int | 0 | — | Items to skip |

---

## 5. Consistent Response Envelope

### 5.1 Success Response (Single Resource)

```json
{
  "status": "SUCCESS",
  "data": {
    "id": 1,
    "name": "Example",
    "created_at": "2026-07-21T12:00:00Z"
  },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2026-07-21T12:00:00Z"
  }
}
```

### 5.2 Success Response (Create - 201)

```json
{
  "status": "CREATED",
  "data": {
    "id": 1,
    "name": "Example",
    "created_at": "2026-07-21T12:00:00Z"
  },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2026-07-21T12:00:00Z"
  }
}
```

### 5.3 Success Response (Delete - 204)

```json
{
  "status": "DELETED",
  "data": null,
  "meta": {
    "request_id": "uuid",
    "timestamp": "2026-07-21T12:00:00Z"
  }
}
```

### 5.4 List Response with Summary Metrics + Pagination

```json
{
  "status": "SUCCESS",
  "data": [
    {"id": 1, "name": "Item 1", "status": "active"},
    {"id": 2, "name": "Item 2", "status": "inactive"}
  ],
  "summary": {
    "total_items": 1234,
    "active_items": 980,
    "inactive_items": 254,
    "recent_activity": {
      "created_today": 15,
      "updated_today": 42,
      "deleted_today": 3
    }
  },
  "pagination": {
    "total": 1234,
    "limit": 10,
    "offset": 0,
    "has_more": true
  },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2026-07-21T12:00:00Z"
  }
}
```

### 5.5 Error Response

```json
{
  "status": "VALIDATION_ERROR",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email is required",
    "details": [
      {"field": "email", "message": "is required"}
    ]
  },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2026-07-21T12:00:00Z"
  }
}
```

---

## 6. Authentication APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | `SUCCESS` | `200` |
| `POST` | `/api/v1/auth/refresh` | `SUCCESS` | `200` |
| `POST` | `/api/v1/auth/register` | `CREATED` | `201` |
| `GET` | `/api/v1/auth/me` | `SUCCESS` | `200` |
| `POST` | `/api/v1/auth/change-password` | `SUCCESS` | `200` |

---

## 7. Customer APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/customers` | `CREATED` | `201` |
| `GET` | `/api/v1/customers/{id}` | `SUCCESS` | `200` |
| `GET` | `/api/v1/customers?limit=10&offset=0` | `SUCCESS` | `200` |
| `PATCH` | `/api/v1/customers/{id}` | `UPDATED` | `200` |
| `DELETE` | `/api/v1/customers/{id}` | `DELETED` | `204` |
| `POST` | `/api/v1/customers/{id}/suspend` | `UPDATED` | `200` |
| `POST` | `/api/v1/customers/{id}/activate` | `UPDATED` | `200` |

**List Summary Fields:**
```json
{
  "summary": {
    "total_items": 500,
    "active_items": 420,
    "inactive_items": 80,
    "recent_activity": {
      "created_today": 5,
      "updated_today": 12,
      "deleted_today": 1
    }
  }
}
```

---

## 8. AI Chat APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/ai/chat` | `SUCCESS` | `200` |
| `WS` | `wss://host/ws/v1/chat/{conversation_id}` | — | — |

---

## 9. Agent APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/agents/execute` | `CREATED` | `201` |
| `GET` | `/api/v1/agents/executions/{id}` | `SUCCESS` | `200` |
| `POST` | `/api/v1/agents/{id}/document-sets` | `CREATED` | `201` |
| `GET` | `/api/v1/agents/{id}/document-sets?limit=10&offset=0` | `SUCCESS` | `200` |
| `DELETE` | `/api/v1/agents/{id}/document-sets/{set_id}` | `DELETED` | `204` |
| `POST` | `/api/v1/agents/{id}/sql-connections/test` | `SUCCESS` | `200` |
| `POST` | `/api/v1/agents/{id}/sql-connections/discover` | `SUCCESS` | `200` |
| `POST` | `/api/v1/agents/{id}/sql-connections/tables` | `SUCCESS` | `200` |

**List Summary Fields:**
```json
{
  "summary": {
    "total_items": 50,
    "bound_items": 35,
    "unbound_items": 15
  }
}
```

---

## 10. RAG APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/rag/documents` | `CREATED` | `201` |
| `GET` | `/api/v1/rag/documents/{id}/status` | `SUCCESS` | `200` |
| `GET` | `/api/v1/rag/documents?limit=10&offset=0` | `SUCCESS` | `200` |
| `DELETE` | `/api/v1/rag/documents/{id}` | `DELETED` | `204` |
| `POST` | `/api/v1/rag/search` | `SUCCESS` | `200` |
| `POST` | `/api/v1/document-sets` | `CREATED` | `201` |
| `GET` | `/api/v1/document-sets?limit=10&offset=0` | `SUCCESS` | `200` |
| `PATCH` | `/api/v1/document-sets/{id}` | `UPDATED` | `200` |
| `DELETE` | `/api/v1/document-sets/{id}` | `DELETED` | `204` |

**List Summary Fields:**
```json
{
  "summary": {
    "total_items": 200,
    "processed_items": 180,
    "processing_items": 15,
    "failed_items": 5,
    "total_chunks": 45000,
    "total_size_bytes": 1073741824
  }
}
```

---

## 11. Vision APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/vision/analyze` | `SUCCESS` | `200` |
| `POST` | `/api/v1/vision/ocr` | `SUCCESS` | `200` |
| `POST` | `/api/v1/vision/batch` | `SUCCESS` | `200` |

---

## 12. SQL Intelligence APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/sql/query` | `SUCCESS` | `200` |
| `POST` | `/api/v1/sql/generate` | `SUCCESS` | `200` |

---

## 13. Memory APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/memory` | `CREATED` | `201` |
| `GET` | `/api/v1/memory/search?q=...&limit=10&offset=0` | `SUCCESS` | `200` |
| `GET` | `/api/v1/memory/context/{session_id}` | `SUCCESS` | `200` |
| `DELETE` | `/api/v1/memory/{id}` | `DELETED` | `204` |

**List Summary Fields:**
```json
{
  "summary": {
    "total_items": 500,
    "preference_items": 200,
    "fact_items": 250,
    "conversation_items": 50,
    "avg_importance": 0.72
  }
}
```

---

## 14. Workflow APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/workflows/start` | `CREATED` | `201` |
| `GET` | `/api/v1/workflows/{id}` | `SUCCESS` | `200` |
| `GET` | `/api/v1/workflows?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/workflows/{id}/steps/{step_id}/approve` | `UPDATED` | `200` |
| `POST` | `/api/v1/workflows/{id}/cancel` | `UPDATED` | `200` |

**List Summary Fields:**
```json
{
  "summary": {
    "total_items": 100,
    "running_items": 25,
    "completed_items": 70,
    "failed_items": 5,
    "avg_duration_seconds": 342
  }
}
```

---

## 15. Model Management APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `GET` | `/api/v1/models?limit=10&offset=0` | `SUCCESS` | `200` |
| `GET` | `/api/v1/models/{name}` | `SUCCESS` | `200` |
| `POST` | `/api/v1/models/pull` | `CREATED` | `201` |
| `DELETE` | `/api/v1/models/{name}` | `DELETED` | `204` |
| `GET` | `/api/v1/models/usage` | `SUCCESS` | `200` |

---

## 16. KYC APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `GET` | `/api/v1/kyc/status` | `SUCCESS` | `200` |
| `POST` | `/api/v1/kyc/documents` | `CREATED` | `201` |
| `GET` | `/api/v1/kyc/documents?limit=10&offset=0` | `SUCCESS` | `200` |
| `DELETE` | `/api/v1/kyc/documents/{id}` | `DELETED` | `204` |
| `POST` | `/api/v1/kyc/submit` | `UPDATED` | `200` |
| `POST` | `/api/v1/kyc/review` | `UPDATED` | `200` |

---

## 17. Security APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/security/scan` | `CREATED` | `201` |
| `POST` | `/api/v1/security/review` | `SUCCESS` | `200` |

---

## 18. Telephony APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/telephony/webhook/inbound` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/outbound` | `CREATED` | `201` |
| `GET` | `/api/v1/telephony/calls/{call_id}` | `SUCCESS` | `200` |
| `GET` | `/api/v1/telephony/calls?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/hold` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/resume` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/transfer` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/end` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/recording/start` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/recording/stop` | `UPDATED` | `200` |
| `GET` | `/api/v1/telephony/calls/{call_id}/transcript` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/auth/verify-pin` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/auth/verify-voice` | `SUCCESS` | `200` |
| `GET` | `/api/v1/telephony/calls/{call_id}/auth/status` | `SUCCESS` | `200` |
| `GET` | `/api/v1/telephony/voicemails?limit=10&offset=0` | `SUCCESS` | `200` |
| `GET` | `/api/v1/telephony/voicemails/{id}` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/voicemails/{id}/handle` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/ivr-flows` | `CREATED` | `201` |
| `GET` | `/api/v1/telephony/ivr-flows?limit=10&offset=0` | `SUCCESS` | `200` |
| `PATCH` | `/api/v1/telephony/ivr-flows/{id}` | `UPDATED` | `200` |
| `DELETE` | `/api/v1/telephony/ivr-flows/{id}` | `DELETED` | `204` |
| `POST` | `/api/v1/telephony/calls/{call_id}/monitor/listen` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/monitor/whisper` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/monitor/barge-in` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/monitor/stop` | `UPDATED` | `200` |
| `WS` | `wss://host/ws/v1/telephony/{call_id}` | — | — |
| `WS` | `wss://host/ws/v1/telephony/monitor/{call_id}` | — | — |

**List Summary Fields (Calls):**
```json
{
  "summary": {
    "total_items": 5000,
    "active_calls": 12,
    "completed_calls": 4900,
    "failed_calls": 88,
    "avg_duration_seconds": 245,
    "avg_wait_seconds": 32
  }
}
```

**List Summary Fields (Voicemails):**
```json
{
  "summary": {
    "total_items": 150,
    "new_voicemails": 8,
    "listened_voicemails": 100,
    "handled_voicemails": 42,
    "avg_duration_seconds": 45
  }
}
```

---

## 19. Conversation APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/conversations` | `CREATED` | `201` |
| `GET` | `/api/v1/conversations/{id}` | `SUCCESS` | `200` |
| `GET` | `/api/v1/conversations?limit=10&offset=0` | `SUCCESS` | `200` |
| `GET` | `/api/v1/conversations/{id}/messages?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/conversations/{id}/messages` | `CREATED` | `201` |
| `GET` | `/api/v1/conversations/{id}/state` | `SUCCESS` | `200` |
| `POST` | `/api/v1/conversations/{id}/end` | `UPDATED` | `200` |
| `POST` | `/api/v1/conversations/{id}/branch` | `CREATED` | `201` |
| `DELETE` | `/api/v1/conversations/{id}` | `DELETED` | `204` |

**List Summary Fields:**
```json
{
  "summary": {
    "total_items": 2000,
    "active_conversations": 42,
    "completed_conversations": 1900,
    "escalated_conversations": 58,
    "avg_duration_seconds": 342,
    "avg_csat_score": 4.2
  }
}
```

---

## 20. STT APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/stt/sessions` | `CREATED` | `201` |
| `POST` | `/api/v1/stt/sessions/{session_id}/audio` | `SUCCESS` | `200` |
| `POST` | `/api/v1/stt/sessions/{session_id}/end` | `UPDATED` | `200` |
| `POST` | `/api/v1/stt/transcribe` | `SUCCESS` | `200` |
| `GET` | `/api/v1/stt/sessions/{session_id}` | `SUCCESS` | `200` |

---

## 21. TTS APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/tts/synthesize` | `SUCCESS` | `200` |
| `POST` | `/api/v1/tts/ssml` | `SUCCESS` | `200` |
| `GET` | `/api/v1/tts/voices?limit=10&offset=0` | `SUCCESS` | `200` |
| `GET` | `/api/v1/tts/voices/{voice_id}` | `SUCCESS` | `200` |
| `GET` | `/api/v1/tts/voices/{voice_id}/preview?text=Hello` | `SUCCESS` | `200` |
| `POST` | `/api/v1/tts/voices/clone` | `CREATED` | `201` |
| `DELETE` | `/api/v1/tts/voices/clone/{clone_id}` | `DELETED` | `204` |

---

## 22. Outbound APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/outbound/campaigns` | `CREATED` | `201` |
| `GET` | `/api/v1/outbound/campaigns/{id}` | `SUCCESS` | `200` |
| `GET` | `/api/v1/outbound/campaigns?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/outbound/campaigns/{id}/start` | `UPDATED` | `200` |
| `POST` | `/api/v1/outbound/campaigns/{id}/pause` | `UPDATED` | `200` |
| `POST` | `/api/v1/outbound/campaigns/{id}/cancel` | `UPDATED` | `200` |
| `GET` | `/api/v1/outbound/campaigns/{id}/stats` | `SUCCESS` | `200` |
| `POST` | `/api/v1/outbound/callbacks` | `CREATED` | `201` |
| `GET` | `/api/v1/outbound/callbacks?limit=10&offset=0` | `SUCCESS` | `200` |
| `DELETE` | `/api/v1/outbound/callbacks/{id}` | `DELETED` | `204` |
| `POST` | `/api/v1/outbound/dnc` | `CREATED` | `201` |
| `GET` | `/api/v1/outbound/dnc?limit=10&offset=0` | `SUCCESS` | `200` |
| `DELETE` | `/api/v1/outbound/dnc/{id}` | `DELETED` | `204` |

**List Summary Fields (Campaigns):**
```json
{
  "summary": {
    "total_items": 50,
    "active_campaigns": 5,
    "completed_campaigns": 40,
    "failed_campaigns": 5,
    "total_calls_attempted": 10000,
    "total_calls_connected": 8500,
    "avg_success_rate": 0.85
  }
}
```

**List Summary Fields (Callbacks):**
```json
{
  "summary": {
    "total_items": 200,
    "scheduled_callbacks": 30,
    "completed_callbacks": 160,
    "cancelled_callbacks": 10
  }
}
```

---

## 23. Webhook APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/webhooks` | `CREATED` | `201` |
| `GET` | `/api/v1/webhooks/{id}` | `SUCCESS` | `200` |
| `GET` | `/api/v1/webhooks?limit=10&offset=0` | `SUCCESS` | `200` |
| `PATCH` | `/api/v1/webhooks/{id}` | `UPDATED` | `200` |
| `DELETE` | `/api/v1/webhooks/{id}` | `DELETED` | `204` |
| `POST` | `/api/v1/webhooks/{id}/test` | `SUCCESS` | `200` |
| `GET` | `/api/v1/webhooks/{id}/deliveries?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/webhooks/{id}/deliveries/{delivery_id}/retry` | `SUCCESS` | `200` |

**List Summary Fields (Webhooks):**
```json
{
  "summary": {
    "total_items": 25,
    "active_webhooks": 20,
    "paused_webhooks": 3,
    "failed_webhooks": 2
  }
}
```

**List Summary Fields (Deliveries):**
```json
{
  "summary": {
    "total_items": 5000,
    "delivered": 4800,
    "failed": 150,
    "pending": 50,
    "avg_latency_ms": 245
  }
}
```

---

## 24. Analytics APIs

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `GET` | `/api/v1/analytics/dashboard?start=...&end=...` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/realtime` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/conversations?start=...&end=...` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/calls?start=...&end=...` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/agents?limit=10&offset=0` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/agents/{agent_id}/performance` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/costs?start=...&end=...` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/costs/customers/{customer_id}` | `SUCCESS` | `200` |
| `POST` | `/api/v1/analytics/reports` | `CREATED` | `201` |
| `GET` | `/api/v1/analytics/reports?limit=10&offset=0` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/reports/{id}` | `SUCCESS` | `200` |

**List Summary Fields (Agent Metrics):**
```json
{
  "summary": {
    "total_items": 50,
    "avg_response_time_ms": 1250,
    "avg_csat_score": 4.2,
    "avg_resolution_rate": 0.87,
    "total_tokens_used": 1500000,
    "total_cost": 1250.50
  }
}
```

---

## 25. Health & Observability

| Method | Endpoint | Status | HTTP |
|---|---|---|---|
| `GET` | `/health` | `SUCCESS` | `200` |
| `GET` | `/metrics` | `SUCCESS` | `200` |

---

## 26. Error Codes

| Code | HTTP | Business Status | Description |
|---|---|---|---|
| `VALIDATION_ERROR` | 400 | `VALIDATION_ERROR` | Request body validation failed |
| `UNAUTHORIZED` | 401 | `UNAUTHORIZED` | Missing or invalid JWT |
| `TOKEN_EXPIRED` | 401 | `TOKEN_EXPIRED` | JWT has expired |
| `FORBIDDEN` | 403 | `FORBIDDEN` | Insufficient permissions |
| `TENANT_VIOLATION` | 403 | `FORBIDDEN` | Cross-tenant access attempt |
| `NOT_FOUND` | 404 | `NOT_FOUND` | Resource not found |
| `CONFLICT` | 409 | `CONFLICT` | Resource already exists |
| `UNPROCESSABLE_ENTITY` | 422 | `UNPROCESSABLE_ENTITY` | Business rule violation |
| `RATE_LIMIT_EXCEEDED` | 429 | `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `AI_MODEL_TIMEOUT` | 504 | `INTERNAL_ERROR` | Model inference timeout |
| `INTERNAL_ERROR` | 500 | `INTERNAL_ERROR` | Server error |
| `SERVICE_UNAVAILABLE` | 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

---

## 27. Performance Targets

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
