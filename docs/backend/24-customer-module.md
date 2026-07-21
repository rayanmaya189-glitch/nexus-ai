# AeroXe Nexus AI — Customer Module

## Customer Aggregate, Profiles, Status Management & Addresses

> **Modular Monolith Module:** This document describes the `customer` module at `src/modules/customer/`. It communicates with other modules via Rust trait interfaces (see [Communication Architecture](12-communication-architecture.md)). All database access uses SeaORM — no raw SQL.

---

## 1. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `customer` |
| Module Path | `src/modules/customer/` |
| Bounded Context | Customer Management |
| Domain Type | Supporting Domain |
| Language | Rust (edition 2024) |
| ORM | SeaORM (no raw SQL) |
| Schema | `customer_` (in shared PostgreSQL) |
| Dependencies | NATS (event publishing), PostgreSQL (persistence) |

---

## 2. Folder Structure

```
src/modules/customer/
├── mod.rs                        # Public re-exports + CustomerService trait
├── domain/
│   ├── mod.rs
│   ├── aggregates/
│   │   └── customer/
│   │       ├── customer.rs       # Customer aggregate root
│   │       ├── profile.rs        # Profile entity
│   │       ├── status.rs         # Status value object / enum
│   │       └── tests/
│   │           └── customer_tests.rs  # Domain tests for Customer aggregate
│   ├── value_objects/
│   │   ├── customer_id.rs        # CustomerId value object
│   │   ├── email.rs              # Email value object (validated)
│   │   └── phone.rs              # Phone value object (validated)
│   └── rules/
│       └── customer_rules.rs     # Customer business rules (status transitions)
├── application/
│   ├── mod.rs
│   ├── commands/
│   │   ├── create_customer.rs    # CreateCustomer command
│   │   ├── suspend_customer.rs   # SuspendCustomer command
│   │   ├── activate_customer.rs  # ActivateCustomer command
│   │   └── tests/                # Handler tests (mock repo)
│   ├── queries/
│   │   └── get_customer.rs       # GetCustomer query
│   └── services/
│       └── customer_service.rs   # Customer orchestration service
├── infrastructure/
│   ├── mod.rs
│   ├── repository/
│   │   └── postgres_customer_repository.rs  # SeaORM repository
│   └── messaging/
│       ├── publishers/
│       │   └── customer_event_publisher.rs  # NATS event publisher
│       └── subscribers/
│           └── payment_event_subscriber.rs  # NATS event subscriber
├── api/
│   ├── mod.rs
│   ├── http/
│   │   ├── customer_controller.rs  # Axum HTTP handlers
│   │   └── tests/                  # Endpoint tests
│   └── grpc/
│       └── customer_service.rs     # gRPC service (tonic)
└── migrations/                   # SeaORM migration files
```

---

## 3. Purpose

The Customer module manages the lifecycle of customers across the AeroXe ecosystem. It handles:

- Customer creation and profile management
- Customer status management (active, suspended, inactive, archived)
- Address management (billing, shipping, physical)
- Customer search and querying
- Customer lifecycle events (via NATS)
- Integration with payment workflows (via NATS subscription)

---

## 4. Aggregate Design

### Customer Aggregate

```
Customer (Aggregate Root)
├── Profile
│   ├── CustomerId (value object)
│   ├── Name
│   ├── EmailAddress (value object)
│   ├── PhoneNumber (value object)
│   └── Addresses[] (entities)
├── Status (value object / enum)
│   ├── Active
│   ├── Suspended
│   ├── Inactive
│   └── Archived
├── KYC
│   ├── KYC Status
│   ├── Document References
│   └── Verification Date
└── Metadata
    ├── Tags[]
    ├── CustomFields (JSON)
    └── Notes[]
```

### Entities

| Entity | Attributes |
|---|---|
| Customer | CustomerId, TenantId, Name, Email, Phone, Status, Tags, CustomFields, Notes, CreatedAt, UpdatedAt |
| Profile | ProfileId, CustomerId, Avatar, Department, JobTitle, Language, Timezone |
| Address | AddressId, CustomerId, Type, Line1, Line2, City, State, PostalCode, Country, IsDefault |

### Value Objects

| Value Object | Rust Type | Validation |
|---|---|---|
| `CustomerId` | `i64` | Positive |
| `Email` | `String` | Regex validated (`validator` crate) |
| `Phone` | `String` | E.164 format validated |
| `CustomerStatus` | Enum | Active, Suspended, Inactive, Archived |

