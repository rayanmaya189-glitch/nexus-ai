# AeroXe Nexus AI — Agent Management

## Agent CRUD, Configuration, Execution, Document Set Binding, Database Binding, Schema Discovery

---

## 1. Agent Management Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                     Agent Management Module                        │
│                                                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │  Agent List   │  │ Agent Create │  │ Agent Detail │           │
│  │  Page         │  │ Page         │  │ Page         │           │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘           │
│         │                  │                  │                     │
│         └──────────────────┼──────────────────┘                     │
│                            │                                        │
│                  ┌─────────▼─────────┐                             │
│                  │   Agent Store      │                             │
│                  │   (Zustand)        │                             │
│                  └─────────┬─────────┘                             │
│                            │                                        │
│         ┌──────────────────┼──────────────────┐                    │
│         │                  │                   │                    │
│  ┌──────▼───────┐  ┌──────▼───────┐  ┌──────▼───────┐           │
│  │ Agent Config  │  │ Agent        │  │ Agent        │           │
│  │ Page          │  │ Execute Page │  │ History Page │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
│                                                                    │
│  ┌──────────────────────────────────────────────────────┐        │
│  │  Sub-modules:                                         │        │
│  │  ├─ Document Set Binding                              │        │
│  │  ├─ Database Connection Binding                       │        │
│  │  ├─ Schema Discovery                                  │        │
│  │  ├─ Table Binding                                     │        │
│  │  └─ Test Connection                                   │        │
│  └──────────────────────────────────────────────────────┘        │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. Agent Types & Models

### 2.1 Agent Type Reference

| Type | Model | Purpose | Key Tools |
|---|---|---|---|
| `planner` | lfm2.5-thinking:1.2b | Intent detection, task planning | Agent selection, step decomposition |
| `customer` | phi4-mini:3.8b | Customer support, FAQ | customer.lookup, ticket.create |
| `developer` | qwen2.5-coder:3b | Code generation, review | git.search, code.analyze, test.execute |
| `rag` | command-r7b:7b | Knowledge reasoning | Search knowledge, document retrieval |
| `vision` | qwen3-vl:4b | Image understanding, OCR | Image analysis, text extraction |
| `security` | whiterabbitneo:7b | Security analysis | Vulnerability scan, threat assessment |
| `business` | llama3.1:7b | Business intelligence | Revenue analysis, forecasting |

### 2.2 Agent Entity

```typescript
// types/agent.types.ts
interface Agent {
  id: string;
  name: string;
  type: AgentType;
  model: string;
  systemPrompt: string;
  description: string;
  capabilities: string[];
  status: "active" | "inactive";
  temperature: number;
  maxTokens: number;
  documentSets: DocumentSetBinding[];
  databases: DatabaseBinding[];
  createdAt: string;
  updatedAt: string;
  createdBy: string;
}

type AgentType = "planner" | "customer" | "developer" | "rag" | "vision" | "security" | "business";

interface AgentExecution {
  id: string;
  agentId: string;
  task: string;
  status: ExecutionStatus;
  steps: ExecutionStep[];
  result?: string;
  tokensUsed: number;
  latencyMs: number;
  startedAt: string;
  completedAt?: string;
}

type ExecutionStatus = "pending" | "planning" | "executing" | "waiting" | "completed" | "failed" | "cancelled";

interface ExecutionStep {
  id: string;
  stepNumber: number;
  agentType: string;
  action: string;
  toolName?: string;
  toolParams?: Record<string, unknown>;
  result?: unknown;
  status: ExecutionStatus;
  startedAt: string;
  completedAt?: string;
}
```

---

## 3. Agent List Page

### 3.1 Page Layout

```
┌──────────────────────────────────────────────────────────────┐
│  Agents                                          [+ Create]   │
│                                                                │
│  ┌──────────────────────┐ ┌────────┐ ┌────────┐ ┌────────┐  │
│  │ 🔍 Search agents...  │ │ Type ▼ │ │Status ▼│ │ Sort ▼ │  │
│  └──────────────────────┘ └────────┘ └────────┘ └────────┘  │
│                                                                │
│  View: [Grid] [List]                                           │
│                                                                │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐                      │
│  │ Agent 1  │ │ Agent 2  │ │ Agent 3  │                      │
│  │ 🟢 Active│ │ 🟢 Active│ │ 🟡 Idle  │                      │
│  │ phi4-mini│ │ qwen-    │ │ command- │                      │
│  │          │ │ coder    │ │ r7b      │                      │
│  │ 1.2K req │ │ 856 req  │ │ 2.1K req │                      │
│  │ 340ms    │ │ 280ms    │ │ 520ms    │                      │
│  └──────────┘ └──────────┘ └──────────┘                      │
│                                                                │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐                      │
│  │ Agent 4  │ │ Agent 5  │ │ Agent 6  │                      │
│  └──────────┘ └──────────┘ └──────────┘                      │
│                                                                │
│  Showing 6 of 12 agents     < 1 2 3 >                         │
└──────────────────────────────────────────────────────────────┘
```

