# Voice Assistant

The voice assistant module provides speech-to-text input throughout the app, enabling users to interact with AI agents and search documents using voice. This document covers every aspect of the voice input feature.

## Architecture Overview

```
+------------------------------------------------------------------+
|                    Voice Assistant                                 |
+------------------------------------------------------------------+
|                                                                    |
|  +----------------+    +----------------+    +----------------+  |
|  | Voice Button   |--->| Speech        |--->| Text Output    |  |
|  | (Tap to Record)|    | Recognizer    |    | (PromptBox)    |  |
|  +----------------+    +----------------+    +----------------+  |
|         |                     |                     |              |
|         v                     v                     v              |
|  +----------------+    +----------------+    +----------------+  |
|  | Animation      |    | Permission     |    | API            |  |
|  | (Pulse/Wave)   |    | Manager        |    | Integration    |  |
|  +----------------+    +----------------+    +----------------+  |
|                                                                    |
+------------------------------------------------------------------+
```

## Data Flow

```
User presses mic button
         |
         v
+------------------+
| Request mic      |
| permission       |
+------------------+
         | (granted)
         v
+------------------+
| Start recording  |
| + animation      |
+------------------+
         |
         v
+------------------+
| Speech frames    |<---- SFSpeechRecognizer
| processed        |      continuous recognition
+------------------+
         |
         v
+------------------+
| Partial results  |----> Live text preview
| (real-time)      |
+------------------+
         |
         v (user releases)
+------------------+
| Final recognition|
+------------------+
         |
         v
+------------------+
| Text sent to     |
| PromptBox        |
+------------------+
```

## Voice Input States

```
+--------+    +------------+    +-------------+    +---------+
|  Idle  |--->| Listening  |--->| Processing  |--->| Success |
+--------+    +------------+    +-------------+    +---------+
     ^              |                  |                |
     |              v                  v                |
     |        +------------+    +-----------+          |
     +--------| Permission |    |  Error    |<---------+
              | Denied     |    +-----------+
              +------------+
```

## Data Models

