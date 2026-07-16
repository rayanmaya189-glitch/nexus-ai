# 22 — Frontend Security

## XSS + CSRF + CSP + Token Security + Input Validation + Secure Headers

---

## 1. Security Model

AeroXe Nexus AI frontend follows a **defense-in-depth** security model. No single security control is trusted alone. Multiple layers protect against attack vectors.

```
┌─────────────────────────────────────────────────┐
│                  Browser                         │
│                                                  │
│  ┌──────────────────────────────────────────┐   │
│  │  Content Security Policy (CSP)           │   │
│  │  X-Frame-Options                         │   │
│  │  Referrer-Policy                         │   │
│  ├──────────────────────────────────────────┤   │
│  │  Input Validation (Zod)                  │   │
│  │  Output Encoding (React auto-escape)     │   │
│  │  Sanitization (DOMPurify if needed)      │   │
│  ├──────────────────────────────────────────┤   │
│  │  Auth Token (HttpOnly cookie / memory)   │   │
│  │  CSRF Protection (SameSite cookies)      │   │
│  │  Secure Headers                          │   │
│  ├──────────────────────────────────────────┤   │
│  │  HTTPS Only (HSTS)                       │   │
│  │  Subresource Integrity (SRI)             │   │
│  └──────────────────────────────────────────┘   │
│                     │                           │
│                     ▼                           │
│              API Gateway                        │
└─────────────────────────────────────────────────┘
```

---

## 2. XSS Prevention

### React Auto-Escaping

React automatically escapes all rendered content. This prevents most XSS attacks by default.

```tsx
// SAFE: React auto-escapes
function SafeComponent({ userInput }: { userInput: string }) {
  return <div>{userInput}</div>; // Always safe
}

// SAFE: Attribute values are also escaped
function SafeLink({ url }: { url: string }) {
  return <a href={url}>Link</a>; // Safe in React
}

// DANGEROUS: Never use unless absolutely necessary
function DangerousComponent({ html }: { html: string }) {
  // Only use with sanitized content!
  return <div dangerouslySetInnerHTML={{ __html: domPurify.sanitize(html) }} />;
}
```

### Input Sanitization

```typescript
// src/lib/sanitize.ts
import DOMPurify from 'dompurify';

// Sanitize HTML content (for rich text displays)
export function sanitizeHTML(dirty: string): string {
  return DOMPurify.sanitize(dirty, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'p', 'br', 'ul', 'ol', 'li', 'code', 'pre'],
    ALLOWED_ATTR: [],
  });
}

// Sanitize plain text (strip all HTML)
export function sanitizeText(dirty: string): string {
  return DOMPurify.sanitize(dirty, { ALLOWED_TAGS: [] });
}

// Sanitize URL (prevent javascript: protocol)
export function sanitizeURL(url: string): string {
  const parsed = new URL(url, window.location.origin);
  if (!['http:', 'https:', 'mailto:'].includes(parsed.protocol)) {
    return '#';
  }
  return parsed.href;
}

// Sanitize file name
export function sanitizeFileName(name: string): string {
  return name
    .replace(/[^a-zA-Z0-9._-]/g, '_')
    .replace(/_{2,}/g, '_')
    .slice(0, 255);
}
```

### XSS Prevention Checklist

| Vector | Protection | Implementation |
|--------|-----------|----------------|
| HTML injection | React auto-escape | Default React behavior |
| Script injection | CSP script-src | Strict CSP header |
| Event handler XSS | CSP + React | No inline event handlers |
| DOM-based XSS | DOMPurify | For `dangerouslySetInnerHTML` only |
| URL XSS | URL sanitization | Validate protocols |
| CSS XSS | CSP style-src | Restrict inline styles |

---

## 3. CSRF Protection

### SameSite Cookies

```
Set-Cookie: access_token=xxx; SameSite=Strict; Secure; HttpOnly; Path=/
Set-Cookie: refresh_token=xxx; SameSite=Strict; Secure; HttpOnly; Path=/api/v1/auth
```

