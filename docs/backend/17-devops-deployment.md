# AeroXe Nexus AI — DevOps & Deployment

## Docker, Kubernetes, CI/CD, GPU Infrastructure & Observability

---

## 1. Infrastructure Philosophy

**Private Infrastructure First:**
- Run offline after model download
- Support local GPU inference
- Scale AI workloads independently
- Maintain enterprise security
- Enable future cloud/hybrid deployment

---

## 2. Deployment Architecture

```
                    Users
                      |
                Load Balancer
                      |
               Nexus API Gateway
                      |
  ====================================================
                     Kubernetes Cluster
  ====================================================
  Identity | AI Gateway | Agent Orchestrator | RAG
  Vision | SQL Agent | Memory | Workflow | Audit
  ====================================================
                      |
  ====================================================
                  Infrastructure Layer
  ====================================================
  PostgreSQL | Redis | NATS JetStream | Elasticsearch | MinIO
  ====================================================
                      |
  ====================================================
                  AI Compute Layer
  ====================================================
                  Ollama GPU Nodes
  ====================================================
```

---

## 3. Container Architecture

Every microservice runs as an independent container:

```dockerfile
# Rust service example
FROM rust:1.80 AS builder
WORKDIR /app
COPY .
RUN cargo build --release

FROM debian:bookworm-slim
COPY --from=builder /app/target/release/service /app/service
CMD ["/app/service"]
```

### Required Files Per Service

```
service-name/
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── healthcheck.sh
├── README.md
└── migrations/
```

---

## 4. Local Development

### Docker Compose Stack

| Service | Port | Purpose |
|---|---|---|
| PostgreSQL | 5432 | Primary database |
| Redis | 6379 | Cache, short-term memory |
| NATS | 4222 | Event streaming |
| MinIO | 9000 | Object storage |
| Ollama | 11434 | AI inference |
| Elasticsearch | 9200 | Full-text search |

---

## 5. Kubernetes Production

### Namespaces

| Namespace | Contents |
|---|---|
| `aeroxe-system` | System services, ingress |
| `aeroxe-ai` | AI microservices |
| `aeroxe-data` | PostgreSQL, Redis, NATS, MinIO, ES |
| `aeroxe-monitoring` | Prometheus, Grafana, Loki, Tempo |
| `aeroxe-gpu` | Ollama GPU nodes |

### Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: agent-service
  template:
    spec:
      containers:
      - name: agent
        image: aeroxe/agent-service:v1
        ports:
        - containerPort: 50051
        resources:
          limits:
            memory: "2Gi"
            cpu: "2"
```

---

## 6. CI/CD Pipeline

```
Developer
    |
    v
Git Push
    |
    v
GitLab CI
    |
    v
[Stage 1] Code Quality
    |  cargo fmt / clippy (Rust)
    |
    v
[Stage 2] Testing
    |  Unit Tests
    |  Integration Tests
    |  Contract Tests
    |
    v
[Stage 3] Security Scan
    |  Trivy (container scan)
    |  OWASP Dependency Check
    |  SAST Scanner
    |
    v
[Stage 4] Build
    |  Docker Image
    |
    v
[Stage 5] Push to Registry
    |
    v
[Stage 6] Deploy to Kubernetes
    |  Rolling Update
    |
    v
Production
```

---

## 7. Git Repository Structure

```
aeroxe-nexus-ai/
├── services/
│   ├── identity-service/
│   ├── agent-service/
│   ├── rag-service/
│   ├── vision-service/
│   └── ...
├── proto/
├── infrastructure/
│   ├── kubernetes/
│   └── docker/
├── docs/
│   ├── srs/
│   └── backend/
└── tests/
```

---

## 8. Observability Stack

| Component | Technology | Purpose |
|---|---|---|
| Metrics | Prometheus | Collect and store metrics |
| Logs | Loki | Log aggregation |
| Tracing | Tempo | Distributed tracing |
| Dashboard | Grafana | Visualization |
| Instrumentation | OpenTelemetry | Auto-instrumentation |

### Application Metrics

Every service exposes `/metrics`:

| Category | Metrics |
|---|---|
| API | request_count, latency, error_rate |
| AI | tokens_generated, model_latency, prompt_size |
| Database | connection_pool, query_time, slow_queries |
| Infrastructure | cpu, memory, gpu_usage |

### Logging Standard

```json
{
  "timestamp": "2026-07-15",
  "service": "agent-service",
  "level": "INFO",
  "trace_id": "123",
  "request_id": "456",
  "tenant_id": "789",
  "message": "Agent completed execution"
}
```

### Health Check Standard

```
GET /health/live    -> 200 OK (process alive)
GET /health/ready   -> 200 OK (dependencies available)
```

---

## 9. Scaling Architecture

### Horizontal Scaling

```
Agent Service: 1 Replica
        |
   High Traffic
        |
Agent Service: 10 Replicas
```

### AI Scaling Strategy

| Node Type | Hardware | Models |
|---|---|---|
| Small AI Node | RTX 3060 12GB | LFM, Hermes, Phi, Qwen Coder, Qwen3-VL |
| Large AI Node | RTX 4090 24GB | Command-R, Llama, WhiteRabbitNeo |
| Enterprise | A6000 / L40S | All models, parallel inference |

---

## 10. Production Hardware

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
| Application Server | 32 cores, 128GB RAM, 2TB NVMe |
| Database Server | 32 cores, 128GB RAM, 4TB NVMe RAID |
| AI Server | 16 cores, 64GB RAM, RTX 4090 |

### Enterprise (100K-1M users)

| Component | Spec |
|---|---|
| K8s Control Plane | 3 nodes, 16 cores, 64GB each |
| Worker Nodes | 4+ nodes, 32-64 cores, 128-256GB |
| GPU Node 1 | 2x RTX 4090 (small models) |
| GPU Node 2 | A6000 / L40S (large models) |

---

## 11. Backup & Disaster Recovery

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

## 12. Production Alerting

| Alert | Threshold |
|---|---|
| CPU | > 85% |
| RAM | > 90% |
| Disk | > 85% |
| GPU Temperature | > 85C |
| Model Unavailable | Any |
| API Latency | > 5s |
| Failed Logins | > 10/min |
| Prompt Injection | Any |
