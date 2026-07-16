# 23 — Deployment & DevOps

## Vercel + Docker + Kubernetes + CI/CD + Monitoring + Rollback

---

## 1. Deployment Architecture

```
┌──────────────────────────────────────────────────────────┐
│                       Git Push                           │
│                          │                               │
│                   ┌──────▼──────┐                        │
│                   │ GitHub Actions│                       │
│                   │ CI Pipeline   │                       │
│                   └──────┬──────┘                        │
│                    ┌─────┴─────┐                         │
│                    │           │                          │
│              ┌─────▼───┐ ┌────▼────┐                    │
│              │ Lint &  │ │  Unit   │                     │
│              │ TypeCheck│ │  Tests  │                     │
│              └─────┬───┘ └────┬────┘                    │
│                    │           │                          │
│              ┌─────▼───┐ ┌────▼────┐                    │
│              │   E2E   │ │ Visual  │                     │
│              │  Tests  │ │ Tests   │                     │
│              └─────┬───┘ └────┬────┘                    │
│                    │           │                          │
│                    └─────┬─────┘                         │
│                          │                               │
│              ┌───────────▼───────────┐                   │
│              │     Build (Vite/Next) │                    │
│              └───────────┬───────────┘                   │
│                          │                               │
│         ┌────────────────┼────────────────┐              │
│         │                │                │               │
│   ┌─────▼─────┐   ┌─────▼─────┐   ┌─────▼─────┐       │
│   │  Vercel   │   │  Docker   │   │ Kubernetes │       │
│   │  (Cloud)  │   │  (Self)   │   │  (Cluster) │       │
│   └─────┬─────┘   └─────┬─────┘   └─────┬─────┘       │
│         │                │                │               │
│   ┌─────▼─────────────────▼────────────────▼─────┐      │
│   │              CDN (Cloudflare/CloudFront)      │      │
│   └──────────────────────┬───────────────────────┘      │
│                          │                               │
│                   ┌──────▼──────┐                        │
│                   │   Users     │                        │
│                   └─────────────┘                        │
└──────────────────────────────────────────────────────────┘
```

---

## 2. Vercel Deployment

### next.config.js

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  // Performance
  reactStrictMode: true,
  swcMinify: true,

  // Image optimization
  images: {
    domains: ['cdn.aeroxe.com', 'avatars.aeroxe.com'],
    formats: ['image/avif', 'image/webp'],
    minimumCacheTTL: 60 * 60 * 24 * 30, // 30 days
  },

  // Headers
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: securityHeaders,
      },
      {
        source: '/_next/static/(.*)',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
        ],
      },
    ];
  },

  // Redirects
  async redirects() {
    return [
      { source: '/chat', destination: '/dashboard/chat', permanent: true },
      { source: '/admin', destination: '/dashboard/admin', permanent: true },
    ];
  },

  // Rewrites (proxy API)
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.BACKEND_URL}/api/:path*`,
      },
    ];
  },

  // Output standalone for Docker
  output: 'standalone',
};

module.exports = nextConfig;
```

### Vercel Configuration

```json
{
  "framework": "nextjs",
  "buildCommand": "npm run build",
  "outputDirectory": ".next",
  "installCommand": "npm ci",
  "regions": ["sin1"],
  "functions": {
    "app/api/**/*.ts": {
      "maxDuration": 30
    }
  },
  "crons": [
    {
      "path": "/api/cron/health",
      "schedule": "*/5 * * * *"
    }
  ]
}
```

### Environment Variables (Vercel)

| Variable | Environment | Description |
|----------|------------|-------------|
| `NEXT_PUBLIC_API_URL` | All | Backend API base URL |
| `NEXT_PUBLIC_WS_URL` | All | WebSocket URL |
| `NEXT_PUBLIC_AEROXE_API_KEY` | All | Widget API key |
| `NEXT_PUBLIC_GA_ID` | Production | Google Analytics ID |
| `NEXT_PUBLIC_SENTRY_DSN` | All | Sentry DSN |
| `SENTRY_AUTH_TOKEN` | Build | Sentry upload token |

---

## 3. Docker Deployment

### Dockerfile (Multi-Stage)