### CSRF Token Double-Submit

```typescript
// src/lib/csrf.ts
import { apiClient } from './api-client';

let csrfToken: string | null = null;

export async function getCSRFToken(): Promise<string> {
  if (csrfToken) return csrfToken;

  const response = await fetch('/api/v1/csrf-token', {
    credentials: 'include',
  });
  const data = await response.json();
  csrfToken = data.token;
  return csrfToken;
}

// Add to all state-changing requests
apiClient.interceptors.request.use(async (config) => {
  if (['post', 'put', 'patch', 'delete'].includes(config.method || '')) {
    const token = await getCSRFToken();
    config.headers['X-CSRF-Token'] = token;
  }
  return config;
});
```

---

## 4. Content Security Policy (CSP)

### CSP Headers

```typescript
// next.config.js (Vercel) or server middleware
const securityHeaders = [
  {
    key: 'Content-Security-Policy',
    value: [
      "default-src 'self'",
      "script-src 'self' 'nonce-{NONCE}' https://cdn.aeroxe.com",
      "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
      "img-src 'self' data: https://cdn.aeroxe.com https://*.tile.openstreetmap.org",
      "font-src 'self' https://fonts.gstatic.com",
      "connect-src 'self' https://api.aeroxe.com wss://ws.aeroxe.com",
      "frame-src 'self' https://widget.aeroxe.com",
      "frame-ancestors 'none'",
      "base-uri 'self'",
      "form-action 'self'",
      "object-src 'none'",
      "upgrade-insecure-requests",
    ].join('; '),
  },
  {
    key: 'X-Frame-Options',
    value: 'DENY',
  },
  {
    key: 'X-Content-Type-Options',
    value: 'nosniff',
  },
  {
    key: 'X-XSS-Protection',
    value: '1; mode=block',
  },
  {
    key: 'Referrer-Policy',
    value: 'strict-origin-when-cross-origin',
  },
  {
    key: 'Permissions-Policy',
    value: 'camera=(), microphone=(self), geolocation=()',
  },
  {
    key: 'Strict-Transport-Security',
    value: 'max-age=31536000; includeSubDomains; preload',
  },
];
```

### CSP Directives Reference

| Directive | Value | Purpose |
|-----------|-------|---------|
| `default-src` | `'self'` | Default for all resources |
| `script-src` | `'self' nonce-xxx` | Only allow scripts with nonce |
| `style-src` | `'self' 'unsafe-inline'` | Allow styles (shadcn needs inline) |
| `img-src` | `'self' data: https` | Allow images from trusted sources |
| `connect-src` | `'self' https://api.aeroxe.com` | API and WebSocket connections |
| `frame-ancestors` | `'none'` | Prevent framing (clickjacking) |
| `base-uri` | `'self'` | Prevent base tag injection |
| `form-action` | `'self'` | Prevent form hijacking |

---

## 5. Token Security

### Token Storage Strategy

| Token | Storage | Rationale |
|-------|---------|-----------|
| Access Token | Memory (Zustand) | Not accessible via XSS in HttpOnly cookies |
| Refresh Token | HttpOnly cookie | Server-only access, never in JS |
| CSRF Token | Memory | Short-lived, per-session |
| Session ID | HttpOnly cookie | Server-only |

### Token Refresh Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│ Frontend │     │ API GW   │     │ Auth Svc │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │  Request + Token                │
     │──────────────>│                 │
     │               │  Validate       │
     │               │────────────────>│
     │               │  401 Unauthorized
     │               │<────────────────│
     │               │                 │
     │               │  Refresh Token  │
     │               │────────────────>│
     │               │  New Tokens     │
     │               │<────────────────│
     │  200 + Response│                │
     │<──────────────│                 │
