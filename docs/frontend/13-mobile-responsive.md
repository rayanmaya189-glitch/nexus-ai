# Mobile & Responsive Design

> Mobile-first responsive architecture for Nexus AI — breakpoints, touch interactions, offline support, and platform-adaptive layouts.

## 1. Responsive Breakpoints

| Name | Range | Columns | Sidebar | Grid Gap |
|---|---|---|---|---|
| Mobile S | 0–480px | 4 | Hidden | 12px |
| Mobile L | 481–640px | 4 | Hidden | 16px |
| Tablet | 641–1024px | 8 | Collapsible | 20px |
| Desktop | 1025–1440px | 12 | Persistent | 24px |
| Wide | 1441px+ | 12 | Persistent | 24px (max-width container) |

```ts
// tailwind.config.ts (custom breakpoints)
export default {
  theme: {
    screens: {
      xs: "480px",
      sm: "640px",
      md: "768px",
      lg: "1024px",
      xl: "1280px",
      "2xl": "1440px",
    },
  },
};
```

### Breakpoint Visual Reference

```
0          480       640       768       1024      1280      1440
│──────────│─────────│─────────│─────────│─────────│─────────│──────►
│ Mobile S │Mobile L │         Tablet    │         Desktop   │  Wide
│──────────│─────────│─────────│─────────│─────────│─────────│──────►
│ 4 cols   │ 4 cols  │   8 cols         │    12 cols        │ 12 cols
│ No side  │ No side │ Collapsible side │  Persistent side  │
```

## 2. Mobile-First Design Approach

Every component is written mobile-first using Tailwind's prefix system:

```tsx
// components/dashboard/DashboardGrid.tsx
export function DashboardGrid({ widgets }: Props) {
  return (
    <div className="
      grid
      grid-cols-1           {/* Mobile: 1 column */}
      sm:grid-cols-2         {/* Tablet: 2 columns */}
      lg:grid-cols-3         {/* Desktop: 3 columns */}
      xl:grid-cols-4         {/* Wide: 4 columns */}
      gap-4                  {/* Mobile: 16px */}
      sm:gap-5               {/* Tablet+: 20px */}
      p-4                    {/* Mobile: 16px padding */}
      sm:p-6                 {/* Tablet+: 24px padding */}
    ">
      {widgets.map((widget) => (
        <DashboardWidget key={widget.id} widget={widget} />
      ))}
    </div>
  );
}
```

### Layout Composition Strategy

| Layer | Mobile | Tablet | Desktop |
|---|---|---|---|
| Page layout | Stacked full-width | Two-column split | Sidebar + content |
| Navigation | Bottom nav bar | Collapsible sidebar | Persistent sidebar |
| Chat interface | Full-screen overlay | Right panel | Right panel |
| Dashboard | Single column scroll | 2-col grid | 3-4 col grid |
| Modals | Full-screen sheet | Centered modal | Centered modal |
| Tables | Card-based view | Scrollable table | Full table |

## 3. Mobile Navigation

### Bottom Navigation Bar

```tsx
// components/mobile/BottomNav.tsx
"use client";

import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  MessageSquare,
  Users,
  FileText,
  Settings,
} from "lucide-react";

const NAV_ITEMS = [
  { href: "/dashboard", icon: LayoutDashboard, label: "Dashboard" },
  { href: "/chat", icon: MessageSquare, label: "Chat" },
  { href: "/agents", icon: Users, label: "Agents" },
  { href: "/documents", icon: FileText, label: "Documents" },
  { href: "/settings", icon: Settings, label: "Settings" },
];

export function BottomNav() {
  const pathname = usePathname();

  return (
    <nav
      className="fixed bottom-0 inset-x-0 z-40 bg-white dark:bg-neutral-900
                 border-t border-neutral-200 dark:border-neutral-800
                 safe-area-bottom lg:hidden"
      role="navigation"
      aria-label="Main navigation"
    >
      <div className="flex items-center justify-around h-16">
        {NAV_ITEMS.map((item) => {
          const isActive = pathname.startsWith(item.href);
          return (
            <a
              key={item.href}
              href={item.href}
              className={`flex flex-col items-center justify-center gap-0.5
                         min-w-[64px] py-2 rounded-lg transition-colors
                         ${isActive
                           ? "text-blue-600 dark:text-blue-400"
                           : "text-neutral-500 dark:text-neutral-400"
                         }`}
              aria-current={isActive ? "page" : undefined}
            >
              <item.icon className="h-5 w-5" />
              <span className="text-[10px] font-medium">{item.label}</span>
            </a>
          );
        })}
      </div>
    </nav>
  );
}
```

