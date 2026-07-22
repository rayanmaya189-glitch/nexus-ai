# AeroXe Nexus AI — Speech-to-Text (STT) Module

## Real-Time Speech Recognition, Streaming Transcription & Audio Processing

> **Modular Monolith Module:** This document describes the `nexus-stt` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via gRPC (sync) or NATS (async) between services.

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-stt` |
| Crate | `nexus-stt` (workspace member) |
| Bounded Context | Speech Processing |
| Domain Type | Core Domain |
| Language | Rust (edition 2024) |
| Schema | `stt_` (in shared PostgreSQL) |
| AI Model | `whisper` (via Ollama or local runtime) |
| Dependencies | `nexus-telephony` (audio input), Ollama (Whisper model) |

---

## 2. Purpose

The STT module converts spoken audio into text in real-time:

- Streaming audio transcription (real-time, chunk-by-chunk)
- Batch audio file transcription
- Multi-language speech recognition
- Speaker identification (who is speaking)
- PII redaction in transcripts
- Profanity filtering in speech
- Voice command detection

---

## 3. Aggregate Design

### TranscriptionSession Aggregate

```
TranscriptionSession (Aggregate Root)
├── SessionMetadata
│   ├── SessionId
│   ├── CallId (optional)
│   ├── TenantId
│   ├── Language
│   └── Model
├── AudioConfig
│   ├── SampleRate
│   ├── Channels
│   ├── Codec
│   └── ChunkSizeMs
├── TranscriptionState
│   ├── PartialText (current in-progress)
│   ├── FinalText[] (completed segments)
│   ├── SpeakerLabels[]
│   └── Timestamps[]
└── Result
    ├── FullTranscript
    ├── WordTimestamps[]
    └── Confidence
```

---

## 4. Public API Trait

```rust
// nexus-stt/src/interfaces/api.rs
#[async_trait]
pub trait STTService: Send + Sync {
    // Real-time streaming transcription
    async fn start_streaming_session(&self, req: StreamingSessionRequest) -> Result<StreamingSessionHandle, STTError>;
    async fn send_audio_chunk(&self, session_id: SessionId, chunk: AudioChunk) -> Result<PartialTranscript, STTError>;
    async fn end_streaming_session(&self, session_id: SessionId) -> Result<FinalTranscript, STTError>;
    async fn get_partial_transcript(&self, session_id: SessionId) -> Result<PartialTranscript, STTError>;
    
    // Batch transcription (for recordings)
    async fn transcribe_audio(&self, req: TranscribeRequest) -> Result<Transcript, STTError>;
    async fn transcribe_file(&self, req: FileTranscribeRequest) -> Result<Transcript, STTError>;
    
    // Language detection
    async fn detect_language(&self, audio: Vec<u8>) -> Result<Language, STTError>;
    
    // Speaker diarization
    async fn diarize_speakers(&self, audio: Vec<u8>, num_speakers: Option<u32>) -> Result<Vec<SpeakerSegment>, STTError>;
}

pub struct StreamingSessionRequest {
    pub tenant_id: TenantId,
    pub call_id: Option<CallId>,
    pub language: Option<String>,       // Auto-detect if None
    pub model: Option<String>,          // Whisper variant
    pub sample_rate: u32,               // 8000, 16000, etc.
    pub enable_punctuation: bool,
    pub enable_speaker_labels: bool,
    pub redact_pii: bool,
}

pub struct AudioChunk {
    pub session_id: SessionId,
    pub data: Vec<u8>,                  // Raw PCM audio
    pub is_final: bool,                 // Last chunk of utterance
    pub timestamp: u64,
}

pub struct PartialTranscript {
    pub text: String,
    pub is_final: bool,
    pub confidence: f32,
    pub words: Vec<WordTimestamp>,
}

