# Notifications System

> Comprehensive notification architecture for Nexus AI — real-time delivery, multi-channel preferences, and action-driven alerts.

## 1. Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    NOTIFICATION FLOW                      │
│                                                          │
│  Backend Event ──► WebSocket ──► Notification Store      │
│       │                              │                   │
│       ├──► Push (FCM/APNs)          ├──► Bell Icon       │
│       ├──► Email (SendGrid)         ├──► Dropdown Panel  │
│       └──► SMS (Twilio)             ├──► Toast Alert     │
│                                      └──► Notification   │
│                                           Center (Page)  │
└─────────────────────────────────────────────────────────┘
```

## 2. Notification Bell Icon

The bell icon lives in the header and serves as the primary entry point.

```tsx
// components/header/NotificationBell.tsx
import { Bell } from "lucide-react";
import { useNotifications } from "@/hooks/useNotifications";
import { NotificationDropdown } from "./NotificationDropdown";

export function NotificationBell() {
  const { unreadCount, isConnected } = useNotifications();
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        aria-label={`Notifications. ${unreadCount} unread`}
        aria-expanded={isOpen}
        aria-haspopup="true"
        className="relative p-2 rounded-lg hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors"
      >
        <Bell className="h-5 w-5 text-neutral-600 dark:text-neutral-300" />
        {unreadCount > 0 && (
          <span
            className="absolute -top-0.5 -right-0.5 flex items-center justify-center
                        min-w-[18px] h-[18px] px-1 text-[10px] font-bold
                        text-white bg-red-500 rounded-full"
            aria-hidden="true"
          >
            {unreadCount > 99 ? "99+" : unreadCount}
          </span>
        )}
        {!isConnected && (
          <span className="absolute bottom-0 right-0 h-2 w-2 rounded-full bg-yellow-400" />
        )}
      </button>
      {isOpen && (
        <NotificationDropdown onClose={() => setIsOpen(false)} />
      )}
    </div>
  );
}
```

### Badge Count Logic

| Condition | Display |
|---|---|
| `count === 0` | No badge |
| `count > 0 && count <= 99` | Exact count |
| `count > 99` | `99+` |
| WebSocket disconnected | Yellow dot indicator |

## 3. Notification Dropdown Panel

```tsx
// components/header/NotificationDropdown.tsx
import { useNotifications } from "@/hooks/useNotifications";
import { NotificationItem } from "./NotificationItem";

interface Props {
  onClose: () => void;
}

export function NotificationDropdown({ onClose }: Props) {
  const {
    notifications,
    unreadCount,
    markAsRead,
    markAllAsRead,
    clearAll,
  } = useNotifications();

  const dropdownRef = useRef<HTMLDivElement>(null);
  useClickOutside(dropdownRef, onClose);
  useFocusTrap(dropdownRef);

  return (
    <div
      ref={dropdownRef}
      role="menu"
      aria-label="Notifications"
      className="absolute right-0 top-full mt-2 w-96 max-h-[480px]
                 bg-white dark:bg-neutral-900 rounded-xl shadow-2xl
                 border border-neutral-200 dark:border-neutral-700
                 overflow-hidden z-50"
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-neutral-100 dark:border-neutral-800">
        <h2 className="text-sm font-semibold">Notifications</h2>
        <div className="flex gap-2">
          {unreadCount > 0 && (
            <button
              onClick={markAllAsRead}
              className="text-xs text-blue-600 hover:text-blue-700"
            >
              Mark all read
            </button>
          )}
          <button
            onClick={clearAll}
            className="text-xs text-neutral-500 hover:text-red-600"
          >
            Clear all
          </button>
        </div>
      </div>

      {/* Notification List */}
      <div className="overflow-y-auto max-h-[380px]" role="list">
        {notifications.length === 0 ? (
          <div className="px-4 py-8 text-center text-neutral-500 text-sm">
            No notifications
          </div>
        ) : (
          notifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              onRead={markAsRead}
            />
          ))
        )}
      </div>

      {/* Footer */}
      <div className="border-t border-neutral-100 dark:border-neutral-800 px-4 py-2">
        <Link
          href="/notifications"
          onClick={onClose}
          className="text-xs text-blue-600 hover:underline"
        >
          View all notifications
        </Link>
      </div>
    </div>
  );
}
```

## 4. Notification Types

| Type | Icon | Category | Default Severity |
|---|---|---|---|
| `ai_task_complete` | ✅ | AI | `success` |
| `document_processed` | 📄 | Documents | `success` |
| `security_alert` | 🛡️ | Security | `error` |
| `workflow_step` | ⚙️ | Workflows | `info` |
| `approval_request` | ✋ | Approvals | `warning` |
| `system_alert` | 🖥️ | System | `warning` |
| `agent_failure` | ❌ | Agents | `error` |
| `mention` | 💬 | Chat | `info` |
| `billing` | 💳 | Billing | `info` |

### Notification Type Definition

```ts
// types/notification.ts
export type NotificationType =
  | "ai_task_complete"
  | "document_processed"
  | "security_alert"
  | "workflow_step"
  | "approval_request"
  | "system_alert"
  | "agent_failure"
  | "mention"
  | "billing";

