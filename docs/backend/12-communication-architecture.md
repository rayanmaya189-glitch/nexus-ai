# AeroXe Nexus AI — Communication Architecture

## Trait-Based Internal Calls + Versioned NATS JetStream Event-Driven Design

---

## 1. Communication Strategy Overview

AeroXe Nexus AI follows a **hybrid communication architecture** optimized for modular monoliths:

```
                     External World
                           |
                  REST / WebSocket / HTTPS (versioned: /api/v1/)
                           |
               +---------------------------+
               |    gateway Module         |
               |    (axum HTTP/WS Server)  |
               +---------------------------+
                           |
          ======================================
            Internal Communication (in-process)
          ======================================
              Synchronous          Asynchronous
              Trait Methods          NATS JetStream
                  |                   (versioned subjects)
          Module-to-Module         Event-Driven Flow
          (direct fn calls)        (background jobs)
          ======================================
              External gRPC (optional, versioned service names)
              tonic — for SDK / partner integrations
```

---

## 2. Communication Rules

| Layer | Protocol | Versioning | Usage |
|---|---|---|---|
| External (Web/Mobile) | HTTPS REST | `/api/v{version}/` | Client API |
| Real-time Chat | WebSocket | `/ws/v{version}/` | Token streaming |
| **Internal Synchronous** | **Rust trait methods** | N/A (compile-time) | Module-to-module (in-process) |
| **Internal Asynchronous** | **NATS JetStream (via Outbox Pattern)** | `aeroxe.v{version}.module.event` | Events, background jobs |
| AI Runtime | Ollama HTTP API | N/A | Model inference |
| External gRPC (optional) | tonic | `package.module.v{version}.Service` | SDK / partner integrations |

---

## 3. Versioning Standards

| Medium | Format | Example |
|---|---|---|
| REST API | `/api/v{version}/<resource>` | `/api/v1/auth/login` |
| WebSocket | `/ws/v{version}/<channel>` | `/ws/v1/chat/{conv_id}` |
| NATS Subject | `aeroxe.v{version}.<module>.<event>` | `aeroxe.v1.identity.user.created` |
| External gRPC Package | `<module>.v{version}` | `identity.v1.AuthService` |
| Event Envelope | `"api_version": "v{version}"` | `"api_version": "v1"` |

---

## 4. Trait-Based Internal Communication (Replaces gRPC)

### Design Principles

Every module communicates with other modules through **Rust trait interfaces**:

1. **Modules depend on traits, not implementations** — enables testing with mocks
2. **Traits are defined in each module** — in `api/mod.rs`
3. **Requests are in-process** — `Arc<dyn Trait>` references wired at startup
4. **Strong typing** — all inputs/outputs are typed Rust structs
5. **No serialization overhead** — direct struct passing
6. **Compile-time safety** — missing methods are compile errors

### Trait Definition Pattern

```rust
// src/modules/rag/api/mod.rs
#[async_trait]
pub trait RagService: Send + Sync {
    async fn search(&self, query: SearchQuery) -> Result<SearchResults, RagError>;
    async fn upload_document(&self, request: UploadRequest) -> Result<DocumentStatus>;
}

// src/modules/agent/application/services/agent_service.rs — uses RagService as dependency
pub struct AgentServiceImpl {
    rag: Arc<dyn RagService>,
    memory: Arc<dyn MemoryService>,
    ollama: OllamaClient,
}

#[async_trait]
impl AgentService for AgentServiceImpl {
    async fn start_execution(&self, request: StartAgentRequest) -> Result<ExecutionResponse> {
        // Call RAG synchronously — no gRPC, no serialization
        let docs = self.rag.search(SearchQuery {
            query: request.task.clone(),
            tenant_id: request.tenant_id,
            limit: 5,
        }).await?;

        // Call Ollama directly
        let plan = self.ollama.chat(/* ... */).await?;

        Ok(ExecutionResponse { /* ... */ })
    }
}
```

### Wiring at Application Startup

