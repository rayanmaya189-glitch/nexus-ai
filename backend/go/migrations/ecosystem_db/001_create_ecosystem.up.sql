-- Ecosystem Service - Integrations, MCP Tools

CREATE TABLE IF NOT EXISTS integrations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('github', 'slack', 'postgresql', 'mysql', 'mongodb', 'jira', 'custom')),
    config JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'error')),
    credentials_ref VARCHAR(255),
    last_health_check TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_integrations_tenant ON integrations(tenant_id);
CREATE INDEX idx_integrations_type ON integrations(type);

CREATE TABLE IF NOT EXISTS mcp_tools (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    input_schema JSONB DEFAULT '{}',
    integration_id BIGINT REFERENCES integrations(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_mcp_tools_tenant ON mcp_tools(tenant_id);
CREATE INDEX idx_mcp_tools_integration ON mcp_tools(integration_id);

CREATE TABLE IF NOT EXISTS mcp_tool_invocations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tool_id BIGINT NOT NULL REFERENCES mcp_tools(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    input JSONB,
    output JSONB,
    status VARCHAR(20) DEFAULT 'completed',
    latency_ms FLOAT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_mcp_invocations_tool ON mcp_tool_invocations(tool_id);
CREATE INDEX idx_mcp_invocations_tenant ON mcp_tool_invocations(tenant_id);

-- Default integrations
INSERT INTO integrations (tenant_id, name, type, config, status) VALUES
(0, 'GitHub', 'github', '{"org": "", "repos": []}', 'active'),
(0, 'Slack', 'slack', '{"workspace": "", "channels": []}', 'active'),
(0, 'PostgreSQL', 'postgresql', '{"host": "localhost", "port": 5432, "database": "nexus_main"}', 'active');

-- Default MCP tools
INSERT INTO mcp_tools (tenant_id, name, description, input_schema) VALUES
(0, 'execute_sql', 'Execute SQL queries against connected databases', '{"type": "object", "properties": {"query": {"type": "string"}, "database": {"type": "string"}}}'),
(0, 'search_code', 'Search code repositories', '{"type": "object", "properties": {"query": {"type": "string"}, "repo": {"type": "string"}}}'),
(0, 'send_message', 'Send messages to Slack channels', '{"type": "object", "properties": {"channel": {"type": "string"}, "message": {"type": "string"}}}');
