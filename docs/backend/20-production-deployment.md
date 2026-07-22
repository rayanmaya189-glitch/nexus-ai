# 20 — Production Deployment & Infrastructure

## Hardware Sizing + Network Design + High Availability + Disaster Recovery

> **Modular Monolith Context:** This document covers deployment of the single `aeroxe-nexus` binary. For detailed container/K8s configuration, see [DevOps Deployment](17-devops-deployment.md). This document focuses on hardware sizing, zones, and operational considerations.

---

## 1. Production Deployment Goals

AeroXe Nexus AI production platform must support:

- Enterprise AI assistant
- Multiple AeroXe products integration
- Multi-tenant SaaS
- Private/on-premise deployment
- Offline AI operation after model download
- GPU-based local inference
- High availability
- Secure AI operations
- Future cloud/hybrid expansion

---

## 2. Production Architecture Overview

```
                           INTERNET
                               |
                         Firewall / WAF
                               |
                         Load Balancer
                               |
================================================================
                      Kubernetes Cluster
================================================================
+---------------------------+
|   aeroxe-nexus (Monolith)  |  ← Single binary, all modules
|   Port 8080 (HTTP/WS)      |
+---------------------------+
================================================================
                      Data Layer
================================================================
  PostgreSQL Cluster      Redis Cluster       NATS JetStream Cluster
  Elasticsearch Cluster   MinIO Storage
================================================================
                      AI Compute Layer
================================================================
  Ollama GPU Servers
  LFM   Hermes3   Phi4   Qwen Coder   Qwen3-VL   Command-R   Llama   WhiteRabbitNeo
================================================================
                      Monitoring Layer
================================================================
  Prometheus   Grafana   Loki   Tempo   OpenTelemetry
================================================================
```

> **Key Difference:** A single `aeroxe-nexus` deployment replaces multiple separate service deployments.

---

## 3. Production Infrastructure Zones

AeroXe Nexus AI uses security zones.

### Zone 1 — Public Zone

Contains:

| Component      |
| -------------- |
| Internet       |
| Users          |
| Mobile Apps    |
| Web Apps       |
| Partners       |

Protection:

- WAF
- DDoS Protection
- Rate Limiting

### Zone 2 — API Zone

Contains:

| Component          |
| ------------------ |
| Load Balancer      |
| API Gateway        |
| WebSocket Gateway  |

Security:

- TLS 1.3
- JWT validation
- Request filtering

### Zone 3 — Application Zone

Contains:

| Component            |
| -------------------- |
| Microservices        |
| AI Agents            |
| Workflow Engine      |
| Integration Services |

Communication:

- Trait-based dispatch (in-process)
- NATS

### Zone 4 — Data Zone

Contains:

| Component      |
| -------------- |
| PostgreSQL     |
| Redis          |
| Elastic        |
| MinIO          |
| NATS           |

Access: Only internal services.

### Zone 5 — AI Compute Zone

Contains:

| Component      |
| -------------- |
| Ollama Servers |
| GPU Nodes      |
| Model Storage  |

---

## 4. Small Production Hardware (10K–50K Users)

### Application Server

| Spec    | Value         |
| ------- | ------------- |
| CPU     | 32 cores      |
| RAM     | 128GB         |
| Storage | 2TB NVMe      |
| OS      | Ubuntu Server 24.04 |

### Database Server

| Spec    | Value              |
| ------- | ------------------ |
| CPU     | 32 cores           |
| RAM     | 128GB              |
| Storage | 4TB NVMe RAID      |
| Services| PostgreSQL, Redis  |

### AI Server

| Spec    | Value                  |
| ------- | ---------------------- |
| CPU     | 16 cores               |
| RAM     | 64GB                   |
| GPU     | RTX 4090 24GB or RTX A5000 24GB |

---

## 5. Enterprise Production Hardware (100K–1M Users)

### Kubernetes Control Plane

3 Nodes — each:

| Spec    | Value    |
| ------- | -------- |
| CPU     | 16 cores |
| RAM     | 64GB     |
| Storage | 1TB NVMe |

### Worker Nodes

Minimum 4 nodes:

| Spec    | Value          |
| ------- | -------------- |
| CPU     | 32–64 cores    |
| RAM     | 128–256GB      |

### AI GPU Nodes

| Node  | GPU                      | Purpose                           |
| ----- | ------------------------ | --------------------------------- |
| Node 1| RTX 4090 x2              | Small models, Agents              |
| Node 2| RTX A6000 / L40S         | Large reasoning, Vision, Heavy RAG|

---

## 6. GPU Model Allocation Strategy

### RTX 3060 12GB

| Model              | Parameters |
| ------------------ | ---------- |
| LFM2.5 Thinking    | 1.2B       |
| Hermes3            | 3B         |
| Phi-4 Mini         | 3.8B       |
| Qwen Coder         | 3B         |
| Qwen3-VL           | 4B         |

### RTX 4090 24GB

| Model              | Parameters |
| ------------------ | ---------- |
| Command-R          | 7B         |
| Llama 3.1          | 7B         |
| WhiteRabbitNeo     | 7B         |

---

## 7. Final Ollama Model Deployment

| Model                | Purpose       | Priority  |
| -------------------- | ------------- | --------- |
| Command-R 7B         | RAG           | HIGH      |
| Qwen3-VL 4B          | Vision        | HIGH      |
| Qwen Coder 3B        | Development   | HIGH      |
| Llama 3.1 7B         | Reasoning     | HIGH      |
| Hermes3 3B           | Agent Control | MEDIUM    |
| LFM2.5 Thinking 1.2B | Planning      | MEDIUM    |
| Phi-4 Mini 3.8B      | General Chat  | MEDIUM    |
| WhiteRabbitNeo 7B    | Security      | ON DEMAND |
| **Whisper Small**    | **STT**       | **HIGH**  |
| **Piper TTS**        | **TTS**       | **HIGH**  |

