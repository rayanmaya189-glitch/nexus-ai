# AeroXe Nexus AI — Dashboard

## KPI Cards, Charts, Real-Time Metrics, Widget Layout, Dashboard Customization

---

## 1. Dashboard Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│  Dashboard Header                                                 │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ Dashboard  │  Date Range: [Last 7d ▼]  │ Tenant: [All ▼]  │  │
│  │            │  Model: [All ▼]            │    ↻ Refresh     │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐                            │
│  │ KPI  │ │ KPI  │ │ KPI  │ │ KPI  │  ← KPI Row                 │
│  │Total │ │Active│ │Token │ │Cost  │                             │
│  │Reqs  │ │Users │ │Usage │ │This  │                             │
│  │12.4K │ │  342 │ │ 8.2M │ │$142  │                             │
│  └──────┘ └──────┘ └──────┘ └──────┘                            │
│                                                                    │
│  ┌───────────────────────┐ ┌───────────────────────┐             │
│  │                       │ │                        │             │
│  │   AI Usage Chart      │ │  Token Consumption     │             │
│  │   (Line/Bar)          │ │  (Stacked Bar)         │             │
│  │                       │ │                        │             │
│  └───────────────────────┘ └───────────────────────┘             │
│                                                                    │
│  ┌───────────────────────┐ ┌───────────────────────┐             │
│  │                       │ │                        │             │
│  │  Model Performance    │ │  Active Agents         │             │
│  │  (Latency/Throughput) │ │  (Status Cards)        │             │
│  │                       │ │                        │             │
│  └───────────────────────┘ └───────────────────────┘             │
│                                                                    │
│  ┌───────────────────────┐ ┌───────────────────────┐             │
│  │                       │ │                        │             │
│  │  System Health        │ │  Recent Activity       │             │
│  │  (Gauges)             │ │  (Feed)                │             │
│  │                       │ │                        │             │
│  └───────────────────────┘ └───────────────────────┘             │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. KPI Cards

### 2.1 KPI Card Component

```tsx
// components/dashboard/KPICard.tsx
interface KPICardProps {
  title: string;
  value: string | number;
  change?: number;           // Percentage change from previous period
  changeType?: "positive" | "negative" | "neutral";
  icon: LucideIcon;
  description?: string;
  loading?: boolean;
}

export function KPICard({ title, value, change, changeType, icon: Icon, description, loading }: KPICardProps) {
  if (loading) {
    return (
      <Card>
        <CardContent className="p-6">
          <Skeleton className="h-4 w-24 mb-2" />
          <Skeleton className="h-8 w-16 mb-2" />
          <Skeleton className="h-3 w-32" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium text-muted-foreground">{title}</p>
          <Icon className="h-4 w-4 text-muted-foreground" />
        </div>
        <div className="mt-2">
          <p className="text-3xl font-bold">{value}</p>
        </div>
        {change !== undefined && (
          <div className="mt-2 flex items-center gap-1">
            {changeType === "positive" ? (
              <TrendingUp className="h-4 w-4 text-green-500" />
            ) : changeType === "negative" ? (
              <TrendingDown className="h-4 w-4 text-red-500" />
            ) : null}
            <span className={cn(
              "text-sm font-medium",
              changeType === "positive" && "text-green-500",
              changeType === "negative" && "text-red-500",
              changeType === "neutral" && "text-muted-foreground"
            )}>
              {change > 0 ? "+" : ""}{change}%
            </span>
            <span className="text-sm text-muted-foreground">vs last period</span>
          </div>
        )}
        {description && (
          <p className="mt-1 text-xs text-muted-foreground">{description}</p>
        )}
      </CardContent>
    </Card>
  );
}
```

### 2.2 KPI Definitions

| KPI | Icon | Value Format | Change Metric |
|---|---|---|---|
| Total Requests | `Activity` | `12.4K` | vs last 7 days |
| Active Users | `Users` | `342` | vs last 7 days |
| Token Usage | `Zap` | `8.2M` | vs last 7 days |
| Monthly Cost | `DollarSign` | `$142` | vs last month |
| Avg Response Time | `Clock` | `1.2s` | vs last 7 days |
| Success Rate | `CheckCircle` | `99.4%` | vs last 7 days |
| Documents Processed | `FileText` | `1,247` | vs last 7 days |
| Active Agents | `Bot` | `12` | current |

### 2.3 KPI Row Layout

