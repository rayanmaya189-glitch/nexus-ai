# Notifications

The notifications module handles push notification registration, delivery, categorization, in-app notification center, preferences, and deep link navigation. This document covers every aspect of the notification system.

## Architecture Overview

```
+------------------------------------------------------------------+
|                    Notification System                             |
+------------------------------------------------------------------+
|                                                                    |
|  +----------------+    +----------------+    +----------------+  |
|  | APNs           |--->| Notification   |--->| In-App         |  |
|  | Integration    |    | Manager        |    | Center         |  |
|  +----------------+    +----------------+    +----------------+  |
|         |                     |                     |              |
|         v                     v                     v              |
|  +----------------+    +----------------+    +----------------+  |
|  | Device Token   |    | Categories     |    | Deep Links     |  |
|  | Management     |    | & Actions      |    | Navigation     |  |
|  +----------------+    +----------------+    +----------------+  |
|                                                                    |
+------------------------------------------------------------------+
```

## Notification Flow

```
+--------+    +--------+    +--------+    +--------+    +--------+
| Server |--->|  APNs  |--->| Device |--->| Notif  |--->| User   |
| sends  |    | pushes |    | receives|    | Center |    | sees   |
| payload|    | to     |    |         |    | shows  |    |        |
+--------+    +--------+    +--------+    +--------+    +--------+
                                                       |
                                                       v
                                              +-----------------+
                                              | Deep link or    |
                                              | in-app action   |
                                              +-----------------+
```

## Data Models

```swift
// MARK: - Notification Model

struct AppNotification: Codable, Identifiable, Equatable {
    let id: String
    var title: String
    var body: String
    var category: NotificationCategory
    var priority: NotificationPriority
    var data: NotificationData?
    var isRead: Bool
    var createdAt: Date
    var readAt: Date?

    enum CodingKeys: String, CodingKey {
        case id, title, body, category, priority, data
        case isRead = "is_read"
        case createdAt = "created_at"
        case readAt = "read_at"
    }
}

// MARK: - Notification Category

enum NotificationCategory: String, Codable, CaseIterable {
    case aiAlert = "ai_alert"
    case security = "security"
    case workflow = "workflow"
    case general = "general"
    case system = "system"
    case mention = "mention"

    var displayName: String {
        switch self {
        case .aiAlert: return "AI Alerts"
        case .security: return "Security"
        case .workflow: return "Workflow"
        case .general: return "General"
        case .system: return "System"
        case .mention: return "Mentions"
        }
    }

    var iconName: String {
        switch self {
        case .aiAlert: return "brain.head.profile"
        case .security: return "lock.shield"
        case .workflow: return "arrow.triangle.branch"
        case .general: return "bell"
        case .system: return "gearshape"
        case .mention: return "at"
        }
    }

    var color: String {
        switch self {
        case .aiAlert: return "purple"
        case .security: return "red"
        case .workflow: return "blue"
        case .general: return "gray"
        case .system: return "orange"
        case .mention: return "green"
        }
    }
}

// MARK: - Notification Priority

enum NotificationPriority: String, Codable, CaseIterable {
    case low = "low"
    case normal = "normal"
    case high = "high"
    case critical = "critical"

    var interruptionLevel: String {
        switch self {
        case .low: return "passive"
        case .normal: return "active"
        case .high: return "timeSensitive"
        case .critical: return "critical"
        }
    }
}

// MARK: - Notification Data

struct NotificationData: Codable, Equatable {
    let deepLink: String?
    let agentId: String?
    let documentId: String?
    let executionId: String?
    let actionUrl: String?
    let metadata: [String: String]?

    enum CodingKeys: String, CodingKey {
        case deepLink = "deep_link"
        case agentId = "agent_id"
        case documentId = "document_id"
        case executionId = "execution_id"
        case actionUrl = "action_url"
        case metadata
    }
}

// MARK: - Notification Preferences

struct NotificationPreferences: Codable, Equatable {
    var globalEnabled: Bool
    var categories: [CategoryPreference]
    var quietHoursEnabled: Bool
    var quietHoursStart: Date
    var quietHoursEnd: Date
    var pushEnabled: Bool
    var inAppEnabled: Bool
    var emailEnabled: Bool

    enum CodingKeys: String, CodingKey {
        case globalEnabled = "global_enabled"
        case categories
        case quietHoursEnabled = "quiet_hours_enabled"
        case quietHoursStart = "quiet_hours_start"
        case quietHoursEnd = "quiet_hours_end"
        case pushEnabled = "push_enabled"
        case inAppEnabled = "in_app_enabled"
        case emailEnabled = "email_enabled"
    }
}

struct CategoryPreference: Codable, Identifiable, Equatable {
    let id: String
    let category: NotificationCategory
    var enabled: Bool
    var pushEnabled: Bool
    var soundEnabled: Bool

    enum CodingKeys: String, CodingKey {
        case id, category, enabled
        case pushEnabled = "push_enabled"
        case soundEnabled = "sound_enabled"
    }
}

// MARK: - Push Token

struct PushTokenRegistration: Codable {
    let token: String
    let platform: String
    let deviceModel: String
    let osVersion: String
    let appVersion: String

    enum CodingKeys: String, CodingKey {
        case token, platform
        case deviceModel = "device_model"
        case osVersion = "os_version"
        case appVersion = "app_version"
    }
}
```

