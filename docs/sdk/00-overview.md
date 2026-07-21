# AeroXe Nexus AI — SDK Overview

## Multi-Language Client SDK Architecture & Common Patterns

---

## 1. SDK Purpose

The AeroXe Nexus AI SDK provides official client libraries for six programming languages, enabling developers to integrate with the AeroXe Nexus AI Platform from any backend technology stack.

### Supported Languages

| Language | Package | Version | Status |
|---|---|---|---|
| Go | `github.com/aeroxe/nexus-sdk-go` | v1.x | Stable |
| Java | `com.aeroxe:nexus-sdk-java` | 1.x | Stable |
| Python | `aeronexus` | 1.x | Stable |
| Rust | `nexus-sdk` | 1.x | Stable |
| Node.js | `@aeronexus/sdk` | 1.x | Stable |
| Elixir | `nexus_sdk` | 1.x | Stable |

---

## 2. Platform Architecture

```
Client Applications
    |
    | HTTPS / WebSocket / gRPC
    v
Nexus API Gateway (Axum + Tonic)
    |
    | Internal gRPC + NATS JetStream
    v
Microservices:
    - identity-service (Auth, Users, Tenants, RBAC)
    - ai-gateway-service (Chat, Streaming)
    - agent-orchestrator-service (Agent Execution)
    - rag-service (Document Upload, Knowledge Search)
    - vision-intelligence-service (Image Analysis, OCR)
    - sql-intelligence-service (Natural Language SQL)
    - memory-service (User Memory)
    - workflow-service (Workflow Automation)
    - security-ai-service (Threat Detection)
    - model-registry-service (AI Models)
    - audit-service (Compliance Logging)
```

---

## 3. SDK Feature Matrix

| Feature | REST API | WebSocket | gRPC | Go | Java | Python | Rust | Node.js | Elixir |
|---|---|---|---|---|---|---|---|---|---|
| Authentication | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Token Refresh | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| AI Chat | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Streaming Chat | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Agent Execution | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| RAG Document Upload | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Knowledge Search | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Vision Analysis | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| OCR | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| SQL Intelligence | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Memory Store/Search | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Workflow Start/Status | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Model List | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| WebSocket Events | - | ✓ | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Retry Logic | - | - | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Rate Limit Handling | - | - | - | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

---

## 4. Common SDK Architecture

All SDKs share a consistent architecture:

```
nexus-sdk/
├── client/              # HTTP/WS client configuration
├── auth/                # Authentication & token management
├── services/            # Service-specific API wrappers
│   ├── ai_chat.go       # AI Chat + Streaming
│   ├── agents.go        # Agent Execution
│   ├── rag.go           # RAG Document & Search
│   ├── vision.go        # Vision & OCR
│   ├── sql.go           # SQL Intelligence
│   ├── memory.go        # Memory Store/Search
│   ├── workflow.go      # Workflow Automation
│   ├── models.go        # Model Management
│   └── admin.go         # Admin Operations
├── models/              # Data models & DTOs
├── websocket/           # WebSocket client
├── errors/              # Error types & handling
├── middleware/           # Request interceptors
└── utils/               # Helper functions
```

---

## 5. Base URL Configuration

| Environment | Base URL |
|---|---|
| Production | `https://api.aeroxenexus.com` |
| Staging | `https://staging-api.aeroxenexus.com` |
| Development | `http://localhost:8080` |
| WebSocket | `wss://api.aeroxenexus.com` |

---

## 6. Authentication

### JWT Token Flow

```
1. Authenticate: POST /api/v1/auth/login
2. Receive: access_token, refresh_token, expires_in
3. Use: Authorization: Bearer <access_token>
4. Refresh: POST /api/v1/auth/refresh (before expiry)
5. Auto-refresh: SDK handles transparently
```

### Token Storage

Each SDK manages token storage differently:

| Language | Default Storage | Secure Option |
|---|---|---|
| Go | In-memory | `keyring` package |
| Java | In-memory | `java.security.KeyStore` |
| Python | In-memory | `keyring` package |
| Rust | In-memory | OS keychain via `keyring` crate |
| Node.js | In-memory | `keytar` package |
| Elixir | In-memory | Agent state |

---

## 7. Request/Response Format

### Standard Request Headers

```http
Authorization: Bearer <jwt>
Content-Type: application/json
X-Request-ID: <uuid>
X-Tenant-ID: <tenant_id>
User-Agent: nexus-sdk-<language>/<version>
```

