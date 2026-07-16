# Dashboard

The dashboard is the primary landing screen of the Nexus AI iOS app. It provides at-a-glance visibility into AI usage, agent status, system health, and recent activity. This document covers every aspect of the dashboard feature.

## Architecture Overview

```
┌──────────────────────────────────────────────────────────┐
│                     DashboardView                        │
├──────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │ KPI Cards   │  │ AI Usage    │  │ Token       │     │
│  │ (4 metrics) │  │ Chart       │  │ Chart       │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │ Model Perf  │  │ Active      │  │ System      │     │
│  │ Chart       │  │ Agents      │  │ Health      │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
│  ┌──────────────────────────────────────────────────┐   │
│  │ Recent Activity Feed                             │   │
│  └──────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────┐   │
│  │ Quick Actions                                    │   │
│  └──────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────┘
```

## Data Flow

```
┌──────────┐    ┌────────────────┐    ┌──────────────┐
│ API      │───▶│ DashboardVM    │───▶│ DashboardView│
│ Service  │    │ (ObservableObj)│    │ (SwiftUI)    │
└──────────┘    └────────────────┘    └──────────────┘
                       │                      │
                       ▼                      ▼
                ┌────────────┐       ┌──────────────┐
                │ Dashboard  │       │ Charts, Cards│
                │ Cache      │       │ Widgets      │
                └────────────┘       └──────────────┘
```

## Dashboard Data Models

```swift
struct DashboardData: Codable {
    let kpis: KPIData
    let aiUsage: AIUsageData
    let tokenConsumption: TokenConsumptionData
    let modelPerformance: [ModelPerformanceData]
    let activeAgents: [AgentStatusData]
    let recentActivity: [ActivityItem]
    let systemHealth: SystemHealthData
    let lastUpdated: Date
}

struct KPIData: Codable {
    let totalRequests: Int
    let activeUsers: Int
    let modelUsage: Double
    let totalCost: Double
    let requestsTrend: Double
    let usersTrend: Double
    let usageTrend: Double
    let costTrend: Double

    enum CodingKeys: String, CodingKey {
        case totalRequests = "total_requests"
        case activeUsers = "active_users"
        case modelUsage = "model_usage"
        case totalCost = "total_cost"
        case requestsTrend = "requests_trend"
        case usersTrend = "users_trend"
        case usageTrend = "usage_trend"
        case costTrend = "cost_trend"
    }
}

struct AIUsageData: Codable {
    let daily: [UsagePoint]
    let weekly: [UsagePoint]
    let monthly: [UsagePoint]
}

struct UsagePoint: Codable, Identifiable {
    let id = UUID()
    let date: Date
    let requests: Int
    let successful: Int
    let failed: Int

    enum CodingKeys: String, CodingKey {
        case date, requests, successful, failed
    }
}

struct TokenConsumptionData: Codable {
    let byModel: [ModelTokenUsage]
    let byDay: [DailyTokenUsage]
}

struct ModelTokenUsage: Codable, Identifiable {
    let id = UUID()
    let modelName: String
    let inputTokens: Int
    let outputTokens: Int

    enum CodingKeys: String, CodingKey {
        case modelName = "model_name"
        case inputTokens = "input_tokens"
        case outputTokens = "output_tokens"
    }
}

struct DailyTokenUsage: Codable, Identifiable {
    let id = UUID()
    let date: Date
    let inputTokens: Int
    let outputTokens: Int

    enum CodingKeys: String, CodingKey {
        case date
        case inputTokens = "input_tokens"
        case outputTokens = "output_tokens"
    }
}

struct ModelPerformanceData: Codable, Identifiable {
    let id = UUID()
    let modelName: String
    let avgLatency: Double
    let p95Latency: Double
    let throughput: Double
    let errorRate: Double
    let totalRequests: Int

    enum CodingKeys: String, CodingKey {
        case modelName = "model_name"
        case avgLatency = "avg_latency"
        case p95Latency = "p95_latency"
        case throughput
        case errorRate = "error_rate"
        case totalRequests = "total_requests"
    }
}

struct AgentStatusData: Codable, Identifiable {
    let id: String
    let name: String
    let status: AgentStatus
    let currentTasks: Int
    let completedTasks: Int
    let avgLatency: Double
    let lastActive: Date

    enum CodingKeys: String, CodingKey {
        case id, name, status
        case currentTasks = "current_tasks"
        case completedTasks = "completed_tasks"
        case avgLatency = "avg_latency"
        case lastActive = "last_active"
    }
}

enum AgentStatus: String, Codable {
    case online
    case offline
    case busy
    case error
}

struct ActivityItem: Codable, Identifiable {
    let id: String
    let type: ActivityType
    let title: String
    let description: String
    let timestamp: Date
    let agentId: String?
    let metadata: [String: String]?

    enum CodingKeys: String, CodingKey {
        case id, type, title, description, timestamp
        case agentId = "agent_id"
        case metadata
    }
}

enum ActivityType: String, Codable {
    case conversationStarted = "conversation_started"
    case documentProcessed = "document_processed"
    case toolExecution = "tool_execution"
    case agentCreated = "agent_created"
    case systemAlert = "system_alert"
}

struct SystemHealthData: Codable {
    let cpu: Double
    let ram: Double
    let gpu: Double
    let storage: Double
    let networkLatency: Double
    let uptime: TimeInterval
    let status: HealthStatus

    enum CodingKeys: String, CodingKey {
        case cpu, ram, gpu, storage
        case networkLatency = "network_latency"
        case uptime, status
    }
}

enum HealthStatus: String, Codable {
    case healthy
    case degraded
    case critical
}

struct DashboardFilters: Equatable {
    var dateRange: DateRange
    var tenantId: String?
    var userId: String?
    var modelId: String?

    static let `default` = DashboardFilters(
        dateRange: .last7Days,
        tenantId: nil,
        userId: nil,
        modelId: nil
    )
}

enum DateRange: String, CaseIterable, Identifiable {
    case last24Hours = "24 Hours"
    case last7Days = "7 Days"
    case last30Days = "30 Days"
    case last90Days = "90 Days"
    case custom = "Custom"

    var id: String { rawValue }
}
```

