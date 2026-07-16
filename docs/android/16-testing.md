# 16 — Testing

## 1. Testing Strategy Overview

```
                    /\
                   /  \          UI / E2E Tests
                  / E2E\         (Few, Slow, Expensive)
                 /______\
                /        \       Integration Tests
               / Integr.  \     (Moderate, Medium Speed)
              /____________\
             /              \    Unit Tests
            /   Unit Tests   \   (Many, Fast, Cheap)
           /__________________\
```

### Testing Pyramid Targets

| Layer          | % of Tests | Speed     | Cost   | Scope                  |
|----------------|-----------|-----------|--------|------------------------|
| Unit           | 70%       | <10ms     | Low    | Single class/function  |
| Integration    | 20%       | <500ms    | Medium | Multiple components    |
| UI / E2E       | 10%       | <5s       | High   | Full user flow         |

### Test Coverage Targets

| Module              | Target Coverage |
|---------------------|----------------|
| Domain (Use Cases)  | 95%+           |
| Data (Repository)   | 90%+           |
| ViewModel           | 85%+           |
| UI (Compose)        | 70%+           |
| Overall             | 80%+           |

---

## 2. Unit Testing

### 2.1 JUnit 5 Setup

```kotlin
// build.gradle.kts (testImplementation)
testImplementation("org.junit.jupiter:junit-jupiter-api:5.10.1")
testImplementation("org.junit.jupiter:junit-jupiter-params:5.10.1")
testRuntimeOnly("org.junit.jupiter:junit-jupiter-engine:5.10.1")
```

### 2.2 Basic JUnit 5 Tests

```kotlin
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.DisplayName

class CurrencyFormatterTest {

    private lateinit var formatter: CurrencyFormatter

    @BeforeEach
    fun setup() {
        formatter = CurrencyFormatter(locale = Locale.US)
    }

    @Nested
    @DisplayName("formatAmount")
    inner class FormatAmountTests {

        @Test
        fun `should format positive amount with dollar sign`() {
            val result = formatter.formatAmount(1234.50)
            assertEquals("$1,234.50", result)
        }

        @Test
        fun `should format zero as $0_00`() {
            val result = formatter.formatAmount(0.0)
            assertEquals("$0.00", result)
        }

        @Test
        fun `should format negative amount with parentheses`() {
            val result = formatter.formatAmount(-500.0)
            assertEquals("($500.00)", result)
        }

        @Test
        fun `should round to two decimal places`() {
            val result = formatter.formatAmount(19.999)
            assertEquals("$20.00", result)
        }
    }
}
```

### 2.3 Parameterized Tests

```kotlin
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.CsvSource
import org.junit.jupiter.params.provider.ValueSource

class EmailValidatorTest {

    private val validator = EmailValidator()

    @ParameterizedTest
    @ValueSource(
        strings = [
            "user@example.com",
            "test.user@domain.co",
            "admin+tag@company.org"
        ]
    )
    fun `should accept valid emails`(email: String) {
        assertTrue(validator.isValid(email))
    }

    @ParameterizedTest
    @CsvSource(
        "invalid, false",
        "@nodomain.com, false",
        "no-at-sign.com, false",
        ", false",
        "'', false"
    )
    fun `should reject invalid emails`(email: String, expected: Boolean) {
        assertEquals(expected, validator.isValid(email))
    }

    @ParameterizedTest
    @CsvSource(
        "100.00, USD, $100.00",
        "100.00, EUR, €100.00",
        "100.00, JPY, ¥100"
    )
    fun `should format amount in different currencies`(
        amount: Double, currency: String, expected: String
    ) {
        val formatter = CurrencyFormatter(currency = currency)
        assertEquals(expected, formatter.formatAmount(amount))
    }
}
```

### 2.4 MockK Basics

```kotlin
import io.mockk.*
import io.mockk.impl.annotations.MockK
import io.mockk.junit5.MockKExtension
import org.junit.jupiter.api.extension.ExtendWith

@ExtendWith(MockKExtension::class)
class LoginUseCaseTest {

    @MockK
    private lateinit var authRepository: AuthRepository

    @MockK
    private lateinit var tokenStorage: TokenStorage

    private lateinit var loginUseCase: LoginUseCase

    @BeforeEach
    fun setup() {
        loginUseCase = LoginUseCase(authRepository, tokenStorage)
    }

    @Test
    fun `should save token on successful login`() = runTest {
        // Arrange
        val token = AuthToken(accessToken = "abc", refreshToken = "xyz")
        coEvery { authRepository.login("user@test.com", "pass123") } returns
            Result.success(token)
        coEvery { tokenStorage.save(any()) } just Runs

        // Act
        val result = loginUseCase("user@test.com", "pass123")

        // Assert
        assertTrue(result.isSuccess)
        coVerify { tokenStorage.save(token) }
    }

    @Test
    fun `should return error on invalid credentials`() = runTest {
        // Arrange
        coEvery { authRepository.login(any(), any()) } returns
            Result.failure(InvalidCredentialsException())

        // Act
        val result = loginUseCase("user@test.com", "wrong")

        // Assert
        assertTrue(result.isFailure)
        assertInstanceOf(InvalidCredentialsException::class.java, result.exceptionOrNull())
        coVerify(exactly = 0) { tokenStorage.save(any()) }
    }
}
```

