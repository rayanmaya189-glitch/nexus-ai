# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 9 — AI Agent Framework & Advanced RAG Intelligence Engine

## Agentic AI + Hybrid RAG + Memory + Knowledge Graph + Self-Improvement Architecture

---

# 1. AI Intelligence Architecture Overview

AeroXe Nexus AI is not a simple chatbot.

It is an **Agentic AI Platform** where multiple specialized AI agents collaborate within a modular monolith.

Architecture:

```text
                    User


                      |

                      |

              gateway module

              src/modules/gateway/


                      |

                      |

              ai-gateway module

              (trait call → agent module)


                      |

================================================


              AI Intelligence Layer

              (src/modules/*)


================================================


 ai-gateway    agent    rag     vision

 memory        sql-agent  security


                      |

================================================


             Knowledge & Data Layer

             (Schema-per-Module, SeaORM)


================================================


 PostgreSQL 16 (shared cluster)

 pgvector

 Apache AGE

 Elasticsearch

 Redis

 MinIO


================================================


                      |

                Ollama Models


================================================

```

---

# 2. Agent Framework Design

## 2.1 Agent Definition

An Agent is an autonomous software entity capable of:

* Understanding goals
* Creating plans
* Calling tools
* Accessing knowledge
* Using memory
* Producing decisions

---

# 3. Agent Architecture

Every agent contains:

```text id="n1q2ko"
Agent


|

+-- Identity

|

+-- System Prompt

|

+-- Model

|

+-- Tools

|

+-- Memory

|

+-- Knowledge Sources

|

+-- Policies

|

+-- Evaluation Rules

```

---

# 4. Agent Lifecycle

```text id="pl7v5x"
          User Request


              |

              |

          Intent Detection


              |

              |

          Task Planning


              |

              |

          Tool Selection


              |

              |

          Execution


              |

              |

          Knowledge Retrieval


              |

              |

          Reasoning


              |

              |

          Response Generation


              |

              |

          Memory Update


```

---

# 5. Agent Types

AeroXe Nexus AI includes:

---

# 5.1 Planner Agent

Model:

```
lfm2.5-thinking:1.2b
```

Responsibilities:

* Understand user objective
* Break tasks into steps
* Select agents

Example:

User:

> Analyze customer complaint

Planner:

```json
{
"steps":[

"Get customer details",

"Check billing",

"Check network",

"Search knowledge",

"Generate solution"

]

}

```

---

# 5.2 Customer Support Agent

Model:

```
Phi-4-mini:3.8b
```

Purpose:

* Customer support
* FAQ
* Ticket creation

Tools:

```
customer.lookup()

ticket.create()

billing.check()

network.status()

```

---

# 5.3 Developer Agent

Model:

```
Qwen2.5-Coder:3B
```

Capabilities:

* Code generation
* Code review
* Debugging
* Architecture suggestions

Tools:

```
git.search()

code.analyze()

test.execute()

```

---

# 5.4 RAG Knowledge Agent

Model:

```
Command-R7B
```

Purpose:

Enterprise knowledge reasoning.

Sources:

```
Documents

Policies

Manuals

Tickets

Database

```

---

# 5.5 Vision Agent

Model:

```
Qwen3-VL:4B
```

Capabilities:

* Image understanding
* OCR
* Screenshot analysis
* Document extraction

---

# 5.6 Security Agent

Model:

```
WhiteRabbitNeo:7B
```

Capabilities:

* Code security review
* Vulnerability analysis
* Threat analysis

---

# 5.7 Business Intelligence Agent

Model:

```
Llama3.1:7B
```

Capabilities:

* Business analysis
* Reports
* Forecasting

---

# 6. Agent Module Design

Module:

```
agent (src/modules/agent/)
```

Responsibilities:

* Agent selection
* Execution planning
* Tool management
* Context management
* Published versioned NATS events

---

# 7. Agent Decision Flow

Example:

User:

"Why is customer internet slow?"

Flow:

```text
User


 |

Planner Agent


 |

Task Plan


 |

===========================


Customer Agent


      |

Customer Data


      |


Network Agent


      |

Network Status


      |


RAG Agent


      |

Troubleshooting Guide


===========================


 |

Final Answer

```

---

# 8. Tool Calling Architecture

Agents do not directly access systems.

They use controlled tools.

Architecture:

```text
 id="7h01kv"
Agent


 |

Tool Request


 |

Tool Gateway


 |

Permission Engine


 |

Service API


 |

Result


```

---

# 9. Tool Definition Format

Example:

```json id="3r5w9n"
{
"name":
"customer.lookup",

"description":
"Get customer information",

"parameters":{

 "customer_id":"string"

}

}

```

---

# 10. Model Router Architecture

Purpose:

Select best model for the task.

Implemented as a service within the `ai-gateway` module (not a separate service).

Flow:

```text
Request


 |

Classifier


 |

Model Router (ai-gateway module)


 |

=====================


Simple Query

     |

Phi-4 Mini


Coding

     |

Qwen Coder


Vision

     |

Qwen3-VL


Security

     |

WhiteRabbitNeo


Complex Reasoning

     |

Llama3.1


RAG

     |

Command-R


=====================


```

---

# 11. Memory Architecture

AI memory has three layers.

---

# 11.1 Short-Term Memory

Technology:

```
Redis
```

