# AeroXe Nexus AI — Communication Architecture

## gRPC (sync) / NATS (async) Internal Calls + Protobuf + Versioned NATS JetStream Event-Driven Design

---

## 1. Communication Strategy Overview

AeroXe Nexus AI follows a **hybrid communication architecture** optimized for modular monoliths:

```
                     External World
                           |
                  HTTPS / WebSocket (Protobuf serialized as JSON)
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
              gRPC (tonic)          NATS JetStream
                  |                   (Protobuf payloads)
          Module-to-Module         Event-Driven Flow
          (direct tonic calls)     (background jobs)
          ======================================
              External gRPC (optional, versioned service names)
              tonic — for SDK / partner integrations
```

---

## 2. Communication Rules

| Layer | Protocol | Format | Versioning | Usage |
|---|---|---|---|---|
| External (Web/Mobile) | HTTPS | **Protobuf (proto3) serialized as JSON** | `/api/v{version}/` | Client API |
| Real-time Chat | WebSocket | Protobuf chunks | `/ws/v{version}/` | Token streaming |
| **Internal Synchronous** | **gRPC (tonic)** | **Protobuf** | `aeroxe.v{version}.<service>` | Module-to-module (in-process) |
| **Internal Asynchronous** | **NATS JetStream (via Outbox Pattern)** | **Protobuf payloads** | `aeroxe.v{version}.module.event` | Events, background jobs |
| AI Runtime | Ollama HTTP API | JSON | N/A | Model inference |
| External gRPC (optional) | tonic | Protobuf | `package.module.v{version}.Service` | SDK / partner integrations |

**HTTP API Rules:**
- **Allowed methods:** PATCH, POST, DELETE only
- **Read operations:** Use POST with a request body (no GET)
- **No PUT** — use PATCH for partial updates

---

## 3. Versioning Standards

| Medium | Format | Example |
|---|---|---|
| HTTP API | `/api/v{version}/<resource>` | `/api/v1/auth/login` |
| WebSocket | `/ws/v{version}/<channel>` | `/ws/v1/chat/{conv_id}` |
| Internal gRPC | `aeroxe.v{version}.<service>.<method>` | `aeroxe.v1.vision.VisionService.AnalyzeImage` |
| NATS Subject | `aeroxe.v{version}.<module>.<event>` | `aeroxe.v1.identity.user.created` |
| External gRPC Package | `<module>.v{version}` | `identity.v1.AuthService` |
| Event Envelope | `"api_version": "v{version}"` | `"api_version": "v1"` |

---

## 4. gRPC Internal Communication (tonic)

### Design Principles

Every module communicates with other modules through **gRPC service interfaces** via tonic:

1. **Modules expose gRPC services** — each module defines `.proto` service definitions
2. **Protobuf messages** — all request/response are proto3 messages serialized as JSON over HTTP (external) or binary (internal)
3. **In-process tonic** — modules call each other via tonic channels within the same binary (no network overhead)
4. **Strong typing** — all inputs/outputs are Protobuf messages with codegen
5. **Compile-time safety** — proto codegen catches mismatched types at build time

### Proto Service Definition Pattern

```protobuf
// proto/vision/v1/vision_service.proto
syntax = "proto3";
package aeroxe.vision.v1;

service VisionService {
  rpc AnalyzeImage(AnalyzeImageRequest) returns (AnalyzeImageResponse);
  rpc ExtractText(ExtractTextRequest) returns (ExtractTextResponse);
  rpc TroubleshootDevice(TroubleshootDeviceRequest) returns (TroubleshootDeviceResponse);
}

message AnalyzeImageRequest {
  bytes image = 1;
  string image_type = 2;
  string task = 3;
  int64 tenant_id = 4;
}

message AnalyzeImageResponse {
  string description = 1;
  float confidence = 2;
  repeated DetectedObject objects = 3;
  map<string, string> metadata = 4;
  double latency_ms = 5;
}
```

### In-Process gRPC Wiring

