# AeroXe Nexus AI — Agent Orchestrator Service

## AI Agent Lifecycle, Planning, Tool Execution & Orchestration

---

## 1. Service Identity

| Attribute | Value |
|---|---|
| Service Name | `agent-orchestrator-service` |
| Bounded Context | Agent Orchestration |
| Domain Type | Core Domain |
| Language | Rust |
| Database | `agent_db` (PostgreSQL) |
| gRPC Port | 50052 |

---

## 2. Purpose

The Agent Orchestrator is the brain of the AI platform. It manages:

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

## 4. gRPC Contract

```protobuf
syntax = "proto3";
package aeroxe.agent;

service AgentService {
  rpc StartExecution(StartAgentRequest) returns (AgentExecutionResponse);
  rpc GetExecutionStatus(StatusRequest) returns (StatusResponse);
  rpc StreamExecution(StreamRequest) returns (stream ExecutionEvent);
  rpc CancelExecution(CancelRequest) returns (CancelResponse);
}

message StartAgentRequest {
  string task = 1;
  string agent_type = 2;
  string context = 3;
  map<string, string> metadata = 4;
  string tenant_id = 5;
  string user_id = 6;
}

message AgentExecutionResponse {
  string execution_id = 1;
  string status = 2;
  string agent_type = 3;
}

message ExecutionEvent {
  string event_type = 1; // "thinking", "tool_call", "tool_result", "token", "completed"
  string content = 2;
  string step = 3;
  bool is_final = 4;
}
```

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
Tool Execution (gRPC to target service)
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

## 9. Database Schema (agent_db)

### agents

```sql
CREATE TABLE agents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
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
CREATE TABLE agent_executions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    agent_id BIGINT NOT NULL REFERENCES agents(id),
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
CREATE TABLE agent_steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    execution_id BIGINT NOT NULL REFERENCES agent_executions(id),
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

---

## 10. NATS Events

### Published

| Subject | Event |
|---|---|
| `aeroxe.agent.started` | `AgentStarted` |
| `aeroxe.agent.completed` | `AgentCompleted` |
| `aeroxe.agent.failed` | `AgentFailed` |
| `aeroxe.agent.tool.executed` | `ToolExecuted` |

### Subscribed

| Subject | Handler |
|---|---|
| `aeroxe.ai.request.created` | Start agent execution |
| `aeroxe.rag.completed` | Process RAG results |
| `aeroxe.vision.completed` | Process vision results |

---

## 11. Human Approval Workflow

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

## 12. Observability

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
