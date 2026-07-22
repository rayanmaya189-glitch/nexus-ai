# AeroXe Nexus AI — Infrastructure Patterns

## Double Entry Ledger, Distributed Caching, Distributed Locking, Outbox Pattern

> This document defines critical infrastructure patterns for production reliability, data integrity, and performance in the `aeroxe-nexus` modular monolith.

---

## 1. Outbox Pattern (Transactional Outbox)

### 1.1 Problem

Current design: Module writes to PostgreSQL AND publishes to NATS separately. If NATS publish fails after PostgreSQL commit, events are lost.

```
Module → PostgreSQL (commit) → NATS (publish) → FAIL = EVENT LOST
```

### 1.2 Solution: Transactional Outbox

Events are stored in PostgreSQL (outbox table) within the SAME transaction as business data. A background poller reads the outbox and publishes to NATS.

```
Module → PostgreSQL (business data + outbox entry) → COMMIT
                                                          ↓
                                               Background Poller
                                                          ↓
                                               NATS (publish) → Success → Mark as published
```

### 1.3 Outbox Table Schema

```sql
CREATE TABLE outbox_.events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    aggregate_type VARCHAR(100) NOT NULL,    -- e.g., 'customer', 'call', 'conversation'
    aggregate_id VARCHAR(100) NOT NULL,      -- e.g., customer_id, call_id
    event_type VARCHAR(100) NOT NULL,        -- e.g., 'CustomerCreated', 'CallEnded'
    payload JSONB NOT NULL,                  -- Event data
    nats_subject VARCHAR(200) NOT NULL,      -- e.g., 'aeroxe.v1.customer.customer.created'
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending | published | failed
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 5,
    next_retry_at TIMESTAMP,
    published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_pending ON outbox_.events(status, created_at)
    WHERE status = 'pending';
CREATE INDEX idx_outbox_retry ON outbox_.events(status, next_retry_at)
    WHERE status = 'failed';
```

### 1.4 Outbox Implementation

```rust
// Within a module's transactional context
pub async fn create_customer(
    db: &DatabaseConnection,
    nats: &NatsClient,
    req: CreateCustomerRequest,
) -> Result<Customer> {
    // Start transaction
    let txn = db.begin().await?;

    // 1. Insert business data
    let customer = customer::ActiveModel { ... }.insert(&txn).await?;

    // 2. Insert outbox event (SAME transaction)
    let outbox_event = outbox::ActiveModel {
        aggregate_type: Set("customer".to_string()),
        aggregate_id: Set(customer.id.to_string()),
        event_type: Set("CustomerCreated".to_string()),
        payload: Set(serde_json::to_value(&customer)?),
        nats_subject: Set("aeroxe.v1.customer.customer.created".to_string()),
        status: Set("pending".to_string()),
        ..Default::default()
    };
    outbox_event.insert(&txn).await?;

    // 3. Commit transaction (business data + outbox event atomically)
    txn.commit().await?;

    Ok(customer)
}
```

### 1.5 Outbox Poller (Background Worker)

```rust
pub async fn outbox_poller(db: DatabaseConnection, nats: NatsClient) {
    loop {
        // 1. Fetch pending events (batch of 100)
        let events = outbox::Entity::find()
            .filter(outbox::Column::Status.eq("pending"))
            .order_by_asc(outbox::Column::CreatedAt)
            .limit(100)
            .all(&db).await?;

        for event in events {
            // 2. Publish to NATS
            match nats.publish(&event.nats_subject, &event.payload).await {
                Ok(_) => {
                    // 3. Mark as published
                    let mut active: outbox::ActiveModel = event.into();
                    active.status = Set("published".to_string());
                    active.published_at = Set(Some(Utc::now()));
                    active.update(&db).await?;
                }
                Err(e) => {
                    // 4. Mark as failed, schedule retry
                    let mut active: outbox::ActiveModel = event.into();
                    active.retry_count = Set(active.retry_count.unwrap() + 1);
                    active.status = Set("failed".to_string());
                    active.next_retry_at = Set(Some(Utc::now() + exponential_backoff(retry_count)));
                    active.update(&db).await?;
                }
            }
        }

        // Sleep 1 second before next poll
        tokio::time::sleep(Duration::from_secs(1)).await;
    }
}
```

### 1.6 Outbox Event Lifecycle

```
Created (pending) → Published → Archived
                 ↘ Failed → Retry → Published
                           ↘ Max Retries → Dead Letter
```

### 1.7 Benefits

| Benefit | Description |
|---|---|
| Atomicity | Business data + event committed together |
| Reliability | No event loss on NATS failure |
| Replay | Can re-publish failed events |
| Auditability | Complete event history in PostgreSQL |
| Ordering | Events processed in creation order |

