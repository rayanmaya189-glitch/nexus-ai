# 17 — Performance

## 1. Performance Targets

| Metric                  | Target       | Budget     | Critical  |
|-------------------------|-------------|------------|-----------|
| Cold Start              | <1.0s       | <2.0s      | >3.0s     |
| Warm Start              | <0.5s       | <1.0s      | >2.0s     |
| Hot Start               | <0.2s       | <0.5s      | >1.0s     |
| Frame Time (P99)        | <16ms       | <32ms      | >50ms     |
| Frame Time (P50)        | <8ms        | <16ms      | >32ms     |
| Jank Rate               | <1%         | <3%        | >5%       |
| APK Size                | <20MB       | <50MB      | >80MB     |
| AAB Size                | <15MB       | <40MB      | >60MB     |
| Memory (Heap P95)       | <80MB       | <128MB     | >200MB    |
| Memory (Heap P50)       | <40MB       | <64MB      | >100MB    |
| Network Latency (P50)   | <200ms      | <500ms     | >1000ms   |
| Network Latency (P99)   | <1000ms     | <2000ms    | >5000ms   |
| Battery Drain (per hr)  | <2%         | <5%        | >10%      |
| App Size Download       | <15MB       | <30MB      | >50MB     |

```
Performance Budget Dashboard:
┌─────────────────────────────────────────────────────────────┐
│  STARTUP        ████████░░░░  820ms    ✅ (target: <1s)   │
│  FRAME P99      ████░░░░░░░░  14.2ms   ✅ (target: <16ms) │
│  MEMORY P95     ██████████░░  112MB    ⚠️ (target: <80MB)  │
│  APK SIZE       ██████░░░░░░  18MB     ✅ (target: <20MB)  │
│  NETWORK P50    ███████░░░░░  340ms    ✅ (target: <500ms) │
│  BATTERY/HR     ███░░░░░░░░░  1.8%     ✅ (target: <2%)    │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. Startup Optimization

### 2.1 Cold / Warm / Hot Start

```
Cold Start (process + activity):
┌────────────────────────────────────────────────────────────┐
│ Fork │ ClassLoad │ Application │ inflate │ first frame     │
│  5ms │   200ms   │    300ms    │  200ms  │    50ms         │
│◄──────────── Total: ~800ms ───────────────────────────────►│
└────────────────────────────────────────────────────────────┘

Warm Start (activity recreate):
┌────────────────────────────────────────────────────────────┐
│ inflate │ first frame                                     │
│  200ms  │   50ms                                          │
│◄──── Total: ~250ms ───────────────────────────────────────►│
└────────────────────────────────────────────────────────────┘

Hot Start (resume):
┌────────────────────────────────────────────────────────────┐
│ first frame                                              │
│   50ms                                                   │
│◄ Total: ~50ms ───────────────────────────────────────────►│
└────────────────────────────────────────────────────────────┘
```

### 2.2 App Startup Library

```kotlin
// build.gradle.kts
implementation("androidx.startup:startup-runtime:1.1.1")

class AnalyticsInitializer : Initializer<AnalyticsService> {
    override fun create(context: Context): AnalyticsService {
        return AnalyticsService.init(context)
    }

    override fun dependencies(): List<Class<out Initializer<*>>> {
        return emptyList() // No dependencies
    }
}

// AndroidManifest.xml
<provider
    android:name="androidx.startup.InitializationProvider"
    android:authorities="${applicationId}.androidx-startup"
    android:exported="false">
    <meta-data
        android:name="com.nexusai.AnalyticsInitializer"
        android:value="androidx.startup" />
