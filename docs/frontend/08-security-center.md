# Security Center

## Table of Contents

- [Overview](#overview)
- [Page Layout](#page-layout)
- [Security Alerts](#security-alerts)
- [Security Alert Detail](#security-alert-detail)
- [Security Alert Cards](#security-alert-cards)
- [Audit Log Viewer](#audit-log-viewer)
- [Audit Log Detail](#audit-log-detail)
- [Audit Log Filters](#audit-log-filters)
- [Audit Log Export](#audit-log-export)
- [Access Control Management](#access-control-management)
- [Role Management](#role-management)
- [Permission Matrix](#permission-matrix)
- [User Access Review](#user-access-review)
- [Threat Detection Dashboard](#threat-detection-dashboard)
- [Prompt Injection Detection](#prompt-injection-detection)
- [Sensitive Words Filter](#sensitive-words-filter)
- [IP Whitelisting](#ip-whitelisting)
- [Session Management](#session-management)
- [Two-Factor Authentication](#two-factor-authentication)
- [API Integration](#api-integration)
- [Hooks](#hooks)
- [Stores](#stores)
- [Error Handling](#error-handling)
- [Responsive Design](#responsive-design)

---

## Overview

The Security Center provides centralized visibility and control over system security. It covers alerting, audit logging, access control, threat detection, and security configuration.

```
+------------------------------------------------------------------+
|  Security Center                                                  |
+--------------+---------------------------------------------------+
|              |                                                   |
|  Dashboard   |  Security Score: 92/100   ###########  92%       |
|  Alerts (3!) |                                                   |
|  Audit Logs  |  Active Alerts: 3 (1 High, 1 Med, 1 Low)        |
|  Access Ctrl |  Audit Events (24h): 12,450                      |
|  Threats     |  Failed Logins (24h): 7                           |
|  IP Filter   |  Prompt Injections Blocked: 23                    |
|  Sessions    |                                                   |
|  2FA         |  Recent Alerts                                    |
|              |    HIGH: Unusual API access pattern               |
|              |    MED:  3 failed login attempts                  |
|              |    LOW:  New IP address detected                  |
+--------------+---------------------------------------------------+
```

---

## Page Layout

### Component Tree

```
SecurityCenterPage
+-- SecuritySidebar
|   +-- NavigationLinks (Dashboard, Alerts, Audit, Access, Threats, Settings)
|   +-- SecurityScore
|   +-- AlertSummary
+-- MainContent
    +-- DashboardTab
    |   +-- AlertSummaryCards
    |   +-- ThreatTrendChart
    |   +-- RecentEventsList
    +-- AlertsTab
    |   +-- AlertFilters
    |   +-- AlertList
    |   +-- AlertDetail (expandable/modal)
    +-- AuditTab
    |   +-- AuditFilters
    |   +-- AuditTable
    |   +-- AuditDetail (expandable/modal)
    |   +-- ExportButton
    +-- AccessTab
    |   +-- RoleList
    |   +-- PermissionMatrix
    |   +-- UserAccessReview
    +-- ThreatsTab
    |   +-- ThreatDashboard
    |   +-- PromptInjectionLog
    |   +-- SensitiveWordsConfig
    +-- SettingsTab
        +-- IPWhitelist
        +-- SessionManagement
        +-- TwoFactorConfig
```

---

## Security Alerts

### Alert List

```
+----------------------------------------------------------------------+
|  Security Alerts                                    [Mark All Read]   |
+----------------------------------------------------------------------+
|  Filter: [All Severities v] [All Types v] [All Status v]            |
|                                                                      |
|  +----------------------------------------------------------------+  |
|  | HIGH   | Unusual API access pattern detected                  |  |
|  |        | Source: 203.0.113.42 (unknown IP)                    |  |
|  |        | Time: Jul 16, 2026 10:32 AM                          |  |
|  |        | Status: Unresolved                                    |  |
|  |        | [View Details] [Block IP] [Dismiss]                  |  |
|  +----------------------------------------------------------------+  |
|  | MEDIUM | 3 failed login attempts for user@acme.com            |  |
|  |        | Source: 198.51.100.23                                 |  |
|  |        | Time: Jul 16, 2026 09:15 AM                          |  |
|  |        | Status: Investigating                                 |  |
|  |        | [View Details] [Block IP] [Dismiss]                  |  |
|  +----------------------------------------------------------------+  |
|  | LOW    | New IP address detected for admin@example.com         |  |
|  |        | Source: 192.0.2.100 (Berlin, DE)                     |  |
|  |        | Time: Jul 16, 2026 08:00 AM                          |  |
|  |        | Status: Auto-resolved                                 |  |
|  |        | [View Details] [Dismiss]                              |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Showing 1-10 of 47 alerts    < 1 2 3 ... 5 >                       |
+----------------------------------------------------------------------+
```

### Alert Types

| Type                    | Description                                   | Severity Range |
|-------------------------|-----------------------------------------------|----------------|
| Failed Login            | Multiple failed authentication attempts       | LOW - HIGH     |
| Unusual API Access      | Anomalous API call patterns                   | MEDIUM - HIGH  |
| Permission Escalation   | Unauthorized permission change attempt        | HIGH           |
| Data Exfiltration       | Unusual data download patterns                | HIGH           |
| Prompt Injection        | Detected prompt injection attempt             | MEDIUM - HIGH  |
| New IP Login            | Login from previously unseen IP               | LOW            |
| Session Anomaly         | Concurrent sessions from different locations  | MEDIUM         |
| Admin Action            | Administrative changes to security settings   | INFO           |
| Rate Limiting           | Rate limit exceeded threshold                 | LOW - MEDIUM   |
| Model Abuse             | Excessive model usage or abuse pattern        | MEDIUM - HIGH  |

---

## Security Alert Detail

```
+----------------------------------------------------------------------+
|  <- Back to Alerts                                                   |
+----------------------------------------------------------------------+
|                                                                      |
|  Unusual API Access Pattern Detected                          HIGH   |
|  Alert ID: sec_alert_abc123                                          |
|  Status: [Unresolved v]                                              |
|                                                                      |
|  Timeline:                                                           |
|  +----------------------------------------------------------------+  |
|  |  10:30:00  First request from IP 203.0.113.42                  |  |
|  |  10:30:15  2nd request - different endpoint                    |  |
|  |  10:30:16  3rd request - bulk data endpoint                    |  |
|  |  10:30:17  4th request - different API key                     |  |
|  |  10:32:00  ALERT triggered - pattern detected                  |  |
|  |  10:32:01  IP flagged for review                               |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Event Details:                                                      |
|  +----------------------------------------------------------------+  |
|  |  Source IP:       203.0.113.42                                 |  |
|  |  User Agent:      Python-urllib/3.11                           |  |
|  |  API Key Used:    key_***7x2k                                  |  |
|  |  Endpoints Hit:   /api/documents, /api/export                  |  |
|  |  Request Count:   47 in 2 minutes                              |  |
|  |  Geolocation:     Unknown / Proxy                               |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Actions:                                                            |
|  [Block IP] [Revoke API Key] [Disable User] [Dismiss]               |
|  [Assign to Admin] [Add Note]                                        |
+----------------------------------------------------------------------+
```

### Alert Statuses

| Status         | Description                                  |
|----------------|----------------------------------------------|
| Unresolved     | New alert, needs investigation               |
| Investigating  | Being actively reviewed                      |
| Confirmed      | Threat confirmed, action needed              |
| Auto-resolved  | Automatically handled by system              |
| Dismissed      | Reviewed and dismissed as false positive     |
| Resolved       | Issue addressed and resolved                 |

---

## Security Alert Cards

### Severity Badge Styling

| Severity | Color  | Background  | Border         |
|----------|--------|-------------|----------------|
| HIGH     | Red    | bg-red-50   | border-red-200 |
| MEDIUM   | Yellow | bg-yellow-50| border-yellow-200 |
| LOW      | Blue   | bg-blue-50  | border-blue-200 |
| INFO     | Gray   | bg-gray-50  | border-gray-200 |

### Alert Card Component

```tsx
// components/security/AlertCard.tsx
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { AlertTriangle, Shield, Info } from 'lucide-react';

interface AlertCardProps {
  alert: SecurityAlert;
  onClick: () => void;
  onBlock?: () => void;
  onDismiss?: () => void;
}

const SEVERITY_CONFIG = {
  high: { icon: AlertTriangle, color: 'text-red-600', bg: 'bg-red-50', border: 'border-red-200' },
  medium: { icon: Shield, color: 'text-yellow-600', bg: 'bg-yellow-50', border: 'border-yellow-200' },
  low: { icon: Info, color: 'text-blue-600', bg: 'bg-blue-50', border: 'border-blue-200' },
} as const;

export function AlertCard({ alert, onClick, onBlock, onDismiss }: AlertCardProps) {
  const config = SEVERITY_CONFIG[alert.severity];
  const Icon = config.icon;

  return (
    <Card
      className={`cursor-pointer hover:shadow-md transition-shadow ${config.bg} ${config.border}`}
      onClick={onClick}
      role="button"
      aria-label={`${alert.severity} alert: ${alert.title}`}
    >
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Icon className={`w-5 h-5 ${config.color}`} />
            <Badge variant={alert.severity === 'high' ? 'destructive' : 'secondary'}>
              {alert.severity.toUpperCase()}
            </Badge>
            <span className="font-semibold">{alert.title}</span>
          </div>
          <Badge variant={alert.status === 'unresolved' ? 'outline' : 'default'}>
            {alert.status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">{alert.description}</p>
        <div className="flex items-center justify-between mt-3">
          <span className="text-xs text-muted-foreground">
            {alert.source_ip} - {formatTimeAgo(alert.created_at)}
          </span>
          <div className="flex gap-2" onClick={(e) => e.stopPropagation()}>
            {onBlock && <Button size="sm" variant="destructive" onClick={onBlock}>Block IP</Button>}
            {onDismiss && <Button size="sm" variant="outline" onClick={onDismiss}>Dismiss</Button>}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
```

---

## Audit Log Viewer

### Table Layout

```
+----------------------------------------------------------------------+
|  Audit Logs                                            [Export v]    |
+----------------------------------------------------------------------+
|  Search: [_________________________]  Date: [Jul 15] - [Jul 16]     |
|  Event Type: [All v]  User: [All v]  Action: [All v]               |
+----------------------------------------------------------------------+
|                                                                      |
|  Time              | User            | Event Type    | Action        |
|  ------------------+-----------------+---------------+---------------|
|  10:32:15 AM       | admin@co.com    | Auth          | Login         |
|  10:30:00 AM       | system          | Security      | Alert         |
|  10:28:45 AM       | john@acme.com   | Document      | Upload        |
|  10:25:12 AM       | jane@acme.com   | Model         | Start         |
|  10:20:00 AM       | admin@co.com    | User          | Create        |
|  10:15:33 AM       | system          | Security      | Scan          |
|  10:10:00 AM       | john@acme.com   | Chat          | Send          |
|  10:05:22 AM       | admin@co.com    | Settings      | Update        |
|  10:00:00 AM       | system          | Health        | Check         |
|  09:55:10 AM       | jane@acme.com   | Agent         | Execute       |
|                                                                      |
|  Showing 1-10 of 12,450    < 1 2 3 ... 1245 >                       |
+----------------------------------------------------------------------+
```

### Audit Event Types

| Category    | Event Types                                              |
|-------------|----------------------------------------------------------|
| Auth        | Login, Logout, Login Failed, Password Change, 2FA Verify |
| User        | Create, Update, Delete, Role Change, Status Change       |
| Document    | Upload, Delete, View, Edit, Search, Download             |
| Model       | Start, Stop, Configure, Download, Remove                 |
| Agent       | Create, Execute, Update, Delete                          |
| Chat        | Send, Receive, Feedback                                  |
| Security    | Alert, Scan, Block IP, Revoke Key, Session Revoke        |
| Settings    | Update, Export, Import                                   |
| Billing     | Plan Change, Payment, Invoice                            |
| System      | Health Check, Backup, Restore, Migration                 |

---

## Audit Log Detail

```
+----------------------------------------------------------------------+
|  Audit Event Detail                                                  |
+----------------------------------------------------------------------+
|                                                                      |
|  Event ID:    audit_evt_xyz789                                       |
|  Timestamp:   2026-07-16T10:28:45.123Z                              |
|  User:        john@acme.com (ID: usr_abc123)                        |
|  Event Type:  Document                                               |
|  Action:      Upload                                                 |
|  IP Address:  192.168.1.100                                         |
|  User Agent:  Mozilla/5.0 (Macintosh; Intel Mac OS X...)            |
|  Status:      Success                                                |
|                                                                      |
|  Payload:                                                            |
|  +----------------------------------------------------------------+  |
|  |  {                                                              |  |
|  |    "document_id": "doc_def456",                                |  |
|  |    "document_name": "report.pdf",                               |  |
|  |    "file_size": 2411724,                                        |  |
|  |    "file_type": "application/pdf",                              |  |
|  |    "set_id": "set_ghi789",                                     |  |
|  |    "tags": ["finance", "quarterly"]                             |  |
|  |  }                                                              |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Metadata:                                                           |
|  +----------------------------------------------------------------+  |
|  |  Request ID:    req_abc123xyz                                  |  |
|  |  Session ID:    sess_xyz789                                    |  |
|  |  Tenant ID:     tenant_acme                                    |  |
|  |  Response Time: 245ms                                           |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

---

## Audit Log Filters

### Filter Options

| Filter       | Type     | Options                                         |
|--------------|----------|-------------------------------------------------|
| Date Range   | date     | Start date - End date (default: last 24h)       |
| Event Type   | multi    | Auth, User, Document, Model, Agent, Chat, etc.  |
| User         | multi    | User emails/names                               |
| Action       | multi    | Specific actions (login, create, delete, etc.)  |
| Status       | multi    | Success, Failure                                |
| IP Address   | text     | Filter by source IP                             |
| Resource ID  | text     | Filter by affected resource                     |
| Tenant       | single   | Filter by organization                          |

### Filter Component

```tsx
// components/security/AuditFilters.tsx
export function AuditFilters({ filters, onChange }: AuditFilterProps) {
  return (
    <div className="flex flex-wrap gap-3 p-4 border rounded-lg">
      <div className="flex-1 min-w-[200px]">
        <Label>Date Range</Label>
        <div className="flex items-center gap-2">
          <DatePicker value={filters.date_from}
            onChange={(d) => onChange({ ...filters, date_from: d })} />
          <span>to</span>
          <DatePicker value={filters.date_to}
            onChange={(d) => onChange({ ...filters, date_to: d })} />
        </div>
      </div>
      <div>
        <Label>Event Type</Label>
        <MultiSelect
          value={filters.event_types || []}
          onChange={(types) => onChange({ ...filters, event_types: types })}
          options={EVENT_TYPES}
        />
      </div>
      <div>
        <Label>User</Label>
        <MultiSelect
          value={filters.users || []}
          onChange={(users) => onChange({ ...filters, users })}
          options={USER_OPTIONS}
        />
      </div>
      <div>
        <Label>Action</Label>
        <MultiSelect
          value={filters.actions || []}
          onChange={(actions) => onChange({ ...filters, actions })}
          options={ACTION_OPTIONS}
        />
      </div>
      <div>
        <Label>Status</Label>
        <MultiSelect
          value={filters.statuses || []}
          onChange={(statuses) => onChange({ ...filters, statuses })}
          options={['Success', 'Failure']}
        />
      </div>
      <div className="flex items-end">
        <Button variant="outline" onClick={() => onChange({})}>Clear All</Button>
      </div>
    </div>
  );
}
```

---

## Audit Log Export

### Export Options

| Format | Best For                  | Includes                      |
|--------|---------------------------|-------------------------------|
| CSV    | Spreadsheet analysis      | All visible columns           |
| JSON   | Programmatic processing   | Full event payloads           |
| PDF    | Compliance reports        | Formatted table + summary     |

### Export Dialog

```tsx
// components/security/ExportAuditLogs.tsx
interface ExportOptions {
  format: 'csv' | 'json' | 'pdf';
  dateRange: { from: string; to: string };
  filters: AuditFilters;
  columns: string[];
}

export function ExportAuditLogs({ onExport }: { onExport: (opts: ExportOptions) => Promise<void> }) {
  const [options, setOptions] = useState<ExportOptions>({
    format: 'csv',
    dateRange: { from: '', to: '' },
    filters: {},
    columns: ['time', 'user', 'event_type', 'action', 'status'],
  });
  const [exporting, setExporting] = useState(false);

  const handleExport = async () => {
    setExporting(true);
    try {
      await onExport(options);
    } finally {
      setExporting(false);
    }
  };

  return (
    <Dialog>
      <DialogHeader>
        <DialogTitle>Export Audit Logs</DialogTitle>
      </DialogHeader>
      <div className="space-y-4">
        <div>
          <Label>Format</Label>
          <RadioGroup value={options.format}
            onValueChange={(v) => setOptions({ ...options, format: v as any })}>
            <RadioGroupItem value="csv">CSV</RadioGroupItem>
            <RadioGroupItem value="json">JSON</RadioGroupItem>
            <RadioGroupItem value="pdf">PDF</RadioGroupItem>
          </RadioGroup>
        </div>
        <div>
          <Label>Date Range</Label>
          <DatePicker value={options.dateRange.from}
            onChange={(d) => setOptions({ ...options, dateRange: { ...options.dateRange, from: d } })} />
          <DatePicker value={options.dateRange.to}
            onChange={(d) => setOptions({ ...options, dateRange: { ...options.dateRange, to: d } })} />
        </div>
        <div>
          <Label>Columns to Include</Label>
          <CheckboxGroup value={options.columns}
            onChange={(cols) => setOptions({ ...options, columns: cols })}>
            <CheckboxItem value="time">Timestamp</CheckboxItem>
            <CheckboxItem value="user">User</CheckboxItem>
            <CheckboxItem value="event_type">Event Type</CheckboxItem>
            <CheckboxItem value="action">Action</CheckboxItem>
            <CheckboxItem value="status">Status</CheckboxItem>
            <CheckboxItem value="ip">IP Address</CheckboxItem>
            <CheckboxItem value="payload">Full Payload</CheckboxItem>
          </CheckboxGroup>
        </div>
        <DialogFooter>
          <Button onClick={handleExport} disabled={exporting}>
            {exporting ? 'Exporting...' : 'Export'}
          </Button>
        </DialogFooter>
      </div>
    </Dialog>
  );
}
```

---

## Access Control Management

### Roles Overview

| Role        | Description                              | Users | Permissions Count |
|-------------|------------------------------------------|-------|-------------------|
| Super Admin | Full system access                       | 2     | All               |
| Admin       | Organization-level management            | 5     | 45                |
| Editor      | Content and document management          | 12    | 28                |
| Viewer      | Read-only access                         | 35    | 12                |
| API Only    | API access only, no UI                   | 3     | 8                 |
| Custom      | Custom role with selected permissions    | 8     | Variable          |

---

## Role Management

### Create/Edit Role Dialog

```
+--------------------------------------------------------------+
|  Create Role                                              [X]|
+--------------------------------------------------------------+
|                                                              |
|  Name:        [Custom Analyst                    ]           |
|  Description: [Can view and analyze data, no edits ]        |
|  Color:       #6366F1 (click to change)                     |
|                                                              |
|  Permissions:                                                |
|  +------------------------------------------------------+   |
|  |  Documents                                           |   |
|  |  [x] View documents    [x] Search documents         |   |
|  |  [ ] Upload documents  [ ] Edit documents           |   |
|  |  [ ] Delete documents  [ ] Manage sets              |   |
|  +------------------------------------------------------+   |
|  |  Models                                             |   |
|  |  [x] View models       [ ] Start/Stop models       |   |
|  |  [ ] Configure models  [ ] Download models          |   |
|  +------------------------------------------------------+   |
|  |  Agents                                             |   |
|  |  [x] View agents       [ ] Create agents           |   |
|  |  [ ] Execute agents    [ ] Configure agents         |   |
|  +------------------------------------------------------+   |
|  |  Security                                           |   |
|  |  [x] View audit logs   [ ] Manage alerts           |   |
|  |  [ ] Manage roles      [ ] Manage access            |   |
|  +------------------------------------------------------+   |
|  |  Billing                                            |   |
|  |  [x] View billing      [ ] Manage plan              |   |
|  +------------------------------------------------------+   |
|                                                              |
|         [Cancel]              [Create Role]                  |
+--------------------------------------------------------------+
```

### Role Component

```tsx
// components/security/RoleEditor.tsx
import { useState } from 'react';

const PERMISSION_GROUPS = [
  {
    group: 'Documents',
    permissions: [
      { id: 'doc.view', label: 'View documents' },
      { id: 'doc.search', label: 'Search documents' },
      { id: 'doc.upload', label: 'Upload documents' },
      { id: 'doc.edit', label: 'Edit documents' },
      { id: 'doc.delete', label: 'Delete documents' },
      { id: 'doc.manage_sets', label: 'Manage sets' },
    ],
  },
  {
    group: 'Models',
    permissions: [
      { id: 'model.view', label: 'View models' },
      { id: 'model.start_stop', label: 'Start/Stop models' },
      { id: 'model.configure', label: 'Configure models' },
      { id: 'model.download', label: 'Download models' },
    ],
  },
  {
    group: 'Agents',
    permissions: [
      { id: 'agent.view', label: 'View agents' },
      { id: 'agent.create', label: 'Create agents' },
      { id: 'agent.execute', label: 'Execute agents' },
      { id: 'agent.configure', label: 'Configure agents' },
    ],
  },
  {
    group: 'Security',
    permissions: [
      { id: 'security.view_audit', label: 'View audit logs' },
      { id: 'security.manage_alerts', label: 'Manage alerts' },
      { id: 'security.manage_roles', label: 'Manage roles' },
      { id: 'security.manage_access', label: 'Manage access' },
    ],
  },
  {
    group: 'Billing',
    permissions: [
      { id: 'billing.view', label: 'View billing' },
      { id: 'billing.manage', label: 'Manage plan' },
    ],
  },
];

interface RoleEditorProps {
  role?: Role;
  onSave: (data: CreateRoleRequest) => Promise<void>;
  onCancel: () => void;
}

export function RoleEditor({ role, onSave, onCancel }: RoleEditorProps) {
  const [name, setName] = useState(role?.name || '');
  const [description, setDescription] = useState(role?.description || '');
  const [permissions, setPermissions] = useState<Set<string>>(
    new Set(role?.permissions || [])
  );

  const togglePermission = (permId: string) => {
    setPermissions((prev) => {
      const next = new Set(prev);
      if (next.has(permId)) next.delete(permId);
      else next.add(permId);
      return next;
    });
  };

  return (
    <div className="space-y-4">
      <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Role name" />
      <Textarea value={description} onChange={(e) => setDescription(e.target.value)}
        placeholder="Description" />
      {PERMISSION_GROUPS.map((group) => (
        <div key={group.group} className="border rounded-lg p-4">
          <h4 className="font-semibold mb-2">{group.group}</h4>
          <div className="grid grid-cols-2 gap-2">
            {group.permissions.map((perm) => (
              <label key={perm.id} className="flex items-center gap-2 text-sm">
                <Checkbox
                  checked={permissions.has(perm.id)}
                  onCheckedChange={() => togglePermission(perm.id)}
                />
                {perm.label}
              </label>
            ))}
          </div>
        </div>
      ))}
      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={onCancel}>Cancel</Button>
        <Button onClick={() => onSave({ name, description, permissions: Array.from(permissions) })}>
          {role ? 'Save Changes' : 'Create Role'}
        </Button>
      </div>
    </div>
  );
}
```

---

## Permission Matrix

### Resource x Action Grid

| Resource       | View | Create | Edit | Delete | Share | Export |
|----------------|------|--------|------|--------|-------|--------|
| Documents      | Y    | Y      | Y    | N      | Y     | Y      |
| Document Sets  | Y    | Y      | Y    | Y      | N     | N      |
| Models         | Y    | N      | Y    | N      | N     | N      |
| Agents         | Y    | Y      | Y    | Y      | N     | N      |
| Chats          | Y    | Y      | N    | Y      | N     | Y      |
| Users          | Y    | Y      | Y    | Y      | N     | N      |
| Roles          | Y    | Y      | Y    | Y      | N     | N      |
| Audit Logs     | Y    | N      | N    | N      | N     | Y      |
| Security       | Y    | N      | Y    | N      | N     | N      |
| Billing        | Y    | N      | Y    | N      | N     | Y      |
| Settings       | Y    | N      | Y    | N      | N     | N      |
| Workflows      | Y    | Y      | Y    | Y      | N     | Y      |

### Permission Matrix Component

```tsx
// components/security/PermissionMatrix.tsx
export function PermissionMatrix({ roles, resources }: PermissionMatrixProps) {
  const actions = ['view', 'create', 'edit', 'delete', 'share', 'export'];

  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse">
        <thead>
          <tr>
            <th className="border p-2 text-left">Resource</th>
            {roles.map((role) => (
              <th key={role.id} className="border p-2 text-center" colSpan={actions.length}>
                {role.name}
              </th>
            ))}
          </tr>
          <tr>
            <th className="border p-2">Action</th>
            {roles.map((role) =>
              actions.map((action) => (
                <th key={`${role.id}-${action}`} className="border p-2 text-center text-xs">
                  {action.charAt(0).toUpperCase()}
                </th>
              ))
            )}
          </tr>
        </thead>
        <tbody>
          {resources.map((resource) => (
            <tr key={resource}>
              <td className="border p-2 font-medium">{resource}</td>
              {roles.map((role) =>
                actions.map((action) => (
                  <td key={`${role.id}-${resource}-${action}`}
                    className="border p-2 text-center">
                    {role.permissions.includes(`${resource}.${action}`) ? 'Y' : 'N'}
                  </td>
                ))
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

---

## User Access Review

```
+----------------------------------------------------------------------+
|  User Access Review                                                  |
+----------------------------------------------------------------------+
|                                                                      |
|  User: [john@acme.com v]                                             |
|                                                                      |
|  Assigned Roles:                                                     |
|  +----------------------------------------------------------------+  |
|  |  Role         | Assigned By      | Since           | Status   |  |
|  |---------------|------------------+-----------------+----------|  |
|  |  Editor       | admin@co.com     | Jan 15, 2026    | Active   |  |
|  |  Doc Manager  | admin@co.com     | Mar 20, 2026    | Active   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Effective Permissions:                                              |
|  +----------------------------------------------------------------+  |
|  |  Documents:  View, Search, Upload, Edit, Manage Sets           |  |
|  |  Models:     View                                              |  |
|  |  Agents:     View, Create, Execute                             |  |
|  |  Chats:      View, Create, Delete                              |  |
|  |  Billing:    (none)                                            |  |
|  |  Security:   View audit logs                                   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Recent Access (Last 7 days):                                        |
|  +----------------------------------------------------------------+  |
|  |  Jul 16  10:28  Document Upload   192.168.1.100  Success       |  |
|  |  Jul 16  09:15  Agent Execute     192.168.1.100  Success       |  |
|  |  Jul 15  16:30  Chat Send         192.168.1.100  Success       |  |
|  |  Jul 15  11:00  Login             192.168.1.100  Success       |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

---

## Threat Detection Dashboard

```
+----------------------------------------------------------------------+
|  Threat Detection Dashboard                                          |
+----------------------------------------------------------------------+
|                                                                      |
|  Threat Level: LOW                                                   |
|  Score: 12/100                                                       |
|  +----------------------------------------------------------------+  |
|  |  Threat Trend (Last 30 Days)                                   |  |
|  |                                                                |  |
|  |  30 |                                                          |  |
|  |  20 |     *                                                    |  |
|  |  10 |  *     *   *                                            |  |
|  |   0 +--*-----*---*-----*----*----*----*----*----*----*--------|  |
|  |     Day1  Day5  Day10  Day15  Day20  Day25  Day30             |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Threat Types:                                                       |
|  +----------------------------------------------------------------+  |
|  |  Type                | 24h  | 7d   | 30d  | Trend               |  |
|  |----------------------+------+------+------+---------------------|  |
|  |  Failed Login        |  7   |  42  | 180  | Decreasing  -15%   |  |
|  |  Prompt Injection    |  23  | 156  | 890  | Increasing  +12%   |  |
|  |  API Abuse           |  3   |  18  |  67  | Stable       0%   |  |
|  |  Permission Denied   |  12  |  89  | 340  | Decreasing   -8%  |  |
|  |  Data Exfiltration   |  0   |   2  |   5  | Decreasing -100%   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Geographic Hotspots:                                                |
|  +----------------------------------------------------------------+  |
|  |  Region         | Threats | Top Type           | Action Taken  |  |
|  |-----------------+---------+--------------------+---------------|  |
|  |  Unknown/Proxy  |   45    | API Abuse          | 12 IPs blocked|  |
|  |  Eastern Europe |   23    | Failed Login       |  5 IPs blocked|  |
|  |  Southeast Asia |   18    | Prompt Injection   |  3 IPs blocked|  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

---

## Prompt Injection Detection

### Detection Log

```
+----------------------------------------------------------------------+
|  Prompt Injection Detection Log                                      |
+----------------------------------------------------------------------+
|                                                                      |
|  Summary: 23 attempts blocked in last 24 hours                       |
|  Patterns: 12 unique patterns detected                               |
|                                                                      |
|  Recent Blocked Attempts:                                            |
|  +----------------------------------------------------------------+  |
|  | Time       | User         | Pattern          | Action           |  |
|  |------------|--------------|------------------+------------------|  |
|  | 10:32:15   | anonymous    | Instruction      | Blocked + Logged |  |
|  | 10:28:00   | user_7x2k    | Role Play        | Blocked + Logged |  |
|  | 10:15:33   | user_9m4p    | Instruction      | Blocked + Logged |  |
|  | 10:00:00   | anonymous    | System Override  | Blocked + Logged |  |
|  | 09:45:12   | user_2j8n    | Instruction      | Blocked + Logged |  |
|  | 09:30:00   | anonymous    | Data Extraction  | Blocked + Logged |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Detection Patterns:                                                 |
|  +----------------------------------------------------------------+  |
|  | Pattern            | Category        | Enabled | Block Count   |  |
|  |--------------------+-----------------+---------+---------------|  |
|  | Instruction Inject | Direct          | Yes     | 890           |  |
|  | Role Play          | Persona         | Yes     | 456           |  |
|  | System Override    | Privilege       | Yes     | 234           |  |
|  | Data Extraction    | Information     | Yes     | 178           |  |
|  | Context Manipulation| Context        | Yes     | 123           |  |
|  | Encoding Bypass    | Evasion         | Yes     | 89            |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

### Injection Detection Component

```tsx
// components/security/PromptInjectionLog.tsx
import { useQuery } from '@tanstack/react-query';
import { Badge } from '@/components/ui/badge';
import { api } from '@/lib/api';

interface InjectionEvent {
  id: string;
  timestamp: string;
  user_id: string;
  user_display: string;
  pattern: string;
  category: string;
  input_preview: string;
  action: 'blocked' | 'flagged' | 'passed';
  confidence: number;
}

export function PromptInjectionLog() {
  const { data: events, isLoading } = useQuery({
    queryKey: ['security', 'injection-log'],
    queryFn: () => api.post('/security/injection-log', {}),
    refetchInterval: 10_000,
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Prompt Injection Log</h3>
        <Badge variant="destructive">
          {events?.summary?.blocked_24h || 0} blocked (24h)
        </Badge>
      </div>
      <div className="border rounded-lg overflow-hidden">
        <table className="w-full">
          <thead className="bg-muted">
            <tr>
              <th className="p-3 text-left text-sm">Time</th>
              <th className="p-3 text-left text-sm">User</th>
              <th className="p-3 text-left text-sm">Pattern</th>
              <th className="p-3 text-left text-sm">Confidence</th>
              <th className="p-3 text-left text-sm">Action</th>
            </tr>
          </thead>
          <tbody>
            {(events?.data || []).map((event: InjectionEvent) => (
              <tr key={event.id} className="border-t hover:bg-muted/50">
                <td className="p-3 text-sm">{formatTime(event.timestamp)}</td>
                <td className="p-3 text-sm">{event.user_display}</td>
                <td className="p-3 text-sm">
                  <Badge variant="outline">{event.pattern}</Badge>
                </td>
                <td className="p-3 text-sm">{Math.round(event.confidence * 100)}%</td>
                <td className="p-3 text-sm">
                  <Badge variant={event.action === 'blocked' ? 'destructive' : 'secondary'}>
                    {event.action}
                  </Badge>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
```

---

## Sensitive Words Filter

### Configuration

```
+----------------------------------------------------------------------+
|  Sensitive Words Filter Configuration                                |
+----------------------------------------------------------------------+
|                                                                      |
|  +----------------------------------------------------------------+  |
|  |  Word List: Profanity                                          |  |
|  |  Status: Active   |   Words: 1,245   |   Added today: 3       |  |
|  |  [Edit List] [Disable] [Export]                                 |  |
|  +----------------------------------------------------------------+  |
|  |  Word List: PII Patterns                                        |  |
|  |  Status: Active   |   Patterns: 42    |   Regex-based          |  |
|  |  [Edit List] [Disable] [Export]                                 |  |
|  +----------------------------------------------------------------+  |
|  |  Word List: Competitor Names                                    |  |
|  |  Status: Active   |   Words: 28      |   Case-insensitive      |  |
|  |  [Edit List] [Disable] [Export]                                 |  |
|  +----------------------------------------------------------------+  |
|  |  Word List: Internal Code Names                                 |  |
|  |  Status: Inactive |   Words: 15      |   Exact match           |  |
|  |  [Edit List] [Enable] [Export]                                  |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Action on Match:                                                    |
|  ( ) Log only                                                       |
|  (X) Block and log                                                  |
|  ( ) Flag for review                                                |
|                                                                      |
|  Scope:                                                              |
|  [X] User inputs  [X] AI outputs  [ ] System messages              |
|                                                                      |
|  [Save Configuration]  [Test Filter]                                 |
+----------------------------------------------------------------------+
```

### Filter Component

```tsx
// components/security/SensitiveWordsConfig.tsx
interface WordList {
  id: string;
  name: string;
  status: 'active' | 'inactive';
  word_count: number;
  type: 'exact' | 'regex' | 'fuzzy';
  last_updated: string;
}

interface FilterConfig {
  action: 'log' | 'block' | 'flag';
  scope: {
    user_inputs: boolean;
    ai_outputs: boolean;
    system_messages: boolean;
  };
}

export function SensitiveWordsConfig() {
  const [config, setConfig] = useState<FilterConfig>({
    action: 'block',
    scope: { user_inputs: true, ai_outputs: true, system_messages: false },
  });

  return (
    <div className="space-y-6">
      <h3 className="text-lg font-semibold">Sensitive Words Filter</h3>
      <WordListManager />
      <div className="space-y-4 border rounded-lg p-4">
        <h4 className="font-medium">Filter Actions</h4>
        <RadioGroup
          value={config.action}
          onValueChange={(v) => setConfig({ ...config, action: v as any })}
        >
          <div className="flex items-center gap-2">
            <RadioGroupItem value="log" id="action-log" />
            <Label htmlFor="action-log">Log only</Label>
          </div>
          <div className="flex items-center gap-2">
            <RadioGroupItem value="block" id="action-block" />
            <Label htmlFor="action-block">Block and log</Label>
          </div>
          <div className="flex items-center gap-2">
            <RadioGroupItem value="flag" id="action-flag" />
            <Label htmlFor="action-flag">Flag for review</Label>
          </div>
        </RadioGroup>
      </div>
      <div className="space-y-4 border rounded-lg p-4">
        <h4 className="font-medium">Filter Scope</h4>
        <div className="flex gap-4">
          <label className="flex items-center gap-2">
            <Checkbox checked={config.scope.user_inputs}
              onCheckedChange={(c) =>
                setConfig({ ...config, scope: { ...config.scope, user_inputs: !!c } })
              } />
            User inputs
          </label>
          <label className="flex items-center gap-2">
            <Checkbox checked={config.scope.ai_outputs}
              onCheckedChange={(c) =>
                setConfig({ ...config, scope: { ...config.scope, ai_outputs: !!c } })
              } />
            AI outputs
          </label>
          <label className="flex items-center gap-2">
            <Checkbox checked={config.scope.system_messages}
              onCheckedChange={(c) =>
                setConfig({ ...config, scope: { ...config.scope, system_messages: !!c } })
              } />
            System messages
          </label>
        </div>
      </div>
    </div>
  );
}
```

---

## IP Whitelisting

### Whitelist Management

```
+----------------------------------------------------------------------+
|  IP Whitelist / Blacklist                                            |
+----------------------------------------------------------------------+
|                                                                      |
|  Tab: [Whitelisted IPs] [Blacklisted IPs] [Blocked Attempts]        |
|                                                                      |
|  +----------------------------------------------------------------+  |
|  |  IP Address       | Label            | Added By    | Actions   |  |
|  |-------------------|------------------+-------------+-----------|  |
|  |  192.168.1.0/24   | Office Network   | admin       | Edit Del  |  |
|  |  10.0.0.0/8       | VPN Range        | admin       | Edit Del  |  |
|  |  203.0.113.42     | CI Server        | admin       | Edit Del  |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [+ Add IP Range]  [+ Add Single IP]                                |
|                                                                      |
|  Blacklist:                                                          |
|  +----------------------------------------------------------------+  |
|  |  IP Address       | Reason                | Added   | Expires  |  |
|  |-------------------+-----------------------+---------+----------|  |
|  |  198.51.100.0/24  | Brute force attempts  | Jul 15  | Jul 22   |  |
|  |  100.64.0.5       | Data exfiltration     | Jul 14  | Permanent|  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

### IP Filter Component

```tsx
// components/security/IPWhitelist.tsx
interface IPEntry {
  id: string;
  ip: string;
  type: 'whitelist' | 'blacklist';
  label: string;
  added_by: string;
  added_at: string;
  expires_at?: string;
}

export function IPWhitelist() {
  const [tab, setTab] = useState<'whitelist' | 'blacklist'>('whitelist');

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">IP Access Control</h3>
        <Button onClick={() => setTab(tab)}>
          {tab === 'whitelist' ? '+ Add to Whitelist' : '+ Add to Blacklist'}
        </Button>
      </div>
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          <TabsTrigger value="whitelist">Whitelisted</TabsTrigger>
          <TabsTrigger value="blacklist">Blacklisted</TabsTrigger>
        </TabsList>
        <TabsContent value="whitelist">
          <IPTable type="whitelist" />
        </TabsContent>
        <TabsContent value="blacklist">
          <IPTable type="blacklist" />
        </TabsContent>
      </Tabs>
    </div>
  );
}
```

---

## Session Management

### Active Sessions

```
+----------------------------------------------------------------------+
|  Session Management                                                  |
+----------------------------------------------------------------------+
|                                                                      |
|  Active Sessions: 5                                                  |
|                                                                      |
|  +----------------------------------------------------------------+  |
|  |  Device         | IP Address     | Last Active  | Status       |  |
|  |-----------------+----------------+--------------+--------------|  |
|  |  Chrome/Mac     | 192.168.1.100  | Now          | Current (Y)  |  |
|  |  Safari/iOS     | 192.168.1.105  | 5 min ago    | Active  [X]  |  |
|  |  Firefox/Linux  | 10.0.0.50      | 1 hour ago   | Active  [X]  |  |
|  |  Chrome/Win     | 10.0.0.55      | 3 hours ago  | Idle     [X] |  |
|  |  API Client     | 203.0.113.42   | 30 min ago   | Active  [X]  |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Revoke All Other Sessions]  [Revoke Selected]                      |
|                                                                      |
|  Session Policy:                                                     |
|  +----------------------------------------------------------------+  |
|  |  Max Sessions Per User:    [5]                                  |  |
|  |  Session Timeout:          [30] minutes                         |  |
|  |  Idle Timeout:             [15] minutes                         |  |
|  |  Require Re-auth for Sensitive Actions: [Yes]                   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Save Policy]                                                       |
+----------------------------------------------------------------------+
```

### Session Component

```tsx
// components/security/SessionManager.tsx
interface Session {
  id: string;
  device: string;
  browser: string;
  os: string;
  ip_address: string;
  last_active: string;
  is_current: boolean;
  created_at: string;
  user_agent: string;
}

export function SessionManager() {
  const { data: sessions, isLoading } = useQuery({
    queryKey: ['security', 'sessions'],
    queryFn: () => api.post('/security/sessions', {}),
  });

  const revokeSession = useMutation({
    mutationFn: (sessionId: string) => api.delete(`/security/sessions/${sessionId}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['security', 'sessions'] }),
  });

  const revokeAll = useMutation({
    mutationFn: () => api.delete('/security/sessions', { data: { except_current: true } }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['security', 'sessions'] }),
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Active Sessions ({sessions?.length || 0})</h3>
        <Button variant="destructive" onClick={() => revokeAll.mutate()}>
          Revoke All Others
        </Button>
      </div>
      <div className="space-y-2">
        {sessions?.map((session: Session) => (
          <div key={session.id} className="flex items-center justify-between p-3 border rounded-lg">
            <div>
              <p className="font-medium">
                {session.browser}/{session.os}
                {session.is_current && <Badge className="ml-2">Current</Badge>}
              </p>
              <p className="text-sm text-muted-foreground">
                {session.ip_address} - Last active: {formatTimeAgo(session.last_active)}
              </p>
            </div>
            {!session.is_current && (
              <Button
                size="sm"
                variant="outline"
                onClick={() => revokeSession.mutate(session.id)}
              >
                Revoke
              </Button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## Two-Factor Authentication

### 2FA Configuration

```
+----------------------------------------------------------------------+
|  Two-Factor Authentication                                           |
+----------------------------------------------------------------------+
|                                                                      |
|  Status: Enabled                                                     |
|  Method: Authenticator App                                           |
|                                                                      |
|  Your recovery codes (save these somewhere safe):                    |
|  +----------------------------------------------------------------+  |
|  |  ABCD-EFGH-IJKL       MNOP-QRST-UVWX                        |  |
|  |  1234-5678-9ABC       DEF0-1234-5678                        |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Download Recovery Codes]  [Regenerate Codes]                       |
|                                                                      |
|  Device Registration:                                                |
|  +----------------------------------------------------------------+  |
|  |  Device              | Last Used        | Actions               |  |
|  |----------------------+------------------+-----------------------|  |
|  |  Google Authenticator| Jul 16, 10:28 AM | [Revoke]              |  |
|  |  Backup Codes        | Never            | [View] [Regenerate]   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Disable 2FA]                                                       |
+----------------------------------------------------------------------+
```

### 2FA Setup Flow

```
Step 1: Scan QR Code
+--------------------------+
|  [QR CODE IMAGE]         |
|                          |
|  Or enter manually:      |
|  JBSWY3DPEHPK3PXP        |
+--------------------------+

Step 2: Verify Code
+--------------------------+
|  Enter the 6-digit code  |
|  from your app:          |
|  [______]                |
|                          |
|  [Verify]  [Cancel]      |
+--------------------------+

Step 3: Save Recovery Codes
+--------------------------+
|  Save these codes:       |
|  ABCD-EFGH-IJKL          |
|  MNOP-QRST-UVWX          |
|  ...                     |
|                          |
|  [Download] [I've Saved] |
+--------------------------+
```

---

## API Integration

### Endpoints

| Method   | Endpoint                           | Description                    |
|----------|------------------------------------|--------------------------------|
| `POST`   | `/api/security/alerts`             | List security alerts           |
| `POST`   | `/api/security/alerts/:id`         | Get alert details              |
| `PATCH`  | `/api/security/alerts/:id`         | Update alert status            |
| `DELETE` | `/api/security/alerts/:id`         | Dismiss/delete alert           |
| `POST`   | `/api/security/alerts/:id/block`   | Block IP from alert            |
| `POST`   | `/api/security/audit`              | List audit logs                |
| `POST`   | `/api/security/audit/:id`          | Get audit event detail         |
| `POST`   | `/api/security/audit/export`       | Export audit logs              |
| `POST`   | `/api/security/roles`              | List roles                     |
| `POST`   | `/api/security/roles`              | Create role                    |
| `PATCH`  | `/api/security/roles/:id`          | Update role                    |
| `DELETE` | `/api/security/roles/:id`          | Delete role                    |
| `POST`   | `/api/security/permissions`        | Get permission matrix          |
| `POST`   | `/api/security/user-access/:uid`   | Get user access review         |
| `POST`   | `/api/security/threats`            | Get threat dashboard data      |
| `POST`   | `/api/security/injection-log`      | Get injection detection log    |
| `POST`   | `/api/security/sensitive-words`    | Get word filter config         |
| `PATCH`  | `/api/security/sensitive-words`    | Update word filter config      |
| `POST`   | `/api/security/ip-list`            | Get IP whitelist/blacklist     |
| `POST`   | `/api/security/ip-list`            | Add IP to list                 |
| `DELETE` | `/api/security/ip-list/:id`        | Remove IP from list            |
| `POST`   | `/api/security/sessions`           | List active sessions           |
| `DELETE` | `/api/security/sessions/:id`       | Revoke a session               |
| `DELETE` | `/api/security/sessions`           | Revoke multiple sessions       |
| `POST`   | `/api/security/2fa/enable`         | Enable 2FA                     |
| `POST`   | `/api/security/2fa/verify`         | Verify 2FA code                |
| `DELETE` | `/api/security/2fa/disable`        | Disable 2FA                    |
| `POST`   | `/api/security/2fa/recovery-codes` | Get recovery codes             |

---

## Hooks

### useSecurityAlerts

```typescript
// hooks/security/useSecurityAlerts.ts
export function useSecurityAlerts(filters?: AlertFilters) {
  const queryClient = useQueryClient();

  const alerts = useQuery({
    queryKey: ['security-alerts', filters],
    queryFn: () => api.post('/security/alerts', filters),
    refetchInterval: 15_000,
  });

  const updateAlert = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { status: string } }) =>
      api.patch(`/security/alerts/${id}`, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['security-alerts'] }),
  });

  const blockIP = useMutation({
    mutationFn: (alertId: string) =>
      api.post(`/security/alerts/${alertId}/block`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['security-alerts'] }),
  });

  return { alerts, updateAlert, blockIP };
}
```

### useAuditLogs

```typescript
// hooks/security/useAuditLogs.ts
export function useAuditLogs(filters?: AuditFilters) {
  const logs = useQuery({
    queryKey: ['audit-logs', filters],
    queryFn: () => api.post('/security/audit', filters),
    staleTime: 5_000,
  });

  const exportLogs = useMutation({
    mutationFn: (options: ExportOptions) =>
      api.post('/security/audit/export', { ...options, responseType: 'blob' }),
  });

  return { logs, exportLogs };
}
```

### useAccessControl

```typescript
// hooks/security/useAccessControl.ts
export function useAccessControl() {
  const queryClient = useQueryClient();

  const roles = useQuery({
    queryKey: ['roles'],
    queryFn: () => api.post('/security/roles', {}),
  });

  const createRole = useMutation({
    mutationFn: (data: CreateRoleRequest) => api.post('/security/roles', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['roles'] }),
  });

  const updateRole = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Role> }) =>
      api.patch(`/security/roles/${id}`, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['roles'] }),
  });

  const deleteRole = useMutation({
    mutationFn: (id: string) => api.delete(`/security/roles/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['roles'] }),
  });

  return { roles, createRole, updateRole, deleteRole };
}
```

---

## Stores

### Security Store (Zustand)

```typescript
// stores/security/security.ts
import { create } from 'zustand';

interface SecurityState {
  alerts: SecurityAlert[];
  unreadCount: number;
  auditLogs: AuditEvent[];
  roles: Role[];
  activeSessions: Session[];
  threatScore: number;

  setAlerts: (alerts: SecurityAlert[]) => void;
  markAlertRead: (id: string) => void;
  markAllAlertsRead: () => void;
  setAuditLogs: (logs: AuditEvent[]) => void;
  setRoles: (roles: Role[]) => void;
  setActiveSessions: (sessions: Session[]) => void;
  setThreatScore: (score: number) => void;
  removeSession: (id: string) => void;
}

export const useSecurityStore = create<SecurityState>((set) => ({
  alerts: [],
  unreadCount: 0,
  auditLogs: [],
  roles: [],
  activeSessions: [],
  threatScore: 0,

  setAlerts: (alerts) => set({
    alerts,
    unreadCount: alerts.filter((a) => !a.read).length,
  }),

  markAlertRead: (id) => set((state) => ({
    alerts: state.alerts.map((a) =>
      a.id === id ? { ...a, read: true } : a
    ),
    unreadCount: state.unreadCount - 1,
  })),

  markAllAlertsRead: () => set((state) => ({
    alerts: state.alerts.map((a) => ({ ...a, read: true })),
    unreadCount: 0,
  })),

  setAuditLogs: (auditLogs) => set({ auditLogs }),
  setRoles: (roles) => set({ roles }),

  setActiveSessions: (activeSessions) => set({ activeSessions }),

  setThreatScore: (threatScore) => set({ threatScore }),

  removeSession: (id) => set((state) => ({
    activeSessions: state.activeSessions.filter((s) => s.id !== id),
  })),
}));
```

---

## Error Handling

### Error Types

| Error                    | Message                          | Action                        |
|--------------------------|----------------------------------|-------------------------------|
| Alert fetch failed       | "Failed to load alerts"          | Auto-retry, show toast        |
| Audit log fetch failed   | "Failed to load audit logs"      | Show cached data, retry       |
| Export failed            | "Export failed, try again"       | Retry button                  |
| Role create failed       | "Could not create role"          | Show validation errors        |
| Session revoke failed    | "Could not revoke session"       | Retry button                  |
| IP block failed          | "Failed to block IP"             | Retry, show error details     |
| 2FA setup failed         | "2FA setup interrupted"          | Restart flow                  |
| Permission denied        | "Insufficient permissions"       | Show restricted view          |
| Scan timeout             | "Security scan timed out"        | Auto-retry with backoff       |

### Error Boundary

```tsx
// components/security/SecurityErrorBoundary.tsx
import { ErrorBoundary } from '@/components/ErrorBoundary';

export function SecurityErrorBoundary({ children }: { children: React.ReactNode }) {
  return (
    <ErrorBoundary
      fallback={({ error, reset }) => (
        <div className="p-6 text-center space-y-4">
          <ShieldAlert className="w-12 h-12 text-destructive mx-auto" />
          <h3 className="text-lg font-semibold">Security Module Error</h3>
          <p className="text-muted-foreground">{error.message}</p>
          <p className="text-sm text-muted-foreground">
            If this persists, contact your system administrator.
          </p>
          <Button onClick={reset}>Try Again</Button>
        </div>
      )}
    >
      {children}
    </ErrorBoundary>
  );
}
```

---

## Responsive Design

### Mobile Layout (<=640px)

```
+----------------------+
| Security Center      |
| [Dashboard] [Alerts] |
| [Audit] [Access]     |
+----------------------+
|                      |
| Security Score: 92   |
| ################# 92%|
|                      |
| Active Alerts: 3     |
| Failed Logins: 7     |
|                      |
| Recent Alerts:       |
| +------------------+ |
| | HIGH: Unusual API | |
| | MED:  3 failed    | |
| | LOW:  New IP      | |
| +------------------+ |
+----------------------+
```

### Responsive Breakpoints

| Breakpoint  | Layout                                            |
|-------------|---------------------------------------------------|
| <=640px     | Stacked cards, tabbed navigation, single column   |
| 641-1024px  | 2-column grid, collapsible sidebar                |
| 1025-1280px | Full sidebar + content, split-panel details       |
| >1280px     | 3-column layout, side-by-side detail views        |

### Accessibility

- All alerts announced via `aria-live="polite"` region
- Severity badges use text + icon (not color alone)
- Keyboard navigation: Tab through all interactive elements
- Screen reader labels on all buttons and links
- Focus traps in modal dialogs
- Skip navigation link for audit log tables
- Color contrast meets WCAG 2.1 AA (4.5:1 for text, 3:1 for large text)
