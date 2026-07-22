# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 2 — Complete DDD Architecture Design

## Domain-Driven Design (DDD) + Modular Monolith + SeaORM + TDD

---

# 1. DDD Architecture Overview

AeroXe Nexus AI is designed using **Domain-Driven Design principles** as a **Modular Monolith**.

The system is divided into independent **Bounded Contexts** under `src/modules/`.

Each bounded context:

* Owns its business logic
* Owns its database **schema** (shared PostgreSQL via SeaORM)
* Exposes public API through **gRPC** (sync) or **NATS** (async)
* Publishes events through **NATS JetStream** (versioned subjects)
* Deploys as part of a single binary (extractable later)

---

# 2. Bounded Context Map

```text
                         AeroXe Nexus AI


                               |

              ---------------------------------

              Core AI Intelligence Domain


              ---------------------------------


     identity    customer     ai-gateway    agent


        |           |             |           |


        |           |             |           |


    security    workflow      rag        vision


                               |


              ---------------------------------

              Intelligence Domains


              ---------------------------------


    sql-agent    memory       audit        notification


        |          |             |              |


        |          |             |              |


   Business     Context      Compliance    Alerts

    Data        Memory


              ---------------------------------

              Infrastructure Modules


              ---------------------------------


   model-registry    config         ecosystem


              (stateless gateway module)

```

---

# 3. Core Domain Classification

DDD classification:

| Domain                | Type              | Module Name | Schema Prefix |
| --------------------- | ----------------- | ----------- | ------------- |
| Agent Orchestration   | Core Domain       | `agent`     | `agent_`      |
| AI Gateway            | Core Domain       | `ai-gateway`| `ai_`         |
| RAG Intelligence      | Core Domain       | `rag`       | `rag_`        |
| Vision Intelligence   | Core Domain       | `vision`    | `vision_`     |
| SQL Intelligence      | Core Domain       | `sql-agent` | `sql_`        |
| Security Intelligence | Core Domain       | `security`  | `security_`   |
| Identity              | Supporting Domain | `identity`  | `identity_`   |
| Customer              | Supporting Domain | `customer`  | `customer_`   |
| Memory                | Supporting Domain | `memory`    | `memory_`     |
| Audit                 | Supporting Domain | `audit`     | `audit_`      |
| Workflow              | Supporting Domain | `workflow`  | `workflow_`   |

---

# 4. Module Design Rules

Every module under `src/modules/<name>/` must follow:

```
src/modules/<name>/

├── mod.rs                         # Public API trait + re-exports
├── domain/
│   ├── mod.rs
│   ├── aggregates/                # Aggregate roots with invariants
│   ├── entities/                  # Mutable domain objects (with IDs)
│   ├── value_objects/             # Immutable validated types
│   ├── events/                    # Domain event structs
│   └── rules/                     # Business rules
├── application/
│   ├── mod.rs
│   ├── commands/                  # CQRS command structs
│   ├── queries/                   # CQRS query structs
│   ├── handlers/                  # Command/query handlers
│   └── services/                  # Application services
├── infrastructure/
│   ├── mod.rs
│   ├── repository/                # SeaORM repositories
│   ├── messaging/                 # NATS publishers/subscribers
│   └── security/                  # JWT, hashing
├── api/
│   ├── mod.rs
│   ├── http/                      # Axum HTTP handlers
│   └── http/                      # Axum HTTP handlers
├── migrations/                    # SeaORM migration files
└── tests/
    ├── unit/
    ├── integration/
    └── contract/

```

---

# 5. Identity Bounded Context

## Module

```
identity (src/modules/identity/)
```

Schema:

```
identity_
```

Purpose:

Manage users, tenants, roles, permissions.

---

# Aggregate

## User Aggregate

```
User
 |
 +-- Profile
 |
 +-- Roles
 |
 +-- Permissions
 |
 +-- Sessions
```

---

# Entities

## User

Attributes:

```
UserId
TenantId
Email
Status
CreatedAt
```

## Role

```
RoleId
Name
Permissions
```

## Session

```
SessionId
UserId
Token
RefreshToken
Expiry
```

---

# Value Objects

```
UserId

TenantId

EmailAddress

PasswordHash (bcrypt)

Permission

```

---

# Domain Events (Versioned NATS Subjects)

```
aeroxe.v1.identity.user.created

aeroxe.v1.identity.user.updated

aeroxe.v1.identity.role.assigned

aeroxe.v1.identity.permission.changed

```