```

### Token Rotation

```typescript
// Refresh token rotation with reuse detection
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = useAuthStore.getState().refreshToken;

        // Rotate: use current refresh token, get new pair
        const { data } = await axios.post('/api/v1/auth/refresh', {
          refresh_token: refreshToken,
        });

        // If server returns error about reused token, force logout
        if (data.error === 'Token reuse detected') {
          useAuthStore.getState().logout();
          window.location.href = '/login';
          return Promise.reject(error);
        }

        useAuthStore.getState().setToken(data.access_token);
        useAuthStore.getState().setRefreshToken(data.refresh_token);

        originalRequest.headers.Authorization = `Bearer ${data.access_token}`;
        return apiClient(originalRequest);
      } catch {
        useAuthStore.getState().logout();
        window.location.href = '/login';
        return Promise.reject(error);
      }
    }
    return Promise.reject(error);
  }
);
```

### Secure Cookie Configuration

```
Set-Cookie:
  refresh_token=xxx;
  HttpOnly
  Secure
  SameSite=Strict
  Path=/api/v1/auth
  Max-Age=604800
  Domain=.aeroxe.com
```

---

## 6. Input Validation

### Zod Schemas

```typescript
// src/lib/validations.ts
import { z } from 'zod';

// Login validation
export const loginSchema = z.object({
  email: z
    .string()
    .email('Invalid email format')
    .max(255, 'Email too long')
    .toLowerCase()
    .trim(),
  password: z
    .string()
    .min(8, 'Password must be at least 8 characters')
    .max(128, 'Password too long')
    .regex(/[A-Z]/, 'Password must contain an uppercase letter')
    .regex(/[a-z]/, 'Password must contain a lowercase letter')
    .regex(/[0-9]/, 'Password must contain a number')
    .regex(/[^A-Za-z0-9]/, 'Password must contain a special character'),
});

// Chat message validation
export const chatMessageSchema = z.object({
  content: z
    .string()
    .min(1, 'Message cannot be empty')
    .max(4000, 'Message too long')
    .trim(),
  agent_id: z.string().uuid('Invalid agent ID'),
  conversation_id: z.string().uuid().optional(),
});

// File upload validation
export const fileUploadSchema = z.object({
  file: z
    .instanceof(File)
    .refine((f) => f.size <= 10 * 1024 * 1024, 'File must be under 10MB')
    .refine(
      (f) => [
        'image/png', 'image/jpeg', 'image/gif', 'image/webp',
        'application/pdf',
        'text/plain', 'text/csv',
      ].includes(f.type),
      'File type not allowed'
    ),
});

// User profile validation
export const profileSchema = z.object({
  name: z.string().min(1).max(100).trim(),
  email: z.string().email().max(255),
  phone: z.string().regex(/^\+?[1-9]\d{1,14}$/, 'Invalid phone number').optional(),
});
```

### Validation Middleware

```typescript
// src/hooks/use-validated-form.ts
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ZodSchema } from 'zod';

export function useValidatedForm<T extends ZodSchema>(schema: T) {
  return useForm({
    resolver: zodResolver(schema),
    mode: 'onBlur', // Validate on blur for better UX
  });
}
```

---

## 7. Clickjacking Protection

```typescript
// CSP frame-ancestors prevents framing
// X-Frame-Options as backup
// Additionally, frame-busting script as defense-in-depth

// src/lib/frame-buster.ts
if (window.self !== window.top) {
  window.top!.location = window.self.location;
}
```

---

## 8. Open Redirect Prevention

```typescript
// src/lib/url-utils.ts
const ALLOWED_REDIRECT_HOSTS = [
  'aeroxe.com',
  'app.aeroxe.com',
  'localhost',
];