```rust
// src/main.rs
async fn main() {
    let db = init_db().await;
    let redis = RedisClient::open(config.redis_url)?;
    let nats = NatsClient::connect(config.nats_url).await?;
    let ollama = OllamaClient::new(config.ollama_url);

    // Each module runs its gRPC server on an in-process tonic channel
    let identity_server = identity::start_grpc_server(db.clone(), redis.clone());
    let customer_server = customer::start_grpc_server(db.clone(), nats.clone());
    let rag_server = rag::start_grpc_server(db.clone(), nats.clone(), ollama.clone());
    let memory_server = memory::start_grpc_server(db.clone(), redis.clone(), ollama.clone());
    let vision_server = vision::start_grpc_server(db.clone(), ollama.clone());
    let agent_server = agent::start_grpc_server(
        rag_server.clone(),
        memory_server.clone(),
        vision_server.clone(),
        ollama.clone(),
    );

    // Gateway connects to module gRPC servers via in-process channels
    let app = gateway::build_router(AppState {
        identity: identity_server.clone(),
        customer: customer_server.clone(),
        agent: agent_server.clone(),
        rag: rag_server.clone(),
        memory: memory_server.clone(),
        vision: vision_server.clone(),
    });

    axum::serve(listener, app).await?;
}
```

### Performance: In-Process gRPC vs Network gRPC

| Aspect | External gRPC (network) | In-Process gRPC (modular monolith) |
|---|---|---|
| Latency per call | 2-5ms (network + serialization) | < 50μs (tonic channel, no network) |
| Overhead | Protobuf encode/decode + network | Protobuf encode/decode only |
| Error handling | gRPC status codes | gRPC status codes (consistent) |
| Testing | Need running services | tonic `Channel::from_static` mocks |
| Type safety | Protobuf codegen | Protobuf codegen |
| Serialization format | Protobuf binary | Protobuf binary (same as external) |

---

## 5. When to Use NATS vs gRPC Calls

| Scenario | Use | Reason |
|---|---|---|
| User requests AI chat | gRPC call | Need immediate response |
| Agent needs RAG data | gRPC call | Synchronous data needed for reasoning |
| Agent needs memory | gRPC call | Synchronous context retrieval |
| Agent execution completes | NATS (Protobuf) | Other modules need to react asynchronously |
| Customer created | NATS (Protobuf) | Notify other systems |
| Document uploaded | NATS (Protobuf) | Long-running processing, don't block client |
| Audit event | NATS (Protobuf) | Fire-and-forget, must not impact latency |
| Notification send | NATS (Protobuf) | Background delivery, retry on failure |
| Workflow step completed | NATS (Protobuf) | Decoupled step orchestration |
| Config change broadcast | NATS (Protobuf) | All modules must eventually reconfigure |

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

## 7. Module gRPC Service Catalogue

Each module exposes a gRPC service defined in `.proto` files. Below are the service interfaces (proto-generated Rust traits):

### nexus-identity (gRPC Service)

```protobuf
// proto/identity/v1/identity_service.proto
service IdentityService {
  rpc Authenticate(AuthRequest) returns (AuthResponse);
  rpc VerifyToken(VerifyTokenRequest) returns (JWTClaims);
  rpc CheckPermission(PermissionRequest) returns (PermissionResponse);
  rpc ValidateTenant(ValidateTenantRequest) returns (Tenant);
  rpc CreateUser(CreateUserRequest) returns (User);
}
```

### nexus-ai-gateway (gRPC Service)

```protobuf
// proto/ai-gateway/v1/ai_gateway_service.proto
service AIGatewayService {
  rpc SubmitRequest(AIRequest) returns (AIResponse);
  rpc StreamResponse(AIRequest) returns (stream AIChunk);
  rpc CancelRequest(CancelRequest) returns (Empty);
  rpc GetSessionStatus(SessionStatusRequest) returns (SessionStatus);
}
```

### nexus-agent (gRPC Service)

