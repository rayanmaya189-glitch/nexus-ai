# AeroXe Nexus AI — Rust SDK

## Official Rust Client Library for AeroXe Nexus AI Platform

---

## 1. Overview

The Rust SDK provides a safe, performant, async-first client for the AeroXe Nexus AI Platform. Built with Rust 1.75+, it leverages `tokio` for async runtime and `reqwest` for HTTP.

### Crate

```
nexus-sdk
```

### Requirements

- Rust 1.75+
- Tokio async runtime

---

## 2. Installation

```toml
[dependencies]
nexus-sdk = "1.0"

# With optional features
nexus-sdk = { version = "1.0", features = ["websocket", "rustls-tls"] }
```

### Features

| Feature | Default | Description |
|---|---|---|
| `rustls-tls` | ✓ | TLS via rustls |
| `native-tls` | - | TLS via native-tls |
| `websocket` | - | WebSocket support |
| `json` | ✓ | JSON serialization |
| `stream` | ✓ | Streaming support |

---

## 3. Client Initialization

### Basic Configuration

```rust
use nexus_sdk::{Client, ClientConfig};

#[tokio::main]
async fn main() -> Result<(), nexus_sdk::Error> {
    let client = Client::builder()
        .base_url("https://api.aeroxenexus.com")
        .api_key("your-api-key")
        .build()?;

    let response = client.ai().chat(&nexus_sdk::ChatRequest {
        message: "Hello, Nexus!".to_string(),
        agent: Some("general".to_string()),
        ..Default::default()
    }).await?;

    println!("{}", response.answer);
    Ok(())
}
```

### Builder Pattern

```rust
use std::time::Duration;
use nexus_sdk::{Client, ClientConfig};

let client = Client::builder()
    .base_url("https://api.aeroxenexus.com")
    .api_key("your-api-key")
    .timeout(Duration::from_secs(30))
    .connect_timeout(Duration::from_secs(10))
    .max_retries(3)
    .retry_delay(Duration::from_secs(2))
    .max_retry_delay(Duration::from_secs(30))
    .user_agent("my-app/1.0")
    .debug(true)
    .insecure(false) // Skip TLS verification (dev only)
    .build()?;
```

### Configuration from Environment

```rust
use nexus_sdk::Client;

let client = Client::from_env()?;

// Or with custom env vars
let client = Client::builder()
    .base_url(&std::env::var("NEXUS_BASE_URL")?)
    .api_key(&std::env::var("NEXUS_API_KEY")?)
    .timeout(Duration::from_secs(
        std::env::var("NEXUS_TIMEOUT")
            .unwrap_or_else(|_| "30".to_string())
            .parse()?
    ))
    .build()?;
```

---

## 4. Authentication

### Login

```rust
let token = client.auth().login(&nexus_sdk::LoginRequest {
    email: "admin@company.com".to_string(),
    password: "password".to_string(),
}).await?;

println!("{}", token.access_token);
println!("{}", token.refresh_token);
println!("{}", token.expires_in);
```

### Token Refresh

```rust
let new_token = client.auth().refresh(&nexus_sdk::RefreshRequest {
    refresh_token: token.refresh_token,
}).await?;
```

### Auto-Refresh

The SDK automatically refreshes tokens before expiry:

```rust
let client = Client::builder()
    .base_url("https://api.aeroxenexus.com")
    .api_key("your-api-key")
    .auto_refresh(true) // Default: true
    .build()?;
```

### Logout

```rust
client.auth().logout().await?;
```

---

## 5. AI Chat

### Basic Chat

```rust
let response = client.ai().chat(&nexus_sdk::ChatRequest {
    message: "Explain my customer complaint".to_string(),
    agent: Some("customer-agent".to_string()),
    conversation_id: Some("conv-123".to_string()),
    ..Default::default()
}).await?;

println!("{}", response.conversation_id);
println!("{}", response.answer);
println!("{}", response.model);
```

### Streaming Chat

```rust
use futures::StreamExt;

let mut stream = client.ai().chat_stream(&nexus_sdk::ChatStreamRequest {
    message: "Analyze my broadband issue".to_string(),
    agent: Some("customer-agent".to_string()),
    ..Default::default()
}).await?;

while let Some(event) = stream.next().await {
    let event = event?;
    match event.type_.as_str() {
        "token" => print!("{}", event.content),
        "tool_call" => println!("\n[Tool Call] {}", event.content),
        "tool_result" => println!("[Tool Result] {}", event.content),
        "completed" => println!("\n[Stream Complete]"),
        "error" => println!("\n[Error] {}", event.content),
        _ => {}
    }
}
```

