# 19 — Error Handling

## 1. Error Handling Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                     Error Handling Layers                        │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  UI Layer          │ Snackbars, Dialogs, Empty States   │   │
│  ├────────────────────┼─────────────────────────────────────┤   │
│  │  ViewModel Layer   │ Result<>, Sealed Classes, Events    │   │
│  ├────────────────────┼─────────────────────────────────────┤   │
│  │  Domain Layer      │ Use Case errors, Validation         │   │
│  ├────────────────────┼─────────────────────────────────────┤   │
│  │  Data Layer        │ Network, DB, WebSocket errors       │   │
│  ├────────────────────┼─────────────────────────────────────┤   │
│  │  Platform Layer    │ System errors, permissions, certs   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  Cross-cutting: Crashlytics, Timber, Error Boundaries            │
└──────────────────────────────────────────────────────────────────┘
```

### Result Type

```kotlin
sealed class Result<out T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val exception: NexusException) : Result<Nothing>()
    data object Loading : Result<Nothing>()

    val isSuccess: Boolean get() = this is Success
    val isFailure: Boolean get() = this is Error

    fun getOrNull(): T? = (this as? Success)?.data
    fun exceptionOrNull(): NexusException? = (this as? Error)?.exception

    fun <R> map(transform: (T) -> R): Result<R> = when (this) {
        is Success -> Success(transform(data))
        is Error -> this
        is Loading -> this
    }

    fun onSuccess(block: (T) -> Unit): Result<T> {
        if (this is Success) block(data)
        return this
    }

    fun onError(block: (NexusException) -> Unit): Result<T> {
        if (this is Error) block(exception)
        return this
    }
}

suspend fun <T> runCatchingResult(block: suspend () -> T): Result<T> {
    return try {
        Result.Success(block())
    } catch (e: NexusException) {
        Result.Error(e)
    } catch (e: Exception) {
        Result.Error(e.toNexusException())
    }
}
```

### Sealed Error Classes

```kotlin
sealed class NexusException(
    override val message: String,
    open val code: String = "UNKNOWN",
    open val recoverable: Boolean = true,
    open val userMessage: String = "Something went wrong"
) : Exception(message) {

    data class NetworkException(
        override val message: String = "Network error",
        val cause: Throwable? = null
    ) : NexusException(
        message = message,
        code = "NETWORK_ERROR",
        recoverable = true,
        userMessage = "Check your internet connection and try again"
    )

    data class AuthException(
        override val message: String = "Authentication failed",
        val authCode: AuthErrorCode = AuthErrorCode.UNKNOWN
    ) : NexusException(
        message = message,
        code = "AUTH_ERROR",
        recoverable = when (authCode) {
            AuthErrorCode.TOKEN_EXPIRED, AuthErrorCode.INVALID_CREDENTIALS -> true
            AuthErrorCode.ACCOUNT_LOCKED, AuthErrorCode.KYC_PENDING -> false
            else -> false
        },
        userMessage = when (authCode) {
            AuthErrorCode.TOKEN_EXPIRED -> "Session expired. Please log in again."
            AuthErrorCode.INVALID_CREDENTIALS -> "Invalid email or password."
            AuthErrorCode.ACCOUNT_LOCKED -> "Account locked. Contact support."
            AuthErrorCode.KYC_PENDING -> "KYC verification pending."
            else -> "Authentication failed. Please try again."
        }
    )

    data class ValidationException(
        override val message: String,
        val field: String? = null,
        val errors: Map<String, String> = emptyMap()
    ) : NexusException(
        message = message,
        code = "VALIDATION_ERROR",
        recoverable = true,
        userMessage = if (field != null) message else "Please check your input"
    )

    data class ServerException(
        override val message: String = "Server error",
        val httpCode: Int = 500
    ) : NexusException(
        message = message,
        code = "SERVER_ERROR_$httpCode",
        recoverable = httpCode in 500..503,
        userMessage = when (httpCode) {
            500 -> "Server error. Please try again later."
            502 -> "Service temporarily unavailable."
            503 -> "Service is under maintenance."
            504 -> "Server timeout. Please try again."
            else -> "Something went wrong."
        }
    )

    data class RateLimitException(
        val retryAfterMs: Long = 60_000
    ) : NexusException(
        message = "Rate limit exceeded",
        code = "RATE_LIMIT",
        recoverable = true,
        userMessage = "Too many requests. Please wait a moment."
    )

    data class WebSocketException(
        override val message: String = "WebSocket error",
        val wsCode: Int = 1000
    ) : NexusException(
        message = message,
        code = "WS_ERROR",
        recoverable = wsCode in listOf(1000, 1001, 1006, 1011),
        userMessage = "Connection lost. Reconnecting..."
    )

    data class CacheException(
        override val message: String = "Cache error"
    ) : NexusException(
        message = message,
        code = "CACHE_ERROR",
        recoverable = true,
        userMessage = "Unable to load cached data"
    )

    data class UnknownException(
        override val message: String = "Unknown error",
        val cause: Throwable? = null
    ) : NexusException(
        message = message,
        code = "UNKNOWN",
        recoverable = false,
        userMessage = "Something went wrong. Please try again."
    )
}

