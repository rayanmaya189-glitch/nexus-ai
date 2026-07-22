# AeroXe Nexus AI — Agent Orchestrator Module

## AI Agent Lifecycle, Planning, Tool Execution & Orchestration

> **Modular Monolith Module:** This document describes the `nexus-agent` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces (see [Communication Architecture](12-communication-architecture.md)).

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-agent` |
| Crate | `nexus-agent` (workspace member) |
| Bounded Context | Agent Orchestration |
| Domain Type | Core Domain |
| Language | Rust |
| Schema | `agent_` (in shared PostgreSQL) |
| Dependencies | `nexus-rag` (RagService trait), `nexus-memory` (MemoryService trait) + Ollama |

---

## 2. Purpose

The Agent Orchestrator module is the brain of the AI platform. It manages:

- Agent lifecycle (creation, execution, completion, failure)
- Task planning and decomposition
- Tool selection and invocation
- Context management across multi-step executions
- Agent routing to specialized models
- Result aggregation from multiple agents

---

## 3. Aggregate Design

### AgentExecution Aggregate

```
AgentExecution (Aggregate Root)
├── Task
│   ├── TaskId
│   ├── Description
│   ├── Priority
│   └── Status
├── Plan
│   ├── Steps[]
│   └── Dependencies
├── ToolExecution[]
│   ├── ToolName
│   ├── Parameters
│   ├── Result
│   └── Status
└── Result
    ├── Output
    ├── TokensUsed
    └── LatencyMs
```

### Entities

| Entity | Attributes |
|---|---|
| Agent | AgentId, Name, Type, Capabilities[], Model, SystemPrompt |
| Task | TaskId, Description, Status, Priority, CreatedAt |
| ExecutionStep | StepId, ExecutionId, StepNumber, Action, Result, Status |

### Value Objects

| Value Object | Type | Description |
|---|---|---|
| `AgentId` | i64 | Unique agent identifier |
| `TaskId` | i64 | Unique task identifier |
| `ExecutionId` | i64 | Unique execution run identifier |
| `AgentType` | Enum | planner, customer, developer, rag, vision, security, business |
| `TaskStatus` | Enum | pending, planning, executing, waiting, completed, failed |

---

## 4. Public API Trait

```rust
// nexus-agent/src/interfaces/api.rs
#[async_trait]
pub trait AgentService: Send + Sync {
    async fn start_execution(&self, req: StartAgentRequest) -> Result<ExecutionResponse, AgentError>;
    async fn get_execution_status(&self, id: ExecutionId) -> Result<ExecutionStatus, AgentError>;
    async fn stream_execution(&self, req: StreamRequest) -> Result<Receiver<ExecutionEvent>, AgentError>;
    async fn cancel_execution(&self, id: ExecutionId) -> Result<(), AgentError>;
}

pub struct StartAgentRequest {
    pub task: String,
    pub agent_type: AgentType,
    pub context: String,
    pub metadata: HashMap<String, String>,
    pub tenant_id: TenantId,
    pub user_id: UserId,
}

