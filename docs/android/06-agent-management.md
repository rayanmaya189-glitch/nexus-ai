# 06 - Agent Management

Agent management provides the full lifecycle for AI agents: list, create, configure,
execute, and monitor. Agents are the primary units of work in Nexus AI — each agent
wraps a model with specific tools, permissions, document set bindings, and SQL
connections.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Agent List Screen](#agent-list-screen)
3. [Agent List ViewModel](#agent-list-viewmodel)
4. [Agent Card Composable](#agent-card-composable)
5. [Agent Detail Screen](#agent-detail-screen)
6. [Agent Creation Screen](#agent-creation-screen)
7. [Agent Configuration Screen](#agent-configuration-screen)
8. [Agent Execution Screen](#agent-execution-screen)
9. [Agent Execution History](#agent-execution-history)
10. [Agent Performance Metrics](#agent-performance-metrics)
11. [Agent Settings](#agent-settings)
12. [Agent Tools Configuration](#agent-tools-configuration)
13. [Agent-Document Set Binding](#agent-document-set-binding)
14. [Agent-Database Connection](#agent-database-connection)
15. [Agent Test Connection Flow](#agent-test-connection-flow)
16. [Agent Schema Discovery](#agent-schema-discovery)
17. [Agent Table Binding](#agent-table-binding)
18. [Agent Execution Workflow](#agent-execution-workflow)
19. [Agent Execution Visualization](#agent-execution-visualization)
20. [Agent Error Handling](#agent-error-handling)
21. [Agent API Integration](#agent-api-integration)
22. [Agent Data Models](#agent-data-models)
23. [Agent Caching](#agent-caching)
24. [Pull-to-Refresh](#pull-to-refresh)
25. [Search and Filter](#search-and-filter)
26. [Empty States](#empty-states)
27. [Loading States](#loading-states)
28. [Accessibility](#accessibility)
29. [Responsive Design](#responsive-design)

---

## Architecture Overview

```
┌────────────────────────────────────────────────────────────────┐
│                        UI Layer                                │
│  AgentListScreen ──► AgentDetailScreen ──► AgentExecutionView   │
│       │                    │                      │             │
│       ▼                    ▼                      ▼             │
│  AgentListVM          AgentDetailVM          ExecutionVM        │
│       │                    │                      │             │
│       ▼                    ▼                      ▼             │
│  ┌─────────────────────────────────────────────────────┐       │
│  │                AgentRepository                      │       │
│  │  ├── getAgents()        ├── createAgent()           │       │
│  │  ├── getAgent(id)       ├── updateAgent()           │       │
│  │  ├── deleteAgent(id)    ├── executeAgent()          │       │
│  │  ├── bindDocumentSets() ├── bindSqlConnections()    │       │
│  │  └── getExecutionHistory()                          │       │
│  └─────────────────────────────────────────────────────┘       │
│                         │                                      │
│  Data Layer            ▼                                      │
│  ├── AgentApiService (Retrofit)                                │
│  ├── AgentDao (Room)                                           │
│  └── AgentPreferences (DataStore)                              │
└────────────────────────────────────────────────────────────────┘
```

---

## Agent List Screen

The agent list supports grid and list views, with search, filter, and sort controls.

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AgentListScreen(
    viewModel: AgentListViewModel = hiltViewModel(),
    onAgentClick: (String) -> Unit,
    onCreateAgent: () -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val isRefreshing = viewModel.isRefreshing.collectAsStateWithLifecycle()

    Scaffold(
        topBar = {
            AgentListTopBar(
                searchQuery = state.searchQuery,
                onSearchChange = { viewModel.onAction(AgentListAction.UpdateSearch(it)) },
                onFilterClick = { viewModel.onAction(AgentListAction.OpenFilters) },
                onSortClick = { viewModel.onAction(AgentListAction.OpenSort) },
                viewMode = state.viewMode,
                onViewModeToggle = { viewModel.onAction(AgentListAction.ToggleViewMode) }
            )
        },
        floatingActionButton = {
            ExtendedFloatingActionButton(
                onClick = onCreateAgent,
                icon = { Icon(Icons.Default.Add, "Create agent") },
                text = { Text("New Agent") }
            )
        }
    ) { padding ->
        SwipeRefresh(
            state = rememberSwipeRefreshState(isRefreshing),
            onRefresh = { viewModel.onAction(AgentListAction.Refresh) },
            modifier = Modifier.padding(padding)
        ) {
            when {
                state.isLoading -> AgentListSkeleton()
                state.error != null -> AgentListError(
                    error = state.error!!,
                    onRetry = { viewModel.onAction(AgentListAction.Refresh) }
                )
                state.agents.isEmpty() && state.searchQuery.isBlank() -> AgentListEmpty(
                    onCreateAgent = onCreateAgent
                )
                else -> {
                    val grouped = state.agents.groupBy { it.status }
                    LazyColumn(
                        contentPadding = PaddingValues(16.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        AgentStatusGroup.entries.forEach { group ->
                            val agents = grouped[group.status] ?: emptyList()
                            if (agents.isNotEmpty()) {
                                item(key = "header_${group.name}") {
                                    Text(
                                        text = "${group.label} (${agents.size})",
                                        style = MaterialTheme.typography.titleSmall,
                                        fontWeight = FontWeight.SemiBold,
                                        modifier = Modifier.padding(vertical = 8.dp)
                                    )
                                }
                                items(agents, key = { it.agentId }) { agent ->
                                    AgentCard(
                                        agent = agent,
                                        onClick = { onAgentClick(agent.agentId) }
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

enum class AgentStatusGroup(val label: String, val status: AgentStatusType) {
    ONLINE("Online", AgentStatusType.ONLINE),
    RUNNING("Running", AgentStatusType.RUNNING),
    OFFLINE("Offline", AgentStatusType.OFFLINE),
    ERROR("Error", AgentStatusType.ERROR)
}
```

### Grid View Alternative

```kotlin
@Composable
fun AgentGrid(
    agents: List<AgentSummary>,
    onAgentClick: (String) -> Unit
) {
    LazyVerticalStaggeredGrid(
        columns = StaggeredGridCells.Adaptive(minSize = 280.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        contentPadding = PaddingValues(16.dp)
    ) {
        items(agents, key = { it.agentId }) { agent ->
            AgentCardCompact(agent = agent, onClick = { onAgentClick(agent.agentId) })
        }
    }
}
```

---

## Agent List ViewModel

```kotlin
@HiltViewModel
class AgentListViewModel @Inject constructor(
    private val agentRepository: AgentRepository,
    private val agentPreferences: AgentPreferences,
    @IoDispatcher private val ioDispatcher: CoroutineDispatcher
) : ViewModel() {

    private val _state = MutableStateFlow(AgentListState())
    val state: StateFlow<AgentListState> = _state.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = _isRefreshing.asStateFlow()

    private var allAgents: List<AgentSummary> = emptyList()

    init {
        loadAgents()
    }

    fun onAction(action: AgentListAction) {
        when (action) {
            is AgentListAction.Refresh -> refresh()
            is AgentListAction.UpdateSearch -> {
                _state.update { it.copy(searchQuery = action.query) }
                filterAgents()
            }
            is AgentListAction.SortBy -> {
                _state.update { it.copy(sortBy = action.sort) }
                filterAgents()
            }
            is AgentListAction.FilterByStatus -> {
                _state.update {
                    it.copy(
                        statusFilter = if (it.statusFilter == action.status) null else action.status
                    )
                }
                filterAgents()
            }
            is AgentListAction.ToggleViewMode -> {
                val newMode = when (_state.value.viewMode) {
                    ViewMode.LIST -> ViewMode.GRID
                    ViewMode.GRID -> ViewMode.LIST
                }
                _state.update { it.copy(viewMode = newMode) }
            }
            is AgentListAction.OpenFilters -> {
                _state.update { it.copy(showFilters = true) }
            }
            is AgentListAction.OpenSort -> {
                _state.update { it.copy(showSortSheet = true) }
            }
            is AgentListAction.DismissDialogs -> {
                _state.update { it.copy(showFilters = false, showSortSheet = false) }
            }
        }
    }

    private fun loadAgents() {
        viewModelScope.launch(ioDispatcher) {
            _state.update { it.copy(isLoading = true, error = null) }
            try {
                allAgents = agentRepository.getAgents()
                filterAgents()
                _state.update { it.copy(isLoading = false) }
            } catch (e: Exception) {
                _state.update {
                    it.copy(isLoading = false, error = e.message ?: "Failed to load agents")
                }
            }
        }
    }

    private fun filterAgents() {
        viewModelScope.launch(ioDispatcher) {
            val state = _state.value
            var filtered = allAgents

            // Search
            if (state.searchQuery.isNotBlank()) {
                filtered = filtered.filter {
                    it.name.contains(state.searchQuery, ignoreCase = true) ||
                    it.model.contains(state.searchQuery, ignoreCase = true) ||
                    it.description.contains(state.searchQuery, ignoreCase = true)
                }
            }

            // Status filter
            if (state.statusFilter != null) {
                filtered = filtered.filter { it.status == state.statusFilter }
            }

            // Sort
            filtered = when (state.sortBy) {
                SortOption.NAME -> filtered.sortedBy { it.name.lowercase() }
                SortOption.LAST_USED -> filtered.sortedByDescending { it.lastUsedAt }
                SortOption.TASKS -> filtered.sortedByDescending { it.totalTasks }
                SortOption.CREATED -> filtered.sortedByDescending { it.createdAt }
                SortOption.LATENCY -> filtered.sortedBy { it.avgLatencyMs }
            }

            _state.update { it.copy(agents = filtered) }
        }
    }

    private fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            loadAgents()
            _isRefreshing.value = false
        }
    }
}

data class AgentListState(
    val isLoading: Boolean = false,
    val agents: List<AgentSummary> = emptyList(),
    val searchQuery: String = "",
    val sortBy: SortOption = SortOption.NAME,
    val statusFilter: AgentStatusType? = null,
    val viewMode: ViewMode = ViewMode.LIST,
    val showFilters: Boolean = false,
    val showSortSheet: Boolean = false,
    val error: String? = null
)

sealed interface AgentListAction {
    data object Refresh : AgentListAction
    data class UpdateSearch(val query: String) : AgentListAction
    data class SortBy(val sort: SortOption) : AgentListAction
    data class FilterByStatus(val status: AgentStatusType) : AgentListAction
    data object ToggleViewMode : AgentListAction
    data object OpenFilters : AgentListAction
    data object OpenSort : AgentListAction
    data object DismissDialogs : AgentListAction
}

enum class SortOption(val label: String) {
    NAME("Name"), LAST_USED("Last Used"), TASKS("Total Tasks"),
    CREATED("Created Date"), LATENCY("Avg Latency")
}

enum class ViewMode { LIST, GRID }
```

---

## Agent Card Composable

```kotlin
@Composable
fun AgentCard(
    agent: AgentSummary,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val statusColor = when (agent.status) {
        AgentStatusType.ONLINE -> Color(0xFF4CAF50)
        AgentStatusType.RUNNING -> MaterialTheme.colorScheme.primary
        AgentStatusType.OFFLINE -> Color(0xFF9E9E9E)
        AgentStatusType.ERROR -> MaterialTheme.colorScheme.error
    }

    Card(
        onClick = onClick,
        modifier = modifier.fillMaxWidth(),
        elevation = CardDefaults.cardElevation(defaultElevation = 1.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Agent avatar
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(CircleShape)
                    .background(statusColor.copy(alpha = 0.12f)),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = Icons.Default.SmartToy,
                    contentDescription = agent.name,
                    tint = statusColor,
                    modifier = Modifier.size(24.dp)
                )
            }

            Spacer(modifier = Modifier.width(16.dp))

            Column(modifier = Modifier.weight(1f)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = agent.name,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    StatusChip(status = agent.status, color = statusColor)
                }
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = agent.model,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.height(4.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    AgentStat(icon = Icons.Default.Task, value = "${agent.totalTasks} tasks")
                    AgentStat(icon = Icons.Default.Speed, value = "${agent.avgLatencyMs}ms")
                    AgentStat(icon = Icons.Default.CheckCircle, value = "${agent.successRate}%")
                }
            }

            Icon(
                imageVector = Icons.Default.ChevronRight,
                contentDescription = "View agent",
                tint = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
fun AgentStat(icon: ImageVector, value: String) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        Icon(
            imageVector = icon,
            contentDescription = null,
            modifier = Modifier.size(12.dp),
            tint = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Spacer(modifier = Modifier.width(4.dp))
        Text(
            text = value,
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
fun StatusChip(status: AgentStatusType, color: Color) {
    Surface(
        shape = RoundedCornerShape(12.dp),
        color = color.copy(alpha = 0.12f)
    ) {
        Text(
            text = status.name,
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp),
            style = MaterialTheme.typography.labelSmall,
            color = color,
            fontWeight = FontWeight.Medium
        )
    }
}
```

### Compact Card (Grid View)

```kotlin
@Composable
fun AgentCardCompact(
    agent: AgentSummary,
    onClick: () -> Unit
) {
    Card(onClick = onClick) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    imageVector = Icons.Default.SmartToy,
                    contentDescription = null,
                    modifier = Modifier.size(32.dp),
                    tint = MaterialTheme.colorScheme.primary
                )
                StatusDot(status = agent.status)
            }
            Spacer(modifier = Modifier.height(12.dp))
            Text(
                text = agent.name,
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
            Text(
                text = agent.model,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Spacer(modifier = Modifier.height(8.dp))
            LinearProgressIndicator(
                progress = { agent.successRate / 100f },
                modifier = Modifier.fillMaxWidth().height(4.dp).clip(RoundedCornerShape(2.dp)),
                color = Color(0xFF4CAF50)
            )
        }
    }
}
```

---

## Agent Detail Screen

Tabbed layout with Overview, Configuration, Execution, and History tabs.

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AgentDetailScreen(
    agentId: String,
    viewModel: AgentDetailViewModel = hiltViewModel(),
    onBack: () -> Unit,
    onExecute: (String) -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var selectedTab by remember { mutableIntStateOf(0) }

    val tabs = listOf("Overview", "Config", "Execute", "History")

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(state.agent?.name ?: "Agent Detail") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { viewModel.onAction(AgentDetailAction.EditAgent) }) {
                        Icon(Icons.Default.Edit, "Edit")
                    }
                    IconButton(onClick = { viewModel.onAction(AgentDetailAction.DeleteAgent) }) {
                        Icon(Icons.Default.Delete, "Delete", tint = MaterialTheme.colorScheme.error)
                    }
                }
            )
        },
        floatingActionButton = {
            if (selectedTab == 2) {
                FloatingActionButton(
                    onClick = { onExecute(agentId) }
                ) {
                    Icon(Icons.Default.PlayArrow, "Execute agent")
                }
            }
        }
    ) { padding ->
        Column(modifier = Modifier.padding(padding)) {
            TabRow(selectedTabIndex = selectedTab) {
                tabs.forEachIndexed { index, title ->
                    Tab(
                        selected = selectedTab == index,
                        onClick = { selectedTab = index },
                        text = { Text(title) }
                    )
                }
            }

            when (selectedTab) {
                0 -> AgentOverviewTab(agent = state.agent)
                1 -> AgentConfigTab(
                    agent = state.agent,
                    onConfigChange = { viewModel.onAction(AgentDetailAction.UpdateConfig(it)) }
                )
                2 -> AgentExecutionTab(
                    agentId = agentId,
                    onExecute = { viewModel.onAction(AgentDetailAction.ExecuteAgent) }
                )
                3 -> AgentHistoryTab(
                    executions = state.executions,
                    onLoadMore = { viewModel.onAction(AgentDetailAction.LoadMoreHistory) }
                )
            }
        }
    }
}

@Composable
fun AgentOverviewTab(agent: Agent?) {
    if (agent == null) return

    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // Summary card
        item {
            Card {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Summary", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.height(12.dp))
                    InfoRow("Model", agent.model)
                    InfoRow("Status", agent.status.name)
                    InfoRow("Created", agent.createdAt.toFormattedDate())
                    InfoRow("Last Used", agent.lastUsedAt?.toFormattedDate() ?: "Never")
                    InfoRow("Total Tasks", "${agent.totalTasks}")
                    InfoRow("Success Rate", "${agent.successRate}%")
                    InfoRow("Avg Latency", "${agent.avgLatencyMs}ms")
                }
            }
        }

        // Description
        item {
            Card {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Description", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.height(8.dp))
                    Text(
                        text = agent.description.ifEmpty { "No description" },
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
            }
        }

        // System prompt preview
        item {
            Card {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("System Prompt", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.height(8.dp))
                    Text(
                        text = agent.systemPrompt.take(200) + if (agent.systemPrompt.length > 200) "..." else "",
                        style = MaterialTheme.typography.bodySmall,
                        fontFamily = FontFamily.Monospace,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }

        // Tools bound
        item {
            Card {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Tools (${agent.tools.size})", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.height(8.dp))
                    FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        agent.tools.forEach { tool ->
                            AssistChip(
                                onClick = {},
                                label = { Text(tool.name) },
                                leadingIcon = {
                                    Icon(tool.icon, contentDescription = null, modifier = Modifier.size(16.dp))
                                }
                            )
                        }
                    }
                }
            }
        }

        // Document sets
        item {
            Card {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Document Sets (${agent.documentSets.size})", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.height(8.dp))
                    agent.documentSets.forEach { ds ->
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(Icons.Default.Folder, contentDescription = null, modifier = Modifier.size(16.dp))
                            Spacer(Modifier.width(8.dp))
                            Text(ds.name, style = MaterialTheme.typography.bodyMedium)
                            Spacer(Modifier.weight(1f))
                            Text(ds.permissionLevel.name, style = MaterialTheme.typography.labelSmall)
                        }
                    }
                }
            }
        }

        // SQL connections
        item {
            Card {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Database Connections (${agent.sqlConnections.size})", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.height(8.dp))
                    agent.sqlConnections.forEach { conn ->
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(Icons.Default.Storage, contentDescription = null, modifier = Modifier.size(16.dp))
                            Spacer(Modifier.width(8.dp))
                            Column(modifier = Modifier.weight(1f)) {
                                Text(conn.connectionName, style = MaterialTheme.typography.bodyMedium)
                                Text(
                                    "${conn.databaseType} · ${conn.boundTables.size} tables",
                                    style = MaterialTheme.typography.labelSmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun InfoRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(label, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Text(value, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Medium)
    }
}
```

---

## Agent Creation Screen

Form with all fields for creating a new agent.

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AgentCreationScreen(
    viewModel: AgentCreationViewModel = hiltViewModel(),
    onBack: () -> Unit,
    onAgentCreated: (String) -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Create Agent") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back")
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { padding ->
        LazyColumn(
            modifier = Modifier.padding(padding),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Name
            item {
                OutlinedTextField(
                    value = state.name,
                    onValueChange = { viewModel.onAction(AgentCreationAction.UpdateName(it)) },
                    label = { Text("Agent Name") },
                    modifier = Modifier.fillMaxWidth(),
                    isError = state.nameError != null,
                    supportingText = state.nameError?.let { { Text(it) } },
                    singleLine = true
                )
            }

            // Description
            item {
                OutlinedTextField(
                    value = state.description,
                    onValueChange = { viewModel.onAction(AgentCreationAction.UpdateDescription(it)) },
                    label = { Text("Description") },
                    modifier = Modifier.fillMaxWidth(),
                    minLines = 2,
                    maxLines = 4
                )
            }

            // Model selection
            item {
                Text("Model", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                Spacer(Modifier.height(8.dp))
                ModelSelector(
                    selectedModel = state.selectedModel,
                    models = state.availableModels,
                    onSelect = { viewModel.onAction(AgentCreationAction.SelectModel(it)) }
                )
            }

            // System prompt
            item {
                OutlinedTextField(
                    value = state.systemPrompt,
                    onValueChange = { viewModel.onAction(AgentCreationAction.UpdateSystemPrompt(it)) },
                    label = { Text("System Prompt") },
                    modifier = Modifier.fillMaxWidth().heightIn(min = 120.dp),
                    minLines = 4,
                    maxLines = 10,
                    textStyle = MaterialTheme.typography.bodySmall.copy(fontFamily = FontFamily.Monospace)
                )
            }

            // Temperature
            item {
                Column {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text("Temperature", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                        Text(
                            text = String.format("%.2f", state.temperature),
                            style = MaterialTheme.typography.bodyMedium
                        )
                    }
                    Slider(
                        value = state.temperature,
                        onValueChange = { viewModel.onAction(AgentCreationAction.UpdateTemperature(it)) },
                        valueRange = 0f..2f,
                        steps = 19
                    )
                }
            }

            // Max tokens
            item {
                OutlinedTextField(
                    value = state.maxTokens.toString(),
                    onValueChange = { viewModel.onAction(AgentCreationAction.UpdateMaxTokens(it.toIntOrNull() ?: 4096)) },
                    label = { Text("Max Tokens") },
                    modifier = Modifier.fillMaxWidth(),
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                    singleLine = true
                )
            }

            // Tools selection
            item {
                Text("Tools", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                Spacer(Modifier.height(8.dp))
                FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    state.availableTools.forEach { tool ->
                        val isSelected = tool in state.selectedTools
                        FilterChip(
                            selected = isSelected,
                            onClick = { viewModel.onAction(AgentCreationAction.ToggleTool(tool)) },
                            label = { Text(tool.name) },
                            leadingIcon = if (isSelected) {
                                { Icon(Icons.Default.Check, contentDescription = null, modifier = Modifier.size(16.dp)) }
                            } else null
                        )
                    }
                }
            }

            // Document sets
            item {
                DocumentSetSelector(
                    selectedSets = state.selectedDocumentSets,
                    availableSets = state.availableDocumentSets,
                    onToggle = { viewModel.onAction(AgentCreationAction.ToggleDocumentSet(it)) },
                    onPermissionChange = { set, perm ->
                        viewModel.onAction(AgentCreationAction.UpdateDocSetPermission(set, perm))
                    }
                )
            }

            // SQL connections
            item {
                SqlConnectionSelector(
                    selectedConnections = state.selectedConnections,
                    availableConnections = state.availableConnections,
                    onToggle = { viewModel.onAction(AgentCreationAction.ToggleSqlConnection(it)) }
                )
            }

            // Create button
            item {
                Button(
                    onClick = { viewModel.onAction(AgentCreationAction.Create) },
                    modifier = Modifier.fillMaxWidth(),
                    enabled = state.isValid && !state.isCreating
                ) {
                    if (state.isCreating) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(20.dp),
                            strokeWidth = 2.dp,
                            color = MaterialTheme.colorScheme.onPrimary
                        )
                        Spacer(Modifier.width(8.dp))
                    }
                    Text("Create Agent")
                }
            }
        }
    }
}

@Composable
fun ModelSelector(
    selectedModel: ModelOption?,
    models: List<ModelOption>,
    onSelect: (ModelOption) -> Unit
) {
    var expanded by remember { mutableStateOf(false) }

    ExposedDropdownMenuBox(
        expanded = expanded,
        onExpandedChange = { expanded = it }
    ) {
        OutlinedTextField(
            value = selectedModel?.displayName ?: "Select model",
            onValueChange = {},
            readOnly = true,
            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
            modifier = Modifier.fillMaxWidth().menuAnchor()
        )
        ExposedDropdownMenu(
            expanded = expanded,
            onDismissRequest = { expanded = false }
        ) {
            models.forEach { model ->
                DropdownMenuItem(
                    text = {
                        Column {
                            Text(model.displayName)
                            Text(
                                model.provider,
                                style = MaterialTheme.typography.labelSmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    },
                    onClick = {
                        onSelect(model)
                        expanded = false
                    },
                    leadingIcon = {
                        AsyncImage(
                            model = model.iconUrl,
                            contentDescription = model.provider,
                            modifier = Modifier.size(24.dp)
                        )
                    }
                )
            }
        }
    }
}
```

---

## Agent Configuration Screen

Full configuration with tabs for tools, document sets, and SQL connections.

```kotlin
@Composable
fun AgentConfigTab(
    agent: Agent?,
    onConfigChange: (AgentConfigUpdate) -> Unit
) {
    if (agent == null) return

    var configTab by remember { mutableIntStateOf(0) }
    val configTabs = listOf("Model & Prompt", "Tools", "Documents", "Database")

    Column {
        TabRow(selectedTabIndex = configTab) {
            configTabs.forEachIndexed { index, title ->
                Tab(
                    selected = configTab == index,
                    onClick = { configTab = index },
                    text = { Text(title, style = MaterialTheme.typography.labelMedium) }
                )
            }
        }

        when (configTab) {
            0 -> ModelPromptConfig(agent = agent, onConfigChange = onConfigChange)
            1 -> ToolsConfig(agent = agent, onConfigChange = onConfigChange)
            2 -> DocumentSetsConfig(agent = agent, onConfigChange = onConfigChange)
            3 -> DatabaseConfig(agent = agent, onConfigChange = onConfigChange)
        }
    }
}

@Composable
fun ModelPromptConfig(
    agent: Agent,
    onConfigChange: (AgentConfigUpdate) -> Unit
) {
    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            var model by remember { mutableStateOf(agent.model) }
            ModelSelector(
                selectedModel = ModelOption(agent.model, agent.model, ""),
                models = emptyList(),
                onSelect = { onConfigChange(AgentConfigUpdate.Model(it.displayName)) }
            )
        }
        item {
            var prompt by remember { mutableStateOf(agent.systemPrompt) }
            OutlinedTextField(
                value = prompt,
                onValueChange = { prompt = it },
                label = { Text("System Prompt") },
                modifier = Modifier.fillMaxWidth().heightIn(min = 160.dp),
                minLines = 6,
                textStyle = MaterialTheme.typography.bodySmall.copy(fontFamily = FontFamily.Monospace)
            )
        }
        item {
            var temp by remember { mutableFloatStateOf(agent.temperature) }
            Column {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text("Temperature")
                    Text(String.format("%.2f", temp))
                }
                Slider(
                    value = temp,
                    onValueChange = { temp = it },
                    valueRange = 0f..2f
                )
            }
        }
        item {
            var maxTok by remember { mutableIntStateOf(agent.maxTokens) }
            OutlinedTextField(
                value = maxTok.toString(),
                onValueChange = { maxTok = it.toIntOrNull() ?: 4096 },
                label = { Text("Max Tokens") },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                modifier = Modifier.fillMaxWidth(),
                singleLine = true
            )
        }
        item {
            Button(
                onClick = {
                    onConfigChange(AgentConfigUpdate.Save(
                        model = agent.model,
                        prompt = agent.systemPrompt,
                        temperature = temp,
                        maxTokens = maxTok
                    ))
                },
                modifier = Modifier.fillMaxWidth()
            ) { Text("Save Configuration") }
        }
    }
}
```

---

## Agent Execution Screen

Real-time execution view with step-by-step progress.

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AgentExecutionScreen(
    agentId: String,
    viewModel: AgentExecutionViewModel = hiltViewModel(),
    onBack: () -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Execute Agent") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back")
                    }
                },
                actions = {
                    if (state.isRunning) {
                        IconButton(onClick = { viewModel.onAction(ExecutionAction.Cancel) }) {
                            Icon(Icons.Default.Stop, "Cancel", tint = MaterialTheme.colorScheme.error)
                        }
                    }
                }
            )
        }
    ) { padding ->
        Column(modifier = Modifier.padding(padding)) {
            // Input area
            ExecutionInput(
                query = state.inputQuery,
                onQueryChange = { viewModel.onAction(ExecutionAction.UpdateInput(it)) },
                onExecute = { viewModel.onAction(ExecutionAction.Execute) },
                isRunning = state.isRunning,
                modifier = Modifier.padding(16.dp)
            )

            // Execution steps
            if (state.steps.isNotEmpty()) {
                ExecutionSteps(
                    steps = state.steps,
                    modifier = Modifier.weight(1f)
                )
            }

            // Output area
            if (state.output.isNotEmpty()) {
                ExecutionOutput(
                    output = state.output,
                    modifier = Modifier.padding(16.dp)
                )
            }
        }
    }
}

@Composable
fun ExecutionInput(
    query: String,
    onQueryChange: (String) -> Unit,
    onExecute: () -> Unit,
    isRunning: Boolean,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier,
        verticalAlignment = Alignment.Bottom
    ) {
        OutlinedTextField(
            value = query,
            onValueChange = onQueryChange,
            label = { Text("Enter your query") },
            modifier = Modifier.weight(1f),
            minLines = 2,
            maxLines = 5,
            enabled = !isRunning
        )
        Spacer(Modifier.width(8.dp))
        Button(
            onClick = onExecute,
            enabled = query.isNotBlank() && !isRunning,
            modifier = Modifier.height(56.dp)
        ) {
            if (isRunning) {
                CircularProgressIndicator(
                    modifier = Modifier.size(20.dp),
                    strokeWidth = 2.dp,
                    color = MaterialTheme.colorScheme.onPrimary
                )
            } else {
                Icon(Icons.Default.PlayArrow, "Execute")
            }
        }
    }
}

@Composable
fun ExecutionSteps(
    steps: List<ExecutionStep>,
    modifier: Modifier = Modifier
) {
    LazyColumn(
        modifier = modifier,
        contentPadding = PaddingValues(horizontal = 16.dp)
    ) {
        items(steps, key = { it.id }) { step ->
            ExecutionStepRow(step = step)
        }
    }
}

@Composable
fun ExecutionStepRow(step: ExecutionStep) {
    val (icon, color) = when (step.status) {
        StepStatus.RUNNING -> Icons.Default.Refresh to MaterialTheme.colorScheme.primary
        StepStatus.COMPLETED -> Icons.Default.CheckCircle to Color(0xFF4CAF50)
        StepStatus.FAILED -> Icons.Default.Error to MaterialTheme.colorScheme.error
        StepStatus.PENDING -> Icons.Default.Schedule to MaterialTheme.colorScheme.outline
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.Top
    ) {
        // Step indicator
        Box {
            if (step.status == StepStatus.RUNNING) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    strokeWidth = 2.dp
                )
            } else {
                Icon(
                    imageVector = icon,
                    contentDescription = step.status.name,
                    tint = color,
                    modifier = Modifier.size(24.dp)
                )
            }
        }
        Spacer(Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = step.title,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium
            )
            if (step.description.isNotEmpty()) {
                Text(
                    text = step.description,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            // Tool call details
            if (step.toolCall != null) {
                ToolCallCard(toolCall = step.toolCall)
            }
        }
        Text(
            text = step.timestamp.toFormattedTime(),
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
fun ToolCallCard(toolCall: ToolCall) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(top = 8.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.Build, contentDescription = null, modifier = Modifier.size(14.dp))
                Spacer(Modifier.width(6.dp))
                Text(
                    text = toolCall.name,
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.SemiBold
                )
            }
            Spacer(Modifier.height(4.dp))
            Text(
                text = "Input: ${toolCall.input.take(100)}${if (toolCall.input.length > 100) "..." else ""}",
                style = MaterialTheme.typography.bodySmall,
                fontFamily = FontFamily.Monospace,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            if (toolCall.output != null) {
                Spacer(Modifier.height(4.dp))
                Text(
                    text = "Output: ${toolCall.output.take(200)}${if (toolCall.output.length > 200) "..." else ""}",
                    style = MaterialTheme.typography.bodySmall,
                    fontFamily = FontFamily.Monospace,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

@Composable
fun ExecutionOutput(
    output: String,
    modifier: Modifier = Modifier
) {
    Card(modifier = modifier) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text("Output", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
            Spacer(Modifier.height(8.dp))
            Text(
                text = output,
                style = MaterialTheme.typography.bodyMedium
            )
        }
    }
}
```

---

## Agent Execution History

```kotlin
@Composable
fun AgentHistoryTab(
    executions: List<AgentExecution>,
    onLoadMore: () -> Unit
) {
    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        items(executions, key = { it.executionId }) { execution ->
            ExecutionHistoryItem(execution = execution)
        }

        item {
            LaunchedEffect(Unit) { onLoadMore() }
            CircularProgressIndicator(
                modifier = Modifier.size(24.dp).padding(8.dp)
            )
        }
    }
}

@Composable
fun ExecutionHistoryItem(execution: AgentExecution) {
    val statusColor = when (execution.status) {
        ExecutionStatus.COMPLETED -> Color(0xFF4CAF50)
        ExecutionStatus.FAILED -> MaterialTheme.colorScheme.error
        ExecutionStatus.CANCELLED -> MaterialTheme.colorScheme.outline
        ExecutionStatus.RUNNING -> MaterialTheme.colorScheme.primary
    }

    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = execution.query.take(60),
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.Medium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f)
                )
                Surface(
                    shape = RoundedCornerShape(12.dp),
                    color = statusColor.copy(alpha = 0.12f)
                ) {
                    Text(
                        text = execution.status.name,
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp),
                        style = MaterialTheme.typography.labelSmall,
                        color = statusColor
                    )
                }
            }
            Spacer(Modifier.height(8.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Text("Duration: ${execution.durationMs}ms", style = MaterialTheme.typography.labelSmall)
                Text("Steps: ${execution.stepCount}", style = MaterialTheme.typography.labelSmall)
                Text("Tokens: ${execution.tokensUsed}", style = MaterialTheme.typography.labelSmall)
            }
            Text(
                text = execution.startedAt.toFormattedDateTime(),
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}
```

---

## Agent Performance Metrics

```kotlin
@Composable
fun AgentPerformanceMetrics(agent: Agent, modifier: Modifier = Modifier) {
    LazyColumn(
        modifier = modifier,
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            MetricsSummaryRow(
                metrics = listOf(
                    MetricItem("Total Executions", "${agent.totalTasks}", Icons.Default.PlayArrow),
                    MetricItem("Success Rate", "${agent.successRate}%", Icons.Default.CheckCircle),
                    MetricItem("Avg Latency", "${agent.avgLatencyMs}ms", Icons.Default.Speed),
                    MetricItem("Total Tokens", formatTokenCount(agent.totalTokensUsed), Icons.Default.DataUsage)
                )
            )
        }
        item {
            AgentLatencyChart(agentId = agent.agentId)
        }
        item {
            AgentErrorRateChart(agentId = agent.agentId)
        }
        item {
            AgentTokenUsageChart(agentId = agent.agentId)
        }
    }
}

@Composable
fun MetricsSummaryRow(metrics: List<MetricItem>) {
    LazyRow(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        items(metrics) { metric ->
            Card(modifier = Modifier.width(140.dp)) {
                Column(modifier = Modifier.padding(12.dp)) {
                    Icon(
                        metric.icon, contentDescription = null,
                        modifier = Modifier.size(20.dp),
                        tint = MaterialTheme.colorScheme.primary
                    )
                    Spacer(Modifier.height(8.dp))
                    Text(metric.value, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Text(metric.label, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
            }
        }
    }
}
```

---

## Agent Settings

```kotlin
data class AgentSettings(
    val systemPrompt: String = "",
    val temperature: Float = 0.7f,
    val maxTokens: Int = 4096,
    val topP: Float = 1f,
    val frequencyPenalty: Float = 0f,
    val presencePenalty: Float = 0f,
    val stopSequences: List<String> = emptyList(),
    val enableStreaming: Boolean = true,
    val enableFunctionCalling: Boolean = true,
    val maxIterations: Int = 10
)

@Composable
fun AgentSettingsForm(
    settings: AgentSettings,
    onSettingsChange: (AgentSettings) -> Unit,
    modifier: Modifier = Modifier
) {
    LazyColumn(
        modifier = modifier,
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            OutlinedTextField(
                value = settings.systemPrompt,
                onValueChange = { onSettingsChange(settings.copy(systemPrompt = it)) },
                label = { Text("System Prompt") },
                modifier = Modifier.fillMaxWidth().heightIn(min = 120.dp),
                minLines = 4
            )
        }
        item {
            SliderSetting("Temperature", settings.temperature, 0f..2f) {
                onSettingsChange(settings.copy(temperature = it))
            }
        }
        item {
            SliderSetting("Top P", settings.topP, 0f..1f) {
                onSettingsChange(settings.copy(topP = it))
            }
        }
        item {
            SliderSetting("Frequency Penalty", settings.frequencyPenalty, -2f..2f) {
                onSettingsChange(settings.copy(frequencyPenalty = it))
            }
        }
        item {
            SliderSetting("Presence Penalty", settings.presencePenalty, -2f..2f) {
                onSettingsChange(settings.copy(presencePenalty = it))
            }
        }
        item {
            OutlinedTextField(
                value = settings.maxTokens.toString(),
                onValueChange = { onSettingsChange(settings.copy(maxTokens = it.toIntOrNull() ?: 4096)) },
                label = { Text("Max Tokens") },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            OutlinedTextField(
                value = settings.maxIterations.toString(),
                onValueChange = { onSettingsChange(settings.copy(maxIterations = it.toIntOrNull() ?: 10)) },
                label = { Text("Max Iterations") },
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("Streaming", style = MaterialTheme.typography.bodyMedium)
                Switch(
                    checked = settings.enableStreaming,
                    onCheckedChange = { onSettingsChange(settings.copy(enableStreaming = it)) }
                )
            }
        }
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("Function Calling", style = MaterialTheme.typography.bodyMedium)
                Switch(
                    checked = settings.enableFunctionCalling,
                    onCheckedChange = { onSettingsChange(settings.copy(enableFunctionCalling = it)) }
                )
            }
        }
    }
}

@Composable
fun SliderSetting(label: String, value: Float, range: ClosedFloatingPointRange<Float>, onValueChange: (Float) -> Unit) {
    Column {
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            Text(label, style = MaterialTheme.typography.bodyMedium)
            Text(String.format("%.2f", value), style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.SemiBold)
        }
        Slider(value = value, onValueChange = onValueChange, valueRange = range)
    }
}
```

---

## Agent Tools Configuration

```kotlin
@Composable
fun ToolsConfig(
    agent: Agent,
    onConfigChange: (AgentConfigUpdate) -> Unit
) {
    val (selectedTools, setSelectedTools) = remember { mutableStateOf(agent.tools.toSet()) }

    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item {
            Text(
                "Available Tools",
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold
            )
            Spacer(Modifier.height(4.dp))
            Text(
                "Select tools the agent can use during execution",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        items(AvailableTools.all, key = { it.id }) { tool ->
            val isSelected = tool in selectedTools
            Card(
                onClick = {
                    val newSet = if (isSelected) selectedTools - tool else selectedTools + tool
                    setSelectedTools(newSet)
                },
                colors = CardDefaults.cardColors(
                    containerColor = if (isSelected)
                        MaterialTheme.colorScheme.primaryContainer
                    else MaterialTheme.colorScheme.surface
                )
            ) {
                Row(
                    modifier = Modifier.padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Checkbox(
                        checked = isSelected,
                        onCheckedChange = { checked ->
                            val newSet = if (checked) selectedTools + tool else selectedTools - tool
                            setSelectedTools(newSet)
                        }
                    )
                    Spacer(Modifier.width(12.dp))
                    Icon(tool.icon, contentDescription = null, modifier = Modifier.size(24.dp))
                    Spacer(Modifier.width(12.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(tool.name, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Medium)
                        Text(tool.description, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
        }

        item {
            Button(
                onClick = { onConfigChange(AgentConfigUpdate.Tools(selectedTools.toList())) },
                modifier = Modifier.fillMaxWidth()
            ) { Text("Save Tool Configuration") }
        }
    }
}

object AvailableTools {
    val all = listOf(
        ToolOption("web_search", "Web Search", "Search the web for information", Icons.Default.Search),
        ToolOption("calculator", "Calculator", "Perform calculations", Icons.Default.Calculate),
        ToolOption("code_interpreter", "Code Interpreter", "Execute code snippets", Icons.Default.Code),
        ToolOption("file_reader", "File Reader", "Read uploaded files", Icons.Default.FolderOpen),
        ToolOption("database_query", "Database Query", "Execute SQL queries", Icons.Default.Storage),
        ToolOption("api_caller", "API Caller", "Make external API calls", Icons.Default.Cloud),
        ToolOption("image_gen", "Image Generator", "Generate images from text", Icons.Default.Image),
        ToolOption("email_sender", "Email Sender", "Send emails", Icons.Default.Email)
    )
}
```

---

## Agent-Document Set Binding

```kotlin
@Composable
fun DocumentSetSelector(
    selectedSets: Map<String, PermissionLevel>,
    availableSets: List<DocumentSetInfo>,
    onToggle: (DocumentSetInfo) -> Unit,
    onPermissionChange: (DocumentSetInfo, PermissionLevel) -> Unit
) {
    Column {
        Text("Document Sets", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
        Spacer(Modifier.height(8.dp))

        availableSets.forEach { ds ->
            val isSelected = ds.id in selectedSets
            val permission = selectedSets[ds.id] ?: PermissionLevel.READ

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { onToggle(ds) }
                    .padding(vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Checkbox(
                    checked = isSelected,
                    onCheckedChange = { onToggle(ds) }
                )
                Spacer(Modifier.width(8.dp))
                Icon(Icons.Default.Folder, contentDescription = null, modifier = Modifier.size(20.dp))
                Spacer(Modifier.width(8.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(ds.name, style = MaterialTheme.typography.bodyMedium)
                    Text("${ds.documentCount} documents", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                if (isSelected) {
                    PermissionDropdown(
                        selected = permission,
                        onSelect = { onPermissionChange(ds, it) }
                    )
                }
            }
        }
    }
}

enum class PermissionLevel { READ, WRITE, ADMIN }

@Composable
fun PermissionDropdown(
    selected: PermissionLevel,
    onSelect: (PermissionLevel) -> Unit
) {
    var expanded by remember { mutableStateOf(false) }

    ExposedDropdownMenuBox(expanded = expanded, onExpandedChange = { expanded = it }) {
        AssistChip(
            onClick = { expanded = true },
            label = { Text(selected.name, style = MaterialTheme.typography.labelSmall) },
            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
            modifier = Modifier.menuAnchor()
        )
        ExposedDropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
            PermissionLevel.entries.forEach { level ->
                DropdownMenuItem(
                    text = { Text(level.name) },
                    onClick = { onSelect(level); expanded = false }
                )
            }
        }
    }
}
```

---

## Agent-Database Connection

```kotlin
@Composable
fun SqlConnectionSelector(
    selectedConnections: Set<String>,
    availableConnections: List<SqlConnectionInfo>,
    onToggle: (SqlConnectionInfo) -> Unit
) {
    Column {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text("Database Connections", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            TextButton(onClick = { /* navigate to add connection */ }) {
                Icon(Icons.Default.Add, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(Modifier.width(4.dp))
                Text("Add")
            }
        }

        availableConnections.forEach { conn ->
            val isSelected = conn.id in selectedConnections
            Card(
                onClick = { onToggle(conn) },
                colors = CardDefaults.cardColors(
                    containerColor = if (isSelected) MaterialTheme.colorScheme.primaryContainer
                    else MaterialTheme.colorScheme.surface
                )
            ) {
                Row(
                    modifier = Modifier.padding(12.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Checkbox(checked = isSelected, onCheckedChange = { onToggle(conn) })
                    Spacer(Modifier.width(8.dp))
                    Icon(Icons.Default.Storage, contentDescription = null)
                    Spacer(Modifier.width(8.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(conn.connectionName, style = MaterialTheme.typography.bodyMedium)
                        Text(
                            "${conn.databaseType} · ${conn.host}:${conn.port} · ${conn.boundTables.size} tables",
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                    StatusDot(isOnline = conn.isConnected)
                }
            }
        }
    }
}
```

---

## Agent Test Connection Flow

```kotlin
@Composable
fun TestConnectionButton(
    connectionId: String,
    viewModel: AgentDetailViewModel
) {
    val testState by viewModel.testConnectionState.collectAsStateWithLifecycle()

    Column {
        Button(
            onClick = { viewModel.onAction(AgentDetailAction.TestConnection(connectionId)) },
            enabled = testState !is ConnectionState.Testing
        ) {
            if (testState is ConnectionState.Testing) {
                CircularProgressIndicator(modifier = Modifier.size(16.dp), strokeWidth = 2.dp)
                Spacer(Modifier.width(8.dp))
                Text("Testing...")
            } else {
                Icon(Icons.Default.WifiFind, contentDescription = null, modifier = Modifier.size(16.dp))
                Spacer(Modifier.width(8.dp))
                Text("Test Connection")
            }
        }

        when (val state = testState) {
            is ConnectionState.Success -> {
                Card(
                    colors = CardDefaults.cardColors(containerColor = Color(0xFF4CAF50).copy(alpha = 0.1f))
                ) {
                    Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Icon(Icons.Default.CheckCircle, contentDescription = null, tint = Color(0xFF4CAF50))
                        Spacer(Modifier.width(8.dp))
                        Column {
                            Text("Connection successful", style = MaterialTheme.typography.bodyMedium)
                            Text("Latency: ${state.latencyMs}ms", style = MaterialTheme.typography.labelSmall)
                        }
                    }
                }
            }
            is ConnectionState.Failed -> {
                Card(
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.error.copy(alpha = 0.1f))
                ) {
                    Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Icon(Icons.Default.Error, contentDescription = null, tint = MaterialTheme.colorScheme.error)
                        Spacer(Modifier.width(8.dp))
                        Column {
                            Text("Connection failed", style = MaterialTheme.typography.bodyMedium)
                            Text(state.message, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.error)
                        }
                    }
                }
            }
            else -> {}
        }
    }
}

sealed interface ConnectionState {
    data object Idle : ConnectionState
    data object Testing : ConnectionState
    data class Success(val latencyMs: Long, val version: String) : ConnectionState
    data class Failed(val message: String) : ConnectionState
}
```

---

## Agent Schema Discovery

```kotlin
@Composable
fun SchemaDiscoveryView(
    connectionId: String,
    viewModel: AgentDetailViewModel
) {
    val schema by viewModel.schemaState.collectAsStateWithLifecycle()

    Column(modifier = Modifier.padding(16.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text("Schema Discovery", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
            Button(onClick = { viewModel.onAction(AgentDetailAction.DiscoverSchema(connectionId)) }) {
                Text("Discover")
            }
        }

        when (val s = schema) {
            is SchemaState.Loading -> {
                CircularProgressIndicator(modifier = Modifier.padding(32.dp))
            }
            is SchemaState.Success -> {
                LazyColumn(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    items(s.tables, key = { it.name }) { table ->
                        SchemaTableCard(table = table)
                    }
                }
            }
            is SchemaState.Error -> {
                Text(s.message, color = MaterialTheme.colorScheme.error)
            }
            else -> {}
        }
    }
}

@Composable
fun SchemaTableCard(table: TableInfo) {
    var expanded by remember { mutableStateOf(false) }

    Card(
        onClick = { expanded = !expanded },
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.TableChart, contentDescription = null, modifier = Modifier.size(20.dp))
                    Spacer(Modifier.width(8.dp))
                    Text(table.name, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.SemiBold)
                }
                Column(horizontalAlignment = Alignment.End) {
                    Text("${table.columns.size} columns", style = MaterialTheme.typography.labelSmall)
                    Text("~${table.estimatedRows} rows", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
            }

            if (expanded) {
                Spacer(Modifier.height(12.dp))
                HorizontalDivider()
                Spacer(Modifier.height(8.dp))

                table.columns.forEach { col ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 2.dp),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(col.name, style = MaterialTheme.typography.bodySmall, fontFamily = FontFamily.Monospace)
                        Text(col.type, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
        }
    }
}
```

---

## Agent Table Binding

```kotlin
@Composable
fun TableBindingSection(
    connectionId: String,
    tables: List<TableInfo>,
    boundTables: Map<String, TableBinding>,
    onBind: (String, TableBinding) -> Unit,
    onUnbind: (String) -> Unit
) {
    Column(modifier = Modifier.padding(16.dp)) {
        Text("Table Binding", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(8.dp))

        tables.forEach { table ->
            val binding = boundTables[table.name]
            val isBound = binding != null

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(vertical = 4.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Checkbox(
                    checked = isBound,
                    onCheckedChange = { checked ->
                        if (checked) {
                            onBind(table.name, TableBinding(table.name, ColumnPermission.ALL, true))
                        } else {
                            onUnbind(table.name)
                        }
                    }
                )
                Column(modifier = Modifier.weight(1f)) {
                    Text(table.name, style = MaterialTheme.typography.bodyMedium)
                }
                if (isBound) {
                    AssistChip(
                        onClick = { /* open column permission editor */ },
                        label = { Text(binding!!.columnPermission.name) },
                        modifier = Modifier.height(28.dp)
                    )
                }
            }
        }
    }
}

data class TableBinding(
    val tableName: String,
    val columnPermission: ColumnPermission,
    val enabled: Boolean
)

enum class ColumnPermission { ALL, READ_ONLY, SELECTED }
```

---

## Agent Execution Workflow

```
┌────────────────────────────────────────────────────────────────────┐
│                    Agent Execution Flow                            │
│                                                                    │
│  User Query ──► Validate ──► Load Agent Config ──► Build Context   │
│      │                                                   │        │
│      │              ┌────────────────────────────────────┘        │
│      │              ▼                                              │
│      │         LLM API Call (streaming)                           │
│      │              │                                              │
│      │         ┌────┴────┐                                         │
│      │         ▼         ▼                                         │
│      │    Text Response  Tool Call                                │
│      │         │         │                                        │
│      │         │    ┌────┘                                         │
│      │         │    ▼                                              │
│      │         │  Execute Tool                                     │
│      │         │    │                                              │
│      │         │    ▼                                              │
│      │         │  Feed Result Back to LLM                         │
│      │         │    │                                              │
│      │         └────┘                                             │
│      │              │                                              │
│      │         Final Response                                     │
│      │              │                                              │
│      ▼              ▼                                              │
│  Display Response + Update Metrics                                │
└────────────────────────────────────────────────────────────────────┘
```

```kotlin
class AgentExecutionViewModel @Inject constructor(
    private val agentRepository: AgentRepository,
    private val executionRepository: ExecutionRepository,
    @IoDispatcher private val ioDispatcher: CoroutineDispatcher
) : ViewModel() {

    private val _state = MutableStateFlow(ExecutionState())
    val state: StateFlow<ExecutionState> = _state.asStateFlow()

    private var executionJob: Job? = null

    fun onAction(action: ExecutionAction) {
        when (action) {
            is ExecutionAction.UpdateInput -> {
                _state.update { it.copy(inputQuery = action.query) }
            }
            is ExecutionAction.Execute -> executeQuery()
            is ExecutionAction.Cancel -> cancelExecution()
        }
    }

    private fun executeQuery() {
        val query = _state.value.inputQuery.trim()
        if (query.isBlank()) return

        executionJob = viewModelScope.launch(ioDispatcher) {
            _state.update {
                it.copy(
                    isRunning = true,
                    steps = listOf(ExecutionStep(
                        id = "init",
                        title = "Initializing",
                        description = "Validating query and loading agent configuration",
                        status = StepStatus.RUNNING,
                        timestamp = Instant.now()
                    )),
                    output = ""
                )
            }

            try {
                val flow = agentRepository.executeAgent(
                    agentId = _state.value.agentId,
                    query = query
                )

                flow.collect { event ->
                    when (event) {
                        is ExecutionEvent.StepStarted -> {
                            _state.update { state ->
                                state.copy(
                                    steps = state.steps + ExecutionStep(
                                        id = event.stepId,
                                        title = event.title,
                                        description = event.description,
                                        status = StepStatus.RUNNING,
                                        timestamp = Instant.now()
                                    )
                                )
                            }
                        }
                        is ExecutionEvent.StepCompleted -> {
                            _state.update { state ->
                                state.copy(
                                    steps = state.steps.map {
                                        if (it.id == event.stepId) it.copy(
                                            status = StepStatus.COMPLETED,
                                            description = event.result
                                        ) else it
                                    }
                                )
                            }
                        }
                        is ExecutionEvent.ToolCallStarted -> {
                            _state.update { state ->
                                state.copy(
                                    steps = state.steps.map {
                                        if (it.id == event.stepId) it.copy(
                                            toolCall = ToolCall(
                                                name = event.toolName,
                                                input = event.input
                                            )
                                        ) else it
                                    }
                                )
                            }
                        }
                        is ExecutionEvent.ToolCallCompleted -> {
                            _state.update { state ->
                                state.copy(
                                    steps = state.steps.map {
                                        if (it.id == event.stepId) it.copy(
                                            toolCall = it.toolCall?.copy(output = event.output)
                                        ) else it
                                    }
                                )
                            }
                        }
                        is ExecutionEvent.TextChunk -> {
                            _state.update { it.copy(output = it.output + event.text) }
                        }
                        is ExecutionEvent.Completed -> {
                            _state.update { it.copy(isRunning = false) }
                        }
                        is ExecutionEvent.Error -> {
                            _state.update { state ->
                                state.copy(
                                    isRunning = false,
                                    steps = state.steps.map {
                                        if (it.id == event.stepId) it.copy(
                                            status = StepStatus.FAILED,
                                            description = event.message
                                        ) else it
                                    }
                                )
                            }
                        }
                    }
                }
            } catch (e: CancellationException) {
                _state.update {
                    it.copy(
                        isRunning = false,
                        steps = it.steps.map {
                            if (it.status == StepStatus.RUNNING) it.copy(status = StepStatus.CANCELLED) else it
                        }
                    )
                }
            } catch (e: Exception) {
                _state.update {
                    it.copy(
                        isRunning = false,
                        output = "Error: ${e.message}"
                    )
                }
            }
        }
    }

    private fun cancelExecution() {
        executionJob?.cancel()
    }
}
```

---

## Agent Execution Visualization

```kotlin
@Composable
fun ExecutionVisualization(
    steps: List<ExecutionStep>,
    modifier: Modifier = Modifier
) {
    LazyColumn(
        modifier = modifier,
        contentPadding = PaddingValues(16.dp)
    ) {
        itemsIndexed(steps) { index, step ->
            Row(modifier = Modifier.fillMaxWidth()) {
                // Timeline line
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    modifier = Modifier.width(32.dp)
                ) {
                    StepIcon(step.status)
                    if (index < steps.lastIndex) {
                        Box(
                            modifier = Modifier
                                .width(2.dp)
                                .height(40.dp)
                                .background(
                                    if (step.status == StepStatus.COMPLETED)
                                        Color(0xFF4CAF50)
                                    else MaterialTheme.colorScheme.outline.copy(alpha = 0.3f)
                                )
                        )
                    }
                }
                // Step content
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .padding(start = 8.dp, bottom = 16.dp)
                ) {
                    Text(
                        text = step.title,
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.Medium
                    )
                    if (step.description.isNotEmpty()) {
                        Text(
                            text = step.description,
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                    if (step.toolCall != null) {
                        ExecutionToolCallCard(toolCall = step.toolCall!!)
                    }
                }
            }
        }
    }
}

@Composable
fun StepIcon(status: StepStatus) {
    val (icon, color) = when (status) {
        StepStatus.RUNNING -> Icons.Default.Refresh to MaterialTheme.colorScheme.primary
        StepStatus.COMPLETED -> Icons.Default.CheckCircle to Color(0xFF4CAF50)
        StepStatus.FAILED -> Icons.Default.Cancel to MaterialTheme.colorScheme.error
        StepStatus.PENDING -> Icons.Default.Schedule to MaterialTheme.colorScheme.outline
    }
    Box(
        modifier = Modifier
            .size(32.dp)
            .clip(CircleShape)
            .background(color.copy(alpha = 0.12f)),
        contentAlignment = Alignment.Center
    ) {
        if (status == StepStatus.RUNNING) {
            CircularProgressIndicator(modifier = Modifier.size(20.dp), strokeWidth = 2.dp)
        } else {
            Icon(icon, contentDescription = status.name, tint = color, modifier = Modifier.size(18.dp))
        }
    }
}
```

---

## Agent Error Handling

```kotlin
sealed interface AgentError {
    data class CreationFailed(val message: String) : AgentError
    data class ExecutionFailed(val message: String, val stepId: String?) : AgentError
    data class ConnectionFailed(val connectionId: String, val message: String) : AgentError
    data object NotFound : AgentError
    data object PermissionDenied : AgentError
}

@Composable
fun AgentErrorBanner(
    error: AgentError,
    onDismiss: () -> Unit,
    onRetry: (() -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier.fillMaxWidth(),
        color = MaterialTheme.colorScheme.errorContainer,
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                Icons.Default.Error,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.onErrorContainer
            )
            Spacer(Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = when (error) {
                        is AgentError.CreationFailed -> "Creation Failed"
                        is AgentError.ExecutionFailed -> "Execution Failed"
                        is AgentError.ConnectionFailed -> "Connection Failed"
                        is AgentError.NotFound -> "Agent Not Found"
                        is AgentError.PermissionDenied -> "Permission Denied"
                    },
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.Medium,
                    color = MaterialTheme.colorScheme.onErrorContainer
                )
                Text(
                    text = when (error) {
                        is AgentError.CreationFailed -> error.message
                        is AgentError.ExecutionFailed -> error.message
                        is AgentError.ConnectionFailed -> error.message
                        is AgentError.NotFound -> "The requested agent could not be found"
                        is AgentError.PermissionDenied -> "You don't have permission to perform this action"
                    },
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onErrorContainer.copy(alpha = 0.8f)
                )
            }
            if (onRetry != null) {
                TextButton(onClick = onRetry) { Text("Retry") }
            }
            IconButton(onClick = onDismiss) {
                Icon(Icons.Default.Close, "Dismiss", tint = MaterialTheme.colorScheme.onErrorContainer)
            }
        }
    }
}
```

### Execution Error Recovery

```kotlin
when (val status = step.status) {
    StepStatus.FAILED -> {
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            OutlinedButton(onClick = { onRetryStep(step.id) }) {
                Icon(Icons.Default.Refresh, contentDescription = null, modifier = Modifier.size(14.dp))
                Spacer(Modifier.width(4.dp))
                Text("Retry", style = MaterialTheme.typography.labelSmall)
            }
            OutlinedButton(onClick = { onSkipStep(step.id) }) {
                Text("Skip", style = MaterialTheme.typography.labelSmall)
            }
        }
    }
}
```

---

## Agent API Integration

```kotlin
interface AgentApiService {

    @GET("api/v1/agents")
    suspend fun getAgents(
        @Query("search") search: String? = null,
        @Query("status") status: String? = null,
        @Query("sort") sort: String = "name",
        @Query("page") page: Int = 1,
        @Query("limit") limit: Int = 20
    ): PaginatedResponse<AgentDto>

    @GET("api/v1/agents/{id}")
    suspend fun getAgent(@Path("id") agentId: String): AgentDto

    @POST("api/v1/agents")
    suspend fun createAgent(@Body request: CreateAgentRequest): AgentDto

    @PUT("api/v1/agents/{id}")
    suspend fun updateAgent(@Path("id") agentId: String, @Body request: UpdateAgentRequest): AgentDto

    @DELETE("api/v1/agents/{id}")
    suspend fun deleteAgent(@Path("id") agentId: String)

    @POST("api/v1/agents/{id}/execute")
    suspend fun executeAgent(@Path("id") agentId: String, @Body request: ExecuteAgentRequest): ResponseBody

    @POST("api/v1/agents/{id}/document-sets")
    suspend fun bindDocumentSets(@Path("id") agentId: String, @Body request: BindDocumentSetsRequest): AgentDto

    @POST("api/v1/agents/{id}/sql-connections")
    suspend fun bindSqlConnections(@Path("id") agentId: String, @Body request: BindSqlConnectionsRequest): AgentDto

    @GET("api/v1/agents/{id}/executions")
    suspend fun getExecutionHistory(
        @Path("id") agentId: String,
        @Query("page") page: Int = 1,
        @Query("limit") limit: Int = 20
    ): PaginatedResponse<ExecutionDto>

    @POST("api/v1/sql-connections/{id}/test")
    suspend fun testConnection(@Path("id") connectionId: String): TestConnectionResponse

    @POST("api/v1/sql-connections/{id}/discover")
    suspend fun discoverSchema(@Path("id") connectionId: String): SchemaResponse
}
```

### Request/Response DTOs

```kotlin
@Serializable
data class CreateAgentRequest(
    val name: String,
    val description: String,
    val model: String,
    val systemPrompt: String,
    val temperature: Float,
    val maxTokens: Int,
    val toolIds: List<String>,
    val documentSetIds: List<String>,
    val sqlConnectionIds: List<String>
)

@Serializable
data class AgentDto(
    val id: String,
    val name: String,
    val description: String,
    val model: String,
    val systemPrompt: String,
    val temperature: Float,
    val maxTokens: Int,
    val status: String,
    val tools: List<ToolDto>,
    val documentSets: List<DocumentSetBindingDto>,
    val sqlConnections: List<SqlConnectionBindingDto>,
    val totalTasks: Int,
    val successRate: Float,
    val avgLatencyMs: Int,
    val createdAt: String,
    val lastUsedAt: String?
)

@Serializable
data class ExecuteAgentRequest(
    val query: String,
    val context: Map<String, String>? = null
)

@Serializable
data class TestConnectionResponse(
    val success: Boolean,
    val latencyMs: Long,
    val version: String?,
    val errorMessage: String? = null
)

@Serializable
data class SchemaResponse(
    val tables: List<TableDto>
)

@Serializable
data class TableDto(
    val name: String,
    val columns: List<ColumnDto>,
    val estimatedRows: Long
)

@Serializable
data class ColumnDto(
    val name: String,
    val type: String,
    val nullable: Boolean,
    val isPrimaryKey: Boolean
)
```

---

## Agent Data Models

```kotlin
data class AgentSummary(
    val agentId: String,
    val name: String,
    val description: String,
    val model: String,
    val status: AgentStatusType,
    val totalTasks: Int,
    val successRate: Float,
    val avgLatencyMs: Int,
    val createdAt: Instant,
    val lastUsedAt: Instant?
)

data class Agent(
    val agentId: String,
    val name: String,
    val description: String,
    val model: String,
    val systemPrompt: String,
    val temperature: Float,
    val maxTokens: Int,
    val status: AgentStatusType,
    val tools: List<AgentTool>,
    val documentSets: List<DocumentSetBinding>,
    val sqlConnections: List<SqlConnectionBinding>,
    val settings: AgentSettings,
    val totalTasks: Int,
    val successRate: Float,
    val avgLatencyMs: Int,
    val totalTokensUsed: Long,
    val createdAt: Instant,
    val lastUsedAt: Instant?
)

enum class AgentStatusType { ONLINE, OFFLINE, RUNNING, ERROR }

data class AgentTool(
    val id: String,
    val name: String,
    val description: String,
    val icon: ImageVector
)

data class DocumentSetBinding(
    val id: String,
    val name: String,
    val documentCount: Int,
    val permissionLevel: PermissionLevel
)

data class SqlConnectionBinding(
    val id: String,
    val connectionName: String,
    val databaseType: String,
    val host: String,
    val port: Int,
    val isConnected: Boolean,
    val boundTables: List<String>
)

data class AgentExecution(
    val executionId: String,
    val agentId: String,
    val query: String,
    val status: ExecutionStatus,
    val durationMs: Long,
    val stepCount: Int,
    val tokensUsed: Int,
    val startedAt: Instant,
    val completedAt: Instant?
)

enum class ExecutionStatus { RUNNING, COMPLETED, FAILED, CANCELLED }

data class ExecutionStep(
    val id: String,
    val title: String,
    val description: String,
    val status: StepStatus,
    val timestamp: Instant,
    val toolCall: ToolCall? = null
)

enum class StepStatus { PENDING, RUNNING, COMPLETED, FAILED }

data class ToolCall(
    val name: String,
    val input: String,
    val output: String? = null
)

data class ModelOption(
    val id: String,
    val displayName: String,
    val provider: String,
    val iconUrl: String? = null
)
```

---

## Agent Caching

```kotlin
@Entity(tableName = "agents")
data class AgentCacheEntity(
    @PrimaryKey val agentId: String,
    val name: String,
    val description: String,
    val model: String,
    val systemPrompt: String,
    val temperature: Float,
    val maxTokens: Int,
    val status: String,
    val toolsJson: String,
    val documentSetsJson: String,
    val sqlConnectionsJson: String,
    val totalTasks: Int,
    val successRate: Float,
    val avgLatencyMs: Int,
    val lastUpdated: Long
)

@Dao
interface AgentDao {

    @Query("SELECT * FROM agents ORDER BY name ASC")
    fun getAllAgents(): Flow<List<AgentCacheEntity>>

    @Query("SELECT * FROM agents WHERE agentId = :agentId")
    suspend fun getAgent(agentId: String): AgentCacheEntity?

    @Upsert
    suspend fun upsertAll(agents: List<AgentCacheEntity>)

    @Upsert
    suspend fun upsert(agent: AgentCacheEntity)

    @Query("DELETE FROM agents WHERE agentId = :agentId")
    suspend fun delete(agentId: String)

    @Query("DELETE FROM agents")
    suspend fun deleteAll()

    @Query("SELECT * FROM agents WHERE name LIKE '%' || :query || '%' OR model LIKE '%' || :query || '%'")
    suspend fun search(query: String): List<AgentCacheEntity>
}
```

---

## Pull-to-Refresh

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AgentListPullToRefresh(
    isRefreshing: Boolean,
    onRefresh: () -> Unit,
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit
) {
    PullToRefreshBox(
        isRefreshing = isRefreshing,
        onRefresh = onRefresh,
        modifier = modifier
    ) {
        content()
    }
}
```

---

## Search and Filter

```kotlin
@Composable
fun AgentSearchBar(
    query: String,
    onQueryChange: (String) -> Unit,
    onFilterClick: () -> Unit,
    onSortClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier.padding(horizontal = 16.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        OutlinedTextField(
            value = query,
            onValueChange = onQueryChange,
            placeholder = { Text("Search agents...") },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = "Search") },
            trailingIcon = {
                if (query.isNotEmpty()) {
                    IconButton(onClick = { onQueryChange("") }) {
                        Icon(Icons.Default.Clear, "Clear")
                    }
                }
            },
            modifier = Modifier.weight(1f),
            singleLine = true
        )
        Spacer(Modifier.width(8.dp))
        IconButton(onClick = onFilterClick) {
            Icon(Icons.Default.FilterList, "Filters")
        }
        IconButton(onClick = onSortClick) {
            Icon(Icons.Default.Sort, "Sort")
        }
    }
}
```

---

## Empty States

```kotlin
@Composable
fun AgentListEmpty(onCreateAgent: () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            Icons.Default.SmartToy,
            contentDescription = null,
            modifier = Modifier.size(80.dp),
            tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.4f)
        )
        Spacer(Modifier.height(16.dp))
        Text("No agents yet", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
        Spacer(Modifier.height(8.dp))
        Text(
            "Create your first AI agent to get started",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
        Spacer(Modifier.height(24.dp))
        Button(onClick = onCreateAgent) {
            Icon(Icons.Default.Add, contentDescription = null)
            Spacer(Modifier.width(8.dp))
            Text("Create Agent")
        }
    }
}

@Composable
fun AgentListError(error: String, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            Icons.Default.ErrorOutline,
            contentDescription = null,
            modifier = Modifier.size(64.dp),
            tint = MaterialTheme.colorScheme.error
        )
        Spacer(Modifier.height(16.dp))
        Text("Failed to load agents", style = MaterialTheme.typography.titleMedium)
        Spacer(Modifier.height(4.dp))
        Text(error, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Spacer(Modifier.height(16.dp))
        Button(onClick = onRetry) { Text("Retry") }
    }
}
```

---

## Loading States

```kotlin
@Composable
fun AgentListSkeleton() {
    val shimmer = shimmerBrush()

    LazyColumn(contentPadding = PaddingValues(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        items(6) {
            Card(modifier = Modifier.fillMaxWidth()) {
                Row(
                    modifier = Modifier.padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Box(modifier = Modifier.size(48.dp).clip(CircleShape).background(shimmer))
                    Spacer(Modifier.width(16.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Box(modifier = Modifier.width(120.dp).height(16.dp).background(shimmer, RoundedCornerShape(4.dp)))
                        Spacer(Modifier.height(8.dp))
                        Box(modifier = Modifier.width(80.dp).height(12.dp).background(shimmer, RoundedCornerShape(4.dp)))
                        Spacer(Modifier.height(8.dp))
                        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                            Box(modifier = Modifier.width(60.dp).height(10.dp).background(shimmer, RoundedCornerShape(4.dp)))
                            Box(modifier = Modifier.width(50.dp).height(10.dp).background(shimmer, RoundedCornerShape(4.dp)))
                        }
                    }
                }
            }
        }
    }
}
```

---

## Accessibility

```kotlin
// Agent card
Card(
    modifier = Modifier.semantics(mergeDescendants = true) {
        contentDescription = "Agent ${agent.name}, model ${agent.model}, status ${agent.status.name}, ${agent.totalTasks} tasks, ${agent.successRate}% success rate"
    }
)

// Execution step
Row(modifier = Modifier.semantics {
    stateDescription = "${step.title}: ${step.status.name}"
})

// Test connection
Button(
    onClick = { ... },
    modifier = Modifier.semantics {
        stateDescription = if (isSuccess) "Connection test passed" else "Test connection"
    }
)

// Schema table
Box(modifier = Modifier.semantics {
    stateDescription = "Table ${table.name}, ${table.columns.size} columns, approximately ${table.estimatedRows} rows"
})

// Navigation
Icon(
    imageVector = Icons.Default.ChevronRight,
    contentDescription = "View details for ${agent.name}"
)
```

---

## Responsive Design

| Screen Width       | Agent List Layout     | Detail Layout        |
|--------------------|-----------------------|----------------------|
| Compact (< 600dp)  | Single column list    | Full-screen tabs     |
| Medium (600–839dp) | 2-column grid         | Side-by-side config  |
| Expanded (840dp+)  | 3-column grid         | Navigation rail tabs |

```kotlin
@Composable
fun AdaptiveAgentLayout(
    agents: List<AgentSummary>,
    onAgentClick: (String) -> Unit,
    onCreateAgent: () -> Unit
) {
    val windowSize = currentWindowAdaptiveInfo().windowSizeClass

    when (windowSize.windowWidthSizeClass) {
        WindowWidthSizeClass.COMPACT -> {
            AgentListScreen(
                agents = agents,
                viewMode = ViewMode.LIST,
                onAgentClick = onAgentClick,
                onCreateAgent = onCreateAgent
            )
        }
        WindowWidthSizeClass.MEDIUM -> {
            AgentGrid(
                agents = agents,
                columns = 2,
                onAgentClick = onAgentClick,
                onCreateAgent = onCreateAgent
            )
        }
        WindowWidthSizeClass.EXPANDED -> {
            AgentGrid(
                agents = agents,
                columns = 3,
                onAgentClick = onAgentClick,
                onCreateAgent = onCreateAgent
            )
        }
    }
}
```

---

## Summary

| Feature                     | Screen / Composable                  | ViewModel              |
|-----------------------------|--------------------------------------|------------------------|
| Agent List                  | `AgentListScreen`                    | `AgentListViewModel`   |
| Agent Card                  | `AgentCard` / `AgentCardCompact`     | —                      |
| Agent Detail                | `AgentDetailScreen` (tabs)           | `AgentDetailViewModel` |
| Agent Creation              | `AgentCreationScreen`                | `AgentCreationViewModel`|
| Agent Configuration         | `AgentConfigTab` (sub-tabs)          | `AgentDetailViewModel` |
| Agent Execution             | `AgentExecutionScreen`               | `AgentExecutionViewModel`|
| Agent Execution History     | `AgentHistoryTab`                    | `AgentDetailViewModel` |
| Agent Performance           | `AgentPerformanceMetrics`            | `AgentDetailViewModel` |
| Agent Settings              | `AgentSettingsForm`                  | `AgentDetailViewModel` |
| Tools Config                | `ToolsConfig`                        | `AgentDetailViewModel` |
| Document Set Binding        | `DocumentSetSelector`                | `AgentCreationViewModel`|
| SQL Connection              | `SqlConnectionSelector`              | `AgentDetailViewModel` |
| Test Connection             | `TestConnectionButton`               | `AgentDetailViewModel` |
| Schema Discovery            | `SchemaDiscoveryView`                | `AgentDetailViewModel` |
| Table Binding               | `TableBindingSection`                | `AgentDetailViewModel` |
| Execution Visualization     | `ExecutionVisualization`             | `AgentExecutionViewModel`|
| Error Handling              | `AgentErrorBanner`                   | All ViewModels         |
| Loading                     | `AgentListSkeleton`                  | —                      |
| Empty State                 | `AgentListEmpty`                     | —                      |
| Search & Filter             | `AgentSearchBar`                     | `AgentListViewModel`   |