enum class AuthErrorCode {
    TOKEN_EXPIRED,
    INVALID_CREDENTIALS,
    ACCOUNT_LOCKED,
    KYC_PENDING,
    UNKNOWN
}

fun Throwable.toNexusException(): NexusException = when (this) {
    is NexusException -> this
    is java.net.UnknownHostException ->
        NexusException.NetworkException("No internet connection", this)
    is java.net.SocketTimeoutException ->
        NexusException.NetworkException("Connection timed out", this)
    is java.io.IOException ->
        NexusException.NetworkException("Network error", this)
    is kotlinx.serialization.SerializationException ->
        NexusException.UnknownException("Data parsing error", this)
    is SecurityException ->
        NexusException.AuthException("Security error", AuthErrorCode.UNKNOWN)
    else ->
        NexusException.UnknownException(this.message ?: "Unknown error", this)
}
```

---

## 2. Error Classification

```
┌──────────────────────────────────────────────────────────────────┐
│                    Error Classification Matrix                   │
│                                                                  │
│  Category        │ Recoverable │ Retry?  │ User Action Needed   │
├──────────────────────────────────────────────────────────────────┤
│  Network         │ Yes         │ Yes     │ Check connection     │
│  Auth            │ Varies      │ Maybe   │ Re-login / fix       │
│  Validation      │ Yes         │ No      │ Fix input            │
│  Server (5xx)    │ Yes         │ Yes     │ Wait / retry         │
│  Client (4xx)    │ Varies      │ No*     │ Fix request          │
│  Rate Limit      │ Yes         │ Yes*    │ Wait                 │
│  WebSocket       │ Yes         │ Auto    │ None (auto-reconnect)│
│  Cache           │ Yes         │ Yes     │ None (fallback)      │
│  Unknown         │ Varies      │ Maybe   │ Report / retry       │
└──────────────────────────────────────────────────────────────────┘
```

### Recovery Strategies

```
Network Error ──────► Cache Fallback ──► Offline Mode
Server Error ──────► Retry (3x) ──────► Degraded Mode
Auth Error ────────► Refresh Token ───► Re-login
Validation ────────► Inline Error ────► Fix Input
Rate Limit ────────► Wait + Retry ────► Backoff
WebSocket ─────────► Auto Reconnect ──► FCM Fallback
Unknown ───────────► Log + Retry ─────► User Report
```

### Error Classification Table

| Error Type        | Code Prefix   | Recoverable | Retry    | User Sees               |
|-------------------|--------------|-------------|----------|--------------------------|
| No Internet       | NETWORK      | Yes         | Auto     | Snackbar + retry         |
| Timeout           | NETWORK      | Yes         | Yes      | Snackbar + retry         |
| SSL Error         | NETWORK      | No          | No       | Full-screen error        |
| Token Expired     | AUTH         | Yes         | Auto     | Redirect to login        |
| Invalid Creds     | AUTH         | No          | No       | Inline field error       |
| Account Locked    | AUTH         | No          | No       | Full-screen + support    |
| 500 Server Error  | SERVER_500   | Yes         | Yes      | Snackbar + retry         |
| 503 Maintenance   | SERVER_503   | Yes         | Yes      | Full-screen + countdown  |
| 429 Rate Limit    | RATE_LIMIT   | Yes         | Wait     | Snackbar + countdown     |
| Validation        | VALIDATION   | Yes         | No       | Inline field errors      |
| WebSocket Drop    | WS_ERROR     | Yes         | Auto     | "Reconnecting..." banner |
| Cache Miss        | CACHE        | Yes         | Yes      | Empty state              |
| Unknown           | UNKNOWN      | No          | Maybe    | Generic error dialog     |

---

## 3. Network Errors

```kotlin
class NetworkErrorHandler(
    private val connectivityChecker: ConnectivityChecker,
    private val cacheManager: CacheManager
) {

    suspend fun <T> handleNetworkCall(
        cacheKey: String? = null,
        block: suspend () -> T
    ): Result<T> {
        return try {
            if (!connectivityChecker.isConnected()) {
                if (cacheKey != null) {
                    val cached = cacheManager.get<T>(cacheKey)
                    if (cached != null) return Result.Success(cached)
                }
                return Result.Error(
                    NexusException.NetworkException("No internet connection")
                )
            }
            val result = block()
            if (cacheKey != null) cacheManager.put(cacheKey, result)
            Result.Success(result)
        } catch (e: java.net.UnknownHostException) {
            handleNoInternet(cacheKey)
        } catch (e: java.net.SocketTimeoutException) {
            handleTimeout(cacheKey)
        } catch (e: javax.net.ssl.SSLException) {
            handleSSLError(e)
        } catch (e: IOException) {
            handleIOError(e, cacheKey)
        } catch (e: HttpException) {
            handleHttpError(e, cacheKey)
        }
    }

    private suspend fun <T> handleNoInternet(cacheKey: String?): Result<T> {
        if (cacheKey != null) {
            val cached = cacheManager.get<T>(cacheKey)
            if (cached != null) return Result.Success(cached)
        }
        return Result.Error(NexusException.NetworkException("No internet connection"))
    }

    private suspend fun <T> handleTimeout(cacheKey: String?): Result<T> {
        if (cacheKey != null) {
            val cached = cacheManager.get<T>(cacheKey)
            if (cached != null) return Result.Success(cached)
        }
        return Result.Error(NexusException.NetworkException("Connection timed out"))
    }

    private fun handleSSLError(e: javax.net.ssl.SSLException): Result.Error {
        Timber.e(e, "SSL error")
        return Result.Error(NexusException.NetworkException("Secure connection failed"))
    }

    private suspend fun <T> handleIOError(e: IOException, cacheKey: String?): Result<T> {
        if (cacheKey != null) {
            val cached = cacheManager.get<T>(cacheKey)
            if (cached != null) return Result.Success(cached)
        }
        return Result.Error(NexusException.NetworkException("Network error", e))
    }

    private suspend fun <T> handleHttpError(e: HttpException, cacheKey: String?): Result<T> {
        return when (e.code()) {
            400 -> Result.Error(NexusException.ServerException("Bad request", 400))
            401 -> Result.Error(NexusException.AuthException(
                authCode = AuthErrorCode.TOKEN_EXPIRED))
            403 -> Result.Error(NexusException.AuthException("Access denied"))
            404 -> Result.Error(NexusException.ServerException("Not found", 404))
            409 -> Result.Error(NexusException.ServerException("Conflict", 409))
            422 -> Result.Error(NexusException.ValidationException("Invalid data"))
            429 -> {
                val retryAfter = e.response()?.headers()
                    ?.get("Retry-After")?.toLongOrNull() ?: 60
                Result.Error(NexusException.RateLimitException(retryAfter * 1000))
            }
            in 500..599 -> {
                if (cacheKey != null) {
                    val cached = cacheManager.get<T>(cacheKey)
                    if (cached != null) return Result.Success(cached)
                }
                Result.Error(NexusException.ServerException("Server error", e.code()))
            }
            else -> Result.Error(NexusException.ServerException("HTTP ${e.code()}", e.code()))
        }
    }
}
```

---

## 4. Auth Errors

```kotlin
class AuthErrorHandler(
    private val tokenRefreshService: TokenRefreshService,
    private val sessionManager: SessionManager
) {

    suspend fun <T> handleAuthCall(block: suspend () -> T): Result<T> {
        return try {
            Result.Success(block())
        } catch (e: HttpException) {
            when (e.code()) {
                401 -> handleUnauthorized(e, block)
                403 -> Result.Error(NexusException.AuthException("Access denied"))
                else -> Result.Error(e.toNexusException())
            }
        }
    }

    private suspend fun <T> handleUnauthorized(
        e: HttpException,
        originalBlock: suspend () -> T
    ): Result<T> {
        val errorBody = e.response()?.errorBody()?.string()
        val errorResponse = parseErrorResponse(errorBody)

        return when (errorResponse?.error) {
            "token_expired" -> {
                val refreshed = tokenRefreshService.refreshToken()
                if (refreshed) {
                    try {
                        Result.Success(originalBlock())
                    } catch (retryError: Exception) {
                        sessionManager.clearSession()
                        Result.Error(NexusException.AuthException(
                            authCode = AuthErrorCode.TOKEN_EXPIRED))
                    }
                } else {
                    sessionManager.clearSession()
                    Result.Error(NexusException.AuthException(
                        authCode = AuthErrorCode.TOKEN_EXPIRED))
                }
            }
            "invalid_credentials" -> {
                Result.Error(NexusException.AuthException(
                    authCode = AuthErrorCode.INVALID_CREDENTIALS))
            }
            "account_locked" -> {
                Result.Error(NexusException.AuthException(
                    message = "Account locked",
                    authCode = AuthErrorCode.ACCOUNT_LOCKED))
            }
            "kyc_pending" -> {
                Result.Error(NexusException.AuthException(
                    authCode = AuthErrorCode.KYC_PENDING))
            }
            else -> {
                sessionManager.clearSession()
                Result.Error(NexusException.AuthException(
                    authCode = AuthErrorCode.TOKEN_EXPIRED))
            }
        }
    }
}
```

---

## 5. Validation Errors

```kotlin
class TransactionValidator {

