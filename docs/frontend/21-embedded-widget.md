# 21 — Embedded AI Assistant Widget

## Embeddable AI Chat in AeroXe Ecosystem Applications

---

## 1. Widget Concept

The AeroXe Embedded Widget is a self-contained, embeddable AI chat component that can be integrated into any AeroXe application (or third-party apps) via a simple script tag, iframe, or web component. It provides the full AI assistant experience without requiring the host application to build the chat UI.

```
┌──────────────────────────────────────────────────┐
│  Host Application (ERP, CRM, Billing, etc.)      │
│                                                   │
│  ┌─────────────────────────────┐                  │
│  │  App Content                │                  │
│  │                             │                  │
│  │                             │                  │
│  │                    ┌────────┴────────┐         │
│  │                    │  AeroXe Widget  │         │
│  │                    │  ┌──────────┐   │         │
│  │                    │  │ 💬 Chat  │   │         │
│  │                    │  └──────────┘   │         │
│  │                    │  Floating Button│         │
│  │                    └─────────────────┘         │
│  └─────────────────────────────┘                  │
└──────────────────────────────────────────────────┘
```

---

## 2. Widget Architecture

### Integration Methods

| Method | Complexity | Customization | Best For |
|--------|-----------|---------------|----------|
| Script Tag | Low | Limited | Quick embed, third-party |
| Web Component | Medium | Full | Modern frameworks |
| React Component | Low | Full | React applications |
| iframe | Low | None | Sandboxed, third-party |

### Script Tag Integration

```html
<!-- Add to any HTML page -->
<script
  src="https://cdn.aeroxe.com/widget/v1/aeroxe-widget.js"
  data-api-key="YOUR_API_KEY"
  data-tenant-id="YOUR_TENANT_ID"
  data-position="bottom-right"
  data-theme="light"
  data-agent="default"
  async
  defer
></script>
```

### Web Component Integration

```html
<!-- Register web component -->
<script src="https://cdn.aeroxe.com/widget/v1/aeroxe-widget.js"></script>

<!-- Use anywhere -->
<aeroxe-chat
  api-key="YOUR_API_KEY"
  tenant-id="YOUR_TENANT_ID"
  position="bottom-right"
  theme="light"
></aeroxe-chat>
```

### React Component Integration

```tsx
import { AeroxeWidget } from '@aeroxe/widget-react';

function App() {
  return (
    <div>
      <h1>My ERP Application</h1>
      <AeroxeWidget
        apiKey={process.env.NEXT_PUBLIC_AEROXE_API_KEY}
        tenantId="tenant-1"
        position="bottom-right"
        theme="light"
        agent="erp-assistant"
        onOpen={() => console.log('Widget opened')}
        onMessage={(msg) => console.log('Message:', msg)}
      />
    </div>
  );
}
```

---

## 3. Widget Initialization

### Configuration Object

```typescript
interface AeroxeWidgetConfig {
  // Required
  apiKey: string;
  tenantId: string;

  // Position
  position?: 'bottom-right' | 'bottom-left' | 'top-right' | 'top-left';
  offsetX?: number;
  offsetY?: number;

  // Theme
  theme?: 'light' | 'dark' | 'auto';
  primaryColor?: string;
  fontFamily?: string;
  borderRadius?: number;

  // Agent
  agent?: string;
  agentName?: string;
  greeting?: string;

  // UI
  title?: string;
  subtitle?: string;
  placeholder?: string;
  showAgentAvatar?: boolean;
  showTimestamp?: boolean;
  allowFileUpload?: boolean;
  allowVoiceInput?: boolean;

  // Behavior
  openOnLoad?: boolean;
  persistConversation?: boolean;
  maxMessageLength?: number;

  // Auth
  userToken?: string;
  userId?: string;
  userEmail?: string;
  userName?: string;

  // Events
  onOpen?: () => void;
  onClose?: () => void;
  onMessage?: (message: WidgetMessage) => void;
  onSend?: (content: string) => void;
  onError?: (error: WidgetError) => void;
  onAgentSwitch?: (agentId: string) => void;

  // Advanced
  apiBase?: string;
  wsUrl?: string;
  locale?: string;
  debug?: boolean;
}
```

