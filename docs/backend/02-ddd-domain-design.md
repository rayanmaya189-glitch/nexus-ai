# AeroXe Nexus AI — DDD Domain Design

## Domain-Driven Design + Bounded Contexts + Aggregate Design

---

## 1. DDD Architecture Overview

AeroXe Nexus AI is designed using Domain-Driven Design principles. The system is divided into independent Bounded Contexts, each owning its business logic, database, gRPC contracts, and NATS events.

---

## 2. Core Domain Classification

| Domain | Type | Service | Language |
|---|---|---|---|
| Agent Orchestration | Core Domain | `agent-orchestrator-service` | Rust |
| AI Gateway | Core Domain | `ai-gateway-service` | Go |
| RAG Intelligence | Core Domain | `rag-service` | Rust |
| Vision Intelligence | Core Domain | `vision-service` | Rust |
| SQL Intelligence | Core Domain | `sql-agent-service` | Go |
| Security Intelligence | Core Domain | `security-ai-service` | Rust |
| Identity | Supporting Domain | `identity-service` | Go |
| Memory | Supporting Domain | `memory-service` | Rust |
| Audit | Supporting Domain | `audit-service` | Go |
| Workflow | Supporting Domain | `workflow-service` | Go |

### Infrastructure Services

| Service | Language |
|---|---|
| `model-registry-service` | Go |
| `notification-service` | Go |
| `configuration-service` | Go |

---

## 3. Microservice Code Structure

Every microservice follows a strict layered architecture:

```
service-name/
├── domain/
│   ├── entities/
│   ├── aggregates/
│   ├── value_objects/
│   ├── domain_events/
│   └── repositories/
├── application/
│   ├── commands/
│   ├── queries/
│   ├── handlers/
│   └── use_cases/
├── infrastructure/
│   ├── database/
│   ├── grpc/
│   ├── nats/
│   └── external/
├── interfaces/
│   ├── rest/
│   ├── websocket/
│   └── grpc/
└── tests/
    ├── unit/
    ├── integration/
    ├── contract/
    ├── performance/
    └── security/
```

---

## 4. Identity Bounded Context

**Service:** `identity-service`
**Purpose:** Manage users, tenants, roles, permissions

### Aggregate: User

```
User
├── Profile
├── Roles[]
└── Permissions[]
```

### Entities

| Entity | Attributes |
|---|---|
| User | UserId, TenantId, Email, Status, CreatedAt |
| Role | RoleId, Name, Permissions[] |
| Tenant | TenantId, Name, Plan, Settings |

### Value Objects

- `UserId` (UUID)
- `TenantId` (UUID)
- `EmailAddress` (validated string)
- `Permission` (resource + action pair)

### Domain Events

| Event | Trigger |
|---|---|
| `UserCreated` | New user registration |
| `UserUpdated` | Profile change |
| `RoleAssigned` | Role assignment |
| `PermissionChanged` | Permission modification |

### Commands / Queries

| Command | Query |
|---|---|
| `CreateUserCommand` | `GetUserQuery` |
| `AssignRoleCommand` | `GetPermissionsQuery` |
| `UpdatePermissionCommand` | `GetTenantUsersQuery` |

---

## 5. AI Gateway Bounded Context

**Service:** `ai-gateway-service`
**Purpose:** Central AI request management and routing

### Aggregate: AIRequest

```
AIRequest
├── RequestContext
├── SecurityContext
└── ExecutionPlan
```

### Entities

| Entity | Attributes |
|---|---|
| AISession | SessionId, UserId, TenantId, StartedAt, Status |
| AIRequest | RequestId, SessionId, Prompt, Model, Status |

### Value Objects

- `Prompt` (sanitized text)
- `ModelName` (validated model identifier)
- `RequestId` (UUID)
- `TenantId` (UUID)

### Domain Events

| Event | Description |
|---|---|
| `AIRequestReceived` | New request enters the gateway |
| `AIResponseGenerated` | Model response produced |
| `AIRequestFailed` | Request could not be fulfilled |

