# Error Handling

> Error handling architecture for Nexus AI — boundaries, toasts, validation, recovery, and monitoring.

## 1. Architecture Overview

```
┌───────────────────────────────────────────────────────────────┐
│                    ERROR HANDLING LAYERS                        │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                    Error Boundaries                      │  │
│  │  Page Level → Feature Level → Component Level            │  │
│  └──────────────────────┬──────────────────────────────────┘  │
│                         │                                     │
│  ┌──────────────────────┴──────────────────────────────────┐  │
│  │                   Error Classification                    │  │
│  │  Network │ Auth │ Validation │ Server │ Client │ Timeout │  │
│  └──────────────────────┬──────────────────────────────────┘  │
│                         │                                     │
│  ┌──────────────────────┴──────────────────────────────────┐  │
│  │                   Error Presentation                     │  │
│  │  Toast │ Inline Error │ Error Page │ Modal │ Banner      │  │
│  └──────────────────────┬──────────────────────────────────┘  │
│                         │                                     │
│  ┌──────────────────────┴──────────────────────────────────┐  │
│  │                   Error Monitoring                       │  │
│  │  Sentry │ Error Rate │ Alerts │ Logging                   │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

## 2. Error Types & Classification

| Type | Code Range | Example | Recoverable? |
|---|---|---|---|
| **Network** | 0 | Offline, DNS failure | Yes (retry) |
| **Auth** | 401 | Token expired, invalid credentials | Yes (refresh) |
| **Forbidden** | 403 | Insufficient permissions | No |
| **Validation** | 422 | Invalid form input | Yes (fix input) |
| **Rate Limit** | 429 | Too many requests | Yes (backoff) |
| **Server** | 500 | Internal server error | Yes (retry) |
| **Bad Gateway** | 502 | Service unavailable | Yes (retry) |
| **Timeout** | 408 | Request timeout | Yes (retry) |
| **Client** | N/A | Bug in frontend code | No (needs fix) |

### Error Response Schema

```ts
// types/errors.ts
export interface ApiError {
  status: number;
  code: string;
  message: string;
  details?: Record<string, string[]>;
  requestId?: string;
  timestamp: string;
}

export interface ValidationError {
  field: string;
  message: string;
  code: string;
}

export type ErrorSeverity = "low" | "medium" | "high" | "critical";

export interface AppError {
  id: string;
  type: "network" | "auth" | "validation" | "server" | "client" | "timeout";
  severity: ErrorSeverity;
  message: string;
  originalError?: Error;
  context?: Record<string, unknown>;
  recoverable: boolean;
  retryable: boolean;
  timestamp: number;
}

// Error classification
export function classifyError(error: unknown): AppError {
  if (error instanceof AxiosError) {
    const status = error.response?.status;

    if (!error.response) {
      return {
        id: crypto.randomUUID(),
        type: "network",
        severity: "medium",
        message: "Network error. Please check your connection.",
        originalError: error,
        recoverable: true,
        retryable: true,
        timestamp: Date.now(),
      };
    }

    if (status === 401) {
      return {
        id: crypto.randomUUID(),
        type: "auth",
        severity: "high",
        message: "Your session has expired. Please log in again.",
        originalError: error,
        recoverable: true,
        retryable: false,
        timestamp: Date.now(),
      };
    }

    if (status === 422) {
      return {
        id: crypto.randomUUID(),
        type: "validation",
        severity: "low",
        message: error.response?.data?.message ?? "Validation failed",
        originalError: error,
        recoverable: true,
        retryable: false,
        context: { details: error.response?.data?.details },
        timestamp: Date.now(),
      };
    }

    if (status === 429) {
      return {
        id: crypto.randomUUID(),
        type: "network",
        severity: "medium",
        message: "Too many requests. Please try again later.",
        originalError: error,
        recoverable: true,
        retryable: true,
        timestamp: Date.now(),
      };
    }

    if (status && status >= 500) {
      return {
        id: crypto.randomUUID(),
        type: "server",
        severity: "high",
        message: "Server error. Our team has been notified.",
        originalError: error,
        recoverable: true,
        retryable: true,
        timestamp: Date.now(),
      };
    }
  }

  if (error instanceof Error && error.name === "AbortError") {
    return {
      id: crypto.randomUUID(),
      type: "timeout",
      severity: "medium",
      message: "Request timed out. Please try again.",
      originalError: error,
      recoverable: true,
      retryable: true,
      timestamp: Date.now(),
    };
  }

  // Unknown error
  return {
    id: crypto.randomUUID(),
    type: "client",
    severity: "critical",
    message: "An unexpected error occurred.",
    originalError: error instanceof Error ? error : new Error(String(error)),
    recoverable: false,
    retryable: false,
    timestamp: Date.now(),
  };
}
```

## 3. Error Boundary Architecture

### Page-Level Error Boundary

```tsx
// components/error/PageErrorBoundary.tsx
"use client";

