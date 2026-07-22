# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 13 — Telephony & Voice Architecture

## SIP/WebRTC Integration, Call Management, Caller Authentication, Real-Time Audio

---

## 1. Telephony Module Overview

The Telephony module is the **voice channel gateway** for AeroXe Nexus AI. It enables:

- Receiving inbound voice calls (SIP, WebRTC, PSTN)
- Making outbound voice calls
- Real-time audio streaming between caller and AI agent
- Call lifecycle management (create, hold, transfer, end)
- DTMF (touch-tone) input handling
- Call recording and transcription
- Call routing to appropriate AI agents
- Queue management and ACD (Automatic Call Distribution)
- Caller authentication (PIN, voice biometrics)
- Anti-fraud and anti-spoofing
- Voicemail system
- IVR (Interactive Voice Response)
- Real-time call monitoring (listen-in, whisper, barge-in)
- Audio quality monitoring (MOS, jitter, packet loss)

---

## 2. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `telephony` |
| Bounded Context | Telephony |
| Domain Type | Core Domain |
| Schema | `telephony_` (shared PostgreSQL) |
| Dependencies | `nexus-agent`, `nexus-stt`, `nexus-tts`, `nexus-conversation`, `nexus-memory`, `nexus-audit` |

---

## 3. Aggregate Design

### Call Aggregate

```
Call (Aggregate Root)
├── CallMetadata (CallId, TenantId, Direction, Channel, CallerInfo, CalleeInfo)
├── CallState (Status, HoldCount, TransferCount, HangupCause)
├── AudioSession (StreamId, Codec, SampleRate, RecordingId)
├── ConversationContext (ConversationId, AgentId, Transcript)
├── CallerAuth (AuthMethod, AuthStatus, Attempts)
├── FraudChecks (CheckType, Result, Score)
└── BillingInfo (DurationSeconds, Cost)
```

### Call States

```
Ringing → Connecting → Active → OnHold → Transferring → Completed
                  ↘              ↘
                 Failed        Abandoned
```

---

## 4. Caller Authentication (CRITICAL)

### 4.1 Authentication Methods

| Method | Security Level | Use Case |
|---|---|---|
| Caller ID Match | Low | Match number to customer record |
| PIN Verification | Medium | 4-6 digit PIN entry |
| Voice Biometrics | High | Speaker verification |
| Knowledge-Based | Medium | Security questions |
| Multi-Factor | High | Number + PIN + Voice |

### 4.2 Authentication Flow

```
Inbound Call → Extract Caller Number → Lookup Customer
  → Determine Auth Level → Perform Authentication → Auth Result
```

### 4.3 Entities

```sql
CREATE TABLE telephony.caller_auth (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony.calls(id),
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT,
    auth_method VARCHAR(30) NOT NULL,
    auth_status VARCHAR(20) NOT NULL,
    attempt_count INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    verified_at TIMESTAMP,
    failure_reason VARCHAR(100),
    voice_biometric_score FLOAT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 5. Anti-Fraud & Anti-Spoofing (CRITICAL)

### 5.1 Fraud Detection Rules

| Threat | Detection | Action |
|---|---|---|
| Caller ID Spoofing | ANI validation, carrier lookup | Flag + require additional auth |
| SIM Swap | Recent number change detection | Block + alert |
| Voice Clone | Liveness detection, deepfake detection | Block + alert |
| Replay Attack | Audio nonce verification | Reject audio |
| Toll Fraud | Outbound call pattern analysis | Block + alert |
| Brute Force PIN | Failed attempt rate limiting | Lock account |

### 5.2 Entities

```sql
CREATE TABLE telephony.fraud_checks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES telephony.calls(id),
    tenant_id BIGINT NOT NULL,
    check_type VARCHAR(50) NOT NULL,
    result VARCHAR(20) NOT NULL,
    score FLOAT,
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 6. Audio Quality Monitoring

### 6.1 Metrics

| Metric | Target | Alert Threshold |
|---|---|---|
| MOS (Mean Opinion Score) | > 4.0 | < 3.0 |
| Jitter | < 30ms | > 50ms |
| Packet Loss | < 1% | > 3% |
| Round-trip Latency | < 150ms | > 300ms |

---

## 7. Voicemail System

### 7.1 Flow

