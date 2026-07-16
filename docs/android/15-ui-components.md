# 15. UI Components

## Table of Contents

1. [Design System](#1-design-system)
2. [Theme Configuration](#2-theme-configuration)
3. [Dynamic Color](#3-dynamic-color)
4. [Color System](#4-color-system)
5. [Typography System](#5-typography-system)
6. [Shape System](#6-shape-system)
7. [Spacing System](#7-spacing-system)
8. [Button Components](#8-button-components)
9. [Card Components](#9-card-components)
10. [Modal Components](#10-modal-components)
11. [List Components](#11-list-components)
12. [Input Components](#12-input-components)
13. [Feedback Components](#13-feedback-components)
14. [Navigation Components](#14-navigation-components)
15. [Media Components](#15-media-components)
16. [Chat Components](#16-chat-components)
17. [Agent Components](#17-agent-components)
18. [Document Components](#18-document-components)
19. [Chart Components](#19-chart-components)
20. [Loading Components](#20-loading-components)
21. [Empty State Components](#21-empty-state-components)
22. [Error State Components](#22-error-state-components)
23. [Animation Components](#23-animation-components)
24. [Icon System](#24-icon-system)
25. [Responsive Layout](#25-responsive-layout)
26. [Foldable Device Support](#26-foldable-device-support)
27. [Tablet Layout](#27-tablet-layout)
28. [Dark Mode Support](#28-dark-mode-support)
29. [High Contrast Mode](#29-high-contrast-mode)
30. [Font Scaling](#30-font-scaling)
31. [RTL Support](#31-rtl-support)

---

## 1. Design System

### Design System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Design System Architecture                 │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                    Design Tokens                      │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │
│  │  │  Color   │  │ Typography│  │  Shape   │          │  │
│  │  │  Tokens  │  │  Tokens   │  │  Tokens  │          │  │
│  │  └──────────┘  └──────────┘  └──────────┘          │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │
│  │  │ Spacing  │  │  Elevation│  │ Duration │          │  │
│  │  │  Tokens  │  │  Tokens   │  │  Tokens  │          │  │
│  │  └──────────┘  └──────────┘  └──────────┘          │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                 │
│                           ▼                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  Theme (Material 3)                    │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │
│  │  │  Light   │  │   Dark   │  │ Dynamic  │          │  │
│  │  │  Theme   │  │  Theme   │  │  Color   │          │  │
│  │  └──────────┘  └──────────┘  └──────────┘          │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                 │
│                           ▼                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  Component Library                     │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │
│  │  │ Buttons  │  │  Cards   │  │  Inputs  │          │  │
│  │  └──────────┘  └──────────┘  └──────────┘          │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │
│  │  │  Modals  │  │  Lists   │  │ Feedback │          │  │
│  │  └──────────┘  └──────────┘  └──────────┘          │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                 │
│                           ▼                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                  Application Screens                    │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Dependencies

```kotlin
// build.gradle.kts
dependencies {
    // Material 3
    implementation("androidx.compose.material3:material3:1.2.1")

    // Material Icons Extended
    implementation("androidx.compose.material:material-icons-extended:1.6.0")

    // Compose Foundation
    implementation("androidx.compose.foundation:foundation:1.6.0")

    // Compose Animation
    implementation("androidx.compose.animation:animation:1.6.0")

    // Coil for image loading
    implementation("io.coil-kt:coil-compose:2.5.0")

    // Accompanist
    implementation("com.google.accompanist:accompanist-systemuicontroller:0.34.0")
    implementation("com.google.accompanist:accompanist-placeholder:0.34.0")
}
```

---

## 2. Theme Configuration

### Complete Theme Setup

```kotlin
// ─── Theme Definition ──────────────────────────────────────

@Composable
fun NexusTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    dynamicColor: Boolean = true,
    content: @Composable () -> Unit
) {
    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context)
            else dynamicLightColorScheme(context)
        }
        darkTheme -> darkNexusColorScheme()
        else -> lightNexusColorScheme()
    }

    val view = LocalView.current
    if (!view.isInEditMode) {
        SideEffect {
            val window = (view.context as Activity).window
            window.statusBarColor = colorScheme.surface.toArgb()
            WindowCompat.getInsetsController(window, view).isAppearanceLightStatusBars = !darkTheme
        }
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = NexusTypography,
        shapes = NexusShapes,
        content = content
    )
}

// ─── Light Color Scheme ────────────────────────────────────

fun lightNexusColorScheme(
    primary: Color = Color(0xFF1A73E8),
    onPrimary: Color = Color(0xFFFFFFFF),
    primaryContainer: Color = Color(0xFFD3E3FD),
    onPrimaryContainer: Color = Color(0xFF001D36),
    secondary: Color = Color(0xFF545F71),
    onSecondary: Color = Color(0xFFFFFFFF),
    secondaryContainer: Color = Color(0xFFD8E3F8),
    onSecondaryContainer: Color = Color(0xFF111C2B),
    tertiary: Color = Color(0xFF6D5677),
    onTertiary: Color = Color(0xFFFFFFFF),
    tertiaryContainer: Color = Color(0xFFF6D9FF),
    onTertiaryContainer: Color = Color(0xFF271430),
    error: Color = Color(0xFFBA1A1A),
    onError: Color = Color(0xFFFFFFFF),
    errorContainer: Color = Color(0xFFFFDAD6),
    onErrorContainer: Color = Color(0xFF410002),
    background: Color = Color(0xFFFDFBFF),
    onBackground: Color = Color(0xFF1A1C1E),
    surface: Color = Color(0xFFFDFBFF),
    onSurface: Color = Color(0xFF1A1C1E),
    surfaceVariant: Color = Color(0xFFE0E2EC),
    onSurfaceVariant: Color = Color(0xFF44474E),
    outline: Color = Color(0xFF74777F),
    outlineVariant: Color = Color(0xFFC4C6CF),
    inverseSurface: Color = Color(0xFF2F3033),
    inverseOnSurface: Color = Color(0xFFF1F0F4),
    inversePrimary: Color = Color(0xFFA4C9FA),
    surfaceTint: Color = Color(0xFF1A73E8),
    surfaceDim: Color = Color(0xFFDADADE),
    surfaceBright: Color = Color(0xFFFDFBFF),
    surfaceContainerLowest: Color = Color(0xFFFFFFFF),
    surfaceContainerLow: Color = Color(0xFFF8F6FA),
    surfaceContainer: Color = Color(0xFFF2F0F4),
    surfaceContainerHigh: Color = Color(0xFFECEAEF),
    surfaceContainerHighest: Color = Color(0xFFE6E4E9)
)

// ─── Dark Color Scheme ─────────────────────────────────────

fun darkNexusColorScheme(
    primary: Color = Color(0xFFA4C9FA),
    onPrimary: Color = Color(0xFF003258),
    primaryContainer: Color = Color(0xFF004A7C),
    onPrimaryContainer: Color = Color(0xFFD3E3FD),
    secondary: Color = Color(0xFFBCC7DB),
    onSecondary: Color = Color(0xFF263141),
    secondaryContainer: Color = Color(0xFF3C4758),
    onSecondaryContainer: Color = Color(0xFFD8E3F8),
    tertiary: Color = Color(0xFFD9BDE3),
    onTertiary: Color = Color(0xFF3D2946),
    tertiaryContainer: Color = Color(0xFF543F5E),
    onTertiaryContainer: Color = Color(0xFFF6D9FF),
    error: Color = Color(0xFFFFB4AB),
    onError: Color = Color(0xFF690005),
    errorContainer: Color = Color(0xFF93000A),
    onErrorContainer: Color = Color(0xFFFFDAD6),
    background: Color = Color(0xFF1A1C1E),
    onBackground: Color = Color(0xFFE2E2E6),
    surface: Color = Color(0xFF1A1C1E),
    onSurface: Color = Color(0xFFE2E2E6),
    surfaceVariant: Color = Color(0xFF44474E),
    onSurfaceVariant: Color = Color(0xFFC4C6CF),
    outline: Color = Color(0xFF8E9099),
    outlineVariant: Color = Color(0xFF44474E),
    inverseSurface: Color = Color(0xFFE2E2E6),
    inverseOnSurface: Color = Color(0xFF2F3033),
    inversePrimary: Color = Color(0xFF1A73E8),
    surfaceTint: Color = Color(0xFFA4C9FA),
    surfaceDim: Color = Color(0xFF1A1C1E),
    surfaceBright: Color = Color(0xFF3F4043),
    surfaceContainerLowest: Color = Color(0xFF141416),
    surfaceContainerLow: Color = Color(0xFF222224),
    surfaceContainer: Color = Color(0xFF262629),
    surfaceContainerHigh: Color = Color(0xFF313134),
    surfaceContainerHighest: Color = Color(0xFF3C3C3F)
)
```

---

## 3. Dynamic Color

### Material You Dynamic Color

```kotlin
@Composable
fun NexusDynamicTheme(
    useDynamicColor: Boolean = true,
    content: @Composable () -> Unit
) {
    val context = LocalContext.current

    val colorScheme = when {
        useDynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            if (isSystemInDarkTheme()) {
                dynamicDarkColorScheme(context)
            } else {
                dynamicLightColorScheme(context)
            }
        }
        isSystemInDarkTheme() -> darkNexusColorScheme()
        else -> lightNexusColorScheme()
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = NexusTypography,
        shapes = NexusShapes,
        content = content
    )
}

// ─── Color Palette Extraction ──────────────────────────────

@Composable
fun DynamicColorInfo() {
    val colorScheme = MaterialTheme.colorScheme

    Column {
        ColorSwatch("Primary", colorScheme.primary)
        ColorSwatch("On Primary", colorScheme.onPrimary)
        ColorSwatch("Primary Container", colorScheme.primaryContainer)
        ColorSwatch("Secondary", colorScheme.secondary)
        ColorSwatch("Tertiary", colorScheme.tertiary)
        ColorSwatch("Error", colorScheme.error)
        ColorSwatch("Background", colorScheme.background)
        ColorSwatch("Surface", colorScheme.surface)
    }
}

@Composable
fun ColorSwatch(name: String, color: Color) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(40.dp)
                .clip(RoundedCornerShape(8.dp))
                .background(color)
        )
        Spacer(modifier = Modifier.width(12.dp))
        Column {
            Text(text = name, style = MaterialTheme.typography.labelMedium)
            Text(
                text = "#${color.value.toString(16).take(8).uppercase()}",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}
```

### Dynamic Color Fallback

```
┌─────────────────────────────────────────────────────────┐
│             Dynamic Color Decision Flow                   │
│                                                         │
│  App Launch                                             │
│       │                                                 │
│       ▼                                                 │
│  ┌──────────────────┐                                   │
│  │ Android 12+?     │                                   │
│  └──────┬───────────┘                                   │
│     Yes │  No                                           │
│     │   │                                              │
│     ▼   ▼                                              │
│  ┌──────┐ ┌──────────────┐                             │
│  │Dynmc │ │ Use static   │                             │
│  │Color │ │ color scheme │                             │
│  │      │ │              │                             │
│  └──┬───┘ └──────────────┘                             │
│     │                                                   │
│     ▼                                                   │
│  ┌──────────────────┐                                   │
│  │ User preference   │                                 │
│  │ "Dynamic Color"   │                                 │
│  └──────┬───────────┘                                   │
│   On    │  Off                                          │
│   │     │                                              │
│   ▼     ▼                                              │
│  Use   Use fallback                                    │
│  dynmc  scheme                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 4. Color System

### Color Tokens Table

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `primary` | #1A73E8 | #A4C9FA | Primary actions, links |
| `onPrimary` | #FFFFFF | #003258 | Text on primary |
| `primaryContainer` | #D3E3FD | #004A7C | Chips, badges |
| `onPrimaryContainer` | #001D36 | #D3E3FD | Text on primary container |
| `secondary` | #545F71 | #BCC7DB | Secondary actions |
| `secondaryContainer` | #D8E3F8 | #3C4758 | Secondary containers |
| `tertiary` | #6D5677 | #D9BDE3 | Accent, highlights |
| `tertiaryContainer` | #F6D9FF | #543F5E | Tertiary containers |
| `error` | #BA1A1A | #FFB4AB | Error states |
| `errorContainer` | #FFDAD6 | #93000A | Error containers |
| `background` | #FDFBFF | #1A1C1E | Screen background |
| `surface` | #FDFBFF | #1A1C1E | Card/sheet background |
| `surfaceVariant` | #E0E2EC | #44474E | Subtle containers |
| `outline` | #74777F | #8E9099 | Borders, dividers |

### Custom App Colors

```kotlin
object NexusColors {
    // Chat-specific colors
    val userMessageBubble = Color(0xFF1A73E8)
    val userMessageText = Color(0xFFFFFFFF)
    val assistantMessageBubble = Color(0xFFF1F3F4)
    val assistantMessageText = Color(0xFF202124)

    // Status colors
    val online = Color(0xFF34A853)
    val offline = Color(0xFF9AA0A6)
    val away = Color(0xFFFBBC04)
    val busy = Color(0xFFEA4335)

    // Sync status
    val synced = Color(0xFF34A853)
    val pending = Color(0xFFFBBC04)
    val failed = Color(0xFFEA4335)
    val conflict = Color(0xFF9C27B0)

    // Processing status
    val processing = Color(0xFF1A73E8)
    val completed = Color(0xFF34A853)
    val queued = Color(0xFF9AA0A6)

    // Sentiment colors (for document analysis)
    val positive = Color(0xFF34A853)
    val negative = Color(0xFFEA4335)
    val neutral = Color(0xFF9AA0A6)
}
```

---

## 5. Typography System

### Typography Scale

```kotlin
val NexusTypography = Typography(
    // Display
    displayLarge = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Normal,
        fontSize = 57.sp,
        lineHeight = 64.sp,
        letterSpacing = (-0.25).sp
    ),
    displayMedium = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Normal,
        fontSize = 45.sp,
        lineHeight = 52.sp,
        letterSpacing = 0.sp
    ),
    displaySmall = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Normal,
        fontSize = 36.sp,
        lineHeight = 44.sp,
        letterSpacing = 0.sp
    ),

    // Headline
    headlineLarge = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.SemiBold,
        fontSize = 32.sp,
        lineHeight = 40.sp,
        letterSpacing = 0.sp
    ),
    headlineMedium = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.SemiBold,
        fontSize = 28.sp,
        lineHeight = 36.sp,
        letterSpacing = 0.sp
    ),
    headlineSmall = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.SemiBold,
        fontSize = 24.sp,
        lineHeight = 32.sp,
        letterSpacing = 0.sp
    ),

    // Title
    titleLarge = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Medium,
        fontSize = 22.sp,
        lineHeight = 28.sp,
        letterSpacing = 0.sp
    ),
    titleMedium = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Medium,
        fontSize = 16.sp,
        lineHeight = 24.sp,
        letterSpacing = 0.15.sp
    ),
    titleSmall = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Medium,
        fontSize = 14.sp,
        lineHeight = 20.sp,
        letterSpacing = 0.1.sp
    ),

    // Body
    bodyLarge = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Normal,
        fontSize = 16.sp,
        lineHeight = 24.sp,
        letterSpacing = 0.5.sp
    ),
    bodyMedium = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Normal,
        fontSize = 14.sp,
        lineHeight = 20.sp,
        letterSpacing = 0.25.sp
    ),
    bodySmall = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Normal,
        fontSize = 12.sp,
        lineHeight = 16.sp,
        letterSpacing = 0.4.sp
    ),

    // Label
    labelLarge = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Medium,
        fontSize = 14.sp,
        lineHeight = 20.sp,
        letterSpacing = 0.1.sp
    ),
    labelMedium = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Medium,
        fontSize = 12.sp,
        lineHeight = 16.sp,
        letterSpacing = 0.5.sp
    ),
    labelSmall = TextStyle(
        fontFamily = NexusFontFamily.Default,
        fontWeight = FontWeight.Medium,
        fontSize = 11.sp,
        lineHeight = 16.sp,
        letterSpacing = 0.5.sp
    )
)

// ─── Font Families ─────────────────────────────────────────

object NexusFontFamily {
    val Default = FontFamily.Default
    val Monospace = FontFamily.Monospace
    val Serif = FontFamily.Serif

    // Custom fonts (if needed)
    // val Custom = FontFamily(Font(R.font.custom_font))
}
```

### Typography Usage Table

| Style | Size | Weight | Use For |
|-------|------|--------|---------|
| displayLarge | 57sp | Normal | Splash screen title |
| displayMedium | 45sp | Normal | Onboarding |
| displaySmall | 36sp | Normal | Section headers |
| headlineLarge | 32sp | SemiBold | Screen titles |
| headlineMedium | 28sp | SemiBold | Dialog titles |
| headlineSmall | 24sp | SemiBold | Card titles |
| titleLarge | 22sp | Medium | App bar titles |
| titleMedium | 16sp | Medium | List item titles |
| titleSmall | 14sp | Medium | Subtitles |
| bodyLarge | 16sp | Normal | Main content |
| bodyMedium | 14sp | Normal | Secondary content |
| bodySmall | 12sp | Normal | Captions |
| labelLarge | 14sp | Medium | Button text |
| labelMedium | 12sp | Medium | Badge text |
| labelSmall | 11sp | Medium | Timestamps |

---

## 6. Shape System

### Shape Definitions

```kotlin
val NexusShapes = Shapes(
    none = RoundedCornerShape(0.dp),
    extraSmall = RoundedCornerShape(4.dp),
    small = RoundedCornerShape(8.dp),
    medium = RoundedCornerShape(12.dp),
    large = RoundedCornerShape(16.dp),
    extraLarge = RoundedCornerShape(28.dp),
    full = CircleShape
)

// ─── Custom Shapes ─────────────────────────────────────────

object NexusCustomShapes {
    val chatBubbleUser = RoundedCornerShape(
        topStart = 16.dp,
        topEnd = 16.dp,
        bottomStart = 16.dp,
        bottomEnd = 4.dp
    )

    val chatBubbleAssistant = RoundedCornerShape(
        topStart = 16.dp,
        topEnd = 16.dp,
        bottomStart = 4.dp,
        bottomEnd = 16.dp
    )

    val cardRounded = RoundedCornerShape(16.dp)
    val cardSharp = RoundedCornerShape(4.dp)
    val bottomSheet = RoundedCornerShape(topStart = 28.dp, topEnd = 28.dp)
    val dialog = RoundedCornerShape(28.dp)
    val button = RoundedCornerShape(20.dp)
    val chip = RoundedCornerShape(8.dp)
    val avatar = CircleShape
    val notification = RoundedCornerShape(12.dp)
}
```

### Shape Usage Guide

| Shape | Radius | Use For |
|-------|--------|---------|
| none | 0dp | Dividers, full-width elements |
| extraSmall | 4dp | Subtle rounding |
| small | 8dp | Chips, small buttons, tags |
| medium | 12dp | Cards, text fields |
| large | 16dp | Bottom sheets, dialogs |
| extraLarge | 28dp | Bottom sheet drag handle |
| full | Circle | Avatars, FABs, icons |

---

## 7. Spacing System

### Spacing Scale

```kotlin
object NexusSpacing {
    val xxxs = 2.dp
    val xxs = 4.dp
    val xs = 8.dp
    val sm = 12.dp
    val md = 16.dp
    val lg = 20.dp
    val xl = 24.dp
    val xxl = 32.dp
    val xxxl = 40.dp
    val xxxxl = 48.dp
    val xxxxxl = 64.dp
}

// ─── Padding Extensions ────────────────────────────────────

fun Modifier.nexusPadding(
    horizontal: Dp = NexusSpacing.md,
    vertical: Dp = NexusSpacing.md
) = padding(horizontal = horizontal, vertical = vertical)

fun Modifier.nexusPaddingAll() = padding(NexusSpacing.md)

fun Modifier.nexusPaddingHorizontal() = padding(horizontal = NexusSpacing.md)

fun Modifier.nexusPaddingVertical() = padding(vertical = NexusSpacing.md)

// ─── Spacing Modifier ──────────────────────────────────────

@Composable
fun Modifier.nexusSpacing(
    after: Dp = NexusSpacing.md
) = this.then(Modifier.padding(bottom = after))

// ─── Gap Modifier ──────────────────────────────────────────

@Composable
fun RowScope.nexusGap(
    gap: Dp = NexusSpacing.xs
) = this.then(Modifier.weight(1f))
```

### Spacing Reference

```
┌─────────────────────────────────────────────────────────┐
│                  Spacing Scale                           │
│                                                         │
│  xxxxs: 64dp  ████████████████████████                  │
│  xxxl:  40dp  ████████████████                          │
│  xxl:   32dp  ████████████                              │
│  xl:    24dp  █████████                                 │
│  lg:    20dp  ████████                                  │
│  md:    16dp  ██████                                    │
│  sm:    12dp  ████                                      │
│  xs:     8dp  ███                                       │
│  xxs:    4dp  ██                                        │
│  xxxs:   2dp  █                                         │
│                                                         │
│  Usage:                                                 │
│  • Screen padding: md (16dp)                            │
│  • Card padding: md-lg (16-20dp)                        │
│  • List item padding: md (16dp)                         │
│  • Between items: xs-sm (8-12dp)                        │
│  • Between sections: xl-xxl (24-32dp)                   │
│  • Button internal padding: xs-sm (8-12dp)              │
└─────────────────────────────────────────────────────────┘
```

---

## 8. Button Components

### Button Variants

```kotlin
// ─── Filled Button ─────────────────────────────────────────

@Composable
fun NexusFilledButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    isLoading: Boolean = false
) {
    Button(
        onClick = onClick,
        modifier = modifier.height(48.dp),
        enabled = enabled && !isLoading,
        shape = NexusCustomShapes.button
    ) {
        if (isLoading) {
            CircularProgressIndicator(
                modifier = Modifier.size(20.dp),
                strokeWidth = 2.dp,
                color = MaterialTheme.colorScheme.onPrimary
            )
        } else {
            Text(text = text)
        }
    }
}

// ─── Outlined Button ───────────────────────────────────────

@Composable
fun NexusOutlinedButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    icon: ImageVector? = null
) {
    OutlinedButton(
        onClick = onClick,
        modifier = modifier.height(48.dp),
        enabled = enabled,
        shape = NexusCustomShapes.button
    ) {
        if (icon != null) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                modifier = Modifier.size(18.dp)
            )
            Spacer(modifier = Modifier.width(8.dp))
        }
        Text(text = text)
    }
}

// ─── Text Button ───────────────────────────────────────────

@Composable
fun NexusTextButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true
) {
    TextButton(
        onClick = onClick,
        modifier = modifier,
        enabled = enabled
    ) {
        Text(text = text)
    }
}

// ─── Icon Button ───────────────────────────────────────────

@Composable
fun NexusIconButton(
    icon: ImageVector,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    contentDescription: String? = null,
    enabled: Boolean = true,
    tint: Color = MaterialTheme.colorScheme.onSurface
) {
    IconButton(
        onClick = onClick,
        modifier = modifier.size(48.dp),
        enabled = enabled
    ) {
        Icon(
            imageVector = icon,
            contentDescription = contentDescription,
            tint = tint
        )
    }
}

// ─── Tonal Button ──────────────────────────────────────────

@Composable
fun NexusTonalButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    icon: ImageVector? = null
) {
    FilledTonalButton(
        onClick = onClick,
        modifier = modifier.height(48.dp),
        enabled = enabled,
        shape = NexusCustomShapes.button
    ) {
        if (icon != null) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                modifier = Modifier.size(18.dp)
            )
            Spacer(modifier = Modifier.width(8.dp))
        }
        Text(text = text)
    }
}

// ─── Extended FAB ──────────────────────────────────────────

@Composable
fun NexusExtendedFab(
    text: String,
    icon: ImageVector,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    ExtendedFloatingActionButton(
        onClick = onClick,
        modifier = modifier,
        icon = { Icon(icon, contentDescription = null) },
        text = { Text(text) },
        containerColor = MaterialTheme.colorScheme.primaryContainer,
        contentColor = MaterialTheme.colorScheme.onPrimaryContainer
    )
}
```

### Button Reference

| Button | Height | Padding | Use For |
|--------|--------|---------|---------|
| Filled | 48dp | 24dp horizontal | Primary actions |
| Outlined | 48dp | 24dp horizontal | Secondary actions |
| Text | - | 12dp horizontal | Tertiary actions |
| Icon | 48dp | - | Icon-only actions |
| Tonal | 48dp | 24dp horizontal | Emphasized secondary |
| FAB | 56dp | - | Floating actions |

---

## 9. Card Components

### Card Variants

```kotlin
// ─── Elevated Card ─────────────────────────────────────────

@Composable
fun NexusElevatedCard(
    modifier: Modifier = Modifier,
    onClick: (() -> Unit)? = null,
    content: @Composable ColumnScope.() -> Unit
) {
    if (onClick != null) {
        Card(
            onClick = onClick,
            modifier = modifier.fillMaxWidth(),
            shape = NexusCustomShapes.cardRounded,
            colors = CardDefaults.elevatedCardColors(
                containerColor = MaterialTheme.colorScheme.surface
            ),
            elevation = CardDefaults.cardElevation(
                defaultElevation = 2.dp,
                pressedElevation = 4.dp
            ),
            content = content
        )
    } else {
        Card(
            modifier = modifier.fillMaxWidth(),
            shape = NexusCustomShapes.cardRounded,
            colors = CardDefaults.elevatedCardColors(
                containerColor = MaterialTheme.colorScheme.surface
            ),
            elevation = CardDefaults.cardElevation(
                defaultElevation = 2.dp
            ),
            content = content
        )
    }
}

// ─── Filled Card ───────────────────────────────────────────

@Composable
fun NexusFilledCard(
    modifier: Modifier = Modifier,
    onClick: (() -> Unit)? = null,
    content: @Composable ColumnScope.() -> Unit
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        onClick = onClick ?: {},
        shape = NexusCustomShapes.cardRounded,
        colors = CardDefaults.filledCardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        ),
        content = content
    )
}

// ─── Outlined Card ─────────────────────────────────────────

@Composable
fun NexusOutlinedCard(
    modifier: Modifier = Modifier,
    onClick: (() -> Unit)? = null,
    content: @Composable ColumnScope.() -> Unit
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        onClick = onClick ?: {},
        shape = NexusCustomShapes.cardRounded,
        colors = CardDefaults.outlinedCardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        border = CardDefaults.outlinedCardBorder(),
        content = content
    )
}

// ─── Content Card (for messages, documents) ────────────────

@Composable
fun NexusContentCard(
    title: String,
    subtitle: String? = null,
    leading: @Composable (() -> Unit)? = null,
    trailing: @Composable (() -> Unit)? = null,
    onClick: (() -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        onClick = onClick ?: {},
        shape = NexusCustomShapes.cardRounded,
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            if (leading != null) {
                leading()
                Spacer(modifier = Modifier.width(16.dp))
            }

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                if (subtitle != null) {
                    Text(
                        text = subtitle,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis
                    )
                }
            }

            if (trailing != null) {
                Spacer(modifier = Modifier.width(8.dp))
                trailing()
            }
        }
    }
}
```

---

## 10. Modal Components

### Dialog

```kotlin
@Composable
fun NexusAlertDialog(
    title: String,
    message: String,
    confirmText: String = "OK",
    dismissText: String? = "Cancel",
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
    icon: ImageVector? = null
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        icon = icon?.let {
            {
                Icon(
                    it,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.primary
                )
            }
        },
        title = { Text(text = title) },
        text = { Text(text = message) },
        confirmButton = {
            Button(onClick = onConfirm) {
                Text(text = confirmText)
            }
        },
        dismissButton = dismissText?.let {
            {
                TextButton(onClick = onDismiss) {
                    Text(text = it)
                }
            }
        }
    )
}

@Composable
fun NexusConfirmDialog(
    title: String,
    message: String,
    confirmText: String = "Confirm",
    dismissText: String = "Cancel",
    isDestructive: Boolean = false,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(text = title) },
        text = { Text(text = message) },
        confirmButton = {
            Button(
                onClick = onConfirm,
                colors = if (isDestructive) {
                    ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.error
                    )
                } else ButtonDefaults.buttonColors()
            ) {
                Text(text = confirmText)
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text(text = dismissText)
            }
        }
    )
}
```

### Bottom Sheet

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NexusBottomSheet(
    onDismiss: () -> Unit,
    title: String? = null,
    content: @Composable ColumnScope.() -> Unit
) {
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        dragHandle = { BottomSheetDefaults.DragHandle() },
        shape = NexusCustomShapes.bottomSheet,
        containerColor = MaterialTheme.colorScheme.surface
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(bottom = 32.dp)
        ) {
            if (title != null) {
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleLarge,
                    modifier = Modifier.padding(bottom = 16.dp)
                )
            }
            content()
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NexusListBottomSheet(
    onDismiss: () -> Unit,
    title: String? = null,
    items: List<BottomSheetItem>,
    onItemClick: (BottomSheetItem) -> Unit
) {
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        dragHandle = { BottomSheetDefaults.DragHandle() },
        shape = NexusCustomShapes.bottomSheet
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 32.dp)
        ) {
            if (title != null) {
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleLarge,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
                )
            }

            items.forEach { item ->
                ListItem(
                    headlineContent = { Text(item.title) },
                    supportingContent = item.subtitle?.let { { Text(it) } },
                    leadingContent = item.icon?.let {
                        { Icon(it, contentDescription = null) }
                    },
                    modifier = Modifier.clickable {
                        onItemClick(item)
                        onDismiss()
                    }
                )
            }
        }
    }
}

data class BottomSheetItem(
    val id: String,
    val title: String,
    val subtitle: String? = null,
    val icon: ImageVector? = null,
    val isDestructive: Boolean = false
)
```

