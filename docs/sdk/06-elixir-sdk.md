# AeroXe Nexus AI — Elixir SDK

## Official Elixir Client Library for AeroXe Nexus AI Platform

---

## 1. Overview

The Elixir SDK provides an OTP-compliant, fault-tolerant client for the AeroXe Nexus AI Platform. Built with Elixir 1.15+, it leverages GenServer for state management, Tesla for HTTP, and websockets_client for real-time communication.

### Package

```
nexus_sdk
```

### Requirements

- Elixir 1.15+
- OTP 26+

---

## 2. Installation

### mix.exs

```elixir
defmodule MyApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :my_app,
      version: "0.1.0",
      elixir: "~> 1.15",
      deps: deps(),
    ]
  end

  def application do
    [
      extra_applications: [:logger],
      mod: {MyApp.Application, []},
    ]
  end

  defp deps do
    [
      {:nexus_sdk, "~> 1.0"},
      # Optional: for secure token storage
      {:keyring, "~> 1.0"},
    ]
  end
end
```

```bash
mix deps.get
```

---

## 3. Client Initialization

### Basic Configuration

```elixir
alias NexusSdk.Client

{:ok, client} = Client.new(
  base_url: "https://api.aeroxenexus.com",
  api_key: "your-api-key",
)

{:ok, response} = Client.ai_chat(client, %{
  message: "Hello, Nexus!",
  agent: "general",
})

IO.puts(response.answer)
```

### Configuration Options

```elixir
alias NexusSdk.Client

{:ok, client} = Client.new(
  base_url: "https://api.aeroxenexus.com",
  api_key: "your-api-key",
  timeout: 30_000,
  connect_timeout: 10_000,
  max_retries: 3,
  retry_delay: 2_000,
  max_retry_delay: 30_000,
  user_agent: "my-app/1.0",
  debug: false,
  insecure: false,  # Skip TLS verification (dev only)
)
```

### Application Configuration

```elixir
# config/config.exs
config :nexus_sdk,
  base_url: "https://api.aeroxenexus.com",
  api_key: System.get_env("NEXUS_API_KEY"),
  timeout: 30_000,
  max_retries: 3
```

```elixir
# lib/my_app/application.ex
defmodule MyApp.Application do
  use Application

  def start(_type, _args) do
    children = [
      {NexusSdk.Client, Application.get_env(:nexus_sdk, NexusSdk.Client)},
    ]

    Supervisor.start_link(children, strategy: :one_for_one)
  end
end
```

### Environment Variables

```elixir
alias NexusSdk.Client

{:ok, client} = Client.new(
  base_url: System.get_env("NEXUS_BASE_URL", "https://api.aeroxenexus.com"),
  api_key: System.get_env("NEXUS_API_KEY"),
  timeout: String.to_integer(System.get_env("NEXUS_TIMEOUT", "30000")),
)
```

---

## 4. Authentication

### Login

```elixir
alias NexusSdk.Client

{:ok, token} = Client.auth_login(client, %{
  email: "admin@company.com",
  password: "password",
})

IO.puts(token.access_token)
IO.puts(token.refresh_token)
IO.puts(token.expires_in)
```

### Token Refresh

```elixir
{:ok, new_token} = Client.auth_refresh(client, %{
  refresh_token: token.refresh_token,
})
```

### Auto-Refresh

The SDK automatically refreshes tokens before expiry:

```elixir
{:ok, client} = Client.new(
  base_url: "https://api.aeroxenexus.com",
  api_key: "your-api-key",
  auto_refresh: true,  # Default: true
)
```

### Logout

```elixir
:ok = Client.auth_logout(client)
```

---

## 5. AI Chat

### Basic Chat

```elixir
{:ok, response} = Client.ai_chat(client, %{
  message: "Explain my customer complaint",
  agent: "customer-agent",
  conversation_id: "conv-123",  # Optional: continue existing conversation
})

IO.puts(response.conversation_id)
IO.puts(response.answer)
IO.puts(response.model)
```

