# Agent Management

Agent management is the core feature for creating, configuring, monitoring, and executing AI agents. This document covers every aspect of the agent management module including creation, configuration, document set binding, database connections, execution, and history.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Agent Management                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐    ┌──────────────┐    ┌────────────────┐   │
│  │ Agent    │───▶│ Agent Detail │───▶│ Agent Execute  │   │
│  │ List     │    │ (Tabs)       │    │ (Real-time)    │   │
│  └──────────┘    └──────────────┘    └────────────────┘   │
│       │                                    │               │
│       ▼                                    ▼               │
│  ┌──────────┐    ┌──────────────┐    ┌────────────────┐   │
│  │ Agent    │    │ Doc Sets     │    │ SQL Connection │   │
│  │ Create   │    │ Binding      │    │ Binding        │   │
│  └──────────┘    └──────────────┘    └────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow

```
┌────────────┐     ┌──────────────┐     ┌────────────────┐
│ API Layer  │────▶│ ViewModel    │────▶│ Views          │
│ (REST +    │     │ (Published)  │     │ (SwiftUI)      │
│ Protobuf)  │     │              │     │                │
└────────────┘     └──────────────┘     └────────────────┘
       │                  │
       ▼                  ▼
┌────────────┐     ┌──────────────┐
│ CoreData   │     │ WebSocket    │
│ (Cache)    │     │ (Exec Feed)  │
└────────────┘     └──────────────┘
```

## Data Models

```swift
// MARK: - Agent Model

struct Agent: Codable, Identifiable, Equatable {
    let id: String
    var name: String
    var description: String?
    var modelId: String
    var systemPrompt: String?
    var temperature: Double
    var maxTokens: Int
    var status: AgentStatus
    var tools: [AgentToolConfig]
    var documentSetIds: [String]
    var sqlConnectionId: String?
    var boundTables: [BoundTable]
    var createdAt: Date
    var updatedAt: Date
    var createdBy: String?

    enum CodingKeys: String, CodingKey {
        case id, name, description
        case modelId = "model_id"
        case systemPrompt = "system_prompt"
        case temperature
        case maxTokens = "max_tokens"
        case status, tools
        case documentSetIds = "document_set_ids"
        case sqlConnectionId = "sql_connection_id"
        case boundTables = "bound_tables"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case createdBy = "created_by"
    }
}

enum AgentStatus: String, Codable, CaseIterable {
    case active
    case inactive
    case error
    case training

    var displayName: String {
        switch self {
        case .active: return "Active"
        case .inactive: return "Inactive"
        case .error: return "Error"
        case .training: return "Training"
        }
    }

    var color: String {
        switch self {
        case .active: return "green"
        case .inactive: return "gray"
        case .error: return "red"
        case .training: return "orange"
        }
    }
}

// MARK: - Agent Tool Configuration

struct AgentToolConfig: Codable, Identifiable, Equatable {
    let id: String
    let toolId: String
    let name: String
    var enabled: Bool
    var parameters: [String: ToolParameterValue]

    enum ToolParameterValue: Codable, Equatable {
        case string(String)
        case int(Int)
        case double(Double)
        case bool(Bool)
    }

    enum CodingKeys: String, CodingKey {
        case id, toolId, name, enabled, parameters
    }
}

// MARK: - Document Sets

struct DocumentSet: Codable, Identifiable, Equatable {
    let id: String
    var name: String
    var description: String?
    var documentCount: Int
    var totalChunks: Int
    var createdAt: Date
    var updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id, name, description
        case documentCount = "document_count"
        case totalChunks = "total_chunks"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

// MARK: - SQL Connection

struct SQLConnection: Codable, Identifiable, Equatable {
    let id: String
    var name: String
    var type: DatabaseType
    var host: String
    var port: Int
    var database: String
    var username: String
    var isConnected: Bool
    var lastTested: Date?
    var createdAt: Date

    enum CodingKeys: String, CodingKey {
        case id, name, type, host, port, database, username
        case isConnected = "is_connected"
        case lastTested = "last_tested"
        case createdAt = "created_at"
    }
}

enum DatabaseType: String, Codable, CaseIterable {
    case postgresql = "PostgreSQL"
    case mysql = "MySQL"
    case sqlite = "SQLite"
    case mssql = "MS SQL Server"

    var defaultPort: Int {
        switch self {
        case .postgresql: return 5432
        case .mysql: return 3306
        case .sqlite: return 0
        case .mssql: return 1433
        }
    }

    var iconName: String {
        switch self {
        case .postgresql: return "cylinder.fill"
        case .mysql: return "cylinder.fill"
        case .sqlite: return "document.fill"
        case .mssql: return "server.rack"
        }
    }
}

// MARK: - Table Binding

struct BoundTable: Codable, Identifiable, Equatable {
    let id: String
    let tableName: String
    var permissionLevel: TablePermission
    var selectedColumns: [String]?
    var rowEstimate: Int?
    var columnDetails: [ColumnInfo]?

    enum CodingKeys: String, CodingKey {
        case id, tableName, permissionLevel
        case selectedColumns = "selected_columns"
        case rowEstimate = "row_estimate"
        case columnDetails = "column_details"
    }
}

enum TablePermission: String, Codable, CaseIterable {
    case readOnly = "read_only"
    case readWrite = "read_write"
    case fullAccess = "full_access"

    var displayName: String {
        switch self {
        case .readOnly: return "Read Only"
        case .readWrite: return "Read & Write"
        case .fullAccess: return "Full Access"
        }
    }
}

struct ColumnInfo: Codable, Identifiable, Equatable {
    let id: String
    let name: String
    let dataType: String
    let isNullable: Bool
    let isPrimaryKey: Bool
    let description: String?

    enum CodingKeys: String, CodingKey {
        case id, name, dataType, isNullable, isPrimaryKey, description
    }
}

// MARK: - Schema Discovery

struct SchemaDiscoveryResult: Codable {
    let tables: [DiscoveredTable]
}

struct DiscoveredTable: Codable, Identifiable {
    let id = UUID()
    let name: String
    let rowCount: Int
    let columns: [ColumnInfo]

    enum CodingKeys: String, CodingKey {
        case name, rowCount, columns
    }
}

// MARK: - Execution

struct AgentExecution: Codable, Identifiable, Equatable {
    let id: String
    let agentId: String
    var status: ExecutionStatus
    var input: String
    var output: String?
    var steps: [ExecutionStep]
    var startedAt: Date
    var completedAt: Date?
    var duration: TimeInterval?
    var tokenUsage: TokenUsage?
    var error: String?

    enum CodingKeys: String, CodingKey {
        case id, agentId, status, input, output, steps
        case startedAt = "started_at"
        case completedAt = "completed_at"
        case duration, tokenUsage, error
    }
}

enum ExecutionStatus: String, Codable {
    case pending
    case running
    case completed
    case failed
    case cancelled
}

struct ExecutionStep: Codable, Identifiable, Equatable {
    let id: String
    let type: StepType
    var content: String
    var toolName: String?
    var toolInput: String?
    var toolOutput: String?
    var status: ExecutionStatus
    var timestamp: Date

    enum CodingKeys: String, CodingKey {
        case id, type, content
        case toolName = "tool_name"
        case toolInput = "tool_input"
        case toolOutput = "tool_output"
        case status, timestamp
    }
}

enum StepType: String, Codable {
    case thought
    case toolCall
    case toolResult
    case answer
    case error
}

struct TokenUsage: Codable, Equatable {
    let inputTokens: Int
    let outputTokens: Int
    let totalTokens: Int

    enum CodingKeys: String, CodingKey {
        case inputTokens = "input_tokens"
        case outputTokens = "output_tokens"
        case totalTokens = "total_tokens"
    }
}

// MARK: - Agent Performance

struct AgentPerformance: Codable, Identifiable {
    let id = UUID()
    let agentId: String
    let date: Date
    let executions: Int
    let successRate: Double
    let avgLatency: Double
    let totalTokens: Int
    let totalCost: Double

    enum CodingKeys: String, CodingKey {
        case agentId, date, executions
        case successRate = "success_rate"
        case avgLatency = "avg_latency"
        case totalTokens = "total_tokens"
        case totalCost = "total_cost"
    }
}

// MARK: - Available Models

struct AIModel: Codable, Identifiable {
    let id: String
    let name: String
    let provider: String
    let maxTokens: Int
    let supportsTools: Bool
    let supportsVision: Bool

    enum CodingKeys: String, CodingKey {
        case id, name, provider
        case maxTokens = "max_tokens"
        case supportsTools = "supports_tools"
        case supportsVision = "supports_vision"
    }
}

// MARK: - Available Tools

struct AITool: Codable, Identifiable {
    let id: String
    let name: String
    let description: String
    let category: String
    let parameters: [ToolParameter]
}

struct ToolParameter: Codable, Identifiable {
    let id: String
    let name: String
    let type: String
    let required: Bool
    let description: String
    let defaultValue: String?

    enum CodingKeys: String, CodingKey {
        case id, name, type, required, description
        case defaultValue = "default_value"
    }
}
```