```dockerfile
# Stage 1: Dependencies
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --only=production

# Stage 2: Build
FROM node:20-alpine AS builder
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN npm run build

# Stage 3: Production
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  frontend:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - NEXT_PUBLIC_API_URL=http://api-gateway:8000
      - NEXT_PUBLIC_WS_URL=ws://api-gateway:8000/ws
    depends_on:
      - api-gateway
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - frontend
    restart: unless-stopped

  api-gateway:
    image: aeroxe/api-gateway:latest
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql://...
      - REDIS_URL=redis://redis:6379
    depends_on:
      - redis
      - postgres

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  postgres:
    image: postgres:18-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=nexus_ai
      - POSTGRES_USER=aeroxe
      - POSTGRES_PASSWORD_FILE=/run/secrets/db_password
    volumes:
      - pg_data:/var/lib/postgresql/data
    secrets:
      - db_password

volumes:
  redis_data:
  pg_data:

secrets:
  db_password:
    file: ./secrets/db_password.txt
```

### Nginx Configuration

```nginx
upstream frontend {
    server frontend:3000;
}

upstream api {
    server api-gateway:8000;
}

server {
    listen 80;
    server_name app.aeroxe.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name app.aeroxe.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security headers
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Gzip
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml;
    gzip_min_length 256;

    # Frontend
    location / {
        proxy_pass http://frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API proxy
    location /api/ {
        proxy_pass http://api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 30s;
    }

    # WebSocket proxy
    location /ws {
        proxy_pass http://api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400;
    }

    # Static assets cache
    location /_next/static/ {
        proxy_pass http://frontend;
        expires 365d;
        add_header Cache-Control "public, immutable";
    }
}
```

---

## 4. Kubernetes Deployment

### Deployment

```yaml
# k8s/frontend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nexus-frontend
  namespace: aeroxe
  labels:
    app: nexus-frontend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nexus-frontend
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
  template:
    metadata:
      labels:
        app: nexus-frontend
    spec:
      containers:
        - name: frontend
          image: aeroxe/nexus-frontend:latest
          ports:
            - containerPort: 3000
          env:
            - name: NODE_ENV
              value: "production"
            - name: NEXT_PUBLIC_API_URL
              value: "http://api-gateway.aeroxe.svc.cluster.local:8000"
            - name: NEXT_PUBLIC_WS_URL
              value: "ws://api-gateway.aeroxe.svc.cluster.local:8000/ws"
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 5
            periodSeconds: 10
```

### Service

```yaml
# k8s/frontend-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: nexus-frontend
  namespace: aeroxe
spec:
  selector:
    app: nexus-frontend
  ports:
    - port: 80
      targetPort: 3000
  type: ClusterIP
```

### Ingress

```yaml
# k8s/frontend-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nexus-frontend
  namespace: aeroxe
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - app.aeroxe.com
      secretName: aeroxe-tls
  rules:
    - host: app.aeroxe.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: nexus-frontend
                port:
                  number: 80
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: api-gateway
                port:
                  number: 8000
          - path: /ws
            pathType: Prefix
            backend:
              service:
                name: api-gateway
                port:
                  number: 8000
```

### HPA (Horizontal Pod Autoscaler)

```yaml
# k8s/frontend-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: nexus-frontend
  namespace: aeroxe
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: nexus-frontend
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

### ConfigMap

```yaml
# k8s/frontend-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nexus-frontend-config
  namespace: aeroxe
data:
  NEXT_PUBLIC_API_URL: "https://api.aeroxe.com"
  NEXT_PUBLIC_WS_URL: "wss://ws.aeroxe.com"
  NEXT_PUBLIC_WIDGET_CDN: "https://cdn.aeroxe.com/widget"
```

### Secrets

```yaml
# k8s/frontend-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: nexus-frontend-secrets
  namespace: aeroxe
type: Opaque
stringData:
  NEXT_PUBLIC_AEROXE_API_KEY: "ak_live_xxx"
  NEXT_PUBLIC_SENTRY_DSN: "https://xxx@sentry.io/123"
```

---

## 5. Environment Configuration

### Environment Tiers

| Tier | URL | Backend | Database | Purpose |
|------|-----|---------|----------|---------|
| Development | localhost:3000 | localhost:8000 | Local | Local dev |
| Staging | staging.aeroxe.com | api-staging.aeroxe.com | Staging | Testing |
| Production | app.aeroxe.com | api.aeroxe.com | Production | Live |

### .env Files

```bash
# .env.development
NEXT_PUBLIC_API_URL=http://localhost:8000
NEXT_PUBLIC_WS_URL=ws://localhost:8000/ws
NEXT_PUBLIC_ENV=development
NEXT_PUBLIC_DEBUG=true

# .env.staging
NEXT_PUBLIC_API_URL=https://api-staging.aeroxe.com
NEXT_PUBLIC_WS_URL=wss://ws-staging.aeroxe.com
NEXT_PUBLIC_ENV=staging
NEXT_PUBLIC_DEBUG=true