### 3.2 Agent List Component

```tsx
// components/agent/AgentList.tsx
export function AgentList() {
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [searchQuery, setSearchQuery] = useState("");
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const { data: agents, isLoading } = useQuery({
    queryKey: ["agents", { search: searchQuery, type: typeFilter, status: statusFilter }],
    queryFn: () => agentApi.list({ search: searchQuery, type: typeFilter, status: statusFilter }),
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Agents</h1>
          <p className="text-muted-foreground">Manage your AI agents</p>
        </div>
        <Button onClick={() => router.push("/dashboard/agents/new")}>
          <Plus className="mr-2 h-4 w-4" />
          Create Agent
        </Button>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-3">
        <SearchInput value={searchQuery} onChange={setSearchQuery} placeholder="Search agents..." />
        <Select value={typeFilter} onValueChange={setTypeFilter}>
          <SelectTrigger className="w-40"><SelectValue placeholder="All Types" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            {AGENT_TYPES.map((t) => <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>)}
          </SelectContent>
        </Select>
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-32"><SelectValue placeholder="All Status" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="inactive">Inactive</SelectItem>
          </SelectContent>
        </Select>
        <div className="ml-auto flex gap-1">
          <Button variant={viewMode === "grid" ? "default" : "ghost"} size="icon" onClick={() => setViewMode("grid")}>
            <LayoutGrid className="h-4 w-4" />
          </Button>
          <Button variant={viewMode === "list" ? "default" : "ghost"} size="icon" onClick={() => setViewMode("list")}>
            <List className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Agent Grid/List */}
      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-48" />)}
        </div>
      ) : viewMode === "grid" ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {agents?.map((agent) => <AgentCard key={agent.id} agent={agent} />)}
        </div>
      ) : (
        <AgentTable agents={agents || []} />
      )}
    </div>
  );
}
```

### 3.3 Agent Card Component

```tsx
// components/agent/AgentCard.tsx
export function AgentCard({ agent }: { agent: Agent }) {
  return (
    <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => router.push(`/dashboard/agents/${agent.id}`)}>
      <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-2">
        <div className="flex items-center gap-2">
          <div className={cn(
            "h-2 w-2 rounded-full",
            agent.status === "active" ? "bg-green-500" : "bg-gray-400"
          )} />
          <CardTitle className="text-base">{agent.name}</CardTitle>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
            <Button variant="ghost" size="icon">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuItem onClick={() => router.push(`/dashboard/agents/${agent.id}/configure`)}>
              <Settings className="mr-2 h-4 w-4" />Configure
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => router.push(`/dashboard/agents/${agent.id}/execute`)}>
              <Play className="mr-2 h-4 w-4" />Execute
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => router.push(`/dashboard/agents/${agent.id}/history`)}>
              <History className="mr-2 h-4 w-4" />History
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="text-destructive">
              <Trash2 className="mr-2 h-4 w-4" />Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-2 mb-3">
          <Badge variant="secondary">{agent.type}</Badge>
          <Badge variant="outline">{agent.model}</Badge>
        </div>
        <p className="text-sm text-muted-foreground line-clamp-2 mb-3">{agent.description}</p>
        <div className="grid grid-cols-3 gap-2 text-center text-xs">
          <div>
            <p className="font-medium">{agent.metrics?.totalExecutions ?? 0}</p>
            <p className="text-muted-foreground">Executions</p>
          </div>
          <div>
            <p className="font-medium">{agent.metrics?.avgLatency ?? 0}ms</p>
            <p className="text-muted-foreground">Avg Latency</p>
          </div>
          <div>
            <p className="font-medium">{agent.metrics?.successRate ?? 100}%</p>
            <p className="text-muted-foreground">Success</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
```

---

## 4. Agent Creation Page

### 4.1 Create Form Schema

```typescript
const agentCreateSchema = z.object({
  name: z.string()
    .min(2, "Name must be at least 2 characters")
    .max(100, "Name must be at most 100 characters"),
  type: z.enum(["customer", "developer", "rag", "vision", "security", "business"]),
  model: z.string().min(1, "Model is required"),
  description: z.string().min(10, "Description must be at least 10 characters"),
  systemPrompt: z.string().min(20, "System prompt must be at least 20 characters"),
  temperature: z.number().min(0).max(2).default(0.7),
  maxTokens: z.number().min(256).max(32768).default(4096),
  capabilities: z.array(z.string()).min(1, "Select at least one capability"),
  documentSets: z.array(z.string()).optional(),
  databases: z.array(z.string()).optional(),
});
```

### 4.2 Create Form Layout

