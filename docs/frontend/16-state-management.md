# State Management

> Zustand + TanStack Query architecture for Nexus AI — store design, query patterns, persistence, and debugging.

## 1. Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                     STATE ARCHITECTURE                           │
│                                                                  │
│  ┌────────────────────────┐  ┌───────────────────────────────┐   │
│  │      Zustand Stores    │  │     TanStack Query            │   │
│  │      (Client State)    │  │     (Server State)            │   │
│  │                        │  │                               │   │
│  │  • Auth Store          │  │  • Query Cache               │   │
│  │  • UI Store            │  │  • Mutation Cache            │   │
│  │  • Chat Store          │  │  • Infinite Queries          │   │
│  │  • Notification Store  │  │  • Optimistic Updates        │   │
│  │  • Settings Store      │  │  • Background Refetch        │   │
│  │  • Model Store         │  │                               │   │
│  └──────────┬─────────────┘  └───────────────┬───────────────┘   │
│             │                                │                   │
│             └──────────┐  ┌──────────────────┘                   │
│                        ▼  ▼                                      │
│              ┌─────────────────┐                                 │
│              │   Components    │                                 │
│              │   (React)       │                                 │
│              └─────────────────┘                                 │
└──────────────────────────────────────────────────────────────────┘
```

### State Classification

| Category | Technology | Examples |
|---|---|---|
| **Client State** | Zustand | Theme, sidebar, user, modals, toasts |
| **Server State** | TanStack Query | Agents, documents, models, dashboard |
| **URL State** | Next.js searchParams | Filters, pagination, active tab |
| **Form State** | React Hook Form + Zod | Form values, validation errors |
| **Derived State** | Selectors / useMemo | Filtered lists, computed metrics |

## 2. Zustand Architecture

### Store Design Principles

| Principle | Description |
|---|---|
| Single Responsibility | Each store manages one domain |
| Normalized State | IDs as keys, entities in flat maps |
| Minimal Subscriptions | Selectors prevent unnecessary re-renders |
| Action Colocation | Actions live in the same store as state |
| Middleware Layer | persist, devtools, immer, logger |

### Store Composition

```
stores/
├── authStore.ts          (user, token, tenant, permissions)
├── uiStore.ts            (sidebar, modals, toasts, theme)
├── chatStore.ts          (conversations, messages, streaming)
├── notificationStore.ts  (notifications, preferences)
├── settingsStore.ts      (user settings, app preferences)
├── documentStore.ts      (documents, sets, processing state)
├── modelStore.ts         (models, GPU metrics, performance)
├── agentStore.ts         (agents, executions, configs)
├── dashboardStore.ts     (metrics, health, activity)
├── securityStore.ts      (alerts, audit logs, access)
├── adminStore.ts         (users, tenants, KYC)
├── workflowStore.ts      (workflows, executions, builder)
└── performanceStore.ts   (metrics, budgets, violations)
```

## 3. Auth Store

```ts
// stores/authStore.ts
import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";

interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  role: "owner" | "admin" | "member" | "viewer";
  tenantId: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  permissions: string[];
  tenantId: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  lastActivity: number;
}

interface AuthActions {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  setToken: (token: string, refreshToken: string) => void;
  updateUser: (updates: Partial<User>) => void;
  checkPermission: (permission: string) => boolean;
  refreshSession: () => Promise<void>;
  updateActivity: () => void;
}

