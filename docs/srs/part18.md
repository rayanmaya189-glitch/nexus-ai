# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 18 — Outbound Campaign Architecture

## Proactive AI Calls, Campaign Management, DNC Compliance, Scheduling

---

## 1. Outbound Module Overview

The Outbound module enables AI-initiated communications:

- Outbound voice calls (AI calling customers)
- Proactive chat messages
- Campaign management (bulk outbound)
- Callback scheduling
- Do-Not-Call (DNC) compliance
- Call scheduling and timezone management
- Outbound rate limiting

---

## 2. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-outbound` |
| Bounded Context | Outbound Communication |
| Schema | `outbound_` (shared PostgreSQL) |

---

## 3. Campaign Execution Flow

```
Campaign Created (Draft) → Target Selection → Schedule
  → Execution Loop → Conversation → Post-Call → Repeat
```

---

## 4. DNC Compliance

| Rule | Description |
|---|---|
| Universal DNC | Check national DNC registry |
| Tenant DNC | Tenant-specific block list |
| Customer DNC | Customer-level opt-out |
| Time-based DNC | No calls outside business hours |
| Frequency limit | Max N calls per customer per period |

---

## 5. REST API Endpoints

| Method | Endpoint | Business Status | HTTP |
|---|---|---|---|
| `POST` | `/api/v1/outbound/campaigns` | `CREATED` | `201` |
| `GET` | `/api/v1/outbound/campaigns/{id}` | `SUCCESS` | `200` |
| `GET` | `/api/v1/outbound/campaigns?limit=10&offset=0` | `SUCCESS` | `200` |
| `POST` | `/api/v1/outbound/campaigns/{id}/start` | `UPDATED` | `200` |
| `POST` | `/api/v1/outbound/callbacks` | `CREATED` | `201` |
| `GET` | `/api/v1/outbound/callbacks?limit=10&offset=0` | `SUCCESS` | `200` |
| `DELETE` | `/api/v1/outbound/callbacks/{id}` | `DELETED` | `204` |
| `POST` | `/api/v1/outbound/dnc` | `CREATED` | `201` |
| `GET` | `/api/v1/outbound/dnc?limit=10&offset=0` | `SUCCESS` | `200` |
| `DELETE` | `/api/v1/outbound/dnc/{id}` | `DELETED` | `204` |

---

## 6. Database Tables

| Table | Purpose |
|---|---|
| `outbound_.campaigns` | Campaign definitions |
| `outbound_.targets` | Campaign target customers |
| `outbound_.callbacks` | Scheduled callbacks |
| `outbound_.dnc_list` | Do-Not-Call list |

---

# End of Part 18
