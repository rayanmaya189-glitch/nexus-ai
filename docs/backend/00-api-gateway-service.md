# AeroXe Nexus AI — API Gateway Service

## External Traffic Entry Point, Request Routing & Cross-Cutting Concerns

---

## 1. Service Identity

| Attribute | Value |
|---|---|
| Service Name | `nexus-api-gateway` |
| Bounded Context | API Gateway |
| Domain Type | Supporting Domain |
| Language | Go |
| Framework | Hertz + Kratos |
| Database | None (stateless) |
| Cache | Redis (rate limiting, session) |
| gRPC Port | 50050 |
| HTTP Port | 8080 |
| WebSocket Port | 8081 |

---

## 2. Purpose

The API Gateway is the **single external entry point** for all client traffic into the AeroXe Nexus AI platform. It handles:

- TLS termination (TLS 1.3)
- Request authentication (JWT validation)
- Tenant extraction and validation
- Rate limiting (per tenant, per user, per endpoint)
- Request routing to internal services
- WebSocket connection management and streaming relay
- gRPC-Gateway REST-to-gRPC translation
- Request/response logging and audit trail
- DDoS protection and IP filtering
- CORS handling
- API versioning enforcement

---

## 3. Architecture

```
                           INTERNET
                               |
                        Firewall / WAF
                               |
                        Load Balancer
                               |
                    +----------+----------+
                    |   Nexus API Gateway  |
                    |   (Hertz + Kratos)   |
                    +----------+----------+
                               |
              +----------------+----------------+
              |                |                |
         REST Routes     WebSocket Gateway    gRPC Gateway
              |                |                |
              v                v                v
        Internal gRPC Services (via Kratos)
```

---

## 4. Request Processing Pipeline

Every inbound request passes through the middleware chain:

```
Inbound Request
  → TLS Termination
    → Request ID Generation
      → IP Filtering / DDoS Check
        → CORS Handler
          → Authentication (JWT Validation)
            → Tenant Extraction
              → Rate Limit Check (Redis)
                → Authorization (RBAC/ABAC)
                  → Request Logging
                    → Routing to Backend Service
```

---

## 5. Middleware Stack

### 5.1 Request ID Middleware

| Property | Value |
|---|---|
| Header | `X-Request-ID` |
| Format | UUID v4 |
| Propagation | Injected into all downstream gRPC calls |

### 5.2 Authentication Middleware

| Step | Description |
|---|---|
| Extract | `Authorization: Bearer <jwt>` header |
| Validate | JWT signature, expiry, issuer |
| Decode | Extract `user_id`, `tenant_id`, `roles`, `permissions` |
| Attach | Set in request context for downstream services |
| Skip | Public endpoints (`/api/v1/auth/login`, `/api/v1/auth/register`, `/health`) |

### 5.3 Tenant Middleware

| Step | Description |
|---|---|
| Extract | `tenant_id` from JWT claims |
| Validate | Tenant exists and is active |
| Propagate | `X-Tenant-ID` header to all internal services |
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

### 5.6 Logging Middleware

| Field | Source |
|---|---|
| request_id | Generated |
| tenant_id | JWT |
| user_id | JWT |
| method | Request |
| path | Request |
| status | Response |
| latency_ms | Calculated |
| user_agent | Request header |
| ip | Connection |

---

## 6. Routing Architecture

### 6.1 REST Routes

| Route Pattern | Backend Service | Protocol |
|---|---|---|
| `/api/v1/auth/*` | identity-service | gRPC |
| `/api/v1/ai/chat` | ai-gateway-service | gRPC |
| `/api/v1/ai/stream` | ai-gateway-service | gRPC (streaming) |
| `/api/v1/agents/*` | agent-orchestrator-service | gRPC |
| `/api/v1/rag/*` | rag-service | gRPC |
| `/api/v1/vision/*` | vision-intelligence-service | gRPC |
| `/api/v1/sql/*` | sql-intelligence-service | gRPC |
| `/api/v1/workflows/*` | workflow-service | gRPC |
| `/api/v1/memory/*` | memory-service | gRPC |
| `/api/v1/security/*` | security-ai-service | gRPC |
| `/api/v1/admin/*` | identity-service | gRPC |
| `/health` | Gateway (self) | HTTP |
| `/metrics` | Gateway (self) | HTTP |