```protobuf
// proto/agent/v1/agent_service.proto
service AgentService {
  rpc StartExecution(StartAgentRequest) returns (ExecutionResponse);
  rpc GetExecutionStatus(ExecutionStatusRequest) returns (ExecutionStatus);
  rpc StreamExecution(StreamRequest) returns (stream ExecutionEvent);
  rpc CancelExecution(CancelRequest) returns (Empty);
}
```

### nexus-rag (gRPC Service)

```protobuf
// proto/rag/v1/rag_service.proto
service RagService {
  rpc Search(SearchRequest) returns (SearchResults);
  rpc UploadDocument(UploadRequest) returns (DocumentStatus);
  rpc GetDocumentStatus(DocumentStatusRequest) returns (DocumentStatus);
  rpc DeleteDocument(DeleteRequest) returns (Empty);
}
```

### nexus-vision (gRPC Service)

```protobuf
// proto/vision/v1/vision_service.proto
service VisionService {
  rpc AnalyzeImage(AnalyzeImageRequest) returns (AnalyzeImageResponse);
  rpc ExtractText(ExtractTextRequest) returns (ExtractTextResponse);
  rpc TroubleshootDevice(TroubleshootDeviceRequest) returns (TroubleshootDeviceResponse);
}
```

### nexus-sql-agent (gRPC Service)

```protobuf
// proto/sql-agent/v1/sql_agent_service.proto
service SQLAgentService {
  rpc GenerateQuery(GenerateQueryRequest) returns (SQLResponse);
  rpc ExecuteQuery(ExecuteQueryRequest) returns (ResultResponse);
  rpc TestConnection(TestConnectionRequest) returns (ConnectionStatus);
}
```

### nexus-memory (gRPC Service)

```protobuf
// proto/memory/v1/memory_service.proto
service MemoryService {
  rpc Store(StoreMemoryRequest) returns (Empty);
  rpc Search(SearchMemoryRequest) returns (SearchMemoryResponse);
  rpc GetConversationContext(GetContextRequest) returns (ConversationContext);
  rpc ClearSession(ClearSessionRequest) returns (Empty);
}
```

### nexus-workflow (gRPC Service)

```protobuf
// proto/workflow/v1/workflow_service.proto
service WorkflowService {
  rpc StartWorkflow(StartWorkflowRequest) returns (WorkflowResponse);
  rpc GetStatus(GetStatusRequest) returns (WorkflowStatus);
  rpc ApproveStep(ApproveRequest) returns (Empty);
  rpc CancelWorkflow(CancelWorkflowRequest) returns (Empty);
}
```

### nexus-security-ai (gRPC Service)

```protobuf
// proto/security-ai/v1/security_service.proto
service SecurityService {
  rpc AnalyzeCode(CodeReviewRequest) returns (CodeReviewResponse);
  rpc ScanSecurity(SecurityScanRequest) returns (SecurityReport);
  rpc ScanPrompt(ScanPromptRequest) returns (PromptScanResult);
}
```

### nexus-audit (gRPC Service)

```protobuf
// proto/audit/v1/audit_service.proto
service AuditService {
  rpc LogEvent(AuditEvent) returns (Empty);
  rpc QueryEvents(AuditQuery) returns (AuditEvents);
  rpc GenerateReport(ReportRequest) returns (Report);
}
```

### nexus-telephony (gRPC Service)

