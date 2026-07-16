# AeroXe Nexus AI — Frontend Architecture Overview

## Next.js 16+, TypeScript, shadcn/ui, Zustand, TanStack Query — Complete Architecture

---

## 1. Platform Identity

| Attribute | Value |
|---|---|
| Product | AeroXe Nexus AI |
| Domain | aeroxenexus.com |
| Category | Enterprise Agentic AI Platform |
| Frontend Framework | Next.js 16+ (App Router) |
| Language | TypeScript 5.5+ |
| Styling | Tailwind CSS 4 |
| UI Library | shadcn/ui |
| Animation | Framer Motion |
| State Management | Zustand |
| Data Fetching | TanStack Query v5 |
| Forms | React Hook Form + Zod |
| Deployment | Vercel / Docker SSR |

---

## 2. System Architecture Diagram

```
                    ┌──────────────────────────────┐
                    │       Browser / Mobile        │
                    └──────────────┬───────────────┘
                                   │
                         HTTPS + WebSocket
                                   │
                    ┌──────────────▼───────────────┐
                    │     Next.js Application       │
                    │     ─────────────────────     │
                    │  SSR / SSG / ISR / CSR        │
                    │  App Router (Edge + Node)     │
                    └──────────────┬───────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                     │
    ┌─────────▼──────┐  ┌─────────▼──────┐  ┌─────────▼──────┐
    │  Nexus API GW  │  │  WebSocket GW  │  │  Static Assets │
    │  (REST + gRPC) │  │  (Real-time)   │  │  (CDN / S3)    │
    └────────────────┘  └────────────────┘  └────────────────┘
              │                    │
    ┌─────────▼────────────────────▼────────────────────┐
    │              Backend Microservices                  │
    │  identity-service │ agent-orchestrator │ rag-svc   │
    │  vision-service   │ sql-agent-service  │ memory-svc│
    └───────────────────────────────────────────────────┘
```

---

## 3. Tech Stack

### 3.1 Core Framework

| Technology | Version | Purpose |
|---|---|---|
| Next.js | 16+ | React meta-framework (SSR, SSG, ISR, API routes) |
| React | 19 | UI library with Server Components, Suspense, transitions |
| TypeScript | 5.5+ | Type safety across the entire codebase |
| Node.js | 22 LTS | Server runtime for SSR and API routes |

### 3.2 Styling & UI

| Technology | Purpose |
|---|---|
| Tailwind CSS 4 | Utility-first CSS framework |
| shadcn/ui | Pre-built accessible component library (Radix + Tailwind) |
| Framer Motion | Animation library for React |
| Lucide React | Icon library (1000+ icons) |
| clsx + tailwind-merge | Conditional class merging utility |

### 3.3 State Management

| Technology | Purpose |
|---|---|
| Zustand | Lightweight global state (auth, chat, theme, UI) |
| TanStack Query v5 | Server state, caching, background refetch |
| React Hook Form | Form state management |
| Zod | Schema validation (forms, API responses) |

### 3.4 Real-Time

| Technology | Purpose |
|---|---|
| WebSocket (native) | Chat streaming, live metrics |
| Server-Sent Events | Fallback for streaming responses |

### 3.5 Build & Deploy

| Technology | Purpose |
|---|---|
| Vercel | Primary deployment (edge functions, ISR) |
| Docker | Self-hosted / air-gapped deployment |
| Turbopack | Next.js dev server (faster HMR) |
| pnpm | Package manager |

---

## 4. Project Structure