import { Component, type ReactNode } from "react";
import * as Sentry from "@sentry/nextjs";

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class PageErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    Sentry.captureException(error, {
      contexts: {
        react: {
          componentStack: errorInfo.componentStack,
        },
      },
    });

    console.error("[PageErrorBoundary]", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback;
      return <ErrorPage error={this.state.error} />;
    }
    return this.props.children;
  }
}
```

### Feature-Level Error Boundary

```tsx
// components/error/FeatureErrorBoundary.tsx
interface FeatureErrorBoundaryProps {
  children: ReactNode;
  featureName: string;
  fallback?: ReactNode;
}

export class FeatureErrorBoundary extends Component<FeatureErrorBoundaryProps, State> {
  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    Sentry.captureException(error, {
      tags: { feature: this.props.featureName },
      contexts: { react: { componentStack: errorInfo.componentStack } },
    });
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback ?? (
          <div className="p-4 border border-red-200 rounded-xl bg-red-50 dark:bg-red-950/20 dark:border-red-800">
            <div className="flex items-center gap-2 text-red-700 dark:text-red-400">
              <AlertTriangle className="h-5 w-5" />
              <span className="font-medium">Something went wrong with {this.props.featureName}</span>
            </div>
            <p className="text-sm text-red-600 dark:text-red-500 mt-1">
              {this.state.error?.message}
            </p>
            <button
              onClick={() => this.setState({ hasError: false, error: null })}
              className="mt-3 px-3 py-1.5 text-sm font-medium text-red-700
                         bg-red-100 rounded-lg hover:bg-red-200 transition-colors"
            >
              Try again
            </button>
          </div>
        )
      );
    }
    return this.props.children;
  }
}
```

### Component-Level Error Boundary

```tsx
// Inline error boundary for individual components
function SafeComponent({ children }: { children: ReactNode }) {
  return (
    <ErrorBoundary
      fallbackRender={({ error, resetErrorBoundary }) => (
        <div className="p-3 rounded-lg bg-neutral-50 dark:bg-neutral-800 text-sm">
          <p className="text-neutral-600 dark:text-neutral-400">
            This component failed to render.
          </p>
          <button
            onClick={resetErrorBoundary}
            className="mt-2 text-blue-600 hover:underline text-xs"
          >
            Retry
          </button>
        </div>
      )}
    >
      {children}
    </ErrorBoundary>
  );
}
```

### Boundary Placement Strategy

| Level | Scope | Fallback UI | Examples |
|---|---|---|---|
| Page | Entire page | Full error page | `/dashboard` crash |
| Feature | Feature section | Error card with retry | Chat panel, document viewer |
| Component | Single component | Inline error message | Chart, table, form |

## 4. Error Fallback UI

### Full Error Page

```tsx
// components/error/ErrorPage.tsx
interface ErrorPageProps {
  error?: Error | null;
  title?: string;
  description?: string;
  showRetry?: boolean;
  showHome?: boolean;
}

