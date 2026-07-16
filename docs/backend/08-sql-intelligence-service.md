# AeroXe Nexus AI — SQL Intelligence Service

## Natural Language SQL Generation, Safe Query Execution & Business Analytics

---

## 1. Service Identity

| Attribute | Value |
|---|---|
| Service Name | `sql-agent-service` |
| Bounded Context | SQL Intelligence |
| Domain Type | Core Domain |
| Language | Go |
| AI Model | `qwen2.5-coder:3b` (Ollama) |
| gRPC Port | 50056 |

---

## 2. Purpose

The SQL Intelligence Service enables natural language business intelligence. Users ask questions in plain English, and the service generates safe, validated SQL queries against business databases, returning results with explanations.

---

## 3. Aggregate Design

### QueryExecution Aggregate

```
QueryExecution (Aggregate Root)
├── GeneratedSQL
│   ├── RawSQL
│   ├── ParsedAST
│   └── ValidationStatus
├── ValidationResult
│   ├── IsSafe
│   ├── BlockedOperations
│   └── Permissions
└── ResultSet
    ├── Columns[]
    ├── Rows[]
    └── Summary
```

### Value Objects

| Value Object | Type |
|---|---|
| `DatabaseId` | i64 |
| `SQLStatement` | Validated SQL string |
| `QueryPermission` | User role + database access |

---

## 4. gRPC Contract

```protobuf
syntax = "proto3";
package aeroxe.sql;

service SQLService {
  rpc GenerateQuery(QueryRequest) returns (SQLResponse);
  rpc ExecuteQuery(SQLRequest) returns (ResultResponse);
  rpc ExplainQuery(ExplainRequest) returns (ExplainResponse);
  rpc StreamQuery(SQLRequest) returns (stream ResultChunk);
}

message QueryRequest {
  string question = 1;
  string database = 2;
  string tenant_id = 3;
  string user_id = 4;
}

message SQLResponse {
  string sql = 1;
  string explanation = 2;
  bool is_safe = 3;
  repeated string warnings = 4;
}

message SQLRequest {
  string sql = 1;
  string database = 2;
  string tenant_id = 3;
}

message ResultResponse {
  repeated ColumnInfo columns = 1;
  repeated Row rows = 2;
  int32 row_count = 3;
  float execution_time_ms = 4;
}
```

---

## 5. Query Processing Pipeline

```
User Question: "Show monthly revenue for 2026"
    |
    v
[1] Intent Classification
    |  - Data query vs. operational request
    |  - Identify target database
    |
    v
[2] Schema Context Assembly
    |  - Load relevant table schemas
    |  - Include column descriptions
    |  - Add relationships
    |
    v
[3] SQL Generation (LLM)
    |  - Model: qwen2.5-coder:3b
    |  - System prompt with schema context
    |  - Generate SQL from natural language
    |
    v
[4] SQL Validation
    |  - Parse SQL AST
    |  - Check allowed operations
    |  - Verify table/column existence
    |  - Check for injection patterns
    |
    v
[5] Permission Check
    |  - User role verification
    |  - Table-level access control
    |  - Column-level masking (if applicable)
    |
    v
[6] Execute Against Read Replica
    |  - NEVER against primary
    |  - Query timeout: 30s
    |  - Row limit: 10,000
    |
    v
[7] Result Post-processing
    |  - Format for display
    |  - Generate explanation
    |  - Create summary
    |
    v
[8] Audit Event
    |  - Log query + execution + user
```

---

## 6. SQL Safety Rules

### Agent Scope Isolation

Every query is validated against the agent's bound databases and tables:

| Check | Description |
|---|---|
| Table whitelist | All tables in query must be in agent's bound tables |
| Column whitelist | All columns must exist in bound table definitions |
| No cross-database | Query cannot reference multiple databases |
| Tenant filter | WHERE clause must include tenant_id |

### Allowed Operations

| Operation | Status |
|---|---|
| SELECT | Allowed |
| JOIN | Allowed |
| GROUP BY | Allowed |
| ORDER BY | Allowed |
| WHERE | Allowed |
| HAVING | Allowed |
| COUNT, SUM, AVG, MIN, MAX | Allowed |
| Subqueries | Allowed (with timeout) |
| CTEs (WITH clause) | Allowed |

### Blocked Operations

| Operation | Status | Reason |
|---|---|---|
| DELETE | BLOCKED | Destructive |
| UPDATE | BLOCKED | Destructive |
| INSERT | BLOCKED | Destructive |
| DROP | BLOCKED | Destructive |
| ALTER | BLOCKED | Schema change |
| TRUNCATE | BLOCKED | Destructive |
| CREATE | BLOCKED | Schema change |
| GRANT | BLOCKED | Security |
| REVOKE | BLOCKED | Security |

### Injection Protection