### Status Transitions

```
CREATED → ACTIVE → SUSPENDED → ACTIVE (re-activated)
                   → INACTIVE → ARCHIVED
```

| Transition | Rule |
|---|---|
| CREATED → ACTIVE | Default on creation |
| ACTIVE → SUSPENDED | Requires admin permission, reason recorded |
| ACTIVE → INACTIVE | No activity for 90 days |
| SUSPENDED → ACTIVE | Requires admin action, all payments cleared |
| INACTIVE → ARCHIVED | No activity for 365 days |
| ARCHIVED | Final state, cannot transition out |

---

## 5. SeaORM Entities (No Raw SQL)

### Customer Entity

```rust
// src/modules/customer/infrastructure/repository/customer_entity.rs
use sea_orm::entity::prelude::*;

#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "customers", schema_name = "customer")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub tenant_id: i64,
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
    pub status: String,           // active, suspended, inactive, archived
    pub kyc_status: String,       // pending, verified, rejected
    pub tags: Option<Json>,
    pub custom_fields: Option<Json>,
    pub notes: Option<Json>,
    pub created_at: DateTime,
    pub updated_at: DateTime,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(has_many = "super::addresses::Entity")]
    Addresses,
}

impl ActiveModelBehavior for ActiveModel {}
```

### Address Entity

```rust
#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "addresses", schema_name = "customer")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub customer_id: i64,
    pub address_type: String,     // billing, shipping, physical
    pub line1: String,
    pub line2: Option<String>,
    pub city: String,
    pub state: Option<String>,
    pub postal_code: String,
    pub country: String,
    pub is_default: bool,
    pub created_at: DateTime,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(
        belongs_to = "super::customers::Entity",
        from = "Column::CustomerId",
        to = "super::customers::Column::Id"
    )]
    Customer,
}

impl ActiveModelBehavior for ActiveModel {}
```

---

## 6. Repository Pattern

```rust
// src/modules/customer/infrastructure/repository/postgres_customer_repository.rs
#[async_trait]
pub trait CustomerRepository: Send + Sync {
    async fn find_by_id(&self, id: i64, tenant_id: i64) -> Result<Option<Customer>, RepositoryError>;
    async fn find_by_email(&self, email: &str, tenant_id: i64) -> Result<Option<Customer>, RepositoryError>;
    async fn find_by_tenant(&self, tenant_id: i64, page: u32, per_page: u32) -> Result<Vec<Customer>, RepositoryError>;
    async fn search(&self, query: CustomerSearchQuery) -> Result<Vec<Customer>, RepositoryError>;
    async fn save(&self, customer: &Customer) -> Result<Customer, RepositoryError>;
    async fn update(&self, customer: &Customer) -> Result<Customer, RepositoryError>;
    async fn update_status(&self, id: i64, tenant_id: i64, status: &str) -> Result<(), RepositoryError>;
}

pub struct PostgresCustomerRepository {
    db: DatabaseConnection,
}

#[async_trait]
impl CustomerRepository for PostgresCustomerRepository {
    async fn find_by_id(&self, id: i64, tenant_id: i64) -> Result<Option<Customer>, RepositoryError> {
        let model = customers::Entity::find_by_id(id)
            .filter(customers::Column::TenantId.eq(tenant_id))
            .one(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        
        if let Some(customer) = model {
            let addresses = addresses::Entity::find()
                .filter(addresses::Column::CustomerId.eq(customer.id))
                .all(&self.db)
                .await
                .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
            
            Ok(Some(Customer::from((customer, addresses))))
        } else {
            Ok(None)
        }
    }

    async fn save(&self, customer: &Customer) -> Result<Customer, RepositoryError> {
        let active = customers::ActiveModel {
            tenant_id: Set(customer.tenant_id),
            name: Set(customer.name.clone()),
            email: Set(customer.email.clone()),
            phone: Set(customer.phone.clone()),
            status: Set(customer.status.to_string()),
            tags: Set(customer.tags.clone().map(Json)),
            custom_fields: Set(customer.custom_fields.clone().map(Json)),
            ..Default::default()
        };
        let result = active.insert(&self.db)
            .await
            .map_err(|e| RepositoryError::DatabaseError(e.to_string()))?;
        Ok(Customer::from(result))
    }
}
```