    fun validate(request: TransactionRequest): Result<Unit> {
        val errors = mutableMapOf<String, String>()

        if (request.toAddress.isBlank()) {
            errors["toAddress"] = "Recipient address is required"
        } else if (!isValidAddress(request.toAddress)) {
            errors["toAddress"] = "Invalid recipient address"
        }

        if (request.amount <= BigDecimal.ZERO) {
            errors["amount"] = "Amount must be greater than zero"
        } else if (request.amount > BigDecimal("1000000")) {
            errors["amount"] = "Amount exceeds maximum limit"
        }

        if (request.currency.isBlank()) {
            errors["currency"] = "Currency is required"
        } else if (request.currency !in listOf("USDC", "ETH", "BTC")) {
            errors["currency"] = "Unsupported currency"
        }

        return if (errors.isEmpty()) {
            Result.Success(Unit)
        } else {
            Result.Error(NexusException.ValidationException(
                message = "Validation failed",
                errors = errors
            ))
        }
    }

    private fun isValidAddress(address: String): Boolean {
        return address.startsWith("0x") && address.length == 42
    }
}
```

### Form Validation in Compose

```kotlin
@Composable
fun TransactionForm(onConfirm: (TransactionRequest) -> Unit) {
    var address by remember { mutableStateOf("") }
    var amount by remember { mutableStateOf("") }
    var errors by remember { mutableStateOf<Map<String, String>>(emptyMap()) }
    val validator = remember { TransactionValidator() }

    Column(modifier = Modifier.padding(16.dp)) {
        OutlinedTextField(
            value = address,
            onValueChange = { address = it; errors = errors - "toAddress" },
            label = { Text("Recipient Address") },
            isError = errors.containsKey("toAddress"),
            supportingText = errors["toAddress"]?.let {
                { Text(it, color = MaterialTheme.colorScheme.error) }
            },
            modifier = Modifier.fillMaxWidth()
        )

        Spacer(Modifier.height(8.dp))

        OutlinedTextField(
            value = amount,
            onValueChange = { amount = it; errors = errors - "amount" },
            label = { Text("Amount") },
            isError = errors.containsKey("amount"),
            supportingText = errors["amount"]?.let {
                { Text(it, color = MaterialTheme.colorScheme.error) }
            },
            modifier = Modifier.fillMaxWidth()
        )

        Spacer(Modifier.height(16.dp))

        Button(
            onClick = {
                val request = TransactionRequest(
                    toAddress = address,
                    amount = amount.toBigDecimalOrNull() ?: BigDecimal.ZERO,
                    currency = "USDC"
                )
                when (val result = validator.validate(request)) {
                    is Result.Success -> onConfirm(request)
                    is Result.Error -> {
                        errors = (result.exception
                            as NexusException.ValidationException).errors
                    }
                    else -> {}
                }
            },
            modifier = Modifier.fillMaxWidth()
        ) { Text("Send") }
    }
}
```

---

## 6. WebSocket Errors

```kotlin
class WebSocketErrorHandler(
    private val reconnectManager: ReconnectManager
) {

    fun handleWsError(code: Int, reason: String): NexusException.WebSocketException {
        return when (code) {
            1000 -> NexusException.WebSocketException("Normal closure", code)
            1001 -> NexusException.WebSocketException("Going away", code)
            1006 -> NexusException.WebSocketException("Abnormal closure", code)
            1008 -> NexusException.WebSocketException("Policy violation", code)
            1009 -> NexusException.WebSocketException("Message too large", code)
            1011 -> NexusException.WebSocketException("Server error", code)
            1012 -> NexusException.WebSocketException("Server restarting", code)
            4001 -> NexusException.AuthException(authCode = AuthErrorCode.TOKEN_EXPIRED)
            4003 -> NexusException.AuthException("Unauthorized")
            4008 -> NexusException.RateLimitException(5000)
            else -> NexusException.WebSocketException("WS error: $reason", code)
        }
    }

    suspend fun onError(code: Int, reason: String) {
        val exception = handleWsError(code, reason)
        Timber.w("WebSocket error: code=$code reason=$reason")

        when {
            exception is NexusException.AuthException -> {
                reconnectManager.stopReconnect()
            }
            exception.recoverable -> {
                reconnectManager.scheduleReconnect()
            }
            else -> {
                reconnectManager.stopReconnect()
            }
        }
    }
}
```

---

## 7. Error UI Patterns

### 7.1 Snackbar

```kotlin
@Composable
fun ErrorSnackbar(
    error: NexusException?,
    onDismiss: () -> Unit,
    onRetry: (() -> Unit)? = null,
    snackbarHostState: SnackbarHostState = remember { SnackbarHostState() }
) {
    LaunchedEffect(error) {
        error?.let {
            val result = snackbarHostState.showSnackbar(
                message = it.userMessage,
                duration = SnackbarDuration.Short,
                actionLabel = if (it.recoverable && onRetry != null) "Retry" else null
            )
            when (result) {
                SnackbarResult.ActionPerformed -> onRetry?.invoke()
                SnackbarResult.Dismissed -> onDismiss()
            }
        }
    }

    SnackbarHost(hostState = snackbarHostState)
}
```

### 7.2 Error Dialog

```kotlin
@Composable
fun ErrorDialog(
    error: NexusException?,
    onDismiss: () -> Unit,
    onRetry: (() -> Unit)? = null
) {
    if (error == null) return

    AlertDialog(
        onDismissRequest = onDismiss,
        icon = {
            Icon(Icons.Default.Warning, null,
                tint = MaterialTheme.colorScheme.error)
        },
        title = { Text("Error") },
        text = {
            Column {
                Text(error.userMessage)
                if (BuildConfig.DEBUG) {
                    Spacer(Modifier.height(8.dp))
                    Text(
                        "Code: ${error.code}\nMessage: ${error.message}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.outline
                    )
                }
            }
        },
        confirmButton = {
            if (onRetry != null && error.recoverable) {
                TextButton(onClick = onRetry) { Text("Retry") }
            }
            TextButton(onClick = onDismiss) { Text("OK") }
        }
    )
}
```

### 7.3 Inline Error

```kotlin
@Composable
fun InlineFieldError(error: String?, modifier: Modifier = Modifier) {
    AnimatedVisibility(
        visible = error != null,
        enter = fadeIn() + expandVertically(),
        exit = fadeOut() + shrinkVertically()
    ) {
        Row(
            modifier = modifier.padding(top = 4.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(Icons.Default.ErrorOutline, null,
                tint = MaterialTheme.colorScheme.error,
                modifier = Modifier.size(16.dp))
            Spacer(Modifier.width(4.dp))
            Text(error ?: "", color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall)
        }
    }
}
```

### 7.4 Full-Screen Error

```kotlin
@Composable
fun FullScreenError(
    error: NexusException,
    onRetry: (() -> Unit)? = null,
    onReport: (() -> Unit)? = null
) {
    Column(
        modifier = Modifier.fillMaxSize().padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            imageVector = when (error) {
                is NexusException.NetworkException -> Icons.Default.WifiOff
                is NexusException.ServerException -> Icons.Default.CloudOff
                is NexusException.AuthException -> Icons.Default.Lock
                else -> Icons.Default.ErrorOutline
            },
            contentDescription = null,
            modifier = Modifier.size(80.dp),
            tint = MaterialTheme.colorScheme.outline
        )

        Spacer(Modifier.height(16.dp))

        Text(
            text = when (error) {
                is NexusException.NetworkException -> "No Internet Connection"
                is NexusException.ServerException -> "Server Unavailable"
                is NexusException.AuthException -> "Authentication Required"
                else -> "Something Went Wrong"
            },
            style = MaterialTheme.typography.headlineSmall,
            textAlign = TextAlign.Center
        )

        Spacer(Modifier.height(8.dp))
        Text(error.userMessage, style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.outline,
            textAlign = TextAlign.Center)

        Spacer(Modifier.height(24.dp))

        if (onRetry != null && error.recoverable) {
            Button(onClick = onRetry) {
                Icon(Icons.Default.Refresh, null)
                Spacer(Modifier.width(8.dp))
                Text("Try Again")
            }
        }

        if (onReport != null) {
            Spacer(Modifier.height(8.dp))
            TextButton(onClick = onReport) { Text("Report Problem") }
        }
    }
}
```

---

## 8. Empty & Loading States

```kotlin
@Composable
fun EmptyState(
    icon: ImageVector,
    title: String,
    description: String,
    actionLabel: String? = null,
    onAction: (() -> Unit)? = null
) {
    Column(
        modifier = Modifier.fillMaxSize().padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(icon, null, Modifier.size(64.dp),
            tint = MaterialTheme.colorScheme.outline)
        Spacer(Modifier.height(16.dp))
        Text(title, style = MaterialTheme.typography.titleMedium)
        Spacer(Modifier.height(8.dp))
        Text(description, style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.outline,
            textAlign = TextAlign.Center)
        if (actionLabel != null && onAction != null) {
            Spacer(Modifier.height(16.dp))
            Button(onClick = onAction) { Text(actionLabel) }
        }
    }
}

