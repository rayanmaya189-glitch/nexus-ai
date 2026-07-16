# AeroXe Nexus AI — Node.js SDK

## Official Node.js Client Library for AeroXe Nexus AI Platform

---

## 1. Overview

The Node.js SDK provides a modern, TypeScript-first client for the AeroXe Nexus AI Platform. Built with Node.js 18+, it supports both CommonJS and ESM modules.

### Package

```
@aeronexus/sdk
```

### Requirements

- Node.js 18+
- npm 9+ or yarn 1.22+ or pnpm 8+

---

## 2. Installation

```bash
# npm
npm install @aeronexus/sdk

# yarn
yarn add @aeronexus/sdk

# pnpm
pnpm add @aeronexus/sdk
```

### Optional Dependencies

```bash
# With WebSocket support
npm install @aeronexus/sdk ws

# With keytar for secure token storage
npm install @aeronexus/sdk keytar
```

---

## 3. Client Initialization

### Basic Configuration

```typescript
import { NexusClient } from '@aeronexus/sdk';

const client = new NexusClient({
  baseUrl: 'https://api.aeroxenexus.com',
  apiKey: 'your-api-key',
});

const response = await client.ai.chat({
  message: 'Hello, Nexus!',
  agent: 'general',
});

console.log(response.answer);
```

### Configuration Options

```typescript
import { NexusClient, ClientConfig } from '@aeronexus/sdk';

const config: ClientConfig = {
  baseUrl: 'https://api.aeroxenexus.com',
  apiKey: 'your-api-key',
  timeout: 30000,
  connectTimeout: 10000,
  maxRetries: 3,
  retryDelay: 2000,
  maxRetryDelay: 30000,
  userAgent: 'my-app/1.0',
  debug: false,
  insecure: false, // Skip TLS verification (dev only)
};

const client = new NexusClient(config);
```

### Environment Variables

```typescript
import { NexusClient } from '@aeronexus/sdk';

const client = new NexusClient({
  baseUrl: process.env.NEXUS_BASE_URL || 'https://api.aeroxenexus.com',
  apiKey: process.env.NEXUS_API_KEY,
  timeout: parseInt(process.env.NEXUS_TIMEOUT || '30000'),
});
```

---

## 4. Authentication

### Login

```typescript
const token = await client.auth.login({
  email: 'admin@company.com',
  password: 'password',
});

console.log(token.accessToken);
console.log(token.refreshToken);
console.log(token.expiresIn);
```

### Token Refresh

```typescript
const newToken = await client.auth.refresh({
  refreshToken: token.refreshToken,
});
```

### Auto-Refresh

The SDK automatically refreshes tokens before expiry:

```typescript
const client = new NexusClient({
  baseUrl: 'https://api.aeroxenexus.com',
  apiKey: 'your-api-key',
  autoRefresh: true, // Default: true
});
```

### Logout

```typescript
await client.auth.logout();
```

---

## 5. AI Chat

### Basic Chat

```typescript
const response = await client.ai.chat({
  message: 'Explain my customer complaint',
  agent: 'customer-agent',
  conversationId: 'conv-123', // Optional: continue existing conversation
});

console.log(response.conversationId);
console.log(response.answer);
console.log(response.model);
```

### Streaming Chat

```typescript
const stream = await client.ai.chatStream({
  message: 'Analyze my broadband issue',
  agent: 'customer-agent',
});

for await (const event of stream) {
  switch (event.type) {
    case 'token':
      process.stdout.write(event.content);
      break;
    case 'tool_call':
      console.log(`\n[Tool Call] ${event.content}`);
      break;
    case 'tool_result':
      console.log(`[Tool Result] ${event.content}`);
      break;
    case 'completed':
      console.log('\n[Stream Complete]');
      break;
    case 'error':
      console.error(`\n[Error] ${event.content}`);
      break;
  }
}
```

### Async Iterators

```typescript
async function* chatStream(message: string) {
  const stream = await client.ai.chatStream({ message });
  yield* stream;
}

for await (const event of chatStream('Hello')) {
  if (event.type === 'token') {
    process.stdout.write(event.content);
  }
}
```

