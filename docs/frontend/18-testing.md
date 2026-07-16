# 18 — Frontend Testing Strategy & Quality Engineering

## Jest + React Testing Library + Playwright + Accessibility + Performance + Visual Regression

---

## 1. Testing Philosophy

AeroXe Nexus AI frontend follows **Test-Driven Development (TDD)** with AI-assisted quality gates.

Core rules:

- **No component ships without tests.**
- **Every API integration has a mock-based test.**
- **Accessibility is tested, not assumed.**
- **Performance budgets are enforced in CI.**

### Development Cycle

```
Component Design → Write Test → Implement → Refactor → Integration Test → Deploy
```

---

## 2. Testing Pyramid

```
                    /\
                   /  \
                  / E2E\              ← Few, expensive, high confidence
                 /------\
                /Visual  \
               /Regression\           ← Visual diff per component
              /------------\
             / Accessibility \
            /----------------\
           / Integration      \       ← API, hooks, stores
          /--------------------\
         /   Unit Tests         \     ← Fast, cheap, many
        /________________________\
```

### Layer Distribution

| Layer              | Count Target | Speed      | Confidence   | Cost   |
|--------------------|-------------|------------|--------------|--------|
| Unit Tests         | 200+        | < 1s total | Low–Medium   | Low    |
| Component Tests    | 100+        | < 5s total | Medium       | Low    |
| Integration Tests  | 50+         | < 15s total| Medium–High  | Medium |
| E2E Tests          | 20+         | < 60s total| High         | High   |
| Visual Regression  | Per component| < 30s    | High         | Medium |
| Accessibility      | Per page    | < 10s      | High         | Low    |
| Performance        | Per build   | < 120s     | High         | Medium |

---

## 3. Test Environment Setup

### jsdom vs happy-dom

| Feature           | jsdom      | happy-dom    |
|-------------------|------------|--------------|
| Speed             | Slower     | 2–5x faster  |
| DOM Accuracy      | High       | High         |
| Web APIs          | Broad      | Growing      |
| Community         | Large      | Growing      |
| Recommended For   | Complex DOM| Fast unit    |

### Vitest Configuration (Recommended)

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@components': path.resolve(__dirname, './src/components'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@stores': path.resolve(__dirname, './src/stores'),
      '@lib': path.resolve(__dirname, './src/lib'),
      '@types': path.resolve(__dirname, './src/types'),
    },
  },
  test: {
    environment: 'happy-dom',
    globals: true,
    setupFiles: ['./tests/setup.ts'],
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'tests/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/index.ts',
        '**/*.stories.tsx',
      ],
      thresholds: {
        lines: 80,
        branches: 90,
        functions: 80,
        statements: 80,
      },
    },
  },
});
```

### Setup File

```typescript
// tests/setup.ts
import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach, beforeAll, afterAll } from 'vitest';
import { server } from './mocks/server';

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => {
  cleanup();
  server.resetHandlers();
});
afterAll(() => server.close());
```

---

## 4. Unit Testing

### Utility Functions

```typescript
// src/lib/format.ts
export function formatCurrency(amount: number, currency = 'INR'): string {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amount);
}

export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength - 3) + '...';
}
```

```typescript
// src/lib/__tests__/format.test.ts
import { describe, it, expect } from 'vitest';
import { formatCurrency, truncateText } from '../format';

describe('formatCurrency', () => {
  it('formats INR currency', () => {
    expect(formatCurrency(150000)).toBe('₹1,50,000');
  });

  it('formats USD currency', () => {
    expect(formatCurrency(1500, 'USD')).toBe('$1,500');
  });

  it('handles zero', () => {
    expect(formatCurrency(0)).toBe('₹0');
  });

  it('handles decimals', () => {
    expect(formatCurrency(1234.56)).toBe('₹1,235');
  });
});

describe('truncateText', () => {
  it('returns full text when shorter than max', () => {
    expect(truncateText('hello', 10)).toBe('hello');
  });

  it('truncates and adds ellipsis', () => {
    expect(truncateText('hello world test', 10)).toBe('hello w...');
  });

  it('returns exact length when equal', () => {
    expect(truncateText('12345', 5)).toBe('12345');
  });
});
```

### Validation Schemas

```typescript
// src/lib/__tests__/validation.test.ts
import { describe, it, expect } from 'vitest';
import { loginSchema, chatMessageSchema } from '../validation';

describe('loginSchema', () => {
  it('accepts valid login', () => {
    const result = loginSchema.safeParse({
      email: 'user@aeroxe.com',
      password: 'SecurePass123!',
    });
    expect(result.success).toBe(true);
  });

  it('rejects invalid email', () => {
    const result = loginSchema.safeParse({
      email: 'not-an-email',
      password: 'SecurePass123!',
    });
    expect(result.success).toBe(false);
    expect(result.error?.errors[0].path).toContain('email');
  });

  it('rejects short password', () => {
    const result = loginSchema.safeParse({
      email: 'user@aeroxe.com',
      password: '123',
    });
    expect(result.success).toBe(false);
  });
});

