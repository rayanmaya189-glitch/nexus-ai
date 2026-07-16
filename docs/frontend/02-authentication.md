# AeroXe Nexus AI — Authentication & Authorization

## Login, Registration, JWT, RBAC, ABAC, Multi-Tenant Auth, KYC, Session Management

---

## 1. Authentication Overview

```
┌──────────────────────────────────────────────────────────┐
│                  Authentication Architecture               │
│                                                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │  Login    │  │ Register │  │  OTP     │  │  SSO     │ │
│  │  Page     │  │  Page    │  │ Verify   │  │ (Future) │ │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘ │
│        │              │              │              │       │
│        └──────────────┴──────┬───────┴──────────────┘       │
│                              │                              │
│                    ┌─────────▼─────────┐                    │
│                    │   Auth Store       │                    │
│                    │   (Zustand)        │                    │
│                    └─────────┬─────────┘                    │
│                              │                              │
│                    ┌─────────▼─────────┐                    │
│                    │  API Client        │                    │
│                    │  (JWT Interceptor) │                    │
│                    └─────────┬─────────┘                    │
│                              │                              │
│                    ┌─────────▼─────────┐                    │
│                    │  identity-service  │                    │
│                    │  (Backend API)     │                    │
│                    └───────────────────┘                    │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

### Supported Methods

| Method | Status | Description |
|---|---|---|
| Email + Password | Active | Primary authentication |
| OTP (Email/SMS) | Active | Two-factor verification |
| SSO / OAuth2 | Future | Single sign-on integration |
| Biometric (Mobile) | Future | Fingerprint, Face ID |

---

## 2. Login Page

### 2.1 Page Design

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│                   [Nexus Logo]                        │
│                                                      │
│            Welcome back to AeroXe Nexus AI           │
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │                                                │   │
│  │  Email Address                                 │   │
│  │  ┌──────────────────────────────────────────┐ │   │
│  │  │ admin@aeroxenexus.com                    │ │   │
│  │  └──────────────────────────────────────────┘ │   │
│  │                                                │   │
│  │  Password                                      │   │
│  │  ┌──────────────────────────────────────────┐ │   │
│  │  │ ••••••••••••••                   👁       │ │   │
│  │  └──────────────────────────────────────────┘ │   │
│  │                                                │   │
│  │  ☐ Remember me              Forgot password?   │   │
│  │                                                │   │
│  │  ┌──────────────────────────────────────────┐ │   │
│  │  │           Sign In                         │ │   │
│  │  └──────────────────────────────────────────┘ │   │
│  │                                                │   │
│  │  ──────────── OR ────────────                  │   │
│  │                                                │   │
│  │  ┌──────────────────────────────────────────┐ │   │
│  │  │     🔑 Continue with SSO (Coming Soon)    │ │   │
│  │  └──────────────────────────────────────────┘ │   │
│  │                                                │   │
│  │  Don't have an account? Sign up               │   │
│  │                                                │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  © 2026 AeroXe. All rights reserved.                │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 2.2 Login Form Component

```tsx
// components/forms/LoginForm.tsx
"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { useAuth } from "@/hooks/auth/useAuth";
import { Eye, EyeOff, Loader2 } from "lucide-react";
import { useState } from "react";