```tsx
// Dashboard KPI row
<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
  <KPICard
    title="Total Requests"
    value={formatNumber(metrics.totalRequests)}
    change={metrics.requestsChange}
    changeType={metrics.requestsChange >= 0 ? "positive" : "negative"}
    icon={Activity}
    loading={isLoading}
  />
  <KPICard
    title="Active Users"
    value={metrics.activeUsers}
    change={metrics.usersChange}
    changeType={metrics.usersChange >= 0 ? "positive" : "negative"}
    icon={Users}
    loading={isLoading}
  />
  <KPICard
    title="Token Usage"
    value={formatTokens(metrics.tokenUsage)}
    change={metrics.tokensChange}
    changeType={metrics.tokensChange >= 0 ? "negative" : "positive"} // More tokens = higher cost
    icon={Zap}
    loading={isLoading}
  />
  <KPICard
    title="Monthly Cost"
    value={`$${formatCurrency(metrics.monthlyCost)}`}
    change={metrics.costChange}
    changeType={metrics.costChange <= 0 ? "positive" : "negative"}
    icon={DollarSign}
    loading={isLoading}
  />
</div>
```

---

## 3. Charts

### 3.1 AI Usage Chart (Requests Over Time)

```tsx
// components/dashboard/AIUsageChart.tsx
import { Line, LineChart, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from "recharts";

interface UsageDataPoint {
  date: string;
  requests: number;
  successful: number;
  failed: number;
}

export function AIUsageChart({ data, isLoading }: { data: UsageDataPoint[]; isLoading: boolean }) {
  if (isLoading) return <Card><CardContent><Skeleton className="h-80" /></CardContent></Card>;

  return (
    <Card>
      <CardHeader>
        <CardTitle>AI Usage</CardTitle>
        <CardDescription>Requests over the last 7 days</CardDescription>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
            <XAxis dataKey="date" className="text-xs" tick={{ fill: "var(--muted-foreground)" }} />
            <YAxis className="text-xs" tick={{ fill: "var(--muted-foreground)" }} />
            <Tooltip
              contentStyle={{
                backgroundColor: "var(--card)",
                border: "1px solid var(--border)",
                borderRadius: "8px",
              }}
            />
            <Legend />
            <Line type="monotone" dataKey="requests" stroke="var(--accent)" strokeWidth={2} name="Total" />
            <Line type="monotone" dataKey="successful" stroke="var(--success)" strokeWidth={2} name="Successful" />
            <Line type="monotone" dataKey="failed" stroke="var(--destructive)" strokeWidth={2} name="Failed" />
          </LineChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
```

### 3.2 Token Consumption Chart

```tsx
// components/dashboard/TokenChart.tsx
import { Bar, BarChart, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from "recharts";

interface TokenDataPoint {
  date: string;
  inputTokens: number;
  outputTokens: number;
}

export function TokenChart({ data, isLoading }: { data: TokenDataPoint[]; isLoading: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Token Consumption</CardTitle>
        <CardDescription>Input vs Output tokens</CardDescription>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
            <XAxis dataKey="date" className="text-xs" />
            <YAxis className="text-xs" />
            <Tooltip />
            <Legend />
            <Bar dataKey="inputTokens" stackId="tokens" fill="var(--accent)" name="Input Tokens" radius={[0, 0, 0, 0]} />
            <Bar dataKey="outputTokens" stackId="tokens" fill="var(--primary)" name="Output Tokens" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
```

### 3.3 Model Performance Chart

```tsx
// components/dashboard/ModelPerformanceChart.tsx
interface ModelPerformance {
  model: string;
  avgLatency: number;    // ms
  throughput: number;    // requests/min
  errorRate: number;     // percentage
}

export function ModelPerformanceChart({ data }: { data: ModelPerformance[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Model Performance</CardTitle>
        <CardDescription>Average latency by model</CardDescription>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={data} layout="vertical">
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis type="number" unit="ms" />
            <YAxis dataKey="model" type="category" width={120} />
            <Tooltip formatter={(value: number) => `${value}ms`} />
            <Bar dataKey="avgLatency" fill="var(--accent)" radius={[0, 4, 4, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
```

### 3.4 Chart Color Palette

| Data Series | Color | CSS Variable |
|---|---|---|
| Primary metric | Blue | `var(--accent)` |
| Secondary metric | Navy | `var(--primary)` |
| Success/Positive | Green | `var(--success)` |
| Error/Negative | Red | `var(--destructive)` |
| Warning | Amber | `var(--warning)` |
| Neutral | Gray | `var(--muted-foreground)` |

---

## 4. Active Agents Widget