### Full Screen Dialog

```kotlin
@Composable
fun NexusFullScreenDialog(
    onDismiss: () -> Unit,
    title: String,
    actions: @Composable RowScope.() -> Unit = {},
    content: @Composable () -> Unit
) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(
            dismissOnBackPress = true,
            dismissOnClickOutside = false,
            usePlatformDefaultWidth = false
        )
    ) {
        Surface(
            modifier = Modifier
                .fillMaxSize()
                .statusBarsPadding(),
            shape = RoundedCornerShape(bottomStart = 28.dp, bottomEnd = 28.dp)
        ) {
            Scaffold(
                topBar = {
                    TopAppBar(
                        title = { Text(title) },
                        navigationIcon = {
                            IconButton(onClick = onDismiss) {
                                Icon(Icons.Filled.Close, contentDescription = "Close")
                            }
                        },
                        actions = { actions() }
                    )
                }
            ) { innerPadding ->
                Box(modifier = Modifier.padding(innerPadding)) {
                    content()
                }
            }
        }
    }
}
```

---

## 11. List Components

### List Item Variants

```kotlin
// ─── Standard List Item ────────────────────────────────────

@Composable
fun NexusListItem(
    title: String,
    subtitle: String? = null,
    leading: @Composable (() -> Unit)? = null,
    trailing: @Composable (() -> Unit)? = null,
    onClick: (() -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    ListItem(
        headlineContent = {
            Text(
                text = title,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
        },
        supportingContent = subtitle?.let {
            {
                Text(
                    text = it,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        },
        leadingContent = leading,
        trailingContent = trailing,
        modifier = modifier.then(
            if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier
        )
    )
}

// ─── Section Header ────────────────────────────────────────

@Composable
fun NexusSectionHeader(
    title: String,
    actionText: String? = null,
    onActionClick: (() -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleSmall,
            color = MaterialTheme.colorScheme.onSurface
        )

        if (actionText != null && onActionClick != null) {
            TextButton(
                onClick = onActionClick,
                contentPadding = PaddingValues(0.dp)
            ) {
                Text(text = actionText)
            }
        }
    }
}

// ─── Divider ───────────────────────────────────────────────

@Composable
fun NexusDivider(
    modifier: Modifier = Modifier,
    color: Color = MaterialTheme.colorScheme.outlineVariant,
    thickness: Dp = 0.5.dp
) {
    HorizontalDivider(
        modifier = modifier,
        color = color,
        thickness = thickness
    )
}

// ─── Sticky Header ─────────────────────────────────────────

@Composable
fun NexusStickyHeader(
    text: String,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier.fillMaxWidth(),
        color = MaterialTheme.colorScheme.surfaceContainer
    ) {
        Text(
            text = text,
            style = MaterialTheme.typography.labelMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
        )
    }
}
```

