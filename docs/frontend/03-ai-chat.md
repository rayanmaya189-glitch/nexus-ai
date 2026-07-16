# AeroXe Nexus AI — AI Chat

## Chat UI, WebSocket Streaming, Message Types, Conversation Management, Real-Time Rendering

---

## 1. Chat Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                        Chat Page Layout                        │
│                                                                │
│  ┌────────────┐ ┌────────────────────────────────────────────┐│
│  │ Conversation│ │  Chat Header                               ││
│  │  Sidebar    │ │  ┌──────────────────────────────────────┐  ││
│  │             │ │  │ AgentSelector │ ModelIndicator │ ⚙   │  ││
│  │  ┌───────┐ │ │  └──────────────────────────────────────┘  ││
│  │  │ Chat 1│ │ ├────────────────────────────────────────────┤│
│  │  │ Chat 2│ │ │  MessageList (scrollable)                   ││
│  │  │ Chat 3│ │ │  ┌──────────────────────────────────────┐  ││
│  │  │ ...   │ │ │  │  User: "Analyze customer #1234"      │  ││
│  │  └───────┘ │ │  ├──────────────────────────────────────┤  ││
│  │             │ │  │  🤖 Thinking...                      │  ││
│  │  Search     │ │  │  ✓ customer.lookup(#1234)           │  ││
│  │  Filter     │ │  │  ✓ billing.check(#1234)             │  ││
│  │             │ │  │  AI: "Customer has 3 open tickets..."│  ││
│  │  ┌───────┐ │ │  └──────────────────────────────────────┘  ││
│  │  │ New   │ │ ├────────────────────────────────────────────┤│
│  │  │ Chat  │ │ │  PromptBox                                 ││
│  │  └───────┘ │ │  ┌──────────────────────────────────────┐  ││
│  └────────────┘ │  │ 📎 🎤 Type your message...     Send  │  ││
│                  │  └──────────────────────────────────────┘  ││
│                  └────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
```

---

## 2. Component Architecture

### 2.1 Component Hierarchy

```
ChatPage
├── ConversationSidebar
│   ├── SearchInput
│   ├── ConversationList
│   │   └── ConversationItem[]
│   └── NewChatButton
│
├── ChatWindow
│   ├── ChatHeader
│   │   ├── AgentSelector
│   │   ├── ModelIndicator
│   │   └── ChatSettings
│   │
│   ├── MessageList
│   │   ├── MessageBubble[] (user messages)
│   │   ├── AIResponse[] (assistant messages)
│   │   │   ├── ThinkingIndicator
│   │   │   ├── ToolExecutionDisplay[]
│   │   │   ├── TokenStreamViewer
│   │   │   └── MarkdownRenderer
│   │   ├── FileUploader (attached files)
│   │   └── ScrollToBottomButton
│   │
│   └── PromptBox
│       ├── FileUploader (inline)
│       ├── VoiceInput
│       ├── TextInput
│       └── SendButton
│
└── ChatSettings (sheet)
    ├── ModelSelection
    ├── TemperatureSlider
    └── MaxTokensInput
```

### 2.2 Component Specifications

| Component | Purpose | Key Features |
|---|---|---|
| `ChatWindow` | Main chat container | Layout, state coordination |
| `MessageList` | Scrollable message area | Auto-scroll, virtualization |
| `MessageBubble` | Single message display | User vs AI styling |
| `PromptBox` | User input area | Multi-line, keyboard shortcuts |
| `FileUploader` | File attachment | Drag-and-drop, progress |
| `VoiceInput` | Voice-to-text | Microphone toggle, recording |
| `AgentSelector` | Choose agent | Dropdown with agent cards |
| `ModelIndicator` | Show active model | Name, status, latency |
| `TokenStreamViewer` | Real-time token rendering | Streaming cursor, markdown |
| `ToolExecutionDisplay` | Tool call visualization | Steps, status, collapsible |
| `ThinkingIndicator` | AI thinking state | Animated, contextual text |
| `ConversationSidebar` | History panel | Search, filter, list |
| `ChatSettings` | Configuration panel | Model, temperature, limits |

---

## 3. Chat Store (Zustand)

### 3.1 Store Definition

```typescript
// stores/chat-store.ts
interface ChatState {
  // Conversations
  conversations: Conversation[];
  activeConversationId: string | null;

  // Messages
  messages: Message[];

  // Streaming
  isStreaming: boolean;
  streamingMessageId: string | null;
  streamingContent: string;

  // UI State
  sidebarOpen: boolean;
  settingsOpen: boolean;

  // Agent/Model
  selectedAgentId: string | null;
  selectedModel: string;

  // Actions
  setActiveConversation: (id: string | null) => void;
  addMessage: (message: Message) => void;
  updateMessage: (id: string, updates: Partial<Message>) => void;
  appendToStreamingMessage: (token: string) => void;
  setStreaming: (value: boolean, messageId?: string) => void;
  setToolExecution: (messageId: string, execution: ToolExecution) => void;
  clearConversation: () => void;
  deleteConversation: (id: string) => void;
  setSelectedAgent: (agentId: string) => void;
  setSelectedModel: (model: string) => void;
  toggleSidebar: () => void;
}
```

### 3.2 Message Types

```typescript
// types/chat.types.ts
type MessageType = "user" | "assistant" | "system" | "tool_result";
type MessageStatus = "sending" | "streaming" | "completed" | "error";

interface Message {
  id: string;
  conversationId: string;
  type: MessageType;
  content: string;
  status: MessageStatus;
  agent?: string;
  model?: string;
  toolCalls?: ToolCall[];
  tokenUsage?: TokenUsage;
  createdAt: string;
  metadata?: Record<string, unknown>;
}

interface ToolCall {
  id: string;
  name: string;
  arguments: Record<string, unknown>;
  result?: unknown;
  status: "pending" | "running" | "completed" | "error";
  startedAt?: string;
  completedAt?: string;
  error?: string;
}

interface TokenUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
}

interface Conversation {
  id: string;
  title: string;
  agent: string;
  model: string;
  messageCount: number;
  lastMessageAt: string;
  createdAt: string;
}
```

### 3.3 Message Flow State Machine

```
                    ┌─────────────┐
                    │   IDLE      │
                    └──────┬──────┘
                           │ User sends message
                           ▼
                    ┌─────────────┐
                    │  SENDING    │
                    └──────┬──────┘
                           │ Server acknowledged
                           ▼
                    ┌─────────────┐
                    │  THINKING   │ ← Planning, intent detection
                    └──────┬──────┘
                           │ Agent selected
                           ▼
                    ┌─────────────┐
                    │ TOOL_CALL   │ ← Tool execution phase
                    └──────┬──────┘
                           │ Tool result received
                           ▼
                    ┌─────────────┐
                    │  STREAMING  │ ← Token-by-token response
                    └──────┬──────┘
                           │ Stream completed
                           ▼
                    ┌─────────────┐
                    │  COMPLETED  │
                    └─────────────┘
```

---

## 4. WebSocket Connection Management

### 4.1 WebSocket Client

```typescript
// lib/websocket.ts
class NexusWebSocket {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private messageHandlers: Map<string, (data: unknown) => void> = new Map();

  connect(conversationId: string) {
    const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL}/ws/chat`;
    const token = useAuthStore.getState().accessToken;

    this.ws = new WebSocket(`${wsUrl}?token=${token}&conversation=${conversationId}`);

    this.ws.onopen = () => {
      console.log("[WS] Connected");
      this.reconnectAttempts = 0;
      this.startHeartbeat();
      this.send({ type: "authenticate", token });
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };

    this.ws.onclose = (event) => {
      console.log("[WS] Disconnected", event.code);
      this.stopHeartbeat();
      if (event.code !== 1000) {
        this.attemptReconnect(conversationId);
      }
    };

    this.ws.onerror = (error) => {
      console.error("[WS] Error:", error);
    };
  }

  private handleMessage(message: WSInbound) {
    const handler = this.messageHandlers.get(message.type);
    if (handler) handler(message);

    // Global handlers
    switch (message.type) {
      case "token":
        useChatStore.getState().appendToStreamingMessage(message.content);
        break;
      case "thinking":
        useChatStore.getState().setThinking(message.content);
        break;
      case "tool_call":
        useChatStore.getState().addToolCall(message);
        break;
      case "tool_result":
        useChatStore.getState().updateToolCall(message);
        break;
      case "completed":
        useChatStore.getState().setStreaming(false);
        break;
      case "error":
        useChatStore.getState().setError(message);
        break;
      case "pong":
        // Heartbeat acknowledged
        break;
    }
  }

  send(data: WSOutbound) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  private attemptReconnect(conversationId: string) {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      useChatStore.getState().setConnectionError("Unable to reconnect. Please refresh.");
      return;
    }

    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts),
      8000
    );

    this.reconnectAttempts++;
    console.log(`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

    setTimeout(() => this.connect(conversationId), delay);
  }

  private startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      this.send({ type: "ping" });
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  disconnect() {
    this.stopHeartbeat();
    this.ws?.close(1000, "User disconnect");
    this.ws = null;
  }
}