describe('chatMessageSchema', () => {
  it('accepts valid message', () => {
    const result = chatMessageSchema.safeParse({
      content: 'Hello, what is my inventory status?',
      agent_id: 'agent-uuid-1',
    });
    expect(result.success).toBe(true);
  });

  it('rejects empty content', () => {
    const result = chatMessageSchema.safeParse({
      content: '',
      agent_id: 'agent-uuid-1',
    });
    expect(result.success).toBe(false);
  });
});
```

---

## 5. Component Testing

### Testing shadcn/ui Components

```typescript
// src/components/ui/__tests__/button.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from '../button';

describe('Button', () => {
  it('renders with default variant', () => {
    render(<Button>Click me</Button>);
    const button = screen.getByRole('button', { name: /click me/i });
    expect(button).toBeInTheDocument();
  });

  it('renders with destructive variant', () => {
    render(<Button variant="destructive">Delete</Button>);
    expect(screen.getByRole('button')).toHaveClass('bg-destructive');
  });

  it('calls onClick when clicked', async () => {
    const user = userEvent.setup();
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click</Button>);

    await user.click(screen.getByRole('button'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('is disabled when disabled prop is true', async () => {
    const user = userEvent.setup();
    const handleClick = vi.fn();
    render(<Button disabled onClick={handleClick}>Submit</Button>);

    await user.click(screen.getByRole('button'));
    expect(handleClick).not.toHaveBeenCalled();
  });

  it('shows loading spinner when loading', () => {
    render(<Button loading>Submit</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
  });
});
```

### Testing Form Components

```typescript
// src/components/__tests__/chat-input.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ChatInput } from '../chat-input';

describe('ChatInput', () => {
  it('renders input field', () => {
    render(<ChatInput onSend={vi.fn()} />);
    expect(screen.getByPlaceholderText(/type your message/i)).toBeInTheDocument();
  });

  it('calls onSend with message text on Enter', async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();
    render(<ChatInput onSend={onSend} />);

    const input = screen.getByPlaceholderText(/type your message/i);
    await user.type(input, 'Show me my inventory');
    await user.keyboard('{Enter}');

    expect(onSend).toHaveBeenCalledWith('Show me my inventory');
  });

  it('does not send empty message', async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();
    render(<ChatInput onSend={onSend} />);

    await user.keyboard('{Enter}');
    expect(onSend).not.toHaveBeenCalled();
  });

  it('clears input after sending', async () => {
    const user = userEvent.setup();
    render(<ChatInput onSend={vi.fn()} />);

    const input = screen.getByPlaceholderText(/type your message/i);
    await user.type(input, 'Hello');
    await user.keyboard('{Enter}');

    expect(input).toHaveValue('');
  });

  it('disables input when sending', () => {
    render(<ChatInput onSend={vi.fn()} isSending />);
    expect(screen.getByPlaceholderText(/type your message/i)).toBeDisabled();
  });
});
```

### Testing Data Display Components

```typescript
// src/components/__tests__/agent-card.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AgentCard } from '../agent-card';
import type { Agent } from '@/types/agent';

const mockAgent: Agent = {
  id: 'agent-1',
  name: 'Broadband Assistant',
  description: 'Handles broadband customer queries',
  status: 'active',
  avatar: '/avatars/broadband.png',
  capabilities: ['chat', 'voice', 'document-analysis'],
  tenant_id: 'tenant-1',
  created_at: '2024-01-15T10:00:00Z',
};

describe('AgentCard', () => {
  it('displays agent name and description', () => {
    render(<AgentCard agent={mockAgent} onClick={vi.fn()} />);
    expect(screen.getByText('Broadband Assistant')).toBeInTheDocument();
    expect(screen.getByText('Handles broadband customer queries')).toBeInTheDocument();
  });

  it('displays agent status badge', () => {
    render(<AgentCard agent={mockAgent} onClick={vi.fn()} />);
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('displays capability tags', () => {
    render(<AgentCard agent={mockAgent} onClick={vi.fn()} />);
    expect(screen.getByText('chat')).toBeInTheDocument();
    expect(screen.getByText('voice')).toBeInTheDocument();
  });

  it('calls onClick when card is clicked', async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    render(<AgentCard agent={mockAgent} onClick={onClick} />);

    await user.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalledWith(mockAgent);
  });
});
```

---

## 6. Integration Testing

### API Integration Tests

```typescript
// src/hooks/__tests__/use-chat.test.ts
import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useChat } from '../use-chat';
import { server } from '@/tests/mocks/server';
import { http, HttpResponse } from 'msw';

describe('useChat', () => {
  it('sends message and receives response', async () => {
    const { result } = renderHook(() => useChat({ agentId: 'agent-1' }));

    act(() => {
      result.current.sendMessage('Hello');
    });

    expect(result.current.isLoading).toBe(true);

    await act(async () => {
      await vi.waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });
    });

    expect(result.current.messages).toHaveLength(2);
    expect(result.current.messages[0].role).toBe('user');
    expect(result.current.messages[1].role).toBe('assistant');
  });

  it('handles API error gracefully', async () => {
    server.use(
      http.post('/api/v1/chat/send', () => {
        return HttpResponse.json(
          { error: 'Agent not found' },
          { status: 404 }
        );
      })
    );

    const { result } = renderHook(() => useChat({ agentId: 'nonexistent' }));

    act(() => {
      result.current.sendMessage('Hello');
    });

    await act(async () => {
      await vi.waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });
    });

    expect(result.current.error).toBe('Agent not found');
  });

  it('cancels in-flight request', async () => {
    const { result } = renderHook(() => useChat({ agentId: 'agent-1' }));

    act(() => {
      result.current.sendMessage('Long query');
    });

    act(() => {
      result.current.cancel();
    });

    expect(result.current.isCancelled).toBe(true);
  });
});
```

### Zustand Store Tests

```typescript
// src/stores/__tests__/chat-store.test.ts
import { describe, it, expect, beforeEach } from 'vitest';
import { act } from '@testing-library/react';
import { useChatStore } from '../chat-store';