### List Patterns

```
┌─────────────────────────────────────────────────────────┐
│                  List Layout Pattern                     │
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │ Section Header: "Recent Conversations"    [View]  │  │
│  ├───────────────────────────────────────────────────┤  │
│  │ ┌───┐ Conversation Title                  ... │  │  │
│  │ │AVT│ Last message preview text               │  │  │
│  │ └───┘ 2 min ago                               │  │  │
│  ├───────────────────────────────────────────────────┤  │
│  │ ┌───┐ Another Conversation               ... │  │  │
│  │ │AVT│ Different message preview              │  │  │
│  │ └───┘ 1 hour ago                              │  │  │
│  ├───────────────────────────────────────────────────┤  │
│  │ ┌───┐ Third Conversation                 ... │  │  │
│  │ │AVT│ Yet another message preview            │  │  │
│  │ └───┘ Yesterday                               │  │  │
│  ├───────────────────────────────────────────────────┤  │
│  │ Section Header: "Documents"               [View] │  │
│  ├───────────────────────────────────────────────────┤  │
│  │ ┌───┐ document.pdf                    2.4 MB │  │  │
│  │ │📄│ Processed • 12 chunks                    │  │  │
│  │ └───┘ Uploaded 2 days ago                      │  │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 12. Input Components

### Input Variants

```kotlin
// ─── Text Field ────────────────────────────────────────────

@Composable
fun NexusTextField(
    value: String,
    onValueChange: (String) -> Unit,
    label: String,
    modifier: Modifier = Modifier,
    placeholder: String? = null,
    leadingIcon: ImageVector? = null,
    trailingIcon: @Composable (() -> Unit)? = null,
    error: String? = null,
    enabled: Boolean = true,
    readOnly: Boolean = false,
    singleLine: Boolean = true,
    maxLines: Int = 1
) {
    OutlinedTextField(
        value = value,
        onValueChange = onValueChange,
        label = { Text(label) },
        modifier = modifier.fillMaxWidth(),
        placeholder = placeholder?.let { { Text(it) } },
        leadingIcon = leadingIcon?.let {
            { Icon(it, contentDescription = null) }
        },
        trailingIcon = trailingIcon,
        isError = error != null,
        supportingText = error?.let {
            { Text(it, color = MaterialTheme.colorScheme.error) }
        },
        enabled = enabled,
        readOnly = readOnly,
        singleLine = singleLine,
        maxLines = maxLines,
        shape = NexusCustomShapes.small
    )
}

// ─── Search Bar ────────────────────────────────────────────

@Composable
fun NexusSearchBar(
    query: String,
    onQueryChange: (String) -> Unit,
    onSearch: () -> Unit,
    onClear: () -> Unit,
    placeholder: String = "Search...",
    modifier: Modifier = Modifier
) {
    OutlinedTextField(
        value = query,
        onValueChange = onQueryChange,
        modifier = modifier
            .fillMaxWidth()
            .height(52.dp),
        placeholder = { Text(placeholder) },
        leadingIcon = {
            Icon(
                Icons.Filled.Search,
                contentDescription = "Search",
                tint = MaterialTheme.colorScheme.onSurfaceVariant
            )
        },
        trailingIcon = {
            if (query.isNotEmpty()) {
                IconButton(onClick = onClear, modifier = Modifier.size(32.dp)) {
                    Icon(
                        Icons.Filled.Clear,
                        contentDescription = "Clear",
                        modifier = Modifier.size(18.dp)
                    )
                }
            }
        },
        keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
        keyboardActions = KeyboardActions(onSearch = { onSearch() }),
        singleLine = true,
        shape = NexusCustomShapes.extraLarge,
        colors = OutlinedTextFieldDefaults.colors(
            unfocusedBorderColor = MaterialTheme.colorScheme.surfaceVariant,
            focusedBorderColor = MaterialTheme.colorScheme.primary
        )
    )
}