export const wsClient = new NexusWebSocket();
```

### 4.2 Connection States

| State | Visual Indicator | Behavior |
|---|---|---|
| Connecting | Blue pulsing dot | Attempting connection |
| Connected | Green dot | Ready for messages |
| Reconnecting | Yellow pulsing dot | Auto-retry with backoff |
| Disconnected | Red dot | Manual reconnect required |
| Error | Red dot + toast | Show error, offer retry |

---

## 5. Chat UI Components

### 5.1 PromptBox Component

```tsx
// components/chat/PromptBox.tsx
interface PromptBoxProps {
  onSend: (message: string, files?: File[]) => void;
  disabled?: boolean;
  placeholder?: string;
}

export function PromptBox({ onSend, disabled, placeholder }: PromptBoxProps) {
  const [input, setInput] = useState("");
  const [files, setFiles] = useState<File[]>([]);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const isStreaming = useChatStore((s) => s.isStreaming);

  // Auto-resize textarea
  useEffect(() => {
    const el = textareaRef.current;
    if (el) {
      el.style.height = "auto";
      el.style.height = Math.min(el.scrollHeight, 200) + "px";
    }
  }, [input]);

  // Keyboard shortcut: Enter to send, Shift+Enter for newline
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleSend = () => {
    if (!input.trim() && files.length === 0) return;
    onSend(input, files);
    setInput("");
    setFiles([]);
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }
  };

  return (
    <div className="border-t bg-background p-4">
      {/* Attached files preview */}
      {files.length > 0 && (
        <div className="mb-3 flex flex-wrap gap-2">
          {files.map((file, i) => (
            <Badge key={i} variant="secondary" className="gap-1">
              <FileText className="h-3 w-3" />
              {file.name}
              <button onClick={() => removeFile(i)}>
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}

      <div className="flex items-end gap-2">
        {/* File upload button */}
        <FileUploader onFilesSelected={setFiles} />

        {/* Voice input button */}
        <VoiceInput onTranscript={(text) => setInput((prev) => prev + text)} />

        {/* Text input */}
        <div className="relative flex-1">
          <textarea
            ref={textareaRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder || "Type your message... (Shift+Enter for new line)"}
            disabled={disabled || isStreaming}
            rows={1}
            className="w-full resize-none rounded-lg border bg-background px-4 py-3 pr-12 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        {/* Send button */}
        <Button
          onClick={handleSend}
          disabled={disabled || isStreaming || (!input.trim() && files.length === 0)}
          size="icon"
          className="shrink-0"
        >
          {isStreaming ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Send className="h-4 w-4" />
          )}
        </Button>
      </div>
    </div>
  );
}
```

### 5.2 MessageBubble Component

```tsx
// components/chat/MessageBubble.tsx
interface MessageBubbleProps {
  message: Message;
}

export function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.type === "user";

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className={cn(
        "flex gap-3",
        isUser ? "justify-end" : "justify-start"
      )}
    >
      {/* Avatar */}
      {!isUser && (
        <Avatar className="h-8 w-8 shrink-0">
          <AvatarFallback className="bg-accent text-accent-foreground">
            <Bot className="h-4 w-4" />
          </AvatarFallback>
        </Avatar>
      )}

      {/* Content */}
      <div
        className={cn(
          "max-w-[70%] rounded-lg px-4 py-3",
          isUser
            ? "bg-primary text-primary-foreground"
            : "bg-muted"
        )}
      >
        {isUser ? (
          <p className="text-sm whitespace-pre-wrap">{message.content}</p>
        ) : (
          <>
            {/* Thinking phase */}
            {message.thinking && (
              <ThinkingIndicator content={message.thinking} />
            )}

            {/* Tool executions */}
            {message.toolCalls?.map((tc) => (
              <ToolExecutionDisplay key={tc.id} toolCall={tc} />
            ))}

            {/* Response content */}
            <MarkdownRenderer content={message.content} />

            {/* Streaming cursor */}
            {message.status === "streaming" && (
              <span className="streaming-cursor" />
            )}

            {/* Token usage */}
            {message.tokenUsage && (
              <div className="mt-2 text-xs text-muted-foreground">
                {message.tokenUsage.totalTokens} tokens
              </div>
            )}
          </>
        )}
      </div>

      {/* User avatar */}
      {isUser && (
        <Avatar className="h-8 w-8 shrink-0">
          <AvatarFallback>
            <User className="h-4 w-4" />
          </AvatarFallback>
        </Avatar>
      )}
    </motion.div>
  );
}
```

### 5.3 ThinkingIndicator Component

```tsx
// components/chat/ThinkingIndicator.tsx
const thinkingPhases = [
  "Planning request...",
  "Analyzing intent...",
  "Selecting agent...",
  "Searching knowledge base...",
  "Executing tools...",
  "Generating response...",
];

