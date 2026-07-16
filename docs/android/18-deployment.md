# 18 — Deployment

## 1. Play Store Preparation

### 1.1 Store Listing

| Field                    | Character Limit | Description                              |
|--------------------------|----------------|------------------------------------------|
| App Title                | 30             | Short, memorable brand name              |
| Short Description        | 80             | Key value proposition                    |
| Full Description         | 4000           | Detailed feature overview                |
| Developer Name           | —              | Company / personal name                  |
| Website                  | —              | Support or product website URL           |
| Email                    | —              | Support contact email                    |
| Phone                    | Optional       | Support phone number                     |
| Privacy Policy URL       | —              | Required for data collection             |

### 1.2 Store Listing Content

```markdown
Title: NexusAI — AI Chat Assistant

Short Description:
Secure AI-powered chat with wallet integration and end-to-end encryption.

Full Description:
NexusAI is a secure, privacy-first AI chat application with integrated
cryptocurrency wallet support.

Features:
• AI-powered conversational assistant
• End-to-end encrypted messaging
• Built-in USDC wallet management
• Multi-conversation support
• WebSocket real-time messaging
• Biometric authentication
• Cross-device sync

Security:
• E2E encryption using X25519 key exchange
• Biometric lock for sensitive operations
• Certificate pinning for API security
• SOC 2 compliant infrastructure

Privacy:
• We do not sell personal data
• Messages encrypted at rest and in transit
• Minimal data collection policy
```

### 1.3 Graphics Requirements

```
Play Store Graphics:
┌──────────────────────────────────────────────────────────┐
│  Asset               │ Size              │ Format         │
├──────────────────────────────────────────────────────────┤
│  App Icon            │ 512x512           │ PNG (32-bit)   │
│  Feature Graphic     │ 1024x500          │ PNG/JPEG       │
│  Phone Screenshots   │ 16:9 or 9:16      │ PNG/JPEG       │
│  7-inch Tablet       │ 16:9 or 9:16      │ PNG/JPEG       │
│  10-inch Tablet      │ 16:9 or 9:16      │ PNG/JPEG       │
│  Wear Screenshots    │ 384x384           │ PNG/JPEG       │
│  TV Banner           │ 320x180           │ PNG            │
│  TV Screenshots      │ 1920x1080         │ PNG/JPEG       │
└──────────────────────────────────────────────────────────┘
```

---

## 2. Signing Configuration

### 2.1 Keystore Setup

```kotlin
// build.gradle.kts
android {
    signingConfigs {
        create("release") {
            storeFile = file(System.getenv("KEYSTORE_PATH") ?: "release.keystore")
            storePassword = System.getenv("KEYSTORE_PASSWORD")
            keyAlias = System.getenv("KEY_ALIAS")
            keyPassword = System.getenv("KEY_PASSWORD")
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            signingConfig = signingConfigs.getByName("release")
        }
    }
}
```

### 2.2 Google Play App Signing

```
┌──────────────────────────────────────────────────────────┐
│                Signing Key Architecture                  │
│                                                          │
│  Developer Machine        Google Play                    │
│  ┌──────────────┐        ┌──────────────┐               │
│  │ Upload Key   │───────►│ App Signing  │               │
│  │ (your local) │        │ Key (Google) │               │
│  └──────────────┘        └──────┬───────┘               │
│                                 │                        │
│                                 ▼                        │
│                          ┌──────────────┐               │
│                          │ Signed APK/  │               │
│                          │ AAB for user │               │
│                          └──────────────┘               │
│                                                          │
│  Benefits:                                               │
│  • Google manages the app signing key                   │
│  • Upload key can be rotated if compromised             │
│  • Play App Signing enables App Bundle + splits         │
│  • Key upgrade available for older apps                 │
└──────────────────────────────────────────────────────────┘
```

### 2.3 Signing Security

| Practice                              | Description                          |
|---------------------------------------|--------------------------------------|
| Never commit keystore to git          | Use `.gitignore` + CI secrets        |
| Use separate keys per environment    | staging.keystore, release.keystore   |
| Enable Play App Signing              | Google manages production key        |
| Rotate upload key periodically        | Play Console → Setup → App signing   |
| Use hardware-backed keys (KMS)       | Google Cloud KMS for CI/CD           |
| Require 2+ approvals for key access   | Team policy                          |

---

## 3. Build Variants