pub struct ExecutionEvent {
    pub event_type: EventType, // Thinking, ToolCall, ToolResult, Token, Completed
    pub content: String,
    pub step: u32,
    pub is_final: bool,
}
```

> **Note:** AgentService is consumed by `nexus-ai-gateway` and `nexus-gateway` via trait dispatch — no network overhead.

---

## 5. Agent Types

### 5.1 Planner Agent

| Attribute | Value |
|---|---|
| Model | `lfm2.5-thinking:1.2b` |
| Purpose | Understand user goals, break into steps, select agents |

**Responsibilities:**
- Intent detection from user input
- Task decomposition into ordered steps
- Agent selection for each step
- Dependency resolution between steps

**Example Output:**
```json
{
  "steps": [
    "Get customer details",
    "Check billing status",
    "Check network status",
    "Search knowledge base",
    "Generate solution"
  ],
  "agents": [
    "customer-agent",
    "billing-agent",
    "network-agent",
    "rag-agent",
    "response-agent"
  ]
}
```

### 5.2 Customer Support Agent

| Attribute | Value |
|---|---|
| Model | `phi4-mini:3.8b` |
| Purpose | Customer support, FAQ, ticket creation |

**Tools:**
- `customer.lookup(customer_id)` - Retrieve customer info
- `ticket.create(subject, body)` - Create support ticket
- `billing.check(customer_id)` - Check billing status
- `network.status(customer_id)` - Check network connectivity

### 5.3 Developer Agent

| Attribute | Value |
|---|---|
| Model | `qwen2.5-coder:3b` |
| Purpose | Code generation, review, debugging |

**Tools:**
- `git.search(query)` - Search codebase
- `code.analyze(file_path)` - Analyze code quality
- `test.execute(test_suite)` - Run test suite
- `security.scan(code)` - Security review

### 5.4 RAG Knowledge Agent

| Attribute | Value |
|---|---|
| Model | `command-r7b:7b` |
| Purpose | Enterprise knowledge reasoning |

**Knowledge Sources:**
- Documents (PDF, DOCX, HTML, Markdown)
- Policies and manuals
- Support tickets
- Database records

### 5.5 Vision Agent

| Attribute | Value |
|---|---|
| Model | `qwen3-vl:4b` |
| Purpose | Image understanding, OCR, visual analysis |

**Capabilities:**
- Image description and analysis
- Text extraction (OCR)
- Screenshot analysis
- Document extraction
- Device troubleshooting (ONU/router LED analysis)

### 5.6 Security Agent

| Attribute | Value |
|---|---|
| Model | `whiterabbitneo:7b` |
| Purpose | Security analysis, vulnerability detection |

**Capabilities:**
- Code security review
- Vulnerability analysis
- Threat assessment
- Secure coding recommendations

### 5.7 Business Intelligence Agent

| Attribute | Value |
|---|---|
| Model | `llama3.1:7b` |
| Purpose | Business analysis, reports, forecasting |

**Capabilities:**
- Revenue analysis
- Customer churn prediction
- Operational insights
- Strategic recommendations

---

## 6. Agent Lifecycle

```
User Request
    |
    v
[1] Intent Detection (Planner Agent)
    |
    v
[2] Task Planning (Step Decomposition)
    |
    v
[3] Agent Selection (Match capabilities)
    |
    v
[4] Context Assembly (Gather data)
    |
    v
[5] Tool Selection (Determine tools needed)
    |
    v
[6] Execution (Call tools, gather data)
    |
    v
[7] Knowledge Retrieval (RAG if needed)
    |
    v
[8] Reasoning (LLM reasoning over gathered data)
    |
    v
[9] Response Generation
    |
    v
[10] Memory Update (Store conversation context)
    |
    v
[11] NATS Event Published (AgentCompleted)
```

---

## 7. Tool Calling Architecture

Agents cannot directly access external systems. All tool calls go through a controlled pipeline:

```
Agent Tool Request
    |
    v
Tool Gateway
    |
    v
Permission Engine (RBAC + ABAC check)
    |
    v
Rate Limiter (Per-tool limits)
    |
    v
Tool Execution (trait method call to module)
    |
    v
Result returned to Agent
```

### Tool Definition Format

```json
{
  "name": "customer.lookup",
  "description": "Retrieve customer information by ID",
  "parameters": {
    "customer_id": {
      "type": "string",
      "required": true,
      "description": "Customer ID"
    }
  },
  "permissions": ["customer.read"],
  "category": "data_access"
}
```

### Tool Security Rules

| Action | Policy |
|---|---|
| Read customer data | Allowed with customer.read permission |
| Create ticket | Allowed with ticket.create permission |
| Delete customer data | BLOCKED - Requires human approval |
| Refund payment | BLOCKED - Requires human approval |
| Execute SQL | Read-only, validated, against read replica |

---

## 8. Multi-Agent Collaboration

### Sequential Execution

```
Step 1: Customer Agent -> Get customer data
Step 2: Network Agent -> Check network status
Step 3: RAG Agent -> Find troubleshooting guide
Step 4: Response Agent -> Generate final answer
```

### Parallel Execution

```
Step 1 (parallel):
    ├── Customer Agent -> Get customer data
    ├── Billing Agent -> Check billing status
    └── Network Agent -> Check network status

Step 2 (after all complete):
    └── RAG Agent -> Generate answer from combined context
```

---

## 9. Database Schema (agent_ schema)

### agents

```sql
-- Schema: agent_
CREATE TABLE agent.agents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    system_prompt TEXT,
    capabilities JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### agent_executions