```
┌──────────────────────────────────────────────────────────────┐
│  Create New Agent                                             │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ Basic Information                                     │    │
│  │                                                        │    │
│  │ Agent Name                                            │    │
│  │ ┌──────────────────────────────────────────────────┐  │    │
│  │ │ Customer Support Agent                           │  │    │
│  │ └──────────────────────────────────────────────────┘  │    │
│  │                                                        │    │
│  │ Type                                                  │    │
│  │ ┌────────────────┐ ┌──────────┐ ┌────────────────┐  │    │
│  │ │   🤝 Customer  │ │💻 Dev    │ │ 📚 RAG         │  │    │
│  │ └────────────────┘ └──────────┘ └────────────────┘  │    │
│  │ ┌────────────────┐ ┌──────────┐ ┌────────────────┐  │    │
│  │ │ 👁 Vision      │ │🔒 Sec   │ │ 📊 Business    │  │    │
│  │ └────────────────┘ └──────────┘ └────────────────┘  │    │
│  │                                                        │    │
│  │ Model                                                 │    │
│  │ ┌──────────────────────────────────────────────────┐  │    │
│  │ │ phi4-mini:3.8b                                  ▼│  │    │
│  │ └──────────────────────────────────────────────────┘  │    │
│  │                                                        │    │
│  │ Description                                           │    │
│  │ ┌──────────────────────────────────────────────────┐  │    │
│  │ │ Handles customer support queries, FAQ, and ticket│  │    │
│  │ └──────────────────────────────────────────────────┘  │    │
│  │                                                        │    │
│  │ System Prompt                                         │    │
│  │ ┌──────────────────────────────────────────────────┐  │    │
│  │ │ You are a customer support agent for AeroXe...   │  │    │
│  │ │ You have access to customer database and...      │  │    │
│  │ │                                                   │  │    │
│  │ └──────────────────────────────────────────────────┘  │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ Configuration                                         │    │
│  │                                                        │    │
│  │ Temperature ───────●──── 0.7     Max Tokens: 4096     │    │
│  │                                                        │    │
│  │ Capabilities:                                          │    │
│  │ ☑ Customer Lookup  ☑ Ticket Creation                  │    │
│  │ ☑ Billing Check    ☐ Network Status                   │    │
│  │ ☑ Knowledge Search ☐ Code Generation                  │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ Resources (Optional)                                  │    │
│  │                                                        │    │
│  │ Document Sets: [Select sets to bind]                  │    │
│  │ Databases:     [Select databases to bind]             │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌──────────────────────────────────────────────┐           │
│  │ [Cancel]                    [Create Agent]     │           │
│  └──────────────────────────────────────────────┘           │
└──────────────────────────────────────────────────────────────┘
```

### 4.3 Create Form Component

```tsx
// components/agent/AgentForm.tsx
export function AgentForm({ mode = "create", initialData }: AgentFormProps) {
  const form = useForm<AgentFormData>({
    resolver: zodResolver(mode === "create" ? agentCreateSchema : agentUpdateSchema),
    defaultValues: initialData ?? {
      name: "",
      type: "customer",
      model: "",
      description: "",
      systemPrompt: "",
      temperature: 0.7,
      maxTokens: 4096,
      capabilities: [],
    },
  });

  const selectedType = form.watch("type");
  const models = MODELS_BY_TYPE[selectedType] || [];

  // Auto-select first model when type changes
  useEffect(() => {
    if (models.length > 0 && !models.includes(form.getValues("model"))) {
      form.setValue("model", models[0]);
    }
  }, [selectedType, models, form]);

  const createMutation = useMutation({
    mutationFn: agentApi.create,
    onSuccess: (agent) => {
      toast.success("Agent created successfully");
      queryClient.invalidateQueries({ queryKey: ["agents"] });
      router.push(`/dashboard/agents/${agent.id}`);
    },
    onError: (error) => {
      toast.error("Failed to create agent", { description: error.message });
    },
  });

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit((data) => createMutation.mutate(data))} className="space-y-6">
        {/* Type Selection Grid */}
        <FormField
          control={form.control}
          name="type"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Agent Type</FormLabel>
              <div className="grid grid-cols-3 gap-3">
                {AGENT_TYPES.map((type) => (
                  <Button
                    key={type.value}
                    type="button"
                    variant={field.value === type.value ? "default" : "outline"}
                    className="flex flex-col items-center gap-2 h-20"
                    onClick={() => field.onChange(type.value)}
                  >
                    <type.icon className="h-5 w-5" />
                    <span className="text-xs">{type.label}</span>
                  </Button>
                ))}
              </div>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Name, Model, Description, System Prompt, etc. */}

        {/* Temperature & Max Tokens */}
        <div className="grid grid-cols-2 gap-4">
          <FormField
            control={form.control}
            name="temperature"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Temperature: {field.value}</FormLabel>
                <FormControl>
                  <Slider
                    min={0}
                    max={2}
                    step={0.1}
                    value={[field.value]}
                    onValueChange={(v) => field.onChange(v[0])}
                  />
                </FormControl>
                <FormDescription>Higher = more creative, lower = more deterministic</FormDescription>
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="maxTokens"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Max Tokens</FormLabel>
                <FormControl>
                  <Input type="number" {...field} />
                </FormControl>
              </FormItem>
            )}
          />
        </div>

        <Button type="submit" disabled={createMutation.isPending}>
          {createMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Create Agent
        </Button>
      </form>
    </Form>
  );
}
```

