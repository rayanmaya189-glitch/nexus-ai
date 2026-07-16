# AeroXe Nexus AI — Java SDK

## Official Java Client Library for AeroXe Nexus AI Platform

---

## 1. Overview

The Java SDK provides a type-safe, fluent client for the AeroXe Nexus AI Platform. Built with Java 17+, it uses modern Java features including records, sealed interfaces, and virtual threads.

### Maven Coordinates

```xml
<groupId>com.aeroxe</groupId>
<artifactId>nexus-sdk-java</artifactId>
<version>1.0.0</version>
```

### Gradle

```groovy
implementation 'com.aeroxe:nexus-sdk-java:1.0.0'
```

---

## 2. Requirements

- Java 17+
- Maven 3.8+ or Gradle 8+

### Dependencies

```xml
<dependencies>
    <dependency>
        <groupId>com.aeroxe</groupId>
        <artifactId>nexus-sdk-java</artifactId>
        <version>1.0.0</version>
    </dependency>
</dependencies>
```

---

## 3. Client Initialization

### Builder Pattern

```java
import com.aeroxe.nexus.NexusClient;
import com.aeroxe.nexus.config.ClientConfig;

NexusClient client = NexusClient.builder()
    .baseUrl("https://api.aeroxenexus.com")
    .apiKey("your-api-key")
    .timeout(Duration.ofSeconds(30))
    .retry(3, Duration.ofSeconds(2))
    .userAgent("my-app/1.0")
    .debug(true)
    .build();
```

### Configuration Options

```java
ClientConfig config = ClientConfig.builder()
    .baseUrl("https://api.aeroxenexus.com")
    .apiKey("your-api-key")
    .timeout(Duration.ofSeconds(30))
    .connectTimeout(Duration.ofSeconds(10))
    .readTimeout(Duration.ofSeconds(30))
    .writeTimeout(Duration.ofSeconds(30))
    .maxRetries(3)
    .retryDelay(Duration.ofSeconds(2))
    .maxRetryDelay(Duration.ofSeconds(30))
    .rateLimitBurst(100)
    .rateLimitWindow(Duration.ofMinutes(1))
    .userAgent("my-app/1.0")
    .debug(false)
    .insecure(false) // Skip TLS verification (dev only)
    .build();

NexusClient client = NexusClient.create(config);
```

### Environment-Based Configuration

```java
NexusClient client = NexusClient.builder()
    .baseUrl(System.getenv("NEXUS_BASE_URL"))
    .apiKey(System.getenv("NEXUS_API_KEY"))
    .timeout(Duration.ofSeconds(
        Integer.parseInt(System.getenv().getOrDefault("NEXUS_TIMEOUT", "30"))
    ))
    .build();
```

---

## 4. Authentication

### Login

```java
import com.aeroxe.nexus.models.*;

Token token = client.auth().login(
    LoginRequest.builder()
        .email("admin@company.com")
        .password("password")
        .build()
);

System.out.println(token.accessToken());
System.out.println(token.refreshToken());
System.out.println(token.expiresIn());
```

### Token Refresh

```java
Token newToken = client.auth().refresh(
    RefreshRequest.builder()
        .refreshToken(token.refreshToken())
        .build()
);
```

### Auto-Refresh

The SDK automatically refreshes tokens before expiry:

```java
NexusClient client = NexusClient.builder()
    .baseUrl("https://api.aeroxenexus.com")
    .apiKey("your-api-key")
    .autoRefresh(true) // Default: true
    .build();
```

### Logout

```java
client.auth().logout();
```

---

## 5. AI Chat

### Basic Chat

```java
ChatResponse response = client.ai().chat(
    ChatRequest.builder()
        .message("Explain my customer complaint")
        .agent("customer-agent")
        .conversationId("conv-123") // Optional: continue existing conversation
        .build()
);

System.out.println(response.conversationId());
System.out.println(response.answer());
System.out.println(response.model());
```

### Streaming Chat

```java
try (Stream<ChatEvent> stream = client.ai().chatStream(
    ChatStreamRequest.builder()
        .message("Analyze my broadband issue")
        .agent("customer-agent")
        .build()
)) {
    stream.forEach(event -> {
        switch (event.type()) {
            case "token" -> System.out.print(event.content());
            case "tool_call" -> System.out.println("\n[Tool Call] " + event.content());
            case "tool_result" -> System.out.println("[Tool Result] " + event.content());
            case "completed" -> System.out.println("\n[Stream Complete]");
            case "error" -> System.err.println("\n[Error] " + event.content());
        }
    });
}
```

### Async Chat