export function ThinkingIndicator({ content }: { content: string }) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="mb-3 flex items-center gap-2 text-sm text-muted-foreground"
    >
      <div className="flex gap-1">
        <span className="h-2 w-2 animate-bounce rounded-full bg-accent [animation-delay:-0.3s]" />
        <span className="h-2 w-2 animate-bounce rounded-full bg-accent [animation-delay:-0.15s]" />
        <span className="h-2 w-2 animate-bounce rounded-full bg-accent" />
      </div>
      <span>{content || "Thinking..."}</span>
    </motion.div>
  );
}
```

### 5.4 ToolExecutionDisplay Component

```tsx
// components/chat/ToolExecutionDisplay.tsx
interface ToolExecutionDisplayProps {
  toolCall: ToolCall;
}

const statusIcons = {
  pending: <Clock className="h-4 w-4 text-muted-foreground" />,
  running: <Loader2 className="h-4 w-4 animate-spin text-blue-500" />,
  completed: <Check className="h-4 w-4 text-green-500" />,
  error: <X className="h-4 w-4 text-red-500" />,
};

export function ToolExecutionDisplay({ toolCall }: ToolExecutionDisplayProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <Collapsible open={expanded} onOpenChange={setExpanded}>
      <CollapsibleTrigger asChild>
        <button className="flex w-full items-center gap-2 rounded-md border p-2 text-sm hover:bg-muted/50">
          {statusIcons[toolCall.status]}
          <span className="font-mono text-xs">{toolCall.name}</span>
          <ChevronDown className={cn("h-3 w-3 ml-auto transition-transform", expanded && "rotate-180")} />
        </button>
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-2 space-y-2">
        {/* Arguments */}
        <div className="rounded-md bg-muted p-2 text-xs">
          <p className="mb-1 font-medium text-muted-foreground">Arguments:</p>
          <pre className="overflow-x-auto">{JSON.stringify(toolCall.arguments, null, 2)}</pre>
        </div>
        {/* Result */}
        {toolCall.result && (
          <div className="rounded-md bg-muted p-2 text-xs">
            <p className="mb-1 font-medium text-muted-foreground">Result:</p>
            <pre className="overflow-x-auto">{JSON.stringify(toolCall.result, null, 2)}</pre>
          </div>
        )}
        {/* Error */}
        {toolCall.error && (
          <div className="rounded-md bg-destructive/10 p-2 text-xs text-destructive">
            <p className="font-medium">Error:</p>
            <p>{toolCall.error}</p>
          </div>
        )}
      </CollapsibleContent>
    </Collapsible>
  );
}
```

### 5.5 TokenStreamViewer Component

```tsx
// components/chat/TokenStreamViewer.tsx
// Renders streaming tokens with markdown support and performance optimization