```swift
// MARK: - Voice Input State

enum VoiceInputState: Equatable {
    case idle
    case listening
    case processing
    case success(String)
    case error(VoiceError)

    var isActive: Bool {
        switch self {
        case .listening, .processing: return true
        default: return false
        }
    }
}

enum VoiceError: LocalizedError {
    case microphonePermissionDenied
    case speechRecognitionPermissionDenied
    case noMicrophone
    case recognitionFailed(String)
    case networkError(String)
    case noSpeechDetected
    case audioSessionFailed
    case notAvailable
    case timeout

    var errorDescription: String? {
        switch self {
        case .microphonePermissionDenied:
            return "Microphone access is required for voice input. Please enable it in Settings."
        case .speechRecognitionPermissionDenied:
            return "Speech recognition permission is required. Please enable it in Settings."
        case .noMicrophone:
            return "No microphone detected on this device."
        case .recognitionFailed(let msg):
            return "Recognition failed: \(msg)"
        case .networkError(let msg):
            return "Network error: \(msg)"
        case .noSpeechDetected:
            return "No speech detected. Please try again."
        case .audioSessionFailed:
            return "Failed to configure audio session."
        case .notAvailable:
            return "Speech recognition is not available on this device."
        case .timeout:
            return "Recognition timed out. Please try again."
        }
    }

    var icon: String {
        switch self {
        case .microphonePermissionDenied, .speechRecognitionPermissionDenied:
            return "mic.slash.fill"
        case .noMicrophone: return "mic.slash"
        case .recognitionFailed, .audioSessionFailed: return "exclamationmark.triangle"
        case .networkError: return "wifi.slash"
        case .noSpeechDetected: return "waveform"
        case .notAvailable: return "xmark.circle"
        case .timeout: return "clock.fill"
        }
    }
}

// MARK: - Voice Configuration

struct VoiceConfiguration {
    var language: VoiceLanguage
    var sensitivity: RecognitionSensitivity
    var offlineMode: Bool
    var hapticFeedback: Bool
    var autoSend: Bool
    var maxDuration: TimeInterval

    static let `default` = VoiceConfiguration(
        language: .english,
        sensitivity: .medium,
        offlineMode: false,
        hapticFeedback: true,
        autoSend: false,
        maxDuration: 60
    )
}

enum VoiceLanguage: String, CaseIterable, Identifiable {
    case english = "en-US"
    case spanish = "es-ES"
    case french = "fr-FR"
    case german = "de-DE"
    case japanese = "ja-JP"
    case chinese = "zh-CN"
    case arabic = "ar-SA"
    case portuguese = "pt-BR"
    case russian = "ru-RU"
    case korean = "ko-KR"

    var id: String { rawValue }

    var displayName: String {
        switch self {
        case .english: return "English"
        case .spanish: return "Spanish"
        case .french: return "French"
        case .german: return "German"
        case .japanese: return "Japanese"
        case .chinese: return "Chinese"
        case .arabic: return "Arabic"
        case .portuguese: return "Portuguese"
        case .russian: return "Russian"
        case .korean: return "Korean"
        }
    }

    var flag: String {
        switch self {
        case .english: return "🇺🇸"
        case .spanish: return "🇪🇸"
        case .french: return "🇫🇷"
        case .german: return "🇩🇪"
        case .japanese: return "🇯🇵"
        case .chinese: return "🇨🇳"
        case .arabic: return "🇸🇦"
        case .portuguese: return "🇧🇷"
        case .russian: return "🇷🇺"
        case .korean: return "🇰🇷"
        }
    }
}

enum RecognitionSensitivity: String, CaseIterable, Identifiable {
    case low, medium, high
    var id: String { rawValue }

    var silenceTimeout: TimeInterval {
        switch self {
        case .low: return 3.0
        case .medium: return 2.0
        case .high: return 1.0
        }
    }
}

// MARK: - Voice Command

struct VoiceCommand: Identifiable {
    let id = UUID()
    let phrase: String
    let action: VoiceCommandAction
    let description: String
}

enum VoiceCommandAction {
    case startNewChat
    case searchDocuments
    case openDashboard
    case openAgents
    case openKnowledgeCenter
    case goBack
    case cancel
    case submit
    case custom(String)
}

// MARK: - Voice History

struct VoiceInputRecord: Codable, Identifiable {
    let id: String
    let text: String
    let language: String
    let duration: TimeInterval
    let confidence: Double
    let timestamp: Date
    let screen: String?

    enum CodingKeys: String, CodingKey {
        case id, text, language, duration, confidence, timestamp, screen
    }
}
```

## Voice Manager