</provider>
```

### 2.3 Defer Non-Critical Initialization

```kotlin
class NexusApplication : Application() {
    override fun onCreate() {
        super.onCreate()

        // Critical: init immediately
        TokenManager.init(this)
        CrashReporter.init(this)

        // Defer: init on first frame
        Handler(Looper.getMainLooper()).post {
            AnalyticsService.init(this)
            FeatureFlags.init(this)
        }

        // Defer: init on idle
        Looper.myQueue().addIdleHandler {
            ImageLoader.init(this)
            false // Don't keep the idle handler
        }
    }
}
```

---

## 3. Baseline Profiles

```
┌──────────────────────────────────────────────────────────────┐
│                    Baseline Profile Flow                     │
│                                                              │
│  App Launch ──► Measure Hot Methods ──► Generate Profile     │
│                        │                      │              │
│                        ▼                      ▼              │
│              Systrace/Perfetto        baseline-prof.txt      │
│                        │                      │              │
│                        ▼                      ▼              │
│              Identify Hot Paths     Ship with APK/AAB       │
│                                             │               │
│                                             ▼               │
│                                    ART pre-compiles on       │
│                                    first install             │
└──────────────────────────────────────────────────────────────┘
```

### Generate Baseline Profile

```kotlin
// build.gradle.kts
plugins {
    id("androidx.baselineprofile")
}

dependencies {
    baselineProfile(project(":benchmarks"))
}

// In benchmark module
@RunWith(AndroidJUnit4::class)
class BaselineProfileGenerator {

    @get:Rule
    val baselineRule = BaselineProfileRule()

    @Test
    fun generate() {
        baselineRule.collect(
            packageName = "com.nexusai.app",
            profileBlock = {
                pressHome()
                startActivityAndWait()

                // Navigate through critical paths
                device.findObject(By.res("wallet_card")).click()
                device.waitForIdle()

                device.findObject(By.res("chat_button")).click()
                device.waitForIdle()

                device.findObject(By.res("chat_input")).text = "Hello"
                device.findObject(By.res("send_button")).click()
                device.waitForIdle()
            }
        )
    }
}

// Generate with:
// ./gradlew :benchmark:pixel6Api33BaselineProfileGenerator
```

### Baseline Profile Content

```text
# baseline-prof.txt (auto-generated, subset)
HSPLcom/nexusai/app/MainActivity;->onCreate(Landroid/os/Bundle;)V
HSPLcom/nexusai/app/ui/chat/ChatScreen;->Content(Landroidx/compose/runtime/Composer;I)V
HSPLcom/nexusai/app/data/remote/ApiClient;->getWallet()Lkotlinx/coroutines/flow/Flow;
```

---

## 4. R8 Optimization

### 4.1 R8 Rules

```proguard
# proguard-rules.pro

# ─── Keep rules ─────────────────────────────────────────────
# Keep data classes for JSON serialization
-keep class com.nexusai.data.model.** { *; }
-keepclassmembers class com.nexusai.data.model.** {
    <fields>;
}

# Keep Retrofit interfaces
-keep,allowobfuscation interface * extends retrofit2.http.* {
    @retrofit2.http.* <methods>;
}

# Keep Kotlin serialization
-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.AnnotationsKt

# Keep Room entities
-keep class * extends androidx.room.RoomDatabase
-keep @androidx.room.Entity class *

# ─── Optimization flags ─────────────────────────────────────
# Enable aggressive optimizations
-optimizationpasses 5
-allowaccessmodification
-repackageclasses ''

# Remove logging in release
-assumenosideeffects class android.util.Log {
    public static int v(...);
    public static int d(...);
    public static int i(...);
}

# Remove Timber debug trees
-assumenosideeffects class timber.log.Timber {
    public static void d(...);
    public static void v(...);
}
```

### 4.2 R8 Impact

| Metric              | Before R8 | After R8  | Reduction |
|---------------------|-----------|-----------|-----------|
| Method Count        | 45,000    | 28,000    | 38%       |
| APK Size            | 32MB      | 18MB      | 44%       |
| String Pool         | 120KB     | 65KB      | 46%       |
| DEX File Size       | 8MB       | 4.5MB     | 44%       |

---

## 5. Build Optimization

```kotlin
// gradle.properties
org.gradle.parallel=true
org.gradle.caching=true
org.gradle.configureondemand=true
org.gradle.daemon=true
org.gradle.jvmargs=-Xmx4g -XX:+UseParallelGC

kotlin.incremental=true
kotlin.caching.enabled=true

# Kotlin compilation avoidance
kotlin.compiler.execution.strategy=in-process