---

## 5. Agent Configuration Page

### 5.1 Configuration Sections

```
┌──────────────────────────────────────────────────────────────┐
│  Agent: Customer Support Agent                    [Execute]   │
│                                                                │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐     │
│  │General  │ │Document  │ │Database  │ │  Permissions  │     │
│  │Settings │ │Sets      │ │Connections│ │              │     │
│  └─────────┘ └──────────┘ └──────────┘ └──────────────┘     │
│                                                                │
│  ═══════════════════════════════════════════════════════════   │
│                                                                │
│  General Settings Tab (active)                                 │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ Model: phi4-mini:3.8b                                │    │
│  │ Temperature: 0.7                                     │    │
│  │ Max Tokens: 4096                                     │    │
│  │ System Prompt: [editable textarea]                   │    │
│  │ Capabilities: [tag list, editable]                   │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                                │
│  Document Sets Tab                                             │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ Bound Sets:                                          │    │
│  │ ┌────────────────────────────────────────────────┐   │    │
│  │ │ 📁 Customer FAQ (read)           [Unbind]      │   │    │
│  │ │ 📁 Troubleshooting Guide (read)  [Unbind]      │   │    │
│  │ └────────────────────────────────────────────────┘   │    │
│  │                                                      │    │
│  │ Available Sets:                                      │    │
│  │ ┌────────────────────────────────────────────────┐   │    │
│  │ │ 📁 Billing Policies          [Bind]            │   │    │
│  │ │ 📁 Network Documentation     [Bind]            │   │    │
│  │ └────────────────────────────────────────────────┘   │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                                │
│  Database Connections Tab                                       │
│  ┌──────────────────────────────────────────────────────┐    │
│  │ [+ Add Connection]                                   │    │
│  │                                                      │    │
│  │ Connection: PostgreSQL (billing-db)  ✅ Connected     │    │
│  │ ├─ Bound Tables:                                     │    │
│  │ │  ☑ customers  ☑ invoices  ☐ payments              │    │
│  │ ├─ [Test Connection] [Discover Schema] [Unbind]      │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                                │
└──────────────────────────────────────────────────────────────┘
```

---

## 6. Agent-Document Set Binding

### 6.1 Binding Component

```tsx
// components/agent/AgentDocSetBinding.tsx
export function AgentDocSetBinding({ agentId }: { agentId: string }) {
  const { data: boundSets } = useQuery({
    queryKey: ["agents", agentId, "document-sets"],
    queryFn: () => agentApi.getDocumentSets(agentId),
  });

  const { data: availableSets } = useQuery({
    queryKey: ["document-sets", "available"],
    queryFn: documentSetApi.list,
  });

  const bindMutation = useMutation({
    mutationFn: (data: { documentSetId: string; permissionLevel: string }) =>
      agentApi.bindDocumentSet(agentId, data),
    onSuccess: () => {
      toast.success("Document set bound");
      queryClient.invalidateQueries({ queryKey: ["agents", agentId, "document-sets"] });
    },
  });

  const unbindMutation = useMutation({
    mutationFn: (setId: string) => agentApi.unbindDocumentSet(agentId, setId),
    onSuccess: () => {
      toast.success("Document set unbound");
      queryClient.invalidateQueries({ queryKey: ["agents", agentId, "document-sets"] });
    },
  });

  const unboundSets = availableSets?.filter(
    (s) => !boundSets?.some((b) => b.documentSetId === s.id)
  );

  return (
    <div className="space-y-6">
      {/* Currently Bound Sets */}
      <div>
        <h3 className="text-lg font-medium mb-3">Bound Document Sets</h3>
        {boundSets?.length === 0 ? (
          <p className="text-sm text-muted-foreground">No document sets bound. Agent cannot search knowledge.</p>
        ) : (
          <div className="space-y-2">
            {boundSets?.map((binding) => (
              <div key={binding.id} className="flex items-center justify-between rounded-lg border p-3">
                <div className="flex items-center gap-3">
                  <FolderOpen className="h-5 w-5 text-muted-foreground" />
                  <div>
                    <p className="font-medium">{binding.documentSetName}</p>
                    <p className="text-xs text-muted-foreground">
                      {binding.documentCount} documents · {binding.totalChunks} chunks
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant="secondary">{binding.permissionLevel}</Badge>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => unbindMutation.mutate(binding.documentSetId)}
                  >
                    <Unlink className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Available Sets to Bind */}
      {unboundSets && unboundSets.length > 0 && (
        <div>
          <h3 className="text-lg font-medium mb-3">Available Document Sets</h3>
          <div className="space-y-2">
            {unboundSets.map((set) => (
              <div key={set.id} className="flex items-center justify-between rounded-lg border p-3">
                <div className="flex items-center gap-3">
                  <Folder className="h-5 w-5 text-muted-foreground" />
                  <div>
                    <p className="font-medium">{set.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {set.documentCount} documents
                    </p>
                  </div>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => bindMutation.mutate({
                    documentSetId: set.id,
                    permissionLevel: "read",
                  })}
                >
                  <Link2 className="mr-1 h-4 w-4" />
                  Bind
                </Button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
```

