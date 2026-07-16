# AeroXe Nexus AI — Design System

## Design Tokens, shadcn/ui Components, Theming, Typography, Colors, Animations

---

## 1. Design System Overview

```
┌──────────────────────────────────────────────────────────┐
│                  AeroXe Nexus AI Design System             │
│                                                            │
│  Layer 1: Design Tokens (CSS Variables)                    │
│  ├── Colors, Typography, Spacing, Shadows, Borders         │
│  ├── Light Mode + Dark Mode                                │
│  └── Enterprise Branding Overrides                         │
│                                                            │
│  Layer 2: UI Primitives (shadcn/ui)                        │
│  ├── 40+ accessible components                             │
│  ├── Radix UI primitives (headless)                        │
│  └── Tailwind CSS styling                                  │
│                                                            │
│  Layer 3: Composite Components                             │
│  ├── DataTable, ChatWindow, AgentCard, KPICard             │
│  └── Domain-specific patterns                              │
│                                                            │
│  Layer 4: Layout & Pages                                   │
│  ├── App Shell, Sidebar, Header                            │
│  └── Full page compositions                                │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

---

## 2. Design Tokens

### 2.1 Color System — CSS Variables

```css
/* src/app/globals.css */

@layer base {
  :root {
    /* Primary — Navy Blue (Brand) */
    --primary: 222 47% 11%;          /* #0f1d3d */
    --primary-foreground: 0 0% 98%;  /* #fafafa */

    /* Secondary — Slate Blue */
    --secondary: 217 33% 17%;
    --secondary-foreground: 0 0% 98%;

    /* Accent — Electric Blue */
    --accent: 217 91% 60%;           /* #2563eb */
    --accent-foreground: 0 0% 100%;

    /* Background */
    --background: 0 0% 100%;         /* #ffffff */
    --foreground: 222 47% 11%;       /* #0f1d3d */

    /* Card */
    --card: 0 0% 100%;
    --card-foreground: 222 47% 11%;

    /* Popover */
    --popover: 0 0% 100%;
    --popover-foreground: 222 47% 11%;

    /* Muted */
    --muted: 210 40% 96%;
    --muted-foreground: 215 16% 47%;

    /* Border / Input / Ring */
    --border: 214 32% 91%;
    --input: 214 32% 91%;
    --ring: 217 91% 60%;

    /* Destructive */
    --destructive: 0 84% 60%;
    --destructive-foreground: 0 0% 98%;

    /* Success */
    --success: 142 76% 36%;
    --success-foreground: 0 0% 98%;

    /* Warning */
    --warning: 38 92% 50%;
    --warning-foreground: 0 0% 100%;

    /* Radius */
    --radius: 0.5rem;

    /* Shadows */
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
    --shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.1);
  }

  .dark {
    --primary: 217 33% 17%;
    --primary-foreground: 0 0% 98%;

    --secondary: 217 20% 25%;
    --secondary-foreground: 0 0% 98%;

    --accent: 217 91% 60%;
    --accent-foreground: 0 0% 100%;

    --background: 222 47% 6%;
    --foreground: 210 40% 96%;

    --card: 222 47% 9%;
    --card-foreground: 210 40% 96%;

    --popover: 222 47% 9%;
    --popover-foreground: 210 40% 96%;

    --muted: 217 20% 15%;
    --muted-foreground: 215 20% 55%;

    --border: 217 20% 15%;
    --input: 217 20% 15%;
    --ring: 217 91% 60%;

    --destructive: 0 62% 50%;
    --destructive-foreground: 0 0% 98%;

    --success: 142 76% 36%;
    --warning: 38 92% 50%;
  }
}
```

### 2.2 Color Reference Table

| Token | Light Mode | Dark Mode | Usage |
|---|---|---|---|
| `--primary` | `#0f1d3d` | `#1a2744` | Buttons, headers, primary actions |
| `--primary-foreground` | `#fafafa` | `#fafafa` | Text on primary backgrounds |
| `--accent` | `#2563eb` | `#2563eb` | Links, highlights, active states |
| `--background` | `#ffffff` | `#0c1425` | Page background |
| `--foreground` | `#0f1d3d` | `#e2e8f0` | Default text color |
| `--muted` | `#f1f5f9` | `#1a2438` | Subtle backgrounds, placeholders |
| `--muted-foreground` | `#64748b` | `#7c8db5` | Secondary text, labels |
| `--border` | `#e2e8f0` | `#1a2438` | Borders, dividers |
| `--destructive` | `#dc2626` | `#ef4444` | Error states, delete actions |
| `--success` | `#16a34a` | `#22c55e` | Success states, completions |
| `--warning` | `#f59e0b` | `#fbbf24` | Warning states, cautions |