### Chat with Context

```rust
let response = client.ai().chat(&nexus_sdk::ChatRequest {
    message: "What's the status of ticket #123?".to_string(),
    agent: Some("customer-agent".to_string()),
    context: Some(serde_json::json!({
        "ticket_id": "tkt_123",
        "customer_id": "cust_456"
    })),
    model: Some("command-r7b".to_string()),
    temperature: Some(0.7),
    max_tokens: Some(1000),
    ..Default::default()
}).await?;
```

### Conversation History

```rust
let history = client.ai().get_conversation("conv-123").await?;

for msg in &history.messages {
    println!("[{}] {}: {}", msg.role, msg.timestamp, msg.content);
}
```

---

## 6. Agent Execution

### Execute Agent

```rust
let execution = client.agents().execute(&nexus_sdk::AgentExecuteRequest {
    agent: "developer-agent".to_string(),
    task: "Review this Go code for security vulnerabilities".to_string(),
    context: Some(serde_json::json!({
        "repository": "backend",
        "file": "pkg/auth/handler.go"
    })),
}).await?;

println!("{}", execution.execution_id);
println!("{}", execution.status); // "started"
```

### Get Execution Status

```rust
let status = client.agents().get_execution("exec-abc123").await?;

println!("{}", status.status);  // "completed"
println!("{}", status.result);  // "Code review complete. 3 issues found."

for step in &status.steps {
    println!("Step: {} - Status: {}", step.step, step.status);
}
```

### Stream Agent Execution

```rust
use futures::StreamExt;

let mut stream = client.agents().stream_execution("exec-abc123").await?;

while let Some(event) = stream.next().await {
    let event = event?;
    println!("[{}] {}", event.step, event.status);
}
```

### List Agents

```rust
let agents = client.agents().list(&nexus_sdk::AgentListRequest {
    page: Some(1),
    page_size: Some(20),
}).await?;

for agent in &agents.data {
    println!("Agent: {} - {}", agent.name, agent.description);
}
```

### Get Agent Details

```rust
let agent = client.agents().get("customer-agent").await?;

println!("{}", agent.name);
println!("{:?}", agent.tools);
println!("{}", agent.model);
```

---

## 7. RAG / Knowledge Management

### Upload Document

```rust
let file_bytes = tokio::fs::read("network-guide.pdf").await?;

let doc = client.rag().upload_document(&nexus_sdk::DocumentUploadRequest {
    file: file_bytes,
    file_name: "network-guide.pdf".to_string(),
    content_type: "application/pdf".to_string(),
    metadata: Some(serde_json::json!({
        "category": "network",
        "version": "1.0"
    })),
}).await?;

println!("{}", doc.document_id); // "doc-uuid"
println!("{}", doc.status);      // "processing"
```

### Get Document Status

```rust
let status = client.rag().get_document_status("doc-uuid").await?;

println!("{}", status.status);  // "completed"
println!("{}", status.chunks);  // 42
println!("{}", status.size);    // 1024000
```

### List Documents

```rust
let docs = client.rag().list_documents(&nexus_sdk::DocumentListRequest {
    page: Some(1),
    page_size: Some(20),
    status: Some("completed".to_string()),
}).await?;

for doc in &docs.data {
    println!("Document: {} - Status: {}", doc.file_name, doc.status);
}
```

### Delete Document

```rust
client.rag().delete_document("doc-uuid").await?;
```

### Search Knowledge Base

```rust
let results = client.rag().search(&nexus_sdk::SearchRequest {
    query: "How to configure ONU?".to_string(),
    limit: Some(5),
    filters: Some(serde_json::json!({
        "category": "network"
    })),
}).await?;

for result in &results.results {
    println!("Title: {}", result.title);
    println!("Score: {:.2}", result.score);
    println!("Content: {}", result.content);
    println!("Source: {}\n", result.source);
}
```

---

## 8. Vision Intelligence

### Analyze Image