# .env.production
NEXT_PUBLIC_API_URL=https://api.aeroxe.com
NEXT_PUBLIC_WS_URL=wss://ws.aeroxe.com
NEXT_PUBLIC_ENV=production
NEXT_PUBLIC_DEBUG=false
```

### Environment Validation

```typescript
// src/lib/env.ts
import { z } from 'zod';

const envSchema = z.object({
  NEXT_PUBLIC_API_URL: z.string().url(),
  NEXT_PUBLIC_WS_URL: z.string().url(),
  NEXT_PUBLIC_ENV: z.enum(['development', 'staging', 'production']),
  NEXT_PUBLIC_DEBUG: z.enum(['true', 'false']).transform((v) => v === 'true'),
  NEXT_PUBLIC_GA_ID: z.string().optional(),
  NEXT_PUBLIC_SENTRY_DSN: z.string().url().optional(),
});

export const env = envSchema.parse(process.env);
```

---

## 6. CI/CD Pipeline (GitHub Actions)

```yaml
# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  NODE_VERSION: 20
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # ─── Quality Gate ───
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
      - run: npm ci
      - run: npm run lint
      - run: npm run typecheck
      - run: npm run test:unit -- --coverage
      - name: Check coverage threshold
        run: npx vitest run --coverage --reporter=json && node scripts/check-coverage.js

  # ─── Build ───
  build:
    runs-on: ubuntu-latest
    needs: quality
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
      - run: npm ci
      - run: npm run build
      - uses: actions/upload-artifact@v4
        with:
          name: build-output
          path: .next/

  # ─── E2E Tests ───
  e2e:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
      - run: npm ci
      - run: npx playwright install --with-deps
      - uses: actions/download-artifact@v4
        with:
          name: build-output
          path: .next/
      - run: npm run test:e2e
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: test-results/

  # ─── Deploy Staging ───
  deploy-staging:
    runs-on: ubuntu-latest
    needs: [build, e2e]
    if: github.ref == 'refs/heads/main'
    environment: staging
    steps:
      - uses: actions/checkout@v4
      - uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: '--pre'

  # ─── Deploy Production ───
  deploy-production:
    runs-on: ubuntu-latest
    needs: deploy-staging
    if: github.ref == 'refs/heads/main'
    environment: production
    steps:
      - uses: actions/checkout@v4
      - uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: '--prod'

  # ─── Docker Build (Self-hosted) ───
  docker:
    runs-on: ubuntu-latest
    needs: [build, e2e]
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
```

---

## 7. Build Optimization

### Vite Configuration

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig({
  plugins: [
    react(),
    visualizer({
      open: true,
      gzipSize: true,
      filename: 'build-analysis.html',
    }),
  ],
  build: {
    target: 'es2020',
    minify: 'esbuild',
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          ui: ['@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu'],
          charts: ['recharts'],
          forms: ['react-hook-form', '@hookform/resolvers', 'zod'],
          query: ['@tanstack/react-query'],
          state: ['zustand'],
        },
      },
    },
    chunkSizeWarningLimit: 500,
  },
  optimizeDeps: {
    include: ['react', 'react-dom', 'zustand'],
  },
});
```

### Bundle Size Budget

| Chunk | Max Size (gzipped) |
|-------|-------------------|
| vendor (React) | 45 KB |
| ui (shadcn) | 30 KB |
| charts (Recharts) | 40 KB |
| forms (RHF + Zod) | 15 KB |
| query (TanStack) | 12 KB |
| state (Zustand) | 5 KB |
| app code | 50 KB |
| **Total** | **~200 KB** |

---

## 8. Static Asset Deployment

### Cache Headers

```
# Immutable assets (fingerprinted)
/_next/static/** → Cache-Control: public, max-age=31536000, immutable

# HTML pages
/ → Cache-Control: no-cache, must-revalidate

# API responses
/api/** → Cache-Control: private, no-store

# Static files
/public/** → Cache-Control: public, max-age=86400
```

### CDN Configuration

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  User    │────>│   CDN    │────>│  Origin  │
│          │     │ (Edge)   │     │ (Vercel) │
└──────────┘     └──────────┘     └──────────┘
                      │
              Cache Hit? ──Yes──> Return cached
                      │
                     No
                      │
              Fetch from origin
              Cache at edge
              Return to user