### 2.3 Semantic Color Usage

| Context | Color | CSS Variable |
|---|---|---|
| Online status | Green | `var(--success)` |
| Offline status | Gray | `var(--muted-foreground)` |
| Error status | Red | `var(--destructive)` |
| Warning status | Amber | `var(--warning)` |
| Processing status | Blue | `var(--accent)` |
| Link hover | Blue | `var(--accent)` |
| Selected item | Blue tint | `var(--accent) / 10%` |
| AI thinking | Purple | `#8b5cf6` |
| AI streaming | Blue pulse | `var(--accent)` |

---

## 3. Typography System

### 3.1 Font Families

```css
@layer base {
  :root {
    --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    --font-mono: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
    --font-display: 'Inter', var(--font-sans);
  }
}
```

### 3.2 Font Sizes Scale

| Token | Size | Line Height | Usage |
|---|---|---|---|
| `text-xs` | 0.75rem (12px) | 1rem | Captions, labels |
| `text-sm` | 0.875rem (14px) | 1.25rem | Secondary text, metadata |
| `text-base` | 1rem (16px) | 1.5rem | Body text, default |
| `text-lg` | 1.125rem (18px) | 1.75rem | Subheadings |
| `text-xl` | 1.25rem (20px) | 1.75rem | Section titles |
| `text-2xl` | 1.5rem (24px) | 2rem | Page titles |
| `text-3xl` | 1.875rem (30px) | 2.25rem | Hero headings |
| `text-4xl` | 2.25rem (36px) | 2.5rem | Display headings |

### 3.3 Font Weights

| Token | Weight | Usage |
|---|---|---|
| `font-normal` | 400 | Body text |
| `font-medium` | 500 | Labels, emphasis |
| `font-semibold` | 600 | Headings, buttons |
| `font-bold` | 700 | Strong emphasis, page titles |

### 3.4 Typography Components

```tsx
// Heading hierarchy
<h1 className="text-4xl font-bold tracking-tight">Page Title</h1>
<h2 className="text-2xl font-semibold tracking-tight">Section Title</h2>
<h3 className="text-xl font-semibold">Card Title</h3>
<h4 className="text-lg font-medium">Subsection</h4>

// Body text
<p className="text-base leading-7">Body paragraph text</p>
<p className="text-sm text-muted-foreground">Secondary or meta text</p>

// Code
<code className="text-sm font-mono bg-muted px-1.5 py-0.5 rounded">const x = 1</code>

// Labels
<label className="text-sm font-medium leading-none">Form Label</label>
```

---

## 4. Spacing System

### 4.1 Spacing Scale

| Token | Size | Pixels | Usage |
|---|---|---|---|
| `0` | 0px | 0 | Reset |
| `px` | 1px | 1 | Hairline borders |
| `0.5` | 0.125rem | 2px | Tight spacing |
| `1` | 0.25rem | 4px | Compact elements |
| `1.5` | 0.375rem | 6px | Small gaps |
| `2` | 0.5rem | 8px | Icon gaps, inline spacing |
| `3` | 0.75rem | 12px | Card padding, form gaps |
| `4` | 1rem | 16px | Standard spacing |
| `5` | 1.25rem | 20px | Section spacing |
| `6` | 1.5rem | 24px | Card internal padding |
| `8` | 2rem | 32px | Page section spacing |
| `10` | 2.5rem | 40px | Large section spacing |
| `12` | 3rem | 48px | Page margins |
| `16` | 4rem | 64px | Hero spacing |
| `20` | 5rem | 80px | Extra large spacing |

### 4.2 Spacing Usage Patterns

