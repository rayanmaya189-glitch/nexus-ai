# AeroXe Nexus AI — Python SDK

## Official Python Client Library for AeroXe Nexus AI Platform

---

## 1. Overview

The Python SDK provides a Pythonic, async-first client for the AeroXe Nexus AI Platform. Built with Python 3.10+, it supports both synchronous and asynchronous usage patterns.

### Package

```
aeronexus
```

### Requirements

- Python 3.10+
- pip 21.0+

---

## 2. Installation

```bash
pip install aeronexus

# With async support
pip install aeronexus[async]

# With all optional features
pip install aeronexus[all]
```

### Optional Dependencies

```bash
pip install aeronexus[keyring]    # Secure token storage
pip install aeronexus[websocket]  # WebSocket support
pip install aeronexus[async]      # Async support (aiohttp)
```

---

## 3. Client Initialization

### Basic Configuration

```python
from aeronexus import NexusClient

client = NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
)

response = client.ai.chat(
    message="Hello, Nexus!",
    agent="general",
)
print(response.answer)
```

### Configuration Options

```python
from aeronexus import NexusClient
from aeronexus.config import ClientConfig

config = ClientConfig(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
    timeout=30.0,
    connect_timeout=10.0,
    max_retries=3,
    retry_delay=2.0,
    max_retry_delay=30.0,
    rate_limit_burst=100,
    rate_limit_window=60.0,
    user_agent="my-app/1.0",
    debug=False,
    insecure=False,  # Skip TLS verification (dev only)
)

client = NexusClient(config)
```

### Environment Variables

```python
import os
from aeronexus import NexusClient

client = NexusClient(
    base_url=os.getenv("NEXUS_BASE_URL", "https://api.aeroxenexus.com"),
    api_key=os.getenv("NEXUS_API_KEY"),
    timeout=float(os.getenv("NEXUS_TIMEOUT", "30")),
)
```

---

## 4. Authentication

### Login

```python
token = client.auth.login(
    email="admin@company.com",
    password="password",
)

print(token.access_token)
print(token.refresh_token)
print(token.expires_in)
```

### Token Refresh

```python
new_token = client.auth.refresh(
    refresh_token=token.refresh_token,
)
```

### Auto-Refresh

The SDK automatically refreshes tokens before expiry:

```python
from aeronexus import NexusClient

client = NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
    auto_refresh=True,  # Default: True
)
```

### Logout

```python
client.auth.logout()
```

---

## 5. AI Chat

### Basic Chat

```python
response = client.ai.chat(
    message="Explain my customer complaint",
    agent="customer-agent",
    conversation_id="conv-123",  # Optional: continue existing conversation
)

print(response.conversation_id)
print(response.answer)
print(response.model)
```

### Streaming Chat

```python
for event in client.ai.chat_stream(
    message="Analyze my broadband issue",
    agent="customer-agent",
):
    match event.type:
        case "token":
            print(event.content, end="", flush=True)
        case "tool_call":
            print(f"\n[Tool Call] {event.content}")
        case "tool_result":
            print(f"[Tool Result] {event.content}")
        case "completed":
            print("\n[Stream Complete]")
        case "error":
            print(f"\n[Error] {event.content}")
```

### Async Chat

```python
import asyncio
from aeronexus import AsyncNexusClient

async def main():
    client = AsyncNexusClient(
        base_url="https://api.aeroxenexus.com",
        api_key="your-api-key",
    )

    response = await client.ai.chat(
        message="Hello, Nexus!",
        agent="general",
    )
    print(response.answer)

asyncio.run(main())
```

### Async Streaming

```python
async def stream_chat():
    client = AsyncNexusClient(
        base_url="https://api.aeroxenexus.com",
        api_key="your-api-key",
    )

    async for event in client.ai.chat_stream(
        message="Analyze my broadband issue",
        agent="customer-agent",
    ):
        match event.type:
            case "token":
                print(event.content, end="", flush=True)
            case "completed":
                print("\n[Done]")

asyncio.run(stream_chat())
```

### Chat with Context

```python
response = client.ai.chat(
    message="What's the status of ticket #123?",
    agent="customer-agent",
    context={
        "ticket_id": "tkt_123",
        "customer_id": "cust_456",
    },
    model="command-r7b",
    temperature=0.7,
    max_tokens=1000,
)
```

### Conversation History

```python
history = client.ai.get_conversation("conv-123")

for msg in history.messages:
    print(f"[{msg.role}] {msg.timestamp}: {msg.content}")
```

---