---

## 2. Distributed Locking

### 2.1 Problem

Multiple instances of the monolith may try to process the same resource simultaneously:
- Same callback scheduled twice
- Same campaign started twice
- Same webhook delivery attempted twice
- Same agent execution triggered twice

### 2.2 Solution: Redis Distributed Lock (Redlock)

Use Redis-based distributed locks to ensure only one instance processes a resource at a time.

### 2.3 Lock Table Schema

```sql
CREATE TABLE distributed_locks_.locks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    lock_key VARCHAR(200) NOT NULL UNIQUE,   -- e.g., 'callback:123', 'campaign:456'
    lock_owner VARCHAR(100) NOT NULL,        -- Instance ID + thread ID
    acquired_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    ttl_seconds INT NOT NULL DEFAULT 30,
    metadata JSONB
);

CREATE INDEX idx_locks_key ON distributed_locks_.locks(lock_key);
CREATE INDEX idx_locks_expiry ON distributed_locks_.locks(expires_at);
```

### 2.4 Lock Implementation

```rust
pub struct DistributedLock {
    redis: RedisPool,
    db: DatabaseConnection,
}

impl DistributedLock {
    pub async fn acquire(
        &self,
        key: &str,
        ttl_seconds: u32,
    ) -> Result<Option<LockGuard>> {
        let lock_owner = format!("{}:{}", instance_id(), thread_id());

        // Try Redis first (fast path)
        let redis_key = format!("lock:{}", key);
        let acquired = self.redis.set_nx(
            &redis_key,
            &lock_owner,
            Some(Duration::from_secs(ttl_seconds as u64)),
        ).await?;

        if acquired {
            // Also store in PostgreSQL for persistence
            let lock = distributed_locks::ActiveModel {
                lock_key: Set(key.to_string()),
                lock_owner: Set(lock_owner.clone()),
                expires_at: Set(Utc::now() + chrono::Duration::seconds(ttl_seconds as i64)),
                ttl_seconds: Set(ttl_seconds as i32),
                ..Default::default()
            };
            lock.insert(&self.db).await?;

            return Ok(Some(LockGuard {
                key: key.to_string(),
                owner: lock_owner,
                redis: self.redis.clone(),
                db: self.db.clone(),
            }));
        }

        Ok(None)
    }
}

impl Drop for LockGuard {
    fn drop(&mut self) {
        // Release lock on drop
        let redis = self.redis.clone();
        let key = self.key.clone();
        let owner = self.owner.clone();
        tokio::spawn(async move {
            let redis_key = format!("lock:{}", key);
            let _: () = redis.del(&redis_key).await.unwrap_or(());
            // Also delete from PostgreSQL
        });
    }
}
```

### 2.5 Lock Usage Examples

```rust
// Callback scheduling
let lock = distributed_lock.acquire(&format!("callback:{}", callback_id), 30).await?;
if let Some(_guard) = lock {
    // Process callback (only one instance at a time)
    process_callback(callback_id).await?;
}

// Campaign execution
let lock = distributed_lock.acquire(&format!("campaign:{}", campaign_id), 300).await?;
if let Some(_guard) = lock {
    execute_campaign(campaign_id).await?;
}

// Webhook delivery
let lock = distributed_lock.acquire(&format!("webhook:{}", delivery_id), 10).await?;
if let Some(_guard) = lock {
    deliver_webhook(delivery_id).await?;
}
```

### 2.6 Lock Safety

| Mechanism | Description |
|---|---|
| TTL | Lock auto-expires if holder crashes |
| Owner verification | Only owner can release lock |
| Fencing token | Monotonic token prevents stale writes |
| Heartbeat | Extend lock if operation takes longer than TTL |

---

## 3. Distributed Caching

### 3.1 Problem

Multiple instances of the monolith have separate in-process caches. When one instance updates data, others have stale cache.

### 3.2 Solution: Multi-Tier Caching with Redis

```
L1: In-Process Cache (per instance, fast, small)
    ↓ Miss
L2: Redis Cache (shared, fast, large)
    ↓ Miss
L3: PostgreSQL (source of truth, slow)
```

### 3.3 Cache Strategy

| Strategy | Use Case | Consistency |
|---|---|---|
| **Cache-Aside** | Read-heavy data (config, models) | Eventual |
| **Write-Through** | Critical data (customer, agent) | Strong |
| **Write-Behind** | Analytics, metrics | Eventual |
| **Read-Through** | RAG documents, embeddings | Eventual |

### 3.4 Cache Table Schema

