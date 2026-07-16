# AeroXe Nexus AI — Communication Architecture

## gRPC, Protocol Buffers, NATS JetStream & Event-Driven Design

---

## 1. Communication Strategy Overview

AeroXe Nexus AI follows a hybrid communication architecture:

```
                     External World
                           |
                  REST / WebSocket / HTTPS
                           |
                     API Gateway
                           |
  ================================================
                 Internal Communication
  ================================================
        Synchronous              Asynchronous
            gRPC                  NATS JetStream
             |                         |
     Service-to-Service        Event Driven Flow
  ================================================
```

---

## 2. Communication Rules

| Layer | Protocol | Usage |
|---|---|---|
| External (Web/Mobile) | HTTPS REST | Client API |
| Real-time Chat | WebSocket | Token streaming |
| Internal Synchronous | gRPC + Protobuf | Service-to-service |
| Internal Asynchronous | NATS JetStream | Events, background jobs |
| AI Runtime | Ollama HTTP API | Model inference |

---

## 3. gRPC Architecture

### Design Principles

Every gRPC service must have:
- Versioned protobuf contracts (package `aeroxe.{service}`)
- Backward compatibility (field deprecation, not removal)
- Strong typing through Protocol Buffers
- Standardized error codes
- Authentication metadata propagation

### Required Metadata

Every gRPC request must include:

```text
authorization     - JWT token
tenant-id         - Tenant ID
user-id           - User ID
request-id        - Request UUID for tracing
trace-id          - Distributed trace ID
```

---

## 4. Proto Repository Structure

```
aeroxe-proto/
├── common/
│   └── common.proto
├── identity/
│   └── identity.proto
├── agent/
│   └── agent.proto
├── rag/
│   └── rag.proto
├── vision/
│   └── vision.proto
├── sql/
│   └── sql.proto
├── memory/
│   └── memory.proto
├── workflow/
│   └── workflow.proto
├── security/
│   └── security.proto
├── audit/
│   └── audit.proto
└── gateway/
    └── gateway.proto
```

---

## 5. Common Proto Definitions

### common/common.proto

```protobuf
syntax = "proto3";
package aeroxe.common;

message RequestContext {
  string request_id = 1;
  string tenant_id = 2;
  string user_id = 3;
  string trace_id = 4;
}

message ErrorResponse {
  string code = 1;
  string message = 2;
  string details = 3;
  string request_id = 4;
}

enum ErrorCode {
  UNKNOWN = 0;
  INVALID_REQUEST = 1;
  UNAUTHORIZED = 2;
  FORBIDDEN = 3;
  NOT_FOUND = 4;
  TIMEOUT = 5;
  MODEL_ERROR = 6;
  DATABASE_ERROR = 7;
}
```

---

## 6. Service-to-Service gRPC Flows

### Agent -> RAG

```protobuf
service RagService {
  rpc SearchKnowledge(SearchRequest) returns (SearchResponse);
}
```

### Agent -> Vision

```protobuf
service VisionService {
  rpc AnalyzeImage(ImageRequest) returns (ImageAnalysisResponse);
}
```

### Agent -> SQL

```protobuf
service SQLService {
  rpc GenerateQuery(QueryRequest) returns (SQLResponse);
  rpc ExecuteQuery(SQLRequest) returns (ResultResponse);
}
```

### Agent -> Memory

```protobuf
service MemoryService {
  rpc StoreMemory(StoreMemoryRequest) returns (MemoryResponse);
  rpc SearchMemory(SearchMemoryRequest) returns (MemoryList);
}
```

---

## 7. NATS JetStream Architecture

### Subject Naming Standard

Format: `aeroxe.<domain>.<event>`

### All Subjects

| Subject | Domain | Description |
|---|---|---|
| `aeroxe.ai.request.created` | AI Gateway | New AI request |
| `aeroxe.ai.response.generated` | AI Gateway | Response ready |
| `aeroxe.ai.failed` | AI Gateway | Request failed |
| `aeroxe.agent.started` | Agent | Agent execution started |
| `aeroxe.agent.completed` | Agent | Agent execution done |
| `aeroxe.agent.failed` | Agent | Agent execution failed |
| `aeroxe.agent.tool.executed` | Agent | Tool call made |
| `aeroxe.rag.document.uploaded` | RAG | Document received |
| `aeroxe.rag.document.processed` | RAG | Processing done |
| `aeroxe.rag.embedding.created` | RAG | Embeddings stored |
| `aeroxe.vision.image.received` | Vision | Image uploaded |
| `aeroxe.vision.analysis.completed` | Vision | Analysis done |
| `aeroxe.workflow.started` | Workflow | Workflow started |
| `aeroxe.workflow.completed` | Workflow | Workflow done |
| `aeroxe.workflow.failed` | Workflow | Workflow error |
| `aeroxe.security.scan.started` | Security | Scan initiated |
| `aeroxe.security.threat.detected` | Security | Threat found |
| `aeroxe.audit.*` | Audit | All audit events |