export function ErrorPage({
  error,
  title = "Something went wrong",
  description = "An unexpected error occurred. Please try again or contact support.",
  showRetry = true,
  showHome = true,
}: ErrorPageProps) {
  const router = useRouter();

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="max-w-md w-full text-center space-y-6">
        {/* Illustration */}
        <div className="flex justify-center">
          <div className="w-24 h-24 rounded-full bg-red-100 dark:bg-red-950/30 flex items-center justify-center">
            <AlertTriangle className="h-12 w-12 text-red-500" />
          </div>
        </div>

        {/* Content */}
        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-neutral-900 dark:text-neutral-100">
            {title}
          </h1>
          <p className="text-neutral-600 dark:text-neutral-400">{description}</p>
        </div>

        {/* Error details (dev only) */}
        {process.env.NODE_ENV === "development" && error && (
          <div className="text-left p-4 bg-neutral-100 dark:bg-neutral-800 rounded-xl
                          text-xs font-mono text-neutral-600 dark:text-neutral-400 overflow-auto max-h-40">
            <p className="font-semibold mb-1">{error.name}: {error.message}</p>
            <pre className="whitespace-pre-wrap">{error.stack}</pre>
          </div>
        )}

        {/* Actions */}
        <div className="flex justify-center gap-3">
          {showRetry && (
            <button
              onClick={() => window.location.reload()}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg
                         hover:bg-blue-700 transition-colors font-medium"
            >
              Try again
            </button>
          )}
          {showHome && (
            <button
              onClick={() => router.push("/dashboard")}
              className="px-4 py-2 bg-neutral-200 dark:bg-neutral-700 text-neutral-700
                         dark:text-neutral-300 rounded-lg hover:bg-neutral-300
                         dark:hover:bg-neutral-600 transition-colors font-medium"
            >
              Go to Dashboard
            </button>
          )}
        </div>

        {/* Support link */}
        <p className="text-sm text-neutral-500">
          If this persists,{" "}
          <a href="mailto:support@nexus-ai.com" className="text-blue-600 hover:underline">
            contact support
          </a>
          {" "}or{" "}
          <a
            href="https://github.com/nexus-ai/issues"
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-600 hover:underline"
          >
            report an issue
          </a>
        </p>
      </div>
    </div>
  );
}
```

### 404 Page

```tsx
// app/not-found.tsx
export default function NotFound() {
  const router = useRouter();

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="max-w-md w-full text-center space-y-6">
        <div className="text-8xl font-bold text-neutral-200 dark:text-neutral-700">
          404
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-bold">Page not found</h1>
          <p className="text-neutral-600 dark:text-neutral-400">
            The page you're looking for doesn't exist or has been moved.
          </p>
        </div>
        <SearchBar placeholder="Search for pages..." />
        <div className="flex justify-center gap-3">
          <button
            onClick={() => router.back()}
            className="px-4 py-2 text-sm font-medium border rounded-lg
                       hover:bg-neutral-50 dark:hover:bg-neutral-800 transition-colors"
          >
            Go back
          </button>
          <button
            onClick={() => router.push("/dashboard")}
            className="px-4 py-2 text-sm font-medium bg-blue-600 text-white
                       rounded-lg hover:bg-blue-700 transition-colors"
          >
            Go to Dashboard
          </button>
        </div>
        <div className="pt-4 border-t">
          <p className="text-sm text-neutral-500 mb-2">Popular pages:</p>
          <div className="flex flex-wrap justify-center gap-2">
            {[
              { label: "Dashboard", href: "/dashboard" },
              { label: "Chat", href: "/chat" },
              { label: "Agents", href: "/agents" },
              { label: "Documents", href: "/documents" },
            ].map((link) => (
              <a
                key={link.href}
                href={link.href}
                className="px-3 py-1.5 text-sm bg-neutral-100 dark:bg-neutral-800
                           rounded-full hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors"
              >
                {link.label}
              </a>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
```

### 500 Page

```tsx
// app/error.tsx
"use client";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    Sentry.captureException(error);
  }, [error]);

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="max-w-md w-full text-center space-y-6">
        <div className="flex justify-center">
          <div className="w-24 h-24 rounded-full bg-red-100 dark:bg-red-950/30 flex items-center justify-center">
            <ServerCrash className="h-12 w-12 text-red-500" />
          </div>
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-bold">Server Error</h1>
          <p className="text-neutral-600 dark:text-neutral-400">
            Something went wrong on our end. Our team has been notified.
          </p>
          {error.digest && (
            <p className="text-xs text-neutral-500 font-mono">
              Error ID: {error.digest}
            </p>
          )}
        </div>
        <div className="flex justify-center gap-3">
          <button
            onClick={reset}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg
                       hover:bg-blue-700 transition-colors font-medium"
          >
            Try again
          </button>
          <a
            href="mailto:support@nexus-ai.com?subject=Error%20Report&body=Error%20ID:%20${error.digest}"
            className="px-4 py-2 border rounded-lg font-medium
                       hover:bg-neutral-50 dark:hover:bg-neutral-800 transition-colors"
          >
            Contact Support
          </a>
        </div>
      </div>
    </div>
  );
}
```

## 5. Toast Notification System

```ts
// hooks/useToast.ts
import { create } from "zustand";

interface Toast {
  id: string;
  type: "success" | "error" | "warning" | "info";
  title: string;
  description?: string;
  duration?: number;
  action?: { label: string; onClick: () => void };
  dismissible?: boolean;
}

interface ToastStore {
  toasts: Toast[];
  addToast: (toast: Omit<Toast, "id">) => string;
  removeToast: (id: string) => void;
  dismissAll: () => void;
}