### Hamburger Menu

```tsx
// components/mobile/HamburgerMenu.tsx
export function HamburgerMenu() {
  const [isOpen, setIsOpen] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);

  useFocusTrap(panelRef, isOpen);

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="lg:hidden p-2 rounded-lg hover:bg-neutral-100 dark:hover:bg-neutral-800"
        aria-label="Open navigation menu"
        aria-expanded={isOpen}
      >
        <Menu className="h-5 w-5" />
      </button>

      <AnimatePresence>
        {isOpen && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 z-50 lg:hidden"
              onClick={() => setIsOpen(false)}
            />
            <motion.div
              ref={panelRef}
              initial={{ x: "-100%" }}
              animate={{ x: 0 }}
              exit={{ x: "-100%" }}
              transition={{ type: "spring", damping: 25, stiffness: 300 }}
              className="fixed inset-y-0 left-0 z-50 w-72
                         bg-white dark:bg-neutral-900 shadow-xl
                         lg:hidden"
              role="dialog"
              aria-modal="true"
              aria-label="Navigation menu"
            >
              <SidebarContent onClose={() => setIsOpen(false)} />
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
}
```

### Swipe Gestures

```ts
// hooks/useSwipeGesture.ts
interface SwipeHandlers {
  onSwipeLeft?: () => void;
  onSwipeRight?: () => void;
  onSwipeUp?: () => void;
  onSwipeDown?: () => void;
  threshold?: number;
}

export function useSwipeGesture(ref: RefObject<HTMLElement>, handlers: SwipeHandlers) {
  const touchStart = useRef<{ x: number; y: number } | null>(null);
  const threshold = handlers.threshold ?? 50;

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    const handleTouchStart = (e: TouchEvent) => {
      touchStart.current = {
        x: e.touches[0].clientX,
        y: e.touches[0].clientY,
      };
    };

    const handleTouchEnd = (e: TouchEvent) => {
      if (!touchStart.current) return;

      const dx = e.changedTouches[0].clientX - touchStart.current.x;
      const dy = e.changedTouches[0].clientY - touchStart.current.y;

      if (Math.abs(dx) > Math.abs(dy)) {
        if (Math.abs(dx) > threshold) {
          dx > 0 ? handlers.onSwipeRight?.() : handlers.onSwipeLeft?.();
        }
      } else {
        if (Math.abs(dy) > threshold) {
          dy > 0 ? handlers.onSwipeDown?.() : handlers.onSwipeUp?.();
        }
      }
      touchStart.current = null;
    };

    el.addEventListener("touchstart", handleTouchStart, { passive: true });
    el.addEventListener("touchend", handleTouchEnd, { passive: true });

    return () => {
      el.removeEventListener("touchstart", handleTouchStart);
      el.removeEventListener("touchend", handleTouchEnd);
    };
  }, [ref, handlers, threshold]);
}

// Usage: swipe between chat conversations
const chatRef = useRef<HTMLDivElement>(null);
useSwipeGesture(chatRef, {
  onSwipeRight: () => navigateToPreviousConversation(),
  onSwipeLeft: () => navigateToNextConversation(),
});
```

## 4. Mobile Chat Interface