### Chat with Context

```typescript
const response = await client.ai.chat({
  message: "What's the status of ticket #123?",
  agent: 'customer-agent',
  context: {
    ticket_id: 'tkt_123',
    customer_id: 'cust_456',
  },
  model: 'command-r7b',
  temperature: 0.7,
  maxTokens: 1000,
});
```

### Conversation History

```typescript
const history = await client.ai.getConversation('conv-123');

for (const msg of history.messages) {
  console.log(`[${msg.role}] ${msg.timestamp}: ${msg.content}`);
}
```

---

## 6. Agent Execution

### Execute Agent

```typescript
const execution = await client.agents.execute({
  agent: 'developer-agent',
  task: 'Review this Go code for security vulnerabilities',
  context: {
    repository: 'backend',
    file: 'pkg/auth/handler.go',
  },
});

console.log(execution.executionId);
console.log(execution.status); // "started"
```

### Get Execution Status

```typescript
const status = await client.agents.getExecution('exec-abc123');

console.log(status.status); // "completed"
console.log(status.result); // "Code review complete. 3 issues found."

for (const step of status.steps) {
  console.log(`Step: ${step.step} - Status: ${step.status}`);
}
```

### Stream Agent Execution

```typescript
const stream = await client.agents.streamExecution('exec-abc123');

for await (const event of stream) {
  console.log(`[${event.step}] ${event.status}`);
}
```

### List Agents

```typescript
const agents = await client.agents.list({
  page: 1,
  pageSize: 20,
});

for (const agent of agents.data) {
  console.log(`Agent: ${agent.name} - ${agent.description}`);
}
```

### Get Agent Details

```typescript
const agent = await client.agents.get('customer-agent');

console.log(agent.name);
console.log(agent.tools);
console.log(agent.model);
```

---

## 7. RAG / Knowledge Management

### Upload Document

```typescript
import { readFileSync } from 'fs';

const fileBuffer = readFileSync('network-guide.pdf');

const doc = await client.rag.uploadDocument({
  file: fileBuffer,
  fileName: 'network-guide.pdf',
  contentType: 'application/pdf',
  metadata: {
    category: 'network',
    version: '1.0',
  },
});

console.log(doc.documentId); // "doc-uuid"
console.log(doc.status);      // "processing"
```

### Get Document Status

```typescript
const status = await client.rag.getDocumentStatus('doc-uuid');

console.log(status.status);  // "completed"
console.log(status.chunks);  // 42
console.log(status.size);    // 1024000
```

### List Documents

```typescript
const docs = await client.rag.listDocuments({
  page: 1,
  pageSize: 20,
  status: 'completed',
});

for (const doc of docs.data) {
  console.log(`Document: ${doc.fileName} - Status: ${doc.status}`);
}
```

### Delete Document

```typescript
await client.rag.deleteDocument('doc-uuid');
```

### Search Knowledge Base

```typescript
const results = await client.rag.search({
  query: 'How to configure ONU?',
  limit: 5,
  filters: { category: 'network' },
});

for (const result of results.results) {
  console.log(`Title: ${result.title}`);
  console.log(`Score: ${result.score.toFixed(2)}`);
  console.log(`Content: ${result.content}`);
  console.log(`Source: ${result.source}\n`);
}
```

---

## 8. Vision Intelligence

### Analyze Image

```typescript
import { readFileSync } from 'fs';

const fileBuffer = readFileSync('router-photo.jpg');

const analysis = await client.vision.analyze({
  file: fileBuffer,
  fileName: 'router-photo.jpg',
  task: 'identify problem',
});

console.log(analysis.description); // "Router LED is showing red"
console.log(analysis.confidence);  // 0.94
```

### OCR

```typescript
import { readFileSync } from 'fs';

const fileBuffer = readFileSync('document-scan.png');

const ocrResult = await client.vision.ocr({
  file: fileBuffer,
  fileName: 'document-scan.png',
  language: 'eng',
});

console.log(ocrResult.text);
console.log(ocrResult.confidence);
```