@Composable
fun ShimmerBox(modifier: Modifier = Modifier) {
    val shimmer = rememberInfiniteTransition(label = "shimmer")
    val alpha by shimmer.animateFloat(
        initialValue = 0.2f, targetValue = 0.9f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000),
            repeatMode = RepeatMode.Reverse
        ), label = "shimmer_alpha"
    )
    Box(modifier = modifier.clip(RoundedCornerShape(8.dp))
        .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = alpha)))
}

@Composable
fun LoadingSkeleton() {
    Column(Modifier.padding(16.dp)) {
        repeat(5) {
            ShimmerBox(Modifier.fillMaxWidth().height(60.dp)
                .padding(vertical = 4.dp))
        }
    }
}
```

---

## 9. Retry Logic

```kotlin
class RetryManager(
    private val maxRetries: Int = 3,
    private val initialDelayMs: Long = 1000,
    private val maxDelayMs: Long = 30_000,
    private val factor: Double = 2.0
) {
    suspend fun <T> retryWithBackoff(
        retryable: (NexusException) -> Boolean = { it.recoverable },
        block: suspend () -> T
    ): Result<T> {
        var currentDelay = initialDelayMs
        var lastError: NexusException? = null

        repeat(maxRetries) { attempt ->
            when (val result = runCatchingResult(block)) {
                is Result.Success -> return result
                is Result.Error -> {
                    lastError = result.exception
                    if (!retryable(result.exception) || attempt == maxRetries - 1) {
                        return result
                    }
                    Timber.w("Retry ${attempt + 1}/$maxRetries after ${currentDelay}ms")
                    delay(currentDelay)
                    currentDelay = (currentDelay * factor).toLong()
                        .coerceAtMost(maxDelayMs)
                }
                is Result.Loading -> {}
            }
        }
        return Result.Error(lastError ?: NexusException.UnknownException())
    }
}
```

### Retry Timing

```
Attempt │ Delay      │ Cumulative
────────┼────────────┼───────────
   1    │ 1,000 ms   │ 1,000 ms
   2    │ 2,000 ms   │ 3,000 ms
   3    │ 4,000 ms   │ 7,000 ms
   4    │ 8,000 ms   │ 15,000 ms
   5    │ 16,000 ms  │ 31,000 ms
   6    │ 30,000 ms  │ 61,000 ms  (capped at maxDelay)
