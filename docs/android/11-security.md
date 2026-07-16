# 11. Security

## Table of Contents

1. [Biometric Authentication](#1-biometric-authentication)
2. [Biometric Enrollment Flow](#2-biometric-enrollment-flow)
3. [Biometric Login Flow](#3-biometric-login-flow)
4. [Biometric Types](#4-biometric-types)
5. [Biometric Fallback](#5-biometric-fallback)
6. [Biometric Error Handling](#6-biometric-error-handling)
7. [Secure Storage](#7-secure-storage)
8. [Android Keystore](#8-android-keystore)
9. [EncryptedSharedPreferences](#9-encryptedsharedpreferences)
10. [Token Storage Security](#10-token-storage-security)
11. [Certificate Pinning](#11-certificate-pinning)
12. [Network Security Config](#12-network-security-config)
13. [SSL/TLS Configuration](#13-ssltls-configuration)
14. [Device Binding](#14-device-binding)
15. [Root Detection](#15-root-detection)
16. [Debugger Detection](#16-debugger-detection)
17. [App Integrity Verification](#17-app-integrity-verification)
18. [ProGuard/R8 Obfuscation](#18-proguardr8-obfuscation)
19. [Code Obfuscation Rules](#19-code-obfuscation-rules)
20. [Secure Logging](#20-secure-logging)
21. [Secure Network Communication](#21-secure-network-communication)
22. [Secure Deep Links](#22-secure-deep-links)
23. [Input Validation](#23-input-validation)
24. [SQL Injection Prevention](#24-sql-injection-prevention)
25. [XSS Prevention](#25-xss-prevention)
26. [CSRF Protection](#26-csrf-protection)
27. [Session Management](#27-session-management)
28. [Device Security Checks](#28-device-security-checks)
29. [App Data Backup](#29-app-data-backup)
30. [Clipboard Security](#30-clipboard-security)
31. [Screenshot Prevention](#31-screenshot-prevention)
32. [Root Detection Responses](#32-root-detection-responses)
33. [Security Testing](#33-security-testing)
34. [Security Audit Logging](#34-security-audit-logging)

---

## 1. Biometric Authentication

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Biometric Auth Flow                       │
│                                                             │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────┐  │
│  │  App UI  │───▶│BiometricAuth │───▶│  BiometricPrompt │  │
│  │          │    │   Manager    │    │     (System)      │  │
│  └──────────┘    └──────┬───────┘    └────────┬─────────┘  │
│                         │                     │             │
│                  ┌──────▼───────┐    ┌────────▼─────────┐  │
│                  │   Keystore   │    │  Biometric Sensor │  │
│                  │   Manager    │    │  (Fingerprint/    │  │
│                  └──────┬───────┘    │   Face/Iris)      │  │
│                         │            └──────────────────┘  │
│                  ┌──────▼───────┐                           │
│                  │Token Manager │                           │
│                  │(EncryptedSP) │                           │
│                  └──────────────┘                           │
└─────────────────────────────────────────────────────────────┘
```

### Biometric Authentication Manager

```kotlin
import androidx.biometric.BiometricManager
import androidx.biometric.BiometricPrompt
import androidx.core.content.ContextCompat
import androidx.fragment.app.FragmentActivity

class BiometricAuthManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val tokenManager: TokenManager,
    private val keyStoreManager: KeyStoreManager
) {
    private val biometricManager = BiometricManager.from(context)

    sealed class AuthResult {
        data object Success : AuthResult()
        data class Error(val code: Int, val message: String) : AuthResult()
        data object NotAvailable : AuthResult()
        data object NotEnrolled : AuthResult()
        data object UserCancelled : AuthResult()
    }

    fun canAuthenticate(): BiometricAvailability {
        return when (biometricManager.canAuthenticate(
            BiometricManager.Authenticators.BIOMETRIC_STRONG or
            BiometricManager.Authenticators.DEVICE_CREDENTIAL
        )) {
            BiometricManager.BIOMETRIC_SUCCESS -> BiometricAvailability.Available
            BiometricManager.BIOMETRIC_ERROR_NO_HARDWARE -> BiometricAvailability.NoHardware
            BiometricManager.BIOMETRIC_ERROR_HW_UNAVAILABLE -> BiometricAvailability.HardwareUnavailable
            BiometricManager.BIOMETRIC_ERROR_NONE_ENROLLED -> BiometricAvailability.NoneEnrolled
            BiometricManager.BIOMETRIC_ERROR_SECURITY_UPDATE_REQUIRED ->
                BiometricAvailability.SecurityUpdateRequired
            BiometricManager.BIOMETRIC_ERROR_UNSUPPORTED -> BiometricAvailability.Unsupported
            BiometricManager.BIOMETRIC_STATUS_UNKNOWN -> BiometricAvailability.Unknown
            else -> BiometricAvailability.Unknown
        }
    }

    fun authenticate(
        activity: FragmentActivity,
        title: String = "Authenticate",
        subtitle: String = "Verify your identity to continue",
        negativeButtonText: String = "Use Password",
        allowDeviceCredential: Boolean = false,
        onResult: (AuthResult) -> Unit
    ) {
        val executor = ContextCompat.getMainExecutor(context)

        val callback = object : BiometricPrompt.AuthenticationCallback() {
            override fun onAuthenticationSucceeded(result: BiometricPrompt.AuthenticationResult) {
                super.onAuthenticationSucceeded(result)
                onResult(AuthResult.Success)
            }

            override fun onAuthenticationError(errorCode: Int, errString: CharSequence) {
                super.onAuthenticationError(errorCode, errString)
                when (errorCode) {
                    BiometricPrompt.ERROR_NEGATIVE_BUTTON,
                    BiometricPrompt.ERROR_USER_CANCELED -> {
                        onResult(AuthResult.UserCancelled)
                    }
                    BiometricPrompt.ERROR_LOCKOUT -> {
                        onResult(AuthResult.Error(errorCode, "Too many attempts. Device locked."))
                    }
                    BiometricPrompt.ERROR_LOCKOUT_PERMANENT -> {
                        onResult(AuthResult.Error(errorCode, "Biometric locked permanently. Use device credentials."))
                    }
                    else -> {
                        onResult(AuthResult.Error(errorCode, errString.toString()))
                    }
                }
            }

            override fun onAuthenticationFailed() {
                super.onAuthenticationFailed()
                // Called when a biometric is recognized but not accepted
                // (e.g., wrong finger). Do not dismiss the prompt.
            }
        }

        val promptInfoBuilder = BiometricPrompt.PromptInfo.Builder()
            .setTitle(title)
            .setSubtitle(subtitle)

        if (allowDeviceCredential) {
            promptInfoBuilder.setAllowedAuthenticators(
                BiometricManager.Authenticators.BIOMETRIC_STRONG or
                BiometricManager.Authenticators.DEVICE_CREDENTIAL
            )
        } else {
            promptInfoBuilder
                .setNegativeButtonText(negativeButtonText)
                .setAllowedAuthenticators(BiometricManager.Authenticators.BIOMETRIC_STRONG)
        }

        val promptInfo = promptInfoBuilder.build()
        val biometricPrompt = BiometricPrompt(activity, executor, callback)

        // Optional: CryptoObject for added security
        val cryptoObject = createCryptoObject()

        if (cryptoObject != null) {
            biometricPrompt.authenticate(promptInfo, cryptoObject)
        } else {
            biometricPrompt.authenticate(promptInfo)
        }
    }

    private fun createCryptoObject(): BiometricPrompt.CryptoObject? {
        return try {
            val key = keyStoreManager.getOrCreateBiometricKey()
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")
            cipher.init(Cipher.ENCRYPT_MODE, key)
            BiometricPrompt.CryptoObject(cipher)
        } catch (e: Exception) {
            null
        }
    }
}

enum class BiometricAvailability {
    Available,
    NoHardware,
    HardwareUnavailable,
    NoneEnrolled,
    SecurityUpdateRequired,
    Unsupported,
    Unknown
}
```

---

## 2. Biometric Enrollment Flow

### Enrollment Process Diagram

```
┌─────────────────────────────────────────────────────────┐
│              Biometric Enrollment Flow                   │
│                                                         │
│  ┌──────────┐                                           │
│  │ User taps│                                           │
│  │ "Enable  │                                           │
│  │Biometric"│                                           │
│  └────┬─────┘                                           │
│       │                                                 │
│       ▼                                                 │
│  ┌──────────────────┐                                   │
│  │ Check             │                                  │
│  │ Availability      │                                  │
│  └──────┬───────────┘                                   │
│    Available │ Not Available                             │
│    │         │                                          │
│    ▼         ▼                                          │
│  ┌────────┐ ┌────────────────┐                          │
│  │Prompt  │ │Show message:   │                          │
│  │Biometric│ │"Biometric not  │                          │
│  │Prompt  │ │ available"     │                          │
│  └───┬────┘ └────────────────┘                          │
│      │                                                  │
│  ┌───▼────────────┐                                     │
│  │ User authenticates                                   │
│  └───┬────────────┘                                     │
│      │                                                  │
│  ┌───▼────────────┐     ┌──────────────────┐           │
│  │ Generate        │────▶│ Store key in     │           │
│  │ CryptoObject    │     │ Android Keystore │           │
│  │ Key             │     └──────────────────┘           │
│  └───┬────────────┘                                     │
│      │                                                  │
│  ┌───▼────────────┐                                     │
│  │ Save biometric  │                                    │
│  │ enabled flag    │                                    │
│  │ in EncryptedSP  │                                    │
│  └───┬────────────┘                                     │
│      │                                                  │
│  ┌───▼────────────┐                                     │
│  │ Show success    │                                    │
│  │ message         │                                    │
│  └────────────────┘                                     │
└─────────────────────────────────────────────────────────┘
```

### Enrollment Implementation

```kotlin
class BiometricEnrollmentManager @Inject constructor(
    private val biometricAuthManager: BiometricAuthManager,
    private val keyStoreManager: KeyStoreManager,
    private val tokenManager: TokenManager,
    private val securePrefs: EncryptedSharedPreferences
) {
    sealed class EnrollmentState {
        data object Idle : EnrollmentState()
        data object CheckingAvailability : EnrollmentState()
        data object Prompting : EnrollmentState()
        data object Enabling : EnrollmentState()
        data object Success : EnrollmentState()
        data class Error(val message: String) : EnrollmentState()
    }

    private val _state = MutableStateFlow<EnrollmentState>(EnrollmentState.Idle)
    val state: StateFlow<EnrollmentState> = _state.asStateFlow()

    fun isBiometricEnabled(): Boolean {
        return securePrefs.getBoolean("biometric_enabled", false)
    }

    suspend fun enroll(activity: FragmentActivity) {
        _state.value = EnrollmentState.CheckingAvailability

        when (val availability = biometricAuthManager.canAuthenticate()) {
            BiometricAvailability.Available -> {
                _state.value = EnrollmentState.Prompting
                biometricAuthManager.authenticate(
                    activity = activity,
                    title = "Enable Biometric Login",
                    subtitle = "Verify your identity to enable biometric authentication",
                    negativeButtonText = "Cancel",
                    onResult = { result ->
                        handleAuthResult(result)
                    }
                )
            }
            BiometricAvailability.NoneEnrolled -> {
                _state.value = EnrollmentState.Error(
                    "Please set up a fingerprint or face recognition in your device settings first."
                )
            }
            else -> {
                _state.value = EnrollmentState.Error(
                    "Biometric authentication is not available: $availability"
                )
            }
        }
    }

    private fun handleAuthResult(result: BiometricAuthManager.AuthResult) {
        when (result) {
            is BiometricAuthManager.AuthResult.Success -> {
                CoroutineScope(Dispatchers.IO).launch {
                    try {
                        _state.value = EnrollmentState.Enabling

                        // Generate and store biometric-bound key
                        keyStoreManager.generateBiometricKey()

                        // Mark biometric as enabled
                        securePrefs.edit().putBoolean("biometric_enabled", true).apply()

                        _state.value = EnrollmentState.Success
                    } catch (e: Exception) {
                        _state.value = EnrollmentState.Error(
                            "Failed to enable biometric: ${e.message}"
                        )
                    }
                }
            }
            is BiometricAuthManager.AuthResult.UserCancelled -> {
                _state.value = EnrollmentState.Idle
            }
            is BiometricAuthManager.AuthResult.Error -> {
                _state.value = EnrollmentState.Error(result.message)
            }
            else -> {
                _state.value = EnrollmentState.Idle
            }
        }
    }

    fun disable() {
        securePrefs.edit().putBoolean("biometric_enabled", false).apply()
        keyStoreManager.deleteBiometricKey()
        _state.value = EnrollmentState.Idle
    }
}
```

---

## 3. Biometric Login Flow

### Login Flow Diagram

```
┌─────────────────────────────────────────────────────────┐
│               Biometric Login Flow                       │
│                                                         │
│  ┌──────────┐                                           │
│  │ App      │                                           │
│  │ Launches │                                           │
│  └────┬─────┘                                           │
│       │                                                 │
│       ▼                                                 │
│  ┌──────────────────┐                                   │
│  │ Biometric         │                                  │
│  │ enabled?          │                                  │
│  └──────┬───────────┘                                   │
│    Yes   │  No                                           │
│    │     │                                              │
│    ▼     ▼                                              │
│  ┌──────┐ ┌──────────┐                                  │
│  │Prompt│ │Navigate  │                                  │
│  │Bio   │ │to Login  │                                  │
│  └──┬───┘ └──────────┘                                  │
│     │                                                   │
│  ┌──▼─────────────┐                                     │
│  │ Verify with     │                                    │
│  │ CryptoObject    │                                    │
│  └──┬──────────────┘                                    │
│  Success │ Failure                                       │
│  │       │                                              │
│  ▼       ▼                                              │
│  ┌──────┐ ┌──────────┐                                  │
│  │Decrypt│ │Show error│                                 │
│  │token  │ │or fall   │                                 │
│  └──┬───┘ │back to   │                                  │
│     │     │password  │                                  │
│     │     └──────────┘                                  │
│     ▼                                                   │
│  ┌──────────┐                                           │
│  │ Check    │                                           │
│  │ token    │                                           │
│  │ expiry   │                                           │
│  └────┬─────┘                                           │
│  Valid │ Expired                                         │
│  │     │                                                │
│  ▼     ▼                                                │
│  ┌──────┐ ┌──────────┐                                  │
│  │Navigate│ │Refresh  │                                  │
│  │to Home │ │token    │                                  │
│  └────────┘ └─────────┘                                  │
└─────────────────────────────────────────────────────────┘
```

### Login Implementation

```kotlin
class BiometricLoginManager @Inject constructor(
    private val biometricAuthManager: BiometricAuthManager,
    private val tokenManager: TokenManager,
    private val keyStoreManager: KeyStoreManager,
    private val securePrefs: EncryptedSharedPreferences,
    private val apiService: NexusApiService
) {
    sealed class LoginState {
        data object Initial : LoginState()
        data object BiometricPrompted : LoginState()
        data object Authenticating : LoginState()
        data object RefreshingToken : LoginState()
        data object Authenticated : LoginState()
        data class Failed(val reason: String) : LoginState()
    }

    private val _loginState = MutableStateFlow<LoginState>(LoginState.Initial)
    val loginState: StateFlow<LoginState> = _loginState.asStateFlow()

    fun isBiometricLoginEnabled(): Boolean {
        return securePrefs.getBoolean("biometric_enabled", false)
    }

    fun hasStoredToken(): Boolean {
        return tokenManager.getAccessToken() != null
    }

    fun promptBiometricLogin(activity: FragmentActivity) {
        if (!isBiometricLoginEnabled()) {
            _loginState.value = LoginState.Failed("Biometric login not enabled")
            return
        }

        _loginState.value = LoginState.BiometricPrompted

        biometricAuthManager.authenticate(
            activity = activity,
            title = "Welcome Back",
            subtitle = "Authenticate to continue to Nexus AI",
            negativeButtonText = "Use Password",
            onResult = { result ->
                when (result) {
                    is BiometricAuthManager.AuthResult.Success -> {
                        handleBiometricSuccess()
                    }
                    is BiometricAuthManager.AuthResult.UserCancelled -> {
                        _loginState.value = LoginState.Failed("Cancelled")
                    }
                    is BiometricAuthManager.AuthResult.Error -> {
                        _loginState.value = LoginState.Failed(result.message)
                    }
                    else -> {
                        _loginState.value = LoginState.Failed("Authentication failed")
                    }
                }
            }
        )
    }

    private fun handleBiometricSuccess() {
        CoroutineScope(Dispatchers.IO).launch {
            _loginState.value = LoginState.Authenticating

            try {
                // Verify CryptoObject and decrypt token
                val cryptoResult = keyStoreManager.verifyBiometricCrypto()
                if (!cryptoResult) {
                    _loginState.value = LoginState.Failed("Crypto verification failed")
                    return@launch
                }

                // Check token validity
                val accessToken = tokenManager.getAccessToken()

                if (accessToken != null && !tokenManager.isTokenExpired()) {
                    _loginState.value = LoginState.Authenticated
                    return@launch
                }

                // Token expired, try refresh
                if (tokenManager.isTokenExpiringSoon()) {
                    _loginState.value = LoginState.RefreshingToken
                    val refreshToken = tokenManager.getRefreshToken()

                    if (refreshToken != null) {
                        try {
                            val response = apiService.refreshToken(refreshToken)
                            tokenManager.saveTokens(
                                accessToken = response.accessToken,
                                refreshToken = response.refreshToken,
                                expiresIn = response.expiresIn
                            )
                            _loginState.value = LoginState.Authenticated
                        } catch (e: Exception) {
                            // Refresh failed, need full re-login
                            tokenManager.clearTokens()
                            _loginState.value = LoginState.Failed("Session expired. Please login again.")
                        }
                    } else {
                        _loginState.value = LoginState.Failed("No refresh token available")
                    }
                } else {
                    _loginState.value = LoginState.Failed("No valid session")
                }
            } catch (e: Exception) {
                _loginState.value = LoginState.Failed("Authentication error: ${e.message}")
            }
        }
    }

    fun logout() {
        tokenManager.clearTokens()
        keyStoreManager.deleteBiometricKey()
        securePrefs.edit().putBoolean("biometric_enabled", false).apply()
        _loginState.value = LoginState.Initial
    }
}
```

---

## 4. Biometric Types

### Biometric Support Matrix

| Biometric Type | Hardware Required | Security Level | Crypto Support | API Level |
|---------------|-------------------|----------------|----------------|-----------|
| Fingerprint | Fingerprint sensor | `BIOMETRIC_STRONG` | Yes | 23+ |
| Face | Face recognition camera | `BIOMETRIC_STRONG` | Yes | 29+ |
| Iris | Iris scanner | `BIOMETRIC_STRONG` | Yes | 29+ |
| Face (Weak) | Front camera | `BIOMETRIC_WEAK` | No | 29+ |
| Device Credential | PIN/Pattern/Password | `DEVICE_CREDENTIAL` | No | 23+ |

### Biometric Type Detection

```kotlin
class BiometricTypeDetector @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun getAvailableBiometrics(): List<BiometricType> {
        val manager = BiometricManager.from(context)
        val available = mutableListOf<BiometricType>()

        // Check strong biometric
        if (manager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)
            == BiometricManager.BIOMETRIC_SUCCESS
        ) {
            available.add(BiometricType.STRONG)
        }

        // Check weak biometric (face unlock without liveness)
        if (manager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_WEAK)
            == BiometricManager.BIOMETRIC_SUCCESS
        ) {
            available.add(BiometricType.WEAK)
        }

        // Check device credential
        if (manager.canAuthenticate(BiometricManager.Authenticators.DEVICE_CREDENTIAL)
            == BiometricManager.BIOMETRIC_SUCCESS
        ) {
            available.add(BiometricType.DEVICE_CREDENTIAL)
        }

        // Detect specific hardware
        val packageManager = context.packageManager
        if (packageManager.hasSystemFeature(PackageManager.FEATURE_FINGERPRINT)) {
            available.add(BiometricType.FINGERPRINT)
        }

        return available
    }

    fun getRecommendedAuthenticator(): Int {
        val available = getAvailableBiometrics()
        return when {
            available.contains(BiometricType.FINGERPRINT) ->
                BiometricManager.Authenticators.BIOMETRIC_STRONG
            available.contains(BiometricType.STRONG) ->
                BiometricManager.Authenticators.BIOMETRIC_STRONG
            available.contains(BiometricType.WEAK) ->
                BiometricManager.Authenticators.BIOMETRIC_WEAK
            else ->
                BiometricManager.Authenticators.DEVICE_CREDENTIAL
        }
    }
}

enum class BiometricType {
    STRONG,
    WEAK,
    FINGERPRINT,
    DEVICE_CREDENTIAL
}
```

---

## 5. Biometric Fallback

### Fallback Chain

```
┌─────────────────────────────────────────────────────────┐
│                Authentication Fallback                    │
│                                                         │
│  ┌──────────────────┐                                   │
│  │ Primary:         │                                   │
│  │ Biometric (Strong)│                                  │
│  └──────┬───────────┘                                   │
│    Not   │  Success                                      │
│    avail │    │                                          │
│    │     │    ▼                                          │
│    ▼     │  ┌────────────┐                              │
│  ┌───────┐│  │ Authenticated│                            │
│  │Fallback││  └────────────┘                            │
│  │to:    ││                                              │
│  │Device ││                                              │
│  │Cred   ││                                              │
│  └──┬────┘│                                              │
│     │     │                                              │
│  Not avail                                               │
│     │                                                    │
│     ▼                                                    │
│  ┌──────────────┐                                        │
│  │ Fallback to:  │                                       │
│  │ Password Login│                                       │
│  └──────┬───────┘                                        │
│         │                                                │
│         ▼                                                │
│  ┌──────────────┐                                        │
│  │ Email/Password│                                       │
│  │ Login Screen  │                                       │
│  └──────────────┘                                        │
└─────────────────────────────────────────────────────────┘
```

### Fallback Implementation

```kotlin
class AuthFallbackManager @Inject constructor(
    private val biometricAuthManager: BiometricAuthManager,
    private val loginApiService: LoginApiService,
    private val securePrefs: EncryptedSharedPreferences
) {
    sealed class AuthMethod {
        data object Biometric : AuthMethod()
        data object DeviceCredential : AuthMethod()
        data object Password : AuthMethod()
    }

    fun determineAuthMethod(): AuthMethod {
        val biometricAvailability = biometricAuthManager.canAuthenticate()

        return when {
            biometricAvailability == BiometricAvailability.Available &&
            securePrefs.getBoolean("biometric_enabled", false) -> {
                AuthMethod.Biometric
            }
            else -> AuthMethod.Password
        }
    }

    suspend fun loginWithPassword(
        email: String,
        password: String
    ): Result<AuthToken> {
        return try {
            val response = loginApiService.login(LoginRequest(email, password))
            Result.success(response)
        } catch (e: HttpException) {
            Result.failure(Exception("Invalid credentials"))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

data class LoginRequest(
    val email: String,
    val password: String
)

data class AuthToken(
    val accessToken: String,
    val refreshToken: String,
    val expiresIn: Long
)
```

---

## 6. Biometric Error Handling

### Error Codes and Responses

| Error Code | Constant | User Message | Action |
|-----------|----------|--------------|--------|
| 1 | `ERROR_HW_UNAVAILABLE` | Biometric hardware is unavailable | Show password fallback |
| 2 | `ERROR_UNABLE_TO_PROCESS` | Unable to process biometric | Retry or show password |
| 3 | `ERROR_TIMEOUT` | Biometric scan timed out | Auto-retry prompt |
| 4 | `ERROR_NO_SPACE` | Not enough space for biometric | Clear sensor, retry |
| 5 | `ERROR_CANCELED` | Biometric was canceled | Return to previous screen |
| 7 | `ERROR_LOCKOUT` | Too many attempts. Wait 30s | Show countdown, lock |
| 8 | `ERROR_VENDOR` | Vendor-specific error | Show generic error |
| 9 | `ERROR_LOCKOUT_PERMANENT` | Biometric locked permanently | Require password |
| 10 | `ERROR_USER_CANCELED` | User canceled | Return to previous |
| 11 | `ERROR_NEGATIVE_BUTTON` | Negative button pressed | Use password fallback |
| 12 | `ERROR_NEGATIVE_BUTTON` | Password fallback requested | Navigate to password |
| 13 | `ERROR_NO_BIOMETRICS` | No biometrics enrolled | Go to settings |
| 14 | `ERROR_HW_NOT_PRESENT` | No biometric hardware | Use password only |
| 15 | `ERROR_SECURITY_UPDATE_REQUIRED` | Security update needed | Prompt update |
| 16 | `ERROR_UNSUPPORTED` | Not supported on device | Use password |
| 17 | `ERROR_STATUS_UNKNOWN` | Unknown status | Retry |

### Error Handler Implementation

```kotlin
class BiometricErrorHandler @Inject constructor(
    @ApplicationContext private val context: Context
) {
    data class ErrorInfo(
        val code: Int,
        val userMessage: String,
        val technicalMessage: String,
        val suggestedAction: SuggestedAction,
        val shouldLog: Boolean = true
    )

    sealed class SuggestedAction {
        data object Retry : SuggestedAction()
        data object UsePassword : SuggestedAction()
        data object GoToSettings : SuggestedAction()
        data object ShowError : SuggestedAction()
        data object WaitAndRetry : SuggestedAction()
        data object RequirePassword : SuggestedAction()
    }

    fun handleError(errorCode: Int, errString: CharSequence): ErrorInfo {
        return when (errorCode) {
            BiometricPrompt.ERROR_HW_UNAVAILABLE -> ErrorInfo(
                code = errorCode,
                userMessage = "Biometric sensor is unavailable. Use your password instead.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.UsePassword
            )

            BiometricPrompt.ERROR_UNABLE_TO_PROCESS -> ErrorInfo(
                code = errorCode,
                userMessage = "Unable to process. Please try again.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.Retry
            )

            BiometricPrompt.ERROR_TIMEOUT -> ErrorInfo(
                code = errorCode,
                userMessage = "Scan timed out. Please try again.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.Retry
            )

            BiometricPrompt.ERROR_LOCKOUT -> ErrorInfo(
                code = errorCode,
                userMessage = "Too many failed attempts. Please wait 30 seconds.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.WaitAndRetry
            )

            BiometricPrompt.ERROR_LOCKOUT_PERMANENT -> ErrorInfo(
                code = errorCode,
                userMessage = "Biometric locked. Please use your device password.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.RequirePassword
            )

            BiometricPrompt.ERROR_USER_CANCELED -> ErrorInfo(
                code = errorCode,
                userMessage = "Authentication canceled.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.ShowError,
                shouldLog = false
            )

            BiometricPrompt.ERROR_NEGATIVE_BUTTON -> ErrorInfo(
                code = errorCode,
                userMessage = "Use your password to login.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.UsePassword,
                shouldLog = false
            )

            BiometricPrompt.ERROR_NO_BIOMETRICS -> ErrorInfo(
                code = errorCode,
                userMessage = "No biometrics enrolled. Set up a fingerprint or face ID in Settings.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.GoToSettings
            )

            BiometricPrompt.ERROR_HW_NOT_PRESENT -> ErrorInfo(
                code = errorCode,
                userMessage = "This device doesn't support biometric authentication.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.UsePassword
            )

            BiometricPrompt.ERROR_SECURITY_UPDATE_REQUIRED -> ErrorInfo(
                code = errorCode,
                userMessage = "A security update is required. Please update your device.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.ShowError
            )

            BiometricPrompt.ERROR_UNSUPPORTED -> ErrorInfo(
                code = errorCode,
                userMessage = "Biometric authentication is not supported on this device.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.UsePassword
            )

            else -> ErrorInfo(
                code = errorCode,
                userMessage = "Authentication failed. Please try again or use your password.",
                technicalMessage = errString.toString(),
                suggestedAction = SuggestedAction.UsePassword
            )
        }
    }
}
```

---

## 7. Secure Storage

### Storage Security Layers

```
┌─────────────────────────────────────────────────────────┐
│               Security Storage Layers                    │
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │ Layer 1: Android Keystore (Hardware-backed)       │  │
│  │ ┌─────────────────────────────────────────────┐   │  │
│  │ │ • AES-256-GCM encryption keys               │   │  │
│  │ │ • Biometric-bound keys                      │   │  │
│  │ │ • Certificate chain                         │   │  │
│  │ │ • Hardware Security Module (HSM)             │   │  │
│  │ └─────────────────────────────────────────────┘   │  │
│  └───────────────────────────────────────────────────┘  │
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │ Layer 2: EncryptedSharedPreferences              │  │
│  │ ┌─────────────────────────────────────────────┐   │  │
│  │ │ • Access tokens                             │   │  │
│  │ │ • Refresh tokens                            │   │  │
│  │ │ • User preferences (sensitive)              │   │  │
│  │ │ • Biometric key reference                   │   │  │
│  │ └─────────────────────────────────────────────┘   │  │
│  └───────────────────────────────────────────────────┘  │
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │ Layer 3: DataStore (Lightweight)                 │  │
│  │ ┌─────────────────────────────────────────────┐   │  │
│  │ │ • User settings (non-sensitive)             │   │  │
│  │ │ • Theme preferences                        │   │  │
│  │ │ • Feature flags                            │   │  │
│  │ └─────────────────────────────────────────────┘   │  │
│  └───────────────────────────────────────────────────┘  │
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │ Layer 4: Room Database (Encrypted)               │  │
│  │ ┌─────────────────────────────────────────────┐   │  │
│  │ │ • Conversations, messages                   │   │  │
│  │ │ • Documents metadata                        │   │  │
│  │ │ • Agent configurations                      │   │  │
│  │ │ • SQLCipher encryption                      │   │  │
│  │ └─────────────────────────────────────────────┘   │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Secure Storage Module

```kotlin
@Module
@InstallIn(SingletonComponent::class)
object SecureStorageModule {

    @Provides
    @Singleton
    fun provideMasterKey(
        @ApplicationContext context: Context
    ): MasterKey {
        return MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .setRequestStrongBoxBacked(true)
            .build()
    }

    @Provides
    @Singleton
    fun provideEncryptedSharedPreferences(
        @ApplicationContext context: Context,
        masterKey: MasterKey
    ): EncryptedSharedPreferences {
        return EncryptedSharedPreferences.create(
            context,
            "nexus_secure_storage",
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        ) as EncryptedSharedPreferences
    }

    @Provides
    @Singleton
    fun provideKeyStoreManager(): KeyStoreManager {
        return KeyStoreManager()
    }

    @Provides
    @Singleton
    fun provideTokenManager(
        encryptedPrefs: EncryptedSharedPreferences,
        keyStoreManager: KeyStoreManager
    ): TokenManager {
        return TokenManager(encryptedPrefs, keyStoreManager)
    }
}
```

---

## 8. Android Keystore

### Keystore Operations

```kotlin
class KeyStoreManager @Inject constructor() {

    companion object {
        private const val KEYSTORE_PROVIDER = "AndroidKeyStore"
        private const val KEY_ALIAS_PREFIX = "nexus_ai_"
    }

    // ─── Key Generation ────────────────────────────────────

    fun generateSecretKey(alias: String): SecretKey {
        val keyGenerator = KeyGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_AES,
            KEYSTORE_PROVIDER
        )
        val spec = KeyGenParameterSpec.Builder(
            "$KEY_ALIAS_PREFIX$alias",
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        )
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            .setKeySize(256)
            .setUserAuthenticationRequired(false)
            .setRandomizedEncryptionRequired(true)
            .build()

        keyGenerator.init(spec)
        return keyGenerator.generateKey()
    }

    fun generateBiometricKey(alias: String = "biometric"): SecretKey {
        val keyGenerator = KeyGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_AES,
            KEYSTORE_PROVIDER
        )
        val spec = KeyGenParameterSpec.Builder(
            "$KEY_ALIAS_PREFIX$alias",
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        )
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            .setKeySize(256)
            .setUserAuthenticationRequired(true)
            .setInvalidatedByBiometricEnrollment(true)
            .build()

        keyGenerator.init(spec)
        return keyGenerator.generateKey()
    }

    // ─── Key Retrieval ─────────────────────────────────────

    fun getSecretKey(alias: String): SecretKey? {
        return try {
            val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER).apply { load(null) }
            val entry = keyStore.getEntry(
                "$KEY_ALIAS_PREFIX$alias", null
            ) as? KeyStore.SecretKeyEntry
            entry?.secretKey
        } catch (e: Exception) {
            null
        }
    }

    // ─── Key Storage ───────────────────────────────────────

    fun storeKeyPair(alias: String, keyPair: KeyPair) {
        val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER).apply { load(null) }
        keyStore.setKeyEntry(
            "$KEY_ALIAS_PREFIX$alias",
            keyPair.private,
            null,
            arrayOf(keyPair.certificate)
        )
    }

    // ─── Key Attestation ───────────────────────────────────

    fun getKeyAttestation(): ListCertificateChain? {
        return try {
            val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER).apply { load(null) }
            val key = keyStore.getKey("${KEY_ALIAS_PREFIX}attestation", null)
            val certificateChain = keyStore.getCertificateChain("${KEY_ALIAS_PREFIX}attestation")
            certificateChain?.map { it as X509Certificate }
        } catch (e: Exception) {
            null
        }
    }

    // ─── Key Deletion ──────────────────────────────────────

    fun deleteKey(alias: String) {
        try {
            val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER).apply { load(null) }
            keyStore.deleteEntry("$KEY_ALIAS_PREFIX$alias")
        } catch (e: Exception) {
            // Log securely
        }
    }

    fun deleteAllKeys() {
        try {
            val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER).apply { load(null) }
            keyStore.aliases().toList().forEach { alias ->
                if (alias.startsWith(KEY_ALIAS_PREFIX)) {
                    keyStore.deleteEntry(alias)
                }
            }
        } catch (e: Exception) {
            // Log securely
        }
    }

    // ─── Key Existence Check ───────────────────────────────

    fun keyExists(alias: String): Boolean {
        return try {
            val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER).apply { load(null) }
            keyStore.containsAlias("$KEY_ALIAS_PREFIX$alias")
        } catch (e: Exception) {
            false
        }
    }

    // ─── Encrypt / Decrypt ─────────────────────────────────

    fun encrypt(alias: String, plaintext: String): String {
        val key = getSecretKey(alias) ?: generateSecretKey(alias)
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, key)

        val iv = cipher.iv
        val encrypted = cipher.doFinal(plaintext.toByteArray(Charsets.UTF_8))

        return Base64.encodeToString(iv + encrypted, Base64.NO_WRAP)
    }

    fun decrypt(alias: String, encryptedBase64: String): String {
        val key = getSecretKey(alias) ?: throw IllegalStateException("Key not found: $alias")
        val combined = Base64.decode(encryptedBase64, Base64.NO_WRAP)

        val iv = combined.sliceArray(0 until 12)
        val ciphertext = combined.sliceArray(12 until combined.size)

        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.DECRYPT_MODE, key, GCMParameterSpec(128, iv))

        return String(cipher.doFinal(ciphertext), Charsets.UTF_8)
    }
}
```

---

## 9. EncryptedSharedPreferences

### Configuration

```kotlin
// Already shown in Section 7, but here's the complete setup with error handling

class SecurePreferencesManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val prefs: EncryptedSharedPreferences by lazy {
        try {
            val masterKey = MasterKey.Builder(context)
                .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
                .setRequestStrongBoxBacked(true)
                .build()

            EncryptedSharedPreferences.create(
                context,
                "nexus_secure_prefs",
                masterKey,
                EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
                EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
            ) as EncryptedSharedPreferences
        } catch (e: Exception) {
            // Fallback: recreate if corrupted
            context.deleteSharedPreferences("nexus_secure_prefs")
            val masterKey = MasterKey.Builder(context)
                .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
                .build()

            EncryptedSharedPreferences.create(
                context,
                "nexus_secure_prefs",
                masterKey,
                EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
                EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
            ) as EncryptedSharedPreferences
        }
    }

    // String operations
    fun putString(key: String, value: String) {
        prefs.edit().putString(key, value).apply()
    }

    fun getString(key: String, default: String = ""): String {
        return prefs.getString(key, default) ?: default
    }

    // Boolean operations
    fun putBoolean(key: String, value: Boolean) {
        prefs.edit().putBoolean(key, value).apply()
    }

    fun getBoolean(key: String, default: Boolean = false): Boolean {
        return prefs.getBoolean(key, default)
    }

    // Long operations
    fun putLong(key: String, value: Long) {
        prefs.edit().putLong(key, value).apply()
    }

    fun getLong(key: String, default: Long = 0L): Long {
        return prefs.getLong(key, default)
    }

    // Clear specific keys
    fun remove(key: String) {
        prefs.edit().remove(key).apply()
    }

    // Clear all
    fun clearAll() {
        prefs.edit().clear().apply()
    }

    // Check if key exists
    fun contains(key: String): Boolean {
        return prefs.contains(key)
    }
}
```

---

## 10. Token Storage Security

### Token Storage Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Token Storage Architecture                   │
│                                                         │
│  ┌───────────────────────┐  ┌───────────────────────┐  │
│  │    Access Token        │  │    Refresh Token       │  │
│  │    ─────────────       │  │    ──────────────      │  │
│  │    Storage:            │  │    Storage:            │  │
│  │    EncryptedSharedPref │  │    EncryptedSharedPref │  │
│  │    Key: access_token   │  │    Key: refresh_token  │  │
│  │    Encryption: AES256  │  │    Encryption: AES256  │  │
│  │    Expiry: tracked     │  │    Expiry: tracked     │  │
│  │    Auto-refresh: yes   │  │    Rotation: on use    │  │
│  └───────────────────────┘  └───────────────────────┘  │
│                                                         │
│  ┌───────────────────────┐  ┌───────────────────────┐  │
│  │    Biometric Key       │  │    Device ID          │  │
│  │    ──────────────      │  │    ─────────          │  │
│  │    Storage:            │  │    Storage:           │  │
│  │    Android Keystore    │  │    Keystore +         │  │
│  │    Bound: biometric    │  │    EncryptedSharedPref│  │
│  │    Invalidated: on     │  │    Bound: device      │  │
│  │    enrollment change   │  │    hardware           │  │
│  └───────────────────────┘  └───────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Token Security Implementation

```kotlin
class SecureTokenManager @Inject constructor(
    private val encryptedPrefs: EncryptedSharedPreferences,
    private val keyStoreManager: KeyStoreManager
) {
    companion object {
        private const val KEY_ACCESS_TOKEN = "secure_access_token"
        private const val KEY_REFRESH_TOKEN = "secure_refresh_token"
        private const val KEY_ACCESS_TOKEN_EXPIRY = "access_token_expiry"
        private const val KEY_REFRESH_TOKEN_EXPIRY = "refresh_token_expiry"
        private const val KEY_TOKEN_FAMILY = "token_family"
    }

    fun saveTokens(response: AuthTokenResponse) {
        val now = System.currentTimeMillis()

        // Generate or update token family ID for rotation tracking
        val family = encryptedPrefs.getString(KEY_TOKEN_FAMILY, null)
            ?: UUID.randomUUID().toString()

        encryptedPrefs.edit().apply {
            // Encrypt tokens before storage
            putString(KEY_ACCESS_TOKEN, keyStoreManager.encrypt("tokens", response.accessToken))
            putString(KEY_REFRESH_TOKEN, keyStoreManager.encrypt("tokens", response.refreshToken))
            putLong(KEY_ACCESS_TOKEN_EXPIRY, now + (response.expiresIn * 1000))
            putLong(KEY_REFRESH_TOKEN_EXPIRY, now + (response.refreshTokenExpiresIn * 1000))
            putString(KEY_TOKEN_FAMILY, family)
            apply()
        }
    }

    fun getAccessToken(): String? {
        val encrypted = encryptedPrefs.getString(KEY_ACCESS_TOKEN, null) ?: return null
        return try {
            keyStoreManager.decrypt("tokens", encrypted)
        } catch (e: Exception) {
            null
        }
    }

    fun getRefreshToken(): String? {
        val encrypted = encryptedPrefs.getString(KEY_REFRESH_TOKEN, null) ?: return null
        return try {
            keyStoreManager.decrypt("tokens", encrypted)
        } catch (e: Exception) {
            null
        }
    }

    fun isAccessTokenExpired(): Boolean {
        val expiry = encryptedPrefs.getLong(KEY_ACCESS_TOKEN_EXPIRY, 0)
        return System.currentTimeMillis() >= expiry
    }

    fun isRefreshTokenExpired(): Boolean {
        val expiry = encryptedPrefs.getLong(KEY_REFRESH_TOKEN_EXPIRY, 0)
        return System.currentTimeMillis() >= expiry
    }

    fun isAccessTokenExpiringSoon(thresholdMs: Long = 300_000): Boolean {
        val expiry = encryptedPrefs.getLong(KEY_ACCESS_TOKEN_EXPIRY, 0)
        return System.currentTimeMillis() >= (expiry - thresholdMs)
    }

    fun getTokenFamily(): String? {
        return encryptedPrefs.getString(KEY_TOKEN_FAMILY, null)
    }

    fun clearTokens() {
        encryptedPrefs.edit().apply {
            remove(KEY_ACCESS_TOKEN)
            remove(KEY_REFRESH_TOKEN)
            remove(KEY_ACCESS_TOKEN_EXPIRY)
            remove(KEY_REFRESH_TOKEN_EXPIRY)
            remove(KEY_TOKEN_FAMILY)
            apply()
        }
    }

    fun validateTokenIntegrity(): Boolean {
        val accessToken = getAccessToken()
        val refreshToken = getRefreshToken()
        return accessToken != null && refreshToken != null
    }
}

data class AuthTokenResponse(
    val accessToken: String,
    val refreshToken: String,
    val expiresIn: Long,
    val refreshTokenExpiresIn: Long,
    val tokenType: String = "Bearer"
)
```

---

## 11. Certificate Pinning

### OkHttp Certificate Pinner

```kotlin
@Module
@InstallIn(SingletonComponent::class)
object NetworkSecurityModule {

    @Provides
    @Singleton
    fun provideCertificatePinner(): CertificatePinner {
        return CertificatePinner.Builder()
            .add(
                "api.nexus-ai.com",
                "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" // Primary
            )
            .add(
                "api.nexus-ai.com",
                "sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=" // Backup
            )
            .add(
                "*.nexus-ai.com",
                "sha256/CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=" // Wildcard
            )
            .build()
    }

    @Provides
    @Singleton
    fun provideOkHttpClient(
        certificatePinner: CertificatePinner,
        tokenManager: SecureTokenManager,
        @ApplicationContext context: Context
    ): OkHttpClient {
        return OkHttpClient.Builder()
            .certificatePinner(certificatePinner)
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .addInterceptor(AuthInterceptor(tokenManager))
            .addInterceptor(LoggingInterceptor())
            .addNetworkInterceptor(SecurityInterceptor())
            .build()
    }

    @Provides
    @Singleton
    fun provideRetrofit(okHttpClient: OkHttpClient): Retrofit {
        return Retrofit.Builder()
            .baseUrl("https://api.nexus-ai.com/")
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }
}

class AuthInterceptor @Inject constructor(
    private val tokenManager: SecureTokenManager
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()

        // Skip auth for login endpoints
        if (request.url.encodedPath.contains("/auth/")) {
            return chain.proceed(request)
        }

        val token = tokenManager.getAccessToken()
        val authenticatedRequest = if (token != null) {
            request.newBuilder()
                .header("Authorization", "Bearer $token")
                .header("X-Token-Family", tokenManager.getTokenFamily() ?: "")
                .build()
        } else {
            request
        }

        val response = chain.proceed(authenticatedRequest)

        // Handle token refresh on 401
        if (response.code == 401 && !request.url.encodedPath.contains("/auth/")) {
            response.close()
            return handleTokenRefresh(chain, request)
        }

        return response
    }

    private fun handleTokenRefresh(
        chain: Interceptor.Chain,
        originalRequest: Request
    ): Response {
        val refreshToken = tokenManager.getRefreshToken()
            ?: throw IOException("No refresh token")

        // Refresh token request
        val refreshRequest = Request.Builder()
            .url("https://api.nexus-ai.com/auth/refresh")
            .post(
                """{"refresh_token": "$refreshToken"}""".toRequestBody(
                    "application/json".toMediaTypeOrNull()
                )
            )
            .build()

        val refreshResponse = chain.proceed(refreshRequest)

        if (refreshResponse.isSuccessful) {
            val body = refreshResponse.body?.string()
            val tokenResponse = Gson().fromJson(body, AuthTokenResponse::class.java)
            tokenManager.saveTokens(tokenResponse)

            // Retry original request with new token
            val newToken = tokenManager.getAccessToken()
            val retryRequest = originalRequest.newBuilder()
                .header("Authorization", "Bearer $newToken")
                .build()

            return chain.proceed(retryRequest)
        }

        // Refresh failed, clear tokens
        tokenManager.clearTokens()
        throw IOException("Token refresh failed")
    }
}

class SecurityInterceptor @Inject constructor() : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()

        // Ensure HTTPS
        if (!request.url.isHttps) {
            throw IOException("Insecure connection detected")
        }

        // Add security headers
        val secureRequest = request.newBuilder()
            .header("X-Content-Type-Options", "nosniff")
            .header("X-Frame-Options", "DENY")
            .header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
            .build()

        return chain.proceed(secureRequest)
    }
}
```

### Pin Rotation Strategy

```
┌─────────────────────────────────────────────────────────┐
│              Certificate Pin Rotation                     │
│                                                         │
│  Current Pin Set:                                       │
│  ┌─────────┬──────────────────┬──────────────────────┐ │
│  │ Pin ID  │ SHA-256 Hash     │ Status               │ │
│  ├─────────┼──────────────────┼──────────────────────┤ │
│  │ Pin-1   │ AAAA...          │ Active (Primary)     │ │
│  │ Pin-2   │ BBBB...          │ Active (Backup)      │ │
│  │ Pin-3   │ CCCC...          │ Rotating out (90 days)│ │
│  └─────────┴──────────────────┴──────────────────────┘ │
│                                                         │
│  Rotation Process:                                      │
│  1. Generate new certificate with new key               │
│  2. Deploy new cert to server                           │
│  3. Add new pin to app (next release)                   │
│  4. Monitor traffic on old pin                          │
│  5. Remove old pin after migration period               │
│                                                         │
│  Emergency: If pin fails, use backup pins               │
│  or force app update via server-side flag               │
└─────────────────────────────────────────────────────────┘
```

---

## 12. Network Security Config

### network_security_config.xml

```xml
<?xml version="1.0" encoding="utf-8"?>
<network-security-config>

    <!-- Base configuration for all domains -->
    <base-config cleartextTrafficPermitted="false">
        <trust-anchors>
            <certificates src="system" />
        </trust-anchors>
        <pin-set expiration="2025-12-31">
            <pin digest="SHA-256">
                AAAA...
            </pin>
            <pin digest="SHA-256">
                BBBB...
            </pin>
        </pin-set>
    </base-config>

    <!-- API domain with certificate pinning -->
    <domain-config cleartextTrafficPermitted="false">
        <domain includeSubdomains="true">api.nexus-ai.com</domain>
        <pin-set expiration="2025-06-30">
            <pin digest="SHA-256">
                AAAA...
            </pin>
            <pin digest="SHA-256">
                BBBB...
            </pin>
        </pin-set>
        <trust-anchors>
            <certificates src="system" />
        </trust-anchors>
    </domain-config>

    <!-- Development domain (debug builds only) -->
    <domain-config cleartextTrafficPermitted="true">
        <domain includeSubdomains="false">10.0.2.2</domain>
        <domain includeSubdomains="false">localhost</domain>
        <trust-anchors>
            <certificates src="system" />
            <certificates src="user" />
        </trust-anchors>
    </domain-config>

</network-security-config>
```

### AndroidManifest.xml Configuration

```xml
<application
    android:name=".NexusApplication"
    android:allowBackup="false"
    android:fullBackupContent="@xml/backup_rules"
    android:networkSecurityConfig="@xml/network_security_config"
    android:usesCleartextTraffic="false"
    android:supportsRtl="true"
    android:theme="@style/Theme.NexusAI">

    <!-- Disable backup for sensitive data -->
    <meta-data
        android:name="com.google.android.backup.api_key"
        android:value="" />

</application>
```

---

## 13. SSL/TLS Configuration

### TLS Configuration

```kotlin
object TLSConfiguration {
    // Minimum TLS version
    const val MIN_TLS_VERSION = "TLSv1.2"

    // Allowed cipher suites (ordered by preference)
    val ALLOWED_CIPHER_SUITES = arrayOf(
        "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
        "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
        "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
        "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
        "TLS_DHE_RSA_WITH_AES_256_GCM_SHA384",
        "TLS_DHE_RSA_WITH_AES_128_GCM_SHA256"
    )

    // Blocked cipher suites (weak/known vulnerable)
    val BLOCKED_CIPHER_SUITES = arrayOf(
        "TLS_RSA_WITH_AES_128_CBC_SHA",
        "TLS_RSA_WITH_AES_256_CBC_SHA",
        "TLS_RSA_WITH_3DES_EDE_CBC_SHA",
        "TLS_RSA_WITH_RC4_128_SHA",
        "TLS_RSA_WITH_RC4_128_MD5",
        "SSL_RSA_WITH_3DES_EDE_CBC_SHA"
    )
}

class SecureSocketFactory @Inject constructor() {

    fun createSSLSocketFactory(): SSLSocketFactory {
        val context = SSLContext.getInstance("TLSv1.3")
        context.init(null, null, SecureRandom())
        return context.socketFactory
    }

    fun configureConnection(sslSocket: SSLSocket) {
        sslSocket.enabledProtocols = sslSocket.enabledProtocols.filter {
            it.startsWith("TLSv1.2") || it.startsWith("TLSv1.3")
        }.toTypedArray()

        sslSocket.enabledCipherSuites = sslSocket.enabledCipherSuites.filter { suite ->
            TLSConfiguration.ALLOWED_CIPHER_SUITES.contains(suite) &&
            !TLSConfiguration.BLOCKED_CIPHER_SUITES.contains(suite)
        }.toTypedArray()
    }
}
```

---

## 14. Device Binding

### Device Binding Implementation

```kotlin
class DeviceBindingManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val keyStoreManager: KeyStoreManager,
    private val securePrefs: EncryptedSharedPreferences
) {
    data class DeviceInfo(
        val deviceId: String,
        val deviceName: String,
        val manufacturer: String,
        val model: String,
        val osVersion: String,
        val securityPatchLevel: String,
        val isEmulator: Boolean,
        val isRooted: Boolean
    )

    fun getDeviceId(): String {
        var deviceId = securePrefs.getString("device_id", null)
        if (deviceId == null) {
            deviceId = generateDeviceId()
            securePrefs.edit().putString("device_id", deviceId).apply()
        }
        return deviceId
    }

    private fun generateDeviceId(): String {
        val androidId = Settings.Secure.getString(
            context.contentResolver,
            Settings.Secure.ANDROID_ID
        )
        val manufacturer = Build.MANUFACTURER
        val model = Build.MODEL
        val fingerprint = Build.FINGERPRINT

        // Create a stable, unique device ID
        val rawId = "$androidId:$manufacturer:$model:$fingerprint"
        return rawId.sha256Hash()
    }

    fun getDeviceInfo(): DeviceInfo {
        return DeviceInfo(
            deviceId = getDeviceId(),
            deviceName = Build.DEVICE,
            manufacturer = Build.MANUFACTURER,
            model = Build.MODEL,
            osVersion = "Android ${Build.VERSION.RELEASE} (API ${Build.VERSION.SDK_INT})",
            securityPatchLevel = Build.VERSION.SECURITY_PATCH,
            isEmulator = detectEmulator(),
            isRooted = RootDetector.isRooted(context)
        )
    }

    fun bindSession(sessionToken: String): String {
        val deviceId = getDeviceId()
        val bindingData = "$sessionToken:$deviceId:${System.currentTimeMillis()}"
        return bindingData.hmacSha256(keyStoreManager.getDeviceBindingKey())
    }

    fun validateBinding(boundDeviceId: String, currentDeviceId: String): Boolean {
        return boundDeviceId == currentDeviceId
    }

    private fun detectEmulator(): Boolean {
        return (Build.FINGERPRINT.startsWith("generic")
                || Build.FINGERPRINT.startsWith("unknown")
                || Build.MODEL.contains("google_sdk")
                || Build.MODEL.contains("Emulator")
                || Build.MODEL.contains("Android SDK built for x86")
                || Build.MANUFACTURER.contains("Genymotion")
                || Build.BRAND.startsWith("generic") && Build.DEVICE.startsWith("generic")
                || "google_sdk" == Build.PRODUCT)
    }
}

// Extension function for SHA-256 hashing
fun String.sha256Hash(): String {
    val bytes = MessageDigest.getInstance("SHA-256").digest(this.toByteArray())
    return bytes.joinToString("") { "%02x".format(it) }
}

// Extension function for HMAC-SHA256
fun String.hmacSha256(key: SecretKey): String {
    val mac = Mac.getInstance("HmacSHA256")
    mac.init(key)
    val hash = mac.doFinal(this.toByteArray())
    return hash.joinToString("") { "%02x".format(it) }
}
```

---

## 15. Root Detection

### Root Detection Implementation

```kotlin
class RootDetector @Inject constructor(
    @ApplicationContext private val context: Context
) {
    data class RootCheckResult(
        val isRooted: Boolean,
        val confidence: Float, // 0.0 to 1.0
        val detectedMethods: List<RootMethod>
    )

    enum class RootMethod {
        SU_BINARY,
        SU_PATH,
        ROOT_MANAGERS,
        TEST_KEYS,
        DANGEROUS_PROPS,
        rw_SYSTEM,
        BUSYBOX_BINARY,
        XPOSED,
        MAGISK,
        SuperUser_APK,
        ROOT_CERTS
    }

    fun check(): RootCheckResult {
        val detectedMethods = mutableListOf<RootMethod>()

        if (checkSuBinary()) detectedMethods.add(RootMethod.SU_BINARY)
        if (checkSuPath()) detectedMethods.add(RootMethod.SU_PATH)
        if (checkRootManagers()) detectedMethods.add(RootMethod.ROOT_MANAGERS)
        if (checkTestKeys()) detectedMethods.add(RootMethod.TEST_KEYS)
        if (checkDangerousProps()) detectedMethods.add(RootMethod.DANGEROUS_PROPS)
        if (checkRwSystem()) detectedMethods.add(RootMethod.RW_SYSTEM)
        if (checkBusybox()) detectedMethods.add(RootMethod.BUSYBOX_BINARY)
        if (checkXposed()) detectedMethods.add(RootMethod.XPOSED)
        if (checkMagisk()) detectedMethods.add(RootMethod.MAGISK)
        if (checkSuperUserApk()) detectedMethods.add(RootMethod.SUPERUSER_APK)
        if (checkRootCerts()) detectedMethods.add(RootMethod.ROOT_CERTS)

        val confidence = (detectedMethods.size.toFloat() / RootMethod.values().size).coerceIn(0f, 1f)

        return RootCheckResult(
            isRooted = detectedMethods.isNotEmpty(),
            confidence = confidence,
            detectedMethods = detectedMethods
        )
    }

    private fun checkSuBinary(): Boolean {
        val paths = arrayOf(
            "/system/bin/su",
            "/system/xbin/su",
            "/sbin/su",
            "/data/local/xbin/su",
            "/data/local/bin/su",
            "/system/sd/xbin/su",
            "/system/bin/failsafe/su",
            "/data/local/su"
        )
        return paths.any { File(it).exists() }
    }

    private fun checkSuPath(): Boolean {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("which", "su"))
            val reader = BufferedReader(InputStreamReader(process.inputStream))
            reader.readLine() != null
        } catch (e: Exception) {
            false
        }
    }

    private fun checkRootManagers(): Boolean {
        val managers = arrayOf(
            "com.topjohnwu.magisk",
            "eu.chainfire.supersu",
            "com.koushikdutta.superuser",
            "com.thirdparty.superuser",
            "com.noshufou.android.su"
        )
        return managers.any { isPackageInstalled(it) }
    }

    private fun checkTestKeys(): Boolean {
        return Build.TAGS?.contains("test-keys") == true
    }

    private fun checkDangerousProps(): Boolean {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("getprop", "ro.debuggable"))
            val reader = BufferedReader(InputStreamReader(process.inputStream))
            val value = reader.readLine()?.trim()
            value == "1"
        } catch (e: Exception) {
            false
        }
    }

    private fun checkRwSystem(): Boolean {
        return try {
            val mounts = File("/proc/mounts").readText()
            mounts.contains("/system") && !mounts.contains("ro,")
        } catch (e: Exception) {
            false
        }
    }

    private fun checkBusybox(): Boolean {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("which", "busybox"))
            val reader = BufferedReader(InputStreamReader(process.inputStream))
            reader.readLine() != null
        } catch (e: Exception) {
            false
        }
    }

    private fun checkXposed(): Boolean {
        return try {
            val xposedBridge = ClassLoader.getSystemClassLoader()
                .loadClass("de.robv.android.xposed.XposedBridge")
            xposedBridge != null
        } catch (e: ClassNotFoundException) {
            false
        }
    }

    private fun checkMagisk(): Boolean {
        return isPackageInstalled("com.topjohnwu.magisk") ||
            File("/sbin/.magisk").exists() ||
            try {
                val process = Runtime.getRuntime().exec(arrayOf("magisk", "--version"))
                val reader = BufferedReader(InputStreamReader(process.inputStream))
                reader.readLine() != null
            } catch (e: Exception) {
                false
            }
    }

    private fun checkSuperUserApk(): Boolean {
        return isPackageInstalled("com.thirdparty.superuser") ||
            isPackageInstalled("com.koushikdutta.superuser") ||
            isPackageInstalled("com.noshufou.android.su")
    }

    private fun checkRootCerts(): Boolean {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("pm", "list", "packages"))
            val reader = BufferedReader(InputStreamReader(process.inputStream))
            reader.readLines().any { line ->
                line.contains("com.topjohnwu") || line.contains("eu.chainfire")
            }
        } catch (e: Exception) {
            false
        }
    }

    private fun isPackageInstalled(packageName: String): Boolean {
        return try {
            context.packageManager.getPackageInfo(packageName, 0)
            true
        } catch (e: PackageManager.NameNotFoundException) {
            false
        }
    }
}
```

---

## 16. Debugger Detection

### Debugger Detection

```kotlin
class DebuggerDetector @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun isDebuggerAttached(): Boolean {
        return (context.applicationContext as? Application)
            ?.let { isDebuggable(it) || isNativeDebuggerAttached() }
            ?: false
    }

    private fun isDebuggable(application: Application): Boolean {
        val flags = application.applicationInfo.flags
        return (flags and ApplicationInfo.FLAG_DEBUGGABLE) != 0
    }

    private fun isNativeDebuggerAttached(): Boolean {
        return Debug.isDebuggerConnected()
    }

    fun checkAndRespond(): SecurityResponse {
        if (isDebuggerAttached()) {
            // Log security event
            SecurityAuditLogger.logEvent(
                SecurityEvent.DEBUGGER_DETECTED,
                mapOf(
                    "timestamp" to System.currentTimeMillis(),
                    "type" to if (Debug.isDebuggerConnected()) "java" else "build"
                )
            )

            return SecurityResponse.BLOCK
        }
        return SecurityResponse.ALLOW
    }

    fun startDebuggerWatchdog(scope: CoroutineScope) {
        scope.launch {
            while (isActive) {
                if (Debug.isDebuggerConnected()) {
                    SecurityAuditLogger.logEvent(SecurityEvent.DEBUGGER_ATTACHED_RUNTIME)
                    // Take defensive action
                    // Options: exit, limit functionality, show warning
                }
                delay(5000) // Check every 5 seconds
            }
        }
    }
}

enum class SecurityResponse {
    ALLOW,
    WARN,
    BLOCK,
    LIMIT_FEATURES
}
```

---

## 17. App Integrity Verification

### Play Integrity API

```kotlin
class AppIntegrityManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val tokenManager: SecureTokenManager
) {
    private val integrityManager = PlayIntegrityManager(context)

    sealed class IntegrityResult {
        data object Genuine : IntegrityResult()
        data class Compromised(val reason: String) : IntegrityResult()
        data class Unavailable(val reason: String) : IntegrityResult()
        data class Error(val exception: Exception) : IntegrityResult()
    }

    suspend fun verifyIntegrity(): IntegrityResult {
        return try {
            val nonce = generateNonce()
            val integrityTokenRequest = IntegrityTokenRequest.builder()
                .setNonce(nonce)
                .setCloudProjectNumber(CLOUD_PROJECT_NUMBER)
                .build()

            val response = integrityManager.requestIntegrityToken(integrityTokenRequest)

            if (response.token().isNotEmpty()) {
                // Send token to server for verification
                val verificationResult = verifyWithServer(response.token())
                verificationResult
            } else {
                IntegrityResult.Unavailable("Empty integrity token")
            }
        } catch (e: Exception) {
            IntegrityResult.Error(e)
        }
    }

    private fun generateNonce(): String {
        val nonceBytes = ByteArray(16)
        SecureRandom().nextBytes(nonceBytes)
        return Base64.encodeToString(nonceBytes, Base64.URL_SAFE or Base64.NO_WRAP)
    }

    private suspend fun verifyWithServer(token: String): IntegrityResult {
        return try {
            val response = apiService.verifyIntegrity(
                IntegrityRequest(
                    integrityToken = token,
                    deviceId = tokenManager.getDeviceId(),
                    timestamp = System.currentTimeMillis()
                )
            )

            when (response.verdict) {
                "LEGITIMATE" -> IntegrityResult.Genuine
                "TAMPERED" -> IntegrityResult.Compromised("App integrity check failed")
                "UNKNOWN" -> IntegrityResult.Unavailable("Could not verify integrity")
                else -> IntegrityResult.Compromised("Unknown verdict: ${response.verdict}")
            }
        } catch (e: Exception) {
            IntegrityResult.Error(e)
        }
    }

    companion object {
        private const val CLOUD_PROJECT_NUMBER = 123456789L
    }
}

data class IntegrityRequest(
    val integrityToken: String,
    val deviceId: String,
    val timestamp: Long
)

data class IntegrityResponse(
    val verdict: String,
    val requestDetails: String?,
    val error: String?
)
```

---

## 18. ProGuard/R8 Obfuscation

### build.gradle.kts Configuration

```kotlin
android {
    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            isOptimizeResources = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
        debug {
            isMinifyEnabled = false
            isDebuggable = true
        }
    }

    // R8 full mode for better optimization
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}
```

---

## 19. Code Obfuscation Rules

### proguard-rules.pro

```proguard
# ─── General Rules ──────────────────────────────────────────

# Keep annotations
-keepattributes *Annotation*
-keepattributes Signature
-keepattributes InnerClasses
-keepattributes EnclosingMethod

# Keep source file names for crash reports
-keepattributes SourceFile,LineNumberTable
-renamesourcefileattribute SourceFile

# ─── Kotlin Serialization ───────────────────────────────────

-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.AnnotationsKt

-keepclassmembers class kotlinx.serialization.json.** {
    *** Companion;
}
-keepclasseswithmembers class kotlinx.serialization.json.** {
    kotlinx.serialization.KSerializer serializer(...);
}

-keep,includedescriptorclasses class com.nexusai.**$$serializer { *; }
-keepclassmembers class com.nexusai.** {
    *** Companion;
}
-keepclasseswithmembers class com.nexusai.** {
    kotlinx.serialization.KSerializer serializer(...);
}

# ─── Room Database ──────────────────────────────────────────

-keep class * extends androidx.room.RoomDatabase
-keep @androidx.room.Entity class *
-keep @androidx.room.Dao class *
-keep @androidx.room.Embedded class *
-keep @androidx.room.Relation class *
-keep class androidx.room.** { *; }

# Keep Room type converters
-keep class * extends androidx.room.TypeConverter
-keepclassmembers class * {
    @androidx.room.* <methods>;
}

# ─── Hilt / Dependency Injection ────────────────────────────

-keep class dagger.hilt.** { *; }
-keep class javax.inject.** { *; }
-keep class * extends dagger.hilt.android.internal.managers.ViewComponentManager$FragmentContextWrapper { *; }

-keep @dagger.hilt.android.lifecycle.HiltViewModel class * {
    <init>(...);
}

-keep @dagger.hilt.InstallIn class * { *; }
-keep @dagger.Module class * { *; }

# ─── Retrofit / OkHttp ──────────────────────────────────────

-keepattributes Signature, InnerClasses, EnclosingMethod
-keepattributes RuntimeVisibleAnnotations, RuntimeVisibleParameterAnnotations
-keepattributes AnnotationDefault

-keepclassmembers,allowshrinking,allowobfuscation interface * {
    @retrofit2.http.* <methods>;
}

-dontwarn org.codehaus.mojo.animal_sniffer.IgnoreJRERequirement
-dontwarn javax.annotation.**
-dontwarn kotlin.Unit
-dontwarn retrofit2.KotlinExtensions
-dontwarn retrofit2.KotlinExtensions$*

-keep class retrofit2.** { *; }
-keepclasseswithmembers class * {
    @retrofit2.http.* <methods>;
}

-keep class okhttp3.** { *; }
-dontwarn okhttp3.**
-dontwarn okio.**

# ─── Biometric ──────────────────────────────────────────────

-keep class androidx.biometric.** { *; }

# ─── DataStore ──────────────────────────────────────────────

-keep class androidx.datastore.** { *; }
-keepclassmembers class * extends com.google.protobuf.GeneratedMessageLite {
    <fields>;
}

# ─── WorkManager ────────────────────────────────────────────

-keep class * extends androidx.work.Worker
-keep class * extends androidx.work.ListenableWorker
-keepclassmembers class * {
    @androidx.work.* <methods>;
}

# ─── Compose ────────────────────────────────────────────────

-keep class androidx.compose.** { *; }
-dontwarn androidx.compose.**

# ─── Model Classes ──────────────────────────────────────────

-keep class com.nexusai.data.model.** { *; }
-keep class com.nexusai.data.local.entity.** { *; }
-keep class com.nexusai.data.remote.dto.** { *; }

# ─── Enum Classes ───────────────────────────────────────────

-keepclassmembers enum * {
    public static **[] values();
    public static ** valueOf(java.lang.String);
}

# ─── Parcelable ─────────────────────────────────────────────

-keepclassmembers class * implements android.os.Parcelable {
    public static final ** CREATOR;
}

# ─── R8 Full Mode ───────────────────────────────────────────

-allowaccessmodification
-repackageclasses ''
-optimizationpasses 5
```

---

## 20. Secure Logging

### Timber Release Tree

```kotlin
class SecureReleaseTree @Inject constructor() : Timber.Tree() {

    companion object {
        private val SENSITIVE_PATTERNS = listOf(
            Regex("(?i)password"),
            Regex("(?i)token"),
            Regex("(?i)secret"),
            Regex("(?i)api[_-]?key"),
            Regex("(?i)access[_-]?token"),
            Regex("(?i)refresh[_-]?token"),
            Regex("(?i)bearer"),
            Regex("(?i)authorization"),
            Regex("(?i)credit[_-]?card"),
            Regex("(?i)ssn"),
            Regex("(?i)email"),
            Regex("[A-Za-z0-9+/]{40,}"), // Base64 tokens
            Regex("eyJ[A-Za-z0-9-_]+\\.eyJ[A-Za-z0-9-_]+") // JWT tokens
        )

        private val REDACTION_PLACEHOLDERS = mapOf(
            "password" to "[REDACTED:password]",
            "token" to "[REDACTED:token]",
            "secret" to "[REDACTED:secret]",
            "email" to "[REDACTED:email]"
        )
    }

    override fun log(priority: Int, tag: String?, message: String, t: Throwable?) {
        if (BuildConfig.DEBUG) return // Only active in release

        // Don't log sensitive data in release builds
        val sanitizedMessage = sanitizeMessage(message)

        // In production, you might send to a crash reporting service
        // but only sanitized data
        Crashlytics.log(priority, tag ?: "NexusAI", sanitizedMessage)

        if (t != null && priority >= Log.ERROR) {
            Crashlytics.recordException(t)
        }
    }

    private fun sanitizeMessage(message: String): String {
        var sanitized = message
        SENSITIVE_PATTERNS.forEach { pattern ->
            sanitized = pattern.replace(sanitized) { matchResult ->
                val matched = matchResult.value.lowercase()
                REDACTION_PATTERNS.entries.find { (key, _) ->
                    matched.contains(key)
                }?.value ?: "[REDACTED]"
            }
        }
        return sanitized
    }
}

// Application setup
class NexusApplication : Application() {
    override fun onCreate() {
        super.onCreate()

        if (BuildConfig.DEBUG) {
            Timber.plant(Timber.DebugTree())
        } else {
            Timber.plant(SecureReleaseTree())
        }
    }
}

// Usage
class SomeRepository @Inject constructor() {
    fun login(email: String, password: String) {
        // WRONG: Timber.d("Login attempt for $email with password $password")
        // RIGHT:
        Timber.d("Login attempt for user")
    }

    fun saveToken(token: String) {
        // WRONG: Timber.d("Token saved: $token")
        // RIGHT:
        Timber.d("Token saved successfully")
    }
}
```

---

## 21. Secure Network Communication

### HTTPS-Only Enforcement

```kotlin
class SecureNetworkConfig @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun validateUrl(url: String): Boolean {
        val uri = Uri.parse(url)
        return uri.scheme == "https" && !isDevelopmentHost(uri.host)
    }

    private fun isDevelopmentHost(host: String?): Boolean {
        if (!BuildConfig.DEBUG) return false
        val devHosts = listOf("localhost", "10.0.2.2", "127.0.0.1")
        return devHosts.contains(host)
    }

    fun getSecureBaseUrl(): String {
        return if (BuildConfig.DEBUG) {
            "http://10.0.2.2:8080/" // Local dev server
        } else {
            "https://api.nexus-ai.com/"
        }
    }
}

// Network interceptor that blocks HTTP
class HttpBlockerInterceptor @Inject constructor() : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()

        if (request.url.scheme != "https" && !BuildConfig.DEBUG) {
            throw SecurityException(
                "Insecure HTTP request blocked: ${request.url}"
            )
        }

        return chain.proceed(request)
    }
}
```

---

## 22. Secure Deep Links

### Deep Link Security

```kotlin
class SecureDeepLinkHandler @Inject constructor(
    private val signatureVerifier: SignatureVerifier
) {
    sealed class DeepLinkResult {
        data class Valid(val route: String, val params: Map<String, String>) : DeepLinkResult()
        data object Invalid : DeepLinkResult()
        data object Tampered : DeepLinkResult()
    }

    fun handleDeepLink(intent: Intent): DeepLinkResult {
        val data = intent.data ?: return DeepLinkResult.Invalid

        // 1. Verify the scheme
        if (data.scheme != "nexusai") return DeepLinkResult.Invalid

        // 2. Verify the host
        val validHosts = listOf("app", "api", "share")
        if (data.host !in validHosts) return DeepLinkResult.Invalid

        // 3. Verify signature if present
        val signature = data.getQueryParameter("sig")
        val timestamp = data.getQueryParameter("ts")

        if (signature != null && timestamp != null) {
            val payload = "${data.host}:${data.path}:$timestamp"
            if (!signatureVerifier.verify(payload, signature)) {
                return DeepLinkResult.Tampered
            }

            // Check timestamp freshness (5 minutes)
            val ts = timestamp.toLongOrNull() ?: return DeepLinkResult.Invalid
            if (System.currentTimeMillis() - ts > 5 * 60 * 1000) {
                return DeepLinkResult.Invalid
            }
        }

        // 4. Parse route
        return parseRoute(data)
    }

    private fun parseRoute(data: Uri): DeepLinkResult.Valid {
        val path = data.pathSegments
        val params = mutableMapOf<String, String>()

        data.queryParameterNames.forEach { key ->
            data.getQueryParameter(key)?.let { params[key] = it }
        }

        val route = when {
            path.contains("chat") -> "/chat/${path.getOrNull(1) ?: ""}"
            path.contains("agent") -> "/agent/${path.getOrNull(1) ?: ""}"
            path.contains("document") -> "/document/${path.getOrNull(1) ?: ""}"
            path.contains("share") -> "/share"
            else -> "/"
        }

        return DeepLinkResult.Valid(route, params)
    }
}

class SignatureVerifier @Inject constructor(
    private val keyStoreManager: KeyStoreManager
) {
    fun verify(payload: String, signature: String): Boolean {
        return try {
            val expectedSignature = keyStoreManager.sign(payload)
            expectedSignature == signature
        } catch (e: Exception) {
            false
        }
    }
}
```

---

## 23. Input Validation

### Client-Side Validation

```kotlin
class InputValidator @Inject constructor() {

    sealed class ValidationResult {
        data object Valid : ValidationResult()
        data class Invalid(val message: String, val field: String) : ValidationResult()
    }

    fun validateEmail(email: String): ValidationResult {
        if (email.isBlank()) return ValidationResult.Invalid("Email is required", "email")
        if (email.length > 254) return ValidationResult.Invalid("Email is too long", "email")

        val emailRegex = "^[A-Za-z0-9+_.-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$".toRegex()
        if (!emailRegex.matches(email)) {
            return ValidationResult.Invalid("Invalid email format", "email")
        }

        return ValidationResult.Valid
    }

    fun validatePassword(password: String): ValidationResult {
        if (password.length < 8) {
            return ValidationResult.Invalid("Password must be at least 8 characters", "password")
        }
        if (password.length > 128) {
            return ValidationResult.Invalid("Password is too long", "password")
        }
        if (!password.contains(Regex("[A-Z]"))) {
            return ValidationResult.Invalid("Password must contain an uppercase letter", "password")
        }
        if (!password.contains(Regex("[a-z]"))) {
            return ValidationResult.Invalid("Password must contain a lowercase letter", "password")
        }
        if (!password.contains(Regex("[0-9]"))) {
            return ValidationResult.Invalid("Password must contain a number", "password")
        }
        return ValidationResult.Valid
    }

    fun validateSearchQuery(query: String): ValidationResult {
        if (query.length > 200) {
            return ValidationResult.Invalid("Search query is too long", "search")
        }

        // Block SQL injection patterns
        val dangerousPatterns = listOf(
            "'", "\"", ";", "--", "/*", "*/",
            "UNION", "SELECT", "DROP", "DELETE",
            "INSERT", "UPDATE", "ALTER", "EXEC"
        )
        val upperQuery = query.uppercase()
        if (dangerousPatterns.any { upperQuery.contains(it) }) {
            return ValidationResult.Invalid("Invalid characters in search", "search")
        }

        return ValidationResult.Valid
    }

    fun sanitizeInput(input: String): String {
        return input
            .replace("<", "&lt;")
            .replace(">", "&gt;")
            .replace("\"", "&quot;")
            .replace("'", "&#x27;")
            .trim()
            .take(10000) // Limit length
    }
}
```

---

## 24. SQL Injection Prevention

### Room Parameterized Queries

```kotlin
// CORRECT: Parameterized query (safe)
@Dao
interface SafeDao {
    @Query("SELECT * FROM conversations WHERE title = :title")
    suspend fun getByTitle(title: String): ConversationEntity?

    @Query("SELECT * FROM messages WHERE content LIKE '%' || :query || '%'")
    fun search(query: String): Flow<List<MessageEntity>>

    @Query("UPDATE messages SET content = :content WHERE id = :id")
    suspend fun updateContent(id: String, content: String)
}

// NEVER DO: String concatenation (vulnerable)
// @Query("SELECT * FROM conversations WHERE title = '$title'") // NEVER!

// NEVER DO: Raw query with user input (vulnerable)
// database.execSQL("SELECT * FROM conversations WHERE title = '$userInput'") // NEVER!

// If raw query is absolutely necessary, use parameterized form:
fun safeRawQuery(database: SupportSQLiteDatabase, userInput: String) {
    val sanitized = sanitizeForRawQuery(userInput)
    database.execSQL(
        "SELECT * FROM conversations WHERE title = ?",
        arrayOf(sanitized)
    )
}

fun sanitizeForRawQuery(input: String): String {
    // Additional sanitization for edge cases
    return input.replace("'", "''").trim()
}
```

### SQL Injection Prevention Table

| Technique | Safe? | Example |
|-----------|-------|---------|
| Room `@Query` with params | Yes | `@Query("...WHERE id = :id")` |
| Room `@RawQuery` with args | Yes | `database.query(sql, args)` |
| `SupportSQLiteDatabase.execSQL(sql, args)` | Yes | `execSQL("...WHERE id = ?", arrayOf(id))` |
| String concatenation | **NO** | `"WHERE id = '$id'"` |
| `.execSQL(stringInterpolation)` | **NO** | `execSQL("...$input")` |
| Room auto-generated queries | Yes | Insert/Update/Delete |

---

## 25. XSS Prevention

### WebView Security

```kotlin
// PREFERRED: Don't use WebView for user-generated content
// If WebView is absolutely necessary:

@Composable
fun SecureWebView(
    htmlContent: String,
    modifier: Modifier = Modifier
) {
    // Sanitize HTML content
    val sanitized = HtmlSanitizer.sanitize(htmlContent)

    AndroidView(
        factory = { context ->
            WebView(context).apply {
                settings.javaScriptEnabled = false // DISABLE JavaScript
                settings.allowFileAccess = false
                settings.allowContentAccess = false
                settings.domStorageEnabled = false
                settings.databaseEnabled = false
                settings.setSupportMultipleWindows(false)
                settings.javaScriptCanOpenWindowsAutomatically = false
                settings.blockNetworkImage = false
                settings.loadsImagesAutomatically = true

                webViewClient = object : WebViewClient() {
                    override fun shouldOverrideUrlLoading(
                        view: WebView?,
                        request: WebResourceRequest?
                    ): Boolean {
                        // Block all navigation
                        return true
                    }

                    override fun shouldInterceptRequest(
                        view: WebView?,
                        request: WebResourceRequest?
                    ): WebResourceResponse? {
                        // Only allow same-origin requests
                        if (request?.url?.host != "app.nexus-ai.com") {
                            return WebResourceResponse(
                                "text/plain", "utf-8",
                                ByteArrayInputStream("Blocked".toByteArray())
                            )
                        }
                        return super.shouldInterceptRequest(view, request)
                    }
                }

                loadDataWithBaseURL(
                    "https://app.nexus-ai.com",
                    sanitized,
                    "text/html",
                    "UTF-8",
                    null
                )
            }
        },
        modifier = modifier
    )
}

object HtmlSanitizer {
    fun sanitize(html: String): String {
        // Simple sanitization - remove script tags and event handlers
        return html
            .replace(Regex("<script[^>]*>.*?</script>", RegexOption.DOT_MATCHES_ALL), "")
            .replace(Regex("<iframe[^>]*>.*?</iframe>", RegexOption.DOT_MATCHES_ALL), "")
            .replace(Regex("<object[^>]*>.*?</object>", RegexOption.DOT_MATCHES_ALL), "")
            .replace(Regex("<embed[^>]*>.*?</embed>", RegexOption.DOT_MATCHES_ALL), "")
            .replace(Regex("on\\w+\\s*=", RegexOption.IGNORE_CASE), "")
            .replace(Regex("javascript:", RegexOption.IGNORE_CASE), "")
            .replace(Regex("data:", RegexOption.IGNORE_CASE), "")
    }
}
```

---

## 26. CSRF Protection

### Token Rotation

```kotlin
class CSRFProtectionManager @Inject constructor(
    private val securePrefs: EncryptedSharedPreferences
) {
    private val csrfTokenKey = "csrf_token"
    private val csrfTokenExpiryKey = "csrf_token_expiry"

    fun generateCSRFToken(): String {
        val token = ByteArray(32).also { SecureRandom().nextBytes(it) }
        val tokenBase64 = Base64.encodeToString(token, Base64.URL_SAFE or Base64.NO_WRAP)

        securePrefs.edit().apply {
            putString(csrfTokenKey, tokenBase64)
            putLong(csrfTokenExpiryKey, System.currentTimeMillis() + 3_600_000) // 1 hour
            apply()
        }

        return tokenBase64
    }

    fun getCSRFToken(): String? {
        val expiry = securePrefs.getLong(csrfTokenExpiryKey, 0)
        if (System.currentTimeMillis() > expiry) {
            return null // Token expired
        }
        return securePrefs.getString(csrfTokenKey, null)
    }

    fun validateCSRFToken(token: String): Boolean {
        val stored = getCSRFToken() ?: return false
        return MessageDigest.isEqual(
            stored.toByteArray(),
            token.toByteArray()
        )
    }

    fun rotateToken(): String {
        // Generate new token and invalidate old one
        return generateCSRFToken()
    }
}
```

---

## 27. Session Management

### Session Manager

```kotlin
class SessionManager @Inject constructor(
    private val securePrefs: EncryptedSharedPreferences,
    private val tokenManager: SecureTokenManager
) {
    companion object {
        private const val SESSION_TIMEOUT_MS = 30 * 60 * 1000L // 30 minutes
        private const val ABSOLUTE_TIMEOUT_MS = 24 * 60 * 60 * 1000L // 24 hours
        private const val MAX_CONCURRENT_SESSIONS = 3
    }

    data class Session(
        val sessionId: String,
        val userId: String,
        val createdAt: Long,
        val lastActiveAt: Long,
        val deviceId: String,
        val isActive: Boolean
    )

    fun createSession(userId: String, deviceId: String): Session {
        // Check concurrent session limit
        val activeSessions = getActiveSessions()
        if (activeSessions.size >= MAX_CONCURRENT_SESSIONS) {
            // Terminate oldest session
            val oldest = activeSessions.minByOrNull { it.createdAt }
            oldest?.let { terminateSession(it.sessionId) }
        }

        val session = Session(
            sessionId = UUID.randomUUID().toString(),
            userId = userId,
            createdAt = System.currentTimeMillis(),
            lastActiveAt = System.currentTimeMillis(),
            deviceId = deviceId,
            isActive = true
        )

        saveSession(session)
        return session
    }

    fun validateSession(): Boolean {
        val session = getCurrentSession() ?: return false

        if (!session.isActive) return false

        // Check relative timeout (inactivity)
        val inactiveTime = System.currentTimeMillis() - session.lastActiveAt
        if (inactiveTime > SESSION_TIMEOUT_MS) {
            terminateSession(session.sessionId)
            return false
        }

        // Check absolute timeout
        val sessionAge = System.currentTimeMillis() - session.createdAt
        if (sessionAge > ABSOLUTE_TIMEOUT_MS) {
            terminateSession(session.sessionId)
            return false
        }

        // Update last active time
        updateLastActive(session.sessionId)
        return true
    }

    fun terminateSession(sessionId: String) {
        val sessions = getStoredSessions().toMutableList()
        val updated = sessions.map {
            if (it.sessionId == sessionId) it.copy(isActive = false) else it
        }
        saveSessions(updated)

        if (getCurrentSession()?.sessionId == sessionId) {
            tokenManager.clearTokens()
        }
    }

    fun terminateAllSessions() {
        val sessions = getStoredSessions().map { it.copy(isActive = false) }
        saveSessions(sessions)
        tokenManager.clearTokens()
    }

    private fun getCurrentSession(): Session? {
        return getStoredSessions().firstOrNull { it.isActive }
    }

    private fun getActiveSessions(): List<Session> {
        return getStoredSessions().filter { it.isActive }
    }

    private fun saveSession(session: Session) {
        val sessions = getStoredSessions().toMutableList()
        sessions.add(session)
        saveSessions(sessions)
    }

    private fun updateLastActive(sessionId: String) {
        val sessions = getStoredSessions().toMutableList()
        val updated = sessions.map {
            if (it.sessionId == sessionId) it.copy(lastActiveAt = System.currentTimeMillis())
            else it
        }
        saveSessions(updated)
    }

    private fun getStoredSessions(): List<Session> {
        val json = securePrefs.getString("sessions", "[]") ?: "[]"
        return try {
            Gson().fromJson(json, object : TypeToken<List<Session>>() {}.type)
        } catch (e: Exception) {
            emptyList()
        }
    }

    private fun saveSessions(sessions: List<Session>) {
        securePrefs.edit().putString("sessions", Gson().toJson(sessions)).apply()
    }
}
```

---

## 28. Device Security Checks

### Device Security Manager

```kotlin
class DeviceSecurityManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    data class DeviceSecurityStatus(
        val hasScreenLock: Boolean,
        val isOSVersionSupported: Boolean,
        val isBootloaderLocked: Boolean,
        val isEncryptionEnabled: Boolean,
        val securityPatchDate: String,
        val isSecurityPatchCurrent: Boolean,
        val overallScore: SecurityScore
    )

    enum class SecurityScore {
        STRONG,    // All checks pass
        MODERATE,  // Most checks pass
        WEAK,      // Some checks fail
        COMPROMISED // Critical checks fail
    )

    fun checkDeviceSecurity(): DeviceSecurityStatus {
        val hasScreenLock = checkScreenLock()
        val isOSVersionSupported = checkOSVersion()
        val isBootloaderLocked = checkBootloader()
        val isEncryptionEnabled = checkEncryption()
        val securityPatchDate = getSecurityPatchDate()
        val isSecurityPatchCurrent = checkSecurityPatchCurrent(securityPatchDate)

        val score = calculateScore(
            hasScreenLock, isOSVersionSupported,
            isBootloaderLocked, isEncryptionEnabled,
            isSecurityPatchCurrent
        )

        return DeviceSecurityStatus(
            hasScreenLock = hasScreenLock,
            isOSVersionSupported = isOSVersionSupported,
            isBootloaderLocked = isBootloaderLocked,
            isEncryptionEnabled = isEncryptionEnabled,
            securityPatchDate = securityPatchDate,
            isSecurityPatchCurrent = isSecurityPatchCurrent,
            overallScore = score
        )
    }

    private fun checkScreenLock(): Boolean {
        val keyguardManager = context.getSystemService<KeyguardManager>()
        return keyguardManager?.isKeyguardSecure == true
    }

    private fun checkOSVersion(): Boolean {
        return Build.VERSION.SDK_INT >= Build.VERSION_CODES.S // Android 12+
    }

    private fun checkBootloader(): Boolean {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("getprop", "ro.boot.verifiedbootstate"))
            val reader = BufferedReader(InputStreamReader(process.inputStream))
            val state = reader.readLine()?.trim()
            state == "green" || state == "yellow"
        } catch (e: Exception) {
            true // Assume locked if can't check
        }
    }

    private fun checkEncryption(): Boolean {
        return try {
            val ci = CryptoIntents()
            val result = ci.isDeviceEncryptionPolicy()
            result == 1
        } catch (e: Exception) {
            true
        }
    }

    private fun getSecurityPatchDate(): String {
        return Build.VERSION.SECURITY_PATCH ?: "Unknown"
    }

    private fun checkSecurityPatchCurrent(dateString: String): Boolean {
        if (dateString == "Unknown") return false
        return try {
            val patchDate = SimpleDateFormat("yyyy-MM-dd", Locale.US).parse(dateString)
            val threeMonthsAgo = Calendar.getInstance().apply {
                add(Calendar.MONTH, -3)
            }.time
            patchDate?.after(threeMonthsAgo) == true
        } catch (e: Exception) {
            false
        }
    }

    private fun calculateScore(
        screenLock: Boolean,
        osVersion: Boolean,
        bootloaderLocked: Boolean,
        encryption: Boolean,
        patchCurrent: Boolean
    ): SecurityScore {
        val checks = listOf(screenLock, osVersion, bootloaderLocked, encryption, patchCurrent)
        val passCount = checks.count { it }

        return when {
            passCount == 5 -> SecurityScore.STRONG
            passCount >= 3 -> SecurityScore.MODERATE
            passCount >= 1 -> SecurityScore.WEAK
            else -> SecurityScore.COMPROMISED
        }
    }
}
```

---

## 29. App Data Backup

### Backup Rules

```xml
<!-- res/xml/backup_rules.xml (Android 12-) -->
<?xml version="1.0" encoding="utf-8"?>
<full-backup-content>
    <!-- Exclude all sensitive data -->
    <exclude domain="sharedpref" path="nexus_secure_prefs.xml" />
    <exclude domain="sharedpref" path="nexus_secure_storage.xml" />
    <exclude domain="sharedpref" path="db_crypto.xml" />
    <exclude domain="database" path="nexus_ai_database" />
    <exclude domain="database" path="nexus_ai_database-shm" />
    <exclude domain="database" path="nexus_ai_database-wal" />
    <exclude domain="file" path="documents/" />
    <exclude domain="file" path="key_store/" />
    <exclude domain="external" path="." />

    <!-- Include only non-sensitive preferences -->
    <include domain="sharedpref" path="user_settings.xml" />
    <include domain="sharedpref" path="theme_prefs.xml" />
</full-backup-content>

<!-- res/xml/data_extraction_rules.xml (Android 12+) -->
<?xml version="1.0" encoding="utf-8"?>
<data-extraction-rules>
    <cloud-backup>
        <exclude domain="sharedpref" path="nexus_secure_prefs.xml" />
        <exclude domain="sharedpref" path="nexus_secure_storage.xml" />
        <exclude domain="sharedpref" path="db_crypto.xml" />
        <exclude domain="database" path="nexus_ai_database" />
        <exclude domain="file" path="documents/" />
        <exclude domain="file" path="key_store/" />
        <include domain="sharedpref" path="user_settings.xml" />
    </cloud-backup>
    <device-transfer>
        <include domain="sharedpref" path="user_settings.xml" />
    </device-transfer>
</data-extraction-rules>
```

### AndroidManifest.xml Backup Configuration

```xml
<application
    android:allowBackup="false"
    android:fullBackupContent="@xml/backup_rules"
    android:dataExtractionRules="@xml/data_extraction_rules">

    <!-- Disable auto backup for security -->
    <meta-data
        android:name="com.google.android.backup.api_key"
        android:value=""
        tools:node="remove" />

</application>
```

---

## 30. Clipboard Security

### Clipboard Manager

```kotlin
class ClipboardSecurityManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val clipboardManager = context.getSystemService<ClipboardManager>()

    fun copySensitive(text: String, label: String = "Sensitive Data") {
        val clip = ClipData.newPlainText(label, text)
        clipboardManager?.setPrimaryClip(clip)

        // Clear clipboard after delay
        CoroutineScope(Dispatchers.Main).launch {
            delay(30_000) // 30 seconds
            clearClipboard()
        }
    }

    fun clearClipboard() {
        val clip = ClipData.newPlainText("", "")
        clipboardManager?.setPrimaryClip(clip)
    }

    fun isClipboardEmpty(): Boolean {
        return clipboardManager?.primaryClip == null
    }

    fun monitorClipboard(onSensitiveContentDetected: (String) -> Unit) {
        val clipBoardManager = context.getSystemService<ClipboardManager>()
        clipManager?.addPrimaryClipChangedListener {
            val clip = clipManager.primaryClip
            val text = clip?.getItemAt(0)?.text?.toString() ?: return@addPrimaryClipChangedListener

            // Check if copied text might be sensitive
            if (isPotentialSensitiveContent(text)) {
                onSensitiveContentDetected(text)
            }
        }
    }

    private fun isPotentialSensitiveContent(text: String): Boolean {
        val patterns = listOf(
            Regex("\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}\\b"), // Email
            Regex("\\b\\d{4}[- ]?\\d{4}[- ]?\\d{4}[- ]?\\d{4}\\b"), // Credit card
            Regex("eyJ[A-Za-z0-9_-]+\\.eyJ[A-Za-z0-9_-]+"), // JWT
            Regex("\\b[A-Za-z0-9]{32,}\\b") // Long token-like strings
        )
        return patterns.any { it.containsMatchIn(text) }
    }
}
```

---

## 31. Screenshot Prevention

### Screenshot Prevention Implementation

```kotlin
class ScreenshotPreventionManager @Inject constructor(
    private val activity: FragmentActivity
) {
    fun preventScreenshots() {
        activity.window.setFlags(
            WindowManager.LayoutParams.FLAG_SECURE,
            WindowManager.LayoutParams.FLAG_SECURE
        )
    }

    fun allowScreenshots() {
        activity.window.clearFlags(
            WindowManager.LayoutParams.FLAG_SECURE
        )
    }
}

// Composable wrapper
@Composable
fun SecureScreen(
    preventScreenshot: Boolean = true,
    content: @Composable () -> Unit
) {
    val activity = LocalContext.current as? FragmentActivity

    DisposableEffect(activity) {
        if (preventScreenshot && activity != null) {
            activity.window.setFlags(
                WindowManager.LayoutParams.FLAG_SECURE,
                WindowManager.LayoutParams.FLAG_SECURE
            )
        }

        onDispose {
            activity?.window?.clearFlags(
                WindowManager.LayoutParams.FLAG_SECURE
            )
        }
    }

    content()
}

// Usage: Wrap sensitive screens
@Composable
fun LoginScreen() {
    SecureScreen(preventScreenshot = true) {
        // Login UI content
    }
}

@Composable
fun ChatScreen() {
    SecureScreen(preventScreenshot = false) {
        // Chat content (screenshots allowed)
    }
}
```

---

## 32. Root Detection Responses

### Response Strategy

```
┌─────────────────────────────────────────────────────────┐
│           Root Detection Response Matrix                 │
│                                                         │
│  ┌──────────────┬──────────────┬────────────────────┐  │
│  │ Root Level   │ Response     │ Features Affected  │  │
│  ├──────────────┼──────────────┼────────────────────┤  │
│  │ None         │ Allow all    │ None               │  │
│  │ Low (1-2     │ Warn + allow │ No sensitive        │  │
│  │  indicators) │              │ features           │  │
│  │ Medium (3-4) │ Warn + limit │ Limited chat,       │  │
│  │              │              │ no file upload      │  │
│  │ High (5+)    │ Block app    │ App unusable        │  │
│  └──────────────┴──────────────┴────────────────────┘  │
│                                                         │
│  Detection Methods:                                     │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 1. Root binary check (su, busybox)               │  │
│  │ 2. Root app check (Magisk, SuperSU)              │  │
│  │ 3. System property check (test-keys)             │  │
│  │ 4. File system check (/system writable)          │  │
│  │ 5. Xposed framework check                        │  │
│  │ 6. SafetyNet/Play Integrity check                │  │
│  │ 7. Emulator detection                            │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Response Implementation

```kotlin
class RootResponseManager @Inject constructor(
    private val rootDetector: RootDetector,
    private val securityLogger: SecurityAuditLogger
) {
    sealed class SecurityAction {
        data object AllowAll : SecurityAction()
        data object WarnAndAllow : SecurityAction()
        data object WarnAndLimit : SecurityAction()
        data object BlockApp : SecurityAction()
    }

    fun determineAction(): SecurityAction {
        val result = rootDetector.check()

        securityLogger.logEvent(
            SecurityEvent.ROOT_CHECK,
            mapOf(
                "isRooted" to result.isRooted,
                "confidence" to result.confidence,
                "methods" to result.detectedMethods.map { it.name }
            )
        )

        return when {
            !result.isRooted -> SecurityAction.AllowAll
            result.confidence < 0.3f -> SecurityAction.WarnAndAllow
            result.confidence < 0.6f -> SecurityAction.WarnAndLimit
            else -> SecurityAction.BlockApp
        }
    }

    @Composable
    fun RootWarningDialog(
        action: SecurityAction,
        onDismiss: () -> Unit,
        onContinue: () -> Unit
    ) {
        AlertDialog(
            onDismissRequest = onDismiss,
            icon = {
                Icon(
                    Icons.Filled.Warning,
                    contentDescription = "Security Warning",
                    tint = MaterialTheme.colorScheme.error
                )
            },
            title = {
                Text("Security Warning")
            },
            text = {
                Text(
                    when (action) {
                        is SecurityAction.WarnAndAllow ->
                            "This device appears to have been modified. Some features may be limited for your security."
                        is SecurityAction.WarnAndLimit ->
                            "Security risk detected. Chat and document features will be limited."
                        is SecurityAction.BlockApp ->
                            "This device has been significantly modified. The app cannot run on this device for security reasons."
                        else -> ""
                    }
                )
            },
            confirmButton = {
                when (action) {
                    is SecurityAction.WarnAndAllow,
                    is SecurityAction.WarnAndLimit -> {
                        TextButton(onClick = onContinue) {
                            Text("Continue")
                        }
                    }
                    else -> {}
                }
            },
            dismissButton = {
                TextButton(onClick = onDismiss) {
                    Text("Exit")
                }
            }
        )
    }
}
```

---

## 33. Security Testing

### OWASP MASVS Checklist

| Category | Test | Tool | Priority |
|----------|------|------|----------|
| **V1: Architecture** | Threat modeling | Manual | High |
| | Secure coding guidelines | Manual | High |
| **V2: Data Storage** | No sensitive data in logs | Log analysis | Critical |
| | No sensitive data in clipboard | Manual | High |
| | Encrypted storage | Check EncryptedSP | Critical |
| | No backup of sensitive data | adb backup test | High |
| **V3: Cryptography** | Key management in Keystore | Check key usage | Critical |
| | No hardcoded keys | Static analysis | Critical |
| | Strong algorithms (AES-256) | Crypto review | High |
| **V4: Authentication** | Biometric authentication | Manual test | High |
| | Session management | Session testing | High |
| | Token expiry | Token testing | Medium |
| **V5: Network** | HTTPS only | Network config | Critical |
| | Certificate pinning | SSL Labs test | High |
| | No cleartext traffic | Network config | Critical |
| **V6: Platform** | No exported components | Manifest review | High |
| | Secure intent handling | Intent testing | Medium |
| | Input validation | Fuzzing | High |
| **V7: Code** | ProGuard/R8 enabled | Build config | Medium |
| | No debuggable in release | Build config | High |
| | Root detection | Manual test | Medium |
| **V8: Resilience** | Anti-debugging | Debugger test | Medium |
| | Root detection bypass testing | Manual test | Medium |

### Security Scan Integration

```kotlin
class SecurityScanner @Inject constructor(
    @ApplicationContext private val context: Context
) {
    data class ScanResult(
        val vulnerabilities: List<Vulnerability>,
        val score: Int,
        val scanDate: Long
    )

    data class Vulnerability(
        val severity: Severity,
        val category: String,
        val description: String,
        val recommendation: String
    )

    enum class Severity { CRITICAL, HIGH, MEDIUM, LOW, INFO }

    suspend fun performFullScan(): ScanResult {
        val vulnerabilities = mutableListOf<Vulnerability>()

        // 1. Check network security
        vulnerabilities.addAll(checkNetworkSecurity())

        // 2. Check storage security
        vulnerabilities.addAll(checkStorageSecurity())

        // 3. Check code security
        vulnerabilities.addAll(checkCodeSecurity())

        // 4. Check device integrity
        vulnerabilities.addAll(checkDeviceIntegrity())

        val score = calculateScore(vulnerabilities)

        return ScanResult(
            vulnerabilities = vulnerabilities,
            score = score,
            scanDate = System.currentTimeMillis()
        )
    }

    private fun checkNetworkSecurity(): List<Vulnerability> {
        val vulns = mutableListOf<Vulnerability>()

        // Check if cleartext traffic is allowed
        val appInfo = context.applicationInfo
        if ((appInfo.flags and ApplicationInfo.FLAG_DEBUGGABLE) != 0) {
            // Debug builds are exempt
        }

        return vulns
    }

    private fun checkStorageSecurity(): List<Vulnerability> {
        val vulns = mutableListOf<Vulnerability>()

        // Check for SharedPreferences that aren't encrypted
        val sharedPrefsDir = File(context.sharedPreferencesDir)
        sharedPrefsDir.listFiles()?.forEach { file ->
            if (file.name.endsWith(".xml") && !file.name.startsWith("nexus_secure")) {
                vulns.add(Vulnerability(
                    severity = Severity.MEDIUM,
                    category = "Data Storage",
                    description = "Unencrypted SharedPreferences file: ${file.name}",
                    recommendation = "Move sensitive data to EncryptedSharedPreferences"
                ))
            }
        }

        return vulns
    }

    private fun checkCodeSecurity(): List<Vulnerability> {
        val vulns = mutableListOf<Vulnerability>()

        // Check if app is debuggable
        if ((context.applicationInfo.flags and ApplicationInfo.FLAG_DEBUGGABLE) != 0) {
            vulns.add(Vulnerability(
                severity = Severity.HIGH,
                category = "Code Security",
                description = "App is debuggable in current build",
                recommendation = "Disable debuggable for release builds"
            ))
        }

        return vulns
    }

    private fun checkDeviceIntegrity(): List<Vulnerability> {
        val vulns = mutableListOf<Vulnerability>()

        val rootDetector = RootDetector(context)
        val rootResult = rootDetector.check()
        if (rootResult.isRooted) {
            vulns.add(Vulnerability(
                severity = Severity.HIGH,
                category = "Device Integrity",
                description = "Device appears to be rooted",
                recommendation = "Implement root detection responses"
            ))
        }

        return vulns
    }

    private fun calculateScore(vulnerabilities: List<Vulnerability>): Int {
        var score = 100
        vulnerabilities.forEach { vuln ->
            when (vuln.severity) {
                Severity.CRITICAL -> score -= 25
                Severity.HIGH -> score -= 15
                Severity.MEDIUM -> score -= 10
                Severity.LOW -> score -= 5
                Severity.INFO -> score -= 1
            }
        }
        return score.coerceIn(0, 100)
    }
}
```

---

## 34. Security Audit Logging

### Audit Logger

```kotlin
class SecurityAuditLogger @Inject constructor(
    @ApplicationContext private val context: Context,
    private val securePrefs: EncryptedSharedPreferences
) {
    companion object {
        private const val MAX_LOG_ENTRIES = 10000
        private const val LOG_RETENTION_DAYS = 90L
    }

    enum class SecurityEvent {
        // Authentication
        BIOMETRIC_AUTH_ATTEMPT,
        BIOMETRIC_AUTH_SUCCESS,
        BIOMETRIC_AUTH_FAILURE,
        BIOMETRIC_AUTH_LOCKOUT,
        PASSWORD_LOGIN_ATTEMPT,
        PASSWORD_LOGIN_SUCCESS,
        PASSWORD_LOGIN_FAILURE,
        SESSION_CREATED,
        SESSION_EXPIRED,
        SESSION_TERMINATED,
        TOKEN_REFRESH,
        TOKEN_REFRESH_FAILED,

        // Device
        ROOT_DETECTED,
        ROOT_CHECK,
        DEBUGGER_DETECTED,
        DEBUGGER_ATTACHED_RUNTIME,
        EMULATOR_DETECTED,
        DEVICE_SECURITY_CHECK,
        APP_INTEGRITY_CHECK,

        // Data
        DATA_EXPORT_ATTEMPT,
        DATA_PURGE,
        ENCRYPTION_KEY_GENERATED,
        ENCRYPTION_KEY_DELETED,
        SECURE_STORAGE_ACCESS,
        SECURE_STORAGE_CORRUPTED,

        // Network
        CERTIFICATE_PIN_FAILURE,
        INSECURE_REQUEST_BLOCKED,
        API_RATE_LIMITED,
        API_AUTHORIZATION_FAILURE,

        // App
        APP_STARTUP,
        APP_BACKGROUND,
        APP_TAMPER_DETECTED,
        BACKUP_ATTEMPT,
        SCREENSHOT_ATTEMPT
    }

    data class AuditEntry(
        val timestamp: Long,
        val event: SecurityEvent,
        val details: Map<String, Any>,
        val deviceId: String,
        val sessionId: String?,
        val severity: Severity
    ) {
        enum class Severity { INFO, WARNING, ERROR, CRITICAL }
    }

    fun logEvent(
        event: SecurityEvent,
        details: Map<String, Any> = emptyMap(),
        severity: AuditEntry.Severity = AuditEntry.Severity.INFO
    ) {
        val entry = AuditEntry(
            timestamp = System.currentTimeMillis(),
            event = event,
            details = details,
            deviceId = getDeviceId(),
            sessionId = getSessionId(),
            severity = severity
        )

        // Write to local audit log
        writeToLocalLog(entry)

        // Send critical events to server
        if (severity == AuditEntry.Severity.CRITICAL ||
            severity == AuditEntry.Severity.ERROR) {
            sendToServer(entry)
        }

        // Trim old entries
        trimOldEntries()
    }

    private fun writeToLocalLog(entry: AuditEntry) {
        CoroutineScope(Dispatchers.IO).launch {
            try {
                val json = Gson().toJson(entry)
                val logFile = getAuditLogFile()
                logFile.appendText(json + "\n")
            } catch (e: Exception) {
                // Don't let logging failures crash the app
            }
        }
    }

    private suspend fun sendToServer(entry: AuditEntry) {
        try {
            apiService.logSecurityEvent(entry)
        } catch (e: Exception) {
            // Queue for later if offline
        }
    }

    private fun trimOldEntries() {
        CoroutineScope(Dispatchers.IO).launch {
            val cutoff = System.currentTimeMillis() - (LOG_RETENTION_DAYS * 24 * 60 * 60 * 1000)
            val logFile = getAuditLogFile()

            if (logFile.exists()) {
                val entries = logFile.readLines()
                    .mapNotNull { line ->
                        try {
                            Gson().fromJson(line, AuditEntry::class.java)
                        } catch (e: Exception) {
                            null
                        }
                    }
                    .filter { it.timestamp > cutoff }
                    .takeLast(MAX_LOG_ENTRIES)

                logFile.writeText(entries.joinToString("\n") { Gson().toJson(it) })
            }
        }
    }

    private fun getAuditLogFile(): File {
        val dir = File(context.filesDir, "audit_logs")
        dir.mkdirs()
        return File(dir, "security_audit.log")
    }

    private fun getDeviceId(): String {
        return securePrefs.getString("device_id", "unknown") ?: "unknown"
    }

    private fun getSessionId(): String? {
        return securePrefs.getString("current_session_id", null)
    }

    // Query methods
    fun getRecentEvents(
        limit: Int = 100,
        severity: AuditEntry.Severity? = null
    ): List<AuditEntry> {
        return try {
            val logFile = getAuditLogFile()
            if (!logFile.exists()) return emptyList()

            logFile.readLines()
                .mapNotNull { line ->
                    try {
                        Gson().fromJson(line, AuditEntry::class.java)
                    } catch (e: Exception) {
                        null
                    }
                }
                .filter { entry ->
                    severity == null || entry.severity == severity
                }
                .sortedByDescending { it.timestamp }
                .take(limit)
        } catch (e: Exception) {
            emptyList()
        }
    }

    fun getEventsByType(
        event: SecurityEvent,
        since: Long = System.currentTimeMillis() - 24 * 60 * 60 * 1000
    ): List<AuditEntry> {
        return getRecentEvents(Int.MAX_VALUE)
            .filter { it.event == event && it.timestamp > since }
    }
}

// Usage in code
class SomeSecureOperation @Inject constructor(
    private val securityLogger: SecurityAuditLogger
) {
    suspend fun performSecureAction() {
        securityLogger.logEvent(
            SecurityEvent.BIOMETRIC_AUTH_ATTEMPT,
            mapOf("action" to "login")
        )

        try {
            // ... secure operation
            securityLogger.logEvent(
                SecurityEvent.BIOMETRIC_AUTH_SUCCESS,
                mapOf("action" to "login")
            )
        } catch (e: Exception) {
            securityLogger.logEvent(
                SecurityEvent.BIOMETRIC_AUTH_FAILURE,
                mapOf("error" to e.message.orEmpty()),
                AuditEntry.Severity.WARNING
            )
            throw e
        }
    }
}
```

### Audit Log Dashboard

```
┌─────────────────────────────────────────────────────────┐
│           Security Audit Log Summary                     │
│                                                         │
│  Date Range: Last 24 hours                              │
│  ─────────────────────────────────────────────          │
│                                                         │
│  ┌─────────────┬───────────┬──────────────────────┐    │
│  │ Severity    │ Count     │ Trend                │    │
│  ├─────────────┼───────────┼──────────────────────┤    │
│  │ CRITICAL    │ 0         │ -                    │    │
│  │ ERROR       │ 2         │ ▲ +1 from yesterday  │    │
│  │ WARNING     │ 15        │ ▼ -3 from yesterday  │    │
│  │ INFO        │ 342       │ ▼ -20 from yesterday │    │
│  └─────────────┴───────────┴──────────────────────┘    │
│                                                         │
│  Recent Events:                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 14:23  INFO    BIOMETRIC_AUTH_SUCCESS   login    │  │
│  │ 14:20  INFO    APP_STARTUP              normal   │  │
│  │ 14:15  WARNING ROOT_CHECK               rooted   │  │
│  │ 14:10  ERROR   PASSWORD_LOGIN_FAILURE   3 tries  │  │
│  │ 14:05  INFO    SESSION_CREATED          new      │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## Summary

| Security Layer | Technology | Protection |
|---------------|------------|------------|
| **Authentication** | BiometricPrompt + Keystore | Device biometrics |
| **Token Storage** | EncryptedSharedPreferences + Keystore | Encrypted tokens |
| **Key Management** | Android Keystore (HSM) | Hardware-backed keys |
| **Network Security** | Certificate Pinning + TLS 1.2/1.3 | MITM prevention |
| **Data Encryption** | SQLCipher + AES-256-GCM | Data at rest |
| **Code Protection** | R8/ProGuard obfuscation | Reverse engineering |
| **Device Security** | Root + Debugger detection | Tamper detection |
| **App Integrity** | Play Integrity API | APK tampering |
| **Logging** | Timber secure tree | No sensitive data in logs |
| **Backup Security** | allowBackup=false + rules | Data extraction prevention |
| **Screenshot Prevention** | FLAG_SECURE | Screen capture prevention |
| **Clipboard Security** | Auto-clear clipboard | Clipboard snooping |
| **Session Management** | Timeout + rotation | Session hijacking |
| **Audit Trail** | Security audit logging | Incident forensics |
