# Performance

> Performance engineering for Nexus AI — bundle optimization, caching, rendering strategies, and monitoring.

## 1. Performance Targets

| Metric | Target | Measurement |
|---|---|---|
| First Contentful Paint (FCP) | < 1.5s | Lighthouse |
| Largest Contentful Paint (LCP) | < 2.5s | Web Vitals |
| First Input Delay (FID) | < 100ms | Web Vitals |
| Cumulative Layout Shift (CLS) | < 0.1 | Web Vitals |
| Time to First Byte (TTFB) | < 200ms | Network |
| Initial JS Bundle | < 150KB gzip | Bundle analyzer |
| Total Page Weight | < 500KB | Network |
| API Response (p95) | < 500ms | APM |
| WebSocket Message Latency | < 100ms | Custom metric |
| Chat First Token | < 2s | Streaming metric |

## 2. Bundle Optimization

### Bundle Size Budget

| Category | Budget (gzip) | Current | Status |
|---|---|---|---|
| Framework (React, Next.js) | < 45KB | ~42KB | ✅ |
| UI Components | < 30KB | ~28KB | ✅ |
| State Management | < 10KB | ~8KB | ✅ |
| Router | < 5KB | ~4KB | ✅ |
| Utilities | < 15KB | ~12KB | ✅ |
| Charts/Visualization | < 20KB | ~18KB | ✅ |
| Vendor (3rd party) | < 25KB | ~22KB | ✅ |
| **Total** | **< 150KB** | **~134KB** | ✅ |

### Webpack/Turbopack Configuration

```ts
// next.config.ts
const config: NextConfig = {
  // Enable Turbopack for dev
  experimental: {
    turbo: true,
  },

  // Production optimizations
  webpack: (config, { isServer }) => {
    // Tree shaking
    config.optimization = {
      ...config.optimization,
      usedExports: true,
      sideEffects: true,
      splitChunks: {
        chunks: "all",
        maxInitialRequests: 25,
        minSize: 20000,
        cacheGroups: {
          vendor: {
            test: /[\\/]node_modules[\\/]/,
            name(module) {
              const packageName = module.context.match(
                /[\\/]node_modules[\\/](.*?)([\\/]|$)/
              )?.[1];
              return `vendor.${packageName?.replace("@", "")}`;
            },
          },
          charts: {
            test: /[\\/]node_modules[\\/](recharts|d3|plotly)/,
            name: "charts",
            priority: 10,
          },
          icons: {
            test: /[\\/]node_modules[\\/](@?lucide|@heroicons)/,
            name: "icons",
          },
        },
      },
    };

    // Replace large libraries
    config.resolve.alias = {
      ...config.resolve.alias,
      "date-fns": "date-fns", // only import needed functions
    };

    return config;
  },

  // Headers for caching
  async headers() {
    return [
      {
        source: "/_next/static/(.*)",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },
      {
        source: "/fonts/(.*)",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },
    ];
  },
};
```

## 3. Code Splitting Strategies

### Route-Based Splitting

```tsx
// app/(dashboard)/layout.tsx — dashboard routes chunk
import dynamic from "next/dynamic";

const Dashboard = dynamic(() => import("@/pages/Dashboard"), {
  loading: () => <DashboardSkeleton />,
});

const Chat = dynamic(() => import("@/pages/Chat"), {
  loading: () => <ChatSkeleton />,
});

const Agents = dynamic(() => import("@/pages/Agents"), {
  loading: () => <AgentListSkeleton />,
});
```

### Component-Based Splitting

```tsx
// Lazy load heavy components
const RichTextEditor = dynamic(() => import("@/components/RichTextEditor"), {
  loading: () => <div className="h-64 bg-neutral-100 rounded-lg animate-pulse" />,
  ssr: false,
});

const CodeEditor = dynamic(() => import("@/components/CodeEditor"), {
  loading: () => <EditorSkeleton />,
  ssr: false,
});

const PDFViewer = dynamic(() => import("@/components/PDFViewer"), {
  loading: () => <div className="h-[600px] bg-neutral-100 rounded-lg animate-pulse" />,
  ssr: false,
});

const DataVisualization = dynamic(() => import("@/components/charts/DataVisualization"), {
  loading: () => <ChartSkeleton />,
});
```

### Model-Based Splitting (AI model weights)