## Agent List View Model

```swift
import SwiftUI
import Combine

@MainActor
final class AgentListViewModel: ObservableObject {
    @Published var agents: [Agent] = []
    @Published var isLoading = false
    @Published var error: AgentError?
    @Published var searchText = ""
    @Published var sortOrder: SortOrder = .newest
    @Published var filterStatus: AgentStatus?

    private let apiService: AgentAPIService
    private let cacheService: AgentCacheService
    private var cancellables = Set<AnyCancellable>()

    enum SortOrder: String, CaseIterable {
        case newest = "Newest"
        case oldest = "Oldest"
        case nameAZ = "Name A-Z"
        case nameZA = "Name Z-A"
        case mostExecutions = "Most Executions"
    }

    var filteredAgents: [Agent] {
        var result = agents

        if !searchText.isEmpty {
            result = result.filter {
                $0.name.localizedCaseInsensitiveContains(searchText) ||
                ($0.description?.localizedCaseInsensitiveContains(searchText) ?? false)
            }
        }

        if let status = filterStatus {
            result = result.filter { $0.status == status }
        }

        switch sortOrder {
        case .newest: result.sort { $0.createdAt > $1.createdAt }
        case .oldest: result.sort { $0.createdAt < $1.createdAt }
        case .nameAZ: result.sort { $0.name < $1.name }
        case .nameZA: result.sort { $0.name > $1.name }
        case .mostExecutions: break
        }

        return result
    }

    init(
        apiService: AgentAPIService = .shared,
        cacheService: AgentCacheService = .shared
    ) {
        self.apiService = apiService
        self.cacheService = cacheService
    }

    func loadAgents() async {
        isLoading = true
        error = nil

        if let cached = await cacheService.getCachedAgents() {
            agents = cached
            isLoading = false
            await refreshAgents()
            return
        }

        do {
            agents = try await apiService.fetchAgents()
            await cacheService.cacheAgents(agents)
        } catch {
            self.error = AgentError.from(error)
        }

        isLoading = false
    }

    func refreshAgents() async {
        do {
            agents = try await apiService.fetchAgents()
            await cacheService.cacheAgents(agents)
        } catch {
            self.error = AgentError.from(error)
        }
    }

    func deleteAgent(_ agent: Agent) async {
        do {
            try await apiService.deleteAgent(id: agent.id)
            agents.removeAll { $0.id == agent.id }
            await cacheService.cacheAgents(agents)
        } catch {
            self.error = AgentError.from(error)
        }
    }

    func toggleAgentStatus(_ agent: Agent) async {
        var updated = agent
        updated.status = agent.status == .active ? .inactive : .active
        do {
            let saved = try await apiService.updateAgent(updated)
            if let index = agents.firstIndex(where: { $0.id == agent.id }) {
                agents[index] = saved
            }
            await cacheService.cacheAgents(agents)
        } catch {
            self.error = AgentError.from(error)
        }
    }
}

enum AgentError: LocalizedError {
    case network(String)
    case server(Int, String?)
    case decoding
    case unauthorized
    case notFound
    case validation(String)
    case unknown(String)

    var errorDescription: String? {
        switch self {
        case .network(let msg): return "Network error: \(msg)"
        case .server(let code, let msg): return "Server error (\(code)): \(msg ?? "Unknown")"
        case .decoding: return "Failed to parse response"
        case .unauthorized: return "Session expired"
        case .notFound: return "Agent not found"
        case .validation(let msg): return msg
        case .unknown(let msg): return msg
        }
    }

    static func from(_ error: Error) -> AgentError {
        if let apiError = error as? APIError {
            switch apiError {
            case .network(let err): return .network(err.localizedDescription)
            case .http(let code, let msg):
                if code == 404 { return .notFound }
                return .server(code, msg)
            case .decoding: return .decoding
            case .unauthorized: return .unauthorized
            }
        }
        return .unknown(error.localizedDescription)
    }
}
```

## Agent List Screen

```swift
struct AgentListView: View {
    @StateObject private var viewModel = AgentListViewModel()
    @State private var showCreateAgent = false
    @State private var showSortOptions = false
    @State private var agentToDelete: Agent?

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.agents.isEmpty {
                    AgentListLoadingView()
                } else if let error = viewModel.error, viewModel.agents.isEmpty {
                    AgentListErrorView(error: error) {
                        Task { await viewModel.loadAgents() }
                    }
                } else if viewModel.filteredAgents.isEmpty && viewModel.searchText.isEmpty {
                    AgentListEmptyView {
                        showCreateAgent = true
                    }
                } else {
                    agentList
                }
            }
            .navigationTitle("Agents")
            .searchable(text: $viewModel.searchText, prompt: "Search agents...")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Menu {
                        sortMenu
                        statusFilter
                    } label: {
                        Image(systemName: "line.3.horizontal.decrease.circle")
                    }
                }
                ToolbarItem(placement: .topBarTrailing) {
                    Button(action: { showCreateAgent = true }) {
                        Image(systemName: "plus")
                    }
                }
            }
            .sheet(isPresented: $showCreateAgent) {
                AgentCreateSheet()
            }
            .refreshable {
                await viewModel.refreshAgents()
            }
            .task {
                await viewModel.loadAgents()
            }
            .confirmationDialog("Delete Agent", isPresented: .constant(agentToDelete != nil)) {
                Button("Delete", role: .destructive) {
                    if let agent = agentToDelete {
                        Task { await viewModel.deleteAgent(agent) }
                    }
                    agentToDelete = nil
                }
                Button("Cancel", role: .cancel) { agentToDelete = nil }
            } message: {
                Text("Are you sure you want to delete this agent? This action cannot be undone.")
            }
        }
    }

    private var agentList: some View {
        List {
            ForEach(viewModel.filteredAgents) { agent in
                NavigationLink(destination: AgentDetailView(agentId: agent.id)) {
                    AgentCardView(agent: agent)
                }
                .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                    Button(role: .destructive) {
                        agentToDelete = agent
                    } label: {
                        Label("Delete", systemImage: "trash")
                    }
                }
                .swipeActions(edge: .leading) {
                    Button {
                        Task { await viewModel.toggleAgentStatus(agent) }
                    } label: {
                        Label(
                            agent.status == .active ? "Deactivate" : "Activate",
                            systemImage: agent.status == .active ? "pause.fill" : "play.fill"
                        )
                    }
                    .tint(agent.status == .active ? .orange : .green)
                }
            }
        }
        .listStyle(.insetGrouped)
    }

    private var sortMenu: some View {
        Menu("Sort") {
            ForEach(AgentListViewModel.SortOrder.allCases, id: \.self) { order in
                Button {
                    viewModel.sortOrder = order
                } label: {
                    HStack {
                        Text(order.rawValue)
                        if viewModel.sortOrder == order {
                            Image(systemName: "checkmark")
                        }
                    }
                }
            }
        }
    }

    private var statusFilter: some View {
        Menu("Filter Status") {
            Button {
                viewModel.filterStatus = nil
            } label: {
                HStack {
                    Text("All")
                    if viewModel.filterStatus == nil { Image(systemName: "checkmark") }
                }
            }
            ForEach(AgentStatus.allCases, id: \.self) { status in
                Button {
                    viewModel.filterStatus = status
                } label: {
                    HStack {
                        Text(status.displayName)
                        if viewModel.filterStatus == status { Image(systemName: "checkmark") }
                    }
                }
            }
        }
    }
}

struct AgentCardView: View {
    let agent: Agent

    var body: some View {
        HStack(spacing: 12) {
            statusIndicator
            agentInfo
            Spacer()
            executionInfo
        }
        .padding(.vertical, 4)
    }

    private var statusIndicator: some View {
        Circle()
            .fill(statusColor)
            .frame(width: 10, height: 10)
            .overlay(
                Circle()
                    .stroke(statusColor.opacity(0.3), lineWidth: 3)
            )
    }

    private var agentInfo: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(agent.name)
                .font(.headline)
            if let desc = agent.description {
                Text(desc)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }
            HStack(spacing: 8) {
                Label(agent.modelId, systemImage: "cpu")
                if !agent.tools.isEmpty {
                    Label("\(agent.tools.count) tools", systemImage: "wrench")
                }
                if !agent.documentSetIds.isEmpty {
                    Label("\(agent.documentSetIds.count) sets", systemImage: "doc")
                }
            }
            .font(.caption2)
            .foregroundStyle(.secondary)
        }
    }

    private var executionInfo: some View {
        VStack(alignment: .trailing, spacing: 4) {
            Text(agent.status.displayName)
                .font(.caption2.bold())
                .foregroundStyle(statusColor)
        }
    }

    private var statusColor: Color {
        switch agent.status {
        case .active: return .green
        case .inactive: return .gray
        case .error: return .red
        case .training: return .orange
        }
    }
}
```