### Standard Response Headers

```http
X-Request-ID: <uuid>
X-Rate-Limit-Remaining: 499
X-Rate-Limit-Reset: 1700000000
```

### Error Response Format

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

---

## 8. Error Handling

### Error Codes

| Code | HTTP Status | Description | SDK Handling |
|---|---|---|---|
| `UNAUTHORIZED` | 401 | Missing or invalid JWT | Auto-refresh, then re-throw |
| `TOKEN_EXPIRED` | 401 | JWT expired | Auto-refresh, retry once |
| `FORBIDDEN` | 403 | Insufficient permissions | Throw immediately |
| `TENANT_VIOLATION` | 403 | Cross-tenant access | Throw immediately |
| `NOT_FOUND` | 404 | Resource not found | Return null/None |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit hit | Auto-retry with backoff |
| `REQUEST_TIMEOUT` | 504 | Backend timeout | Retry with exponential backoff |
| `AI_MODEL_TIMEOUT` | 504 | AI model timeout | Retry with backoff |
| `INTERNAL_ERROR` | 500 | Server error | Retry up to 3 times |
| `SERVICE_UNAVAILABLE` | 503 | Service down | Retry with circuit breaker |

### Retry Strategy

```
Attempt 1: Immediate
Attempt 2: Wait 1s (exponential backoff)
Attempt 3: Wait 2s
Attempt 4: Wait 4s
Max retries: 4 (configurable)
Jitter: ±20%
```

### Rate Limit Handling

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "retry_after": 30
  }
}
```

SDKs automatically:
1. Parse `retry_after` from response
2. Sleep for the specified duration
3. Retry the request
4. Respect `X-Rate-Limit-Remaining` header

---

## 9. WebSocket Protocol

### Connection

```
wss://api.aeroxenexus.com/ws/chat/{conversation_id}
```

### Client → Server Messages

```json
{ "type": "message", "content": "Analyze my broadband issue" }
{ "type": "ping" }
```

### Server → Client Messages

```json
{ "type": "token", "content": "Customer" }
{ "type": "token", "content": " network" }
{ "type": "tool_call", "content": "customer.lookup()" }
{ "type": "tool_result", "content": "{ ... }" }
{ "type": "completed" }
{ "type": "error", "content": "Model timeout" }
{ "type": "pong" }
```

### WebSocket Connection Options

| Option | Default | Description |
|---|---|---|
| `reconnect` | `true` | Auto-reconnect on disconnect |
| `max_reconnect_attempts` | `5` | Max reconnection attempts |
| `heartbeat_interval` | `30s` | Ping/pong interval |
| `message_buffer_size` | `64KB` | Max message size |
| `idle_timeout` | `5m` | Connection idle timeout |

---

## 10. Pagination

### Offset-Based Pagination

```http
GET /api/v1/agents?limit=10&offset=0
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "total": 150,
    "limit": 10,
    "offset": 0,
    "has_more": true
  }
}
```

### Cursor-Based Pagination (for streaming)

```http
GET /api/v1/memory/search?q=customer&limit=10&cursor=abc123
```

**Response:**
```json
{
  "data": [...],
  "cursor": "next_cursor_value",
  "has_more": true
}
```

---

## 11. Rate Limits

| Tier | Limit | Burst |
|---|---|---|
| Free | 100 AI requests/min | 10 |
| Customer | 500 AI requests/min | 50 |
| Enterprise | 10,000 AI requests/min | 500 |
| Admin | Unlimited | Unlimited |

### SDK Rate Limit Headers

```http
X-Rate-Limit-Limit: 500
X-Rate-Limit-Remaining: 499
X-Rate-Limit-Reset: 1700000000
```

---

## 12. API Endpoint Reference

### Authentication

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/auth/login` | Login with email/password |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Invalidate session |

### AI Chat

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/ai/chat` | Send chat message |
| WS | `/ws/chat/{conversation_id}` | Streaming chat |

### Agents

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/agents/execute` | Execute agent task |
| GET | `/api/v1/agents/execution/{id}` | Get execution status |
| GET | `/api/v1/agents` | List available agents |
| GET | `/api/v1/agents/{id}` | Get agent details |