```ts
// Load tokenizer/model weights only when chat is opened
export async function loadChatTokenizer() {
  const { AutoTokenizer } = await import("@huggingface/transformers");
  return AutoTokenizer.from_pretrained("Xenova/gpt2-tokenizer");
}

// Load document parser only when document view opens
export async function loadPDFParser() {
  const pdfjs = await import("pdfjs-dist");
  pdfjs.GlobalWorkerOptions.workerSrc = `//cdnjs.cloudflare.com/ajax/libs/pdf.js/${pdfjs.version}/pdf.worker.min.js`;
  return pdfjs;
}
```

## 4. Lazy Loading

### React.lazy with Suspense

```tsx
// App.tsx / layout.tsx
import { Suspense, lazy } from "react";
import { PageSkeleton } from "@/components/skeletons/PageSkeleton";

const DashboardPage = lazy(() => import("@/pages/dashboard"));
const ChatPage = lazy(() => import("@/pages/chat"));
const AgentsPage = lazy(() => import("@/pages/agents"));

export function AppRoutes() {
  return (
    <Suspense fallback={<PageSkeleton />}>
      <Routes>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/chat/*" element={<ChatPage />} />
        <Route path="/agents/*" element={<AgentsPage />} />
      </Routes>
    </Suspense>
  );
}
```

### Image Lazy Loading

```tsx
// components/ui/LazyImage.tsx
import { useRef, useState, useEffect } from "react";

interface LazyImageProps {
  src: string;
  alt: string;
  width?: number;
  height?: number;
  className?: string;
  placeholder?: string;
}

export function LazyImage({
  src,
  alt,
  width,
  height,
  className,
  placeholder = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzIwIiBoZWlnaHQ9IjI0MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCBmaWxsPSIjZTBlMGUwIiB3aWR0aD0iMzIwIiBoZWlnaHQ9IjI0MCIvPjwvc3ZnPg==",
}: LazyImageProps) {
  const imgRef = useRef<HTMLImageElement>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInView, setIsInView] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true);
          observer.disconnect();
        }
      },
      { rootMargin: "200px" }
    );
    if (imgRef.current) observer.observe(imgRef.current);
    return () => observer.disconnect();
  }, []);

  return (
    <img
      ref={imgRef}
      src={isInView ? src : placeholder}
      alt={alt}
      width={width}
      height={height}
      loading="lazy"
      decoding="async"
      onLoad={() => setIsLoaded(true)}
      className={`transition-opacity duration-300 ${
        isLoaded ? "opacity-100" : "opacity-0"
      } ${className}`}
    />
  );
}
```

## 5. Image Optimization

```ts
// next.config.ts — image optimization
const config: NextConfig = {
  images: {
    formats: ["image/avif", "image/webp"],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
    minimumCacheTTL: 60 * 60 * 24 * 365, // 1 year
    remotePatterns: [
      {
        protocol: "https",
        hostname: "cdn.nexus-ai.com",
        pathname: "/images/**",
      },
    ],
  },
};
```

| Format | Size Reduction | Use Case |
|---|---|---|
| Original JPEG | Baseline | Source |
| WebP | ~30% smaller | Default output |
| AVIF | ~50% smaller | Hero, large images |
| Responsive srcset | Per-device optimal | All images |

## 6. Font Optimization

```tsx
// app/layout.tsx
import { Inter, JetBrains_Mono } from "next/font/google";

const inter = Inter({
  subsets: ["latin"],
  display: "swap", // Show fallback immediately
  preload: true,
  fallback: ["system-ui", "-apple-system", "sans-serif"],
  adjustFontFallback: true,
  variable: "--font-inter",
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  display: "swap",
  preload: false, // Only preload primary font
  variable: "--font-mono",
});

export default function RootLayout({ children }) {
  return (
    <html lang="en" className={`${inter.variable} ${jetbrainsMono.variable}`}>
      <head>
        {/* Preload critical font files */}
        <link rel="preload" href={inter.src} as="font" crossOrigin="anonymous" />
      </head>
      <body className="font-sans">{children}</body>
    </html>
  );
}
```

### Font Loading Strategy

| Strategy | Use Case |
|---|---|
| `display: swap` | All fonts (avoid invisible text) |
| `preload` | Primary body font only |
| `font-display: optional` | Non-critical decorative fonts |
| System font stack | Fallback chain |

## 7. Caching Strategies

```
┌─────────────────────────────────────────────────────────────┐
│                    CACHING LAYERS                            │
│                                                             │
│  ┌─────────────┐   ┌──────────────┐   ┌─────────────────┐  │
│  │   Browser    │   │  Service     │   │    CDN          │  │
│  │   Cache      │   │  Worker      │   │    (Cloudflare) │  │
│  │              │   │  Cache       │   │                 │  │
│  │  HTTP Cache  │   │  Network-first│  │  Edge cache     │  │
│  │  Static: 1yr │   │  API-first   │   │  Static: 1yr    │  │
│  │  HTML: 0     │   │  Fallback    │   │  Pages: 60s     │  │
│  └─────────────┘   └──────────────┘   └─────────────────┘  │
│                                                             │
│  ┌─────────────┐   ┌──────────────┐                        │
│  │  React      │   │  IndexedDB   │                        │
│  │  Query      │   │  (offline)   │                        │
│  │  Cache      │   │              │                        │
│  │  Stale-while│   │  Documents   │                        │
│  │  -revalidate│   │  Chat history│                        │
│  └─────────────┘   └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