### Streaming Chat

```elixir
{:ok, stream} = Client.ai_chat_stream(client, %{
  message: "Analyze my broadband issue",
  agent: "customer-agent",
})

Stream.each(stream, fn event ->
  case event.type do
    "token" -> IO.write(event.content)
    "tool_call" -> IO.puts("\n[Tool Call] #{event.content}")
    "tool_result" -> IO.puts("[Tool Result] #{event.content}")
    "completed" -> IO.puts("\n[Stream Complete]")
    "error" -> IO.puts("\n[Error] #{event.content}")
  end
end)
|> Stream.run()
```

### Async Chat

```elixir
Task.async(fn ->
  {:ok, response} = Client.ai_chat(client, %{
    message: "Hello, Nexus!",
    agent: "general",
  })
  response.answer
end)
|> Task.await()
```

### Chat with Context

```elixir
{:ok, response} = Client.ai_chat(client, %{
  message: "What's the status of ticket #123?",
  agent: "customer-agent",
  context: %{
    ticket_id: "tkt_123",
    customer_id: "cust_456",
  },
  model: "command-r7b",
  temperature: 0.7,
  max_tokens: 1000,
})
```

### Conversation History

```elixir
{:ok, history} = Client.ai_get_conversation(client, "conv-123")

Enum.each(history.messages, fn msg ->
  IO.puts("[#{msg.role}] #{msg.timestamp}: #{msg.content}")
end)
```

---

## 6. Agent Execution

### Execute Agent

```elixir
{:ok, execution} = Client.agents_execute(client, %{
  agent: "developer-agent",
  task: "Review this Go code for security vulnerabilities",
  context: %{
    repository: "backend",
    file: "pkg/auth/handler.go",
  },
})

IO.puts(execution.execution_id)
IO.puts(execution.status)  # "started"
```

### Get Execution Status

```elixir
{:ok, status} = Client.agents_get_execution(client, "exec-abc123")

IO.puts(status.status)  # "completed"
IO.puts(status.result)  # "Code review complete. 3 issues found."

Enum.each(status.steps, fn step ->
  IO.puts("Step: #{step.step} - Status: #{step.status}")
end)
```

### Stream Agent Execution

```elixir
{:ok, stream} = Client.agents_stream_execution(client, "exec-abc123")

Stream.each(stream, fn event ->
  IO.puts("[#{event.step}] #{event.status}")
end)
|> Stream.run()
```

### List Agents

```elixir
{:ok, agents} = Client.agents_list(client, %{
  page: 1,
  page_size: 20,
})

Enum.each(agents.data, fn agent ->
  IO.puts("Agent: #{agent.name} - #{agent.description}")
end)
```

### Get Agent Details

```elixir
{:ok, agent} = Client.agents_get(client, "customer-agent")

IO.puts(agent.name)
IO.inspect(agent.tools)
IO.puts(agent.model)
```

---

## 7. RAG / Knowledge Management

### Upload Document

```elixir
{:ok, doc} = Client.rag_upload_document(client, %{
  file: File.read!("network-guide.pdf"),
  file_name: "network-guide.pdf",
  content_type: "application/pdf",
  metadata: %{
    category: "network",
    version: "1.0",
  },
})

IO.puts(doc.document_id)  # "doc-uuid"
IO.puts(doc.status)       # "processing"
```

### Get Document Status

```elixir
{:ok, status} = Client.rag_get_document_status(client, "doc-uuid")

IO.puts(status.status)   # "completed"
IO.puts(status.chunks)   # 42
IO.puts(status.size)     # 1024000
```

### List Documents

```elixir
{:ok, docs} = Client.rag_list_documents(client, %{
  page: 1,
  page_size: 20,
  status: "completed",
})

Enum.each(docs.data, fn doc ->
  IO.puts("Document: #{doc.file_name} - Status: #{doc.status}")
end)
```