const loginSchema = z.object({
  email: z.string().email("Please enter a valid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
  rememberMe: z.boolean().default(false),
});

type LoginFormData = z.infer<typeof loginSchema>;

export function LoginForm() {
  const { login, isLoading } = useAuth();
  const [showPassword, setShowPassword] = useState(false);

  const form = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "", rememberMe: false },
  });

  const onSubmit = async (data: LoginFormData) => {
    try {
      await login(data.email, data.password, data.rememberMe);
    } catch (error) {
      if (error.code === "INVALID_CREDENTIALS") {
        form.setError("email", { message: "Invalid email or password" });
      } else if (error.code === "ACCOUNT_LOCKED") {
        form.setError("email", { message: "Account locked. Try again in 15 minutes." });
      } else if (error.code === "KYC_PENDING") {
        // Redirect to KYC page
      }
    }
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email Address</FormLabel>
              <FormControl>
                <Input type="email" placeholder="admin@aeroxenexus.com" {...field} />
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
                <div className="relative">
                  <Input
                    type={showPassword ? "text" : "password"}
                    placeholder="••••••••••••"
                    {...field}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="absolute right-2 top-1/2 -translate-y-1/2"
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </Button>
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex items-center justify-between">
          <FormField
            control={form.control}
            name="rememberMe"
            render={({ field }) => (
              <FormItem className="flex items-center space-x-2">
                <FormControl>
                  <Checkbox checked={field.value} onCheckedChange={field.onChange} />
                </FormControl>
                <FormLabel className="text-sm font-normal">Remember me</FormLabel>
              </FormItem>
            )}
          />
          <a href="/forgot-password" className="text-sm text-accent hover:underline">
            Forgot password?
          </a>
        </div>

        <Button type="submit" className="w-full" disabled={isLoading}>
          {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Sign In
        </Button>
      </form>
    </Form>
  );
}
```

---

## 3. Registration Page

### 3.1 Registration Flow

```
┌──────────────────────────────────────────────────────────┐
│                                                            │
│  Step 1          Step 2          Step 3          Step 4    │
│  ─────          ─────          ─────          ─────       │
│  Account Info → Organization → Verify Email → Complete     │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

### 3.2 Registration Form Schema

```typescript
const registerSchema = z.object({
  // Step 1: Account Info
  fullName: z.string().min(2, "Name is required"),
  email: z.string().email("Valid email required"),
  password: z.string()
    .min(12, "Password must be at least 12 characters")
    .regex(/[A-Z]/, "Must contain uppercase")
    .regex(/[a-z]/, "Must contain lowercase")
    .regex(/[0-9]/, "Must contain a number")
    .regex(/[^A-Za-z0-9]/, "Must contain a symbol"),
  confirmPassword: z.string(),
  acceptTerms: z.literal(true, {
    errorMap: () => ({ message: "You must accept the terms" }),
  }),

  // Step 2: Organization
  companyName: z.string().min(2, "Company name is required"),
  companySize: z.enum(["1-10", "11-50", "51-200", "201-1000", "1000+"]),
  industry: z.string().min(1, "Industry is required"),
  useCase: z.string().min(10, "Please describe your use case"),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});
```

### 3.3 Registration Success State

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│                   ✅ Account Created                  │
│                                                      │
│          We've sent a verification email to           │
│          admin@aeroxenexus.com                        │
│                                                      │
│          Please check your inbox and click            │
│          the verification link to activate            │
│          your account.                                │
│                                                      │
│          Didn't receive the email? Resend             │
│                                                      │
│          ┌────────────────────────────────────────┐  │
│          │          Open Email Client              │  │
│          └────────────────────────────────────────┘  │
│                                                      │
└──────────────────────────────────────────────────────┘
```

---

## 4. OTP Verification Page

### 4.1 OTP Flow

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│                   [Nexus Logo]                        │
│                                                      │
│              Verify Your Identity                     │
│                                                      │
│          Enter the 6-digit code sent to               │
│          admin@aeroxenexus.com                        │
│                                                      │
│          ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐      │
│          │ 3 │ │ 7 │ │ _ │ │ _ │ │ _ │ │ _ │      │
│          └───┘ └───┘ └───┘ └───┘ └───┘ └───┘      │
│                                                      │
│          ┌────────────────────────────────────────┐  │
│          │           Verify Code                   │  │
│          └────────────────────────────────────────┘  │
│                                                      │
│          Resend code in 00:45                        │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 4.2 OTP Input Component

```tsx
interface OTPInputProps {
  length: 6;
  onComplete: (otp: string) => void;
  error?: string;
}