---

## 8. Event Schema Standard

Every NATS event follows this structure:

```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "AgentCompleted",
  "timestamp": "2026-07-15T12:00:00Z",
  "tenant_id": "tenant-uuid",
  "service": "agent-orchestrator-service",
  "version": "1.0",
  "data": {
    "execution_id": "12345",
    "agent": "customer-agent",
    "status": "success",
    "tokens_used": 1250,
    "latency_ms": 3500
  }
}
```

---

## 9. JetStream Stream Design

### AI Stream

| Parameter | Value |
|---|---|
| Stream Name | `AI_EVENTS` |
| Subjects | `aeroxe.ai.*` |
| Retention | 7 days |
| Storage | File |
| Replication | 3 |

### Agent Stream

| Parameter | Value |
|---|---|
| Stream Name | `AGENT_EVENTS` |
| Subjects | `aeroxe.agent.*` |
| Retention | 30 days |
| Storage | File |
| Replication | 3 |

### RAG Stream

| Parameter | Value |
|---|---|
| Stream Name | `RAG_EVENTS` |
| Subjects | `aeroxe.rag.*` |
| Retention | 14 days |
| Storage | File |

### Audit Stream

| Parameter | Value |
|---|---|
| Stream Name | `AUDIT_EVENTS` |
| Subjects | `aeroxe.audit.*` |
| Retention | 365 days |
| Storage | File |
| Replication | 3 |

### Workflow Stream

| Parameter | Value |
|---|---|
| Stream Name | `WORKFLOW_EVENTS` |
| Subjects | `aeroxe.workflow.*` |
| Retention | 30 days |

---

## 10. Request Flow Example

**User asks:** "Why is customer internet slow?"

```
User
    |
    v
API Gateway (REST/WS)
    |
    v
AI Gateway Service (JWT validation, rate limiting)
    |
    v
Agent Orchestrator (gRPC)
    |
    v
LFM2.5 Thinking Model (Intent detection)
    |
    v
Plan Created
    |
    v
NATS: aeroxe.agent.execution.started
    |
    ├── Customer Agent (gRPC -> Broadband Service)
    ├── SQL Agent (gRPC -> Customer DB)
    ├── RAG Agent (gRPC -> Knowledge Base)
    |
    v
Results Aggregated
    |
    v
Command-R 7B (Generate final answer)
    |
    v
Response streamed back to user via WebSocket
```

---

## 11. Streaming Response Architecture

For Chat UI with token streaming:

```
User
    |
    v
WebSocket Connection
    |
    v
AI Gateway
    |
    v
gRPC Stream
    |
    v
Ollama (Token generation)
    |
    v
Token Stream
    |
    v
WebSocket -> Frontend
```

---

## 12. Security Requirements

### gRPC Security

| Requirement | Implementation |
|---|---|
| TLS Encryption | All gRPC traffic encrypted |
| Service Authentication | mTLS certificates |
| Metadata Validation | JWT + tenant validation |
| Rate Limiting | Per-service limits |

### NATS Security

| Requirement | Implementation |
|---|---|
| TLS | All NATS connections encrypted |
| Account Isolation | Separate accounts per service |
| Subject Permissions | Publish/subscribe ACLs |
| Authentication | NKey or JWT-based auth |

### Service Permission Matrix

| Service | Publish | Subscribe |
|---|---|---|
| agent-orchestrator | `aeroxe.agent.*` | `aeroxe.rag.*`, `aeroxe.vision.*` |
| rag-service | `aeroxe.rag.*` | `aeroxe.rag.document.*` |
| vision-service | `aeroxe.vision.*` | `aeroxe.vision.image.*` |
| audit-service | - | `aeroxe.*` (all events) |
| workflow-service | `aeroxe.workflow.*` | `aeroxe.workflow.*` |

---

## 13. Final Communication Stack

| Layer | Technology |
|---|---|
| Mobile/Web API | REST |
| Real-time Chat | WebSocket |
| Internal RPC | gRPC |
| Contract | Protocol Buffers |
| Event Bus | NATS JetStream |
| AI Runtime API | Ollama API |
| Database Access | Repository Pattern |