```swift
import SwiftUI
import Speech
import AVFoundation
import Combine

@MainActor
final class VoiceManager: ObservableObject {
    @Published var state: VoiceInputState = .idle
    @Published var partialText = ""
    @Published var finalText = ""
    @Published var configuration = VoiceConfiguration.default
    @Published var permissionGranted = false

    private let speechRecognizer: SFSpeechRecognizer?
    private var recognitionRequest: SFSpeechAudioBufferRecognitionRequest?
    private var recognitionTask: SFSpeechRecognitionTask?
    private let audioEngine = AVAudioEngine()
    private var inputNode: AVAudioInputNode?
    private var silenceTimer: Timer?
    private var durationTimer: Timer?
    private var recordingStartTime: Date?

    private let historyService: VoiceHistoryService
    private let hapticGenerator = UIImpactFeedbackGenerator(style: .medium)

    init(historyService: VoiceHistoryService = .shared) {
        self.historyService = historyService
        self.speechRecognizer = SFSpeechRecognizer(locale: Locale(identifier: configuration.language.rawValue))
        self.inputNode = audioEngine.inputNode
        checkPermissions()
    }

    // MARK: - Permission Management

    func checkPermissions() {
        Task {
            let micStatus = AVCaptureDevice.authorizationStatus(for: .audio)
            let speechStatus = SFSpeechRecognizer.authorizationStatus()

            switch (micStatus, speechStatus) {
            case (.authorized, .authorized):
                permissionGranted = true
            case (.notDetermined, _):
                let granted = await AVCaptureDevice.requestAccess(for: .audio)
                permissionGranted = granted && speechStatus == .authorized
            case (_, .notDetermined):
                let status = await withCheckedContinuation { continuation in
                    SFSpeechRecognizer.requestAuthorization { status in
                        continuation.resume(returning: status)
                    }
                }
                permissionGranted = micStatus == .authorized && status == .authorized
            default:
                permissionGranted = false
            }
        }
    }

    func openSettings() {
        guard let url = URL(string: UIApplication.openSettingsURLString) else { return }
        UIApplication.shared.open(url)
    }

    // MARK: - Recording Control

    func startListening() {
        guard permissionGranted else {
            state = .error(.microphonePermissionDenied)
            return
        }
        guard let recognizer = speechRecognizer, recognizer.isAvailable else {
            state = .error(.notAvailable)
            return
        }

        do {
            try configureAudioSession()
        } catch {
            state = .error(.audioSessionFailed)
            return
        }

        recognitionRequest = SFSpeechAudioBufferRecognitionRequest()
        guard let recognitionRequest = recognitionRequest else {
            state = .error(.recognitionFailed("Failed to create request"))
            return
        }

        recognitionRequest.shouldReportPartialResults = true
        recognitionRequest.taskHint = .dictation

        if #available(iOS 16, *) {
            recognitionRequest.addsPunctuation = true
        }

        recognitionTask = recognizer.recognitionTask(with: recognitionRequest) { [weak self] result, error in
            Task { @MainActor [weak self] in
                guard let self else { return }

                if let result = result {
                    let text = result.bestTranscription.formattedString
                    self.partialText = text
                    self.resetSilenceTimer()
                }

                if error != nil || (result?.isFinal ?? false) {
                    self.finishListening()
                }
            }
        }

        let recordingFormat = inputNode?.outputFormat(forBus: 0)
        inputNode?.installTap(onBus: 0, bufferSize: 1024, format: recordingFormat) { [weak self] buffer, _ in
            self?.recognitionRequest?.append(buffer)
        }

        audioEngine.prepare()
        try audioEngine.start()

        recordingStartTime = Date()
        state = .listening
        partialText = ""
        finalText = ""

        if configuration.hapticFeedback {
            hapticGenerator.impactOccurred()
        }

        startDurationTimer()
        resetSilenceTimer()
    }

    func stopListening() {
        silenceTimer?.invalidate()
        durationTimer?.invalidate()
        recognitionTask?.cancel()
        recognitionTask = nil
        recognitionRequest?.endAudio()
        recognitionRequest = nil
        audioEngine.stop()
        audioEngine.inputNode.removeTap(onBus: 0)
        try? AVAudioSession.sharedInstance().setActive(false)

        if configuration.hapticFeedback {
            let generator = UINotificationFeedbackGenerator()
            generator.notificationOccurred(.success)
        }
    }

    func cancelListening() {
        stopListening()
        state = .idle
        partialText = ""
        finalText = ""
    }

    // MARK: - Internal

    private func finishListening() {
        stopListening()

        let text = partialText.trimmingCharacters(in: .whitespacesAndNewlines)

        if text.isEmpty {
            state = .error(.noSpeechDetected)
            return
        }

        finalText = text
        state = .success(text)

        let duration = recordingStartTime.map { Date().timeIntervalSince($0) } ?? 0
        Task {
            await historyService.saveRecord(
                VoiceInputRecord(
                    id: UUID().uuidString,
                    text: text,
                    language: configuration.language.rawValue,
                    duration: duration,
                    confidence: 1.0,
                    timestamp: Date(),
                    screen: nil
                )
            )
        }
    }

    private func configureAudioSession() throws {
        let session = AVAudioSession.sharedInstance()
        try session.setCategory(.record, mode: .measurement, options: .duckOthers)
        try session.setActive(true, options: .notifyOthersOnDeactivation)
    }

    private func resetSilenceTimer() {
        silenceTimer?.invalidate()
        silenceTimer = Timer.scheduledTimer(
            withTimeInterval: configuration.sensitivity.silenceTimeout,
            repeats: false
        ) { [weak self] _ in
            Task { @MainActor [weak self] in
                self?.stopListening()
            }
        }
    }

    private func startDurationTimer() {
        durationTimer = Timer.scheduledTimer(withTimeInterval: 0.1, repeats: true) { [weak self] _ in
            Task { @MainActor [weak self] in
                guard let self, let start = self.recordingStartTime else { return }
                let elapsed = Date().timeIntervalSince(start)
                if elapsed >= self.configuration.maxDuration {
                    self.stopListening()
                }
            }
        }
    }

    func updateLanguage(_ language: VoiceLanguage) {
        configuration.language = language
        speechRecognizer?.delegate = nil
        // Reinitialize would happen on next listen
    }
}
```