export function OTPInput({ length = 6, onComplete, error }: OTPInputProps) {
  const [digits, setDigits] = useState<string[]>(new Array(length).fill(""));
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const handleChange = (index: number, value: string) => {
    if (!/^\d$/.test(value) && value !== "") return;

    const newDigits = [...digits];
    newDigits[index] = value;
    setDigits(newDigits);

    // Auto-advance to next input
    if (value && index < length - 1) {
      inputRefs.current[index + 1]?.focus();
    }

    // Check if complete
    const otp = newDigits.join("");
    if (otp.length === length && !newDigits.includes("")) {
      onComplete(otp);
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    if (e.key === "Backspace" && !digits[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const pasted = e.clipboardData.getData("text").slice(0, length);
    if (/^\d+$/.test(pasted)) {
      const newDigits = pasted.split("").concat(new Array(length - pasted.length).fill(""));
      setDigits(newDigits);
      inputRefs.current[Math.min(pasted.length, length - 1)]?.focus();
      if (pasted.length === length) onComplete(pasted);
    }
  };

  return (
    <div className="flex gap-2">
      {digits.map((digit, i) => (
        <Input
          key={i}
          ref={(el) => { inputRefs.current[i] = el; }}
          type="text"
          inputMode="numeric"
          maxLength={1}
          value={digit}
          onChange={(e) => handleChange(i, e.target.value)}
          onKeyDown={(e) => handleKeyDown(i, e)}
          onPaste={handlePaste}
          className="w-12 h-12 text-center text-lg font-mono"
          aria-label={`Digit ${i + 1}`}
        />
      ))}
    </div>
  );
}
```

---

## 5. JWT Token Management

### 5.1 Token Structure

```typescript
interface AccessTokenPayload {
  sub: string;          // User UUID
  tenant_id: string;    // Tenant UUID
  roles: string[];      // ["admin", "ai_operator"]
  permissions: string[];// ["ai.execute", "document.read"]
  email: string;        // admin@aeroxenexus.com
  iat: number;          // Issued at (Unix timestamp)
  exp: number;          // Expires at (Unix timestamp)
  iss: string;          // "aeroxe-nexus-ai"
}
```

### 5.2 Token Lifecycle

```
┌─────────────────────────────────────────────────────┐
│                Token Lifecycle Flow                    │
│                                                       │
│  Login                                               │
│    │                                                 │
│    ├─► Access Token (1 hour)                         │
│    │     └─► Stored in memory (Zustand)              │
│    │     └─► Attached to API requests (Header)       │
│    │                                                 │
│    └─► Refresh Token (7 days)                        │
│          └─► Stored in HTTP-only secure cookie        │
│          └─► Used for silent token refresh            │
│                                                       │
│  Before Access Token Expires:                        │
│    │                                                 │
│    ├─► Auto-refresh (5 min before expiry)            │
│    │     POST /api/v1/auth/refresh                   │
│    │     ├─► New access token returned               │
│    │     ├─► New refresh token (rotation)            │
│    │     └─► Zustand state updated                   │
│    │                                                 │
│    └─► If refresh fails:                             │
│          ├─► Redirect to login                       │
│          └─► Clear all tokens                        │
│                                                       │
│  Logout:                                             │
│    │                                                 │
│    ├─► POST /api/v1/auth/logout                      │
│    ├─► Server invalidates refresh token              │
│    ├─► Client clears Zustand state                   │
│    └─► Redirect to /login                            │
│                                                       │
└─────────────────────────────────────────────────────┘
```

### 5.3 Token Refresh Implementation

```typescript
// lib/token.ts
const REFRESH_THRESHOLD_MS = 5 * 60 * 1000; // 5 minutes before expiry

export function isTokenExpiringSoon(token: string): boolean {
  const payload = decodeJWT(token);
  if (!payload?.exp) return true;
  const expiresAt = payload.exp * 1000;
  return Date.now() >= expiresAt - REFRESH_THRESHOLD_MS;
}

export function decodeJWT(token: string): AccessTokenPayload | null {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64).split('').map(c =>
        '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
      ).join('')
    );
    return JSON.parse(jsonPayload);
  } catch {
    return null;
  }
}