```tsx
// components/mobile/MobileChat.tsx
export function MobileChat() {
  const [inputHeight, setInputHeight] = useState(0);
  const chatContainerRef = useRef<HTMLDivElement>(null);
  const { messages, isStreaming } = useChatMessages();

  // Handle virtual keyboard
  useEffect(() => {
    const viewport = document.querySelector("meta[name=viewport]");
    if (viewport) {
      viewport.setAttribute(
        "content",
        "width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no, viewport-fit=cover"
      );
    }
  }, []);

  return (
    <div className="fixed inset-0 flex flex-col bg-white dark:bg-neutral-950">
      {/* Full-screen header */}
      <header className="flex items-center gap-3 px-4 py-3 border-b safe-area-top">
        <Button variant="ghost" size="sm" asChild>
          <a href="/chat">
            <ChevronLeft className="h-5 w-5" />
          </a>
        </Button>
        <ConversationHeader />
      </header>

      {/* Scrollable messages */}
      <div
        ref={chatContainerRef}
        className="flex-1 overflow-y-auto overscroll-contain px-4"
        style={{ paddingBottom: inputHeight + 16 }}
      >
        {messages.map((msg) => (
          <MobileChatMessage key={msg.id} message={msg} />
        ))}
        {isStreaming && <TypingIndicator />}
      </div>

      {/* Input area — adapts to keyboard */}
      <div className="border-t bg-white dark:bg-neutral-900 safe-area-bottom">
        <ChatInput
          onHeightChange={setInputHeight}
          placeholder="Type a message..."
        />
      </div>
    </div>
  );
}
```

### Keyboard Handling

```ts
// hooks/useVirtualKeyboard.ts
export function useVirtualKeyboard() {
  const [isVisible, setIsVisible] = useState(false);
  const [height, setHeight] = useState(0);

  useEffect(() => {
    const handleVisualViewport = () => {
      if (!window.visualViewport) return;
      const newHeight = window.innerHeight - window.visualViewport.height;
      setIsVisible(newHeight > 0);
      setHeight(newHeight);
    };

    window.visualViewport?.addEventListener("resize", handleVisualViewport);
    window.visualViewport?.addEventListener("scroll", handleVisualViewport);

    return () => {
      window.visualViewport?.removeEventListener("resize", handleVisualViewport);
      window.visualViewport?.removeEventListener("scroll", handleVisualViewport);
    };
  }, []);

  return { isKeyboardVisible: isVisible, keyboardHeight: height };
}
```

## 5. Mobile Dashboard

```tsx
// components/mobile/MobileDashboard.tsx
export function MobileDashboard() {
  return (
    <div className="lg:hidden space-y-4 p-4 pb-24">
      {/* Quick stats — horizontal scroll */}
      <div className="flex gap-3 overflow-x-auto snap-x snap-mandatory -mx-4 px-4
                      scrollbar-hide">
        {quickStats.map((stat) => (
          <div
            key={stat.label}
            className="snap-start flex-shrink-0 w-[calc(50%-6px)] p-4
                       bg-white dark:bg-neutral-900 rounded-xl border"
          >
            <StatCard stat={stat} compact />
          </div>
        ))}
      </div>

      {/* Stacked widget cards */}
      <div className="space-y-4">
        {widgets.map((widget) => (
          <MobileWidgetCard key={widget.id} widget={widget} />
        ))}
      </div>
    </div>
  );
}

// Swipeable card carousel
function SwipeableCardList({ items }: { items: CardItem[] }) {
  const [currentIndex, setCurrentIndex] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);

  useSwipeGesture(containerRef, {
    onSwipeLeft: () => setCurrentIndex((i) => Math.min(i + 1, items.length - 1)),
    onSwipeRight: () => setCurrentIndex((i) => Math.max(i - 1, 0)),
  });

  return (
    <div ref={containerRef} className="relative overflow-hidden">
      <motion.div
        className="flex"
        animate={{ x: `-${currentIndex * 100}%` }}
        transition={{ type: "spring", damping: 25 }}
      >
        {items.map((item) => (
          <div key={item.id} className="flex-shrink-0 w-full px-4">
            <Card>{item.content}</Card>
          </div>
        ))}
      </motion.div>
      {/* Dots indicator */}
      <div className="flex justify-center gap-1.5 mt-3">
        {items.map((_, i) => (
          <span
            key={i}
            className={`h-1.5 rounded-full transition-all ${
              i === currentIndex ? "w-4 bg-blue-500" : "w-1.5 bg-neutral-300"
            }`}
          />
        ))}
      </div>
    </div>
  );
}
```