```sql
CREATE TABLE distributed_cache_.cache_metadata (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cache_key VARCHAR(200) NOT NULL UNIQUE,
    cache_tier VARCHAR(20) NOT NULL,        -- l1 | l2
    ttl_seconds INT NOT NULL DEFAULT 300,
    last_accessed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    access_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 3.5 Cache Implementation

```rust
pub struct DistributedCache {
    redis: RedisPool,
    local_cache: DashMap<String, CacheEntry>,  // L1 in-process
}

impl DistributedCache {
    pub async fn get<T: DeserializeOwned>(&self, key: &str) -> Result<Option<T>> {
        // L1: Check in-process cache
        if let Some(entry) = self.local_cache.get(key) {
            if !entry.is_expired() {
                return Ok(Some(entry.value.clone()));
            }
        }

        // L2: Check Redis
        if let Some(data) = self.redis.get(key).await? {
            let value: T = serde_json::from_str(&data)?;
            // Populate L1
            self.local_cache.insert(key.to_string(), CacheEntry::new(value.clone(), 60));
            return Ok(Some(value));
        }

        Ok(None)
    }

    pub async fn set<T: Serialize>(
        &self,
        key: &str,
        value: &T,
        ttl_seconds: u32,
    ) -> Result<()> {
        let data = serde_json::to_string(value)?;

        // L2: Write to Redis
        self.redis.setex(key, &data, ttl_seconds as u64).await?;

        // L1: Write to in-process cache
        self.local_cache.insert(
            key.to_string(),
            CacheEntry::new(value.clone(), ttl_seconds),
        );

        // Publish cache invalidation event
        self.publish_invalidation(key).await?;

        Ok(())
    }

    pub async fn invalidate(&self, key: &str) -> Result<()> {
        // Remove from L1
        self.local_cache.remove(key);

        // Remove from L2
        self.redis.del(key).await?;

        // Publish invalidation to other instances
        self.publish_invalidation(key).await?;

        Ok(())
    }

    async fn publish_invalidation(&self, key: &str) -> Result<()> {
        // Use NATS to notify other instances
        self.nats.publish("aeroxe.v1.cache.invalidated", key).await?;
        Ok(())
    }
}
```

### 3.6 Cache Invalidation

```
Instance A updates data
    |
    v
[1] Write to PostgreSQL
    |
    v
[2] Write to Redis (L2)
    |
    v
[3] Publish invalidation event to NATS
    |
    v
Instance B receives invalidation
    |
    v
[4] Remove from L1 (in-process cache)
    |
    v
[5] Remove from L2 (Redis)
```

### 3.7 Cache Stampede Protection

| Technique | Description |
|---|---|
| **Singleflight** | Only one instance fetches from DB, others wait |
| **Early Expiration** | Refresh cache before it expires |
| **Probabilistic Refresh** | Random instances refresh proactively |
| **Lock-based Refresh** | Distributed lock prevents concurrent refreshes |

---

## 4. Double Entry Ledger

### 4.1 Problem

Financial transactions (billing, payments, refunds, invoicing) require:
- Complete audit trail
- Balance verification
- Error recovery
- Multi-currency support
- Regulatory compliance (SOX, PCI-DSS)

### 4.2 Solution: Double Entry Accounting

Every financial transaction has equal and opposite entries (debit + credit).

### 4.3 Ledger Schema

```sql
CREATE TABLE ledger_.accounts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    account_code VARCHAR(50) NOT NULL,
    account_name VARCHAR(100) NOT NULL,
    account_type VARCHAR(20) NOT NULL,      -- asset | liability | equity | revenue | expense
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, account_code)
);

CREATE TABLE ledger_.transactions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    transaction_id UUID NOT NULL UNIQUE,
    reference_type VARCHAR(50),             -- invoice | payment | refund | adjustment
    reference_id BIGINT,                    -- e.g., invoice_id, payment_id
    description TEXT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    status VARCHAR(20) NOT NULL DEFAULT 'posted',  -- pending | posted | reversed
    posted_at TIMESTAMP,
    reversed_at TIMESTAMP,
    reverse_reason TEXT,
    created_by BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_.entries (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    transaction_id BIGINT NOT NULL REFERENCES ledger_.transactions(id),
    account_id BIGINT NOT NULL REFERENCES ledger_.accounts(id),
    entry_type VARCHAR(10) NOT NULL,        -- debit | credit
    amount DECIMAL(18,4) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    exchange_rate DECIMAL(18,8) DEFAULT 1.0,
    base_amount DECIMAL(18,4) NOT NULL,     -- Amount in base currency
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_.balances (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL REFERENCES ledger_.accounts(id),
    period VARCHAR(7) NOT NULL,             -- YYYY-MM
    debit_total DECIMAL(18,4) NOT NULL DEFAULT 0,
    credit_total DECIMAL(18,4) NOT NULL DEFAULT 0,
    balance DECIMAL(18,4) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, account_id, period)
);
```

### 4.4 Account Types

| Type | Normal Balance | Examples |
|---|---|---|
| **Asset** | Debit | Cash, Accounts Receivable, Prepaid |
| **Liability** | Credit | Accounts Payable, Deferred Revenue |
| **Equity** | Credit | Owner's Equity, Retained Earnings |
| **Revenue** | Credit | Sales Revenue, Service Revenue |
| **Expense** | Debit | Operating Expenses, COGS |

### 4.5 Transaction Implementation

```rust
pub struct LedgerService {
    db: DatabaseConnection,
    distributed_lock: DistributedLock,
}