### Default Configuration

```typescript
const DEFAULT_CONFIG: Partial<AeroxeWidgetConfig> = {
  position: 'bottom-right',
  offsetX: 20,
  offsetY: 20,
  theme: 'light',
  primaryColor: '#6366f1',
  fontFamily: 'Inter, system-ui, sans-serif',
  borderRadius: 12,
  title: 'AeroXe Assistant',
  subtitle: 'AI-powered help',
  placeholder: 'Type your message...',
  showAgentAvatar: true,
  showTimestamp: true,
  allowFileUpload: true,
  allowVoiceInput: false,
  openOnLoad: false,
  persistConversation: true,
  maxMessageLength: 4000,
  locale: 'en',
  debug: false,
};
```

---

## 4. Widget UI

### Floating Button

```
    ┌─────────────────┐
    │  💬  (badge: 3) │  ← Floating button with unread count
    └─────────────────┘
         ↓ click
    ┌─────────────────┐
    │  ═══            │  ← Close icon
    │  AeroXe Chat    │
    │  AI Assistant   │
    │                 │
    │  User: Hello... │
    │  Bot: Hi! How   │
    │  can I help?    │
    │                 │
    │ [📎] [Type...] [🎤] [➤]│
    └─────────────────┘
```

### Chat Bubble States

```typescript
// Widget button states
type WidgetButtonState = 'idle' | 'active' | 'loading' | 'unread';

// Button animations
const buttonStates = {
  idle: 'scale(1)',
  active: 'scale(1.1)',
  loading: 'pulse 1.5s infinite',
  unread: 'bounce 0.5s ease-in-out',
};
```

### Expanded Panel

```typescript
// src/widget/components/chat-panel.tsx
export function ChatPanel({ config, messages, onSend, onClose }: ChatPanelProps) {
  return (
    <div
      className="aeroxe-widget-panel"
      style={{
        position: 'fixed',
        bottom: config.offsetY + 70,
        right: config.offsetX,
        width: 380,
        height: 520,
        borderRadius: config.borderRadius,
        fontFamily: config.fontFamily,
        boxShadow: '0 8px 32px rgba(0,0,0,0.15)',
        zIndex: 2147483647,
      }}
    >
      <WidgetHeader
        title={config.title}
        subtitle={config.subtitle}
        agentName={config.agentName}
        onClose={onClose}
      />
      <MessageList messages={messages} showTimestamp={config.showTimestamp} />
      <InputBar
        placeholder={config.placeholder}
        onSend={onSend}
        allowFileUpload={config.allowFileUpload}
        allowVoiceInput={config.allowVoiceInput}
        maxLength={config.maxMessageLength}
      />
    </div>
  );
}
```

---

## 5. Widget Theming

### CSS Custom Properties

```css
/* Widget theme variables */
:root {
  --aeroxe-widget-primary: #6366f1;
  --aeroxe-widget-primary-hover: #4f46e5;
  --aeroxe-widget-bg: #ffffff;
  --aeroxe-widget-bg-secondary: #f8fafc;
  --aeroxe-widget-text: #1e293b;
  --aeroxe-widget-text-muted: #64748b;
  --aeroxe-widget-border: #e2e8f0;
  --aeroxe-widget-user-bubble: #6366f1;
  --aeroxe-widget-user-text: #ffffff;
  --aeroxe-widget-bot-bubble: #f1f5f9;
  --aeroxe-widget-bot-text: #1e293b;
  --aeroxe-widget-font: 'Inter', system-ui, sans-serif;
  --aeroxe-widget-radius: 12px;
  --aeroxe-widget-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
}

/* Dark theme */
[data-aeroxe-theme="dark"] {
  --aeroxe-widget-bg: #1e293b;
  --aeroxe-widget-bg-secondary: #334155;
  --aeroxe-widget-text: #f8fafc;
  --aeroxe-widget-text-muted: #94a3b8;
  --aeroxe-widget-border: #475569;
  --aeroxe-widget-bot-bubble: #334155;
  --aeroxe-widget-bot-text: #f8fafc;
}
```