export const useAuthStore = create<AuthState & AuthActions>()(
  devtools(
    persist(
      immer((set, get) => ({
        user: null,
        token: null,
        refreshToken: null,
        permissions: [],
        tenantId: null,
        isAuthenticated: false,
        isLoading: true,
        lastActivity: Date.now(),

        login: async (credentials) => {
          set({ isLoading: true });
          try {
            const response = await api.post("/api/auth/login", credentials);
            const { user, token, refreshToken, permissions } = response.data;

            set((state) => {
              state.user = user;
              state.token = token;
              state.refreshToken = refreshToken;
              state.permissions = permissions;
              state.tenantId = user.tenantId;
              state.isAuthenticated = true;
              state.isLoading = false;
            });

            api.defaults.headers.common.Authorization = `Bearer ${token}`;
          } catch (error) {
            set({ isLoading: false });
            throw error;
          }
        },

        logout: () => {
          set((state) => {
            state.user = null;
            state.token = null;
            state.refreshToken = null;
            state.permissions = [];
            state.tenantId = null;
            state.isAuthenticated = false;
          });
          delete api.defaults.headers.common.Authorization;
        },

        setToken: (token, refreshToken) =>
          set((state) => {
            state.token = token;
            state.refreshToken = refreshToken;
          }),

        updateUser: (updates) =>
          set((state) => {
            if (state.user) Object.assign(state.user, updates);
          }),

        checkPermission: (permission) => {
          const { permissions, user } = get();
          if (user?.role === "owner") return true;
          return permissions.includes(permission);
        },

        refreshSession: async () => {
          const { refreshToken } = get();
          if (!refreshToken) return get().logout();

          try {
            const response = await api.post("/api/auth/refresh", { refreshToken });
            get().setToken(response.data.token, response.data.refreshToken);
          } catch {
            get().logout();
          }
        },

        updateActivity: () => set({ lastActivity: Date.now() }),
      })),
      {
        name: "nexus-auth",
        partialize: (state) => ({
          token: state.token,
          refreshToken: state.refreshToken,
          user: state.user,
          permissions: state.permissions,
          tenantId: state.tenantId,
        }),
      }
    ),
    { name: "AuthStore" }
  )
);
```

## 4. Chat Store

```ts
// stores/chatStore.ts
interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: string;
  model?: string;
  tokenCount?: number;
  isStreaming?: boolean;
  attachments?: Attachment[];
}

interface Conversation {
  id: string;
  title: string;
  messages: Message[];
  model: string;
  createdAt: string;
  updatedAt: string;
  tokenCount: number;
  isPinned: boolean;
}

interface ChatState {
  conversations: Conversation[];
  activeConversationId: string | null;
  isStreaming: boolean;
  streamingMessageId: string | null;
  selectedModel: string;
  temperature: number;
  maxTokens: number;
}

interface ChatActions {
  createConversation: () => string;
  setActiveConversation: (id: string) => void;
  deleteConversation: (id: string) => void;
  addMessage: (conversationId: string, message: Message) => void;
  updateStreamingMessage: (conversationId: string, messageId: string, content: string) => void;
  setStreaming: (isStreaming: boolean, messageId?: string) => void;
  setSelectedModel: (model: string) => void;
  setTemperature: (temp: number) => void;
  setMaxTokens: (max: number) => void;
  pinConversation: (id: string) => void;
}

export const useChatStore = create<ChatState & ChatActions>()(
  devtools(
    persist(
      immer((set, get) => ({
        conversations: [],
        activeConversationId: null,
        isStreaming: false,
        streamingMessageId: null,
        selectedModel: "gpt-4o",
        temperature: 0.7,
        maxTokens: 4096,

        createConversation: () => {
          const id = crypto.randomUUID();
          const conversation: Conversation = {
            id,
            title: "New Conversation",
            messages: [],
            model: get().selectedModel,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            tokenCount: 0,
            isPinned: false,
          };
          set((state) => {
            state.conversations.unshift(conversation);
            state.activeConversationId = id;
          });
          return id;
        },

        addMessage: (conversationId, message) =>
          set((state) => {
            const conv = state.conversations.find((c) => c.id === conversationId);
            if (conv) {
              conv.messages.push(message);
              conv.tokenCount += message.tokenCount ?? 0;
              conv.updatedAt = new Date().toISOString();
              if (conv.messages.length === 1 && message.role === "user") {
                conv.title = message.content.slice(0, 50) + (message.content.length > 50 ? "..." : "");
              }
            }
          }),

        updateStreamingMessage: (conversationId, messageId, content) =>
          set((state) => {
            const conv = state.conversations.find((c) => c.id === conversationId);
            const msg = conv?.messages.find((m) => m.id === messageId);
            if (msg) msg.content = content;
          }),

        setStreaming: (isStreaming, messageId) =>
          set((state) => {
            state.isStreaming = isStreaming;
            state.streamingMessageId = messageId ?? null;
          }),

        deleteConversation: (id) =>
          set((state) => {
            state.conversations = state.conversations.filter((c) => c.id !== id);
            if (state.activeConversationId === id) {
              state.activeConversationId = state.conversations[0]?.id ?? null;
            }
          }),

        setActiveConversation: (id) => set({ activeConversationId: id }),
        setSelectedModel: (model) => set({ selectedModel: model }),
        setTemperature: (temp) => set({ temperature: temp }),
        setMaxTokens: (max) => set({ maxTokens: max }),

        pinConversation: (id) =>
          set((state) => {
            const conv = state.conversations.find((c) => c.id === id);
            if (conv) conv.isPinned = !conv.isPinned;
          }),
      })),
      { name: "nexus-chat" }
    ),
    { name: "ChatStore" }
  )
);
```

## 5. Agent Store

```ts
// stores/agentStore.ts
interface Agent {
  id: string;
  name: string;
  description: string;
  model: string;
  systemPrompt: string;
  tools: string[];
  status: "active" | "inactive" | "error";
  executionCount: number;
  lastExecutedAt?: string;
  config: AgentConfig;
}