// Auto-refresh interceptor
let refreshPromise: Promise<boolean> | null = null;

export async function refreshAccessToken(apiClient: AxiosInstance): Promise<boolean> {
  // Deduplicate concurrent refresh requests
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const response = await apiClient.post('/api/v1/auth/refresh');
      const { access_token, refresh_token } = response.data;

      useAuthStore.getState().setTokens(access_token, refresh_token);
      return true;
    } catch (error) {
      useAuthStore.getState().clearAuth();
      window.location.href = '/login';
      return false;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}
```

---

## 6. Session Management

### 6.1 Session Storage Strategy

| Data | Storage | Lifetime | Security |
|---|---|---|---|
| Access Token | Zustand (memory) | Until page reload | Never persisted to disk |
| Refresh Token | HTTP-only cookie | 7 days | Secure, SameSite=Strict, HttpOnly |
| User Profile | Zustand (memory) | Session | Re-fetched on mount |
| Theme Preference | localStorage | Until cleared | Non-sensitive |
| Remember Me | localStorage | 30 days | Non-sensitive |
| Tenant Context | Zustand (memory) | Session | Re-fetched on mount |

### 6.2 Session Handling

```typescript
// stores/auth-store.ts
interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  login: (email: string, password: string, rememberMe: boolean) => Promise<void>;
  logout: () => Promise<void>;
  setTokens: (accessToken: string, refreshToken: string) => void;
  clearAuth: () => void;
  checkAndRefreshToken: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: null,
  isAuthenticated: false,
  isLoading: true,

  login: async (email, password, rememberMe) => {
    const response = await authApi.login({ email, password });
    set({
      user: response.user,
      accessToken: response.access_token,
      isAuthenticated: true,
      isLoading: false,
    });
  },

  logout: async () => {
    try {
      await authApi.logout();
    } finally {
      set({
        user: null,
        accessToken: null,
        isAuthenticated: false,
        isLoading: false,
      });
      window.location.href = '/login';
    }
  },

  setTokens: (accessToken) => {
    set({ accessToken });
  },

  clearAuth: () => {
    set({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: false,
    });
  },
}));
```

---

## 7. Password Management

### 7.1 Change Password Flow

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│  Settings > Security > Change Password                │
│                                                      │
│  Current Password                                    │
│  ┌──────────────────────────────────────────────┐   │
│  │ ••••••••••••••••••                         👁  │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  New Password                                        │
│  ┌──────────────────────────────────────────────┐   │
│  │ ••••••••••••••••••                         👁  │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  Password Strength: ██████████ Strong                │
│                                                      │
│  Requirements:                                       │
│  ✅ At least 12 characters                          │
│  ✅ Contains uppercase letter                       │
│  ✅ Contains lowercase letter                       │
│  ✅ Contains a number                               │
│  ✅ Contains a symbol                               │
│                                                      │
│  Confirm New Password                                │
│  ┌──────────────────────────────────────────────┐   │
│  │ ••••••••••••••••••                         👁  │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │         Update Password                        │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  ⚠ You will be logged out from all other sessions.   │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### 7.2 Password Reset Flow

```
Request Reset                    Enter Code                    Set New Password
────────────                    ──────────                    ────────────────