export type NotificationSeverity = "info" | "warning" | "error" | "success";

export interface Notification {
  id: string;
  type: NotificationType;
  severity: NotificationSeverity;
  title: string;
  message: string;
  data?: Record<string, unknown>;
  read: boolean;
  createdAt: string;
  actions?: NotificationAction[];
  group?: string;
  tenantId: string;
}

export interface NotificationAction {
  id: string;
  label: string;
  variant: "primary" | "danger" | "ghost";
  endpoint?: string;
  method?: "POST" | "PUT" | "DELETE";
}

export interface NotificationPreference {
  type: NotificationType;
  channels: {
    inApp: boolean;
    email: boolean;
    push: boolean;
    sms: boolean;
  };
}
```

### NotificationItem Component

```tsx
// components/notifications/NotificationItem.tsx
import { formatDistanceToNow } from "date-fns";

const SEVERITY_STYLES: Record<NotificationSeverity, string> = {
  info: "bg-blue-50 border-l-blue-500 dark:bg-blue-950/30",
  warning: "bg-amber-50 border-l-amber-500 dark:bg-amber-950/30",
  error: "bg-red-50 border-l-red-500 dark:bg-red-950/30",
  success: "bg-green-50 border-l-green-500 dark:bg-green-950/30",
};

const TYPE_ICONS: Record<NotificationType, React.ReactNode> = {
  ai_task_complete: <CheckCircle className="h-5 w-5 text-green-500" />,
  security_alert: <ShieldAlert className="h-5 w-5 text-red-500" />,
  agent_failure: <AlertTriangle className="h-5 w-5 text-red-500" />,
  approval_request: <Hand className="h-5 w-5 text-amber-500" />,
  // ... other types
};

export function NotificationItem({
  notification,
  onRead,
}: {
  notification: Notification;
  onRead: (id: string) => void;
}) {
  return (
    <div
      role="listitem"
      onClick={() => !notification.read && onRead(notification.id)}
      className={`px-4 py-3 border-l-4 cursor-pointer transition-colors
                  hover:bg-neutral-50 dark:hover:bg-neutral-800/50
                  ${SEVERITY_STYLES[notification.severity]}
                  ${notification.read ? "opacity-60" : ""}`}
      aria-read={notification.read}
    >
      <div className="flex items-start gap-3">
        <div className="mt-0.5">{TYPE_ICONS[notification.type]}</div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <p className="text-sm font-medium truncate">{notification.title}</p>
            {!notification.read && (
              <span className="h-2 w-2 rounded-full bg-blue-500 flex-shrink-0" />
            )}
          </div>
          <p className="text-xs text-neutral-600 dark:text-neutral-400 mt-0.5 line-clamp-2">
            {notification.message}
          </p>
          <time className="text-[10px] text-neutral-400 mt-1 block">
            {formatDistanceToNow(new Date(notification.createdAt), {
              addSuffix: true,
            })}
          </time>
        </div>
      </div>
      {notification.actions && (
        <div className="flex gap-2 mt-2 ml-8">
          {notification.actions.map((action) => (
            <NotificationAction key={action.id} action={action} />
          ))}
        </div>
      )}
    </div>
  );
}
```

## 5. Notification Severity

| Severity | Visual Treatment | Sound | Auto-dismiss |
|---|---|---|---|
| `info` | Blue accent, info icon | Subtle chime | After 5s |
| `success` | Green accent, check icon | Success tone | After 3s |
| `warning` | Amber accent, warning icon | Alert tone | No auto-dismiss |
| `error` | Red accent, error icon | Error tone | No auto-dismiss |

## 6. Real-Time Delivery via WebSocket

```ts
// hooks/useNotifications.ts
import { useNotificationStore } from "@/stores/notificationStore";
import { useWebSocket } from "@/hooks/useWebSocket";

