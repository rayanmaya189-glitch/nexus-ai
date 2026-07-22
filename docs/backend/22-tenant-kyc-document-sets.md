# AeroXe Nexus AI — Tenant KYC, Document Sets & Agent Isolation

## Tenant Onboarding + KYC Verification + Document Set Management + Agent-Document Binding

> **Modular Monolith Context:** KYC handling is in `identity` module (`src/modules/identity/`, schema `identity_`), document sets in `rag` module (`src/modules/rag/`, schema `rag_`), agent bindings in `agent` module (`src/modules/agent/`, schema `agent_`), and customer data in `customer` module (`src/modules/customer/`, schema `customer_`). All modules communicate via trait interfaces — not gRPC. All database access uses SeaORM — no raw SQL. See [DDD Domain Design](02-ddd-domain-design.md).

---

## 1. Overview

This document defines five critical flows:

1. **Tenant KYC Flow** — Verification before platform access (in `identity` module)
2. **Document Set Creation** — Organizing documents into scoped collections (in `rag` module)
3. **Agent-Document Set Binding** — Connecting agents to specific document sets (in `agent` module)
4. **Agent-Database Binding** — Connecting agents to specific databases and tables with credential validation (in `agent` module)
5. **Agent Isolation** — Enforcing that agents access ONLY their bound resources
6. **Customer Module Integration** — Customer lifecycle events (in `customer` module)

---

## 2. Tenant KYC Flow

### 2.1 Purpose

Every tenant must complete KYC (Know Your Customer) verification before accessing AI features. This ensures compliance, accountability, and abuse prevention.

### 2.2 KYC States

```
PENDING → DOCUMENTS_SUBMITTED → UNDER_REVIEW → APPROVED
                                                → REJECTED
                                                → REQUIRES_ADDITIONAL_INFO
```

| State | Description |
|---|---|
| `PENDING` | Tenant registered, KYC not started |
| `DOCUMENTS_SUBMITTED` | KYC documents uploaded |
| `UNDER_REVIEW` | Admin reviewing documents |
| `APPROVED` | KYC passed, full platform access |
| `REJECTED` | KYC failed, access restricted |
| `REQUIRES_ADDITIONAL_INFO` | More documents needed |

### 2.3 KYC Document Types

| Document | Required | Purpose |
|---|---|---|
| Business Registration | Yes | Legal entity verification |
| Tax ID / GST Certificate | Yes | Tax compliance |
| Director/Owner ID | Yes | Identity verification |
| Address Proof | Yes | Location verification |
| Bank Statement | No (Enterprise) | Financial verification |
| Industry License | Conditional | Regulated industries |

### 2.4 KYC Flow Diagram

```
Tenant Registers
  → Account Created (status: pending_kyc)
    → KYC Portal Accessible
      → Upload Documents
        → Submit for Review
          → Admin Notification
            → Manual Review
              → APPROVED: Account activated, status: active
              → REJECTED: Account locked, notification sent
              → REQUIRES_ADDITIONAL_INFO: Partial access, re-upload allowed
```

### 2.5 KYC API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/kyc/status` | Get KYC verification status |
| POST | `/api/v1/kyc/documents` | Upload KYC documents |
| GET | `/api/v1/kyc/documents` | List submitted documents |
| DELETE | `/api/v1/kyc/documents/:id` | Remove a KYC document |
| POST | `/api/v1/kyc/submit` | Submit for review |
| POST | `/api/v1/kyc/review` | Admin: approve/reject |
| GET | `/api/v1/kyc/history` | KYC audit trail |

### 2.6 KYC Enforcement

| Feature | Pre-KYC | Post-KYC |
|---|---|---|
| Login | Allowed | Allowed |
| Dashboard | Read-only | Full |
| AI Chat | Blocked | Allowed |
| Document Upload | Blocked | Allowed |
| Agent Creation | Blocked | Allowed |
| API Access | Blocked | Allowed |
| Admin Panel | Blocked | Allowed |

---

## 3. Document Set Creation Flow

### 3.1 Purpose

Document Sets group related documents into scoped collections. Agents are bound to document sets and can ONLY access documents within their bound sets.

### 3.2 Document Set States

```
DRAFT → ACTIVE → ARCHIVED
```