## Dashboard ViewModel

```swift
import SwiftUI
import Combine

@MainActor
final class DashboardViewModel: ObservableObject {
    @Published var dashboardData: DashboardData?
    @Published var isLoading = false
    @Published var error: DashboardError?
    @Published var filters = DashboardFilters.default
    @Published var lastRefreshDate: Date?
    @Published var isRefreshing = false

    private let apiService: DashboardAPIService
    private let cacheService: DashboardCacheService
    private let webSocketService: LiveMetricsService
    private var cancellables = Set<AnyCancellable>()

    init(
        apiService: DashboardAPIService = .shared,
        cacheService: DashboardCacheService = .shared,
        webSocketService: LiveMetricsService = .shared
    ) {
        self.apiService = apiService
        self.cacheService = cacheService
        self.webSocketService = webSocketService
        setupBindings()
    }

    func loadDashboard() async {
        isLoading = true
        error = nil

        if let cached = await cacheService.getCachedDashboard(filters: filters) {
            dashboardData = cached
            isLoading = false
            await refreshDashboard()
            return
        }

        do {
            let data = try await apiService.fetchDashboard(filters: filters)
            dashboardData = data
            lastRefreshDate = Date()
            await cacheService.cacheDashboard(data, filters: filters)
        } catch {
            self.error = DashboardError.from(error)
        }

        isLoading = false
    }

    func refreshDashboard() async {
        isRefreshing = true
        do {
            let data = try await apiService.fetchDashboard(filters: filters)
            dashboardData = data
            lastRefreshDate = Date()
            await cacheService.cacheDashboard(data, filters: filters)
        } catch {
            self.error = DashboardError.from(error)
        }
        isRefreshing = false
    }

    func updateFilters(_ newFilters: DashboardFilters) async {
        filters = newFilters
        await loadDashboard()
    }

    private func setupBindings() {
        webSocketService.metricsPublisher
            .receive(on: DispatchQueue.main)
            .sink { [weak self] update in
                self?.applyLiveUpdate(update)
            }
            .store(in: &cancellables)
    }

    private func applyLiveUpdate(_ update: LiveMetricsUpdate) {
        guard var data = dashboardData else { return }
        data.kpis.totalRequests += update.newRequests
        data.kpis.activeUsers = update.activeUsers
        data.systemHealth.cpu = update.cpuUsage
        data.systemHealth.ram = update.ramUsage
        dashboardData = data
    }
}

enum DashboardError: LocalizedError {
    case networkError(String)
    case serverError(Int, String?)
    case decodingError
    case unauthorized
    case unknown(String)

    var errorDescription: String? {
        switch self {
        case .networkError(let msg): return "Network error: \(msg)"
        case .serverError(let code, let msg):
            return "Server error (\(code)): \(msg ?? "Unknown")"
        case .decodingError: return "Failed to parse server response"
        case .unauthorized: return "Session expired. Please log in again."
        case .unknown(let msg): return msg
        }
    }

    static func from(_ error: Error) -> DashboardError {
        if let apiError = error as? APIError {
            switch apiError {
            case .network(let err): return .networkError(err.localizedDescription)
            case .http(let code, let msg): return .serverError(code, msg)
            case .decoding: return .decodingError
            case .unauthorized: return .unauthorized
            }
        }
        return .unknown(error.localizedDescription)
    }
}
```

## Dashboard View

