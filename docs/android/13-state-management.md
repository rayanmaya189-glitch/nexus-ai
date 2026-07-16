# 13. State Management

## Table of Contents

1. [MVVM Architecture](#1-mvvm-architecture)
2. [ViewModel Lifecycle](#2-viewmodel-lifecycle)
3. [StateFlow for UI State](#3-stateflow-for-ui-state)
4. [SharedFlow for Events](#4-sharedflow-for-events)
5. [State Hoisting](#5-state-hoisting)
6. [UI State Sealed Class](#6-ui-state-sealed-class)
7. [Screen State](#7-screen-state)
8. [Event System](#8-event-system)
9. [Navigation State](#9-navigation-state)
10. [Form State](#10-form-state)
11. [Search State](#11-search-state)
12. [Pagination State](#12-pagination-state)
13. [Optimistic Updates](#13-optimistic-updates)
14. [Cache Strategy](#14-cache-strategy)
15. [Reactive Data Flow](#15-reactive-data-flow)
16. [Dependency Injection](#16-dependency-injection)
17. [Hilt Modules](#17-hilt-modules)
18. [Hilt Scoping](#18-hilt-scoping)
19. [Hilt Testing](#19-hilt-testing)
20. [Coroutine Scope Management](#20-coroutine-scope-management)
21. [Flow Collection in Compose](#21-flow-collection-in-compose)
22. [State Preservation](#22-state-preservation)
23. [Configuration Change Handling](#23-configuration-change-handling)
24. [Process Death Handling](#24-process-death-handling)
25. [Memory Management](#25-memory-management)
26. [State Debugging](#26-state-debugging)

---

## 1. MVVM Architecture

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        MVVM Architecture                     │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                     View Layer                        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │  │
│  │  │ Compose  │  │ Compose  │  │ Compose  │           │  │
│  │  │ Screen A │  │ Screen B │  │ Screen C │           │  │
│  │  └────┬─────┘  └────┬─────┘  └────┬─────┘           │  │
│  │       │              │              │                  │  │
│  │       │    observes  │              │                  │  │
│  │       │   StateFlow  │              │                  │  │
│  │       │              │              │                  │  │
│  └───────┼──────────────┼──────────────┼─────────────────┘  │
│          │              │              │                     │
│  ┌───────▼──────┐ ┌─────▼──────┐ ┌────▼───────┐           │
│  │  ViewModel A  │ │ ViewModel B│ │ ViewModel C│           │
│  │               │ │            │ │            │           │
│  │ - uiState:    │ │ - uiState: │ │ - uiState: │           │
│  │   StateFlow   │ │   StateFlow│ │   StateFlow│           │
│  │ - events:     │ │ - events:  │ │ - events:  │           │
│  │   SharedFlow  │ │  SharedFlow│ │  SharedFlow│           │
│  │               │ │            │ │            │           │
│  │ + onAction()  │ │ + onAction │ │ + onAction │           │
│  └───────┬──────┘ └─────┬──────┘ └────┬───────┘           │
│          │              │              │                     │
├──────────┼──────────────┼──────────────┼─────────────────────┤
│          │              │              │                     │
│  ┌───────▼──────────────▼──────────────▼─────────────────┐  │
│  │                   Domain Layer                         │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │  │
│  │  │  UseCase A   │  │  UseCase B   │  │  UseCase C  │ │  │
│  │  │              │  │              │  │             │ │  │
│  │  │ + execute()  │  │ + execute()  │  │ + execute() │ │  │
│  │  │ + observe()  │  │ + observe()  │  │ + observe() │ │  │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘ │  │
│  └─────────┼─────────────────┼─────────────────┼─────────┘  │
│            │                 │                 │             │
├────────────┼─────────────────┼─────────────────┼─────────────┤
│            │                 │                 │             │
│  ┌─────────▼─────────────────▼─────────────────▼─────────┐  │
│  │                  Data Layer                            │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │  │
│  │  │ Repository A │  │ Repository B │  │Repository C  │ │  │
│  │  │              │  │              │  │             │ │  │
│  │  │ - remote API │  │ - remote API │  │ - remote    │ │  │
│  │  │ - local DB   │  │ - local DB   │  │ - local DB  │ │  │
│  │  │ - cache      │  │ - cache      │  │ - cache     │ │  │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘ │  │
│  └─────────┼─────────────────┼─────────────────┼─────────┘  │
│            │                 │                 │             │
├────────────┼─────────────────┼─────────────────┼─────────────┤
│            ▼                 ▼                 ▼             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                  Data Sources                         │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │   │
│  │  │  Room    │  │  Retrofit│  │ DataStore│          │   │
│  │  │  (Local) │  │  (Remote)│  │(Prefs)   │          │   │
│  │  └──────────┘  └──────────┘  └──────────┘          │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow Direction

```
┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐
│  User    │      │   View   │      │ ViewModel│      │Repository│
│ Action   │      │  (Compose)│      │          │      │          │
└────┬─────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘
     │                 │                 │                  │
     │  onClick/onInput│                 │                  │
     │────────────────▶│                 │                  │
     │                 │                 │                  │
     │                 │  call viewModel │                  │
     │                 │  .onAction()    │                  │
     │                 │────────────────▶│                  │
     │                 │                 │                  │
     │                 │                 │  fetch/update    │
     │                 │                 │─────────────────▶│
     │                 │                 │                  │
     │                 │                 │   Flow<Result>   │
     │                 │                 │◀─────────────────│
     │                 │                 │                  │
     │                 │  StateFlow      │                  │
     │                 │  emits new state│                  │
     │                 │◀────────────────│                  │
     │                 │                 │                  │
     │  Recomposition  │                 │                  │
     │◀────────────────│                 │                  │
     │                 │                 │                  │
```

---

## 2. ViewModel Lifecycle

### ViewModel Lifecycle Management

```kotlin
@HiltViewModel
class ChatViewModel @Inject constructor(
    private val getConversationsUseCase: GetConversationsUseCase,
    private val sendMessageUseCase: SendMessageUseCase,
    private val syncManager: SyncManager,
    private val savedStateHandle: SavedStateHandle
) : ViewModel() {

    // ─── State ─────────────────────────────────────────────

    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    private val _events = MutableSharedFlow<ChatEvent>()
    val events: SharedFlow<ChatEvent> = _events.asSharedFlow()

    // ─── Initialization ────────────────────────────────────

    init {
        loadConversations()
        observeSyncState()
    }

    // ─── Data Loading ──────────────────────────────────────

    private fun loadConversations() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            getConversationsUseCase()
                .catch { e ->
                    _uiState.update { it.copy(isLoading = false, error = e.message) }
                    _events.emit(ChatEvent.ShowError(e.message ?: "Unknown error"))
                }
                .collect { conversations ->
                    _uiState.update {
                        it.copy(
                            conversations = conversations,
                            isLoading = false,
                            error = null
                        )
                    }
                }
        }
    }

    private fun observeSyncState() {
        viewModelScope.launch {
            syncManager.syncState.collect { state ->
                _uiState.update { it.copy(syncState = state) }
            }
        }
    }

    // ─── Actions ───────────────────────────────────────────

    fun onAction(action: ChatAction) {
        when (action) {
            is ChatAction.SendMessage -> sendMessage(action.content)
            is ChatAction.SelectConversation -> selectConversation(action.id)
            is ChatAction.DeleteConversation -> deleteConversation(action.id)
            is ChatAction.Refresh -> refreshConversations()
            is ChatAction.ClearError -> clearError()
        }
    }

    private fun sendMessage(content: String) {
        viewModelScope.launch {
            val conversationId = _uiState.value.selectedConversationId ?: return@launch

            _uiState.update { it.copy(isSendingMessage = true) }

            try {
                val result = sendMessageUseCase(
                    conversationId = conversationId,
                    content = content
                )

                result.fold(
                    onSuccess = { message ->
                        _events.emit(ChatEvent.MessageSent(message))
                        _uiState.update { it.copy(isSendingMessage = false) }
                    },
                    onFailure = { error ->
                        _events.emit(ChatEvent.ShowError(error.message ?: "Failed to send"))
                        _uiState.update { it.copy(isSendingMessage = false) }
                    }
                )
            } catch (e: Exception) {
                _events.emit(ChatEvent.ShowError(e.message ?: "Unknown error"))
                _uiState.update { it.copy(isSendingMessage = false) }
            }
        }
    }

    private fun selectConversation(id: String) {
        _uiState.update { it.copy(selectedConversationId = id) }
    }

    private fun deleteConversation(id: String) {
        viewModelScope.launch {
            try {
                // Perform deletion
                _events.emit(ChatEvent.ConversationDeleted(id))
                if (_uiState.value.selectedConversationId == id) {
                    _uiState.update { it.copy(selectedConversationId = null) }
                }
            } catch (e: Exception) {
                _events.emit(ChatEvent.ShowError(e.message ?: "Failed to delete"))
            }
        }
    }

    private fun refreshConversations() {
        viewModelScope.launch {
            _uiState.update { it.copy(isRefreshing = true) }
            loadConversations()
            _uiState.update { it.copy(isRefreshing = false) }
        }
    }

    private fun clearError() {
        _uiState.update { it.copy(error = null) }
    }

    // ─── Cleanup ───────────────────────────────────────────

    override fun onCleared() {
        super.onCleared()
        // Cleanup resources if needed
    }
}

// ─── UI State ──────────────────────────────────────────────

data class ChatUiState(
    val conversations: List<Conversation> = emptyList(),
    val selectedConversationId: String? = null,
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val isSendingMessage: Boolean = false,
    val error: String? = null,
    val syncState: SyncState = SyncState.Idle
)

// ─── Actions ───────────────────────────────────────────────

sealed class ChatAction {
    data class SendMessage(val content: String) : ChatAction()
    data class SelectConversation(val id: String) : ChatAction()
    data class DeleteConversation(val id: String) : ChatAction()
    data object Refresh : ChatAction()
    data object ClearError : ChatAction()
}

// ─── Events ────────────────────────────────────────────────

sealed class ChatEvent {
    data class MessageSent(val message: Message) : ChatEvent()
    data class ConversationDeleted(val id: String) : ChatEvent()
    data class ShowError(val message: String) : ChatEvent()
    data object NavigateToLogin : ChatEvent()
}
```

---

## 3. StateFlow for UI State

### StateFlow Usage Patterns

```kotlin
// ─── Pattern 1: Simple State ───────────────────────────────

class ProfileViewModel @Inject constructor(
    private val getUserProfileUseCase: GetUserProfileUseCase
) : ViewModel() {

    data class UiState(
        val profile: UserProfile? = null,
        val isLoading: Boolean = true,
        val error: String? = null
    )

    private val _uiState = MutableStateFlow(UiState())
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    init {
        viewModelScope.launch {
            getUserProfileUseCase()
                .onStart { _uiState.update { it.copy(isLoading = true) } }
                .catch { e -> _uiState.update { it.copy(error = e.message, isLoading = false) } }
                .collect { profile ->
                    _uiState.update { it.copy(profile = profile, isLoading = false) }
                }
        }
    }
}

// ─── Pattern 2: Derived State ──────────────────────────────

class DashboardViewModel @Inject constructor() : ViewModel() {

    data class UiState(
        val conversations: List<Conversation> = emptyList(),
        val documents: List<Document> = emptyList(),
        val agents: List<Agent> = emptyList(),
        val isLoading: Boolean = false,
        val searchQuery: String = "",
        val selectedTab: Int = 0
    )

    private val _uiState = MutableStateFlow(UiState())
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    // Derived state: filtered conversations based on search query
    val filteredConversations: StateFlow<List<Conversation>> =
        combine(
            _uiState.map { it.conversations },
            _uiState.map { it.searchQuery }
        ) { conversations, query ->
            if (query.isBlank()) conversations
            else conversations.filter {
                it.title.contains(query, ignoreCase = true)
            }
        }.stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = emptyList()
        )

    // Derived state: has data
    val hasData: StateFlow<Boolean> = _uiState
        .map { it.conversations.isNotEmpty() || it.documents.isNotEmpty() }
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = false
        )
}

// ─── Pattern 3: Multiple Flows ─────────────────────────────

class SettingsViewModel @Inject constructor(
    private val userSettingsRepository: UserSettingsRepository
) : ViewModel() {

    data class UiState(
        val settings: UserSettings = UserSettings(),
        val isSaving: Boolean = false,
        val saveSuccess: Boolean = false
    )

    private val _uiState = MutableStateFlow(UiState())
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    init {
        viewModelScope.launch {
            userSettingsRepository.settings.collect { settings ->
                _uiState.update { it.copy(settings = settings) }
            }
        }
    }

    fun updateThemeMode(mode: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true) }
            userSettingsRepository.updateThemeMode(mode)
            _uiState.update { it.copy(isSaving = false, saveSuccess = true) }
        }
    }
}

// ─── Pattern 4: State with Side Effects ────────────────────

class FormViewModel @Inject constructor() : ViewModel() {

    data class UiState(
        val name: String = "",
        val email: String = "",
        val isValid: Boolean = false,
        val isSubmitting: Boolean = false
    )

    private val _uiState = MutableStateFlow(UiState())
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    private val _sideEffect = Channel<SideEffect>(Channel.BUFFERED)
    val sideEffect: Flow<SideEffect> = _sideEffect.receiveAsFlow()

    fun updateName(name: String) {
        _uiState.update { it.copy(name = name) }
        validateForm()
    }

    fun updateEmail(email: String) {
        _uiState.update { it.copy(email = email) }
        validateForm()
    }

    private fun validateForm() {
        val state = _uiState.value
        val isValid = state.name.isNotBlank() &&
            state.email.contains("@") &&
            state.email.contains(".")
        _uiState.update { it.copy(isValid = isValid) }
    }

    fun submit() {
        viewModelScope.launch {
            _uiState.update { it.copy(isSubmitting = true) }
            try {
                // Submit logic
                _sideEffect.send(SideEffect.ShowToast("Submitted successfully"))
                _sideEffect.send(SideEffect.NavigateBack)
            } catch (e: Exception) {
                _sideEffect.send(SideEffect.ShowError(e.message ?: "Error"))
            } finally {
                _uiState.update { it.copy(isSubmitting = false) }
            }
        }
    }

    sealed class SideEffect {
        data class ShowToast(val message: String) : SideEffect()
        data class ShowError(val message: String) : SideEffect()
        data object NavigateBack : SideEffect()
    }
}
```

### StateFlow vs SharedFlow Decision Matrix

```
┌──────────────────────────────────────────────────────────┐
│           StateFlow vs SharedFlow Decision                 │
│                                                          │
│  ┌─────────────────────┬──────────────────────────────┐  │
│  │ Use StateFlow when: │ Use SharedFlow when:         │  │
│  ├─────────────────────┼──────────────────────────────┤  │
│  │ • UI state          │ • One-time events            │  │
│  │ • Needs current     │ • Navigation                 │  │
│  │   value             │ • Toasts/Snackbars           │  │
│  │ • Data binding      │ • Dialogs                    │  │
│  │ • Screen state      │ • Side effects               │  │
│  │ • Form fields       │ • Analytics events           │  │
│  │ • List data         │ • Error emissions            │  │
│  └─────────────────────┴──────────────────────────────┘  │
│                                                          │
│  Key Differences:                                        │
│  ┌──────────────┬──────────────┬─────────────────────┐  │
│  │ Property     │ StateFlow    │ SharedFlow          │  │
│  ├──────────────┼──────────────┼─────────────────────┤  │
│  │ Replay       │ Always 1     │ Configurable (0+)   │  │
│  │ Initial val  │ Required     │ Not required        │  │
│  │ Conflation   │ Yes          │ Configurable        │  │
│  │ Distinct     │ Yes (equals) │ No                  │  │
│  │ Subscription │ Only latest  │ All emissions       │  │
│  └──────────────┴──────────────┴─────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

---

## 4. SharedFlow for Events

### Event System Implementation

```kotlin
class EventViewModel @Inject constructor() : ViewModel() {

    // ─── UI Events (one-time) ───────────────────────────────

    private val _uiEvents = MutableSharedFlow<UiEvent>(
        replay = 0,
        extraBufferCapacity = 64,
        onBufferOverflow = BufferOverflow.DROP_OLDEST
    )
    val uiEvents: SharedFlow<UiEvent> = _uiEvents.asSharedFlow()

    // ─── Navigation Events ──────────────────────────────────

    private val _navigationEvents = Channel<NavigationEvent>(
        capacity = Channel.BUFFERED,
        onBufferOverflow = BufferOverflow.DROP_OLDEST
    )
    val navigationEvents: Flow<NavigationEvent> = _navigationEvents.receiveAsFlow()

    // ─── Emit Events ────────────────────────────────────────

    fun showError(message: String) {
        viewModelScope.launch {
            _uiEvents.emit(UiEvent.ShowSnackbar(message))
        }
    }

    fun navigate(route: String) {
        viewModelScope.launch {
            _navigationEvents.send(NavigationEvent.Navigate(route))
        }
    }

    fun navigateBack() {
        viewModelScope.launch {
            _navigationEvents.send(NavigationEvent.NavigateBack)
        }
    }

    fun showDialog(dialog: DialogConfig) {
        viewModelScope.launch {
            _uiEvents.emit(UiEvent.ShowDialog(dialog))
        }
    }

    // ─── Event Definitions ──────────────────────────────────

    sealed class UiEvent {
        data class ShowSnackbar(
            val message: String,
            val actionLabel: String? = null,
            val duration: SnackbarDuration = SnackbarDuration.Short
        ) : UiEvent()

        data class ShowToast(val message: String) : UiEvent()
        data class ShowDialog(val config: DialogConfig) : UiEvent()
        data object HideDialog : UiEvent()
        data class CopyToClipboard(val text: String, val label: String) : UiEvent()
        data class Vibrate(val durationMs: Long = 100) : UiEvent()
    }

    sealed class NavigationEvent {
        data class Navigate(val route: String) : NavigationEvent()
        data object NavigateBack : NavigationEvent()
        data class NavigateAndPopUpTo(
            val route: String,
            val popUpTo: String
        ) : NavigationEvent()
    }
}

data class DialogConfig(
    val title: String,
    val message: String,
    val positiveButton: String = "OK",
    val negativeButton: String? = null,
    val onPositive: () -> Unit = {},
    val onNegative: (() -> Unit)? = null
)
```

### Event Collection in Composable

```kotlin
@Composable
fun EventHandlingScreen(
    viewModel: EventViewModel = hiltViewModel()
) {
    val snackbarHostState = remember { SnackbarHostState() }

    // Collect one-time events
    LaunchedEffect(Unit) {
        viewModel.uiEvents.collect { event ->
            when (event) {
                is UiEvent.ShowSnackbar -> {
                    snackbarHostState.showSnackbar(
                        message = event.message,
                        actionLabel = event.actionLabel,
                        duration = event.duration
                    )
                }
                is UiEvent.ShowToast -> {
                    Toast.makeText(context, event.message, Toast.LENGTH_SHORT).show()
                }
                is UiEvent.ShowDialog -> {
                    // Show dialog
                }
                is UiEvent.HideDialog -> {
                    // Hide dialog
                }
                is UiEvent.CopyToClipboard -> {
                    clipboardManager.setPrimaryClip(
                        ClipData.newPlainText(event.label, event.text)
                    )
                }
                is UiEvent.Vibrate -> {
                    // Vibrate
                }
            }
        }
    }

    // Collect navigation events
    LaunchedEffect(Unit) {
        viewModel.navigationEvents.collect { event ->
            when (event) {
                is NavigationEvent.Navigate -> {
                    navController.navigate(event.route)
                }
                is NavigationEvent.NavigateBack -> {
                    navController.popBackStack()
                }
                is NavigationEvent.NavigateAndPopUpTo -> {
                    navController.navigate(event.route) {
                        popUpTo(event.popUpTo) { inclusive = true }
                    }
                }
            }
        }
    }

    // UI content
    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { padding ->
        // Screen content
    }
}
```

---

## 5. State Hoisting

### State Hoisting Patterns

```kotlin
// ─── Stateless Component ───────────────────────────────────

@Composable
fun SearchBar(
    query: String,
    onQueryChange: (String) -> Unit,
    onSearch: () -> Unit,
    onClear: () -> Unit,
    modifier: Modifier = Modifier
) {
    OutlinedTextField(
        value = query,
        onValueChange = onQueryChange,
        placeholder = { Text("Search...") },
        leadingIcon = {
            Icon(Icons.Filled.Search, contentDescription = "Search")
        },
        trailingIcon = {
            if (query.isNotEmpty()) {
                IconButton(onClick = onClear) {
                    Icon(Icons.Filled.Clear, contentDescription = "Clear")
                }
            }
        },
        keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
        keyboardActions = KeyboardActions(onSearch = { onSearch() }),
        singleLine = true,
        modifier = modifier.fillMaxWidth()
    )
}

// ─── Stateful Container ────────────────────────────────────

@Composable
fun SearchBarContainer(
    onSearch: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    var query by remember { mutableStateOf("") }

    SearchBar(
        query = query,
        onQueryChange = { query = it },
        onSearch = { onSearch(query) },
        onClear = { query = "" },
        modifier = modifier
    )
}

// ─── Complex Stateful Container ────────────────────────────

@Composable
fun ChatInputContainer(
    onSendMessage: (String, List<Attachment>) -> Unit,
    modifier: Modifier = Modifier
) {
    var text by remember { mutableStateOf("") }
    var attachments by remember { mutableStateOf(emptyList<Attachment>()) }
    var isUploading by remember { mutableStateOf(false) }

    val canSend = text.isNotBlank() || attachments.isNotEmpty()

    ChatInputBar(
        text = text,
        onTextChange = { text = it },
        attachments = attachments,
        onRemoveAttachment = { attachments = attachments - it },
        isUploading = isUploading,
        canSend = canSend,
        onSend = {
            onSendMessage(text, attachments)
            text = ""
            attachments = emptyList()
        },
        onAttach = { /* Open file picker */ },
        modifier = modifier
    )
}

// ─── List Item State ───────────────────────────────────────

@Composable
fun ConversationItem(
    conversation: Conversation,
    isSelected: Boolean,
    onClick: () -> Unit,
    onLongClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val backgroundColor by animateColorAsState(
        targetValue = if (isSelected) {
            MaterialTheme.colorScheme.primaryContainer
        } else {
            MaterialTheme.colorScheme.surface
        },
        label = "bgColor"
    )

    Card(
        modifier = modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .combinedClickable(
                onClick = onClick,
                onLongClick = onLongClick
            ),
        colors = CardDefaults.cardColors(containerColor = backgroundColor)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = conversation.title,
                    style = MaterialTheme.typography.titleMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                if (conversation.lastMessage != null) {
                    Text(
                        text = conversation.lastMessage,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis
                    )
                }
            }

            if (conversation.unreadCount > 0) {
                Badge {
                    Text(text = conversation.unreadCount.toString())
                }
            }
        }
    }
}
```

### State Hoisting Decision Flow

```
┌─────────────────────────────────────────────────────────┐
│              State Hoisting Decision                      │
│                                                         │
│  Does the component need state?                         │
│       │                                                 │
│    Yes│  No                                             │
│       │  │                                              │
│       ▼  ▼                                              │
│  ┌──────────┐  ┌─────────────────┐                     │
│  │ Who owns │  │ Stateless       │                     │
│  │ the state?│  │ component       │                     │
│  └────┬─────┘  └─────────────────┘                     │
│       │                                                 │
│    ┌──┼──────────┬──────────────┐                       │
│    ▼  │          ▼              ▼                       │
│  ┌────┴───┐  ┌──────────┐  ┌──────────┐               │
│  │ Only   │  │ Multiple │  │ Complex  │               │
│  │ this   │  │ siblings │  │ logic    │               │
│  │ comp.  │  │ need it  │  │ needed   │               │
│  └────┬───┘  └────┬─────┘  └────┬─────┘               │
│       │           │              │                      │
│       ▼           ▼              ▼                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │ Keep in  │  │ Hoist to │  │ Use      │             │
│  │ composable│  │ common  │  │ ViewModel│             │
│  │ (local   │  │ parent  │  │          │             │
│  │ state)   │  │          │  │          │             │
│  └──────────┘  └──────────┘  └──────────┘             │
└─────────────────────────────────────────────────────────┘
```

---

## 6. UI State Sealed Class

### Comprehensive UI State Pattern

```kotlin
// ─── Generic UI State ──────────────────────────────────────

sealed class UiState<out T> {
    data object Idle : UiState<Nothing>()
    data object Loading : UiState<Nothing>()
    data class Success<T>(val data: T) : UiState<T>()
    data class Error(
        val message: String,
        val throwable: Throwable? = null,
        val retryAction: (() -> Unit)? = null
    ) : UiState<Nothing>()

    val isLoading: Boolean get() = this is Loading
    val isSuccess: Boolean get() = this is Success
    val isError: Boolean get() = this is Error
    val data: T? get() = (this as? Success)?.data

    fun <R> map(transform: (T) -> R): UiState<R> {
        return when (this) {
            is Idle -> Idle
            is Loading -> Loading
            is Success -> Success(transform(data))
            is Error -> this
        }
    }

    fun getOrNull(): T? = data

    fun getOrDefault(default: @UnsafeVariance T): T = data ?: default

    inline fun onSuccess(action: (T) -> Unit): UiState<T> {
        if (this is Success) action(data)
        return this
    }

    inline fun onError(action: (String) -> Unit): UiState<T> {
        if (this is Error) action(message)
        return this
    }
}

// ─── Screen-Specific States ────────────────────────────────

data class ChatListUiState(
    val conversations: List<Conversation> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val error: String? = null,
    val searchQuery: String = "",
    val isSearchActive: Boolean = false,
    val selectedConversations: Set<String> = emptySet(),
    val isMultiSelectMode: Boolean = false,
    val sortBy: SortBy = SortBy.LAST_MESSAGE,
    val filterBy: FilterBy = FilterBy.ALL
) {
    val filteredConversations: List<Conversation>
        get() = conversations
            .filter { matchesFilter(it) }
            .filter {
                searchQuery.isBlank() || it.title.contains(searchQuery, ignoreCase = true)
            }
            .sortedWith(sortComparator)

    val hasData: Boolean get() = conversations.isNotEmpty()
    val isEmpty: Boolean get() = !isLoading && conversations.isEmpty()
    val hasSelection: Boolean get() = selectedConversations.isNotEmpty()

    private fun matchesFilter(conversation: Conversation): Boolean = when (filterBy) {
        FilterBy.ALL -> true
        FilterBy.PINNED -> conversation.isPinned
        FilterBy.UNREAD -> conversation.unreadCount > 0
        FilterBy.ARCHIVED -> conversation.isArchived
    }

    private val sortComparator: Comparator<Conversation>
        get() = when (sortBy) {
            SortBy.LAST_MESSAGE -> compareByDescending<Conversation> { it.lastMessageTime }
            SortBy.NAME -> compareBy { it.title }
            SortBy.CREATED -> compareByDescending<Conversation> { it.createdAt }
            SortBy.MESSAGE_COUNT -> compareByDescending<Conversation> { it.messageCount }
        }

    enum class SortBy { LAST_MESSAGE, NAME, CREATED, MESSAGE_COUNT }
    enum class FilterBy { ALL, PINNED, UNREAD, ARCHIVED }
}

data class ConversationDetailUiState(
    val conversation: Conversation? = null,
    val messages: List<Message> = emptyList(),
    val isLoading: Boolean = false,
    val isSendingMessage: Boolean = false,
    val error: String? = null,
    val inputText: String = "",
    val attachments: List<Attachment> = emptyList(),
    val replyingTo: Message? = null,
    val isStreaming: Boolean = false,
    val streamingContent: String = ""
) {
    val canSend: Boolean
        get() = inputText.isNotBlank() || attachments.isNotEmpty()

    val hasMessages: Boolean
        get() = messages.isNotEmpty()
}

data class SettingsUiState(
    val userSettings: UserSettings = UserSettings(),
    val themeMode: ThemeMode = ThemeMode.SYSTEM,
    val isDarkMode: Boolean = false,
    val fontSize: FontSize = FontSize.MEDIUM,
    val language: String = "en",
    val notificationsEnabled: Boolean = true,
    val biometricEnabled: Boolean = false,
    val offlineModeEnabled: Boolean = false,
    val storageUsedBytes: Long = 0,
    val storageMaxBytes: Long = 500 * 1024 * 1024,
    val appVersion: String = "",
    val isSaving: Boolean = false,
    val error: String? = null
) {
    val storageUsedPercentage: Float
        get() = if (storageMaxBytes > 0) {
            storageUsedBytes.toFloat() / storageMaxBytes
        } else 0f

    val storageUsedFormatted: String
        get() = formatBytes(storageUsedBytes)

    val storageMaxFormatted: String
        get() = formatBytes(storageMaxBytes)

    private fun formatBytes(bytes: Long): String {
        val kb = bytes / 1024.0
        val mb = kb / 1024.0
        val gb = mb / 1024.0
        return when {
            gb >= 1.0 -> "%.1f GB".format(gb)
            mb >= 1.0 -> "%.1f MB".format(mb)
            else -> "%.0f KB".format(kb)
        }
    }
}

enum class ThemeMode { LIGHT, DARK, SYSTEM }
enum class FontSize { SMALL, MEDIUM, LARGE, EXTRA_LARGE }
```

---

## 7. Screen State

### Complete Screen State Pattern

```kotlin
data class DashboardScreenState(
    // ─── Data ──────────────────────────────────────────────
    val recentConversations: List<Conversation> = emptyList(),
    val recentDocuments: List<Document> = emptyList(),
    val activeAgents: List<Agent> = emptyList(),
    val stats: DashboardStats = DashboardStats(),

    // ─── Loading States ────────────────────────────────────
    val isLoadingConversations: Boolean = false,
    val isLoadingDocuments: Boolean = false,
    val isLoadingAgents: Boolean = false,
    val isLoadingStats: Boolean = false,

    // ─── Error States ──────────────────────────────────────
    val conversationsError: String? = null,
    val documentsError: String? = null,
    val agentsError: String? = null,
    val statsError: String? = null,

    // ─── UI State ──────────────────────────────────────────
    val selectedTab: Int = 0,
    val searchQuery: String = "",
    val isSearchActive: Boolean = false,
    val isRefreshing: Boolean = false,
    val showOnboarding: Boolean = false,
    val showUpdateBanner: Boolean = false,

    // ─── Sync State ────────────────────────────────────────
    val syncState: SyncState = SyncState.Idle,
    val pendingSyncCount: Int = 0,
    val lastSyncTime: Long = 0
) {
    val isLoading: Boolean
        get() = isLoadingConversations || isLoadingDocuments ||
            isLoadingAgents || isLoadingStats

    val hasError: Boolean
        get() = conversationsError != null || documentsError != null ||
            agentsError != null || statsError != null

    val errorMessages: List<String>
        get() = listOfNotNull(
            conversationsError,
            documentsError,
            agentsError,
            statsError
        )

    val isOnline: Boolean
        get() = syncState != SyncState.Offline

    val needsRefresh: Boolean
        get() = System.currentTimeMillis() - lastSyncTime > 30 * 60 * 1000
}

data class DashboardStats(
    val totalConversations: Int = 0,
    val totalDocuments: Int = 0,
    val totalMessages: Int = 0,
    val storageUsed: Long = 0,
    val lastActiveAgent: String? = null
)

// ─── ViewModel using screen state ──────────────────────────

@HiltViewModel
class DashboardViewModel @Inject constructor(
    private val getConversationsUseCase: GetConversationsUseCase,
    private val getDocumentsUseCase: GetDocumentsUseCase,
    private val getAgentsUseCase: GetAgentsUseCase,
    private val getStatsUseCase: GetDashboardStatsUseCase,
    private val syncManager: SyncManager
) : ViewModel() {

    private val _state = MutableStateFlow(DashboardScreenState())
    val state: StateFlow<DashboardScreenState> = _state.asStateFlow()

    init {
        loadAllData()
        observeSyncState()
    }

    fun onAction(action: DashboardAction) {
        when (action) {
            is DashboardAction.Refresh -> refreshAll()
            is DashboardAction.SelectTab -> selectTab(action.index)
            is DashboardAction.Search -> search(action.query)
            is DashboardAction.ClearSearch -> clearSearch()
            is DashboardAction.DismissError -> clearError(action.errorType)
        }
    }

    private fun loadAllData() {
        loadConversations()
        loadDocuments()
        loadAgents()
        loadStats()
    }

    private fun loadConversations() {
        viewModelScope.launch {
            _state.update { it.copy(isLoadingConversations = true, conversationsError = null) }
            getConversationsUseCase()
                .catch { e -> _state.update { s -> s.copy(conversationsError = e.message, isLoadingConversations = false) } }
                .collect { data -> _state.update { s -> s.copy(recentConversations = data.take(5), isLoadingConversations = false) } }
        }
    }

    private fun loadDocuments() {
        viewModelScope.launch {
            _state.update { it.copy(isLoadingDocuments = true, documentsError = null) }
            getDocumentsUseCase()
                .catch { e -> _state.update { s -> s.copy(documentsError = e.message, isLoadingDocuments = false) } }
                .collect { data -> _state.update { s -> s.copy(recentDocuments = data.take(5), isLoadingDocuments = false) } }
        }
    }

    private fun loadAgents() {
        viewModelScope.launch {
            _state.update { it.copy(isLoadingAgents = true, agentsError = null) }
            getAgentsUseCase()
                .catch { e -> _state.update { s -> s.copy(agentsError = e.message, isLoadingAgents = false) } }
                .collect { data -> _state.update { s -> s.copy(activeAgents = data.filter { it.isActive }, isLoadingAgents = false) } }
        }
    }

    private fun loadStats() {
        viewModelScope.launch {
            _state.update { it.copy(isLoadingStats = true, statsError = null) }
            getStatsUseCase()
                .catch { e -> _state.update { s -> s.copy(statsError = e.message, isLoadingStats = false) } }
                .collect { data -> _state.update { s -> s.copy(stats = data, isLoadingStats = false) } }
        }
    }

    private fun refreshAll() {
        _state.update { it.copy(isRefreshing = true) }
        loadAllData()
        viewModelScope.launch {
            delay(1000) // Minimum refresh animation time
            _state.update { it.copy(isRefreshing = false) }
        }
    }

    private fun selectTab(index: Int) {
        _state.update { it.copy(selectedTab = index) }
    }

    private fun search(query: String) {
        _state.update { it.copy(searchQuery = query, isSearchActive = query.isNotBlank()) }
    }

    private fun clearSearch() {
        _state.update { it.copy(searchQuery = "", isSearchActive = false) }
    }

    private fun clearError(errorType: ErrorType) {
        _state.update { state ->
            when (errorType) {
                ErrorType.CONVERSATIONS -> state.copy(conversationsError = null)
                ErrorType.DOCUMENTS -> state.copy(documentsError = null)
                ErrorType.AGENTS -> state.copy(agentsError = null)
                ErrorType.STATS -> state.copy(statsError = null)
            }
        }
    }

    private fun observeSyncState() {
        viewModelScope.launch {
            syncManager.syncState.collect { syncState ->
                _state.update { it.copy(syncState = syncState) }
            }
        }
    }
}

sealed class DashboardAction {
    data object Refresh : DashboardAction()
    data class SelectTab(val index: Int) : DashboardAction()
    data class Search(val query: String) : DashboardAction()
    data object ClearSearch : DashboardAction()
    data class DismissError(val errorType: ErrorType) : DashboardAction()
}

enum class ErrorType { CONVERSATIONS, DOCUMENTS, AGENTS, STATS }
```

---

## 8. Event System

### Comprehensive Event Architecture

```kotlin
// ─── Event Definitions ─────────────────────────────────────

sealed class AppEvent {
    // Navigation
    data class Navigate(val route: String, val popUpTo: String? = null) : AppEvent()
    data object NavigateBack : AppEvent()
    data object NavigateToLogin : AppEvent()

    // UI Feedback
    data class ShowSnackbar(
        val message: String,
        val action: SnackBarAction? = null,
        val duration: SnackbarDuration = SnackbarDuration.Short
    ) : AppEvent()

    data class ShowToast(val message: String) : AppEvent()
    data class ShowDialog(val config: DialogConfig) : AppEvent()
    data object DismissDialog : AppEvent()

    // System
    data object RequestPermission : AppEvent()
    data class OpenUrl(val url: String) : AppEvent()
    data class ShareText(val text: String, val title: String) : AppEvent()
    data class CopyToClipboard(val text: String) : AppEvent()

    // Auth
    data object SessionExpired : AppEvent()
    data object Logout : AppEvent()
}

data class SnackBarAction(
    val label: String,
    val action: suspend () -> Unit
)

// ─── Event Emitter ─────────────────────────────────────────

class EventProducer @Inject constructor() {
    private val _events = MutableSharedFlow<AppEvent>(
        replay = 0,
        extraBufferCapacity = 64,
        onBufferOverflow = BufferOverflow.DROP_OLDEST
    )
    val events: SharedFlow<AppEvent> = _events.asSharedFlow()

    suspend fun emit(event: AppEvent) {
        _events.emit(event)
    }

    fun tryEmit(event: AppEvent) {
        _events.tryEmit(event)
    }
}

// ─── Event Consumer ────────────────────────────────────────

@Composable
fun <T> EventConsumer(
    eventFlow: Flow<T>,
    onEvent: (T) -> Unit
) {
    LaunchedEffect(Unit) {
        eventFlow.collect { event ->
            onEvent(event)
        }
    }
}

// ─── Usage ─────────────────────────────────────────────────

@Composable
fun MainScreen(
    viewModel: MainViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }

    // Collect events
    LaunchedEffect(Unit) {
        viewModel.events.collect { event ->
            when (event) {
                is AppEvent.ShowSnackbar -> {
                    val result = snackbarHostState.showSnackbar(
                        message = event.message,
                        actionLabel = event.action?.label,
                        duration = event.duration
                    )
                    if (result == SnackbarResult.ActionPerformed) {
                        event.action?.action?.invoke()
                    }
                }
                is AppEvent.ShowToast -> {
                    Toast.makeText(context, event.message, Toast.LENGTH_SHORT).show()
                }
                is AppEvent.Navigate -> {
                    if (event.popUpTo != null) {
                        navController.navigate(event.route) {
                            popUpTo(event.popUpTo) { inclusive = true }
                        }
                    } else {
                        navController.navigate(event.route)
                    }
                }
                is AppEvent.NavigateBack -> navController.popBackStack()
                is AppEvent.NavigateToLogin -> {
                    navController.navigate("login") {
                        popUpTo(0) { inclusive = true }
                    }
                }
                is AppEvent.ShowDialog -> { /* Show dialog */ }
                is AppEvent.DismissDialog -> { /* Dismiss dialog */ }
                is AppEvent.Logout -> { /* Handle logout */ }
                else -> {}
            }
        }
    }

    Scaffold(snackbarHost = { SnackbarHost(snackbarHostState) }) { padding ->
        // Screen content
    }
}
```

---

## 9. Navigation State

### Navigation State Management

```kotlin
data class NavigationState(
    val currentRoute: String = "dashboard",
    val previousRoute: String? = null,
    val backStack: List<String> = listOf("dashboard"),
    val isNavigating: Boolean = false,
    val arguments: Map<String, Any> = emptyMap()
) {
    val canGoBack: Boolean get() = backStack.size > 1
    val isAuthenticated: Boolean get() = currentRoute != "login" && currentRoute != "splash"

    fun navigateTo(route: String): NavigationState {
        return copy(
            previousRoute = currentRoute,
            currentRoute = route,
            backStack = backStack + route,
            isNavigating = false
        )
    }

    fun navigateBack(): NavigationState {
        if (!canGoBack) return this
        val newStack = backStack.dropLast(1)
        return copy(
            previousRoute = currentRoute,
            currentRoute = newStack.lastOrNull() ?: "dashboard",
            backStack = newStack,
            isNavigating = false
        )
    }

    fun navigateAndClear(route: String): NavigationState {
        return copy(
            previousRoute = currentRoute,
            currentRoute = route,
            backStack = listOf(route),
            isNavigating = false
        )
    }
}

@HiltViewModel
class NavigationViewModel @Inject constructor() : ViewModel() {

    private val _state = MutableStateFlow(NavigationState())
    val state: StateFlow<NavigationState> = _state.asStateFlow()

    fun navigate(route: String) {
        _state.update { it.navigateTo(route) }
    }

    fun navigateBack() {
        _state.update { it.navigateBack() }
    }

    fun navigateAndClear(route: String) {
        _state.update { it.navigateAndClear(route) }
    }

    fun setNavigating(isNavigating: Boolean) {
        _state.update { it.copy(isNavigating = isNavigating) }
    }
}
```

---

## 10. Form State

### Form State Management

```kotlin
data class FormState(
    val fields: Map<String, FieldState> = emptyMap(),
    val isSubmitting: Boolean = false,
    val submitError: String? = null,
    val isDirty: Boolean = false,
    val submitCount: Int = 0
) {
    val isValid: Boolean get() = fields.values.all { it.isValid }
    val hasErrors: Boolean get() = fields.values.any { it.error != null }
    val errorCount: Int get() = fields.values.count { it.error != null }

    fun getField(name: String): FieldState = fields[name] ?: FieldState()

    fun getFieldValue(name: String): String = getField(name).value

    fun getError(name: String): String? = getField(name).error

    fun isFieldValid(name: String): Boolean = getField(name).isValid
}

data class FieldState(
    val value: String = "",
    val error: String? = null,
    val isTouched: Boolean = false,
    val isFocused: Boolean = false,
    val isValidators: List<(String) -> String?> = emptyList()
) {
    val isValid: Boolean get() = error == null && (value.isNotEmpty() || !isTouched)
    val hasError: Boolean get() = error != null && isTouched

    fun validate(): FieldState {
        val newError = isValidators.firstNotNullOfOrNull { it(value) }
        return copy(error = newError)
    }
}

// ─── Form ViewModel ────────────────────────────────────────

class LoginFormViewModel @Inject constructor(
    private val loginUseCase: LoginUseCase
) : ViewModel() {

    private val _formState = MutableStateFlow(FormState())
    val formState: StateFlow<FormState> = _formState.asStateFlow()

    private val _events = Channel<FormEvent>(Channel.BUFFERED)
    val events: Flow<FormEvent> = _events.receiveAsFlow()

    init {
        initializeFields()
    }

    private fun initializeFields() {
        _formState.update { state ->
            state.copy(
                fields = mapOf(
                    "email" to FieldState(
                        isValidators = listOf(
                            { v -> if (v.isBlank()) "Email is required" else null },
                            { v -> if (!v.contains("@")) "Invalid email" else null }
                        )
                    ),
                    "password" to FieldState(
                        isValidators = listOf(
                            { v -> if (v.isBlank()) "Password is required" else null },
                            { v -> if (v.length < 8) "Password must be at least 8 characters" else null }
                        )
                    )
                )
            )
        }
    }

    fun updateField(name: String, value: String) {
        _formState.update { state ->
            val field = state.fields[name] ?: return@update state
            val updatedField = field.copy(value = value, isTouched = true).validate()
            state.copy(
                fields = state.fields + (name to updatedField),
                isDirty = true,
                submitError = null
            )
        }
    }

    fun touchField(name: String) {
        _formState.update { state ->
            val field = state.fields[name] ?: return@update state
            state.copy(fields = state.fields + (name to field.copy(isTouched = true)))
        }
    }

    fun focusField(name: String) {
        _formState.update { state ->
            val updatedFields = state.fields.map { (key, field) ->
                key to if (key == name) field.copy(isFocused = true) else field.copy(isFocused = false)
            }.toMap()
            state.copy(fields = updatedFields)
        }
    }

    fun submit() {
        // Validate all fields
        _formState.update { state ->
            val validatedFields = state.fields.map { (name, field) ->
                name to field.copy(isTouched = true).validate()
            }.toMap()
            state.copy(fields = validatedFields)
        }

        if (!_formState.value.isValid) return

        viewModelScope.launch {
            _formState.update { it.copy(isSubmitting = true, submitError = null) }

            try {
                val email = _formState.value.getFieldValue("email")
                val password = _formState.value.getFieldValue("password")

                loginUseCase(email, password)
                    .onSuccess { _events.send(FormEvent.Success) }
                    .onFailure { e ->
                        _formState.update { it.copy(submitError = e.message) }
                        _events.send(FormEvent.Error(e.message ?: "Login failed"))
                    }
            } catch (e: Exception) {
                _formState.update { it.copy(submitError = e.message) }
                _events.send(FormEvent.Error(e.message ?: "Login failed"))
            } finally {
                _formState.update { it.copy(isSubmitting = false, submitCount = it.submitCount + 1) }
            }
        }
    }

    fun reset() {
        _formState.value = FormState()
        initializeFields()
    }

    sealed class FormEvent {
        data object Success : FormEvent()
        data class Error(val message: String) : FormEvent()
    }
}
```

---

## 11. Search State

### Search State Management

```kotlin
data class SearchState(
    val query: String = "",
    val isSearchActive: Boolean = false,
    val isSearching: Boolean = false,
    val results: SearchResults = SearchResults(),
    val recentSearches: List<String> = emptyList(),
    val suggestions: List<String> = emptyList(),
    val filters: SearchFilters = SearchFilters(),
    val error: String? = null,
    val searchHistory: List<SearchHistoryItem> = emptyList()
) {
    val hasResults: Boolean get() = results.totalCount > 0
    val hasQuery: Boolean get() = query.isNotBlank()
    val canSearch: Boolean get() = query.isNotBlank() && !isSearching
    val isEmpty: Boolean get() = !isSearching && hasQuery && !hasResults
}

data class SearchResults(
    val conversations: List<Conversation> = emptyList(),
    val messages: List<Message> = emptyList(),
    val documents: List<Document> = emptyList(),
    val agents: List<Agent> = emptyList()
) {
    val totalCount: Int get() = conversations.size + messages.size + documents.size + agents.size
    val isEmpty: Boolean get() = totalCount == 0
}

data class SearchFilters(
    val type: SearchResultType = SearchResultType.ALL,
    val dateRange: DateRange? = null,
    val sortBy: SearchSortBy = SearchSortBy.RELEVANCE
)

enum class SearchResultType { ALL, CONVERSATIONS, MESSAGES, DOCUMENTS, AGENTS }
enum class SearchSortBy { RELEVANCE, DATE, NAME }

data class SearchHistoryItem(
    val query: String,
    val timestamp: Long,
    val resultCount: Int
)

// ─── Search ViewModel ──────────────────────────────────────

@HiltViewModel
class SearchViewModel @Inject constructor(
    private val searchUseCase: SearchUseCase,
    private val searchHistoryRepository: SearchHistoryRepository
) : ViewModel() {

    private val _state = MutableStateFlow(SearchState())
    val state: StateFlow<SearchState> = _state.asStateFlow()

    private var searchJob: Job? = null

    init {
        loadRecentSearches()
    }

    fun onQueryChange(query: String) {
        _state.update { it.copy(query = query, isSearchActive = true) }
        debounceSearch(query)
    }

    fun onSearch() {
        val query = _state.value.query
        if (query.isBlank()) return
        performSearch(query)
    }

    fun onClearQuery() {
        _state.update { it.copy(query = "", results = SearchResults(), isSearchActive = false) }
        searchJob?.cancel()
    }

    fun onFilterChange(type: SearchResultType) {
        _state.update { it.copy(filters = it.filters.copy(type = type)) }
        performSearch(_state.value.query)
    }

    fun onSortChange(sortBy: SearchSortBy) {
        _state.update { it.copy(filters = it.filters.copy(sortBy = sortBy)) }
    }

    fun selectRecentSearch(query: String) {
        _state.update { it.copy(query = query, isSearchActive = true) }
        performSearch(query)
    }

    fun clearSearchHistory() {
        viewModelScope.launch {
            searchHistoryRepository.clearAll()
            _state.update { it.copy(recentSearches = emptyList()) }
        }
    }

    private fun debounceSearch(query: String) {
        searchJob?.cancel()
        searchJob = viewModelScope.launch {
            delay(300) // Debounce 300ms
            if (query.isNotBlank()) {
                performSearch(query)
            }
        }
    }

    private fun performSearch(query: String) {
        viewModelScope.launch {
            _state.update { it.copy(isSearching = true, error = null) }

            try {
                val filters = _state.value.filters
                searchUseCase(query, filters)
                    .catch { e ->
                        _state.update { it.copy(isSearching = false, error = e.message) }
                    }
                    .collect { results ->
                        _state.update { it.copy(results = results, isSearching = false) }
                        saveSearchToHistory(query, results.totalCount)
                    }
            } catch (e: Exception) {
                _state.update { it.copy(isSearching = false, error = e.message) }
            }
        }
    }

    private fun loadRecentSearches() {
        viewModelScope.launch {
            searchHistoryRepository.getRecentSearches()
                .collect { searches ->
                    _state.update { it.copy(recentSearches = searches.map { s -> s.query }) }
                }
        }
    }

    private suspend fun saveSearchToHistory(query: String, resultCount: Int) {
        searchHistoryRepository.save(query, resultCount)
    }
}
```

---

## 12. Pagination State

### Pagination Implementation

```kotlin
data class PaginationState<T>(
    val items: List<T> = emptyList(),
    val currentPage: Int = 0,
    val pageSize: Int = 20,
    val isLoading: Boolean = false,
    val isLoadingMore: Boolean = false,
    val hasMore: Boolean = true,
    val error: String? = null,
    val totalItems: Int = 0
) {
    val canLoadMore: Boolean get() = hasMore && !isLoadingMore && !isLoading
    val isEmpty: Boolean get() = !isLoading && items.isEmpty()
    val isLastPage: Boolean get() = !hasMore
    val itemCount: Int get() = items.size

    companion object {
        fun <T> loading(): PaginationState<T> = PaginationState(isLoading = true)
        fun <T> error(message: String): PaginationState<T> = PaginationState(error = message)
    }
}

// ─── Pagination ViewModel ──────────────────────────────────

@HiltViewModel
class PaginatedListViewModel @Inject constructor(
    private val getItemsUseCase: GetPaginatedItemsUseCase
) : ViewModel() {

    private val _state = MutableStateFlow(PaginationState<ListItem>())
    val state: StateFlow<PaginationState<ListItem>> = _state.asStateFlow()

    init {
        loadFirstPage()
    }

    fun loadFirstPage() {
        viewModelScope.launch {
            _state.update { PaginationState.loading() }

            getItemsUseCase(page = 0, pageSize = _state.value.pageSize)
                .catch { e ->
                    _state.update { PaginationState.error(e.message ?: "Error") }
                }
                .collect { result ->
                    _state.update {
                        PaginationState(
                            items = result.items,
                            currentPage = 0,
                            pageSize = it.pageSize,
                            hasMore = result.hasMore,
                            totalItems = result.total
                        )
                    }
                }
        }
    }

    fun loadNextPage() {
        val current = _state.value
        if (!current.canLoadMore) return

        viewModelScope.launch {
            _state.update { it.copy(isLoadingMore = true, error = null) }

            val nextPage = current.currentPage + 1

            getItemsUseCase(page = nextPage, pageSize = current.pageSize)
                .catch { e ->
                    _state.update { it.copy(isLoadingMore = false, error = e.message) }
                }
                .collect { result ->
                    _state.update {
                        it.copy(
                            items = it.items + result.items,
                            currentPage = nextPage,
                            isLoadingMore = false,
                            hasMore = result.hasMore,
                            totalItems = result.total,
                            error = null
                        )
                    }
                }
        }
    }

    fun refresh() {
        loadFirstPage()
    }

    fun retry() {
        val current = _state.value
        if (current.items.isEmpty()) {
            loadFirstPage()
        } else {
            loadNextPage()
        }
    }

    fun removeItem(id: String) {
        _state.update { state ->
            state.copy(
                items = state.items.filter { it.id != id },
                totalItems = state.totalItems - 1
            )
        }
    }

    data class GetResult<T>(
        val items: List<T>,
        val hasMore: Boolean,
        val total: Int
    )
}

// ─── Composable with Pagination ────────────────────────────

@Composable
fun PaginatedListScreen(
    viewModel: PaginatedListViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()

    // Trigger next page load when near bottom
    val shouldLoadMore by remember {
        derivedStateOf {
            val lastVisibleItem = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            val totalItems = listState.layoutInfo.totalItemsCount
            lastVisibleItem >= totalItems - 3 && state.canLoadMore
        }
    }

    LaunchedEffect(shouldLoadMore) {
        if (shouldLoadMore) {
            viewModel.loadNextPage()
        }
    }

    LazyColumn(
        state = listState,
        modifier = Modifier.fillMaxSize()
    ) {
        items(
            items = state.items,
            key = { it.id }
        ) { item ->
            ListItem(item = item)
        }

        if (state.isLoadingMore) {
            item {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator()
                }
            }
        }

        if (state.error != null) {
            item {
                ErrorRetryItem(
                    message = state.error!!,
                    onRetry = { viewModel.retry() }
                )
            }
        }
    }

    if (state.isLoading && state.items.isEmpty()) {
        Box(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center
        ) {
            CircularProgressIndicator()
        }
    }

    if (state.isEmpty) {
        EmptyState(
            message = "No items found",
            modifier = Modifier.fillMaxSize()
        )
    }
}
```

---

## 13. Optimistic Updates

### Optimistic Update Pattern

```kotlin
class OptimisticUpdateManager @Inject constructor(
    private val repository: ConversationRepository,
    private val syncManager: SyncManager
) {
    sealed class UpdateResult<out T> {
        data class Optimistic<T>(val data: T) : UpdateResult<T>()
        data class Confirmed<T>(val data: T) : UpdateResult<T>()
        data class RolledBack(val originalData: Any, val error: String) : UpdateResult<Nothing>()
    }

    suspend fun <T> performOptimisticUpdate(
        localUpdate: suspend () -> T,
        networkUpdate: suspend () -> T,
        rollback: suspend (T) -> Unit,
        onError: suspend (Exception) -> Unit
    ): Flow<UpdateResult<T>> = flow {
        // 1. Perform optimistic local update
        val optimisticResult = localUpdate()
        emit(UpdateResult.Optimistic(optimisticResult))

        // 2. Queue for sync
        try {
            // 3. Attempt network update
            val confirmedResult = networkUpdate()
            emit(UpdateResult.Confirmed(confirmedResult))
        } catch (e: Exception) {
            // 4. Rollback on failure
            try {
                rollback(optimisticResult)
            } catch (rollbackException: Exception) {
                onError(rollbackException)
            }
            emit(UpdateResult.RolledBack(optimisticResult, e.message ?: "Update failed"))
            onError(e)
        }
    }.flowOn(Dispatchers.IO)
}

// ─── Usage in ViewModel ────────────────────────────────────

@HiltViewModel
class ConversationViewModel @Inject constructor(
    private val repository: ConversationRepository,
    private val optimisticManager: OptimisticUpdateManager
) : ViewModel() {

    fun pinConversation(id: String) {
        viewModelScope.launch {
            optimisticManager.performOptimisticUpdate(
                localUpdate = {
                    repository.updatePinned(id, pinned = true)
                    repository.getConversation(id)
                },
                networkUpdate = {
                    repository.syncPinnedStatus(id, pinned = true)
                },
                rollback = { previousState ->
                    repository.updatePinned(id, pinned = false)
                },
                onError = { error ->
                    _events.emit(UiEvent.ShowSnackbar("Failed to pin: ${error.message}"))
                }
            ).collect { result ->
                when (result) {
                    is UpdateResult.Optimistic -> {
                        // UI already updated optimistically
                    }
                    is UpdateResult.Confirmed -> {
                        // Server confirmed, no UI change needed
                    }
                    is UpdateResult.RolledBack -> {
                        _events.emit(UiEvent.ShowSnackbar("Update failed. Reverted."))
                    }
                }
            }
        }
    }
}
```

---

## 14. Cache Strategy

### Repository Cache Implementation

```kotlin
abstract class CachedRepository<Key : Any, Data : Any> {

    private val cache = LruCache<Key, CachedData<Data>>(50)

    data class CachedData<T>(
        val data: T,
        val timestamp: Long,
        val freshness: DataFreshness
    )

    abstract suspend fun fetchFromNetwork(key: Key): Data?
    abstract suspend fun fetchFromLocal(key: Key): Data?
    abstract suspend fun saveToLocal(key: Key, data: Data)
    abstract fun getDataFreshness(): DataFreshness

    suspend fun get(
        key: Key,
        strategy: CacheStrategy = CacheStrategy.CACHE_FIRST
    ): Result<Data> {
        return when (strategy) {
            CacheStrategy.CACHE_FIRST -> getCacheFirst(key)
            CacheStrategy.NETWORK_FIRST -> getNetworkFirst(key)
            CacheStrategy.CACHE_ONLY -> getCacheOnly(key)
            CacheStrategy.NETWORK_ONLY -> getNetworkOnly(key)
        }
    }

    private suspend fun getCacheFirst(key: Key): Result<Data> {
        // Check in-memory cache
        val memoryCached = cache.get(key)
        if (memoryCached != null && !isStale(memoryCached)) {
            return Result.success(memoryCached.data)
        }

        // Check local DB
        val localData = fetchFromLocal(key)
        if (localData != null) {
            cache.put(key, CachedData(localData, System.currentTimeMillis(), getDataFreshness()))
            if (!isStale(CachedData(localData, System.currentTimeMillis(), getDataFreshness()))) {
                return Result.success(localData)
            }
        }

        // Fetch from network
        return try {
            val networkData = fetchFromNetwork(key)
            if (networkData != null) {
                saveToLocal(key, networkData)
                cache.put(key, CachedData(networkData, System.currentTimeMillis(), getDataFreshness()))
                Result.success(networkData)
            } else if (localData != null) {
                Result.success(localData) // Stale cache
            } else {
                Result.failure(Exception("No data available"))
            }
        } catch (e: Exception) {
            if (localData != null) {
                Result.success(localData)
            } else {
                Result.failure(e)
            }
        }
    }

    private suspend fun getNetworkFirst(key: Key): Result<Data> {
        return try {
            val networkData = fetchFromNetwork(key)
            if (networkData != null) {
                saveToLocal(key, networkData)
                cache.put(key, CachedData(networkData, System.currentTimeMillis(), getDataFreshness()))
                Result.success(networkData)
            } else {
                val localData = fetchFromLocal(key)
                if (localData != null) Result.success(localData)
                else Result.failure(Exception("No data available"))
            }
        } catch (e: Exception) {
            val localData = fetchFromLocal(key)
            if (localData != null) Result.success(localData)
            else Result.failure(e)
        }
    }

    private suspend fun getCacheOnly(key: Key): Result<Data> {
        val memoryCached = cache.get(key)
        if (memoryCached != null) return Result.success(memoryCached.data)

        val localData = fetchFromLocal(key)
        return if (localData != null) {
            cache.put(key, CachedData(localData, System.currentTimeMillis(), getDataFreshness()))
            Result.success(localData)
        } else {
            Result.failure(Exception("No cached data"))
        }
    }

    private suspend fun getNetworkOnly(key: Key): Result<Data> {
        return try {
            val networkData = fetchFromNetwork(key)
            if (networkData != null) Result.success(networkData)
            else Result.failure(Exception("No data from network"))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun isStale(cached: CachedData<*>): Boolean {
        val age = System.currentTimeMillis() - cached.timestamp
        return age > cached.freshness.maxAgeMs
    }

    fun invalidate(key: Key) {
        cache.remove(key)
    }

    fun invalidateAll() {
        cache.evictAll()
    }
}
```

---

## 15. Reactive Data Flow

### Complete Reactive Flow

```
┌─────────────────────────────────────────────────────────────┐
│              Reactive Data Flow Architecture                 │
│                                                             │
│  ┌──────────┐                                               │
│  │ Remote   │    Flow<Result<T>>                            │
│  │ API      │──────────────┐                                │
│  └──────────┘              │                                │
│                            ▼                                │
│  ┌──────────┐    ┌─────────────────┐                       │
│  │ Local DB │───▶│   Repository    │                       │
│  │ (Room)   │    │                 │                       │
│  └──────────┘    │ • Flow<Result<T>>│                       │
│                  │ • emit()        │                       │
│                  └────────┬────────┘                       │
│                           │                                 │
│                           ▼                                 │
│                  ┌─────────────────┐                       │
│                  │    UseCase      │                       │
│                  │                 │                       │
│                  │ • operator {}   │                       │
│                  │ • map/filter    │                       │
│                  │ • catch {}      │                       │
│                  └────────┬────────┘                       │
│                           │                                 │
│                           ▼                                 │
│                  ┌─────────────────┐                       │
│                  │   ViewModel     │                       │
│                  │                 │                       │
│                  │ StateFlow +     │                       │
│                  │ SharedFlow      │                       │
│                  └────────┬────────┘                       │
│                           │                                 │
│                           ▼                                 │
│                  ┌─────────────────┐                       │
│                  │   Compose UI    │                       │
│                  │                 │                       │
│                  │ collectAsState  │                       │
│                  │ + LaunchedEffect│                       │
│                  └─────────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

### Repository Implementation

```kotlin
class ConversationRepositoryImpl @Inject constructor(
    private val conversationDao: ConversationDao,
    private val apiService: NexusApiService,
    private val networkMonitor: NetworkMonitor,
    private val syncQueueManager: SyncQueueManager
) : ConversationRepository {

    override fun getConversations(): Flow<Result<List<Conversation>>> = flow {
        // Emit local data immediately
        conversationDao.getAllConversations()
            .map { entities -> entities.map { it.toDomain() } }
            .collect { localConversations ->
                emit(Result.success(localConversations))
            }
    }.catch { e ->
        emit(Result.failure(e))
    }.flowOn(Dispatchers.IO)

    override fun getConversation(id: String): Flow<Result<Conversation>> = flow {
        // Emit local data
        conversationDao.getConversationById(id)
            .map { it?.toDomain() }
            .filterNotNull()
            .collect { conversation ->
                emit(Result.success(conversation))
            }

        // Try to refresh from network if online
        if (networkMonitor.isOnline()) {
            try {
                val remote = apiService.getConversation(id)
                val entity = remote.toEntity()
                conversationDao.upsert(entity)
            } catch (e: Exception) {
                // Non-fatal, local data is already emitted
            }
        }
    }.catch { e ->
        emit(Result.failure(e))
    }.flowOn(Dispatchers.IO)

    override suspend fun createConversation(
        title: String,
        agentId: String
    ): Result<Conversation> {
        return try {
            val id = UUID.randomUUID().toString()
            val entity = ConversationEntity(
                id = id,
                title = title,
                agentId = agentId,
                syncStatus = SyncStatus.PENDING
            )

            // Save locally first
            conversationDao.insertConversation(entity)

            // Try to sync
            if (networkMonitor.isOnline()) {
                try {
                    apiService.createConversation(entity.toDto())
                    conversationDao.markAsSynced(id)
                } catch (e: Exception) {
                    // Queued for sync
                    syncQueueManager.enqueue(
                        EntityType.CONVERSATION, id, SyncOperation.CREATE,
                        Json.encodeToString(entity.toDto())
                    )
                }
            } else {
                syncQueueManager.enqueue(
                    EntityType.CONVERSATION, id, SyncOperation.CREATE,
                    Json.encodeToString(entity.toDto())
                )
            }

            Result.success(entity.toDomain())
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun deleteConversation(id: String): Result<Unit> {
        return try {
            conversationDao.softDelete(id)
            syncQueueManager.enqueue(
                EntityType.CONVERSATION, id, SyncOperation.DELETE, ""
            )
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
```

---

## 16. Dependency Injection

### Hilt Setup

```kotlin
// ─── Application ───────────────────────────────────────────

@HiltAndroidApp
class NexusApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        // Initialize services
    }
}

// ─── Activity ──────────────────────────────────────────────

@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            NexusTheme {
                NexusNavHost()
            }
        }
    }
}

// ─── Fragment ──────────────────────────────────────────────

@AndroidEntryPoint
class SettingsFragment : Fragment() {
    private val viewModel: SettingsViewModel by viewModels()
}

// ─── ViewModel ─────────────────────────────────────────────

@HiltViewModel
class DashboardViewModel @Inject constructor(
    private val getConversationsUseCase: GetConversationsUseCase,
    private val getDocumentsUseCase: GetDocumentsUseCase,
    private val syncManager: SyncManager
) : ViewModel() {
    // ...
}

// ─── Repository ────────────────────────────────────────────

class ConversationRepositoryImpl @Inject constructor(
    private val conversationDao: ConversationDao,
    private val apiService: NexusApiService
) : ConversationRepository {
    // ...
}

// ─── UseCase ───────────────────────────────────────────────

class GetConversationsUseCase @Inject constructor(
    private val repository: ConversationRepository
) {
    operator fun invoke(): Flow<List<Conversation>> {
        return repository.getConversations()
            .map { result ->
                result.getOrElse { emptyList() }
            }
    }
}

// ─── Service ───────────────────────────────────────────────

class SyncManager @Inject constructor(
    private val syncQueueDao: SyncQueueDao,
    private val apiService: NexusApiService
) {
    // ...
}
```

---

## 17. Hilt Modules

### Module Definitions

```kotlin
// ─── Network Module ────────────────────────────────────────

@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    @Provides
    @Singleton
    fun provideOkHttpClient(
        certificatePinner: CertificatePinner,
        tokenManager: TokenManager
    ): OkHttpClient {
        return OkHttpClient.Builder()
            .certificatePinner(certificatePinner)
            .addInterceptor(AuthInterceptor(tokenManager))
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .build()
    }

    @Provides
    @Singleton
    fun provideRetrofit(okHttpClient: OkHttpClient): Retrofit {
        return Retrofit.Builder()
            .baseUrl(BuildConfig.BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }

    @Provides
    @Singleton
    fun provideApiService(retrofit: Retrofit): NexusApiService {
        return retrofit.create(NexusApiService::class.java)
    }

    @Provides
    @Singleton
    fun provideCertificatePinner(): CertificatePinner {
        return CertificatePinner.Builder()
            .add("api.nexus-ai.com", "sha256/AAAA...")
            .build()
    }
}

// ─── Database Module ───────────────────────────────────────

@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideDatabase(@ApplicationContext context: Context): NexusDatabase {
        return Room.databaseBuilder(context, NexusDatabase::class.java, "nexus.db")
            .addMigrations(*allMigrations)
            .build()
    }

    @Provides
    fun provideConversationDao(db: NexusDatabase): ConversationDao = db.conversationDao()

    @Provides
    fun provideMessageDao(db: NexusDatabase): MessageDao = db.messageDao()

    @Provides
    fun provideDocumentDao(db: NexusDatabase): DocumentDao = db.documentDao()

    @Provides
    fun provideAgentDao(db: NexusDatabase): AgentDao = db.agentDao()

    @Provides
    fun provideSyncQueueDao(db: NexusDatabase): SyncQueueDao = db.syncQueueDao()
}

// ─── Repository Module ─────────────────────────────────────

@Module
@InstallIn(SingletonComponent::class)
abstract class RepositoryModule {

    @Binds
    @Singleton
    abstract fun bindConversationRepository(
        impl: ConversationRepositoryImpl
    ): ConversationRepository

    @Binds
    @Singleton
    abstract fun bindDocumentRepository(
        impl: DocumentRepositoryImpl
    ): DocumentRepository

    @Binds
    @Singleton
    abstract fun bindAgentRepository(
        impl: AgentRepositoryImpl
    ): AgentRepository
}

// ─── Security Module ───────────────────────────────────────

@Module
@InstallIn(SingletonComponent::class)
object SecurityModule {

    @Provides
    @Singleton
    fun provideMasterKey(@ApplicationContext context: Context): MasterKey {
        return MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()
    }

    @Provides
    @Singleton
    fun provideEncryptedSharedPreferences(
        @ApplicationContext context: Context,
        masterKey: MasterKey
    ): EncryptedSharedPreferences {
        return EncryptedSharedPreferences.create(
            context, "nexus_secure", masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        ) as EncryptedSharedPreferences
    }

    @Provides
    @Singleton
    fun provideKeyStoreManager(): KeyStoreManager = KeyStoreManager()

    @Provides
    @Singleton
    fun provideTokenManager(
        encryptedPrefs: EncryptedSharedPreferences,
        keyStoreManager: KeyStoreManager
    ): TokenManager = TokenManager(encryptedPrefs, keyStoreManager)
}

// ─── UseCase Module ────────────────────────────────────────

@Module
@InstallIn(SingletonComponent::class)
object UseCaseModule {

    @Provides
    @Singleton
    fun provideGetConversationsUseCase(
        repository: ConversationRepository
    ): GetConversationsUseCase {
        return GetConversationsUseCase(repository)
    }

    @Provides
    @Singleton
    fun provideSendMessageUseCase(
        repository: MessageRepository,
        syncManager: SyncManager
    ): SendMessageUseCase {
        return SendMessageUseCase(repository, syncManager)
    }
}
```

---

## 18. Hilt Scoping

### Scoping Guide

| Scope | Lifetime | Use For | Example |
|-------|----------|---------|---------|
| `@Singleton` | App lifetime | Database, API, Repos | Retrofit, Room, API |
| `@ActivityScoped` | Activity lifetime | Activity-specific state | Activity ViewModel |
| `@ViewModelScoped` | ViewModel lifetime | ViewModel dependencies | UseCase, Repository |
| `@FragmentScoped` | Fragment lifetime | Fragment-specific state | Fragment arguments |
| `@ViewScoped` | View lifetime | View-specific state | View binding |
| `@ServiceScoped` | Service lifetime | Service dependencies | Service repository |

```kotlin
// ─── Singleton Scope ───────────────────────────────────────

@Singleton
class SyncManager @Inject constructor() // Lives for app lifetime

@Singleton
class TokenManager @Inject constructor() // Lives for app lifetime

// ─── ViewModel Scope ───────────────────────────────────────

@ViewModelScoped
class ChatUseCase @Inject constructor() // Lives per ViewModel

@ViewModelScoped
class MessageFormatter @Inject constructor() // Lives per ViewModel

// ─── Custom Scope ──────────────────────────────────────────

@Scope
@Retention(AnnotationRetention.RUNTIME)
annotation class SessionScoped

@SessionScoped
class SessionRepository @Inject constructor() // Lives per session
```

---

## 19. Hilt Testing

### Test Setup

```kotlin
// ─── Test Application ──────────────────────────────────────

@Module
@TestInstallIn(
    components = [SingletonComponent::class],
    replaces = [NetworkModule::class]
)
object TestNetworkModule {
    @Provides
    @Singleton
    fun provideMockApiService(): NexusApiService {
        return MockApiService()
    }
}

// ─── ViewModel Test ────────────────────────────────────────

@HiltAndroidTest
@RunWith(AndroidJUnit4::class)
class DashboardViewModelTest {

    @get:Rule
    val hiltRule = HiltAndroidRule(this)

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Inject
    lateinit var viewModel: DashboardViewModel

    @Before
    fun setup() {
        hiltRule.inject()
    }

    @Test
    fun `loading data shows loading state`() = runTest {
        // When
        viewModel.onAction(DashboardAction.Refresh)

        // Then
        val state = viewModel.state.value
        // Verify loading state was shown
    }

    @Test
    fun `selecting tab updates state`() = runTest {
        // When
        viewModel.onAction(DashboardAction.SelectTab(2))

        // Then
        val state = viewModel.state.value
        assertEquals(2, state.selectedTab)
    }
}

// ─── Repository Test ───────────────────────────────────────

@HiltAndroidTest
@RunWith(AndroidJUnit4::class)
class ConversationRepositoryTest {

    @get:Rule
    val hiltRule = HiltAndroidRule(this)

    @Inject
    lateinit var repository: ConversationRepository

    @Before
    fun setup() {
        hiltRule.inject()
    }

    @Test
    fun `getConversations returns flow`() = runTest {
        val result = repository.getConversations()
        assertNotNull(result)
    }
}

// ─── Custom Test Runner ────────────────────────────────────

@HiltAndroidTest
@RunWith(AndroidJUnit4::class)
class IntegrationTest {

    @get:Rule
    val hiltRule = HiltAndroidRule(this)

    @get:Rule
    val composeRule = createAndroidComposeRule<MainActivity>()

    @Test
    fun appLaunches() {
        composeRule.onNodeWithText("Dashboard").assertIsDisplayed()
    }
}
```

---

## 20. Coroutine Scope Management

### Coroutine Scope Patterns

```kotlin
// ─── ViewModel Scope ───────────────────────────────────────

class MyViewModel @Inject constructor() : ViewModel() {

    fun loadData() {
        // Automatically cancelled when ViewModel is cleared
        viewModelScope.launch(Dispatchers.IO) {
            val data = repository.getData()
            withContext(Dispatchers.Main) {
                _uiState.value = UiState.Success(data)
            }
        }
    }

    fun loadDataWithTimeout() {
        viewModelScope.launch {
            try {
                withTimeout(10_000L) { // 10 second timeout
                    val data = repository.getData()
                    _uiState.value = UiState.Success(data)
                }
            } catch (e: TimeoutCancellationException) {
                _uiState.value = UiState.Error("Request timed out")
            }
        }
    }

    fun loadMultipleInParallel() {
        viewModelScope.launch {
            val deferred1 = async(Dispatchers.IO) { repository.getConversations() }
            val deferred2 = async(Dispatchers.IO) { repository.getDocuments() }
            val deferred3 = async(Dispatchers.IO) { repository.getAgents() }

            try {
                val conversations = deferred1.await()
                val documents = deferred2.await()
                val agents = deferred3.await()

                _uiState.update {
                    it.copy(
                        conversations = conversations,
                        documents = documents,
                        agents = agents
                    )
                }
            } catch (e: Exception) {
                _uiState.update { UiState.Error(e.message ?: "Error") }
            }
        }
    }
}

// ─── Lifecycle Scope ───────────────────────────────────────

@Composable
fun MyScreen() {
    val lifecycleOwner = LocalLifecycleOwner.current

    // Cancelled when screen leaves composition
    LaunchedEffect(Unit) {
        lifecycleOwner.lifecycle.repeatOnLifecycle(Lifecycle.State.STARTED) {
            // Only runs when screen is visible
            viewModel.events.collect { event ->
                handleEvent(event)
            }
        }
    }
}

// ─── Custom Coroutine Scope ────────────────────────────────

@Composable
fun rememberCoroutineScope(): CoroutineScope {
    val lifecycleOwner = LocalLifecycleOwner.current
    return remember(lifecycleOwner) {
        CoroutineScope(
            SupervisorJob() + lifecycleOwner.lifecycle.coroutineScope.coroutineContext
        )
    }
}

// ─── Error Handling in Coroutines ──────────────────────────

class SafeCoroutineScope @Inject constructor() {

    fun launchSafely(
        scope: CoroutineScope,
        onError: (Throwable) -> Unit = {},
        block: suspend CoroutineScope.() -> Unit
    ): Job {
        return scope.launch {
            try {
                block()
            } catch (e: CancellationException) {
                throw e // Never catch cancellation
            } catch (e: Exception) {
                onError(e)
            }
        }
    }

    suspend fun <T> runSafely(
        onError: (Throwable) -> T? = { null },
        block: suspend () -> T
    ): T? {
        return try {
            block()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            onError(e)
        }
    }
}
```

---

## 21. Flow Collection in Compose

### Compose Flow Collection Patterns

```kotlin
// ─── Pattern 1: collectAsStateWithLifecycle ─────────────────

@Composable
fun StateScreen(viewModel: MyViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    when (uiState) {
        is UiState.Loading -> LoadingScreen()
        is UiState.Success -> SuccessScreen(data = uiState.data)
        is UiState.Error -> ErrorScreen(message = uiState.message)
    }
}

// ─── Pattern 2: Multiple Flows ─────────────────────────────

@Composable
fun MultiFlowScreen(viewModel: MultiFlowViewModel = hiltViewModel()) {
    val state1 by viewModel.state1.collectAsStateWithLifecycle()
    val state2 by viewModel.state2.collectAsStateWithLifecycle()
    val state3 by viewModel.state3.collectAsStateWithLifecycle()

    Column {
        Text("State 1: $state1")
        Text("State 2: $state2")
        Text("State 3: $state3")
    }
}

// ─── Pattern 3: Events with LaunchedEffect ─────────────────

@Composable
fun EventScreen(viewModel: EventViewModel = hiltViewModel()) {
    val snackbarHostState = remember { SnackbarHostState() }

    // Collect events - survives recomposition
    LaunchedEffect(Unit) {
        viewModel.events.collect { event ->
            when (event) {
                is UiEvent.ShowSnackbar -> {
                    snackbarHostState.showSnackbar(event.message)
                }
                is UiEvent.Navigate -> {
                    // navigation handled elsewhere
                }
            }
        }
    }

    Scaffold(snackbarHost = { SnackbarHost(snackbarHostState) }) { padding ->
        // Content
    }
}

// ─── Pattern 4: State with Side Effects ────────────────────

@Composable
fun SideEffectScreen(viewModel: SideEffectViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val sideEffect by viewModel.sideEffect.collectAsStateWithLifecycle(initialValue = null)

    // Handle side effects
    LaunchedEffect(sideEffect) {
        sideEffect?.let { effect ->
            when (effect) {
                is SideEffect.ShowToast -> {
                    Toast.makeText(context, effect.message, Toast.LENGTH_SHORT).show()
                }
                is SideEffect.Navigate -> {
                    navController.navigate(effect.route)
                }
            }
        }
    }

    // UI
    Box(modifier = Modifier.fillMaxSize()) {
        when (state) {
            is UiState.Loading -> CircularProgressIndicator()
            is UiState.Success -> Content(state.data)
            is UiState.Error -> ErrorMessage(state.message)
        }
    }
}

// ─── Pattern 5: Derived State ──────────────────────────────

@Composable
fun DerivedStateScreen(viewModel: DerivedStateViewModel = hiltViewModel()) {
    val items by viewModel.items.collectAsStateWithLifecycle()
    val searchQuery by viewModel.searchQuery.collectAsStateWithLifecycle()

    val filteredItems by remember(items, searchQuery) {
        derivedStateOf {
            if (searchQuery.isBlank()) items
            else items.filter { it.name.contains(searchQuery, ignoreCase = true) }
        }
    }

    val itemCount by remember(items) {
        derivedStateOf { items.size }
    }

    Column {
        Text("Total: $itemCount items")
        SearchBar(query = searchQuery, onQueryChange = viewModel::updateSearchQuery)
        LazyColumn {
            items(filteredItems) { item ->
                ItemRow(item)
            }
        }
    }
}
```

---

## 22. State Preservation

### State Preservation Patterns

```kotlin
// ─── remember ──────────────────────────────────────────────

@Composable
fun Counter() {
    // Survives recomposition, lost on configuration change
    var count by remember { mutableIntStateOf(0) }

    Button(onClick = { count++ }) {
        Text("Count: $count")
    }
}

// ─── rememberSaveable ──────────────────────────────────────

@Composable
fun SearchScreen() {
    // Survives configuration changes and process death
    var query by rememberSaveable { mutableStateOf("") }
    var isExpanded by rememberSaveable { mutableStateOf(false) }

    Column {
        OutlinedTextField(
            value = query,
            onValueChange = { query = it },
            modifier = Modifier.fillMaxWidth()
        )
    }
}

// ─── rememberSaveable with custom type ─────────────────────

@Parcelize
data class FormState(
    val name: String = "",
    val email: String = "",
    val step: Int = 0
) : Parcelable

@Composable
fun MultiStepForm() {
    var formState by rememberSaveable { mutableStateOf(FormState()) }

    // ...
}

// ─── ViewModel survives configuration change ───────────────

@Composable
fun ScreenWithViewModel() {
    // ViewModel survives configuration changes
    val viewModel: MyViewModel = hiltViewModel()

    val state by viewModel.state.collectAsStateWithLifecycle()

    // State in ViewModel is automatically preserved
    // Only local UI state needs rememberSaveable
    var isDialogShown by rememberSaveable { mutableStateOf(false) }

    if (isDialogShown) {
        MyDialog(onDismiss = { isDialogShown = false })
    }
}

// ─── SavedStateHandle in ViewModel ─────────────────────────

@HiltViewModel
class SavedStateViewModel @Inject constructor(
    private val savedStateHandle: SavedStateHandle
) : ViewModel() {

    // Automatically saved and restored
    val query: StateFlow<String> = savedStateHandle.getStateFlow("query", "")
    val page: StateFlow<Int> = savedStateHandle.getStateFlow("page", 0)

    fun updateQuery(newQuery: String) {
        savedStateHandle["query"] = newQuery
    }

    fun nextPage() {
        val currentPage = savedStateHandle.get<Int>("page") ?: 0
        savedStateHandle["page"] = currentPage + 1
    }
}
```

### State Preservation Decision Matrix

```
┌──────────────────────────────────────────────────────────┐
│          State Preservation Decision Matrix               │
│                                                          │
│  ┌──────────────────┬───────────────┬──────────────────┐ │
│  │ State Type       │ Survives      │ Survives Process │ │
│  │                  │ Config Change │ Death            │ │
│  ├──────────────────┼───────────────┼──────────────────┤ │
│  │ val in composable│ No            │ No               │ │
│  │ remember {}      │ No            │ No               │ │
│  │ rememberSaveable │ Yes           │ Yes              │ │
│  │ ViewModel        │ Yes           │ Yes*             │ │
│  │ SavedStateHandle │ Yes           │ Yes              │ │
│  │ DataStore        │ Yes           │ Yes              │ │
│  │ Room             │ Yes           │ Yes              │ │
│  └──────────────────┴───────────────┴──────────────────┘ │
│                                                          │
│  * ViewModel survives process death only when using       │
│    SavedStateHandle or SavedStateViewModelFactory         │
└──────────────────────────────────────────────────────────┘
```

---

## 23. Configuration Change Handling

### Configuration Change Strategy

```kotlin
// ViewModel survives configuration changes automatically

@HiltViewModel
class MainViewModel @Inject constructor(
    private val savedStateHandle: SavedStateHandle
) : ViewModel() {

    // This state survives configuration changes
    private val _uiState = MutableStateFlow(MainUiState())
    val uiState: StateFlow<MainUiState> = _uiState.asStateFlow()

    // SavedStateHandle for process death
    val scrollPosition: StateFlow<Int> = savedStateHandle.getStateFlow("scroll", 0)
    val selectedTab: StateFlow<Int> = savedStateHandle.getStateFlow("tab", 0)

    fun saveScrollPosition(position: Int) {
        savedStateHandle["scroll"] = position
    }

    fun selectTab(index: Int) {
        savedStateHandle["tab"] = index
    }
}

// Activity configuration
@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // ViewModel and its state are automatically restored
        setContent {
            NexusTheme {
                NexusNavHost()
            }
        }
    }
}

// Fragment configuration
@AndroidEntryPoint
class ChatFragment : Fragment() {
    // ViewModel scoped to Fragment, survives config changes
    private val viewModel: ChatViewModel by viewModels()

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        // State is automatically restored
    }
}
```

---

## 24. Process Death Handling

### Process Death Recovery

```kotlin
@HiltViewModel
class RecoverableViewModel @Inject constructor(
    private val savedStateHandle: SavedStateHandle,
    private val repository: ConversationRepository
) : ViewModel() {

    // Keys for saved state
    companion object {
        const val KEY_CONVERSATION_ID = "conversation_id"
        const val KEY_INPUT_TEXT = "input_text"
        const val KEY_SCROLL_POSITION = "scroll_position"
        const val KEY_DATA_LOADED = "data_loaded"
    }

    private val _uiState = MutableStateFlow(UiState())
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    // Restore state from SavedStateHandle
    private val conversationId: String?
        get() = savedStateHandle[KEY_CONVERSATION_ID]

    private val inputText: String
        get() = savedStateHandle[KEY_INPUT_TEXT] ?: ""

    private val dataLoaded: Boolean
        get() = savedStateHandle[KEY_DATA_LOADED] ?: false

    init {
        // Check if we need to reload data after process death
        if (!dataLoaded) {
            loadInitialData()
        } else {
            // Restore from saved state
            restoreState()
        }
    }

    private fun loadInitialData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            try {
                val conversations = repository.getConversations().first()
                val documents = repository.getDocuments().first()

                _uiState.update {
                    it.copy(
                        conversations = conversations,
                        documents = documents,
                        isLoading = false,
                        inputText = inputText
                    )
                }

                savedStateHandle[KEY_DATA_LOADED] = true
            } catch (e: Exception) {
                _uiState.update { it.copy(error = e.message, isLoading = false) }
            }
        }
    }

    private fun restoreState() {
        // Minimal restoration - data is still valid
        _uiState.update {
            it.copy(
                inputText = inputText,
                isLoading = false
            )
        }
    }

    fun updateInput(text: String) {
        savedStateHandle[KEY_INPUT_TEXT] = text
        _uiState.update { it.copy(inputText = text) }
    }

    fun selectConversation(id: String) {
        savedStateHandle[KEY_CONVERSATION_ID] = id
        _uiState.update { it.copy(selectedConversationId = id) }
    }

    fun saveScrollPosition(position: Int) {
        savedStateHandle[KEY_SCROLL_POSITION] = position
    }

    data class UiState(
        val conversations: List<Conversation> = emptyList(),
        val documents: List<Document> = emptyList(),
        val selectedConversationId: String? = null,
        val inputText: String = "",
        val isLoading: Boolean = true,
        val error: String? = null
    )
}
```

### Process Death Recovery Flow

```
┌─────────────────────────────────────────────────────────┐
│          Process Death Recovery Flow                     │
│                                                         │
│  ┌──────────┐                                           │
│  │ App in   │                                           │
│  │ Background│                                          │
│  └────┬─────┘                                           │
│       │                                                 │
│       ▼                                                 │
│  ┌──────────┐                                           │
│  │ System   │                                           │
│  │ kills    │                                           │
│  │ process  │                                           │
│  └────┬─────┘                                           │
│       │                                                 │
│       ▼                                                 │
│  ┌──────────┐    ┌──────────────────┐                   │
│  │ SavedState│    │ ViewModel is     │                   │
│  │ Handle   │───▶│ recreated        │                   │
│  │ preserved│    │ from saved state │                   │
│  └──────────┘    └───────┬──────────┘                   │
│                          │                              │
│                          ▼                              │
│  ┌──────────────────────────────────┐                   │
│  │ ViewModel.init {                │                   │
│  │   if (!dataLoaded) {            │                   │
│  │     loadInitialData()           │                   │
│  │   } else {                      │                   │
│  │     restoreState()              │                   │
│  │   }                             │                   │
│  │ }                               │                   │
│  └──────────────────────────────────┘                   │
│                          │                              │
│                          ▼                              │
│  ┌──────────────────────────────────┐                   │
│  │ UI restored with correct state   │                   │
│  └──────────────────────────────────┘                   │
└─────────────────────────────────────────────────────────┘
```

---

## 25. Memory Management

### Memory Management Practices

```kotlin
// ─── Leak Detection ────────────────────────────────────────

class MemoryManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun checkMemoryUsage(): MemoryInfo {
        val runtime = Runtime.getRuntime()
        val maxMemory = runtime.maxMemory()
        val totalMemory = runtime.totalMemory()
        val freeMemory = runtime.freeMemory()
        val usedMemory = totalMemory - freeMemory

        return MemoryInfo(
            maxMemory = maxMemory,
            totalMemory = totalMemory,
            usedMemory = usedMemory,
            freeMemory = freeMemory,
            usagePercentage = (usedMemory.toFloat() / maxMemory * 100)
        )
    }

    fun getMemoryWarningLevel(): WarningLevel {
        val info = checkMemoryUsage()
        return when {
            info.usagePercentage > 90 -> WarningLevel.CRITICAL
            info.usagePercentage > 75 -> WarningLevel.HIGH
            info.usagePercentage > 60 -> WarningLevel.MEDIUM
            else -> WarningLevel.LOW
        }
    }

    data class MemoryInfo(
        val maxMemory: Long,
        val totalMemory: Long,
        val usedMemory: Long,
        val freeMemory: Long,
        val usagePercentage: Float
    )

    enum class WarningLevel { LOW, MEDIUM, HIGH, CRITICAL }
}

// ─── Proper Flow Collection ────────────────────────────────

// WRONG: Collecting in a non-lifecycle-aware way
class BadViewModel : ViewModel() {
    fun loadData() {
        // This will leak if the coroutine outlives the ViewModel
        CoroutineScope(Dispatchers.IO).launch {
            flow.collect { /* ... */ }
        }
    }
}

// RIGHT: Using viewModelScope
class GoodViewModel : ViewModel() {
    fun loadData() {
        // This is automatically cancelled when ViewModel is cleared
        viewModelScope.launch {
            flow.collect { /* ... */ }
        }
    }
}

// ─── WeakReference Usage ───────────────────────────────────

class WeakReferenceCache {
    private val cache = mutableMapOf<String, WeakReference<Any>>()

    fun get(key: String): Any? {
        return cache[key]?.get()
    }

    fun put(key: String, value: Any) {
        cache[key] = WeakReference(value)
    }

    fun cleanup() {
        val iterator = cache.entries.iterator()
        while (iterator.hasNext()) {
            val entry = iterator.next()
            if (entry.value.get() == null) {
                iterator.remove()
            }
        }
    }
}

// ─── Image Cache Management ────────────────────────────────

class ImageCacheManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val memoryCache = LruCache<String, Bitmap>(
        (Runtime.getRuntime().maxMemory() / 8).toInt()
    )

    fun getFromMemoryCache(key: String): Bitmap? {
        return memoryCache.get(key)
    }

    fun putToMemoryCache(key: String, bitmap: Bitmap) {
        if (getBitmapSize(bitmap) > memoryCache.maxSize() / 2) {
            return // Don't cache overly large images
        }
        memoryCache.put(key, bitmap)
    }

    fun clearMemoryCache() {
        memoryCache.evictAll()
    }

    fun clearDiskCache() {
        context.cacheDir.listFiles()?.forEach { it.delete() }
    }

    private fun getBitmapSize(bitmap: Bitmap): Int {
        return bitmap.byteCount
    }
}
```

### Memory Leak Prevention

```
┌─────────────────────────────────────────────────────────┐
│            Memory Leak Prevention                         │
│                                                         │
│  DO:                                                    │
│  ✓ Use viewModelScope for coroutines                    │
│  ✓ Cancel flows when ViewModel is cleared               │
│  ✓ Use WeakReference for caches                         │
│  ✓ Release resources in onCleared()                     │
│  ✓ Use DisposableEffect for lifecycle cleanup           │
│  ✓ Limit cache sizes with LruCache                      │
│  ✓ Collect flows with collectAsStateWithLifecycle       │
│                                                         │
│  DON'T:                                                 │
│  ✗ Hold Activity/Context in ViewModel                   │
│  ✗ Use CoroutineScope() instead of viewModelScope       │
│  ✗ Store Bitmaps without size limits                    │
│  ✗ Keep static references to Views                      │
│  ✗ Forget to cancel registered callbacks                │
│  ✗ Use inner classes without WeakReference              │
│  ✗ Store large objects in SavedStateHandle              │
└─────────────────────────────────────────────────────────┘
```

---

## 26. State Debugging

### Debug Tools and Techniques

```kotlin
// ─── Compose Preview ───────────────────────────────────────

@Preview(showBackground = true)
@Composable
fun ChatScreenPreview() {
    NexusTheme {
        ChatScreen(
            uiState = ChatUiState(
                conversations = listOf(
                    Conversation(
                        id = "1",
                        title = "Test Chat",
                        lastMessage = "Hello world",
                        unreadCount = 2
                    ),
                    Conversation(
                        id = "2",
                        title = "Another Chat",
                        lastMessage = "How are you?",
                        unreadCount = 0
                    )
                ),
                isLoading = false
            ),
            onAction = {}
        )
    }
}

@Preview(showBackground = true, name = "Loading State")
@Composable
fun ChatScreenLoadingPreview() {
    NexusTheme {
        ChatScreen(
            uiState = ChatUiState(isLoading = true),
            onAction = {}
        )
    }
}

@Preview(showBackground = true, name = "Error State")
@Composable
fun ChatScreenErrorPreview() {
    NexusTheme {
        ChatScreen(
            uiState = ChatUiState(error = "Failed to load conversations"),
            onAction = {}
        )
    }
}

// ─── State Inspector (Debug) ───────────────────────────────

@Composable
fun <T> StateInspector(
    name: String,
    state: T,
    enabled: Boolean = BuildConfig.DEBUG
) {
    if (!enabled) return

    var expanded by remember { mutableStateOf(false) }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(4.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.tertiaryContainer.copy(alpha = 0.3f)
        )
    ) {
        Column(modifier = Modifier.padding(8.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "🔍 $name",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onTertiaryContainer
                )
                IconButton(
                    onClick = { expanded = !expanded },
                    modifier = Modifier.size(24.dp)
                ) {
                    Icon(
                        if (expanded) Icons.Filled.ExpandLess else Icons.Filled.ExpandMore,
                        contentDescription = "Toggle",
                        modifier = Modifier.size(16.dp)
                    )
                }
            }

            if (expanded) {
                Text(
                    text = state.toString(),
                    style = MaterialTheme.typography.labelSmall.copy(
                        fontFamily = FontFamily.Monospace
                    ),
                    color = MaterialTheme.colorScheme.onTertiaryContainer
                )
            }
        }
    }
}

// ─── Usage ─────────────────────────────────────────────────

@Composable
fun DebuggableScreen(viewModel: MyViewModel = hiltViewModel()) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize()) {
        // Debug state inspector
        StateInspector("UiState", state)

        // Actual screen content
        when (state) {
            is UiState.Loading -> LoadingScreen()
            is UiState.Success -> ContentScreen(state.data)
            is UiState.Error -> ErrorScreen(state.message)
        }
    }
}
```

### Layout Inspector Tags

```kotlin
// Compose semantics for debugging
@Composable
fun DebuggableItem(
    item: ListItem,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier
            .semantics {
                // Adds debug tag for Layout Inspector
                debugLabel = "ListItem:${item.id}"
            }
            .padding(16.dp)
    ) {
        Text(text = item.title)
    }
}
```

---

## Summary

| Component | Purpose | Technology |
|-----------|---------|------------|
| **ViewModel** | UI state holder | ViewModel + Hilt |
| **StateFlow** | Reactive state | Kotlin Flow |
| **SharedFlow** | One-time events | Kotlin Flow |
| **State Hoisting** | Stateless composables | Compose |
| **Sealed Classes** | Type-safe states | Kotlin |
| **SavedStateHandle** | Process death recovery | AndroidX |
| **Hilt** | Dependency injection | Dagger/Hilt |
| **Coroutines** | Async operations | Kotlin Coroutines |
| **Flow** | Reactive streams | Kotlin Flow |
| **LruCache** | Memory cache | Android |
| **DataStore** | Persistent prefs | AndroidX |

The state management architecture ensures:
- **Unidirectional data flow** for predictable state changes
- **Type-safe states** via sealed classes
- **Lifecycle-aware collection** via `collectAsStateWithLifecycle`
- **Process death recovery** via `SavedStateHandle`
- **Memory safety** via `viewModelScope` and proper cleanup
- **Testability** via Hilt and stateless composables
