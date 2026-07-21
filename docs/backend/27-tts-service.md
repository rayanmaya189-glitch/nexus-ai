# AeroXe Nexus AI — Text-to-Speech (TTS) Module

## Natural Voice Synthesis, Streaming Audio Output & Voice Personalization

> **Modular Monolith Module:** This document describes the `nexus-tts` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces.

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-tts` |
| Crate | `nexus-tts` (workspace member) |
| Bounded Context | Speech Processing |
| Domain Type | Core Domain |
| Language | Rust (edition 2024) |
| Schema | `tts_` (in shared PostgreSQL) |
| AI Model | `piper` (local) or `edge-tts` / `coqui-tts` |
| Dependencies | `nexus-telephony` (audio output), MinIO (voice assets) |

---

## 2. Purpose

The TTS module converts text into natural-sounding speech:

- Real-time streaming TTS (text → audio chunks)
- Multiple voice options (male, female, accent variants)
- SSML support for prosody control
- Emotion/intent-based voice modulation
- Voice cloning (enterprise feature)
- Pre-recorded message management
- Multi-language voice synthesis

---

## 3. Aggregate Design

### VoiceSynthesis Aggregate

```
VoiceSynthesis (Aggregate Root)
├── SynthesisRequest
│   ├── RequestId
│   ├── Text
│   ├── Voice
│   │   ├── VoiceId
│   │   ├── Language
│   │   ├── Gender
│   │   └── Style (neutral, cheerful, sad, serious)
│   ├── SSML (optional)
│   └── OutputFormat
├── SynthesisResult
│   ├── AudioChunks[]
│   ├── Duration
│   ├── SampleRate
│   └── Codec
└── VoiceProfile
    ├── VoiceId
    ├── Name
    ├── Language
    ├── Gender
    ├── SampleAudio
    └── Config
```

---

## 4. Public API Trait

```rust
// nexus-tts/src/interfaces/api.rs
#[async_trait]
pub trait TTSService: Send + Sync {
    // Streaming synthesis (for real-time calls)
    async fn start_streaming_synthesis(&self, req: StreamingSynthesisRequest) -> Result<StreamingSynthesisHandle, TTSError>;
    async fn synthesize_chunk(&self, session_id: SessionId, text: String) -> Result<Receiver<AudioChunk>, TTSError>;
    async fn end_streaming_synthesis(&self, session_id: SessionId) -> Result<(), TTSError>;
    
    // Batch synthesis (for pre-recorded messages)
    async fn synthesize(&self, req: SynthesisRequest) -> Result<SynthesisResult, TTSError>;
    async fn synthesize_to_file(&self, req: SynthesisRequest) -> Result<String, TTSError>;
    
    // Voice management
    async fn list_voices(&self, tenant_id: TenantId) -> Result<Vec<VoiceProfile>, TTSError>;
    async fn get_voice(&self, voice_id: VoiceId) -> Result<VoiceProfile, TTSError>;
    async fn preview_voice(&self, voice_id: VoiceId, text: String) -> Result<Vec<u8>, TTSError>;
    
    // SSML synthesis
    async fn synthesize_ssml(&self, req: SSMLRequest) -> Result<SynthesisResult, TTSError>;
}

pub struct StreamingSynthesisRequest {
    pub tenant_id: TenantId,
    pub voice_id: Option<VoiceId>,
    pub language: String,
    pub sample_rate: u32,
    pub codec: AudioCodec,
    pub chunk_size_ms: u32,
}

pub struct SynthesisRequest {
    pub text: String,
    pub voice_id: Option<VoiceId>,
    pub language: String,
    pub speed: f32,            // 0.5 - 2.0
    pub pitch: f32,            // -1.0 to 1.0
    pub volume: f32,           // 0.0 to 2.0
    pub emotion: Option<String>,
    pub output_format: AudioCodec,
}

pub struct SSMLRequest {
    pub ssml: String,
    pub voice_id: Option<VoiceId>,
    pub language: String,
}