| Pattern | Classes | Example |
|---|---|---|
| Inline elements | `gap-1` to `gap-2` | Icon + text in button |
| Card internal | `p-4` to `p-6` | Card padding |
| Between cards | `gap-4` to `gap-6` | Grid gap |
| Page sections | `space-y-6` to `space-y-8` | Vertical section spacing |
| Form fields | `space-y-4` | Between form groups |
| Modal padding | `p-6` | Dialog internal padding |
| Sidebar items | `px-3 py-2` | Nav item padding |
| Table cells | `px-4 py-3` | Cell padding |

---

## 5. Shadow System

| Token | Value | Usage |
|---|---|---|
| `shadow-sm` | `0 1px 2px 0 rgb(0 0 0 / 0.05)` | Subtle elevation (badges, tags) |
| `shadow` | `0 1px 3px 0 rgb(0 0 0 / 0.1)` | Default cards |
| `shadow-md` | `0 4px 6px -1px rgb(0 0 0 / 0.1)` | Dropdowns, popovers |
| `shadow-lg` | `0 10px 15px -3px rgb(0 0 0 / 0.1)` | Modals, dialogs |
| `shadow-xl` | `0 20px 25px -5px rgb(0 0 0 / 0.1)` | Floating panels |
| `ring` | `0 0 0 2px var(--ring)` | Focus rings |

---

## 6. Border System

| Token | Value | Usage |
|---|---|---|
| `border` | `1px solid var(--border)` | Default borders |
| `border-2` | `2px solid var(--border)` | Emphasized borders |
| `border-dashed` | `1px dashed var(--border)` | Drop zones, optional areas |
| `rounded-sm` | `0.125rem` | Small radius (badges) |
| `rounded` | `0.25rem` | Default radius |
| `rounded-md` | `0.375rem` | Medium radius (buttons) |
| `rounded-lg` | `0.5rem` | Large radius (cards) |
| `rounded-xl` | `0.75rem` | Extra large (modals) |
| `rounded-full` | `9999px` | Circular (avatars, pills) |

---

## 7. shadcn/ui Component Library

### 7.1 Component Catalogue

| Component | File | Description | Key Props |
|---|---|---|---|
| Button | `button.tsx` | Interactive button | `variant`, `size`, `disabled`, `asChild` |
| Card | `card.tsx` | Content container | `CardHeader`, `CardContent`, `CardFooter` |
| Dialog | `dialog.tsx` | Modal dialog | `open`, `onOpenChange` |
| Sheet | `sheet.tsx` | Slide-out panel | `side`, `open`, `onOpenChange` |
| Form | `form.tsx` | React Hook Form wrapper | `control`, `schema` |
| Input | `input.tsx` | Text input | `type`, `placeholder`, `disabled` |
| Textarea | `textarea.tsx` | Multi-line text | `rows`, `disabled` |
| Select | `select.tsx` | Dropdown select | `value`, `onValueChange` |
| Table | `table.tsx` | Data table | `TableHeader`, `TableBody`, `TableRow` |
| Tabs | `tabs.tsx` | Tab navigation | `value`, `onValueChange` |
| Toast | `toast.tsx` | Notification popup | `variant`, `title`, `description` |
| Command | `command.tsx` | Command palette | `value`, `onValueChange` |
| Badge | `badge.tsx` | Status label | `variant` |
| Alert | `alert.tsx` | Alert banner | `variant` |
| AlertDialog | `alert-dialog.tsx` | Confirmation modal | `onConfirm`, `onCancel` |
| Accordion | `accordion.tsx` | Collapsible content | `type`, `collapsible` |
| Avatar | `avatar.tsx` | User image | `src`, `fallback` |
| Checkbox | `checkbox.tsx` | Boolean toggle | `checked`, `onCheckedChange` |
| Switch | `switch.tsx` | Toggle switch | `checked`, `onCheckedChange` |
| RadioGroup | `radio-group.tsx` | Radio selection | `value`, `onValueChange` |
| Label | `label.tsx` | Form label | `htmlFor` |
| Separator | `separator.tsx` | Visual divider | `orientation` |
| ScrollArea | `scroll-area.tsx` | Custom scrollbar | `orientation` |
| Skeleton | `skeleton.tsx` | Loading placeholder | `className` |
| Tooltip | `tooltip.tsx` | Hover tooltip | `content`, `side` |
| Popover | `popover.tsx` | Click popover | `open`, `onOpenChange` |
| Progress | `progress.tsx` | Progress bar | `value`, `max` |
| Chart | `chart.tsx` | Recharts wrapper | `data`, `config` |

