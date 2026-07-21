# AeroXe Nexus AI — Telephony Module

## Voice Channel, SIP/WebRTC Integration, Call Management & Real-Time Audio

> **Modular Monolith Module:** This document describes the `nexus-telephony` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces (see [Communication Architecture](12-communication-architecture.md)).

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-telephony` |
| Crate | `nexus-telephony` (workspace member) |
| Bounded Context | Telephony |
| Domain Type | Core Domain |
| Language | Rust (edition 2024) |
| Schema | `telephony_` (in shared PostgreSQL) |
| Dependencies | `nexus-agent` (AI execution), `nexus-stt` (speech-to-text), `nexus-tts` (text-to-speech), `nexus-conversation` (conversation state), `nexus-memory` (context), `nexus-audit` (logging) |

---

## 2. Purpose

The Telephony module is the **voice channel gateway** for AeroXe Nexus AI. It enables:

- Receiving inbound voice calls (SIP, WebRTC, PSTN)
- Making outbound voice calls
- Real-time audio streaming between caller and AI agent
- Call lifecycle management (create, hold, transfer, end)
- DTMF (touch-tone) input handling
- Call recording and transcription
- Call routing to appropriate AI agents
- Queue management and ACD (Automatic Call Distribution)

---

## 3. Aggregate Design

### Call Aggregate

```
Call (Aggregate Root)
├── CallMetadata
│   ├── CallId (UUID)
│   ├── TenantId
│   ├── Direction (Inbound | Outbound)
│   ├── Channel (PSTN | SIP | WebRTC | Mobile)
│   ├── CallerInfo
│   │   ├── PhoneNumber
│   │   ├── CallerName
│   │   └── CustomerId (if matched)
│   ├── CalleeInfo
│   │   ├── PhoneNumber
│   │   └── AgentId
│   └── Timestamps
│       ├── CreatedAt
│       ├── AnsweredAt
│       ├── EndedAt
│       └── Duration
├── CallState
│   ├── Status (Ringing | Connecting | Active | OnHold | Transferring | Completed | Failed | Abandoned)
│   ├── HoldCount
│   ├── TransferCount
│   └── HangupCause
├── AudioSession
│   ├── StreamId
│   ├── Codec (G711Ulaw | G711Alaw | Opus | GSM)
│   ├── SampleRate (8000 | 16000 | 48000)
│   ├── Direction (Bidirectional)
│   └── RecordingId
├── ConversationContext
│   ├── ConversationId
│   ├── AgentId
│   ├── Transcript[]
│   └── TokensUsed
└── BillingInfo
    ├── DurationSeconds
    ├── BilledDuration
    └── Cost
```

### Entities

| Entity | Attributes |
|---|---|
| Call | CallId, TenantId, Direction, Channel, CallerNumber, CalleeNumber, Status, AgentId, StartedAt, EndedAt, Duration |
| CallRecording | RecordingId, CallId, StoragePath, Format, Duration, Size |
| CallTranscript | TranscriptId, CallId, Speaker, Content, StartTime, EndTime, Confidence |
| CallTransfer | TransferId, CallId, FromAgent, ToAgent, Reason, Timestamp |
| DTMFEvent | EventId, CallId, Digit, Timestamp |
| PhoneNumber | PhoneId, TenantId, Number, Type (DID | TollFree | Mobile), Status |

### Value Objects

| Value Object | Type | Description |
|---|---|---|
| `CallId` | UUID | Unique call identifier |
| `PhoneNumber` | String | E.164 format phone number |
| `CallDirection` | Enum | Inbound, Outbound |
| `CallStatus` | Enum | Ringing, Connecting, Active, OnHold, Transferring, Completed, Failed, Abandoned |
| `AudioCodec` | Enum | G711Ulaw, G711Alaw, Opus, GSM |
| `HangupCause` | Enum | NormalClearing, NoAnswer, Busy, Failed, Congestion, Rejected |

---

## 4. Public API Trait

```rust
// nexus-telephony/src/interfaces/api.rs
#[async_trait]
pub trait TelephonyService: Send + Sync {
    // Inbound call management
    async fn handle_inbound_call(&self, req: InboundCallRequest) -> Result<CallResponse, TelephonyError>;
    async fn answer_call(&self, call_id: CallId) -> Result<(), TelephonyError>;
    async fn reject_call(&self, call_id: CallId, reason: String) -> Result<(), TelephonyError>;
    