---

## 7. Public API Trait

```rust
// src/modules/customer/api/mod.rs
#[async_trait]
pub trait CustomerService: Send + Sync {
    async fn create_customer(&self, req: CreateCustomerRequest) -> Result<Customer, CustomerError>;
    async fn get_customer(&self, id: CustomerId, tenant_id: TenantId) -> Result<Option<Customer>, CustomerError>;
    async fn list_customers(&self, query: CustomerListQuery) -> Result<PaginatedCustomers, CustomerError>;
    async fn search_customers(&self, query: CustomerSearchQuery) -> Result<Vec<Customer>, CustomerError>;
    async fn suspend_customer(&self, id: CustomerId, tenant_id: TenantId, reason: String) -> Result<Customer, CustomerError>;
    async fn activate_customer(&self, id: CustomerId, tenant_id: TenantId) -> Result<Customer, CustomerError>;
    async fn update_customer(&self, req: UpdateCustomerRequest) -> Result<Customer, CustomerError>;
    async fn delete_customer(&self, id: CustomerId, tenant_id: TenantId) -> Result<(), CustomerError>;
}
```

---

## 8. Commands & Queries

### Commands

| Command | Handler | Description |
|---|---|---|
| `CreateCustomerCommand` | `CreateCustomerHandler` | Creates a new customer with profile and addresses |
| `SuspendCustomerCommand` | `SuspendCustomerHandler` | Suspends a customer with reason |
| `ActivateCustomerCommand` | `ActivateCustomerHandler` | Re-activates a suspended customer |
| `UpdateCustomerCommand` | `UpdateCustomerHandler` | Updates customer profile and metadata |

### Queries

| Query | Handler | Description |
|---|---|---|
| `GetCustomerQuery` | `GetCustomerHandler` | Retrieves a single customer by ID |
| `ListCustomersQuery` | `ListCustomersHandler` | Paginated customer list filtered by tenant |
| `SearchCustomersQuery` | `SearchCustomersHandler` | Full-text search across customer fields |

---

## 9. REST API Endpoints

### Create Customer

```
POST /api/v1/customers
Authorization: Bearer <jwt>
```

### Get Customer

```
GET /api/v1/customers/{id}
Authorization: Bearer <jwt>
```

### List Customers

```
GET /api/v1/customers?page=1&per_page=20&status=active
Authorization: Bearer <jwt>
```

### Suspend Customer

```
POST /api/v1/customers/{id}/suspend
Authorization: Bearer <jwt>
```

### Activate Customer

```
POST /api/v1/customers/{id}/activate
Authorization: Bearer <jwt>
```

### Update Customer

```
PUT /api/v1/customers/{id}
Authorization: Bearer <jwt>
```

### Delete Customer (Soft)

```
DELETE /api/v1/customers/{id}
Authorization: Bearer <jwt>
```

---

## 10. gRPC Service (Versioned)

```protobuf
// proto/customer/v1/customer_service.proto
package customer.v1;

service CustomerService {
    rpc CreateCustomer(CreateCustomerRequest) returns (Customer);
    rpc GetCustomer(GetCustomerRequest) returns (Customer);
    rpc ListCustomers(ListCustomersRequest) returns (ListCustomersResponse);
    rpc SearchCustomers(SearchCustomersRequest) returns (SearchCustomersResponse);
    rpc SuspendCustomer(SuspendCustomerRequest) returns (Customer);
    rpc ActivateCustomer(ActivateCustomerRequest) returns (Customer);
    rpc UpdateCustomer(UpdateCustomerRequest) returns (Customer);
    rpc DeleteCustomer(DeleteCustomerRequest) returns (Empty);
}
```

Service version is embedded in the package name (`customer.v1`).

---

## 11. NATS Events (Versioned Subjects)

### Published Events

| Subject | Event | Trigger |
|---|---|---|
| `aeroxe.v1.customer.customer.created` | `CustomerCreated` | Customer created |
| `aeroxe.v1.customer.customer.activated` | `CustomerActivated` | Customer activated |
| `aeroxe.v1.customer.customer.suspended` | `CustomerSuspended` | Customer suspended |
| `aeroxe.v1.customer.customer.updated` | `CustomerUpdated` | Customer profile updated |

### Subscribed Events

