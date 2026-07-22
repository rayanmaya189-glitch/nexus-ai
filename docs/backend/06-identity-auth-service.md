# AeroXe Nexus AI — Identity & Authentication Module

## IAM, JWT, RBAC, ABAC, Multi-Tenant User Management

> **Modular Monolith Module:** This document describes the `identity` module at `src/modules/identity/`. It communicates with other modules via Rust trait interfaces (see [Communication Architecture](12-communication-architecture.md)).

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `identity` |
| Module Path | `src/modules/identity/` |
| Bounded Context | Identity |
| Domain Type | Supporting Domain |
| Language | Rust |
| ORM | SeaORM (no raw SQL) |
| Schema | `identity_` (in shared PostgreSQL) |
| Dependencies | Redis (sessions, rate limiting) |

---

## 2. Folder Structure

```
src/modules/identity/
├── mod.rs                    # Public re-exports + IdentityService trait
├── domain/
│   ├── mod.rs
│   ├── aggregates/
│   │   └── user/
│   │       ├── user.rs       # User aggregate root
│   │       └── tests/        # Domain tests for User aggregate
│   ├── entities/
│   │   └── session.rs        # Session entity
│   ├── value_objects/
│   │   ├── email.rs          # EmailAddress value object
│   │   └── password.rs       # Password value object (hashing)
│   └── rules/
│       └── auth_rules.rs     # Authentication business rules
├── application/
│   ├── mod.rs
│   ├── commands/
│   │   ├── login.rs          # Login command
│   │   └── tests/            # Command handler tests (mock repo)
│   ├── queries/
│   │   └── get_user.rs       # GetUser query
│   └── services/
│       └── auth_service.rs   # Auth orchestration service
├── infrastructure/
│   ├── mod.rs
│   ├── repository/
│   │   └── postgres_user_repository.rs  # SeaORM implementation
│   └── security/
│       └── jwt.rs            # JWT generation/validation
├── api/
│   ├── mod.rs
│   ├── http/
│   │   ├── auth_controller.rs  # Axum HTTP handlers
│   │   └── tests/              # API integration tests
│   └── external/
│       └── auth_grpc.rs        # External gRPC adapter (tonic, optional)
└── migrations/               # SeaORM migration files
```

---

## 3. Purpose

The Identity module is the foundation of platform security. It manages:

- User registration and authentication
- JWT token generation and validation
- Role-based access control (RBAC)
- Attribute-based access control (ABAC)
- Multi-tenant user isolation
- API key management
- Session management (Session entity)
- KYC document management

---

## 4. Aggregate Design

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
| **Session** | SessionId, UserId, Token, RefreshToken, ExpiresAt, CreatedAt |

### Value Objects

| Value Object | Type | Validation |
|---|---|---|
| `UserId` | `i64` | Positive |
| `TenantId` | `i64` | Positive |
| `EmailAddress` | `String` | Regex validated (`validator` crate) |
| `PasswordHash` | `String` | bcrypt verified (`password` module) |
| `JWTToken` | `String` | RS256 signed + exp check |
| `Permission` | `String` | Format: `{resource}.{action}` |

---

## 5. Authentication Architecture

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
Identity Module (api/http/auth_controller.rs)
    |
    v
AuthService (application/services/auth_service.rs)
    |
    ├── Validate email exists (SeaORM query → postgres_user_repository)
    ├── Verify password (bcrypt → domain/value_objects/password.rs)
    ├── Check account status (domain/rules/auth_rules.rs)
    ├── Generate JWT access token (infrastructure/security/jwt.rs)
    ├── Generate refresh token
    ├── Create Session entity
    |
    v
Return tokens to client
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

## 6. Authorization Model

### 6.1 RBAC (Role-Based Access Control)

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

### 6.2 ABAC (Attribute-Based Access Control)

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

## 7. Multi-Tenant Architecture

### Tenant Isolation Strategy

**Phase 1: Shared Database + Tenant Column**

Every business table includes `tenant_id`. All access goes through SeaORM filters:

```rust
// SeaORM query with tenant isolation (no raw SQL)
let users = Entity::find()
    .filter(ModelColumn::TenantId.eq(tenant_id))
    .all(&self.db)
    .await?;
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

## 8. Public API Trait

```rust
// src/modules/identity/api/mod.rs
#[async_trait]
pub trait IdentityService: Send + Sync {
    async fn authenticate(&self, req: AuthRequest) -> Result<AuthResponse, IdentityError>;
    async fn verify_token(&self, token: &str) -> Result<JWTClaims, IdentityError>;
    async fn check_permission(&self, req: PermissionRequest) -> Result<bool, IdentityError>;
    async fn validate_tenant(&self, tenant_id: TenantId) -> Result<Tenant, IdentityError>;
    async fn create_user(&self, req: CreateUserRequest) -> Result<User, IdentityError>;
    async fn get_user(&self, id: UserId) -> Result<Option<User>, IdentityError>;
    async fn assign_role(&self, req: AssignRoleRequest) -> Result<(), IdentityError>;
}

pub struct AuthRequest {
    pub email: String,
    pub password: String,
    pub tenant_id: TenantId,
}