### Commands

| Command | Description |
|---|---|
| `SubmitAIRequestCommand` | Queue a new AI request |
| `CancelAIRequestCommand` | Cancel in-flight request |

---

## 6. Agent Orchestration Bounded Context

**Service:** `agent-orchestrator-service`
**Purpose:** AI agent lifecycle, planning, tool selection, execution

### Aggregate: AgentExecution

```
AgentExecution
├── Task
├── Plan
├── ToolExecution[]
└── Result
```

### Entities

| Entity | Attributes |
|---|---|
| Agent | AgentId, AgentType, Capabilities[], Model |
| Task | TaskId, Status, Priority, CreatedAt |
| ExecutionStep | StepId, ExecutionId, StepNumber, Action, Result |

### Value Objects

- `AgentId` (UUID)
- `TaskId` (UUID)
- `ExecutionId` (UUID)
- `AgentType` (enum: planner, customer, developer, rag, vision, security, business)

### Domain Events

| Event | Description |
|---|---|
| `AgentStarted` | Agent begins execution |
| `AgentCompleted` | Agent finishes successfully |
| `AgentFailed` | Agent execution error |
| `ToolExecuted` | A tool call was made |

### Commands

| Command | Description |
|---|---|
| `StartAgentCommand` | Initialize agent execution |
| `ExecuteToolCommand` | Invoke a tool through the policy engine |
| `CompleteTaskCommand` | Mark task as done |

### Agent Routing Logic

```
User Request
    |
    v
Planner Agent (lfm2.5-thinking:1.2b)
    |
    v
Intent Classification
    |
    ├── Coding     -> Qwen2.5-Coder:3B
    ├── Security   -> WhiteRabbitNeo:7B
    ├── Image      -> Qwen3-VL:4B
    ├── Document   -> Command-R:7B
    ├── Business   -> Llama3.1:7B
    └── General    -> Phi-4-Mini:3.8B
```

---

## 7. RAG Knowledge Bounded Context

**Service:** `rag-service`
**Purpose:** Enterprise knowledge intelligence

### Aggregate: KnowledgeDocument

```
KnowledgeDocument
├── Metadata
├── Chunks[]
└── Embeddings[]
```

### Entities

| Entity | Attributes |
|---|---|
| Document | DocumentId, TenantId, FileName, Type, Status |
| Chunk | ChunkId, Content, Position, Embedding (vector) |

### Value Objects

- `DocumentId` (UUID)
- `EmbeddingVector` (768-dimensional float array)
- `DocumentType` (enum: pdf, docx, html, markdown, code)

### Domain Events

| Event | Description |
|---|---|
| `DocumentUploaded` | File received |
| `DocumentProcessed` | Parsing complete |
| `EmbeddingCreated` | Vector embeddings generated |
| `KnowledgeUpdated` | Knowledge base modified |

### Commands

| Command | Description |
|---|---|
| `UploadDocumentCommand` | Submit a document for ingestion |
| `ProcessDocumentCommand` | Trigger parsing + chunking + embedding |
| `SearchKnowledgeCommand` | Query the knowledge base |

### RAG Processing Flow

```
Upload -> Parser -> Chunking -> Embedding -> Vector Store
                                                |
                                                v
Query -> Hybrid Search -> Re-ranking -> Command-R 7B -> Answer
```

---

## 8. Vision Intelligence Bounded Context

**Service:** `vision-service`
**Model:** `Qwen3-VL:4B`

### Aggregate: VisionAnalysis

```
VisionAnalysis
├── Image
├── Detection
└── Extraction
```

### Entities

| Entity | Attributes |
|---|---|
| Image | ImageId, URL, Type, Size, TenantId |
| AnalysisResult | ResultId, Confidence, Description, Metadata |

### Domain Events

| Event | Description |
|---|---|
| `ImageUploaded` | Image received |
| `VisionProcessingStarted` | Analysis begins |
| `VisionCompleted` | Analysis complete |