export function sanitizeRedirect(url: string): string {
  try {
    const parsed = new URL(url, window.location.origin);

    // Only allow relative paths or same-origin redirects
    if (parsed.origin !== window.location.origin) {
      // Check if host is in allowlist
      const hostAllowed = ALLOWED_REDIRECT_HOSTS.some(
        (h) => parsed.hostname === h || parsed.hostname.endsWith(`.${h}`)
      );
      if (!hostAllowed) {
        return '/dashboard'; // Default safe redirect
      }
    }

    // Prevent protocol-relative URLs
    if (parsed.protocol === 'javascript:') {
      return '/dashboard';
    }

    return parsed.pathname + parsed.search;
  } catch {
    return '/dashboard';
  }
}
```

---

## 9. File Upload Security

```typescript
// src/lib/file-security.ts
const ALLOWED_MIME_TYPES = [
  'image/png', 'image/jpeg', 'image/gif', 'image/webp',
  'application/pdf',
  'text/plain', 'text/csv',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
];

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
const MAGIC_BYTES: Record<string, number[]> = {
  'image/png': [0x89, 0x50, 0x4e, 0x47],
  'image/jpeg': [0xff, 0xd8, 0xff],
  'application/pdf': [0x25, 0x50, 0x44, 0x46],
};

export async function validateFile(file: File): Promise<{ valid: boolean; error?: string }> {
  // Size check
  if (file.size > MAX_FILE_SIZE) {
    return { valid: false, error: `File exceeds ${MAX_FILE_SIZE / 1024 / 1024}MB limit` };
  }

  // MIME type check
  if (!ALLOWED_MIME_TYPES.includes(file.type)) {
    return { valid: false, error: `File type ${file.type} not allowed` };
  }

  // Magic byte verification
  const expectedBytes = MAGIC_BYTES[file.type];
  if (expectedBytes) {
    const buffer = await file.slice(0, 4).arrayBuffer();
    const bytes = new Uint8Array(buffer);
    const matches = expectedBytes.every((b, i) => bytes[i] === b);
    if (!matches) {
      return { valid: false, error: 'File content does not match declared type' };
    }
  }

  // Dangerous extension check
  const dangerousExtensions = ['.exe', '.bat', '.cmd', '.sh', '.ps1', '.vbs', '.js', '.msi'];
  const ext = '.' + file.name.split('.').pop()?.toLowerCase();
  if (dangerousExtensions.includes(ext)) {
    return { valid: false, error: `File extension ${ext} not allowed` };
  }

  return { valid: true };
}
```

---

## 10. Sensitive Data Handling

### Never Store Secrets in Frontend

```typescript
// WRONG: Never do this
const API_SECRET = 'sk_live_xxxxxxxxxxxx'; // Exposed in browser

// CORRECT: Environment variables (bundled at build time, not secrets)
const API_URL = process.env.NEXT_PUBLIC_API_URL; // Non-secret config only

// CORRECT: Secrets stay on server
// API secret key is only in the backend .env, never exposed to frontend
```

### PII Masking

```typescript
// src/lib/pii-masking.ts
export function maskEmail(email: string): string {
  const [local, domain] = email.split('@');
  if (!domain) return email;
  const masked = local[0] + '***' + local[local.length - 1];
  return `${masked}@${domain}`;
}

export function maskPhone(phone: string): string {
  if (phone.length < 4) return '****';
  return phone.slice(0, 2) + '****' + phone.slice(-2);
}

export function maskCardNumber(number: string): string {
  return '****-****-****-' + number.slice(-4);
}
```

---

## 11. Dependency Security

### npm Audit

```bash
# Check for vulnerabilities
npm audit

# Fix automatically
npm audit fix

# Force fix (breaking changes possible)
npm audit fix --force
```

### Dependabot Configuration

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "daily"
    open-pull-requests-limit: 10
    labels: ["dependencies", "security"]
    reviewers: ["security-team"]
```

### Lockfile Integrity

```bash
# Verify lockfile integrity
npm ci --ignore-scripts

# Use npm ci instead of npm install in CI
# This ensures exact dependency resolution from lockfile
```

---

## 12. Security Headers Summary