┌──────────────┐               ┌──────────────┐            ┌──────────────┐
│ Enter email  │  ─────────►  │ Enter 6-digit│  ────────► │ Enter new    │
│              │  Email sent   │ code from    │  Code      │ password     │
│ Send Reset   │               │ email        │  verified  │              │
│ Link         │               │              │            │ Reset done   │
└──────────────┘               └──────────────┘            └──────────────┘
```

### 7.3 Password Validation Schema

```typescript
const passwordSchema = z.string()
  .min(12, "Password must be at least 12 characters")
  .regex(/[A-Z]/, "Must contain at least one uppercase letter")
  .regex(/[a-z]/, "Must contain at least one lowercase letter")
  .regex(/[0-9]/, "Must contain at least one number")
  .regex(/[^A-Za-z0-9]/, "Must contain at least one symbol");

const changePasswordSchema = z.object({
  currentPassword: z.string().min(1, "Current password is required"),
  newPassword: passwordSchema,
  confirmPassword: z.string(),
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
}).refine((data) => data.currentPassword !== data.newPassword, {
  message: "New password must be different from current password",
  path: ["newPassword"],
});
```

---

## 8. Role-Based Access Control (RBAC) UI

### 8.1 Role Definitions

| Role | Description | UI Capabilities |
|---|---|---|
| `SUPER_ADMIN` | Platform-wide admin | All features, admin panel |
| `TENANT_ADMIN` | Tenant-level admin | All tenant features, user management |
| `AI_OPERATOR` | AI management | Agent CRUD, chat, knowledge |
| `DEVELOPER` | Developer access | Code, API, testing tools |
| `CUSTOMER_SUPPORT` | Support staff | Chat, tickets, customer data |
| `USER` | Standard user | Chat, document read |
| `AUDITOR` | Read-only audit | Audit logs, reports |

### 8.2 Permission-Based UI Rendering

```tsx
// hooks/auth/usePermission.ts
export function usePermission(permission: string): boolean {
  const { user } = useAuthStore();
  return user?.permissions.includes(permission) ?? false;
}

// hooks/auth/useRole.ts
export function useRole(role: string): boolean {
  const { user } = useAuthStore();
  return user?.roles.includes(role) ?? false;
}

// Usage in components
function AgentManagementSection() {
  const canManageAgents = usePermission("ai.manage");
  const isAdmin = useRole("TENANT_ADMIN") || useRole("SUPER_ADMIN");

  return (
    <div>
      {canManageAgents && (
        <Button onClick={() => createAgent()}>Create Agent</Button>
      )}
      {isAdmin && (
        <Button variant="destructive" onClick={() => deleteAgent()}>
          Delete Agent
        </Button>
      )}
    </div>
  );
}
```

### 8.3 Conditional Rendering Patterns

```tsx
// Pattern 1: Permission gate component
function PermissionGate({ permission, children, fallback }: {
  permission: string;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}) {
  const hasPermission = usePermission(permission);
  return hasPermission ? <>{children}</> : <>{fallback}</>;
}

// Usage
<PermissionGate permission="ai.manage" fallback={<LockIcon />}>
  <Button>Configure Agent</Button>
</PermissionGate>