export function TokenStreamViewer({ content, isStreaming }: {
  content: string;
  isStreaming: boolean;
}) {
  // Buffer management: only re-render markdown every 100ms during streaming
  const [displayContent, setDisplayContent] = useState(content);

  useEffect(() => {
    if (!isStreaming) {
      setDisplayContent(content);
      return;
    }

    const timer = setTimeout(() => {
      setDisplayContent(content);
    }, 100);

    return () => clearTimeout(timer);
  }, [content, isStreaming]);

  return (
    <div className="prose prose-sm dark:prose-invert max-w-none">
      <MarkdownRenderer content={displayContent} />
      {isStreaming && <span className="streaming-cursor" />}
    </div>
  );
}
```

---

## 6. Streaming Protocol

### 6.1 Token-by-Token Rendering

```
Server sends:                    Client renders:
─────────────                    ──────────────

{ type: "token",                │ (no change)
  content: "Customer" }         │ "Customer"

{ type: "token",                │ (debounce buffer)
  content: " has" }             │ "Customer has"

{ type: "token",                │ (flush buffer)
  content: " 3 open" }          │ "Customer has 3 open"

{ type: "tool_call",            │ Show tool call card
  name: "billing.check" }       │ "Customer has 3 open"

{ type: "tool_result",          │ Update tool card
  result: { total: 142 } }      │ ✓ billing.check

