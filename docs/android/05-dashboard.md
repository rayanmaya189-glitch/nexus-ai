# 05 - Dashboard

The dashboard is the central hub of the Nexus AI Android application. It provides
a real-time overview of system health, AI usage, agent activity, and key performance
indicators. This document covers every aspect of the dashboard feature from layout
through to accessibility.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Dashboard Screen Layout](#dashboard-screen-layout)
3. [DashboardViewModel](#dashboardviewmodel)
4. [KPI Cards Composable](#kpi-cards-composable)
5. [AI Usage Chart Composable](#ai-usage-chart-composable)
6. [Token Consumption Chart Composable](#token-consumption-chart-composable)
7. [Model Performance Chart Composable](#model-performance-chart-composable)
8. [Active Agents Widget Composable](#active-agents-widget-composable)
9. [Recent Activity Feed Composable](#recent-activity-feed-composable)
10. [System Health Indicators Composable](#system-health-indicators-composable)
11. [Quick Actions Composable](#quick-actions-composable)
12. [Dashboard Filters](#dashboard-filters)
13. [Dashboard API Integration](#dashboard-api-integration)
14. [Pull-to-Refresh](#pull-to-refresh)
15. [Real-time Updates](#real-time-updates)
16. [Dashboard Data Models](#dashboard-data-models)
17. [Dashboard Caching](#dashboard-caching)
18. [Error Handling](#error-handling)
19. [Loading States](#loading-states)
20. [Responsive Design](#responsive-design)
21. [Accessibility](#accessibility)
22. [Performance](#performance)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────┐
│                      UI Layer                            │
│  DashboardScreen ──► KpiCards / Charts / Widgets          │
│        │                                                 │
│        ▼                                                 │
│  DashboardViewModel (Hilt)                               │
│        │                                                 │
│        ├── DashboardRepository (data aggregation)        │
│        ├── MetricsRepository (metrics + cache)           │
│        ├── AgentRepository (agent status)                │
│        └── WebSocketManager (real-time updates)          │
│                                                          │
│  Data Layer                                              │
│  ├── Room Database (DashboardCacheEntity)                │
│  ├── DataStore (DashboardPreferences)                    │
│  ├── Retrofit (DashboardApiService)                      │
│  └── WebSocket (LiveMetricsSocket)                       │
└──────────────────────────────────────────────────────────┘
```

Data flow is unidirectional: API → Repository → ViewModel → UI. Real-time updates
via WebSocket are merged into the ViewModel state through a dedicated
`SharedFlow<MetricsUpdate>`.

---

## Dashboard Screen Layout

The dashboard uses a `LazyColumn` for phone layouts and a staggered grid for tablets.
Sections are rendered as composable blocks inside the lazy list.

```kotlin
@Composable
fun DashboardScreen(
    viewModel: DashboardViewModel = hiltViewModel(),
    onNavigateToChat: () -> Unit,
    onNavigateToAgentDetail: (String) -> Unit,
    onNavigateToAlerts: () -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val isRefreshing = viewModel.isRefreshing.collectAsStateWithLifecycle()

    Scaffold(
        topBar = {
            DashboardTopBar(
                onFilterClick = { viewModel.onAction(DashboardAction.OpenFilters) },
                onRefreshClick = { viewModel.onAction(DashboardAction.Refresh) }
            )
        }
    ) { padding ->
        SwipeRefresh(
            state = rememberSwipeRefreshState(isRefreshing),
            onRefresh = { viewModel.onAction(DashboardAction.Refresh) },
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            DashboardContent(
                state = state,
                onKpiClick = { /* expand detail */ },
                onAgentClick = onNavigateToAgentDetail,
                onQuickAction = { action ->
                    when (action) {
                        QuickAction.START_CHAT -> onNavigateToChat()
                        QuickAction.UPLOAD_DOC -> viewModel.onAction(DashboardAction.UploadDocument)
                        QuickAction.VIEW_ALERTS -> onNavigateToAlerts()
                    }
                },
                onRetry = { viewModel.onAction(DashboardAction.Refresh) }
            )
        }
    }
}
```

### Phone vs Tablet Layout

```
Phone (< 600dp)              Tablet (>= 600dp)
┌───────────────────┐       ┌─────────────────────┬──────────────┐
│  KPI Row (scroll) │       │  KPI Cards (grid)   │              │
├───────────────────┤       ├─────────────────────┤   Activity   │
│  AI Usage Chart   │       │  AI Usage Chart     │   Feed       │
├───────────────────┤       ├─────────────────────┤   + Agent    │
│  Token Chart      │       │  Token / Perf Chart │   Status     │
├───────────────────┤       ├─────────────────────┤              │
│  Agent Status     │       │  System Health      │              │
├───────────────────┤       └─────────────────────┴──────────────┘
│  Activity Feed    │
├───────────────────┤
│  System Health    │
├───────────────────┤
│  Quick Actions    │
└───────────────────┘
```

```kotlin
@Composable
fun DashboardContent(
    state: DashboardState,
    onKpiClick: (KpiType) -> Unit,
    onAgentClick: (String) -> Unit,
    onQuickAction: (QuickAction) -> Unit,
    onRetry: () -> Unit
) {
    val windowSize = currentWindowAdaptiveInfo().windowSizeClass
    val isTablet = windowSize.windowWidthSizeClass ==
        WindowWidthSizeClass.EXPANDED

    if (isTablet) {
        TabletDashboardLayout(
            state = state,
            onKpiClick = onKpiClick,
            onAgentClick = onAgentClick,
            onQuickAction = onQuickAction,
            onRetry = onRetry
        )
    } else {
        PhoneDashboardLayout(
            state = state,
            onKpiClick = onKpiClick,
            onAgentClick = onAgentClick,
            onQuickAction = onQuickAction,
            onRetry = onRetry
        )
    }
}
```

---

## DashboardViewModel

The ViewModel manages all dashboard state, handles user actions, fetches data from
multiple repositories, and merges WebSocket updates.

```kotlin
@HiltViewModel
class DashboardViewModel @Inject constructor(
    private val metricsRepository: MetricsRepository,
    private val agentRepository: AgentRepository,
    private val activityRepository: ActivityRepository,
    private val healthRepository: HealthRepository,
    private val webSocketManager: LiveMetricsSocket,
    private val dashboardPreferences: DashboardPreferences,
    @IoDispatcher private val ioDispatcher: CoroutineDispatcher
) : ViewModel() {

    private val _state = MutableStateFlow(DashboardState())
    val state: StateFlow<DashboardState> = _state.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = _isRefreshing.asStateFlow()

    private val _event = Channel<DashboardEvent>(Channel.BUFFERED)
    val event: Flow<DashboardEvent> = _event.receiveAsFlow()

    init {
        loadDashboard()
        observeRealtimeUpdates()
    }

    fun onAction(action: DashboardAction) {
        when (action) {
            is DashboardAction.Refresh -> refresh()
            is DashboardAction.OpenFilters -> {
                _state.update { it.copy(showFilters = true) }
            }
            is DashboardAction.ApplyFilters -> {
                _state.update { it.copy(filters = action.filters, showFilters = false) }
                loadDashboard()
            }
            is DashboardAction.SelectDateRange -> {
                _state.update { it.copy(filters = it.filters.copy(dateRange = action.range)) }
            }
            is DashboardAction.UploadDocument -> {
                viewModelScope.launch { _event.send(DashboardEvent.NavigateToUpload) }
            }
        }
    }

    private fun loadDashboard() {
        viewModelScope.launch(ioDispatcher) {
            _state.update { it.copy(isLoading = true) }
            try {
                val filters = _state.value.filters
                val metricsDeferred = async { metricsRepository.getDashboardMetrics(filters) }
                val agentsDeferred = async { agentRepository.getActiveAgents() }
                val activityDeferred = async { activityRepository.getRecentActivity(limit = 20) }
                val healthDeferred = async { healthRepository.getSystemHealth() }

                _state.update {
                    it.copy(
                        isLoading = false,
                        kpis = metricsDeferred.await().kpis,
                        aiUsage = metricsDeferred.await().aiUsage,
                        tokenConsumption = metricsDeferred.await().tokenConsumption,
                        modelPerformance = metricsDeferred.await().modelPerformance,
                        activeAgents = agentsDeferred.await(),
                        recentActivity = activityDeferred.await(),
                        systemHealth = healthDeferred.await()
                    )
                }
            } catch (e: Exception) {
                _state.update {
                    it.copy(
                        isLoading = false,
                        error = DashboardError.LoadFailed(e.message ?: "Unknown error")
                    )
                }
            }
        }
    }

    private fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            loadDashboard()
            _isRefreshing.value = false
        }
    }

    private fun observeRealtimeUpdates() {
        viewModelScope.launch {
            webSocketManager.metricsUpdates.collect { update ->
                _state.update { state ->
                    state.copy(
                        kpis = applyKpiUpdate(state.kpis, update),
                        systemHealth = applyHealthUpdate(state.systemHealth, update)
                    )
                }
            }
        }
    }
}
```

### DashboardState

```kotlin
data class DashboardState(
    val isLoading: Boolean = false,
    val kpis: List<KpiCard> = emptyList(),
    val aiUsage: AiUsageChart = AiUsageChart.empty(),
    val tokenConsumption: TokenChart = TokenChart.empty(),
    val modelPerformance: ModelPerformanceChart = ModelPerformanceChart.empty(),
    val activeAgents: List<AgentStatus> = emptyList(),
    val recentActivity: List<ActivityEvent> = emptyList(),
    val systemHealth: SystemHealth = SystemHealth.empty(),
    val filters: DashboardFilters = DashboardFilters(),
    val showFilters: Boolean = false,
    val error: DashboardError? = null
)
```

### DashboardAction

```kotlin
sealed interface DashboardAction {
    data object Refresh : DashboardAction
    data object OpenFilters : DashboardAction
    data class ApplyFilters(val filters: DashboardFilters) : DashboardAction
    data class SelectDateRange(val range: DateRange) : DashboardAction
    data object UploadDocument : DashboardAction
}
```

### DashboardEvent

```kotlin
sealed interface DashboardEvent {
    data object NavigateToUpload : DashboardEvent
    data class ShowError(val message: String) : DashboardEvent
}
```

---

## KPI Cards Composable

Four KPI cards sit at the top of the dashboard: Total Requests, Active Users,
Model Usage, and Cost. Each card shows a value, trend indicator, and mini
sparkline.

```kotlin
@Composable
fun KpiCardsRow(
    kpis: List<KpiCard>,
    onKpiClick: (KpiType) -> Unit,
    modifier: Modifier = Modifier
) {
    LazyRow(
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        contentPadding = PaddingValues(horizontal = 16.dp),
        modifier = modifier
    ) {
        items(kpis, key = { it.type }) { kpi ->
            KpiCard(
                kpi = kpi,
                onClick = { onKpiClick(kpi.type) },
                modifier = Modifier.width(160.dp)
            )
        }
    }
}

@Composable
fun KpiCard(
    kpi: KpiCard,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val trendColor = when (kpi.trend) {
        Trend.UP -> MaterialTheme.colorScheme.primary
        Trend.DOWN -> MaterialTheme.colorScheme.error
        Trend.STABLE -> MaterialTheme.colorScheme.outline
    }

    Card(
        onClick = onClick,
        modifier = modifier,
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
                modifier = Modifier.fillMaxWidth()
            ) {
                Icon(
                    imageVector = kpi.icon,
                    contentDescription = kpi.label,
                    tint = MaterialTheme.colorScheme.primary,
                    modifier = Modifier.size(24.dp)
                )
                TrendIndicator(
                    value = kpi.trendValue,
                    trend = kpi.trend,
                    color = trendColor
                )
            }
            Spacer(modifier = Modifier.height(12.dp))
            Text(
                text = kpi.formattedValue,
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = kpi.label,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            if (kpi.sparklineData.isNotEmpty()) {
                Spacer(modifier = Modifier.height(8.dp))
                MiniSparkline(
                    data = kpi.sparklineData,
                    color = trendColor,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(32.dp)
                )
            }
        }
    }
}

@Composable
fun TrendIndicator(
    value: Float,
    trend: Trend,
    color: Color
) {
    val icon = when (trend) {
        Trend.UP -> Icons.Default.TrendingUp
        Trend.DOWN -> Icons.Default.TrendingDown
        Trend.STABLE -> Icons.Default.TrendingFlat
    }
    Row(verticalAlignment = Alignment.CenterVertically) {
        Icon(
            imageVector = icon,
            contentDescription = "Trend",
            tint = color,
            modifier = Modifier.size(16.dp)
        )
        Text(
            text = "${String.format("%.1f", value)}%",
            style = MaterialTheme.typography.labelSmall,
            color = color
        )
    }
}
```

### KPI Card Data

| KPI              | Icon                    | Format      | Trend Source          |
|------------------|-------------------------|-------------|-----------------------|
| Total Requests   | `Icons.AutoMirrored.Filled.Send` | `#,###`  | vs previous period    |
| Active Users     | `Icons.Default.People`  | `#,###`     | vs previous period    |
| Model Usage      | `Icons.Default.SmartToy` | `#,##0.0B` | vs previous period    |
| Cost             | `Icons.Default.AttachMoney` | `$#,##0.00` | vs previous period |

---

## AI Usage Chart Composable

Displays requests over time as a line chart or bar chart with toggle.

```kotlin
@Composable
fun AiUsageChart(
    data: AiUsageChart,
    modifier: Modifier = Modifier
) {
    var chartType by remember { mutableStateOf(ChartType.LINE) }

    Card(modifier = modifier) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "AI Usage",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold
                )
                ChartTypeToggle(
                    selected = chartType,
                    onToggle = { chartType = it }
                )
            }
            Spacer(modifier = Modifier.height(16.dp))

            when (chartType) {
                ChartType.LINE -> LineChart(
                    dataPoints = data.dataPoints.map {
                        PointF(it.timestamp.toFloat(), it.requestCount.toFloat())
                    },
                    lineColor = MaterialTheme.colorScheme.primary,
                    fillColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(200.dp)
                )
                ChartType.BAR -> BarChart(
                    dataPoints = data.dataPoints.map {
                        BarData(
                            label = it.label,
                            value = it.requestCount.toFloat(),
                            color = MaterialTheme.colorScheme.primary
                        )
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(200.dp)
                )
            }

            Spacer(modifier = Modifier.height(8.dp))
            LegendRow(
                items = listOf(
                    LegendItem("Total", MaterialTheme.colorScheme.primary),
                    LegendItem("Success", MaterialTheme.colorScheme.tertiary),
                    LegendItem("Failed", MaterialTheme.colorScheme.error)
                )
            )
        }
    }
}
```

### Line Chart Composable (Canvas)

```kotlin
@Composable
fun LineChart(
    dataPoints: List<PointF>,
    lineColor: Color,
    fillColor: Color,
    modifier: Modifier = Modifier
) {
    Canvas(modifier = modifier) {
        if (dataPoints.isEmpty()) return@Canvas

        val maxVal = dataPoints.maxOf { it.y }
        val minVal = dataPoints.minOf { it.y }
        val range = (maxVal - minVal).coerceAtLeast(1f)

        val stepX = size.width / (dataPoints.size - 1).coerceAtLeast(1)

        val path = Path()
        val fillPath = Path()

        dataPoints.forEachIndexed { index, point ->
            val x = index * stepX
            val y = size.height - ((point.y - minVal) / range * size.height)

            if (index == 0) {
                path.moveTo(x, y)
                fillPath.moveTo(x, size.height)
                fillPath.lineTo(x, y)
            } else {
                val prevX = (index - 1) * stepX
                val prevY = size.height - ((dataPoints[index - 1].y - minVal) / range * size.height)
                val controlX1 = prevX + stepX / 3f
                val controlX2 = x - stepX / 3f
                path.cubicTo(controlX1, prevY, controlX2, y, x, y)
                fillPath.cubicTo(controlX1, prevY, controlX2, y, x, y)
            }
        }

        fillPath.lineTo(size.width, size.height)
        fillPath.close()

        drawPath(fillPath, fillColor)
        drawPath(path, lineColor, style = Stroke(width = 3.dp.toPx()))
    }
}
```

---

## Token Consumption Chart Composable

Stacked bar chart showing input vs output tokens over time.

```kotlin
@Composable
fun TokenConsumptionChart(
    data: TokenChart,
    modifier: Modifier = Modifier
) {
    Card(modifier = modifier) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "Token Consumption",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold
            )
            Spacer(modifier = Modifier.height(16.dp))

            StackedBarChart(
                entries = data.entries.map { entry ->
                    StackedBarEntry(
                        label = entry.label,
                        segments = listOf(
                            BarSegment(
                                value = entry.inputTokens.toFloat(),
                                color = MaterialTheme.colorScheme.primary,
                                label = "Input"
                            ),
                            BarSegment(
                                value = entry.outputTokens.toFloat(),
                                color = MaterialTheme.colorScheme.secondary,
                                label = "Output"
                            )
                        )
                    )
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(200.dp)
            )

            Spacer(modifier = Modifier.height(8.dp))
            LegendRow(
                items = listOf(
                    LegendItem("Input Tokens", MaterialTheme.colorScheme.primary),
                    LegendItem("Output Tokens", MaterialTheme.colorScheme.secondary)
                )
            )

            Spacer(modifier = Modifier.height(12.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                TokenStat(
                    label = "Total Input",
                    value = data.totalInputTokens,
                    color = MaterialTheme.colorScheme.primary
                )
                TokenStat(
                    label = "Total Output",
                    value = data.totalOutputTokens,
                    color = MaterialTheme.colorScheme.secondary
                )
                TokenStat(
                    label = "Avg/Turn",
                    value = data.avgTokensPerTurn,
                    color = MaterialTheme.colorScheme.tertiary
                )
            }
        }
    }
}

@Composable
fun TokenStat(label: String, value: Long, color: Color) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(
            text = formatTokenCount(value),
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold,
            color = color
        )
        Text(
            text = label,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

fun formatTokenCount(count: Long): String = when {
    count >= 1_000_000 -> "${count / 1_000_000}M"
    count >= 1_000 -> "${count / 1_000}K"
    else -> count.toString()
}
```

### StackedBarChart Canvas

```kotlin
@Composable
fun StackedBarChart(
    entries: List<StackedBarEntry>,
    modifier: Modifier = Modifier
) {
    Canvas(modifier = modifier) {
        val maxValue = entries.maxOf { entry ->
            entry.segments.sumOf { it.value.toDouble() }.toFloat()
        }
        val barWidth = size.width / entries.size * 0.6f
        val spacing = size.width / entries.size * 0.4f

        entries.forEachIndexed { index, entry ->
            val x = index * (barWidth + spacing) + spacing / 2
            var currentY = size.height
            val totalValue = entry.segments.sumOf { it.value.toDouble() }.toFloat()

            entry.segments.forEach { segment ->
                val segHeight = (segment.value / maxValue) * size.height
                drawRoundRect(
                    color = segment.color,
                    topLeft = Offset(x, currentY - segHeight),
                    size = Size(barWidth, segHeight),
                    cornerRadius = CornerRadius(4.dp.toPx(), 4.dp.toPx())
                )
                currentY -= segHeight
            }

            // Label below bar
            drawContext.canvas.nativeCanvas.apply {
                drawText(
                    entry.label,
                    x + barWidth / 2,
                    size.height + 16.dp.toPx(),
                    android.graphics.Paint().apply {
                        textSize = 10.sp.toPx()
                        textAlign = android.graphics.Paint.Align.CENTER
                        color = android.graphics.Color.GRAY
                    }
                )
            }
        }
    }
}
```

---

## Model Performance Chart Composable

Displays latency, throughput, and error rate per model.

```kotlin
@Composable
fun ModelPerformanceChart(
    data: ModelPerformanceChart,
    modifier: Modifier = Modifier
) {
    var selectedMetric by remember { mutableStateOf(PerformanceMetric.LATENCY) }

    Card(modifier = modifier) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Model Performance",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold
                )
                MetricFilterChips(
                    selected = selectedMetric,
                    onSelect = { selectedMetric = it }
                )
            }
            Spacer(modifier = Modifier.height(16.dp))

            val chartData = data.models.map { model ->
                val value = when (selectedMetric) {
                    PerformanceMetric.LATENCY -> model.avgLatencyMs
                    PerformanceMetric.THROUGHPUT -> model.throughputPerMin
                    PerformanceMetric.ERROR_RATE -> model.errorRate * 100f
                }
                BarData(
                    label = model.modelName,
                    value = value,
                    color = getMetricColor(selectedMetric)
                )
            }

            HorizontalBarChart(
                bars = chartData,
                modifier = Modifier
                    .fillMaxWidth()
                    .height((data.models.size * 48).dp)
            )
        }
    }
}

@Composable
fun MetricFilterChips(
    selected: PerformanceMetric,
    onSelect: (PerformanceMetric) -> Unit
) {
    Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
        PerformanceMetric.entries.forEach { metric ->
            FilterChip(
                selected = selected == metric,
                onClick = { onSelect(metric) },
                label = { Text(metric.label, style = MaterialTheme.typography.labelSmall) }
            )
        }
    }
}

enum class PerformanceMetric(val label: String, val unit: String) {
    LATENCY("Latency", "ms"),
    THROUGHPUT("Throughput", "req/min"),
    ERROR_RATE("Error Rate", "%")
}
```

---

## Active Agents Widget Composable

Shows online/offline agents with status indicators.

```kotlin
@Composable
fun ActiveAgentsWidget(
    agents: List<AgentStatus>,
    onAgentClick: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    Card(modifier = modifier) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Active Agents",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold
                )
                Badge(
                    containerColor = MaterialTheme.colorScheme.primaryContainer,
                    contentColor = MaterialTheme.colorScheme.onPrimaryContainer
                ) {
                    Text("${agents.count { it.isOnline }}")
                }
            }
            Spacer(modifier = Modifier.height(12.dp))

            if (agents.isEmpty()) {
                EmptyState(
                    icon = Icons.Default.SmartToy,
                    message = "No agents configured"
                )
            } else {
                agents.take(5).forEach { agent ->
                    AgentStatusRow(
                        agent = agent,
                        onClick = { onAgentClick(agent.agentId) }
                    )
                    if (agent != agents.last()) {
                        HorizontalDivider(modifier = Modifier.padding(vertical = 4.dp))
                    }
                }
                if (agents.size > 5) {
                    TextButton(
                        onClick = { /* navigate to full agent list */ },
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text("View all ${agents.size} agents")
                    }
                }
            }
        }
    }
}

@Composable
fun AgentStatusRow(
    agent: AgentStatus,
    onClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(10.dp)
                .clip(CircleShape)
                .background(if (agent.isOnline) Color(0xFF4CAF50) else Color(0xFF9E9E9E))
        )
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = agent.name,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium
            )
            Text(
                text = agent.model,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
        Column(horizontalAlignment = Alignment.End) {
            Text(
                text = "${agent.activeTasks} tasks",
                style = MaterialTheme.typography.bodySmall
            )
            Text(
                text = "${agent.avgLatencyMs}ms avg",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}
```

---

## Recent Activity Feed Composable

Lists latest conversations, tool executions, and system events.

```kotlin
@Composable
fun RecentActivityFeed(
    activities: List<ActivityEvent>,
    modifier: Modifier = Modifier
) {
    Card(modifier = modifier) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "Recent Activity",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold
            )
            Spacer(modifier = Modifier.height(12.dp))

            if (activities.isEmpty()) {
                EmptyState(
                    icon = Icons.Default.History,
                    message = "No recent activity"
                )
            } else {
                activities.forEach { event ->
                    ActivityEventRow(event = event)
                    if (event != activities.last()) {
                        HorizontalDivider(modifier = Modifier.padding(vertical = 4.dp))
                    }
                }
            }
        }
    }
}

@Composable
fun ActivityEventRow(event: ActivityEvent) {
    val (icon, color) = when (event.type) {
        ActivityType.CONVERSATION -> Icons.Default.Chat to MaterialTheme.colorScheme.primary
        ActivityType.TOOL_EXECUTION -> Icons.Default.Build to MaterialTheme.colorScheme.secondary
        ActivityType.DOCUMENT_UPLOAD -> Icons.Default.Upload to MaterialTheme.colorScheme.tertiary
        ActivityType.AGENT_ACTION -> Icons.Default.SmartToy to MaterialTheme.colorScheme.primary
        ActivityType.SYSTEM_ALERT -> Icons.Default.Warning to MaterialTheme.colorScheme.error
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 6.dp),
        verticalAlignment = Alignment.Top
    ) {
        Box(
            modifier = Modifier
                .size(32.dp)
                .clip(CircleShape)
                .background(color.copy(alpha = 0.12f)),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = icon,
                contentDescription = event.type.name,
                tint = color,
                modifier = Modifier.size(16.dp)
            )
        }
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = event.title,
                style = MaterialTheme.typography.bodyMedium
            )
            if (event.description.isNotEmpty()) {
                Text(
                    text = event.description,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis
                )
            }
        }
        Text(
            text = event.timestamp.toRelativeTimeString(),
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}
```

---

## System Health Indicators Composable

Gauges for CPU, RAM, GPU, and storage utilization.

```kotlin
@Composable
fun SystemHealthIndicators(
    health: SystemHealth,
    modifier: Modifier = Modifier
) {
    Card(modifier = modifier) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "System Health",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold
            )
            Spacer(modifier = Modifier.height(16.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                HealthGauge(label = "CPU", usage = health.cpuUsage)
                HealthGauge(label = "RAM", usage = health.ramUsage)
                HealthGauge(label = "GPU", usage = health.gpuUsage)
                HealthGauge(label = "Disk", usage = health.diskUsage)
            }

            Spacer(modifier = Modifier.height(12.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                HealthDetail("Uptime", health.uptime)
                HealthDetail("Latency", "${health.avgLatencyMs}ms")
                HealthDetail("Error Rate", "${String.format("%.2f", health.errorRate)}%")
            }
        }
    }
}

@Composable
fun HealthGauge(
    label: String,
    usage: Float,
    modifier: Modifier = Modifier
) {
    val color = when {
        usage < 0.5f -> Color(0xFF4CAF50)
        usage < 0.8f -> Color(0xFFFFC107)
        else -> Color(0xFFF44336)
    }
    val sweepAngle = usage * 270f

    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = modifier
    ) {
        Canvas(modifier = Modifier.size(64.dp)) {
            val strokeWidth = 8.dp.toPx()
            val radius = (size.minDimension - strokeWidth) / 2

            // Background arc
            drawArc(
                color = color.copy(alpha = 0.15f),
                startAngle = 135f,
                sweepAngle = 270f,
                useCenter = false,
                style = Stroke(strokeWidth, cap = StrokeCap.Round),
                topLeft = Offset(strokeWidth / 2, strokeWidth / 2),
                size = Size(radius * 2, radius * 2)
            )

            // Usage arc
            drawArc(
                color = color,
                startAngle = 135f,
                sweepAngle = sweepAngle,
                useCenter = false,
                style = Stroke(strokeWidth, cap = StrokeCap.Round),
                topLeft = Offset(strokeWidth / 2, strokeWidth / 2),
                size = Size(radius * 2, radius * 2)
            )
        }
        Text(
            text = "${(usage * 100).toInt()}%",
            style = MaterialTheme.typography.labelLarge,
            fontWeight = FontWeight.Bold,
            color = color
        )
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
fun HealthDetail(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(
            text = value,
            style = MaterialTheme.typography.bodyMedium,
            fontWeight = FontWeight.SemiBold
        )
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}
```

---

## Quick Actions Composable

Three action buttons: Start Chat, Upload Document, View Alerts.

```kotlin
enum class QuickAction(val label: String, val icon: ImageVector) {
    START_CHAT("Start Chat", Icons.Default.Chat),
    UPLOAD_DOC("Upload Document", Icons.Default.Upload),
    VIEW_ALERTS("View Alerts", Icons.Default.Notifications)
}

@Composable
fun QuickActionsRow(
    onAction: (QuickAction) -> Unit,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        QuickAction.entries.forEach { action ->
            FilledTonalButton(
                onClick = { onAction(action) },
                modifier = Modifier.weight(1f),
                contentPadding = PaddingValues(vertical = 12.dp)
            ) {
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Icon(
                        imageVector = action.icon,
                        contentDescription = action.label,
                        modifier = Modifier.size(24.dp)
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = action.label,
                        style = MaterialTheme.typography.labelMedium
                    )
                }
            }
        }
    }
}
```

---

## Dashboard Filters

A bottom sheet with date range, tenant, user, and model filters.

```kotlin
data class DashboardFilters(
    val dateRange: DateRange = DateRange.LAST_7_DAYS,
    val tenantId: String? = null,
    val userId: String? = null,
    val modelId: String? = null
)

enum class DateRange(val label: String) {
    LAST_24H("Last 24 Hours"),
    LAST_7_DAYS("Last 7 Days"),
    LAST_30_DAYS("Last 30 Days"),
    LAST_90_DAYS("Last 90 Days"),
    CUSTOM("Custom Range")
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DashboardFiltersSheet(
    filters: DashboardFilters,
    onApply: (DashboardFilters) -> Unit,
    onDismiss: () -> Unit
) {
    var selectedRange by remember { mutableStateOf(filters.dateRange) }
    var selectedTenant by remember { mutableStateOf(filters.tenantId) }
    var selectedUser by remember { mutableStateOf(filters.userId) }
    var selectedModel by remember { mutableStateOf(filters.modelId) }

    ModalBottomSheet(onDismissRequest = onDismiss) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = "Dashboard Filters",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(20.dp))

            Text("Date Range", style = MaterialTheme.typography.titleSmall)
            Spacer(modifier = Modifier.height(8.dp))
            SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                DateRange.entries.filter { it != DateRange.CUSTOM }.forEachIndexed { i, range ->
                    SegmentedButton(
                        selected = selectedRange == range,
                        onClick = { selectedRange = range },
                        shape = SegmentedButtonDefaults.itemShape(i, DateRange.entries.size - 1)
                    ) {
                        Text(range.label, style = MaterialTheme.typography.labelSmall)
                    }
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            // Tenant dropdown
            FilterDropdown(
                label = "Tenant",
                selected = selectedTenant,
                options = listOf("All Tenants", "Tenant A", "Tenant B"),
                onSelect = { selectedTenant = it }
            )

            Spacer(modifier = Modifier.height(12.dp))

            // User search
            FilterSearchField(
                label = "User",
                query = selectedUser ?: "",
                onQueryChange = { selectedUser = it }
            )

            Spacer(modifier = Modifier.height(12.dp))

            // Model dropdown
            FilterDropdown(
                label = "Model",
                selected = selectedModel,
                options = listOf("All Models", "GPT-4", "Claude 3", "Gemini"),
                onSelect = { selectedModel = it }
            )

            Spacer(modifier = Modifier.height(24.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                OutlinedButton(
                    onClick = {
                        selectedRange = DateRange.LAST_7_DAYS
                        selectedTenant = null
                        selectedUser = null
                        selectedModel = null
                    },
                    modifier = Modifier.weight(1f)
                ) { Text("Reset") }
                Button(
                    onClick = {
                        onApply(
                            DashboardFilters(
                                dateRange = selectedRange,
                                tenantId = selectedTenant,
                                userId = selectedUser,
                                modelId = selectedModel
                            )
                        )
                    },
                    modifier = Modifier.weight(1f)
                ) { Text("Apply") }
            }

            Spacer(modifier = Modifier.height(16.dp))
        }
    }
}
```

---

## Dashboard API Integration

### Retrofit Service

```kotlin
interface DashboardApiService {

    @GET("api/v1/metrics/dashboard")
    suspend fun getDashboardMetrics(
        @Query("dateRange") dateRange: String,
        @Query("tenantId") tenantId: String? = null,
        @Query("userId") userId: String? = null,
        @Query("modelId") modelId: String? = null
    ): DashboardMetricsResponse

    @GET("api/v1/metrics/ai-usage")
    suspend fun getAiUsage(
        @Query("dateRange") dateRange: String,
        @Query("granularity") granularity: String = "hourly"
    ): AiUsageResponse

    @GET("api/v1/metrics/tokens")
    suspend fun getTokenConsumption(
        @Query("dateRange") dateRange: String
    ): TokenConsumptionResponse

    @GET("api/v1/metrics/model-performance")
    suspend fun getModelPerformance(
        @Query("dateRange") dateRange: String
    ): ModelPerformanceResponse

    @GET("api/v1/health")
    suspend fun getSystemHealth(): SystemHealthResponse

    @GET("api/v1/agents/active")
    suspend fun getActiveAgents(): List<AgentStatusResponse>

    @GET("api/v1/activity/recent")
    suspend fun getRecentActivity(
        @Query("limit") limit: Int = 20
    ): List<ActivityEventResponse>
}
```

### Response Models

```kotlin
@Serializable
data class DashboardMetricsResponse(
    val kpis: KpisDto,
    val aiUsage: AiUsageDto,
    val tokenConsumption: TokenConsumptionDto,
    val modelPerformance: ModelPerformanceDto
)

@Serializable
data class KpisDto(
    val totalRequests: Long,
    val totalRequestsTrend: Float,
    val activeUsers: Int,
    val activeUsersTrend: Float,
    val modelUsage: Double,
    val modelUsageTrend: Float,
    val totalCost: Double,
    val totalCostTrend: Float,
    val sparklineData: List<SparklinePointDto>
)

@Serializable
data class AiUsageDto(
    val dataPoints: List<AiUsagePointDto>,
    val totalRequests: Long,
    val successRate: Float
)

@Serializable
data class AiUsagePointDto(
    val timestamp: String,
    val label: String,
    val requestCount: Long,
    val successCount: Long,
    val failureCount: Long
)
```

---

## Pull-to-Refresh

Uses the Accompanist `SwipeRefresh` or Material 3 `PullToRefreshBox`.

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DashboardPullToRefresh(
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

## Real-time Updates

WebSocket connection streams live metrics to the dashboard.

```kotlin
class LiveMetricsSocket @Inject constructor(
    private val tokenProvider: AuthTokenProvider,
    private val baseUrl: String
) {
    private var webSocket: WebSocket? = null
    private val _metricsUpdates = MutableSharedFlow<MetricsUpdate>(
        replay = 0,
        extraBufferCapacity = 64
    )
    val metricsUpdates: SharedFlow<MetricsUpdate> = _metricsUpdates.asSharedFlow()

    fun connect() {
        val client = OkHttpClient.Builder()
            .readTimeout(0, TimeUnit.MILLISECONDS)
            .build()

        val request = Request.Builder()
            .url("$baseUrl/ws/metrics")
            .addHeader("Authorization", "Bearer ${tokenProvider.getToken()}")
            .build()

        webSocket = client.newWebSocket(request, object : WebSocketListener() {
            override fun onMessage(ws: WebSocket, text: String) {
                try {
                    val update = Json.decodeFromString<MetricsUpdate>(text)
                    _metricsUpdates.tryEmit(update)
                } catch (e: Exception) {
                    Log.e("LiveMetrics", "Parse error", e)
                }
            }

            override fun onFailure(ws: WebSocket, t: Throwable, resp: Response?) {
                Log.e("LiveMetrics", "Connection failed", t)
                reconnectWithBackoff()
            }

            override fun onClosed(ws: WebSocket, code: Int, reason: String) {
                reconnectWithBackoff()
            }
        })
    }

    private fun reconnectWithBackoff() {
        CoroutineScope(Dispatchers.IO).launch {
            var delay = 1000L
            while (true) {
                delay(delay.coerceAtMost(30000L))
                connect()
                return@launch
            }
        }
    }

    fun disconnect() {
        webSocket?.close(1000, "Closing")
        webSocket = null
    }
}

@Serializable
data class MetricsUpdate(
    val type: String,
    val cpuUsage: Float? = null,
    val ramUsage: Float? = null,
    val gpuUsage: Float? = null,
    val activeRequests: Int? = null,
    val timestamp: String
)
```

---

## Dashboard Data Models

### Kotlin Data Classes (Domain)

```kotlin
data class KpiCard(
    val type: KpiType,
    val label: String,
    val formattedValue: String,
    val rawValue: Double,
    val trend: Trend,
    val trendValue: Float,
    val icon: ImageVector,
    val sparklineData: List<Float> = emptyList()
)

enum class KpiType { TOTAL_REQUESTS, ACTIVE_USERS, MODEL_USAGE, COST }
enum class Trend { UP, DOWN, STABLE }

data class AiUsageChart(
    val dataPoints: List<AiUsagePoint>,
    val totalRequests: Long,
    val successRate: Float
) {
    companion object {
        fun empty() = AiUsageChart(emptyList(), 0, 0f)
    }
}

data class AiUsagePoint(
    val timestamp: Long,
    val label: String,
    val requestCount: Long,
    val successCount: Long,
    val failureCount: Long
)

data class TokenChart(
    val entries: List<TokenEntry>,
    val totalInputTokens: Long,
    val totalOutputTokens: Long,
    val avgTokensPerTurn: Long
) {
    companion object {
        fun empty() = TokenChart(emptyList(), 0, 0, 0)
    }
}

data class TokenEntry(
    val label: String,
    val inputTokens: Long,
    val outputTokens: Long
)

data class ModelPerformanceChart(
    val models: List<ModelPerformanceEntry>
) {
    companion object {
        fun empty() = ModelPerformanceChart(emptyList())
    }
}

data class ModelPerformanceEntry(
    val modelName: String,
    val avgLatencyMs: Float,
    val throughputPerMin: Float,
    val errorRate: Float
)

data class AgentStatus(
    val agentId: String,
    val name: String,
    val model: String,
    val isOnline: Boolean,
    val activeTasks: Int,
    val avgLatencyMs: Int
)

data class ActivityEvent(
    val id: String,
    val type: ActivityType,
    val title: String,
    val description: String,
    val timestamp: Instant,
    val metadata: Map<String, String> = emptyMap()
)

enum class ActivityType {
    CONVERSATION, TOOL_EXECUTION, DOCUMENT_UPLOAD,
    AGENT_ACTION, SYSTEM_ALERT
}

data class SystemHealth(
    val cpuUsage: Float,
    val ramUsage: Float,
    val gpuUsage: Float,
    val diskUsage: Float,
    val uptime: String,
    val avgLatencyMs: Int,
    val errorRate: Float
) {
    companion object {
        fun empty() = SystemHealth(0f, 0f, 0f, 0f, "0s", 0, 0f)
    }
}

data class DashboardError(val message: String)
```

---

## Dashboard Caching

### Room Entity

```kotlin
@Entity(tableName = "dashboard_cache")
data class DashboardCacheEntity(
    @PrimaryKey val id: String = "dashboard",
    val kpisJson: String,
    val aiUsageJson: String,
    val tokenJson: String,
    val modelPerfJson: String,
    val healthJson: String,
    val agentsJson: String,
    val activityJson: String,
    val lastUpdated: Long
)

@Dao
interface DashboardCacheDao {

    @Query("SELECT * FROM dashboard_cache WHERE id = 'dashboard'")
    suspend fun getCached(): DashboardCacheEntity?

    @Upsert
    suspend fun upsert(entity: DashboardCacheEntity)

    @Query("DELETE FROM dashboard_cache")
    suspend fun clear()
}
```

### DataStore Preferences

```kotlin
class DashboardPreferences @Inject constructor(
    private val dataStore: DataStore<Preferences>
) {
    val lastFilterJson: Flow<String?> = dataStore.data.map {
        it[stringPreferencesKey("last_dashboard_filters")]
    }

    suspend fun saveFilters(filters: DashboardFilters) {
        dataStore.edit { prefs ->
            prefs[stringPreferencesKey("last_dashboard_filters")] =
                Json.encodeToString(filters)
        }
    }

    suspend fun getSavedFilters(): DashboardFilters {
        return lastFilterJson.first()?.let {
            Json.decodeFromString(it)
        } ?: DashboardFilters()
    }
}
```

---

## Error Handling

### Error States

```kotlin
sealed interface DashboardError {
    data class LoadFailed(val message: String) : DashboardError
    data class WebSocketError(val message: String) : DashboardError
    data object NoConnection : DashboardError
}

@Composable
fun DashboardErrorState(
    error: DashboardError,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier.fillMaxWidth().padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Icon(
            imageVector = when (error) {
                is DashboardError.NoConnection -> Icons.Default.WifiOff
                else -> Icons.Default.ErrorOutline
            },
            contentDescription = "Error",
            modifier = Modifier.size(64.dp),
            tint = MaterialTheme.colorScheme.error
        )
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = when (error) {
                is DashboardError.LoadFailed -> "Failed to load dashboard"
                is DashboardError.WebSocketError -> "Real-time connection lost"
                is DashboardError.NoConnection -> "No internet connection"
            },
            style = MaterialTheme.typography.titleMedium,
            textAlign = TextAlign.Center
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = when (error) {
                is DashboardError.LoadFailed -> error.message
                is DashboardError.WebSocketError -> error.message
                is DashboardError.NoConnection -> "Please check your network settings"
            },
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
        Spacer(modifier = Modifier.height(16.dp))
        Button(onClick = onRetry) {
            Icon(Icons.Default.Refresh, contentDescription = null)
            Spacer(modifier = Modifier.width(8.dp))
            Text("Retry")
        }
    }
}
```

### Empty State

```kotlin
@Composable
fun EmptyState(
    icon: ImageVector,
    message: String,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier.fillMaxWidth().padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Icon(
            imageVector = icon,
            contentDescription = null,
            modifier = Modifier.size(48.dp),
            tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f)
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = message,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
    }
}
```

---

## Loading States

### Shimmer Skeleton

```kotlin
@Composable
fun DashboardSkeleton() {
    val shimmerBrush = shimmerBrush()

    LazyColumn(
        verticalArrangement = Arrangement.spacedBy(16.dp),
        contentPadding = PaddingValues(16.dp)
    ) {
        // KPI skeleton
        items(4) {
            SkeletonKpiCard(shimmerBrush)
        }
        // Chart skeletons
        item { SkeletonChart(shimmerBrush, height = 240.dp) }
        item { SkeletonChart(shimmerBrush, height = 200.dp) }
        // Agent skeleton
        items(3) {
            SkeletonAgentRow(shimmerBrush)
        }
        // Activity skeleton
        items(5) {
            SkeletonActivityRow(shimmerBrush)
        }
    }
}

@Composable
fun SkeletonKpiCard(brush: Brush) {
    Card(
        modifier = Modifier
            .width(160.dp)
            .height(120.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Box(
                modifier = Modifier
                    .size(24.dp)
                    .background(brush, RoundedCornerShape(4.dp))
            )
            Spacer(modifier = Modifier.height(12.dp))
            Box(
                modifier = Modifier
                    .width(80.dp)
                    .height(20.dp)
                    .background(brush, RoundedCornerShape(4.dp))
            )
            Spacer(modifier = Modifier.height(4.dp))
            Box(
                modifier = Modifier
                    .width(60.dp)
                    .height(12.dp)
                    .background(brush, RoundedCornerShape(4.dp))
            )
        }
    }
}

@Composable
fun shimmerBrush(): Brush {
    val transition = rememberInfiniteTransition()
    val translateAnim by transition.animateFloat(
        initialValue = 0f,
        targetValue = 1000f,
        animationSpec = infiniteRepeatable(
            animation = tween(1200, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        )
    )
    return Brush.linearGradient(
        colors = listOf(
            MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
            MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.2f),
            MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
        ),
        start = Offset(translateAnim, 0f),
        end = Offset(translateAnim + 200f, 200f)
    )
}

@Composable
fun SkeletonChart(brush: Brush, height: Dp) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(height)
                .padding(16.dp)
                .background(brush, RoundedCornerShape(8.dp))
        )
    }
}

@Composable
fun SkeletonAgentRow(brush: Brush) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(10.dp)
                .background(brush, CircleShape)
        )
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Box(
                modifier = Modifier
                    .width(120.dp)
                    .height(14.dp)
                    .background(brush, RoundedCornerShape(4.dp))
            )
            Spacer(modifier = Modifier.height(4.dp))
            Box(
                modifier = Modifier
                    .width(80.dp)
                    .height(10.dp)
                    .background(brush, RoundedCornerShape(4.dp))
            )
        }
    }
}

@Composable
fun SkeletonActivityRow(brush: Brush) {
    Row(
        modifier = Modifier.fillMaxWidth().padding(vertical = 6.dp),
        verticalAlignment = Alignment.Top
    ) {
        Box(
            modifier = Modifier
                .size(32.dp)
                .background(brush, CircleShape)
        )
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(0.7f)
                    .height(12.dp)
                    .background(brush, RoundedCornerShape(4.dp))
            )
            Spacer(modifier = Modifier.height(4.dp))
            Box(
                modifier = Modifier
                    .fillMaxWidth(0.5f)
                    .height(10.dp)
                    .background(brush, RoundedCornerShape(4.dp))
            )
        }
    }
}
```

---

## Responsive Design

### Adaptive Layout

```kotlin
@Composable
fun PhoneDashboardLayout(
    state: DashboardState,
    onKpiClick: (KpiType) -> Unit,
    onAgentClick: (String) -> Unit,
    onQuickAction: (QuickAction) -> Unit,
    onRetry: () -> Unit
) {
    LazyColumn(
        verticalArrangement = Arrangement.spacedBy(16.dp),
        contentPadding = PaddingValues(16.dp)
    ) {
        item { KpiCardsRow(kpis = state.kpis, onKpiClick = onKpiClick) }
        item {
            AiUsageChart(
                data = state.aiUsage,
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            TokenConsumptionChart(
                data = state.tokenConsumption,
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            ModelPerformanceChart(
                data = state.modelPerformance,
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            ActiveAgentsWidget(
                agents = state.activeAgents,
                onAgentClick = onAgentClick,
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            RecentActivityFeed(
                activities = state.recentActivity,
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            SystemHealthIndicators(
                health = state.systemHealth,
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            QuickActionsRow(
                onAction = onQuickAction,
                modifier = Modifier.fillMaxWidth()
            )
        }
    }
}

@Composable
fun TabletDashboardLayout(
    state: DashboardState,
    onKpiClick: (KpiType) -> Unit,
    onAgentClick: (String) -> Unit,
    onQuickAction: (QuickAction) -> Unit,
    onRetry: () -> Unit
) {
    Row(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        // Left column: main content
        LazyColumn(
            modifier = Modifier.weight(1.5f),
            verticalArrangement = Arrangement.spacedBy(16.dp),
            contentPadding = PaddingValues(end = 8.dp)
        ) {
            item {
                LazyRow(
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    items(state.kpis) { kpi ->
                        KpiCard(kpi = kpi, onClick = { onKpiClick(kpi.type) }, Modifier.width(180.dp))
                    }
                }
            }
            item { AiUsageChart(data = state.aiUsage, Modifier.fillMaxWidth()) }
            item { TokenConsumptionChart(data = state.tokenConsumption, Modifier.fillMaxWidth()) }
            item { ModelPerformanceChart(data = state.modelPerformance, Modifier.fillMaxWidth()) }
            item { SystemHealthIndicators(health = state.systemHealth, Modifier.fillMaxWidth()) }
            item { QuickActionsRow(onAction = onQuickAction, Modifier.fillMaxWidth()) }
        }

        Spacer(modifier = Modifier.width(16.dp))

        // Right column: sidebar
        LazyColumn(
            modifier = Modifier.weight(1f),
            verticalArrangement = Arrangement.spacedBy(16.dp),
            contentPadding = PaddingValues(start = 8.dp)
        ) {
            item {
                ActiveAgentsWidget(
                    agents = state.activeAgents,
                    onAgentClick = onAgentClick,
                    Modifier.fillMaxWidth()
                )
            }
            item {
                RecentActivityFeed(
                    activities = state.recentActivity,
                    Modifier.fillMaxWidth()
                )
            }
        }
    }
}
```

### Window Size Classes

| Class              | Width Range      | Layout                     |
|--------------------|------------------|----------------------------|
| Compact            | 0–599dp          | Single column, phone       |
| Medium             | 600–839dp        | Two columns, compact tablet|
| Expanded           | 840dp+           | Two columns, full tablet   |

---

## Accessibility

### Content Descriptions

```kotlin
// KPI card
Icon(
    imageVector = kpi.icon,
    contentDescription = "${kpi.label}: ${kpi.formattedValue}, trending ${kpi.trend.name}"
)

// Health gauge
Canvas(modifier = Modifier.semantics {
    stateDescription = "${label} usage at ${(usage * 100).toInt()} percent"
})

// Agent status
Box(
    modifier = Modifier.semantics {
        stateDescription = if (agent.isOnline) "Online" else "Offline"
    }
)

// Activity event
Row(modifier = Modifier.semantics(mergeDescendants = true) {
    contentDescription = "${event.title}. ${event.description}. ${event.timestamp.toRelativeTimeString()}"
})
```

### TalkBack

```kotlin
// Custom live region for real-time updates
Text(
    text = "Dashboard updated. Current CPU: ${(health.cpuUsage * 100).toInt()}%",
    modifier = Modifier.semantics {
        liveRegion = LiveRegionMode.Polite
    }
)

// Label all interactive elements
Button(
    onClick = { ... },
    modifier = Modifier.semantics {
        role = Role.Button
    }
) {
    Text("Refresh Dashboard")
}
```

### Color Contrast

All text meets WCAG AA contrast ratios (4.5:1 for body text, 3:1 for large text).
Health gauge colors are paired with explicit percentage labels so color-blind users
can distinguish states.

---

## Performance

### Lazy Loading and Pagination

```kotlin
// Activity feed pagination
@Composable
fun PaginatedActivityFeed(
    viewModel: DashboardViewModel
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()

    LazyColumn(state = listState) {
        items(state.recentActivity, key = { it.id }) { event ->
            ActivityEventRow(event = event)
        }

        // Trigger next page when near end
        if (state.hasMoreActivity) {
            item {
                LaunchedEffect(Unit) {
                    viewModel.onAction(DashboardAction.LoadMoreActivity)
                }
                CircularProgressIndicator(
                    modifier = Modifier.padding(16.dp).size(24.dp)
                )
            }
        }
    }
}
```

### Key Performance Metrics

| Metric                        | Target     | Measurement              |
|-------------------------------|------------|--------------------------|
| Initial render                | < 500ms    | Time to first frame      |
| Chart render                  | < 200ms    | Canvas draw time         |
| Shimmer duration              | 1.2s       | Animation loop           |
| WebSocket reconnect           | 1–30s      | Exponential backoff      |
| Cache read                    | < 10ms     | Room query               |
| LazyColumn scroll FPS         | 60fps      | Frame drops < 1%         |
| Data refresh                  | < 2s       | Network + render         |

### Debouncing

```kotlin
// Debounce filter changes
private val filterDebounce = debounce<DashboardFilters>(300L)

fun onAction(action: DashboardAction) {
    when (action) {
        is DashboardAction.ApplyFilters -> {
            filterDebounce { filters ->
                viewModelScope.launch {
                    dashboardPreferences.saveFilters(filters)
                    loadDashboard()
                }
            }
        }
        // ...
    }
}
```

### Image and Icon Caching

Vector icons from Material Icons are resolved locally. Sparkline data is kept
as `List<Float>` to avoid object allocation during recomposition. Charts use
`remember` with stable keys to avoid unnecessary redraws.

---

## Summary

| Component              | Composable                    | Data Source             |
|------------------------|-------------------------------|-------------------------|
| KPI Cards              | `KpiCardsRow` / `KpiCard`     | MetricsRepository       |
| AI Usage Chart         | `AiUsageChart`                | MetricsRepository       |
| Token Chart            | `TokenConsumptionChart`       | MetricsRepository       |
| Model Performance      | `ModelPerformanceChart`       | MetricsRepository       |
| Active Agents          | `ActiveAgentsWidget`          | AgentRepository         |
| Activity Feed          | `RecentActivityFeed`          | ActivityRepository      |
| System Health          | `SystemHealthIndicators`      | HealthRepository        |
| Quick Actions          | `QuickActionsRow`             | Local                   |
| Filters                | `DashboardFiltersSheet`       | DashboardPreferences    |
| Skeletons              | `DashboardSkeleton`           | None (UI only)          |
| Error States           | `DashboardErrorState`         | ViewModel               |

All state is managed through `DashboardViewModel` with a single `StateFlow<DashboardState>`.
Real-time updates arrive via `LiveMetricsSocket` and are merged into state. Caching is
handled through Room for data and DataStore for preferences.