export const useToastStore = create<ToastStore>((set, get) => ({
  toasts: [],

  addToast: (toast) => {
    const id = crypto.randomUUID();
    const duration = toast.duration ?? (toast.type === "error" ? 8000 : 5000);

    set((state) => ({
      toasts: [...state.toasts, { ...toast, id }].slice(-5), // Max 5 visible
    }));

    if (duration > 0) {
      setTimeout(() => get().removeToast(id), duration);
    }

    return id;
  },

  removeToast: (id) =>
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    })),

  dismissAll: () => set({ toasts: [] }),
}));

// Convenience methods
export const toast = {
  success: (title: string, opts?: Partial<Toast>) =>
    useToastStore.getState().addToast({ type: "success", title, ...opts }),
  error: (title: string, opts?: Partial<Toast>) =>
    useToastStore.getState().addToast({ type: "error", title, ...opts }),
  warning: (title: string, opts?: Partial<Toast>) =>
    useToastStore.getState().addToast({ type: "warning", title, ...opts }),
  info: (title: string, opts?: Partial<Toast>) =>
    useToastStore.getState().addToast({ type: "info", title, ...opts }),
};
```

### Toast UI Component

```tsx
// components/ui/ToastContainer.tsx
function ToastContainer() {
  const { toasts, removeToast } = useToastStore();

  return (
    <div
      aria-live="polite"
      aria-label="Notifications"
      className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2
                 max-w-sm w-full pointer-events-none"
    >
      <AnimatePresence mode="popLayout">
        {toasts.map((t) => (
          <motion.div
            key={t.id}
            layout
            initial={{ opacity: 0, y: 20, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, x: 100, scale: 0.95 }}
            transition={{ type: "spring", damping: 25, stiffness: 300 }}
            role={t.type === "error" ? "alert" : "status"}
            className={`pointer-events-auto p-4 rounded-xl shadow-lg border
                       flex items-start gap-3 ${TOAST_STYLES[t.type]}`}
          >
            <ToastIcon type={t.type} />
            <div className="flex-1 min-w-0">
              <p className="font-medium text-sm">{t.title}</p>
              {t.description && (
                <p className="text-xs mt-0.5 opacity-80">{t.description}</p>
              )}
              {t.action && (
                <button
                  onClick={t.action.onClick}
                  className="mt-2 text-xs font-semibold underline"
                >
                  {t.action.label}
                </button>
              )}
            </div>
            <button
              onClick={() => removeToast(t.id)}
              aria-label="Dismiss"
              className="p-1 rounded-lg hover:bg-black/10 dark:hover:bg-white/10
                         transition-colors flex-shrink-0"
            >
              <X className="h-4 w-4" />
            </button>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  );
}

const TOAST_STYLES: Record<string, string> = {
  success: "bg-green-50 border-green-200 text-green-800 dark:bg-green-950 dark:border-green-800 dark:text-green-300",
  error: "bg-red-50 border-red-200 text-red-800 dark:bg-red-950 dark:border-red-800 dark:text-red-300",
  warning: "bg-amber-50 border-amber-200 text-amber-800 dark:bg-amber-950 dark:border-amber-800 dark:text-amber-300",
  info: "bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-950 dark:border-blue-800 dark:text-blue-300",
};
```

### Toast Configuration

| Type | Default Duration | Dismissible | Sound | Position |
|---|---|---|---|---|
| `success` | 5s | ✅ | Success chime | Bottom-right |
| `error` | 8s (no auto-dismiss) | ✅ | Error tone | Bottom-right |
| `warning` | 6s | ✅ | Alert tone | Bottom-right |
| `info` | 5s | ✅ | None | Bottom-right |

## 6. Form Validation Errors

```ts
// lib/validation/schemas.ts
import { z } from "zod";

export const createAgentSchema = z.object({
  name: z
    .string()
    .min(1, "Agent name is required")
    .max(100, "Name must be 100 characters or less"),
  description: z
    .string()
    .max(500, "Description must be 500 characters or less")
    .optional(),
  model: z.string().min(1, "Please select a model"),
  systemPrompt: z
    .string()
    .min(10, "System prompt must be at least 10 characters")
    .max(10000, "System prompt is too long"),
  temperature: z
    .number()
    .min(0, "Temperature must be at least 0")
    .max(2, "Temperature must be at most 2"),
  maxTokens: z
    .number()
    .int()
    .min(1, "Max tokens must be at least 1")
    .max(128000, "Max tokens cannot exceed 128,000"),
  tools: z.array(z.string()).optional(),
});

export type CreateAgentInput = z.infer<typeof createAgentSchema>;
```

### Form Error Display

```tsx
// components/ui/FormError.tsx
interface FormErrorProps {
  errors: Record<string, string[]>;
  fieldName: string;
}

export function FormError({ errors, fieldName }: FormErrorProps) {
  const fieldErrors = errors[fieldName];
  if (!fieldErrors || fieldErrors.length === 0) return null;

  return (
    <div role="alert" aria-live="polite" className="mt-1 space-y-0.5">
      {fieldErrors.map((error, i) => (
        <p key={i} className="text-xs text-red-600 dark:text-red-400 flex items-center gap-1">
          <AlertCircle className="h-3 w-3 flex-shrink-0" />
          {error}
        </p>
      ))}
    </div>
  );
}

// Form with validation
function CreateAgentForm() {
  const [errors, setErrors] = useState<Record<string, string[]>>({});

  const form = useForm<CreateAgentInput>({
    resolver: zodResolver(createAgentSchema),
    mode: "onBlur", // Validate on blur
  });

  const onSubmit = async (data: CreateAgentInput) => {
    try {
      await api.post("/api/agents", data);
      toast.success("Agent created successfully");
    } catch (error) {
      if (isAxiosError(error) && error.response?.status === 422) {
        const validationErrors = error.response.data.details;
        setErrors(validationErrors);

        // Focus first error field
        const firstField = Object.keys(validationErrors)[0];
        if (firstField) {
          form.setFocus(firstField as keyof CreateAgentInput);
        }
      }
    }
  };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <label htmlFor="name" className="block text-sm font-medium mb-1">
          Name <span className="text-red-500">*</span>
        </label>
        <input
          id="name"
          {...form.register("name")}
          aria-invalid={!!form.formState.errors.name || !!errors.name}
          aria-describedby={errors.name ? "name-error" : undefined}
          className="w-full px-3 py-2 border rounded-lg"
        />
        {form.formState.errors.name && (
          <p id="name-error" role="alert" className="text-xs text-red-600 mt-1">
            {form.formState.errors.name.message}
          </p>
        )}
        <FormError errors={errors} fieldName="name" />
      </div>
      {/* ... other fields ... */}
    </form>
  );
}
```

## 7. Network Error Handling

### Retry with Exponential Backoff

```ts
// lib/api/retry.ts
interface RetryConfig {
  maxRetries: number;
  baseDelay: number;
  maxDelay: number;
  retryableStatuses: number[];
}

const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxRetries: 3,
  baseDelay: 1000,
  maxDelay: 30000,
  retryableStatuses: [408, 429, 500, 502, 503, 504],
};

export async function withRetry<T>(
  fn: () => Promise<T>,
  config: Partial<RetryConfig> = {}
): Promise<T> {
  const { maxRetries, baseDelay, maxDelay, retryableStatuses } = {
    ...DEFAULT_RETRY_CONFIG,
    ...config,
  };

  let lastError: unknown;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;

      if (attempt === maxRetries) break;

      if (isAxiosError(error)) {
        const status = error.response?.status;
        const retryAfter = error.response?.headers?.["retry-after"];

        if (status && !retryableStatuses.includes(status)) break;

        const delay = retryAfter
          ? parseInt(retryAfter) * 1000
          : Math.min(baseDelay * 2 ** attempt + Math.random() * 1000, maxDelay);

        console.warn(`[Retry] Attempt ${attempt + 1}/${maxRetries} after ${delay}ms`);

        await new Promise((resolve) => setTimeout(resolve, delay));
      } else {
        break;
      }
    }
  }

  throw lastError;
}

