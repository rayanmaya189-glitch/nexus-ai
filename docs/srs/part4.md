# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 4 — Database Architecture & Data Design

## PostgreSQL + pgvector + Apache AGE + Redis + Elasticsearch + MinIO

---

# 1. Database Architecture Principles

AeroXe Nexus AI follows **Database-per-Microservice** architecture.

Rules:

✅ Each microservice owns its database
✅ No direct database access between services
✅ Communication through gRPC/NATS only
✅ Data consistency through events
✅ Read models can be optimized separately
✅ Tenant isolation mandatory

---

# 2. Data Architecture Overview

```text
                         AeroXe Nexus AI


                               |

                  Microservice Data Ownership


                               |


================================================================


 Identity Service

        |

        PostgreSQL

        identity_db



 Agent Service

        |

        PostgreSQL

        agent_db



 RAG Service

        |

        PostgreSQL + pgvector

        rag_db



 Vision Service

        |

        PostgreSQL

        vision_db



 Memory Service

        |

        PostgreSQL + Redis

        memory_db



 Workflow Service

        |

        PostgreSQL

        workflow_db



 Audit Service

        |

        PostgreSQL + Elasticsearch

        audit_db


================================================================

```

---

# 3. Storage Technology Selection

| Requirement      | Technology     |
| ---------------- | -------------- |
| Transaction Data | PostgreSQL 16  |
| Vector Search    | pgvector       |
| Knowledge Graph  | Apache AGE     |
| Cache            | Redis          |
| Full Text Search | Elasticsearch  |
| File Storage     | MinIO          |
| Event Storage    | NATS JetStream |

---

# 4. Multi-Tenant Data Architecture

AeroXe Nexus AI supports:

* Multiple companies
* Multiple business units
* Multiple users
* Multiple AI agents

Every business table must include:

```sql
tenant_id UUID NOT NULL
```

Example:

```sql
CREATE TABLE ai_sessions (

id UUID PRIMARY KEY,

tenant_id UUID NOT NULL,

user_id UUID NOT NULL,

created_at TIMESTAMP NOT NULL

);

```

---

# 5. Identity Service Database

Database:

```text
identity_db
```

Purpose:

Authentication and authorization.

---

# 5.1 Users Table

```sql
CREATE TABLE users (

id UUID PRIMARY KEY,

tenant_id UUID NOT NULL,

email VARCHAR(255) UNIQUE,

password_hash TEXT,

status VARCHAR(50),

created_at TIMESTAMP,

updated_at TIMESTAMP

);

```

---

# 5.2 Roles

```sql
CREATE TABLE roles (

id UUID PRIMARY KEY,

tenant_id UUID,

name VARCHAR(100),

description TEXT

);

```

---

# 5.3 Permissions

```sql
CREATE TABLE permissions (

id UUID PRIMARY KEY,

name VARCHAR(100),

resource VARCHAR(100),

action VARCHAR(50)

);

```

---

# 5.4 User Roles

```sql
CREATE TABLE user_roles (

user_id UUID,

role_id UUID,

PRIMARY KEY(user_id,role_id)

);

```

---

# 6. AI Gateway Database

Database:

```text
gateway_db
```

Purpose:

Store AI sessions and requests.

---

# 6.1 AI Session

```sql
CREATE TABLE ai_sessions (

id UUID PRIMARY KEY,

tenant_id UUID,

user_id UUID,

started_at TIMESTAMP,

status VARCHAR(50)

);

```

---

# 6.2 AI Requests

```sql
CREATE TABLE ai_requests (

id UUID PRIMARY KEY,

session_id UUID,

prompt TEXT,

model VARCHAR(100),

status VARCHAR(50),

created_at TIMESTAMP

);

```

---

# 7. Agent Orchestrator Database

Database:

```text
agent_db
```

Purpose:

Track AI agent execution.

---

# 7.1 Agents

```sql
CREATE TABLE agents (

id UUID PRIMARY KEY,

name VARCHAR(100),

type VARCHAR(100),

model VARCHAR(100),

status VARCHAR(50)

);

```

---

# 7.2 Agent Executions

```sql
CREATE TABLE agent_executions (

id UUID PRIMARY KEY,

tenant_id UUID,

agent_id UUID,

task TEXT,

status VARCHAR(50),

started_at TIMESTAMP,

completed_at TIMESTAMP

);

```

---

# 7.3 Agent Steps

```sql
CREATE TABLE agent_steps (

id UUID PRIMARY KEY,

execution_id UUID,

step_number INT,

action TEXT,

result JSONB

);

```

---

# 8. RAG Database Architecture

Database:

```text
rag_db
```

Technology:

```text
PostgreSQL

+

pgvector

```

---

# 8.1 Documents Table

```sql
CREATE TABLE documents (

id UUID PRIMARY KEY,

tenant_id UUID,

filename TEXT,

type VARCHAR(50),

status VARCHAR(50),

created_at TIMESTAMP

);

```

---

# 8.2 Document Chunks

```sql
CREATE TABLE document_chunks (

id UUID PRIMARY KEY,

document_id UUID,

content TEXT,

chunk_index INT,

embedding vector(768)

);

```

---

# 8.3 Vector Index

```sql
CREATE INDEX embedding_index

ON document_chunks

USING ivfflat

(embedding vector_cosine_ops);

```

---

# 8.4 Metadata

```sql
CREATE TABLE document_metadata (

id UUID PRIMARY KEY,

document_id UUID,

metadata JSONB

);

```