interface AgentExecution {
  id: string;
  agentId: string;
  status: "running" | "completed" | "failed";
  input: string;
  output?: string;
  tokenCount: number;
  duration: number;
  startedAt: string;
  completedAt?: string;
  error?: string;
}

interface AgentState {
  agents: Map<string, Agent>;
  executions: Map<string, AgentExecution[]>;
  activeAgentId: string | null;
  isExecuting: boolean;
}

export const useAgentStore = create<AgentState & AgentActions>()(
  devtools(
    immer((set, get) => ({
      agents: new Map(),
      executions: new Map(),
      activeAgentId: null,
      isExecuting: false,

      setAgents: (agents) =>
        set((state) => {
          state.agents = new Map(agents.map((a) => [a.id, a]));
        }),

      addExecution: (execution) =>
        set((state) => {
          const existing = state.executions.get(execution.agentId) ?? [];
          state.executions.set(execution.agentId, [execution, ...existing].slice(0, 50));
        }),

      updateExecution: (agentId, executionId, updates) =>
        set((state) => {
          const executions = state.executions.get(agentId);
          const execution = executions?.find((e) => e.id === executionId);
          if (execution) Object.assign(execution, updates);
        }),

      setActiveAgent: (id) => set({ activeAgentId: id }),
      setExecuting: (isExecuting) => set({ isExecuting }),
    })),
    { name: "AgentStore" }
  )
);
```

## 6. UI Store

```ts
// stores/uiStore.ts
type ModalId =
  | "createAgent"
  | "editAgent"
  | "deleteConfirm"
  | "settings"
  | "notificationPreferences"
  | "commandPalette"
  | null;

type ToastVariant = "success" | "error" | "warning" | "info";

interface Toast {
  id: string;
  title: string;
  description?: string;
  variant: ToastVariant;
  action?: { label: string; onClick: () => void };
  duration?: number;
}

interface UIState {
  sidebarOpen: boolean;
  sidebarCollapsed: boolean;
  activeModal: ModalId;
  toasts: Toast[];
  theme: "light" | "dark" | "system";
  commandPaletteOpen: boolean;
  activePanel: "chat" | "preview" | "none";
}

interface UIActions {
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  openModal: (modal: ModalId) => void;
  closeModal: () => void;
  addToast: (toast: Omit<Toast, "id">) => void;
  removeToast: (id: string) => void;
  setTheme: (theme: UIState["theme"]) => void;
  toggleCommandPalette: () => void;
  setActivePanel: (panel: UIState["activePanel"]) => void;
}

