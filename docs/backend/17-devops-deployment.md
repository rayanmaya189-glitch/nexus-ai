# AeroXe Nexus AI — DevOps & Deployment

## Docker, Kubernetes, CI/CD, GPU Infrastructure (Modular Monolith Deployment)

---

## 1. Infrastructure Philosophy

**Private Infrastructure First:**
- Run offline after model download
- Support local GPU inference
- Scale the binary, not individual services
- Maintain enterprise security
- Enable future cloud/hybrid deployment

**Key Difference from Microservices:**
The entire backend is a **single Rust binary**. You deploy one container instead of 10+.

---

## 2. Deployment Architecture

```
                    Users
                      |
                Load Balancer
                      |
          +-----------------------+
          |   aeroxe-nexus:latest  |  ← Single binary
          |   (modular monolith)   |
          +-----------------------+
                      |
  ==========================================
              Infrastructure Layer
  ==========================================
  PostgreSQL | Redis | NATS | MinIO | ES
  ==========================================
                      |
  ==========================================
              AI Compute Layer
  ==========================================
              Ollama GPU Nodes
  ==========================================
```

---

## 3. Container Architecture

### Single Dockerfile

```dockerfile
# Multi-stage Rust build
FROM rust:1.84-slim-bookworm AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
COPY src/ src/
COPY nexus-gateway/ nexus-gateway/
COPY nexus-identity/ nexus-identity/
COPY nexus-agent/ nexus-agent/
COPY nexus-rag/ nexus-rag/
COPY nexus-vision/ nexus-vision/
COPY nexus-sql-agent/ nexus-sql-agent/
COPY nexus-memory/ nexus-memory/
COPY nexus-workflow/ nexus-workflow/
COPY nexus-security-ai/ nexus-security-ai/
COPY nexus-audit/ nexus-audit/
COPY nexus-notification/ nexus-notification/
COPY nexus-model-registry/ nexus-model-registry/
COPY nexus-config/ nexus-config/
COPY nexus-ecosystem/ nexus-ecosystem/
RUN cargo build --release --bin aeroxe-nexus

# Runtime image
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/aeroxe-nexus /usr/local/bin/
COPY migrations/ /app/migrations/
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=30s \
  CMD curl -f http://localhost:8080/health || exit 1
CMD ["aeroxe-nexus"]
```

### Image Size

| Component | Size |
|---|---|
| Builder image | ~4GB (build dependencies) |
| Runtime image | **~50MB** (static binary + Debian slim) |
| Startup time | < 1 second (no JVM, no interpreter) |

---

## 4. Local Development

### Docker Compose (Development Stack)

```yaml
version: "3.9"
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: aeroxe_nexus
      POSTGRES_USER: aeroxe
      POSTGRES_PASSWORD: aeroxe_dev
    ports: ["5432:5432"]
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  nats:
    image: nats:2.10-alpine
    ports: ["4222:4222"]

  minio:
    image: minio/minio
    ports: ["9000:9000", "9001:9001"]
    command: server /data --console-address ":9001"

  elasticsearch:
    image: elasticsearch:8.12
    environment:
      - discovery.type=single-node
    ports: ["9200:9200"]

  ollama:
    image: ollama/ollama:latest
    ports: ["11434:11434"]
    volumes:
      - ollama_models:/root/.ollama
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]

  app:
    build: .
    ports: ["8080:8080"]
    environment:
      DATABASE_URL: postgres://aeroxe:aeroxe_dev@postgres:5432/aeroxe_nexus
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
      OLLAMA_URL: http://ollama:11434
      JWT_SECRET: dev-secret-do-not-use-in-prod
    depends_on:
      - postgres
      - redis
      - nats
      - ollama
```

| Service | Port | Purpose |
|---|---|---|
| PostgreSQL | 5432 | Primary database |
| Redis | 6379 | Cache, sessions |
| NATS | 4222 | Event streaming |
| MinIO | 9000 | Object storage |
| Elasticsearch | 9200 | Full-text search |
| Ollama | 11434 | AI inference |
| **app** | **8080** | **Single monolith binary** |

---

## 5. Kubernetes Production

### Namespaces

| Namespace | Contents |
|---|---|
| `aeroxe-system` | nexus monolith, ingress |
| `aeroxe-data` | PostgreSQL, Redis, NATS, MinIO, ES |
| `aeroxe-monitoring` | Prometheus, Grafana, Loki, Tempo |
| `aeroxe-gpu` | Ollama GPU nodes |

### Deployment (Single Monolith)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aeroxe-nexus
  namespace: aeroxe-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aeroxe-nexus
  template:
    metadata:
      labels:
        app: aeroxe-nexus
    spec:
      containers:
      - name: nexus
        image: aeroxe/nexus:v1.0.0
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: nexus-db
              key: url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: nexus-jwt
              key: secret
        - name: REDIS_URL
          value: redis://redis.aeroxe-data:6379
        - name: NATS_URL
          value: nats://nats.aeroxe-data:4222
        - name: OLLAMA_URL
          value: http://ollama.aeroxe-gpu:11434
        - name: LOG_LEVEL
          value: info
        resources:
          requests:
            memory: "1Gi"
            cpu: "1"
          limits:
            memory: "4Gi"
            cpu: "4"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 15
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: aeroxe-nexus
  namespace: aeroxe-system
spec:
  selector:
    app: aeroxe-nexus
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

### Scaling

| Metric | Scale Up | Scale Down |
|---|---|---|
| CPU | >70% for 2 min | <30% for 5 min |
| Active requests | >1000 per instance | <100 for 5 min |
| WebSocket connections | >500 per instance | <50 for 5 min |