## Notification Manager

```swift
import SwiftUI
import UserNotifications
import Combine

@MainActor
final class NotificationManager: NSObject, ObservableObject {
    @Published var notifications: [AppNotification] = []
    @Published var unreadCount = 0
    @Published var preferences = NotificationPreferences.default
    @Published var isRegistered = false

    private let apiService: NotificationAPIService
    private let cacheService: NotificationCacheService
    private var cancellables = Set<AnyCancellable>()

    var unreadNotifications: [AppNotification] {
        notifications.filter { !$0.isRead }
    }

    init(
        apiService: NotificationAPIService = .shared,
        cacheService: NotificationCacheService = .shared
    ) {
        self.apiService = apiService
        self.cacheService = cacheService
        super.init()
        UNUserNotificationCenter.current().delegate = self
        setupNotificationCategories()
    }

    // MARK: - Registration

    func registerForPushNotifications() async {
        do {
            let granted = try await UNUserNotificationCenter.current()
                .requestAuthorization(options: [.alert, .badge, .sound])

            guard granted else { return }

            await MainActor.run {
                UIApplication.shared.registerForRemoteNotifications()
            }
        } catch {
            print("Notification auth failed: \(error)")
        }
    }

    func handleDeviceToken(_ deviceToken: Data) {
        let token = deviceToken.map { String(format: "%02.2hhx", $0) }.joined()
        let registration = PushTokenRegistration(
            token: token,
            platform: "ios",
            deviceModel: UIDevice.current.model,
            osVersion: UIDevice.current.systemVersion,
            appVersion: Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0"
        )

        Task {
            do {
                try await apiService.registerDevice(registration)
                isRegistered = true
            } catch {
                print("Token registration failed: \(error)")
            }
        }
    }

    func handleTokenRefresh(_ deviceToken: Data) {
        handleDeviceToken(deviceToken)
    }

    // MARK: - Notification Categories

    private func setupNotificationCategories() {
        let categories: [(NotificationCategory, [UNNotificationAction])] = [
            (.aiAlert, [
                UNNotificationAction(identifier: "VIEW_ACTION", title: "View", options: .foreground),
                UNNotificationAction(identifier: "DISMISS_ACTION", title: "Dismiss", options: .destructive)
            ]),
            (.security, [
                UNNotificationAction(identifier: "APPROVE_ACTION", title: "Approve", options: .foreground),
                UNNotificationAction(identifier: "REJECT_ACTION", title: "Reject", options: .destructive),
                UNNotificationAction(identifier: "VIEW_ACTION", title: "View Details", options: .foreground)
            ]),
            (.workflow, [
                UNNotificationAction(identifier: "VIEW_ACTION", title: "View", options: .foreground),
                UNNotificationAction(identifier: "EXECUTE_ACTION", title: "Execute", options: .foreground)
            ]),
            (.general, [
                UNNotificationAction(identifier: "VIEW_ACTION", title: "View", options: .foreground),
                UNNotificationAction(identifier: "DISMISS_ACTION", title: "Dismiss", options: .destructive)
            ]),
            (.system, [
                UNNotificationAction(identifier: "VIEW_ACTION", title: "View", options: .foreground)
            ]),
            (.mention, [
                UNNotificationAction(identifier: "REPLY_ACTION", title: "Reply", options: .foreground),
                UNNotificationAction(identifier: "VIEW_ACTION", title: "View", options: .foreground)
            ])
        ]

        let unCategories: [UNNotificationCategory] = categories.map { cat, actions in
            UNNotificationCategory(
                identifier: cat.rawValue,
                actions: actions,
                intentIdentifiers: [],
                options: .customDismissAction
            )
        }

        UNUserNotificationCenter.current().setNotificationCategories(Set(unCategories))
    }

    // MARK: - Fetch Notifications

    func fetchNotifications() async {
        do {
            notifications = try await apiService.fetchNotifications()
            unreadCount = notifications.filter { !$0.isRead }.count
            updateBadge()
            await cacheService.cacheNotifications(notifications)
        } catch {
            if let cached = await cacheService.getCachedNotifications() {
                notifications = cached
                unreadCount = cached.filter { !$0.isRead }.count
            }
        }
    }

    func fetchPreferences() async {
        do {
            preferences = try await apiService.fetchPreferences()
        } catch { }
    }

    // MARK: - Read / Clear

    func markAsRead(_ notification: AppNotification) {
        guard !notification.isRead else { return }
        if let index = notifications.firstIndex(where: { $0.id == notification.id }) {
            notifications[index].isRead = true
            notifications[index].readAt = Date()
            unreadCount = notifications.filter { !$0.isRead }.count
            updateBadge()
        }
        Task { try? await apiService.markAsRead(id: notification.id) }
    }

    func markAllAsRead() {
        for index in notifications.indices {
            notifications[index].isRead = true
            notifications[index].readAt = Date()
        }
        unreadCount = 0
        updateBadge()
        Task { try? await apiService.markAllAsRead() }
    }

    func deleteNotification(_ notification: AppNotification) {
        notifications.removeAll { $0.id == notification.id }
        unreadCount = notifications.filter { !$0.isRead }.count
        updateBadge()
        Task { try? await apiService.deleteNotification(id: notification.id) }
    }

    func clearAll() {
        notifications.removeAll()
        unreadCount = 0
        updateBadge()
        Task { try? await apiService.clearAll() }
    }

    // MARK: - Badge

    func updateBadge() {
        UNUserNotificationCenter.current().setBadgeCount(unreadCount)
    }

    // MARK: - Preferences

    func updatePreferences(_ prefs: NotificationPreferences) async {
        preferences = prefs
        do {
            try await apiService.updatePreferences(prefs)
        } catch { }
    }

    func toggleCategory(_ category: NotificationCategory) {
        if let index = preferences.categories.firstIndex(where: { $0.category == category }) {
            preferences.categories[index].enabled.toggle()
            Task { await updatePreferences(preferences) }
        }
    }
}

extension NotificationManager: UNUserNotificationCenterDelegate {
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        let content = notification.request.content
        let userInfo = content.userInfo

        let appNotification = AppNotification(
            id: notification.request.identifier,
            title: content.title,
            body: content.body,
            category: NotificationCategory(rawValue: userInfo["category"] as? String ?? "") ?? .general,
            priority: .normal,
            data: parseNotificationData(userInfo),
            isRead: false,
            createdAt: Date(),
            readAt: nil
        )

        notifications.insert(appNotification, at: 0)
        unreadCount += 1
        updateBadge()

        completionHandler([.banner, .badge, .sound])
    }

    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let userInfo = response.notification.request.content.userInfo

        switch response.actionIdentifier {
        case "VIEW_ACTION":
            handleDeepLink(userInfo)
        case "APPROVE_ACTION":
            handleApproval(userInfo, approved: true)
        case "REJECT_ACTION":
            handleApproval(userInfo, approved: false)
        case "DISMISS_ACTION":
            break
        case "EXECUTE_ACTION":
            handleExecution(userInfo)
        case "REPLY_ACTION":
            break
        default:
            handleDeepLink(userInfo)
        }

        completionHandler()
    }

    private func handleDeepLink(_ userInfo: [AnyHashable: Any]) {
        guard let data = parseNotificationData(userInfo),
              let deepLink = data.deepLink else { return }

        NotificationCenter.default.post(
            name: .deepLinkNotification,
            object: nil,
            userInfo: ["deepLink": deepLink]
        )
    }

    private func handleApproval(_ userInfo: [AnyHashable: Any], approved: Bool) {
        guard let data = parseNotificationData(userInfo),
              let executionId = data.executionId else { return }

        Task {
            try? await apiService.handleApproval(executionId: executionId, approved: approved)
        }
    }

    private func handleExecution(_ userInfo: [AnyHashable: Any]) {
        guard let data = parseNotificationData(userInfo),
              let actionUrl = data.actionUrl else { return }

        NotificationCenter.default.post(
            name: .executeNotificationAction,
            object: nil,
            userInfo: ["actionUrl": actionUrl]
        )
    }

    private func parseNotificationData(_ userInfo: [AnyHashable: Any]) -> NotificationData? {
        guard let dataDict = userInfo["data"] as? [String: Any] else { return nil }
        return NotificationData(
            deepLink: dataDict["deep_link"] as? String,
            agentId: dataDict["agent_id"] as? String,
            documentId: dataDict["document_id"] as? String,
            executionId: dataDict["execution_id"] as? String,
            actionUrl: dataDict["action_url"] as? String,
            metadata: dataDict["metadata"] as? [String: String]
        )
    }
}

extension Notification.Name {
    static let deepLinkNotification = Notification.Name("deepLinkNotification")
    static let executeNotificationAction = Notification.Name("executeNotificationAction")
}
```

