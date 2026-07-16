# 21 — Audit Trail & Compliance

## Comprehensive Audit Logging + Regulatory Compliance + Data Governance

---

## 1. Purpose

AeroXe Nexus AI maintains complete audit trails for:

- Regulatory compliance
- Security forensics
- Operational debugging
- Tenant accountability
- AI decision traceability

---

## 2. Audit Architecture

```
Application Layer → Audit Event → NATS JetStream → Audit Service → Elasticsearch + PostgreSQL
```

---

## 3. Audit Event Schema

```json
{
  "event_id": "uuid-v4",
  "timestamp": "2025-01-15T10:30:00Z",
  "event_type": "agent.tool.execute",
  "actor": {
    "user_id": "user-123",
    "tenant_id": "tenant-aeroxe",
    "role": "admin",
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0"
  },
  "resource": {
    "type": "agent",
    "id": "agent-456",
    "name": "Customer Support Agent"
  },
  "action": "tool.execute",
  "result": "success",
  "details": {
    "tool": "database.query",
    "query": "SELECT * FROM customers WHERE id = $1",
    "params": ["12345"],
    "duration_ms": 45
  },
  "trace_id": "abc-123-def-456",
  "request_id": "req-789"
}
```

---

## 4. Audit Event Types

### Authentication Events

| Event Type                     | Description                    |
| ------------------------------ | ------------------------------ |
| auth.login.success             | Successful login               |
| auth.login.failure             | Failed login attempt           |
| auth.logout                    | User logout                    |
| auth.token.refresh             | Token refreshed                |
| auth.token.revoked             | Token revoked                  |
| auth.password.changed          | Password changed               |
| auth.mfa.enabled               | MFA enabled                    |
| auth.mfa.disabled              | MFA disabled                   |
| auth.session.expired           | Session expired                |

### Authorization Events

| Event Type                     | Description                    |
| ------------------------------ | ------------------------------ |
| authz.permission.granted       | Permission granted             |
| authz.permission.denied        | Permission denied              |
| authz.role.assigned            | Role assigned                  |
| authz.role.revoked             | Role revoked                   |
| authz.tenant.violation         | Cross-tenant access attempt    |

### AI Agent Events

| Event Type                     | Description                    |
| ------------------------------ | ------------------------------ |
| agent.created                  | Agent created                  |
| agent.started                  | Agent session started          |
| agent.completed                | Agent session completed        |
| agent.failed                   | Agent session failed           |
| agent.tool.execute             | Tool execution requested       |
| agent.tool.approved            | Tool execution approved        |
| agent.tool.denied              | Tool execution denied          |
| agent.planning.start           | Planning phase started         |
| agent.planning.complete        | Planning phase completed       |

### Data Events

| Event Type                     | Description                    |
| ------------------------------ | ------------------------------ |
| data.read                      | Data read                      |
| data.write                     | Data written                   |
| data.delete                    | Data deleted                   |
| data.export                    | Data exported                  |
| data.import                    | Data imported                  |
| data.upload                    | File uploaded                  |
| data.download                  | File downloaded                |

### Security Events

| Event Type                     | Description                    |
| ------------------------------ | ------------------------------ |
| security.prompt.injection     | Prompt injection detected      |
| security.data.leakage         | Data leakage attempt           |
| security.unauthorized.access  | Unauthorized access attempt    |
| security.tenant.violation     | Tenant boundary violation      |
| security.suspicious.activity  | Suspicious activity detected   |

### Workflow Events

| Event Type                     | Description                    |
| ------------------------------ | ------------------------------ |
| workflow.created               | Workflow created               |
| workflow.started               | Workflow started               |
| workflow.step.completed        | Workflow step completed        |
| workflow.approval.requested    | Approval requested             |
| workflow.approval.granted      | Approval granted               |
| workflow.approval.denied       | Approval denied                |
| workflow.completed             | Workflow completed             |
| workflow.failed                | Workflow failed                |

