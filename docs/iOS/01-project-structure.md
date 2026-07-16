# iOS Project Structure

## Table of Contents

- [Xcode Project Configuration](#xcode-project-configuration)
- [Swift Package Manager Dependencies](#swift-package-manager-dependencies)
- [Module Structure](#module-structure)
- [Feature Modules](#feature-modules)
- [Domain Layer](#domain-layer)
- [Data Layer](#data-layer)
- [Presentation Layer](#presentation-layer)
- [Core Module](#core-module)
- [DesignSystem Module](#designsystem-module)
- [DI Architecture](#di-architecture)
- [Build Configurations](#build-configurations)
- [Swift Optimization](#swift-optimization)
- [Swift Code Conventions](#swift-code-conventions)
- [Package Dependencies](#package-dependencies)
- [Code Generation](#code-generation)

---

## Xcode Project Configuration

### Targets

| Target | Type | Bundle ID | Platform | Minimum OS |
|--------|------|-----------|----------|------------|
| NexusAI | iOS App | com.nexusai.app | iOS | 17.0 |
| NexusAIUnitTests | Unit Test Bundle | com.nexusai.app.tests | iOS | 17.0 |
| NexusAIUITests | UI Test Bundle | com.nexusai.app.uitests | iOS | 17.0 |
| NexusAICore | Framework | com.nexusai.core | iOS | 17.0 |
| NexusAIDesignSystem | Framework | com.nexusai.designsystem | iOS | 17.0 |
| NexusAIFeatures | Framework | com.nexusai.features | iOS | 17.0 |

### Build Settings

```
┌─────────────────────────────────────────────────────────────┐
│                    BUILD SETTINGS                            │
├─────────────────────────────────────────────────────────────┤
│ Swift Language Version:     5.9                             │
│ Swift Strict Concurrency:   Complete                        │
│ Deployment Target:          iOS 17.0                        │
│ Supported Destinations:     iPhone, iPad                    │
│ Architecture:               arm64                           │
│ Swift Optimization:         -O (Release), -Onone (Debug)   │
│ Enable Module Stability:    YES                             │
│ Code Signing:               Automatic                       │
│ Development Team:           XXXXXXXXXX                      │
│ Product Bundle Identifier:  com.nexusai.app                 │
│ Info.plist:                 Generated                       │
│ Assets:                     Assets.xcassets                 │
└─────────────────────────────────────────────────────────────┘
```

### Schemes

| Scheme | Build Configuration | Test Plans | Arguments |
|--------|-------------------|------------|-----------|
| NexusAI | Debug / Release | Tests.xctestplan | -ENABLE_DEBUG_LOGGING YES |
| NexusAI-Staging | Debug | Tests.xctestplan | -API_BASE_URL https://staging.api.nexusai.com |
| NexusAI-Production | Release | - | -API_BASE_URL https://api.nexusai.com |

### Xcode Project Tree

```
NexusAI.xcodeproj/
├── project.pbxproj
├── xcshareddata/
│   ├── xcschemes/
│   │   ├── NexusAI.xcscheme
│   │   ├── NexusAI-Staging.xcscheme
│   │   └── NexusAI-Production.xcscheme
│   └── TestPlans/
│       └── Tests.xctestplan
└── xcuserdata/
```

---

## Swift Package Manager Dependencies

```swift
// Package.swift
// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "NexusAI",
    platforms: [.iOS(.v17)],
    products: [
        .library(name: "NexusAICore", targets: ["NexusAICore"]),
        .library(name: "NexusAIDesignSystem", targets: ["NexusAIDesignSystem"]),
        .library(name: "NexusAIFeatures", targets: ["NexusAIFeatures"]),
    ],
    dependencies: [
        // MARK: - Networking
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.9.0"),
        .package(url: "https://github.com/onevcat/Kingfisher.git", from: "7.10.0"),

        // MARK: - Architecture
        .package(url: "https://github.com/pointfreeco/swift-composable-architecture", from: "1.7.0"),

        // MARK: - Persistence
        .package(url: "https://github.com/nicklockwood/SwiftFormat", from: "0.52.0"),

        // MARK: - Firebase
        .package(url: "https://github.com/firebase/firebase-ios-sdk.git", from: "10.18.0"),

        // MARK: - Utilities
        .package(url: "https://github.com/krzysztofzablocki/Sourcery.git", from: "2.0.0"),
        .package(url: "https://github.com/SwiftGen/SwiftGen.git", from: "6.6.0"),
        .package(url: "https://github.com/pointfreeco/swift-dependencies", from: "1.2.0"),
        .package(url: "https://github.com/ReactorKit/ReactorKit", from: "3.2.0"),
        .package(url: "https://github.com/Moya/Moya.git", from: "15.0.0"),

        // MARK: - Testing
        .package(url: "https://github.com/Quick/Quick.git", from: "7.0.0"),
        .package(url: "https://github.com/Quick/Nimble.git", from: "13.0.0"),
    ],
    targets: [
        .target(
            name: "NexusAICore",
            dependencies: [
                "Alamofire",
                "Kingfisher",
                "FirebaseAnalytics",
                "FirebaseCrashlytics",
            ]
        ),
        .target(
            name: "NexusAIDesignSystem",
            dependencies: []
        ),
        .target(
            name: "NexusAIFeatures",
            dependencies: [
                "NexusAICore",
                "NexusAIDesignSystem",
                "ComposableArchitecture",
            ]
        ),
        .testTarget(
            name: "NexusAIUnitTests",
            dependencies: [
                "NexusAICore",
                "NexusAIFeatures",
                "Quick",
                "Nimble",
            ]
        ),
        .testTarget(
            name: "NexusAIUITests",
            dependencies: [
                "NexusAIFeatures",
            ]
        ),
    ]
)
```

### Dependency Table

| Package | Version | Purpose | License |
|---------|---------|---------|---------|
| Alamofire | 5.9+ | HTTP networking | MIT |
| Kingfisher | 7.10+ | Image downloading/caching | MIT |
| ComposableArchitecture | 1.7+ | State management | MIT |
| Firebase | 10.18+ | Analytics, Crashlytics | Apache 2.0 |
| Sourcery | 2.0+ | Code generation | MIT |
| SwiftGen | 6.6+ | Asset generation | MIT |
| Quick | 7.0+ | BDD testing | Apache 2.0 |
| Nimble | 13.0+ | Matcher framework | Apache 2.0 |
| SwiftFormat | 0.52+ | Code formatting | MIT |

---

## Module Structure

```
nexus-ai/
├── NexusAI/                                    # Main App Target
│   ├── App/
│   │   ├── NexusAIApp.swift
│   │   ├── AppDelegate.swift
│   │   └── Configuration/
│   │       ├── AppConfig.swift
│   │       ├── EnvironmentConfig.swift
│   │       └── FeatureFlags.swift
│   ├── Resources/
│   │   ├── Assets.xcassets/
│   │   ├── Localizable.xcstrings
│   │   ├── Preview Content/
│   │   └── Info.plist
│   └── NexusAI.entitlements
│
├── NexusAICore/                                # Core Infrastructure
│   ├── Sources/
│   │   ├── Networking/
│   │   │   ├── APIClient.swift
│   │   │   ├── Endpoint.swift
│   │   │   ├── HTTPMethod.swift
│   │   │   ├── Interceptor/
│   │   │   │   ├── AuthInterceptor.swift
│   │   │   │   ├── TokenRefreshInterceptor.swift
│   │   │   │   ├── LoggingInterceptor.swift
│   │   │   │   └── RateLimitInterceptor.swift
│   │   │   ├── Security/
│   │   │   │   ├── CertificatePinning.swift
│   │   │   │   └── NetworkSecurityConfig.swift
│   │   │   └── Monitor/
│   │   │       └── NetworkMonitor.swift
│   │   ├── WebSocket/
│   │   │   ├── WebSocketClient.swift
│   │   │   ├── WebSocketMessage.swift
│   │   │   ├── WebSocketConfiguration.swift
│   │   │   ├── WebSocketReconnect.swift
│   │   │   └── WebSocketHeartbeat.swift
│   │   ├── Security/
│   │   │   ├── KeychainService.swift
│   │   │   ├── BiometricService.swift
│   │   │   ├── CryptoService.swift
│   │   │   └── SecureEnclaveService.swift
│   │   ├── Persistence/
│   │   │   ├── CoreDataStack.swift
│   │   │   ├── CoreDataModel.xcdatamodeld
│   │   │   ├── Entities/
│   │   │   │   ├── CDMessage+CoreDataClass.swift
│   │   │   │   ├── CDConversation+CoreDataClass.swift
│   │   │   │   ├── CDUser+CoreDataClass.swift
│   │   │   │   └── CDAgent+CoreDataClass.swift
│   │   │   └── SyncEngine.swift
│   │   ├── DI/
│   │   │   ├── DependencyContainer.swift
│   │   │   ├── Dependencies.swift
│   │   │   └── EnvironmentKeys.swift
│   │   ├── Logging/
│   │   │   ├── AppLogger.swift
│   │   │   └── LogCategory.swift
│   │   ├── Analytics/
│   │   │   ├── AnalyticsManager.swift
│   │   │   └── AnalyticsEvent.swift
│   │   └── Utilities/
│   │       ├── DateExtensions.swift
│   │       ├── StringExtensions.swift
│   │       ├── ViewExtensions.swift
│   │       └── ColorExtensions.swift
│   └── Tests/
│       ├── NetworkingTests/
│       ├── WebSocketTests/
│       ├── KeychainTests/
│       └── SyncEngineTests/
│
├── NexusAIDesignSystem/                        # UI Components & Theme
│   ├── Sources/
│   │   ├── Components/
│   │   │   ├── Buttons/
│   │   │   │   ├── PrimaryButton.swift
│   │   │   │   ├── SecondaryButton.swift
│   │   │   │   ├── IconButton.swift
│   │   │   │   └── DestructiveButton.swift
│   │   │   ├── TextFields/
│   │   │   │   ├── NexusTextField.swift
│   │   │   │   ├── SecureTextField.swift
│   │   │   │   └── SearchTextField.swift
│   │   │   ├── Cards/
│   │   │   │   ├── ContentCard.swift
│   │   │   │   ├── AgentCard.swift
│   │   │   │   └── StatCard.swift
│   │   │   ├── Alerts/
│   │   │   │   ├── NexusAlert.swift
│   │   │   │   └── ToastView.swift
│   │   │   ├── Indicators/
│   │   │   │   ├── LoadingIndicator.swift
│   │   │   │   ├── ProgressIndicator.swift
│   │   │   │   └── StatusBadge.swift
│   │   │   └── Misc/
│   │   │       ├── AvatarView.swift
│   │   │       ├── BadgeView.swift
│   │   │       └── Divider.swift
│   │   ├── Theme/
│   │   │   ├── ColorTheme.swift
│   │   │   ├── TypographyTheme.swift
│   │   │   ├── SpacingTheme.swift
│   │   │   ├── ShadowTheme.swift
│   │   │   └── AnimationTheme.swift
│   │   └── Tokens/
│   │       ├── Colors.swift
│   │       ├── Fonts.swift
│   │       ├── Dimensions.swift
│   │       └── Icons.swift
│   └── Tests/
│       └── ComponentTests/
│
├── NexusAIFeatures/                            # Feature Modules
│   ├── Sources/
│   │   ├── Auth/
│   │   │   ├── Presentation/
│   │   │   │   ├── Views/
│   │   │   │   │   ├── LoginView.swift
│   │   │   │   │   ├── RegisterView.swift
│   │   │   │   │   ├── BiometricSetupView.swift
│   │   │   │   │   ├── ForgotPasswordView.swift
│   │   │   │   │   └── Components/
│   │   │   │   │       ├── LoginHeader.swift
│   │   │   │   │       ├── EmailField.swift
│   │   │   │   │       └── PasswordField.swift
│   │   │   │   └── ViewModels/
│   │   │   │       ├── LoginViewModel.swift
│   │   │   │       ├── RegisterViewModel.swift
│   │   │   │       └── BiometricViewModel.swift
│   │   │   ├── Domain/
│   │   │   │   ├── Entities/
│   │   │   │   │   ├── User.swift
│   │   │   │   │   ├── AuthTokens.swift
│   │   │   │   │   └── AuthError.swift
│   │   │   │   ├── UseCases/
│   │   │   │   │   ├── LoginUseCase.swift
│   │   │   │   │   ├── RegisterUseCase.swift
│   │   │   │   │   ├── BiometricUseCase.swift
│   │   │   │   │   ├── RefreshTokenUseCase.swift
│   │   │   │   │   └── LogoutUseCase.swift
│   │   │   │   └── Repository/
│   │   │   │       ├── AuthRepository.swift
│   │   │   │       └── AuthRepositoryProtocol.swift
│   │   │   └── Data/
│   │   │       ├── Remote/
│   │   │       │   ├── AuthAPIService.swift
│   │   │       │   └── AuthEndpoints.swift
│   │   │       └── Local/
│   │   │           ├── AuthLocalStore.swift
│   │   │           └── AuthKeychainStore.swift
│   │   ├── Chat/
│   │   │   ├── Presentation/
│   │   │   │   ├── Views/
│   │   │   │   │   ├── ChatView.swift
│   │   │   │   │   ├── ChatListView.swift
│   │   │   │   │   ├── MessageListView.swift
│   │   │   │   │   ├── MessageBubble.swift
│   │   │   │   │   ├── PromptBox.swift
│   │   │   │   │   ├── FileUploader.swift
│   │   │   │   │   ├── VoiceInput.swift
│   │   │   │   │   ├── AgentSelector.swift
│   │   │   │   │   ├── ModelIndicator.swift
│   │   │   │   │   ├── TokenStreamViewer.swift
│   │   │   │   │   ├── ThinkingIndicator.swift
│   │   │   │   │   ├── ToolExecutionDisplay.swift
│   │   │   │   │   └── Components/
│   │   │   │   │       ├── MessageDateHeader.swift
│   │   │   │   │       ├── TypingIndicator.swift
│   │   │   │   │       └── ConnectionStatus.swift
│   │   │   │   └── ViewModels/
│   │   │   │       ├── ChatViewModel.swift
│   │   │   │       └── ChatListViewModel.swift
│   │   │   ├── Domain/
│   │   │   │   ├── Entities/
│   │   │   │   │   ├── Message.swift
│   │   │   │   │   ├── Conversation.swift
│   │   │   │   │   ├── ToolCall.swift
│   │   │   │   │   ├── Attachment.swift
│   │   │   │   │   └── ChatConfig.swift
│   │   │   │   ├── UseCases/
│   │   │   │   │   ├── SendMessageUseCase.swift
│   │   │   │   │   ├── StreamResponseUseCase.swift
│   │   │   │   │   ├── LoadHistoryUseCase.swift
│   │   │   │   │   └── ConversationUseCase.swift
│   │   │   │   └── Repository/
│   │   │   │       ├── ChatRepository.swift
│   │   │   │       └── ChatRepositoryProtocol.swift
│   │   │   └── Data/
│   │   │       ├── Remote/
│   │   │       │   ├── ChatAPIService.swift
│   │   │       │   └── ChatEndpoints.swift
│   │   │       └── Local/
│   │   │           ├── ChatLocalStore.swift
│   │   │           └── MessageCache.swift
│   │   ├── Dashboard/
│   │   │   ├── Presentation/
│   │   │   │   ├── Views/
│   │   │   │   │   ├── DashboardView.swift
│   │   │   │   │   ├── StatsView.swift
│   │   │   │   │   └── ActivityView.swift
│   │   │   │   └── ViewModels/
│   │   │   │       └── DashboardViewModel.swift
│   │   │   ├── Domain/
│   │   │   │   ├── Entities/
│   │   │   │   │   ├── DashboardStats.swift
│   │   │   │   │   └── Activity.swift
│   │   │   │   └── UseCases/
│   │   │   │       └── DashboardUseCase.swift
│   │   │   └── Data/
│   │   │       └── Remote/
│   │   │           └── DashboardAPIService.swift
│   │   ├── Agents/
│   │   │   ├── Presentation/
│   │   │   │   ├── Views/
│   │   │   │   │   ├── AgentsView.swift
│   │   │   │   │   ├── AgentDetailView.swift
│   │   │   │   │   └── AgentCreateView.swift
│   │   │   │   └── ViewModels/
│   │   │   │       └── AgentsViewModel.swift
│   │   │   ├── Domain/
│   │   │   │   ├── Entities/
│   │   │   │   │   ├── Agent.swift
│   │   │   │   │   └── AgentConfig.swift
│   │   │   │   └── UseCases/
│   │   │   │       └── AgentUseCase.swift
│   │   │   └── Data/
│   │   │       └── Remote/
│   │   │           └── AgentAPIService.swift
│   │   ├── Knowledge/
│   │   │   ├── Presentation/
│   │   │   ├── Domain/
│   │   │   └── Data/
│   │   ├── Models/
│   │   │   ├── Presentation/
│   │   │   ├── Domain/
│   │   │   └── Data/
│   │   ├── Settings/
│   │   │   ├── Presentation/
│   │   │   ├── Domain/
│   │   │   └── Data/
│   │   └── Notifications/
│   │       ├── Presentation/
│   │       ├── Domain/
│   │       └── Data/
│   └── Tests/
│       ├── AuthTests/
│       ├── ChatTests/
│       ├── DashboardTests/
│       └── AgentsTests/
│
└── Resources/                                  # Shared Resources
    ├── Assets.xcassets/
    │   ├── Colors/
    │   ├── Images/
    │   ├── Icons/
    │   └── AccentColor.colorset
    ├── Localizable.xcstrings
    └── Info.plist
```

---

## Feature Modules

### Auth Module

```swift
// Auth Module Structure
struct AuthModule {
    static func register(in container: DependencyContainer) {
        // Repository
        container.register(AuthRepositoryProtocol.self) { resolver in
            AuthRepository(
                apiService: resolver.resolve(AuthAPIServiceProtocol.self)!,
                localStore: resolver.resolve(AuthLocalStoreProtocol.self)!
            )
        }

        // Use Cases
        container.register(LoginUseCaseProtocol.self) { resolver in
            LoginUseCase(repository: resolver.resolve(AuthRepositoryProtocol.self)!)
        }

        container.register(RegisterUseCaseProtocol.self) { resolver in
            RegisterUseCase(repository: resolver.resolve(AuthRepositoryProtocol.self)!)
        }

        container.register(BiometricUseCaseProtocol.self) { resolver in
            BiometricUseCase(
                repository: resolver.resolve(AuthRepositoryProtocol.self)!,
                biometricService: resolver.resolve(BiometricServiceProtocol.self)!
            )
        }

        // Services
        container.register(AuthAPIServiceProtocol.self) { _ in
            AuthAPIService()
        }

        container.register(AuthLocalStoreProtocol.self) { _ in
            AuthLocalStore()
        }
    }
}
```

```swift
// Auth Feature File Listing
// Presentation/Views/
//   LoginView.swift           - Email/password login form
//   RegisterView.swift        - New user registration
//   BiometricSetupView.swift  - Face ID/Touch ID setup
//   ForgotPasswordView.swift  - Password reset flow
//   Components/
//     LoginHeader.swift       - App logo and welcome text
//     EmailField.swift        - Validated email input
//     PasswordField.swift     - Password input with toggle
//
// Presentation/ViewModels/
//   LoginViewModel.swift      - Login state management
//   RegisterViewModel.swift   - Registration state management
//   BiometricViewModel.swift  - Biometric state management
//
// Domain/Entities/
//   User.swift                - User model
//   AuthTokens.swift          - JWT token pair
//   AuthError.swift           - Auth-specific errors
//
// Domain/UseCases/
//   LoginUseCase.swift        - Login business logic
//   RegisterUseCase.swift     - Registration business logic
//   BiometricUseCase.swift    - Biometric business logic
//   RefreshTokenUseCase.swift - Token refresh logic
//   LogoutUseCase.swift       - Logout business logic
//
// Domain/Repository/
//   AuthRepository.swift              - Auth data access
//   AuthRepositoryProtocol.swift      - Auth repository interface
//
// Data/Remote/
//   AuthAPIService.swift      - Auth API calls
//   AuthEndpoints.swift       - Auth endpoint definitions
//
// Data/Local/
//   AuthLocalStore.swift      - UserDefaults storage
//   AuthKeychainStore.swift   - Keychain token storage
```

### Chat Module

```swift
// Chat Module Registration
struct ChatModule {
    static func register(in container: DependencyContainer) {
        container.register(ChatRepositoryProtocol.self) { resolver in
            ChatRepository(
                apiService: resolver.resolve(ChatAPIServiceProtocol.self)!,
                webSocketClient: resolver.resolve(WebSocketClientProtocol.self)!,
                localStore: resolver.resolve(ChatLocalStoreProtocol.self)!
            )
        }

        container.register(SendMessageUseCaseProtocol.self) { resolver in
            SendMessageUseCase(
                repository: resolver.resolve(ChatRepositoryProtocol.self)!
            )
        }

        container.register(StreamResponseUseCaseProtocol.self) { resolver in
            StreamResponseUseCase(
                repository: resolver.resolve(ChatRepositoryProtocol.self)!
            )
        }

        container.register(WebSocketClientProtocol.self) { resolver in
            WebSocketClient(
                configuration: resolver.resolve(WebSocketConfiguration.self)!,
                authInterceptor: resolver.resolve(AuthInterceptorProtocol.self)!
            )
        }
    }
}
```

```swift
// Chat Module Feature Files
// Presentation/Views/
//   ChatView.swift                - Main chat container
//   ChatListView.swift            - Conversation list
//   MessageListView.swift         - Messages scroll view
//   MessageBubble.swift           - Individual message cell
//   PromptBox.swift               - Input area
//   FileUploader.swift            - File upload UI
//   VoiceInput.swift              - Voice recording UI
//   AgentSelector.swift           - Agent picker
//   ModelIndicator.swift          - Current model badge
//   TokenStreamViewer.swift       - Real-time tokens
//   ThinkingIndicator.swift       - AI thinking state
//   ToolExecutionDisplay.swift    - Tool call progress
//
// Presentation/ViewModels/
//   ChatViewModel.swift           - Chat state + WebSocket
//   ChatListViewModel.swift       - Conversation list state
//
// Domain/Entities/
//   Message.swift                 - Message model
//   Conversation.swift            - Conversation model
//   ToolCall.swift                - Tool call model
//   Attachment.swift              - File attachment model
//   ChatConfig.swift              - Chat configuration
//
// Domain/UseCases/
//   SendMessageUseCase.swift      - Send message logic
//   StreamResponseUseCase.swift   - Stream response logic
//   LoadHistoryUseCase.swift      - Load chat history
//   ConversationUseCase.swift     - Conversation CRUD
//
// Domain/Repository/
//   ChatRepository.swift          - Chat data access
//   ChatRepositoryProtocol.swift  - Chat repository interface
//
// Data/Remote/
//   ChatAPIService.swift          - Chat API calls
//   ChatEndpoints.swift           - Chat endpoint definitions
//
// Data/Local/
//   ChatLocalStore.swift          - CoreData chat store
//   MessageCache.swift            - In-memory message cache
```

### Dashboard Module

```swift
// Dashboard Module Structure
struct DashboardModule {
    static func register(in container: DependencyContainer) {
        container.register(DashboardUseCaseProtocol.self) { resolver in
            DashboardUseCase(
                apiService: resolver.resolve(DashboardAPIServiceProtocol.self)!
            )
        }
    }
}

// Dashboard Feature Files
// Presentation/Views/
//   DashboardView.swift    - Main dashboard with stats
//   StatsView.swift        - Statistics cards
//   ActivityView.swift     - Recent activity list
//
// Presentation/ViewModels/
//   DashboardViewModel.swift - Dashboard state
//
// Domain/Entities/
//   DashboardStats.swift   - Statistics model
//   Activity.swift         - Activity item model
//
// Domain/UseCases/
//   DashboardUseCase.swift - Dashboard data fetching
//
// Data/Remote/
//   DashboardAPIService.swift - Dashboard API calls
```

### Agents Module

```swift
// Agents Module Structure
// Presentation/Views/
//   AgentsView.swift        - Agent list with search
//   AgentDetailView.swift   - Agent detail and config
//   AgentCreateView.swift   - Create new agent
//
// Presentation/ViewModels/
//   AgentsViewModel.swift   - Agents state management
//
// Domain/Entities/
//   Agent.swift             - Agent model
//   AgentConfig.swift       - Agent configuration
//
// Domain/UseCases/
//   AgentUseCase.swift      - Agent CRUD operations
//
// Data/Remote/
//   AgentAPIService.swift   - Agent API calls
```

### Knowledge Module

```swift
// Knowledge Module Structure
// Presentation/Views/
//   KnowledgeView.swift      - Knowledge base list
//   KnowledgeDetailView.swift - Knowledge entry detail
//   KnowledgeUploadView.swift - Upload new knowledge
//
// Presentation/ViewModels/
//   KnowledgeViewModel.swift - Knowledge state management
//
// Domain/Entities/
//   Knowledge.swift          - Knowledge entry model
//   KnowledgeCategory.swift  - Category model
//
// Domain/UseCases/
//   KnowledgeUseCase.swift   - Knowledge CRUD operations
//
// Data/Remote/
//   KnowledgeAPIService.swift - Knowledge API calls
```

### Models Module

```swift
// Models Module Structure
// Presentation/Views/
//   ModelsView.swift         - Available AI models list
//   ModelDetailView.swift    - Model details and usage
//   ModelComparisonView.swift - Compare models
//
// Presentation/ViewModels/
//   ModelsViewModel.swift    - Models state management
//
// Domain/Entities/
//   AIModel.swift            - AI model definition
//   ModelCapability.swift    - Model capabilities
//   ModelPricing.swift       - Model cost information
//
// Domain/UseCases/
//   ModelUseCase.swift       - Model fetching and selection
//
// Data/Remote/
//   ModelAPIService.swift    - Model API calls
```

### Settings Module

```swift
// Settings Module Structure
// Presentation/Views/
//   SettingsView.swift       - Main settings screen
//   ProfileView.swift        - User profile editing
//   NotificationSettings.swift - Notification preferences
//   AppearanceSettings.swift - Theme and appearance
//   SecuritySettings.swift   - Security options
//   AboutView.swift          - App info and version
//
// Presentation/ViewModels/
//   SettingsViewModel.swift  - Settings state management
//   ProfileViewModel.swift   - Profile state management
//
// Domain/Entities/
//   UserSettings.swift       - Settings model
//   UserPreferences.swift    - User preferences
//
// Domain/UseCases/
//   SettingsUseCase.swift    - Settings persistence
//   ProfileUseCase.swift     - Profile management
//
// Data/Local/
//   SettingsLocalStore.swift - UserDefaults storage
```

### Notifications Module

```swift
// Notifications Module Structure
// Presentation/Views/
//   NotificationsView.swift  - Notification list
//   NotificationDetailView.swift - Notification detail
//
// Presentation/ViewModels/
//   NotificationsViewModel.swift - Notifications state
//
// Domain/Entities/
//   AppNotification.swift    - Notification model
//   NotificationType.swift   - Notification types
//
// Domain/UseCases/
//   NotificationUseCase.swift - Notification handling
//
// Data/Local/
//   NotificationStore.swift  - Notification persistence
```

---

## Domain Layer

### Entities

```swift
// Core entities shared across features
struct User: Identifiable, Codable, Hashable {
    let id: String
    let email: String
    let name: String
    let avatarUrl: String?
    let tenantId: String
    let createdAt: Date
    let updatedAt: Date
    let biometricEnabled: Bool
    let kycStatus: KYCStatus

    enum KYCStatus: String, Codable {
        case pending
        case submitted
        case verified
        case rejected
    }
}

struct AuthTokens: Codable {
    let accessToken: String
    let refreshToken: String
    let expiresIn: TimeInterval
    let tokenType: String

    var isExpired: Bool {
        Date() >= expiryDate
    }

    var expiryDate: Date {
        Date().addingTimeInterval(expiresIn)
    }

    var needsRefresh: Bool {
        Date() >= expiryDate.addingTimeInterval(-300) // 5 min before expiry
    }
}

struct Message: Identifiable, Codable, Hashable {
    let id: UUID
    let content: String
    let role: MessageRole
    let timestamp: Date
    var attachments: [Attachment]
    var toolCalls: [ToolCall]
    var isStreaming: Bool
    var tokens: [String]

    enum MessageRole: String, Codable {
        case user
        case assistant
        case system
        case tool
    }

    var displayedContent: String {
        if isStreaming {
            return tokens.joined()
        }
        return content
    }
}

struct Conversation: Identifiable, Codable, Hashable {
    let id: String
    var title: String
    let agentId: String?
    let modelId: String?
    let createdAt: Date
    var updatedAt: Date
    var messageCount: Int
    var lastMessage: Message?

    var formattedDate: String {
        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .abbreviated
        return formatter.localizedString(for: updatedAt, relativeTo: Date())
    }
}

struct Agent: Identifiable, Codable, Hashable {
    let id: String
    let name: String
    let description: String
    let systemPrompt: String
    let modelId: String
    let avatarUrl: String?
    let capabilities: [String]
    let isDefault: Bool
    let createdAt: Date
}

struct ToolCall: Identifiable, Codable, Hashable {
    let id: String
    let name: String
    let arguments: [String: AnyCodable]
    var status: Status
    var result: String?

    enum Status: String, Codable {
        case pending
        case running
        case completed
        case failed
    }
}

struct Attachment: Identifiable, Codable, Hashable {
    let id: UUID
    let name: String
    let type: AttachmentType
    let url: String?
    let data: Data?
    let size: Int64

    enum AttachmentType: String, Codable {
        case image
        case document
        case audio
        case video
    }
}
```

### Use Cases

```swift
// Use case protocols and implementations
protocol LoginUseCaseProtocol {
    func execute(email: String, password: String) async throws -> AuthTokens
}

class LoginUseCase: LoginUseCaseProtocol {
    private let repository: AuthRepositoryProtocol

    init(repository: AuthRepositoryProtocol) {
        self.repository = repository
    }

    func execute(email: String, password: String) async throws -> AuthTokens {
        guard EmailValidator.isValid(email) else {
            throw AuthError.invalidEmail
        }

        guard password.count >= 12 else {
            throw AuthError.passwordTooShort
        }

        let tokens = try await repository.login(email: email, password: password)
        try await repository.saveTokens(tokens)

        return tokens
    }
}

protocol SendMessageUseCaseProtocol {
    func execute(
        message: String,
        conversationId: String,
        agentId: String?,
        attachments: [Attachment]
    ) async throws -> Message
}

class SendMessageUseCase: SendMessageUseCaseProtocol {
    private let repository: ChatRepositoryProtocol

    init(repository: ChatRepositoryProtocol) {
        self.repository = repository
    }

    func execute(
        message: String,
        conversationId: String,
        agentId: String?,
        attachments: [Attachment]
    ) async throws -> Message {
        let userMessage = Message(
            id: UUID(),
            content: message,
            role: .user,
            timestamp: Date(),
            attachments: attachments,
            toolCalls: [],
            isStreaming: false,
            tokens: []
        )

        try await repository.saveMessage(userMessage, conversationId: conversationId)

        return userMessage
    }
}
```

### Repository Protocols

```swift
// Repository protocol definitions
protocol AuthRepositoryProtocol {
    func login(email: String, password: String) async throws -> AuthTokens
    func register(email: String, password: String, name: String) async throws -> User
    func refreshToken(_ token: String) async throws -> AuthTokens
    func logout() async throws
    func saveTokens(_ tokens: AuthTokens) async throws
    func getTokens() async throws -> AuthTokens?
    func getCurrentUser() async throws -> User?
}

protocol ChatRepositoryProtocol {
    func sendMessage(
        _ message: String,
        conversationId: String,
        agentId: String?
    ) async throws -> Message

    func getConversations(limit: Int, offset: Int) async throws -> [Conversation]
    func getMessages(conversationId: String, limit: Int) async throws -> [Message]
    func createConversation(title: String, agentId: String?) async throws -> Conversation
    func deleteConversation(_ id: String) async throws
    func saveMessage(_ message: Message, conversationId: String) async throws
}

protocol AgentRepositoryProtocol {
    func getAgents() async throws -> [Agent]
    func getAgent(id: String) async throws -> Agent
    func createAgent(_ agent: Agent) async throws -> Agent
    func updateAgent(_ agent: Agent) async throws -> Agent
    func deleteAgent(_ id: String) async throws
}
```

---

## Data Layer

### Remote Data Sources

```swift
// Remote API service implementations
class AuthAPIService: AuthAPIServiceProtocol {
    private let apiClient: APIClientProtocol

    init(apiClient: APIClientProtocol = APIClient()) {
        self.apiClient = apiClient
    }

    func login(email: String, password: String) async throws -> AuthTokensResponse {
        return try await apiClient.request(
            AuthEndpoint.login(email: email, password: password)
        )
    }

    func register(
        email: String,
        password: String,
        name: String
    ) async throws -> UserResponse {
        return try await apiClient.request(
            AuthEndpoint.register(email: email, password: password, name: name)
        )
    }

    func refreshToken(_ token: String) async throws -> AuthTokensResponse {
        return try await apiClient.request(
            AuthEndpoint.refreshToken(token: token)
        )
    }
}

class ChatAPIService: ChatAPIServiceProtocol {
    private let apiClient: APIClientProtocol

    init(apiClient: APIClientProtocol = APIClient()) {
        self.apiClient = apiClient
    }

    func sendMessage(
        _ message: String,
        conversationId: String,
        agentId: String?
    ) async throws -> MessageResponse {
        return try await apiClient.request(
            ChatEndpoint.sendMessage(
                message: message,
                conversationId: conversationId,
                agentId: agentId
            )
        )
    }

    func getConversations(limit: Int, offset: Int) async throws -> [ConversationResponse] {
        return try await apiClient.request(
            ChatEndpoint.getConversations(limit: limit, offset: offset)
        )
    }

    func getMessages(conversationId: String, limit: Int) async throws -> [MessageResponse] {
        return try await apiClient.request(
            ChatEndpoint.getMessages(conversationId: conversationId, limit: limit)
        )
    }
}
```

### Local Data Sources

```swift
// Local persistence implementations
class AuthKeychainStore: AuthLocalStoreProtocol {
    private let keychain = KeychainService.shared

    func saveTokens(_ tokens: AuthTokens) throws {
        let data = try JSONEncoder().encode(tokens)
        try keychain.save(data, for: "auth_tokens")
    }

    func getTokens() throws -> AuthTokens? {
        guard let data = try keychain.load(for: "auth_tokens") else { return nil }
        return try JSONDecoder().decode(AuthTokens.self, from: data)
    }

    func deleteTokens() throws {
        try keychain.delete(for: "auth_tokens")
    }
}

class ChatLocalStore: ChatLocalStoreProtocol {
    private let coreDataStack: CoreDataStack

    init(coreDataStack: CoreDataStack = .shared) {
        self.coreDataStack = coreDataStack
    }

    func saveMessage(_ message: Message, conversationId: String) throws {
        let context = coreDataStack.newBackgroundContext()
        let cdMessage = CDMessage(context: context)
        cdMessage.id = message.id
        cdMessage.content = message.content
        cdMessage.role = message.role.rawValue
        cdMessage.timestamp = message.timestamp
        cdMessage.conversationId = conversationId

        try context.save()
    }

    func getMessages(conversationId: String, limit: Int) throws -> [Message] {
        let context = coreDataStack.context
        let request = CDMessage.fetchRequest()
        request.predicate = NSPredicate(format: "conversationId == %@", conversationId)
        request.sortDescriptors = [NSSortDescriptor(key: "timestamp", ascending: false)]
        request.fetchLimit = limit

        let results = try context.fetch(request)
        return results.map { $0.toDomain() }
    }
}
```

### Repository Implementations

```swift
// Repository implementations that compose remote + local
class AuthRepository: AuthRepositoryProtocol {
    private let apiService: AuthAPIServiceProtocol
    private let localStore: AuthLocalStoreProtocol

    init(
        apiService: AuthAPIServiceProtocol,
        localStore: AuthLocalStoreProtocol
    ) {
        self.apiService = apiService
        self.localStore = localStore
    }

    func login(email: String, password: String) async throws -> AuthTokens {
        let response = try await apiService.login(email: email, password: password)
        return response.toDomain()
    }

    func saveTokens(_ tokens: AuthTokens) async throws {
        try localStore.saveTokens(tokens)
    }

    func getTokens() async throws -> AuthTokens? {
        try localStore.getTokens()
    }

    func logout() async throws {
        try localStore.deleteTokens()
    }
}

class ChatRepository: ChatRepositoryProtocol {
    private let apiService: ChatAPIServiceProtocol
    private let webSocketClient: WebSocketClientProtocol
    private let localStore: ChatLocalStoreProtocol

    init(
        apiService: ChatAPIServiceProtocol,
        webSocketClient: WebSocketClientProtocol,
        localStore: ChatLocalStoreProtocol
    ) {
        self.apiService = apiService
        self.webSocketClient = webSocketClient
        self.localStore = localStore
    }

    func sendMessage(
        _ message: String,
        conversationId: String,
        agentId: String?
    ) async throws -> Message {
        // Save locally first (optimistic)
        let localMessage = Message(
            id: UUID(),
            content: message,
            role: .user,
            timestamp: Date(),
            attachments: [],
            toolCalls: [],
            isStreaming: false,
            tokens: []
        )
        try localStore.saveMessage(localMessage, conversationId: conversationId)

        // Send via WebSocket for streaming
        webSocketClient.send(.chat(
            conversationId: conversationId,
            message: message,
            agentId: agentId,
            attachments: []
        ))

        return localMessage
    }
}
```

### Data Layer Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      DATA LAYER                              │
│                                                             │
│  ┌──────────────────────┐    ┌──────────────────────────┐  │
│  │    REMOTE SOURCES     │    │      LOCAL SOURCES        │  │
│  │                      │    │                          │  │
│  │  ┌────────────────┐ │    │  ┌────────────────────┐ │  │
│  │  │ AuthAPIService │ │    │  │ AuthKeychainStore  │ │  │
│  │  └───────┬────────┘ │    │  └─────────┬──────────┘ │  │
│  │          │          │    │            │            │  │
│  │  ┌───────┴────────┐ │    │  ┌─────────┴──────────┐ │  │
│  │  │ ChatAPIService │ │    │  │ ChatLocalStore     │ │  │
│  │  └───────┬────────┘ │    │  └─────────┬──────────┘ │  │
│  │          │          │    │            │            │  │
│  │  ┌───────┴────────┐ │    │  ┌─────────┴──────────┐ │  │
│  │  │ AgentAPIService│ │    │  │ SettingsLocalStore │ │  │
│  │  └───────┬────────┘ │    │  └─────────┬──────────┘ │  │
│  │          │          │    │            │            │  │
│  └──────────┼──────────┘    └────────────┼────────────┘  │
│             │                            │                │
│             ▼                            ▼                │
│  ┌──────────────────────────────────────────────────────┐│
│  │                  REPOSITORY LAYER                     ││
│  │                                                      ││
│  │  ┌──────────────┐  ┌────────────┐  ┌──────────────┐││
│  │  │AuthRepository│  │ChatRepo    │  │AgentRepo     │││
│  │  └──────────────┘  └────────────┘  └──────────────┘││
│  └──────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

---

## Presentation Layer

### View Components

```swift
// Reusable view components
struct MessageBubble: View {
    let message: Message
    let isStreaming: Bool

    var body: some View {
        HStack(alignment: message.role == .user ? .trailing : .leading) {
            if message.role == .assistant {
                AvatarView(name: "AI", size: 32)
            }

            VStack(alignment: message.role == .user ? .trailing : .leading, spacing: 4) {
                Text(message.displayedContent)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                    .background(message.role == .user ? Color.accentColor : Color(.systemGray5))
                    .foregroundColor(message.role == .user ? .white : .primary)
                    .clipShape(RoundedRectangle(cornerRadius: 16))

                if !message.toolCalls.isEmpty {
                    ForEach(message.toolCalls) { toolCall in
                        ToolExecutionDisplay(toolCall: toolCall)
                    }
                }

                Text(message.timestamp, style: .time)
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
        }
    }
}

struct PromptBox: View {
    @Binding var text: String
    let isStreaming: Bool
    let onSend: () -> Void
    let onAttach: () -> Void
    let onVoice: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            Divider()

            HStack(alignment: .bottom, spacing: 12) {
                Button(action: onAttach) {
                    Image(systemName: "paperclip")
                        .font(.title3)
                        .foregroundColor(.secondary)
                }

                TextField("Message...", text: $text, axis: .vertical)
                    .textFieldStyle(.plain)
                    .lineLimit(1...5)
                    .onSubmit { onSend() }

                Button(action: onVoice) {
                    Image(systemName: "mic.fill")
                        .font(.title3)
                        .foregroundColor(.secondary)
                }

                Button(action: onSend) {
                    Image(systemName: "arrow.up.circle.fill")
                        .font(.title2)
                        .foregroundColor(text.isEmpty ? .gray : .accentColor)
                }
                .disabled(text.isEmpty || isStreaming)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
        }
    }
}
```

### ViewModels

```swift
// ViewModel protocol
@MainActor
protocol ViewModelProtocol: ObservableObject {
    associatedtype State: Equatable
    associatedtype Action

    var state: State { get }
    func send(_ action: Action)
}

// ViewModel base class
@MainActor
class BaseViewModel<State>: ObservableObject {
    @Published var state: State
    @Published var error: Error?
    @Published var isLoading: Bool = false

    private var cancellables = Set<AnyCancellable>()

    init(initialState: State) {
        self.state = initialState
    }

    func handleError(_ error: Error) {
        self.error = error
        isLoading = false
    }

    func withLoading<T>(_ operation: () async throws -> T) async rethrows -> T {
        isLoading = true
        defer { isLoading = false }
        return try await operation()
    }
}
```

---

## Core Module

### Dependency Injection

```swift
// Dependency Injection Container
class DependencyContainer {
    static let shared = DependencyContainer()

    private var factories: [String: () -> Any] = [:]
    private var singletons: [String: Any] = [:]

    func register<T>(_ type: T.Type, factory: @escaping () -> T) {
        let key = String(describing: type)
        factories[key] = factory
    }

    func register<T>(_ type: T.Type, singleton: @autoclosure @escaping () -> T) {
        let key = String(describing: type)
        factories[key] = { [weak self] in
            if let existing = self?.singletons[key] as? T {
                return existing
            }
            let instance = singleton()
            self?.singletons[key] = instance
            return instance
        }
    }

    func resolve<T>(_ type: T.Type) -> T? {
        let key = String(describing: type)
        guard let factory = factories[key] else { return nil }
        return factory() as? T
    }

    func reset() {
        factories.removeAll()
        singletons.removeAll()
    }
}

// Registration helper
extension DependencyContainer {
    func registerAll() {
        // Core Services
        register(APIClientProtocol.self, singleton: APIClient())
        register(WebSocketClientProtocol.self, singleton: WebSocketClient())
        register(KeychainServiceProtocol.self, singleton: KeychainService())
        register(BiometricServiceProtocol.self, singleton: BiometricService())
        register(CoreDataStackProtocol.self, singleton: CoreDataStack.shared)
        register(NetworkMonitorProtocol.self, singleton: NetworkMonitor())
        register(AnalyticsManagerProtocol.self, singleton: AnalyticsManager.shared)

        // Auth
        AuthModule.register(in: self)

        // Chat
        ChatModule.register(in: self)

        // Dashboard
        DashboardModule.register(in: self)

        // Agents
        AgentsModule.register(in: self)
    }
}
```

```swift
// Environment-based DI
struct EnvironmentKeys {
    struct AuthRepository: EnvironmentKey {
        static let defaultValue: AuthRepositoryProtocol = AuthRepository(
            apiService: AuthAPIService(),
            localStore: AuthKeychainStore()
        )
    }

    struct ChatRepository: EnvironmentKey {
        static let defaultValue: ChatRepositoryProtocol = ChatRepository(
            apiService: ChatAPIService(),
            webSocketClient: WebSocketClient(),
            localStore: ChatLocalStore()
        )
    }

    struct NetworkMonitor: EnvironmentKey {
        static let defaultValue: NetworkMonitorProtocol = NetworkMonitor()
    }
}

extension EnvironmentValues {
    var authRepository: AuthRepositoryProtocol {
        get { self[EnvironmentKeys.AuthRepository.self] }
        set { self[EnvironmentKeys.AuthRepository.self] = newValue }
    }

    var chatRepository: ChatRepositoryProtocol {
        get { self[EnvironmentKeys.ChatRepository.self] }
        set { self[EnvironmentKeys.ChatRepository.self] = newValue }
    }

    var networkMonitor: NetworkMonitorProtocol {
        get { self[EnvironmentKeys.NetworkMonitor.self] }
        set { self[EnvironmentKeys.NetworkMonitor.self] = newValue }
    }
}

// Usage in views
struct ChatView: View {
    @Environment(\.chatRepository) private var chatRepository
    @Environment(\.networkMonitor) private var networkMonitor

    @StateObject private var viewModel: ChatViewModel

    init(conversationId: String) {
        _viewModel = StateObject(wrappedValue: ChatViewModel(
            conversationId: conversationId
        ))
    }
}
```

### DI Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                 DEPENDENCY INJECTION                      │
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │              DependencyContainer                    │ │
│  │                                                   │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │ │
│  │  │  Register   │  │  Resolve    │  │  Reset   │ │ │
│  │  │  <T>(:)     │  │  <T>(:)     │  │  ()      │ │ │
│  │  └──────┬──────┘  └──────┬──────┘  └──────────┘ │ │
│  │         │                │                        │ │
│  │  ┌──────▼────────────────▼────────────────────┐  │ │
│  │  │              Factory Registry               │  │ │
│  │  │                                            │  │ │
│  │  │  APIClient ──▶ APIClient()                 │  │ │
│  │  │  AuthRepo ──▶ AuthRepository()             │  │ │
│  │  │  ChatRepo ──▶ ChatRepository()             │  │ │
│  │  │  WebSocket ──▶ WebSocketClient()           │  │ │
│  │  │  Keychain ──▶ KeychainService()            │  │ │
│  │  │  CoreData ──▶ CoreDataStack.shared         │  │ │
│  │  └────────────────────────────────────────────┘  │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │            SwiftUI @Environment                    │ │
│  │                                                   │ │
│  │  @Environment(\\.authRepository)                   │ │
│  │  @Environment(\\.chatRepository)                   │ │
│  │  @EnvironmentObject var appState: AppState        │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

---

## DesignSystem Module

### Theme

```swift
// Color Theme
struct ColorTheme {
    // Primary
    static let primary = Color("Primary", bundle: .module)
    static let primaryLight = Color("PrimaryLight", bundle: .module)
    static let primaryDark = Color("PrimaryDark", bundle: .module)

    // Secondary
    static let secondary = Color("Secondary", bundle: .module)

    // Backgrounds
    static let background = Color("Background", bundle: .module)
    static let surface = Color("Surface", bundle: .module)
    static let surfaceElevated = Color("SurfaceElevated", bundle: .module)

    // Text
    static let textPrimary = Color("TextPrimary", bundle: .module)
    static let textSecondary = Color("TextSecondary", bundle: .module)
    static let textTertiary = Color("TextTertiary", bundle: .module)

    // Status
    static let success = Color("Success", bundle: .module)
    static let warning = Color("Warning", bundle: .module)
    static let error = Color("Error", bundle: .module)
    static let info = Color("Info", bundle: .module)

    // Semantic
    static let userMessage = Color("UserMessage", bundle: .module)
    static let assistantMessage = Color("AssistantMessage", bundle: .module)
}

// Typography Theme
struct TypographyTheme {
    static let largeTitle = Font.largeTitle.bold()
    static let title1 = Font.title.bold()
    static let title2 = Font.title2.bold()
    static let title3 = Font.title3
    static let headline = Font.headline
    static let body = Font.body
    static let callout = Font.callout
    static let subheadline = Font.subheadline
    static let footnote = Font.footnote
    static let caption = Font.caption
    static let caption2 = Font.caption2
}

// Spacing Theme
struct SpacingTheme {
    static let xs: CGFloat = 4
    static let sm: CGFloat = 8
    static let md: CGFloat = 12
    static let lg: CGFloat = 16
    static let xl: CGFloat = 24
    static let xxl: CGFloat = 32
    static let xxxl: CGFloat = 48

    static func padding(_ size: SpacingSize = .md) -> CGFloat {
        switch size {
        case .xs: return xs
        case .sm: return sm
        case .md: return md
        case .lg: return lg
        case .xl: return xl
        case .xxl: return xxl
        case .xxxl: return xxxl
        }
    }

    enum SpacingSize {
        case xs, sm, md, lg, xl, xxl, xxxl
    }
}
```

### Components

```swift
// Primary Button Component
struct PrimaryButton: View {
    let title: String
    let action: () -> Void
    var isLoading: Bool = false
    var isDisabled: Bool = false

    var body: some View {
        Button(action: action) {
            HStack {
                if isLoading {
                    ProgressView()
                        .tint(.white)
                } else {
                    Text(title)
                }
            }
            .frame(maxWidth: .infinity)
            .frame(height: 50)
            .background(isDisabled ? Color.gray : ColorTheme.primary)
            .foregroundColor(.white)
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
        .disabled(isDisabled || isLoading)
    }
}

// Toast View
struct ToastView: View {
    let message: String
    let type: ToastType

    enum ToastType {
        case success, error, warning, info

        var color: Color {
            switch self {
            case .success: return ColorTheme.success
            case .error: return ColorTheme.error
            case .warning: return ColorTheme.warning
            case .info: return ColorTheme.info
            }
        }

        var icon: String {
            switch self {
            case .success: return "checkmark.circle.fill"
            case .error: return "xmark.circle.fill"
            case .warning: return "exclamationmark.triangle.fill"
            case .info: return "info.circle.fill"
            }
        }
    }

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: type.icon)
                .foregroundColor(type.color)

            Text(message)
                .font(.subheadline)
                .foregroundColor(.primary)

            Spacer()
        }
        .padding()
        .background(type.color.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.1), radius: 4, y: 2)
    }
}
```

---

## Build Configurations

```swift
// Environment Configuration
enum BuildConfig {
    case debug
    case staging
    case release

    static var current: BuildConfig {
        #if DEBUG
        return .debug
        #elseif STAGING
        return .staging
        #else
        return .release
        #endif
    }

    var apiBaseURL: String {
        switch self {
        case .debug: return "http://localhost:8080"
        case .staging: return "https://staging.api.nexusai.com"
        case .release: return "https://api.nexusai.com"
        }
    }

    var wsBaseURL: String {
        switch self {
        case .debug: return "ws://localhost:8080"
        case .staging: return "wss://staging.ws.nexusai.com"
        case .release: return "wss://ws.nexusai.com"
        }
    }

    var firebaseConfig: String {
        switch self {
        case .debug: return "GoogleService-Debug.plist"
        case .staging: return "GoogleService-Staging.plist"
        case .release: return "GoogleService-Release.plist"
        }
    }

    var logLevel: LogLevel {
        switch self {
        case .debug: return .verbose
        case .staging: return .info
        case .release: return .error
        }
    }
}
```

### Build Configuration Table

| Configuration | API URL | WebSocket | Logging | Crashlytics | Analytics |
|--------------|---------|-----------|---------|-------------|-----------|
| Debug | localhost:8080 | ws://localhost:8080 | Verbose | Disabled | Test |
| Staging | staging.api.nexusai.com | wss://staging.ws.nexusai.com | Info | Enabled | Staging |
| Release | api.nexusai.com | wss://ws.nexusai.com | Error | Enabled | Production |

---

## Swift Optimization

```swift
// Swift Compiler Optimization Flags
//
// Debug Build (-Onone):
//   - No optimization
//   - Full debug info
//   - Runtime checks enabled
//   - -swift-version 5.9
//   - -enable-upcoming-feature StrictConcurrency
//
// Release Build (-O):
//   - Whole module optimization
//   - -O (optimize for speed)
//   - -whole-module-optimization
//   - -enable-upcoming-feature StrictConcurrency
//   - Dead code stripping
//   - Strip debug symbols
//
// Size-Optimized Build (-Osize):
//   - -Osize (optimize for size)
//   - -swc (swiftc optimization)
//   - Link-Time Optimization (LTO)
//   - Strip unused symbols

// Performance-Optimized Code Examples

// 1. Use value types (structs) over reference types (classes)
struct Message {  // ✅ Value type
    let id: UUID
    let content: String
}

// 2. Use lazy properties for expensive computations
class ChatViewModel {
    lazy var filteredMessages: [Message] = {
        return messages.filter { $0.role == .user }
    }()
}

// 3. Use actor for thread-safe state
actor MessageStore {
    private var messages: [Message] = []

    func append(_ message: Message) {
        messages.append(message)
    }

    func getMessages() -> [Message] {
        messages
    }
}

// 4. Use Sendable for concurrency safety
struct Message: Sendable {
    let id: UUID
    let content: String
    let role: MessageRole
}

// 5. Use @inlinable for performance-critical code
@inlinable
func formatTokenCount(_ count: Int) -> String {
    if count >= 1000 {
        return String(format: "%.1fK", Double(count) / 1000)
    }
    return "\(count)"
}
```

---

## Swift Code Conventions

### Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| Types | PascalCase | `ChatViewModel`, `MessageBubble` |
| Protocols | PascalCase | `ChatRepositoryProtocol` |
| Methods | camelCase | `sendMessage()`, `fetchConversations()` |
| Properties | camelCase | `isLoading`, `errorMessage` |
| Constants | camelCase | `maxRetryCount`, `defaultTimeout` |
| Enums | PascalCase | `MessageRole`, `ChatState` |
| Enum Cases | camelCase | `.user`, `.assistant`, `.streaming` |
| Closures | camelCase | `onSend`, `onError` |
| Private | Prefix `_` | `_isLoading` (internal tracking) |

### File Organization

```swift
// MARK: - File Organization Template

import SwiftUI
// MARK: - Imports (alphabetical)

// MARK: - Type Declaration
struct MyView: View {
    // MARK: - Properties
    @State private var counter = 0
    @Environment(\.dismiss) private var dismiss

    // MARK: - Body
    var body: some View {
        // MARK: View Content
        VStack {
            // ...
        }
        // MARK: - Modifiers
        .navigationTitle("My View")
    }

    // MARK: - Private Methods
    private func loadData() {
        // ...
    }
}

// MARK: - Preview
#Preview {
    MyView()
}
```

### Documentation Standards

```swift
/// Sends a message to the AI assistant via WebSocket.
///
/// This method creates a user message, saves it locally,
/// and sends it through the WebSocket connection for streaming.
///
/// - Parameters:
///   - message: The text content of the message.
///   - conversationId: The ID of the current conversation.
///   - agentId: Optional agent ID to route the message to.
///   - attachments: Array of file attachments to include.
/// - Returns: The created `Message` object.
/// - Throws: `ChatError.networkUnavailable` if offline,
///           `ChatError.messageTooLong` if exceeds limit.
///
/// - Note: Messages are saved optimistically before server confirmation.
/// - Important: Maximum message length is 10,000 characters.
///
/// ```swift
/// let msg = try await chatRepo.sendMessage(
///     "Hello, AI!",
///     conversationId: "conv-123",
///     agentId: "agent-456",
///     attachments: []
/// )
/// ```
func sendMessage(
    _ message: String,
    conversationId: String,
    agentId: String?,
    attachments: [Attachment]
) async throws -> Message
```

---

## Code Generation

### Sourcery Configuration

```yaml
# .sourcery.yml
sources:
  - NexusAICore/Sources
  - NexusAIFeatures/Sources
templates:
  - Templates/
output:
  Generated/
args:
  autoMockable: true
  testableImports: true
```

```swift
// Sourcery Template for Auto-Mocking
// Templates/AutoMockable.stencil

{% for type in types.protocols %}
{% if type.name|hasPrefix:"I" %}

class Mock{{ type.name }}: {{ type.name }} {
    {% for method in type.methods %}
    var {{ method.name }}CallsCount = 0
    var {{ method.name }}ReturnValue: {{ method.returnType|default:"Void" }}!
    var {{ method.name }}Handler: (({{ method.arguments|join:", " }}) -> {{ method.returnType|default:"Void" }})?

    func {{ method.name }}({{ method.arguments|join:", " }}) {{ method.returnType|default:"-> Void" }} {
        {{ method.name }}CallsCount += 1
        return {{ method.name }}Handler ?? {{ method.name }}ReturnValue
    }
    {% endfor %}
}

{% endif %}
{% endfor %}
```

### SwiftGen Configuration

```yaml
# swiftgen.yml
strings:
  inputs:
    - Resources/Localizable.xcstrings
  outputs:
    - templateName: structured-swift5
      output: Generated/Strings.swift

colors:
  inputs:
    - DesignSystem/Resources/Colors/Colors.xcassets
  outputs:
    - templateName: swift5
      output: Generated/Colors.swift

images:
  inputs:
    - Resources/Assets.xcassets
  outputs:
    - templateName: runtime-swift5
      output: Generated/Images.swift
```

### Generated Code Examples

```swift
// Generated/Strings.swift
// swiftlint:disable all
// Generated by SwiftGen — https://github.com/SwiftGen/SwiftGen

extension L10n {
    enum Chat {
        /// Send a message
        static let sendButton = L10n.tr("Localizable", "chat.send_button")
        /// Message...
        static let placeholder = L10n.tr("Localizable", "chat.placeholder")
        /// Type a message...
        static let typing = L10n.tr("Localizable", "chat.typing")
    }

    enum Auth {
        /// Log In
        static let loginButton = L10n.tr("Localizable", "auth.login_button")
        /// Email address
        static let emailPlaceholder = L10n.tr("Localizable", "auth.email_placeholder")
        /// Password
        static let passwordPlaceholder = L10n.tr("Localizable", "auth.password_placeholder")
    }
}
```

```swift
// Generated/Colors.swift
// Generated by SwiftGen

extension Color {
    enum NexusAI {
        static let primary = Color("Primary", bundle: .module)
        static let secondary = Color("Secondary", bundle: .module)
        static let accent = Color("Accent", bundle: .module)
        static let background = Color("Background", bundle: .module)
        static let surface = Color("Surface", bundle: .module)
    }
}
```