```sql
CREATE TABLE agent.executions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    agent_id BIGINT NOT NULL REFERENCES agent.agents(id),
    task TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    plan JSONB,
    result JSONB,
    tokens_used INT,
    latency_ms FLOAT,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);
```

### agent_steps

```sql
CREATE TABLE agent.steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    execution_id BIGINT NOT NULL REFERENCES agent.executions(id),
    step_number INT NOT NULL,
    agent_type VARCHAR(100),
    action TEXT NOT NULL,
    tool_name VARCHAR(100),
    tool_params JSONB,
    result JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);
```

### agent_document_sets

```sql
CREATE TABLE agent.document_sets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES agent.agents(id) ON DELETE CASCADE,
    document_set_id BIGINT NOT NULL,  -- references rag.document_sets
    tenant_id BIGINT NOT NULL,
    permission_level VARCHAR(50) NOT NULL DEFAULT 'read',
    bound_by BIGINT NOT NULL,
    bound_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, document_set_id)
);

CREATE INDEX idx_agent_docsets_agent ON agent.document_sets(agent_id);
CREATE INDEX idx_agent_docsets_tenant ON agent.document_sets(tenant_id);
```

### agent_databases

```sql
CREATE TABLE agent.databases (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES agent.agents(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    connection_name VARCHAR(100) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL DEFAULT 5432,
    database_name VARCHAR(100) NOT NULL,
    password_encrypted TEXT NOT NULL,
    ssl_mode VARCHAR(20) NOT NULL DEFAULT 'require',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    UNIQUE(agent_id, database_name)
);

CREATE INDEX idx_agent_dbs_agent ON agent.databases(agent_id);
CREATE INDEX idx_agent_dbs_tenant ON agent.databases(tenant_id);
```

### agent_database_tables

```sql
CREATE TABLE agent.database_tables (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_database_id BIGINT NOT NULL REFERENCES agent.databases(id) ON DELETE CASCADE,
    agent_id BIGINT NOT NULL REFERENCES agent.agents(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    columns JSONB NOT NULL,
    primary_key JSONB,
    UNIQUE(agent_database_id, table_name)
);

CREATE INDEX idx_agent_dbtables_agent ON agent.database_tables(agent_id);
CREATE INDEX idx_agent_dbtables_db ON agent.database_tables(agent_database_id);
```

---

## 10. Agent-Document Set Binding

### Binding Rule

Every agent MUST be bound to at least one document set for RAG operations. An agent can ONLY access documents from its bound sets.

### Binding Flow

```
Admin selects agent → Views available document sets → Binds agent to sets → Agent scope defined
```

### Scope Enforcement

```
Agent Request → Query agent_document_sets for bound set IDs
  → Get document IDs from bound sets
    → RAG search filtered to those documents only
      → Results returned from scoped documents
```

### Binding API

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/agents/:id/document-sets` | Bind agent to sets |
| GET | `/api/v1/agents/:id/document-sets` | List bound sets |
| DELETE | `/api/v1/agents/:id/document-sets/:set_id` | Unbind agent |

---

## 11. Agent-Database Binding (SQL Agent)

### Binding Rule

SQL agents MUST be bound to specific databases and tables. The agent can ONLY query bound resources. Connection credentials must be tested before binding.

### Test Connection Flow

```
Admin enters credentials → Click "Test Connection"
  → TCP connection to host:port
    → Authentication with username/password
      → SSL/TLS handshake
        → Database accessible
          → IF FAILED: Show error
          → IF SUCCESS: Discover schema
```

### Database Binding API

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/agents/:id/sql-connections/test` | Test database connection |
| POST | `/api/v1/agents/:id/sql-connections/discover` | Discover schema |
| POST | `/api/v1/agents/:id/sql-connections/tables` | Bind tables |
| GET | `/api/v1/agents/:id/sql-connections` | List bound databases |
| DELETE | `/api/v1/agents/:id/sql-connections/:conn_id` | Unbind database |

### Scope Enforcement

```
Agent SQL Query → Query agent_databases for bound connections
  → Query agent_database_tables for bound tables
    → Validate query only references bound tables
      → Execute on bound connection only
        → Return scoped results
```

---

## 12. NATS Events

### Published