```rust
// src/main.rs
async fn main() {
    let db = init_db().await;          // SeaORM DatabaseConnection
    let redis = RedisClient::open(config.redis_url)?;
    let nats = NatsClient::connect(config.nats_url).await?;
    let ollama = OllamaClient::new(config.ollama_url);

    let identity: Arc<dyn IdentityService> = Arc::new(
        identity::IdentityServiceImpl::new(db.clone(), redis.clone())
    );
    let customer: Arc<dyn CustomerService> = Arc::new(
        customer::CustomerServiceImpl::new(db.clone(), nats.clone())
    );
    let rag: Arc<dyn RagService> = Arc::new(
        rag::RagServiceImpl::new(db.clone(), nats.clone(), ollama.clone())
    );
    let memory: Arc<dyn MemoryService> = Arc::new(
        memory::MemoryServiceImpl::new(db.clone(), redis.clone(), ollama.clone())
    );
    let agent: Arc<dyn AgentService> = Arc::new(
        agent::AgentServiceImpl::new(rag.clone(), memory.clone(), ollama.clone())
    );

    let app = gateway::build_router(AppState {
        identity, customer, agent, rag, memory, // ...
    });

    axum::serve(listener, app).await?;
}
```

### Performance Comparison

| Aspect | External gRPC | Modular Monolith (Trait) |
|---|---|---|
| Latency per call | 2-5ms (network + serialization) | < 1μs (trait vtable dispatch) |
| Overhead | Protobuf encode/decode | Zero — direct struct passing |
| Error handling | External gRPC status codes | Rust `Result<T, E>` |
| Testing | Need running services | `Mockall` mocks |
| Type safety | Protobuf codegen | Rust compiler |

---

## 5. When to Use NATS vs Trait Calls

| Scenario | Use | Reason |
|---|---|---|
| User requests AI chat | Trait method | Need immediate response |
| Agent needs RAG data | Trait method | Synchronous data needed for reasoning |
| Agent needs memory | Trait method | Synchronous context retrieval |
| Agent execution completes | NATS (versioned) | Other modules need to react asynchronously |
| Customer created | NATS (versioned) | Notify other systems |
| Document uploaded | NATS (versioned) | Long-running processing, don't block client |
| Audit event | NATS (versioned) | Fire-and-forget, must not impact latency |
| Notification send | NATS (versioned) | Background delivery, retry on failure |
| Workflow step completed | NATS (versioned) | Decoupled step orchestration |
| Config change broadcast | NATS (versioned) | All modules must eventually reconfigure |

---

## 6. Module Service Dependency Graph

```
                      gateway
                         |
        +----------------+----------------+----------------+
        |                |                |                |
   identity        ai-gateway        model-registry       |
        |                |                                 |
        |           agent                                  |
        |                |       |        |        |        |
        |                v       v        v        v        |
        |            rag     vision  sql-agent  memory     |
        |                |         |                        |
        +----------------+---------+----------------------+
                             |     |
                      customer  workflow
                             |     |
                             v     v
                      notification  audit

   +=============================================================+
   |              Voice / Telephony Channel                       |
   +=============================================================+
   gateway (webhooks) --> telephony --> stt --> agent --> tts
                             |                    |
                             v                    v
                        conversation          analytics
                             |                    |
                             v                    v
                        outbound              webhook
```

---

## 7. Module API Trait Catalogue

### nexus-identity

```rust
#[async_trait]
pub trait IdentityService: Send + Sync {
    async fn authenticate(&self, req: AuthRequest) -> Result<AuthResponse>;
    async fn verify_token(&self, token: &str) -> Result<JWTClaims>;
    async fn check_permission(&self, req: PermissionRequest) -> Result<bool>;
    async fn validate_tenant(&self, tenant_id: TenantId) -> Result<Tenant>;
    async fn create_user(&self, req: CreateUserRequest) -> Result<User>;
}
```

### nexus-ai-gateway

```rust
#[async_trait]
pub trait AIGatewayService: Send + Sync {
    async fn submit_request(&self, req: AIRequest) -> Result<AIResponse>;
    async fn stream_response(&self, req: AIRequest) -> Result<Receiver<AIChunk>>;
    async fn cancel_request(&self, id: RequestId) -> Result<()>;
    async fn get_session_status(&self, id: SessionId) -> Result<SessionStatus>;
}
```

### nexus-agent

```rust
#[async_trait]
pub trait AgentService: Send + Sync {
    async fn start_execution(&self, req: StartAgentRequest) -> Result<ExecutionResponse>;
    async fn get_execution_status(&self, id: ExecutionId) -> Result<ExecutionStatus>;
    async fn stream_execution(&self, req: StreamRequest) -> Result<Receiver<ExecutionEvent>>;
    async fn cancel_execution(&self, id: ExecutionId) -> Result<()>;
}
```

### nexus-rag