### HTTP Cache Headers

| Resource | Cache-Control | CDN Cache |
|---|---|---|
| Static assets (`/_next/static`) | `public, max-age=31536000, immutable` | Yes |
| Images (`/images`) | `public, max-age=31536000, immutable` | Yes |
| API responses (`/api`) | `no-cache, must-revalidate` | No |
| HTML pages | `public, s-maxage=60, stale-while-revalidate=300` | Yes (60s) |
| Fonts | `public, max-age=31536000, immutable` | Yes |
| WebSocket | N/A | No |

### Stale-While-Revalidate Pattern

```ts
// TanStack Query config
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,           // 30s before considered stale
      gcTime: 5 * 60_000,          // 5min before garbage collected
      refetchOnWindowFocus: true,
      refetchOnReconnect: true,
    },
  },
});

// Stale-while-revalidate for specific queries
function useDashboardMetrics() {
  return useQuery({
    queryKey: ["dashboard-metrics"],
    queryFn: fetchDashboardMetrics,
    staleTime: 10_000,        // 10s
    gcTime: 60_000,           // 1min
    refetchInterval: 30_000,  // Poll every 30s
  });
}
```

## 8. ISR & SSR

### Incremental Static Regeneration

```ts
// app/dashboard/page.tsx
export const revalidate = 60; // Revalidate every 60 seconds

export default async function DashboardPage() {
  const metrics = await fetchDashboardMetrics(); // Cached at edge
  return <Dashboard metrics={metrics} />;
}

// app/agents/[id]/page.tsx
export const revalidate = 300; // 5 minutes for agent detail pages

export async function generateStaticParams() {
  const agents = await fetchAgentIds();
  return agents.map((id) => ({ id }));
}
```

### SSR for Dynamic Pages

```ts
// app/chat/[conversationId]/page.tsx — always fresh
export const dynamic = "force-dynamic";
export const runtime = "edge"; // Edge runtime for low latency

export default async function ChatPage({ params }) {
  const conversation = await fetchConversation(params.conversationId);
  return <Chat conversation={conversation} />;
}
```

### Render Strategy Matrix

| Page | Strategy | Revalidation | Reason |
|---|---|---|---|
| Landing | SSG | On deploy | Static content |
| Dashboard | ISR | 60s | Semi-dynamic metrics |
| Agent List | ISR | 120s | Changes infrequently |
| Agent Detail | ISR | 300s | Rarely changes |
| Chat | SSR (Edge) | N/A | Always fresh |
| Documents | ISR | 60s | Moderate update rate |
| Settings | SSR | N/A | User-specific data |
| Admin | SSR | N/A | Always fresh, auth-gated |

## 9. Preloading & Prefetching

```tsx
// Link prefetching (Next.js)
import Link from "next/link";

<Link href="/dashboard" prefetch={true}>
  Dashboard
</Link>

// Route prefetching on hover
<Link href="/agents" prefetch={true}>Agents</Link>

// DNS prefetch for external services
<head>
  <link rel="dns-prefetch" href="https://api.nexus-ai.com" />
  <link rel="dns-prefetch" href="https://cdn.nexus-ai.com" />
  <link rel="preconnect" href="https://api.nexus-ai.com" crossOrigin="anonymous" />
</head>

// Preload critical resources
<link rel="preload" href="/fonts/Inter-Bold.woff2" as="font" type="font/woff2" crossOrigin="anonymous" />
<link rel="preload" href="/api/dashboard/metrics" as="fetch" crossOrigin="anonymous" />
```

