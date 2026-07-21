# AeroXe Nexus AI — Security Architecture

## Zero Trust, RBAC, ABAC, AI Security & Data Protection (Modular Monolith)

> This document applies to the `aeroxe-nexus` modular monolith. Unlike microservice architectures, internal module-to-module calls do not use mTLS because they are in-process trait dispatches. Security focuses on: (1) external API security via `nexus-gateway`, (2) authentication via `nexus-identity`, (3) AI safety via `nexus-security-ai`, and (4) audit via `nexus-audit`.

---

## 1. Security Model

AeroXe Nexus AI uses a **Zero Trust Security Model**: "Never trust, always verify."

Every request must be authenticated, authorized, validated, and audited.

---

## 2. Security Layers

```
User / Application
       |
       v
API Gateway
       |
  ================================================
  Security Enforcement Layer (nexus-gateway)
  ================================================
  JWT Validation | RBAC/ABAC | Tenant Extraction
  Rate Limiting  | Input Validation | Audit
  ================================================
       |
       v
Internal Modules (in-process trait dispatch)
       |
  ================================================
  RequestContext Propagation | Permission Trait Calls
  NATS Subject Permissions | Database Security
  ================================================
       |
       v
Data Layer
  ================================================
  Encryption | Access Control | Audit Logging
  ================================================
```

---

## 3. Authentication

### JWT Token Architecture

**Supported Methods:**

| Method | Status |
|---|---|
| Email + Password | Active |
| OTP (Email/SMS) | Active |
| API Keys | Active |
| SSO | Future |
| OAuth2 | Future |

### JWT Token Structure

```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "roles": ["admin"],
  "permissions": ["ai.execute", "document.read"],
  "email": "admin@company.com",
  "exp": 1780003600,
  "iss": "aeroxe-nexus-ai"
}
```

### Token Management

| Token | Lifetime | Storage |
|---|---|---|
| Access Token | 1 hour | HTTP-only cookie / secure storage |
| Refresh Token | 7 days | Secure storage (one-time use) |

---

## 4. Authorization

### RBAC (Role-Based Access Control)

| Role | Description |
|---|---|
| SUPER_ADMIN | Platform-wide admin |
| TENANT_ADMIN | Tenant-level admin |
| AI_OPERATOR | AI management |
| DEVELOPER | Developer access |
| CUSTOMER_SUPPORT | Support staff |
| USER | Standard user |
| AUDITOR | Read-only audit |

### ABAC (Attribute-Based Access Control)

Authorization based on:

| Attribute Type | Examples |
|---|---|
| User Attributes | Department, clearance level, location |
| Resource Attributes | Document classification, data sensitivity |
| Environment Attributes | Time of day, IP range, device type |

### Authorization Flow

```
User Request -> JWT Validation -> Extract Claims
    -> RBAC Check -> ABAC Policy Engine -> Permission Granted/Denied
```

---

## 5. Multi-Tenant Isolation

### Strategy

Every request carries `tenant_id`. Enforcement:

1. JWT contains `tenant_id` (validated by `nexus-identity` trait call)
2. All database queries filter by `tenant_id`
3. `RequestContext` propagates `tenant_id` to all module trait calls
4. NATS events include `tenant_id`
5. Cross-tenant access returns `403 Forbidden`

### Database Isolation

**Phase 1 (Current):** Shared database + tenant column
```sql
SELECT * FROM documents WHERE tenant_id = $1;
```

**Phase 2 (Future):** Database per tenant for enterprise customers

---

## 6. Module-to-Module Security

> In the modular monolith, modules communicate via Rust trait methods **within the same process**. There is no network between modules, so mTLS is not needed internally.

### Security Enforcement

| Layer | Mechanism |
|---|---|
| API Boundary | `nexus-gateway` validates JWT + tenant before trait dispatch |
| Module Entry | Modules receive pre-validated `RequestContext` (no re-validation needed) |
| Permission Check | Modules call `nexus-identity::check_permission()` trait method |
| Tenant Isolation | All queries include `tenant_id` — enforced at database level |