| Header | Value | Purpose |
|--------|-------|---------|
| `Content-Security-Policy` | Strict policy | Prevent XSS, data injection |
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-XSS-Protection` | `1; mode=block` | Legacy XSS protection |
| `Strict-Transport-Security` | `max-age=31536000` | Force HTTPS |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer info |
| `Permissions-Policy` | `camera=(), microphone=(self)` | Control browser features |

---

## 13. Secure Authentication Flow

### Password Policy

```typescript
// Enforced by both frontend and backend
const PASSWORD_POLICY = {
  minLength: 8,
  maxLength: 128,
  requireUppercase: true,
  requireLowercase: true,
  requireNumber: true,
  requireSpecial: true,
  disallowCommon: true, // Check against common password list
  disallowPersonalInfo: true, // No name/email in password
};
```

### Account Lockout

```typescript
// Frontend shows lockout message
// Backend enforces lockout after 5 failed attempts
const LOCKOUT_CONFIG = {
  maxAttempts: 5,
  lockoutDuration: 15 * 60 * 1000, // 15 minutes
  resetAttemptsAfter: 30 * 60 * 1000, // 30 minutes of inactivity
};
```

### Session Management

```typescript
// Frontend session handling
class SessionManager {
  private timeout: ReturnType<typeof setTimeout> | null = null;
  private readonly SESSION_TIMEOUT = 30 * 60 * 1000; // 30 minutes

  start() {
    this.timeout = setTimeout(() => {
      useAuthStore.getState().logout();
      window.location.href = '/login?reason=timeout';
    }, this.SESSION_TIMEOUT);
  }

  reset() {
    if (this.timeout) clearTimeout(this.timeout);
    this.start();
  }

  stop() {
    if (this.timeout) clearTimeout(this.timeout);
  }
}
```

---

## 14. API Security

### Request Signing

```typescript
// HMAC signature for API requests
import CryptoJS from 'crypto-js';

function signRequest(method: string, path: string, body: string, timestamp: string): string {
  const message = `${method}\n${path}\n${timestamp}\n${body}`;
  return CryptoJS.HmacSHA256(message, API_SECRET).toString();
}
```

### Rate Limiting (Client-Side)

```typescript
// src/lib/rate-limiter.ts
class RateLimiter {
  private attempts: Map<string, number[]> = new Map();

  isAllowed(key: string, maxAttempts: number, windowMs: number): boolean {
    const now = Date.now();
    const attempts = this.attempts.get(key) || [];
    const recentAttempts = attempts.filter((t) => now - t < windowMs);

    if (recentAttempts.length >= maxAttempts) {
      return false;
    }

    recentAttempts.push(now);
    this.attempts.set(key, recentAttempts);
    return true;
  }
}

export const loginRateLimiter = new RateLimiter();
// 5 attempts per 15 minutes
// loginRateLimiter.isAllowed('login', 5, 15 * 60 * 1000)
```

---

## 15. Security Monitoring

### Client-Side Error Reporting

```typescript
// src/lib/error-reporter.ts
export function initErrorReporting() {
  // Global error handler
  window.addEventListener('error', (event) => {
    reportError({
      type: 'javascript_error',
      message: event.message,
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      stack: event.error?.stack,
    });
  });

  // Unhandled promise rejection
  window.addEventListener('unhandledrejection', (event) => {
    reportError({
      type: 'unhandled_rejection',
      message: String(event.reason),
      stack: event.reason?.stack,
    });
  });
}