```protobuf
// proto/telephony/v1/telephony_service.proto
service TelephonyService {
  // Inbound/Outbound
  rpc HandleInboundCall(InboundCallRequest) returns (CallResponse);
  rpc InitiateOutboundCall(OutboundCallRequest) returns (CallResponse);
  rpc AnswerCall(AnswerCallRequest) returns (Empty);
  rpc EndCall(EndCallRequest) returns (Empty);
  // Call control
  rpc HoldCall(HoldCallRequest) returns (Empty);
  rpc ResumeCall(ResumeCallRequest) returns (Empty);
  rpc TransferCall(TransferRequest) returns (Empty);
  // Audio
  rpc SendAudio(SendAudioRequest) returns (Empty);
  rpc ReceiveAudio(ReceiveAudioRequest) returns (stream AudioFrame);
  // Caller authentication
  rpc AuthenticateCaller(AuthenticateCallerRequest) returns (CallerAuthResult);
  rpc VerifyPin(VerifyPinRequest) returns (VerifyPinResponse);
  rpc VerifyVoiceBiometric(VerifyVoiceBiometricRequest) returns (BiometricResult);
  // Anti-fraud
  rpc CheckFraud(CheckFraudRequest) returns (FraudCheckResult);
  // Voicemail
  rpc StartVoicemail(StartVoicemailRequest) returns (VoicemailResponse);
  rpc EndVoicemail(EndVoicemailRequest) returns (VoicemailResponse);
  // IVR
  rpc StartIVRFlow(StartIVRRequest) returns (Empty);
  rpc HandleDTMFInput(DTMFRequest) returns (IVRResponse);
  // Live monitoring
  rpc StartMonitoring(StartMonitoringRequest) returns (Empty);
  rpc EndMonitoring(EndMonitoringRequest) returns (Empty);
  // Query
  rpc GetCallStatus(GetCallStatusRequest) returns (CallStatus);
  rpc GetVoicemails(GetVoicemailsRequest) returns (Voicemails);
}
```

### nexus-conversation (gRPC Service)

```protobuf
// proto/conversation/v1/conversation_service.proto
service ConversationService {
  rpc CreateConversation(CreateConversationRequest) returns (Conversation);
  rpc GetConversation(GetConversationRequest) returns (Conversation);
  rpc TransitionState(TransitionStateRequest) returns (ConversationState);
  rpc AddMessage(AddMessageRequest) returns (Message);
  rpc GetContext(GetContextRequest) returns (ConversationContext);
  rpc EndConversation(EndConversationRequest) returns (Empty);
}
```

### nexus-stt (gRPC Service)

```protobuf
// proto/stt/v1/stt_service.proto
service STTService {
  rpc StartStreamingSession(StreamingSessionRequest) returns (StreamingSessionHandle);
  rpc SendAudioChunk(SendAudioChunkRequest) returns (PartialTranscript);
  rpc EndStreamingSession(EndStreamingSessionRequest) returns (FinalTranscript);
  rpc TranscribeAudio(TranscribeRequest) returns (Transcript);
  rpc GetConfig(GetConfigRequest) returns (STTConfig);
  rpc UpdateConfig(UpdateConfigRequest) returns (Empty);
  rpc CheckLiveness(CheckLivenessRequest) returns (LivenessResult);
}
```

### nexus-tts (gRPC Service)

```protobuf
// proto/tts/v1/tts_service.proto
service TTSService {
  rpc StartStreamingSynthesis(StreamingSynthesisRequest) returns (StreamingSynthesisHandle);
  rpc SynthesizeChunk(SynthesizeChunkRequest) returns (stream AudioChunk);
  rpc Synthesize(SynthesizeRequest) returns (SynthesisResult);
  rpc ListVoices(ListVoicesRequest) returns (Voices);
  rpc CloneVoice(VoiceCloneRequest) returns (VoiceClone);
  rpc RevokeClone(RevokeCloneRequest) returns (Empty);
  rpc SynthesizeWithSentiment(SentimentSynthesisRequest) returns (SynthesisResult);
  rpc PlaySurveyPrompt(PlaySurveyRequest) returns (Empty);
}
```

### nexus-analytics (gRPC Service)

```protobuf
// proto/analytics/v1/analytics_service.proto
service AnalyticsService {
  rpc GetDashboard(DashboardRequest) returns (Dashboard);
  rpc GetConversationMetrics(ConversationMetricsRequest) returns (ConversationMetrics);
  rpc GetCallMetrics(CallMetricsRequest) returns (CallMetrics);
  rpc GetAgentPerformance(AgentPerformanceRequest) returns (AgentPerformance);
  rpc GetCostBreakdown(CostBreakdownRequest) returns (CostBreakdown);
}
```