Example:

```json
{
 "department":"billing",
 "category":"invoice",
 "security":"private"
}

```

---

# 9. RAG Data Flow

```text
Document Upload


      |

      |

MinIO Storage


      |

      |

Parser


      |

      |

Chunk Generator


      |

      |

Embedding Model


      |

      |

pgvector


      |

      |

Command-R 7B


      |

      |

Answer

```

---

# 10. Knowledge Graph Database

Technology:

```text
Apache AGE
```

Purpose:

Relationship intelligence.

---

# Example Knowledge Graph

```text
Customer

 |

has

 |

Subscription

 |

connected_to

 |

ONU Device

 |

belongs_to

 |

OLT

 |

located_at

 |

City

```

---

# 10.1 Graph Entities

Nodes:

```text
Customer

Company

Device

Network

Document

Agent

```

---

# Relationships:

```text
OWNS

CONNECTED_TO

DEPENDS_ON

RELATED_TO

BELONGS_TO

```

---

# 11. Vision Service Database

Database:

```text
vision_db
```

---

# 11.1 Images

```sql
CREATE TABLE images (

id UUID PRIMARY KEY,

tenant_id UUID,

storage_path TEXT,

type VARCHAR(50),

created_at TIMESTAMP

);

```

---

# 11.2 Vision Analysis

```sql
CREATE TABLE vision_analysis (

id UUID PRIMARY KEY,

image_id UUID,

model VARCHAR(100),

description TEXT,

confidence FLOAT,

metadata JSONB

);

```

---

# 11.3 OCR Results

```sql
CREATE TABLE ocr_results (

id UUID PRIMARY KEY,

image_id UUID,

text TEXT

);

```

---

# 12. Memory Service Database

Database:

```text
memory_db
```

---

# 12.1 Short Term Memory

Technology:

```text
Redis
```

Example:

```
conversation:{user_id}

```

Stores:

* Current conversation
* Temporary context
* Active tasks

---

# 12.2 Long Term Memory

PostgreSQL:

```sql
CREATE TABLE memories (

id UUID PRIMARY KEY,

user_id UUID,

content TEXT,

embedding vector(768),

importance FLOAT,

created_at TIMESTAMP

);

```

---

# 13. Workflow Database

Database:

```text
workflow_db
```

---

# 13.1 Workflow Definition

```sql
CREATE TABLE workflows (

id UUID PRIMARY KEY,

name VARCHAR(100),

definition JSONB

);

```

---

# 13.2 Workflow Execution

```sql
CREATE TABLE workflow_instances (

id UUID PRIMARY KEY,

workflow_id UUID,

status VARCHAR(50),

started_at TIMESTAMP

);

```

---

# 14. Audit Database

Database:

```text
audit_db
```

Purpose:

Complete compliance tracking.

---

# 14.1 Audit Events

```sql
CREATE TABLE audit_events (

id UUID PRIMARY KEY,

tenant_id UUID,

service VARCHAR(100),

event_type VARCHAR(100),

payload JSONB,

created_at TIMESTAMP

);

```

---

# 15. Elasticsearch Usage

Used for:

* Logs
* AI conversations
* Audit search
* Knowledge search

Indexes:

```
ai_logs

audit_events

documents

security_events

```

---

# 16. MinIO Storage Design

Purpose:

Object storage.

Buckets:

```
aeroxe-documents

aeroxe-images

aeroxe-model-files

aeroxe-backups

```

---

# 17. Database Event Synchronization

Example:

Document Processing:

```text
rag-service


DocumentProcessed Event


        |

        |

NATS JetStream


        |

        |

knowledge-graph-service


Update Relationships


```

---

# 18. Repository Pattern

DDD Repository Example:

```rust
trait AgentRepository {


async fn save(

agent: Agent

);



async fn find_by_id(

id: AgentId

);


}

```

---

# 19. Database Migration Strategy

Technology:

Recommended:

```text
EntORM Migrate (Go)
```

or

```text
SeaORM Migrate (Rust)
```

Structure:

```
migrations/


001_create_users.sql

002_create_roles.sql

003_add_permissions.sql

```

---

# 20. Backup Strategy

## PostgreSQL

* Daily full backup
* WAL archiving
* Point-in-time recovery

## MinIO

* Versioning enabled
* Replication

## NATS JetStream

* Snapshot backup

---

# 21. Performance Requirements

| Component            | Target |
| -------------------- | ------ |
| Vector Search        | <200ms |
| SQL Query            | <2s    |
| Redis Lookup         | <10ms  |
| PostgreSQL API Query | <100ms |
| Elasticsearch Search | <300ms |

---

# 22. Final Database Architecture

```text
                     AeroXe Nexus AI


                            |

==================================================


PostgreSQL

├── identity_db

├── gateway_db

├── agent_db

├── rag_db

├── vision_db

├── memory_db

├── workflow_db

└── audit_db



pgvector

└── Semantic Search



Apache AGE

└── Knowledge Graph



Redis

└── Memory Cache



Elasticsearch

└── Search + Analytics



MinIO

└── Object Storage


==================================================

```

---

# Part 4 Completed

Covered:

✅ Database-per-service architecture
✅ PostgreSQL schema design
✅ pgvector RAG database
✅ Knowledge Graph model
✅ Memory architecture
✅ Multi-tenancy
✅ Repository pattern
✅ Backup strategy