### Integration Events

| Event Type                     | Description                    |
| ------------------------------ | ------------------------------ |
| integration.connected          | External system connected      |
| integration.disconnected       | External system disconnected   |
| integration.sync.start         | Data sync started              |
| integration.sync.complete      | Data sync completed            |
| integration.error              | Integration error              |

---

## 5. Audit Service Implementation

### Service Identity

| Property     | Value                              |
| ------------ | ---------------------------------- |
| Language     | Go                                 |
| Port         | 50060 (gRPC), 8060 (REST)         |
| Database     | PostgreSQL + Elasticsearch         |
| Storage      | Hot (30d) → Warm (90d) → Cold (1y)|

### Responsibilities

- Ingest audit events from all services
- Validate event schema
- Index events in Elasticsearch
- Store events in PostgreSQL
- Provide audit query API
- Generate compliance reports
- Detect anomalies

---

## 6. Audit Database Schema

### chat_trail Table (Partitioned)

Every message in every conversation is recorded with full context.

```sql
CREATE TABLE chat_trail (
    id              BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id       BIGINT NOT NULL,
    session_id      VARCHAR(100) NOT NULL,
    conversation_id VARCHAR(100) NOT NULL,
    message_id      VARCHAR(100) NOT NULL,
    customer_id     BIGINT,
    user_id         BIGINT,
    role            VARCHAR(20) NOT NULL,           -- user | assistant | system | tool
    content         TEXT NOT NULL,
    content_type    VARCHAR(50) DEFAULT 'text',     -- text | image | audio | tool_call | tool_result
    model_used      VARCHAR(100),                   -- which AI model responded
    tokens_input    INTEGER,
    tokens_output   INTEGER,
    latency_ms      INTEGER,
    tool_name       VARCHAR(100),                   -- tool invoked (if any)
    tool_input      JSONB,                          -- tool call arguments
    tool_output     JSONB,                          -- tool call result
    tool_status     VARCHAR(20),                    -- success | error | blocked
    safety_flag     VARCHAR(50),                    -- none | prompt_injection | jailbreak | policy_violation
    metadata        JSONB,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Monthly partitions (auto-created by pg_partman)
CREATE TABLE chat_trail_2026_01 PARTITION OF chat_trail
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE chat_trail_2026_02 PARTITION OF chat_trail
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

-- Indexes
CREATE INDEX idx_ct_tenant_session ON chat_trail(tenant_id, session_id, created_at DESC);
CREATE INDEX idx_ct_conversation ON chat_trail(conversation_id, created_at);
CREATE INDEX idx_ct_customer ON chat_trail(customer_id, created_at DESC);
CREATE INDEX idx_ct_safety ON chat_trail(safety_flag) WHERE safety_flag != 'none';
```

### audit_events Table (Partitioned)

```sql
CREATE TABLE audit_events (
    id              BIGINT GENERATED ALWAYS AS IDENTITY,
    tenant_id       BIGINT NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    actor_user_id   BIGINT,
    actor_role      VARCHAR(50),
    actor_ip        INET,
    actor_user_agent TEXT,
    resource_type   VARCHAR(50),
    resource_id     BIGINT,
    action          VARCHAR(100) NOT NULL,
    result          VARCHAR(20) NOT NULL,           -- success | failure | blocked
    details         JSONB,
    trace_id        VARCHAR(100),
    request_id      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE audit_events_2026_01 PARTITION OF audit_events
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE audit_events_2026_02 PARTITION OF audit_events
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

-- Indexes
CREATE INDEX idx_audit_tenant ON audit_events(tenant_id, created_at DESC);
CREATE INDEX idx_audit_event_type ON audit_events(event_type, created_at DESC);
CREATE INDEX idx_audit_actor ON audit_events(actor_user_id, created_at DESC);
CREATE INDEX idx_audit_trace ON audit_events(trace_id);
```

### Partition Management (pg_partman)