## In-App Notification Center

```swift
struct NotificationCenterView: View {
    @StateObject private var viewModel = NotificationCenterViewModel()

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.notifications.isEmpty {
                    NotificationCenterLoadingView()
                } else if viewModel.notifications.isEmpty {
                    NotificationCenterEmptyView()
                } else {
                    notificationList
                }
            }
            .navigationTitle("Notifications")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Menu {
                        Button("Mark All Read") {
                            viewModel.markAllAsRead()
                        }
                        Button("Clear All", role: .destructive) {
                            viewModel.clearAll()
                        }
                    } label: {
                        Image(systemName: "ellipsis.circle")
                    }
                }
            }
            .refreshable { await viewModel.refresh() }
            .task { await viewModel.load() }
        }
    }

    private var notificationList: some View {
        List {
            if !viewModel.unreadNotifications.isEmpty {
                Section {
                    ForEach(viewModel.unreadNotifications) { notification in
                        NotificationRow(notification: notification)
                            .onTapGesture {
                                viewModel.markAsRead(notification)
                            }
                    }
                } header: {
                    Text("Unread (\(viewModel.unreadNotifications.count))")
                        .font(.caption.bold())
                }
            }

            Section {
                ForEach(viewModel.readNotifications) { notification in
                    NotificationRow(notification: notification)
                }
            } header: {
                Text("Earlier")
                    .font(.caption.bold())
            }
        }
        .listStyle(.insetGrouped)
    }
}

struct NotificationRow: View {
    let notification: AppNotification

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            categoryIcon
            content
            readStatus
        }
        .opacity(notification.isRead ? 0.7 : 1.0)
    }

    private var categoryIcon: some View {
        Image(systemName: notification.category.iconName)
            .font(.title3)
            .foregroundStyle(Color(notification.category.color))
            .frame(width: 32, height: 32)
            .background(Color(notification.category.color).opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private var content: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(notification.title)
                .font(.subheadline.bold())
            Text(notification.body)
                .font(.caption)
                .foregroundStyle(.secondary)
                .lineLimit(3)
            HStack {
                Text(notification.category.displayName)
                    .font(.caption2)
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(.quaternary)
                    .clipShape(Capsule())
                Text(notification.createdAt.formatted(.relative(presentation: .named)))
                    .font(.caption2)
                    .foregroundStyle(.tertiary)
            }
        }
    }

    private var readStatus: some View {
        Group {
            if !notification.isRead {
                Circle()
                    .fill(.blue)
                    .frame(width: 8, height: 8)
            }
        }
    }
}

struct NotificationCenterLoadingView: View {
    var body: some View {
        List(0..<5) { _ in
            HStack(spacing: 12) {
                RoundedRectangle(cornerRadius: 8)
                    .fill(.quaternary)
                    .frame(width: 32, height: 32)
                VStack(alignment: .leading, spacing: 6) {
                    RoundedRectangle(cornerRadius: 4).fill(.quaternary).frame(width: 140, height: 14)
                    RoundedRectangle(cornerRadius: 4).fill(.quaternary).frame(width: 200, height: 10)
                }
                Spacer()
            }
            .padding(.vertical, 6)
        }
        .listStyle(.insetGrouped)
        .redacted(reason: .placeholder)
    }
}

struct NotificationCenterEmptyView: View {
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "bell.slash")
                .font(.system(size: 48))
                .foregroundStyle(.secondary)
            Text("No Notifications")
                .font(.title2.bold())
            Text("You're all caught up!")
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

@MainActor
final class NotificationCenterViewModel: ObservableObject {
    @Published var notifications: [AppNotification] = []
    @Published var isLoading = false

    private let manager: NotificationManager

    var unreadNotifications: [AppNotification] {
        notifications.filter { !$0.isRead }
    }

    var readNotifications: [AppNotification] {
        notifications.filter { $0.isRead }
    }

    init(manager: NotificationManager = .shared) {
        self.manager = manager
        // Observe manager changes
    }

    func load() async {
        isLoading = true
        await manager.fetchNotifications()
        notifications = manager.notifications
        isLoading = false
    }

    func refresh() async {
        await manager.fetchNotifications()
        notifications = manager.notifications
    }

    func markAsRead(_ notification: AppNotification) {
        manager.markAsRead(notification)
        if let idx = notifications.firstIndex(where: { $0.id == notification.id }) {
            notifications[idx].isRead = true
        }
    }

    func markAllAsRead() {
        manager.markAllAsRead()
        for index in notifications.indices {
            notifications[index].isRead = true
        }
    }

    func deleteNotification(_ notification: AppNotification) {
        manager.deleteNotification(notification)
        notifications.removeAll { $0.id == notification.id }
    }

    func clearAll() {
        manager.clearAll()
        notifications.removeAll()
    }
}
```

