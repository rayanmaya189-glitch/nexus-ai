# AeroXe Nexus AI — Identity & Authentication Service

## IAM, JWT, RBAC, ABAC, Multi-Tenant User Management

---

## 1. Service Identity

| Attribute | Value |
|---|---|
| Service Name | `identity-service` |
| Bounded Context | Identity |
| Domain Type | Supporting Domain |
| Language | Go |
| Database | `identity_db` (PostgreSQL) |
| gRPC Port | 50054 |

---

## 2. Purpose

The Identity Service is the foundation of platform security. It manages:

- User registration and authentication
- JWT token generation and validation
- Role-based access control (RBAC)
- Attribute-based access control (ABAC)
- Multi-tenant user isolation
- API key management
- Session management

---

## 3. Aggregate Design

### User Aggregate

```
User (Aggregate Root)
├── Profile
│   ├── Email
│   ├── DisplayName
│   └── Avatar
├── Authentication
│   ├── PasswordHash
│   ├── OTP Secret
│   └── MFA Settings
├── Roles[]
│   ├── RoleId
│   ├── Name
│   └── Permissions[]
└── TenantMembership
    ├── TenantId
    └── TenantRole
```

### Entities

| Entity | Attributes |
|---|---|
| User | UserId, TenantId, Email, PasswordHash, Status, CreatedAt |
| Role | RoleId, TenantId, Name, Description, Permissions[] |
| Permission | PermissionId, Name, Resource, Action |
| Tenant | TenantId, Name, Plan, Settings, Status |
| APIKey | KeyId, TenantId, UserId, KeyHash, Scopes[], ExpiresAt |

### Value Objects

| Value Object | Type |
|---|---|
| `UserId` | i64 |
| `TenantId` | i64 |
| `EmailAddress` | Validated string |
| `PasswordHash` | bcrypt hash |
| `JWTToken` | Signed JWT |
| `Permission` | Resource + Action pair |

---

## 4. Authentication Architecture

### Supported Methods

| Method | Status | Description |
|---|---|---|
| Email + Password | Active | Primary authentication |
| OTP (Email/SMS) | Active | Two-factor authentication |
| SSO | Future | Single sign-on integration |
| OAuth2 | Future | External identity providers |

### Login Flow

```
User
    |
    v
POST /api/v1/auth/login
    |
    v
Identity Service
    |
    ├── Validate email exists
    ├── Verify password (bcrypt)
    ├── Check account status
    ├── Generate JWT access token
    ├── Generate refresh token
    |
    v
Return tokens to client
    |
    v
Frontend stores session
```

### JWT Token Structure

```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "roles": ["admin", "ai_operator"],
  "permissions": ["ai.execute", "document.read", "customer.read"],
  "email": "admin@company.com",
  "iat": 1780000000,
  "exp": 1780003600,
  "iss": "aeroxe-nexus-ai"
}
```

### Token Lifecycle

| Token | Lifetime | Refreshable |
|---|---|---|
| Access Token | 1 hour | No |
| Refresh Token | 7 days | Yes (one-time use) |

---

## 5. Authorization Model

### 5.1 RBAC (Role-Based Access Control)

**Predefined Roles:**

| Role | Description | Permissions |
|---|---|---|
| `SUPER_ADMIN` | Platform-wide admin | All permissions |
| `TENANT_ADMIN` | Tenant-level admin | All tenant permissions |
| `AI_OPERATOR` | AI management | ai.*, document.*, knowledge.* |
| `DEVELOPER` | Developer access | code.*, api.*, test.* |
| `CUSTOMER_SUPPORT` | Support staff | customer.*, ticket.*, knowledge.search |
| `USER` | Standard user | chat.*, document.read |
| `AUDITOR` | Read-only audit | audit.*, report.read |

### Permission Model

Each permission follows: `{resource}.{action}`

```
ai.execute          - Execute AI requests
ai.manage           - Manage AI agents
document.read       - Read documents
document.write      - Upload documents
document.delete     - Delete documents
customer.read       - View customer data
customer.write      - Modify customer data
ticket.create       - Create support tickets
knowledge.search    - Search knowledge base
billing.read        - View billing data
code.read           - View source code
code.write          - Modify source code
audit.read          - View audit logs
admin.manage        - Platform administration
```

### 5.2 ABAC (Attribute-Based Access Control)

ABAC adds context-aware authorization based on:

**User Attributes:**
- Department (support, engineering, finance, executive)
- Clearance level (standard, privileged, restricted)
- Location

**Resource Attributes:**
- Document classification (public, internal, confidential, restricted)
- Data sensitivity (low, medium, high, critical)
- Owner tenant