describe('chatStore', () => {
  beforeEach(() => {
    useChatStore.getState().reset();
  });

  it('adds message to conversation', () => {
    const { addMessage } = useChatStore.getState();

    act(() => {
      addMessage({
        id: 'msg-1',
        role: 'user',
        content: 'Hello',
        timestamp: new Date().toISOString(),
      });
    });

    const { messages } = useChatStore.getState();
    expect(messages).toHaveLength(1);
    expect(messages[0].content).toBe('Hello');
  });

  it('switches active conversation', () => {
    const { createConversation, setActiveConversation } =
      useChatStore.getState();

    let conv1: string;
    let conv2: string;

    act(() => {
      conv1 = createConversation('Chat 1');
      conv2 = createConversation('Chat 2');
    });

    act(() => {
      setActiveConversation(conv1!);
    });

    expect(useChatStore.getState().activeConversationId).toBe(conv1!);
  });

  it('deletes conversation and its messages', () => {
    const { createConversation, deleteConversation, addMessage } =
      useChatStore.getState();

    let convId: string;

    act(() => {
      convId = createConversation('To Delete');
      addMessage({
        id: 'msg-1',
        role: 'user',
        content: 'Test',
        timestamp: new Date().toISOString(),
        conversationId: convId,
      });
    });

    act(() => {
      deleteConversation(convId!);
    });

    const { conversations, messages } = useChatStore.getState();
    expect(conversations.find((c) => c.id === convId!)).toBeUndefined();
    expect(messages.filter((m) => m.conversationId === convId!)).toHaveLength(0);
  });
});
```

---

## 7. E2E Testing (Playwright)

### Configuration

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html', { outputFolder: 'test-results/playwright' }],
    ['junit', { outputFile: 'test-results/junit.xml' }],
  ],
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    { name: 'webkit', use: { ...devices['Desktop Safari'] } },
    { name: 'mobile-chrome', use: { ...devices['Pixel 7'] } },
  ],
  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: !process.env.CI,
  },
});
```

### Login Flow E2E

```typescript
// e2e/auth/login.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Login Flow', () => {
  test('successful login redirects to dashboard', async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel('Email').fill('admin@aeroxe.com');
    await page.getByLabel('Password').fill('SecurePass123!');
    await page.getByRole('button', { name: 'Sign In' }).click();

    await expect(page).toHaveURL('/dashboard');
    await expect(page.getByText('Welcome')).toBeVisible();
  });

  test('invalid credentials shows error', async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel('Email').fill('wrong@email.com');
    await page.getByLabel('Password').fill('wrong');
    await page.getByRole('button', { name: 'Sign In' }).click();

    await expect(page.getByText('Invalid credentials')).toBeVisible();
  });

  test('form validation prevents empty submission', async ({ page }) => {
    await page.goto('/login');
    await page.getByRole('button', { name: 'Sign In' }).click();

    await expect(page.getByText('Email is required')).toBeVisible();
    await expect(page.getByText('Password is required')).toBeVisible();
  });
});
```

### Chat Flow E2E