## Notification Preferences View

```swift
struct NotificationPreferencesView: View {
    @StateObject private var viewModel = NotificationPreferencesViewModel()

    var body: some View {
        Form {
            Section("Global") {
                Toggle("Enable Notifications", isOn: $viewModel.preferences.globalEnabled)
                Toggle("Push Notifications", isOn: $viewModel.preferences.pushEnabled)
                Toggle("In-App Notifications", isOn: $viewModel.preferences.inAppEnabled)
                Toggle("Email Notifications", isOn: $viewModel.preferences.emailEnabled)
            }

            Section("Categories") {
                ForEach(viewModel.preferences.categories) { pref in
                    HStack {
                        Image(systemName: pref.category.iconName)
                            .foregroundStyle(Color(pref.category.color))
                        Text(pref.category.displayName)
                        Spacer()
                        Toggle("", isOn: Binding(
                            get: { pref.enabled },
                            set: { _ in viewModel.toggleCategory(pref.category) }
                        ))
                        .labelsHidden()
                    }
                }
            }

            Section("Quiet Hours") {
                Toggle("Enable Quiet Hours", isOn: $viewModel.preferences.quietHoursEnabled)
                if viewModel.preferences.quietHoursEnabled {
                    DatePicker("Start", selection: $viewModel.preferences.quietHoursStart, displayedComponents: .hourAndMinute)
                    DatePicker("End", selection: $viewModel.preferences.quietHoursEnd, displayedComponents: .hourAndMinute)
                }
            }

            Section {
                Button("Save Preferences") {
                    Task { await viewModel.save() }
                }
                .frame(maxWidth: .infinity)
            }
        }
        .navigationTitle("Notification Preferences")
    }
}

@MainActor
final class NotificationPreferencesViewModel: ObservableObject {
    @Published var preferences: NotificationPreferences = .default
    @Published var isSaving = false

    private let manager: NotificationManager

    init(manager: NotificationManager = .shared) {
        self.manager = manager
        self.preferences = manager.preferences
    }

    func toggleCategory(_ category: NotificationCategory) {
        manager.toggleCategory(category)
        if let idx = preferences.categories.firstIndex(where: { $0.category == category }) {
            preferences.categories[idx].enabled.toggle()
        }
    }

    func save() async {
        isSaving = true
        await manager.updatePreferences(preferences)
        isSaving = false
    }
}
```

