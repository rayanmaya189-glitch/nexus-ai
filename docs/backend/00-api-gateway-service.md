# AeroXe Nexus AI — API Gateway Module

## External Traffic Entry Point, Request Routing & Cross-Cutting Concerns

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-gateway` |
| Bounded Context | API Gateway |
| Domain Type | Supporting Domain |
| Language | Rust (edition 2024) |
| Framework | **axum** (HTTP + WebSocket) |
| Internal Protocol | gRPC (synchronous) or NATS (async) between services |
| External Protocol | Protobuf JSON over HTTP (PATCH/POST/DELETE only) / WebSocket |
| Database | Redis (rate limiting, session caching) — **no PostgreSQL** |
| HTTP Port | 8080 |
| WebSocket Endpoint | `/ws/*` (integrated in same server) |

---

## 2. Folder Structure

```
src/modules/gateway/
├── mod.rs                        # Public re-exports
├── auth/
│   ├── mod.rs
│   ├── jwt_validator.rs          # JWT validation logic
│   └── api_key_validator.rs      # API key authentication
├── rate_limiter/
│   ├── mod.rs
│   └── token_bucket.rs           # Token bucket algorithm (Redis)
├── request_validator/
│   ├── mod.rs
│   └── schemas/
│       ├── mod.rs
│       ├── auth_schemas.rs       # Login/register request schemas
│       ├── ai_schemas.rs         # AI chat request schemas
│       └── customer_schemas.rs   # Customer request schemas
├── api_versioning/
│   ├── mod.rs
│   └── router.rs                 # API version routing middleware
└── tests/                        # Gateway integration tests
```

---

## 3. Purpose

The API Gateway is the **single external entry point** into the AeroXe Nexus AI monolith. It routes requests to the appropriate service module via **gRPC** (synchronous operations) or **NATS** (async fire-and-forget operations). All request/response bodies are **Protobuf messages serialized as JSON** over HTTP.

It handles:

- TLS termination (TLS 1.3)
- Request authentication (JWT validation, delegated to `nexus-identity`)
- API key validation (for programmatic access)
- Tenant extraction and validation
- Rate limiting (per tenant, per user, per endpoint) — Token Bucket algorithm
- Request routing to internal services via gRPC (sync) or NATS (async)
- API versioning enforcement (`/api/v1/`, `/api/v2/`, etc.)
- WebSocket connection management and streaming relay
- Request/response logging and audit trail
- DDoS protection and IP filtering
- CORS handling
- Request schema validation
- Metrics exposition (`/metrics`)

---

## 4. Architecture

```
                           INTERNET
                               |
                        Firewall / WAF
                               |
                        Load Balancer
                               |
              +------------------------------------+
              |         nexus-gateway Module        |
              |         (axum HTTP/WS Server)       |
              |  src/modules/gateway/               |
              +------------------------------------+
                               |
              +------------------------------------+
              |         Middleware Pipeline         |
              |  Auth → Tenant → Rate-Limit →      |
              |  Authorization → Audit              |
              |  (src/modules/gateway/auth/)        |
              |  (src/modules/gateway/rate_limiter/)|
              +------------------------------------+
                               |
              +------------------------------------+
              |   API Versioning Router             |
              |   (src/modules/gateway/             |
              |    api_versioning/router.rs)        |
              +------------------------------------+
                               |
              +------------------------------------+
              |   Service Dispatch Layer            |
              |  gRPC (sync) / NATS (async)        |
              +------------------------------------+
                               |
   +----------+---------+---------+---------+------+
   |          |         |         |         |      |
   v          v         v         v         v      v
nexus-  nexus-ai-  nexus-   nexus-   nexus-  nexus-
identity gateway   agent    rag      memory  audit
(via     (via      (via     (via     (via     (via
gRPC)    gRPC)     gRPC)    gRPC)    gRPC)   NATS)
```

---

## 5. Request Processing Pipeline

Every inbound request passes through the middleware chain. All middleware are axum layers within the same process:

```
Inbound Request
  → TLS Termination
    → Request ID Generation
      → IP Filtering / DDoS Check
        → CORS Handler
          → API Version Extraction (/api/v{version}/...)
            → Authentication (JWT Validation → identity service via gRPC)
              → Tenant Extraction
                → Rate Limit Check (Redis — Token Bucket)
                  → Request Schema Validation (Protobuf)
                    → Authorization (RBAC/ABAC → identity service via gRPC)
                      → Request Logging
                        → Versioned Router Dispatch
                          → Service Dispatch (gRPC sync / NATS async)
                            → Response Serialization (Protobuf JSON)
                              → Audit Event (→ audit service via NATS)
```

---

## 6. API Versioning Strategy

### 6.1 Versioning Middleware

The gateway enforces API versioning at the router level. All routes are prefixed with `/api/v{version}/`.

| Version | Status | Date |
|---|---|---|
| `v1` | Active | Launch |
| `v2` | Planned | Future |

### 6.2 Version Negotiation

Supported methods (in priority order):

1. **URL Path** (default): `/api/v1/auth/login`
2. **Header** (optional): `Accept-Version: v1`
3. **Query Parameter** (optional): `?api_version=1`

### 6.3 Breaking Change Policy

| Change Type | Version Bump | Deprecation Period |
|---|---|---|
| New endpoint | Minor (same version) | — |
| New optional field | Minor (same version) | — |
| Remove field | Major (new version) | 6 months deprecation warning |
| Change field semantics | Major (new version) | 6 months deprecation warning |
| Remove endpoint | Major (new version) | 6 months deprecation warning |

Deprecated endpoints return a `Sunset` header and `X-Deprecated` header.

### 6.4 Version Router Implementation

```rust
// src/modules/gateway/api_versioning/router.rs
pub enum ApiVersion {
    V1,
    V2,
}

pub fn build_versioned_router(state: AppState) -> Router {
    Router::new()
        .nest("/api/v1", build_v1_routes(state.clone()))
        .nest("/api/v2", build_v2_routes(state))
        .route("/health", get(health_check))
        .route("/metrics", get(metrics_handler))
}
```

---

## 7. Middleware Stack (axum Layers)

### 7.1 Request ID Middleware

| Property | Value |
|---|---|
| Header | `X-Request-ID` |
| Format | UUID v4 |
| Propagation | Injected into gRPC request context |

### 7.2 Authentication Middleware

Located in `src/modules/gateway/auth/`.

| Step | Description |
|---|---|
| Extract | `Authorization: Bearer <jwt>` header or `X-API-Key` header |
| Validate | JWT signature, expiry, issuer (calls `nexus-identity::verify_token()`) |
| Validate (API Key) | Hash comparison with stored keys |
| Decode | Extract `user_id`, `tenant_id`, `roles`, `permissions` |
| Attach | Set in axum request extension |
| Skip | Public endpoints (`/api/v1/auth/login`, `/api/v1/auth/register`, `/health`) |

### 7.3 Tenant Middleware

| Step | Description |
|---|---|
| Extract | `tenant_id` from JWT claims |
| Validate | Tenant exists and is active (calls `nexus-identity::validate_tenant()`) |
| Propagate | Attached to request extensions for downstream use |
| Enforce | Cross-tenant access blocked |

### 7.4 Rate Limiting Middleware

Located in `src/modules/gateway/rate_limiter/token_bucket.rs`.

| Tier | Limit | Burst |
|---|---|---|
| Free | 100 AI requests/min | 10 |
| Customer | 500 AI requests/min | 50 |
| Enterprise | 10,000 AI requests/min | 500 |
| Admin | Unlimited | Unlimited |

Algorithm: **Token Bucket** (Redis-backed)

| Endpoint Category | Rate Limit Key |
|---|---|
| AI Chat | `rate:v1:{tenant_id}:ai_chat` |
| RAG Search | `rate:v1:{tenant_id}:rag` |
| Auth | `rate:v1:{ip}:auth` |
| General API | `rate:v1:{tenant_id}:api` |

Note: Rate limit keys include the API version prefix to allow separate limits per version.

### 7.5 Request Validation Middleware

Located in `src/modules/gateway/request_validator/schemas/`.

Validates request bodies against versioned schemas before dispatching to handlers.

```rust
// src/modules/gateway/request_validator/schemas/auth_schemas.rs
#[derive(Deserialize, Validate)]
pub struct V1LoginRequest {
    #[validate(email)]
    pub email: String,
    #[validate(length(min = 8, max = 128))]
    pub password: String,
}

#[derive(Deserialize, Validate)]
pub struct V2LoginRequest {
    #[validate(email)]
    pub email: String,
    #[validate(length(min = 8, max = 128))]
    pub password: String,
    #[serde(default)]
    pub mfa_code: Option<String>,
}
```

### 7.6 Authorization Middleware

| Check | Description |
|---|---|
| Role | User has required role for endpoint |
| Permission | User has required permission for resource |
| ABAC | Context-aware policies (time, location, resource state) |

All checks delegate to `nexus-identity::check_permission()` via gRPC.

### 7.7 Logging Middleware

```rust
// tracing spans for every request
#[instrument(skip_all, fields(
    request_id = %request_id,
    api_version = %api_version,
    tenant_id = %tenant_id,
    user_id = %user_id,
    method = %method,
    path = %path,
))]
pub async fn log_middleware(req: Request, next: Next) -> Response { ... }
```

---

## 8. Routing Architecture

### 8.1 Routes — Gateway → Service Dispatch

Read operations use **POST** with a query/read request body (e.g., `POST /api/v1/customers/read` with `{id: 123}`).

```
/api/v1/auth/*             → identity service (gRPC)
/api/v1/customers/*        → customer service (gRPC)
/api/v1/ai/chat            → ai-gateway service (gRPC)
/api/v1/ai/stream          → ai-gateway service (WebSocket upgrade)
/api/v1/agents/*           → agent service (gRPC)
/api/v1/rag/*              → rag service (gRPC)
/api/v1/vision/*           → vision service (gRPC)
/api/v1/sql/*              → sql-agent service (gRPC)
/api/v1/workflows/*        → workflow service (gRPC)
/api/v1/memory/*           → memory service (gRPC)
/api/v1/security/*         → security service (gRPC)
/api/v1/admin/*            → identity / audit service (gRPC)
/api/v1/document-sets/*    → rag service (gRPC)
/api/v1/kyc/*              → identity service (gRPC)
/api/v1/telephony/*        → telephony service (gRPC)
/api/v1/conversations/*    → conversation service (gRPC)
/api/v1/stt/*              → stt service (gRPC)
/api/v1/tts/*              → tts service (gRPC)
/api/v1/outbound/*         → outbound service (gRPC)
/api/v1/webhooks/*         → webhook service (gRPC)
/api/v1/analytics/*        → analytics service (gRPC)
/health                    → Gateway (self) — module health
/metrics                   → Gateway (self) — prometheus metrics
```

### 8.2 WebSocket Routes

| Route | Purpose |
|---|---|
| `ws://host/ws/v1/chat/{conversation_id}` | AI chat streaming |
| `ws://host/ws/v1/agent/{agent_id}` | Agent execution streaming |
| `ws://host/ws/v1/notifications` | Real-time notifications |
| `ws://host/ws/v1/telephony/{call_id}` | Audio streaming for voice calls |
| `ws://host/ws/v1/telephony/monitor/{call_id}` | Live call monitoring |

### 8.3 API Method Standards

| Method | Usage | Allowed |
|---|---|---|
| `PATCH` | Partial update | **ALLOWED** |
| `POST` | Create resource, execute operation, read/query | **ALLOWED** |
| `DELETE` | Remove/deactivate resource | **ALLOWED** |
| `GET` | — | **NOT ALLOWED** |
| `PUT` | — | **NOT ALLOWED** |

**Read operations** use `POST` with a query/read request body (e.g., `POST /api/v1/customers/read` with `{id: 123}`).

**Write operations** use `PATCH` for partial updates and `POST` for creation.

**Resource IDs** may appear in the URL path or request body.

### 8.4 Request/Response Format (Protobuf JSON)

All request and response bodies are **Protobuf messages** (proto3) serialized as JSON over HTTP. The gateway serializes/deserializes Protobuf JSON.

**Request Envelope (Protobuf):**
```json
{
  "operation": "GetCustomer",
  "request_id": "uuid",
  "tenant_id": 1,
  "user_id": 123,
  "data": {"customer_id": 456}
}
```

**Response Envelope (Protobuf):**
```json
{
  "status": "SUCCESS",
  "operation": "GetCustomer",
  "request_id": "uuid",
  "data": {"id": 456, "name": "Acme Corp"},
  "meta": {"timestamp": "2026-07-21T12:00:00Z"}
}
```

**List Response Envelope (Protobuf):**
```json
{
  "status": "SUCCESS",
  "operation": "ListCustomers",
  "request_id": "uuid",
  "data": [...],
  "summary": {"total_items": 1234, "active_items": 980},
  "pagination": {"total": 1234, "limit": 10, "offset": 0, "has_more": true},
  "meta": {"timestamp": "2026-07-21T12:00:00Z"}
}
```

### 8.5 HTTP Status Codes

| Code | Usage |
|---|---|
| `200` | Successful operation |
| `400` | Bad request / validation error |
| `401` | Unauthorized (missing/invalid JWT) |
| `403` | Forbidden (insufficient permissions) |
| `404` | Resource not found |
| `409` | Conflict (already exists) |
| `422` | Business rule violation |
| `429` | Rate limit exceeded |
| `500` | Internal server error |
| `503` | Service unavailable |

**Note:** HTTP status is always 200 for successful operations. Business status (`SUCCESS`, `CREATED`, `UPDATED`, `DELETED`) is in the response body.

### 8.6 Route Registration (axum Router Example)

```rust
// src/modules/gateway/api_versioning/router.rs
pub fn build_v1_routes(state: AppState) -> Router {
    Router::new()
        // Auth routes - ALL POST
        .route("/auth/login", post(handlers::auth::v1_login))
        .route("/auth/register", post(handlers::auth::v1_register))
        .route("/auth/refresh", post(handlers::auth::v1_refresh))
        .route("/auth/me", post(handlers::auth::v1_me))
        // Customer routes
        .route("/customers", post(handlers::customer::v1_create))
        .route("/customers/read", post(handlers::customer::v1_read))
        .route("/customers/{id}", patch(handlers::customer::v1_update))
        .route("/customers/{id}/suspend", post(handlers::customer::v1_suspend))
        .route("/customers/{id}/activate", post(handlers::customer::v1_activate))
        .route("/customers/{id}", delete(handlers::customer::v1_delete))
        // AI routes
        .route("/ai/chat", post(handlers::ai::v1_chat))
        .route("/ai/stream", post(handlers::ai::v1_stream_ws))
        // ... etc
        .layer(middleware_stack)
        .with_state(state)
}
```

---

## 9. AppState — Service Clients

The gateway holds gRPC clients (for sync operations) and NATS publishers (for async operations) for each service:

```rust
// src/modules/gateway/mod.rs
pub struct AppState {
    pub identity_client: IdentityServiceClient,
    pub customer_client: CustomerServiceClient,
    pub ai_gateway_client: AIGatewayServiceClient,
    pub agent_client: AgentServiceClient,
    pub rag_client: RagServiceClient,
    pub vision_client: VisionServiceClient,
    pub sql_agent_client: SQLAgentServiceClient,
    pub memory_client: MemoryServiceClient,
    pub workflow_client: WorkflowServiceClient,
    pub security_client: SecurityServiceClient,
    pub notification_client: NotificationServiceClient,
    pub model_registry_client: ModelRegistryServiceClient,
    pub config_client: ConfigServiceClient,
    pub ecosystem_client: EcosystemServiceClient,
    // Voice/Telephony services
    pub telephony_client: TelephonyServiceClient,
    pub conversation_client: ConversationServiceClient,
    pub stt_client: STTServiceClient,
    pub tts_client: TTSServiceClient,
    pub analytics_client: AnalyticsServiceClient,
    pub webhook_client: WebhookServiceClient,
    pub outbound_client: OutboundServiceClient,
    // Async event publishers
    pub audit_publisher: NatsPublisher,
    pub notification_publisher: NatsPublisher,
}
```

Each client connects to its service via gRPC (synchronous) or NATS (async). Services CAN be extracted to separate binaries later.

---

## 10. TDD Contract — Gateway Module Tests

### 10.1 Unit Tests (Mocked Module Traits)

```rust
#[tokio::test]
async fn test_ai_chat_handler_with_mocked_agent_v1() {
    let mut mock_ai = MockAIGatewayService::new();
    mock_ai.expect_submit_request()
        .returning(|_| Ok(AIResponse { .. }));

    let app = build_v1_router(state_with_mocks(mock_ai));
    let response = app
        .oneshot(Request::builder()
            .uri("/api/v1/ai/chat")
            .method("POST")
            .header("Authorization", "Bearer valid-jwt")
            .body(Body::from(r#"{"message":"hello"}"#))
            .unwrap())
        .await;

    assert_eq!(response.status(), 200);
}
```

### 10.2 Integration Tests (Full Module Stack)

```rust
#[tokio::test]
async fn test_full_chat_flow_v1() {
    // Spin up real modules with test DB + Ollama mock
    let state = test_app_state().await;
    let app = build_versioned_router(state);
    // Send request + assert response + assert audit event
}
```

### 10.3 Middleware Tests

| Test | Assertion |
|---|---|
| Missing JWT → 401 | `UNAUTHORIZED` error code |
| Expired JWT → 401 | `TOKEN_EXPIRED` error code |
| Wrong tenant → 403 | `TENANT_VIOLATION` error code |
| Rate limit hit → 429 | `RATE_LIMIT_EXCEEDED` error code |
| Invalid API version → 404 | `NOT_FOUND` error code |
| Valid request → 200 | Correct Protobuf JSON body |

---

## 11. WebSocket Gateway

### Connection Lifecycle

```
Client connects: ws://host/ws/v1/chat/{conversation_id}
  → Upgrade to WebSocket (axum built-in)
    → Authenticate (query param or first message)
    → Validate tenant (identity service — gRPC)
    → Create streaming channel to nexus-ai-gateway
    → Relay messages bidirectionally via tokio channels
    → On disconnect: cleanup
```

### Connection Management

| Property | Value |
|---|---|
| Max connections per tenant | 100 |
| Max connections per user | 10 |
| Idle timeout | 5 minutes |
| Message buffer size | 64KB |
| Heartbeat interval | 30 seconds |

### WebSocket Routes

| Route | Purpose |
|---|---|
| `ws://host/ws/v1/chat/{conversation_id}` | AI chat streaming |
| `ws://host/ws/v1/agent/{agent_id}` | Agent execution streaming |
| `ws://host/ws/v1/notifications` | Real-time notifications |
| `ws://host/ws/v1/telephony/{call_id}` | Audio streaming for voice calls |
| `ws://host/ws/v1/telephony/monitor/{call_id}` | Live call monitoring |

---

## 12. API Method Standards

| Method | Usage | Request Body | Idempotent |
|---|---|---|---|
| `PATCH` | Partial update | Yes | Yes |
| `POST` | Create / Read / Execute action | Yes | No |
| `DELETE` | Remove resource | No | Yes |

**GET and PUT are NOT allowed.** Read operations use POST with a query/read request body. Use PATCH for updates.

### Pagination Standards

All list endpoints MUST support server-side pagination via POST with a read request body:

```
POST /api/v1/resource/read
{
  "limit": 10,
  "offset": 0
}
```

| Parameter | Default | Max | Description |
|---|---|---|---|
| `limit` | 10 | 100 | Items per page |
| `offset` | 0 | — | Items to skip |

---

## 13. Error Response Standard

### Business Status Codes

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

### Consistent Response Envelope

**Success (Single Resource):**
```json
{
  "status": "SUCCESS",
  "data": {"id": 1, "name": "Example"},
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

**Success (Create):**
```json
{
  "status": "CREATED",
  "data": {"id": 1, "name": "Example"},
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

**Success (Delete):**
```json
{
  "status": "DELETED",
  "data": null,
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

**List with Summary + Pagination:**
```json
{
  "status": "SUCCESS",
  "data": [...],
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
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

**Error:**
```json
{
  "status": "VALIDATION_ERROR",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email is required",
    "details": [{"field": "email", "message": "is required"}]
  },
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

### Summary Fields Per Resource

| Resource | Summary Fields |
|---|---|
| **Customers** | total_items, active_items, inactive_items, created_today, updated_today |
| **Agents** | total_items, bound_items, unbound_items |
| **Documents** | total_items, processed_items, processing_items, failed_items, total_chunks |
| **Workflows** | total_items, running_items, completed_items, failed_items, avg_duration |
| **Calls** | total_items, active_calls, completed_calls, failed_calls, avg_duration |
| **Voicemails** | total_items, new_voicemails, listened, handled, avg_duration |
| **Conversations** | total_items, active, completed, escalated, avg_duration, avg_csat |
| **Campaigns** | total_items, active, completed, failed, total_calls, avg_success_rate |
| **Callbacks** | total_items, scheduled, completed, cancelled |
| **Webhooks** | total_items, active, paused, failed |
| **Deliveries** | total_items, delivered, failed, pending, avg_latency |
| **Agent Metrics** | total_items, avg_response_time, avg_csat, total_tokens, total_cost |

---

## 14. Health Checks

```
POST /health
```

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "api_versions": ["v1"],
  "uptime_seconds": 86400,
  "checks": {
    "redis": "healthy",
    "postgresql": "healthy",
    "nats": "healthy",
    "ollama": "healthy"
  }
}
```

---

## 15. Configuration

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `HTTP_PORT` | HTTP listen port | 8080 |
| `JWT_ISSUER` | Expected JWT issuer | nexus-ai |
| `JWT_SECRET` | JWT verification secret | (required) |
| `REDIS_URL` | Redis connection string | redis://localhost:6379 |
| `DATABASE_URL` | PostgreSQL connection string | postgres://... |
| `NATS_URL` | NATS connection string | nats://localhost:4222 |
| `OLLAMA_URL` | Ollama API URL | http://localhost:11434 |
| `RATE_LIMIT_ENABLED` | Enable rate limiting | true |
| `CORS_ALLOWED_ORIGINS` | Comma-separated origins | * |
| `ENABLED_API_VERSIONS` | Comma-separated enabled versions | v1 |
| `LOG_LEVEL` | Logging level | info |

---

## 16. Observability

### Metrics (Prometheus)

| Metric | Type | Labels |
|---|---|---|
| `gateway_requests_total` | Counter | method, path, api_version, status |
| `gateway_request_duration_ms` | Histogram | method, path, api_version |
| `gateway_websocket_connections` | Gauge | tenant_id |
| `gateway_rate_limit_rejections_total` | Counter | tenant_id, tier, api_version |
| `gateway_auth_failures_total` | Counter | reason, api_version |

### Tracing (OpenTelemetry)

Every request generates a trace with spans for each middleware + gRPC service call:

```
Trace: gateway.request
  ├── Span: api_version_check
  ├── Span: authentication
  ├── Span: tenant_validation
  ├── Span: rate_limit_check
  ├── Span: request_validation
  ├── Span: handler (gRPC service dispatch)
  └── Span: audit_logging
```

---

## 17. NATS Events (Protobuf Payloads)

All NATS event payloads are **Protobuf messages** serialized as binary or JSON. Subject format: `aeroxe.v1.<service>.<event>`.

### Published

| Subject | When |
|---|---|
| `aeroxe.v1.gateway.request.completed` | After every successful request |
| `aeroxe.v1.gateway.auth.failure` | On authentication failure |
| `aeroxe.v1.gateway.rate_limit.exceeded` | On rate limit rejection |

### Subscribed

| Subject | Purpose |
|---|---|
| `aeroxe.v1.gateway.config.reload` | Dynamic configuration updates |

> **Versioning:** All NATS subjects include the `v1` prefix to allow future version coexistence and prevent message format conflicts.