```swift
import SwiftUI
import Charts

struct DashboardView: View {
    @StateObject private var viewModel = DashboardViewModel()
    @State private var showFilters = false

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.dashboardData == nil {
                    DashboardLoadingView()
                } else if let error = viewModel.error, viewModel.dashboardData == nil {
                    DashboardErrorView(error: error) {
                        Task { await viewModel.loadDashboard() }
                    }
                } else if let data = viewModel.dashboardData {
                    dashboardContent(data)
                } else {
                    DashboardEmptyView()
                }
            }
            .navigationTitle("Dashboard")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button(action: { showFilters = true }) {
                        Image(systemName: "line.3.horizontal.decrease.circle")
                    }
                }
            }
            .sheet(isPresented: $showFilters) {
                DashboardFilterSheet(filters: viewModel.filters) { newFilters in
                    Task { await viewModel.updateFilters(newFilters) }
                }
            }
            .refreshable {
                await viewModel.refreshDashboard()
            }
            .task {
                await viewModel.loadDashboard()
            }
        }
    }

    private func dashboardContent(_ data: DashboardData) -> some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                lastUpdatedHeader(data.lastUpdated)
                KPICardsView(kpis: data.kpis)
                AIUsageChartView(usage: data.aiUsage)
                TokenConsumptionChartView(consumption: data.tokenConsumption)
                ModelPerformanceChartView(performance: data.modelPerformance)
                ActiveAgentsWidgetView(agents: data.activeAgents)
                SystemHealthIndicatorsView(health: data.systemHealth)
                RecentActivityFeedView(items: data.recentActivity)
                QuickActionsView()
            }
            .padding()
        }
    }

    private func lastUpdatedHeader(_ date: Date) -> some View {
        HStack {
            Text("Last updated: \(date.formatted(.relative(presentation: .named)))")
                .font(.caption)
                .foregroundStyle(.secondary)
            Spacer()
            if viewModel.isRefreshing {
                ProgressView()
                    .scaleEffect(0.8)
            }
        }
    }
}
```

## KPI Cards View

```swift
struct KPICardsView: View {
    let kpis: KPIData

    private let columns = [
        GridItem(.flexible()),
        GridItem(.flexible())
    ]

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Key Metrics")
                .font(.headline)

            LazyVGrid(columns: columns, spacing: 12) {
                KPICard(
                    title: "Total Requests",
                    value: "\(kpis.totalRequests.formatted())",
                    trend: kpis.requestsTrend,
                    icon: "arrow.up.arrow.down",
                    color: .blue
                )
                KPICard(
                    title: "Active Users",
                    value: "\(kpis.activeUsers)",
                    trend: kpis.usersTrend,
                    icon: "person.2.fill",
                    color: .green
                )
                KPICard(
                    title: "Model Usage",
                    value: String(format: "%.1f%%", kpis.modelUsage),
                    trend: kpis.usageTrend,
                    icon: "brain.head.profile",
                    color: .purple
                )
                KPICard(
                    title: "Total Cost",
                    value: String(format: "$%.2f", kpis.totalCost),
                    trend: kpis.costTrend,
                    icon: "dollarsign.circle.fill",
                    color: .orange
                )
            }
        }
    }
}

struct KPICard: View {
    let title: String
    let value: String
    let trend: Double
    let icon: String
    let color: Color

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: icon)
                    .foregroundStyle(color)
                Spacer()
                trendBadge
            }
            Text(value)
                .font(.title2.bold())
                .foregroundStyle(.primary)
            Text(title)
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .accessibilityElement(children: .combine)
        .accessibilityLabel("\(title): \(value), trend \(trend > 0 ? "up" : "down") \(String(format: "%.1f", abs(trend))) percent")
    }

    private var trendBadge: some View {
        HStack(spacing: 2) {
            Image(systemName: trend >= 0 ? "arrow.up" : "arrow.down")
            Text(String(format: "%.1f%%", abs(trend)))
        }
        .font(.caption2.bold())
        .foregroundStyle(trend >= 0 ? .green : .red)
        .padding(.horizontal, 6)
        .padding(.vertical, 2)
        .background(trend >= 0 ? Color.green.opacity(0.1) : Color.red.opacity(0.1))
        .clipShape(Capsule())
    }
}
```

## AI Usage Chart View

```swift
struct AIUsageChartView: View {
    let usage: AIUsageData
    @State private var selectedTimeframe: Timeframe = .daily

    enum Timeframe: String, CaseIterable {
        case daily, weekly, monthly
    }

    private var chartData: [UsagePoint] {
        switch selectedTimeframe {
        case .daily: return usage.daily
        case .weekly: return usage.weekly
        case .monthly: return usage.monthly
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("AI Usage")
                    .font(.headline)
                Spacer()
                Picker("Timeframe", selection: $selectedTimeframe) {
                    ForEach(Timeframe.allCases, id: \.self) { Text($0.rawValue.capitalized) }
                }
                .pickerStyle(.segmented)
                .frame(width: 200)
            }

            if chartData.isEmpty {
                ChartEmptyPlaceholder()
            } else {
                Chart(chartData) { point in
                    LineMark(
                        x: .value("Date", point.date),
                        y: .value("Requests", point.requests)
                    )
                    .foregroundStyle(.blue)
                    .interpolationMethod(.catmullRom)

                    AreaMark(
                        x: .value("Date", point.date),
                        y: .value("Requests", point.requests)
                    )
                    .foregroundStyle(.blue.opacity(0.15))
                    .interpolationMethod(.catmullRom)

                    RuleMark(y: .value("Average", averageRequests))
                        .foregroundStyle(.red.opacity(0.5))
                        .lineStyle(StrokeStyle(lineWidth: 1, dash: [5]))
                }
                .chartYAxis {
                    AxisMarks(position: .leading)
                }
                .chartOverlay { proxy in
                    GeometryReader { geometry in
                        Rectangle()
                            .fill(.clear)
                            .contentShape(Rectangle())
                            .onTapGesture { location in
                                handleChartTap(location: location, proxy: proxy, geometry: geometry)
                            }
                    }
                }
                .frame(height: 200)
            }
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var averageRequests: Double {
        guard !chartData.isEmpty else { return 0 }
        return Double(chartData.map(\.requests).reduce(0, +)) / Double(chartData.count)
    }

    private func handleChartTap(location: CGPoint, proxy: ChartProxy, geometry: GeometryProxy) {
        let origin = geometry[proxy.plotAreaFrame].origin
        let location = CGPoint(
            x: location.x - origin.x,
            y: location.y - origin.y
        }
        if let date: Date = proxy.value(atX: location.x) {
            print("Tapped date: \(date)")
        }
    }
}
```