### 7.2 Button Component

```tsx
// Variants
<Button variant="default">Primary Action</Button>
<Button variant="secondary">Secondary</Button>
<Button variant="destructive">Delete</Button>
<Button variant="outline">Cancel</Button>
<Button variant="ghost">Ghost</Button>
<Button variant="link">Link Style</Button>

// Sizes
<Button size="default">Normal</Button>
<Button size="sm">Small</Button>
<Button size="lg">Large</Button>
<Button size="icon"><Icon /></Button>

// With loading state
<Button disabled={isLoading}>
  {isLoading && <Loader2 className="animate-spin mr-2 h-4 w-4" />}
  Save Changes
</Button>
```

### 7.3 Card Component

```tsx
<Card>
  <CardHeader>
    <CardTitle>Agent: Customer Support</CardTitle>
    <CardDescription>phi4-mini:3.8b model</CardDescription>
  </CardHeader>
  <CardContent>
    <p>Handles customer support queries.</p>
  </CardContent>
  <CardFooter>
    <Button size="sm">Configure</Button>
    <Button size="sm" variant="outline">Execute</Button>
  </CardFooter>
</Card>
```

### 7.4 Table Component

```tsx
<Table>
  <TableHeader>
    <TableRow>
      <TableHead>Name</TableHead>
      <TableHead>Model</TableHead>
      <TableHead>Status</TableHead>
      <TableHead className="text-right">Actions</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    {agents.map((agent) => (
      <TableRow key={agent.id}>
        <TableCell className="font-medium">{agent.name}</TableCell>
        <TableCell>{agent.model}</TableCell>
        <TableCell><StatusBadge status={agent.status} /></TableCell>
        <TableCell className="text-right">
          <Button variant="ghost" size="icon">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </TableCell>
      </TableRow>
    ))}
  </TableBody>
</Table>
```

### 7.5 Toast Notifications

```tsx
import { toast } from "@/components/ui/sonner";

// Success
toast.success("Agent created successfully");

// Error
toast.error("Failed to connect to database", {
  description: "Check your credentials and try again.",
});

// Warning
toast.warning("Rate limit approaching", {
  description: "You've used 90% of your monthly quota.",
});

// Info
toast.info("Document processing started", {
  description: "This may take a few minutes.",
});

// With action
toast("Unsaved changes", {
  action: { label: "Save", onClick: () => save() },
});
```

### 7.6 Command Menu (Cmd+K)

```tsx
<CommandDialog open={open} onOpenChange={setOpen}>
  <CommandInput placeholder="Type a command..." />
  <CommandList>
    <CommandEmpty>No results found.</CommandEmpty>
    <CommandGroup heading="Actions">
      <CommandItem onSelect={() => startChat()}>
        <MessageSquare className="mr-2 h-4 w-4" />
        Start New Chat
      </CommandItem>
      <CommandItem onSelect={() => uploadDoc()}>
        <Upload className="mr-2 h-4 w-4" />
        Upload Document
      </CommandItem>
    </CommandGroup>
    <CommandGroup heading="Navigation">
      <CommandItem onSelect={() => router.push("/dashboard")}>
        <LayoutDashboard className="mr-2 h-4 w-4" />
        Dashboard
      </CommandItem>
    </CommandGroup>
  </CommandList>
</CommandDialog>
```

### 7.7 Form Component (React Hook Form + Zod)

```tsx
const formSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  model: z.string().min(1, "Model is required"),
  systemPrompt: z.string().optional(),
  temperature: z.number().min(0).max(2).default(0.7),
});

<Form {...form}>
  <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
    <FormField
      control={form.control}
      name="name"
      render={({ field }) => (
        <FormItem>
          <FormLabel>Agent Name</FormLabel>
          <FormControl>
            <Input placeholder="Customer Support Agent" {...field} />
          </FormControl>
          <FormDescription>Choose a descriptive name.</FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
    <FormField
      control={form.control}
      name="model"
      render={({ field }) => (
        <FormItem>
          <FormLabel>Model</FormLabel>
          <Select onValueChange={field.onChange} defaultValue={field.value}>
            <FormControl>
              <SelectTrigger>
                <SelectValue placeholder="Select model" />
              </SelectTrigger>
            </FormControl>
            <SelectContent>
              <SelectItem value="phi4-mini:3.8b">Phi-4 Mini (3.8B)</SelectItem>
              <SelectItem value="qwen2.5-coder:3b">Qwen2.5 Coder (3B)</SelectItem>
              <SelectItem value="command-r7b:7b">Command-R (7B)</SelectItem>
            </SelectContent>
          </Select>
          <FormMessage />
        </FormItem>
      )}
    />
    <Button type="submit">Create Agent</Button>
  </form>
</Form>
```