// Axios interceptor with retry
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.config?._retryCount === undefined) {
      error.config._retryCount = 0;
    }

    const { maxRetries = 3, baseDelay = 1000 } = error.config?.retryConfig ?? {};

    if (
      error.config._retryCount < maxRetries &&
      isRetryableError(error)
    ) {
      error.config._retryCount += 1;
      const delay = baseDelay * 2 ** error.config._retryCount;
      await new Promise((resolve) => setTimeout(resolve, delay));
      return api(error.config);
    }

    return Promise.reject(error);
  }
);
```

### Offline Detection

```tsx
// components/ui/OfflineBanner.tsx
function OfflineBanner() {
  const [isOffline, setIsOffline] = useState(!navigator.onLine);
  const [isRetrying, setIsRetrying] = useState(false);

  useEffect(() => {
    const handleOnline = () => {
      setIsOffline(false);
      toast.success("You're back online");
    };
    const handleOffline = () => setIsOffline(true);

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);
    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  if (!isOffline) return null;

  return (
    <div
      role="alert"
      className="fixed top-0 inset-x-0 z-[100] bg-amber-500 text-white
                 text-center text-sm py-2 safe-area-top"
    >
      <div className="flex items-center justify-center gap-2">
        <WifiOff className="h-4 w-4" />
        <span>You are offline. Some features may be unavailable.</span>
        <button
          onClick={async () => {
            setIsRetrying(true);
            try {
              await fetch("/api/health");
              setIsOffline(false);
              toast.success("Connection restored");
            } catch {
              // Still offline
            } finally {
              setIsRetrying(false);
            }
          }}
          disabled={isRetrying}
          className="underline font-medium disabled:opacity-50"
        >
          {isRetrying ? "Checking..." : "Retry"}
        </button>
      </div>
    </div>
  );
}
```

## 8. Auth Error Handling

```ts
// lib/api/authInterceptor.ts
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (error: unknown) => void;
}> = [];