| State | Description |
|---|---|
| `DRAFT` | Set created, documents being added |
| `ACTIVE` | Set is live, agents can query it |
| `ARCHIVED` | Set deprecated, removed from search |

### 3.3 Document Set Creation Flow

```
User Creates Document Set
  → Set created (status: draft)
    → Name + Description + Tags
      → Add Documents (upload or select existing)
        → Documents processed (chunked, embedded)
          → Activate Set (status: active)
            → Set available for agent binding
```

### 3.4 Document Set Properties

| Property | Type | Description |
|---|---|---|
| id | BIGINT | Auto-generated PK |
| tenant_id | BIGINT | Owning tenant |
| name | VARCHAR(255) | Human-readable name |
| description | TEXT | Purpose of the set |
| tags | JSONB | Classification tags |
| status | VARCHAR(50) | draft / active / archived |
| document_count | INT | Number of documents |
| total_chunks | INT | Total embedded chunks |
| created_by | BIGINT | User who created it |
| created_at | TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | Last modification |

### 3.5 Document Set API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/document-sets` | Create document set |
| GET | `/api/v1/document-sets` | List all sets for tenant |
| GET | `/api/v1/document-sets/:id` | Get set details |
| PATCH | `/api/v1/document-sets/:id` | Update set metadata |
| DELETE | `/api/v1/document-sets/:id` | Delete set (soft) |
| POST | `/api/v1/document-sets/:id/activate` | Activate set |
| POST | `/api/v1/document-sets/:id/archive` | Archive set |
| POST | `/api/v1/document-sets/:id/documents` | Add documents to set |
| DELETE | `/api/v1/document-sets/:id/documents/:doc_id` | Remove document from set |
| GET | `/api/v1/document-sets/:id/documents` | List documents in set |

### 3.6 Adding Documents to a Set

```
User selects documents
  → Validation: documents belong to same tenant
    → Validation: document status is 'processed'
      → Documents linked to set
        → Set document_count updated
          → Set total_chunks recalculated
```

### 3.7 Rules

| Rule | Description |
|---|---|
| Tenant isolation | Sets are scoped to tenant |
| No cross-tenant docs | Cannot add docs from other tenants |
| processed docs only | Only fully processed documents can be added |
| max 1000 docs per set | Prevents oversized sets |
| archival cleanup | Archived sets excluded from search indexing |

---

## 4. Agent-Document Set Binding

### 4.1 Purpose

Each agent is explicitly bound to one or more document sets. This binding determines which documents the agent can access during RAG operations.

### 4.2 Binding Flow

```
Admin Opens Agent Configuration
  → Selects Agent
    → Views available Document Sets (for their tenant)
      → Binds agent to selected sets
        → Binding saved
          → Agent re-indexed for new scope
            → Agent can now query bound sets only
```

### 4.3 Binding Properties

| Property | Type | Description |
|---|---|---|
| id | BIGINT | Auto-generated PK |
| agent_id | BIGINT | References agents(id) |
| document_set_id | BIGINT | References document_sets(id) |
| tenant_id | BIGINT | Tenant scope |
| permission_level | VARCHAR(50) | read / read_write |
| bound_by | BIGINT | User who created binding |
| bound_at | TIMESTAMP | Binding creation time |

### 4.4 Permission Levels

| Level | Agent Can |
|---|---|
| `read` | Query documents, generate answers |
| `read_write` | Query, plus suggest document updates |

### 4.5 Binding API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/agents/:id/document-sets` | Bind agent to document sets |
| GET | `/api/v1/agents/:id/document-sets` | List bound document sets |
| PATCH | `/api/v1/agents/:id/document-sets/:set_id` | Update binding permissions |
| DELETE | `/api/v1/agents/:id/document-sets/:set_id` | Unbind agent from set |
| GET | `/api/v1/document-sets/:id/agents` | List agents bound to set |

### 4.6 Binding Rules

| Rule | Description |
|---|---|
| Tenant scope | Can only bind sets from same tenant |
| At least one set | Agent must have ≥1 document set for RAG |
| Max 10 sets per agent | Prevents scope explosion |
| No orphan agents | Unbinding last set requires confirmation |
| Audit logged | All binding changes are audited |

---

## 5. Agent-Database Binding (SQL Agent)

### 5.1 Purpose

When creating a SQL agent, it must be bound to specific business databases and tables. The agent can ONLY query the bound databases/tables. Connection credentials are validated before binding is allowed.

