# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 16 — Analytics & Business Intelligence Architecture

## Conversation Metrics, Call Center Metrics, Cost Allocation, Agent Scoring

---

## 1. Analytics Module Overview

The Analytics module provides business intelligence and operational insights:

- Real-time conversation metrics
- AI agent performance tracking
- Customer satisfaction analytics
- Voice call analytics (call center metrics)
- Cost and usage tracking
- Anomaly detection
- Custom report generation
- Dashboard data APIs

---

## 2. Module Identity

| Attribute | Value |
|---|---|
| Module Name | `nexus-analytics` |
| Bounded Context | Analytics & Intelligence |
| Schema | `analytics_` (shared PostgreSQL) |

---

## 3. Call Center Metrics

| Metric | Description | Target |
|---|---|---|
| AHT (Average Handle Time) | Total talk + hold + wrap-up time | < 5 min |
| ASA (Average Speed of Answer) | Time to answer in queue | < 30s |
| ACW (After Call Work) | Post-call processing time | < 30s |
| FCR (First Call Resolution) | Resolved without callback | > 80% |
| Abandonment Rate | Calls abandoned in queue | < 5% |
| Service Level | % answered within threshold | > 80% |
| Occupancy Rate | Agent busy time / available time | 70-85% |

---

## 4. Conversation Cost Allocation

```rust
pub struct ConversationCost {
    pub conversation_id: ConversationId,
    pub llm_cost: f64,
    pub llm_tokens_input: u64,
    pub llm_tokens_output: u64,
    pub stt_cost: f64,
    pub tts_cost: f64,
    pub telephony_cost: f64,
    pub storage_cost: f64,
    pub total_cost: f64,
}
```

---

## 5. Agent Performance Scoring

| Factor | Weight | Description |
|---|---|---|
| Customer Satisfaction | 30% | CSAT scores |
| Resolution Rate | 25% | Issues resolved without escalation |
| Response Time | 20% | Speed of response |
| First Contact Resolution | 15% | Resolved in first interaction |
| Compliance Score | 10% | Script adherence |

---

## 6. REST API Endpoints

| Method | Endpoint | Business Status | HTTP |
|---|---|---|---|
| `GET` | `/api/v1/analytics/dashboard?start=...&end=...` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/realtime` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/agents?limit=10&offset=0` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/agents/{agent_id}/performance` | `SUCCESS` | `200` |
| `GET` | `/api/v1/analytics/costs?start=...&end=...` | `SUCCESS` | `200` |
| `POST` | `/api/v1/analytics/reports` | `CREATED` | `201` |

---

## 7. Database Tables

| Table | Purpose |
|---|---|
| `analytics.conversation_metrics` | Conversation metrics (partitioned) |
| `analytics.call_metrics` | Call metrics (partitioned) |
| `analytics.agent_metrics` | Agent metrics (partitioned) |
| `analytics.cost_tracking` | Cost tracking (partitioned) |
| `analytics.snapshots` | Pre-aggregated snapshots |
| `analytics.conversation_costs` | Per-conversation costs |

---

# End of Part 16