```
nexus-ai-web/
├── public/
│   ├── favicon.ico
│   ├── logo.svg
│   ├── logos/
│   │   ├── nexus-logo-light.svg
│   │   ├── nexus-logo-dark.svg
│   │   └── tenant-logo-placeholder.svg
│   └── robots.txt
│
├── src/
│   ├── app/                          # App Router (routes)
│   │   ├── (auth)/                   # Auth route group (no layout)
│   │   │   ├── login/
│   │   │   │   └── page.tsx
│   │   │   ├── register/
│   │   │   │   └── page.tsx
│   │   │   ├── forgot-password/
│   │   │   │   └── page.tsx
│   │   │   ├── reset-password/
│   │   │   │   └── page.tsx
│   │   │   └── otp-verify/
│   │   │       └── page.tsx
│   │   │
│   │   ├── (dashboard)/              # Dashboard route group
│   │   │   ├── layout.tsx            # Dashboard sidebar + header
│   │   │   ├── page.tsx              # Main dashboard
│   │   │   ├── agents/
│   │   │   │   ├── page.tsx          # Agent list
│   │   │   │   ├── new/
│   │   │   │   │   └── page.tsx      # Create agent
│   │   │   │   └── [id]/
│   │   │   │       ├── page.tsx      # Agent detail
│   │   │   │       ├── configure/
│   │   │   │       │   └── page.tsx  # Agent config
│   │   │   │       ├── execute/
│   │   │   │       │   └── page.tsx  # Execute agent
│   │   │   │       └── history/
│   │   │   │           └── page.tsx  # Execution history
│   │   │   ├── chat/
│   │   │   │   ├── page.tsx          # New chat
│   │   │   │   └── [conversationId]/
│   │   │   │       └── page.tsx      # Resume chat
│   │   │   ├── documents/
│   │   │   │   ├── page.tsx          # Document list
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx      # Document detail
│   │   │   ├── knowledge/
│   │   │   │   ├── page.tsx          # Knowledge base
│   │   │   │   └── sets/
│   │   │   │       ├── page.tsx      # Document sets
│   │   │   │       └── [id]/
│   │   │   │           └── page.tsx  # Set detail
│   │   │   ├── settings/
│   │   │   │   ├── page.tsx          # General settings
│   │   │   │   ├── profile/
│   │   │   │   │   └── page.tsx      # User profile
│   │   │   │   ├── security/
│   │   │   │   │   └── page.tsx      # Security settings
│   │   │   │   ├── tenants/
│   │   │   │   │   └── page.tsx      # Tenant management
│   │   │   │   └── billing/
│   │   │   │       └── page.tsx      # Billing settings
│   │   │   ├── audit/
│   │   │   │   └── page.tsx          # Audit logs
│   │   │   └── kyc/
│   │   │       └── page.tsx          # KYC verification
│   │   │
│   │   ├── api/                      # Next.js API routes (BFF)
│   │   │   └── [...proxy]/
│   │   │       └── route.ts          # Proxy to backend
│   │   │
│   │   ├── layout.tsx                # Root layout
│   │   ├── page.tsx                  # Landing / redirect
│   │   ├── loading.tsx               # Global loading UI
│   │   ├── error.tsx                 # Global error boundary
│   │   ├── not-found.tsx             # 404 page
│   │   └── globals.css               # Tailwind + CSS variables
│   │
│   ├── components/                   # Shared components
│   │   ├── ui/                       # shadcn/ui primitives
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── dropdown-menu.tsx
│   │   │   ├── form.tsx
│   │   │   ├── input.tsx
│   │   │   ├── label.tsx
│   │   │   ├── select.tsx
│   │   │   ├── table.tsx
│   │   │   ├── tabs.tsx
│   │   │   ├── toast.tsx
│   │   │   ├── toaster.tsx
│   │   │   ├── command.tsx
│   │   │   ├── sheet.tsx
│   │   │   ├── skeleton.tsx
│   │   │   ├── badge.tsx
│   │   │   ├── separator.tsx
│   │   │   ├── scroll-area.tsx
│   │   │   ├── tooltip.tsx
│   │   │   ├── popover.tsx
│   │   │   ├── avatar.tsx
│   │   │   ├── switch.tsx
│   │   │   ├── checkbox.tsx
│   │   │   ├── radio-group.tsx
│   │   │   ├── textarea.tsx
│   │   │   ├── progress.tsx
│   │   │   ├── alert.tsx
│   │   │   ├── alert-dialog.tsx
│   │   │   ├── accordion.tsx
│   │   │   └── chart.tsx
│   │   │
│   │   ├── chat/                     # Chat-specific components
│   │   │   ├── ChatWindow.tsx
│   │   │   ├── MessageList.tsx
│   │   │   ├── MessageBubble.tsx
│   │   │   ├── PromptBox.tsx
│   │   │   ├── FileUploader.tsx
│   │   │   ├── VoiceInput.tsx
│   │   │   ├── AgentSelector.tsx
│   │   │   ├── ModelIndicator.tsx
│   │   │   ├── TokenStreamViewer.tsx
│   │   │   ├── ToolExecutionDisplay.tsx
│   │   │   ├── ThinkingIndicator.tsx
│   │   │   ├── ConversationSidebar.tsx
│   │   │   └── ChatSettings.tsx
│   │   │
│   │   ├── agent/                    # Agent management components
│   │   │   ├── AgentCard.tsx
│   │   │   ├── AgentList.tsx
│   │   │   ├── AgentForm.tsx
│   │   │   ├── AgentConfig.tsx
│   │   │   ├── AgentExecutionView.tsx
│   │   │   ├── AgentPerformance.tsx
│   │   │   ├── AgentDocSetBinding.tsx
│   │   │   ├── AgentDbBinding.tsx
│   │   │   ├── AgentTestConnection.tsx
│   │   │   └── AgentSchemaDiscovery.tsx
│   │   │
│   │   ├── dashboard/                # Dashboard widgets
│   │   │   ├── KPICard.tsx
│   │   │   ├── AIUsageChart.tsx
│   │   │   ├── TokenChart.tsx
│   │   │   ├── ModelPerformanceChart.tsx
│   │   │   ├── ActiveAgentsWidget.tsx
│   │   │   ├── ActivityFeed.tsx
│   │   │   ├── SystemHealthGauge.tsx
│   │   │   └── QuickActions.tsx
│   │   │
│   │   ├── forms/                    # Form components
│   │   │   ├── LoginForm.tsx
│   │   │   ├── RegisterForm.tsx
│   │   │   ├── OTPForm.tsx
│   │   │   ├── ChangePasswordForm.tsx
│   │   │   ├── AgentCreateForm.tsx
│   │   │   ├── DocumentUploadForm.tsx
│   │   │   ├── DBConnectionForm.tsx
│   │   │   └── KYCUploadForm.tsx
│   │   │
│   │   ├── layout/                   # Layout components
│   │   │   ├── AppShell.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Header.tsx
│   │   │   ├── Breadcrumb.tsx
│   │   │   ├── CommandMenu.tsx
│   │   │   ├── ThemeToggle.tsx
│   │   │   ├── TenantSwitcher.tsx
│   │   │   └── NotificationBell.tsx
│   │   │
│   │   └── shared/                   # Shared utility components
│   │       ├── ConfirmDialog.tsx
│   │       ├── EmptyState.tsx
│   │       ├── ErrorBoundary.tsx
│   │       ├── LoadingSpinner.tsx
│   │       ├── Pagination.tsx
│   │       ├── SearchInput.tsx
│   │       ├── SortableHeader.tsx
│   │       ├── DataTable.tsx
│   │       └── StatusBadge.tsx
│   │
│   ├── stores/                       # Zustand stores
│   │   ├── auth-store.ts
│   │   ├── chat-store.ts
│   │   ├── agent-store.ts
│   │   ├── dashboard-store.ts
│   │   ├── theme-store.ts
│   │   ├── ui-store.ts
│   │   └── tenant-store.ts
│   │
│   ├── hooks/                        # Custom React hooks
│   │   ├── auth/
│   │   │   ├── useAuth.ts
│   │   │   ├── usePermission.ts
│   │   │   └── useRole.ts
│   │   ├── chat/
│   │   │   ├── useChat.ts
│   │   │   ├── useWebSocket.ts
│   │   │   └── useStreaming.ts
│   │   ├── agent/
│   │   │   ├── useAgents.ts
│   │   │   ├── useAgentExecution.ts
│   │   │   └── useAgentConfig.ts
│   │   ├── dashboard/
│   │   │   ├── useMetrics.ts
│   │   │   ├── useHealth.ts
│   │   │   └── useActivity.ts
│   │   └── shared/
│   │       ├── useDebounce.ts
│   │       ├── useLocalStorage.ts
│   │       └── useMediaQuery.ts
│   │
│   ├── api/                          # API client layer
│   │   ├── client.ts                 # Axios instance + interceptors
│   │   ├── auth.api.ts               # Auth endpoints
│   │   ├── chat.api.ts               # Chat endpoints
│   │   ├── agent.api.ts              # Agent endpoints
│   │   ├── dashboard.api.ts          # Dashboard endpoints
│   │   ├── document.api.ts           # Document endpoints
│   │   └── types.ts                  # API response types
│   │
│   ├── lib/                          # Utility libraries
│   │   ├── utils.ts                  # cn(), formatDate(), etc.
│   │   ├── constants.ts              # API URLs, routes, config
│   │   ├── validations.ts            # Zod schemas
│   │   ├── websocket.ts              # WebSocket manager
│   │   └── token.ts                  # JWT helpers
│   │
│   ├── types/                        # Shared TypeScript types
│   │   ├── auth.types.ts
│   │   ├── chat.types.ts
│   │   ├── agent.types.ts
│   │   ├── dashboard.types.ts
│   │   ├── document.types.ts
│   │   └── api.types.ts
│   │
│   └── providers/                    # Context providers
│       ├── ThemeProvider.tsx
│       ├── AuthProvider.tsx
│       ├── QueryProvider.tsx
│       ├── ToastProvider.tsx
│       └── TenantProvider.tsx
│
├── tests/                            # Test files
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── .env.local                        # Environment variables
├── .env.example                      # Environment template
├── next.config.ts                    # Next.js configuration
├── tailwind.config.ts                # Tailwind configuration
├── components.json                   # shadcn/ui configuration
├── tsconfig.json                     # TypeScript configuration
├── postcss.config.mjs                # PostCSS configuration
├── pnpm-lock.yaml                    # Lock file
├── package.json                      # Dependencies
├── .eslintrc.json                    # ESLint configuration
├── .prettierrc                       # Prettier configuration
├── Dockerfile                        # Docker build
├── docker-compose.yml                # Docker compose
└── README.md                         # Project documentation
```