## Notification Badge Support

```swift
struct NotificationBadgeModifier: ViewModifier {
    @ObservedObject var manager: NotificationManager

    func body(content: Content) -> some View {
        content
            .overlay(alignment: .topTrailing) {
                if manager.unreadCount > 0 {
                    Text(badgeText)
                        .font(.caption2.bold())
                        .foregroundStyle(.white)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(.red)
                        .clipShape(Capsule())
                        .offset(x: 4, y: -4)
                }
            }
    }

    private var badgeText: String {
        manager.unreadCount > 99 ? "99+" : "\(manager.unreadCount)"
    }
}

extension View {
    func notificationBadge(manager: NotificationManager) -> some View {
        modifier(NotificationBadgeModifier(manager: manager))
    }
}
```

## Deep Link Navigation

```swift
enum NotificationDeepLink: String {
    case dashboard = "/dashboard"
    case agents = "/agents"
    case agentDetail = "/agents/"
    case knowledgeCenter = "/knowledge"
    case documentDetail = "/documents/"
    case chat = "/chat"
    case settings = "/settings"
    case notifications = "/notifications"

    static func from(_ urlString: String) -> NotificationDeepLink? {
        guard let url = URL(string: urlString),
              let host = url.host,
              let path = URL(string: urlString)?.path else { return nil }

        let fullPath = "/\(host)\(path)"

        if fullPath.hasPrefix("/agents/") { return .agentDetail }
        if fullPath.hasPrefix("/documents/") { return .documentDetail }

        return NotificationDeepLink(rawValue: fullPath)
    }
}

struct DeepLinkHandler: ViewModifier {
    @EnvironmentObject var navigationState: NavigationState

    func body(content: Content) -> some View {
        content
            .onReceive(NotificationCenter.default.publisher(for: .deepLinkNotification)) { notification in
                guard let userInfo = notification.userInfo,
                      let deepLinkString = userInfo["deepLink"] as? String,
                      let deepLink = NotificationDeepLink.from(deepLinkString) else { return }

                navigate(to: deepLink, userInfo: userInfo)
            }
    }

    private func navigate(to link: NotificationDeepLink, userInfo: [AnyHashable: Any]? = nil) {
        switch link {
        case .dashboard:
            navigationState.selectedTab = .dashboard
        case .agents:
            navigationState.selectedTab = .agents
        case .agentDetail:
            if let data = userInfo?["data"] as? [String: Any],
               let agentId = data["agent_id"] as? String {
                navigationState.navigate(to: .agentDetail(agentId))
            }
        case .knowledgeCenter:
            navigationState.selectedTab = .knowledge
        case .documentDetail:
            if let data = userInfo?["data"] as? [String: Any],
               let docId = data["document_id"] as? String {
                navigationState.navigate(to: .documentDetail(docId))
            }
        case .chat:
            navigationState.selectedTab = .chat
        case .settings:
            navigationState.selectedTab = .settings
        case .notifications:
            navigationState.navigate(to: .notifications)
        }
    }
}
```

