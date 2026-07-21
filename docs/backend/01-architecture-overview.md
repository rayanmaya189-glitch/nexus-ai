# AeroXe Nexus AI — Backend Architecture Overview

## Modular Monolithic DDD + Test-Driven Architecture

---

## 1. Platform Identity

| Attribute | Value |
|---|---|
| Product | AeroXe Nexus AI |
| Domain | aeroxenexus.com |
| Category | Enterprise Agentic AI Platform |
| Backend Language | **Rust** (edition 2024) |
| Architecture | **Modular Monolith** (DDD Bounded Contexts → Rust modules under `src/modules/`) |
| ORM | **SeaORM** (no raw SQL — all database access through SeaORM entities & models) |
| AI Runtime | Ollama (local GPU inference) |
| Internal Communication | Rust trait interfaces + NATS JetStream (async) |
| External API | REST + WebSocket (via API Gateway module) |
| API Versioning | URL-based (`/api/v1/`, `/api/v2/`, etc.) |
| NATS Versioning | Subject-based (`aeroxe.v1.module.event`) |
| gRPC Versioning | Service-level versioning (`identity.v1.AuthService`) |
| Deployment | Single binary with optional sidecar services |
| Testing | **TDD-First**: No code without tests |

---

## 2. Architecture Principles

### 2.1 Modular Monolithic Architecture

The entire backend is a **single Rust binary** organized into DDD modules under `src/modules/`:

```
+-------------------------------------------------------------------+
|                    AeroXe Nexus AI Monolith                        |
|                                                                   |
|  src/modules/                                                     |
|  +------------------+  +------------------+  +------------------+ |
|  |   gateway        |  |   ai-gateway     |  |   agent          | |
|  |                  |  |                  |  |                  | |
|  +------------------+  +------------------+  +------------------+ |
|                                                                   |
|  +------------------+  +------------------+  +------------------+ |
|  |   rag            |  |   vision         |  |   sql-agent      | |
|  +------------------+  +------------------+  +------------------+ |
|                                                                   |
|  +------------------+  +------------------+  +------------------+ |
|  |   identity       |  |   customer       |  |   memory         | |
|  +------------------+  +------------------+  +------------------+ |
|                                                                   |
|  +------------------+  +------------------+  +------------------+ |
|  |   workflow       |  |   security       |  |   audit          | |
|  +------------------+  +------------------+  +------------------+ |
+-------------------------------------------------------------------+
                              |
        +---------------------------------------------------+
        |              Shared Infrastructure                 |
        |  PostgreSQL | Redis | NATS | MinIO | Elasticsearch |
        +---------------------------------------------------+
                              |
        +---------------------------------------------------+
        |              AI Compute (Ollama)                   |
        |  LFM | Hermes3 | Phi4 | Qwen | Command-R | Llama  |
        +---------------------------------------------------+
```

### 2.2 Folder Structure

Every module follows a strict DDD layered structure within `src/modules/<name>/`:

```
src/
├── main.rs                        # Binary entry point
├── lib.rs                         # Common library utilities
├── config/                        # Application configuration
│   ├── database.rs                # SeaORM database config
│   ├── redis.rs                   # Redis connection config
│   ├── nats.rs                    # NATS connection config
│   └── settings.rs                # Global settings (env-based)
│
└── modules/                       # Bounded contexts (business domains)
    ├── gateway/                   # API Gateway layer
    │   ├── mod.rs
    │   ├── auth/
    │   │   ├── jwt_validator.rs
    │   │   └── api_key_validator.rs
    │   ├── rate_limiter/
    │   │   └── token_bucket.rs
    │   ├── request_validator/
    │   │   └── schemas/
    │   ├── api_versioning/
    │   │   └── router.rs
    │   └── tests/
    │
    ├── identity/                  # Authentication & authorisation
    │   ├── domain/
    │   │   ├── aggregates/user/
    │   │   ├── entities/session.rs
    │   │   ├── value_objects/
    │   │   └── rules/
    │   ├── application/
    │   │   ├── commands/
    │   │   ├── queries/
    │   │   └── services/
    │   ├── infrastructure/
    │   │   ├── repository/        # SeaORM repositories
    │   │   └── security/          # JWT, hashing
    │   ├── api/
    │   │   ├── http/
    │   │   └── grpc/
    │   └── migrations/            # SeaORM migration files
    │
    ├── customer/                  # Customer aggregate, KYC, addresses
    │   ├── domain/
    │   │   ├── aggregates/customer/
    │   │   ├── value_objects/
    │   │   └── rules/
    │   ├── application/
    │   │   ├── commands/
    │   │   ├── queries/
    │   │   └── services/
    │   ├── infrastructure/
    │   │   ├── repository/        # SeaORM repositories
    │   │   └── messaging/         # NATS publishers/subscribers
    │   ├── api/
    │   │   ├── http/
    │   │   └── grpc/
    │   └── migrations/            # SeaORM migration files
    │
    ├── ai-gateway/                # AI Gateway
    ├── agent/                     # Agent Orchestration
    ├── rag/                       # RAG Intelligence
    ├── vision/                    # Vision Intelligence
    ├── sql-agent/                 # SQL Intelligence
    ├── memory/                    # Memory
    ├── workflow/                  # Workflow
    ├── security/                  # Security Intelligence
    ├── audit/                     # Audit & Compliance
    ├── notification/              # Notifications
    ├── model-registry/            # Model Management
    ├── config/                    # Dynamic Configuration
    └── ecosystem/                 # Ecosystem Integration
```

### 2.3 Domain-Driven Design (DDD)

- **Bounded Contexts** → Rust modules (`src/modules/<name>/`)
- **Aggregates** → Rust structs with invariant enforcement
- **Entities + Value Objects** → Typed structs with validation (`validator` crate)
- **Domain Events** → NATS JetStream (async) + in-process (sync)
- **Repository Traits** → `#[async_trait]` with **SeaORM** implementations
- **No Raw SQL** → All database access through SeaORM entity models

### 2.4 Test-Driven Development (TDD)

**Hard rule:** No production code without automated tests.

```
RED (write failing test) → GREEN (implement) → REFACTOR → INTEGRATE
```

Every module enforces:
- **Unit tests** for domain logic (entities, value objects, aggregates)
- **Integration tests** for repository + infrastructure (SeaORM + test DB)
- **Module boundary tests** for cross-module contracts
- **API contract tests** for external interfaces

### 2.5 Architecture Decisions

| Decision | Rationale |
|---|---|
| **Single binary** vs microservices | Eliminates gRPC overhead, simplifies deployment, enables stronger compile-time guarantees |
| **SeaORM** over raw SQL | Type-safe queries, migration tooling, compile-time checked schemas |
| **No raw SQL anywhere** | All DB access through SeaORM entity models — prevents SQL injection, enables schema migrations |
| **Schema-per-BoundedContext** | Logical database isolation without physical separation — enables future extraction to microservices |
| **Trait-based module interfaces** | Modules communicate through Rust traits — type-safe, testable, mockable |
| **NATS for async events only** | Background jobs, cross-module notifications, audit logging |
| **API versioning in URL** | `/api/v1/` prefix — clear, cacheable, easy to route |
| **NATS subject versioning** | `aeroxe.v1.module.event` — prevents event format conflicts |
| **gRPC service versioning** | `package.v1.ServiceName` — supports multiple API versions |
| **TDD enforced at module boundaries** | Each module has a public API surface + comprehensive test suite |

---

## 3. Backend Technology Stack

### 3.1 Language & Core

| Component | Technology |
|---|---|
| Language | **Rust** (edition 2024) |
| Async Runtime | Tokio (multi-threaded) |
| ORM | **SeaORM** (with `sea-orm` crate, migration tooling) |
| HTTP / WS Server | **axum** |
| gRPC (external only) | tonic (optional, for SDK/partner integrations) |
| Serialization | serde + serde_json |
| Configuration | environment-based + config files |
| Logging | tracing + OpenTelemetry |

### 3.2 DDD / Testing Libraries

| Component | Technology |
|---|---|
| Async Traits | `async-trait` crate |
| Mocking | `mockall` crate |
| Validation | `validator` crate (derive macros for struct validation) |
| Test Runners | `cargo test`, `cargo nextest` |
| Property Testing | `proptest` crate |

### 3.3 Infrastructure