    // Outbound call management
    async fn initiate_outbound_call(&self, req: OutboundCallRequest) -> Result<CallResponse, TelephonyError>;
    
    // Call control
    async fn hold_call(&self, call_id: CallId) -> Result<(), TelephonyError>;
    async fn resume_call(&self, call_id: CallId) -> Result<(), TelephonyError>;
    async fn transfer_call(&self, req: TransferRequest) -> Result<(), TelephonyError>;
    async fn end_call(&self, call_id: CallId, reason: String) -> Result<(), TelephonyError>;
    async fn bridge_calls(&self, call_id_1: CallId, call_id_2: CallId) -> Result<(), TelephonyError>;
    
    // Audio streaming
    async fn send_audio(&self, call_id: CallId, audio: AudioFrame) -> Result<(), TelephonyError>;
    async fn receive_audio(&self, call_id: CallId) -> Result<Receiver<AudioFrame>, TelephonyError>;
    
    // DTMF
    async fn send_dtmf(&self, call_id: CallId, digits: String) -> Result<(), TelephonyError>;
    async fn receive_dtmf(&self, call_id: CallId) -> Result<Receiver<DTMFEvent>, TelephonyError>;
    
    // Recording
    async fn start_recording(&self, call_id: CallId) -> Result<RecordingId, TelephonyError>;
    async fn stop_recording(&self, call_id: CallId) -> Result<RecordingId, TelephonyError>;
    
    // Query
    async fn get_call_status(&self, call_id: CallId) -> Result<CallStatus, TelephonyError>;
    async fn list_active_calls(&self, tenant_id: TenantId) -> Result<Vec<CallSummary>, TelephonyError>;
    async fn get_call_history(&self, query: CallHistoryQuery) -> Result<Vec<CallSummary>, TelephonyError>;
}

pub struct InboundCallRequest {
    pub caller_number: PhoneNumber,
    pub callee_number: PhoneNumber,
    pub channel: CallChannel,
    pub sip_headers: HashMap<String, String>,
    pub tenant_id: TenantId,
}

pub struct OutboundCallRequest {
    pub caller_number: PhoneNumber,
    pub callee_number: PhoneNumber,
    pub agent_id: Option<AgentId>,
    pub context: HashMap<String, String>,
    pub tenant_id: TenantId,
    pub user_id: UserId,
}

pub struct CallResponse {
    pub call_id: CallId,
    pub status: CallStatus,
    pub sip_call_id: Option<String>,
}

pub struct TransferRequest {
    pub call_id: CallId,
    pub target_agent_id: Option<AgentId>,
    pub target_phone: Option<PhoneNumber>,
    pub transfer_type: TransferType, // Blind | Attended
    pub reason: String,
}

pub struct AudioFrame {
    pub call_id: CallId,
    pub sequence: u64,
    pub timestamp: u64,
    pub codec: AudioCodec,
    pub data: Vec<u8>,
}

pub enum TransferType {
    Blind,      // Immediate transfer, caller doesn't speak to new agent first
    Attended,   // New agent is introduced before transfer completes
}
```

---

## 5. Call Processing Pipeline

### 5.1 Inbound Call Flow

```
Caller Dials AeroXe Number
    |
    v
[PSTN/SIP Provider] → Webhook/SIP INVITE
    |
    v
[1] Call Reception (nexus-telephony)
    |  - Parse SIP headers / Webhook payload
    |  - Extract caller/callee numbers
    |  - Create Call entity (status: Ringing)
    |
    v
[2] Number Routing
    |  - Match callee_number to tenant DID
    |  - Determine tenant and routing rules
    |
    v
[3] Customer Identification
    |  - Lookup caller_number in customer database
    |  - If matched: attach customer context
    |  - If unmatched: guest mode
    |
    v
[4] Agent Selection (via nexus-agent)
    |  - Determine intent from IVR / direct dial
    |  - Select appropriate AI agent
    |  - Load agent configuration
    |
    v
[5] Audio Session Setup
    |  - Establish RTP/WebRTC audio stream
    |  - Negotiate codec (G.711 / Opus)
    |  - Start bidirectional audio relay
    |
    v
