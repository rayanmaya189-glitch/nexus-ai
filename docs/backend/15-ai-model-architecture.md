# AeroXe Nexus AI — AI Model Architecture

## Ollama Runtime, Model Routing, Multi-Model Strategy & GPU Management

---

## 1. Architecture Overview

AeroXe Nexus AI uses **Ollama** as the local AI inference runtime. The platform does NOT use a single model for everything. Instead, it employs a multi-model strategy where specialized models handle specific domains.

```
User Request
    |
    v
Agent Router (LFM2.5 Thinking)
    |
    v
Intent Classification
    |
    ├── Coding     -> Qwen2.5-Coder:3B
    ├── Security   -> WhiteRabbitNeo:7B
    ├── Vision     -> Qwen3-VL:4B
    ├── RAG        -> Command-R:7B
    ├── Reasoning  -> Llama3.1:7B
    └── General    -> Phi-4-Mini:3.8B
```

---

## 2. Model Catalogue

| Model | Size | Purpose | GPU Requirement | Priority |
|---|---|---|---|---|
| LFM2.5 Thinking | 1.2B | Planning engine, intent detection | RTX 3060 (12GB) | MEDIUM |
| Hermes3 | 3B | Agent controller, tool calling, MCP | RTX 3060 (12GB) | MEDIUM |
| Phi-4 Mini | 3.8B | General assistant, chatbot, FAQ | RTX 3060 (12GB) | MEDIUM |
| Qwen2.5 Coder | 3B | Code generation, debugging, SQL | RTX 3060 (12GB) | HIGH |
| Qwen3-VL | 4B | Vision: image, OCR, screenshots | RTX 3060 (12GB) | HIGH |
| Command-R | 7B | Enterprise RAG, knowledge reasoning | RTX 4090 (24GB) | HIGH |
| Llama 3.1 | 7B | Advanced reasoning, architecture | RTX 4090 (24GB) | HIGH |
| WhiteRabbitNeo | 7B | Security analysis, vulnerability | RTX 4090 (24GB) | ON DEMAND |

---

## 3. Model Purpose Details

### 3.1 LFM2.5 Thinking 1.2B — Planning Engine

**Responsibilities:**
- Intent detection from user input
- Task planning and decomposition
- Agent selection for each sub-task

**Example:**
```
User: "Analyze customer complaint"

Planner output:
{
  "steps": [
    "Get customer details",
    "Check ticket history",
    "Check network status",
    "Search knowledge articles",
    "Generate solution"
  ]
}
```

### 3.2 Hermes3 3B — Agent Controller

**Responsibilities:**
- Tool calling and function execution
- MCP (Model Context Protocol) integration
- Workflow control and sequencing

### 3.3 Phi-4 Mini 3.8B — General Assistant

**Use Cases:**
- Customer chatbot
- Employee assistant
- FAQ responses
- Simple conversational queries

### 3.4 Qwen2.5 Coder 3B — Developer AI

**Functions:**
- Code generation (Rust, Go, TypeScript, Python)
- Code review and debugging
- SQL query generation
- API design suggestions

### 3.5 Qwen3-VL 4B — Vision Intelligence

**Capabilities:**
- Image understanding and description
- OCR (Optical Character Recognition)
- Screenshot analysis
- Diagram and chart analysis

**Use Cases:**
- ONU/router LED troubleshooting
- Invoice processing
- Document extraction
- UI/UX analysis

### 3.6 Command-R 7B — Enterprise Knowledge AI

**Responsibilities:**
- RAG-powered answers from enterprise knowledge
- Policy search and interpretation
- Documentation reasoning
- Multi-source information synthesis

### 3.7 Llama 3.1 7B — Advanced Reasoning

**Responsibilities:**
- Complex business analysis
- Architecture decisions
- Root cause analysis
- Strategic recommendations

### 3.8 WhiteRabbitNeo 7B — Security AI

**Responsibilities:**
- Security code review
- Vulnerability analysis
- Threat detection
- Secure coding recommendations

---

## 4. Model Router Architecture

### Routing Flow

```
Request
    |
    v
Intent Classifier (LFM2.5 Thinking)
    |
    v
Model Router
    |
    ├── Simple Query     -> Phi-4 Mini (3.8B)
    ├── Coding Request   -> Qwen Coder (3B)
    ├── Vision Request   -> Qwen3-VL (4B)
    ├── Security Review  -> WhiteRabbitNeo (7B)
    ├── Complex Reasoning -> Llama 3.1 (7B)
    └── RAG / Knowledge  -> Command-R (7B)
```

### Routing Decision Matrix

| Signal | Model |
|---|---|
| "How do I code..." | Qwen2.5-Coder |
| "Analyze this image" | Qwen3-VL |
| "Is this code secure?" | WhiteRabbitNeo |
| "What does the policy say?" | Command-R |
| "Why did revenue decrease?" | Llama 3.1 |
| "Hello, how are you?" | Phi-4 Mini |
| "Plan the approach" | LFM2.5 Thinking |

---

## 5. Ollama Deployment

### Service

| Attribute | Value |
|---|---|
| Service Name | `ollama-service` |
| Protocol | HTTP |
| Default Port | 11434 |
| Model Storage | `/usr/share/ollama/.ollama/models` |

### API Endpoints Used

| Endpoint | Purpose |
|---|---|
| `POST /api/generate` | Text generation |
| `POST /api/chat` | Chat completion |
| `POST /api/embeddings` | Generate embeddings |
| `GET /api/tags` | List available models |
| `POST /api/pull` | Download model |

---

## 6. GPU Scheduling

### Kubernetes Resource Requests

```yaml
resources:
  limits:
    nvidia.com/gpu: 1
  requests:
    memory: "8Gi"
    cpu: "4"
```

### GPU Node Allocation

| Node | GPU | Models |
|---|---|---|
| Small AI Node | RTX 3060 12GB | LFM2.5, Hermes3, Phi-4, Qwen Coder, Qwen3-VL |
| Large AI Node | RTX 4090 24GB / A6000 | Command-R, Llama 3.1, WhiteRabbitNeo |
| Enterprise | A6000 / L40S | All models, parallel inference |

---

## 7. Embedding Model

| Attribute | Value |
|---|---|
| Model | Via Ollama (embedding endpoint) |
| Dimension | 768 |
| Used By | RAG Service, Memory Service |
| Storage | pgvector |

---

## 8. Model Performance Targets

| Metric | Target |
|---|---|
| First Token Latency | < 2s |
| Tokens/Second | > 30 tok/s |
| Concurrent Requests | 10+ per GPU |
| Model Load Time | < 30s |
| Embedding Latency | < 500ms |

---

## 9. Model Registry Service

The `model-registry-service` manages Ollama models:

### REST API

```
GET /api/v1/models          - List available models
GET /api/v1/models/{name}   - Get model details
POST /api/v1/models/pull    - Download a model
DELETE /api/v1/models/{name} - Remove a model
GET /api/v1/models/usage    - Usage statistics
```

### Model Status Response

```json
{
  "name": "qwen3-vl:4b",
  "type": "vision",
  "status": "available",
  "size_bytes": 2500000000,
  "vram_required": "5.2GB",
  "requests_today": 500,
  "avg_latency_ms": 1200
}
```

---

## 10. Model Monitoring

| Metric | Description |
|---|---|
| `ollama_requests_total` | Total requests by model |
| `ollama_request_duration_ms` | Inference latency |
| `ollama_tokens_generated` | Tokens produced |
| `ollama_gpu_utilization` | GPU memory and compute usage |
| `ollama_model_load_time` | Cold start time |
| `ollama_concurrent_requests` | Active concurrent inferences |