## Token Consumption Chart View

```swift
struct TokenConsumptionChartView: View {
    let consumption: TokenConsumptionData
    @State private var chartType: ChartType = .stackedBar

    enum ChartType {
        case stackedBar, pie, line
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Token Consumption")
                    .font(.headline)
                Spacer()
                Menu {
                    Button("Stacked Bar") { chartType = .stackedBar }
                    Button("Pie Chart") { chartType = .pie }
                    Button("Line Chart") { chartType = .line }
                } label: {
                    Image(systemName: "chart.xyaxis.line")
                }
            }

            legend

            if chartType == .stackedBar {
                stackedBarChart
            } else if chartType == .line {
                lineChart
            } else {
                pieChart
            }
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var legend: some View {
        HStack(spacing: 16) {
            HStack(spacing: 4) {
                Circle().fill(.blue).frame(width: 8, height: 8)
                Text("Input Tokens").font(.caption)
            }
            HStack(spacing: 4) {
                Circle().fill(.green).frame(width: 8, height: 8)
                Text("Output Tokens").font(.caption)
            }
            Spacer()
            Text("Total: \(totalTokens.formatted())")
                .font(.caption.bold())
        }
    }

    private var totalTokens: Int {
        consumption.byDay.reduce(0) { $0 + $1.inputTokens + $1.outputTokens }
    }

    private var stackedBarChart: some View {
        Chart(consumption.byDay) { day in
            BarMark(
                x: .value("Date", day.date, unit: .day),
                y: .value("Tokens", day.inputTokens),
                stacking: .standard
            )
            .foregroundStyle(.blue)

            BarMark(
                x: .value("Date", day.date, unit: .day),
                y: .value("Tokens", day.outputTokens),
                stacking: .standard
            )
            .foregroundStyle(.green)
        }
        .chartYAxis {
            AxisMarks(position: .leading)
        }
        .frame(height: 200)
    }

    private var lineChart: some View {
        Chart(consumption.byDay) { day in
            LineMark(
                x: .value("Date", day.date, unit: .day),
                y: .value("Tokens", day.inputTokens)
            )
            .foregroundStyle(.blue)

            LineMark(
                x: .value("Date", day.date, unit: .day),
                y: .value("Tokens", day.outputTokens)
            )
            .foregroundStyle(.green)
        }
        .frame(height: 200)
    }

    private var pieChart: some View {
        let totalInput = consumption.byModel.reduce(0) { $0 + $1.inputTokens }
        let totalOutput = consumption.byModel.reduce(0) { $0 + $1.outputTokens }

        return Chart {
            SectorMark(angle: .value("Input", totalInput))
                .foregroundStyle(.blue)
                .annotation(position: .overlay) {
                    Text("Input\n\(totalInput.formatted())")
                        .font(.caption2)
                        .multilineTextAlignment(.center)
                }
            SectorMark(angle: .value("Output", totalOutput))
                .foregroundStyle(.green)
                .annotation(position: .overlay) {
                    Text("Output\n\(totalOutput.formatted())")
                        .font(.caption2)
                        .multilineTextAlignment(.center)
                }
        }
        .frame(height: 200)
    }
}
```

## Model Performance Chart View

```swift
struct ModelPerformanceChartView: View {
    let performance: [ModelPerformanceData]
    @State private var selectedMetric: Metric = .avgLatency

    enum Metric: String, CaseIterable {
        case avgLatency = "Avg Latency"
        case p95Latency = "P95 Latency"
        case throughput = "Throughput"
        case errorRate = "Error Rate"
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Model Performance")
                    .font(.headline)
                Spacer()
                Picker("Metric", selection: $selectedMetric) {
                    ForEach(Metric.allCases, id: \.self) { Text($0.rawValue) }
                }
                .pickerStyle(.menu)
            }

            Chart(performance) { model in
                BarMark(
                    x: .value("Model", model.modelName),
                    y: .value(selectedMetric.rawValue, metricValue(for: model))
                )
                .foregroundStyle(colorForMetric)
                .annotation(position: .top) {
                    Text(metricLabel(for: model))
                        .font(.caption2)
                        .foregroundStyle(.secondary)
                }
            }
            .frame(height: 200)
            .chartYAxis {
                AxisMarks(position: .leading)
            }
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func metricValue(for model: ModelPerformanceData) -> Double {
        switch selectedMetric {
        case .avgLatency: return model.avgLatency
        case .p95Latency: return model.p95Latency
        case .throughput: return model.throughput
        case .errorRate: return model.errorRate
        }
    }

    private func metricLabel(for model: ModelPerformanceData) -> String {
        let value = metricValue(for: model)
        switch selectedMetric {
        case .avgLatency, .p95Latency: return String(format: "%.0fms", value)
        case .throughput: return String(format: "%.1f req/s", value)
        case .errorRate: return String(format: "%.1f%%", value)
        }
    }

    private var colorForMetric: Color {
        switch selectedMetric {
        case .avgLatency: return .blue
        case .p95Latency: return .orange
        case .throughput: return .green
        case .errorRate: return .red
        }
    }
}
```