```tsx
// components/dashboard/ActiveAgentsWidget.tsx
interface AgentStatus {
  id: string;
  name: string;
  model: string;
  status: "active" | "idle" | "offline" | "error";
  currentTasks: number;
  avgLatency: number;
  lastActiveAt: string;
}

export function ActiveAgentsWidget({ agents, isLoading }: { agents: AgentStatus[]; isLoading: boolean }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>Active Agents</CardTitle>
          <CardDescription>{agents.filter(a => a.status === "active").length} currently active</CardDescription>
        </div>
        <Button variant="ghost" size="sm" onClick={() => router.push("/dashboard/agents")}>
          View All <ArrowRight className="ml-1 h-4 w-4" />
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {isLoading ? (
            Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-16 w-full" />
            ))
          ) : (
            agents.slice(0, 5).map((agent) => (
              <div key={agent.id} className="flex items-center justify-between rounded-lg border p-3">
                <div className="flex items-center gap-3">
                  <div className={cn(
                    "h-2 w-2 rounded-full",
                    agent.status === "active" && "bg-green-500",
                    agent.status === "idle" && "bg-yellow-500",
                    agent.status === "offline" && "bg-gray-400",
                    agent.status === "error" && "bg-red-500"
                  )} />
                  <div>
                    <p className="text-sm font-medium">{agent.name}</p>
                    <p className="text-xs text-muted-foreground">{agent.model}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm">{agent.currentTasks} tasks</p>
                  <p className="text-xs text-muted-foreground">{agent.avgLatency}ms avg</p>
                </div>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}
```

---

## 5. System Health Gauges

```tsx
// components/dashboard/SystemHealthGauge.tsx
interface HealthMetric {
  name: string;
  value: number;        // 0-100
  unit: string;
  status: "healthy" | "warning" | "critical";
}

function GaugeBar({ metric }: { metric: HealthMetric }) {
  const colorMap = {
    healthy: "bg-green-500",
    warning: "bg-yellow-500",
    critical: "bg-red-500",
  };

  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-sm">
        <span>{metric.name}</span>
        <span className="text-muted-foreground">{metric.value}{metric.unit}</span>
      </div>
      <div className="h-2 w-full rounded-full bg-muted">
        <div
          className={cn("h-2 rounded-full transition-all duration-500", colorMap[metric.status])}
          style={{ width: `${metric.value}%` }}
        />
      </div>
    </div>
  );
}

export function SystemHealthGauge({ metrics, isLoading }: { metrics: HealthMetric[]; isLoading: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>System Health</CardTitle>
        <CardDescription>Real-time resource utilization</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {isLoading ? (
          Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-10 w-full" />)
        ) : (
          metrics.map((metric) => (
            <GaugeBar key={metric.name} metric={metric} />
          ))
        )}
      </CardContent>
    </Card>
  );
}
```

### Health Metrics

| Metric | Warning Threshold | Critical Threshold | Unit |
|---|---|---|---|
| CPU Usage | > 70% | > 90% | % |
| Memory Usage | > 75% | > 90% | % |
| GPU Utilization | > 80% | > 95% | % |
| Disk Usage | > 70% | > 85% | % |
| Active Connections | > 800 | > 950 | count |

---

## 6. Activity Feed

```tsx
// components/dashboard/ActivityFeed.tsx
interface ActivityItem {
  id: string;
  type: "chat" | "agent" | "document" | "system" | "alert";
  title: string;
  description: string;
  timestamp: string;
  user?: string;
}

const typeIcons = {
  chat: MessageSquare,
  agent: Bot,
  document: FileText,
  system: Settings,
  alert: AlertTriangle,
};

const typeColors = {
  chat: "text-blue-500 bg-blue-500/10",
  agent: "text-purple-500 bg-purple-500/10",
  document: "text-green-500 bg-green-500/10",
  system: "text-gray-500 bg-gray-500/10",
  alert: "text-red-500 bg-red-500/10",
};

export function ActivityFeed({ items, isLoading }: { items: ActivityItem[]; isLoading: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
        <CardDescription>Latest events across the platform</CardDescription>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-80">
          {isLoading ? (
            Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-14 w-full mb-3" />)
          ) : (
            <div className="space-y-3">
              {items.map((item) => {
                const Icon = typeIcons[item.type];
                return (
                  <div key={item.id} className="flex items-start gap-3">
                    <div className={cn("rounded-full p-2", typeColors[item.type])}>
                      <Icon className="h-4 w-4" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{item.title}</p>
                      <p className="text-xs text-muted-foreground truncate">{item.description}</p>
                      <p className="text-xs text-muted-foreground mt-1">
                        {item.user && `${item.user} · `}{formatRelativeTime(item.timestamp)}
                      </p>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
```