### RAG / Knowledge

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/rag/documents` | Upload document |
| GET | `/api/v1/rag/documents/{id}/status` | Get processing status |
| GET | `/api/v1/rag/documents` | List documents |
| DELETE | `/api/v1/rag/documents/{id}` | Delete document |
| POST | `/api/v1/rag/search` | Search knowledge base |

### Vision

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/vision/analyze` | Analyze image |
| POST | `/api/v1/vision/ocr` | Extract text from image |

### SQL Intelligence

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/sql/query` | Natural language SQL query |
| GET | `/api/v1/sql/databases` | List available databases |

### Memory

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/memory` | Store memory |
| GET | `/api/v1/memory/search` | Search memory |
| DELETE | `/api/v1/memory/{id}` | Delete memory |

### Workflow

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/workflows/start` | Start workflow |
| GET | `/api/v1/workflows/{id}` | Get workflow status |
| GET | `/api/v1/workflows` | List workflows |

### Models

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/models` | List available models |

### Admin

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/admin/users` | List users |
| POST | `/api/v1/admin/users` | Create user |
| PUT | `/api/v1/admin/users/{id}` | Update user |
| DELETE | `/api/v1/admin/users/{id}` | Delete user |
| GET | `/api/v1/admin/tenants` | List tenants |
| POST | `/api/v1/admin/tenants` | Create tenant |

---

## 13. Performance Requirements

| Operation | Target Latency |
|---|---|
| Authentication | < 200ms |
| Chat First Token | < 2s |
| RAG Search | < 500ms |
| Agent Start | < 300ms |
| Vision Request | < 5s |
| SQL Query | < 3s |
| Memory Search | < 200ms |
| Workflow Start | < 300ms |

---

## 14. Security Requirements

| Requirement | Implementation |
|---|---|
| TLS 1.3 | All connections |
| JWT Validation | Signature, expiry, issuer |
| Token Rotation | Refresh tokens rotate on use |
| Secure Storage | OS keychain where available |
| Certificate Pinning | Optional (recommended for mobile) |
| Request Signing | HMAC-SHA256 for sensitive operations |
| Audit Logging | All API calls logged server-side |

---

## 15. SDK Installation

### Go

```bash
go get github.com/aeroxe/nexus-sdk-go@latest
```

### Java

```xml
<dependency>
    <groupId>com.aeroxe</groupId>
    <artifactId>nexus-sdk-java</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Python

```bash
pip install aeronexus
```

### Rust

```toml
[dependencies]
nexus-sdk = "1.0"
```

### Node.js

```bash
npm install @aeronexus/sdk
# or
yarn add @aeronexus/sdk
# or
pnpm add @aeronexus/sdk
```

### Elixir

```elixir
# mix.exs
{:nexus_sdk, "~> 1.0"}
```

---

## 16. Quick Start

### Go

```go
client := nexus.NewClient(
    nexus.WithBaseURL("https://api.aeroxenexus.com"),
    nexus.WithAPIKey("your-api-key"),
)

response, err := client.AI.Chat(&nexus.ChatRequest{
    Message: "Hello, Nexus!",
    Agent:   "general",
})
```

### Java

```java
NexusClient client = NexusClient.builder()
    .baseUrl("https://api.aeroxenexus.com")
    .apiKey("your-api-key")
    .build();

ChatResponse response = client.ai().chat(
    ChatRequest.builder()
        .message("Hello, Nexus!")
        .agent("general")
        .build()
);
```

### Python

```python
from aeronexus import NexusClient

client = NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
)

response = client.ai.chat(
    message="Hello, Nexus!",
    agent="general"
)
```

### Rust

```rust
use nexus_sdk::{Client, ChatRequest};

let client = Client::builder()
    .base_url("https://api.aeroxenexus.com")
    .api_key("your-api-key")
    .build()?;

let response = client.ai().chat(&ChatRequest {
    message: "Hello, Nexus!".to_string(),
    agent: Some("general".to_string()),
}).await?;
```

### Node.js

```javascript
import { NexusClient } from '@aeronexus/sdk';

const client = new NexusClient({
  baseUrl: 'https://api.aeroxenexus.com',
  apiKey: 'your-api-key',
});

const response = await client.ai.chat({
  message: 'Hello, Nexus!',
  agent: 'general',
});
```

### Elixir

```elixir
alias NexusSdk.Client

{:ok, client} = Client.new(
  base_url: "https://api.aeroxenexus.com",
  api_key: "your-api-key"
)

{:ok, response} = Client.ai_chat(client, %{
  message: "Hello, Nexus!",
  agent: "general"
})
```