### 5.2 Binding States

```
CREDENTIALS_TESTING → CONNECTED → TABLES_DISCOVERED → BOUND → ACTIVE
                          ↓
                     CONNECTION_FAILED
```

| State | Description |
|---|---|
| `CREDENTIALS_TESTING` | Test connection in progress |
| `CONNECTED` | Credentials validated, connection successful |
| `TABLES_DISCOVERED` | Schema introspection complete, tables listed |
| `BOUND` | Agent bound to selected databases/tables |
| `ACTIVE` | Binding live, agent can query |
| `CONNECTION_FAILED` | Credentials invalid or connection refused |

### 5.3 Binding Flow

```
Admin Opens Agent SQL Configuration
  → Enter database credentials:
      host, port, database_name, username, password, ssl_mode
    → Click "Test Connection" button
      → System validates credentials:
          1. TCP connection to host:port
          2. Authentication with username/password
          3. SSL/TLS handshake (if required)
          4. Database exists and is accessible
        → IF FAILED: Show error, allow retry
        → IF SUCCESS:
            → Introspect schema:
                - List all tables
                - List columns per table
                - Identify primary keys
                - Detect relationships
              → Show discovered tables to admin
                → Admin selects which tables agent can access
                  → Admin selects which columns per table (optional)
                    → Save binding
                      → Credentials encrypted and stored
                        → Agent ready for SQL queries
```

### 5.4 Test Connection Button

**Endpoint:**

```
POST /api/v1/agents/:agent_id/sql-connections/test
```

**Request:**

```json
{
  "host": "db.example.com",
  "port": 5432,
  "database_name": "aeroxe_billing_db",
  "username": "readonly_agent",
  "password": "encrypted_password",
  "ssl_mode": "require"
}
```

**Response (Success):**

```json
{
  "status": "connected",
  "server_version": "PostgreSQL 16.0",
  "tables_found": 45,
  "latency_ms": 120
}
```

**Response (Failure):**

```json
{
  "status": "connection_failed",
  "error": "password authentication failed for user \"readonly_agent\"",
  "error_code": "AUTH_FAILED"
}
```

### 5.5 Schema Discovery After Successful Connection

After test connection succeeds, system introspects the database:

**Endpoint:**

```
POST /api/v1/agents/:agent_id/sql-connections/discover
```

**Response:**

```json
{
  "tables": [
    {
      "name": "invoices",
      "columns": [
        {"name": "id", "type": "bigint", "nullable": false},
        {"name": "customer_id", "type": "bigint", "nullable": false},
        {"name": "amount", "type": "numeric", "nullable": false},
        {"name": "status", "type": "varchar", "nullable": false},
        {"name": "invoice_date", "type": "date", "nullable": false}
      ],
      "primary_key": ["id"],
      "row_count_estimate": 125000
    },
    {
      "name": "customers",
      "columns": [
        {"name": "id", "type": "bigint", "nullable": false},
        {"name": "name", "type": "varchar", "nullable": false},
        {"name": "email", "type": "varchar", "nullable": true}
      ],
      "primary_key": ["id"],
      "row_count_estimate": 8500
    }
  ],
  "total_tables": 45
}
```

### 5.6 Table Binding After Discovery

Admin selects tables from discovered list:

**Endpoint:**

```
POST /api/v1/agents/:agent_id/sql-connections/tables
```

**Request:**

```json
{
  "connection_id": 123,
  "tables": [
    {
      "table_name": "invoices",
      "columns": ["id", "customer_id", "amount", "status", "invoice_date"]
    },
    {
      "table_name": "customers",
      "columns": ["id", "name", "email"]
    }
  ]
}
```

### 5.7 Database Credential Storage

| Property | Description |
|---|---|
| Encryption | AES-256-GCM at rest |
| Key management | Vault / KMS |
| Access | Only SQL agent service decrypts |
| Rotation | Supported via re-test flow |
| Audit | All access logged |

### 5.8 Binding Rules

| Rule | Description |
|---|---|
| Test before bind | Connection must pass test before table selection |
| Credential encryption | Passwords encrypted before storage |
| Read-only user | Platform enforces readonly database user |
| Tenant isolation | Can only bind databases for own tenant |
| Max 10 databases per agent | Prevents scope explosion |
| Max 50 tables per database | Prevents excessive scope |
| Admin-only binding | Only tenant admins can configure |
| Re-test on update | Credential changes require new test |