## 10. Virtual Scrolling

```tsx
// components/ui/VirtualList.tsx
import { FixedSizeList as List } from "react-window";

interface VirtualListProps<T> {
  items: T[];
  height: number;
  itemHeight: number;
  renderItem: (props: { index: number; style: React.CSSProperties; data: T }) => React.ReactNode;
}

export function VirtualList<T>({
  items,
  height,
  itemHeight,
  renderItem,
}: VirtualListProps<T>) {
  return (
    <List
      height={height}
      itemCount={items.length}
      itemSize={itemHeight}
      width="100%"
      itemData={items}
      overscanCount={5}
    >
      {renderItem}
    </List>
  );
}

// Usage: virtual notification list
<VirtualList
  items={notifications}
  height={480}
  itemHeight={72}
  renderItem={({ index, style, data }) => (
    <div style={style}>
      <NotificationItem notification={data[index]} />
    </div>
  )}
/>

// Usage: virtual chat history (thousands of messages)
<VirtualList
  items={messages}
  height={containerHeight}
  itemHeight={80}
  renderItem={({ index, style, data }) => (
    <div style={style}>
      <ChatMessage message={data[index]} />
    </div>
  )}
/>
```

## 11. Memoization

```tsx
// React.memo for expensive renders
const ExpensiveChart = React.memo(function ExpensiveChart({ data, options }) {
  return <HighchartsReact options={buildChartOptions(data, options)} />;
});

// useMemo for computed values
function FilteredList({ items, query, category }: Props) {
  const filteredItems = useMemo(() => {
    return items.filter(
      (item) =>
        item.name.includes(query) &&
        (category === "all" || item.category === category)
    );
  }, [items, query, category]); // Only recompute when these change

  return <List items={filteredItems} />;
}

// useCallback for stable references
function Parent() {
  const handleClick = useCallback((id: string) => {
    console.log("Clicked:", id);
  }, []);

  return (
    <>
      {/* handleClick reference is stable — Child won't re-render */}
      {items.map((item) => (
        <Child key={item.id} item={item} onClick={handleClick} />
      ))}
    </>
  );
}
```

## 12. Core Web Vitals

```ts
// lib/performance/webVitals.ts
import { onLCP, onFID, onCLS, onFCP, onTTFB } from "web-vitals";
import { useMetricsStore } from "@/stores/performanceStore";

function sendToAnalytics(metric) {
  const body = JSON.stringify({
    name: metric.name,
    value: metric.value,
    rating: metric.rating, // "good" | "needs-improvement" | "poor"
    delta: metric.delta,
    id: metric.id,
    navigationType: metric.navigationType,
    url: window.location.href,
    userAgent: navigator.userAgent,
  });

  // Use sendBeacon for non-blocking analytics
  if (navigator.sendBeacon) {
    navigator.sendBeacon("/api/metrics/vitals", body);
  } else {
    fetch("/api/metrics/vitals", { body, method: "POST", keepalive: true });
  }

  // Update local store for dashboard display
  useMetricsStore.getState().addMetric({
    name: metric.name,
    value: metric.value,
    rating: metric.rating,
    timestamp: Date.now(),
  });
}

export function initWebVitals() {
  onLCP(sendToAnalytics);
  onFID(sendToAnalytics);
  onCLS(sendToAnalytics);
  onFCP(sendToAnalytics);
  onTTFB(sendToAnalytics);
}
```

### Web Vitals Targets

| Metric | Good | Needs Improvement | Poor |
|---|---|---|---|
| LCP | ≤ 2.5s | ≤ 4.0s | > 4.0s |
| FID | ≤ 100ms | ≤ 300ms | > 300ms |
| CLS | ≤ 0.1 | ≤ 0.25 | > 0.25 |
| FCP | ≤ 1.8s | ≤ 3.0s | > 3.0s |
| TTFB | ≤ 800ms | ≤ 1800ms | > 1800ms |

## 13. Performance Monitoring

