# Admin Panel

## Table of Contents

- [Overview](#overview)
- [Admin Panel Layout](#admin-panel-layout)
- [User Management](#user-management)
- [User List Table](#user-list-table)
- [User Creation Form](#user-creation-form)
- [User Edit Form](#user-edit-form)
- [User Detail Page](#user-detail-page)
- [Organization (Tenant) Management](#organization-tenant-management)
- [Tenant List Table](#tenant-list-table)
- [Tenant Creation Form](#tenant-creation-form)
- [Tenant Detail Page](#tenant-detail-page)
- [KYC Management](#kyc-management)
- [KYC Document Viewer](#kyc-document-viewer)
- [KYC Status Tracking](#kyc-status-tracking)
- [Plan Management](#plan-management)
- [Billing Overview](#billing-overview)
- [Usage Analytics](#usage-analytics)
- [System Settings](#system-settings)
- [Feature Flags](#feature-flags)
- [API Integration](#api-integration)
- [Hooks](#hooks)
- [Stores](#stores)
- [Error Handling](#error-handling)
- [Responsive Design](#responsive-design)
- [Admin Role Protection](#admin-role-protection)

---

## Overview

The Admin Panel provides system administrators with tools for managing users, organizations (tenants), billing, KYC verification, and system configuration. Only users with admin or super-admin roles can access this section.

```
+------------------------------------------------------------------+
|  Admin Panel                                                      |
+--------------+---------------------------------------------------+
|              |                                                   |
|  Dashboard   |  Overview Stats                                  |
|  Users       |  +------------------+------------------+         |
|  Tenants     |  | Users: 156       | Active: 142      |         |
|  KYC         |  | Tenants: 23      | Pending KYC: 3   |         |
|  Billing     |  | Revenue: $12,450 | MRR: $8,900      |         |
|  Plans       |  +------------------+------------------+         |
|  Settings    |                                                   |
|  Flags       |  Recent Activity                                 |
|              |  - New user: john@acme.com (Editor)              |
|              |  - Tenant upgrade: TechStart -> Pro               |
|              |  - KYC submitted: DataCorp                       |
+--------------+---------------------------------------------------+
```

---

## Admin Panel Layout

### Component Tree

```
AdminPanel
+-- AdminSidebar
|   +-- NavItem: Dashboard
|   +-- NavItem: Users (badge: count)
|   +-- NavItem: Organizations
|   +-- NavItem: KYC (badge: pending count)
|   +-- NavItem: Billing
|   +-- NavItem: Plans
|   +-- NavItem: Settings
|   +-- NavItem: Feature Flags
+-- AdminContent
|   +-- AdminHeader (breadcrumb, admin avatar, notifications)
|   +-- AdminRouter (nested routes)
|   |   +-- DashboardPage
|   |   +-- UsersPage
|   |   +-- UserDetailPage
|   |   +-- TenantsPage
|   |   +-- TenantDetailPage
|   |   +-- KYCPage
|   |   +-- BillingPage
|   |   +-- PlansPage
|   |   +-- SettingsPage
|   |   +-- FeatureFlagsPage
|   +-- AdminFooter
+-- AdminGuards (role check wrapper)
```

### Layout Code

```tsx
// pages/admin/AdminPanel.tsx
import { AdminSidebar } from '@/components/admin/AdminSidebar';
import { AdminHeader } from '@/components/admin/AdminHeader';
import { AdminGuard } from '@/components/admin/AdminGuard';
import { Outlet } from 'react-router-dom';

export function AdminPanel() {
  return (
    <AdminGuard requiredRole="admin">
      <div className="flex h-screen bg-background">
        <AdminSidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <AdminHeader />
          <main className="flex-1 overflow-y-auto p-6">
            <Outlet />
          </main>
        </div>
      </div>
    </AdminGuard>
  );
}
```

### Sidebar Navigation

```tsx
// components/admin/AdminSidebar.tsx
import { NavLink } from 'react-router-dom';
import { Users, Building2, Shield, CreditCard, Settings, Flag, LayoutDashboard } from 'lucide-react';

const NAV_ITEMS = [
  { to: '/admin', icon: LayoutDashboard, label: 'Dashboard', end: true },
  { to: '/admin/users', icon: Users, label: 'Users' },
  { to: '/admin/tenants', icon: Building2, label: 'Organizations' },
  { to: '/admin/kyc', icon: Shield, label: 'KYC' },
  { to: '/admin/billing', icon: CreditCard, label: 'Billing' },
  { to: '/admin/plans', icon: CreditCard, label: 'Plans' },
  { to: '/admin/settings', icon: Settings, label: 'Settings' },
  { to: '/admin/flags', icon: Flag, label: 'Feature Flags' },
];

export function AdminSidebar() {
  return (
    <aside className="w-64 border-r bg-card flex flex-col">
      <div className="p-4 border-b">
        <h2 className="text-lg font-bold">Admin Panel</h2>
      </div>
      <nav className="flex-1 p-2 space-y-1">
        {NAV_ITEMS.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.end}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors
              ${isActive ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'}`
            }
          >
            <item.icon className="w-4 h-4" />
            {item.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
```

---

## User Management

### User List Table

```
+----------------------------------------------------------------------+
|  Users                                              [+ Create User]  |
+----------------------------------------------------------------------+
|  Search: [_________________________]  Role: [All v]  Status: [All v]|
|                                                                      |
|  Name            | Email              | Role    | Status | Last Login|
|  ----------------+--------------------+---------+--------+----------|
|  John Smith      | john@acme.com      | Editor  | Active | 2h ago   |
|  Jane Doe        | jane@acme.com      | Admin   | Active | 1h ago   |
|  Bob Wilson      | bob@techstart.io   | Viewer  | Active | 1d ago   |
|  Alice Brown     | alice@datacorp.com | Editor  | Invited| Never    |
|  Charlie Lee     | charlie@acme.com   | Viewer  | Banned | 5d ago   |
|                                                                      |
|  Showing 1-10 of 156 users    < 1 2 3 ... 16 >                      |
+----------------------------------------------------------------------+
```

### Table Columns

| Column       | Width    | Sortable | Filterable | Description                |
|--------------|----------|----------|------------|----------------------------|
| Checkbox     | 40px     | No       | No         | Bulk selection             |
| Avatar       | 40px     | No       | No         | User avatar                |
| Name         | Flexible | Yes      | No         | Display name               |
| Email        | Flexible | Yes      | Yes        | Email address              |
| Role         | 100px    | Yes      | Yes        | Assigned role              |
| Status       | 80px     | Yes      | Yes        | Active/Invited/Banned      |
| Tenant       | 120px    | Yes      | Yes        | Organization name          |
| Last Login   | 100px    | Yes      | No         | Last login time            |
| Actions      | 100px    | No       | No         | Edit / Delete / Ban        |

---

## User Creation Form

```
+----------------------------------------------------------------------+
|  Create User                                                         |
+----------------------------------------------------------------------+
|                                                                      |
|  Email:       [____________________________]                         |
|  First Name:  [____________________________]                         |
|  Last Name:   [____________________________]                         |
|  Password:    [____________________________]                         |
|  Confirm:     [____________________________]                         |
|                                                                      |
|  Role:        [Editor v]                                             |
|  Organization:[Acme Corp v]                                          |
|  Status:      (X) Active  ( ) Invited                                |
|                                                                      |
|  Options:                                                            |
|  [ ] Send welcome email                                              |
|  [ ] Require password change on first login                          |
|  [ ] Enable 2FA                                                      |
|                                                                      |
|         [Cancel]                 [Create User]                        |
+----------------------------------------------------------------------+
```

### Validation Rules

| Field      | Rule                                           | Error Message                    |
|------------|------------------------------------------------|----------------------------------|
| Email      | Required, valid email format, unique           | "Email is required" / "Invalid"  |
| First Name | Required, 2-50 characters                      | "Required, 2-50 chars"           |
| Last Name  | Required, 2-50 characters                      | "Required, 2-50 chars"           |
| Password   | Min 8 chars, 1 uppercase, 1 number, 1 special  | "Must meet complexity"           |
| Confirm    | Must match password                            | "Passwords do not match"         |
| Role       | Required, valid role                           | "Select a role"                  |

### Form Component

```tsx
// components/admin/UserCreateForm.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const createUserSchema = z.object({
  email: z.string().email('Invalid email address'),
  first_name: z.string().min(2, 'Min 2 characters').max(50),
  last_name: z.string().min(2, 'Min 2 characters').max(50),
  password: z.string().min(8, 'Min 8 characters')
    .regex(/[A-Z]/, 'Must contain uppercase')
    .regex(/[0-9]/, 'Must contain a number')
    .regex(/[^A-Za-z0-9]/, 'Must contain special character'),
  role: z.enum(['admin', 'editor', 'viewer']),
  tenant_id: z.string().optional(),
  send_welcome: z.boolean().default(true),
  require_password_change: z.boolean().default(false),
  enable_2fa: z.boolean().default(false),
});

type CreateUserFormData = z.infer<typeof createUserSchema>;

export function UserCreateForm({ onSubmit }: { onSubmit: (data: CreateUserFormData) => Promise<void> }) {
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<CreateUserFormData>({
    resolver: zodResolver(createUserSchema),
    defaultValues: { role: 'editor', send_welcome: true },
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 max-w-lg">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label>First Name</Label>
          <Input {...register('first_name')} />
          {errors.first_name && <p className="text-sm text-destructive">{errors.first_name.message}</p>}
        </div>
        <div>
          <Label>Last Name</Label>
          <Input {...register('last_name')} />
          {errors.last_name && <p className="text-sm text-destructive">{errors.last_name.message}</p>}
        </div>
      </div>
      <div>
        <Label>Email</Label>
        <Input type="email" {...register('email')} />
        {errors.email && <p className="text-sm text-destructive">{errors.email.message}</p>}
      </div>
      <div>
        <Label>Password</Label>
        <Input type="password" {...register('password')} />
        {errors.password && <p className="text-sm text-destructive">{errors.password.message}</p>}
      </div>
      <div>
        <Label>Role</Label>
        <Select {...register('role')}>
          <SelectItem value="admin">Admin</SelectItem>
          <SelectItem value="editor">Editor</SelectItem>
          <SelectItem value="viewer">Viewer</SelectItem>
        </Select>
      </div>
      <div>
        <Label>Organization</Label>
        <Select {...register('tenant_id')}>
          <SelectItem value="">None</SelectItem>
          {/* Dynamically populated from tenant list */}
        </Select>
      </div>
      <div className="space-y-2">
        <label className="flex items-center gap-2">
          <Checkbox {...register('send_welcome')} /> Send welcome email
        </label>
        <label className="flex items-center gap-2">
          <Checkbox {...register('require_password_change')} /> Require password change on first login
        </label>
      </div>
      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Creating...' : 'Create User'}
      </Button>
    </form>
  );
}
```

---

## User Edit Form

```
+----------------------------------------------------------------------+
|  Edit User: john@acme.com                                            |
+----------------------------------------------------------------------+
|                                                                      |
|  Profile:                                                            |
|  +----------------------------------------------------------------+  |
|  |  First Name:  [John                              ]             |  |
|  |  Last Name:   [Smith                             ]             |  |
|  |  Email:       [john@acme.com                     ] (verified)  |  |
|  |  Avatar:      [Upload Image]                                 |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Role & Status:                                                      |
|  +----------------------------------------------------------------+  |
|  |  Role:      [Editor v]                                         |  |
|  |  Status:    (X) Active  ( ) Invited  ( ) Banned                |  |
|  |  Tenant:    [Acme Corp v]                                     |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Security:                                                           |
|  +----------------------------------------------------------------+  |
|  |  [Reset Password]  [Send Password Reset Email]                 |  |
|  |  [Enable 2FA]      [Revoke All Sessions]                      |  |
|  |  [Ban User]        [Delete User]                               |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|         [Cancel]                 [Save Changes]                       |
+----------------------------------------------------------------------+
```

---

## User Detail Page

```
+----------------------------------------------------------------------+
|  <- Back to Users                                                    |
+----------------------------------------------------------------------+
|                                                                      |
|  John Smith (john@acme.com)                                          |
|  Role: Editor  |  Status: Active  |  Org: Acme Corp                  |
|  Joined: Jan 15, 2026  |  Last Login: 2 hours ago                   |
|                                                                      |
|  [Overview] [Activity] [Permissions] [Sessions] [API Keys]          |
+----------------------------------------------------------------------+
|                                                                      |
|  Overview Tab:                                                       |
|  +----------------------------------------------------------------+  |
|  |  Total Chats:     234                                         |  |
|  |  Total Messages:  4,567                                       |  |
|  |  Documents:       45 uploaded                                 |  |
|  |  Agents Created:  12                                          |  |
|  |  API Calls (30d): 1,234                                       |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Activity Log:                                                       |
|  +----------------------------------------------------------------+  |
|  |  Jul 16 10:28  Document Upload  report.pdf                     |  |
|  |  Jul 16 09:15  Agent Execute   Customer Bot                    |  |
|  |  Jul 15 16:30  Chat Message    "Summarize Q2 data"             |  |
|  |  Jul 15 11:00  Login           192.168.1.100 (Chrome/Mac)      |  |
|  |  Jul 14 14:22  Model Config    Llama 3 70B (temp: 0.7)        |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

---

## Organization (Tenant) Management

### Tenant List Table

```
+----------------------------------------------------------------------+
|  Organizations                                     [+ Create Org]     |
+----------------------------------------------------------------------+
|                                                                      |
|  Name           | Plan     | KYC    | Users | AI Usage | Revenue    |
|  ---------------+----------+--------+-------+----------+------------|
|  Acme Corp      | Pro      | Approved|  25   | 45,230   | $499/mo   |
|  TechStart Inc  | Enterprise| Approved|  12   | 89,120   | $999/mo   |
|  DataCorp       | Free     | Pending |   5   | 12,340   | $0/mo     |
|  GlobalTech     | Pro      | Review  |  18   | 34,560   | $499/mo   |
|  StartupXYZ     | Free     | Submitted|  3   |  2,100   | $0/mo     |
|                                                                      |
|  Showing 1-10 of 23 organizations                                   |
+----------------------------------------------------------------------+
```

### Tenant Columns

| Column       | Description                                   |
|--------------|-----------------------------------------------|
| Name         | Organization name and logo                    |
| Plan         | Current subscription plan                     |
| KYC          | KYC verification status                       |
| Users        | Number of active users                        |
| AI Usage     | Monthly AI requests count                     |
| Revenue      | Monthly recurring revenue                     |
| Domain       | Organization domain                           |
| Created      | Account creation date                         |
| Actions      | View / Edit / Manage                          |

---

## Tenant Creation Form

```
+----------------------------------------------------------------------+
|  Create Organization                                                 |
+----------------------------------------------------------------------+
|                                                                      |
|  Name:        [____________________________]                         |
|  Domain:      [____________________________]                         |
|  Plan:        [Free v]                                               |
|                                                                      |
|  Admin User:                                                         |
|  Email:       [____________________________]                         |
|  First Name:  [____________________________]                         |
|  Last Name:   [____________________________]                         |
|                                                                      |
|  Settings:                                                           |
|  Max Users:        [50]                                              |
|  Max AI Requests:  [10000/month]                                     |
|  Storage Limit:    [5GB]                                             |
|                                                                      |
|  Branding:                                                           |
|  Logo:           [Upload]                                            |
|  Primary Color:  [#3B82F6]                                           |
|                                                                      |
|         [Cancel]                 [Create Organization]                |
+----------------------------------------------------------------------+
```

---

## Tenant Detail Page

```
+----------------------------------------------------------------------+
|  <- Back to Organizations                                            |
+----------------------------------------------------------------------+
|                                                                      |
|  Acme Corp                                                           |
|  Plan: Pro  |  Domain: acme.com  |  Created: Jan 10, 2026           |
|                                                                      |
|  [Overview] [Users] [Settings] [Usage] [KYC] [Billing]              |
+----------------------------------------------------------------------+
|                                                                      |
|  Overview:                                                           |
|  +----------------------------------------------------------------+  |
|  |  Users: 25 active    |  Documents: 342                         |  |
|  |  AI Requests: 45,230 |  Storage: 3.2GB / 10GB                  |  |
|  |  Agent Executions: 890 |  Monthly Cost: $499                    |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Usage Chart (Last 30 Days):                                         |
|  +----------------------------------------------------------------+  |
|  |  AI Requests per Day                                          |  |
|  |  2000 |        *                                              |  |
|  |  1500 |   *  *   *  *                                        |  |
|  |  1000 |  * * * * * **                                        |  |
|  |   500 | * * * * * ****                                       |  |
|  |     0 +------------------                                    |  |
|  |       Week1  Week2  Week3  Week4                              |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Top Users:                                                          |
|  +----------------------------------------------------------------+  |
|  |  User            | Requests | Tokens    | Cost                 |  |
|  |------------------+----------+-----------+----------------------|  |
|  |  john@acme.com   | 12,450   | 3,240,000 | $12.45               |  |
|  |  jane@acme.com   | 8,230    | 2,100,000 | $8.23                |  |
|  |  bob@acme.com    | 5,120    | 1,280,000 | $5.12                |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

---

## KYC Management

### KYC List

```
+----------------------------------------------------------------------+
|  KYC Management                                                      |
+----------------------------------------------------------------------+
|  Tab: [Pending (3)] [Submitted (5)] [Under Review (2)] [Approved (18)] [Rejected (1)] |
|                                                                      |
|  Pending/Submitted KYC Applications:                                 |
|  +----------------------------------------------------------------+  |
|  |  Org           | Submitted    | Documents | Status      | Act  |  |
|  |----------------+--------------+-----------+-------------+------|  |
|  |  DataCorp      | Jul 15, 2026 | 4/4       | Submitted   | View |  |
|  |  StartupXYZ    | Jul 14, 2026 | 2/4       | Partial     | View |  |
|  |  NewVenture    | Jul 16, 2026 | 0/4       | Pending     | View |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

### KYC Status Tracking

```
KYC Lifecycle:

  Pending -> Submitted -> Under Review -> Approved / Rejected
     |           |              |              |
     v           v              v              v
  No docs    Docs uploaded   Admin reviewing  Final decision
  uploaded   (4 required)    (5 business days)
```

| Status      | Description                                    | Badge Color |
|-------------|------------------------------------------------|-------------|
| Pending     | Organization created, no documents uploaded    | Gray        |
| Submitted   | All required documents uploaded                | Blue        |
| Under Review| Admin is reviewing the documents               | Yellow      |
| Approved    | KYC verification passed                        | Green       |
| Rejected    | KYC verification failed                        | Red         |
| Expired     | KYC needs renewal (annual)                     | Orange      |

### KYC Status Component

```tsx
// components/admin/KYCStatusBadge.tsx
import { Badge } from '@/components/ui/badge';

const KYC_STATUS_CONFIG = {
  pending: { label: 'Pending', variant: 'outline' as const },
  submitted: { label: 'Submitted', variant: 'secondary' as const },
  'under-review': { label: 'Under Review', variant: 'default' as const },
  approved: { label: 'Approved', variant: 'default' as const },
  rejected: { label: 'Rejected', variant: 'destructive' as const },
  expired: { label: 'Expired', variant: 'destructive' as const },
};

export function KYCStatusBadge({ status }: { status: string }) {
  const config = KYC_STATUS_CONFIG[status as keyof typeof KYC_STATUS_CONFIG];
  return <Badge variant={config?.variant || 'outline'}>{config?.label || status}</Badge>;
}
```

---

## KYC Document Viewer

```
+----------------------------------------------------------------------+
|  KYC Review: DataCorp                                                |
+----------------------------------------------------------------------+
|                                                                      |
|  Organization: DataCorp                                              |
|  Submitted: Jul 15, 2026                                             |
|  Review Deadline: Jul 22, 2026                                       |
|                                                                      |
|  Required Documents:                                                 |
|  +----------------------------------------------------------------+  |
|  |  1. Business Registration     [View]  [Download]  Status: OK   |  |
|  |  2. Tax ID Certificate        [View]  [Download]  Status: OK   |  |
|  |  3. Proof of Address          [View]  [Download]  Status: OK   |  |
|  |  4. Director ID               [View]  [Download]  Status: OK   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Document Preview:                                                   |
|  +----------------------------------------------------------------+  |
|  |                                                                 |  |
|  |     [Document Preview Area]                                     |  |
|  |     PDF/Image viewer with zoom, pan, rotate controls            |  |
|  |                                                                 |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Review Decision:                                                    |
|  Notes: [_________________________________________________]         |
|          [_________________________________________________]         |
|                                                                      |
|  [Reject with Reason]  [Request Additional Info]  [Approve KYC]      |
+----------------------------------------------------------------------+
```

### KYC Review Component

```tsx
// components/admin/KYCReview.tsx
import { useState } from 'react';

interface KYCReviewProps {
  application: KYCApplication;
  onApprove: (notes: string) => Promise<void>;
  onReject: (reason: string) => Promise<void>;
  onRequestInfo: (message: string) => Promise<void>;
}

export function KYCReview({ application, onApprove, onReject, onRequestInfo }: KYCReviewProps) {
  const [notes, setNotes] = useState('');
  const [rejectReason, setRejectReason] = useState('');

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div>
          <h3 className="font-semibold mb-3">Required Documents</h3>
          {application.documents.map((doc) => (
            <div key={doc.id} className="flex items-center justify-between p-3 border rounded mb-2">
              <div>
                <p className="font-medium">{doc.type}</p>
                <p className="text-sm text-muted-foreground">{doc.filename}</p>
              </div>
              <div className="flex gap-2">
                <Button size="sm" variant="outline" onClick={() => window.open(doc.preview_url)}>
                  View
                </Button>
                <Button size="sm" variant="outline" onClick={() => window.open(doc.download_url)}>
                  Download
                </Button>
              </div>
            </div>
          ))}
        </div>
        <div>
          <h3 className="font-semibold mb-3">Document Preview</h3>
          <div className="border rounded-lg h-96 flex items-center justify-center bg-muted">
            <p className="text-muted-foreground">Select a document to preview</p>
          </div>
        </div>
      </div>
      <div className="border-t pt-4">
        <Label>Review Notes</Label>
        <Textarea value={notes} onChange={(e) => setNotes(e.target.value)}
          placeholder="Add review notes..." className="mt-2" />
        <div className="flex gap-2 mt-4">
          <Button variant="destructive" onClick={() => onReject(rejectReason)}>
            Reject
          </Button>
          <Button variant="outline" onClick={() => onRequestInfo(notes)}>
            Request Info
          </Button>
          <Button onClick={() => onApprove(notes)}>
            Approve KYC
          </Button>
        </div>
      </div>
    </div>
  );
}
```

---

## Plan Management

### Plans Overview

| Plan       | Price    | Users | AI Requests | Storage | Features                  |
|------------|----------|-------|-------------|---------|---------------------------|
| Free       | $0/mo    | 5     | 1,000/mo    | 1GB     | Basic chat, 1 agent       |
| Pro        | $49/mo   | 25    | 50,000/mo   | 10GB    | RAG, agents, priority     |
| Enterprise | $199/mo  | 100   | 500,000/mo  | 100GB   | Custom models, API, SSO   |
| Custom     | Variable | Varies| Variable    | Variable| All features, dedicated   |

### Plan Card

```
+------------------------------+
|  Pro                         |
|  $49/month                   |
|                              |
|  - Up to 25 users            |
|  - 50,000 AI requests/mo     |
|  - 10GB storage              |
|  - RAG & Agents              |
|  - Priority support          |
|  - Custom branding           |
|                              |
|  [Edit Plan] [View Tenants]  |
+------------------------------+
```

### Plan Editor

```tsx
// components/admin/PlanEditor.tsx
interface PlanEditorProps {
  plan: Plan;
  onSave: (data: PlanData) => Promise<void>;
}

export function PlanEditor({ plan, onSave }: PlanEditorProps) {
  const [form, setForm] = useState({
    name: plan.name,
    price: plan.price,
    max_users: plan.max_users,
    max_ai_requests: plan.max_ai_requests,
    max_storage_gb: plan.max_storage_gb,
    features: plan.features,
  });

  return (
    <div className="space-y-4 max-w-lg">
      <div>
        <Label>Plan Name</Label>
        <Input value={form.name}
          onChange={(e) => setForm({ ...form, name: e.target.value })} />
      </div>
      <div>
        <Label>Monthly Price ($)</Label>
        <Input type="number" value={form.price}
          onChange={(e) => setForm({ ...form, price: +e.target.value })} />
      </div>
      <div className="grid grid-cols-3 gap-4">
        <div>
          <Label>Max Users</Label>
          <Input type="number" value={form.max_users}
            onChange={(e) => setForm({ ...form, max_users: +e.target.value })} />
        </div>
        <div>
          <Label>Max AI Requests</Label>
          <Input type="number" value={form.max_ai_requests}
            onChange={(e) => setForm({ ...form, max_ai_requests: +e.target.value })} />
        </div>
        <div>
          <Label>Storage (GB)</Label>
          <Input type="number" value={form.max_storage_gb}
            onChange={(e) => setForm({ ...form, max_storage_gb: +e.target.value })} />
        </div>
      </div>
      <div>
        <Label>Features</Label>
        <div className="space-y-2 mt-2">
          {ALL_FEATURES.map((feature) => (
            <label key={feature} className="flex items-center gap-2">
              <Checkbox
                checked={form.features.includes(feature)}
                onCheckedChange={(c) => {
                  const features = c
                    ? [...form.features, feature]
                    : form.features.filter((f) => f !== feature);
                  setForm({ ...form, features });
                }}
              />
              {feature}
            </label>
          ))}
        </div>
      </div>
      <Button onClick={() => onSave(form)}>Save Plan</Button>
    </div>
  );
}
```

---

## Billing Overview

```
+----------------------------------------------------------------------+
|  Billing Overview                                                    |
+----------------------------------------------------------------------+
|                                                                      |
|  Revenue Summary:                                                    |
|  +----------------------------------------------------------------+  |
|  |  MRR: $8,900       |  ARR: $106,800                          |  |
|  |  Active Subs: 18   |  Pending Invoices: 2                    |  |
|  |  Collected (30d): $12,450 |  Outstanding: $1,200             |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Revenue by Plan:                                                    |
|  +----------------------------------------------------------------+  |
|  |  Enterprise: $3,996  ████████████████  (45%)                  |  |
|  |  Pro:        $3,484  ████████████      (39%)                  |  |
|  |  Free:       $0      -                   (0%)                  |  |
|  |  Custom:     $1,420  ██████            (16%)                  |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Recent Invoices:                                                    |
|  +----------------------------------------------------------------+  |
|  |  Org        | Amount | Status   | Due Date  | Actions           |  |
|  |-------------+--------+----------+-----------+-------------------|  |
|  |  Acme Corp  | $499   | Paid     | Jul 1     | [View] [Receipt]  |  |
|  |  TechStart  | $999   | Paid     | Jul 1     | [View] [Receipt]  |  |
|  |  GlobalTech | $499   | Pending  | Aug 1     | [View] [Remind]   |  |
|  |  NewVenture | $0     | N/A      | N/A       | Free plan         |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

---

## Usage Analytics

### Per-Tenant Usage

| Tenant       | AI Requests | Tokens Used | Avg Latency | Cost    | Trend |
|--------------|-------------|-------------|-------------|---------|-------|
| TechStart    | 89,120      | 22,400,000  | 115ms       | $112.00 | +15%  |
| Acme Corp    | 45,230      | 11,300,000  | 120ms       | $56.50  | +8%   |
| GlobalTech   | 34,560      | 8,640,000   | 130ms       | $43.20  | +22%  |
| DataCorp     | 12,340      | 3,080,000   | 125ms       | $15.40  | +45%  |
| StartupXYZ   | 2,100       | 520,000     | 140ms       | $6.50   | +10%  |

### Per-User Usage

| User           | Org      | Requests | Tokens   | Models Used       | Cost   |
|----------------|----------|----------|----------|-------------------|--------|
| john@acme.com  | Acme     | 12,450   | 3,240,000| Llama 3, GPT-4o   | $12.45 |
| jane@acme.com  | Acme     | 8,230    | 2,100,000| Llama 3           | $8.23  |
| bob@techstart  | TechStart| 15,600   | 3,900,000| GPT-4o, Embedding | $19.50 |

### Usage Chart Component

```tsx
// components/admin/UsageChart.tsx
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';

interface UsageChartProps {
  data: UsageDataPoint[];
  metric: 'requests' | 'tokens' | 'cost';
}

export function UsageChart({ data, metric }: UsageChartProps) {
  const labels = { requests: 'AI Requests', tokens: 'Tokens Used', cost: 'Cost ($)' };

  return (
    <div className="border rounded-lg p-4">
      <h4 className="font-medium mb-4">{labels[metric]} by Tenant</h4>
      <BarChart data={data} width={600} height={300}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="tenant_name" />
        <YAxis />
        <Tooltip />
        <Legend />
        <Bar dataKey={metric} fill="hsl(var(--primary))" radius={[4, 4, 0, 0]} />
      </BarChart>
    </div>
  );
}
```

---

## System Settings

```
+----------------------------------------------------------------------+
|  System Settings                                                     |
+----------------------------------------------------------------------+
|                                                                      |
|  General:                                                            |
|  +----------------------------------------------------------------+  |
|  |  System Name:           [Nexus AI                    ]         |  |
|  |  Default Language:      [English v]                            |  |
|  |  Default Timezone:      [UTC v]                                |  |
|  |  Registration Open:     [x] Allow new user registrations      |  |
|  |  Default Plan:          [Free v]                               |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Security:                                                           |
|  +----------------------------------------------------------------+  |
|  |  Session Timeout:       [30] minutes                           |  |
|  |  Max Login Attempts:    [5]                                    |  |
|  |  Lockout Duration:      [15] minutes                           |  |
|  |  Require 2FA for Admin: [x]                                    |  |
|  |  Password Min Length:   [8]                                    |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  AI Defaults:                                                        |
|  +----------------------------------------------------------------+  |
|  |  Default Model:         [Llama 3 70B v]                       |  |
|  |  Embedding Model:       [text-embedding-3-small v]             |  |
|  |  Max Tokens per Request:[4096]                                 |  |
|  |  Rate Limit (req/min):  [60]                                   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Storage:                                                            |
|  +----------------------------------------------------------------+  |
|  |  Max Upload Size:       [50] MB                                |  |
|  |  Total Storage Quota:   [500] GB                               |  |
|  |  Retention Period:      [365] days                             |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Save Settings]                                                     |
+----------------------------------------------------------------------+
```

---

## Feature Flags

```
+----------------------------------------------------------------------+
|  Feature Flags                                                       |
+----------------------------------------------------------------------+
|                                                                      |
|  Global Flags:                                                       |
|  +----------------------------------------------------------------+  |
|  |  Feature              | Enabled | Rollout | Actions            |  |
|  |-----------------------+---------+---------+--------------------|  |
|  |  RAG System           | Yes     | 100%    | [Edit] [Disable]   |  |
|  |  Agent Marketplace    | Yes     | 50%     | [Edit] [Disable]   |  |
|  |  Workflow Builder     | No      | 0%      | [Edit] [Enable]    |  |
|  |  Voice Chat           | No      | 0%      | [Edit] [Enable]    |  |
|  |  Custom Models        | Yes     | 100%    | [Edit] [Disable]   |  |
|  |  API Access           | Yes     | 100%    | [Edit] [Disable]   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Per-Tenant Overrides:                                               |
|  +----------------------------------------------------------------+  |
|  |  Tenant     | Feature        | Override | Value                 |  |
|  |-------------+----------------+----------+-----------------------|  |
|  |  TechStart  | Agent Market   | Custom   | 100%                  |  |
|  |  Acme Corp  | Workflow Bld   | Custom   | Enabled               |  |
|  |  DataCorp   | Custom Models  | Custom   | Disabled              |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [+ Add Flag]  [+ Add Tenant Override]                              |
+----------------------------------------------------------------------+
```

### Feature Flag Component

```tsx
// components/admin/FeatureFlags.tsx
interface FeatureFlag {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  rollout_percentage: number;
  tenant_overrides: TenantOverride[];
}

interface TenantOverride {
  tenant_id: string;
  tenant_name: string;
  override: boolean;
  value: boolean | number;
}

export function FeatureFlags() {
  const { data: flags, isLoading } = useQuery({
    queryKey: ['admin', 'feature-flags'],
    queryFn: () => api.post('/admin/feature-flags', {}),
  });

  const updateFlag = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<FeatureFlag> }) =>
      api.patch(`/admin/feature-flags/${id}`, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin', 'feature-flags'] }),
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Feature Flags</h3>
        <Button>+ Add Flag</Button>
      </div>
      <div className="space-y-2">
        {flags?.map((flag: FeatureFlag) => (
          <div key={flag.id} className="flex items-center justify-between p-4 border rounded-lg">
            <div className="flex-1">
              <div className="flex items-center gap-3">
                <Switch
                  checked={flag.enabled}
                  onCheckedChange={(enabled) =>
                    updateFlag.mutate({ id: flag.id, data: { enabled } })
                  }
                />
                <div>
                  <p className="font-medium">{flag.name}</p>
                  <p className="text-sm text-muted-foreground">{flag.description}</p>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="text-sm text-muted-foreground">
                Rollout: {flag.rollout_percentage}%
              </div>
              <Slider
                value={[flag.rollout_percentage]}
                onValueChange={([v]) =>
                  updateFlag.mutate({ id: flag.id, data: { rollout_percentage: v } })
                }
                max={100}
                step={5}
                className="w-32"
              />
              <Button size="sm" variant="outline">Edit</Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## API Integration

### Endpoints

| Method   | Endpoint                         | Description              |
|----------|----------------------------------|--------------------------|
| `POST`   | `/api/admin/users`               | List all users           |
| `POST`   | `/api/admin/users`               | Create user              |
| `POST`   | `/api/admin/users/:id`           | Get user details         |
| `PATCH`  | `/api/admin/users/:id`           | Update user              |
| `DELETE` | `/api/admin/users/:id`           | Delete user              |
| `POST`   | `/api/admin/users/:id/ban`       | Ban user                 |
| `POST`   | `/api/admin/users/:id/unban`     | Unban user               |
| `POST`   | `/api/admin/tenants`             | List tenants             |
| `POST`   | `/api/admin/tenants`             | Create tenant            |
| `POST`   | `/api/admin/tenants/:id`         | Get tenant details       |
| `PATCH`  | `/api/admin/tenants/:id`         | Update tenant            |
| `DELETE` | `/api/admin/tenants/:id`         | Delete tenant            |
| `POST`   | `/api/admin/kyc`                 | List KYC applications    |
| `POST`   | `/api/admin/kyc/:id`             | Get KYC details          |
| `PATCH`  | `/api/admin/kyc/:id/approve`     | Approve KYC              |
| `PATCH`  | `/api/admin/kyc/:id/reject`      | Reject KYC               |
| `POST`   | `/api/admin/billing`             | Get billing overview     |
| `POST`   | `/api/admin/billing/invoices`    | List invoices            |
| `POST`   | `/api/admin/billing/invoices/:id/send` | Resend invoice  |
| `POST`   | `/api/admin/plans`               | List plans               |
| `POST`   | `/api/admin/plans`               | Create plan              |
| `PATCH`  | `/api/admin/plans/:id`           | Update plan              |
| `POST`   | `/api/admin/usage`               | Get usage analytics      |
| `POST`   | `/api/admin/settings`            | Get system settings      |
| `PATCH`  | `/api/admin/settings`            | Update system settings   |
| `POST`   | `/api/admin/feature-flags`       | List feature flags       |
| `PATCH`  | `/api/admin/feature-flags/:id`   | Update feature flag      |

---

## Hooks

### useUsers

```typescript
// hooks/admin/useUsers.ts
export function useUsers(filters?: UserFilters) {
  const queryClient = useQueryClient();

  const users = useQuery({
    queryKey: ['admin-users', filters],
    queryFn: () => api.post('/admin/users', filters),
  });

  const createUser = useMutation({
    mutationFn: (data: CreateUserRequest) => api.post('/admin/users', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  });

  const updateUser = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<User> }) =>
      api.patch(`/admin/users/${id}`, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  });

  const deleteUser = useMutation({
    mutationFn: (id: string) => api.delete(`/admin/users/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  });

  const banUser = useMutation({
    mutationFn: (id: string) => api.post(`/admin/users/${id}/ban`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  });

  return { users, createUser, updateUser, deleteUser, banUser };
}
```

### useTenants

```typescript
// hooks/admin/useTenants.ts
export function useTenants(filters?: TenantFilters) {
  const queryClient = useQueryClient();

  const tenants = useQuery({
    queryKey: ['admin-tenants', filters],
    queryFn: () => api.post('/admin/tenants', filters),
  });

  const createTenant = useMutation({
    mutationFn: (data: CreateTenantRequest) => api.post('/admin/tenants', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-tenants'] }),
  });

  const updateTenant = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Tenant> }) =>
      api.patch(`/admin/tenants/${id}`, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-tenants'] }),
  });

  return { tenants, createTenant, updateTenant };
}
```

### useKYC

```typescript
// hooks/admin/useKYC.ts
export function useKYC(filters?: KYCFilters) {
  const queryClient = useQueryClient();

  const applications = useQuery({
    queryKey: ['admin-kyc', filters],
    queryFn: () => api.post('/admin/kyc', filters),
  });

  const approveKYC = useMutation({
    mutationFn: ({ id, notes }: { id: string; notes: string }) =>
      api.patch(`/admin/kyc/${id}/approve`, { notes }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-kyc'] }),
  });

  const rejectKYC = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      api.patch(`/admin/kyc/${id}/reject`, { reason }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-kyc'] }),
  });

  return { applications, approveKYC, rejectKYC };
}
```

### useBilling

```typescript
// hooks/admin/useBilling.ts
export function useBilling() {
  const overview = useQuery({
    queryKey: ['admin-billing'],
    queryFn: () => api.post('/admin/billing', {}),
  });

  const invoices = useQuery({
    queryKey: ['admin-invoices'],
    queryFn: () => api.post('/admin/billing/invoices', {}),
  });

  const sendInvoice = useMutation({
    mutationFn: (id: string) => api.post(`/admin/billing/invoices/${id}/send`),
  });

  return { overview, invoices, sendInvoice };
}
```

---

## Stores

### Admin Store (Zustand)

```typescript
// stores/admin/admin.ts
import { create } from 'zustand';

interface AdminState {
  users: User[];
  tenants: Tenant[];
  kycApplications: KYCApplication[];
  billing: BillingOverview;
  systemSettings: SystemSettings;
  featureFlags: FeatureFlag[];

  setUsers: (users: User[]) => void;
  updateUserInList: (id: string, data: Partial<User>) => void;
  removeUserFromList: (id: string) => void;
  setTenants: (tenants: Tenant[]) => void;
  updateTenantInList: (id: string, data: Partial<Tenant>) => void;
  setKYCApplications: (apps: KYCApplication[]) => void;
  updateKYCInList: (id: string, data: Partial<KYCApplication>) => void;
  setBilling: (billing: BillingOverview) => void;
  setSystemSettings: (settings: SystemSettings) => void;
  setFeatureFlags: (flags: FeatureFlag[]) => void;
  updateFeatureFlag: (id: string, data: Partial<FeatureFlag>) => void;
}

export const useAdminStore = create<AdminState>((set) => ({
  users: [],
  tenants: [],
  kycApplications: [],
  billing: {} as BillingOverview,
  systemSettings: {} as SystemSettings,
  featureFlags: [],

  setUsers: (users) => set({ users }),
  updateUserInList: (id, data) => set((state) => ({
    users: state.users.map((u) => u.id === id ? { ...u, ...data } : u),
  })),
  removeUserFromList: (id) => set((state) => ({
    users: state.users.filter((u) => u.id !== id),
  })),
  setTenants: (tenants) => set({ tenants }),
  updateTenantInList: (id, data) => set((state) => ({
    tenants: state.tenants.map((t) => t.id === id ? { ...t, ...data } : t),
  })),
  setKYCApplications: (kycApplications) => set({ kycApplications }),
  updateKYCInList: (id, data) => set((state) => ({
    kycApplications: state.kycApplications.map((k) =>
      k.id === id ? { ...k, ...data } : k
    ),
  })),
  setBilling: (billing) => set({ billing }),
  setSystemSettings: (systemSettings) => set({ systemSettings }),
  setFeatureFlags: (featureFlags) => set({ featureFlags }),
  updateFeatureFlag: (id, data) => set((state) => ({
    featureFlags: state.featureFlags.map((f) =>
      f.id === id ? { ...f, ...data } : f
    ),
  })),
}));
```

---

## Error Handling

### Error Types

| Error                    | Message                          | Action                        |
|--------------------------|----------------------------------|-------------------------------|
| User create failed       | "Could not create user"          | Show validation errors        |
| User update failed       | "Could not update user"          | Show conflict details         |
| User delete failed       | "Cannot delete user"             | Show related resources        |
| Tenant create failed     | "Organization already exists"    | Show name conflict            |
| KYC approve failed       | "KYC approval failed"            | Retry with notes              |
| Billing fetch failed     | "Failed to load billing data"    | Show cached data              |
| Settings save failed     | "Settings not saved"             | Show what changed             |
| Feature flag update fail | "Flag update failed"             | Revert toggle state           |
| Permission denied        | "Admin access required"          | Redirect to dashboard         |

### Optimistic Updates

```tsx
// Example optimistic update for user status change
const toggleUserStatus = useMutation({
  mutationFn: ({ id, status }: { id: string; status: string }) =>
    api.patch(`/admin/users/${id}`, { status }),
  onMutate: async ({ id, status }) => {
    await queryClient.cancelQueries({ queryKey: ['admin-users'] });
    const previous = queryClient.getQueryData(['admin-users']);
    queryClient.setQueryData(['admin-users'], (old: any) => ({
      ...old,
      data: old.data.map((u: User) => u.id === id ? { ...u, status } : u),
    }));
    return { previous };
  },
  onError: (_err, _vars, context) => {
    queryClient.setQueryData(['admin-users'], context?.previous);
    toast({ title: 'Failed to update user status', variant: 'destructive' });
  },
  onSettled: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
});
```

---

## Responsive Design

### Mobile Layout (<=640px)

```
+------------------+
| Admin Panel  [H] |
+------------------+
| [Hamburger Menu] |
+------------------+
|                  |
| Stats Grid:      |
| Users: 156       |
| Tenants: 23      |
| Revenue: $12,450 |
|                  |
| Recent Activity: |
| - User created   |
| - Tenant upgraded|
| - KYC submitted  |
+------------------+
```

### Responsive Breakpoints

| Breakpoint | Layout                                               |
|------------|------------------------------------------------------|
| <=640px    | Stacked cards, hamburger nav, full-width forms       |
| 641-1024px | 2-column grid, collapsible sidebar                   |
| 1025+      | Full sidebar, 3-column grid, split-panel details     |

---

## Admin Role Protection

### Route Guard

```tsx
// components/admin/AdminGuard.tsx
import { useAuth } from '@/hooks/auth/useAuth';
import { Navigate } from 'react-router-dom';

interface AdminGuardProps {
  children: React.ReactNode;
  requiredRole?: 'admin' | 'super_admin';
}

export function AdminGuard({ children, requiredRole = 'admin' }: AdminGuardProps) {
  const { user, isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (requiredRole === 'super_admin' && user?.role !== 'super_admin') {
    return <Navigate to="/admin" replace />;
  }

  if (user?.role !== 'admin' && user?.role !== 'super_admin') {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}
```

### Component Guard

```tsx
// components/admin/AdminOnly.tsx
import { useAuth } from '@/hooks/auth/useAuth';

export function AdminOnly({ children, fallback }: {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}) {
  const { user } = useAuth();

  if (user?.role !== 'admin' && user?.role !== 'super_admin') {
    return fallback ? <>{fallback}</> : null;
  }

  return <>{children}</>;
}

// Usage
<AdminOnly fallback={<p className="text-muted-foreground">Admin access required</p>}>
  <DeleteUserButton userId={userId} />
</AdminOnly>
```

### Route Configuration

```tsx
// routes/admin.tsx
import { AdminPanel } from '@/pages/admin/AdminPanel';
import { AdminGuard } from '@/components/admin/AdminGuard';
import { lazy } from 'react';

const DashboardPage = lazy(() => import('@/pages/admin/DashboardPage'));
const UsersPage = lazy(() => import('@/pages/admin/UsersPage'));
const TenantsPage = lazy(() => import('@/pages/admin/TenantsPage'));
const KYCPage = lazy(() => import('@/pages/admin/KYCPage'));
const BillingPage = lazy(() => import('@/pages/admin/BillingPage'));
const SettingsPage = lazy(() => import('@/pages/admin/SettingsPage'));
const FeatureFlagsPage = lazy(() => import('@/pages/admin/FeatureFlagsPage'));

export const adminRoutes = {
  path: '/admin',
  element: (
    <AdminGuard>
      <AdminPanel />
    </AdminGuard>
  ),
  children: [
    { index: true, element: <DashboardPage /> },
    { path: 'users', element: <UsersPage /> },
    { path: 'users/:id', element: <UserDetailPage /> },
    { path: 'tenants', element: <TenantsPage /> },
    { path: 'tenants/:id', element: <TenantDetailPage /> },
    { path: 'kyc', element: <KYCPage /> },
    { path: 'billing', element: <BillingPage /> },
    { path: 'plans', element: <PlansPage /> },
    { path: 'settings', element: <SettingsPage /> },
    { path: 'flags', element: <FeatureFlagsPage /> },
  ],
};
```