---

## 5. App Router Structure

### 5.1 Route Groups

| Route Group | Layout | Purpose |
|---|---|---|
| `(auth)` | Minimal (centered card) | Login, register, password reset |
| `(dashboard)` | App shell (sidebar + header) | Authenticated application pages |

### 5.2 Route Hierarchy

```
/                           → Redirect to /dashboard or /login
/login                      → Login page (public)
/register                   → Register page (public)
/forgot-password            → Password reset request (public)
/reset-password/:token      → Password reset form (public)
/otp-verify                 → OTP verification (public)

/dashboard                  → Main dashboard (protected)
/dashboard/agents           → Agent list page
/dashboard/agents/new       → Agent creation page
/dashboard/agents/:id       → Agent detail page
/dashboard/agents/:id/configure → Agent configuration
/dashboard/agents/:id/execute   → Agent execution
/dashboard/agents/:id/history   → Execution history

/dashboard/chat             → New conversation
/dashboard/chat/:id         → Resume conversation

/dashboard/documents        → Document management
/dashboard/knowledge        → Knowledge base
/dashboard/knowledge/sets   → Document sets
/dashboard/knowledge/sets/:id → Set detail

/dashboard/settings         → General settings
/dashboard/settings/profile → User profile
/dashboard/settings/security → Security settings
/dashboard/settings/tenants → Tenant management

/dashboard/audit            → Audit logs
/dashboard/kyc              → KYC verification
```