---

## 9. SQL Intelligence

### Natural Language Query

```typescript
const result = await client.sql.query({
  question: 'Show monthly revenue for last 6 months',
  database: 'aeroxe_billing_db',
});

console.log(result.sql);       // "SELECT SUM(amount)..."
console.log(result.rowCount);  // 6

for (const row of result.data) {
  console.log(row);
}
```

### List Available Databases

```typescript
const databases = await client.sql.listDatabases();

for (const db of databases) {
  console.log(`Database: ${db.name} - ${db.description}`);
}
```

---

## 10. Memory

### Store Memory

```typescript
const memory = await client.memory.store({
  userId: 'user-123',
  memory: 'Customer prefers Hindi support',
  type: 'preference',
  metadata: { language: 'hindi' },
});

console.log(memory.id);
```

### Search Memory

```typescript
const results = await client.memory.search({
  query: 'customer language preference',
  limit: 5,
});

for (const result of results.data) {
  console.log(`Memory: ${result.memory} (Score: ${result.score.toFixed(2)})`);
}
```

### Delete Memory

```typescript
await client.memory.delete('mem-uuid');
```

---

## 11. Workflow Automation

### Start Workflow

```typescript
const workflow = await client.workflow.start({
  workflow: 'customer-support-flow',
  context: { ticket_id: 'tkt_123' },
});

console.log(workflow.workflowId); // "wf-123"
console.log(workflow.status);     // "running"
```

### Get Workflow Status

```typescript
const status = await client.workflow.getStatus('wf-123');

console.log(status.status); // "completed"
for (const step of status.steps) {
  console.log(`Step: ${step.step} - Status: ${step.status}`);
}
```

### List Workflows

```typescript
const workflows = await client.workflow.list({
  page: 1,
  pageSize: 20,
});
```

---

## 12. Model Management

### List Available Models

```typescript
const models = await client.models.list();

for (const model of models) {
  console.log(`Model: ${model.name} - Type: ${model.type} - Status: ${model.status}`);
}
```

---

## 13. WebSocket Client

### Connect to Chat Stream

```typescript
import { WebSocketClient } from '@aeronexus/sdk';

const ws = new WebSocketClient({
  url: 'wss://api.aeroxenexus.com/ws/chat/conv-123',
  headers: { Authorization: `Bearer ${token.accessToken}` },
});

// Send message
await ws.send({
  type: 'message',
  content: 'Analyze my broadband issue',
});

// Receive messages
ws.on('message', (msg) => {
  switch (msg.type) {
    case 'token':
      process.stdout.write(msg.content);
      break;
    case 'tool_call':
      console.log(`\n[Tool] ${msg.content}`);
      break;
    case 'completed':
      console.log('\n[Done]');
      ws.close();
      break;
  }
});

ws.on('error', (error) => {
  console.error('WebSocket error:', error);
});

ws.on('close', () => {
  console.log('Connection closed');
});

ws.connect();
```

### Async WebSocket

```typescript
import { WebSocketClient } from '@aeronexus/sdk';

async function streamChat(message: string) {
  const ws = new WebSocketClient({
    url: 'wss://api.aeroxenexus.com/ws/chat/conv-123',
    headers: { Authorization: `Bearer ${token.accessToken}` },
  });

  await ws.connect();
  await ws.send({ type: 'message', content: message });

  for await (const msg of ws) {
    switch (msg.type) {
      case 'token':
        process.stdout.write(msg.content);
        break;
      case 'completed':
        console.log('\n[Done]');
        return;
    }
  }
}

await streamChat('Hello');
```

### WebSocket Event Handler

```typescript
import { WebSocketClient, EventHandler } from '@aeronexus/sdk';

class MyHandler implements EventHandler {
  onToken(content: string): void {
    process.stdout.write(content);
  }

  onToolCall(content: string): void {
    console.log(`\n[Tool] ${content}`);
  }

  onError(error: Error): void {
    console.error(`\n[Error] ${error.message}`);
  }

  onComplete(): void {
    console.log('\n[Done]');
  }
}

const ws = new WebSocketClient({
  url: 'wss://api.aeroxenexus.com/ws/chat/conv-123',
  handler: new MyHandler(),
});

ws.connect();
```