// ─── Dropdown ──────────────────────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NexusDropdown(
    value: String,
    onValueChange: (String) -> Unit,
    options: List<String>,
    label: String,
    modifier: Modifier = Modifier,
    enabled: Boolean = true
) {
    var expanded by remember { mutableStateOf(false) }

    ExposedDropdownMenuBox(
        expanded = expanded,
        onExpandedChange = { expanded = !expanded },
        modifier = modifier
    ) {
        OutlinedTextField(
            value = value,
            onValueChange = onValueChange,
            label = { Text(label) },
            readOnly = true,
            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) },
            modifier = Modifier
                .fillMaxWidth()
                .menuAnchor(),
            enabled = enabled,
            shape = NexusCustomShapes.small
        )

        ExposedDropdownMenu(
            expanded = expanded,
            onDismissRequest = { expanded = false }
        ) {
            options.forEach { option ->
                DropdownMenuItem(
                    text = { Text(option) },
                    onClick = {
                        onValueChange(option)
                        expanded = false
                    }
                )
            }
        }
    }
}

// ─── Checkbox, Radio, Switch ───────────────────────────────

@Composable
fun NexusCheckbox(
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
    label: String,
    modifier: Modifier = Modifier,
    enabled: Boolean = true
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .clickable(enabled = enabled) { onCheckedChange(!checked) }
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Checkbox(
            checked = checked,
            onCheckedChange = onCheckedChange,
            enabled = enabled
        )
        Spacer(modifier = Modifier.width(8.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.bodyLarge,
            color = if (enabled) {
                MaterialTheme.colorScheme.onSurface
            } else {
                MaterialTheme.colorScheme.onSurface.copy(alpha = 0.38f)
            }
        )
    }
}

@Composable
fun NexusRadio(
    selected: Boolean,
    onClick: () -> Unit,
    label: String,
    modifier: Modifier = Modifier,
    enabled: Boolean = true
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .clickable(enabled = enabled) { onClick() }
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        RadioButton(
            selected = selected,
            onClick = onClick,
            enabled = enabled
        )
        Spacer(modifier = Modifier.width(8.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.bodyLarge,
            color = if (enabled) {
                MaterialTheme.colorScheme.onSurface
            } else {
                MaterialTheme.colorScheme.onSurface.copy(alpha = 0.38f)
            }
        )
    }
}

@Composable
fun NexusSwitch(
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
    label: String,
    subtitle: String? = null,
    modifier: Modifier = Modifier,
    enabled: Boolean = true
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .clickable(enabled = enabled) { onCheckedChange(!checked) }
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = label,
                style = MaterialTheme.typography.bodyLarge,
                color = if (enabled) {
                    MaterialTheme.colorScheme.onSurface
                } else {
                    MaterialTheme.colorScheme.onSurface.copy(alpha = 0.38f)
                }
            )
            if (subtitle != null) {
                Text(
                    text = subtitle,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
        Switch(
            checked = checked,
            onCheckedChange = onCheckedChange,
            enabled = enabled
        )
    }
}
```

---

## 13. Feedback Components

### Snackbar and Toast

```kotlin
@Composable
fun NexusSnackbarHost(
    snackbarHostState: SnackbarHostState,
    modifier: Modifier = Modifier
) {
    SnackbarHost(
        hostState = snackbarHostState,
        modifier = modifier,
        snackbar = { data ->
            Snackbar(
                snackbarData = data,
                shape = NexusCustomShapes.small,
                containerColor = MaterialTheme.colorScheme.inverseSurface,
                contentColor = MaterialTheme.colorScheme.inverseOnSurface,
                actionColor = MaterialTheme.colorScheme.inversePrimary
            )
        }
    )
}

@Composable
fun NexusToast(
    message: String,
    duration: Int = Toast.LENGTH_SHORT
) {
    val context = LocalContext.current
    LaunchedEffect(message) {
        Toast.makeText(context, message, duration).show()
    }
}
```

### Progress Indicators

```kotlin
@Composable
fun NexusCircularProgress(
    modifier: Modifier = Modifier,
    message: String? = null
) {
    Column(
        modifier = modifier,
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        CircularProgressIndicator(
            modifier = Modifier.size(48.dp),
            color = MaterialTheme.colorScheme.primary,
            strokeWidth = 4.dp
        )

        if (message != null) {
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = message,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
fun NexusLinearProgress(
    progress: Float,
    modifier: Modifier = Modifier,
    label: String? = null
) {
    Column(modifier = modifier) {
        if (label != null) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = label,
                    style = MaterialTheme.typography.labelMedium
                )
                Text(
                    text = "${(progress * 100).toInt()}%",
                    style = MaterialTheme.typography.labelMedium
                )
            }
            Spacer(modifier = Modifier.height(4.dp))
        }

        LinearProgressIndicator(
            progress = { progress },
            modifier = Modifier
                .fillMaxWidth()
                .height(8.dp)
                .clip(RoundedCornerShape(4.dp))
        )
    }
}

@Composable
fun NexusStepProgress(
    currentStep: Int,
    totalSteps: Int,
    steps: List<String>,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        steps.forEachIndexed { index, step ->
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                modifier = Modifier.weight(1f)
            ) {
                // Step circle
                Box(
                    modifier = Modifier
                        .size(32.dp)
                        .clip(CircleShape)
                        .background(
                            when {
                                index < currentStep -> MaterialTheme.colorScheme.primary
                                index == currentStep -> MaterialTheme.colorScheme.primaryContainer
                                else -> MaterialTheme.colorScheme.surfaceVariant
                            }
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    if (index < currentStep) {
                        Icon(
                            Icons.Filled.Check,
                            contentDescription = "Completed",
                            tint = MaterialTheme.colorScheme.onPrimary,
                            modifier = Modifier.size(16.dp)
                        )
                    } else {
                        Text(
                            text = "${index + 1}",
                            style = MaterialTheme.typography.labelSmall,
                            color = when {
                                index <= currentStep -> MaterialTheme.colorScheme.onPrimaryContainer
                                else -> MaterialTheme.colorScheme.onSurfaceVariant
                            }
                        )
                    }
                }

                Spacer(modifier = Modifier.height(4.dp))

                Text(
                    text = step,
                    style = MaterialTheme.typography.labelSmall,
                    color = when {
                        index <= currentStep -> MaterialTheme.colorScheme.onSurface
                        else -> MaterialTheme.colorScheme.onSurfaceVariant
                    },
                    textAlign = TextAlign.Center
                )
            }
        }
    }
}
```

### Skeleton Loading

```kotlin
@Composable
fun NexusSkeleton(
    modifier: Modifier = Modifier,
    width: Dp = Dp.Unspecified,
    height: Dp = 16.dp,
    shape: Shape = NexusCustomShapes.small
) {
    val transition = rememberInfiniteTransition(label = "skeleton")
    val alpha by transition.animateFloat(
        initialValue = 0.3f,
        targetValue = 0.7f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "skeletonAlpha"
    )

    Box(
        modifier = modifier
            .then(if (width != Dp.Unspecified) Modifier.width(width) else Modifier)
            .height(height)
            .clip(shape)
            .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = alpha))
    )
}

@Composable
fun NexusSkeletonListItem(
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(16.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        // Avatar skeleton
        NexusSkeleton(
            modifier = Modifier
                .size(48.dp)
                .clip(CircleShape)
        )

        Spacer(modifier = Modifier.width(16.dp))

        // Text skeletons
        Column(modifier = Modifier.weight(1f)) {
            NexusSkeleton(
                modifier = Modifier.fillMaxWidth(0.7f),
                height = 16.dp
            )
            Spacer(modifier = Modifier.height(8.dp))
            NexusSkeleton(
                modifier = Modifier.fillMaxWidth(0.5f),
                height = 12.dp
            )
        }
    }
}
```

---

## 14. Navigation Components

### Navigation Components

```kotlin
// ─── Top App Bar ───────────────────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NexusTopAppBar(
    title: String,
    onBackClick: (() -> Unit)? = null,
    actions: @Composable RowScope.() -> Unit = {},
    scrollBehavior: TopAppBarScrollBehavior? = null
) {
    TopAppBar(
        title = { Text(text = title) },
        navigationIcon = {
            if (onBackClick != null) {
                IconButton(onClick = onBackClick) {
                    Icon(Icons.Filled.ArrowBack, contentDescription = "Back")
                }
            }
        },
        actions = actions,
        scrollBehavior = scrollBehavior,
        colors = TopAppBarDefaults.topAppBarColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    )
}

// ─── Bottom Navigation ─────────────────────────────────────

@Composable
fun NexusBottomNavigation(
    items: List<BottomNavItem>,
    currentRoute: String?,
    onItemClick: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    NavigationBar(
        modifier = modifier,
        containerColor = MaterialTheme.colorScheme.surfaceContainer
    ) {
        items.forEach { item ->
            NavigationBarItem(
                selected = currentRoute == item.route,
                onClick = { onItemClick(item.route) },
                icon = {
                    BadgedBox(
                        badge = {
                            if (item.badgeCount != null) {
                                Badge { Text("${item.badgeCount}") }
                            }
                        }
                    ) {
                        Icon(
                            imageVector = if (currentRoute == item.route) {
                                item.selectedIcon
                            } else {
                                item.unselectedIcon
                            },
                            contentDescription = item.label
                        )
                    }
                },
                label = {
                    Text(
                        text = item.label,
                        style = MaterialTheme.typography.labelSmall
                    )
                }
            )
        }
    }
}

data class BottomNavItem(
    val route: String,
    val label: String,
    val selectedIcon: ImageVector,
    val unselectedIcon: ImageVector,
    val badgeCount: Int? = null
)

// ─── Tabs ──────────────────────────────────────────────────

@Composable
fun NexusTabs(
    tabs: List<String>,
    selectedIndex: Int,
    onTabSelected: (Int) -> Unit,
    modifier: Modifier = Modifier
) {
    TabRow(
        selectedTabIndex = selectedIndex,
        modifier = modifier,
        containerColor = MaterialTheme.colorScheme.surface,
        contentColor = MaterialTheme.colorScheme.primary
    ) {
        tabs.forEachIndexed { index, title ->
            Tab(
                selected = selectedIndex == index,
                onClick = { onTabSelected(index) },
                text = {
                    Text(
                        text = title,
                        style = if (selectedIndex == index) {
                            MaterialTheme.typography.labelLarge
                        } else {
                            MaterialTheme.typography.labelMedium
                        }
                    )
                }
            )
        }
    }
}
```

---

## 15. Media Components

### Image Component

```kotlin
@Composable
fun NexusImage(
    url: String?,
    contentDescription: String?,
    modifier: Modifier = Modifier,
    contentScale: ContentScale = ContentScale.Crop,
    placeholder: @Composable (() -> Unit)? = {
        Box(
            modifier = modifier
                .clip(RoundedCornerShape(8.dp))
                .background(MaterialTheme.colorScheme.surfaceVariant)
        )
    },
    error: @Composable (() -> Unit)? = {
        Box(
            modifier = modifier
                .clip(RoundedCornerShape(8.dp))
                .background(MaterialTheme.colorScheme.errorContainer),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                Icons.Filled.BrokenImage,
                contentDescription = "Error loading image",
                tint = MaterialTheme.colorScheme.onErrorContainer
            )
        }
    }
) {
    AsyncImage(
        model = url,
        contentDescription = contentDescription,
        modifier = modifier.clip(RoundedCornerShape(8.dp)),
        contentScale = contentScale,
        placeholder = placeholder?.let {
            { it() }
        },
        error = error?.let {
            { it() }
        }
    )
}

