# 20 — Ecosystem Dashboards

## Central Business Assistant + Domain Dashboards + AI-Powered Intelligence Views

---

## 1. Dashboard Architecture Overview

AeroXe Nexus AI provides intelligent dashboards for every AeroXe ecosystem product, powered by AI analysis, real-time data, and predictive insights.

```
┌─────────────────────────────────────────────────────────┐
│                    Dashboard Shell                       │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────────────┐ │
│  │ Sidebar│ │ Topbar │ │Search  │ │ Notification Bell│ │
│  └────────┘ └────────┘ └────────┘ └──────────────────┘ │
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │               Dashboard Content Area              │  │
│  │                                                   │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │  │
│  │  │ Stat     │ │ Stat     │ │ Stat Card        │  │  │
│  │  │ Card 1   │ │ Card 2   │ │ Card 3           │  │  │
│  │  └──────────┘ └──────────┘ └──────────────────┘  │  │
│  │                                                   │  │
│  │  ┌──────────────────┐ ┌──────────────────────┐   │  │
│  │  │ Chart Area       │ │ AI Insights Panel    │   │  │
│  │  │ (Recharts)       │ │ (Predictions, etc)   │   │  │
│  │  └──────────────────┘ └──────────────────────┘   │  │
│  │                                                   │  │
│  │  ┌─────────────────────────────────────────────┐  │  │
│  │  │ Data Table / Activity Feed                  │  │  │
│  │  └─────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 2. Central Business Assistant

The cross-platform AI chat interface is the primary entry point for all AeroXe products.

### Features

| Feature | Description |
|---------|-------------|
| Multi-agent chat | Switch between domain-specific agents |
| Conversation history | Persistent chat history per user |
| File upload | Attach documents, images, reports |
| Voice input | Speech-to-text for hands-free queries |
| Streaming responses | Real-time token-by-token display |
| Tool call visualization | Show agent tool calls and results |
| Markdown rendering | Rich text in assistant responses |
| Code highlighting | Syntax-highlighted code blocks |
| Export conversation | Download as PDF or Markdown |
| Share conversation | Generate shareable link |

### Chat Interface Layout

```
┌─────────────────────────────────────────────────────┐
│  Agent: [Broadband Assistant ▼]   [New Chat] [Menu] │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌─────────────────────────────────────────────┐    │
│  │  👤 User: Show me network outage status     │    │
│  └─────────────────────────────────────────────┘    │
│                                                     │
│  ┌─────────────────────────────────────────────┐    │
│  │  🤖 Assistant: Here is the current network  │    │
│  │  status for your region:                     │    │
│  │                                              │    │
│  │  ┌─────────────────────────────────────┐     │    │
│  │  │  OLT-01: Online    ████████░░ 80%   │     │    │
│  │  │  OLT-02: Degraded  ██████░░░░ 60%   │     │    │
│  │  │  OLT-03: Online    █████████░ 90%   │     │    │
│  │  └─────────────────────────────────────┘     │    │
│  │                                              │    │
│  │  ⚠️ Predictive Alert: OLT-02 may need       │    │
│  │  maintenance within 48 hours based on        │    │
│  │  historical patterns.                        │    │
│  │                                              │    │
│  │  [Tool: check_network_status] ✓             │    │
│  └─────────────────────────────────────────────┘    │
│                                                     │
├─────────────────────────────────────────────────────┤
│  📎  [Type your message...              ] [🎤] [➤] │
└─────────────────────────────────────────────────────┘
```

---

## 3. Broadband Network Intelligence Dashboard

### Overview

Real-time monitoring of broadband network infrastructure with AI-powered predictive analytics.

### OLT/ONU Status Cards

```
┌────────────────┐  ┌────────────────┐  ┌────────────────┐
│  OLT-01        │  │  OLT-02        │  │  OLT-03        │
│  ● Online      │  │  ◉ Degraded    │  │  ● Online      │
│                │  │                │  │                │
│  ONUs: 256     │  │  ONUs: 198     │  │  ONUs: 240     │
│  Active: 245   │  │  Active: 175   │  │  Active: 238   │
│  Bandwidth:    │  │  Bandwidth:    │  │  Bandwidth:    │
│  850 Mbps      │  │  620 Mbps      │  │  910 Mbps      │
│                │  │                │  │                │
│  Uptime: 99.9% │  │  Uptime: 97.2% │  │  Uptime: 99.7% │
│  Latency: 2ms  │  │  Latency: 8ms  │  │  Latency: 3ms  │
└────────────────┘  └────────────────┘  └────────────────┘
```

### Dashboard Data Model

```typescript
interface BroadbandDashboard {
  olt_status: OLTStatus[];
  network_outages: NetworkOutage[];
  bandwidth_utilization: BandwidthMetric[];
  predictive_alerts: PredictiveAlert[];
  customer_complaints: ComplaintCorrelation[];
}