### 5.3 Server vs Client Components

```
┌──────────────────────────────────────────────────────┐
│                   Server Components                   │
│  (Default in Next.js App Router)                     │
│                                                      │
│  - Page shells and layouts                           │
│  - Data fetching (server-side)                       │
│  - SEO metadata generation                           │
│  - Static content rendering                          │
│  - Redirect logic (auth checks)                      │
├──────────────────────────────────────────────────────┤
│                    Client Components                  │
│  ("use client" directive)                            │
│                                                      │
│  - Interactive forms                                 │
│  - Real-time updates (WebSocket)                     │
│  - State management (Zustand)                        │
│  - Animations (Framer Motion)                        │
│  - Browser APIs (localStorage, notifications)        │
│  - Streaming token display                           │
└──────────────────────────────────────────────────────┘
```

---

## 6. Component Architecture

### 6.1 Component Layers

```
┌─────────────────────────────────────────┐
│           Page Components               │  Route-level (Server or Client)
├─────────────────────────────────────────┤
│         Feature Components              │  Domain-specific (Chat, Agent, Dashboard)
├─────────────────────────────────────────┤
│         Shared Components               │  Cross-cutting (DataTable, EmptyState)
├─────────────────────────────────────────┤
│           UI Primitives                 │  shadcn/ui (Button, Card, Dialog)
└─────────────────────────────────────────┘
```

### 6.2 Component Categories