```ts
// hooks/usePerformance.ts
export function usePerformance() {
  const { addMetric } = useMetricsStore();

  const measureApiCall = useCallback(
    async <T>(name: string, fn: () => Promise<T>): Promise<T> => {
      const start = performance.now();
      try {
        const result = await fn();
        const duration = performance.now() - start;
        addMetric({ name, value: duration, rating: duration < 500 ? "good" : "poor", timestamp: Date.now() });
        return result;
      } catch (error) {
        const duration = performance.now() - start;
        addMetric({ name, value: duration, rating: "poor", timestamp: Date.now() });
        throw error;
      }
    },
    [addMetric]
  );

  const measureRender = useCallback(
    (name: string) => {
      const start = performance.now();
      return () => {
        const duration = performance.now() - start;
        addMetric({ name, value: duration, rating: duration < 16 ? "good" : "poor", timestamp: Date.now() });
      };
    },
    [addMetric]
  );

  return { measureApiCall, measureRender };
}

// Usage
function AgentList() {
  const { measureApiCall } = usePerformance();

  useEffect(() => {
    measureApiCall("fetch-agents", () => fetch("/api/agents"));
  }, []);
}
```

### Performance Dashboard

```
┌──────────────────────────────────────────────────────┐
│ Performance Dashboard                                 │
├──────────────────────────────────────────────────────┤
│ Core Web Vitals                                      │
│ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐         │
│ │  LCP   │ │  FID   │ │  CLS   │ │  FCP   │         │
│ │ 1.2s   │ │ 45ms   │ │ 0.03   │ │ 0.8s   │         │
│ │  ✅    │ │  ✅    │ │  ✅    │ │  ✅    │         │
│ └────────┘ └────────┘ └────────┘ └────────┘         │
│                                                      │
│ API Performance (p95)                                │
│ ┌──────────────────────────────────────────────┐    │
│ │ /api/agents          ████████░░  320ms       │    │
│ │ /api/chat/messages   ██████████  480ms       │    │
│ │ /api/documents       ██████░░░░  210ms       │    │
│ │ /api/models          ███░░░░░░░  120ms       │    │
│ └──────────────────────────────────────────────┘    │
│                                                      │
│ Bundle Size                                          │
│ Main: 134KB │ Chunks: 42 │ Lazy: 12                 │
└──────────────────────────────────────────────────────┘
```

## 14. Performance Stores

```ts
// stores/performanceStore.ts
import { create } from "zustand";

interface Metric {
  name: string;
  value: number;
  rating: "good" | "needs-improvement" | "poor";
  timestamp: number;
}

interface PerformanceState {
  metrics: Metric[];
  budgets: Record<string, number>;
  violations: string[];
}

interface PerformanceActions {
  addMetric: (metric: Metric) => void;
  checkBudget: (name: string, value: number) => boolean;
  clearMetrics: () => void;
}

const DEFAULT_BUDGETS: Record<string, number> = {
  LCP: 2500,
  FID: 100,
  CLS: 0.1,
  FCP: 1800,
  "api-response": 500,
  "chat-first-token": 2000,
  "websocket-latency": 100,
  "render-time": 16,
};

export const useMetricsStore = create<PerformanceState & PerformanceActions>()(
  (set, get) => ({
    metrics: [],
    budgets: DEFAULT_BUDGETS,
    violations: [],

    addMetric: (metric) =>
      set((state) => {
        const newMetrics = [...state.metrics, metric].slice(-200); // Keep last 200
        const budget = state.budgets[metric.name];
        const violations = budget && metric.value > budget
          ? [...state.violations, `${metric.name}: ${metric.value} > ${budget}`]
          : state.violations;
        return { metrics: newMetrics, violations };
      }),

    checkBudget: (name, value) => {
      const budget = get().budgets[name];
      return budget ? value <= budget : true;
    },

    clearMetrics: () => set({ metrics: [], violations: [] }),
  })
);
```

## 15. Database Query Optimization

```ts
// lib/api/queryOptimization.ts

// ❌ N+1 problem
async function getAgentsWithMetricsBad() {
  const agents = await db.agent.findMany(); // 1 query
  for (const agent of agents) {
    agent.metrics = await db.agentMetrics.findMany({ // N queries
      where: { agentId: agent.id },
    });
  }
  return agents;
}

// ✅ Batched query
async function getAgentsWithMetricsGood() {
  const agents = await db.agent.findMany({
    include: {
      metrics: {
        orderBy: { timestamp: "desc" },
        take: 10,
      },
    },
  });
  return agents;
}

// ✅ DataLoader pattern for GraphQL/batched APIs
import DataLoader from "dataloader";

const metricsLoader = new DataLoader(async (agentIds: string[]) => {
  const metrics = await db.agentMetrics.findMany({
    where: { agentId: { in: agentIds } },
    orderBy: { timestamp: "desc" },
  });
  return agentIds.map((id) => metrics.filter((m) => m.agentId === id));
});
```