## Agent Detail Screen

```swift
struct AgentDetailView: View {
    let agentId: String
    @StateObject private var viewModel: AgentDetailViewModel
    @State private var selectedTab: DetailTab = .overview

    init(agentId: String) {
        self.agentId = agentId
        _viewModel = StateObject(wrappedValue: AgentDetailViewModel(agentId: agentId))
    }

    enum DetailTab: String, CaseIterable {
        case overview = "Overview"
        case configuration = "Config"
        case execution = "Execute"
        case history = "History"
        case performance = "Metrics"
    }

    var body: some View {
        Group {
            if viewModel.isLoading {
                ProgressView("Loading agent...")
            } else if let agent = viewModel.agent {
                DetailContent(agent: agent)
            } else if let error = viewModel.error {
                VStack {
                    Text(error.localizedDescription)
                    Button("Retry") {
                        Task { await viewModel.loadAgent() }
                    }
                }
            }
        }
        .navigationTitle(viewModel.agent?.name ?? "Agent")
        .navigationBarTitleDisplayMode(.inline)
        .task { await viewModel.loadAgent() }
    }

    @ViewBuilder
    private func DetailContent(agent: Agent) -> some View {
        VStack(spacing: 0) {
            Picker("Tab", selection: $selectedTab) {
                ForEach(DetailTab.allCases, id: \.self) { tab in
                    Text(tab.rawValue).tag(tab)
                }
            }
            .pickerStyle(.segmented)
            .padding(.horizontal)

            switch selectedTab {
            case .overview:
                AgentOverviewTab(agent: agent)
            case .configuration:
                AgentConfigurationTab(agent: agent) { updated in
                    Task { await viewModel.updateAgent(updated) }
                }
            case .execution:
                AgentExecutionTab(agent: agent)
            case .history:
                AgentHistoryTab(agentId: agent.id)
            case .performance:
                AgentPerformanceTab(agentId: agent.id)
            }
        }
    }
}
```

## Agent Detail ViewModel

```swift
@MainActor
final class AgentDetailViewModel: ObservableObject {
    @Published var agent: Agent?
    @Published var isLoading = false
    @Published var error: AgentError?

    let agentId: String
    private let apiService: AgentAPIService

    init(agentId: String, apiService: AgentAPIService = .shared) {
        self.agentId = agentId
        self.apiService = apiService
    }

    func loadAgent() async {
        isLoading = true
        do {
            agent = try await apiService.fetchAgent(id: agentId)
        } catch {
            self.error = AgentError.from(error)
        }
        isLoading = false
    }

    func updateAgent(_ agent: Agent) async {
        do {
            let updated = try await apiService.updateAgent(agent)
            self.agent = updated
        } catch {
            self.error = AgentError.from(error)
        }
    }
}
```

## Agent Overview Tab

```swift
struct AgentOverviewTab: View {
    let agent: Agent

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                statusCard
                detailsCard
                toolsCard
                documentSetsCard
                sqlConnectionCard
            }
            .padding()
        }
    }

    private var statusCard: some View {
        VStack(spacing: 12) {
            HStack {
                Circle()
                    .fill(statusColor)
                    .frame(width: 12, height: 12)
                Text(agent.status.displayName)
                    .font(.title3.bold())
                Spacer()
            }
            HStack {
                StatView(label: "Model", value: agent.modelId)
                StatView(label: "Temp", value: String(format: "%.1f", agent.temperature))
                StatView(label: "Max Tokens", value: "\(agent.maxTokens)")
            }
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var detailsCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("System Prompt")
                .font(.headline)
            Text(agent.systemPrompt ?? "No system prompt set")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var toolsCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Tools")
                .font(.headline)
            if agent.tools.isEmpty {
                Text("No tools configured")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            } else {
                ForEach(agent.tools) { tool in
                    HStack {
                        Image(systemName: "wrench.fill")
                            .foregroundStyle(.orange)
                        Text(tool.name)
                        Spacer()
                        Circle()
                            .fill(tool.enabled ? .green : .gray)
                            .frame(width: 8, height: 8)
                    }
                    .font(.subheadline)
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var documentSetsCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Document Sets")
                .font(.headline)
            if agent.documentSetIds.isEmpty {
                Text("No document sets bound")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            } else {
                ForEach(agent.documentSetIds, id: \.self) { setId in
                    HStack {
                        Image(systemName: "doc.fill")
                            .foregroundStyle(.blue)
                        Text(setId)
                        Spacer()
                    }
                    .font(.subheadline)
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var sqlConnectionCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Database Connection")
                .font(.headline)
            if let connId = agent.sqlConnectionId {
                HStack {
                    Image(systemName: "cylinder.fill")
                        .foregroundStyle(.purple)
                    Text("Connection: \(connId)")
                    Spacer()
                    Button("Manage") { }
                        .font(.caption)
                }
            } else {
                Text("No database connection")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
            if !agent.boundTables.isEmpty {
                Divider()
                Text("Bound Tables").font(.subheadline.bold())
                ForEach(agent.boundTables) { table in
                    HStack {
                        Image(systemName: "tablecells")
                            .foregroundStyle(.indigo)
                        Text(table.tableName)
                        Spacer()
                        Text(table.permissionLevel.displayName)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                    .font(.subheadline)
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var statusColor: Color {
        switch agent.status {
        case .active: return .green
        case .inactive: return .gray
        case .error: return .red
        case .training: return .orange
        }
    }
}

struct StatView: View {
    let label: String
    let value: String

    var body: some View {
        VStack(spacing: 2) {
            Text(value)
                .font(.subheadline.bold())
            Text(label)
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
    }
}
```

## Agent Creation Screen

