# 19 — Integration Guide

## API Client + WebSocket + React Hooks + Forms + Error Handling + Caching + Analytics

---

## 1. Overview

AeroXe Nexus AI frontend integrates with the backend via REST APIs, WebSocket connections, and event-driven systems. This guide covers every integration point.

```
┌─────────────────────────────────────────────────────┐
│                  Frontend App                        │
│                                                     │
│  ┌─────────┐  ┌──────────┐  ┌────────────────────┐ │
│  │ React   │  │ Zustand  │  │ TanStack Query     │ │
│  │ Hooks   │←→│ Stores   │←→│ Cache Layer        │ │
│  └────┬────┘  └────┬─────┘  └────────┬───────────┘ │
│       │            │                  │              │
│  ┌────▼────────────▼──────────────────▼───────────┐ │
│  │              API Client (Axios)                 │ │
│  │         Request/Response Interceptors           │ │
│  └────┬──────────────────────────┬────────────────┘ │
│       │                          │                  │
│  ┌────▼────┐              ┌──────▼──────┐          │
│  │ REST    │              │ WebSocket   │          │
│  │ API     │              │ Client      │          │
│  └────┬────┘              └──────┬──────┘          │
└───────┼──────────────────────────┼──────────────────┘
        │                          │
   ┌────▼────┐              ┌──────▼──────┐
   │ API     │              │ WebSocket   │
   │ Gateway │              │ Server      │
   └─────────┘              └─────────────┘
```

---

## 2. API Client Setup

### Axios Instance

```typescript
// src/lib/api-client.ts
import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { useAuthStore } from '@/stores/auth-store';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    // Add tenant ID header
    const tenantId = useAuthStore.getState().tenantId;
    if (tenantId) {
      config.headers['X-Tenant-ID'] = tenantId;
    }

    // Add request ID for tracing
    config.headers['X-Request-ID'] = crypto.randomUUID();

    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    // Handle 401 Unauthorized
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = useAuthStore.getState().refreshToken;
        if (!refreshToken) {
          throw new Error('No refresh token');
        }

        const { data } = await axios.post(`${API_BASE_URL}/api/v1/auth/refresh`, {
          refresh_token: refreshToken,
        });

        useAuthStore.getState().setToken(data.access_token);
        useAuthStore.getState().setRefreshToken(data.refresh_token);

        originalRequest.headers.Authorization = `Bearer ${data.access_token}`;
        return apiClient(originalRequest);
      } catch (refreshError) {
        useAuthStore.getState().logout();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    // Transform error
    const errorData = (error.response?.data as any)?.error;
    const transformedError = {
      message: errorData?.message || error.message || 'Unknown error',
      status: error.response?.status,
      code: errorData?.code,
      request_id: errorData?.request_id,
    };

    return Promise.reject(transformedError);
  }
);
```

### API Methods

```typescript
// src/lib/api.ts
import { apiClient } from './api-client';
import type {
  LoginRequest,
  LoginResponse,
  Agent,
  Message,
  Conversation,
  DashboardStats,
} from '@/types/api';

export const authApi = {
  login: (data: LoginRequest) =>
    apiClient.post<LoginResponse>('/api/v1/auth/login', data),

  logout: () => apiClient.post('/api/v1/auth/logout'),

  refreshToken: (refreshToken: string) =>
    apiClient.post('/api/v1/auth/refresh', { refresh_token: refreshToken }),

  getProfile: () => apiClient.get('/api/v1/auth/profile'),
};

export const agentApi = {
  list: () => apiClient.get<{ agents: Agent[] }>('/api/v1/agents'),

  get: (id: string) => apiClient.get<Agent>(`/api/v1/agents/${id}`),

  create: (data: Partial<Agent>) =>
    apiClient.post<Agent>('/api/v1/agents', data),

  update: (id: string, data: Partial<Agent>) =>
    apiClient.put<Agent>(`/api/v1/agents/${id}`, data),

  delete: (id: string) => apiClient.delete(`/api/v1/agents/${id}`),
};

export const chatApi = {
  sendMessage: (data: { message: string; agent: string; conversation_id?: string }) =>
    apiClient.post<Message>('/api/v1/ai/chat', data),

  getConversations: () =>
    apiClient.get<{ conversations: Conversation[] }>('/api/v1/ai/conversations'),

  getMessages: (conversationId: string) =>
    apiClient.get<{ messages: Message[] }>(
      `/api/v1/ai/conversations/${conversationId}/messages`
    ),

  deleteConversation: (id: string) =>
    apiClient.delete(`/api/v1/ai/conversations/${id}`),
};

export const dashboardApi = {
  getStats: () => apiClient.get<DashboardStats>('/api/v1/dashboard/stats'),

  getBroadbandMetrics: () =>
    apiClient.get('/api/v1/dashboard/broadband'),

  getERPMetrics: () =>
    apiClient.get('/api/v1/dashboard/erp'),

  getCRMMetrics: () =>
    apiClient.get('/api/v1/dashboard/crm'),
};

export const fileApi = {
  upload: (file: File, onProgress?: (pct: number) => void) => {
    const formData = new FormData();
    formData.append('file', file);

    return apiClient.post('/api/v1/rag/documents', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (e.total && onProgress) {
          onProgress(Math.round((e.loaded / e.total) * 100));
        }
      },
    });
  },

  download: (fileId: string) =>
    apiClient.get(`/api/v1/rag/documents/${fileId}`, { responseType: 'blob' }),
};
```