## 6. Agent Execution

### Execute Agent

```python
execution = client.agents.execute(
    agent="developer-agent",
    task="Review this Go code for security vulnerabilities",
    context={
        "repository": "backend",
        "file": "pkg/auth/handler.go",
    },
)

print(execution.execution_id)
print(execution.status)  # "started"
```

### Get Execution Status

```python
status = client.agents.get_execution("exec-abc123")

print(status.status)  # "completed"
print(status.result)  # "Code review complete. 3 issues found."

for step in status.steps:
    print(f"Step: {step.step} - Status: {step.status}")
```

### Stream Agent Execution

```python
for event in client.agents.stream_execution("exec-abc123"):
    print(f"[{event.step}] {event.status}")
```

### List Agents

```python
agents = client.agents.list(
    page=1,
    page_size=20,
)

for agent in agents.data:
    print(f"Agent: {agent.name} - {agent.description}")
```

### Get Agent Details

```python
agent = client.agents.get("customer-agent")

print(agent.name)
print(agent.tools)
print(agent.model)
```

---

## 7. RAG / Knowledge Management

### Upload Document

```python
with open("network-guide.pdf", "rb") as f:
    doc = client.rag.upload_document(
        file=f,
        file_name="network-guide.pdf",
        content_type="application/pdf",
        metadata={
            "category": "network",
            "version": "1.0",
        },
    )

print(doc.document_id)  # "doc-uuid"
print(doc.status)       # "processing"
```

### Get Document Status

```python
status = client.rag.get_document_status("doc-uuid")

print(status.status)   # "completed"
print(status.chunks)   # 42
print(status.size)     # 1024000
```

### List Documents

```python
docs = client.rag.list_documents(
    page=1,
    page_size=20,
    status="completed",
)

for doc in docs.data:
    print(f"Document: {doc.file_name} - Status: {doc.status}")
```

### Delete Document

```python
client.rag.delete_document("doc-uuid")
```

### Search Knowledge Base

```python
results = client.rag.search(
    query="How to configure ONU?",
    limit=5,
    filters={"category": "network"},
)

for result in results.results:
    print(f"Title: {result.title}")
    print(f"Score: {result.score:.2f}")
    print(f"Content: {result.content}")
    print(f"Source: {result.source}\n")
```

---

## 8. Vision Intelligence

### Analyze Image

```python
with open("router-photo.jpg", "rb") as f:
    analysis = client.vision.analyze(
        file=f,
        file_name="router-photo.jpg",
        task="identify problem",
    )

print(analysis.description)  # "Router LED is showing red"
print(analysis.confidence)   # 0.94
```

### OCR

```python
with open("document-scan.png", "rb") as f:
    ocr_result = client.vision.ocr(
        file=f,
        file_name="document-scan.png",
        language="eng",
    )

print(ocr_result.text)
print(ocr_result.confidence)
```

---

## 9. SQL Intelligence

### Natural Language Query

```python
result = client.sql.query(
    question="Show monthly revenue for last 6 months",
    database="aeroxe_billing_db",
)

print(result.sql)      # "SELECT SUM(amount)..."
print(result.row_count) # 6

for row in result.data:
    print(row)
```

### List Available Databases

```python
databases = client.sql.list_databases()

for db in databases:
    print(f"Database: {db.name} - {db.description}")
```

---

## 10. Memory

### Store Memory

```python
memory = client.memory.store(
    user_id="user-123",
    memory="Customer prefers Hindi support",
    type="preference",
    metadata={"language": "hindi"},
)

print(memory.id)
```

### Search Memory

```python
results = client.memory.search(
    query="customer language preference",
    limit=5,
)

for result in results.data:
    print(f"Memory: {result.memory} (Score: {result.score:.2f})")
```

### Delete Memory

```python
client.memory.delete("mem-uuid")
```

---

## 11. Workflow Automation

### Start Workflow

```python
workflow = client.workflow.start(
    workflow="customer-support-flow",
    context={"ticket_id": "tkt_123"},
)

print(workflow.workflow_id)  # "wf-123"
print(workflow.status)       # "running"
```

### Get Workflow Status

```python
status = client.workflow.get_status("wf-123")

print(status.status)  # "completed"
for step in status.steps:
    print(f"Step: {step.step} - Status: {step.status}")
```

### List Workflows

```python
workflows = client.workflow.list(
    page=1,
    page_size=20,
)
```

---

## 12. Model Management

### List Available Models

```python
models = client.models.list()

for model in models:
    print(f"Model: {model.name} - Type: {model.type} - Status: {model.status}")
```