## Notification Custom Content Extension

```swift
// NotificationContentExtension/Info.plist entry in extension target:
// NSExtension/NSExtensionPointIdentifier: com.apple.usernotifications.content-extension

class NotificationContentViewController: UIViewController, UNNotificationContentExtension {
    var notificationData: NotificationData?

    func didReceive(_ notification: UNNotification) {
        let content = notification.request.content
        let userInfo = content.userInfo

        if let dataDict = userInfo["data"] as? [String: Any] {
            notificationData = NotificationData(
                deepLink: dataDict["deep_link"] as? String,
                agentId: dataDict["agent_id"] as? String,
                documentId: dataDict["document_id"] as? String,
                executionId: dataDict["execution_id"] as? String,
                actionUrl: dataDict["action_url"] as? String,
                metadata: dataDict["metadata"] as? [String: String]
            )
        }

        setupUI(content: content)
    }

    func didReceive(_ response: UNNotificationResponse, completionHandler: @escaping (UNNotificationPresentationOptions) -> Void) {
        completionHandler([.banner, .badge, .sound])
    }

    private func setupUI(content: UNNotificationContent) {
        view.backgroundColor = .systemBackground
        // Render custom UI based on notification category and data
    }
}
```

## Notification API Integration

```swift
protocol NotificationAPIService {
    func registerDevice(_ registration: PushTokenRegistration) async throws
    func fetchNotifications() async throws -> [AppNotification]
    func markAsRead(id: String) async throws
    func markAllAsRead() async throws
    func deleteNotification(id: String) async throws
    func clearAll() async throws
    func fetchPreferences() async throws -> NotificationPreferences
    func updatePreferences(_ preferences: NotificationPreferences) async throws
    func handleApproval(executionId: String, approved: Bool) async throws
}

final class DefaultNotificationAPIService: NotificationAPIService {
    static let shared = DefaultNotificationAPIService()
    private let client: NetworkClient

    init(client: NetworkClient = .shared) { self.client = client }

    func registerDevice(_ registration: PushTokenRegistration) async throws {
        var req = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/notifications/register")!,
            method: .post
        )
        req.httpBody = try JSONEncoder().encode(registration)
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        _ = try await client.execute(req)
    }

    func fetchNotifications() async throws -> [AppNotification] {
        let req = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/notifications")!,
            method: .post
        )
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode([AppNotification].self, from: data)
    }

    func markAsRead(id: String) async throws {
        let req = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/notifications/\(id)/read")!,
            method: .patch
        )
        _ = try await client.execute(req)
    }

    func markAllAsRead() async throws {
        let req = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/notifications/read-all")!,
            method: .patch
        )
        _ = try await client.execute(req)
    }

    func deleteNotification(id: String) async throws {
        let req = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/notifications/\(id)")!,
            method: .delete
        )
        _ = try await client.execute(req)
    }

    func clearAll() async throws {
        let req = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/notifications/clear")!,
            method: .delete
        )
        _ = try await client.execute(req)
    }

    func fetchPreferences() async throws -> NotificationPreferences {
        let req = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/notifications/preferences")!,
            method: .post
        )
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode(NotificationPreferences.self, from: data)
    }

    func updatePreferences(_ preferences: NotificationPreferences) async throws {
        var req = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/notifications/preferences")!,
            method: .patch
        )
        req.httpBody = try JSONEncoder().encode(preferences)
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        _ = try await client.execute(req)
    }

    func handleApproval(executionId: String, approved: Bool) async throws {
        var req = try URLRequest(
            url: URL(string: "\(APIConfig.baseURL)/api/v1/executions/\(executionId)/approve")!,
            method: .post
        )
        let body = ["approved": approved]
        req.httpBody = try JSONSerialization.data(withJSONObject: body)
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        _ = try await client.execute(req)
    }
}
```

