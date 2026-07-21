-- Memory Service - Short-term, Long-term, Episodic, Semantic Memory

CREATE TABLE IF NOT EXISTS memory_entries (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    agent_id BIGINT,
    session_id VARCHAR(100),
    memory_type VARCHAR(20) NOT NULL CHECK (memory_type IN ('short_term', 'long_term', 'episodic', 'semantic')),
    content TEXT NOT NULL,
    summary TEXT,
    importance FLOAT DEFAULT 0.5 CHECK (importance >= 0 AND importance <= 1),
    access_count INT DEFAULT 0,
    embedding vector(384),
    metadata JSONB DEFAULT '{}',
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memory_entries_tenant ON memory_entries(tenant_id);
CREATE INDEX idx_memory_entries_user ON memory_entries(user_id);
CREATE INDEX idx_memory_entries_agent ON memory_entries(agent_id);
CREATE INDEX idx_memory_entries_type ON memory_entries(memory_type);
CREATE INDEX idx_memory_entries_session ON memory_entries(session_id);
CREATE INDEX idx_memory_entries_embedding ON memory_entries USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Memory consolidation tracking
CREATE TABLE IF NOT EXISTS memory_consolidations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    agent_id BIGINT,
    source_memory_ids BIGINT[] NOT NULL,
    consolidated_memory_id BIGINT REFERENCES memory_entries(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_memory_consolidations_tenant ON memory_consolidations(tenant_id);
