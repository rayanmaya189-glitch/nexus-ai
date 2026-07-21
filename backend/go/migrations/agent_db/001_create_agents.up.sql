-- Agent Orchestrator - Agents, Executions, Steps

CREATE TABLE IF NOT EXISTS agents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    agent_type VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    system_prompt TEXT,
    capabilities JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_agents_tenant ON agents(tenant_id);
CREATE INDEX idx_agents_type ON agents(agent_type);

CREATE TABLE IF NOT EXISTS agent_executions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_id BIGINT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    task TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    result TEXT,
    tokens_used INT DEFAULT 0,
    latency_ms FLOAT DEFAULT 0,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_agent_executions_agent ON agent_executions(agent_id);
CREATE INDEX idx_agent_executions_tenant ON agent_executions(tenant_id);

CREATE TABLE IF NOT EXISTS agent_steps (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    execution_id BIGINT NOT NULL REFERENCES agent_executions(id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    agent_type VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    tool_name VARCHAR(100),
    tool_params JSONB,
    result TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_agent_steps_execution ON agent_steps(execution_id);

INSERT INTO agents (tenant_id, name, agent_type, model, system_prompt, capabilities) VALUES
(0, 'Planner Agent', 'planner', 'lfm2.5-thinking:1.2b', 'You are a planning agent. Break down complex tasks into smaller steps.', '{"planning": true, "task_decomposition": true}'),
(0, 'Customer Agent', 'customer', 'command-r7b:7b', 'You are a helpful customer support agent.', '{"customer_support": true, "multi_turn": true}'),
(0, 'Developer Agent', 'developer', 'qwen2.5-coder:3b', 'You are a software development assistant.', '{"code_review": true, "code_generation": true}'),
(0, 'Vision Agent', 'vision', 'qwen3-vl:4b', 'You are a vision analysis agent.', '{"image_analysis": true, "ocr": true}'),
(0, 'Security Agent', 'security', 'whiterabbitneo:7b', 'You are a security analysis agent.', '{"security_analysis": true, "vulnerability_detection": true}'),
(0, 'Business Agent', 'business', 'llama3.1:7b', 'You are a business intelligence agent.', '{"data_analysis": true, "reporting": true}'),
(0, 'RAG Agent', 'rag', 'phi4-mini:3.8b', 'You are a document Q&A agent.', '{"question_answering": true, "document_retrieval": true}'),
(0, 'SQL Agent', 'sql', 'qwen2.5-coder:3b', 'You are a SQL intelligence agent.', '{"sql_generation": true, "data_analysis": true}');