@Composable
fun NexusAvatar(
    url: String?,
    name: String,
    size: Dp = 40.dp,
    modifier: Modifier = Modifier
) {
    val initials = name.split(" ")
        .mapNotNull { it.firstOrNull()?.toString() }
        .take(2)
        .joinToString("")

    Box(
        modifier = modifier
            .size(size)
            .clip(CircleShape),
        contentAlignment = Alignment.Center
    ) {
        if (url != null) {
            AsyncImage(
                model = url,
                contentDescription = name,
                modifier = Modifier.fillMaxSize(),
                contentScale = ContentScale.Crop
            )
        } else {
            Surface(
                modifier = Modifier.fillMaxSize(),
                color = MaterialTheme.colorScheme.primaryContainer
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(
                        text = initials,
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.onPrimaryContainer
                    )
                }
            }
        }
    }
}
```

---

## 16. Chat Components

### Message Bubble

```kotlin
@Composable
fun MessageBubble(
    message: Message,
    isQueued: Boolean = false,
    isFailed: Boolean = false,
    onLongClick: (() -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    val isUser = message.role == MessageRole.USER

    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 2.dp),
        horizontalArrangement = if (isUser) Arrangement.End else Arrangement.Start
    ) {
        if (!isUser) {
            NexusAvatar(
                url = null,
                name = "AI",
                size = 28.dp
            )
            Spacer(modifier = Modifier.width(8.dp))
        }

        Column(
            modifier = Modifier
                .widthIn(max = 280.dp)
                .combinedClickable(
                    onClick = {},
                    onLongClick = onLongClick
                )
                .background(
                    color = if (isUser) {
                        MaterialTheme.colorScheme.primaryContainer
                    } else {
                        MaterialTheme.colorScheme.surfaceVariant
                    },
                    shape = NexusCustomShapes.chatBubbleUser
                )
                .padding(12.dp)
        ) {
            Text(
                text = message.content,
                style = MaterialTheme.typography.bodyMedium,
                color = if (isUser) {
                    MaterialTheme.colorScheme.onPrimaryContainer
                } else {
                    MaterialTheme.colorScheme.onSurfaceVariant
                }
            )

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = formatTime(message.timestamp),
                    style = MaterialTheme.typography.labelSmall,
                    color = if (isUser) {
                        MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                    } else {
                        MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f)
                    }
                )

                if (isUser) {
                    Spacer(modifier = Modifier.width(4.dp))
                    Icon(
                        imageVector = when {
                            isFailed -> Icons.Filled.ErrorOutline
                            isQueued -> Icons.Filled.Schedule
                            else -> Icons.Filled.Check
                        },
                        contentDescription = when {
                            isFailed -> "Failed"
                            isQueued -> "Queued"
                            else -> "Sent"
                        },
                        modifier = Modifier.size(12.dp),
                        tint = when {
                            isFailed -> MaterialTheme.colorScheme.error
                            isQueued -> MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                            else -> MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                        }
                    )
                }
            }
        }

        if (isUser) {
            Spacer(modifier = Modifier.width(8.dp))
            NexusAvatar(
                url = null,
                name = "You",
                size = 28.dp
            )
        }
    }
}

// ─── Chat Input Bar ────────────────────────────────────────

@Composable
fun ChatInputBar(
    text: String,
    onTextChange: (String) -> Unit,
    onSend: () -> Unit,
    onAttach: () -> Unit,
    isOnline: Boolean,
    isSending: Boolean = false,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier,
        tonalElevation = 2.dp,
        shadowElevation = 8.dp
    ) {
        Row(
            modifier = Modifier
                .padding(horizontal = 8.dp, vertical = 8.dp)
                .imePadding(),
            verticalAlignment = Alignment.Bottom
        ) {
            IconButton(onClick = onAttach) {
                Icon(
                    Icons.Filled.AttachFile,
                    contentDescription = "Attach file"
                )
            }

            OutlinedTextField(
                value = text,
                onValueChange = onTextChange,
                modifier = Modifier
                    .weight(1f)
                    .heightIn(min = 48.dp, max = 120.dp),
                placeholder = {
                    Text(
                        if (isOnline) "Type a message..." else "Offline - message will be queued"
                    )
                },
                maxLines = 4,
                shape = NexusCustomShapes.extraLarge,
                colors = OutlinedTextFieldDefaults.colors(
                    unfocusedBorderColor = Color.Transparent,
                    focusedBorderColor = Color.Transparent
                )
            )

            Spacer(modifier = Modifier.width(4.dp))

            FilledIconButton(
                onClick = onSend,
                enabled = text.isNotBlank() && !isSending
            ) {
                if (isSending) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(20.dp),
                        strokeWidth = 2.dp,
                        color = MaterialTheme.colorScheme.onPrimary
                    )
                } else {
                    Icon(
                        Icons.Filled.Send,
                        contentDescription = "Send"
                    )
                }
            }
        }
    }
}

private fun formatTime(timestamp: Long): String {
    val now = System.currentTimeMillis()
    val diff = now - timestamp

    return when {
        diff < 60_000 -> "now"
        diff < 3_600_000 -> "${diff / 60_000}m ago"
        diff < 86_400_000 -> SimpleDateFormat("HH:mm", Locale.getDefault()).format(Date(timestamp))
        else -> SimpleDateFormat("MMM d", Locale.getDefault()).format(Date(timestamp))
    }
}
```

---

## 17. Agent Components

### Agent Card

```kotlin
@Composable
fun AgentCard(
    agent: Agent,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    NexusElevatedCard(
        onClick = onClick,
        modifier = modifier
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically
            ) {
                NexusAvatar(
                    url = agent.avatarUrl,
                    name = agent.name,
                    size = 48.dp
                )

                Spacer(modifier = Modifier.width(12.dp))

                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = agent.name,
                        style = MaterialTheme.typography.titleMedium
                    )
                    Text(
                        text = agent.model,
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }

                // Status badge
                AgentStatusBadge(isActive = agent.isActive)
            }

            if (agent.description != null) {
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    text = agent.description,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis
                )
            }
        }
    }
}

@Composable
fun AgentStatusBadge(
    isActive: Boolean,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier,
        shape = NexusCustomShapes.chip,
        color = if (isActive) {
            MaterialTheme.colorScheme.primaryContainer
        } else {
            MaterialTheme.colorScheme.surfaceVariant
        }
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            Box(
                modifier = Modifier
                    .size(6.dp)
                    .clip(CircleShape)
                    .background(
                        if (isActive) NexusColors.online
                        else NexusColors.offline
                    )
            )
            Text(
                text = if (isActive) "Active" else "Inactive",
                style = MaterialTheme.typography.labelSmall,
                color = if (isActive) {
                    MaterialTheme.colorScheme.onPrimaryContainer
                } else {
                    MaterialTheme.colorScheme.onSurfaceVariant
                }
            )
        }
    }
}

// ─── Execution Step ────────────────────────────────────────