---

## 8. Theme System

### 8.1 Theme Architecture

```
┌─────────────────────────────────────────────────┐
│                  Theme System                     │
│                                                   │
│  ┌──────────────┐  ┌──────────────┐              │
│  │  Light Mode   │  │  Dark Mode   │              │
│  │  (default)    │  │  (.dark)     │              │
│  └──────┬───────┘  └──────┬───────┘              │
│         │                  │                      │
│         └───────┬──────────┘                      │
│                 │                                  │
│         ┌───────▼───────┐                         │
│         │  ThemeProvider  │                         │
│         │  (next-themes)  │                         │
│         └───────┬───────┘                         │
│                 │                                  │
│         ┌───────▼───────┐                         │
│         │  CSS Variables  │                         │
│         │  (globals.css)  │                         │
│         └───────────────┘                         │
│                                                   │
└─────────────────────────────────────────────────┘
```

### 8.2 Theme Toggle Implementation

```tsx
"use client";

import { useTheme } from "next-themes";

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
      aria-label="Toggle theme"
    >
      <Sun className="h-5 w-5 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      <Moon className="absolute h-5 w-5 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
    </Button>
  );
}
```

### 8.3 Enterprise Branding

```typescript
// Tenant-specific theme overrides
interface TenantBranding {
  logo: {
    light: string;   // URL to light mode logo
    dark: string;    // URL to dark mode logo
    favicon: string; // URL to favicon
  };
  colors: {
    primary: string;   // Override primary color
    accent: string;    // Override accent color
  };
  fonts: {
    heading: string;   // Custom heading font
    body: string;      // Custom body font
  };
  name: string;        // Tenant display name
}
```

### 8.4 Theme Persistence

| Storage | Scope | Duration |
|---|---|---|
| `localStorage` (`theme`) | Theme preference | Until cleared |
| `sessionStorage` (`tenant`) | Tenant context | Session only |
| HTTP-only cookie (`session`) | Auth session | 7 days |

---

## 9. Layout System

### 9.1 Grid System

```tsx
// Dashboard grid
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
  <KPICard />
  <KPICard />
  <KPICard />
  <KPICard />
</div>

// Agent cards
<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
  <AgentCard />
  <AgentCard />
  <AgentCard />
</div>

// Chart layout
<div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
  <ChartCard />
  <ChartCard />
</div>
```

### 9.2 Responsive Breakpoints

| Prefix | Width | Columns | Gap |
|---|---|---|---|
| Default (mobile) | < 640px | 1 | 4 |
| `sm:` | ≥ 640px | 2 | 4 |
| `md:` | ≥ 768px | 2-3 | 6 |
| `lg:` | ≥ 1024px | 3-4 | 6 |
| `xl:` | ≥ 1280px | 4-5 | 6 |

### 9.3 App Shell Layout

```
┌─────────────────────────────────────────────────┐
│  Header (fixed top, h-14)                        │
│  ┌─────────┬───────────────────────────────────┐│
│  │ Sidebar │  Main Content Area                 ││
│  │ (w-60)  │  (scrollable, p-6)                ││
│  │         │                                    ││
│  │ - Logo  │  ┌─────────────────────────────┐  ││
│  │ - Nav   │  │  Page Content               │  ││
│  │ - Menu  │  │                             │  ││
│  │         │  └─────────────────────────────┘  ││
│  │         │                                    ││
│  └─────────┴───────────────────────────────────┘│
└─────────────────────────────────────────────────┘
```

---

## 10. Animation System

### 10.1 Framer Motion Presets