# Android build optimization
android.enableBuildCache=true
android.enableR8.fullMode=true
android.nonTransitiveRClass=true
```

### Build Cache Performance

```
┌──────────────────────────────────────────────────────────┐
│  Build Type         │ Clean   │ Cached   │ Up-to-Date   │
├──────────────────────────────────────────────────────────┤
│  Full build         │ 45s     │ 25s      │ 2s           │
│  Single file change │ 30s     │ 12s      │ 3s           │
│  Resource change    │ 25s     │ 8s       │ 2s           │
│  No change          │ 20s     │ 3s       │ 1s           │
└──────────────────────────────────────────────────────────┘
```

---

## 6. Memory Optimization

### 6.1 Leak Detection

```kotlin
// Application class
class NexusApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        if (BuildConfig.DEBUG) {
            StrictMode.setVmPolicy(
                StrictMode.VmPolicy.Builder()
                    .detectLeakedClosableObjects()
                    .detectLeakedSqlLiteObjects()
                    .detectActivityLeaks()
                    .penaltyLog()
                    .build()
            )
        }
    }
}

// LeakCanary (debug only)
// build.gradle.kts
debugImplementation("com.squareup.leakcanary:leakcanary-android:2.13")
```

### 6.2 Object Pooling

```kotlin
class MessagePool(private val maxSize: Int = 50) {
    private val pool = ArrayDeque<MessageUi>(maxSize)

    fun obtain(content: String, isUser: Boolean): MessageUi {
        return pool.removeFirstOrNull()?.copy(
            content = content, isUser = isUser
        ) ?: MessageUi(content = content, isUser = isUser)
    }

    fun recycle(message: MessageUi) {
        if (pool.size < maxSize) {
            pool.addLast(message)
        }
    }

    fun clear() = pool.clear()
}
```

### 6.3 Bitmap Management

```kotlin
class BitmapOptimizer {

    // Calculate inSampleSize for efficient decoding
    fun calculateInSampleSize(
        options: BitmapFactory.Options,
        reqWidth: Int,
        reqHeight: Int
    ): Int {
        val (height, width) = options.outHeight to options.outWidth
        var inSampleSize = 1

        if (height > reqHeight || width > reqWidth) {
            val halfHeight = height / 2
            val halfWidth = width / 2
            while (halfHeight / inSampleSize >= reqHeight &&
                   halfWidth / inSampleSize >= reqWidth) {
                inSampleSize *= 2
            }
        }
        return inSampleSize
    }

    // Decode with downsampling
    fun decodeSampledBitmap(
        res: Resources,
        resId: Int,
        reqWidth: Int,
        reqHeight: Int
    ): Bitmap {
        val options = BitmapFactory.Options().apply {
            inJustDecodeBounds = true
        }
        BitmapFactory.decodeResource(res, resId, options)

        options.inSampleSize = calculateInSampleSize(options, reqWidth, reqHeight)
        options.inJustDecodeBounds = false

        return BitmapFactory.decodeResource(res, resId, options)
    }
}
```

### 6.4 Memory Profiling

```
Memory Leak Detection Workflow:
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Android     │     │  LeakCanary  │     │  MAT /       │
│  Studio      │────►│  Auto-detect │────►│  Heap Dump   │
│  Profiler    │     │  in debug    │     │  Analysis    │
└──────────────┘     └──────────────┘     └──────────────┘
        │                    │                    │
        ▼                    ▼                    ▼
  ┌──────────┐        ┌──────────┐        ┌──────────┐
  │ Allocations│      │ Leak     │        │ Retained  │
  │ Graph      │      │ Traces   │        │ Size      │
  └──────────┘        └──────────┘        └──────────┘
```

| Memory Metric            | Healthy    | Warning    | Critical    |
|--------------------------|-----------|------------|-------------|
| Java Heap (P95)          | <80MB     | 80-128MB   | >128MB      |
| Native Heap              | <50MB     | 50-100MB   | >100MB      |
| Graphics                 | <30MB     | 30-60MB    | >60MB       |
| Allocation Rate          | <1MB/s    | 1-5MB/s    | >5MB/s      |
| GC Pauses                | <5ms      | 5-10ms     | >10ms       |
| GC Frequency             | <1/min    | 1-5/min    | >5/min      |
| Bitmap Memory            | <20MB     | 20-40MB    | >40MB       |

---

## 7. Image Loading Optimization

### 7.1 Coil Configuration

```kotlin
class CoilImageLoaderFactory(private val context: Context) {

