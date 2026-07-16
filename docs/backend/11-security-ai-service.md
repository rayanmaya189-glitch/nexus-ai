# AeroXe Nexus AI — Security AI Service

## AI-Powered Security Analysis, Vulnerability Detection & Threat Intelligence

---

## 1. Service Identity

| Attribute | Value |
|---|---|
| Service Name | `security-ai-service` |
| Bounded Context | Security Intelligence |
| Domain Type | Core Domain |
| Language | Rust |
| AI Model | `whiterabbitneo:7b` (Ollama) |
| gRPC Port | 50059 |

---

## 2. Purpose

The Security AI Service provides AI-powered security intelligence:

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

## 4. gRPC Contract

```protobuf
syntax = "proto3";
package aeroxe.security;

service SecurityService {
  rpc AnalyzeSecurity(SecurityRequest) returns (SecurityReport);
  rpc ReviewCode(CodeReviewRequest) returns (CodeReviewResponse);
  rpc ScanInfrastructure(ScanRequest) returns (ScanResponse);
}

message SecurityRequest {
  string target = 1;
  string type = 2; // "code", "infrastructure", "dependency"
  string tenant_id = 3;
}

message SecurityReport {
  float risk_score = 1;
  repeated Finding findings = 2;
  repeated Recommendation recommendations = 3;
  string summary = 4;
}

message Finding {
  string severity = 1;
  string category = 2;
  string description = 3;
  string location = 4;
  string recommendation = 5;
}

message CodeReviewRequest {
  string code = 1;
  string language = 2;
  string context = 3;
}

message CodeReviewResponse {
  repeated Finding issues = 1;
  string secure_alternative = 2;
  float risk_score = 3;
}
```

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
