# AeroXe Nexus AI — Workflow Module

## Business Automation, Approvals & Task Management

> **Modular Monolith Module:** This document describes the `nexus-workflow` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces (see [Communication Architecture](12-communication-architecture.md)).

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-workflow` |
| Crate | `nexus-workflow` (workspace member) |
| Bounded Context | Workflow |
| Domain Type | Supporting Domain |
| Language | Rust |
| Schema | `workflow_` (in shared PostgreSQL) |
| Dependencies | `nexus-agent` (AI task execution), `nexus-notification` (notifications) |

---

## 2. Purpose

The Workflow module provides business process automation capabilities within the `aeroxe-nexus` monolith:

- Define and execute automated workflows
- Manage approval chains
- Coordinate multi-step business processes
- Integrate with AI agents for intelligent automation
- Track workflow execution and history

---

## 3. Aggregate Design

### WorkflowInstance Aggregate

```
WorkflowInstance (Aggregate Root)
├── WorkflowDefinition
│   ├── Name
│   ├── Steps[]
│   └── Triggers
├── Steps[]
│   ├── StepId
│   ├── Type
│   ├── Status
│   ├── Assignee
│   └── Result
├── Approvals[]
│   ├── ApprovalId
│   ├── Approver
│   ├── Status
│   └── Comment
└── Actions[]
    ├── ActionType
    ├── Target
    └── Payload
```

### Entities

| Entity | Attributes |
|---|---|
| Workflow | WorkflowId, Name, TenantId, Definition (JSONB), Status |
| WorkflowInstance | InstanceId, WorkflowId, Status, StartedAt, CompletedAt |
| WorkflowStep | StepId, InstanceId, StepNumber, Type, Status, Assignee |
| Approval | ApprovalId, StepId, ApproverId, Status, Comment |

### Value Objects

| Value Object | Type |
|---|---|
| `WorkflowId` | i64 |
| `InstanceId` | i64 |
| `StepType` | Enum (ai_task, approval, notification, api_call, condition) |
| `StepStatus` | Enum (pending, running, waiting, completed, failed, skipped) |

---

## 4. Public API Trait

```rust
// nexus-workflow/src/interfaces/api.rs
#[async_trait]
pub trait WorkflowService: Send + Sync {
    async fn start_workflow(&self, req: StartWorkflowRequest) -> Result<WorkflowResponse, WorkflowError>;
    async fn get_status(&self, id: WorkflowId) -> Result<WorkflowStatus, WorkflowError>;
    async fn approve_step(&self, req: ApproveRequest) -> Result<(), WorkflowError>;
    async fn cancel_workflow(&self, id: WorkflowId) -> Result<(), WorkflowError>;
    async fn list_workflows(&self, req: ListRequest) -> Result<Vec<WorkflowSummary>, WorkflowError>;
}

pub struct StartWorkflowRequest {
    pub workflow_name: String,
    pub tenant_id: TenantId,
    pub user_id: UserId,
    pub context: HashMap<String, String>,
}

pub struct WorkflowResponse {
    pub workflow_id: WorkflowId,
    pub status: WorkflowStatus,
    pub message: String,
}

pub struct ApproveRequest {
    pub instance_id: WorkflowId,
    pub step_id: StepId,
    pub approver_id: UserId,
    pub approved: bool,
    pub comment: Option<String>,
}
```

> **Note:** `WorkflowService` is consumed by `nexus-gateway` (HTTP handlers) and interacts with `nexus-agent` for AI task steps via trait dispatch.

---

## 5. Workflow Types

### 5.1 Customer Support Flow

```
New Ticket
    |
    v
AI Analysis (Agent)
    |  - Classify issue
    |  - Check customer account
    |  - Check network status
    |
    v
Route to Department
    |
    ├── Technical -> Technical Agent
    ├── Billing -> Billing Agent
    └── General -> Support Agent
    |
    v
Generate Response (AI)
    |
    v
Human Approval (if needed)
    |
    v
Send Response
    |
    v
Update Ticket
```

### 5.2 New Customer Onboarding

```
Customer Registration
    |
    v
CRM Create Customer
    |
    v
Billing Setup
    |
    v
Network Provisioning
    |
    v
Welcome Message (AI)
    |
    v
Follow-up Schedule (AI)
```

### 5.3 Document Approval

```
Document Uploaded
    |
    v
AI Review (Content check)
    |
    v
Manager Approval
    |
    v
Legal Review (if contract)
    |
    v
Final Approval
    |
    v
Publish
```

---

## 6. Step Types

| Type | Description | Execution |
|---|---|---|
| `ai_task` | AI agent processes data | gRPC to agent-orchestrator |
| `approval` | Human approval required | Wait for human action |
| `notification` | Send notification | gRPC to notification-service |
| `api_call` | External API call | HTTP/gRPC to external service |
| `condition` | Branch based on result | Evaluate condition |
| `delay` | Wait for specified time | Timer-based |
| `parallel` | Execute multiple steps | Fan-out/fan-in |

---

## 7. Database Schema (workflow_db)

### workflows

```sql
CREATE TABLE workflow.definitions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    definition JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### workflow_instances

```sql
CREATE TABLE workflow.instances (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES workflow.definitions(id),
    tenant_id BIGINT NOT NULL,
    initiated_by BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    context JSONB,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    error_message TEXT
);
```

### workflow_steps

```sql
CREATE TABLE workflow.steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES workflow.instances(id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    assignee_id BIGINT,
    input JSONB,
    output JSONB,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

### workflow_approvals

```sql
CREATE TABLE workflow.approvals (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    step_id BIGINT NOT NULL REFERENCES workflow.steps(id) ON DELETE CASCADE,
    approver_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    comment TEXT,
    decided_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 8. REST API Endpoints

### Start Workflow

```
POST /api/v1/workflows/start
```

**Request:**
```json
{
  "workflow": "customer-support-flow",
  "context": {
    "ticket_id": "tkt_123",
    "customer_id": "cust_456"
  }
}
```

**Response:**
```json
{
  "workflow_id": "wfl_789",
  "status": "running"
}
```

### Get Workflow Status

```
GET /api/v1/workflows/{id}
```

### Approve Step

```
POST /api/v1/workflows/{id}/steps/{step_id}/approve
```

### List Active Workflows

```
GET /api/v1/workflows?status=running
```

---

## 9. NATS Events

| Subject | Event |
|---|---|
| `aeroxe.workflow.started` | Workflow execution begins |
| `aeroxe.workflow.step.completed` | Step finished |
| `aeroxe.workflow.completed` | All steps complete |
| `aeroxe.workflow.failed` | Workflow error |
| `aeroxe.workflow.approval.requested` | Approval needed |
| `aeroxe.workflow.approval.decided` | Approval given/denied |

---

## 10. Observability

| Metric | Description |
|---|---|
| `workflow_instances_total` | Total workflow instances |
| `workflow_step_duration_ms` | Step execution time |
| `workflow_approval_wait_ms` | Time waiting for approval |
| `workflow_failures_total` | Failed workflows by type |
| `workflow_active_count` | Currently running workflows |