---

## 3. TypeScript API Types

```typescript
// src/types/api.ts
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

export interface User {
  id: string;
  email: string;
  name: string;
  roles: string[];
  permissions: string[];
  tenant_id: string;
  avatar: string | null;
  created_at: string;
  last_login: string | null;
}

export interface Agent {
  id: string;
  name: string;
  description: string;
  status: 'active' | 'inactive' | 'error';
  system_prompt?: string;
  capabilities: string[];
  knowledge_base_ids: string[];
  tools: string[];
  tenant_id: string;
  created_at: string;
  updated_at: string;
  max_tokens?: number;
  temperature?: number;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  agent_id?: string;
  conversationId?: string;
  tool_calls?: ToolCall[];
  metadata?: Record<string, unknown>;
}

export interface ToolCall {
  id: string;
  name: string;
  arguments: string;
  result?: string;
}

export interface Conversation {
  id: string;
  title: string;
  agent_id: string;
  created_at: string;
  updated_at: string;
  message_count: number;
}

export interface DashboardStats {
  total_conversations: number;
  total_messages: number;
  active_agents: number;
  response_time_avg: number;
  satisfaction_score: number;
}

export interface APIError {
  message: string;
  status: number;
  code?: string;
  details?: Record<string, unknown>;
}
```

---

## 4. WebSocket Client Setup

### Connection Manager

```typescript
// src/lib/websocket-client.ts
import { useAuthStore } from '@/stores/auth-store';
import { useChatStore } from '@/stores/chat-store';

type WebSocketMessage = {
  type: 'token' | 'tool_call' | 'tool_result' | 'completed' | 'error' | 'heartbeat';
  [key: string]: unknown;
};

type MessageHandler = (message: WebSocketMessage) => void;
type StatusHandler = (status: 'connecting' | 'connected' | 'disconnected' | 'error') => void;

class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private statusHandlers: Set<StatusHandler> = new Set();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private isIntentionalClose = false;

  constructor(url: string) {
    this.url = url;
  }

  connect() {
    const token = useAuthStore.getState().token;
    if (!token) return;

    this.isIntentionalClose = false;
    this.notifyStatus('connecting');

    this.ws = new WebSocket(`${this.url}?token=${token}`);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.notifyStatus('connected');
      this.startHeartbeat();
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.dispatch(message);
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };

    this.ws.onclose = (event) => {
      this.stopHeartbeat();
      if (!this.isIntentionalClose) {
        this.notifyStatus('disconnected');
        this.reconnect();
      }
    };

    this.ws.onerror = () => {
      this.notifyStatus('error');
    };
  }

  disconnect() {
    this.isIntentionalClose = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    this.stopHeartbeat();
    this.ws?.close();
    this.ws = null;
  }

  send(data: object) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  on(type: string, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => this.handlers.get(type)?.delete(handler);
  }

  onStatus(handler: StatusHandler) {
    this.statusHandlers.add(handler);
    return () => this.statusHandlers.delete(handler);
  }

  private dispatch(message: WebSocketMessage) {
    const typeHandlers = this.handlers.get(message.type);
    if (typeHandlers) {
      typeHandlers.forEach((handler) => handler(message));
    }
    const allHandlers = this.handlers.get('*');
    if (allHandlers) {
      allHandlers.forEach((handler) => handler(message));
    }
  }

  private notifyStatus(status: string) {
    this.statusHandlers.forEach((handler) => handler(status as any));
  }

  private reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.notifyStatus('error');
      return;
    }

    const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 30000);
    this.reconnectAttempts++;

    this.reconnectTimer = setTimeout(() => {
      this.connect();
    }, delay);
  }

  private startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      this.send({ type: 'heartbeat' });
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }
}

export const wsClient = new WebSocketClient(
  process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8000/ws/chat'
);
```

