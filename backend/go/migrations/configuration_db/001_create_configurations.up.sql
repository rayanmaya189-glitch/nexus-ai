-- Configuration Service - Platform Config, Tenant Config

CREATE TABLE IF NOT EXISTS platform_configurations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    description TEXT,
    data_type VARCHAR(20) DEFAULT 'string' CHECK (data_type IN ('string', 'int', 'float', 'bool', 'json')),
    is_sensitive BOOLEAN DEFAULT FALSE,
    default_value TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_platform_config_category ON platform_configurations(category);

CREATE TABLE IF NOT EXISTS tenant_configurations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    config_key VARCHAR(100) NOT NULL,
    config_value TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, config_key)
);

CREATE INDEX idx_tenant_config_tenant ON tenant_configurations(tenant_id);

-- Default configurations
INSERT INTO platform_configurations (config_key, config_value, category, description, data_type) VALUES
('default_model', 'lfm2.5-thinking:1.2b', 'ai', 'Default AI model for chat', 'string'),
('max_tokens', '4096', 'ai', 'Maximum tokens per request', 'int'),
('temperature', '0.7', 'ai', 'Default temperature', 'float'),
('enable_content_filter', 'true', 'security', 'Enable sensitive content filtering', 'bool'),
('enable_injection_detection', 'true', 'security', 'Enable prompt injection detection', 'bool'),
('audit_retention_days', '365', 'audit', 'Audit log retention in days', 'int'),
('rag_chunk_size', '512', 'rag', 'Default chunk size for document ingestion', 'int'),
('rag_chunk_overlap', '50', 'rag', 'Default chunk overlap', 'int');