| Category | Directory | Components |
|---|---|---|
| UI Primitives | `components/ui/` | Button, Card, Dialog, Form, Table, Toast, etc. |
| Chat | `components/chat/` | ChatWindow, MessageList, PromptBox, etc. |
| Agent | `components/agent/` | AgentCard, AgentForm, AgentConfig, etc. |
| Dashboard | `components/dashboard/` | KPICard, AIUsageChart, ActivityFeed, etc. |
| Forms | `components/forms/` | LoginForm, RegisterForm, AgentCreateForm, etc. |
| Layout | `components/layout/` | Sidebar, Header, CommandMenu, etc. |
| Shared | `components/shared/` | DataTable, EmptyState, ErrorBoundary, etc. |

### 6.3 Component Composition Pattern

```tsx
// Page = Layout + Feature Components + Shared Components
export default function AgentListPage() {
  return (
    <PageHeader title="Agents" description="Manage your AI agents" />
    <AgentFilters />
    <AgentList>
      <AgentCard />
      <AgentCard />
    </AgentList>
    <Pagination />
  );
}
```

---

## 7. State Management Architecture

### 7.1 Zustand Store Map

```
┌─────────────────────────────────────────────────┐
│                  Zustand Stores                   │
├─────────────────────────────────────────────────┤
│                                                   │
│  auth-store.ts          → User, tokens, session  │
│  chat-store.ts          → Messages, streaming     │
│  agent-store.ts         → Agents, executions      │
│  dashboard-store.ts     → Metrics, filters        │
│  theme-store.ts         → Light/dark mode         │
│  ui-store.ts            → Sidebar, modals, toast  │
│  tenant-store.ts        → Current tenant, list    │
│                                                   │
└─────────────────────────────────────────────────┘
```

### 7.2 Store Design Pattern

```typescript
interface ChatStore {
  // State
  messages: Message[];
  isStreaming: boolean;
  activeConversationId: string | null;

  // Actions
  addMessage: (message: Message) => void;
  appendToLastMessage: (token: string) => void;
  setStreaming: (value: boolean) => void;
  clearConversation: () => void;
}
```

### 7.3 State Separation

| State Type | Solution | Examples |
|---|---|---|
| UI State | Zustand | Sidebar open, modals, theme |
| Auth State | Zustand + cookies | User, tokens, permissions |
| Server State | TanStack Query | API data, cached responses |
| Form State | React Hook Form | Form values, validation |
| URL State | Next.js searchParams | Filters, pagination, sort |

---

## 8. Data Fetching Architecture

### 8.1 TanStack Query Integration

```
┌─────────────────────────────────────────────────────┐
│                TanStack Query Flow                    │
│                                                       │
│  Component                                           │
│     │                                                │
│     ▼                                                │
│  useQuery() / useMutation()                          │
│     │                                                │
│     ▼                                                │
│  Query Cache (stale-while-revalidate)                │
│     │                                                │
│     ▼                                                │
│  API Client (Axios + JWT interceptors)               │
│     │                                                │
│     ▼                                                │
│  Backend API (REST + WebSocket)                      │
│                                                       │
└─────────────────────────────────────────────────────┘
```

### 8.2 Query Key Strategy

```typescript
// Convention: [domain, entity, ...identifiers]
queryKeys.agents.all        → ['agents']
queryKeys.agents.detail(id) → ['agents', id]
queryKeys.agents.history(id) → ['agents', id, 'history']
queryKeys.chat.messages(id) → ['chat', 'messages', id]
queryKeys.dashboard.metrics → ['dashboard', 'metrics']
```

### 8.3 Caching Strategy

| Data Type | Stale Time | Cache Time | Refetch |
|---|---|---|---|
| User profile | 5 min | 30 min | On mount |
| Agent list | 2 min | 10 min | On mount + refetchInterval |
| Chat messages | 0 (never) | 0 (never) | Manual only |
| Dashboard metrics | 30 sec | 5 min | refetchInterval: 30s |
| Document list | 1 min | 10 min | On mount |

---

## 9. Real-Time Architecture

### 9.1 WebSocket Client