---

## 7. Agent-Database Connection Binding

### 7.1 Database Connection Form

```tsx
// components/agent/AgentDbBinding.tsx
const dbConnectionSchema = z.object({
  connectionName: z.string().min(1, "Connection name is required"),
  host: z.string().min(1, "Host is required"),
  port: z.number().min(1).max(65535).default(5432),
  databaseName: z.string().min(1, "Database name is required"),
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
  sslMode: z.enum(["disable", "require", "verify-ca", "verify-full"]).default("require"),
});

export function AgentDbBinding({ agentId }: { agentId: string }) {
  const [showForm, setShowForm] = useState(false);
  const { data: connections, isLoading } = useQuery({
    queryKey: ["agents", agentId, "sql-connections"],
    queryFn: () => agentApi.getConnections(agentId),
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium">Database Connections</h3>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Connection
        </Button>
      </div>

      {showForm && (
        <DBConnectionForm
          agentId={agentId}
          onSubmit={(data) => createConnection.mutate(data)}
          onCancel={() => setShowForm(false)}
        />
      )}

      <div className="space-y-3">
        {connections?.map((conn) => (
          <DatabaseConnectionCard key={conn.id} connection={conn} agentId={agentId} />
        ))}
      </div>
    </div>
  );
}
```

### 7.2 Database Connection Card

```tsx
function DatabaseConnectionCard({ connection, agentId }: {
  connection: DatabaseConnection;
  agentId: string;
}) {
  const statusColors = {
    pending: "bg-yellow-500",
    connected: "bg-green-500",
    failed: "bg-red-500",
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="flex items-center gap-2">
          <Database className="h-5 w-5" />
          <CardTitle className="text-base">{connection.connectionName}</CardTitle>
          <div className={cn("h-2 w-2 rounded-full", statusColors[connection.status])} />
          <Badge variant={connection.status === "connected" ? "default" : "destructive"}>
            {connection.status}
          </Badge>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => testConnection(connection)}>
            <Zap className="mr-1 h-4 w-4" />Test
          </Button>
          <Button variant="outline" size="sm" onClick={() => discoverSchema(connection)}>
            <Search className="mr-1 h-4 w-4" />Discover
          </Button>
          <Button variant="destructive" size="sm" onClick={() => removeConnection(connection.id)}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-4 gap-4 text-sm text-muted-foreground">
          <div><span className="font-medium text-foreground">Host:</span> {connection.host}:{connection.port}</div>
          <div><span className="font-medium text-foreground">Database:</span> {connection.databaseName}</div>
          <div><span className="font-medium text-foreground">SSL:</span> {connection.sslMode}</div>
          <div><span className="font-medium text-foreground">Tables:</span> {connection.discoveredTablesCount}</div>
        </div>

        {/* Bound Tables */}
        {connection.tables && connection.tables.length > 0 && (
          <div className="mt-4">
            <p className="text-sm font-medium mb-2">Bound Tables:</p>
            <div className="flex flex-wrap gap-2">
              {connection.tables.map((table) => (
                <Badge key={table.name} variant="secondary">
                  {table.name}
                  <button className="ml-1" onClick={() => unbindTable(connection.id, table.name)}>
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              ))}
            </div>
          </div>
        )}

        {connection.lastTestResult && (
          <div className="mt-3 text-xs">
            Last tested: {formatRelativeTime(connection.lastTestedAt)} —
            <span className={connection.lastTestResult === "success" ? "text-green-500" : "text-red-500"}>
              {connection.lastTestResult}
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
```

---

## 8. Test Connection Flow

### 8.1 Connection Test UI