```kotlin
android {
    buildFeatures {
        buildConfig = true
    }

    buildTypes {
        debug {
            isDebuggable = true
            applicationIdSuffix = ".debug"
            versionNameSuffix = "-debug"
            buildConfigField("String", "API_BASE_URL", "\"https://dev-api.nexusai.com\"")
            buildConfigField("String", "WS_URL", "\"wss://dev-ws.nexusai.com\"")
            buildConfigField("boolean", "ENABLE_ANALYTICS", "false")
        }

        staging {
            initWith(buildTypes.getByName("release"))
            isDebuggable = true
            applicationIdSuffix = ".staging"
            versionNameSuffix = "-staging"
            buildConfigField("String", "API_BASE_URL", "\"https://staging-api.nexusai.com\"")
            buildConfigField("String", "WS_URL", "\"wss://staging-ws.nexusai.com\"")
            buildConfigField("boolean", "ENABLE_ANALYTICS", "true")
        }

        release {
            isMinifyEnabled = true
            isShrinkResources = true
            signingConfig = signingConfigs.getByName("release")
            buildConfigField("String", "API_BASE_URL", "\"https://api.nexusai.com\"")
            buildConfigField("String", "WS_URL", "\"wss://ws.nexusai.com\"")
            buildConfigField("boolean", "ENABLE_ANALYTICS", "true")
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }

    productFlavors {
        create("free") {
            dimension = "tier"
            applicationIdSuffix = ""
            buildConfigField("int", "MAX_CONVERSATIONS", "5")
            buildConfigField("boolean", "PREMIUM_FEATURES", "false")
        }
        create("premium") {
            dimension = "tier"
            applicationIdSuffix = ".premium"
            buildConfigField("int", "MAX_CONVERSATIONS", "-1")
            buildConfigField("boolean", "PREMIUM_FEATURES", "true")
        }
    }
}
```

### Variant Matrix

| Variant          | Minify | Debuggable | Suffix       | API URL        |
|------------------|--------|-----------|--------------|----------------|
| freeDebug        | No     | Yes       | .debug       | dev            |
| freeStaging      | Yes    | Yes       | .staging     | staging        |
| freeRelease      | Yes    | No        | —            | production     |
| premiumDebug     | No     | Yes       | .premium.debug | dev          |
| premiumStaging   | Yes    | Yes       | .premium.staging | staging    |
| premiumRelease   | Yes    | No        | .premium     | production     |

---

## 4. Version Management

```kotlin
// build.gradle.kts
android {
    defaultConfig {
        versionCode = calculateVersionCode()  // Auto-increment
        versionName = "1.4.2"                  // Semantic versioning
    }
}

fun calculateVersionCode(): Int {
    val versionFile = file("version.properties")
    val props = java.util.Properties().apply {
        if (versionFile.exists()) load(versionFile.inputStream())
    }

    val major = props.getProperty("MAJOR", "1").toInt()
    val minor = props.getProperty("MINOR", "0").toInt()
    val patch = props.getProperty("PATCH", "0").toInt()
    val build = props.getProperty("BUILD", "0").toInt()

    // Format: MMmmPPB (e.g., 1.4.2.3 = 10402003)
    return major * 10_000_000 + minor * 100_000 + patch * 1000 + build
}
```

### Semantic Versioning

```
Format: MAJOR.MINOR.PATCH-BUILD

Examples:
  1.0.0-1    → Initial release
  1.1.0-1    → New feature (backward compatible)
  1.1.1-1    → Bug fix
  2.0.0-1    → Breaking changes
  1.2.0-rc1  → Release candidate
  1.2.0-beta1→ Beta release

versionCode:
  MAJOR * 10000000 + MINOR * 100000 + PATCH * 1000 + BUILD

  1.0.0-1   → 10000001
  1.1.0-1   → 10100001
  1.1.1-1   → 10101001
  2.0.0-1   → 20000001
```

---

## 5. GitHub Actions CI/CD