Contains:

* Current conversation
* Temporary context
* Active tasks

TTL:

Example:

```
24 hours
```

---

# 11.2 Long-Term Memory

Technology:

```
PostgreSQL + pgvector
```

Stores:

* User preferences
* Previous conversations
* Important facts

---

# 11.3 Organizational Memory

Technology:

```
Apache AGE
```

Stores:

Relationships:

```
Customer

 |

uses

 |

Product

 |

related_to

 |

Issue

```

---

# 12. Advanced RAG Architecture

Traditional RAG:

```
Question

 |

Search

 |

Answer

```

AeroXe Nexus AI uses:

## Advanced Hybrid RAG

```text
 id="9smz2o"
Question


 |

Query Understanding


 |

================================


Vector Search

(pgvector)


Keyword Search

(ElasticSearch)


Knowledge Graph Search

(Apache AGE)


Database Query

(PostgreSQL)


================================


 |

Result Fusion


 |

Re-ranking


 |

LLM Reasoning


 |

Answer

```

---

# 13. RAG Pipeline

## Step 1 — Document Ingestion

Sources:

```
PDF

DOCX

HTML

Database

API

Code Repository

```

---

# Step 2 — Processing

Pipeline:

```
Document


 |

Parser


 |

Cleaning


 |

Chunking


 |

Metadata Extraction


 |

Embedding (nomic-embed-text: 768 dimensions, via Ollama)


 |

Storage

```

---

# 14. Intelligent Chunking

Instead of fixed chunks:

Traditional:

```
500 tokens
```

AeroXe:

```
Semantic Chunking

```

Example:

Document:

```
Billing Policy

Section 1

Section 2

Section 3

```

Chunks preserve meaning.

---

# 15. Embedding Architecture

Embedding model: **nomic-embed-text (768 dimensions, via Ollama)**

Embedding stores:

```
Text Meaning

Context

Metadata

Relationships

```

Stored:

```
pgvector

nomic-embed-text (768 dimensions, via Ollama)

```

---

# 16. Hybrid Search

Combines:

## Vector Search

Finds:

```
similar meaning
```

## Keyword Search

Finds:

```
exact terms
```

## Graph Search

Finds:

```
relationships
```

---

# 17. Re-ranking System

Problem:

First search results may not be best.

Solution:

Reranker.

Flow:

```
100 Results


 |

Reranker


 |

Top 5 Results


 |

LLM

```

---

# 18. RAG Security Layer

Before retrieval:

Checks:

* Tenant permission
* Document access
* Data classification

Example:

User:

```
Support Employee
```

Cannot retrieve:

```
Financial Reports

```

---

# 19. Real-Time Data RAG

AeroXe Nexus AI supports live business data.

Architecture:

```text
User Question


 |

SQL Agent


 |

Database Query


 |

Result


 |

RAG Context


 |

LLM


 |

Answer

```

---

# 20. SQL + RAG Combined Intelligence

Example:

Question:

"Why did revenue decrease this month?"

System:

```
RAG:

Find sales policies


+

SQL:

Calculate revenue


+

Graph:

Find business relationships


+

LLM:

Explain reason

```

---

# 21. Knowledge Graph Reasoning

Apache AGE graph example:

```text
Customer

 |

owns

 |

Router

 |

connected_to

 |

OLT

 |

located_in

 |

Jalgaon

```

AI can reason:

"Customers connected to this OLT are affected."

---

# 22. Agent Self-Improvement Loop

Architecture:

```text
 id="w9i4kl"
Agent Response


 |

User Feedback


 |

Evaluation


 |

Performance Score


 |

Prompt Improvement


 |

Knowledge Update


 |

Better Agent


```

---

# 23. AI Evaluation Framework

Every response evaluated:

Metrics:

| Metric    | Purpose           |
| --------- | ----------------- |
| Accuracy  | Correct answer    |
| Relevance | Useful response   |
| Safety    | No harmful output |
| Latency   | Speed             |
| Cost      | Resource usage    |

---

# 24. Human Approval Workflow

For sensitive actions:

Example:

```
AI wants:

Refund ₹50,000

```

Flow:

```
AI Recommendation

       |

Human Approval

       |

Execution

```

---

# 25. Agent Observability

Track:

```
Agent ID

Model Used

Tokens

Tools Used

Latency

Decision Path

Errors

```

---

# 26. Final AI Architecture

```text
                         User


                           |


                    AI Gateway


                           |


                 Agent Orchestrator


                           |


================================================


Planning Agent

Reasoning Agent

Tool Agent

Memory Agent

RAG Agent


================================================


                           |


================================================


Knowledge Intelligence


pgvector

Elastic

Apache AGE

PostgreSQL


================================================


                           |


                      Ollama


================================================


LFM Thinking

Hermes3

Phi4 Mini

Qwen Coder

Qwen3-VL

Command-R

Llama3.1

WhiteRabbitNeo


================================================

```

---

# Part 9 Completed

Covered:

✅ Agent Framework
✅ Agent Lifecycle
✅ Model Router
✅ Tool Calling
✅ Memory Architecture
✅ Hybrid RAG
✅ Vector Search
✅ Knowledge Graph
✅ SQL Intelligence Integration
✅ Real-Time Data RAG
✅ AI Evaluation Loop
✅ Human Approval Workflow

---
