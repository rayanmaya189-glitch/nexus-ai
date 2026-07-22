# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 6 — DevOps, Deployment & Infrastructure Architecture

## Single Binary + Modular Monolith + GPU AI Runtime + CI/CD + Observability

---

# 1. Infrastructure Architecture Overview

AeroXe Nexus AI is designed for **single binary deployment** (modular monolith).

Deployment goals:

* **Single Rust binary** containing all modules
* Run offline after model download
* Support local GPU inference
* Maintain enterprise security
* Enable future cloud/hybrid deployment

---

# 2. Deployment Architecture

```text
                              Users


                                |

                                |

                         Load Balancer


                                |

                                |

                     gateway (axum HTTP/WS)

                     Single Binary: aeroxe-nexus


================================================================


                         Application Layer

                         (all in one process)


================================================================


 All Modules (src/modules/)

 identity | customer | agent | rag | vision
 sql-agent | memory | workflow | security
 audit | notification | model-registry
 config | ecosystem


================================================================


                         Infrastructure Layer


================================================================


 PostgreSQL Cluster 16 (Schema-per-Module)

 Redis Cluster

 NATS JetStream Cluster

 Elasticsearch Cluster

 MinIO Cluster


================================================================


                         AI Compute Layer


================================================================


 Ollama GPU Nodes


 LFM | Hermes3 | Qwen | Llama | Command-R | WhiteRabbitNeo


```

---

# 3. Single Binary Architecture

The entire application compiles to a single Rust binary.

Example:

```text
aeroxe-nexus (single binary)

├── All modules compiled in
├── Single axum HTTP/WS server (port 8080)
├── Shared PostgreSQL connection pool (SeaORM)
├── Shared Redis connection
├── Shared NATS connection
├── Shared Ollama HTTP client

```

Benefits:

| Aspect | Modular Monolith (1 binary) |
|---|---|
| Deployment | Single systemd service |
| Startup | Minutes | < 1 second |
| Debugging | Cross-container tracing | Single process |
| Latency | < 1ms (gRPC in-process) |
| Resource usage | N runtimes | 1 runtime |

├── NATS Client

├── Health Check

└── Metrics Endpoint

```

---

# 4. Docker Standards

Every service must provide:

```text
Dockerfile

docker-compose.yml

.env.example

healthcheck

README

migration scripts

```

---

# 5. Example Dockerfile

For Rust:

```dockerfile
FROM rust:1.80 AS builder


WORKDIR /app


COPY .

RUN cargo build --release



FROM debian:bookworm-slim


COPY --from=builder \
/app/target/release/service \
/app/service


CMD ["/app/service"]

```

---

# 6. Local Development Environment

Developer machine:

```text
Docker Compose


Services:

- PostgreSQL

- Redis

- NATS

- MinIO

- Ollama

- Elasticsearch

- aeroxe-nexus (single binary)

```

---

# 7. Development Docker Compose Architecture

```text
                 docker-compose


                       |


================================================


PostgreSQL

Redis

NATS JetStream

MinIO

Ollama

Qdrant/pgvector

ElasticSearch


================================================


Application (single binary)


aeroxe-nexus


src/modules/identity

src/modules/customer

src/modules/agent

src/modules/rag

src/modules/vision


```

---

# 8. Kubernetes Production Architecture

Production:

```text
Kubernetes Cluster


Namespaces:


aeroxe-system

aeroxe-ai

aeroxe-data

aeroxe-monitoring

```

---

# 9. Kubernetes Namespace Design

## aeroxe-ai

Contains:

```text
aeroxe-nexus (single binary)

src/modules/ai-gateway

src/modules/agent

src/modules/rag

src/modules/vision

src/modules/sql-agent

src/modules/workflow

```

---

## aeroxe-data

Contains:

```text
PostgreSQL

Redis

NATS

MinIO

ElasticSearch

```

---

## aeroxe-monitoring

Contains:

```text
Prometheus

Grafana

Loki

Tempo

OpenTelemetry

```

---

# 10. Kubernetes Deployment Example

Single Binary Deployment:

```yaml
apiVersion: apps/v1

kind: Deployment


metadata:

 name: aeroxe-nexus


spec:

 replicas: 3


 selector:

  matchLabels:

   app: aeroxe-nexus


 template:


  spec:


   containers:


   - name: nexus


     image: aeroxe/nexus:latest


     ports:


     - containerPort: 8080

```

---

# 11. AI GPU Infrastructure

AI inference runs separately.

Architecture:

```text
                AI Modules


                    |


                    | HTTP


                    |


             Ollama GPU Server


                    |


==================================


RTX 3060

RTX 4090

A6000

L40S

==================================


```

---

# 12. Ollama Deployment

Service:

```text
ollama-service
```

Responsibilities:

* Model management
* GPU inference
* Token streaming
* **Circuit Breaker**: 5 consecutive failures → Open state; 30s timeout → Half-Open (3 probe requests); 3 successful probes → Closed state

---

# 13. Model Deployment Strategy

Models:

| Model                | Purpose           |
| -------------------- | ----------------- |
| LFM2.5 Thinking 1.2B | Planning          |
| Hermes3 3B           | Agent Control     |
| Phi-4 Mini 3.8B      | General Assistant |
| Qwen2.5 Coder 3B     | Development       |
| Qwen3-VL 4B          | Vision            |
| Command-R 7B         | RAG               |
| Llama 3.1 7B         | Reasoning         |
| WhiteRabbitNeo 7B    | Security          |

---

# 14. GPU Scheduling

Kubernetes:

```yaml
resources:

 limits:

   nvidia.com/gpu: 1