[6] AI Conversation Loop
    |  - Audio → nexus-stt (transcription)
    |  - Transcription → nexus-agent (processing)
    |  - Agent response → nexus-tts (synthesis)
    |  - TTS audio → caller (playback)
    |
    v
[7] Call Conclusion
    |  - End call (normal / transfer / abandon)
    |  - Save recording
    |  - Save transcript
    |  - Publish audit event
    |  - Calculate billing
```

### 5.2 Outbound Call Flow

```
AI Decision / Campaign Trigger
    |
    v
[1] Initiate Call (nexus-telephony)
    |  - Validate caller/callee numbers
    |  - Check outbound permissions
    |  - Check DNC list
    |
    v
[2] Dial via Provider
    |  - SIP INVITE / API call to provider
    |  - Create Call entity (status: Connecting)
    |
    v
[3] Call Connected
    |  - Answer detection
    |  - Start audio stream
    |
    v
[4] AI Conversation
    |  - Same as inbound flow
    |
    v
[5] Call Conclusion
    |  - Same as inbound flow
```

---

## 6. Audio Streaming Architecture

### 6.1 Real-Time Audio Pipeline

```
Caller Audio (RTP/WebSocket)
    |
    v
[Audio Receiver] (nexus-telephony)
    |  - Decode RTP/G.711 → PCM
    |  - Jitter buffer (50ms)
    |  - Silence detection (VAD)
    |
    v
[Voice Activity Detection]
    |  - Detect speech start/end
    |  - Filter background noise
    |  - Segment speech utterances
    |
    v
[Utterance Complete]
    |  - "chunk of speech ready"
    |
    v
[Speech-to-Text] (nexus-stt)
    |  - Stream PCM to STT engine
    |  - Real-time transcription
    |  - Partial results for low latency
    |
    v
[Transcription Text]
    |
    v
[Agent Processing] (nexus-agent)
    |  - Process text as chat message
    |  - Generate response text
    |
    v
[Text-to-Speech] (nexus-tts)
    |  - Convert response to speech
    |  - Stream audio chunks
    |
    v
[Audio Sender] (nexus-telephony)
    |  - Encode PCM → G.711/Opus
    |  - Send RTP packets
    |
    v
Caller Hears Response
```

### 6.2 Latency Budget

| Segment | Target | Method |
|---|---|---|
| Audio capture → STT | < 50ms | Local PCM buffering |
| STT processing | < 200ms | Streaming STT (Whisper) |
| Agent processing | < 500ms | Optimized prompt + streaming |
| TTS generation | < 200ms | Streaming TTS |
| Audio playback | < 50ms | Pre-buffering |
| **Total round-trip** | **< 1000ms** | Optimized pipeline |

### 6.3 Audio Frame Format

```rust
pub struct AudioFrame {
    pub call_id: CallId,
    pub sequence: u64,          // Monotonic sequence number
    pub timestamp: u64,         // RTP timestamp (8kHz or 16kHz)
    pub codec: AudioCodec,      // G.711, Opus, etc.
    pub data: Vec<u8>,          // Raw audio bytes
    pub is_speech: bool,        // VAD result
    pub is_final: bool,         // Final frame of utterance
}

pub enum AudioCodec {
    G711Ulaw,   // 8kHz, 64kbps (PSTN standard)
    G711Alaw,   // 8kHz, 64kbps (European PSTN)
    Opus,       // 8-48kHz, variable bitrate (WebRTC)
    GSM,        // 8kHz, 13kbps (mobile)
    PCM,        // 16kHz, 128kbps (internal)
}
```

---

## 7. Call Routing

### 7.1 Routing Rules

| Rule | Description |
|---|---|
| DID-based routing | Match inbound number to tenant |
| IVR routing | Menu-based agent selection |
| Skill-based routing | Route to agent with matching capability |
| Queue routing | Round-robin / least-busy / priority |
| Time-based routing | Business hours vs after-hours |
| Geographic routing | Route based on caller location |

### 7.2 Agent Selection for Calls

```
Incoming Call
    |
    v
Determine Intent
    |  - DTMF menu selection
    |  - First utterance analysis
    |  - Caller history lookup
    |
    v
Select Agent
    |  - Match intent to agent capabilities
    |  - Check agent availability
    |  - Load agent voice/personality config
    |
    v
Assign Agent
    |  - Bind agent to call session
    |  - Initialize conversation context
    |  - Start audio relay