export const useUIStore = create<UIState & UIActions>()(
  devtools(
    persist(
      immer((set, get) => ({
        sidebarOpen: true,
        sidebarCollapsed: false,
        activeModal: null,
        toasts: [],
        theme: "system",
        commandPaletteOpen: false,
        activePanel: "none",

        toggleSidebar: () => set((state) => { state.sidebarOpen = !state.sidebarOpen; }),
        setSidebarOpen: (open) => set({ sidebarOpen: open }),
        setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

        openModal: (modal) => set({ activeModal: modal }),
        closeModal: () => set({ activeModal: null }),

        addToast: (toast) =>
          set((state) => {
            const id = crypto.randomUUID();
            state.toasts.push({ ...toast, id });
            // Auto-remove after duration
            setTimeout(() => get().removeToast(id), toast.duration ?? 5000);
          }),

        removeToast: (id) =>
          set((state) => {
            state.toasts = state.toasts.filter((t) => t.id !== id);
          }),

        setTheme: (theme) => set({ theme }),

        toggleCommandPalette: () =>
          set((state) => { state.commandPaletteOpen = !state.commandPaletteOpen; }),

        setActivePanel: (panel) => set({ activePanel: panel }),
      })),
      {
        name: "nexus-ui",
        partialize: (state) => ({
          theme: state.theme,
          sidebarCollapsed: state.sidebarCollapsed,
        }),
      }
    ),
    { name: "UIStore" }
  )
);
```

## 7. Settings Store

```ts
// stores/settingsStore.ts
interface SettingsState {
  settings: UserSettings;
  isLoading: boolean;
}

interface UserSettings {
  chat: {
    defaultModel: string;
    temperature: number;
    maxTokens: number;
    systemPrompt: string;
    streamingEnabled: boolean;
  };
  documents: {
    autoProcess: boolean;
    defaultFormat: "pdf" | "markdown" | "json";
    maxFileSize: number;
  };
  notifications: {
    email: boolean;
    push: boolean;
    sound: boolean;
    quietHoursStart: string;
    quietHoursEnd: string;
  };
  appearance: {
    theme: "light" | "dark" | "system";
    fontSize: "small" | "medium" | "large";
    sidebarPosition: "left" | "right";
    compactMode: boolean;
  };
}

export const useSettingsStore = create<SettingsState & SettingsActions>()(
  devtools(
    persist(
      immer((set) => ({
        settings: DEFAULT_SETTINGS,
        isLoading: false,

        updateSettings: (section, updates) =>
          set((state) => {
            Object.assign(state.settings[section], updates);
          }),

        resetSettings: () =>
          set({ settings: DEFAULT_SETTINGS }),

        loadSettings: async () => {
          set({ isLoading: true });
          const response = await api.post("/api/user/settings", {});
          set({ settings: response.data, isLoading: false });
        },

        saveSettings: async () => {
          const { settings } = get();
          await api.patch("/api/user/settings", settings);
        },
      })),
      { name: "nexus-settings" }
    ),
    { name: "SettingsStore" }
  )
);
```

## 8. Zustand Middleware

| Middleware | Purpose | Configuration |
|---|---|---|
| `persist` | localStorage persistence | `name`, `partialize`, `storage` |
| `devtools` | Redux DevTools integration | `name`, `enabled` |
| `immer` | Immutable updates | Direct mutation syntax |
| `logger` | Console logging | `collapsed`, `diff` |

```ts
// Custom logger middleware
export const logger = <T extends object>(
  config: (set: any, get: any, store: any) => T
) => (set: any, get: any, store: any) => {
  const loggedSet: typeof set = (...args) => {
    const prevState = get();
    set(...args);
    const nextState = get();
    console.groupCollapsed(`%c${store.name}`, "color: #888; font-weight: normal");
    console.log("%cprev", "color: #9E9E9E; font-weight: bold", prevState);
    console.log("%cnext", "color: #4CAF50; font-weight: bold", nextState);
    console.groupEnd();
  };
  return config(loggedSet, get, store);
};

// Custom session storage middleware (for non-persisted stores)
export function sessionStorage<T>(name: string) {
  return {
    name,
    storage: {
      getItem: (name: string) => {
        const str = sessionStorage.getItem(name);
        return str ? JSON.parse(str) : null;
      },
      setItem: (name: string, value: unknown) => {
        sessionStorage.setItem(name, JSON.stringify(value));
      },
      removeItem: (name: string) => {
        sessionStorage.removeItem(name);
      },
    },
  };
}
```

## 9. Zustand Selectors

```ts
// Memoized selectors with shallow equality
import { useShallow } from "zustand/react/shallow";

// ❌ Bad: creates new object every render
const { user, token, isAuthenticated } = useAuthStore();

// ✅ Good: shallow comparison prevents unnecessary re-renders
const { user, token, isAuthenticated } = useAuthStore(
  useShallow((state) => ({
    user: state.user,
    token: state.token,
    isAuthenticated: state.isAuthenticated,
  }))
);