```rust
#[async_trait]
pub trait RagService: Send + Sync {
    async fn search(&self, query: SearchQuery) -> Result<SearchResults>;
    async fn upload_document(&self, req: UploadRequest) -> Result<DocumentStatus>;
    async fn get_document_status(&self, id: DocumentId) -> Result<Option<DocumentStatus>>;
    async fn delete_document(&self, id: DocumentId) -> Result<()>;
}
```

### nexus-vision

```rust
#[async_trait]
pub trait VisionService: Send + Sync {
    async fn analyze_image(&self, req: ImageRequest) -> Result<ImageAnalysisResponse>;
    async fn extract_text(&self, req: ImageRequest) -> Result<OCRResponse>;
    async fn troubleshoot_device(&self, req: DeviceImageRequest) -> Result<TroubleshootResponse>;
}
```

### nexus-sql-agent

```rust
#[async_trait]
pub trait SQLAgentService: Send + Sync {
    async fn generate_query(&self, req: QueryRequest) -> Result<SQLResponse>;
    async fn execute_query(&self, req: SQLRequest) -> Result<ResultResponse>;
    async fn test_connection(&self, req: TestConnectionRequest) -> Result<ConnectionStatus>;
}
```

### nexus-memory

```rust
#[async_trait]
pub trait MemoryService: Send + Sync {
    async fn store(&self, req: StoreMemoryRequest) -> Result<()>;
    async fn search(&self, req: SearchMemoryRequest) -> Result<Vec<MemoryItem>>;
    async fn get_conversation_context(&self, session_id: SessionId) -> Result<Vec<Message>>;
    async fn clear_session(&self, session_id: SessionId) -> Result<()>;
}
```

### nexus-workflow

```rust
#[async_trait]
pub trait WorkflowService: Send + Sync {
    async fn start_workflow(&self, req: StartWorkflowRequest) -> Result<WorkflowResponse>;
    async fn get_status(&self, id: WorkflowId) -> Result<WorkflowStatus>;
    async fn approve_step(&self, req: ApproveRequest) -> Result<()>;
    async fn cancel_workflow(&self, id: WorkflowId) -> Result<()>;
}
```

### nexus-security-ai

```rust
#[async_trait]
pub trait SecurityService: Send + Sync {
    async fn analyze_code(&self, req: CodeReviewRequest) -> Result<CodeReviewResponse>;
    async fn scan_security(&self, req: SecurityScanRequest) -> Result<SecurityReport>;
    async fn scan_prompt(&self, prompt: &str) -> Result<PromptScanResult>;
}
```

### nexus-audit

```rust
#[async_trait]
pub trait AuditService: Send + Sync {
    async fn log_event(&self, event: AuditEvent) -> Result<()>;
    async fn query_events(&self, query: AuditQuery) -> Result<Vec<AuditEvent>>;
    async fn generate_report(&self, req: ReportRequest) -> Result<Report>;
}
```

### nexus-telephony (NEW)

```rust
#[async_trait]
pub trait TelephonyService: Send + Sync {
    // Inbound/Outbound
    async fn handle_inbound_call(&self, req: InboundCallRequest) -> Result<CallResponse>;
    async fn initiate_outbound_call(&self, req: OutboundCallRequest) -> Result<CallResponse>;
    async fn answer_call(&self, call_id: CallId) -> Result<()>;
    async fn end_call(&self, call_id: CallId, reason: String) -> Result<()>;
    // Call control
    async fn hold_call(&self, call_id: CallId) -> Result<()>;
    async fn resume_call(&self, call_id: CallId) -> Result<()>;
    async fn transfer_call(&self, req: TransferRequest) -> Result<()>;
    // Audio
    async fn send_audio(&self, call_id: CallId, audio: AudioFrame) -> Result<()>;
    async fn receive_audio(&self, call_id: CallId) -> Result<Receiver<AudioFrame>>;
    // Caller authentication (NEW)
    async fn authenticate_caller(&self, call_id: CallId, method: AuthMethod) -> Result<CallerAuthResult>;
    async fn verify_pin(&self, call_id: CallId, pin: String) -> Result<bool>;
    async fn verify_voice_biometric(&self, call_id: CallId, voice_sample: Vec<u8>) -> Result<f32>;
    // Anti-fraud (NEW)
    async fn check_fraud(&self, call_id: CallId) -> Result<FraudCheckResult>;
    // Voicemail (NEW)
    async fn start_voicemail(&self, call_id: CallId) -> Result<VoicemailId>;
    async fn end_voicemail(&self, call_id: CallId) -> Result<VoicemailId>;
    // IVR (NEW)
    async fn start_ivr_flow(&self, call_id: CallId, flow_id: FlowId) -> Result<()>;
    async fn handle_dtmf_input(&self, call_id: CallId, digit: char) -> Result<IVRResponse>;
    // Live monitoring (NEW)
    async fn start_monitoring(&self, call_id: CallId, supervisor_id: UserId, action: MonitorAction) -> Result<()>;
    async fn end_monitoring(&self, call_id: CallId) -> Result<()>;
    // Query
    async fn get_call_status(&self, call_id: CallId) -> Result<CallStatus>;
    async fn get_voicemails(&self, tenant_id: TenantId) -> Result<Vec<Voicemail>>;
}
```