```java
CompletableFuture<ChatResponse> future = client.ai().chatAsync(
    ChatRequest.builder()
        .message("Hello, Nexus!")
        .agent("general")
        .build()
);

future.thenAccept(response -> {
    System.out.println("Answer: " + response.answer());
}).exceptionally(ex -> {
    System.err.println("Error: " + ex.getMessage());
    return null;
});
```

### Chat with Context

```java
ChatResponse response = client.ai().chat(
    ChatRequest.builder()
        .message("What's the status of ticket #123?")
        .agent("customer-agent")
        .context(Map.of(
            "ticket_id", "tkt_123",
            "customer_id", "cust_456"
        ))
        .model("command-r7b")
        .temperature(0.7)
        .maxTokens(1000)
        .build()
);
```

---

## 6. Agent Execution

### Execute Agent

```java
AgentExecution execution = client.agents().execute(
    AgentExecuteRequest.builder()
        .agent("developer-agent")
        .task("Review this Go code for security vulnerabilities")
        .context(Map.of(
            "repository", "backend",
            "file", "pkg/auth/handler.go"
        ))
        .build()
);

System.out.println(execution.executionId());
System.out.println(execution.status()); // "started"
```

### Get Execution Status

```java
AgentExecution status = client.agents().getExecution("exec-abc123");

System.out.println(status.status());  // "completed"
System.out.println(status.result());  // "Code review complete. 3 issues found."

status.steps().forEach(step -> {
    System.out.printf("Step: %s - Status: %s%n", step.step(), step.status());
});
```

### Stream Agent Execution

```java
try (Stream<AgentEvent> stream = client.agents().streamExecution("exec-abc123")) {
    stream.forEach(event -> {
        System.out.printf("[%s] %s%n", event.step(), event.status());
    });
}
```

### List Agents

```java
AgentListResponse agents = client.agents().list(
    AgentListRequest.builder()
        .page(1)
        .pageSize(20)
        .build()
);

agents.data().forEach(agent -> {
    System.out.printf("Agent: %s - %s%n", agent.name(), agent.description());
});
```

### Get Agent Details

```java
Agent agent = client.agents().get("customer-agent");

System.out.println(agent.name());
System.out.println(agent.tools());
System.out.println(agent.model());
```

---

## 7. RAG / Knowledge Management

### Upload Document

```java
Path filePath = Path.of("network-guide.pdf");
byte[] fileBytes = Files.readAllBytes(filePath);

DocumentUploadResponse doc = client.rag().uploadDocument(
    DocumentUploadRequest.builder()
        .file(fileBytes)
        .fileName("network-guide.pdf")
        .contentType("application/pdf")
        .metadata(Map.of(
            "category", "network",
            "version", "1.0"
        ))
        .build()
);

System.out.println(doc.documentId());
System.out.println(doc.status());
```

### Get Document Status

```java
DocumentStatus status = client.rag().getDocumentStatus("doc-uuid");

System.out.println(status.status());
System.out.println(status.chunks());
System.out.println(status.size());
```

### List Documents

```java
DocumentListResponse docs = client.rag().listDocuments(
    DocumentListRequest.builder()
        .page(1)
        .pageSize(20)
        .status("completed")
        .build()
);

docs.data().forEach(doc -> {
    System.out.printf("Document: %s - Status: %s%n", doc.fileName(), doc.status());
});
```

### Delete Document

```java
client.rag().deleteDocument("doc-uuid");
```

### Search Knowledge Base

```java
SearchResponse results = client.rag().search(
    SearchRequest.builder()
        .query("How to configure ONU?")
        .limit(5)
        .filters(Map.of("category", "network"))
        .build()
);

results.results().forEach(result -> {
    System.out.println("Title: " + result.title());
    System.out.printf("Score: %.2f%n", result.score());
    System.out.println("Content: " + result.content());
    System.out.println("Source: " + result.source());
});
```

---

## 8. Vision Intelligence

### Analyze Image

```java
Path filePath = Path.of("router-photo.jpg");
byte[] fileBytes = Files.readAllBytes(filePath);

ImageAnalysis analysis = client.vision().analyze(
    ImageAnalyzeRequest.builder()
        .file(fileBytes)
        .fileName("router-photo.jpg")
        .task("identify problem")
        .build()
);

System.out.println(analysis.description());
System.out.println(analysis.confidence());
```

### OCR

```java
Path filePath = Path.of("document-scan.png");
byte[] fileBytes = Files.readAllBytes(filePath);

OCRResult ocrResult = client.vision().ocr(
    OCRRequest.builder()
        .file(fileBytes)
        .fileName("document-scan.png")
        .language("eng")
        .build()
);

System.out.println(ocrResult.text());
System.out.println(ocrResult.confidence());
```

---

## 9. SQL Intelligence

### Natural Language Query