```yaml
# .github/workflows/android.yml
name: Android CI/CD

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  GRADLE_OPTS: "-Dorg.gradle.daemon=false"

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          java-version: '17'
          distribution: 'temurin'
      - uses: gradle/actions/setup-gradle@v3
      - run: ./gradlew lintDebug
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: lint-results
          path: build/reports/lint-results-debug.html

  test:
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          java-version: '17'
          distribution: 'temurin'
      - uses: gradle/actions/setup-gradle@v3
      - run: ./gradlew testDebugUnitTest
      - run: ./gradlew jacocoTestReport
      - uses: codecov/codecov-action@v3
        with:
          files: build/reports/jacoco/jacocoTestReport.xml

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          java-version: '17'
          distribution: 'temurin'
      - uses: gradle/actions/setup-gradle@v3

      - name: Build AAB
        run: ./gradlew bundleRelease

      - name: Build APK
        run: ./gradlew assembleRelease

      - uses: actions/upload-artifact@v3
        with:
          name: release-artifacts
          path: |
            app/build/outputs/bundle/release/*.aab
            app/build/outputs/apk/release/*.apk

  deploy-staging:
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/develop'
    environment: staging
    steps:
      - uses: actions/download-artifact@v3
        with:
          name: release-artifacts

      - name: Deploy to Internal Track
        uses: r0adkll/upload-google-play@v1
        with:
          serviceAccountJsonPlainText: ${{ secrets.PLAY_SERVICE_ACCOUNT }}
          packageName: com.nexusai.app
          releaseFiles: app/build/outputs/bundle/release/app-release.aab
          track: internal
          status: completed

  deploy-production:
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main'
    environment: production
    steps:
      - uses: actions/download-artifact@v3
        with:
          name: release-artifacts

      - name: Deploy to Production
        uses: r0adkll/upload-google-play@v1
        with:
          serviceAccountJsonPlainText: ${{ secrets.PLAY_SERVICE_ACCOUNT }}
          packageName: com.nexusai.app
          releaseFiles: app/build/outputs/bundle/release/app-release.aab
          track: production
          status: completed
          rolloutFraction: 0.1
```

---

## 6. Fastlane Integration

### 6.1 Fastfile

```ruby
# fastlane/Fastfile
default_platform(:android)

platform :android do
  desc "Build and deploy to internal track"
  lane :internal do
    gradle(
      task: "bundle",
      build_type: "Release",
      project_dir: "./"
    )

    upload_to_play_store(
      track: "internal",
      aab: "app/build/outputs/bundle/release/app-release.aab",
      skip_upload_metadata: true,
      skip_upload_images: true,
      skip_upload_screenshots: true
    )
  end

  desc "Build and deploy to production"
  lane :production do
    gradle(
      task: "bundle",
      build_type: "Release",
      project_dir: "./"
    )

    upload_to_play_store(
      track: "production",
      aab: "app/build/outputs/bundle/release/app-release.aab",
      rollout_fraction: 0.1,
      skip_upload_metadata: true,
      skip_upload_images: true,
      skip_upload_screenshots: true
    )
  end

  desc "Promote internal to production"
  lane :promote_to_production do
    upload_to_play_store(
      track: "internal",
      track_promote_to: "production",
      version_code: 42
    )
  end

  desc "Run tests"
  lane :test do
    gradle(task: "testDebugUnitTest")
    gradle(task: "lintDebug")
  end
end
```

### 6.2 Appfile

```ruby
# fastlane/Appfile
json_key_file("path/to/play-store-key.json")
package_name("com.nexusai.app")
```

---

## 7. Deployment Tracks

```
Deployment Pipeline:
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │ Internal │───►│ Closed   │───►│ Open     │───►│Production│  │
│  │ Track    │    │ Track    │    │ Track    │    │ Track    │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│       │              │               │               │          │
│   Testers        Alpha/Beta      Public Beta    100% users     │
│   (10-100)       (100-1000)     (1000-10000)   (unlimited)    │
│                                                                  │
│  Timeline: ──────────────────────────────────────────────►      │
│            Day 1    Day 3-7    Day 7-14    Day 14+             │
└──────────────────────────────────────────────────────────────────┘
```

### Staged Rollout

| Stage       | Percentage | Duration | Monitoring                  |
|-------------|-----------|----------|-----------------------------|
| Internal    | 0%        | 1-3 days | Crash rate, ANR, manual QA  |
| Closed Beta | 1-5%      | 3-7 days | Crash rate, ratings, ANR    |
| Open Beta   | 5-20%     | 7-14 days| Ratings, reviews, crash     |
| Production  | 20-100%   | 14+ days | All metrics, revenue        |

```yaml
# Staged rollout via GitHub Actions
- name: Staged rollout
  uses: r0adkll/upload-google-play@v1
  with:
    serviceAccountJsonPlainText: ${{ secrets.PLAY_SERVICE_ACCOUNT }}
    packageName: com.nexusai.app
    releaseFiles: app-release.aab
    track: production
    status: inProgress
    rolloutFraction: 0.25
```

---

## 8. App Bundle & Dynamic Features

### 8.1 Dynamic Feature Modules