export function useNotifications() {
  const store = useNotificationStore();

  const { isConnected } = useWebSocket({
    url: `${WS_BASE_URL}/notifications`,
    onMessage: (event) => {
      const data = JSON.parse(event.data);

      switch (data.type) {
        case "notification.new":
          store.addNotification(data.payload);
          if (data.payload.severity === "error") {
            toast.error(data.payload.title, {
              description: data.payload.message,
              action: data.payload.actions,
            });
          }
          break;
        case "notification.read":
          store.markAsRead(data.payload.id);
          break;
        case "notification.cleared":
          store.clearAll();
          break;
        case "notification.batch":
          store.addNotifications(data.payload.notifications);
          break;
      }
    },
    reconnectAttempts: 10,
    reconnectInterval: (attempt) => Math.min(1000 * 2 ** attempt, 30000),
  });

  return {
    notifications: store.notifications,
    unreadCount: store.unreadCount,
    isConnected,
    markAsRead: store.markAsRead,
    markAllAsRead: store.markAllAsRead,
    clearAll: store.clearAll,
  };
}
```

### WebSocket Message Protocol

| Event | Direction | Payload |
|---|---|---|
| `notification.new` | Server → Client | `Notification` |
| `notification.read` | Both | `{ id: string }` |
| `notification.cleared` | Server → Client | `{}` |
| `notification.batch` | Server → Client | `{ notifications: Notification[] }` |
| `notification.subscribe` | Client → Server | `{ types: NotificationType[] }` |
| `notification.ack` | Client → Server | `{ id: string }` |

## 7. Notification Stores (Zustand)

```ts
// stores/notificationStore.ts
import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";

interface NotificationState {
  notifications: Notification[];
  preferences: NotificationPreference[];
  unreadCount: number;
  filter: NotificationFilter | null;
}

interface NotificationActions {
  addNotification: (notification: Notification) => void;
  addNotifications: (notifications: Notification[]) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  clearAll: () => void;
  setFilter: (filter: NotificationFilter | null) => void;
  updatePreference: (type: NotificationType, channels: Partial<Channels>) => void;
}

