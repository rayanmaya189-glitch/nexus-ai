# iOS Networking

## Table of Contents

- [URLSession Configuration](#urlsession-configuration)
- [APIClient](#apiclient)
- [Authentication Interceptor](#authentication-interceptor)
- [Token Refresh Interceptor](#token-refresh-interceptor)
- [Certificate Pinning](#certificate-pinning)
- [Network Security Configuration](#network-security-configuration)
- [WebSocket Client](#websocket-client)
- [WebSocket Message Protocol](#websocket-message-protocol)
- [WebSocket Authentication](#websocket-authentication)
- [WebSocket Reconnection](#websocket-reconnection)
- [API Endpoint Definitions](#api-endpoint-definitions)
- [Request/Response Models](#requestresponse-models)
- [Error Handling](#error-handling)
- [Offline Detection](#offline-detection)
- [Retry Logic](#retry-logic)
- [Request Cancellation](#request-cancellation)
- [Mock Data for Testing](#mock-data-for-testing)
- [Network Monitoring](#network-monitoring)

---

## URLSession Configuration

```swift
// Custom URLSession Configuration
extension URLSession {
    static let apiSession: URLSession = {
        let config = URLSessionConfiguration.default

        // Timeout Configuration
        config.timeoutIntervalForRequest = 30
        config.timeoutIntervalForResource = 300

        // Cache Policy
        config.requestCachePolicy = .reloadIgnoringLocalCacheData
        config.urlCache = nil // No caching for API calls

        // Connection Pool
        config.httpMaximumConnectionsPerHost = 6
        config.httpShouldSetCookies = false
        config.httpCookieAcceptPolicy = .never
        config.httpCookieStorage = nil

        // TLS
        config.tlsMinimumSupportedProtocolVersion = .TLSv12
        config.tlsMaximumSupportedProtocolVersion = .TLSv13

        // Misc
        config.waitsForConnectivity = true
        config.networkServiceType = .default
        config.allowsExpensiveNetworkAccess = true
        config.allowsConstrainedNetworkAccess = true

        // Headers
        config.httpAdditionalHeaders = [
            "Accept": "application/json",
            "Accept-Language": Locale.current.language.languageCode?.identifier ?? "en",
            "X-Platform": "iOS",
            "X-App-Version": Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "unknown"
        ]

        return config
    }()

    static let webSocketSession: URLSession = {
        let config = URLSessionConfiguration.default
        config.waitsForConnectivity = true
        return URLSession(configuration: config)
    }()
}
```

### URLSession Configuration Table

| Setting | Value | Purpose |
|---------|-------|---------|
| `timeoutIntervalForRequest` | 30s | Per-request timeout |
| `timeoutIntervalForResource` | 300s | Total resource timeout |
| `httpMaximumConnectionsPerHost` | 6 | Connection pool limit |
| `requestCachePolicy` | `.reloadIgnoringLocalCacheData` | No API caching |
| `tlsMinimumSupportedProtocolVersion` | TLS 1.2 | Minimum TLS version |
| `waitsForConnectivity` | true | Wait for network |
| `httpShouldSetCookies` | false | No cookie management |

---

## APIClient

```swift
// APIClient Protocol
protocol APIClientProtocol: Sendable {
    func request<T: Decodable>(_ endpoint: Endpoint) async throws -> T
    func request(_ endpoint: Endpoint) async throws -> (Data, HTTPURLResponse)
    func upload<T: Decodable>(_ endpoint: Endpoint, data: Data) async throws -> T
}

// APIClient Implementation
final class APIClient: APIClientProtocol, @unchecked Sendable {
    private let session: URLSession
    private let interceptors: [RequestInterceptor]
    private let decoder: JSONDecoder
    private let logger: AppLogger

    init(
        session: URLSession = .apiSession,
        interceptors: [RequestInterceptor] = [],
        logger: AppLogger = AppLogger(category: .network)
    ) {
        self.session = session
        self.interceptors = interceptors
        self.decoder = JSONDecoder()
        self.decoder.keyDecodingStrategy = .convertFromSnakeCase
        self.decoder.dateDecodingStrategy = .iso8601
        self.logger = logger
    }

    // MARK: - Async/Await Request
    func request<T: Decodable>(_ endpoint: Endpoint) async throws -> T {
        var urlRequest = try buildRequest(from: endpoint)

        // Apply interceptors
        for interceptor in interceptors {
            try await interceptor.intercept(&urlRequest)
        }

        logger.info("\(endpoint.method.rawValue) \(endpoint.url)")

        let startTime = CFAbsoluteTimeGetCurrent()

        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await session.data(for: urlRequest)
        } catch let error as URLError {
            throw mapURLError(error)
        }

        let duration = (CFAbsoluteTimeGetCurrent() - startTime) * 1000
        logger.info("Response: \(response.statusCode) (\(String(format: "%.0f", duration))ms)")

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        // Handle interceptors response
        for interceptor in interceptors {
            try await interceptor.didReceive(data, response: httpResponse)
        }

        return try handleResponse(data, response: httpResponse)
    }

    // MARK: - Raw Request
    func request(_ endpoint: Endpoint) async throws -> (Data, HTTPURLResponse) {
        var urlRequest = try buildRequest(from: endpoint)

        for interceptor in interceptors {
            try await interceptor.intercept(&urlRequest)
        }

        let (data, response) = try await session.data(for: urlRequest)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        return (data, httpResponse)
    }

    // MARK: - Upload Request
    func upload<T: Decodable>(_ endpoint: Endpoint, data: Data) async throws -> T {
        var urlRequest = try buildRequest(from: endpoint)
        urlRequest.httpBody = data
        urlRequest.setValue("multipart/form-data", forHTTPHeaderField: "Content-Type")

        for interceptor in interceptors {
            try await interceptor.intercept(&urlRequest)
        }

        let (responseData, response) = try await session.data(for: urlRequest)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        return try handleResponse(responseData, response: httpResponse)
    }

    // MARK: - Build Request
    private func buildRequest(from endpoint: Endpoint) throws -> URLRequest {
        var components = URLComponents(url: endpoint.baseURL.appendingPathComponent(endpoint.path), resolvingAgainstBaseURL: false)

        if !endpoint.queryItems.isEmpty {
            components?.queryItems = endpoint.queryItems
        }

        guard let url = components?.url else {
            throw APIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = endpoint.method.rawValue
        request.timeoutInterval = endpoint.timeout

        for (key, value) in endpoint.headers {
            request.setValue(value, forHTTPHeaderField: key)
        }

        if let body = endpoint.body {
            request.httpBody = body
        }

        return request
    }

    // MARK: - Handle Response
    private func handleResponse<T: Decodable>(_ data: Data, response: HTTPURLResponse) throws -> T {
        switch response.statusCode {
        case 200...299:
            do {
                return try decoder.decode(T.self, from: data)
            } catch {
                logger.error("Decoding failed: \(error.localizedDescription)")
                throw APIError.decodingError(error)
            }
        case 401:
            throw APIError.unauthorized
        case 403:
            throw APIError.forbidden
        case 404:
            throw APIError.notFound
        case 409:
            throw APIError.conflict
        case 422:
            throw APIError.validationError(data: data)
        case 429:
            let retryAfter = response.value(forHTTPHeaderField: "Retry-After")
            throw APIError.rateLimited(retryAfter: retryAfter)
        case 500...599:
            throw APIError.serverError(statusCode: response.statusCode)
        default:
            throw APIError.httpError(statusCode: response.statusCode, data: data)
        }
    }

    // MARK: - Map URL Error
    private func mapURLError(_ error: URLError) -> APIError {
        switch error.code {
        case .notConnectedToInternet:
            return APIError.noConnection
        case .timedOut:
            return APIError.timeout
        case .cannotFindHost, .cannotConnectToHost:
            return APIError.hostUnreachable
        case .networkConnectionLost:
            return APIError.connectionLost
        case .secureConnectionFailed:
            return APIError.tlsError
        default:
            return APIError.urlError(error)
        }
    }
}
```

### APIClient Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      APIClient Flow                          │
│                                                             │
│  ┌──────────┐     ┌──────────────┐     ┌────────────────┐ │
│  │ Endpoint │────▶│ Build Request │────▶│ Interceptors   │ │
│  │          │     │              │     │ - Auth         │ │
│  └──────────┘     └──────────────┘     │ - Logging      │ │
│                                        │ - Rate Limit   │ │
│                                        └───────┬────────┘ │
│                                                │          │
│                                                ▼          │
│                                        ┌────────────────┐ │
│                                        │  URLSession    │ │
│                                        │  .data(for:)   │ │
│                                        └───────┬────────┘ │
│                                                │          │
│                                                ▼          │
│                                        ┌────────────────┐ │
│                                        │ Handle Response│ │
│                                        │ - Status Code  │ │
│                                        │ - Decode JSON  │ │
│                                        │ - Error Map    │ │
│                                        └───────┬────────┘ │
│                                                │          │
│  ┌──────────┐     ┌──────────────┐     ┌───────▼────────┐ │
│  │ Response │◀────│  Interceptor │◀────│  HTTP Response  │ │
│  │ Model    │     │  Response    │     │                 │ │
│  └──────────┘     └──────────────┘     └────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## Authentication Interceptor

```swift
// Request Interceptor Protocol
protocol RequestInterceptor: Sendable {
    func intercept(_ request: inout URLRequest) async throws
    func didReceive(_ data: Data, response: HTTPURLResponse) async throws
}

// Auth Interceptor
final class AuthInterceptor: RequestInterceptor {
    private let tokenProvider: TokenProviderProtocol
    private let logger: AppLogger

    init(
        tokenProvider: TokenProviderProtocol = KeychainTokenProvider(),
        logger: AppLogger = AppLogger(category: .auth)
    ) {
        self.tokenProvider = tokenProvider
        self.logger = logger
    }

    func intercept(_ request: inout URLRequest) async throws {
        // Skip auth for public endpoints
        guard let endpoint = request.url?.absoluteString,
              !endpoint.contains("/auth/login"),
              !endpoint.contains("/auth/register"),
              !endpoint.contains("/auth/refresh") else {
            return
        }

        guard let tokens = try tokenProvider.getTokens() else {
            throw APIError.unauthorized
        }

        // Check if token needs refresh
        if tokens.needsRefresh {
            logger.info("Token needs refresh, refreshing...")
            let newTokens = try await tokenProvider.refreshToken()
            request.setValue(
                "Bearer \(newTokens.accessToken)",
                forHTTPHeaderField: "Authorization"
            )
        } else {
            request.setValue(
                "Bearer \(tokens.accessToken)",
                forHTTPHeaderField: "Authorization"
            )
        }
    }

    func didReceive(_ data: Data, response: HTTPURLResponse) async throws {
        // Handle 401 responses
        if response.statusCode == 401 {
            logger.error("Received 401, refreshing token...")
            try await tokenProvider.refreshToken()
            throw APIError.tokenExpired
        }
    }
}
```

```swift
// Token Provider
protocol TokenProviderProtocol: Sendable {
    func getTokens() async throws -> AuthTokens?
    func refreshToken() async throws -> AuthTokens
    func saveTokens(_ tokens: AuthTokens) async throws
    func clearTokens() async throws
}

final class KeychainTokenProvider: TokenProviderProtocol {
    private let keychain = KeychainService.shared
    private let apiClient: APIClientProtocol

    init(apiClient: APIClientProtocol = APIClient()) {
        self.apiClient = apiClient
    }

    func getTokens() async throws -> AuthTokens? {
        guard let data = try keychain.load(for: "auth_tokens") else { return nil }
        return try JSONDecoder().decode(AuthTokens.self, from: data)
    }

    func refreshToken() async throws -> AuthTokens {
        guard let tokens = try await getTokens() else {
            throw APIError.unauthorized
        }

        let newTokens: AuthTokensResponse = try await apiClient.request(
            AuthEndpoint.refreshToken(token: tokens.refreshToken)
        )

        let domainTokens = AuthTokens(
            accessToken: newTokens.accessToken,
            refreshToken: newTokens.refreshToken,
            expiresIn: newTokens.expiresIn,
            tokenType: newTokens.tokenType
        )

        try await saveTokens(domainTokens)
        return domainTokens
    }

    func saveTokens(_ tokens: AuthTokens) async throws {
        let data = try JSONEncoder().encode(tokens)
        try keychain.save(data, for: "auth_tokens")
    }

    func clearTokens() async throws {
        try keychain.delete(for: "auth_tokens")
    }
}
```

### Interceptor Chain

```
┌─────────────────────────────────────────────────────────────┐
│                   INTERCEPTOR CHAIN                          │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                  OUTBOUND (Request)                  │   │
│  │                                                     │   │
│  │  Request ──▶ [Logging] ──▶ [Auth] ──▶ [RateLimit]  │   │
│  │                   │           │          │          │   │
│  │                   ▼           ▼          ▼          │   │
│  │              Log Request  Attach Token  Check Rate  │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                  INBOUND (Response)                  │   │
│  │                                                     │   │
│  │  Response ──▶ [RateLimit] ──▶ [Auth] ──▶ [Logging] │   │
│  │                   │           │          │          │   │
│  │                   ▼           ▼          ▼          │   │
│  │            Update Quota   Handle 401  Log Response  │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## Token Refresh Interceptor

```swift
// Token Refresh Interceptor with Queue Management
final class TokenRefreshInterceptor: RequestInterceptor {
    private let tokenProvider: TokenProviderProtocol
    private let logger: AppLogger

    // Queue management
    private var isRefreshing = false
    private var pendingRequests: [CheckedContinuation<AuthTokens, Error>] = []
    private let lock = NSLock()

    init(
        tokenProvider: TokenProviderProtocol = KeychainTokenProvider(),
        logger: AppLogger = AppLogger(category: .auth)
    ) {
        self.tokenProvider = tokenProvider
        self.logger = logger
    }

    func intercept(_ request: inout URLRequest) async throws {
        guard let tokens = try tokenProvider.getTokens() else {
            throw APIError.unauthorized
        }

        if tokens.needsRefresh {
            let newTokens = try await refreshIfNeeded()
            request.setValue(
                "Bearer \(newTokens.accessToken)",
                forHTTPHeaderField: "Authorization"
            )
        } else {
            request.setValue(
                "Bearer \(tokens.accessToken)",
                forHTTPHeaderField: "Authorization"
            )
        }
    }

    func didReceive(_ data: Data, response: HTTPURLResponse) async throws {
        if response.statusCode == 401 {
            _ = try await refreshIfNeeded()
            throw APIError.tokenRefreshed
        }
    }

    // MARK: - Token Refresh with Queue
    private func refreshIfNeeded() async throws -> AuthTokens {
        lock.lock()

        if isRefreshing {
            // Queue the request to wait for refresh
            lock.unlock()
            return try await withCheckedThrowingContinuation { continuation in
                lock.lock()
                pendingRequests.append(continuation)
                lock.unlock()
            }
        }

        isRefreshing = true
        lock.unlock()

        do {
            let newTokens = try await tokenProvider.refreshToken()
            logger.info("Token refreshed successfully")

            // Resume all pending requests
            lock.lock()
            for continuation in pendingRequests {
                continuation.resume(returning: newTokens)
            }
            pendingRequests.removeAll()
            isRefreshing = false
            lock.unlock()

            return newTokens
        } catch {
            // Resume all pending requests with error
            lock.lock()
            for continuation in pendingRequests {
                continuation.resume(throwing: error)
            }
            pendingRequests.removeAll()
            isRefreshing = false
            lock.unlock()

            throw error
        }
    }
}
```

### Token Refresh Flow

```
┌─────────────────────────────────────────────────────────────┐
│                TOKEN REFRESH FLOW                            │
│                                                             │
│  Request 1 (token expired)                                  │
│  ┌──────────┐                                               │
│  │ API Call │──▶ 401 Response ──┐                           │
│  └──────────┘                   │                           │
│                                 ▼                           │
│                         ┌───────────────┐                   │
│                         │ Is Refreshing?│                   │
│                         └───────┬───────┘                   │
│                                 │                           │
│                    ┌────────────┴────────────┐              │
│                    │ YES                     │ NO           │
│                    ▼                         ▼              │
│          ┌──────────────┐        ┌──────────────────┐      │
│          │ Queue Request│        │ Start Refresh    │      │
│          └──────┬───────┘        └────────┬─────────┘      │
│                 │                         │                 │
│                 │                  ┌──────▼───────┐        │
│                  │                 │ POST /refresh │        │
│                  │                 └──────┬───────┘        │
│                  │                        │                 │
│                  │                 ┌──────▼───────┐        │
│                  │                 │ Save Tokens  │        │
│                  │                 └──────┬───────┘        │
│                  │                        │                 │
│                  │                 ┌──────▼───────┐        │
│                  │                 │ Resume All   │        │
│                  │                 │ Queued       │        │
│                  │                 └──────┬───────┘        │
│                  │                        │                 │
│                  ▼                        ▼                 │
│          ┌──────────────────────────────────────┐          │
│          │        Retry Original Request         │          │
│          └──────────────────────────────────────┘          │
│                                                             │
│  Request 2 (queued during refresh)                          │
│  ┌──────────┐                                               │
│  │ API Call │──▶ Wait for Refresh ──▶ Retry with New Token  │
│  └──────────┘                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Certificate Pinning

```swift
// Certificate Pinning using URLSessionDelegate
final class CertificatePinner: NSObject, URLSessionDelegate {
    private let pinnedCertificates: [SecCertificate]
    private let logger: AppLogger

    init(certificates: [SecCertificate] = [], logger: AppLogger = AppLogger(category: .network)) {
        self.pinnedCertificates = certificates
        self.logger = logger
        super.init()
    }

    func urlSession(
        _ session: URLSession,
        didReceive challenge: URLAuthenticationChallenge
    ) async -> (URLSession.AuthChallengeDisposition, URLCredential?) {
        guard challenge.protectionSpace.authenticationMethod == NSURLAuthenticationMethodServerTrust,
              let serverTrust = challenge.protectionSpace.serverTrust else {
            return (.performDefaultHandling, nil)
        }

        // Validate certificate chain
        let policies = [SecPolicyCreateSSL(true, challenge.protectionSpace.host as CFString)]
        SecTrustSetPolicies(serverTrust, policies as CFTypeRef)

        var error: CFError?
        guard SecTrustEvaluateWithError(serverTrust, &error) else {
            logger.error("Certificate validation failed: \(error?.localizedDescription ?? "unknown")")
            return (.cancelAuthenticationChallenge, nil)
        }

        // Check pinned certificates
        if !pinnedCertificates.isEmpty {
            guard let serverCertificate = SecTrustGetCertificateAtIndex(serverTrust, 0) else {
                return (.cancelAuthenticationChallenge, nil)
            }

            let serverCertData = SecCertificateCopyData(serverCertificate) as Data
            let isPinned = pinnedCertificates.contains { pinnedCert in
                let pinnedData = SecCertificateCopyData(pinnedCert) as Data
                return serverCertData == pinnedData
            }

            if !isPinned {
                logger.error("Certificate not pinned")
                return (.cancelAuthenticationChallenge, nil)
            }
        }

        return (.useCredential, URLCredential(trust: serverTrust))
    }
}
```

```swift
// TrustKit-based Certificate Pinning
import TrustKit

class TrustKitPinner {
    static let shared = TrustKitPinner()

    func configure() {
        let config: [String: Any] = [
            kTSKPinnedDomains: [
                "api.nexusai.com": [
                    kTSKEnforcePinning: true,
                    kTSKIncludeSubdomains: true,
                    kTSKPinnedLeafCertificateSHA256: [
                        "base64_encoded_sha256_of_leaf_cert_1",
                        "base64_encoded_sha256_of_leaf_cert_2"
                    ],
                    kTSKPublicKeyHashes: [
                        "base64_encoded_sha256_of_public_key_1",
                        "base64_encoded_sha256_of_public_key_2"
                    ],
                    kTSKReportUris: ["https://report.nexusai.com/pinning"]
                ],
                "ws.nexusai.com": [
                    kTSKEnforcePinning: true,
                    kTSKIncludeSubdomains: true,
                    kTSKPinnedLeafCertificateSHA256: [
                        "base64_encoded_sha256_of_leaf_cert"
                    ]
                ]
            ]
        ]

        TrustKit.initSharedInstance(withConfiguration: config)
    }
}
```

### Certificate Pinning Flow

```
┌─────────────────────────────────────────────────────────────┐
│              CERTIFICATE PINNING FLOW                        │
│                                                             │
│  ┌──────────┐     ┌──────────────┐     ┌────────────────┐ │
│  │ HTTPS    │────▶│ TLS Handshake│────▶│ Receive Cert   │ │
│  │ Request  │     │              │     │                │ │
│  └──────────┘     └──────────────┘     └───────┬────────┘ │
│                                                │          │
│                                                ▼          │
│                                        ┌────────────────┐ │
│                                        │ Validate Chain │ │
│                                        │ - Trust Store  │ │
│                                        │ - Expiry       │ │
│                                        │ - Revocation   │ │
│                                        └───────┬────────┘ │
│                                                │          │
│                                    ┌───────────┴───────┐  │
│                                    │ Valid             │  │
│                                    └───────────┬───────┘  │
│                                                │          │
│                                    ┌───────────┴───────┐  │
│                                    │ Check Pinned      │  │
│                                    │ Certificates      │  │
│                                    └───────────┬───────┘  │
│                                                │          │
│                                    ┌───────────┴───────┐  │
│                                    │ Match Found?      │  │
│                                    └───────────┬───────┘  │
│                                                │          │
│                                    ┌───────────┴───────┐  │
│                                    │ YES    │    NO    │  │
│                                    ▼        │         ▼  │
│                              ┌──────────┐   │  ┌────────┐│
│                              │ Allow    │   │  │ Block  ││
│                              │ Request  │   │  │ Request││
│                              └──────────┘   │  └────────┘│
└─────────────────────────────────────────────────────────────┘
```

---

## Network Security Configuration

```xml
<!-- NetworkSecurityConfig.xml -->
<network-security-config>
    <!-- Debug configuration -->
    <domain-config cleartextTrafficPermitted="true">
        <domain includeSubdomains="true">localhost</domain>
        <domain includeSubdomains="true">127.0.0.1</domain>
        <domain includeSubdomains="true">10.0.2.2</domain>
        <trust-anchors>
            <certificates src="system" />
        </trust-anchors>
    </domain-config>

    <!-- Production configuration -->
    <domain-config>
        <domain includeSubdomains="true">api.nexusai.com</domain>
        <domain includeSubdomains="true">ws.nexusai.com</domain>
        <pin-set expiration="2025-01-01">
            <pin digest="SHA-256">
                base64_encoded_primary_key_hash
            </pin>
            <pin digest="SHA-256">
                base64_encoded_backup_key_hash
            </pin>
        </pin-set>
        <trust-anchors>
            <certificates src="system" />
            <certificates src="user" />
        </trust-anchors>
    </domain-config>
</network-security-config>
```

```swift
// App Transport Security exceptions in Info.plist
/*
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSExceptionDomains</key>
    <dict>
        <key>localhost</key>
        <dict>
            <key>NSExceptionAllowsInsecureHTTPLoads</key>
            <true/>
            <key>NSIncludesSubdomains</key>
            <true/>
        </dict>
    </dict>
</dict>
*/
```

---

## WebSocket Client

```swift
// WebSocket Client Protocol
protocol WebSocketClientProtocol: Sendable {
    var isConnected: Bool { get }
    func connect(token: String) async throws
    func disconnect()
    func send(_ message: WebSocketMessage)
    func onTokenReceived(_ handler: @escaping @Sendable (String) -> Void)
    func onCompleted(_ handler: @escaping @Sendable (ChatResponse) -> Void)
    func onToolCall(_ handler: @escaping @Sendable (ToolCall) -> Void)
    func onThinking(_ handler: @escaping @Sendable (String) -> Void)
    func onError(_ handler: @escaping @Sendable (Error) -> Void)
    func onConnectionStateChange(_ handler: @escaping @Sendable (WebSocketState) -> Void)
}

// WebSocket State
enum WebSocketState: Equatable {
    case disconnected
    case connecting
    case connected
    case reconnecting(attempt: Int)
    case failed(Error)

    static func == (lhs: WebSocketState, rhs: WebSocketState) -> Bool {
        switch (lhs, rhs) {
        case (.disconnected, .disconnected): return true
        case (.connecting, .connecting): return true
        case (.connected, .connected): return true
        case (.reconnecting(let a), .reconnecting(let b)): return a == b
        case (.failed, .failed): return true
        default: return false
        }
    }
}

// WebSocket Client Implementation
final class WebSocketClient: WebSocketClientProtocol {
    private var webSocketTask: URLSessionWebSocketTask?
    private var session: URLSession
    private var reconnectTimer: Timer?
    private var heartbeatTimer: Timer?
    private var token: String?

    private var tokenHandlers: [@Sendable (String) -> Void] = []
    private var completedHandlers: [@Sendable (ChatResponse) -> Void] = []
    private var toolCallHandlers: [@Sendable (ToolCall) -> Void] = []
    private var thinkingHandlers: [@Sendable (String) -> Void] = []
    private var errorHandlers: [@Sendable (Error) -> Void] = []
    private var stateHandlers: [@Sendable (WebSocketState) -> Void] = []

    private var state: WebSocketState = .disconnected {
        didSet {
            for handler in stateHandlers {
                handler(state)
            }
        }
    }

    private let reconnectConfig = WebSocketReconnectConfig()
    private var currentAttempt = 0
    private let logger: AppLogger

    var isConnected: Bool {
        if case .connected = state { return true }
        return false
    }

    init(
        session: URLSession = .webSocketSession,
        logger: AppLogger = AppLogger(category: .chat)
    ) {
        self.session = session
        self.logger = logger
    }

    // MARK: - Connect
    func connect(token: String) async throws {
        self.token = token
        self.currentAttempt = 0

        var components = URLComponents()
        components.scheme = "wss"
        components.host = "ws.nexusai.com"
        components.path = "/chat"
        components.queryItems = [URLQueryItem(name: "token", value: token)]

        guard let url = components.url else {
            throw WebSocketError.invalidURL
        }

        state = .connecting
        logger.info("Connecting to WebSocket: \(url.host ?? "")")

        webSocketTask = session.webSocketTask(with: url)
        webSocketTask?.resume()

        // Start receiving messages
        receiveMessages()

        // Wait for connection
        try await Task.sleep(for: .seconds(3))

        if isConnected {
            state = .connected
            startHeartbeat()
            logger.info("WebSocket connected")
        } else {
            throw WebSocketError.connectionFailed
        }
    }

    // MARK: - Disconnect
    func disconnect() {
        stopHeartbeat()
        stopReconnect()

        webSocketTask?.cancel(with: .goingAway, reason: nil)
        webSocketTask = nil
        state = .disconnected

        logger.info("WebSocket disconnected")
    }

    // MARK: - Send
    func send(_ message: WebSocketMessage) {
        guard let data = try? JSONEncoder().encode(message),
              let jsonString = String(data: data, encoding: .utf8) else {
            logger.error("Failed to encode WebSocket message")
            return
        }

        let wsMessage = URLSessionWebSocketTask.Message.string(jsonString)

        webSocketTask?.send(wsMessage) { [weak self] error in
            if let error = error {
                self?.logger.error("Failed to send WebSocket message: \(error.localizedDescription)")
                self?.handleError(error)
            } else {
                self?.logger.debug("Sent WebSocket message: \(message)")
            }
        }
    }

    // MARK: - Receive Messages
    private func receiveMessages() {
        webSocketTask?.receive { [weak self] result in
            guard let self = self else { return }

            switch result {
            case .success(let message):
                self.handleMessage(message)
                self.receiveMessages() // Continue receiving
            case .failure(let error):
                self.logger.error("WebSocket receive error: \(error.localizedDescription)")
                self.handleError(error)
            }
        }
    }

    // MARK: - Handle Message
    private func handleMessage(_ message: URLSessionWebSocketTask.Message) {
        switch message {
        case .string(let text):
            guard let data = text.data(using: .utf8),
                  let payload = try? JSONDecoder().decode(WebSocketPayload.self, from: data) else {
                logger.error("Failed to parse WebSocket message")
                return
            }

            DispatchQueue.main.async { [weak self] in
                self?.processPayload(payload)
            }

        case .data(let data):
            logger.debug("Received binary data: \(data.count) bytes")

        @unknown default:
            break
        }
    }

    // MARK: - Process Payload
    private func processPayload(_ payload: WebSocketPayload) {
        switch payload.type {
        case .token:
            if let token = payload.data?["text"] as? String {
                for handler in tokenHandlers {
                    handler(token)
                }
            }

        case .completed:
            if let response = payload.data {
                let chatResponse = ChatResponse(
                    conversationId: response["conversationId"] as? String ?? "",
                    messageId: response["messageId"] as? String ?? "",
                    content: response["content"] as? String ?? "",
                    toolCalls: []
                )
                for handler in completedHandlers {
                    handler(chatResponse)
                }
            }

        case .toolCall:
            if let toolData = payload.data {
                let toolCall = ToolCall(
                    id: toolData["id"] as? String ?? UUID().uuidString,
                    name: toolData["name"] as? String ?? "",
                    arguments: [:],
                    status: .pending
                )
                for handler in toolCallHandlers {
                    handler(toolCall)
                }
            }

        case .toolResult:
            // Handle tool result
            break

        case .thinking:
            if let text = payload.data?["text"] as? String {
                for handler in thinkingHandlers {
                    handler(text)
                }
            }

        case .error:
            if let errorMessage = payload.data?["message"] as? String {
                let error = WebSocketError.serverError(errorMessage)
                for handler in errorHandlers {
                    handler(error)
                }
            }
        }
    }

    // MARK: - Error Handling
    private func handleError(_ error: Error) {
        state = .failed(error)

        for handler in errorHandlers {
            handler(error)
        }

        // Auto-reconnect
        if let wsError = error as? WebSocketError, wsError.shouldReconnect {
            attemptReconnect()
        }
    }

    // MARK: - Heartbeat
    private func startHeartbeat() {
        heartbeatTimer = Timer.scheduledTimer(withTimeInterval: 30, repeats: true) { [weak self] _ in
            self?.send(.heartbeat)
        }
    }

    private func stopHeartbeat() {
        heartbeatTimer?.invalidate()
        heartbeatTimer = nil
    }

    // MARK: - Reconnect
    private func attemptReconnect() {
        guard currentAttempt < reconnectConfig.maxRetries else {
            logger.error("Max reconnection attempts reached")
            state = .failed(WebSocketError.maxRetriesExceeded)
            return
        }

        currentAttempt += 1
        state = .reconnecting(attempt: currentAttempt)

        let delay = reconnectConfig.delay(for: currentAttempt)
        logger.info("Reconnecting in \(delay)s (attempt \(currentAttempt))")

        reconnectTimer = Timer.scheduledTimer(withTimeInterval: delay, repeats: false) { [weak self] _ in
            guard let self = self, let token = self.token else { return }

            Task {
                do {
                    try await self.connect(token: token)
                } catch {
                    self.logger.error("Reconnection failed: \(error.localizedDescription)")
                    self.attemptReconnect()
                }
            }
        }
    }

    private func stopReconnect() {
        reconnectTimer?.invalidate()
        reconnectTimer = nil
        currentAttempt = 0
    }

    // MARK: - Handler Registration
    func onTokenReceived(_ handler: @escaping @Sendable (String) -> Void) {
        tokenHandlers.append(handler)
    }

    func onCompleted(_ handler: @escaping @Sendable (ChatResponse) -> Void) {
        completedHandlers.append(handler)
    }

    func onToolCall(_ handler: @escaping @Sendable (ToolCall) -> Void) {
        toolCallHandlers.append(handler)
    }

    func onThinking(_ handler: @escaping @Sendable (String) -> Void) {
        thinkingHandlers.append(handler)
    }

    func onError(_ handler: @escaping @Sendable (Error) -> Void) {
        errorHandlers.append(handler)
    }

    func onConnectionStateChange(_ handler: @escaping @Sendable (WebSocketState) -> Void) {
        stateHandlers.append(handler)
    }
}
```

---

## WebSocket Message Protocol

```swift
// WebSocket Message Types
enum WebSocketMessageType: String, Codable {
    case chat
    case heartbeat
    case pong
    case cancel
    case ping
}

// WebSocket Outgoing Message
enum WebSocketMessage: Codable {
    case chat(conversationId: String, message: String, agentId: String?, attachments: [Attachment])
    case heartbeat
    case pong
    case cancel(messageId: String)

    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)

        switch self {
        case .chat(let conversationId, let message, let agentId, let attachments):
            try container.encode("chat", forKey: .type)
            try container.encode(conversationId, forKey: .conversationId)
            try container.encode(message, forKey: .message)
            if let agentId = agentId {
                try container.encode(agentId, forKey: .agentId)
            }
            try container.encode(attachments, forKey: .attachments)

        case .heartbeat:
            try container.encode("heartbeat", forKey: .type)

        case .pong:
            try container.encode("pong", forKey: .type)

        case .cancel(let messageId):
            try container.encode("cancel", forKey: .type)
            try container.encode(messageId, forKey: .messageId)
        }
    }

    enum CodingKeys: String, CodingKey {
        case type, conversationId, message, agentId, attachments, messageId
    }
}

// WebSocket Incoming Payload
struct WebSocketPayload: Codable {
    let type: PayloadType
    let data: [String: Any]?

    enum PayloadType: String, Codable {
        case token
        case completed
        case toolCall = "tool_call"
        case toolResult = "tool_result"
        case thinking
        case error
        case ping
        case pong
    }
}

// Chat Response from WebSocket
struct ChatResponse: Codable {
    let conversationId: String
    let messageId: String
    let content: String
    let toolCalls: [ToolCall]
}
```

### WebSocket Message Flow

```
┌─────────────────────────────────────────────────────────────┐
│                WEBSOCKET MESSAGE FLOW                        │
│                                                             │
│  Client                          Server                     │
│  ┌──────────┐                    ┌──────────┐              │
│  │ Connect  │─── WS Connect ────▶│ Accept   │              │
│  │ w/token  │◀── Connected ──────│          │              │
│  └──────────┘                    └──────────┘              │
│                                                             │
│  ┌──────────┐                    ┌──────────┐              │
│  │ Auth     │─── Auth Message ──▶│ Validate │              │
│  │ Message  │◀── Auth OK ────────│ Token    │              │
│  └──────────┘                    └──────────┘              │
│                                                             │
│  ┌──────────┐                    ┌──────────┐              │
│  │ Chat     │─── Chat Message ──▶│ Process  │              │
│  │ Request  │                    │          │              │
│  └──────────┘                    └────┬─────┘              │
│                                       │                    │
│  ┌──────────┐                    ┌────▼─────┐              │
│  │ Receive  │◀── Thinking ──────│ Generate │              │
│  │ Thinking │                    │          │              │
│  └──────────┘                    │          │              │
│                                   │          │              │
│  ┌──────────┐                    │          │              │
│  │ Receive  │◀── Tool Call ─────│ Call     │              │
│  │ Tool Call│                    │ Tool     │              │
│  └──────────┘                    │          │              │
│                                   │          │              │
│  ┌──────────┐                    │          │              │
│  │ Send     │─── Tool Result ──▶│ Process  │              │
│  │ Result   │                    │ Result   │              │
│  └──────────┘                    │          │              │
│                                   │          │              │
│  ┌──────────┐                    │          │              │
│  │ Receive  │◀── Stream ────────│ Stream   │              │
│  │ Tokens   │◀── Tokens ────────│ Response │              │
│  └──────────┘                    │          │              │
│                                   │          │              │
│  ┌──────────┐                    │          │              │
│  │ Receive  │◀── Completed ─────│ Complete │              │
│  │ Complete │                    │          │              │
│  └──────────┘                    └──────────┘              │
│                                                             │
│  ┌──────────┐                    ┌──────────┐              │
│  │ Heartbeat│─── Ping ──────────▶│ Pong     │              │
│  │ (30s)    │◀── Pong ──────────│          │              │
│  └──────────┘                    └──────────┘              │
└─────────────────────────────────────────────────────────────┘
```

---

## WebSocket Authentication

```swift
// WebSocket Authentication
extension WebSocketClient {
    func authenticate() async throws {
        guard let token = self.token else {
            throw WebSocketError.notAuthenticated
        }

        // Method 1: Query parameter (already set during connect)
        // wss://ws.nexusai.com/chat?token=<jwt>

        // Method 2: First message authentication
        let authMessage: [String: Any] = [
            "type": "auth",
            "token": token,
            "client_id": UIDevice.current.identifierForVendor?.uuidString ?? UUID().uuidString,
            "platform": "ios",
            "version": Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "unknown"
        ]

        if let data = try? JSONSerialization.data(withJSONObject: authMessage),
           let jsonString = String(data: data, encoding: .utf8) {
            webSocketTask?.send(.string(jsonString)) { error in
                if let error = error {
                    self.logger.error("Auth message failed: \(error.localizedDescription)")
                }
            }
        }
    }
}
```

---

## WebSocket Reconnection

```swift
// Reconnection Configuration
struct WebSocketReconnectConfig {
    let maxRetries: Int = 10
    let baseDelay: TimeInterval = 1.0
    let maxDelay: TimeInterval = 60.0
    let backoffMultiplier: Double = 2.0
    let jitterRange: ClosedRange<Double> = 0...0.5

    func delay(for attempt: Int) -> TimeInterval {
        let exponentialDelay = baseDelay * pow(backoffMultiplier, Double(attempt - 1))
        let clampedDelay = min(exponentialDelay, maxDelay)
        let jitter = Double.random(in: jitterRange)
        return clampedDelay + jitter
    }
}

// Reconnection State Machine
/*
┌─────────────────────────────────────────────────────────────┐
│            WEBSOCKET RECONNECTION STATE MACHINE              │
│                                                             │
│  ┌───────────┐                                              │
│  │Connected   │◀──────────────────────────────────────┐    │
│  └─────┬─────┘                                         │    │
│        │ Error                                         │    │
│        ▼                                               │    │
│  ┌───────────┐                                         │    │
│  │Disconnecting│                                        │    │
│  └─────┬─────┘                                         │    │
│        │                                               │    │
│        ▼                                               │    │
│  ┌───────────┐     ┌──────────────┐     ┌───────────┐│    │
│  │Reconnecting├────▶│ Attempt 1    ├────▶│ Attempt 2 ││    │
│  │(attempt: N)│     │ Delay: 1s    │     │ Delay: 2s ││    │
│  └───────────┘     └──────────────┘     └─────┬─────┘│    │
│                                               │      │    │
│                                               ▼      │    │
│                                         ┌───────────┐│    │
│                                         │ Attempt 3 ││    │
│                                         │ Delay: 4s ││    │
│                                         └─────┬─────┘│    │
│                                               │      │    │
│                                               ▼      │    │
│                                         ┌───────────┐│    │
│                                    ┌───▶│ Attempt N ││    │
│                                    │    │ Delay: 60s││    │
│                                    │    └─────┬─────┘│    │
│                                    │          │      │    │
│                                    │    Success     Failure │
│                                    │          │      │    │
│                                    ▼          ▼      ▼    │
│                              ┌──────────┐  ┌──────────┐  │
│                              │ Connected │  │ Failed   │  │
│                              └──────────┘  │ (Max)    │  │
│                                            └──────────┘  │
└─────────────────────────────────────────────────────────────┘
*/
```

---

## API Endpoint Definitions

```swift
// Endpoint Protocol
protocol Endpoint {
    var baseURL: URL { get }
    var path: String { get }
    var method: HTTPMethod { get }
    var headers: [String: String] { get }
    var queryItems: [URLQueryItem] { get }
    var body: Data? { get }
    var timeout: TimeInterval { get }
}

// HTTP Method
enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case patch = "PATCH"
    case delete = "DELETE"
}

// API Endpoints
enum APIEndpoint: Endpoint {
    // Auth
    case login(email: String, password: String)
    case register(email: String, password: String, name: String)
    case refreshToken(token: String)
    case logout

    // Chat
    case sendMessage(conversationId: String, message: String, agentId: String?)
    case getConversations(limit: Int, offset: Int)
    case getMessages(conversationId: String, limit: Int)
    case createConversation(title: String, agentId: String?)
    case deleteConversation(id: String)

    // Agents
    case getAgents
    case getAgent(id: String)
    case createAgent(name: String, systemPrompt: String, modelId: String)
    case updateAgent(id: String, name: String, systemPrompt: String)
    case deleteAgent(id: String)

    // Knowledge
    case getKnowledge
    case uploadKnowledge(data: Data, filename: String)
    case deleteKnowledge(id: String)

    // Models
    case getModels
    case getModel(id: String)

    var baseURL: URL {
        URL(string: BuildConfig.current.apiBaseURL)!
    }

    var path: String {
        switch self {
        case .login: return "/api/v1/auth/login"
        case .register: return "/api/v1/auth/register"
        case .refreshToken: return "/api/v1/auth/refresh"
        case .logout: return "/api/v1/auth/logout"

        case .sendMessage: return "/api/v1/chat/send"
        case .getConversations: return "/api/v1/conversations"
        case .getMessages(let id, _): return "/api/v1/conversations/\(id)/messages"
        case .createConversation: return "/api/v1/conversations"
        case .deleteConversation(let id): return "/api/v1/conversations/\(id)"

        case .getAgents: return "/api/v1/agents"
        case .getAgent(let id): return "/api/v1/agents/\(id)"
        case .createAgent: return "/api/v1/agents"
        case .updateAgent(let id, _, _): return "/api/v1/agents/\(id)"
        case .deleteAgent(let id): return "/api/v1/agents/\(id)"

        case .getKnowledge: return "/api/v1/knowledge"
        case .uploadKnowledge: return "/api/v1/knowledge"
        case .deleteKnowledge(let id): return "/api/v1/knowledge/\(id)"

        case .getModels: return "/api/v1/models"
        case .getModel(let id): return "/api/v1/models/\(id)"
        }
    }

    var method: HTTPMethod {
        switch self {
        case .login, .register, .refreshToken, .sendMessage, .createConversation,
             .createAgent, .uploadKnowledge:
            return .post
        case .getConversations, .getMessages, .getAgents, .getAgent,
             .getKnowledge, .getModels, .getModel:
            return .get
        case .updateAgent:
            return .patch
        case .logout, .deleteConversation, .deleteAgent, .deleteKnowledge:
            return .delete
        }
    }

    var headers: [String: String] {
        var headers = [
            "Content-Type": "application/json",
            "Accept": "application/json"
        ]

        switch self {
        case .uploadKnowledge:
            headers["Content-Type"] = "multipart/form-data"
        default:
            break
        }

        return headers
    }

    var queryItems: [URLQueryItem] {
        switch self {
        case .getConversations(let limit, let offset):
            return [
                URLQueryItem(name: "limit", value: "\(limit)"),
                URLQueryItem(name: "offset", value: "\(offset)")
            ]
        case .getMessages(_, let limit):
            return [URLQueryItem(name: "limit", value: "\(limit)")]
        default:
            return []
        }
    }

    var body: Data? {
        var params: [String: Any] = [:]

        switch self {
        case .login(let email, let password):
            params = ["email": email, "password": password]
        case .register(let email, let password, let name):
            params = ["email": email, "password": password, "name": name]
        case .refreshToken(let token):
            params = ["refresh_token": token]
        case .sendMessage(let conversationId, let message, let agentId):
            params = ["conversation_id": conversationId, "message": message]
            if let agentId = agentId {
                params["agent_id"] = agentId
            }
        case .createConversation(let title, let agentId):
            params = ["title": title]
            if let agentId = agentId {
                params["agent_id"] = agentId
            }
        case .createAgent(let name, let systemPrompt, let modelId):
            params = ["name": name, "system_prompt": systemPrompt, "model_id": modelId]
        case .updateAgent(_, let name, let systemPrompt):
            params = ["name": name, "system_prompt": systemPrompt]
        default:
            return nil
        }

        return try? JSONSerialization.data(withJSONObject: params)
    }

    var timeout: TimeInterval {
        switch self {
        case .uploadKnowledge:
            return 120
        case .login, .register:
            return 15
        default:
            return 30
        }
    }
}
```

### Endpoint Table

| Endpoint | Method | Path | Auth | Timeout |
|----------|--------|------|------|---------|
| Login | POST | /api/v1/auth/login | No | 15s |
| Register | POST | /api/v1/auth/register | No | 15s |
| Refresh Token | POST | /api/v1/auth/refresh | No | 10s |
| Logout | DELETE | /api/v1/auth/logout | Yes | 10s |
| Send Message | POST | /api/v1/chat/send | Yes | 30s |
| Get Conversations | GET | /api/v1/conversations | Yes | 15s |
| Get Messages | GET | /api/v1/conversations/{id}/messages | Yes | 15s |
| Create Conversation | POST | /api/v1/conversations | Yes | 15s |
| Delete Conversation | DELETE | /api/v1/conversations/{id} | Yes | 10s |
| Get Agents | GET | /api/v1/agents | Yes | 15s |
| Create Agent | POST | /api/v1/agents | Yes | 15s |
| Upload Knowledge | POST | /api/v1/knowledge | Yes | 120s |
| Get Models | GET | /api/v1/models | Yes | 15s |

---

## Request/Response Models

```swift
// Request Models
struct LoginRequest: Codable {
    let email: String
    let password: String
}

struct RegisterRequest: Codable {
    let email: String
    let password: String
    let name: String
}

struct SendMessageRequest: Codable {
    let conversationId: String
    let message: String
    let agentId: String?
    let attachments: [AttachmentRequest]?
}

struct AttachmentRequest: Codable {
    let filename: String
    let mimeType: String
    let data: String // Base64 encoded
}

// Response Models
struct AuthTokensResponse: Codable {
    let accessToken: String
    let refreshToken: String
    let expiresIn: TimeInterval
    let tokenType: String

    func toDomain() -> AuthTokens {
        AuthTokens(
            accessToken: accessToken,
            refreshToken: refreshToken,
            expiresIn: expiresIn,
            tokenType: tokenType
        )
    }
}

struct UserResponse: Codable {
    let id: String
    let email: String
    let name: String
    let avatarUrl: String?
    let tenantId: String
    let createdAt: String
    let updatedAt: String
    let biometricEnabled: Bool
    let kycStatus: String

    func toDomain() -> User {
        User(
            id: id,
            email: email,
            name: name,
            avatarUrl: avatarUrl,
            tenantId: tenantId,
            createdAt: ISO8601DateFormatter().date(from: createdAt) ?? Date(),
            updatedAt: ISO8601DateFormatter().date(from: updatedAt) ?? Date(),
            biometricEnabled: biometricEnabled,
            kycStatus: User.KYCStatus(rawValue: kycStatus) ?? .pending
        )
    }
}

struct MessageResponse: Codable {
    let id: String
    let content: String
    let role: String
    let timestamp: String
    let attachments: [AttachmentResponse]?
    let toolCalls: [ToolCallResponse]?

    func toDomain() -> Message {
        Message(
            id: UUID(uuidString: id) ?? UUID(),
            content: content,
            role: Message.MessageRole(rawValue: role) ?? .user,
            timestamp: ISO8601DateFormatter().date(from: timestamp) ?? Date(),
            attachments: attachments?.map { $0.toDomain() } ?? [],
            toolCalls: toolCalls?.map { $0.toDomain() } ?? [],
            isStreaming: false,
            tokens: []
        )
    }
}

struct AttachmentResponse: Codable {
    let id: String
    let name: String
    let type: String
    let url: String?
    let size: Int64

    func toDomain() -> Attachment {
        Attachment(
            id: UUID(uuidString: id) ?? UUID(),
            name: name,
            type: Attachment.AttachmentType(rawValue: type) ?? .document,
            url: url,
            data: nil,
            size: size
        )
    }
}

struct ToolCallResponse: Codable {
    let id: String
    let name: String
    let arguments: [String: String]
    let status: String
    let result: String?

    func toDomain() -> ToolCall {
        ToolCall(
            id: id,
            name: name,
            arguments: arguments.mapValues { AnyCodable($0) },
            status: ToolCall.Status(rawValue: status) ?? .pending,
            result: result
        )
    }
}
```

### Model Mapping Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                 MODEL MAPPING                                │
│                                                             │
│  ┌──────────────────┐         ┌──────────────────┐        │
│  │  Response Model   │  toDomain()  │  Domain Entity  │        │
│  │  (Codable)       │────────▶│  (Pure Swift)   │        │
│  └──────────────────┘         └──────────────────┘        │
│                                                             │
│  AuthTokensResponse ──────────▶ AuthTokens                  │
│  UserResponse ────────────────▶ User                        │
│  MessageResponse ─────────────▶ Message                     │
│  ConversationResponse ────────▶ Conversation                │
│  AgentResponse ───────────────▶ Agent                       │
│  AttachmentResponse ──────────▶ Attachment                  │
│  ToolCallResponse ────────────▶ ToolCall                    │
│                                                             │
│  ┌──────────────────┐         ┌──────────────────┐        │
│  │  Domain Entity    │  toRequest()  │  Request Model   │        │
│  │  (Pure Swift)    │────────▶│  (Codable)       │        │
│  └──────────────────┘         └──────────────────┘        │
│                                                             │
│  User ───────────────────────▶ UserRequest                  │
│  Message ────────────────────▶ SendMessageRequest           │
│  Conversation ───────────────▶ CreateConversationRequest    │
└─────────────────────────────────────────────────────────────┘
```

---

## Error Handling

```swift
// API Error Types
enum APIError: AppError, Equatable {
    case invalidURL
    case invalidResponse
    case httpError(statusCode: Int, data: Data?)
    case decodingError(Error)
    case rateLimited(retryAfter: String?)
    case timeout
    case noConnection
    case unauthorized
    case forbidden
    case notFound
    case conflict
    case validationError(data: Data)
    case serverError(statusCode: Int)
    case urlError(URLError)
    case tokenExpired
    case tokenRefreshed
    case hostUnreachable
    case connectionLost
    case tlsError

    var code: Int {
        switch self {
        case .invalidURL: return 1001
        case .invalidResponse: return 1002
        case .httpError(let code, _): return code
        case .decodingError: return 1003
        case .rateLimited: return 1029
        case .timeout: return 1004
        case .noConnection: return 1005
        case .unauthorized: return 1006
        case .forbidden: return 1007
        case .notFound: return 1008
        case .conflict: return 1009
        case .validationError: return 1010
        case .serverError: return 1011
        case .urlError: return 1012
        case .tokenExpired: return 1013
        case .tokenRefreshed: return 1014
        case .hostUnreachable: return 1015
        case .connectionLost: return 1016
        case .tlsError: return 1017
        }
    }

    var domain: String { "com.nexusai.api" }

    var errorDescription: String? {
        switch self {
        case .invalidURL: return "Invalid URL"
        case .invalidResponse: return "Invalid server response"
        case .httpError(let code, _): return "HTTP Error: \(code)"
        case .decodingError: return "Failed to decode response"
        case .rateLimited: return "Rate limited. Please try again later"
        case .timeout: return "Request timed out"
        case .noConnection: return "No internet connection"
        case .unauthorized: return "Please log in again"
        case .forbidden: return "Access denied"
        case .notFound: return "Resource not found"
        case .conflict: return "Resource conflict"
        case .validationError: return "Validation error"
        case .serverError(let code): return "Server error: \(code)"
        case .urlError(let error): return error.localizedDescription
        case .tokenExpired: return "Session expired"
        case .tokenRefreshed: return "Token refreshed"
        case .hostUnreachable: return "Server unreachable"
        case .connectionLost: return "Connection lost"
        case .tlsError: return "Secure connection failed"
        }
    }

    var shouldRetry: Bool {
        switch self {
        case .timeout, .connectionLost, .serverError:
            return true
        case .noConnection, .hostUnreachable:
            return false
        default:
            return false
        }
    }
}

// WebSocket Error Types
enum WebSocketError: AppError {
    case invalidURL
    case connectionFailed
    case notAuthenticated
    case serverError(String)
    case maxRetriesExceeded
    case messageTooLarge
    case invalidMessage

    var shouldReconnect: Bool {
        switch self {
        case .connectionFailed, .serverError:
            return true
        case .notAuthenticated, .maxRetriesExceeded:
            return false
        default:
            return false
        }
    }
}
```

### Error Handling Flow

```
┌─────────────────────────────────────────────────────────────┐
│                  ERROR HANDLING FLOW                         │
│                                                             │
│  ┌──────────┐     ┌──────────────┐     ┌────────────────┐ │
│  │ API Call │────▶│ Error Occurs │────▶│ Classify Error │ │
│  └──────────┘     └──────────────┘     └───────┬────────┘ │
│                                                │          │
│                    ┌───────────────────────────┴───┐      │
│                    │                               │      │
│              ┌─────▼─────┐  ┌─────────────┐  ┌───▼────┐ │
│              │ Client    │  │ Server      │  │Network │ │
│              │ Error     │  │ Error       │  │ Error  │ │
│              └─────┬─────┘  └──────┬──────┘  └───┬────┘ │
│                    │               │             │       │
│              ┌─────▼─────┐  ┌─────▼─────┐  ┌───▼─────┐│
│              │ Handle    │  │ Retry     │  │ Offline ││
│              │ Locally   │  │ Logic     │  │ Mode    ││
│              └─────┬─────┘  └─────┬─────┘  └───┬─────┘│
│                    │               │             │       │
│              ┌─────▼─────┐  ┌─────▼─────┐  ┌───▼─────┐│
│              │ Show Error│  │ Exponential│  │ Queue   ││
│              │ to User   │  │ Backoff   │  │ Request ││
│              └───────────┘  └───────────┘  └─────────┘│
│                                                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Offline Detection

```swift
// Network Monitor
import Network

final class NetworkMonitor: ObservableObject, NetworkMonitorProtocol {
    static let shared = NetworkMonitor()

    private let monitor = NWPathMonitor()
    private let queue = DispatchQueue(label: "NetworkMonitor")

    @Published var isConnected = true
    @Published var connectionType: ConnectionType = .unknown
    @Published var isExpensive = false

    enum ConnectionType {
        case wifi
        case cellular
        case ethernet
        case unknown
    }

    func startMonitoring() {
        monitor.pathUpdateHandler = { [weak self] path in
            DispatchQueue.main.async {
                self?.isConnected = path.status == .satisfied
                self?.isExpensive = path.isExpensive
                self?.connectionType = self?.getConnectionType(path) ?? .unknown

                if path.status == .satisfied {
                    self?.handleOnline()
                } else {
                    self?.handleOffline()
                }
            }
        }

        monitor.start(queue: queue)
    }

    func stopMonitoring() {
        monitor.cancel()
    }

    private func getConnectionType(_ path: NWPath) -> ConnectionType {
        if path.usesInterfaceType(.wifi) {
            return .wifi
        } else if path.usesInterfaceType(.cellular) {
            return .cellular
        } else if path.usesInterfaceType(.wiredEthernet) {
            return .ethernet
        }
        return .unknown
    }

    private func handleOnline() {
        os_log(.info, "Network: Connected via %{public}@", "\(connectionType)")
        NotificationCenter.default.post(name: .networkConnected, object: nil)
    }

    private func handleOffline() {
        os_log(.info, "Network: Disconnected")
        NotificationCenter.default.post(name: .networkDisconnected, object: nil)
    }
}

extension Notification.Name {
    static let networkConnected = Notification.Name("networkConnected")
    static let networkDisconnected = Notification.Name("networkDisconnected")
}
```

---

## Retry Logic

```swift
// Retry with Exponential Backoff
struct RetryPolicy {
    let maxRetries: Int
    let baseDelay: TimeInterval
    let maxDelay: TimeInterval
    let backoffMultiplier: Double
    let retryableErrors: [APIError]

    static let `default` = RetryPolicy(
        maxRetries: 3,
        baseDelay: 1.0,
        maxDelay: 30.0,
        backoffMultiplier: 2.0,
        retryableErrors: [.timeout, .connectionLost, .serverError(statusCode: 500)]
    )

    func delay(for attempt: Int) -> TimeInterval {
        let delay = baseDelay * pow(backoffMultiplier, Double(attempt))
        return min(delay, maxDelay)
    }

    func shouldRetry(error: Error, attempt: Int) -> Bool {
        guard attempt < maxRetries else { return false }

        if let apiError = error as? APIError {
            return retryableErrors.contains(apiError)
        }
        return false
    }
}

// Retry Extension for APIClient
extension APIClient {
    func requestWithRetry<T: Decodable>(
        _ endpoint: Endpoint,
        policy: RetryPolicy = .default
    ) async throws -> T {
        var lastError: Error?

        for attempt in 0...policy.maxRetries {
            do {
                return try await request(endpoint)
            } catch {
                lastError = error

                if policy.shouldRetry(error: error, attempt: attempt) {
                    let delay = policy.delay(for: attempt)
                    try await Task.sleep(for: .seconds(delay))
                    continue
                }

                throw error
            }
        }

        throw lastError ?? APIError.invalidResponse
    }
}
```

### Retry Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    RETRY FLOW                                │
│                                                             │
│  ┌──────────┐     ┌──────────────┐     ┌────────────────┐ │
│  │ API Call │────▶│ Success?     │────▶│ Return Result  │ │
│  └──────────┘     └──────┬───────┘     └────────────────┘ │
│                          │ NO                               │
│                          ▼                                  │
│                  ┌──────────────┐                           │
│                  │ Retryable?   │                           │
│                  └──────┬───────┘                           │
│                         │ NO                                │
│                         ▼                                   │
│                 ┌──────────────┐                            │
│                 │ Throw Error  │                            │
│                 └──────────────┘                            │
│                         │ YES                               │
│                         ▼                                   │
│                 ┌──────────────┐                            │
│                 │ Max Retries? │                            │
│                 └──────┬───────┘                            │
│                        │ YES                                │
│                        ▼                                    │
│                ┌──────────────┐                             │
│                │ Throw Error  │                             │
│                └──────────────┘                             │
│                        │ NO                                 │
│                        ▼                                    │
│                ┌──────────────┐                             │
│                │ Wait (backoff)│                             │
│                └──────┬───────┘                             │
│                       │                                     │
│                       ▼                                     │
│               ┌──────────────┐                              │
│               │ Retry Call   │──────▶ Back to API Call      │
│               └──────────────┘                              │
│                                                             │
│  Attempt 1: 1s delay                                        │
│  Attempt 2: 2s delay                                        │
│  Attempt 3: 4s delay                                        │
│  Attempt 4: 8s delay                                        │
│  Attempt 5: 16s delay                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## Request Cancellation

```swift
// Task Cancellation
class RequestManager {
    private var activeTasks: [String: Task<Void, Never>] = [:]
    private let lock = NSLock()

    func cancelRequest(id: String) {
        lock.lock()
        activeTasks[id]?.cancel()
        activeTasks.removeValue(forKey: id)
        lock.unlock()
    }

    func cancelAllRequests() {
        lock.lock()
        for (_, task) in activeTasks {
            task.cancel()
        }
        activeTasks.removeAll()
        lock.unlock()
    }

    func makeRequest<T>(
        id: String = UUID().uuidString,
        operation: () async throws -> T
    ) async throws -> T {
        let task = Task { try await operation() }
        lock.lock()
        activeTasks[id] = Task { await task.result }
        lock.unlock()

        defer {
            lock.lock()
            activeTasks.removeValue(forKey: id)
            lock.unlock()
        }

        return try await task.value
    }
}

// Cancellable Request
extension APIClient {
    func cancellableRequest<T: Decodable>(
        _ endpoint: Endpoint,
        cancellationToken: String = UUID().uuidString,
        manager: RequestManager = RequestManager()
    ) async throws -> T {
        try Task.checkCancellation()

        let (data, response) = try await manager.makeRequest {
            try await self.request(endpoint)
        }

        return try handleResponse(data, response: response)
    }
}
```

```swift
// Combine-based Cancellation
extension APIClient {
    func requestPublisher<T: Decodable>(_ endpoint: Endpoint) -> AnyPublisher<T, APIError> {
        var urlRequest: URLRequest
        do {
            urlRequest = try buildRequest(from: endpoint)
        } catch {
            return Fail(error: APIError.invalidURL).eraseToAnyPublisher()
        }

        return URLSession.shared.dataTaskPublisher(for: urlRequest)
            .tryMap { data, response in
                guard let httpResponse = response as? HTTPURLResponse else {
                    throw APIError.invalidResponse
                }
                return (data, httpResponse)
            }
            .flatMap { data, response -> AnyPublisher<T, APIError> in
                do {
                    let decoded = try self.handleResponse(data, response: response)
                    return Just(decoded).setFailureType(to: APIError.self).eraseToAnyPublisher()
                } catch {
                    return Fail(error: error as! APIError).eraseToAnyPublisher()
                }
            }
            .eraseToAnyPublisher()
    }
}
```

---

## Mock Data for Testing

```swift
// Mock URLProtocol for testing
class MockURLProtocol: URLProtocol {
    static var mockData: Data?
    static var mockError: Error?
    static var mockStatusCode: Int = 200
    static var mockHeaders: [String: String] = [:]
    static var requestHandler: ((URLRequest) throws -> (Data, HTTPURLResponse))?

    override class func canInit(with request: URLRequest) -> Bool {
        true
    }

    override class func canonicalRequest(for request: URLRequest) -> URLRequest {
        request
    }

    override func startLoading() {
        if let handler = requestHandler {
            do {
                let (data, response) = try handler(request)
                client?.urlProtocol(self, didLoad: data)
                client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            } catch {
                client?.urlProtocol(self, didFailWithError: error)
            }
        } else if let error = MockURLProtocol.mockError {
            client?.urlProtocol(self, didFailWithError: error)
        } else if let data = MockURLProtocol.mockData {
            let response = HTTPURLResponse(
                url: request.url!,
                statusCode: MockURLProtocol.mockStatusCode,
                httpVersion: "HTTP/1.1",
                headerFields: MockURLProtocol.mockHeaders
            )!
            client?.urlProtocol(self, didLoad: data)
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
        }

        client?.urlProtocolDidFinishLoading(self)
    }

    override func stopLoading() {}

    static func setupMock<T: Encodable>(data: T, statusCode: Int = 200) {
        let encoded = try? JSONEncoder().encode(data)
        mockData = encoded
        mockStatusCode = statusCode
    }

    static func setupMockError(_ error: Error) {
        mockError = error
    }

    static func setupMockJSON(_ jsonString: String) {
        mockData = jsonString.data(using: .utf8)
    }

    static func reset() {
        mockData = nil
        mockError = nil
        mockStatusCode = 200
        mockHeaders = [:]
        requestHandler = nil
    }
}
```

```swift
// Mock API Client for testing
class MockAPIClient: APIClientProtocol {
    var requestResult: Any?
    var requestError: Error?
    var requestCallCount = 0
    var lastEndpoint: Endpoint?

    func request<T: Decodable>(_ endpoint: Endpoint) async throws -> T {
        requestCallCount += 1
        lastEndpoint = endpoint

        if let error = requestError {
            throw error
        }

        if let result = requestResult as? T {
            return result
        }

        throw APIError.invalidResponse
    }

    func request(_ endpoint: Endpoint) async throws -> (Data, HTTPURLResponse) {
        requestCallCount += 1
        lastEndpoint = endpoint

        if let error = requestError {
            throw error
        }

        let data = MockURLProtocol.mockData ?? Data()
        let response = HTTPURLResponse(
            url: URL(string: "https://api.test.com")!,
            statusCode: MockURLProtocol.mockStatusCode,
            httpVersion: "HTTP/1.1",
            headerFields: nil
        )!

        return (data, response)
    }

    func upload<T: Decodable>(_ endpoint: Endpoint, data: Data) async throws -> T {
        try await request(endpoint)
    }
}

// Mock WebSocket Client for testing
class MockWebSocketClient: WebSocketClientProtocol {
    var isConnected = false
    var sentMessages: [WebSocketMessage] = []
    var tokenHandler: ((String) -> Void)?
    var completedHandler: ((ChatResponse) -> Void)?
    var errorHandler: ((Error) -> Void)?

    func connect(token: String) async throws {
        isConnected = true
    }

    func disconnect() {
        isConnected = false
    }

    func send(_ message: WebSocketMessage) {
        sentMessages.append(message)
    }

    func onTokenReceived(_ handler: @escaping @Sendable (String) -> Void) {
        tokenHandler = handler
    }

    func onCompleted(_ handler: @escaping @Sendable (ChatResponse) -> Void) {
        completedHandler = handler
    }

    func onToolCall(_ handler: @escaping @Sendable (ToolCall) -> Void) {}

    func onThinking(_ handler: @escaping @Sendable (String) -> Void) {}

    func onError(_ handler: @escaping @Sendable (Error) -> Void) {
        errorHandler = handler
    }

    func onConnectionStateChange(_ handler: @escaping @Sendable (WebSocketState) -> Void) {}

    // Test helpers
    func simulateToken(_ token: String) {
        tokenHandler?(token)
    }

    func simulateCompletion(_ response: ChatResponse) {
        completedHandler?(response)
    }

    func simulateError(_ error: Error) {
        errorHandler?(error)
    }
}
```

### Test Data Factory

```swift
// Test Data Factory
enum TestDataFactory {
    static func makeUser() -> User {
        User(
            id: "user-1",
            email: "test@example.com",
            name: "Test User",
            avatarUrl: nil,
            tenantId: "tenant-1",
            createdAt: Date(),
            updatedAt: Date(),
            biometricEnabled: false,
            kycStatus: .verified
        )
    }

    static func makeMessage(content: String = "Hello, AI!") -> Message {
        Message(
            id: UUID(),
            content: content,
            role: .user,
            timestamp: Date(),
            attachments: [],
            toolCalls: [],
            isStreaming: false,
            tokens: []
        )
    }

    static func makeConversation() -> Conversation {
        Conversation(
            id: "conv-1",
            title: "Test Conversation",
            agentId: "agent-1",
            modelId: "model-1",
            createdAt: Date(),
            updatedAt: Date(),
            messageCount: 5,
            lastMessage: makeMessage()
        )
    }

    static func makeAgent() -> Agent {
        Agent(
            id: "agent-1",
            name: "Test Agent",
            description: "A test agent",
            systemPrompt: "You are a helpful assistant.",
            modelId: "gpt-4",
            avatarUrl: nil,
            capabilities: ["chat", "code"],
            isDefault: true,
            createdAt: Date()
        )
    }

    static func makeAuthTokens() -> AuthTokens {
        AuthTokens(
            accessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
            refreshToken: "refresh_token_123",
            expiresIn: 3600,
            tokenType: "Bearer"
        )
    }

    static func makeLoginResponse() -> AuthTokensResponse {
        AuthTokensResponse(
            accessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
            refreshToken: "refresh_token_123",
            expiresIn: 3600,
            tokenType: "Bearer"
        )
    }
}
```

---

## Network Monitoring

```swift
// Network Status Display
struct NetworkStatusView: View {
    @ObservedObject var networkMonitor = NetworkMonitor.shared

    var body: some View {
        HStack {
            Circle()
                .fill(networkMonitor.isConnected ? Color.green : Color.red)
                .frame(width: 8, height: 8)

            Text(networkMonitor.isConnected ? "Online" : "Offline")
                .font(.caption)

            if networkMonitor.isConnected {
                Text(networkMonitor.connectionType.label)
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
        }
    }
}

extension NetworkMonitor.ConnectionType {
    var label: String {
        switch self {
        case .wifi: return "WiFi"
        case .cellular: return "Cellular"
        case .ethernet: return "Ethernet"
        case .unknown: return "Unknown"
        }
    }
}

// Network Quality Indicator
struct NetworkQualityIndicator: View {
    @ObservedObject var networkMonitor = NetworkMonitor.shared

    var quality: Quality {
        if !networkMonitor.isConnected { return .offline }
        if networkMonitor.isExpensive { return .poor }
        return .good
    }

    enum Quality {
        case good, poor, offline

        var color: Color {
            switch self {
            case .good: return .green
            case .poor: return .yellow
            case .offline: return .red
            }
        }

        var label: String {
            switch self {
            case .good: return "Good"
            case .poor: return "Slow"
            case .offline: return "Offline"
            }
        }
    }

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: quality == .offline ? "wifi.slash" : "wifi")
                .foregroundColor(quality.color)
            Text(quality.label)
                .font(.caption2)
                .foregroundColor(.secondary)
        }
    }
}
```

### Network Monitoring Diagram

```
┌─────────────────────────────────────────────────────────────┐
│              NETWORK MONITORING ARCHITECTURE                  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                   NWPathMonitor                       │  │
│  │                                                      │  │
│  │  pathUpdateHandler:                                  │  │
│  │    ├── status: .satisfied / .unsatisfied              │  │
│  │    ├── isExpensive: true / false                      │  │
│  │    └── interfaceType: .wifi / .cellular / ...         │  │
│  └──────────────────────┬───────────────────────────────┘  │
│                         │                                  │
│                         ▼                                  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                NetworkMonitor                          │  │
│  │                                                      │  │
│  │  @Published var isConnected: Bool                    │  │
│  │  @Published var connectionType: ConnectionType       │  │
│  │  @Published var isExpensive: Bool                    │  │
│  └──────────────────────┬───────────────────────────────┘  │
│                         │                                  │
│           ┌─────────────┴─────────────┐                    │
│           ▼                           ▼                    │
│  ┌──────────────────┐      ┌──────────────────┐           │
│  │     Views         │      │   Services        │           │
│  │  - Status Bar     │      │ - API Client      │           │
│  │  - Offline Banner │      │ - WebSocket       │           │
│  │  - Sync Manager   │      │ - Sync Engine     │           │
│  └──────────────────┘      └──────────────────┘           │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Network Status Flow                       │  │
│  │                                                      │  │
│  │  Online ──────────────────────────────────────────▶  │  │
│  │    ├── API calls allowed                             │  │
│  │    ├── WebSocket connected                           │  │
│  │    └── Sync engine active                            │  │
│  │                                                      │  │
│  │  Offline ─────────────────────────────────────────▶  │  │
│  │    ├── Queue API calls                               │  │
│  │    ├── Disconnect WebSocket                          │  │
│  │    ├── Use cached data                               │  │
│  │    └── Show offline banner                           │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```