```swift
struct AgentCreateSheet: View {
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel = AgentCreateViewModel()

    var body: some View {
        NavigationStack {
            Form {
                Section("Basic Info") {
                    TextField("Agent Name", text: $viewModel.name)
                    TextField("Description (optional)", text: Binding(
                        get: { viewModel.description ?? "" },
                        set: { viewModel.description = $0.isEmpty ? nil : $0 }
                    ))
                }

                Section("Model") {
                    Picker("Model", selection: $viewModel.selectedModelId) {
                        Text("Select Model").tag("")
                        ForEach(viewModel.availableModels) { model in
                            VStack(alignment: .leading) {
                                Text(model.name)
                                Text(model.provider)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                            .tag(model.id)
                        }
                    }
                }

                Section("System Prompt") {
                    TextEditor(text: $viewModel.systemPrompt)
                        .frame(minHeight: 100)
                }

                Section("Parameters") {
                    HStack {
                        Text("Temperature")
                        Spacer()
                        Text(String(format: "%.1f", viewModel.temperature))
                            .foregroundStyle(.secondary)
                    }
                    Slider(value: $viewModel.temperature, in: 0...2, step: 0.1)

                    Stepper(
                        "Max Tokens: \(viewModel.maxTokens)",
                        value: $viewModel.maxTokens,
                        in: 256...128000,
                        step: 256
                    )
                }

                Section("Tools") {
                    ForEach(viewModel.availableTools) { tool in
                        HStack {
                            VStack(alignment: .leading) {
                                Text(tool.name).font(.subheadline)
                                Text(tool.description)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                            Spacer()
                            Toggle("", isOn: Binding(
                                get: { viewModel.enabledToolIds.contains(tool.id) },
                                set: { enabled in
                                    if enabled {
                                        viewModel.enabledToolIds.insert(tool.id)
                                    } else {
                                        viewModel.enabledToolIds.remove(tool.id)
                                    }
                                }
                            ))
                            .labelsHidden()
                        }
                    }
                }

                Section {
                    Button(action: { Task { await viewModel.createAgent() } }) {
                        if viewModel.isCreating {
                            HStack {
                                ProgressView()
                                    .scaleEffect(0.8)
                                Text("Creating...")
                            }
                        } else {
                            Text("Create Agent")
                        }
                    }
                    .disabled(!viewModel.isValid || viewModel.isCreating)
                    .frame(maxWidth: .infinity)
                }
            }
            .navigationTitle("New Agent")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
            .alert("Error", isPresented: $viewModel.showError) {
                Button("OK") { }
            } message: {
                Text(viewModel.errorMessage ?? "Unknown error")
            }
            .onChange(of: viewModel.createdAgent) { agent in
                if agent != nil { dismiss() }
            }
        }
    }
}

@MainActor
final class AgentCreateViewModel: ObservableObject {
    @Published var name = ""
    @Published var description: String?
    @Published var selectedModelId = ""
    @Published var systemPrompt = ""
    @Published var temperature: Double = 0.7
    @Published var maxTokens: Int = 4096
    @Published var enabledToolIds: Set<String> = []
    @Published var isCreating = false
    @Published var showError = false
    @Published var errorMessage: String?
    @Published var createdAgent: Agent?

    var availableModels: [AIModel] = []
    var availableTools: [AITool] = []

    var isValid: Bool {
        !name.trimmingCharacters(in: .whitespaces).isEmpty &&
        !selectedModelId.isEmpty
    }

    private let apiService: AgentAPIService

    init(apiService: AgentAPIService = .shared) {
        self.apiService = apiService
        Task { await loadOptions() }
    }

    func loadOptions() async {
        do {
            async let models = apiService.fetchAvailableModels()
            async let tools = apiService.fetchAvailableTools()
            availableModels = try await models
            availableTools = try await tools
            if selectedModelId.isEmpty, let first = availableModels.first {
                selectedModelId = first.id
            }
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }

    func createAgent() async {
        isCreating = true
        let tools = enabledToolIds.map { toolId in
            AgentToolConfig(
                id: UUID().uuidString,
                toolId: toolId,
                name: availableTools.first { $0.id == toolId }?.name ?? "",
                enabled: true,
                parameters: [:]
            )
        }

        let agent = CreateAgentRequest(
            name: name,
            description: description,
            modelId: selectedModelId,
            systemPrompt: systemPrompt,
            temperature: temperature,
            maxTokens: maxTokens,
            tools: tools
        )

        do {
            createdAgent = try await apiService.createAgent(agent)
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
        isCreating = false
    }
}
```

## Agent Configuration Tab

```swift
struct AgentConfigurationTab: View {
    let agent: Agent
    let onSave: (Agent) -> Void

    @State private var editedAgent: Agent
    @State private var showDocumentSetSelector = false
    @State private var showSQLConnectionSheet = false

    init(agent: Agent, onSave: @escaping (Agent) -> Void) {
        self.agent = agent
        self.onSave = onSave
        _editedAgent = State(initialValue: agent)
    }

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                modelSection
                promptSection
                parametersSection
                toolsSection
                documentSetsSection
                sqlConnectionSection
                saveButton
            }
            .padding()
        }
    }

    private var modelSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Model").font(.headline)
            Picker("Model", selection: $editedAgent.modelId) {
                // Populated from available models
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var promptSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("System Prompt").font(.headline)
            TextEditor(text: Binding(
                get: { editedAgent.systemPrompt ?? "" },
                set: { editedAgent.systemPrompt = $0.isEmpty ? nil : $0 }
            ))
            .frame(minHeight: 120)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var parametersSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Parameters").font(.headline)
            HStack {
                Text("Temperature: \(String(format: "%.1f", editedAgent.temperature))")
                Spacer()
            }
            Slider(value: $editedAgent.temperature, in: 0...2, step: 0.1)

            Stepper(
                "Max Tokens: \(editedAgent.maxTokens)",
                value: $editedAgent.maxTokens,
                in: 256...128000,
                step: 256
            )
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var toolsSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Tools").font(.headline)
            ForEach(editedAgent.tools) { tool in
                HStack {
                    VStack(alignment: .leading) {
                        Text(tool.name)
                        if let params = tool.parameters as? [String: AgentToolConfig.ToolParameterValue],
                           !params.isEmpty {
                            Text("\(params.count) parameters")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                    Spacer()
                    Toggle("", isOn: Binding(
                        get: { tool.enabled },
                        set: { newValue in
                            if let idx = editedAgent.tools.firstIndex(where: { $0.id == tool.id }) {
                                editedAgent.tools[idx].enabled = newValue
                            }
                        }
                    ))
                    .labelsHidden()
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var documentSetsSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text("Document Sets").font(.headline)
                Spacer()
                Button("Add") { showDocumentSetSelector = true }
            }
            if editedAgent.documentSetIds.isEmpty {
                Text("No document sets bound")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            } else {
                ForEach(editedAgent.documentSetIds, id: \.self) { setId in
                    HStack {
                        Image(systemName: "doc.fill")
                        Text(setId)
                        Spacer()
                        Button(action: {
                            editedAgent.documentSetIds.removeAll { $0 == setId }
                        }) {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundStyle(.red)
                        }
                    }
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .sheet(isPresented: $showDocumentSetSelector) {
            DocumentSetSelectorSheet(selectedIds: $editedAgent.documentSetIds)
        }
    }

    private var sqlConnectionSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text("Database").font(.headline)
                Spacer()
                Button(editedAgent.sqlConnectionId == nil ? "Connect" : "Manage") {
                    showSQLConnectionSheet = true
                }
            }
            if let connId = editedAgent.sqlConnectionId {
                Text("Connected: \(connId)")
                    .font(.subheadline)
            } else {
                Text("No database connected")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .sheet(isPresented: $showSQLConnectionSheet) {
            SQLConnectionSheet(
                connectionId: editedAgent.sqlConnectionId,
                boundTables: $editedAgent.boundTables,
                onConnect: { connId in
                    editedAgent.sqlConnectionId = connId
                }
            )
        }
    }

    private var saveButton: some View {
        Button("Save Changes") { onSave(editedAgent) }
            .buttonStyle(.borderedProminent)
            .disabled(editedAgent == agent)
    }
}
```