```rust
let file_bytes = tokio::fs::read("router-photo.jpg").await?;

let analysis = client.vision().analyze(&nexus_sdk::ImageAnalyzeRequest {
    file: file_bytes,
    file_name: "router-photo.jpg".to_string(),
    task: Some("identify problem".to_string()),
}).await?;

println!("{}", analysis.description); // "Router LED is showing red"
println!("{}", analysis.confidence);  // 0.94
```

### OCR

```rust
let file_bytes = tokio::fs::read("document-scan.png").await?;

let ocr_result = client.vision().ocr(&nexus_sdk::OCRRequest {
    file: file_bytes,
    file_name: "document-scan.png".to_string(),
    language: Some("eng".to_string()),
}).await?;

println!("{}", ocr_result.text);
println!("{}", ocr_result.confidence);
```

---

## 9. SQL Intelligence

### Natural Language Query

```rust
let result = client.sql().query(&nexus_sdk::SQLQueryRequest {
    question: "Show monthly revenue for last 6 months".to_string(),
    database: Some("aeroxe_billing_db".to_string()),
}).await?;

println!("{}", result.sql);       // "SELECT SUM(amount)..."
println!("{}", result.row_count);  // 6

for row in &result.data {
    println!("{:?}", row);
}
```

### List Available Databases

```rust
let databases = client.sql().list_databases().await?;

for db in &databases {
    println!("Database: {} - {}", db.name, db.description);
}
```

---

## 10. Memory

### Store Memory

```rust
let memory = client.memory().store(&nexus_sdk::MemoryStoreRequest {
    user_id: "user-123".to_string(),
    memory: "Customer prefers Hindi support".to_string(),
    type_: "preference".to_string(),
    metadata: Some(serde_json::json!({
        "language": "hindi"
    })),
}).await?;

println!("{}", memory.id);
```

### Search Memory

```rust
let results = client.memory().search(&nexus_sdk::MemorySearchRequest {
    query: "customer language preference".to_string(),
    limit: Some(5),
}).await?;

for result in &results.data {
    println!("Memory: {} (Score: {:.2})", result.memory, result.score);
}
```

### Delete Memory

```rust
client.memory().delete("mem-uuid").await?;
```

---

## 11. Workflow Automation

### Start Workflow

```rust
let workflow = client.workflow().start(&nexus_sdk::WorkflowStartRequest {
    workflow: "customer-support-flow".to_string(),
    context: Some(serde_json::json!({
        "ticket_id": "tkt_123"
    })),
}).await?;

println!("{}", workflow.workflow_id); // "wf-123"
println!("{}", workflow.status);      // "running"
```

### Get Workflow Status

```rust
let status = client.workflow().get_status("wf-123").await?;

println!("{}", status.status); // "completed"
for step in &status.steps {
    println!("Step: {} - Status: {}", step.step, step.status);
}
```

### List Workflows

```rust
let workflows = client.workflow().list(&nexus_sdk::WorkflowListRequest {
    page: Some(1),
    page_size: Some(20),
}).await?;
```

---

## 12. Model Management

### List Available Models

```rust
let models = client.models().list().await?;

for model in &models {
    println!("Model: {} - Type: {} - Status: {}",
        model.name, model.type_, model.status);
}
```

---

## 13. WebSocket Client

### Connect to Chat Stream

```rust
use nexus_sdk::websocket::{WebSocketClient, Message};

let mut ws = WebSocketClient::connect(
    "wss://api.aeroxenexus.com/ws/chat/conv-123",
    Some(vec![("Authorization".to_string(), format!("Bearer {}", token.access_token))]),
).await?;

// Send message
ws.send(Message::text(serde_json::json!({
    "type": "message",
    "content": "Analyze my broadband issue"
}))).await?;

// Receive messages
while let Some(msg) = ws.next().await {
    let msg: serde_json::Value = serde_json::from_str(&msg?.to_text()?)?;
    match msg["type"].as_str() {
        Some("token") => print!("{}", msg["content"]),
        Some("tool_call") => println!("\n[Tool] {}", msg["content"]),
        Some("completed") => {
            println!("\n[Done]");
            break;
        }
        _ => {}
    }
}
```

### WebSocket Event Handler

```rust
use nexus_sdk::websocket::{WebSocketClient, EventHandler};

struct MyHandler;

#[async_trait]
impl EventHandler for MyHandler {
    async fn on_token(&self, content: &str) {
        print!("{}", content);
    }

    async fn on_tool_call(&self, content: &str) {
        println!("\n[Tool] {}", content);
    }

    async fn on_error(&self, error: &str) {
        println!("\n[Error] {}", error);
    }

    async fn on_complete(&self) {
        println!("\n[Done]");
    }
}

let mut ws = WebSocketClient::new(
    "wss://api.aeroxenexus.com/ws/chat/conv-123",
    MyHandler,
).await?;

ws.connect().await?;
```

