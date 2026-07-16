# 16. Testing

## Table of Contents

- [Testing Strategy](#testing-strategy)
- [Unit Testing](#unit-testing)
- [ViewModel Testing](#viewmodel-testing)
- [Use Case Testing](#use-case-testing)
- [Repository Testing](#repository-testing)
- [Network Testing](#network-testing)
- [CoreData Testing](#coredata-testing)
- [WebSocket Testing](#websocket-testing)
- [Integration Testing](#integration-testing)
- [UI Testing](#ui-testing)
- [Screen Testing](#screen-testing)
- [Component Testing](#component-testing)
- [Navigation Testing](#navigation-testing)
- [Screenshot Testing](#screenshot-testing)
- [Accessibility Testing](#accessibility-testing)
- [Performance Testing](#performance-testing)
- [Security Testing](#security-testing)
- [Test Fixtures](#test-fixtures)
- [Test Utilities](#test-utilities)
- [Mock Strategies](#mock-strategies)
- [Test Data Management](#test-data-management)
- [Test Coverage](#test-coverage)
- [CI/CD Integration](#cicd-integration)
- [Test Reporting & Debugging](#test-reporting--debugging)
- [Test Patterns](#test-patterns)
- [Test Naming & Organization](#test-naming--organization)

---

## Testing Strategy

### The Testing Pyramid

```
        ╱╲
       ╱  ╲        E2E / UI Tests (10%) — Slow, few, high confidence
      ╱────╲
     ╱      ╲      Integration Tests (20%) — Medium speed
    ╱────────╲
   ╱          ╲    Unit Tests (70%) — Fast, many, low cost
  ╱────────────╲
```

| Tier       | Count | Speed    | Confidence |
|------------|-------|----------|------------|
| Unit       | 70%   | < 1ms    | Component  |
| Integration| 20%   | 10-100ms | Interface  |
| UI / E2E   | 10%   | 1-10s    | Full flow  |

---

## Unit Testing

```swift
import XCTest
@testable import NexusApp

final class StringValidatorTests: XCTestCase {
    var sut: StringValidator!

    override func setUp() { super.setUp(); sut = StringValidator() }
    override func tearDown() { sut = nil; super.tearDown() }

    func testValidateEmail_whenValid_returnsTrue() {
        XCTAssertTrue(sut.validateEmail("user@example.com"))
    }

    func testValidateEmail_whenMissingAt_returnsFalse() {
        XCTAssertFalse(sut.validateEmail("userexample.com"))
    }

    func testValidateEmail_performance() {
        measure {
            for _ in 0..<10_000 { _ = sut.validateEmail("user@example.com") }
        }
    }
}
```

---

## ViewModel Testing

```swift
@MainActor
final class ProfileViewModelTests: XCTestCase {
    var sut: ProfileViewModel!
    var mockRepo: MockProfileRepository!

    override func setUp() {
        super.setUp()
        mockRepo = MockProfileRepository()
        sut = ProfileViewModel(repository: mockRepo)
    }

    func testLoadProfile_whenSuccess_updatesProfile() async {
        mockRepo.profileResult = .success(Profile(id: "1", name: "John", email: "j@t.com"))
        await sut.loadProfile()
        XCTAssertEqual(sut.profile?.name, "John")
        XCTAssertFalse(sut.isLoading)
    }

    func testLoadProfile_whenFailure_setsError() async {
        mockRepo.profileResult = .failure(.networkError(.noConnection))
        await sut.loadProfile()
        XCTAssertNotNil(sut.errorMessage)
    }

    func testUpdateProfile_whenSuccess_updatesLocally() async {
        sut.profile = Profile.mock
        mockRepo.updateResult = .success(Profile.mock.copy(name: "Jane"))
        await sut.updateProfile(name: "Jane")
        XCTAssertEqual(sut.profile?.name, "Jane")
    }
}
```

---

## Use Case Testing

```swift
final class SendMoneyUseCaseTests: XCTestCase {
    var sut: SendMoneyUseCase!
    var mockWallet: MockWalletRepository!
    var mockTx: MockTransactionRepository!

    override func setUp() {
        super.setUp()
        mockWallet = MockWalletRepository()
        mockTx = MockTransactionRepository()
        sut = SendMoneyUseCase(walletRepository: mockWallet, transactionRepository: mockTx)
    }

    func testExecute_whenValid_succeeds() async {
        mockWallet.balanceResult = .success(1000.0)
        mockTx.sendResult = .success(Transaction.mock(id: "tx-1"))
        let result = await sut.execute(amount: 50, recipientId: "u-2", currency: .usd)
        if case .success(let tx) = result { XCTAssertEqual(tx.id, "tx-1") } else { XCTFail() }
    }

    func testExecute_whenInsufficientBalance_returnsError() async {
        mockWallet.balanceResult = .success(10.0)
        let result = await sut.execute(amount: 50, recipientId: "u-2", currency: .usd)
        if case .failure(let e) = result { XCTAssertEqual(e, .insufficientBalance) } else { XCTFail() }
    }
}
```

---

## Repository Testing

### CoreData (In-Memory)

```swift
final class LocalUserRepositoryTests: XCTestCase {
    var sut: LocalUserRepository!
    var container: NSPersistentContainer!

    override func setUp() {
        super.setUp()
        container = NSPersistentContainer.inMemoryContainer()
        sut = LocalUserRepository(context: container.viewContext)
    }

    func testSave_andFetch() async throws {
        try await sut.save(User.mock)
        XCTAssertEqual(try await sut.fetch(id: User.mock.id)?.name, User.mock.name)
    }

    func testDelete_removes() async throws {
        try await sut.save(User.mock)
        try await sut.delete(id: User.mock.id)
        XCTAssertNil(try await sut.fetch(id: User.mock.id))
    }
}
```

### Remote

```swift
final class RemoteUserRepositoryTests: XCTestCase {
    var sut: RemoteUserRepository!
    var mockClient: MockNetworkClient!

    override func setUp() { super.setUp(); mockClient = MockNetworkClient(); sut = RemoteUserRepository(networkClient: mockClient) }

    func testFetch_whenSuccess() async throws {
        mockClient.responseData = try JSONEncoder().encode(User.mock)
        XCTAssertEqual(try await sut.fetch(id: "1")?.id, User.mock.id)
    }

    func testFetch_when404_returnsNil() async throws {
        mockClient.error = APIError.notFound
        XCTAssertNil(try await sut.fetch(id: "999"))
    }
}
```

---

## Network Testing

### URLProtocol Mock

```swift
final class MockURLProtocol: URLProtocol {
    static var mockData: Data?
    static var mockError: Error?
    static var mockStatusCode: Int = 200
    static var capturedRequests: [URLRequest] = []

    override class func canInit(with request: URLRequest) -> Bool { capturedRequests.append(request); return true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }
    override func startLoading() {
        if let e = MockURLProtocol.mockError { client?.urlProtocol(self, didFailWithError: e); return }
        let r = HTTPURLResponse(url: request.url!, statusCode: MockURLProtocol.mockStatusCode, httpVersion: nil, headerFields: nil)!
        client?.urlProtocol(self, didReceive: r, cacheStoragePolicy: .notAllowed)
        if let d = MockURLProtocol.mockData { client?.urlProtocol(self, didLoad: d) }
        client?.urlProtocolDidFinishLoading(self)
    }
    override func stopLoading() {}
    static func reset() { mockData = nil; mockError = nil; mockStatusCode = 200; capturedRequests = [] }
}
```

```swift
final class APIClientTests: XCTestCase {
    var sut: APIClient!

    override func setUp() {
        super.setUp()
        let c = URLSessionConfiguration.ephemeral; c.protocolClasses = [MockURLProtocol.self]
        sut = APIClient(session: URLSession(configuration: c))
    }
    override func tearDown() { MockURLProtocol.reset(); super.tearDown() }

    func testFetchUsers_whenSuccess() async throws {
        MockURLProtocol.mockData = try JSONEncoder().encode(APIResponse(data: [User.mock], meta: .init(total: 1)))
        let r: [User] = try await sut.fetch(path: "/users")
        XCTAssertEqual(r.count, 1)
    }

    func testFetchUsers_when401() async throws {
        MockURLProtocol.mockStatusCode = 401; MockURLProtocol.mockError = APIError.unauthorized
        do { let _: [User] = try await sut.fetch(path: "/users"); XCTFail() }
        catch let e as APIError { XCTAssertEqual(e, .unauthorized) }
    }
}
```

---

## CoreData Testing

### In-Memory Container

```swift
extension NSPersistentContainer {
    static func inMemoryContainer() -> NSPersistentContainer {
        let d = NSPersistentStoreDescription(); d.type = NSInMemoryStoreType
        let c = NSPersistentContainer(name: "NexusModel",
            managedObjectModel: NSManagedObjectModel.mergedModel(from: [Bundle(for: AppDelegate.self)])!)
        c.persistentStoreDescriptions = [d]
        c.loadPersistentStores { _, error in if let e = error { fatalError("\(e)") } }
        c.viewContext.automaticallyMergesChangesFromParent = true
        return c
    }
}
```

```swift
final class CoreDataTransactionTests: XCTestCase {
    var container: NSPersistentContainer!
    var context: NSManagedObjectContext!

    override func setUp() { super.setUp(); container = NSPersistentContainer.inMemoryContainer(); context = container.viewContext }

    func testFetch_sortedByDate() throws {
        let tx1 = TransactionEntity(context: context); tx1.id = "1"; tx1.date = Date(timeIntervalSince1970: 1000)
        let tx2 = TransactionEntity(context: context); tx2.id = "2"; tx2.date = Date(timeIntervalSince1970: 2000)
        try context.save()
        let r: [TransactionEntity] = try context.fetch(TransactionEntity.fetchRequest().sorted(by: "date", ascending: false))
        XCTAssertEqual(r[0].id, "2")
    }

    func testBatchDelete() throws {
        for i in 1...100 { let t = TransactionEntity(context: context); t.id = "\(i)"; t.amount = Double(i) }
        try context.save()
        try context.execute(NSBatchDeleteRequest(fetchRequest: TransactionEntity.fetchRequest() as! NSFetchRequest<NSFetchRequestResult>))
        XCTAssertEqual(try context.count(for: TransactionEntity.fetchRequest()), 0)
    }
}
```

---

## WebSocket Testing

```swift
final class MockWebSocket: WebSocketProtocol {
    var onConnect: (() -> Void)?
    var onDisconnect: ((Error?) -> Void)?
    var onText: ((WebSocketMessage) -> Void)?
    var sentMessages: [String] = []
    var isConnected = false

    func connect() { isConnected = true; onConnect?() }
    func disconnect() { isConnected = false; onDisconnect?(nil) }
    func send(_ m: String) { sentMessages.append(m) }
    func simulateReceive(_ t: String) { onText?(.text(t)) }
    func simulateDisconnect(error: Error) { isConnected = false; onDisconnect?(error) }
}

final class ChatViewModelWebSocketTests: XCTestCase {
    var sut: ChatViewModel!
    var ws: MockWebSocket!

    override func setUp() { ws = MockWebSocket(); sut = ChatViewModel(webSocket: ws) }

    func testConnect_setsState() { sut.connect(); XCTAssertTrue(ws.isConnected) }
    func testSendMessage_sendsViaSocket() { sut.connect(); sut.sendMessage("Hi"); XCTAssertEqual(ws.sentMessages.count, 1) }
    func testReceive_addsToMessages() {
        sut.connect(); ws.simulateReceive(#"{"type":"message","content":"Hello","sender":"u2"}"#)
        XCTAssertEqual(sut.messages.first?.content, "Hello")
    }
}
```

---

## Integration Testing

```swift
final class TransactionFlowIntegrationTests: XCTestCase {
    func testCreateTransaction_savesLocalAndSyncsRemote() async throws {
        let container = NSPersistentContainer.inMemoryContainer()
        let local = LocalTransactionRepository(context: container.viewContext)
        let mockAPI = MockAPIClient()
        mockAPI.responseData = try JSONEncoder().encode(Transaction.mock(id: "tx-1"))
        let remote = RemoteTransactionRepository(apiClient: mockAPI)
        let ws = MockWebSocket()
        let coord = TransactionCoordinator(repository: CombinedTransactionRepository(local: local, remote: remote), webSocket: ws)

        let tx = try await coord.createTransaction(CreateTransactionRequest(amount: 100, recipientId: "u-2", currency: .usd))
        XCTAssertEqual(tx.id, "tx-1")
        XCTAssertTrue(ws.sentMessages.contains { $0.contains("tx-1") })
    }
}
```

---

## UI Testing

```swift
final class LoginFlowUITests: XCTestCase {
    var app: XCUIApplication!

    override func setUp() { continueAfterFailure = false; app = XCUIApplication(); app.launchArguments += ["--uitesting"]; app.launch() }

    func testLogin_validCredentials_navigatesHome() {
        app.textFields["login_email_field"].tap(); app.textFields["login_email_field"].typeText("u@t.com")
        app.secureTextFields["login_password_field"].tap(); app.secureTextFields["login_password_field"].typeText("Pass123!")
        app.buttons["login_submit_button"].tap()
        XCTAssertTrue(app.otherElements["home_screen"].waitForExistence(timeout: 5))
    }

    func testLogin_invalidCredentials_showsError() {
        app.textFields["login_email_field"].tap(); app.textFields["login_email_field"].typeText("bad@t.com")
        app.secureTextFields["login_password_field"].tap(); app.secureTextFields["login_password_field"].typeText("wrong")
        app.buttons["login_submit_button"].tap()
        XCTAssertTrue(app.alerts["login_error_alert"].waitForExistence(timeout: 5))
    }
}
```

---

## Screen Testing

### Page Object Pattern

```swift
struct LoginScreen {
    let app: XCUIApplication
    var emailField: XCUIElement { app.textFields["login_email_field"] }
    var passwordField: XCUIElement { app.secureTextFields["login_password_field"] }
    var loginButton: XCUIElement { app.buttons["login_submit_button"] }

    @discardableResult
    func login(email: String, password: String) -> HomeScreen {
        emailField.tap(); emailField.typeText(email)
        passwordField.tap(); passwordField.typeText(password)
        loginButton.tap()
        return HomeScreen(app: app)
    }
}

struct HomeScreen {
    let app: XCUIApplication
    var feedTable: XCUIElement { app.tables["feed_table"] }
    func navigateToProfile() -> ProfileScreen { app.tabBars.buttons["Profile"].tap(); return ProfileScreen(app: app) }
}
```

---

## Component Testing

### Accessibility Identifiers

```swift
enum AccessibilityIdentifier {
    enum Login { static let emailField = "login_email_field"; static let passwordField = "login_password_field"; static let submitButton = "login_submit_button" }
    enum Home { static let feedTable = "home_feed_table"; static let profileTab = "home_tab_profile" }
}

struct LoginView: View {
    var body: some View {
        VStack {
            TextField("Email", text: $email).accessibilityIdentifier(AccessibilityIdentifier.Login.emailField)
            SecureField("Password", text: $password).accessibilityIdentifier(AccessibilityIdentifier.Login.passwordField)
            Button("Login") { login() }.accessibilityIdentifier(AccessibilityIdentifier.Login.submitButton)
        }
    }
}
```

---

## Navigation Testing

```swift
final class NavigationTests: XCTestCase {
    var app: XCUIApplication!
    override func setUp() { app = XCUIApplication(); app.launch() }

    func testTabBar_switchesTabs() {
        app.tabBars.firstMatch.buttons["Feed"].tap()
        XCTAssertTrue(app.tables["feed_table"].waitForExistence(timeout: 2))
        app.tabBars.firstMatch.buttons["Profile"].tap()
        XCTAssertTrue(app.otherElements["profile_screen"].waitForExistence(timeout: 2))
    }

    func testDeepLink_opensScreen() {
        app.launchEnvironment["DEEP_LINK"] = "nexus://profile/user-123"
        XCTAssertTrue(app.otherElements["profile_screen"].waitForExistence(timeout: 5))
    }
}
```

---

## Screenshot Testing

```swift
import SnapshotTesting

final class ComponentSnapshotTests: XCTestCase {
    func testLoginView_lightMode() {
        assertSnapshot(matching: LoginView(), as: .image(layout: .device(config: .iPhone13)))
    }
    func testLoginView_darkMode() {
        assertSnapshot(matching: LoginView().environment(\.colorScheme, .dark), as: .image(layout: .device(config: .iPhone13)))
    }
    func testTransactionCell_states() {
        for (name, state) in [("pending", TransactionCell.State.pending), ("completed", .completed), ("failed", .failed)] {
            assertSnapshot(matching: TransactionCell(state: state), as: .image, identifier: name)
        }
    }
}
```

---

## Accessibility Testing

```swift
final class AccessibilityTests: XCTestCase {
    var app: XCUIApplication!
    override func setUp() { app = XCUIApplication(); app.launch() }

    func testLogin_hasVoiceOverLabels() {
        XCTAssertTrue(app.textFields["login_email_field"].isAccessibilityElement)
        XCTAssertTrue(app.secureTextFields["login_password_field"].isAccessibilityElement)
        XCTAssertTrue(app.buttons["login_submit_button"].isAccessibilityElement)
    }

    func testLogin_supportsDynamicType() {
        app.launchArguments += ["--UITestingDynamicType"]; app.launch()
        XCTAssertTrue(app.textFields["login_email_field"].isHittable)
    }
}
```

---

## Performance Testing

```swift
final class PerformanceTests: XCTestCase {
    func testStartup() { measure(metrics: [XCTApplicationLaunchMetric()]) { XCUIApplication().launch() } }

    func testScrollPerformance() {
        let app = XCUIApplication(); app.launch()
        measure(metrics: [XCTCPUMetric(), XCTMemoryMetric()]) {
            app.tables["feed_table"].swipeUp(); app.tables["feed_table"].swipeDown()
        }
    }

    func testImageDecoding() {
        let d = UIImage(named: "large_photo")!.pngData()!
        measure(metrics: [XCTCPUMetric()]) { _ = UIImage(data: d) }
    }
}
```

---

## Security Testing

```swift
final class SecurityTests: XCTestCase {
    func testBiometricAuth_whenDeviceLocked() async {
        let ctx = MockLAContext(); ctx.canEvaluatePolicyResult = false
        ctx.evaluatePolicyError = LAError(.deviceLocked)
        let sut = BiometricAuthService(context: ctx)
        if case .failure(let e) = await sut.authenticate() { XCTAssertEqual(e, .deviceLocked) }
    }

    func testKeychain_savesAndRetrieves() throws {
        let kc = KeychainStore(service: "com.nexus.test")
        try kc.save("token", forKey: "auth_token")
        XCTAssertEqual(try kc.fetch("auth_token"), "token")
        try kc.delete("auth_token"); XCTAssertNil(try? kc.fetch("auth_token"))
    }
}
```

---

## Test Fixtures

```swift
struct UserFactory {
    static func build(id: String = UUID().uuidString, name: String = "John") -> User { User(id: id, name: name, email: "\(name.lowercased())@t.com") }
    static func buildArray(count: Int) -> [User] { (0..<count).map { build(id: "\($0)", name: "User \($0)") } }
}

struct TransactionFactory {
    static func build(amount: Double = 100, status: TransactionStatus = .completed) -> Transaction {
        Transaction(id: UUID().uuidString, amount: amount, currency: .usd, status: status, date: Date(), recipientId: "r-1")
    }
}
```

---

## Test Utilities

```swift
extension XCTestCase {
    func wait(for condition: () -> Bool, timeout: TimeInterval = 5, file: StaticString = #file, line: UInt = #line) {
        let deadline = Date().addingTimeInterval(timeout)
        while Date() < deadline { if condition() { return }; RunLoop.current.run(until: Date().addingTimeInterval(0.1)) }
        XCTFail("Condition not met", file: file, line: line)
    }
}

extension String { static func randomEmail() -> String { "\(UUID().uuidString.lowercased())@t.com" } }
func XCTAssertNear(_ v: Double, _ e: Double, accuracy: Double = 0.01, file: StaticString = #file, line: UInt = #line) {
    XCTAssertEqual(v, e, accuracy: accuracy, file: file, line: line)
}
```

---

## Mock Strategies

```swift
// Protocol-based
protocol UserRepositoryProtocol { func fetch(id: String) async throws -> User?; func save(_ user: User) async throws }

class MockUserRepository: UserRepositoryProtocol {
    var users: [String: User] = [:]; var fetchResult: Result<User?, Error> = .success(nil)
    func fetch(id: String) async throws -> User? { try fetchResult.get() }
    func save(_ user: User) async throws { users[user.id] = user }
}

// Spy
class SpyUserRepository: MockUserRepository {
    var fetchCallCount = 0; var fetchIds: [String] = []
    override func fetch(id: String) async throws -> User? { fetchCallCount += 1; fetchIds.append(id); return try await super.fetch(id: id) }
}
```

---

## Test Data Management

```swift
enum TestJSON {
    static func decode<T: Decodable>(_ type: T.Type, named name: String) throws -> T {
        guard let url = Bundle(for: BundleToken.self).url(forResource: name, withExtension: "json") else { throw TestError.notFound }
        return try JSONDecoder.apiDecoder.decode(type, from: Data(contentsOf: url))
    }
}

struct TestDatabaseManager {
    static func createInMemoryContainer() -> NSPersistentContainer {
        let c = NSPersistentContainer.inMemoryContainer()
        for i in 1...5 { let u = UserEntity(context: c.viewContext); u.id = "user-\(i)"; u.name = "User \(i)" }
        for i in 1...20 { let t = TransactionEntity(context: c.viewContext); t.id = "tx-\(i)"; t.amount = Double(i * 10) }
        try? c.viewContext.save()
        return c
    }
}
```

---

## Test Coverage

```bash
xcodebuild test -project NexusApp.xcodeproj -scheme NexusApp \
  -destination 'platform=iOS Simulator,name=iPhone 15' \
  -enableCodeCoverage YES -resultBundlePath ./TestResults.xcresult

xcrun xccov view --report ./TestResults.xcresult
```

| Module              | Target |
|---------------------|--------|
| Core Business Logic | > 90%  |
| ViewModels          | > 85%  |
| Repositories        | > 80%  |
| **Overall**         | > 80%  |

---

## CI/CD Integration

```yaml
name: iOS Tests
on:
  push: { branches: [main] }
  pull_request: { branches: [main] }
jobs:
  test:
    runs-on: macos-14
    strategy:
      matrix:
        destination:
          - 'platform=iOS Simulator,name=iPhone 15,OS=17.2'
          - 'platform=iOS Simulator,name=iPad Pro (11-inch),OS=17.2'
    steps:
      - uses: actions/checkout@v4
      - run: sudo xcode-select -s /Applications/Xcode_15.2.app
      - name: Unit Tests
        run: xcodebuild test -project NexusApp.xcodeproj -scheme NexusApp -destination '${{ matrix.destination }}' -enableCodeCoverage YES -resultBundlePath ./results/unit.xcresult
      - name: UI Tests
        run: xcodebuild test -project NexusApp.xcodeproj -scheme NexusAppUITests -destination '${{ matrix.destination }}' -resultBundlePath ./results/ui.xcresult
      - uses: actions/upload-artifact@v4
        if: always()
        with: { name: results-${{ matrix.destination }}, path: ./results/ }
```

---

## Test Reporting & Debugging

```bash
xcpretty --report junit -o ./test-results.xml < build.log
```

```swift
func testFetchUser() async throws {
    let result = try await repo.fetch(id: "1")
    dump(result)  // Debug output
    let user = try XCTUnwrap(result)  // Crash with clear message
    XCTAssertEqual(user.name, "John")
}
```

---

## Test Patterns

### AAA (Arrange, Act, Assert)

```swift
func testCalculateTotal_withTax() {
    let calc = TaxCalculator(taxRate: 0.1)                          // Arrange
    let items = [CartItem(price: 100, quantity: 2)]                  // Arrange
    XCTAssertEqual(calc.calculateTotal(items: items), 220.0)        // Act + Assert
}
```

### Given-When-Then

```swift
func testTransfer_sufficientBalance() {
    let wallet = Wallet(balance: 1000)                               // Given
    TransferUseCase(wallet: wallet).execute(amount: 200, to: "r-1")  // When
    XCTAssertEqual(wallet.balance, 800)                              // Then
}
```

---

## Test Naming & Organization

```
test_<method>_<scenario>_<expectedResult>
test_validateEmail_validAddress_returnsTrue
test_loadProfile_networkFailure_showsError
```

### File Structure

```
NexusAppTests/
├── Unit/
│   ├── ViewModel/    (LoginViewModelTests, FeedViewModelTests)
│   ├── UseCase/      (SendMoneyUseCaseTests)
│   ├── Repository/   (LocalUserRepositoryTests, RemoteUserRepositoryTests)
│   └── Network/      (APIClientTests)
├── Integration/      (TransactionFlowTests)
├── UI/               (LoginFlowUITests, NavigationUITests)
├── Screenshot/       (ComponentSnapshotTests)
├── Performance/      (ScrollPerformanceTests)
├── Fixtures/         (*.json)
└── Helpers/          (MockFactory, TestExtensions)
```

Mirror the source structure:
```
NexusApp/Features/Login/LoginViewModel.swift
  ↔ NexusAppTests/Unit/ViewModel/LoginViewModelTests.swift
```
