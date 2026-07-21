# AeroXe Nexus AI — Security AI Module

## AI-Powered Security Analysis, Vulnerability Detection & Threat Intelligence

> **Modular Monolith Module:** This document describes the `nexus-security-ai` crate — a module within the single `aeroxe-nexus` binary. It communicates with other modules via Rust trait interfaces (see [Communication Architecture](12-communication-architecture.md)).

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-security-ai` |
| Crate | `nexus-security-ai` (workspace member) |
| Bounded Context | Security Intelligence |
| Domain Type | Core Domain |
| Language | Rust |
| AI Model | `whiterabbitneo:7b` (Ollama) |
| Dependencies | Ollama (inference), Elasticsearch (event storage) |

---

## 2. Purpose

The Security AI module provides AI-powered security intelligence within the `aeroxe-nexus` monolith:

- Code security review and vulnerability detection
- Threat analysis and risk assessment
- Secure coding recommendations
- Infrastructure security scanning
- Security audit report generation

---

## 3. Aggregate Design

### SecurityAnalysis Aggregate

```
SecurityAnalysis (Aggregate Root)
├── Finding[]
│   ├── Severity (CRITICAL, HIGH, MEDIUM, LOW, INFO)
│   ├── Category
│   ├── Description
│   ├── Location
│   └── Recommendation
├── Recommendation[]
│   ├── Priority
│   ├── Action
│   └── Impact
└── RiskScore
    ├── Overall
    ├── Breakdown[]
    └── Trend
```

### Entities

| Entity | Attributes |
|---|---|
| SecurityScan | ScanId, Target, Type, Status, StartedAt, CompletedAt |
| Finding | FindingId, ScanId, Severity, Category, Description, Location |
| Threat | ThreatId, Type, Severity, Description, Indicators[] |

### Value Objects

| Value Object | Type |
|---|---|
| `ScanType` | Enum (code, infrastructure, dependency, configuration) |
| `Severity` | Enum (critical, high, medium, low, info) |
| `RiskScore` | Float (0.0 - 10.0) |

---

## 4. Public API Trait

```rust
// nexus-security-ai/src/interfaces/api.rs
#[async_trait]
pub trait SecurityService: Send + Sync {
    async fn analyze_security(&self, req: SecurityRequest) -> Result<SecurityReport, SecurityError>;
    async fn review_code(&self, req: CodeReviewRequest) -> Result<CodeReviewResponse, SecurityError>;
    async fn scan_infrastructure(&self, req: ScanRequest) -> Result<ScanResponse, SecurityError>;
}

pub struct SecurityRequest {
    pub target: String,
    pub scan_type: ScanType, // Code, Infrastructure, Dependency
    pub tenant_id: TenantId,
}

pub struct SecurityReport {
    pub risk_score: f32,
    pub findings: Vec<Finding>,
    pub recommendations: Vec<Recommendation>,
    pub summary: String,
}

pub struct CodeReviewRequest {
    pub code: String,
    pub language: String,
    pub context: Option<String>,
}

pub struct CodeReviewResponse {
    pub issues: Vec<Finding>,
    pub secure_alternative: String,
    pub risk_score: f32,
}
```

> **Note:** `SecurityService` is consumed by `nexus-ai-gateway` (prompt injection scanning) and `nexus-gateway` (API security checks) — all via in-process trait dispatch.

---

## 5. Security Analysis Capabilities

### 5.1 Code Security Review

| Check | Description |
|---|---|
| SQL Injection | Detect unsafe SQL construction |
| XSS Vulnerabilities | Cross-site scripting risks |
| Command Injection | OS command execution risks |
| Hardcoded Secrets | Passwords, API keys in code |
| Insecure Deserialization | Unsafe data parsing |
| Path Traversal | File system access risks |
| SSRF | Server-side request forgery |
| Authentication Bypass | Auth logic weaknesses |

### 5.2 Infrastructure Security

| Check | Description |
|---|---|
| Configuration Review | Docker, K8s, NATS config |
| Network Security | Open ports, exposed services |
| TLS Configuration | Certificate validity, cipher suites |
| Secrets Management | Vault usage, key rotation |
| Container Security | Image scanning, privilege checks |

### 5.3 Dependency Security

| Check | Description |
|---|---|
| Known CVEs | Vulnerable package versions |
| License Compliance | GPL, MIT, Apache licenses |
| Outdated Packages | End-of-life dependencies |
| Supply Chain Risk | Package source verification |

---

## 6. Threat Detection

### Threat Categories

| Category | Examples |
|---|---|
| Prompt Injection | "Ignore previous instructions" |
| Data Exfiltration | Unauthorized data access attempts |
| Privilege Escalation | Unauthorized permission requests |
| Denial of Service | Resource exhaustion attacks |
| Social Engineering | Manipulation attempts |

### Detection Pipeline

```
Input Analysis
    |
    v
[1] Pattern Matching
    |  - Known attack signatures
    |  - Regex-based detection
    |
    v
[2] LLM Analysis (WhiteRabbitNeo)
    |  - Context-aware threat detection
    |  - Novel attack identification
    |  - Risk assessment
    |
    v
[3] Behavioral Analysis
    |  - Anomaly detection
    |  - Baseline comparison
    |
    v
[4] Correlation
    |  - Cross-reference findings
    |  - Escalation rules
    |
    v
[5] Response
    |  - Block if critical
    |  - Alert if high
    |  - Log if medium/low
```

---

## 7. Prompt Injection Protection

### Detection Rules

| Pattern | Action |
|---|---|
| "Ignore previous instructions" | BLOCKED |
| "Show system prompt" | BLOCKED |
| "Reveal database password" | BLOCKED |
| "You are now..." | BLOCKED |
| "Act as..." (role hijack) | FLAGGED |
| Base64 encoded instructions | FLAGGED |

### Protection Pipeline

```
User Input
    |
    v
Input Security Scanner
    |  - Pattern detection
    |  - Encoding detection
    |  - Length analysis
    |
    v
Prompt Sanitizer
    |  - Remove injection patterns
    |  - Escape special characters
    |  - Truncate if oversized
    |
    v
AI Agent (with sanitized input)
    |
    v
Response Validator
    |  - Check for leaked system prompts
    |  - Check for PII in output
    |  - Check for harmful content
    |
    v
Safe Response
```

---

## 8. REST API Endpoints

### Security Scan

```
POST /api/v1/security/scan
```

**Request:**
```json
{
  "target": "repository://aeroxe/backend",
  "type": "code"
}
```

**Response:**
```json
{
  "scan_id": "uuid",
  "risk_score": 3.5,
  "findings": [
    {
      "severity": "HIGH",
      "category": "sql_injection",
      "description": "User input directly interpolated into SQL query",
      "location": "services/sql-agent/main.go:142",
      "recommendation": "Use parameterized queries instead of string concatenation"
    }
  ],
  "summary": "3 high, 5 medium, 2 low issues found"
}
```

### Code Review

```
POST /api/v1/security/review
```

---

## 9. NATS Events

| Subject | Event |
|---|---|
| `aeroxe.security.scan.started` | Scan initiated |
| `aeroxe.security.threat.detected` | Threat found |
| `aeroxe.security.report.generated` | Analysis complete |
| `aeroxe.security.alert.triggered` | Critical alert |

---

## 10. Observability

| Metric | Description |
|---|---|
| `security_scans_total` | Total scans by type |
| `security_findings_total` | Findings by severity |
| `security_threats_detected_total` | Threats detected |
| `security_scan_duration_ms` | Scan execution time |
| `security_injection_attempts_total` | Blocked injection attempts |