### Request Context Propagation

```rust
// Every module trait method receives a context with authenticated claims
pub struct RequestContext {
    pub tenant_id: TenantId,
    pub user_id: UserId,
    pub roles: Vec<String>,
    pub permissions: Vec<String>,
    pub request_id: String,
    pub trace_id: String,
}
```

---

## 7. NATS JetStream Security

### Account Isolation

```
AI_ACCOUNT
  subjects: aeroxe.ai.*
```

### Subject Permissions

**Agent Service:**
```
publish:   aeroxe.agent.*
subscribe: aeroxe.rag.*
```

**RAG Service:**
```
subscribe: aeroxe.rag.document.*
publish:   aeroxe.rag.completed
```

---

## 8. API Gateway Security

| Responsibility | Implementation |
|---|---|
| Authentication | JWT validation |
| Authorization | RBAC + ABAC |
| Rate Limiting | Token Bucket (Redis) |
| Request Filtering | Input validation |
| DDoS Protection | Connection limits, IP blocking |
| API Logging | Request/response audit |

### Rate Limiting

| Tier | Limit |
|---|---|
| Free | 100 requests/min |
| Customer | 1,000 requests/min |
| Enterprise | 10,000 requests/min |

---

## 9. AI Security

### Prompt Injection Protection

```
User Input -> Input Security Scanner -> Prompt Sanitizer
    -> AI Agent -> Response Validator
```

**Blocked Patterns:**
- "Ignore previous instructions"
- "Show system prompt"
- "Reveal database password"

### AI Tool Security

Agents cannot directly execute tools:

```
Agent -> Tool Request -> Policy Engine -> Permission Check
    -> Tool Execution -> Result
```

Example: AI requesting "DELETE customer data" -> BLOCKED (requires human approval)

### SQL Agent Security

| Allowed | Blocked |
|---|---|
| SELECT, JOIN, GROUP BY | DELETE, UPDATE, DROP |
| ORDER BY, COUNT, SUM, AVG | ALTER, TRUNCATE, INSERT |

SQL flow: Question -> LLM -> SQL Validator -> Permission Check -> Read Replica -> Result

---

## 10. Data Protection

### Encryption at Rest

| Component | Algorithm |
|---|---|
| Database | AES-256 |
| MinIO Files | Server-side encryption |
| Backups | Encrypted |

### Encryption in Transit

| Component | Protocol |
|---|---|
| All External | TLS 1.3 |
| Internal Module Calls | In-process (no network) |
| NATS (optional) | TLS |
| Database | TLS |
| Optional gRPC (external) | TLS 1.3 + optional mTLS |

---

## 11. Secrets Management

Never store in code:
- Passwords
- API Keys
- JWT Secrets
- Database Credentials

**Solutions:**
- Hashicorp Vault
- Kubernetes Secrets
- Environment Variables (dev only)

---

## 12. Telephony Security (NEW)

### Voice Channel Security

| Requirement | Implementation |
|---|---|
| Call encryption | SRTP for RTP, TLS for SIP |
| Recording consent | Configurable per-tenant (required/optional) |
| PII in audio | Auto-redact credit cards, SSN in transcripts |
| DNC compliance | Check outbound against DNC list |
| Phone number validation | E.164 format enforcement |
| Rate limiting | Per-tenant concurrent call limits |
| Call recording access | RBAC on recording playback |
| STT data retention | Configurable transcript retention |
| TTS voice cloning | Tenant-scoped voice profiles |
| Webhook security | HMAC signature verification |

### Call Recording Compliance

| Regulation | Requirement |
|---|---|
| GDPR | Consent before recording, right to deletion |
| CCPA | Disclosure of recording, opt-out option |
| TCPA | DNC compliance for outbound calls |
| Industry-specific | Varies by sector |

### Audio Data Protection

