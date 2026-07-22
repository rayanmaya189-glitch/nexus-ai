# iOS Architecture Overview

## Table of Contents

- [Tech Stack](#tech-stack)
- [Architecture Layers](#architecture-layers)
- [Module Structure](#module-structure)
- [Dependency Flow](#dependency-flow)
- [Navigation Architecture](#navigation-architecture)
- [State Management Architecture](#state-management-architecture)
- [Networking Architecture](#networking-architecture)
- [Local Storage Architecture](#local-storage-architecture)
- [Background Work Architecture](#background-work-architecture)
- [Push Notification Architecture](#push-notification-architecture)
- [Biometric Architecture](#biometric-architecture)
- [Offline Architecture](#offline-architecture)
- [Error Handling Architecture](#error-handling-architecture)
- [Logging Architecture](#logging-architecture)
- [Analytics Architecture](#analytics-architecture)
- [Testing Architecture](#testing-architecture)
- [CI/CD Architecture](#cicd-architecture)
- [Performance Architecture](#performance-architecture)

---

## Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Language | Swift 5.9+ | Type-safe, memory-safe language |
| UI Framework | SwiftUI | Declarative UI rendering |
| Reactive | Combine | Asynchronous event handling |
| Persistence | CoreData | Structured local storage |
| Networking | URLSession | HTTP/HTTPS requests |
| Real-time | WebSocket | Live AI chat streaming |
| DI | EnvironmentObject | Dependency injection |
| Navigation | NavigationStack | Stack-based navigation |
| Security | Keychain Services | Secure credential storage |
| Biometric | LAContext | Face ID / Touch ID |
| Logging | OSLog | System logging |
| Analytics | Firebase Analytics | Event tracking |

```swift
// Package.swift dependencies
dependencies: [
    .package(url: "https://github.com/firebase/firebase-ios-sdk.git", from: "10.18.0"),
    .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.9.0"),
    .package(url: "https://github.com/pointfreeco/swift-composable-architecture", from: "1.7.0"),
    .package(url: "https://github.com/krzysztofzablocki/Sourcery.git", from: "2.0.0"),
    .package(url: "https://github.com/SwiftGen/SwiftGen.git", from: "6.6.0")
]
```

---

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    PRESENTATION LAYER                        │
│  ┌─────────┐  ┌─────────┐  ┌──────────┐  ┌──────────────┐ │
│  │  Views   │  │ViewModels│  │Components│  │NavigationCtrl│ │
│  └────┬────┘  └────┬────┘  └─────┬────┘  └──────┬───────┘ │
│       │            │             │               │          │
├───────┴────────────┴─────────────┴───────────────┴──────────┤
│                     DOMAIN LAYER                             │
│  ┌──────────┐  ┌──────────┐  ┌────────────┐  ┌──────────┐ │
│  │ Entities  │  │ UseCases │  │ Repository │  │  Errors  │ │
│  │           │  │          │  │ Protocols  │  │          │ │
│  └─────┬────┘  └─────┬────┘  └──────┬─────┘  └────┬─────┘ │
│        │             │              │              │         │
├────────┴─────────────┴──────────────┴──────────────┴────────┤
│                       DATA LAYER                             │
│  ┌──────────┐  ┌──────────┐  ┌────────────┐  ┌──────────┐ │
│  │ CoreData  │  │API Client│  │ WebSocket  │  │Keychain  │ │
│  │  Stack    │  │          │  │  Client    │  │ Service  │ │
│  └──────────┘  └──────────┘  └────────────┘  └──────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Responsibilities | Dependencies |
|-------|-----------------|--------------|
| Presentation | UI rendering, user interaction, view state | Domain |
| Domain | Business logic, entities, use case orchestration | None (pure) |
| Data | API calls, local persistence, data transformation | Domain |

---

## Module Structure

```
nexus-ai/
├── NexusAI/                          # App target
│   ├── App/
│   │   ├── NexusAIApp.swift          # @main entry point
│   │   ├── AppDelegate.swift         # UIApplicationDelegate
│   │   └── SceneDelegate.swift       # UISceneDelegate
│   ├── Resources/
│   │   ├── Assets.xcassets/
│   │   ├── Localizable.strings
│   │   └── Preview Content/
│   └── Info.plist
│
├── Features/                         # Feature modules
│   ├── Auth/
│   │   ├── Presentation/
│   │   │   ├── Views/
│   │   │   │   ├── LoginView.swift
│   │   │   │   ├── RegisterView.swift
│   │   │   │   ├── BiometricSetupView.swift
│   │   │   │   └── ForgotPasswordView.swift
│   │   │   └── ViewModels/
│   │   │       ├── LoginViewModel.swift
│   │   │       ├── RegisterViewModel.swift
│   │   │       └── BiometricViewModel.swift
│   │   └── Domain/
│   │       ├── Entities/
│   │       │   ├── User.swift
│   │       │   └── AuthTokens.swift
│   │       ├── UseCases/
│   │       │   ├── LoginUseCase.swift
│   │       │   ├── RegisterUseCase.swift
│   │       │   └── BiometricUseCase.swift
│   │       └── Repository/
│   │           └── AuthRepository.swift
│   │
│   ├── Chat/
│   │   ├── Presentation/
│   │   │   ├── Views/
│   │   │   │   ├── ChatView.swift
│   │   │   │   ├── MessageListView.swift
│   │   │   │   ├── MessageBubble.swift
│   │   │   │   ├── PromptBox.swift
│   │   │   │   ├── FileUploader.swift
│   │   │   │   ├── VoiceInput.swift
│   │   │   │   ├── AgentSelector.swift
│   │   │   │   ├── ModelIndicator.swift
│   │   │   │   ├── TokenStreamViewer.swift
│   │   │   │   ├── ThinkingIndicator.swift
│   │   │   │   └── ToolExecutionDisplay.swift
│   │   │   └── ViewModels/
│   │   │       └── ChatViewModel.swift
│   │   └── Domain/
│   │       ├── Entities/
│   │       │   ├── Message.swift
│   │       │   ├── Conversation.swift
│   │       │   └── ToolCall.swift
│   │       ├── UseCases/
│   │       │   ├── SendMessageUseCase.swift
│   │       │   ├── StreamResponseUseCase.swift
│   │       │   └── LoadHistoryUseCase.swift
│   │       └── Repository/
│   │           └── ChatRepository.swift
│   │
│   ├── Dashboard/
│   │   ├── Presentation/
│   │   └── Domain/
│   ├── Agents/
│   │   ├── Presentation/
│   │   └── Domain/
│   ├── Knowledge/
│   │   ├── Presentation/
│   │   └── Domain/
│   ├── Models/
│   │   ├── Presentation/
│   │   └── Domain/
│   ├── Settings/
│   │   ├── Presentation/
│   │   └── Domain/
│   └── Notifications/
│       ├── Presentation/
│       └── Domain/
│
├── Core/                             # Core infrastructure
│   ├── Networking/
│   │   ├── APIClient.swift
│   │   ├── URLSession+Configuration.swift
│   │   ├── AuthInterceptor.swift
│   │   ├── TokenRefreshInterceptor.swift
│   │   ├── CertificatePinning.swift
│   │   └── NetworkMonitor.swift
│   ├── WebSocket/
│   │   ├── WebSocketClient.swift
│   │   ├── WebSocketMessage.swift
│   │   ├── WebSocketReconnect.swift
│   │   └── WebSocketHeartbeat.swift
│   ├── Security/
│   │   ├── KeychainService.swift
│   │   ├── BiometricService.swift
│   │   └── CertificatePinner.swift
│   ├── Persistence/
│   │   ├── CoreDataStack.swift
│   │   ├── CoreDataModel.swift
│   │   └── SyncEngine.swift
│   └── DI/
│       └── DependencyContainer.swift
│
├── DesignSystem/                     # Reusable UI components
│   ├── Components/
│   │   ├── Buttons/
│   │   ├── TextFields/
│   │   ├── Cards/
│   │   ├── Alerts/
│   │   └── Indicators/
│   ├── Theme/
│   │   ├── ColorTheme.swift
│   │   ├── TypographyTheme.swift
│   │   └── SpacingTheme.swift
│   └── Tokens/
│       ├── Colors.swift
│       ├── Fonts.swift
│       └── Dimensions.swift
│
├── Resources/                        # Shared resources
│   ├── Assets.xcassets/
│   ├── Localizable.strings
│   └── Info.plist
│
├── Tests/
│   ├── UnitTests/
│   ├── IntegrationTests/
│   └── UITests/
│
└── Packages/                         # Local Swift packages
    ├── NexusAICore/
    ├── NexusAIDesignSystem/
    └── NexusAIFeatures/
```

---

## Dependency Flow

```
┌─────────────────────────────────────────────────────────┐
│                    OUTER LAYERS                          │
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │                FEATURES LAYER                     │  │
│  │  Auth | Chat | Dashboard | Agents | Knowledge    │  │
│  │  Models | Settings | Notifications               │  │
│  └──────────────────────┬───────────────────────────┘  │
│                         │ depends on                    │
│                         ▼                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │                DESIGN SYSTEM                      │  │
│  │  Components | Theme | Tokens                     │  │
│  └──────────────────────┬───────────────────────────┘  │
│                         │ depends on                    │
│                         ▼                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │                 CORE LAYER                        │  │
│  │  Networking | Security | Persistence | DI        │  │
│  └──────────────────────┬───────────────────────────┘  │
│                         │ depends on                    │
│                         ▼                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │              RESOURCES / FRAMEWORKS               │  │
│  │  SwiftUI | Combine | CoreData | Keychain         │  │
│  └──────────────────────────────────────────────────┘  │
│                                                         │
│                   INNER LAYERS                          │
└─────────────────────────────────────────────────────────┘
```

### Dependency Rules

| Rule | Description |
|------|-------------|
| Outer → Inner | Outer layers may depend on inner layers |
| No Inner → Outer | Inner layers must never depend on outer layers |
| Feature Isolation | Features must not depend on each other |
| Core Independence | Core is independent of Features |
| DI for Cross-Cutting | Use DependencyContainer for cross-cutting concerns |

---

## Navigation Architecture

```swift
// Root navigation with NavigationStack
@main
struct NexusAIApp: App {
    @StateObject private var appRouter = AppRouter()

    var body: some Scene {
        WindowGroup {
            NavigationStack(path: $appRouter.path) {
                RootView()
                    .navigationDestination(for: Route.self) { route in
                        route.destination
                    }
                    .sheet(item: $appRouter.sheet) { sheet in
                        sheet.destination
                    }
                    .fullScreenCover(item: $appRouter.fullScreen) { cover in
                        cover.destination
                    }
            }
            .environmentObject(appRouter)
        }
    }
}
```

```swift
// Route definitions
enum Route: Hashable {
    case login
    case register
    case chat(conversationId: String)
    case dashboard
    case agents
    case knowledge
    case settings
    case profile

    @ViewBuilder
    var destination: some View {
        switch self {
        case .login:
            LoginView()
        case .register:
            RegisterView()
        case .chat(let id):
            ChatView(conversationId: id)
        case .dashboard:
            DashboardView()
        case .agents:
            AgentsView()
        case .knowledge:
            KnowledgeView()
        case .settings:
            SettingsView()
        case .profile:
            ProfileView()
        }
    }
}
```

```swift
// AppRouter with deep link support
@MainActor
class AppRouter: ObservableObject {
    @Published var path = NavigationPath()
    @Published var sheet: Sheet?
    @Published var fullScreen: FullScreenCover?

    func navigate(to route: Route) {
        path.append(route)
    }

    func navigateBack() {
        path.removeLast()
    }

    func navigateToRoot() {
        path = NavigationPath()
    }

    func presentSheet(_ sheet: Sheet) {
        self.sheet = sheet
    }

    func handleDeepLink(_ url: URL) {
        guard let components = URLComponents(url: url, resolvingAgainstBaseURL: true),
              let host = components.host else { return }

        switch host {
        case "chat":
            if let id = components.queryItems?.first(where: { $0.name == "id" })?.value {
                navigate(to: .chat(conversationId: id))
            }
        case "dashboard":
            navigate(to: .dashboard)
        case "settings":
            navigate(to: .settings)
        default:
            break
        }
    }
}
```

```swift
// Navigation path data for state restoration
struct NavigationPathData: Codable {
    let routes: [RouteData]

    enum RouteData: Codable {
        case login
        case chat(conversationId: String)
        case dashboard
        case agents
        case knowledge
        case settings
    }
}
```

### Navigation Flow Diagram

```
┌──────────────┐
│  RootView     │
│  (Auth Check) │
└──────┬───────┘
       │
       ├── Not Authenticated ──→ LoginView ──→ RegisterView
       │
       ├── Authenticated ──→ TabView
       │                       │
       │                       ├── Dashboard Tab
       │                       │   └── NavigationStack
       │                       │       ├── DashboardView
       │                       │       ├── StatsView
       │                       │       └── DetailView
       │                       │
       │                       ├── Chat Tab
       │                       │   └── NavigationStack
       │                       │       ├── ChatListView
       │                       │       ├── ChatView
       │                       │       └── Settings
       │                       │
       │                       ├── Agents Tab
       │                       │   └── NavigationStack
       │                       │       ├── AgentsListView
       │                       │       └── AgentDetailView
       │                       │
       │                       └── Settings Tab
       │                           └── NavigationStack
       │                               ├── SettingsView
       │                               ├── ProfileView
       │                               └── AboutView
       │
       └── Deep Link ──→ Direct Route
```

---

## State Management Architecture

```swift
// MVVM pattern with ObservableObject
@MainActor
class ChatViewModel: ObservableObject {
    // MARK: - Published State
    @Published var messages: [Message] = []
    @Published var inputText: String = ""
    @Published var isLoading: Bool = false
    @Published var streamingToken: String?
    @Published var error: ChatError?
    @Published var isTyping: Bool = false
    @Published var selectedAgent: Agent?
    @Published var attachments: [Attachment] = []
    @Published var isConnected: Bool = false

    // MARK: - Dependencies
    private let chatUseCase: ChatUseCaseProtocol
    private let webSocketClient: WebSocketClientProtocol
    private var cancellables = Set<AnyCancellable>()
    private let taskId = UUID()

    // MARK: - State Machine
    enum State: Equatable {
        case idle
        case connecting
        case connected
        case streaming
        case error(String)
    }

    @Published var state: State = .idle

    // MARK: - Init
    init(
        chatUseCase: ChatUseCaseProtocol = ChatUseCase(),
        webSocketClient: WebSocketClientProtocol = WebSocketClient()
    ) {
        self.chatUseCase = chatUseCase
        self.webSocketClient = webSocketClient
        setupWebSocketBindings()
    }

    // MARK: - WebSocket Bindings
    private func setupWebSocketBindings() {
        webSocketClient.onTokenReceived { [weak self] token in
            self?.state = .streaming
            self?.streamingToken = (self?.streamingToken ?? "") + token
        }

        webSocketClient.onCompleted { [weak self] response in
            self?.handleCompletion(response)
        }

        webSocketClient.onToolCall { [weak self] toolCall in
            self?.handleToolCall(toolCall)
        }

        webSocketClient.onThinking { [weak self] thinking in
            self?.handleThinking(thinking)
        }

        webSocketClient.onError { [weak self] error in
            self?.state = .error(error.localizedDescription)
        }
    }

    // MARK: - Actions
    func sendMessage() {
        let text = inputText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }

        let userMessage = Message(
            id: UUID(),
            content: text,
            role: .user,
            timestamp: Date(),
            attachments: attachments
        )

        messages.append(userMessage)
        inputText = ""
        attachments = []
        state = .streaming
        isLoading = true

        webSocketClient.send(
            .chat(
                conversationId: currentConversationId,
                message: text,
                agentId: selectedAgent?.id,
                attachments: userMessage.attachments
            )
        )
    }
}
```

```swift
// Combine-based state pipeline
extension ChatViewModel {
    func setupCombinePipeline() {
        // Debounce input for search suggestions
        $inputText
            .debounce(for: .milliseconds(300), scheduler: RunLoop.main)
            .removeDuplicates()
            .filter { !$0.isEmpty }
            .sink { [weak self] text in
                self?.fetchSuggestions(for: text)
            }
            .store(in: &cancellables)

        // Auto-reconnect on connection loss
        $state
            .filter { $0 == .error("Connection lost") }
            .delay(for: .seconds(5), scheduler: RunLoop.main)
            .sink { [weak self] _ in
                self?.reconnect()
            }
            .store(in: &cancellables)

        // Batch streaming tokens
        $streamingToken
            .compactMap { $0 }
            .collect(.byTime(RunLoop.main, .milliseconds(50)))
            .sink { [weak self] tokens in
                self?.renderBatch(tokens)
            }
            .store(in: &cancellables)
    }
}
```

---

## Networking Architecture

```swift
// APIClient with async/await and Combine
protocol APIClientProtocol {
    func request<T: Decodable>(_ endpoint: Endpoint) async throws -> T
    func request<T: Decodable>(_ endpoint: Endpoint) -> AnyPublisher<T, APIError>
}

class APIClient: APIClientProtocol {
    private let session: URLSession
    private let interceptor: AuthInterceptorProtocol
    private let decoder: JSONDecoder

    init(
        session: URLSession = .shared,
        interceptor: AuthInterceptorProtocol = AuthInterceptor()
    ) {
        self.session = session
        self.interceptor = interceptor
        self.decoder = JSONDecoder()
        self.decoder.keyDecodingStrategy = .convertFromSnakeCase
        self.decoder.dateDecodingStrategy = .iso8601
    }

    func request<T: Decodable>(_ endpoint: Endpoint) async throws -> T {
        var request = endpoint.urlRequest
        try await interceptor.intercept(&request)

        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        switch httpResponse.statusCode {
        case 200...299:
            return try decoder.decode(T.self, from: data)
        case 401:
            try await interceptor.handleUnauthorized()
            return try await request(endpoint)
        case 429:
            throw APIError.rateLimited
        default:
            throw APIError.httpError(statusCode: httpResponse.statusCode)
        }
    }
}
```

```swift
// Endpoint definitions
protocol Endpoint {
    var path: String { get }
    var method: HTTPMethod { get }
    var headers: [String: String] { get }
    var body: Data? { get }
    var queryItems: [URLQueryItem] { get }

    var urlRequest: URLRequest { get }
}

enum APIEndpoint: Endpoint {
    case login(email: String, password: String)
    case refreshToken(token: String)
    case chat(conversationId: String, message: String)
    case conversations
    case agents
    case knowledge

    var path: String {
        switch self {
        case .login: return "/api/v1/auth/login"
        case .refreshToken: return "/api/v1/auth/refresh"
        case .chat(let id, _): return "/api/v1/chat/\(id)"
        case .conversations: return "/api/v1/conversations"
        case .agents: return "/api/v1/agents"
        case .knowledge: return "/api/v1/knowledge"
        }
    }

    var method: HTTPMethod {
        switch self {
        case .login, .refreshToken, .chat:
            return .post
        case .conversations, .agents, .knowledge:
            return .post
        }
    }
}
```

### Network Flow Diagram

```
┌─────────┐     ┌──────────────┐     ┌───────────────┐     ┌────────┐
│  View   │────▶│  ViewModel   │────▶│  UseCase      │────▶│  API   │
│         │     │              │     │               │     │ Client │
└─────────┘     └──────────────┘     └───────────────┘     └───┬────┘
                                                               │
                                                               ▼
                                                         ┌───────────┐
                                                         │Interceptors│
                                                         │ - Auth    │
                                                         │ - Refresh │
                                                         │ - Pinning │
                                                         └─────┬─────┘
                                                               │
                                                               ▼
                                                         ┌───────────┐
                                                         │ URLSession│
                                                         │ (Server)  │
                                                         └───────────┘
```

---

## Local Storage Architecture

```swift
// CoreData Stack
class CoreDataStack {
    static let shared = CoreDataStack()

    lazy var persistentContainer: NSPersistentContainer = {
        let container = NSPersistentContainer(name: "NexusAI")
        container.loadPersistentStores { description, error in
            if let error = error {
                fatalError("CoreData load failed: \(error)")
            }
        }
        container.viewContext.automaticallyMergesChangesFromParent = true
        container.viewContext.mergePolicy = NSMergeByPropertyObjectTrumpMergePolicy
        return container
    }()

    var context: NSManagedObjectContext {
        persistentContainer.viewContext
    }

    func newBackgroundContext() -> NSManagedObjectContext {
        persistentContainer.newBackgroundContext()
    }

    func save() {
        guard context.hasChanges else { return }
        do {
            try context.save()
        } catch {
            os_log(.error, "CoreData save failed: %@", error.localizedDescription)
        }
    }
}
```

```swift
// Keychain Service
class KeychainService {
    static let shared = KeychainService()

    func save(_ data: Data, for key: String) throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]

        SecItemDelete(query as CFDictionary)
        let status = SecItemAdd(query as CFDictionary, nil)

        guard status == errSecSuccess else {
            throw KeychainError.saveFailed(status)
        }
    }

    func load(for key: String) throws -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]

        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)

        guard status == errSecSuccess else {
            if status == errSecItemNotFound { return nil }
            throw KeychainError.loadFailed(status)
        }

        return item as? Data
    }

    func delete(for key: String) throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key
        ]

        let status = SecItemDelete(query as CFDictionary)
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw KeychainError.deleteFailed(status)
        }
    }
}
```

### Storage Architecture Table

| Storage Type | Use Case | Security | Persistence |
|-------------|----------|----------|-------------|
| CoreData | Messages, Conversations, Agents | File Protection | Until deleted |
| Keychain | Tokens, Biometric keys | Hardware-encrypted | Until deleted |
| UserDefaults | Preferences, Settings | None | Until deleted |
| File System | Images, Documents | File Protection | Until deleted |
| In-Memory | Cache, State | None | App lifecycle |

---

## Background Work Architecture

```swift
// BGTaskScheduler for background tasks
class BackgroundTaskManager {
    static let shared = BackgroundTaskManager()

    private let refreshTaskId = "com.nexusai.refresh"
    private let syncTaskId = "com.nexusai.sync"

    func registerTasks() {
        BGTaskScheduler.shared.register(
            forTaskWithIdentifier: refreshTaskId,
            using: nil
        ) { task in
            self.handleRefresh(task: task as! BGAppRefreshTask)
        }

        BGTaskScheduler.shared.register(
            forTaskWithIdentifier: syncTaskId,
            using: nil
        ) { task in
            self.handleSync(task: task as! BGProcessingTask)
        }
    }

    func scheduleRefresh() {
        let request = BGAppRefreshTaskRequest(identifier: refreshTaskId)
        request.earliestBeginDate = Date(timeIntervalSinceNow: 15 * 60)

        try? BGTaskScheduler.shared.submit(request)
    }

    func scheduleSync() {
        let request = BGProcessingTaskRequest(identifier: syncTaskId)
        request.requiresNetworkConnectivity = true
        request.requiresExternalPower = false

        try? BGTaskScheduler.shared.submit(request)
    }

    private func handleRefresh(task: BGAppRefreshTask) {
        let refreshOperation = RefreshOperation()
        task.expirationHandler = {
            refreshOperation.cancel()
        }

        refreshOperation.completionBlock = {
            task.setTaskCompleted(success: !refreshOperation.isCancelled)
            self.scheduleRefresh()
        }

        OperationQueue.main.addOperation(refreshOperation)
    }

    private func handleSync(task: BGProcessingTask) {
        let syncOperation = SyncOperation()
        task.expirationHandler = {
            syncOperation.cancel()
        }

        syncOperation.completionBlock = {
            task.setTaskCompleted(success: !syncOperation.isCancelled)
        }

        OperationQueue.main.addOperation(syncOperation)
    }
}
```

---

## Push Notification Architecture

```swift
// Push Notification Manager
class PushNotificationManager: NSObject, ObservableObject {
    static let shared = PushNotificationManager()

    @Published var deviceToken: Data?

    func registerForNotifications() {
        UNUserNotificationCenter.current().requestAuthorization(
            options: [.alert, .badge, .sound]
        ) { granted, error in
            DispatchQueue.main.async {
                if granted {
                    UIApplication.shared.registerForRemoteNotifications()
                }
            }
        }

        UNUserNotificationCenter.current().delegate = self
    }

    func didRegisterForRemoteNotifications(with deviceToken: Data) {
        self.deviceToken = deviceToken
        let token = deviceToken.map { String(format: "%02.2hhx", $0) }.joined()
        uploadToken(token)
    }

    private func uploadToken(_ token: String) {
        // Send token to server
    }
}

extension PushNotificationManager: UNUserNotificationCenterDelegate {
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification
    ) async -> UNNotificationPresentationOptions {
        [.banner, .badge, .sound]
    }

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse
    ) async {
        let userInfo = response.notification.request.content.userInfo
        handleNotification(userInfo)
    }

    private func handleNotification(_ userInfo: [AnyHashable: Any]) {
        guard let type = userInfo["type"] as? String else { return }

        switch type {
        case "new_message":
            if let conversationId = userInfo["conversationId"] as? String {
                // Navigate to chat
            }
        case "agent_complete":
            // Navigate to agents
            break
        default:
            break
        }
    }
}
```

---

## Biometric Architecture

```swift
// Biometric Authentication Service
class BiometricService {
    static let shared = BiometricService()

    private let context = LAContext()

    enum BiometricType {
        case none
        case touchID
        case faceID
    }

    var biometricType: BiometricType {
        guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: nil) else {
            return .none
        }

        switch context.biometryType {
        case .touchID: return .touchID
        case .faceID: return .faceID
        case .opticID: return .none
        @unknown default: return .none
        }
    }

    var isAvailable: Bool {
        var error: NSError?
        return context.canEvaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics,
            error: &error
        )
    }

    func authenticate(reason: String) async throws -> Bool {
        return try await withCheckedThrowingContinuation { continuation in
            context.evaluatePolicy(
                .deviceOwnerAuthenticationWithBiometrics,
                localizedReason: reason
            ) { success, error in
                if let error = error {
                    continuation.resume(throwing: error)
                } else {
                    continuation.resume(returning: success)
                }
            }
        }
    }

    func storeBiometricKey(_ key: String) throws {
        guard isAvailable else { throw BiometricError.notAvailable }

        let accessControl = SecAccessControlCreateWithFlags(
            kCFAllocatorDefault,
            kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
            .biometryCurrentSet,
            nil
        )!

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: "biometric_key",
            kSecValueData as String: key.data(using: .utf8)!,
            kSecAttrAccessControl as String: accessControl
        ]

        SecItemDelete(query as CFDictionary)
        let status = SecItemAdd(query as CFDictionary, nil)

        guard status == errSecSuccess else {
            throw BiometricError.keychainError(status)
        }
    }
}
```

---

## Offline Architecture

```swift
// Sync Engine for offline-first architecture
class SyncEngine: ObservableObject {
    @Published var syncStatus: SyncStatus = .idle
    @Published var lastSyncDate: Date?

    private let coreDataStack: CoreDataStack
    private let apiClient: APIClientProtocol
    private var syncTimer: Timer?

    enum SyncStatus {
        case idle
        case syncing(progress: Double)
        case completed
        case error(String)
    }

    init(
        coreDataStack: CoreDataStack = .shared,
        apiClient: APIClientProtocol = APIClient()
    ) {
        self.coreDataStack = coreDataStack
        self.apiClient = apiClient
        setupSyncTimer()
        observeConnectivity()
    }

    func syncAll() async {
        syncStatus = .syncing(progress: 0)

        do {
            try await syncMessages()
            syncStatus = .syncing(progress: 0.5)

            try await syncConversations()
            syncStatus = .syncing(progress: 0.8)

            try await syncKnowledge()
            syncStatus = .completed
            lastSyncDate = Date()
        } catch {
            syncStatus = .error(error.localizedDescription)
        }
    }

    private func syncMessages() async throws {
        let context = coreDataStack.newBackgroundContext()
        let pendingMessages = try context.fetch(PendingMessage.fetchRequest())

        for (index, message) in pendingMessages.enumerated() {
            try await sendPendingMessage(message)
            let progress = Double(index) / Double(pendingMessages.count) * 0.5
            await MainActor.run { syncStatus = .syncing(progress: progress) }

            message.synced = true
        }

        try context.save()
    }

    private func observeConnectivity() {
        NWPathMonitor().pathUpdateHandler = { [weak self] path in
            if path.status == .satisfied {
                Task { await self?.syncAll() }
            }
        }
    }
}
```

### Offline Architecture Diagram

```
┌────────────────────────────────────────────────────────┐
│                 OFFLINE-ARCHITECTURE                    │
│                                                        │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────┐│
│  │  User    │    │   CoreData   │    │   Sync Queue ││
│  │  Input   │───▶│   (Pending)  │───▶│   (Outbox)   ││
│  └──────────┘    └──────────────┘    └──────┬───────┘│
│                                             │        │
│                                        ┌────▼────┐   │
│                                        │Network  │   │
│                                        │Monitor  │   │
│                                        └────┬────┘   │
│                                             │        │
│  ┌──────────┐    ┌──────────────┐    ┌─────▼──────┐│
│  │  Local   │◀───│  CoreData    │◀───│  API       ││
│  │  Cache   │    │  (Synced)    │    │  Client    ││
│  └──────────┘    └──────────────┘    └────────────┘│
└────────────────────────────────────────────────────────┘
```

---

## Error Handling Architecture

```swift
// Structured error handling
protocol AppError: LocalizedError {
    var code: Int { get }
    var domain: String { get }
    var underlyingError: Error? { get }
}

// Network Errors
enum APIError: AppError {
    case invalidURL
    case invalidResponse
    case httpError(statusCode: Int)
    case decodingError(Error)
    case rateLimited
    case timeout
    case noConnection
    case unauthorized
    case serverError(String)

    var code: Int {
        switch self {
        case .invalidURL: return 1001
        case .invalidResponse: return 1002
        case .httpError(let code): return code
        case .decodingError: return 1003
        case .rateLimited: return 1029
        case .timeout: return 1004
        case .noConnection: return 1005
        case .unauthorized: return 1006
        case .serverError: return 1007
        }
    }

    var domain: String { "com.nexusai.api" }

    var underlyingError: Error? { nil }

    var errorDescription: String? {
        switch self {
        case .invalidURL: return "Invalid URL"
        case .invalidResponse: return "Invalid server response"
        case .httpError(let code): return "HTTP Error: \(code)"
        case .decodingError: return "Failed to decode response"
        case .rateLimited: return "Rate limited. Please try again later"
        case .timeout: return "Request timed out"
        case .noConnection: return "No internet connection"
        case .unauthorized: return "Authentication required"
        case .serverError(let msg): return "Server error: \(msg)"
        }
    }
}

// Error Boundary for ViewModels
@MainActor
class ErrorBoundary: ObservableObject {
    @Published var currentError: AppError?
    @Published var showError: Bool = false

    func handle<T>(_ operation: () async throws -> T) async -> T? {
        do {
            return try await operation()
        } catch let error as AppError {
            currentError = error
            showError = true
            return nil
        } catch {
            currentError = APIError.serverError(error.localizedDescription)
            showError = true
            return nil
        }
    }
}
```

```swift
// Error alert modifier
struct ErrorAlertModifier: ViewModifier {
    @EnvironmentObject var errorBoundary: ErrorBoundary

    func body(content: Content) -> some View {
        content
            .alert(
                "Error",
                isPresented: $errorBoundary.showError,
                presenting: errorBoundary.currentError
            ) { _ in
                Button("OK") { errorBoundary.currentError = nil }
            } message: { error in
                Text(error.localizedDescription)
            }
    }
}
```

---

## Logging Architecture

```swift
// Centralized logging with OSLog
import OSLog

enum LogCategory: String {
    case network = "Network"
    case ui = "UI"
    case data = "Data"
    case auth = "Auth"
    case chat = "Chat"
    case analytics = "Analytics"
    case general = "General"
}

struct AppLogger {
    private let logger: os.Logger

    init(category: LogCategory) {
        self.logger = os.Logger(
            subsystem: Bundle.main.bundleIdentifier ?? "com.nexusai",
            category: category.rawValue
        )
    }

    func debug(_ message: String) {
        logger.debug("\(message, privacy: .public)")
    }

    func info(_ message: String) {
        logger.info("\(message, privacy: .public)")
    }

    func error(_ message: String, error: Error? = nil) {
        if let error = error {
            logger.error("\(message): \(error.localizedDescription, privacy: .public)")
        } else {
            logger.error("\(message, privacy: .public)")
        }
    }

    func fault(_ message: String) {
        logger.fault("\(message, privacy: .public)")
    }
}

// Usage
let networkLogger = AppLogger(category: .network)
let chatLogger = AppLogger(category: .chat)

networkLogger.info("POST /api/v1/conversations - 200 OK")
chatLogger.error("WebSocket connection failed", error: connectionError)
```

---

## Analytics Architecture

```swift
// Firebase Analytics wrapper
enum AnalyticsEvent {
    case login(method: String)
    case logout
    case sendMessage(conversationId: String, agentId: String?)
    case viewAgent(agentId: String)
    case viewKnowledge(knowledgeId: String)
    case search(query: String)
    case error(domain: String, code: Int)
    case performance(name: String, duration: TimeInterval)

    var name: String {
        switch self {
        case .login: return "login"
        case .logout: return "logout"
        case .sendMessage: return "send_message"
        case .viewAgent: return "view_agent"
        case .viewKnowledge: return "view_knowledge"
        case .search: return "search"
        case .error: return "error"
        case .performance: return "performance"
        }
    }

    var parameters: [String: Any] {
        switch self {
        case .login(let method):
            return ["method": method]
        case .logout:
            return [:]
        case .sendMessage(let conversationId, let agentId):
            var params: [String: Any] = ["conversation_id": conversationId]
            if let agentId = agentId { params["agent_id"] = agentId }
            return params
        case .viewAgent(let id):
            return ["agent_id": id]
        case .viewKnowledge(let id):
            return ["knowledge_id": id]
        case .search(let query):
            return ["query": query]
        case .error(let domain, let code):
            return ["domain": domain, "code": code]
        case .performance(let name, let duration):
            return ["name": name, "duration_ms": duration * 1000]
        }
    }
}

class AnalyticsManager {
    static let shared = AnalyticsManager()

    func log(_ event: AnalyticsEvent) {
        Analytics.logEvent(event.name, parameters: event.parameters)
    }

    func setUserProperty(_ value: String?, forName name: String) {
        Analytics.setUserProperty(value, forName: name)
    }

    func setUserID(_ userID: String?) {
        Analytics.setUserID(userID)
    }
}
```

---

## Testing Architecture

```swift
// Unit Test with Mock
class ChatViewModelTests: XCTestCase {
    var sut: ChatViewModel!
    var mockChatUseCase: MockChatUseCase!
    var mockWebSocket: MockWebSocketClient!

    override func setUp() {
        super.setUp()
        mockChatUseCase = MockChatUseCase()
        mockWebSocket = MockWebSocketClient()
        sut = ChatViewModel(
            chatUseCase: mockChatUseCase,
            webSocketClient: mockWebSocket
        )
    }

    override func tearDown() {
        sut = nil
        mockChatUseCase = nil
        mockWebSocket = nil
        super.tearDown()
    }

    func testSendMessage() {
        // Given
        sut.inputText = "Hello, AI"
        let expectation = XCTestExpectation(description: "Message sent")

        // When
        sut.sendMessage()

        // Then
        XCTAssertEqual(sut.messages.count, 1)
        XCTAssertEqual(sut.messages.first?.content, "Hello, AI")
        XCTAssertEqual(sut.messages.first?.role, .user)
        XCTAssertTrue(sut.inputText.isEmpty)
    }

    func testSendEmptyMessageDoesNothing() {
        sut.inputText = ""
        sut.sendMessage()
        XCTAssertTrue(sut.messages.isEmpty)
    }
}

// Mock for testing
class MockChatUseCase: ChatUseCaseProtocol {
    var sendMessageResult: Result<Message, Error> = .success(
        Message(id: UUID(), content: "Mock response", role: .assistant, timestamp: Date())
    )

    func sendMessage(_ message: String, conversationId: String) async throws -> Message {
        return try sendMessageResult.get()
    }
}

// UI Test
class ChatUITests: XCTestCase {
    var app: XCUIApplication!

    override func setUp() {
        super.setUp()
        app = XCUIApplication()
        app.launch()
    }

    func testSendMessage() {
        let textField = app.textFields["prompt_input"]
        let sendButton = app.buttons["send_button"]

        textField.tap()
        textField.typeText("Hello")
        sendButton.tap()

        let messageCell = app.staticTexts["Hello"]
        XCTAssertTrue(messageCell.exists)
    }
}
```

---

## CI/CD Architecture

```yaml
# Xcode Cloud workflow
name: iOS CI/CD
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: macos-14
    steps:
      - uses: actions/checkout@v4

      - name: Select Xcode
        run: sudo xcode-select -s /Applications/Xcode_15.2.app

      - name: Build
        run: |
          xcodebuild build \
            -scheme NexusAI \
            -destination 'platform=iOS Simulator,name=iPhone 15 Pro' \
            -configuration Debug

      - name: Unit Tests
        run: |
          xcodebuild test \
            -scheme NexusAI \
            -destination 'platform=iOS Simulator,name=iPhone 15 Pro' \
            -only-testing:NexusAIUnitTests

      - name: UI Tests
        run: |
          xcodebuild test \
            -scheme NexusAI \
            -destination 'platform=iOS Simulator,name=iPhone 15 Pro' \
            -only-testing:NexusAIUITests

  deploy:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: macos-14
    steps:
      - name: Fastlane Deploy
        run: |
          fastlane ios beta
```

```ruby
# Fastlane configuration
default_platform(:ios)

platform :ios do
  desc "Run unit tests"
  lane :test do
    scan(
      scheme: "NexusAI",
      devices: ["iPhone 15 Pro"],
      clean: true
    )
  end

  desc "Build and deploy to TestFlight"
  lane :beta do
    increment_build_number
    build_app(
      scheme: "NexusAI",
      export_method: "app-store"
    )
    upload_to_testflight
  end

  desc "Deploy to App Store"
  lane :release do
    build_app(
      scheme: "NexusAI",
      export_method: "app-store"
    )
    upload_to_app_store(
      force: true,
      skip_screenshots: true
    )
  end
end
```

---

## Performance Architecture

```swift
// Performance monitoring
class PerformanceMonitor {
    static let shared = PerformanceMonitor()

    func measure<T>(_ name: String, operation: () throws -> T) rethrows -> T {
        let start = CFAbsoluteTimeGetCurrent()
        let result = try operation()
        let end = CFAbsoluteTimeGetCurrent()
        let duration = (end - start) * 1000

        os_log(.info, "⏱ %{public}s: %.2fms", name, duration)

        AnalyticsManager.shared.log(.performance(name: name, duration: duration / 1000))

        return result
    }

    func measureAsync<T>(_ name: String, operation: () async throws -> T) async rethrows -> T {
        let start = CFAbsoluteTimeGetCurrent()
        let result = try await operation()
        let end = CFAbsoluteTimeGetCurrent()
        let duration = (end - start) * 1000

        os_log(.info, "⏱ %{public}s: %.2fms", name, duration)

        return result
    }
}
```

```swift
// Image caching for performance
class ImageCache {
    static let shared = ImageCache()

    private let cache = NSCache<NSString, UIImage>()
    private let memoryLimit = 50 * 1024 * 1024 // 50MB

    init() {
        cache.totalCostLimit = memoryLimit
    }

    func image(for url: URL) -> UIImage? {
        cache.object(forKey: url.absoluteString as NSString)
    }

    func setImage(_ image: UIImage, for url: URL) {
        let cost = Int(image.size.width * image.size.height * image.scale * 4)
        cache.setObject(image, forKey: url.absoluteString as NSString, cost: cost)
    }
}
```

### Performance Benchmarks

| Metric | Target | Measurement Tool |
|--------|--------|-----------------|
| App Launch | < 2s | Instruments - App Launch |
| Screen Load | < 500ms | Instruments - Time Profiler |
| API Response | < 300ms | Network Instrument |
| WebSocket Connect | < 1s | Custom logging |
| Memory Usage | < 150MB | Instruments - Leaks |
| CPU Usage | < 30% idle | Instruments - CPU |
| Battery Impact | Minimal | Energy Diagnostics |
| App Size | < 50MB | Xcode Organizer |