```

---

## 10. Error Boundaries

```kotlin
@Composable
fun ErrorBoundary(
    fallback: @Composable (error: Throwable, retry: () -> Unit) -> Unit,
    content: @Composable () -> Unit
) {
    var error by remember { mutableStateOf<Throwable?>(null) }

    if (error != null) {
        fallback(error!!) { error = null }
    } else {
        content()
    }
}

@Composable
fun AppLevelErrorBoundary(content: @Composable () -> Unit) {
    ErrorBoundary(
        fallback = { error, retry ->
            FullScreenError(
                error = error.toNexusException(),
                onRetry = retry,
                onReport = { /* open bug report */ }
            )
        }
    ) {
        content()
    }
}

@Composable
fun FeatureErrorBoundary(
    featureName: String,
    content: @Composable () -> Unit
) {
    ErrorBoundary(
        fallback = { error, retry ->
            Column(
                Modifier.fillMaxWidth().padding(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text("$featureName is unavailable",
                    style = MaterialTheme.typography.titleSmall)
                Spacer(Modifier.height(8.dp))
                Button(onClick = retry) { Text("Try Again") }
            }
        }
    ) {
        content()
    }
}
```

---

## 11. Error Logging & Monitoring

```kotlin
class ErrorLogger(
    private val crashlytics: FirebaseCrashlytics,
    private val analytics: FirebaseAnalytics
) {
    fun logError(error: NexusException, context: Map<String, String> = emptyMap()) {
        when {
            error is NexusException.NetworkException -> Timber.w(error)
            error.recoverable -> Timber.w(error, "Recoverable error")
            else -> Timber.e(error, "Non-recoverable error")
        }

        crashlytics.apply {
            setCustomKey("error_code", error.code)
            setCustomKey("recoverable", error.recoverable)
            context.forEach { (key, value) -> setCustomKey(key, value) }
            recordException(error)
        }

        val bundle = Bundle().apply {
            putString("error_type", error::class.simpleName)
            putString("error_code", error.code)
            putBoolean("recoverable", error.recoverable)
        }
        analytics.logEvent("app_error", bundle)
    }

    fun logBreadcrumb(message: String) {
        Timber.d("Breadcrumb: $message")
        crashlytics.log(message)
    }
}
```

### Error Monitoring Dashboard

```
Error Rate Dashboard:
┌──────────────────────────────────────────────────────────────┐
│  Error Type        │ Count │ % Total │ Trend  │ Severity    │
├──────────────────────────────────────────────────────────────┤
│  Network           │ 1,200 │ 45%     │ ↓ -5%  │ Low         │
│  Auth (token)      │ 800   │ 30%     │ → 0%   │ Medium      │
│  Server 5xx        │ 200   │ 7.5%    │ ↑ +10% │ High        │
│  Validation        │ 300   │ 11%     │ → 0%   │ Low         │
│  WebSocket         │ 100   │ 3.7%    │ ↓ -20% │ Low         │
│  Unknown           │ 60    │ 2.8%    │ ↑ +5%  │ Medium      │
├──────────────────────────────────────────────────────────────┤
│  TOTAL             │ 2,660 │ 100%    │        │             │
│  Crash-free rate   │ 99.7%                                │
│  ANR-free rate     │ 99.95%                               │
└──────────────────────────────────────────────────────────────┘
```

---

## 12. Error Recovery Strategies

```kotlin
class ErrorRecoveryStrategy(
    private val cacheManager: CacheManager,
    private val connectivityChecker: ConnectivityChecker,
    private val featureFlags: FeatureFlags
) {
    suspend fun <T> recover(
        cacheKey: String,
        networkCall: suspend () -> T
    ): Result<T> {
        when (val result = runCatchingResult(networkCall)) {
            is Result.Success -> return result
            is Result.Error -> {
                if (!result.exception.recoverable) return result

                val cached = cacheManager.get<T>(cacheKey)
                if (cached != null) {
                    FirebaseCrashlytics.getInstance()
                        .log("Cache fallback: $cacheKey")
                    return Result.Success(cached)
                }

                return result
            }
            is Result.Loading -> {}
        }
        return Result.Error(NexusException.UnknownException())
    }
}
```

### Degraded Mode

```
Degraded Mode Levels:
┌──────────────────────────────────────────────────────────────┐
│  Level 1 (Degraded):                                        │
│    - WebSocket offline, use polling                          │
│    - Cache-first for all reads                              │
│    - Disable non-essential analytics                        │
│                                                              │
│  Level 2 (Limited):                                         │
│    - Read-only mode (no writes)                             │
│    - Cached content only                                    │
│    - Simplified UI                                          │
│                                                              │
│  Level 3 (Offline):                                         │
│    - Full offline mode                                      │
│    - Queue writes for later sync                            │
│    - Show cached data with "offline" badge                  │
└──────────────────────────────────────────────────────────────┘
```

---

## 13. Error Accessibility

```kotlin
@Composable
fun AccessibleErrorState(error: NexusException, modifier: Modifier = Modifier) {
    val focusRequester = remember { FocusRequester() }

    LaunchedEffect(error) { focusRequester.requestFocus() }

    Box(
        modifier = modifier
            .focusRequester(focusRequester)
            .focusable()
            .semantics {
                contentDescription = "Error: ${error.userMessage}"
                liveRegion = LiveRegionMode.Assertive
            }
    ) {
        Text(error.userMessage, modifier = Modifier.semantics {
            liveRegion = LiveRegionMode.Assertive
        })
    }
}
```

### Accessibility Guidelines for Errors

| Error UI           | TalkBack Announcement          | Focus Behavior            |
|--------------------|--------------------------------|---------------------------|
| Snackbar           | Auto-announced via liveRegion  | No focus change           |
| Dialog             | Title + message announced      | Focus moves to dialog     |
| Inline error       | Announced when appears         | Linked to input field     |
| Full-screen error  | Title + description announced  | Focus moves to retry btn  |
| Connection banner  | Announced once, not repeated   | No focus change           |

---

## 14. Error Testing

```kotlin
class ErrorHandlingTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun should_showNetworkError_when_noInternet() {
        coEvery { connectivityChecker.isConnected() } returns false

        viewModel.onAction(ChatAction.LoadMessages)

        composeTestRule.onNodeWithText("No internet connection")
            .assertIsDisplayed()
        composeTestRule.onNodeWithText("Try Again")
            .assertIsDisplayed()
    }

    @Test
    fun should_showRetryButton_when_serverError() {
        coEvery { apiClient.getMessages() } returns
            Result.failure(NexusException.ServerException("Server error", 500))

        viewModel.onAction(ChatAction.LoadMessages)

        composeTestRule.onNodeWithText(
            "Server error. Please try again later."
        ).assertIsDisplayed()
    }

    @Test
    fun should_showInlineValidation_when_invalidInput() {
        composeTestRule.setContent {
            TransactionForm(onConfirm = {})
        }

        composeTestRule.onNodeWithTag("amount_input").performTextInput("0")
        composeTestRule.onNodeWithTag("confirm_button").performClick()

        composeTestRule.onNodeWithText("Amount must be greater than zero")
            .assertIsDisplayed()
    }

    @Test
    fun should_dismissError_when_tappedDismiss() {
        viewModel.onErrorShown()

        composeTestRule.onNodeWithText("Error message").assertDoesNotExist()
    }

    @Test
    fun should_retry_when_retryButtonClicked() {
        composeTestRule.onNodeWithText("Try Again").performClick()

        coVerify { viewModel.onAction(ChatAction.LoadMessages) }
    }

    @Test
    fun should_showOfflineState_when_networkLost() {
        connectivityChecker.setConnected(false)

        viewModel.onNetworkChanged(false)

        composeTestRule.onNodeWithTag("offline_banner").assertIsDisplayed()
        composeTestRule.onNodeWithText("You are offline").assertIsDisplayed()
    }
}
```

---

## 15. Error Debugging

```
Error Debug Workflow:
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  User        │────►│  Crashlytics │────►│  Log Viewer  │
│  Report      │     │  Dashboard   │     │  (Logcat)    │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                            ▼
                     ┌──────────────┐     ┌──────────────┐
                     │  Symbol      │────►│  Stack Trace │
                     │  Upload      │     │  Resolution  │
                     └──────────────┘     └──────────────┘