    fun create(): ImageLoader {
        return ImageLoader.Builder(context)
            .memoryCache {
                MemoryCache.Builder(context)
                    .maxSizePercent(0.25) // 25% of available heap
                    .strongReferencesEnabled(true)
                    .build()
            }
            .diskCache {
                DiskCache.Builder()
                    .directory(context.cacheDir.resolve("image_cache"))
                    .maxSizePercent(0.02) // 2% of device storage
                    .build()
            }
            .components {
                add(SvgDecoder.Factory())
                add(GifDecoder.Factory())
            }
            .crossfade(true)
            .crossfade(300)
            .respectCacheHeaders(false) // Always use cache
            .build()
    }
}

// Usage in Compose
@Composable
fun UserAvatar(url: String, modifier: Modifier = Modifier) {
    AsyncImage(
        model = ImageRequest.Builder(LocalContext.current)
            .data(url)
            .crossfade(true)
            .size(Size.ORIGINAL)
            .diskCachePolicy(CachePolicy.ENABLED)
            .memoryCachePolicy(CachePolicy.ENABLED)
            .build(),
        contentDescription = "User avatar",
        modifier = modifier
            .size(48.dp)
            .clip(CircleShape),
        placeholder = painterResource(R.drawable.placeholder_avatar),
        error = painterResource(R.drawable.error_avatar)
    )
}
```

### 7.2 Image Format Comparison

| Format       | Size vs JPEG | Quality   | Animation | Transparency | Use Case          |
|-------------|-------------|-----------|-----------|-------------|-------------------|
| JPEG        | Baseline    | Lossy     | No        | No          | Photos            |
| PNG         | +25-50%     | Lossless  | No        | Yes         | Screenshots, UI   |
| WebP (lossy)| -25-35%     | Lossy     | No        | Yes         | Photos            |
| WebP (lossless)| +15-20% | Lossless  | No        | Yes         | Screenshots       |
| AVIF         | -50%        | Excellent | No        | Yes         | Next-gen photos   |
| SVG         | Small       | Vector    | Yes       | Yes         | Icons, logos      |
| Vector Drawable| Tiny      | Vector    | No        | Yes         | UI icons          |

---

## 8. Network Optimization

### 8.1 OkHttp Configuration

```kotlin
class NetworkModule {

