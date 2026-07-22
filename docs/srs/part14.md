# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 14 — Speech Processing (STT/TTS) Architecture

## Speech-to-Text, Text-to-Speech, Multi-Language, Voice Cloning

---

## 1. STT Module Overview

The STT module converts spoken audio into text in real-time:

- Streaming audio transcription (real-time, chunk-by-chunk)
- Batch audio file transcription
- Multi-language speech recognition
- Speaker identification
- PII redaction in transcripts
- Voice command detection
- Confidence threshold configuration
- Anti-injection protection

---

## 2. STT Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-stt` |
| Bounded Context | Speech Processing |
| Schema | `stt_` (shared PostgreSQL) |
| AI Model | `whisper` (via Ollama) |

---

## 3. STT Public API Trait

```rust
#[async_trait]
pub trait STTService: Send + Sync {
    async fn start_streaming_session(&self, req: StreamingSessionRequest) -> Result<StreamingSessionHandle>;
    async fn send_audio_chunk(&self, session_id: SessionId, chunk: AudioChunk) -> Result<PartialTranscript>;
    async fn end_streaming_session(&self, session_id: SessionId) -> Result<FinalTranscript>;
    async fn transcribe_audio(&self, req: TranscribeRequest) -> Result<Transcript>;
    async fn get_config(&self, tenant_id: TenantId) -> Result<STTConfig>;
    async fn check_liveness(&self, audio: Vec<u8>) -> Result<LivenessResult>;
}
```

---

## 4. Confidence Threshold Configuration

```rust
pub struct STTConfig {
    pub tenant_id: TenantId,
    pub min_confidence_threshold: f32,
    pub low_confidence_action: LowConfidenceAction,
    pub language_override: Option<String>,
    pub enable_pii_redaction: bool,
    pub enable_diarization: bool,
}

pub enum LowConfidenceAction {
    AcceptAll, Reject, FlagForReview, FallbackToHuman, AskConfirmation,
}
```

---

## 5. Anti-Injection Protection

| Attack | Detection | Action |
|---|---|---|
| Pre-recorded speech | Liveness detection | Reject audio |
| Deepfake voice | Deepfake detection model | Block call |
| Replay attack | Nonce verification | Reject audio |

---

## 6. STT Entities

| Table | Purpose |
|---|---|
| `stt.sessions` | Transcription sessions |
| `stt.segments` | Transcription segments |
| `stt.models` | STT model registry |

---

## 7. REST API Endpoints

| Method | Endpoint | Business Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/stt/sessions` | `CREATED` | `201` |
| `POST` | `/api/v1/stt/sessions/{session_id}/audio` | `SUCCESS` | `200` |
| `POST` | `/api/v1/stt/sessions/{session_id}/end` | `UPDATED` | `200` |
| `POST` | `/api/v1/stt/transcribe` | `SUCCESS` | `200` |
| `POST` | `/api/v1/stt/sessions/{session_id}` | `SUCCESS` | `200` |

---

## 8. TTS Module Overview

The TTS module converts text into natural-sounding speech:

- Real-time streaming TTS (text → audio chunks)
- Multiple voice options (male, female, accent variants)
- SSML support for prosody control
- Emotion/intent-based voice modulation
- Voice cloning (enterprise feature)
- Multi-language voice synthesis
- Sentiment-aware voice adaptation

---

## 9. TTS Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-tts` |
| Bounded Context | Speech Processing |
| Schema | `tts_` (shared PostgreSQL) |
| AI Model | `piper` (local) or `edge-tts` / `coqui-tts` |

---

## 10. TTS Public API Trait

```rust
#[async_trait]
pub trait TTSService: Send + Sync {
    async fn start_streaming_synthesis(&self, req: StreamingSynthesisRequest) -> Result<StreamingSynthesisHandle>;
    async fn synthesize_chunk(&self, session_id: SessionId, text: String) -> Result<Receiver<AudioChunk>>;
    async fn synthesize(&self, req: SynthesisRequest) -> Result<SynthesisResult>;
    async fn list_voices(&self, tenant_id: TenantId) -> Result<Vec<VoiceProfile>>;
    async fn clone_voice(&self, req: VoiceCloneRequest) -> Result<VoiceClone>;
    async fn revoke_clone(&self, clone_id: CloneId) -> Result<()>;
    async fn synthesize_with_sentiment(&self, req: SentimentSynthesisRequest) -> Result<SynthesisResult>;
}
```

---

## 11. Voice Clone Authorization

```
Tenant Requests Voice Clone → Authorization Check → Source Audio Collection
  → Clone Training → Quality Validation → Deployment with Usage Limits
```

### Entities

```sql
CREATE TABLE tts.voice_clones (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    clone_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    source_speaker VARCHAR(100),
    consent_recorded BOOLEAN NOT NULL DEFAULT false,
    consent_storage_path TEXT,
    reference_audio_path TEXT,
    model_path TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'training',
    usage_count INT NOT NULL DEFAULT 0,
    max_usage INT,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP
);
```

---

## 12. Sentiment-Aware Voice Adaptation

| Sentiment | TTS Adaptation |
|---|---|
| Happy | Cheerful tone, normal pace |
| Frustrated | Empathetic tone, slightly slower |
| Angry | Calm tone, slower pace, softer volume |
| Confused | Clear articulation, slower pace |

---

## 13. TTS Entities

| Table | Purpose |
|---|---|
| `tts.voices` | Voice profiles |
| `tts.voice_clones` | Voice clone authorizations |
| `tts.synthesis_log` | Synthesis audit trail |

---

## 14. REST API Endpoints

| Method | Endpoint | Business Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/tts/synthesize` | `SUCCESS` | `200` |
| `POST` | `/api/v1/tts/ssml` | `SUCCESS` | `200` |
| `POST` | `/api/v1/tts/voices?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/tts/voices/clone` | `CREATED` | `201` |
| `DELETE` | `/api/v1/tts/voices/clone/{clone_id}` | `DELETED` | `204` |

---

# End of Part 14