```kotlin
// build.gradle.kts (dynamic feature module)
plugins {
    id("com.android.dynamic-feature")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "com.nexusai.feature.premium"
}

dependencies {
    implementation(project(":app"))
    implementation("androidx.core:core-ktx:1.12.0")
}
```

```
Dynamic Feature Delivery:
┌──────────────────────────────────────────────────────────────┐
│  Module              │ Delivery     │ Size    │ When          │
├──────────────────────────────────────────────────────────────┤
│  :app                │ Install-time │ 15MB    │ Always        │
│  :feature:chat       │ Install-time │ 3MB     │ Always        │
│  :feature:wallet     │ On-demand    │ 4MB     │ First use     │
│  :feature:premium    │ On-demand    │ 2MB     │ Subscription  │
│  :feature:settings   │ Install-time │ 1MB     │ Always        │
└──────────────────────────────────────────────────────────────┘
```

```kotlin
// On-demand module loading
class DynamicFeatureLoader(private val context: Context) {

    suspend fun loadPremiumFeature(): Result<PremiumFeature> {
        return try {
            val intent = Intent()
            intent.action = "com.nexusai.feature.premium.LAUNCH"
            intent.setPackage(context.packageName)

            val installRequest = SplitInstallRequest.newBuilder()
                .addModule("premium")
                .build()

            SplitInstallManagerFactory.create().startInstall(installRequest)

            Result.success(PremiumFeature())
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
```

---

## 9. Play Store Review Process

### 9.1 Common Rejection Reasons

| Reason                        | Prevention                                |
|-------------------------------|-------------------------------------------|
| Crash on launch               | Test on clean install, multiple devices   |
| Missing privacy policy        | Add URL in Play Console                   |
| Permissions without disclosure| Declare in data safety section            |
| Inappropriate content         | Follow content guidelines                 |
| Misleading description        | Accurate store listing                    |
| In-app purchase issues        | Proper billing implementation             |
| API level too low             | Target latest API, test on minimum        |
| WebView without controls      | Use Chrome Custom Tabs                    |

### 9.2 Review Checklist

```
Pre-Submission Checklist:
┌──────────────────────────────────────────────────────────────┐
│  □ App runs without crashes on clean install                │
│  □ Target SDK is latest (API 34+)                           │
│  □ Privacy policy URL is live and accessible                │
│  □ Data safety section is accurate                          │
│  □ Content rating questionnaire completed                   │
│  □ Screenshots match current app UI                         │
│  □ Store description is accurate                            │
│  □ No placeholder text or TODO comments                     │
│  □ All permissions have in-app disclosure                   │
│  □ App icons render correctly                               │
│  □ Tested on phone + tablet (if applicable)                 │
│  □ Tested on API 24 (minSdk) and API 34                    │
│  □ Sensitive permissions have justification                 │
│  □ Feature graphic is high quality                          │
│  □ Contact information is valid                             │
└──────────────────────────────────────────────────────────────┘
```

---

## 10. Play Store Updates

### 10.1 In-App Update API

```kotlin
class UpdateManager(private val activity: ComponentActivity) {

    private val appUpdateManager = AppUpdateManagerFactory.create(activity)

    fun checkForUpdate() {
        appUpdateManager.appUpdateInfo.addOnSuccessListener { info ->
            when {
                info.updateAvailability() == UpdateAvailability.UPDATE_AVAILABLE -> {
                    if (info.isImmediateUpdateAllowed) {
                        startImmediateUpdate(info)
                    } else if (info.isFlexibleUpdateAllowed) {
                        startFlexibleUpdate(info)
                    }
                }
                info.installStatus() == InstallStatus.DOWNLOADED -> {
                    showUpdateInstalledSnackbar()
                }
            }
        }
    }

    private fun startImmediateUpdate(info: AppUpdateInfo) {
        appUpdateManager.startUpdateForResult(
            info,
            REQUEST_CODE_IMMEDIATE,
            activity
        )
    }

    private fun startFlexibleUpdate(info: AppUpdateInfo) {
        appUpdateManager.startUpdateFlowForResult(
            info,
            AppUpdateType.FLEXIBLE,
            activity,
            REQUEST_CODE_FLEXIBLE
        )
    }

    private fun showUpdateInstalledSnackbar() {
        appUpdateManager.completeUpdate()
    }
}
```

### 10.2 Update Types

| Type         | Behavior                     | UX Impact        | Use When               |
|-------------|------------------------------|------------------|------------------------|
| Immediate   | Blocks until downloaded      | Low (fast DL)    | Critical security fix  |
| Flexible    | Downloads in background      | None             | Feature updates        |
| Postponable | Can defer for 30 days        | None             | Non-critical updates   |