### Commands

| Command | Description |
|---|---|
| `AnalyzeImageCommand` | General image understanding |
| `ExtractTextCommand` | OCR extraction |
| `DetectObjectCommand` | Object detection |

---

## 9. SQL Intelligence Bounded Context

**Service:** `sql-agent-service`
**Purpose:** Natural language business intelligence

### Aggregate: QueryExecution

```
QueryExecution
├── GeneratedSQL
├── ValidationResult
└── ResultSet
```

### SQL Safety Rules

**Allowed:** SELECT, JOIN, GROUP BY, ORDER BY, COUNT, SUM, AVG
**Blocked:** DELETE, UPDATE, DROP, ALTER, TRUNCATE, INSERT

### Domain Events

| Event | Description |
|---|---|
| `QueryGenerated` | SQL generated from natural language |
| `QueryApproved` | Passed validation |
| `QueryExecuted` | Executed against read replica |

---

## 10. Memory Bounded Context

**Service:** `memory-service`
**Purpose:** Maintain AI memory across sessions

### Aggregate: MemoryProfile

```
MemoryProfile
├── ShortTermMemory (Redis)
└── LongTermMemory (PostgreSQL + pgvector)
```

### Storage Strategy

| Type | Technology | TTL | Purpose |
|---|---|---|---|
| Short-Term | Redis | 24h | Current conversation, temporary context |
| Long-Term | PostgreSQL + pgvector | Permanent | User preferences, past conversations |
| Organizational | Apache AGE | Permanent | Entity relationships |

---

## 11. Workflow Bounded Context

**Service:** `workflow-service`
**Purpose:** Business automation and approvals

### Aggregate: WorkflowInstance

```
WorkflowInstance
├── Steps[]
├── Approvals[]
└── Actions[]
```

### Domain Events

| Event | Description |
|---|---|
| `WorkflowStarted` | Workflow execution begins |
| `StepCompleted` | Individual step done |
| `WorkflowFinished` | All steps complete |

---

## 12. Security Intelligence Bounded Context

**Service:** `security-ai-service`
**Model:** `WhiteRabbitNeo:7B`

### Aggregate: SecurityAnalysis

```
SecurityAnalysis
├── Finding
└── Recommendation
```

### Domain Events

| Event | Description |
|---|---|
| `SecurityScanStarted` | Scan initiated |
| `ThreatDetected` | Security issue found |
| `ReportGenerated` | Analysis report produced |

---

## 13. Audit Bounded Context

**Service:** `audit-service`
**Purpose:** Complete AI activity tracking for compliance

### Domain Events

| Event | Description |
|---|---|
| `AIRequestLogged` | AI interaction recorded |
| `DataAccessLogged` | Data access tracked |
| `ToolExecutionLogged` | Tool call recorded |
| `SecurityEventLogged` | Security event tracked |

---

## 14. Domain Event Architecture

All domains communicate through events via NATS JetStream.

### Event Flow Example

```
rag-service
    |
    v  DocumentProcessed event
NATS JetStream
    |
    v  Consumer
knowledge-graph-service
    |
    v  Update Relationships
```

### Event Schema Standard

Every event follows this structure:

```json
{
  "event_id": "uuid",
  "event_type": "AgentCompleted",
  "timestamp": "2026-07-15T12:00:00Z",
  "tenant_id": "uuid",
  "service": "agent-service",
  "version": "1.0",
  "data": {}
}
```

---

## 15. Final DDD Service Map

```
AeroXe Nexus AI

Core Domain
├── ai-gateway-service
├── agent-orchestrator-service
├── rag-service
├── vision-service
├── sql-agent-service
├── security-ai-service

Supporting Domain
├── identity-service
├── memory-service
├── workflow-service
├── audit-service

Infrastructure
├── model-registry-service
├── notification-service
├── configuration-service
├── ecosystem-integration-service
```