---

## 7. Quick Actions

```tsx
// components/dashboard/QuickActions.tsx
const quickActions = [
  { label: "Start Chat", icon: MessageSquare, href: "/dashboard/chat", color: "bg-blue-500" },
  { label: "Upload Document", icon: Upload, href: "/dashboard/documents", color: "bg-green-500" },
  { label: "Create Agent", icon: Bot, href: "/dashboard/agents/new", color: "bg-purple-500" },
  { label: "View Alerts", icon: Bell, href: "/dashboard/audit", color: "bg-red-500" },
];

export function QuickActions() {
  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
      {quickActions.map((action) => (
        <Button
          key={action.label}
          variant="outline"
          className="flex flex-col items-center gap-2 h-20"
          onClick={() => router.push(action.href)}
        >
          <div className={cn("rounded-full p-2 text-white", action.color)}>
            <action.icon className="h-4 w-4" />
          </div>
          <span className="text-xs">{action.label}</span>
        </Button>
      ))}
    </div>
  );
}
```

---

## 8. Dashboard Filters

```tsx
// Dashboard filter bar
export function DashboardFilters() {
  const { filters, setDateRange, setTenantFilter, setModelFilter } = useDashboardStore();

  return (
    <div className="flex flex-wrap items-center gap-3">
      {/* Date Range Picker */}
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline" size="sm">
            <CalendarDays className="mr-2 h-4 w-4" />
            {formatDateRange(filters.dateRange)}
          </Button>
        </PopoverTrigger>
        <PopoverContent>
          <DateRangePicker
            value={filters.dateRange}
            onChange={setDateRange}
            presets={[
              { label: "Last 24 hours", value: "24h" },
              { label: "Last 7 days", value: "7d" },
              { label: "Last 30 days", value: "30d" },
              { label: "This month", value: "month" },
              { label: "Custom", value: "custom" },
            ]}
          />
        </PopoverContent>
      </Popover>

      {/* Tenant Filter */}
      <Select value={filters.tenantId} onValueChange={setTenantFilter}>
        <SelectTrigger className="w-40">
          <Building2 className="mr-2 h-4 w-4" />
          <SelectValue placeholder="All Tenants" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Tenants</SelectItem>
          {tenants.map((t) => (
            <SelectItem key={t.id} value={t.id}>{t.name}</SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Model Filter */}
      <Select value={filters.model} onValueChange={setModelFilter}>
        <SelectTrigger className="w-40">
          <Cpu className="mr-2 h-4 w-4" />
          <SelectValue placeholder="All Models" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Models</SelectItem>
          {models.map((m) => (
            <SelectItem key={m} value={m}>{m}</SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Refresh */}
      <Button variant="ghost" size="icon" onClick={refreshDashboard}>
        <RefreshCw className={cn("h-4 w-4", isRefreshing && "animate-spin")} />
      </Button>
    </div>
  );
}
```

---

## 9. Real-Time Updates

### 9.1 WebSocket Integration

```typescript
// hooks/dashboard/useMetrics.ts
export function useMetrics(filters: DashboardFilters) {
  const queryClient = useQueryClient();

  // API query for initial data
  const { data: metrics, isLoading } = useQuery({
    queryKey: ["dashboard", "metrics", filters],
    queryFn: () => dashboardApi.getMetrics(filters),
    refetchInterval: 30_000, // Poll every 30s as fallback
  });

  // WebSocket for real-time updates
  useEffect(() => {
    const ws = new WebSocket(`${WS_URL}/ws/metrics`);

    ws.onmessage = (event) => {
      const update = JSON.parse(event.data);

      // Update query cache optimistically
      queryClient.setQueryData(["dashboard", "metrics", filters], (old: DashboardMetrics) => ({
        ...old,
        totalRequests: update.totalRequests,
        activeUsers: update.activeUsers,
        tokenUsage: update.tokenUsage,
        health: update.health,
      }));
    };

    return () => ws.close();
  }, [filters, queryClient]);

  return { metrics, isLoading };
}
```

### 9.2 Refresh Strategy

| Data | API Poll | WebSocket | Manual Refresh |
|---|---|---|---|
| KPI values | 30s interval | ✅ Live | ✅ |
| Charts | 60s interval | ✅ Live | ✅ |
| Agent status | 15s interval | ✅ Live | ✅ |
| Health metrics | 10s interval | ✅ Live | ✅ |
| Activity feed | 60s interval | ✅ Live | ✅ |

---

## 10. Dashboard Stores