### Delete Document

```elixir
:ok = Client.rag_delete_document(client, "doc-uuid")
```

### Search Knowledge Base

```elixir
{:ok, results} = Client.rag_search(client, %{
  query: "How to configure ONU?",
  limit: 5,
  filters: %{category: "network"},
})

Enum.each(results.results, fn result ->
  IO.puts("Title: #{result.title}")
  IO.puts("Score: #{:erlang.float_to_binary(result.score, decimals: 2)}")
  IO.puts("Content: #{result.content}")
  IO.puts("Source: #{result.source}\n")
end)
```

---

## 8. Vision Intelligence

### Analyze Image

```elixir
{:ok, analysis} = Client.vision_analyze(client, %{
  file: File.read!("router-photo.jpg"),
  file_name: "router-photo.jpg",
  task: "identify problem",
})

IO.puts(analysis.description)  # "Router LED is showing red"
IO.puts(analysis.confidence)   # 0.94
```

### OCR

```elixir
{:ok, ocr_result} = Client.vision_ocr(client, %{
  file: File.read!("document-scan.png"),
  file_name: "document-scan.png",
  language: "eng",
})

IO.puts(ocr_result.text)
IO.puts(ocr_result.confidence)
```

---

## 9. SQL Intelligence

### Natural Language Query

```elixir
{:ok, result} = Client.sql_query(client, %{
  question: "Show monthly revenue for last 6 months",
  database: "aeroxe_billing_db",
})

IO.puts(result.sql)        # "SELECT SUM(amount)..."
IO.puts(result.row_count)  # 6

Enum.each(result.data, fn row ->
  IO.inspect(row)
end)
```

### List Available Databases

```elixir
{:ok, databases} = Client.sql_list_databases(client)

Enum.each(databases, fn db ->
  IO.puts("Database: #{db.name} - #{db.description}")
end)
```

---

## 10. Memory

### Store Memory

```elixir
{:ok, memory} = Client.memory_store(client, %{
  user_id: "user-123",
  memory: "Customer prefers Hindi support",
  type: "preference",
  metadata: %{language: "hindi"},
})

IO.puts(memory.id)
```

### Search Memory

```elixir
{:ok, results} = Client.memory_search(client, %{
  query: "customer language preference",
  limit: 5,
})

Enum.each(results.data, fn result ->
  IO.puts("Memory: #{result.memory} (Score: #{result.score})")
end)
```

### Delete Memory

```elixir
:ok = Client.memory_delete(client, "mem-uuid")
```

---

## 11. Workflow Automation

### Start Workflow

```elixir
{:ok, workflow} = Client.workflow_start(client, %{
  workflow: "customer-support-flow",
  context: %{ticket_id: "tkt_123"},
})

IO.puts(workflow.workflow_id)  # "wf-123"
IO.puts(workflow.status)       # "running"
```

### Get Workflow Status

```elixir
{:ok, status} = Client.workflow_get_status(client, "wf-123")

IO.puts(status.status)  # "completed"
Enum.each(status.steps, fn step ->
  IO.puts("Step: #{step.step} - Status: #{step.status}")
end)
```

### List Workflows

```elixir
{:ok, workflows} = Client.workflow_list(client, %{
  page: 1,
  page_size: 20,
})
```

---

## 12. Model Management

### List Available Models

```elixir
{:ok, models} = Client.models_list(client)

Enum.each(models, fn model ->
  IO.puts("Model: #{model.name} - Type: #{model.type} - Status: #{model.status}")
end)
```

---

## 13. WebSocket Client

### Connect to Chat Stream