### WebSocket Message Types

| Type | Direction | Payload | Description |
|------|-----------|---------|-------------|
| `token` | Server→Client | `{ content: string }` | Streaming text token |
| `tool_call` | Server→Client | `{ content: string }` | Agent calling a tool |
| `tool_result` | Server→Client | `{ content: string }` | Tool execution result |
| `thinking` | Server→Client | `{ content: string }` | Agent reasoning |
| `completed` | Server→Client | `{}` | Response complete |
| `error` | Server→Client | `{ message, code }` | Error occurred |
| `ping` | Client→Server | `{}` | Keep-alive ping |
| `pong` | Server→Client | `{}` | Keep-alive pong |
| `message` | Client→Server | `{ type: "message", content: string }` | Send message |

---

## 5. React Hooks for API

### useQuery Patterns

```typescript
// src/hooks/use-agents.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { agentApi } from '@/lib/api';

export function useAgents() {
  return useQuery({
    queryKey: ['agents'],
    queryFn: async () => {
      const { data } = await agentApi.list();
      return data.agents;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 30 * 60 * 1000, // 30 minutes
  });
}

export function useAgent(id: string) {
  return useQuery({
    queryKey: ['agents', id],
    queryFn: async () => {
      const { data } = await agentApi.get(id);
      return data;
    },
    enabled: !!id,
  });
}

export function useCreateAgent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<Agent>) => agentApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] });
    },
  });
}
```

### useMutation Patterns

```typescript
// src/hooks/use-chat.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { chatApi } from '@/lib/api';
import { useChatStore } from '@/stores/chat-store';

export function useSendMessage() {
  const queryClient = useQueryClient();
  const addMessage = useChatStore((s) => s.addMessage);

  return useMutation({
    mutationFn: (data: { content: string; agent_id: string; conversation_id?: string }) =>
      chatApi.sendMessage(data),
    onSuccess: (response) => {
      addMessage(response.data);
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
    },
  });
}
```

---

## 6. React Hooks for WebSocket

```typescript
// src/hooks/use-websocket.ts
import { useEffect, useRef, useCallback } from 'react';
import { wsClient } from '@/lib/websocket-client';
import { useChatStore } from '@/stores/chat-store';

export function useWebSocket() {
  const statusRef = useRef<'connecting' | 'connected' | 'disconnected' | 'error'>('disconnected');
  const appendToken = useChatStore((s) => s.appendStreamingToken);

  useEffect(() => {
    wsClient.connect();

    const unsubStatus = wsClient.onStatus((status) => {
      statusRef.current = status;
    });

    const unsubToken = wsClient.on('token', (msg) => {
      appendToken(msg.token as string);
    });

    const unsubCompleted = wsClient.on('completed', (msg) => {
      useChatStore.getState().finalizeStreamingMessage(msg.message_id as string);
    });

    const unsubError = wsClient.on('error', (msg) => {
      console.error('WebSocket error:', msg);
    });

    return () => {
      unsubStatus();
      unsubToken();
      unsubCompleted();
      unsubError();
      wsClient.disconnect();
    };
  }, []);

  const send = useCallback((data: object) => {
    wsClient.send(data);
  }, []);

  return { send, status: statusRef };
}
```

---

## 7. React Hooks for Auth

```typescript
// src/hooks/use-auth.ts
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth-store';
import { authApi } from '@/lib/api';
import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';

export function useAuth() {
  const { user, token, isAuthenticated, setToken, setUser, setRefreshToken, logout: storeLogout } =
    useAuthStore();
  const navigate = useNavigate();

  const loginMutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: (response) => {
      const { access_token, refresh_token, user: userData } = response.data;
      setToken(access_token);
      setRefreshToken(refresh_token);
      setUser(userData);
      navigate('/dashboard');
      toast.success(`Welcome, ${userData.name}!`);
    },
    onError: (error: any) => {
      toast.error(error.message || 'Login failed');
    },
  });

  const logout = async () => {
    try {
      await authApi.logout();
    } catch {
      // Ignore logout API errors
    } finally {
      storeLogout();
      navigate('/login');
    }
  };

  return {
    user,
    token,
    isAuthenticated,
    login: loginMutation.mutate,
    loginLoading: loginMutation.isPending,
    logout,
  };
}

// Permission hook
export function usePermission(permission: string): boolean {
  const user = useAuthStore((s) => s.user);
  return user?.permissions.includes(permission) ?? false;
}

// Role hook
export function useRole(role: string): boolean {
  const user = useAuthStore((s) => s.user);
  return user?.roles.includes(role) ?? false;
}
```

