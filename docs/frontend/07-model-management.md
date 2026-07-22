# Model Management

## Table of Contents

- [Overview](#overview)
- [Model List Page](#model-list-page)
- [Model Card Component](#model-card-component)
- [Model Detail Page](#model-detail-page)
- [Model Download](#model-download)
- [Model Removal](#model-removal)
- [GPU Usage Visualization](#gpu-usage-visualization)
- [Model Performance Charts](#model-performance-charts)
- [Model Routing Configuration](#model-routing-configuration)
- [Model Priority Settings](#model-priority-settings)
- [Model Status Monitoring](#model-status-monitoring)
- [Model Health Checks](#model-health-checks)
- [Model Configuration](#model-configuration)
- [Model Comparison](#model-comparison)
- [Model Usage Analytics](#model-usage-analytics)
- [Model Cost Tracking](#model-cost-tracking)
- [API Integration](#api-integration)
- [Hooks](#hooks)
- [Stores](#stores)
- [Error Handling](#error-handling)
- [Responsive Design](#responsive-design)

---

## Overview

Model Management provides centralized control over AI models deployed in the system. It covers model deployment, monitoring, configuration, routing, and performance tracking across GPU resources.

```
┌──────────────────────────────────────────────────────────────┐
│  Model Management                                            │
├──────────────┬───────────────────────────────────────────────┤
│              │                                               │
│  Dashboard   │  ┌─────────┐ ┌─────────┐ ┌─────────┐        │
│  Models      │  │ Llama 3 │ │ GPT-4o  │ │ Gemma 2 │        │
│  GPU         │  │ ● Run   │ │ ● Run   │ │ ○ Stop  │        │
│  Analytics   │  │ 8GB/24GB│ │ API     │ │ 4GB/24GB│        │
│  Settings    │  └─────────┘ └─────────┘ └─────────┘        │
│              │                                               │
│              │  ┌─────────────────────────────────────┐     │
│              │  │  GPU Utilization: 58% (42.5GB/72GB) │     │
│              │  │  ████████████████░░░░░░░░░░░░░░░░░░ │     │
│              │  └─────────────────────────────────────┘     │
│              │                                               │
│              │  Model Performance Overview                   │
│              │  ┌─────────────────────────────────────┐     │
│              │  │  Avg Latency: 120ms  Throughput: 45 │     │
│              │  │  Requests/hr: 1,234   Errors: 0.2%  │     │
│              │  └─────────────────────────────────────┘     │
└──────────────┴───────────────────────────────────────────────┘
```

---

## Model List Page

### View Toggle

| View    | Layout         | Best For                          |
|---------|----------------|-----------------------------------|
| Grid    | Card-based     | Visual overview, many models      |
| Table   | Row-based      | Detailed comparison, sorting      |

### Grid View

```
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│  🧠 Llama 3 70B  │ │  🧠 GPT-4o       │ │  🧠 Gemma 2 27B  │
│  ● Running       │ │  ● Running       │ │  ○ Stopped       │
│                  │ │                  │ │                  │
│  Type: Chat      │ │  Type: Chat      │ │  Type: Chat      │
│  VRAM: 14GB      │ │  Type: API       │ │  VRAM: 8GB       │
│  Requests: 450/d │ │  Requests: 890/d │ │  Requests: 0/d   │
│                  │ │                  │ │                  │
│  [View Details]  │ │  [View Details]  │ │  [Start] [View]  │
└──────────────────┘ └──────────────────┘ └──────────────────┘
```

### Table View

| Name       | Type     | Status    | VRAM/Quota | Requests/day | Avg Latency | Errors | Actions       |
|------------|----------|-----------|------------|--------------|-------------|--------|---------------|
| Llama 3 70B| Local    | Running   | 14GB/24GB  | 450          | 120ms       | 0.1%   | Edit Stop Log |
| GPT-4o     | API      | Running   | API        | 890          | 85ms        | 0.0%   | Edit Stop Log |
| Gemma 2 27B| Local    | Stopped   | 0GB/24GB   | 0            | —           | —      | Start Edit    |
| Mistral 7B | Local    | Error     | 4GB/24GB   | 12           | 200ms       | 5.2%   | Restart Edit  |

### Filter Options

```tsx
// components/models/ModelFilters.tsx
export function ModelFilters() {
  return (
    <div className="flex flex-wrap gap-3">
      <Select placeholder="Status">
        <SelectItem value="all">All Status</SelectItem>
        <SelectItem value="running">Running</SelectItem>
        <SelectItem value="stopped">Stopped</SelectItem>
        <SelectItem value="loading">Loading</SelectItem>
        <SelectItem value="error">Error</SelectItem>
      </Select>

      <Select placeholder="Type">
        <SelectItem value="all">All Types</SelectItem>
        <SelectItem value="local">Local</SelectItem>
        <SelectItem value="api">API</SelectItem>
        <SelectItem value="hybrid">Hybrid</SelectItem>
      </Select>

      <Select placeholder="Task">
        <SelectItem value="all">All Tasks</SelectItem>
        <SelectItem value="chat">Chat</SelectItem>
        <SelectItem value="completion">Completion</SelectItem>
        <SelectItem value="embedding">Embedding</SelectItem>
        <SelectItem value="vision">Vision</SelectItem>
      </Select>

      <Select placeholder="Sort">
        <SelectItem value="name">Name</SelectItem>
        <SelectItem value="status">Status</SelectItem>
        <SelectItem value="vram">VRAM Usage</SelectItem>
        <SelectItem value="requests">Request Count</SelectItem>
        <SelectItem value="latency">Avg Latency</SelectItem>
      </Select>
    </div>
  );
}
```

---

## Model Card Component

```
┌──────────────────────────────────────────┐
│  🧠 Llama 3 70B Instruct                │
│  ● Running                               │
├──────────────────────────────────────────┤
│  Type: Local    Task: Chat               │
│  Format: GGUF   Quantization: Q4_K_M     │
│                                          │
│  VRAM Usage                             │
│  ████████████░░░░░░░░  14GB / 24GB      │
│                                          │
│  Performance (24h)                       │
│  Requests: 450/day                       │
│  Avg Latency: 120ms                      │
│  P99 Latency: 340ms                      │
│  Tokens/s: 42                            │
│  Error Rate: 0.1%                        │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │  Latency (last 24h)               │  │
│  │  ▁▂▃▂▁▂▃▄▃▂▁▂▃▄▃▂▁▂▃▄▃▂▁▂▃▄▃▂ │  │
│  └────────────────────────────────────┘  │
│                                          │
│  [View Details] [Configure] [Stop]       │
└──────────────────────────────────────────┘
```

### Model Card Code

```tsx
// components/models/ModelCard.tsx
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { MiniChart } from '@/components/ui/mini-chart';
import { ModelActions } from '@/components/models/ModelActions';

interface ModelCardProps {
  model: Model;
  onClick: () => void;
}

export function ModelCard({ model, onClick }: ModelCardProps) {
  const vramPercent = model.vram_used && model.vram_total
    ? Math.round((model.vram_used / model.vram_total) * 100)
    : null;

  return (
    <div
      className="border rounded-xl p-4 hover:shadow-md transition-shadow cursor-pointer"
      onClick={onClick}
      role="article"
      aria-label={`Model ${model.name}`}
    >
      <div className="flex items-start justify-between">
        <div>
          <h3 className="font-semibold text-lg">{model.name}</h3>
          <StatusIndicator status={model.status} />
        </div>
        <Badge variant={model.type === 'api' ? 'secondary' : 'default'}>
          {model.type}
        </Badge>
      </div>

      <div className="mt-3 space-y-2 text-sm text-muted-foreground">
        <p>Task: {model.task}</p>
        {model.format && <p>Format: {model.format} ({model.quantization})</p>}
      </div>

      {vramPercent !== null && (
        <div className="mt-3">
          <p className="text-xs text-muted-foreground mb-1">
            VRAM: {model.vram_used}GB / {model.vram_total}GB
          </p>
          <Progress value={vramPercent} className="h-1.5" />
        </div>
      )}

      <div className="mt-3 grid grid-cols-2 gap-2 text-xs">
        <div>Requests: {model.requests_24h}/day</div>
        <div>Latency: {model.avg_latency_ms}ms</div>
        <div>Tokens/s: {model.tokens_per_second ?? '—'}</div>
        <div>Errors: {model.error_rate ?? '—'}%</div>
      </div>

      {model.latency_chart && (
        <div className="mt-3">
          <MiniChart data={model.latency_chart} height={32} color="hsl(var(--primary))" />
        </div>
      )}

      <div className="mt-4 flex gap-2" onClick={(e) => e.stopPropagation()}>
        <ModelActions model={model} />
      </div>
    </div>
  );
}
```

---

## Model Detail Page

### Detail Layout

```
┌─────────────────────────────────────────────────────────────┐
│  ← Back to Models                                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  🧠 Llama 3 70B Instruct                ● Running          │
│  Type: Local  |  Task: Chat  |  Format: GGUF Q4_K_M        │
│                                                             │
│  [Overview] [Performance] [Configuration] [Logs] [Routing]  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Overview Tab:                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  GPU Allocation: GPU 0 (14GB used of 24GB)         │   │
│  │  Loaded: Jul 15, 2026 08:30:00                      │   │
│  │  Uptime: 25h 30m                                    │   │
│  │  Total Requests: 11,250                             │   │
│  │  Total Tokens: 2,847,392                            │   │
│  │                                                     │   │
│  │  Last 24h Metrics:                                  │   │
│  │  ┌──────────┬──────────┬──────────┬──────────┐     │   │
│  │  │ Requests │ Latency  │ Tokens/s │ Errors   │     │   │
│  │  │   450    │  120ms   │   42     │  0.1%    │     │   │
│  │  └──────────┴──────────┴──────────┴──────────┘     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Model Download

### Available Models

```
┌──────────────────────────────────────────────────────────────┐
│  Download Models                                             │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Search: [_________________________]                         │
│  Filter: [All] [Chat] [Embedding] [Vision] [Code]          │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Llama 3 8B Instruct                                  │  │
│  │  Size: 4.5GB  |  Format: GGUF Q4_K_M                  │  │
│  │  Context: 8K  |  Task: Chat                            │  │
│  │  Downloads: 12,450  |  Rating: ★★★★☆                  │  │
│  │  [Download] [Preview]                                  │  │
│  ├────────────────────────────────────────────────────────┤  │
│  │  Gemma 2 9B                                           │  │
│  │  Size: 5.8GB  |  Format: GGUF Q5_K_M                  │  │
│  │  Context: 8K  |  Task: Chat                            │  │
│  │  Downloads: 8,230  |  Rating: ★★★★☆                  │  │
│  │  [Download] [Preview]                                  │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Download Progress

```
┌──────────────────────────────────────────────────────────────┐
│  Downloading: Llama 3 70B Instruct                           │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ████████████████████████░░░░░░░░░░░░░░░░░░  58%            │
│                                                              │
│  Downloaded: 8.2GB / 14.1GB                                 │
│  Speed: 45 MB/s                                              │
│  ETA: 2 min 15 sec                                           │
│                                                              │
│  Verifying checksum...  ✅                                   │
│  Extracting...        ⏳ In progress                         │
│  Loading into GPU...   ⏳ Waiting                            │
│                                                              │
│  [Cancel Download]                                           │
└──────────────────────────────────────────────────────────────┘
```

### Download Pipeline

```
Download       Verify         Extract        Load into       Ready
   │              │              │            GPU              │
   ▼              ▼              ▼              ▼              ▼
┌────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐
│ Fetching│─▶│Checksum  │─▶│ Extract  │─▶│ GPU      │─▶│ Model  │
│ file    │  │ verify   │  │ archive  │  │ memory   │  │ ready  │
└────────┘  └──────────┘  └──────────┘  └──────────┘  └────────┘
   Status:      Status:       Status:       Status:      Status:
   downloading  verifying     extracting    loading      ready
```

---

## Model Removal

### Confirmation Dialog

```tsx
// components/models/RemoveModelDialog.tsx
import { AlertDialog, AlertDialogAction, AlertDialogCancel,
         AlertDialogContent, AlertDialogDescription,
         AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog';

interface RemoveModelDialogProps {
  model: Model;
  onConfirm: () => Promise<void>;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function RemoveModelDialog({ model, onConfirm, open, onOpenChange }: RemoveModelDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove Model</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to remove <strong>{model.name}</strong>?
            {model.status === 'running' && (
              <span className="block mt-2 text-destructive font-medium">
                This model is currently running and will be stopped first.
              </span>
            )}
            <span className="block mt-2">
              This will free {model.vram_used || model.disk_size}GB of {model.type === 'local' ? 'disk' : ''} space.
              This action cannot be undone.
            </span>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            className="bg-destructive text-destructive-foreground"
          >
            Remove Model
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
```

### Removal Cleanup Steps

```
User confirms removal
       │
       ▼
┌──────────────┐
│ Stop model   │ (if running)
│ gracefully   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Unload from  │
│ GPU memory   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Delete model │
│ files from   │
│ disk         │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Remove       │
│ routing      │
│ config       │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Update model │
│ list in DB   │
└──────────────┘
```

---

## GPU Usage Visualization

### GPU Dashboard

```
┌──────────────────────────────────────────────────────────────┐
│  GPU Overview                                                 │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  GPU 0: NVIDIA RTX 4090 (24GB)           Utilization: 82%   │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  VRAM: ██████████████████████░░░░  20GB / 24GB        │  │
│  │  Util: ████████████████░░░░░░░░░░  82%                │  │
│  │  Temp: ████████████████████░░░░░░  72°C               │  │
│  │  Power: ████████████████████░░░░░  280W / 350W        │  │
│  │                                                        │  │
│  │  Models loaded:                                        │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ Llama 3 70B  ████████████████  14GB (58%)      │  │  │
│  │  │ Embedding    ████████          4GB  (17%)       │  │  │
│  │  │ Free         ████████          6GB  (25%)       │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  GPU 1: NVIDIA RTX 4090 (24GB)           Utilization: 35%   │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  VRAM: ████████░░░░░░░░░░░░░░░░  8GB / 24GB          │  │
│  │  Util: ███████░░░░░░░░░░░░░░░░░  35%                  │  │
│  │  Temp: ██████████████░░░░░░░░░░  54°C                 │  │
│  │  Power: ██████████░░░░░░░░░░░░░  150W / 350W         │  │
│  │                                                        │  │
│  │  Models loaded:                                        │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ Gemma 2 27B  ████████          8GB  (33%)       │  │  │
│  │  │ Free         ████████████████  16GB (67%)       │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Per-Model GPU Allocation Table

| Model        | GPU    | VRAM Used | VRAM % | Utilization | Temp | Power |
|--------------|--------|-----------|--------|-------------|------|-------|
| Llama 3 70B  | GPU 0  | 14GB      | 58%    | 78%         | 70°C | 260W  |
| Embedding    | GPU 0  | 4GB       | 17%    | 45%         | 55°C | 80W   |
| Gemma 2 27B  | GPU 1  | 8GB       | 33%    | 32%         | 52°C | 120W  |
| Free         | GPU 0  | 6GB       | 25%    | —           | —    | —     |
| Free         | GPU 1  | 16GB      | 67%    | —           | —    | —     |

---

## Model Performance Charts

### Latency Chart

```
Response Latency (Last 24 Hours)
350ms ┤
      │                        ╷
300ms ┤                        │
      │              ╷         │
250ms ┤              │    ╷    │
      │    ╷    ╷    │    │    │
200ms ┤    │    │    │    │    │
      │    │    │    │    │    │    ╷
150ms ┤────┼────┼────┼────┼────┼────┼─────────
      │    ╵    ╵    ╵    ╵    ╵    ╵
100ms ┤
      │
 50ms ┤
      └──────────────────────────────────────────
       00:00  04:00  08:00  12:00  16:00  20:00

  ── P50    ── P95    ── P99
```

### Throughput Chart

```
Throughput (Tokens/Second)
60 ┤
   │              ┌─┐
50 ┤         ┌─┐  │ │        ┌─┐
   │    ┌─┐  │ │  │ │  ┌─┐  │ │
40 ┤    │ │  │ │  │ │  │ │  │ │
   │ ┌─┐│ │  │ │  │ │  │ │  │ │  ┌─┐
30 ┤ │ ││ │  │ │  │ │  │ │  │ │  │ │
   │ │ ││ │  │ │  │ │  │ │  │ │  │ │
20 ┤ │ ││ │  │ │  │ │  │ │  │ │  │ │
   │ │ ││ │  │ │  │ │  │ │  │ │  │ │
10 ┤ │ ││ │  │ │  │ │  │ │  │ │  │ │
   │ │ ││ │  │ │  │ │  │ │  │ │  │ │
 0 ┼─┴─┴┴─┴──┴─┴──┴─┴──┴─┴──┴─┴──┴─┴──
   00  02  04  06  08  10  12  14  16  18
```

### Performance Metrics

| Metric              | Description                          | Unit   |
|---------------------|--------------------------------------|--------|
| P50 Latency        | Median response time                 | ms     |
| P95 Latency        | 95th percentile latency              | ms     |
| P99 Latency        | 99th percentile latency              | ms     |
| Throughput         | Tokens generated per second          | tok/s  |
| Requests/sec       | Concurrent request handling          | req/s  |
| Time to First Token| Latency for first token              | ms     |
| Error Rate         | Failed request percentage            | %      |
| Queue Depth        | Pending requests in queue            | count  |

---

## Model Routing Configuration

### Routing Rules Table

```
┌──────────────────────────────────────────────────────────────┐
│  Model Routing Rules                                         │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Task              Primary Model      Fallback    Priority  │
│  ────────────────────────────────────────────────────────── │
│  Chat              Llama 3 70B        GPT-4o      1         │
│  Code Generation   CodeLlama 34B      GPT-4o      1         │
│  Embedding         text-embedding-3   —           —         │
│  Vision            GPT-4o Vision      LLaVA 13B   2         │
│  Summarization     Llama 3 70B        Mistral 7B  2         │
│  Translation       GPT-4o             Llama 3 70B 2         │
│  SQL Generation    CodeLlama 34B      GPT-4o      1         │
│                                                              │
│  [+ Add Rule]  [Edit]  [Reset to Defaults]                  │
└──────────────────────────────────────────────────────────────┘
```

### Routing Edit Dialog

```tsx
// components/models/RoutingRuleEditor.tsx
interface RoutingRuleEditorProps {
  rule?: RoutingRule;
  availableModels: Model[];
  onSave: (rule: RoutingRule) => Promise<void>;
  onCancel: () => void;
}

export function RoutingRuleEditor({ rule, availableModels, onSave, onCancel }: RoutingRuleEditorProps) {
  const [form, setForm] = useState<RoutingRule>(
    rule ?? { task: '', primary_model: '', fallback_model: '', priority: 1 }
  );

  return (
    <Dialog>
      <DialogHeader>
        <DialogTitle>{rule ? 'Edit' : 'Add'} Routing Rule</DialogTitle>
      </DialogHeader>
      <div className="space-y-4">
        <div>
          <Label>Task Type</Label>
          <Select value={form.task} onValueChange={(v) => setForm({ ...form, task: v })}>
            <SelectContent>
              <SelectItem value="chat">Chat</SelectItem>
              <SelectItem value="completion">Completion</SelectItem>
              <SelectItem value="embedding">Embedding</SelectItem>
              <SelectItem value="vision">Vision</SelectItem>
              <SelectItem value="code">Code Generation</SelectItem>
              <SelectItem value="sql">SQL Generation</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label>Primary Model</Label>
          <Select value={form.primary_model}
            onValueChange={(v) => setForm({ ...form, primary_model: v })}>
            <SelectContent>
              {availableModels.map((m) => (
                <SelectItem key={m.id} value={m.id}>
                  {m.name} ({m.status})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label>Fallback Model (optional)</Label>
          <Select value={form.fallback_model || ''}
            onValueChange={(v) => setForm({ ...form, fallback_model: v || undefined })}>
            <SelectContent>
              <SelectItem value="">None</SelectItem>
              {availableModels.filter((m) => m.id !== form.primary_model).map((m) => (
                <SelectItem key={m.id} value={m.id}>{m.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label>Priority (lower = higher priority)</Label>
          <Input type="number" value={form.priority}
            onChange={(e) => setForm({ ...form, priority: +e.target.value })} />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>Cancel</Button>
          <Button onClick={() => onSave(form)}>Save Rule</Button>
        </DialogFooter>
      </div>
    </Dialog>
  );
}
```

---

## Model Priority Settings

### Priority and Load Balancing

| Strategy       | Description                                    | Use Case           |
|----------------|------------------------------------------------|--------------------|
| Priority-based | Use primary, fallback to secondary on failure  | Critical tasks     |
| Round-robin    | Distribute requests equally across models      | High availability  |
| Least-loaded   | Route to model with lowest queue depth         | Performance        |
| Cost-optimized | Route to cheapest model that meets requirements| Cost control       |
| Latency-based  | Route to fastest-responding model              | User experience    |

### Configuration Panel

```
┌──────────────────────────────────────────────────┐
│  Load Balancing Strategy                         │
├──────────────────────────────────────────────────┤
│                                                  │
│  Strategy: (●) Priority-based                    │
│            (○) Round-robin                       │
│            (○) Least-loaded                      │
│            (○) Cost-optimized                    │
│            (○) Latency-based                     │
│                                                  │
│  Health Check Interval: [30] seconds             │
│  Max Retries:           [3]                      │
│  Retry Delay:           [1000] ms                │
│  Circuit Breaker:       [✓] Enabled              │
│  Failure Threshold:     [5] consecutive failures  │
│  Recovery Timeout:      [60] seconds              │
│                                                  │
│                    [Save Configuration]           │
└──────────────────────────────────────────────────┘
```

---

## Model Status Monitoring

### Status States

| Status    | Indicator | Description                              |
|-----------|-----------|------------------------------------------|
| Running   | ● Green   | Model loaded and accepting requests      |
| Stopped   | ○ Gray    | Model unloaded from memory               |
| Loading   | ◐ Yellow  | Model being loaded into GPU memory       |
| Error     | ● Red     | Model encountered an error               |
| Unloading | ◑ Yellow  | Model being removed from GPU memory      |
| Queued    | ◐ Blue    | Model download/load queued               |

### Status Dashboard

```tsx
// components/models/ModelStatusDashboard.tsx
export function ModelStatusDashboard({ models }: { models: Model[] }) {
  const statusCounts = models.reduce((acc, m) => {
    acc[m.status] = (acc[m.status] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  return (
    <div className="grid grid-cols-5 gap-4">
      {Object.entries(statusCounts).map(([status, count]) => (
        <div key={status} className="border rounded-lg p-4 text-center">
          <StatusIcon status={status} className="w-6 h-6 mx-auto" />
          <p className="mt-2 text-2xl font-bold">{count}</p>
          <p className="text-sm text-muted-foreground capitalize">{status}</p>
        </div>
      ))}
      <div className="border rounded-lg p-4 text-center">
        <Box className="w-6 h-6 mx-auto text-muted-foreground" />
        <p className="mt-2 text-2xl font-bold">{models.length}</p>
        <p className="text-sm text-muted-foreground">Total</p>
      </div>
    </div>
  );
}
```

---

## Model Health Checks

### Health Check Configuration

```typescript
// lib/models/health.ts
interface HealthCheckConfig {
  interval_seconds: number;
  timeout_seconds: number;
  endpoints: HealthEndpoint[];
}

interface HealthEndpoint {
  model_id: string;
  check_type: 'inference' | 'status' | 'gpu_memory' | 'disk';
  expected_response?: string;
  max_latency_ms?: number;
}

interface HealthStatus {
  model_id: string;
  status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  last_check: string;
  latency_ms: number;
  details: {
    inference_works: boolean;
    gpu_memory_ok: boolean;
    disk_ok: boolean;
    error_message?: string;
  };
}
```

### Health Status Table

| Model         | Status    | Last Check       | Latency | Inference | GPU  | Disk |
|---------------|-----------|------------------|---------|-----------|------|------|
| Llama 3 70B   | Healthy   | 10s ago          | 45ms    | ✅        | ✅   | ✅   |
| GPT-4o        | Healthy   | 10s ago          | 12ms    | ✅        | N/A  | N/A  |
| Gemma 2 27B   | Healthy   | 10s ago          | 52ms    | ✅        | ✅   | ✅   |
| Mistral 7B    | Degraded  | 10s ago          | 890ms   | ⚠️ Slow   | ✅   | ✅   |
| CodeLlama     | Unhealthy | 10s ago          | timeout | ❌        | ❌   | ✅   |

---

## Model Configuration

### Configuration Panel

```
┌──────────────────────────────────────────────────────────────┐
│  Model Configuration: Llama 3 70B                            │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Generation Parameters:                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Temperature:      [0.7]      (0.0 - 2.0)          │    │
│  │  Top-P:            [0.9]      (0.0 - 1.0)          │    │
│  │  Top-K:            [50]       (1 - 100)            │    │
│  │  Max Tokens:       [4096]     (1 - 128000)         │    │
│  │  Repeat Penalty:   [1.1]      (1.0 - 2.0)          │    │
│  │  Min P:            [0.05]     (0.0 - 1.0)          │    │
│  │  Frequency Penalty: [0.0]     (0.0 - 2.0)          │    │
│  │  Presence Penalty:  [0.0]     (0.0 - 2.0)          │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  System Prompt:                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  You are a helpful assistant...                    │    │
│  │                                                     │    │
│  │                                                     │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  GPU Settings:                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  GPU Layers:         [-1] (all, or 0-80)            │    │
│  │  Context Length:     [8192]                         │    │
│  │  Batch Size:         [512]                          │    │
│  │  Thread Count:       [8]                            │    │
│  │  Rope Scaling:       [linear]                       │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  [Reset to Defaults]  [Apply]  [Save as Template]           │
└──────────────────────────────────────────────────────────────┘
```

### Parameter Reference

| Parameter        | Range      | Default | Description                          |
|------------------|------------|---------|--------------------------------------|
| Temperature      | 0.0 – 2.0 | 0.7     | Randomness in token selection         |
| Top-P            | 0.0 – 1.0 | 0.9     | Nucleus sampling threshold            |
| Top-K            | 1 – 100   | 50      | Number of top tokens to consider      |
| Max Tokens       | 1 – 128K  | 4096    | Maximum output length                 |
| Repeat Penalty   | 1.0 – 2.0 | 1.1     | Penalize repeated tokens              |
| Min P            | 0.0 – 1.0 | 0.05    | Minimum probability threshold         |
| Frequency Penalty| 0.0 – 2.0 | 0.0     | Penalize frequent tokens              |
| Presence Penalty | 0.0 – 2.0 | 0.0     | Penalize tokens that appeared         |
| GPU Layers       | -1 / 0-80 | -1      | Number of layers on GPU (-1 = all)    |
| Context Length    | 512 – 128K| 8192    | Maximum context window                |
| Batch Size       | 1 – 2048  | 512     | Processing batch size                 |

---

## Model Comparison

### Side-by-Side Comparison

```
┌────────────────────────────────────────────────────────────────┐
│  Model Comparison                                               │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  [Llama 3 70B    ▼]  vs  [GPT-4o         ▼]  [+ Add Model]   │
│                                                                │
│  ┌──────────────┬──────────────┬──────────────┐               │
│  │  Metric      │ Llama 3 70B  │ GPT-4o       │               │
│  ├──────────────┼──────────────┼──────────────┤               │
│  │  Type        │ Local        │ API          │               │
│  │  Context     │ 8K           │ 128K         │               │
│  │  VRAM        │ 14GB         │ N/A          │               │
│  │  Cost/1K tok │ $0.00 (local)│ $0.005       │               │
│  │  Latency P50 │ 120ms        │ 85ms         │               │
│  │  Latency P99 │ 340ms        │ 180ms        │               │
│  │  Throughput  │ 42 tok/s     │ 85 tok/s     │               │
│  │  MMLU Score  │ 79.5         │ 88.7         │               │
│  │  HumanEval   │ 67.0         │ 90.2         │               │
│  │  MATH        │ 35.2         │ 76.6         │               │
│  │  Error Rate  │ 0.1%         │ 0.0%         │               │
│  └──────────────┴──────────────┴──────────────┘               │
│                                                                │
│  Charts:                                                      │
│  ┌────────────────────────┐ ┌────────────────────────┐        │
│  │  Latency Comparison    │ │  Throughput Comparison │        │
│  │  ┌────┐                │ │  ┌────┐                │        │
│  │  │120 │  ┌────┐        │ │  │ 42 │  ┌────┐        │        │
│  │  │ ms │  │ 85 │        │ │  │t/s │  │ 85 │        │        │
│  │  └────┘  └────┘        │ │  └────┘  └────┘        │        │
│  │  Llama  GPT-4o         │ │  Llama  GPT-4o         │        │
│  └────────────────────────┘ └────────────────────────┘        │
└────────────────────────────────────────────────────────────────┘
```

---

## Model Usage Analytics

### Per-Tenant Usage

| Tenant        | Model        | Requests (30d) | Tokens Used | Avg Latency | Cost      |
|---------------|--------------|----------------|-------------|-------------|-----------|
| Acme Corp     | Llama 3 70B  | 12,450         | 3,240,000   | 125ms       | $0.00     |
| Acme Corp     | GPT-4o       | 8,230          | 2,100,000   | 88ms        | $10.50    |
| TechStart     | Llama 3 70B  | 5,120          | 1,280,000   | 118ms       | $0.00     |
| TechStart     | Embedding    | 15,600         | 780,000     | 12ms        | $0.00     |

### Per-User Usage

| User           | Model        | Requests (7d) | Tokens (7d) | Avg Latency | Top Task     |
|----------------|--------------|---------------|-------------|-------------|--------------|
| john@acme.com  | Llama 3 70B  | 340           | 89,000      | 130ms       | Chat         |
| john@acme.com  | GPT-4o       | 120           | 45,000      | 82ms        | Code Gen     |
| jane@acme.com  | Llama 3 70B  | 210           | 56,000      | 115ms       | Summarize    |

### Per-Agent Usage

| Agent          | Model        | Requests (24h) | Tokens (24h) | Success Rate |
|----------------|--------------|----------------|--------------|--------------|
| Customer Bot   | Llama 3 70B  | 890            | 234,000      | 99.2%        |
| Code Assistant | CodeLlama    | 450            | 178,000      | 97.8%        |
| Document RAG   | Embedding    | 2,100          | 420,000      | 100%         |

---

## Model Cost Tracking

### Cost Dashboard

```
┌──────────────────────────────────────────────────────────────┐
│  Model Cost Overview (Last 30 Days)                           │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Total Cost: $142.50                                         │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  API Costs by Model:                                   │  │
│  │                                                        │  │
│  │  GPT-4o:     $98.20   ████████████████████  (69%)     │  │
│  │  GPT-4o-mini: $32.10  ███████              (22%)      │  │
│  │  Embedding:   $12.20  ███                   (9%)      │  │
│  │  Local:       $0.00   (GPU electricity ≈ $4.50)       │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  Cost Trend (Last 30 Days)                                  │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  $10 ┤         ╷                                       │  │
│  │      │    ╷    │    ╷                                  │  │
│  │   $8 ┤    │    │    │         ╷                       │  │
│  │      │    │    │    │    ╷    │                       │  │
│  │   $6 ┤────┼────┼────┼────┼────┼────                   │  │
│  │      │    ╵    ╵    ╵    ╵    ╵                       │  │
│  │   $4 ┤                                                 │  │
│  │      └─────────────────────────────────────────        │  │
│  │       Week 1   Week 2   Week 3   Week 4               │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  Budget: $200/month    Spent: $142.50    Remaining: $57.50  │
│  ████████████████████████████████░░░░░░░░░░░░░░  71%        │
│                                                              │
│  [Set Budget Alert]  [Export Report]                         │
└──────────────────────────────────────────────────────────────┘
```

### Cost Per Request

| Model        | Input Cost/1K tokens | Output Cost/1K tokens | Avg Tokens/Request | Cost/Request |
|--------------|----------------------|-----------------------|--------------------|--------------|
| GPT-4o       | $0.005               | $0.015                | 850                | $0.009       |
| GPT-4o-mini  | $0.00015             | $0.0006               | 850                | $0.0004      |
| text-embedding-3 | $0.0001          | —                     | 500                | $0.00005     |
| Llama 3 70B  | $0.00 (local)        | $0.00 (local)         | 850                | ~$0.00001    |

---

## API Integration

### Endpoints

| Method   | Endpoint                       | Description                  |
|----------|--------------------------------|------------------------------|
| `POST`   | `/api/models`                  | List all models              |
| `POST`   | `/api/models/:id`              | Get model details            |
| `POST`   | `/api/models/download`         | Start model download         |
| `DELETE` | `/api/models/:id`              | Remove model                 |
| `POST`   | `/api/models/:id/start`        | Start/load model             |
| `POST`   | `/api/models/:id/stop`         | Stop/unload model            |
| `PATCH`  | `/api/models/:id/config`       | Update model configuration   |
| `POST`   | `/api/models/:id/performance`  | Get performance metrics      |
| `POST`   | `/api/models/:id/logs`         | Get model logs               |
| `POST`   | `/api/gpu`                     | Get GPU status               |
| `POST`   | `/api/gpu/usage`               | Get GPU usage history        |
| `POST`   | `/api/models/routing`          | Get routing rules            |
| `PATCH`  | `/api/models/routing`          | Update routing rules         |
| `POST`   | `/api/models/health`           | Get health status            |
| `POST`   | `/api/models/:id/health-check` | Trigger health check         |
| `POST`   | `/api/models/compare`          | Compare models               |
| `POST`   | `/api/models/analytics`        | Get usage analytics          |
| `POST`   | `/api/models/costs`            | Get cost data                |

---

## Hooks

### useModels

```typescript
// hooks/models/useModels.ts
export function useModels(filters?: ModelFilters) {
  const queryClient = useQueryClient();

  const models = useQuery({
    queryKey: ['models', filters],
    queryFn: () => api.post('/models', filters),
    staleTime: 10_000,
  });

  const startModel = useMutation({
    mutationFn: (id: string) => api.post(`/models/${id}/start`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['models'] }),
  });

  const stopModel = useMutation({
    mutationFn: (id: string) => api.post(`/models/${id}/stop`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['models'] }),
  });

  const removeModel = useMutation({
    mutationFn: (id: string) => api.delete(`/models/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['models'] }),
  });

  return { models, startModel, stopModel, removeModel };
}
```

### useModelPerformance

```typescript
// hooks/models/useModelPerformance.ts
export function useModelPerformance(modelId: string, timeRange: TimeRange) {
  return useQuery({
    queryKey: ['model-performance', modelId, timeRange],
    queryFn: () => api.post(`/models/${modelId}/performance`, { range: timeRange }),
    refetchInterval: 30_000,
  });
}
```

### useGPU

```typescript
// hooks/models/useGPU.ts
export function useGPU() {
  const gpuStatus = useQuery({
    queryKey: ['gpu'],
    queryFn: () => api.post('/gpu', {}),
    refetchInterval: 5_000,
  });

  const gpuHistory = useQuery({
    queryKey: ['gpu-history'],
    queryFn: () => api.post('/gpu/usage', {}),
  });

  return { gpuStatus, gpuHistory };
}
```

---

## Stores

### Models Store (Zustand)

```typescript
// stores/models/models.ts
import { create } from 'zustand';

interface ModelsState {
  models: Model[];
  selectedModel: Model | null;
  filters: ModelFilters;
  view: 'grid' | 'table';
  comparisonIds: string[];

  setModels: (models: Model[]) => void;
  setSelectedModel: (model: Model | null) => void;
  setFilters: (filters: ModelFilters) => void;
  setView: (view: 'grid' | 'table') => void;
  addToComparison: (id: string) => void;
  removeFromComparison: (id: string) => void;
  clearComparison: () => void;
}

export const useModelsStore = create<ModelsState>((set) => ({
  models: [],
  selectedModel: null,
  filters: {},
  view: 'grid',
  comparisonIds: [],

  setModels: (models) => set({ models }),
  setSelectedModel: (model) => set({ selectedModel: model }),
  setFilters: (filters) => set({ filters }),
  setView: (view) => set({ view }),
  addToComparison: (id) => set((state) => ({
    comparisonIds: state.comparisonIds.length < 3
      ? [...state.comparisonIds, id]
      : state.comparisonIds,
  })),
  removeFromComparison: (id) => set((state) => ({
    comparisonIds: state.comparisonIds.filter((cid) => cid !== id),
  })),
  clearComparison: () => set({ comparisonIds: [] }),
}));
```

### GPU Store

```typescript
// stores/models/gpu.ts
import { create } from 'zustand';

interface GPUState {
  gpus: GPUInfo[];
  utilizationHistory: GPUUtilizationPoint[];
  selectedGPU: number | null;

  setGPUs: (gpus: GPUInfo[]) => void;
  setUtilizationHistory: (history: GPUUtilizationPoint[]) => void;
  setSelectedGPU: (index: number | null) => void;
}

export const useGPUStore = create<GPUState>((set) => ({
  gpus: [],
  utilizationHistory: [],
  selectedGPU: null,

  setGPUs: (gpus) => set({ gpus }),
  setUtilizationHistory: (utilizationHistory) => set({ utilizationHistory }),
  setSelectedGPU: (selectedGPU) => set({ selectedGPU }),
}));
```

---

## Error Handling

### Error Types

| Error                | Message                           | Action                            |
|----------------------|-----------------------------------|-----------------------------------|
| Model not found      | "Model {name} not found"          | Refresh list                      |
| Download failed      | "Download interrupted"            | Retry download                    |
| OOM (Out of Memory)  | "Insufficient GPU memory"         | Stop another model or upgrade GPU |
| Model unavailable    | "API model temporarily unavailable"| Switch to fallback                |
| Config invalid       | "Invalid parameter value"         | Highlight invalid field           |
| Health check timeout | "Health check timed out"          | Auto-retry in 30s                 |
| GPU driver error     | "GPU driver not responding"       | Restart GPU service               |
| Model corrupted      | "Model files corrupted"           | Re-download model                 |

### Error Display

```tsx
// components/models/ModelError.tsx
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { AlertCircle, RefreshCw, Trash2 } from 'lucide-react';

interface ModelErrorProps {
  model: Model;
  error: ModelError;
  onRetry: () => void;
  onRemove: () => void;
}

export function ModelError({ model, error, onRetry, onRemove }: ModelErrorProps) {
  return (
    <Alert variant="destructive">
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>Error with {model.name}</AlertTitle>
      <AlertDescription className="mt-2">
        <p>{error.message}</p>
        {error.details && (
          <pre className="mt-2 text-xs bg-muted p-2 rounded overflow-auto max-h-20">
            {error.details}
          </pre>
        )}
        <div className="mt-3 flex gap-2">
          <Button size="sm" variant="outline" onClick={onRetry}>
            <RefreshCw className="w-3 h-3 mr-1" /> Retry
          </Button>
          <Button size="sm" variant="destructive" onClick={onRemove}>
            <Trash2 className="w-3 h-3 mr-1" /> Remove
          </Button>
        </div>
      </AlertDescription>
    </Alert>
  );
}
```

---

## Responsive Design

### Mobile Layout (≤640px)

```
┌──────────────────────┐
│  Model Management    │
│  [Grid] [Table]      │
├──────────────────────┤
│  ┌──────────────────┐│
│  │ 🧠 Llama 3 70B   ││
│  │ ● Running        ││
│  │ 14GB / 24GB      ││
│  │ 450 req/day      ││
│  │ [View] [Stop]    ││
│  └──────────────────┘│
│  ┌──────────────────┐│
│  │ 🧠 GPT-4o        ││
│  │ ● Running        ││
│  │ API              ││
│  │ 890 req/day      ││
│  │ [View] [Stop]    ││
│  └──────────────────┘│
└──────────────────────┘
```

### Tablet Layout (641px–1024px)

```
┌──────────────────────────────────────────┐
│  Model Management          [Grid] [Table]│
├──────────────────────────────────────────┤
│  ┌──────────────────┐ ┌──────────────────┐│
│  │ 🧠 Llama 3 70B   │ │ 🧠 GPT-4o        ││
│  │ ● Running        │ │ ● Running        ││
│  │ 14GB / 24GB      │ │ API              ││
│  │ 450 req/day      │ │ 890 req/day      ││
│  │ Latency: 120ms   │ │ Latency: 85ms    ││
│  │ [View] [Stop]    │ │ [View] [Stop]    ││
│  └──────────────────┘ └──────────────────┘│
│  ┌──────────────────┐ ┌──────────────────┐│
│  │ 🧠 Gemma 2 27B   │ │ 🧠 Mistral 7B    ││
│  │ ○ Stopped        │ │ ● Error          ││
│  │ 8GB / 24GB       │ │ 4GB / 24GB       ││
│  │ [Start] [View]   │ │ [Restart] [View] ││
│  └──────────────────┘ └──────────────────┘│
└──────────────────────────────────────────┘
```

### Responsive Breakpoints

| Breakpoint | Layout                                    |
|------------|-------------------------------------------|
| ≤640px     | Single column cards, stacked metrics      |
| 641–1024px | 2-column card grid, side-by-side charts   |
| 1025–1280px| 3-column card grid, full dashboard        |
| >1280px    | 4-column grid, split-panel detail view    |
