# AeroXe Nexus AI — Telephony Module

## Voice Channel, SIP/WebRTC Integration, Call Management & Real-Time Audio

> **Modular Monolith Module:** This document describes the `nexus-telephony` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via gRPC (sync) or NATS (async) between services.

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
CREATE TABLE telephony_.calls (
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

CREATE INDEX idx_calls_tenant ON telephony_.calls(tenant_id, created_at DESC);
CREATE INDEX idx_calls_customer ON telephony_.calls(customer_id, created_at DESC);
CREATE INDEX idx_calls_status ON telephony_.calls(status) WHERE status NOT IN ('completed', 'failed');
```

### call_recordings

```sql
CREATE TABLE telephony_.recordings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony_.calls(id) ON DELETE CASCADE,
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
CREATE TABLE telephony_.transcripts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony_.calls(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    speaker VARCHAR(20) NOT NULL,             -- caller | agent | system
    content TEXT NOT NULL,
    start_time_ms INT,
    end_time_ms INT,
    confidence FLOAT,
    model VARCHAR(100),                       -- STT model used
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transcripts_call ON telephony_.transcripts(call_id);
```

### phone_numbers

```sql
CREATE TABLE telephony_.phone_numbers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    number VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,                -- did | tollfree | mobile | sip
    provider VARCHAR(50),
    provider_config JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_phone_unique ON telephony_.phone_numbers(number);
```

### call_queues

```sql
CREATE TABLE telephony_.queues (
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
CREATE TABLE telephony_.dtmf_events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony_.calls(id) ON DELETE CASCADE,
    digit CHAR(1) NOT NULL,
    timestamp_ms INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 12. REST API Endpoints (Protobuf over HTTP)

| Method | Endpoint | Business Status | HTTP | Description |
|---|---|---|---|---|
| `POST` | `/api/v1/telephony/webhook/inbound` | `SUCCESS` | `200` | Inbound call webhook |
| `POST` | `/api/v1/telephony/calls/outbound` | `CREATED` | `201` | Initiate outbound call |
| `POST` | `/api/v1/telephony/calls` (read body) | `SUCCESS` | `200` | Get call details |
| `POST` | `/api/v1/telephony/calls/list` (read body) | `SUCCESS` | `200` | List calls |
| `POST` | `/api/v1/telephony/calls/{call_id}/hold` | `UPDATED` | `200` | Hold call |
| `POST` | `/api/v1/telephony/calls/{call_id}/resume` | `UPDATED` | `200` | Resume call |
| `POST` | `/api/v1/telephony/calls/{call_id}/transfer` | `UPDATED` | `200` | Transfer call |
| `POST` | `/api/v1/telephony/calls/{call_id}/end` | `UPDATED` | `200` | End call |
| `POST` | `/api/v1/telephony/calls/{call_id}/recording/start` | `UPDATED` | `200` | Start recording |
| `POST` | `/api/v1/telephony/calls/{call_id}/recording/stop` | `UPDATED` | `200` | Stop recording |
| `POST` | `/api/v1/telephony/calls/{call_id}/transcript` (read body) | `SUCCESS` | `200` | Get transcript |
| `POST` | `/api/v1/telephony/calls/{call_id}/auth/verify-pin` | `SUCCESS` | `200` | Verify PIN |
| `POST` | `/api/v1/telephony/calls/{call_id}/auth/verify-voice` | `SUCCESS` | `200` | Verify voice |
| `POST` | `/api/v1/telephony/calls/{call_id}/auth/status` (read body) | `SUCCESS` | `200` | Auth status |
| `POST` | `/api/v1/telephony/voicemails/list` (read body) | `SUCCESS` | `200` | List voicemails |
| `POST` | `/api/v1/telephony/voicemails` (read body) | `SUCCESS` | `200` | Get voicemail |
| `POST` | `/api/v1/telephony/voicemails/{id}/handle` | `UPDATED` | `200` | Handle voicemail |
| `POST` | `/api/v1/telephony/ivr-flows` | `CREATED` | `201` | Create IVR flow |
| `POST` | `/api/v1/telephony/ivr-flows/list` (read body) | `SUCCESS` | `200` | List IVR flows |
| `PATCH` | `/api/v1/telephony/ivr-flows/{id}` | `UPDATED` | `200` | Update IVR flow |
| `DELETE` | `/api/v1/telephony/ivr-flows/{id}` | `DELETED` | `204` | Delete IVR flow |
| `POST` | `/api/v1/telephony/calls/{call_id}/monitor/listen` | `SUCCESS` | `200` | Listen-in |
| `POST` | `/api/v1/telephony/calls/{call_id}/monitor/whisper` | `SUCCESS` | `200` | Whisper to agent |
| `POST` | `/api/v1/telephony/calls/{call_id}/monitor/barge-in` | `SUCCESS` | `200` | Barge-in |
| `POST` | `/api/v1/telephony/calls/{call_id}/monitor/stop` | `UPDATED` | `200` | Stop monitoring |
| `WS` | `wss://host/ws/v1/telephony/{call_id}` | — | — | Audio stream |
| `WS` | `wss://host/ws/v1/telephony/monitor/{call_id}` | — | — | Live monitoring |

### List Calls Response

```json
{
  "status": "SUCCESS",
  "data": [...],
  "summary": {
    "total_items": 5000,
    "active_calls": 12,
    "completed_calls": 4900,
    "failed_calls": 88,
    "abandoned_calls": 256,
    "avg_duration_seconds": 245,
    "avg_wait_seconds": 32,
    "avg_handle_time_seconds": 180,
    "recent_activity": {
      "calls_today": 150,
      "calls_this_hour": 15
    }
  },
  "pagination": {"total": 5000, "limit": 10, "offset": 0, "has_more": true},
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

### List Voicemails Response

```json
{
  "status": "SUCCESS",
  "data": [...],
  "summary": {
    "total_items": 150,
    "new_voicemails": 8,
    "listened_voicemails": 100,
    "handled_voicemails": 42,
    "avg_duration_seconds": 45,
    "recent_activity": {
      "new_today": 3,
      "handled_today": 5
    }
  },
  "pagination": {"total": 150, "limit": 10, "offset": 0, "has_more": true},
  "meta": {"request_id": "uuid", "timestamp": "2026-07-21T12:00:00Z"}
}
```

**Note:** All endpoints use Protobuf (proto3) serialized as JSON over HTTP. Content-Type: `application/json` (Protobuf messages serialized as JSON). No PUT method — use PATCH for updates. Read operations use POST with a read body. All list endpoints support `limit` (default 10) and `offset`.

---

## 13. NATS Events

### Published

| Subject | Event |
|---|---|
| `aeroxe.v1.telephony_.call.inbound` | `InboundCallReceived` |
| `aeroxe.v1.telephony_.call.outbound` | `OutboundCallInitiated` |
| `aeroxe.v1.telephony_.call.answered` | `CallAnswered` |
| `aeroxe.v1.telephony_.call.ended` | `CallEnded` |
| `aeroxe.v1.telephony_.call.transferred` | `CallTransferred` |
| `aeroxe.v1.telephony_.call.hold` | `CallOnHold` |
| `aeroxe.v1.telephony_.call.recording.started` | `RecordingStarted` |
| `aeroxe.v1.telephony_.call.recording.stopped` | `RecordingStopped` |
| `aeroxe.v1.telephony_.call.dtmf` | `DTMFReceived` |
| `aeroxe.v1.telephony_.call.transcript.ready` | `TranscriptReady` |

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

## 17. Caller Authentication (NEW - CRITICAL)

### 17.1 Authentication Methods

| Method | Security Level | Use Case |
|---|---|---|
| **Caller ID Match** | Low | Match number to customer record |
| **PIN Verification** | Medium | 4-6 digit PIN entry |
| **Voice Biometrics** | High | Speaker verification |
| **Knowledge-Based** | Medium | Security questions |
| **Multi-Factor** | High | Number + PIN + Voice |

### 17.2 Caller Authentication Flow

```
Inbound Call Received
    |
    v
[1] Extract Caller Number (ANI/CLI)
    |
    v
[2] Lookup Customer Record
    |  - Match phone number to customer
    |  - Check verification level required
    |
    v
[3] Determine Auth Level Required
    |  - General inquiry: No auth needed
    |  - Account info: PIN required
    |  - Sensitive data: Voice biometrics + PIN
    |
    v
[4] Perform Authentication
    |  - Play PIN prompt: "Please enter your 4-digit PIN"
    |  - Collect DTMF digits
    |  - Validate PIN against customer record
    |  - Optional: Voice sample comparison
    |
    v
[5] Auth Result
    |  - Success: Attach verified_customer context
    |  - Failure: Limit data access, offer fallback
    |  - Max attempts (3): Transfer to human
```

### 17.3 Caller Authentication Entities

```sql
CREATE TABLE telephony_.caller_auth (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony_.calls(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT,
    auth_method VARCHAR(30) NOT NULL,       -- caller_id | pin | voice_biometric | knowledge
    auth_status VARCHAR(20) NOT NULL,       -- pending | success | failed | skipped
    attempt_count INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    verified_at TIMESTAMP,
    failure_reason VARCHAR(100),
    voice_biometric_score FLOAT,            -- 0.0-1.0 similarity score
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 18. Anti-Fraud & Anti-Spoofing (NEW - CRITICAL)

### 18.1 Fraud Detection Rules

| Threat | Detection Method | Action |
|---|---|---|
| Caller ID Spoofing | ANI validation, carrier lookup | Flag + require additional auth |
| SIM Swap | Recent number change detection | Block + alert |
| Voice Clone | Liveness detection, deepfake detection | Block + alert |
| Replay Attack | Audio nonce verification | Reject audio |
| Toll Fraud | Outbound call pattern analysis | Block + alert |
| Brute Force PIN | Failed attempt rate limiting | Lock account |
| Social Engineering | Conversation pattern analysis | Alert supervisor |

### 18.2 Fraud Detection Pipeline

```
Call Received
    |
    v
[1] Pre-Call Fraud Check
    |  - Caller reputation score (external service)
    |  - Number validity check (carrier API)
    |  - Recent fraud patterns
    |
    v
[2] During-Call Monitoring
    |  - PIN attempt counting
    |  - Voice biometric confidence
    |  - Conversation anomaly detection
    |
    v
[3] Post-Call Analysis
    |  - Call pattern analysis
    |  - Fraud scoring
    |  - Alert if suspicious
```

### 18.3 Fraud Entities

```sql
CREATE TABLE telephony_.fraud_checks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony_.calls(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    check_type VARCHAR(50) NOT NULL,       -- caller_reputation | number_validity | sim_swap | deepfake
    result VARCHAR(20) NOT NULL,           -- pass | flag | block
    score FLOAT,                           -- 0.0-1.0 risk score
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 19. Audio Quality Monitoring (NEW)

### 19.1 Quality Metrics

| Metric | Target | Alert Threshold |
|---|---|---|
| MOS (Mean Opinion Score) | > 4.0 | < 3.0 |
| Jitter | < 30ms | > 50ms |
| Packet Loss | < 1% | > 3% |
| Round-trip Latency | < 150ms | > 300ms |
| Echo | None | Detected |
| Noise Floor | < -40dB | > -30dB |

### 19.2 Quality Monitoring Pipeline

```
RTP Audio Stream
    |
    v
[1] Quality Analysis (per 5-second window)
    |  - Calculate MOS
    |  - Measure jitter
    |  - Detect packet loss
    |
    v
[2] Alert Evaluation
    |  - Compare to thresholds
    |  - Check degradation trend
    |
    v
[3] Actions
    |  - Degrading: Log warning
    |  - Critical: Alert supervisor
    |  - Unusable: Offer callback
```

### 19.3 Quality Entities

```sql
CREATE TABLE telephony_.audio_quality (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony_.calls(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    window_start_ms INT NOT NULL,
    window_end_ms INT NOT NULL,
    mos_score FLOAT,
    jitter_ms FLOAT,
    packet_loss_percent FLOAT,
    latency_ms FLOAT,
    noise_floor_db FLOAT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 20. Voicemail System (NEW)

### 20.1 Voicemail Flow

```
Call Not Answered / Agent Unavailable
    |
    v
[1] Voicemail Prompt
    |  - "Please leave a message after the tone"
    |  - Play tone
    |
    v
[2] Record Voicemail
    |  - Record caller audio
    |  - Max duration: 3 minutes
    |  - Store in MinIO
    |
    v
[3] Post-Processing
    |  - Transcribe voicemail (STT)
    |  - Detect caller number → customer match
    |  - Generate summary (AI)
    |
    v
[4] Notification
    |  - Notify assigned agent
    |  - Send email with transcript
    |  - Add to agent queue
    |
    v
[5] Agent Action
    |  - Listen to voicemail
    |  - Read transcript
    |  - Call back customer
    |  - Mark as handled
```

### 20.2 Voicemail Entities

```sql
CREATE TABLE telephony_.voicemails (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    voicemail_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    call_id BIGINT REFERENCES telephony_.calls(id),
    customer_id BIGINT,
    caller_number VARCHAR(20) NOT NULL,
    agent_id BIGINT,
    storage_path TEXT NOT NULL,
    duration_seconds INT,
    transcription TEXT,
    summary TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'new',  -- new | listened | handled | archived
    handled_by BIGINT,
    handled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 21. IVR System (NEW)

### 21.1 IVR Flow

```
Call Answered
    |
    v
[1] Play Welcome Message
    |  - "Welcome to AeroXe. Press 1 for support, press 2 for billing..."
    |
    v
[2] Collect Input
    |  - DTMF digit
    |  - Or speech: "Say sales or support"
    |
    v
[3] Route Based on Input
    |  - 1 → Support queue
    |  - 2 → Billing queue
    |  - 3 → Technical support
    |  - 0 → Operator
    |
    v
[4] Nested Menus
    |  - Sub-menus for department
    |  - Agent selection
    |
    v
[5] Fallback
    |  - Invalid input → repeat menu
    |  - No input → repeat menu
    |  - Max 3 repeats → transfer to operator
```

### 21.2 IVR Configuration

```rust
pub struct IVRFlow {
    pub flow_id: FlowId,
    pub tenant_id: TenantId,
    pub name: String,
    pub nodes: Vec<IVRNode>,
    pub default_timeout_ms: u32,
    pub max_retries: u32,
}

pub struct IVRNode {
    pub node_id: String,
    pub node_type: IVRNodeType,
    pub prompt: String,
    pub options: Vec<IVROption>,
}

pub enum IVRNodeType {
    Menu,           // Press 1 for X, press 2 for Y
    PlayMessage,    // Play audio message
    CollectDigits,  // Collect DTMF input
    SpeechRecognize, // Voice command
    Transfer,       // Transfer to agent/queue
    Voicemail,      // Route to voicemail
    Hangup,         // End call
}

pub struct IVROption {
    pub digit: Option<char>,           // DTMF option
    pub speech_text: Option<String>,   // Voice command text
    pub next_node: String,             // Next IVR node
    pub action: IVRAction,
}

pub enum IVRAction {
    TransferToQueue(QueueId),
    TransferToAgent(AgentId),
    PlayMessage(String),
    CollectDigits(u32), // max digits
    Voicemail,
    Hangup,
}
```

### 21.3 IVR Entities

```sql
CREATE TABLE telephony_.ivr_flows (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    flow_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    definition JSONB NOT NULL,
    default_timeout_ms INT NOT NULL DEFAULT 5000,
    max_retries INT NOT NULL DEFAULT 3,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 22. Real-Time Call Monitoring (NEW)

### 22.1 Supervisor Capabilities

| Capability | Description |
|---|---|
| **Listen-In** | Hear call without being detected |
| **Whisper** | Coach agent privately (caller can't hear) |
| **Barge-In** | Join call as third party |
| **Transfer** | Force transfer to human |
| **End Call** | Terminate call |
| **View Transcript** | See live transcription |

### 22.2 Monitoring Flow

```
Supervisor Monitors Active Call
    |
    v
[1] WebSocket Connection
    |  - Supervisor connects to monitoring stream
    |  - Authenticate with supervisor role
    |
    v
[2] Live Data Stream
    |  - Real-time transcription
    |  - Agent/customer audio (listen-in)
    |  - Sentiment indicators
    |  - Call metadata
    |
    v
[3] Supervisor Actions
    |  - Whisper: Private channel to agent
    |  - Barge-in: Join conversation
    |  - Transfer: Force escalation
```

### 22.3 Monitoring Entities

```sql
CREATE TABLE telephony_.call_monitoring (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony_.calls(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    supervisor_id BIGINT NOT NULL,
    action VARCHAR(20) NOT NULL,          -- listen | whisper | barge_in
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP
);
```

---

## 23. Security (Updated)

| Requirement | Implementation |
|---|---|
| **Caller authentication** | **Multi-factor: caller ID + PIN + voice biometrics** |
| **Anti-fraud** | **Caller reputation, SIM swap detection, deepfake detection** |
| **Audio injection prevention** | **Liveness detection, nonce verification** |
| Call encryption | SRTP for RTP, TLS for SIP |
| **Recording consent** | **Mandatory consent prompt, audit trail, two-party support** |
| PII in recordings | Auto-redact credit cards, SSN |
| Access control | RBAC on recordings, playback audit |
| DNC compliance | Check outbound against DNC list |
| Recording retention | Configurable, auto-delete after period |
| Phone number validation | E.164 format enforcement |
| Rate limiting | Per-tenant concurrent call limits |
| **Audio quality monitoring** | **MOS, jitter, packet loss tracking** |
| **Voicemail security** | **Encrypted storage, access-controlled playback** |
| **IVR security** | **Input validation, rate limiting per menu** |
