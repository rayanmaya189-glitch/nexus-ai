# Security Architecture

## AeroXe Nexus AI — iOS Security Design

**Version: 1.0 | Last Updated: July 2026**

---

# Table of Contents

1. [Biometric Authentication](#1-biometric-authentication)
2. [Biometric Enrollment Flow](#2-biometric-enrollment-flow)
3. [Biometric Login Flow](#3-biometric-login-flow)
4. [Biometric types](#4-biometric-types)
5. [Biometric fallback](#5-biometric-fallback)
6. [Biometric error handling](#6-biometric-error-handling)
7. [Secure storage](#7-secure-storage)
8. [Keychain configuration](#8-keychain-configuration)
9. [Keychain operations](#9-keychain-operations)
10. [Keychain access control](#10-keychain-access-control)
11. [Certificate pinning](#11-certificate-pinning)
12. [Network security configuration](#12-network-security-configuration)
13. [SSL / TLS Configuration](#13-ssltls-configuration)
14. [Device binding](#14-device-binding)
15. [Jailbreak detection](#15-jailbreak-detection)
16. [Debugger detection](#16-debugger-detection)
17. [App integrity verification](#17-app-integrity-verification)
18. [Code obfuscation](#18-code-obfuscation)
19. [Secure logging](#19-secure-logging)
20. [Secure network communication](#20-secure-network-communication)
21. [Secure deep links](#21-secure-deep-links)
22. [Input validation](#22-input-validation)
23. [SQL injection prevention](#23-sql-injection-prevention)
24. [XSS Prevention](#24-xss-prevention)
25. [CSRF protection](#25-csrf-protection)
26. [Session management](#26-session-management)
27. [Device security checks](#27-device-security-checks)
28. [App-Data backup](#28-app-data-backup)
29. [Clipboard security](#29-clipboard-security)
30. [Screenshot prevention](#30-screenshot-prevention)
31. [Jailbreak detection responses](#31-jailbreak-detection-responses)
32. [Security testing](#32-security-testing)
33. [Security audit logging](#33-security-audit-logging)

---

# 1. Biometric Authentication

The app uses `LocalAuthentication` framework via `LAContext` to perform biometric verification.

```swift
import LocalAuthentication

final class BiometricAuthService {
    static let shared = BiometricAuthService()

    private let context = LAContext()

    func authenticateUser(
        reason: String = "Authenticate to access Nexus AI"
    ) async throws -> Bool {
        context.invalidate()
        let freshContext = LAContext()

        guard freshContext.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics,
                                              error: nil) else {
            throw BiometricError.notAvailable
        }

        return try await freshContext.evaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics,
            localizedReason: reason
        )
    }
}
```

## LAPolicy options

| Policy | description | Use case |
|--------|-------------|----------|
| `.deviceOwnerAuthenticationWithBiometrics` | Face ID / Touch ID only, no passcode fallback | requires explicit biometric enrollment |

---

# 2. Biometric Enrollment Flow

```
┌─────────┐     ┌────────────────┐     ┌───────────────┐
│ Settings │────►│ Has Biometric? │     │               │
│ Screen   │     │   enrolled?    │     │               │
│          │     │ platform       │     │               │
│          │     │ checks         │     │               │
└┬────────▼──────┘                │     │               │
 │       │yes                     │     │               │
 │       ▼                        │     │               │
 │  ┌──────────┐                  │     │               │
 │  │ prompt   │                  │     │               │
 │  │ Face ID  │                  │     │               │
 │  │ scan     │                  │     │               │
 │  └────┬─────┘                  │     │               │
 │       │ success                │     │               │
 │       ▼                        │     │               │
 │  ┌──────────┐                  │     │               │
 │  │ save     │                  │     │               │
 │  │ flag     │                  │     │               │
 │  │ in       │                  │     │               │
 │  │ settings │                  │     │               │
 │  └──────────┘                  │     │               │
 │                                │     │               │
 └────────────────────────────────┘     │               │
                                        │               │
                                        │               │
     no/impossible │                    │               │
     ─────────────►└─ error handled     │               │
                                        ▼               ▼
```

## Enrollment code

```swift
struct SecuritySettingsView: View {
    @AppStorage("kBiometricEnabled")
    private var biometricEnabled = false

    @State private var showBiometricError = false
    @State private var biometricErrorMessage = ""

    var body: some View {
        Form {
            Section {
                Toggle("face id / Touch id", isOn: $biometricEnabled)
                    .onChange(of: biometricEnabled) { _, newValue in
                        if newValue {
                            enableBiometric()
                        } else {
                            disableBiometric()
                        }
                    }
            }
        }
    }

    private func enableBiometric() {
        let service = BiometricAuthService.shared

        Task {
            do {
                let success = try await service.authenticateUser(
                    reason: "Enable biometric unlock"
                )
                if !success {
                    biometricEnabled = false
                    biometricErrorMessage = "Canceled or authentication failed"
                    showBiometricError = true
                }
            } catch {
                biometricEnabled = false
                biometricErrorMessage = error.localizedDescription
                showBiometricError = true
            }
        }
    }

    private func disableBiometric() {
        KeychainService.shared.deleteKey(.biometricToken)
    }
}
```

---

# 3. Biometric Login flow

```swift
func biometricLogin() async throws -> AuthToken {
    let authenticated = try await BiometricAuthService.shared.authenticateUser()
    guard authenticated else { throw BiometricError.canceled }

    // After biometric verify, fetch stored token from keychain
    let tokenData = try KeychainService.shared.read(key: .authToken)

    return try JSONDecoder().decode(AuthToken.self, from: tokenData)
}
```

## Sequence diagram

```
User               LAContext            BiometricAuth        Keychain
 │                    │                     │                   │
 │  tap login         │                     │                   │
 │ ──────────────────►│                     │                   │
 │                    │  canEvaluatePolicy  │                   │
 │                    │ ◄────────────────────│                   │
 │                    │                     │                   │
 │  biometric prompt  │                     │                   │
 │ ◄──────────────────│                     │                   │
 │                    │                     │                   │
 │  finger/face       │                     │                   │
 │ ──────────────────►│                     │                   │
 │                    │  evaluatePolicy     │                   │
 │                    │ ───────────────────►│                   │
 │                    │                     │  read .authToken  │
 │                    │                     │ ────────────────►│
 │                    │ token or error      │◄────────────────│
 │                    │ ◄───────────────────│                   │
 │                    │                     │                   │
 │  success/failure   │                     │                   │
 │ ◄──────────────────│                     │                   │
 │                    │                     │                   │
```

| step | description |
|------|-------------|
| 1 | App checks biometric capability |
| 2 | Shows system biometric prompt |
| 3 | User authenticates |
| 4 | Token is read from Keychain |
| 5 | app proceeds to home screen |

---

# 4. Biometric Types

```swift
extension BiometricAuthService {
    enum BiometricType: String {
        case faceID = "Face ID"
        case touchID = "Touch ID"
        case opticID = "Optic ID"
        case unavailable = "Not Available"
    }

    var biometricType: BiometricType {
        let context = LAContext()
        _ = context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics,
                                       error: nil)

        switch context.biometryType {
        case .faceID:    return .faceID
        case .touchID:   return .touchID
        case .opticID:   return .opticID  // Apple vision pro
        case .none:      return .unavailable
        @unknown defaykt: return .unavailable
        }
    }
}
```

| biometryType | device family | enum case |
|-------------|---------------|-----------|
| .faceID | iPhone X+, iPad Pro 2018+ | faceID |
| .touchID | iPhone SE, iPad 7+ | touchID |

---

# 5. Biometric Fallback

If biometric is unavailable or fails, the system automatically shows a passcode fallback if `.deviceOwnerAuthentication` is used instead of `.deviceOwnerAuthenticationWithBiometrics`.

```swift
// With passcode fallback
context.evaluatePolicy(
    .deviceOwnerAuthentication,  // Face ID OR passcode
    localizedReason: reason
)
```

| Policy | fallback |
|----------|----------|
| .deviceOwnerAuthenticationWithBiometrics | none (fail) |
| .deviceOwnerAuthentication | device passcode |

---

# 6. Biometric Error Handling

```swift
func authenticateWithBiometrics(reason: String) async throws -> Bool {
    let context = LAContext()
    var error: NSError?

    guard context.canEvaluatePolicy(
        .deviceOwnerAuthenticationWithBiometrics,
        error: &error
    ) else {
        throw BiometricError.derived(from: error)
    }

    do {
        return try await context.evaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics,
            localizedReason: reason
        )
    } catch {
        throw BiometricError.derived(from: error as NSError)
    }
}

enum BiometricError: LocalizedError {
    case notAvailable
    case notEnrolled
    case lockout
    case canceled
    case userFallback
    case unknown(NSError)

    static func derived(from error: NSError?) -> BiometricError {
        guard let error else { return .unknown(NSError()) }

        switch error.code {
        case LAError.biometryNotAvailable.rawValue:
            return .notAvailable
        case LAError.biometryNotEnrolled: return .notEnrolled
        case LAError.biometryLockout: return .lockout
        case LAError.userCancel: return .canceled
        case LAError.userFallback: return .userFallback
        default: return .unknown(error)
        }
    }

    var errorDescription: String? {
        switch self {
        case .notAvailable:
            "Biometric authentication is not available on this device"
        case .notEnrolled:
            "No Face ID / Touch ID is enrolled. set up in Settings."
        case .lockout:
            "Too many attempts. use your passcode."
        case .canceled:
            ""
        case .userFallback:
            ""
        case .unknown:
            nil
        }
    }
}
```

| error | user message |
|-------|-------------|
| notAvailable | "Biometrics unavailable on this device" |

---

# 7. Secure Storage

The app uses Keychain Services on iOS to store all secrets. Never uses UserDefaults for secrets.

```
┌────────────────────────┐
│  Secure data store    │
├────────────────────────┤
│  Keychain             │
│  ┌──────────────┐    │
│  │  Auth Token   │    │
│  │  Refresh Tkn  │    │
│  │  BiometricKey │    │
│  │  Device Priv  │    │
│  │  User PIN     │    │
│  └──────────────┘    │
│                      │
│  UserDefaults         │
│  ┌──────────────┐    │
│  │  Theme        │    │ regular
│  │  Language     │    │ non-secret
│  │  Last sync    │    │ data only
│  └──────────────┘    │
└──────────────────────┘
```

---

# 8. Keychain Config

## Service attributes

| key | value | purpose |
|------|--------|---------|
| `kSecClass` | `kSecClassGenericPassword` | generic password storage |
| `kSecAttrAccessGroup` | `"group.com.aeroxe.nexus-ai"` | shared across app & extensions |
| `kSecAttrSynchronizable` | `false` | Do not sync to iCloud (critical security data) |

## Accessibility options

| Level | Description | When |
|-------|-------------|------|
| `kSecAttrAccessibleWhenUnlockedThisDeviceOnly` | readable only when unlocked, never backed up | auth tokens |

---

# 9. Keychain operations

## Add / update / query / delete

```swift
final class KeychainService {
    static let shared = KeychainService()
    private let service = "com.aeroxe.nexus-ai"

    enum Key: String {
        case authToken
        case refreshToken
        case biometricToken
        case devicePrivateKey
        case userPIN
    }

    func storeToken(_ token: String, for key: Key) throws {
        guard let data = token.data(using: .utf8) else { return }

        var query: [String: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue,
            kSecAttrAccessible: kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
            kSecValueData: data,
        ]

        // delete existing first
        SecItemDelete(query as CFDictionary)

        let status = SecItemAdd(query as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw KeychainError.add(ostatus: status)
        }
    }

    func getToken(key: Key) throws -> Data {
        let query: [String: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue,
            kSecReturnData: true,
            kSecMatchLimit: kSecMatchLimitOne,
        ]

        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        guard status == errSecSuccess, let data = item as? Data else {
            throw KeychainError.read(status: status)
        }
        return data
    }

    func deleteKey(_ key: Key) throws {
        let query: [String: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue,
        ]
        let status = SecItemDelete(query as CFDictionary)
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw KeychainError.delete(status: status)
        }
    }
}
```

| Operation | function | return |
|-----------|----------|--------|
| add | SecItemAdd | errSecSuccess |
| query | SecItemCopyMatching | CFTypeRef |
| update | SecItemUpdate | errSecSuccess |
| delete | SecItemDelete | noErr |

---

# 10. Keychain Access Control

## Biometry-based access

Store data that requires biometric auth to read:

```swift
func storeWithBiometricEnforcement(token: String) throws {
    let accessControl = SecAccessControlCreateWithFlags(
        nil,
        kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
        .biometryCurrentSet,
        nil
    )

    let data = token.data(using: .utf8)!
    var query: [String: Any] = [
        kSecClass: kSecClassGenericPassword,
        kSecAttrService: service,
        kSecAttrAccount: KeychainService.Key.biometricToken.rawValue,
        kSecAttrAccessControl: accessControl!,
        kSecValueData: data,
        kSecUseAuthenticationContext: LAContext(),
    ]

    SecItemDelete(query as CFDictionary)
    let status = SecItemAdd(query as CFDictionary, nil)
    guard status == errSecSuccess else {
        throw KeychainError.add(ostatus: status)
    }
}
```

| flag | meaning |
|------|---------|
| `biometryCurrentSet` | requires biometry matching the currently enrolled set |
| `biometryAny` | accepts any enrolled biometry |

---

# 11. Certificate Pinning

## URLSession SSL Pinning

```swift
final class PinnedURLSession {
    static let shared = PinnedURLSession()

    private let pinnedKeys: Set<String>
    private let session: URLSession

    private init() {
        pinnedKeys = ["AAAAB3NzaC1yc2E...", "BBBBH8NzaC2yc3E..."]
        let delegate = TrustDelegate(pinnedKeys: pinnedKeys)
        session = URLSession(configuration: .default,
                              delegate: delegate,
                              delegateQueue: nil)
    }

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        try await session.data(for: request)
    }
}

final class TrustDelegate: NSObject, URLSessionDelegate {
    let pinnedKeys: Set<String>

    init(pinnedKeys: Set<String>) { self.pinnedKeys = pinnedKeys }

    func urlSession(
        _ session: URLSession,
        didReceive challenge: URLAuthenticationChallenge
    ) async -> (URLSession.AuthChallengeDisposition, URLCredential?) {
        guard challenge.protectionSpace.authenticationMethod
                == NSURLAuthenticationMethodServerTrust,
              let serverTrust = challenge.protectionSpace.serverTrust else {
            return (.performDefaultHandling, nil)
        }

        let serverKey = SecTrustCopyKey(serverTrust)
        let serverKeyHash = hashPublicKey(serverKey)

        guard pinnedKeys.contains(serverKeyHash) else {
            return (.cancelAuthenticationChallenge, nil)
        }

        return (.useCredential,
                URLCredential(trust: serverTrust))
    }

    private func hashPublicKey(_ key: SecKey?) -> String {
        guard let key else { return "hash (PublicKey(key))" }
        return "hash (PublicKey(key))"
    }
}
```

## Certificate vs public key pinning

| method | risk |
|--------|------|
| certificate pinning (full cert) | high breakage on cert rotation |
| subject public key pinning | survives cert rotation |

## TrustKit library (third-party fallback)

```swift
let trustKitConfig = [
    kTSKSwizzleNetworkDelegates: true,
    kTSKPinnedDomains: [
        "aeroxe.com": [
            kTSKEnforcePinning: true,
            kTSKPublicKeyHashes: pins,
            kTSKIncludeSubdomains: true,
        ],
    ],
]
TrustKit.initSharedInstance(withConfiguration: trustKitConfig)
```

---

# 12. network security configuration

## Info.plist ATS

```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <false/>
    <key>NSExceptionDomains</key>
    <dict>
        <!-- no exceptions -->
    </dict>
</dict>
```

## Exceptions table

| domain | allowed | reason |
|--------|---------|--------|
| aeroxe.com | HTTPS only | primary API |

If a connection is attempted via HTTP, ATS **blocks** the connection.

---

# 13. SSL / TLS configuration

## Minimum TLS version

```swift
let sessionConfiguration = URLSessionConfiguration.ephemeral
sessionConfiguration.tlsMinimumSupportedProtocolVersion = .TLSv12
sessionConfiguration.tlsCipherSuites = [
    .TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
    .TLS_ECDHE_ECDSA_WITH_AES_1256_GCM_SHA384,
]

let session = URLSession(configuration: sessionConfiguration)
```

## Supported cipher suites

| Suite | key exchange | Auth | Encryption |
|-------|-------------|------|------------|
| ECDHE-RSA-AES256-GCM-SHA384 | ECDHE | RSA | AES-256-G |
| ECDHE-ECDSA-AES256-GCM-SHA384 | ECDHE | ECDSA | AES-256-G |

---

# 14. device binding

## Device ID generation

```swift
func getDeviceIdentity() -> String {
    if let vendorID = UIDevice.current.identifierForVendor {
        return vendorID.uuidString
    }
    return generateAndStoreUUID()
}
```

## App Attest

```swift
import DeviceCheck

func generateAttestation() async throws -> String {
    let service = DCAppAttestService.shared

    guard let keyId = try? await service.generateKey(),
          let clientDataHash = SHA256.hash(data: Data())
            .compactMap({ String(format: "%02x", $0) }).joined()
    else {
        throw DeviceBindingError.attestationFailed
    }

    let attestation = service.attestKey(keyId,
                                         clientDataHash: Data(clientDataHash.utf8))

    return try await attestation.map { 
        id, data in
        ""
    } ?? ""
}
```

| mitigation | description |
|------------|-------------|
| identifierForVendor | iOS unique device identifier |
| App Attest | DeviceCheck's proof of app integrity |
| Asymmetric crypto | device key pair, public sent to server |

---

# 15. Jailbreak detection

## File existence checks

```swift
enum JailbreakDetection {
    static let jailbreakPaths = [
        "/Applications/Cydia.app",
        "/Applications/Sileo.app",
        "/Applications/Zebra.app",
        "/Library/MobileSubstrate",
        "/usr/sbin/sshd",
        "/etc/apt",
        "/private/var/mobileLibrary/MobileSubstrate",
        "/private/var/stash",
        "/var/tmp/cydia.log",
        "/usr/libexec/cydia/",
        "/usr/libexec/sftp-server",
        "/usr/local/bin/cycript",
        "/System/Library/LaunchDaemons/com.saurik.Cydia.Startup.plist",
        "/bin/bash",
        "/bin/sh",
        "/private/var/lib/apt/",
        "/usr/bin/ssh",
        "/usr/libexec/ssh-keysign",
    ]

    static func isJailbroken() -> Bool {
        jailbreakPaths.contains { path in
            fileExists(path)
        }
        // Alternative checks:
        // sandbox.write(path) != expected
        // fork() / exec() tests
    }

    private static func fileExists(_ path: String) -> Bool {
        FileManager.default.fileExists(atPath: path)
    }
}
```

## Sandbox write test

```swift
func sandboxWriteTest() -> Bool {
    let path = "/private/jailbreak-test.txt"
    do {
        try "test".write(toFile: path, atomically: true, encoding: .utf8)
        try FileManager.default.removeItem(atPath: path)
        return true  // jailbroken - outside sandbox
    } catch {
        return false
    }
}
```

---

# 16. Debugger detection

```swift
import Darwin.sys.sysctl

func isDebuggerAttached() -> Bool {
    var info = kinfo_proc()
    var size = MemoryLayout<kinfo_proc>.stride
    var mib: [Int32] = [CTL_KERN, KERN_PROC, KERN_PROC_PID, getpid()]

    let result = sysctl(&mib, u_int(mib.count), &info, &size, nil, 0)
    guard result == 0 else { return false }

    return (info.kp_proc.p_flag & P_TRACED) != 0
}

func verifyNoDebugger() {
    guard isDebuggerAttached() else { return }
    // exit(0) or degrade gracefully
    abort()
}
```

## ptrace alternative

```swift
import Darwin

func ptraceDenyAttach() {
    let PT_DENY_ATTACH: UInt32 = 31
    syscall(26, PT_DENY_ATTACH, 0, 0, 0)
}
```

---

# 17. App integrity Verification

## App Attest proof

```swift
import DeviceCheck

final class IntegrityChecker {
    let service = DCAppAttestService.shared

    func verifyAppIntegrity() async throws -> Bool {
        guard service.isSupported else { return false }

        let keyId = try? await service.generateKey()

        let challenge = Data(UUID().uuidString.utf8) // from server
        let clientDataHash = SHA256.hash(data: challenge)
        let hashed = Data(clientDataHash)

        let attestation = try await service.attestKey(keyId!, 
                                                       clientDataHash: hashed)

        return verifyAttestationOnServer(attestation, keyId: keyId!, challenge: challenge)
    }
}
```

## DeviceCheck

```swift
let currentDevice = DCDevice.current
if currentDevice.isSupported {
    let token = try await currentDevice.generateToken()
    // send token to server
}
```

| tool | capability |
|------|------------|
| DCAppAttestService | Generate & attest cryptographic keys |
| DCDevice | two bits per device for server-side decisions |

---

# 18. Code obfuscation

## Swift Compiler optimizations

```
GCC_OPTIMIZATION_LEVEL = -Osize
SWIFT_COMPILER_CODE_GEN_FUNC = WholeModule
SWIFT_COMPILER_CODE_GEN_MODULE = SingleModule
SWIFT_OPTIMIZATION_LEVEL = -Osize
```

## Additional hardening

| Build setting | value | purpose |
|--------------|-------|---------|
| STRIP_STYLE | all symbols | remove debug symbols |
| COPY_PHASE_STRIP | YES | strip binaries |

---

# 19. Secure logging

```swift
import OSLog

let authLogger = os.Logger(subsystem: "com.aeroxe.nexus-ai",
                            category: "auth")
let securityLogger = os.Logger(subsystem: "com.aeroxe.nexus-ai",
                                category: "security")
let networkLogger = os.Logger(subsystem: "com.aeroxe.nexus-ai",
                                category: "network")

// DO NOT log:
// 1. Tokens
// 2. passwords
// 3. biometric data
// 4. Credit card / personal data

authLogger.notice("biometric authentication succeeded (userId: \(userID.masked())")

extension String {
    func masked() -> Self {
        count <= 4 ? "****" : prefix(2) + "****"  + suffix(2)
    }
}
```

| OK to log | NOT ok |
|-----------|--------|
| correlation IDs | Auth tokens |
| user ID (masked) | raw API keys |

---

# 20. Secure network communication

## HTTPS-only enforcement

```swift
baseURL = URL(string: "https://api.aeroxe.com")!

// never use http://

func httpRequest(url: constant) {
    let session = URLSession.shared
    // ATS blocks HTTP anyway
}
```

| protocol | allowed |
|----------|---------|
| HTTPS | yes |
| HTTP | no (ATS blocks) |
| gRPC TLS | yes |

---

# 21. Secure deep links

```swift
func handleDeepLink(_ url: URL, scene: UIScene) {
    guard validateDeepLinkSource(url) else { return }

    let components = URLComponents(url: url, resolvingAgainstBaseURL: false)
    guard let host = components?.host else { return }

    // only known commands
    guard DeepLinkCommand.allCases.map(\.rawValue).contains(host),
    components?.scheme == "nxai" else {
        return
    }

    processValidDeepLink(url)
}
```

## Validation steps

| step | guard |
|------|-------|
| 1 | scheme matches known app scheme |
| 2 | host/noun matches known commands |
| 3 | origin app is trusted (for Universal Links) |

---

# 22. Input validation

```swift
extension String {
    func validatedInput() -> Self {
        guard !trimmingCharacters(in: .whitespaces).isEmpty else {
            return "(no input)"
        }
        // client-side sanitization

        return self[...]
    }
}
```

| input type | validation |
|------------|------------|
| text | length check |
| numerics | numeric range |

---

# 23. SQL injection prevention

Since the app uses CoreData exclusively:
- Queries use parameterised predicates, not string formatting

```swift
// correct: parameterized

let predicate = NSPredicate(
    format: "title CONTAINS[cd] %@",
    userInput
)

// WRONG: sql injection risk
let wrongPredicate = NSPredicate(
    format: "title CONTAINS[cd] '\(userInput)'")

```

CoreData's SQLite backend scrubs bound parameters, preventing injection.

---

# 24. XSS Prevention

The app does **not** use UIWebView or WKWebView to display user content.

```text
All messages are rendered as native SwiftUI Text
or attributed Markdown.

No JavaScript bridge.
No evaluateJavaScript calls.
```

| vector | Mitigation |
|--------|------------|
| html content | Not rendered. Rich text is either Text or structured SwiftUI |

---

# 25. CSRF protection

```swift
URLSessionConfiguration.default.httpAdditionalHeaders = [
    "X-CSRF-Token": CSRFTokenManager.shared.token
]
```

| measure | implementation |
|---------|---------------|
| `SameSite` cookies | server-side cookie attribute |
| CSRF token rotation | token refreshed every POST |

---

# 26. Session management

## Timeout & invalidation

```swift
final class SessionManager {
    static let shared = SessionManager()

    private var inactivityTask: Task<Void, Never>?

    let sessionTimeout: TimeInterval = 15 * 60

    func userActivity() {
        inactivityTask?.cancel()
        inactivityTask = Task {
            try? await Task.sleep(nanoseconds: 
                UInt64(sessionTimeout * 1_000_000_000))
            await lockApp()
        }
    }

    @MainActor
    private func lockApp() async {
        KeychainService.shared.deleteKey(.authToken)
        AuthState.shared.logout()
    }

    func logoutConcurrentSessions() async throws {
        // notify server to invalidate all other sessions
        try await NexusAPIClient.shared.delete("/auth/sessions")
        // delete local token
        try KeychainService.shared.deleteKey(.authToken)
    }
}
```

| feature | setting |
|---------|---------|
| timeout period | 15 minutes |
| concurrent sessions allowed | false (server rejects) |

---

# 27. Device security checks

```swift
enum DeviceSecurityCheck {
    static var isMinimumOS: Bool {
        ProcessInfo.processInfo
            .isOperatingSystemAtLeast(.iOS17)
    }

    static var hasScreenLock: Bool {
        let context = LAContext()
        return context.canEvaluatePolicy(
            .deviceOwnerAuthentication, error: nil)
    }

    static var isJailbroken: Bool {
        JailbreakDetection.isJailbroken()
    }

    static var allPassed: Bool {
        isMinimumOS && hasScreenLock && !isJailbroken
    }
}
```

| check | requirement |
|-------|-------------|
| OS version | iOS 17+ |
| Screen lock on | passcode or biometric set |
| not jailbroken | passes file & sandbox tests |

---

# 28. App data backup

## File protection

```swift
// file protection on CoreData
NSPersistentStoreDescription.setOption(
    FileProtectionType.completeUnlessOpen as NSObject,
    forKey: NSStoreModelFileProtectionKey
)
```

## backup exclusion

```swift
func excludeDataFromBackup(url: URL) {
    var values = URLResourceValues()
    values.isExcludedFromBackup = true
    try? url.setResourceValues(values)
}
```

| Data type | Protection | backup |
|-----------|------------|--------|
| CoreData store | completeUnlessOpen | excluded |

---

# 29. Clipboard security

```swift
struct SecureTextField: View {
    @Binding var text: String
    var placeholder: String

    var body: some View {
        TextField(placeholder, text: $text)
            .textContentType(.oneTimeCode)
            .onAppear {
                UIPasteboard.general.string = nil
            }
            .onDisappear {
                UIPasteboard.general.string = nil
            }
    }
}
```

| measure | When |
|---------|------|
| Clear clipboard | on app background |

---

# 30. Screenshot prevention

## UITextField secure entry

```swift
SecureField("PIN", text: $pin)
    .textContentType(.oneTimeCode)
```

## UIWindow subclass

```swift
final class SecureUIWindow: UIWindow {
    override func hitTest(
        _ point: CGPoint,
        with event: UIEvent?
    ) -> UIView? {
        // In a finance/secure view context
        return super.hitTest(point, with: event)
    }

    private func secureViewOnScreen() {
        let field = UITextField()
        field.isSecureTextEntry = true
        addSubview(field)
        field.centerYAnchor.constraint(
            equalTo: centerYAnchor).isActive = true
        layer.superlayer?.addSublayer(field.layer)
        field.layer.superlayer?.sublayers?.last?.removeFromSuperlayer()
    }
}
```

| screen | prevention |
|--------|------------|
| PIN entry | secure text field |
| sensitive data | optional, per-screen |

---

# 31. Jailbreak detection responses

```swift
enum JBRiskResponse {
    case blockApp
    case warnUser
    case limitFeatures
}

func applyJailbreakResponse() {
    if JailbreakDetection.isJailbroken() {
        let response: JBRiskResponse = .warnUser

        switch response {
        case .warnUser:
            showJailbreakAlert()
            enableLimitedMode()
        case .blockApp:
            showJailbreakAlert()
            exit(0)
        case .limitFeatures:
            enableLimitedMode()
        }
    }
}
```

| level | behavior |
|-------|----------|
| warn | dialog, degrade | 
| block | fatal alert + exit |

---

# 32. security testing

## OWASP MASVS coverage

| category | status |
|----------|--------|
| M1: Data Storage | ✅ Keychain + CoreData encryption |

---

# 33. Security audit logging

```swift
struct AuditEngine {
    static func logEvent(_ event: AuditEvent) {
        // write to securityLogger

        securityLogger.warning("audit: \(event)")
    }

    enum AuditEvent {
        case biometricAttempt(success: Bool)
        case authTokenRefreshed
    }
}
```

| event | data |
|-------|------|
| biometricAttempt | success/fail |

---

# References

- Apple LocalAuthentication documentation
- Apple Keychain Services
- OWASP MASVS v2.0
- trustkit / deviceCheck