### nexus-conversation (NEW)

```rust
#[async_trait]
pub trait ConversationService: Send + Sync {
    async fn create_conversation(&self, req: CreateConversationRequest) -> Result<Conversation>;
    async fn get_conversation(&self, id: ConversationId) -> Result<Option<Conversation>>;
    async fn transition_state(&self, id: ConversationId, trigger: TransitionTrigger) -> Result<ConversationState>;
    async fn add_message(&self, id: ConversationId, msg: NewMessage) -> Result<Message>;
    async fn get_context(&self, id: ConversationId) -> Result<ConversationContext>;
    async fn end_conversation(&self, id: ConversationId, outcome: ConversationOutcome) -> Result<()>;
}
```

### nexus-stt (NEW)

```rust
#[async_trait]
pub trait STTService: Send + Sync {
    async fn start_streaming_session(&self, req: StreamingSessionRequest) -> Result<StreamingSessionHandle>;
    async fn send_audio_chunk(&self, session_id: SessionId, chunk: AudioChunk) -> Result<PartialTranscript>;
    async fn end_streaming_session(&self, session_id: SessionId) -> Result<FinalTranscript>;
    async fn transcribe_audio(&self, req: TranscribeRequest) -> Result<Transcript>;
    // Confidence threshold (NEW)
    async fn get_config(&self, tenant_id: TenantId) -> Result<STTConfig>;
    async fn update_config(&self, tenant_id: TenantId, config: STTConfig) -> Result<()>;
    // Anti-injection (NEW)
    async fn check_liveness(&self, audio: Vec<u8>) -> Result<LivenessResult>;
}
```

### nexus-tts (NEW)

```rust
#[async_trait]
pub trait TTSService: Send + Sync {
    async fn start_streaming_synthesis(&self, req: StreamingSynthesisRequest) -> Result<StreamingSynthesisHandle>;
    async fn synthesize_chunk(&self, session_id: SessionId, text: String) -> Result<Receiver<AudioChunk>>;
    async fn synthesize(&self, req: SynthesisRequest) -> Result<SynthesisResult>;
    async fn list_voices(&self, tenant_id: TenantId) -> Result<Vec<VoiceProfile>>;
    // Voice cloning (NEW)
    async fn clone_voice(&self, req: VoiceCloneRequest) -> Result<VoiceClone>;
    async fn revoke_clone(&self, clone_id: CloneId) -> Result<()>;
    // Sentiment adaptation (NEW)
    async fn synthesize_with_sentiment(&self, req: SentimentSynthesisRequest) -> Result<SynthesisResult>;
    // Post-call survey (NEW)
    async fn play_survey_prompt(&self, call_id: CallId) -> Result<()>;
}
```

### nexus-analytics (NEW)

```rust
#[async_trait]
pub trait AnalyticsService: Send + Sync {
    async fn get_dashboard(&self, tenant_id: TenantId, time_range: TimeRange) -> Result<Dashboard>;
    async fn get_conversation_metrics(&self, req: ConversationMetricsRequest) -> Result<ConversationMetrics>;
    async fn get_call_metrics(&self, req: CallMetricsRequest) -> Result<CallMetrics>;
    async fn get_agent_performance(&self, agent_id: AgentId, time_range: TimeRange) -> Result<AgentPerformance>;
    async fn get_cost_breakdown(&self, tenant_id: TenantId, time_range: TimeRange) -> Result<CostBreakdown>;
}
```

### nexus-webhook (NEW)

```rust
#[async_trait]
pub trait WebhookService: Send + Sync {
    async fn create_subscription(&self, req: CreateWebhookRequest) -> Result<WebhookSubscription>;
    async fn delete_subscription(&self, id: SubscriptionId) -> Result<()>;
    async fn list_subscriptions(&self, tenant_id: TenantId) -> Result<Vec<WebhookSubscription>>;
    async fn test_webhook(&self, id: SubscriptionId) -> Result<WebhookTestResult>;
}
```