## Active Agents Widget View

```swift
struct ActiveAgentsWidgetView: View {
    let agents: [AgentStatusData]

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Active Agents")
                    .font(.headline)
                Spacer()
                Text("\(agents.filter { $0.status == .online }.count) online")
                    .font(.caption)
                    .foregroundStyle(.green)
            }

            if agents.isEmpty {
                Text("No agents configured")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, minHeight: 80)
            } else {
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 12) {
                        ForEach(agents) { agent in
                            AgentStatusCard(agent: agent)
                        }
                    }
                }
            }
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct AgentStatusCard: View {
    let agent: AgentStatusData

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Circle()
                    .fill(statusColor)
                    .frame(width: 8, height: 8)
                Text(agent.name)
                    .font(.subheadline.bold())
                    .lineLimit(1)
            }

            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 4) {
                    Image(systemName: "cpu")
                        .font(.caption2)
                    Text("\(agent.currentTasks) tasks")
                        .font(.caption2)
                }
                HStack(spacing: 4) {
                    Image(systemName: "clock")
                        .font(.caption2)
                    Text("\(Int(agent.avgLatency))ms avg")
                        .font(.caption2)
                }
            }
            .foregroundStyle(.secondary)

            Text("Active \(agent.lastActive.formatted(.relative(presentation: .named)))")
                .font(.caption2)
                .foregroundStyle(.tertiary)
        }
        .padding(12)
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 10))
        .frame(width: 140)
    }

    private var statusColor: Color {
        switch agent.status {
        case .online: return .green
        case .offline: return .gray
        case .busy: return .orange
        case .error: return .red
        }
    }
}
```

## Recent Activity Feed View

```swift
struct RecentActivityFeedView: View {
    let items: [ActivityItem]

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Recent Activity")
                    .font(.headline)
                Spacer()
                Button("See All") { }
                    .font(.subheadline)
            }

            if items.isEmpty {
                Text("No recent activity")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, minHeight: 80)
            } else {
                ForEach(items.prefix(10)) { item in
                    ActivityRow(item: item)
                    if item.id != items.prefix(10).last?.id {
                        Divider()
                    }
                }
            }
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct ActivityRow: View {
    let item: ActivityItem

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: iconForType(item.type))
                .foregroundStyle(colorForType(item.type))
                .frame(width: 24)

            VStack(alignment: .leading, spacing: 2) {
                Text(item.title)
                    .font(.subheadline.bold())
                Text(item.description)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
            }

            Spacer()

            Text(item.timestamp.formatted(.relative(presentation: .named)))
                .font(.caption2)
                .foregroundStyle(.tertiary)
        }
    }

    private func iconForType(_ type: ActivityType) -> String {
        switch type {
        case .conversationStarted: return "bubble.left.fill"
        case .documentProcessed: return "doc.fill"
        case .toolExecution: return "wrench.and.screwdriver.fill"
        case .agentCreated: return "person.badge.plus"
        case .systemAlert: return "exclamationmark.triangle.fill"
        }
    }

    private func colorForType(_ type: ActivityType) -> Color {
        switch type {
        case .conversationStarted: return .blue
        case .documentProcessed: return .green
        case .toolExecution: return .orange
        case .agentCreated: return .purple
        case .systemAlert: return .red
        }
    }
}
```

## System Health Indicators View

```swift
struct SystemHealthIndicatorsView: View {
    let health: SystemHealthData

    private let columns = [
        GridItem(.flexible()),
        GridItem(.flexible()),
        GridItem(.flexible()),
        GridItem(.flexible())
    ]

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("System Health")
                    .font(.headline)
                Spacer()
                HealthStatusBadge(status: health.status)
            }

            LazyVGrid(columns: columns, spacing: 12) {
                HealthGauge(value: health.cpu, label: "CPU", icon: "cpu")
                HealthGauge(value: health.ram, label: "RAM", icon: "memorychip")
                HealthGauge(value: health.gpu, label: "GPU", icon: "gpu")
                HealthGauge(value: health.storage, label: "Storage", icon: "internaldrive")
            }

            HStack {
                Label("Network: \(Int(health.networkLatency))ms", systemImage: "wifi")
                    .font(.caption)
                Spacer()
                Label("Uptime: \(formatUptime(health.uptime))", systemImage: "clock")
                    .font(.caption)
            }
            .foregroundStyle(.secondary)
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func formatUptime(_ interval: TimeInterval) -> String {
        let days = Int(interval) / 86400
        let hours = Int(interval) % 86400 / 3600
        if days > 0 { return "\(days)d \(hours)m" }
        return "\(hours)m"
    }
}

struct HealthGauge: View {
    let value: Double
    let label: String
    let icon: String

    private var gaugeColor: Color {
        if value < 60 { return .green }
        if value < 85 { return .orange }
        return .red
    }

    var body: some View {
        VStack(spacing: 6) {
            Gauge(value: value, in: 0...100) {
                Image(systemName: icon)
                    .font(.caption2)
            }
            .gaugeStyle(.accessoryCircular)
            .tint(gaugeColor)

            Text(label)
                .font(.caption2.bold())
            Text("\(Int(value))%")
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
    }
}

struct HealthStatusBadge: View {
    let status: HealthStatus

    var body: some View {
        HStack(spacing: 4) {
            Circle()
                .fill(color)
                .frame(width: 6, height: 6)
            Text(status.rawValue.capitalized)
                .font(.caption2.bold())
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(color.opacity(0.15))
        .clipShape(Capsule())
    }

    private var color: Color {
        switch status {
        case .healthy: return .green
        case .degraded: return .orange
        case .critical: return .red
        }
    }
}
```