function processQueue(error: unknown, token?: string) {
  failedQueue.forEach((promise) => {
    if (error) {
      promise.reject(error);
    } else {
      promise.resolve(token!);
    }
  });
  failedQueue = [];
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then((token) => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          return api(originalRequest);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const { refreshToken } = useAuthStore.getState();
        if (!refreshToken) throw new Error("No refresh token");

        const response = await axios.post("/api/auth/refresh", { refreshToken });
        const { token, refreshToken: newRefreshToken } = response.data;

        useAuthStore.getState().setToken(token, newRefreshToken);
        api.defaults.headers.common.Authorization = `Bearer ${token}`;

        processQueue(null, token);

        originalRequest.headers.Authorization = `Bearer ${token}`;
        return api(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError, undefined);
        useAuthStore.getState().logout();
        window.location.href = "/login";
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);
```

## 9. WebSocket Error Handling

```ts
// lib/websocket/reconnection.ts
class ReconnectingWebSocket {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private baseDelay = 1000;

  connect(url: string) {
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      console.log("[WS] Connected");
    };

    this.ws.onclose = (event) => {
      if (!event.wasClean) {
        console.warn(`[WS] Connection lost (code: ${event.code})`);
        this.reconnect(url);
      }
    };

    this.ws.onerror = (error) => {
      console.error("[WS] Error:", error);
    };
  }

  private reconnect(url: string) {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error("[WS] Max reconnect attempts reached");
      toast.error("Connection lost. Please refresh the page.");
      return;
    }

    const delay = Math.min(
      this.baseDelay * 2 ** this.reconnectAttempts + Math.random() * 1000,
      30000
    );

    this.reconnectAttempts += 1;

    console.log(
      `[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
    );

    setTimeout(() => this.connect(url), delay);
  }

  close() {
    this.ws?.close(1000, "Client closed");
    this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnect
  }
}
```

## 10. Streaming Error Handling

```ts
// lib/chat/streaming.ts
export async function streamChatResponse(
  messages: Message[],
  onChunk: (chunk: string) => void,
  onError: (error: AppError) => void,
  signal?: AbortSignal
) {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 120_000); // 2min timeout

  try {
    const response = await fetch("/api/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ messages }),
      signal: signal ?? controller.signal,
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message ?? "Chat request failed");
    }

    const reader = response.body!.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    let lastChunkTime = Date.now();
    const TIMEOUT = 30_000; // 30s between chunks

    while (true) {
      const { done, value } = await reader.read();

      if (done) break;

      if (Date.now() - lastChunkTime > TIMEOUT) {
        throw new Error("Stream timed out — no data received for 30 seconds");
      }

      lastChunkTime = Date.now();
      buffer += decoder.decode(value, { stream: true });

      const lines = buffer.split("\n");
      buffer = lines.pop() ?? "";

      for (const line of lines) {
        if (line.startsWith("data: ")) {
          const data = line.slice(6);
          if (data === "[DONE]") return;

          try {
            const parsed = JSON.parse(data);
            if (parsed.error) {
              throw new Error(parsed.error);
            }
            onChunk(parsed.content ?? "");
          } catch (e) {
            if (e instanceof SyntaxError) continue; // Skip malformed JSON
            throw e;
          }
        }
      }
    }
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") {
      onError(classifyError(new Error("Request was cancelled")));
      return;
    }
    onError(classifyError(error));
  } finally {
    clearTimeout(timeoutId);
  }
}
```

## 11. Error Logging & Monitoring

### Sentry Integration

```ts
// lib/sentry.ts
import * as Sentry from "@sentry/nextjs";

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
  environment: process.env.NODE_ENV,
  tracesSampleRate: process.env.NODE_ENV === "production" ? 0.1 : 1.0,
  replaysSessionSampleRate: 0.1,
  replaysOnErrorSampleRate: 1.0,
  integrations: [
    Sentry.replayIntegration(),
  ],
  beforeSend(event) {
    // Strip sensitive data
    if (event.request?.headers) {
      delete event.request.headers["Authorization"];
      delete event.request.headers["Cookie"];
    }
    return event;
  },
});
```

### Error Rate Monitoring

```ts
// lib/monitoring/errorTracker.ts
class ErrorTracker {
  private errors: Map<string, { count: number; lastSeen: number }> = new Map();
  private thresholds = {
    "network-error": { count: 10, window: 60_000 },
    "api-error": { count: 20, window: 60_000 },
    "render-error": { count: 5, window: 60_000 },
  };

  track(type: string, error: AppError) {
    const existing = this.errors.get(type);
    const now = Date.now();

    if (existing && now - existing.lastSeen < this.thresholds[type]?.window) {
      existing.count += 1;
      existing.lastSeen = now;

      if (existing.count >= this.thresholds[type]?.count) {
        this.alert(type, existing.count);
      }
    } else {
      this.errors.set(type, { count: 1, lastSeen: now });
    }
  }

  private alert(type: string, count: number) {
    Sentry.captureMessage(
      `Error rate threshold exceeded: ${type} (${count} occurrences)`,
      "error"
    );
  }
}

export const errorTracker = new ErrorTracker();
```

## 12. Error Recovery Strategies

| Strategy | Use Case | Implementation |
|---|---|---|
| **Retry** | Network, timeout, 5xx | Exponential backoff |
| **Fallback** | Data loading failure | Cached data, placeholder |
| **Graceful degradation** | Feature unavailable | Disable feature, show message |
| **Circuit breaker** | Repeated failures | Stop calling, show cached |
| **User action** | Auth, validation | Guide user to fix |

```tsx
// Error recovery component
function DataWithErrorRecovery<T>({
  queryKey,
  fetchFn,
  renderFn,
  fallbackData,
}: {
  queryKey: string[];
  fetchFn: () => Promise<T>;
  renderFn: (data: T) => React.ReactNode;
  fallbackData?: T;
}) {
  const { data, error, isLoading, refetch, failureCount } = useQuery({
    queryKey,
    queryFn: fetchFn,
    retry: 3,
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 10_000),
  });

  if (isLoading) return <Skeleton />;
  if (error && fallbackData) return <>{renderFn(fallbackData)}</>;
  if (error) {
    return (
      <div className="p-4 border border-red-200 rounded-xl bg-red-50 dark:bg-red-950/20">
        <p className="text-sm text-red-700 dark:text-red-400">
          Failed to load data. Attempted {failureCount} times.
        </p>
        <button
          onClick={() => refetch()}
          className="mt-2 text-sm font-medium text-red-700 underline
                     hover:text-red-800 dark:text-red-400"
        >
          Try again
        </button>
      </div>
    );
  }

  return <>{renderFn(data!)}</>;
}
```

## 13. Error Accessibility

| Requirement | Implementation |
|---|---|
| Error announced to screen readers | `role="alert"`, `aria-live="assertive"` |
| Form errors linked to fields | `aria-describedby` pointing to error |
| Toast notifications | `role="alert"` for errors, `role="status"` for info |
| Focus management | Focus moved to error summary on submit |
| Error contrast | Error text passes AA contrast ratio |
| Not color-only | Icon + text + color for all errors |

```tsx
// Accessible error summary
function ErrorSummary({ errors }: { errors: Record<string, string[]> }) {
  const errorCount = Object.values(errors).flat().length;
  if (errorCount === 0) return null;

  return (
    <div
      role="alert"
      aria-live="assertive"
      className="p-4 bg-red-50 dark:bg-red-950/20 border border-red-200
                 dark:border-red-800 rounded-xl"
      tabIndex={-1}
      ref={(el) => el?.focus()}
    >
      <h3 className="text-sm font-semibold text-red-800 dark:text-red-300">
        {errorCount} {errorCount === 1 ? "error" : "errors"} found
      </h3>
      <ul className="mt-2 space-y-1">
        {Object.entries(errors).map(([field, messages]) =>
          messages.map((msg, i) => (
            <li key={`${field}-${i}`} className="text-sm text-red-700 dark:text-red-400">
              <a href={`#${field}`} className="underline hover:no-underline">
                {field}: {msg}
              </a>
            </li>
          ))
        )}
      </ul>
    </div>
  );
}
```

## 14. Error Testing

```ts
// e2e/error-handling.spec.ts
test.describe("Error Handling", () => {
  test("shows error page when API returns 500", async ({ page }) => {
    await page.route("**/api/agents", (route) =>
      route.fulfill({ status: 500, body: JSON.stringify({ message: "Internal Server Error" }) })
    );
    await page.goto("/agents");
    await expect(page.getByText("Server Error")).toBeVisible();
  });

  test("shows 404 page for unknown routes", async ({ page }) => {
    await page.goto("/unknown-page");
    await expect(page.getByText("Page not found")).toBeVisible();
  });

  test("handles network disconnection gracefully", async ({ page, context }) => {
    await page.goto("/dashboard");
    await context.setOffline(true);
    // Should show offline banner
    await expect(page.getByText("You are offline")).toBeVisible();
  });

  test("form validation shows errors", async ({ page }) => {
    await page.goto("/agents/new");
    await page.click("button[type=submit]");
    await expect(page.getByText("Agent name is required")).toBeVisible();
  });

  test("toast dismisses after timeout", async ({ page }) => {
    // Trigger a success toast
    await page.evaluate(() => {
      (window as any).toast.success("Test notification");
    });
    await expect(page.getByText("Test notification")).toBeVisible();
    await page.waitForTimeout(5500);
    await expect(page.getByText("Test notification")).not.toBeVisible();
  });
});
```

## 15. Error Stores

```ts
// stores/errorStore.ts
import { create } from "zustand";

