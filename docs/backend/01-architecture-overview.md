# AeroXe Nexus AI — Backend Architecture Overview

## System-Level Backend Documentation

---

## 1. Platform Identity

| Attribute | Value |
|---|---|
| Product | AeroXe Nexus AI |
| Domain | aeroxenexus.com |
| Category | Enterprise Agentic AI Platform |
| Backend Languages | Rust + Go |
| AI Runtime | Ollama (local GPU inference) |
| Communication | gRPC + Protocol Buffers + NATS JetStream |
| External API | REST + WebSocket + gRPC Gateway |
| Deployment | Private Infrastructure First |

---

## 2. System Architecture Diagram

```
                         Users / Applications
                                |
                        Nexus API Gateway
                                |
                  AI Agent Orchestrator
                                |
                        NATS JetStream
                                |
  ============================================================
                     AI Microservices
  ============================================================
    identity-service        ai-gateway-service
    agent-orchestrator-service    rag-service
    vision-service          sql-agent-service
    memory-service          workflow-service
    security-ai-service     audit-service
    notification-service    model-registry-service
    configuration-service   ecosystem-integration-service
  ============================================================
                                |
                        Model Router
                                |
  ============================================================
                     Ollama Runtime
  ============================================================
    lfm2.5-thinking:1.2b    hermes3:3b
    phi4-mini:3.8b           qwen2.5-coder:3b
    qwen3-vl:4b              command-r7b:7b
    llama3.1:7b              whiterabbitneo:7b
  ============================================================
                                |
  ============================================================
                   Intelligence Platform
  ============================================================
    RAG Engine    SQL Agent    Knowledge Graph
    Memory System    Workflow Engine
  ============================================================
```

---

## 3. Architecture Principles

### 3.1 Domain-Driven Design (DDD)

The platform is divided into independent bounded contexts. Each bounded context owns its business logic, database, and exposes contracts through gRPC while publishing events via NATS JetStream.

### 3.2 Test-Driven Development (TDD)

No production code exists without automated tests. The cycle: Write Test -> Implement -> Refactor -> Integration Test -> Production.

### 3.3 Microservice Architecture

Each service:
- Owns its database (Database-per-Microservice pattern)
- Owns its business rules
- Communicates through versioned contracts (gRPC)
- Deploys independently
- Has independent scaling lifecycle

### 3.4 Hybrid Communication Strategy

| Layer | Technology |
|---|---|
| External (Mobile/Web) | HTTPS REST + WebSocket |
| Internal Synchronous | gRPC + Protocol Buffers |
| Internal Asynchronous | NATS JetStream |
| AI Runtime | Ollama HTTP API |
| Database Access | Repository Pattern |

---

## 4. Backend Technology Stack

### 4.1 Languages

| Language | Usage | Rationale |
|---|---|---|
| Rust | Core AI services, performance-critical paths | Memory safety, zero-cost abstractions, async runtime (Tokio) |
| Go | Infrastructure services, API gateway, integration services | Concurrency model, fast compilation, ecosystem |

### 4.2 Core Libraries

| Component | Rust | Go |
|---|---|---|
| gRPC | tonic | kratos |
| NATS | async-nats | nats.go |
| Database | SeaORM | EntORM |
| Vector | pgvector (SeaORM) | pgvector (EntORM) |
| Serialization | serde + serde_json | encoding/json |
| Async Runtime | Tokio | goroutines |
| HTTP | axum / actix-web | hertz |
| WebSocket | tokio-tungstenite | coder/websocket |

### 4.3 Infrastructure

| Component | Technology |
|---|---|
| Container Runtime | Docker |
| Orchestration | Kubernetes |
| CI/CD | GitLab CI |
| Monitoring | OpenTelemetry + Prometheus + Grafana |
| Logging | Loki |
| Tracing | Tempo |
| Object Storage | MinIO |
| Search | Elasticsearch |
| Cache | Redis |
| Graph | Apache AGE (PostgreSQL extension) |
| Vector Search | pgvector (PostgreSQL extension) |
| Primary Database | PostgreSQL 18 |