## Quick Actions View

```swift
struct QuickActionsView: View {
    @State private var showNewChat = false
    @State private var showUploadDocument = false
    @State private var showAlerts = false

    private let actions: [QuickAction] = [
        QuickAction(icon: "bubble.left.fill", title: "Start Chat", color: .blue),
        QuickAction(icon: "arrow.up.doc.fill", title: "Upload Document", color: .green),
        QuickAction(icon: "bell.badge.fill", title: "View Alerts", color: .red)
    ]

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Quick Actions")
                .font(.headline)

            HStack(spacing: 12) {
                ForEach(actions) { action in
                    Button {
                        handleAction(action)
                    } label: {
                        VStack(spacing: 8) {
                            Image(systemName: action.icon)
                                .font(.title2)
                                .foregroundStyle(action.color)
                            Text(action.title)
                                .font(.caption.bold())
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 16)
                        .background(action.color.opacity(0.1))
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                    }
                }
            }
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func handleAction(_ action: QuickAction) {
        switch action.title {
        case "Start Chat": showNewChat = true
        case "Upload Document": showUploadDocument = true
        case "View Alerts": showAlerts = true
        default: break
        }
    }
}

struct QuickAction: Identifiable {
    let id = UUID()
    let icon: String
    let title: String
    let color: Color
}
```

## Dashboard Filters

```swift
struct DashboardFilterSheet: View {
    @Environment(\.dismiss) private var dismiss
    @State var filters: DashboardFilters
    let onApply: (DashboardFilters) -> Void

    @State private var customStartDate = Date()
    @State private var customEndDate = Date()

    var body: some View {
        NavigationStack {
            Form {
                Section("Date Range") {
                    Picker("Range", selection: $filters.dateRange) {
                        ForEach(DateRange.allCases) { range in
                            Text(range.rawValue).tag(range)
                        }
                    }
                    .pickerStyle(.inline)

                    if filters.dateRange == .custom {
                        DatePicker("Start", selection: $customStartDate, displayedComponents: .date)
                        DatePicker("End", selection: $customEndDate, displayedComponents: .date)
                    }
                }

                Section("Filters") {
                    TextField("Tenant ID", text: Binding(
                        get: { filters.tenantId ?? "" },
                        set: { filters.tenantId = $0.isEmpty ? nil : $0 }
                    ))

                    TextField("User ID", text: Binding(
                        get: { filters.userId ?? "" },
                        set: { filters.userId = $0.isEmpty ? nil : $0 }
                    ))

                    TextField("Model ID", text: Binding(
                        get: { filters.modelId ?? "" },
                        set: { filters.modelId = $0.isEmpty ? nil : $0 }
                    ))
                }
            }
            .navigationTitle("Filters")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Apply") {
                        onApply(filters)
                        dismiss()
                    }
                }
            }
        }
    }
}
```

## Loading States

```swift
struct DashboardLoadingView: View {
    @State private var isAnimating = false

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                ForEach(0..<5) { _ in
                    ShimmerCard()
                }
            }
            .padding()
        }
    }
}

struct ShimmerCard: View {
    @State private var phase: CGFloat = 0

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            RoundedRectangle(cornerRadius: 8)
                .fill(.quaternary)
                .frame(width: 120, height: 20)

            HStack(spacing: 12) {
                ForEach(0..<4) { _ in
                    RoundedRectangle(cornerRadius: 8)
                        .fill(.quaternary)
                        .frame(height: 80)
                }
            }
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .overlay(
            LinearGradient(
                colors: [.clear, .white.opacity(0.4), .clear],
                startPoint: .leading,
                endPoint: .trailing
            )
            .rotationEffect(.degrees(20))
            .offset(x: phase)
            .clipped()
        )
        .onAppear {
            withAnimation(.linear(duration: 1.5).repeatForever(autoreverses: false)) {
                phase = 400
            }
        }
    }
}
```

## Error and Empty States