---

## 13. WebSocket Client

### Connect to Chat Stream

```python
from aeronexus.websocket import WebSocketClient

ws = WebSocketClient(
    url="wss://api.aeroxenexus.com/ws/chat/conv-123",
    headers={"Authorization": f"Bearer {token.access_token}"},
)

# Send message
ws.send({"type": "message", "content": "Analyze my broadband issue"})

# Receive messages
for msg in ws:
    match msg["type"]:
        case "token":
            print(msg["content"], end="", flush=True)
        case "tool_call":
            print(f"\n[Tool] {msg['content']}")
        case "completed":
            print("\n[Done]")
            break
```

### Async WebSocket

```python
import asyncio
from aeronexus.websocket import AsyncWebSocketClient

async def main():
    ws = AsyncWebSocketClient(
        url="wss://api.aeroxenexus.com/ws/chat/conv-123",
        headers={"Authorization": f"Bearer {token.access_token}"},
    )

    await ws.connect()
    await ws.send({"type": "message", "content": "Hello"})

    async for msg in ws:
        match msg["type"]:
            case "token":
                print(msg["content"], end="", flush=True)
            case "completed":
                print("\n[Done]")
                break

    await ws.close()

asyncio.run(main())
```

### WebSocket Event Handler

```python
from aeronexus.websocket import WebSocketClient, EventHandler

class MyHandler(EventHandler):
    def on_token(self, content: str):
        print(content, end="", flush=True)

    def on_tool_call(self, content: str):
        print(f"\n[Tool] {content}")

    def on_error(self, error: Exception):
        print(f"\n[Error] {error}")

    def on_complete(self):
        print("\n[Done]")

ws = WebSocketClient(
    url="wss://api.aeroxenexus.com/ws/chat/conv-123",
    handler=MyHandler(),
)
ws.connect()
```

---

## 14. Error Handling

### Custom Exception Types

```python
from aeronexus.exceptions import (
    NexusError,
    UnauthorizedError,
    TokenExpiredError,
    RateLimitError,
    NotFoundError,
    TimeoutError,
)

try:
    response = client.ai.chat(message="hello")
except UnauthorizedError:
    # Handle unauthorized
    pass
except TokenExpiredError:
    # Handle token expired
    pass
except RateLimitError as e:
    # Handle rate limit
    time.sleep(e.retry_after)
    # Retry request
except NotFoundError:
    # Handle not found
    pass
except TimeoutError:
    # Handle timeout
    pass
except NexusError as e:
    # Handle other errors
    print(f"Error code: {e.code}")
    print(f"Error message: {e.message}")
    print(f"Request ID: {e.request_id}")
```

### Exception Hierarchy

```
NexusError (base)
├── UnauthorizedError (401)
├── TokenExpiredError (401)
├── ForbiddenError (403)
├── TenantViolationError (403)
├── NotFoundError (404)
├── RateLimitError (429)
├── TimeoutError (504)
├── ServiceUnavailableError (503)
└── InternalServerError (500)
```

---

## 15. Context Manager

### Using with Statement

```python
with NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
) as client:
    response = client.ai.chat(message="Hello")
    print(response.answer)
# Client is automatically closed
```

### Async Context Manager

```python
async with AsyncNexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
) as client:
    response = await client.ai.chat(message="Hello")
    print(response.answer)
# Client is automatically closed
```

---

## 16. Retry & Rate Limiting

### Custom Retry Configuration

```python
from aeronexus import NexusClient
from aeronexus.retry import RetryConfig

client = NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
    retry=RetryConfig(
        max_retries=5,
        retry_delay=1.0,
        max_retry_delay=60.0,
        exponential_backoff=True,
        jitter=True,
        retry_on=[429, 500, 502, 503, 504],
    ),
)
```

### Rate Limit Handling

```python
from aeronexus import NexusClient
from aeronexus.retry import RateLimitHandler

client = NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
    rate_limit_handler=RateLimitHandler(
        auto_retry=True,
        max_wait=120.0,
        respect_headers=True,
    ),
)
```

---

## 17. Testing

### Mock Server

```python
import pytest
from unittest.mock import Mock, patch
from aeronexus import NexusClient

def test_chat():
    with patch("aeronexus.http.HTTPClient") as mock_http:
        mock_response = Mock()
        mock_response.json.return_value = {
            "conversation_id": "conv-123",
            "answer": "Test response",
            "model": "test-model",
        }
        mock_http.post.return_value = mock_response

        client = NexusClient(
            base_url="http://localhost:8080",
            api_key="test-key",
        )

        response = client.ai.chat(message="Test")

        assert response.answer == "Test response"
```

