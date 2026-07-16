# 17. Performance

## Table of Contents

- [Performance Targets](#performance-targets)
- [Startup Optimization](#startup-optimization)
- [App Launch Time](#app-launch-time)
- [Time Profiler](#time-profiler)
- [Allocations Instrument](#allocations-instrument)
- [Leaks Instrument](#leaks-instrument)
- [Core Animation Instrument](#core-animation-instrument)
- [Network Instrument](#network-instrument)
- [Swift Optimization](#swift-optimization)
- [Build Optimization](#build-optimization)
- [Memory Optimization](#memory-optimization)
- [Image Caching](#image-caching)
- [Image Formats](#image-formats)
- [Network Optimization](#network-optimization)
- [CoreData Optimization](#coredata-optimization)
- [List Performance](#list-performance)
- [View Composition](#view-composition)
- [Animation Optimization](#animation-optimization)
- [Startup Metrics](#startup-metrics)
- [Rendering Metrics](#rendering-metrics)
- [Network Metrics](#network-metrics)
- [Memory Metrics](#memory-metrics)
- [Battery Optimization](#battery-optimization)
- [Network Sensitivity](#network-sensitivity)
- [App Size Optimization](#app-size-optimization)
- [Performance Monitoring](#performance-monitoring)
- [Performance Testing](#performance-testing)
- [Performance Budget](#performance-budget)
- [Performance Debugging](#performance-debugging)
- [Performance CI/CD](#performance-cicd)

---

## Performance Targets

| Metric                    | Target      | Critical     |
|---------------------------|-------------|--------------|
| Cold Start (pre-main)     | < 200ms     | 400ms        |
| Cold Start (post-main)    | < 1s        | 2s           |
| Warm Start                | < 300ms     | 500ms        |
| Hot Start                 | < 100ms     | 200ms        |
| Frame Rate                | 60 fps      | 30 fps       |
| Frame Time                | < 16.67ms   | 33.33ms      |
| Main Thread Usage         | < 25%       | 40%          |
| Peak Memory               | < 150MB     | 250MB        |
| Bundle Size               | < 50MB      | 100MB        |
| Network Request (p50)     | < 200ms     | 1s           |
| Network Request (p95)     | < 500ms     | 2s           |
| Scroll Jank Rate          | < 1%        | 5%           |
| Crash Rate                | < 0.1%      | 1%           |

---

## Startup Optimization

### App Startup Flow

```
┌─────────────────────────────────────────────────────────────┐
│                      PRE-MAIN                                 │
│  ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │  dylib   │→ │  Rebase  │→ │   Bind   │→ │   ObjC Setup │  │
│  │  Load    │  │          │  │          │  │              │  │
│  └─────────┘  └──────────┘  └──────────┘  └──────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                      POST-MAIN                                │
│  ┌──────────┐  ┌────────────┐  ┌──────────┐  ┌──────────┐   │
│  │ @main /  │→ │ App Init   │→ │  Window  │→ │  First   │   │
│  │ AppDelegate│ │ (services) │  │  Setup   │  │  Render  │   │
│  └──────────┘  └────────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Lazy Initialization

```swift
@main
struct NexusApp: App {
    private lazy var database: DatabaseManager = { DatabaseManager() }()
    private lazy var networkClient: NetworkClient = {
        NetworkClient(baseURL: Configuration.apiBaseURL)
    }()

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(\.database, database)
                .environment(\.networkClient, networkClient)
        }
    }
}
```

### Prewarming

```swift
actor PrewarmManager {
    static let shared = PrewarmManager()

    func prewarm() async {
        await withTaskGroup(of: Void.self) { group in
            group.addTask { _ = try? await DatabaseManager.shared.warmUp() }
            group.addTask { ImageCache.shared.calculateMemoryCache() }
            group.addTask { _ = await URLSession.shared.data(from: URL(string: "about:blank")!) }
        }
    }
}

// In AppDelegate
func application(_ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
    Task.detached(priority: .userInitiated) { await PrewarmManager.shared.prewarm() }
    return true
}
```

---

## App Launch Time

### Cold vs Warm vs Hot Start

| Type        | Trigger                        | Target    | Bottleneck              |
|-------------|--------------------------------|-----------|-------------------------|
| Cold Start  | App killed, user taps icon     | < 1.5s    | dylib loading           |
| Warm Start  | App in background, terminated  | < 0.5s    | State restoration       |
| Hot Start   | App in background, suspended   | < 0.1s    | Scene activation        |

### Measuring

```swift
let launchLog = OSLog(subsystem: "com.nexus.app", category: "Launch")
func application(_ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
    let id = OSSignpostID(log: launchLog)
    os_signpost(.begin, log: launchLog, name: "AppLaunch", signpostID: id)
    // Setup...
    os_signpost(.end, log: launchLog, name: "AppLaunch", signpostID: id)
    return true
}
```

### Pre-Main Optimization

```bash
# Check dylib load times
DYLD_PRINT_STATISTICS=1 ./NexusApp.app/NexusApp
```

Reduce dylib count by:
- Linking only needed frameworks
- Using static libraries where possible
- Avoiding massive umbrella frameworks

---

## Time Profiler

### Reading the Call Tree

```
┌──────────────────────────────────────────────────────────────────┐
│ Weight    Self     Symbol                                         │
│ ──────── ──────── ──────────────────────────────────────────      │
│ 100.0%   0.0%     main                                           │
│  ├─ 45.2% 0.1%    RunLoop                                        │
│  │  ├─ 22.1% 15.2% CellContentRenderer.draw()                    │
│  │  │  └─ 16.5% 14.8% ImageDecoder.decode()       ← HOT         │
│  │  └─ 8.3%  0.5%  TableView._updateVisibleCells()               │
│  └─ 53.3% 0.2%    NetworkActivity.processResponse()              │
│     └─ 42.1% 38.5% JSONDecoder.decode()              ← HOT       │
└──────────────────────────────────────────────────────────────────┘
```

| Column   | Meaning                                                |
|----------|--------------------------------------------------------|
| Weight   | Total time in symbol including callees                 |
| Self     | Time spent directly in this symbol (excluding callees) |

High **Self** = the function itself is slow. High **Weight** = it calls slow functions.

---

## Allocations Instrument

### Retain Cycle Detection

```
Persistent  #   Bytes   Category    Context
YES         1   2,048   ViewModel   FeedViewModel
  → Cycle: FeedViewModel → APIClient → FeedViewModel
  → Fix: Use [weak self] in closure
```

### Fixing Retain Cycles

```swift
// BAD
func fetchData() {
    apiClient.fetch { data in
        self.processData(data)  // LEAK
    }
}

// GOOD
func fetchData() {
    apiClient.fetch { [weak self] data in
        self?.processData(data)
    }
}

// GOOD - async/await
func fetchData() async {
    let data = await apiClient.fetch()
    await MainActor.run { [weak self] in
        self?.processData(data)
    }
}
```

---

## Leaks Instrument

Enable **Zombie Objects** for debugging deallocated access:
```
Product → Scheme → Run → Diagnostics → ☑ Enable Zombie Objects
```

---

## Core Animation Instrument

### Frame Budget

```
┌──────────────────────────────────────────────────────────┐
│               16.67ms Budget (60fps)                      │
│  ┌─────────┐  ┌──────────┐  ┌─────────┐  ┌───────────┐  │
│  │ Commit  │→ │ Render   │→ │ GPU     │→ │ Composite │  │
│  │ (CPU)   │  │ (CPU)    │  │ Draw    │  │ Display   │  │
│  │ 2ms     │  │ 5ms      │  │ 4ms     │  │ 2ms       │  │
│  └─────────┘  └──────────┘  └─────────┘  └───────────┘  │
│  If total > 16.67ms → Dropped Frame (jank)               │
└──────────────────────────────────────────────────────────┘
```

### Frame Rate Monitor

```swift
class FrameRateMonitor {
    private var displayLink: CADisplayLink?
    private var lastTimestamp: CFTimeInterval = 0
    private var frameCount = 0
    private var jankCount = 0

    var onJankDetected: ((Double) -> Void)?

    func start() {
        displayLink = CADisplayLink(target: self, selector: #selector(tick))
        displayLink?.add(to: .main, forMode: .common)
    }

    func stop() { displayLink?.invalidate(); displayLink = nil }

    @objc private func tick(_ link: CADisplayLink) {
        frameCount += 1
        let elapsed = link.timestamp - lastTimestamp
        if elapsed > 0.01867 {
            jankCount += 1
            onJankDetected?((elapsed - 0.01667) * 1000)
        }
        lastTimestamp = link.timestamp
    }

    var jankRate: Double {
        frameCount > 0 ? Double(jankCount) / Double(frameCount) : 0
    }
}
```

### Offscreen Rendering Triggers

```
Core Animation flags:
- Color Composites (blue) → Offscreen rendering
- Color Blended (red)     → Alpha blending

Common triggers:
- cornerRadius + masksToBounds
- Shadow without shadowPath
- masks with shouldRasterize
```

---

## Network Instrument

```
Request Timeline:
┌──────────────────────────────────────────────────────────────┐
│ Request     DNS     Connect  TLS     TTFB    Download  Done │
│ GET /users  12ms    45ms     68ms     120ms   35ms      155ms│
│ GET /txns   8ms     30ms     52ms     95ms    180ms     275ms│
└──────────────────────────────────────────────────────────────┘
```

---

## Swift Optimization

| Flag                  | Effect                                    |
|-----------------------|-------------------------------------------|
| `-O`                  | Basic optimizations (Release default)     |
| `-Osize`              | Optimize for size over speed              |
| `-whole-module-optimization` | Optimize across all module files    |

### Release Build Settings

```
SWIFT_OPTIMIZATION_LEVEL = -O
SWIFT_COMPILATION_MODE = wholemodule
GCC_OPTIMIZATION_LEVEL = s
DEAD_CODE_STRIPPING = YES
STRIP_INSTALLED_PRODUCT = YES
COPY_PHASE_STRIP = YES
```

---

## Build Optimization

```bash
# Check slow-to-compile functions
swiftc -debug-time-function-bodies YourFile.swift
swiftc -debug-time-expression-type-checking YourFile.swift
```

- Avoid changing files that import everything
- Use `@testable import` only in test targets
- Keep public API surface minimal

---

## Memory Optimization

### Image Downsampling

```swift
extension UIImage {
    static func downsample(imageAt url: URL, to pointSize: CGSize,
                           scale: CGFloat = UIScreen.main.scale) -> UIImage? {
        let maxDimension = max(pointSize.width, pointSize.height) * scale
        let opts = [kCGImageSourceShouldCache: false] as CFDictionary
        guard let src = CGImageSourceCreateWithURL(url as CFURL, opts) else { return nil }
        let downsampleOpts = [
            kCGImageSourceCreateThumbnailFromImageAlways: true,
            kCGImageSourceShouldCacheImmediately: true,
            kCGImageSourceCreateThumbnailWithTransform: true,
            kCGImageSourceThumbnailMaxPixelSize: maxDimension
        ] as CFDictionary
        guard let cg = CGImageSourceCreateThumbnailAtIndex(src, 0, downsampleOpts) else { return nil }
        return UIImage(cgImage: cg)
    }
}
```

### Autoreleasepool

```swift
func processItems(_ items: [DataItem]) {
    for batch in items.chunked(into: 100) {
        autoreleasepool {
            let processed = batch.map { process($0) }
            saveToDatabase(processed)
        }
    }
}
```

---

## Image Caching

```swift
actor ImageCache {
    static let shared = ImageCache()
    private let memoryCache = NSCache<NSString, UIImage>()
    private let diskCacheURL: URL

    init() {
        let dir = FileManager.default.urls(for: .cachesDirectory, in: .userDomainMask)[0]
        diskCacheURL = dir.appendingPathComponent("ImageCache")
        try? FileManager.default.createDirectory(at: diskCacheURL, withIntermediateDirectories: true)
        memoryCache.totalCostLimit = 50 * 1024 * 1024
        memoryCache.countLimit = 200
    }

    func image(forKey key: String) -> UIImage? {
        if let cached = memoryCache.object(forKey: key as NSString) { return cached }
        let fileURL = diskCacheURL.appendingPathComponent(key.md5Hash)
        guard let data = try? Data(contentsOf: fileURL), let img = UIImage(data: data) else { return nil }
        memoryCache.setObject(img, forKey: key as NSString, cost: data.count)
        return img
    }

    func store(_ image: UIImage, forKey key: String) {
        let data = image.pngData()
        memoryCache.setObject(image, forKey: key as NSString, cost: data?.count ?? 0)
        try? data?.write(to: diskCacheURL.appendingPathComponent(key.md5Hash))
    }
}
```

---

## Image Formats

| Format  | Use Case                  | Size vs JPEG | Alpha | Animation |
|---------|---------------------------|-------------|-------|-----------|
| HEIC    | Photos, user content      | -50%        | Yes   | No        |
| WebP    | Web assets, thumbnails    | -30%        | Yes   | Yes       |
| Vector  | Icons, logos              | -90%        | Yes   | No        |
| PNG     | UI elements               | +50%        | Yes   | No        |

---

## Network Optimization

```swift
extension URLSession {
    static let cachedSession: URLSession = {
        let config = URLSessionConfiguration.default
        config.urlCache = URLCache(memoryCapacity: 10_000_000, diskCapacity: 100_000_000, diskPath: "net_cache")
        config.requestCachePolicy = .returnCacheDataElseLoad
        return URLSession(configuration: config)
    }()
}

// HTTP/2 reuse
class NetworkManager {
    static let shared = NetworkManager()
    private let session: URLSession
    init() {
        let config = URLSessionConfiguration.default
        config.httpMaximumConnectionsPerHost = 6
        config.httpShouldUsePipelining = true
        self.session = URLSession(configuration: config)
    }
}
```

---

## CoreData Optimization

```swift
extension NSManagedObjectContext {
    func batchFetch<T: NSManagedObject>(_ type: T.Type, predicate: NSPredicate? = nil,
                                        batchSize: Int = 20) throws -> [T] {
        let request = NSFetchRequest<T>(entityName: String(describing: type))
        request.predicate = predicate
        request.fetchBatchSize = batchSize
        request.returnsObjectsAsFaults = false
        return try fetch(request)
    }

    func efficientCount<T: NSManagedObject>(for type: T.Type) throws -> Int {
        let request = NSFetchRequest<T>(entityName: String(describing: type))
        request.resultType = .countResultType
        return (try fetch(request).first as? Int) ?? 0
    }
}
```

### Indexing (xcdatamodeld)

| Entity      | Indexed Attributes                    |
|-------------|---------------------------------------|
| Transaction | date, amount, category, status       |
| User        | email, id                            |

---

## List Performance

```swift
// BAD: All items created at once
ScrollView { VStack { ForEach(items) { ItemRow(item: $0) } } }

// GOOD: LazyVStack creates on demand
ScrollView { LazyVStack { ForEach(items) { ItemRow(item: $0) } } }

// BEST: With stable IDs and content shape
ScrollView {
    LazyVStack(spacing: 0) {
        ForEach(items) { item in
            ItemRow(item: item).id(item.id)
        }
    }
    .contentShape(Rectangle())
}
```

### Diffable Data Source (UIKit)

```swift
typealias Section = TransactionSection

enum TransactionSection: Hashable { case today, thisWeek, older }

func updateSnapshot(with transactions: [Item]) {
    var snapshot = NSDiffableDataSourceSnapshot<TransactionSection, Item>()
    snapshot.appendSections([.today, .thisWeek, .older])
    snapshot.appendItems(transactions.filter { Calendar.current.isDateInToday($0.date) }, toSection: .today)
    snapshot.appendItems(transactions.filter { !Calendar.current.isDateInToday($0.date) }, toSection: .thisWeek)
    dataSource.apply(snapshot, animatingDifferences: true)
}
```

---

## View Composition

```swift
// Rasterize complex views into single GPU texture
ComplexChartView(data: chartData)
    .drawingGroup()

// Composite opacity layers together
ZStack {
    BackgroundView()
    OverlayView().opacity(0.8)
}
.compositingGroup()
```

---

## Animation Optimization

```swift
// BAD: Implicit animation on frequent values
Circle().animation(.easeInOut, value: frequentlyChangingValue)

// GOOD: Wrap state change in withAnimation
withAnimation(.spring()) { items = newItems }

// GOOD: Transition for enter/exit
if showBanner {
    BannerView()
        .transition(.move(edge: .top).combined(with: .opacity))
}
```

---

## Startup Metrics

| Metric                  | Tool                          | Target |
|-------------------------|-------------------------------|--------|
| Pre-main time           | DYLD_PRINT_STATISTICS=1       | < 200ms|
| Post-main to first frame| Instruments App Launch         | < 1s   |
| Time to interactive     | Custom os_signpost             | < 1.5s |

---

## Rendering Metrics

| Metric              | Good       | Warning    | Bad        |
|---------------------|------------|------------|------------|
| Frame Rate          | 60 fps     | 45-59 fps  | < 45 fps   |
| Frame Time          | < 16.67ms  | 16-25ms    | > 25ms     |
| Jank Rate           | < 1%       | 1-5%       | > 5%       |
| Offscreen Renders   | 0          | 1-2        | > 3        |

---

## Network & Memory Metrics

| Metric              | Target   | Alert    | Critical |
|---------------------|----------|----------|----------|
| Network p50         | < 200ms  | 500ms    | 1s       |
| Network p95         | < 500ms  | 1s       | 2s       |
| Error Rate          | < 1%     | 3%       | 5%       |
| Peak Memory         | < 100MB  | 150MB    | 250MB    |
| Average Memory      | < 60MB   | 80MB     | 120MB    |
| Leaked Objects      | 0        | 1-5      | > 5      |

---

## Battery & Network Sensitivity

```swift
class NetworkMonitor {
    static let shared = NetworkMonitor()
    private let monitor = NWPathMonitor()
    private let queue = DispatchQueue(label: "NetworkMonitor")
    private(set) var isConnected = true
    private(set) var isExpensive = false
    private(set) var isConstrained = false

    var currentQuality: NetworkQuality {
        if !isConnected { return .offline }
        if isConstrained { return .constrained }
        if isExpensive { return .expensive }
        return .good
    }

    func startMonitoring() {
        monitor.pathUpdateHandler = { [weak self] path in
            self?.isConnected = path.status == .satisfied
            self?.isExpensive = path.isExpensive
            self?.isConstrained = path.isConstrained
        }
        monitor.start(queue: queue)
    }
}

enum NetworkQuality {
    case offline, constrained, expensive, good
    var imageQuality: CGFloat {
        switch self { case .offline: 0; case .constrained: 0.4; case .expensive: 0.6; case .good: 0.8 }
    }
}
```

---

## App Size Optimization

```
App Thinning Report:
┌──────────────────┬─────────┬────────────────┬────────────┐
│ Component        │ Size    │ After Thinning │ Savings    │
├──────────────────┼─────────┼────────────────┼────────────┤
│ Executable       │ 18MB    │ 12MB           │ 33%        │
│ Assets (2x)      │ 25MB    │ 12MB           │ 52%        │
│ Assets (3x)      │ 25MB    │ 8MB            │ 68%        │
│ Resources        │ 5MB     │ 3MB            │ 40%        │
├──────────────────┼─────────┼────────────────┼────────────┤
│ TOTAL            │ 73MB    │ 35MB           │ 52%        │
└──────────────────┴─────────┴────────────────┴────────────┘
```

Techniques: App Thinning, On-Demand Resources, remove @1x assets, dead code stripping.

---

## Performance Monitoring

### MetricKit

```swift
import MetricKit

class PerformanceMonitor: NSObject, MXMetricManagerSubscriber {
    static let shared = PerformanceMonitor()

    func start() { MXMetricManager.shared.add(self) }

    func didReceive(_ payloads: [MXMetricPayload]) {
        for payload in payloads {
            if let launch = payload.applicationLaunchMetrics {
                os_log(.info, "TTFT: %f ms", launch.timeToFirstDraw.rawValue * 1000)
            }
            if let memory = payload.memoryMetrics {
                os_log(.info, "Peak: %f MB", memory.peakMemoryUsage.rawValue / 1_000_000)
            }
        }
    }

    func didReceive(_ payloads: [MXDiagnosticPayload]) {
        for payload in payloads {
            if let hangs = payload.hangDiagnostics {
                for hang in hangs { os_log(.error, "Hang: %f s", hang.hangDuration.rawValue) }
            }
        }
    }
}
```

### os_signpost

```swift
import os.signpost

let log = OSLog(subsystem: "com.nexus.app", category: "Perf")

func traceOperation<T>(_ name: String, work: () throws -> T) rethrows -> T {
    let id = OSSignpostID(log: log)
    os_signpost(.begin, log: log, name: name, signpostID: id)
    let result = try work()
    os_signpost(.end, log: log, name: name, signpostID: id)
    return result
}
```

---

## Performance Testing

```swift
final class PerformanceTests: XCTestCase {
    func testDecodePerformance() throws {
        let data = try JSONEncoder().encode(Transaction.mockArray(count: 1000))
        measure(metrics: [XCTCPUMetric()]) {
            _ = try? JSONDecoder().decode([Transaction].self, from: data)
        }
    }

    func testCoreDataFetchPerformance() throws {
        let container = NSPersistentContainer.inMemoryContainer()
        for i in 0..<10_000 {
            let e = TransactionEntity(context: container.viewContext)
            e.id = "\(i)"; e.amount = Double(i)
        }
        try container.viewContext.save()

        measure(metrics: [XCTMemoryMetric()]) {
            let exp = XCTestExpectation(description: "Fetch")
            Task { _ = try? await LocalTransactionRepository(context: container.viewContext).fetchAll(); exp.fulfill() }
            wait(for: [exp], timeout: 5)
        }
    }

    func testImageProcessingPerformance() throws {
        let image = UIImage(named: "large_image")!
        measure(metrics: [XCTCPUMetric()]) {
            _ = image.downsample(to: CGSize(width: 200, height: 200), scale: UIScreen.main.scale)
        }
    }
}
```

---

## Performance Budget

| Category        | Budget    | Enforcement            |
|-----------------|-----------|------------------------|
| Binary Size     | < 30MB    | CI check, fail build   |
| Total IPA       | < 50MB    | CI check, fail build   |
| Cold Start      | < 1.5s    | XCTest performance     |
| Memory Peak     | < 150MB   | XCTest performance     |
| Frame Drops     | < 1%      | XCTest UI performance  |

```yaml
# CI Binary Size Check
- name: Check Binary Size
  run: |
    SIZE=$(stat -f%z NexusApp.app/NexusApp)
    [ "$SIZE" -lt 31457280 ] || exit 1  # 30MB
```

---

## Performance Debugging

```swift
enum Log {
    static let network = Logger(subsystem: "com.nexus.app", category: "Network")
    static let ui = Logger(subsystem: "com.nexus.app", category: "UI")
    static let perf = Logger(subsystem: "com.nexus.app", category: "Perf")
}
Log.perf.info("Feed loaded in \(elapsed, format: .fixed(precision: 3))s")
```

---

## Performance CI/CD

```yaml
name: Performance Check
on:
  pull_request:
    branches: [main]
jobs:
  performance:
    runs-on: macos-14
    steps:
      - uses: actions/checkout@v4
      - name: Run Perf Tests
        run: |
          xcodebuild test -scheme NexusApp \
            -destination 'platform=iOS Simulator,name=iPhone 15' \
            -only-testing:NexusAppTests/PerformanceTests \
            -resultBundlePath ./perf-results.xcresult
      - name: Check Binary Size
        run: |
          xcodebuild archive -scheme NexusApp -archivePath ./build/NexusApp.xcarchive
          SIZE=$(stat -f%z ./build/NexusApp.xcarchive/Products/Applications/NexusApp.app/NexusApp)
          [ "$SIZE" -lt 31457280 ] || exit 1
```