// ✅ Good: single value selector (primitive, no shallow needed)
const sidebarOpen = useUIStore((state) => state.sidebarOpen);

// ✅ Good: derived selector with memoization
const activeConversation = useChatStore(
  (state) => state.conversations.find((c) => c.id === state.activeConversationId)
);

// Selector factory for parameterized selectors
const selectAgentExecutions = (agentId: string) => (state: AgentState) =>
  state.executions.get(agentId) ?? [];

function AgentHistory({ agentId }: { agentId: string }) {
  const executions = useAgentStore(selectAgentExecutions(agentId));
  return <ExecutionList executions={executions} />;
}
```

### Selector Performance

| Pattern | Re-render Frequency | Use Case |
|---|---|---|
| `useStore(s => s.primitive)` | On change only | Single primitive values |
| `useStore(useShallow({...}))` | On shallow diff | Multiple values |
| `useStore(s => s.deep.field)` | On field change only | Deep nested access |
| `useStore()` (no selector) | On any store change | ❌ Avoid |

## 10. TanStack Query Architecture

### Query Key Convention

```ts
// lib/queryKeys.ts
export const queryKeys = {
  agents: {
    all: ["agents"] as const,
    lists: () => [...queryKeys.agents.all, "list"] as const,
    list: (filters: AgentFilters) => [...queryKeys.agents.lists(), filters] as const,
    details: () => [...queryKeys.agents.all, "detail"] as const,
    detail: (id: string) => [...queryKeys.agents.details(), id] as const,
    executions: (id: string) => [...queryKeys.agents.detail(id), "executions"] as const,
  },
  documents: {
    all: ["documents"] as const,
    lists: () => [...queryKeys.documents.all, "list"] as const,
    list: (filters: DocFilters) => [...queryKeys.documents.lists(), filters] as const,
    detail: (id: string) => [...queryKeys.documents.all, "detail", id] as const,
    sets: () => [...queryKeys.documents.all, "sets"] as const,
  },
  models: {
    all: ["models"] as const,
    list: () => [...queryKeys.models.all, "list"] as const,
    gpu: () => [...queryKeys.models.all, "gpu"] as const,
    performance: (id: string) => [...queryKeys.models.all, "performance", id] as const,
  },
  dashboard: {
    all: ["dashboard"] as const,
    metrics: () => [...queryKeys.dashboard.all, "metrics"] as const,
    health: () => [...queryKeys.dashboard.all, "health"] as const,
    activity: () => [...queryKeys.dashboard.all, "activity"] as const,
  },
  security: {
    all: ["security"] as const,
    alerts: () => [...queryKeys.security.all, "alerts"] as const,
    auditLogs: (filters: AuditFilters) => [...queryKeys.security.all, "audit", filters] as const,
  },
  admin: {
    all: ["admin"] as const,
    users: () => [...queryKeys.admin.all, "users"] as const,
    tenants: () => [...queryKeys.admin.all, "tenants"] as const,
  },
} as const;
```

### Query Hooks

```ts
// hooks/queries/useAgents.ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/queryKeys";

export function useAgents(filters?: AgentFilters) {
  return useQuery({
    queryKey: queryKeys.agents.list(filters ?? {}),
    queryFn: () => api.post("/api/agents", { filters }).then((r) => r.data),
    staleTime: 30_000,
    placeholderData: keepPreviousData,
  });
}

export function useAgent(id: string) {
  return useQuery({
    queryKey: queryKeys.agents.detail(id),
    queryFn: () => api.post(`/api/agents/${id}`, {}).then((r) => r.data),
    enabled: !!id,
  });
}

export function useAgentExecutions(agentId: string) {
  return useQuery({
    queryKey: queryKeys.agents.executions(agentId),
    queryFn: () =>
      api.post(`/api/agents/${agentId}/executions`, {}).then((r) => r.data),
    enabled: !!agentId,
  });
}

export function useCreateAgent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAgentInput) =>
      api.post("/api/agents", data).then((r) => r.data),

    onSuccess: (newAgent) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.agents.all });
      toast.success(`Agent "${newAgent.name}" created`);
    },

    onError: (error: ApiError) => {
      toast.error("Failed to create agent", { description: error.message });
    },
  });
}