**Environment Attributes:**
- Time of day
- IP address range
- Device type

**Example Policy:**
```
IF user.department == "support"
AND document.classification == "financial"
THEN DENY
```

### Authorization Flow

```
User Request
    |
    v
JWT Validation (verify signature, expiry)
    |
    v
Extract Claims (user_id, tenant_id, roles, permissions)
    |
    v
RBAC Check (does role have required permission?)
    |
    v
ABAC Policy Engine (context-aware rules)
    |
    v
Permission Granted / Denied
```

---

## 6. Multi-Tenant Architecture

### Tenant Isolation Strategy

**Phase 1: Shared Database + Tenant Column**

Every business table includes `tenant_id`:

```sql
SELECT * FROM documents WHERE tenant_id = $1;
```

**Phase 2 (Future): Database per Tenant**

For enterprise customers requiring complete isolation:
```
tenant_a_database
tenant_b_database
tenant_c_database
```

### Tenant Examples

| Tenant | Type |
|---|---|
| AeroXe Broadband | Internal |
| AeroXe ERP Customer A | SaaS |
| AeroXe Enterprise Customer B | Enterprise |

---

## 7. gRPC Contract

```protobuf
syntax = "proto3";
package aeroxe.identity;

service IdentityService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc Authenticate(AuthRequest) returns (AuthResponse);
  rpc CheckPermission(PermissionRequest) returns (PermissionResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc AssignRole(AssignRoleRequest) returns (AssignRoleResponse);
}

message CreateUserRequest {
  string tenant_id = 1;
  string email = 2;
  string password = 3;
  repeated string roles = 4;
}

message AuthRequest {
  string email = 1;
  string password = 2;
  string tenant_id = 3;
}

message AuthResponse {
  string access_token = 1;
  string refresh_token = 2;
  int64 expires_in = 3;
  User user = 4;
}

message PermissionRequest {
  string user_id = 1;
  string resource = 2;
  string action = 3;
  string tenant_id = 4;
}

message PermissionResponse {
  bool allowed = 1;
  string reason = 2;
}
```

---

## 8. Database Schema (identity_db)

### users

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret TEXT,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### roles

```sql
CREATE TABLE roles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### permissions

```sql
CREATE TABLE permissions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT
);
```

### user_roles

```sql
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by BIGINT,
    PRIMARY KEY(user_id, role_id)
);
```

### role_permissions

```sql
CREATE TABLE role_permissions (
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY(role_id, permission_id)
);
```

### tenants

```sql
CREATE TABLE tenants (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    settings JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### api_keys

```sql
CREATE TABLE api_keys (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    key_hash TEXT NOT NULL,
    key_prefix VARCHAR(10) NOT NULL,
    scopes TEXT[],
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 9. REST API Endpoints

### Login

```
POST /api/v1/auth/login
```

**Request:**
```json
{
  "email": "admin@company.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJl...",
  "expires_in": 3600,
  "user": {
    "id": "uuid",
    "email": "admin@company.com",
    "roles": ["admin"]
  }
}
```

### Refresh Token

```
POST /api/v1/auth/refresh
```

### Register User (Admin Only)

```
POST /api/v1/auth/register
```

### Get Current User

```
GET /api/v1/auth/me
```

### Change Password

```
POST /api/v1/auth/change-password
```

---

## 10. Security Requirements

| Requirement | Implementation |
|---|---|
| Password Hashing | bcrypt (cost factor 12) |
| JWT Signing | HS256 (symmetric) or RS256 (asymmetric) |
| Token Storage | HTTP-only secure cookies or secure storage |
| Rate Limiting | 5 failed login attempts -> 15 min lockout |
| Password Policy | Min 12 chars, mixed case, numbers, symbols |
| Session Management | Invalidate on password change |
| API Key Security | Hashed storage, prefix for identification |

---

## 11. NATS Events

### Published

| Subject | Event |
|---|---|
| `aeroxe.identity.user.created` | `UserCreated` |
| `aeroxe.identity.user.updated` | `UserUpdated` |
| `aeroxe.identity.role.assigned` | `RoleAssigned` |
| `aeroxe.identity.permission.changed` | `PermissionChanged` |

---

## 12. Observability

| Metric | Description |
|---|---|
| `auth_login_attempts_total` | Login attempts by result |
| `auth_login_duration_ms` | Login latency |
| `auth_tokens_issued_total` | Tokens generated |
| `auth_permission_checks_total` | Authorization checks |
| `auth_rate_limit_hits_total` | Rate limit rejections |