```
┌─────────────────────────────────────────────────────┐
│              WebSocket Connection Flow                │
│                                                       │
│  1. Connect: wss://api.aeroxenexus.com/ws/chat       │
│  2. Authenticate: Send JWT token                      │
│  3. Subscribe: Join conversation room                 │
│  4. Send: User message                               │
│  5. Receive: Token stream, tool calls, completion    │
│  6. Disconnect: On conversation end or timeout       │
│                                                       │
└─────────────────────────────────────────────────────┘
```

### 9.2 WebSocket Message Protocol

```typescript
// Outbound (Client → Server)
type WSOutbound =
  | { type: 'authenticate'; token: string }
  | { type: 'message'; content: string; conversation_id: string }
  | { type: 'ping' };

// Inbound (Server → Client)
type WSInbound =
  | { type: 'token'; content: string }
  | { type: 'tool_call'; name: string; params: Record<string, unknown> }
  | { type: 'tool_result'; name: string; result: unknown }
  | { type: 'thinking'; content: string }
  | { type: 'completed'; conversation_id: string }
  | { type: 'error'; code: string; message: string }
  | { type: 'pong' };
```

### 9.3 Reconnection Strategy

| Attempt | Delay | Max Attempts |
|---|---|---|
| 1st | 1s | |
| 2nd | 2s | |
| 3rd | 4s | |
| 4th+ | 8s (max) | 10 |
| After 10 fails | Show offline UI | |

---

## 10. API Client Architecture

### 10.1 Axios Configuration

```typescript
// src/api/client.ts
const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
});

// Request interceptor: Attach JWT
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken;
  if (token) config.headers.Authorization = `Bearer ${token}`;
  config.headers['X-Tenant-ID'] = useTenantStore.getState().currentTenantId;
  return config;
});

// Response interceptor: Handle 401, refresh token
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      const refreshed = await refreshAccessToken();
      if (!refreshed) return redirectToLogin();
    }
    return Promise.reject(error);
  }
);
```

### 10.2 API Error Standard

```typescript
interface APIError {
  error: {
    code: string;         // "AI_MODEL_TIMEOUT"
    message: string;      // Human-readable message
    request_id: string;   // UUID for tracing
    details?: unknown;    // Additional context
  };
}
```

---

## 11. Error Handling Architecture

### 11.1 Error Boundary Hierarchy

```
┌──────────────────────────────────────┐
│         Global Error Boundary         │  Catches: catastrophic errors
├──────────────────────────────────────┤
│         Route Error Boundary          │  Catches: page-level errors
├──────────────────────────────────────┤
│       Feature Error Boundary          │  Catches: component tree errors
├──────────────────────────────────────┤
│         Toast Notifications           │  Catches: API errors, validation
└──────────────────────────────────────┘
```

### 11.2 Error Handling Strategy

| Error Type | Handler | User Experience |
|---|---|---|
| Network error | Toast + retry button | "Connection lost. Retrying..." |
| 401 Unauthorized | Redirect to login | Session expired message |
| 403 Forbidden | Toast notification | "You don't have permission" |
| 404 Not Found | not-found.tsx page | Custom 404 page |
| 429 Rate Limited | Toast + backoff | "Too many requests. Wait..." |
| 500 Server Error | Error boundary | Fallback UI with retry |
| Validation error | Form field errors | Inline validation messages |
| WebSocket error | Reconnect + toast | "Reconnecting..." |

---

## 12. Performance Architecture

### 12.1 Code Splitting Strategy

```
┌──────────────────────────────────────────────────┐
│              Loading Hierarchy                     │
│                                                    │
│  Root Layout        → Always loaded (shared nav)   │
│  Dashboard Layout   → Dashboard shell (sidebar)    │
│  Page Components    → Dynamic imports (lazy)       │
│  Chart Libraries    → Dynamic imports (lazy)       │
│  Heavy Components   → Dynamic imports (lazy)       │
│                                                    │
└──────────────────────────────────────────────────┘
```

### 12.2 Rendering Strategy by Page

| Page | Strategy | Rationale |
|---|---|---|
| Landing / Marketing | SSG (Static) | SEO, performance |
| Login / Register | SSR | Dynamic, minimal data |
| Dashboard | ISR (60s) | Semi-dynamic metrics |
| Chat | CSR | Real-time, interactive |
| Agent List | SSR + Streaming | Initial data fast |
| Agent Detail | SSR + Hydration | SEO + interactivity |
| Settings | CSR | User-specific, private |

### 12.3 Performance Targets