export const useNotificationStore = create<NotificationState & NotificationActions>()(
  devtools(
    persist(
      immer((set, get) => ({
        notifications: [],
        preferences: DEFAULT_PREFERENCES,
        unreadCount: 0,
        filter: null,

        addNotification: (notification) =>
          set((state) => {
            state.notifications.unshift(notification);
            if (!notification.read) state.unreadCount += 1;
          }),

        markAsRead: (id) =>
          set((state) => {
            const n = state.notifications.find((n) => n.id === id);
            if (n && !n.read) {
              n.read = true;
              state.unreadCount = Math.max(0, state.unreadCount - 1);
            }
          }),

        markAllAsRead: () =>
          set((state) => {
            state.notifications.forEach((n) => (n.read = true));
            state.unreadCount = 0;
          }),

        clearAll: () =>
          set((state) => {
            state.notifications = [];
            state.unreadCount = 0;
          }),

        setFilter: (filter) => set({ filter }),

        updatePreference: (type, channels) =>
          set((state) => {
            const pref = state.preferences.find((p) => p.type === type);
            if (pref) Object.assign(pref.channels, channels);
          }),
      })),
      { name: "nexus-notifications" }
    ),
    { name: "NotificationStore" }
  )
);
```

## 8. Notification Preferences

```tsx
// components/notifications/NotificationPreferences.tsx
export function NotificationPreferences() {
  const { preferences, updatePreference } = useNotificationStore();

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-semibold">Notification Preferences</h2>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b">
              <th className="text-left py-3">Type</th>
              <th className="text-center py-3">In-App</th>
              <th className="text-center py-3">Email</th>
              <th className="text-center py-3">Push</th>
              <th className="text-center py-3">SMS</th>
            </tr>
          </thead>
          <tbody>
            {NOTIFICATION_TYPES.map((type) => {
              const pref = preferences.find((p) => p.type === type.value);
              return (
                <tr key={type.value} className="border-b last:border-0">
                  <td className="py-3">
                    <div className="flex items-center gap-2">
                      {type.icon}
                      <span>{type.label}</span>
                    </div>
                  </td>
                  {(["inApp", "email", "push", "sms"] as const).map((ch) => (
                    <td key={ch} className="text-center py-3">
                      <Switch
                        checked={pref?.channels[ch] ?? false}
                        onCheckedChange={(checked) =>
                          updatePreference(type.value, { [ch]: checked })
                        }
                      />
                    </td>
                  ))}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
```

### Preference Defaults

| Type | In-App | Email | Push | SMS |
|---|---|---|---|---|
| AI Task Complete | ✅ | ✅ | ✅ | ❌ |
| Document Processed | ✅ | ✅ | ❌ | ❌ |
| Security Alert | ✅ | ✅ | ✅ | ✅ |
| Workflow Step | ✅ | ❌ | ❌ | ❌ |
| Approval Request | ✅ | ✅ | ✅ | ❌ |
| System Alert | ✅ | ✅ | ✅ | ✅ |
| Agent Failure | ✅ | ✅ | ✅ | ❌ |

## 9. Push Notification Integration

### Firebase Cloud Messaging (FCM)

```ts
// lib/push/fcm.ts
import { getMessaging, getToken } from "firebase/messaging";
import { firebaseApp } from "@/lib/firebase";

export async function requestFCMPermission(): Promise<string | null> {
  const messaging = getMessaging(firebaseApp);
  const permission = await Notification.requestPermission();

  if (permission !== "granted") return null;

  const token = await getToken(messaging, {
    vapidKey: process.env.NEXT_PUBLIC_FCM_VAPID_KEY,
  });

  // Register token with backend
  await fetch("/api/notifications/push/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ token, provider: "fcm" }),
  });

  return token;
}

// Background message handler
export function setupBackgroundHandler() {
  const messaging = getMessaging(firebaseApp);
  onBackgroundMessage(messaging, (payload) => {
    const { title, body, icon } = payload.notification ?? {};
    self.registration.showNotification(title ?? "", {
      body,
      icon: icon ?? "/icons/notification-icon.png",
      data: payload.data,
      actions: [
        { action: "open", title: "Open" },
        { action: "dismiss", title: "Dismiss" },
      ],
    });
  });
}
```

### APNs (iOS Safari / PWA)

```ts
// lib/push/apns.ts
export async function registerAPNs(): Promise<string | null> {
  if (!("serviceWorker" in navigator) || !("PushManager" in window)) return null;

  const registration = await navigator.serviceWorker.ready;
  const subscription = await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(
      process.env.NEXT_PUBLIC_VAPID_KEY!
    ),
  });

  await fetch("/api/notifications/push/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ subscription, provider: "apns" }),
  });

  return JSON.stringify(subscription);
}
```

## 10. Email Notification Templates

```html
<!-- templates/emails/notification.html -->
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 24px; }
    .container { max-width: 600px; margin: 0 auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
    .header { padding: 24px; background: linear-gradient(135deg, #6366f1, #8b5cf6); color: white; }
    .header h1 { margin: 0; font-size: 20px; font-weight: 600; }
    .content { padding: 24px; }
    .badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; text-transform: uppercase; }
    .badge-success { background: #dcfce7; color: #166534; }
    .badge-error { background: #fee2e2; color: #991b1b; }
    .badge-warning { background: #fef3c7; color: #92400e; }
    .badge-info { background: #dbeafe; color: #1e40af; }
    .action-btn { display: inline-block; padding: 10px 24px; border-radius: 8px; text-decoration: none; font-weight: 600; margin: 8px 4px; }
    .action-primary { background: #6366f1; color: white; }
    .action-danger { background: #ef4444; color: white; }
    .footer { padding: 16px 24px; background: #f9fafb; border-top: 1px solid #e5e7eb; font-size: 12px; color: #6b7280; text-align: center; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>{{title}}</h1>
    </div>
    <div class="content">
      <span class="badge badge-{{severity}}">{{severity}}</span>
      <p style="margin-top: 16px; font-size: 15px; color: #374151; line-height: 1.6;">
        {{message}}
      </p>
      {{#if actions.length}}
      <div style="margin-top: 20px;">
        {{#each actions}}
        <a href="{{../baseUrl}}/api/notifications/action/{{id}}?token={{../token}}"
           class="action-btn action-{{variant}}">{{label}}</a>
        {{/each}}
      </div>
      {{/if}}
    </div>
    <div class="footer">
      Nexus AI — <a href="{{preferencesUrl}}">Manage notification preferences</a>
    </div>
  </div>
</body>
</html>
```

## 11. Notification Grouping

```ts
// lib/notifications/grouping.ts
export function groupNotifications(
  notifications: Notification[]
): NotificationGroup[] {
  const groups = new Map<string, Notification[]>();

  for (const n of notifications) {
    const key = n.group ?? n.type;
    if (!groups.has(key)) groups.set(key, []);
    groups.get(key)!.push(n);
  }

  return Array.from(groups.entries())
    .map(([key, items]) => ({
      key,
      label: getGroupLabel(key),
      notifications: items.sort(
        (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      ),
      unreadCount: items.filter((n) => !n.read).length,
      latestAt: items[0].createdAt,
    }))
    .sort(
      (a, b) => new Date(b.latestAt).getTime() - new Date(a.latestAt).getTime()
    );
}

function getGroupLabel(key: string): string {
  const labels: Record<string, string> = {
    ai_task_complete: "AI Tasks",
    security_alert: "Security",
    approval_request: "Approvals",
    workflow_step: "Workflows",
    agent_failure: "Agent Issues",
  };
  return labels[key] ?? "Other";
}
```

## 12. Notification Actions

```tsx
// components/notifications/NotificationAction.tsx
export function NotificationAction({ action }: { action: NotificationAction }) {
  const [loading, setLoading] = useState(false);

  const handleClick = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setLoading(true);
    try {
      await fetch(action.endpoint!, {
        method: action.method ?? "POST",
        headers: { "Content-Type": "application/json" },
      });
      toast.success(`Action "${action.label}" completed`);
    } catch {
      toast.error(`Failed to ${action.label.toLowerCase()}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button
      size="sm"
      variant={action.variant === "primary" ? "default" : action.variant}
      onClick={handleClick}
      disabled={loading}
      className="text-xs"
    >
      {loading ? <Loader2 className="h-3 w-3 animate-spin" /> : action.label}
    </Button>
  );
}
```

### Real-World Action Examples

| Notification | Actions |
|---|---|
| Approval Request | Approve / Reject / View Details |
| Security Alert | Investigate / Dismiss / Block IP |
| Agent Failure | Retry / View Logs / Restart |
| AI Task Complete | View Result / Share / Download |

## 13. Notification Center (Full Page)

```
┌─────────────────────────────────────────────────────────┐
│ Notifications                                   ⚙️ Prefs │
├─────────────────────────────────────────────────────────┤
│ Filter: [All ▾] [Unread ▾] [Type ▾] [Date Range ▾]     │
├─────────────────────────────────────────────────────────┤
│ TODAY                                                   │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ 🛡️ Security Alert: Unusual login detected           │ │
│ │ Login from new IP in Berlin, Germany.     2m ago    │ │
│ │ [Investigate] [Dismiss]                             │ │
│ ├─────────────────────────────────────────────────────┤ │
│ │ ✅ AI Task Complete: Report generated               │ │
│ │ Quarterly analysis report is ready.      15m ago    │ │
│ │ [View Result]                                       │ │
│ └─────────────────────────────────────────────────────┘ │
│ YESTERDAY                                               │
│ ...                                                     │
├─────────────────────────────────────────────────────────┤
│ ← Page 1 of 5 →                                        │
└─────────────────────────────────────────────────────────┘
```

## 14. Notification API Integration

| Endpoint | Method | Description |
|---|---|---|
| `GET /api/notifications` | `GET` | List notifications (paginated) |
| `GET /api/notifications/unread-count` | `GET` | Get unread count |
| `PUT /api/notifications/:id/read` | `PUT` | Mark single as read |
| `PUT /api/notifications/read-all` | `PUT` | Mark all as read |
| `DELETE /api/notifications` | `DELETE` | Clear all notifications |
| `GET /api/notifications/preferences` | `GET` | Get preferences |
| `PUT /api/notifications/preferences` | `PUT` | Update preferences |
| `POST /api/notifications/action/:id` | `POST` | Execute action |

## 15. Responsive Design

| Breakpoint | Bell Behavior | Dropdown Behavior |
|---|---|---|
| Mobile (<640px) | Bell in header | Full-screen slide-up panel |
| Tablet (641–1024px) | Bell in header | Dropdown panel (w-96) |
| Desktop (>1025px) | Bell in header | Dropdown panel (w-96) |

```tsx
// Mobile: full-screen notification panel
function MobileNotificationPanel({ onClose }) {
  return (
    <div className="fixed inset-0 z-50 bg-white dark:bg-neutral-900
                    animate-in slide-in-from-bottom duration-300">
      <div className="flex items-center justify-between px-4 py-3 border-b">
        <h2 className="text-lg font-semibold">Notifications</h2>
        <Button variant="ghost" onClick={onClose}>
          <X className="h-5 w-5" />
        </Button>
      </div>
      <NotificationList />
    </div>
  );
}
```

## 16. Accessibility

| Requirement | Implementation |
|---|---|
| Bell button | `aria-label`, `aria-expanded`, `aria-haspopup` |
| Badge count | Announced via `aria-label` on button |
| Dropdown | `role="menu"`, `role="menuitem"` |
| Notification item | `role="listitem"`, `aria-read` |
| Actions | `aria-label` describing action purpose |
| Keyboard navigation | ↑↓ to navigate, Enter to select, Escape to close |
| Live region | New notifications announced via `aria-live="polite"` |
| Screen reader | Title + message read aloud on focus |

```tsx
// Accessible live region for new notifications
function NotificationLiveRegion() {
  const { latestNotification } = useNotifications();
  return (
    <div aria-live="polite" aria-atomic="true" className="sr-only">
      {latestNotification && (
        <span>
          New {latestNotification.severity} notification: {latestNotification.title}
        </span>
      )}
    </div>
  );
}
```

## 17. Notification Hooks

```ts
// hooks/useNotificationPreferences.ts
export function useNotificationPreferences() {
  const { data: preferences, isLoading } = useQuery({
    queryKey: ["notification-preferences"],
    queryFn: () => api.get("/api/notifications/preferences").then((r) => r.data),
  });

  const updateMutation = useMutation({
    mutationFn: (prefs: Partial<NotificationPreference>[]) =>
      api.put("/api/notifications/preferences", { preferences: prefs }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notification-preferences"] });
    },
  });

  return { preferences, isLoading, updatePreference: updateMutation.mutate };
}

// hooks/useNotificationSound.ts
export function useNotificationSound() {
  const audioRef = useRef<HTMLAudioElement | null>(null);

  const playSound = useCallback((severity: NotificationSeverity) => {
    if (!audioRef.current) {
      audioRef.current = new Audio();
    }
    const sounds: Record<string, string> = {
      info: "/sounds/notification-info.mp3",
      success: "/sounds/notification-success.mp3",
      warning: "/sounds/notification-warning.mp3",
      error: "/sounds/notification-error.mp3",
    };
    audioRef.current.src = sounds[severity];
    audioRef.current.play().catch(() => {});
  }, []);

  return { playSound };
}
```

## 18. File Structure

```
src/
├── components/
│   └── notifications/
│       ├── NotificationBell.tsx
│       ├── NotificationDropdown.tsx
│       ├── NotificationItem.tsx
│       ├── NotificationAction.tsx
│       ├── NotificationPreferences.tsx
│       ├── NotificationCenter.tsx
│       ├── NotificationLiveRegion.tsx
│       └── MobileNotificationPanel.tsx
├── hooks/
│   ├── useNotifications.ts
│   ├── useNotificationPreferences.ts
│   └── useNotificationSound.ts
├── stores/
│   └── notificationStore.ts
├── types/
│   └── notification.ts
├── lib/
│   ├── notifications/
│   │   ├── grouping.ts
│   │   └── templates.ts
│   └── push/
│       ├── fcm.ts
│       └── apns.ts
└── templates/
    └── emails/
        └── notification.html
```
