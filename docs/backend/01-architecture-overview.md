# AeroXe Nexus AI — Backend Architecture Overview

## Modular Monolithic DDD + Test-Driven Architecture

---

## 1. Platform Identity

| Attribute | Value |
|---|---|
| Product | AeroXe Nexus AI |
| Domain | aeroxenexus.com |
| Category | Enterprise Agentic AI Platform |
| Backend Language | Rust |
| Architecture | **Modular Monolith** (DDD Bounded Contexts → Rust Modules) |
| AI Runtime | Ollama (local GPU inference) |
| Internal Communication | Rust trait interfaces + NATS JetStream (async) |
| External API | REST + WebSocket (via API Gateway module) |
| Deployment | Single binary with optional sidecar services |
| Testing | **TDD-First**: No code without tests |

---

## 2. Architecture Principles

### 2.1 Modular Monolithic Architecture

The entire backend is a **single Rust binary** organized into DDD modules:

```
+-------------------------------------------------------------------+
|                    AeroXe Nexus AI Monolith                        |
|                                                                   |
|  +------------------+  +------------------+  +------------------+  |
|  |   API Gateway    |  |  AI Gateway      |  | Agent Orchestrator| |
|  |   Module         |  |  Module          |  | Module           | |
|  +------------------+  +------------------+  +------------------+  |
|                                                                   |
|  +------------------+  +------------------+  +------------------+  |
|  |   RAG Module     |  |  Vision Module   |  | SQL Agent Module | |
|  +------------------+  +------------------+  +------------------+  |
|                                                                   |
|  +------------------+  +------------------+  +------------------+  |
|  |   Identity Module |  |  Memory Module   |  | Workflow Module  | |
|  +------------------+  +------------------+  +------------------+  |
|                                                                   |
|  +------------------+  +------------------+  +------------------+  |
|  |   Security Module |  |  Audit Module    |  | Integration M.   | |
|  +------------------+  +------------------+  +------------------+  |
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

### 2.2 Domain-Driven Design (DDD)

- **Bounded Contexts** → Rust modules (`crate::domain::*`)
- **Aggregates** → Rust structs with invariant enforcement
- **Entities + Value Objects** → Typed structs with validation
- **Domain Events** → NATS JetStream (async) + in-process (sync)
- **Repository Traits** → `#[async_trait]` with SeaORM implementations

### 2.3 Test-Driven Development (TDD)

**Hard rule:** No production code without automated tests.

```
RED (write failing test) → GREEN (implement) → REFACTOR → INTEGRATE
```

Every module enforces:
- **Unit tests** for domain logic (entities, value objects, aggregates)
- **Integration tests** for repository + infrastructure
- **Module boundary tests** for cross-module contracts
- **API contract tests** for external interfaces

### 2.4 Architecture Decisions

| Decision | Rationale |
|---|---|
| **Single binary** vs microservices | Eliminates gRPC overhead, simplifies deployment, enables stronger compile-time guarantees |
| **Trait-based module interfaces** | Modules communicate through Rust traits — type-safe, testable, mockable |
| **NATS for async events only** | Background jobs, cross-module notifications, audit logging |
| **Schema-per-BoundedContext** | Logical database isolation without physical separation — enables future extraction to microservices |
| **TDD enforced at module boundaries** | Each module has a public API surface + comprehensive test suite |

---

## 3. Backend Technology Stack

### 3.1 Language & Core

| Component | Technology |
|---|---|
| Language | **Rust** (edition 2024) |
| Async Runtime | Tokio (multi-threaded) |
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
| Test Runners | `cargo test`, `cargo nextest` |
| Property Testing | `proptest` crate |
| Fuzzing | `cargo-fuzz` |

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

| Module | Bounded Context | Purpose |
|---|---|---|
| `nexus-gateway` | API Gateway | HTTP/WS server, auth, rate-limit, routing |
| `nexus-ai-gateway` | AI Gateway | AI request lifecycle, prompt safety |
| `nexus-agent` | Agent Orchestration | Agent lifecycle, planning, tool execution |
| `nexus-rag` | RAG Intelligence | Document ingestion, embeddings, search |
| `nexus-vision` | Vision Intelligence | Image analysis, OCR |
| `nexus-sql-agent` | SQL Intelligence | NL→SQL, safe query execution |
| `nexus-security-ai` | Security Intelligence | Code review, threat detection |

### 4.2 Supporting Domain Modules

| Module | Bounded Context | Purpose |
|---|---|---|
| `nexus-identity` | Identity & Auth | IAM, JWT, RBAC/ABAC, tenant mgmt |
| `nexus-memory` | Memory | Short/long-term AI memory |
| `nexus-workflow` | Workflow | Business process automation |
| `nexus-audit` | Audit & Compliance | Full audit trail, compliance |

### 4.3 Infrastructure Modules

| Module | Bounded Context | Purpose |
|---|---|---|
| `nexus-model-registry` | Model Management | Ollama model lifecycle |
| `nexus-notification` | Notifications | Email, WhatsApp, push |
| `nexus-config` | Configuration | Dynamic config, feature flags |
| `nexus-ecosystem` | Ecosystem Integration | AeroXe product connectors |