## 6. Mobile Document Upload

```tsx
// components/mobile/MobileDocumentUpload.tsx
export function MobileDocumentUpload() {
  const [uploadMethod, setUploadMethod] = useState<"camera" | "file" | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleCameraCapture = async () => {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = "image/*";
    input.capture = "environment";
    input.onchange = (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (file) uploadFile(file);
    };
    input.click();
  };

  return (
    <div className="p-4 space-y-4">
      {/* Upload methods */}
      <div className="grid grid-cols-2 gap-3">
        <button
          onClick={handleCameraCapture}
          className="flex flex-col items-center gap-2 p-6 rounded-xl
                     border-2 border-dashed border-neutral-300
                     hover:border-blue-400 transition-colors"
        >
          <Camera className="h-8 w-8 text-neutral-500" />
          <span className="text-sm font-medium">Take Photo</span>
        </button>
        <button
          onClick={() => fileInputRef.current?.click()}
          className="flex flex-col items-center gap-2 p-6 rounded-xl
                     border-2 border-dashed border-neutral-300
                     hover:border-blue-400 transition-colors"
        >
          <Upload className="h-8 w-8 text-neutral-500" />
          <span className="text-sm font-medium">Choose File</span>
        </button>
      </div>

      {/* Drag & drop zone (tablet+) */}
      <div className="hidden sm:block">
        <DropZone onFilesSelected={uploadFiles} />
      </div>

      <input
        ref={fileInputRef}
        type="file"
        multiple
        className="hidden"
        accept=".pdf,.doc,.docx,.txt,.csv,.xlsx"
        onChange={(e) => {
          const files = Array.from(e.target.files ?? []);
          files.forEach(uploadFile);
        }}
      />

      {/* Upload progress */}
      <UploadProgressList uploads={activeUploads} />
    </div>
  );
}
```

## 7. Tablet Layout

```tsx
// components/layout/TabletLayout.tsx
export function TabletLayout({ children }: { children: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-neutral-950">
      {/* Collapsible sidebar */}
      <aside
        className={`fixed inset-y-0 left-0 z-30 transition-all duration-300
                    ${sidebarOpen ? "w-64" : "w-0"} overflow-hidden
                    md:block lg:hidden`}
      >
        <SidebarContent />
      </aside>

      {/* Overlay when sidebar is open */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/30 z-20 md:block lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Main content */}
      <div className={`transition-all duration-300 ${sidebarOpen ? "md:ml-64" : ""}`}>
        <header className="flex items-center gap-3 px-4 py-3 border-b">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setSidebarOpen(!sidebarOpen)}
            aria-label={sidebarOpen ? "Close sidebar" : "Open sidebar"}
          >
            {sidebarOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </Button>
          <HeaderContent />
        </header>

        {/* Two-column content on tablet */}
        <div className="grid grid-cols-2 gap-4 p-4">
          <div className="col-span-2 lg:col-span-1">{children}</div>
          <div className="hidden sm:block lg:hidden">
            <TabletSecondaryPanel />
          </div>
        </div>
      </div>
    </div>
  );
}
```

### Split View (iPad-style)

```tsx
function SplitView() {
  const [ratio, setRatio] = useState(50); // percent

  return (
    <div className="hidden md:flex h-[calc(100vh-64px)]">
      <div
        className="overflow-y-auto border-r"
        style={{ width: `${ratio}%` }}
      >
        <PrimaryPanel />
      </div>
      <div
        className="cursor-col-resize w-1 hover:bg-blue-500 transition-colors"
        onMouseDown={startResize}
        role="separator"
        aria-orientation="vertical"
        aria-valuenow={ratio}
        aria-valuemin={30}
        aria-valuemax={70}
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === "ArrowLeft") setRatio((r) => Math.max(30, r - 5));
          if (e.key === "ArrowRight") setRatio((r) => Math.min(70, r + 5));
        }}
      />
      <div style={{ width: `${100 - ratio}%` }}>
        <SecondaryPanel />
      </div>
    </div>
  );
}
```

