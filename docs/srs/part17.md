# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 17 — Webhook & Event Delivery Architecture

## Outbound Webhooks, Event Subscription, HMAC Security, Retry Logic

---

## 1. Webhook Module Overview

The Webhook module enables external systems to subscribe to AeroXe events:

- Tenant-configurable outbound webhooks
- Event subscription management
- Guaranteed delivery with retry
- Webhook signature verification (HMAC)
- Delivery logging and debugging
- Rate limiting per webhook
- IP whitelisting
- Event filtering

---

## 2. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-webhook` |
| Bounded Context | Integration |
| Schema | `webhook_` (shared PostgreSQL) |

---

## 3. Webhook Security

| Requirement | Implementation |
|---|---|
| IP Whitelist | CIDR ranges allowed |
| TLS Required | Reject non-HTTPS |
| HMAC Signature | `X-AeroXe-Signature: sha256=<hex-digest>` |
| Timestamp Validation | Prevent replay attacks |
| Rate Limiting | Per-subscription + per-IP |
| Payload Size Limit | Max 1MB |

---

## 4. Delivery Pipeline

```
NATS Event → Event Filter → Rate Limit Check → Payload Construction
  → HMAC Signature → HTTP POST → Response Handling → Retry Logic
```

---

## 5. Supported Event Types

| Event Type | Description |
|---|---|
| `conversation.created` | New conversation started |
| `conversation.ended` | Conversation completed |
| `call.inbound` | Inbound call received |
| `call.completed` | Call ended |
| `customer.created` | New customer |
| `agent.completed` | Agent execution done |
| `security.threat.detected` | Security event |
| `kyc.approved` | KYC verified |

---

## 6. REST API Endpoints

| Method | Endpoint | Business Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/webhooks` | `CREATED` | `201` |
| `GET` | `/api/v1/webhooks/{id}` | `SUCCESS` | `200` |
| `GET` | `/api/v1/webhooks?limit=10&offset=0` | `SUCCESS` | `200` |
| `PATCH` | `/api/v1/webhooks/{id}` | `UPDATED` | `200` |
| `DELETE` | `/api/v1/webhooks/{id}` | `DELETED` | `204` |
| `POST` | `/api/v1/webhooks/{id}/test` | `SUCCESS` | `200` |
| `GET` | `/api/v1/webhooks/{id}/deliveries?limit=10&offset=0` | `SUCCESS` | `200` |

---

## 7. Database Tables

| Table | Purpose |
|---|---|
| `webhook.subscriptions` | Webhook subscriptions |
| `webhook.deliveries` | Delivery attempts |

---

# End of Part 17