```typescript
// Fade in
const fadeIn = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
  transition: { duration: 0.2 },
};

// Slide up
const slideUp = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: 20 },
  transition: { duration: 0.3, ease: "easeOut" },
};

// Scale in (for modals)
const scaleIn = {
  initial: { opacity: 0, scale: 0.95 },
  animate: { opacity: 1, scale: 1 },
  exit: { opacity: 0, scale: 0.95 },
  transition: { duration: 0.2 },
};

// Slide from right (sidebar)
const slideRight = {
  initial: { x: -20, opacity: 0 },
  animate: { x: 0, opacity: 1 },
  transition: { duration: 0.3 },
};
```

### 10.2 CSS Animation Utilities

| Class | Effect | Usage |
|---|---|---|
| `animate-pulse` | Fade in/out pulse | Loading states |
| `animate-spin` | 360° rotation | Loading spinners |
| `animate-bounce` | Bounce up/down | Attention indicators |
| `animate-ping` | Ripple effect | Active status dots |
| `transition-all` | Smooth transitions | Theme toggles |
| `transition-colors` | Color transitions | Hover effects |
| `transition-transform` | Transform transitions | Hover scale |

### 10.3 Streaming Cursor Animation

```css
/* Blinking cursor for AI typing */
@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.streaming-cursor::after {
  content: '▊';
  animation: blink 1s step-end infinite;
  color: var(--accent);
}
```

### 10.4 Chat Message Entry Animation

```tsx
<motion.div
  initial={{ opacity: 0, y: 10 }}
  animate={{ opacity: 1, y: 0 }}
  transition={{ duration: 0.2 }}
  className="message-bubble"
>
  {content}
</motion.div>
```

---

## 11. Icon System

### 11.1 Lucide Icons

```tsx
import {
  MessageSquare, Bot, Settings, User, LogOut,
  ChevronRight, Search, Bell, Moon, Sun,
  Play, Pause, Check, X, AlertTriangle,
  Upload, Download, FileText, Database,
  Shield, Key, Globe, Zap, Cpu,
  BarChart3, TrendingUp, Activity,
} from "lucide-react";

// Size variants
<Icon className="h-4 w-4" />  // Small (buttons, inline)
<Icon className="h-5 w-5" />  // Default (navigation)
<Icon className="h-6 w-6" />  // Large (headers)
<Icon className="h-8 w-8" />  // XL (empty states)

// With color
<Icon className="h-5 w-5 text-green-500" />   // Success
<Icon className="h-5 w-5 text-red-500" />     // Error
<Icon className="h-5 w-5 text-blue-500" />    // Info
<Icon className="h-5 w-5 text-amber-500" />   // Warning
```

### 11.2 Icon Mapping by Context

| Context | Icon | Source |
|---|---|---|
| Chat | `MessageSquare` | Lucide |
| Agent | `Bot` | Lucide |
| Dashboard | `LayoutDashboard` | Lucide |
| Settings | `Settings` | Lucide |
| Documents | `FileText` | Lucide |
| Database | `Database` | Lucide |
| Security | `Shield` | Lucide |
| Analytics | `BarChart3` | Lucide |
| Notifications | `Bell` | Lucide |
| Search | `Search` | Lucide |

---

## 12. Accessibility

### 12.1 Focus Management

| Pattern | Implementation |
|---|---|
| Focus ring | `focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2` |
| Skip to content | `<a href="#main" className="sr-only focus:not-sr-only">Skip to content</a>` |
| Focus trap in modals | Radix Dialog handles focus trapping |
| Return focus on close | Radix manages focus restoration |

### 12.2 Contrast Ratios

| Pair | Ratio | WCAG Level |
|---|---|---|
| Foreground on Background | ≥ 7:1 | AAA |
| Muted text on Background | ≥ 4.5:1 | AA |
| Primary on White | ≥ 7:1 | AAA |
| Accent on White | ≥ 4.5:1 | AA |
| Destructive on White | ≥ 4.5:1 | AA |
| Success on White | ≥ 4.5:1 | AA |

### 12.3 Hit Target Sizes

| Element | Minimum Size | Recommended |
|---|---|---|
| Button | 44x44px | 48x48px |
| Link | 44x44px | 48x48px |
| Icon button | 44x44px | 48x48px |
| Checkbox | 20x20px | 24x24px |
| Toggle | 44x24px | 48x24px |
| Tab | 44x44px | 48x48px |

### 12.4 Screen Reader Support