## 8. Desktop Layout

```
┌──────────────────────────────────────────────────────────────┐
│ ┌──────────┐ ┌────────────────────────────────────────────┐  │
│ │          │ │  Header: Breadcrumb  Search  🔔  👤        │  │
│ │ Sidebar  │ ├────────────────────────────────────────────┤  │
│ │          │ │                                            │  │
│ │ 📊 Dash  │ │                                            │  │
│ │ 💬 Chat  │ │           Main Content Area                │  │
│ │ 🤖 Agents│ │                                            │  │
│ │ 📄 Docs  │ │                                            │  │
│ │ ⚙️ Work  │ │                                            │  │
│ │ 🔒 Sec   │ │                                            │  │
│ │ 👥 Admin │ │                                            │  │
│ │          │ │                                            │  │
│ │ ──────── │ │                                            │  │
│ │ ⚙️ Sett  │ │                                            │  │
│ └──────────┘ └────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

```tsx
// components/layout/DesktopLayout.tsx
export function DesktopLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="hidden lg:flex h-screen overflow-hidden">
      {/* Persistent sidebar */}
      <aside className="w-64 flex-shrink-0 border-r bg-white dark:bg-neutral-900
                        flex flex-col">
        <SidebarContent />
      </aside>

      {/* Main area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="h-16 flex items-center justify-between px-6
                          border-b bg-white dark:bg-neutral-900">
          <Breadcrumb />
          <div className="flex items-center gap-4">
            <SearchBar />
            <NotificationBell />
            <UserMenu />
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-6">
          {children}
        </main>
      </div>
    </div>
  );
}
```

## 9. Responsive Images

```tsx
// components/ui/ResponsiveImage.tsx
interface ResponsiveImageProps {
  src: string;
  alt: string;
  widths?: number[];
  sizes?: string;
  priority?: boolean;
}

export function ResponsiveImage({
  src,
  alt,
  widths = [320, 640, 960, 1280],
  sizes = "(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw",
  priority = false,
}: ResponsiveImageProps) {
  const srcSet = widths
    .map((w) => `${src}?w=${w}&f=webp ${w}w`)
    .join(", ");

  return (
    <picture>
      <source srcSet={srcSet} sizes={sizes} type="image/webp" />
      <img
        src={`${src}?w=640&f=webp`}
        alt={alt}
        loading={priority ? "eager" : "lazy"}
        decoding="async"
        className="w-full h-auto object-cover"
        fetchPriority={priority ? "high" : "auto"
      }
      />
    </picture>
  );
}
```

### Image Format Support

| Format | Browser Support | Use Case |
|---|---|---|
| WebP | 97%+ | Default for all images |
| AVIF | 90%+ | Hero images, large photos |
| JPEG | 100% | Fallback for older browsers |
| PNG | 100% | Transparency required |

## 10. Responsive Typography

```css
/* globals.css — fluid type scale */
:root {
  --font-xs: clamp(0.694rem, 0.66rem + 0.17vw, 0.8rem);
  --font-sm: clamp(0.833rem, 0.79rem + 0.22vw, 0.96rem);
  --font-base: clamp(1rem, 0.94rem + 0.29vw, 1.15rem);
  --font-lg: clamp(1.2rem, 1.13rem + 0.35vw, 1.38rem);
  --font-xl: clamp(1.44rem, 1.35rem + 0.42vw, 1.66rem);
  --font-2xl: clamp(1.728rem, 1.62rem + 0.51vw, 1.99rem);
  --font-3xl: clamp(2.074rem, 1.94rem + 0.62vw, 2.39rem);
}

body {
  font-size: var(--font-base);
  line-height: 1.6;
}

h1 { font-size: var(--font-3xl); font-weight: 800; line-height: 1.2; }
h2 { font-size: var(--font-2xl); font-weight: 700; line-height: 1.25; }
h3 { font-size: var(--font-xl);  font-weight: 600; line-height: 1.3; }
```

## 11. Touch Targets & Spacing

| Element | Minimum Size | Recommended |
|---|---|---|
| Button (icon) | 44×44px | 48×48px |
| Button (text) | 44px height | 48px height |
| Link in nav | 44px height | 48px height |
| Checkbox | 44×44px | 48×48px |
| Toggle switch | 44×24px | 48×28px |
| Tab | 44px height | 48px height |
| Adjacent targets | 8px gap | 12px gap |

```tsx
// Accessible touch targets
function TouchButton({ children, ...props }) {
  return (
    <button
      className="min-h-[44px] min-w-[44px] flex items-center justify-center
                 active:scale-95 transition-transform"
      {...props}
    >
      {children}
    </button>
  );
}
```

## 12. Viewport & Safe Area

```html
<!-- layouts/layout.tsx meta tags -->
<meta
  name="viewport"
  content="width=device-width, initial-scale=1, maximum-scale=5, viewport-fit=cover"
/>
<meta name="theme-color" content="#ffffff" media="(prefers-color-scheme: light)" />
<meta name="theme-color" content="#0a0a0a" media="(prefers-color-scheme: dark)" />
```

```css
/* Safe area CSS */
.safe-area-top { padding-top: env(safe-area-inset-top); }
.safe-area-bottom { padding-bottom: env(safe-area-inset-bottom); }
.safe-area-left { padding-left: env(safe-area-inset-left); }
.safe-area-right { padding-right: env(safe-area-inset-right); }

/* Notch-aware header */
@supports (padding: env(safe-area-inset-top)) {
  .notch-header {
    padding-top: env(safe-area-inset-top);
  }
}
```

## 13. Offline Mode

```ts
// lib/offline/serviceWorker.ts
const CACHE_NAME = "nexus-ai-v1";
const STATIC_ASSETS = [
  "/",
  "/dashboard",
  "/offline",
  "/icons/icon-192.png",
  "/icons/icon-512.png",
];

// Service worker install
self.addEventListener("install", (event) => {
  (event as ExtendableEvent).waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS))
  );
});