### nexus-outbound (NEW)

```rust
#[async_trait]
pub trait OutboundService: Send + Sync {
    async fn create_campaign(&self, req: CreateCampaignRequest) -> Result<Campaign>;
    async fn start_campaign(&self, id: CampaignId) -> Result<()>;
    async fn make_outbound_call(&self, req: OutboundCallRequest) -> Result<OutboundCallResult>;
    async fn schedule_callback(&self, req: ScheduleCallbackRequest) -> Result<ScheduledCallback>;
    async fn check_dnc(&self, phone: PhoneNumber, tenant_id: TenantId) -> Result<bool>;
}
```

---

## 8. NATS JetStream Architecture

### Subject Naming Standard

Format: `aeroxe.v{version}.<module>.<event>`

Currently active version: **`v1`**

### All Subjects (Versioned)

| Subject | Module | Description |
|---|---|---|
| `aeroxe.v1.ai.request.created` | AI Gateway | New AI request |
| `aeroxe.v1.ai.response.generated` | AI Gateway | Response ready |
| `aeroxe.v1.ai.failed` | AI Gateway | Request failed |
| `aeroxe.v1.agent.started` | Agent | Agent execution started |
| `aeroxe.v1.agent.completed` | Agent | Agent execution done |
| `aeroxe.v1.agent.failed` | Agent | Agent execution failed |
| `aeroxe.v1.agent.tool.executed` | Agent | Tool call made |
| `aeroxe.v1.rag.document.uploaded` | RAG | Document received |
| `aeroxe.v1.rag.document.processed` | RAG | Processing done |
| `aeroxe.v1.rag.embedding.created` | RAG | Embeddings stored |
| `aeroxe.v1.rag.knowledge.updated` | RAG | Knowledge base modified |
| `aeroxe.v1.vision.image.received` | Vision | Image uploaded |
| `aeroxe.v1.vision.analysis.completed` | Vision | Analysis done |
| `aeroxe.v1.workflow.started` | Workflow | Workflow started |
| `aeroxe.v1.workflow.step.completed` | Workflow | Step finished |
| `aeroxe.v1.workflow.completed` | Workflow | All steps complete |
| `aeroxe.v1.workflow.failed` | Workflow | Workflow error |
| `aeroxe.v1.security.scan.started` | Security | Scan initiated |
| `aeroxe.v1.security.threat.detected` | Security | Threat found |
| `aeroxe.v1.customer.customer.created` | Customer | Customer created |
| `aeroxe.v1.customer.customer.activated` | Customer | Customer activated |
| `aeroxe.v1.customer.customer.suspended` | Customer | Customer suspended |
| `aeroxe.v1.customer.customer.updated` | Customer | Customer updated |
| `aeroxe.v1.audit.*` | Audit | All audit events |
| `aeroxe.v1.identity.*` | Identity | User/tenant events |
| `aeroxe.v1.memory.*` | Memory | Memory lifecycle events |
| `aeroxe.v1.gateway.*` | Gateway | Gateway operational events |
| `aeroxe.v1.config.*` | Config | Configuration changes |
| `aeroxe.v1.telephony.call.*` | Telephony | Call lifecycle events |
| `aeroxe.v1.conversation.*` | Conversation | Conversation state events |
| `aeroxe.v1.outbound.*` | Outbound | Campaign and callback events |
| `aeroxe.v1.webhook.*` | Webhook | Webhook delivery events |

---

## 9. Event Schema Standard (Versioned)

Every NATS event includes the API version in the envelope:

```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "AgentCompleted",
  "version": "1.0",
  "api_version": "v1",
  "timestamp": "2026-07-21T12:00:00Z",
  "tenant_id": 1,
  "module": "agent",
  "data": {
    "execution_id": 12345,
    "agent": "customer-agent",
    "status": "success",
    "tokens_used": 1250,
    "latency_ms": 3500
  }
}
```

---

## 10. JetStream Stream Design