## Voice Button View

```swift
struct VoiceButtonView: View {
    @ObservedObject var voiceManager: VoiceManager
    let onTextCaptured: (String) -> Void
    var size: ButtonSize = .medium

    enum ButtonSize {
        var diameter: CGFloat {
            switch self {
            case .small: return 36
            case .medium: return 44
            case .large: return 56
            }
        }
    }

    @State private var isAnimating = false

    var body: some View {
        Button(action: handleTap) {
            ZStack {
                if voiceManager.state == .listening {
                    pulseRings
                }

                Circle()
                    .fill(backgroundColor)
                    .frame(width: size.diameter, height: size.diameter)
                    .overlay(
                        Image(systemName: iconName)
                            .font(.system(size: size.diameter * 0.4))
                            .foregroundStyle(iconColor)
                    )
            }
        }
        .buttonStyle(.plain)
        .accessibilityLabel(accessibilityLabel)
        .accessibilityHint("Double tap to start voice input")
    }

    private var pulseRings: some View {
        ForEach(0..<3) { index in
            Circle()
                .stroke(.red.opacity(0.3), lineWidth: 2)
                .frame(width: size.diameter, height: size.diameter)
                .scaleEffect(isAnimating ? 2.0 : 1.0)
                .opacity(isAnimating ? 0 : 0.6)
                .animation(
                    .easeOut(duration: 1.5)
                    .repeatForever(autoreverses: false)
                    .delay(Double(index) * 0.3),
                    value: isAnimating
                )
        }
    }

    private var iconName: String {
        switch voiceManager.state {
        case .idle: return "mic.fill"
        case .listening: return "stop.fill"
        case .processing: return "waveform"
        case .success: return "checkmark"
        case .error: return "mic.slash.fill"
        }
    }

    private var backgroundColor: Color {
        switch voiceManager.state {
        case .idle: return .blue
        case .listening: return .red
        case .processing: return .orange
        case .success: return .green
        case .error: return .gray
        }
    }

    private var iconColor: Color {
        switch voiceManager.state {
        case .idle: return .white
        case .listening: return .white
        case .processing: return .white
        case .success: return .white
        case .error: return .white
        }
    }

    private var accessibilityLabel: String {
        switch voiceManager.state {
        case .idle: return "Start voice input"
        case .listening: return "Stop voice input"
        case .processing: return "Processing speech"
        case .success: return "Voice input complete"
        case .error(let err): return "Voice input error: \(err.localizedDescription)"
        }
    }

    private func handleTap() {
        switch voiceManager.state {
        case .idle:
            voiceManager.startListening()
            withAnimation { isAnimating = true }
        case .listening:
            voiceManager.stopListening()
            withAnimation { isAnimating = false }
            if case .success(let text) = voiceManager.state {
                onTextCaptured(text)
            }
        case .error:
            voiceManager.state = .idle
            voiceManager.startListening()
        default:
            break
        }
    }
}
```

## Voice Waveform Visualization

```swift
struct VoiceWaveformView: View {
    @ObservedObject var voiceManager: VoiceManager
    @State private var amplitudes: [CGFloat] = Array(repeating: 0, count: 50)
    @State private var timer: Timer?

    var body: some View {
        HStack(spacing: 2) {
            ForEach(amplitudes.indices, id: \.self) { index in
                RoundedRectangle(cornerRadius: 1)
                    .fill(barColor)
                    .frame(width: 3, height: max(2, amplitudes[index] * 40))
                    .animation(.easeInOut(duration: 0.1), value: amplitudes[index])
            }
        }
        .frame(height: 44)
        .onAppear { startMonitoring() }
        .onDisappear { stopMonitoring() }
    }

    private var barColor: Color {
        voiceManager.state == .listening ? .red : .gray.opacity(0.3)
    }

    private func startMonitoring() {
        timer = Timer.scheduledTimer(withTimeInterval: 0.05, repeats: true) { _ in
            if voiceManager.state == .listening {
                let amplitude = CGFloat.random(in: 0.1...1.0)
                amplitudes.removeFirst()
                amplitudes.append(amplitude)
            } else {
                amplitudes = amplitudes.map { max($0 - 0.1, 0) }
            }
        }
    }

    private func stopMonitoring() {
        timer?.invalidate()
        timer = nil
    }
}
```