### 2.5 Turbine (Flow Testing)

```kotlin
import app.cash.turbine.test
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Test

class AuthStateFlowTest {

    private val viewModel = AuthViewModel(loginUseCase = mockk())

    @Test
    fun `should emit Loading then Success on login`() = runTest {
        viewModel.uiState.test {
            // Initial state
            assertEquals(AuthUiState.Idle, awaitItem())

            // Trigger login
            viewModel.onAction(AuthAction.Login("user@test.com", "pass"))

            // Loading state
            assertEquals(AuthUiState.Loading, awaitItem())

            // Success state
            val success = awaitItem()
            assertInstanceOf(AuthUiState.Success::class.java, success)

            cancelAndIgnoreRemainingEvents()
        }
    }

    @Test
    fun `should emit Idle then Error on login failure`() = runTest {
        viewModel.uiState.test {
            assertEquals(AuthUiState.Idle, awaitItem())

            viewModel.onAction(AuthAction.Login("user@test.com", "wrong"))

            assertEquals(AuthUiState.Loading, awaitItem())

            val error = awaitItem()
            assertInstanceOf(AuthUiState.Error::class.java, error)
            assertEquals("Invalid credentials", (error as AuthUiState.Error).message)

            cancelAndIgnoreRemainingEvents()
        }
    }

    @Test
    fun `should collect websocket messages in order`() = runTest {
        val messageFlow = webSocketRepository.observeMessages()

        messageFlow.test {
            val msg1 = awaitItem()
            assertEquals("Hello", msg1.content)

            val msg2 = awaitItem()
            assertEquals("World", msg2.content)

            cancelAndIgnoreRemainingEvents()
        }
    }
}
```

---

## 3. ViewModel Testing

### 3.1 State Testing

```kotlin
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.jupiter.api.*

@OptIn(ExperimentalCoroutinesApi::class)
class WalletViewModelTest {

    private val testDispatcher = UnconfinedTestDispatcher()
    private lateinit var viewModel: WalletViewModel
    private lateinit var getWalletUseCase: GetWalletUseCase
    private lateinit var sendTransactionUseCase: SendTransactionUseCase

    @BeforeEach
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        getWalletUseCase = mockk()
        sendTransactionUseCase = mockk()
        viewModel = WalletViewModel(getWalletUseCase, sendTransactionUseCase)
    }

    @AfterEach
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `should load wallet balance on init`() = runTest {
        val wallet = Wallet(balance = BigDecimal("1500.00"), currency = "USD")
        coEvery { getWalletUseCase() } returns Result.success(wallet)

        viewModel.uiState.test {
            skipItems(1) // skip idle

            viewModel.onAction(WalletAction.LoadWallet)

            val loading = awaitItem()
            assertTrue(loading is WalletUiState.Loading)

            val success = awaitItem() as WalletUiState.Success
            assertEquals("1,500.00", success.balance)
            assertEquals("USD", success.currency)

            cancelAndIgnoreRemainingEvents()
        }
    }

    @Test
    fun `should update balance optimistically on send`() = runTest {
        val wallet = Wallet(balance = BigDecimal("1500.00"), currency = "USD")
        coEvery { getWalletUseCase() } returns Result.success(wallet)
        coEvery { sendTransactionUseCase(any()) } returns Result.success(Unit)

        viewModel.onAction(WalletAction.LoadWallet)
        advanceUntilIdle()

        viewModel.onAction(
            WalletAction.SendTransaction(
                to = "0x123",
                amount = BigDecimal("500.00")
            )
        )

        val state = viewModel.uiState.value as WalletUiState.Success
        assertEquals("1,000.00", state.balance)
    }
}
```

### 3.2 Action/Event Testing

