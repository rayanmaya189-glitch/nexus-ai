# AeroXe Nexus AI — Go SDK

## Official Go Client Library for AeroXe Nexus AI Platform

---

## 1. Overview

The Go SDK provides an idiomatic Go client for the AeroXe Nexus AI Platform. Built with Go 1.22+, it leverages Go's concurrency model, context propagation, and interface-based design.

### Package

```
github.com/aeroxe/nexus-sdk-go
```

### Requirements

- Go 1.22+
- No external dependencies (stdlib only for core)

---

## 2. Installation

```bash
go get github.com/aeroxe/nexus-sdk-go@latest
```

---

## 3. Client Initialization

### Basic Configuration

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    nexus "github.com/aeroxe/nexus-sdk-go"
)

func main() {
    client, err := nexus.NewClient(
        nexus.WithBaseURL("https://api.aeroxenexus.com"),
        nexus.WithAPIKey("your-api-key"),
        nexus.WithTimeout(30*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    response, err := client.AI.Chat(ctx, &nexus.ChatRequest{
        Message: "Hello, Nexus!",
        Agent:   "general",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(response.Answer)
}
```

### Functional Options Pattern

```go
client, err := nexus.NewClient(
    nexus.WithBaseURL("https://api.aeroxenexus.com"),
    nexus.WithAPIKey("your-api-key"),
    nexus.WithTimeout(60*time.Second),
    nexus.WithRetry(3, 2*time.Second),
    nexus.WithRateLimit(100, 1*time.Minute),
    nexus.WithUserAgent("my-app/1.0"),
    nexus.WithDebug(true),
    nexus.WithTransport(&http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    }),
)
```

### Configuration Options

| Option | Default | Description |
|---|---|---|
| `WithBaseURL(url)` | `https://api.aeroxenexus.com` | API base URL |
| `WithAPIKey(key)` | - | API key for authentication |
| `WithTimeout(d)` | `30s` | HTTP client timeout |
| `WithRetry(max, delay)` | `3, 1s` | Retry configuration |
| `WithRateLimit(burst, window)` | - | Client-side rate limiting |
| `WithUserAgent(ua)` | `nexus-sdk-go/1.0` | Custom user agent |
| `WithDebug(bool)` | `false` | Enable debug logging |
| `WithTransport(t)` | `http.DefaultTransport` | Custom HTTP transport |
| `WithInsecure(bool)` | `false` | Skip TLS verification (dev only) |

---

## 4. Authentication

### Login

```go
ctx := context.Background()
token, err := client.Auth.Login(ctx, &nexus.LoginRequest{
    Email:    "admin@company.com",
    Password: "password",
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(token.AccessToken)
fmt.Println(token.RefreshToken)
fmt.Println(token.ExpiresIn)
```

### Token Refresh

```go
newToken, err := client.Auth.Refresh(ctx, &nexus.RefreshRequest{
    RefreshToken: token.RefreshToken,
})
if err != nil {
    log.Fatal(err)
}
```

### Auto-Refresh

The SDK automatically refreshes tokens before expiry:

```go
// Tokens are managed internally
// The SDK will refresh transparently when needed
client, err := nexus.NewClient(
    nexus.WithBaseURL("https://api.aeroxenexus.com"),
    nexus.WithAPIKey("your-api-key"),
    nexus.WithAutoRefresh(true), // Default: true
)
```

### Logout

```go
err := client.Auth.Logout(ctx)
if err != nil {
    log.Fatal(err)
}
```

---

## 5. AI Chat

### Basic Chat

```go
response, err := client.AI.Chat(ctx, &nexus.ChatRequest{
    Message:        "Explain my customer complaint",
    Agent:          "customer-agent",
    ConversationID: "conv-123", // Optional: continue existing conversation
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.ConversationID)
fmt.Println(response.Answer)
fmt.Println(response.Model)
```

### Streaming Chat

```go
stream, err := client.AI.ChatStream(ctx, &nexus.ChatStreamRequest{
    Message: "Analyze my broadband issue",
    Agent:   "customer-agent",
})
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    event, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }

    switch event.Type {
    case "token":
        fmt.Print(event.Content)
    case "tool_call":
        fmt.Printf("\n[Tool Call] %s\n", event.Content)
    case "tool_result":
        fmt.Printf("[Tool Result] %s\n", event.Content)
    case "completed":
        fmt.Println("\n[Stream Complete]")
    case "error":
        fmt.Printf("\n[Error] %s\n", event.Content)
    }
}
```

### Chat with Context

```go
response, err := client.AI.Chat(ctx, &nexus.ChatRequest{
    Message: "What's the status of ticket #123?",
    Agent:   "customer-agent",
    Context: map[string]interface{}{
        "ticket_id": "tkt_123",
        "customer_id": "cust_456",
    },
    Model: "command-r7b", // Optional: specify model
    Temperature: 0.7,      // Optional: control randomness
    MaxTokens: 1000,       // Optional: limit response length
})
```

### Conversation History

```go
// Get conversation history
history, err := client.AI.GetConversation(ctx, "conv-123")
if err != nil {
    log.Fatal(err)
}

for _, msg := range history.Messages {
    fmt.Printf("[%s] %s: %s\n", msg.Role, msg.Timestamp, msg.Content)
}
```

---

## 6. Agent Execution

### Execute Agent

```go
execution, err := client.Agents.Execute(ctx, &nexus.AgentExecuteRequest{
    Agent: "developer-agent",
    Task:  "Review this Go code for security vulnerabilities",
    Context: map[string]interface{}{
        "repository": "backend",
        "file":       "pkg/auth/handler.go",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(execution.ExecutionID)
fmt.Println(execution.Status) // "started"
```

### Get Execution Status

```go
status, err := client.Agents.GetExecution(ctx, "exec-abc123")
if err != nil {
    log.Fatal(err)
}

fmt.Println(status.Status)  // "completed"
fmt.Println(status.Result)  // "Code review complete. 3 issues found."

for _, step := range status.Steps {
    fmt.Printf("Step: %s - Status: %s\n", step.Step, step.Status)
}
```

### Stream Agent Execution

```go
stream, err := client.Agents.StreamExecution(ctx, "exec-abc123")
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    event, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("[%s] %s\n", event.Step, event.Status)
}
```

### List Agents

```go
agents, err := client.Agents.List(ctx, &nexus.AgentListRequest{
    Page:     1,
    PageSize: 20,
})
if err != nil {
    log.Fatal(err)
}

for _, agent := range agents.Data {
    fmt.Printf("Agent: %s - %s\n", agent.Name, agent.Description)
}
```

### Get Agent Details

```go
agent, err := client.Agents.Get(ctx, "customer-agent")
if err != nil {
    log.Fatal(err)
}

fmt.Println(agent.Name)
fmt.Println(agent.Tools)
fmt.Println(agent.Model)
```

---

## 7. RAG / Knowledge Management

### Upload Document

```go
file, err := os.Open("network-guide.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

doc, err := client.RAG.UploadDocument(ctx, &nexus.DocumentUploadRequest{
    File:        file,
    FileName:    "network-guide.pdf",
    ContentType: "application/pdf",
    Metadata: map[string]string{
        "category": "network",
        "version":  "1.0",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(doc.DocumentID) // "doc-uuid"
fmt.Println(doc.Status)     // "processing"
```

### Get Document Status

```go
status, err := client.RAG.GetDocumentStatus(ctx, "doc-uuid")
if err != nil {
    log.Fatal(err)
}

fmt.Println(status.Status)     // "completed"
fmt.Println(status.Chunks)     // 42
fmt.Println(status.Size)       // 1024000
```

### List Documents

```go
docs, err := client.RAG.ListDocuments(ctx, &nexus.DocumentListRequest{
    Page:     1,
    PageSize: 20,
    Status:   "completed", // Optional filter
})
if err != nil {
    log.Fatal(err)
}

for _, doc := range docs.Data {
    fmt.Printf("Document: %s - Status: %s\n", doc.FileName, doc.Status)
}
```

### Delete Document

```go
err := client.RAG.DeleteDocument(ctx, "doc-uuid")
if err != nil {
    log.Fatal(err)
}
```

### Search Knowledge Base

```go
results, err := client.RAG.Search(ctx, &nexus.SearchRequest{
    Query: "How to configure ONU?",
    Limit: 5,
    Filters: map[string]interface{}{
        "category": "network",
    },
})
if err != nil {
    log.Fatal(err)
}

for _, result := range results.Results {
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Score: %.2f\n", result.Score)
    fmt.Printf("Content: %s\n", result.Content)
    fmt.Printf("Source: %s\n\n", result.Source)
}
```

---

## 8. Vision Intelligence

### Analyze Image

```go
file, err := os.Open("router-photo.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

analysis, err := client.Vision.Analyze(ctx, &nexus.ImageAnalyzeRequest{
    File:     file,
    FileName: "router-photo.jpg",
    Task:     "identify problem",
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(analysis.Description)  // "Router LED is showing red"
fmt.Println(analysis.Confidence)   // 0.94
```

### OCR (Optical Character Recognition)

```go
file, err := os.Open("document-scan.png")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

ocrResult, err := client.Vision.OCR(ctx, &nexus.OCRRequest{
    File:     file,
    FileName: "document-scan.png",
    Language: "eng", // Optional
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(ocrResult.Text)
fmt.Println(ocrResult.Confidence)
```

---

## 9. SQL Intelligence

### Natural Language Query

```go
result, err := client.SQL.Query(ctx, &nexus.SQLQueryRequest{
    Question: "Show monthly revenue for last 6 months",
    Database: "aeroxe_billing_db", // Optional: specify database
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.SQL)     // "SELECT SUM(amount)..."
fmt.Println(result.RowCount) // 6

for _, row := range result.Data {
    fmt.Println(row)
}
```

### List Available Databases

```go
databases, err := client.SQL.ListDatabases(ctx)
if err != nil {
    log.Fatal(err)
}

for _, db := range databases {
    fmt.Printf("Database: %s - %s\n", db.Name, db.Description)
}
```

---

## 10. Memory

### Store Memory

```go
memory, err := client.Memory.Store(ctx, &nexus.MemoryStoreRequest{
    UserID:   "user-123",
    Memory:   "Customer prefers Hindi support",
    Type:     "preference",
    Metadata: map[string]interface{}{
        "language": "hindi",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(memory.ID)
```

### Search Memory

```go
results, err := client.Memory.Search(ctx, &nexus.MemorySearchRequest{
    Query: "customer language preference",
    Limit: 5,
})
if err != nil {
    log.Fatal(err)
}

for _, result := range results.Data {
    fmt.Printf("Memory: %s (Score: %.2f)\n", result.Memory, result.Score)
}
```

### Delete Memory

```go
err := client.Memory.Delete(ctx, "mem-uuid")
if err != nil {
    log.Fatal(err)
}
```

---

## 11. Workflow Automation

### Start Workflow

```go
workflow, err := client.Workflow.Start(ctx, &nexus.WorkflowStartRequest{
    Workflow: "customer-support-flow",
    Context: map[string]interface{}{
        "ticket_id": "tkt_123",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(workflow.WorkflowID) // "wf-123"
fmt.Println(workflow.Status)     // "running"
```

### Get Workflow Status

```go
status, err := client.Workflow.GetStatus(ctx, "wf-123")
if err != nil {
    log.Fatal(err)
}

fmt.Println(status.Status)  // "completed"
fmt.Println(status.Steps)   // [{step: "analyze", status: "done"}]
```

### List Workflows

```go
workflows, err := client.Workflow.List(ctx, &nexus.WorkflowListRequest{
    Page:     1,
    PageSize: 20,
})
if err != nil {
    log.Fatal(err)
}
```

---

## 12. Model Management

### List Available Models

```go
models, err := client.Models.List(ctx)
if err != nil {
    log.Fatal(err)
}

for _, model := range models {
    fmt.Printf("Model: %s - Type: %s - Status: %s\n",
        model.Name, model.Type, model.Status)
}
```

---

## 13. WebSocket Client

### Connect to Chat Stream

```go
ws, err := client.WebSocket.Connect(ctx, &nexus.WebSocketConfig{
    URL: "wss://api.aeroxenexus.com/ws/chat/conv-123",
    Headers: map[string]string{
        "Authorization": "Bearer " + token.AccessToken,
    },
})
if err != nil {
    log.Fatal(err)
}
defer ws.Close()

// Send message
err = ws.Send(&nexus.WSMessage{
    Type:    "message",
    Content: "Analyze my broadband issue",
})
if err != nil {
    log.Fatal(err)
}

// Receive messages
for {
    msg, err := ws.Recv()
    if err != nil {
        break
    }

    switch msg.Type {
    case "token":
        fmt.Print(msg.Content)
    case "tool_call":
        fmt.Printf("\n[Tool] %s\n", msg.Content)
    case "completed":
        fmt.Println("\n[Done]")
        return
    }
}
```

### WebSocket Event Handler

```go
handler := nexus.NewWSHandler().
    OnToken(func(content string) {
        fmt.Print(content)
    }).
    OnToolCall(func(content string) {
        fmt.Printf("\n[Tool] %s\n", content)
    }).
    OnError(func(err error) {
        fmt.Printf("\n[Error] %v\n", err)
    }).
    OnComplete(func() {
        fmt.Println("\n[Done]")
    })

err := ws.Listen(ctx, handler)
```

---

## 14. Error Handling

### Custom Error Types

```go
var (
    nexusErr *nexus.Error
    if errors.As(err, &nexusErr) {
        switch nexusErr.Code {
        case nexus.ErrUnauthorized:
            // Handle unauthorized
        case nexus.ErrRateLimit:
            retryAfter := nexusErr.RetryAfter
            time.Sleep(retryAfter)
            // Retry request
        case nexus.ErrNotFound:
            // Resource not found
        default:
            // Handle other errors
        }
    }
)
```

### Error Codes

| Constant | Code | HTTP Status |
|---|---|---|
| `nexus.ErrUnauthorized` | `UNAUTHORIZED` | 401 |
| `nexus.ErrTokenExpired` | `TOKEN_EXPIRED` | 401 |
| `nexus.ErrForbidden` | `FORBIDDEN` | 403 |
| `nexus.ErrNotFound` | `NOT_FOUND` | 404 |
| `nexus.ErrRateLimit` | `RATE_LIMIT_EXCEEDED` | 429 |
| `nexus.ErrTimeout` | `REQUEST_TIMEOUT` | 504 |
| `nexus.ErrInternal` | `INTERNAL_ERROR` | 500 |
| `nexus.ErrUnavailable` | `SERVICE_UNAVAILABLE` | 503 |

---

## 15. Context & Cancellation

### Request Context

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

response, err := client.AI.Chat(ctx, &nexus.ChatRequest{
    Message: "Hello",
})
```

### Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel from another goroutine
go func() {
    time.Sleep(5 * time.Second)
    cancel()
}()

response, err := client.AI.Chat(ctx, &nexus.ChatRequest{
    Message: "Long running query",
})
```

---

## 16. Concurrency

### Parallel Requests

```go
var wg sync.WaitGroup
errors := make(chan error, 3)

// Execute 3 requests in parallel
wg.Add(3)

go func() {
    defer wg.Done()
    _, err := client.AI.Chat(ctx, &nexus.ChatRequest{Message: "Query 1"})
    errors <- err
}()

go func() {
    defer wg.Done()
    _, err := client.RAG.Search(ctx, &nexus.SearchRequest{Query: "Query 2"})
    errors <- err
}()

go func() {
    defer wg.Done()
    _, err := client.SQL.Query(ctx, &nexus.SQLQueryRequest{Question: "Query 3"})
    errors <- err
}()

wg.Wait()
close(errors)

for err := range errors {
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
```

### Worker Pool

```go
type WorkerPool struct {
    tasks   chan func() error
    workers int
}

func (p *WorkerPool) Run() {
    var wg sync.WaitGroup
    for i := 0; i < p.workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range p.tasks {
                if err := task(); err != nil {
                    log.Printf("Worker error: %v", err)
                }
            }
        }()
    }
    wg.Wait()
}
```

---

## 17. Middleware

### Request Interceptor

```go
type LoggingInterceptor struct{}

func (l *LoggingInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
    start := time.Now()
    log.Printf("[%s] %s %s", req.Method, req.URL.Path, req.Header.Get("X-Request-ID"))

    resp, err := http.DefaultTransport.RoundTrip(req)

    log.Printf("[%s] %s completed in %v", req.Method, req.URL.Path, time.Since(start))
    return resp, err
}

client, _ := nexus.NewClient(
    nexus.WithBaseURL("https://api.aeroxenexus.com"),
    nexus.WithAPIKey("your-api-key"),
    nexus.WithTransport(&LoggingInterceptor{}),
)
```

### Custom Middleware

```go
type MetricsInterceptor struct {
    metrics *prometheus.CounterVec
}

func (m *MetricsInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
    m.metrics.WithLabelValues(req.Method, req.URL.Path).Inc()
    return http.DefaultTransport.RoundTrip(req)
}
```

---

## 18. Testing

### Mock Server

```go
func TestChat(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(nexus.ChatResponse{
            ConversationID: "conv-123",
            Answer:         "Test response",
            Model:          "test-model",
        })
    }))
    defer server.Close()

    client, _ := nexus.NewClient(
        nexus.WithBaseURL(server.URL),
        nexus.WithAPIKey("test-key"),
    )

    response, err := client.AI.Chat(context.Background(), &nexus.ChatRequest{
        Message: "Test",
    })

    assert.NoError(t, err)
    assert.Equal(t, "Test response", response.Answer)
}
```

### Interface-Based Mocking

```go
type AIService interface {
    Chat(ctx context.Context, req *nexus.ChatRequest) (*nexus.ChatResponse, error)
    ChatStream(ctx context.Context, req *nexus.ChatStreamRequest) (*nexus.Stream, error)
}

type MockAIService struct {
    ChatFunc func(ctx context.Context, req *nexus.ChatRequest) (*nexus.ChatResponse, error)
}

func (m *MockAIService) Chat(ctx context.Context, req *nexus.ChatRequest) (*nexus.ChatResponse, error) {
    return m.ChatFunc(ctx, req)
}

func TestWithMock(t *testing.T) {
    mock := &MockAIService{
        ChatFunc: func(ctx context.Context, req *nexus.ChatRequest) (*nexus.ChatResponse, error) {
            return &nexus.ChatResponse{Answer: "mock response"}, nil
        },
    }

    response, err := mock.Chat(context.Background(), &nexus.ChatRequest{Message: "test"})
    assert.NoError(t, err)
    assert.Equal(t, "mock response", response.Answer)
}
```

---

## 19. Examples

### Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    nexus "github.com/aeroxe/nexus-sdk-go"
)

func main() {
    // Create client
    client, err := nexus.NewClient(
        nexus.WithBaseURL("https://api.aeroxenexus.com"),
        nexus.WithAPIKey(os.Getenv("NEXUS_API_KEY")),
        nexus.WithTimeout(60*time.Second),
        nexus.WithRetry(3, 2*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Login
    token, err := client.Auth.Login(ctx, &nexus.LoginRequest{
        Email:    "admin@company.com",
        Password: os.Getenv("NEXUS_PASSWORD"),
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Logged in, token expires in %d seconds\n", token.ExpiresIn)

    // Chat with AI
    chatResponse, err := client.AI.Chat(ctx, &nexus.ChatRequest{
        Message: "Analyze my customer complaints for the last week",
        Agent:   "customer-agent",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("AI Response: %s\n", chatResponse.Answer)

    // Search knowledge base
    searchResults, err := client.RAG.Search(ctx, &nexus.SearchRequest{
        Query: "network outage procedures",
        Limit: 5,
    })
    if err != nil {
        log.Fatal(err)
    }
    for _, result := range searchResults.Results {
        fmt.Printf("Found: %s (Score: %.2f)\n", result.Title, result.Score)
    }

    // Execute agent
    execution, err := client.Agents.Execute(ctx, &nexus.AgentExecuteRequest{
        Agent: "developer-agent",
        Task:  "Review security of authentication module",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Agent execution started: %s\n", execution.ExecutionID)

    // Poll for completion
    for {
        status, err := client.Agents.GetExecution(ctx, execution.ExecutionID)
        if err != nil {
            log.Fatal(err)
        }
        if status.Status == "completed" {
            fmt.Printf("Agent completed: %s\n", status.Result)
            break
        }
        time.Sleep(2 * time.Second)
    }
}
```

---

## 20. Best Practices

### Connection Pooling

```go
// Reuse client across requests
client, _ := nexus.NewClient(
    nexus.WithBaseURL("https://api.aeroxenexus.com"),
    nexus.WithAPIKey("your-api-key"),
    nexus.WithTransport(&http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    }),
)
```

### Context Propagation

```go
// Always pass context for cancellation and timeouts
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // Propagate request context
    response, err := client.AI.Chat(ctx, &nexus.ChatRequest{
        Message: r.FormValue("message"),
    })
    // ...
}
```

### Error Handling

```go
response, err := client.AI.Chat(ctx, &nexus.ChatRequest{Message: "hello"})
if err != nil {
    var nexusErr *nexus.Error
    if errors.As(err, &nexusErr) {
        if nexusErr.Code == nexus.ErrRateLimit {
            time.Sleep(nexusErr.RetryAfter)
            // Retry
        }
    }
    log.Printf("API error: %v", err)
}
```

### Graceful Shutdown

```go
// Clean up resources on shutdown
defer client.Close()

// Or with context
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Wait for pending requests to complete
```
