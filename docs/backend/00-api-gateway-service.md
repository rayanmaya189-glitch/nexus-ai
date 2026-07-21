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
| Internal Protocol | Rust trait interfaces (no gRPC internally) |
| External Protocol | REST / WebSocket / optionally gRPC for partners |
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

The API Gateway is the **single external entry point** into the AeroXe Nexus AI monolith. Unlike a microservice gateway that proxies to separate services, this module **directly calls internal Rust modules** through trait interfaces — no HTTP/gRPC hop required.

It handles:

- TLS termination (TLS 1.3)
- Request authentication (JWT validation, delegated to `nexus-identity`)
- API key validation (for programmatic access)
- Tenant extraction and validation
- Rate limiting (per tenant, per user, per endpoint) — Token Bucket algorithm
- Request routing to internal module trait methods
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
              |   Trait-based Module Dispatch       |
              |  (direct Rust fn calls, no network) |
              +------------------------------------+
                               |
   +----------+---------+---------+---------+------+
   |          |         |         |         |      |
   v          v         v         v         v      v
nexus-  nexus-ai-  nexus-   nexus-   nexus-  nexus-
identity gateway   agent    rag      memory  audit
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
            → Authentication (JWT Validation → nexus-identity trait call)
              → Tenant Extraction
                → Rate Limit Check (Redis — Token Bucket)
                  → Request Schema Validation
                    → Authorization (RBAC/ABAC → nexus-identity trait call)
                      → Request Logging
                        → Versioned Router Dispatch
                          → Module Trait Dispatch
                            → Response Audit (→ nexus-audit trait/NATS)
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
| Propagation | Injected into module trait call context |

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

All checks delegate to `nexus-identity::check_permission()` trait method.

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

### 8.1 REST Routes — All Handlers Are Module Trait Calls

```
/api/v1/auth/*             → nexus-identity trait calls
/api/v1/customers/*        → nexus-customer trait calls       ← NEW
/api/v1/ai/chat            → nexus-ai-gateway trait calls
/api/v1/ai/stream          → nexus-ai-gateway (WebSocket upgrade)
/api/v1/agents/*           → nexus-agent trait calls
/api/v1/rag/*              → nexus-rag trait calls
/api/v1/vision/*           → nexus-vision trait calls
/api/v1/sql/*              → nexus-sql-agent trait calls
/api/v1/workflows/*        → nexus-workflow trait calls
/api/v1/memory/*           → nexus-memory trait calls
/api/v1/security/*         → nexus-security-ai trait calls
/api/v1/admin/*            → nexus-identity / nexus-audit trait calls
/api/v1/document-sets/*    → nexus-rag trait calls
/api/v1/kyc/*              → nexus-identity trait calls
/health                    → Gateway (self) — module health
/metrics                   → Gateway (self) — prometheus metrics
```

### 8.2 WebSocket Routes

| Route | Purpose |
|---|---|
| `ws://host/ws/v1/chat/{conversation_id}` | AI chat streaming (versioned) |
| `ws://host/ws/v1/agent/{agent_id}` | Agent execution streaming (versioned) |
| `ws://host/ws/v1/notifications` | Real-time notifications (versioned) |

### 8.3 Route Registration (axum Router Example)

```rust
// src/modules/gateway/api_versioning/router.rs
pub fn build_v1_routes(state: AppState) -> Router {
    Router::new()
        // Auth routes (identity module)
        .route("/auth/login", post(handlers::auth::v1_login))
        .route("/auth/register", post(handlers::auth::v1_register))
        .route("/auth/refresh", post(handlers::auth::v1_refresh))
        .route("/auth/me", get(handlers::auth::v1_me))
        // Customer routes (customer module)
        .route("/customers", post(handlers::customer::v1_create))
        .route("/customers/{id}", get(handlers::customer::v1_get))
        .route("/customers/{id}/suspend", post(handlers::customer::v1_suspend))
        .route("/customers/{id}/activate", post(handlers::customer::v1_activate))
        // AI routes
        .route("/ai/chat", post(handlers::ai::v1_chat))
        .route("/ai/stream", get(handlers::ai::v1_stream_ws))
        // ... etc
        .layer(middleware_stack)
        .with_state(state)
}
```

---

## 9. AppState — Shared Module References

The gateway holds trait object references to all modules:

```rust
// src/modules/gateway/mod.rs
pub struct AppState {
    pub identity: Arc<dyn IdentityService>,
    pub customer: Arc<dyn CustomerService>,           // ← NEW
    pub ai_gateway: Arc<dyn AIGatewayService>,
    pub agent: Arc<dyn AgentService>,
    pub rag: Arc<dyn RagService>,
    pub vision: Arc<dyn VisionService>,
    pub sql_agent: Arc<dyn SQLAgentService>,
    pub memory: Arc<dyn MemoryService>,
    pub workflow: Arc<dyn WorkflowService>,
    pub security: Arc<dyn SecurityService>,
    pub audit: Arc<dyn AuditService>,
    pub notifications: Arc<dyn NotificationService>,
    pub model_registry: Arc<dyn ModelRegistryService>,
    pub config: Arc<dyn ConfigService>,
    pub ecosystem: Arc<dyn EcosystemService>,
}
```

Each trait is defined in the module's public API and implemented by the module's application layer. Because all modules live in the same binary, **there is zero network overhead** between gateway and domain logic.

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
| Valid request → 200 | Correct JSON body |

---

## 11. WebSocket Gateway

### Connection Lifecycle

```
Client connects: ws://host/ws/v1/chat/{conversation_id}
  → Upgrade to WebSocket (axum built-in)
    → Authenticate (query param or first message)
    → Validate tenant (nexus-identity trait call)
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

---

## 12. Error Response Standard

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

| Code | HTTP Status | Description |
|---|---|---|
| `UNAUTHORIZED` | 401 | Missing or invalid JWT |
| `TOKEN_EXPIRED` | 401 | JWT has expired |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `TENANT_VIOLATION` | 403 | Cross-tenant access attempt |
| `NOT_FOUND` | 404 | Resource not found |
| `VERSION_NOT_FOUND` | 404 | API version does not exist |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit hit |
| `REQUEST_TIMEOUT` | 504 | Module processing timeout |
| `AI_MODEL_TIMEOUT` | 504 | AI model inference timeout |
| `INTERNAL_ERROR` | 500 | Unexpected server error |
| `SERVICE_UNAVAILABLE` | 503 | Module unavailable |
| `VERSION_DEPRECATED` | 400 | Deprecated API version |

---

## 13. Health Checks

```
GET /health
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

## 14. Configuration

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

## 15. Observability

### Metrics (Prometheus)

| Metric | Type | Labels |
|---|---|---|
| `gateway_requests_total` | Counter | method, path, api_version, status |
| `gateway_request_duration_ms` | Histogram | method, path, api_version |
| `gateway_websocket_connections` | Gauge | tenant_id |
| `gateway_rate_limit_rejections_total` | Counter | tenant_id, tier, api_version |
| `gateway_auth_failures_total` | Counter | reason, api_version |

### Tracing (OpenTelemetry)

Every request generates a trace with spans for each middleware + module trait call:

```
Trace: gateway.request
  ├── Span: api_version_check
  ├── Span: authentication
  ├── Span: tenant_validation
  ├── Span: rate_limit_check
  ├── Span: request_validation
  ├── Span: handler (module trait dispatch)
  └── Span: audit_logging
```

---

## 16. NATS Events

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
