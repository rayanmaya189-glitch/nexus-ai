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

### Synthesize Speech

```
POST /api/v1/tts/synthesize
Authorization: Bearer <jwt>
```

**Request:**
```json
{
  "text": "Hello! Your account balance is ₹2,450.",
  "voice_id": "en-us-female-1",
  "speed": 1.0,
  "pitch": 0.0,
  "emotion": "cheerful"
}
```

**Response:** Audio stream (audio/wav)

### List Voices

```
GET /api/v1/tts/voices?language=en&gender=female
```

### Preview Voice

```
GET /api/v1/tts/voices/{voice_id}/preview?text=Hello+world
```

### Synthesize SSML

```
POST /api/v1/tts/ssml
Content-Type: application/xml
```

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