{ type: "token",                │ Continue rendering
  content: " tickets" }         │ "Customer has 3 open tickets"
```

### 6.2 Buffer Strategy

| Phase | Buffer Interval | Rationale |
|---|---|---|
| First 500ms | Immediate | Fast initial response |
| After 500ms | 50ms batch | Reduce DOM thrashing |
| Markdown heavy | 100ms batch | Avoid partial markdown |
| Tool calls | Immediate | Show tool execution |
| Completion | Immediate | Final render |

---

## 7. Conversation Management

### 7.1 Conversation Sidebar

```tsx
// components/chat/ConversationSidebar.tsx
export function ConversationSidebar() {
  const { conversations, activeConversationId, setActiveConversation } = useChatStore();
  const [searchQuery, setSearchQuery] = useState("");

  const filteredConversations = conversations.filter((c) =>
    c.title.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const groupedConversations = useMemo(() => {
    return groupByDate(filteredConversations);
  }, [filteredConversations]);

  return (
    <div className="flex h-full w-60 flex-col border-r">
      {/* New Chat Button */}
      <div className="p-3">
        <Button onClick={createNewChat} className="w-full" variant="outline">
          <Plus className="mr-2 h-4 w-4" />
          New Chat
        </Button>
      </div>

      {/* Search */}
      <div className="px-3 pb-2">
        <SearchInput
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="Search conversations..."
        />
      </div>

      {/* Conversation List */}
      <ScrollArea className="flex-1">
        {Object.entries(groupedConversations).map(([date, convs]) => (
          <div key={date}>
            <p className="px-3 py-2 text-xs font-medium text-muted-foreground">{date}</p>
            {convs.map((conv) => (
              <button
                key={conv.id}
                onClick={() => setActiveConversation(conv.id)}
                className={cn(
                  "flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-muted",
                  conv.id === activeConversationId && "bg-muted"
                )}
              >
                <MessageSquare className="h-4 w-4 shrink-0" />
                <span className="truncate">{conv.title}</span>
              </button>
            ))}
          </div>
        ))}
      </ScrollArea>
    </div>
  );
}
```

### 7.2 Conversation Operations

| Operation | Method | Endpoint |
|---|---|---|
| List conversations | `GET` | `/api/v1/chat/conversations` |
| Get conversation | `GET` | `/api/v1/chat/conversations/:id` |
| Create conversation | `POST` | `/api/v1/chat/conversations` |
| Update title | `PUT` | `/api/v1/chat/conversations/:id` |
| Delete conversation | `DELETE` | `/api/v1/chat/conversations/:id` |
| Get messages | `GET` | `/api/v1/chat/conversations/:id/messages` |

### 7.3 Auto-Title Generation

```typescript
// After first exchange, auto-generate conversation title
async function generateConversationTitle(conversationId: string, firstMessage: string) {
  const response = await api.post("/api/v1/ai/chat", {
    message: `Generate a short title (max 50 chars) for this conversation: "${firstMessage}"`,
    agent: "planner",
    system_prompt: "Respond with ONLY the title, no quotes or punctuation.",
  });
  await api.patch(`/api/v1/chat/conversations/${conversationId}`, {
    title: response.data.answer,
  });
}
```

---

## 8. File Upload in Chat

### 8.1 FileUploader Component

```tsx
// components/chat/FileUploader.tsx
interface FileUploaderProps {
  onFilesSelected: (files: File[]) => void;
  maxFiles?: number;
  maxSize?: number; // bytes
  accept?: string;
}

export function FileUploader({ onFilesSelected, maxFiles = 5, maxSize = 10 * 1024 * 1024, accept }: FileUploaderProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<Map<string, number>>(new Map());
  const inputRef = useRef<HTMLInputElement>(null);

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setIsDragging(true);
    } else if (e.type === "dragleave") {
      setIsDragging(false);
    }
  };

  const processFiles = (fileList: FileList) => {
    const validFiles = Array.from(fileList)
      .filter((f) => f.size <= maxSize)
      .slice(0, maxFiles);

    if (validFiles.length < fileList.length) {
      toast.warning(`${fileList.length - validFiles.length} files were too large and skipped.`);
    }

    onFilesSelected(validFiles);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files) processFiles(e.dataTransfer.files);
  };

  return (
    <>
      <input
        ref={inputRef}
        type="file"
        multiple
        accept={accept}
        className="hidden"
        onChange={(e) => e.target.files && processFiles(e.target.files)}
      />
      <Button
        variant="ghost"
        size="icon"
        onClick={() => inputRef.current?.click()}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
        aria-label="Upload files"
      >
        <Paperclip className="h-4 w-4" />
      </Button>
    </>
  );
}
```

### 8.2 Supported File Types

| Type | Extensions | Max Size | Processing |
|---|---|---|---|
| Documents | .pdf, .docx, .txt, .md | 10MB | RAG ingestion |
| Images | .jpg, .png, .gif, .webp | 10MB | Vision analysis |
| Code | .py, .js, .ts, .go, .rs | 5MB | Code review |
| Data | .csv, .json, .xlsx | 10MB | Data analysis |

---

## 9. Voice Input

### 9.1 VoiceInput Component

```tsx
// components/chat/VoiceInput.tsx
export function VoiceInput({ onTranscript }: { onTranscript: (text: string) => void }) {
  const [isRecording, setIsRecording] = useState(false);
  const [isSupported, setIsSupported] = useState(false);
  const recognitionRef = useRef<SpeechRecognition | null>(null);

  useEffect(() => {
    const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
    setIsSupported(!!SpeechRecognition);
  }, []);

  const toggleRecording = () => {
    if (isRecording) {
      recognitionRef.current?.stop();
      setIsRecording(false);
    } else {
      const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
      const recognition = new SpeechRecognition();
      recognition.continuous = false;
      recognition.interimResults = false;
      recognition.lang = "en-US";

      recognition.onresult = (event) => {
        const transcript = event.results[0][0].transcript;
        onTranscript(transcript);
      };

      recognition.onerror = () => {
        setIsRecording(false);
      };

      recognition.onend = () => {
        setIsRecording(false);
      };

      recognition.start();
      recognitionRef.current = recognition;
      setIsRecording(true);
    }
  };

  if (!isSupported) return null;

  return (
    <Button
      variant={isRecording ? "destructive" : "ghost"}
      size="icon"
      onClick={toggleRecording}
      aria-label={isRecording ? "Stop recording" : "Start voice input"}
    >
      {isRecording ? (
        <MicOff className="h-4 w-4 animate-pulse" />
      ) : (
        <Mic className="h-4 w-4" />
      )}
    </Button>
  );
}
```

---

## 10. Agent Selector

### 10.1 AgentSelector Component

```tsx
// components/chat/AgentSelector.tsx
interface AgentInfo {
  id: string;
  name: string;
  model: string;
  description: string;
  capabilities: string[];
  status: "active" | "inactive";
}