```java
SQLQueryResult result = client.sql().query(
    SQLQueryRequest.builder()
        .question("Show monthly revenue for last 6 months")
        .database("aeroxe_billing_db")
        .build()
);

System.out.println(result.sql());
System.out.println(result.rowCount());

result.data().forEach(row -> {
    System.out.println(row);
});
```

### List Available Databases

```java
List<DatabaseInfo> databases = client.sql().listDatabases();

databases.forEach(db -> {
    System.out.printf("Database: %s - %s%n", db.name(), db.description());
});
```

---

## 10. Memory

### Store Memory

```java
MemoryEntry memory = client.memory().store(
    MemoryStoreRequest.builder()
        .userId("user-123")
        .memory("Customer prefers Hindi support")
        .type("preference")
        .metadata(Map.of("language", "hindi"))
        .build()
);

System.out.println(memory.id());
```

### Search Memory

```java
MemorySearchResponse results = client.memory().search(
    MemorySearchRequest.builder()
        .query("customer language preference")
        .limit(5)
        .build()
);

results.data().forEach(result -> {
    System.out.printf("Memory: %s (Score: %.2f)%n", result.memory(), result.score());
});
```

### Delete Memory

```java
client.memory().delete("mem-uuid");
```

---

## 11. Workflow Automation

### Start Workflow

```java
WorkflowExecution workflow = client.workflow().start(
    WorkflowStartRequest.builder()
        .workflow("customer-support-flow")
        .context(Map.of("ticket_id", "tkt_123"))
        .build()
);

System.out.println(workflow.workflowId());
System.out.println(workflow.status());
```

### Get Workflow Status

```java
WorkflowStatus status = client.workflow().getStatus("wf-123");

System.out.println(status.status());
status.steps().forEach(step -> {
    System.out.printf("Step: %s - Status: %s%n", step.step(), step.status());
});
```

### List Workflows

```java
WorkflowListResponse workflows = client.workflow().list(
    WorkflowListRequest.builder()
        .page(1)
        .pageSize(20)
        .build()
);
```

---

## 12. Model Management

### List Available Models

```java
List<ModelInfo> models = client.models().list();

models.forEach(model -> {
    System.out.printf("Model: %s - Type: %s - Status: %s%n",
        model.name(), model.type(), model.status());
});
```

---

## 13. WebSocket Client

### Connect to Chat Stream

```java
import com.aeroxe.nexus.websocket.*;

WebSocket ws = client.webSocket().connect(
    WebSocketConfig.builder()
        .url("wss://api.aeroxenexus.com/ws/chat/conv-123")
        .header("Authorization", "Bearer " + token.accessToken())
        .build()
);

// Send message
ws.send(WSMessage.builder()
    .type("message")
    .content("Analyze my broadband issue")
    .build());

// Receive messages
ws.onMessage(event -> {
    switch (event.type()) {
        case "token" -> System.out.print(event.content());
        case "tool_call" -> System.out.println("\n[Tool] " + event.content());
        case "completed" -> System.out.println("\n[Done]");
    }
});

ws.onClose((code, reason) -> {
    System.out.printf("Connection closed: %d - %s%n", code, reason);
});

ws.onError(error -> {
    System.err.println("WebSocket error: " + error.getMessage());
});

ws.connect();
```

### WebSocket with Virtual Threads

```java
// Java 21+ virtual threads
try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
    executor.submit(() -> {
        ws.connect();
    });
}
```

---

## 14. Error Handling

### Custom Exception Types

```java
try {
    ChatResponse response = client.ai().chat(
        ChatRequest.builder().message("hello").build()
    );
} catch (UnauthorizedException e) {
    // Handle unauthorized
} catch (TokenExpiredException e) {
    // Handle token expired
} catch (RateLimitException e) {
    // Handle rate limit
    Thread.sleep(e.getRetryAfter().toMillis());
    // Retry request
} catch (NotFoundException e) {
    // Handle not found
} catch (TimeoutException e) {
    // Handle timeout
} catch (NexusException e) {
    // Handle other errors
    System.err.println("Error code: " + e.code());
    System.err.println("Error message: " + e.getMessage());
    System.err.println("Request ID: " + e.requestId());
}
```

### Exception Hierarchy

```
NexusException (base)
├── UnauthorizedException (401)
├── TokenExpiredException (401)
├── ForbiddenException (403)
├── TenantViolationException (403)
├── NotFoundException (404)
├── RateLimitException (429)
├── TimeoutException (504)
├── ServiceUnavailableException (503)
└── InternalServerException (500)
```

---

## 15. Reactive Streams

### Using with Project Reactor

```java
import reactor.core.publisher.Flux;

Flux<ChatEvent> stream = client.ai().chatStreamReactive(
    ChatStreamRequest.builder()
        .message("Analyze my data")
        .agent("data-agent")
        .build()
);

stream.filter(event -> event.type().equals("token"))
    .map(ChatEvent::content)
    .doOnNext(System.out::print)
    .doOnComplete(() -> System.out.println("\n[Done]"))
    .subscribe();
```