@Composable
fun ExecutionStep(
    step: ExecutionStep,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.Top
    ) {
        // Status icon
        Box(
            modifier = Modifier
                .size(24.dp)
                .clip(CircleShape)
                .background(
                    when (step.status) {
                        ExecutionStatus.COMPLETED -> MaterialTheme.colorScheme.primary
                        ExecutionStatus.RUNNING -> MaterialTheme.colorScheme.tertiary
                        ExecutionStatus.PENDING -> MaterialTheme.colorScheme.surfaceVariant
                        ExecutionStatus.FAILED -> MaterialTheme.colorScheme.error
                    }
                ),
            contentAlignment = Alignment.Center
        ) {
            when (step.status) {
                ExecutionStatus.COMPLETED -> Icon(
                    Icons.Filled.Check,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.onPrimary,
                    modifier = Modifier.size(14.dp)
                )
                ExecutionStatus.RUNNING -> CircularProgressIndicator(
                    modifier = Modifier.size(14.dp),
                    strokeWidth = 2.dp,
                    color = MaterialTheme.colorScheme.onTertiary
                )
                ExecutionStatus.PENDING -> Icon(
                    Icons.Filled.Schedule,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.size(14.dp)
                )
                ExecutionStatus.FAILED -> Icon(
                    Icons.Filled.Close,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.onError,
                    modifier = Modifier.size(14.dp)
                )
            }
        }

        Spacer(modifier = Modifier.width(12.dp))

        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = step.title,
                style = MaterialTheme.typography.bodyMedium
            )
            if (step.description != null) {
                Text(
                    text = step.description,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

data class ExecutionStep(
    val title: String,
    val description: String? = null,
    val status: ExecutionStatus = ExecutionStatus.PENDING
)

enum class ExecutionStatus {
    COMPLETED, RUNNING, PENDING, FAILED
}
```

---

## 18. Document Components

### Document Card

```kotlin
@Composable
fun DocumentCard(
    document: Document,
    onClick: () -> Unit,
    onLongClick: (() -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    NexusElevatedCard(
        onClick = onClick,
        modifier = modifier
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Document type icon
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(NexusCustomShapes.small)
                    .background(getDocumentTypeColor(document.mimeType)),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = getDocumentTypeIcon(document.mimeType),
                    contentDescription = null,
                    tint = getDocumentTypeIconColor(document.mimeType),
                    modifier = Modifier.size(24.dp)
                )
            }

            Spacer(modifier = Modifier.width(12.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = document.name,
                    style = MaterialTheme.typography.titleSmall,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )

                Spacer(modifier = Modifier.height(4.dp))

                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    // Processing status
                    ProcessingStatusBadge(status = document.processingStatus)

                    // File size
                    Text(
                        text = formatFileSize(document.sizeBytes),
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }

                if (document.chunkCount > 0) {
                    Text(
                        text = "${document.chunkCount} chunks",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            // Sync status
            SyncStatusIcon(status = document.syncStatus)
        }
    }
}

@Composable
fun ProcessingStatusBadge(
    status: ProcessingStatus,
    modifier: Modifier = Modifier
) {
    val (color, text) = when (status) {
        ProcessingStatus.COMPLETED -> MaterialTheme.colorScheme.primary to "Processed"
        ProcessingStatus.PROCESSING -> MaterialTheme.colorScheme.tertiary to "Processing"
        ProcessingStatus.PENDING -> MaterialTheme.colorScheme.surfaceVariant to "Pending"
        ProcessingStatus.FAILED -> MaterialTheme.colorScheme.error to "Failed"
    }

    Surface(
        modifier = modifier,
        shape = NexusCustomShapes.chip,
        color = color.copy(alpha = 0.12f)
    ) {
        Text(
            text = text,
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
            style = MaterialTheme.typography.labelSmall,
            color = color
        )
    }
}

@Composable
fun SyncStatusIcon(
    status: SyncStatus,
    modifier: Modifier = Modifier
) {
    Icon(
        imageVector = when (status) {
            SyncStatus.SYNCED -> Icons.Filled.CloudDone
            SyncStatus.PENDING -> Icons.Filled.CloudQueue
            SyncStatus.FAILED -> Icons.Filled.CloudOff
            SyncStatus.CONFLICT -> Icons.Filled.CloudSync
        },
        contentDescription = when (status) {
            SyncStatus.SYNCED -> "Synced"
            SyncStatus.PENDING -> "Pending sync"
            SyncStatus.FAILED -> "Sync failed"
            SyncStatus.CONFLICT -> "Sync conflict"
        },
        modifier = modifier.size(16.dp),
        tint = when (status) {
            SyncStatus.SYNCED -> NexusColors.synced
            SyncStatus.PENDING -> NexusColors.pending
            SyncStatus.FAILED -> NexusColors.failed
            SyncStatus.CONFLICT -> NexusColors.conflict
        }
    )
}

fun getDocumentTypeIcon(mimeType: String): ImageVector = when {
    mimeType.startsWith("image/") -> Icons.Filled.Image
    mimeType.startsWith("video/") -> Icons.Filled.VideoFile
    mimeType.startsWith("audio/") -> Icons.Filled.AudioFile
    mimeType == "application/pdf" -> Icons.Filled.PictureAsPdf
    mimeType.contains("word") || mimeType.contains("document") -> Icons.Filled.Description
    mimeType.contains("sheet") || mimeType.contains("excel") -> Icons.Filled.TableChart
    mimeType.contains("presentation") -> Icons.Filled.Slideshow
    else -> Icons.Filled.InsertDriveFile
}

fun getDocumentTypeColor(mimeType: String): Color = when {
    mimeType.startsWith("image/") -> Color(0xFFE8F5E9)
    mimeType.startsWith("video/") -> Color(0xFFFCE4EC)
    mimeType.startsWith("audio/") -> Color(0xFFF3E5F5)
    mimeType == "application/pdf" -> Color(0xFFFFEBEE)
    mimeType.contains("word") -> Color(0xFFE3F2FD)
    mimeType.contains("sheet") -> Color(0xFFE8F5E9)
    else -> Color(0xFFF5F5F5)
}

fun getDocumentTypeIconColor(mimeType: String): Color = when {
    mimeType.startsWith("image/") -> Color(0xFF2E7D32)
    mimeType.startsWith("video/") -> Color(0xFFC62828)
    mimeType.startsWith("audio/") -> Color(0xFF6A1B9A)
    mimeType == "application/pdf" -> Color(0xFFC62828)
    mimeType.contains("word") -> Color(0xFF1565C0)
    mimeType.contains("sheet") -> Color(0xFF2E7D32)
    else -> Color(0xFF616161)
}

fun formatFileSize(bytes: Long): String {
    val kb = bytes / 1024.0
    val mb = kb / 1024.0
    val gb = mb / 1024.0
    return when {
        gb >= 1.0 -> "%.1f GB".format(gb)
        mb >= 1.0 -> "%.1f MB".format(mb)
        kb >= 1.0 -> "%.0f KB".format(kb)
        else -> "$bytes B"
    }
}
```

---

## 19. Chart Components

### Line Chart

```kotlin
@Composable
fun NexusLineChart(
    data: List<Float>,
    labels: List<String>,
    modifier: Modifier = Modifier,
    lineColor: Color = MaterialTheme.colorScheme.primary,
    showGrid: Boolean = true,
    showLabels: Boolean = true
) {
    val primaryColor = MaterialTheme.colorScheme.primary
    val outlineColor = MaterialTheme.colorScheme.outlineVariant

    Canvas(
        modifier = modifier
            .fillMaxWidth()
            .height(200.dp)
    ) {
        if (data.isEmpty()) return@Canvas

        val maxValue = data.max()
        val minValue = data.min()
        val range = maxValue - minValue
        val stepX = size.width / (data.size - 1).coerceAtLeast(1)

        // Draw grid lines
        if (showGrid) {
            for (i in 0..4) {
                val y = size.height * i / 4
                drawLine(
                    color = outlineColor,
                    start = Offset(0f, y),
                    end = Offset(size.width, y),
                    strokeWidth = 1.dp.toPx(),
                    pathEffect = PathEffect.dashPathEffect(floatArrayOf(8f, 4f))
                )
            }
        }

        // Draw line
        val path = Path()
        data.forEachIndexed { index, value ->
            val x = index * stepX
            val y = size.height - ((value - minValue) / range * size.height)

            if (index == 0) {
                path.moveTo(x, y)
            } else {
                path.lineTo(x, y)
            }
        }

        drawPath(
            path = path,
            color = lineColor,
            style = Stroke(
                width = 2.dp.toPx(),
                cap = StrokeCap.Round,
                join = StrokeJoin.Round
            )
        )

        // Draw dots
        data.forEachIndexed { index, value ->
            val x = index * stepX
            val y = size.height - ((value - minValue) / range * size.height)

            drawCircle(
                color = lineColor,
                radius = 4.dp.toPx(),
                center = Offset(x, y)
            )
        }
    }
}

// ─── Bar Chart ─────────────────────────────────────────────

@Composable
fun NexusBarChart(
    data: List<Pair<String, Float>>,
    modifier: Modifier = Modifier,
    barColor: Color = MaterialTheme.colorScheme.primary,
    showValues: Boolean = true
) {
    val maxValue = data.maxOfOrNull { it.second } ?: 0f

    Column(modifier = modifier) {
        Canvas(
            modifier = Modifier
                .fillMaxWidth()
                .height(150.dp)
        ) {
            val barWidth = size.width / (data.size * 1.5f)
            val spacing = barWidth / 2

            data.forEachIndexed { index, (_, value) ->
                val barHeight = (value / maxValue) * size.height
                val x = index * (barWidth + spacing) + spacing

                drawRect(
                    color = barColor,
                    topLeft = Offset(x, size.height - barHeight),
                    size = Size(barWidth, barHeight)
                )
            }
        }

        if (showValues) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                data.forEach { (label, _) ->
                    Text(
                        text = label,
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
    }
}

// ─── Gauge ─────────────────────────────────────────────────

@Composable
fun NexusGauge(
    value: Float,
    maxValue: Float = 100f,
    label: String,
    modifier: Modifier = Modifier,
    color: Color = MaterialTheme.colorScheme.primary
) {
    val progress = (value / maxValue).coerceIn(0f, 1f)

    Column(
        modifier = modifier,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Box(
            modifier = Modifier.size(80.dp),
            contentAlignment = Alignment.Center
        ) {
            CircularProgressIndicator(
                progress = { progress },
                modifier = Modifier.fillMaxSize(),
                color = color,
                strokeWidth = 8.dp,
                trackColor = MaterialTheme.colorScheme.surfaceVariant
            )

            Text(
                text = "${(progress * 100).toInt()}%",
                style = MaterialTheme.typography.titleMedium,
                color = color
            )
        }

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = label,
            style = MaterialTheme.typography.labelMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
    }
}
```

---

## 20. Loading Components

### Loading Patterns

```kotlin
// ─── Shimmer Loading ───────────────────────────────────────

@Composable
fun NexusShimmer(
    modifier: Modifier = Modifier
) {
    val transition = rememberInfiniteTransition(label = "shimmer")
    val translateAnim = transition.animateFloat(
        initialValue = 0f,
        targetValue = 1000f,
        animationSpec = infiniteRepeatable(
            animation = tween(1200, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "shimmerTranslate"
    )

    Brush.linearGradient(
        colors = listOf(
            MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f),
            MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.8f),
            MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f)
        ),
        start = Offset.Zero,
        end = Offset(x = translateAnim.value, y = translateAnim.value)
    )
}

@Composable
fun NexusShimmerListItem(modifier: Modifier = Modifier) {
    val shimmerBrush = NexusShimmer()

    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(16.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(48.dp)
                .clip(CircleShape)
                .background(shimmerBrush)
        )

        Spacer(modifier = Modifier.width(16.dp))

        Column(modifier = Modifier.weight(1f)) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(0.7f)
                    .height(16.dp)
                    .clip(NexusCustomShapes.small)
                    .background(shimmerBrush)
            )
            Spacer(modifier = Modifier.height(8.dp))
            Box(
                modifier = Modifier
                    .fillMaxWidth(0.5f)
                    .height(12.dp)
                    .clip(NexusCustomShapes.small)
                    .background(shimmerBrush)
            )
        }
    }
}

// ─── Full Screen Loading ───────────────────────────────────

@Composable
fun NexusFullScreenLoading(
    message: String = "Loading...",
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            CircularProgressIndicator(
                modifier = Modifier.size(48.dp),
                color = MaterialTheme.colorScheme.primary,
                strokeWidth = 4.dp
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = message,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

// ─── Loading Overlay ───────────────────────────────────────

@Composable
fun NexusLoadingOverlay(
    isLoading: Boolean,
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit
) {
    Box(modifier = modifier) {
        content()

        if (isLoading) {
            Surface(
                modifier = Modifier.fillMaxSize(),
                color = MaterialTheme.colorScheme.surface.copy(alpha = 0.7f)
            ) {
                Box(contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(48.dp)
                    )
                }
            }
        }
    }
}

// ─── Pull to Refresh ───────────────────────────────────────

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NexusPullToRefresh(
    isRefreshing: Boolean,
    onRefresh: () -> Unit,
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit
) {
    PullToRefreshBox(
        isRefreshing = isRefreshing,
        onRefresh = onRefresh,
        modifier = modifier
    ) {
        content()
    }
}
```

---

## 21. Empty State Components

### Empty State Patterns

```kotlin
@Composable
fun NexusEmptyState(
    icon: ImageVector,
    title: String,
    message: String,
    actionText: String? = null,
    onAction: (() -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            imageVector = icon,
            contentDescription = null,
            modifier = Modifier.size(80.dp),
            tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f)
        )

        Spacer(modifier = Modifier.height(24.dp))

        Text(
            text = title,
            style = MaterialTheme.typography.headlineSmall,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurface
        )

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = message,
            style = MaterialTheme.typography.bodyMedium,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        if (actionText != null && onAction != null) {
            Spacer(modifier = Modifier.height(24.dp))

            NexusFilledButton(
                text = actionText,
                onClick = onAction
            )
        }
    }
}

// ─── Pre-built Empty States ────────────────────────────────

@Composable
fun EmptyConversations(modifier: Modifier = Modifier) {
    NexusEmptyState(
        icon = Icons.Filled.ChatBubbleOutline,
        title = "No conversations yet",
        message = "Start a new conversation to begin chatting with AI agents.",
        actionText = "New Conversation",
        modifier = modifier
    )
}

@Composable
fun EmptyDocuments(modifier: Modifier = Modifier) {
    NexusEmptyState(
        icon = Icons.Filled.FolderOpen,
        title = "No documents",
        message = "Upload documents to build your knowledge base.",
        actionText = "Upload Document",
        modifier = modifier
    )
}

@Composable
fun EmptyAgents(modifier: Modifier = Modifier) {
    NexusEmptyState(
        icon = Icons.Filled.SmartToy,
        title = "No agents configured",
        message = "Create or configure AI agents to get started.",
        actionText = "Create Agent",
        modifier = modifier
    )
}

@Composable
fun EmptySearchResults(query: String, modifier: Modifier = Modifier) {
    NexusEmptyState(
        icon = Icons.Filled.SearchOff,
        title = "No results for \"$query\"",
        message = "Try different keywords or check your spelling.",
        modifier = modifier
    )
}

@Composable
fun EmptyNotifications(modifier: Modifier = Modifier) {
    NexusEmptyState(
        icon = Icons.Filled.NotificationsNone,
        title = "All caught up!",
        message = "You have no new notifications.",
        modifier = modifier
    )
}
```

---

## 22. Error State Components

### Error State Patterns

```kotlin
@Composable
fun NexusErrorState(
    title: String = "Something went wrong",
    message: String = "An unexpected error occurred. Please try again.",
    actionText: String = "Retry",
    onRetry: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            imageVector = Icons.Filled.ErrorOutline,
            contentDescription = null,
            modifier = Modifier.size(80.dp),
            tint = MaterialTheme.colorScheme.error.copy(alpha = 0.7f)
        )

        Spacer(modifier = Modifier.height(24.dp))

        Text(
            text = title,
            style = MaterialTheme.typography.headlineSmall,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurface
        )

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = message,
            style = MaterialTheme.typography.bodyMedium,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Spacer(modifier = Modifier.height(24.dp))

        NexusFilledButton(
            text = actionText,
            onClick = onRetry
        )
    }
}

// ─── Pre-built Error States ────────────────────────────────

@Composable
fun NetworkError(onRetry: () -> Unit, modifier: Modifier = Modifier) {
    NexusErrorState(
        title = "No internet connection",
        message = "Please check your network settings and try again.",
        actionText = "Retry",
        onRetry = onRetry,
        modifier = modifier
    )
}

@Composable
fun ServerError(onRetry: () -> Unit, modifier: Modifier = Modifier) {
    NexusErrorState(
        title = "Server error",
        message = "Our servers are experiencing issues. Please try again later.",
        actionText = "Retry",
        onRetry = onRetry,
        modifier = modifier
    )
}

@Composable
fun PermissionError(onGrant: () -> Unit, modifier: Modifier = Modifier) {
    NexusErrorState(
        title = "Permission required",
        message = "This feature requires additional permissions to work.",
        actionText = "Grant Permission",
        onRetry = onGrant,
        modifier = modifier
    )
}

@Composable
fun ErrorRetryItem(
    message: String,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier
            .fillMaxWidth()
            .padding(16.dp),
        shape = NexusCustomShapes.small,
        color = MaterialTheme.colorScheme.errorContainer
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                Icons.Filled.Error,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.onErrorContainer,
                modifier = Modifier.size(24.dp)
            )

            Spacer(modifier = Modifier.width(12.dp))

            Text(
                text = message,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onErrorContainer,
                modifier = Modifier.weight(1f)
            )

            NexusTextButton(
                text = "Retry",
                onClick = onRetry
            )
        }
    }
}
```

---

## 23. Animation Components

### Animation Components

```kotlin
// ─── Pulse Animation ───────────────────────────────────────