### nexus-webhook (gRPC Service)

```protobuf
// proto/webhook/v1/webhook_service.proto
service WebhookService {
  rpc CreateSubscription(CreateWebhookRequest) returns (WebhookSubscription);
  rpc DeleteSubscription(DeleteSubscriptionRequest) returns (Empty);
  rpc ListSubscriptions(ListSubscriptionsRequest) returns (WebhookSubscriptions);
  rpc TestWebhook(TestWebhookRequest) returns (WebhookTestResult);
}
```

### nexus-outbound (gRPC Service)

```protobuf
// proto/outbound/v1/outbound_service.proto
service OutboundService {
  rpc CreateCampaign(CreateCampaignRequest) returns (Campaign);
  rpc StartCampaign(StartCampaignRequest) returns (Empty);
  rpc MakeOutboundCall(OutboundCallRequest) returns (OutboundCallResult);
  rpc ScheduleCallback(ScheduleCallbackRequest) returns (ScheduledCallback);
  rpc CheckDNC(CheckDNCRequest) returns (DNCResponse);
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

## 9. Event Schema Standard (Versioned — Protobuf)

Every NATS event is a **Protobuf message** serialized as JSON. The envelope includes the API version:

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

**All event payloads are Protobuf messages** — the JSON representation is for human readability; wire format is Protobuf binary.

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
User → HTTPS POST /api/v1/ai/chat (Protobuf JSON body)
  → gateway (auth, rate-limit, version check)
    → ai-gateway::SubmitRequest() [gRPC call]
      → agent::StartExecution() [gRPC call]
        → Ollama: LFM2.5 Thinking (intent detection)
          → Plan: check customer, check network, search knowledge
        → rag::Search() [gRPC call] — document knowledge
        → memory::Search() [gRPC call] — past conversations
        → Ollama: Command-R 7B (generate final answer)
      ← Response + audit event (NATS: aeroxe.v1.audit.ai.request)
    ← HTTP 200 + Protobuf JSON response
```

Note: **Every internal arrow is a gRPC call via tonic** — binary Protobuf serialization, in-process channels, no network overhead. The entire flow completes in < 3 seconds.

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

### Internal gRPC Calls

| Requirement | Implementation |
|---|---|
| Authentication | JWT validated by gateway, claims attached to RequestContext |
| Authorization | identity::CheckPermission() gRPC call |
| Tenant Isolation | tenant_id propagated via RequestContext in all gRPC calls |
| Rate Limiting | Token bucket in gateway middleware |

### NATS Security

| Requirement | Implementation |
|---|---|
| TLS | All NATS connections encrypted |
| Payload Format | All payloads are Protobuf messages |
| Account Isolation | Separate accounts per module |
| Subject Permissions | Publish/subscribe ACLs per version |
| Authentication | NKey or JWT-based auth |

---

## 15. Testing Communication Contracts

### Module Boundary Tests (TDD)

```rust
/// Contract: agent must be able to call rag::Search() via gRPC
#[tokio::test]
async fn test_agent_rag_integration() {
    let rag_server = MockRagServiceServer::new();
    let agent = AgentServiceImpl::new(rag_server.clone(), /* ... */);

    let result = agent.start_execution(/* ... */).await;
    assert!(result.is_ok());
}
```

### NATS Contract Tests (Versioned Subjects — Protobuf Payloads)

```rust
#[tokio::test]
async fn test_agent_completed_event_is_published() {
    let nats = nats_test_server().await;
    let agent = AgentServiceImpl::new(/* ... with nats */);

    agent.start_execution(/* ... */).await;

    // Verify Protobuf event was published with correct versioned subject
    let msg = nats.subscribe("aeroxe.v1.agent.completed").await?;
    assert_eq!(msg.subject, "aeroxe.v1.agent.completed");
    // Deserialize as Protobuf
    let event = AgentCompletedEvent::decode(msg.data.as_ref())?;
    // ...
}
```