```kotlin
class ChatViewModelTest {

    @Test
    fun `should add message to list on send action`() = runTest {
        coEvery { sendMessageUseCase(any()) } returns Result.success(Unit)

        viewModel.onAction(ChatAction.SendMessage("Hello"))

        val state = viewModel.uiState.value as ChatUiState.Success
        assertEquals(1, state.messages.size)
        assertEquals("Hello", state.messages[0].content)
        assertEquals(MessageRole.USER, state.messages[0].role)
    }

    @Test
    fun `should append incoming websocket message`() = runTest {
        viewModel.onEvent(ChatEvent.WebSocketMessage("Hi there"))

        val state = viewModel.uiState.value as ChatUiState.Success
        assertEquals(1, state.messages.size)
        assertEquals("Hi there", state.messages[0].content)
        assertEquals(MessageRole.ASSISTANT, state.messages[0].role)
    }

    @Test
    fun `should clear error on dismiss action`() = runTest {
        viewModel.onAction(ChatAction.DismissError)

        val state = viewModel.uiState.value
        assertNull(state.error)
    }
}
```

---

## 4. Use Case Testing

```kotlin
class CreateTransactionUseCaseTest {

    @MockK private lateinit var walletRepository: WalletRepository
    @MockK private lateinit var transactionRepository: TransactionRepository
    @MockK private lateinit var notificationService: NotificationService

    private lateinit var useCase: CreateTransactionUseCase

    @BeforeEach
    fun setup() {
        MockKAnnotations.init(this)
        useCase = CreateTransactionUseCase(
            walletRepository, transactionRepository, notificationService
        )
    }

    @Test
    fun `should create transaction and debit wallet`() = runTest {
        val request = TransactionRequest(
            toAddress = "0xabc",
            amount = BigDecimal("100.00"),
            currency = "USDC"
        )

        coEvery { walletRepository.getBalance("USDC") } returns
            Result.success(BigDecimal("500.00"))
        coEvery { walletRepository.debit("USDC", BigDecimal("100.00")) } returns
            Result.success(Unit)
        coEvery { transactionRepository.create(request) } returns
            Result.success(Transaction(id = "tx_1", status = "pending"))

        val result = useCase(request)

        assertTrue(result.isSuccess)
        coVerifyOrder {
            walletRepository.getBalance("USDC")
            walletRepository.debit("USDC", BigDecimal("100.00"))
            transactionRepository.create(request)
        }
    }

    @Test
    fun `should reject transaction when insufficient balance`() = runTest {
        coEvery { walletRepository.getBalance("USDC") } returns
            Result.success(BigDecimal("50.00"))

        val request = TransactionRequest(
            toAddress = "0xabc",
            amount = BigDecimal("100.00"),
            currency = "USDC"
        )

        val result = useCase(request)

        assertTrue(result.isFailure)
        assertInstanceOf(
            InsufficientBalanceException::class.java,
            result.exceptionOrNull()
        )
        coVerify(exactly = 0) { transactionRepository.create(any()) }
    }
}
```

---

## 5. Repository Testing

### 5.1 Local + Remote Data Source

```kotlin
class ChatRepositoryImplTest {

    @MockK private lateinit var remoteDataSource: ChatRemoteDataSource
    @MockK private lateinit var localDataSource: ChatLocalDataSource
    @MockK private lateinit var networkChecker: NetworkChecker

    private lateinit var repository: ChatRepositoryImpl

    @BeforeEach
    fun setup() {
        MockKAnnotations.init(this)
        repository = ChatRepositoryImpl(
            remoteDataSource, localDataSource, networkChecker
        )
    }

    @Test
    fun `should return cached messages when offline`() = runTest {
        coEvery { networkChecker.isConnected() } returns false
        coEvery { localDataSource.getMessages("conv_1") } returns
            listOf(cachedMessage)

        val result = repository.getMessages("conv_1")

        assertTrue(result.isSuccess)
        assertEquals(listOf(cachedMessage), result.getOrNull())
        coVerify(exactly = 0) { remoteDataSource.getMessages(any()) }
    }

    @Test
    fun `should fetch remote and cache when online`() = runTest {
        coEvery { networkChecker.isConnected() } returns true
        coEvery { remoteDataSource.getMessages("conv_1") } returns
            Result.success(listOf(remoteMessage))
        coEvery { localDataSource.upsertMessages(any()) } just Runs

        val result = repository.getMessages("conv_1")

        assertTrue(result.isSuccess)
        coVerify { localDataSource.upsertMessages(listOf(remoteMessage)) }
    }

    @Test
    fun `should fallback to cache on remote failure`() = runTest {
        coEvery { networkChecker.isConnected() } returns true
        coEvery { remoteDataSource.getMessages("conv_1") } returns
            Result.failure(IOException("Timeout"))
        coEvery { localDataSource.getMessages("conv_1") } returns
            listOf(cachedMessage)

        val result = repository.getMessages("conv_1")

        assertTrue(result.isSuccess)
        assertEquals(listOf(cachedMessage), result.getOrNull())
    }
}
```

---

## 6. Network Testing (MockWebServer)