```typescript
// stores/dashboard-store.ts
interface DashboardState {
  // Filters
  filters: DashboardFilters;
  setDateRange: (range: DateRange) => void;
  setTenantFilter: (tenantId: string) => void;
  setModelFilter: (model: string) => void;

  // Widget visibility
  widgetConfig: WidgetConfig[];
  toggleWidget: (widgetId: string) => void;
  reorderWidgets: (from: number, to: number) => void;

  // UI state
  isRefreshing: boolean;
  refreshDashboard: () => Promise<void>;
}

interface DashboardFilters {
  dateRange: DateRange;
  tenantId: string;
  model: string;
}

interface WidgetConfig {
  id: string;
  title: string;
  visible: boolean;
  order: number;
  size: "sm" | "md" | "lg" | "full";
}

const defaultWidgets: WidgetConfig[] = [
  { id: "kpi", title: "KPI Cards", visible: true, order: 0, size: "full" },
  { id: "ai-usage", title: "AI Usage", visible: true, order: 1, size: "lg" },
  { id: "token-consumption", title: "Token Consumption", visible: true, order: 2, size: "lg" },
  { id: "model-performance", title: "Model Performance", visible: true, order: 3, size: "lg" },
  { id: "active-agents", title: "Active Agents", visible: true, order: 4, size: "lg" },
  { id: "system-health", title: "System Health", visible: true, order: 5, size: "lg" },
  { id: "activity-feed", title: "Activity Feed", visible: true, order: 6, size: "lg" },
];
```

---

## 11. Dashboard API Integration

### 11.1 API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/dashboard/metrics` | Main dashboard metrics |
| `POST` | `/api/v1/dashboard/metrics/requests` | Request count over time |
| `POST` | `/api/v1/dashboard/metrics/tokens` | Token consumption over time |
| `POST` | `/api/v1/dashboard/metrics/models` | Model performance data |
| `POST` | `/api/v1/dashboard/agents/status` | Agent status overview |
| `POST` | `/api/v1/dashboard/health` | System health metrics |
| `POST` | `/api/v1/dashboard/activity` | Recent activity feed |
| `POST` | `/api/v1/dashboard/costs` | Cost breakdown data |

### 11.2 Response Types

```typescript
interface DashboardMetrics {
  totalRequests: number;
  requestsChange: number;
  activeUsers: number;
  usersChange: number;
  tokenUsage: number;
  tokensChange: number;
  monthlyCost: number;
  costChange: number;
  health: HealthMetric[];
  recentActivity: ActivityItem[];
  modelPerformance: ModelPerformance[];
}
```

---

## 12. Dashboard Hooks

| Hook | Purpose | Returns |
|---|---|---|
| `useMetrics(filters)` | Fetch dashboard metrics | `{ metrics, isLoading, refresh }` |
| `useHealth()` | System health data | `{ health, isLoading }` |
| `useActivity(limit)` | Recent activity feed | `{ activities, isLoading, fetchMore }` |
| `useChart(metric, range)` | Chart data | `{ data, isLoading }` |
| `useAgentStatus()` | Agent status list | `{ agents, isLoading }` |

---

## 13. Responsive Dashboard

| Screen | Layout |
|---|---|
| Mobile (< 768px) | Single column, KPI cards stack vertically, charts full-width |
| Tablet (768-1024px) | 2-column grid for KPIs, charts stack |
| Desktop (1024-1440px) | 4-column KPI grid, 2-column charts |
| Wide (> 1440px) | Same as desktop with more whitespace |

### Responsive KPI Grid

```tsx
<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6">
  {/* KPIs adapt: 1 col → 2 col → 4 col */}
</div>

<div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
  {/* Charts: 1 col → 2 col */}
</div>
```

---

## 14. Dashboard Performance

### 14.1 Optimization Techniques

| Technique | Implementation |
|---|---|
| Data caching | TanStack Query stale-while-revalidate |
| Incremental updates | WebSocket push, not full refresh |
| Lazy chart rendering | Load chart libs dynamically |
| Skeleton loading | Show placeholders during fetch |
| Memoized calculations | `useMemo` for derived data |
| Debounced filters | 300ms debounce on filter changes |
| Image optimization | Next.js `Image` component |

### 14.2 Performance Targets

| Metric | Target |
|---|---|
| Initial dashboard load | < 2s |
| Filter change response | < 500ms |
| Real-time update latency | < 1s |
| Chart re-render | < 100ms |
| Memory usage | < 50MB |

---

*Document Version: 1.0 — Last Updated: July 2026*