    fun provideOkHttpClient(): OkHttpClient {
        return OkHttpClient.Builder()
            // Connection pool
            .connectionPool(
                ConnectionPool(
                    maxIdleConnections = 5,
                    keepAliveDuration = 5,
                    timeUnit = TimeUnit.MINUTES
                )
            )
            // Timeouts
            .connectTimeout(15, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(15, TimeUnit.SECONDS)
            .callTimeout(60, TimeUnit.SECONDS)
            // Caching
            .cache(Cache(
                directory = File(context.cacheDir, "http_cache"),
                maxSize = 50L * 1024 * 1024 // 50MB
            ))
            // Interceptors
            .addInterceptor(AuthInterceptor(tokenProvider))
            .addInterceptor(RetryInterceptor(maxRetries = 3))
            .addNetworkInterceptor(CacheControlInterceptor())
            // HTTP/2
            .protocols(listOf(Protocol.HTTP_2, Protocol.HTTP_1_1))
            .build()
    }
}
```

### 8.2 Cache Strategy

```
Cache Control Flow:
┌────────────┐     ┌────────────┐     ┌────────────┐
│  Request   │────►│  Check     │────►│  Cache     │
│            │     │  Cache     │     │  Hit?      │
└────────────┘     └────────────┘     └─────┬──────┘
                                     Yes │    │ No
                                       ▼      ▼
                              ┌──────────┐  ┌──────────┐
                              │ Return   │  │ Network  │
                              │ Cached   │  │ Request  │
                              └──────────┘  └─────┬────┘
                                            Yes │    │ No
                                              ▼      ▼
                                     ┌──────────┐ ┌────────┐
                                     │ Cache    │ │ Return │
                                     │ Response │ │ Error  │
                                     └──────────┘ └────────┘
```

| Endpoint Type    | Cache Strategy      | TTL      | stale-while-revalidate |
|-----------------|--------------------|----------|------------------------|
| User Profile     | Cache + Network    | 5 min    | 30 min                 |
| Wallet Balance   | Network Only       | 0        | 0                      |
| Chat History     | Cache + Network    | 1 min    | 10 min                 |
| Static Config    | Cache Only         | 24 hr    | 7 days                 |
| Transaction List | Cache + Network    | 2 min    | 15 min                 |

---

## 9. Database Optimization

### 9.1 Room Indexing

```kotlin
@Entity(
    tableName = "messages",
    indices = [
        Index(value = ["conversation_id"]),
        Index(value = ["timestamp"]),
        Index(value = ["conversation_id", "timestamp"])
    ]
)
data class MessageEntity(
    @PrimaryKey val id: String,
    @ColumnInfo(name = "conversation_id") val conversationId: String,
    val content: String,
    val role: String,
    val timestamp: Long
)
```

### 9.2 WAL Mode & Performance

```kotlin
class AppDatabase : RoomDatabase() {
    companion object {
        fun build(context: Context): AppDatabase {
            return Room.databaseBuilder(context, AppDatabase::class.java, "app.db")
                .setJournalMode(JournalMode.TRUNCATE) // Better for reads
                .set WAL (Write-Ahead Logging)        // Enable WAL
                .setQueryExecutor(
                    Executors.newFixedThreadPool(4)
                )
                .setTransactionExecutor(
                    Executors.newSingleThreadExecutor()
                )
                .fallbackToDestructiveMigration()
                .build()
        }
    }
}
```

### 9.3 Query Optimization

```kotlin
@Dao
interface MessageDao {
    // ❌ Bad: full table scan
    @Query("SELECT * FROM messages WHERE conversation_id = :convId")
    fun getMessagesSlow(convId: String): Flow<List<MessageEntity>>

    // ✅ Good: indexed query with projection
    @Query("""
        SELECT id, content, role, timestamp
        FROM messages
        WHERE conversation_id = :convId
        ORDER BY timestamp DESC
        LIMIT :limit
    """)
    fun getMessagesFast(convId: String, limit: Int = 50): Flow<List<MessageEntity>>

    // ✅ Good: paginated query
    @Query("""
        SELECT * FROM messages
        WHERE conversation_id = :convId
        AND timestamp < :before
        ORDER BY timestamp DESC
        LIMIT :limit
    """)
    suspend fun getMessagesBefore(
        convId: String, before: Long, limit: Int = 20
    ): List<MessageEntity>
}
```

---

## 10. LazyColumn Optimization

```kotlin
@Composable
fun ChatMessageList(messages: List<MessageUi>) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        // Key for stable IDs across recompositions
        key = { messages[it].id },
        // Content type for view holder reuse
        contentType = { messages[it].contentType }
    ) {
        items(
            items = messages,
            key = { it.id },
            contentType = { it.contentType }
        ) { message ->
            ChatBubble(message = message)
        }
    }
}

// Content type mapping for ViewType reuse
val MessageUi.contentType: String
    get() = when {
        isUser -> "user_bubble"
        isSystem -> "system_message"
        else -> "assistant_bubble"
    }
```

### Recomposition Optimization

```kotlin
// ❌ Bad: Unstable class causes recomposition
data class ChatMessage(
    val id: String,
    val content: String,
    val isUser: Boolean,
    val metadata: Map<String, Any> // Unstable!
)

// ✅ Good: Stable class avoids recomposition
@Immutable
data class ChatMessage(
    val id: String,
    val content: String,
    val isUser: Boolean,
    val metadata: Metadata // Stable
)