```elixir
alias NexusSdk.WebSocket

{:ok, ws} = WebSocket.connect(
  "wss://api.aeroxenexus.com/ws/chat/conv-123",
  headers: [{"Authorization", "Bearer #{token.access_token}"}],
)

# Send message
:ok = WebSocket.send(ws, %{
  type: "message",
  content: "Analyze my broadband issue",
})

# Receive messages
receive do
  {:websocket, msg} ->
    case msg.type do
      "token" -> IO.write(msg.content)
      "tool_call" -> IO.puts("\n[Tool] #{msg.content}")
      "completed" -> IO.puts("\n[Done]")
    end
after
  30_000 -> IO.puts("Timeout")
end
```

### GenServer WebSocket

```elixir
defmodule MyApp.ChatWebSocket do
  use GenServer

  alias NexusSdk.WebSocket

  def start_link(opts) do
    GenServer.start_link(__MODULE__, opts)
  end

  def init(opts) do
    {:ok, ws} = WebSocket.connect(
      opts.url,
      headers: opts.headers,
    )

    {:ok, %{ws: ws, buffer: ""}}
  end

  def send_message(pid, message) do
    GenServer.cast(pid, {:send, message})
  end

  def handle_cast({:send, message}, %{ws: ws} = state) do
    :ok = WebSocket.send(ws, %{
      type: "message",
      content: message,
    })
    {:noreply, state}
  end

  def handle_info({:websocket, msg}, state) do
    case msg.type do
      "token" ->
        IO.write(msg.content)
        {:noreply, %{state | buffer: state.buffer <> msg.content}}

      "completed" ->
        IO.puts("\n[Done]")
        {:stop, :normal, state}

      _ ->
        {:noreply, state}
    end
  end
end
```

### Phoenix Channel Integration

```elixir
defmodule MyAppWeb.ChatChannel do
  use Phoenix.Channel

  alias NexusSdk.WebSocket

  def join("chat:" <> conversation_id, _params, socket) do
    {:ok, ws} = WebSocket.connect(
      "wss://api.aeroxenexus.com/ws/chat/#{conversation_id}",
      headers: [{"Authorization", "Bearer #{socket.assigns.token}"}],
    )

    {:ok, %{ws: ws}, assign(socket, :ws, ws)}
  end

  def handle_in("message", %{"content" => content}, socket) do
    :ok = WebSocket.send(socket.assigns.ws, %{
      type: "message",
      content: content,
    })

    {:noreply, socket}
  end

  def handle_info({:websocket, msg}, socket) do
    push(socket, msg.type, msg)
    {:noreply, socket}
  end
end
```

---

## 14. Error Handling

### Custom Error Types

```elixir
alias NexusSdk.Error

case Client.ai_chat(client, %{message: "hello"}) do
  {:ok, response} ->
    IO.puts(response.answer)

  {:error, %Error{code: "UNAUTHORIZED", message: msg}} ->
    IO.puts("Unauthorized: #{msg}")

  {:error, %Error{code: "TOKEN_EXPIRED", message: msg}} ->
    IO.puts("Token expired: #{msg}")

  {:error, %Error{code: "RATE_LIMIT_EXCEEDED", retry_after: retry_after}} ->
    Process.sleep(retry_after * 1000)
    # Retry request

  {:error, %Error{code: "NOT_FOUND", message: msg}} ->
    IO.puts("Not found: #{msg}")

  {:error, %Error{code: "REQUEST_TIMEOUT", message: msg}} ->
    IO.puts("Timeout: #{msg}")

  {:error, %Error{code: code, message: msg}} ->
    IO.puts("Error [#{code}]: #{msg}")
end
```

### Error Struct

```elixir
defmodule NexusSdk.Error do
  defstruct [
    :code,
    :message,
    :request_id,
    :retry_after,
    :status,
  ]
end
```

---

## 15. Retry & Rate Limiting

### Custom Retry Configuration