## 16. Streaming Optimization

```ts
// lib/api/streaming.ts
export async function streamChatResponse(
  messages: Message[],
  onChunk: (chunk: string) => void,
  signal?: AbortSignal
) {
  const response = await fetch("/api/chat", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ messages }),
    signal,
  });

  const reader = response.body!.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });

    // Process complete SSE events
    const lines = buffer.split("\n");
    buffer = lines.pop() ?? "";

    for (const line of lines) {
      if (line.startsWith("data: ")) {
        const data = line.slice(6);
        if (data === "[DONE]") return;

        try {
          const parsed = JSON.parse(data);
          onChunk(parsed.content ?? "");
        } catch {
          // Skip malformed events
        }
      }
    }
  }
}
```

## 17. Performance Testing

```ts
// e2e/performance.spec.ts
import { test, expect } from "@playwright/test";

test("dashboard loads within performance budget", async ({ page }) => {
  const startTime = Date.now();

  await page.goto("/dashboard");
  await page.waitForLoadState("networkidle");

  const loadTime = Date.now() - startTime;
  expect(loadTime).toBeLessThan(3000); // 3s budget

  // Check LCP
  const lcp = await page.evaluate(() => {
    return new Promise<number>((resolve) => {
      new PerformanceObserver((list) => {
        const entries = list.getEntries();
        resolve(entries[entries.length - 1].startTime);
      }).observe({ type: "largest-contentful-paint", buffered: true });
      setTimeout(() => resolve(0), 5000);
    });
  });

  expect(lcp).toBeLessThan(2500);

  // Check CLS
  const cls = await page.evaluate(() => {
    return new Promise<number>((resolve) => {
      let score = 0;
      new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (!(entry as any).hadRecentInput) {
            score += (entry as any).value;
          }
        }
        resolve(score);
      }).observe({ type: "layout-shift", buffered: true });
      setTimeout(() => resolve(score), 3000);
    });
  });

  expect(cls).toBeLessThan(0.1);
});

test("chat streaming first token latency", async ({ page }) => {
  await page.goto("/chat");
  const startTime = Date.now();

  await page.fill("textarea", "Hello, how are you?");
  await page.click("button[type=submit]");

  await page.waitForSelector("[data-testid=streaming-message]", { timeout: 2000 });
  const firstTokenLatency = Date.now() - startTime;

  expect(firstTokenLatency).toBeLessThan(2000);
});
```

### k6 Load Test

```ts
// perf/load-test.js
import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 20 },   // Ramp up
    { duration: "1m", target: 50 },     // Sustained load
    { duration: "30s", target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"],
    http_req_failed: ["rate<0.01"],
  },
};

export default function () {
  const res = http.get("https://api.nexus-ai.com/api/agents");
  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time < 500ms": (r) => r.timings.duration < 500,
  });
  sleep(1);
}
```

## 18. Performance Hooks & Stores Summary

| Hook / Store | Purpose |
|---|---|
| `usePerformance` | Measure API calls, render times |
| `useMetricsStore` | Store/display performance metrics |
| `useLazyLoad` | Intersection observer lazy loading |
| `useVirtualScroll` | Virtual list management |
| `useDebounce` | Debounce expensive computations |
| `useThrottle` | Throttle event handlers |
| `useIntersection` | Track element visibility |

## 19. File Structure

```
src/
├── lib/
│   └── performance/
│       ├── webVitals.ts
│       ├── queryOptimization.ts
│       ├── streaming.ts
│       └── budgets.ts
├── hooks/
│   ├── usePerformance.ts
│   ├── useLazyLoad.ts
│   ├── useVirtualScroll.ts
│   └── useIntersection.ts
├── stores/
│   └── performanceStore.ts
├── components/
│   └── ui/
│       ├── LazyImage.tsx
│       ├── VirtualList.tsx
│       └── LazySection.tsx
├── e2e/
│   └── performance.spec.ts
└── perf/
    └── load-test.js
```
