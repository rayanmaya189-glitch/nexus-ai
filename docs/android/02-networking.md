# Networking

## Table of Contents

1. [OkHttp Configuration](#okhttp-configuration)
2. [Retrofit Configuration](#retrofit-configuration)
3. [Authentication Interceptor](#authentication-interceptor)
4. [Token Refresh Interceptor](#token-refresh-interceptor)
5. [Logging Interceptor](#logging-interceptor)
6. [Certificate Pinning](#certificate-pinning)
7. [Network Security Config](#network-security-config)
8. [WebSocket Client](#websocket-client)
9. [WebSocket Message Protocol](#websocket-message-protocol)
10. [WebSocket Authentication](#websocket-authentication)
11. [WebSocket Reconnection](#websocket-reconnection)
12. [API Endpoint Definitions](#api-endpoint-definitions)
13. [Request/Response DTOs](#requestresponse-dtos)
14. [Error Handling](#error-handling)
15. [Offline Detection](#offline-detection)
16. [Retry Logic](#retry-logic)
17. [Request Cancellation](#request-cancellation)
18. [Mock Data for Testing](#mock-data-for-testing)
19. [Network Monitoring](#network-monitoring)

---

## OkHttp Configuration

```kotlin
// core-network/src/main/java/com/nexus/ai/core/network/OkHttpModule.kt
@Module
@InstallIn(SingletonComponent::class)
object OkHttpModule {

    @Provides
    @Singleton
    fun provideConnectionPool(): ConnectionPool {
        return ConnectionPool(
            maxIdleConnections = 5,
            keepAliveDuration = 5,
            timeUnit = TimeUnit.MINUTES
        )
    }

    @Provides
    @Singleton
    fun provideDispatcher(): Dispatcher {
        return Dispatcher().apply {
            maxRequests = 64
            maxRequestsPerHost = 10
        }
    }

    @Provides
    @Singleton
    fun provideOkHttpClient(
        authInterceptor: AuthInterceptor,
        tokenRefreshInterceptor: TokenRefreshInterceptor,
        loggingInterceptor: HttpLoggingInterceptor,
        connectionPool: ConnectionPool,
        dispatcher: Dispatcher,
        @ApplicationContext context: Context
    ): OkHttpClient {
        return OkHttpClient.Builder()
            // Connection settings
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(60, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .callTimeout(120, TimeUnit.SECONDS)
            .connectionPool(connectionPool)
            .dispatcher(dispatcher)

            // Interceptors (order matters)
            .addInterceptor(authInterceptor)
            .addInterceptor(tokenRefreshInterceptor)
            .addNetworkInterceptor(OkHttpNetworkInterceptor())

            // Logging (debug only)
            .apply {
                if (BuildConfig.DEBUG) {
                    addInterceptor(loggingInterceptor)
                }
            }

            // Security
            .sslSocketFactory(
                SSLContext.getInstance("TLS").apply {
                    init(null, null, null)
                }.socketFactory,
                TrustAllCertsManager()
            )
            .hostnameVerifier { _, _ -> true }

            // Retry
            .retryOnConnectionFailure(true)

            .build()
    }
}
```

**OkHttp Configuration Table:**

| Setting | Value | Rationale |
|---------|-------|-----------|
| `connectTimeout` | 30s | Balance between patience and responsiveness |
| `readTimeout` | 60s | Allows for long AI responses |
| `writeTimeout` | 30s | File uploads may take time |
| `callTimeout` | 120s | Overall timeout for entire operation |
| `maxIdleConnections` | 5 | Reuse connections efficiently |
| `keepAliveDuration` | 5min | Keep connections warm |
| `maxRequests` | 64 | Handle concurrent requests |
| `maxRequestsPerHost` | 10 | Prevent server overload |

**Connection Flow:**

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Request    │───►│   OkHttp     │───►│   Server     │
│              │    │   Client     │    │              │
└──────────────┘    └──────┬───────┘    └──────┬───────┘
                           │                    │
                    ┌──────┴───────┐    ┌──────┴───────┐
                    │  Connection  │    │   Response   │
                    │    Pool      │    │              │
                    └──────────────┘    └──────────────┘
```

---

## Retrofit Configuration

```kotlin
// core-network/src/main/java/com/nexus/ai/core/network/RetrofitModule.kt
@Module
@InstallIn(SingletonComponent::class)
object RetrofitModule {

    @Provides
    @Singleton
    fun provideJson(): Json {
        return Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            isLenient = true
            prettyPrint = BuildConfig.DEBUG
            encodeDefaults = true
        }
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
            .addConverterFactory(
                json.asConverterFactory("application/json".toMediaType())
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

**Retrofit Setup:**

```
┌─────────────────────────────────────────────────────────┐
│                    Retrofit Client                       │
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Base URL: https://api.nexus-ai.com/              │  │
│  │  Converter: Kotlinx.serialization                 │  │
│  │  Call Adapter: Coroutine (default)                │  │
│  └───────────────────────────────────────────────────┘  │
│                         │                               │
│                         ▼                               │
│  ┌───────────────────────────────────────────────────┐  │
│  │  OkHttp Client                                    │  │
│  │  ├── Auth Interceptor (JWT token)                 │  │
│  │  ├── Token Refresh Interceptor                    │  │
│  │  ├── Logging Interceptor (debug)                  │  │
│  │  └── Network Interceptor (cache control)          │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## Authentication Interceptor

```kotlin
// data/remote/interceptor/AuthInterceptor.kt
class AuthInterceptor @Inject constructor(
    private val tokenManager: TokenManager
) : Interceptor {

    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()

        // Skip auth for login/refresh endpoints
        if (originalRequest.url.encodedPath in PUBLIC_ENDPOINTS) {
            return chain.proceed(originalRequest)
        }

        val token = tokenManager.getAccessToken()
            ?: return chain.proceed(originalRequest)

        val authenticatedRequest = originalRequest.newBuilder()
            .header("Authorization", "Bearer $token")
            .header("Content-Type", "application/json")
            .header("Accept", "application/json")
            .header("X-Request-Id", UUID.randomUUID().toString())
            .header("X-Client-Version", BuildConfig.VERSION_NAME)
            .build()

        return chain.proceed(authenticatedRequest)
    }

    companion object {
        val PUBLIC_ENDPOINTS = setOf(
            "/api/v1/auth/login",
            "/api/v1/auth/register",
            "/api/v1/auth/refresh",
            "/api/v1/auth/forgot-password",
            "/api/v1/health"
        )
    }
}
```

**Auth Interceptor Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Request   │───►│  Is Public  │───►│  Add Token  │
│             │    │  Endpoint?  │    │             │
└─────────────┘    └──────┬──────┘    └──────┬──────┘
                          │                   │
                   No     │    Yes            │
                    │     │     │             │
                    ▼     │     ▼             │
             ┌──────┴────┐│   ┌──────────┐    │
             │  Skip Auth ││   │  Proceed │    │
             │  Proceed   ││   │  Without │    │
             └───────────┘│   └──────────┘    │
                          │                   │
                          └───────┬───────────┘
                                  ▼
                           ┌─────────────┐
                           │  Server     │
                           └─────────────┘
```

---

## Token Refresh Interceptor

```kotlin
// data/remote/interceptor/TokenRefreshInterceptor.kt
class TokenRefreshInterceptor @Inject constructor(
    private val tokenManager: TokenManager,
    private val authApi: Provider<AuthApi>,
    private val json: Json
) : Interceptor {

    private val mutex = Mutex()
    private val refreshScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    @Volatile
    private var refreshToken: String? = null

    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()

        // Skip for refresh endpoint to avoid loop
        if (originalRequest.url.encodedPath == "/api/v1/auth/refresh") {
            return chain.proceed(originalRequest)
        }

        val response = chain.proceed(originalRequest)

        if (response.code == 401) {
            response.close()
            return handleUnauthorized(chain)
        }

        return response
    }

    private fun handleUnauthorized(chain: Interceptor.Chain): Response {
        return runBlocking {
            mutex.withLock {
                // Check if token was already refreshed
                val currentToken = tokenManager.getAccessToken()
                val requestToken = chain.request()
                    .header("Authorization")
                    ?.removePrefix("Bearer ")

                if (currentToken != requestToken) {
                    // Token was refreshed, retry with new token
                    val newRequest = chain.request().newBuilder()
                        .header("Authorization", "Bearer $currentToken")
                        .build()
                    return@withLock chain.proceed(newRequest)
                }

                // Need to refresh token
                val refreshToken = tokenManager.getRefreshToken()
                    ?: throw AuthException.NoRefreshToken()

                try {
                    val refreshResponse = authApi.get().refreshToken(
                        RefreshTokenRequest(refreshToken = refreshToken)
                    )

                    if (refreshResponse.isSuccessful) {
                        val body = refreshResponse.body()!!
                        tokenManager.saveTokens(
                            accessToken = body.data.accessToken,
                            refreshToken = body.data.refreshToken
                        )

                        // Retry original request
                        val newRequest = chain.request().newBuilder()
                            .header("Authorization", "Bearer ${body.data.accessToken}")
                            .build()
                        chain.proceed(newRequest)
                    } else {
                        // Refresh failed, clear tokens
                        tokenManager.clearTokens()
                        throw AuthException.TokenExpired()
                    }
                } catch (e: Exception) {
                    tokenManager.clearTokens()
                    throw AuthException.TokenExpired()
                }
            }
        }
    }
}
```

**Token Refresh Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Request   │───►│   Server    │───►│  401 Error  │
└─────────────┘    └─────────────┘    └──────┬──────┘
                                             │
                                             ▼
                                      ┌─────────────┐
                                      │   Mutex     │
                                      │   Lock      │
                                      └──────┬──────┘
                                             │
                                    ┌────────┴────────┐
                                    │                 │
                                    ▼                 ▼
                             ┌─────────────┐  ┌─────────────┐
                             │  Token      │  │  Refresh    │
                             │  Changed?   │  │  Token      │
                             └──────┬──────┘  └──────┬──────┘
                                    │                 │
                           Yes      │    No           │
                            │       │     │           │
                            ▼       │     ▼           │
                     ┌──────────┐   │  ┌──────────┐   │
                     │  Retry   │   │  │  API     │   │
                     │  with    │   │  │  Call    │   │
                     │  new     │   │  └────┬─────┘   │
                     │  token   │   │       │         │
                     └──────────┘   │  ┌────┴─────┐   │
                                    │  │ Success? │   │
                                    │  └────┬─────┘   │
                                    │       │         │
                                    │  Yes  │  No     │
                                    │   │   │   │     │
                                    │   ▼   │   ▼     │
                                    │  ┌────┴──┐┌────┴──┐
                                    │  │ Save  ││ Clear │
                                    │  │ Tokens││ Tokens│
                                    │  └───────┘└───────┘
```

---

## Logging Interceptor

```kotlin
// core-network/src/main/java/com/nexus/ai/core/network/LoggingModule.kt
@Module
@InstallIn(SingletonComponent::class)
object LoggingModule {

    @Provides
    @Singleton
    fun provideLoggingInterceptor(): HttpLoggingInterceptor {
        return HttpLoggingInterceptor { message ->
            Timber.tag("HTTP").d(message)
        }.apply {
            level = if (BuildConfig.DEBUG) {
                HttpLoggingInterceptor.Level.BODY
            } else {
                HttpLoggingInterceptor.Level.NONE
            }
            redactHeader("Authorization")
            redactHeader("Cookie")
        }
    }
}

// Custom network interceptor for detailed logging
class OkHttpNetworkInterceptor : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        val startTime = System.nanoTime()

        Timber.d("HTTP ${request.method} ${request.url}")

        val response = chain.proceed(request)
        val duration = (System.nanoTime() - startTime) / 1_000_000

        Timber.d(
            "HTTP ${response.code} ${request.method} ${request.url} (${duration}ms)"
        )

        return response
    }
}
```

---

## Certificate Pinning

```kotlin
// core-network/src/main/java/com/nexus/ai/core/network/CertificatePinning.kt
object CertificatePinning {

    fun createPinnedOkHttpClient(
        authInterceptor: AuthInterceptor,
        tokenRefreshInterceptor: TokenRefreshInterceptor
    ): OkHttpClient {
        val certificatePinner = CertificatePinner.Builder()
            .add(
                "api.nexus-ai.com",
                "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" // Primary
            )
            .add(
                "api.nexus-ai.com",
                "sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=" // Backup
            )
            .add(
                "ws.nexus-ai.com",
                "sha256/CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=" // WebSocket
            )
            .build()

        return OkHttpClient.Builder()
            .certificatePinner(certificatePinner)
            .addInterceptor(authInterceptor)
            .addInterceptor(tokenRefreshInterceptor)
            .build()
    }
}
```

---

## Network Security Config

```xml
<!-- res/xml/network_security_config.xml -->
<?xml version="1.0" encoding="utf-8"?>
<network-security-config>
    <!-- Debug: Trust user-added CAs -->
    <domain-config cleartextTrafficPermitted="false">
        <domain includeSubdomains="true">api.nexus-ai.com</domain>
        <domain includeSubdomains="true">ws.nexus-ai.com</domain>

        <!-- Production certificates -->
        <trust-anchors>
            <certificates src="system" />
            <certificates src="user" /> <!-- For debugging with proxy -->
        </trust-anchors>
    </domain-config>

    <!-- Certificate pinning -->
    <domain-config>
        <domain includeSubdomains="true">api.nexus-ai.com</domain>
        <pin-set expiration="2027-01-01">
            <pin digest="SHA-256">AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=</pin>
            <pin digest="SHA-256">BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=</pin>
        </pin-set>
    </domain-config>

    <!-- WebSocket pinning -->
    <domain-config>
        <domain includeSubdomains="true">ws.nexus-ai.com</domain>
        <pin-set expiration="2027-01-01">
            <pin digest="SHA-256">CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=</pin>
            <pin digest="SHA-256">DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD=</pin>
        </pin-set>
    </domain-config>
</network-security-config>
```

**Network Security Flow:**

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   App        │───►│   Network    │───►│   Server     │
│   Request    │    │   Security   │    │   HTTPS      │
└──────────────┘    │   Config     │    └──────────────┘
                    └──────┬───────┘
                           │
                    ┌──────┴───────┐
                    │   Check      │
                    │   Domain     │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
       ┌──────────┐ ┌──────────┐ ┌──────────┐
       │  Cleartext│ │  Cert    │ │  Pin     │
       │  Allowed? │ │  Trust   │ │  Check   │
       └──────────┘ └──────────┘ └──────────┘
```

---

## WebSocket Client

```kotlin
// data/remote/websocket/WebSocketManager.kt
@Singleton
class WebSocketManager @Inject constructor(
    private val okHttpClient: OkHttpClient,
    private val tokenManager: TokenManager,
    private val json: Json
) {
    private var webSocket: WebSocket? = null
    private val _events = MutableSharedFlow<WebSocketEvent>(
        replay = 1,
        extraBufferCapacity = 64
    )
    val events: SharedFlow<WebSocketEvent> = _events.asSharedFlow()

    private val _connectionState = MutableStateFlow(ConnectionState.DISCONNECTED)
    val connectionState: StateFlow<ConnectionState> = _connectionState.asStateFlow()

    private var reconnectJob: Job? = null
    private var heartbeatJob: Job? = null
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    private var currentConversationId: String? = null
    private var reconnectAttempts = 0
    private val maxReconnectAttempts = 10
    private val baseReconnectDelay = 1000L
    private val maxReconnectDelay = 30000L

    fun connect(conversationId: String) {
        currentConversationId = conversationId
        val token = tokenManager.getAccessToken() ?: return

        val request = Request.Builder()
            .url("${BuildConfig.WS_URL}/ws?token=$token")
            .build()

        _connectionState.value = ConnectionState.CONNECTING

        webSocket = okHttpClient.newWebSocket(request, createWebSocketListener())
    }

    private fun createWebSocketListener(): WebSocketListener {
        return object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                super.onOpen(webSocket, response)
                Timber.d("WebSocket connected")
                _connectionState.value = ConnectionState.CONNECTED
                reconnectAttempts = 0
                startHeartbeat()
                sendAuthMessage()
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                super.onMessage(webSocket, text)
                handleMessage(text)
            }

            override fun onMessage(webSocket: WebSocket, bytes: ByteString) {
                super.onMessage(webSocket, bytes)
                handleMessage(bytes.utf8())
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                super.onClosing(webSocket, code, reason)
                Timber.d("WebSocket closing: $code $reason")
                webSocket.close(1000, null)
                _connectionState.value = ConnectionState.DISCONNECTED
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                super.onClosed(webSocket, code, reason)
                Timber.d("WebSocket closed: $code $reason")
                _connectionState.value = ConnectionState.DISCONNECTED
                stopHeartbeat()
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                super.onFailure(webSocket, t, response)
                Timber.e(t, "WebSocket failure")
                _connectionState.value = ConnectionState.DISCONNECTED
                _events.tryEmit(WebSocketEvent.Error(t.message ?: "Unknown error"))
                stopHeartbeat()
                scheduleReconnect()
            }
        }
    }

    private fun handleMessage(text: String) {
        try {
            val event = json.decodeFromString<WebSocketEvent>(text)
            _events.tryEmit(event)
        } catch (e: Exception) {
            Timber.e(e, "Failed to parse WebSocket message")
        }
    }

    fun sendMessage(text: String, conversationId: String, agentId: String? = null) {
        val message = WebSocketMessage(
            type = "user_message",
            content = text,
            conversationId = conversationId,
            agentId = agentId
        )
        val jsonMessage = json.encodeToString(WebSocketMessage.serializer(), message)
        webSocket?.send(jsonMessage)
    }

    fun disconnect() {
        reconnectJob?.cancel()
        stopHeartbeat()
        webSocket?.close(1000, "Client disconnect")
        webSocket = null
        _connectionState.value = ConnectionState.DISCONNECTED
    }

    private fun sendAuthMessage() {
        val token = tokenManager.getAccessToken() ?: return
        val authMessage = WebSocketAuthMessage(
            type = "auth",
            token = token
        )
        val jsonMessage = json.encodeToString(WebSocketAuthMessage.serializer(), authMessage)
        webSocket?.send(jsonMessage)
    }

    private fun startHeartbeat() {
        heartbeatJob = scope.launch {
            while (isActive) {
                delay(30_000) // 30 seconds
                webSocket?.send("""{"type":"ping"}""")
            }
        }
    }

    private fun stopHeartbeat() {
        heartbeatJob?.cancel()
    }

    private fun scheduleReconnect() {
        if (reconnectAttempts >= maxReconnectAttempts) {
            Timber.e("Max reconnect attempts reached")
            _events.tryEmit(WebSocketEvent.Error("Connection lost. Please check your network."))
            return
        }

        val delay = minOf(
            baseReconnectDelay * (1 shl reconnectAttempts),
            maxReconnectDelay
        )

        reconnectJob = scope.launch {
            delay(delay)
            reconnectAttempts++
            Timber.d("Reconnecting... attempt $reconnectAttempts")
            currentConversationId?.let { connect(it) }
        }
    }
}
```

**WebSocket Connection State Machine:**

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  DISCONNECTED│───►│  CONNECTING  │───►│  CONNECTED   │
└──────────────┘    └──────────────┘    └──────┬───────┘
       ▲                                       │
       │                                       │
       │    ┌──────────────┐    ┌──────────────┴──┐
       │    │  RECONNECTING│◄───│  DISCONNECTING   │
       │    └──────────────┘    └─────────────────┘
       │           │
       │           ▼
       │    ┌──────────────┐
       └────│  FAILED      │
            └──────────────┘
```

---

## WebSocket Message Protocol

```kotlin
// data/remote/websocket/WebSocketProtocol.kt
@Serializable
sealed class WebSocketEvent {
    @Serializable
    @SerialName("token")
    data class Token(
        val content: String,
        val messageId: String
    ) : WebSocketEvent()

    @Serializable
    @SerialName("thinking")
    data class Thinking(
        val content: String,
        val messageId: String
    ) : WebSocketEvent()

    @Serializable
    @SerialName("tool_call")
    data class ToolCall(
        val id: String,
        val name: String,
        val arguments: String,
        val messageId: String
    ) : WebSocketEvent()

    @Serializable
    @SerialName("tool_result")
    data class ToolResult(
        val toolCallId: String,
        val content: String,
        val success: Boolean,
        val messageId: String
    ) : WebSocketEvent()

    @Serializable
    @SerialName("completed")
    data class Completed(
        val messageId: String,
        val tokenUsage: TokenUsage? = null
    ) : WebSocketEvent()

    @Serializable
    @SerialName("error")
    data class Error(
        val message: String,
        val code: String? = null,
        val messageId: String? = null
    ) : WebSocketEvent()

    @Serializable
    @SerialName("connected")
    data class Connected(
        val conversationId: String,
        val agentId: String?
    ) : WebSocketEvent()
}

@Serializable
data class WebSocketMessage(
    val type: String,
    val content: String,
    val conversationId: String,
    val agentId: String? = null,
    val modelId: String? = null
)

@Serializable
data class WebSocketAuthMessage(
    val type: String = "auth",
    val token: String
)
```

**Message Protocol Flow:**

```
Client                                          Server
  │                                               │
  │──── {"type":"auth","token":"..."} ────────────►│
  │                                               │
  │◄──── {"type":"connected","conversationId":"..."} │
  │                                               │
  │──── {"type":"user_message","content":"..."} ──►│
  │                                               │
  │◄──── {"type":"thinking","content":"..."}       │
  │                                               │
  │◄──── {"type":"token","content":"Hello"}        │
  │◄──── {"type":"token","content":" world"}       │
  │◄──── {"type":"token","content":"!"}            │
  │                                               │
  │◄──── {"type":"tool_call","name":"search",...}  │
  │                                               │
  │◄──── {"type":"tool_result","success":true}     │
  │                                               │
  │◄──── {"type":"completed","tokenUsage":{...}}   │
  │                                               │
  │──── {"type":"ping"} ──────────────────────────►│
  │◄──── {"type":"pong"} ─────────────────────────│
  │                                               │
```

---

## WebSocket Authentication

```kotlin
// Option 1: Query parameter token
fun connectWithQueryToken(conversationId: String) {
    val token = tokenManager.getAccessToken() ?: return
    val request = Request.Builder()
        .url("${BuildConfig.WS_URL}/ws?token=$token&conversation_id=$conversationId")
        .build()
    webSocket = okHttpClient.newWebSocket(request, listener)
}

// Option 2: First message authentication
fun connectWithFirstMessageAuth(conversationId: String) {
    val request = Request.Builder()
        .url("${BuildConfig.WS_URL}/ws")
        .build()
    webSocket = okHttpClient.newWebSocket(request, listener)

    // Send auth message immediately
    val token = tokenManager.getAccessToken() ?: return
    val authMessage = json.encodeToString(
        WebSocketAuthMessage.serializer(),
        WebSocketAuthMessage(token = token)
    )
    webSocket?.send(authMessage)
}

// Option 3: Header authentication
fun connectWithHeaderAuth(conversationId: String) {
    val token = tokenManager.getAccessToken() ?: return
    val request = Request.Builder()
        .url("${BuildConfig.WS_URL}/ws")
        .header("Authorization", "Bearer $token")
        .build()
    webSocket = okHttpClient.newWebSocket(request, listener)
}
```

---

## WebSocket Reconnection

```kotlin
// Exponential backoff reconnection
class ReconnectManager(
    private val scope: CoroutineScope,
    private val maxAttempts: Int = 10,
    private val baseDelay: Long = 1000L,
    private val maxDelay: Long = 30000L
) {
    private var attempt = 0
    private var reconnectJob: Job? = null

    fun scheduleReconnect(onReconnect: () -> Unit) {
        if (attempt >= maxAttempts) {
            Timber.e("Max reconnect attempts reached")
            return
        }

        val delay = calculateDelay()
        Timber.d("Scheduling reconnect in ${delay}ms (attempt ${attempt + 1})")

        reconnectJob = scope.launch {
            delay(delay)
            attempt++
            onReconnect()
        }
    }

    private fun calculateDelay(): Long {
        // Exponential backoff with jitter
        val exponentialDelay = baseDelay * (1 shl attempt)
        val jitter = Random.nextLong(0, exponentialDelay / 4)
        return minOf(exponentialDelay + jitter, maxDelay)
    }

    fun reset() {
        attempt = 0
        reconnectJob?.cancel()
    }
}

// Reconnection delays table
// Attempt 1: ~1000ms
// Attempt 2: ~2000ms
// Attempt 3: ~4000ms
// Attempt 4: ~8000ms
// Attempt 5: ~16000ms
// Attempt 6+: ~30000ms (max)
```

**Reconnection Timeline:**

```
Time (ms)  0    1000   3000   7000   15000  31000  61000
           │     │      │      │      │      │      │
           ▼     ▼      ▼      ▼      ▼      ▼      ▼
         ┌───┐ ┌───┐  ┌───┐  ┌───┐  ┌───┐  ┌───┐  ┌───┐
         │   │ │   │  │   │  │   │  │   │  │   │  │   │
         │Fail│ │Try│  │Try│  │Try│  │Try│  │Try│  │Try│
         │   │ │ 1 │  │ 2 │  │ 3 │  │ 4 │  │ 5 │  │ 6 │
         └───┘ └───┘  └───┘  └───┘  └───┘  └───┘  └───┘
```

---

## API Endpoint Definitions

```kotlin
// data/remote/ApiService.kt
interface ApiService {

    // ==================== Auth ====================
    @POST("api/v1/auth/login")
    suspend fun login(@Body request: LoginRequest): Response<LoginResponse>

    @POST("api/v1/auth/register")
    suspend fun register(@Body request: RegisterRequest): Response<RegisterResponse>

    @POST("api/v1/auth/refresh")
    suspend fun refreshToken(@Body request: RefreshTokenRequest): Response<RefreshTokenResponse>

    @POST("api/v1/auth/logout")
    suspend fun logout(): Response<Unit>

    @POST("api/v1/auth/forgot-password")
    suspend fun forgotPassword(@Body request: ForgotPasswordRequest): Response<Unit>

    @POST("api/v1/auth/reset-password")
    suspend fun resetPassword(@Body request: ResetPasswordRequest): Response<Unit>

    // ==================== Users ====================
    @GET("api/v1/users/me")
    suspend fun getCurrentUser(): Response<ApiResponse<UserDto>>

    @PUT("api/v1/users/me")
    suspend fun updateCurrentUser(@Body request: UpdateUserRequest): Response<ApiResponse<UserDto>>

    @GET("api/v1/users/me/profile")
    suspend fun getUserProfile(): Response<ApiResponse<UserProfileDto>>

    // ==================== Agents ====================
    @GET("api/v1/agents")
    suspend fun getAgents(
        @Query("page") page: Int = 1,
        @Query("limit") limit: Int = 20
    ): Response<ApiResponse<List<AgentDto>>>

    @GET("api/v1/agents/{id}")
    suspend fun getAgent(@Path("id") agentId: String): Response<ApiResponse<AgentDto>>

    @POST("api/v1/agents")
    suspend fun createAgent(@Body request: CreateAgentRequest): Response<ApiResponse<AgentDto>>

    @PUT("api/v1/agents/{id}")
    suspend fun updateAgent(
        @Path("id") agentId: String,
        @Body request: UpdateAgentRequest
    ): Response<ApiResponse<AgentDto>>

    @DELETE("api/v1/agents/{id}")
    suspend fun deleteAgent(@Path("id") agentId: String): Response<Unit>

    // ==================== Conversations ====================
    @GET("api/v1/ai/conversations")
    suspend fun getConversations(
        @Query("page") page: Int = 1,
        @Query("limit") limit: Int = 20
    ): Response<ApiResponse<List<ConversationDto>>>

    @GET("api/v1/ai/conversations/{id}")
    suspend fun getConversation(@Path("id") conversationId: String): Response<ApiResponse<ConversationDto>>

    @POST("api/v1/ai/conversations")
    suspend fun createConversation(@Body request: CreateConversationRequest): Response<ApiResponse<ConversationDto>>

    @DELETE("api/v1/ai/conversations/{id}")
    suspend fun deleteConversation(@Path("id") conversationId: String): Response<Unit>

    @GET("api/v1/ai/conversations/{id}/messages")
    suspend fun getMessages(
        @Path("id") conversationId: String,
        @Query("limit") limit: Int = 50,
        @Query("offset") offset: Int = 0
    ): Response<ApiResponse<List<MessageDto>>>

    // ==================== AI Chat ====================
    @POST("api/v1/ai/chat")
    suspend fun sendChatMessage(@Body request: ChatRequest): Response<ApiResponse<ChatResponse>>

    // ==================== Models ====================
    @GET("api/v1/models")
    suspend fun getModels(): Response<ApiResponse<List<ModelDto>>>

    @GET("api/v1/models/{id}")
    suspend fun getModel(@Path("id") modelId: String): Response<ApiResponse<ModelDto>>

    // ==================== Knowledge ====================
    @GET("api/v1/knowledge")
    suspend fun getKnowledge(
        @Query("page") page: Int = 1,
        @Query("limit") limit: Int = 20
    ): Response<ApiResponse<List<KnowledgeDto>>>

    // ==================== RAG Knowledge ====================
    @GET("api/v1/rag/documents")
    suspend fun getDocuments(
        @Query("page") page: Int = 1,
        @Query("limit") limit: Int = 20
    ): Response<ApiResponse<List<DocumentDto>>>

    @Multipart
    @POST("api/v1/rag/documents")
    suspend fun uploadDocument(
        @Part file: MultipartBody.Part
    ): Response<ApiResponse<DocumentUploadResponse>>

    @GET("api/v1/rag/documents/{id}/status")
    suspend fun getDocumentStatus(
        @Path("id") documentId: String
    ): Response<ApiResponse<DocumentStatusResponse>>

    @POST("api/v1/rag/search")
    suspend fun searchKnowledge(
        @Body request: SearchRequest
    ): Response<ApiResponse<SearchResponse>>

    @DELETE("api/v1/rag/documents/{id}")
    suspend fun deleteDocument(
        @Path("id") documentId: String
    ): Response<Unit>

    // ==================== Vision ====================
    @Multipart
    @POST("api/v1/vision/analyze")
    suspend fun analyzeImage(
        @Part image: MultipartBody.Part,
        @Part("task") task: RequestBody?
    ): Response<ApiResponse<VisionAnalysisResponse>>

    // ==================== SQL Intelligence ====================
    @POST("api/v1/sql/query")
    suspend fun executeSqlQuery(
        @Body request: SqlQueryRequest
    ): Response<ApiResponse<SqlQueryResponse>>

    // ==================== Memory ====================
    @POST("api/v1/memory")
    suspend fun storeMemory(
        @Body request: StoreMemoryRequest
    ): Response<Unit>

    @GET("api/v1/memory/search")
    suspend fun searchMemory(
        @Query("q") query: String,
        @Query("limit") limit: Int = 5
    ): Response<ApiResponse<List<MemoryDto>>>

    // ==================== Workflows ====================
    @POST("api/v1/workflows/start")
    suspend fun startWorkflow(
        @Body request: StartWorkflowRequest
    ): Response<ApiResponse<WorkflowResponse>>

    @GET("api/v1/workflows/{id}")
    suspend fun getWorkflow(
        @Path("id") workflowId: String
    ): Response<ApiResponse<WorkflowStatusResponse>>

    // ==================== KYC ====================
    @GET("api/v1/kyc/status")
    suspend fun getKycStatus(): Response<ApiResponse<KycStatusResponse>>

    @Multipart
    @POST("api/v1/kyc/documents")
    suspend fun uploadKycDocument(
        @Part file: MultipartBody.Part
    ): Response<ApiResponse<KycDocumentResponse>>

    @POST("api/v1/kyc/submit")
    suspend fun submitKyc(): Response<Unit>
}
```

**API Endpoint Map:**

```
┌─────────────────────────────────────────────────────────┐
│                      API Endpoints                       │
├─────────────────────────────────────────────────────────┤
│  Auth                                                    │
│  ├── POST   /api/v1/auth/login                          │
│  ├── POST   /api/v1/auth/register                       │
│  ├── POST   /api/v1/auth/refresh                        │
│  ├── POST   /api/v1/auth/logout                         │
│  ├── POST   /api/v1/auth/forgot-password                │
│  └── POST   /api/v1/auth/reset-password                 │
│                                                          │
│  Users                                                   │
│  ├── GET    /api/v1/users/me                            │
│  ├── PUT    /api/v1/users/me                            │
│  └── GET    /api/v1/users/me/profile                    │
│                                                          │
│  Agents                                                  │
│  ├── GET    /api/v1/agents                              │
│  ├── GET    /api/v1/agents/{id}                         │
│  ├── POST   /api/v1/agents                              │
│  ├── PUT    /api/v1/agents/{id}                         │
│  └── DELETE /api/v1/agents/{id}                         │
│                                                          │
│  Conversations                                           │
│  ├── GET    /api/v1/ai/conversations                       │
│  ├── GET    /api/v1/ai/conversations/{id}                  │
│  ├── POST   /api/v1/ai/conversations                       │
│  ├── DELETE /api/v1/ai/conversations/{id}                  │
│  └── GET    /api/v1/ai/conversations/{id}/messages         │
│                                                          │
│  AI Chat                                                 │
│  └── POST   /api/v1/ai/chat                               │
│                                                          │
│  Models                                                  │
│  ├── GET    /api/v1/models                              │
│  └── GET    /api/v1/models/{id}                         │
│                                                          │
│  Knowledge                                               │
│  ├── GET    /api/v1/rag/documents                        │
│  ├── POST   /api/v1/rag/documents                        │
│  ├── GET    /api/v1/rag/documents/{id}/status             │
│  ├── POST   /api/v1/rag/search                           │
│  └── DELETE /api/v1/rag/documents/{id}                    │
│                                                          │
│  Vision                                                  │
│  ├── POST   /api/v1/vision/analyze                       │
│  ├── POST   /api/v1/vision/ocr                           │
│  └── POST   /api/v1/vision/batch                         │
│                                                          │
│  SQL Intelligence                                        │
│  └── POST   /api/v1/sql/query                            │
│                                                          │
│  Memory                                                  │
│  ├── POST   /api/v1/memory                               │
│  ├── GET    /api/v1/memory/search                         │
│  └── GET    /api/v1/memory/context/{session_id}           │
│                                                          │
│  Workflows                                               │
│  ├── POST   /api/v1/workflows/start                       │
│  └── GET    /api/v1/workflows/{id}                        │
│                                                          │
│  KYC                                                     │
│  ├── GET    /api/v1/kyc/status                            │
│  ├── POST   /api/v1/kyc/documents                         │
│  └── POST   /api/v1/kyc/submit                            │
└─────────────────────────────────────────────────────────┘
```

---

## Request/Response DTOs

```kotlin
// data/remote/dto/LoginRequest.kt
@Serializable
data class LoginRequest(
    @SerialName("email") val email: String,
    @SerialName("password") val password: String,
    @SerialName("device_id") val deviceId: String? = null
)

@Serializable
data class LoginResponse(
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String,
    @SerialName("token_type") val tokenType: String = "Bearer",
    @SerialName("expires_in") val expiresIn: Long,
    @SerialName("user") val user: UserDto
)

// data/remote/dto/UserDto.kt
@Serializable
data class UserDto(
    @SerialName("id") val id: String,
    @SerialName("email") val email: String,
    @SerialName("name") val name: String,
    @SerialName("avatar_url") val avatarUrl: String? = null,
    @SerialName("role") val role: String,
    @SerialName("tenant_id") val tenantId: String,
    @SerialName("created_at") val createdAt: String,
    @SerialName("updated_at") val updatedAt: String
)

// data/remote/dto/AgentDto.kt
@Serializable
data class AgentDto(
    @SerialName("id") val id: String,
    @SerialName("name") val name: String,
    @SerialName("description") val description: String?,
    @SerialName("model_id") val modelId: String,
    @SerialName("system_prompt") val systemPrompt: String?,
    @SerialName("avatar_url") val avatarUrl: String?,
    @SerialName("capabilities") val capabilities: List<String>,
    @SerialName("is_active") val isActive: Boolean = true,
    @SerialName("created_at") val createdAt: String
)

// data/remote/dto/MessageDto.kt
@Serializable
data class MessageDto(
    @SerialName("id") val id: String,
    @SerialName("conversation_id") val conversationId: String,
    @SerialName("role") val role: String,
    @SerialName("content") val content: String,
    @SerialName("timestamp") val timestamp: String,
    @SerialName("tool_calls") val toolCalls: List<ToolCallDto>? = null,
    @SerialName("tool_results") val toolResults: List<ToolResultDto>? = null,
    @SerialName("token_usage") val tokenUsage: TokenUsageDto? = null,
    @SerialName("model_id") val modelId: String? = null
)

@Serializable
data class ToolCallDto(
    @SerialName("id") val id: String,
    @SerialName("name") val name: String,
    @SerialName("arguments") val arguments: String
)

@Serializable
data class ToolResultDto(
    @SerialName("tool_call_id") val toolCallId: String,
    @SerialName("content") val content: String,
    @SerialName("success") val success: Boolean
)

@Serializable
data class TokenUsageDto(
    @SerialName("prompt_tokens") val promptTokens: Int,
    @SerialName("completion_tokens") val completionTokens: Int,
    @SerialName("total_tokens") val totalTokens: Int
)

// data/remote/dto/ApiResponse.kt
@Serializable
data class ApiResponse<T>(
    @SerialName("success") val success: Boolean,
    @SerialName("data") val data: T,
    @SerialName("message") val message: String? = null,
    @SerialName("error") val error: String? = null
)

@Serializable
data class ErrorResponse(
    @SerialName("code") val code: String,
    @SerialName("message") val message: String,
    @SerialName("details") val details: Map<String, String>? = null
)
```

---

## Error Handling

```kotlin
// common/type/NetworkResult.kt
sealed class NetworkResult<out T> {
    data class Success<T>(val data: T) : NetworkResult<T>()
    data class Error(val code: Int, val message: String, val exception: Throwable? = null) : NetworkResult<Nothing>()
    data object Loading : NetworkResult<Nothing>()
}

// Extension function to convert Response to NetworkResult
suspend fun <T> safeApiCall(apiCall: suspend () -> Response<T>): NetworkResult<T> {
    return try {
        val response = apiCall()
        if (response.isSuccessful) {
            val body = response.body()
            if (body != null) {
                NetworkResult.Success(body)
            } else {
                NetworkResult.Error(0, "Empty response body")
            }
        } else {
            val errorBody = response.errorBody()?.string()
            val errorResponse = try {
                Json.decodeFromString<ErrorResponse>(errorBody ?: "")
            } catch (e: Exception) {
                null
            }
            NetworkResult.Error(
                code = response.code(),
                message = errorResponse?.message ?: response.message(),
                exception = ApiException(response.code(), errorBody)
            )
        }
    } catch (e: HttpException) {
        NetworkResult.Error(code = e.code(), message = e.message(), exception = e)
    } catch (e: IOException) {
        NetworkResult.Error(code = -1, message = "Network error: ${e.message}", exception = e)
    } catch (e: Exception) {
        NetworkResult.Error(code = -1, message = e.message ?: "Unknown error", exception = e)
    }
}

// Error hierarchy
sealed class NetworkException : Exception() {
    data class Api(val code: Int, override val message: String) : NetworkException()
    data class Connectivity(override val message: String = "No internet connection") : NetworkException()
    data class Timeout(override val message: String = "Request timed out") : NetworkException()
    data class Unauthorized(override val message: String = "Authentication required") : NetworkException()
    data class Forbidden(override val message: String = "Access denied") : NetworkException()
    data class NotFound(override val message: String = "Resource not found") : NetworkException()
    data class RateLimited(override val message: String = "Too many requests") : NetworkException()
    data class ServerError(override val message: String = "Server error") : NetworkException()
    data class Parsing(override val message: String = "Failed to parse response") : NetworkException()
}

// Error mapping
fun mapHttpError(code: Int, message: String?): NetworkException {
    return when (code) {
        401 -> NetworkException.Unauthorized(message ?: "Unauthorized")
        403 -> NetworkException.Forbidden(message ?: "Forbidden")
        404 -> NetworkException.NotFound(message ?: "Not found")
        429 -> NetworkException.RateLimited(message ?: "Rate limited")
        in 500..599 -> NetworkException.ServerError(message ?: "Server error")
        else -> NetworkException.Api(code, message ?: "API error")
    }
}
```

**Error Handling Flow:**

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   API Call   │───►│   Try/Catch  │───►│   Success    │
└──────────────┘    └──────┬───────┘    └──────┬───────┘
                           │                    │
                    Exception                  │
                           │                    │
                    ┌──────┴───────┐            │
                    │  Exception   │            │
                    │  Type?       │            │
                    └──────┬───────┘            │
                           │                    │
              ┌────────────┼────────────┐       │
              ▼            ▼            ▼       │
       ┌──────────┐ ┌──────────┐ ┌──────────┐  │
       │  HTTP    │ │  IO      │ │  Other   │  │
       │  Error   │ │  Error   │ │  Error   │  │
       └────┬─────┘ └────┬─────┘ └────┬─────┘  │
            │            │            │         │
            ▼            ▼            ▼         │
     ┌──────────┐ ┌──────────┐ ┌──────────┐    │
     │  Map to  │ │  Network │ │  Generic │    │
     │  Status  │ │  Error   │ │  Error   │    │
     │  Code    │ │          │ │          │    │
     └────┬─────┘ └────┬─────┘ └────┬─────┘    │
          │            │            │           │
          └────────────┼────────────┘           │
                       ▼                        │
                ┌─────────────┐                 │
                │ NetworkError│                 │
                └──────┬──────┘                 │
                       │                        │
                       ▼                        │
                ┌─────────────┐                 │
                │  UI Layer   │◄────────────────┘
                │  (Display)  │
                └─────────────┘
```

---

## Offline Detection

```kotlin
// core-network/src/main/java/com/nexus/ai/core/network/NetworkMonitor.kt
@Singleton
class NetworkMonitor @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val connectivityManager =
        context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager

    private val _isOnline = MutableStateFlow(false)
    val isOnline: StateFlow<Boolean> = _isOnline.asStateFlow()

    private val _networkType = MutableStateFlow(NetworkType.UNKNOWN)
    val networkType: StateFlow<NetworkType> = _networkType.asStateFlow()

    private val networkCallback = object : ConnectivityManager.NetworkCallback() {
        override fun onAvailable(network: Network) {
            super.onAvailable(network)
            _isOnline.value = true
            updateNetworkType()
        }

        override fun onLost(network: Network) {
            super.onLost(network)
            _isOnline.value = false
            _networkType.value = NetworkType.UNKNOWN
        }

        override fun onCapabilitiesChanged(
            network: Network,
            capabilities: NetworkCapabilities
        ) {
            super.onCapabilitiesChanged(network, capabilities)
            updateNetworkType(capabilities)
        }
    }

    fun startMonitoring() {
        val request = NetworkRequest.Builder()
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .build()
        connectivityManager.registerNetworkCallback(request, networkCallback)

        // Check initial state
        val currentNetwork = connectivityManager.activeNetwork
        val capabilities = connectivityManager.getNetworkCapabilities(currentNetwork)
        _isOnline.value = capabilities != null
        updateNetworkType(capabilities)
    }

    fun stopMonitoring() {
        connectivityManager.unregisterNetworkCallback(networkCallback)
    }

    private fun updateNetworkType(capabilities: NetworkCapabilities? = null) {
        val caps = capabilities ?: run {
            val network = connectivityManager.activeNetwork
            connectivityManager.getNetworkCapabilities(network)
        }

        _networkType.value = when {
            caps == null -> NetworkType.UNKNOWN
            caps.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) -> NetworkType.WIFI
            caps.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) -> NetworkType.CELLULAR
            caps.hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET) -> NetworkType.ETHERNET
            caps.hasTransport(NetworkCapabilities.TRANSPORT_VPN) -> NetworkType.VPN
            else -> NetworkType.OTHER
        }
    }
}

enum class NetworkType {
    WIFI, CELLULAR, ETHERNET, VPN, OTHER, UNKNOWN
}
```

---

## Retry Logic

```kotlin
// core-network/src/main/java/com/nexus/ai/core/network/RetryPolicy.kt
class RetryPolicy(
    private val maxRetries: Int = 3,
    private val initialDelay: Long = 1000L,
    private val maxDelay: Long = 10000L,
    private val factor: Double = 2.0,
    private val retryOn: (Throwable) -> Boolean = { it is IOException }
) {
    suspend fun <T> execute(block: suspend () -> T): T {
        var currentDelay = initialDelay
        var lastException: Exception? = null

        repeat(maxRetries + 1) { attempt ->
            try {
                return block()
            } catch (e: Exception) {
                lastException = e
                if (attempt < maxRetries && retryOn(e)) {
                    Timber.w(e, "Retry attempt ${attempt + 1}/$maxRetries after ${currentDelay}ms")
                    delay(currentDelay)
                    currentDelay = (currentDelay * factor).toLong().coerceAtMost(maxDelay)
                }
            }
        }

        throw lastException ?: RuntimeException("Retry failed")
    }
}

// Usage
val retryPolicy = RetryPolicy(
    maxRetries = 3,
    initialDelay = 1000L,
    retryOn = { it is IOException || it is HttpException }
)

val result = retryPolicy.execute {
    apiService.getAgents()
}
```

**Retry Backoff Schedule:**

```
Attempt  Delay (ms)  Total Wait
─────────────────────────────────
1        0           0
2        1000        1000
3        2000        3000
4        4000        7000
5        8000        15000
6        10000       25000 (capped)
```

---

## Request Cancellation

```kotlin
// data/remote/interceptor/CancelInterceptor.kt
class CancelInterceptor @Inject constructor() : Interceptor {
    private val activeCalls = ConcurrentHashMap<String, Call>()

    override fun intercept(chain: Interceptor.Chain): Response {
        val call = chain.call()
        val requestId = chain.request().tag(String::class.java) ?: UUID.randomUUID().toString()

        activeCalls[requestId] = call

        return try {
            chain.proceed(chain.request())
        } finally {
            activeCalls.remove(requestId)
        }
    }

    fun cancelRequest(requestId: String) {
        activeCalls[requestId]?.cancel()
        activeCalls.remove(requestId)
    }

    fun cancelAll() {
        activeCalls.values.forEach { it.cancel() }
        activeCalls.clear()
    }
}

// ViewModel usage with CoroutineScope
class ChatViewModel @Inject constructor(
    private val apiService: ApiService
) : ViewModel() {

    private var fetchJob: Job? = null

    fun loadAgents() {
        fetchJob?.cancel() // Cancel previous request
        fetchJob = viewModelScope.launch {
            apiService.getAgents()
        }
    }

    fun cancelAllRequests() {
        fetchJob?.cancel()
    }

    override fun onCleared() {
        super.onCleared()
        cancelAllRequests()
    }
}
```

---

## Mock Data for Testing

```kotlin
// test-network/src/main/java/com/nexus/ai/test/network/MockApiService.kt
class MockApiService : ApiService {

    private val mockAgents = listOf(
        AgentDto(
            id = "agent-1",
            name = "Nexus Assistant",
            description = "General purpose AI assistant",
            modelId = "model-1",
            systemPrompt = "You are a helpful assistant.",
            avatarUrl = null,
            capabilities = listOf("chat", "search", "code"),
            isActive = true,
            createdAt = "2026-01-01T00:00:00Z"
        )
    )

    override suspend fun getAgents(page: Int, limit: Int): Response<ApiResponse<List<AgentDto>>> {
        return Response.success(
            ApiResponse(
                success = true,
                data = mockAgents
            )
        )
    }

    override suspend fun login(request: LoginRequest): Response<LoginResponse> {
        return if (request.email == "test@example.com" && request.password == "password123") {
            Response.success(
                LoginResponse(
                    accessToken = "mock-access-token",
                    refreshToken = "mock-refresh-token",
                    expiresIn = 3600,
                    user = UserDto(
                        id = "user-1",
                        email = "test@example.com",
                        name = "Test User",
                        role = "user",
                        tenantId = "tenant-1",
                        createdAt = "2026-01-01T00:00:00Z",
                        updatedAt = "2026-01-01T00:00:00Z"
                    )
                )
            )
        } else {
            Response.error(401, "{\"error\":\"Invalid credentials\"}".toResponseBody())
        }
    }
}

// test-network/src/main/java/com/nexus/ai/test/network/MockWebSocketServer.kt
class MockWebSocketServer {
    private val server = MockWebServer()
    private val mockMessages = mutableListOf<String>()

    fun start() {
        server.start(8080)
    }

    fun enqueueMessage(message: String) {
        mockMessages.add(message)
    }

    fun enqueueDisconnect() {
        server.enqueue(MockResponse().withSocketClose())
    }

    fun enqueueError(code: Int, message: String) {
        server.enqueue(MockResponse().setResponseCode(code).setBody(message))
    }

    fun getBaseUrl(): String = server.url("/").toString()

    fun shutdown() {
        server.shutdown()
    }
}
```

---

## Network Monitoring

```kotlin
// core-network/src/main/java/com/nexus/ai/core/network/NetworkCallback.kt
@Singleton
class NetworkCallback @Inject constructor(
    @ApplicationContext private val context: Context,
    private val networkMonitor: NetworkMonitor
) {
    private val callback = object : ConnectivityManager.NetworkCallback() {
        override fun onAvailable(network: Network) {
            Timber.d("Network available: $network")
            // Trigger sync when network becomes available
            CoroutineScope(Dispatchers.IO).launch {
                syncRepository.syncPendingData()
            }
        }

        override fun onLost(network: Network) {
            Timber.d("Network lost: $network")
            // Queue operations for later
        }

        override fun onCapabilitiesChanged(
            network: Network,
            networkCapabilities: NetworkCapabilities
        ) {
            val downSpeed = networkCapabilities.getLinkDownstreamBandwidthKbps()
            val upSpeed = networkCapabilities.getLinkUpstreamBandwidthKbps()
            Timber.d("Network capabilities changed: down=${downSpeed}kbps, up=${upSpeed}kbps")
        }
    }

    fun register() {
        val request = NetworkRequest.Builder()
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .build()
        val connectivityManager =
            context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        connectivityManager.registerNetworkCallback(request, callback)
    }

    fun unregister() {
        val connectivityManager =
            context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        connectivityManager.unregisterNetworkCallback(callback)
    }
}
```

**Network Monitoring Dashboard:**

```
┌─────────────────────────────────────────────────────────┐
│                 Network Status Monitor                   │
├─────────────────────────────────────────────────────────┤
│  Connection: ● Connected (WiFi)                         │
│  Latency:    45ms                                       │
│  Bandwidth:  150 Mbps down / 30 Mbps up                 │
│                                                          │
│  Active Connections: 3                                   │
│  ├── api.nexus-ai.com:443 (keep-alive)                 │
│  ├── ws.nexus-ai.com:443 (websocket)                   │
│  └── firebase.googleapis.com:443 (push)                │
│                                                          │
│  Request Queue:                                          │
│  ├── GET /api/v1/agents      [pending]    12ms          │
│  ├── GET /api/v1/models      [completed]  45ms          │
│  └── POST /api/v1/ai/chat    [completed]  120ms         │
│                                                          │
│  Cache:                                                  │
│  ├── Hit Rate: 85%                                      │
│  ├── Size: 2.3 MB                                       │
│  └── Evictions: 12                                      │
└─────────────────────────────────────────────────────────┘
```

---

*Document Version: 1.0 | Last Updated: 2026-07-16*