```

---

## 9. Domain Configuration

### DNS Records

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A | app.aeroxe.com | 76.76.21.21 (Vercel) | 300 |
| CNAME | www | app.aeroxe.com | 300 |
| CNAME | cdn | cname.vercel-dns.com | 300 |
| TXT | _vercel | vc-domain-verify=xxx | 300 |

### SSL

- Vercel provides automatic SSL certificates
- Custom domain SSL via Let's Encrypt
- HSTS preload enabled

---

## 10. Monitoring

### Sentry Integration

```typescript
// src/lib/sentry.ts
import * as Sentry from '@sentry/nextjs';

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
  environment: process.env.NEXT_PUBLIC_ENV,
  tracesSampleRate: 0.1, // 10% of transactions
  replaysSessionSampleRate: 0.01, // 1% of sessions
  replaysOnErrorSampleRate: 1.0, // 100% of errors
  integrations: [
    new Sentry.BrowserTracing(),
    new Sentry.Replay(),
  ],
});
```

### LogRocket

```typescript
// src/lib/logrocket.ts
import LogRocket from 'logrocket';

if (process.env.NODE_ENV === 'production') {
  LogRocket.init('aeroxe/nexus-ai');

  // Identify user
  LogRocket.identify(userId, {
    name: userName,
    email: userEmail,
    tenantId: tenantId,
  });
}
```

### Web Vitals

```typescript
// src/lib/vitals.ts
import { onCLS, onFCP, onLCP, onTTFB, onINP } from 'web-vitals';

function sendToAnalytics(metric: any) {
  fetch('/api/v1/metrics/vitals', {
    method: 'POST',
    body: JSON.stringify(metric),
    keepalive: true,
  });
}

onCLS(sendToAnalytics);
onFCP(sendToAnalytics);
onLCP(sendToAnalytics);
onTTFB(sendToAnalytics);
onINP(sendToAnalytics);
```

---

## 11. Logging

### Structured Client-Side Logs

```typescript
// src/lib/logger.ts
type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LogEntry {
  level: LogLevel;
  message: string;
  context?: Record<string, unknown>;
  timestamp: string;
  url: string;
  userId?: string;
}

class ClientLogger {
  private buffer: LogEntry[] = [];
  private maxBuffer = 50;

  log(level: LogLevel, message: string, context?: Record<string, unknown>) {
    const entry: LogEntry = {
      level,
      message,
      context,
      timestamp: new Date().toISOString(),
      url: window.location.href,
    };

    if (process.env.NODE_ENV === 'development') {
      console[level](message, context);
    }

    this.buffer.push(entry);
    if (this.buffer.length >= this.maxBuffer) this.flush();
  }

  debug(msg: string, ctx?: Record<string, unknown>) { this.log('debug', msg, ctx); }
  info(msg: string, ctx?: Record<string, unknown>) { this.log('info', msg, ctx); }
  warn(msg: string, ctx?: Record<string, unknown>) { this.log('warn', msg, ctx); }
  error(msg: string, ctx?: Record<string, unknown>) { this.log('error', msg, ctx); }

  async flush() {
    const entries = [...this.buffer];
    this.buffer = [];
    try {
      await fetch('/api/v1/logs/client', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ logs: entries }),
        keepalive: true,
      });
    } catch {
      this.buffer.unshift(...entries);
    }
  }
}

export const logger = new ClientLogger();
```

---

## 12. Alerting

### Error Rate Alerts

```yaml
# prometheus/alerts.yml (frontend metrics)
groups:
  - name: frontend
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High frontend error rate"

      - alert: HighLCP
        expr: histogram_quantile(0.95, le, web_vitals_lcp_seconds) > 2.5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "LCP exceeds 2.5s"

      - alert: HighCLS
        expr: web_vitals_cls > 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "CLS exceeds 0.1"
```

---

## 13. Rollback Strategy

### Vercel Rollback

```bash
# Instant rollback to previous deployment
npx vercel rollback

# Rollback to specific deployment
npx vercel rollback <deployment-url>
```

### Kubernetes Rollback

```bash
# Rollback to previous version
kubectl rollout undo deployment/nexus-frontend -n aeroxe

# Rollback to specific revision
kubectl rollout undo deployment/nexus-frontend --to-revision=3 -n aeroxe

# Check rollout status
kubectl rollout status deployment/nexus-frontend -n aeroxe
```

### Blue-Green Deployment

```
┌────────────────┐    ┌────────────────┐
│  Blue (Current)│    │  Green (New)   │
│  v1.2.0        │    │  v1.3.0        │
│  100% traffic  │    │  0% traffic    │
└────────┬───────┘    └────────┬───────┘
         │                     │
    ┌────▼─────────────────────▼────┐
    │         Load Balancer          │
    └────────────────────────────────┘
              │
    Test green, then switch:
    Blue: 0%  → Green: 100%
```

### Canary Deployment

```yaml
# Gradual traffic shift
# Step 1: 5% traffic to new version
# Step 2: 25% traffic
# Step 3: 50% traffic
# Step 4: 100% traffic