```elixir
alias NexusSdk.Client

{:ok, client} = Client.new(
  base_url: "https://api.aeroxenexus.com",
  api_key: "your-api-key",
  retry: [
    max_retries: 5,
    retry_delay: 1_000,
    max_retry_delay: 60_000,
    exponential_backoff: true,
    jitter: true,
    retry_on: [429, 500, 502, 503, 504],
  ],
)
```

### Rate Limit Handling

```elixir
alias NexusSdk.Client

{:ok, client} = Client.new(
  base_url: "https://api.aeroxenexus.com",
  api_key: "your-api-key",
  rate_limit: [
    auto_retry: true,
    max_wait: 120_000,
    respect_headers: true,
  ],
)
```

---

## 16. Testing

### Mox

```elixir
# test/test_helper.exs
Mox.defmock(NexusSdk.MockHTTP, for: NexusSdk.HTTP Behaviour)

# test/my_app_test.exs
defmodule MyAppTest do
  use ExUnit.Case
  import Mox

  test "chat message" do
    expect(NexusSdk.MockHTTP, :post, fn url, body, headers ->
      assert url == "https://api.aeroxenexus.com/api/v1/ai/chat"
      assert body.message == "Test"

      {:ok, %{
        status: 200,
        body: %{
          "conversation_id" => "conv-123",
          "answer" => "Test response",
          "model" => "test-model",
        },
      }}
    end)

    {:ok, client} = Client.new(
      base_url: "https://api.aeroxenexus.com",
      api_key: "test-key",
      http: NexusSdk.MockHTTP,
    )

    {:ok, response} = Client.ai_chat(client, %{message: "Test"})

    assert response.answer == "Test response"
  end
end
```

### ExUnit

```elixir
defmodule NexusSdk.ClientTest do
  use ExUnit.Case, async: true

  alias NexusSdk.Client

  describe "ai_chat/2" do
    test "sends chat message" do
      # Use bypass or similar for HTTP mocking
      bypass = Bypass.open()

      Bypass.expect(bypass, fn conn ->
        assert conn.method == "POST"
        assert conn.request_path == "/api/v1/ai/chat"

        Plug.Conn.resp(conn, 200, Jason.encode!(%{
          "conversation_id" => "conv-123",
          "answer" => "Test response",
          "model" => "test-model",
        }))
      end)

      {:ok, client} = Client.new(
        base_url: "http://localhost:#{bypass.port}",
        api_key: "test-key",
      )

      {:ok, response} = Client.ai_chat(client, %{message: "Test"})

      assert response.answer == "Test response"

      Bypass.down(bypass)
    end
  end
end
```

---

## 17. Examples

### Complete Example

```elixir
defmodule MyApp.NexusExample do
  alias NexusSdk.Client

  def run do
    # Create client
    {:ok, client} = Client.new(
      base_url: "https://api.aeroxenexus.com",
      api_key: System.get_env("NEXUS_API_KEY"),
      timeout: 60_000,
      max_retries: 3,
      retry_delay: 2_000,
    )

    # Login
    {:ok, token} = Client.auth_login(client, %{
      email: "admin@company.com",
      password: System.get_env("NEXUS_PASSWORD"),
    })
    IO.puts("Logged in, token expires in #{token.expires_in} seconds")

    # Chat with AI
    {:ok, chat_response} = Client.ai_chat(client, %{
      message: "Analyze my customer complaints for the last week",
      agent: "customer-agent",
    })
    IO.puts("AI Response: #{chat_response.answer}")

    # Search knowledge base
    {:ok, search_results} = Client.rag_search(client, %{
      query: "network outage procedures",
      limit: 5,
    })

    Enum.each(search_results.results, fn result ->
      IO.puts("Found: #{result.title} (Score: #{result.score})")
    end)

    # Execute agent
    {:ok, execution} = Client.agents_execute(client, %{
      agent: "developer-agent",
      task: "Review security of authentication module",
    })
    IO.puts("Agent execution started: #{execution.execution_id}")

    # Poll for completion
    poll_execution(client, execution.execution_id)
  end

  defp poll_execution(client, execution_id) do
    case Client.agents_get_execution(client, execution_id) do
      {:ok, %{status: "completed"} = status} ->
        IO.puts("Agent completed: #{status.result}")

      {:ok, _status} ->
        Process.sleep(2_000)
        poll_execution(client, execution_id)

      {:error, error} ->
        IO.puts("Error: #{error.message}")
    end
  end
end
```