```swift
struct DashboardErrorView: View {
    let error: DashboardError
    let onRetry: () -> Void

    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 48))
                .foregroundStyle(.orange)

            Text("Something went wrong")
                .font(.title2.bold())

            Text(error.localizedDescription)
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)

            Button(action: onRetry) {
                Label("Retry", systemImage: "arrow.clockwise")
            }
            .buttonStyle(.borderedProminent)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct DashboardEmptyView: View {
    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "chart.bar.doc.horizontal")
                .font(.system(size: 48))
                .foregroundStyle(.secondary)

            Text("No Data Available")
                .font(.title2.bold())

            Text("Start using the platform to see your dashboard data.")
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct ChartEmptyPlaceholder: View {
    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: "chart.line.downtrend.xyaxis")
                .font(.title2)
                .foregroundStyle(.secondary)
            Text("No data for this period")
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .frame(height: 120)
        .frame(maxWidth: .infinity)
    }
}
```

## Dashboard API Integration

```swift
protocol DashboardAPIService {
    func fetchDashboard(filters: DashboardFilters) async throws -> DashboardData
    func fetchMetrics(filters: DashboardFilters) async throws -> [MetricPoint]
}

final class DefaultDashboardAPIService: DashboardAPIService {
    static let shared = DefaultDashboardAPIService()
    private let networkClient: NetworkClient

    init(networkClient: NetworkClient = .shared) {
        self.networkClient = networkClient
    }

    func fetchDashboard(filters: DashboardFilters) async throws -> DashboardData {
        var components = URLComponents(string: "\(APIConfig.baseURL)/api/v1/dashboard")!
        var queryItems: [URLQueryItem] = [
            URLQueryItem(name: "date_range", value: filters.dateRange.rawValue)
        ]
        if let tenant = filters.tenantId {
            queryItems.append(URLQueryItem(name: "tenant_id", value: tenant))
        }
        if let user = filters.userId {
            queryItems.append(URLQueryItem(name: "user_id", value: user))
        }
        if let model = filters.modelId {
            queryItems.append(URLQueryItem(name: "model_id", value: model))
        }
        components.queryItems = queryItems

        let request = try URLRequest(url: components.url!, method: .get)
        let (data, response) = try await networkClient.execute(request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.network(NSError(domain: "", code: -1))
        }
        switch httpResponse.statusCode {
        case 200...299:
            return try JSONDecoder().decode(DashboardData.self, from: data)
        case 401:
            throw APIError.unauthorized
        default:
            throw APIError.http(httpResponse.statusCode, String(data: data, encoding: .utf8))
        }
    }

    func fetchMetrics(filters: DashboardFilters) async throws -> [MetricPoint] {
        var components = URLComponents(string: "\(APIConfig.baseURL)/api/v1/metrics")!
        components.queryItems = [
            URLQueryItem(name: "date_range", value: filters.dateRange.rawValue)
        ]

        let request = try URLRequest(url: components.url!, method: .get)
        let (data, _) = try await networkClient.execute(request)
        return try JSONDecoder().decode([MetricPoint].self, from: data)
    }
}
```

## Live Metrics via WebSocket

```swift
struct LiveMetricsUpdate: Codable {
    let newRequests: Int
    let activeUsers: Int
    let cpuUsage: Double
    let ramUsage: Double
    let gpuUsage: Double
    let timestamp: Date

    enum CodingKeys: String, CodingKey {
        case newRequests = "new_requests"
        case activeUsers = "active_users"
        case cpuUsage = "cpu_usage"
        case ramUsage = "ram_usage"
        case gpuUsage = "gpu_usage"
        case timestamp
    }
}

final class LiveMetricsService: ObservableObject {
    static let shared = LiveMetricsService()
    private let webSocket: WebSocketClient

    private var metricsSubject = PassthroughSubject<LiveMetricsUpdate, Never>()
    var metricsPublisher: AnyPublisher<LiveMetricsUpdate, Never> {
        metricsSubject.eraseToAnyPublisher()
    }

    init(webSocket: WebSocketClient = .shared) {
        self.webSocket = webSocket
    }

    func connect() {
        webSocket.connect(to: "wss://api.nexus.ai/ws/metrics") { [weak self] result in
            switch result {
            case .success(let data):
                if let update = try? JSONDecoder().decode(LiveMetricsUpdate.self, from: data) {
                    self?.metricsSubject.send(update)
                }
            case .failure(let error):
                print("WebSocket error: \(error)")
                DispatchQueue.main.asyncAfter(deadline: .now() + 5) {
                    self?.connect()
                }
            }
        }
    }

    func disconnect() {
        webSocket.disconnect()
    }
}
```

## Dashboard Caching

