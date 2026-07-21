# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 20 — Infrastructure Patterns Architecture

## Outbox Pattern, Distributed Locking, Distributed Caching, Circuit Breaker

---

## 1. Transactional Outbox Pattern

### 1.1 Problem

Events lost when NATS publish fails after PostgreSQL commit.

### 1.2 Solution

Events stored in `outbox.events` table within the SAME transaction as business data. Background poller reads outbox and publishes to NATS.

### 1.3 Outbox Table

```sql
CREATE TABLE outbox.events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    nats_subject VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 5,
    next_retry_at TIMESTAMP,
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 1.4 Event Lifecycle

```
Created (pending) → Published → Archived
                 ↘ Failed → Retry → Published
                           ↘ Max Retries → Dead Letter
```

---

## 2. Distributed Locking

### 2.1 Problem

Multiple instances processing same resource simultaneously.

### 2.2 Solution

Redis-based distributed locks with TTL and owner verification.

### 2.3 Lock Table

```sql
CREATE TABLE distributed_locks.locks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    lock_key VARCHAR(200) NOT NULL UNIQUE,
    lock_owner VARCHAR(100) NOT NULL,
    acquired_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    ttl_seconds INT NOT NULL DEFAULT 30,
    metadata JSONB
);
```

### 2.4 Lock Usage

| Resource | Lock Key | TTL |
|---|---|---|
| Callback scheduling | `callback:{id}` | 30s |
| Campaign execution | `campaign:{id}` | 300s |
| Webhook delivery | `webhook:{delivery_id}` | 10s |
| Agent execution | `agent:{execution_id}` | 60s |
| Ledger transaction | `ledger:{tenant_id}` | 30s |

---

## 3. Distributed Caching

### 3.1 Multi-Tier Architecture

```
L1: In-Process Cache (per instance, fast, small)
    ↓ Miss
L2: Redis Cache (shared, fast, large)
    ↓ Miss
L3: PostgreSQL (source of truth)
```

### 3.2 Cache Strategies

| Strategy | Use Case | Consistency |
|---|---|---|
| Cache-Aside | Read-heavy data (config, models) | Eventual |
| Write-Through | Critical data (customer, agent) | Strong |
| Write-Behind | Analytics, metrics | Eventual |

### 3.3 Cache Invalidation

```
Instance A updates data → Write to PostgreSQL → Write to Redis
  → Publish invalidation event to NATS → Other instances invalidate L1
```

---

## 4. Circuit Breaker Pattern

### 4.1 Problem

External dependencies (Ollama, telephony providers) can fail and cause cascading failures.

### 4.2 Solution

Circuit breaker wraps external calls with failure detection and fallback.

### 4.3 States

```
Closed (normal) → Open (failing) → Half-Open (testing) → Closed
```

### 4.4 Configuration

```rust
pub struct CircuitBreakerConfig {
    pub failure_threshold: u32,      // Failures before opening
    pub success_threshold: u32,      // Successes before closing
    pub timeout_ms: u32,             // Time before half-open
    pub half_open_max_calls: u32,    // Calls in half-open state
}
```

### 4.5 Usage

| Dependency | Circuit Breaker | Fallback |
|---|---|---|
| Ollama | YES | Return cached response / error |
| Telephony Provider | YES | Try secondary provider |
| External API | YES | Return cached data |
| NATS | NO (JetStream handles) | — |

---

## 5. Retry with Exponential Backoff

### 5.1 Configuration

```rust
pub struct RetryConfig {
    pub max_retries: u32,           // Default: 3
    pub initial_delay_ms: u32,      // Default: 1000
    pub backoff_multiplier: f32,    // Default: 2.0
    pub max_delay_ms: u32,          // Default: 30000
    pub jitter: bool,               // Default: true
}
```

### 5.2 Retry Schedule

| Attempt | Delay (with jitter) |
|---|---|
| 1 | ~1000ms |
| 2 | ~2000ms |
| 3 | ~4000ms |
| 4 | ~8000ms |
| 5 | ~16000ms |

---

## 6. Bulkhead Pattern

### 6.1 Purpose

Isolate failures between modules to prevent cascading.

### 6.2 Implementation

```rust
pub struct BulkheadConfig {
    pub max_concurrent: u32,        // Max concurrent calls
    pub max_queue: u32,             // Max waiting calls
    pub timeout_ms: u32,            // Call timeout
}
```

### 6.3 Bulkhead Allocation

| Bulkhead | Max Concurrent | Purpose |
|---|---|---|
| AI Inference | 50 | Ollama calls |
| Database | 100 | PostgreSQL connections |
| Telephony | 200 | Concurrent calls |
| External API | 20 | Third-party API calls |

---

## 7. Database Schema Summary

| Schema | Purpose |
|---|---|
| `outbox_` | Transactional outbox for reliable event delivery |
| `distributed_locks_` | Distributed lock management |
| `distributed_cache_` | Cache metadata tracking |
| `circuit_breaker_` | Circuit breaker state tracking |

---

# End of Part 20