```kotlin
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.MockResponse

class ApiClientTest {

    private lateinit var server: MockWebServer
    private lateinit var apiClient: ApiClient

    @BeforeEach
    fun setup() {
        server = MockWebServer()
        server.start()

        apiClient = ApiClient(
            baseUrl = server.url("/").toString(),
            httpClient = OkHttpClient.Builder().build()
        )
    }

    @AfterEach
    fun tearDown() {
        server.shutdown()
    }

    @Test
    fun `should parse wallet response correctly`() = runTest {
        server.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "application/json")
                .setBody("""
                    {
                        "balance": "1500.00",
                        "currency": "USDC",
                        "address": "0xabc"
                    }
                """.trimIndent())
        )

        val result = apiClient.getWallet()

        assertTrue(result.isSuccess)
        assertEquals("1500.00", result.getOrNull()?.balance)

        val request = server.takeRequest()
        assertEquals("/wallet", request.path)
        assertEquals("GET", request.method)
    }

    @Test
    fun `should handle 401 unauthorized`() = runTest {
        server.enqueue(
            MockResponse()
                .setResponseCode(401)
                .setBody("""{"error": "token_expired"}""")
        )

        val result = apiClient.getWallet()

        assertTrue(result.isFailure)
        assertInstanceOf(AuthException::class.java, result.exceptionOrNull())
    }

    @Test
    fun `should handle 429 rate limit with retry`() = runTest {
        server.enqueue(
            MockResponse()
                .setResponseCode(429)
                .setHeader("Retry-After", "1")
                .setBody("""{"error": "rate_limited"}""")
        )
        server.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setBody("""{"balance": "100", "currency": "USDC", "address": "0x1"}""")
        )

        val result = apiClient.getWallet()

        assertTrue(result.isSuccess)
        assertEquals(2, server.requestCount)
    }

    @Test
    fun `should send correct headers`() = runTest {
        server.enqueue(MockResponse().setResponseCode(200).setBody("{}"))

        apiClient.getWallet()

        val request = server.takeRequest()
        assertEquals("Bearer test_token", request.getHeader("Authorization"))
        assertEquals("application/json", request.getHeader("Content-Type"))
    }
}
```

---

## 7. Room Testing

```kotlin
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider

class MessageDaoTest {

    private lateinit var database: AppDatabase
    private lateinit var messageDao: MessageDao

    @BeforeEach
    fun setup() {
        database = Room.inMemoryDatabaseBuilder(
            ApplicationProvider.getApplicationContext(),
            AppDatabase::class.java
        ).allowMainThreadQueries().build()

        messageDao = database.messageDao()
    }

    @AfterEach
    fun tearDown() {
        database.close()
    }

    @Test
    fun `should insert and retrieve messages`() = runTest {
        val messages = listOf(
            MessageEntity(id = "1", content = "Hello", role = "user", timestamp = 1000L),
            MessageEntity(id = "2", content = "Hi", role = "assistant", timestamp = 2000L)
        )

        messageDao.upsertAll(messages)
        val result = messageDao.getMessagesByConversation("conv_1")

        assertEquals(2, result.size)
        assertEquals("Hello", result[0].content)
        assertEquals("Hi", result[1].content)
    }

    @Test
    fun `should delete old messages beyond limit`() = runTest {
        val messages = (1..100).map {
            MessageEntity(
                id = "msg_$it",
                content = "Message $it",
                role = "user",
                timestamp = it.toLong()
            )
        }

        messageDao.upsertAll(messages)
        messageDao.deleteOlderThan(limit = 50)

        val result = messageDao.getMessagesByConversation("conv_1")
        assertEquals(50, result.size)
    }

    @Test
    fun `should use UPSERT for conflict resolution`() = runTest {
        val msg = MessageEntity(id = "1", content = "v1", role = "user", timestamp = 1000L)
        messageDao.upsertAll(listOf(msg))

        val updated = msg.copy(content = "v2")
        messageDao.upsertAll(listOf(updated))

        val result = messageDao.getById("1")
        assertEquals("v2", result?.content)
    }

    @Test
    fun `should observe message count as flow`() = runTest {
        messageDao.observeMessageCount("conv_1").test {
            assertEquals(0, awaitItem())

            messageDao.upsertAll(listOf(
                MessageEntity(id = "1", content = "Hi", role = "user", timestamp = 1L)
            ))
            assertEquals(1, awaitItem())

            cancelAndIgnoreRemainingEvents()
        }
    }
}
```

---

## 8. WebSocket Testing