```swift
final class DashboardCacheService {
    static let shared = DashboardCacheService()
    private let cache = NSCache<NSString, CachedDashboard>()
    private let defaults = UserDefaults.standard
    private let cacheKey = "dashboard_cache"
    private let cacheExpiry: TimeInterval = 300 // 5 minutes

    func getCachedDashboard(filters: DashboardFilters) async -> DashboardData? {
        let key = cacheKeyFor(filters: filters)
        if let cached = cache.object(forKey: key as NSString) {
            guard Date().timeIntervalSince(cached.timestamp) < cacheExpiry else {
                cache.removeObject(forKey: key as NSString)
                return nil
            }
            return cached.data
        }
        return nil
    }

    func cacheDashboard(_ data: DashboardData, filters: DashboardFilters) async {
        let cached = CachedDashboard(data: data, timestamp: Date())
        cache.setObject(cached, forKey: cacheKeyFor(filters: filters) as NSString)
    }

    func clearCache() {
        cache.removeAllObjects()
    }

    private func cacheKeyFor(filters: DashboardFilters) -> String {
        "\(cacheKey)_\(filters.dateRange.rawValue)_\(filters.tenantId ?? "all")_\(filters.userId ?? "all")_\(filters.modelId ?? "all")"
    }
}

final class CachedDashboard: NSObject {
    let data: DashboardData
    let timestamp: Date

    init(data: DashboardData, timestamp: Date) {
        self.data = data
        self.timestamp = timestamp
    }
}
```

## Responsive Design

```swift
struct DashboardResponsiveLayout: View {
    @Environment(\.horizontalSizeClass) var sizeClass

    var body: some View {
        if sizeClass == .compact {
            iPhoneLayout()
        } else {
            iPadLayout()
        }
    }
}

struct iPhoneLayout: View {
    var body: some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                KPICardsView(kpis: mockKPIs)
                AIUsageChartView(usage: mockUsage)
                TokenConsumptionChartView(consumption: mockTokens)
                ModelPerformanceChartView(performance: mockPerformance)
                ActiveAgentsWidgetView(agents: mockAgents)
                SystemHealthIndicatorsView(health: mockHealth)
                RecentActivityFeedView(items: mockActivity)
                QuickActionsView()
            }
            .padding()
        }
    }
}

struct iPadLayout: View {
    let columns = [
        GridItem(.flexible()),
        GridItem(.flexible())
    ]

    var body: some View {
        ScrollView {
            LazyVGrid(columns: columns, spacing: 16) {
                KPICardsView(kpis: mockKPIs)
                AIUsageChartView(usage: mockUsage)
                TokenConsumptionChartView(consumption: mockTokens)
                ModelPerformanceChartView(performance: mockPerformance)
                ActiveAgentsWidgetView(agents: mockAgents)
                SystemHealthIndicatorsView(health: mockHealth)
                RecentActivityFeedView(items: mockActivity)
                QuickActionsView()
            }
            .padding()
        }
    }
}
```

## Accessibility

```swift
// VoiceOver support across all dashboard components
struct DashboardAccessibilityModifier: ViewModifier {
    func body(content: Content) -> some View {
        content
            .accessibilityElement(children: .contain)
            .accessibilityLabel("Dashboard")
    }
}

// KPICard accessibility enhancements
// (already included in KPICard above)

// Dynamic Type support
// All text uses system fonts which automatically scale
// Charts respect accessibility settings

// Reduce Motion support
struct AccessibleChartAnimation: ViewModifier {
    @Environment(\.accessibilityReduceMotion) var reduceMotion

    func body(content: Content) -> some View {
        content
            .animation(reduceMotion ? nil : .easeInOut, value: UUID())
    }
}

// VoiceOver rotor actions for charts
extension AIUsageChartView {
    var accessibilityCustomActions: [AccessibilityCustomAction] {
        [
            AccessibilityCustomAction("Show daily usage") {
                print("Switch to daily")
                return true
            },
            AccessibilityCustomAction("Show weekly usage") {
                print("Switch to weekly")
                return true
            }
        ]
    }
}
```

## Performance Considerations

| Aspect | Strategy |
|---|---|
| Data loading | Async/await with progressive loading |
| Caching | 5-minute in-memory cache with filter-aware keys |
| Charts | Lazy rendering, only visible chart data |
| Images | AsyncImage with placeholder |
| Memory | NSCache for automatic eviction |
| Network | WebSocket for real-time, REST for initial load |
| Pagination | Cursor-based pagination for activity feed |
| Refresh | Pull-to-refresh with debounce |
| Animations | Respect Reduce Motion |
| Prefetching | Prefetch next page of activity |

## File Structure

```
Dashboard/
├── Views/
│   ├── DashboardView.swift
│   ├── KPICardsView.swift
│   ├── AIUsageChartView.swift
│   ├── TokenConsumptionChartView.swift
│   ├── ModelPerformanceChartView.swift
│   ├── ActiveAgentsWidgetView.swift
│   ├── AgentStatusCard.swift
│   ├── RecentActivityFeedView.swift
│   ├── ActivityRow.swift
│   ├── SystemHealthIndicatorsView.swift
│   ├── HealthGauge.swift
│   ├── QuickActionsView.swift
│   ├── DashboardFilterSheet.swift
│   ├── DashboardLoadingView.swift
│   ├── DashboardErrorView.swift
│   └── DashboardEmptyView.swift
├── ViewModels/
│   └── DashboardViewModel.swift
├── Models/
│   ├── DashboardData.swift
│   ├── KPIData.swift
│   ├── UsageData.swift
│   ├── TokenData.swift
│   ├── PerformanceData.swift
│   └── HealthData.swift
├── Services/
│   ├── DashboardAPIService.swift
│   ├── DashboardCacheService.swift
│   └── LiveMetricsService.swift
└── Extensions/
    ├── DashboardAccessibility.swift
    └── DashboardResponsive.swift
```