```sql
-- Auto-create next 3 months of partitions
SELECT partman.create_parent(
    p_parent_table := 'public.chat_trail',
    p_control := 'created_at',
    p_type := 'range',
    p_premake := 3
);

SELECT partman.create_parent(
    p_parent_table := 'public.audit_events',
    p_control := 'created_at',
    p_type := 'range',
    p_premake := 3
);

-- Retention: drop partitions older than 2 years
UPDATE partman.part_config
SET retention = '2 years', retention_keep_table = false
WHERE parent_table = 'public.chat_trail';

UPDATE partman.part_config
SET retention = '2 years', retention_keep_table = false
WHERE parent_table = 'public.audit_events';
```

### audit_retention Policies

```sql
-- Hot: 0-3 months (NVMe, full detail, no compression)
-- Warm: 3-12 months (SSD, LZ4 compression)
-- Cold: 1-2 years (HDD, ZSTD compression)
-- Archive: 2+ years (exported to MinIO, deleted from PG)
-- Permanent: compliance-critical events (never deleted)
```

### chat_trail Usage

| Use Case | Query |
|---|---|
| Conversation replay | `SELECT * FROM chat_trail WHERE conversation_id = ? ORDER BY created_at` |
| Customer chat history | `SELECT * FROM chat_trail WHERE customer_id = ? AND created_at > ?` |
| Tool usage analytics | `SELECT tool_name, COUNT(*), AVG(latency_ms) FROM chat_trail WHERE tool_name IS NOT NULL GROUP BY tool_name` |
| Model performance | `SELECT model_used, AVG(tokens_output), AVG(latency_ms) FROM chat_trail WHERE role = 'assistant' GROUP BY model_used` |
| Safety incidents | `SELECT * FROM chat_trail WHERE safety_flag != 'none' AND created_at > ?` |
| Compliance export | `SELECT to_jsonb(chat_trail.*) FROM chat_trail WHERE conversation_id = ?` |

---

## 7. Compliance Requirements

### Data Protection

| Regulation    | Requirements                                    |
| ------------- | ----------------------------------------------- |
| GDPR          | Data minimization, right to erasure, consent    |
| SOC 2         | Access controls, audit logging, encryption      |
| ISO 27001     | Information security management                 |
| HIPAA         | PHI protection (if healthcare integration)      |

### Compliance Checks

| Check                              | Frequency   |
| ---------------------------------- | ----------- |
| Access control review              | Weekly      |
| Audit log integrity                | Daily       |
| Encryption key rotation            | Monthly     |
| Penetration testing                | Quarterly   |
| Compliance audit                   | Annually    |
| Disaster recovery test             | Quarterly   |

---

## 8. Data Governance

### Data Classification

| Level        | Description                    | Examples                     |
| ------------ | ------------------------------ | ---------------------------- |
| Public       | Non-sensitive                  | Marketing content            |
| Internal     | Business operations            | Employee directory           |
| Confidential | Business sensitive             | Financial reports            |
| Restricted   | Highly sensitive               | PII, payment data, health    |

### Data Retention Policy

| Data Type          | Retention Period | Archive Location |
| ------------------ | ---------------- | ---------------- |
| User data          | Active + 7 years | MinIO Cold       |
| Audit logs         | 1 year           | MinIO Cold       |
| AI conversations   | Active + 1 year  | Elasticsearch    |
| Session data       | 30 days          | Deleted          |
| Backup data        | 90 days          | MinIO            |

---

## 9. Audit Query API

### Query Audit Events

```
GET /api/v1/audit/events
Query params:
  - tenant_id
  - event_type
  - actor_user_id
  - resource_type
  - resource_id
  - start_date
  - end_date
  - result
  - page, limit
```

### Get Audit Event Detail

```
GET /api/v1/audit/events/:event_id
```

### Generate Compliance Report

```
POST /api/v1/audit/reports
Body:
{
  "report_type": "access_summary",
  "start_date": "2025-01-01",
  "end_date": "2025-01-31",
  "tenant_id": "tenant-aeroxe"
}
```