```

### 7.3 Queue Management

```rust
pub struct CallQueue {
    pub queue_id: QueueId,
    pub tenant_id: TenantId,
    pub name: String,
    pub max_wait_seconds: u32,
    pub max_queue_size: u32,
    pub overflow_action: OverflowAction, // Voicemail | Transfer | Hangup
    pub music_on_hold: Option<String>,   // Audio file path
}

pub enum OverflowAction {
    Voicemail,
    TransferToNumber(PhoneNumber),
    Hangup,
    PlayMessage(String),
}
```

---

## 8. Call Transfer

### 8.1 Transfer Types

| Type | Description | Flow |
|---|---|---|
| **Blind Transfer** | Immediate transfer without introduction | AI → New Agent (caller hears ring) |
| **Attended Transfer** | AI introduces new agent first | AI → New Agent (introduction) → Transfer |
| **Warm Transfer** | AI speaks to new agent before connecting | AI → Agent私下交谈 → Connect caller |
| **Conference** | Three-way conversation | AI + Caller + New Agent |
| **Queue Transfer** | Transfer to human queue | AI → Human Queue → Available Agent |

### 8.2 Transfer Flow

```
AI Agent decides to transfer
    |
    v
[1] Agent calls TelephonyService::transfer_call()
    |
    v
[2] Check transfer permissions
    |  - Agent allowed to transfer?
    |  - Target agent available?
    |
    v
[3] Put caller on hold
    |  - Play hold music/message
    |  - Maintain audio session
    |
    v
[4] Connect to target
    |  - Blind: Direct connect
    |  - Attended: Ring target, wait for answer
    |
    v
[5] Complete transfer
    |  - Bridge audio streams
    |  - Update call record
    |  - Transfer context to new agent
    |
    v
[6] Audit event published
```

---

## 9. DTMF Handling

```rust
pub struct DTMFConfig {
    pub enabled: bool,
    pub timeout_ms: u32,              // Max wait between digits
    pub max_digits: u32,              // Max digits to collect
    pub termination_digit: char,      // Digit that ends input (#)
    pub prompt_text: Option<String>,  // TTS prompt before collection
}

pub struct DTMFMenu {
    pub menu_id: String,
    pub prompt: String,               // "Press 1 for sales, press 2 for support"
    pub options: Vec<DTMFOption>,
    pub timeout_action: TimeoutAction,
}

pub struct DTMFOption {
    pub digit: char,
    pub action: DTMFAction,
    pub label: String,
}

pub enum DTMFAction {
    TransferToAgent(AgentId),
    TransferToQueue(QueueId),
    PlayMessage(String),
    RunWorkflow(WorkflowId),
    Hangup,
}
```

---

## 10. Call Recording

### 10.1 Recording Configuration

| Setting | Description |
|---|---|
| Always record | Record all calls |
| On-demand | Agent starts/stops recording |
| Consent-based | Record only after consent prompt |
| Dual-channel | Separate recordings for caller and agent |
| Format | WAV (uncompressed) or MP3 (compressed) |
| Storage | MinIO (aeroxe-call-recordings bucket) |
| Retention | Configurable per tenant |

### 10.2 Recording Storage

```
Call Recording
    |
    v
[1] Capture audio stream (both channels)
    |
    v
[2] Write to temp file (WAV)
    |
    v
[3] On call end:
    |  - Convert to MP3 (configurable)
    |  - Upload to MinIO
    |  - Store metadata in PostgreSQL
    |
    v
[4] Background job:
    |  - Transcribe recording (STT)
    |  - Store transcript
    |  - Generate call summary (AI)
