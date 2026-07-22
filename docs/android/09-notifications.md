# 09 - Notifications

The notification system provides real-time alerts for AI responses, security events,
workflow updates, and system status. It spans FCM push notifications, local
notification channels, an in-app notification center, and user preferences.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Push Notification Architecture](#push-notification-architecture)
3. [FCM Token Management](#fcm-token-management)
4. [Notification Channels](#notification-channels)
5. [Notification Builder](#notification-builder)
6. [Notification Content](#notification-content)
7. [Notification Actions](#notification-actions)
8. [Notification Priority](#notification-priority)
9. [Notification Sound](#notification-sound)
10. [Notification Vibration](#notification-vibration)
11. [Notification LED](#notification-led)
12. [Notification Large Icon](#notification-large-icon)
13. [Notification Styles](#notification-styles)
14. [Notification Custom Layout](#notification-custom-layout)
15. [Notification Click Handling](#notification-click-handling)
16. [Notification Dismiss Handling](#notification-dismiss-handling)
17. [Notification Badge](#notification-badge)
18. [In-App Notification Center](#in-app-notification-center)
19. [Notification Preferences](#notification-preferences)
20. [Notification Data Models](#notification-data-models)
21. [Notification Caching](#notification-caching)
22. [Notification API Integration](#notification-api-integration)
23. [Notification Permissions](#notification-permissions)
24. [Notification on Different Android Versions](#notification-on-different-android-versions)
25. [Notification Testing](#notification-testing)
26. [Notification Accessibility](#notification-accessibility)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                     Server Side                                  │
│  Event System ──► FCM Admin SDK ──► FCM Server                  │
│                                            │                     │
└────────────────────────────────────────────┼─────────────────────┘
                                             │
                     ┌───────────────────────┘
                     ▼
┌──────────────────────────────────────────────────────────────────┐
│                     Android App                                  │
│                                                                  │
│  FirebaseMessagingService                                        │
│       │                                                          │
│       ├── TokenManager (register/refresh)                        │
│       ├── NotificationBuilder (create notifications)             │
│       ├── ChannelManager (create/update channels)                │
│       └── DeepLinkResolver (route to screens)                    │
│                                                                  │
│  Local                                                           │
│  ├── NotificationDao (Room)                                      │
│  ├── NotificationPreferences (DataStore)                         │
│  └── InAppNotificationCenter (Compose)                           │
│                                                                  │
│  NotificationCompat.Builder ──► NotificationManager ──► System   │
│  InAppNotificationCenter ──► LazyColumn (notification list)      │
└──────────────────────────────────────────────────────────────────┘
```

---

## Push Notification Architecture

### Server → FCM → Device Flow

```
┌─────────┐     ┌──────────┐     ┌─────────┐     ┌───────────┐
│ Server  │────►│ FCM HTTP │────►│  FCM    │────►│ Android   │
│ Event   │     │ v1 API   │     │ Server  │     │ Device    │
└─────────┘     └──────────┘     └─────────┘     └───────────┘
                                             │
                              ┌───────────────┘
                              ▼
                    FirebaseMessagingService
                              │
                    ┌─────────┴──────────┐
                    ▼                    ▼
              Show Local         Save to Room
              Notification      + Sync to Server
```

```kotlin
class FirebaseNotificationService : FirebaseMessagingService() {

    @Inject lateinit var tokenManager: FcmTokenManager
    @Inject lateinit var notificationBuilder: AppNotificationBuilder
    @Inject lateinit var channelManager: NotificationChannelManager
    @Inject lateinit var deepLinkResolver: DeepLinkResolver
    @Inject lateinit var notificationRepository: NotificationRepository

    override fun onNewToken(token: String) {
        super.onNewToken(token)
        CoroutineScope(Dispatchers.IO).launch {
            tokenManager.registerToken(token)
        }
    }

    override fun onMessageReceived(message: RemoteMessage) {
        super.onMessageReceived(message)

        val data = message.data
        val notificationType = data["type"] ?: "general"

        val notification = AppNotification(
            id = data["id"] ?: UUID.randomUUID().toString(),
            type = NotificationType.fromString(notificationType),
            title = message.notification?.title ?: data["title"] ?: "",
            body = message.notification?.body ?: data["body"] ?: "",
            deepLink = data["deepLink"] ?: "",
            priority = data["priority"] ?: "default",
            metadata = data.filterKeys { it !in setOf("id", "type", "title", "body", "deepLink", "priority") },
            receivedAt = Instant.now(),
            isRead = false
        )

        // Save to local database
        CoroutineScope(Dispatchers.IO).launch {
            notificationRepository.insertNotification(notification)
        }

        // Show system notification
        val channel = channelManager.getChannel(notification.type)
        notificationBuilder.show(notification, channel)

        // Update badge
        notificationRepository.incrementUnreadCount()
    }
}
```

---

## FCM Token Management

```kotlin
class FcmTokenManager @Inject constructor(
    private val apiService: NotificationApiService,
    private val dataStore: DataStore<Preferences>,
    @IoDispatcher private val ioDispatcher: CoroutineDispatcher
) {
    companion object {
        private val KEY_FCM_TOKEN = stringPreferencesKey("fcm_token")
        private val KEY_TOKEN_SYNCED = booleanPreferencesKey("fcm_token_synced")
    }

    suspend fun registerToken(token: String) {
        // Save locally
        dataStore.edit { prefs ->
            prefs[KEY_FCM_TOKEN] = token
            prefs[KEY_TOKEN_SYNCED] = false
        }

        // Sync to server
        try {
            apiService.registerFcmToken(FcmTokenRequest(token = token, platform = "android"))
            dataStore.edit { it[KEY_TOKEN_SYNCED] = true }
        } catch (e: Exception) {
            // Will retry on next app launch
        }
    }

    suspend fun getToken(): String? {
        return dataStore.data.first()[KEY_FCM_TOKEN]
    }

    suspend fun syncIfNeeded() {
        val synced = dataStore.data.first()[KEY_TOKEN_SYNCED] ?: false
        if (!synced) {
            val token = getToken() ?: return
            try {
                apiService.registerFcmToken(FcmTokenRequest(token = token, platform = "android"))
                dataStore.edit { it[KEY_TOKEN_SYNCED] = true }
            } catch (_: Exception) {}
        }
    }

    suspend fun unregisterToken() {
        val token = getToken() ?: return
        try {
            apiService.unregisterFcmToken(FcmTokenRequest(token = token, platform = "android"))
        } catch (_: Exception) {}
        dataStore.edit {
            it[KEY_FCM_TOKEN] = null
            it[KEY_TOKEN_SYNCED] = false
        }
    }
}

data class FcmTokenRequest(
    val token: String,
    val platform: String
)
```

---

## Notification Channels

```kotlin
class NotificationChannelManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun createAllChannels() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val manager = context.getSystemService(NotificationManager::class.java)

            val channels = listOf(
                createChannel(
                    id = CHANNEL_AI_ALERTS,
                    name = "AI Alerts",
                    description = "Notifications for AI responses and completions",
                    importance = NotificationManager.IMPORTANCE_HIGH,
                    sound = Uri.parse("android.resource://${context.packageName}/raw/notification_ai"),
                    vibration = longArrayOf(0, 200),
                    enableLights = true,
                    lightColor = Color.BLUE
                ),
                createChannel(
                    id = CHANNEL_SECURITY,
                    name = "Security",
                    description = "Security alerts and suspicious activity",
                    importance = NotificationManager.IMPORTANCE_HIGH,
                    sound = Uri.parse("android.resource://${context.packageName}/raw/notification_security"),
                    vibration = longArrayOf(0, 100, 100, 100),
                    enableLights = true,
                    lightColor = Color.RED
                ),
                createChannel(
                    id = CHANNEL_WORKFLOW,
                    name = "Workflow Updates",
                    description = "Agent execution and workflow status updates",
                    importance = NotificationManager.IMPORTANCE_DEFAULT,
                    vibration = longArrayOf(0, 100),
                    enableLights = true,
                    lightColor = Color.GREEN
                ),
                createChannel(
                    id = CHANNEL_DOCUMENTS,
                    name = "Document Processing",
                    description = "Document upload and processing status",
                    importance = NotificationManager.IMPORTANCE_LOW
                ),
                createChannel(
                    id = CHANNEL_GENERAL,
                    name = "General",
                    description = "General notifications and announcements",
                    importance = NotificationManager.IMPORTANCE_DEFAULT
                ),
                createChannel(
                    id = CHANNEL_SYSTEM,
                    name = "System",
                    description = "System status and maintenance notices",
                    importance = NotificationManager.IMPORTANCE_LOW
                )
            )

            manager.createNotificationChannels(channels)
        }
    }

    @RequiresApi(Build.VERSION_CODES.O)
    private fun createChannel(
        id: String,
        name: String,
        description: String,
        importance: Int,
        sound: Uri? = null,
        vibration: LongArray? = null,
        enableLights: Boolean = false,
        lightColor: Int = 0
    ): NotificationChannel {
        return NotificationChannel(id, name, importance).apply {
            this.description = description
            sound?.let {
                val audioAttributes = AudioAttributes.Builder()
                    .setUsage(AudioAttributes.USAGE_NOTIFICATION)
                    .setContentType(AudioAttributes.CONTENT_TYPE_SONIFICATION)
                    .build()
                setSound(it, audioAttributes)
            }
            vibration?.let { vibrationPattern = it }
            enableLights(enableLights)
            if (enableLights) { this.lightColor = lightColor }
        }
    }

    fun getChannel(type: NotificationType): String = when (type) {
        NotificationType.AI_ALERT -> CHANNEL_AI_ALERTS
        NotificationType.SECURITY -> CHANNEL_SECURITY
        NotificationType.WORKFLOW -> CHANNEL_WORKFLOW
        NotificationType.DOCUMENT -> CHANNEL_DOCUMENTS
        NotificationType.GENERAL -> CHANNEL_GENERAL
        NotificationType.SYSTEM -> CHANNEL_SYSTEM
    }

    companion object {
        const val CHANNEL_AI_ALERTS = "ai_alerts"
        const val CHANNEL_SECURITY = "security"
        const val CHANNEL_WORKFLOW = "workflow"
        const val CHANNEL_DOCUMENTS = "documents"
        const val CHANNEL_GENERAL = "general"
        const val CHANNEL_SYSTEM = "system"
    }
}
```

### Channel Configuration Table

| Channel       | Importance | Sound        | Vibration      | Lights | Default ON |
|---------------|------------|--------------|----------------|--------|------------|
| AI Alerts     | HIGH       | Custom AI    | [0, 200]ms     | Blue   | Yes        |
| Security      | HIGH       | Custom Sec   | [0, 100,100,100]ms | Red | Yes     |
| Workflow      | DEFAULT    | Default      | [0, 100]ms     | Green  | Yes        |
| Documents     | LOW        | None         | None           | No     | Yes        |
| General       | DEFAULT    | Default      | None           | No     | Yes        |
| System        | LOW        | None         | None           | No     | No         |

---

## Notification Builder

```kotlin
class AppNotificationBuilder @Inject constructor(
    @ApplicationContext private val context: Context,
    private val deepLinkResolver: DeepLinkResolver
) {
    private val notificationManager: NotificationManagerCompat =
        NotificationManagerCompat.from(context)

    fun show(notification: AppNotification, channelId: String) {
        val pendingIntent = deepLinkResolver.createPendingIntent(notification.deepLink)

        val builder = NotificationCompat.Builder(context, channelId)
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle(notification.title)
            .setContentText(notification.body)
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)
            .setGroup(notification.type.groupKey)
            .setPriority(getPriority(notification.priority))
            .setCategory(getCategory(notification.type))
            .setVisibility(getVisibility(notification.type))
            .setColor(getColor(notification.type))
            .setLargeIcon(loadLargeIcon(notification))
            .setWhen(notification.receivedAt.toEpochMilli())
            .setShowWhen(true)

        // Style based on type
        when (notification.type) {
            NotificationType.AI_ALERT -> {
                builder.setStyle(
                    NotificationCompat.BigTextStyle()
                        .bigText(notification.body)
                        .setBigContentTitle(notification.title)
                )
                addAiAlertActions(builder, notification)
            }
            NotificationType.WORKFLOW -> {
                builder.setStyle(
                    NotificationCompat.InboxStyle()
                        .setBigContentTitle(notification.title)
                )
            }
            NotificationType.SECURITY -> {
                builder.setOngoing(true) // Cannot be swiped away
                addSecurityActions(builder, notification)
            }
            else -> {
                builder.setStyle(
                    NotificationCompat.BigTextStyle().bigText(notification.body)
                )
            }
        }

        // Custom layout for certain types
        if (notification.type == NotificationType.AI_ALERT && notification.metadata.containsKey("modelName")) {
            val customView = RemoteViews(context.packageName, R.layout.notification_ai_response)
            customView.setTextViewText(R.id.title, notification.title)
            customView.setTextViewText(R.id.body, notification.body)
            customView.setTextViewText(R.id.modelName, notification.metadata["modelName"] ?: "")
            builder.setCustomContentView(customView)
        }

        try {
            notificationManager.notify(notification.id.hashCode(), builder.build())
        } catch (e: SecurityException) {
            // Notification permission not granted
        }
    }

    private fun getPriority(priority: String): Int = when (priority) {
        "high" -> NotificationCompat.PRIORITY_HIGH
        "low" -> NotificationCompat.PRIORITY_LOW
        "min" -> NotificationCompat.PRIORITY_MIN
        else -> NotificationCompat.PRIORITY_DEFAULT
    }

    private fun getCategory(type: NotificationType): String = when (type) {
        NotificationType.AI_ALERT -> NotificationCompat.CATEGORY_MESSAGE
        NotificationType.SECURITY -> NotificationCompat.CATEGORY_ALARM
        NotificationType.WORKFLOW -> NotificationCompat.CATEGORY_PROGRESS
        NotificationType.DOCUMENT -> NotificationCompat.CATEGORY_STATUS
        NotificationType.GENERAL -> NotificationCompat.CATEGORY_SOCIAL
        NotificationType.SYSTEM -> NotificationCompat.CATEGORY_SYSTEM
    }

    private fun getVisibility(type: NotificationType): Int = when (type) {
        NotificationType.SECURITY -> NotificationCompat.VISIBILITY_PUBLIC
        NotificationType.AI_ALERT -> NotificationCompat.VISIBILITY_PRIVATE
        else -> NotificationCompat.VISIBILITY_PRIVATE
    }

    private fun getColor(type: NotificationType): Int = when (type) {
        NotificationType.AI_ALERT -> ContextCompat.getColor(context, R.color.notification_ai)
        NotificationType.SECURITY -> ContextCompat.getColor(context, R.color.notification_security)
        NotificationType.WORKFLOW -> ContextCompat.getColor(context, R.color.notification_workflow)
        else -> ContextCompat.getColor(context, R.color.notification_default)
    }

    private fun loadLargeIcon(notification: AppNotification): Bitmap? {
        return when (notification.type) {
            NotificationType.AI_ALERT -> {
                BitmapFactory.decodeResource(context.resources, R.drawable.ic_ai_large)
            }
            NotificationType.SECURITY -> {
                BitmapFactory.decodeResource(context.resources, R.drawable.ic_security_large)
            }
            else -> null
        }
    }

    fun cancelAll() {
        notificationManager.cancelAll()
    }

    fun cancel(notificationId: String) {
        notificationManager.cancel(notificationId.hashCode())
    }
}
```

---

## Notification Content

```kotlin
data class AppNotification(
    val id: String,
    val type: NotificationType,
    val title: String,
    val body: String,
    val deepLink: String,
    val priority: String = "default",
    val metadata: Map<String, String> = emptyMap(),
    val receivedAt: Instant = Instant.now(),
    val isRead: Boolean = false
)

enum class NotificationType(val channelName: String, val groupKey: String) {
    AI_ALERT("AI Alerts", "group_ai"),
    SECURITY("Security", "group_security"),
    WORKFLOW("Workflow Updates", "group_workflow"),
    DOCUMENT("Document Processing", "group_documents"),
    GENERAL("General", "group_general"),
    SYSTEM("System", "group_system");

    companion object {
        fun fromString(value: String): NotificationType {
            return entries.find { it.name.equals(value, ignoreCase = true) } ?: GENERAL
        }
    }
}
```

---

## Notification Actions

```kotlin
private fun addAiAlertActions(
    builder: NotificationCompat.Builder,
    notification: AppNotification
) {
    val viewIntent = deepLinkResolver.createPendingIntent(notification.deepLink)
    builder.addAction(
        R.drawable.ic_notification_action,
        "View",
        viewIntent
    )

    val dismissIntent = createDismissPendingIntent(notification.id)
    builder.addAction(
        R.drawable.ic_notification_dismiss,
        "Dismiss",
        dismissIntent
    )
}

private fun addSecurityActions(
    builder: NotificationCompat.Builder,
    notification: AppNotification
) {
    // Approve action
    val approveIntent = createActionPendingIntent(
        notification.id, NotificationAction.APPROVE
    )
    builder.addAction(
        R.drawable.ic_approve,
        "Approve",
        approveIntent
    )

    // Reject action
    val rejectIntent = createActionPendingIntent(
        notification.id, NotificationAction.REJECT
    )
    builder.addAction(
        R.drawable.ic_reject,
        "Reject",
        rejectIntent
    )

    // View details
    val viewIntent = deepLinkResolver.createPendingIntent(notification.deepLink)
    builder.addAction(
        R.drawable.ic_notification_action,
        "View Details",
        viewIntent
    )
}

private fun createActionPendingIntent(
    notificationId: String,
    action: NotificationAction
): PendingIntent {
    val intent = Intent(context, NotificationActionReceiver::class.java).apply {
        action = action.name
        putExtra("notification_id", notificationId)
    }
    return PendingIntent.getBroadcast(
        context,
        notificationId.hashCode() + action.ordinal,
        intent,
        PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
    )
}

private fun createDismissPendingIntent(notificationId: String): PendingIntent {
    val intent = Intent(context, NotificationDismissReceiver::class.java).apply {
        putExtra("notification_id", notificationId)
    }
    return PendingIntent.getBroadcast(
        context,
        notificationId.hashCode(),
        intent,
        PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
    )
}

enum class NotificationAction { APPROVE, REJECT, DISMISS }
```

### Broadcast Receivers

```kotlin
class NotificationActionReceiver : BroadcastReceiver() {
    @Inject lateinit var notificationRepository: NotificationRepository

    override fun onReceive(context: Context, intent: Intent) {
        val notificationId = intent.getStringExtra("notification_id") ?: return
        val action = intent.action ?: return

        goAsync().let { pendingResult ->
            CoroutineScope(Dispatchers.IO).launch {
                try {
                    when (NotificationAction.valueOf(action)) {
                        NotificationAction.APPROVE -> {
                            notificationRepository.handleAction(notificationId, "approve")
                        }
                        NotificationAction.REJECT -> {
                            notificationRepository.handleAction(notificationId, "reject")
                        }
                        NotificationAction.DISMISS -> {
                            notificationRepository.markDismissed(notificationId)
                        }
                    }
                } finally {
                    pendingResult.finish()
                }
            }
        }
    }
}

class NotificationDismissReceiver : BroadcastReceiver() {
    @Inject lateinit var notificationRepository: NotificationRepository

    override fun onReceive(context: Context, intent: Intent) {
        val notificationId = intent.getStringExtra("notification_id") ?: return
        goAsync().let { pendingResult ->
            CoroutineScope(Dispatchers.IO).launch {
                try {
                    notificationRepository.markDismissed(notificationId)
                } finally {
                    pendingResult.finish()
                }
            }
        }
    }
}
```

---

## Notification Priority

| Priority | Use Case                   | Behavior                               |
|----------|----------------------------|----------------------------------------|
| HIGH     | Security alerts, AI errors | Heads-up, sound, vibration             |
| DEFAULT  | Workflow updates, AI replies | Status bar, sound per channel       |
| LOW      | Document processing        | Status bar, no sound                   |
| MIN      | Background sync            | Collapsed in shade                     |

---

## Notification Sound

Custom notification sounds per channel:

```kotlin
// Raw resource files
// res/raw/notification_ai.ogg     - Short chime for AI responses
// res/raw/notification_security.ogg - Alert tone for security
// res/raw/notification_workflow.ogg - Soft tone for workflow

// Programmatic sound setting (Android 8.0+)
if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
    val channel = NotificationChannel(
        CHANNEL_AI_ALERTS,
        "AI Alerts",
        NotificationManager.IMPORTANCE_HIGH
    ).apply {
        setSound(
            Uri.parse("android.resource://${context.packageName}/raw/notification_ai"),
            AudioAttributes.Builder()
                .setUsage(AudioAttributes.USAGE_NOTIFICATION)
                .setContentType(AudioAttributes.CONTENT_TYPE_SONIFICATION)
                .build()
        )
    }
    manager.createNotificationChannel(channel)
}
```

---

## Notification Vibration

| Channel       | Vibration Pattern           | Description              |
|---------------|-----------------------------|--------------------------|
| AI Alerts     | `[0, 200]`                  | Single short buzz        |
| Security      | `[0, 100, 100, 100]`        | Triple urgent buzz       |
| Workflow      | `[0, 100]`                  | Single gentle buzz       |
| Documents     | None                        | Silent                   |
| General       | None                        | System default           |

```kotlin
// Custom vibration pattern
if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
    channel.vibrationPattern = longArrayOf(0, 100, 50, 100)
    // Format: delay, vibrate, sleep, vibrate...
}

// Per-notification vibration override (pre-Oreo)
builder.setVibrate(pattern)
```

---

## Notification LED

```kotlin
// Channel configuration (Android 8.0+)
channel.enableLights(true)
channel.lightColor = when (type) {
    NotificationType.AI_ALERT -> Color.BLUE
    NotificationType.SECURITY -> Color.RED
    NotificationType.WORKFLOW -> Color.GREEN
    else -> Color.WHITE
}

// Pre-Oreo
builder.setLights(color, 500, 500) // color, onMs, offMs
```

---

## Notification Large Icon

```kotlin
// App icon
builder.setLargeIcon(
    BitmapFactory.decodeResource(context.resources, R.drawable.ic_launcher_foreground)
)

// User avatar for personalized notifications
fun loadUserAvatar(url: String): Bitmap? {
    return try {
        val connection = URL(url).openConnection()
        BitmapFactory.decodeStream(connection.getInputStream())
    } catch (e: Exception) {
        BitmapFactory.decodeResource(context.resources, R.drawable.ic_default_avatar)
    }
}

// AI model icon
builder.setLargeIcon(
    BitmapFactory.decodeResource(context.resources, R.drawable.ic_ai_model)
)
```

---

## Notification Styles

### Big Text Style

```kotlin
builder.setStyle(
    NotificationCompat.BigTextStyle()
        .bigText(notification.body)
        .setBigContentTitle(notification.title)
        .setSummaryText("Nexus AI")
        .setSummaryText("Tap to view full response")
)
```

### Inbox Style

```kotlin
builder.setStyle(
    NotificationCompat.InboxStyle()
        .setBigContentTitle("3 new notifications")
        .setSummaryText("Nexus AI")
        .addLine("Agent completed: Data Analysis")
        .addLine("Document processed: report.pdf")
        .addLine("Security: New login detected")
)
```

### Media Style

```kotlin
builder.setStyle(
    NotificationCompat.MediaStyle()
        .setMediaSession(mediaSession.sessionToken)
        .setShowActionsInCompactView(0, 1)
        .setShowCancelButton(true)
        .setCancelButtonAction(0)
).addAction(R.drawable.ic_prev, "Previous", prevPending)
 .addAction(R.drawable.ic_pause, "Pause", pausePending)
 .addAction(R.drawable.ic_next, "Next", nextPending)
```

### Custom Layout

```kotlin
// layout/notification_ai_response.xml
// <RelativeLayout>
//   <ImageView android:id="@+id/icon" />
//   <TextView android:id="@+id/title" />
//   <TextView android:id="@+id/modelName" />
//   <TextView android:id="@+id/body" />
//   <ProgressBar android:id="@+id/progress" />
// </RelativeLayout>

val customView = RemoteViews(context.packageName, R.layout.notification_ai_response).apply {
    setTextViewText(R.id.title, notification.title)
    setTextViewText(R.id.body, notification.body.take(100) + "...")
    setTextViewText(R.id.modelName, notification.metadata["modelName"] ?: "AI")
    setImageViewResource(R.id.icon, R.drawable.ic_ai_notification)
}

val customBigView = RemoteViews(context.packageName, R.layout.notification_ai_response_expanded).apply {
    setTextViewText(R.id.title, notification.title)
    setTextViewText(R.id.body, notification.body)
    setTextViewText(R.id.modelName, notification.metadata["modelName"] ?: "AI")
    setTextViewText(R.id.timestamp, notification.receivedAt.toFormattedDateTime())
}

builder.setCustomContentView(customView)
    .setCustomBigContentView(customBigView)
    .setCustomSmallIcon(R.drawable.ic_ai_notification)
```

---

## Notification Click Handling

### Deep Link Resolution

```kotlin
class DeepLinkResolver @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun createPendingIntent(deepLink: String): PendingIntent {
        val intent = createDeepLinkIntent(deepLink).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        return PendingIntent.getActivity(
            context,
            deepLink.hashCode(),
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }

    private fun createDeepLinkIntent(deepLink: String): Intent {
        return when {
            deepLink.startsWith("/chat/") -> {
                val chatId = deepLink.removePrefix("/chat/")
                Intent(context, MainActivity::class.java).apply {
                    putExtra("destination", "chat")
                    putExtra("chatId", chatId)
                }
            }
            deepLink.startsWith("/agent/") -> {
                val agentId = deepLink.removePrefix("/agent/")
                Intent(context, MainActivity::class.java).apply {
                    putExtra("destination", "agent_detail")
                    putExtra("agentId", agentId)
                }
            }
            deepLink.startsWith("/document/") -> {
                val docId = deepLink.removePrefix("/document/")
                Intent(context, MainActivity::class.java).apply {
                    putExtra("destination", "document_detail")
                    putExtra("documentId", docId)
                }
            }
            deepLink.startsWith("/alerts/") -> {
                Intent(context, MainActivity::class.java).apply {
                    putExtra("destination", "notifications")
                }
            }
            else -> {
                Intent(context, MainActivity::class.java).apply {
                    putExtra("destination", "dashboard")
                }
            }
        }
    }
}
```

### Intent Handling in MainActivity

```kotlin
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        handleNotificationIntent(intent)
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        handleNotificationIntent(intent)
    }

    private fun handleNotificationIntent(intent: Intent?) {
        val destination = intent?.getStringExtra("destination") ?: return
        val chatId = intent.getStringExtra("chatId")
        val agentId = intent.getStringExtra("agentId")
        val documentId = intent.getStringExtra("documentId")

        // Navigate using Navigation Component
        when (destination) {
            "chat" -> navController.navigate("chat/$chatId")
            "agent_detail" -> navController.navigate("agent/$agentId")
            "document_detail" -> navController.navigate("document/$documentId")
            "notifications" -> navController.navigate("notifications")
            "dashboard" -> navController.navigate("dashboard")
        }
    }
}
```

---

## Notification Dismiss Handling

```kotlin
class NotificationDismissReceiver : BroadcastReceiver() {
    @Inject lateinit var notificationRepository: NotificationRepository

    override fun onReceive(context: Context, intent: Intent) {
        val notificationId = intent.getStringExtra("notification_id") ?: return

        goAsync().let { pendingResult ->
            CoroutineScope(Dispatchers.IO).launch {
                try {
                    notificationRepository.markDismissed(notificationId)
                    // Track dismiss analytics
                } finally {
                    pendingResult.finish()
                }
            }
        }
    }
}
```

---

## Notification Badge

### App Icon Badge

```kotlin
class BadgeManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun updateBadgeCount(count: Int) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val badgeManager = ShortcutBadger.applyCountOrThrow(context, count)
        }

        // Alternative for API 26+
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val launcherApps = context.getSystemService(LauncherApps::class.java)
            val componentName = ComponentName(context, MainActivity::class.java)
            launcherApps?.shortcutHostManager?.apply {
                // Badge is managed through notification channels
            }
        }
    }

    fun clearBadge() {
        ShortcutBadger.removeCountOrThrow(context)
    }
}

// In NotificationRepository
fun incrementUnreadCount() {
    viewModelScope.launch {
        val current = _unreadCount.value
        _unreadCount.value = current + 1
        badgeManager.updateBadgeCount(_unreadCount.value)
    }
}

fun markAllRead() {
    viewModelScope.launch {
        _unreadCount.value = 0
        badgeManager.clearBadge()
    }
}
```

---

## In-App Notification Center

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationCenterScreen(
    viewModel: NotificationCenterViewModel = hiltViewModel(),
    onNotificationClick: (AppNotification) -> Unit,
    onBack: () -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val isRefreshing by viewModel.isRefreshing.collectAsStateWithLifecycle()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Notifications") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back")
                    }
                },
                actions = {
                    if (state.notifications.any { !it.isRead }) {
                        TextButton(onClick = { viewModel.onAction(NotificationAction.MarkAllRead) }) {
                            Text("Mark all read")
                        }
                    }
                    IconButton(onClick = { viewModel.onAction(NotificationAction.ClearAll) }) {
                        Icon(Icons.Default.DeleteSweep, "Clear all")
                    }
                }
            )
        }
    ) { padding ->
        SwipeRefresh(
            state = rememberSwipeRefreshState(isRefreshing),
            onRefresh = { viewModel.onAction(NotificationAction.Refresh) },
            modifier = Modifier.padding(padding)
        ) {
            when {
                state.isLoading -> NotificationListSkeleton()
                state.notifications.isEmpty() -> NotificationEmpty()
                else -> {
                    val grouped = state.notifications.groupBy { notification ->
                        when {
                            notification.receivedAt.isAfter(Instant.now().minus(Duration.ofHours(24))) -> "Today"
                            notification.receivedAt.isAfter(Instant.now().minus(Duration.ofDays(7))) -> "This Week"
                            else -> "Earlier"
                        }
                    }

                    LazyColumn(
                        contentPadding = PaddingValues(16.dp),
                        verticalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        grouped.forEach { (section, notifications) ->
                            item(key = "header_$section") {
                                Text(
                                    text = section,
                                    style = MaterialTheme.typography.titleSmall,
                                    fontWeight = FontWeight.SemiBold,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    modifier = Modifier.padding(vertical = 8.dp)
                                )
                            }

                            items(notifications, key = { it.id }) { notification ->
                                NotificationItem(
                                    notification = notification,
                                    onClick = {
                                        viewModel.onAction(NotificationAction.MarkRead(notification.id))
                                        onNotificationClick(notification)
                                    },
                                    onDismiss = {
                                        viewModel.onAction(NotificationAction.Dismiss(notification.id))
                                    }
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun NotificationItem(
    notification: AppNotification,
    onClick: () -> Unit,
    onDismiss: () -> Unit
) {
    val dismissState = rememberDismissState(
        confirmValueChange = { value ->
            if (value == DismissValue.DismissedToStart) {
                onDismiss()
                true
            } else false
        }
    )

    SwipeToDismiss(
        state = dismissState,
        background = {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(MaterialTheme.colorScheme.errorContainer)
                    .padding(horizontal = 20.dp),
                contentAlignment = Alignment.CenterEnd
            ) {
                Icon(
                    Icons.Default.Delete,
                    contentDescription = "Dismiss",
                    tint = MaterialTheme.colorScheme.onErrorContainer
                )
            }
        },
        dismissContent = {
            NotificationItemContent(notification = notification, onClick = onClick)
        },
        directions = setOf(DismissDirection.EndToStart)
    )
}

@Composable
fun NotificationItemContent(
    notification: AppNotification,
    onClick: () -> Unit
) {
    val (icon, color) = when (notification.type) {
        NotificationType.AI_ALERT -> Icons.Default.SmartToy to MaterialTheme.colorScheme.primary
        NotificationType.SECURITY -> Icons.Default.Security to MaterialTheme.colorScheme.error
        NotificationType.WORKFLOW -> Icons.Default.AccountTree to MaterialTheme.colorScheme.tertiary
        NotificationType.DOCUMENT -> Icons.Default.Description to MaterialTheme.colorScheme.secondary
        NotificationType.GENERAL -> Icons.Default.Notifications to MaterialTheme.colorScheme.onSurface
        NotificationType.SYSTEM -> Icons.Default.Settings to MaterialTheme.colorScheme.onSurfaceVariant
    }

    Card(
        onClick = onClick,
        colors = CardDefaults.cardColors(
            containerColor = if (notification.isRead)
                MaterialTheme.colorScheme.surface
            else MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.15f)
        ),
        modifier = Modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.Top
        ) {
            // Icon
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .clip(CircleShape)
                    .background(color.copy(alpha = 0.12f)),
                contentAlignment = Alignment.Center
            ) {
                Icon(icon, contentDescription = null, tint = color, modifier = Modifier.size(20.dp))
            }

            Spacer(Modifier.width(12.dp))

            Column(modifier = Modifier.weight(1f)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        text = notification.title,
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = if (notification.isRead) FontWeight.Normal else FontWeight.Bold,
                        modifier = Modifier.weight(1f)
                    )
                    if (!notification.isRead) {
                        Box(
                            modifier = Modifier
                                .size(8.dp)
                                .clip(CircleShape)
                                .background(MaterialTheme.colorScheme.primary)
                        )
                    }
                }

                Spacer(Modifier.height(4.dp))

                Text(
                    text = notification.body,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis
                )

                Spacer(Modifier.height(4.dp))

                Text(
                    text = notification.receivedAt.toRelativeTimeString(),
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f)
                )
            }
        }
    }
}
```

---

## Notification Preferences

```kotlin
@Composable
fun NotificationPreferencesScreen(
    viewModel: NotificationPreferencesViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // Master toggle
        item {
            Card {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Push Notifications", fontWeight = FontWeight.SemiBold)
                        Text("Enable all push notifications", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                    Switch(
                        checked = state.masterEnabled,
                        onCheckedChange = { viewModel.onAction(PrefAction.ToggleMaster(it)) }
                    )
                }
            }
        }

        // Per-type toggles
        item { Text("Notification Types", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold) }

        NotificationType.entries.forEach { type ->
            item(key = type.name) {
                NotificationTypePreference(
                    type = type,
                    enabled = state.typeEnabled[type] ?: true,
                    onToggle = { viewModel.onAction(PrefAction.ToggleType(type, it)) }
                )
            }
        }

        // Sound preferences
        item { Text("Sound & Vibration", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold) }

        item {
            NotificationTypePreference(
                type = NotificationType.AI_ALERT,
                enabled = state.soundEnabled,
                onToggle = { viewModel.onAction(PrefAction.ToggleSound(it)) }
            )
        }

        item {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("Vibration", style = MaterialTheme.typography.bodyMedium)
                Switch(
                    checked = state.vibrationEnabled,
                    onCheckedChange = { viewModel.onAction(PrefAction.ToggleVibration(it)) }
                )
            }
        }

        // Quiet hours
        item { Text("Quiet Hours", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold) }

        item {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text("Enable Quiet Hours")
                    Text(
                        "No notifications during quiet hours",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Switch(
                    checked = state.quietHoursEnabled,
                    onCheckedChange = { viewModel.onAction(PrefAction.ToggleQuietHours(it)) }
                )
            }
        }
    }
}

@Composable
fun NotificationTypePreference(
    type: NotificationType,
    enabled: Boolean,
    onToggle: (Boolean) -> Unit
) {
    val (icon, label) = when (type) {
        NotificationType.AI_ALERT -> Icons.Default.SmartToy to "AI Alerts"
        NotificationType.SECURITY -> Icons.Default.Security to "Security Alerts"
        NotificationType.WORKFLOW -> Icons.Default.AccountTree to "Workflow Updates"
        NotificationType.DOCUMENT -> Icons.Default.Description to "Document Processing"
        NotificationType.GENERAL -> Icons.Default.Notifications to "General"
        NotificationType.SYSTEM -> Icons.Default.Settings to "System"
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(icon, contentDescription = null, modifier = Modifier.size(20.dp), tint = MaterialTheme.colorScheme.onSurfaceVariant)
        Spacer(Modifier.width(12.dp))
        Text(label, modifier = Modifier.weight(1f), style = MaterialTheme.typography.bodyMedium)
        Switch(checked = enabled, onCheckedChange = onToggle)
    }
}
```

---

## Notification Data Models

```kotlin
data class AppNotification(
    val id: String,
    val type: NotificationType,
    val title: String,
    val body: String,
    val deepLink: String,
    val priority: String = "default",
    val metadata: Map<String, String> = emptyMap(),
    val receivedAt: Instant = Instant.now(),
    val isRead: Boolean = false
)

enum class NotificationType(val channelName: String, val groupKey: String) {
    AI_ALERT("AI Alerts", "group_ai"),
    SECURITY("Security", "group_security"),
    WORKFLOW("Workflow Updates", "group_workflow"),
    DOCUMENT("Document Processing", "group_documents"),
    GENERAL("General", "group_general"),
    SYSTEM("System", "group_system");

    companion object {
        fun fromString(value: String): NotificationType =
            entries.find { it.name.equals(value, ignoreCase = true) } ?: GENERAL
    }
}

data class NotificationPreferences(
    val masterEnabled: Boolean = true,
    val typeEnabled: Map<NotificationType, Boolean> = NotificationType.entries.associateWith { true },
    val soundEnabled: Boolean = true,
    val vibrationEnabled: Boolean = true,
    val quietHoursEnabled: Boolean = false,
    val quietHoursStart: LocalTime = LocalTime.of(22, 0),
    val quietHoursEnd: LocalTime = LocalTime.of(7, 0)
)

data class NotificationCenterState(
    val isLoading: Boolean = false,
    val notifications: List<AppNotification> = emptyList(),
    val unreadCount: Int = 0,
    val error: String? = null
)
```

---

## Notification Caching

```kotlin
@Entity(tableName = "notifications")
data class NotificationEntity(
    @PrimaryKey val id: String,
    val type: String,
    val title: String,
    val body: String,
    val deepLink: String,
    val priority: String,
    val metadataJson: String,
    val receivedAt: Long,
    val isRead: Boolean,
    val isDismissed: Boolean = false
)

@Dao
interface NotificationDao {

    @Query("SELECT * FROM notifications WHERE isDismissed = 0 ORDER BY receivedAt DESC")
    fun getAllNotifications(): Flow<List<NotificationEntity>>

    @Query("SELECT COUNT(*) FROM notifications WHERE isRead = 0 AND isDismissed = 0")
    fun getUnreadCount(): Flow<Int>

    @Query("SELECT * FROM notifications WHERE id = :id")
    suspend fun getById(id: String): NotificationEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsert(notification: NotificationEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsertAll(notifications: List<NotificationEntity>)

    @Query("UPDATE notifications SET isRead = 1 WHERE id = :id")
    suspend fun markRead(id: String)

    @Query("UPDATE notifications SET isRead = 1")
    suspend fun markAllRead()

    @Query("UPDATE notifications SET isDismissed = 1 WHERE id = :id")
    suspend fun markDismissed(id: String)

    @Query("DELETE FROM notifications WHERE receivedAt < :before")
    suspend fun deleteOlderThan(before: Long)

    @Query("DELETE FROM notifications")
    suspend fun deleteAll()

    @Query("SELECT * FROM notifications WHERE isDismissed = 0 ORDER BY receivedAt DESC LIMIT :limit OFFSET :offset")
    suspend fun getPage(limit: Int, offset: Int): List<NotificationEntity>
}
```

---

## Notification API Integration

```kotlin
interface NotificationApiService {

    @POST("api/v1/notifications/fcm-token")
    suspend fun registerFcmToken(@Body request: FcmTokenRequest): Response<Unit>

    @HTTP(method = "DELETE", path = "api/v1/notifications/fcm-token")
    suspend fun unregisterFcmToken(@Body request: FcmTokenRequest): Response<Unit>

    @POST("api/v1/notifications")
    suspend fun getNotifications(
        @Body request: GetNotificationsRequest
    ): PaginatedResponse<NotificationDto>

    @PATCH("api/v1/notifications/{id}/read")
    suspend fun markRead(@Path("id") notificationId: String): Response<Unit>

    @PATCH("api/v1/notifications/read-all")
    suspend fun markAllRead(): Response<Unit>

    @DELETE("api/v1/notifications/{id}")
    suspend fun deleteNotification(@Path("id") notificationId: String): Response<Unit>

    @POST("api/v1/notifications/preferences")
    suspend fun getPreferences(@Body request: GetPreferencesRequest): NotificationPreferencesDto

    @PATCH("api/v1/notifications/preferences")
    suspend fun updatePreferences(@Body request: UpdatePreferencesRequest): NotificationPreferencesDto

    @POST("api/v1/notifications/{id}/action")
    suspend fun handleAction(
        @Path("id") notificationId: String,
        @Body request: NotificationActionRequest
    ): Response<Unit>
}

@Serializable
data class NotificationDto(
    val id: String,
    val type: String,
    val title: String,
    val body: String,
    val deepLink: String,
    val priority: String,
    val metadata: Map<String, String>,
    val receivedAt: String,
    val isRead: Boolean
)

@Serializable
data class NotificationPreferencesDto(
    val masterEnabled: Boolean,
    val typeEnabled: Map<String, Boolean>,
    val soundEnabled: Boolean,
    val vibrationEnabled: Boolean,
    val quietHoursEnabled: Boolean,
    val quietHoursStart: String,
    val quietHoursEnd: String
)

@Serializable
data class UpdatePreferencesRequest(
    val masterEnabled: Boolean? = null,
    val typeEnabled: Map<String, Boolean>? = null,
    val soundEnabled: Boolean? = null,
    val vibrationEnabled: Boolean? = null,
    val quietHoursEnabled: Boolean? = null,
    val quietHoursStart: String? = null,
    val quietHoursEnd: String? = null
)

@Serializable
data class NotificationActionRequest(
    val action: String,
    val payload: Map<String, String>? = null
)
```

---

## Notification Permissions

### Android 13+ (API 33) Runtime Permission

```kotlin
@Composable
fun NotificationPermissionHandler(
    onGranted: @Composable () -> Unit,
    onDenied: @Composable () -> Unit
) {
    val context = LocalContext.current
    var hasPermission by remember {
        mutableStateOf(
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                ContextCompat.checkSelfPermission(
                    context, Manifest.permission.POST_NOTIFICATIONS
                ) == PackageManager.PERMISSION_GRANTED
            } else {
                true // Pre-13: Permission granted at install
            }
        )
    }

    val launcher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { granted -> hasPermission = granted }

    LaunchedEffect(Unit) {
        if (!hasPermission && Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            launcher.launch(Manifest.permission.POST_NOTIFICATIONS)
        }
    }

    if (hasPermission) onGranted() else onDenied()
}

@Composable
fun NotificationPermissionRationale(
    onRequestPermission: () -> Unit,
    onSkip: () -> Unit
) {
    AlertDialog(
        onDismissRequest = onSkip,
        icon = { Icon(Icons.Default.NotificationsActive, contentDescription = null) },
        title = { Text("Enable Notifications") },
        text = {
            Text("Allow Nexus AI to send you notifications about AI responses, security alerts, and workflow updates.")
        },
        confirmButton = {
            Button(onClick = onRequestPermission) { Text("Enable") }
        },
        dismissButton = {
            TextButton(onClick = onSkip) { Text("Not now") }
        }
    )
}
```

---

## Notification on Different Android Versions

| Version | API | Behavior                                               |
|---------|-----|--------------------------------------------------------|
| 8.0+    | 26  | Notification channels required                         |
| 8.1+    | 27  | Background notification limits                         |
| 9.0     | 28  | Notification bubbles (optional)                        |
| 10      | 29  | Notification history, bubble improvements              |
| 11      | 30  | Notification permission auto-reset                     |
| 12      | 31  | Custom notification trampolines restricted             |
| 13      | 33  | POST_NOTIFICATIONS runtime permission required         |
| 14      | 34  | Foreground service notification restrictions           |

```kotlin
fun handleVersionSpecificBehavior() {
    when {
        Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU -> {
            // Android 13+: Check POST_NOTIFICATIONS permission
            if (ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS)
                != PackageManager.PERMISSION_GRANTED) {
                // Request permission
            }
        }
        Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            // Android 12+: Notification trampoline restrictions
            // Use PendingIntent directly, not through broadcast
        }
        Build.VERSION.SDK_INT >= Build.VERSION_CODES.O -> {
            // Android 8+: Channels are required
            channelManager.createAllChannels()
        }
    }
}
```

---

## Notification Testing

### Unit Tests

```kotlin
class NotificationChannelManagerTest {

    private lateinit var channelManager: NotificationChannelManager

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        channelManager = NotificationChannelManager(context)
    }

    @Test
    fun `createAllChannels creates all required channels`() {
        channelManager.createAllChannels()

        val manager = ApplicationProvider.getApplicationContext<Context>()
            .getSystemService(NotificationManager::class.java)

        val channels = manager.notificationChannels
        assertEquals(6, channels.size)
        assertTrue(channels.any { it.id == "ai_alerts" })
        assertTrue(channels.any { it.id == "security" })
        assertTrue(channels.any { it.id == "workflow" })
        assertTrue(channels.any { it.id == "documents" })
        assertTrue(channels.any { it.id == "general" })
        assertTrue(channels.any { it.id == "system" })
    }

    @Test
    fun `getChannel returns correct channel for AI alert`() {
        assertEquals("ai_alerts", channelManager.getChannel(NotificationType.AI_ALERT))
    }

    @Test
    fun `getChannel returns correct channel for security`() {
        assertEquals("security", channelManager.getChannel(NotificationType.SECURITY))
    }
}

class DeepLinkResolverTest {

    private lateinit var resolver: DeepLinkResolver

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        resolver = DeepLinkResolver(context)
    }

    @Test
    fun `chat deep link creates correct intent`() {
        val intent = resolver.createDeepLinkIntent("/chat/123")
        assertEquals("chat", intent.getStringExtra("destination"))
        assertEquals("123", intent.getStringExtra("chatId"))
    }

    @Test
    fun `agent deep link creates correct intent`() {
        val intent = resolver.createDeepLinkIntent("/agent/abc")
        assertEquals("agent_detail", intent.getStringExtra("destination"))
        assertEquals("abc", intent.getStringExtra("agentId"))
    }
}
```

### UI Tests

```kotlin
@Test
fun notificationCenter_showsEmptyState() {
    composeTestRule.setContent {
        NotificationEmpty()
    }
    composeTestRule.onNodeWithText("No notifications").assertIsDisplayed()
}

@Test
fun notificationItem_displaysContent() {
    val notification = AppNotification(
        id = "1",
        type = NotificationType.AI_ALERT,
        title = "AI Response Ready",
        body = "Your query has been processed",
        deepLink = "/chat/123"
    )

    composeTestRule.setContent {
        NotificationItemContent(notification = notification, onClick = {})
    }

    composeTestRule.onNodeWithText("AI Response Ready").assertIsDisplayed()
    composeTestRule.onNodeWithText("Your query has been processed").assertIsDisplayed()
}

@Test
fun notificationPreferences_togglesAreClickable() {
    composeTestRule.setContent {
        NotificationTypePreference(
            type = NotificationType.AI_ALERT,
            enabled = true,
            onToggle = {}
        )
    }
    composeTestRule.onNodeWithText("AI Alerts").assertIsDisplayed()
}
```

---

## Notification Accessibility

```kotlin
// Notification content descriptions
NotificationItem(
    notification = notification,
    onClick = { ... },
    modifier = Modifier.semantics(mergeDescendants = true) {
        contentDescription = buildString {
            append("${notification.type.name} notification. ")
            append("${notification.title}. ")
            append("${notification.body}. ")
            append(notification.receivedAt.toRelativeTimeString())
            if (!notification.isRead) append(". Unread")
        }
    }
)

// Unread indicator
Box(
    modifier = Modifier.semantics {
        stateDescription = if (!notification.isRead) "Unread" else "Read"
    }
)

// Action buttons
IconButton(
    onClick = { ... },
    modifier = Modifier.semantics { role = Role.Button }
) {
    Icon(
        Icons.Default.Delete,
        contentDescription = "Dismiss notification"
    )
}

// Preference toggle
Switch(
    checked = enabled,
    onCheckedChange = onToggle,
    modifier = Modifier.semantics {
        stateDescription = "$label notifications: ${if (enabled) "enabled" else "disabled"}"
    }
)

// Live region for new notifications
LazyColumn(modifier = Modifier.semantics {
    liveRegion = LiveRegionMode.Polite
})
```

---

## Summary

| Component               | File / Class                    | Notes                          |
|-------------------------|----------------------------------|--------------------------------|
| FCM Service             | `FirebaseNotificationService`    | Handles push from FCM          |
| Token Manager           | `FcmTokenManager`                | Register/refresh FCM token     |
| Channel Manager         | `NotificationChannelManager`     | Create 6 notification channels |
| Notification Builder    | `AppNotificationBuilder`         | Build and show notifications   |
| Deep Link Resolver      | `DeepLinkResolver`               | Route clicks to screens        |
| Badge Manager           | `BadgeManager`                   | App icon badge count           |
| Notification Center     | `NotificationCenterScreen`       | In-app notification list       |
| Preferences             | `NotificationPreferencesScreen`  | Per-type toggles, quiet hours  |
| Action Receiver         | `NotificationActionReceiver`     | Handle approve/reject actions  |
| Dismiss Receiver        | `NotificationDismissReceiver`    | Track dismissals               |
| Permission Handler      | `NotificationPermissionHandler`  | POST_NOTIFICATIONS (API 33)    |
| Cache (Room)            | `NotificationDao`                | Local persistence              |
| API Service             | `NotificationApiService`         | CRUD + preferences + actions   |
| Data Models             | `AppNotification`, enums         | Domain models                  |
