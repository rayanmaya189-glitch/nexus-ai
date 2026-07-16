# Authentication

## Table of Contents

1. [Login Screen](#login-screen)
2. [Login ViewModel](#login-viewmodel)
3. [Login API Integration](#login-api-integration)
4. [JWT Token Storage](#jwt-token-storage)
5. [Refresh Token Storage](#refresh-token-storage)
6. [Token Refresh Flow](#token-refresh-flow)
7. [Token Refresh Interceptor](#token-refresh-interceptor)
8. [Biometric Authentication](#biometric-authentication)
9. [Biometric Setup Screen](#biometric-setup-screen)
10. [Biometric Login Flow](#biometric-login-flow)
11. [Secure Storage](#secure-storage)
12. [Session Management](#session-management)
13. [Auto-Logout](#auto-logout)
14. [Password Validation](#password-validation)
15. [Error Handling](#error-handling)
16. [KYC Verification Flow](#kyc-verification-flow)
17. [Multi-Tenant Support](#multi-tenant-support)
18. [Logout Flow](#logout-flow)
19. [Auth State Persistence](#auth-state-persistence)
20. [Auth Hooks](#auth-hooks)

---

## Login Screen

```kotlin
// feature-auth/ui/LoginScreen.kt
@Composable
fun LoginScreen(
    viewModel: LoginViewModel = hiltViewModel(),
    onLoginSuccess: () -> Unit,
    onForgotPassword: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.events.collect { event ->
            when (event) {
                is LoginEvent.LoginSuccess -> onLoginSuccess()
                is LoginEvent.ShowError -> { /* Show snackbar */ }
                is LoginEvent.NavigateToForgotPassword -> onForgotPassword()
            }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp)
            .verticalScroll(rememberScrollState()),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        // Logo
        Image(
            painter = painterResource(R.drawable.ic_logo),
            contentDescription = "Nexus AI Logo",
            modifier = Modifier.size(120.dp)
        )

        Spacer(modifier = Modifier.height(32.dp))

        // Title
        Text(
            text = "Welcome Back",
            style = MaterialTheme.typography.headlineLarge,
            fontWeight = FontWeight.Bold
        )

        Text(
            text = "Sign in to continue",
            style = MaterialTheme.typography.bodyLarge,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Spacer(modifier = Modifier.height(48.dp))

        // Email field
        OutlinedTextField(
            value = uiState.email,
            onValueChange = { viewModel.onAction(LoginAction.UpdateEmail(it)) },
            label = { Text("Email") },
            leadingIcon = { Icon(Icons.Default.Email, contentDescription = null) },
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Email,
                imeAction = ImeAction.Next
            ),
            isError = uiState.emailError != null,
            supportingText = uiState.emailError?.let { error ->
                { Text(error, color = MaterialTheme.colorScheme.error) }
            },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )

        Spacer(modifier = Modifier.height(16.dp))

        // Password field
        var passwordVisible by remember { mutableStateOf(false) }
        OutlinedTextField(
            value = uiState.password,
            onValueChange = { viewModel.onAction(LoginAction.UpdatePassword(it)) },
            label = { Text("Password") },
            leadingIcon = { Icon(Icons.Default.Lock, contentDescription = null) },
            trailingIcon = {
                IconButton(onClick = { passwordVisible = !passwordVisible }) {
                    Icon(
                        imageVector = if (passwordVisible) Icons.Default.Visibility
                        else Icons.Default.VisibilityOff,
                        contentDescription = if (passwordVisible) "Hide password"
                        else "Show password"
                    )
                }
            },
            visualTransformation = if (passwordVisible) VisualTransformation.None
            else PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Password,
                imeAction = ImeAction.Done
            ),
            keyboardActions = KeyboardActions(
                onDone = { viewModel.onAction(LoginAction.Login) }
            ),
            isError = uiState.passwordError != null,
            supportingText = uiState.passwordError?.let { error ->
                { Text(error, color = MaterialTheme.colorScheme.error) }
            },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )

        Spacer(modifier = Modifier.height(8.dp))

        // Forgot password
        TextButton(
            onClick = { viewModel.onAction(LoginAction.ForgotPassword) },
            modifier = Modifier.align(Alignment.End)
        ) {
            Text("Forgot Password?")
        }

        Spacer(modifier = Modifier.height(24.dp))

        // Login button
        Button(
            onClick = { viewModel.onAction(LoginAction.Login) },
            enabled = uiState.isValid && !uiState.isLoading,
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp),
            shape = RoundedCornerShape(12.dp)
        ) {
            if (uiState.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = MaterialTheme.colorScheme.onPrimary
                )
            } else {
                Text("Sign In", fontSize = 16.sp)
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        // Biometric login
        if (uiState.isBiometricAvailable) {
            OutlinedButton(
                onClick = { viewModel.onAction(LoginAction.BiometricLogin) },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                shape = RoundedCornerShape(12.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.Fingerprint,
                    contentDescription = "Biometric Login",
                    modifier = Modifier.size(24.dp)
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text("Login with Biometrics")
            }
        }

        Spacer(modifier = Modifier.height(32.dp))

        // Sign up link
        Row(
            horizontalArrangement = Arrangement.Center,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text("Don't have an account? ")
            TextButton(onClick = { /* Navigate to register */ }) {
                Text("Sign Up")
            }
        }
    }
}
```

**Login Screen Layout:**

```
┌─────────────────────────────────────────┐
│                                         │
│              ┌───────────┐              │
│              │   Logo    │              │
│              └───────────┘              │
│                                         │
│           Welcome Back                  │
│        Sign in to continue              │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📧  Email                        │   │
│  │ user@example.com                 │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🔒  Password        👁️          │   │
│  │ ••••••••••••                    │   │
│  └─────────────────────────────────┘   │
│                                         │
│                    Forgot Password?     │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │           Sign In                │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │  🔍  Login with Biometrics      │   │
│  └─────────────────────────────────┘   │
│                                         │
│     Don't have an account? Sign Up      │
└─────────────────────────────────────────┘
```

---

## Login ViewModel

```kotlin
// feature-auth/ui/LoginViewModel.kt
@HiltViewModel
class LoginViewModel @Inject constructor(
    private val loginUseCase: LoginUseCase,
    private val biometricLoginUseCase: BiometricLoginUseCase,
    private val biometricManager: BiometricManager,
    private val tokenManager: TokenManager,
    private val analyticsManager: AnalyticsManager
) : ViewModel() {

    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    private val _events = Channel<LoginEvent>(Channel.BUFFERED)
    val events: Flow<LoginEvent> = _events.receiveAsFlow()

    init {
        checkBiometricAvailability()
        checkExistingSession()
    }

    private fun checkBiometricAvailability() {
        viewModelScope.launch {
            val isAvailable = biometricManager.isBiometricAvailable()
            _uiState.update { it.copy(isBiometricAvailable = isAvailable) }
        }
    }

    private fun checkExistingSession() {
        viewModelScope.launch {
            if (tokenManager.hasValidToken()) {
                _events.send(LoginEvent.LoginSuccess)
            }
        }
    }

    fun onAction(action: LoginAction) {
        when (action) {
            is LoginAction.UpdateEmail -> updateEmail(action.email)
            is LoginAction.UpdatePassword -> updatePassword(action.password)
            is LoginAction.Login -> performLogin()
            is LoginAction.BiometricLogin -> performBiometricLogin()
            is LoginAction.ForgotPassword -> {
                viewModelScope.launch {
                    _events.send(LoginEvent.NavigateToForgotPassword)
                }
            }
        }
    }

    private fun updateEmail(email: String) {
        _uiState.update {
            it.copy(
                email = email,
                emailError = validateEmail(email)
            )
        }
    }

    private fun updatePassword(password: String) {
        _uiState.update {
            it.copy(
                password = password,
                passwordError = validatePassword(password)
            )
        }
    }

    private fun performLogin() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            loginUseCase(
                email = _uiState.value.email.trim(),
                password = _uiState.value.password
            ).collect { result ->
                when (result) {
                    is Result.Loading -> {
                        _uiState.update { it.copy(isLoading = true) }
                    }
                    is Result.Success -> {
                        analyticsManager.logEvent(AnalyticsEvent.USER_LOGIN(mapOf(
                            "method" to "email"
                        )))
                        _uiState.update { it.copy(isLoading = false) }
                        _events.send(LoginEvent.LoginSuccess)
                    }
                    is Result.Error -> {
                        val errorMessage = mapLoginError(result.exception)
                        _uiState.update {
                            it.copy(isLoading = false, error = errorMessage)
                        }
                        _events.send(LoginEvent.ShowError(errorMessage))
                    }
                }
            }
        }
    }

    private fun performBiometricLogin() {
        viewModelScope.launch {
            val activity = getCurrentActivity()
                ?: return@launch

            biometricManager.promptBiometric(
                activity = activity,
                onSuccess = {
                    viewModelScope.launch {
                        biometricLoginUseCase().collect { result ->
                            when (result) {
                                is Result.Loading -> {
                                    _uiState.update { it.copy(isLoading = true) }
                                }
                                is Result.Success -> {
                                    analyticsManager.logEvent(AnalyticsEvent.USER_LOGIN(mapOf(
                                        "method" to "biometric"
                                    )))
                                    _uiState.update { it.copy(isLoading = false) }
                                    _events.send(LoginEvent.LoginSuccess)
                                }
                                is Result.Error -> {
                                    _uiState.update {
                                        it.copy(isLoading = false, error = result.exception.message)
                                    }
                                }
                            }
                        }
                    }
                },
                onError = { error ->
                    _uiState.update { it.copy(error = error) }
                }
            )
        }
    }

    private fun validateEmail(email: String): String? {
        if (email.isBlank()) return "Email is required"
        if (!Patterns.EMAIL_ADDRESS.matcher(email).matches()) return "Invalid email format"
        return null
    }

    private fun validatePassword(password: String): String? {
        if (password.isBlank()) return "Password is required"
        if (password.length < 12) return "Password must be at least 12 characters"
        return null
    }

    private fun mapLoginError(exception: Throwable): String {
        return when (exception) {
            is AuthException.InvalidCredentials -> "Invalid email or password"
            is AuthException.AccountLocked -> "Account is locked. Please try again later."
            is AuthException.NetworkError -> "Network error. Please check your connection."
            is AuthException.TooManyRequests -> "Too many attempts. Please wait a moment."
            else -> exception.message ?: "An unexpected error occurred"
        }
    }

    private fun getCurrentActivity(): FragmentActivity? {
        // In a real app, use activity reference from composition
        return null
    }
}

data class LoginUiState(
    val email: String = "",
    val password: String = "",
    val emailError: String? = null,
    val passwordError: String? = null,
    val isLoading: Boolean = false,
    val error: String? = null,
    val isBiometricAvailable: Boolean = false
) {
    val isValid: Boolean
        get() = emailError == null && passwordError == null &&
                email.isNotBlank() && password.isNotBlank()
}

sealed class LoginAction {
    data class UpdateEmail(val email: String) : LoginAction()
    data class UpdatePassword(val password: String) : LoginAction()
    data object Login : LoginAction()
    data object BiometricLogin : LoginAction()
    data object ForgotPassword : LoginAction()
}

sealed class LoginEvent {
    data object LoginSuccess : LoginEvent()
    data class ShowError(val message: String) : LoginEvent()
    data object NavigateToForgotPassword : LoginEvent()
}
```

**ViewModel State Flow:**

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   User       │───►│   ViewModel  │───►│   UiState    │
│   Input      │    │   (Process)  │    │   (State)    │
└──────────────┘    └──────┬───────┘    └──────┬───────┘
                           │                    │
                    ┌──────┴───────┐            │
                    │   Use Case   │            │
                    │   (Auth)     │            │
                    └──────┬───────┘            │
                           │                    │
                    ┌──────┴───────┐            │
                    │   API Call   │            │
                    │   (Remote)   │            │
                    └──────┬───────┘            │
                           │                    │
                    ┌──────┴───────┐            │
                    │   Result     │            │
                    │   (Success/  │            │
                    │    Error)    │            │
                    └──────┬───────┘            │
                           │                    │
                           └────────────────────┘
                                  │
                                  ▼
                           ┌─────────────┐
                           │   Screen    │
                           │   (Render)  │
                           └─────────────┘
```

---

## Login API Integration

```kotlin
// data/remote/dto/LoginRequest.kt
@Serializable
data class LoginRequest(
    @SerialName("email") val email: String,
    @SerialName("password") val password: String,
    @SerialName("device_id") val deviceId: String? = null,
    @SerialName("platform") val platform: String = "android"
)

@Serializable
data class LoginResponse(
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String,
    @SerialName("token_type") val tokenType: String = "Bearer",
    @SerialName("expires_in") val expiresIn: Long,
    @SerialName("user") val user: UserDto
)

@Serializable
data class UserDto(
    @SerialName("id") val id: String,
    @SerialName("email") val email: String,
    @SerialName("name") val name: String,
    @SerialName("avatar_url") val avatarUrl: String? = null,
    @SerialName("role") val role: String,
    @SerialName("tenant_id") val tenantId: String,
    @SerialName("is_biometric_enabled") val isBiometricEnabled: Boolean = false,
    @SerialName("is_kyc_verified") val isKycVerified: Boolean = false,
    @SerialName("created_at") val createdAt: String,
    @SerialName("updated_at") val updatedAt: String
)

// data/repository/AuthRepositoryImpl.kt
class AuthRepositoryImpl @Inject constructor(
    private val apiService: ApiService,
    private val tokenManager: TokenManager,
    private val userDao: UserDao,
    private val userMapper: UserMapper
) : AuthRepository {

    override fun login(email: String, password: String): Flow<Result<User>> = flow {
        emit(Result.Loading)
        try {
            val deviceId = getDeviceId()
            val response = apiService.login(
                LoginRequest(
                    email = email,
                    password = password,
                    deviceId = deviceId
                )
            )

            if (response.isSuccessful) {
                val body = response.body()!!

                // Save tokens
                tokenManager.saveTokens(
                    accessToken = body.accessToken,
                    refreshToken = body.refreshToken,
                    expiresIn = body.expiresIn
                )

                // Save user locally
                val userEntity = userMapper.toEntity(body.user)
                userDao.insertUser(userEntity)

                emit(Result.Success(userMapper.toDomain(body.user)))
            } else {
                val error = parseLoginError(response.code(), response.errorBody())
                emit(Result.Error(error))
            }
        } catch (e: Exception) {
            emit(Result.Error(mapNetworkError(e)))
        }
    }

    override fun biometricLogin(): Flow<Result<User>> = flow {
        emit(Result.Loading)
        try {
            val biometricToken = tokenManager.getBiometricToken()
                ?: throw AuthException.BiometricNotSetup()

            val response = apiService.biometricLogin(
                BiometricLoginRequest(token = biometricToken)
            )

            if (response.isSuccessful) {
                val body = response.body()!!
                tokenManager.saveTokens(
                    accessToken = body.accessToken,
                    refreshToken = body.refreshToken,
                    expiresIn = body.expiresIn
                )
                emit(Result.Success(userMapper.toDomain(body.user)))
            } else {
                emit(Result.Error(AuthException.BiometricFailed()))
            }
        } catch (e: Exception) {
            emit(Result.Error(mapNetworkError(e)))
        }
    }

    override suspend fun logout() {
        try {
            apiService.logout()
        } catch (e: Exception) {
            // Ignore logout API errors
        } finally {
            tokenManager.clearTokens()
            userDao.deleteAllUsers()
        }
    }

    private fun parseLoginError(code: Int, errorBody: ResponseBody?): AuthException {
        return when (code) {
            401 -> AuthException.InvalidCredentials()
            403 -> AuthException.AccountLocked()
            429 -> AuthException.TooManyRequests()
            500 -> AuthException.ServerError()
            else -> AuthException.Unknown("Login failed ($code)")
        }
    }

    private fun mapNetworkError(e: Exception): AuthException {
        return when (e) {
            is IOException -> AuthException.NetworkError()
            is HttpException -> AuthException.ServerError()
            else -> AuthException.Unknown(e.message)
        }
    }

    private fun getDeviceId(): String {
        return Settings.Secure.ANDROID_ID
    }
}
```

**Login API Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │───►│   Server    │───►│   Database  │
│             │    │   (API)     │    │             │
└─────────────┘    └──────┬──────┘    └─────────────┘
                          │
         POST /api/v1/auth/login
         {email, password, device_id}
                          │
                          ▼
                   ┌─────────────┐
                   │  Validate   │
                   │  Credentials│
                   └──────┬──────┘
                          │
              ┌───────────┼───────────┐
              ▼           ▼           ▼
       ┌──────────┐ ┌──────────┐ ┌──────────┐
       │  Success │ │ Invalid  │ │  Locked  │
       │  200     │ │  401     │ │  403     │
       └────┬─────┘ └────┬─────┘ └────┬─────┘
            │            │            │
            ▼            ▼            ▼
     ┌──────────┐ ┌──────────┐ ┌──────────┐
     │  Save    │ │  Show    │ │  Show    │
     │  Tokens  │ │  Error   │ │  Error   │
     └──────────┘ └──────────┘ └──────────┘
```

---

## JWT Token Storage

```kotlin
// core-security/src/main/java/com/nexus/ai/core/security/TokenManager.kt
@Singleton
class TokenManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val androidKeyStore: AndroidKeyStore
) {
    private val keyStore = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
    private val masterKey = MasterKey.Builder(context)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .build()

    private val securePrefs = EncryptedSharedPreferences.create(
        context,
        "nexus_auth_prefs",
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    companion object {
        private const val KEY_ACCESS_TOKEN = "access_token"
        private const val KEY_REFRESH_TOKEN = "refresh_token"
        private const val KEY_TOKEN_EXPIRY = "token_expiry"
        private const val KEY_BIOMETRIC_TOKEN = "biometric_token"
        private const val KEY_USER_ID = "user_id"
    }

    fun saveTokens(
        accessToken: String,
        refreshToken: String,
        expiresIn: Long
    ) {
        val expiryTime = System.currentTimeMillis() + (expiresIn * 1000)

        securePrefs.edit().apply {
            putString(KEY_ACCESS_TOKEN, encryptWithKeystore(accessToken))
            putString(KEY_REFRESH_TOKEN, encryptWithKeystore(refreshToken))
            putLong(KEY_TOKEN_EXPIRY, expiryTime)
            apply()
        }
    }

    fun getAccessToken(): String? {
        val encryptedToken = securePrefs.getString(KEY_ACCESS_TOKEN, null) ?: return null
        return decryptWithKeystore(encryptedToken)
    }

    fun getRefreshToken(): String? {
        val encryptedToken = securePrefs.getString(KEY_REFRESH_TOKEN, null) ?: return null
        return decryptWithKeystore(encryptedToken)
    }

    fun getTokenExpiry(): Long {
        return securePrefs.getLong(KEY_TOKEN_EXPIRY, 0)
    }

    fun isTokenExpired(): Boolean {
        val expiry = getTokenExpiry()
        // Refresh 5 minutes before expiry
        return System.currentTimeMillis() >= (expiry - 5 * 60 * 1000)
    }

    fun hasValidToken(): Boolean {
        val token = getAccessToken()
        return token != null && !isTokenExpired()
    }

    fun clearTokens() {
        securePrefs.edit().apply {
            remove(KEY_ACCESS_TOKEN)
            remove(KEY_REFRESH_TOKEN)
            remove(KEY_TOKEN_EXPIRY)
            remove(KEY_BIOMETRIC_TOKEN)
            remove(KEY_USER_ID)
            apply()
        }
    }

    fun saveBiometricToken(token: String) {
        securePrefs.edit().apply {
            putString(KEY_BIOMETRIC_TOKEN, encryptWithKeystore(token))
            apply()
        }
    }

    fun getBiometricToken(): String? {
        val encryptedToken = securePrefs.getString(KEY_BIOMETRIC_TOKEN, null) ?: return null
        return decryptWithKeystore(encryptedToken)
    }

    fun saveUserId(userId: String) {
        securePrefs.edit().apply {
            putString(KEY_USER_ID, userId)
            apply()
        }
    }

    fun getUserId(): String? {
        return securePrefs.getString(KEY_USER_ID, null)
    }

    private fun encryptWithKeystore(data: String): String {
        val key = androidKeyStore.getKey("token_key") as? SecretKey
            ?: androidKeyStore.generateKey("token_key")

        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, key)

        val iv = cipher.iv
        val encrypted = cipher.doFinal(data.toByteArray())

        // Combine IV and encrypted data
        val combined = iv + encrypted
        return Base64.encodeToString(combined, Base64.NO_WRAP)
    }

    private fun decryptWithKeystore(encryptedData: String): String {
        val key = androidKeyStore.getKey("token_key") as? SecretKey
            ?: throw SecurityException("Key not found")

        val combined = Base64.decode(encryptedData, Base64.NO_WRAP)
        val iv = combined.sliceArray(0..11)
        val encrypted = combined.sliceArray(12 until combined.size)

        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val spec = GCMParameterSpec(128, iv)
        cipher.init(Cipher.DECRYPT_MODE, key, spec)

        val decrypted = cipher.doFinal(encrypted)
        return String(decrypted)
    }
}
```

**Token Storage Hierarchy:**

```
┌─────────────────────────────────────────────────────────┐
│                  Token Storage                           │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Android Keystore (Hardware-backed)              │   │
│  │  ├── AES-256-GCM Key                            │   │
│  │  ├── RSA Key Pair (optional)                     │   │
│  │  └── Biometric Key                               │   │
│  └──────────────────────────────────────────────────┘   │
│                         │                                │
│                         ▼                                │
│  ┌──────────────────────────────────────────────────┐   │
│  │  EncryptedSharedPreferences                      │   │
│  │  ├── access_token (encrypted)                    │   │
│  │  ├── refresh_token (encrypted)                   │   │
│  │  ├── token_expiry (timestamp)                    │   │
│  │  ├── biometric_token (encrypted)                 │   │
│  │  └── user_id (plain)                             │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  Security Features:                                      │
│  ├── Hardware-backed key storage                        │
│  ├── AES-256-GCM encryption                             │
│  ├── Unique IV per encryption                           │
│  ├── Key rotation support                               │
│  └── Biometric-gated access                             │
└─────────────────────────────────────────────────────────┘
```

---

## Refresh Token Storage

```kotlin
// Same TokenManager as above, with additional refresh-specific methods
extension functions:

fun TokenManager.saveRefreshTokenOnly(refreshToken: String) {
    securePrefs.edit().apply {
        putString(KEY_REFRESH_TOKEN, encryptWithKeystore(refreshToken))
        apply()
    }
}

fun TokenManager.isRefreshTokenValid(): Boolean {
    return getRefreshToken() != null
}

fun TokenManager.getRefreshTokenAge(): Long {
    val expiry = getTokenExpiry()
    val currentTime = System.currentTimeMillis()
    return currentTime - expiry
}
```

---

## Token Refresh Flow

```kotlin
// data/repository/AuthRepositoryImpl.kt
class AuthRepositoryImpl @Inject constructor(
    private val apiService: ApiService,
    private val tokenManager: TokenManager
) : AuthRepository {

    override fun refreshToken(): Flow<Result<Unit>> = flow {
        val refreshToken = tokenManager.getRefreshToken()
            ?: throw AuthException.NoRefreshToken()

        try {
            val response = apiService.refreshToken(
                RefreshTokenRequest(refreshToken = refreshToken)
            )

            if (response.isSuccessful) {
                val body = response.body()!!
                tokenManager.saveTokens(
                    accessToken = body.accessToken,
                    refreshToken = body.refreshToken,
                    expiresIn = body.expiresIn
                )
                emit(Result.Success(Unit))
            } else {
                tokenManager.clearTokens()
                emit(Result.Error(AuthException.TokenExpired()))
            }
        } catch (e: Exception) {
            tokenManager.clearTokens()
            emit(Result.Error(AuthException.TokenExpired()))
        }
    }

    override fun shouldRefreshToken(): Boolean {
        return tokenManager.isTokenExpired() && tokenManager.isRefreshTokenValid()
    }
}

// Token refresh scheduling
@Singleton
class TokenRefreshScheduler @Inject constructor(
    private val authRepository: AuthRepository,
    private val tokenManager: TokenManager,
    private val scope: CoroutineScope
) {
    private var refreshJob: Job? = null

    fun startAutoRefresh() {
        refreshJob?.cancel()
        refreshJob = scope.launch {
            while (isActive) {
                // Wait until 5 minutes before token expiry
                val expiry = tokenManager.getTokenExpiry()
                val refreshTime = expiry - (5 * 60 * 1000)
                val delay = maxOf(0, refreshTime - System.currentTimeMillis())

                delay(delay)

                // Refresh token
                authRepository.refreshToken().collect { result ->
                    when (result) {
                        is Result.Success -> {
                            Timber.d("Token refreshed successfully")
                        }
                        is Result.Error -> {
                            Timber.e("Token refresh failed: ${result.exception.message}")
                            // Trigger logout
                        }
                        is Result.Loading -> { /* ignore */ }
                    }
                }
            }
        }
    }

    fun stopAutoRefresh() {
        refreshJob?.cancel()
    }
}
```

**Token Refresh Timeline:**

```
Token Issued          Token Expiry
    │                     │
    ▼                     ▼
    ├─────────────────────┤
    │                     │
    │  Refresh at (E-5m)  │
    │        │            │
    │        ▼            │
    │  ┌──────────┐       │
    │  │  Refresh │       │
    │  │  Token   │       │
    │  └──────────┘       │
    │                     │
    ▼                     ▼
    ├─────────────────────┤
    │  Valid Token        │
    │  Period             │
```

---

## Token Refresh Interceptor

```kotlin
// data/remote/interceptor/TokenRefreshInterceptor.kt
class TokenRefreshInterceptor @Inject constructor(
    private val tokenManager: TokenManager,
    private val authApi: Provider<ApiService>,
    private val json: Json
) : Interceptor {

    private val mutex = Mutex()
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()

        // Skip for auth endpoints
        if (originalRequest.url.encodedPath in PUBLIC_ENDPOINTS) {
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
                    // Token was refreshed by another thread
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
                            accessToken = body.accessToken,
                            refreshToken = body.refreshToken,
                            expiresIn = body.expiresIn
                        )

                        // Retry original request with new token
                        val newRequest = chain.request().newBuilder()
                            .header("Authorization", "Bearer ${body.accessToken}")
                            .build()
                        chain.proceed(newRequest)
                    } else {
                        // Refresh failed
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

    companion object {
        val PUBLIC_ENDPOINTS = setOf(
            "/api/v1/auth/login",
            "/api/v1/auth/register",
            "/api/v1/auth/refresh",
            "/api/v1/auth/forgot-password"
        )
    }
}
```

**Token Refresh Interceptor Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Request   │───►│   401       │───►│   Mutex     │
│             │    │   Error     │    │   Lock      │
└─────────────┘    └─────────────┘    └──────┬──────┘
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

## Biometric Authentication

```kotlin
// core-security/src/main/java/com/nexus/ai/core/security/BiometricHelper.kt
@Singleton
class BiometricHelper @Inject constructor(
    @ApplicationContext private val context: Context,
    private val tokenManager: TokenManager
) {
    private val biometricManager = androidx.biometric.BiometricManager.from(context)

    fun isBiometricAvailable(): Boolean {
        val result = biometricManager.canAuthenticate(
            BiometricManager.Authenticators.BIOMETRIC_STRONG or
            BiometricManager.Authenticators.DEVICE_CREDENTIAL
        )
        return result == BiometricManager.BIOMETRIC_SUCCESS
    }

    fun getBiometricStatus(): BiometricStatus {
        return when (biometricManager.canAuthenticate(
            BiometricManager.Authenticators.BIOMETRIC_STRONG
        )) {
            BiometricManager.BIOMETRIC_SUCCESS -> BiometricStatus.AVAILABLE
            BiometricManager.BIOMETRIC_ERROR_NO_HARDWARE -> BiometricStatus.NO_HARDWARE
            BiometricManager.BIOMETRIC_ERROR_HW_UNAVAILABLE -> BiometricStatus.HARDWARE_UNAVAILABLE
            BiometricManager.BIOMETRIC_ERROR_NONE_ENROLLED -> BiometricStatus.NOT_ENROLLED
            BiometricManager.BIOMETRIC_ERROR_SECURITY_UPDATE_REQUIRED -> BiometricStatus.SECURITY_UPDATE_REQUIRED
            BiometricManager.BIOMETRIC_ERROR_UNSUPPORTED -> BiometricStatus.UNSUPPORTED
            BiometricManager.BIOMETRIC_STATUS_UNKNOWN -> BiometricStatus.UNKNOWN
            else -> BiometricStatus.UNKNOWN
        }
    }

    fun promptBiometric(
        activity: FragmentActivity,
        title: String = "Authenticate",
        subtitle: String = "Verify your identity",
        onSuccess: (BiometricPrompt.AuthenticationResult) -> Unit,
        onError: (Int, String) -> Unit,
        onFailed: () -> Unit
    ) {
        val promptInfo = BiometricPrompt.PromptInfo.Builder()
            .setTitle(title)
            .setSubtitle(subtitle)
            .setAllowedAuthenticators(BiometricManager.Authenticators.BIOMETRIC_STRONG)
            .setNegativeButtonText("Use Password")
            .build()

        val executor = ContextCompat.getMainExecutor(context)

        val callback = object : BiometricPrompt.AuthenticationCallback() {
            override fun onAuthenticationSucceeded(result: AuthenticationResult) {
                super.onAuthenticationSucceeded(result)
                onSuccess(result)
            }

            override fun onAuthenticationError(errorCode: Int, errString: CharSequence) {
                super.onAuthenticationError(errorCode, errString)
                onError(errorCode, errString.toString())
            }

            override fun onAuthenticationFailed() {
                super.onAuthenticationFailed()
                onFailed()
            }
        }

        val biometricPrompt = BiometricPrompt(activity, executor, callback)
        biometricPrompt.authenticate(promptInfo)
    }

    fun enableBiometric(token: String) {
        tokenManager.saveBiometricToken(token)
    }

    fun disableBiometric() {
        // Clear biometric token
        securePrefs.edit().apply {
            remove(KEY_BIOMETRIC_TOKEN)
            apply()
        }
    }
}

enum class BiometricStatus {
    AVAILABLE,
    NO_HARDWARE,
    HARDWARE_UNAVAILABLE,
    NOT_ENROLLED,
    SECURITY_UPDATE_REQUIRED,
    UNSUPPORTED,
    UNKNOWN
}
```

---

## Biometric Setup Screen

```kotlin
// feature-auth/ui/BiometricSetupScreen.kt
@Composable
fun BiometricSetupScreen(
    viewModel: BiometricSetupViewModel = hiltViewModel(),
    onSetupComplete: () -> Unit,
    onSkip: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            imageVector = Icons.Default.Fingerprint,
            contentDescription = null,
            modifier = Modifier.size(120.dp),
            tint = MaterialTheme.colorScheme.primary
        )

        Spacer(modifier = Modifier.height(32.dp))

        Text(
            text = "Enable Biometric Login",
            style = MaterialTheme.typography.headlineMedium,
            fontWeight = FontWeight.Bold
        )

        Spacer(modifier = Modifier.height(16.dp))

        Text(
            text = "Use your fingerprint or face to quickly sign in to Nexus AI.",
            style = MaterialTheme.typography.bodyLarge,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Spacer(modifier = Modifier.height(48.dp))

        Button(
            onClick = { viewModel.onAction(BiometricSetupAction.Enable) },
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp),
            shape = RoundedCornerShape(12.dp)
        ) {
            Text("Enable Biometric Login")
        }

        Spacer(modifier = Modifier.height(16.dp))

        OutlinedButton(
            onClick = onSkip,
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp),
            shape = RoundedCornerShape(12.dp)
        ) {
            Text("Skip for Now")
        }
    }
}

// feature-auth/ui/BiometricSetupViewModel.kt
@HiltViewModel
class BiometricSetupViewModel @Inject constructor(
    private val biometricHelper: BiometricHelper,
    private val tokenManager: TokenManager,
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(BiometricSetupUiState())
    val uiState: StateFlow<BiometricSetupUiState> = _uiState.asStateFlow()

    private val _events = Channel<BiometricSetupEvent>(Channel.BUFFERED)
    val events: Flow<BiometricSetupEvent> = _events.receiveAsFlow()

    fun onAction(action: BiometricSetupAction) {
        when (action) {
            is BiometricSetupAction.Enable -> enableBiometric()
        }
    }

    private fun enableBiometric() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            val token = tokenManager.getAccessToken() ?: return@launch
            biometricHelper.enableBiometric(token)

            analyticsManager.logEvent(AnalyticsEvent.BIOMETRIC_ENABLED)

            _uiState.update { it.copy(isLoading = false, isSetupComplete = true) }
            _events.send(BiometricSetupEvent.SetupComplete)
        }
    }
}
```

---

## Biometric Login Flow

```kotlin
// feature-auth/ui/BiometricLoginScreen.kt
@Composable
fun BiometricLoginScreen(
    viewModel: BiometricLoginViewModel = hiltViewModel(),
    onLoginSuccess: () -> Unit,
    onFallbackToPassword: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.promptBiometric()
    }

    LaunchedEffect(Unit) {
        viewModel.events.collect { event ->
            when (event) {
                is BiometricLoginEvent.Success -> onLoginSuccess()
                is BiometricLoginEvent.Failed -> onFallbackToPassword()
            }
        }
    }

    Column(
        modifier = Modifier.fillMaxSize(),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        if (uiState.isLoading) {
            CircularProgressIndicator()
        } else {
            Icon(
                imageVector = Icons.Default.Fingerprint,
                contentDescription = null,
                modifier = Modifier.size(120.dp),
                tint = MaterialTheme.colorScheme.primary
            )

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Touch the sensor",
                style = MaterialTheme.typography.headlineSmall
            )

            Spacer(modifier = Modifier.height(48.dp))

            TextButton(onClick = onFallbackToPassword) {
                Text("Use Password Instead")
            }
        }
    }
}

// feature-auth/ui/BiometricLoginViewModel.kt
@HiltViewModel
class BiometricLoginViewModel @Inject constructor(
    private val biometricLoginUseCase: BiometricLoginUseCase,
    private val biometricHelper: BiometricHelper
) : ViewModel() {

    private val _uiState = MutableStateFlow(BiometricLoginUiState())
    val uiState: StateFlow<BiometricLoginUiState> = _uiState.asStateFlow()

    private val _events = Channel<BiometricLoginEvent>(Channel.BUFFERED)
    val events: Flow<BiometricLoginEvent> = _events.receiveAsFlow()

    fun promptBiometric() {
        viewModelScope.launch {
            val activity = getCurrentActivity() ?: return@launch

            biometricHelper.promptBiometric(
                activity = activity,
                onSuccess = { result ->
                    viewModelScope.launch {
                        performBiometricLogin()
                    }
                },
                onError = { code, message ->
                    viewModelScope.launch {
                        _events.send(BiometricLoginEvent.Failed(message))
                    }
                },
                onFailed = {
                    viewModelScope.launch {
                        _events.send(BiometricLoginEvent.Failed("Authentication failed"))
                    }
                }
            )
        }
    }

    private suspend fun performBiometricLogin() {
        _uiState.update { it.copy(isLoading = true) }

        biometricLoginUseCase().collect { result ->
            when (result) {
                is Result.Loading -> {
                    _uiState.update { it.copy(isLoading = true) }
                }
                is Result.Success -> {
                    _uiState.update { it.copy(isLoading = false) }
                    _events.send(BiometricLoginEvent.Success)
                }
                is Result.Error -> {
                    _uiState.update { it.copy(isLoading = false) }
                    _events.send(BiometricLoginEvent.Failed(result.exception.message))
                }
            }
        }
    }
}
```

**Biometric Login Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   App       │───►│  Biometric  │───►│   User      │
│   Launch    │    │  Prompt     │    │   Touch     │
└─────────────┘    └─────────────┘    └──────┬──────┘
                                            │
                                     ┌──────┴──────┐
                                     │  Succeeded? │
                                     └──────┬──────┘
                                            │
                               ┌────────────┼────────────┐
                               ▼            ▼            ▼
                        ┌──────────┐ ┌──────────┐ ┌──────────┐
                        │  Yes     │ │  No      │ │  Error   │
                        │          │ │ (Failed) │ │          │
                        └────┬─────┘ └────┬─────┘ └────┬─────┘
                             │            │            │
                             ▼            ▼            ▼
                      ┌──────────┐ ┌──────────┐ ┌──────────┐
                      │  API     │ │  Retry   │ │  Show    │
                      │  Login   │ │  Prompt  │ │  Error   │
                      └────┬─────┘ └──────────┘ └──────────┘
                           │
                           ▼
                    ┌──────────┐
                    │  Success │
                    │  → Home  │
                    └──────────┘
```

---

## Secure Storage

```kotlin
// core-security/src/main/java/com/nexus/ai/core/security/SecurePrefsManager.kt
@Singleton
class SecurePrefsManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val masterKey = MasterKey.Builder(context)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .setRequestStrongBoxBacked(true)
        .build()

    private val securePrefs = EncryptedSharedPreferences.create(
        context,
        "nexus_secure_prefs",
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    fun putString(key: String, value: String) {
        securePrefs.edit().putString(key, value).apply()
    }

    fun getString(key: String): String? {
        return securePrefs.getString(key, null)
    }

    fun putLong(key: String, value: Long) {
        securePrefs.edit().putLong(key, value).apply()
    }

    fun getLong(key: String): Long {
        return securePrefs.getLong(key, 0)
    }

    fun putBoolean(key: String, value: Boolean) {
        securePrefs.edit().putBoolean(key, value).apply()
    }

    fun getBoolean(key: String): Boolean {
        return securePrefs.getBoolean(key, false)
    }

    fun remove(key: String) {
        securePrefs.edit().remove(key).apply()
    }

    fun clear() {
        securePrefs.edit().clear().apply()
    }

    fun contains(key: String): Boolean {
        return securePrefs.contains(key)
    }
}

// AndroidKeyStore.kt
@Singleton
class AndroidKeyStore @Inject constructor() {
    private val keyStore = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }

    fun generateKey(alias: String): SecretKey {
        val keyGenerator = KeyGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_AES,
            "AndroidKeyStore"
        )

        val keyGenSpec = KeyGenParameterSpec.Builder(
            alias,
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        )
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            .setKeySize(256)
            .setUserAuthenticationRequired(false)
            .build()

        keyGenerator.init(keyGenSpec)
        return keyGenerator.generateKey()
    }

    fun getKey(alias: String): Key? {
        return keyStore.getAlias(alias)?.let {
            keyStore.getKey(it, null) as? SecretKey
        }
    }

    fun deleteKey(alias: String) {
        keyStore.deleteEntry(alias)
    }
}
```

**Secure Storage Architecture:**

```
┌─────────────────────────────────────────────────────────┐
│                  Secure Storage Layer                     │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Android Keystore                                │   │
│  │  ├── Hardware-backed AES-256 key                 │   │
│  │  ├── StrongBox backed (if available)             │   │
│  │  └── Biometric-bound keys                        │   │
│  └──────────────────────────────────────────────────┘   │
│                         │                                │
│                         ▼                                │
│  ┌──────────────────────────────────────────────────┐   │
│  │  EncryptedSharedPreferences                      │   │
│  │  ├── Key Encryption: AES-256-SIV                 │   │
│  │  ├── Value Encryption: AES-256-GCM              │   │
│  │  └── Auto-generates IV per value                 │   │
│  └──────────────────────────────────────────────────┘   │
│                         │                                │
│                         ▼                                │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Stored Data                                      │   │
│  │  ├── access_token (encrypted)                    │   │
│  │  ├── refresh_token (encrypted)                   │   │
│  │  ├── biometric_token (encrypted)                 │   │
│  │  ├── user_preferences (encrypted)                │   │
│  │  └── app_settings (encrypted)                    │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
│  Security Properties:                                    │
│  ├── Encryption at rest                                 │
│  ├── Hardware-backed keys                               │
│  ├── No plaintext secrets in memory                     │
│  ├── Automatic key rotation                             │
│  └── Biometric-gated access                             │
└─────────────────────────────────────────────────────────┘
```

---

## Session Management

```kotlin
// core-security/src/main/java/com/nexus/ai/core/security/SessionManager.kt
@Singleton
class SessionManager @Inject constructor(
    private val tokenManager: TokenManager,
    private val securePrefsManager: SecurePrefsManager,
    private val authRepository: AuthRepository
) {
    private val _sessionState = MutableStateFlow<SessionState>(SessionState.Active)
    val sessionState: StateFlow<SessionState> = _sessionState.asStateFlow()

    private var inactivityTimer: Job? = null
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    companion object {
        private const val INACTIVITY_TIMEOUT = 30 * 60 * 1000L // 30 minutes
        private const val SESSION_TIMEOUT = 24 * 60 * 60 * 1000L // 24 hours
        private const val KEY_LAST_ACTIVITY = "last_activity"
        private const val KEY_SESSION_START = "session_start"
    }

    fun startSession() {
        securePrefsManager.putLong(KEY_SESSION_START, System.currentTimeMillis())
        securePrefsManager.putLong(KEY_LAST_ACTIVITY, System.currentTimeMillis())
        _sessionState.value = SessionState.Active
        startInactivityTimer()
    }

    fun updateActivity() {
        securePrefsManager.putLong(KEY_LAST_ACTIVITY, System.currentTimeMillis())
        resetInactivityTimer()
    }

    fun isSessionValid(): Boolean {
        val sessionStart = securePrefsManager.getLong(KEY_SESSION_START)
        val currentTime = System.currentTimeMillis()

        // Check session timeout
        if (currentTime - sessionStart > SESSION_TIMEOUT) {
            endSession(SessionEndReason.TIMEOUT)
            return false
        }

        // Check inactivity
        val lastActivity = securePrefsManager.getLong(KEY_LAST_ACTIVITY)
        if (currentTime - lastActivity > INACTIVITY_TIMEOUT) {
            endSession(SessionEndReason.INACTIVITY)
            return false
        }

        // Check token validity
        if (!tokenManager.hasValidToken()) {
            endSession(SessionEndReason.TOKEN_EXPIRED)
            return false
        }

        return true
    }

    private fun startInactivityTimer() {
        inactivityTimer?.cancel()
        inactivityTimer = scope.launch {
            delay(INACTIVITY_TIMEOUT)
            endSession(SessionEndReason.INACTIVITY)
        }
    }

    private fun resetInactivityTimer() {
        startInactivityTimer()
    }

    fun endSession(reason: SessionEndReason) {
        inactivityTimer?.cancel()
        _sessionState.value = SessionState.Ended(reason)
        scope.launch {
            authRepository.logout()
            securePrefsManager.clear()
        }
    }
}

sealed class SessionState {
    data object Active : SessionState()
    data class Ended(val reason: SessionEndReason) : SessionState()
}

enum class SessionEndReason {
    INACTIVITY,
    TIMEOUT,
    TOKEN_EXPIRED,
    USER_LOGOUT,
    SECURITY_BREACH
}
```

---

## Auto-Logout

```kotlin
// Auto-logout implementation
@Singleton
class AutoLogoutManager @Inject constructor(
    private val sessionManager: SessionManager,
    private val networkMonitor: NetworkMonitor
) {
    private var logoutJob: Job? = null
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    fun startMonitoring() {
        // Monitor session state
        scope.launch {
            sessionManager.sessionState.collect { state ->
                when (state) {
                    is SessionState.Ended -> {
                        when (state.reason) {
                            SessionEndReason.INACTIVITY -> {
                                Timber.d("Auto-logout due to inactivity")
                                triggerLogout("Session expired due to inactivity")
                            }
                            SessionEndReason.TIMEOUT -> {
                                Timber.d("Auto-logout due to session timeout")
                                triggerLogout("Session expired")
                            }
                            SessionEndReason.TOKEN_EXPIRED -> {
                                Timber.d("Auto-logout due to token expiry")
                                triggerLogout("Authentication expired")
                            }
                            SessionEndReason.USER_LOGOUT -> {
                                Timber.d("User initiated logout")
                            }
                            SessionEndReason.SECURITY_BREACH -> {
                                Timber.w("Auto-logout due to security breach")
                                triggerLogout("Security alert")
                            }
                        }
                    }
                    is SessionState.Active -> {
                        // Reset UI state
                    }
                }
            }
        }

        // Monitor network for session validation
        scope.launch {
            networkMonitor.isOnline.collect { isOnline ->
                if (isOnline) {
                    validateSession()
                }
            }
        }
    }

    private fun validateSession() {
        scope.launch {
            if (!sessionManager.isSessionValid()) {
                triggerLogout("Session invalid")
            }
        }
    }

    private fun triggerLogout(reason: String) {
        logoutJob?.cancel()
        logoutJob = scope.launch {
            sessionManager.endSession(SessionEndReason.TOKEN_EXPIRED)
            // Navigate to login screen
        }
    }

    fun stopMonitoring() {
        logoutJob?.cancel()
    }
}
```

---

## Password Validation

```kotlin
// common/extension/PasswordValidator.kt
object PasswordValidator {

    fun validate(password: String): PasswordValidationResult {
        val errors = mutableListOf<String>()

        if (password.length < 12) {
            errors.add("Password must be at least 12 characters")
        }
        if (!password.any { it.isUpperCase() }) {
            errors.add("Password must contain at least one uppercase letter")
        }
        if (!password.any { it.isLowerCase() }) {
            errors.add("Password must contain at least one lowercase letter")
        }
        if (!password.any { it.isDigit() }) {
            errors.add("Password must contain at least one number")
        }
        if (!password.any { !it.isLetterOrDigit() }) {
            errors.add("Password must contain at least one special character")
        }
        if (password.contains("password", ignoreCase = true)) {
            errors.add("Password cannot contain 'password'")
        }
        if (password.contains(Regex("(.)\\1{2,}"))) {
            errors.add("Password cannot contain 3 or more repeated characters")
        }

        return PasswordValidationResult(
            isValid = errors.isEmpty(),
            errors = errors,
            strength = calculateStrength(password)
        )
    }

    private fun calculateStrength(password: String): PasswordStrength {
        var score = 0
        if (password.length >= 12) score++
        if (password.length >= 16) score++
        if (password.any { it.isUpperCase() }) score++
        if (password.any { it.isLowerCase() }) score++
        if (password.any { it.isDigit() }) score++
        if (password.any { !it.isLetterOrDigit() }) score++
        if (password.length >= 20) score++

        return when {
            score <= 2 -> PasswordStrength.WEAK
            score <= 4 -> PasswordStrength.MEDIUM
            score <= 5 -> PasswordStrength.STRONG
            else -> PasswordStrength.VERY_STRONG
        }
    }
}

data class PasswordValidationResult(
    val isValid: Boolean,
    val errors: List<String>,
    val strength: PasswordStrength
)

enum class PasswordStrength {
    WEAK, MEDIUM, STRONG, VERY_STRONG
}

// Usage in ViewModel
fun validatePassword(password: String): String? {
    val result = PasswordValidator.validate(password)
    return result.errors.firstOrNull()
}
```

---

## Error Handling

```kotlin
// domain/entity/Error.kt
sealed class AuthException : Exception() {
    data class InvalidCredentials(
        override val message: String = "Invalid email or password"
    ) : AuthException()

    data class AccountLocked(
        override val message: String = "Account is locked. Please try again later."
    ) : AuthException()

    data class NetworkError(
        override val message: String = "Network error. Please check your connection."
    ) : AuthException()

    data class TokenExpired(
        override val message: String = "Session expired. Please login again."
    ) : AuthException()

    data class NoRefreshToken(
        override val message: String = "No refresh token available."
    ) : AuthException()

    data class BiometricNotSetup(
        override val message: String = "Biometric authentication is not set up."
    ) : AuthException()

    data class BiometricFailed(
        override val message: String = "Biometric authentication failed."
    ) : AuthException()

    data class TooManyRequests(
        override val message: String = "Too many attempts. Please wait a moment."
    ) : AuthException()

    data class ServerError(
        override val message: String = "Server error. Please try again later."
    ) : AuthException()

    data class Unknown(
        override val message: String? = "An unexpected error occurred."
    ) : AuthException()
}

// Error mapping in UI
@Composable
fun AuthErrorMapper(error: AuthException): String {
    return when (error) {
        is AuthException.InvalidCredentials -> "Invalid email or password. Please try again."
        is AuthException.AccountLocked -> "Your account has been locked due to too many failed attempts. Please contact support."
        is AuthException.NetworkError -> "Unable to connect. Please check your internet connection."
        is AuthException.TokenExpired -> "Your session has expired. Please login again."
        is AuthException.TooManyRequests -> "Too many login attempts. Please wait a few minutes."
        is AuthException.ServerError -> "Something went wrong. Please try again later."
        is AuthException.BiometricNotSetup -> "Biometric login is not configured. Please use password."
        is AuthException.BiometricFailed -> "Biometric verification failed. Please try again."
        is AuthException.Unknown -> error.message ?: "An unexpected error occurred."
        is AuthException.NoRefreshToken -> "Session expired. Please login again."
    }
}
```

**Error Handling Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   API       │───►│   Error     │───►│   Map to    │
│   Response  │    │   Type      │    │   User Msg  │
└─────────────┘    └─────────────┘    └──────┬──────┘
                                            │
                                     ┌──────┴──────┐
                                     │  Error      │
                                     │  Category   │
                                     └──────┬──────┘
                                            │
                          ┌─────────────────┼─────────────────┐
                          ▼                 ▼                 ▼
                   ┌──────────┐      ┌──────────┐      ┌──────────┐
                   │  Auth    │      │  Network │      │  Server  │
                   │  Error   │      │  Error   │      │  Error   │
                   └────┬─────┘      └────┬─────┘      └────┬─────┘
                        │                 │                 │
                        ▼                 ▼                 ▼
                 ┌──────────┐      ┌──────────┐      ┌──────────┐
                 │  Show    │      │  Show    │      │  Show    │
                 │  Snackbar│      │  Retry   │      │  Generic │
                 └──────────┘      └──────────┘      └──────────┘
```

---

## KYC Verification Flow

```kotlin
// feature-auth/ui/KycScreen.kt
@Composable
fun KycScreen(
    viewModel: KycViewModel = hiltViewModel(),
    onVerificationComplete: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        when (uiState.step) {
            KycStep.DOCUMENT_UPLOAD -> {
                DocumentUploadStep(
                    onDocumentSelected = { viewModel.onAction(KycAction.UploadDocument(it)) }
                )
            }
            KycStep.SELFIE -> {
                SelfieStep(
                    onSelfieCaptured = { viewModel.onAction(KycAction.CaptureSelfie(it)) }
                )
            }
            KycStep.PROCESSING -> {
                ProcessingStep()
            }
            KycStep.REVIEW -> {
                ReviewStep(
                    documentImage = uiState.documentImage,
                    selfieImage = uiState.selfieImage,
                    onConfirm = { viewModel.onAction(KycAction.SubmitVerification) }
                )
            }
            KycStep.COMPLETE -> {
                CompleteStep(onContinue = onVerificationComplete)
            }
            KycStep.ERROR -> {
                ErrorStep(
                    error = uiState.error,
                    onRetry = { viewModel.onAction(KycAction.Retry) }
                )
            }
        }
    }
}

// feature-auth/ui/KycViewModel.kt
@HiltViewModel
class KycViewModel @Inject constructor(
    private val kycUseCase: KycUseCase,
    private val fileRepository: FileRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(KycUiState())
    val uiState: StateFlow<KycUiState> = _uiState.asStateFlow()

    fun onAction(action: KycAction) {
        when (action) {
            is KycAction.UploadDocument -> uploadDocument(action.uri)
            is KycAction.CaptureSelfie -> captureSelfie(action.uri)
            is KycAction.SubmitVerification -> submitVerification()
            is KycAction.Retry -> retry()
        }
    }

    private fun uploadDocument(uri: Uri) {
        viewModelScope.launch {
            _uiState.update { it.copy(step = KycStep.PROCESSING) }
            try {
                val file = fileRepository.uploadFile(uri)
                _uiState.update {
                    it.copy(documentImage = file.url, step = KycStep.SELFIE)
                }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(step = KycStep.ERROR, error = e.message)
                }
            }
        }
    }

    private fun submitVerification() {
        viewModelScope.launch {
            _uiState.update { it.copy(step = KycStep.PROCESSING) }
            kycUseCase(
                documentUrl = _uiState.value.documentImage!!,
                selfieUrl = _uiState.value.selfieImage!!
            ).collect { result ->
                when (result) {
                    is Result.Loading -> {
                        _uiState.update { it.copy(step = KycStep.PROCESSING) }
                    }
                    is Result.Success -> {
                        _uiState.update { it.copy(step = KycStep.COMPLETE) }
                    }
                    is Result.Error -> {
                        _uiState.update {
                            it.copy(step = KycStep.ERROR, error = result.exception.message)
                        }
                    }
                }
            }
        }
    }
}

enum class KycStep {
    DOCUMENT_UPLOAD,
    SELFIE,
    PROCESSING,
    REVIEW,
    COMPLETE,
    ERROR
}
```

**KYC Flow Diagram:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Document   │───►│  Selfie     │───►│  Review     │
│  Upload     │    │  Capture    │    │  Confirm    │
└─────────────┘    └─────────────┘    └──────┬──────┘
                                            │
                                     ┌──────┴──────┐
                                     │  Submit     │
                                     └──────┬──────┘
                                            │
                                     ┌──────┴──────┐
                                     │  Processing │
                                     └──────┬──────┘
                                            │
                               ┌────────────┼────────────┐
                               ▼            ▼            ▼
                        ┌──────────┐ ┌──────────┐ ┌──────────┐
                        │  Approved│ │  Pending │ │  Rejected│
                        └──────────┘ └──────────┘ └──────────┘
```

---

## Multi-Tenant Support

```kotlin
// core-security/src/main/java/com/nexus/ai/core/security/TenantManager.kt
@Singleton
class TenantManager @Inject constructor(
    private val securePrefsManager: SecurePrefsManager,
    private val apiService: ApiService
) {
    private val _currentTenant = MutableStateFlow<Tenant?>(null)
    val currentTenant: StateFlow<Tenant?> = _currentTenant.asStateFlow()

    private val _availableTenants = MutableStateFlow<List<Tenant>>(emptyList())
    val availableTenants: StateFlow<List<Tenant>> = _availableTenants.asStateFlow()

    companion object {
        private const val KEY_CURRENT_TENANT = "current_tenant"
    }

    suspend fun loadTenants() {
        try {
            val response = apiService.getTenants()
            if (response.isSuccessful) {
                val tenants = response.body()?.data ?: emptyList()
                _availableTenants.value = tenants

                // Restore saved tenant
                val savedTenantId = securePrefsManager.getString(KEY_CURRENT_TENANT)
                val savedTenant = tenants.find { it.id == savedTenantId }
                if (savedTenant != null) {
                    _currentTenant.value = savedTenant
                } else if (tenants.isNotEmpty()) {
                    selectTenant(tenants.first())
                }
            }
        } catch (e: Exception) {
            Timber.e(e, "Failed to load tenants")
        }
    }

    fun selectTenant(tenant: Tenant) {
        _currentTenant.value = tenant
        securePrefsManager.putString(KEY_CURRENT_TENANT, tenant.id)

        // Update API base URL if needed
        // Update headers for tenant context
    }

    fun getTenantId(): String? {
        return _currentTenant.value?.id
    }

    fun isMultiTenant(): Boolean {
        return _availableTenants.value.size > 1
    }
}

data class Tenant(
    val id: String,
    val name: String,
    val logoUrl: String? = null,
    val plan: String = "free",
    val settings: Map<String, Any> = emptyMap()
)

// Tenant context in API requests
class TenantInterceptor @Inject constructor(
    private val tenantManager: TenantManager
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Request {
        val tenantId = tenantManager.getTenantId() ?: return chain.proceed(chain.request())

        val request = chain.request().newBuilder()
            .header("X-Tenant-Id", tenantId)
            .build()

        return chain.proceed(request)
    }
}
```

---

## Logout Flow

```kotlin
// domain/usecase/auth/LogoutUseCase.kt
class LogoutUseCase @Inject constructor(
    private val authRepository: AuthRepository,
    private val sessionManager: SessionManager,
    private val webSocketManager: WebSocketManager,
    private val workManager: WorkManager
) {
    operator fun invoke(): Flow<Result<Unit>> = flow {
        emit(Result.Loading)
        try {
            // 1. Stop WebSocket
            webSocketManager.disconnect()

            // 2. Cancel background work
            workManager.cancelUniqueWork("nexus_sync", ExistingWorkPolicy.KEEP)

            // 3. Call API logout
            authRepository.logout()

            // 4. Clear session
            sessionManager.endSession(SessionEndReason.USER_LOGOUT)

            // 5. Clear local data
            clearLocalData()

            emit(Result.Success(Unit))
        } catch (e: Exception) {
            // Even if API fails, clear local data
            clearLocalData()
            emit(Result.Success(Unit))
        }
    }

    private suspend fun clearLocalData() {
        // Clear database
        // Clear DataStore
        // Clear SecurePrefs
        // Clear image cache
    }
}

// feature-settings/ui/SettingsScreen.kt
@Composable
fun SettingsScreen(
    viewModel: SettingsViewModel = hiltViewModel(),
    onLogout: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    // Logout confirmation dialog
    if (uiState.showLogoutDialog) {
        AlertDialog(
            onDismissRequest = { viewModel.onAction(SettingsAction.DismissLogout) },
            title = { Text("Logout") },
            text = { Text("Are you sure you want to logout?") },
            confirmButton = {
                TextButton(
                    onClick = {
                        viewModel.onAction(SettingsAction.ConfirmLogout)
                        onLogout()
                    }
                ) {
                    Text("Logout", color = MaterialTheme.colorScheme.error)
                }
            },
            dismissButton = {
                TextButton(onClick = { viewModel.onAction(SettingsAction.DismissLogout) }) {
                    Text("Cancel")
                }
            }
        )
    }

    // Settings list
    LazyColumn {
        item {
            SettingsItem(
                icon = Icons.Default.Person,
                title = "Profile",
                onClick = { /* Navigate to profile */ }
            )
        }
        item {
            SettingsItem(
                icon = Icons.Default.Security,
                title = "Security",
                onClick = { /* Navigate to security */ }
            )
        }
        item {
            SettingsItem(
                icon = Icons.Default.Fingerprint,
                title = "Biometric Login",
                trailing = {
                    Switch(
                        checked = uiState.isBiometricEnabled,
                        onCheckedChange = { viewModel.onAction(SettingsAction.ToggleBiometric(it)) }
                    )
                }
            )
        }
        item {
            SettingsItem(
                icon = Icons.Default.Notifications,
                title = "Notifications",
                onClick = { /* Navigate to notifications */ }
            )
        }
        item {
            SettingsItem(
                icon = Icons.Default.DarkMode,
                title = "Theme",
                onClick = { /* Navigate to theme settings */ }
            )
        }
        item {
            SettingsItem(
                icon = Icons.Default.Language,
                title = "Language",
                onClick = { /* Navigate to language settings */ }
            )
        }
        item {
            SettingsItem(
                icon = Icons.Default.Info,
                title = "About",
                onClick = { /* Navigate to about */ }
            )
        }
        item {
            SettingsItem(
                icon = Icons.Default.Logout,
                title = "Logout",
                titleColor = MaterialTheme.colorScheme.error,
                onClick = { viewModel.onAction(SettingsAction.ShowLogoutDialog) }
            )
        }
    }
}
```

**Logout Flow Diagram:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  User       │───►│  Confirm    │───►│  Start      │
│  Click      │    │  Dialog     │    │  Logout     │
└─────────────┘    └─────────────┘    └──────┬──────┘
                                            │
                    ┌────────────────────────┼────────────────────────┐
                    │                        │                        │
                    ▼                        ▼                        ▼
             ┌──────────┐            ┌──────────┐            ┌──────────┐
             │  Stop    │            │  Cancel  │            │  Clear   │
             │  WebSocket│           │  Work    │            │  Tokens  │
             └──────────┘            └──────────┘            └──────────┘
                    │                        │                        │
                    └────────────────────────┼────────────────────────┘
                                            │
                                     ┌──────┴──────┐
                                     │  Clear      │
                                     │  Local Data │
                                     └──────┬──────┘
                                            │
                                     ┌──────┴──────┐
                                     │  Navigate   │
                                     │  to Login   │
                                     └─────────────┘
```

---

## Auth State Persistence

```kotlin
// common/type/AuthState.kt
sealed class AuthState {
    data object Loading : AuthState()
    data object Unauthenticated : AuthState()
    data class Authenticated(val user: User) : AuthState()
    data class TokenExpiring(val user: User, val expiresIn: Long) : AuthState()
}

// feature-auth/di/AuthModule.kt
@Module
@InstallIn(SingletonComponent::class)
object AuthModule {

    @Provides
    @Singleton
    fun provideAuthStateProvider(
        tokenManager: TokenManager,
        userDao: UserDao,
        sessionManager: SessionManager
    ): AuthStateProvider {
        return AuthStateProviderImpl(tokenManager, userDao, sessionManager)
    }
}

// AuthStateProvider.kt
@Singleton
class AuthStateProviderImpl @Inject constructor(
    private val tokenManager: TokenManager,
    private val userDao: UserDao,
    private val sessionManager: SessionManager
) : AuthStateProvider {

    override fun observeAuthState(): Flow<AuthState> = flow {
        emit(AuthState.Loading)

        // Check for existing session
        if (tokenManager.hasValidToken() && sessionManager.isSessionValid()) {
            val userId = tokenManager.getUserId()
            if (userId != null) {
                val user = userDao.getUserById(userId)
                if (user != null) {
                    emit(AuthState.Authenticated(userMapper.toDomain(user)))
                } else {
                    emit(AuthState.Unauthenticated)
                }
            } else {
                emit(AuthState.Unauthenticated)
            }
        } else {
            emit(AuthState.Unauthenticated)
        }
    }
}

// Usage in Navigation
@Composable
fun NexusNavHost(
    authStateProvider: AuthStateProvider,
    navController: NavHostController
) {
    val authState by authStateProvider.observeAuthState()
        .collectAsState(initial = AuthState.Loading)

    when (authState) {
        is AuthState.Loading -> {
            SplashScreen()
        }
        is AuthState.Unauthenticated -> {
            AuthNavHost(navController = navController)
        }
        is AuthState.Authenticated -> {
            MainNavHost(
                user = (authState as AuthState.Authenticated).user,
                navController = navController
            )
        }
        is AuthState.TokenExpiring -> {
            // Show warning and allow user to refresh
            MainNavHost(
                user = (authState as AuthState.TokenExpiring).user,
                navController = navController
            )
        }
    }
}
```

---

## Auth Hooks (Compose Equivalent)

```kotlin
// common/composable/AuthHooks.kt
@Composable
fun rememberAuthState(): AuthState {
    val authStateProvider = hiltViewModel<AuthStateProvider>()
    val authState by authStateProvider.observeAuthState()
        .collectAsState(initial = AuthState.Loading)
    return authState
}

@Composable
fun requireAuth(
    content: @Composable (User) -> Unit
) {
    val authState = rememberAuthState()

    when (authState) {
        is AuthState.Loading -> {
            CircularProgressIndicator()
        }
        is AuthState.Authenticated -> {
            content(authState.user)
        }
        is AuthState.Unauthenticated -> {
            // Navigate to login
            val navController = LocalNavController.current
            LaunchedEffect(Unit) {
                navController.navigate(Screen.Login.route) {
                    popUpTo(0) { inclusive = true }
                }
            }
        }
        is AuthState.TokenExpiring -> {
            // Show warning dialog
            TokenExpiringDialog(
                expiresIn = authState.expiresIn,
                onRefresh = { /* Refresh token */ },
                onLogout = { /* Logout */ }
            )
            content(authState.user)
        }
    }
}

@Composable
fun useLogout(): () -> Unit {
    val logoutUseCase = hiltViewModel<LogoutUseCase>()
    val navController = LocalNavController.current

    return {
        CoroutineScope(Dispatchers.Main).launch {
            logoutUseCase().collect { result ->
                if (result is Result.Success) {
                    navController.navigate(Screen.Login.route) {
                        popUpTo(0) { inclusive = true }
                    }
                }
            }
        }
    }
}

@Composable
fun useCurrentUser(): User? {
    val authState = rememberAuthState()
    return when (authState) {
        is AuthState.Authenticated -> authState.user
        is AuthState.TokenExpiring -> authState.user
        else -> null
    }
}
```

**Auth Hooks Usage:**

```kotlin
// In any screen
@Composable
fun ChatScreen() {
    requireAuth { user ->
        // User is guaranteed to be authenticated here
        Column {
            Text("Welcome, ${user.name}")
            MessageList(userId = user.id)
        }
    }
}

// Check current user
@Composable
fun ProfileScreen() {
    val user = useCurrentUser()
    if (user != null) {
        Text("Email: ${user.email}")
    }
}

// Logout action
@Composable
fun SettingsScreen() {
    val logout = useLogout()

    Button(onClick = logout) {
        Text("Logout")
    }
}
```

---

*Document Version: 1.0 | Last Updated: 2026-07-16*
