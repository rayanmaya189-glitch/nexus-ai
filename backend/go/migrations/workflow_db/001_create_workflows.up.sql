-- Workflow Service - Workflows, Executions, Steps

CREATE TABLE IF NOT EXISTS workflows (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    trigger_type VARCHAR(50) DEFAULT 'manual' CHECK (trigger_type IN ('manual', 'event', 'schedule', 'webhook')),
    config JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    version INT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_workflows_tenant ON workflows(tenant_id);

CREATE TABLE IF NOT EXISTS workflow_steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    agent_type VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    config JSONB DEFAULT '{}',
    timeout INT DEFAULT 300,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workflow_id, step_number)
);

CREATE INDEX idx_workflow_steps_workflow ON workflow_steps(workflow_id);

CREATE TABLE IF NOT EXISTS workflow_executions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    input TEXT,
    output TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    error_message TEXT,
    latency_ms FLOAT DEFAULT 0,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_workflow_executions_workflow ON workflow_executions(workflow_id);
CREATE INDEX idx_workflow_executions_tenant ON workflow_executions(tenant_id);
CREATE INDEX idx_workflow_executions_status ON workflow_executions(status);

CREATE TABLE IF NOT EXISTS workflow_execution_steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    execution_id BIGINT NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    step_id BIGINT NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    input TEXT,
    output TEXT,
    error_message TEXT,
    agent_execution_id BIGINT,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_workflow_execution_steps_execution ON workflow_execution_steps(execution_id);

-- Default workflows
INSERT INTO workflows (tenant_id, name, description, trigger_type, config, status) VALUES
(0, 'Customer Support Pipeline', 'Automated customer support with escalation', 'event', '{"escalation_threshold": 3, "max_retries": 2}', 'active'),
(0, 'Document Processing Pipeline', 'Ingest, chunk, embed, and index documents', 'manual', '{"chunk_size": 512, "chunk_overlap": 50}', 'active'),
(0, 'Security Review Pipeline', 'Automated security scanning and reporting', 'schedule', '{"schedule": "0 2 * * *", "models": ["whiterabbitneo:7b"]}', 'active');