```
┌──────────────────────────────────────────────────────────┐
│  Test Database Connection                                  │
│                                                            │
│  Connection: PostgreSQL (billing-db)                       │
│  Host: db.aeroxenexus.com:5432                             │
│                                                            │
│  ┌──────────────────────────────────────────────────┐    │
│  │  🔍 Testing connection...                         │    │
│  │                                                    │    │
│  │  Step 1/4: TCP Connection      ✅ (12ms)          │    │
│  │  Step 2/4: Authentication      ✅ (8ms)           │    │
│  │  Step 3/4: SSL Handshake       ✅ (15ms)          │    │
│  │  Step 4/4: Database Access     ⏳ Checking...     │    │
│  └──────────────────────────────────────────────────┘    │
│                                                            │
│  Status: Connected ✅                                      │
│  Server Version: PostgreSQL 16.0                           │
│  Total Tables: 47                                         │
│  Response Time: 42ms                                      │
│                                                            │
│  [Discover Schema]                                        │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

### 8.2 Test Connection Implementation

```tsx
// hooks/agent/useAgentDb.ts
export function useTestConnection() {
  const [testResult, setTestResult] = useState<TestResult | null>(null);
  const [isTesting, setIsTesting] = useState(false);
  const [steps, setSteps] = useState<TestStep[]>([]);

  const testConnection = async (connectionId: string) => {
    setIsTesting(true);
    setSteps([]);
    setTestResult(null);

    const stepDefs = ["TCP Connection", "Authentication", "SSL Handshake", "Database Access"];

    for (let i = 0; i < stepDefs.length; i++) {
      setSteps((prev) => [...prev, { name: stepDefs[i], status: "running" }]);

      try {
        const result = await agentApi.testConnection(connectionId, i);
        setSteps((prev) => prev.map((s, idx) =>
          idx === i ? { ...s, status: "success", duration: result.stepDuration } : s
        ));
      } catch (error) {
        setSteps((prev) => prev.map((s, idx) =>
          idx === i ? { ...s, status: "failed", error: error.message } : s
        ));
        setTestResult({ success: false, error: error.message });
        setIsTesting(false);
        return;
      }
    }

    setTestResult({ success: true, serverVersion: "...", tableCount: 47 });
    setIsTesting(false);
  };

  return { testConnection, testResult, isTesting, steps };
}
```

---

## 9. Schema Discovery

### 9.1 Schema Discovery UI

```
┌──────────────────────────────────────────────────────────┐
│  Schema Discovery — billing-db                             │
│                                                            │
│  Discovered 47 tables                                      │
│                                                            │
│  ┌──────────────────────────────────────────────────┐    │
│  │ ☐ | Table Name        | Columns | Rows (est) |   │    │
│  │---|-------------------|---------|-------------|   │    │
│  │ ☑ | customers         | 12      | ~50,000     |   │    │
│  │ ☑ | invoices          | 8       | ~120,000    |   │    │
│  │ ☐ | payments          | 10      | ~95,000     |   │    │
│  │ ☐ | subscriptions     | 6       | ~48,000     |   │    │
│  │ ☑ | support_tickets   | 14      | ~30,000     |   │    │
│  │ ☐ | audit_logs        | 5       | ~500,000    |   │    │
│  │ ☐ | products          | 9       | ~200        |   │    │
│  │ ☑ | billing_plans     | 7       | ~15         |   │    │
│  └──────────────────────────────────────────────────┘    │
│                                                            │
│  Selected: 4 tables                                        │
│                                                            │
│  ┌──────────────────────────┐  ┌────────────────────┐    │
│  │ [Cancel]                  │  │ [Bind Selected]    │    │
│  └──────────────────────────┘  └────────────────────┘    │
└──────────────────────────────────────────────────────────┘
```

### 9.2 Schema Discovery Component

```tsx
// components/agent/AgentSchemaDiscovery.tsx
export function AgentSchemaDiscovery({ connectionId, agentId }: {
  connectionId: string;
  agentId: string;
}) {
  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());

  const { data: tables, isLoading, refetch } = useQuery({
    queryKey: ["agents", agentId, "sql-connections", connectionId, "tables"],
    queryFn: () => agentApi.discoverSchema(agentId, connectionId),
  });

  const bindTablesMutation = useMutation({
    mutationFn: (tableNames: string[]) =>
      agentApi.bindTables(agentId, connectionId, { tables: tableNames }),
    onSuccess: () => {
      toast.success(`${selectedTables.size} tables bound`);
      queryClient.invalidateQueries({ queryKey: ["agents", agentId, "sql-connections"] });
    },
  });

  const toggleTable = (tableName: string) => {
    setSelectedTables((prev) => {
      const next = new Set(prev);
      if (next.has(tableName)) next.delete(tableName);
      else next.add(tableName);
      return next;
    });
  };

  const toggleAll = () => {
    if (selectedTables.size === tables?.length) {
      setSelectedTables(new Set());
    } else {
      setSelectedTables(new Set(tables?.map((t) => t.name) || []));
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium">Schema Discovery</h3>
        <Button onClick={() => refetch()} disabled={isLoading}>
          <RefreshCw className={cn("mr-2 h-4 w-4", isLoading && "animate-spin")} />
          Re-Discover
        </Button>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-10">
                <Checkbox checked={selectedTables.size === tables?.length} onCheckedChange={toggleAll} />
              </TableHead>
              <TableHead>Table Name</TableHead>
              <TableHead>Columns</TableHead>
              <TableHead>Rows (est.)</TableHead>
              <TableHead>Primary Key</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell colSpan={5}><Skeleton className="h-8" /></TableCell>
                </TableRow>
              ))
            ) : (
              tables?.map((table) => (
                <TableRow key={table.name}>
                  <TableCell>
                    <Checkbox
                      checked={selectedTables.has(table.name)}
                      onCheckedChange={() => toggleTable(table.name)}
                    />
                  </TableCell>
                  <TableCell className="font-mono text-sm">{table.name}</TableCell>
                  <TableCell>{table.columns.length}</TableCell>
                  <TableCell>{formatNumber(table.rowCountEstimate)}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="font-mono text-xs">
                      {table.primaryKey?.join(", ") || "—"}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {selectedTables.size > 0 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">{selectedTables.size} tables selected</p>
          <Button onClick={() => bindTablesMutation.mutate(Array.from(selectedTables))}>
            Bind Selected Tables
          </Button>
        </div>
      )}
    </div>
  );
}
```

---

## 10. Agent Execution Page

### 10.1 Execution View

```
┌──────────────────────────────────────────────────────────┐
│  Execute Agent: Customer Support                          │
│                                                            │
│  ┌──────────────────────────────────────────────────┐    │
│  │ Task                                              │    │
│  │ ┌──────────────────────────────────────────────┐  │    │
│  │ │ Analyze customer complaint #12345            │  │    │
│  │ └──────────────────────────────────────────────┘  │    │
│  │                                                    │    │
│  │ [Start Execution]                                  │    │
│  └──────────────────────────────────────────────────┘    │
│                                                            │
│  ═══════════════════════════════════════════════════════   │
│                                                            │
│  Execution Progress                                         │
│                                                            │
│  Step 1: Planning           ✅ Completed (320ms)           │
│  ┌──────────────────────────────────────────────────┐    │
│  │ Agent: planner (lfm2.5-thinking)                  │    │
│  │ Plan: 1. Get customer data                        │    │
│  │       2. Check billing status                     │    │
│  │       3. Search knowledge base                    │    │
│  │       4. Generate response                        │    │
│  └──────────────────────────────────────────────────┘    │
│                                                            │
│  Step 2: Data Gathering     ✅ Completed (1.2s)           │
│  ┌──────────────────────────────────────────────────┐    │
│  │ ✓ customer.lookup(#12345) — 200ms               │    │
│  │ ✓ billing.check(#12345) — 180ms                 │    │
│  │ ✓ ticket.history(#12345) — 250ms                │    │
│  └──────────────────────────────────────────────────┘    │
│                                                            │
│  Step 3: Knowledge Search   ✅ Completed (800ms)           │
│  ┌──────────────────────────────────────────────────┐    │
│  │ ✓ RAG search — 3 documents found                │    │
│  └──────────────────────────────────────────────────┘    │
│                                                            │
│  Step 4: Response           ⏳ Streaming...                │
│  ┌──────────────────────────────────────────────────┐    │
│  │ Customer #12345 has 3 open support tickets...    │    │
│  │ Their billing status shows...                    │    │
│  │ Based on the knowledge base, the recommended     │    │
│  │ resolution is...                                 │    │
│  └──────────────────────────────────────────────────┘    │
│                                                            │
│  ────────────────────────────────────────────────────────  │
│  Tokens: 1,247  │  Latency: 2.8s  │  Status: Running      │
│                                                            │
│  [Cancel Execution]                                        │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

### 10.2 Execution View Component

```tsx
// components/agent/AgentExecutionView.tsx
export function AgentExecutionView({ agentId }: { agentId: string }) {
  const [task, setTask] = useState("");
  const [executionId, setExecutionId] = useState<string | null>(null);
  const { execution, isStreaming, cancelExecution } = useAgentExecution(executionId);

  const startMutation = useMutation({
    mutationFn: () => agentApi.execute(agentId, { task }),
    onSuccess: (data) => {
      setExecutionId(data.execution_id);
    },
  });

  return (
    <div className="space-y-6">
      {/* Task Input */}
      <Card>
        <CardHeader>
          <CardTitle>Task</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <Textarea
            value={task}
            onChange={(e) => setTask(e.target.value)}
            placeholder="Describe the task for this agent..."
            rows={3}
          />
          <div className="flex gap-2">
            <Button
              onClick={() => startMutation.mutate()}
              disabled={!task.trim() || startMutation.isPending || isStreaming}
            >
              {startMutation.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Play className="mr-2 h-4 w-4" />
              )}
              Start Execution
            </Button>
            {isStreaming && (
              <Button variant="destructive" onClick={cancelExecution}>
                <Square className="mr-2 h-4 w-4" />
                Cancel
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Execution Progress */}
      {execution && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              Execution Progress
              <Badge variant={execution.status === "completed" ? "default" : "secondary"}>
                {execution.status}
              </Badge>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {execution.steps.map((step) => (
              <ExecutionStepCard key={step.id} step={step} />
            ))}
          </CardContent>
          <CardFooter className="text-sm text-muted-foreground">
            Tokens: {execution.tokensUsed} · Latency: {execution.latencyMs}ms
          </CardFooter>
        </Card>
      )}
    </div>
  );
}
```

---

## 11. Execution History

```tsx
// Agent execution history table
export function AgentExecutionHistory({ agentId }: { agentId: string }) {
  const { data: executions, isLoading } = useQuery({
    queryKey: ["agents", agentId, "history"],
    queryFn: () => agentApi.getExecutionHistory(agentId),
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Execution History</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Task</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Tokens</TableHead>
              <TableHead>Latency</TableHead>
              <TableHead>Started</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {executions?.map((exec) => (
              <TableRow key={exec.id}>
                <TableCell className="font-mono text-xs">{exec.id.slice(0, 8)}</TableCell>
                <TableCell className="max-w-xs truncate">{exec.task}</TableCell>
                <TableCell>
                  <StatusBadge status={exec.status} />
                </TableCell>
                <TableCell>{exec.tokensUsed}</TableCell>
                <TableCell>{exec.latencyMs}ms</TableCell>
                <TableCell>{formatRelativeTime(exec.startedAt)}</TableCell>
                <TableCell>
                  <Button variant="ghost" size="sm" onClick={() => viewExecution(exec.id)}>
                    <Eye className="h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
```

---

## 12. Agent Store (Zustand)

```typescript
// stores/agent-store.ts
interface AgentState {
  // List
  agents: Agent[];
  setAgents: (agents: Agent[]) => void;

  // Current agent (detail/config pages)
  currentAgent: Agent | null;
  setCurrentAgent: (agent: Agent | null) => void;

  // Executions
  activeExecution: AgentExecution | null;
  setActiveExecution: (exec: AgentExecution | null) => void;

  // Config editing
  isDirty: boolean;
  setDirty: (value: boolean) => void;

  // Filters (list page)
  searchQuery: string;
  typeFilter: string;
  statusFilter: string;
  setSearchQuery: (query: string) => void;
  setTypeFilter: (type: string) => void;
  setStatusFilter: (status: string) => void;
}

export const useAgentStore = create<AgentState>((set) => ({
  agents: [],
  currentAgent: null,
  activeExecution: null,
  isDirty: false,
  searchQuery: "",
  typeFilter: "all",
  statusFilter: "all",

  setAgents: (agents) => set({ agents }),
  setCurrentAgent: (agent) => set({ currentAgent: agent }),
  setActiveExecution: (exec) => set({ activeExecution: exec }),
  setDirty: (value) => set({ isDirty: value }),
  setSearchQuery: (query) => set({ searchQuery: query }),
  setTypeFilter: (type) => set({ typeFilter: type }),
  setStatusFilter: (status) => set({ statusFilter: status }),
}));
```

---

## 13. Agent API Integration

### 13.1 CRUD Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/agents` | List agents |
| `POST` | `/api/v1/agents/:id` | Get agent details |
| `POST` | `/api/v1/agents` | Create agent |
| `PATCH` | `/api/v1/agents/:id` | Update agent |
| `DELETE` | `/api/v1/agents/:id` | Delete agent |

### 13.2 Execution Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/agents/:id/execute` | Start execution |
| `POST` | `/api/v1/agents/execution/:execId` | Get execution status |
| `DELETE` | `/api/v1/agents/execution/:execId` | Cancel execution |
| `POST` | `/api/v1/agents/:id/history` | Get execution history |

### 13.3 Document Set Binding Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/agents/:id/document-sets` | Bind document sets |
| `POST` | `/api/v1/agents/:id/document-sets` | List bound sets |
| `DELETE` | `/api/v1/agents/:id/document-sets/:setId` | Unbind set |

### 13.4 Database Binding Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/agents/:id/sql-connections` | Add database connection |
| `POST` | `/api/v1/agents/:id/sql-connections/test` | Test connection |
| `POST` | `/api/v1/agents/:id/sql-connections/discover` | Discover schema |
| `POST` | `/api/v1/agents/:id/sql-connections/:connId/tables` | Bind tables |
| `POST` | `/api/v1/agents/:id/sql-connections` | List connections |
| `DELETE` | `/api/v1/agents/:id/sql-connections/:connId` | Remove connection |

---

## 14. Agent Hooks

| Hook | Purpose | Returns |
|---|---|---|
| `useAgents(filters)` | List agents with filters | `{ agents, isLoading, pagination }` |
| `useAgent(id)` | Single agent detail | `{ agent, isLoading }` |
| `useAgentExecution(execId)` | Real-time execution status | `{ execution, isStreaming, cancel }` |
| `useAgentConfig(id)` | Agent configuration CRUD | `{ config, update, isDirty }` |
| `useAgentDocSets(id)` | Document set binding | `{ bound, available, bind, unbind }` |
| `useAgentDb(id)` | Database connections | `{ connections, test, discover, bind }` |
| `useTestConnection()` | Connection test flow | `{ test, result, steps, isTesting }` |
| `useSchemaDiscovery(connId)` | Schema discovery | `{ tables, isLoading, refetch }` |

---

*Document Version: 1.0 — Last Updated: July 2026*