impl LedgerService {
    pub async fn post_transaction(
        &self,
        tenant_id: TenantId,
        entries: Vec<EntryRequest>,
        description: String,
        reference: Option<(String, i64)>,
    ) -> Result<Transaction> {
        // Validate: debits must equal credits
        let total_debit: Decimal = entries.iter()
            .filter(|e| e.entry_type == "debit")
            .map(|e| e.amount)
            .sum();
        let total_credit: Decimal = entries.iter()
            .filter(|e| e.entry_type == "credit")
            .map(|e| e.amount)
            .sum();

        if total_debit != total_credit {
            return Err(LedgerError::ImbalancedTransaction);
        }

        // Acquire lock to prevent concurrent modifications
        let lock = self.distributed_lock
            .acquire(&format!("ledger:{}", tenant_id), 30).await?;

        if lock.is_none() {
            return Err(LedgerError::ConcurrentModification);
        }

        let txn = self.db.begin().await?;

        // Create transaction
        let transaction = transactions::ActiveModel { ... }.insert(&txn).await?;

        // Create entries
        for entry in entries {
            entries::ActiveModel {
                transaction_id: Set(transaction.id),
                account_id: Set(entry.account_id),
                entry_type: Set(entry.entry_type),
                amount: Set(entry.amount),
                currency: Set(entry.currency),
                exchange_rate: Set(entry.exchange_rate),
                base_amount: Set(entry.base_amount),
                description: Set(entry.description),
                ..Default::default()
            }.insert(&txn).await?;
        }

        // Update balances
        for entry in &entries {
            self.update_balance(&txn, tenant_id, entry.account_id, entry).await?;
        }

        txn.commit().await?;

        Ok(transaction)
    }

    pub async fn verify_balances(
        &self,
        tenant_id: TenantId,
    ) -> Result<BalanceVerification> {
        // Verify all transactions are balanced
        let unbalanced = transactions::Entity::find()
            .filter(transactions::Column::TenantId.eq(tenant_id))
            .filter(transactions::Column::Status.eq("posted"))
            .all(&self.db).await?;

        let mut verification = BalanceVerification::new();
        for txn in unbalanced {
            let entries = entries::Entity::find()
                .filter(entries::Column::TransactionId.eq(txn.id))
                .all(&self.db).await?;

            let total_debit: Decimal = entries.iter()
                .filter(|e| e.entry_type == "debit")
                .map(|e| e.amount)
                .sum();
            let total_credit: Decimal = entries.iter()
                .filter(|e| e.entry_type == "credit")
                .map(|e| e.amount)
                .sum();

            if total_debit != total_credit {
                verification.add_imbalanced(txn.id, total_debit, total_credit);
            }
        }

        Ok(verification)
    }
}
```

### 4.6 Standard Account Chart

```sql
-- Asset Accounts (Debit Normal)
INSERT INTO ledger_.accounts (tenant_id, account_code, account_name, account_type) VALUES
(1, '1000', 'Cash', 'asset'),
(1, '1100', 'Accounts Receivable', 'asset'),
(1, '1200', 'Prepaid Expenses', 'asset');

-- Liability Accounts (Credit Normal)
INSERT INTO ledger_.accounts (tenant_id, account_code, account_name, account_type) VALUES
(1, '2000', 'Accounts Payable', 'liability'),
(1, '2100', 'Deferred Revenue', 'liability'),
(1, '2200', 'Tax Payable', 'liability');

-- Revenue Accounts (Credit Normal)
INSERT INTO ledger_.accounts (tenant_id, account_code, account_name, account_type) VALUES
(1, '4000', 'Subscription Revenue', 'revenue'),
(1, '4100', 'Usage Revenue', 'revenue'),
(1, '4200', 'API Revenue', 'revenue');