### Custom Logo

```typescript
// Widget with custom branding
<AeroxeWidget
  logo="https://mysite.com/logo.png"
  favicon="https://mysite.com/favicon.ico"
  primaryColor="#059669"
  secondaryColor="#d1fae5"
  fontFamily="'Poppins', sans-serif"
  title="My Company Assistant"
  subtitle="Powered by AeroXe AI"
/>
```

---

## 6. Widget Authentication

### Token Passing

```typescript
// Via SDK
AeroxeWidget.init({
  apiKey: 'YOUR_API_KEY',
  tenantId: 'tenant-1',
  userToken: 'jwt-token-from-host-app',
  userId: 'user-123',
  userEmail: 'user@company.com',
  userName: 'John Doe',
});
```

### SSO Integration

```typescript
// Widget receives auth from parent app via postMessage
window.addEventListener('message', (event) => {
  if (event.origin === 'https://app.aeroxe.com') {
    if (event.data.type === 'AEROXE_AUTH') {
      AeroxeWidget.setAuth({
        token: event.data.token,
        refreshToken: event.data.refreshToken,
        user: event.data.user,
      });
    }
  }
});

// Parent app sends auth
window.aeroxeWidgetFrame.contentWindow.postMessage({
  type: 'AEROXE_AUTH',
  token: jwtToken,
  refreshToken: refreshToken,
  user: { id, email, name },
}, 'https://widget.aeroxe.com');
```

---

## 7. Widget API Integration

### Streaming Message Display

```typescript
// Widget handles streaming tokens
class WidgetChatManager {
  private ws: WebSocket;
  private onToken: (token: string) => void;
  private onCompleted: (messageId: string) => void;

  constructor(config: AeroxeWidgetConfig) {
    this.ws = new WebSocket(`${config.wsUrl}?token=${config.userToken}`);

    this.ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      switch (msg.type) {
        case 'token':
          this.onToken(msg.token);
          break;
        case 'completed':
          this.onCompleted(msg.message_id);
          break;
        case 'tool_call':
          this.handleToolCall(msg);
          break;
        case 'error':
          this.handleError(msg);
          break;
      }
    };
  }

  sendMessage(content: string, agentId: string) {
    this.ws.send(JSON.stringify({
      type: 'chat',
      content,
      agent_id: agentId,
    }));
  }

  handleToolCall(msg: any) {
    // Show tool call indicator in UI
    this.onToken(`\n🔧 Using ${msg.name}...\n`);
  }
}
```

---

## 8. Widget File Upload

```typescript
// Drag-and-drop file upload
class WidgetFileUpload {
  private maxSize = 10 * 1024 * 1024; // 10MB
  private acceptedTypes = [
    'image/png', 'image/jpeg', 'image/gif',
    'application/pdf',
    'text/plain', 'text/csv',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  ];

  validate(file: File): { valid: boolean; error?: string } {
    if (file.size > this.maxSize) {
      return { valid: false, error: `File exceeds ${this.maxSize / 1024 / 1024}MB limit` };
    }
    if (!this.acceptedTypes.includes(file.type)) {
      return { valid: false, error: `File type ${file.type} not supported` };
    }
    return { valid: true };
  }

  async upload(file: File, config: AeroxeWidgetConfig): Promise<string> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('tenant_id', config.tenantId);

    const response = await fetch(`${config.apiBase}/api/v1/widget/upload`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${config.userToken}` },
      body: formData,
    });

    const data = await response.json();
    return data.url;
  }
}
```

---

## 9. Widget Voice Input

```typescript
// Speech-to-text in widget
class WidgetVoiceInput {
  private recognition: SpeechRecognition | null = null;
  private isRecording = false;

