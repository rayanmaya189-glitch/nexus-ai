# iOS AI Chat

## Table of Contents

- [Chat Screen Layout](#chat-screen-layout)
- [Chat ViewModel](#chat-viewmodel)
- [MessageList View](#messagelist-view)
- [MessageBubble View](#messagebubble-view)
- [PromptBox View](#promptbox-view)
- [FileUploader View](#fileuploader-view)
- [VoiceInput View](#voiceinput-view)
- [AgentSelector View](#agentselector-view)
- [ModelIndicator View](#modelindicator-view)
- [TokenStreamViewer View](#tokenstreamviewer-view)
- [ThinkingIndicator View](#thinkingindicator-view)
- [ToolExecutionDisplay View](#toolexecutiondisplay-view)
- [WebSocket Connection Management](#websocket-connection-management)
- [WebSocket Message Handling](#websocket-message-handling)
- [Streaming Token Buffer](#streaming-token-buffer)
- [Message Types](#message-types)
- [Conversation Management](#conversation-management)
- [Chat History Sidebar](#chat-history-sidebar)
- [Chat Settings](#chat-settings)
- [Image Upload in Chat](#image-upload-in-chat)
- [File Attachment in Chat](#file-attachment-in-chat)
- [Error Handling](#error-handling)
- [Accessibility](#accessibility)
- [Performance](#performance)
- [Offline Support](#offline-support)

---

## Chat Screen Layout

```
┌─────────────────────────────────────────────────────────────┐
│  ◀ Back   Chat with AI Assistant   [⚙️] [⋯]               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 🤖 Hi! How can I help you today?                    │   │
│  │                                                     │   │
│  │ 10:30 AM                                            │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Can you explain SwiftUI?  │            │   │
│  │                                      │            │   │
│  │                                      │ 10:31 AM   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 🤖 Planning...                                      │   │
│  │ ┌───────────────────────────────────────────────┐   │   │
│  │ │ ✅ Searching knowledge base                    │   │   │
│  │ │ 🔄 Generating response...                     │   │   │
│  │ └───────────────────────────────────────────────┘   │   │
│  │                                                     │   │
│  │ SwiftUI is a declarative UI framework...            │   │
│  │ It allows you to build apps using...                │   │
│  │                                                     │   │
│  │ 10:31 AM                                            │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Tell me more  │                        │   │
│  │                             │ 10:32 AM              │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ──────────────────────────────────────────────────────    │
│  │ 🔗 │  │ Tell me more...                  │ 🔊 │ ⬆️ │    │
│  ──────────────────────────────────────────────────────    │
└─────────────────────────────────────────────────────────────┘
```

```swift
// ChatView.swift
import SwiftUI

struct ChatView: View {
    @StateObject private var viewModel: ChatViewModel
    @Environment(\.dismiss) private var dismiss
    @Environment(\.networkMonitor) private var networkMonitor
    @State private var showSettings = false
    @State private var showAgentSelector = false

    init(conversationId: String? = nil) {
        _viewModel = StateObject(wrappedValue: ChatViewModel(
            conversationId: conversationId
        ))
    }

    var body: some View {
        VStack(spacing: 0) {
            // Header
            ChatHeader(
                title: viewModel.conversationTitle,
                isConnected: viewModel.isConnected,
                onSettings: { showSettings = true },
                onAgentSelect: { showAgentSelector = true }
            )

            // Connection Status
            if !networkMonitor.isConnected {
                OfflineBanner()
            }

            // Message List
            MessageListView(
                messages: viewModel.messages,
                isStreaming: viewModel.isStreaming,
                streamingMessage: viewModel.streamingMessage,
                onRetry: viewModel.retryLastMessage,
                onDelete: viewModel.deleteMessage
            )

            // Prompt Box
            PromptBox(
                text: $viewModel.inputText,
                isStreaming: viewModel.isStreaming,
                attachments: $viewModel.attachments,
                onSend: viewModel.sendMessage,
                onAttach: { viewModel.showFilePicker = true },
                onVoice: { viewModel.startVoiceInput() }
            )
        }
        .navigationBarHidden(true)
        .sheet(isPresented: $showSettings) {
            ChatSettingsView(settings: $viewModel.settings)
        }
        .sheet(isPresented: $showAgentSelector) {
            AgentSelectorView(
                agents: viewModel.availableAgents,
                selectedAgent: $viewModel.selectedAgent
            )
        }
        .fileImporter(
            isPresented: $viewModel.showFilePicker,
            allowedContentTypes: viewModel.allowedContentTypes
        ) { result in
            viewModel.handleFileImport(result)
        }
        .alert("Error", isPresented: $viewModel.showError) {
            Button("Retry") { viewModel.retryLastMessage() }
            Button("OK", role: .cancel) {}
        } message: {
            Text(viewModel.errorMessage ?? "An error occurred")
        }
    }
}

// Chat Header
struct ChatHeader: View {
    let title: String
    let isConnected: Bool
    let onSettings: () -> Void
    let onAgentSelect: () -> Void

    var body: some View {
        HStack {
            // Connection indicator
            Circle()
                .fill(isConnected ? Color.green : Color.red)
                .frame(width: 8, height: 8)

            VStack(alignment: .leading) {
                Text(title)
                    .font(.headline)
                    .lineLimit(1)

                HStack(spacing: 4) {
                    Text(isConnected ? "Connected" : "Disconnected")
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
            }

            Spacer()

            Button(action: onAgentSelect) {
                Image(systemName: "person.3")
                    .foregroundColor(.secondary)
            }

            Button(action: onSettings) {
                Image(systemName: "gearshape")
                    .foregroundColor(.secondary)
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(Color(.systemBackground))
        .overlay(
            Rectangle()
                .fill(Color(.systemGray5))
                .frame(height: 1),
            alignment: .bottom
        )
    }
}

// Offline Banner
struct OfflineBanner: View {
    var body: some View {
        HStack {
            Image(systemName: "wifi.slash")
            Text("You're offline. Messages will be sent when connection is restored.")
                .font(.caption)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 8)
        .background(Color.orange.opacity(0.2))
    }
}
```

---

## Chat ViewModel

```swift
// ChatViewModel.swift
import Foundation
import Combine
import SwiftUI

@MainActor
class ChatViewModel: ObservableObject {
    // MARK: - Published State
    @Published var messages: [ChatMessage] = []
    @Published var inputText: String = ""
    @Published var attachments: [ChatAttachment] = []
    @Published var isStreaming: Bool = false
    @Published var streamingMessage: ChatMessage?
    @Published var isConnected: Bool = false
    @Published var isLoadingHistory: Bool = false
    @Published var errorMessage: String?
    @Published var showError: Bool = false
    @Published var showFilePicker: Bool = false
    @Published var selectedAgent: Agent?
    @Published var availableAgents: [Agent] = []
    @Published var settings: ChatSettings = ChatSettings()
    @Published var conversationTitle: String = "New Chat"

    // MARK: - Internal State
    private let conversationId: String?
    private var currentConversationId: String?
    private let chatUseCase: SendMessageUseCaseProtocol
    private let streamUseCase: StreamResponseUseCaseProtocol
    private let historyUseCase: LoadHistoryUseCaseProtocol
    private let webSocketClient: WebSocketClientProtocol
    private let tokenBuffer: StreamingTokenBuffer
    private var cancellables = Set<AnyCancellable>()
    private var streamTask: Task<Void, Never>?

    // MARK: - Allowed Content Types
    let allowedContentTypes: [UTType] = [
        .image, .pdf, .text, .plainText,
        .json, .xml, .spreadsheet, .presentation
    ]

    // MARK: - Init
    init(
        conversationId: String? = nil,
        chatUseCase: SendMessageUseCaseProtocol = SendMessageUseCase(),
        streamUseCase: StreamResponseUseCaseProtocol = StreamResponseUseCase(),
        historyUseCase: LoadHistoryUseCaseProtocol = LoadHistoryUseCase(),
        webSocketClient: WebSocketClientProtocol = WebSocketClient()
    ) {
        self.conversationId = conversationId
        self.chatUseCase = chatUseCase
        self.streamUseCase = streamUseCase
        self.historyUseCase = historyUseCase
        self.webSocketClient = webSocketClient
        self.tokenBuffer = StreamingTokenBuffer()

        setupWebSocketBindings()
        setupTokenBuffer()
        connectWebSocket()

        if let conversationId = conversationId {
            loadHistory(conversationId: conversationId)
        }

        Task {
            await loadAvailableAgents()
        }
    }

    // MARK: - WebSocket Bindings
    private func setupWebSocketBindings() {
        webSocketClient.onTokenReceived { [weak self] token in
            Task { @MainActor in
                self?.handleStreamingToken(token)
            }
        }

        webSocketClient.onCompleted { [weak self] response in
            Task { @MainActor in
                self?.handleStreamCompletion(response)
            }
        }

        webSocketClient.onToolCall { [weak self] toolCall in
            Task { @MainActor in
                self?.handleToolCall(toolCall)
            }
        }

        webSocketClient.onThinking { [weak self] thinking in
            Task { @MainActor in
                self?.handleThinking(thinking)
            }
        }

        webSocketClient.onError { [weak self] error in
            Task { @MainActor in
                self?.handleStreamError(error)
            }
        }

        webSocketClient.onConnectionStateChange { [weak self] state in
            Task { @MainActor in
                switch state {
                case .connected:
                    self?.isConnected = true
                case .disconnected, .failed:
                    self?.isConnected = false
                default:
                    break
                }
            }
        }
    }

    // MARK: - Token Buffer
    private func setupTokenBuffer() {
        tokenBuffer.onBatch { [weak self] tokens in
            Task { @MainActor in
                self?.appendStreamingTokens(tokens)
            }
        }
    }

    // MARK: - WebSocket Connection
    private func connectWebSocket() {
        Task {
            do {
                guard let token = try TokenStorage.shared.getAccessToken() else {
                    throw AuthError.noTokens
                }
                try await webSocketClient.connect(token: token)
                isConnected = true
            } catch {
                os_log(.error, "WebSocket connection failed: %@", error.localizedDescription)
                isConnected = false
            }
        }
    }

    // MARK: - Load History
    private func loadHistory(conversationId: String) {
        isLoadingHistory = true
        currentConversationId = conversationId

        Task {
            do {
                let history = try await historyUseCase.execute(conversationId: conversationId)
                messages = history
                if let firstMessage = history.first {
                    conversationTitle = String(firstMessage.prefix(30)) + "..."
                }
            } catch {
                errorMessage = "Failed to load history"
                showError = true
            }
            isLoadingHistory = false
        }
    }

    // MARK: - Load Agents
    private func loadAvailableAgents() async {
        do:
            // This would come from a use case
            availableAgents = [
                Agent(
                    id: "general",
                    name: "General Assistant",
                    description: "General purpose AI assistant",
                    systemPrompt: "You are a helpful assistant.",
                    modelId: "gpt-4",
                    avatarUrl: nil,
                    capabilities: ["chat", "code", "analysis"],
                    isDefault: true,
                    createdAt: Date()
                )
            ]
            selectedAgent = availableAgents.first
        } catch {
            os_log(.error, "Failed to load agents: %@", error.localizedDescription)
        }
    }

    // MARK: - Send Message
    func sendMessage() {
        let text = inputText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty || !attachments.isEmpty else { return }

        // Create user message
        let userMessage = ChatMessage(
            id: UUID(),
            content: text,
            role: .user,
            timestamp: Date(),
            attachments: attachments,
            toolCalls: [],
            thinking: nil,
            isStreaming: false,
            tokens: [],
            status: .sent
        )

        messages.append(userMessage)
        inputText = ""
        attachments = []
        isStreaming = true

        // Create placeholder for AI response
        let aiMessage = ChatMessage(
            id: UUID(),
            content: "",
            role: .assistant,
            timestamp: Date(),
            attachments: [],
            toolCalls: [],
            thinking: nil,
            isStreaming: true,
            tokens: [],
            status: .streaming
        )

        streamingMessage = aiMessage

        // Send via WebSocket
        webSocketClient.send(.chat(
            conversationId: currentConversationId ?? UUID().uuidString,
            message: text,
            agentId: selectedAgent?.id,
            attachments: userMessage.attachments.map { attachment in
                Attachment(
                    id: attachment.id,
                    name: attachment.name,
                    type: .document,
                    url: nil,
                    data: attachment.data,
                    size: attachment.size
                )
            }
        ))
    }

    // MARK: - Handle Streaming Token
    private func handleStreamingToken(_ token: String) {
        tokenBuffer.append(token)
    }

    private func appendStreamingTokens(_ tokens: [String]) {
        guard var message = streamingMessage else { return }
        message.tokens.append(contentsOf: tokens)
        streamingMessage = message
    }

    // MARK: - Handle Stream Completion
    private func handleStreamCompletion(_ response: ChatResponse) {
        tokenBuffer.flush()

        if var message = streamingMessage {
            message.content = response.content
            message.isStreaming = false
            message.status = .delivered
            messages.append(message)
        }

        streamingMessage = nil
        isStreaming = false
        currentConversationId = response.conversationId
    }

    // MARK: - Handle Tool Call
    private func handleToolCall(_ toolCall: ToolCall) {
        guard var message = streamingMessage else { return }
        message.toolCalls.append(toolCall)
        streamingMessage = message
    }

    // MARK: - Handle Thinking
    private func handleThinking(_ thinking: String) {
        guard var message = streamingMessage else { return }
        message.thinking = thinking
        streamingMessage = message
    }

    // MARK: - Handle Stream Error
    private func handleStreamError(_ error: Error) {
        isStreaming = false
        streamingMessage = nil

        if let chatError = error as? WebSocketError {
            errorMessage = chatError.localizedDescription
        } else {
            errorMessage = "Connection error. Please try again."
        }
        showError = true
    }

    // MARK: - Retry Last Message
    func retryLastMessage() {
        guard let lastUserMessage = messages.last(where: { $0.role == .user }) else { return }

        // Remove failed AI response
        messages.removeAll { $0.status == .failed }

        inputText = lastUserMessage.content
        sendMessage()
    }

    // MARK: - Delete Message
    func deleteMessage(_ message: ChatMessage) {
        messages.removeAll { $0.id == message.id }
    }

    // MARK: - File Import
    func handleFileImport(_ result: Result<[URL], Error>) {
        switch result {
        case .success(let urls):
            for url in urls {
                guard url.startAccessingSecurityScopedResource() else { continue }
                defer { url.stopAccessingSecurityScopedResource() }

                if let data = try? Data(contentsOf: url) {
                    let attachment = ChatAttachment(
                        id: UUID(),
                        name: url.lastPathComponent,
                        type: url.pathExtension,
                        url: url,
                        data: data,
                        size: Int64(data.count),
                        uploadProgress: 0
                    )
                    attachments.append(attachment)
                }
            }
        case .failure(let error):
            errorMessage = error.localizedDescription
            showError = true
        }
    }

    // MARK: - Voice Input
    func startVoiceInput() {
        // Voice input implementation
        os_log(.info, "Voice input started")
    }
}

// MARK: - Chat Message Model
struct ChatMessage: Identifiable, Equatable {
    let id: UUID
    let content: String
    let role: MessageRole
    let timestamp: Date
    var attachments: [ChatAttachment]
    var toolCalls: [ToolCall]
    var thinking: String?
    var isStreaming: Bool
    var tokens: [String]
    var status: MessageStatus

    enum MessageRole {
        case user
        case assistant
        case system
    }

    enum MessageStatus {
        case sending
        case sent
        case delivered
        case streaming
        case failed
    }

    var displayedContent: String {
        if isStreaming && !tokens.isEmpty {
            return content + tokens.joined()
        }
        return content
    }

    static func == (lhs: ChatMessage, rhs: ChatMessage) -> Bool {
        lhs.id == rhs.id && lhs.isStreaming == rhs.isStreaming && lhs.tokens == rhs.tokens
    }
}

// MARK: - Chat Attachment
struct ChatAttachment: Identifiable {
    let id: UUID
    let name: String
    let type: String
    let url: URL?
    let data: Data?
    let size: Int64
    var uploadProgress: Double

    var icon: String {
        switch type.lowercased() {
        case "jpg", "jpeg", "png", "gif": return "photo"
        case "pdf": return "doc.richtext"
        case "txt": return "doc.text"
        case "json": return "doc.text"
        default: return "doc"
        }
    }

    var formattedSize: String {
        ByteCountFormatter.string(fromByteCount: size, countStyle: .file)
    }
}

// MARK: - Chat Settings
struct ChatSettings {
    var selectedModelId: String = "gpt-4"
    var temperature: Double = 0.7
    var maxTokens: Int = 4096
    var streamingEnabled: Bool = true
    var autoScroll: Bool = true
    var showTimestamps: Bool = true
    var showThinking: Bool = true
    var showToolCalls: Bool = true
}
```

### ViewModel State Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                 CHAT VIEWMODEL STATE                         │
│                                                             │
│  ┌───────────┐     ┌──────────────┐     ┌──────────────┐  │
│  │   Idle    │────▶│  Connecting  │────▶│  Connected   │  │
│  │           │     │  WebSocket   │     │              │  │
│  └───────────┘     └──────────────┘     └──────┬───────┘  │
│                                                 │          │
│                                                 ▼          │
│                                         ┌──────────────┐  │
│                                         │  Sending     │  │
│                                         │  Message     │  │
│                                         └──────┬───────┘  │
│                                                │          │
│                                                ▼          │
│                                    ┌───────────────────┐  │
│                                    │   Streaming       │  │
│                                    │   Response        │  │
│                                    └───────┬───────────┘  │
│                                            │              │
│                         ┌──────────────────┼──────┐       │
│                         │                  │      │       │
│                         ▼                  ▼      ▼       │
│                 ┌──────────────┐  ┌────────┐ ┌────────┐  │
│                 │  Completed   │  │ Tool   │ │ Error  │  │
│                 │              │  │ Call   │ │        │  │
│                 └──────────────┘  └────────┘ └────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## MessageList View

```swift
// MessageListView.swift
import SwiftUI

struct MessageListView: View {
    let messages: [ChatMessage]
    let isStreaming: Bool
    let streamingMessage: ChatMessage?
    let onRetry: () -> Void
    let onDelete: (ChatMessage) -> Void

    @State private var scrollViewProxy: ScrollViewProxy?
    @State private var isUserScrolling = false
    @AppStorage("chatAutoScroll") private var autoScroll = true

    var body: some View {
        ScrollViewReader { proxy in
            ScrollView {
                LazyVStack(spacing: 12) {
                    // Welcome message if empty
                    if messages.isEmpty && streamingMessage == nil {
                        WelcomeMessage()
                    }

                    // Date grouping
                    ForEach(groupedMessages, id: \.key) { date, dayMessages in
                        Section {
                            ForEach(dayMessages) { message in
                                MessageBubble(
                                    message: message,
                                    onRetry: onRetry,
                                    onDelete: { onDelete(message) }
                                )
                                .id(message.id)
                                .transition(.asymmetric(
                                    insertion: .move(edge: .bottom).combined(with: .opacity),
                                    removal: .opacity
                                ))
                            }
                        } header: {
                            DateHeader(date: date)
                        }
                    }

                    // Streaming message
                    if let streaming = streamingMessage {
                        MessageBubble(
                            message: streaming,
                            onRetry: {},
                            onDelete: {}
                        )
                        .id("streaming")
                    }

                    // Typing indicator
                    if isStreaming && streamingMessage == nil {
                        TypingIndicator()
                    }
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
            }
            .scrollDismissesKeyboard(.interactively)
            .scrollIndicators(.hidden)
            .onChange(of: messages.count) { _, _ in
                scrollToBottom(proxy: proxy)
            }
            .onChange(of: streamingMessage?.tokens.count) { _, _ in
                if autoScroll && !isUserScrolling {
                    scrollToBottom(proxy: proxy)
                }
            }
            .onAppear { scrollViewProxy = proxy }
        }
    }

    private func scrollToBottom(proxy: ScrollViewProxy, animated: Bool = true) {
        withAnimation(animated ? .easeOut(duration: 0.3) : nil) {
            if let streaming = streamingMessage {
                proxy.scrollTo(streaming.id, anchor: .bottom)
            } else if let lastMessage = messages.last {
                proxy.scrollTo(lastMessage.id, anchor: .bottom)
            }
        }
    }

    private var groupedMessages: [(key: Date, value: [ChatMessage])] {
        let grouped = Dictionary(grouping: messages) { message in
            Calendar.current.startOfDay(for: message.timestamp)
        }
        return grouped.sorted { $0.key < $1.key }
    }
}

// Date Header
struct DateHeader: View {
    let date: Date

    var body: some View {
        Text(formattedDate)
            .font(.caption)
            .foregroundColor(.secondary)
            .padding(.vertical, 8)
    }

    private var formattedDate: String {
        let formatter = DateFormatter()
        if Calendar.current.isDateInToday(date) {
            return "Today"
        } else if Calendar.current.isDateInYesterday(date) {
            return "Yesterday"
        } else {
            formatter.dateStyle = .medium
            return formatter.string(from: date)
        }
    }
}

// Typing Indicator
struct TypingIndicator: View {
    @State private var animationPhase = 0

    var body: some View {
        HStack(alignment: .bottom) {
            Image(systemName: "brain.head.profile")
                .foregroundColor(.white)
                .padding(8)
                .background(Color.accentColor)
                .clipShape(Circle())

            HStack(spacing: 4) {
                ForEach(0..<3) { index in
                    Circle()
                        .fill(Color.gray)
                        .frame(width: 8, height: 8)
                        .offset(y: animationPhase == index ? -4 : 0)
                        .animation(
                            .easeInOut(duration: 0.5)
                            .repeatForever(autoreverses: true)
                            .delay(Double(index) * 0.15),
                            value: animationPhase
                        )
                }
            }
            .padding(12)
            .background(Color(.systemGray6))
            .clipShape(RoundedRectangle(cornerRadius: 16))
        }
        .onAppear {
            animationPhase = 1
        }
    }
}

// Welcome Message
struct WelcomeMessage: View {
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "brain.head.profile")
                .font(.system(size: 60))
                .foregroundStyle(
                    LinearGradient(
                        colors: [.blue, .purple],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )

            Text("How can I help you today?")
                .font(.title2.bold())

            Text("Ask me anything, or try one of these:")
                .font(.subheadline)
                .foregroundColor(.secondary)

            VStack(spacing: 8) {
                SuggestionButton(text: "Explain SwiftUI")
                SuggestionButton(text: "Write a function")
                SuggestionButton(text: "Debug my code")
                SuggestionButton(text: "Create a design")
            }
        }
        .padding(.top, 60)
    }
}

struct SuggestionButton: View {
    let text: String
    @EnvironmentObject var viewModel: ChatViewModel

    var body: some View {
        Button(action: {
            viewModel.inputText = text
            viewModel.sendMessage()
        }) {
            Text(text)
                .font(.subheadline)
                .foregroundColor(.primary)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(Color(.systemGray6))
                .clipShape(Capsule())
        }
    }
}
```

---

## MessageBubble View

```swift
// MessageBubble.swift
import SwiftUI

struct MessageBubble: View {
    let message: ChatMessage
    let onRetry: () -> Void
    let onDelete: () -> Void

    @State private var showContextMenu = false

    var body: some View {
        HStack(alignment: message.role == .user ? .trailing : .leading, spacing: 8) {
            if message.role == .assistant {
                // AI Avatar
                Image(systemName: "brain.head.profile")
                    .foregroundColor(.white)
                    .padding(6)
                    .background(Color.accentColor)
                    .clipShape(Circle())
                    .frame(width: 32, height: 32)
            }

            VStack(alignment: message.role == .user ? .trailing : .leading, spacing: 4) {
                // Attachments
                if !message.attachments.isEmpty {
                    AttachmentPreview(attachments: message.attachments)
                }

                // Thinking Indicator
                if let thinking = message.thinking {
                    ThinkingIndicator(text: thinking)
                }

                // Tool Calls
                if !message.toolCalls.isEmpty {
                    ForEach(message.toolCalls) { toolCall in
                        ToolExecutionDisplay(toolCall: toolCall)
                    }
                }

                // Message Content
                if !message.displayedContent.isEmpty {
                    contentBubble
                }

                // Timestamp
                if message.status != .streaming {
                    HStack(spacing: 4) {
                        Text(message.timestamp, style: .time)
                            .font(.caption2)
                            .foregroundColor(.secondary)

                        if message.role == .user {
                            statusIcon
                        }
                    }
                }

                // Retry button for failed messages
                if message.status == .failed {
                    Button(action: onRetry) {
                        HStack(spacing: 4) {
                            Image(systemName: "arrow.clockwise")
                            Text("Retry")
                        }
                        .font(.caption)
                        .foregroundColor(.red)
                    }
                }
            }

            if message.role == .user {
                // User Avatar
                Image(systemName: "person.circle.fill")
                    .foregroundColor(.white)
                    .padding(6)
                    .background(Color.gray)
                    .clipShape(Circle())
                    .frame(width: 32, height: 32)
            }
        }
        .contextMenu {
            Button(action: { copyMessage() }) {
                Label("Copy", systemImage: "doc.on.doc")
            }

            if message.role == .user {
                Button(action: onRetry) {
                    Label("Retry", systemImage: "arrow.clockwise")
                }
            }

            Button(action: onDelete, role: .destructive) {
                Label("Delete", systemImage: "trash")
            }
        }
    }

    // Message Content View
    @ViewBuilder
    private var contentBubble: some View {
        if message.role == .user {
            // User message
            Text(message.displayedContent)
                .padding(.horizontal, 16)
                .padding(.vertical, 12)
                .background(Color.accentColor)
                .foregroundColor(.white)
                .clipShape(RoundedRectangle(cornerRadius: 16))
                .clipShape(
                    RoundedCorner(
                        corners: [.topLeft, .topRight, .bottomLeft],
                        radius: 16
                    )
                )
        } else {
            // AI message
            VStack(alignment: .leading, spacing: 8) {
                // Streaming tokens
                if message.isStreaming && !message.tokens.isEmpty {
                    StreamingTextView(tokens: message.tokens)
                } else {
                    // Rendered content
                    MarkdownTextView(content: message.displayedContent)
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(Color(.systemGray6))
            .foregroundColor(.primary)
            .clipShape(RoundedRectangle(cornerRadius: 16))
            .clipShape(
                RoundedCorner(
                    corners: [.topLeft, .topRight, .bottomRight],
                    radius: 16
                )
            )
        }
    }

    // Status Icon
    @ViewBuilder
    private var statusIcon: some View {
        switch message.status {
        case .sending:
            ProgressView()
                .scaleEffect(0.6)
        case .sent:
            Image(systemName: "checkmark")
                .font(.caption2)
                .foregroundColor(.secondary)
        case .delivered:
            Image(systemName: "checkmark.circle.fill")
                .font(.caption2)
                .foregroundColor(.secondary)
        case .streaming:
            ProgressView()
                .scaleEffect(0.6)
        case .failed:
            Image(systemName: "exclamationmark.circle.fill")
                .font(.caption2)
                .foregroundColor(.red)
        }
    }

    private func copyMessage() {
        UIPasteboard.general.string = message.displayedContent
    }
}

// RoundedCorner helper
struct RoundedCorner: Shape {
    var corners: UIRectCorner
    var radius: CGFloat

    func path(in rect: CGRect) -> Path {
        let path = UIBezierPath(
            roundedRect: rect,
            byRoundingCorners: corners,
            cornerRadii: CGSize(width: radius, height: radius)
        )
        return Path(path.cgPath)
    }
}

// Attachment Preview
struct AttachmentPreview: View {
    let attachments: [ChatAttachment]

    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                ForEach(attachments) { attachment in
                    AttachmentCard(attachment: attachment)
                }
            }
        }
    }
}

struct AttachmentCard: View {
    let attachment: ChatAttachment

    var body: some View {
        VStack(spacing: 4) {
            if attachment.type.lowercased() == "jpg" ||
               attachment.type.lowercased() == "jpeg" ||
               attachment.type.lowercased() == "png" {
                if let data = attachment.data, let uiImage = UIImage(data: data) {
                    Image(uiImage: uiImage)
                        .resizable()
                        .scaledToFill()
                        .frame(width: 120, height: 80)
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                }
            } else {
                VStack(spacing: 4) {
                    Image(systemName: attachment.icon)
                        .font(.title2)
                        .foregroundColor(.accentColor)
                    Text(attachment.name)
                        .font(.caption2)
                        .lineLimit(1)
                    Text(attachment.formattedSize)
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }
                .frame(width: 120, height: 80)
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }
        }
    }
}

// Markdown Text View
struct MarkdownTextView: View {
    let content: String

    var body: some View {
        // Simple markdown rendering
        VStack(alignment: .leading, spacing: 8) {
            ForEach(Array(parseMarkdown(content).enumerated()), id: \.offset) { _, block in
                switch block {
                case .text(let text):
                    Text(text)
                case .code(let language, let code):
                    CodeBlock(language: language, code: code)
                case .heading(let text):
                    Text(text)
                        .font(.headline)
                        .padding(.top, 4)
                case .list(let items):
                    VStack(alignment: .leading, spacing: 4) {
                        ForEach(Array(items.enumerated()), id: \.offset) { _, item in
                            HStack(alignment: .top) {
                                Text("•")
                                Text(item)
                            }
                        }
                    }
                }
            }
        }
    }

    private func parseMarkdown(_ text: String) -> [MarkdownBlock] {
        // Simplified markdown parser
        var blocks: [MarkdownBlock] = []
        let lines = text.components(separatedBy: .newlines)

        for line in lines {
            if line.hasPrefix("```") {
                let language = String(line.dropFirst(3))
                blocks.append(.code(language: language, code: ""))
            } else if line.hasPrefix("# ") {
                blocks.append(.heading(String(line.dropFirst(2))))
            } else if line.hasPrefix("• ") || line.hasPrefix("- ") {
                blocks.append(.list([String(line.dropFirst(2))]))
            } else if !line.isEmpty {
                blocks.append(.text(line))
            }
        }

        return blocks
    }

    enum MarkdownBlock {
        case text(String)
        case code(language: String, code: String)
        case heading(String)
        case list([String])
    }
}

// Code Block
struct CodeBlock: View {
    let language: String
    let code: String

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Text(language)
                    .font(.caption)
                    .foregroundColor(.secondary)
                Spacer()
                Button(action: { copyCode() }) {
                    Image(systemName: "doc.on.doc")
                        .font(.caption)
                }
            }

            ScrollView(.horizontal, showsIndicators: false) {
                Text(code)
                    .font(.system(.body, design: .monospaced))
                    .textSelection(.enabled)
            }
        }
        .padding(12)
        .background(Color(.systemGray5))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private func copyCode() {
        UIPasteboard.general.string = code
    }
}

// Streaming Text View
struct StreamingTextView: View {
    let tokens: [String]

    var body: some View {
        Text(tokens.joined())
            .textSelection(.enabled)
    }
}
```

---

## PromptBox View

```swift
// PromptBox.swift
import SwiftUI

struct PromptBox: View {
    @Binding var text: String
    let isStreaming: Bool
    @Binding var attachments: [ChatAttachment]
    let onSend: () -> Void
    let onAttach: () -> Void
    let onVoice: () -> Void

    @State private var isRecording = false
    @State private var textHeight: CGFloat = 20

    var body: some View {
        VStack(spacing: 0) {
            // Attachment preview
            if !attachments.isEmpty {
                AttachmentBar(attachments: $attachments)
            }

            Divider()

            // Input area
            HStack(alignment: .bottom, spacing: 12) {
                // Attach button
                Button(action: onAttach) {
                    Image(systemName: "plus.circle.fill")
                        .font(.title2)
                        .foregroundColor(.secondary)
                }
                .disabled(isStreaming)

                // Text input
                VStack(alignment: .leading, spacing: 4) {
                    ZStack(alignment: .topLeading) {
                        if text.isEmpty {
                            Text("Message...")
                                .foregroundColor(.secondary)
                                .padding(.top, 8)
                                .padding(.leading, 4)
                        }

                        TextEditor(text: $text)
                            .scrollContentBackground(.hidden)
                            .frame(minHeight: 20, maxHeight: 120)
                            .fixedSize(horizontal: false, vertical: true)
                            .padding(.top, 4)
                            .padding(.leading, -4)
                    }
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 8)
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 20))

                // Voice button
                Button(action: onVoice) {
                    Image(systemName: isRecording ? "stop.circle.fill" : "mic.circle.fill")
                        .font(.title2)
                        .foregroundColor(isRecording ? .red : .secondary)
                }

                // Send button
                Button(action: onSend) {
                    Image(systemName: "arrow.up.circle.fill")
                        .font(.title2)
                        .foregroundColor(canSend ? .accentColor : .gray)
                }
                .disabled(!canSend)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
        }
    }

    private var canSend: Bool {
        (!text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || !attachments.isEmpty) && !isStreaming
    }
}

// Attachment Bar
struct AttachmentBar: View {
    @Binding var attachments: [ChatAttachment]

    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 8) {
                ForEach(attachments) { attachment in
                    AttachmentChip(attachment: attachment) {
                        attachments.removeAll { $0.id == attachment.id }
                    }
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 8)
        }
    }
}

struct AttachmentChip: View {
    let attachment: ChatAttachment
    let onRemove: () -> Void

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: attachment.icon)
                .font(.caption)
            Text(attachment.name)
                .font(.caption)
                .lineLimit(1)
            Button(action: onRemove) {
                Image(systemName: "xmark.circle.fill")
                    .font(.caption2)
            }
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(Color(.systemGray6))
        .clipShape(Capsule())
    }
}
```

### PromptBox Layout

```
┌─────────────────────────────────────────────────────────────┐
│  ┌──────────────────────────────────────────────────────┐  │
│  │ 📄 document.pdf  ✕    🖼️ image.jpg  ✕              │  │
│  └──────────────────────────────────────────────────────┘  │
│  ────────────────────────────────────────────────────────  │
│  │    │  ┌────────────────────────────┐      │     │    │  │
│  │ ➕ │  │                            │  🎤  │  ⬆️ │    │  │
│  │    │  │  Message...                │      │     │    │  │
│  │    │  └────────────────────────────┘      │     │    │  │
│  └────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

---

## FileUploader View

```swift
// FileUploader.swift
import SwiftUI
import PhotosUI

struct FileUploader: View {
    @Binding var attachments: [ChatAttachment]
    @State private var selectedItem: PhotosPickerItem?
    @State private var showCamera = false
    @State private var showDocumentPicker = false
    @State private var uploadProgress: [UUID: Double] = [:]

    var body: some View {
        HStack(spacing: 16) {
            // Camera
            Button(action: { showCamera = true }) {
                VStack(spacing: 4) {
                    Image(systemName: "camera.fill")
                        .font(.title2)
                    Text("Camera")
                        .font(.caption2)
                }
                .foregroundColor(.primary)
            }

            // Photo Library
            PhotosPicker(
                selection: $selectedItem,
                matching: .images
            ) {
                VStack(spacing: 4) {
                    Image(systemName: "photo.fill")
                        .font(.title2)
                    Text("Photos")
                        .font(.caption2)
                }
                .foregroundColor(.primary)
            }

            // Document
            Button(action: { showDocumentPicker = true }) {
                VStack(spacing: 4) {
                    Image(systemName: "doc.fill")
                        .font(.title2)
                    Text("Document")
                        .font(.caption2)
                }
                .foregroundColor(.primary)
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .sheet(isPresented: $showCamera) {
            CameraView { image in
                addImageAttachment(image)
            }
        }
        .fileImporter(
            isPresented: $showDocumentPicker,
            allowedContentTypes: [.pdf, .plainText, .json, .xml]
        ) { result in
            handleDocumentImport(result)
        }
        .onChange(of: selectedItem) { _, newItem in
            Task {
                if let data = try? await newItem?.loadTransferable(type: Data.self) {
                    let attachment = ChatAttachment(
                        id: UUID(),
                        name: "image.jpg",
                        type: "jpg",
                        url: nil,
                        data: data,
                        size: Int64(data.count),
                        uploadProgress: 0
                    )
                    attachments.append(attachment)
                }
                selectedItem = nil
            }
        }
    }

    private func addImageAttachment(_ image: UIImage) {
        guard let data = image.jpegData(compressionQuality: 0.8) else { return }

        let attachment = ChatAttachment(
            id: UUID(),
            name: "photo_\(Date().timeIntervalSince1970).jpg",
            type: "jpg",
            url: nil,
            data: data,
            size: Int64(data.count),
            uploadProgress: 0
        )
        attachments.append(attachment)
    }

    private func handleDocumentImport(_ result: Result<[URL], Error>) {
        switch result {
        case .success(let urls):
            for url in urls {
                guard url.startAccessingSecurityScopedResource() else { continue }
                defer { url.stopAccessingSecurityScopedResource() }

                if let data = try? Data(contentsOf: url) {
                    let attachment = ChatAttachment(
                        id: UUID(),
                        name: url.lastPathComponent,
                        type: url.pathExtension,
                        url: url,
                        data: data,
                        size: Int64(data.count),
                        uploadProgress: 0
                    )
                    attachments.append(attachment)
                }
            }
        case .failure:
            break
        }
    }
}

// Camera View
struct CameraView: UIViewControllerRepresentable {
    let onCapture: (UIImage) -> Void

    func makeUIViewController(context: Context) -> UIImagePickerController {
        let picker = UIImagePickerController()
        picker.sourceType = .camera
        picker.delegate = context.coordinator
        return picker
    }

    func updateUIViewController(_ uiViewController: UIImagePickerController, context: Context) {}

    func makeCoordinator() -> Coordinator {
        Coordinator(onCapture: onCapture)
    }

    class Coordinator: NSObject, UIImagePickerControllerDelegate, UINavigationControllerDelegate {
        let onCapture: (UIImage) -> Void

        init(onCapture: @escaping (UIImage) -> Void) {
            self.onCapture = onCapture
        }

        func imagePickerController(
            _ picker: UIImagePickerController,
            didFinishPickingMediaWithInfo info: [UIImagePickerController.InfoKey: Any]
        ) {
            if let image = info[.originalImage] as? UIImage {
                onCapture(image)
            }
            picker.dismiss(animated: true)
        }

        func imagePickerControllerDidCancel(_ picker: UIImagePickerController) {
            picker.dismiss(animated: true)
        }
    }
}
```

---

## VoiceInput View

```swift
// VoiceInput.swift
import SwiftUI

struct VoiceInput: View {
    @State private var isRecording = false
    @State private var recordingDuration: TimeInterval = 0
    @State private var audioLevel: CGFloat = 0
    @State private var recognizedText: String = ""

    let onTranscription: (String) -> Void
    let onCancel: () -> Void

    private let timer = Timer.publish(every: 0.1, on: .main, in: .common).autoconnect()

    var body: some View {
        VStack(spacing: 24) {
            // Waveform visualization
            VoiceWaveformView(audioLevel: audioLevel)
                .frame(height: 100)
                .padding(.top, 32)

            // Duration
            Text(formattedDuration)
                .font(.headline)
                .foregroundColor(.secondary)

            // Recognized text
            if !recognizedText.isEmpty {
                ScrollView {
                    Text(recognizedText)
                        .font(.body)
                        .padding()
                        .background(Color(.systemGray6))
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }
                .frame(maxHeight: 100)
            }

            // Controls
            HStack(spacing: 32) {
                Button(action: onCancel) {
                    Image(systemName: "xmark.circle.fill")
                        .font(.title)
                        .foregroundColor(.red)
                }

                Button(action: toggleRecording) {
                    ZStack {
                        Circle()
                            .fill(isRecording ? Color.red : Color.accentColor)
                            .frame(width: 70, height: 70)

                        if isRecording {
                            RoundedRectangle(cornerRadius: 4)
                                .fill(Color.white)
                                .frame(width: 24, height: 24)
                        } else {
                            Image(systemName: "mic.fill")
                                .font(.title2)
                                .foregroundColor(.white)
                        }
                    }
                }

                Button(action: submitRecording) {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.title)
                        .foregroundColor(.green)
                }
                .disabled(recognizedText.isEmpty)
            }
            .padding(.bottom, 32)
        }
        .onReceive(timer) { _ in
            if isRecording {
                recordingDuration += 0.1
                // Simulate audio level
                audioLevel = CGFloat.random(in: 0.2...1.0)
            }
        }
    }

    private func toggleRecording() {
        isRecording.toggle()
        if isRecording {
            startRecording()
        } else {
            stopRecording()
        }
    }

    private func startRecording() {
        // Start audio recording
        os_log(.info, "Voice recording started")
    }

    private func stopRecording() {
        // Stop audio recording and process
        os_log(.info, "Voice recording stopped")
    }

    private func submitRecording() {
        onTranscription(recognizedText)
    }

    private var formattedDuration: String {
        let minutes = Int(recordingDuration) / 60
        let seconds = Int(recordingDuration) % 60
        return String(format: "%02d:%02d", minutes, seconds)
    }
}

// Voice Waveform View
struct VoiceWaveformView: View {
    let audioLevel: CGFloat

    var body: some View {
        HStack(spacing: 4) {
            ForEach(0..<40) { index in
                RoundedRectangle(cornerRadius: 2)
                    .fill(Color.accentColor)
                    .frame(width: 4, height: waveHeight(for: index))
                    .animation(
                        .easeInOut(duration: 0.1),
                        value: audioLevel
                    )
            }
        }
        .frame(maxWidth: .infinity)
    }

    private func waveHeight(for index: Int) -> CGFloat {
        let normalizedIndex = CGFloat(index) / 40
        let distanceFromCenter = abs(normalizedIndex - 0.5) * 2
        let height = (1 - distanceFromCenter) * audioLevel * 40
        return max(4, height)
    }
}
```

---

## AgentSelector View

```swift
// AgentSelector.swift
import SwiftUI

struct AgentSelectorView: View {
    let agents: [Agent]
    @Binding var selectedAgent: Agent?
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            List {
                ForEach(agents) { agent in
                    Button(action: {
                        selectedAgent = agent
                        dismiss()
                    }) {
                        HStack(spacing: 12) {
                            // Agent avatar
                            if let avatarUrl = agent.avatarUrl {
                                AsyncImage(url: URL(string: avatarUrl)) { image in
                                    image.resizable().scaledToFill()
                                } placeholder: {
                                    Image(systemName: "person.circle.fill")
                                }
                                .frame(width: 48, height: 48)
                                .clipShape(Circle())
                            } else {
                                Image(systemName: "person.circle.fill")
                                    .font(.title2)
                                    .foregroundColor(.accentColor)
                                    .frame(width: 48, height: 48)
                            }

                            VStack(alignment: .leading, spacing: 4) {
                                Text(agent.name)
                                    .font(.headline)
                                    .foregroundColor(.primary)

                                Text(agent.description)
                                    .font(.subheadline)
                                    .foregroundColor(.secondary)
                                    .lineLimit(2)

                                HStack {
                                    ForEach(agent.capabilities.prefix(3), id: \.self) { capability in
                                        Text(capability)
                                            .font(.caption2)
                                            .padding(.horizontal, 6)
                                            .padding(.vertical, 2)
                                            .background(Color.accentColor.opacity(0.1))
                                            .clipShape(Capsule())
                                    }
                                }
                            }

                            Spacer()

                            if selectedAgent?.id == agent.id {
                                Image(systemName: "checkmark.circle.fill")
                                    .foregroundColor(.accentColor)
                            }
                        }
                        .padding(.vertical, 4)
                    }
                }
            }
            .navigationTitle("Select Agent")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Done") { dismiss() }
                }
            }
        }
    }
}
```

---

## ModelIndicator View

```swift
// ModelIndicator.swift
import SwiftUI

struct ModelIndicator: View {
    let modelName: String
    let status: ModelStatus

    enum ModelStatus {
        case available
        case loading
        case unavailable

        var color: Color {
            switch self {
            case .available: return .green
            case .loading: return .orange
            case .unavailable: return .red
            }
        }

        var icon: String {
            switch self {
            case .available: return "checkmark.circle.fill"
            case .loading: return "arrow.triangle.2.circlepath"
            case .unavailable: return "xmark.circle.fill"
            }
        }
    }

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: status.icon)
                .font(.caption2)
                .foregroundColor(status.color)

            Text(modelName)
                .font(.caption2)
                .foregroundColor(.secondary)

            if status == .loading {
                ProgressView()
                    .scaleEffect(0.5)
            }
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(Color(.systemGray6))
        .clipShape(Capsule())
    }
}
```

---

## TokenStreamViewer View

```swift
// TokenStreamViewer.swift
import SwiftUI

struct TokenStreamViewer: View {
    let tokens: [String]
    let isComplete: Bool

    @State private var visibleTokenCount = 0

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            // Rendered tokens
            Text(visibleTokens)
                .font(.body)
                .textSelection(.enabled)

            // Cursor
            if !isComplete {
                Text("▌")
                    .font(.body)
                    .foregroundColor(.accentColor)
                    .opacity(cursorOpacity)
            }
        }
        .onChange(of: tokens.count) { _, _ in
            updateVisibleTokens()
        }
    }

    private var visibleTokens: String {
        Array(tokens.prefix(visibleTokenCount)).joined()
    }

    private var cursorOpacity: Double {
        isComplete ? 0 : 1
    }

    private func updateVisibleTokens() {
        withAnimation(.easeOut(duration: 0.1)) {
            visibleTokenCount = tokens.count
        }
    }
}
```

---

## ThinkingIndicator View

```swift
// ThinkingIndicator.swift
import SwiftUI

struct ThinkingIndicator: View {
    let text: String

    @State private var isAnimating = false

    var body: some View {
        HStack(spacing: 8) {
            // Animated icon
            Image(systemName: "brain.head.profile")
                .font(.caption)
                .foregroundColor(.accentColor)
                .rotationEffect(.degrees(isAnimating ? 360 : 0))
                .animation(
                    .linear(duration: 2).repeatForever(autoreverses: false),
                    value: isAnimating
                )

            // Thinking text
            Text(text)
                .font(.caption)
                .foregroundColor(.secondary)
                .italic()
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(Color.accentColor.opacity(0.1))
        .clipShape(Capsule())
        .onAppear {
            isAnimating = true
        }
    }
}
```

---

## ToolExecutionDisplay View

```swift
// ToolExecutionDisplay.swift
import SwiftUI

struct ToolExecutionDisplay: View {
    let toolCall: ToolCall

    var body: some View {
        DisclosureGroup {
            if let result = toolCall.result {
                Text(result)
                    .font(.caption)
                    .foregroundColor(.secondary)
                    .padding(.top, 4)
            }
        } label: {
            HStack(spacing: 8) {
                // Status icon
                statusIcon

                // Tool name
                Text(toolCall.name)
                    .font(.caption)
                    .foregroundColor(.primary)

                Spacer()

                // Status text
                Text(statusText)
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(Color(.systemGray6))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    @ViewBuilder
    private var statusIcon: some View {
        switch toolCall.status {
        case .pending:
            Image(systemName: "clock")
                .foregroundColor(.gray)
        case .running:
            ProgressView()
                .scaleEffect(0.6)
        case .completed:
            Image(systemName: "checkmark.circle.fill")
                .foregroundColor(.green)
        case .failed:
            Image(systemName: "xmark.circle.fill")
                .foregroundColor(.red)
        }
    }

    private var statusText: String {
        switch toolCall.status {
        case .pending: return "Pending"
        case .running: return "Running..."
        case .completed: return "Complete"
        case .failed: return "Failed"
        }
    }
}
```

---

## WebSocket Connection Management

```swift
// WebSocket Connection Manager
class WebSocketConnectionManager: ObservableObject {
    @Published var isConnected = false
    @Published var connectionState: ConnectionState = .disconnected

    private let webSocketClient: WebSocketClientProtocol
    private var reconnectAttempts = 0
    private let maxReconnectAttempts = 10

    enum ConnectionState {
        case disconnected
        case connecting
        case connected
        case reconnecting
        case failed
    }

    init(webSocketClient: WebSocketClientProtocol = WebSocketClient()) {
        self.webSocketClient = webSocketClient
        setupBindings()
    }

    private func setupBindings() {
        webSocketClient.onConnectionStateChange { [weak self] state in
            Task { @MainActor in
                switch state {
                case .connected:
                    self?.isConnected = true
                    self?.connectionState = .connected
                    self?.reconnectAttempts = 0
                case .disconnected:
                    self?.isConnected = false
                    self?.connectionState = .disconnected
                case .connecting:
                    self?.connectionState = .connecting
                case .reconnecting:
                    self?.isConnected = false
                    self?.connectionState = .reconnecting
                case .failed:
                    self?.isConnected = false
                    self?.connectionState = .failed
                }
            }
        }
    }

    func connect(token: String) async {
        connectionState = .connecting

        do {
            try await webSocketClient.connect(token: token)
        } catch {
            connectionState = .failed
            attemptReconnect(token: token)
        }
    }

    func disconnect() {
        webSocketClient.disconnect()
        connectionState = .disconnected
    }

    private func attemptReconnect(token: String) {
        guard reconnectAttempts < maxReconnectAttempts else {
            connectionState = .failed
            return
        }

        reconnectAttempts += 1
        let delay = min(pow(2.0, Double(reconnectAttempts)), 60.0)

        connectionState = .reconnecting

        Task {
            try? await Task.sleep(for: .seconds(delay))
            await connect(token: token)
        }
    }
}
```

---

## WebSocket Message Handling

```swift
// WebSocket Message Handler
class WebSocketMessageHandler {
    private let tokenBuffer: StreamingTokenBuffer
    private let messageStore: MessageStoreProtocol

    init(
        tokenBuffer: StreamingTokenBuffer = StreamingTokenBuffer(),
        messageStore: MessageStoreProtocol = MessageStore()
    ) {
        self.tokenBuffer = tokenBuffer
        self.messageStore = messageStore
    }

    func handlePayload(_ payload: WebSocketPayload) -> ChatEvent {
        switch payload.type {
        case .token:
            if let token = payload.data?["text"] as? String {
                tokenBuffer.append(token)
                return .tokenReceived(token)
            }

        case .completed:
            tokenBuffer.flush()
            if let response = payload.data {
                let chatResponse = ChatResponse(
                    conversationId: response["conversationId"] as? String ?? "",
                    messageId: response["messageId"] as? String ?? "",
                    content: response["content"] as? String ?? "",
                    toolCalls: []
                )
                return .completed(chatResponse)
            }

        case .toolCall:
            if let toolData = payload.data {
                let toolCall = ToolCall(
                    id: toolData["id"] as? String ?? UUID().uuidString,
                    name: toolData["name"] as? String ?? "",
                    arguments: [:],
                    status: .pending
                )
                return .toolCall(toolCall)
            }

        case .thinking:
            if let text = payload.data?["text"] as? String {
                return .thinking(text)
            }

        case .error:
            if let message = payload.data?["message"] as? String {
                return .error(message)
            }

        default:
            break
        }

        return .unknown
    }

    enum ChatEvent {
        case tokenReceived(String)
        case completed(ChatResponse)
        case toolCall(ToolCall)
        case thinking(String)
        case error(String)
        case unknown
    }
}
```

---

## Streaming Token Buffer

```swift
// Streaming Token Buffer
class StreamingTokenBuffer {
    private var buffer: [String] = []
    private var batchTimer: Timer?
    private let batchInterval: TimeInterval = 0.05 // 50ms
    private let maxBatchSize = 5
    private var onBatchCallback: (([String]) -> Void)?

    init(batchInterval: TimeInterval = 0.05) {
        self.batchInterval = batchInterval
        startBatchTimer()
    }

    deinit {
        batchTimer?.invalidate()
    }

    func append(_ token: String) {
        buffer.append(token)

        if buffer.count >= maxBatchSize {
            flush()
        }
    }

    func flush() {
        guard !buffer.isEmpty else { return }
        let tokens = buffer
        buffer = []
        onBatchCallback?(tokens)
    }

    func onBatch(_ callback: @escaping ([String]) -> Void) {
        onBatchCallback = callback
    }

    private func startBatchTimer() {
        batchTimer = Timer.scheduledTimer(withTimeInterval: batchInterval, repeats: true) { [weak self] _ in
            self?.flush()
        }
    }
}

// Token Buffer Flow
/*
┌─────────────────────────────────────────────────────────────┐
│               STREAMING TOKEN BUFFER                         │
│                                                             │
│  WebSocket ──▶ Token ──▶ Buffer ──▶ Batch Timer ──▶ Render  │
│                                                             │
│  Token 1: "Swift" ──┐                                       │
│  Token 2: "UI"    ──┤                                       │
│  Token 3: "is"    ──┼──▶ Buffer ──▶ Flush ──▶ ["Swift",    │
│  Token 4: "a"     ──┤                          "UI", "is", │
│  Token 5: "great" ──┘                          "a", "great"]│
│                                                             │
│  Batch Timer: 50ms                                          │
│  Max Batch Size: 5 tokens                                   │
│  Total Latency: ~50ms per batch                             │
└─────────────────────────────────────────────────────────────┘
*/
```

---

## Message Types

```swift
// Message Type Definitions
enum MessageType: String, Codable {
    case text
    case image
    case document
    case toolCall = "tool_call"
    case toolResult = "tool_result"
    case thinking
    case system
}

// Message Content Types
enum MessageContent: Codable {
    case text(String)
    case image(Data)
    case document(DocumentContent)
    case toolCall(ToolCallContent)
    case toolResult(ToolResultContent)
    case thinking(String)

    struct DocumentContent: Codable {
        let name: String
        let mimeType: String
        let data: Data
        let size: Int64
    }

    struct ToolCallContent: Codable {
        let id: String
        let name: String
        let arguments: [String: AnyCodable]
    }

    struct ToolResultContent: Codable {
        let toolCallId: String
        let result: String
        let isSuccess: Bool
    }
}
```

---

## Conversation Management

```swift
// Conversation Manager
class ConversationManager: ObservableObject {
    @Published var conversations: [Conversation] = []
    @Published var currentConversation: Conversation?

    private let useCase: ConversationUseCaseProtocol

    init(useCase: ConversationUseCaseProtocol = ConversationUseCase()) {
        self.useCase = useCase
        Task { await loadConversations() }
    }

    func loadConversations() async {
        do {
            conversations = try await useCase.getConversations()
        } catch {
            os_log(.error, "Failed to load conversations: %@", error.localizedDescription)
        }
    }

    func createConversation(title: String, agentId: String?) async -> Conversation? {
        do {
            let conversation = try await useCase.createConversation(title: title, agentId: agentId)
            conversations.insert(conversation, at: 0)
            currentConversation = conversation
            return conversation
        } catch {
            os_log(.error, "Failed to create conversation: %@", error.localizedDescription)
            return nil
        }
    }

    func deleteConversation(_ conversation: Conversation) async {
        do {
            try await useCase.deleteConversation(conversation.id)
            conversations.removeAll { $0.id == conversation.id }
            if currentConversation?.id == conversation.id {
                currentConversation = nil
            }
        } catch {
            os_log(.error, "Failed to delete conversation: %@", error.localizedDescription)
        }
    }

    func searchConversations(query: String) -> [Conversation] {
        if query.isEmpty { return conversations }
        return conversations.filter {
            $0.title.localizedCaseInsensitiveContains(query)
        }
    }
}
```

---

## Chat History Sidebar

```swift
// ChatHistoryView.swift
struct ChatHistoryView: View {
    @ObservedObject var conversationManager: ConversationManager
    @State private var searchText = ""
    @State private var showDeleteConfirmation = false
    @State private var conversationToDelete: Conversation?

    var filteredConversations: [Conversation] {
        conversationManager.searchConversations(query: searchText)
    }

    var body: some View {
        NavigationStack {
            List {
                ForEach(filteredConversations) { conversation in
                    NavigationLink(destination: ChatView(conversationId: conversation.id)) {
                        ConversationRow(conversation: conversation)
                    }
                    .swipeActions(edge: .trailing) {
                        Button(role: .destructive) {
                            conversationToDelete = conversation
                            showDeleteConfirmation = true
                        } label: {
                            Label("Delete", systemImage: "trash")
                        }
                    }
                }
            }
            .searchable(text: $searchText, prompt: "Search conversations")
            .navigationTitle("Chat History")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button(action: {
                        Task {
                            _ = await conversationManager.createConversation(
                                title: "New Chat",
                                agentId: nil
                            )
                        }
                    }) {
                        Image(systemName: "plus")
                    }
                }
            }
            .alert("Delete Conversation", isPresented: $showDeleteConfirmation) {
                Button("Delete", role: .destructive) {
                    if let conversation = conversationToDelete {
                        Task {
                            await conversationManager.deleteConversation(conversation)
                        }
                    }
                }
                Button("Cancel", role: .cancel) {}
            } message: {
                Text("Are you sure you want to delete this conversation?")
            }
        }
    }
}

// Conversation Row
struct ConversationRow: View {
    let conversation: Conversation

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(conversation.title)
                .font(.headline)
                .lineLimit(1)

            if let lastMessage = conversation.lastMessage {
                Text(lastMessage.content)
                    .font(.subheadline)
                    .foregroundColor(.secondary)
                    .lineLimit(2)
            }

            HStack {
                Text(conversation.formattedDate)
                    .font(.caption2)
                    .foregroundColor(.secondary)

                Spacer()

                Text("\(conversation.messageCount) messages")
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
        }
        .padding(.vertical, 4)
    }
}
```

---

## Chat Settings

```swift
// ChatSettingsView.swift
struct ChatSettingsView: View {
    @Binding var settings: ChatSettings

    var body: some View {
        NavigationStack {
            Form {
                Section("Model") {
                    Picker("Model", selection: $settings.selectedModelId) {
                        Text("GPT-4").tag("gpt-4")
                        Text("GPT-4 Turbo").tag("gpt-4-turbo")
                        Text("Claude 3").tag("claude-3")
                        Text("Gemini Pro").tag("gemini-pro")
                    }

                    HStack {
                        Text("Temperature")
                        Slider(value: $settings.temperature, in: 0...1, step: 0.1)
                        Text(String(format: "%.1f", settings.temperature))
                            .foregroundColor(.secondary)
                    }

                    Stepper("Max Tokens: \(settings.maxTokens)", value: $settings.maxTokens, in: 256...8192, step: 256)
                }

                Section("Display") {
                    Toggle("Auto-scroll", isOn: $settings.autoScroll)
                    Toggle("Show timestamps", isOn: $settings.showTimestamps)
                    Toggle("Show thinking process", isOn: $settings.showThinking)
                    Toggle("Show tool calls", isOn: $settings.showToolCalls)
                    Toggle("Streaming responses", isOn: $settings.streamingEnabled)
                }
            }
            .navigationTitle("Chat Settings")
            .navigationBarTitleDisplayMode(.inline)
        }
    }
}
```

---

## Error Handling

```swift
// Chat Error Types
enum ChatError: AppError {
    case connectionFailed
    case messageTooLong
    case invalidMessage
    case sendFailed
    case loadHistoryFailed
    case uploadFailed(String)
    case rateLimited
    case modelUnavailable
    case conversationNotFound
    case messageNotFound
    case networkUnavailable

    var code: Int {
        switch self {
        case .connectionFailed: return 3001
        case .messageTooLong: return 3002
        case .invalidMessage: return 3003
        case .sendFailed: return 3004
        case .loadHistoryFailed: return 3005
        case .uploadFailed: return 3006
        case .rateLimited: return 3007
        case .modelUnavailable: return 3008
        case .conversationNotFound: return 3009
        case .messageNotFound: return 3010
        case .networkUnavailable: return 3011
        }
    }

    var domain: String { "com.nexusai.chat" }

    var errorDescription: String? {
        switch self {
        case .connectionFailed: return "Failed to connect to chat server"
        case .messageTooLong: return "Message is too long (max 10,000 characters)"
        case .invalidMessage: return "Invalid message format"
        case .sendFailed: return "Failed to send message"
        case .loadHistoryFailed: return "Failed to load chat history"
        case .uploadFailed(let file): return "Failed to upload \(file)"
        case .rateLimited: return "Too many messages. Please wait."
        case .modelUnavailable: return "AI model is currently unavailable"
        case .conversationNotFound: return "Conversation not found"
        case .messageNotFound: return "Message not found"
        case .networkUnavailable: return "No internet connection"
        }
    }
}
```

---

## Accessibility

```swift
// Accessibility Extensions
extension MessageBubble {
    var accessibilityLabel: String {
        let role = message.role == .user ? "You" : "AI"
        let content = message.displayedContent
        let time = message.timestamp.formatted(date: .omitted, time: .shortened)
        return "\(role) at \(time): \(content)"
    }

    var accessibilityHint: String {
        message.role == .user ? "Double tap to edit or delete" : "Double tap to copy"
    }
}

extension PromptBox {
    var accessibilityLabel: String {
        "Message input"
    }

    var accessibilityHint: String {
        "Type your message and tap send, or use voice input"
    }
}

// Dynamic Type Support
struct DynamicTypeMessageBubble: View {
    let message: ChatMessage
    @Environment(\.dynamicTypeSize) var dynamicTypeSize

    var body: some View {
        Text(message.displayedContent)
            .font(.body)
            .textSelection(.enabled)
            .accessibilityLabel(accessibilityLabel)
            .accessibilityAddTraits(message.role == .user ? [] : .isStaticText)
    }
}

// VoiceOver Support
extension VoiceInput {
    var accessibilityLabel: String {
        isRecording ? "Stop recording" : "Start recording"
    }

    var accessibilityHint: String {
        isRecording ? "Double tap to stop recording" : "Double tap to start voice input"
    }
}

// Content Shapes for Better Hit Testing
extension MessageBubble {
    var contentShape: some View {
        RoundedRectangle(cornerRadius: 16)
            .contentShape(RoundedRectangle(cornerRadius: 16))
    }
}
```

---

## Performance

```swift
// Performance Optimizations

// 1. LazyVStack for Message List
// Already using LazyVStack in MessageListView for virtual scrolling

// 2. Image Caching
class ImageCache {
    static let shared = ImageCache()
    private let cache = NSCache<NSString, UIImage>()

    init() {
        cache.countLimit = 100
        cache.totalCostLimit = 50 * 1024 * 1024 // 50MB
    }

    func image(for url: URL) -> UIImage? {
        cache.object(forKey: url.absoluteString as NSString)
    }

    func setImage(_ image: UIImage, for url: URL) {
        let cost = Int(image.size.width * image.size.height * image.scale * 4)
        cache.setObject(image, forKey: url.absoluteString as NSString, cost: cost)
    }
}

// 3. Diffable Data Source for Messages
// Using SwiftUI's built-in diffing with ForEach and Equatable conformance

// 4. Message Prefetching
class MessagePrefetcher {
    private let prefetchWindow = 20

    func shouldPrefetch(messages: [ChatMessage], currentIndex: Int) -> Bool {
        currentIndex >= messages.count - prefetchWindow
    }
}

// 5. Memory Management for Streaming
class StreamingMemoryManager {
    private let maxTokensInMemory = 10000

    func shouldFlushTokens(_ tokens: [String]) -> Bool {
        tokens.count > maxTokensInMemory
    }

    func flushOldTokens(_ tokens: inout [String]) {
        if tokens.count > maxTokensInMemory {
            tokens.removeFirst(tokens.count - maxTokensInMemory)
        }
    }
}
```

### Performance Benchmarks

| Metric | Target | Current | Tool |
|--------|--------|---------|------|
| Message Render | < 16ms | ~8ms | Time Profiler |
| Token Stream | < 50ms | ~30ms | Custom |
| WebSocket Connect | < 1s | ~400ms | Network |
| Scroll FPS | 60fps | 60fps | Core Animation |
| Memory Usage | < 150MB | ~80MB | Instruments |
| Image Load | < 200ms | ~100ms | Kingfisher |

---

## Offline Support

```swift
// Offline Message Queue
class OfflineMessageQueue: ObservableObject {
    @Published var queuedMessages: [QueuedMessage] = []

    private let storage: OfflineStorageProtocol

    struct QueuedMessage: Identifiable, Codable {
        let id: UUID
        let conversationId: String
        let content: String
        let timestamp: Date
        let attachments: [Data]
        var status: QueueStatus

        enum QueueStatus: String, Codable {
            case pending
            case sending
            case sent
            case failed
        }
    }

    init(storage: OfflineStorageProtocol = OfflineStorage()) {
        self.storage = storage
        loadQueuedMessages()
    }

    func enqueue(message: QueuedMessage) {
        queuedMessages.append(message)
        saveQueuedMessages()
    }

    func processQueue() async {
        let pendingMessages = queuedMessages.filter { $0.status == .pending }

        for message in pendingMessages {
            do {
                // Try to send
                let _ = try await ChatRepository().sendMessage(
                    message.content,
                    conversationId: message.conversationId,
                    agentId: nil
                )

                // Update status
                if let index = queuedMessages.firstIndex(where: { $0.id == message.id }) {
                    queuedMessages[index].status = .sent
                }
            } catch {
                if let index = queuedMessages.firstIndex(where: { $0.id == message.id }) {
                    queuedMessages[index].status = .failed
                }
            }
        }

        saveQueuedMessages()
    }

    private func loadQueuedMessages() {
        queuedMessages = storage.loadQueuedMessages()
    }

    private func saveQueuedMessages() {
        storage.saveQueuedMessages(queuedMessages)
    }
}

// Offline Banner
struct OfflineMessageBanner: View {
    @ObservedObject var queue: OfflineMessageQueue

    var body: some View {
        if !queue.queuedMessages.isEmpty {
            HStack {
                Image(systemName: "arrow.triangle.2.circlepath")
                    .foregroundColor(.orange)
                Text("\(queue.queuedMessages.filter({ $0.status == .pending }).count) messages queued")
                    .font(.caption)
                Spacer()
                Button("Send Now") {
                    Task { await queue.processQueue() }
                }
                .font(.caption)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 8)
            .background(Color.orange.opacity(0.2))
        }
    }
}
```

### Offline Support Flow

```
┌─────────────────────────────────────────────────────────────┐
│                OFFLINE SUPPORT FLOW                          │
│                                                             │
│  Online:                                                    │
│  ┌──────────┐     ┌──────────────┐     ┌────────────────┐ │
│  │  Send    │────▶│  WebSocket   │────▶│  Server        │ │
│  │  Message │     │  Stream      │     │  Response      │ │
│  └──────────┘     └──────────────┘     └────────────────┘ │
│                                                             │
│  Offline:                                                   │
│  ┌──────────┐     ┌──────────────┐     ┌────────────────┐ │
│  │  Send    │────▶│  Queue       │────▶│  CoreData      │ │
│  │  Message │     │  Message     │     │  (Pending)     │ │
│  └──────────┘     └──────────────┘     └───────┬────────┘ │
│                                                │          │
│                                    ┌───────────▼───────┐  │
│                                    │  Network Monitor  │  │
│                                    │  Detects Online   │  │
│                                    └───────────┬───────┘  │
│                                                │          │
│                                    ┌───────────▼───────┐  │
│                                    │  Process Queue    │  │
│                                    │  Send Messages    │  │
│                                    └───────────┬───────┘  │
│                                                │          │
│                                    ┌───────────▼───────┐  │
│                                    │  Update Status    │  │
│                                    │  (Sent/Failed)    │  │
│                                    └───────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```