```typescript
// e2e/chat/chat-flow.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Chat Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel('Email').fill('admin@aeroxe.com');
    await page.getByLabel('Password').fill('SecurePass123!');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await expect(page).toHaveURL('/dashboard');
  });

  test('send message and receive response', async ({ page }) => {
    await page.getByText('AI Assistant').click();

    const chatInput = page.getByPlaceholder('Type your message...');
    await chatInput.fill('Show me current inventory levels');
    await chatInput.press('Enter');

    await expect(
      page.locator('[data-role="assistant"]').first()
    ).toBeVisible({ timeout: 15000 });
  });

  test('agent selection changes context', async ({ page }) => {
    await page.getByText('AI Assistant').click();

    await page.getByRole('combobox', { name: 'Select Agent' }).click();
    await page.getByText('Broadband Assistant').click();

    await expect(page.getByText('Broadband Assistant')).toBeVisible();
  });

  test('file upload works in chat', async ({ page }) => {
    await page.getByText('AI Assistant').click();

    const fileInput = page.getByLabel('Upload file');
    await fileInput.setInputFiles({
      name: 'report.pdf',
      mimeType: 'application/pdf',
      buffer: Buffer.from('test content'),
    });

    await expect(page.getByText('report.pdf')).toBeVisible();
  });
});
```

---

## 8. Visual Regression Testing

### Snapshot Testing

```typescript
// src/components/__tests__/button.visual.test.tsx
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { Button } from '../button';

describe('Button Visual', () => {
  const variants = ['default', 'destructive', 'outline', 'ghost', 'link'] as const;
  const sizes = ['default', 'sm', 'lg', 'icon'] as const;

  variants.forEach((variant) => {
    sizes.forEach((size) => {
      it(`${variant}-${size} renders correctly`, () => {
        const { container } = render(
          <Button variant={variant} size={size}>Click me</Button>
        );
        expect(container.firstChild).toMatchSnapshot();
      });
    });
  });
});
```

### Chromatic Integration

```typescript
// .storybook/test-runner.ts
import { TestRunnerConfig } from '@storybook/test-runner';

const config: TestRunnerConfig = {
  async preVisit({ page }) {
    await page.setViewportSize({ width: 1280, height: 720 });
  },
};

export default config;
```

### Storybook Stories for Visual Tests

```typescript
// src/components/button.stories.tsx
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './button';

const meta: Meta<typeof Button> = {
  title: 'UI/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'destructive', 'outline', 'ghost', 'link'],
    },
    size: {
      control: 'select',
      options: ['default', 'sm', 'lg', 'icon'],
    },
  },
};

export default meta;
type Story = StoryObj<typeof Button>;

export const AllVariants: Story = {
  render: () => (
    <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
      {(['default', 'destructive', 'outline', 'ghost', 'link'] as const).map(
        (variant) => (
          <Button key={variant} variant={variant}>{variant}</Button>
        )
      )}
    </div>
  ),
};
```

---

## 9. Accessibility Testing

### axe-core Integration

```typescript
// src/components/__tests__/dialog.a11y.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
import { Dialog, DialogContent, DialogTrigger } from '../ui/dialog';
import { Button } from '../ui/button';

expect.extend(toHaveNoViolations);

describe('Dialog Accessibility', () => {
  it('has no axe violations', async () => {
    const { container } = render(
      <Dialog>
        <DialogTrigger asChild>
          <Button>Open</Button>
        </DialogTrigger>
        <DialogContent>
          <h2>Dialog Title</h2>
          <p>Dialog content goes here</p>
        </DialogContent>
      </Dialog>
    );

    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });
});
```

### Page-Level A11y Tests

```typescript
// src/pages/__tests__/dashboard.a11y.test.tsx
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
import { MemoryRouter } from 'react-router-dom';
import { DashboardPage } from '../dashboard';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

expect.extend(toHaveNoViolations);

const queryClient = new QueryClient();

describe('Dashboard Accessibility', () => {
  it('dashboard page has no violations', async () => {
    const { container } = render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });
});
```

### Lighthouse CI

```yaml
# .github/workflows/lighthouse.yml
name: Lighthouse CI
on:
  pull_request:
    branches: [main]

jobs:
  lighthouse:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npm ci
      - run: npm run build
      - uses: treosh/lighthouse-ci-action@v10
        with:
          urls: |
            http://localhost:3000/dashboard
            http://localhost:3000/chat
          budgetPath: ./lighthouse-budget.json
          uploadArtifacts: true
```

### Lighthouse Budget

```json
[
  {
    "path": "/*",
    "timings": {
      "first-contentful-paint": { "budget": 1500 },
      "largest-contentful-paint": { "budget": 2500 },
      "cumulative-layout-shift": { "budget": 0.1 },
      "total-blocking-time": { "budget": 300 },
      "interactive": { "budget": 3500 }
    },
    "resourceSizes": {
      "script": { "budget": 300 },
      "stylesheet": { "budget": 50 },
      "image": { "budget": 200 },
      "font": { "budget": 100 },
      "total": { "budget": 700 }
    }
  }
]
```