// Use derivedStateOf for computed values
@Composable
fun WalletSummary(transactions: List<Transaction>) {
    val totalIncome by remember {
        derivedStateOf {
            transactions.filter { it.amount > 0 }.sumOf { it.amount }
        }
    }
    val totalExpense by remember {
        derivedStateOf {
            transactions.filter { it.amount < 0 }.sumOf { it.amount }
        }
    }
}
```

---

## 11. State Optimization

```kotlin
// ❌ Bad: Whole state recomposes on any change
data class ChatUiState(
    val messages: List<MessageUi>,
    val isTyping: Boolean,
    val inputValue: String,
    val scrollPosition: Int
)

// ✅ Good: Split into granular states
@Stable
class ChatState {
    val messages = mutableStateListOf<MessageUi>()
    var isTyping by mutableStateOf(false)
    var inputValue by mutableStateOf("")
    var scrollPosition by mutableStateOf(0)
}

// Use select for flow collection
viewModel.uiState
    .map { it.messages }
    .distinctUntilChanged()
    .collect { messages ->
        // Only updates when messages change
    }

// Use collectAsStateWithLifecycle in Compose
@Composable
fun ChatScreen(viewModel: ChatViewModel) {
    val messages by viewModel.messages.collectAsStateWithLifecycle()
    val isTyping by viewModel.isTyping.collectAsStateWithLifecycle()
}
```

---

## 12. Performance Monitoring

### 12.1 Custom Metrics

```kotlin
class PerformanceMonitor(private val firebase: FirebasePerformance) {

    fun <T> trace(name: String, block: suspend () -> T): suspend () -> T = {
        val trace = firebase.newTrace(name)
        trace.start()
        try {
            val result = block()
            trace.putAttribute("status", "success")
            result
        } catch (e: Exception) {
            trace.putAttribute("status", "error")
            trace.putAttribute("error", e.message ?: "unknown")
            throw e
        } finally {
            trace.stop()
        }
    }

    fun reportStartupMetric(stage: String, durationMs: Long) {
        val metric = TraceMetric("startup_stage").apply {
            putAttribute("stage", stage)
            putAttribute("duration_ms", durationMs.toString())
        }
        firebase.recordTrace(metric)
    }

    fun reportNetworkMetric(
        endpoint: String,
        latencyMs: Long,
        statusCode: Int
    ) {
        val metric = TraceMetric("network_request").apply {
            putAttribute("endpoint", endpoint)
            putAttribute("latency_ms", latencyMs.toString())
            putAttribute("status_code", statusCode.toString())
        }
        firebase.recordTrace(metric)
    }
}
```

### 12.2 Startup Metrics Collection

```kotlin
class StartupTracer {
    private val milestones = mutableMapOf<String, Long>()
    private val startTime = SystemClock.elapsedRealtime()

    fun mark(stage: String) {
        milestones[stage] = SystemClock.elapsedRealtime() - startTime
    }

    fun report() {
        val trace = Performance.newTrace("app_startup")
        trace.start()
        milestones.forEach { (stage, duration) ->
            trace.putMetric(stage, duration)
        }
        trace.stop()
    }
}

// Usage
class NexusApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        val tracer = StartupTracer()

        tracer.mark("application_onCreate")

        // ... init code

        tracer.mark("application_initialized")
        tracer.report()
    }
}
```

---

## 13. App Size Optimization

| Technique                 | Savings   | Complexity | Impact     |
|--------------------------|-----------|-----------|------------|
| R8 / ProGuard            | 30-50%    | Low        | High       |
| WebP images              | 25-35%    | Low        | High       |
| Vector Drawables         | 60-80%    | Medium     | High       |
| Dynamic Feature Modules  | 20-40%    | High       | High       |
| Remove unused resources  | 5-15%     | Low        | Medium     |
| SVG icons                | 70-90%    | Low        | Medium     |
| Composable vs XML        | 10-20%    | Medium     | Low        |
| String compression       | 10-20%    | Low        | Low        |

### APK Analyzer Report

```
APK Size Breakdown:
┌──────────────────────────────────────────────────────┐
│  Total: 18.2 MB                                      │
│                                                      │
│  classes.dex     ████████████░░░░░░  6.2 MB (34%)   │
│  res/            ████████░░░░░░░░░░  4.8 MB (26%)   │
│  lib/            ██████░░░░░░░░░░░░  3.2 MB (18%)   │
│  assets/         ████░░░░░░░░░░░░░░  2.1 MB (12%)   │
│  META-INF/       ██░░░░░░░░░░░░░░░░  1.0 MB (5%)    │
│  other           ██░░░░░░░░░░░░░░░░  0.9 MB (5%)    │
└──────────────────────────────────────────────────────┘
```

---

## 14. Animation Optimization

```kotlin
// ❌ Bad: Expensive recomposition animation
@Composable
fun ChatBubbleOptimized(message: MessageUi) {
    var visible by remember { mutableStateOf(false) }

    LaunchedEffect(message.id) {
        visible = true
    }

    AnimatedVisibility(
        visible = visible,
        enter = fadeIn() + slideInVertically()
    ) {
        MessageContent(message)
    }
}