```
Call Not Answered → Play Voicemail Prompt → Record Voicemail
  → Transcribe (STT) → Generate Summary (AI) → Notify Agent
  → Agent Handles → Mark as Handled
```

### 7.2 Entities

```sql
CREATE TABLE telephony.voicemails (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    voicemail_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    call_id BIGINT,
    customer_id BIGINT,
    caller_number VARCHAR(20) NOT NULL,
    agent_id BIGINT,
    storage_path TEXT NOT NULL,
    duration_seconds INT,
    transcription TEXT,
    summary TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    handled_by BIGINT,
    handled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 8. IVR System

### 8.1 IVR Flow

```
Call Answered → Play Welcome Message → Collect Input (DTMF/Speech)
  → Route Based on Input → Nested Menus → Fallback
```

### 8.2 IVR Configuration

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
    Menu, PlayMessage, CollectDigits, SpeechRecognize, Transfer, Voicemail, Hangup,
}
```

---

## 9. Real-Time Call Monitoring

### 9.1 Supervisor Capabilities

| Capability | Description |
|---|---|
| Listen-In | Hear call without being detected |
| Whisper | Coach agent privately (caller can't hear) |
| Barge-In | Join call as third party |
| Transfer | Force transfer to human |
| End Call | Terminate call |

---

## 10. REST API Endpoints

| Method | Endpoint | Business Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/telephony/webhook/inbound` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/outbound` | `CREATED` | `201` |
| `POST` | `/api/v1/telephony/calls/{call_id}` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/hold` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/resume` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/transfer` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/end` | `UPDATED` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/auth/verify-pin` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/calls/{call_id}/auth/verify-voice` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/voicemails?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/telephony/ivr-flows` | `CREATED` | `201` |
| `PATCH` | `/api/v1/telephony/ivr-flows/{id}` | `UPDATED` | `200` |
| `DELETE` | `/api/v1/telephony/ivr-flows/{id}` | `DELETED` | `204` |
| `POST` | `/api/v1/telephony/calls/{call_id}/monitor/listen` | `SUCCESS` | `200` |
| `WS` | `wss://host/ws/v1/telephony/{call_id}` | — | — |
| `WS` | `wss://host/ws/v1/telephony/monitor/{call_id}` | — | — |

---

## 11. Provider Integration

| Provider | Protocol | Features |
|---|---|---|
| Twilio | SIP/Webhook | PSTN, Recording, STT, TTS |
| Vonage (Nexmo) | SIP/Webhook | PSTN, Recording |
| FreeSWITCH | SIP | Self-hosted PBX |
| Asterisk | SIP | Self-hosted PBX |
| Custom SIP | SIP/RTP | Any SIP provider |

---

## 12. Database Tables

| Table | Purpose |
|---|---|
| `telephony.calls` | Call records |
| `telephony.recordings` | Call recordings |
| `telephony.transcripts` | Call transcripts |
| `telephony.phone_numbers` | Tenant phone numbers |
| `telephony.queues` | Call queues |
| `telephony.dtmf_events` | DTMF input events |
| `telephony.caller_auth` | Caller authentication records |
| `telephony.fraud_checks` | Anti-fraud check results |
| `telephony.audio_quality` | Audio quality metrics |
| `telephony.voicemails` | Voicemail recordings |
| `telephony.ivr_flows` | IVR flow configurations |
| `telephony.call_monitoring` | Supervisor monitoring sessions |

---

## 13. Security

| Requirement | Implementation |
|---|---|
| Caller authentication | Multi-factor: caller ID + PIN + voice biometrics |
| Anti-fraud | Caller reputation, SIM swap detection, deepfake detection |
| Audio injection prevention | Liveness detection, nonce verification |
| Call encryption | SRTP for RTP, TLS for SIP |
| Recording consent | Mandatory consent prompt, audit trail |
| DNC compliance | Check outbound against DNC list |
| Audio quality monitoring | MOS, jitter, packet loss tracking |

---

## 14. Performance Targets

| Operation | Target |
|---|---|
| Inbound call setup | < 100ms |
| Audio latency (one-way) | < 100ms |
| STT latency | < 200ms |
| TTS latency | < 200ms |
| Total round-trip | < 1s |
| DTMF detection | < 50ms |

---

# End of Part 13