export function useUpdateAgent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateAgentInput }) =>
      api.patch(`/api/agents/${id}`, data).then((r) => r.data),

    // Optimistic update
    onMutate: async ({ id, data }) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.agents.detail(id) });

      const previous = queryClient.getQueryData(queryKeys.agents.detail(id));

      queryClient.setQueryData(queryKeys.agents.detail(id), (old: Agent) => ({
        ...old,
        ...data,
      }));

      return { previous };
    },

    onError: (_err, { id }, context) => {
      queryClient.setQueryData(queryKeys.agents.detail(id), context?.previous);
      toast.error("Failed to update agent");
    },

    onSettled: (_data, _error, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.agents.all });
    },
  });
}
```

## 11. Query Patterns

### Infinite Scroll

```ts
// hooks/queries/useInfiniteDocuments.ts
export function useInfiniteDocuments(filters: DocFilters) {
  return useInfiniteQuery({
    queryKey: [...queryKeys.documents.all, "infinite", filters],
    queryFn: ({ pageParam = 1 }) =>
      api
        .post("/api/documents", { filters, page: pageParam, limit: 20 })
        .then((r) => r.data),
    getNextPageParam: (lastPage) =>
      lastPage.page < lastPage.totalPages ? lastPage.page + 1 : undefined,
    initialPageParam: 1,
  });
}

// Usage in component
function DocumentList() {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteDocuments({ status: "processed" });

  const allDocuments = data?.pages.flatMap((p) => p.documents) ?? [];

  return (
    <div>
      {allDocuments.map((doc) => (
        <DocumentCard key={doc.id} document={doc} />
      ))}
      {hasNextPage && (
        <button onClick={() => fetchNextPage()} disabled={isFetchingNextPage}>
          {isFetchingNextPage ? "Loading..." : "Load more"}
        </button>
      )}
    </div>
  );
}
```

### Optimistic Updates

```ts
// Optimistic mutation pattern
export function useToggleDocumentFavorite() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (docId: string) =>
      api.patch(`/api/documents/${docId}/favorite`).then((r) => r.data),

    onMutate: async (docId) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.documents.all });

      const previous = queryClient.getQueriesData({ queryKey: queryKeys.documents.all });

      queryClient.setQueriesData(
        { queryKey: queryKeys.documents.all },
        (old: any) => {
          if (!old?.pages) return old;
          return {
            ...old,
            pages: old.pages.map((page: any) => ({
              ...page,
              documents: page.documents.map((doc: any) =>
                doc.id === docId ? { ...doc, isFavorite: !doc.isFavorite } : doc
              ),
            })),
          };
        }
      );

      return { previous };
    },

    onError: (_err, _docId, context) => {
      // Rollback
      context?.previous.forEach(([key, data]) => {
        queryClient.setQueryData(key, data);
      });
      toast.error("Failed to update favorite");
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.documents.all });
    },
  });
}
```

## 12. Query Cache Management

```ts
// Cache invalidation patterns
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,        // 30s before refetch
      gcTime: 5 * 60_000,      // 5min garbage collection
      refetchOnWindowFocus: true,
      refetchOnReconnect: true,
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30_000),
    },
  },
});

// Selective invalidation
await queryClient.invalidateQueries({ queryKey: queryKeys.agents.all });       // All agent queries
await queryClient.invalidateQueries({ queryKey: queryKeys.agents.detail("1") }); // Specific agent
await queryClient.invalidateQueries({
  queryKey: queryKeys.agents.all,
  exact: true, // Only exact match, not child queries
});

// Prefetch on hover
const prefetchAgent = useCallback(
  (id: string) => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.agents.detail(id),
      queryFn: () => api.post(`/api/agents/${id}`, {}).then((r) => r.data),
      staleTime: 60_000,
    });
  },
  [queryClient]
);
```

## 13. Query Error Handling

```tsx
// Error boundary per query
function AgentDetail({ id }: { id: string }) {
  const { data, error, isLoading, isError } = useAgent(id);

  if (isLoading) return <AgentSkeleton />;
  if (isError) return <QueryError error={error} queryKey={queryKeys.agents.detail(id)} />;

  return <AgentCard agent={data} />;
}