---

## 6. Agent Isolation — Database & Table Scope

### 6.1 Core Principle

**An agent can ONLY query databases and tables it is explicitly bound to. No exceptions.**

### 6.2 SQL Agent Isolation Architecture

```
Agent SQL Request
  → Extract agent_id
    → Query agent_databases for bound connection IDs
      → Query agent_database_tables for bound tables
        → Validate query only references bound tables
          → Execute query against bound connection only
            → Return scoped results
```

### 6.3 Isolation Flow

```
                    +-------------------+
                    |  Agent SQL Query  |
                    +--------+----------+
                             |
                    +--------v----------+
                    |  Extract agent_id  |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Query bindings:    |
                    | agent_databases    |
                    | WHERE agent_id = ? |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Get bound          |
                    | connection_ids:    |
                    | [conn_1, conn_2]   |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Query tables:      |
                    | agent_database_    |
                    | tables WHERE       |
                    | agent_id = ?       |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Validate query:    |
                    | - All tables in    |
                    |   bound list       |
                    | - No destructive   |
                    |   operations       |
                    | - No injection     |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Execute on bound   |
                    | connection only    |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Scoped Results     |
                    +-------------------+
```

### 6.4 Query Validation

Every SQL query from an agent is validated before execution:

| Check | Description |
|---|---|
| Table whitelist | All tables in query exist in agent's bound tables |
| Column whitelist | All columns exist in bound table definitions |
| No cross-database | Query cannot reference multiple databases |
| No destructive ops | SELECT only (no INSERT, UPDATE, DELETE, DROP) |
| No injection | SQL injection patterns blocked |
| Tenant filter | WHERE clause must include tenant_id |
| Row limit | Maximum 10,000 rows returned |
| Timeout | 30 second query timeout |

### 6.5 Query Validation Example

```sql
-- Agent bound to: invoices, customers (in aeroxe_billing_db)
-- Query: SELECT * FROM invoices JOIN customers ON ...

-- VALIDATION:
-- ✅ invoices is in bound tables
-- ✅ customers is in bound tables
-- ✅ JOIN is allowed
-- ✅ SELECT only
-- ✅ tenant_id filter present
-- EXECUTED

-- Query: SELECT * FROM employees
-- ❌ employees is NOT in bound tables
-- BLOCKED: Table 'employees' not in agent scope
```

### 6.6 Multi-Database Agent Example

Agent: "Finance Analytics Agent"

```
Bound Databases:
  1. aeroxe_billing_db (connection: db-billing-01)
     Bound Tables: invoices, payments, credit_notes
  2. aeroxe_erp_db (connection: db-erp-01)
     Bound Tables: products, orders, inventory

NOT bound to:
  - aeroxe_hrms_db (HR data - irrelevant scope)
  - aeroxe_crm_db (CRM data - different department)
  - aeroxe_broadband_db (network ops - irrelevant)

Result: Agent can ONLY query invoices/payments/credit_notes from billing
        and products/orders/inventory from ERP
```

### 6.7 Isolation Enforcement Points

| Layer | Enforcement |
|---|---|
| `nexus-gateway` | Agent SQL request validated |
| `nexus-sql-agent` | Table whitelist check before generation (trait call) |
| LLM Prompt | Schema context only includes bound tables |
| Query Validator | AST analysis against bound schema |
| Connection Router | Execute only on bound connection |
| Result Filter | Strip any leaked metadata |
| `nexus-audit` | All queries logged with scope info (trait call) |

### 6.8 Isolation Violations

| Violation | Detection | Response |
|---|---|---|
| Query unbound table | Table whitelist check | Query rejected + audit log |
| Cross-database join | Connection router check | Query rejected + alert |
| Credential leak | Encryption + access log | Immediate rotation + alert |
| Schema drift | Periodic re-introspection | Alert admin, re-validate |

---

## 7. Complete Flow — End to End

### 5.1 Core Principle

**An agent can ONLY access documents from its bound document sets. No exceptions.**

```
Agent Request
  → Extract agent_id
    → Query agent_document_sets for bound set IDs
      → Filter document_chunks WHERE document_id IN (set's documents)
        → Return only scoped results
```

### 5.2 Isolation Architecture