```kotlin
class WebSocketManagerTest {

    private lateinit var webSocketManager: WebSocketManagerImpl
    private lateinit var mockWebSocket: MockWebSocket

    @BeforeEach
    fun setup() {
        mockWebSocket = MockWebSocket()
        webSocketManager = WebSocketManagerImpl(
            okHttpClient = OkHttpClient.Builder().build(),
            baseUrl = "wss://test.example.com/ws"
        )
    }

    @Test
    fun `should emit messages when received`() = runTest {
        webSocketManager.connect()

        mockWebSocket.onMessage("""{"type":"chat","content":"Hello"}""")

        webSocketManager.messages.test {
            val msg = awaitItem()
            assertEquals("Hello", msg.content)
            cancelAndIgnoreRemainingEvents()
        }
    }

    @Test
    fun `should reconnect on unexpected disconnect`() = runTest {
        webSocketManager.connect()
        mockWebSocket.onClosed(1000, "Normal")

        advanceTimeBy(1000) // First retry delay

        coVerify { webSocketManager.connect() }
    }

    @Test
    fun `should use exponential backoff for reconnection`() = runTest {
        webSocketManager.connect()

        // First disconnect
        mockWebSocket.onClosed(1000, "Normal")
        advanceTimeBy(1000) // 1s delay

        // Second disconnect
        mockWebSocket.onClosed(1000, "Normal")
        advanceTimeBy(2000) // 2s delay

        // Third disconnect
        mockWebSocket.onClosed(1000, "Normal")
        advanceTimeBy(4000) // 4s delay

        assertEquals(4, webSocketManager.connectionAttempts)
    }

    @Test
    fun `should send messages through websocket`() = runTest {
        webSocketManager.connect()

        webSocketManager.send(ChatMessage(content = "Hello"))

        val sent = mockWebSocket.lastSentMessage
        assertTrue(sent.contains("Hello"))
    }
}
```

---

## 9. Integration Testing

```kotlin
@RunWith(AndroidJUnit4::class)
class ChatIntegrationTest {

    private lateinit var database: AppDatabase
    private lateinit var apiClient: ApiClient
    private lateinit var repository: ChatRepositoryImpl
    private lateinit var viewModel: ChatViewModel

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()

        database = Room.databaseBuilder(context, AppDatabase::class.java, "test_db")
            .build()

        apiClient = ApiClient(
            baseUrl = "https://jsonplaceholder.typicode.com/",
            httpClient = OkHttpClient.Builder().build()
        )

        repository = ChatRepositoryImpl(
            remoteDataSource = ChatRemoteDataSourceImpl(apiClient),
            localDataSource = ChatLocalDataSourceImpl(database.messageDao()),
            networkChecker = NetworkCheckerImpl(context)
        )

        viewModel = ChatViewModel(
            sendMessageUseCase = SendMessageUseCase(repository),
            getMessagesUseCase = GetMessagesUseCase(repository)
        )
    }

    @After
    fun teardown() {
        database.close()
    }

    @Test
    fun fullChatFlow_sendMessage_persistsLocally() = runTest {
        viewModel.onAction(ChatAction.SendMessage("Hello AI"))

        val state = viewModel.uiState.value as ChatUiState.Success
        assertEquals(1, state.messages.size)

        val persisted = database.messageDao()
            .getMessagesByConversation("default")
        assertEquals(1, persisted.size)
        assertEquals("Hello AI", persisted[0].content)
    }
}
```

---

## 10. UI Testing (Compose)

### 10.1 Screen Testing

```kotlin
import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule

class ChatScreenTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    private val viewModel = mockk<ChatViewModel>(relaxed = true)

    @Test
    fun should_displayEmptyState_when_noMessages() {
        every { viewModel.uiState } returns mutableStateOf(
            ChatUiState.Success(messages = emptyList())
        )

        composeTestRule.setContent {
            ChatScreen(viewModel = viewModel)
        }

        composeTestRule.onNodeWithText("No messages yet").assertIsDisplayed()
    }

    @Test
    fun should_displayMessages_when_loaded() {
        val messages = listOf(
            ChatMessageUi(content = "Hello", isUser = true),
            ChatMessageUi(content = "Hi there!", isUser = false)
        )
        every { viewModel.uiState } returns mutableStateOf(
            ChatUiState.Success(messages = messages)
        )

        composeTestRule.setContent {
            ChatScreen(viewModel = viewModel)
        }

        composeTestRule.onNodeWithText("Hello").assertIsDisplayed()
        composeTestRule.onNodeWithText("Hi there!").assertIsDisplayed()
    }

    @Test
    fun should_callViewModel_when_messageSent() {
        every { viewModel.uiState } returns mutableStateOf(
            ChatUiState.Success(messages = emptyList())
        )

        composeTestRule.setContent {
            ChatScreen(viewModel = viewModel)
        }

        composeTestRule.onNodeWithTag("chat_input")
            .performTextInput("New message")

        composeTestRule.onNodeWithTag("send_button").performClick()

        verify { viewModel.onAction(ChatAction.SendMessage("New message")) }
    }

    @Test
    fun should_showLoading_when_sending() {
        every { viewModel.uiState } returns mutableStateOf(ChatUiState.Loading)

        composeTestRule.setContent {
            ChatScreen(viewModel = viewModel)
        }

        composeTestRule.onNodeWithTag("loading_indicator").assertIsDisplayed()
    }
}
```