// Network-first for API, cache-first for static
self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);

  if (url.pathname.startsWith("/api/")) {
    // Network first, fallback to cache
    event.respondWith(
      fetch(event.request)
        .then((response) => {
          const clone = response.clone();
          caches.open(CACHE_NAME).then((cache) => cache.put(event.request, clone));
          return response;
        })
        .catch(() => caches.match(event.request))
    );
  } else {
    // Cache first, fallback to network
    event.respondWith(
      caches.match(event.request).then((cached) => cached ?? fetch(event.request))
    );
  }
});
```

### Offline Detection & UI

```ts
// hooks/useOfflineStatus.ts
export function useOfflineStatus() {
  const [isOffline, setIsOffline] = useState(!navigator.onLine);

  useEffect(() => {
    const handleOnline = () => setIsOffline(false);
    const handleOffline = () => setIsOffline(true);

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  return isOffline;
}

// components/OfflineBanner.tsx
export function OfflineBanner() {
  const isOffline = useOfflineStatus();

  if (!isOffline) return null;

  return (
    <div className="fixed top-0 inset-x-0 z-[100] bg-amber-500 text-white
                    text-center text-sm py-2 px-4 safe-area-top">
      You are offline. Some features may be limited.
    </div>
  );
}
```

## 14. Performance on Mobile

| Strategy | Target | Implementation |
|---|---|---|
| Initial JS bundle | <150KB gzipped | Code splitting, dynamic imports |
| First Contentful Paint | <1.5s | SSR + inline critical CSS |
| Largest Contentful Paint | <2.5s | Image optimization, font preload |
| Time to Interactive | <3.5s | Lazy load non-critical routes |
| Total weight | <500KB | Tree shaking, bundle analysis |

```ts
// Dynamic imports for route-based splitting
const Dashboard = lazy(() => import("@/pages/Dashboard"));
const Chat = lazy(() => import("@/pages/Chat"));
const Agents = lazy(() => import("@/pages/Agents"));

// Component-based splitting
const NotificationCenter = lazy(
  () => import("@/components/notifications/NotificationCenter")
);

// Intersection observer for below-fold content
function LazySection({ children }: { children: React.ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) setIsVisible(true); },
      { rootMargin: "200px" }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);

  return <div ref={ref}>{isVisible ? children : null}</div>;
}
```

## 15. Responsive Hook

```ts
// hooks/useResponsive.ts
export function useResponsive() {
  const [breakpoint, setBreakpoint] = useState<"mobile" | "tablet" | "desktop" | "wide">("desktop");

  useEffect(() => {
    const check = () => {
      const w = window.innerWidth;
      if (w <= 640) setBreakpoint("mobile");
      else if (w <= 1024) setBreakpoint("tablet");
      else if (w <= 1440) setBreakpoint("desktop");
      else setBreakpoint("wide");
    };

    check();
    window.addEventListener("resize", check);
    return () => window.removeEventListener("resize", check);
  }, []);

  return {
    breakpoint,
    isMobile: breakpoint === "mobile",
    isTablet: breakpoint === "tablet",
    isDesktop: breakpoint === "desktop",
    isWide: breakpoint === "wide",
    isMobileOrTablet: breakpoint === "mobile" || breakpoint === "tablet",
  };
}
```

## 16. Mobile Testing Strategy

| Tool | Purpose | Priority |
|---|---|---|
| Chrome DevTools Device Mode | Quick responsive checks | High |
| BrowserStack | Real device testing | High |
| Lighthouse (mobile) | Performance & accessibility | High |
| Xcode Simulator | iOS-specific testing | Medium |
| Android Emulator | Android-specific testing | Medium |
| Playwright mobile viewport | Automated E2E tests | High |

### Automated Responsive Tests

```ts
// e2e/responsive.spec.ts
import { test, expect } from "@playwright/test";