## Voice Input in Chat (PromptBox)

```swift
struct PromptBox: View {
    @Binding var text: String
    @ObservedObject var voiceManager: VoiceManager
    @State private var showVoicePreview = false
    let onSubmit: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            if showVoicePreview, case .success(let recognized) = voiceManager.state {
                voicePreviewBanner(text: recognized)
            }

            HStack(spacing: 12) {
                voiceButton
                textInput
                sendButton
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .background(.regularMaterial)
            .clipShape(RoundedRectangle(cornerRadius: 20))
        }
    }

    private var voiceButton: some View {
        VoiceButtonView(voiceManager: voiceManager) { recognizedText in
            text = recognizedText
            showVoicePreview = true
        }
    }

    private var textInput: some View {
        TextField("Type or speak...", text: $text, axis: .vertical)
            .textFieldStyle(.plain)
            .lineLimit(1...5)
            .onChange(of: voiceManager.partialText) { newValue in
                if voiceManager.state == .listening {
                    text = newValue
                }
            }
    }

    private var sendButton: some View {
        Button(action: {
            onSubmit()
            showVoicePreview = false
        }) {
            Image(systemName: "arrow.up.circle.fill")
                .font(.title2)
                .foregroundStyle(text.trimmingCharacters(in: .whitespaces).isEmpty ? .gray : .blue)
        }
        .disabled(text.trimmingCharacters(in: .whitespaces).isEmpty)
    }

    private func voicePreviewBanner(text: String) -> some View {
        HStack {
            Image(systemName: "mic.fill")
                .foregroundStyle(.red)
            Text(text)
                .font(.caption)
                .lineLimit(1)
            Spacer()
            Button("Use") { showVoicePreview = false }
                .font(.caption.bold())
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 6)
        .background(.blue.opacity(0.1))
    }
}
```

## Voice Input Error View

```swift
struct VoiceErrorView: View {
    let error: VoiceError
    let onRetry: () -> Void
    let onOpenSettings: () -> Void

    var body: some View {
        VStack(spacing: 12) {
            Image(systemName: error.icon)
                .font(.system(size: 32))
                .foregroundStyle(.red)
            Text(error.localizedDescription)
                .font(.subheadline)
                .multilineTextAlignment(.center)

            if isPermissionError {
                Button("Open Settings", action: onOpenSettings)
                    .buttonStyle(.borderedProminent)
            } else {
                Button("Try Again", action: onRetry)
                    .buttonStyle(.borderedProminent)
            }
        }
        .padding()
    }

    private var isPermissionError: Bool {
        if case .microphonePermissionDenied = error { return true }
        if case .speechRecognitionPermissionDenied = error { return true }
        return false
    }
}
```

## Voice Commands

```swift
struct VoiceCommandsRegistry {
    static let commands: [VoiceCommand] = [
        VoiceCommand(phrase: "new chat", action: .startNewChat, description: "Start a new conversation"),
        VoiceCommand(phrase: "search documents", action: .searchDocuments, description: "Search the knowledge center"),
        VoiceCommand(phrase: "open dashboard", action: .openDashboard, description: "Navigate to dashboard"),
        VoiceCommand(phrase: "open agents", action: .openAgents, description: "Navigate to agent management"),
        VoiceCommand(phrase: "knowledge center", action: .openKnowledgeCenter, description: "Open knowledge center"),
        VoiceCommand(phrase: "go back", action: .goBack, description: "Navigate back"),
        VoiceCommand(phrase: "cancel", action: .cancel, description: "Cancel current action"),
        VoiceCommand(phrase: "submit", action: .submit, description: "Submit current input")
    ]

    static func matchCommand(_ text: String) -> VoiceCommand? {
        let lowered = text.lowercased().trimmingCharacters(in: .whitespacesAndNewlines)
        return commands.first { command in
            lowered == command.phrase || lowered.hasPrefix(command.phrase)
        }
    }
}
```