pub struct FinalTranscript {
    pub text: String,
    pub confidence: f32,
    pub words: Vec<WordTimestamp>,
    pub duration_ms: u32,
    pub language: String,
    pub speakers: Vec<SpeakerLabel>,
}

pub struct WordTimestamp {
    pub word: String,
    pub start_ms: u32,
    pub end_ms: u32,
    pub confidence: f32,
}

pub struct SpeakerLabel {
    pub speaker_id: String,
    pub start_ms: u32,
    pub end_ms: u32,
}
```

---

## 5. Real-Time Streaming Architecture

### 5.1 Streaming Pipeline

```
Audio Stream (from Telephony)
    |
    v
[1] Audio Buffer
    |  - Accumulate audio chunks
    |  - Configurable chunk size (100ms-500ms)
    |
    v
[2] Voice Activity Detection (VAD)
    |  - Detect speech vs silence
    |  - Start transcription on speech
    |  - End transcription on silence
    |
    v
[3] Audio Pre-processing
    |  - Noise reduction
    |  - Normalization
    |  - Resampling (if needed)
    |
    v
[4] Whisper Streaming
    |  - Feed audio chunks to Whisper
    |  - Generate partial transcripts
    |  - Generate final transcripts on silence
    |
    v
[5] Post-processing
    |  - Punctuation insertion
    |  - Capitalization
    |  - PII redaction
    |  - Speaker labeling
    |
    v
[6] Output
    |  - Partial transcript (streaming)
    |  - Final transcript (on utterance end)
    |  - Word timestamps
    |  - Confidence scores
```

### 5.2 Whisper Model Selection

| Model | Size | Speed | Accuracy | Use Case |
|---|---|---|---|---|
| whisper-tiny | 39M | Fastest | Lower | Low-latency, simple commands |
| whisper-base | 74M | Fast | Good | Balanced speed/accuracy |
| whisper-small | 244M | Medium | High | Production quality |
| whisper-medium | 769M | Slow | Very High | High-accuracy needs |
| whisper-large-v3 | 1.5G | Slowest | Highest | Maximum accuracy |

### 5.3 Language Support

| Language | Code | Status |
|---|---|---|
| English | en | Primary |
| Hindi | hi | Primary |
| Marathi | mr | Supported |
| Tamil | ta | Supported |
| Telugu | te | Supported |
| Bengali | bn | Supported |
| Gujarati | gu | Supported |
| Kannada | kn | Supported |
| Malayalam | ml | Supported |
| Punjabi | pa | Supported |
| Auto-detect | auto | Supported |

---

## 6. Database Schema (stt_ schema)

### transcription_sessions

```sql
CREATE TABLE stt_.sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    call_id BIGINT,
    language VARCHAR(10) DEFAULT 'auto',
    model VARCHAR(50) NOT NULL DEFAULT 'whisper-small',
    sample_rate INT NOT NULL DEFAULT 16000,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    partial_text TEXT,
    final_text TEXT,
    word_count INT DEFAULT 0,
    confidence FLOAT,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### transcription_segments

```sql
CREATE TABLE stt_.segments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES stt_.sessions(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL,
    text TEXT NOT NULL,
    speaker VARCHAR(20),
    start_ms INT NOT NULL,
    end_ms INT NOT NULL,
    confidence FLOAT,
    is_final BOOLEAN NOT NULL DEFAULT true,
    word_timestamps JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stt_segments_session ON stt_.segments(session_id, start_ms);
```

### stt_models

