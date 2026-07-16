# AeroXe Nexus AI — Vision Intelligence Service

## Image Processing, OCR, Visual Reasoning & Device Analysis

---

## 1. Service Identity

| Attribute | Value |
|---|---|
| Service Name | `vision-service` |
| Bounded Context | Vision Intelligence |
| Domain Type | Core Domain |
| Language | Rust |
| Database | `vision_db` (PostgreSQL) |
| AI Model | `qwen3-vl:4b` (Ollama) |
| gRPC Port | 50055 |
| Object Storage | MinIO (`aeroxe-images` bucket) |

---

## 2. Purpose

The Vision Service provides AI-powered image understanding capabilities:

- Image analysis and description
- Optical Character Recognition (OCR)
- Device troubleshooting (ONU/router LED analysis)
- Invoice and document extraction
- Screenshot analysis
- Diagram and chart understanding

---

## 3. Aggregate Design

### VisionAnalysis Aggregate

```
VisionAnalysis (Aggregate Root)
├── Image
│   ├── ImageId
│   ├── StoragePath
│   ├── FileType
│   └── SizeBytes
├── Detection
│   ├── Description
│   ├── Confidence
│   └── DetectedObjects[]
└── Extraction
    ├── ExtractedText (OCR)
    ├── StructuredData
    └── Metadata
```

### Entities

| Entity | Attributes |
|---|---|
| Image | ImageId, TenantId, StoragePath, Type, Size, CreatedAt |
| AnalysisResult | ResultId, ImageId, Model, Description, Confidence, Metadata |
| OCRResult | ResultId, ImageId, Text, Language, Confidence |

### Value Objects

| Value Object | Type |
|---|---|
| `ImageId` | UUID |
| `ImageType` | Enum (png, jpg, jpeg, webp, gif, bmp, tiff) |
| `AnalysisType` | Enum (describe, ocr, troubleshoot, extract, detect) |

---

## 4. gRPC Contract

```protobuf
syntax = "proto3";
package aeroxe.vision;

service VisionService {
  rpc AnalyzeImage(ImageRequest) returns (ImageAnalysisResponse);
  rpc ExtractText(ImageRequest) returns (OCRResponse);
  rpc TroubleshootDevice(DeviceImageRequest) returns (TroubleshootResponse);
  rpc BatchAnalyze(BatchImageRequest) returns (BatchAnalysisResponse);
}

message ImageRequest {
  bytes image = 1;
  string type = 2;
  string task = 3;
  string tenant_id = 4;
}

message ImageAnalysisResponse {
  string description = 1;
  float confidence = 2;
  repeated DetectedObject objects = 3;
  map<string, string> metadata = 4;
}

message OCRResponse {
  string text = 1;
  float confidence = 2;
  string language = 3;
  repeated TextRegion regions = 4;
}

message DeviceImageRequest {
  bytes image = 1;
  string device_type = 2;
  string tenant_id = 3;
}

message TroubleshootResponse {
  string diagnosis = 1;
  float confidence = 2;
  repeated string recommendations = 3;
  string severity = 4;
}
```

---

## 5. Image Analysis Pipeline

```
Image Upload
    |
    v
[1] Image Validation
    |  - File type check
    |  - Size limit (max 20MB)
    |  - Format verification
    |
    v
[2] Store in MinIO
    |  - aeroxe-images bucket
    |  - Path: /{tenant_id}/{date}/{uuid}.{ext}
    |
    v
[3] Create Database Record
    |  - vision_db.images table
    |
    v
[4] Pre-processing
    |  - Resize if too large (max 1024x1024 for analysis)
    |  - Normalize format
    |  - Convert to base64 for Ollama
    |
    v
[5] AI Model Invocation
    |  - Ollama Qwen3-VL:4B
    |  - Send image + task prompt
    |
    v
[6] Post-processing
    |  - Parse model output
    |  - Extract structured data
    |  - Confidence scoring
    |
    v
[7] Store Results
    |  - vision_db.vision_analysis table
    |  - vision_db.ocr_results table (if OCR)
    |
    v
[8] Return Response
```

---

## 6. Use Cases

### 6.1 Device Troubleshooting (ISP)

**Input:** Photo of ONU/router showing LED indicators

**AI Analysis:**
- Detect LED colors and patterns
- Cross-reference with device model knowledge
- Identify potential issues