```
                    +-------------------+
                    |   Agent Request   |
                    +--------+----------+
                             |
                    +--------v----------+
                    |  Extract agent_id  |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Query bindings:    |
                    | agent_document_    |
                    | sets WHERE         |
                    | agent_id = ?       |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Get document_set_  |
                    | ids: [1, 3, 7]     |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Get all document   |
                    | ids from sets:     |
                    | [101,102,103,104]  |
                    +--------+----------+
                             |
                    +--------v----------+
                    | RAG search with    |
                    | WHERE document_id  |
                    | IN (101,102,103,   |
                    | 104)               |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Scoped Results     |
                    | Only bound docs    |
                    +-------------------+
```

### 5.3 Database Query (RAG Search with Isolation)

```sql
-- Get document IDs from agent's bound document sets
WITH agent_scope AS (
    SELECT dsd.document_id
    FROM agent_document_sets ads
    JOIN document_set_documents dsd ON dsd.document_set_id = ads.document_set_id
    WHERE ads.agent_id = $1
      AND ads.tenant_id = $2
)
-- Search only within scoped documents
SELECT dc.id, dc.content, dc.metadata,
       dc.embedding <=> $3 AS distance
FROM document_chunks dc
WHERE dc.document_id IN (SELECT document_id FROM agent_scope)
ORDER BY distance
LIMIT $4;
```

### 5.4 Isolation Enforcement Points

| Layer | Enforcement |
|---|---|
| `nexus-gateway` | Agent request validated against binding |
| `nexus-agent` | Scope injected into execution context |
| `nexus-rag` | Query filtered by document set scope (trait call) |
| Vector Search | pgvector query scoped to bound documents |
| `nexus-memory` | Agent memory scoped to tenant + sets (trait call) |
| `nexus-audit` | All access logged with scope info (trait call) |

### 5.5 Isolation Violations

| Violation | Detection | Response |
|---|---|---|
| Agent queries unbound doc | RAG query filter blocks | Empty results + audit log |
| Cross-tenant document access | Tenant middleware blocks | 403 Forbidden + alert |
| Manual scope bypass | Schema-level enforcement | Query rejected at DB |
| Scope tampering | Binding immutable without auth | Audit + admin alert |

### 5.6 Multi-Set Agent Example

Agent: "Customer Support Agent"

```
Bound Document Sets:
  1. "Product Manuals" (read)
  2. "FAQ Database" (read)
  3. "Troubleshooting Guides" (read)

NOT bound to:
  - "Financial Reports" (different department)
  - "HR Policies" (irrelevant scope)
  - "Legal Documents" (restricted)

Result: Agent can ONLY answer from Product Manuals, FAQ, and Troubleshooting
```

---

## 8. Complete Flow — End to End

### 8.1 Tenant Onboarding to Agent Deployment

```
1. Tenant Registers
   → Account created (status: pending_kyc)

2. Tenant Completes KYC
   → Upload business docs, ID, tax info
   → Submit for review
   → Admin approves
   → Account status: active

3. Tenant Creates Document Sets
   → "Product Knowledge Base" (draft)
   → Upload product manuals, guides
   → Process documents (chunking + embedding)
   → Activate set (status: active)

4. Tenant Creates Agent
   → "Product Support Agent"
   → Configure model, system prompt, tools

5. Admin Binds Agent to Document Sets
   → Bind to "Product Knowledge Base" (read)
   → Agent scope defined

6. Admin Binds Agent to Databases (SQL Agent)
   → Enter database credentials
   → Click "Test Connection" → Success
   → Discover tables → Select tables
   → Save binding
   → Agent can ONLY query bound tables

7. Agent Handles Queries
   → User asks product question
   → Agent orchestrator checks document set bindings
   → RAG searches ONLY within bound sets
   → User asks SQL question
   → SQL agent checks database bindings
   → Query validated against bound tables only
   → Response generated from scoped data
   → Audit log records all access
```

---

## 9. NATS Events

### KYC Events

| Subject | Event | When |
|---|---|---|
| `aeroxe.v1.kyc.submitted` | `KYCDocumentsSubmitted` | Documents uploaded |
| `aeroxe.v1.kyc.approved` | `KYCApproved` | Verification passed |
| `aeroxe.v1.kyc.rejected` | `KYCRejected` | Verification failed |

### Document Set Events