| Metric | Target |
|---|---|
| First Contentful Paint | < 1.5s |
| Largest Contentful Paint | < 2.5s |
| Cumulative Layout Shift | < 0.1 |
| First Input Delay | < 100ms |
| Time to Interactive | < 3s |
| Bundle Size (initial) | < 200KB gzipped |

### 12.4 Optimization Techniques

| Technique | Implementation |
|---|---|
| Image optimization | next/image with WebP, lazy loading |
| Font optimization | next/font with self-hosted fonts |
| Script optimization | next/script with lazyOnload |
| Component lazy loading | React.lazy + Suspense |
| Virtual scrolling | TanStack Virtual for long lists |
| Debounced search | useDebounce hook (300ms) |
| Memoization | React.memo for expensive renders |
| Prefetching | next/link + route prefetch |

---

## 13. Security Architecture

### 13.1 Security Layers

```
┌──────────────────────────────────────────────────┐
│              Security Architecture                 │
│                                                    │
│  Layer 1: Transport Security                       │
│  ├── HTTPS everywhere (TLS 1.3)                   │
│  ├── HSTS headers                                  │
│  └── Certificate pinning (mobile)                  │
│                                                    │
│  Layer 2: Authentication                           │
│  ├── JWT access tokens (1hr)                       │
│  ├── Refresh token rotation (7 days)               │
│  ├── HTTP-only secure cookies                      │
│  └── CSRF protection (SameSite cookies)            │
│                                                    │
│  Layer 3: Authorization                            │
│  ├── Role-based UI rendering                       │
│  ├── Permission-gated components                   │
│  └── Route guards (middleware)                      │
│                                                    │
│  Layer 4: Input Validation                         │
│  ├── Zod schemas on all forms                      │
│  ├── Sanitized output (XSS prevention)             │
│  └── Content Security Policy headers               │
│                                                    │
│  Layer 5: Data Protection                          │
│  ├── No secrets in client code                     │
│  ├── Environment variables (.env.local only)       │
│  └── Token rotation on compromise                  │
│                                                    │
└──────────────────────────────────────────────────┘
```

### 13.2 Token Management

| Token | Storage | Lifetime | Security |
|---|---|---|---|
| Access Token | Memory (Zustand) | 1 hour | Not persisted to disk |
| Refresh Token | HTTP-only cookie | 7 days | Secure, SameSite=Strict |
| Session Data | Zustand + sessionStorage | Session | Cleared on logout |

### 13.3 XSS Prevention

| Measure | Implementation |
|---|---|
| React auto-escaping | JSX default behavior |
| dangerouslySetInnerHTML | Never used |
| CSP headers | Strict Content-Security-Policy |
| Input sanitization | DOMPurify for rich text |
| Output encoding | React rendering engine |

---

## 14. Deployment Architecture

### 14.1 Deployment Options

```
┌─────────────────────────────────────────────┐
│          Deployment Architecture              │
│                                               │
│  Option A: Vercel (Primary)                  │
│  ├── Edge Runtime (middleware)               │
│  ├── Serverless Functions (API routes)       │
│  ├── ISR (static pages)                      │
│  ├── Edge CDN (global distribution)          │
│  └── Preview deployments (PRs)               │
│                                               │
│  Option B: Docker (Self-Hosted)              │
│  ├── Node.js server (SSR + API routes)       │
│  ├── Nginx reverse proxy                     │
│  ├── Redis (server cache)                    │
│  └── Kubernetes orchestration                │
│                                               │
└─────────────────────────────────────────────┘
```

### 14.2 Environment Configuration

| Variable | Purpose | Required |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | Backend API base URL | Yes |
| `NEXT_PUBLIC_WS_URL` | WebSocket endpoint | Yes |
| `NEXT_PUBLIC_APP_NAME` | Application name | No |
| `NEXT_PUBLIC_TENANT_ID` | Default tenant | No |
| `NODE_ENV` | Environment mode | Yes |

### 14.3 Docker Configuration

```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM node:22-alpine AS runner
WORKDIR /app
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]
```

---

## 15. Mobile Architecture

### 15.1 Responsive Web Strategy

| Breakpoint | Width | Layout |
|---|---|---|
| Mobile | < 768px | Single column, bottom nav, collapsed sidebar |
| Tablet | 768px - 1024px | Two columns, collapsible sidebar |
| Desktop | 1024px - 1440px | Full layout, sidebar always visible |
| Wide | > 1440px | Expanded layout, extra panels |

