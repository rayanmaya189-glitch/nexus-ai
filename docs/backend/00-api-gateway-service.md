# AeroXe Nexus AI — API Gateway Module

## External Traffic Entry Point, Request Routing & Cross-Cutting Concerns

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-gateway` |
| Bounded Context | API Gateway |
| Domain Type | Supporting Domain |
| Language | Rust |
| Framework | **axum** (HTTP + WebSocket) |
| Internal Protocol | Rust trait interfaces (no gRPC internally) |
| External Protocol | REST / WebSocket / optionally gRPC for partners |
| Cache | Redis (rate limiting, session) |
| HTTP Port | 8080 |
| WebSocket Endpoint | `/ws/*` (integrated in same server) |

---

## 2. Purpose

The API Gateway is the **single external entry point** into the AeroXe Nexus AI monolith. Unlike a microservice gateway that proxies to separate services, this module **directly calls internal Rust modules** through trait interfaces — no HTTP/gRPC hop required.

It handles:

- TLS termination (TLS 1.3)
- Request authentication (JWT validation, delegated to `nexus-identity`)
- Tenant extraction and validation
- Rate limiting (per tenant, per user, per endpoint)
- Request routing to internal module trait methods
- WebSocket connection management and streaming relay
- Request/response logging and audit trail
- DDoS protection and IP filtering
- CORS handling
- API versioning enforcement
- Metrics exposition (`/metrics`)

---

## 3. Architecture

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
              +------------------------------------+
                               |
              +------------------------------------+
              |         Middleware Pipeline         |
              |  Auth → Tenant → Rate-Limit →      |
              |  Authorization → Audit              |
              +------------------------------------+
                               |
              +------------------------------------+
              |   Trait-based Module Dispatch       |
              |  (direct Rust fn calls, no network) |
              +------------------------------------+
              |                                     |
   +----------+---------+---------+---------+------+
   |          |         |         |         |      |
   v          v         v         v         v      v
nexus-  nexus-ai-  nexus-   nexus-   nexus-  nexus-
identity gateway   agent    rag      memory  audit
```

---

## 4. Request Processing Pipeline

Every inbound request passes through the middleware chain. All middleware are axum layers within the same process:

```
Inbound Request
  → TLS Termination
    → Request ID Generation
      → IP Filtering / DDoS Check
        → CORS Handler
          → Authentication (JWT Validation → nexus-identity trait call)
            → Tenant Extraction
              → Rate Limit Check (Redis)
                → Authorization (RBAC/ABAC → nexus-identity trait call)
                  → Request Logging
                    → Module Trait Dispatch
                      → Response Audit (→ nexus-audit trait/NATS)
```

---

## 5. Middleware Stack (axum Layers)

### 5.1 Request ID Middleware

| Property | Value |
|---|---|
| Header | `X-Request-ID` |
| Format | UUID v4 |
| Propagation | Injected into module trait call context |

### 5.2 Authentication Middleware

| Step | Description |
|---|---|
| Extract | `Authorization: Bearer <jwt>` header |
| Validate | JWT signature, expiry, issuer (calls `nexus-identity::verify_token()`) |
| Decode | Extract `user_id`, `tenant_id`, `roles`, `permissions` |
| Attach | Set in axum request extension |
| Skip | Public endpoints (`/api/v1/auth/login`, `/api/v1/auth/register`, `/health`) |

### 5.3 Tenant Middleware

| Step | Description |
|---|---|
| Extract | `tenant_id` from JWT claims |
| Validate | Tenant exists and is active (calls `nexus-identity::validate_tenant()`) |
| Propagate | Attached to request extensions for downstream use |
| Enforce | Cross-tenant access blocked |

### 5.4 Rate Limiting Middleware

| Tier | Limit | Burst |
|---|---|---|
| Free | 100 AI requests/min | 10 |
| Customer | 500 AI requests/min | 50 |
| Enterprise | 10,000 AI requests/min | 500 |
| Admin | Unlimited | Unlimited |

Algorithm: **Token Bucket** (Redis-backed)

| Endpoint Category | Rate Limit Key |
|---|---|
| AI Chat | `rate:{tenant_id}:ai_chat` |
| RAG Search | `rate:{tenant_id}:rag` |
| Auth | `rate:{ip}:auth` |
| General API | `rate:{tenant_id}:api` |

### 5.5 Authorization Middleware

| Check | Description |
|---|---|
| Role | User has required role for endpoint |
| Permission | User has required permission for resource |
| ABAC | Context-aware policies (time, location, resource state) |

All checks delegate to `nexus-identity::check_permission()` trait method.

### 5.6 Logging Middleware

```rust
// tracing spans for every request
#[instrument(skip_all, fields(
    request_id = %request_id,
    tenant_id = %tenant_id,
    user_id = %user_id,
    method = %method,
    path = %path,
))]
pub async fn log_middleware(req: Request, next: Next) -> Response { ... }
```

---

## 6. Routing Architecture

### 6.1 REST Routes — All Handlers Are Module Trait Calls

```
/api/v1/auth/*             → nexus-identity trait calls
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

### 6.2 WebSocket Routes

| Route | Purpose |
|---|---|
| `ws://host/ws/chat/{conversation_id}` | AI chat streaming |
| `ws://host/ws/agent/{agent_id}` | Agent execution streaming |
| `ws://host/ws/notifications` | Real-time notifications |

### 6.3 Route Registration (axum Router Example)

```rust
pub fn build_router<A, I, G, R, ...>(/* module traits */) -> Router {
    Router::new()
        // Auth routes
        .route("/api/v1/auth/login", post(handlers::auth::login))
        .route("/api/v1/auth/register", post(handlers::auth::register))
        // AI routes
        .route("/api/v1/ai/chat", post(handlers::ai::chat))
        .route("/api/v1/ai/stream", get(handlers::ai::stream_ws))
        // Agent routes
        .route("/api/v1/agents/execute", post(handlers::agent::execute))
        // ... etc
        .layer(middleware_stack)
        .with_state(AppState { identity, ai_gateway, agent, ... })
}
```

---

## 7. AppState — Shared Module References

The gateway holds trait object references to all modules:

```rust
pub struct AppState {
    pub identity: Arc<dyn IdentityService>,
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

## 8. TDD Contract — Gateway Module Tests

### 8.1 Unit Tests (Mocked Module Traits)

```rust
#[tokio::test]
async fn test_ai_chat_handler_with_mocked_agent() {
    let mut mock_ai = MockAIGatewayService::new();
    mock_ai.expect_submit_request()
        .returning(|_| Ok(AIResponse { .. }));

    let app = build_router(mock_ai, /* other mocks */);
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

### 8.2 Integration Tests (Full Module Stack)

```rust
#[tokio::test]
async fn test_full_chat_flow() {
    // Spin up real modules with test DB + Ollama mock
    let state = test_app_state().await;
    let app = build_router_with_state(state);
    // Send request + assert response + assert audit event
}
```

### 8.3 Middleware Tests

| Test | Assertion |
|---|---|
| Missing JWT → 401 | `UNAUTHORIZED` error code |
| Expired JWT → 401 | `TOKEN_EXPIRED` error code |
| Wrong tenant → 403 | `TENANT_VIOLATION` error code |
| Rate limit hit → 429 | `RATE_LIMIT_EXCEEDED` error code |
| Valid request → 200 | Correct JSON body |

---

## 9. WebSocket Gateway

### Connection Lifecycle

```
Client connects: ws://host/ws/chat/{conversation_id}
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

## 10. Error Response Standard

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

| Code | HTTP Status | Description |
|---|---|---|
| `UNAUTHORIZED` | 401 | Missing or invalid JWT |
| `TOKEN_EXPIRED` | 401 | JWT has expired |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `TENANT_VIOLATION` | 403 | Cross-tenant access attempt |
| `NOT_FOUND` | 404 | Resource not found |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit hit |
| `REQUEST_TIMEOUT` | 504 | Module processing timeout |
| `AI_MODEL_TIMEOUT` | 504 | AI model inference timeout |
| `INTERNAL_ERROR` | 500 | Unexpected server error |
| `SERVICE_UNAVAILABLE` | 503 | Module unavailable |

---

## 11. Health Checks

```
GET /health
```

```json
{
  "status": "healthy",
  "version": "1.0.0",
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

## 12. Configuration

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
| `LOG_LEVEL` | Logging level | info |

---

## 13. Observability

### Metrics (Prometheus)

| Metric | Type | Labels |
|---|---|---|
| `gateway_requests_total` | Counter | method, path, status |
| `gateway_request_duration_ms` | Histogram | method, path |
| `gateway_websocket_connections` | Gauge | tenant_id |
| `gateway_rate_limit_rejections_total` | Counter | tenant_id, tier |
| `gateway_auth_failures_total` | Counter | reason |

### Tracing (OpenTelemetry)

Every request generates a trace with spans for each middleware + module trait call:

```
Trace: gateway.request
  ├── Span: authentication
  ├── Span: tenant_validation
  ├── Span: rate_limit_check
  ├── Span: handler (module trait dispatch)
  └── Span: audit_logging
```

---

## 14. NATS Events

### Published

| Subject | When |
|---|---|
| `aeroxe.gateway.request.completed` | After every successful request |
| `aeroxe.gateway.auth.failure` | On authentication failure |
| `aeroxe.gateway.rate_limit.exceeded` | On rate limit rejection |

### Subscribed

| Subject | Purpose |
|---|---|
| `aeroxe.gateway.config.reload` | Dynamic configuration updates |