// Pattern 2: Role-based navigation items
const navigationItems = [
  { label: "Dashboard", href: "/dashboard", requiredPermission: null },
  { label: "Agents", href: "/dashboard/agents", requiredPermission: "ai.manage" },
  { label: "Documents", href: "/dashboard/documents", requiredPermission: "document.read" },
  { label: "Audit", href: "/dashboard/audit", requiredPermission: "audit.read" },
  { label: "Settings", href: "/dashboard/settings", requiredPermission: "admin.manage" },
].filter(item => !item.requiredPermission || usePermission(item.requiredPermission));
```

### 8.4 Full Permission Matrix

| Permission | SUPER_ADMIN | TENANT_ADMIN | AI_OPERATOR | DEVELOPER | SUPPORT | USER | AUDITOR |
|---|---|---|---|---|---|---|---|
| `ai.execute` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `ai.manage` | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| `document.read` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `document.write` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `document.delete` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `customer.read` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ |
| `customer.write` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ |
| `ticket.create` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ |
| `knowledge.search` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `audit.read` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| `admin.manage` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

---

## 9. Auth Guards

### 9.1 Route Protection (Middleware)

```typescript
// middleware.ts (Next.js middleware)
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const publicRoutes = ["/login", "/register", "/forgot-password", "/reset-password"];
const kycRequiredRoutes = ["/dashboard/agents", "/dashboard/chat"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Allow public routes
  if (publicRoutes.some(route => pathname.startsWith(route))) {
    return NextResponse.next();
  }

  // Check for session cookie
  const session = request.cookies.get("session");
  if (!session) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }

  // Check KYC status for protected routes
  if (kycRequiredRoutes.some(route => pathname.startsWith(route))) {
    const kycStatus = request.cookies.get("kyc_status");
    if (kycStatus?.value !== "approved") {
      return NextResponse.redirect(new URL("/dashboard/kyc", request.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
};
```

### 9.2 Component-Level Protection

```tsx
// components/shared/AuthGuard.tsx
function AuthGuard({ children, fallback }: {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}) {
  const { isAuthenticated, isLoading } = useAuthStore();

  if (isLoading) return <LoadingSpinner />;
  if (!isAuthenticated) return <>{fallback ?? <Redirect to="/login" />}</>;
  return <>{children}</>;
}

// Usage in layout
export default function DashboardLayout({ children }) {
  return (
    <AuthGuard>
      <AppShell>
        {children}
      </AppShell>
    </AuthGuard>
  );
}
```

---

## 10. Multi-Tenant Authentication

### 10.1 Tenant Switching

```tsx
// components/layout/TenantSwitcher.tsx
function TenantSwitcher() {
  const { currentTenant, availableTenants, switchTenant } = useTenantStore();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className="w-full justify-start gap-2">
          <Building2 className="h-4 w-4" />
          <span className="truncate">{currentTenant?.name}</span>
          <ChevronDown className="h-4 w-4 ml-auto" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuLabel>Switch Tenant</DropdownMenuLabel>
        {availableTenants.map((tenant) => (
          <DropdownMenuItem
            key={tenant.id}
            onClick={() => switchTenant(tenant.id)}
            className={tenant.id === currentTenant?.id ? "bg-accent" : ""}
          >
            <Building2 className="mr-2 h-4 w-4" />
            {tenant.name}
            {tenant.id === currentTenant?.id && <Check className="ml-auto h-4 w-4" />}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
```

### 10.2 Tenant Context

```typescript
// stores/tenant-store.ts
interface TenantState {
  currentTenant: Tenant | null;
  availableTenants: Tenant[];
  switchTenant: (tenantId: string) => Promise<void>;
}

export const useTenantStore = create<TenantState>((set, get) => ({
  currentTenant: null,
  availableTenants: [],

  switchTenant: async (tenantId) => {
    // API call to switch context
    await api.post(`/api/v1/tenants/${tenantId}/switch`);

    // Update all cached data for new tenant
    queryClient.clear();

    // Update store
    const tenant = get().availableTenants.find(t => t.id === tenantId);
    set({ currentTenant: tenant });

    // Reload page to apply new tenant context
    window.location.reload();
  },
}));
```

---

## 11. KYC Verification Flow

### 11.1 KYC States

```
PENDING → DOCUMENTS_SUBMITTED → UNDER_REVIEW → APPROVED
                                                → REJECTED
                                                → REQUIRES_ADDITIONAL_INFO
```

### 11.2 KYC Page Layout

```
┌──────────────────────────────────────────────────────────┐
│                                                            │
│  KYC Verification                              Status: ⏳  │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  Step 1: Business Registration                       │ │
│  │  ✅ Uploaded: business-registration.pdf              │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  Step 2: Tax ID / GST Certificate                    │ │
│  │  ✅ Uploaded: gst-certificate.pdf                    │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  Step 3: Director/Owner ID                           │ │
│  │  ┌────────────────────────────────────────────────┐  │ │
│  │  │  📁 Drop files here or click to upload          │  │ │
│  │  │     Supports: PDF, JPG, PNG (max 5MB)           │  │ │
│  │  └────────────────────────────────────────────────┘  │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  Step 4: Address Proof                               │ │
│  │  📁 Drop files here or click to upload                │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                            │
│  ┌──────────────────────────────────────────────┐        │
│  │        Submit for Review                        │        │
│  └──────────────────────────────────────────────┘        │
│                                                            │
│  ── KYC Enforcement ──                                    │
│  Pre-KYC: AI Chat ❌ | Documents ❌ | Agents ❌           │
│  Post-KYC: AI Chat ✅ | Documents ✅ | Agents ✅          │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

### 11.3 KYC Upload Component

```tsx
function KYCUploadForm() {
  const { uploadDocument, submitForReview } = useKYC();
  const [documents, setDocuments] = useState<KYCDocument[]>([]);

  const requiredDocs = [
    { type: "business_registration", label: "Business Registration", required: true },
    { type: "tax_id", label: "Tax ID / GST Certificate", required: true },
    { type: "director_id", label: "Director/Owner ID", required: true },
    { type: "address_proof", label: "Address Proof", required: true },
  ];

  return (
    <div className="space-y-4">
      {requiredDocs.map((doc) => (
        <Card key={doc.type}>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              {getDocStatusIcon(doc.type)}
              {doc.label}
              {doc.required && <Badge variant="destructive">Required</Badge>}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {getDocument(doc.type) ? (
              <div className="flex items-center gap-2">
                <FileText className="h-4 w-4" />
                <span>{getDocument(doc.type)?.filename}</span>
                <Badge variant="secondary">Uploaded</Badge>
              </div>
            ) : (
              <FileUploader
                accept=".pdf,.jpg,.jpeg,.png"
                maxSize={5 * 1024 * 1024} // 5MB
                onUpload={(file) => uploadDocument(doc.type, file)}
              />
            )}
          </CardContent>
        </Card>
      ))}

      <Button
        onClick={submitForReview}
        disabled={!allRequiredDocsUploaded()}
        className="w-full"
      >
        Submit for Review
      </Button>
    </div>
  );
}
```

---

## 12. Auth API Integration

### 12.1 API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/auth/login` | User login |
| `POST` | `/api/v1/auth/register` | User registration |
| `POST` | `/api/v1/auth/refresh` | Token refresh |
| `POST` | `/api/v1/auth/logout` | User logout |
| `GET` | `/api/v1/auth/me` | Get current user |
| `POST` | `/api/v1/auth/change-password` | Change password |
| `POST` | `/api/v1/auth/forgot-password` | Request password reset |
| `POST` | `/api/v1/auth/reset-password` | Reset with token |
| `POST` | `/api/v1/auth/verify-otp` | Verify OTP code |
| `GET` | `/api/v1/kyc/status` | Get KYC status |
| `POST` | `/api/v1/kyc/documents` | Upload KYC document |
| `POST` | `/api/v1/kyc/submit` | Submit KYC for review |

### 12.2 API Client Functions

```typescript
// api/auth.api.ts
export const authApi = {
  login: (data: LoginRequest) =>
    apiClient.post<LoginResponse>("/api/v1/auth/login", data),

  register: (data: RegisterRequest) =>
    apiClient.post<RegisterResponse>("/api/v1/auth/register", data),

  refresh: () =>
    apiClient.post<RefreshResponse>("/api/v1/auth/refresh"),

  logout: () =>
    apiClient.post("/api/v1/auth/logout"),

  getMe: () =>
    apiClient.get<User>("/api/v1/auth/me"),

  changePassword: (data: ChangePasswordRequest) =>
    apiClient.post("/api/v1/auth/change-password", data),

  forgotPassword: (email: string) =>
    apiClient.post("/api/v1/auth/forgot-password", { email }),

  resetPassword: (token: string, password: string) =>
    apiClient.post("/api/v1/auth/reset-password", { token, password }),

  verifyOTP: (otp: string) =>
    apiClient.post("/api/v1/auth/verify-otp", { otp }),
};
```

---

## 13. Auth Error Handling

### 13.1 Error Codes

| Code | HTTP Status | Message | Action |
|---|---|---|---|
| `INVALID_CREDENTIALS` | 401 | Invalid email or password | Show form error |
| `ACCOUNT_LOCKED` | 423 | Account locked | Show lockout timer |
| `KYC_PENDING` | 403 | KYC verification pending | Redirect to KYC |
| `KYC_REJECTED` | 403 | KYC verification rejected | Show rejection reason |
| `TOKEN_EXPIRED` | 401 | Session expired | Auto-refresh or login |
| `TOKEN_INVALID` | 401 | Invalid token | Force logout |
| `RATE_LIMITED` | 429 | Too many attempts | Show backoff timer |
| `EMAIL_NOT_VERIFIED` | 403 | Email not verified | Resend verification |

### 13.2 Error Display

```tsx
// Error toast messages
const authErrorMessages: Record<string, string> = {
  INVALID_CREDENTIALS: "Invalid email or password. Please try again.",
  ACCOUNT_LOCKED: "Your account has been locked due to too many failed attempts. Try again in 15 minutes.",
  KYC_PENDING: "Please complete KYC verification to access AI features.",
  KYC_REJECTED: "Your KYC verification was rejected. Please re-submit with updated documents.",
  TOKEN_EXPIRED: "Your session has expired. Please log in again.",
  RATE_LIMITED: "Too many login attempts. Please wait before trying again.",
  NETWORK_ERROR: "Connection failed. Please check your internet connection.",
};
```

---

## 14. Auth State Persistence

### 14.1 State Initialization

```typescript
// App initialization flow
async function initializeAuth() {
  const store = useAuthStore.getState();

  // 1. Check if refresh token cookie exists
  const hasRefreshToken = document.cookie.includes("refresh_token");

  if (!hasRefreshToken) {
    store.setLoading(false);
    return;
  }

  // 2. Try to refresh access token
  try {
    const response = await authApi.refresh();
    store.setTokens(response.data.access_token, response.data.refresh_token);

    // 3. Fetch user profile
    const userResponse = await authApi.getMe();
    store.setUser(userResponse.data);
    store.setAuthenticated(true);
  } catch {
    store.clearAuth();
  } finally {
    store.setLoading(false);
  }
}

// Run on app mount
initializeAuth();
```

### 14.2 Token Refresh Timer

```typescript
// hooks/auth/useTokenRefresh.ts
export function useTokenRefresh() {
  const { accessToken, checkAndRefreshToken } = useAuthStore();

  useEffect(() => {
    if (!accessToken) return;

    const interval = setInterval(async () => {
      if (isTokenExpiringSoon(accessToken)) {
        await checkAndRefreshToken();
      }
    }, 60_000); // Check every minute

    return () => clearInterval(interval);
  }, [accessToken, checkAndRefreshToken]);
}
```

---

## 15. Complete Auth Hooks Reference

| Hook | Purpose | Returns |
|---|---|---|
| `useAuth()` | Login, logout, user state | `{ user, login, logout, isLoading, isAuthenticated }` |
| `usePermission(permission)` | Check single permission | `boolean` |
| `useRole(role)` | Check single role | `boolean` |
| `usePermissions()` | Get all user permissions | `string[]` |
| `useRoles()` | Get all user roles | `string[]` |
| `useTenant()` | Current tenant context | `{ tenant, switchTenant, available }` |
| `useKYC()` | KYC operations | `{ status, upload, submit }` |
| `useTokenRefresh()` | Auto token refresh | `void` (side effect) |

---

*Document Version: 1.0 — Last Updated: July 2026*