---

## 5. Module Dependency Flow

```
                   +----------------+
                   | nexus-gateway  |  ← External HTTP/WS
                   +----------------+
                          |
          +---------------+---------------+
          |               |               |
          v               v               v
   +-----------+   +-----------+   +-----------+
   | nexus-    |   | nexus-    |   | nexus-    |
   | identity  |   | ai-gateway|   | model-    |
   +-----------+   +-----------+   | registry  |
          |               |        +-----------+
          |               v
          |        +-----------+
          |        | nexus-    |
          |        | agent     |
          |        +-----------+
          |          |     |     |     |
          v          v     v     v     v
   +-----------+   +-----+ +---+ +---+ +--------+
   | nexus-    |   | RAG | | V. | |SQL| |memory |
   | workflow  |   +-----+ +---+ +---+ +--------+
   +-----------+         |     |     |
          |              v     v     v
          v        +-------------------+
   +-----------+   |   nexus-security  |
   | nexus-    |   +-------------------+
   | notification|          |
   +-----------+          v
                   +-----------+
                   | nexus-    |
                   | audit     |
                   +-----------+
```

---

## 6. Request Flow

### 6.1 AI Chat Request (Modular Monolith)

```
User → HTTP POST /api/v1/ai/chat
  → nexus-gateway::router
    → [Middleware] Auth, Tenant, Rate-Limit
    → nexus-gateway::handlers::ai_chat
      → nexus-ai-gateway::AIGatewayService::submit_request()
        → nexus-agent::AgentService::start_execution()
          → [Planning]   LFM2.5 Thinking (Ollama)
          → [Execution]   Qwen2.5-Coder / Command-R / etc. (Ollama)
          → [Tools]       nexus-rag / nexus-sql-agent / nexus-vision (trait calls)
          → [Memory]      nexus-memory::MemoryService (trait call)
        ← Response (streamed via axum WebSocket)
      → nexus-audit::AuditService::log_event() (NATS + trait)
    ← HTTP 200 + JSON body or WebSocket stream
```

### 6.2 Document Upload (Async via NATS)

```
User → POST /api/v1/rag/documents
  → nexus-gateway → nexus-rag::RagService::upload_document()
    → Store file in MinIO
    → Publish NATS: aeroxe.rag.document.uploaded
    ← HTTP 202 Accepted

[NATS Consumer]
  → nexus-rag::DocumentProcessor (async worker)
    → Parse → Chunk → Embed (Ollama) → Store pgvector
    → Publish NATS: aeroxe.rag.document.processed
    → Update knowledge graph (Apache AGE)
```

---

## 7. Module Internal Structure (DDD Layers)

Every module follows the same layered structure:

```
nexus-<name>/
├── Cargo.toml
├── src/
│   ├── lib.rs                      # Public API (re-exports)
│   ├── domain/
│   │   ├── mod.rs
│   │   ├── aggregates/             # Aggregate roots with invariants
│   │   ├── entities/               # Mutable domain objects
│   │   ├── value_objects/          # Immutable validated types
│   │   └── events/                 # Domain event definitions
│   ├── application/
│   │   ├── mod.rs
│   │   ├── commands/               # CQRS command structs
│   │   ├── queries/                # CQRS query structs
│   │   ├── handlers/               # Command/query handler impls
│   │   └── services/               # Application services (use cases)
│   ├── infrastructure/
│   │   ├── mod.rs
│   │   ├── persistence/            # SeaORM repos + migrations
│   │   ├── ollama/                 # Ollama HTTP client
│   │   └── nats/                   # NATS publisher/subscriber
│   └── interfaces/
│       ├── mod.rs
│       ├── api/                    # Module's internal API traits
│       └── events/                 # NATS event handlers
├── tests/
│   ├── unit/                       # Domain unit tests
│   ├── integration/                # Integration tests (DB, NATS, Ollama)
│   ├── contract/                   # Module boundary contract tests
│   └── e2e/                        # End-to-end flow tests
└── migrations/                     # SeaORM migration files
```

---

## 8. Performance Targets

| Component | Target |
|---|---|
| API Gateway (HTTP serving) | 50,000 req/sec |
| Module trait dispatch | < 1μs overhead |
| PostgreSQL query (indexed) | < 10ms |
| Vector Search (pgvector) | < 200ms |
| Redis Lookup | < 10ms |
| Chat First Token | < 2s |
| RAG Search | < 500ms |
| Vision Analysis | < 5s |

---

## 9. Cross-Cutting Concerns

| Concern | Implementation |
|---|---|
| Authentication | JWT + API Keys (`nexus-identity`) |
| Authorization | RBAC + ABAC Hybrid |
| Multi-Tenancy | `tenant_id` column in all tables |
| Input Validation | `validator` crate + serde deserialization |
| Rate Limiting | Token Bucket (Redis) |
| Observability | OpenTelemetry (metrics, logs, traces via `tracing`) |
| Audit | Every sensitive action logged via `nexus-audit` |
| Secrets | Environment variables + Hashicorp Vault |
| Backup | Daily full + WAL archiving |
| DR | RPO < 15min, RTO < 2hr |