### 6.2 WebSocket Routes

| Route | Backend | Purpose |
|---|---|---|
| `wss://host/ws/chat/{conversation_id}` | ai-gateway-service | AI chat streaming |
| `wss://host/ws/agent/{agent_id}` | agent-orchestrator-service | Agent execution streaming |
| `wss://host/ws/notifications` | notification-service | Real-time notifications |

### 6.3 gRPC-Gateway Translation

REST requests are translated to internal gRPC calls via Kratos:

```
HTTP POST /api/v1/auth/login
  → Kratos gRPC-Gateway
    → identity-service.Login() gRPC call
      → HTTP JSON response
```

---

## 7. WebSocket Gateway

### Connection Lifecycle

```
Client connects: wss://host/ws/chat/{conversation_id}
  → Upgrade to WebSocket
    → Authenticate (query param or first message)
    → Validate tenant
    → Create backend gRPC stream
    → Relay messages bidirectionally
    → On disconnect: close gRPC stream, cleanup
```

### Connection Management

| Property | Value |
|---|---|
| Max connections per tenant | 100 |
| Max connections per user | 10 |
| Idle timeout | 5 minutes |
| Message buffer size | 64KB |
| Heartbeat interval | 30 seconds |

### Message Protocol

```json
{
  "type": "message|tool_call|tool_result|error|done",
  "conversation_id": "uuid",
  "data": {},
  "timestamp": "2025-01-15T10:30:00Z"
}
```

---

## 8. Error Response Standard

All errors follow a consistent format:

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Retry after 30 seconds.",
    "request_id": "uuid",
    "timestamp": "2025-01-15T10:30:00Z"
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
| `REQUEST_TIMEOUT` | 504 | Backend service timeout |
| `AI_MODEL_TIMEOUT` | 504 | AI model inference timeout |
| `INTERNAL_ERROR` | 500 | Unexpected server error |
| `SERVICE_UNAVAILABLE` | 503 | Backend service unavailable |

---

## 9. Health Checks

### Endpoint

```
GET /health
```

### Response

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime_seconds": 86400,
  "checks": {
    "redis": "healthy",
    "identity_service": "healthy",
    "agent_orchestrator": "healthy"
  }
}
```

### Check Frequency

| Check | Interval | Timeout |
|---|---|---|
| Self | N/A | N/A |
| Redis | 10s | 2s |
| Downstream services | 15s | 5s |

---

## 10. Load Balancing

### Client → Gateway

| Algorithm | Method |
|---|---|
| External | DNS round-robin or cloud LB |
| Internal | Kubernetes Service (kube-proxy) |

### Gateway → Backend Services

| Method | Description |
|---|---|
| gRPC | Kratos built-in load balancing (round-robin) |
| Health-aware | Unhealthy instances removed from pool |

---

## 11. Security

### TLS Configuration

| Property | Value |
|---|---|
| Protocol | TLS 1.3 only |
| Ciphers | AES-256-GCM, ChaCha20-Poly1305 |
| Certificates | Let's Encrypt or internal CA |
| mTLS | Between gateway and internal services |

### DDoS Protection

| Layer | Protection |
|---|---|
| Network | Cloud LB rate limiting |
| Application | Token bucket per tenant/IP |
| Adaptive | Block IPs with >50 failed auth in 5 min |

### CORS Configuration

```
Allowed Origins: Configured per tenant
Allowed Methods: GET, POST, PUT, DELETE, OPTIONS
Allowed Headers: Authorization, Content-Type, X-Tenant-ID, X-Request-ID
Max Age: 86400 seconds
Credentials: true
```

---

## 12. Observability

### Metrics (Prometheus)

| Metric | Type | Labels |
|---|---|---|
| `gateway_requests_total` | Counter | method, path, status |
| `gateway_request_duration_ms` | Histogram | method, path |
| `gateway_websocket_connections` | Gauge | tenant_id |
| `gateway_rate_limit_rejections_total` | Counter | tenant_id, tier |
| `gateway_auth_failures_total` | Counter | reason |
| `gateway_active_requests` | Gauge | method |
| `gateway_upstream_latency_ms` | Histogram | service, method |

### Tracing (OpenTelemetry)

Every request generates a trace:

```
Trace: gateway.request
  ├── Span: tls_termination
  ├── Span: authentication
  ├── Span: tenant_validation
  ├── Span: rate_limit_check
  ├── Span: authorization
  ├── Span: routing
  └── Span: upstream_call (service_name)