---

## 10. Performance Testing

### Web Vitals Monitoring

```typescript
// src/lib/web-vitals.ts
import type { Metric } from 'web-vitals';

function sendToAnalytics(metric: Metric) {
  const body = JSON.stringify({
    name: metric.name,
    value: metric.value,
    id: metric.id,
    delta: metric.delta,
    rating: metric.rating,
  });

  if (navigator.sendBeacon) {
    navigator.sendBeacon('/api/v1/metrics/web-vitals', body);
  }
}

export function reportWebVitals() {
  if (typeof window !== 'undefined') {
    import('web-vitals').then(({ onCLS, onFCP, onLCP, onTTFB, onINP }) => {
      onCLS(sendToAnalytics);
      onFCP(sendToAnalytics);
      onLCP(sendToAnalytics);
      onTTFB(sendToAnalytics);
      onINP(sendToAnalytics);
    });
  }
}
```

### Performance Budget Test

```typescript
// e2e/performance/budget.spec.ts
import { test, expect } from '@playwright/test';

test('homepage meets performance budget', async ({ page }) => {
  const startTime = Date.now();
  await page.goto('/');
  await page.waitForLoadState('networkidle');
  const loadTime = Date.now() - startTime;

  expect(loadTime).toBeLessThan(3000);

  const performanceMetrics = await page.evaluate(() => {
    const paint = performance.getEntriesByType('paint');
    const lcp = performance.getEntriesByType('largest-contentful-paint');
    return {
      fcp: paint.find((e) => e.name === 'first-contentful-paint')?.startTime,
      lcp: lcp.length > 0 ? lcp[0].startTime : null,
    };
  });

  if (performanceMetrics.fcp) {
    expect(performanceMetrics.fcp).toBeLessThan(1500);
  }
  if (performanceMetrics.lcp) {
    expect(performanceMetrics.lcp).toBeLessThan(2500);
  }
});
```

---

## 11. Mock Strategies

### MSW Handlers

```typescript
// tests/mocks/handlers.ts
import { http, HttpResponse } from 'msw';

export const handlers = [
  http.post('/api/v1/auth/login', async ({ request }) => {
    const body = await request.json() as { email: string; password: string };

    if (body.email === 'admin@aeroxe.com' && body.password === 'SecurePass123!') {
      return HttpResponse.json({
        access_token: 'mock-jwt-token',
        refresh_token: 'mock-refresh-token',
        user: {
          id: 'user-1',
          email: body.email,
          name: 'Admin User',
          roles: ['admin'],
        },
      });
    }

    return HttpResponse.json(
      { error: 'Invalid credentials' },
      { status: 401 }
    );
  }),

  http.get('/api/v1/agents', () => {
    return HttpResponse.json({
      agents: [
        {
          id: 'agent-1',
          name: 'Broadband Assistant',
          description: 'Handles broadband queries',
          status: 'active',
        },
        {
          id: 'agent-2',
          name: 'ERP Assistant',
          description: 'Manages ERP operations',
          status: 'active',
        },
      ],
    });
  }),

  http.post('/api/v1/chat/send', async ({ request }) => {
    const body = await request.json() as { content: string; agent_id: string };
    return HttpResponse.json({
      id: 'msg-response-1',
      role: 'assistant',
      content: `I received your message: "${body.content}". Here is my analysis...`,
      agent_id: body.agent_id,
      timestamp: new Date().toISOString(),
    });
  }),
];
```

### MSW Server

```typescript
// tests/mocks/server.ts
import { setupServer } from 'msw/node';
import { handlers } from './handlers';

export const server = setupServer(...handlers);
```

### WebSocket Mock

```typescript
// tests/mocks/websocket.ts
export class MockWebSocket {
  static instances: MockWebSocket[] = [];
  readyState = 1; // OPEN
  url: string;
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  private closeRequested = false;

  constructor(url: string) {
    this.url = url;
    MockWebSocket.instances.push(this);
    setTimeout(() => this.onopen?.(), 0);
  }

  send(data: string) {
    const message = JSON.parse(data);
    if (message.type === 'chat') {
      this.simulateMessage({
        type: 'token',
        token: 'Hello',
      });
      setTimeout(() => {
        this.simulateMessage({
          type: 'token',
          token: ' world',
        });
      }, 50);
      setTimeout(() => {
        this.simulateMessage({
          type: 'completed',
          message_id: 'msg-1',
        });
      }, 100);
    }
  }

  simulateMessage(data: object) {
    this.onmessage?.({ data: JSON.stringify(data) });
  }

  close() {
    this.closeRequested = true;
    this.readyState = 3; // CLOSED
    this.onclose?.();
  }

  static reset() {
    MockWebSocket.instances = [];
  }
}
```