### pytest Fixture

```python
import pytest
from aeronexus import NexusClient

@pytest.fixture
def client():
    return NexusClient(
        base_url="http://localhost:8080",
        api_key="test-key",
    )

def test_chat(client):
    response = client.ai.chat(message="Test")
    assert response.answer is not None
```

### Response Recording

```python
from aeronexus.testing import record_responses, replay_responses

@record_responses("fixtures/chat_response.json")
def test_chat_recorded():
    client = NexusClient(base_url="http://localhost:8080", api_key="test-key")
    response = client.ai.chat(message="Test")
    assert response.answer == "Test response"

@replay_responses("fixtures/chat_response.json")
def test_chat_replay():
    client = NexusClient(base_url="http://localhost:8080", api_key="test-key")
    response = client.ai.chat(message="Test")
    assert response.answer == "Test response"
```

---

## 18. Examples

### Complete Example

```python
import os
import time
from aeronexus import NexusClient

def main():
    # Create client
    client = NexusClient(
        base_url="https://api.aeroxenexus.com",
        api_key=os.getenv("NEXUS_API_KEY"),
        timeout=60.0,
        max_retries=3,
    )

    # Login
    token = client.auth.login(
        email="admin@company.com",
        password=os.getenv("NEXUS_PASSWORD"),
    )
    print(f"Logged in, token expires in {token.expires_in} seconds")

    # Chat with AI
    chat_response = client.ai.chat(
        message="Analyze my customer complaints for the last week",
        agent="customer-agent",
    )
    print(f"AI Response: {chat_response.answer}")

    # Search knowledge base
    search_results = client.rag.search(
        query="network outage procedures",
        limit=5,
    )
    for result in search_results.results:
        print(f"Found: {result.title} (Score: {result.score:.2f})")

    # Execute agent
    execution = client.agents.execute(
        agent="developer-agent",
        task="Review security of authentication module",
    )
    print(f"Agent execution started: {execution.execution_id}")

    # Poll for completion
    while True:
        status = client.agents.get_execution(execution.execution_id)
        if status.status == "completed":
            print(f"Agent completed: {status.result}")
            break
        time.sleep(2)

if __name__ == "__main__":
    main()
```

### Async Example

```python
import asyncio
import os
from aeronexus import AsyncNexusClient

async def main():
    async with AsyncNexusClient(
        base_url="https://api.aeroxenexus.com",
        api_key=os.getenv("NEXUS_API_KEY"),
    ) as client:
        # Chat with AI
        response = await client.ai.chat(
            message="Hello, Nexus!",
            agent="general",
        )
        print(f"Response: {response.answer}")

        # Search knowledge
        results = await client.rag.search(
            query="network procedures",
            limit=5,
        )
        for result in results.results:
            print(f"Found: {result.title}")

if __name__ == "__main__":
    asyncio.run(main())
```

### Jupyter Notebook Example

```python
from aeronexus import NexusClient
import pandas as pd

client = NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
)

# SQL Intelligence
result = client.sql.query(
    question="Show monthly revenue for last 12 months",
    database="aeroxe_billing_db",
)

# Convert to DataFrame
df = pd.DataFrame(result.data)
print(df)

# Plot
df.plot(x="month", y="revenue", kind="bar")
```

---

## 19. Best Practices

### Connection Pooling

```python
from aeronexus import NexusClient
from aeronexus.http import ConnectionPool

# Reuse client across requests
client = NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
    connection_pool=ConnectionPool(
        max_connections=100,
        max_keepalive=10,
        keepalive_expiry=30.0,
    ),
)
```

### Thread Safety

```python
import threading
from aeronexus import NexusClient

# NexusClient is thread-safe
client = NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
)

def worker(query):
    response = client.ai.chat(message=query)
    print(response.answer)

threads = [
    threading.Thread(target=worker, args=(f"Query {i}",))
    for i in range(10)
]

for t in threads:
    t.start()
for t in threads:
    t.join()
```

### Graceful Shutdown

```python
import signal
from aeronexus import NexusClient

client = NexusClient(
    base_url="https://api.aeroxenexus.com",
    api_key="your-api-key",
)

def shutdown_handler(signum, frame):
    client.close()
    exit(0)

signal.signal(signal.SIGINT, shutdown_handler)
signal.signal(signal.SIGTERM, shutdown_handler)
```