function reportError(error: Record<string, unknown>) {
  // Don't report in development
  if (process.env.NODE_ENV === 'development') {
    console.error('[Error Report]', error);
    return;
  }

  // Send to error reporting service
  fetch('/api/v1/errors/report', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      ...error,
      url: window.location.href,
      userAgent: navigator.userAgent,
      timestamp: new Date().toISOString(),
    }),
    keepalive: true,
  });
}
```

### Suspicious Activity Detection

```typescript
// Detect potential XSS attempts in URL
function detectSuspiciousActivity() {
  const url = window.location.href;
  const suspiciousPatterns = [
    /javascript:/i,
    /on\w+\s*=/i,
    /<script/i,
    /data:text\/html/i,
  ];

  const isSuspicious = suspiciousPatterns.some((p) => p.test(url));
  if (isSuspicious) {
    reportError({
      type: 'suspicious_activity',
      message: 'Potentially malicious URL detected',
      url,
    });
    window.location.href = '/'; // Redirect to safe page
  }
}

detectSuspiciousActivity();
```

---

## 16. Security Testing

### SAST (Static Application Security Testing)

```yaml
# .github/workflows/security.yml
name: Security Scan
on: [push, pull_request]

jobs:
  sast:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run ESLint Security Plugin
        run: npx eslint --plugin security src/
      - name: Run Semgrep
        uses: semgrep/semgrep-action@v1
        with:
          config: p/typescript
```

### DAST (Dynamic Application Security Testing)

```yaml
  dast:
    runs-on: ubuntu-latest
    needs: deploy-staging
    steps:
      - name: Run OWASP ZAP
        uses: zaproxy/action-full-scan@v0.8.0
        with:
          target: https://staging.aeroxe.com
          rules_file_name: '.zap/rules.tsv'
```

---

## 17. Security Hooks

```typescript
// src/hooks/use-secure-storage.ts
export function useSecureStorage() {
  const set = (key: string, value: string) => {
    // Use sessionStorage for sensitive data (cleared on tab close)
    // Never use localStorage for tokens or secrets
    sessionStorage.setItem(key, value);
  };

  const get = (key: string): string | null => {
    return sessionStorage.getItem(key);
  };

  const remove = (key: string) => {
    sessionStorage.removeItem(key);
  };

  const clear = () => {
    sessionStorage.clear();
  };

  return { set, get, remove, clear };
}

// src/hooks/use-csp.ts
export function useCSP() {
  useEffect(() => {
    // Monitor CSP violations
    document.addEventListener('securitypolicyviolation', (event) => {
      fetch('/api/v1/security/csp-violation', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          violatedDirective: event.violatedDirective,
          blockedURI: event.blockedURI,
          documentURI: event.documentURI,
          originalPolicy: event.originalPolicy,
        }),
      });
    });
  }, []);
}
```

---

## 18. Security Checklist

| Category | Check | Status |
|----------|-------|--------|
| **XSS** | React auto-escape for all output | Required |
| **XSS** | DOMPurify for any HTML rendering | Required |
| **XSS** | No `eval()` or `new Function()` | Required |
| **CSRF** | SameSite=Strict cookies | Required |
| **CSRF** | CSRF token for state-changing requests | Required |
| **CSP** | Strict Content-Security-Policy header | Required |
| **CSP** | Nonce-based script loading | Required |
| **Tokens** | Access token in memory only | Required |
| **Tokens** | Refresh token in HttpOnly cookie | Required |
| **Tokens** | Token rotation on refresh | Required |
| **Input** | Zod validation on all forms | Required |
| **Input** | Server-side validation (never trust client) | Required |
| **Files** | MIME type + magic byte verification | Required |
| **Files** | Size limits enforced client + server | Required |
| **Headers** | All security headers configured | Required |
| **HTTPS** | HSTS with preload | Required |
| **Dependencies** | npm audit in CI | Required |
| **Dependencies** | Dependabot enabled | Required |
| **Errors** | Global error reporting | Required |
| **Monitoring** | CSP violation reporting | Required |
| **Auth** | Password policy enforced | Required |
| **Auth** | Account lockout after failed attempts | Required |
| **Session** | Timeout after inactivity | Required |
| **PII** | Mask sensitive data in UI | Required |
| **Secrets** | No secrets in frontend code | Required |