---

# Commands

```
CreateUserCommand

AssignRoleCommand

UpdatePermissionCommand

```

---

# Queries

```
GetUserQuery

GetPermissionsQuery

```

---

# 6. AI Gateway Bounded Context

Module:

```
ai-gateway (src/modules/ai-gateway/)
```

Schema:

```
ai_
```

Purpose:

Central AI request management.

---

# Aggregate

## AIRequest

```
AIRequest

 |
 +-- RequestContext

 |
 +-- SecurityContext

 |
 +-- ExecutionPlan

```

---

# Entities

## AI Session

```
SessionId

UserId

TenantId

StartedAt

```

---

# Value Objects

```
Prompt

ModelName

TenantId

RequestId

```

---

# Domain Events

```
AIRequestReceived

AIResponseGenerated

AIRequestFailed

```

---

# Commands

```
SubmitAIRequestCommand

CancelAIRequestCommand

```

---

# 7. Agent Orchestration Context

Module:

```
agent (src/modules/agent/)
```

Schema:

```
agent_
```

Core domain.

---

# Purpose

Control AI agents.

Responsibilities:

* Planning
* Routing
* Tool selection
* Execution

---

# Aggregate

## AgentExecution

```
AgentExecution


 |
 +-- Task

 |
 +-- Plan

 |
 +-- ToolExecution

 |
 +-- Result

```

---

# Entities

## Agent

```
AgentId

AgentType

Capabilities

Model

```

---

## Task

```
TaskId

Status

Priority

CreatedAt

```

---

# Value Objects

```
AgentId

TaskId

ExecutionId

```

---

# Domain Events (Versioned NATS Subjects)

```
aeroxe.v1.agent.started

aeroxe.v1.agent.completed

aeroxe.v1.agent.failed

aeroxe.v1.agent.tool.executed

```

---

# Commands

```
StartAgentCommand

ExecuteToolCommand

CompleteTaskCommand

```

---

# Agent Routing Logic

Example:

```
User Request

        |

Planner

(lfm2.5-thinking)


        |

Decision


        |

---------------------

Coding

 |

Qwen Coder


Security

 |

WhiteRabbitNeo


Image

 |

Qwen3-VL


Document

 |

Command-R


Business

 |

Llama

---------------------

```

---

# 8. RAG Knowledge Context

Module:

```
rag (src/modules/rag/)
```

Schema:

```
rag_
```

Purpose:

Enterprise knowledge intelligence.

---

# Aggregate

## KnowledgeDocument

```
Document


 |
 +-- Metadata

 |
 +-- Chunks

 |
 +-- Embeddings

```

---

# Entities

## Document

```
DocumentId

TenantId

FileName

Status

```

---

## Chunk

```
ChunkId

Content

Position

Embedding

```

---

# Value Objects

```
DocumentId

EmbeddingVector

DocumentType

```

---

# Domain Events (Versioned NATS Subjects)

```
aeroxe.v1.rag.document.uploaded

aeroxe.v1.rag.document.processed

aeroxe.v1.rag.embedding.created

aeroxe.v1.rag.knowledge.updated

```

---

# Commands

```
UploadDocumentCommand

ProcessDocumentCommand

SearchKnowledgeCommand

```

---

# RAG Flow

```
Upload

 |

Parser

 |

Chunking

 |

Embedding

 |

Vector Store

 |

Retriever

 |

Command-R

 |

Answer

```

---

# 9. Vision Intelligence Context

Module:

```
vision (src/modules/vision/)
```

Schema:

```
vision_
```

Model:

```
Qwen3-VL:4B
```

---

# Aggregate

## VisionAnalysis

```
VisionAnalysis

 |
 +-- Image

 |
 +-- Detection

 |
 +-- Extraction

```

---

# Entities

## Image

```
ImageId

URL

Type

Size

```

---

## AnalysisResult

```
ResultId

Confidence

Description

```

---

# Domain Events

```
ImageUploaded

VisionProcessingStarted

VisionCompleted

```

---

# Commands

```
AnalyzeImageCommand

ExtractTextCommand

DetectObjectCommand

```

---

# 10. SQL Intelligence Context

Module:

```
sql-agent (src/modules/sql-agent/)
```

Schema:

```
sql_
```

Purpose:

Natural language business intelligence.

---