```

---

## 11. Database Schema (telephony_ schema)

### calls

```sql
CREATE TABLE telephony.calls (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    call_id UUID NOT NULL UNIQUE,
    direction VARCHAR(10) NOT NULL,           -- inbound | outbound
    channel VARCHAR(20) NOT NULL,             -- pstn | sip | webrtc | mobile
    caller_number VARCHAR(20) NOT NULL,
    callee_number VARCHAR(20) NOT NULL,
    customer_id BIGINT,                       -- matched customer
    agent_id BIGINT,                          -- assigned AI agent
    status VARCHAR(20) NOT NULL DEFAULT 'ringing',
    sip_call_id VARCHAR(100),
    queue_id BIGINT,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    answered_at TIMESTAMP,
    ended_at TIMESTAMP,
    duration_seconds INT,
    hangup_cause VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_calls_tenant ON telephony.calls(tenant_id, created_at DESC);
CREATE INDEX idx_calls_customer ON telephony.calls(customer_id, created_at DESC);
CREATE INDEX idx_calls_status ON telephony.calls(status) WHERE status NOT IN ('completed', 'failed');
```

### call_recordings

```sql
CREATE TABLE telephony.recordings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony.calls(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    format VARCHAR(10) NOT NULL,              -- wav | mp3
    duration_seconds INT,
    file_size_bytes BIGINT,
    channels INT NOT NULL DEFAULT 2,          -- 1=mono, 2=stereo
    sample_rate INT NOT NULL DEFAULT 8000,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### call_transcripts

```sql
CREATE TABLE telephony.transcripts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony.calls(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    speaker VARCHAR(20) NOT NULL,             -- caller | agent | system
    content TEXT NOT NULL,
    start_time_ms INT,
    end_time_ms INT,
    confidence FLOAT,
    model VARCHAR(100),                       -- STT model used
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transcripts_call ON telephony.transcripts(call_id);
```

### phone_numbers

```sql
CREATE TABLE telephony.phone_numbers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    number VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,                -- did | tollfree | mobile | sip
    provider VARCHAR(50),
    provider_config JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_phone_unique ON telephony.phone_numbers(number);
```

### call_queues

```sql
CREATE TABLE telephony.queues (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    max_wait_seconds INT NOT NULL DEFAULT 300,
    max_queue_size INT NOT NULL DEFAULT 50,
    overflow_action VARCHAR(50) NOT NULL DEFAULT 'voicemail',
    overflow_target VARCHAR(100),
    music_on_hold_path TEXT,
    priority INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### dtmf_events

```sql
CREATE TABLE telephony.dtmf_events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony.calls(id) ON DELETE CASCADE,
    digit CHAR(1) NOT NULL,
    timestamp_ms INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 12. REST API Endpoints

### Inbound Call Webhook

```
POST /api/v1/telephony/webhook/inbound
Content-Type: application/json
```

**Request (from telephony provider):**
```json
{
  "call_id": "uuid",
  "caller_number": "+919876543210",
  "callee_number": "+911800123456",
  "channel": "pstn",
  "sip_headers": {
    "From": "<sip:+919876543210@aeroxe.com>",
    "To": "<sip:+911800123456@aeroxe.com>"
  }
}
```

### Initiate Outbound Call

```
POST /api/v1/telephony/calls/outbound
Authorization: Bearer <jwt>
```

**Request:**
```json
{
  "callee_number": "+919876543210",
  "agent_id": "support-agent",
  "context": {
    "reason": "follow_up",
    "ticket_id": "tkt_123"
  }
}
```

### Call Control

```
POST /api/v1/telephony/calls/{call_id}/hold
POST /api/v1/telephony/calls/{call_id}/resume
POST /api/v1/telephony/calls/{call_id}/transfer
POST /api/v1/telephony/calls/{call_id}/end
POST /api/v1/telephony/calls/{call_id}/recording/start
POST /api/v1/telephony/calls/{call_id}/recording/stop
```

### Call Query

```
GET /api/v1/telephony/calls/{call_id}
GET /api/v1/telephony/calls?status=active&tenant_id=1
GET /api/v1/telephony/calls/{call_id}/transcript
GET /api/v1/telephony/calls/{call_id}/recording
```

### WebSocket Audio Stream

```
wss://host/ws/v1/telephony/{call_id}
```

**Binary frames:** Audio data (G.711/Opus)
**Text frames:** Control messages (DTMF, status)

---

## 13. NATS Events

### Published

| Subject | Event |
|---|---|
| `aeroxe.v1.telephony.call.inbound` | `InboundCallReceived` |
| `aeroxe.v1.telephony.call.outbound` | `OutboundCallInitiated` |
| `aeroxe.v1.telephony.call.answered` | `CallAnswered` |
| `aeroxe.v1.telephony.call.ended` | `CallEnded` |
| `aeroxe.v1.telephony.call.transferred` | `CallTransferred` |
| `aeroxe.v1.telephony.call.hold` | `CallOnHold` |
| `aeroxe.v1.telephony.call.recording.started` | `RecordingStarted` |
| `aeroxe.v1.telephony.call.recording.stopped` | `RecordingStopped` |
| `aeroxe.v1.telephony.call.dtmf` | `DTMFReceived` |
| `aeroxe.v1.telephony.call.transcript.ready` | `TranscriptReady` |

### Subscribed

| Subject | Handler |
|---|---|
| `aeroxe.v1.agent.call.started` | Attach agent to call |
| `aeroxe.v1.agent.call.ended` | End call session |
| `aeroxe.v1.conversation.transfer` | Handle transfer request |

---

## 14. Provider Integration

### Supported Providers

| Provider | Protocol | Features |
|---|---|---|
| Twilio | SIP/Webhook | PSTN, Recording, STT, TTS |
| Vonage (Nexmo) | SIP/Webhook | PSTN, Recording |
| Bland.ai | API | AI-native calling |
| Retell.ai | API | AI voice agents |
| FreeSWITCH | SIP | Self-hosted PBX |
| Asterisk | SIP | Self-hosted PBX |
| Custom SIP | SIP/RTP | Any SIP provider |

### Provider Adapter Pattern

```rust
#[async_trait]
pub trait TelephonyProvider: Send + Sync {
    async fn make_call(&self, req: MakeCallRequest) -> Result<ProviderCallId, ProviderError>;
    async fn answer_call(&self, provider_call_id: &str) -> Result<(), ProviderError>;
    async fn hangup_call(&self, provider_call_id: &str) -> Result<(), ProviderError>;
    async fn send_dtmf(&self, provider_call_id: &str, digits: &str) -> Result<(), ProviderError>;
    async fn get_call_status(&self, provider_call_id: &str) -> Result<CallStatus, ProviderError>;
    fn get_audio_stream(&self, call_id: &CallId) -> Result<AudioStream, ProviderError>;
}

// Implementations:
pub struct TwilioProvider { ... }
pub struct VonageProvider { ... }
pub struct FreeSWITCHProvider { ... }
pub struct CustomSipProvider { ... }
```

---

## 15. Observability

### Metrics

| Metric | Type | Description |
|---|---|---|
| `telephony_calls_total` | Counter | Total calls by direction/status |
| `telephony_call_duration_seconds` | Histogram | Call duration |
| `telephony_calls_active` | Gauge | Active concurrent calls |
| `telephony_queue_size` | Gauge | Calls waiting in queue |
| `telephony_queue_wait_seconds` | Histogram | Time in queue |
| `telephony_transfer_total` | Counter | Transfers by type |
| `telephony_recording_total` | Counter | Recordings created |
| `telephony_stt_latency_ms` | Histogram | Speech-to-text latency |
| `telephony_tts_latency_ms` | Histogram | Text-to-speech latency |
| `telephony_audio_latency_ms` | Histogram | End-to-end audio latency |
| `telephony_dtmf_events_total` | Counter | DTMF digits received |

### Grafana Dashboard

| Panel | Description |
|---|---|
| Active Calls | Real-time call count |
| Call Volume | Calls per hour/day |
| Queue Depth | Calls waiting |
| Average Wait Time | Queue performance |
| Average Handle Time | Agent efficiency |
| Transfer Rate | Escalation tracking |
| Call Success Rate | Completion percentage |
| Recording Storage | Storage usage |

---

## 16. Performance Targets

| Operation | Target |
|---|---|
| Inbound call setup | < 100ms |
| Outbound call connect | < 2s |
| Audio latency (one-way) | < 100ms |
| STT latency | < 200ms |
| TTS latency | < 200ms |
| AI response latency | < 500ms |
| Total round-trip | < 1s |
| DTMF detection | < 50ms |
| Recording start | < 1s |

---

## 17. Security

| Requirement | Implementation |
|---|---|
| Call encryption | SRTP for RTP, TLS for SIP |
| Recording consent | Configurable per-tenant (required/optional) |
| PII in recordings | Auto-redact credit cards, SSN |
| Access control | Only authorized users can listen to recordings |
| DNC compliance | Check outbound against DNC list |
| Recording retention | Configurable, auto-delete after period |
| Phone number validation | E.164 format enforcement |
| Rate limiting | Per-tenant concurrent call limits |
