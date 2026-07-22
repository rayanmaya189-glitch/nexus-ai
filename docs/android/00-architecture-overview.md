# Android Architecture Overview

## Table of Contents

1. [Tech Stack](#tech-stack)
2. [Architecture Layers](#architecture-layers)
3. [Module Structure](#module-structure)
4. [Dependency Flow](#dependency-flow)
5. [Navigation Architecture](#navigation-architecture)
6. [State Management](#state-management)
7. [Networking Architecture](#networking-architecture)
8. [Local Storage](#local-storage)
9. [Background Work](#background-work)
10. [Push Notifications](#push-notifications)
11. [Biometric Authentication](#biometric-authentication)
12. [Offline Architecture](#offline-architecture)
13. [Error Handling](#error-handling)
14. [Logging](#logging)
15. [Analytics](#analytics)
16. [Testing Architecture](#testing-architecture)
17. [CI/CD Pipeline](#cicd-pipeline)
18. [Performance](#performance)

---

## Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Language | Kotlin 2.0 | Primary language, coroutines, flows |
| UI Framework | Jetpack Compose | Declarative UI |
| Architecture | Clean Architecture + MVVM | Separation of concerns |
| DI | Hilt (Dagger) | Dependency injection |
| Local DB | Room | SQLite abstraction |
| Networking | OkHttp + Retrofit | HTTP client + REST + Protobuf API |
| WebSocket | OkWs | Real-time streaming |
| Serialization | Kotlinx.serialization | JSON parsing |
| Image Loading | Coil | Compose image loading |
| Navigation | Navigation Compose | Screen navigation |
| Background | WorkManager | Periodic/one-time work |
| Security | EncryptedSharedPreferences | Secure token storage |
| Logging | Timber | Extensible logging |
| Analytics | Firebase Analytics | Event tracking |
| Crash Reporting | Firebase Crashlytics | Crash reporting |
| Testing | JUnit5 + Mockk + Turbine | Unit testing |
| UI Testing | Compose Testing + Espresso | Integration testing |
| CI/CD | GitHub Actions | Automated pipeline |
| Profiling | Baseline Profiles | Startup optimization |

---

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                        Presentation Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Screens   │  │  ViewModels │  │    Compose Components   │ │
│  │  (Compose)  │  │ (StateFlow) │  │  (Reusable UI Widgets)  │ │
│  └──────┬──────┘  └──────┬──────┘  └─────────────┬───────────┘ │
│         │                │                       │             │
│         └────────────────┼───────────────────────┘             │
│                          │                                     │
├──────────────────────────┼─────────────────────────────────────┤
│                     Domain Layer                               │
│  ┌─────────────┐  ┌─────┴──────┐  ┌─────────────────────────┐ │
│  │  Use Cases  │  │  Entities  │  │  Repository Interfaces  │ │
│  │  (Actions)  │  │  (Models)  │  │     (Contracts)         │ │
│  └──────┬──────┘  └────────────┘  └─────────────┬───────────┘ │
│         │                                        │             │
├─────────┼────────────────────────────────────────┼─────────────┤
│                      Data Layer                               │
│  ┌──────┴──────┐  ┌─────────────┐  ┌────────────┴──────────┐  │
│  │  Repository │  │  Remote DS  │  │      Local DS         │  │
│  │   Impl      │  │  (Retrofit) │  │      (Room)           │  │
│  └──────┬──────┘  └──────┬──────┘  └────────────┬──────────┘  │
│         │                │                       │             │
│         └────────────────┼───────────────────────┘             │
│                          │                                     │
├──────────────────────────┼─────────────────────────────────────┤
│                        Core Layer                              │
│  ┌─────────────┐  ┌─────┴──────┐  ┌─────────────────────────┐ │
│  │   Network   │  │  Database  │  │       Security          │ │
│  │  (OkHttp)   │  │   (Room)   │  │  (Keystore, Encrypted)  │ │
│  └─────────────┘  └────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

**Key Principles:**

- **Presentation** observes Domain via `StateFlow`
- **Domain** defines repository interfaces (no implementation details)
- **Data** implements repository interfaces
- **Core** provides cross-cutting infrastructure
- Dependencies flow inward: Presentation → Domain → Data → Core

---

## Module Structure

```
nexus-ai-android/
├── app/                          # Application module
│   ├── build.gradle.kts
│   ├── src/main/
│   │   ├── AndroidManifest.xml
│   │   ├── java/.../App.kt
│   │   └── java/.../MainActivity.kt
│   └── src/test/
├── build-logic/                  # Convention plugins
│   ├── build.gradle.kts
│   └── src/main/kotlin/
│       ├── android-feature.gradle.kts
│       ├── android-library.gradle.kts
│       ├── android-domain.gradle.kts
│       └── android-data.gradle.kts
├── feature/                      # Feature modules
│   ├── feature-chat/
│   ├── feature-dashboard/
│   ├── feature-agents/
│   ├── feature-knowledge/
│   ├── feature-models/
│   ├── feature-settings/
│   ├── feature-auth/
│   └── feature-notifications/
├── domain/                       # Domain module
│   └── domain/
│       ├── src/main/java/
│       │   ├── entity/
│       │   ├── repository/
│       │   ├── usecase/
│       │   └── di/
│       └── build.gradle.kts
├── data/                         # Data module
│   └── data/
│       ├── src/main/java/
│       │   ├── local/
│       │   ├── remote/
│       │   ├── repository/
│       │   └── di/
│       └── build.gradle.kts
├── core/                         # Core infrastructure
│   ├── core-network/
│   ├── core-database/
│   ├── core-security/
│   └── core-utils/
├── common/                       # Shared utilities
│   └── common/
│       ├── src/main/java/
│       │   ├── extension/
│       │   ├── constant/
│       │   └── type/
│       └── build.gradle.kts
└── gradle/
    └── libs.versions.toml        # Version catalog
```

---

## Dependency Flow

```
┌──────────────────────────────────────────────────────────┐
│                     Dependency Graph                      │
│                                                          │
│   app ──────────────────────────────────────────────┐    │
│     ├── feature-chat ──┐                            │    │
│     ├── feature-dash ──┤                            │    │
│     ├── feature-agent ─┤                            │    │
│     ├── feature-model ─┤                            │    │
│     ├── feature-auth ──┼── domain ──┐               │    │
│     ├── feature-know ──┤            │               │    │
│     ├── feature-noti ──┘            │               │    │
│     │                               │               │    │
│     ├── data ───────────────────────┤               │    │
│     │   ├── core-network            │               │    │
│     │   ├── core-database           │               │    │
│     │   └── core-security           │               │    │
│     │                               │               │    │
│     ├── core-network ───────────────┘               │    │
│     ├── core-database ──────────────────────────────┘    │
│     ├── core-security                                    │
│     ├── core-utils                                       │
│     └── common                                           │
└──────────────────────────────────────────────────────────┘
```

**Rule:** Outer layers depend on inner layers. Domain never depends on Data or Core directly.

---

## Navigation Architecture

```kotlin
// NexusNavHost.kt
@Composable
fun NexusNavHost(
    navController: NavHostController,
    authState: AuthState
) {
    NavHost(
        navController = navController,
        startDestination = if (authState is AuthState.Authenticated)
            Screen.Chat.route else Screen.Login.route
    ) {
        // Auth flow
        composable(Screen.Login.route) {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigate(Screen.Chat.route) {
                        popUpTo(Screen.Login.route) { inclusive = true }
                    }
                }
            )
        }

        // Main flow
        navigation(startDestination = Screen.Chat.route, route = "main") {
            composable(Screen.Chat.route) {
                ChatScreen(onNavigateToAgents = { navController.navigate(Screen.Agents.route) })
            }
            composable(Screen.Dashboard.route) {
                DashboardScreen()
            }
            composable(Screen.Agents.route) {
                AgentsScreen(onAgentSelected = { agentId ->
                    navController.navigate("chat/$agentId")
                })
            }
            composable(Screen.Knowledge.route) {
                KnowledgeScreen()
            }
            composable(Screen.Settings.route) {
                SettingsScreen(
                    onLogout = {
                        navController.navigate(Screen.Login.route) {
                            popUpTo(0) { inclusive = true }
                        }
                    }
                )
            }
        }

        // Deep links
        deepLink { uriPattern = "nexus-ai://chat/{agentId}" }
    }
}

sealed class Screen(val route: String) {
    object Login : Screen("login")
    object Chat : Screen("chat")
    object Dashboard : Screen("dashboard")
    object Agents : Screen("agents")
    object Knowledge : Screen("knowledge")
    object Models : Screen("models")
    object Settings : Screen("settings")
    object Notifications : Screen("notifications")
}
```

**Navigation Diagram:**

```
                    ┌─────────┐
                    │  Login  │
                    └────┬────┘
                         │ (success)
                         ▼
                    ┌─────────┐
                    │   Chat  │ ◄─────────────────┐
                    └────┬────┘                    │
                         │                         │
          ┌──────────────┼──────────────┐          │
          ▼              ▼              ▼          │
    ┌───────────┐ ┌────────────┐ ┌──────────┐     │
    │ Dashboard │ │  Agents    │ │  Know.   │     │
    └───────────┘ └──────┬─────┘ └──────────┘     │
                         │ (select)                │
                         ▼                         │
                    ┌─────────┐                    │
                    │Chat/Agent├───────────────────┘
                    └────┬────┘
                         │ (back)
                         ▼
                    ┌─────────┐
                    │ Settings├─── logout ──► Login
                    └─────────┘
```

---

## State Management

```kotlin
// ChatViewModel.kt
@HiltViewModel
class ChatViewModel @Inject constructor(
    private val sendMessageUseCase: SendMessageUseCase,
    private val observeMessagesUseCase: ObserveMessagesUseCase,
    private val webSocketManager: WebSocketManager
) : ViewModel() {

    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    private val _events = Channel<ChatEvent>(Channel.BUFFERED)
    val events: Flow<ChatEvent> = _events.receiveAsFlow()

    init {
        observeMessages()
        observeWebSocketEvents()
    }

    private fun observeMessages() {
        observeMessagesUseCase()
            .onEach { messages ->
                _uiState.update { it.copy(messages = messages) }
            }
            .launchIn(viewModelScope)
    }

    private fun observeWebSocketEvents() {
        webSocketManager.events
            .onEach { event ->
                when (event) {
                    is WsEvent.Token -> handleToken(event)
                    is WsEvent.Thinking -> handleThinking(event)
                    is WsEvent.ToolCall -> handleToolCall(event)
                    is WsEvent.Completed -> handleCompleted()
                    is WsEvent.Error -> handleError(event)
                    is WsEvent.Connected -> handleConnected()
                    is WsEvent.Disconnected -> handleDisconnected()
                }
            }
            .launchIn(viewModelScope)
    }

    fun onAction(action: ChatAction) {
        when (action) {
            is ChatAction.SendMessage -> sendMessage(action.text)
            is ChatAction.UploadFile -> uploadFile(action.uri)
            is ChatAction.SelectAgent -> selectAgent(action.agentId)
            is ChatAction.CancelGeneration -> cancelGeneration()
            is ChatAction.RetryMessage -> retryMessage(action.messageId)
            is ChatAction.DeleteMessage -> deleteMessage(action.messageId)
            is ChatAction.LoadMore -> loadMoreHistory()
        }
    }

    private fun sendMessage(text: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            sendMessageUseCase(text)
                .onFailure { error ->
                    _events.send(ChatEvent.ShowError(error.message))
                    _uiState.update { it.copy(isLoading = false) }
                }
        }
    }
}

data class ChatUiState(
    val messages: List<Message> = emptyList(),
    val isLoading: Boolean = false,
    val isStreaming: Boolean = false,
    val streamingText: String = "",
    val selectedAgent: Agent? = null,
    val isWebSocketConnected: Boolean = false,
    val error: String? = null
)

sealed class ChatAction {
    data class SendMessage(val text: String) : ChatAction()
    data class UploadFile(val uri: Uri) : ChatAction()
    data class SelectAgent(val agentId: String) : ChatAction()
    data object CancelGeneration : ChatAction()
    data class RetryMessage(val messageId: String) : ChatAction()
    data class DeleteMessage(val messageId: String) : ChatAction()
    data object LoadMore : ChatAction()
}

sealed class ChatEvent {
    data class ShowError(val message: String?) : ChatEvent()
    data class NavigateToLogin(val reason: String) : ChatEvent()
}
```

**State Flow Diagram:**

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Action     │────►│  ViewModel   │────►│  UiState     │
│  (UserInput) │     │  (Process)   │     │  (StateFlow) │
└──────────────┘     └──────┬───────┘     └──────┬───────┘
                            │                     │
                            │                     ▼
                     ┌──────┴───────┐     ┌──────────────┐
                     │   Use Case   │     │   Compose    │
                     │  (Business)  │     │   (Render)   │
                     └──────┬───────┘     └──────────────┘
                            │
                            ▼
                     ┌──────────────┐
                     │ Repository   │
                     │   (Data)     │
                     └──────────────┘
```

---

## Networking Architecture

```kotlin
// NetworkModule.kt
@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    @Provides
    @Singleton
    fun provideOkHttpClient(
        authInterceptor: AuthInterceptor,
        tokenRefreshInterceptor: TokenRefreshInterceptor,
        loggingInterceptor: HttpLoggingInterceptor
    ): OkHttpClient {
        return OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(60, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .connectionPool(ConnectionPool(5, 5, TimeUnit.MINUTES))
            .addInterceptor(authInterceptor)
            .addInterceptor(tokenRefreshInterceptor)
            .addInterceptor(loggingInterceptor)
            .build()
    }

    @Provides
    @Singleton
    fun provideRetrofit(okHttpClient: OkHttpClient): Retrofit {
        return Retrofit.Builder()
            .baseUrl(BuildConfig.BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(
                Json.asConverterFactory("application/json".toMediaType())
            )
            .build()
    }

    @Provides
    @Singleton
    fun provideApiService(retrofit: Retrofit): ApiService {
        return retrofit.create(ApiService::class.java)
    }
}
```

**Network Request Flow:**

```
┌───────────┐    ┌──────────────┐    ┌──────────────┐
│  App Code │───►│  Auth Intcpt │───►│ TokenRefresh │
└───────────┘    │  (Add Token) │    │  (401 Retry) │
                 └──────┬───────┘    └──────┬───────┘
                        │                    │
                        ▼                    ▼
                 ┌──────────────┐    ┌──────────────┐
                 │   Logging    │───►│   OkHttp     │
                 │  (Debug Log) │    │  (Execute)   │
                 └──────────────┘    └──────┬───────┘
                                            │
                                     ┌──────┴───────┐
                                     │   Server     │
                                     └──────────────┘
```

---

## Local Storage

```kotlin
// DatabaseModule.kt
@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideDatabase(@ApplicationContext context: Context): NexusDatabase {
        return Room.databaseBuilder(
            context,
            NexusDatabase::class.java,
            "nexus_ai.db"
        )
        .addMigrations(MIGRATION_1_2, MIGRATION_2_3)
        .build()
    }

    @Provides
    fun provideMessageDao(db: NexusDatabase): MessageDao = db.messageDao()

    @Provides
    fun provideConversationDao(db: NexusDatabase): ConversationDao = db.conversationDao()
}

// SecurityModule.kt
@Module
@InstallIn(SingletonComponent::class)
object SecurityModule {

    @Provides
    @Singleton
    fun provideEncryptedSharedPreferences(
        @ApplicationContext context: Context
    ): SharedPreferences {
        return EncryptedSharedPreferences.create(
            "nexus_secure_prefs",
            MasterKey.Builder(context)
                .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
                .build(),
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        )
    }

    @Provides
    @Singleton
    fun provideDataStore(
        @ApplicationContext context: Context
    ): DataStore<Preferences> {
        return context.createDataStore(name = "nexus_preferences")
    }
}
```

**Storage Hierarchy:**

```
┌─────────────────────────────────────────────────────┐
│                   Storage Options                    │
├─────────────┬───────────────┬───────────────────────┤
│   Room DB   │  DataStore    │ EncryptedSharedPrefs   │
│             │               │                       │
│ Messages    │ User Prefs    │ Access Token          │
│ Conversations│ Theme        │ Refresh Token         │
│ Agents      │ Language      │ Biometric Key         │
│ Knowledge   │ Last Sync     │ API Key               │
│             │ Onboarding    │ User ID               │
│             │               │                       │
│ Structured  │ Key-Value     │ Encrypted Key-Value   │
│ Queryable   │ Async         │ Hardware-Backed       │
└─────────────┴───────────────┴───────────────────────┘
```

---

## Background Work

```kotlin
// SyncWorker.kt
@HiltWorker
class SyncWorker @AssistedInject constructor(
    @Assisted context: Context,
    @Assisted params: WorkerParameters,
    private val syncRepository: SyncRepository
) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result {
        return try {
            syncRepository.syncAll()
            Result.success()
        } catch (e: Exception) {
            if (runAttemptCount < 3) Result.retry()
            else Result.failure()
        }
    }
}

// WorkManager Setup
@Singleton
class BackgroundWorkManager @Inject constructor(
    private val workManager: WorkManager
) {
    fun schedulePeriodicSync() {
        val syncRequest = PeriodicWorkRequestBuilder<SyncWorker>(
            15, TimeUnit.MINUTES
        )
            .setConstraints(
                Constraints.Builder()
                    .setRequiredNetworkType(NetworkType.CONNECTED)
                    .setRequiresBatteryNotLow(true)
                    .build()
            )
            .setBackoffCriteria(
                BackoffPolicy.EXPONENTIAL,
                30, TimeUnit.SECONDS
            )
            .build()

        workManager.enqueueUniquePeriodicWork(
            "nexus_sync",
            ExistingPeriodicWorkPolicy.KEEP,
            syncRequest
        )
    }
}
```

---

## Push Notifications

```kotlin
// FirebaseMessagingService.kt
@AndroidEntryPoint
class NexusMessagingService : FirebaseMessagingService() {

    @Inject lateinit var notificationManager: NexusNotificationManager

    override fun onNewToken(token: String) {
        super.onNewToken(token)
        CoroutineScope(Dispatchers.IO).launch {
            notificationManager.registerToken(token)
        }
    }

    override fun onMessageReceived(message: RemoteMessage) {
        super.onMessageReceived(message)
        notificationManager.handleRemoteMessage(message)
    }
}

// NexusNotificationManager.kt
@Singleton
class NexusNotificationManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val apiService: ApiService
) {
    fun handleRemoteMessage(message: RemoteMessage) {
        val notification = NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle(message.data["title"])
            .setContentText(message.data["body"])
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setAutoCancel(true)
            .build()

        notificationManager.notify(
            message.data["id"]?.toInt() ?: System.currentTimeMillis().toInt(),
            notification
        )
    }
}
```

---

## Biometric Authentication

```kotlin
// BiometricManager.kt
@Singleton
class BiometricManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val keyStore: AndroidKeyStore
) {
    fun promptBiometric(
        activity: FragmentActivity,
        onSuccess: () -> Unit,
        onError: (String) -> Unit
    ) {
        val promptInfo = BiometricPrompt.PromptInfo.Builder()
            .setTitle("Authenticate")
            .setSubtitle("Use your fingerprint to login")
            .setAllowedAuthenticators(BiometricManager.Authenticators.BIOMETRIC_STRONG)
            .setNegativeButtonText("Use Password")
            .build()

        val biometricPrompt = BiometricPrompt(
            activity,
            ContextCompat.getMainExecutor(context),
            object : BiometricPrompt.AuthenticationCallback() {
                override fun onAuthenticationSucceeded(result: AuthenticationResult) {
                    super.onAuthenticationSucceeded(result)
                    onSuccess()
                }

                override fun onAuthenticationError(errorCode: Int, errString: CharSequence) {
                    super.onAuthenticationError(errorCode, errString)
                    onError(errString.toString())
                }
            }
        )

        biometricPrompt.authenticate(promptInfo)
    }
}
```

---

## Offline Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Offline Sync Flow                      │
│                                                         │
│  ┌───────────┐     ┌───────────┐     ┌───────────┐     │
│  │  User     │────►│   Room    │────►│  Pending  │     │
│  │  Action   │     │   DB      │     │  Queue    │     │
│  └───────────┘     └───────────┘     └─────┬─────┘     │
│                                            │           │
│                                     ┌──────┴───────┐   │
│                                     │  Network     │   │
│                                     │  Available?  │   │
│                                     └──────┬───────┘   │
│                                      Yes   │   No      │
│                                       │    │    │      │
│                                       ▼    │    ▼      │
│                                ┌──────────┐│  ┌──────┐  │
│                                │  Sync    ││  │ Wait │  │
│                                │  Now     ││  └──────┘  │
│                                └──────────┘│            │
└─────────────────────────────────────────────────────────┘
```

---

## Error Handling

```kotlin
// Result wrapper
sealed class Result<out T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val exception: Throwable) : Result<Nothing>()
    data object Loading : Result<Nothing>()

    val isSuccess get() = this is Success
    val isError get() = this is Error
    val isLoading get() = this is Loading

    fun getOrNull(): T? = (this as? Success)?.data
    fun exceptionOrNull(): Throwable? = (this as? Error)?.exception
}

// Extension functions
suspend fun <T> safeCall(block: suspend () -> T): Result<T> {
    return try {
        Result.Success(block())
    } catch (e: HttpException) {
        Result.Error(ApiException(e.code(), e.message(), e))
    } catch (e: IOException) {
        Result.Error(NetworkException(e))
    } catch (e: Exception) {
        Result.Error(e)
    }
}

// Domain use case pattern
class SendMessageUseCase @Inject constructor(
    private val messageRepository: MessageRepository,
    private val webSocketManager: WebSocketManager
) {
    operator fun invoke(text: String): Flow<Result<Message>> = flow {
        emit(Result.Loading)
        val message = messageRepository.createLocalMessage(text)
        emit(Result.Success(message))
        webSocketManager.sendMessage(text)
    }
}
```

**Error Hierarchy:**

```
Throwable
├── ApiException (HTTP errors)
│   ├── 401 Unauthorized
│   ├── 403 Forbidden
│   ├── 404 Not Found
│   ├── 429 Rate Limited
│   └── 500 Server Error
├── NetworkException (Connectivity)
│   ├── NoInternetException
│   ├── TimeoutException
│   └── SSLException
├── AuthException (Auth errors)
│   ├── TokenExpiredException
│   ├── InvalidCredentialsException
│   └── AccountLockedException
├── DatabaseException (Room errors)
│   ├── MigrationException
│   └── ConstraintException
└── ValidationException (Input errors)
    ├── InvalidEmailException
    ├── WeakPasswordException
    └── EmptyFieldException
```

---

## Logging

```kotlin
// LoggingModule.kt
@Module
@InstallIn(SingletonComponent::class)
object LoggingModule {

    @Provides
    @Singleton
    fun provideTimberTree(): Timber.Tree {
        return if (BuildConfig.DEBUG) {
           object : Timber.DebugTree() {
                override fun createStackElementTag(element: StackTraceElement): String {
                    return "(${element.fileName}:${element.lineNumber})#${element.methodName}"
                }
            }
        } else {
            CrashReportingTree()
        }
    }
}

class CrashReportingTree : Timber.Tree() {
    override fun log(priority: Int, tag: String?, message: String, t: Throwable?) {
        if (priority < Log.WARN) return

        if (t != null) {
            Firebase.crashlytics.recordException(t)
        }
        if (priority >= Log.ERROR) {
            Firebase.crashlytics.log(message)
        }
    }
}

// Usage
Timber.d("Sending message: %s", text)
Timber.e(exception, "WebSocket connection failed")
Timber.w("Token expiring soon, refreshing...")
```

---

## Analytics

```kotlin
// AnalyticsModule.kt
@Singleton
class AnalyticsManager @Inject constructor(
    private val firebaseAnalytics: FirebaseAnalytics
) {
    fun logEvent(event: AnalyticsEvent) {
        firebaseAnalytics.logEvent(event.name) {
            event.params.forEach { (key, value) ->
                when (value) {
                    is String -> param(key, value)
                    is Int -> param(key, value.toLong())
                    is Long -> param(key, value)
                    is Double -> param(key, value)
                    is Boolean -> param(key, if (value) "true" else "false")
                }
            }
        }
    }

    fun logScreenView(screenName: String) {
        logEvent(AnalyticsEvent.SCREEN_VIEW(mapOf("screen_name" to screenName)))
    }

    fun logChatMessage(agentId: String, modelId: String, messageLength: Int) {
        logEvent(AnalyticsEvent.CHAT_MESSAGE(mapOf(
            "agent_id" to agentId,
            "model_id" to modelId,
            "message_length" to messageLength
        )))
    }
}

sealed class AnalyticsEvent(val name: String, val params: Map<String, Any>) {
    class SCREEN_VIEW(params: Map<String, Any>) : AnalyticsEvent("screen_view", params)
    class CHAT_MESSAGE(params: Map<String, Any>) : AnalyticsEvent("chat_message", params)
    class USER_LOGIN(params: Map<String, Any>) : AnalyticsEvent("user_login", params)
    class AGENT_SELECTED(params: Map<String, Any>) : AnalyticsEvent("agent_selected", params)
}
```

---

## Testing Architecture

```kotlin
// Unit Test
@ExtendWith(MockkExtension::class)
class SendMessageUseCaseTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Mockk lateinit var messageRepository: MessageRepository
    @Mockk lateinit var webSocketManager: WebSocketManager
    private lateinit var useCase: SendMessageUseCase

    @BeforeEach
    fun setup() {
        useCase = SendMessageUseCase(messageRepository, webSocketManager)
    }

    @Test
    fun `should create message and send via websocket`() = runTest {
        // Given
        val text = "Hello, AI"
        val expectedMessage = Message(id = "1", content = text, role = Role.USER)
        coEvery { messageRepository.createLocalMessage(text) } returns expectedMessage
        every { webSocketManager.sendMessage(text) } returns Unit

        // When
        useCase(text).test {
            assertThat(awaitItem()).isInstanceOf(Result.Loading::class.java)
            assertThat(awaitItem()).isEqualTo(Result.Success(expectedMessage))
            cancelAndIgnoreRemainingEvents()
        }
    }
}

// Compose UI Test
@HiltAndroidTest
class ChatScreenTest {

    @get:Rule
    val hiltRule = HiltAndroidRule(this)

    @get:Rule
    val composeRule = createAndroidComposeRule<MainActivity>()

    @Inject lateinit var viewModel: ChatViewModel

    @Before
    fun setup() {
        hiltRule.inject()
    }

    @Test
    fun shouldDisplayMessages() {
        composeRule.setContent {
            NexusTheme {
                ChatScreen()
            }
        }

        composeRule.onNodeWithText("Hello AI").assertIsDisplayed()
        composeRule.onNodeWithTag("message_input").assertExists()
        composeRule.onNodeWithTag("send_button").assertIsEnabled()
    }

    @Test
    fun shouldSendMessage() {
        composeRule.setContent {
            NexusTheme {
                ChatScreen()
            }
        }

        composeRule.onNodeWithTag("message_input").performTextInput("Hello")
        composeRule.onNodeWithTag("send_button").performClick()
        composeRule.onNodeWithText("Hello").assertIsDisplayed()
    }
}
```

**Testing Pyramid:**

```
                    ╱╲
                   ╱  ╲         E2E Tests (5%)
                  ╱    ╲        Espresso, Compose
                 ╱──────╲
                ╱        ╲      Integration Tests (15%)
               ╱          ╲     ViewModel + Repository
              ╱────────────╲
             ╱              ╲   Unit Tests (80%)
            ╱                ╲  Use Cases, Domain Logic
           ╱──────────────────╲
```

---

## CI/CD Pipeline

```yaml
# .github/workflows/android.yml
name: Android CI/CD

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up JDK 21
        uses: actions/setup-java@v4
        with:
          java-version: '21'
          distribution: 'temurin'

      - name: Cache Gradle
        uses: actions/cache@v4
        with:
          path: |
            ~/.gradle/caches
            ~/.gradle/wrapper
          key: gradle-${{ runner.os }}-${{ hashFiles('**/*.gradle*') }}

      - name: Run Tests
        run: ./gradlew testDebugUnitTest

      - name: Run Lint
        run: ./gradlew lintDebug

      - name: Build APK
        run: ./gradlew assembleRelease

      - name: Upload APK
        uses: actions/upload-artifact@v4
        with:
          name: app-release
          path: app/build/outputs/apk/release/

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy to Play Store
        uses: r0adkll/upload-google-play@v1
        with:
          serviceAccountJsonPlainText: ${{ secrets.PLAY_SERVICE_ACCOUNT }}
          packageName: com.nexus.ai
          releaseFiles: app/build/outputs/apk/release/app-release.apk
          track: production
```

---

## Performance

### Baseline Profiles

```kotlin
// BaselineProfileGenerator.kt
@RunWith(AndroidJUnit4::class)
class BaselineProfileGenerator {

    @get:Rule
    val rule = BaselineProfileRule()

    @Test
    fun generate() {
        rule.collect(
            packageName = "com.nexus.ai",
            includeInStartupProfile = true
        ) {
            pressHome()
            startActivityAndWait()

            // Navigate to chat
            onNodeWithTag("chat_tab").performClick()
            waitForIdle()

            // Send a message
            onNodeWithTag("message_input").performTextInput("Hello")
            onNodeWithTag("send_button").performClick()
            waitForIdle()
        }
    }
}
```

### R8 Rules

```proguard
# proguard-rules.pro
-keepattributes Signature
-keepattributes *Annotation*

# Kotlinx.serialization
-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.AnnotationsKt

-keepclassmembers @kotlinx.serialization.Serializable class ** {
    *** Companion;
}
-keepclasseswithmembers class **$$serializer {
    *** INSTANCE;
}

# Room
-keep class * extends androidx.room.RoomDatabase
-keep @androidx.room.Entity class *
-dontwarn androidx.room.paging.**

# Retrofit
-keepattributes Exceptions
-keepattributes Signature
-keepattributes GenericSignature
-dontwarn retrofit2.**
-keep class retrofit2.** { *; }

# Hilt
-keep class dagger.hilt.** { *; }
-keep class javax.inject.** { *; }
```

### Memory Leak Detection

```kotlin
// DebugApplication.kt
@HiltAndroidApp
class DebugApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        if (BuildConfig.DEBUG) {
            StrictMode.setThreadPolicy(
                StrictMode.ThreadPolicy.Builder()
                    .detectAll()
                    .penaltyLog()
                    .build()
            )
            StrictMode.setVmPolicy(
                StrictMode.VmPolicy.Builder()
                    .detectLeakedSqlLiteObjects()
                    .detectLeakedClosableObjects()
                    .detectActivityLeaks()
                    .penaltyLog()
                    .build()
            )
        }
    }
}
```

**Performance Budget:**

| Metric | Target | Tool |
|--------|--------|------|
| Cold Startup | < 2s | Baseline Profiles |
| Frame Duration | < 16ms | FrameMetrics |
| APK Size | < 30MB | R8 + Bundler |
| Memory | < 150MB | Profiler |
| Network Calls | < 500ms | OkHttp Metrics |
| DB Queries | < 50ms | Room Tracing |
| Crash Rate | < 0.1% | Crashlytics |
| ANR Rate | < 0.05% | Play Console |

---

*Document Version: 1.0 | Last Updated: 2026-07-16*