```

---

# 15. Model Routing Architecture

The platform does not use one model for everything.

Flow:

```text
User Request


        |


Agent Router


        |


Intent Detection


        |


==============================


Coding

  |

Qwen Coder



Vision

  |

Qwen3-VL



Security

  |

WhiteRabbitNeo



RAG

  |

Command-R



Reasoning

  |

Llama



==============================

```

---

# 16. CI/CD Architecture

Pipeline:

```text
Developer


 |

Git Push


 |

GitLab


 |

CI Pipeline


 |

Build


 |

Test


 |

Security Scan


 |

Docker Image


 |

Registry


 |

Kubernetes Deploy


```

---

# 17. Git Repository Structure

```text
aeroxe-nexus-ai/


├── src/


│
├── modules/


│   ├── identity/

│   ├── customer/

│   ├── agent/

│   ├── rag/

│   ├── vision/


├── proto/


├── infrastructure/


├── kubernetes/


├── docker/


├── docs/


└── tests/

```

---

# 18. CI Pipeline Stages## Stage 1
Code Quality
Tools:
```text
Rust:
cargo fmt
cargo clippy
cargo audit
```

---

# Stage 2

Testing

```text
Unit Tests

Integration Tests

Contract Tests

```

---

# Stage 3

Security

Tools:

```text
Trivy

OWASP Dependency Check

SAST Scanner

```

---

# Stage 4

Build

Create:

```text
Docker Image

```

---

# Stage 5

Deploy

```text
Kubernetes Rolling Update

```

---

# 19. Observability Architecture

Stack:

```text
OpenTelemetry


        |


        |


-----------------------------

Metrics

Prometheus


Logs

Loki


Tracing

Tempo


Dashboard

Grafana


-----------------------------

```

---

# 20. Application Metrics

Every service exposes:

```text
/metrics
```

Metrics:

## API

```
request_count

latency

error_rate

```

---

## AI

```
tokens_generated

model_latency

prompt_size

completion_size

```

---

## Database

```
connection_pool

query_time

slow_queries

```

---

# 21. Logging Standard

Format:

JSON structured logs

Example:

```json
{
 "timestamp":"2026-07-15",

 "service":"agent-service",

 "level":"INFO",

 "trace_id":"123",

 "message":"Agent completed"

}

```

---

# 22. Health Check Standard

Every service:

## Liveness

```
POST /health/live

```

## Readiness

```
POST /health/ready

```

---

# 23. Scaling Architecture

## Horizontal Scaling

Example:

Single Binary:

```text
1 Replica


        |

High Traffic


        |

10 Replicas

```

---

# 24. AI Scaling Strategy

AI workloads are separated.

Example:

## Small Model Node

Hardware:

```
RTX 3060 12GB

```

Runs:

```
LFM

Hermes

Phi

Qwen Coder

```

---

## Large AI Node

Hardware:

```
RTX 4090

A6000

L40S

```

Runs:

```
Command-R

Llama

WhiteRabbitNeo

```

---

# 25. Backup Architecture

## Database

PostgreSQL:

```
Daily Full Backup

WAL Archive

Point-in-Time Recovery

```

---

## Object Storage

MinIO:

```
Versioning

Replication

Encryption

```

---

# 26. Disaster Recovery

Targets:

## RPO

```
15 minutes
```

## RTO

```
2 hours
```

---

# 27. Production Hardware Recommendation

## Development

Your RTX 3060 system:

```
CPU:
i5/i7


RAM:
32GB+


GPU:
RTX 3060 12GB


Storage:
1TB SSD


```

Suitable for:

* Development
* Testing
* Small users

---

## Production Small

```
CPU:
16-32 cores


RAM:
128GB


GPU:
RTX 4090 / A5000


Storage:
NVMe SSD

```

---

## Enterprise Production

```
CPU:
64+ cores


RAM:
256GB-512GB


GPU:
A6000/L40S


Storage:
NVMe RAID


```

---

# 28. Final DevOps Architecture

```text
                    AeroXe Nexus AI


                         |


                   Kubernetes


                         |


================================================


Applications


Single Binary (Modular Monolith)


gRPC / NATS (inter-service)


NATS JetStream


================================================


Data Layer


PostgreSQL

Redis

Elastic

MinIO

pgvector


================================================


AI Layer


Ollama

GPU Nodes

Model Router


================================================


Operations


CI/CD

Prometheus

Grafana

OpenTelemetry


================================================

```

---

# Part 6 Completed

Covered:

✅ Docker Architecture
✅ Kubernetes Design
✅ Ollama GPU Deployment
✅ Model Routing
✅ CI/CD Pipeline
✅ Git Strategy
✅ Monitoring
✅ Logging
✅ Scaling
✅ Backup
✅ Disaster Recovery
✅ Production Infrastructure

---