### 10.2 Component Testing

```kotlin
class WalletCardTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun should_displayBalance_andCurrency() {
        composeTestRule.setContent {
            WalletCard(balance = "1,500.00", currency = "USDC")
        }

        composeTestRule.onNodeWithText("1,500.00").assertIsDisplayed()
        composeTestRule.onNodeWithText("USDC").assertIsDisplayed()
    }

    @Test
    fun should_callOnClick_when_tapped() {
        var clicked = false

        composeTestRule.setContent {
            WalletCard(
                balance = "100",
                currency = "USD",
                onClick = { clicked = true }
            )
        }

        composeTestRule.onNodeWithTag("wallet_card").performClick()
        assertTrue(clicked)
    }
}

class TransactionFormTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun should_showError_when_amountIsZero() {
        composeTestRule.setContent {
            TransactionForm(
                onConfirm = {},
                onCancel = {}
            )
        }

        composeTestRule.onNodeWithTag("amount_input")
            .performTextInput("0")
        composeTestRule.onNodeWithTag("confirm_button").performClick()

        composeTestRule.onNodeWithText("Amount must be greater than 0")
            .assertIsDisplayed()
    }

    @Test
    fun should_disableConfirm_when_formIsInvalid() {
        composeTestRule.setContent {
            TransactionForm(onConfirm = {}, onCancel = {})
        }

        composeTestRule.onNodeWithTag("confirm_button").assertIsNotEnabled()
    }
}
```

### 10.3 Navigation Testing

```kotlin
class NavigationTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun should_navigateToChat_when_walletCardClicked() {
        composeTestRule.setContent {
            AppNavHost(startDestination = "wallet")
        }

        composeTestRule.onNodeWithTag("wallet_card").performClick()

        composeTestRule.onNodeWithText("Chat").assertIsDisplayed()
    }

    @Test
    fun should_navigateBack_when_backButtonPressed() {
        composeTestRule.setContent {
            AppNavHost(startDestination = "chat")
        }

        composeTestRule.onNodeWithTag("back_button").performClick()

        composeTestRule.onNodeWithTag("wallet_card").assertIsDisplayed()
    }
}
```

---

## 11. Screenshot Testing

```kotlin
// Using Roborazzi
import com.github.takahirom.roborazzi.RobolectricTestRunner
import com.github.takahirom.roborazzi.captureRoboImage

@RunWith(RobolectricTestRunner::class)
class ScreenshotTests {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun chatScreen_emptyState() {
        composeTestRule.setContent {
            Theme {
                ChatScreen(uiState = ChatUiState.Success(emptyList()))
            }
        }
        composeTestRule.onRoot().captureRoboImage()
    }

    @Test
    fun chatScreen_withMessages() {
        composeTestRule.setContent {
            Theme {
                ChatScreen(
                    uiState = ChatUiState.Success(
                        messages = listOf(
                            MessageUi("Hello", isUser = true),
                            MessageUi("Hi there!", isUser = false)
                        )
                    )
                )
            }
        }
        composeTestRule.onRoot().captureRoboImage()
    }

    @Test
    fun walletCard_largeBalance() {
        composeTestRule.setContent {
            Theme {
                WalletCard(balance = "1,234,567.89", currency = "USDC")
            }
        }
        composeTestRule.onRoot().captureRoboImage()
    }
}
```

---

## 12. Accessibility Testing

```kotlin
class AccessibilityTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun sendButton_shouldHaveContentDescription() {
        composeTestRule.setContent {
            SendButton(onClick = {})
        }

        composeTestRule.onNodeWithContentDescription("Send message")
            .assertIsDisplayed()
    }

    @Test
    fun chatMessage_shouldHaveRole() {
        composeTestRule.setContent {
            ChatBubble(
                content = "Hello",
                isUser = true
            )
        }

        composeTestRule.onNodeWithText("Hello")
            .assertContentDescriptionEquals("You said: Hello")
    }

    @Test
    fun balance_shouldBeReadableByScreenReader() {
        composeTestRule.setContent {
            BalanceDisplay(balance = "1500.00", currency = "USDC")
        }

        composeTestRule.onNodeWithContentDescription(
            "Balance: 1,500.00 USDC"
        ).assertExists()
    }
}
```