const VIEWPORTS = [
  { name: "iPhone SE", width: 375, height: 667 },
  { name: "iPhone 14", width: 390, height: 844 },
  { name: "iPad", width: 810, height: 1080 },
  { name: "Desktop", width: 1280, height: 720 },
  { name: "Wide", width: 1920, height: 1080 },
];

for (const vp of VIEWPORTS) {
  test(`layout renders correctly on ${vp.name}`, async ({ page }) => {
    await page.setViewportSize({ width: vp.width, height: vp.height });
    await page.goto("/dashboard");

    // Verify appropriate layout
    if (vp.width < 640) {
      await expect(page.locator("[data-testid=bottom-nav]")).toBeVisible();
      await expect(page.locator("[data-testid=sidebar]")).not.toBeVisible();
    } else if (vp.width < 1024) {
      await expect(page.locator("[data-testid=hamburger]")).toBeVisible();
    } else {
      await expect(page.locator("[data-testid=sidebar]")).toBeVisible();
    }
  });
}
```

## 17. File Structure

```
src/
├── components/
│   ├── layout/
│   │   ├── DesktopLayout.tsx
│   │   ├── TabletLayout.tsx
│   │   └── MobileLayout.tsx
│   ├── mobile/
│   │   ├── BottomNav.tsx
│   │   ├── HamburgerMenu.tsx
│   │   ├── MobileChat.tsx
│   │   ├── MobileDashboard.tsx
│   │   ├── MobileDocumentUpload.tsx
│   │   ├── MobileNotificationPanel.tsx
│   │   └── SwipeableCardList.tsx
│   └── ui/
│       ├── ResponsiveImage.tsx
│       └── TouchButton.tsx
├── hooks/
│   ├── useResponsive.ts
│   ├── useSwipeGesture.ts
│   ├── useVirtualKeyboard.ts
│   ├── useOfflineStatus.ts
│   └── useSafeArea.ts
├── lib/
│   ├── offline/
│   │   ├── serviceWorker.ts
│   │   └── cache.ts
│   └── push/
│       ├── fcm.ts
│       └── apns.ts
└── styles/
    └── globals.css          (fluid type, safe areas)
```