-- Expense Accounts (Debit Normal)
INSERT INTO ledger_.accounts (tenant_id, account_code, account_name, account_type) VALUES
(1, '5000', 'LLM Cost', 'expense'),
(1, '5100', 'Telephony Cost', 'expense'),
(1, '5200', 'Infrastructure Cost', 'expense');
```

### 4.7 Financial Reports

| Report | Description |
|---|---|
| **Balance Sheet** | Assets = Liabilities + Equity at a point in time |
| **Income Statement** | Revenue - Expenses over a period |
| **Trial Balance** | Verify all debits = all credits |
| **General Ledger** | Complete transaction history |
| **Aging Report** | Receivables/payables by age |

---

## 5. Integration with Existing Modules

### 5.1 Outbox Pattern Integration

| Module | Events to Outbox |
|---|---|
| `identity` | UserCreated, UserUpdated, RoleAssigned |
| `customer` | CustomerCreated, CustomerSuspended, CustomerActivated |
| `agent` | AgentCompleted, AgentFailed, ToolExecuted |
| `telephony` | CallAnswered, CallEnded, CallTransferred |
| `conversation` | ConversationCreated, ConversationEnded, ConversationEscalated |
| `outbound` | CampaignStarted, CampaignCompleted, CallbackScheduled |
| `webhook` | WebhookDeliveryCompleted, WebhookDeliveryFailed |

### 5.2 Distributed Locking Integration

| Resource | Lock Key | TTL |
|---|---|---|
| Callback scheduling | `callback:{id}` | 30s |
| Campaign execution | `campaign:{id}` | 300s |
| Webhook delivery | `webhook:{delivery_id}` | 10s |
| Agent execution | `agent:{execution_id}` | 60s |
| Ledger transaction | `ledger:{tenant_id}` | 30s |
| Configuration reload | `config:{tenant_id}` | 10s |

### 5.3 Distributed Caching Integration

| Cache Key Pattern | TTL | Strategy |
|---|---|---|
| `config:{tenant_id}:{key}` | 300s | Cache-Aside |
| `agent:{agent_id}:config` | 600s | Cache-Aside |
| `model:{model_name}:status` | 60s | Write-Through |
| `customer:{id}:profile` | 300s | Write-Through |
| `ivr:{flow_id}:definition` | 600s | Cache-Aside |
| `dnc:{tenant_id}:{phone}` | 3600s | Write-Through |

---

## 6. Database Schema Updates

### 6.1 New Schemas Required

| Schema | Purpose |
|---|---|
| `outbox_` | Transactional outbox for reliable event delivery |
| `distributed_locks_` | Distributed lock management |
| `distributed_cache_` | Cache metadata tracking |
| `ledger_` | Double entry financial ledger |

### 6.2 Schema Map Update

```
PostgreSQL 16 Cluster
├── identity_        (Identity module)
├── customer_        (Customer module)
├── ai_              (AI Gateway module)
├── agent_           (Agent module)
├── rag_             (RAG module)
├── vision_          (Vision module)
├── memory_          (Memory module)
├── workflow_        (Workflow module)
├── telephony_       (Telephony module)
├── conversation_    (Conversation module)
├── stt_             (STT module)
├── tts_             (TTS module)
├── analytics_       (Analytics module)
├── webhook_         (Webhook module)
├── outbound_        (Outbound module)
├── audit_           (Audit module)
├── outbox_          (Outbox Pattern)           ← NEW
├── distributed_locks_ (Distributed Locking)   ← NEW
├── distributed_cache_ (Distributed Caching)   ← NEW
└── ledger_          (Double Entry Ledger)      ← NEW
```

---

## 7. NATS Events for Cache Invalidation

| Subject | Event | Purpose |
|---|---|---|
| `aeroxe.v1.cache.invalidated` | `CacheInvalidated` | Notify instances to invalidate cache |
| `aeroxe.v1.cache.refreshed` | `CacheRefreshed` | Notify instances to refresh cache |

---

## 8. Observability

| Pattern | Metrics |
|---|---|
| **Outbox** | `outbox_events_pending`, `outbox_events_published`, `outbox_events_failed`, `outbox_poll_latency_ms` |
| **Locking** | `locks_acquired_total`, `locks_wait_time_ms`, `locks_expired_total`, `locks_contention_total` |
| **Caching** | `cache_hits_total`, `cache_misses_total`, `cache_hit_ratio`, `cache_evictions_total`, `cache_invalidation_latency_ms` |
| **Ledger** | `ledger_transactions_total`, `ledger_entries_total`, `ledger_balance_verification_failures` |