---

## 8. Form Integration (React Hook Form + Zod)

### Login Form

```typescript
// src/components/login-form.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '@/hooks/use-auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';

const loginFormSchema = z.object({
  email: z.string().email('Please enter a valid email'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
});

type LoginFormValues = z.infer<typeof loginFormSchema>;

export function LoginForm() {
  const { login, loginLoading } = useAuth();

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginFormSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit((values) => login(values))} className="space-y-4">
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input placeholder="admin@aeroxe.com" type="email" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <Input placeholder="Enter password" type="password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit" loading={loginLoading} className="w-full">
          Sign In
        </Button>
      </form>
    </Form>
  );
}
```

### Chat Input Form

```typescript
// src/components/chat-input.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';

const chatFormSchema = z.object({
  message: z.string().min(1, 'Message cannot be empty').max(4000, 'Message too long'),
});

type ChatFormValues = z.infer<typeof chatFormSchema>;

interface ChatInputProps {
  onSend: (message: string) => void;
  isSending?: boolean;
  onFileUpload?: (file: File) => void;
}

export function ChatInput({ onSend, isSending, onFileUpload }: ChatInputProps) {
  const form = useForm<ChatFormValues>({
    resolver: zodResolver(chatFormSchema),
    defaultValues: { message: '' },
  });

  const handleSubmit = (values: ChatFormValues) => {
    onSend(values.message);
    form.reset();
  };

  return (
    <form onSubmit={form.handleSubmit(handleSubmit)} className="flex gap-2">
      <Textarea
        {...form.register('message')}
        placeholder="Type your message..."
        disabled={isSending}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            form.handleSubmit(handleSubmit)();
          }
        }}
      />
      <Button type="submit" disabled={isSending || !form.watch('message')}>
        Send
      </Button>
    </form>
  );
}
```

---

## 9. File Upload Integration

```typescript
// src/hooks/use-file-upload.ts
import { useState, useCallback } from 'react';
import { fileApi } from '@/lib/api';

interface UseFileUploadOptions {
  maxSize?: number; // bytes
  acceptedTypes?: string[];
  onSuccess?: (url: string) => void;
  onError?: (error: string) => void;
}

export function useFileUpload(options: UseFileUploadOptions = {}) {
  const { maxSize = 10 * 1024 * 1024, acceptedTypes = [], onSuccess, onError } = options;
  const [progress, setProgress] = useState(0);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const upload = useCallback(async (file: File) => {
    if (file.size > maxSize) {
      const msg = `File size exceeds ${maxSize / 1024 / 1024}MB limit`;
      setError(msg);
      onError?.(msg);
      return;
    }

    if (acceptedTypes.length > 0 && !acceptedTypes.includes(file.type)) {
      const msg = `File type ${file.type} not accepted`;
      setError(msg);
      onError?.(msg);
      return;
    }

    setIsUploading(true);
    setProgress(0);
    setError(null);

    try {
      const { data } = await fileApi.upload(file, (pct) => setProgress(pct));
      onSuccess?.(data.url);
      return data.url;
    } catch (err: any) {
      const msg = err.message || 'Upload failed';
      setError(msg);
      onError?.(msg);
    } finally {
      setIsUploading(false);
    }
  }, [maxSize, acceptedTypes, onSuccess, onError]);

  return { upload, progress, isUploading, error };
}
```

---

## 10. Error Integration

### Error Boundary

```typescript
// src/components/error-boundary.tsx
import { Component, type ReactNode } from 'react';
import { Button } from '@/components/ui/button';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught:', error, errorInfo);
    // Send to Sentry
    if (window.Sentry) {
      window.Sentry.captureException(error, { extra: errorInfo });
    }
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback || (
          <div className="flex flex-col items-center justify-center p-8">
            <h2 className="text-lg font-semibold">Something went wrong</h2>
            <p className="text-muted-foreground mt-2">
              {this.state.error?.message}
            </p>
            <Button
              className="mt-4"
              onClick={() => this.setState({ hasError: false })}
            >
              Try Again
            </Button>
          </div>
        )
      );
    }

    return this.props.children;
  }
}
```