# If error rate increases at any step, rollback
```

---

## 14. Feature Flags

```typescript
// src/lib/feature-flags.ts
import { create } from 'zustand';

interface FeatureFlags {
  newDashboard: boolean;
  voiceInputBeta: boolean;
  advancedAnalytics: boolean;
  experimentalChatUI: boolean;
}

// Environment-based feature flags
const ENV_FLAGS: Record<string, FeatureFlags> = {
  development: {
    newDashboard: true,
    voiceInputBeta: true,
    advancedAnalytics: true,
    experimentalChatUI: true,
  },
  staging: {
    newDashboard: true,
    voiceInputBeta: true,
    advancedAnalytics: false,
    experimentalChatUI: false,
  },
  production: {
    newDashboard: false,
    voiceInputBeta: false,
    advancedAnalytics: false,
    experimentalChatUI: false,
  },
};

export const useFeatureFlags = create<FeatureFlags>(() => {
  const env = process.env.NEXT_PUBLIC_ENV || 'development';
  return ENV_FLAGS[env] || ENV_FLAGS.development;
});
```

---

## 15. Post-Deployment Verification

### Smoke Tests

```yaml
# .github/workflows/smoke-test.yml
name: Post-Deploy Smoke Tests
on:
  deployment:
    types: [completed]

jobs:
  smoke:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run smoke tests
        run: |
          # Health check
          curl -f https://app.aeroxe.com/health || exit 1

          # Login page loads
          curl -f https://app.aeroxe.com/login || exit 1

          # API reachable
          curl -f https://api.aeroxe.com/health || exit 1

          # Static assets load
          curl -f https://app.aeroxe.com/_next/static/chunk-main.js || exit 1
```

### Health Check Endpoint

```typescript
// app/health/route.ts (Next.js App Router)
import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json({
    status: 'healthy',
    version: process.env.NEXT_PUBLIC_APP_VERSION || 'unknown',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
  });
}
```

---

## 16. Internationalization (i18n)

```typescript
// src/i18n/config.ts
export const locales = ['en', 'hi', 'bn', 'ta', 'te', 'mr'] as const;
export type Locale = (typeof locales)[number];
export const defaultLocale: Locale = 'en';

// src/i18n/messages/en.json
{
  "common": {
    "loading": "Loading...",
    "error": "Something went wrong",
    "retry": "Try Again"
  },
  "chat": {
    "placeholder": "Type your message...",
    "send": "Send",
    "typing": "Assistant is typing..."
  },
  "dashboard": {
    "welcome": "Welcome back, {name}!",
    "stats": "Dashboard Statistics"
  }
}
```

### RTL Support

```css
[dir="rtl"] .sidebar {
  direction: rtl;
  border-left: none;
  border-right: 1px solid var(--border);
}

[dir="rtl"] .chat-input {
  direction: rtl;
  text-align: right;
}
```

---

## 17. Incident Response

### Error Escalation Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Frontend │────>│ Sentry   │────>│ PagerDuty│────>│  On-Call │
│ Error    │     │ Alert    │     │ Alert    │     │ Engineer │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
```

### Escalation Rules

| Severity | Criteria | Response Time | Escalation |
|----------|----------|---------------|------------|
| P0 | Site down, data loss | 15 minutes | Immediate page |
| P1 | Major feature broken | 1 hour | Slack + email |
| P2 | Minor feature issue | 4 hours | Slack |
| P3 | Cosmetic issue | Next sprint | Backlog |

### Communication Template

```markdown
## Incident Report — [Date]

**Status**: Investigating / Identified / Monitoring / Resolved
**Impact**: [Description of user impact]
**Duration**: [Start time] — [End time]
**Root Cause**: [Brief description]
**Resolution**: [What was done]
**Prevention**: [Steps to prevent recurrence]
```

---

## 18. Deployment Checklist

| Step | Description | Automated |
|------|-------------|-----------|
| Lint | ESLint + Prettier | Yes |
| TypeCheck | TypeScript strict | Yes |
| Unit Tests | Vitest with coverage | Yes |
| E2E Tests | Playwright | Yes |
| Visual Tests | Snapshots/Chromatic | Yes |
| Build | Production build | Yes |
| Security Scan | SAST + npm audit | Yes |
| Deploy Staging | Auto on main merge | Yes |
| Smoke Tests | Health checks | Yes |
| Manual QA | Staging verification | No |
| Deploy Production | Manual approval | Yes |
| Post-Deploy Verification | Health + smoke | Yes |
| Monitor | Sentry + logs | Yes |
| Rollback Ready | If errors spike | Manual |