### Mock Helpers

```typescript
// tests/mocks/helpers.ts
import { vi } from 'vitest';

export function mockFetch(data: unknown, status = 200) {
  global.fetch = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
  });
}

export function mockLocalStorage() {
  const store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { Object.keys(store).forEach((k) => delete store[k]); }),
    get length() { return Object.keys(store).length; },
    key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
  };
}
```

---

## 12. Test Fixtures

### User Fixtures

```typescript
// tests/fixtures/users.ts
import type { User } from '@/types/user';

export const adminUser: User = {
  id: 'user-admin-1',
  email: 'admin@aeroxe.com',
  name: 'Admin User',
  roles: ['admin'],
  permissions: ['ai.execute', 'document.read', 'user.manage', 'settings.write'],
  tenant_id: 'tenant-1',
  avatar: null,
  created_at: '2024-01-01T00:00:00Z',
  last_login: '2024-06-15T10:00:00Z',
};

export const regularUser: User = {
  id: 'user-regular-1',
  email: 'user@aeroxe.com',
  name: 'Regular User',
  roles: ['user'],
  permissions: ['ai.execute', 'document.read'],
  tenant_id: 'tenant-1',
  avatar: null,
  created_at: '2024-03-01T00:00:00Z',
  last_login: '2024-06-14T08:00:00Z',
};

export const users = { adminUser, regularUser };
```

### Chat Message Fixtures

```typescript
// tests/fixtures/messages.ts
import type { Message } from '@/types/chat';

export const chatMessages: Message[] = [
  {
    id: 'msg-1',
    role: 'user',
    content: 'Show me my current inventory levels',
    timestamp: '2024-06-15T10:00:00Z',
    conversationId: 'conv-1',
  },
  {
    id: 'msg-2',
    role: 'assistant',
    content: 'Here are your current inventory levels across all warehouses. Total SKUs: 1,247. Low stock items: 23. Reorder needed: 15.',
    timestamp: '2024-06-15T10:00:05Z',
    conversationId: 'conv-1',
    agent_id: 'agent-erp',
  },
  {
    id: 'msg-3',
    role: 'user',
    content: 'Which items need immediate reorder?',
    timestamp: '2024-06-15T10:01:00Z',
    conversationId: 'conv-1',
  },
  {
    id: 'msg-4',
    role: 'assistant',
    content: 'The following 15 items need immediate reorder based on current consumption rates and lead times...',
    timestamp: '2024-06-15T10:01:10Z',
    conversationId: 'conv-1',
    agent_id: 'agent-erp',
  },
];
```

### Agent Config Fixtures

```typescript
// tests/fixtures/agents.ts
import type { AgentConfig } from '@/types/agent';

export const broadbandAgent: AgentConfig = {
  id: 'agent-broadband-1',
  name: 'Broadband Assistant',
  description: 'Handles customer queries for broadband services',
  status: 'active',
  system_prompt: 'You are a helpful broadband support agent...',
  capabilities: ['chat', 'voice', 'document-analysis', 'ticket-creation'],
  knowledge_base_ids: ['kb-broadband-1'],
  tools: ['check_network_status', 'create_ticket', 'lookup_customer'],
  tenant_id: 'tenant-1',
  created_at: '2024-01-15T10:00:00Z',
  updated_at: '2024-06-01T12:00:00Z',
  max_tokens: 4096,
  temperature: 0.7,
};

export const erpAgent: AgentConfig = {
  id: 'agent-erp-1',
  name: 'ERP Intelligence Agent',
  description: 'Manages ERP operations, inventory, and finance queries',
  status: 'active',
  system_prompt: 'You are an ERP intelligence assistant...',
  capabilities: ['chat', 'sql-query', 'report-generation'],
  knowledge_base_ids: ['kb-erp-1', 'kb-finance-1'],
  tools: ['query_inventory', 'generate_report', 'check_orders'],
  tenant_id: 'tenant-1',
  created_at: '2024-02-01T10:00:00Z',
  updated_at: '2024-06-10T12:00:00Z',
  max_tokens: 8192,
  temperature: 0.3,
};
```

### Factory Pattern for Test Data

