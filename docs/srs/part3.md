# AeroXe Nexus AI Platform

# Software Requirements Specification (SRS) v1.0

# Part 3 — Communication Architecture Design

## gRPC + Protocol Buffers + NATS JetStream Event Architecture

---

# 1. Communication Strategy Overview

AeroXe Nexus AI follows a **hybrid communication architecture**.

The communication model:

```text
                     External World


                           |

                  REST / WebSocket / HTTPS


                           |

                    API Gateway


                           |

================================================

                 Internal Communication


================================================


        Synchronous              Asynchronous


            gRPC                  NATS JetStream


             |                         |

             |                         |

     Service-to-Service        Event Driven Flow


```

---

# 2. Communication Rules

## External Communication

Used by:

* Web applications
* Mobile applications
* AeroXe products
* Third-party integrations

Protocol:

```
HTTPS REST
WebSocket
GraphQL (optional future)
```

---

## Internal Communication

Used by:

* Microservices
* AI agents
* Infrastructure services

Protocol:

```
gRPC + Protobuf
```

---

## Event Communication

Used for:

* Background jobs
* Notifications
* AI workflow
* Data synchronization

Protocol:

```
NATS JetStream
```

---

# 3. gRPC Architecture

## Service Communication

Example:

```text
agent-orchestrator-service


          |

          | gRPC

          |

rag-service


          |

          | gRPC

          |

vector-search-service

```

---

# 4. gRPC Design Principles

Every gRPC service must have:

* Versioned protobuf contracts
* Backward compatibility
* Strong typing
* Error standards
* Authentication metadata

Example:

```
grpc.metadata

tenant-id
user-id
request-id
authorization
```

---

# 5. Proto Repository Structure

Recommended:

```
aeroxe-proto/


├── common/

│
├── identity/

│
├── agent/

│
├── rag/

│
├── vision/

│
├── workflow/

│
├── memory/

│
├── security/

│
└── audit/

```

---

# 6. Common Proto Definition

File:

```
common/common.proto
```

```protobuf
syntax = "proto3";


package aeroxe.common;


message RequestContext {

 string request_id = 1;

 string tenant_id = 2;

 string user_id = 3;

 string trace_id = 4;

}


message ErrorResponse {

 string code = 1;

 string message = 2;

 string details = 3;

}

```

---

# 7. Identity Service gRPC

Service:

```
identity-service
```

Proto:

```
identity.proto
```

```protobuf
syntax = "proto3";


package aeroxe.identity;



service IdentityService {


 rpc CreateUser(CreateUserRequest)
 returns(CreateUserResponse);


 rpc Authenticate(AuthRequest)
 returns(AuthResponse);


 rpc CheckPermission(PermissionRequest)
 returns(PermissionResponse);


}


message CreateUserRequest {

string tenant_id = 1;

string email = 2;

string password = 3;

}


message CreateUserResponse {

string user_id = 1;

}

```

---

# 8. AI Gateway gRPC

Service:

```
ai-gateway-service
```

Purpose:

Central AI request processing.

```protobuf
service AIGatewayService {


rpc SubmitRequest(AIRequest)

returns(AIResponse);



rpc StreamResponse(AIRequest)

returns(stream AIChunk);


}


message AIRequest {


string session_id = 1;

string prompt = 2;

string agent = 3;

}


message AIResponse {


string response = 1;

string model = 2;

}

```

---

# 9. Agent Orchestrator gRPC

Service:

```
agent-orchestrator-service
```

Purpose:

Manage AI execution.

```protobuf
service AgentService {


rpc StartExecution(StartAgentRequest)

returns(AgentExecutionResponse);



rpc GetExecutionStatus(StatusRequest)

returns(StatusResponse);


}


message StartAgentRequest {


string task = 1;

string agent_type = 2;

string context = 3;


}

```

---

# 10. RAG Service gRPC

Service:

```
rag-service
```

Purpose:

Knowledge retrieval.

```protobuf
service RagService {


rpc SearchKnowledge(SearchRequest)

returns(SearchResponse);



rpc UploadDocument(DocumentRequest)

returns(DocumentResponse);



}



message SearchRequest {


string query = 1;


int32 limit = 2;


}


message SearchResponse {


repeated Document documents = 1;


}

```

---

# 11. Vision Service gRPC

Service:

```
vision-service
```

Model:

```
Ollama Qwen3-VL:4B
```

```protobuf
service VisionService {


rpc AnalyzeImage(ImageRequest)

returns(ImageAnalysisResponse);



rpc ExtractText(ImageRequest)

returns(OCRResponse);


}


message ImageRequest {


bytes image = 1;

string type = 2;


}


message ImageAnalysisResponse {


string description = 1;


float confidence = 2;


}

```

---

# 12. SQL Intelligence Service gRPC

Service:

```
sql-agent-service
```

Purpose:

Safe business data intelligence.

```protobuf
service SQLService {


rpc GenerateQuery(QueryRequest)

returns(SQLResponse);



rpc ExecuteQuery(SQLRequest)

returns(ResultResponse);


}


message QueryRequest {


string question = 1;


string database = 2;


}


```

---

# 13. Memory Service gRPC