### Export Audit Data

```
POST /api/v1/audit/export
Body:
{
  "format": "csv",
  "filters": {...}
}
```

---

## 10. Anomaly Detection

### Automatic Detection Rules

| Rule                              | Alert Level |
| --------------------------------- | ----------- |
| 10+ failed logins in 5 minutes   | HIGH        |
| Cross-tenant access attempt       | CRITICAL    |
| Prompt injection detected         | CRITICAL    |
| Unusual data export volume        | HIGH        |
| Off-hours administrative access   | MEDIUM      |
| Privilege escalation attempt      | CRITICAL    |
| Tool execution without approval   | HIGH        |

### Response Actions

| Alert Level | Action                                 |
| ----------- | -------------------------------------- |
| LOW         | Log only                              |
| MEDIUM      | Log + Notify admin                     |
| HIGH        | Log + Notify admin + Block IP          |
| CRITICAL    | Log + Notify admin + Block + Lock account |

---

## 11. AI Decision Traceability

### Why It Matters

AI agents make autonomous decisions. Every decision must be traceable.

### Traceability Chain

```
User Request
  → AI Gateway (routing decision)
    → Agent Orchestrator (planning)
      → Model Selection (which model)
        → Tool Execution (which tools)
          → Response Generation
            → Output to User
```

### Stored per AI Decision

| Field                 | Description                        |
| --------------------- | ---------------------------------- |
| input_prompt          | Original user input                |
| model_used            | Which model processed the request  |
| tools_called          | List of tools invoked              |
| tool_parameters       | Parameters sent to each tool       |
| tool_results          | Results returned by tools          |
| reasoning_trace       | Model reasoning steps              |
| final_response        | Response sent to user              |
| confidence_score      | Model confidence                   |
| safety_checks_passed  | Prompt injection checks            |
| latency_ms            | Total processing time              |
| tokens_used           | Token consumption                  |

---

## 12. Multi-Tenant Audit Isolation

### Requirements

Each tenant's audit data is completely isolated.

### Implementation

- Tenant ID in every audit event row
- Row-level security on audit tables
- Cross-tenant queries blocked at service level
- Audit export includes only requesting tenant's data

### Test

```
Tenant A queries audit → Only sees tenant_id=A events
Tenant B queries audit → Only sees tenant_id=B events
Admin queries audit → Sees all events with tenant filter
```

---

## 13. Audit Log Integrity

### Tamper Protection

| Mechanism                         | Purpose                         |
| --------------------------------- | ------------------------------- |
| Append-only audit table           | Prevent modification            |
| Cryptographic hash chain          | Detect tampering                |
| Regular integrity checks          | Verify hash chain               |
| Separate audit storage            | Isolate from application data   |

### Hash Chain

```
Event N hash = SHA256(event_N_data + event_(N-1)_hash)
```

---

## 14. Compliance Reports

### Standard Reports

| Report                      | Frequency   | Audience        |
| --------------------------- | ----------- | --------------- |
| Access Summary              | Weekly      | Security team   |
| Security Incident Report    | As needed   | CISO            |
| Data Access Report          | Monthly     | Compliance      |
| AI Decision Audit           | Monthly     | AI team         |
| Tenant Activity Report      | Monthly     | Account managers|
| Regulatory Compliance       | Annually    | Auditors        |

---

## 15. Audit Monitoring Dashboard

### Grafana Dashboard Panels

| Panel                       | Description                    |
| --------------------------- | ------------------------------ |
| Events per second           | Real-time event throughput     |
| Failed authentications      | Security monitoring            |
| Cross-tenant violations     | Isolation enforcement          |
| AI tool executions          | Agent activity                 |
| Prompt injection attempts   | AI safety                      |
| Data export volume          | Data governance                |
| Top actors by activity      | Usage patterns                 |
| Error rate by service       | Reliability                    |