  constructor(onResult: (text: string) => void) {
    if ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window) {
      const SpeechRecognition = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
      this.recognition = new SpeechRecognition();
      this.recognition.continuous = false;
      this.recognition.interimResults = true;
      this.recognition.lang = 'en-US';

      this.recognition.onresult = (event: SpeechRecognitionEvent) => {
        const transcript = Array.from(event.results)
          .map((result) => result[0].transcript)
          .join('');
        onResult(transcript);
      };
    }
  }

  start() {
    if (this.recognition && !this.isRecording) {
      this.recognition.start();
      this.isRecording = true;
    }
  }

  stop() {
    if (this.recognition && this.isRecording) {
      this.recognition.stop();
      this.isRecording = false;
    }
  }

  get isSupported(): boolean {
    return this.recognition !== null;
  }
}
```

---

## 10. Widget Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | string | required | API authentication key |
| `tenantId` | string | required | Tenant identifier |
| `position` | enum | `bottom-right` | Button position |
| `offsetX` | number | `20` | Horizontal offset (px) |
| `offsetY` | number | `20` | Vertical offset (px) |
| `theme` | enum | `light` | Color theme |
| `primaryColor` | string | `#6366f1` | Primary brand color |
| `fontFamily` | string | `Inter` | Widget font |
| `borderRadius` | number | `12` | Panel border radius |
| `agent` | string | `default` | Default agent ID |
| `title` | string | `AeroXe Assistant` | Panel header title |
| `subtitle` | string | `AI-powered help` | Panel header subtitle |
| `placeholder` | string | `Type your message...` | Input placeholder |
| `greeting` | string | `Hello! How can I help?` | Initial bot message |
| `showAgentAvatar` | boolean | `true` | Show agent avatar |
| `showTimestamp` | boolean | `true` | Show message timestamps |
| `allowFileUpload` | boolean | `true` | Enable file upload |
| `allowVoiceInput` | boolean | `false` | Enable voice input |
| `openOnLoad` | boolean | `false` | Auto-open widget |
| `persistConversation` | boolean | `true` | Save conversations |
| `maxMessageLength` | number | `4000` | Max input length |
| `locale` | string | `en` | Language code |
| `debug` | boolean | `false` | Enable debug logs |

---

## 11. Widget Events

```typescript
// Event system
interface WidgetEvents {
  onOpen: () => void;
  onClose: () => void;
  onMessage: (message: WidgetMessage) => void;
  onSend: (content: string) => void;
  onError: (error: WidgetError) => void;
  onAgentSwitch: (agentId: string) => void;
  onFileUpload: (file: File) => void;
  onVoiceStart: () => void;
  onVoiceEnd: (transcript: string) => void;
  onConversationStart: (conversationId: string) => void;
  onConversationEnd: (conversationId: string) => void;
}

// SDK event binding
AeroxeWidget.on('open', () => analytics.track('widget_opened'));
AeroxeWidget.on('message', (msg) => {
  analytics.track('widget_message', {
    role: msg.role,
    length: msg.content.length,
  });
});
AeroxeWidget.on('error', (err) => {
  sentry.captureException(err);
});
```

---

## 12. Widget SDK

### JavaScript SDK