### Toast Error Hook

```typescript
// src/hooks/use-error-toast.ts
import { toast } from 'sonner';
import { useCallback } from 'react';

export function useErrorToast() {
  const showError = useCallback((error: unknown) => {
    const message =
      error instanceof Error
        ? error.message
        : typeof error === 'string'
        ? error
        : 'An unexpected error occurred';

    toast.error(message);
  }, []);

  return { showError };
}
```

---

## 11. Cache Integration

```typescript
// src/lib/query-client.ts
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,      // 5 minutes
      gcTime: 30 * 60 * 1000,         // 30 minutes
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 1,
    },
  },
});

// Prefetch utilities
export async function prefetchAgents() {
  await queryClient.prefetchQuery({
    queryKey: ['agents'],
    queryFn: () => agentApi.list(),
    staleTime: 5 * 60 * 1000,
  });
}

export async function prefetchConversations() {
  await queryClient.prefetchQuery({
    queryKey: ['conversations'],
    queryFn: () => chatApi.getConversations(),
    staleTime: 2 * 60 * 1000,
  });
}
```

---

## 12. Date Handling (date-fns)

```typescript
// src/lib/date-utils.ts
import { format, formatDistanceToNow, parseISO, isValid, startOfDay, endOfDay } from 'date-fns';

export function formatDate(date: string | Date, formatStr = 'MMM dd, yyyy'): string {
  const parsed = typeof date === 'string' ? parseISO(date) : date;
  return isValid(parsed) ? format(parsed, formatStr) : 'Invalid date';
}

export function formatRelativeTime(date: string | Date): string {
  const parsed = typeof date === 'string' ? parseISO(date) : date;
  return isValid(parsed) ? formatDistanceToNow(parsed, { addSuffix: true }) : 'Unknown';
}

export function getDayRange(date: string | Date): { start: Date; end: Date } {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return { start: startOfDay(d), end: endOfDay(d) };
}
```

---

## 13. Number Formatting

```typescript
// src/lib/number-utils.ts
export function formatCurrency(amount: number, currency = 'INR'): string {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amount);
}

export function formatPercentage(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat('en-IN').format(value);
}

export function formatCompact(value: number): string {
  return new Intl.NumberFormat('en-IN', {
    notation: 'compact',
    compactDisplay: 'short',
  }).format(value);
}
```

---

## 14. Analytics Integration

```typescript
// src/lib/analytics.ts
type AnalyticsEvent = {
  event: string;
  properties?: Record<string, unknown>;
  timestamp: string;
};

class Analytics {
  private queue: AnalyticsEvent[] = [];
  private flushInterval: ReturnType<typeof setInterval> | null = null;

  init() {
    this.flushInterval = setInterval(() => this.flush(), 30000);
  }

  track(event: string, properties?: Record<string, unknown>) {
    this.queue.push({
      event,
      properties,
      timestamp: new Date().toISOString(),
    });

    if (this.queue.length >= 10) {
      this.flush();
    }
  }

  page(pageName: string) {
    this.track('page_view', { page: pageName });
  }

  private async flush() {
    if (this.queue.length === 0) return;

    const events = [...this.queue];
    this.queue = [];

    try {
      await fetch('/api/v1/analytics/events', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ events }),
      });
    } catch {
      this.queue.unshift(...events);
    }
  }

  destroy() {
    if (this.flushInterval) clearInterval(this.flushInterval);
    this.flush();
  }
}

export const analytics = new Analytics();
```

---

## 15. Feature Flags

```typescript
// src/lib/feature-flags.ts
import { useFeatureFlagStore } from '@/stores/feature-flag-store';

export function useFeatureFlag(flag: string): boolean {
  const flags = useFeatureFlagStore((s) => s.flags);
  return flags[flag] ?? false;
}

// Usage
function DashboardPage() {
  const showNewUI = useFeatureFlag('new_dashboard_ui');
  const enableVoiceInput = useFeatureFlag('voice_input');

  return (
    <div>
      {showNewUI ? <NewDashboard /> : <OldDashboard />}
      {enableVoiceInput && <VoiceInputButton />}
    </div>
  );
}
```

---

## 16. Logging Integration