---

## 14. Error Handling

### Error Types

```rust
use nexus_sdk::Error;

match client.ai().chat(&request).await {
    Ok(response) => {
        println!("{}", response.answer);
    }
    Err(Error::Unauthorized(msg)) => {
        eprintln!("Unauthorized: {}", msg);
    }
    Err(Error::TokenExpired(msg)) => {
        eprintln!("Token expired: {}", msg);
    }
    Err(Error::RateLimit { retry_after }) => {
        eprintln!("Rate limited, retry after {:?}', retry_after);
        tokio::time::sleep(retry_after).await;
    }
    Err(Error::NotFound(msg)) => {
        eprintln!("Not found: {}", msg);
    }
    Err(Error::Timeout(msg)) => {
        eprintln!("Timeout: {}", msg);
    }
    Err(e) => {
        eprintln!("Error: {}", e);
    }
}
```

### Error Hierarchy

```rust
pub enum Error {
    Unauthorized(String),
    TokenExpired(String),
    Forbidden(String),
    TenantViolation(String),
    NotFound(String),
    RateLimit { retry_after: Duration },
    Timeout(String),
    ServiceUnavailable(String),
    Internal(String),
    Http(reqwest::Error),
    Json(serde_json::Error),
    WebSocket(tokio_tungstenite::tungstenite::Error),
}
```

---

## 15. Retry & Rate Limiting

### Custom Retry Configuration

```rust
use nexus_sdk::retry::RetryConfig;

let client = Client::builder()
    .base_url("https://api.aeroxenexus.com")
    .api_key("your-api-key")
    .retry(RetryConfig {
        max_retries: 5,
        retry_delay: Duration::from_secs(1),
        max_retry_delay: Duration::from_secs(60),
        exponential_backoff: true,
        jitter: true,
        retry_on: vec![429, 500, 502, 503, 504],
    })
    .build()?;
```

### Rate Limit Handling

```rust
use nexus_sdk::retry::RateLimitHandler;

let client = Client::builder()
    .base_url("https://api.aeroxenexus.com")
    .api_key("your-api-key")
    .rate_limit_handler(RateLimitHandler {
        auto_retry: true,
        max_wait: Duration::from_secs(120),
        respect_headers: true,
    })
    .build()?;
```

---

## 16. Testing

### Mock Server

```rust
#[cfg(test)]
mod tests {
    use mockito::Server;
    use nexus_sdk::Client;

    #[tokio::test]
    async fn test_chat() {
        let mut server = Server::new();
        let mock = server.mock("POST", "/api/v1/ai/chat")
            .with_status(200)
            .with_body(serde_json::json!({
                "conversation_id": "conv-123",
                "answer": "Test response",
                "model": "test-model"
            }).to_string())
            .create();

        let client = Client::builder()
            .base_url(&server.url())
            .api_key("test-key")
            .build()
            .unwrap();

        let response = client.ai().chat(&nexus_sdk::ChatRequest {
            message: "Test".to_string(),
            ..Default::default()
        }).await.unwrap();

        assert_eq!(response.answer, "Test response");
        mock.assert();
    }
}
```

### WireMock

```rust
#[tokio::test]
async fn test_with_wiremock() {
    let mock_server = wiremock::MockServer::start().await;

    wiremock::Mock::given(wiremock::matchers::method("POST"))
        .and(wiremock::matchers::path("/api/v1/ai/chat"))
        .respond_with(wiremock::ResponseTemplate::new(200)
            .set_body_json(serde_json::json!({
                "conversation_id": "conv-123",
                "answer": "Mock response",
                "model": "test-model"
            })))
        .mount(&mock_server)
        .await;

    let client = Client::builder()
        .base_url(&mock_server.uri())
        .api_key("test-key")
        .build()
        .unwrap();

    let response = client.ai().chat(&nexus_sdk::ChatRequest {
        message: "Test".to_string(),
        ..Default::default()
    }).await.unwrap();

    assert_eq!(response.answer, "Mock response");
}
```

---

## 17. Examples

### Complete Example