| Pattern | Action |
|---|---|
| `' OR 1=1 --` | Blocked |
| `'; DROP TABLE` | Blocked |
| `UNION SELECT` | Allowed only if safe columns |
| Comment injection (`--`, `/*`) | Stripped and validated |
| Multiple statements (`;`) | Blocked |

---

## 7. LLM Prompt Engineering

### System Prompt Template

```
You are a SQL expert for the AeroXe database. Generate ONLY SELECT queries.

Database Schema:
{schema_context}

Rules:
- Only generate SELECT statements
- Use proper JOIN syntax
- Include appropriate WHERE clauses for tenant isolation
- Use aggregate functions when appropriate
- Never use DELETE, UPDATE, INSERT, DROP, ALTER, TRUNCATE
- Always qualify column names with table aliases
- Use LIMIT for large result sets

User Question: {question}
```

---

## 8. Example Queries

### User: "Show monthly revenue for 2026"

**Generated SQL:**
```sql
SELECT
    DATE_TRUNC('month', invoice_date) AS month,
    SUM(amount) AS total_revenue,
    COUNT(*) AS invoice_count
FROM invoices
WHERE tenant_id = $1
  AND invoice_date >= '2026-01-01'
  AND invoice_date < '2027-01-01'
GROUP BY DATE_TRUNC('month', invoice_date)
ORDER BY month;
```

### User: "Which customers have pending payments?"

**Generated SQL:**
```sql
SELECT
    c.customer_id,
    c.name,
    c.email,
    SUM(i.amount) AS pending_amount
FROM customers c
JOIN invoices i ON c.customer_id = i.customer_id
WHERE c.tenant_id = $1
  AND i.status = 'pending'
GROUP BY c.customer_id, c.name, c.email
ORDER BY pending_amount DESC;
```

---

## 9. Database Connection

### Agent-Database Binding

Each SQL agent is bound to specific databases and tables. The agent can ONLY query bound resources.

### Connection Flow

```
Admin enters database credentials
  → Click "Test Connection"
    → System validates credentials
      → IF FAILED: Show error, allow retry
      → IF SUCCESS: Introspect schema
        → Admin selects tables
          → Binding saved
            → Agent ready
```

### Test Connection Button

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
  "server_version": "PostgreSQL 18.0",
  "tables_found": 45,
  "latency_ms": 120
}
```

**Response (Failure):**

```json
{
  "status": "connection_failed",
  "error": "password authentication failed",
  "error_code": "AUTH_FAILED"
}
```

### Schema Discovery

After successful connection, system introspects schema:

```
POST /api/v1/agents/:agent_id/sql-connections/discover
```

Returns all tables, columns, primary keys, and row count estimates.

### Table Binding

Admin selects tables from discovered list:

```
POST /api/v1/agents/:agent_id/sql-connections/tables
```

### Read Replica Strategy

| Operation | Target |
|---|---|
| All SQL queries | Read Replica (never primary) |
| Connection pooling | EntORM pool (max 10 connections per tenant) |
| Query timeout | 30 seconds |
| Row limit | 10,000 rows |

### Supported Databases

The SQL Agent can query any PostgreSQL database connected to the platform:

| Database | Purpose |
|---|---|
| `aeroxe_broadband_db` | ISP operations |
| `aeroxe_erp_db` | Enterprise resource planning |
| `aeroxe_crm_db` | Customer relationship management |
| `aeroxe_billing_db` | Billing and payments |
| `aeroxe_hrms_db` | Human resources |

---

## 10. REST API Endpoints

### Generate SQL

```
POST /api/v1/sql/generate
```

**Request:**
```json
{
  "question": "Show monthly revenue",
  "database": "aeroxe_billing_db"
}
```

**Response:**
```json
{
  "sql": "SELECT DATE_TRUNC('month', invoice_date)...",
  "explanation": "This query groups invoices by month and calculates total revenue...",
  "is_safe": true,
  "warnings": []
}
```

### Execute Query

```
POST /api/v1/sql/query
```

**Request:**
```json
{
  "question": "Show monthly revenue",
  "database": "aeroxe_billing_db"
}
```

**Response:**
```json
{
  "sql": "SELECT DATE_TRUNC('month', invoice_date)...",
  "data": [
    { "month": "2026-01-01", "total_revenue": 500000, "invoice_count": 120 },
    { "month": "2026-02-01", "total_revenue": 620000, "invoice_count": 145 }
  ],
  "row_count": 7,
  "execution_time_ms": 45.2
}
```

---

## 11. Observability

| Metric | Description |
|---|---|
| `sql_queries_generated_total` | Queries generated by LLM |
| `sql_queries_executed_total` | Queries executed against database |
| `sql_queries_blocked_total` | Blocked safety violations |
| `sql_query_execution_ms` | Query execution time |
| `sql_llm_generation_ms` | LLM SQL generation time |
| `sql_injection_attempts_total` | Injection attempts detected |