### Voice/Telephony Hardware (NEW)

| Component | Spec | Purpose |
|---|---|---|
| STT Server | 8 cores, 32GB RAM, GPU optional | Real-time transcription |
| TTS Server | 4 cores, 16GB RAM | Voice synthesis |
| SIP Server | 4 cores, 8GB RAM | Call routing (FreeSWITCH) |
| Audio Processing | 4 cores, 8GB RAM | Noise reduction, VAD |

---

## 8. Network Architecture

```
                Internet
                   |
               Firewall
                   |
              DMZ Network
                   |
          API Gateway Network
                   |
        Application Network
                   |
          Database Network
                   |
             AI Network
```

---

## 9. Internal Communication

Modules within the monolith communicate via **Rust trait method calls** (in-process). No gRPC or mTLS needed internally.

```
nexus-agent → RagService::search() trait method call → same process, no network
```

External gRPC is available only for optional partner SDK integrations via `nexus-gateway`.

---

## 10. Event Architecture (NATS JetStream)

```
customer.created
  → CRM Service
    → AI Sales Agent
      → Notification Service
```

---

## 11. Storage Architecture

### PostgreSQL

Stores: Users, Tenants, Agents, Workflows, Permissions, Transactions, Configurations

### pgvector

Stores: Document Embeddings, Memory Embeddings, Semantic Knowledge

### Elasticsearch

Stores: Logs, Audit, Search Index, Analytics

### MinIO

Stores: Documents, Images, AI Files, Backups, Models

---

## 12. High Availability Design

### Application Layer (Modular Monolith)

Multiple replicas of the single binary:

```
aeroxe-nexus → Replica 1, Replica 2, Replica 3 (all modules scale together)
```

### PostgreSQL HA

```
Primary → Replica 1 → Replica 2
Technology: Patroni + etcd
```

### Redis HA

```
Redis Cluster: Master → Replica → Shard
```

### NATS HA

```
NATS Node 1 + NATS Node 2 + NATS Node 3
JetStream Replication Factor: 3
```

---

## 13. Backup Architecture

### Database Backup

| Frequency  | Type          |
| ---------- | ------------- |
| Daily      | Full Backup   |
| Continuous | WAL Archive   |

### Object Backup (MinIO)

- Versioning
- Replication
- Lifecycle Management

---

## 14. Disaster Recovery

| Metric | Target     |
| ------ | ---------- |
| RPO    | <15 minutes|
| RTO    | <2 hours   |

---

## 15. Monitoring Architecture

### Stack

```
Application → OpenTelemetry →
  Metrics: Prometheus
  Logs: Loki
  Tracing: Tempo
  Dashboard: Grafana
```

### Production Alerting

#### Infrastructure Alerts

| Alert                |
| -------------------- |
| CPU >85%             |
| RAM >90%             |
| Disk >85%            |
| GPU Temperature      |
| Network Failure      |

#### AI Alerts

| Alert                |
| -------------------- |
| Model unavailable    |
| High latency         |
| Token failure        |
| GPU overload         |

#### Security Alerts

| Alert                 |
| --------------------- |
| Multiple failed login |
| Prompt injection      |
| Unauthorized access   |
| Tenant violation      |

---

## 16. Deployment Pipeline

```
Developer → Git Push → GitHub Actions → Testing (cargo test, clippy) → Security Scan (cargo-audit, trivy)
  → Docker Build → Container Registry → Kubernetes Rolling Update → Production
```

---

## 17. Production Environment Separation

```
Development → Testing → Staging → Production
```

---

## 18. Security Final Checklist

### Application

| Check              |
| ------------------ |
| JWT Authentication |
| RBAC               |
| ABAC               |
| API Security       |
| Input Validation   |

### AI

| Check                        |
| ---------------------------- |
| Prompt Injection Protection  |
| Tool Permission Control      |
| Model Isolation              |
| Data Leakage Prevention      |

### Infrastructure

| Check              |
| ------------------ |
| TLS Everywhere     |
| Firewall           |
| Secrets Management |
| Backup Encryption  |

---

## 19. Final Production Architecture

```
                        AeroXe Nexus AI
================================================================
                           Clients
================================================================
  Web   Mobile   Enterprise Apps   APIs
================================================================
                        |
================================================================
+------------------------------+  ← Single Deployment
|     aeroxe-nexus (Monolith)   |
|  http://:8080  /health       |
|  +-----+  +------+  +-----+ |
|  |gate |  |ai-gw |  |agent| |
|  |way  |  |ateway|  |     | |
|  +-----+  +------+  +-----+ |
|  +-----+  +------+  +-----+ |
|  |rag  |  |vision|  |sql  | |
|  +-----+  +------+  +-----+ |
|  +-----+  +------+  +-----+ |
|  |mem  |  |work  |  |sec  | |
|  +-----+  +------+  +-----+ |
+------------------------------+
================================================================
                      Infrastructure
================================================================
  PostgreSQL  Elasticsearch  Redis  NATS  MinIO
================================================================
                             |
================================================================
                      AI Runtime (Ollama)
================================================================
  LFM   Hermes3   Phi4   Qwen   Qwen3-VL   Command-R   Llama   WRNeo
================================================================
================================================================
                       Observability
================================================================
  Prometheus  Grafana  Loki  Tempo  OpenTelemetry
================================================================
```