### Using with RxJava

```java
import io.reactivex.rxjava3.core.Observable;

Observable<ChatEvent> stream = client.ai().chatStreamRx(
    ChatStreamRequest.builder()
        .message("Analyze my data")
        .agent("data-agent")
        .build()
);

stream.filter(event -> event.type().equals("token"))
    .map(ChatEvent::content)
    .doOnNext(System.out::print)
    .doOnComplete(() -> System.out.println("\n[Done]"))
    .subscribe();
```

---

## 16. Testing

### Mock Server

```java
@Test
void testChat() throws Exception {
    MockWebServer server = new MockWebServer();
    server.enqueue(new MockResponse()
        .setBody("""
            {
                "conversation_id": "conv-123",
                "answer": "Test response",
                "model": "test-model"
            }
            """)
        .setHeader("Content-Type", "application/json"));

    NexusClient client = NexusClient.builder()
        .baseUrl(server.url("/").toString())
        .apiKey("test-key")
        .build();

    ChatResponse response = client.ai().chat(
        ChatRequest.builder().message("Test").build()
    );

    assertEquals("Test response", response.answer());
}
```

### WireMock

```java
@WireMockTest(httpPort = 8080)
void testWithWireMock() {
    stubFor(post("/api/v1/ai/chat")
        .willReturn(okJson("""
            {
                "conversation_id": "conv-123",
                "answer": "Mock response",
                "model": "test-model"
            }
            """)));

    NexusClient client = NexusClient.builder()
        .baseUrl("http://localhost:8080")
        .apiKey("test-key")
        .build();

    ChatResponse response = client.ai().chat(
        ChatRequest.builder().message("Test").build()
    );

    assertEquals("Mock response", response.answer());
}
```

---

## 17. Examples

### Complete Example

```java
package com.aeroxe.nexus.example;

import com.aeroxe.nexus.NexusClient;
import com.aeroxe.nexus.models.*;

import java.time.Duration;
import java.util.Map;

public class NexusExample {
    public static void main(String[] args) throws Exception {
        // Create client
        NexusClient client = NexusClient.builder()
            .baseUrl("https://api.aeroxenexus.com")
            .apiKey(System.getenv("NEXUS_API_KEY"))
            .timeout(Duration.ofSeconds(60))
            .retry(3, Duration.ofSeconds(2))
            .build();

        // Login
        Token token = client.auth().login(
            LoginRequest.builder()
                .email("admin@company.com")
                .password(System.getenv("NEXUS_PASSWORD"))
                .build()
        );
        System.out.printf("Logged in, token expires in %d seconds%n", token.expiresIn());

        // Chat with AI
        ChatResponse chatResponse = client.ai().chat(
            ChatRequest.builder()
                .message("Analyze my customer complaints for the last week")
                .agent("customer-agent")
                .build()
        );
        System.out.println("AI Response: " + chatResponse.answer());

        // Search knowledge base
        SearchResponse searchResults = client.rag().search(
            SearchRequest.builder()
                .query("network outage procedures")
                .limit(5)
                .build()
        );
        searchResults.results().forEach(result -> {
            System.out.printf("Found: %s (Score: %.2f)%n", result.title(), result.score());
        });

        // Execute agent
        AgentExecution execution = client.agents().execute(
            AgentExecuteRequest.builder()
                .agent("developer-agent")
                .task("Review security of authentication module")
                .build()
        );
        System.out.println("Agent execution started: " + execution.executionId());

        // Poll for completion
        while (true) {
            AgentExecution status = client.agents().getExecution(execution.executionId());
            if ("completed".equals(status.status())) {
                System.out.println("Agent completed: " + status.result());
                break;
            }
            Thread.sleep(2000);
        }
    }
}
```

---

## 18. Best Practices

### Connection Pooling

```java
// Reuse client across requests
NexusClient client = NexusClient.builder()
    .baseUrl("https://api.aeroxenexus.com")
    .apiKey("your-api-key")
    .connectionPoolSize(100)
    .keepAliveDuration(Duration.ofMinutes(5))
    .build();
```

### Thread Safety

```java
// NexusClient is thread-safe
// Share a single instance across threads
ExecutorService executor = Executors.newFixedThreadPool(10);

IntStream.range(0, 100).forEach(i -> {
    executor.submit(() -> {
        client.ai().chat(
            ChatRequest.builder()
                .message("Query " + i)
                .build()
        );
    });
});
```

### Graceful Shutdown

```java
// Clean up resources on shutdown
Runtime.getRuntime().addShutdownHook(new Thread(() -> {
    client.close();
}));
```
