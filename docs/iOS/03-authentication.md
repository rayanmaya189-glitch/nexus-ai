# iOS Authentication

## Table of Contents

- [Login Screen](#login-screen)
- [Login ViewModel](#login-viewmodel)
- [Login API Integration](#login-api-integration)
- [JWT Token Storage](#jwt-token-storage)
- [Refresh Token Storage](#refresh-token-storage)
- [Token Refresh Flow](#token-refresh-flow)
- [Token Refresh Interceptor](#token-refresh-interceptor)
- [Biometric Authentication](#biometric-authentication)
- [Biometric Setup Screen](#biometric-setup-screen)
- [Biometric Login Flow](#biometric-login-flow)
- [Secure Storage](#secure-storage)
- [Session Management](#session-management)
- [Auto-Logout](#auto-logout)
- [Password Validation](#password-validation)
- [Error Handling](#error-handling)
- [KYC Verification Flow](#kyc-verification-flow)
- [Multi-Tenant Support](#multi-tenant-support)
- [Logout Flow](#logout-flow)
- [Auth State Persistence](#auth-state-persistence)
- [Auth Environment Object](#auth-environment-object)

---

## Login Screen

```swift
// LoginView.swift
import SwiftUI

struct LoginView: View {
    @StateObject private var viewModel = LoginViewModel()
    @Environment(\.dismiss) private var dismiss
    @Environment(\.authRepository) private var authRepository

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 32) {
                    // Header
                    LoginHeader()

                    // Form
                    VStack(spacing: 20) {
                        EmailField(
                            text: $viewModel.email,
                            error: viewModel.emailError
                        )

                        PasswordField(
                            text: $viewModel.password,
                            error: viewModel.passwordError,
                            isVisible: $viewModel.showPassword
                        )

                        if let error = viewModel.errorMessage {
                            ErrorBanner(message: error)
                        }

                        PrimaryButton(
                            title: "Log In",
                            isLoading: viewModel.isLoading,
                            action: viewModel.login
                        )
                        .accessibilityIdentifier("login_button")
                    }
                    .padding(.horizontal, 24)

                    // Divider
                    HStack {
                        Rectangle()
                            .fill(Color(.systemGray4))
                            .frame(height: 1)
                        Text("or")
                            .font(.subheadline)
                            .foregroundColor(.secondary)
                        Rectangle()
                            .fill(Color(.systemGray4))
                            .frame(height: 1)
                    }
                    .padding(.horizontal, 24)

                    // Biometric Login
                    if viewModel.isBiometricAvailable {
                        BiometricLoginButton(action: viewModel.biometricLogin)
                    }

                    // Social Login
                    SocialLoginButtons()

                    Spacer()

                    // Footer
                    VStack(spacing: 8) {
                        Text("Don't have an account?")
                            .foregroundColor(.secondary)

                        NavigationLink("Sign Up") {
                            RegisterView()
                        }
                    }
                    .padding(.bottom, 16)
                }
                .padding(.top, 60)
            }
            .scrollDismissesKeyboard(.interactively)
            .navigationBarHidden(true)
            .alert("Error", isPresented: $viewModel.showError) {
                Button("OK") { viewModel.showError = false }
            } message: {
                Text(viewModel.errorMessage ?? "An error occurred")
            }
        }
    }
}

// Login Header Component
struct LoginHeader: View {
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "brain.head.profile")
                .font(.system(size: 60))
                .foregroundStyle(
                    LinearGradient(
                        colors: [.blue, .purple],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )

            Text("Welcome Back")
                .font(.largeTitle.bold())

            Text("Sign in to continue")
                .font(.subheadline)
                .foregroundColor(.secondary)
        }
        .padding(.bottom, 16)
    }
}

// Email Field Component
struct EmailField: View {
    @Binding var text: String
    let error: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            TextField("Email address", text: $text)
                .textFieldStyle(.plain)
                .textContentType(.emailAddress)
                .keyboardType(.emailAddress)
                .autocapitalization(.none)
                .disableAutocorrection(true)
                .padding()
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .overlay(
                    RoundedRectangle(cornerRadius: 12)
                        .stroke(error != nil ? Color.red : Color.clear, lineWidth: 1)
                )

            if let error = error {
                Text(error)
                    .font(.caption)
                    .foregroundColor(.red)
            }
        }
    }
}

// Password Field Component
struct PasswordField: View {
    @Binding var text: String
    let error: String?
    @Binding var isVisible: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                if isVisible {
                    TextField("Password", text: $text)
                        .textFieldStyle(.plain)
                        .textContentType(.password)
                } else {
                    SecureField("Password", text: $text)
                        .textFieldStyle(.plain)
                        .textContentType(.password)
                }

                Button(action: { isVisible.toggle() }) {
                    Image(systemName: isVisible ? "eye.slash" : "eye")
                        .foregroundColor(.secondary)
                }
            }
            .padding()
            .background(Color(.systemGray6))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(error != nil ? Color.red : Color.clear, lineWidth: 1)
            )

            if let error = error {
                Text(error)
                    .font(.caption)
                    .foregroundColor(.red)
            }
        }
    }
}

// Error Banner
struct ErrorBanner: View {
    let message: String

    var body: some View {
        HStack {
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundColor(.red)
            Text(message)
                .font(.subheadline)
                .foregroundColor(.red)
        }
        .padding()
        .frame(maxWidth: .infinity)
        .background(Color.red.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// Biometric Login Button
struct BiometricLoginButton: View {
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack {
                Image(systemName: biometricIcon)
                    .font(.title3)
                Text("Sign in with \(biometricName)")
                    .fontWeight(.medium)
            }
            .frame(maxWidth: .infinity)
            .frame(height: 50)
            .background(Color(.systemGray6))
            .foregroundColor(.primary)
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
        .padding(.horizontal, 24)
    }

    private var biometricIcon: String {
        BiometricService.shared.biometricType == .faceID ? "faceid" : "touchid"
    }

    private var biometricName: String {
        BiometricService.shared.biometricType == .faceID ? "Face ID" : "Touch ID"
    }
}

// Social Login Buttons
struct SocialLoginButtons: View {
    var body: some View {
        VStack(spacing: 12) {
            SocialButton(
                title: "Continue with Google",
                icon: "g.circle.fill",
                color: .blue
            ) {}

            SocialButton(
                title: "Continue with Apple",
                icon: "apple.logo",
                color: .black
            ) {}
        }
        .padding(.horizontal, 24)
    }
}

struct SocialButton: View {
    let title: String
    let icon: String
    let color: Color
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack {
                Image(systemName: icon)
                Text(title)
                    .fontWeight(.medium)
            }
            .frame(maxWidth: .infinity)
            .frame(height: 50)
            .background(Color(.systemGray6))
            .foregroundColor(.primary)
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
    }
}
```

### Login Screen Layout Diagram

```
┌─────────────────────────────────────────────┐
│                                             │
│              ┌─────────────┐                │
│              │  🧠 Logo    │                │
│              └─────────────┘                │
│                                             │
│           Welcome Back                      │
│        Sign in to continue                  │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │ 📧 Email address                    │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │ 🔒 Password            [👁]        │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │          ⚠️ Error message           │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │          Log In                     │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  ─────────── or ───────────                │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │ 🔐 Sign in with Face ID            │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │ 🟢 Continue with Google             │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │ 🍎 Continue with Apple              │   │
│  └─────────────────────────────────────┘   │
│                                             │
│        Don't have an account?               │
│            Sign Up →                        │
│                                             │
└─────────────────────────────────────────────┘
```

---

## Login ViewModel

```swift
// LoginViewModel.swift
import Foundation
import Combine
import LocalAuthentication

@MainActor
class LoginViewModel: ObservableObject {
    // MARK: - Published State
    @Published var email: String = ""
    @Published var password: String = ""
    @Published var showPassword: Bool = false
    @Published var isLoading: Bool = false
    @Published var errorMessage: String?
    @Published var showError: Bool = false
    @Published var emailError: String?
    @Published var passwordError: String?
    @Published var isBiometricAvailable: Bool = false
    @Published var isAuthenticated: Bool = false

    // MARK: - Dependencies
    private let authUseCase: LoginUseCaseProtocol
    private let biometricUseCase: BiometricUseCaseProtocol
    private let biometricService: BiometricServiceProtocol
    private var cancellables = Set<AnyCancellable>()

    // MARK: - Init
    init(
        authUseCase: LoginUseCaseProtocol = LoginUseCase(),
        biometricUseCase: BiometricUseCaseProtocol = BiometricUseCase(),
        biometricService: BiometricServiceProtocol = BiometricService.shared
    ) {
        self.authUseCase = authUseCase
        self.biometricUseCase = biometricUseCase
        self.biometricService = biometricService
        checkBiometricAvailability()
        setupValidation()
    }

    // MARK: - Setup
    private func setupValidation() {
        $email
            .debounce(for: .milliseconds(300), scheduler: RunLoop.main)
            .sink { [weak self] email in
                self?.validateEmail(email)
            }
            .store(in: &cancellables)

        $password
            .debounce(for: .milliseconds(300), scheduler: RunLoop.main)
            .sink { [weak self] password in
                self?.validatePassword(password)
            }
            .store(in: &cancellables)
    }

    // MARK: - Validation
    private func validateEmail(_ email: String) {
        if email.isEmpty {
            emailError = nil
        } else if !EmailValidator.isValid(email) {
            emailError = "Please enter a valid email"
        } else {
            emailError = nil
        }
    }

    private func validatePassword(_ password: String) {
        if password.isEmpty {
            passwordError = nil
        } else if password.count < 12 {
            passwordError = "Password must be at least 12 characters"
        } else {
            passwordError = nil
        }
    }

    var isFormValid: Bool {
        EmailValidator.isValid(email) &&
        password.count >= 12 &&
        emailError == nil &&
        passwordError == nil
    }

    // MARK: - Login
    func login() {
        guard isFormValid else { return }

        isLoading = true
        errorMessage = nil

        Task {
            do {
                let tokens = try await authUseCase.execute(
                    email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                    password: password
                )

                isAuthenticated = true

                AnalyticsManager.shared.log(.login(method: "email"))

                os_log(.info, "Login successful for email: %{public}@", email)
            } catch {
                handleError(error)
            }

            isLoading = false
        }
    }

    // MARK: - Biometric Login
    func biometricLogin() {
        isLoading = true
        errorMessage = nil

        Task {
            do {
                let success = try await biometricUseCase.authenticateWithBiometrics()

                if success {
                    isAuthenticated = true
                    AnalyticsManager.shared.log(.login(method: "biometric"))
                    os_log(.info, "Biometric login successful")
                }
            } catch {
                handleError(error)
            }

            isLoading = false
        }
    }

    // MARK: - Biometric Availability
    private func checkBiometricAvailability() {
        isBiometricAvailable = biometricService.isAvailable
    }

    // MARK: - Error Handling
    private func handleError(_ error: Error) {
        if let authError = error as? AuthError {
            switch authError {
            case .invalidCredentials:
                errorMessage = "Invalid email or password"
            case .accountLocked:
                errorMessage = "Account locked. Please try again later."
            case .tooManyAttempts:
                errorMessage = "Too many attempts. Please try again in 15 minutes."
            case .networkError:
                errorMessage = "Network error. Please check your connection."
            case .serverError:
                errorMessage = "Server error. Please try again later."
            default:
                errorMessage = authError.localizedDescription
            }
        } else {
            errorMessage = "An unexpected error occurred"
        }

        showError = true
    }
}
```

### Login ViewModel State Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                LOGIN VIEWMODEL STATE                         │
│                                                             │
│  ┌───────────┐     ┌──────────────┐     ┌──────────────┐  │
│  │   Idle    │────▶│  Validating  │────▶│   Valid      │  │
│  │           │     │              │     │              │  │
│  └───────────┘     └──────────────┘     └──────┬───────┘  │
│                                                 │          │
│                                                 ▼          │
│                                         ┌──────────────┐  │
│                                         │   Loading    │  │
│                                         │              │  │
│                                         └──────┬───────┘  │
│                                    ┌───────────┴───────┐  │
│                                    │                   │  │
│                                    ▼                   ▼  │
│                            ┌──────────────┐    ┌──────────┐│
│                            │   Success    │    │  Error   ││
│                            │              │    │          ││
│                            └──────────────┘    └────┬─────┘│
│                                                      │     │
│                                                      ▼     │
│                                              ┌───────────┐│
│                                              │   Idle    ││
│                                              │ (retry)   ││
│                                              └───────────┘│
└─────────────────────────────────────────────────────────────┘
```

---

## Login API Integration

```swift
// Login API Service
protocol AuthAPIServiceProtocol {
    func login(email: String, password: String) async throws -> AuthTokensResponse
    func register(email: String, password: String, name: String) async throws -> UserResponse
    func refreshToken(token: String) async throws -> AuthTokensResponse
    func logout(token: String) async throws
    func resetPassword(email: String) async throws
}

class AuthAPIService: AuthAPIServiceProtocol {
    private let apiClient: APIClientProtocol

    init(apiClient: APIClientProtocol = APIClient()) {
        self.apiClient = apiClient
    }

    func login(email: String, password: String) async throws -> AuthTokensResponse {
        let endpoint = AuthEndpoint.login(email: email, password: password)
        return try await apiClient.request(endpoint)
    }

    func register(
        email: String,
        password: String,
        name: String
    ) async throws -> UserResponse {
        let endpoint = AuthEndpoint.register(email: email, password: password, name: name)
        return try await apiClient.request(endpoint)
    }

    func refreshToken(token: String) async throws -> AuthTokensResponse {
        let endpoint = AuthEndpoint.refreshToken(token: token)
        return try await apiClient.request(endpoint)
    }

    func logout(token: String) async throws {
        let endpoint = AuthEndpoint.logout
        let _: EmptyResponse = try await apiClient.request(endpoint)
    }

    func resetPassword(email: String) async throws {
        let endpoint = AuthEndpoint.resetPassword(email: email)
        let _: EmptyResponse = try await apiClient.request(endpoint)
    }
}

// Auth Endpoints
enum AuthEndpoint: Endpoint {
    case login(email: String, password: String)
    case register(email: String, password: String, name: String)
    case refreshToken(token: String)
    case logout
    case resetPassword(email: String)

    var path: String {
        switch self {
        case .login: return "/api/v1/auth/login"
        case .register: return "/api/v1/auth/register"
        case .refreshToken: return "/api/v1/auth/refresh"
        case .logout: return "/api/v1/auth/logout"
        case .resetPassword: return "/api/v1/auth/reset-password"
        }
    }

    var method: HTTPMethod {
        switch self {
        case .login, .register, .refreshToken, .resetPassword: return .post
        case .logout: return .delete
        }
    }

    var headers: [String: String] {
        ["Content-Type": "application/json"]
    }

    var body: Data? {
        switch self {
        case .login(let email, let password):
            return try? JSONSerialization.data(withJSONObject: [
                "email": email,
                "password": password
            ])
        case .register(let email, let password, let name):
            return try? JSONSerialization.data(withJSONObject: [
                "email": email,
                "password": password,
                "name": name
            ])
        case .refreshToken(let token):
            return try? JSONSerialization.data(withJSONObject: [
                "refresh_token": token
            ])
        case .logout:
            return nil
        case .resetPassword(let email):
            return try? JSONSerialization.data(withJSONObject: [
                "email": email
            ])
        }
    }
}
```

### Login API Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    LOGIN API FLOW                            │
│                                                             │
│  ┌──────────┐     ┌──────────────┐     ┌────────────────┐ │
│  │  Email   │────▶│  Validate    │────▶│  POST /login   │ │
│  │ Password │     │  Input       │     │                │ │
│  └──────────┘     └──────────────┘     └───────┬────────┘ │
│                                                │          │
│                                    ┌───────────┴───────┐  │
│                                    │                   │  │
│                                    ▼                   ▼  │
│                            ┌──────────────┐    ┌──────────┐│
│                            │  200 OK      │    │  Error   ││
│                            │  {tokens}    │    │  {error} ││
│                            └──────┬───────┘    └────┬─────┘│
│                                   │                 │      │
│                                   ▼                 ▼      │
│                           ┌──────────────┐    ┌──────────┐│
│                           │  Save Tokens │    │  Show    ││
│                           │  (Keychain)  │    │  Error   ││
│                           └──────┬───────┘    └──────────┘│
│                                   │                        │
│                                   ▼                        │
│                           ┌──────────────┐                │
│                           │  Navigate to │                │
│                           │  Dashboard   │                │
│                           └──────────────┘                │
│                                                           │
│  Response:                                                │
│  {                                                        │
│    "access_token": "eyJhbGci...",                         │
│    "refresh_token": "dGhpcyBpc...",                       │
│    "expires_in": 3600,                                    │
│    "token_type": "Bearer"                                 │
│  }                                                        │
└─────────────────────────────────────────────────────────────┘
```

---

## JWT Token Storage

```swift
// JWT Token Structure
struct JWTToken: Codable {
    let header: JWTHeader
    let payload: JWTPayload
    let signature: String

    struct JWTHeader: Codable {
        let alg: String
        let typ: String
    }

    struct JWTPayload: Codable {
        let sub: String
        let email: String
        let name: String
        let tenantId: String
        let iat: TimeInterval
        let exp: TimeInterval
        let iss: String

        var isExpired: Bool {
            Date().timeIntervalSince1970 > exp
        }

        var timeUntilExpiry: TimeInterval {
            exp - Date().timeIntervalSince1970
        }

        var needsRefresh: Bool {
            timeUntilExpiry < 300 // 5 minutes
        }
    }
}

// Token Parser
class JWTTokenParser {
    static func parse(_ token: String) -> JWTToken? {
        let parts = token.split(separator: ".")
        guard parts.count == 3 else { return nil }

        guard let headerData = Data(base64Encoded: String(parts[0])),
              let payloadData = Data(base64Encoded: String(parts[1])) else {
            return nil
        }

        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase

        guard let header = try? decoder.decode(JWTToken.JWTHeader.self, from: headerData),
              let payload = try? decoder.decode(JWTToken.JWTPayload.self, from: payloadData) else {
            return nil
        }

        return JWTToken(
            header: header,
            payload: payload,
            signature: String(parts[2])
        )
    }
}

// Token Storage in Keychain
class TokenStorage {
    static let shared = TokenStorage()

    private let keychain = KeychainService.shared
    private let accessTokenKey = "com.nexusai.auth.access_token"
    private let refreshTokenKey = "com.nexusai.auth.refresh_token"

    func saveTokens(_ tokens: AuthTokens) throws {
        let encoder = JSONEncoder()

        let accessData = try encoder.encode(tokens)
        try keychain.save(accessData, for: accessTokenKey)

        let refreshData = try encoder.encode(RefreshTokenWrapper(
            refreshToken: tokens.refreshToken,
            expiresAt: tokens.expiryDate
        ))
        try keychain.save(refreshData, for: refreshTokenKey)
    }

    func getAccessToken() throws -> String? {
        guard let data = try keychain.load(for: accessTokenKey),
              let tokens = try? JSONDecoder().decode(AuthTokens.self, from: data) else {
            return nil
        }

        if tokens.isExpired {
            return nil
        }

        return tokens.accessToken
    }

    func getRefreshToken() throws -> String? {
        guard let data = try keychain.load(for: refreshTokenKey),
              let wrapper = try? JSONDecoder().decode(RefreshTokenWrapper.self, from: data) else {
            return nil
        }

        if wrapper.expiresAt < Date() {
            return nil
        }

        return wrapper.refreshToken
    }

    func getTokens() throws -> AuthTokens? {
        guard let data = try keychain.load(for: accessTokenKey) else {
            return nil
        }
        return try JSONDecoder().decode(AuthTokens.self, from: data)
    }

    func clearTokens() throws {
        try keychain.delete(for: accessTokenKey)
        try keychain.delete(for: refreshTokenKey)
    }

    func hasValidTokens() -> Bool {
        guard let tokens = try? getTokens() else { return false }
        return !tokens.isExpired
    }
}

struct RefreshTokenWrapper: Codable {
    let refreshToken: String
    let expiresAt: Date
}
```

### Token Storage Security

```
┌─────────────────────────────────────────────────────────────┐
│                 TOKEN STORAGE SECURITY                        │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                    Keychain                            │  │
│  │                                                      │  │
│  │  ┌────────────────────────────────────────────────┐  │  │
│  │  │  Access Token                                   │  │  │
│  │  │  ─────────────────────────────────────────────  │  │  │
│  │  │  Key: com.nexusai.auth.access_token             │  │  │
│  │  │  Class: kSecClassGenericPassword                │  │  │
│  │  │  Access: kSecAttrAccessibleWhenUnlocked          │  │  │
│  │  │  ThisDeviceOnly: YES                            │  │  │
│  │  └────────────────────────────────────────────────┘  │  │
│  │                                                      │  │
│  │  ┌────────────────────────────────────────────────┐  │  │
│  │  │  Refresh Token                                  │  │  │
│  │  │  ─────────────────────────────────────────────  │  │  │
│  │  │  Key: com.nexusai.auth.refresh_token            │  │  │
│  │  │  Class: kSecClassGenericPassword                │  │  │
│  │  │  Access: kSecAttrAccessibleAfterFirstUnlock      │  │  │
│  │  │  ThisDeviceOnly: YES                            │  │  │
│  │  └────────────────────────────────────────────────┘  │  │
│  │                                                      │  │
│  │  ┌────────────────────────────────────────────────┐  │  │
│  │  │  Biometric Key                                  │  │  │
│  │  │  ─────────────────────────────────────────────  │  │  │
│  │  │  Key: com.nexusai.auth.biometric_key            │  │  │
│  │  │  Class: kSecClassGenericPassword                │  │  │
│  │  │  Access: kSecAttrAccessibleWhenUnlocked          │  │  │
│  │  │  AccessControl: .biometryCurrentSet             │  │  │
│  │  │  ThisDeviceOnly: YES                            │  │  │
│  │  └────────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  Security Properties:                                       │
│  ✅ Hardware-encrypted storage                              │
│  ✅ Not backed up to iCloud                                 │
│  ✅ Requires device unlock                                   │
│  ✅ Biometric protection for sensitive keys                 │
│  ✅ Device-only access (no migration)                       │
└─────────────────────────────────────────────────────────────┘
```

---

## Refresh Token Storage

```swift
// Refresh Token Manager
class RefreshTokenManager: ObservableObject {
    @Published var isRefreshing = false

    private let tokenProvider: TokenProviderProtocol
    private let apiClient: APIClientProtocol
    private let logger: AppLogger

    private var refreshTask: Task<Void, Never>?
    private let lock = NSLock()

    init(
        tokenProvider: TokenProviderProtocol = KeychainTokenProvider(),
        apiClient: APIClientProtocol = APIClient(),
        logger: AppLogger = AppLogger(category: .auth)
    ) {
        self.tokenProvider = tokenProvider
        self.apiClient = apiClient
        self.logger = logger
    }

    func refreshTokenIfNeeded() async throws -> AuthTokens {
        guard let tokens = try await tokenProvider.getTokens() else {
            throw AuthError.noTokens
        }

        if tokens.needsRefresh {
            return try await refreshToken()
        }

        return tokens
    }

    func refreshToken() async throws -> AuthTokens {
        lock.lock()

        if isRefreshing {
            lock.unlock()
            return try await withCheckedThrowingContinuation { continuation in
                // Wait for ongoing refresh
                Task {
                    while isRefreshing {
                        try await Task.sleep(for: .milliseconds(100))
                    }
                    do {
                        let tokens = try await tokenProvider.getTokens()!
                        continuation.resume(returning: tokens)
                    } catch {
                        continuation.resume(throwing: error)
                    }
                }
            }
        }

        isRefreshing = true
        lock.unlock()

        do {
            let tokens = try await tokenProvider.refreshToken()
            logger.info("Token refreshed successfully")

            lock.lock()
            isRefreshing = false
            lock.unlock()

            return tokens
        } catch {
            lock.lock()
            isRefreshing = false
            lock.unlock()

            logger.error("Token refresh failed: \(error.localizedDescription)")
            throw error
        }
    }

    func scheduleTokenRefresh() {
        Task {
            guard let tokens = try await tokenProvider.getTokens() else { return }

            let timeUntilRefresh = tokens.expiryDate.timeIntervalSinceNow - 300
            if timeUntilRefresh > 0 {
                try await Task.sleep(for: .seconds(timeUntilRefresh))
                _ = try? await refreshToken()
            }
        }
    }
}
```

### Token Refresh Timeline

```
┌─────────────────────────────────────────────────────────────┐
│                  TOKEN REFRESH TIMELINE                       │
│                                                             │
│  Time ──────────────────────────────────────────────────▶  │
│                                                             │
│  0s           3300s (55min)    3600s (1hr)                 │
│  │                │                  │                      │
│  ▼                ▼                  ▼                      │
│  ┌──────────────────────────────────────────────┐          │
│  │  Token Valid                                  │          │
│  │  ─────────────────────────────────────────── │          │
│  │  ✅ API calls work                           │          │
│  │  ✅ WebSocket connected                      │          │
│  └──────────────────────────────────────────────┘          │
│                                                             │
│  ┌──────────────────────────────────────────────┐          │
│  │  Token Needs Refresh                          │          │
│  │  ─────────────────────────────────────────── │          │
│  │  ⚠️ Auto-refresh triggered                   │          │
│  │  🔄 Background refresh in progress           │          │
│  └──────────────────────────────────────────────┘          │
│                                                             │
│  ┌──────────────────────────────────────────────┐          │
│  │  Token Expired                                │          │
│  │  ─────────────────────────────────────────── │          │
│  │  ❌ API calls fail with 401                  │          │
│  │  ❌ WebSocket disconnected                   │          │
│  │  🔄 Refresh attempted                         │          │
│  └──────────────────────────────────────────────┘          │
│                                                             │
│  Refresh Strategy:                                          │
│  1. Proactive: Refresh 5 min before expiry                 │
│  2. Reactive: Refresh on 401 response                      │
│  3. Periodic: Background refresh every 55 minutes          │
└─────────────────────────────────────────────────────────────┘
```

---

## Token Refresh Interceptor

```swift
// See 02-networking.md for the full TokenRefreshInterceptor implementation

// Summary of the refresh interceptor flow:

// 1. Request intercepted
//    └── Check if token needs refresh
//        ├── No: Attach current token
//        └── Yes: Trigger refresh
//
// 2. Refresh triggered
//    ├── Is refresh in progress?
//    │   ├── Yes: Queue request
//    │   └── No: Start refresh
//    │
//    └── Refresh completed
//        ├── Success: Resume all queued requests
//        └── Failure: Fail all queued requests
//
// 3. Response received
//    └── Check status code
//        ├── 200-299: Return response
//        ├── 401: Trigger refresh + retry
//        └── Other: Return error
```

---

## Biometric Authentication

```swift
// Biometric Service
import LocalAuthentication

protocol BiometricServiceProtocol {
    var biometricType: BiometricType { get }
    var isAvailable: Bool { get }
    func authenticate(reason: String) async throws -> Bool
    func storeBiometricKey(_ key: String) throws
    func getBiometricKey() throws -> String?
    func deleteBiometricKey() throws
}

enum BiometricType {
    case none
    case touchID
    case faceID
}

enum BiometricError: LocalizedError {
    case notAvailable
    case cancelled
    case fallback
    case lockout
    case passcodeRequired
    case authenticationFailed
    case keychainError(OSStatus)

    var errorDescription: String? {
        switch self {
        case .notAvailable: return "Biometric authentication not available"
        case .cancelled: return "Authentication was cancelled"
        case .fallback: return "User chose to use passcode"
        case .lockout: return "Biometric authentication is locked"
        case .passcodeRequired: return "Passcode is required"
        case .authenticationFailed: return "Authentication failed"
        case .keychainError(let status): return "Keychain error: \(status)"
        }
    }
}

class BiometricService: BiometricServiceProtocol {
    static let shared = BiometricService()

    private let context = LAContext()

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
                    let laError = error as? LAError
                    switch laError?.code {
                    case .userCancel:
                        continuation.resume(throwing: BiometricError.cancelled)
                    case .userFallback:
                        continuation.resume(throwing: BiometricError.fallback)
                    case .biometryLockout:
                        continuation.resume(throwing: BiometricError.lockout)
                    case .biometryNotAvailable:
                        continuation.resume(throwing: BiometricError.notAvailable)
                    case .biometryNotEnrolled:
                        continuation.resume(throwing: BiometricError.notAvailable)
                    default:
                        continuation.resume(throwing: BiometricError.authenticationFailed)
                    }
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
            kSecAttrAccount as String: "com.nexusai.biometric_key",
            kSecValueData as String: key.data(using: .utf8)!,
            kSecAttrAccessControl as String: accessControl
        ]

        SecItemDelete(query as CFDictionary)
        let status = SecItemAdd(query as CFDictionary, nil)

        guard status == errSecSuccess else {
            throw BiometricError.keychainError(status)
        }
    }

    func getBiometricKey() throws -> String? {
        let context = LAContext()
        context.localizedReason = "Access biometric key"

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: "com.nexusai.biometric_key",
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne,
            kSecUseAuthenticationContext as String: context
        ]

        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)

        guard status == errSecSuccess else {
            if status == errSecItemNotFound { return nil }
            throw BiometricError.keychainError(status)
        }

        return item as? Data.flatMap { String(data: $0, encoding: .utf8) }
    }

    func deleteBiometricKey() throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: "com.nexusai.biometric_key"
        ]

        let status = SecItemDelete(query as CFDictionary)
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw BiometricError.keychainError(status)
        }
    }
}
```

### Biometric Authentication Flow

```
┌─────────────────────────────────────────────────────────────┐
│             BIOMETRIC AUTHENTICATION FLOW                    │
│                                                             │
│  ┌──────────┐     ┌──────────────┐     ┌────────────────┐ │
│  │  User    │────▶│  App         │────▶│  LAContext     │ │
│  │  Taps    │     │  Requests    │     │  Evaluate      │ │
│  │  Bio     │     │  Auth        │     │                │ │
│  └──────────┘     └──────────────┘     └───────┬────────┘ │
│                                                │          │
│                                    ┌───────────┴───────┐  │
│                                    │                   │  │
│                                    ▼                   ▼  │
│                            ┌──────────────┐    ┌──────────┐│
│                            │  Face ID /   │    │  User    ││
│                            │  Touch ID    │    │  Prompt  ││
│                            │  Scan        │    │  Shown   ││
│                            └──────┬───────┘    └──────────┘│
│                                   │                        │
│                          ┌────────┴────────┐              │
│                          │                 │              │
│                          ▼                 ▼              │
│                  ┌──────────────┐  ┌──────────────┐      │
│                  │  Success     │  │  Failure     │      │
│                  │              │  │              │      │
│                  └──────┬───────┘  └──────┬───────┘      │
│                         │                 │              │
│                         ▼                 ▼              │
│                 ┌──────────────┐  ┌──────────────┐      │
│                 │  Access      │  │  Show Error  │      │
│                 │  Keychain    │  │  / Retry     │      │
│                 │  with Bio    │  │              │      │
│                 └──────┬───────┘  └──────────────┘      │
│                         │                                │
│                         ▼                                │
│                 ┌──────────────┐                         │
│                 │  Get Token   │                         │
│                 │  from Bio    │                         │
│                 │  Keychain    │                         │
│                 └──────┬───────┘                         │
│                         │                                │
│                         ▼                                │
│                 ┌──────────────┐                         │
│                 │  Authenticate│                         │
│                 │  with Token  │                         │
│                 └──────┬───────┘                         │
│                         │                                │
│                         ▼                                │
│                 ┌──────────────┐                         │
│                 │  Navigate to │                         │
│                 │  Dashboard   │                         │
│                 └──────────────┘                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Biometric Setup Screen

```swift
// BiometricSetupView.swift
struct BiometricSetupView: View {
    @StateObject private var viewModel = BiometricViewModel()
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 32) {
            Spacer()

            // Icon
            Image(systemName: viewModel.biometricIcon)
                .font(.system(size: 80))
                .foregroundStyle(
                    LinearGradient(
                        colors: [.blue, .purple],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )

            // Title
            Text("Enable \(viewModel.biometricName)")
                .font(.largeTitle.bold())

            // Description
            Text("Use \(viewModel.biometricName) for quick and secure sign-in.")
                .font(.body)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)

            // Benefits
            VStack(alignment: .leading, spacing: 16) {
                BenefitRow(icon: "lock.shield", text: "Secure biometric protection")
                BenefitRow(icon: "bolt.fill", text: "Quick one-tap sign-in")
                BenefitRow(icon: "hand.raised.fill", text: "No password required")
            }
            .padding(.horizontal, 32)

            Spacer()

            // Buttons
            VStack(spacing: 12) {
                PrimaryButton(
                    title: "Enable \(viewModel.biometricName)",
                    isLoading: viewModel.isLoading,
                    action: viewModel.enableBiometric
                )

                Button("Skip for now") {
                    dismiss()
                }
                .foregroundColor(.secondary)
            }
            .padding(.horizontal, 24)
            .padding(.bottom, 32)
        }
        .navigationBarTitleDisplayMode(.inline)
        .alert("Success", isPresented: $viewModel.showSuccess) {
            Button("OK") { dismiss() }
        } message: {
            Text("\(viewModel.biometricName) has been enabled successfully.")
        }
        .alert("Error", isPresented: $viewModel.showError) {
            Button("OK") {}
        } message: {
            Text(viewModel.errorMessage ?? "An error occurred")
        }
    }
}

struct BenefitRow: View {
    let icon: String
    let text: String

    var body: some View {
        HStack(spacing: 16) {
            Image(systemName: icon)
                .foregroundColor(.blue)
                .frame(width: 24)
            Text(text)
                .foregroundColor(.primary)
        }
    }
}
```

```swift
// BiometricViewModel
@MainActor
class BiometricViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var showSuccess = false
    @Published var showError = false
    @Published var errorMessage: String?

    private let biometricService: BiometricServiceProtocol
    private let tokenStorage: TokenStorage

    var biometricType: BiometricType {
        biometricService.biometricType
    }

    var biometricName: String {
        switch biometricType {
        case .faceID: return "Face ID"
        case .touchID: return "Touch ID"
        case .none: return "Biometrics"
        }
    }

    var biometricIcon: String {
        switch biometricType {
        case .faceID: return "faceid"
        case .touchID: return "touchid"
        case .none: return "lock.shield"
        }
    }

    init(
        biometricService: BiometricServiceProtocol = BiometricService.shared,
        tokenStorage: TokenStorage = .shared
    ) {
        self.biometricService = biometricService
        self.tokenStorage = tokenStorage
    }

    func enableBiometric() {
        guard biometricService.isAvailable else {
            errorMessage = "Biometrics not available on this device"
            showError = true
            return
        }

        isLoading = true

        Task {
            do {
                // Authenticate with biometrics
                let success = try await biometricService.authenticate(
                    reason: "Enable \(biometricName) for quick sign-in"
                )

                guard success else {
                    isLoading = false
                    return
                }

                // Store biometric key
                guard let accessToken = try tokenStorage.getAccessToken() else {
                    throw AuthError.noTokens
                }

                try biometricService.storeBiometricKey(accessToken)

                showSuccess = true
                AnalyticsManager.shared.log(.biometricEnabled(type: biometricName))
            } catch {
                errorMessage = error.localizedDescription
                showError = true
            }

            isLoading = false
        }
    }
}
```

---

## Biometric Login Flow

```swift
// Biometric Login Flow
class BiometricLoginFlow {
    private let biometricService: BiometricServiceProtocol
    private let tokenProvider: TokenProviderProtocol
    private let authRepository: AuthRepositoryProtocol

    init(
        biometricService: BiometricServiceProtocol = BiometricService.shared,
        tokenProvider: TokenProviderProtocol = KeychainTokenProvider(),
        authRepository: AuthRepositoryProtocol = AuthRepository()
    ) {
        self.biometricService = biometricService
        self.tokenProvider = tokenProvider
        self.authRepository = authRepository
    }

    func login() async throws -> Bool {
        // 1. Check if biometrics are available
        guard biometricService.isAvailable else {
            throw BiometricError.notAvailable
        }

        // 2. Check if biometric key exists
        guard let biometricKey = try biometricService.getBiometricKey() else {
            throw AuthError.biometricNotSetup
        }

        // 3. Prompt for biometric authentication
        let success = try await biometricService.authenticate(
            reason: "Sign in to NexusAI"
        )

        guard success else {
            throw BiometricError.authenticationFailed
        }

        // 4. Use the biometric key to get tokens
        // The key is actually the refresh token stored securely
        let tokens = try await authRepository.refreshToken(biometricKey)

        // 5. Save the new tokens
        try await authRepository.saveTokens(tokens)

        return true
    }
}
```

---

## Secure Storage

```swift
// Keychain Service
class KeychainService {
    static let shared = KeychainService()

    // Save data to Keychain
    func save(_ data: Data, for key: String, accessibility: CFString = kSecAttrAccessibleWhenUnlockedThisDeviceOnly) throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: accessibility
        ]

        // Delete existing item
        SecItemDelete(query as CFDictionary)

        // Add new item
        let status = SecItemAdd(query as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw KeychainError.saveFailed(status)
        }
    }

    // Load data from Keychain
    func load(for key: String) throws -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]

        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)

        switch status {
        case errSecSuccess:
            return item as? Data
        case errSecItemNotFound:
            return nil
        default:
            throw KeychainError.loadFailed(status)
        }
    }

    // Delete data from Keychain
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

    // Delete all Keychain data
    func deleteAll() throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword
        ]

        let status = SecItemDelete(query as CFDictionary)
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw KeychainError.deleteFailed(status)
        }
    }
}

enum KeychainError: LocalizedError {
    case saveFailed(OSStatus)
    case loadFailed(OSStatus)
    case deleteFailed(OSStatus)
    case encodingFailed

    var errorDescription: String? {
        switch self {
        case .saveFailed(let status): return "Keychain save failed: \(status)"
        case .loadFailed(let status): return "Keychain load failed: \(status)"
        case .deleteFailed(let status): return "Keychain delete failed: \(status)"
        case .encodingFailed: return "Failed to encode data"
        }
    }
}
```

### Secure Storage Table

| Storage | Key | Accessibility | Protection |
|---------|-----|---------------|------------|
| Access Token | `auth.access_token` | WhenUnlocked | Device-only |
| Refresh Token | `auth.refresh_token` | AfterFirstUnlock | Device-only |
| Biometric Key | `auth.biometric_key` | WhenUnlocked + Biometric | Bio-protected |
| Tenant ID | `auth.tenant_id` | AfterFirstUnlock | Device-only |
| User ID | `auth.user_id` | AfterFirstUnlock | Device-only |

---

## Session Management

```swift
// Session Manager
class SessionManager: ObservableObject {
    @Published var isAuthenticated = false
    @Published var currentUser: User?

    private let tokenStorage: TokenStorage
    private let authRepository: AuthRepositoryProtocol
    private var inactivityTimer: Timer?
    private var sessionTimer: Timer?
    private let sessionTimeout: TimeInterval = 3600 // 1 hour
    private let inactivityTimeout: TimeInterval = 900 // 15 minutes

    init(
        tokenStorage: TokenStorage = .shared,
        authRepository: AuthRepositoryProtocol = AuthRepository()
    ) {
        self.tokenStorage = tokenStorage
        self.authRepository = authRepository
        checkSession()
    }

    func checkSession() {
        if tokenStorage.hasValidTokens() {
            isAuthenticated = true
            loadCurrentUser()
            startInactivityTimer()
            startSessionTimer()
        } else {
            isAuthenticated = false
        }
    }

    func startInactivityTimer() {
        inactivityTimer?.invalidate()
        inactivityTimer = Timer.scheduledTimer(
            withTimeInterval: inactivityTimeout,
            repeats: false
        ) { [weak self] _ in
            self?.handleInactivityTimeout()
        }
    }

    func resetInactivityTimer() {
        startInactivityTimer()
    }

    private func startSessionTimer() {
        sessionTimer?.invalidate()
        sessionTimer = Timer.scheduledTimer(
            withTimeInterval: sessionTimeout,
            repeats: false
        ) { [weak self] _ in
            self?.handleSessionTimeout()
        }
    }

    private func handleInactivityTimeout() {
        os_log(.info, "Session expired due to inactivity")
        logout()
    }

    private func handleSessionTimeout() {
        os_log(.info, "Session timeout reached")
        logout()
    }

    func logout() {
        inactivityTimer?.invalidate()
        sessionTimer?.invalidate()

        Task {
            try? await authRepository.logout()
            try? tokenStorage.clearTokens()

            await MainActor.run {
                isAuthenticated = false
                currentUser = nil
            }

            AnalyticsManager.shared.log(.logout)
        }
    }

    private func loadCurrentUser() {
        Task {
            do {
                let user = try await authRepository.getCurrentUser()
                await MainActor.run {
                    currentUser = user
                }
            } catch {
                os_log(.error, "Failed to load user: %@", error.localizedDescription)
            }
        }
    }
}
```

---

## Auto-Logout

```swift
// Auto-Logout on Token Expiry
extension SessionManager {
    func setupTokenExpiryMonitoring() {
        Timer.scheduledTimer(withTimeInterval: 60, repeats: true) { [weak self] _ in
            self?.checkTokenExpiry()
        }
    }

    private func checkTokenExpiry() {
        Task {
            do {
                guard let tokens = try tokenStorage.getTokens() else {
                    logout()
                    return
                }

                if tokens.isExpired {
                    // Try to refresh
                    do {
                        let _ = try await authRepository.refreshToken(tokens.refreshToken)
                    } catch {
                        logout()
                    }
                }
            } catch {
                logout()
            }
        }
    }
}

// Background Task for Token Refresh
class BackgroundTokenRefresh {
    func scheduleRefresh() {
        let request = BGAppRefreshTaskRequest(identifier: "com.nexusai.tokenRefresh")
        request.earliestBeginDate = Date(timeIntervalSinceNow: 55 * 60) // 55 minutes

        try? BGTaskScheduler.shared.submit(request)
    }

    func handleRefresh(task: BGAppRefreshTask) {
        let operation = Task {
            do {
                let tokens = try TokenStorage.shared.getTokens()
                if let tokens = tokens, tokens.needsRefresh {
                    let _ = try await AuthRepository().refreshToken(tokens.refreshToken)
                }
                return true
            } catch {
                return false
            }
        }

        task.expirationHandler = {
            operation.cancel()
        }

        Task {
            let success = await operation.value
            task.setTaskCompleted(success: success)
            scheduleRefresh()
        }
    }
}
```

---

## Password Validation

```swift
// Password Validator
struct PasswordValidator {
    struct ValidationResult {
        let isValid: Bool
        let errors: [PasswordError]

        var errorMessages: [String] {
            errors.map { $0.message }
        }
    }

    enum PasswordError: LocalizedError {
        case tooShort(minimum: Int)
        case noUppercase
        case noLowercase
        case noNumber
        case noSymbol
        case commonPassword
        case containsEmail

        var message: String {
            switch self {
            case .tooShort(let min): return "Must be at least \(min) characters"
            case .noUppercase: return "Must contain an uppercase letter"
            case .noLowercase: return "Must contain a lowercase letter"
            case .noNumber: return "Must contain a number"
            case .noSymbol: return "Must contain a special character"
            case .commonPassword: return "This password is too common"
            case .containsEmail: return "Cannot contain your email"
            }
        }
    }

    static func validate(_ password: String, email: String = "") -> ValidationResult {
        var errors: [PasswordError] = []

        // Minimum length
        if password.count < 12 {
            errors.append(.tooShort(minimum: 12))
        }

        // Uppercase
        if !password.contains(where: { $0.isUppercase }) {
            errors.append(.noUppercase)
        }

        // Lowercase
        if !password.contains(where: { $0.isLowercase }) {
            errors.append(.noLowercase)
        }

        // Number
        if !password.contains(where: { $0.isNumber }) {
            errors.append(.noNumber)
        }

        // Symbol
        let symbols = CharacterSet(charactersIn: "!@#$%^&*()_+-=[]{}|;:'\",.<>?/`~")
        if password.unicodeScalars.first(where: { symbols.contains($0) }) == nil {
            errors.append(.noSymbol)
        }

        // Common password
        if commonPasswords.contains(password.lowercased()) {
            errors.append(.commonPassword)
        }

        // Contains email
        if !email.isEmpty, password.lowercased().contains(email.lowercased()) {
            errors.append(.containsEmail)
        }

        return ValidationResult(
            isValid: errors.isEmpty,
            errors: errors
        )
    }

    private static let commonPasswords = [
        "password", "123456", "12345678", "qwerty", "abc123",
        "monkey", "master", "dragon", "login", "princess",
        "football", "shadow", "sunshine", "trustno1", "iloveyou"
    ]
}

// Password Strength Indicator
struct PasswordStrengthView: View {
    let password: String

    var strength: Strength {
        let result = PasswordValidator.validate(password)
        if result.errors.isEmpty { return .strong }
        if result.errors.count <= 2 { return .medium }
        return .weak
    }

    enum Strength: String {
        case weak = "Weak"
        case medium = "Medium"
        case strong = "Strong"

        var color: Color {
            switch self {
            case .weak: return .red
            case .medium: return .orange
            case .strong: return .green
            }
        }
    }

    var body: some View {
        if !password.isEmpty {
            HStack(spacing: 8) {
                ProgressView(value: Double(strengthValue), total: 4)
                    .tint(strength.color)

                Text(strength.rawValue)
                    .font(.caption)
                    .foregroundColor(strength.color)
            }
        }
    }

    private var strengthValue: Int {
        let result = PasswordValidator.validate(password)
        return 4 - result.errors.count
    }
}
```

---

## Error Handling

```swift
// Auth Error Types
enum AuthError: AppError {
    case invalidCredentials
    case invalidEmail
    case passwordTooShort
    case passwordTooWeak(errors: [PasswordValidator.PasswordError])
    case accountLocked
    case tooManyAttempts
    case networkError
    case serverError(String)
    case noTokens
    case tokenExpired
    case refreshTokenExpired
    case biometricNotSetup
    case biometricFailed
    case kycRequired
    case kycPending
    case kycRejected
    case tenantNotFound
    case userNotFound
    case unknown(String)

    var code: Int {
        switch self {
        case .invalidCredentials: return 2001
        case .invalidEmail: return 2002
        case .passwordTooShort: return 2003
        case .passwordTooWeak: return 2004
        case .accountLocked: return 2005
        case .tooManyAttempts: return 2006
        case .networkError: return 2007
        case .serverError: return 2008
        case .noTokens: return 2009
        case .tokenExpired: return 2010
        case .refreshTokenExpired: return 2011
        case .biometricNotSetup: return 2012
        case .biometricFailed: return 2013
        case .kycRequired: return 2014
        case .kycPending: return 2015
        case .kycRejected: return 2016
        case .tenantNotFound: return 2017
        case .userNotFound: return 2018
        case .unknown: return 2099
        }
    }

    var domain: String { "com.nexusai.auth" }

    var errorDescription: String? {
        switch self {
        case .invalidCredentials: return "Invalid email or password"
        case .invalidEmail: return "Please enter a valid email address"
        case .passwordTooShort: return "Password must be at least 12 characters"
        case .passwordTooWeak(let errors): return errors.map(\.message).joined(separator: "\n")
        case .accountLocked: return "Account locked. Please contact support."
        case .tooManyAttempts: return "Too many attempts. Please try again later."
        case .networkError: return "Network error. Please check your connection."
        case .serverError(let msg): return "Server error: \(msg)"
        case .noTokens: return "No authentication tokens found"
        case .tokenExpired: return "Session expired. Please log in again."
        case .refreshTokenExpired: return "Refresh token expired. Please log in again."
        case .biometricNotSetup: return "Biometric authentication not set up"
        case .biometricFailed: return "Biometric authentication failed"
        case .kycRequired: return "KYC verification required"
        case .kycPending: return "KYC verification pending"
        case .kycRejected: return "KYC verification rejected"
        case .tenantNotFound: return "Organization not found"
        case .userNotFound: return "User not found"
        case .unknown(let msg): return msg
        }
    }

    var shouldLogout: Bool {
        switch self {
        case .tokenExpired, .refreshTokenExpired, .noTokens:
            return true
        default:
            return false
        }
    }
}
```

### Error Handling Table

| Error | Code | User Message | Action |
|-------|------|-------------|--------|
| Invalid Credentials | 2001 | Invalid email or password | Show error |
| Account Locked | 2005 | Account locked | Contact support |
| Too Many Attempts | 2006 | Try again in 15 min | Wait + retry |
| Network Error | 2007 | Check connection | Retry |
| Token Expired | 2010 | Session expired | Auto-refresh or logout |
| KYC Required | 2014 | Verification required | Navigate to KYC |
| Biometric Failed | 2013 | Authentication failed | Retry or password |

---

## KYC Verification Flow

```swift
// KYC Verification View
struct KYCVerificationView: View {
    @StateObject private var viewModel = KYCViewModel()

    var body: some View {
        VStack(spacing: 24) {
            // Status Header
            KYCStatusHeader(status: viewModel.status)

            switch viewModel.status {
            case .pending:
                KYCPendingView()

            case .documentUpload:
                DocumentUploadView(
                    onFrontCapture: viewModel.uploadFrontDocument,
                    onBackCapture: viewModel.uploadBackDocument,
                    onSelfieCapture: viewModel.uploadSelfie
                )

            case .submitted:
                KYCSubmittedView()

            case .verified:
                KYCVerifiedView()

            case .rejected:
                KYCRejectedView(reason: viewModel.rejectionReason)
            }
        }
        .navigationTitle("Verification")
        .navigationBarTitleDisplayMode(.inline)
    }
}

struct DocumentUploadView: View {
    let onFrontCapture: () -> Void
    let onBackCapture: () -> Void
    let onSelfieCapture: () -> Void

    var body: some View {
        VStack(spacing: 16) {
            Text("Upload Documents")
                .font(.headline)

            Text("Please provide a valid government-issued ID")
                .font(.subheadline)
                .foregroundColor(.secondary)

            Button(action: onFrontCapture) {
                VStack {
                    Image(systemName: "doc.text")
                        .font(.title2)
                    Text("Front of ID")
                        .font(.subheadline)
                }
                .frame(maxWidth: .infinity)
                .frame(height: 100)
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }

            Button(action: onBackCapture) {
                VStack {
                    Image(systemName: "doc.text")
                        .font(.title2)
                    Text("Back of ID")
                        .font(.subheadline)
                }
                .frame(maxWidth: .infinity)
                .frame(height: 100)
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }

            Button(action: onSelfieCapture) {
                VStack {
                    Image(systemName: "camera")
                        .font(.title2)
                    Text("Take Selfie")
                        .font(.subheadline)
                }
                .frame(maxWidth: .infinity)
                .frame(height: 100)
                .background(Color(.systemGray6))
                .clipShape(RoundedRectangle(cornerRadius: 12))
            }
        }
        .padding()
    }
}
```

---

## Multi-Tenant Support

```swift
// Tenant Manager
class TenantManager: ObservableObject {
    @Published var currentTenant: Tenant?
    @Published var availableTenants: [Tenant] = []

    private let tenantRepository: TenantRepositoryProtocol

    struct Tenant: Identifiable, Codable, Hashable {
        let id: String
        let name: String
        let logo: String?
        let plan: Plan

        enum Plan: String, Codable {
            case free
            case pro
            case enterprise
        }
    }

    init(tenantRepository: TenantRepositoryProtocol = TenantRepository()) {
        self.tenantRepository = tenantRepository
    }

    func loadTenants() async {
        do {
            availableTenants = try await tenantRepository.getTenants()
            if let firstTenant = availableTenants.first {
                currentTenant = firstTenant
            }
        } catch {
            os_log(.error, "Failed to load tenants: %@", error.localizedDescription)
        }
    }

    func switchTenant(_ tenant: Tenant) async {
        currentTenant = tenant

        do {
            try await tenantRepository.switchTenant(tenant.id)
        } catch {
            os_log(.error, "Failed to switch tenant: %@", error.localizedDescription)
        }
    }
}

// Tenant Context
struct TenantContext {
    let tenantId: String
    let tenantName: String
    let plan: TenantManager.Tenant.Plan

    var isEnterprise: Bool { plan == .enterprise }
    var isPro: Bool { plan == .pro || plan == .enterprise }
}
```

---

## Logout Flow

```swift
// Logout Flow
class LogoutManager {
    private let authRepository: AuthRepositoryProtocol
    private let tokenStorage: TokenStorage
    private let sessionManager: SessionManager
    private let coreDataStack: CoreDataStack

    func logout() async {
        // 1. Notify server
        do {
            try await authRepository.logout()
        } catch {
            os_log(.error, "Logout API failed: %@", error.localizedDescription)
        }

        // 2. Clear tokens
        try? tokenStorage.clearTokens()

        // 3. Clear local data
        await clearLocalData()

        // 4. Update session state
        await MainActor.run {
            sessionManager.isAuthenticated = false
            sessionManager.currentUser = nil
        }

        // 5. Clear biometric key
        try? BiometricService.shared.deleteBiometricKey()

        // 6. Log analytics
        AnalyticsManager.shared.log(.logout)

        os_log(.info, "Logout completed")
    }

    private func clearLocalData() async {
        let context = coreDataStack.newBackgroundContext()

        // Delete all messages
        let messageRequest = NSFetchRequest<NSFetchRequestResult>(entityName: "CDMessage")
        let messageDelete = NSBatchDeleteRequest(fetchRequest: messageRequest)
        try? context.execute(messageDelete)

        // Delete all conversations
        let conversationRequest = NSFetchRequest<NSFetchRequestResult>(entityName: "CDConversation")
        let conversationDelete = NSBatchDeleteRequest(fetchRequest: conversationRequest)
        try? context.execute(conversationDelete)

        // Delete all agents
        let agentRequest = NSFetchRequest<NSFetchRequestResult>(entityName: "CDAgent")
        let agentDelete = NSBatchDeleteRequest(fetchRequest: agentRequest)
        try? context.execute(agentDelete)

        try? context.save()
    }
}
```

---

## Auth State Persistence

```swift
// Auth State Manager
@MainActor
class AuthStateManager: ObservableObject {
    @Published var state: AuthState = .unknown

    enum AuthState {
        case unknown
        case authenticated(User)
        case unauthenticated
        case expired
    }

    private let tokenStorage: TokenStorage
    private let authRepository: AuthRepositoryProtocol

    init(
        tokenStorage: TokenStorage = .shared,
        authRepository: AuthRepositoryProtocol = AuthRepository()
    ) {
        self.tokenStorage = tokenStorage
        self.authRepository = authRepository
    }

    func checkAuthState() async {
        state = .unknown

        guard let tokens = try? tokenStorage.getTokens(),
              !tokens.isExpired else {
            state = .unauthenticated
            return
        }

        // Try to load user from cache
        if let user = try? await authRepository.getCurrentUser() {
            state = .authenticated(user)
        } else {
            // Refresh tokens and try again
            do {
                let newTokens = try await authRepository.refreshToken(tokens.refreshToken)
                try tokenStorage.saveTokens(newTokens)

                if let user = try? await authRepository.getCurrentUser() {
                    state = .authenticated(user)
                } else {
                    state = .unauthenticated
                }
            } catch {
                state = .unauthenticated
            }
        }
    }
}
```

---

## Auth Environment Object

```swift
// Auth Environment
@MainActor
class AuthEnvironment: ObservableObject {
    @Published var isAuthenticated = false
    @Published var currentUser: User?
    @Published var currentTenant: TenantManager.Tenant?

    private let sessionManager: SessionManager
    private let tenantManager: TenantManager
    private let logoutManager: LogoutManager

    init(
        sessionManager: SessionManager = SessionManager(),
        tenantManager: TenantManager = TenantManager(),
        logoutManager: LogoutManager = LogoutManager()
    ) {
        self.sessionManager = sessionManager
        self.tenantManager = tenantManager
        self.logoutManager = logoutManager

        // Observe session changes
        sessionManager.$isAuthenticated.assign(to: &$isAuthenticated)
        sessionManager.$currentUser.assign(to: &$currentUser)
        tenantManager.$currentTenant.assign(to: &$currentTenant)
    }

    var currentUserEmail: String {
        currentUser?.email ?? ""
    }

    var currentUserName: String {
        currentUser?.name ?? ""
    }

    var tenantName: String {
        currentTenant?.name ?? "Personal"
    }

    func logout() async {
        await logoutManager.logout()
    }

    func switchTenant(_ tenant: TenantManager.Tenant) async {
        await tenantManager.switchTenant(tenant)
    }
}

// Usage in Views
struct MainTabView: View {
    @EnvironmentObject var authEnvironment: AuthEnvironment

    var body: some View {
        TabView {
            DashboardView()
                .tabItem {
                    Label("Dashboard", systemImage: "house")
                }

            ChatListView()
                .tabItem {
                    Label("Chat", systemImage: "message")
                }

            AgentsView()
                .tabItem {
                    Label("Agents", systemImage: "person.3")
                }

            SettingsView()
                .tabItem {
                    Label("Settings", systemImage: "gear")
                }
        }
        .tint(.accentColor)
    }
}
```

### Auth Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                  AUTH ARCHITECTURE                           │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  PRESENTATION                         │  │
│  │                                                      │  │
│  │  LoginView ──▶ LoginViewModel ──▶ AuthEnvironment    │  │
│  │  RegisterView ──▶ RegisterViewModel                   │  │
│  │  BiometricSetupView ──▶ BiometricViewModel           │  │
│  │  KYCView ──▶ KYCViewModel                            │  │
│  └──────────────────────┬───────────────────────────────┘  │
│                         │                                  │
│                         ▼                                  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  DOMAIN                               │  │
│  │                                                      │  │
│  │  LoginUseCase ──▶ AuthRepositoryProtocol             │  │
│  │  RegisterUseCase ──▶ AuthRepositoryProtocol          │  │
│  │  BiometricUseCase ──▶ AuthRepositoryProtocol         │  │
│  │  RefreshTokenUseCase ──▶ AuthRepositoryProtocol      │  │
│  │  LogoutUseCase ──▶ AuthRepositoryProtocol            │  │
│  └──────────────────────┬───────────────────────────────┘  │
│                         │                                  │
│                         ▼                                  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  DATA                                 │  │
│  │                                                      │  │
│  │  AuthRepository ──┬──▶ AuthAPIService                │  │
│  │                   └──▶ TokenStorage (Keychain)       │  │
│  │                                                      │  │
│  │  AuthInterceptor ──▶ TokenRefreshInterceptor         │  │
│  │  BiometricService ──▶ LAContext                      │  │
│  │  SessionManager ──▶ Timer + CoreData                 │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  SECURITY                             │  │
│  │                                                      │  │
│  │  Keychain ──▶ Tokens, Biometric Keys                 │  │
│  │  LAContext ──▶ Face ID / Touch ID                    │  │
│  │  Certificate Pinning ──▶ TLS Validation              │  │
│  │  Token Expiry ──▶ Auto-refresh                       │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```