| Pattern | Implementation |
|---|---|
| Image alt text | All `<img>` have descriptive `alt` |
| ARIA labels | Interactive elements have `aria-label` |
| ARIA live regions | Toast notifications use `aria-live="polite"` |
| Role attributes | Semantic HTML + ARIA roles where needed |
| Keyboard navigation | All interactive elements reachable via Tab |
| Reduced motion | `prefers-reduced-motion` media query respected |

---

## 13. Component Variants Reference

### 13.1 Button Variants

| Variant | Background | Text | Border | Use Case |
|---|---|---|---|---|
| `default` | Primary | Primary-fg | None | Primary actions |
| `secondary` | Secondary | Secondary-fg | None | Secondary actions |
| `destructive` | Destructive | Destructive-fg | None | Delete, remove |
| `outline` | Transparent | Foreground | Border | Cancel, back |
| `ghost` | Transparent | Foreground | None | Icon buttons, nav |
| `link` | Transparent | Accent | None | Text links |

### 13.2 Badge Variants

| Variant | Background | Text | Usage |
|---|---|---|---|
| `default` | Primary | Primary-fg | Default status |
| `secondary` | Muted | Muted-fg | Neutral status |
| `destructive` | Destructive | Destructive-fg | Error, offline |
| `outline` | Transparent | Foreground | Tag, label |

### 13.3 Status Badge Mapping

| Status | Color | Badge Variant |
|---|---|---|
| `online` / `active` | Green | `success` (custom) |
| `offline` / `inactive` | Gray | `secondary` |
| `processing` / `running` | Blue | `default` |
| `error` / `failed` | Red | `destructive` |
| `warning` / `pending` | Amber | `warning` (custom) |

---

## 14. Dark Mode Implementation

### 14.1 Class Toggle Strategy

```tsx
// providers/ThemeProvider.tsx
"use client";
import { ThemeProvider as NextThemesProvider } from "next-themes";

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return (
    <NextThemesProvider
      attribute="class"
      defaultTheme="dark"
      enableSystem
      disableTransitionOnChange={false}
    >
      {children}
    </NextThemesProvider>
  );
}
```

### 14.2 Dark Mode Checklist

| Element | Approach |
|---|---|
| Background colors | CSS variables (`var(--background)`) |
| Text colors | CSS variables (`var(--foreground)`) |
| Border colors | CSS variables (`var(--border)`) |
| Images | Dark variants via `dark:` prefix or `<picture>` |
| Charts | Theme-aware color palette |
| Code blocks | Dark syntax highlighting theme |
| Shadows | Heavier shadows in dark mode |
| Focus rings | Adjusted ring color for dark backgrounds |

---

## 15. Responsive Design Patterns

### 15.1 Mobile-First Approach

```tsx
// Base = mobile, then override for larger screens
<div className="
  flex flex-col           // Mobile: stack vertically
  md:flex-row             // Tablet+: side by side
  gap-4                   // Mobile gap
  md:gap-6                // Tablet+ gap
  p-4                     // Mobile padding
  md:p-6                  // Tablet+ padding
">
```

### 15.2 Responsive Patterns

| Pattern | Mobile | Tablet | Desktop |
|---|---|---|---|
| Navigation | Bottom tabs | Collapsible sidebar | Fixed sidebar |
| Dashboard grid | 1 column | 2 columns | 4 columns |
| Chat layout | Full screen | Split view | 3-panel |
| Agent cards | 1 per row | 2 per row | 3 per row |
| Data table | Card view | Scrollable table | Full table |
| Modals | Full screen | Centered | Centered |

---

## 16. Enterprise Branding

### 16.1 Tenant Logo Integration

```tsx
function TenantLogo({ className }: { className?: string }) {
  const { branding } = useTenant();
  const { theme } = useTheme();

  const logoSrc = theme === "dark" ? branding.logo.dark : branding.logo.light;

  return (
    <img
      src={logoSrc || "/logos/nexus-logo-light.svg"}
      alt={`${branding.name} logo`}
      className={cn("h-8 w-auto", className)}
    />
  );
}
```

### 16.2 Custom Color Overrides

```css
/* Tenant-specific theme class */
.tenant-aeroxenexus {
  --primary: 222 47% 11%;
  --accent: 217 91% 60%;
}

.tenant-custom {
  --primary: 142 76% 36%;
  --accent: 262 83% 58%;
}
```

---

*Document Version: 1.0 — Last Updated: July 2026*