export function AgentSelector({ agents, selected, onSelect }: {
  agents: AgentInfo[];
  selected: string;
  onSelect: (agentId: string) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <Bot className="h-4 w-4" />
          {agents.find((a) => a.id === selected)?.name || "Select Agent"}
          <ChevronDown className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-80">
        <DropdownMenuLabel>Select Agent</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {agents.map((agent) => (
          <DropdownMenuItem
            key={agent.id}
            onClick={() => onSelect(agent.id)}
            className="flex flex-col items-start gap-1 p-3"
          >
            <div className="flex w-full items-center justify-between">
              <span className="font-medium">{agent.name}</span>
              <Badge variant={agent.status === "active" ? "default" : "secondary"}>
                {agent.model}
              </Badge>
            </div>
            <span className="text-xs text-muted-foreground">{agent.description}</span>
            <div className="flex gap-1">
              {agent.capabilities.map((cap) => (
                <Badge key={cap} variant="outline" className="text-xs">{cap}</Badge>
              ))}
            </div>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
```

---

## 11. Chat API Integration

### 11.1 REST Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/ai/chat` | Submit chat (non-streaming) |
| `GET` | `/api/v1/chat/conversations` | List conversations |
| `POST` | `/api/v1/chat/conversations` | Create conversation |
| `GET` | `/api/v1/chat/conversations/:id/messages` | Get messages |
| `DELETE` | `/api/v1/chat/conversations/:id` | Delete conversation |
| `POST` | `/api/v1/rag/search` | Search knowledge (in-chat) |
| `POST` | `/api/v1/vision/analyze` | Analyze uploaded image |

### 11.2 WebSocket Endpoints

| Endpoint | Purpose |
|---|---|
| `wss://api.aeroxenexus.com/ws/chat` | Main chat streaming |

### 11.3 Chat Hooks

```typescript
// hooks/chat/useChat.ts
export function useChat(conversationId?: string) {
  const { addMessage, setStreaming, appendToStreamingMessage } = useChatStore();
  const queryClient = useQueryClient();

  // Fetch existing messages
  const { data: messages = [], isLoading } = useQuery({
    queryKey: ["chat", "messages", conversationId],
    queryFn: () => chatApi.getMessages(conversationId!),
    enabled: !!conversationId,
  });

  // Send message
  const sendMessage = useCallback(async (content: string, files?: File[]) => {
    // Add user message
    const userMessage: Message = {
      id: crypto.randomUUID(),
      conversationId: conversationId || "",
      type: "user",
      content,
      status: "completed",
      createdAt: new Date().toISOString(),
    };
    addMessage(userMessage);

    // Start streaming response
    setStreaming(true);
    const aiMessageId = crypto.randomUUID();

    // Connect WebSocket and send
    wsClient.send({
      type: "message",
      content,
      conversation_id: conversationId || "",
    });
  }, [conversationId, addMessage, setStreaming]);

  return { messages, isLoading, sendMessage };
}
```

---

## 12. Performance Optimization

### 12.1 Virtual Scrolling

```tsx
// For long conversations (100+ messages)
import { useVirtualizer } from "@tanstack/react-virtual";

function MessageList({ messages }: { messages: Message[] }) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: messages.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 80, // Estimated message height
    overscan: 5,
  });

  return (
    <div ref={parentRef} className="h-full overflow-auto">
      <div style={{ height: `${virtualizer.getTotalSize()}px`, position: "relative" }}>
        {virtualizer.getVirtualItems().map((virtualRow) => (
          <div
            key={virtualRow.key}
            style={{
              position: "absolute",
              top: 0,
              left: 0,
              width: "100%",
              transform: `translateY(${virtualRow.start}px)`,
            }}
          >
            <MessageBubble message={messages[virtualRow.index]} />
          </div>
        ))}
      </div>
    </div>
  );
}
```

### 12.2 Auto-Scroll Behavior

```typescript
// Scroll to bottom when new messages arrive, but not if user scrolled up
function useAutoScroll(messageCount: number) {
  const containerRef = useRef<HTMLDivElement>(null);
  const isUserScrolledUp = useRef(false);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      isUserScrolledUp.current = scrollHeight - scrollTop - clientHeight > 100;
    };

    container.addEventListener("scroll", handleScroll);
    return () => container.removeEventListener("scroll", handleScroll);
  }, []);

  useEffect(() => {
    if (!isUserScrolledUp.current) {
      containerRef.current?.scrollTo({ top: 999999, behavior: "smooth" });
    }
  }, [messageCount]);

  return containerRef;
}
```

### 12.3 Performance Targets

| Metric | Target |
|---|---|
| First token render | < 50ms from WebSocket receive |
| Markdown re-render | Debounced to 100ms intervals |
| Message list scroll | 60fps |
| Conversation switch | < 200ms |
| History load (50 msgs) | < 500ms |
| File upload start | < 100ms |

---

## 13. Error Handling

### 13.1 Chat Error States

| Error | Display | Action |
|---|---|---|
| WebSocket disconnected | Yellow banner "Reconnecting..." | Auto-retry |
| Model timeout | Toast error | Retry button |
| Rate limited | Toast with timer | Wait for reset |
| Invalid message | Inline error on message | Edit and resend |
| File too large | Toast warning | Re-select file |
| Network offline | Banner "You're offline" | Queue message |

### 13.2 Error Recovery

```typescript
// Error handling in WebSocket
wsClient.on("error", (error) => {
  switch (error.code) {
    case "MODEL_TIMEOUT":
      toast.error("AI model timed out", {
        action: { label: "Retry", onClick: retryLastMessage },
      });
      break;
    case "RATE_LIMITED":
      toast.warning("Rate limit exceeded", {
        description: `Try again in ${error.retryAfter}s`,
      });
      break;
    case "AGENT_FAILED":
      toast.error("Agent encountered an error", {
        description: error.message,
        action: { label: "Try Different Agent", onClick: openAgentSelector },
      });
      break;
  }
});
```

---

## 14. Chat Accessibility

### 14.1 Keyboard Shortcuts

| Shortcut | Action |
|---|---|
| `Enter` | Send message |
| `Shift+Enter` | New line in input |
| `Cmd/Ctrl+K` | Open command palette |
| `Cmd/Ctrl+Shift+N` | New conversation |
| `Cmd/Ctrl+/` | Toggle sidebar |
| `Up Arrow` | Edit last message (when input empty) |
| `Tab` | Navigate between messages |
| `Escape` | Close modals/sheets |

### 14.2 ARIA Attributes

| Element | ARIA |
|---|---|
| Message list | `role="log"`, `aria-live="polite"`, `aria-label="Chat messages"` |
| User message | `role="article"`, `aria-label="Your message"` |
| AI message | `role="article"`, `aria-label="Assistant response"` |
| Prompt input | `aria-label="Type your message"` |
| Send button | `aria-label="Send message"` |
| File upload | `aria-label="Attach file"` |
| Voice input | `aria-label="Voice input"` |
| Thinking indicator | `aria-live="polite"`, `aria-label="Assistant is thinking"` |
| Tool execution | `aria-label="Tool: {name}, status: {status}"` |

---

*Document Version: 1.0 — Last Updated: July 2026*