| Subject | Event | When |
|---|---|---|
| `aeroxe.v1.docset.created` | `DocumentSetCreated` | Set created |
| `aeroxe.v1.docset.activated` | `DocumentSetActivated` | Set activated |
| `aeroxe.v1.docset.document.added` | `DocumentAddedToSet` | Doc linked |
| `aeroxe.v1.docset.document.removed` | `DocumentRemovedFromSet` | Doc unlinked |
| `aeroxe.v1.docset.archived` | `DocumentSetArchived` | Set archived |

### Agent Document Binding Events

| Subject | Event | When |
|---|---|---|
| `aeroxe.v1.agent.bound` | `AgentBoundToDocumentSet` | Binding created |
| `aeroxe.v1.agent.unbound` | `AgentUnboundFromDocumentSet` | Binding removed |

### Agent Database Binding Events

| Subject | Event | When |
|---|---|---|
| `aeroxe.v1.agent.db.test.success` | `AgentDBConnectionTested` | Test connection passed |
| `aeroxe.v1.agent.db.test.failed` | `AgentDBConnectionTestFailed` | Test connection failed |
| `aeroxe.v1.agent.db.bound` | `AgentBoundToDatabase` | Database binding created |
| `aeroxe.v1.agent.db.unbound` | `AgentUnboundFromDatabase` | Database binding removed |
| `aeroxe.v1.agent.db.table.bound` | `AgentBoundToTable` | Table binding created |
| `aeroxe.v1.agent.db.table.unbound` | `AgentUnboundFromTable` | Table binding removed |
| `aeroxe.v1.agent.db.schema.drift` | `AgentDBSchemaDrift` | Schema changed after binding |
| `aeroxe.v1.agent.scope.changed` | `AgentScopeChanged` | Permissions updated |

### Customer Module Events

| Subject | Event | When |
|---|---|---|
| `aeroxe.v1.customer.customer.created` | `CustomerCreated` | Customer created |
| `aeroxe.v1.customer.customer.suspended` | `CustomerSuspended` | Customer suspended |
| `aeroxe.v1.customer.customer.activated` | `CustomerActivated` | Customer activated |

---

## 10. Observability

### Metrics

| Metric | Labels | Description |
|---|---|---|
| `kyc_documents_uploaded_total` | tenant_id, doc_type | KYC uploads |
| `kyc_review_duration_seconds` | tenant_id | Review time |
| `kyc_status_changes_total` | tenant_id, from, to | KYC transitions |
| `document_sets_created_total` | tenant_id | Sets created |
| `document_set_documents_total` | set_id | Docs per set |
| `agent_doc_bindings_total` | agent_id, set_id | Document set bindings |
| `agent_db_bindings_total` | agent_id, db_name | Database bindings |
| `agent_db_table_bindings_total` | agent_id, table_name | Table bindings |
| `agent_db_test_connections_total` | agent_id, result | Test connection results |
| `agent_db_query_scoped_total` | agent_id, db_name | Scoped SQL queries |
| `agent_db_query_violations_total` | agent_id, violation_type | Blocked SQL violations |
| `agent_scope_queries_total` | agent_id, result | Scoped RAG queries |
| `agent_scope_violations_total` | agent_id | Blocked RAG violations |

### Grafana Dashboard Panels

| Panel | Description |
|---|---|
| KYC Pipeline Status | Tenants by KYC state |
| Document Set Health | Sets by status, doc count |
| Agent Document Binding Map | Agents → Document Sets |
| Agent Database Binding Map | Agents → Databases → Tables |
| SQL Query Scope Latency | Scoped SQL query time |
| DB Connection Test Results | Pass/fail rates |
| Isolation Violations | Blocked unauthorized access |

---

## 11. Security

| Control | Implementation |
|---|---|
| KYC before access | Middleware blocks pre-KYC tenants |
| Document set ownership | tenant_id enforced at DB level |
| Database credential encryption | AES-256-GCM at rest |
| Test before bind | Connection must pass before table selection |
| Agent scope enforcement | Query-level filtering, not just middleware |
| Binding audit trail | All changes logged to audit_events |
| No cross-tenant sets | DB constraint + service validation |
| No cross-tenant databases | Database binding tenant-scoped |
| Admin-only binding | Only tenant admins can bind agents |
| Read-only DB user | Platform enforces readonly database access |