Service:

```
memory-service
```

```protobuf
service MemoryService {


rpc StoreMemory(StoreMemoryRequest)

returns(MemoryResponse);



rpc SearchMemory(SearchMemoryRequest)

returns(MemoryList);


}


message StoreMemoryRequest {


string user_id = 1;


string content = 2;


string type = 3;


}

```

---

# 14. Workflow Service gRPC

Service:

```
workflow-service
```

```protobuf
service WorkflowService {


rpc StartWorkflow(StartWorkflowRequest)

returns(WorkflowResponse);



rpc GetWorkflowStatus(StatusRequest)

returns(StatusResponse);


}

```

---

# 15. Security AI Service gRPC

Service:

```
security-ai-service
```

Model:

```
WhiteRabbitNeo 7B
```

```protobuf
service SecurityService {


rpc AnalyzeSecurity(SecurityRequest)

returns(SecurityReport);


}


message SecurityRequest {


string target = 1;


string type = 2;


}

```

---

# 16. gRPC Error Standards

All services use:

```protobuf
enum ErrorCode {


UNKNOWN = 0;


INVALID_REQUEST = 1;


UNAUTHORIZED = 2;


FORBIDDEN = 3;


NOT_FOUND = 4;


TIMEOUT = 5;


MODEL_ERROR = 6;


DATABASE_ERROR = 7;


}

```

---

# 17. NATS JetStream Architecture

NATS is the event backbone.

Used for:

* AI tasks
* Document processing
* Agent lifecycle
* Audit events
* Workflow events

---

# 18. NATS Subject Naming Standard

Format:

```
aeroxe.<domain>.<event>
```

Example:

```
aeroxe.ai.request.created

aeroxe.agent.execution.started

aeroxe.rag.document.processed

aeroxe.vision.analysis.completed

```

---

# 19. Core NATS Subjects

## AI Events

```
aeroxe.ai.request.created

aeroxe.ai.response.generated

aeroxe.ai.failed

```

---

## Agent Events

```
aeroxe.agent.started

aeroxe.agent.completed

aeroxe.agent.failed

```

---

## RAG Events

```
aeroxe.rag.document.uploaded

aeroxe.rag.document.processed

aeroxe.rag.embedding.created

```

---

## Vision Events

```
aeroxe.vision.image.received

aeroxe.vision.analysis.completed

```

---

## Workflow Events

```
aeroxe.workflow.started

aeroxe.workflow.completed

aeroxe.workflow.failed

```

---

## Security Events

```
aeroxe.security.scan.started

aeroxe.security.threat.detected

```

---

# 20. Event Schema Standard

Every event:

```json
{
 "event_id":"uuid",

 "event_type":"AgentCompleted",

 "timestamp":"2026-07-15T12:00:00Z",

 "tenant_id":"uuid",

 "service":"agent-service",

 "version":"1.0",

 "data":{

 }

}

```

---

# 21. Example Event

Agent Completed:

```json
{
 "event_type":"AgentCompleted",

 "service":"agent-orchestrator",

 "data":{

    "execution_id":"12345",

    "agent":"customer-agent",

    "status":"success"

 }

}

```

---

# 22. JetStream Stream Design

## AI Stream

```
AI_EVENTS
```

Subjects:

```
aeroxe.ai.*
```

Retention:

```
7 days
```

---

## Agent Stream

```
AGENT_EVENTS
```

Subjects:

```
aeroxe.agent.*
```

Retention:

```
30 days
```

---

## Audit Stream

```
AUDIT_EVENTS
```

Subjects:

```
aeroxe.audit.*
```

Retention:

```
365 days
```

---

# 23. Request Flow Example

## User asks:

"Why is customer internet slow?"

Flow:

```
User

 |

API Gateway

 |

AI Gateway

 |

Agent Orchestrator

 |

LFM Thinking Model

 |

Plan Created


 |

NATS Event

aeroxe.agent.execution.started


 |

Customer Agent


 |

gRPC


 |

Broadband Service


 |

SQL Agent


 |

Customer DB


 |

RAG Service


 |

Command-R


 |

Final Response


```

---

# 24. Streaming Response Architecture

For Chat UI:

```
User

 |

WebSocket

 |

AI Gateway

 |

gRPC Stream

 |

Ollama


Token Streaming


 |

User Interface

```

---

# 25. Security Requirements

gRPC:

* TLS encryption
* Service authentication
* Metadata validation

NATS:

* TLS
* Account isolation
* Subject permissions

---

# 26. Final Communication Stack

| Layer           | Technology         |
| --------------- | ------------------ |
| Mobile/Web API  | REST               |
| Real-time Chat  | WebSocket          |
| Internal RPC    | gRPC               |
| Contract        | Protocol Buffers   |
| Event Bus       | NATS JetStream     |
| AI Runtime API  | Ollama API         |
| Database Access | Repository Pattern |

---

# Part 3 Completed

The AeroXe Nexus AI communication foundation is now defined:

✅ gRPC Service Contracts
✅ Protobuf Structure
✅ NATS JetStream Event Architecture
✅ Event Naming Standards
✅ Streaming Design
✅ Security Rules
