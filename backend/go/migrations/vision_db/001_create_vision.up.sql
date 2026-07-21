-- Vision Service - Image Analysis, OCR, Classification Results

CREATE TABLE IF NOT EXISTS vision_analyses (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    image_source VARCHAR(50) NOT NULL CHECK (image_source IN ('base64', 'url', 'file')),
    image_ref TEXT NOT NULL,
    analysis_type VARCHAR(50) NOT NULL CHECK (analysis_type IN ('analyze', 'ocr', 'classify', 'describe', 'multimodal')),
    model_used VARCHAR(100) NOT NULL,
    result JSONB NOT NULL DEFAULT '{}',
    confidence FLOAT,
    tokens_used INT DEFAULT 0,
    latency_ms FLOAT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'completed',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_vision_analyses_tenant ON vision_analyses(tenant_id);
CREATE INDEX idx_vision_analyses_user ON vision_analyses(user_id);
CREATE INDEX idx_vision_analyses_type ON vision_analyses(analysis_type);
CREATE INDEX idx_vision_analyses_created ON vision_analyses(created_at);

-- OCR results with structured text extraction
CREATE TABLE IF NOT EXISTS vision_ocr_results (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    analysis_id BIGINT NOT NULL REFERENCES vision_analyses(id) ON DELETE CASCADE,
    extracted_text TEXT NOT NULL,
    confidence FLOAT,
    bounding_boxes JSONB DEFAULT '[]',
    language VARCHAR(10),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_vision_ocr_analysis ON vision_ocr_results(analysis_id);

-- Classification results
CREATE TABLE IF NOT EXISTS vision_classifications (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    analysis_id BIGINT NOT NULL REFERENCES vision_analyses(id) ON DELETE CASCADE,
    category VARCHAR(100) NOT NULL,
    label VARCHAR(255) NOT NULL,
    confidence FLOAT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_vision_classifications_analysis ON vision_classifications(analysis_id);