## Agent Execution Tab

```swift
struct AgentExecutionTab: View {
    let agent: Agent
    @StateObject private var viewModel = AgentExecutionViewModel(agentId: agent.id)
    @State private var inputText = ""

    var body: some View {
        VStack(spacing: 0) {
            executionSteps
            Divider()
            inputBar
        }
    }

    private var executionSteps: some View {
        ScrollViewReader { proxy in
            ScrollView {
                LazyVStack(spacing: 12) {
                    if viewModel.currentExecution == nil && viewModel.steps.isEmpty {
                        VStack(spacing: 12) {
                            Image(systemName: "brain.head.profile")
                                .font(.system(size: 40))
                                .foregroundStyle(.secondary)
                            Text("Enter a prompt to start execution")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                        }
                        .frame(maxWidth: .infinity, minHeight: 300)
                    } else {
                        ForEach(viewModel.steps) { step in
                            ExecutionStepView(step: step)
                                .id(step.id)
                        }

                        if viewModel.isExecuting {
                            HStack(spacing: 8) {
                                ProgressView()
                                    .scaleEffect(0.7)
                                Text("Processing...")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                            .frame(maxWidth: .infinity)
                            .id("loading")
                        }
                    }
                }
                .padding()
            }
            .onChange(of: viewModel.steps.count) { _ in
                withAnimation {
                    if let lastStep = viewModel.steps.last {
                        proxy.scrollTo(lastStep.id, anchor: .bottom)
                    } else {
                        proxy.scrollTo("loading", anchor: .bottom)
                    }
                }
            }
        }
    }

    private var inputBar: some View {
        HStack(spacing: 12) {
            TextField("Enter prompt...", text: $inputText, axis: .vertical)
                .textFieldStyle(.plain)
                .lineLimit(1...5)

            Button(action: {
                let text = inputText
                inputText = ""
                Task { await viewModel.execute(input: text) }
            }) {
                Image(systemName: "arrow.up.circle.fill")
                    .font(.title2)
            }
            .disabled(inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || viewModel.isExecuting)
        }
        .padding()
        .background(.regularMaterial)
    }
}

struct ExecutionStepView: View {
    let step: ExecutionStep

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: iconForStep(step.type))
                    .foregroundStyle(colorForStep(step.type))
                Text(titleForStep(step.type))
                    .font(.caption.bold())
                Spacer()
                Text(step.timestamp.formatted(date: .omitted, time: .shortened))
                    .font(.caption2)
                    .foregroundStyle(.tertiary)
            }

            switch step.type {
            case .thought:
                Text(step.content)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .padding(8)
                    .background(.quaternary)
                    .clipShape(RoundedRectangle(cornerRadius: 8))
            case .toolCall:
                VStack(alignment: .leading, spacing: 4) {
                    if let toolName = step.toolName {
                        Label(toolName, systemImage: "wrench.fill")
                            .font(.subheadline.bold())
                    }
                    if let input = step.toolInput {
                        Text(input)
                            .font(.caption)
                            .padding(6)
                            .background(Color.orange.opacity(0.1))
                            .clipShape(RoundedRectangle(cornerRadius: 6))
                    }
                }
            case .toolResult:
                if let output = step.toolOutput {
                    Text(output)
                        .font(.caption)
                        .padding(6)
                        .background(Color.green.opacity(0.1))
                        .clipShape(RoundedRectangle(cornerRadius: 6))
                }
            case .answer:
                Text(step.content)
                    .font(.body)
            case .error:
                Text(step.content)
                    .font(.subheadline)
                    .foregroundStyle(.red)
            }
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func iconForStep(_ type: StepType) -> String {
        switch type {
        case .thought: return "brain"
        case .toolCall: return "wrench.and.screwdriver"
        case .toolResult: return "checkmark.circle"
        case .answer: return "bubble.left.fill"
        case .error: return "exclamationmark.triangle"
        }
    }

    private func colorForStep(_ type: StepType) -> Color {
        switch type {
        case .thought: return .purple
        case .toolCall: return .orange
        case .toolResult: return .green
        case .answer: return .blue
        case .error: return .red
        }
    }

    private func titleForStep(_ type: StepType) -> String {
        switch type {
        case .thought: return "Thinking"
        case .toolCall: return "Tool Call"
        case .toolResult: return "Tool Result"
        case .answer: return "Answer"
        case .error: return "Error"
        }
    }
}
```

## Agent Execution ViewModel

```swift
@MainActor
final class AgentExecutionViewModel: ObservableObject {
    @Published var currentExecution: AgentExecution?
    @Published var steps: [ExecutionStep] = []
    @Published var isExecuting = false
    @Published var error: AgentError?

    let agentId: String
    private let apiService: AgentAPIService
    private let webSocketService: ExecutionWebSocketService

    init(
        agentId: String,
        apiService: AgentAPIService = .shared,
        webSocketService: ExecutionWebSocketService = .shared
    ) {
        self.agentId = agentId
        self.apiService = apiService
        self.webSocketService = webSocketService
    }

    func execute(input: String) async {
        guard !isExecuting else { return }
        isExecuting = true
        error = nil

        do {
            let execution = try await apiService.startExecution(
                agentId: agentId,
                input: input
            )
            currentExecution = execution
            steps = []

            webSocketService.subscribe(executionId: execution.id) { [weak self] step in
                Task { @MainActor in
                    self?.steps.append(step)
                }
            }

            // Poll for completion
            while isExecuting {
                try await Task.sleep(for: .seconds(1))
                if let exec = try? await apiService.getExecution(id: execution.id) {
                    currentExecution = exec
                    if exec.status == .completed || exec.status == .failed || exec.status == .cancelled {
                        isExecuting = false
                        webSocketService.unsubscribe(executionId: execution.id)
                        break
                    }
                }
            }
        } catch {
            self.error = AgentError.from(error)
            isExecuting = false
        }
    }

    func cancelExecution() async {
        guard let executionId = currentExecution?.id else { return }
        do {
            try await apiService.cancelExecution(id: executionId)
            isExecuting = false
            webSocketService.unsubscribe(executionId: executionId)
        } catch {
            self.error = AgentError.from(error)
        }
    }
}
```

## Agent Execution History Tab