| Component | Technology |
|---|---|
| Primary Database | **PostgreSQL 18** (single cluster, schema-per-module) |
| Vector Search | pgvector (PostgreSQL extension) |
| Knowledge Graph | Apache AGE (PostgreSQL extension) |
| Cache / STM | Redis |
| Event Bus | NATS JetStream |
| File Storage | MinIO |
| Full-Text Search | Elasticsearch |
| AI Runtime | Ollama HTTP API |

### 3.4 Module Communication Matrix

| Caller → Callee | Synchronous | Asynchronous |
|---|---|---|
| API Gateway → any module | `trait` method call | NATS for long ops |
| AI Gateway → Agent | `trait` method call | — |
| Agent Orchestrator → RAG | `trait` method call | NATS for ingestion |
| Agent Orchestrator → Memory | `trait` method call | — |
| Agent Orchestrator → SQL | `trait` method call | — |
| Agent Orchestrator → Vision | `trait` method call | — |
| Workflow → any module | `trait` method call | NATS for step events |
| Audit ← any module | `trait` method call | NATS for fire-and-forget |

---

## 4. Module Catalogue (Bounded Contexts)

### 4.1 Core Domain Modules

| Module | Bounded Context | Schema Prefix | Purpose |
|---|---|---|---|
| `gateway` | API Gateway | — (stateless, Redis) | HTTP/WS server, auth, rate-limit, routing |
| `ai-gateway` | AI Gateway | `ai_` | AI request lifecycle, prompt safety |
| `agent` | Agent Orchestration | `agent_` | Agent lifecycle, planning, tool execution |
| `rag` | RAG Intelligence | `rag_` | Document ingestion, embeddings, search |
| `vision` | Vision Intelligence | `vision_` | Image analysis, OCR |
| `sql-agent` | SQL Intelligence | `sql_` | NL→SQL, safe query execution |
| `security` | Security Intelligence | `security_` | Code review, threat detection |
| **`telephony`** | **Telephony** | **`telephony_`** | **Voice channel, SIP/WebRTC, call management, audio streaming, caller auth, anti-fraud, voicemail, IVR, audio quality, live monitoring** |
| **`conversation`** | **Conversation Management** | **`conversation_`** | **Conversation state machine, context, turn-taking, flow control, SLA management, sentiment tracking, GDPR** |
| **`stt`** | **Speech Processing** | **`stt_`** | **Speech-to-Text, real-time transcription, audio processing, confidence thresholds, anti-injection** |
| **`tts`** | **Speech Processing** | **`tts_`** | **Text-to-Speech, voice synthesis, voice personalization, voice cloning, sentiment adaptation, post-call survey** |

### 4.2 Supporting Domain Modules

| Module | Bounded Context | Schema Prefix | Purpose |
|---|---|---|---|
| `identity` | Identity & Auth | `identity_` | IAM, JWT, RBAC/ABAC, tenant mgmt, KYC |
| `customer` | Customer Management | `customer_` | Customer aggregate, profiles, status, addresses |
| `memory` | Memory | `memory_` | Short/long-term AI memory |
| `workflow` | Workflow | `workflow_` | Business process automation |
| `audit` | Audit & Compliance | `audit_` | Full audit trail, compliance |
| **`analytics`** | **Analytics & Intelligence** | **`analytics_`** | **Conversation analytics, AI performance, business intelligence, call center metrics, cost allocation, agent scoring** |
| **`webhook`** | **Integration** | **`webhook_`** | **Outbound webhooks, event delivery, retry logic** |
| **`outbound`** | **Outbound Communication** | **`outbound_`** | **Proactive AI calls, campaign management, DNC compliance** |

### 4.3 Infrastructure Modules

| Module | Bounded Context | Schema Prefix | Purpose |
|---|---|---|---|
| `model-registry` | Model Management | `models_` | Ollama model lifecycle |
| `notification` | Notifications | `notif_` | Email, WhatsApp, push |
| `config` | Configuration | `config_` | Dynamic config, feature flags |
| `ecosystem` | Ecosystem Integration | `eco_` | AeroXe product connectors |

---

## 5. Module Dependency Flow