// ✅ Good: Use graphicsLayer for GPU-accelerated animation
@Composable
fun ScrollOptimizer(listState: LazyListState) {
    val firstVisibleIndex by remember {
        derivedStateOf { listState.firstVisibleItemIndex }
    }

    // Use graphicsLayer for scroll-based animations
    // (avoids recomposition)
    Box(
        modifier = Modifier
            .graphicsLayer {
                // This runs on the render thread, not composition
                alpha = if (firstVisibleIndex > 0) 0.8f else 1f
            }
    ) {
        // Content
    }
}
```

---

## 15. Battery Optimization

```
Battery Impact by Feature:
┌────────────────────────────────────────────────────────────┐
│  Feature                │ Impact  │ Mitigation            │
├────────────────────────────────────────────────────────────┤
│  WebSocket (persistent) │ HIGH    │ Use FCM fallback      │
│  Location tracking      │ HIGH    │ Batch, reduce freq    │
│  Image loading          │ MEDIUM  │ Cache, downsample     │
│  Background sync        │ MEDIUM  │ WorkManager           │
│  Animations             │ LOW     │ Hardware accel        │
│  Push notifications     │ LOW     │ FCM (Google)          │
└────────────────────────────────────────────────────────────┘
```

```kotlin
// WorkManager for background tasks
class SyncWorker(
    context: Context,
    params: WorkerParameters
) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result {
        return try {
            repository.syncPendingTransactions()
            Result.success()
        } catch (e: Exception) {
            Result.retry()
        }
    }

    companion object {
        fun enqueue(context: Context) {
            val constraints = Constraints.Builder()
                .setRequiredNetworkType(NetworkType.CONNECTED)
                .setRequiresBatteryNotLow(true)
                .build()

            val request = PeriodicWorkRequestBuilder<SyncWorker>(
                15, TimeUnit.MINUTES
            )
                .setConstraints(constraints)
                .setBackoffCriteria(
                    BackoffPolicy.EXPONENTIAL,
                    1, TimeUnit.MINUTES
                )
                .build()

            WorkManager.getInstance(context)
                .enqueueUniquePeriodicWork(
                    "sync",
                    ExistingPeriodicWorkPolicy.KEEP,
                    request
                )
        }
    }
}
```

---

## 16. Performance Testing

### Macrobenchmark

```kotlin
@RunWith(AndroidJUnit4::class)
class ScrollBenchmark {

    @get:Rule
    val benchmarkRule = MacrobenchmarkRule()

    @Test
    fun scrollChatHistory() = benchmarkRule.measureRepeated(
        packageName = "com.nexusai.app",
        metrics = listOf(
            FrameTimingMetric("P50"),
            FrameTimingMetric("P90"),
            FrameTimingMetric("P99")
        ),
        iterations = 10,
        compilationMode = CompilationMode.Partial(
            baselineProfileMode = BaselineProfileMode.Require
        )
    ) {
        startActivityAndWait()
        val list = device.findObject(By.res("chat_list"))
        list.fling(Direction.DOWN)
        device.waitForIdle()
    }
}