| Stream Name | Subjects | Retention | Replication |
|---|---|---|---|
| `AI_EVENTS_V1` | `aeroxe.v1.ai.*` | 7 days | 3 |
| `AGENT_EVENTS_V1` | `aeroxe.v1.agent.*` | 30 days | 3 |
| `RAG_EVENTS_V1` | `aeroxe.v1.rag.*` | 14 days | 1 |
| `CUSTOMER_EVENTS_V1` | `aeroxe.v1.customer.*` | 30 days | 1 |
| `AUDIT_EVENTS_V1` | `aeroxe.v1.audit.*` | 365 days | 3 |
| `WORKFLOW_EVENTS_V1` | `aeroxe.v1.workflow.*` | 30 days | 1 |
| `SECURITY_EVENTS_V1` | `aeroxe.v1.security.*` | 365 days | 3 |
| `IDENTITY_EVENTS_V1` | `aeroxe.v1.identity.*` | 365 days | 3 |

---

## 11. Versioning Strategy Details

### 10.1 When to Bump NATS Subject Version

| Change | Version Bump |
|---|---|
| New optional field in event data | Minor (same v1) |
| New event type | Minor (same v1) |
| Remove field from event data | Major (v1 → v2) |
| Change event semantics | Major (v1 → v2) |
| Change serialization format | Major (v1 → v2) |

### 10.2 Coexistence Strategy

Multiple versions can coexist:
```
aeroxe.v1.customer.customer.created   # Old consumers continue
```

### 10.3 External gRPC Versioning

```protobuf
// proto/identity/v1/auth_service.proto
package identity.v1;

service AuthService {
    rpc Authenticate(AuthRequest) returns (AuthResponse);
}

// proto/customer/v1/customer_service.proto
package customer.v1;

service CustomerService {
    rpc CreateCustomer(CreateCustomerRequest) returns (Customer);
    rpc GetCustomer(GetCustomerRequest) returns (Customer);
    rpc SuspendCustomer(SuspendCustomerRequest) returns (Customer);
}
```

---

## 12. Request Flow Example

**User asks:** "Why is customer internet slow?"

```
User → HTTP POST /api/v1/ai/chat
  → gateway (auth, rate-limit, version check)
    → ai-gateway::submit_request() [trait call]
      → agent::start_execution() [trait call]
        → Ollama: LFM2.5 Thinking (intent detection)
          → Plan: check customer, check network, search knowledge
        → rag::search() [trait call] — document knowledge
        → memory::search() [trait call] — past conversations
        → Ollama: Command-R 7B (generate final answer)
      ← Response + audit event (NATS: aeroxe.v1.audit.ai.request)
    ← HTTP 200 + JSON response
```

Note: **Every arrow is an in-process Rust trait method call**, not a network hop. The entire flow completes in < 3 seconds.

---

## 13. Streaming Response Architecture

```
Client WebSocket (/ws/v1/chat/{conversation_id})
    |
    v
gateway (axum WebSocket handler)
    |
    v
ai-gateway::stream_response()
    |  → returns tokio::sync::mpsc::Receiver<AIChunk>
    v
agent::stream_execution()
    |  → returns tokio::sync::mpsc::Receiver<ExecutionEvent>
    v
Ollama HTTP streaming API
    |
    v
Token stream → Receiver → WebSocket → Client
```

All streams use **tokio channels** for zero-copy token relay between modules.

---

## 14. Security Requirements

### Internal Trait Calls

| Requirement | Implementation |
|---|---|
| Authentication | JWT validated by gateway, claims attached to request |
| Authorization | identity::check_permission() called by gateway |
| Tenant Isolation | tenant_id propagated via RequestContext struct |
| Rate Limiting | Token bucket in gateway middleware |

### NATS Security

| Requirement | Implementation |
|---|---|
| TLS | All NATS connections encrypted |
| Account Isolation | Separate accounts per module |
| Subject Permissions | Publish/subscribe ACLs per version |
| Authentication | NKey or JWT-based auth |

---

## 15. Testing Communication Contracts

### Module Boundary Tests (TDD)

```rust
/// Contract: agent must be able to call rag::search()
#[tokio::test]
async fn test_agent_rag_integration() {
    let rag = MockRagService::new();
    let agent = AgentServiceImpl::new(Arc::new(rag), /* ... */);

    let result = agent.start_execution(/* ... */).await;
    assert!(result.is_ok());
}
```

### NATS Contract Tests (Versioned Subjects)

```rust
#[tokio::test]
async fn test_agent_completed_event_is_published() {
    let nats = nats_test_server().await;
    let agent = AgentServiceImpl::new(/* ... with nats */);

    agent.start_execution(/* ... */).await;

    // Verify event was published with correct versioned subject
    let msg = nats.subscribe("aeroxe.v1.agent.completed").await?;
    assert_eq!(msg.subject, "aeroxe.v1.agent.completed");
    // ...
}
```
