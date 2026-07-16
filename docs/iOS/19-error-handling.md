# 19. Error Handling

## Table of Contents

- [Error Handling Architecture](#error-handling-architecture)
- [Error Types](#error-types)
- [Error Classification](#error-classification)
- [Network Errors](#network-errors)
- [Auth Errors](#auth-errors)
- [Validation Errors](#validation-errors)
- [Server & Client Errors](#server--client-errors)
- [Rate Limit Errors](#rate-limit-errors)
- [WebSocket Errors](#websocket-errors)
- [Error UI Patterns](#error-ui-patterns)
- [Error Alerts](#error-alerts)
- [Error Toasts](#error-toasts)
- [Error Inline](#error-inline)
- [Error Full-Screen](#error-full-screen)
- [Empty State UI](#empty-state-ui)
- [Loading State UI](#loading-state-ui)
- [Retry Logic](#retry-logic)
- [Retry UI](#retry-ui)
- [Error Boundaries](#error-boundaries)
- [Error Logging](#error-logging)
- [Error Monitoring & Reporting](#error-monitoring--reporting)
- [Error Recovery Strategies](#error-recovery-strategies)
- [Error Accessibility](#error-accessibility)
- [Error Testing](#error-testing)
- [Error Debugging & Metrics](#error-debugging--metrics)

---

## Error Handling Architecture

### Error Domain Model

```
NexusError (Protocol)
├── NetworkError: noConnection, timeout, dnsFailure, sslError, requestFailed
├── AuthError: invalidCredentials, tokenExpired, accountLocked, kycPending
├── ValidationError: fieldError, formError, businessRule
├── ServerError: internalServerError, badGateway, serviceUnavailable
├── ClientError: badRequest, forbidden, notFound, conflict, unprocessableEntity
├── RateLimitError: rateLimited(retryAfter:)
├── WebSocketError: connectionDropped, authenticationFailed, parsingFailed
└── CacheError: notFound, corrupted, storageFull
```

### Protocol Definition

```swift
protocol NexusError: Error, LocalizedError {
    var code: String { get }
    var title: String { get }
    var message: String { get }
    var recoverySuggestion: String? { get }
    var isRetryable: Bool { get }
    var severity: ErrorSeverity { get }
}

enum ErrorSeverity { case info, warning, critical, fatal }

struct FieldError: Identifiable, Equatable {
    let id = UUID(); let field: String; let message: String
}
```

---

## Error Types

```swift
enum NexusAppError: NexusError {
    case noConnection, timeout, dnsFailure, sslError(Error), requestFailed(statusCode: Int)
    case invalidCredentials, tokenExpired, accountLocked, kycPending, biometricUnavailable
    case fieldError(field: String, message: String), formError([FieldError]), businessRule(String)
    case internalServerError(Int), badGateway, serviceUnavailable, gatewayTimeout
    case badRequest, forbidden, notFound, conflict, unprocessableEntity([FieldError])
    case rateLimited(retryAfter: TimeInterval)
    case connectionDropped, authenticationFailed, messageParsingFailed, connectionTimeout
    case cacheNotFound, cacheCorrupted, storageFull

    var code: String {
        switch self {
        case .noConnection: return "NET_001"; case .timeout: return "NET_002"
        case .dnsFailure: return "NET_003"; case .sslError: return "NET_004"
        case .requestFailed(let c): return "NET_\(c)"
        case .invalidCredentials: return "AUTH_001"; case .tokenExpired: return "AUTH_002"
        case .accountLocked: return "AUTH_003"; case .kycPending: return "AUTH_004"
        case .biometricUnavailable: return "AUTH_005"
        case .fieldError: return "VAL_001"; case .formError: return "VAL_002"; case .businessRule: return "VAL_003"
        case .internalServerError: return "SRV_001"; case .badGateway: return "SRV_002"
        case .serviceUnavailable: return "SRV_003"; case .gatewayTimeout: return "SRV_004"
        case .badRequest: return "CLI_001"; case .forbidden: return "CLI_003"; case .notFound: return "CLI_004"
        case .conflict: return "CLI_009"; case .unprocessableEntity: return "CLI_022"; case .rateLimited: return "CLI_029"
        case .connectionDropped: return "WS_001"; case .authenticationFailed: return "WS_002"
        case .messageParsingFailed: return "WS_003"; case .connectionTimeout: return "WS_004"
        case .cacheNotFound: return "CACHE_001"; case .cacheCorrupted: return "CACHE_002"; case .storageFull: return "CACHE_003"
        }
    }

    var title: String {
        switch self {
        case .noConnection: return "No Internet Connection"; case .timeout: return "Request Timed Out"
        case .dnsFailure, .requestFailed: return "Connection Error"; case .sslError: return "Security Error"
        case .invalidCredentials: return "Login Failed"; case .tokenExpired: return "Session Expired"
        case .accountLocked: return "Account Locked"; case .kycPending: return "Verification Pending"
        case .biometricUnavailable: return "Biometric Not Available"
        case .fieldError(_, let m): return m; case .formError: return "Invalid Form"; case .businessRule(let m): return m
        case .internalServerError: return "Server Error"; case .badGateway, .serviceUnavailable: return "Service Unavailable"
        case .gatewayTimeout: return "Server Timeout"; case .badRequest: return "Bad Request"
        case .forbidden: return "Access Denied"; case .notFound: return "Not Found"; case .conflict: return "Conflict"
        case .unprocessableEntity: return "Invalid Data"; case .rateLimited: return "Too Many Requests"
        case .connectionDropped: return "Connection Lost"; case .authenticationFailed: return "Auth Failed"
        case .messageParsingFailed: return "Message Error"; case .connectionTimeout: return "Connection Timeout"
        case .cacheNotFound: return "Not Cached"; case .cacheCorrupted: return "Cache Error"; case .storageFull: return "Storage Full"
        }
    }

    var message: String {
        switch self {
        case .noConnection: return "Check your internet connection and try again."
        case .timeout: return "The request took too long. Please try again."
        case .dnsFailure: return "Could not reach the server."
        case .sslError: return "A security error occurred."
        case .requestFailed(let c): return "Request failed (status \(c))."
        case .invalidCredentials: return "Email or password is incorrect."
        case .tokenExpired: return "Session expired. Please log in again."
        case .accountLocked: return "Account locked. Contact support."
        case .kycPending: return "Identity verification is pending."
        case .biometricUnavailable: return "Biometric auth not available on this device."
        case .fieldError(_, let m): return m
        case .formError(let e): return e.map(\.message).joined(separator: "\n")
        case .businessRule(let m): return m
        case .internalServerError: return "Something went wrong. Try again later."
        case .badGateway, .serviceUnavailable: return "Service temporarily unavailable."
        case .gatewayTimeout: return "Server took too long to respond."
        case .badRequest: return "Invalid request."
        case .forbidden: return "You don't have permission."
        case .notFound: return "Resource not found."
        case .conflict: return "Conflict with current state."
        case .unprocessableEntity: return "Invalid data provided."
        case .rateLimited(let t): return "Too many requests. Wait \(Int(t))s."
        case .connectionDropped: return "Connection lost."
        case .authenticationFailed: return "Auth failed. Log in again."
        case .messageParsingFailed: return "Could not process server message."
        case .connectionTimeout: return "Connection timed out."
        case .cacheNotFound: return "Data not available offline."
        case .cacheCorrupted: return "Local data corrupted. Refresh."
        case .storageFull: return "Device storage full."
        }
    }

    var recoverySuggestion: String? {
        switch self {
        case .noConnection: return "Connect to Wi-Fi or enable mobile data."
        case .tokenExpired: return "Tap Log In to re-authenticate."
        case .rateLimited(let t): return "Wait \(Int(t)) seconds."
        case .storageFull: return "Free up space in Settings > General > iPhone Storage."
        default: return nil
        }
    }

    var isRetryable: Bool {
        switch self {
        case .noConnection, .timeout, .dnsFailure, .requestFailed, .internalServerError,
             .badGateway, .serviceUnavailable, .gatewayTimeout, .connectionDropped,
             .connectionTimeout, .rateLimited: return true
        default: return false
        }
    }

    var severity: ErrorSeverity {
        switch self {
        case .fieldError, .cacheNotFound: return .info
        case .noConnection, .timeout, .rateLimited, .connectionDropped: return .warning
        case .tokenExpired, .invalidCredentials, .serviceUnavailable: return .critical
        case .internalServerError: return .fatal
        default: return .warning
        }
    }
}
```

---

## Error Classification

| Category    | Recoverable | Auto-Retry | User Action    | Example            |
|-------------|-------------|------------|----------------|--------------------|
| Network     | Yes         | Yes        | None           | No connection      |
| Auth        | Yes         | No         | Re-login       | Token expired      |
| Validation  | Yes         | No         | Fix input      | Invalid email      |
| Server      | Yes         | Yes        | Wait/Retry     | 500 error          |
| Client      | Depends     | No         | Fix request    | 404 Not Found      |
| Rate Limit  | Yes         | Yes        | Wait           | 429 Too Many       |
| WebSocket   | Yes         | Yes        | None           | Connection drop    |

### Decision Tree

```
Error Received → Retryable?
  YES → Rate-limited?
    YES → Wait retryAfter, then retry
    NO  → Exponential backoff
  NO  → Auth error?
    YES → Force re-login
    NO  → Show error UI
         → Cache fallback? → Show cached + warning
         → No cache → Empty state / error screen
```

---

## Network Errors

```swift
struct NetworkErrorHandler {
    func handle(_ error: Error) -> NexusAppError {
        if let url = error as? URLError {
            switch url.code {
            case .notConnectedToInternet: return .noConnection
            case .timedOut: return .timeout
            case .cannotFindHost: return .dnsFailure
            case .secureConnectionFailed: return .sslError(url)
            case .networkConnectionLost: return .connectionDropped
            default: return .requestFailed(statusCode: url.errorCode)
            }
        }
        if let api = error as? APIError {
            switch api {
            case .unauthorized: return .tokenExpired
            case .notFound: return .notFound
            case .rateLimited(let t): return .rateLimited(retryAfter: t)
            case .serverError(let c): return .internalServerError(c)
            default: return .badRequest
            }
        }
        return .requestFailed(statusCode: -1)
    }
}
```

---

## Auth Errors

```swift
@MainActor
class AuthErrorHandler {
    func handle(_ error: NexusAppError) async -> AuthRecoveryAction {
        switch error {
        case .tokenExpired:
            return (try? await authService.refreshToken()) != nil ? .tokenRefreshed : .requireLogin
        case .kycPending: return .navigateToVerification
        case .biometricUnavailable: return .fallbackToPassword
        default: return .showError(error)
        }
    }
}

enum AuthRecoveryAction {
    case tokenRefreshed, requireLogin, navigateToVerification, fallbackToPassword, showError(NexusAppError)
}
```

---

## Validation Errors

```swift
struct FormValidator {
    func validateLogin(email: String, password: String) -> Result<Void, [FieldError]> {
        var errors: [FieldError] = []
        if email.isEmpty { errors.append(.init(field: "email", message: "Email required")) }
        else if !email.isValidEmail { errors.append(.init(field: "email", message: "Invalid email")) }
        if password.isEmpty { errors.append(.init(field: "password", message: "Password required")) }
        else if password.count < 8 { errors.append(.init(field: "password", message: "Min 8 characters")) }
        return errors.isEmpty ? .success(()) : .failure(errors)
    }

    func validateTransfer(amount: String, recipientId: String, balance: Double) -> Result<Void, [FieldError]> {
        var errors: [FieldError] = []
        guard let v = Double(amount) else { return .failure([.init(field: "amount", message: "Invalid amount")]) }
        if v <= 0 { errors.append(.init(field: "amount", message: "Must be positive")) }
        if v > balance { errors.append(.init(field: "amount", message: "Insufficient balance")) }
        if recipientId.isEmpty { errors.append(.init(field: "recipient", message: "Select recipient")) }
        return errors.isEmpty ? .success(()) : .failure(errors)
    }
}
```

---

## Server & Client Errors

| Code | Name                | Handling                     |
|------|---------------------|------------------------------|
| 500  | Internal Error      | Retry 5s                     |
| 502  | Bad Gateway         | Retry 10s                    |
| 503  | Service Unavailable | Exponential backoff          |
| 504  | Gateway Timeout     | Retry 15s                    |
| 400  | Bad Request         | Log + generic error          |
| 401  | Unauthorized        | Refresh token / re-login     |
| 403  | Forbidden           | "Access Denied"              |
| 404  | Not Found           | Empty state                  |
| 422  | Unprocessable       | Parse field errors           |

---

## Rate Limit Errors

```swift
class RateLimitHandler {
    private var retryAfter: TimeInterval = 0
    private var lastRateLimitTime: Date?

    func handle(retryAfter: TimeInterval) -> RateLimitState {
        self.retryAfter = retryAfter; self.lastRateLimitTime = Date()
        return .init(shouldWait: true, waitDuration: retryAfter, message: "Wait \(Int(retryAfter))s")
    }

    func canRetry() -> Bool {
        guard let t = lastRateLimitTime else { return true }
        return Date().timeIntervalSince(t) >= retryAfter
    }
}
```

---

## WebSocket Errors

```swift
class WebSocketErrorHandler {
    func handle(_ error: WebSocketError) -> WebSocketRecovery {
        switch error {
        case .connectionDropped: return .init(action: .reconnect, delay: 1, maxAttempts: 5, backoff: 2)
        case .authenticationFailed: return .init(action: .reauthenticate, delay: 0, maxAttempts: 1, backoff: 1)
        case .connectionTimeout: return .init(action: .reconnect, delay: 5, maxAttempts: 3, backoff: 1.5)
        case .messageParsingFailed: return .init(action: .ignore, delay: 0, maxAttempts: 0, backoff: 1)
        }
    }
}
```

---

## Error UI Patterns

```
Severity → UI:
  INFO       → Inline message, Toast
  WARNING    → Toast with action, Banner
  CRITICAL   → Alert dialog, Full-screen
  FATAL      → Full-screen with retry
```

---

## Error Alerts

```swift
struct ErrorAlertModifier: ViewModifier {
    @Binding var error: NexusAppError?
    let onRetry: (() -> Void)?

    func body(content: Content) -> some View {
        content.alert(error?.title ?? "Error",
            isPresented: Binding(get: { error != nil }, set: { if !$0 { error = nil } }),
            presenting: error
        ) { err in
            if err.isRetryable, let onRetry { Button("Retry") { onRetry() } }
            Button("OK", role: .cancel) { error = nil }
        } message: { err in
            VStack(alignment: .leading) {
                Text(err.message)
                if let s = err.recoverySuggestion { Text(s).font(.caption).foregroundColor(.secondary) }
            }
        }
    }
}

extension View {
    func errorAlert(error: Binding<NexusAppError?>, onRetry: (() -> Void)? = nil) -> some View {
        modifier(ErrorAlertModifier(error: error, onRetry: onRetry))
    }
}

// Usage
struct LoginView: View {
    @StateObject var vm: LoginViewModel
    var body: some View {
        LoginForm(viewModel: vm).errorAlert(error: $vm.error, onRetry: { vm.login() })
    }
}
```

---

## Error Toasts

```swift
struct ToastView: View {
    let message: ToastMessage
    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: message.icon).foregroundColor(.white)
            Text(message.text).font(.subheadline).foregroundColor(.white)
            if let a = message.action { Button(a.title) { a.handler() }.font(.subheadline.bold()).foregroundColor(.white) }
            Spacer()
        }.padding().background(message.color).cornerRadius(12).padding(.horizontal).padding(.bottom, 8)
    }
}

struct ToastMessage: Identifiable {
    let id = UUID(); let text: String; let icon: String; let color: Color; let duration: TimeInterval
    let action: ToastAction?
    struct ToastAction { let title: String; let handler: () -> Void }
    static func error(_ t: String, retry: (() -> Void)? = nil) -> ToastMessage {
        .init(text: t, icon: "exclamationmark.triangle.fill", color: .red, duration: 5,
              action: retry.map { .init(title: "Retry", handler: $0) })
    }
}
```

---

## Error Inline

```swift
struct FormFieldView: View {
    let label: String; let error: String?; @Binding var text: String
    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            TextField(label, text: $text).textFieldStyle(.roundedBorder)
                .overlay(RoundedRectangle(cornerRadius: 8).stroke(error != nil ? Color.red : .clear, lineWidth: 1))
            if let e = error {
                HStack(spacing: 4) {
                    Image(systemName: "exclamationmark.circle.fill").foregroundColor(.red).font(.caption)
                    Text(e).foregroundColor(.red).font(.caption)
                }.accessibilityElement(children: .combine).accessibilityLabel("Error: \(e)")
            }
        }
    }
}

@MainActor
class TransferViewModel: ObservableObject {
    @Published var amount = ""; @Published var recipientId = ""; @Published var fieldErrors: [String: String] = [:]
    func validate() -> Bool {
        fieldErrors.removeAll()
        if amount.isEmpty || Double(amount) == nil { fieldErrors["amount"] = "Enter valid amount" }
        if recipientId.isEmpty { fieldErrors["recipient"] = "Select recipient" }
        return fieldErrors.isEmpty
    }
}
```

---

## Error Full-Screen

```swift
struct FullScreenErrorView: View {
    let error: NexusAppError; let onRetry: () -> Void; let onContactSupport: (() -> Void)?
    var body: some View {
        VStack(spacing: 24) {
            Spacer()
            Image(systemName: errorIcon).font(.system(size: 64)).foregroundColor(errorColor)
            Text(error.title).font(.title2.bold()).multilineTextAlignment(.center)
            Text(error.message).font(.body).foregroundColor(.secondary).multilineTextAlignment(.center).padding(.horizontal, 32)
            if let s = error.recoverySuggestion { Text(s).font(.subheadline).foregroundColor(.secondary) }
            VStack(spacing: 12) {
                if error.isRetryable {
                    Button(action: onRetry) { Label("Try Again", systemImage: "arrow.clockwise").frame(maxWidth: .infinity) }.buttonStyle(.borderedProminent)
                }
                if let onContactSupport {
                    Button(action: onContactSupport) { Label("Contact Support", systemImage: "envelope").frame(maxWidth: .infinity) }.buttonStyle(.bordered)
                }
            }.padding(.horizontal, 32)
            Spacer(); Spacer()
        }
    }
    private var errorIcon: String {
        switch error { case .noConnection: return "wifi.slash"; case .tokenExpired: return "lock.rotation"
        case .notFound: return "questionmark.folder"; default: return "exclamationmark.circle" }
    }
    private var errorColor: Color {
        switch error.severity { case .info: return .blue; case .warning: return .orange; case .critical, .fatal: return .red }
    }
}
```

---

## Empty State UI

```swift
struct EmptyStateView: View {
    let type: EmptyStateType

    enum EmptyStateType {
        case noData, noConnection, noResults, firstTime, error(NexusAppError)
        var icon: String { switch self { case .noData: "tray"; case .noConnection: "wifi.slash"; case .noResults: "magnifyingglass"; case .firstTime: "sparkles"; case .error: "exclamationmark.triangle" } }
        var title: String { switch self { case .noData: "No Data"; case .noConnection: "No Connection"; case .noResults: "No Results"; case .firstTime: "Welcome!"; case .error(let e): e.title } }
        var message: String { switch self { case .noData: "Nothing here yet."; case .noConnection: "Check your connection."; case .noResults: "Try different search."; case .firstTime: "Add your first item."; case .error(let e): e.message } }
    }

    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: type.icon).font(.system(size: 48)).foregroundColor(.secondary)
            Text(type.title).font(.headline)
            Text(type.message).font(.subheadline).foregroundColor(.secondary).multilineTextAlignment(.center).padding(.horizontal, 32)
        }.frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
```

---

## Loading State UI

```swift
struct LoadingView: View {
    let message: String?
    var body: some View {
        VStack(spacing: 16) {
            ProgressView().scaleEffect(1.2)
            if let m = message { Text(m).font(.subheadline).foregroundColor(.secondary) }
        }.frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct SkeletonView: View {
    var body: some View {
        LinearGradient(colors: [.gray.opacity(0.2), .gray.opacity(0.4), .gray.opacity(0.2)],
            startPoint: .leading, endPoint: .trailing).mask(Rectangle())
    }
}
```

---

## Retry Logic

```swift
actor RetryManager {
    func execute<T>(maxRetries: Int = 3, initialDelay: TimeInterval = 1.0, backoffMultiplier: Double = 2.0,
                    shouldRetry: @escaping (Error) -> Bool = { _ in true },
                    operation: @escaping () async throws -> T) async throws -> T {
        var lastError: Error?; var delay = initialDelay
        for attempt in 0..<maxRetries {
            do { return try await operation() } catch {
                lastError = error
                guard attempt < maxRetries - 1, shouldRetry(error) else { throw error }
                try await Task.sleep(nanoseconds: UInt64(delay * 1_000_000_000))
                delay *= backoffMultiplier
            }
        }
        throw lastError ?? RetryError.maxRetriesExceeded
    }
}

enum RetryError: Error { case maxRetriesExceeded }

// Usage
class TransactionViewModel: ObservableObject {
    private let retryManager = RetryManager()
    func sendMoney(amount: Double, to recipientId: String) async {
        do {
            _ = try await retryManager.execute(maxRetries: 3, shouldRetry: { ($0 as? NexusAppError)?.isRetryable ?? false }) {
                [self] in try await transactionService.send(amount: amount, to: recipientId)
            }
        } catch { /* Handle */ }
    }
}
```

---

## Retry UI

```swift
struct RetryView: View {
    let error: NexusAppError; let onRetry: () -> Void; @State private var isRetrying = false
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "arrow.clockwise.circle").font(.system(size: 36)).foregroundColor(.blue)
            Text(error.message).font(.subheadline).foregroundColor(.secondary)
            Button(action: { isRetrying = true; onRetry() }) {
                if isRetrying { ProgressView().tint(.white) } else { Label("Try Again", systemImage: "arrow.clockwise") }
            }.buttonStyle(.borderedProminent).disabled(isRetrying)
        }
    }
}

struct AutoRetryView: View {
    let retryAfter: TimeInterval; let onRetry: () -> Void; @State private var countdown: Int
    init(retryAfter: TimeInterval, onRetry: @escaping () -> Void) {
        self.retryAfter = retryAfter; self.onRetry = onRetry; _countdown = State(initialValue: Int(retryAfter))
    }
    var body: some View {
        VStack(spacing: 12) {
            Text("Retrying in \(countdown)s...").font(.subheadline).foregroundColor(.secondary)
            ProgressView(value: Double(countdown), total: retryAfter).tint(.blue)
            Button("Retry Now") { onRetry() }.buttonStyle(.bordered)
        }.onAppear { Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { t in if countdown > 0 { countdown -= 1 } else { t.invalidate(); onRetry() } } }
    }
}
```

---

## Error Boundaries

```swift
@MainActor
class GlobalErrorHandler: ObservableObject {
    static let shared = GlobalErrorHandler()
    @Published var currentError: NexusAppError?
    @Published var errorHistory: [ErrorEvent] = []

    struct ErrorEvent: Identifiable {
        let id = UUID(); let error: NexusAppError; let timestamp: Date; let context: String
    }

    func handle(_ error: NexusAppError, context: String = "Unknown") {
        currentError = error
        errorHistory.append(.init(error: error, timestamp: Date(), context: context))
        ErrorLogger.shared.log(error, context: context)
    }
    func dismiss() { currentError = nil }
}
```

---

## Error Logging

```swift
import OSLog
import FirebaseCrashlytics

class ErrorLogger {
    static let shared = ErrorLogger()
    private let logger = Logger(subsystem: "com.nexus.app", category: "Errors")

    func log(_ error: NexusAppError, context: String = "") {
        logger.error("[\(error.code)] \(error.title): \(error.message) | \(context)")
        Crashlytics.crashlytics().record(error: error as NSError, userInfo: [
            "error_code": error.code, "context": context
        ])
    }

    func logBreadcrumb(_ message: String, category: String = "Nav") {
        Crashlytics.crashlytics().log("[\(category)] \(message)")
    }
}
```

---

## Error Monitoring & Reporting

| Metric              | Threshold | Action     |
|---------------------|-----------|------------|
| Crash Rate          | > 0.1%    │ Alert      |
| ANR Rate            | > 0.05%   │ Investigate|
| Non-Fatal Rate      │ > 1%      │ Review     |
| Network Error Rate  │ > 5%      │ Check API  |
| Auth Error Rate     │ > 10%     │ Check Auth |

---

## Error Recovery Strategies

| Error Type      | Primary Strategy     | Fallback         |
|-----------------|----------------------|------------------|
| No Connection   | Cache fallback       | Offline mode     |
| Token Expired   | Silent refresh       | Re-login         |
| Server Error    | Retry + backoff      | Cached data      |
| Rate Limited    | Queue & wait         | Cached data      |
| WebSocket Drop  | Auto-reconnect       | HTTP polling     |
| Validation      | Show field errors    | —                |
| Not Found       | Empty state          | Search           |
| Storage Full    | Cleanup cache        | Prompt user      |

### Cache Fallback

```swift
class CachedRepository<T: Codable> {
    private let remote: () async throws -> T
    private let local: () async throws -> T
    private let saveLocal: (T) async throws -> Void

    func fetch() async throws -> T {
        do { let d = try await remote(); try await saveLocal(d); return d }
        catch let e as NexusAppError where e.isRetryable { return try await local() }
    }
}
```

---

## Error Accessibility

```swift
struct AccessibleErrorModifier: ViewModifier {
    @Binding var error: NexusAppError?
    func body(content: Content) -> some View {
        content.onChange(of: error) { _, newError in
            if let e = newError { UIAccessibility.post(notification: .announcement, argument: "\(e.title). \(e.message)") }
        }
    }
}

struct AccessibleErrorView: View {
    let error: NexusAppError
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle.fill").font(.title).accessibilityHidden(true)
            Text(error.title).font(.headline).accessibilityAddTraits(.isHeader)
            Text(error.message).font(.body)
            if let s = error.recoverySuggestion { Text(s).font(.subheadline).foregroundColor(.secondary) }
        }.accessibilityElement(children: .combine).accessibilityLabel("\(error.title). \(error.message)")
    }
}
```

---

## Error Testing

```swift
final class ErrorHandlerTests: XCTestCase {
    func testNetworkHandler_mapsURLError() {
        let h = NetworkErrorHandler()
        XCTAssertEqual(h.handle(URLError(.notConnectedToInternet)), .noConnection)
        XCTAssertEqual(h.handle(URLError(.timedOut)), .timeout)
    }

    func testFormValidator() {
        let v = FormValidator()
        if case .failure(let e) = v.validateLogin(email: "", password: "") { XCTAssertEqual(e.count, 2) }
        if case .success = v.validateLogin(email: "u@t.com", password: "12345678") {} else { XCTFail() }
    }

    func testRetryManager_retries() async {
        var count = 0
        let r = try? await RetryManager().execute(maxRetries: 3, initialDelay: 0.01) {
            count += 1; if count < 3 { throw NexusAppError.timeout }; return "ok"
        }
        XCTAssertEqual(r, "ok"); XCTAssertEqual(count, 3)
    }

    func testRateLimit_enforcesWait() {
        let h = RateLimitHandler(); _ = h.handle(retryAfter: 30); XCTAssertFalse(h.canRetry())
    }
}
```

---

## Error Debugging & Metrics

```swift
#if DEBUG
struct DebugErrorOverlay: View {
    @State private var showError = false
    @State private var selectedError: NexusAppError = .noConnection
    let errors: [NexusAppError] = [.noConnection, .timeout, .tokenExpired, .rateLimited(retryAfter: 30)]
    var body: some View {
        Menu("Simulate Error") {
            ForEach(errors, id: \.code) { e in Button(e.title) { selectedError = e; showError = true } }
        }.errorAlert(error: showError ? $selectedError : .constant(nil))
    }
}
#endif
```

```swift
struct ErrorMetrics {
    static func calculate(from events: [ErrorEvent]) -> ErrorMetricsReport {
        guard !events.isEmpty else { return .empty }
        return ErrorMetricsReport(
            totalErrors: events.count,
            criticalErrors: events.filter { $0.error.severity == .critical || $0.error.severity == .fatal }.count,
            retryableErrors: events.filter { $0.error.isRetryable }.count)
    }
}

struct ErrorMetricsReport {
    let totalErrors, criticalErrors, retryableErrors: Int
    static let empty = ErrorMetricsReport(totalErrors: 0, criticalErrors: 0, retryableErrors: 0)
}
```