```typescript
// tests/factories/index.ts
import type { User, Message, Agent } from '@/types';

let counter = 0;
const uid = (prefix: string) => `${prefix}-${++counter}`;

export function buildUser(overrides: Partial<User> = {}): User {
  return {
    id: uid('user'),
    email: `test-${counter}@aeroxe.com`,
    name: `Test User ${counter}`,
    roles: ['user'],
    permissions: ['ai.execute'],
    tenant_id: 'tenant-test',
    avatar: null,
    created_at: new Date().toISOString(),
    last_login: null,
    ...overrides,
  };
}

export function buildMessage(overrides: Partial<Message> = {}): Message {
  return {
    id: uid('msg'),
    role: 'user',
    content: `Test message ${counter}`,
    timestamp: new Date().toISOString(),
    conversationId: 'conv-test',
    ...overrides,
  };
}

export function buildAgent(overrides: Partial<Agent> = {}): Agent {
  return {
    id: uid('agent'),
    name: `Agent ${counter}`,
    description: `Test agent ${counter}`,
    status: 'active',
    capabilities: ['chat'],
    tenant_id: 'tenant-test',
    created_at: new Date().toISOString(),
    ...overrides,
  };
}
```

---

## 13. Test Patterns (AAA)

### Arrange-Act-Assert Pattern

```typescript
describe('useAuth', () => {
  it('logs in user and stores token', async () => {
    // Arrange
    const { result } = renderHook(() => useAuth());
    const credentials = {
      email: 'admin@aeroxe.com',
      password: 'SecurePass123!',
    };

    // Act
    await act(async () => {
      await result.current.login(credentials);
    });

    // Assert
    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.user?.email).toBe('admin@aeroxe.com');
    expect(result.current.token).toBeTruthy();
  });

  it('logs out user and clears token', async () => {
    // Arrange
    const { result } = renderHook(() => useAuth());
    await act(async () => {
      await result.current.login({
        email: 'admin@aeroxe.com',
        password: 'SecurePass123!',
      });
    });

    // Act
    await act(async () => {
      await result.current.logout();
    });

    // Assert
    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.user).toBeNull();
    expect(result.current.token).toBeNull();
  });
});
```

---

## 14. Test Coverage Targets

| Area               | Lines | Branches | Functions | Statements |
|--------------------|-------|----------|-----------|------------|
| Components         | 80%   | 90%      | 80%       | 80%        |
| Hooks              | 90%   | 95%      | 90%       | 90%        |
| Stores             | 90%   | 95%      | 90%       | 90%        |
| Lib/Utils          | 95%   | 98%      | 95%       | 95%        |
| Services           | 85%   | 90%      | 85%       | 85%        |
| Types              | 100%  | 100%     | 100%      | 100%       |
| **Overall Target** | **80%**| **90%** | **80%**   | **80%**    |

### Coverage Enforcement

```yaml
# .github/workflows/test.yml
- name: Check coverage
  run: |
    npx vitest run --coverage
    npx istanbul check-coverage \
      --lines 80 \
      --branches 90 \
      --functions 80 \
      --statements 80
```

---

## 15. CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Frontend Tests
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci
      - run: npm run lint
      - run: npm run typecheck
      - run: npm run test:unit -- --coverage
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: coverage-report
          path: coverage/

  e2e-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npm ci
      - run: npx playwright install --with-deps
      - run: npm run build
      - run: npm run test:e2e
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: test-results/playwright/

  visual-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npm ci
      - run: npm run test:visual
```

### NPM Scripts

```json
{
  "scripts": {
    "test": "vitest",
    "test:unit": "vitest run --coverage",
    "test:watch": "vitest --watch",
    "test:ui": "vitest --ui",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "test:visual": "vitest run --reporter=verbose --testPathPattern='visual'",
    "test:a11y": "vitest run --testPathPattern='a11y'",
    "test:performance": "playwright test e2e/performance/",
    "test:all": "npm run test:unit && npm run test:e2e && npm run test:visual"
  }
}
```

---

## 16. Snapshot Testing

### Component Snapshots

```typescript
// src/components/__tests__/sidebar.snapshot.test.tsx
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { Sidebar } from '../sidebar';

describe('Sidebar Snapshot', () => {
  it('renders default sidebar', () => {
    const { container } = render(<Sidebar />);
    expect(container.firstChild).toMatchSnapshot();
  });

  it('renders collapsed sidebar', () => {
    const { container } = render(<Sidebar collapsed />);
    expect(container.firstChild).toMatchSnapshot();
  });
});
```

### Snapshot Update Workflow

```bash
# Update snapshots when UI changes are intentional
npx vitest run --update

# Review snapshot changes
git diff src/components/__tests__/__snapshots__/
```

---

## 17. Contract Testing

### API Contract Tests

```typescript
// tests/contracts/chat-api.contract.ts
import { describe, it, expect } from 'vitest';
import { z } from 'zod';

const ChatMessageResponseSchema = z.object({
  id: z.string().uuid(),
  role: z.enum(['user', 'assistant', 'system']),
  content: z.string().min(1),
  timestamp: z.string().datetime(),
  agent_id: z.string().uuid().optional(),
  conversationId: z.string().uuid().optional(),
});

