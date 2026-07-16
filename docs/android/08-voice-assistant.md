# 08 - Voice Assistant

The voice assistant enables hands-free interaction with Nexus AI through Android's
speech recognition APIs. Users can speak queries, commands, and dictate text across
the application.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Voice Input Architecture](#voice-input-architecture)
3. [Voice Button Composable](#voice-button-composable)
4. [Voice Recording Flow](#voice-recording-flow)
5. [Speech-to-Text Conversion](#speech-to-text-conversion)
6. [Voice Input in Chat](#voice-input-in-chat)
7. [Voice Input States](#voice-input-states)
8. [Voice Animation](#voice-animation)
9. [Voice Input Error Handling](#voice-input-error-handling)
10. [Voice Input Accessibility](#voice-input-accessibility)
11. [Voice Commands](#voice-commands)
12. [Voice Input Settings](#voice-input-settings)
13. [Voice Input in Other Screens](#voice-input-in-other-screens)
14. [Voice Input Offline Support](#voice-input-offline-support)
15. [Voice Input History](#voice-input-history)
16. [Voice Input API Integration](#voice-input-api-integration)
17. [Voice Input Data Models](#voice-input-data-models)
18. [Voice Input Testing](#voice-input-testing)
19. [Voice Input Performance](#voice-input-performance)
20. [Voice Input UX Best Practices](#voice-input-ux-best-practices)
21. [Voice Input Permissions](#voice-input-permissions)
22. [Voice Input on Different Devices](#voice-input-on-different-devices)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                         UI Layer                                 │
│  VoiceButton ──► PromptBox (chat) / SearchBar / CommandBar       │
│       │                                                          │
│       ▼                                                          │
│  VoiceInputManager (Compose wrapper)                             │
│       │                                                          │
│       ├── VoiceRecognitionManager (Android SpeechRecognizer)     │
│       ├── VoiceSettingsManager (DataStore preferences)           │
│       ├── VoiceHistoryRepository (recent inputs)                 │
│       └── VoiceCommandProcessor (predefined commands)            │
│                                                                  │
│  Android Platform                                                │
│  ├── SpeechRecognizer API                                        │
│  ├── AudioRecord / MediaRecorder                                 │
│  └── LanguageRecognizer / Intent                                 │
│                                                                  │
│  Optional Cloud                                                  │
│  ├── Google Cloud Speech-to-Text                                 │
│  └── Whisper API                                                 │
└──────────────────────────────────────────────────────────────────┘
```

### Processing Flow

```
Press ──► Permission Check ──► Start Recording ──► Audio Capture
                                                        │
                                          ┌─────────────┘
                                          ▼
                                   SpeechRecognizer
                                          │
                                    ┌─────┴─────┐
                                    ▼           ▼
                              Partial Result  Final Result
                                    │           │
                                    ▼           ▼
                              UI Update     Process & Send
                                          │
                                    ┌─────┴─────┐
                                    ▼           ▼
                              Voice Command  Chat Input
                              (if matched)   (otherwise)
```

---

## Voice Input Architecture

### Permission Flow

```kotlin
// Manifest declaration
// <uses-permission android:name="android.permission.RECORD_AUDIO" />

@Composable
fun VoicePermissionGate(
    onPermissionGranted: @Composable () -> Unit,
    onPermissionDenied: @Composable () -> Unit
) {
    val context = LocalContext.current
    var hasPermission by remember {
        mutableStateOf(
            ContextCompat.checkSelfPermission(
                context, Manifest.permission.RECORD_AUDIO
            ) == PackageManager.PERMISSION_GRANTED
        )
    }

    val launcher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { granted -> hasPermission = granted }

    if (hasPermission) {
        onPermissionGranted()
    } else {
        onPermissionDenied()
    }
}
```

---

## Voice Button Composable

Three visual states: idle, listening, processing.

```kotlin
@Composable
fun VoiceButton(
    state: VoiceInputState,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    size: Dp = 56.dp
) {
    val animatedScale by animateFloatAsState(
        targetValue = when (state) {
            VoiceInputState.LISTENING -> 1.2f
            VoiceInputState.PROCESSING -> 1.0f
            VoiceInputState.IDLE -> 1.0f
        },
        animationSpec = spring(
            dampingRatio = Spring.DampingRatioMediumBouncy,
            stiffness = Spring.StiffnessLow
        ),
        label = "voice_button_scale"
    )

    val backgroundColor = when (state) {
        VoiceInputState.IDLE -> MaterialTheme.colorScheme.surfaceVariant
        VoiceInputState.LISTENING -> MaterialTheme.colorScheme.error
        VoiceInputState.PROCESSING -> MaterialTheme.colorScheme.primary
    }

    val iconColor = when (state) {
        VoiceInputState.IDLE -> MaterialTheme.colorScheme.onSurfaceVariant
        VoiceInputState.LISTENING -> MaterialTheme.colorScheme.onError
        VoiceInputState.PROCESSING -> MaterialTheme.colorScheme.onPrimary
    }

    Box(
        modifier = modifier
            .size(size)
            .scale(animatedScale)
            .clip(CircleShape)
            .background(backgroundColor)
            .clickable(
                onClick = onClick,
                indication = null,
                interactionSource = remember { MutableInteractionSource() }
            ),
        contentAlignment = Alignment.Center
    ) {
        // Pulse animation when listening
        if (state == VoiceInputState.LISTENING) {
            PulseAnimation(
                color = MaterialTheme.colorScheme.error,
                modifier = Modifier.size(size * 1.4f)
            )
        }

        Icon(
            imageVector = when (state) {
                VoiceInputState.IDLE -> Icons.Default.Mic
                VoiceInputState.LISTENING -> Icons.Default.Mic
                VoiceInputState.PROCESSING -> Icons.Default.Sync
            },
            contentDescription = when (state) {
                VoiceInputState.IDLE -> "Start voice input"
                VoiceInputState.LISTENING -> "Listening... Tap to stop"
                VoiceInputState.PROCESSING -> "Processing speech"
            },
            tint = iconColor,
            modifier = Modifier.size(size * 0.45f)
        )
    }
}

enum class VoiceInputState {
    IDLE, LISTENING, PROCESSING
}
```

### Voice Button with Label

```kotlin
@Composable
fun VoiceButtonWithLabel(
    state: VoiceInputState,
    partialText: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = modifier
    ) {
        VoiceButton(state = state, onClick = onClick)

        Spacer(Modifier.height(8.dp))

        AnimatedVisibility(visible = state == VoiceInputState.LISTENING) {
            Text(
                text = if (partialText.isNotEmpty()) partialText else "Listening...",
                style = MaterialTheme.typography.bodySmall,
                color = if (partialText.isNotEmpty())
                    MaterialTheme.colorScheme.onSurface
                else MaterialTheme.colorScheme.onSurfaceVariant,
                textAlign = TextAlign.Center,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.widthIn(max = 200.dp)
            )
        }

        if (state == VoiceInputState.IDLE) {
            Text(
                text = "Tap to speak",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}
```

---

## Voice Recording Flow

```kotlin
class VoiceRecognitionManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val settingsManager: VoiceSettingsManager
) {
    private var speechRecognizer: SpeechRecognizer? = null
    private val _state = MutableStateFlow(VoiceInputState.IDLE)
    val state: StateFlow<VoiceInputState> = _state.asStateFlow()

    private val _partialResults = MutableStateFlow("")
    val partialResults: StateFlow<String> = _partialResults.asStateFlow()

    private val _finalResult = MutableSharedFlow<String>(extraBufferCapacity = 1)
    val finalResult: SharedFlow<String> = _finalResult.asSharedFlow()

    private val _error = MutableSharedFlow<VoiceError>(extraBufferCapacity = 1)
    val error: SharedFlow<VoiceError> = _error.asSharedFlow()

    fun isAvailable(): Boolean {
        return SpeechRecognizer.isRecognitionAvailable(context)
    }

    fun startListening() {
        if (!isAvailable()) {
            _error.tryEmit(VoiceError.RecognizerNotAvailable)
            return
        }

        speechRecognizer?.destroy()
        speechRecognizer = SpeechRecognizer.createSpeechRecognizer(context).apply {
            setRecognitionListener(createRecognitionListener())
        }

        val intent = Intent(RecognizerIntent.ACTION_RECOGNIZE_SPEECH).apply {
            putExtra(RecognizerIntent.EXTRA_LANGUAGE_MODEL, RecognizerIntent.LANGUAGE_MODEL_FREE_FORM)
            putExtra(RecognizerIntent.EXTRA_LANGUAGE, settingsManager.getLanguage())
            putExtra(RecognizerIntent.EXTRA_PARTIAL_RESULTS, true)
            putExtra(RecognizerIntent.EXTRA_MAX_RESULTS, 1)
            putExtra(RecognizerIntent.EXTRA_SPEECH_INPUT_MINIMUM_LENGTH_MILLIS, 500L)
            putExtra(RecognizerIntent.EXTRA_SPEECH_INPUT_COMPLETE_SILENCE_LENGTH_MILLIS, 2000L)
        }

        _state.value = VoiceInputState.LISTENING
        _partialResults.value = ""
        speechRecognizer?.startListening(intent)
    }

    fun stopListening() {
        speechRecognizer?.stopListening()
    }

    fun cancelListening() {
        _state.value = VoiceInputState.IDLE
        _partialResults.value = ""
        speechRecognizer?.cancel()
        speechRecognizer?.destroy()
        speechRecognizer = null
    }

    fun destroy() {
        speechRecognizer?.destroy()
        speechRecognizer = null
    }

    private fun createRecognitionListener() = object : RecognitionListener {
        override fun onReadyForSpeech(params: Bundle?) {
            _state.value = VoiceInputState.LISTENING
        }

        override fun onBeginningOfSpeech() {}

        override fun onRmsChanged(rmsdB: Float) {
            // Used for waveform visualization
            _rmsLevel.value = rmsdB
        }

        override fun onBufferReceived(buffer: ByteArray?) {}

        override fun onEndOfSpeech() {
            _state.value = VoiceInputState.PROCESSING
        }

        override fun onError(error: Int) {
            _state.value = VoiceInputState.IDLE
            val voiceError = when (error) {
                SpeechRecognizer.ERROR_NO_MATCH -> VoiceError.NoMatch
                SpeechRecognizer.ERROR_SPEECH_TIMEOUT -> VoiceError.Timeout
                SpeechRecognizer.ERROR_AUDIO -> VoiceError.AudioError
                SpeechRecognizer.ERROR_CLIENT -> VoiceError.ClientError
                SpeechRecognizer.ERROR_SERVER -> VoiceError.ServerError
                SpeechRecognizer.ERROR_INSUFFICIENT_PERMISSIONS -> VoiceError.PermissionDenied
                SpeechRecognizer.ERROR_NETWORK,
                SpeechRecognizer.ERROR_NETWORK_TIMEOUT -> VoiceError.NetworkError
                SpeechRecognizer.ERROR_RECOGNIZER_BUSY -> VoiceError.RecognizerBusy
                else -> VoiceError.Unknown(error)
            }
            _error.tryEmit(voiceError)
        }

        override fun onResults(results: Bundle?) {
            val matches = results?.getStringArrayList(SpeechRecognizer.RESULTS_RECOGNITION)
            val text = matches?.firstOrNull() ?: ""
            _finalResult.tryEmit(text)
            _state.value = VoiceInputState.IDLE
        }

        override fun onPartialResults(partialResults: Bundle?) {
            val matches = partialResults?.getStringArrayList(SpeechRecognizer.RESULTS_RECOGNITION)
            _partialResults.value = matches?.firstOrNull() ?: ""
        }

        override fun onEvent(eventType: Int, params: Bundle?) {}
    }
}
```

---

## Speech-to-Text Conversion

### VoiceInputManager (Compose Wrapper)

```kotlin
@Composable
fun rememberVoiceInputManager(
    onResult: (String) -> Unit,
    onError: (VoiceError) -> Unit
): VoiceInputManager {
    val context = LocalContext.current
    val recognitionManager = remember { VoiceRecognitionManager(context) }
    val commandProcessor = remember { VoiceCommandProcessor() }

    LaunchedEffect(Unit) {
        recognitionManager.finalResult.collect { text ->
            val command = commandProcessor.process(text)
            if (command != null) {
                command.execute()
            } else {
                onResult(text)
            }
        }
    }

    LaunchedEffect(Unit) {
        recognitionManager.error.collect { error ->
            onError(error)
        }
    }

    DisposableEffect(Unit) {
        onDispose { recognitionManager.destroy() }
    }

    return remember {
        VoiceInputManager(
            startListening = { recognitionManager.startListening() },
            stopListening = { recognitionManager.stopListening() },
            cancelListening = { recognitionManager.cancelListening() },
            isAvailable = { recognitionManager.isAvailable() },
            state = recognitionManager.state,
            partialResults = recognitionManager.partialResults
        )
    }
}

class VoiceInputManager(
    val startListening: () -> Unit,
    val stopListening: () -> Unit,
    val cancelListening: () -> Unit,
    val isAvailable: () -> Boolean,
    val state: StateFlow<VoiceInputState>,
    val partialResults: StateFlow<String>
)
```

---

## Voice Input in Chat

Voice button integrated into the PromptBox composable.

```kotlin
@Composable
fun PromptBox(
    value: String,
    onValueChange: (String) -> Unit,
    onSend: () -> Unit,
    onVoiceResult: (String) -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true
) {
    val voiceManager = rememberVoiceInputManager(
        onResult = { text ->
            onValueChange(value + text)
        },
        onError = { error ->
            // Show snackbar
        }
    )

    val voiceState by voiceManager.state.collectAsStateWithLifecycle()
    val partialText by voiceManager.partialResults.collectAsStateWithLifecycle()

    Column(modifier = modifier) {
        // Partial text preview
        if (voiceState == VoiceInputState.LISTENING && partialText.isNotEmpty()) {
            Surface(
                color = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp)
            ) {
                Text(
                    text = partialText,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.padding(8.dp),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
            }
        }

        // Input row
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 8.dp, vertical = 8.dp),
            verticalAlignment = Alignment.Bottom
        ) {
            // Text field
            OutlinedTextField(
                value = value,
                onValueChange = onValueChange,
                modifier = Modifier.weight(1f),
                placeholder = { Text("Type a message...") },
                maxLines = 4,
                enabled = enabled && voiceState != VoiceInputState.LISTENING
            )

            Spacer(Modifier.width(8.dp))

            // Voice button
            VoiceButton(
                state = voiceState,
                onClick = {
                    if (voiceState == VoiceInputState.LISTENING) {
                        voiceManager.stopListening()
                    } else {
                        voiceManager.startListening()
                    }
                },
                size = 48.dp
            )

            Spacer(Modifier.width(8.dp))

            // Send button
            IconButton(
                onClick = onSend,
                enabled = value.isNotBlank() && voiceState != VoiceInputState.LISTENING
            ) {
                Icon(
                    Icons.Default.Send,
                    contentDescription = "Send message",
                    tint = if (value.isNotBlank()) MaterialTheme.colorScheme.primary
                    else MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}
```

---

## Voice Input States

```kotlin
sealed interface VoiceInputUiState {
    data object Idle : VoiceInputUiState
    data object Listening : VoiceInputUiState
    data class PartialResult(val text: String, val rmsLevel: Float) : VoiceInputUiState
    data object Processing : VoiceInputUiState
    data class Result(val text: String) : VoiceInputUiState
    data class Error(val error: VoiceError) : VoiceInputUiState
}

sealed interface VoiceError {
    data object RecognizerNotAvailable : VoiceError
    data object NoMatch : VoiceError
    data object Timeout : VoiceError
    data object AudioError : VoiceError
    data object ClientError : VoiceError
    data object ServerError : VoiceError
    data object PermissionDenied : VoiceError
    data object NetworkError : VoiceError
    data object RecognizerBusy : VoiceError
    data class Unknown(val code: Int) : VoiceError

    val message: String
        get() = when (this) {
            is RecognizerNotAvailable -> "Speech recognition not available on this device"
            is NoMatch -> "No speech detected. Try again."
            is Timeout -> "Listening timed out"
            is AudioError -> "Audio recording error"
            is ClientError -> "Client error. Restart and try again."
            is ServerError -> "Server error. Try again later."
            is PermissionDenied -> "Microphone permission required"
            is NetworkError -> "Network error. Check your connection."
            is RecognizerBusy -> "Speech recognizer is busy. Try again."
            is Unknown -> "Unknown error (code: $code)"
        }
}
```

---

## Voice Animation

### Pulse Animation

```kotlin
@Composable
fun PulseAnimation(
    color: Color,
    modifier: Modifier = Modifier
) {
    val infiniteTransition = rememberInfiniteTransition(label = "pulse")

    val scale by infiniteTransition.animateFloat(
        initialValue = 0.8f,
        targetValue = 1.2f,
        animationSpec = infiniteRepeatable(
            animation = tween(800, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulse_scale"
    )

    val alpha by infiniteTransition.animateFloat(
        initialValue = 0.6f,
        targetValue = 0.0f,
        animationSpec = infiniteRepeatable(
            animation = tween(800, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulse_alpha"
    )

    Box(modifier = modifier) {
        // Outer pulse ring
        Box(
            modifier = Modifier
                .fillMaxSize()
                .scale(scale)
                .clip(CircleShape)
                .background(color.copy(alpha = alpha))
        )
    }
}
```

### Waveform Visualization

```kotlin
@Composable
fun WaveformVisualizer(
    rmsLevel: Float,
    isActive: Boolean,
    modifier: Modifier = Modifier,
    barCount: Int = 30,
    barColor: Color = MaterialTheme.colorScheme.primary
) {
    val infiniteTransition = rememberInfiniteTransition(label = "waveform")
    val animatedLevels = remember { mutableStateListOf<Float>() }

    LaunchedEffect(isActive) {
        if (isActive) {
            while (isActive) {
                animatedLevels.clear()
                repeat(barCount) {
                    animatedLevels.add(
                        (0.2f..0.8f).random().toFloat() *
                        ((rmsLevel + 10f) / 20f).coerceIn(0f, 1f)
                    )
                }
                delay(50)
            }
        }
    }

    Row(
        modifier = modifier
            .fillMaxWidth()
            .height(48.dp),
        horizontalArrangement = Arrangement.SpaceEvenly,
        verticalAlignment = Alignment.CenterVertically
    ) {
        repeat(barCount) { index ->
            val level = animatedLevels.getOrElse(index) { 0.1f }
            val animatedHeight by animateFloatAsState(
                targetValue = level,
                animationSpec = tween(100),
                label = "bar_$index"
            )

            Box(
                modifier = Modifier
                    .width(3.dp)
                    .height((animatedHeight * 40.dp).coerceAtLeast(2.dp))
                    .clip(RoundedCornerShape(1.5.dp))
                    .background(barColor.copy(alpha = 0.7f + level * 0.3f))
            )
        }
    }
}
```

### Recording Timer

```kotlin
@Composable
fun RecordingTimer(
    isRecording: Boolean,
    modifier: Modifier = Modifier
) {
    var elapsedSeconds by remember { mutableIntStateOf(0) }

    LaunchedEffect(isRecording) {
        if (isRecording) {
            while (isActive) {
                delay(1000)
                elapsedSeconds++
            }
        } else {
            elapsedSeconds = 0
        }
    }

    if (isRecording) {
        Row(
            modifier = modifier,
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Red recording dot
            Box(
                modifier = Modifier
                    .size(8.dp)
                    .clip(CircleShape)
                    .background(MaterialTheme.colorScheme.error)
            )
            Spacer(Modifier.width(6.dp))
            Text(
                text = formatDuration(elapsedSeconds),
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.error,
                fontWeight = FontWeight.Medium
            )
        }
    }
}

private fun formatDuration(seconds: Int): String {
    val min = seconds / 60
    val sec = seconds % 60
    return "%d:%02d".format(min, sec)
}
```

---

## Voice Input Error Handling

```kotlin
@Composable
fun VoiceErrorSnackbar(
    error: VoiceError?,
    onDismiss: () -> Unit,
    onRetry: (() -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(error) {
        if (error != null) {
            val result = snackbarHostState.showSnackbar(
                message = error.message,
                actionLabel = onRetry?.let { "Retry" },
                duration = SnackbarDuration.Short
            )
            when (result) {
                SnackbarResult.ActionPerformed -> onRetry?.invoke()
                SnackbarResult.Dismissed -> onDismiss()
            }
        }
    }
}

@Composable
fun VoicePermissionDeniedContent(
    onRequestPermission: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier.padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Icon(
            Icons.Default.MicOff,
            contentDescription = null,
            modifier = Modifier.size(48.dp),
            tint = MaterialTheme.colorScheme.error
        )
        Spacer(Modifier.height(12.dp))
        Text(
            "Microphone access needed",
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold
        )
        Spacer(Modifier.height(4.dp))
        Text(
            "Enable microphone permission in Settings to use voice input.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
        Spacer(Modifier.height(16.dp))
        Button(onClick = onRequestPermission) {
            Text("Grant Permission")
        }
        Spacer(Modifier.height(8.dp))
        TextButton(onClick = {
            val intent = Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS).apply {
                data = Uri.fromParts("package", context.packageName, null)
            }
            context.startActivity(intent)
        }) {
            Text("Open Settings")
        }
    }
}

@Composable
fun VoiceNotAvailableContent(modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Icon(
            Icons.Default.RecordVoiceOver,
            contentDescription = null,
            modifier = Modifier.size(48.dp),
            tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f)
        )
        Spacer(Modifier.height(12.dp))
        Text(
            "Voice input not available",
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold
        )
        Spacer(Modifier.height(4.dp))
        Text(
            "Speech recognition is not supported on this device.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
    }
}
```

---

## Voice Input Accessibility

```kotlin
// Voice button
VoiceButton(
    state = voiceState,
    onClick = { ... },
    modifier = Modifier.semantics {
        role = Role.Button
        stateDescription = when (voiceState) {
            VoiceInputState.IDLE -> "Voice input. Tap to start listening"
            VoiceInputState.LISTENING -> "Listening. Tap to stop"
            VoiceInputState.PROCESSING -> "Processing speech. Please wait"
        }
    }
)

// Waveform
WaveformVisualizer(
    rmsLevel = rmsLevel,
    isActive = isListening,
    modifier = Modifier.semantics {
        liveRegion = LiveRegionMode.Polite
        stateDescription = if (isListening) "Listening to voice input" else "Voice input idle"
    }
)

// Partial results
Text(
    text = partialText,
    modifier = Modifier.semantics {
        liveRegion = LiveRegionMode.Polite
    }
)

// Recording timer
RecordingTimer(
    isRecording = isListening,
    modifier = Modifier.semantics {
        stateDescription = "Recording duration: ${formatDuration(elapsedSeconds)}"
    }
)

// Error announcement
Snackbar(
    modifier = Modifier.semantics {
        liveRegion = LiveRegionMode.Assertive
    }
) { Text(error.message) }
```

---

## Voice Commands

Predefined commands for common actions.

```kotlin
class VoiceCommandProcessor {

    private val commands = listOf(
        VoiceCommand(
            patterns = listOf("start chat", "new chat", "begin chat"),
            action = VoiceCommandAction.START_CHAT,
            description = "Start a new chat"
        ),
        VoiceCommand(
            patterns = listOf("upload document", "upload file", "add document"),
            action = VoiceCommandAction.UPLOAD_DOCUMENT,
            description = "Upload a document"
        ),
        VoiceCommand(
            patterns = listOf("search documents", "search knowledge", "find documents"),
            action = VoiceCommandAction.SEARCH_DOCUMENTS,
            description = "Search the knowledge base"
        ),
        VoiceCommand(
            patterns = listOf("show agents", "list agents", "open agents"),
            action = VoiceCommandAction.SHOW_AGENTS,
            description = "View agent list"
        ),
        VoiceCommand(
            patterns = listOf("show dashboard", "open dashboard", "go to dashboard"),
            action = VoiceCommandAction.SHOW_DASHBOARD,
            description = "Navigate to dashboard"
        ),
        VoiceCommand(
            patterns = listOf("show notifications", "open notifications"),
            action = VoiceCommandAction.SHOW_NOTIFICATIONS,
            description = "View notifications"
        ),
        VoiceCommand(
            patterns = listOf("stop", "cancel", "never mind"),
            action = VoiceCommandAction.CANCEL,
            description = "Cancel current action"
        ),
        VoiceCommand(
            patterns = listOf("read that again", "repeat"),
            action = VoiceCommandAction.REPEAT,
            description = "Repeat last response"
        )
    )

    fun process(text: String): VoiceCommandMatch? {
        val normalized = text.lowercase().trim()
        for (command in commands) {
            for (pattern in command.patterns) {
                if (normalized.contains(pattern)) {
                    return VoiceCommandMatch(
                        command = command,
                        originalText = text,
                        extractedQuery = normalized.replace(pattern, "").trim()
                    )
                }
            }
        }
        return null
    }

    fun getCommandSuggestions(): List<VoiceCommand> = commands
}

data class VoiceCommand(
    val patterns: List<String>,
    val action: VoiceCommandAction,
    val description: String
)

enum class VoiceCommandAction {
    START_CHAT,
    UPLOAD_DOCUMENT,
    SEARCH_DOCUMENTS,
    SHOW_AGENTS,
    SHOW_DASHBOARD,
    SHOW_NOTIFICATIONS,
    CANCEL,
    REPEAT
}

data class VoiceCommandMatch(
    val command: VoiceCommand,
    val originalText: String,
    val extractedQuery: String
)
```

### Voice Command Help Dialog

```kotlin
@Composable
fun VoiceCommandHelpDialog(
    onDismiss: () -> Unit
) {
    val processor = remember { VoiceCommandProcessor() }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Voice Commands") },
        text = {
            LazyColumn(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                processor.getCommandSuggestions().forEach { command ->
                    Row(modifier = Modifier.fillMaxWidth()) {
                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = "\"${command.patterns.first()}\"",
                                style = MaterialTheme.typography.bodyMedium,
                                fontWeight = FontWeight.SemiBold,
                                fontFamily = FontFamily.Monospace
                            )
                            Text(
                                text = command.description,
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                }
            }
        },
        confirmButton = {
            TextButton(onClick = onDismiss) { Text("Got it") }
        }
    )
}
```

---

## Voice Input Settings

```kotlin
data class VoiceSettings(
    val language: String = "en-US",
    val sensitivity: Float = 0.5f,
    val autoStopEnabled: Boolean = true,
    val autoStopTimeoutMs: Long = 3000L,
    val hapticFeedbackEnabled: Boolean = true,
    val offlineMode: Boolean = false,
    val showPartialResults: Boolean = true,
    val maxRecordingDurationMs: Long = 60_000L
)

class VoiceSettingsManager @Inject constructor(
    private val dataStore: DataStore<Preferences>
) {
    companion object {
        private val KEY_LANGUAGE = stringPreferencesKey("voice_language")
        private val KEY_SENSITIVITY = floatPreferencesKey("voice_sensitivity")
        private val KEY_AUTO_STOP = booleanPreferencesKey("voice_auto_stop")
        private val KEY_AUTO_STOP_TIMEOUT = longPreferencesKey("voice_auto_stop_timeout")
        private val KEY_HAPTIC = booleanPreferencesKey("voice_haptic_feedback")
        private val KEY_OFFLINE = booleanPreferencesKey("voice_offline_mode")
        private val KEY_PARTIAL_RESULTS = booleanPreferencesKey("voice_partial_results")
        private val KEY_MAX_DURATION = longPreferencesKey("voice_max_duration")
    }

    val settings: Flow<VoiceSettings> = dataStore.data.map { prefs ->
        VoiceSettings(
            language = prefs[KEY_LANGUAGE] ?: "en-US",
            sensitivity = prefs[KEY_SENSITIVITY] ?: 0.5f,
            autoStopEnabled = prefs[KEY_AUTO_STOP] ?: true,
            autoStopTimeoutMs = prefs[KEY_AUTO_STOP_TIMEOUT] ?: 3000L,
            hapticFeedbackEnabled = prefs[KEY_HAPTIC] ?: true,
            offlineMode = prefs[KEY_OFFLINE] ?: false,
            showPartialResults = prefs[KEY_PARTIAL_RESULTS] ?: true,
            maxRecordingDurationMs = prefs[KEY_MAX_DURATION] ?: 60_000L
        )
    }

    suspend fun updateSettings(update: (VoiceSettings) -> VoiceSettings) {
        dataStore.edit { prefs ->
            val current = VoiceSettings(
                language = prefs[KEY_LANGUAGE] ?: "en-US",
                sensitivity = prefs[KEY_SENSITIVITY] ?: 0.5f,
                autoStopEnabled = prefs[KEY_AUTO_STOP] ?: true,
                autoStopTimeoutMs = prefs[KEY_AUTO_STOP_TIMEOUT] ?: 3000L,
                hapticFeedbackEnabled = prefs[KEY_HAPTIC] ?: true,
                offlineMode = prefs[KEY_OFFLINE] ?: false,
                showPartialResults = prefs[KEY_PARTIAL_RESULTS] ?: true,
                maxRecordingDurationMs = prefs[KEY_MAX_DURATION] ?: 60_000L
            )
            val updated = update(current)
            prefs[KEY_LANGUAGE] = updated.language
            prefs[KEY_SENSITIVITY] = updated.sensitivity
            prefs[KEY_AUTO_STOP] = updated.autoStopEnabled
            prefs[KEY_AUTO_STOP_TIMEOUT] = updated.autoStopTimeoutMs
            prefs[KEY_HAPTIC] = updated.hapticFeedbackEnabled
            prefs[KEY_OFFLINE] = updated.offlineMode
            prefs[KEY_PARTIAL_RESULTS] = updated.showPartialResults
            prefs[KEY_MAX_DURATION] = updated.maxRecordingDurationMs
        }
    }

    suspend fun getLanguage(): String = settings.first().language
}

@Composable
fun VoiceSettingsScreen(
    viewModel: VoiceSettingsViewModel = hiltViewModel()
) {
    val settings by viewModel.settings.collectAsStateWithLifecycle()

    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Text("Voice Settings", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
        }

        // Language
        item {
            LanguageSelector(
                selectedLanguage = settings.language,
                onLanguageChange = { viewModel.updateLanguage(it) }
            )
        }

        // Sensitivity
        item {
            Column {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text("Sensitivity", style = MaterialTheme.typography.bodyMedium)
                    Text(
                        when {
                            settings.sensitivity < 0.33f -> "Low"
                            settings.sensitivity < 0.66f -> "Medium"
                            else -> "High"
                        },
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                }
                Slider(
                    value = settings.sensitivity,
                    onValueChange = { viewModel.updateSensitivity(it) },
                    valueRange = 0f..1f
                )
                Text(
                    "Lower sensitivity requires louder speech",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }

        // Auto-stop
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text("Auto-stop", style = MaterialTheme.typography.bodyMedium)
                    Text(
                        "Automatically stop listening after silence",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Switch(
                    checked = settings.autoStopEnabled,
                    onCheckedChange = { viewModel.updateAutoStop(it) }
                )
            }
        }

        // Haptic feedback
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text("Haptic Feedback", style = MaterialTheme.typography.bodyMedium)
                    Text(
                        "Vibrate when recording starts/stops",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Switch(
                    checked = settings.hapticFeedbackEnabled,
                    onCheckedChange = { viewModel.updateHapticFeedback(it) }
                )
            }
        }

        // Show partial results
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text("Show Partial Results", style = MaterialTheme.typography.bodyMedium)
                    Text(
                        "Display text as you speak",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Switch(
                    checked = settings.showPartialResults,
                    onCheckedChange = { viewModel.updatePartialResults(it) }
                )
            }
        }

        // Offline mode
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text("Offline Mode", style = MaterialTheme.typography.bodyMedium)
                    Text(
                        "Use on-device recognition (less accurate)",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Switch(
                    checked = settings.offlineMode,
                    onCheckedChange = { viewModel.updateOfflineMode(it) }
                )
            }
        }
    }
}

@Composable
fun LanguageSelector(
    selectedLanguage: String,
    onLanguageChange: (String) -> Unit
) {
    val languages = listOf(
        "en-US" to "English (US)",
        "en-GB" to "English (UK)",
        "es-ES" to "Spanish",
        "fr-FR" to "French",
        "de-DE" to "German",
        "ja-JP" to "Japanese",
        "ko-KR" to "Korean",
        "zh-CN" to "Chinese (Simplified)",
        "pt-BR" to "Portuguese (BR)",
        "ar-SA" to "Arabic",
        "hi-IN" to "Hindi"
    )

    var expanded by remember { mutableStateOf(false) }

    ExposedDropdownMenuBox(expanded = expanded, onExpandedChange = { expanded = it }) {
        OutlinedTextField(
            value = languages.firstOrNull { it.first == selectedLanguage }?.second ?: "English (US)",
            onValueChange = {},
            readOnly = true,
            label = { Text("Language") },
            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
            modifier = Modifier.fillMaxWidth().menuAnchor()
        )
        ExposedDropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
            languages.forEach { (code, name) ->
                DropdownMenuItem(
                    text = { Text(name) },
                    onClick = { onLanguageChange(code); expanded = false },
                    leadingIcon = {
                        if (code == selectedLanguage) {
                            Icon(Icons.Default.Check, contentDescription = null, modifier = Modifier.size(16.dp))
                        }
                    }
                )
            }
        }
    }
}
```

---

## Voice Input in Other Screens

### Voice Search

```kotlin
@Composable
fun VoiceSearchBar(
    query: String,
    onQueryChange: (String) -> Unit,
    onSearch: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    val voiceManager = rememberVoiceInputManager(
        onResult = { text ->
            onQueryChange(text)
            onSearch(text)
        },
        onError = { /* show error */ }
    )

    val voiceState by voiceManager.state.collectAsStateWithLifecycle()
    val partialText by voiceManager.partialResults.collectAsStateWithLifecycle()

    Row(
        modifier = modifier,
        verticalAlignment = Alignment.CenterVertically
    ) {
        OutlinedTextField(
            value = if (voiceState == VoiceInputState.LISTENING && partialText.isNotEmpty()) partialText else query,
            onValueChange = onQueryChange,
            modifier = Modifier.weight(1f),
            placeholder = { Text("Search documents...") },
            leadingIcon = { Icon(Icons.Default.Search, "Search") },
            trailingIcon = {
                if (query.isNotEmpty()) {
                    IconButton(onClick = { onQueryChange("") }) {
                        Icon(Icons.Default.Clear, "Clear")
                    }
                }
            },
            singleLine = true
        )

        Spacer(Modifier.width(8.dp))

        VoiceButton(
            state = voiceState,
            onClick = {
                if (voiceState == VoiceInputState.LISTENING) {
                    voiceManager.stopListening()
                } else {
                    voiceManager.startListening()
                }
            },
            size = 40.dp
        )
    }
}
```

---

## Voice Input Offline Support

```kotlin
class OfflineSpeechRecognizer @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private var recognizer: android.speech.SpeechRecognizer? = null

    fun isOfflineAvailable(): Boolean {
        return try {
            val intent = Intent(RecognizerIntent.ACTION_RECOGNIZE_SPEECH).apply {
                putExtra(RecognizerIntent.EXTRA_PREFER_OFFLINE, true)
            }
            val resolveInfo = context.packageManager.resolveActivity(intent, 0)
            resolveInfo != null
        } catch (e: Exception) {
            false
        }
    }

    fun getOfflineLanguages(): List<String> {
        // Returns languages available for offline recognition
        return listOf("en-US", "es-ES", "fr-FR", "de-DE", "ja-JP", "ko-KR", "zh-CN")
    }

    fun createOfflineRecognizer(language: String): android.speech.SpeechRecognizer? {
        if (!isOfflineAvailable()) return null

        recognizer = android.speech.SpeechRecognizer.createSpeechRecognizer(context)
        return recognizer
    }

    fun destroy() {
        recognizer?.destroy()
        recognizer = null
    }
}
```

---

## Voice Input History

```kotlin
@Entity(tableName = "voice_history")
data class VoiceHistoryEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val text: String,
    val language: String,
    val durationMs: Long,
    val wasCommand: Boolean,
    val commandAction: String?,
    val timestamp: Long
)

@Dao
interface VoiceHistoryDao {

    @Query("SELECT * FROM voice_history ORDER BY timestamp DESC LIMIT 50")
    suspend fun getRecentInputs(): List<VoiceHistoryEntity>

    @Insert
    suspend fun insert(entity: VoiceHistoryEntity)

    @Query("DELETE FROM voice_history WHERE timestamp < :before")
    suspend fun deleteOlderThan(before: Long)

    @Query("SELECT * FROM voice_history WHERE text LIKE '%' || :query || '%' ORDER BY timestamp DESC")
    suspend fun search(query: String): List<VoiceHistoryEntity>
}

@Composable
fun VoiceHistoryList(
    history: List<VoiceHistoryEntity>,
    onItemClick: (String) -> Unit,
    onClear: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(modifier = modifier) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text("Recent Voice Inputs", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
            TextButton(onClick = onClear) { Text("Clear all") }
        }

        if (history.isEmpty()) {
            Text(
                "No voice input history",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(16.dp)
            )
        } else {
            history.forEach { item ->
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { onItemClick(item.text) }
                        .padding(horizontal = 16.dp, vertical = 10.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        if (item.wasCommand) Icons.Default.Terminal else Icons.Default.Mic,
                        contentDescription = null,
                        modifier = Modifier.size(16.dp),
                        tint = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Spacer(Modifier.width(12.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(item.text, style = MaterialTheme.typography.bodyMedium, maxLines = 1, overflow = TextOverflow.Ellipsis)
                        Text(
                            item.timestamp.toRelativeTimeString(),
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
            }
        }
    }
}
```

---

## Voice Input API Integration

For enhanced accuracy, a cloud speech-to-text API can supplement on-device recognition.

```kotlin
interface SpeechToTextApiService {

    @Multipart
    @POST("api/v1/speech/transcribe")
    suspend fun transcribe(
        @Part audio: MultipartBody.Part,
        @Part("language") language: RequestBody,
        @Part("model") model: RequestBody = "whisper-1".toRequestBody("text/plain".toMediaType()),
        @Part("format") format: RequestBody = "webm".toRequestBody("text/plain".toMediaType())
    ): TranscriptionResponse
}

@Serializable
data class TranscriptionResponse(
    val text: String,
    val language: String,
    val confidence: Float,
    val durationMs: Long
)
```

---

## Voice Input Data Models

```kotlin
data class VoiceInput(
    val id: String,
    val text: String,
    val language: String,
    val durationMs: Long,
    val confidence: Float,
    val isPartial: Boolean,
    val timestamp: Instant
)

data class VoiceInputResult(
    val text: String,
    val isCommand: Boolean,
    val commandMatch: VoiceCommandMatch?,
    val confidence: Float
)

sealed interface RecognitionSource {
    data object Device : RecognitionSource
    data class Cloud(val model: String) : RecognitionSource
}
```

---

## Voice Input Testing

### Unit Tests

```kotlin
class VoiceCommandProcessorTest {

    private val processor = VoiceCommandProcessor()

    @Test
    fun `process detects start chat command`() {
        val result = processor.process("start a new chat")
        assertNotNull(result)
        assertEquals(VoiceCommandAction.START_CHAT, result!!.command.action)
    }

    @Test
    fun `process detects upload command`() {
        val result = processor.process("upload a document")
        assertNotNull(result)
        assertEquals(VoiceCommandAction.UPLOAD_DOCUMENT, result!!.command.action)
    }

    @Test
    fun `process detects cancel command`() {
        val result = processor.process("cancel that")
        assertNotNull(result)
        assertEquals(VoiceCommandAction.CANCEL, result!!.command.action)
    }

    @Test
    fun `process returns null for non-command text`() {
        val result = processor.process("What is the capital of France")
        assertNull(result)
    }

    @Test
    fun `process is case insensitive`() {
        val result = processor.process("START CHAT")
        assertNotNull(result)
        assertEquals(VoiceCommandAction.START_CHAT, result!!.command.action)
    }
}
```

### UI Tests

```kotlin
@Test
fun voiceButton_showsIdleState() {
    composeTestRule.setContent {
        VoiceButton(state = VoiceInputState.IDLE, onClick = {})
    }
    composeTestRule.onNodeWithContentDescription("Start voice input").assertIsDisplayed()
}

@Test
fun voiceButton_showsListeningState() {
    composeTestRule.setContent {
        VoiceButton(state = VoiceInputState.LISTENING, onClick = {})
    }
    composeTestRule.onNodeWithContentDescription("Listening... Tap to stop").assertIsDisplayed()
}

@Test
fun voiceButton_togglesOnTap() {
    var clicked = false
    composeTestRule.setContent {
        VoiceButton(state = VoiceInputState.IDLE, onClick = { clicked = true })
    }
    composeTestRule.onNodeWithContentDescription("Start voice input").performClick()
    assertTrue(clicked)
}
```

---

## Voice Input Performance

| Metric                        | Target     | Notes                            |
|-------------------------------|------------|----------------------------------|
| Recognition start latency     | < 200ms    | Time from tap to "listening"     |
| Partial result update         | < 100ms    | UI refresh during listening      |
| Final result delivery         | < 500ms    | After end of speech              |
| Cloud API round-trip          | < 2s       | For cloud-based recognition      |
| Battery usage (1min session)  | < 1%       | Measured on Pixel 7              |
| Memory footprint              | < 10MB     | SpeechRecognizer instance        |

### Optimization Tips

```kotlin
// Reuse SpeechRecognizer instance instead of recreating
private var recognizer: SpeechRecognizer? = null

fun startListening() {
    if (recognizer == null) {
        recognizer = SpeechRecognizer.createSpeechRecognizer(context)
        recognizer?.setRecognitionListener(listener)
    }
    recognizer?.startListening(intent)
}

// Clean up in onCleared
override fun onCleared() {
    recognizer?.destroy()
    recognizer = null
}
```

---

## Voice Input UX Best Practices

| Practice                          | Implementation                                   |
|-----------------------------------|--------------------------------------------------|
| Visual feedback when recording    | Red pulse animation + waveform + timer           |
| Haptic feedback on start/stop     | VibrationEffect on state transitions             |
| Partial results displayed         | Live text preview below the input field          |
| Clear stop mechanism              | Tap button again or swipe down                   |
| Error feedback                    | Snackbar with retry option                       |
| Permission rationale              | Dialog explaining why mic is needed              |
| Maximum recording duration        | Auto-stop after 60 seconds with warning          |
| Cancellation support              | Cancel button visible during recording           |
| Confirmation before discard      | Dialog if text is unsent when stopping            |
| Accessible labels                 | ContentDescription on all interactive elements   |

### Haptic Feedback

```kotlin
@Composable
fun VoiceInputEffect(
    voiceState: VoiceInputState,
    hapticEnabled: Boolean
) {
    val context = LocalContext.current
    val vibrator = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
        val manager = context.getSystemService(Context.VIBRATOR_MANAGER_SERVICE) as? VibratorManager
        manager?.defaultVibrator
    } else {
        @Suppress("DEPRECATION")
        context.getSystemService(Context.VIBRATOR_SERVICE) as? Vibrator
    }

    LaunchedEffect(voiceState) {
        if (!hapticEnabled || vibrator == null) return@LaunchedEffect

        when (voiceState) {
            VoiceInputState.LISTENING -> {
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                    vibrator.vibrate(
                        VibrationEffect.createOneShot(50, VibrationEffect.DEFAULT_AMPLITUDE)
                    )
                }
            }
            VoiceInputState.PROCESSING -> {
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                    vibrator.vibrate(
                        VibrationEffect.createWaveform(longArrayOf(0, 30, 30, 30), -1)
                    )
                }
            }
            else -> {}
        }
    }
}
```

---

## Voice Input Permissions

### Permission Handling Flow

```kotlin
@Composable
fun VoicePermissionHandler(
    onGranted: @Composable () -> Unit,
    onDenied: @Composable (showRationale: Boolean) -> Unit
) {
    val context = LocalContext.current
    var hasPermission by remember {
        mutableStateOf(
            ContextCompat.checkSelfPermission(context, Manifest.permission.RECORD_AUDIO) ==
            PackageManager.PERMISSION_GRANTED
        )
    }
    var showRationale by remember { mutableStateOf(false) }

    val launcher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { granted ->
        hasPermission = granted
        if (!granted) {
            showRationale = shouldShowRequestPermissionRationale(context, Manifest.permission.RECORD_AUDIO)
        }
    }

    if (hasPermission) {
        onGranted()
    } else {
        LaunchedEffect(Unit) {
            if (!showRationale) {
                launcher.launch(Manifest.permission.RECORD_AUDIO)
            }
        }
        onDenied(showRationale)
    }
}

private fun shouldShowRequestPermissionRationale(context: Context, permission: String): Boolean {
    return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
        context is Activity && (context as Activity).shouldShowRequestPermissionRationale(permission)
    } else false
}
```

---

## Voice Input on Different Devices

| Device Type | Considerations                                   |
|-------------|--------------------------------------------------|
| Phone       | Standard layout, bottom-positioned voice button   |
| Tablet      | Larger button, centered voice overlay             |
| Foldable    | Adaptive layout between folded/unfolded states    |
| Wear OS     | Simplified UI, single button interaction          |
| Android Auto| Hands-free mode, continuous listening             |
| TV          | Remote-triggered voice input via D-pad             |

```kotlin
@Composable
fun AdaptiveVoiceLayout(
    voiceState: VoiceInputState,
    partialText: String,
    onVoiceClick: () -> Unit
) {
    val windowSize = currentWindowAdaptiveInfo().windowSizeClass

    when (windowSize.windowWidthSizeClass) {
        WindowWidthSizeClass.COMPACT -> {
            // Phone: voice button in the input bar
            CompactVoiceLayout(voiceState, partialText, onVoiceClick)
        }
        WindowWidthSizeClass.MEDIUM, WindowWidthSizeClass.EXPANDED -> {
            // Tablet: centered floating voice overlay
            TabletVoiceLayout(voiceState, partialText, onVoiceClick)
        }
    }
}

@Composable
fun CompactVoiceLayout(
    state: VoiceInputState,
    partialText: String,
    onClick: () -> Unit
) {
    Row(
        modifier = Modifier.padding(8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        VoiceButton(state = state, onClick = onClick, size = 48.dp)
        if (state == VoiceInputState.LISTENING && partialText.isNotEmpty()) {
            Spacer(Modifier.width(8.dp))
            Text(partialText, style = MaterialTheme.typography.bodySmall, modifier = Modifier.weight(1f))
        }
    }
}

@Composable
fun TabletVoiceLayout(
    state: VoiceInputState,
    partialText: String,
    onClick: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        VoiceButtonWithLabel(
            state = state,
            partialText = partialText,
            onClick = onClick,
            size = 80.dp
        )
        if (state == VoiceInputState.LISTENING) {
            Spacer(Modifier.height(24.dp))
            WaveformVisualizer(
                rmsLevel = 0f,
                isActive = true,
                modifier = Modifier.width(300.dp)
            )
        }
    }
}
```

---

## Summary

| Component                   | File / Composable              | Notes                         |
|-----------------------------|--------------------------------|-------------------------------|
| Voice Button                | `VoiceButton`                  | 3-state animated button       |
| Pulse Animation             | `PulseAnimation`               | Listening indicator           |
| Waveform                    | `WaveformVisualizer`           | Real-time audio visualization |
| Recording Timer             | `RecordingTimer`               | Elapsed time display          |
| PromptBox Integration       | `PromptBox`                    | Chat voice input              |
| Voice Search                | `VoiceSearchBar`               | Document search voice input   |
| Voice Commands              | `VoiceCommandProcessor`        | Predefined command matching   |
| Voice Settings              | `VoiceSettingsScreen`          | Language, sensitivity, etc.   |
| Permission Handler          | `VoicePermissionHandler`       | RECORD_AUDIO permission       |
| Error Handling              | `VoiceErrorSnackbar`           | Error display + retry         |
| History                     | `VoiceHistoryList`             | Recent voice inputs           |
| Recognition Manager         | `VoiceRecognitionManager`      | Android SpeechRecognizer wrap |
| Settings Manager            | `VoiceSettingsManager`         | DataStore preferences         |
| Offline Support             | `OfflineSpeechRecognizer`      | On-device recognition         |
| Cloud API                   | `SpeechToTextApiService`       | Enhanced accuracy backend     |
| Adaptive Layout             | `AdaptiveVoiceLayout`          | Phone/tablet/foldable         |
| Haptic Feedback             | `VoiceInputEffect`             | Vibration on state changes    |