---

## 14. Error Handling

### Custom Error Types

```typescript
import {
  NexusError,
  UnauthorizedError,
  TokenExpiredError,
  RateLimitError,
  NotFoundError,
  TimeoutError,
} from '@aeronexus/sdk';

try {
  const response = await client.ai.chat({ message: 'hello' });
} catch (error) {
  if (error instanceof UnauthorizedError) {
    // Handle unauthorized
  } else if (error instanceof TokenExpiredError) {
    // Handle token expired
  } else if (error instanceof RateLimitError) {
    // Handle rate limit
    await sleep(error.retryAfter * 1000);
    // Retry request
  } else if (error instanceof NotFoundError) {
    // Handle not found
  } else if (error instanceof TimeoutError) {
    // Handle timeout
  } else if (error instanceof NexusError) {
    // Handle other errors
    console.error(`Error code: ${error.code}`);
    console.error(`Error message: ${error.message}`);
    console.error(`Request ID: ${error.requestId}`);
  }
}
```

### Error Hierarchy

```typescript
class NexusError extends Error {
  code: string;
  requestId: string;
}

class UnauthorizedError extends NexusError {}
class TokenExpiredError extends NexusError {}
class ForbiddenError extends NexusError {}
class TenantViolationError extends NexusError {}
class NotFoundError extends NexusError {}
class RateLimitError extends NexusError {
  retryAfter: number;
}
class TimeoutError extends NexusError {}
class ServiceUnavailableError extends NexusError {}
class InternalServerError extends NexusError {}
```

---

## 15. TypeScript Types

### Request Types

```typescript
interface ChatRequest {
  message: string;
  agent?: string;
  conversationId?: string;
  context?: Record<string, unknown>;
  model?: string;
  temperature?: number;
  maxTokens?: number;
}

interface AgentExecuteRequest {
  agent: string;
  task: string;
  context?: Record<string, unknown>;
}

interface SearchRequest {
  query: string;
  limit?: number;
  filters?: Record<string, unknown>;
}
```

### Response Types

```typescript
interface ChatResponse {
  conversationId: string;
  answer: string;
  model: string;
}

interface AgentExecution {
  executionId: string;
  status: string;
  steps: AgentStep[];
  result?: string;
}

interface SearchResult {
  title: string;
  score: number;
  content: string;
  source: string;
}
```

---

## 16. Testing

### Jest

```typescript
import { NexusClient } from '@aeronexus/sdk';

// Mock the HTTP client
jest.mock('@aeronexus/sdk', () => {
  const mockChat = jest.fn().mockResolvedValue({
    conversationId: 'conv-123',
    answer: 'Test response',
    model: 'test-model',
  });

  return {
    NexusClient: jest.fn().mockImplementation(() => ({
      ai: { chat: mockChat },
    })),
  };
});

describe('NexusClient', () => {
  it('should send chat message', async () => {
    const client = new NexusClient({
      baseUrl: 'http://localhost:8080',
      apiKey: 'test-key',
    });

    const response = await client.ai.chat({ message: 'Test' });

    expect(response.answer).toBe('Test response');
  });
});
```

### Vitest

```typescript
import { describe, it, expect, vi } from 'vitest';
import { NexusClient } from '@aeronexus/sdk';

describe('NexusClient', () => {
  it('should send chat message', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        conversation_id: 'conv-123',
        answer: 'Test response',
        model: 'test-model',
      }),
    });

    global.fetch = mockFetch;

    const client = new NexusClient({
      baseUrl: 'http://localhost:8080',
      apiKey: 'test-key',
    });

    const response = await client.ai.chat({ message: 'Test' });

    expect(response.answer).toBe('Test response');
  });
});
```

---

## 17. Examples

### Complete Example