interface ErrorEntry {
  id: string;
  type: string;
  message: string;
  severity: ErrorSeverity;
  timestamp: number;
  context?: Record<string, unknown>;
  resolved: boolean;
}

interface ErrorState {
  errors: ErrorEntry[];
  recentErrors: ErrorEntry[];
  errorRate: Record<string, number>;
}

interface ErrorActions {
  addError: (type: string, message: string, severity: ErrorSeverity, context?: Record<string, unknown>) => void;
  resolveError: (id: string) => void;
  clearErrors: () => void;
}

export const useErrorStore = create<ErrorState & ErrorActions>()(
  (set) => ({
    errors: [],
    recentErrors: [],
    errorRate: {},

    addError: (type, message, severity, context) =>
      set((state) => {
        const error: ErrorEntry = {
          id: crypto.randomUUID(),
          type,
          message,
          severity,
          timestamp: Date.now(),
          context,
          resolved: false,
        };
        return {
          errors: [...state.errors, error],
          recentErrors: [...state.recentErrors, error].slice(-50),
        };
      }),

    resolveError: (id) =>
      set((state) => ({
        errors: state.errors.map((e) =>
          e.id === id ? { ...e, resolved: true } : e
        ),
      })),

    clearErrors: () => set({ errors: [], recentErrors: [], errorRate: {} }),
  })
);
```

## 16. File Structure

```
src/
├── components/
│   ├── error/
│   │   ├── PageErrorBoundary.tsx
│   │   ├── FeatureErrorBoundary.tsx
│   │   ├── ErrorPage.tsx
│   │   ├── NotFound.tsx
│   │   ├── ServerError.tsx
│   │   └── ErrorSummary.tsx
│   └── ui/
│       ├── ToastContainer.tsx
│       ├── FormError.tsx
│       └── OfflineBanner.tsx
├── hooks/
│   ├── useToast.ts
│   ├── useRetry.ts
│   └── useError.ts
├── stores/
│   ├── errorStore.ts
│   └── toastStore.ts
├── lib/
│   ├── api/
│   │   ├── retry.ts
│   │   ├── authInterceptor.ts
│   │   └── errorHandler.ts
│   ├── websocket/
│   │   └── reconnection.ts
│   ├── chat/
│   │   └── streaming.ts
│   ├── monitoring/
│   │   ├── errorTracker.ts
│   │   └── sentry.ts
│   └── validation/
│       └── schemas.ts
├── types/
│   └── errors.ts
└── e2e/
    └── error-handling.spec.ts
```