| Component | Protection |
|---|---|
| Call audio (in-transit) | SRTP encryption |
| Call recordings (at-rest) | AES-256 encryption in MinIO |
| Transcripts (at-rest) | PostgreSQL TDE or column encryption |
| TTS voice profiles | Encrypted storage |
| DNC list | Access-controlled, audit-logged |

---

## 13. Security Monitoring

| Component | Purpose |
|---|---|
| OpenTelemetry | Distributed tracing |
| Prometheus | Security metrics |
| Grafana | Dashboards |
| Loki | Log aggregation |
| Elasticsearch | Security event search |

### Alert Triggers

| Alert | Severity |
|---|---|
| Multiple failed logins | HIGH |
| Prompt injection attempt | CRITICAL |
| Tenant isolation violation | CRITICAL |
| Unauthorized access | HIGH |
| Data exfiltration attempt | CRITICAL |

---

## 13. Security Testing

| Test Type | Tools |
|---|---|
| SAST | Semgrep, SonarQube, CodeQL |
| Dependency | Trivy, Dependabot, OWASP |
| Container | Trivy, Docker Bench |
| Penetration | Manual + automated |
| API Security | OWASP ZAP |

---

## 14. Backup Security

| Component | Security Measure |
|---|---|
| Database Backups | Encrypted, access-controlled |
| MinIO | Versioning, encryption, replication |
| NATS | Stream snapshots |

---

## 15. Disaster Recovery

| Metric | Target |
|---|---|
| RPO (Recovery Point Objective) | < 15 minutes |
| RTO (Recovery Time Objective) | < 2 hours |

---

## 16. Voice Security (NEW - CRITICAL)

### 16.1 Anti-Spoofing Measures

| Attack | Detection | Prevention |
|---|---|---|
| Caller ID Spoofing | ANI validation, carrier lookup | Require additional auth |
| Voice Clone Attack | Deepfake detection, liveness | Voice biometric verification |
| SIM Swap | Number portability check | Cross-reference account changes |
| Replay Attack | Audio nonce, timestamp | Reject non-real-time audio |
| Toll Fraud | Outbound pattern analysis | Rate limiting + monitoring |

### 16.2 Audio Security

| Layer | Protection |
|---|---|
| In-transit | SRTP encryption (AES-128) |
| At-rest | AES-256 encryption for recordings |
| STT processing | No audio storage after transcription |
| TTS output | Encrypted delivery to call session |

### 16.3 Voice Biometrics Security

| Requirement | Implementation |
|---|---|
| Template storage | Encrypted, isolated per tenant |
| Spoofing detection | Liveness verification |
| Template update | Re-enroll periodically |
| Consent | Explicit consent for biometric collection |
| Deletion | Delete templates on account closure |

---

## 17. Compliance Framework Mapping (NEW)

### 17.1 SOC 2 Controls

| Control | Implementation |
|---|---|
| CC6.1 Logical access | JWT + RBAC + ABAC |
| CC6.2 Authentication | Multi-factor auth |
| CC6.3 Authorization | Role-based permissions |
| CC6.6 Encryption | TLS 1.3, AES-256 |
| CC7.1 Monitoring | OpenTelemetry, audit logs |
| CC7.2 Anomaly detection | Fraud detection, sentiment alerts |
| CC8.1 Change management | Database migrations, code review |

### 17.2 GDPR Controls

| Control | Implementation |
|---|---|
| Art. 6 Lawful basis | Consent, contract, legitimate interest |
| Art. 12 Transparency | Clear privacy policy |
| Art. 15 Right of access | Data export API |
| Art. 17 Right to erasure | Deletion pipeline |
| Art. 20 Data portability | JSON export format |
| Art. 25 Privacy by design | Tenant isolation, encryption |
| Art. 32 Security | Encryption, access control |

### 17.3 HIPAA Controls (if healthcare)

| Control | Implementation |
|---|---|
| Access control | RBAC, minimum necessary |
| Audit controls | Comprehensive audit trail |
| Integrity controls | Data validation, checksums |
| Transmission security | TLS 1.3, SRTP |
| Business associates | BAAs with providers |