interface OLTStatus {
  id: string;
  name: string;
  status: 'online' | 'degraded' | 'offline';
  total_onus: number;
  active_onus: number;
  bandwidth_mbps: number;
  uptime_percent: number;
  latency_ms: number;
  location: { lat: number; lng: number };
}

interface NetworkOutage {
  id: string;
  olt_id: string;
  type: 'partial' | 'complete';
  start_time: string;
  end_time: string | null;
  affected_customers: number;
  cause: string;
  status: 'active' | 'resolved' | 'investigating';
}

interface PredictiveAlert {
  id: string;
  component: string;
  alert_type: 'maintenance' | 'failure' | 'degradation';
  confidence: number;
  predicted_time: string;
  description: string;
  recommended_action: string;
}
```

### Bandwidth Utilization Chart

```typescript
// src/components/broadband/bandwidth-chart.tsx
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface BandwidthChartProps {
  data: BandwidthMetric[];
  title?: string;
}

export function BandwidthChart({ data, title = 'Bandwidth Utilization' }: BandwidthChartProps) {
  return (
    <div className="w-full h-[300px]">
      <h3 className="text-lg font-semibold mb-4">{title}</h3>
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="time" tick={{ fontSize: 12 }} />
          <YAxis tick={{ fontSize: 12 }} />
          <Tooltip />
          <Area type="monotone" dataKey="upload" stackId="1" stroke="#8884d8" fill="#8884d8" fillOpacity={0.6} />
          <Area type="monotone" dataKey="download" stackId="1" stroke="#82ca9d" fill="#82ca9d" fillOpacity={0.6} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
```

### Network Outage Map

```typescript
// src/components/broadband/outage-map.tsx
import { MapContainer, TileLayer, CircleMarker, Popup } from 'react-leaflet';

interface OutageMapProps {
  outages: NetworkOutage[];
  oltLocations: OLTStatus[];
}

export function OutageMap({ outages, oltLocations }: OutageMapProps) {
  return (
    <MapContainer center={[20.5937, 78.9629]} zoom={5} className="h-[400px] w-full rounded-lg">
      <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
      {oltLocations.map((olt) => (
        <CircleMarker
          key={olt.id}
          center={[olt.location.lat, olt.location.lng]}
          radius={8}
          color={olt.status === 'online' ? 'green' : olt.status === 'degraded' ? 'orange' : 'red'}
        >
          <Popup>
            <strong>{olt.name}</strong>
            <br />Status: {olt.status}
            <br />ONUs: {olt.active_onus}/{olt.total_onus}
          </Popup>
        </CircleMarker>
      ))}
    </MapContainer>
  );
}
```

---

## 4. ERP Intelligence Dashboard

### Features

| Feature | Description |
|---------|-------------|
| Inventory levels | Real-time stock with AI-predicted restock |
| Purchase orders | Automated PO generation |
| Finance reports | Revenue, expenses, profit charts |
| Supplier scores | Performance ratings and rankings |

### Dashboard Layout

```
┌──────────────────────────────────────────────────────┐
│  ERP Intelligence                    Last sync: 2m ago│
├───────────┬───────────┬──────────────┬───────────────┤
│  Revenue  │  Expenses │  Profit      │  Inventory    │
│  ₹45.2L   │  ₹32.1L   │  ₹13.1L     │  1,247 SKUs  │
│  ↑ 12%    │  ↑ 8%     │  ↑ 18%      │  ⚠ 23 Low    │
├───────────┴───────────┴──────────────┴───────────────┤
│                                                       │
│  ┌─────────────────────┐  ┌────────────────────────┐ │
│  │ Revenue Trend        │  │ AI Restock Predictions │ │
│  │ ▁▂▃▄▅▆▇█▇▆          │  │                        │ │
│  │ [Line Chart]         │  │ Item A: 3 days left    │ │
│  │                      │  │ Item B: 7 days left    │ │
│  │                      │  │ Item C: 12 days left   │ │
│  └─────────────────────┘  └────────────────────────┘ │
│                                                       │
│  ┌─────────────────────────────────────────────────┐ │
│  │ Recent Purchase Orders                          │ │
│  │ PO-001  | Supplier A  | ₹2.5L  | Status: Sent  │ │
│  │ PO-002  | Supplier B  | ₹1.8L  | Status: Recv'd │ │
│  │ PO-003  | Supplier C  | ₹3.2L  | Status: Pnding │ │
│  └─────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────┘
```

### ERP Data Model

```typescript
interface ERPDashboard {
  finance: FinanceSummary;
  inventory: InventorySummary;
  purchase_orders: PurchaseOrder[];
  supplier_scores: SupplierScore[];
  ai_insights: AIInsight[];
}

interface FinanceSummary {
  revenue: number;
  expenses: number;
  profit: number;
  revenue_trend: TrendPoint[];
  expenses_trend: TrendPoint[];
}

interface TrendPoint {
  date: string;
  value: number;
}

interface InventorySummary {
  total_skus: number;
  low_stock_count: number;
  out_of_stock_count: number;
  total_value: number;
  restock_predictions: RestockPrediction[];
}

interface RestockPrediction {
  item_id: string;
  item_name: string;
  current_stock: number;
  daily_usage: number;
  days_until_stockout: number;
  suggested_reorder_qty: number;
  confidence: number;
}

interface SupplierScore {
  supplier_id: string;
  name: string;
  delivery_score: number;
  quality_score: number;
  price_score: number;
  overall_score: number;
  trend: 'improving' | 'stable' | 'declining';
}
```

### AI Restock Hook

```typescript
// src/hooks/use-erp-dashboard.ts
import { useQuery } from '@tanstack/react-query';
import { dashboardApi } from '@/lib/api';

export function useERPDashboard() {
  return useQuery({
    queryKey: ['dashboard', 'erp'],
    queryFn: async () => {
      const { data } = await dashboardApi.getERPMetrics();
      return data as ERPDashboard;
    },
    refetchInterval: 5 * 60 * 1000, // Refresh every 5 min
  });
}

export function useRestockPredictions() {
  return useQuery({
    queryKey: ['erp', 'restock-predictions'],
    queryFn: async () => {
      const { data } = await apiClient.post('/api/v1/erp/restock-predictions', {});
      return data.predictions as RestockPrediction[];
    },
    refetchInterval: 15 * 60 * 1000, // Refresh every 15 min
  });
}
```

---

## 5. CRM Intelligence Dashboard

### Features

| Feature | Description |
|---------|-------------|
| Lead scoring | AI-calculated lead quality scores |
| Sales pipeline | Visual funnel chart |
| Engagement metrics | Email opens, clicks, meetings |
| Churn prediction | At-risk customer alerts |

### Sales Pipeline Funnel

```typescript
// src/components/crm/pipeline-funnel.tsx
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';

interface PipelineStage {
  name: string;
  count: number;
  value: number;
}

const STAGE_COLORS = ['#6366f1', '#8b5cf6', '#a78bfa', '#c4b5fd', '#ddd6fe'];

export function PipelineFunnel({ stages }: { stages: PipelineStage[] }) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <BarChart data={stages} layout="vertical">
        <XAxis type="number" />
        <YAxis type="category" dataKey="name" width={120} />
        <Tooltip formatter={(value) => [`₹${value.toLocaleString()}`, 'Value']} />
        <Bar dataKey="value" radius={[0, 4, 4, 0]}>
          {stages.map((_, i) => (
            <Cell key={i} fill={STAGE_COLORS[i]} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}
```

### Churn Prediction Alert

```typescript
// src/components/crm/churn-alerts.tsx
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { AlertTriangle } from 'lucide-react';

interface ChurnAlert {
  customer_id: string;
  customer_name: string;
  risk_score: number;
  factors: string[];
  recommended_action: string;
}

export function ChurnAlerts({ alerts }: { alerts: ChurnAlert[] }) {
  return (
    <div className="space-y-3">
      <h3 className="text-lg font-semibold">Churn Prediction Alerts</h3>
      {alerts.map((alert) => (
        <Alert key={alert.customer_id} variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>{alert.customer_name} — Risk: {alert.risk_score}%</AlertTitle>
          <AlertDescription>
            <p className="mt-1">Factors: {alert.factors.join(', ')}</p>
            <p className="mt-1 font-medium">Action: {alert.recommended_action}</p>
          </AlertDescription>
        </Alert>
      ))}
    </div>
  );
}
```

---

## 6. Billing Intelligence Dashboard

### Dashboard Layout

```
┌──────────────────────────────────────────────────────┐
│  Billing Intelligence               Period: Jun 2024  │
├───────────┬───────────┬──────────────┬───────────────┤
│  Total    │  Collected│  Outstanding │  Collection   │
│  Invoiced │           │              │  Rate         │
│  ₹82.3L   │  ₹68.5L   │  ₹13.8L     │  83.2%       │
│  ↑ 15%    │  ↑ 12%    │  ↑ 22% ⚠   │  ↓ 2.1%      │
├───────────┴───────────┴──────────────┴───────────────┤
│                                                       │
│  ┌─────────────────────┐  ┌────────────────────────┐ │
│  │ Revenue Trend        │  │ Payment Methods         │ │
│  │ [Area Chart]         │  │ [Pie Chart]             │ │
│  │                      │  │ UPI: 45%               │ │
│  │                      │  │ Bank: 30%              │ │
│  │                      │  │ Card: 25%              │ │
│  └─────────────────────┘  └────────────────────────┘ │
│                                                       │
│  ⚠ Overdue Payments (15 invoices, ₹5.2L)             │
│  ┌─────────────────────────────────────────────────┐ │
│  │ INV-0042 | ₹45,000 | 15 days overdue | ⚠ High  │ │
│  │ INV-0038 | ₹28,500 | 30 days overdue | 🔴 Crit │ │
│  └─────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────┘
```

### Billing Data Model

```typescript
interface BillingDashboard {
  summary: {
    total_invoiced: number;
    total_collected: number;
    outstanding: number;
    collection_rate: number;
    overdue_count: number;
  };
  revenue_trend: TrendPoint[];
  payment_methods: { method: string; percentage: number }[];
  overdue_invoices: OverdueInvoice[];
}

interface OverdueInvoice {
  invoice_id: string;
  amount: number;
  days_overdue: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
  customer_name: string;
  last_reminder_sent: string;
}
```

---

## 7. Payment Monitoring Dashboard

### Transaction Volume Chart

```typescript
// src/components/payments/transaction-volume.tsx
import { ComposedChart, Bar, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface TransactionData {
  date: string;
  successful: number;
  failed: number;
  total_volume: number;
}

export function TransactionVolumeChart({ data }: { data: TransactionData[] }) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <ComposedChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="date" />
        <YAxis yAxisId="left" />
        <YAxis yAxisId="right" orientation="right" />
        <Tooltip />
        <Bar yAxisId="left" dataKey="successful" fill="#22c55e" stackId="1" />
        <Bar yAxisId="left" dataKey="failed" fill="#ef4444" stackId="1" />
        <Line yAxisId="right" dataKey="total_volume" stroke="#6366f1" strokeWidth={2} />
      </ComposedChart>
    </ResponsiveContainer>
  );
}
```

### Payment Dashboard Features

| Feature | Description |
|---------|-------------|
| Transaction volume | Daily/hourly transaction counts |
| Fraud detection | AI-flagged suspicious transactions |
| Success/failure rates | Real-time payment gateway health |
| Gateway status | Health checks for payment gateways |

---

## 8. Exchange/Blockchain Dashboard

### Market Price Charts

```typescript
// src/components/exchange/price-chart.tsx
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';

interface PriceData {
  timestamp: string;
  price: number;
  volume: number;
}

export function PriceChart({ data, symbol }: { data: PriceData[]; symbol: string }) {
  const currentPrice = data[data.length - 1]?.price;
  const prevPrice = data[data.length - 2]?.price;
  const change = prevPrice ? ((currentPrice - prevPrice) / prevPrice) * 100 : 0;

  return (
    <div className="w-full">
      <div className="flex items-center gap-4 mb-4">
        <h3 className="text-lg font-semibold">{symbol}</h3>
        <span className={change >= 0 ? 'text-green-500' : 'text-red-500'}>
          {change >= 0 ? '+' : ''}{change.toFixed(2)}%
        </span>
      </div>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="timestamp" />
          <YAxis domain={['auto', 'auto']} />
          <Tooltip />
          <ReferenceLine y={currentPrice} stroke="#6366f1" strokeDasharray="3 3" />
          <Line type="monotone" dataKey="price" stroke="#6366f1" dot={false} strokeWidth={2} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
```

### Blockchain Dashboard Features

| Feature | Description |
|---------|-------------|
| Price charts | Real-time crypto/asset prices |
| Portfolio overview | Total value, allocation, P&L |
| AML monitoring | Suspicious transaction alerts |
| Validator status | Staking node health |

---

## 9. Credit (Cibil) Intelligence Dashboard

### Credit Score Distribution

```typescript
// src/components/cibil/score-distribution.tsx
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';

interface ScoreRange {
  range: string;
  count: number;
  color: string;
}

export function ScoreDistribution({ data }: { data: ScoreRange[] }) {
  return (
    <ResponsiveContainer width="100%" height={250}>
      <BarChart data={data}>
        <XAxis dataKey="range" />
        <YAxis />
        <Tooltip />
        <Bar dataKey="count" radius={[4, 4, 0, 0]}>
          {data.map((entry, i) => (
            <Cell key={i} fill={entry.color} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}
```

### AI Report Explanation

```typescript
// src/components/cibil/report-explanation.tsx
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Bot } from 'lucide-react';

interface ReportExplanationProps {
  explanation: string;
  factors: { factor: string; impact: 'positive' | 'negative' | 'neutral'; detail: string }[];
}

export function ReportExplanation({ explanation, factors }: ReportExplanationProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bot className="h-5 w-5" /> AI Credit Analysis
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground mb-4">{explanation}</p>
        <div className="space-y-2">
          {factors.map((f, i) => (
            <div key={i} className="flex items-start gap-2">
              <span className={f.impact === 'positive' ? 'text-green-500' : f.impact === 'negative' ? 'text-red-500' : 'text-gray-500'}>
                {f.impact === 'positive' ? '▲' : f.impact === 'negative' ? '▼' : '●'}
              </span>
              <div>
                <span className="font-medium">{f.factor}</span>
                <p className="text-xs text-muted-foreground">{f.detail}</p>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
```

---

## 10. HR Intelligence Dashboard

### Recruitment Pipeline

```
┌──────────────────────────────────────────────────┐
│  HR Intelligence Dashboard                        │
├──────────┬──────────┬──────────┬─────────────────┤
│  Total   │  Active  │  Avg Time│  Attrition      │
│  Employees│ Openings│ to Hire  │  Prediction     │
│  245     │  12      │  23 days │  ⚠ 8% risk      │
├──────────┴──────────┴──────────┴─────────────────┤
│                                                   │
│  Recruitment Pipeline                             │
│  Applied: 156 → Screened: 89 → Interview: 34     │
│         → Offer: 12 → Hired: 8                   │
│                                                   │
│  Attrition Prediction by Department               │
│  Engineering: 5%  |  Sales: 12%  |  Support: 15% │
│                                                   │
│  Top Attrition Factors (AI Analysis)              │
│  1. Compensation below market (42%)               │
│  2. Limited growth opportunities (28%)             │
│  3. Work-life balance concerns (18%)              │
└──────────────────────────────────────────────────┘
```

---

## 11. Solar/Energy Intelligence Dashboard

### Energy Production Chart

```typescript
// src/components/solar/production-chart.tsx
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';

interface ProductionData {
  time: string;
  actual: number;
  predicted: number;
  weather_factor: number;
}

export function ProductionChart({ data }: { data: ProductionData[] }) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <AreaChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="time" />
        <YAxis />
        <Tooltip />
        <Area type="monotone" dataKey="predicted" stroke="#94a3b8" fill="transparent" strokeDasharray="5 5" />
        <Area type="monotone" dataKey="actual" stroke="#22c55e" fill="#22c55e" fillOpacity={0.3} />
      </AreaChart>
    </ResponsiveContainer>
  );
}
```

### Solar Dashboard Features

| Feature | Description |
|---------|-------------|
| Energy production | Real-time generation output |
| Weather correlation | Solar irradiance vs output |
| Maintenance scheduling | Panel cleaning, inverter checks |
| ROI calculations | Investment return projections |
| Grid feed-in | Export to grid monitoring |

---

## 12. Dashboard API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/dashboard/stats` | POST | Central dashboard stats |
| `/api/v1/dashboard/broadband` | POST | Broadband network metrics |
| `/api/v1/dashboard/erp` | POST | ERP intelligence data |
| `/api/v1/dashboard/crm` | POST | CRM intelligence data |
| `/api/v1/dashboard/billing` | POST | Billing metrics |
| `/api/v1/dashboard/payments` | POST | Payment monitoring data |
| `/api/v1/dashboard/exchange` | POST | Exchange/blockchain data |
| `/api/v1/dashboard/cibil` | POST | Credit intelligence data |
| `/api/v1/dashboard/hr` | POST | HR intelligence data |
| `/api/v1/dashboard/solar` | POST | Solar/energy data |
| `/api/v1/dashboard/ai-insights` | POST | AI-generated insights |

---

## 13. Dashboard Hooks and Stores

```typescript
// src/hooks/use-dashboard.ts
import { useQuery } from '@tanstack/react-query';
import { dashboardApi } from '@/lib/api';

export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: async () => {
      const { data } = await dashboardApi.getStats();
      return data;
    },
    refetchInterval: 30000, // 30 seconds
  });
}

export function useBroadbandDashboard() {
  return useQuery({
    queryKey: ['dashboard', 'broadband'],
    queryFn: async () => {
      const { data } = await dashboardApi.getBroadbandMetrics();
      return data;
    },
    refetchInterval: 10000, // 10 seconds for real-time
  });
}

// Dashboard store for filters and preferences
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface DashboardState {
  selectedTimeRange: 'today' | 'week' | 'month' | 'quarter' | 'year';
  selectedDashboard: string;
  isAutoRefresh: boolean;
  refreshInterval: number;
  setTimeRange: (range: DashboardState['selectedTimeRange']) => void;
  setSelectedDashboard: (dashboard: string) => void;
  setAutoRefresh: (enabled: boolean) => void;
  setRefreshInterval: (ms: number) => void;
}

export const useDashboardStore = create<DashboardState>()(
  persist(
    (set) => ({
      selectedTimeRange: 'today',
      selectedDashboard: 'central',
      isAutoRefresh: true,
      refreshInterval: 30000,
      setTimeRange: (range) => set({ selectedTimeRange: range }),
      setSelectedDashboard: (dashboard) => set({ selectedDashboard: dashboard }),
      setAutoRefresh: (enabled) => set({ isAutoRefresh: enabled }),
      setRefreshInterval: (ms) => set({ refreshInterval: ms }),
    }),
    { name: 'dashboard-preferences' }
  )
);
```

---

## 14. Dashboard Component Architecture

| Component | Purpose | Used In |
|-----------|---------|---------|
| `StatCard` | KPI display card | All dashboards |
| `ChartCard` | Chart container with title | All dashboards |
| `DataTable` | Sortable/filterable table | All dashboards |
| `AlertBanner` | AI-generated alerts | All dashboards |
| `TimeRangeSelector` | Period picker | All dashboards |
| `RefreshToggle` | Auto-refresh control | All dashboards |
| `ExportButton` | Export dashboard data | All dashboards |
| `AIInsightPanel` | AI analysis display | All dashboards |
| `StatusBadge` | Status indicator | Broadband, Billing |
| `FunnelChart` | Pipeline visualization | CRM |
| `GaugeChart` | Score display | Credit, HR |
| `HeatMap` | Geographic data | Broadband, Solar |
| `DonutChart` | Allocation display | Exchange, Billing |

---

## 15. Real-Time Data Updates

```typescript
// src/hooks/use-realtime-dashboard.ts
import { useEffect } from 'react';
import { wsClient } from '@/lib/websocket-client';
import { useQueryClient } from '@tanstack/react-query';

export function useRealtimeDashboard(dashboardType: string) {
  const queryClient = useQueryClient();

  useEffect(() => {
    const unsub = wsClient.on('dashboard_update', (msg) => {
      if (msg.dashboard === dashboardType) {
        queryClient.invalidateQueries({ queryKey: ['dashboard', dashboardType] });
      }
    });

    return unsub;
  }, [dashboardType, queryClient]);
}
```