```swift
struct AgentHistoryTab: View {
    let agentId: String
    @StateObject private var viewModel: AgentHistoryViewModel

    init(agentId: String) {
        self.agentId = agentId
        _viewModel = StateObject(wrappedValue: AgentHistoryViewModel(agentId: agentId))
    }

    var body: some View {
        Group {
            if viewModel.isLoading {
                ProgressView("Loading history...")
            } else if viewModel.executions.isEmpty {
                VStack(spacing: 12) {
                    Image(systemName: "clock.arrow.circlepath")
                        .font(.system(size: 40))
                        .foregroundStyle(.secondary)
                    Text("No execution history")
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                List(viewModel.executions) { execution in
                    NavigationLink(destination: ExecutionDetailView(execution: execution)) {
                        ExecutionHistoryRow(execution: execution)
                    }
                }
                .listStyle(.insetGrouped)
            }
        }
        .task { await viewModel.loadHistory() }
        .refreshable { await viewModel.loadHistory() }
    }
}

struct ExecutionHistoryRow: View {
    let execution: AgentExecution

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                statusIcon
                Text(execution.input)
                    .font(.subheadline.bold())
                    .lineLimit(1)
                Spacer()
                Text(execution.startedAt.formatted(date: .abbreviated, time: .shortened))
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            }
            HStack(spacing: 12) {
                Label(execution.status.rawValue.capitalized, systemImage: statusIconName)
                    .font(.caption)
                if let duration = execution.duration {
                    Label("\(String(format: "%.1f", duration))s", systemImage: "clock")
                        .font(.caption)
                }
                if let tokens = execution.tokenUsage {
                    Label("\(tokens.totalTokens) tokens", systemImage: "cpu")
                        .font(.caption)
                }
            }
            .foregroundStyle(.secondary)
        }
    }

    private var statusIcon: some View {
        Group {
            switch execution.status {
            case .completed: Image(systemName: "checkmark.circle.fill").foregroundStyle(.green)
            case .failed: Image(systemName: "xmark.circle.fill").foregroundStyle(.red)
            case .running: ProgressView().scaleEffect(0.6)
            case .pending: Image(systemName: "clock.fill").foregroundStyle(.orange)
            case .cancelled: Image(systemName: "slash.circle.fill").foregroundStyle(.gray)
            }
        }
        .font(.caption)
    }

    private var statusIconName: String {
        switch execution.status {
        case .completed: return "checkmark.circle"
        case .failed: return "xmark.circle"
        case .running: return "arrow.triangle.2.circlepath"
        case .pending: return "clock"
        case .cancelled: return "slash.circle"
        }
    }
}

@MainActor
final class AgentHistoryViewModel: ObservableObject {
    @Published var executions: [AgentExecution] = []
    @Published var isLoading = false
    @Published var error: AgentError?

    let agentId: String
    private let apiService: AgentAPIService

    init(agentId: String, apiService: AgentAPIService = .shared) {
        self.agentId = agentId
        self.apiService = apiService
    }

    func loadHistory() async {
        isLoading = true
        do {
            executions = try await apiService.fetchExecutionHistory(agentId: agentId)
        } catch {
            self.error = AgentError.from(error)
        }
        isLoading = false
    }
}
```

## Agent Performance Tab

```swift
struct AgentPerformanceTab: View {
    let agentId: String
    @StateObject private var viewModel: AgentPerformanceViewModel

    init(agentId: String) {
        self.agentId = agentId
        _viewModel = StateObject(wrappedValue: AgentPerformanceViewModel(agentId: agentId))
    }

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                summaryCards
                executionsChart
                latencyChart
                tokenUsageChart
            }
            .padding()
        }
        .task { await viewModel.loadPerformance() }
    }

    private var summaryCards: some View {
        LazyVGrid(columns: [
            GridItem(.flexible()),
            GridItem(.flexible())
        ], spacing: 12) {
            StatCard(
                title: "Total Executions",
                value: "\(viewModel.totalExecutions)",
                icon: "play.circle.fill",
                color: .blue
            )
            StatCard(
                title: "Success Rate",
                value: String(format: "%.1f%%", viewModel.successRate),
                icon: "checkmark.seal.fill",
                color: .green
            )
            StatCard(
                title: "Avg Latency",
                value: String(format: "%.0fms", viewModel.avgLatency),
                icon: "clock.fill",
                color: .orange
            )
            StatCard(
                title: "Total Cost",
                value: String(format: "$%.2f", viewModel.totalCost),
                icon: "dollarsign.circle.fill",
                color: .purple
            )
        }
    }

    private var executionsChart: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Executions Over Time")
                .font(.headline)
            Chart(viewModel.performanceData) { data in
                BarMark(
                    x: .value("Date", data.date, unit: .day),
                    y: .value("Executions", data.executions)
                )
                .foregroundStyle(.blue)
            }
            .frame(height: 150)
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var latencyChart: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Average Latency")
                .font(.headline)
            Chart(viewModel.performanceData) { data in
                LineMark(
                    x: .value("Date", data.date, unit: .day),
                    y: .value("Latency (ms)", data.avgLatency)
                )
                .foregroundStyle(.orange)
            }
            .frame(height: 150)
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private var tokenUsageChart: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Token Usage")
                .font(.headline)
            Chart(viewModel.performanceData) { data in
                BarMark(
                    x: .value("Date", data.date, unit: .day),
                    y: .value("Tokens", data.totalTokens)
                )
                .foregroundStyle(.purple)
            }
            .frame(height: 150)
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct StatCard: View {
    let title: String
    let value: String
    let icon: String
    let color: Color

    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: icon)
                .font(.title2)
                .foregroundStyle(color)
            Text(value)
                .font(.title3.bold())
            Text(title)
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

@MainActor
final class AgentPerformanceViewModel: ObservableObject {
    @Published var performanceData: [AgentPerformance] = []
    @Published var isLoading = false

    var totalExecutions: Int { performanceData.reduce(0) { $0 + $1.executions } }
    var successRate: Double {
        guard !performanceData.isEmpty else { return 0 }
        return performanceData.map(\.successRate).reduce(0, +) / Double(performanceData.count)
    }
    var avgLatency: Double {
        guard !performanceData.isEmpty else { return 0 }
        return performanceData.map(\.avgLatency).reduce(0, +) / Double(performanceData.count)
    }
    var totalCost: Double { performanceData.reduce(0) { $0 + $1.totalCost } }

    let agentId: String
    private let apiService: AgentAPIService

    init(agentId: String, apiService: AgentAPIService = .shared) {
        self.agentId = agentId
        self.apiService = apiService
    }

    func loadPerformance() async {
        isLoading = true
        do {
            performanceData = try await apiService.fetchAgentPerformance(agentId: agentId)
        } catch { }
        isLoading = false
    }
}
```

## SQL Connection & Schema Discovery