pub struct AuthResponse {
    pub access_token: String,
    pub refresh_token: String,
    pub expires_in: i64,
    pub user: User,
}

pub struct PermissionRequest {
    pub user_id: UserId,
    pub resource: String,
    pub action: String,
    pub tenant_id: TenantId,
}
```

> **Note:** `IdentityService` is consumed by `gateway` middleware (auth, tenant validation) and all other modules (permission checks) — all via in-process trait dispatch.

---

## 9. SeaORM Entities (No Raw SQL)

### users (SeaORM Entity)

```rust
// src/modules/identity/infrastructure/repository/postgres_user_repository.rs
use sea_orm::entity::prelude::*;

#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "users", schema_name = "identity")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: i64,
    #[sea_orm(unique)]
    pub email: String,
    pub password_hash: String,
    pub display_name: Option<String>,
    pub status: String,
    pub mfa_enabled: bool,
    pub mfa_secret: Option<String>,
    pub last_login_at: Option<chrono::NaiveDateTime>,
    pub created_at: chrono::NaiveDateTime,
    pub updated_at: chrono::NaiveDateTime,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(has_many = "super::user_roles::Entity")]
    UserRoles,
    #[sea_orm(has_many = "super::api_keys::Entity")]
    ApiKeys,
    #[sea_orm(has_many = "super::sessions::Entity")]
    Sessions,
}

impl ActiveModelBehavior for ActiveModel {}
```

### Repository Implementation

```rust
#[async_trait]
pub trait UserRepository: Send + Sync {
    async fn find_by_id(&self, id: i64) -> Result<Option<User>, RepositoryError>;
    async fn find_by_email(&self, email: &str) -> Result<Option<User>, RepositoryError>;
    async fn find_by_tenant(&self, tenant_id: i64) -> Result<Vec<User>, RepositoryError>;
    async fn save(&self, user: &User) -> Result<User, RepositoryError>;
    async fn update(&self, user: &User) -> Result<User, RepositoryError>;
}

pub struct PostgresUserRepository {
    db: DatabaseConnection,
}

#[async_trait]
impl UserRepository for PostgresUserRepository {
    async fn find_by_email(&self, email: &str) -> Result<Option<User>, RepositoryError> {
        let model = Entity::find()
            .filter(ModelColumn::Email.eq(email))
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        
        Ok(model.map(User::from))
    }
    
    async fn save(&self, user: &User) -> Result<User, RepositoryError> {
        let active = ActiveModel {
            tenant_id: Set(user.tenant_id),
            email: Set(user.email.clone()),
            password_hash: Set(user.password_hash.clone()),
            display_name: Set(user.display_name.clone()),
            status: Set(user.status.clone()),
            ..Default::default()
        };
        
        let result = active.insert(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        
        Ok(User::from(result))
    }
}
```

---

## 10. REST API Endpoints

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

## 11. External gRPC Adapter (Versioned, Optional)

```protobuf
// proto/identity/v1/auth_service.proto
package identity.v1;

service AuthService {
    rpc Authenticate(AuthRequest) returns (AuthResponse);
    rpc VerifyToken(VerifyTokenRequest) returns (JWTClaims);
    rpc CheckPermission(PermissionRequest) returns (PermissionResponse);
    rpc ValidateTenant(ValidateTenantRequest) returns (Tenant);
    rpc CreateUser(CreateUserRequest) returns (User);
    rpc GetUser(GetUserRequest) returns (User);
    rpc AssignRole(AssignRoleRequest) returns (Empty);
}
```

Service version is embedded in the package name (`identity.v1`).

---

## 12. Security Requirements

| Requirement | Implementation |
|---|---|
| Password Hashing | bcrypt (cost factor 12) |
| JWT Signing | HS256 (symmetric) or RS256 (asymmetric) |
| Token Storage | HTTP-only secure cookies or secure storage |
| Rate Limiting | 5 failed login attempts -> 15 min lockout |
| Password Policy | Min 12 chars, mixed case, numbers, symbols |
| Session Management | Invalidate on password change |
| API Key Security | Hashed storage, prefix for identification |
| **No Raw SQL** | All DB access through SeaORM entities |

---

## 13. NATS Events (Versioned Subjects)

### Published

| Subject | Event |
|---|---|
| `aeroxe.v1.identity.user.created` | `UserCreated` |
| `aeroxe.v1.identity.user.updated` | `UserUpdated` |
| `aeroxe.v1.identity.role.assigned` | `RoleAssigned` |
| `aeroxe.v1.identity.permission.changed` | `PermissionChanged` |

All NATS subjects follow the `aeroxe.v1.<module>.<event>` format to enable future version coexistence.

---

## 14. Observability

> **Note:** As a module within the monolith, identity metrics are collected via shared OpenTelemetry tracing spans in the binary, not via a separate metrics endpoint.

| Metric | Description |
|---|---|
| `auth_login_attempts_total` | Login attempts by result |
| `auth_login_duration_ms` | Login latency |
| `auth_tokens_issued_total` | Tokens generated |
| `auth_permission_checks_total` | Authorization checks |
| `auth_rate_limit_hits_total` | Rate limit rejections |