```typescript
// src/lib/logger.ts
type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LogEntry {
  level: LogLevel;
  message: string;
  data?: unknown;
  timestamp: string;
  url?: string;
}

class Logger {
  private buffer: LogEntry[] = [];
  private maxBufferSize = 50;

  private log(level: LogLevel, message: string, data?: unknown) {
    const entry: LogEntry = {
      level,
      message,
      data,
      timestamp: new Date().toISOString(),
      url: typeof window !== 'undefined' ? window.location.href : undefined,
    };

    if (process.env.NODE_ENV === 'development') {
      console[level](message, data);
    }

    this.buffer.push(entry);
    if (this.buffer.length >= this.maxBufferSize) {
      this.flush();
    }
  }

  debug(msg: string, data?: unknown) { this.log('debug', msg, data); }
  info(msg: string, data?: unknown) { this.log('info', msg, data); }
  warn(msg: string, data?: unknown) { this.log('warn', msg, data); }
  error(msg: string, data?: unknown) { this.log('error', msg, data); }

  async flush() {
    if (this.buffer.length === 0) return;
    const entries = [...this.buffer];
    this.buffer = [];
    try {
      await fetch('/api/v1/logs/client', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ logs: entries }),
      });
    } catch {
      this.buffer.unshift(...entries);
    }
  }
}

export const logger = new Logger();
```

---

## 17. SEO Integration

```typescript
// src/lib/seo.ts
interface SEOConfig {
  title: string;
  description: string;
  image?: string;
  url?: string;
  type?: string;
}

export function setSEO(config: SEOConfig) {
  document.title = `${config.title} | AeroXe Nexus AI`;

  setMeta('description', config.description);
  setMeta('og:title', config.title);
  setMeta('og:description', config.description);
  setMeta('og:image', config.image || '/og-default.png');
  setMeta('og:url', config.url || window.location.href);
  setMeta('og:type', config.type || 'website');
  setMeta('twitter:card', 'summary_large_image');
  setMeta('twitter:title', config.title);
  setMeta('twitter:description', config.description);
}

function setMeta(name: string, content: string) {
  let el = document.querySelector(`meta[property="${name}"],meta[name="${name}"]`) as HTMLMetaElement;
  if (!el) {
    el = document.createElement('meta');
    if (name.startsWith('og:')) {
      el.setAttribute('property', name);
    } else {
      el.setAttribute('name', name);
    }
    document.head.appendChild(el);
  }
  el.setAttribute('content', content);
}
```

---

## 18. Performance Monitoring

```typescript
// src/lib/performance-monitor.ts
export function initPerformanceMonitoring() {
  if (typeof window === 'undefined') return;

  // Track Web Vitals
  import('web-vitals').then(({ onCLS, onFCP, onLCP, onTTFB, onINP }) => {
    const report = (metric: any) => {
      fetch('/api/v1/metrics/vitals', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: metric.name,
          value: metric.value,
          rating: metric.rating,
          id: metric.id,
        }),
        keepalive: true,
      });
    };

    onCLS(report);
    onFCP(report);
    onLCP(report);
    onTTFB(report);
    onINP(report);
  });

  // Track navigation timing
  window.addEventListener('load', () => {
    setTimeout(() => {
      const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      if (nav) {
        fetch('/api/v1/metrics/navigation', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            dns: nav.domainLookupEnd - nav.domainLookupStart,
            tcp: nav.connectEnd - nav.connectStart,
            ttfb: nav.responseStart - nav.requestStart,
            download: nav.responseEnd - nav.responseStart,
            domInteractive: nav.domInteractive - nav.fetchStart,
            domComplete: nav.domComplete - nav.fetchStart,
            loadEvent: nav.loadEventEnd - nav.fetchStart,
          }),
          keepalive: true,
        });
      }
    }, 0);
  });
}
```

---

## 19. Integration Checklist

| Integration          | Status | Priority |
|----------------------|--------|----------|
| API Client (Axios)   | Required | P0 |
| JWT Interceptor      | Required | P0 |
| Token Refresh        | Required | P0 |
| WebSocket Client     | Required | P0 |
| Auth Hooks           | Required | P0 |
| Chat Hooks           | Required | P0 |
| Form Validation      | Required | P0 |
| Error Boundary       | Required | P0 |
| Toast Notifications  | Required | P0 |
| Cache (TanStack)     | Required | P1 |
| File Upload          | Required | P1 |
| Date/Number Formatting | Required | P1 |
| Analytics            | Required | P1 |
| Logging              | Required | P1 |
| SEO Metadata         | Required | P2 |
| Feature Flags        | Required | P2 |
| Performance Monitoring | Required | P2 |
| A/B Testing          | Optional | P3 |