## Voice History Service

```swift
final class VoiceHistoryService {
    static let shared = VoiceHistoryService()
    private let defaults = UserDefaults.standard
    private let maxRecords = 50
    private let key = "voice_history"

    func saveRecord(_ record: VoiceInputRecord) async {
        var records = await getHistory()
        records.insert(record, at: 0)
        if records.count > maxRecords {
            records = Array(records.prefix(maxRecords))
        }
        if let data = try? JSONEncoder().encode(records) {
            defaults.set(data, forKey: key)
        }
    }

    func getHistory() async -> [VoiceInputRecord] {
        guard let data = defaults.data(forKey: key),
              let records = try? JSONDecoder().decode([VoiceInputRecord].self, from: data) else {
            return []
        }
        return records
    }

    func clearHistory() async {
        defaults.removeObject(forKey: key)
    }
}
```

## Voice Settings View

```swift
struct VoiceSettingsView: View {
    @ObservedObject var voiceManager: VoiceManager
    @AppStorage("voice_offline_mode") private var offlineMode = false
    @AppStorage("voice_haptic_feedback") private var hapticFeedback = true
    @AppStorage("voice_auto_send") private var autoSend = false

    var body: some View {
        Form {
            Section("Language") {
                Picker("Language", selection: Binding(
                    get: { voiceManager.configuration.language },
                    set: { voiceManager.updateLanguage($0) }
                )) {
                    ForEach(VoiceLanguage.allCases) { lang in
                        HStack {
                            Text(lang.flag)
                            Text(lang.displayName)
                        }
                        .tag(lang)
                    }
                }
            }

            Section("Recognition") {
                Picker("Sensitivity", selection: $voiceManager.configuration.sensitivity) {
                    ForEach(RecognitionSensitivity.allCases) { sensitivity in
                        VStack(alignment: .leading) {
                            Text(sensitivity.rawValue.capitalized)
                            Text("Silence timeout: \(sensitivity.silenceTimeout, specifier: "%.1f")s")
                                .font(.caption).foregroundStyle(.secondary)
                        }
                        .tag(sensitivity)
                    }
                }

                Toggle("Offline Mode", isOn: $offlineMode)
                    .onChange(of: offlineMode) { value in
                        voiceManager.configuration.offlineMode = value
                    }
            }

            Section("Feedback") {
                Toggle("Haptic Feedback", isOn: $hapticFeedback)
                    .onChange(of: hapticFeedback) { value in
                        voiceManager.configuration.hapticFeedback = value
                    }
                Toggle("Auto-send on completion", isOn: $autoSend)
                    .onChange(of: autoSend) { value in
                        voiceManager.configuration.autoSend = value
                    }
            }

            Section("Permissions") {
                HStack {
                    Text("Microphone")
                    Spacer()
                    PermissionStatusIndicator(granted: voiceManager.permissionGranted)
                }
                Button("Open Settings") {
                    voiceManager.openSettings()
                }
            }
        }
        .navigationTitle("Voice Settings")
    }
}

struct PermissionStatusIndicator: View {
    let granted: Bool
    var body: some View {
        HStack(spacing: 4) {
            Circle()
                .fill(granted ? .green : .red)
                .frame(width: 8, height: 8)
            Text(granted ? "Granted" : "Denied")
                .font(.caption)
                .foregroundStyle(.secondary)
        }
    }
}
```

## Voice Input in Other Screens

```swift
struct VoiceSearchBar: View {
    @ObservedObject var voiceManager: VoiceManager
    @Binding var searchText: String
    @State private var showVoiceSearch = false

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: "magnifyingglass")
                .foregroundStyle(.secondary)

            TextField("Search...", text: $searchText)
                .textFieldStyle(.plain)

            VoiceButtonView(voiceManager: voiceManager, size: .small) { text in
                searchText = text
            }
        }
        .padding(8)
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 10))
    }
}
```