# Aggregate

## QueryExecution

```
QueryExecution

 |
 +-- GeneratedSQL

 |
 +-- ValidationResult

 |
 +-- ResultSet

```

---

# Entities

## SQLQuery

```
QueryId

SQL

Status

```

---

# Value Objects

```
DatabaseId

SQLStatement

QueryPermission

```

---

# Domain Events

```
QueryGenerated

QueryApproved

QueryExecuted

```

---

# SQL Safety Rules

Allowed:

```
SELECT

JOIN

GROUP BY

ORDER BY

COUNT

SUM

AVG

```

Blocked:

```
DELETE

UPDATE

DROP

ALTER

TRUNCATE

```

---

# 11. Memory Context

Module:

```
memory (src/modules/memory/)
```

Schema:

```
memory_
```

Purpose:

Maintain AI memory.

---

# Aggregate

## MemoryProfile

```
MemoryProfile

 |
 +-- ShortTermMemory

 |
 +-- LongTermMemory

```

---

# Storage

Short Term:

```
Redis
```

Long Term:

```
PostgreSQL
```

---

# Events

```
MemoryCreated

MemoryUpdated

MemoryExpired

```

---

# 12. Workflow Context

Module:

```
workflow (src/modules/workflow/)
```

Schema:

```
workflow_
```

Purpose:

Business automation.

---

# Aggregate

## WorkflowInstance

```
WorkflowInstance

 |
 +-- Steps

 |
 +-- Approvals

 |
 +-- Actions

```

---

# Events

```
WorkflowStarted

StepCompleted

WorkflowFinished

```

---

# 13. Security Intelligence Context

Module:

```
security (src/modules/security/)
```

Schema:

```
security_
```

Model:

```
WhiteRabbitNeo:7B
```

---

# Aggregate

## SecurityAnalysis

```
SecurityAnalysis

 |
 +-- Finding

 |
 +-- Recommendation

```

---

# Events

```
SecurityScanStarted

ThreatDetected

ReportGenerated

```

---

# 14. Audit Context

Module:

```
audit (src/modules/audit/)
```

Schema:

```
audit_
```

Purpose:

Complete AI activity tracking.

Events (Versioned NATS Subjects):

```
aeroxe.v1.audit.ai.request

aeroxe.v1.audit.data.access

aeroxe.v1.audit.tool.execution

aeroxe.v1.audit.security.event

```

---

# 15. Domain Event Architecture (Versioned)

All modules communicate using versioned NATS JetStream events.

Example:

```
rag module


aeroxe.v1.rag.document.processed


        |

        |

NATS JetStream


        |

        |

agent module


Update Knowledge Available


```

---

# 16. Customer Bounded Context (NEW)

Module:

```
customer (src/modules/customer/)
```

Schema:

```
customer_
```

Purpose:

Manage customers, profiles, addresses, lifecycle.

## Aggregate

```
Customer
 |
 +-- Profile (name, email, phone)
 |
 +-- Status (active, suspended, inactive, archived)
 |
 +-- Addresses (billing, shipping, physical)
 |
 +-- Metadata (tags, custom_fields, notes)
```

## Domain Events (Versioned NATS Subjects)

```
aeroxe.v1.customer.customer.created

aeroxe.v1.customer.customer.activated

aeroxe.v1.customer.customer.suspended

aeroxe.v1.customer.customer.updated

```

---

# 17. Final DDD Module Map

```
AeroXe Nexus AI — src/modules/


Core Domain

├── ai-gateway     (schema: ai_)
├── agent          (schema: agent_)
├── rag            (schema: rag_)
├── vision         (schema: vision_)
├── sql-agent      (schema: sql_)
├── security       (schema: security_)


Supporting Domain

├── identity       (schema: identity_)
├── customer       (schema: customer_)  ← NEW
├── memory         (schema: memory_)
├── workflow       (schema: workflow_)
├── audit          (schema: audit_)


Infrastructure

├── gateway        (stateless, Redis)
├── model-registry (schema: models_)
├── notification   (schema: notification_)
├── config         (schema: config_)
├── ecosystem      (schema: ecosystem_)

```

---

# 18. TDD Requirements

Every module must contain:

```
tests/

├── unit/                  # Domain unit tests
├── integration/           # SeaORM + DB integration tests
├── contract/              # Module boundary trait contract tests
├── performance/
└── security/

```

---
# End of Part 2