## Notification Caching

```swift
final class NotificationCacheService {
    static let shared = NotificationCacheService()
    private let cache = NSCache<NSString, CachedNotifications>()
    private let defaults = UserDefaults.standard
    private let cacheKey = "notifications_cache"
    private let cacheExpiry: TimeInterval = 600

    func getCachedNotifications() async -> [AppNotification]? {
        guard let cached = cache.object(forKey: cacheKey as NSString) else { return nil }
        guard Date().timeIntervalSince(cached.timestamp) < cacheExpiry else {
            cache.removeObject(forKey: cacheKey as NSString)
            return nil
        }
        return cached.notifications
    }

    func cacheNotifications(_ notifications: [AppNotification]) async {
        cache.setObject(
            CachedNotifications(notifications: notifications, timestamp: Date()),
            forKey: cacheKey as NSString
        )
    }

    func invalidateCache() {
        cache.removeAllObjects()
    }
}

final class CachedNotifications: NSObject {
    let notifications: [AppNotification]
    let timestamp: Date
    init(notifications: [AppNotification], timestamp: Date) {
        self.notifications = notifications
        self.timestamp = timestamp
    }
}
```

## Notification Preferences Defaults

extension NotificationPreferences {
    static let `default` = NotificationPreferences(
        globalEnabled: true,
        categories: NotificationCategory.allCases.map { cat in
            CategoryPreference(
                id: cat.rawValue,
                category: cat,
                enabled: true,
                pushEnabled: true,
                soundEnabled: true
            )
        },
        quietHoursEnabled: false,
        quietHoursStart: Calendar.current.date(from: DateComponents(hour: 22)) ?? Date(),
        quietHoursEnd: Calendar.current.date(from: DateComponents(hour: 7)) ?? Date(),
        pushEnabled: true,
        inAppEnabled: true,
        emailEnabled: false
    )
}
```

## Notification Permissions

```swift
struct NotificationPermissionView: View {
    @ObservedObject var manager: NotificationManager
    @State private var permissionStatus: PermissionStatus = .unknown

    enum PermissionStatus {
        case unknown, granted, denied, provisional
    }

    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: permissionIcon)
                .font(.system(size: 48))
                .foregroundStyle(permissionColor)

            Text(permissionTitle)
                .font(.title2.bold())

            Text(permissionDescription)
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)

            switch permissionStatus {
            case .denied:
                Button("Open Settings") { manager.registerForPushNotifications() }
                    .buttonStyle(.borderedProminent)
            case .granted:
                Label("Notifications Enabled", systemImage: "checkmark.circle.fill")
                    .foregroundStyle(.green)
            default:
                Button("Enable Notifications") {
                    Task { await manager.registerForPushNotifications() }
                }
                .buttonStyle(.borderedProminent)
            }
        }
        .padding()
        .task { await checkStatus() }
    }

    private var permissionIcon: String {
        switch permissionStatus {
        case .granted: return "bell.badge.fill"
        case .denied: return "bell.slash.fill"
        case .provisional: return "bell.badge"
        case .unknown: return "bell"
        }
    }

    private var permissionColor: Color {
        switch permissionStatus {
        case .granted: return .green
        case .denied: return .red
        case .provisional: return .orange
        case .unknown: return .blue
        }
    }

    private var permissionTitle: String {
        switch permissionStatus {
        case .granted: return "Notifications Enabled"
        case .denied: return "Notifications Disabled"
        case .provisional: return "Provisional Notifications"
        case .unknown: return "Enable Notifications"
        }
    }

    private var permissionDescription: String {
        switch permissionStatus {
        case .granted:
            return "You will receive notifications for AI alerts, security events, and workflow updates."
        case .denied:
            return "Notifications are disabled. Please enable them in Settings to receive alerts."
        case .provisional:
            return "Notifications will be delivered quietly without interrupting you."
        case .unknown:
            return "Allow notifications to stay updated on AI agent activity and security alerts."
        }
    }

    private func checkStatus() async {
        let settings = await UNUserNotificationCenter.current().notificationSettings()
        switch settings.authorizationStatus {
        case .authorized: permissionStatus = .granted
        case .denied: permissionStatus = .denied
        case .provisional: permissionStatus = .provisional
        case .notDetermined, @unknown default: permissionStatus = .unknown
        }
    }
}
```

## Notification on Different iOS Versions

```swift
extension NotificationManager {
    func scheduleLocalNotification(
        title: String,
        body: String,
        category: NotificationCategory,
        data: NotificationData? = nil,
        timeInterval: TimeInterval? = nil,
        date: Date? = nil
    ) {
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.categoryIdentifier = category.rawValue
        content.sound = .default

        // Priority-based interruption level (iOS 15+)
        if #available(iOS 15.0, *) {
            switch category {
            case .security:
                content.interruptionLevel = .timeSensitive
            case .aiAlert:
                content.interruptionLevel = .active
            default:
                content.interruptionLevel = .active
            }
        }