---

## 11. Crash Reporting & Analytics

### 11.1 Firebase Crashlytics

```kotlin
// build.gradle.kts
implementation("com.google.firebase:firebase-crashlytics-ktx:18.6.0")
implementation("com.google.firebase:firebase-analytics-ktx:21.5.0")

// Application setup
class NexusApplication : Application() {
    override fun onCreate() {
        super.onCreate()

        FirebaseCrashlytics.getInstance().apply {
            setCrashlyticsCollectionEnabled(!BuildConfig.DEBUG)
        }
    }
}

// Custom crash reporting
suspend fun sendTransaction(request: TransactionRequest): Result<Unit> {
    return try {
        val result = transactionRepository.create(request)
        FirebaseCrashlytics.getInstance().apply {
            log("Transaction created: ${request.amount} ${request.currency}")
            setCustomKey("transaction_amount", request.amount.toDouble())
        }
        result
    } catch (e: Exception) {
        FirebaseCrashlytics.getInstance().apply {
            recordException(e)
            setCustomKey("failed_amount", request.amount.toDouble())
        }
        Result.failure(e)
    }
}
```

### 11.2 Crash Rate Monitoring

```
Crash Rate Thresholds:
┌──────────────────────────────────────────────────────────────┐
│  Metric              │ Healthy  │ Warning  │ Critical       │
├──────────────────────────────────────────────────────────────┤
│  Crash-free users    │ >99.5%   │ 98-99.5% │ <98%          │
│  ANR-free users      │ >99.9%   │ 99-99.9% │ <99%          │
│  Crash sessions/day  │ <50      │ 50-200   │ >200          │
└──────────────────────────────────────────────────────────────┘
```

---

## 12. Remote Config & Feature Flags

```kotlin
class FeatureFlags(private val remoteConfig: FirebaseRemoteConfig) {

    suspend fun init() {
        remoteConfig.setConfigSettingsAsync(
            RemoteConfigSettings.Builder()
                .setMinimumFetchIntervalInSeconds(
                    if (BuildConfig.DEBUG) 0 else 3600
                )
                .build()
        )

        remoteConfig.setDefaultsAsync(R.xml.remote_config_defaults)
        remoteConfig.fetchAndActivate()
    }

    val isPremiumEnabled: Boolean
        get() = remoteConfig.getBoolean("premium_enabled")

    val maxConversations: Int
        get() = remoteConfig.getLong("max_conversations").toInt()

    val maintenanceMode: Boolean
        get() = remoteConfig.getBoolean("maintenance_mode")

    val walletFeatureEnabled: Boolean
        get() = remoteConfig.getBoolean("wallet_feature_enabled")

    suspend fun getString(key: String): String {
        return remoteConfig.getString(key)
    }
}
```

---

## 13. A/B Testing

```kotlin
class ExperimentManager(private val firebase: Firebase) {

    fun logExperiment(experimentId: String, variant: String) {
        val bundle = Bundle().apply {
            putString("experiment_id", experimentId)
            putString("variant", variant)
        }
        firebase.analytics.logEvent("experiment_exposure", bundle)
    }

    fun isVariant(experimentId: String, variant: String): Boolean {
        val remoteConfig = FirebaseRemoteConfig.getInstance()
        return remoteConfig.getString(experimentId) == variant
    }

    // Example: Chat bubble style experiment
    fun getChatBubbleStyle(): ChatBubbleStyle {
        return when {
            isVariant("chat_bubble_style", "rounded") -> ChatBubbleStyle.ROUNDED
            isVariant("chat_bubble_style", "modern") -> ChatBubbleStyle.MODERN
            else -> ChatBubbleStyle.CLASSIC
        }
    }
}
```

---

## 14. Post-Launch Monitoring

```
Post-Launch Dashboard:
┌──────────────────────────────────────────────────────────────┐
│  Metric              │ 1hr  │ 24hr  │ 7-day │ 30-day      │
├──────────────────────────────────────────────────────────────┤
│  Crash Rate          │ 0.1% │ 0.2%  │ 0.3%  │ 0.25%       │
│  ANR Rate            │ 0.0% │ 0.05% │ 0.08% │ 0.06%       │
│  Install/Uninstall   │ —    │ 1.2   │ 1.1   │ 1.05        │
│  Rating              │ —    │ 4.5   │ 4.4   │ 4.3         │
│  Daily Active Users  │ 500  │ 2000  │ 3500  │ 4000        │
│  Session Duration    │ 5m   │ 4.5m  │ 4.8m  │ 5.2m        │
└──────────────────────────────────────────────────────────────┘
```

