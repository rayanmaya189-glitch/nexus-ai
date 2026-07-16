# Workflow Builder

## Table of Contents

- [Overview](#overview)
- [Workflow Builder Page Layout](#workflow-builder-page-layout)
- [React Flow Integration](#react-flow-integration)
- [Node Types](#node-types)
- [Node Configuration Panel](#node-configuration-panel)
- [Edge Connections](#edge-connections)
- [Drag-and-Drop Node Creation](#drag-and-drop-node-creation)
- [Canvas Controls](#canvas-controls)
- [Workflow Properties](#workflow-properties)
- [Workflow Triggers](#workflow-triggers)
- [Workflow Execution View](#workflow-execution-view)
- [Workflow Execution History](#workflow-execution-history)
- [Workflow Approval Step](#workflow-approval-step)
- [Workflow Error Handling](#workflow-error-handling)
- [Workflow Testing](#workflow-testing)
- [Workflow Templates](#workflow-templates)
- [Workflow Import/Export](#workflow-importexport)
- [Workflow Versioning](#workflow-versioning)
- [API Integration](#api-integration)
- [Hooks](#hooks)
- [Stores](#stores)
- [Responsive Design](#responsive-design)

---

## Overview

The Workflow Builder is a visual, node-based editor for creating multi-step AI workflows. It uses React Flow to provide drag-and-drop canvas editing with configurable nodes representing different AI agents, conditions, and actions.

```
+------------------------------------------------------------------+
|  Workflow Builder - Customer Support Pipeline                     |
+------------------------------------------------------------------+
| [+Node] [Save] [Test] [Run] [Export]    Status: Draft  v1.2     |
+----------+----------------------------------------+-------------+
| Node     |                                        | Properties  |
| Palette  |    [User Input]                        |             |
|          |         |                              | Node:       |
| User     |    [Planner Agent]                     | User Input  |
| Input    |        /    \                          |             |
|          |   [RAG]    [SQL]                       | Label:      |
| Planner  |    |          |                        | "Customer   |
| Agent    |    +----+-----+                        |  Question"  |
|          |         |                              |             |
| RAG      |    [Condition]                         | Type:       |
| Agent    |      /      \                          | Text Input  |
|          |  [Approve]  [Response]                 |             |
| SQL      |     |         |                        | Required:   |
| Agent    |   [Human]   [END]                      | Yes         |
|          |                                        |             |
| Vision   |                                        | Placeholder:|
| Agent    |                                        | "Ask a      |
|          |                                        |  question..."|
| Response |                                        |             |
|          |                                        | [Save]      |
| Condition|                                        |             |
|          |                                        |             |
| Approval |                                        |             |
+----------+----------------------------------------+-------------+
|  Zoom: 100%  |  Nodes: 8  |  Edges: 7  |  Minimap [ON]         |
+------------------------------------------------------------------+
```

---

## Workflow Builder Page Layout

### Component Tree

```
WorkflowBuilderPage
+-- BuilderHeader
|   +-- BackLink
|   +-- WorkflowName (editable)
|   +-- StatusBadge
|   +-- ActionButtons (Save, Test, Run, Export)
+-- BuilderBody
|   +-- NodePalette (sidebar)
|   |   +-- DraggableNode[] (categorized)
|   |   +-- SearchFilter
|   +-- Canvas (ReactFlow)
|   |   +-- NodeTypes (custom nodes)
|   |   +-- EdgeTypes (custom edges)
|   |   +-- Controls
|   |   +-- MiniMap
|   |   +-- Background
|   +-- PropertiesPanel (right sidebar)
|       +-- NodeProperties (when node selected)
|       +-- WorkflowProperties (when no selection)
|       +-- EdgeProperties (when edge selected)
+-- BuilderFooter
    +-- ZoomControls
    +-- StatsDisplay
    +-- MinimapToggle
```

### Page Layout Code

```tsx
// pages/workflows/WorkflowBuilderPage.tsx
import { ReactFlowProvider } from 'reactflow';
import { NodePalette } from '@/components/workflows/NodePalette';
import { BuilderCanvas } from '@/components/workflows/BuilderCanvas';
import { PropertiesPanel } from '@/components/workflows/PropertiesPanel';
import { BuilderHeader } from '@/components/workflows/BuilderHeader';
import { useWorkflowBuilder } from '@/hooks/workflows/useWorkflowBuilder';

export function WorkflowBuilderPage() {
  const { workflow, selectedNode, selectedEdge } = useWorkflowBuilder();

  return (
    <ReactFlowProvider>
      <div className="flex flex-col h-screen">
        <BuilderHeader workflow={workflow} />
        <div className="flex-1 flex overflow-hidden">
          <NodePalette />
          <BuilderCanvas />
          <PropertiesPanel
            node={selectedNode}
            edge={selectedEdge}
          />
        </div>
      </div>
    </ReactFlowProvider>
  );
}
```

---

## React Flow Integration

### Node Types Registration

```tsx
// components/workflows/BuilderCanvas.tsx
import ReactFlow, { Node, Edge, Controls, MiniMap, Background } from 'reactflow';
import 'reactflow/dist/style.css';

import { UserInputNode } from '@/components/workflows/nodes/UserInputNode';
import { PlannerAgentNode } from '@/components/workflows/nodes/PlannerAgentNode';
import { RAGAgentNode } from '@/components/workflows/nodes/RAGAgentNode';
import { SQLAgentNode } from '@/components/workflows/nodes/SQLAgentNode';
import { VisionAgentNode } from '@/components/workflows/nodes/VisionAgentNode';
import { ResponseNode } from '@/components/workflows/nodes/ResponseNode';
import { ConditionNode } from '@/components/workflows/nodes/ConditionNode';
import { ApprovalNode } from '@/components/workflows/nodes/ApprovalNode';

const nodeTypes = {
  userInput: UserInputNode,
  plannerAgent: PlannerAgentNode,
  ragAgent: RAGAgentNode,
  sqlAgent: SQLAgentNode,
  visionAgent: VisionAgentNode,
  response: ResponseNode,
  condition: ConditionNode,
  approval: ApprovalNode,
};

const edgeTypes = {
  default: DefaultEdge,
  conditional: ConditionalEdge,
  approval: ApprovalEdge,
};

interface BuilderCanvasProps {
  nodes: Node[];
  edges: Edge[];
  onNodesChange: OnNodesChange;
  onEdgesChange: OnEdgesChange;
  onConnect: OnConnect;
  onNodeClick: (event: React.MouseEvent, node: Node) => void;
  onEdgeClick: (event: React.MouseEvent, edge: Edge) => void;
}

export function BuilderCanvas({
  nodes, edges, onNodesChange, onEdgesChange, onConnect,
  onNodeClick, onEdgeClick,
}: BuilderCanvasProps) {
  return (
    <div className="flex-1">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onNodeClick={onNodeClick}
        onEdgeClick={onEdgeClick}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        snapToGrid
        snapGrid={[15, 15]}
        defaultEdgeOptions={{
          type: 'default',
          animated: true,
        }}
      >
        <Controls position="bottom-left" />
        <MiniMap
          position="bottom-right"
          nodeStrokeWidth={3}
          zoomable
          pannable
        />
        <Background variant="dots" gap={15} size={1} />
      </ReactFlow>
    </div>
  );
}
```

### Node Data Structure

```typescript
interface WorkflowNodeData {
  label: string;
  type: string;
  config: Record<string, any>;
  status?: 'idle' | 'running' | 'completed' | 'error';
  output?: any;
  error?: string;
}
```

---

## Node Types

### Available Node Types

| Node Type       | Color   | Icon          | Description                              | Outputs |
|-----------------|---------|---------------|------------------------------------------|---------|
| User Input      | Blue    | MessageCircle | Captures user input/prompt               | 1       |
| Planner Agent   | Purple  | Brain         | Plans steps to answer complex queries    | 1+      |
| RAG Agent       | Green   | FileSearch    | Searches knowledge base for context      | 1       |
| SQL Agent       | Orange  | Database      | Queries database for structured data     | 1       |
| Vision Agent    | Pink    | Eye           | Analyzes images and visual content       | 1       |
| Response        | Teal    | Send          | Formats and sends final response         | 0       |
| Condition       | Yellow  | GitBranch     | Routes based on conditions               | 2+      |
| Approval        | Red     | ShieldCheck   | Pauses for human approval                | 2       |

### Node Component Templates

#### User Input Node

```tsx
// components/workflows/nodes/UserInputNode.tsx
import { Handle, Position } from 'reactflow';
import { Card } from '@/components/ui/card';
import { MessageCircle } from 'lucide-react';

interface UserInputNodeProps {
  data: WorkflowNodeData;
  selected?: boolean;
}

export function UserInputNode({ data, selected }: UserInputNodeProps) {
  return (
    <Card className={`min-w-[200px] ${selected ? 'ring-2 ring-primary' : ''}`}>
      <div className="flex items-center gap-2 p-3 border-b bg-blue-50">
        <MessageCircle className="w-4 h-4 text-blue-600" />
        <span className="text-sm font-medium">User Input</span>
        <StatusIndicator status={data.status} />
      </div>
      <div className="p-3 text-sm text-muted-foreground">
        <p className="truncate">{data.config.label || data.label}</p>
        {data.config.input_type && (
          <p className="text-xs mt-1">Type: {data.config.input_type}</p>
        )}
      </div>
      <Handle type="source" position={Position.Bottom} className="w-3 h-3" />
    </Card>
  );
}
```

#### Planner Agent Node

```tsx
// components/workflows/nodes/PlannerAgentNode.tsx
import { Handle, Position } from 'reactflow';
import { Card } from '@/components/ui/card';
import { Brain } from 'lucide-react';

export function PlannerAgentNode({ data, selected }: UserInputNodeProps) {
  return (
    <Card className={`min-w-[220px] ${selected ? 'ring-2 ring-primary' : ''}`}>
      <div className="flex items-center gap-2 p-3 border-b bg-purple-50">
        <Brain className="w-4 h-4 text-purple-600" />
        <span className="text-sm font-medium">Planner Agent</span>
        <StatusIndicator status={data.status} />
      </div>
      <div className="p-3 space-y-2 text-sm">
        <p className="font-medium">{data.config.agent_name || 'Default Planner'}</p>
        <div className="text-xs text-muted-foreground space-y-1">
          <p>Model: {data.config.model || 'Llama 3 70B'}</p>
          <p>Max Steps: {data.config.max_steps || 5}</p>
          <p>Temperature: {data.config.temperature || 0.7}</p>
        </div>
      </div>
      <Handle type="target" position={Position.Top} className="w-3 h-3" />
      <Handle type="source" position={Position.Bottom} id="default" className="w-3 h-3" />
    </Card>
  );
}
```

#### Condition Node

```tsx
// components/workflows/nodes/ConditionNode.tsx
import { Handle, Position } from 'reactflow';
import { Card } from '@/components/ui/card';
import { GitBranch } from 'lucide-react';

export function ConditionNode({ data, selected }: UserInputNodeProps) {
  const branches = data.config.branches || [{ label: 'Yes' }, { label: 'No' }];

  return (
    <Card className={`min-w-[200px] ${selected ? 'ring-2 ring-primary' : ''}`}>
      <div className="flex items-center gap-2 p-3 border-b bg-yellow-50">
        <GitBranch className="w-4 h-4 text-yellow-600" />
        <span className="text-sm font-medium">Condition</span>
        <StatusIndicator status={data.status} />
      </div>
      <div className="p-3 text-sm">
        <p className="text-xs text-muted-foreground mb-2">Expression:</p>
        <code className="text-xs bg-muted p-1 rounded block">
          {data.config.expression || 'input.contains("error")'}
        </code>
      </div>
      <Handle type="target" position={Position.Top} className="w-3 h-3" />
      {branches.map((branch: any, idx: number) => (
        <Handle
          key={branch.label}
          type="source"
          position={Position.Bottom}
          id={branch.label}
          style={{ left: `${((idx + 1) / (branches.length + 1)) * 100}%` }}
          className="w-3 h-3"
        />
      ))}
    </Card>
  );
}
```

#### Approval Node

```tsx
// components/workflows/nodes/ApprovalNode.tsx
import { Handle, Position } from 'reactflow';
import { Card } from '@/components/ui/card';
import { ShieldCheck } from 'lucide-react';

export function ApprovalNode({ data, selected }: UserInputNodeProps) {
  return (
    <Card className={`min-w-[200px] ${selected ? 'ring-2 ring-primary' : ''}`}>
      <div className="flex items-center gap-2 p-3 border-b bg-red-50">
        <ShieldCheck className="w-4 h-4 text-red-600" />
        <span className="text-sm font-medium">Approval Required</span>
        <StatusIndicator status={data.status} />
      </div>
      <div className="p-3 text-sm space-y-1">
        <p className="font-medium">{data.config.approval_type || 'Manual Review'}</p>
        <p className="text-xs text-muted-foreground">
          Approvers: {data.config.approvers?.join(', ') || 'Admin'}
        </p>
        <p className="text-xs text-muted-foreground">
          Timeout: {data.config.timeout || '24h'}
        </p>
      </div>
      <Handle type="target" position={Position.Top} className="w-3 h-3" />
      <Handle type="source" position={Position.Bottom} id="approve"
        style={{ left: '30%' }} className="w-3 h-3" />
      <Handle type="source" position={Position.Bottom} id="reject"
        style={{ left: '70%' }} className="w-3 h-3" />
    </Card>
  );
}
```

---

## Node Configuration Panel

### Panel Layout

```
+------------------------------+
|  Node Properties             |
|  Type: User Input            |
+------------------------------+
|                              |
|  Label:                      |
|  [Customer Question     ]    |
|                              |
|  Input Type:                 |
|  [Text Input v]              |
|                              |
|  Required:                   |
|  [x] Yes                     |
|                              |
|  Placeholder:                |
|  [Ask a question about...]   |
|                              |
|  Max Length:                 |
|  [2000]                      |
|                              |
|  Validation:                 |
|  [ ] Min 10 chars           |
|  [ ] No profanity           |
|  [ ] No injection attempts  |
|                              |
|  Advanced:                   |
|  [Show]                      |
|                              |
+------------------------------+
|  [Delete Node]               |
+------------------------------+
```

### Configuration Forms by Node Type

| Node Type       | Key Properties                                          |
|-----------------|---------------------------------------------------------|
| User Input      | label, input_type, placeholder, required, max_length    |
| Planner Agent   | agent_name, model, max_steps, temperature, system_prompt|
| RAG Agent       | knowledge_set, search_mode, top_k, score_threshold     |
| SQL Agent       | database_connection, max_rows, timeout, read_only       |
| Vision Agent    | vision_model, supported_formats, max_size_mb            |
| Response        | response_format, template, include_sources              |
| Condition       | expression, branches[], default_branch                  |
| Approval        | approval_type, approvers[], timeout, on_timeout_action  |

### Config Panel Component

```tsx
// components/workflows/NodeProperties.tsx
interface NodePropertiesProps {
  node: Node;
  onUpdate: (nodeId: string, data: Partial<WorkflowNodeData>) => void;
  onDelete: (nodeId: string) => void;
}

export function NodeProperties({ node, onUpdate, onDelete }: NodePropertiesProps) {
  const [config, setConfig] = useState(node.data.config);

  const handleSave = () => {
    onUpdate(node.id, { config });
  };

  return (
    <div className="space-y-4 p-4">
      <div className="flex items-center justify-between">
        <h3 className="font-semibold">{node.data.type}</h3>
        <Badge>{node.type}</Badge>
      </div>

      <div>
        <Label>Label</Label>
        <Input
          value={node.data.label}
          onChange={(e) => onUpdate(node.id, { label: e.target.value })}
        />
      </div>

      {renderConfigFields(node.type, config, setConfig)}

      <div className="pt-4 border-t">
        <Button variant="destructive" size="sm" onClick={() => onDelete(node.id)}>
          Delete Node
        </Button>
      </div>
    </div>
  );
}

function renderConfigFields(type: string, config: any, setConfig: any) {
  switch (type) {
    case 'userInput':
      return <UserInputConfig config={config} onChange={setConfig} />;
    case 'plannerAgent':
      return <PlannerAgentConfig config={config} onChange={setConfig} />;
    case 'ragAgent':
      return <RAGAgentConfig config={config} onChange={setConfig} />;
    case 'condition':
      return <ConditionConfig config={config} onChange={setConfig} />;
    case 'approval':
      return <ApprovalConfig config={config} onChange={setConfig} />;
    default:
      return <GenericConfig config={config} onChange={setConfig} />;
  }
}
```

---

## Edge Connections

### Edge Types

| Edge Type    | Style          | Use Case                         |
|--------------|----------------|----------------------------------|
| Default      | Solid, animated| Standard data flow               |
| Conditional  | Dashed, colored| Condition branch output          |
| Approval     | Solid, orange  | Approval wait → approve/reject   |
| Error        | Red, dashed    | Error handling path              |
| Fallback     | Gray, dotted   | Fallback/retry path              |

### Edge Validation

```typescript
// lib/workflows/edgeValidation.ts
import { Connection } from 'reactflow';

interface ValidationResult {
  valid: boolean;
  error?: string;
}

export function validateEdge(connection: Connection): ValidationResult {
  const { source, target, sourceHandle, targetHandle } = connection;

  if (!source || !target) {
    return { valid: false, error: 'Source and target required' };
  }

  if (source === target) {
    return { valid: false, error: 'Cannot connect node to itself' };
  }

  // Prevent cycles (simplified - full implementation would use DFS)
  if (wouldCreateCycle(source, target)) {
    return { valid: false, error: 'Connection would create a cycle' };
  }

  // Validate handle types
  const sourceNode = getNode(source);
  const targetNode = getNode(target);

  if (sourceNode?.type === 'response' && target) {
    return { valid: false, error: 'Response nodes cannot have outgoing connections' };
  }

  return { valid: true };
}
```

### Edge Label Component

```tsx
// components/workflows/edges/ConditionalEdge.tsx
import { BaseEdge, EdgeProps, getBezierPath } from 'reactflow';

export function ConditionalEdge({
  id, sourceX, sourceY, targetX, targetY,
  sourcePosition, targetPosition, data, style, markerEnd,
}: EdgeProps) {
  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX, sourceY, targetX, targetY,
    sourcePosition, targetPosition,
  });

  return (
    <>
      <BaseEdge path={edgePath} markerEnd={markerEnd} style={{
        ...style,
        stroke: data?.condition ? '#eab308' : '#94a3b8',
        strokeDasharray: '5 5',
      }} />
      {data?.label && (
        <foreignObject x={labelX - 50} y={labelY - 12} width={100} height={24}>
          <div className="bg-background border rounded px-2 py-0.5 text-xs text-center">
            {data.label}
          </div>
        </foreignObject>
      )}
    </>
  );
}
```

---

## Drag-and-Drop Node Creation

### Node Palette

```
+------------------+
| Node Palette     |
| [Search nodes..] |
+------------------+
| INPUT            |
| [o] User Input   |
|                  |
| AGENTS           |
| [o] Planner      |
| [o] RAG Agent    |
| [o] SQL Agent    |
| [o] Vision Agent |
|                  |
| FLOW             |
| [o] Condition    |
| [o] Approval     |
|                  |
| OUTPUT           |
| [o] Response     |
+------------------+
```

### Palette Component

```tsx
// components/workflows/NodePalette.tsx
import { useDrag } from 'react-dnd';

const NODE_CATEGORIES = [
  {
    category: 'Input',
    nodes: [
      { type: 'userInput', label: 'User Input', icon: MessageCircle, color: 'blue' },
    ],
  },
  {
    category: 'Agents',
    nodes: [
      { type: 'plannerAgent', label: 'Planner Agent', icon: Brain, color: 'purple' },
      { type: 'ragAgent', label: 'RAG Agent', icon: FileSearch, color: 'green' },
      { type: 'sqlAgent', label: 'SQL Agent', icon: Database, color: 'orange' },
      { type: 'visionAgent', label: 'Vision Agent', icon: Eye, color: 'pink' },
    ],
  },
  {
    category: 'Flow',
    nodes: [
      { type: 'condition', label: 'Condition', icon: GitBranch, color: 'yellow' },
      { type: 'approval', label: 'Approval', icon: ShieldCheck, color: 'red' },
    ],
  },
  {
    category: 'Output',
    nodes: [
      { type: 'response', label: 'Response', icon: Send, color: 'teal' },
    ],
  },
];

function DraggableNode({ nodeConfig }: { nodeConfig: any }) {
  const [{ isDragging }, drag] = useDrag({
    type: 'node',
    item: { type: nodeConfig.type },
    collect: (monitor) => ({ isDragging: monitor.isDragging() }),
  });

  return (
    <div
      ref={drag}
      className={`flex items-center gap-2 p-2 rounded-lg cursor-grab
        hover:bg-muted transition-colors
        ${isDragging ? 'opacity-50' : ''}`}
    >
      <nodeConfig.icon className={`w-4 h-4 text-${nodeConfig.color}-600`} />
      <span className="text-sm">{nodeConfig.label}</span>
    </div>
  );
}

export function NodePalette() {
  const [search, setSearch] = useState('');

  const filteredCategories = NODE_CATEGORIES.map((cat) => ({
    ...cat,
    nodes: cat.nodes.filter((n) =>
      n.label.toLowerCase().includes(search.toLowerCase())
    ),
  })).filter((cat) => cat.nodes.length > 0);

  return (
    <aside className="w-56 border-r bg-card p-3 overflow-y-auto">
      <h3 className="text-sm font-semibold mb-3">Node Palette</h3>
      <Input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search nodes..."
        className="mb-3"
      />
      {filteredCategories.map((cat) => (
        <div key={cat.category} className="mb-4">
          <h4 className="text-xs font-medium text-muted-foreground uppercase mb-2">
            {cat.category}
          </h4>
          <div className="space-y-1">
            {cat.nodes.map((node) => (
              <DraggableNode key={node.type} nodeConfig={node} />
            ))}
          </div>
        </div>
      ))}
    </aside>
  );
}
```

---

## Canvas Controls

| Control     | Keyboard | Description                    |
|-------------|----------|--------------------------------|
| Zoom In     | `+`      | Zoom into canvas               |
| Zoom Out    | `-`      | Zoom out from canvas           |
| Fit View    | `Shift+F`| Fit all nodes in viewport      |
| Delete      | `Del`    | Delete selected node/edge      |
| Copy        | `Ctrl+C` | Copy selected node             |
| Paste       | `Ctrl+V` | Paste copied node              |
| Undo        | `Ctrl+Z` | Undo last action               |
| Redo        | `Ctrl+Y` | Redo last undone action        |
| Select All  | `Ctrl+A` | Select all nodes               |
| Pan         | `Space+Drag` | Pan canvas                |
| Mini Map    | `M`      | Toggle minimap                 |

---

## Workflow Properties

```
+------------------------------+
|  Workflow Properties         |
+------------------------------+
|                              |
|  Name:                       |
|  [Customer Support Pipeline] |
|                              |
|  Description:                |
|  [Multi-step customer       ]|
|  [support with RAG, SQL,   ]|
|  [and human approval.     ] |
|                              |
|  Trigger:                    |
|  [API Call v]                |
|                              |
|  Timeout:                    |
|  [300] seconds               |
|                              |
|  Max Retries:                |
|  [3]                         |
|                              |
|  Tags:                       |
|  [support] [customer] [+ ]   |
|                              |
|  Visibility:                 |
|  ( ) Private                 |
|  (X) Organization            |
|  ( ) Public                  |
|                              |
+------------------------------+
|  [Save] [Cancel]             |
+------------------------------+
```

---

## Workflow Triggers

| Trigger Type | Description                              | Configuration             |
|--------------|------------------------------------------|---------------------------|
| Manual       | User clicks "Run" button                 | None                      |
| API Call     | External system calls workflow endpoint   | API key, rate limit       |
| Schedule     | Cron-like scheduled execution             | Cron expression           |
| Event        | Triggered by system event                 | Event type, filters       |
| Webhook      | HTTP webhook with payload                 | URL, secret, method       |
| Chat Message | Triggered by chat message pattern         | Pattern match, channel    |

---

## Workflow Execution View

```
+----------------------------------------------------------------------+
|  Execution: exec_abc123                    Status: Running           |
|  Workflow: Customer Support Pipeline                                 |
|  Started: Jul 16, 2026 10:30:00 AM    Duration: 12.5s              |
+----------------------------------------------------------------------+
|                                                                      |
|  Step Progress:                                                      |
|  [User Input] -> [Planner] -> [RAG Agent] -> [Condition] -> [...]   |
|       OK           OK          RUNNING       --                      |
|                                                                      |
|  +----------------------------------------------------------------+  |
|  |  Step 1: User Input                              OK (0.2s)     |  |
|  |  +----------------------------------------------------------+  |  |
|  |  | Input: "What was our Q2 revenue and how does it compare  |  |  |
|  |  |         to Q1?"                                          |  |  |
|  |  +----------------------------------------------------------+  |  |
|  |                                                                 |  |
|  |  Step 2: Planner Agent                        OK (2.1s)       |  |
|  |  +----------------------------------------------------------+  |  |
|  |  | Plan:                                                    |  |  |
|  |  | 1. Search knowledge base for Q2 revenue data            |  |  |
|  |  | 2. Query database for Q1 revenue comparison             |  |  |
|  |  | 3. Compare and summarize results                        |  |  |
|  |  +----------------------------------------------------------+  |  |
|  |                                                                 |  |
|  |  Step 3: RAG Agent                             RUNNING (8.2s) |  |
|  |  +----------------------------------------------------------+  |  |
|  |  | Searching: "Q2 revenue report 2026"                     |  |  |
|  |  | Found 3 relevant chunks                                  |  |  |
|  |  | Generating response...                                   |  |  |
|  |  +----------------------------------------------------------+  |  |
|  |                                                                 |  |
|  |  Step 4: Condition                             -- (pending)    |  |
|  |  Step 5: SQL Agent                             -- (pending)    |  |
|  |  Step 6: Response                              -- (pending)    |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Cancel Execution]  [View Logs]  [Retry from Step]                  |
+----------------------------------------------------------------------+
```

---

## Workflow Execution History

```
+----------------------------------------------------------------------+
|  Execution History                                                   |
+----------------------------------------------------------------------+
|                                                                      |
|  Execution ID   | Status   | Started           | Duration | Steps   |
|  ---------------+----------+-------------------+----------+---------|
|  exec_abc123    | Running  | Jul 16 10:30 AM   | 12.5s    | 3/6     |
|  exec_def456    | Success  | Jul 16 09:15 AM   | 8.2s     | 6/6     |
|  exec_ghi789    | Failed   | Jul 16 08:00 AM   | 15.0s    | 4/6     |
|  exec_jkl012    | Success  | Jul 15 16:30 PM   | 6.5s     | 6/6     |
|  exec_mno345    | Approved | Jul 15 14:00 PM   | 120.0s   | 6/6     |
|  exec_pqr678    | Success  | Jul 15 11:00 AM   | 7.8s     | 6/6     |
|                                                                      |
|  Showing 1-10 of 234 executions                                     |
+----------------------------------------------------------------------+
```

### Execution Statuses

| Status    | Description                                   |
|-----------|-----------------------------------------------|
| Running   | Currently executing                           |
| Success   | Completed successfully                        |
| Failed    | Failed at some step                           |
| Pending   | Waiting to start                              |
| Approved  | Completed with approval step                  |
| Rejected  | Rejected at approval step                     |
| Cancelled | Manually cancelled by user                    |
| Timed Out | Exceeded workflow timeout                     |

---

## Workflow Approval Step

### Approval Dialog

```
+----------------------------------------------------------------------+
|  Approval Required                                                   |
+----------------------------------------------------------------------+
|                                                                      |
|  Workflow: Customer Support Pipeline                                 |
|  Execution: exec_mno345                                              |
|  Step: Approval (Step 4 of 6)                                        |
|                                                                      |
|  Requested by: Planner Agent                                         |
|  Time: Jul 15, 2026 2:00 PM                                         |
|                                                                      |
|  Context:                                                            |
|  +----------------------------------------------------------------+  |
|  |  The AI suggests sending the following response to the        |  |
|  |  customer:                                                    |  |
|  |                                                               |  |
|  |  "Based on our Q2 financial report, revenue was $4.2M,       |  |
|  |   a 15% increase from Q1's $3.65M. The growth was driven     |  |
|  |   primarily by enterprise segment expansion..."               |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Your Comment:                                                       |
|  [________________________________________________________]         |
|  [________________________________________________________]         |
|                                                                      |
|  [Reject]  [Request Changes]  [Approve]                              |
+----------------------------------------------------------------------+
```

### Approval Component

```tsx
// components/workflows/ApprovalDialog.tsx
interface ApprovalDialogProps {
  execution: WorkflowExecution;
  step: ExecutionStep;
  onApprove: (comment: string) => Promise<void>;
  onReject: (reason: string) => Promise<void>;
}

export function ApprovalDialog({ execution, step, onApprove, onReject }: ApprovalDialogProps) {
  const [comment, setComment] = useState('');
  const [action, setAction] = useState<'approve' | 'reject'>('approve');

  return (
    <Dialog open onOpenChange={() => {}}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Approval Required</DialogTitle>
          <DialogDescription>
            Workflow: {execution.workflow_name} | Step {step.index + 1} of {execution.total_steps}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="border rounded-lg p-4 bg-muted/50">
            <h4 className="font-medium text-sm mb-2">Context</h4>
            <pre className="text-sm whitespace-pre-wrap">{step.context}</pre>
          </div>

          <div>
            <Label>{action === 'approve' ? 'Comment (optional)' : 'Rejection Reason'}</Label>
            <Textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder={action === 'approve' ? 'Add a comment...' : 'Explain why...'}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="destructive" onClick={() => onReject(comment)}>
            Reject
          </Button>
          <Button variant="outline" onClick={() => { setAction('reject'); }}>
            Request Changes
          </Button>
          <Button onClick={() => onApprove(comment)}>
            Approve
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

---

## Workflow Error Handling

### Error Handling Strategies

| Strategy    | Description                                    |
|-------------|------------------------------------------------|
| Retry       | Retry failed step with backoff                 |
| Skip        | Skip failed step and continue                  |
| Fallback    | Execute alternative path                        |
| Rollback    | Undo completed steps and abort                  |
| Alert       | Send notification and wait for manual action   |
| Abort       | Stop workflow immediately                       |

### Failed Step UI

```
+----------------------------------------------------------------------+
|  Step Failed: RAG Agent (Step 3 of 6)                                |
+----------------------------------------------------------------------+
|                                                                      |
|  Error: Knowledge base "finance-docs" not found                      |
|  Time: Jul 16, 2026 10:30:15 AM                                     |
|  Duration: 8.2s                                                      |
|                                                                      |
|  Input received:                                                     |
|  "Search for Q2 revenue data"                                        |
|                                                                      |
|  Error Details:                                                      |
|  +----------------------------------------------------------------+  |
|  |  {                                                              |  |
|  |    "error": "KNOWLEDGE_SET_NOT_FOUND",                         |  |
|  |    "message": "Knowledge set 'finance-docs' does not exist",   |  |
|  |    "set_id": "finance-docs",                                    |  |
|  |    "suggestion": "Create the set or check the name"             |  |
|  |  }                                                              |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Actions:                                                            |
|  [Retry Step]  [Skip & Continue]  [Try Fallback]  [Abort Workflow]   |
+----------------------------------------------------------------------+
```

### Error Handler Component

```tsx
// components/workflows/StepErrorHandler.tsx
interface StepErrorHandlerProps {
  step: ExecutionStep;
  workflow: Workflow;
  onRetry: () => Promise<void>;
  onSkip: () => Promise<void>;
  onFallback: (fallbackPath: string) => Promise<void>;
  onAbort: () => Promise<void>;
}

export function StepErrorHandler({
  step, workflow, onRetry, onSkip, onFallback, onAbort,
}: StepErrorHandlerProps) {
  return (
    <Card className="border-destructive">
      <CardHeader>
        <CardTitle className="text-destructive flex items-center gap-2">
          <AlertCircle className="w-5 h-5" />
          Step Failed: {step.name}
        </CardTitle>
        <CardDescription>
          Error at step {step.index + 1} of {workflow.total_steps}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="bg-muted p-3 rounded-lg">
          <p className="text-sm font-medium">{step.error?.message}</p>
          {step.error?.details && (
            <pre className="text-xs mt-2 overflow-auto max-h-32">
              {JSON.stringify(step.error.details, null, 2)}
            </pre>
          )}
        </div>
        <div className="flex gap-2">
          <Button onClick={onRetry} variant="default">Retry Step</Button>
          <Button onClick={onSkip} variant="outline">Skip & Continue</Button>
          <Button onClick={onAbort} variant="destructive">Abort Workflow</Button>
        </div>
      </CardContent>
    </Card>
  );
}
```

---

## Workflow Testing

### Test Mode

```
+----------------------------------------------------------------------+
|  Test Mode: Customer Support Pipeline                                |
+----------------------------------------------------------------------+
|                                                                      |
|  Mock Data:                                                          |
|  +----------------------------------------------------------------+  |
|  |  User Input: "What was our Q2 revenue?"                        |  |
|  |  Knowledge Base: [finance-docs v] (mock)                        |  |
|  |  Database: [test-db v] (mock)                                   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Run Test]  [Step Through]  [Dry Run]                               |
+----------------------------------------------------------------------+
|                                                                      |
|  Test Results:                                                       |
|  +----------------------------------------------------------------+  |
|  |  Step 1: User Input                        PASS (0.1s)        |  |
|  |  Step 2: Planner Agent                      PASS (1.8s)        |  |
|  |  Step 3: RAG Agent                          PASS (2.1s)        |  |
|  |  Step 4: Condition (branch: "has_data")     PASS (0.1s)        |  |
|  |  Step 5: SQL Agent                          PASS (1.5s)        |  |
|  |  Step 6: Response                           PASS (0.3s)        |  |
|  |                                                                |  |
|  |  Total: 6/6 passed  |  Time: 5.9s  |  Cost: ~$0.002          |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Output:                                                             |
|  +----------------------------------------------------------------+  |
|  |  "Based on our Q2 financial data, revenue reached $4.2M..."   |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

---

## Workflow Templates

### Template Gallery

| Template                | Description                    | Nodes | Difficulty |
|-------------------------|--------------------------------|-------|------------|
| Customer Support        | Multi-agent customer support   | 6     | Medium     |
| Data Analysis Pipeline  | SQL + RAG for data insights    | 4     | Easy       |
| Content Generator       | Multi-step content creation    | 5     | Medium     |
| Document Review         | RAG + approval for doc review  | 5     | Medium     |
| Image Analysis          | Vision + response pipeline     | 3     | Easy       |
| Complex Research        | Planner + multiple agents      | 8     | Hard       |
| Code Review             | Vision + SQL + response        | 6     | Medium     |
| Executive Summary       | RAG + SQL + approval           | 6     | Medium     |

---

## Workflow Import/Export

### Export Format (JSON)

```json
{
  "name": "Customer Support Pipeline",
  "description": "Multi-step customer support workflow",
  "version": "1.2",
  "nodes": [
    {
      "id": "node_1",
      "type": "userInput",
      "position": { "x": 250, "y": 0 },
      "data": {
        "label": "Customer Question",
        "config": {
          "input_type": "text",
          "placeholder": "Ask a question...",
          "required": true
        }
      }
    },
    {
      "id": "node_2",
      "type": "plannerAgent",
      "position": { "x": 250, "y": 150 },
      "data": {
        "label": "Plan Response",
        "config": {
          "model": "llama-3-70b",
          "max_steps": 5,
          "temperature": 0.7
        }
      }
    }
  ],
  "edges": [
    {
      "id": "edge_1",
      "source": "node_1",
      "target": "node_2",
      "type": "default"
    }
  ],
  "trigger": {
    "type": "api",
    "config": {}
  }
}
```

### Import/Export Components

```tsx
// components/workflows/WorkflowImportExport.tsx
export function WorkflowImportExport() {
  const handleExport = () => {
    const data = serializeWorkflow(workflow);
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${workflow.name.replace(/\s+/g, '-').toLowerCase()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleImport = (file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = JSON.parse(e.target?.result as string);
        const validation = validateWorkflowImport(data);
        if (validation.valid) {
          loadWorkflow(data);
        } else {
          toast({ title: 'Import failed', description: validation.error, variant: 'destructive' });
        }
      } catch {
        toast({ title: 'Invalid JSON file', variant: 'destructive' });
      }
    };
    reader.readAsText(file);
  };

  return (
    <div className="flex gap-2">
      <Button variant="outline" onClick={handleExport}>Export JSON</Button>
      <label className="cursor-pointer">
        <Button variant="outline" asChild>
          <span>Import JSON</span>
        </Button>
        <input type="file" accept=".json" className="hidden" onChange={(e) => {
          if (e.target.files?.[0]) handleImport(e.target.files[0]);
        }} />
      </label>
    </div>
  );
}
```

---

## Workflow Versioning

### Version History

```
+----------------------------------------------------------------------+
|  Version History: Customer Support Pipeline                          |
+----------------------------------------------------------------------+
|                                                                      |
|  Current: v1.2 (Draft)                                               |
|                                                                      |
|  +----------------------------------------------------------------+  |
|  |  Version | Date       | Author         | Changes              |  |
|  |----------+------------+----------------+----------------------|  |
|  |  v1.2    | Jul 16     | admin          | Added approval step  |  |
|  |  v1.1    | Jul 15     | john@acme.com  | Added SQL agent      |  |
|  |  v1.0    | Jul 14     | admin          | Initial workflow     |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Compare Versions]  [Rollback to v1.1]  [Publish v1.2]             |
+----------------------------------------------------------------------+
```

### Version Diff View

```
+----------------------------------------------------------------------+
|  Comparing v1.1 vs v1.2                                              |
+----------------------------------------------------------------------+
|                                                                      |
|  +-------------------+  +-------------------+                        |
|  |  v1.1 (Current)   |  |  v1.2 (Draft)     |                        |
|  +-------------------+  +-------------------+                        |
|  |  Nodes: 5          |  |  Nodes: 6 (+1)    |                        |
|  |  Edges: 4          |  |  Edges: 6 (+2)    |                        |
|  |                    |  |                    |                        |
|  |  [User Input]      |  |  [User Input]      |                        |
|  |       |            |  |       |            |                        |
|  |  [Planner]         |  |  [Planner]         |                        |
|  |    /    \          |  |    /    \          |                        |
|  | [RAG]  [SQL]       |  | [RAG]  [SQL]       |                        |
|  |   |       |        |  |   |       |        |                        |
|  | [Response]         |  | [Condition]  <-- NEW                       |
|  +-------------------+  |  /        \        |                        |
|                          | [Approval] [Response]                      |
|                          +-------------------+                        |
+----------------------------------------------------------------------+
```

---

## API Integration

### Endpoints

| Method   | Endpoint                        | Description                  |
|----------|---------------------------------|------------------------------|
| `GET`    | `/api/workflows`                | List workflows               |
| `POST`   | `/api/workflows`                | Create workflow              |
| `GET`    | `/api/workflows/:id`            | Get workflow details         |
| `PATCH`  | `/api/workflows/:id`            | Update workflow              |
| `DELETE` | `/api/workflows/:id`            | Delete workflow              |
| `POST`   | `/api/workflows/:id/run`        | Start workflow execution     |
| `GET`    | `/api/workflows/:id/versions`   | List workflow versions       |
| `POST`   | `/api/workflows/:id/versions`   | Save new version             |
| `POST`   | `/api/workflows/:id/rollback`   | Rollback to version          |
| `GET`    | `/api/workflows/executions`     | List all executions          |
| `GET`    | `/api/workflows/executions/:id` | Get execution details        |
| `POST`   | `/api/workflows/executions/:id/approve` | Approve step    |
| `POST`   | `/api/workflows/executions/:id/reject`  | Reject step     |
| `POST`   | `/api/workflows/executions/:id/cancel`  | Cancel execution|
| `POST`   | `/api/workflows/:id/test`       | Run in test mode             |
| `POST`   | `/api/workflows/import`         | Import workflow from JSON    |
| `GET`    | `/api/workflows/:id/export`     | Export workflow as JSON      |
| `GET`    | `/api/workflows/templates`      | List workflow templates      |

---

## Hooks

### useWorkflows

```typescript
// hooks/workflows/useWorkflows.ts
export function useWorkflows(filters?: WorkflowFilters) {
  const queryClient = useQueryClient();

  const workflows = useQuery({
    queryKey: ['workflows', filters],
    queryFn: () => api.get('/workflows', { params: filters }),
  });

  const createWorkflow = useMutation({
    mutationFn: (data: CreateWorkflowRequest) => api.post('/workflows', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['workflows'] }),
  });

  const updateWorkflow = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Workflow> }) =>
      api.patch(`/workflows/${id}`, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['workflows'] }),
  });

  const deleteWorkflow = useMutation({
    mutationFn: (id: string) => api.delete(`/workflows/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['workflows'] }),
  });

  return { workflows, createWorkflow, updateWorkflow, deleteWorkflow };
}
```

### useWorkflowExecution

```typescript
// hooks/workflows/useWorkflowExecution.ts
export function useWorkflowExecution(workflowId: string) {
  const executions = useQuery({
    queryKey: ['workflow-executions', workflowId],
    queryFn: () => api.get(`/workflows/${workflowId}/executions`),
  });

  const runWorkflow = useMutation({
    mutationFn: (input: any) => api.post(`/workflows/${workflowId}/run`, { input }),
  });

  return { executions, runWorkflow };
}
```

### useWorkflowBuilder

```typescript
// hooks/workflows/useWorkflowBuilder.ts
export function useWorkflowBuilder() {
  const { workflowId } = useParams();
  const { nodes, setNodes, onNodesChange } = useNodesState([]);
  const { edges, setEdges, onEdgesChange, addEdge } = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null);

  const { data: workflow } = useQuery({
    queryKey: ['workflow', workflowId],
    queryFn: () => api.get(`/workflows/${workflowId}`),
    enabled: !!workflowId,
  });

  useEffect(() => {
    if (workflow) {
      setNodes(workflow.nodes);
      setEdges(workflow.edges);
    }
  }, [workflow]);

  const onConnect = useCallback((params: Connection) => {
    const validation = validateEdge(params);
    if (validation.valid) {
      addEdge({ ...params, type: 'default', animated: true });
    } else {
      toast({ title: validation.error, variant: 'destructive' });
    }
  }, [addEdge]);

  const onNodeClick = useCallback((_: any, node: Node) => {
    setSelectedNode(node);
    setSelectedEdge(null);
  }, []);

  const onEdgeClick = useCallback((_: any, edge: Edge) => {
    setSelectedEdge(edge);
    setSelectedNode(null);
  }, []);

  return {
    workflow, nodes, edges, onNodesChange, onEdgesChange,
    onConnect, onNodeClick, onEdgeClick, selectedNode, selectedEdge,
  };
}
```

---

## Stores

### Workflow Builder Store (Zustand)

```typescript
// stores/workflows/builder.ts
import { create } from 'zustand';

interface BuilderState {
  nodes: Node[];
  edges: Edge[];
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  isDirty: boolean;
  history: { nodes: Node[]; edges: Edge[] }[];
  historyIndex: number;

  setNodes: (nodes: Node[]) => void;
  setEdges: (edges: Edge[]) => void;
  addNode: (node: Node) => void;
  updateNodeData: (nodeId: string, data: Partial<WorkflowNodeData>) => void;
  removeNode: (nodeId: string) => void;
  addEdge: (edge: Edge) => void;
  removeEdge: (edgeId: string) => void;
  selectNode: (nodeId: string | null) => void;
  selectEdge: (edgeId: string | null) => void;
  undo: () => void;
  redo: () => void;
  saveSnapshot: () => void;
  markDirty: () => void;
  markClean: () => void;
}

export const useBuilderStore = create<BuilderState>((set, get) => ({
  nodes: [],
  edges: [],
  selectedNodeId: null,
  selectedEdgeId: null,
  isDirty: false,
  history: [],
  historyIndex: -1,

  setNodes: (nodes) => set({ nodes, isDirty: true }),
  setEdges: (edges) => set({ edges, isDirty: true }),

  addNode: (node) => set((state) => ({
    nodes: [...state.nodes, node],
    isDirty: true,
  })),

  updateNodeData: (nodeId, data) => set((state) => ({
    nodes: state.nodes.map((n) =>
      n.id === nodeId ? { ...n, data: { ...n.data, ...data } } : n
    ),
    isDirty: true,
  })),

  removeNode: (nodeId) => set((state) => ({
    nodes: state.nodes.filter((n) => n.id !== nodeId),
    edges: state.edges.filter((e) => e.source !== nodeId && e.target !== nodeId),
    selectedNodeId: state.selectedNodeId === nodeId ? null : state.selectedNodeId,
    isDirty: true,
  })),

  addEdge: (edge) => set((state) => ({
    edges: [...state.edges, edge],
    isDirty: true,
  })),

  removeEdge: (edgeId) => set((state) => ({
    edges: state.edges.filter((e) => e.id !== edgeId),
    selectedEdgeId: state.selectedEdgeId === edgeId ? null : state.selectedEdgeId,
    isDirty: true,
  })),

  selectNode: (nodeId) => set({ selectedNodeId: nodeId, selectedEdgeId: null }),
  selectEdge: (edgeId) => set({ selectedEdgeId: edgeId, selectedNodeId: null }),

  saveSnapshot: () => set((state) => ({
    history: [...state.history.slice(0, state.historyIndex + 1), {
      nodes: [...state.nodes],
      edges: [...state.edges],
    }],
    historyIndex: state.historyIndex + 1,
  })),

  undo: () => set((state) => {
    if (state.historyIndex <= 0) return {};
    const prev = state.history[state.historyIndex - 1];
    return {
      nodes: prev.nodes,
      edges: prev.edges,
      historyIndex: state.historyIndex - 1,
    };
  }),

  redo: () => set((state) => {
    if (state.historyIndex >= state.history.length - 1) return {};
    const next = state.history[state.historyIndex + 1];
    return {
      nodes: next.nodes,
      edges: next.edges,
      historyIndex: state.historyIndex + 1,
    };
  }),

  markDirty: () => set({ isDirty: true }),
  markClean: () => set({ isDirty: false }),
}));
```

---

## Responsive Design

### Mobile Workflow List (<=640px)

```
+----------------------+
| Workflows      [+]   |
+----------------------+
| [Search workflows..] |
+----------------------+
| +------------------+ |
| | Customer Support | |
| | Draft | v1.2     | |
| | 6 nodes | 23 runs| |
| | [Edit] [Run]     | |
| +------------------+ |
| +------------------+ |
| | Data Analysis    | |
| | Published | v1.0 | |
| | 4 nodes | 89 runs| |
| | [Edit] [Run]     | |
| +------------------+ |
+----------------------+
```

### Builder on Tablet

- Node palette collapses to icon-only sidebar
- Properties panel becomes a bottom sheet
- Canvas takes full width
- MiniMap always visible

### Breakpoints

| Breakpoint | Builder Layout                                    |
|------------|---------------------------------------------------|
| <=640px    | List view only (no builder on mobile)             |
| 641-768px  | Builder with collapsed palette + bottom properties|
| 769-1024px | Builder with narrow palette + right properties    |
| >1024px    | Full builder with all panels                      |