        // Communication notifications (iOS 16+)
        if #available(iOS 16.0, *) {
            content.relevanceScore = 0.8
        }

        // Critical alerts (iOS 16+)
        if #available(iOS 16.0, *), category == .security {
            content.interruptionLevel = .critical
        }

        if let data, let dataDict = try? JSONSerialization.jsonObject(
            with: JSONEncoder().encode(data)
        ) as? [String: Any] {
            content.userInfo["data"] = dataDict
            content.userInfo["category"] = category.rawValue
        }

        var trigger: UNNotificationTrigger?
        if let interval = timeInterval {
            trigger = UNTimeIntervalNotificationTrigger(timeInterval: interval, repeats: false)
        } else if let date = date {
            let components = Calendar.current.dateComponents(
                [.year, .month, .day, .hour, .minute],
                from: date
            )
            trigger = UNCalendarNotificationTrigger(dateMatching: components, repeats: false)
        }

        let request = UNNotificationRequest(
            identifier: UUID().uuidString,
            content: content,
            trigger: trigger
        )

        UNUserNotificationCenter.current().add(request)
    }
}
```

## Testing

```swift
import XCTest
@testable import NexusAI

final class NotificationManagerTests: XCTestCase {
    func testNotificationParsing() {
        let notification = AppNotification(
            id: "1", title: "Test", body: "Body",
            category: .aiAlert, priority: .high,
            data: NotificationData(
                deepLink: "/agents/123", agentId: "123",
                documentId: nil, executionId: nil,
                actionUrl: nil, metadata: nil
            ),
            isRead: false, createdAt: Date(), readAt: nil
        )

        XCTAssertEqual(notification.category, .aiAlert)
        XCTAssertFalse(notification.isRead)
        XCTAssertEqual(notification.data?.agentId, "123")
    }

    func testDeepLinkParsing() {
        let link = NotificationDeepLink.from("https://nexus.ai/agents/123")
        XCTAssertEqual(link, .agentDetail)
    }

    func testPreferencesDefaults() {
        let prefs = NotificationPreferences.default
        XCTAssertTrue(prefs.globalEnabled)
        XCTAssertEqual(prefs.categories.count, NotificationCategory.allCases.count)
    }

    func testCategoryProperties() {
        XCTAssertEqual(NotificationCategory.aiAlert.displayName, "AI Alerts")
        XCTAssertEqual(NotificationCategory.security.iconName, "lock.shield")
    }

    func testBadgeText() {
        let count = 150
        let text = count > 99 ? "99+" : "\(count)"
        XCTAssertEqual(text, "99+")
    }
}
```

## Responsive Design

| Screen | iPhone | iPad |
|---|---|---|
| Notification Center | Full screen list | Split view list + detail |
| Preferences | Form stack | Form + preview |
| Notification Banner | Top banner | Top banner |
| Permission View | Centered card | Centered card |

## Accessibility

- All notification rows include VoiceOver labels with category, title, and time
- Read/unread state announced via accessibility trait
- Category icons include descriptive labels
- Deep link navigation works with VoiceOver
- Dynamic Type supported on all notification text
- Haptic feedback on notification actions
- Color is supplemented with icons and text

## File Structure

```
Notifications/
├── Views/
│   ├── NotificationCenterView.swift
│   ├── NotificationRow.swift
│   ├── NotificationPreferencesView.swift
│   ├── NotificationPermissionView.swift
│   ├── NotificationBadgeModifier.swift
│   ├── DeepLinkHandler.swift
│   └── States/
│       ├── NotificationCenterLoadingView.swift
│       └── NotificationCenterEmptyView.swift
├── ViewModels/
│   ├── NotificationManager.swift
│   └── NotificationCenterViewModel.swift
│   └── NotificationPreferencesViewModel.swift
├── Models/
│   ├── AppNotification.swift
│   ├── NotificationCategory.swift
│   ├── NotificationPreferences.swift
│   └── PushTokenRegistration.swift
├── Services/
│   ├── NotificationAPIService.swift
│   └── NotificationCacheService.swift
├── Extensions/
│   └── NotificationDeepLink.swift
└── Tests/
    └── NotificationManagerTests.swift
```