### Hotfix Process

```
Hotfix Workflow:
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Crash       │────►│  Fix & Test  │────►│  Fast-track  │
│  Detected    │     │  (< 2 hrs)   │     │  Review      │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                  │
                                                  ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Monitor     │◄────│  Rollout 10% │────►│  Approve     │
│  Crash Rate  │     │  (1-2 days)  │     │  100%        │
└──────────────┘     └──────────────┘     └──────────────┘
```

---

## 15. Play Store Policies

| Policy Area             | Key Requirements                                |
|------------------------|-------------------------------------------------|
| Data Safety            | Disclose all data collection and sharing        |
| Permissions            | Justify each permission request                 |
| Content Rating         | Accurate IARC rating                            |
| Target Audience        | Correct age group targeting                     |
| Privacy Policy         | Must be accessible and complete                 |
| Metadata               | Accurate screenshots, descriptions              |
| Monetization           | Clear pricing, no deceptive practices           |
| Security               | No malware, no exploits                         |
| Crashes                | Must not crash on launch or during use          |
| Family Policy          | If targeting children, COPPA compliance         |
| Gambling               | Real money gambling restrictions                 |
| Financial Services     | Licensing requirements for financial apps       |

### App Backup

```kotlin
// AndroidManifest.xml
<application
    android:allowBackup="true"
    android:fullBackupContent="@xml/backup_rules"
    android:dataExtractionRules="@xml/data_extraction_rules">

<!-- res/xml/backup_rules.xml -->
<full-backup-content>
    <include domain="sharedpref" path="settings.xml" />
    <include domain="database" path="chat.db" />
    <exclude domain="sharedpref" path="auth.xml" />
    <exclude domain="external" path="." />
</full-backup-content>

<!-- res/xml/data_extraction_rules.xml -->
<data-extraction-rules>
    <cloud-backup>
        <include domain="sharedpref" path="settings.xml" />
        <include domain="database" path="chat.db" />
    </cloud-backup>
    <device-transfer>
        <include domain="sharedpref" path="settings.xml" />
    </device-transfer>
</data-extraction-rules>
```

---

## 16. Build Automation Scripts

```bash
#!/bin/bash
# scripts/release.sh

set -euo pipefail

VERSION=$1
BRANCH=$(git rev-parse --abbrev-ref HEAD)

echo "🚀 Starting release v${VERSION} from branch ${BRANCH}"

# Validate
if [[ "$BRANCH" != "main" && "$BRANCH" != "release/"* ]]; then
    echo "❌ Releases must be from 'main' or 'release/*' branches"
    exit 1
fi

# Run checks
echo "📋 Running lint..."
./gradlew lintDebug
echo "🧪 Running tests..."
./gradlew testDebugUnitTest
echo "🔒 Running security check..."
./gradlew dependencyCheckAnalyze

# Build
echo "📦 Building AAB..."
./gradlew bundleRelease

# Upload to Play Store
echo "📤 Uploading to internal track..."
fastlane internal

echo "✅ Release v${VERSION} uploaded to internal track"
echo "   Next: Promote through tracks when ready"
```

```yaml
# .github/workflows/release.yml
name: Release
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version number (e.g., 1.4.2)'
        required: true
      track:
        description: 'Play Store track'
        required: true
        type: choice
        options:
          - internal
          - closed
          - open
          - production
      rollout:
        description: 'Rollout percentage (for production)'
        required: false
        default: '10'

jobs:
  release:
    runs-on: ubuntu-latest
    environment: ${{ github.event.inputs.track }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          java-version: '17'
      - uses: gradle/actions/setup-gradle@v3

      - name: Update version
        run: |
          echo "VERSION_NAME=${{ github.event.inputs.version }}" >> gradle.properties

      - name: Build
        run: ./gradlew bundleRelease

      - name: Upload to Play Store
        uses: r0adkll/upload-google-play@v1
        with:
          serviceAccountJsonPlainText: ${{ secrets.PLAY_SERVICE_ACCOUNT }}
          packageName: com.nexusai.app
          releaseFiles: app/build/outputs/bundle/release/app-release.aab
          track: ${{ github.event.inputs.track }}
          rolloutFraction: ${{ github.event.inputs.rollout }}
          status: ${{ github.event.inputs.track == 'production' && 'inProgress' || 'completed' }}
```