@Composable
fun NexusPulse(
    modifier: Modifier = Modifier,
    color: Color = MaterialTheme.colorScheme.primary
) {
    val transition = rememberInfiniteTransition(label = "pulse")
    val scale by transition.animateFloat(
        initialValue = 1f,
        targetValue = 1.2f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulseScale"
    )
    val alpha by transition.animateFloat(
        initialValue = 0.8f,
        targetValue = 0.3f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulseAlpha"
    )

    Box(modifier = modifier) {
        // Outer ring
        Box(
            modifier = Modifier
                .size(24.dp)
                .scale(scale)
                .clip(CircleShape)
                .background(color.copy(alpha = alpha))
        )

        // Inner dot
        Box(
            modifier = Modifier
                .size(12.dp)
                .clip(CircleShape)
                .background(color)
                .align(Alignment.Center)
        )
    }
}

// ─── Spin Animation ────────────────────────────────────────

@Composable
fun NexusSpin(
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit
) {
    val transition = rememberInfiniteTransition(label = "spin")
    val rotation by transition.animateFloat(
        initialValue = 0f,
        targetValue = 360f,
        animationSpec = infiniteRepeatable(
            animation = tween(2000, easing = LinearEasing)
        ),
        label = "spinRotation"
    )

    Box(
        modifier = modifier.rotate(rotation)
    ) {
        content()
    }
}

// ─── Slide Animation ───────────────────────────────────────

@Composable
fun NexusSlideIn(
    visible: Boolean,
    direction: SlideDirection = SlideDirection.Left,
    content: @Composable () -> Unit
) {
    val offsetX by animateFloatAsState(
        targetValue = if (visible) 0f else when (direction) {
            SlideDirection.Left -> -300f
            SlideDirection.Right -> 300f
        },
        animationSpec = tween(300),
        label = "slideOffset"
    )

    val alpha by animateFloatAsState(
        targetValue = if (visible) 1f else 0f,
        animationSpec = tween(300),
        label = "slideAlpha"
    )

    Box(
        modifier = Modifier
            .offset { IntOffset(offsetX.roundToInt(), 0) }
            .alpha(alpha)
    ) {
        content()
    }
}

enum class SlideDirection { Left, Right }

// ─── Fade Animation ────────────────────────────────────────

@Composable
fun NexusFadeIn(
    visible: Boolean,
    content: @Composable () -> Unit
) {
    val alpha by animateFloatAsState(
        targetValue = if (visible) 1f else 0f,
        animationSpec = tween(300),
        label = "fadeIn"
    )

    Box(modifier = Modifier.alpha(alpha)) {
        content()
    }
}

// ─── Scale Animation ───────────────────────────────────────

@Composable
fun NexusScaleIn(
    visible: Boolean,
    content: @Composable () -> Unit
) {
    val scale by animateFloatAsState(
        targetValue = if (visible) 1f else 0.8f,
        animationSpec = spring(
            dampingRatio = Spring.DampingRatioMediumBouncy,
            stiffness = Spring.StiffnessLow
        ),
        label = "scaleIn"
    )
    val alpha by animateFloatAsState(
        targetValue = if (visible) 1f else 0f,
        animationSpec = tween(200),
        label = "scaleAlpha"
    )

    Box(
        modifier = Modifier
            .scale(scale)
            .alpha(alpha)
    ) {
        content()
    }
}

// ─── Online Status Pulse ───────────────────────────────────

@Composable
fun OnlinePulse(
    isOnline: Boolean,
    modifier: Modifier = Modifier
) {
    val color = if (isOnline) NexusColors.online else NexusColors.offline

    if (isOnline) {
        NexusPulse(
            modifier = modifier,
            color = color
        )
    } else {
        Box(
            modifier = modifier
                .size(12.dp)
                .clip(CircleShape)
                .background(color)
        )
    }
}
```

---

## 24. Icon System

### Icon Configuration

```kotlin
// ─── App Icon Set ──────────────────────────────────────────

object NexusIcons {
    // Navigation
    val Dashboard = Icons.Filled.Dashboard
    val Chat = Icons.Filled.Chat
    val Agents = Icons.Filled.SmartToy
    val Knowledge = Icons.Filled.LibraryBooks
    val Settings = Icons.Filled.Settings
    val More = Icons.Filled.MoreHoriz

    // Actions
    val Add = Icons.Filled.Add
    val Edit = Icons.Filled.Edit
    val Delete = Icons.Filled.Delete
    val Share = Icons.Filled.Share
    val Copy = Icons.Filled.ContentCopy
    val Search = Icons.Filled.Search
    val Filter = Icons.Filled.FilterList
    val Sort = Icons.Filled.Sort
    val Refresh = Icons.Filled.Refresh
    val Download = Icons.Filled.Download
    val Upload = Icons.Filled.Upload
    val Send = Icons.Filled.Send
    val Attach = Icons.Filled.AttachFile
    val Record = Icons.Filled.Mic
    val Stop = Icons.Filled.Stop

    // Navigation
    val Back = Icons.Filled.ArrowBack
    val Forward = Icons.Filled.ArrowForward
    val Close = Icons.Filled.Close
    val Expand = Icons.Filled.ExpandMore
    val Collapse = Icons.Filled.ExpandLess

    // Status
    val Success = Icons.Filled.CheckCircle
    val Error = Icons.Filled.Error
    val Warning = Icons.Filled.Warning
    val Info = Icons.Filled.Info
    val Sync = Icons.Filled.Sync
    val Synced = Icons.Filled.CloudDone
    val SyncPending = Icons.Filled.CloudQueue
    val SyncFailed = Icons.Filled.CloudOff

    // Chat
    val User = Icons.Filled.Person
    val Assistant = Icons.Filled.SmartToy
    val System = Icons.Filled.Info
    val Reply = Icons.Filled.Reply
    val Forward = Icons.Filled.Forward
    val Bookmark = Icons.Filled.Bookmark

    // Documents
    val Document = Icons.Filled.Description
    val Image = Icons.Filled.Image
    val Video = Icons.Filled.VideoFile
    val Audio = Icons.Filled.AudioFile
    val PDF = Icons.Filled.PictureAsPdf

    // Empty States
    val EmptyChat = Icons.Outlined.ChatBubbleOutline
    val EmptyDocument = Icons.Outlined.FolderOpen
    val EmptySearch = Icons.Outlined.SearchOff
    val EmptyNotification = Icons.Outlined.NotificationsNone
    val EmptyAgent = Icons.Outlined.SmartToy
}

// ─── Icon Component ────────────────────────────────────────

@Composable
fun NexusIcon(
    icon: ImageVector,
    contentDescription: String?,
    modifier: Modifier = Modifier,
    size: Dp = 24.dp,
    tint: Color = MaterialTheme.colorScheme.onSurface
) {
    Icon(
        imageVector = icon,
        contentDescription = contentDescription,
        modifier = modifier.size(size),
        tint = tint
    )
}

// ─── Vector Icon Definitions ───────────────────────────────

// For custom icons, define in res/drawable/
// Use ImageVector.Builder for programmatic vector icons
```

---

## 25. Responsive Layout

### Responsive Layout System

```kotlin
// ─── BoxWithConstraints ────────────────────────────────────

@Composable
fun ResponsiveLayout(
    modifier: Modifier = Modifier,
    content: @Composable (ScreenSize) -> Unit
) {
    BoxWithConstraints(modifier = modifier) {
        val screenSize = when {
            maxWidth < 600.dp -> ScreenSize.COMPACT
            maxWidth < 840.dp -> ScreenSize.MEDIUM
            else -> ScreenSize.EXPANDED
        }

        content(screenSize)
    }
}

enum class ScreenSize {
    COMPACT,   // Phone (< 600dp)
    MEDIUM,    // Tablet portrait (600-840dp)
    EXPANDED   // Tablet landscape / Desktop (> 840dp)
}

// ─── Responsive Grid ───────────────────────────────────────

@Composable
fun ResponsiveGrid(
    items: List<@Composable () -> Unit>,
    modifier: Modifier = Modifier,
    columns: Int? = null
) {
    BoxWithConstraints(modifier = modifier) {
        val effectiveColumns = columns ?: when {
            maxWidth < 600.dp -> 1
            maxWidth < 840.dp -> 2
            else -> 3
        }

        LazyVerticalGrid(
            columns = GridCells.Fixed(effectiveColumns),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(items.size) { index ->
                items[index]()
            }
        }
    }
}

// ─── Adaptive Layout ───────────────────────────────────────

@Composable
fun AdaptiveLayout(
    listContent: @Composable () -> Unit,
    detailContent: (@Composable () -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    BoxWithConstraints(modifier = modifier) {
        if (maxWidth >= 840.dp && detailContent != null) {
            // Two-pane layout
            Row(modifier = Modifier.fillMaxSize()) {
                Box(modifier = Modifier.weight(1f)) {
                    listContent()
                }

                VerticalDivider(
                    modifier = Modifier
                        .fillMaxHeight()
                        .width(1.dp)
                )

                Box(modifier = Modifier.weight(2f)) {
                    detailContent()
                }
            }
        } else {
            // Single pane layout
            listContent()
        }
    }
}
```

### Responsive Breakpoints

| Breakpoint | Width | Layout | Columns |
|-----------|-------|--------|---------|
| Compact | < 600dp | Single pane | 1 |
| Medium | 600-840dp | Two-pane possible | 2 |
| Expanded | > 840dp | Two-pane / Side nav | 3+ |

---

## 26. Foldable Device Support

### Foldable Layout

```kotlin
@Composable
fun FoldableLayout(
    listContent: @Composable () -> Unit,
    detailContent: @Composable () -> Unit,
    modifier: Modifier = Modifier
) {
    val windowSizeClass = calculateWindowSizeClass(
        LocalContext.current as Activity
    )

    when (windowSizeClass.widthSizeClass) {
        WindowWidthSizeClass.Compact -> {
            // Phone layout
            listContent()
        }
        WindowWidthSizeClass.Medium -> {
            // Tablet / Foldable (half-open)
            Row(modifier = modifier.fillMaxSize()) {
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxHeight()
                ) {
                    listContent()
                }

                VerticalDivider()

                Box(
                    modifier = Modifier
                        .weight(1.5f)
                        .fillMaxHeight()
                ) {
                    detailContent()
                }
            }
        }
        WindowWidthSizeClass.Expanded -> {
            // Desktop / Tablet landscape
            Row(modifier = modifier.fillMaxSize()) {
                // Navigation rail
                NavigationRail(
                    modifier = Modifier
                        .fillMaxHeight()
                        .width(80.dp)
                ) {
                    // Rail items
                }

                // List
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxHeight()
                ) {
                    listContent()
                }

                VerticalDivider()

                // Detail
                Box(
                    modifier = Modifier
                        .weight(2f)
                        .fillMaxHeight()
                ) {
                    detailContent()
                }
            }
        }
    }
}

// ─── Window Size Class Usage ───────────────────────────────

@Composable
fun FoldableAwareContent() {
    val windowSizeClass = calculateWindowSizeClass(
        LocalContext.current as Activity
    )

    when (windowSizeClass.heightSizeClass) {
        WindowHeightSizeClass.Compact -> {
            // Landscape / short screen
        }
        WindowHeightSizeClass.Medium -> {
            // Normal portrait
        }
        WindowHeightSizeClass.Expanded -> {
            // Tall screen / unfolded foldable
        }
    }
}
```

### Foldable Device States

```
┌─────────────────────────────────────────────────────────┐
│              Foldable Device States                       │
│                                                         │
│  Folded:              Half-Open:         Fully Open:    │
│  ┌──────────┐        ┌──────────┐       ┌──────────┐  │
│  │          │        │          │       │          │  │
│  │  Phone   │        │  Tablet  │       │  Desktop │  │
│  │  Layout  │        │  Layout  │       │  Layout  │  │
│  │          │        │          │       │          │  │
│  └──────────┘        └──────────┘       └──────────┘  │
│                                                         │
│  Width: <600dp       Width: 600-840dp   Width: >840dp  │
│  Columns: 1          Columns: 2         Columns: 3     │
│  Pane: Single        Pane: Two-pane     Pane: Multi    │
└─────────────────────────────────────────────────────────┘
```

---

## 27. Tablet Layout

### Tablet Two-Pane Layout

```kotlin
@Composable
fun TabletLayout(
    navController: NavHostController
) {
    var selectedId by rememberSaveable { mutableStateOf<String?>(null) }

    Row(modifier = Modifier.fillMaxSize()) {
        // List pane
        Box(
            modifier = Modifier
                .weight(1f)
                .fillMaxHeight()
        ) {
            // Navigation rail + list
            NavigationRail {
                // Rail items
            }

            // Content list
            LazyColumn(
                modifier = Modifier.padding(start = 80.dp)
            ) {
                items(items) { item ->
                    ListItem(
                        headlineContent = { Text(item.title) },
                        selected = selectedId == item.id,
                        onClick = { selectedId = item.id }
                    )
                }
            }
        }

        VerticalDivider(
            modifier = Modifier
                .fillMaxHeight()
                .width(1.dp)
        )

        // Detail pane
        Box(
            modifier = Modifier
                .weight(2f)
                .fillMaxHeight()
        ) {
            if (selectedId != null) {
                DetailContent(id = selectedId!!)
            } else {
                // Empty state
                NexusEmptyState(
                    icon = Icons.Filled.TouchApp,
                    title = "Select an item",
                    message = "Choose an item from the list to view details."
                )
            }
        }
    }
}