---

## 13. Performance Testing

```kotlin
@RunWith(AndroidJUnit4::class)
class StartupBenchmark {

    @get:Rule
    val benchmarkRule = MacrobenchmarkRule()

    @Test
    fun startupCompilationNone() = startup(CompilationMode.None())

    @Test
    fun startupCompilationBaselineProfiles() = startup(
        CompilationMode.Partial(
            baselineProfileMode = BaselineProfileMode.Require
        )
    )

    private fun startup(compilationMode: CompilationMode) {
        benchmarkRule.measureRepeated(
            packageName = "com.nexusai.app",
            metrics = listOf(StartupTimingMetric()),
            iterations = 5,
            compilationMode = compilationMode,
            startupMode = StartupMode.COLD
        ) {
            pressHome()
            startActivityAndWait()
        }
    }
}

@RunWith(AndroidJUnit4::class)
class ScrollBenchmark {

    @get:Rule
    val benchmarkRule = MacrobenchmarkRule()

    @Test
    fun scrollChatList() {
        benchmarkRule.measureRepeated(
            packageName = "com.nexusai.app",
            metrics = listOf(FrameTimingMetric("P50", "P90", "P99")),
            iterations = 10
        ) {
            startActivityAndWait()
            device.findObject(By.res("chat_list")).scroll(Direction.DOWN, 3f)
        }
    }
}
```

---

## 14. Test Fixtures & Factories

```kotlin
object TestDataFactory {

    fun createWallet(
        balance: String = "1000.00",
        currency: String = "USDC"
    ) = Wallet(balance = BigDecimal(balance), currency = currency)

    fun createMessage(
        id: String = UUID.randomUUID().toString(),
        content: String = "Test message",
        role: MessageRole = MessageRole.USER,
        timestamp: Long = System.currentTimeMillis()
    ) = Message(id = id, content = content, role = role, timestamp = timestamp)

    fun createConversation(
        id: String = UUID.randomUUID().toString(),
        title: String = "Test Conversation",
        messageCount: Int = 5
    ) = Conversation(
        id = id,
        title = title,
        messages = (1..messageCount).map { createMessage() },
        createdAt = System.currentTimeMillis()
    )

    fun createApiError(
        code: Int = 500,
        message: String = "Server error"
    ) = ApiError(code = code, message = message)

    fun createSuccessResponse(body: String) =
        MockResponse()
            .setResponseCode(200)
            .setHeader("Content-Type", "application/json")
            .setBody(body)

    fun createErrorResponse(code: Int, body: String) =
        MockResponse()
            .setResponseCode(code)
            .setBody(body)
}

object Fakes {

    class FakeWalletRepository : WalletRepository {
        private var balance = BigDecimal("1000.00")

        override suspend fun getBalance(currency: String) =
            Result.success(balance)

        override suspend fun debit(currency: String, amount: BigDecimal): Result<Unit> {
            balance -= amount
            return Result.success(Unit)
        }

        override suspend fun credit(currency: String, amount: BigDecimal): Result<Unit> {
            balance += amount
            return Result.success(Unit)
        }

        fun setBalance(newBalance: BigDecimal) { balance = newBalance }
    }

    class FakeTokenStorage : TokenStorage {
        private var token: AuthToken? = null

        override suspend fun save(token: AuthToken) { this.token = token }
        override suspend fun get(): AuthToken? = token
        override suspend fun clear() { token = null }
    }
}
```

---

## 15. Test Utilities & Extensions

```kotlin
// Test extensions
fun <T> Result<T>.assertSuccess(block: (T) -> Unit = {}) {
    assertTrue(isSuccess, "Expected success but was failure: ${exceptionOrNull()?.message}")
    getOrNull()?.let(block)
}

fun <T> Result<T>.assertFailure(block: (Throwable) -> Unit = {}) {
    assertTrue(isFailure, "Expected failure but was success")
    exceptionOrNull()?.let(block)
}

fun runBlockingTest(block: suspend TestScope.() -> Unit) = runTest {
    TestCoroutineScheduler().apply {
        advanceTimeBy(0)
    }
    block()
}

// Custom matchers
fun hasDrawable(@DrawableRes resId: Int): Matcher<View> =
    object : BoundedMatcher<View, ImageView>(ImageView::class.java) {
        override fun matchesSafely(view: ImageView) =
            view.drawable?.constantState ==
                ContextCompat.getDrawable(view.context, resId)?.constantState

        override fun describeTo(description: Description) {
            description.appendText("with drawable resource: $resId")
        }
    }
```

---

## 16. Mock Strategies