Because it's a monolith, scaling means **replicating the entire binary**. All modules scale together.

---

## 6. CI/CD Pipeline

```yaml
name: CI/CD
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions-rust-lang/setup-rust-toolchain@v1
      - name: Run all tests
        run: cargo nextest run --all-features
      - name: Clippy
        run: cargo clippy -- -D warnings
      - name: Format check
        run: cargo fmt --check

  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Trivy scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          severity: 'CRITICAL,HIGH'

  build-and-deploy:
    needs: [test, security-scan]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build Docker image
        run: docker build -t aeroxe/nexus:${{ github.sha }} .
      - name: Push to registry
        run: docker push aeroxe/nexus:${{ github.sha }}
      - name: Deploy to K8s
        run: kubectl set image deployment/aeroxe-nexus nexus=aeroxe/nexus:${{ github.sha }}
```

---

## 7. Git Repository Structure

```
aeroxe-nexus-ai/
├── Cargo.toml                      # Workspace root
├── Cargo.lock
├── src/
│   └── main.rs                     # Binary entry point (wires modules)
├── nexus-gateway/                  # API Gateway module crate
├── nexus-identity/                 # Identity module crate
├── nexus-ai-gateway/              # AI Gateway module crate
├── nexus-agent/                    # Agent Orchestrator module crate
├── nexus-rag/                      # RAG module crate
├── nexus-vision/                   # Vision module crate
├── nexus-sql-agent/                # SQL Agent module crate
├── nexus-memory/                   # Memory module crate
├── nexus-workflow/                 # Workflow module crate
├── nexus-security-ai/              # Security AI module crate
├── nexus-audit/                    # Audit module crate
├── nexus-notification/             # Notifications module crate
├── nexus-model-registry/           # Model registry module crate
├── nexus-config/                   # Config module crate
├── nexus-ecosystem/                # Ecosystem integration module crate
├── nexus-telephony/                # Voice/Telephony module crate        ← NEW
├── nexus-conversation/             # Conversation state machine crate    ← NEW
├── nexus-stt/                      # Speech-to-Text module crate         ← NEW
├── nexus-tts/                      # Text-to-Speech module crate         ← NEW
├── nexus-analytics/                # Analytics module crate              ← NEW
├── nexus-webhook/                  # Webhook module crate                ← NEW
├── nexus-outbound/                 # Outbound campaign module crate      ← NEW
├── nexus-infrastructure/           # Infrastructure patterns crate        ← NEW
│   ├── outbox/                     # Transactional Outbox
│   ├── locking/                    # Distributed Locking
│   ├── caching/                    # Distributed Caching
│   └── ledger/                     # Double Entry Ledger
├── proto/                          # Optional external gRPC protos
├── infrastructure/
│   ├── kubernetes/
│   └── docker/
├── migrations/
├── docs/
└── tests/
    ├── e2e/                        # End-to-end tests
    └── benchmarks/                 # Performance benchmarks
```

---

## 8. Observability Stack

| Component | Technology | Purpose |
|---|---|---|
| Metrics | Prometheus | Collect and store metrics |
| Logs | Loki | Log aggregation |
| Tracing | Tempo | Distributed tracing |
| Dashboard | Grafana | Visualization |
| Instrumentation | OpenTelemetry + `tracing` | Auto-instrumentation |

### Application Metrics (Single Binary)

| Category | Metrics |
|---|---|
| API | `requests_total`, `request_duration_ms`, `error_rate` |
| AI | `tokens_generated`, `model_latency`, `prompt_size` |
| Module | `module_call_count`, `module_latency_ms` (per module) |
| Database | `connection_pool_size`, `query_time`, `slow_queries` |
| Infrastructure | `cpu`, `memory`, `gpu_usage` |

### Logging Standard

```json
{
  "timestamp": "2026-07-21T12:00:00Z",
  "module": "nexus-agent",
  "level": "INFO",
  "trace_id": "abc123",
  "request_id": "def456",
  "tenant_id": 1,
  "message": "Agent execution completed"
}
```

---

## 9. Production Hardware

### Development

| Component | Spec |
|---|---|
| CPU | i5/i7 |
| RAM | 32GB+ |
| GPU | RTX 3060 12GB |
| Storage | 1TB SSD |

### Production Small (10K-50K users)

| Component | Spec |
|---|---|
| Application + DB Server | 32 cores, 128GB RAM, 4TB NVMe |
| AI Server | 16 cores, 64GB RAM, RTX 4090 |

### Enterprise (100K-1M users)

| Component | Spec |
|---|---|
| K8s Workers (3-5 nodes) | 32-64 cores, 128-256GB RAM each |
| GPU Node 1 | 2x RTX 4090 (small models) |
| GPU Node 2 | A6000 / L40S (large models) |

---

## 10. Backup & Disaster Recovery

| Component | Strategy |
|---|---|
| PostgreSQL | Daily full + WAL archiving |
| MinIO | Versioning + replication |
| Redis | RDB + AOF |
| NATS | Stream snapshots |

| Metric | Target |
|---|---|
| RPO | < 15 minutes |
| RTO | < 2 hours |

---

## 11. Production Alerting

| Alert | Threshold |
|---|---|
| CPU | > 85% |
| RAM | > 90% |
| Disk | > 85% |
| GPU Temperature | > 85C |
| Module Latency | > 5s (any module) |
| Model Unavailable | Any |
| Failed Logins | > 10/min |
| Prompt Injection | Any |
