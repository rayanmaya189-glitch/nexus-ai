# AI Chat

## Table of Contents

1. [Chat Screen Layout](#chat-screen-layout)
2. [ChatViewModel](#chatviewmodel)
3. [MessageList Component](#messagelist-component)
4. [MessageBubble Component](#messagebubble-component)
5. [PromptBox Component](#promptbox-component)
6. [FileUploader Component](#fileuploader-component)
7. [VoiceInput Component](#voiceinput-component)
8. [AgentSelector Component](#agentselector-component)
9. [ModelIndicator Component](#modelindicator-component)
10. [TokenStreamViewer Component](#tokenstreamviewer-component)
11. [ThinkingIndicator Component](#thinkingindicator-component)
12. [ToolExecutionDisplay Component](#toolexecutiondisplay-component)
13. [WebSocket Connection Management](#websocket-connection-management)
14. [WebSocket Message Handling](#websocket-message-handling)
15. [Streaming Token Buffer](#streaming-token-buffer)
16. [Message Types](#message-types)
17. [Conversation Management](#conversation-management)
18. [Chat History Sidebar](#chat-history-sidebar)
19. [Chat Settings](#chat-settings)
20. [Image Upload in Chat](#image-upload-in-chat)
21. [File Attachment in Chat](#file-attachment-in-chat)
22. [Error Handling](#error-handling)
23. [Accessibility](#accessibility)
24. [Performance](#performance)
25. [Offline Support](#offline-support)

---

## Chat Screen Layout

```kotlin
// feature-chat/ui/ChatScreen.kt
@Composable
fun ChatScreen(
    viewModel: ChatViewModel = hiltViewModel(),
    onNavigateToAgents: () -> Unit,
    onNavigateToSettings: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val drawerState = rememberDrawerState(DrawerValue.Closed)

    LaunchedEffect(Unit) {
        viewModel.events.collect { event ->
            when (event) {
                is ChatEvent.ShowError -> { /* Show snackbar */ }
                is ChatEvent.NavigateToLogin -> { /* Navigate */ }
                is ChatEvent.MessageSent -> { /* Scroll to bottom */ }
            }
        }
    }

    ModalNavigationDrawer(
        drawerState = drawerState,
        drawerContent = {
            ChatHistoryDrawer(
                conversations = uiState.conversations,
                selectedId = uiState.currentConversationId,
                onConversationSelected = {
                    viewModel.onAction(ChatAction.SelectConversation(it))
                },
                onNewConversation = {
                    viewModel.onAction(ChatAction.NewConversation)
                },
                onDeleteConversation = {
                    viewModel.onAction(ChatAction.DeleteConversation(it))
                }
            )
        }
    ) {
        Scaffold(
            topBar = {
                ChatTopBar(
                    agent = uiState.selectedAgent,
                    model = uiState.selectedModel,
                    isConnected = uiState.isWebSocketConnected,
                    conversationTitle = uiState.conversationTitle,
                    onMenuClick = {
                        CoroutineScope(Dispatchers.Main).launch {
                            drawerState.open()
                        }
                    },
                    onAgentClick = onNavigateToAgents,
                    onSettingsClick = onNavigateToSettings
                )
            },
            bottomBar = {
                PromptBox(
                    isLoading = uiState.isLoading,
                    isStreaming = uiState.isStreaming,
                    onSendMessage = { viewModel.onAction(ChatAction.SendMessage(it)) },
                    onUploadFile = { viewModel.onAction(ChatAction.UploadFile(it)) },
                    onVoiceInput = { viewModel.onAction(ChatAction.StartVoiceInput) },
                    onCancelGeneration = { viewModel.onAction(ChatAction.CancelGeneration) }
                )
            }
        ) { paddingValues ->
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues)
            ) {
                // Agent selector (if in multi-agent mode)
                if (uiState.availableAgents.isNotEmpty()) {
                    AgentSelector(
                        agents = uiState.availableAgents,
                        selectedAgent = uiState.selectedAgent,
                        onAgentSelected = {
                            viewModel.onAction(ChatAction.SelectAgent(it.id))
                        }
                    )
                }

                // Message list
                MessageList(
                    messages = uiState.messages,
                    streamingText = uiState.streamingText,
                    isStreaming = uiState.isStreaming,
                    thinkingContent = uiState.thinkingContent,
                    toolExecutions = uiState.toolExecutions,
                    onLoadMore = { viewModel.onAction(ChatAction.LoadMore) },
                    onRetryMessage = { viewModel.onAction(ChatAction.RetryMessage(it)) },
                    onDeleteMessage = { viewModel.onAction(ChatAction.DeleteMessage(it)) },
                    modifier = Modifier.weight(1f)
                )

                // Model indicator
                if (uiState.selectedModel != null) {
                    ModelIndicator(
                        model = uiState.selectedModel!!,
                        tokenUsage = uiState.currentTokenUsage
                    )
                }
            }
        }
    }
}
```

**Chat Screen Layout Diagram:**

```
┌─────────────────────────────────────────────────────────┐
│ ☰  Nexus AI  │  Agent: Assistant  │  ● Connected  ⚙   │
├─────────────────────────────────────────────────────────┤
│                                                          │
│                    Chat History                           │
│                                                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │  👤 User                                        │    │
│  │  Hello, how are you today?                       │    │
│  │                                   10:30 AM       │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │  🤖 Assistant                                    │    │
│  │  I'm doing well, thank you! How can I help      │    │
│  │  you today?                                      │    │
│  │                                   10:30 AM       │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │  🤖 Assistant (streaming...)                     │    │
│  │  I'm analyzing your request...                   │    │
│  │  ┌──────────────────────────────────────────┐   │    │
│  │  │ 🔍 Searching knowledge base...  ✓        │   │    │
│  │  │ 📝 Generating response...  ✓             │   │    │
│  │  └──────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │  Model: GPT-4  │  Tokens: 1,234  │  Cost: $0.02 │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
├─────────────────────────────────────────────────────────┤
│  📎  │ Type your message...                     │  🎤  │
│       └─────────────────────────────────────────┘  ➤   │
└─────────────────────────────────────────────────────────┘
```

---

## ChatViewModel

```kotlin
// feature-chat/ui/ChatViewModel.kt
@HiltViewModel
class ChatViewModel @Inject constructor(
    private val sendMessageUseCase: SendMessageUseCase,
    private val observeMessagesUseCase: ObserveMessagesUseCase,
    private val loadConversationHistoryUseCase: LoadConversationHistoryUseCase,
    private val createConversationUseCase: CreateConversationUseCase,
    private val getAgentsUseCase: GetAgentsUseCase,
    private val selectAgentUseCase: SelectAgentUseCase,
    private val webSocketManager: WebSocketManager,
    private val tokenManager: TokenManager,
    private val networkMonitor: NetworkMonitor,
    private val analyticsManager: AnalyticsManager
) : ViewModel() {

    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    private val _events = Channel<ChatEvent>(Channel.BUFFERED)
    val events: Flow<ChatEvent> = _events.receiveAsFlow()

    private var streamingBuffer = StreamingTokenBuffer()
    private var currentConversationId: String? = null

    init {
        observeMessages()
        observeWebSocketEvents()
        loadAgents()
        createNewConversation()
    }

    fun onAction(action: ChatAction) {
        when (action) {
            is ChatAction.SendMessage -> sendMessage(action.text)
            is ChatAction.UploadFile -> uploadFile(action.uri)
            is ChatAction.StartVoiceInput -> startVoiceInput()
            is ChatAction.SelectAgent -> selectAgent(action.agentId)
            is ChatAction.SelectConversation -> selectConversation(action.conversationId)
            is ChatAction.NewConversation -> createNewConversation()
            is ChatAction.DeleteConversation -> deleteConversation(action.conversationId)
            is ChatAction.CancelGeneration -> cancelGeneration()
            is ChatAction.RetryMessage -> retryMessage(action.messageId)
            is ChatAction.DeleteMessage -> deleteMessage(action.messageId)
            is ChatAction.LoadMore -> loadMoreHistory()
            is ChatAction.UpdateInput -> updateInput(action.text)
        }
    }

    private fun observeMessages() {
        currentConversationId?.let { conversationId ->
            observeMessagesUseCase(conversationId)
                .onEach { messages ->
                    _uiState.update { it.copy(messages = messages) }
                }
                .launchIn(viewModelScope)
        }
    }

    private fun observeWebSocketEvents() {
        webSocketManager.events
            .onEach { event ->
                when (event) {
                    is WebSocketEvent.Token -> handleToken(event)
                    is WebSocketEvent.Thinking -> handleThinking(event)
                    is WebSocketEvent.ToolCall -> handleToolCall(event)
                    is WebSocketEvent.ToolResult -> handleToolResult(event)
                    is WebSocketEvent.Completed -> handleCompleted(event)
                    is WebSocketEvent.Error -> handleError(event)
                    is WebSocketEvent.Connected -> handleConnected(event)
                }
            }
            .launchIn(viewModelScope)

        webSocketManager.connectionState
            .onEach { state ->
                _uiState.update {
                    it.copy(isWebSocketConnected = state == ConnectionState.CONNECTED)
                }
            }
            .launchIn(viewModelScope)
    }

    private fun sendMessage(text: String) {
        if (text.isBlank()) return

        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    isLoading = true,
                    isStreaming = false,
                    streamingText = "",
                    thinkingContent = "",
                    inputText = ""
                )
            }

            analyticsManager.logChatMessage(
                agentId = _uiState.value.selectedAgent?.id ?: "default",
                modelId = _uiState.value.selectedModel?.id ?: "default",
                messageLength = text.length
            )

            sendMessageUseCase(
                text = text,
                conversationId = currentConversationId ?: return@launch,
                agentId = _uiState.value.selectedAgent?.id
            ).collect { result ->
                when (result) {
                    is Result.Loading -> {
                        _uiState.update { it.copy(isLoading = true) }
                    }
                    is Result.Success -> {
                        _uiState.update { it.copy(isLoading = false) }
                        _events.send(ChatEvent.MessageSent)
                    }
                    is Result.Error -> {
                        _uiState.update { it.copy(isLoading = false) }
                        _events.send(ChatEvent.ShowError(result.exception.message))
                    }
                }
            }
        }
    }

    private fun handleToken(event: WebSocketEvent.Token) {
        streamingBuffer.appendToken(event.content)
        val currentText = streamingBuffer.getCurrentText()

        _uiState.update {
            it.copy(
                isStreaming = true,
                isStreamingComplete = false,
                streamingText = currentText
            )
        }
    }

    private fun handleThinking(event: WebSocketEvent.Thinking) {
        _uiState.update {
            it.copy(thinkingContent = event.content)
        }
    }

    private fun handleToolCall(event: WebSocketEvent.ToolCall) {
        val toolExecution = ToolExecution(
            id = event.id,
            name = event.name,
            arguments = event.arguments,
            status = ToolExecutionStatus.RUNNING
        )
        _uiState.update {
            it.copy(
                toolExecutions = it.toolExecutions + toolExecution
            )
        }
    }

    private fun handleToolResult(event: WebSocketEvent.ToolResult) {
        _uiState.update { state ->
            val updatedExecutions = state.toolExecutions.map {
                if (it.id == event.toolCallId) {
                    it.copy(
                        status = if (event.success) ToolExecutionStatus.COMPLETED
                        else ToolExecutionStatus.FAILED,
                        result = event.content
                    )
                } else it
            }
            state.copy(toolExecutions = updatedExecutions)
        }
    }

    private fun handleCompleted(event: WebSocketEvent.Completed) {
        val finalText = streamingBuffer.getCurrentText()
        streamingBuffer.clear()

        _uiState.update {
            it.copy(
                isStreaming = false,
                isStreamingComplete = true,
                streamingText = "",
                thinkingContent = "",
                toolExecutions = emptyList(),
                currentTokenUsage = event.tokenUsage
            )
        }
    }

    private fun handleError(event: WebSocketEvent.Error) {
        streamingBuffer.clear()
        _uiState.update {
            it.copy(
                isStreaming = false,
                isLoading = false,
                streamingText = "",
                thinkingContent = ""
            )
        }
        _events.send(ChatEvent.ShowError(event.message))
    }

    private fun handleConnected(event: WebSocketEvent.Connected) {
        _uiState.update {
            it.copy(isWebSocketConnected = true)
        }
    }

    private fun cancelGeneration() {
        webSocketManager.cancelGeneration()
        streamingBuffer.clear()
        _uiState.update {
            it.copy(
                isStreaming = false,
                isLoading = false,
                streamingText = "",
                thinkingContent = ""
            )
        }
    }

    private fun loadAgents() {
        viewModelScope.launch {
            getAgentsUseCase().collect { result ->
                when (result) {
                    is Result.Success -> {
                        _uiState.update {
                            it.copy(availableAgents = result.data)
                        }
                    }
                    else -> { /* Handle error */ }
                }
            }
        }
    }

    private fun selectAgent(agentId: String) {
        viewModelScope.launch {
            selectAgentUseCase(agentId).collect { result ->
                when (result) {
                    is Result.Success -> {
                        _uiState.update {
                            it.copy(selectedAgent = result.data)
                        }
                    }
                    else -> { /* Handle error */ }
                }
            }
        }
    }

    private fun createNewConversation() {
        viewModelScope.launch {
            createConversationUseCase().collect { result ->
                when (result) {
                    is Result.Success -> {
                        currentConversationId = result.data.id
                        _uiState.update {
                            it.copy(
                                currentConversationId = result.data.id,
                                conversationTitle = result.data.title
                            )
                        }
                        webSocketManager.connect(result.data.id)
                        observeMessages()
                    }
                    else -> { /* Handle error */ }
                }
            }
        }
    }

    private fun selectConversation(conversationId: String) {
        currentConversationId = conversationId
        webSocketManager.disconnect()
        webSocketManager.connect(conversationId)
        observeMessages()
    }

    private fun loadMoreHistory() {
        viewModelScope.launch {
            currentConversationId?.let { conversationId ->
                loadConversationHistoryUseCase(conversationId, limit = 50)
                    .collect { result ->
                        when (result) {
                            is Result.Success -> {
                                _uiState.update {
                                    it.copy(messages = result.data + it.messages)
                                }
                            }
                            else -> { /* Handle error */ }
                        }
                    }
            }
        }
    }
}

data class ChatUiState(
    val messages: List<Message> = emptyList(),
    val conversations: List<Conversation> = emptyList(),
    val currentConversationId: String? = null,
    val conversationTitle: String = "New Chat",
    val isLoading: Boolean = false,
    val isStreaming: Boolean = false,
    val isStreamingComplete: Boolean = false,
    val streamingText: String = "",
    val thinkingContent: String = "",
    val toolExecutions: List<ToolExecution> = emptyList(),
    val selectedAgent: Agent? = null,
    val selectedModel: Model? = null,
    val availableAgents: List<Agent> = emptyList(),
    val isWebSocketConnected: Boolean = false,
    val inputText: String = "",
    val currentTokenUsage: TokenUsage? = null,
    val error: String? = null
)

sealed class ChatAction {
    data class SendMessage(val text: String) : ChatAction()
    data class UploadFile(val uri: Uri) : ChatAction()
    data object StartVoiceInput : ChatAction()
    data class SelectAgent(val agentId: String) : ChatAction()
    data class SelectConversation(val conversationId: String) : ChatAction()
    data object NewConversation : ChatAction()
    data class DeleteConversation(val conversationId: String) : ChatAction()
    data object CancelGeneration : ChatAction()
    data class RetryMessage(val messageId: String) : ChatAction()
    data class DeleteMessage(val messageId: String) : ChatAction()
    data object LoadMore : ChatAction()
    data class UpdateInput(val text: String) : ChatAction()
}

sealed class ChatEvent {
    data object MessageSent : ChatEvent()
    data class ShowError(val message: String?) : ChatEvent()
    data object NavigateToLogin : ChatEvent()
}
```

**ViewModel State Diagram:**

```
┌─────────────────────────────────────────────────────────┐
│                    ChatViewModel                         │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────┐   │
│  │  UiState     │    │  Events      │    │  Actions │   │
│  │  (StateFlow) │    │  (Channel)   │    │  (Input) │   │
│  └──────┬───────┘    └──────┬───────┘    └────┬─────┘   │
│         │                   │                 │          │
│         │    ┌──────────────┴───────────────┐ │          │
│         │    │                              │ │          │
│         ▼    ▼                              ▼ ▼          │
│  ┌──────────────────────────────────────────────────┐   │
│  │                   Use Cases                       │   │
│  │  ├── SendMessageUseCase                          │   │
│  │  ├── ObserveMessagesUseCase                      │   │
│  │  ├── LoadConversationHistoryUseCase              │   │
│  │  ├── CreateConversationUseCase                   │   │
│  │  ├── GetAgentsUseCase                            │   │
│  │  └── SelectAgentUseCase                          │   │
│  └──────────────────────────────────────────────────┘   │
│                         │                                │
│                         ▼                                │
│  ┌──────────────────────────────────────────────────┐   │
│  │                   Repositories                    │   │
│  │  ├── MessageRepository                           │   │
│  │  ├── ConversationRepository                      │   │
│  │  └── AgentRepository                             │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## MessageList Component

```kotlin
// feature-chat/ui/components/MessageList.kt
@Composable
fun MessageList(
    messages: List<Message>,
    streamingText: String,
    isStreaming: Boolean,
    thinkingContent: String,
    toolExecutions: List<ToolExecution>,
    onLoadMore: () -> Unit,
    onRetryMessage: (String) -> Unit,
    onDeleteMessage: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    val listState = rememberLazyListState()
    val coroutineScope = rememberCoroutineScope()

    // Auto-scroll to bottom when new messages arrive
    LaunchedEffect(messages.size, streamingText) {
        if (messages.isNotEmpty()) {
            listState.animateScrollToItem(messages.size - 1)
        }
    }

    // Load more when reaching top
    LaunchedEffect(listState) {
        snapshotFlow { listState.firstVisibleItemIndex }
            .filter { it == 0 }
            .collect {
                onLoadMore()
            }
    }

    LazyColumn(
        state = listState,
        modifier = modifier
            .fillMaxSize()
            .padding(horizontal = 16.dp),
        contentPadding = PaddingValues(vertical = 8.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        // Messages
        items(
            items = messages,
            key = { it.id }
        ) { message ->
            MessageBubble(
                message = message,
                onRetry = { onRetryMessage(message.id) },
                onDelete = { onDeleteMessage(message.id) }
            )
        }

        // Thinking indicator
        if (thinkingContent.isNotEmpty()) {
            item {
                ThinkingIndicator(content = thinkingContent)
            }
        }

        // Streaming message
        if (isStreaming && streamingText.isNotEmpty()) {
            item {
                StreamingMessageBubble(content = streamingText)
            }
        }

        // Tool executions
        if (toolExecutions.isNotEmpty()) {
            item {
                ToolExecutionDisplay(executions = toolExecutions)
            }
        }

        // Loading indicator
        if (isStreaming && streamingText.isEmpty() && thinkingContent.isEmpty()) {
            item {
                LoadingIndicator()
            }
        }
    }
}

@Composable
private fun LoadingIndicator() {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalArrangement = Arrangement.Start
    ) {
        CircularProgressIndicator(
            modifier = Modifier.size(24.dp),
            strokeWidth = 2.dp
        )
        Spacer(modifier = Modifier.width(8.dp))
        Text(
            text = "Thinking...",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}
```

---

## MessageBubble Component

```kotlin
// feature-chat/ui/components/MessageBubble.kt
@Composable
fun MessageBubble(
    message: Message,
    onRetry: () -> Unit,
    onDelete: () -> Unit
) {
    val isUser = message.role == MessageRole.USER
    var showMenu by remember { mutableStateOf(false) }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        horizontalArrangement = if (isUser) Arrangement.End else Arrangement.Start
    ) {
        if (!isUser) {
            // AI avatar
            AsyncImage(
                model = message.agentAvatarUrl,
                contentDescription = "AI Avatar",
                modifier = Modifier
                    .size(32.dp)
                    .clip(CircleShape)
            )
            Spacer(modifier = Modifier.width(8.dp))
        }

        Column(
            modifier = Modifier
                .widthIn(max = 300.dp)
                .background(
                    color = if (isUser) MaterialTheme.colorScheme.primary
                    else MaterialTheme.colorScheme.surfaceVariant,
                    shape = RoundedCornerShape(
                        topStart = if (isUser) 16.dp else 4.dp,
                        topEnd = if (isUser) 4.dp else 16.dp,
                        bottomStart = 16.dp,
                        bottomEnd = 16.dp
                    )
                )
                .padding(12.dp)
                .combinedClickable(
                    onClick = {},
                    onLongClick = { showMenu = true }
                ),
            horizontalAlignment = if (isUser) Alignment.End else Alignment.Start
        ) {
            // Message content
            Text(
                text = message.content,
                color = if (isUser) MaterialTheme.colorScheme.onPrimary
                else MaterialTheme.colorScheme.onSurfaceVariant,
                style = MaterialTheme.typography.bodyMedium
            )

            // Tool calls display
            if (message.toolCalls.isNotEmpty()) {
                Spacer(modifier = Modifier.height(8.dp))
                message.toolCalls.forEach { toolCall ->
                    ToolCallChip(toolCall = toolCall)
                }
            }

            // Tool results display
            if (message.toolResults.isNotEmpty()) {
                Spacer(modifier = Modifier.height(8.dp))
                message.toolResults.forEach { toolResult ->
                    ToolResultChip(toolResult = toolResult)
                }
            }

            // Timestamp
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = formatTimestamp(message.timestamp),
                style = MaterialTheme.typography.labelSmall,
                color = if (isUser) MaterialTheme.colorScheme.onPrimary.copy(alpha = 0.7f)
                else MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f)
            )
        }

        if (isUser) {
            Spacer(modifier = Modifier.width(8.dp))
            // User avatar
            AsyncImage(
                model = message.userAvatarUrl,
                contentDescription = "User Avatar",
                modifier = Modifier
                    .size(32.dp)
                    .clip(CircleShape)
            )
        }
    }

    // Context menu
    DropdownMenu(
        expanded = showMenu,
        onDismissRequest = { showMenu = false }
    ) {
        DropdownMenuItem(
            text = { Text("Copy") },
            onClick = {
                showMenu = false
                // Copy to clipboard
            },
            leadingIcon = { Icon(Icons.Default.ContentCopy, contentDescription = null) }
        )
        if (!isUser) {
            DropdownMenuItem(
                text = { Text("Retry") },
                onClick = {
                    showMenu = false
                    onRetry()
                },
                leadingIcon = { Icon(Icons.Default.Refresh, contentDescription = null) }
            )
        }
        DropdownMenuItem(
            text = { Text("Delete") },
            onClick = {
                showMenu = false
                onDelete()
            },
            leadingIcon = { Icon(Icons.Default.Delete, contentDescription = null) }
        )
    }
}

@Composable
fun StreamingMessageBubble(content: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        horizontalArrangement = Arrangement.Start
    ) {
        AsyncImage(
            model = null,
            contentDescription = "AI Avatar",
            modifier = Modifier
                .size(32.dp)
                .clip(CircleShape)
                .background(MaterialTheme.colorScheme.surfaceVariant)
        )
        Spacer(modifier = Modifier.width(8.dp))

        Column(
            modifier = Modifier
                .widthIn(max = 300.dp)
                .background(
                    color = MaterialTheme.colorScheme.surfaceVariant,
                    shape = RoundedCornerShape(4.dp, 16.dp, 16.dp, 16.dp)
                )
                .padding(12.dp)
        ) {
            Text(
                text = buildAnnotatedString {
                    append(content)
                    // Blinking cursor
                    withStyle(SpanStyle(color = MaterialTheme.colorScheme.primary)) {
                        append("▌")
                    }
                },
                style = MaterialTheme.typography.bodyMedium
            )
        }
    }
}

private fun formatTimestamp(timestamp: Long): String {
    val sdf = SimpleDateFormat("hh:mm a", Locale.getDefault())
    return sdf.format(Date(timestamp))
}
```

**Message Bubble Variants:**

```
User Message:                    AI Message:
┌──────────────────────┐        ┌──────────────────────┐
│  Hello, how are you? │  👤    │  🤖  I'm doing well!  │
│        10:30 AM      │        │  10:30 AM             │
└──────────────────────┘        └──────────────────────┘

Streaming Message:               Tool Execution:
┌──────────────────────┐        ┌──────────────────────┐
│  🤖  I'm analyzing...│        │  🤖  Found results!   │
│      ▌                │        │  🔍 Search: ✓         │
│        10:30 AM       │        │  10:30 AM             │
└──────────────────────┘        └──────────────────────┘
```

---

## PromptBox Component

```kotlin
// feature-chat/ui/components/PromptBox.kt
@Composable
fun PromptBox(
    isLoading: Boolean,
    isStreaming: Boolean,
    onSendMessage: (String) -> Unit,
    onUploadFile: (Uri) -> Unit,
    onVoiceInput: () -> Unit,
    onCancelGeneration: () -> Unit
) {
    var text by remember { mutableStateOf("") }
    var showAttachmentMenu by remember { mutableStateOf(false) }
    val focusRequester = remember { FocusRequester() }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(MaterialTheme.colorScheme.surface)
            .padding(horizontal = 16.dp, vertical = 8.dp)
    ) {
        // Attachment menu
        if (showAttachmentMenu) {
            AttachmentMenu(
                onCameraCapture = { uri -> onUploadFile(uri) },
                onGalleryPick = { uri -> onUploadFile(uri) },
                onDocumentPick = { uri -> onUploadFile(uri) },
                onDismiss = { showAttachmentMenu = false }
            )
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(
                    color = MaterialTheme.colorScheme.surfaceVariant,
                    shape = RoundedCornerShape(24.dp)
                )
                .padding(4.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Attachment button
            IconButton(
                onClick = { showAttachmentMenu = true }
            ) {
                Icon(
                    imageVector = Icons.Default.AttachFile,
                    contentDescription = "Attach file",
                    tint = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            // Text input
            BasicTextField(
                value = text,
                onValueChange = { text = it },
                modifier = Modifier
                    .weight(1f)
                    .focusRequester(focusRequester)
                    .padding(horizontal = 8.dp),
                textStyle = MaterialTheme.typography.bodyMedium,
                maxLines = 5,
                decorationBox = { innerTextField ->
                    Box {
                        if (text.isEmpty()) {
                            Text(
                                text = "Type your message...",
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f)
                            )
                        }
                        innerTextField()
                    }
                }
            )

            // Voice input or Send button
            if (isStreaming) {
                IconButton(
                    onClick = onCancelGeneration
                ) {
                    Icon(
                        imageVector = Icons.Default.Stop,
                        contentDescription = "Stop generation",
                        tint = MaterialTheme.colorScheme.error
                    )
                }
            } else if (text.isBlank()) {
                IconButton(
                    onClick = onVoiceInput
                ) {
                    Icon(
                        imageVector = Icons.Default.Mic,
                        contentDescription = "Voice input",
                        tint = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            } else {
                IconButton(
                    onClick = {
                        onSendMessage(text)
                        text = ""
                    },
                    enabled = !isLoading
                ) {
                    if (isLoading) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(24.dp),
                            strokeWidth = 2.dp
                        )
                    } else {
                        Icon(
                            imageVector = Icons.Default.Send,
                            contentDescription = "Send message",
                            tint = MaterialTheme.colorScheme.primary
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun AttachmentMenu(
    onCameraCapture: (Uri) -> Unit,
    onGalleryPick: (Uri) -> Unit,
    onDocumentPick: (Uri) -> Unit,
    onDismiss: () -> Unit
) {
    val context = LocalContext.current
    val cameraLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.TakePicturePreview()
    ) { bitmap ->
        bitmap?.let {
            // Convert bitmap to URI and upload
            val uri = saveBitmapToCache(context, it)
            onCameraCapture(uri)
        }
        onDismiss()
    }

    val galleryLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent()
    ) { uri ->
        uri?.let { onGalleryPick(it) }
        onDismiss()
    }

    val documentLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.OpenDocument()
    ) { uri ->
        uri?.let { onDocumentPick(it) }
        onDismiss()
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(bottom = 8.dp),
        horizontalArrangement = Arrangement.SpaceEvenly
    ) {
        AttachmentOption(
            icon = Icons.Default.CameraAlt,
            label = "Camera",
            onClick = { cameraLauncher.launch(null) }
        )
        AttachmentOption(
            icon = Icons.Default.PhotoLibrary,
            label = "Gallery",
            onClick = { galleryLauncher.launch("image/*") }
        )
        AttachmentOption(
            icon = Icons.Default.InsertDriveFile,
            label = "Document",
            onClick = { documentLauncher.launch(arrayOf("*/*")) }
        )
    }
}

@Composable
private fun AttachmentOption(
    icon: ImageVector,
    label: String,
    onClick: () -> Unit
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.clickable(onClick = onClick)
    ) {
        Icon(
            imageVector = icon,
            contentDescription = label,
            modifier = Modifier.size(48.dp),
            tint = MaterialTheme.colorScheme.primary
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}
```

---

## FileUploader Component

```kotlin
// feature-chat/ui/components/FileUploader.kt
@Composable
fun FileUploader(
    uploadedFiles: List<UploadedFile>,
    onFileSelected: (Uri) -> Unit,
    onFileRemoved: (String) -> Unit
) {
    val context = LocalContext.current

    val cameraLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.TakePicturePreview()
    ) { bitmap ->
        bitmap?.let {
            val uri = saveBitmapToCache(context, it)
            onFileSelected(uri)
        }
    }

    val galleryLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent()
    ) { uri ->
        uri?.let { onFileSelected(it) }
    }

    val documentLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.OpenDocument()
    ) { uri ->
        uri?.let { onFileSelected(it) }
    }

    Column {
        // Uploaded files preview
        if (uploadedFiles.isNotEmpty()) {
            LazyRow(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(vertical = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                items(uploadedFiles) { file ->
                    FilePreview(
                        file = file,
                        onRemove = { onFileRemoved(file.id) }
                    )
                }
            }
        }
    }
}

@Composable
private fun FilePreview(
    file: UploadedFile,
    onRemove: () -> Unit
) {
    Box(
        modifier = Modifier
            .size(80.dp)
            .clip(RoundedCornerShape(8.dp))
            .background(MaterialTheme.colorScheme.surfaceVariant)
    ) {
        when (file.type) {
            FileType.IMAGE -> {
                AsyncImage(
                    model = file.uri,
                    contentDescription = file.name,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop
                )
            }
            FileType.DOCUMENT -> {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(8.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Icon(
                        imageVector = Icons.Default.InsertDriveFile,
                        contentDescription = null,
                        modifier = Modifier.size(32.dp),
                        tint = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = file.name,
                        style = MaterialTheme.typography.labelSmall,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis
                    )
                }
            }
        }

        // Progress indicator
        if (file.isUploading) {
            CircularProgressIndicator(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(16.dp),
                progress = { file.progress / 100f }
            )
        }

        // Remove button
        IconButton(
            onClick = onRemove,
            modifier = Modifier.align(Alignment.TopEnd)
        ) {
            Icon(
                imageVector = Icons.Default.Close,
                contentDescription = "Remove",
                modifier = Modifier.size(16.dp),
                tint = MaterialTheme.colorScheme.error
            )
        }
    }
}

data class UploadedFile(
    val id: String,
    val name: String,
    val uri: Uri,
    val type: FileType,
    val size: Long,
    val isUploading: Boolean = false,
    val progress: Int = 0,
    val url: String? = null
)

enum class FileType {
    IMAGE, DOCUMENT, AUDIO, VIDEO
}
```

---

## VoiceInput Component

```kotlin
// feature-chat/ui/components/VoiceInput.kt
@Composable
fun VoiceInput(
    isRecording: Boolean,
    recordingDuration: Long,
    onStartRecording: () -> Unit,
    onStopRecording: () -> Unit,
    onCancelRecording: () -> Unit
) {
    val context = LocalContext.current
    val permissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { granted ->
        if (granted) {
            onStartRecording()
        }
    }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        if (isRecording) {
            // Recording UI
            Text(
                text = "Recording...",
                style = MaterialTheme.typography.titleMedium,
                color = MaterialTheme.colorScheme.error
            )

            Spacer(modifier = Modifier.height(8.dp))

            // Duration
            Text(
                text = formatDuration(recordingDuration),
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Waveform animation
            WaveformAnimation(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(64.dp)
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Control buttons
            Row(
                horizontalArrangement = Arrangement.SpaceEvenly,
                modifier = Modifier.fillMaxWidth()
            ) {
                // Cancel button
                IconButton(
                    onClick = onCancelRecording,
                    modifier = Modifier
                        .size(56.dp)
                        .background(
                            MaterialTheme.colorScheme.surfaceVariant,
                            CircleShape
                        )
                ) {
                    Icon(
                        imageVector = Icons.Default.Close,
                        contentDescription = "Cancel",
                        tint = MaterialTheme.colorScheme.error
                    )
                }

                // Stop button
                IconButton(
                    onClick = onStopRecording,
                    modifier = Modifier
                        .size(72.dp)
                        .background(
                            MaterialTheme.colorScheme.error,
                            CircleShape
                        )
                ) {
                    Icon(
                        imageVector = Icons.Default.Stop,
                        contentDescription = "Stop recording",
                        tint = MaterialTheme.colorScheme.onPrimary,
                        modifier = Modifier.size(32.dp)
                    )
                }

                // Spacer for symmetry
                Spacer(modifier = Modifier.size(56.dp))
            }
        } else {
            // Microphone button
            IconButton(
                onClick = {
                    permissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
                },
                modifier = Modifier
                    .size(72.dp)
                    .background(
                        MaterialTheme.colorScheme.primary,
                        CircleShape
                    )
            ) {
                Icon(
                    imageVector = Icons.Default.Mic,
                    contentDescription = "Start recording",
                    tint = MaterialTheme.colorScheme.onPrimary,
                    modifier = Modifier.size(32.dp)
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "Tap to speak",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
private fun WaveformAnimation(modifier: Modifier = Modifier) {
    val infiniteTransition = rememberInfiniteTransition(label = "waveform")
    val amplitude by infiniteTransition.animateFloat(
        initialValue = 0.3f,
        targetValue = 0.7f,
        animationSpec = infiniteRepeatable(
            animation = tween(500, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "amplitude"
    )

    Canvas(modifier = modifier) {
        val barWidth = 4.dp.toPx()
        val gap = 2.dp.toPx()
        val centerY = size.height / 2

        drawContext.canvas.nativeCanvas.apply {
            val paint = android.graphics.Paint().apply {
                color = android.graphics.Color.RED
                style = android.graphics.Paint.Style.FILL
            }

            var x = 0f
            while (x < size.width) {
                val height = (amplitude * size.height * 0.5f) *
                        (0.5f + Math.random().toFloat() * 0.5f)
                drawRoundRect(
                    x, centerY - height / 2,
                    x + barWidth, centerY + height / 2,
                    barWidth / 2, barWidth / 2,
                    paint
                )
                x += barWidth + gap
            }
        }
    }
}

private fun formatDuration(duration: Long): String {
    val minutes = (duration / 1000) / 60
    val seconds = (duration / 1000) % 60
    return String.format("%02d:%02d", minutes, seconds)
}
```

---

## AgentSelector Component

```kotlin
// feature-chat/ui/components/AgentSelector.kt
@Composable
fun AgentSelector(
    agents: List<Agent>,
    selectedAgent: Agent?,
    onAgentSelected: (Agent) -> Unit
) {
    var expanded by remember { mutableStateOf(false) }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        // Selected agent display
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier.clickable { expanded = true }
        ) {
            AsyncImage(
                model = selectedAgent?.avatarUrl,
                contentDescription = null,
                modifier = Modifier
                    .size(32.dp)
                    .clip(CircleShape)
            )
            Spacer(modifier = Modifier.width(8.dp))
            Column {
                Text(
                    text = selectedAgent?.name ?: "Select Agent",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Medium
                )
                if (selectedAgent?.description != null) {
                    Text(
                        text = selectedAgent.description,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis
                    )
                }
            }
            Spacer(modifier = Modifier.width(4.dp))
            Icon(
                imageVector = Icons.Default.ArrowDropDown,
                contentDescription = null
            )
        }

        // Agent capabilities
        if (selectedAgent != null) {
            Row {
                selectedAgent.capabilities.take(3).forEach { capability ->
                    CapabilityChip(capability = capability)
                    Spacer(modifier = Modifier.width(4.dp))
                }
            }
        }
    }

    // Dropdown menu
    DropdownMenu(
        expanded = expanded,
        onDismissRequest = { expanded = false }
    ) {
        agents.forEach { agent ->
            DropdownMenuItem(
                text = {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        AsyncImage(
                            model = agent.avatarUrl,
                            contentDescription = null,
                            modifier = Modifier
                                .size(40.dp)
                                .clip(CircleShape)
                        )
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text(
                                text = agent.name,
                                style = MaterialTheme.typography.bodyMedium,
                                fontWeight = FontWeight.Medium
                            )
                            Text(
                                text = agent.description ?: "",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                },
                onClick = {
                    onAgentSelected(agent)
                    expanded = false
                },
                leadingIcon = {
                    if (agent.id == selectedAgent?.id) {
                        Icon(
                            imageVector = Icons.Default.Check,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.primary
                        )
                    }
                }
            )
        }
    }
}

@Composable
private fun CapabilityChip(capability: String) {
    Surface(
        shape = RoundedCornerShape(12.dp),
        color = MaterialTheme.colorScheme.primaryContainer
    ) {
        Text(
            text = capability,
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onPrimaryContainer
        )
    }
}
```

---

## ModelIndicator Component

```kotlin
// feature-chat/ui/components/ModelIndicator.kt
@Composable
fun ModelIndicator(
    model: Model,
    tokenUsage: TokenUsage?
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(MaterialTheme.colorScheme.surfaceVariant)
            .padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        // Model info
        Row(verticalAlignment = Alignment.CenterVertically) {
            Icon(
                imageVector = Icons.Default.SmartToy,
                contentDescription = null,
                modifier = Modifier.size(16.dp),
                tint = MaterialTheme.colorScheme.primary
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = model.name,
                style = MaterialTheme.typography.labelMedium,
                fontWeight = FontWeight.Medium
            )
            Spacer(modifier = Modifier.width(8.dp))
            StatusBadge(status = model.status)
        }

        // Token usage
        if (tokenUsage != null) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    imageVector = Icons.Default.Token,
                    contentDescription = null,
                    modifier = Modifier.size(16.dp),
                    tint = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.width(4.dp))
                Text(
                    text = "${tokenUsage.totalTokens} tokens",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

@Composable
private fun StatusBadge(status: String) {
    val color = when (status.lowercase()) {
        "online", "active" -> MaterialTheme.colorScheme.primary
        "busy" -> MaterialTheme.colorScheme.error
        "offline" -> MaterialTheme.colorScheme.onSurfaceVariant
        else -> MaterialTheme.colorScheme.onSurfaceVariant
    }

    Surface(
        shape = RoundedCornerShape(8.dp),
        color = color.copy(alpha = 0.1f)
    ) {
        Text(
            text = status,
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
            style = MaterialTheme.typography.labelSmall,
            color = color
        )
    }
}

data class TokenUsage(
    val promptTokens: Int,
    val completionTokens: Int,
    val totalTokens: Int
)
```

---

## TokenStreamViewer Component

```kotlin
// feature-chat/ui/components/TokenStreamViewer.kt
@Composable
fun TokenStreamViewer(
    content: String,
    isComplete: Boolean
) {
    val annotatedString = buildAnnotatedString {
        append(content)
        if (!isComplete) {
            withStyle(SpanStyle(color = MaterialTheme.colorScheme.primary)) {
                append("▌")
            }
        }
    }

    Text(
        text = annotatedString,
        style = MaterialTheme.typography.bodyMedium,
        modifier = Modifier.padding(8.dp)
    )
}

// Streaming token buffer
class StreamingTokenBuffer {
    private val buffer = StringBuilder()
    private val tokens = mutableListOf<String>()
    private var lastRenderTime = 0L
    private val renderInterval = 50L // 50ms between renders

    @Synchronized
    fun appendToken(token: String) {
        tokens.add(token)
        buffer.append(token)
    }

    @Synchronized
    fun getCurrentText(): String {
        return buffer.toString()
    }

    @Synchronized
    fun shouldRender(): Boolean {
        val currentTime = System.currentTimeMillis()
        if (currentTime - lastRenderTime >= renderInterval) {
            lastRenderTime = currentTime
            return true
        }
        return false
    }

    @Synchronized
    fun getRenderableText(): String {
        return buffer.toString()
    }

    @Synchronized
    fun clear() {
        buffer.clear()
        tokens.clear()
        lastRenderTime = 0L
    }

    @Synchronized
    fun getTokenCount(): Int {
        return tokens.size
    }
}
```

---

## ThinkingIndicator Component

```kotlin
// feature-chat/ui/components/ThinkingIndicator.kt
@Composable
fun ThinkingIndicator(content: String) {
    val infiniteTransition = rememberInfiniteTransition(label = "thinking")
    val dots by infiniteTransition.animateValue(
        initialValue = 0,
        targetValue = 3,
        animationSpec = infiniteRepeatable(
            animation = tween(1000, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "dots"
    )

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalArrangement = Arrangement.Start,
        verticalAlignment = Alignment.CenterVertically
    ) {
        // Animated dots
        repeat(3) { index ->
            Box(
                modifier = Modifier
                    .size(8.dp)
                    .background(
                        color = if (index <= dots)
                            MaterialTheme.colorScheme.primary
                        else MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.3f),
                        shape = CircleShape
                    )
            )
            if (index < 2) Spacer(modifier = Modifier.width(4.dp))
        }

        Spacer(modifier = Modifier.width(12.dp))

        // Thinking text
        Text(
            text = content.ifEmpty { "Thinking" },
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}
```

**Thinking Indicator States:**

```
Planning...          Searching...         Checking...
● ○ ○                ○ ● ○                ○ ○ ●
Planning...          Searching...         Checking knowledge base...
```

---

## ToolExecutionDisplay Component

```kotlin
// feature-chat/ui/components/ToolExecutionDisplay.kt
@Composable
fun ToolExecutionDisplay(
    executions: List<ToolExecution>
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp)
    ) {
        executions.forEach { execution ->
            ToolExecutionItem(execution = execution)
            if (execution != executions.last()) {
                Spacer(modifier = Modifier.height(4.dp))
            }
        }
    }
}

@Composable
private fun ToolExecutionItem(execution: ToolExecution) {
    var expanded by remember { mutableStateOf(false) }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(8.dp))
            .background(MaterialTheme.colorScheme.surfaceVariant)
            .clickable { expanded = !expanded }
            .padding(12.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        // Status icon
        when (execution.status) {
            ToolExecutionStatus.RUNNING -> {
                CircularProgressIndicator(
                    modifier = Modifier.size(16.dp),
                    strokeWidth = 2.dp
                )
            }
            ToolExecutionStatus.COMPLETED -> {
                Icon(
                    imageVector = Icons.Default.CheckCircle,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.primary,
                    modifier = Modifier.size(16.dp)
                )
            }
            ToolExecutionStatus.FAILED -> {
                Icon(
                    imageVector = Icons.Default.Error,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.error,
                    modifier = Modifier.size(16.dp)
                )
            }
        }

        Spacer(modifier = Modifier.width(8.dp))

        // Tool name
        Text(
            text = execution.name,
            style = MaterialTheme.typography.bodyMedium,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.weight(1f)
        )

        // Expand icon
        Icon(
            imageVector = if (expanded) Icons.Default.ExpandLess
            else Icons.Default.ExpandMore,
            contentDescription = null,
            modifier = Modifier.size(16.dp)
        )
    }

    // Expanded content
    if (expanded) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(start = 32.dp, top = 8.dp)
        ) {
            // Arguments
            if (execution.arguments.isNotEmpty()) {
                Text(
                    text = "Arguments:",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Text(
                    text = execution.arguments,
                    style = MaterialTheme.typography.bodySmall,
                    fontFamily = FontFamily.Monospace
                )
            }

            // Result
            if (execution.result != null) {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "Result:",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Text(
                    text = execution.result,
                    style = MaterialTheme.typography.bodySmall,
                    fontFamily = FontFamily.Monospace
                )
            }
        }
    }
}

data class ToolExecution(
    val id: String,
    val name: String,
    val arguments: String,
    val status: ToolExecutionStatus,
    val result: String? = null
)

enum class ToolExecutionStatus {
    RUNNING, COMPLETED, FAILED
}
```

**Tool Execution Display:**

```
┌─────────────────────────────────────────────────┐
│ 🔍 search_knowledge_base                    ▼   │
├─────────────────────────────────────────────────┤
│ Arguments:                                       │
│ {                                               │
│   "query": "Android networking",                │
│   "limit": 5                                    │
│ }                                               │
│                                                  │
│ Result:                                          │
│ Found 5 relevant documents about Android         │
│ networking patterns and best practices.           │
└─────────────────────────────────────────────────┘
│ ✓ generate_response                          ▲   │
└─────────────────────────────────────────────────┘
```

---

## WebSocket Connection Management

```kotlin
// data/remote/websocket/WebSocketManager.kt
@Singleton
class WebSocketManager @Inject constructor(
    private val okHttpClient: OkHttpClient,
    private val tokenManager: TokenManager,
    private val json: Json
) {
    private var webSocket: WebSocket? = null
    private val _events = MutableSharedFlow<WebSocketEvent>(
        replay = 1,
        extraBufferCapacity = 64
    )
    val events: SharedFlow<WebSocketEvent> = _events.asSharedFlow()

    private val _connectionState = MutableStateFlow(ConnectionState.DISCONNECTED)
    val connectionState: StateFlow<ConnectionState> = _connectionState.asStateFlow()

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private val reconnectManager = ReconnectManager(scope)

    fun connect(conversationId: String) {
        val token = tokenManager.getAccessToken() ?: return

        val request = Request.Builder()
            .url("${BuildConfig.WS_URL}/ws?token=$token&conversation_id=$conversationId")
            .build()

        _connectionState.value = ConnectionState.CONNECTING
        webSocket = okHttpClient.newWebSocket(request, createWebSocketListener())
    }

    fun disconnect() {
        reconnectManager.reset()
        webSocket?.close(1000, "Client disconnect")
        webSocket = null
        _connectionState.value = ConnectionState.DISCONNECTED
    }

    fun sendMessage(
        text: String,
        conversationId: String,
        agentId: String? = null,
        modelId: String? = null
    ) {
        val message = WebSocketMessage(
            type = "user_message",
            content = text,
            conversationId = conversationId,
            agentId = agentId,
            modelId = modelId
        )
        val jsonMessage = json.encodeToString(WebSocketMessage.serializer(), message)
        webSocket?.send(jsonMessage)
    }

    fun cancelGeneration() {
        val cancelMessage = """{"type":"cancel"}"""
        webSocket?.send(cancelMessage)
    }

    private fun createWebSocketListener(): WebSocketListener {
        return object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                super.onOpen(webSocket, response)
                _connectionState.value = ConnectionState.CONNECTED
                reconnectManager.reset()
                sendAuthMessage()
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                super.onMessage(webSocket, text)
                try {
                    val event = json.decodeFromString<WebSocketEvent>(text)
                    _events.tryEmit(event)
                } catch (e: Exception) {
                    Timber.e(e, "Failed to parse WebSocket message")
                }
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                super.onFailure(webSocket, t, response)
                _connectionState.value = ConnectionState.DISCONNECTED
                _events.tryEmit(WebSocketEvent.Error(t.message ?: "Connection lost"))
                reconnectManager.scheduleReconnect { connect("") }
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                super.onClosed(webSocket, code, reason)
                _connectionState.value = ConnectionState.DISCONNECTED
            }
        }
    }

    private fun sendAuthMessage() {
        val token = tokenManager.getAccessToken() ?: return
        val authMessage = WebSocketAuthMessage(token = token)
        val jsonMessage = json.encodeToString(WebSocketAuthMessage.serializer(), authMessage)
        webSocket?.send(jsonMessage)
    }
}

enum class ConnectionState {
    DISCONNECTED, CONNECTING, CONNECTED
}
```

---

## WebSocket Message Handling

```kotlin
// Message handling in WebSocketManager
private fun handleWebSocketMessage(text: String) {
    try {
        val jsonElement = json.parseToJsonElement(text)
        val type = jsonElement.jsonObject["type"]?.jsonPrimitive?.content

        when (type) {
            "token" -> {
                val event = json.decodeFromString<WebSocketEvent.Token>(text)
                _events.tryEmit(event)
            }
            "thinking" -> {
                val event = json.decodeFromString<WebSocketEvent.Thinking>(text)
                _events.tryEmit(event)
            }
            "tool_call" -> {
                val event = json.decodeFromString<WebSocketEvent.ToolCall>(text)
                _events.tryEmit(event)
            }
            "tool_result" -> {
                val event = json.decodeFromString<WebSocketEvent.ToolResult>(text)
                _events.tryEmit(event)
            }
            "completed" -> {
                val event = json.decodeFromString<WebSocketEvent.Completed>(text)
                _events.tryEmit(event)
            }
            "error" -> {
                val event = json.decodeFromString<WebSocketEvent.Error>(text)
                _events.tryEmit(event)
            }
            "connected" -> {
                val event = json.decodeFromString<WebSocketEvent.Connected>(text)
                _events.tryEmit(event)
            }
            "pong" -> {
                // Heartbeat response, do nothing
            }
            else -> {
                Timber.w("Unknown WebSocket message type: $type")
            }
        }
    } catch (e: Exception) {
        Timber.e(e, "Failed to handle WebSocket message")
    }
}
```

**Message Handling Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Raw JSON  │───►│   Parse     │───►│   Route     │
│   Message   │    │   Type      │    │   Handler   │
└─────────────┘    └─────────────┘    └──────┬──────┘
                                            │
                    ┌────────────────────────┼────────────────────────┐
                    │            │           │           │            │
                    ▼            ▼           ▼           ▼            ▼
             ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
             │  Token   │ │ Thinking │ │ ToolCall │ │Completed │ │  Error   │
             │  Handler │ │ Handler  │ │ Handler  │ │ Handler  │ │ Handler  │
             └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘
                    │            │           │           │            │
                    └────────────────────────┼────────────────────────┘
                                            │
                                     ┌──────┴──────┐
                                     │  Emit Event │
                                     │  to Flow    │
                                     └─────────────┘
```

---

## Streaming Token Buffer

```kotlin
// Streaming token buffer for smooth rendering
class StreamingTokenBuffer(
    private val renderInterval: Long = 50L,
    private val maxBufferSize: Int = 1000
) {
    private val buffer = StringBuilder()
    private val tokenQueue = ConcurrentLinkedQueue<String>()
    private var lastRenderTime = 0L
    private var renderJob: Job? = null
    private val scope = CoroutineScope(Dispatchers.Default)

    private val _renderedText = MutableStateFlow("")
    val renderedText: StateFlow<String> = _renderedText.asStateFlow()

    @Synchronized
    fun appendToken(token: String) {
        if (buffer.length + token.length > maxBufferSize) {
            // Trim buffer if too large
            val trimAmount = buffer.length / 2
            buffer.delete(0, trimAmount)
        }
        buffer.append(token)
        tokenQueue.offer(token)
    }

    fun startRendering(onRender: (String) -> Unit) {
        renderJob = scope.launch {
            while (isActive) {
                val currentTime = System.currentTimeMillis()
                if (currentTime - lastRenderTime >= renderInterval && tokenQueue.isNotEmpty()) {
                    val text = buffer.toString()
                    _renderedText.value = text
                    onRender(text)
                    lastRenderTime = currentTime
                }
                delay(10) // Small delay to prevent busy waiting
            }
        }
    }

    fun stopRendering() {
        renderJob?.cancel()
        // Final render
        val finalText = buffer.toString()
        _renderedText.value = finalText
    }

    @Synchronized
    fun getCurrentText(): String {
        return buffer.toString()
    }

    @Synchronized
    fun clear() {
        buffer.clear()
        tokenQueue.clear()
        _renderedText.value = ""
        lastRenderTime = 0L
    }

    @Synchronized
    fun getTokenCount(): Int {
        return tokenQueue.size
    }
}
```

---

## Message Types

```kotlin
// domain/entity/Message.kt
data class Message(
    val id: String,
    val conversationId: String,
    val role: MessageRole,
    val content: String,
    val timestamp: Long,
    val toolCalls: List<ToolCall> = emptyList(),
    val toolResults: List<ToolResult> = emptyList(),
    val tokenUsage: TokenUsage? = null,
    val modelId: String? = null,
    val agentId: String? = null,
    val attachments: List<Attachment> = emptyList(),
    val isStreaming: Boolean = false,
    val isEdited: Boolean = false,
    val metadata: Map<String, Any> = emptyMap()
)

enum class MessageRole {
    USER, ASSISTANT, SYSTEM, TOOL
}

data class ToolCall(
    val id: String,
    val name: String,
    val arguments: String,
    val status: ToolCallStatus = ToolCallStatus.PENDING
)

data class ToolResult(
    val toolCallId: String,
    val content: String,
    val success: Boolean,
    val executionTime: Long? = null
)

enum class ToolCallStatus {
    PENDING, RUNNING, COMPLETED, FAILED
}

data class Attachment(
    val id: String,
    val name: String,
    val url: String,
    val type: AttachmentType,
    val size: Long,
    val mimeType: String
)

enum class AttachmentType {
    IMAGE, DOCUMENT, AUDIO, VIDEO, CODE
}
```

---

## Conversation Management

```kotlin
// data/repository/ConversationRepositoryImpl.kt
class ConversationRepositoryImpl @Inject constructor(
    private val conversationDao: ConversationDao,
    private val apiService: ApiService,
    private val conversationMapper: ConversationMapper
) : ConversationRepository {

    override fun observeConversations(): Flow<List<Conversation>> {
        return conversationDao.observeConversations().map { entities ->
            entities.map { conversationMapper.toDomain(it) }
        }
    }

    override suspend fun createConversation(title: String?): Conversation {
        val response = apiService.createConversation(
            CreateConversationRequest(title = title ?: "New Chat")
        )

        if (response.isSuccessful) {
            val dto = response.body()!!.data
            val entity = conversationMapper.toEntity(dto)
            conversationDao.insertConversation(entity)
            return conversationMapper.toDomain(entity)
        } else {
            throw Exception("Failed to create conversation")
        }
    }

    override suspend fun deleteConversation(conversationId: String) {
        apiService.deleteConversation(conversationId)
        conversationDao.deleteConversation(conversationId)
    }

    override suspend fun updateConversationTitle(
        conversationId: String,
        title: String
    ) {
        conversationDao.updateTitle(conversationId, title)
    }

    override suspend fun searchConversations(query: String): List<Conversation> {
        return conversationDao.searchConversations("%$query%").map {
            conversationMapper.toDomain(it)
        }
    }
}

// data/local/entity/ConversationEntity.kt
@Entity(tableName = "conversations")
data class ConversationEntity(
    @PrimaryKey val id: String,
    @ColumnInfo(name = "title") val title: String,
    @ColumnInfo(name = "agent_id") val agentId: String?,
    @ColumnInfo(name = "model_id") val modelId: String?,
    @ColumnInfo(name = "created_at") val createdAt: Long,
    @ColumnInfo(name = "updated_at") val updatedAt: Long,
    @ColumnInfo(name = "message_count") val messageCount: Int = 0,
    @ColumnInfo(name = "is_pinned") val isPinned: Boolean = false
)

// data/local/dao/ConversationDao.kt
@Dao
interface ConversationDao {
    @Query("SELECT * FROM conversations ORDER BY updated_at DESC")
    fun observeConversations(): Flow<List<ConversationEntity>>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertConversation(conversation: ConversationEntity)

    @Query("DELETE FROM conversations WHERE id = :conversationId")
    suspend fun deleteConversation(conversationId: String)

    @Query("UPDATE conversations SET title = :title WHERE id = :conversationId")
    suspend fun updateTitle(conversationId: String, title: String)

    @Query("SELECT * FROM conversations WHERE title LIKE :query ORDER BY updated_at DESC")
    suspend fun searchConversations(query: String): List<ConversationEntity>

    @Query("UPDATE conversations SET updated_at = :timestamp WHERE id = :conversationId")
    suspend fun updateTimestamp(conversationId: String, timestamp: Long)
}
```

---

## Chat History Sidebar

```kotlin
// feature-chat/ui/components/ChatHistoryDrawer.kt
@Composable
fun ChatHistoryDrawer(
    conversations: List<Conversation>,
    selectedId: String?,
    onConversationSelected: (String) -> Unit,
    onNewConversation: () -> Unit,
    onDeleteConversation: (String) -> Unit
) {
    ModalDrawerSheet {
        // Header
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = "Chat History",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold
            )

            Spacer(modifier = Modifier.height(16.dp))

            // New conversation button
            Button(
                onClick = onNewConversation,
                modifier = Modifier.fillMaxWidth()
            ) {
                Icon(Icons.Default.Add, contentDescription = null)
                Spacer(modifier = Modifier.width(8.dp))
                Text("New Chat")
            }
        }

        HorizontalDivider()

        // Search
        var searchQuery by remember { mutableStateOf("") }
        OutlinedTextField(
            value = searchQuery,
            onValueChange = { searchQuery = it },
            placeholder = { Text("Search conversations...") },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = null) },
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 8.dp),
            singleLine = true
        )

        // Conversation list
        val filteredConversations = if (searchQuery.isBlank()) {
            conversations
        } else {
            conversations.filter {
                it.title.contains(searchQuery, ignoreCase = true)
            }
        }

        LazyColumn {
            items(
                items = filteredConversations,
                key = { it.id }
            ) { conversation ->
                ConversationItem(
                    conversation = conversation,
                    isSelected = conversation.id == selectedId,
                    onClick = { onConversationSelected(conversation.id) },
                    onDelete = { onDeleteConversation(conversation.id) }
                )
            }
        }
    }
}

@Composable
private fun ConversationItem(
    conversation: Conversation,
    isSelected: Boolean,
    onClick: () -> Unit,
    onDelete: () -> Unit
) {
    var showMenu by remember { mutableStateOf(false) }

    NavigationDrawerItem(
        label = {
            Column {
                Text(
                    text = conversation.title,
                    style = MaterialTheme.typography.bodyMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                Text(
                    text = formatDate(conversation.updatedAt),
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        },
        selected = isSelected,
        onClick = onClick,
        modifier = Modifier
            .padding(horizontal = 16.dp, vertical = 4.dp)
            .combinedClickable(
                onClick = onClick,
                onLongClick = { showMenu = true }
            )
    )

    DropdownMenu(
        expanded = showMenu,
        onDismissRequest = { showMenu = false }
    ) {
        DropdownMenuItem(
            text = { Text("Delete") },
            onClick = {
                showMenu = false
                onDelete()
            },
            leadingIcon = { Icon(Icons.Default.Delete, contentDescription = null) }
        )
    }
}

private fun formatDate(timestamp: Long): String {
    val sdf = SimpleDateFormat("MMM d, yyyy", Locale.getDefault())
    return sdf.format(Date(timestamp))
}
```

---

## Chat Settings

```kotlin
// feature-chat/ui/ChatSettingsScreen.kt
@Composable
fun ChatSettingsScreen(
    viewModel: ChatSettingsViewModel = hiltViewModel(),
    onBack: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Chat Settings") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { paddingValues ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Model selection
            item {
                SettingsSection(title = "Model") {
                    ModelSelectionDropdown(
                        models = uiState.availableModels,
                        selectedModel = uiState.selectedModel,
                        onModelSelected = {
                            viewModel.onAction(ChatSettingsAction.SelectModel(it.id))
                        }
                    )
                }
            }

            // Agent selection
            item {
                SettingsSection(title = "Default Agent") {
                    AgentSelectionDropdown(
                        agents = uiState.availableAgents,
                        selectedAgent = uiState.selectedAgent,
                        onAgentSelected = {
                            viewModel.onAction(ChatSettingsAction.SelectAgent(it.id))
                        }
                    )
                }
            }

            // Temperature
            item {
                SettingsSection(title = "Temperature") {
                    Slider(
                        value = uiState.temperature,
                        onValueChange = {
                            viewModel.onAction(ChatSettingsAction.UpdateTemperature(it))
                        },
                        valueRange = 0f..2f,
                        steps = 20
                    )
                    Text(
                        text = "%.2f".format(uiState.temperature),
                        style = MaterialTheme.typography.bodySmall
                    )
                }
            }

            // Max tokens
            item {
                SettingsSection(title = "Max Tokens") {
                    Slider(
                        value = uiState.maxTokens.toFloat(),
                        onValueChange = {
                            viewModel.onAction(ChatSettingsAction.UpdateMaxTokens(it.toInt()))
                        },
                        valueRange = 256f..4096f,
                        steps = 15
                    )
                    Text(
                        text = "${uiState.maxTokens}",
                        style = MaterialTheme.typography.bodySmall
                    )
                }
            }

            // System prompt
            item {
                SettingsSection(title = "System Prompt") {
                    OutlinedTextField(
                        value = uiState.systemPrompt,
                        onValueChange = {
                            viewModel.onAction(ChatSettingsAction.UpdateSystemPrompt(it))
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(120.dp),
                        placeholder = { Text("Enter system prompt...") }
                    )
                }
            }
        }
    }
}

@Composable
private fun SettingsSection(
    title: String,
    content: @Composable ColumnScope.() -> Unit
) {
    Column {
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold
        )
        Spacer(modifier = Modifier.height(8.dp))
        content()
    }
}

@Composable
private fun ModelSelectionDropdown(
    models: List<Model>,
    selectedModel: Model?,
    onModelSelected: (Model) -> Unit
) {
    var expanded by remember { mutableStateOf(false) }

    ExposedDropdownMenuBox(
        expanded = expanded,
        onExpandedChange = { expanded = it }
    ) {
        OutlinedTextField(
            value = selectedModel?.name ?: "Select Model",
            onValueChange = {},
            readOnly = true,
            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
            modifier = Modifier
                .fillMaxWidth()
                .menuAnchor()
        )

        ExposedDropdownMenu(
            expanded = expanded,
            onDismissRequest = { expanded = false }
        ) {
            models.forEach { model ->
                DropdownMenuItem(
                    text = {
                        Column {
                            Text(model.name)
                            Text(
                                text = model.description ?: "",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    },
                    onClick = {
                        onModelSelected(model)
                        expanded = false
                    }
                )
            }
        }
    }
}
```

---

## Image Upload in Chat

```kotlin
// Image upload handler
class ImageUploadHandler @Inject constructor(
    private val fileRepository: FileRepository,
    private val scope: CoroutineScope
) {
    private val _uploadState = MutableStateFlow<UploadState>(UploadState.Idle)
    val uploadState: StateFlow<UploadState> = _uploadState.asStateFlow()

    fun uploadImage(
        uri: Uri,
        context: Context,
        onProgress: (Int) -> Unit = {}
    ) {
        scope.launch {
            _uploadState.value = UploadState.Uploading(0)

            try {
                // Compress image
                val compressedUri = compressImage(context, uri)

                // Upload
                val result = fileRepository.uploadFile(
                    uri = compressedUri,
                    onProgress = { progress ->
                        _uploadState.value = UploadState.Uploading(progress)
                        onProgress(progress)
                    }
                )

                _uploadState.value = UploadState.Success(result.url)
            } catch (e: Exception) {
                _uploadState.value = UploadState.Error(e.message)
            }
        }
    }

    private suspend fun compressImage(context: Context, uri: Uri): Uri {
        return withContext(Dispatchers.IO) {
            val bitmap = MediaStore.Images.Media.getBitmap(context.contentResolver, uri)
            val compressed = ImageCompressor.compress(
                bitmap = bitmap,
                maxWidth = 1024,
                maxHeight = 1024,
                quality = 80
            )

            // Save compressed image to cache
            val file = File(context.cacheDir, "compressed_${System.currentTimeMillis()}.jpg")
            file.outputStream().use {
                compressed.compress(Bitmap.CompressFormat.JPEG, 80, it)
            }
            Uri.fromFile(file)
        }
    }
}

sealed class UploadState {
    data object Idle : UploadState()
    data class Uploading(val progress: Int) : UploadState()
    data class Success(val url: String) : UploadState()
    data class Error(val message: String?) : UploadState()
}
```

---

## File Attachment in Chat

```kotlin
// File attachment handler
class FileAttachmentHandler @Inject constructor(
    private val fileRepository: FileRepository,
    private val scope: CoroutineScope
) {
    private val _attachments = MutableStateFlow<List<Attachment>>(emptyList())
    val attachments: StateFlow<List<Attachment>> = _attachments.asStateFlow()

    fun addAttachment(uri: Uri, context: Context) {
        scope.launch {
            try {
                val fileInfo = getFileInfo(context, uri)
                val attachment = Attachment(
                    id = UUID.randomUUID().toString(),
                    name = fileInfo.name,
                    url = uri.toString(),
                    type = fileInfo.type,
                    size = fileInfo.size,
                    mimeType = fileInfo.mimeType
                )
                _attachments.update { it + attachment }

                // Upload in background
                uploadAttachment(attachment, context)
            } catch (e: Exception) {
                Timber.e(e, "Failed to add attachment")
            }
        }
    }

    private suspend fun uploadAttachment(attachment: Attachment, context: Context) {
        try {
            val uri = Uri.parse(attachment.url)
            val result = fileRepository.uploadFile(uri)
            _attachments.update { list ->
                list.map {
                    if (it.id == attachment.id) {
                        it.copy(url = result.url)
                    } else it
                }
            }
        } catch (e: Exception) {
            Timber.e(e, "Failed to upload attachment")
            removeAttachment(attachment.id)
        }
    }

    fun removeAttachment(attachmentId: String) {
        _attachments.update { list ->
            list.filter { it.id != attachmentId }
        }
    }

    fun clearAttachments() {
        _attachments.value = emptyList()
    }

    private fun getFileInfo(context: Context, uri: Uri): FileInfo {
        val cursor = context.contentResolver.query(uri, null, null, null, null)
        cursor?.use {
            if (it.moveToFirst()) {
                val nameIndex = it.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                val sizeIndex = it.getColumnIndex(OpenableColumns.SIZE)
                val name = it.getString(nameIndex) ?: "unknown"
                val size = it.getLong(sizeIndex)
                val mimeType = context.contentResolver.getType(uri) ?: "application/octet-stream"
                val type = when {
                    mimeType.startsWith("image/") -> AttachmentType.IMAGE
                    mimeType.startsWith("audio/") -> AttachmentType.AUDIO
                    mimeType.startsWith("video/") -> AttachmentType.VIDEO
                    mimeType.contains("pdf") || mimeType.contains("document") -> AttachmentType.DOCUMENT
                    else -> AttachmentType.DOCUMENT
                }
                return FileInfo(name, size, mimeType, type)
            }
        }
        return FileInfo("unknown", 0, "application/octet-stream", AttachmentType.DOCUMENT)
    }

    data class FileInfo(
        val name: String,
        val size: Long,
        val mimeType: String,
        val type: AttachmentType
    )
}
```

---

## Error Handling

```kotlin
// Chat error types
sealed class ChatError : Exception() {
    data class ConnectionLost(
        override val message: String = "Connection lost. Trying to reconnect..."
    ) : ChatError()

    data class MessageFailed(
        override val message: String = "Failed to send message"
    ) : ChatError()

    data class StreamingError(
        override val message: String = "Error during streaming"
    ) : ChatError()

    data class RateLimited(
        override val message: String = "Rate limited. Please wait a moment."
    ) : ChatError()

    data class ModelError(
        override val message: String = "Model error occurred"
    ) : ChatError()

    data class FileUploadError(
        override val message: String = "Failed to upload file"
    ) : ChatError()

    data class NetworkError(
        override val message: String = "Network error. Please check your connection."
    ) : ChatError()
}

// Error handler in ChatViewModel
private fun handleChatError(error: Throwable) {
    val chatError = when (error) {
        is IOException -> ChatError.NetworkError()
        is HttpException -> {
            when (error.code()) {
                429 -> ChatError.RateLimited()
                500, 502, 503 -> ChatError.ModelError()
                else -> ChatError.MessageFailed()
            }
        }
        is WebSocketException -> ChatError.ConnectionLost()
        else -> ChatError.MessageFailed(error.message)
    }

    viewModelScope.launch {
        _events.send(ChatEvent.ShowError(chatError.message))
    }
}

// Retry logic for failed messages
fun retryMessage(messageId: String) {
    viewModelScope.launch {
        val message = messageRepository.getMessageById(messageId) ?: return@launch

        _uiState.update { it.copy(isLoading = true) }

        sendMessageUseCase(
            text = message.content,
            conversationId = currentConversationId ?: return@launch,
            agentId = _uiState.value.selectedAgent?.id
        ).collect { result ->
            when (result) {
                is Result.Loading -> {
                    _uiState.update { it.copy(isLoading = true) }
                }
                is Result.Success -> {
                    // Delete old message
                    messageRepository.deleteMessage(messageId)
                    _uiState.update { it.copy(isLoading = false) }
                }
                is Result.Error -> {
                    _uiState.update { it.copy(isLoading = false) }
                    _events.send(ChatEvent.ShowError(result.exception.message))
                }
            }
        }
    }
}
```

**Error Handling Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Error     │───►│   Classify  │───►│   Show      │
│   Occurs    │    │   Error     │    │   Message   │
└─────────────┘    └─────────────┘    └──────┬──────┘
                                            │
                                     ┌──────┴──────┐
                                     │  Retryable? │
                                     └──────┬──────┘
                                            │
                               ┌────────────┼────────────┐
                               ▼            ▼            ▼
                        ┌──────────┐ ┌──────────┐ ┌──────────┐
                        │  Yes     │ │  No      │ │  Auto    │
                        │  Show    │ │  Show    │ │  Retry   │
                        │  Retry   │ │  Error   │ │  (3x)    │
                        └──────────┘ └──────────┘ └──────────┘
```

---

## Accessibility

```kotlin
// Accessibility features implementation
@Composable
fun MessageBubble(
    message: Message,
    onRetry: () -> Unit,
    onDelete: () -> Unit
) {
    val role = if (message.role == MessageRole.USER) "User message" else "AI response"

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .semantics {
                contentDescription = "$role: ${message.content}"
                stateDescription = if (message.isStreaming) "Streaming" else null
            }
            .padding(vertical = 4.dp),
        horizontalArrangement = if (message.role == MessageRole.USER) Arrangement.End
        else Arrangement.Start
    ) {
        // Message content with semantic properties
        Column(
            modifier = Modifier
                .widthIn(max = 300.dp)
                .semantics(mergeDescendants = true) {
                    // Merge all child semantics into one node
                }
        ) {
            // Avatar with content description
            if (message.role != MessageRole.USER) {
                AsyncImage(
                    model = message.agentAvatarUrl,
                    contentDescription = "AI assistant avatar",
                    modifier = Modifier
                        .size(32.dp)
                        .clip(CircleShape)
                )
            }

            // Message text
            Text(
                text = message.content,
                modifier = Modifier.semantics {
                    liveRegion = LiveRegionMode.Polite
                }
            )

            // Timestamp
            Text(
                text = formatTimestamp(message.timestamp),
                modifier = Modifier.semantics {
                    contentDescription = "Sent at ${formatTimeForAccessibility(message.timestamp)}"
                }
            )
        }

        // Action buttons with proper labels
        if (message.role != MessageRole.USER) {
            IconButton(
                onClick = onRetry,
                modifier = Modifier.semantics {
                    contentDescription = "Retry sending this message"
                }
            ) {
                Icon(Icons.Default.Refresh, contentDescription = null)
            }

            IconButton(
                onClick = onDelete,
                modifier = Modifier.semantics {
                    contentDescription = "Delete this message"
                }
            ) {
                Icon(Icons.Default.Delete, contentDescription = null)
            }
        }
    }
}

// Keyboard navigation support
@Composable
fun PromptBox(
    // ... parameters
) {
    val focusRequester = remember { FocusRequester() }
    val keyboardController = LocalSoftwareKeyboardController.current

    BasicTextField(
        value = text,
        onValueChange = { text = it },
        modifier = Modifier
            .focusRequester(focusRequester)
            .onKeyEvent { event ->
                if (event.type == KeyEventType.KeyDown) {
                    when (event.key) {
                        Key.Enter -> {
                            if (event.isCtrlPressed || event.isShiftPressed) {
                                // New line
                                false
                            } else {
                                // Send message
                                onSendMessage(text)
                                keyboardController?.hide()
                                true
                            }
                        }
                        else -> false
                    }
                } else false
            },
        // ... other parameters
    )
}

// Content descriptions for images
@Composable
fun MessageImage(imageUrl: String, description: String?) {
    AsyncImage(
        model = imageUrl,
        contentDescription = description ?: "Image in message",
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(8.dp))
            .semantics {
                if (description != null) {
                    contentDescription = "Image: $description"
                } else {
                    contentDescription = "Image attachment"
                }
            }
    )
}
```

**Accessibility Checklist:**

| Feature | Implementation | Status |
|---------|---------------|--------|
| Screen reader | Semantic properties | ✅ |
| Content descriptions | All images, icons | ✅ |
| Live regions | Streaming messages | ✅ |
| Keyboard navigation | All interactive elements | ✅ |
| Focus management | Auto-focus on input | ✅ |
| Contrast ratios | 4.5:1 minimum | ✅ |
| Touch targets | 48dp minimum | ✅ |
| Error announcements | Error messages | ✅ |
| State changes | Live region updates | ✅ |
| Navigation landmarks | Top bar, content, input | ✅ |

---

## Performance

```kotlin
// LazyColumn optimization
@Composable
fun MessageList(
    messages: List<Message>,
    // ... other parameters
) {
    val listState = rememberLazyListState()

    LazyColumn(
        state = listState,
        // Enable content placement for better performance
        contentPadding = PaddingValues(16.dp),
        // Use keys for stable recomposition
    ) {
        items(
            items = messages,
            key = { it.id },
            //contentType = { it.role } // For better item recycling
        ) { message ->
            // Use remember for expensive computations
            val formattedContent = remember(message.content) {
                formatMarkdown(message.content)
            }

            MessageBubble(
                message = message.copy(content = formattedContent),
                onRetry = { onRetryMessage(message.id) },
                onDelete = { onDeleteMessage(message.id) }
            )
        }
    }
}

// Image caching with Coil
@Composable
fun CachedImage(
    url: String,
    contentDescription: String?,
    modifier: Modifier = Modifier
) {
    AsyncImage(
        model = ImageRequest.Builder(LocalContext.current)
            .data(url)
            .crossfade(true)
            .memoryCachePolicy(CachePolicy.ENABLED)
            .diskCachePolicy(CachePolicy.ENABLED)
            .build(),
        contentDescription = contentDescription,
        modifier = modifier.clip(RoundedCornerShape(8.dp)),
        contentScale = ContentScale.Crop
    )
}

// Debounced search
@Composable
fun SearchBar(
    onSearch: (String) -> Unit
) {
    var query by remember { mutableStateOf("") }

    LaunchedEffect(query) {
        if (query.isNotEmpty()) {
            delay(300) // Debounce
            onSearch(query)
        }
    }

    OutlinedTextField(
        value = query,
        onValueChange = { query = it },
        // ... other parameters
    )
}
```

**Performance Metrics:**

| Metric | Target | Tool |
|--------|--------|------|
| First meaningful paint | < 1s | Baseline Profiles |
| Message render time | < 16ms | FrameMetrics |
| Image load time | < 500ms | Coil metrics |
| Search response | < 300ms | Debounce |
| Scroll performance | 60fps | LazyColumn |
| Memory usage | < 150MB | Profiler |
| Bundle size | < 30MB | R8 |

---

## Offline Support

```kotlin
// Offline message queue
@Singleton
class OfflineMessageQueue @Inject constructor(
    private val messageDao: MessageDao,
    private val networkMonitor: NetworkMonitor
) {
    private val pendingMessages = ConcurrentLinkedQueue<PendingMessage>()

    init {
        // Load pending messages from database
        CoroutineScope(Dispatchers.IO).launch {
            loadPendingMessages()
        }

        // Sync when network becomes available
        CoroutineScope(Dispatchers.Main).launch {
            networkMonitor.isOnline.collect { isOnline ->
                if (isOnline) {
                    syncPendingMessages()
                }
            }
        }
    }

    fun enqueueMessage(message: Message) {
        val pendingMessage = PendingMessage(
            id = message.id,
            conversationId = message.conversationId,
            content = message.content,
            timestamp = message.timestamp
        )
        pendingMessages.offer(pendingMessage)
        savePendingMessage(pendingMessage)
    }

    private suspend fun syncPendingMessages() {
        while (pendingMessages.isNotEmpty()) {
            val message = pendingMessages.poll() ?: break
            try {
                // Send message via WebSocket or API
                sendMessageToServer(message)
                deletePendingMessage(message.id)
            } catch (e: Exception) {
                // Re-enqueue if failed
                pendingMessages.offer(message)
                break
            }
        }
    }

    private suspend fun loadPendingMessages() {
        val messages = messageDao.getPendingMessages()
        messages.forEach { pendingMessages.offer(it) }
    }

    private suspend fun savePendingMessage(message: PendingMessage) {
        messageDao.insertPendingMessage(message)
    }

    private suspend fun deletePendingMessage(messageId: String) {
        messageDao.deletePendingMessage(messageId)
    }

    private suspend fun sendMessageToServer(message: PendingMessage) {
        // Implementation depends on your API
    }
}

data class PendingMessage(
    val id: String,
    val conversationId: String,
    val content: String,
    val timestamp: Long
)

// Offline indicator in ChatScreen
@Composable
fun OfflineIndicator(isOnline: Boolean) {
    if (!isOnline) {
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = MaterialTheme.colorScheme.errorContainer
        ) {
            Row(
                modifier = Modifier.padding(8.dp),
                horizontalArrangement = Arrangement.Center,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    imageVector = Icons.Default.CloudOff,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.onErrorContainer
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "You're offline. Messages will be sent when you're back online.",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onErrorContainer
                )
            }
        }
    }
}
```

**Offline Support Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  User       │───►│  Save to    │───►│  Show in    │
│  Sends      │    │  Local DB   │    │  UI         │
│  Message    │    │  (Pending)  │    │  (Pending)  │
└─────────────┘    └─────────────┘    └─────────────┘
                         │
                  ┌──────┴──────┐
                  │  Network    │
                  │  Available? │
                  └──────┬──────┘
                         │
              ┌──────────┼──────────┐
              ▼          ▼          ▼
       ┌──────────┐ ┌──────────┐ ┌──────────┐
       │  Yes     │ │  No      │ │  Wait    │
       │  Sync    │ │  Queue   │ │  for     │
       │  Now     │ │  Message │ │  Network │
       └──────────┘ └──────────┘ └──────────┘
            │
            ▼
       ┌──────────┐
       │  Send to │
       │  Server  │
       └──────────┘
```

---

*Document Version: 1.0 | Last Updated: 2026-07-16*