| Subject | Event |
|---|---|
| `aeroxe.v1.agent.started` | `AgentStarted` |
| `aeroxe.v1.agent.completed` | `AgentCompleted` |
| `aeroxe.v1.agent.failed` | `AgentFailed` |
| `aeroxe.v1.agent.tool.executed` | `ToolExecuted` |
| `aeroxe.v1.agent.bound` | `AgentBoundToDocumentSet` |
| `aeroxe.v1.agent.unbound` | `AgentUnboundFromDocumentSet` |
| `aeroxe.v1.agent.db.test.success` | `AgentDBConnectionTested` |
| `aeroxe.v1.agent.db.test.failed` | `AgentDBConnectionTestFailed` |
| `aeroxe.v1.agent.db.bound` | `AgentBoundToDatabase` |
| `aeroxe.v1.agent.db.unbound` | `AgentUnboundFromDatabase` |

### Subscribed

| Subject | Handler |
|---|---|
| `aeroxe.v1.ai.request.created` | Start agent execution |
| `aeroxe.v1.rag.completed` | Process RAG results |
| `aeroxe.v1.vision.completed` | Process vision results |

---

## 13. Human Approval Workflow

For sensitive actions (refunds, data deletion, financial transactions):

```
AI Agent recommends action
    |
    v
Policy Engine determines approval needed
    |
    v
Notification sent to supervisor
    |
    v
Human approves / rejects
    |
    v
Action executed (or rejected with reason)
```

---

## 13.1 Voice Call Agent Support (NEW)

The agent orchestrator supports voice call contexts alongside text chat:

### Voice Call Agent Flow

```
Inbound Call Received (via nexus-telephony)
    |
    v
Agent Orchestrator receives call context
    |  - caller_number
    |  - customer_id (if matched)
    |  - call_id
    |
    v
Select Agent
    |  - Match caller intent to agent
    |  - Load agent voice/personality config
    |
    v
Initialize Conversation
    |  - Create conversation (nexus-conversation)
    |  - Set channel: voice
    |
    v
Audio Processing Loop
    |  - Receive audio (nexus-telephony)
    |  - Transcribe (nexus-stt)
    |  - Process text (agent reasoning)
    |  - Generate response text
    |  - Synthesize speech (nexus-tts)
    |  - Send audio (nexus-telephony)
    |
    v
Handle Barge-in
    |  - Detect caller interrupt
    |  - Pause current TTS
    |  - Process new input
    |
    v
Transfer to Human (if needed)
    |  - Queue for human agent
    |  - Transfer call
    |  - Pass conversation context
    |
    v
Call Ended
    |  - Save transcript
    |  - Update analytics
    |  - Publish audit event
```

### Voice-Specific Agent Tools

| Tool | Description |
|---|---|
| `call.hold()` | Put caller on hold |
| `call.transfer(target)` | Transfer call to human/queue |
| `call.play_message(text)` | Play TTS message without waiting for response |
| `call.play_audio(file)` | Play pre-recorded audio |
| `call.dtmf.send(digits)` | Send DTMF tones |
| `call.recording.start()` | Start call recording |
| `call.recording.stop()` | Stop call recording |

### Agent Voice Configuration

```rust
pub struct AgentVoiceConfig {
    pub voice_id: VoiceId,                    // TTS voice
    pub speed: f32,                           // Speech rate (0.5-2.0)
    pub pitch: f32,                           // Voice pitch
    pub emotion_default: String,              // Default emotion style
    pub greeting_message: String,             // Opening script
    pub transfer_message: String,             // Pre-transfer message
    pub hold_music: Option<String>,           // Hold music file
    pub max_silence_seconds: u32,             // Max silence before prompt
    pub barge_in_enabled: bool,               // Allow caller interrupt
    pub speech_timeout_ms: u32,               // Silence before considering speech ended
}
```

---

## 14. Observability

### Tracked Metrics

| Metric | Description |
|---|---|
| `agent_executions_total` | Total executions by agent type |
| `agent_execution_duration_ms` | Execution time |
| `agent_tool_calls_total` | Tool invocations by tool name |
| `agent_tool_latency_ms` | Tool execution latency |
| `agent_tokens_used` | LLM tokens consumed |
| `agent_errors_total` | Failed executions by error type |

### Execution Trace

Every execution stores the complete decision path:
```
Intent -> Plan -> Step 1 (agent, tools, result) -> Step 2 -> ... -> Final Answer
```