```typescript
// Aeroxe Widget SDK
declare global {
  interface Window {
    AeroxeWidget: AeroxeWidgetSDK;
  }
}

class AeroxeWidgetSDK {
  private config: AeroxeWidgetConfig;
  private container: HTMLDivElement | null = null;
  private iframe: HTMLIFrameElement | null = null;
  private isOpen = false;
  private eventEmitter = new EventTarget();

  init(config: AeroxeWidgetConfig) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.createContainer();
    this.createButton();

    if (this.config.openOnLoad) {
      this.open();
    }

    if (this.config.debug) {
      console.log('[Aeroxe Widget] Initialized', this.config);
    }
  }

  open() {
    this.isOpen = true;
    this.createPanel();
    this.eventEmitter.dispatchEvent(new CustomEvent('open'));
    this.config.onOpen?.();
  }

  close() {
    this.isOpen = false;
    this.removePanel();
    this.eventEmitter.dispatchEvent(new CustomEvent('close'));
    this.config.onClose?.();
  }

  toggle() {
    this.isOpen ? this.close() : this.open();
  }

  on(event: string, handler: Function) {
    this.eventEmitter.addEventListener(event, handler as EventListener);
  }

  setAuth(auth: { token: string; user: any }) {
    // Update widget authentication
    this.iframe?.contentWindow?.postMessage({
      type: 'AEROXE_AUTH',
      ...auth,
    }, '*');
  }

  destroy() {
    this.container?.remove();
    this.eventEmitter.dispatchEvent(new CustomEvent('destroy'));
  }

  private createContainer() {
    this.container = document.createElement('div');
    this.container.id = 'aeroxe-widget-root';
    this.container.style.cssText = 'position:fixed;z-index:2147483647;';
    document.body.appendChild(this.container);
  }

  private createButton() {
    const btn = document.createElement('button');
    btn.className = 'aeroxe-widget-button';
    btn.innerHTML = '💬';
    btn.setAttribute('aria-label', 'Open AI Assistant');
    btn.onclick = () => this.toggle();

    const pos = this.getPositionStyles();
    btn.style.cssText = pos;
    this.container!.appendChild(btn);
  }

  private createPanel() {
    const panel = document.createElement('div');
    panel.className = 'aeroxe-widget-panel';
    // Panel rendering logic via iframe or shadow DOM
    this.container!.appendChild(panel);
  }

  private removePanel() {
    const panel = this.container?.querySelector('.aeroxe-widget-panel');
    panel?.remove();
  }

  private getPositionStyles(): string {
    const { position, offsetX = 20, offsetY = 20 } = this.config;
    const posMap: Record<string, string> = {
      'bottom-right': `bottom:${offsetY}px;right:${offsetX}px;`,
      'bottom-left': `bottom:${offsetY}px;left:${offsetX}px;`,
      'top-right': `top:${offsetY}px;right:${offsetX}px;`,
      'top-left': `top:${offsetY}px;left:${offsetX}px;`,
    };
    return `position:fixed;${posMap[position || 'bottom-right']}`;
  }
}

// Global export
if (typeof window !== 'undefined') {
  window.AeroxeWidget = new AeroxeWidgetSDK();

  // Auto-init from script tag attributes
  const script = document.querySelector('script[data-api-key]');
  if (script) {
    const config: Partial<AeroxeWidgetConfig> = {
      apiKey: script.getAttribute('data-api-key') || '',
      tenantId: script.getAttribute('data-tenant-id') || '',
      position: (script.getAttribute('data-position') as any) || 'bottom-right',
      theme: (script.getAttribute('data-theme') as any) || 'light',
      agent: script.getAttribute('data-agent') || 'default',
    };
    window.AeroxeWidget.init(config as AeroxeWidgetConfig);
  }
}
```

---

## 13. Widget Security

### Sandboxing

```html
<!-- iframe sandbox policy -->
<iframe
  src="https://widget.aeroxe.com/chat"
  sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
  allow="microphone; camera"
></iframe>
```

### CSP Considerations

```
Content-Security-Policy:
  script-src 'self' https://cdn.aeroxe.com;
  connect-src 'self' https://api.aeroxe.com wss://ws.aeroxe.com;
  frame-src https://widget.aeroxe.com;
  style-src 'self' 'unsafe-inline';
  img-src 'self' https://cdn.aeroxe.com data:;
```

### XSS Prevention

- All user input sanitized before display
- No `innerHTML` usage — React rendering only
- Widget runs in isolated iframe with CSP
- API responses validated with Zod schemas

---

## 14. Widget Deployment

### CDN Distribution

```bash
# Build for CDN
npm run build:widget

# Output
dist/
  aeroxe-widget.js          # Main bundle (~45KB gzipped)
  aeroxe-widget.css          # Styles (~8KB gzipped)
  aeroxe-widget.min.js       # Minified bundle
  aeroxe-widget.min.css      # Minified styles
```

### npm Package

```bash
# Install
npm install @aeroxe/widget-react

# Usage
import { AeroxeWidget } from '@aeroxe/widget-react';
```

### Version Management