pub struct VoiceProfile {
    pub voice_id: VoiceId,
    pub name: String,
    pub language: String,
    pub gender: String,
    pub sample_rate: u32,
    pub preview_url: String,
    pub styles: Vec<String>,
    pub is_cloned: bool,
}
```

---

## 5. Voice Options

### 5.1 Built-in Voices

| Voice ID | Name | Language | Gender | Style |
|---|---|---|---|---|
| `en-us-female-1` | Emma | English (US) | Female | Neutral, Friendly |
| `en-us-male-1` | James | English (US) | Male | Neutral, Professional |
| `en-gb-female-1` | Sophie | English (UK) | Female | Neutral |
| `hi-in-female-1` | Priya | Hindi | Female | Neutral, Friendly |
| `hi-in-male-1` | Rahul | Hindi | Male | Neutral |
| `mr-in-female-1` | Anjali | Marathi | Female | Neutral |
| `ta-in-female-1` | Kavitha | Tamil | Female | Neutral |
| `te-in-female-1` | Lakshmi | Telugu | Female | Neutral |

### 5.2 Voice Styles

| Style | Description | Use Case |
|---|---|---|
| `neutral` | Standard, balanced | General conversations |
| `cheerful` | Warm, friendly | Customer support |
| `serious` | Formal, authoritative | Business calls |
| `empathetic` | Understanding, caring | Complaint handling |
| `energetic` | Enthusiastic | Sales calls |
| `calm` | Soothing, gentle | Healthcare, support |

---

## 6. SSML Support

```xml
<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis">
    <voice name="en-us-female-1">
        <prosody rate="1.0" pitch="+2st" volume="medium">
            Hello! Welcome to AeroXe support.
        </prosody>
        <break time="500ms"/>
        <prosody rate="0.9">
            How can I help you today?
        </prosody>
    </voice>
</speak>
```

### SSML Tags Supported

| Tag | Purpose |
|---|---|
| `<voice>` | Select voice |
| `<prosody>` | Control rate, pitch, volume |
| `<break>` | Insert pause |
| `<emphasis>` | Emphasize words |
| `<say-as>` | Interpret text (dates, numbers, phone) |
| `<phoneme>` | Pronunciation override |
| `<sub>` | Substitution text |

---

## 7. Streaming Architecture

### 7.1 Real-Time Streaming Pipeline

```
Agent Response Text
    |
    v
[1] Text Buffer
    |  - Accumulate text chunks
    |  - Sentence boundary detection
    |
    v
[2] Text Pre-processing
    |  - Abbreviation expansion
    |  - Number formatting
    |  - Punctuation normalization
    |  - Special character handling
    |
    v
[3] Voice Synthesis (Piper/Coqui)
    |  - Convert text → mel spectrogram
    |  - Convert mel → audio waveform
    |  - Stream audio chunks as generated
    |
    v
[4] Audio Post-processing
    |  - Resample to target rate
    |  - Apply volume normalization
    |  - Convert to target codec
    |
    v
[5] Stream to Telephony
    |  - Send audio chunks to call session
    |  - Maintain timing/pace
    |  - Handle barge-in (caller interrupt)