@RunWith(AndroidJUnit4::class)
class InputBenchmark {

    @get:Rule
    val benchmarkRule = MacrobenchmarkRule()

    @Test
    fun typeInChatInput() = benchmarkRule.measureRepeated(
        packageName = "com.nexusai.app",
        metrics = listOf(AutoscrollingInputMetric()),
        iterations = 5
    ) {
        startActivityAndWait()
        val input = device.findObject(By.res("chat_input"))
        input.text = "Hello, this is a test message for benchmark"
        device.waitForIdle()
    }
}
```

### Microbenchmark

```kotlin
@RunWith(AndroidJUnit4::class)
class SerializationBenchmark {

    @get:Rule
    val benchmarkRule = MicrobenchmarkRule()

    @Test
    fun serializeChatMessage() = benchmarkRule.measureRepeated {
        repeat(1000) {
            Json.encodeToString(testMessage)
        }
    }

    @Test
    fun parseJsonResponse() = benchmarkRule.measureRepeated {
        repeat(1000) {
            Json.decodeFromString<ChatResponse>(jsonString)
        }
    }
}
```

---

## 17. Performance Debugging

### 17.1 Systrace / Perfetto

```
Trace Markers for Custom Sections:
┌──────────────────────────────────────────────────────────┐
│  Section                  │ Method                       │
├──────────────────────────────────────────────────────────┤
│  App startup              │ Trace.beginSection()         │
│  Database query           │ Trace.beginSection()         │
│  Network request          │ Trace.beginSection()         │
│  WebSocket message        │ Trace.beginSection()         │
│  UI recomposition         │ traceComposition()           │
│  Image decode             │ Trace.beginSection()         │
└──────────────────────────────────────────────────────────┘
```

```kotlin
// Custom trace markers
suspend fun fetchWallet(): Wallet {
    return withContext(Dispatchers.IO) {
        Trace.beginSection("fetchWallet")
        try {
            apiClient.getWallet()
        } finally {
            Trace.endSection()
        }
    }
}
```

### 17.2 CPU Profiler

| Profile Type   | Use Case                     | Tool              |
|---------------|------------------------------|-------------------|
| CPU Sampling  | Hot method identification    | Android Studio    |
| CPU Tracing   | Method call analysis         | Perfetto          |
| Method Trace  | Detailed call graphs         | Android Studio    |
| Wall Clock    | Real-time latency            | Custom markers    |

---

## 18. Performance CI/CD

```yaml
# .github/workflows/performance.yml
name: Performance Benchmarks
on:
  pull_request:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          java-version: '17'

      - name: Build benchmark APK
        run: ./gradlew :benchmark:assembleBenchmark

      - name: Run startup benchmark
        uses: reactivecircus/android-emulator-runner@v2
        with:
          api-level: 34
          script: |
            ./gradlew :benchmark:pixel6Api33StartupBenchmark

      - name: Check for regressions
        run: |
          python scripts/check_benchmark_regression.py \
            --baseline benchmark-results/main.json \
            --current benchmark-results/current.json \
            --threshold 10

      - name: Upload benchmark results
        uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: benchmark-results/
```

### Regression Detection

```python
# scripts/check_benchmark_regression.py
import json
import sys

def check_regression(baseline: dict, current: dict, threshold: float = 0.10):
    regressions = []
    for metric, values in current.items():
        base_value = baseline.get(metric, values)
        if values > base_value * (1 + threshold):
            regressions.append({
                "metric": metric,
                "baseline": base_value,
                "current": values,
                "regression_pct": ((values - base_value) / base_value) * 100
            })
    return regressions
```

### Performance Budget Enforcement

```
Performance Gate:
┌────────────────────────────────────────────────────────────┐
│  Step 1: Build APK/AAB                                    │
│  Step 2: Run Macrobenchmark (5 iterations)                │
│  Step 3: Run APK Analyzer (size check)                    │
│  Step 4: Compare against budget                           │
│  Step 5: PASS if all metrics within budget                │
│           FAIL if any metric exceeds by >10%               │
│  Step 6: Post results to PR comment                      │
└────────────────────────────────────────────────────────────┘
```