**Output:**
```json
{
  "description": "ONU shows red PON LED and blinking internet LED",
  "confidence": 0.94,
  "diagnosis": "PON signal loss - fiber connection issue",
  "severity": "HIGH",
  "recommendations": [
    "Check fiber optic cable connection",
    "Verify OLT port status",
    "Check for fiber bend or damage",
    "Contact field technician if issue persists"
  ]
}
```

### 6.2 Invoice Processing

**Input:** Invoice image/PDF

**AI Analysis:**
- OCR text extraction
- Key-value pair extraction
- Amount, date, vendor identification

**Output:**
```json
{
  "extracted_text": "Invoice #12345...",
  "structured_data": {
    "invoice_number": "12345",
    "vendor": "ABC Corp",
    "amount": "50000.00",
    "currency": "INR",
    "date": "2026-07-15",
    "due_date": "2026-08-15"
  }
}
```

### 6.3 Document Extraction

**Input:** Scanned document, form, or screenshot

**Output:**
```json
{
  "text": "Full extracted text...",
  "document_type": "form",
  "fields": {
    "name": "John Doe",
    "id_number": "ABCDE1234F"
  }
}
```

---

## 7. Database Schema (vision_db)

### images

```sql
CREATE TABLE images (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    storage_path TEXT NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_size_bytes BIGINT,
    width INT,
    height INT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### vision_analysis

```sql
CREATE TABLE vision_analysis (
    id UUID PRIMARY KEY,
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    model VARCHAR(100) NOT NULL DEFAULT 'qwen3-vl:4b',
    analysis_type VARCHAR(50) NOT NULL,
    description TEXT,
    confidence FLOAT,
    detected_objects JSONB,
    metadata JSONB,
    latency_ms FLOAT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### ocr_results

```sql
CREATE TABLE ocr_results (
    id UUID PRIMARY KEY,
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    language VARCHAR(10) DEFAULT 'en',
    confidence FLOAT,
    regions JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 8. REST API Endpoints

### Analyze Image

```
POST /api/v1/vision/analyze
Content-Type: multipart/form-data
```

**Parameters:**
- `image` - Image file
- `task` - Analysis task (describe, troubleshoot, extract)

**Response:**
```json
{
  "image_id": "uuid",
  "description": "Router LED is showing red",
  "confidence": 0.94,
  "analysis_type": "troubleshoot",
  "recommendations": ["Check cable", "Restart device"]
}
```

### Extract Text (OCR)

```
POST /api/v1/vision/ocr
Content-Type: multipart/form-data
```

### Batch Analysis

```
POST /api/v1/vision/batch
Content-Type: multipart/form-data
```

---

## 9. NATS Events

| Subject | Event |
|---|---|
| `aeroxe.vision.image.received` | Image uploaded |
| `aeroxe.vision.analysis.completed` | Analysis complete |
| `aeroxe.vision.ocr.completed` | OCR complete |

---

## 10. Model Configuration

| Parameter | Value |
|---|---|
| Model | `qwen3-vl:4b` |
| Context Window | 8192 tokens |
| Max Image Size | 1024x1024 (auto-resized) |
| Supported Formats | PNG, JPG, JPEG, WebP, GIF, BMP, TIFF |
| Inference Backend | Ollama HTTP API |

### Prompt Templates

| Task | System Prompt |
|---|---|
| Describe | "Analyze this image and provide a detailed description." |
| Troubleshoot | "You are an ISP network technician. Analyze the device image and identify any issues based on LED indicators, connections, and visible hardware state." |
| OCR | "Extract all visible text from this image. Preserve formatting and structure." |
| Extract | "Extract structured data from this document image. Identify key fields and values." |

---

## 11. Performance Targets

| Operation | Target |
|---|---|
| Image Upload | < 2s |
| Vision Analysis | < 5s |
| OCR Extraction | < 3s |
| Batch (5 images) | < 15s |
| Image Storage | < 1s |

---

## 12. Security

| Requirement | Implementation |
|---|---|
| File Validation | MIME type + magic bytes check |
| Size Limit | 20MB per image |
| Tenant Isolation | All queries filtered by tenant_id |
| Storage Encryption | MinIO server-side encryption |
| Access Control | User can only access own tenant's images |
| Prompt Injection | Image prompts sanitized before model |