describe('Chat API Contract', () => {
  it('POST /api/v1/chat/send response matches schema', async () => {
    const response = await fetch('/api/v1/chat/send', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: 'Hello', agent_id: 'agent-1' }),
    });
    const data = await response.json();

    const result = ChatMessageResponseSchema.safeParse(data);
    expect(result.success).toBe(true);
  });
});
```

---

## 18. Test Debugging

### Debug Tools

```bash
# Vitest watch mode
npx vitest --watch

# Vitest with UI
npx vitest --ui

# Debug specific test
npx vitest run src/hooks/__tests__/use-chat.test.ts --reporter=verbose

# Playwright debug mode
npx playwright test --debug

# Playwright trace viewer
npx playwright show-trace test-results/chromium/*/trace.zip
```

### Console Spy for Debugging

```typescript
it('logs error on API failure', async () => {
  const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

  // ... test code that triggers error ...

  expect(consoleSpy).toHaveBeenCalledWith(
    expect.stringContaining('API Error')
  );

  consoleSpy.mockRestore();
});
```

---

## 19. Test Reporting

### Coverage Report Configuration

```typescript
// vitest.config.ts (coverage section)
coverage: {
  reporter: ['text', 'json', 'html', 'lcov'],
  reportsDirectory: './coverage',
  exclude: [
    'node_modules/',
    'tests/',
    '**/*.d.ts',
    '**/*.config.*',
  ],
}
```

### Test Results Format

| Reporter | Format | Use Case |
|----------|--------|----------|
| text     | CLI    | Local dev |
| json     | JSON   | CI parsing |
| html     | HTML   | Browser review |
| lcov     | LCOV   | CI coverage tools |
| junit    | XML    | CI integration |

---

## 20. Test Data Management

### Cleanup Strategy

```typescript
// tests/setup.ts
import { afterEach } from 'vitest';
import { useChatStore } from '@/stores/chat-store';
import { useAuthStore } from '@/stores/auth-store';

afterEach(() => {
  // Reset all stores
  useChatStore.getState().reset();
  useAuthStore.getState().reset();

  // Clear localStorage
  localStorage.clear();

  // Clear sessionStorage
  sessionStorage.clear();
});
```

### Test Database Seeds

```typescript
// tests/seeds/index.ts
export async function seedTestData() {
  // Seed via API
  await fetch('/api/v1/test/seed', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      users: [adminUser, regularUser],
      agents: [broadbandAgent, erpAgent],
      conversations: testConversations,
    }),
  });
}

export async function cleanupTestData() {
  await fetch('/api/v1/test/cleanup', { method: 'DELETE' });
}
```

---

## 21. Test Hooks

### Custom Test Hook

```typescript
// tests/hooks/use-mock-auth.ts
import { renderHook, act } from '@testing-library/react';
import { useAuthStore } from '@/stores/auth-store';
import { adminUser } from '../fixtures/users';

export function renderAuthenticatedHook<T>(
  hook: () => T,
  user = adminUser
) {
  act(() => {
    useAuthStore.getState().setUser(user);
    useAuthStore.getState().setToken('mock-jwt-token');
  });

  return renderHook(hook);
}
```

### Usage

```typescript
import { renderAuthenticatedHook } from '../hooks/use-mock-auth';
import { useChat } from '@/hooks/use-chat';

it('authenticated user can send messages', async () => {
  const { result } = renderAuthenticatedHook(() =>
    useChat({ agentId: 'agent-1' })
  );

  act(() => {
    result.current.sendMessage('Hello');
  });

  expect(result.current.isLoading).toBe(true);
});
```

---

## 22. Summary Checklist

| Category              | Tool               | Status |
|-----------------------|--------------------|--------|
| Unit Testing          | Vitest             | Required |
| Component Testing     | React Testing Library | Required |
| E2E Testing           | Playwright         | Required |
| Visual Regression     | Chromatic/Snapshots | Required |
| Accessibility         | axe-core           | Required |
| Performance           | Lighthouse CI      | Required |
| Mocking               | MSW                | Required |
| Coverage              | v8/istanbul        | Required |
| CI Integration        | GitHub Actions     | Required |
| Contract Testing      | Zod schemas        | Required |
| Snapshot Testing      | Vitest snapshots   | Required |
| Debug                 | Vitest UI, Playwright Debug | Required |

### Test Commands Quick Reference

| Command | Description |
|---------|-------------|
| `npm test` | Run all tests in watch mode |
| `npm run test:unit` | Run unit tests with coverage |
| `npm run test:e2e` | Run Playwright E2E tests |
| `npm run test:visual` | Run visual regression tests |
| `npm run test:a11y` | Run accessibility tests |
| `npm run test:all` | Run complete test suite |
| `npx vitest --ui` | Open Vitest UI |
| `npx playwright test --debug` | Debug E2E tests |
