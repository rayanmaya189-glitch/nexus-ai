-- AI Gateway - Sessions, Requests, Chat Messages

CREATE TABLE IF NOT EXISTS ai_sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    agent_id BIGINT,
    model VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ai_sessions_tenant ON ai_sessions(tenant_id);
CREATE INDEX idx_ai_sessions_user ON ai_sessions(user_id);

CREATE TABLE IF NOT EXISTS ai_requests (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES ai_sessions(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    agent_id BIGINT,
    prompt TEXT NOT NULL,
    response TEXT,
    model VARCHAR(100) NOT NULL,
    prompt_tokens INT DEFAULT 0,
    response_tokens INT DEFAULT 0,
    latency_ms FLOAT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ai_requests_session ON ai_requests(session_id);
CREATE INDEX idx_ai_requests_tenant ON ai_requests(tenant_id);

CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES ai_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    tokens INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chat_messages_session ON chat_messages(session_id);