| Subject | Handler | Purpose |
|---|---|---|
| `aeroxe.v1.payment.customer.*` | `PaymentEventSubscriber` | Listen for payment events affecting customer status |

### Event Envelope

```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "CustomerCreated",
  "version": "1.0",
  "api_version": "v1",
  "timestamp": "2026-07-21T12:00:00Z",
  "tenant_id": 1,
  "module": "customer",
  "data": {
    "customer_id": 1,
    "name": "Acme Corp",
    "email": "contact@acme.com",
    "status": "active"
  }
}
```

---

## 12. Business Rules

| Rule | Description | Location |
|---|---|---|
| Unique email per tenant | No duplicate customer emails within a tenant | `domain/rules/customer_rules.rs` |
| Valid status transitions | Only predefined transitions allowed | `domain/aggregates/customer/status.rs` |
| Address validation | At least one address required on creation | `domain/rules/customer_rules.rs` |
| Tenant isolation | Customers only visible within their tenant | `infrastructure/repository/` |
| Soft delete | Customers are archived, never hard-deleted | `application/commands/` |

---

## 13. TDD Contract — Customer Module Tests

### Unit Tests (Domain Logic)

```rust
#[tokio::test]
async fn test_customer_creation_with_valid_email() {
    let customer = Customer::new(CreateCustomerRequest {
        name: "Acme Corp".to_string(),
        email: "contact@acme.com".to_string(),
        tenant_id: 1,
        ..Default::default()
    });
    assert!(customer.is_ok());
    assert_eq!(customer.unwrap().status, CustomerStatus::Active);
}

#[tokio::test]
async fn test_customer_creation_with_invalid_email_fails() {
    let customer = Customer::new(CreateCustomerRequest {
        email: "invalid-email".to_string(),
        ..Default::default()
    });
    assert!(matches!(customer, Err(CustomerError::InvalidEmail)));
}

#[tokio::test]
async fn test_suspend_active_customer() {
    let mut customer = Customer::new(valid_request()).unwrap();
    let result = customer.suspend("Non-payment".to_string());
    assert!(result.is_ok());
    assert_eq!(customer.status, CustomerStatus::Suspended);
}

#[tokio::test]
async fn test_suspend_already_suspended_customer_fails() {
    let mut customer = Customer::new(valid_request()).unwrap();
    customer.suspend("reason".to_string()).unwrap();
    let result = customer.suspend("another reason".to_string());
    assert!(matches!(result, Err(CustomerError::InvalidTransition)));
}
```

### Integration Tests (SeaORM Repository)

```rust
#[tokio::test]
async fn test_customer_repository_save_and_find() {
    let db = test_db().await;  // Creates test PostgreSQL with SeaORM
    
    let repo = PostgresCustomerRepository::new(db);
    let customer = Customer::new(valid_request()).unwrap();
    
    let saved = repo.save(&customer).await.unwrap();
    let found = repo.find_by_id(saved.id, saved.tenant_id).await.unwrap();
    
    assert!(found.is_some());
    assert_eq!(found.unwrap().email, customer.email);
}
```

### API Contract Tests

```rust
#[tokio::test]
async fn test_create_customer_endpoint_v1() {
    let mut mock_service = MockCustomerService::new();
    mock_service.expect_create_customer()
        .returning(|_| Ok(Customer::default()));
    
    let app = build_v1_router(state_with_customer_mock(mock_service));
    let response = app
        .oneshot(Request::builder()
            .uri("/api/v1/customers")
            .method("POST")
            .header("Authorization", "Bearer valid-jwt")
            .header("Content-Type", "application/json")
            .body(Body::from(r#"{"name":"Acme Corp","email":"c@acme.com"}"#))
            .unwrap())
        .await;
    
    assert_eq!(response.status(), 201);
}
```

---

## 14. Dependencies

| Module | Dependency Type | Usage |
|---|---|---|
| `identity` | Trait reference | Tenant validation, permission checks |
| `audit` | NATS event | Audit logging of customer operations |
| `notification` | NATS event | Customer lifecycle notifications |

---

## 15. Configuration

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `CUSTOMER_MAX_ADDRESSES` | Max addresses per customer | 10 |
| `CUSTOMER_INACTIVE_DAYS` | Days before marking inactive | 90 |
| `CUSTOMER_ARCHIVE_DAYS` | Days before archiving inactive | 365 |