// Global query error handler
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      meta: {
        onError: (error: ApiError) => {
          if (error.status === 401) {
            useAuthStore.getState().logout();
            router.push("/login");
          }
          if (error.status === 403) {
            toast.error("You don't have permission for this action");
          }
        },
      },
    },
  },
});
```

## 14. State Synchronization

### Zustand ↔ TanStack Query

```ts
// Sync query results to Zustand store
export function useSyncAgentMetrics() {
  const { data: metrics } = useQuery({
    queryKey: queryKeys.dashboard.metrics(),
    queryFn: fetchDashboardMetrics,
    refetchInterval: 30_000,
  });

  const setMetrics = useDashboardStore((s) => s.setMetrics);

  useEffect(() => {
    if (metrics) setMetrics(metrics);
  }, [metrics, setMetrics]);
}
```

### Cross-Store Communication

```ts
// Auth store listens to UI store events
useUIStore.subscribe(
  (state) => state.theme,
  (theme) => {
    document.documentElement.classList.toggle("dark", theme === "dark");
  }
);

// Global: invalidate all queries on logout
useAuthStore.subscribe(
  (state, prevState) => {
    if (prevState.isAuthenticated && !state.isAuthenticated) {
      queryClient.clear();
    }
  }
);
```

## 15. State Persistence

| Store | Storage | Persistence Scope |
|---|---|---|
| Auth | localStorage | Token, user, permissions |
| Chat | localStorage | Conversations, model config |
| UI | localStorage | Theme, sidebar state |
| Settings | localStorage + API | User preferences |
| Notifications | localStorage | Notifications, preferences |
| Performance | sessionStorage | Metrics (session only) |
| Agents | Server only (Query) | N/A |
| Documents | Server only (Query) | N/A |

```ts
// IndexedDB for large data (chat history)
import { get, set, del } from "idb-keyval";

const dbStorage = {
  getItem: async (name: string) => {
    const value = await get(name);
    return value ? JSON.parse(value) : null;
  },
  setItem: async (name: string, value: unknown) => {
    await set(name, JSON.stringify(value));
  },
  removeItem: async (name: string) => {
    await del(name);
  },
};

// Use for large chat history
persist(
  immer((set) => ({ /* ... */ })),
  { name: "nexus-chat-history", storage: createJSONStorage(() => dbStorage) }
);
```

## 16. State Debugging

```ts
// Zustand DevTools
import { useDevTools } from "zustand-devtools";

// In development
if (process.env.NODE_ENV === "development") {
  // Connect to Redux DevTools
  useAuthStore.subscribe(console.log);
}
```

### DevTools Panel

| Tool | Access | Use Case |
|---|---|---|
| Redux DevTools | Browser extension | State inspection, time travel |
| React DevTools | Browser extension | Component render tracking |
| TanStack Query DevTools | `<ReactQueryDevtools />` | Query cache inspection |
| Network tab | Browser DevTools | API calls, caching verification |
| Performance tab | Browser DevTools | Render performance |

## 17. State Performance

| Strategy | Description | Implementation |
|---|---|---|
| Selective subscriptions | Only subscribe to needed state | `useStore(s => s.field)` |
| Shallow equality | Prevent re-renders for object picks | `useShallow()` |
| Batch updates | Group state changes | Zustand batches automatically |
| Computed values | Derive outside store | `useMemo` in components |
| Query deduplication | Same key = one request | TanStack Query default |

## 18. File Structure

```
src/
├── stores/
│   ├── authStore.ts
│   ├── chatStore.ts
│   ├── agentStore.ts
│   ├── dashboardStore.ts
│   ├── documentStore.ts
│   ├── modelStore.ts
│   ├── notificationStore.ts
│   ├── settingsStore.ts
│   ├── securityStore.ts
│   ├── adminStore.ts
│   ├── workflowStore.ts
│   ├── uiStore.ts
│   └── performanceStore.ts
├── hooks/
│   ├── queries/
│   │   ├── useAgents.ts
│   │   ├── useDocuments.ts
│   │   ├── useModels.ts
│   │   ├── useDashboard.ts
│   │   ├── useSecurity.ts
│   │   └── useAdmin.ts
│   └── useStore.ts
├── lib/
│   └── queryKeys.ts
├── providers/
│   └── QueryProvider.tsx
└── types/
    └── store.ts
```