```typescript
import { NexusClient } from '@aeronexus/sdk';

async function main() {
  // Create client
  const client = new NexusClient({
    baseUrl: 'https://api.aeroxenexus.com',
    apiKey: process.env.NEXUS_API_KEY,
    timeout: 60000,
    maxRetries: 3,
    retryDelay: 2000,
  });

  // Login
  const token = await client.auth.login({
    email: 'admin@company.com',
    password: process.env.NEXUS_PASSWORD,
  });
  console.log(`Logged in, token expires in ${token.expiresIn} seconds`);

  // Chat with AI
  const chatResponse = await client.ai.chat({
    message: 'Analyze my customer complaints for the last week',
    agent: 'customer-agent',
  });
  console.log(`AI Response: ${chatResponse.answer}`);

  // Search knowledge base
  const searchResults = await client.rag.search({
    query: 'network outage procedures',
    limit: 5,
  });
  for (const result of searchResults.results) {
    console.log(`Found: ${result.title} (Score: ${result.score.toFixed(2)})`);
  }

  // Execute agent
  const execution = await client.agents.execute({
    agent: 'developer-agent',
    task: 'Review security of authentication module',
  });
  console.log(`Agent execution started: ${execution.executionId}`);

  // Poll for completion
  while (true) {
    const status = await client.agents.getExecution(execution.executionId);
    if (status.status === 'completed') {
      console.log(`Agent completed: ${status.result}`);
      break;
    }
    await new Promise((resolve) => setTimeout(resolve, 2000));
  }
}

main().catch(console.error);
```

### Streaming Example

```typescript
import { NexusClient } from '@aeronexus/sdk';

async function streamChat() {
  const client = new NexusClient({
    baseUrl: 'https://api.aeroxenexus.com',
    apiKey: process.env.NEXUS_API_KEY,
  });

  const stream = await client.ai.chatStream({
    message: 'Analyze my broadband issue',
    agent: 'customer-agent',
  });

  for await (const event of stream) {
    switch (event.type) {
      case 'token':
        process.stdout.write(event.content);
        break;
      case 'tool_call':
        console.log(`\n[Tool] ${event.content}`);
        break;
      case 'completed':
        console.log('\n[Done]');
        break;
    }
  }
}

streamChat().catch(console.error);
```

---

## 18. Best Practices

### Connection Pooling

```typescript
import { NexusClient } from '@aeronexus/sdk';

// Reuse client across requests
const client = new NexusClient({
  baseUrl: 'https://api.aeroxenexus.com',
  apiKey: 'your-api-key',
  keepAlive: true,
  maxSockets: 100,
  keepAliveMsecs: 30000,
});
```

### Graceful Shutdown

```typescript
import { NexusClient } from '@aeronexus/sdk';

const client = new NexusClient({
  baseUrl: 'https://api.aeroxenexus.com',
  apiKey: 'your-api-key',
});

process.on('SIGINT', async () => {
  console.log('Shutting down...');
  await client.close();
  process.exit(0);
});

process.on('SIGTERM', async () => {
  console.log('Shutting down...');
  await client.close();
  process.exit(0);
});
```

### Error Boundaries

```typescript
import { NexusClient, NexusError } from '@aeronexus/sdk';

async function withRetry<T>(
  fn: () => Promise<T>,
  maxRetries = 3,
): Promise<T> {
  let lastError: Error | undefined;

  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error as Error;

      if (error instanceof NexusError) {
        if (error.code === 'RATE_LIMIT_EXCEEDED') {
          await new Promise((resolve) =>
            setTimeout(resolve, error.retryAfter * 1000)
          );
          continue;
        }

        if (error.code === 'UNAUTHORIZED' || error.code === 'FORBIDDEN') {
          throw error; // Don't retry auth errors
        }
      }

      await new Promise((resolve) =>
        setTimeout(resolve, Math.pow(2, i) * 1000)
      );
    }
  }

  throw lastError;
}

// Usage
const response = await withRetry(() =>
  client.ai.chat({ message: 'Hello' })
);
```