```swift
struct SQLConnectionSheet: View {
    @Environment(\.dismiss) private var dismiss
    let connectionId: String?
    @Binding var boundTables: [BoundTable]
    let onConnect: (String) -> Void

    @StateObject private var viewModel = SQLConnectionViewModel()

    var body: some View {
        NavigationStack {
            Group {
                if let connection = viewModel.connection {
                    connectedView(connection)
                } else {
                    connectionForm
                }
            }
            .navigationTitle("Database Connection")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
    }

    private var connectionForm: some View {
        Form {
            Section("Connection Details") {
                Picker("Type", selection: $viewModel.dbType) {
                    ForEach(DatabaseType.allCases, id: \.self) { type in
                        Text(type.rawValue).tag(type)
                    }
                }
                TextField("Host", text: $viewModel.host)
                TextField("Port", text: $viewModel.port)
                    .keyboardType(.numberPad)
                TextField("Database", text: $viewModel.database)
                TextField("Username", text: $viewModel.username)
                SecureField("Password", text: $viewModel.password)
            }

            Section {
                Button(action: { Task { await viewModel.testConnection() } }) {
                    HStack {
                        if viewModel.isTesting {
                            ProgressView().scaleEffect(0.8)
                        } else {
                            Image(systemName: "bolt.fill")
                        }
                        Text("Test Connection")
                    }
                }
                .disabled(viewModel.isTesting || !viewModel.isValid)

                if let testResult = viewModel.testResult {
                    HStack {
                        Image(systemName: testResult.success ? "checkmark.circle.fill" : "xmark.circle.fill")
                            .foregroundStyle(testResult.success ? .green : .red)
                        Text(testResult.message)
                            .font(.caption)
                    }
                }
            }
        }
    }

    private func connectedView(_ connection: SQLConnection) -> some View {
        List {
            Section("Connection") {
                LabeledContent("Type", value: connection.type.rawValue)
                LabeledContent("Host", value: connection.host)
                LabeledContent("Port", value: "\(connection.port)")
                LabeledContent("Database", value: connection.database)
            }

            Section {
                Button("Discover Schema") {
                    Task { await viewModel.discoverSchema() }
                }
                Button("Disconnect", role: .destructive) {
                    viewModel.disconnect()
                }
            }

            if let schema = viewModel.schema {
                Section("Tables (\(schema.tables.count))") {
                    ForEach(schema.tables) { table in
                        NavigationLink(destination: TableBindingView(
                            table: table,
                            boundTables: $boundTables
                        )) {
                            HStack {
                                Image(systemName: "tablecells")
                                VStack(alignment: .leading) {
                                    Text(table.name)
                                    Text("\(table.rowCount) rows, \(table.columns.count) columns")
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

struct TableBindingView: View {
    let table: DiscoveredTable
    @Binding var boundTables: [BoundTable]
    @Environment(\.dismiss) private var dismiss

    @State private var permission: TablePermission = .readOnly
    @State private var selectedColumns: Set<String> = []

    var body: some View {
        List {
            Section("Table Info") {
                LabeledContent("Name", value: table.name)
                LabeledContent("Rows", value: "\(table.rowCount)")
            }

            Section("Permission") {
                Picker("Access Level", selection: $permission) {
                    ForEach(TablePermission.allCases, id: \.self) { perm in
                        Text(perm.displayName).tag(perm)
                    }
                }
                .pickerStyle(.segmented)
            }

            Section("Columns") {
                ForEach(table.columns) { column in
                    HStack {
                        VStack(alignment: .leading) {
                            Text(column.name)
                                .font(.subheadline)
                            Text(column.dataType)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                        Spacer()
                        if column.isPrimaryKey {
                            Text("PK").font(.caption2.bold()).foregroundStyle(.orange)
                        }
                        Toggle("", isOn: Binding(
                            get: { selectedColumns.contains(column.name) },
                            set: { isOn in
                                if isOn { selectedColumns.insert(column.name) }
                                else { selectedColumns.remove(column.name) }
                            }
                        ))
                        .labelsHidden()
                    }
                }
            }

            Section {
                Button("Bind Table") {
                    let binding = BoundTable(
                        id: UUID().uuidString,
                        tableName: table.name,
                        permissionLevel: permission,
                        selectedColumns: Array(selectedColumns),
                        rowEstimate: table.rowCount,
                        columnDetails: table.columns
                    )
                    boundTables.removeAll { $0.tableName == table.name }
                    boundTables.append(binding)
                    dismiss()
                }
                .disabled(selectedColumns.isEmpty)
            }
        }
        .navigationTitle(table.name)
        .navigationBarTitleDisplayMode(.inline)
        .onAppear {
            if let existing = boundTables.first(where: { $0.tableName == table.name }) {
                permission = existing.permissionLevel
                selectedColumns = Set(existing.selectedColumns ?? [])
            } else {
                selectedColumns = Set(table.columns.map(\.name))
            }
        }
    }
}

@MainActor
final class SQLConnectionViewModel: ObservableObject {
    @Published var dbType: DatabaseType = .postgresql
    @Published var host = ""
    @Published var port = ""
    @Published var database = ""
    @Published var username = ""
    @Published var password = ""
    @Published var isTesting = false
    @Published var testResult: TestResult?
    @Published var connection: SQLConnection?
    @Published var schema: SchemaDiscoveryResult?
    @Published var isLoading = false

    var isValid: Bool {
        !host.isEmpty && !database.isEmpty && !username.isEmpty
    }

    struct TestResult {
        let success: Bool
        let message: String
    }

    private let apiService: AgentAPIService

    init(apiService: AgentAPIService = .shared) {
        self.apiService = apiService
    }

    func testConnection() async {
        isTesting = true
        do {
            let result = try await apiService.testSQLConnection(
                type: dbType, host: host,
                port: Int(port) ?? dbType.defaultPort,
                database: database,
                username: username, password: password
            )
            testResult = TestResult(success: true, message: "Connected successfully")
            connection = result
        } catch {
            testResult = TestResult(success: false, message: error.localizedDescription)
        }
        isTesting = false
    }

    func discoverSchema() async {
        guard let connId = connection?.id else { return }
        isLoading = true
        do {
            schema = try await apiService.discoverSchema(connectionId: connId)
        } catch { }
        isLoading = false
    }

    func disconnect() {
        connection = nil
        schema = nil
    }
}
```

## Document Set Selector

```swift
struct DocumentSetSelectorSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Binding var selectedIds: [String]
    @State private var availableSets: [DocumentSet] = []
    @State private var isLoading = false

    private let apiService: DocumentAPIService

    init(selectedIds: Binding<[String]>, apiService: DocumentAPIService = .shared) {
        _selectedIds = selectedIds
        self.apiService = apiService
    }

    var body: some View {
        NavigationStack {
            Group {
                if isLoading {
                    ProgressView("Loading document sets...")
                } else {
                    List(availableSets) { set in
                        HStack {
                            VStack(alignment: .leading) {
                                Text(set.name)
                                Text("\(set.documentCount) documents")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                            Spacer()
                            if selectedIds.contains(set.id) {
                                Image(systemName: "checkmark.circle.fill")
                                    .foregroundStyle(.blue)
                            }
                        }
                        .contentShape(Rectangle())
                        .onTapGesture {
                            if selectedIds.contains(set.id) {
                                selectedIds.removeAll { $0 == set.id }
                            } else {
                                selectedIds.append(set.id)
                            }
                        }
                    }
                }
            }
            .navigationTitle("Document Sets")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { dismiss() }
                }
            }
            .task {
                isLoading = true
                do {
                    availableSets = try await apiService.fetchDocumentSets()
                } catch { }
                isLoading = false
            }
        }
    }
}
```

## Agent API Integration