### 15.2 Mobile-Specific Considerations

| Feature | Mobile Adaptation |
|---|---|
| Sidebar | Bottom tab navigation |
| Chat input | Full-width, auto-expanding |
| File upload | Camera / gallery picker |
| Voice input | Native microphone API |
| Touch targets | Minimum 44x44px |
| Swipe gestures | Conversation list actions |

---

## 16. Offline Architecture

### 16.1 Service Worker Strategy

| Page | Offline Support | Cache Strategy |
|---|---|---|
| Login | Partial (cached shell) | Cache-first |
| Dashboard | Read-only (last data) | Stale-while-revalidate |
| Chat | Message queue (send on reconnect) | Network-first |
| Documents | None (requires server) | Network-only |

### 16.2 Offline Behavior

```
┌──────────────────────────────────────────────────┐
│              Offline Strategy                      │
│                                                    │
│  1. Detect: navigator.onLine + heartbeat failure   │
│  2. Notify: Show "You're offline" banner           │
│  3. Cache: Serve cached data where possible        │
│  4. Queue: Store outbound messages in IndexedDB    │
│  5. Reconnect: Auto-retry WebSocket + flush queue  │
│  6. Sync: Reconcile offline changes with server    │
│                                                    │
└──────────────────────────────────────────────────┘
```

---

## 17. Development Workflow

### 17.1 Scripts

| Command | Purpose |
|---|---|
| `pnpm dev` | Start development server (Turbopack) |
| `pnpm build` | Production build |
| `pnpm start` | Start production server |
| `pnpm lint` | ESLint check |
| `pnpm typecheck` | TypeScript type check |
| `pnpm test` | Run unit tests (Vitest) |
| `pnpm test:e2e` | Run E2E tests (Playwright) |
| `pnpm format` | Prettier format |

### 17.2 Quality Gates

```
Pre-commit:  lint-staged (eslint + prettier)
CI Pipeline: typecheck → lint → test → build → deploy
PR Review:   Visual regression + Lighthouse score
```

---

## 18. Backend Service Mapping

| Frontend Feature | Backend Service | Protocol |
|---|---|---|
| Login / Register | identity-service | REST |
| Chat Streaming | ai-gateway-service | WebSocket |
| Agent Management | agent-orchestrator-service | REST + gRPC |
| Document Upload | rag-service | REST (multipart) |
| Knowledge Search | rag-service | REST |
| Vision Analysis | vision-service | REST (multipart) |
| SQL Queries | sql-agent-service | REST |
| Dashboard Metrics | ai-gateway-service | REST |
| System Health | All services | REST (aggregated) |
| Audit Logs | audit-service | REST |
| KYC Management | identity-service | REST |

---

## 19. Naming Conventions

| Item | Convention | Example |
|---|---|---|
| Files (components) | PascalCase | `AgentCard.tsx` |
| Files (hooks) | camelCase with `use` | `useAgents.ts` |
| Files (stores) | camelCase with `-store` | `auth-store.ts` |
| Files (API) | camelCase with `.api` | `agent.api.ts` |
| Files (types) | camelCase with `.types` | `agent.types.ts` |
| Components | PascalCase | `AgentCard` |
| Hooks | camelCase with `use` | `useAgents` |
| Stores | camelCase | `useAuthStore` |
| API functions | camelCase | `fetchAgents()` |
| Types/Interfaces | PascalCase | `Agent`, `AgentConfig` |
| Constants | UPPER_SNAKE_CASE | `API_BASE_URL` |
| CSS classes | Tailwind utilities | `flex items-center gap-2` |

---

## 20. Documentation Index

| Document | File | Coverage |
|---|---|---|
| Architecture Overview | `00-architecture-overview.md` | This document |
| Design System | `01-design-system.md` | Tokens, components, themes |
| Authentication | `02-authentication.md` | Login, JWT, RBAC, KYC |
| AI Chat | `03-ai-chat.md` | Chat UI, streaming, WebSocket |
| Dashboard | `04-dashboard.md` | Metrics, charts, widgets |
| Agent Management | `05-agent-management.md` | CRUD, config, execution |

---

*Document Version: 1.0 — Last Updated: July 2026*