```typescript
// Widget version with compatibility check
const WIDGET_VERSION = '1.0.0';
const MIN_API_VERSION = 'v1';

// Load specific version
<script src="https://cdn.aeroxe.com/widget/v1.2.0/aeroxe-widget.js"></script>
```

---

## 15. Widget Responsive Design

```css
/* Mobile breakpoint */
@media (max-width: 480px) {
  .aeroxe-widget-panel {
    width: 100% !important;
    height: 100% !important;
    border-radius: 0 !important;
    top: 0 !important;
    left: 0 !important;
    bottom: 0 !important;
    right: 0 !important;
  }
}

/* Tablet breakpoint */
@media (max-width: 768px) {
  .aeroxe-widget-panel {
    width: 340px !important;
    height: 450px !important;
  }
}
```

---

## 16. Widget Analytics

```typescript
// Widget usage tracking
class WidgetAnalytics {
  private events: any[] = [];

  track(event: string, data?: Record<string, unknown>) {
    this.events.push({
      event,
      data,
      timestamp: new Date().toISOString(),
      widgetVersion: WIDGET_VERSION,
    });

    if (this.events.length >= 20) {
      this.flush();
    }
  }

  async flush() {
    if (this.events.length === 0) return;
    const batch = [...this.events];
    this.events = [];

    try {
      await fetch('/api/v1/widget/analytics', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ events: batch }),
      });
    } catch {
      this.events.unshift(...batch);
    }
  }
}
```

### Tracked Events

| Event | Data |
|-------|------|
| `widget_opened` | timestamp, position, theme |
| `widget_closed` | duration, messages_sent |
| `message_sent` | content_length, agent_id, has_file |
| `message_received` | response_time, token_count, had_tools |
| `file_uploaded` | file_type, file_size |
| `voice_used` | duration, transcript_length |
| `agent_switched` | from_agent, to_agent |
| `error_occurred` | error_type, error_message |
| `conversation_started` | agent_id, user_id |
| `conversation_ended` | message_count, duration |

---

## 17. Widget Documentation (Integration Guide)

### Quick Start

```html
<!DOCTYPE html>
<html>
<head>
  <title>My App</title>
</head>
<body>
  <h1>My Application</h1>

  <!-- Add widget -->
  <script
    src="https://cdn.aeroxe.com/widget/v1/aeroxe-widget.js"
    data-api-key="ak_live_xxxxxxxxxxxx"
    data-tenant-id="tenant-123"
    data-theme="light"
    data-position="bottom-right"
    async
  ></script>
</body>
</html>
```

### Advanced Integration

```html
<script src="https://cdn.aeroxe.com/widget/v1/aeroxe-widget.js"></script>
<script>
  AeroxeWidget.init({
    apiKey: 'ak_live_xxxxxxxxxxxx',
    tenantId: 'tenant-123',
    position: 'bottom-right',
    theme: 'light',
    primaryColor: '#059669',
    title: 'Support Assistant',
    subtitle: 'Ask me anything',
    agent: 'support-agent',
    allowFileUpload: true,
    allowVoiceInput: true,
    onOpen: function() {
      console.log('Widget opened');
    },
    onMessage: function(msg) {
      console.log('New message:', msg);
    },
    onError: function(err) {
      console.error('Widget error:', err);
    }
  });
</script>
```

### API Reference

| Method | Description |
|--------|-------------|
| `AeroxeWidget.init(config)` | Initialize widget |
| `AeroxeWidget.open()` | Open chat panel |
| `AeroxeWidget.close()` | Close chat panel |
| `AeroxeWidget.toggle()` | Toggle panel |
| `AeroxeWidget.setAuth(auth)` | Update authentication |
| `AeroxeWidget.on(event, handler)` | Listen to events |
| `AeroxeWidget.off(event, handler)` | Remove event listener |
| `AeroxeWidget.destroy()` | Remove widget entirely |
| `AeroxeWidget.getConversationId()` | Get current conversation ID |
| `AeroxeWidget.sendMessage(text)` | Programmatically send message |