| Strategy          | When to Use                      | Pros                    | Cons                     |
|-------------------|----------------------------------|-------------------------|--------------------------|
| MockK             | Complex interactions, interfaces | Flexible, expressive    | Slow, verbose            |
| Fake              | Simple interfaces, repositories  | Fast, realistic         | Must maintain fakes      |
| Manual Mock       | Framework classes, 3rd party     | Full control            | Extra code to maintain   |
| Spy               | Partial mocking                  | Calls real impl         | Can be fragile           |

```kotlin
// When to use each:
// MockK: when you need verify(), coEvery(), complex stubbing
// Fake: when you want fast, deterministic tests for repositories
// Manual: when mocking framework can't handle the class (e.g., Context)
```

---

## 17. Test Coverage (JaCoCo)

```kotlin
// build.gradle.kts
plugins {
    jacoco
}

tasks.withType<Test> {
    finalizedBy("jacocoTestReport")
}

tasks.register<JacocoReport>("jacocoTestReport") {
    dependsOn("testDebugUnitTest")

    reports {
        xml.required.set(true)
        html.required.set(true)
        csv.required.set(false)
    }

    sourceDirectories.setFrom("src/main/java")
    classDirectories.setFrom(
        fileTree("build/tmp/kotlin-classes/debug") {
            exclude("**/R.class", "**/R\$*.class", "**/BuildConfig.*")
        }
    )
    executionData.setFrom(
        fileTree("build/jacoco/testDebugUnitTest.exec")
    )
}
```

| Module              | Target | Current | Status  |
|---------------------|--------|---------|---------|
| Domain Use Cases    | 95%    | 92%     | ✅ Pass |
| Data Repositories   | 90%    | 88%     | ✅ Pass |
| ViewModels          | 85%    | 83%     | ✅ Pass |
| UI Composables      | 70%    | 65%     | ⚠️ Gap  |
| Overall             | 80%    | 78%     | ⚠️ Gap  |

---

## 18. CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          java-version: '17'
          distribution: 'temurin'

      - name: Unit Tests
        run: ./gradlew testDebugUnitTest

      - name: Code Coverage
        run: ./gradlew jacocoTestReport

      - name: Coverage Report
        uses: codecov/codecov-action@v3
        with:
          files: build/reports/jacoco/jacocoTestReport/jacocoTestReport.xml

      - name: Instrumented Tests
        uses: reactivecircus/android-emulator-runner@v2
        with:
          api-level: 34
          script: ./gradlew connectedAndroidTest

      - name: Lint
        run: ./gradlew lintDebug

      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: build/reports/
```

---

## 19. Test Naming Conventions

```
Pattern: should_[expected behavior]_when_[condition]

Examples:
  should_returnSuccess_when_validCredentials
  should_returnError_when_networkUnavailable
  should_showSnackbar_when_transactionFails
  should_emitLoading_thenSuccess_when_fetchingWallet
  should_notCallRepository_when_cacheIsValid
  should_reconnect_withBackoff_when_connectionDrops
  should_showEmptyState_when_noConversations
  should_navigateToChat_when_walletCardClicked
```

### Test File Organization

```
src/test/java/com/nexusai/
├── domain/
│   ├── usecase/
│   │   ├── SendMessageUseCaseTest.kt
│   │   └── GetWalletBalanceUseCaseTest.kt
│   └── model/
│       └── TransactionTest.kt
├── data/
│   ├── repository/
│   │   ├── ChatRepositoryImplTest.kt
│   │   └── WalletRepositoryImplTest.kt
│   ├── remote/
│   │   └── ApiClientTest.kt
│   └── local/
│       └── MessageDaoTest.kt
├── presentation/
│   ├── chat/
│   │   ├── ChatViewModelTest.kt
│   │   └── ChatScreenTest.kt (composeTest)
│   └── wallet/
│       ├── WalletViewModelTest.kt
│       └── WalletScreenTest.kt
└── testutil/
    ├── TestDataFactory.kt
    ├── Fakes.kt
    └── Extensions.kt
```

---

## 20. Test Debugging Tips

| Problem                    | Solution                                           |
|----------------------------|----------------------------------------------------|
| Flaky coroutine tests      | Use `UnconfinedTestDispatcher`                     |
| Race condition in Flow     | Use Turbine with `test { }`                        |
| Room on main thread        | `.allowMainThreadQueries()` in test only            |
| MockWebServer port clash   | Use `server.port(0)` for dynamic port               |
| Compose test timing        | Use `waitForIdle()` or `advanceUntilIdle()`         |
| Timeout in integration     | Increase timeout in `runTest { }` with `timeout()` |
| OOM in screenshot tests    | Limit bitmap size, use `Config.Quality.LOW`        |
| Navigation test flakiness  | Use `composeTestRule.onRoot().printToLog()`         |