### Streaming Example

```elixir
defmodule MyApp.StreamExample do
  alias NexusSdk.Client

  def stream_chat do
    {:ok, client} = Client.new(
      base_url: "https://api.aeroxenexus.com",
      api_key: System.get_env("NEXUS_API_KEY"),
    )

    {:ok, stream} = Client.ai_chat_stream(client, %{
      message: "Analyze my broadband issue",
      agent: "customer-agent",
    })

    stream
    |> Stream.each(fn event ->
      case event.type do
        "token" -> IO.write(event.content)
        "tool_call" -> IO.puts("\n[Tool] #{event.content}")
        "completed" -> IO.puts("\n[Done]")
      end
    end)
    |> Stream.run()
  end
end
```

### OTP Application Example

```elixir
defmodule MyApp.Application do
  use Application

  def start(_type, _args) do
    children = [
      # Nexus SDK Client with connection pooling
      {NexusSdk.Client, [
        base_url: "https://api.aeroxenexus.com",
        api_key: System.get_env("NEXUS_API_KEY"),
        pool_size: 10,
        pool_overflow: 5,
      ]},
    ]

    opts = [strategy: :one_for_one, name: MyApp.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
```

---

## 18. Best Practices

### Connection Pooling

```elixir
alias NexusSdk.Client

# Use connection pooling for high throughput
{:ok, client} = Client.new(
  base_url: "https://api.aeroxenexus.com",
  api_key: "your-api-key",
  pool_size: 10,      # Number of connections in pool
  pool_overflow: 5,   # Additional connections when pool is full
  pool_timeout: 5_000, # Timeout to get connection from pool
)
```

### Supervision Trees

```elixir
defmodule MyApp.Supervisor do
  use Supervisor

  def start_link(opts) do
    Supervisor.start_link(__MODULE__, opts, name: __MODULE__)
  end

  @impl true
  def init(_opts) do
    children = [
      {NexusSdk.Client, [
        base_url: "https://api.aeroxenexus.com",
        api_key: System.get_env("NEXUS_API_KEY"),
      ]},
    ]

    Supervisor.init(children, strategy: :one_for_one)
  end
end
```

### Circuit Breaker

```elixir
defmodule MyApp.CircuitBreaker do
  use GenServer

  def start_link(opts) do
    GenServer.start_link(__MODULE__, opts, name: __MODULE__)
  end

  def call(func) do
    GenServer.call(__MODULE__, {:call, func})
  end

  @impl true
  def init(_opts) do
    {:ok, %{failures: 0, state: :closed}}
  end

  @impl true
  def handle_call({:call, func}, _from, %{state: :closed} = state) do
    case func.() do
      {:ok, result} ->
        {:reply, {:ok, result}, %{state | failures: 0}}

      {:error, _} = error ->
        new_failures = state.failures + 1
        new_state = if new_failures >= 5, do: :open, else: :closed
        {:reply, error, %{state | failures: new_failures, state: new_state}}
    end
  end

  def handle_call({:call, _func}, _from, %{state: :open} = state) do
    {:reply, {:error, :circuit_open}, state}
  end
end
```

### Graceful Shutdown

```elixir
defmodule MyApp.Shutdown do
  def shutdown do
    # Clean up resources
    NexusSdk.Client.close(MyApp.NexusClient)
    System.halt(0)
  end
end

# Register shutdown handler
System.at_exit(fn _ ->
  MyApp.Shutdown.shutdown()
end)
```