```sql
CREATE TABLE stt_.models (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    size_bytes BIGINT,
    language VARCHAR(10),
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    avg_latency_ms FLOAT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 7. REST API Endpoints (Protobuf over HTTP)

| Method | Endpoint | Business Status | HTTP | Description |
|---|---|---|---|---|
| `POST` | `/api/v1/stt/sessions` | `CREATED` | `201` | Start streaming session |
| `POST` | `/api/v1/stt/sessions/{session_id}/audio` | `SUCCESS` | `200` | Send audio chunk |
| `POST` | `/api/v1/stt/sessions/{session_id}/end` | `UPDATED` | `200` | End session |
| `POST` | `/api/v1/stt/transcribe` | `SUCCESS` | `200` | Batch transcribe |
| `POST` | `/api/v1/stt/sessions` (read body) | `SUCCESS` | `200` | Get transcript |

**Note:** All endpoints use Protobuf (proto3) serialized as JSON over HTTP. Content-Type: `application/json` (Protobuf messages serialized as JSON). No PUT method — use PATCH for updates.

---

## 8. Observability

| Metric | Description |
|---|---|
| `stt_sessions_total` | Total transcription sessions |
| `stt_audio_chunks_processed` | Audio chunks processed |
| `stt_latency_ms` | Transcription latency |
| `stt_partial_transcripts` | Partial results generated |
| `stt_final_transcripts` | Final results generated |
| `stt_errors_total` | Transcription errors |
| `stt_language_detection_ms` | Language detection time |
| `stt_pii_redactions_total` | PII items redacted |

---

## 9. Performance Targets

| Operation | Target |
|---|---|
| Streaming session start | < 100ms |
| Partial transcript latency | < 150ms |
| Final transcript latency | < 300ms |
| Batch transcription (1 min audio) | < 5s |
| Language detection | < 500ms |
| Speaker diarization | < 10s (1 min audio) |
| PII redaction | < 10ms per segment |

---

## 10. Confidence Threshold Configuration (NEW)

### 10.1 Per-Tenant Configuration

```rust
pub struct STTConfig {
    pub tenant_id: TenantId,
    pub min_confidence_threshold: f32,   // Default: 0.7
    pub low_confidence_action: LowConfidenceAction,
    pub language_override: Option<String>, // Force specific language
    pub enable_pii_redaction: bool,
    pub enable_diarization: bool,
    pub max_silence_ms: u32,             // Default: 3000
    pub min_speech_ms: u32,              // Default: 200
}

pub enum LowConfidenceAction {
    AcceptAll,           // Accept even low confidence
    Reject,              // Reject and ask to repeat
    FlagForReview,       // Accept but flag for quality review
    FallbackToHuman,     // Transfer to human agent
    AskConfirmation,     // "Did you say X? Press 1 to confirm"
}
```

### 10.2 Confidence-Based Routing

```
Transcription Received
    |
    v
[1] Check Confidence Score
    |  - High confidence (>0.85): Process normally
    |  - Medium confidence (0.7-0.85): Ask confirmation
    |  - Low confidence (<0.7): Apply configured action
    |
    v
[2] Apply Action
    |  - AcceptAll: Process
    |  - Reject: "Sorry, I didn't catch that. Could you repeat?"
    |  - FlagForReview: Process + quality flag
    |  - FallbackToHuman: Transfer call
    |  - AskConfirmation: "Did you say [X]? Press 1 for yes, 2 for no"
```

---

## 11. Anti-Injection Protection (NEW)

### 11.1 Audio Injection Detection

| Attack | Detection | Action |
|---|---|---|
| Pre-recorded speech | Liveness detection, audio analysis | Reject audio |
| Deepfake voice | Deepfake detection model | Block call |
| Replay attack | Nonce verification | Reject audio |
| Background audio injection | Audio source analysis | Flag for review |

### 11.2 Liveness Detection Pipeline

```
Audio Stream
    |
    v
[1] Audio Provenance
    |  - Verify RTP sequence continuity
    |  - Check timestamp consistency
    |  - Detect audio source (real-time vs pre-recorded)
    |
    v
[2] Liveness Analysis
    |  - Speech pattern naturalness
    |  - Background noise consistency
    |  - Microphone characteristics
    |
    v
[3] Decision
    |  - Live: Process transcription
    |  - Suspicious: Flag + require additional auth
    |  - Fake: Block + alert
```
