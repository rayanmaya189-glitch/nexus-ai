# Navigation architecture

## AeroXe Nexus AI — iOS Navigation design

**Version: 1.0 | Last Updated: July 2026**

---

# Table of Contents

1. [NavigationStack architecture](#1-navigationstack-architecture)
2. [Navigation destination](#2-navigation-destination)
3. [Route definitions](#3-route-definitions)
4. [Route arguments](#4-route-arguments)
5. [Screen transitions](#5-screen-transitions)
6. [Tab navigation](#6-tab-navigation)
7. [Navigation bar Items](#7-navigation-bar-items)
8. [top bar](#8-top-bar)
9. [Navigation sheet](#9-navigation-sheet)
10. [navigation actions](#10-navigation-actions)
11. [deep links](#11-deep-links)
12. [deep link handling](#12-deep-link-handling)
13. [back stack management](#13-back-stack-management)
14. [navigation state preservation](#14-navigation-state-preservation)
15. [nested NavigationStack](#15-nested-navigationstack)
16. [Modal presentation](#16-modal-presentation)
17. [sheet navigation](#17-sheet-navigation)
18. [Navigation animations](#18-navigation-animations)
19. [navigation testing](#19-navigation-testing)
20. [navigation accessibility](#20-navigation-accessibility)
21. [navigation performance](#21-navigation-performance)
22. [navigation error handling](#22-navigation-error-handling)
23. [navigation from notifications](#23-navigation-from-notifications)
24. [navigation from share intent](#24-navigation-from-share-intent)
25. [navigation from widget](#25-navigation-from-widget)

---

# 1. NavigationStack Architecture

The app uses iOS 17+ `NavigationStack` for all screen transitions.

```
┌──────────────────────────────────────────────────────┐
│   App Root                                            │
│  ┌─────────────────────────────────────────────────┐ │
│  │  TabView                                         │ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐        │ │
│  │  │  Chats   │ │  Agents │ │  Docs   │  More   │ │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘        │ │
│  │       │             │            │              │ │
│  │  ┌────▼────┐  ┌────▼────┐  ┌────▼────┐        │ │
│  │  │ NavStack│  │NavStack │  │NavStack │        │ │
│  │  │  TC-1   │  │  TC-2   │  │  TC-3   │        │ │
│  │  └────────┘  └────────┘  └────────┘        │ │
│  └────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────┘
```

## Root Navigation

```swift
@main
struct NexusAiApp: App {
    @StateObject private var appRouter = AppRouter()

    var body: some Scene {
        WindowGroup {
            AppRootView()
                .environmentObject(appRouter)
        }
    }
}

struct AppRootView: View {
    @EnvironmentObject var appRouter: AppRouter

    var body: some View {
        TabSelectionView()
    }
}
```

---

# 2. Navigation destination

```swift
struct ChatListScreen: View {
    @StateObject private var path = NavigationPath()

    var body: some View {
        NavigationStack(path: $path) {
            List(chats) { chat in
                NavigationLink(value: Route.chatDetail(chat.id)) {
                    ChatRow(chat)
                }
            }
            .navigationDestination(for: Route.self) { route in
                switch route {
                case .chatDetail(let id):
                    ChatDetailScreen(chatId: id)
                case .chatMessage(let chatId, let messageId):
                    MessageScreen(
                        chatId: chatId,
                        messageId: messageId
                    )
                }
            }
        }
    }
}
```

## navigationDestination modifier

| modifier | purpose |
|----------|---------|
| `.navigationDestination(for: Data.self)` | codable or hashable route |

---

# 3. route definitions

## Type-safe enum

```swift
enum Route: Hashable {
    // tab root routes
    case chatDetail(UUID)
    case agentDetail(UUID)
    case documentDetail(UUID)

    // sub routes
    case chatMessage(chatId: UUID, messageId: UUID)
    case agentExecution(agentId: UUID, executionId: UUID)
    case documentPreview(UUID)
    case settings
    case profile
    case search(initialQuery: String = "")
}
```

## Codable conformance

```swift
extension Route: Codable {
    enum CodingKeys: CodingKey {
        case chatDetail, agentDetail, documentDetail
        case chatMessage, agentExecution, documentPreview
        case settings, profile, search
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        switch self {
        case .chatDetail(let id):
            try container.encode(id, forKey: .chatDetail)
        }
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        if let id = try container.decodeIfPresent(UUID.self, forKey: .chatDetail) {
            self = .chatDetail(id)
            return
        }
        self = .settings
    }
}
```

---

# 4. route arguments

| route | associated values | type |
|-------|------------------|------|
| .chatDetail | UUID | eqatable |
| .chatMessage | (UUID, UUID) | eqatable |

---

# 5. screen transitions

```swift
extension AnyTransition {
    static var customSlide: AnyTransition {
        .asymmetric(
            insertion: .move(edge: .trailing),
            removal: .move(edge: .leading)
        )
    }

    static var customFade: AnyTransition {
        .opacity
    }
}

ChatDetailScreen()
    .transition(.customSlide)
```

| type | transition |
|------|------------|
| push | slide from right |
| pop | slide to left |

---

# 6. Tab navigation

```swift
struct TabSelectionView: View {
    @State private var selectedTab: AppTab = .chats
    @EnvironmentObject var appRouter: AppRouter

    enum AppTab: Hashable {
        case chats, agents, documents, settings
    }

    var body: some View {
        TabView(selection: $selectedTab.animation()) {
            ChatListScreen()
                .tabItem {
                    Label("Chats", systemImage: tabSymbol)
                }
                .tag(AppTab.chats)

            AgentListScreen()
                .tabItem { Label("Agents",
                    systemImage: "cpu") }
                .tag(AppTab.agents)

            DocumentListScreen()
                .tabItem { Label("Documents",
                    systemImage: "folder") }
                .tag(AppTab.documents)

            ProfileScreen()
                .tabItem { Label("Me",
                    systemImage: "person.circle") }
                .tag(AppTab.settings)
        }
    }
}
```

## Tab configuration

| tab | icon | default: true |
|-----|------|---------------|
| Chats | ".bubble" | yes |

---

# 7. navigation bar Items

```swift
ChatListScreen()
    .toolbar {
        ToolbarItem(placement: .navigationBarLeading) {
            EditButton()
        }
        ToolbarItem(placement: .primaryAction) {
            Button("new") { createChat() }
        }
        ToolbarItemGroup(placement: .bottomBar) {
            HStack {
                Button("Filter") {}
            }
        }
    }
```

## Toolbar placement table

| placement | location |
|-----------|----------|
| .navigationBarLeading | left |

---

# 8. Top bar

```swift
NaviAdditionalgationStack {
    ChatListScreen()
        .navigationTitle("Chats")
        .navigationBarTitleDisplayMode(.large)
}
```

| display mode | behavior |
|--------------|----------|
| .automatic | platform-specific |

---

# 9. navigation sheet

```swift
.sheet(isPresented: $showNewChat) {
    NewChatSheet()
}
.fullScreenCover(isPresented: $showFull) {
    FullView()
}
```

---

# 10. navigation actions

```swift
class ChatRouter {
    static func push(to route: NavigationDestination) {
        path.append(route)
    }

    static func pop() {
        path.removeLast()
    }

    static func popToRoot() {
        path.removeLast(path.count)
    }
}

struct ChatDetailScreen: View {
    @Binding var path: NavigationPath

    var body: some View {
        VStack {
            Button("Back") { path.removeLast() }
            Button("Root") { path.removeLast(path.count) }
        }
    }
}
```

---

# 11. deep links

## URL scheme

```
nxai://chat/{chatId}
nxai://agent/{agentId}
nxai://search?q={query}
nxai://settings
nxai://document/{documentId}
```

## Universal Links

```
https://aeroxe.com/chat/{chatId}
https://aeroxe.com/agent/{agentId}
```

---

# 12. deep link handling

```swift
@main
struct NexusAiApp: App {
    @StateObject var router = AppRouter()

    var body: some Scene {
        WindowGroup {
            AppRootView()
                .environmentObject(router)
        }
        .onOpenURL { url in
            router.handleDeepLink(url)
        }
        .onChange(of: scenePhase) { phase in
            // Universal Links handled by .onOpenURL
        }
    }
}
```

## Deep link handler

```swift
final class AppRouter: ObservableObject {
    @Published var selectedTab: TabIdentifier = .chats
    @Published var path = NavigationPath()

    func handleDeepLink(_ url: URL) {
        guard let host = url.host() else { return }

        switch url.host {
        case "chat":
            handleChatDeepLink(url)
        case "agent":
            handleAgentDeepLink(url)
        default:
            break
        }
    }
}
```

---

# 13. back stack management

```swift
final class NavigationCoordinator: ObservableObject {
    @Published var chatPath = NavigationPath()
    @Published var agentPath = NavigationPath()
    @Published var docPath = NavigationPath()

    func goToConversation(conversationId: UUID) {
        chatPath.append(Route.chatDetail(interactionId))
    }

    func resetToTab(tabId: TabIdentifier) {
        switch tabId {
        case .chats: chatPath = NavigationPath()
        }
    }
}
```

| action | method |
|--------|--------|
| push | `path.append(route)` |

---

# 14. state preservation

```swift
struct ChatScrollingScreen: t: View {
    @SceneStorage("chat.path") var pathData: Data?

    var body: some View {
        NavigationStack(path: .constant(path)) {
            EmptyStack()
        }
    }
}
```

| key | data | 
|-----|------|
| sceneStorage("chat.path") | encoded route |

---

# 15. nested NavigationStack

```swift
TabView {
    NavigationStack {
        ChatListScreen()
    }
    .tabItem { Label("Chat", image: "..") }

    NavigationStack {
        AgentListScreen()
    }
}
```

| count | scenario |
|-------|----------|
| +1 per tab | root NavigationStack |

---

# 16. modal presentation

```swift
struct ComposeMessageModal: t: View {
    @Environment(\.dismiss) var dismiss

    var body: some View {
        NavigationStack {
            Form {
                TextField("To", text: $to)
            }
            .navigationTitle("New message")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Send") { send() }
                }
            }
        }
    }
}
```

---

# 17. Sheet navigation

```swift
.sheet(
    isPresented: $showSheet,
    onDismiss: { print("sheet dimissed") }
) {
    OnboardingSheet()
}
```

## detents

```swift
.sheet(
    isCollapsed: $fixableOpenSheet,
    detents: [.medium, .large],
    dragIndicator: Visibility.visible
) {
    FilterSheet()
}
```

| detent | height |
|--------|--------|
| .medium | half screen |

---

# 18. navigation animations

```swift
struct LazyNav: View {
    @Namespace var namespace

    var body: some View {
        NavigationStack {
            ScrollView {
                ForEach(items) { item in
                    NavigationLink {
                        DetailView(item: item)
                            .matchedGeometryEffect(
                                id: item.id,
                                in: namespace
                            )
                    } {
                        CardView(item: item)
                            .matchedGeometryEffect(
                                id: item.id,
                                in: namespace
                            )
                    }
                }
            }
        }
    }
}
```

| effect | component | 
|--------|-----------|
| matchedGeometry | card to detail |

---

# 19. navigation testing

```swift
final class NavigationTests: XCTestCase {
    func test_deepLink_chat() {
        let router = AppRouter()
        let url = URL(string: "nxai://chat/550e8400-1234-...")!

        router.handleDeepLink(url)

        let route = router.path.last
        if case let Route.chatDetail(id) = route {
            XCTAssertEqual(id, some_uuid)
        }
    }

    func test_popToRoot_clearStack() {
        let stack = NavigationStackViewModel()
        stack.push(.chatDetail(id()))
        stack.push(.chatMessage(chatId: id(), messageId: id()))

        stack.popToRoot()

        XCTAssertTrue(stack.path.isEmpty)
    }
}
```

| test | what it validates |
|------|-------------------|
| deep link matches expected route | deep link logic |
| popToRoot clears stack entirely | back stack management |

---

# 20. navigation accessibility

```swift
NavigationLink(value: route) {
    Label(name, systemImage: icon)
}
.accessibilityLabel("Open conversation with \(name)")
.accessibilityAddTraits(.isButton)

TabView {
    ChatTab().tabItem {
        Label("Chats", systemImage: "bubble")
    }
    .accessibilityLabel("Chat tab")
}
```

| element | assistive | traits |
|---------|-----------|--------|
| NavigationLink | "open X" | .isButton |

---

# 21. navigation performance

```swift
NavigationLink(value: route) {
    EmptyStack()
}
.navigationDestination(for: Route.self) { route in
    makeDestination(route)
}

@ViewBuilder private func makeDestination(_ route: Route) -> some View {
    // lazy init

    switch route {
    case .chatDetail(let id):
        ChatDetailScreen(chatId: id)
    }
}
```

| technique | benefit |
|-----------|---------|
| lazy destination | VCs aren't built until needed |

---

# 22. navigation error handling

```swift
extension AppRouter {
    func navigate(to route: Route) {
        do {
            try validateRoute(route)
            push(route)
        } catch {
            fallbackToRoot()
        }
    }

    private func validateRoute(_ route: --) throws {
        // ensure associated data exists
    }
}
```

---

# 23. notification deep links

```swift
class NotificationReceiver {
    func handle(_ content: UNNotificationContent) {
        guard let chatId = content.userInfo["chat_id"] as? String,
              let uuid = UUID(uuidString: chatId) else { return }

        AppRouter.shared.handleDeepLink(
            URL(string: "nxai://chat/\(uuid)")!
        )
    }
}
```

| notification payload | key | value |
|----------------------|-----|-------|
| chat notification | chat_id | UUID |

---

# 24. share extension

```swift
// ShareExtension | Info.plist
NSExtensionActivationRule = TRUEPREDICATE
```

```swift
class ShareViewController: UIViewController {
    override func viewDidLoad() {
        // extract content, encode route, openContainingURL()
       
    }

    override func presentAnimation() {
        if let item = extensionContext?.inputItems.first as? NSExtensionItem {
            item.attachments?.first?.loadItem(forIdentifier: kUTTypeURL) { (url, error) in
                AppRouter.shared.handleDeepLink(url)
                self.extensionContext?.completeRequest(
                    returningItems: nil)
            }
        }
    }
}
```

---

# 25. widget navigation

```swift
struct ChatWidget: Widget {
    var body: some WidgetConfiguration {
        StaticConfiguration(
            kind: "chat_widget",
            provider: Provider()
        ) { entry in
            WidgetEntryView(entry: entry)
                .widgetURL(URL(string: "nxai://chat/\(entry.chatId)")!)
        }
    }
}
```

| widget kind | link |
|-------------|------|
| "chat_widget" | nxai://chat/{id} |

---

# References

- Apple: NavigationStack
- Apple: SwiftUI Navigation