```

### 7.2 Streaming Latency Optimization

| Technique | Impact |
|---|---|
| Sentence-level streaming | Start TTS before full response |
| Chunked synthesis | Generate 50-100ms audio chunks |
| Pre-buffering | Buffer 2-3 chunks ahead |
| Model optimization | INT8 quantization, ONNX runtime |
| GPU acceleration | CUDA inference for TTS model |

---

## 8. Database Schema (tts_ schema)

### tts_voices

```sql
CREATE TABLE tts.voices (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT,                       -- NULL = platform-wide
    voice_id VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    language VARCHAR(10) NOT NULL,
    gender VARCHAR(10),
    sample_rate INT NOT NULL DEFAULT 16000,
    styles JSONB,
    preview_path TEXT,
    is_cloned BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### tts_synthesis_log

```sql
CREATE TABLE tts.synthesis_log (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    session_id VARCHAR(100),
    call_id BIGINT,
    voice_id VARCHAR(50) NOT NULL,
    text_length INT NOT NULL,
    audio_duration_ms INT,
    latency_ms FLOAT,
    model VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tts_log_tenant ON tts.synthesis_log(tenant_id, created_at DESC);
```

---

## 9. REST API Endpoints

| Method | Endpoint | Business Status | HTTP | Description |
|---|---|---|---|---|
| `POST` | `/api/v1/tts/synthesize` | `SUCCESS` | `200` | Synthesize speech |
| `POST` | `/api/v1/tts/ssml` | `SUCCESS` | `200` | Synthesize SSML |
| `GET` | `/api/v1/tts/voices?limit=10&offset=0&language=en` | `SUCCESS` | `200` | List voices |
| `GET` | `/api/v1/tts/voices/{voice_id}` | `SUCCESS` | `200` | Get voice |
| `GET` | `/api/v1/tts/voices/{voice_id}/preview?text=Hello` | `SUCCESS` | `200` | Preview voice |
| `POST` | `/api/v1/tts/voices/clone` | `CREATED` | `201` | Clone voice |
| `DELETE` | `/api/v1/tts/voices/clone/{clone_id}` | `DELETED` | `204` | Revoke clone |

**Note:** No PUT method. Use PATCH for updates. All list endpoints support `limit` (default 10) and `offset`.

---

## 10. Observability

| Metric | Description |
|---|---|
| `tts_synthesis_total` | Total synthesis requests |
| `tts_synthesis_latency_ms` | Synthesis latency |
| `tts_audio_duration_seconds` | Audio generated |
| `tts_characters_processed` | Characters synthesized |
| `tts_streaming_sessions` | Active streaming sessions |
| `tts_errors_total` | Synthesis errors |

---

## 11. Performance Targets

| Operation | Target |
|---|---|
| First audio chunk | < 150ms |
| Streaming chunk latency | < 50ms |
| Batch synthesis (100 words) | < 1s |
| Voice list query | < 50ms |
| SSML synthesis | < 200ms |

---

## 12. Voice Clone Authorization (NEW)

### 12.1 Clone Authorization Flow

```
Tenant Requests Voice Clone
    |
    v
[1] Authorization Check
    |  - Tenant has clone permission?
    |  - Clone limit not exceeded?
    |
    v
[2] Source Audio Collection
    |  - Provide reference audio samples (min 30 seconds)
    |  - Verify speaker consent
    |  - Store consent record
    |
    v
[3] Clone Training
    |  - Train voice model from samples
    |  - Generate voice profile
    |  - Quality validation
    |
    v
[4] Deployment
    |  - Add to tenant voice library
    |  - Set usage limits
    |  - Enable audit logging
```

### 12.2 Clone Authorization Entities

```sql
CREATE TABLE tts.voice_clones (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    clone_id UUID NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    source_speaker VARCHAR(100),          -- Person whose voice was cloned
    consent_recorded BOOLEAN NOT NULL DEFAULT false,
    consent_storage_path TEXT,
    reference_audio_path TEXT,
    model_path TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'training', -- training | active | revoked
    usage_count INT NOT NULL DEFAULT 0,
    max_usage INT,                        -- NULL = unlimited
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP
);
```

---

## 13. Multi-Language Voice Switching (NEW)

### 13.1 Language-Voice Mapping

```rust
pub struct LanguageVoiceMapping {
    pub language: String,              // "en", "hi", "mr", etc.
    pub primary_voice: VoiceId,        // Default voice for this language
    pub fallback_voice: VoiceId,       // Backup if primary unavailable
    pub accent: String,                // "us", "uk", "in", etc.
}

pub struct TenantLanguageConfig {
    pub tenant_id: TenantId,
    pub mappings: Vec<LanguageVoiceMapping>,
    pub auto_switch: bool,             // Auto-switch on language change
    pub confirmation_required: bool,   // Ask before switching
}
```

### 13.2 Auto-Switch Flow

```
STT Detects Language Change
    |
    v
[1] Compare to Previous Language
    |  - Same: Continue with current voice
    |  - Different: Check auto-switch config
    |
    v
[2] Auto-Switch Decision
    |  - Auto-switch enabled: Switch voice immediately
    |  - Confirmation required: "I detect you're speaking Hindi. Switch to Hindi voice?"
    |
    v
[3] Voice Switch
    |  - Load new language voice model
    |  - Apply prosody settings
    |  - Continue synthesis
```

---

## 14. Sentiment-Aware Voice Adaptation (NEW)

### 14.1 Sentiment → Voice Mapping

| Sentiment | TTS Adaptation |
|---|---|
| **Happy/Satisfied** | Cheerful tone, normal pace |
| **Neutral** | Neutral tone, normal pace |
| **Frustrated** | Empathetic tone, slightly slower |
| **Angry** | Calm tone, slower pace, softer volume |
| **Confused** | Clear articulation, slower pace |
| **Urgent** | Faster pace, clear enunciation |

### 14.2 Adaptation Flow

```
Sentiment Analysis Result
    |
    v
[1] Map Sentiment to Voice Style
    |  - Angry → empathetic + slow
    |  - Frustrated → calm + slower
    |  - Happy → cheerful + normal
    |
    v
[2] Apply to TTS
    |  - Update SSML prosody
    |  - Adjust rate, pitch, volume
    |  - Select emotion style
    |
    v
[3] Synthesize with Adapted Voice
```

---

## 15. Post-Call Survey (NEW)

### 15.1 Survey Flow

```
Call Ended
    |
    v
[1] Play Survey Prompt
    |  - "Please rate your experience from 1 to 5"
    |  - "Press 1 for poor, 5 for excellent"
    |
    v
[2] Collect Response
    |  - DTMF: 1-5 rating
    |  - Or speech: "Four" → "4"
    |
    v
[3] Optional Follow-up
    |  - "Would you like to leave a comment?"
    |  - Record voicemail-style comment
    |
    v
[4] Store & Process
    |  - Save to CSAT database
    |  - Update customer satisfaction score
    |  - Trigger alert if low score
```

### 15.2 Survey Entities

```sql
CREATE TABLE tts.post_call_surveys (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    survey_id UUID NOT NULL UNIQUE,
    call_id BIGINT NOT NULL REFERENCES telephony.calls(id),
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT,
    rating INT,                          -- 1-5
    comment TEXT,
    comment_audio_path TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```
