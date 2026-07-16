# Project Structure

## Table of Contents

1. [Gradle Configuration](#gradle-configuration)
2. [Module Structure](#module-structure)
3. [Feature Modules](#feature-modules)
4. [Domain Layer](#domain-layer)
5. [Data Layer](#data-layer)
6. [Presentation Layer](#presentation-layer)
7. [Core Module](#core-module)
8. [Common Module](#common-module)
9. [DI Modules](#di-modules)
10. [Build Variants](#build-variants)
11. [ProGuard/R8 Rules](#proguardr8-rules)
12. [Kotlin Code Conventions](#kotlin-code-conventions)
13. [Dependency Versions](#dependency-versions)
14. [Convention Plugins](#convention-plugins)
15. [Code Generation](#code-generation)

---

## Gradle Configuration

### Root build.gradle.kts

```kotlin
// build.gradle.kts (root)
plugins {
    alias(libs.plugins.android.application) apply false
    alias(libs.plugins.android.library) apply false
    alias(libs.plugins.kotlin.android) apply false
    alias(libs.plugins.kotlin.compose) apply false
    alias(libs.plugins.hilt) apply false
    alias(libs.plugins.ksp) apply false
    alias(libs.plugins.kotlin.serialization) apply false
    alias(libs.plugins.google.services) apply false
}

tasks.register("clean", Delete::class) {
    delete(rootProject.layout.buildDirectory)
}
```

### settings.gradle.kts

```kotlin
// settings.gradle.kts
pluginManagement {
    repositories {
        google {
            content {
                includeGroupByRegex("com\\.android.*")
                includeGroupByRegex("com\\.google.*")
                includeGroupByRegex("androidx.*")
            }
        }
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "nexus-ai-android"

include(":app")
include(":build-logic")

// Feature modules
include(":feature:feature-chat")
include(":feature:feature-dashboard")
include(":feature:feature-agents")
include(":feature:feature-knowledge")
include(":feature:feature-models")
include(":feature:feature-settings")
include(":feature:feature-auth")
include(":feature:feature-notifications")

// Layer modules
include(":domain")
include(":data")

// Core modules
include(":core:core-network")
include(":core:core-database")
include(":core:core-security")
include(":core:core-utils")

// Common module
include(":common")
```

### Version Catalog (libs.versions.toml)

```toml
# gradle/libs.versions.toml
[versions]
agp = "8.7.0"
kotlin = "2.0.21"
ksp = "2.0.21-1.0.28"
hilt = "2.53.1"
room = "2.7.0"
retrofit = "2.11.0"
okhttp = "4.12.0"
compose-bom = "2024.12.01"
navigation = "2.8.5"
hilt-navigation = "1.2.0"
work = "2.10.0"
datastore = "1.1.1"
security-crypto = "1.1.0-alpha06"
biometric = "1.2.0-alpha05"
firebase-bom = "33.7.0"
coil = "2.7.0"
timber = "5.0.1"
junit5 = "5.11.3"
mockk = "1.13.13"
turbine = "1.2.0"
espresso = "3.6.1"
kotlinx-serialization = "1.7.3"
kotlinx-coroutines = "1.9.0"
lifecycle = "2.8.7"
activity-compose = "1.9.3"
material3 = "1.3.1"

[libraries]
# AndroidX Core
androidx-core-ktx = { group = "androidx.core", name = "core-ktx", version.ref = "kotlin" }
androidx-activity-compose = { group = "androidx.activity", name = "activity-compose", version.ref = "activity-compose" }

# Compose
compose-bom = { group = "androidx.compose", name = "compose-bom", version.ref = "compose-bom" }
compose-ui = { group = "androidx.compose.ui", name = "ui" }
compose-ui-graphics = { group = "androidx.compose.ui", name = "ui-graphics" }
compose-ui-tooling = { group = "androidx.compose.ui", name = "ui-tooling" }
compose-ui-tooling-preview = { group = "androidx.compose.ui", name = "ui-tooling-preview" }
compose-material3 = { group = "androidx.compose.material3", name = "material3", version.ref = "material3" }
compose-material-icons = { group = "androidx.compose.material", name = "material-icons-extended" }
compose-runtime = { group = "androidx.compose.runtime", name = "runtime" }

# Lifecycle
lifecycle-runtime-ktx = { group = "androidx.lifecycle", name = "lifecycle-runtime-ktx", version.ref = "lifecycle" }
lifecycle-viewmodel-compose = { group = "androidx.lifecycle", name = "lifecycle-viewmodel-compose", version.ref = "lifecycle" }
lifecycle-runtime-compose = { group = "androidx.lifecycle", name = "lifecycle-runtime-compose", version.ref = "lifecycle" }

# Navigation
navigation-compose = { group = "androidx.navigation", name = "navigation-compose", version.ref = "navigation" }

# Hilt
hilt-android = { group = "com.google.dagger", name = "hilt-android", version.ref = "hilt" }
hilt-compiler = { group = "com.google.dagger", name = "hilt-compiler", version.ref = "hilt" }
hilt-navigation-compose = { group = "androidx.hilt", name = "hilt-navigation-compose", version.ref = "hilt-navigation" }
hilt-work = { group = "androidx.hilt", name = "hilt-work", version.ref = "hilt-navigation" }
hilt-work-compiler = { group = "androidx.hilt", name = "hilt-compiler", version.ref = "hilt-navigation" }

# Room
room-runtime = { group = "androidx.room", name = "room-runtime", version.ref = "room" }
room-ktx = { group = "androidx.room", name = "room-ktx", version.ref = "room" }
room-compiler = { group = "androidx.room", name = "room-compiler", version.ref = "room" }
room-paging = { group = "androidx.room", name = "room-paging", version.ref = "room" }

# Networking
retrofit = { group = "com.squareup.retrofit2", name = "retrofit", version.ref = "retrofit" }
retrofit-kotlinx = { group = "com.squareup.retrofit2", name = "converter-kotlinx-serialization", version.ref = "retrofit" }
okhttp = { group = "com.squareup.okhttp3", name = "okhttp", version.ref = "okhttp" }
okhttp-logging = { group = "com.squareup.okhttp3", name = "logging-interceptor", version.ref = "okhttp" }
okhttp-websocket = { group = "com.squareup.okhttp3", name = "okhttp-ws", version.ref = "okhttp" }

# Serialization
kotlinx-serialization-json = { group = "org.jetbrains.kotlinx", name = "kotlinx-serialization-json", version.ref = "kotlinx-serialization" }

# Coroutines
kotlinx-coroutines-core = { group = "org.jetbrains.kotlinx", name = "kotlinx-coroutines-core", version.ref = "kotlinx-coroutines" }
kotlinx-coroutines-android = { group = "org.jetbrains.kotlinx", name = "kotlinx-coroutines-android", version.ref = "kotlinx-coroutines" }

# DataStore
datastore-preferences = { group = "androidx.datastore", name = "datastore-preferences", version.ref = "datastore" }

# Security
security-crypto = { group = "androidx.security", name = "security-crypto", version.ref = "security-crypto" }

# Biometric
biometric = { group = "androidx.biometric", name = "biometric", version.ref = "biometric" }

# WorkManager
work-runtime-ktx = { group = "androidx.work", name = "work-runtime-ktx", version.ref = "work" }

# Firebase
firebase-bom = { group = "com.google.firebase", name = "firebase-bom", version.ref = "firebase-bom" }
firebase-analytics = { group = "com.google.firebase", name = "firebase-analytics-ktx" }
firebase-crashlytics = { group = "com.google.firebase", name = "firebase-crashlytics-ktx" }
firebase-messaging = { group = "com.google.firebase", name = "firebase-messaging-ktx" }

# Image Loading
coil-compose = { group = "io.coil-kt", name = "coil-compose", version.ref = "coil" }

# Logging
timber = { group = "com.jakewharton.timber", name = "timber", version.ref = "timber" }

# Testing
junit5-api = { group = "org.junit.jupiter", name = "junit-jupiter-api", version.ref = "junit5" }
junit5-engine = { group = "org.junit.jupiter", name = "junit-jupiter-engine", version.ref = "junit5" }
mockk = { group = "io.mockk", name = "mockk", version.ref = "mockk" }
mockk-android = { group = "io.mockk", name = "mockk-android", version.ref = "mockk" }
turbine = { group = "app.cash.turbine", name = "turbine", version.ref = "turbine" }
coroutines-test = { group = "org.jetbrains.kotlinx", name = "kotlinx-coroutines-test", version.ref = "kotlinx-coroutines" }
compose-ui-test = { group = "androidx.compose.ui", name = "ui-test-junit4" }
compose-ui-test-manifest = { group = "androidx.compose.ui", name = "ui-test-manifest" }
espresso-core = { group = "androidx.test.espresso", name = "espresso-core", version.ref = "espresso" }
espresso-contrib = { group = "androidx.test.espresso", name = "espresso-contrib", version.ref = "espresso" }
androidx-test-runner = { group = "androidx.test", name = "runner", version = "1.6.2" }
androidx-test-rules = { group = "androidx.test", name = "rules", version = "1.6.1" }

[plugins]
android-application = { id = "com.android.application", version.ref = "agp" }
android-library = { id = "com.android.library", version.ref = "agp" }
kotlin-android = { id = "org.jetbrains.kotlin.android", version.ref = "kotlin" }
kotlin-compose = { id = "org.jetbrains.kotlin.plugin.compose", version.ref = "kotlin" }
kotlin-serialization = { id = "org.jetbrains.kotlin.plugin.serialization", version.ref = "kotlin" }
hilt = { id = "com.google.dagger.hilt.android", version.ref = "hilt" }
ksp = { id = "com.google.devtools.ksp", version.ref = "ksp" }
google-services = { id = "com.google.gms.google-services", version = "4.4.2" }
```

---

## Full Directory Tree

```
nexus-ai-android/
├── .github/
│   └── workflows/
│       ├── android.yml
│       ├── release.yml
│       └── pr-check.yml
├── .gitignore
├── .editorconfig
├── build.gradle.kts
├── settings.gradle.kts
├── gradle.properties
├── gradle/
│   ├── wrapper/
│   │   ├── gradle-wrapper.jar
│   │   └── gradle-wrapper.properties
│   └── libs.versions.toml
├── gradlew
├── gradlew.bat
├── build-logic/
│   ├── build.gradle.kts
│   └── src/main/kotlin/
│       ├── AndroidFeatureConventionPlugin.kt
│       ├── AndroidLibraryConventionPlugin.kt
│       ├── AndroidDomainConventionPlugin.kt
│       ├── AndroidDataConventionPlugin.kt
│       ├── AndroidHiltConventionPlugin.kt
│       └── AndroidRoomConventionPlugin.kt
├── app/
│   ├── build.gradle.kts
│   ├── proguard-rules.pro
│   ├── google-services.json
│   └── src/
│       ├── main/
│       │   ├── AndroidManifest.xml
│       │   ├── java/com/nexus/ai/
│       │   │   ├── NexusApp.kt
│       │   │   ├── MainActivity.kt
│       │   │   ├── di/
│       │   │   │   ├── AppModule.kt
│       │   │   │   └── WorkManagerModule.kt
│       │   │   ├── navigation/
│       │   │   │   ├── NexusNavHost.kt
│       │   │   │   ├── Screen.kt
│       │   │   │   └── Routes.kt
│       │   │   └── worker/
│       │   │       ├── SyncWorker.kt
│       │   │       └── NotificationWorker.kt
│       │   └── res/
│       │       ├── values/
│       │       │   ├── strings.xml
│       │       │   ├── colors.xml
│       │       │   └── themes.xml
│       │       ├── drawable/
│       │       ├── mipmap-hdpi/
│       │       ├── mipmap-mdpi/
│       │       ├── mipmap-xhdpi/
│       │       ├── mipmap-xxhdpi/
│       │       ├── mipmap-xxxhdpi/
│       │       └── xml/
│       │           └── network_security_config.xml
│       └── test/
│           └── java/com/nexus/ai/
│               └── NexusAppTest.kt
├── feature/
│   ├── feature-chat/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/feature/chat/
│   │       ├── ui/
│   │       │   ├── ChatScreen.kt
│   │       │   ├── ChatViewModel.kt
│   │       │   ├── ChatUiState.kt
│   │       │   ├── ChatAction.kt
│   │       │   └── components/
│   │       │       ├── MessageList.kt
│   │       │       ├── MessageBubble.kt
│   │       │       ├── PromptBox.kt
│   │       │       ├── FileUploader.kt
│   │       │       ├── VoiceInput.kt
│   │       │       ├── AgentSelector.kt
│   │       │       ├── ModelIndicator.kt
│   │       │       ├── TokenStreamViewer.kt
│   │       │       ├── ThinkingIndicator.kt
│   │       │       └── ToolExecutionDisplay.kt
│   │       └── di/
│   │           └── ChatModule.kt
│   ├── feature-dashboard/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/feature/dashboard/
│   │       ├── ui/
│   │       │   ├── DashboardScreen.kt
│   │       │   ├── DashboardViewModel.kt
│   │       │   └── components/
│   │       └── di/
│   ├── feature-agents/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/feature/agents/
│   │       ├── ui/
│   │       │   ├── AgentsScreen.kt
│   │       │   ├── AgentsViewModel.kt
│   │       │   ├── AgentDetailScreen.kt
│   │       │   └── components/
│   │       └── di/
│   ├── feature-knowledge/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/feature/knowledge/
│   │       ├── ui/
│   │       │   ├── KnowledgeScreen.kt
│   │       │   ├── KnowledgeViewModel.kt
│   │       │   └── components/
│   │       └── di/
│   ├── feature-models/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/feature/models/
│   │       ├── ui/
│   │       │   ├── ModelsScreen.kt
│   │       │   ├── ModelsViewModel.kt
│   │       │   └── components/
│   │       └── di/
│   ├── feature-settings/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/feature/settings/
│   │       ├── ui/
│   │       │   ├── SettingsScreen.kt
│   │       │   ├── SettingsViewModel.kt
│   │       │   ├── ProfileScreen.kt
│   │       │   ├── BiometricScreen.kt
│   │       │   └── components/
│   │       └── di/
│   ├── feature-auth/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/feature/auth/
│   │       ├── ui/
│   │       │   ├── LoginScreen.kt
│   │       │   ├── LoginViewModel.kt
│   │       │   ├── LoginUiState.kt
│   │       │   ├── RegisterScreen.kt
│   │       │   ├── RegisterViewModel.kt
│   │       │   ├── BiometricLoginScreen.kt
│   │       │   └── components/
│   │       └── di/
│   └── feature-notifications/
│       ├── build.gradle.kts
│       └── src/main/java/com/nexus/ai/feature/notifications/
│           ├── ui/
│           │   ├── NotificationsScreen.kt
│           │   ├── NotificationsViewModel.kt
│           │   └── components/
│           └── di/
├── domain/
│   ├── build.gradle.kts
│   └── src/main/java/com/nexus/ai/domain/
│       ├── entity/
│       │   ├── Message.kt
│       │   ├── Conversation.kt
│       │   ├── Agent.kt
│       │   ├── User.kt
│       │   ├── Model.kt
│       │   ├── Knowledge.kt
│       │   ├── TokenUsage.kt
│       │   └── Error.kt
│       ├── repository/
│       │   ├── AuthRepository.kt
│       │   ├── MessageRepository.kt
│       │   ├── AgentRepository.kt
│       │   ├── UserRepository.kt
│       │   ├── ModelRepository.kt
│       │   ├── KnowledgeRepository.kt
│       │   ├── ConversationRepository.kt
│       │   └── SyncRepository.kt
│       ├── usecase/
│       │   ├── auth/
│       │   │   ├── LoginUseCase.kt
│       │   │   ├── LogoutUseCase.kt
│       │   │   ├── RefreshTokenUseCase.kt
│       │   │   └── BiometricLoginUseCase.kt
│       │   ├── chat/
│       │   │   ├── SendMessageUseCase.kt
│       │   │   ├── ObserveMessagesUseCase.kt
│       │   │   ├── LoadConversationHistoryUseCase.kt
│       │   │   └── CreateConversationUseCase.kt
│       │   ├── agent/
│       │   │   ├── GetAgentsUseCase.kt
│       │   │   ├── GetAgentUseCase.kt
│       │   │   └── SelectAgentUseCase.kt
│       │   ├── model/
│       │   │   ├── GetModelsUseCase.kt
│       │   │   └── SelectModelUseCase.kt
│       │   └── knowledge/
│       │       ├── GetKnowledgeUseCase.kt
│       │       └── SearchKnowledgeUseCase.kt
│       ├── mapper/
│       │   ├── MessageMapper.kt
│       │   ├── AgentMapper.kt
│       │   └── UserMapper.kt
│       └── di/
│           └── DomainModule.kt
├── data/
│   ├── build.gradle.kts
│   └── src/main/java/com/nexus/ai/data/
│       ├── local/
│       │   ├── NexusDatabase.kt
│       │   ├── dao/
│       │   │   ├── MessageDao.kt
│       │   │   ├── ConversationDao.kt
│       │   │   ├── AgentDao.kt
│       │   │   ├── KnowledgeDao.kt
│       │   │   └── UserDao.kt
│       │   ├── entity/
│       │   │   ├── MessageEntity.kt
│       │   │   ├── ConversationEntity.kt
│       │   │   ├── AgentEntity.kt
│       │   │   ├── KnowledgeEntity.kt
│       │   │   └── UserEntity.kt
│       │   └── converter/
│       │       ├── DateConverter.kt
│       │       └── ListConverter.kt
│       ├── remote/
│       │   ├── ApiService.kt
│       │   ├── dto/
│       │   │   ├── LoginRequest.kt
│       │   │   ├── LoginResponse.kt
│       │   │   ├── MessageDto.kt
│       │   │   ├── AgentDto.kt
│       │   │   ├── ModelDto.kt
│       │   │   └── ApiResponse.kt
│       │   ├── interceptor/
│       │   │   ├── AuthInterceptor.kt
│       │   │   ├── TokenRefreshInterceptor.kt
│       │   │   └── LoggingInterceptor.kt
│       │   └── websocket/
│       │       ├── WebSocketManager.kt
│       │       ├── WebSocketEvent.kt
│       │       └── WebSocketProtocol.kt
│       ├── repository/
│       │   ├── AuthRepositoryImpl.kt
│       │   ├── MessageRepositoryImpl.kt
│       │   ├── AgentRepositoryImpl.kt
│       │   ├── UserRepositoryImpl.kt
│       │   ├── ModelRepositoryImpl.kt
│       │   ├── KnowledgeRepositoryImpl.kt
│       │   ├── ConversationRepositoryImpl.kt
│       │   └── SyncRepositoryImpl.kt
│       ├── mapper/
│       │   ├── MessageEntityMapper.kt
│       │   ├── AgentEntityMapper.kt
│       │   └── UserEntityMapper.kt
│       └── di/
│           ├── NetworkModule.kt
│           ├── DatabaseModule.kt
│           ├── RepositoryModule.kt
│           └── DataStoreModule.kt
├── core/
│   ├── core-network/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/core/network/
│   │       ├── NetworkMonitor.kt
│   │       ├── ConnectivityObserver.kt
│   │       ├── RetryPolicy.kt
│   │       └── di/
│   │           └── NetworkCoreModule.kt
│   ├── core-database/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/core/database/
│   │       ├── DatabaseProvider.kt
│   │       ├── MigrationManager.kt
│   │       └── di/
│   │           └── DatabaseCoreModule.kt
│   ├── core-security/
│   │   ├── build.gradle.kts
│   │   └── src/main/java/com/nexus/ai/core/security/
│   │       ├── AndroidKeyStore.kt
│   │       ├── TokenManager.kt
│   │       ├── BiometricHelper.kt
│   │       ├── SecurePrefsManager.kt
│   │       └── di/
│   │           └── SecurityCoreModule.kt
│   └── core-utils/
│       ├── build.gradle.kts
│       └── src/main/java/com/nexus/ai/core/utils/
│           ├── DateFormatter.kt
│           ├── KeyboardUtils.kt
│           ├── NetworkUtils.kt
│           ├── ResourceUtils.kt
│           └── di/
│               └── UtilsCoreModule.kt
├── common/
│   ├── build.gradle.kts
│   └── src/main/java/com/nexus/ai/common/
│       ├── extension/
│       │   ├── StringExt.kt
│       │   ├── DateTimeExt.kt
│       │   ├── CollectionExt.kt
│       │   ├── FlowExt.kt
│       │   ├── ContextExt.kt
│       │   └── CoroutineExt.kt
│       ├── constant/
│       │   ├── ApiConstants.kt
│       │   ├── AppConstants.kt
│       │   ├── NotificationConstants.kt
│       │   └── SecurityConstants.kt
│       ├── type/
│       │   ├── Result.kt
│       │   ├── Either.kt
│       │   ├── AuthState.kt
│       │   ├── ConnectionState.kt
│       │   └── UserRole.kt
│       └── theme/
│           ├── Theme.kt
│           ├── Color.kt
│           ├── Type.kt
│           └── Shape.kt
└── docs/
    └── android/
        ├── 00-architecture-overview.md
        ├── 01-project-structure.md
        ├── 02-networking.md
        ├── 03-authentication.md
        └── 04-ai-chat.md
```

---

## Feature Modules

Each feature module follows this structure:

```kotlin
// feature-chat/build.gradle.kts
plugins {
    alias(libs.plugins.android.library)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.hilt)
    alias(libs.plugins.ksp)
}

android {
    namespace = "com.nexus.ai.feature.chat"
    defaultConfig {
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }
}

dependencies {
    implementation(project(":domain"))
    implementation(project(":common"))
    implementation(project(":core:core-network"))

    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.ui.graphics)
    implementation(libs.compose.material3)
    implementation(libs.compose.material.icons)
    implementation(libs.compose.ui.tooling.preview)

    implementation(libs.lifecycle.runtime.compose)
    implementation(libs.lifecycle.viewmodel.compose)
    implementation(libs.navigation.compose)

    implementation(libs.hilt.android)
    ksp(libs.hilt.compiler)
    implementation(libs.hilt.navigation.compose)

    implementation(libs.coil.compose)
    implementation(libs.timber)

    debugImplementation(libs.compose.ui.tooling)
    debugImplementation(libs.compose.ui.test.manifest)

    testImplementation(libs.junit5.api)
    testImplementation(libs.mockk)
    testImplementation(libs.coroutines.test)
    testImplementation(libs.turbine)
}
```

---

## Domain Layer

```kotlin
// domain/entity/Message.kt
data class Message(
    val id: String,
    val conversationId: String,
    val role: MessageRole,
    val content: String,
    val timestamp: Long,
    val toolCalls: List<ToolCall>? = null,
    val toolResults: List<ToolResult>? = null,
    val tokenUsage: TokenUsage? = null,
    val modelId: String? = null,
    val isStreaming: Boolean = false
)

enum class MessageRole {
    USER, ASSISTANT, SYSTEM, TOOL
}

// domain/repository/MessageRepository.kt
interface MessageRepository {
    fun observeMessages(conversationId: String): Flow<List<Message>>
    suspend fun getMessages(conversationId: String, limit: Int, offset: Int): List<Message>
    suspend fun sendMessage(text: String, conversationId: String): Message
    suspend fun deleteMessage(messageId: String)
    suspend fun getMessageById(messageId: String): Message?
    suspend fun updateStreamingMessage(messageId: String, content: String)
}

// domain/usecase/chat/SendMessageUseCase.kt
class SendMessageUseCase @Inject constructor(
    private val messageRepository: MessageRepository,
    private val webSocketManager: WebSocketManager
) {
    operator fun invoke(
        text: String,
        conversationId: String,
        agentId: String? = null
    ): Flow<Result<Message>> = flow {
        emit(Result.Loading)
        try {
            val message = messageRepository.sendMessage(text, conversationId)
            emit(Result.Success(message))
            webSocketManager.sendMessage(
                text = text,
                conversationId = conversationId,
                agentId = agentId
            )
        } catch (e: Exception) {
            emit(Result.Error(e))
        }
    }
}
```

---

## Data Layer

```kotlin
// data/local/NexusDatabase.kt
@Database(
    entities = [
        MessageEntity::class,
        ConversationEntity::class,
        AgentEntity::class,
        KnowledgeEntity::class,
        UserDao::class
    ],
    version = 3,
    exportSchema = true
)
@TypeConverters(DateConverter::class, ListConverter::class)
abstract class NexusDatabase : RoomDatabase() {
    abstract fun messageDao(): MessageDao
    abstract fun conversationDao(): ConversationDao
    abstract fun agentDao(): AgentDao
    abstract fun knowledgeDao(): KnowledgeDao
    abstract fun userDao(): UserDao
}

// data/local/entity/MessageEntity.kt
@Entity(tableName = "messages")
data class MessageEntity(
    @PrimaryKey val id: String,
    @ColumnInfo(name = "conversation_id") val conversationId: String,
    @ColumnInfo(name = "role") val role: String,
    @ColumnInfo(name = "content") val content: String,
    @ColumnInfo(name = "timestamp") val timestamp: Long,
    @ColumnInfo(name = "tool_calls") val toolCalls: String?,
    @ColumnInfo(name = "tool_results") val toolResults: String?,
    @ColumnInfo(name = "model_id") val modelId: String?,
    @ColumnInfo(name = "is_streaming") val isStreaming: Boolean = false
)

// data/local/dao/MessageDao.kt
@Dao
interface MessageDao {
    @Query("SELECT * FROM messages WHERE conversation_id = :conversationId ORDER BY timestamp ASC")
    fun observeMessages(conversationId: String): Flow<List<MessageEntity>>

    @Query("SELECT * FROM messages WHERE conversation_id = :conversationId ORDER BY timestamp ASC LIMIT :limit OFFSET :offset")
    suspend fun getMessages(conversationId: String, limit: Int, offset: Int): List<MessageEntity>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertMessage(message: MessageEntity)

    @Query("UPDATE messages SET content = :content WHERE id = :messageId")
    suspend fun updateContent(messageId: String, content: String)

    @Query("DELETE FROM messages WHERE id = :messageId")
    suspend fun deleteMessage(messageId: String)

    @Query("SELECT * FROM messages WHERE id = :messageId")
    suspend fun getMessageById(messageId: String): MessageEntity?

    @Query("DELETE FROM messages WHERE conversation_id = :conversationId")
    suspend fun deleteMessagesByConversation(conversationId: String)
}

// data/remote/ApiService.kt
interface ApiService {
    @POST("api/v1/auth/login")
    suspend fun login(@Body request: LoginRequest): LoginResponse

    @POST("api/v1/auth/refresh")
    suspend fun refreshToken(@Body request: RefreshTokenRequest): RefreshTokenResponse

    @GET("api/v1/agents")
    suspend fun getAgents(): ApiResponse<List<AgentDto>>

    @GET("api/v1/models")
    suspend fun getModels(): ApiResponse<List<ModelDto>>

    @GET("api/v1/conversations")
    suspend fun getConversations(): ApiResponse<List<ConversationDto>>

    @GET("api/v1/conversations/{id}/messages")
    suspend fun getMessages(
        @Path("id") conversationId: String,
        @Query("limit") limit: Int = 50,
        @Query("offset") offset: Int = 0
    ): ApiResponse<List<MessageDto>>

    @POST("api/v1/conversations")
    suspend fun createConversation(@Body request: CreateConversationRequest): ApiResponse<ConversationDto>

    @DELETE("api/v1/conversations/{id}")
    suspend fun deleteConversation(@Path("id") conversationId: String): ApiResponse<Unit>
}

// data/repository/MessageRepositoryImpl.kt
class MessageRepositoryImpl @Inject constructor(
    private val messageDao: MessageDao,
    private val apiService: ApiService,
    private val messageMapper: MessageEntityMapper
) : MessageRepository {

    override fun observeMessages(conversationId: String): Flow<List<Message>> {
        return messageDao.observeMessages(conversationId).map { entities ->
            entities.map { messageMapper.toDomain(it) }
        }
    }

    override suspend fun sendMessage(text: String, conversationId: String): Message {
        val localMessage = MessageEntity(
            id = UUID.randomUUID().toString(),
            conversationId = conversationId,
            role = MessageRole.USER.name.lowercase(),
            content = text,
            timestamp = System.currentTimeMillis()
        )
        messageDao.insertMessage(localMessage)
        return messageMapper.toDomain(localMessage)
    }

    override suspend fun updateStreamingMessage(messageId: String, content: String) {
        messageDao.updateContent(messageId, content)
    }
}
```

---

## Presentation Layer

```kotlin
// feature/chat/ui/ChatScreen.kt
@Composable
fun ChatScreen(
    viewModel: ChatViewModel = hiltViewModel(),
    onNavigateToAgents: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.events.collect { event ->
            when (event) {
                is ChatEvent.ShowError -> { /* Show snackbar */ }
                is ChatEvent.NavigateToLogin -> { /* Navigate */ }
            }
        }
    }

    Column(modifier = Modifier.fillMaxSize()) {
        // Top bar
        ChatTopBar(
            agent = uiState.selectedAgent,
            isConnected = uiState.isWebSocketConnected,
            onAgentClick = onNavigateToAgents
        )

        // Message list
        MessageList(
            messages = uiState.messages,
            streamingText = uiState.streamingText,
            isStreaming = uiState.isStreaming,
            modifier = Modifier.weight(1f)
        )

        // Input bar
        PromptBox(
            isLoading = uiState.isLoading,
            onSendMessage = { viewModel.onAction(ChatAction.SendMessage(it)) },
            onUploadFile = { viewModel.onAction(ChatAction.UploadFile(it)) }
        )
    }
}

// feature/chat/ui/ChatViewModel.kt
@HiltViewModel
class ChatViewModel @Inject constructor(
    private val sendMessageUseCase: SendMessageUseCase,
    private val observeMessagesUseCase: ObserveMessagesUseCase,
    private val webSocketManager: WebSocketManager,
    private val analyticsManager: AnalyticsManager
) : ViewModel() {

    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    private val _events = Channel<ChatEvent>(Channel.BUFFERED)
    val events: Flow<ChatEvent> = _events.receiveAsFlow()

    init {
        observeMessages()
        connectWebSocket()
    }

    private fun observeMessages() {
        observeMessagesUseCase()
            .onEach { messages ->
                _uiState.update { it.copy(messages = messages) }
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
            _uiState.update { it.copy(isLoading = true, isStreaming = false, streamingText = "") }
            analyticsManager.logChatMessage(
                agentId = _uiState.value.selectedAgent?.id ?: "default",
                modelId = _uiState.value.selectedAgent?.modelId ?: "default",
                messageLength = text.length
            )
            sendMessageUseCase(text, getCurrentConversationId())
                .onEach { result ->
                    when (result) {
                        is Result.Loading -> _uiState.update { it.copy(isLoading = true) }
                        is Result.Success -> _uiState.update { it.copy(isLoading = false) }
                        is Result.Error -> {
                            _events.send(ChatEvent.ShowError(result.exception.message))
                            _uiState.update { it.copy(isLoading = false) }
                        }
                    }
                }
                .launchIn(viewModelScope)
        }
    }
}
```

---

## DI Modules

```kotlin
// data/di/NetworkModule.kt
@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    @Provides
    @Singleton
    fun provideJson(): Json = Json {
        ignoreUnknownKeys = true
        coerceInputValues = true
        isLenient = true
    }

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
    fun provideRetrofit(
        okHttpClient: OkHttpClient,
        json: Json
    ): Retrofit {
        return Retrofit.Builder()
            .baseUrl(BuildConfig.BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
            .build()
    }

    @Provides
    @Singleton
    fun provideApiService(retrofit: Retrofit): ApiService {
        return retrofit.create(ApiService::class.java)
    }

    @Provides
    @Singleton
    fun provideWebSocketManager(
        okHttpClient: OkHttpClient,
        tokenManager: TokenManager,
        json: Json
    ): WebSocketManager {
        return WebSocketManager(okHttpClient, tokenManager, json)
    }
}

// data/di/DatabaseModule.kt
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

    @Provides fun provideMessageDao(db: NexusDatabase): MessageDao = db.messageDao()
    @Provides fun provideConversationDao(db: NexusDatabase): ConversationDao = db.conversationDao()
    @Provides fun provideAgentDao(db: NexusDatabase): AgentDao = db.agentDao()
    @Provides fun provideKnowledgeDao(db: NexusDatabase): KnowledgeDao = db.knowledgeDao()
}

// data/di/RepositoryModule.kt
@Module
@InstallIn(SingletonComponent::class)
abstract class RepositoryModule {

    @Binds
    @Singleton
    abstract fun bindMessageRepository(impl: MessageRepositoryImpl): MessageRepository

    @Binds
    @Singleton
    abstract fun bindAuthRepository(impl: AuthRepositoryImpl): AuthRepository

    @Binds
    @Singleton
    abstract fun bindAgentRepository(impl: AgentRepositoryImpl): AgentRepository

    @Binds
    @Singleton
    abstract fun bindConversationRepository(impl: ConversationRepositoryImpl): ConversationRepository
}
```

---

## Build Variants

```kotlin
// app/build.gradle.kts
android {
    namespace = "com.nexus.ai"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.nexus.ai"
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "1.0.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        vectorDrawables {
            useSupportLibrary = true
        }

        buildConfigField("String", "BASE_URL", "\"https://api.nexus-ai.com/\"")
        buildConfigField("String", "WS_URL", "\"wss://ws.nexus-ai.com/\"")
        buildConfigField("String", "FIREBASE_PROJECT_ID", "\"nexus-ai-prod\"")
    }

    buildTypes {
        debug {
            isMinifyEnabled = false
            isDebuggable = true
            applicationIdSuffix = ".debug"
            versionNameSuffix = "-debug"
            buildConfigField("String", "BASE_URL", "\"https://staging-api.nexus-ai.com/\"")
        }

        release {
            isMinifyEnabled = true
            isShrinkResources = true
            isDebuggable = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            signingConfig = signingConfigs.getByName("release")
        }

        create("staging") {
            initWith(buildTypes.getByName("debug"))
            applicationIdSuffix = ".staging"
            versionNameSuffix = "-staging"
            buildConfigField("String", "BASE_URL", "\"https://staging-api.nexus-ai.com/\"")
        }
    }

    flavorDimensions += "environment"
    productFlavors {
        create("dev") {
            dimension = "environment"
            applicationIdSuffix = ".dev"
            versionNameSuffix = "-dev"
        }
        create("prod") {
            dimension = "environment"
            applicationIdSuffix = ""
            versionNameSuffix = ""
        }
    }
}
```

---

## ProGuard/R8 Rules

```proguard
# proguard-rules.pro

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
-keepattributes Signature
-keepattributes Exceptions
-dontwarn retrofit2.**
-keep class retrofit2.** { *; }
-keepclassmembers,allowshrinking,allowobfuscation interface * {
    @retrofit2.http.* <methods>;
}

# OkHttp
-dontwarn okhttp3.**
-dontwarn okio.**
-keep class okhttp3.** { *; }

# Hilt
-keep class dagger.hilt.** { *; }
-keep class javax.inject.** { *; }
-keep @dagger.hilt.android.lifecycle.HiltViewModel class * { *; }

# Coil
-keep class coil.** { *; }
-dontwarn coil.**

# Firebase
-keep class com.google.firebase.** { *; }
-dontwarn com.google.firebase.**

# Custom models
-keep class com.nexus.ai.data.remote.dto.** { *; }
-keep class com.nexus.ai.domain.entity.** { *; }
```

---

## Kotlin Code Conventions

### Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Classes | PascalCase | `ChatViewModel` |
| Functions | camelCase | `sendMessage()` |
| Variables | camelCase | `uiState` |
| Constants | SCREAMING_SNAKE | `MAX_RETRY_COUNT` |
| Packages | lowercase | `com.nexus.ai` |
| Composable | PascalCase | `MessageBubble()` |
| Module | kebab-case | `:feature-chat` |
| Database table | snake_case | `chat_messages` |
| Column | snake_case | `conversation_id` |

### File Organization

```
// 1. Package declaration
package com.nexus.ai.feature.chat.ui

// 2. Imports (grouped, no unused)
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier

import com.nexus.ai.domain.entity.Message

// 3. Composable functions
@Composable
fun ChatScreen(...) { ... }

// 4. Private composables
@Composable
private fun ChatTopBar(...) { ... }

// 5. Preview functions
@Preview
@Composable
private fun ChatScreenPreview() { ... }
```

### Style Rules

```kotlin
// 1. Use data classes for state
data class ChatUiState(
    val messages: List<Message> = emptyList(),
    val isLoading: Boolean = false
)

// 2. Use sealed classes for events
sealed class ChatAction {
    data class SendMessage(val text: String) : ChatAction()
    data object CancelGeneration : ChatAction()
}

// 3. Use extension functions
fun String.isValidEmail(): Boolean =
    Patterns.EMAIL_ADDRESS.matcher(this).matches()

// 4. Use scope functions appropriately
viewModel.update { it.copy(isLoading = true) }

// 5. Use named parameters for clarity
RoundedCornerShape(
    topStart = 16.dp,
    topEnd = 16.dp,
    bottomStart = 0.dp,
    bottomEnd = 0.dp
)

// 6. Use collectAsStateWithLifecycle()
val uiState by viewModel.uiState.collectAsStateWithLifecycle()
```

---

## Convention Plugins

```kotlin
// build-logic/src/main/kotlin/AndroidFeatureConventionPlugin.kt
class AndroidFeatureConventionPlugin : Plugin<Project> {
    override fun apply(target: Project) {
        with(target) {
            pluginManager.apply("nexus.android.library")
            pluginManager.apply("nexus.android.hilt")
            pluginManager.apply("nexus.android.compose")

            dependencies {
                implementation(project(":domain"))
                implementation(project(":common"))
                implementation(project(":core:core-network"))
                implementation(project(":core:core-security"))
                implementation(project(":core:core-utils"))

                implementation(libs.lifecycle.runtime.compose)
                implementation(libs.lifecycle.viewmodel.compose)
                implementation(libs.navigation.compose)
                implementation(libs.hilt.navigation.compose)
            }
        }
    }
}

// build-logic/build.gradle.kts
plugins {
    `kotlin-dsl`
}

dependencies {
    compileOnly(libs.android.gradle.plugin)
    compileOnly(libs.kotlin.gradle.plugin)
    compileOnly(libs.hilt.gradle.plugin)
}

gradlePlugin {
    plugins {
        register("androidLibrary") {
            id = "nexus.android.library"
            implementationClass = "AndroidLibraryConventionPlugin"
        }
        register("androidFeature") {
            id = "nexus.android.feature"
            implementationClass = "AndroidFeatureConventionPlugin"
        }
        register("androidDomain") {
            id = "nexus.android.domain"
            implementationClass = "AndroidDomainConventionPlugin"
        }
        register("androidData") {
            id = "nexus.android.data"
            implementationClass = "AndroidDataConventionPlugin"
        }
        register("androidHilt") {
            id = "nexus.android.hilt"
            implementationClass = "AndroidHiltConventionPlugin"
        }
        register("androidRoom") {
            id = "nexus.android.room"
            implementationClass = "AndroidRoomConventionPlugin"
        }
        register("androidCompose") {
            id = "nexus.android.compose"
            implementationClass = "AndroidComposeConventionPlugin"
        }
    }
}
```

---

## Code Generation

### Hilt Generated Files

```
app/build/generated/hilt/
├── sources/
│   └── com/nexus/ai/
│       ├── Hilt_NexusApp.java
│       ├── Hilt_MainActivity.java
│       ├── NexusApp_MembersInjector.java
│       └── MainActivity_GeneratedInjector.java
└── classes/
    └── kotlin/
        └── com/nexus/ai/
            ├── Hilt_MainActivity.java
            └── NexusApp_HiltComponents.java
```

### Room Generated Files

```
data/build/generated/ksp/
├── androidUnitTest/
│   └── com/nexus/ai/data/
│       └── local/
│           ├── NexusDatabase_Impl.java
│           ├── MessageDao_Impl.java
│           ├── ConversationDao_Impl.java
│           └── AgentDao_Impl.java
└── androidTest/
    └── com/nexus/ai/data/
        └── local/
            ├── MessageDao_Impl.java
            └── AgentDao_Impl.java
```

### KSP Configuration

```kotlin
// domain/build.gradle.kts
plugins {
    alias(libs.plugins.android.library)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.ksp)
}

dependencies {
    ksp(libs.hilt.compiler)
}

// data/build.gradle.kts
plugins {
    alias(libs.plugins.android.library)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.ksp)
    alias(libs.plugins.kotlin.serialization)
}

dependencies {
    ksp(libs.hilt.compiler)
    ksp(libs.room.compiler)
}
```

---

*Document Version: 1.0 | Last Updated: 2026-07-16*
