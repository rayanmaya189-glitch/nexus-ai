# AeroXe Nexus AI — Tenant KYC, Document Sets & Agent Isolation

## Tenant Onboarding + KYC Verification + Document Set Management + Agent-Document Binding

---

## 1. Overview

This document defines four critical flows:

1. **Tenant KYC Flow** — Verification before platform access
2. **Document Set Creation** — Organizing documents into scoped collections
3. **Agent-Document Set Binding** — Connecting agents to specific document sets
4. **Agent Isolation** — Enforcing that agents access ONLY their bound document sets

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
| PUT | `/api/v1/document-sets/:id` | Update set metadata |
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
| PUT | `/api/v1/agents/:id/document-sets/:set_id` | Update binding permissions |
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

## 5. Agent Isolation — Document Set Scope

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
| API Gateway | Agent request validated against binding |
| Agent Orchestrator | Scope injected into execution context |
| RAG Service | Query filtered by document set scope |
| Vector Search | pgvector query scoped to bound documents |
| Memory Service | Agent memory scoped to tenant + sets |
| Audit | All access logged with scope info |

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

## 6. Complete Flow — End to End

### 6.1 Tenant Onboarding to Agent Deployment

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

6. Agent Handles Queries
   → User asks product question
   → Agent orchestrator checks bindings
   → RAG searches ONLY within bound sets
   → Response generated from scoped documents
   → Audit log records access
```

---

## 7. NATS Events

### KYC Events

| Subject | Event | When |
|---|---|---|
| `aeroxe.kyc.submitted` | `KYCDocumentsSubmitted` | Documents uploaded |
| `aeroxe.kyc.approved` | `KYCApproved` | Verification passed |
| `aeroxe.kyc.rejected` | `KYCRejected` | Verification failed |

### Document Set Events

| Subject | Event | When |
|---|---|---|
| `aeroxe.docset.created` | `DocumentSetCreated` | Set created |
| `aeroxe.docset.activated` | `DocumentSetActivated` | Set activated |
| `aeroxe.docset.document.added` | `DocumentAddedToSet` | Doc linked |
| `aeroxe.docset.document.removed` | `DocumentRemovedFromSet` | Doc unlinked |
| `aeroxe.docset.archived` | `DocumentSetArchived` | Set archived |

### Agent Binding Events

| Subject | Event | When |
|---|---|---|
| `aeroxe.agent.bound` | `AgentBoundToDocumentSet` | Binding created |
| `aeroxe.agent.unbound` | `AgentUnboundFromDocumentSet` | Binding removed |
| `aeroxe.agent.scope.changed` | `AgentScopeChanged` | Permissions updated |

---

## 8. Observability

### Metrics

| Metric | Labels | Description |
|---|---|---|
| `kyc_documents_uploaded_total` | tenant_id, doc_type | KYC uploads |
| `kyc_review_duration_seconds` | tenant_id | Review time |
| `kyc_status_changes_total` | tenant_id, from, to | KYC transitions |
| `document_sets_created_total` | tenant_id | Sets created |
| `document_set_documents_total` | set_id | Docs per set |
| `agent_bindings_total` | agent_id, set_id | Active bindings |
| `agent_scope_queries_total` | agent_id, result | Scoped RAG queries |
| `agent_scope_violations_total` | agent_id | Blocked violations |

### Grafana Dashboard Panels

| Panel | Description |
|---|---|
| KYC Pipeline Status | Tenants by KYC state |
| Document Set Health | Sets by status, doc count |
| Agent Binding Map | Agents → Document Sets |
| Scope Query Latency | RAG query time with scoping |
| Isolation Violations | Blocked unauthorized access |

---

## 9. Security

| Control | Implementation |
|---|---|
| KYC before access | Middleware blocks pre-KYC tenants |
| Document set ownership | tenant_id enforced at DB level |
| Agent scope enforcement | Query-level filtering, not just middleware |
| Binding audit trail | All changes logged to audit_events |
| No cross-tenant sets | DB constraint + service validation |
| Admin-only binding | Only tenant admins can bind agents |