```

### Logging (Loki)

| Field | Description |
|---|---|
| request_id | Unique request identifier |
| trace_id | Distributed trace ID |
| tenant_id | Tenant context |
| user_id | Authenticated user |
| method | HTTP method |
| path | Request path |
| status | Response status code |
| latency_ms | Total request duration |
| upstream_latency_ms | Backend service duration |
| user_agent | Client user agent |
| ip | Client IP address |

---

## 13. Deployment

### Container Configuration

| Property | Value |
|---|---|
| Image | `nexus-api-gateway:latest` |
| Replicas | 3 (minimum) |
| CPU Request | 500m |
| CPU Limit | 2000m |
| Memory Request | 512Mi |
| Memory Limit | 1Gi |
| Port | 8080 (HTTP), 8081 (WS), 50050 (gRPC) |

### Horizontal Scaling

| Metric | Scale Up | Scale Down |
|---|---|---|
| CPU | >70% for 2 min | <30% for 5 min |
| Active requests | >1000 per instance | <100 for 5 min |
| WebSocket connections | >500 per instance | <50 for 5 min |

---

## 14. Configuration

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `GATEWAY_PORT` | HTTP listen port | 8080 |
| `GATEWAY_WS_PORT` | WebSocket listen port | 8081 |
| `GATEWAY_GRPC_PORT` | gRPC listen port | 50050 |
| `JWT_ISSUER` | Expected JWT issuer | nexus-ai |
| `JWT_SECRET` | JWT verification secret | (required) |
| `REDIS_URL` | Redis connection string | redis://localhost:6379 |
| `RATE_LIMIT_ENABLED` | Enable rate limiting | true |
| `CORS_ALLOWED_ORIGINS` | Comma-separated origins | * |
| `LOG_LEVEL` | Logging level | info |
| `OTEL_EXPORTER_URL` | OpenTelemetry endpoint | http://localhost:4317 |

---

## 15. gRPC Contract

### Health Check

```protobuf
service GatewayHealth {
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

### Internal Service Discovery

The gateway discovers backend services via Kubernetes DNS or static configuration:

| Service | gRPC Address |
|---|---|
| identity-service | `identity-service:50051` |
| ai-gateway-service | `ai-gateway-service:50052` |
| agent-orchestrator-service | `agent-orchestrator-service:50053` |
| rag-service | `rag-service:50054` |
| vision-intelligence-service | `vision-intelligence-service:50055` |
| sql-intelligence-service | `sql-intelligence-service:50056` |
| memory-service | `memory-service:50057` |
| workflow-service | `workflow-service:50058` |
| security-ai-service | `security-ai-service:50059` |

---

## 16. NATS Events

### Published

| Subject | When |
|---|---|
| `gateway.request.completed` | After every successful request (for analytics) |
| `gateway.auth.failure` | On authentication failure |
| `gateway.rate_limit.exceeded` | On rate limit rejection |

### Subscribed

| Subject | Purpose |
|---|---|
| `gateway.config.reload` | Dynamic configuration updates |