// ─── Navigation Rail for Tablets ───────────────────────────

@Composable
fun TabletNavigationRail(
    items: List<BottomNavItem>,
    currentRoute: String?,
    onItemClick: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    NavigationRail(
        modifier = modifier,
        containerColor = MaterialTheme.colorScheme.surfaceContainer
    ) {
        items.forEach { item ->
            NavigationRailItem(
                selected = currentRoute == item.route,
                onClick = { onItemClick(item.route) },
                icon = {
                    Icon(
                        imageVector = if (currentRoute == item.route) {
                            item.selectedIcon
                        } else {
                            item.unselectedIcon
                        },
                        contentDescription = item.label
                    )
                },
                label = { Text(item.label) }
            )
        }
    }
}
```

---

## 28. Dark Mode Support

### Dark Mode Configuration

```kotlin
@Composable
fun NexusTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    dynamicColor: Boolean = true,
    content: @Composable () -> Unit
) {
    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> darkNexusColorScheme()
        else -> lightNexusColorScheme()
    }

    // Status bar
    val view = LocalView.current
    if (!view.isInEditMode) {
        SideEffect {
            val window = (view.context as Activity).window
            window.statusBarColor = colorScheme.surface.toArgb()
            WindowCompat.getInsetsController(window, view).isAppearanceLightStatusBars = !darkTheme
        }
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = NexusTypography,
        shapes = NexusShapes,
        content = content
    )
}

// ─── Theme Toggle ──────────────────────────────────────────

@Composable
fun ThemeToggle(
    isDarkMode: Boolean,
    onToggle: (Boolean) -> Unit,
    modifier: Modifier = Modifier
) {
    NexusSwitch(
        checked = isDarkMode,
        onCheckedChange = onToggle,
        label = "Dark mode",
        subtitle = if (isDarkMode) "Dark theme enabled" else "Light theme enabled",
        modifier = modifier
    )
}

// ─── Theme Selection ───────────────────────────────────────

@Composable
fun ThemeSelection(
    selectedTheme: ThemeMode,
    onThemeSelected: (ThemeMode) -> Unit
) {
    Column {
        NexusRadio(
            selected = selectedTheme == ThemeMode.LIGHT,
            onClick = { onThemeSelected(ThemeMode.LIGHT) },
            label = "Light"
        )
        NexusRadio(
            selected = selectedTheme == ThemeMode.DARK,
            onClick = { onThemeSelected(ThemeMode.DARK) },
            label = "Dark"
        )
        NexusRadio(
            selected = selectedTheme == ThemeMode.SYSTEM,
            onClick = { onThemeSelected(ThemeMode.SYSTEM) },
            label = "System default"
        )
    }
}
```

---

## 29. High Contrast Mode

### High Contrast Support

```kotlin
@Composable
fun NexusTheme(
    isHighContrast: Boolean = false,
    content: @Composable () -> Unit
) {
    val baseColorScheme = if (isSystemInDarkTheme()) {
        darkNexusColorScheme()
    } else {
        lightNexusColorScheme()
    }

    val colorScheme = if (isHighContrast) {
        baseColorScheme.copy(
            primary = if (isSystemInDarkTheme()) Color(0xFF82B1FF) else Color(0xFF003DA5),
            onPrimary = if (isSystemInDarkTheme()) Color(0xFF001B3F) else Color(0xFFFFFFFF),
            surface = if (isSystemInDarkTheme()) Color(0xFF000000) else Color(0xFFFFFFFF),
            onSurface = if (isSystemInDarkTheme()) Color(0xFFFFFFFF) else Color(0xFF000000),
            background = if (isSystemInDarkTheme()) Color(0xFF000000) else Color(0xFFFFFFFF),
            onBackground = if (isSystemInDarkTheme()) Color(0xFFFFFFFF) else Color(0xFF000000),
            outline = if (isSystemInDarkTheme()) Color(0xFFFFFFFF) else Color(0xFF000000)
        )
    } else {
        baseColorScheme
    }

    MaterialTheme(
        colorScheme = colorScheme,
        content = content
    )
}

// ─── High Contrast Text ────────────────────────────────────

@Composable
fun HighContrastText(
    text: String,
    style: TextStyle,
    modifier: Modifier = Modifier,
    color: Color = MaterialTheme.colorScheme.onSurface
) {
    val isHighContrast = LocalConfiguration.current.fontScale > 1.2f

    Text(
        text = text,
        style = if (isHighContrast) {
            style.copy(fontWeight = FontWeight.Bold)
        } else {
            style
        },
        color = color,
        modifier = modifier
    )
}
```

---

## 30. Font Scaling

### Font Scaling Support

```kotlin
// ─── Font Scaling Configuration ────────────────────────────

@Composable
fun NexusTheme(
    fontScale: Float = LocalConfiguration.current.fontScale,
    content: @Composable () -> Unit
) {
    // Respect system font scaling
    // All text uses sp units which scale automatically

    MaterialTheme(
        content = content
    )
}

// ─── Font Size Selection ───────────────────────────────────

@Composable
fun FontSizeSelection(
    selectedSize: FontSize,
    onSizeSelected: (FontSize) -> Unit
) {
    Column {
        FontSize.entries.forEach { size ->
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { onSizeSelected(size) }
                    .padding(vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                NexusRadio(
                    selected = selectedSize == size,
                    onClick = { onSizeSelected(size) },
                    label = size.displayName
                )

                Spacer(modifier = Modifier.width(16.dp))

                Text(
                    text = "Sample text",
                    fontSize = size.sp,
                    color = MaterialTheme.colorScheme.onSurface
                )
            }
        }
    }
}

enum class FontSize(val sp: TextUnit, val displayName: String) {
    SMALL(12.sp, "Small"),
    MEDIUM(14.sp, "Medium"),
    LARGE(16.sp, "Large"),
    EXTRA_LARGE(18.sp, "Extra Large")
}

// ─── Font Scaling Prevention (if needed) ───────────────────

// Some UI elements should not scale with font size
@Composable
fun FixedSizeText(
    text: String,
    fontSize: TextUnit,
    modifier: Modifier = Modifier
) {
    // Use dp instead of sp for fixed-size text
    Text(
        text = text,
        fontSize = fontSize.value.sp,
        modifier = modifier
    )
}
```

---

## 31. RTL Support

### RTL Layout Configuration

```kotlin
// ─── AndroidManifest.xml ───────────────────────────────────

// android:supportsRtl="true" in Application tag

// ─── RTL-Aware Composables ─────────────────────────────────

@Composable
fun RTLLayout(content: @Composable () -> Unit) {
    CompositionLocalProvider(
        LocalLayoutDirection provides LayoutDirection.Rtl // Force RTL for testing
    ) {
        content()
    }
}

// ─── Direction-Aware Padding ───────────────────────────────

@Composable
fun DirectionAwareLayout(
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit
) {
    val layoutDirection = LocalLayoutDirection.current

    Row(
        modifier = modifier.fillMaxWidth(),
        horizontalArrangement = if (layoutDirection == LayoutDirection.Rtl) {
            Arrangement.End
        } else {
            Arrangement.Start
        }
    ) {
        content()
    }
}

// ─── RTL-Aware Icon ────────────────────────────────────────

@Composable
fun RTLAwareIcon(
    icon: ImageVector,
    contentDescription: String?,
    modifier: Modifier = Modifier,
    mirrored: Boolean = false
) {
    val layoutDirection = LocalLayoutDirection.current
    val isMirrored = mirrored && layoutDirection == LayoutDirection.Rtl

    Icon(
        imageVector = icon,
        contentDescription = contentDescription,
        modifier = modifier.graphicsLayer(
            rotationY = if (isMirrored) 180f else 0f
        )
    )
}

// ─── RTL-Aware Padding Examples ────────────────────────────

// Use start/end instead of left/right
@Composable
fun RTLAwarePadding() {
    Column(
        modifier = Modifier.padding(horizontal = 16.dp) // Uses start/end automatically
    ) {
        Text(
            text = "This text is padded correctly in RTL",
            modifier = Modifier.padding(start = 16.dp) // Start = Right in RTL
        )
    }
}

// ─── Direction-Aware Alignment ─────────────────────────────

@Composable
fun DirectionAwareAlignment() {
    // Start/End aligns correctly in both LTR and RTL
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text("First") // Will be on right in RTL
        Text("Last")  // Will be on left in RTL
    }

    // For specific RTL needs
    val isRtl = LocalLayoutDirection.current == LayoutDirection.Rtl
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = if (isRtl) Arrangement.Start else Arrangement.End
    ) {
        Text("Aligned based on direction")
    }
}

// ─── RTL Checklist ─────────────────────────────────────────

/*
┌──────────────────────────────────────────────────────┐
│              RTL Support Checklist                     │
│                                                      │
│  ✓ Use start/end instead of left/right              │
│  ✓ Use paddingStart/paddingEnd                      │
│  ✓ Use marginStart/marginEnd                        │
│  ✓ Icons that indicate direction should mirror      │
│  ✓ Navigation arrows should mirror                  │
│  ✓ Text alignment should be natural                 │
│  ✓ List items should reverse order                  │
│  ✓ Progress bars should mirror                      │
│  ✓ Swipe gestures should mirror                     │
│  ✓ Test with force RTL developer option             │
│  ✓ android:supportsRtl="true" in manifest           │
└──────────────────────────────────────────────────────┘
*/
```

---

## Summary

| Category | Components | Key Features |
|----------|------------|--------------|
| **Design System** | Tokens, Theme | Material 3, Dynamic Color |
| **Buttons** | Filled, Outlined, Text, Icon, Tonal, FAB | 48dp height, consistent shapes |
| **Cards** | Elevated, Filled, Outlined | Consistent padding, rounded corners |
| **Modals** | Dialog, Bottom Sheet, Full Screen | Accessible, animated |
| **Lists** | ListItem, Section Header, Divider | Consistent spacing |
| **Inputs** | TextField, Search, Dropdown, Checkbox, Radio, Switch | Validation, error states |
| **Feedback** | Snackbar, Progress, Skeleton | Animated, accessible |
| **Navigation** | TopBar, BottomNav, Tabs, Drawer | Material 3 patterns |
| **Chat** | MessageBubble, InputBar, FileAttachment | Offline indicators |
| **Agent** | AgentCard, StatusBadge, ExecutionStep | Live status |
| **Document** | DocumentCard, ProcessingStatus, SyncStatus | Type-based icons |
| **Charts** | Line, Bar, Gauge | Canvas-based |
| **Loading** | Shimmer, Skeleton, Progress | Animated |
| **Empty** | NoData, NoConnection, Error | Actionable |
| **Error** | Network, Server, Permission | Retry support |
| **Animation** | Pulse, Spin, Slide, Fade, Scale | Smooth, configurable |
| **Responsive** | Compact, Medium, Expanded | Adaptive layouts |
| **Foldable** | Two-pane, Navigation Rail | Window size classes |
| **Tablet** | Side-by-side, Navigation Rail | Expanded layout |
| **Dark Mode** | Light, Dark, System | Dynamic Color |
| **Accessibility** | High Contrast, Font Scaling, RTL | WCAG compliant |