```swift
protocol AgentAPIService {
    func fetchAgents() async throws -> [Agent]
    func fetchAgent(id: String) async throws -> Agent
    func createAgent(_ request: CreateAgentRequest) async throws -> Agent
    func updateAgent(_ agent: Agent) async throws -> Agent
    func deleteAgent(id: String) async throws
    func startExecution(agentId: String, input: String) async throws -> AgentExecution
    func getExecution(id: String) async throws -> AgentExecution
    func cancelExecution(id: String) async throws
    func fetchExecutionHistory(agentId: String) async throws -> [AgentExecution]
    func fetchAgentPerformance(agentId: String) async throws -> [AgentPerformance]
    func fetchAvailableModels() async throws -> [AIModel]
    func fetchAvailableTools() async throws -> [AITool]
    func testSQLConnection(type: DatabaseType, host: String, port: Int, database: String, username: String, password: String) async throws -> SQLConnection
    func discoverSchema(connectionId: String) async throws -> SchemaDiscoveryResult
}

struct CreateAgentRequest: Codable {
    let name: String
    let description: String?
    let modelId: String
    let systemPrompt: String?
    let temperature: Double
    let maxTokens: Int
    let tools: [AgentToolConfig]

    enum CodingKeys: String, CodingKey {
        case name, description
        case modelId = "model_id"
        case systemPrompt = "system_prompt"
        case temperature
        case maxTokens = "max_tokens"
        case tools
    }
}

final class DefaultAgentAPIService: AgentAPIService {
    static let shared = DefaultAgentAPIService()
    private let client: NetworkClient

    init(client: NetworkClient = .shared) {
        self.client = client
    }

    func fetchAgents() async throws -> [Agent] {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/agents")!,
            method: .post
        )
        let (data, _) = try await client.execute(request)
        return try JSONDecoder().decode([Agent].self, from: data)
    }

    func fetchAgent(id: String) async throws -> Agent {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/agents/\(id)")!,
            method: .post
        )
        let (data, _) = try await client.execute(request)
        return try JSONDecoder().decode(Agent.self, from: data)
    }

    func createAgent(_ request: CreateAgentRequest) async throws -> Agent {
        var urlRequest = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/agents")!,
            method: .post
        )
        urlRequest.httpBody = try JSONEncoder().encode(request)
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let (data, _) = try await client.execute(urlRequest)
        return try JSONDecoder().decode(Agent.self, from: data)
    }

    func updateAgent(_ agent: Agent) async throws -> Agent {
        var urlRequest = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/agents/\(agent.id)")!,
            method: .patch
        )
        urlRequest.httpBody = try JSONEncoder().encode(agent)
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let (data, _) = try await client.execute(urlRequest)
        return try JSONDecoder().decode(Agent.self, from: data)
    }

    func deleteAgent(id: String) async throws {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/agents/\(id)")!,
            method: .delete
        )
        _ = try await client.execute(request)
    }

    func startExecution(agentId: String, input: String) async throws -> AgentExecution {
        var urlRequest = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/agents/\(agentId)/execute")!,
            method: .post
        )
        let body = ["input": input]
        urlRequest.httpBody = try JSONSerialization.data(withJSONObject: body)
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let (data, _) = try await client.execute(urlRequest)
        return try JSONDecoder().decode(AgentExecution.self, from: data)
    }

    func getExecution(id: String) async throws -> AgentExecution {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/executions/\(id)")!,
            method: .post
        )
        let (data, _) = try await client.execute(request)
        return try JSONDecoder().decode(AgentExecution.self, from: data)
    }

    func cancelExecution(id: String) async throws {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/executions/\(id)/cancel")!,
            method: .post
        )
        _ = try await client.execute(request)
    }

    func fetchExecutionHistory(agentId: String) async throws -> [AgentExecution] {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/agents/\(agentId)/executions")!,
            method: .post
        )
        let (data, _) = try await client.execute(request)
        return try JSONDecoder().decode([AgentExecution].self, from: data)
    }

    func fetchAgentPerformance(agentId: String) async throws -> [AgentPerformance] {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/agents/\(agentId)/performance")!,
            method: .post
        )
        let (data, _) = try await client.execute(request)
        return try JSONDecoder().decode([AgentPerformance].self, from: data)
    }

    func fetchAvailableModels() async throws -> [AIModel] {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/models")!,
            method: .post
        )
        let (data, _) = try await client.execute(request)
        return try JSONDecoder().decode([AIModel].self, from: data)
    }

    func fetchAvailableTools() async throws -> [AITool] {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/tools")!,
            method: .post
        )
        let (data, _) = try await client.execute(request)
        return try JSONDecoder().decode([AITool].self, from: data)
    }

    func testSQLConnection(type: DatabaseType, host: String, port: Int, database: String, username: String, password: String) async throws -> SQLConnection {
        var urlRequest = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/sql-connections/test")!,
            method: .post
        )
        let body: [String: Any] = [
            "type": type.rawValue, "host": host, "port": port,
            "database": database, "username": username, "password": password
        ]
        urlRequest.httpBody = try JSONSerialization.data(withJSONObject: body)
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let (data, _) = try await client.execute(urlRequest)
        return try JSONDecoder().decode(SQLConnection.self, from: data)
    }

    func discoverSchema(connectionId: String) async throws -> SchemaDiscoveryResult {
        let request = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/sql-connections/\(connectionId)/schema")!,
            method: .post
        )
        let (data, _) = try await client.execute(request)
        return try JSONDecoder().decode(SchemaDiscoveryResult.self, from: data)
    }
}
```

## Agent Caching

```swift
final class AgentCacheService {
    static let shared = AgentCacheService()
    private let cache = NSCache<NSString, CachedAgents>()
    private let cacheExpiry: TimeInterval = 300

    func getCachedAgents() async -> [Agent]? {
        guard let cached = cache.object(forKey: "agents" as NSString) else { return nil }
        guard Date().timeIntervalSince(cached.timestamp) < cacheExpiry else {
            cache.removeObject(forKey: "agents" as NSString)
            return nil
        }
        return cached.agents
    }

    func cacheAgents(_ agents: [Agent]) async {
        cache.setObject(CachedAgents(agents: agents, timestamp: Date()), forKey: "agents" as NSString)
    }

    func invalidateCache() {
        cache.removeAllObjects()
    }
}

final class CachedAgents: NSObject {
    let agents: [Agent]
    let timestamp: Date
    init(agents: [Agent], timestamp: Date) {
        self.agents = agents
        self.timestamp = timestamp
    }
}
```

## Loading, Error, and Empty States

```swift
struct AgentListLoadingView: View {
    var body: some View {
        List(0..<5) { _ in
            HStack(spacing: 12) {
                Circle().fill(.quaternary).frame(width: 10, height: 10)
                VStack(alignment: .leading, spacing: 6) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(.quaternary)
                        .frame(width: 120, height: 16)
                    RoundedRectangle(cornerRadius: 4)
                        .fill(.quaternary)
                        .frame(width: 200, height: 12)
                }
                Spacer()
            }
            .padding(.vertical, 8)
        }
        .listStyle(.insetGrouped)
        .redacted(reason: .placeholder)
    }
}

struct AgentListErrorView: View {
    let error: AgentError
    let onRetry: () -> Void

    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle")
                .font(.largeTitle)
                .foregroundStyle(.orange)
            Text(error.localizedDescription)
                .multilineTextAlignment(.center)
            Button("Retry", action: onRetry)
                .buttonStyle(.borderedProminent)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct AgentListEmptyView: View {
    let onCreate: () -> Void

    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "person.badge.plus")
                .font(.system(size: 48))
                .foregroundStyle(.secondary)
            Text("No Agents Yet")
                .font(.title2.bold())
            Text("Create your first AI agent to get started.")
                .foregroundStyle(.secondary)
            Button("Create Agent", action: onCreate)
                .buttonStyle(.borderedProminent)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
```

## Responsive Design

| Screen | iPhone | iPad |
|---|---|---|
| Agent List | Single column, full width | Grid layout, 2-3 columns |
| Agent Detail | TabView (scrollable) | Sidebar tabs |
| Agent Create | Full screen sheet | Split view sheet |
| Execution | Vertical scroll | Side-by-side (config + output) |
| History | List | Grid + detail pane |
| SQL Connection | Form stack | Form + schema preview |

```swift
struct AgentResponsiveModifier: ViewModifier {
    @Environment(\.horizontalSizeClass) var sizeClass

    func body(content: Content) -> some View {
        content
            .navigationViewStyle(
                sizeClass == .compact ? .stack : .columns
            )
    }
}
```

## Accessibility

- All agent status indicators include VoiceOver labels
- Tool toggles use `accessibilityLabel` for state
- Execution steps use `accessibilityElement(children: .combine)`
- Dynamic Type supported on all text elements
- Charts include data table fallback for VoiceOver
- Color is never the only indicator of status
- Haptic feedback on execution start/complete

## File Structure

```
AgentManagement/
├── Views/
│   ├── AgentListView.swift
│   ├── AgentCardView.swift
│   ├── AgentDetailView.swift
│   ├── AgentCreateSheet.swift
│   ├── AgentOverviewTab.swift
│   ├── AgentConfigurationTab.swift
│   ├── AgentExecutionTab.swift
│   ├── AgentHistoryTab.swift
│   ├── AgentPerformanceTab.swift
│   ├── ExecutionStepView.swift
│   ├── SQLConnectionSheet.swift
│   ├── TableBindingView.swift
│   ├── DocumentSetSelectorSheet.swift
│   ├── AgentListLoadingView.swift
│   ├── AgentListErrorView.swift
│   └── AgentListEmptyView.swift
├── ViewModels/
│   ├── AgentListViewModel.swift
│   ├── AgentDetailViewModel.swift
│   ├── AgentCreateViewModel.swift
│   ├── AgentExecutionViewModel.swift
│   ├── AgentHistoryViewModel.swift
│   ├── AgentPerformanceViewModel.swift
│   └── SQLConnectionViewModel.swift
├── Models/
│   ├── Agent.swift
│   ├── AgentTool.swift
│   ├── SQLConnection.swift
│   ├── BoundTable.swift
│   ├── AgentExecution.swift
│   └── AgentPerformance.swift
├── Services/
│   ├── AgentAPIService.swift
│   ├── AgentCacheService.swift
│   └── ExecutionWebSocketService.swift
└── Extensions/
    └── AgentAccessibility.swift
```