```rust
use std::time::Duration;
use nexus_sdk::{Client, ChatRequest, AgentExecuteRequest, SearchRequest};

#[tokio::main]
async fn main() -> Result<(), nexus_sdk::Error> {
    // Create client
    let client = Client::builder()
        .base_url("https://api.aeroxenexus.com")
        .api_key(std::env::var("NEXUS_API_KEY")?)
        .timeout(Duration::from_secs(60))
        .max_retries(3)
        .retry_delay(Duration::from_secs(2))
        .build()?;

    // Login
    let token = client.auth().login(&nexus_sdk::LoginRequest {
        email: "admin@company.com".to_string(),
        password: std::env::var("NEXUS_PASSWORD")?,
    }).await?;
    println!("Logged in, token expires in {} seconds", token.expires_in);

    // Chat with AI
    let chat_response = client.ai().chat(&ChatRequest {
        message: "Analyze my customer complaints for the last week".to_string(),
        agent: Some("customer-agent".to_string()),
        ..Default::default()
    }).await?;
    println!("AI Response: {}", chat_response.answer);

    // Search knowledge base
    let search_results = client.rag().search(&SearchRequest {
        query: "network outage procedures".to_string(),
        limit: Some(5),
        ..Default::default()
    }).await?;
    for result in &search_results.results {
        println!("Found: {} (Score: {:.2})", result.title, result.score);
    }

    // Execute agent
    let execution = client.agents().execute(&AgentExecuteRequest {
        agent: "developer-agent".to_string(),
        task: "Review security of authentication module".to_string(),
        ..Default::default()
    }).await?;
    println!("Agent execution started: {}", execution.execution_id);

    // Poll for completion
    loop {
        let status = client.agents().get_execution(&execution.execution_id).await?;
        if status.status == "completed" {
            println!("Agent completed: {}", status.result);
            break;
        }
        tokio::time::sleep(Duration::from_secs(2)).await;
    }

    Ok(())
}
```

### Streaming Example

```rust
use futures::StreamExt;
use nexus_sdk::{Client, ChatStreamRequest};

#[tokio::main]
async fn main() -> Result<(), nexus_sdk::Error> {
    let client = Client::builder()
        .base_url("https://api.aeroxenexus.com")
        .api_key(std::env::var("NEXUS_API_KEY")?)
        .build()?;

    let mut stream = client.ai().chat_stream(&ChatStreamRequest {
        message: "Analyze my broadband issue".to_string(),
        agent: Some("customer-agent".to_string()),
        ..Default::default()
    }).await?;

    while let Some(event) = stream.next().await {
        let event = event?;
        match event.type_.as_str() {
            "token" => print!("{}", event.content),
            "tool_call" => println!("\n[Tool] {}", event.content),
            "completed" => {
                println!("\n[Done]");
                break;
            }
            _ => {}
        }
    }

    Ok(())
}
```

---

## 18. Best Practices

### Connection Pooling

```rust
use reqwest::Client;

// Reuse client across requests
let http_client = Client::builder()
    .pool_max_idle_per_host(10)
    .pool_idle_timeout(Duration::from_secs(90))
    .build()?;

let client = Client::builder()
    .base_url("https://api.aeroxenexus.com")
    .api_key("your-api-key")
    .http_client(http_client)
    .build()?;
```

### Error Handling with `anyhow`

```rust
use anyhow::{Context, Result};

fn main() -> Result<()> {
    let client = Client::builder()
        .base_url("https://api.aeroxenexus.com")
        .api_key("your-api-key")
        .build()
        .context("Failed to create Nexus client")?;

    let response = client.ai().chat(&ChatRequest {
        message: "Hello".to_string(),
        ..Default::default()
    })
    .await
    .context("Failed to send chat message")?;

    println!("{}", response.answer);
    Ok(())
}
```

### Graceful Shutdown

```rust
use tokio::signal;

#[tokio::main]
async fn main() -> Result<(), nexus_sdk::Error> {
    let client = Client::builder()
        .base_url("https://api.aeroxenexus.com")
        .api_key("your-api-key")
        .build()?;

    // Start background tasks
    let client_clone = client.clone();
    let handle = tokio::spawn(async move {
        // Background work
    });

    // Wait for shutdown signal
    signal::ctrl_c().await?;

    // Clean up
    client.close().await;
    handle.abort();

    Ok(())
}
```
