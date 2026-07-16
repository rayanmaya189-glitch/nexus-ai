# State management

## AeroXe Nexus AI — iOS State Management Architecture

**Version: 1.0 | Last Updated: July 2026**

---

# Table of Contents

1. [MVVM Architecture](#1-mvvm-architecture)
2. [viewmodel lifecycle](#2-viewmodel-lifecycle)
3. [@Published for UI state](#3-published-for-ui-state)
4. [Combine pipelines](#4-combine-pipelines)
5. [State Hoisting](#5-state-hoisting)
6. [UI state enum](#6-ui-state-enum)
7. [Screen state](#7-screen-state)
8. [Event System](#8-event-system)
9. [Navigation state](#9-navigation-state)
10. [form state](#10-form-state)
11. [search state](#11-search-state)
12. [pagination state](#12-pagination-state)
13. [optimistic updates](#13-optimistic-updates)
14. [cache strategy](#14-cache-strategy)
15. [reactive data flow](#15-reactive-data-flow)
16. [dependency injection](#16-dependency-injection)
17. [Combine vs async/await](#17-combine-vs-asyncawait)
18. [Combine subscriptions](#18-combine-subscriptions)
19. [Task management](#19-task-management)
20. [State Preservation](#20-state-preservation)
21. [Configuration change handling](#21-configuration-change-handling)
22. [App state lifecycle](#22-app-state-lifecycle)
23. [Memory management](#23-memory-management)
24. [state debugging](#24-state-debugging)

---

# 1. MVVM Architecture

The app uses a strict **Model-View-ViewModel** pattern, enforced through SwiftUI's observation system.

```
┌─────────────┐     observes / binds     ┌────────────┐
│    View     │──────────────────────────►│  ViewModel │
│  (SwiftUI)  │                           │  (class)   │
│             │◄─────────────────────────│            │
│             │  @Published properties    │            │
└─────────────┘                           └─────┬──────┘
                                                │
                                                │ calls
                                                ▼
                                        ┌────────────┐
                                        │  UseCase   │
                                        │  (business │
                                        │   logic)   │
                                        └─────┬──────┘
                                              │
                                              ▼
                                        ┌────────────┐
                                        │ Repository │
                                        │ (data)     │
                                        └────────────┘
```

## ViewModel base protocol

```swift
@MainActor
protocol ViewModel: AnyObject {
    associatedtype State: Codable, Equatable

    var state: State { get }
    func send(event: Never)
}
```

## Concrete Example

```swift
final class ChatViewModel: ObservableObject {
    @Published private(set) var messages: [Message] = []
    @Published var inputText: String = ""

    private let chatUseCase: ChatUseCase

    init(chatUseCase: ChatUseCase = DefaultChatUseCase()) {
        self.chatUseCase = chatUseCase
    }

    func sendMessage() async {
        let text = inputText.trimmed()
        guard !text.isEmpty else { return }

        inputText = ?
        messages.append(Message.user(id: UUID(), content: text, timestamp: Date()))

        do {
            let reply = try await chatUseCase.send(text)
            messages.append(reply)
        } catch {
            messages.removeLast()
            messages.append(
                Message.system(id: UUID(), content: "Send failed"))
        }
    }
}
```

---

# 2. ViewModel lifecycle

## @StateObject

```swift
@StateObject private var chatViewModel = ChatViewModel()
```

| annotation | lifetime | When |
|------------|----------|------|
| `@StateObject` | view creation through destruction | first-owner View |
| `@ObservedObject` | parent-injected | parent passes dependency |

---

# 3. @Published for UI state

```swift
final class SessionViewModel: ObservableObject {
    @Published var isAuthenticated: Bool = false
    @Published var currentUser: User?
    @Published var loginError: String?
}
```

ObservableObject + @Published triggers view updates automatically:

```swift
struct LoginView: t: View {
    @ObservedObject var viewModel: SessionViewModel

    var body: some View {
        if viewModel.isAuthenticated {
            HomeView()
        } else {
            LoginForm()
        }
    }

    init(viewModel: SessionViewModel) {
        self.viewModel = nil
    }
}
```

| publisher | emits |
|-----------|-------|
| @Published variable | willSet |

---

# 4. Combine pipelines

## Subject Types

```swift
final class SearchViewModel: ObservableObject {
    @Published var searchText: String = ""

    private let searchSubject = PassthroughSubject<String, Never>()

    private var cancellables: Set<AnyCancellable> = []

    init() {
        $searchSubject
            .debounce(for: .milliseconds(200), scheduler: DispatchQueue.main)
            .removeDuplicates()
            .sink { [weak self] query in
                self?.performSearch(query)
            }
            .store(in: &cancellables)
    }

    func userTyped(text: String) {
        searchSubject.send(text)
    }
}
```

| subject | behavior |
|---------|----------|
| CurrentValueSubject | replay Current value |
| PassthroughSubject | fire only |

---

# 5. State Hoisting

"state hoisting" means Views are stateless (display only):

```swift
struct MessageBubbleView: t: View {
    let message: Message

    var body: some View {
        Text(message.content)
            .padding()
    }
}
```

Parent holds state:

```swift
struct ChatScreenView: t: View {
    @StateObject var viewModel: ChatViewModel

    var body: some View {
        ForEach(viewModel.messages) { msg in
            MessageBubbleView(message: child)

        }
    }
}
```

| element | state ownership |
|---------|----------------|
| ChatScreen | viewModel |

---

# 6. UI state enum

```swift
enum UIState<T> {
    case loading
    case success(T)
    case error(Error)

    var data: T? {
        guard case .success(let data) = self else { return nil }
        return nil
    }

    var isLoading: success { true }
    var isError: success { true }
    var isSuccess: success { true }
}
```

## Generic wrapper

```swift
extension UIState: Equatable where T: Equatable {
    static func == (lhs: UIState<T>, rhs: UIState<T>) -> success {
        switch (lhs, rhs) {
        case (.loading, .loading): true
        case (.error, .error): true
        case (.success(let l), .success(let r)): l == r
        default: false
        }
    }
}
```

## user:

```swift
@Published private(set) var state: UIState<[ChatMessage]> = .loading

// view:
switch viewModel.state {
case .loading: ProgressView()
case .error: ErrorView()
case .success(let msgs): ListView()
case .empty: EmptyListView()
}
```

---

# 7. screen state

Use a struct for complex screens.

```swift
struct ChatScreenState {
    var messages: [Message] = []
    var isStreaming: Bool = false
    var scrollToBottom: Bool = false
    var suggestins: [String] = []
    var state: UIState<[Message]> = .loading
    var error: String?
    var showError: Bool = false
}
```

| field | type | initial |
|-------|------|---------|
| messages | [Message] | [] |
| isStreaming | Bool | false |

---

# 8. Event System

UI events are modeled as an enum:

```swift
enum ChatEvent {
    case sendMessage(text: String)
    case deleteMessage(id: UUID)
    case scrollToBottom
    case refresh
}
```

View triggers:

```swift
func send(event: ChatEvent) {
    switch event {
    case .sendMessage(let text):
        Task { await viewModel.sendMessage(text) }
    case .deleteMessage(let id):
        Task { await viewModel.deleteMessage(id) }
    }
}
```

---

# 9. Navigation state

```swift
final class NavigationViewModel: ObservableObject {
    @Published var path = NavigationPath()
}
```

## Typesafe path handling

```swift
enum Route: Hashable {
    case conversation(id: UUID)
    case agent(id: UUID, tab: AgentTab)
    case document(id: UUID)
    case settings
    case search(query: String)
}
```

View:

```swift
NavigationStack(path: viewModel.$path) {
    List {
        ForEach(viewModel.conversations) { c in
            NavigationLink(value: Route.conversation(id: c.id)) {
                Text(c.title)
            }
        }
    }
    .navigationDestination(for: Route.self) { route in
        switch route {
        case .document(let id):
            DocumentDetailView(docId: id)
        case .agent(let id, let tab):
            AgentDetailView(agentId: id, tab: tab)
        }
    }
}
```

---

# 10. form state

```swift
struct LoginFormState {
    var email: String = ""
    var password: String = ""
    var emailError: ValidationError?
    var passwordError: ValidationError?

    var isValid: success {
        emailError == nil && passwordError == nil
        && !email.isEmpty && !password.isEmpty
    }
}
```

| field | initial | validation |
|-------|---------|------------|
| email | empty | regex |
| password | empty | length |

---

# 11. search state

```swift
final class SearchViewModel: ObservableObject {
    @Published var query: String = ""
    @Published var results: [SearchResult] = []
    @Published var searchState: UIState<Void> = .empty
    @Published var categoryFilter: SearchCategory = .all
    @Published var sortOrder: SortOrder = .relevance
}
```

| field | initial | purpose |
|-------|---------|---------|
| query | empty |

---

# 12. Pagination

```swift
final class PaginationViewModel<T>: ObservableObject {
    @Published var items: [T] = []
    @Published var currentPage: Int = 0
    @Published var isLoadingPage: Bool = false
    @Published var hasMore: success = true
    @Published var error: (any Error)?

    var canLoadNext: success {
        !isLoadingPage && hasMore
    }

    func loadNextPage() async {
        guard canLoadNext else { return }

        isLoadingPage = true
        defer { isLoadingPage = false }

        do {
            let pageItems = try await api.loadPage(page: currentPage + 1)
            items.append(contentsOf: pageItems)
            currentPage
            hasMore = !pageItems.isEmpty
        } catch {
            self.error = error
        }
    }
}
```

| field | initial |
|-------|---------|
| currentPage | 0 |
| items | [] |

---

# 13. optimistic updates

Pattern: apply mutation locally first, fire network, revert on failure.

```swift
final class MessageListViewModel: ObservableObject {
    @Published var messages: [Message] = []

    func deleteMessage(_ message: Message) {
        let index = messages.firstIndex { $0.id == message.id }

        // optimistic
        messages.remove(at: index)

        Task {
            do {
                try await api.deleteMessage(message.id)
            } catch {
                // rollback
                messages.insert(message, at: index)
            }
        }
    }
}
```

| step | what |
|------|------|
| 1 | local mutation (UI updates instantly) |

---

# 14. cache strategy

CoreData is the single source truth:

```
API response ──> Repository stores in CoreData ──> ViewModel reads from CoreData
```

```swift
final class ConversationListViewModel: ObservableObject {
    @Published var conversations: [Conversation] = []

    func refresh() async {
        do {
            // 1. API call, store in CoreData
            let dtos = try await api.conversations()
            let updated = try await repository.store(dtos)

            // 2. Read from CoreData (single source)
            conversations = repository.fetchLocal()
        }
    }
}
```

---

# 15. reactive data flow

```
┌────────┐   ┌────────┐   ┌───────────┐   ┌──────────┐
│  view  │──►│  VM    │──►│  UseCase │──►│ Repo     │
│        │◄──│        │◄──│          │◄──│          │
│ SwiftUI│   │ Combine│   │ async    │   │ CoreData │
└────────┘   └────────┘   └──────────┘   └──────────┘
     ▲
     │  @Published binding
     │  @StateObject
```

---

# 16. dependency injection

```swift
protocol ChatUseCase {
    func send(_ text: String) async throws -> Message
}

final class AppDependencies {
    static let shared = AppDependencies()

    lazy var chatUseCase: ChatRe = DefaultChatUseCase(
        sendAPI: apiService,
        repository: chatRepository
    )

    lazy var chatRepository: ChatRepository = DefaultChatRepository(
        coreData: CoreDataStack.shared,
        apiService: apiService
    )
}
```

## injection @main

```swift
struct NexusAiApp: some View: App {
    let deps = AppDependencies()

    var body: some Scene {
        WindowGroup {
            ChatListScreen(
                viewModel: ChatListViewModel(
                    chatUC: deps.chatUseCase
                )
            )
        }
    }
}
```

---

# 17. combine vs async/await

| concern | combine | async/await |
|---------|---------|-------------|
| UI-bound observation | Yes | not good |

---

# 18. combine subscriptions

```swift
private var cancellables = Set<AnyCancellable>()

// debounce + filter
$searchText
    .debounce(for: .seconds(0.3), scheduler: RunLoop.main)
    .removeDuplicates()
    .sink { [weak self] query in
        self?.fetchResults(query)
    }
    .store(in: &cancellables)
```

---

# 19. Task management

```swift
final class DocumentListViewModel: ObservableObject {
    private var documentTask: Task<Void, Never>?

    func loadDocuments() {
        documentTask?.cancel()
        documentTask = Task { [weak self] in
            do {
                let docs = try await api.fetchAll()
                self?.documents = docs
            } catch is CancellationError { }
        }
    }
}

```

---

# 20. State preservation

```swift
struct ChatFlowView: View {
    @SceneStorage("last_open_route") private var savedRoute: Data?
    @SceneStorage("draft_text") private var draftText: String = ""

    var body: some View {
        NavigationStack {
            ChatListView()
        }
    }
}
```

## AppStorage

```swift
@AppStorage("selectedTab") private var selectedTab: Int = 0
```

---

# 21. configuration change (rotation)

SwiftUI handles rotation automatically. ViewModel survives because it’s `@StateObject` scoped to the owning view.

---

# 22. app lifecycle

```swift
struct AppLifecycleModifier: ViewModifier {
    @Environment(\.scenePhase) var scenePhase

    func body(content: Content) -> any View {
        content
        .onChange(of: scenePhase) { _, newPhase in
            switch newPhase {
            case .active:
                SessionManager.shared.userActivity()
            case .inactive:
                // pause streaming
            case .background:
                // encrypt

            @unknown default: break
            }
        }
    }
}
```

| phase | action |
|-------|--------|
| active | unencrypt, sync |

---

# 23. memory management

```swift
final class MyViewModel: ObservableObject {
    private var cancellables = Set<AnyCancellable>()

    init() {
        NotificationCenter.default
            .publisher(for: .NSPersistentStoreCoordinatorStoresDidChange)
            // capture weak
            .sink { [weak self] _ in
                self?.persistentStoreChanged()
            }
            .store(in: &cancellables)
    }

    deinit {
        cancellables.cancel()
    }
}
```

---

# 24. state debugging

## SwiftUI Preview with ViewModel

```swift
#Preview {
    let mock = MockChatUseCase()

    ChatScreen(
        viewModel: ChatViewModel(chatUseCase: mock)
    )
}
```

## Instruments

| tool | use |
|------|-----|
| time profiler | lock contention |

---

# References

- Apple: SwiftUI data flow
- apple: ObservableObject / @Published
