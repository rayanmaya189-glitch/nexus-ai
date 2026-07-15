# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 2 — Complete DDD Architecture Design

## Domain-Driven Design (DDD) + Microservices + TDD

---

# 1. DDD Architecture Overview

AeroXe Nexus AI is designed using **Domain-Driven Design principles**.

The system is divided into independent **Bounded Contexts**.

Each bounded context:

* Owns its business logic
* Owns its database
* Exposes contracts through gRPC
* Publishes events through NATS JetStream
* Has independent deployment lifecycle

---

# 2. Bounded Context Map

```text
                         AeroXe Nexus AI


                               |

              ---------------------------------

              Core AI Intelligence Domain


              ---------------------------------


        Identity        AI Gateway       Agent Platform


           |                |                 |


           |                |                 |


        Security       Request Mgmt     AI Execution


                               |


              ---------------------------------

              Intelligence Domains


              ---------------------------------


       RAG          Vision          SQL Intelligence


        |             |                   |


        |             |                   |


   Knowledge     Image AI          Business Data



              ---------------------------------

              Support Domains


              ---------------------------------


       Memory        Workflow        Audit


```

---

# 3. Core Domain Classification

DDD classification:

| Domain                | Type              |
| --------------------- | ----------------- |
| Agent Orchestration   | Core Domain       |
| AI Gateway            | Core Domain       |
| RAG Intelligence      | Core Domain       |
| Vision Intelligence   | Core Domain       |
| SQL Intelligence      | Core Domain       |
| Security Intelligence | Core Domain       |
| Identity              | Supporting Domain |
| Memory                | Supporting Domain |
| Audit                 | Supporting Domain |
| Workflow              | Supporting Domain |

---

# 4. Microservice Design Rules

Every microservice must follow:

```
service-name/

├── domain/
│
│   ├── entities/
│   ├── aggregates/
│   ├── value_objects/
│   ├── domain_events/
│   ├── repositories/
│
├── application/
│
│   ├── commands/
│   ├── queries/
│   ├── handlers/
│   ├── use_cases/
│
├── infrastructure/
│
│   ├── database/
│   ├── grpc/
│   ├── nats/
│   ├── external/
│
├── interfaces/
│
│   ├── rest/
│   ├── websocket/
│   ├── grpc/
│
├── tests/

```

---

# 5. Identity Bounded Context

## Service

```
identity-service
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

---

# Value Objects

```
UserId

TenantId

EmailAddress

Permission

```

---

# Domain Events

```
UserCreated

UserUpdated

RoleAssigned

PermissionChanged

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

Service:

```
ai-gateway-service
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

Service:

```
agent-orchestrator-service
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

# Domain Events

```
AgentStarted

AgentCompleted

AgentFailed

ToolExecuted

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

Service:

```
rag-service
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

# Domain Events

```
DocumentUploaded

DocumentProcessed

EmbeddingCreated

KnowledgeUpdated

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

Service:

```
vision-service
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

Service:

```
sql-agent-service
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

Service:

```
memory-service
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

Service:

```
workflow-service
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

Service:

```
security-ai-service
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

Service:

```
audit-service
```

Purpose:

Complete AI activity tracking.

Events:

```
AIRequestLogged

DataAccessLogged

ToolExecutionLogged

SecurityEventLogged

```

---

# 15. Domain Event Architecture

All domains communicate using events.

Example:

```
rag-service


DocumentProcessed


        |

        |

NATS JetStream


        |

        |

agent-orchestrator


Update Knowledge Available


```

---

# 16. Final DDD Service Map

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

```

---

# 17. TDD Requirements

Every service must contain:

```
tests/


├── unit/

├── integration/

├── contract/

├── performance/

└── security/

```

---
# End of Part 2