## Voice Input Offline Support

```swift
extension SFSpeechRecognizer {
    var isOfflineSupported: Bool {
        return supportsOnDeviceRecognition
    }

    func createOfflineRequest() -> SFSpeechAudioBufferRecognitionRequest {
        let request = SFSpeechAudioBufferRecognitionRequest()
        if #available(iOS 16, *) {
            request.requiresOnDeviceRecognition = true
        }
        return request
    }
}

extension VoiceManager {
    func startOfflineListening() {
        guard let recognizer = speechRecognizer, recognizer.supportsOnDeviceRecognition else {
            state = .error(.notAvailable)
            return
        }

        let offlineRequest = recognizer.createOfflineRequest()
        recognitionRequest = offlineRequest
        // ... same recording setup as startListening
    }
}
```

## Voice Input Accessibility

```swift
struct VoiceAccessibilityModifier: ViewModifier {
    let state: VoiceInputState
    let partialText: String

    func body(content: Content) -> some View {
        content
            .accessibilityElement(children: .combine)
            .accessibilityLabel(accessibilityLabel)
            .accessibilityHint(accessibilityHint)
            .accessibilityAddTraits(state == .listening ? .isPlaying : [])
            .accessibilityRemoveTraits(state == .listening ? .isStaticText : [])
    }

    private var accessibilityLabel: String {
        switch state {
        case .idle: return "Voice input button, idle"
        case .listening: return "Voice input active. Listening..."
        case .processing: return "Processing voice input..."
        case .success(let text): return "Voice recognized: \(text)"
        case .error(let error): return "Voice error: \(error.localizedDescription)"
        }
    }

    private var accessibilityHint: String {
        switch state {
        case .idle: return "Double tap to start voice recognition"
        case .listening: return "Double tap to stop. Currently hearing: \(partialText)"
        case .processing: return "Processing your voice input"
        case .success: return "Text has been inserted"
        case .error: return "Double tap to retry"
        }
    }
}
```

## Performance Considerations

| Aspect | Strategy |
|---|---|
| Latency | On-device recognition when possible |
| Battery | Auto-stop after timeout, silence detection |
| Memory | Release audio buffers after processing |
| Accuracy | Configurable sensitivity per user |
| Network | Offline mode fallback |
| Permissions | Graceful degradation with settings redirect |
| Haptic | Light feedback on state changes |

## Testing

```swift
import XCTest
@testable import NexusAI

final class VoiceManagerTests: XCTestCase {
    func testInitialState() {
        let manager = VoiceManager()
        XCTAssertEqual(manager.state, .idle)
        XCTAssertFalse(manager.partialText.isEmpty || true)
    }

    func testVoiceCommandMatching() {
        let match = VoiceCommandsRegistry.matchCommand("new chat")
        XCTAssertNotNil(match)
        XCTAssertEqual(match?.action, .startNewChat)

        let noMatch = VoiceCommandsRegistry.matchCommand("random text")
        XCTAssertNil(noMatch)
    }

    func testSilenceTimeout() {
        let sensitivity = RecognitionSensitivity.high
        XCTAssertEqual(sensitivity.silenceTimeout, 1.0)
    }

    func testLanguageFlags() {
        XCTAssertEqual(VoiceLanguage.english.flag, "🇺🇸")
        XCTAssertEqual(VoiceLanguage.japanese.flag, "🇯🇵")
    }
}
```

## File Structure

```
VoiceAssistant/
├── Views/
│   ├── VoiceButtonView.swift
│   ├── VoiceWaveformView.swift
│   ├── VoiceErrorView.swift
│   ├── VoiceSettingsView.swift
│   ├── VoiceSearchBar.swift
│   └── PromptBox.swift
├── ViewModels/
│   └── VoiceManager.swift
├── Models/
│   ├── VoiceInputState.swift
│   ├── VoiceConfiguration.swift
│   ├── VoiceCommand.swift
│   └── VoiceInputRecord.swift
├── Services/
│   └── VoiceHistoryService.swift
├── Extensions/
│   └── VoiceAccessibility.swift
└── Tests/
    ├── VoiceManagerTests.swift
    └── VoiceCommandTests.swift
```
