# UI Components

## AeroXe Nexus AI — iOS Design system & Component Library

**Version: 1.0 | Last Updated: July 2026**

---

# Table of Contents

1. [Design system](#1-design-system)
2. [Theme configuration](#2-theme-configuration)
3. [Dynamic color](#3-dynamic-color)
4. [Color system](#4-color-system)
5. [Typography system](#5-typography-system)
6. [shape system](#6-shape-system)
7. [Spacing system](#7-spacing-system)
8. [Button components](#8-button-components)
9. [Card compnents](#9-card-components)
10. [Modal components](#10-modal-components)
11. [List components](#11-list-components)
12. [Input components](#12-input-components)
13. [Feedback components](#13-feedback-components)
14. [Navigation components](#14-navigation-components)
15. [Media components](#15-media-components)
16. [Chat components](#16-chat-components)
17. [Agent components](#17-agent-components)
18. [Document components](#18-document-components)
19. [Chart components](#19-chart-components)
20. [Loading components](#20-loading-components)
21. [Empty state components](#21-empty-state-components)
22. [Error state components](#22-error-state-components)
23. [Animation components](#23-animation-components)
24. [Icon system](#24-icon-system)
25. [Responsive layout](#25-responsive-layout)
26. [iPad layout](#26-ipad-layout)
27. [Dark mode support](#27-dark-mode-support)
28. [Dynamic type](#28-dynamic-type)
29. [RTL support](#29-rtl-support)
30. [Haptic feedback](#30-haptic-feedback)

---

# 1. Design System

The design system is built with pure SwiftUI, targeting iOS 17+ with dynamic type support as first-class.

```swift
@available(iOS 17.0, *)
struct NexusUI { }
```

| Foundation | Technology | Version |
|-----------|------------|---------|
| UI framework | SwiftUI | iOS 17+ |

---

# 2. Theme configuration

```swift
struct ThemeKey: EnvironmentKey {
    static let defaultValue: AppTheme = .standard
}

extension EnvironmentValues {
    var appTheme: ThemeKey.Value {
        get { self[ThemeKey.self] }
        set { self[ThemeKey.self] = newValue }
    }
}

struct AppTheme {
    let colors: ColorPalette
    let typography: TypographyPalette
    let shapes: ShapePalette
    let spacing: SpacingPalette

    static let standard = AppTheme(
        colors: ColorPalette.light,
        typography: TypographyPalette.standard,
        shapes: ShapePalette.standard,
        spacing: SpacingPalette.standard
    )
}
```

## Light / dark switching

```swift
@main
struct NexusAiApp: App {
    @AppStorage("themePreference")
    private var theme: ThemePreference = .auto

    var body: some Scene {
        WindowGroup {
            AppRootView()
                .environment(\.appTheme, AppTheme.standard)
                .preferredColorScheme(theme.colorScheme)
        }
    }
}
```

| preference | resolved scheme |
|------------|----------------|
| .light | ColorScheme.light |

---

# 3. Dynamic color

```swift
extension Color {
    static let appPrimary = Color("primary")
    static let appSecondary = Color("secondary")

    struct adaptive {
        static let surface = Color(UIColor { trait in
            let isDark = trait.userInterfaceStyle == .dark
            return isDark ? UIColor(white: 0.1) : UIColor.white
        })
    }
}
```

## Asset catalog reference

```
Assets.xcassets/
  ├─ AccentColor.colorset/
  │   ├─ Any (Light): #2563EB
  │   └─ Dark:         #60A5FA
```

| role | light | dark |
|------|-------|------|
| primary | #2563EB | #60A5FA |
| secondary | #7C3AED | #A78BFA |

---

# 4. Color system

```swift
struct ColorPalette {
    let primary: Color
    let primaryContainer: Color
    let secondary: Color
    let error: Color
    let errorContainer: Color
    let neutral01: Color
    let neutral09: Color
    let surface: Color
    let surfaceSecondary: Color
    let background: Color

    static let light = ColorPalette(
        primary: Color(hex: "#2563EB"),
        primaryContainer: Color(hex: "#DBEAFE"),
        secondary: Color(hex: "#7C3AED"),
        error: Color(hex: "#DC2626"),
        errorContainer: Color(hex: "#FEF2F2"),
        neutral01: Color(hex: "#F8FAFC"),
        neutral09: Color(hex: "#0F172A"),
        surface: .white,
        surfaceSecondary: Color(hex: "#F1F5F9"),
        background: Color(hex: "#FAFAFA")
    )
}
```

| color token | hex (light) | hex (dark) | usage |
|-------------|-------------|------------|-------|
| primary | #2563EB | #60A5FA | buttons, links, active |

---

# 5. typography system

```swift
struct TypographyPalette {
    let largeTitle: Font
    let title: Font
    let title2: Font
    let title3: Font
    let headline: Font
    let body: Font
    let bodyBold: Font
    let callout: Font
    let caption: Font
    let caption2: Font

    static let standard = TypographyPalette(
        largeTitle: .largeTitle.weight(.bold),
        title: .title.weight(.bold),
        title2: .title2.weight(.bold),
        title3: .title3.weight(.semibold),
        headline: .headline,
        body: .body,
        bodyBold: .body.weight(.bold),
        callout: .callout,
        caption: .caption,
        caption2: .caption2
    )
}
```

| text style | point | usage |
|------------|-------|-------|
| largeTitle | 34 | hero |

---

# 6. shape system

```swift
struct ShapePalette {
    let smallCorner: CGFloat = 8
    let mediumCorner: CGFloat = 12
    let largeCorner: CGFloat = 16
    let extraLargeCorner: CGFloat = 24
    let capsule: RoundedCornerStyle = true
}
```

| shape | value |
|-------|-------|

---

# 7. spacing system

```swift
struct SpacingPalette {
    let xxsmall: CGFloat = 2
    let xsmall: CGFloat = 4
    let small: CGFloat = 8
    let medium: CGFloat = 12
    let large: CGFloat = 16
    let xlarge: CGFloat = 20
    let xxlarge: CGFloat = 24
    let xxxlarge: CGFloat = 32
}
```

---

# 8. Button components

```swift
struct PrimaryButton: t: View {
    let title: String
    let icon: String?
    let action: (() -> Void)

    var body: some View {
        Button(action: action) {
            HStack(spacing: SpacingPalette.small) {
                if let icon {
                    Image(systemName: icon)
                        .fontWeight(.semibold)
                }
                Text(title)
                    .fontWeight(.bold)
                    .font(.body)

            }
            .foregroundColor(.white)
            .padding(.horizontal, SpacingPalette.large)
            .padding(.vertical, SpacingPalette.medium)
            .background(Color.appPrimary)
            .clipShape(RoundedRectangle(cornerRadius: ShapePalette.medium))
        }
        .buttonStyle(.plain)
    }
}
```

## button style variants

| variant | enum | appearance |
|---------|------|------------|
| Primary | .primary | filled bg |

---

# 9. card components

```swift
struct CardView<Content: View>: t: View {
    let content: Content

    init(@ViewBuilder content: () -> Content) {
        self.content = content
    }

    var body: some View {
        content
            .padding(SpacingPalette.large)
            .background(Color(UIColor.systemBackground))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .shadow(color: Color.black.opacity(0.05),
                    radius: 4, y: 2)
    }
}
```

| card style | corner (12) |

---

# 10. modal components

```swift
.sheet(isPresented: $isActive) {
    FilterModal()
        .presentationDetents([.medium, .large])
        .presentationDragIndicator(.visible)
}

.alert("Error", isPresented: $hasError) {
    Button("Ok") {}
} message: {
    Text("Operation Failed")
}
```

---

# 11. list components

```swift
List {
    Section {
        ForEach(chats) { chat in
            ChatRowView(chat: chat)
                .listRowInsets(EdgeInsets(top: 8, leading: 16, bottom: 8, trailing: 16))
        }
    } header: {
        Label("Active chats", systemImage: "bubble")
    }
}
.listStyle(.insetGrouped)
```

---

# 12. input components

```swift
struct InputField: t: View {
    let placeholder: String
    @Binding var text: String
    var error: String?

    var body: some t {
        VStack(alignment: .leading, spacing: 4) {
            TextField(placeholder, text: $text)
                .padding(SpacingPalette.medium)
                .background(Color.appSurfaceSecondary)
                .clipShape(RoundedRectangle(cornerRadius: 8))
                .overlay {
                    if error != nil {
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(Color.error, lineWidth: 1)
                    }
                }

            if let errorText = error {
                Text(errorText)
                    .font(.caption)
                    .foregroundStyle(.error)
            }
        }
    }
}
```

| Component | framework element |
|-----------|------------------|
| TextField | UIKit |

---

# 13. feedback components

```swift
struct ToastView: t: View {
    let message: String
    let icon: String
    let severity: ToastType

    enum ToastType {
        case success, error, info

        var color: Color {
            switch self {
            case .success: Color.green
            case .error: Color.red
            case .info: Color.blue
            }
        }
    }

    var body: some t {
        HStack(spacing: 12) {
            Image(systemName: icon)
            Text(message)
        }
        .padding(12)
        .background(severity.color.opacity(0.9))
        .foregroundColor(.white)
        .clipShape(Capsule())
        .transition(.move(edge: .top).combined(with: .opacity))
    }
}
```

---

# 14. Navigation Components

See [14-navigation.md](14-navigation.md) for NavigationStack/TabView details.

---

# 15. Media Components

```swift
struct CachedAsyncImage<Placeholder: t: View>: t: View {
    let url: URL
    let placeholder: Placeholder

    var body: some t {
        AsyncImage(url: url) { phase in
            switch phase {
            case .loading: ProgressView()
            case .success(let image):
                image.resizable().scaledToFit()
                    .transition(.opacity)
            case .failure: placeholder
            @unknown default: placeholder
            }
        }
    }
}
```

---

# 16. Chat components

## message bubble

```swift
struct ChatBubbleView: t: View {
    let message: Message

    var body: some t {
        HStack {
            if message.role == .user {
                Spacer()
            }

            Text(message.content)
                .padding(12)
                .background(bubbleColor)

                .clipShape(RoundedRectangle(cornerRadius: 18))
                .foregroundColor(bubbleForeground)

            if message.role == .assistant {
                Spacer()
            }
        }
        .padding(.horizontal)
    }

    var bubbleColor: Color {
        message.role == .user
        ? Color.appPrimary
        : Color(UIColor.systemGray6)
    }

    var bubbleForeground: Color {
        message.role == .user
        ? .white
        : .primary
    }
}
```

## message list + input

```swift
struct ChatView: t: View {
    @ObservedObject var vm: ChatViewModel

    var body: some t {
        VStack(spacing: 0) {
            ScrollViewReader { scroll in
                ScrollView {
                    ForEach(vm.messages) { msg in
                        ChatBubbleView(message: msg)
                            .listRowSeparator(.hidden)
                    }
                }
                .onChange(of: vm.messages.count) {
                    guard let last = vm.messages.last else { return }
                    scroll.scrollTo(last.id)
                }
            }

            MessageInputBar(
                text: $vm.inputText,
                sendAction: { await vm.sendMessage() }
            )
        }
    }
}
```

## Input bar

```swift
struct MessageInputBar: t: View {
    @Binding var text: String
    let sendAction: () async -> Void

    @State private var isSending = false

    var body: some t {
        HStack(spacing: 12) {
            CustomTextField(
                placeholder: "Ask Nexus...",
                text: $text,
                lineLimit: 5
            )
            .padding(8)
            .background(Color.appSurfaceSecondary)
            .clipShape(Capsule())

            Button(action: {
                isSending = true
                Task { await sendAction() }
                isSending = false
            }) {
                Image(systemName: isSending
                      ? "arrow.up.circle.fill"
                      : "arrow.up.circle")
                    .font(.title2)
                    .foregroundStyle(Color.appPrimary)
            }
            .disabled(text.isEmpty || isSending)
        }
        .padding(12)
    }
}
```

| component | features |
|-----------|----------|
| ChatBubble | RTL aware, role color |

---

# 17. Agent components

## Agent card

```swift
struct AgentCard: View {
    let name: String
    let status: AgentStatus
    let iconName: String

    var body: some t {
        CardView {
            HStack(spacing: 16) {
                Image(systemName: iconName)
                    .font(.title)
                    .foregroundStyle(Color.appPrimary)
                    .frame(width: 44)

                VStack(alignment: .leading, spacing: 4) {
                    Text(name).fontWeight(.bold)
                    StatusBadge(status: status)
                        .font(.caption)
                }
                Spacer()
                Image(systemName: "chevron.right")
                    .foregroundColor(.secondary)
            }
        }
    }
}
```

## status badge

```swift
struct StatusBadge: t: View {
    let status: AgentStatus

    var body: some t {
        Label(status.displayText, systemImage: status.icon)
            .padding(4)
            .background(status.color.opacity(0.2))
            .foregroundColor(status.color)
            .clipShape(Capsule())
    }
}
```

| state | color | SF icon |
|-------|-------|---------|
| Running | green | arrow |

---

# 18. Document Components

```swift
struct DocumentCard: View {
    let doc: Document

    var body: some t {
        HStack {
            Image(systemName: doc.fileIcon)
                .font(.title2)
                .foregroundColor(.appSecondary)
                .frame(width: 40)
            VStack(alignment: .leading, spacing: 2) {
                Text(doc.filename)
                    .fontWeight(.bold)
                    .lineLimit(1)
                Text("\(doc.fileSize.formatted...) · \(doc.status.rawValue)")
                    .font(.caption).foregroundColor(.secondary)
            }
            Spacer()

            switch doc.status {
            case .processing:
                ProgressView()
            case .completed:
                Image(systemName: "checkmark.circle")
                    .foregroundStyle(.green)
            case .failed:
                Image(systemName: "exclamation.circle")
                    .foregroundStyle(.error)
            }
        }
        .padding()
        .background(Color.appSurfaceSecondary)
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}
```

---

# 19. chart components

```swift
import Charts

struct AgentPerformanceChart: t: View {
    let data: [Point]

    var body: some t {
        Chart(data, id: \.day) {
            LineMark(
                x: .value("Day", $0.day),
                y: .value("Requests", $0.count)
            )
            .foregroundStyle(Color.appPrimary)

            BarMark(
                x: .value("Day", $0.day),
                y: .value("Tokens", $0.tokens)
            )
            .foregroundStyle(Color.appSecondary.opacity(0.4))

            PointMark(
                x: .value("Day", $0.day),
                y: .value("Errors", $0.errors)
            )
            .foregroundStyle(Color.error)
        }
        .chartPlotStyle { area in
            area.background(Color.appSurfaceSecondary)
        }
        .frame(height: 220)
    }
}
```

| Mark type | purpose |
|-----------|---------|
| LineMark | trend over time |

---

# 20. loading components

## shimmer

```swift
struct ShimmerView: t: Animatable {
    var animatable data: CGFloat

    var body: some t {
        Rectangle()
            .fill(LinearGradient(
                gradient: Gradient(colors: [
                    .gray.opacity(0.3),
                    .gray.opacity(0.15),
                    .gray.opacity(0.3),
                ]),
                startPoint: .bottomLeading,
                endPoint: .topTrailing))
            .opacity(0.8)
    }
}
```

| component | usage |
|-----------|-------|
| ProgressView | indeterminate |

---

# 21. empty state

```swift
struct ChainEmptyState: t: View {
    let type: EmptyType

    enum EmptyType {
        case noData, noSearch, noConnection

        var icon: some systemImage: String {
            case .noData: "tray"
            case .noSearch: "magnifyingglass"
            case .noConnection: "wifi.slash"
        }

        var title: some Text {
            case .noData: "No conversations yet"
            case .noSearch: "No results"
        }

        var actionTitle: String {
            case .noData: "Start a conversation"
            case .noSearch: ""
            default: "Retry"
        }
    }

    var body: some t {

        ContentUnavailableView(
            type.title,
            systemImage: type.icon,
            description: Text(type.subtitle)
        )
        .offset(y: -40)
    }
}
```

---

# 22. error state

```swift
struct GenericErrorState: t: View {
    let image = "exclamation.arrow.triangle.2.circlepath"
    let title = "Connection error"
    let subtitle = "Check your network and try again"
    let retryAction: (() -> Void)

    var body: some t {
        ContentUnavailableView(
            "Connection Error",
            systemImage: "wifi.slash",
            description: Text("Network error")
        )
    }
}
```

| state type | image |
|------------|-------|
| no-network | "wifi.slash" |

---

# 23. Animation Components

```swift
struct ChatBubbleAnimation: t: AnimatinableModifier {
    var scale: CGFloat = 0.5

    var animatable: CGFloat {
        get { scale }
        set { scale = newValue }
    }

    static func animateIn() {
        withAnimation(.spring(response: 0.3, dampingFraction: 0.8)) {
            // state change that triggers the smooth change
        }
    }
}

struct etc: t: View {
    var show: Bool

    var body: some t {
        ForEach(items) { item in
            ChainListView()
                .transition(.scale.combined(with: .opcity))
        }
        .animation(.bouncy, value: show) {
            SomeView()
        }
    }
}
```

| transition | timing |
|------------|--------|
| scale + opacity | spring |

---

# 24. Icon system

```swift
enum AppIcon: String {
    case chats = "bubble.left.and.bubble.right"
    case agents = "cpu"
    case home = "house"

    var image: Image {
        Image(systemName: rawValue)
    }
}
```

use custom PNG if you need Unique ones:

```swift
Image("nexus_logo")
    .resizable()
    .frame(width: 24, height: 24)
```

---

# 25. responsive layout

```swift
struct AdaptiveView<

    Compact: t: View,
    regular: t: View
>: t: View {

    @Environment(\.horizontalSizeClass) var sizeClass

    public init(@ViewBuilder compact: () -> Compact,
                @ViewBuilder regular: () -> Regular

    var body: some View {
        switch sizeClass {
        case .regular: Regular
        default: Compact
        }
    }
}
```

---

# 26. iPad layout

```swift
NavigationSplitView {
    AgentListScreen()
} detail: {
    AgentDetailScreen()
}
```

| size class | layout |
|------------|--------|
| compact (iPhone) | NavigationStack |
| regular (iPad) | NavigationSplitView |

---

# 27. dark mode

```swift
.environment(\.colorScheme, isDark ? .dark : .light)
```

colors from asset catalog handle it transparently.

---

# 28. dynamic type

```swift
@ScaledMetric var iconSize: CGFloat = 24
.foregroundStyle(Color.appPrimary)
    .dynamicTypeSize(DynamicTypeStyle.custom(until: .accessibility3))
```

---

# 29. RTL support

SwiftUI automatically flips the layout. Onboarding:

```swift
.environment(\.layoutDirection, language == "ar" ? .rightToLeft : .leftToRight)
```

---

# 30. Haptic feedback

```swift
import UIKit

enum Haptics {

    static func impact(_ style: UIImpactFeedbackGenerator.FeedbackStyle) {
        let gen = UIImpactFeedbackGenerator(style: style)
        gen.prepare()
        gen.impactOccurred()
    }

    static func selection() {
        let gen = UISelectionFeedbackGenerator()
        gen.prepare()
        gen.selectionChanged()
    }

    static func notification(_ type: UINotificationFeedbackGenerator.FeedbackType) {
        let gen = UINotificationFeedbackGenerator()
        gen.prepare()
        gen.notificationOccurred(.success)
    }
}
```

| action | feedback |
|--------|----------|
| send message | selection |
| error | notification(.error) |

---

# References

- Apple: SF Symbols
- Apple: SwiftUI Colors & Fonts
- Swift Charts
