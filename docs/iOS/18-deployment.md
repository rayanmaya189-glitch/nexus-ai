# 18. Deployment

## Table of Contents

- [App Store Preparation](#app-store-preparation)
- [App Store Listing](#app-store-listing)
- [App Store Screenshots](#app-store-screenshots)
- [App Store Graphics](#app-store-graphics)
- [App Store Categories](#app-store-categories)
- [App Store Age Rating](#app-store-age-rating)
- [App Store Privacy](#app-store-privacy)
- [Privacy Nutrition Labels](#privacy-nutrition-labels)
- [App Store Targeting](#app-store-targeting)
- [Provisioning Profiles](#provisioning-profiles)
- [Code Signing](#code-signing)
- [Build Configurations](#build-configurations)
- [Release Build](#release-build)
- [Version Management](#version-management)
- [Build Automation](#build-automation)
- [Xcode Cloud Workflow](#xcode-cloud-workflow)
- [Fastlane Integration](#fastlane-integration)
- [App Store Review Process](#app-store-review-process)
- [App Store Updates](#app-store-updates)
- [Crash Reporting](#crash-reporting)
- [Analytics](#analytics)
- [Remote Config](#remote-config)
- [A/B Testing](#ab-testing)
- [App Store Optimization (ASO)](#app-store-optimization-aso)
- [Post-Launch Monitoring](#post-launch-monitoring)
- [Hotfix Process](#hotfix-process)
- [App Backup](#app-backup)
- [Widget Development](#widget-development)
- [App Clips](#app-clips)
- [CarPlay](#carplay)
- [Apple Watch](#apple-watch)

---

## App Store Preparation

### Pre-Launch Checklist

```
┌────────────────────────────────────────────────────────────────┐
│                    PRE-LAUNCH CHECKLIST                         │
├────────────────────────────────────────────────────────────────┤
│  □ App name and subtitle finalized                             │
│  □ App description written (4000 char max)                     │
│  □ Keywords entered (100 char max, comma-separated)            │
│  □ Screenshots captured for all required sizes                 │
│  □ App icon (1024x1024) uploaded                               │
│  □ Privacy policy URL published                                │
│  □ Privacy nutrition labels filled                             │
│  □ Age rating questionnaire completed                          │
│  □ Categories selected (primary + secondary)                   │
│  □ Pricing set                                                 │
│  □ In-App Purchase products configured                         │
│  □ Build uploaded via Xcode / Transporter / fastlane           │
│  □ Demo account credentials provided for review                │
│  □ Tax and banking information submitted                       │
│  □ Agreements accepted                                         │
│  □ Support URL functional                                      │
└────────────────────────────────────────────────────────────────┘
```

---

## App Store Listing

### Metadata Fields

| Field               | Limit        | Required |
|---------------------|--------------|----------|
| App Name            | 30 chars     | Yes      |
| Subtitle            | 30 chars     | Yes      |
| Description         | 4000 chars   | Yes      |
| Keywords            | 100 chars    | Yes      |
| Promotional Text    | 170 chars    | No       |
| What's New          | 4000 chars   | Yes      |
| Support URL         | URL          | Yes      |
| Marketing URL       | URL          | No       |
| Privacy Policy URL  | URL          | Yes      |

### Best Practices

```markdown
## App Name
- Include primary keyword, keep under 25 chars
- Example: "Nexus - Smart Finance"

## Subtitle
- Complement app name with secondary keywords
- Example: "Track Spending & Investments"

## Keywords
- Comma-separated, no spaces after commas
- Don't repeat words from app name
- Use singular forms, include common misspellings
- Example: "finance,money,budget,tracker,spending,investment,savings,portfolio,banking,pay"

## Description
- First 3 lines critical (shown in search results)
- Use bullet points for features
- Include social proof / awards
```

---

## App Store Screenshots

### Required Sizes

| Device                     | Size (pixels)        | Required |
|----------------------------|----------------------|----------|
| iPhone 6.7" (15 Pro Max)   | 1290 x 2796         | Yes      |
| iPhone 6.5" (14 Plus)      | 1242 x 2688         | Yes      |
| iPhone 5.5" (8 Plus)       | 1242 x 2208         | Yes      |
| iPad Pro 12.9" (6th gen)   | 2048 x 2732         | Yes*     |

*Required if app supports iPad

### Screenshot Order

```
1. Hero screenshot - value proposition + key feature
2. Core feature demo - primary use case
3. Secondary feature - breadth of capabilities
4. Social proof - awards, ratings
5. Call to action
```

---

## App Store Graphics

```
App Store Icon: 1024 x 1024 px, PNG/JPEG (no alpha), no rounded corners, no text
Marketing: 1024 x 500 px feature graphic
```

---

## App Store Categories

| Primary         | Subcategories                        |
|-----------------|--------------------------------------|
| Finance         | Banking, Budgeting, Investing        |
| Productivity    | Utilities, Task Management           |

---

## App Store Age Rating

```
Age Rating Questionnaire:
┌──────────────────────────────────────────────┐
│ Category                │ Rating             │
├─────────────────────────┼────────────────────┤
│ Cartoon/Fantasy Violence│ 4+ (None)          │
│ Realistic Violence      │ None selected      │
│ Sexual Content/Nudity   │ None selected      │
│ Alcohol/Tobacco/Drugs   │ None selected      │
│ Gambling                │ None selected      │
│ Horror/Fear Themes      │ None selected      │
└─────────────────────────┴────────────────────┘
Result: 4+ (suitable for all ages)
```

---

## App Store Privacy

### Privacy Policy Must Include

```
□ What data you collect
□ How you use the data
□ How you share the data
□ How users can control their data
□ Data retention period
□ Contact information
□ Third-party SDK disclosures
□ Children's privacy (COPPA)
□ International data transfers
```

---

## Privacy Nutrition Labels

| Category           | Data Type        | Tracking | Linked to Identity |
|--------------------|------------------|----------|--------------------|
| Contact Info       | Email Address    | No       | Yes                |
| Financial Info     | Payment History  | No       | Yes                |
| Identifiers        | User ID          | No       | Yes                |
| Identifiers        | Device ID        | Yes      | Yes                |
| Usage Data         | Product Interaction | Yes    | Yes                |
| Diagnostics        | Crash Data       | No       | No                 |

---

## App Store Targeting

### Available Territories (175+)

```
Americas    │ Europe           │ Asia Pacific
────────────┼──────────────────┼──────────────────
USA         │ UK               │ Japan
Canada      │ Germany          │ South Korea
Mexico      │ France           │ Australia
Brazil      │ Spain            │ India
Argentina   │ Italy            │ Singapore
Colombia    │ Netherlands      │ Hong Kong
Chile       │ Switzerland      │ Taiwan
            │ Sweden / Norway  │ New Zealand
```

### Localization

| Language               | Locale   | Required |
|------------------------|----------|----------|
| English (US)           | en-US    | Primary  |
| Spanish (Mexico)       | es-MX    | Yes      |
| Portuguese (BR)        | pt-BR    | Yes      |
| French                 | fr       | Yes      |
| German                 | de       | Yes      |
| Japanese               | ja       | Yes      |
| Korean                 | ko       | Yes      |
| Chinese (Simplified)   | zh-Hans  | Yes      |

---

## Provisioning Profiles

| Type                    | Use Case                    | Expiry  |
|-------------------------|-----------------------------|---------|
| iOS App Development     | Development & testing       | 1 year  |
| Ad Hoc                  | Limited device testing      | 1 year  |
| App Store               | Production distribution     | 1 year  |

```
Automatic Signing (Recommended):
  Xcode → Signing & Capabilities
  ☑ Automatically manage signing
  Team: [Your Team]
  Xcode handles certificates, profiles, bundle ID registration.
```

---

## Code Signing

```xml
<!-- NexusApp.entitlements -->
<dict>
    <key>aps-environment</key><string>development</string>
    <key>com.apple.developer.app-groups</key>
    <array><string>group.com.nexus.app</string></array>
    <key>com.apple.developer.healthkit</key><true/>
    <key>keychain-access-groups</key>
    <array><string>$(AppIdentifierPrefix)com.nexus.app</string></array>
</dict>
```

---

## Build Configurations

| Setting                    | Debug        | Release        |
|---------------------------|--------------|----------------|
| SWIFT_OPTIMIZATION_LEVEL  | `-Onone`     | `-O`           |
| SWIFT_COMPILATION_MODE    | singlefile   | wholemodule     |
| DEBUG_INFORMATION_FORMAT  | dwarf        | dwarf-with-dsym |
| ONLY_ACTIVE_ARCH          | YES          | NO             |

### Staging Configuration

```
// Staging.xcconfig
#include "Debug.xcconfig"
API_BASE_URL = https://staging-api.nexusapp.com
BUNDLE_IDENTIFIER = com.nexus.app.staging
ENABLE_ANALYTICS = NO
LOG_LEVEL = verbose
```

---

## Release Build

```bash
# Archive
xcodebuild archive \
  -project NexusApp.xcodeproj -scheme NexusApp \
  -configuration Release \
  -archivePath ./build/NexusApp.xcarchive \
  -destination "generic/platform=iOS" \
  -allowProvisioningUpdates

# Export IPA
xcodebuild -exportArchive \
  -archivePath ./build/NexusApp.xcarchive \
  -exportOptionsPlist ExportOptions.plist \
  -exportPath ./build/ipa
```

```xml
<!-- ExportOptions.plist -->
<dict>
    <key>method</key><string>app-store</string>
    <key>teamID</key><string>XXXXXXXXXX</string>
    <key>uploadBitcode</key><false/>
    <key>uploadSymbols</key><true/>
</dict>
```

---

## Version Management

```
Format: MAJOR.MINOR.PATCH (+ BUILD)
Example: 2.3.1 (build 456)

Info.plist keys:
  CFBundleShortVersionString = 2.3.1
  CFBundleVersion = 456

MAJOR (2): Breaking changes
MINOR (3): New features
PATCH (1): Bug fixes
```

---

## Build Automation

```bash
#!/bin/bash
set -euo pipefail

# Test
xcodebuild test -project NexusApp.xcodeproj -scheme NexusApp \
  -destination "platform=iOS Simulator,name=iPhone 15" \
  -enableCodeCoverage YES \
  -resultBundlePath ./TestResults.xcresult

# Archive
xcodebuild archive -project NexusApp.xcodeproj -scheme NexusApp \
  -configuration Release \
  -archivePath ./build/NexusApp.xcarchive \
  -destination "generic/platform=iOS" \
  -allowProvisioningUpdates

# Export
xcodebuild -exportArchive \
  -archivePath ./build/NexusApp.xcarchive \
  -exportOptionsPlist ExportOptions.plist \
  -exportPath ./build/ipa
```

---

## Xcode Cloud Workflow

```bash
#!/bin/bash
# ci_scripts/ci_post_clone.sh
set -euo pipefail
brew install swiftlint
swift package resolve
```

---

## Fastlane Integration

```ruby
# fastlane/Fastfile
default_platform(:ios)

platform :ios do
  desc "Run tests"
  lane :test do
    scan(project: "NexusApp.xcodeproj", scheme: "NexusApp",
         devices: ["iPhone 15"], code_coverage: true)
  end

  desc "Build and upload to TestFlight"
  lane :beta do
    increment_build_number(xcodeproj: "NexusApp.xcodeproj")
    build_app(project: "NexusApp.xcodeproj", scheme: "NexusApp", configuration: "Release")
    upload_to_testflight(skip_waiting_for_build_processing: true)
  end

  desc "Release to App Store"
  lane :release do
    build_app(project: "NexusApp.xcodeproj", scheme: "NexusApp", configuration: "Release")
    upload_to_app_store(force: true)
  end

  desc "Increment version"
  lane :bump do |options|
    increment_version_number(bump_type: options[:type] || "patch")
    increment_build_number
    version = get_version_number
    git_commit(path: ["NexusApp/Info.plist"], message: "Bump to v#{version}")
    add_git_tag(tag: "v#{version}")
    push_to_git_remote
  end
end
```

```ruby
# fastlane/Appfile
app_identifier("com.nexus.app")
apple_id("developer@nexusapp.com")
team_id("XXXXXXXXXX")
itc_team_id("XXXXXXXXXX")
```

---

## App Store Review Process

### Common Rejection Reasons

| Reason                      | Prevention                          |
|-----------------------------|-------------------------------------|
| Crashes / Bugs              | Thorough testing, crash-free release |
| Incomplete Metadata         | Fill all required fields            |
| Broken Links                | Verify all URLs before submission   |
| Placeholder Content         | Remove all lorem ipsum / test data  |
| Privacy Policy Missing      | Add valid privacy policy URL        |
| Misleading Description      | Accurate app description            |
| Minimum Functionality       | App must have sufficient utility    |
| Third-party Payments        | Use Apple IAP for digital goods     |

### Review Notes Template

```
Demo Account:
  Email: reviewer@nexusapp.com
  Password: TestReview2024!

Notes:
  - App requires authentication to access features
  - Location permission needed for nearby features
  - Push notifications used for transaction alerts

  Login with provided credentials
  Tap "Explore" tab to see main features
```

---

## App Store Updates

### TestFlight Distribution

```
Upload Build → Processing (10-30 min) → TestFlight Available
                                            ├→ Internal Testing (up to 100)
                                            └→ External Testing (up to 10k)
                                                    └→ Submit for Review
```

### Phased Rollout

```
Day 1:  1% of users
Day 2:  2% of users
Day 3:  5% of users
Day 4: 10% of users
Day 5: 20% of users
Day 6: 50% of users
Day 7: 100% of users

Monitor: If crash_rate > 1% → Pause rollout
```

---

## Crash Reporting

```swift
import FirebaseCrashlytics

func application(_ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
    FirebaseApp.configure()

    if let userId = UserDefaults.standard.string(forKey: "userId") {
        Crashlytics.crashlytics().setUserID(userId)
    }
    Crashlytics.crashlytics().setCustomValue("premium", forKey: "subscription_type")

    do {
        try loadData()
    } catch {
        Crashlytics.crashlytics().record(error: error)
    }
    return true
}

// Log non-fatal errors
Crashlytics.crashlytics().record(error: NSError(domain: "com.nexus", code: -1))
Crashlytics.crashlytics().log("User tapped send button")
```

### Symbol Upload

```bash
gcloud firebase crashlytics:symbols:upload \
  --app=com.nexus.app --dsyms-path=./build/dSYMs
```

---

## Analytics

```swift
import FirebaseAnalytics

// Log events
Analytics.logEvent("transaction_created", parameters: [
    "amount": 100.0,
    "currency": "USD"
])

// User properties
Analytics.setUserID(userId)
Analytics.setUserProperty("premium", forName: "subscription_type")

// Screen tracking
Analytics.logEvent(AnalyticsEventScreenView, parameters: [
    AnalyticsParameterScreenName: "HomeScreen",
    AnalyticsParameterScreenClass: "HomeView"
])
```

---

## Remote Config

```swift
import FirebaseRemoteConfig

class FeatureFlags {
    static let shared = FeatureFlags()
    private let remoteConfig: RemoteConfig

    init() {
        remoteConfig = RemoteConfig.remoteConfig()
        let settings = RemoteConfigSettings()
        #if DEBUG
        settings.minimumFetchInterval = 0
        #else
        settings.minimumFetchInterval = 3600
        #endif
        remoteConfig.configSettings = settings
        remoteConfig.setDefaults([
            "show_new_feature": false as NSNumber,
            "maintenance_mode": false as NSNumber
        ])
    }

    func fetchAndActivate() async {
        try? await remoteConfig.fetchAndActivate()
    }

    var isMaintenanceMode: Bool {
        remoteConfig.configValue(forKey: "maintenance_mode").boolValue
    }
}
```

---

## A/B Testing

```swift
// Firebase A/B Testing via Remote Config
func checkExperiment() -> String {
    RemoteConfig.remoteConfig().configValue(forKey: "welcome_experiment").stringValue
}

enum WelcomeVariant: String {
    case `default`, personalized, minimal
}
```

---

## App Store Optimization (ASO)

```
Keyword Strategy:
  1. Research competitor keywords (AppTweak, Sensor Tower)
  2. Track keyword rankings weekly
  3. Update keywords with each release
  4. Localize keywords for each market

A/B Test These Elements:
  □ App icon variants
  □ Screenshot order
  □ Subtitle variations
  □ Description formatting
```

---

## Post-Launch Monitoring

| Metric              │ Target      │ Alert    │ Critical  │
|---------------------|-------------|----------|-----------│
│ Crash Rate          │ < 0.1%      │ 0.5%     │ 1.0%      │
│ ANR Rate            │ < 0.05%     │ 0.1%     │ 0.5%      │
│ Daily Active Users  │ Growing     │ Flat     │ Declining │
│ Session Length      │ > 3 min     │ 1-3 min  │ < 1 min   │
│ Retention (D1)      │ > 40%       │ 20-40%   │ < 20%     │
│ Retention (D7)      │ > 20%       │ 10-20%   │ < 10%     │
│ Rating Average      │ > 4.5       │ 3.5-4.5  │ < 3.5     │

---

## Hotfix Process

```
Critical Bug → Fix on hotfix/* → Fastlane beta → TestFlight
    → Expedited App Review (24-48h) → Phased Rollout
```

```ruby
lane :hotfix do |options|
  sh "git checkout -b hotfix/#{options[:bug_id]} main"
  increment_version_number(bump_type: "patch")
  increment_build_number
  build_app(scheme: "NexusApp")
  upload_to_testflight(changelog: "Hotfix: #{options[:bug_id]}")
end
```

---

## App Backup

```swift
// Exclude non-essential data from iCloud backup
func markAsExcludedFromBackup() {
    let dirs = FileManager.default.urls(for: .cachesDirectory, in: .userDomainMask)
    for dir in dirs {
        var values = URLResourceValues()
        values.isExcludedFromBackup = true
        try? dir.setResourceValues(values)
    }
}
```

### Data Protection

```swift
// CoreData with protection
let description = NSPersistentStoreDescription()
description.url = storeURL
description.setOption(
    FileProtectionType.completeUntilFirstUserAuthentication.rawValue as NSObject,
    forKey: NSPersistentStoreFileProtectionKey
)
```

---

## Widget Development

```swift
import WidgetKit

struct NexusWidget: Widget {
    let kind: String = "NexusWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: NexusProvider()) { entry in
            NexusWidgetView(entry: entry)
        }
        .configurationDisplayName("Nexus Balance")
        .description("View your current balance")
        .supportedFamilies([.systemSmall, .systemMedium, .systemLarge])
    }
}

struct NexusProvider: TimelineProvider {
    func placeholder(in context: Context) -> BalanceEntry {
        BalanceEntry(date: Date(), balance: "$0.00")
    }
    func getSnapshot(in context: Context, completion: @escaping (BalanceEntry) -> Void) {
        completion(placeholder(in: context))
    }
    func getTimeline(in context: Context, completion: @escaping (Timeline<BalanceEntry>) -> Void) {
        Task {
            let entry = await fetchBalanceEntry()
            let next = Calendar.current.date(byAdding: .minute, value: 15, to: Date())!
            completion(Timeline(entries: [entry], policy: .after(next)))
        }
    }
}

struct BalanceEntry: TimelineEntry {
    let date: Date
    let balance: String
}

struct NexusWidgetView: View {
    let entry: BalanceEntry
    var body: some View {
        VStack(alignment: .leading) {
            Text("Balance").font(.caption).foregroundColor(.secondary)
            Text(entry.balance).font(.title.bold())
        }.padding()
    }
}
```

---

## App Clips

```swift
// Configured in Xcode → Target → App Clip
// Invocations: QR Code, NFC, Safari Universal Link, Maps, Messages
// Header Image: 1200x600, Title: 50 chars max
```

---

## CarPlay

```swift
import CarPlay

class CarPlayController: NSObject, CPTemplateApplicationDelegate {
    func templateApplication(_ app: CPApplication, didConnect interface: CPInterfaceController) {
        let tabs = CPTabBarTemplate(templates: [
            CPListTemplate(title: "Nexus", sections: [
                CPListSection(items: [
                    CPListItem(text: "Transactions", detailText: "View all"),
                    CPListItem(text: "Send Money", detailText: "Transfer"),
                    CPListItem(text: "Scan QR", detailText: "Quick pay")
                ])
            ]),
            CPListTemplate(title: "Recent", sections: [
                CPListSection(items: recentTransactions.map { tx in
                    CPListItem(text: tx.recipientName, detailText: tx.amountFormatted)
                })
            ])
        ])
        interface.setRootTemplate(tabs, animated: true)
    }
}
```

---

## Apple Watch

```swift
import SwiftUI
import WatchKit

struct WatchNexusApp: App {
    var body: some Scene {
        WindowGroup {
            NavigationView {
                List {
                    Section("Balance") { Text("$1,234.56").font(.title3.bold()) }
                    Section("Quick Actions") {
                        NavigationLink("Send Money") { SendMoneyView() }
                        NavigationLink("Recent") { RecentTransactionsView() }
                    }
                }
            }
        }
    }
}

// Complications
struct NexusComplication: Widget {
    var body: some WidgetConfiguration {
        StaticConfiguration(kind: "NexusComplication", provider: NexusComplicationProvider()) { entry in
            NexusComplicationView(entry: entry)
        }
        .configurationDisplayName("Nexus")
        .supportedFamilies([.accessoryCircular, .accessoryRectangular, .accessoryInline])
    }
}
```