```

### Debug Log Format

```kotlin
class DebugErrorFormatter {
    fun format(error: NexusException): String = buildString {
        appendLine("=== ERROR DETAILS ===")
        appendLine("Type: ${error::class.simpleName}")
        appendLine("Code: ${error.code}")
        appendLine("Message: ${error.message}")
        appendLine("Recoverable: ${error.recoverable}")
        appendLine("User Message: ${error.userMessage}")
        appendLine("Stack Trace:")
        error.stackTraceToString().lines().take(10).forEach {
            appendLine("  $it")
        }
        appendLine("=====================")
    }
}
```

---

## 16. Error Metrics

| Metric                  | Target     | Alert Threshold | Action           |
|------------------------|-----------|----------------|------------------|
| Crash-free rate         | >99.5%    | <99%           | Hotfix           |
| ANR-free rate           | >99.9%    | <99.5%         | Investigate      |
| Error rate (total)      | <1%       | >3%            | Alert team       |
| Network error rate      | <0.5%     | >2%             | Check API/CDN    |
| Auth error rate         | <0.3%     | >1%             | Check token flow |
| Server 5xx rate         | <0.1%     | >0.5%           | Escalate to SRE  |
| Validation error rate   | <5%       | >10%            | Review UX/forms  |
| WebSocket disconnects   | <2/hr     | >5/hr           | Check server     |
| Mean time to recovery   | <5 min    | >15 min         | Improve retry    |
| Error resolution time   | <24 hr    | >72 hr          | Escalate         |

### Error Rate Tracking

```kotlin
class ErrorMetricsCollector {

    private val errorCounts = mutableMapOf<String, Int>()

    fun recordError(error: NexusException) {
        val key = error::class.simpleName ?: "Unknown"
        errorCounts[key] = (errorCounts[key] ?: 0) + 1

        Firebase.analytics.logEvent("error_occurred") {
            param("error_type", key)
            param("error_code", error.code)
            param("recoverable", error.recoverable.toString())
        }
    }

    fun getErrorRate(): Double {
        val total = errorCounts.values.sum()
        return if (total > 0) total.toDouble() / totalRequests else 0.0
    }

    fun getTopErrors(limit: Int = 5): List<Pair<String, Int>> {
        return errorCounts.entries
            .sortedByDescending { it.value }
            .take(limit)
            .map { it.key to it.value }
    }
}
```