---

## 5. Service Catalogue

### 5.1 Core Domain Services

| Service | Language | Purpose |
|---|---|---|
| `ai-gateway-service` | Go | Central AI request processing, routing, rate limiting |
| `agent-orchestrator-service` | Rust | Agent lifecycle, planning, tool execution |
| `rag-service` | Rust | Document ingestion, embeddings, hybrid search |
| `vision-service` | Rust | Image processing, OCR, visual reasoning |
| `sql-agent-service` | Go | Natural language SQL generation and execution |
| `security-ai-service` | Rust | Security analysis, vulnerability detection |

### 5.2 Supporting Domain Services

| Service | Language | Purpose |
|---|---|---|
| `identity-service` | Go | Users, roles, permissions, tenant management |
| `memory-service` | Rust | Short-term and long-term AI memory |
| `workflow-service` | Go | Business automation, approvals, task management |
| `audit-service` | Go | Compliance tracking, activity logging |

### 5.3 Infrastructure Services

| Service | Language | Purpose |
|---|---|---|
| `model-registry-service` | Go | Ollama model management, routing |
| `notification-service` | Go | Push, email, SMS, WhatsApp notifications |
| `configuration-service` | Go | Feature flags, dynamic configuration |
| `ecosystem-integration-service` | Go | AeroXe product connectors |

---

## 6. Request Flow

### 6.1 Synchronous Request (gRPC)

```
User Request
    |
    v
API Gateway (REST/WS)
    |
    v
AI Gateway Service
    |
    v
Agent Orchestrator (gRPC)
    |
    v
RAG Service / Vision Service / SQL Service (gRPC)
    |
    v
Ollama (HTTP)
    |
    v
Response (streamed back)
```

### 6.2 Asynchronous Request (NATS JetStream)

```
Document Uploaded
    |
    v
nexus.rag.document.created (NATS Event)
    |
    v
RAG Worker
    |
    v
Chunk Generation
    |
    v
Embedding Creation
    |
    v
Vector Store Update
    |
    v
nexus.rag.embedding.created (NATS Event)
```

---

## 7. Deployment Model

### 7.1 Private Infrastructure First

The platform is designed to:
- Run offline after model download
- Support local GPU inference via Ollama
- Scale AI workloads independently
- Maintain enterprise security (Zero Trust)
- Enable future cloud/hybrid deployment

### 7.2 Kubernetes Namespaces

```
aeroxe-system          - System services
aeroxe-ai              - AI microservices
aeroxe-data            - PostgreSQL, Redis, NATS, MinIO, Elasticsearch
aeroxe-monitoring      - Prometheus, Grafana, Loki, Tempo
aeroxe-gpu             - Ollama GPU nodes
```

### 7.3 Environment Separation

```
Development  ->  Testing  ->  Staging  ->  Production
```

---

## 8. Performance Targets

| Component | Target |
|---|---|
| API Gateway | 50,000 req/sec |
| gRPC Internal | 100,000 req/sec |
| Vector Search | < 200ms |
| SQL Query | < 2s |
| Redis Lookup | < 10ms |
| PostgreSQL API Query | < 100ms |
| Elasticsearch Search | < 300ms |
| Chat First Token | < 2s |
| RAG Search | < 500ms |
| Vision Request | < 5s |

---

## 9. Cross-Cutting Concerns

| Concern | Implementation |
|---|---|
| Authentication | JWT + API Keys |
| Authorization | RBAC + ABAC Hybrid |
| Multi-Tenancy | tenant_id column in all tables |
| Encryption at Rest | AES-256 |
| Encryption in Transit | TLS 1.3 + mTLS |
| Secrets | Hashicorp Vault / K8s Secrets |
| Audit | Every sensitive action logged |
| Rate Limiting | Token Bucket (Redis) |
| Observability | OpenTelemetry (metrics, logs, traces) |
| Backup | Daily full + WAL archiving |
| DR | RPO < 15min, RTO < 2hr |
