# Settings

## Table of Contents

- [Overview](#overview)
- [Settings Page Layout](#settings-page-layout)
- [User Profile Settings](#user-profile-settings)
- [Password Change Form](#password-change-form)
- [Notification Settings](#notification-settings)
- [Theme Settings](#theme-settings)
- [Language Settings](#language-settings)
- [Tenant Settings (Admin)](#tenant-settings-admin)
- [Tenant AI Policy Settings](#tenant-ai-policy-settings)
- [Tenant Branding Settings](#tenant-branding-settings)
- [API Key Management](#api-key-management)
- [Webhook Configuration](#webhook-configuration)
- [Developer Settings](#developer-settings)
- [Security Settings](#security-settings)
- [Data Retention Settings](#data-retention-settings)
- [Integration Settings](#integration-settings)
- [API Integration](#api-integration)
- [Hooks](#hooks)
- [Stores](#stores)
- [Error Handling](#error-handling)
- [Responsive Design](#responsive-design)

---

## Overview

The Settings page provides users and administrators with configuration controls for their profile, notifications, security, API keys, integrations, and tenant-level settings. It uses a tabbed navigation pattern with sections grouped by category.

```
+------------------------------------------------------------------+
|  Settings                                                         |
+----------+-------------------------------------------------------+
|          |                                                       |
| Profile  |  Profile Settings                                    |
| Security |                                                     |
| Notif.   |  Avatar:   [Avatar Image]  [Change] [Remove]       |
| Theme    |                                                     |
| Language |  Name:     [John Smith                     ]        |
| API Keys |  Email:    [john@acme.com                  ]        |
| Webhooks |  Bio:      [Senior engineer at Acme Corp   ]        |
| Dev      |           [working on AI integrations...    ]        |
| Integr.  |                                                     |
| Data     |  Timezone: [UTC v]                                   |
| (Admin)  |  Date fmt: [YYYY-MM-DD v]                            |
| Ten.AI   |                                                     |
| Brand    |  [Save Profile]                                     |
| Billing  |                                                     |
+----------+-------------------------------------------------------+
```

---

## Settings Page Layout

### Component Tree

```
SettingsPage
+-- SettingsSidebar
|   +-- SectionGroup: Personal
|   |   +-- NavItem: Profile
|   |   +-- NavItem: Security
|   |   +-- NavItem: Notifications
|   |   +-- NavItem: Appearance
|   |   +-- NavItem: Language
|   +-- SectionGroup: Developer
|   |   +-- NavItem: API Keys
|   |   +-- NavItem: Webhooks
|   |   +-- NavItem: Developer
|   |   +-- NavItem: Integrations
|   +-- SectionGroup: Admin (conditional)
|   |   +-- NavItem: Organization
|   |   +-- NavItem: AI Policy
|   |   +-- NavItem: Branding
|   |   +-- NavItem: Data Retention
|   |   +-- NavItem: Billing
+-- SettingsContent
    +-- SettingsHeader (breadcrumb, title)
    +-- SettingsSections
        +-- ProfileSection
        +-- SecuritySection
        +-- NotificationsSection
        +-- ThemeSection
        +-- LanguageSection
        +-- APIKeysSection
        +-- WebhooksSection
        +-- DeveloperSection
        +-- IntegrationsSection
        +-- TenantSection (admin)
        +-- AIPolicySection (admin)
        +-- BrandingSection (admin)
        +-- DataRetentionSection (admin)
        +-- BillingSection (admin)
```

### Layout Code

```tsx
// pages/settings/SettingsPage.tsx
import { SettingsSidebar } from '@/components/settings/SettingsSidebar';
import { useAuth } from '@/hooks/auth/useAuth';
import { ProfileSection } from '@/components/settings/sections/ProfileSection';
import { SecuritySection } from '@/components/settings/sections/SecuritySection';
import { NotificationsSection } from '@/components/settings/sections/NotificationsSection';
import { ThemeSection } from '@/components/settings/sections/ThemeSection';
import { LanguageSection } from '@/components/settings/sections/LanguageSection';
import { APIKeysSection } from '@/components/settings/sections/APIKeysSection';
import { WebhooksSection } from '@/components/settings/sections/WebhooksSection';
import { DeveloperSection } from '@/components/settings/sections/DeveloperSection';
import { IntegrationsSection } from '@/components/settings/sections/IntegrationsSection';
import { TenantSection } from '@/components/settings/sections/TenantSection';
import { AIPolicySection } from '@/components/settings/sections/AIPolicySection';
import { BrandingSection } from '@/components/settings/sections/BrandingSection';
import { DataRetentionSection } from '@/components/settings/sections/DataRetentionSection';
import { BillingSection } from '@/components/settings/sections/BillingSection';

const SECTION_MAP: Record<string, React.ComponentType> = {
  profile: ProfileSection,
  security: SecuritySection,
  notifications: NotificationsSection,
  appearance: ThemeSection,
  language: LanguageSection,
  'api-keys': APIKeysSection,
  webhooks: WebhooksSection,
  developer: DeveloperSection,
  integrations: IntegrationsSection,
  tenant: TenantSection,
  'ai-policy': AIPolicySection,
  branding: BrandingSection,
  'data-retention': DataRetentionSection,
  billing: BillingSection,
};

export function SettingsPage() {
  const [activeSection, setActiveSection] = useState('profile');
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin' || user?.role === 'super_admin';

  const ActiveComponent = SECTION_MAP[activeSection];

  return (
    <div className="flex h-screen bg-background">
      <SettingsSidebar
        activeSection={activeSection}
        onSectionChange={setActiveSection}
        isAdmin={isAdmin}
      />
      <main className="flex-1 overflow-y-auto">
        <div className="max-w-3xl mx-auto p-6">
          {ActiveComponent && <ActiveComponent />}
        </div>
      </main>
    </div>
  );
}
```

---

## User Profile Settings

### Profile Form

```
+----------------------------------------------------------------------+
|  Profile Settings                                                    |
+----------------------------------------------------------------------+
|                                                                      |
|  Avatar:                                                             |
|  +------+                                                            |
|  |      |   [Upload Photo]  [Remove]                                 |
|  |  JS  |   Recommended: 256x256px, JPG or PNG, max 2MB             |
|  |      |                                                            |
|  +------+                                                            |
|                                                                      |
|  Name:                                                               |
|  +----------------------------------------------------------------+  |
|  |  First Name:  [John                              ]             |  |
|  |  Last Name:   [Smith                             ]             |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Email:  john@acme.com  [Verified] [Change Email]                   |
|                                                                      |
|  Bio:                                                                |
|  +----------------------------------------------------------------+  |
|  |  Senior engineer at Acme Corp working on AI integrations...    |  |
|  |                                                                 |  |
|  +----------------------------------------------------------------+  |
|  Max 200 characters. 45/200 used.                                   |
|                                                                      |
|  Timezone:                                                           |
|  [UTC v]                                                             |
|                                                                      |
|  Date Format:                                                        |
|  [YYYY-MM-DD v]                                                      |
|                                                                      |
|  [Save Profile]                                                      |
+----------------------------------------------------------------------+
```

### Profile Component

```tsx
// components/settings/sections/ProfileSection.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useProfile } from '@/hooks/settings/useProfile';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { toast } from 'sonner';

const profileSchema = z.object({
  first_name: z.string().min(2, 'Min 2 characters').max(50),
  last_name: z.string().min(2, 'Min 2 characters').max(50),
  bio: z.string().max(200).optional(),
  timezone: z.string(),
  date_format: z.string(),
});

type ProfileFormData = z.infer<typeof profileSchema>;

export function ProfileSection() {
  const { profile, updateProfile, uploadAvatar, removeAvatar } = useProfile();
  const { register, handleSubmit, formState: { errors, isSubmitting, isDirty } } =
    useForm<ProfileFormData>({
      resolver: zodResolver(profileSchema),
      values: {
        first_name: profile?.first_name || '',
        last_name: profile?.last_name || '',
        bio: profile?.bio || '',
        timezone: profile?.timezone || 'UTC',
        date_format: profile?.date_format || 'YYYY-MM-DD',
      },
    });

  const onSubmit = async (data: ProfileFormData) => {
    try {
      await updateProfile(data);
      toast({ title: 'Profile updated successfully' });
    } catch {
      toast({ title: 'Failed to update profile', variant: 'destructive' });
    }
  };

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      if (file.size > 2 * 1024 * 1024) {
        toast({ title: 'File too large. Max 2MB.', variant: 'destructive' });
        return;
      }
      await uploadAvatar(file);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <h2 className="text-2xl font-bold">Profile Settings</h2>

      <div className="flex items-center gap-4">
        <Avatar className="w-20 h-20">
          <AvatarImage src={profile?.avatar_url} />
          <AvatarFallback>
            {profile?.first_name?.[0]}{profile?.last_name?.[0]}
          </AvatarFallback>
        </Avatar>
        <div className="space-y-2">
          <label className="cursor-pointer">
            <Button variant="outline" size="sm" asChild>
              <span>Upload Photo</span>
            </Button>
            <input type="file" accept="image/*" className="hidden" onChange={handleAvatarUpload} />
          </label>
          <Button variant="ghost" size="sm" onClick={removeAvatar}>Remove</Button>
          <p className="text-xs text-muted-foreground">
            Recommended: 256x256px, JPG or PNG, max 2MB
          </p>
        </div>
      </div>

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
        <div className="flex items-center gap-2">
          <Input value={profile?.email || ''} disabled />
          <Badge variant="secondary">Verified</Badge>
          <Button variant="outline" size="sm">Change</Button>
        </div>
      </div>

      <div>
        <Label>Bio</Label>
        <Textarea {...register('bio')} maxLength={200} />
        <p className="text-xs text-muted-foreground mt-1">
          {(profile?.bio || '').length}/200 characters
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label>Timezone</Label>
          <Select {...register('timezone')}>
            <SelectItem value="UTC">UTC</SelectItem>
            <SelectItem value="America/New_York">Eastern Time</SelectItem>
            <SelectItem value="America/Chicago">Central Time</SelectItem>
            <SelectItem value="America/Denver">Mountain Time</SelectItem>
            <SelectItem value="America/Los_Angeles">Pacific Time</SelectItem>
            <SelectItem value="Europe/London">London</SelectItem>
            <SelectItem value="Europe/Berlin">Berlin</SelectItem>
            <SelectItem value="Asia/Tokyo">Tokyo</SelectItem>
            <SelectItem value="Asia/Shanghai">Shanghai</SelectItem>
          </Select>
        </div>
        <div>
          <Label>Date Format</Label>
          <Select {...register('date_format')}>
            <SelectItem value="YYYY-MM-DD">YYYY-MM-DD</SelectItem>
            <SelectItem value="MM/DD/YYYY">MM/DD/YYYY</SelectItem>
            <SelectItem value="DD/MM/YYYY">DD/MM/YYYY</SelectItem>
            <SelectItem value="DD.MM.YYYY">DD.MM.YYYY</SelectItem>
          </Select>
        </div>
      </div>

      <Button type="submit" disabled={isSubmitting || !isDirty}>
        {isSubmitting ? 'Saving...' : 'Save Profile'}
      </Button>
    </form>
  );
}
```

---

## Password Change Form

```
+----------------------------------------------------------------------+
|  Change Password                                                     |
+----------------------------------------------------------------------+
|                                                                      |
|  Current Password:                                                   |
|  [••••••••••••••••]                                                  |
|                                                                      |
|  New Password:                                                       |
|  [••••••••••••••••]                                                  |
|                                                                      |
|  Confirm New Password:                                               |
|  [••••••••••••••••]                                                  |
|                                                                      |
|  Password Requirements:                                              |
|  +----------------------------------------------------------------+  |
|  |  [x] At least 8 characters                                    |  |
|  |  [x] Contains uppercase letter                                |  |
|  |  [ ] Contains lowercase letter                                |  |
|  |  [ ] Contains a number                                        |  |
|  |  [x] Contains special character                               |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Password Strength: Strong  [==================]                     |
|                                                                      |
|  [Cancel]  [Change Password]                                         |
+----------------------------------------------------------------------+
```

### Password Component

```tsx
// components/settings/sections/SecuritySection.tsx
import { useState, useMemo } from 'react';
import { useSettings } from '@/hooks/settings/useSettings';

export function SecuritySection() {
  const { changePassword } = useSettings();
  const [form, setForm] = useState({
    current_password: '',
    new_password: '',
    confirm_password: '',
  });
  const [showCurrent, setShowCurrent] = useState(false);
  const [showNew, setShowNew] = useState(false);

  const passwordChecks = useMemo(() => ({
    length: form.new_password.length >= 8,
    uppercase: /[A-Z]/.test(form.new_password),
    lowercase: /[a-z]/.test(form.new_password),
    number: /[0-9]/.test(form.new_password),
    special: /[^A-Za-z0-9]/.test(form.new_password),
  }), [form.new_password]);

  const strength = Object.values(passwordChecks).filter(Boolean).length;
  const strengthLabels = ['Very Weak', 'Weak', 'Fair', 'Good', 'Strong'];
  const strengthColors = ['bg-red-500', 'bg-orange-500', 'bg-yellow-500', 'bg-blue-500', 'bg-green-500'];

  const isValid = form.current_password &&
    Object.values(passwordChecks).every(Boolean) &&
    form.new_password === form.confirm_password;

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Security Settings</h2>

      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Change Password</h3>
        <div>
          <Label>Current Password</Label>
          <div className="relative">
            <Input
              type={showCurrent ? 'text' : 'password'}
              value={form.current_password}
              onChange={(e) => setForm({ ...form, current_password: e.target.value })}
            />
            <Button variant="ghost" size="sm" className="absolute right-0 top-0"
              onClick={() => setShowCurrent(!showCurrent)}>
              {showCurrent ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </Button>
          </div>
        </div>
        <div>
          <Label>New Password</Label>
          <div className="relative">
            <Input
              type={showNew ? 'text' : 'password'}
              value={form.new_password}
              onChange={(e) => setForm({ ...form, new_password: e.target.value })}
            />
            <Button variant="ghost" size="sm" className="absolute right-0 top-0"
              onClick={() => setShowNew(!showNew)}>
              {showNew ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </Button>
          </div>
        </div>
        <div>
          <Label>Confirm New Password</Label>
          <Input
            type="password"
            value={form.confirm_password}
            onChange={(e) => setForm({ ...form, confirm_password: e.target.value })}
          />
          {form.confirm_password && form.new_password !== form.confirm_password && (
            <p className="text-sm text-destructive">Passwords do not match</p>
          )}
        </div>

        {form.new_password && (
          <div className="space-y-2">
            <div className="flex gap-1">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className={`h-2 flex-1 rounded ${
                  i < strength ? strengthColors[strength - 1] : 'bg-muted'
                }`} />
              ))}
            </div>
            <p className="text-sm text-muted-foreground">
              Strength: {strengthLabels[strength - 1] || 'Very Weak'}
            </p>
            <div className="space-y-1">
              {Object.entries(passwordChecks).map(([key, passed]) => (
                <p key={key} className={`text-sm ${passed ? 'text-green-600' : 'text-muted-foreground'}`}>
                  {passed ? 'Y' : ' '} {PASSWORD_LABELS[key]}
                </p>
              ))}
            </div>
          </div>
        )}

        <Button disabled={!isValid} onClick={() => changePassword(form)}>
          Change Password
        </Button>
      </div>
    </div>
  );
}
```

---

## Notification Settings

```
+----------------------------------------------------------------------+
|  Notification Settings                                               |
+----------------------------------------------------------------------+
|                                                                      |
|  Email Notifications:                                                |
|  +----------------------------------------------------------------+  |
|  |  [x] Chat responses received                                  |  |
|  |  [x] Agent execution completed                                |  |
|  |  [x] Document processing completed                             |  |
|  |  [ ] Weekly usage summary                                      |  |
|  |  [x] Security alerts                                           |  |
|  |  [ ] Marketing and product updates                             |  |
|  |  [x] Billing alerts (payment due, invoice)                     |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Push Notifications:                                                 |
|  +----------------------------------------------------------------+  |
|  |  [x] Chat responses received                                  |  |
|  |  [x] Agent execution completed                                |  |
|  |  [ ] Document processing completed                             |  |
|  |  [ ] Workflow approval required                                |  |
|  |  [x] Security alerts                                           |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  In-App Notifications:                                               |
|  +----------------------------------------------------------------+  |
|  |  [x] All notifications                                         |  |
|  |  Sound: [x] Play sound on new notification                    |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Email Digest:                                                       |
|  +----------------------------------------------------------------+  |
|  |  Frequency: [Daily v]                                          |  |
|  |  Time:      [09:00]                                             |  |
|  |  Timezone:  [UTC v]                                             |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Save Notification Settings]                                        |
+----------------------------------------------------------------------+
```

### Notification Component

```tsx
// components/settings/sections/NotificationsSection.tsx
import { useSettings } from '@/hooks/settings/useSettings';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Select, SelectItem } from '@/components/ui/select';
import { Button } from '@/components/ui/button';

interface NotificationPrefs {
  email: Record<string, boolean>;
  push: Record<string, boolean>;
  in_app: { enabled: boolean; sound: boolean };
  digest: { frequency: string; time: string; timezone: string };
}

export function NotificationsSection() {
  const { preferences, updatePreferences } = useSettings();
  const [prefs, setPrefs] = useState<NotificationPrefs>(preferences?.notifications || {
    email: {},
    push: {},
    in_app: { enabled: true, sound: true },
    digest: { frequency: 'daily', time: '09:00', timezone: 'UTC' },
  });

  const toggleEmail = (key: string) => setPrefs((p) => ({
    ...p, email: { ...p.email, [key]: !p.email[key] },
  }));

  const togglePush = (key: string) => setPrefs((p) => ({
    ...p, push: { ...p.push, [key]: !p.push[key] },
  }));

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Notification Settings</h2>

      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Email Notifications</h3>
        {EMAIL_NOTIFICATIONS.map((item) => (
          <div key={item.key} className="flex items-center justify-between">
            <Label>{item.label}</Label>
            <Switch checked={!!prefs.email[item.key]} onCheckedChange={() => toggleEmail(item.key)} />
          </div>
        ))}
      </div>

      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Push Notifications</h3>
        {PUSH_NOTIFICATIONS.map((item) => (
          <div key={item.key} className="flex items-center justify-between">
            <Label>{item.label}</Label>
            <Switch checked={!!prefs.push[item.key]} onCheckedChange={() => togglePush(item.key)} />
          </div>
        ))}
      </div>

      <div className="space-y-4">
        <h3 className="text-lg font-semibold">In-App Notifications</h3>
        <div className="flex items-center justify-between">
          <Label>Enable In-App Notifications</Label>
          <Switch checked={prefs.in_app.enabled}
            onCheckedChange={(c) => setPrefs({ ...prefs, in_app: { ...prefs.in_app, enabled: c } })} />
        </div>
        <div className="flex items-center justify-between">
          <Label>Play Sound</Label>
          <Switch checked={prefs.in_app.sound}
            onCheckedChange={(c) => setPrefs({ ...prefs, in_app: { ...prefs.in_app, sound: c } })} />
        </div>
      </div>

      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Email Digest</h3>
        <div className="grid grid-cols-3 gap-4">
          <div>
            <Label>Frequency</Label>
            <Select value={prefs.digest.frequency}
              onValueChange={(v) => setPrefs({ ...prefs, digest: { ...prefs.digest, frequency: v } })}>
              <SelectItem value="realtime">Real-time</SelectItem>
              <SelectItem value="daily">Daily</SelectItem>
              <SelectItem value="weekly">Weekly</SelectItem>
              <SelectItem value="never">Never</SelectItem>
            </Select>
          </div>
          <div>
            <Label>Time</Label>
            <Input type="time" value={prefs.digest.time}
              onChange={(e) => setPrefs({ ...prefs, digest: { ...prefs.digest, time: e.target.value } })} />
          </div>
          <div>
            <Label>Timezone</Label>
            <Select value={prefs.digest.timezone}
              onValueChange={(v) => setPrefs({ ...prefs, digest: { ...prefs.digest, timezone: v } })}>
              <SelectItem value="UTC">UTC</SelectItem>
              <SelectItem value="America/New_York">Eastern</SelectItem>
              <SelectItem value="America/Los_Angeles">Pacific</SelectItem>
            </Select>
          </div>
        </div>
      </div>

      <Button onClick={() => updatePreferences({ notifications: prefs })}>
        Save Notification Settings
      </Button>
    </div>
  );
}
```

---

## Theme Settings

```
+----------------------------------------------------------------------+
|  Appearance                                                          |
+----------------------------------------------------------------------+
|                                                                      |
|  Theme:                                                              |
|  +----------------------------------------------------------------+  |
|  |  (X) Light           ( ) Dark             ( ) System           |  |
|  |                                                                 |  |
|  |  [Light Preview]     [Dark Preview]      [Auto Preview]        |  |
|  |  +---------------+  +---------------+  +---------------+       |  |
|  |  | White bg      |  | Dark bg       |  | Matches OS    |       |  |
|  |  | Dark text     |  | Light text    |  | preference    |       |  |
|  |  +---------------+  +---------------+  +---------------+       |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Accent Color:                                                       |
|  +----------------------------------------------------------------+  |
|  |  [#3B82F6]  [#8B5CF6]  [#EC4899]  [#10B981]  [#F59E0B]        |  |
|  |  Blue        Purple      Pink        Green       Amber          |  |
|  |                                                                 |  |
|  |  Custom: [________] [Pick]                                      |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Font Size:                                                          |
|  +----------------------------------------------------------------+  |
|  |  [Small]  [X] Medium  [Large]                                   |  |
|  |   14px      16px        18px                                    |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Compact Mode:                                                       |
|  [ ] Enable compact mode (reduces spacing and padding)               |
|                                                                      |
|  Sidebar Position:                                                   |
|  (X) Left  ( ) Right                                                |
|                                                                      |
|  [Save Appearance]                                                   |
+----------------------------------------------------------------------+
```

### Theme Component

```tsx
// components/settings/sections/ThemeSection.tsx
import { useTheme } from 'next-themes';
import { useSettings } from '@/hooks/settings/useSettings';
import { Card } from '@/components/ui/card';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';

const ACCENT_COLORS = [
  { name: 'Blue', value: '#3B82F6', className: 'bg-blue-500' },
  { name: 'Purple', value: '#8B5CF6', className: 'bg-purple-500' },
  { name: 'Pink', value: '#EC4899', className: 'bg-pink-500' },
  { name: 'Green', value: '#10B981', className: 'bg-emerald-500' },
  { name: 'Amber', value: '#F59E0B', className: 'bg-amber-500' },
];

export function ThemeSection() {
  const { theme, setTheme } = useTheme();
  const { preferences, updatePreferences } = useSettings();
  const [accentColor, setAccentColor] = useState(preferences?.accentColor || '#3B82F6');
  const [fontSize, setFontSize] = useState(preferences?.fontSize || 'medium');
  const [compact, setCompact] = useState(preferences?.compactMode || false);

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Appearance</h2>

      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Theme</h3>
        <RadioGroup value={theme} onValueChange={setTheme} className="grid grid-cols-3 gap-4">
          {['light', 'dark', 'system'].map((t) => (
            <Label key={t} htmlFor={`theme-${t}`}>
              <Card className={`p-4 cursor-pointer hover:border-primary
                ${theme === t ? 'border-2 border-primary' : ''}`}>
                <RadioGroupItem value={t} id={`theme-${t}`} className="sr-only" />
                <div className={`w-full h-16 rounded mb-2 ${
                  t === 'light' ? 'bg-white border' :
                  t === 'dark' ? 'bg-gray-900 border' :
                  'bg-gradient-to-r from-white to-gray-900 border'
                }`} />
                <p className="text-sm font-medium capitalize text-center">{t}</p>
              </Card>
            </Label>
          ))}
        </RadioGroup>
      </div>

      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Accent Color</h3>
        <div className="flex gap-3">
          {ACCENT_COLORS.map((color) => (
            <button
              key={color.value}
              className={`w-10 h-10 rounded-full ${color.className}
                ${accentColor === color.value ? 'ring-2 ring-offset-2 ring-primary' : ''}`}
              onClick={() => setAccentColor(color.value)}
              aria-label={`${color.name} accent color`}
            />
          ))}
        </div>
      </div>

      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Font Size</h3>
        <RadioGroup value={fontSize} onValueChange={setFontSize} className="flex gap-4">
          {['small', 'medium', 'large'].map((size) => (
            <Label key={size} className="flex items-center gap-2 cursor-pointer">
              <RadioGroupItem value={size} />
              <span className="capitalize">{size}</span>
            </Label>
          ))}
        </RadioGroup>
      </div>

      <div className="flex items-center justify-between">
        <Label>Compact Mode</Label>
        <Switch checked={compact} onCheckedChange={setCompact} />
      </div>

      <Button onClick={() => updatePreferences({ accentColor, fontSize, compactMode: compact })}>
        Save Appearance
      </Button>
    </div>
  );
}
```

---

## Language Settings

```
+----------------------------------------------------------------------+
|  Language & Region                                                   |
+----------------------------------------------------------------------+
|                                                                      |
|  Language:                                                           |
|  [English v]                                                         |
|  +----------------------------------------------------------------+  |
|  |  English                          |  Bahasa Indonesia         |  |
|  |  Bahasa Melayu                    |  Deutsch                 |  |
|  |  Espa~nol                         |  Fran~cais               |  |
|  |  Italiano                         |  Nederlands              |  |
|  |  Portugues                        |  Russian                 |  |
|  |  Turkish                          |  Arabic (RTL)            |  |
|  |  Chinese (Simplified)             |  Chinese (Traditional)   |  |
|  |  Japanese                         |  Korean                  |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  RTL Support:                                                        |
|  [x] Enable right-to-left layout for Arabic, Hebrew, etc.           |
|                                                                      |
|  Region:                                                             |
|  [United States v]                                                   |
|                                                                      |
|  [Save Language Settings]                                            |
+----------------------------------------------------------------------+
```

---

## Tenant Settings (Admin)

```
+----------------------------------------------------------------------+
|  Organization Settings  [Admin Only]                                 |
+----------------------------------------------------------------------+
|                                                                      |
|  General:                                                            |
|  +----------------------------------------------------------------+  |
|  |  Name:     [Acme Corp                              ]            |  |
|  |  Domain:   [acme.com                               ]            |  |
|  |  Logo:     [Upload]  [Remove]                                  |  |
|  |  Contact:  [admin@acme.com                        ]            |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Usage Limits:                                                       |
|  +----------------------------------------------------------------+  |
|  |  Max Users:         [50]                                       |  |
|  |  Max AI Requests:   [50000/month]                              |  |
|  |  Storage Limit:     [10GB]                                     |  |
|  |  Max Agents:        [20]                                       |  |
|  |  Max Workflows:     [10]                                       |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Plan: Pro ($49/month)  [Change Plan]                                |
|                                                                      |
|  [Save Organization Settings]                                        |
+----------------------------------------------------------------------+
```

---

## Tenant AI Policy Settings

```
+----------------------------------------------------------------------+
|  AI Policy Settings  [Admin Only]                                    |
+----------------------------------------------------------------------+
|                                                                      |
|  Safety Guardrails:                                                  |
|  +----------------------------------------------------------------+  |
|  |  [x] Block harmful content (violence, hate, illegal)          |  |
|  |  [x] Block PII in AI outputs (emails, phones, SSN)           |  |
|  |  [x] Block prompt injection attempts                          |  |
|  |  [x] Content filtering for profanity                          |  |
|  |  [ ] Watermark AI-generated content                            |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Tool Policies:                                                      |
|  +----------------------------------------------------------------+  |
|  |  Tool              | Allowed | Rate Limit | Approval Required  |  |
|  |--------------------+---------+------------+--------------------|  |
|  |  Web Search        | Yes     | 100/hour   | No                 |  |
|  |  Code Execution    | Yes     | 50/hour    | Yes                |  |
|  |  File System       | Yes     | Unlimited  | No                 |  |
|  |  Database Query    | Yes     | 200/hour   | No                 |  |
|  |  Email Send        | Yes     | 10/hour    | Yes                |  |
|  |  API Call (ext)    | Yes     | 500/hour   | No                 |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Model Restrictions:                                                 |
|  +----------------------------------------------------------------+  |
|  |  [x] Enforce model routing rules                              |  |
|  |  [x] Block unapproved models                                   |  |
|  |  [ ] Allow user-selected models                                |  |
|  |  Max tokens per request: [4096]                                |  |
|  |  Max context length:     [8192]                                |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Save AI Policy]                                                    |
+----------------------------------------------------------------------+
```

### AI Policy Component

```tsx
// components/settings/sections/AIPolicySection.tsx
import { useSettings } from '@/hooks/settings/useSettings';

export function AIPolicySection() {
  const { aiPolicy, updateAIPolicy } = useSettings();
  const [policy, setPolicy] = useState(aiPolicy || {
    guardrails: { harmful: true, pii: true, injection: true, profanity: true, watermark: false },
    toolPolicies: [],
    modelRestrictions: { enforce_routing: true, block_unapproved: true, allow_user_select: false },
    max_tokens: 4096,
    max_context: 8192,
  });

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">AI Policy Settings</h2>

      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Safety Guardrails</h3>
        {Object.entries(GUARDRAIL_LABELS).map(([key, label]) => (
          <div key={key} className="flex items-center justify-between">
            <Label>{label}</Label>
            <Switch
              checked={policy.guardrails[key as keyof typeof policy.guardrails]}
              onCheckedChange={(c) => setPolicy({
                ...policy,
                guardrails: { ...policy.guardrails, [key]: c },
              })}
            />
          </div>
        ))}
      </div>

      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Tool Policies</h3>
        <div className="overflow-x-auto">
          <table className="w-full border-collapse">
            <thead>
              <tr className="border-b">
                <th className="p-2 text-left">Tool</th>
                <th className="p-2">Allowed</th>
                <th className="p-2">Rate Limit</th>
                <th className="p-2">Approval Required</th>
              </tr>
            </thead>
            <tbody>
              {policy.toolPolicies.map((tp, idx) => (
                <tr key={tp.tool} className="border-b">
                  <td className="p-2 font-medium">{tp.tool}</td>
                  <td className="p-2 text-center">
                    <Switch checked={tp.allowed}
                      onCheckedChange={(c) => {
                        const updated = [...policy.toolPolicies];
                        updated[idx] = { ...tp, allowed: c };
                        setPolicy({ ...policy, toolPolicies: updated });
                      }} />
                  </td>
                  <td className="p-2 text-center">
                    <Input type="number" value={tp.rate_limit}
                      onChange={(e) => {
                        const updated = [...policy.toolPolicies];
                        updated[idx] = { ...tp, rate_limit: +e.target.value };
                        setPolicy({ ...policy, toolPolicies: updated });
                      }}
                      className="w-24" />
                  </td>
                  <td className="p-2 text-center">
                    <Switch checked={tp.approval_required}
                      onCheckedChange={(c) => {
                        const updated = [...policy.toolPolicies];
                        updated[idx] = { ...tp, approval_required: c };
                        setPolicy({ ...policy, toolPolicies: updated });
                      }} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Model Restrictions</h3>
        <div className="flex items-center justify-between">
          <Label>Enforce model routing rules</Label>
          <Switch checked={policy.modelRestrictions.enforce_routing}
            onCheckedChange={(c) => setPolicy({
              ...policy,
              modelRestrictions: { ...policy.modelRestrictions, enforce_routing: c },
            })} />
        </div>
        <div className="flex items-center justify-between">
          <Label>Block unapproved models</Label>
          <Switch checked={policy.modelRestrictions.block_unapproved}
            onCheckedChange={(c) => setPolicy({
              ...policy,
              modelRestrictions: { ...policy.modelRestrictions, block_unapproved: c },
            })} />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <Label>Max tokens per request</Label>
            <Input type="number" value={policy.max_tokens}
              onChange={(e) => setPolicy({ ...policy, max_tokens: +e.target.value })} />
          </div>
          <div>
            <Label>Max context length</Label>
            <Input type="number" value={policy.max_context}
              onChange={(e) => setPolicy({ ...policy, max_context: +e.target.value })} />
          </div>
        </div>
      </div>

      <Button onClick={() => updateAIPolicy(policy)}>Save AI Policy</Button>
    </div>
  );
}
```

---

## Tenant Branding Settings

```
+----------------------------------------------------------------------+
|  Branding Settings  [Admin Only]                                     |
+----------------------------------------------------------------------+
|                                                                      |
|  Logo:                                                               |
|  +------+                                                            |
|  | Logo |   [Upload Logo]  [Remove]                                  |
|  +------+   Recommended: 200x50px, SVG or PNG, max 500KB            |
|                                                                      |
|  Colors:                                                             |
|  +----------------------------------------------------------------+  |
|  |  Primary:    [#3B82F6]  [Pick]                                 |  |
|  |  Secondary:  [#8B5CF6]  [Pick]                                 |  |
|  |  Accent:     [#10B981]  [Pick]                                 |  |
|  |  Background: [#FFFFFF]  [Pick]                                 |  |
|  |  Text:       [#1F2937]  [Pick]                                 |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Fonts:                                                              |
|  +----------------------------------------------------------------+  |
|  |  Heading Font:  [Inter v]                                      |  |
|  |  Body Font:     [Inter v]                                      |  |
|  |  Mono Font:     [JetBrains Mono v]                             |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Custom CSS:                                                         |
|  +----------------------------------------------------------------+  |
|  |  /* Add custom CSS overrides */                                |  |
|  |  :root {                                                       |  |
|  |    --primary: #3B82F6;                                          |  |
|  |  }                                                             |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Preview:                                                            |
|  +----------------------------------------------------------------+  |
|  |  [Live preview of branding applied to a sample page]           |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Save Branding]  [Reset to Defaults]                                |
+----------------------------------------------------------------------+
```

---

## API Key Management

```
+----------------------------------------------------------------------+
|  API Keys                                                            |
+----------------------------------------------------------------------+
|                                                                      |
|  Your API Keys:                                                      |
|  +----------------------------------------------------------------+  |
|  |  Name            | Key               | Created  | Last Used | A |  |
|  |------------------+-------------------+----------+-----------+---|  |
|  |  Production      | nx_live_***abc123 | Jul 1    | 2h ago    | V |  |
|  |  Development     | nx_test_***def456 | Jun 15   | 5m ago    | V |  |
|  |  CI/CD Pipeline  | nx_live_***ghi789 | May 20   | 1d ago    | V |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Usage This Month:                                                   |
|  +----------------------------------------------------------------+  |
|  |  Production:    45,230 calls  ($4.52)                          |  |
|  |  Development:   12,450 calls  ($1.25)                          |  |
|  |  CI/CD:          2,100 calls  ($0.21)                          |  |
|  |  Total:         59,780 calls  ($5.98)                          |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Rate Limit: 1000 requests/minute (Pro plan)                         |
|  Remaining:  940/min                                                  |
|                                                                      |
|  [+ Generate New Key]                                                |
+----------------------------------------------------------------------+
```

### API Key Component

```tsx
// components/settings/sections/APIKeysSection.tsx
import { useState } from 'react';
import { useSettings } from '@/hooks/settings/useSettings';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { AlertDialog, AlertDialogAction, AlertDialogCancel,
         AlertDialogContent, AlertDialogDescription,
         AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
         AlertDialogTrigger } from '@/components/ui/alert-dialog';
import { Eye, EyeOff, Trash2 } from 'lucide-react';

export function APIKeysSection() {
  const { apiKeys, generateKey, revokeKey } = useSettings();
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set());
  const [newKeyName, setNewKeyName] = useState('');
  const [showGenerate, setShowGenerate] = useState(false);

  const toggleKeyVisibility = (keyId: string) => {
    setVisibleKeys((prev) => {
      const next = new Set(prev);
      if (next.has(keyId)) next.delete(keyId);
      else next.add(keyId);
      return next;
    });
  };

  const maskKey = (key: string) => {
    const prefix = key.substring(0, 8);
    return `${prefix}${'*'.repeat(24)}`;
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">API Keys</h2>

      <div className="border rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Key</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Last Used</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {apiKeys?.map((apiKey) => (
              <TableRow key={apiKey.id}>
                <TableCell className="font-medium">{apiKey.name}</TableCell>
                <TableCell>
                  <div className="flex items-center gap-2 font-mono text-sm">
                    {visibleKeys.has(apiKey.id) ? apiKey.key : maskKey(apiKey.key)}
                    <Button variant="ghost" size="sm" onClick={() => toggleKeyVisibility(apiKey.id)}>
                      {visibleKeys.has(apiKey.id) ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                    </Button>
                  </div>
                </TableCell>
                <TableCell>{formatDate(apiKey.created_at)}</TableCell>
                <TableCell>{apiKey.last_used ? formatTimeAgo(apiKey.last_used) : 'Never'}</TableCell>
                <TableCell>
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <Button variant="ghost" size="sm">
                        <Trash2 className="w-3 h-3 text-destructive" />
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>Revoke API Key</AlertDialogTitle>
                        <AlertDialogDescription>
                          Are you sure you want to revoke <strong>{apiKey.name}</strong>?
                          Any applications using this key will stop working immediately.
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction onClick={() => revokeKey(apiKey.id)}>
                          Revoke Key
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {showGenerate ? (
        <div className="flex gap-2">
          <Input value={newKeyName} onChange={(e) => setNewKeyName(e.target.value)}
            placeholder="Key name (e.g., 'Production')" className="max-w-xs" />
          <Button onClick={() => { generateKey(newKeyName); setShowGenerate(false); setNewKeyName(''); }}>
            Generate
          </Button>
          <Button variant="outline" onClick={() => setShowGenerate(false)}>Cancel</Button>
        </div>
      ) : (
        <Button onClick={() => setShowGenerate(true)}>+ Generate New Key</Button>
      )}
    </div>
  );
}
```

---

## Webhook Configuration

```
+----------------------------------------------------------------------+
|  Webhooks                                                            |
+----------------------------------------------------------------------+
|                                                                      |
|  Active Webhooks:                                                    |
|  +----------------------------------------------------------------+  |
|  |  URL: https://api.acme.com/webhooks/nexus                     |  |
|  |  Events: workflow.completed, alert.triggered                   |  |
|  |  Status: Active  |  Last: 2h ago (200 OK)                      |  |
|  |  [Edit] [Test] [Disable] [Delete]                              |  |
|  +----------------------------------------------------------------+  |
|  |  URL: https://hooks.slack.com/services/T00...                  |  |
|  |  Events: alert.triggered                                       |  |
|  |  Status: Active  |  Last: 30m ago (200 OK)                     |  |
|  |  [Edit] [Test] [Disable] [Delete]                              |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [+ Add Webhook]                                                     |
|                                                                      |
|  Available Events:                                                   |
|  +----------------------------------------------------------------+  |
|  |  Event                    | Description                         |  |
|  |---------------------------+-------------------------------------|  |
|  |  workflow.completed       | Workflow execution finished          |  |
|  |  workflow.failed          | Workflow execution failed            |  |
|  |  alert.triggered          | Security alert triggered             |  |
|  |  document.processed       | Document processing complete         |  |
|  |  model.status_changed     | Model status changed                 |  |
|  |  user.created             | New user registered                  |  |
|  |  billing.invoice_created  | New invoice generated                |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

---

## Developer Settings

```
+----------------------------------------------------------------------+
|  Developer Settings                                                  |
+----------------------------------------------------------------------+
|                                                                      |
|  Sandbox Mode:                                                       |
|  [x] Enable sandbox (test API calls without real execution)         |
|                                                                      |
|  SDK Downloads:                                                      |
|  +----------------------------------------------------------------+  |
|  |  Python SDK    v1.2.3    [Download]  [Documentation]           |  |
|  |  JavaScript    v1.1.0    [Download]  [Documentation]           |  |
|  |  Go SDK        v0.9.5    [Download]  [Documentation]           |  |
|  |  REST API      v2.0      [Download]  [Documentation]           |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  API Documentation:                                                  |
|  [OpenAPI Spec]  [Postman Collection]  [Interactive Docs]           |
|                                                                      |
|  Sandbox Environment:                                                |
|  +----------------------------------------------------------------+  |
|  |  Base URL:  https://sandbox-api.nexus.ai/v2                    |  |
|  |  API Key:   nx_sandbox_***xyz789 [Regenerate]                  |  |
|  |  Rate Limit: 100 requests/minute                               |  |
|  |  Data Retention: 7 days                                        |  |
|  +----------------------------------------------------------------+  |
+----------------------------------------------------------------------+
```

---

## Security Settings

```
+----------------------------------------------------------------------+
|  Security Settings                                                   |
+----------------------------------------------------------------------+
|                                                                      |
|  Session Management:                                                 |
|  +----------------------------------------------------------------+  |
|  |  Session Timeout:         [30] minutes                          |  |
|  |  Idle Timeout:            [15] minutes                          |  |
|  |  Max Concurrent Sessions: [5]                                   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Two-Factor Authentication:                                          |
|  Status: Enabled (Authenticator App)  [Manage] [Disable]            |
|                                                                      |
|  IP Whitelist:                                                       |
|  +----------------------------------------------------------------+  |
|  |  192.168.1.0/24    Office Network    [Remove]                  |  |
|  |  10.0.0.0/8        VPN Range         [Remove]                  |  |
|  +----------------------------------------------------------------+  |
|  |  [Add IP Range]                                                 |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Login History (Last 10):                                            |
|  +----------------------------------------------------------------+  |
|  |  Time          | IP Address     | Device        | Status       |  |
|  |----------------+----------------+---------------+--------------|  |
|  |  Jul 16 10:28  | 192.168.1.100  | Chrome/Mac    | Success      |  |
|  |  Jul 15 16:30  | 192.168.1.100  | Chrome/Mac    | Success      |  |
|  |  Jul 15 09:00  | 10.0.0.50      | Safari/iOS    | Success      |  |
|  |  Jul 14 22:15  | 203.0.113.42   | Unknown       | Failed       |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [Save Security Settings]                                            |
+----------------------------------------------------------------------+
```

---

## Data Retention Settings

```
+----------------------------------------------------------------------+
|  Data Retention  [Admin Only]                                        |
+----------------------------------------------------------------------+
|                                                                      |
|  Retention Policies:                                                 |
|  +----------------------------------------------------------------+  |
|  |  Chat History:           [90] days                             |  |
|  |  Audit Logs:             [365] days                            |  |
|  |  Documents:              [Unlimited]                           |  |
|  |  Model Logs:             [30] days                             |  |
|  |  Deleted Items (Trash):  [30] days                             |  |
|  |  API Usage Data:         [365] days                            |  |
|  |  Session Data:           [30] days                             |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Data Export:                                                        |
|  [Export All Data (JSON)]  [Export Audit Logs (CSV)]                |
|  [Export User Data (GDPR)]                                          |
|                                                                      |
|  Data Deletion:                                                      |
|  [Delete Expired Data Now]  [Schedule Cleanup]                       |
|                                                                      |
|  Warning: Data deletion is permanent and cannot be undone.           |
|                                                                      |
|  [Save Retention Settings]                                           |
+----------------------------------------------------------------------+
```

---

## Integration Settings

```
+----------------------------------------------------------------------+
|  Integrations                                                        |
+----------------------------------------------------------------------+
|                                                                      |
|  Connected Services:                                                 |
|  +----------------------------------------------------------------+  |
|  |  Google Workspace   | Connected (john@acme.com) | [Disconnect] |  |
|  |  Microsoft 365      | Connected (john@acme.com) | [Disconnect] |  |
|  |  Slack              | Connected (#nexus-alerts)  | [Disconnect] |  |
|  |  GitHub             | Connected (john-acme)      | [Disconnect] |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  Available Integrations:                                             |
|  +----------------------------------------------------------------+  |
|  |  Notion            | [Connect]                                 |  |
|  |  Confluence        | [Connect]                                 |  |
|  |  Salesforce        | [Connect]                                 |  |
|  |  HubSpot           | [Connect]                                 |  |
|  |  Jira              | [Connect]                                 |  |
|  |  GitLab            | [Connect]                                 |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  OAuth Apps:                                                         |
|  +----------------------------------------------------------------+  |
|  |  Name              | Scopes            | Status   | Actions    |  |
|  |--------------------+-------------------+----------+------------|  |
|  |  Custom App 1      | read, write       | Active   | [Revoke]   |  |
|  |  CI/CD Integration | read              | Active   | [Revoke]   |  |
|  +----------------------------------------------------------------+  |
|                                                                      |
|  [+ Register New OAuth App]                                          |
+----------------------------------------------------------------------+
```

---

## API Integration

### Endpoints

| Method   | Endpoint                        | Description                  |
|----------|---------------------------------|------------------------------|
| `GET`    | `/api/settings/profile`         | Get user profile             |
| `PATCH`  | `/api/settings/profile`         | Update user profile          |
| `POST`   | `/api/settings/avatar`          | Upload avatar                |
| `DELETE` | `/api/settings/avatar`          | Remove avatar                |
| `POST`   | `/api/settings/password`        | Change password              |
| `GET`    | `/api/settings/notifications`   | Get notification preferences |
| `PATCH`  | `/api/settings/notifications`   | Update notification prefs    |
| `GET`    | `/api/settings/theme`           | Get theme preferences        |
| `PATCH`  | `/api/settings/theme`           | Update theme preferences     |
| `GET`    | `/api/settings/language`        | Get language preferences     |
| `PATCH`  | `/api/settings/language`        | Update language preferences  |
| `GET`    | `/api/settings/api-keys`        | List API keys                |
| `POST`   | `/api/settings/api-keys`        | Generate API key             |
| `DELETE` | `/api/settings/api-keys/:id`    | Revoke API key               |
| `GET`    | `/api/settings/webhooks`        | List webhooks                |
| `POST`   | `/api/settings/webhooks`        | Create webhook               |
| `PATCH`  | `/api/settings/webhooks/:id`    | Update webhook               |
| `DELETE` | `/api/settings/webhooks/:id`    | Delete webhook               |
| `POST`   | `/api/settings/webhooks/:id/test` | Test webhook               |
| `GET`    | `/api/settings/security`        | Get security settings        |
| `PATCH`  | `/api/settings/security`        | Update security settings     |
| `GET`    | `/api/settings/sessions`        | List active sessions         |
| `DELETE` | `/api/settings/sessions/:id`    | Revoke session               |
| `GET`    | `/api/settings/retention`       | Get retention settings       |
| `PATCH`  | `/api/settings/retention`       | Update retention settings    |
| `GET`    | `/api/settings/integrations`    | List integrations            |
| `POST`   | `/api/settings/integrations/connect`    | Connect service    |
| `DELETE` | `/api/settings/integrations/:id`        | Disconnect service  |
| `GET`    | `/api/admin/tenant/settings`    | Get tenant settings (admin)  |
| `PATCH`  | `/api/admin/tenant/settings`    | Update tenant settings       |
| `GET`    | `/api/admin/tenant/ai-policy`   | Get AI policy (admin)        |
| `PATCH`  | `/api/admin/tenant/ai-policy`   | Update AI policy             |
| `GET`    | `/api/admin/tenant/branding`    | Get branding settings        |
| `PATCH`  | `/api/admin/tenant/branding`    | Update branding              |

---

## Hooks

### useSettings

```typescript
// hooks/settings/useSettings.ts
export function useSettings() {
  const queryClient = useQueryClient();

  const profile = useQuery({
    queryKey: ['settings', 'profile'],
    queryFn: () => api.get('/settings/profile'),
  });

  const updateProfile = useMutation({
    mutationFn: (data: Partial<Profile>) => api.patch('/settings/profile', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['settings', 'profile'] }),
  });

  const changePassword = useMutation({
    mutationFn: (data: ChangePasswordRequest) => api.post('/settings/password', data),
  });

  return {
    profile: profile.data,
    updateProfile: updateProfile.mutateAsync,
    changePassword: changePassword.mutateAsync,
    isLoading: profile.isLoading,
  };
}
```

### useProfile

```typescript
// hooks/settings/useProfile.ts
export function useProfile() {
  const queryClient = useQueryClient();

  const profile = useQuery({
    queryKey: ['profile'],
    queryFn: () => api.get('/settings/profile'),
  });

  const updateProfile = useMutation({
    mutationFn: (data: Partial<Profile>) => api.patch('/settings/profile', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['profile'] }),
  });

  const uploadAvatar = useMutation({
    mutationFn: (file: File) => {
      const formData = new FormData();
      formData.append('avatar', file);
      return api.post('/settings/avatar', formData);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['profile'] }),
  });

  const removeAvatar = useMutation({
    mutationFn: () => api.delete('/settings/avatar'),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['profile'] }),
  });

  return {
    profile: profile.data,
    updateProfile: updateProfile.mutateAsync,
    uploadAvatar: uploadAvatar.mutateAsync,
    removeAvatar: removeAvatar.mutateAsync,
  };
}
```

### usePreferences

```typescript
// hooks/settings/usePreferences.ts
export function usePreferences() {
  const preferences = useQuery({
    queryKey: ['preferences'],
    queryFn: () => api.get('/settings/preferences'),
  });

  const updatePreferences = useMutation({
    mutationFn: (data: Partial<UserPreferences>) =>
      api.patch('/settings/preferences', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['preferences'] }),
  });

  return {
    preferences: preferences.data,
    updatePreferences: updatePreferences.mutateAsync,
  };
}
```

---

## Stores

### Settings Store (Zustand)

```typescript
// stores/settings/settings.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface SettingsState {
  theme: 'light' | 'dark' | 'system';
  accentColor: string;
  fontSize: 'small' | 'medium' | 'large';
  compactMode: boolean;
  sidebarPosition: 'left' | 'right';
  language: string;
  timezone: string;
  dateFormat: string;

  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  setAccentColor: (color: string) => void;
  setFontSize: (size: 'small' | 'medium' | 'large') => void;
  setCompactMode: (compact: boolean) => void;
  setSidebarPosition: (pos: 'left' | 'right') => void;
  setLanguage: (language: string) => void;
  setTimezone: (timezone: string) => void;
  setDateFormat: (format: string) => void;
  resetDefaults: () => void;
}

const DEFAULTS = {
  theme: 'system' as const,
  accentColor: '#3B82F6',
  fontSize: 'medium' as const,
  compactMode: false,
  sidebarPosition: 'left' as const,
  language: 'en',
  timezone: 'UTC',
  dateFormat: 'YYYY-MM-DD',
};

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      ...DEFAULTS,

      setTheme: (theme) => set({ theme }),
      setAccentColor: (accentColor) => set({ accentColor }),
      setFontSize: (fontSize) => set({ fontSize }),
      setCompactMode: (compactMode) => set({ compactMode }),
      setSidebarPosition: (sidebarPosition) => set({ sidebarPosition }),
      setLanguage: (language) => set({ language }),
      setTimezone: (timezone) => set({ timezone }),
      setDateFormat: (dateFormat) => set({ dateFormat }),
      resetDefaults: () => set(DEFAULTS),
    }),
    {
      name: 'nexus-settings',
    }
  )
);
```

---

## Error Handling

### Error Types

| Error                    | Message                          | Action                        |
|--------------------------|----------------------------------|-------------------------------|
| Profile save failed      | "Could not save profile"         | Show validation errors        |
| Password change failed   | "Current password incorrect"     | Highlight password field      |
| Avatar upload failed     | "Avatar upload failed"           | Retry button                  |
| Avatar too large         | "File exceeds 2MB limit"         | Show file size info           |
| API key generation fail  | "Could not generate key"         | Retry button                  |
| Webhook test failed      | "Webhook endpoint not reachable" | Show HTTP status/error        |
| Settings save failed     | "Settings not saved"             | Show what changed             |
| Theme apply failed       | "Could not apply theme"          | Revert to previous theme      |
| Integration connect fail | "Authorization failed"           | Re-authenticate               |
| Session revoke failed    | "Could not revoke session"       | Retry button                  |
| Invalid API key name     | "Name already exists"            | Show conflict error           |
| Retention policy invalid | "Minimum 7 days"                 | Show validation error         |

### Save Feedback

```tsx
// components/settings/SaveFeedback.tsx
import { toast } from 'sonner';

export function useSaveFeedback() {
  const saveSuccess = (section?: string) => {
    toast({
      title: 'Settings saved',
      description: section ? `${section} settings updated successfully` : undefined,
    });
  };

  const saveError = (error: any, section?: string) => {
    const message = error?.response?.data?.message || error?.message || 'Unknown error';
    toast({
      title: `Failed to save${section ? ` ${section}` : ''} settings`,
      description: message,
      variant: 'destructive',
    });
  };

  return { saveSuccess, saveError };
}
```

---

## Responsive Design

### Mobile Settings (<=640px)

```
+----------------------+
| Settings       [X]   |
+----------------------+
| Tab: [Profile] [v]   |
+----------------------+
|                      |
| [Avatar]             |
|                      |
| First Name:          |
| [John           ]    |
|                      |
| Last Name:           |
| [Smith          ]    |
|                      |
| Email:               |
| [john@acme.com  ]    |
|                      |
| Bio:                 |
| [Senior engine..]    |
|                      |
| [Save Profile]       |
+----------------------+
```

### Responsive Breakpoints

| Breakpoint | Layout                                               |
|------------|------------------------------------------------------|
| <=640px    | Dropdown nav, full-width sections, stacked forms     |
| 641-1024px | Collapsible sidebar, 2-column forms                  |
| >1024px    | Fixed sidebar, 3-column max-width layout             |

### Settings Navigation Patterns

| Viewport    | Navigation                           |
|-------------|--------------------------------------|
| Mobile      | Dropdown select for sections         |
| Tablet      | Horizontal tab bar at top            |
| Desktop     | Left sidebar with section groups     |
| Wide desktop| Left sidebar + section anchor list   |