```
                   +----------------+
                   |   gateway      |  ← External HTTP/WS + Telephony Webhooks
                   +----------------+
                          |
          +---------------+---------------+
          |               |               |
          v               v               v
   +-----------+   +-----------+   +-----------+
   | identity  |   | ai-gateway|   | model-    |
   +-----------+   +-----------+   | registry  |
          |               |        +-----------+
          |               v
          |        +-----------+
          |        |  agent    |
          +-------------------+
          |          |     |     |     |
          v          v     v     v     v
   +-----------+   +-----+ +---+ +---+ +--------+
   | customer  |   | RAG | | V. | |SQL| | memory |
   +-----------+   +-----+ +---+ +---+ +--------+
   | workflow  |         |     |     |
   +-----------+         v     v     v
          |        +-------------------+
          v        |   security        |
   +-----------+   +-------------------+
   |notifictn  |          |
   +-----------+          v
                   +-----------+
                   |  audit    |
                   +-----------+

   +=============================================================+
   |               Voice / Telephony Channel                      |
   +=============================================================+
   +-----------+   +-----------+   +-----------+   +-----------+
   |telephony  |-->|   stt     |-->|  agent    |-->|   tts     |
   |(SIP/WebRTC|   |(Speech-to |   |(AI Process|   |(Text-to   |
   | Call Mgmt)|   | Text)     |   | ing)      |   | Speech)   |
   +-----------+   +-----------+   +-----------+   +-----------+
          |                            |                |
          v                            v                v
   +-----------+              +-----------+      +-----------+
   |outbound   |              |conversation|     | analytics |
   |(Campaigns |              |(State     |      |(Metrics,  |
   | Proactive)|              | Machine)  |      | Reports)  |
   +-----------+              +-----------+      +-----------+
          |                                                |
          v                                                v
   +-----------+                                  +-----------+
   |  webhook  |                                  |  audit    |
   |(Event     |                                  |(Compliance|
   | Delivery) |                                  | Logging)  |
   +-----------+                                  +-----------+
```

---

## 6. Versioning Standards

### 6.1 API Versioning (REST)

All routes follow `/api/v{version}/<resource>` pattern.

```
/api/v1/auth/login
/api/v1/customers/{id}
/api/v2/customers/{id}         # Future: different response schema
```

Version is extracted by `api_versioning::router` in the gateway module and propagated to all trait calls via `RequestContext`.

### 6.2 NATS Event Versioning

All NATS subjects follow `aeroxe.v{version}.<module>.<event>` pattern.

```
aeroxe.v1.identity.user.created
aeroxe.v1.customer.customer.created
aeroxe.v2.customer.customer.created     # Future: different event schema
```

### 6.3 gRPC Service Versioning

gRPC service names include the version:

```protobuf
package identity.v1;
service AuthService { ... }

package customer.v1;
service CustomerService { ... }
```

### 6.4 Database Schema Versioning

Each module's schema uses SeaORM migrations with sequential versioning:

```rust
// migrations/identity/
// m20250701_000001_create_users_table.rs
// m20250701_000002_create_roles_table.rs
```

Every migration is versioned and reversible via SeaORM's migration system.

---

## 7. Performance Targets

| Component | Target |
|---|---|
| API Gateway (HTTP serving) | 50,000 req/sec |
| Module trait dispatch | < 1μs overhead |
| PostgreSQL query (SeaORM, indexed) | < 10ms |
| Vector Search (pgvector) | < 200ms |
| Redis Lookup | < 10ms |
| Chat First Token | < 2s |
| RAG Search | < 500ms |
| Vision Analysis | < 5s |

---

## 8. Cross-Cutting Concerns

| Concern | Implementation |
|---|---|
| Authentication | JWT + API Keys (`identity` module) |
| Authorization | RBAC + ABAC Hybrid |
| Multi-Tenancy | `tenant_id` column in all tables (SeaORM entity filters) |
| Input Validation | `validator` crate + serde deserialization |
| Rate Limiting | Token Bucket (Redis) |
| Observability | OpenTelemetry (metrics, logs, traces via `tracing`) |
| Audit | Every sensitive action logged via `audit` module |
| Secrets | Environment variables + Hashicorp Vault |
| No Raw SQL | All DB access through SeaORM entities and models |
| Backup | Daily full + WAL archiving |
| DR | RPO < 15min, RTO < 2hr |
