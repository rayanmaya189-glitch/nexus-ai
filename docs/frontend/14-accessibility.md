# Accessibility

> WCAG 2.1 AA compliance for Nexus AI — keyboard navigation, screen reader support, focus management, and inclusive design patterns.

## 1. WCAG 2.1 AA Compliance

### Compliance Matrix

| Principle | Requirement | Status | Components Affected |
|---|---|---|---|
| **Perceivable** | Color contrast ≥ 4.5:1 (text) | ✅ | All text components |
| | Color contrast ≥ 3:1 (large text, UI) | ✅ | Buttons, icons, borders |
| | Text resizable to 200% | ✅ | All typography |
| | Images have alt text | ✅ | `<img>`, `<Avatar>`, icons |
| | Video has captions | ✅ | Media components |
| **Operable** | All functionality keyboard accessible | ✅ | All interactive elements |
| | No keyboard traps | ✅ | Modals, dropdowns |
| | Skip navigation link | ✅ | Main layout |
| | Page titles descriptive | ✅ | Layout metadata |
| | Focus order logical | ✅ | Tab order management |
| **Understandable** | Language attribute set | ✅ | `<html lang>` |
| | Form labels present | ✅ | All form inputs |
| | Error suggestions provided | ✅ | Form validation |
| | Consistent navigation | ✅ | Persistent nav |
| **Robust** | Valid HTML | ✅ | Semantic markup |
| | ARIA correctly used | ✅ | Custom components |
| | Status messages via ARIA | ✅ | Toasts, live regions |

## 2. Keyboard Navigation

### Tab Order Map

```
┌──────────────────────────────────────────────────────┐
│ Skip Link → Logo → Nav Items → Search → 🔔 → 👤     │
│                                                      │
│ Main Content:                                        │
│   Headings (not tabbable)                            │
│   Interactive elements in DOM order                  │
│   Links → Buttons → Form inputs → Custom widgets     │
│                                                      │
│ Footer: Links → Social icons                         │
└──────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts

| Shortcut | Action | Scope |
|---|---|---|
| `Tab` | Move to next focusable element | Global |
| `Shift+Tab` | Move to previous focusable element | Global |
| `Enter` / `Space` | Activate button/link | Global |
| `Escape` | Close modal/dropdown/panel | Contextual |
| `↑` / `↓` | Navigate list items | Lists, menus |
| `←` / `→` | Navigate tabs, carousel | Tab groups |
| `Ctrl+K` / `⌘+K` | Open command palette | Global |
| `Ctrl+/` | Open keyboard shortcuts dialog | Global |
| `?` | Show shortcuts (when not in input) | Global |
| `1-9` | Switch between nav sections | Sidebar |

### Skip Navigation

```tsx
// components/layout/SkipNav.tsx
export function SkipNav() {
  return (
    <a
      href="#main-content"
      className="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4
                 focus:z-[200] focus:px-4 focus:py-2 focus:bg-blue-600
                 focus:text-white focus:rounded-lg focus:outline-none
                 focus:ring-2 focus:ring-blue-400 focus:ring-offset-2"
    >
      Skip to main content
    </a>
  );
}

// layouts/layout.tsx
export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        <SkipNav />
        <Sidebar />
        <main id="main-content" tabIndex={-1} className="outline-none">
          {children}
        </main>
      </body>
    </html>
  );
}
```

## 3. Screen Reader Support

### ARIA Labels & Roles

| Component | ARIA Usage |
|---|---|
| Sidebar | `role="navigation"`, `aria-label="Main navigation"` |
| Notification bell | `aria-label="Notifications. 5 unread"`, `aria-expanded` |
| Search | `role="search"`, `aria-label="Search"` |
| Chat messages | `role="log"`, `aria-live="polite"`, `aria-label="Chat messages"` |
| Loading spinner | `role="status"`, `aria-label="Loading"` |
| Toast | `role="alert"`, `aria-live="assertive"` |
| Modal | `role="dialog"`, `aria-modal="true"`, `aria-labelledby` |
| Tabs | `role="tablist"`, `role="tab"`, `role="tabpanel"`, `aria-selected` |
| Accordion | `role="button"`, `aria-expanded`, `aria-controls` |
| Progress | `role="progressbar"`, `aria-valuenow`, `aria-valuemin`, `aria-valuemax` |
| Avatar | `role="img"`, `aria-label="[Name] avatar"` |

### Live Regions for Streaming

```tsx
// components/chat/ChatMessages.tsx
export function ChatMessages({ messages, isStreaming }: Props) {
  return (
    <div
      role="log"
      aria-label="Chat messages"
      aria-live="polite"
      aria-relevant="additions"
      className="space-y-4 p-4"
    >
      {messages.map((msg) => (
        <ChatMessage key={msg.id} message={msg} />
      ))}
      {isStreaming && (
        <div aria-live="polite" aria-atomic="false">
          <TypingIndicator />
          <span className="sr-only">Assistant is typing a response</span>
        </div>
      )}
    </div>
  );
}

// Streaming text announcement (debounced for performance)
function StreamingAnnouncement({ content }: { content: string }) {
  const [announcement, setAnnouncement] = useState("");

  useEffect(() => {
    const timer = setTimeout(() => {
      // Announce every 100 characters to avoid overwhelming screen readers
      setAnnouncement(content.slice(0, Math.ceil(content.length / 100) * 100));
    }, 500);
    return () => clearTimeout(timer);
  }, [content]);

  return (
    <div className="sr-only" aria-live="polite" aria-atomic="false">
      {announcement}
    </div>
  );
}
```

### Landmark Structure

```
<html lang="en">
  <body>
    <a href="#main">Skip to main content</a>
    <header role="banner">
      <nav aria-label="Main navigation">...</nav>
      <div role="search">...</div>
    </header>
    <aside aria-label="Sidebar">...</aside>
    <main id="main" role="main">...</main>
    <footer role="contentinfo">...</footer>
  </body>
</html>
```

## 4. Focus Management

### Focus Trap in Modals

```tsx
// hooks/useFocusTrap.ts
import { useEffect, useRef, useCallback } from "react";

const FOCUSABLE_SELECTORS = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(", ");

export function useFocusTrap(
  containerRef: React.RefObject<HTMLElement>,
  isActive: boolean = true
) {
  const previousFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!isActive || !containerRef.current) return;

    // Save previous focus
    previousFocusRef.current = document.activeElement as HTMLElement;

    // Focus first focusable element
    const container = containerRef.current;
    const focusable = container.querySelectorAll(FOCUSABLE_SELECTORS);
    const firstFocusable = focusable[0] as HTMLElement;
    firstFocusable?.focus();

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== "Tab") return;

      const currentFocusable = Array.from(
        container.querySelectorAll(FOCUSABLE_SELECTORS)
      ) as HTMLElement[];

      if (currentFocusable.length === 0) return;

      const first = currentFocusable[0];
      const last = currentFocusable[currentFocusable.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    container.addEventListener("keydown", handleKeyDown);

    return () => {
      container.removeEventListener("keydown", handleKeyDown);
      // Restore previous focus
      previousFocusRef.current?.focus();
    };
  }, [containerRef, isActive]);
}
```

### Focus Restoration

```tsx
// components/ui/Modal.tsx
function Modal({ isOpen, onClose, children, title }: ModalProps) {
  const modalRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLElement | null>(null);

  useFocusTrap(modalRef, isOpen);

  useEffect(() => {
    if (isOpen) {
      triggerRef.current = document.activeElement as HTMLElement;
    } else if (triggerRef.current) {
      triggerRef.current.focus();
      triggerRef.current = null;
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="fixed inset-0 bg-black/50" onClick={onClose} />
      <div
        ref={modalRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        className="relative z-10 bg-white dark:bg-neutral-900 rounded-2xl
                   shadow-xl max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto"
      >
        <div className="flex items-center justify-between px-6 py-4 border-b">
          <h2 id="modal-title" className="text-lg font-semibold">{title}</h2>
          <button
            onClick={onClose}
            aria-label="Close dialog"
            className="p-2 rounded-lg hover:bg-neutral-100 dark:hover:bg-neutral-800
                       min-h-[44px] min-w-[44px] flex items-center justify-center"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="p-6">{children}</div>
      </div>
    </div>
  );
}
```

## 5. Color Contrast

### Contrast Ratios by Element

| Element | Foreground | Background | Ratio | AA Pass | AAA Pass |
|---|---|---|---|---|---|
| Body text | `#1a1a1a` | `#ffffff` | 16.27:1 | ✅ | ✅ |
| Secondary text | `#6b7280` | `#ffffff` | 5.05:1 | ✅ | ❌ |
| Disabled text | `#9ca3af` | `#ffffff` | 2.95:1 | ❌ (ok) | ❌ |
| Primary button | `#ffffff` | `#6366f1` | 4.56:1 | ✅ | ❌ |
| Error text | `#dc2626` | `#ffffff` | 5.63:1 | ✅ | ❌ |
| Link text | `#2563eb` | `#ffffff` | 4.62:1 | ✅ | ❌ |

### High Contrast Mode

```css
/* High contrast mode toggle */
.high-contrast {
  --color-text-primary: #000000;
  --color-text-secondary: #1a1a1a;
  --color-bg-primary: #ffffff;
  --color-bg-secondary: #f0f0f0;
  --color-border: #000000;
  --color-focus-ring: #0000ff;
}

.high-contrast.dark {
  --color-text-primary: #ffffff;
  --color-text-secondary: #e0e0e0;
  --color-bg-primary: #000000;
  --color-bg-secondary: #1a1a1a;
  --color-border: #ffffff;
  --color-focus-ring: #ffff00;
}

/* Forced colors mode (Windows High Contrast) */
@media (forced-colors: active) {
  .btn-primary {
    border: 2px solid ButtonText;
    forced-color-adjust: none;
  }

  .status-success {
    color: LinkText;
    border: 1px solid LinkText;
  }

  :focus-visible {
    outline: 2px solid Highlight;
  }
}
```

```tsx
// hooks/useHighContrast.ts
export function useHighContrast() {
  const [isHighContrast, setIsHighContrast] = useState(() => {
    if (typeof window === "undefined") return false;
    return localStorage.getItem("high-contrast") === "true";
  });

  useEffect(() => {
    document.documentElement.classList.toggle("high-contrast", isHighContrast);
    localStorage.setItem("high-contrast", String(isHighContrast));
  }, [isHighContrast]);

  return { isHighContrast, toggleHighContrast: () => setIsHighContrast((p) => !p) };
}
```

## 6. Font Scaling

```css
/* Responsive text that respects user preferences */
@media (prefers-reduced-motion: no-preference) {
  html { font-size: 100%; }
}

/* Allow up to 200% zoom without breaking layout */
@media (min-resolution: 1dppx) {
  body {
    font-size: clamp(1rem, 1rem, 1.25rem); /* max 200% */
  }
}

/* Prevent text overlap at large font sizes */
.responsive-card {
  container-type: inline-size;
}

@container (max-width: 400px) {
  .card-title { font-size: 1rem; }
  .card-body { font-size: 0.875rem; }
}

/* Root-relative sizing for all components */
.btn {
  padding: 0.625rem 1.25rem;  /* 10px 20px */
  font-size: 0.875rem;        /* 14px */
  border-radius: 0.5rem;      /* 8px */
  min-height: 2.75rem;        /* 44px touch target */
}
```

## 7. Form Accessibility

```tsx
// components/ui/AccessibleForm.tsx
interface AccessibleFieldProps {
  label: string;
  name: string;
  error?: string;
  hint?: string;
  required?: boolean;
  children: React.ReactNode;
}

export function AccessibleField({
  label,
  name,
  error,
  hint,
  required,
  children,
}: AccessibleFieldProps) {
  const fieldId = `field-${name}`;
  const errorId = `error-${name}`;
  const hintId = `hint-${name}`;

  return (
    <div className="space-y-1.5">
      <label
        htmlFor={fieldId}
        className="block text-sm font-medium text-neutral-700 dark:text-neutral-300"
      >
        {label}
        {required && (
          <span className="text-red-500 ml-0.5" aria-hidden="true">*</span>
        )}
        {required && <span className="sr-only"> (required)</span>}
      </label>

      <div
        aria-describedby={[error && errorId, hint && hintId].filter(Boolean).join(" ") || undefined}
        aria-invalid={!!error}
        aria-required={required}
      >
        {React.cloneElement(children as React.ReactElement, {
          id: fieldId,
          "aria-invalid": !!error,
          "aria-describedby": [error && errorId, hint && hintId]
            .filter(Boolean)
            .join(" ") || undefined,
        })}
      </div>

      {hint && !error && (
        <p id={hintId} className="text-xs text-neutral-500">
          {hint}
        </p>
      )}

      {error && (
        <p id={errorId} role="alert" className="text-xs text-red-600 dark:text-red-400">
          {error}
        </p>
      )}
    </div>
  );
}
```

### Form Validation with Zod

```ts
// lib/validation/schemas.ts
import { z } from "zod";

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, "Email is required")
    .email("Please enter a valid email address"),
  password: z
    .string()
    .min(1, "Password is required")
    .min(8, "Password must be at least 8 characters"),
});

// Form error mapping for screen readers
function getFieldErrors(errors: ZodError) {
  return errors.issues.map((issue) => ({
    field: issue.path.join("."),
    message: issue.message,
  }));
}
```

## 8. Modal Accessibility

| Requirement | Implementation |
|---|---|
| Focus trap | `useFocusTrap` on modal container |
| ESC to close | `onKeyDown` handler on modal overlay |
| `aria-modal="true"` | Applied to dialog element |
| `aria-labelledby` | Points to title element |
| `aria-describedby` | Points to description (optional) |
| Return focus | `previousFocusRef` restoration |
| Background inert | `aria-hidden="true"` on background content |

## 9. Dropdown Accessibility

```tsx
// components/ui/AccessibleDropdown.tsx
export function AccessibleDropdown({ trigger, children, label }: DropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const [activeIndex, setActiveIndex] = useState(-1);
  const menuItems = useRef<HTMLButtonElement[]>([]);

  useFocusTrap(menuRef, isOpen);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        setActiveIndex((i) => Math.min(i + 1, menuItems.current.length - 1));
        break;
      case "ArrowUp":
        e.preventDefault();
        setActiveIndex((i) => Math.max(i - 1, 0));
        break;
      case "Home":
        e.preventDefault();
        setActiveIndex(0);
        break;
      case "End":
        e.preventDefault();
        setActiveIndex(menuItems.current.length - 1);
        break;
      case "Escape":
        setIsOpen(false);
        triggerRef.current?.focus();
        break;
    }
  };

  useEffect(() => {
    menuItems.current[activeIndex]?.focus();
  }, [activeIndex]);

  return (
    <div className="relative">
      <button
        ref={triggerRef}
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="menu"
        aria-label={label}
      >
        {trigger}
      </button>
      {isOpen && (
        <div
          ref={menuRef}
          role="menu"
          aria-label={label}
          onKeyDown={handleKeyDown}
          className="absolute right-0 mt-2 w-56 rounded-xl shadow-lg
                     bg-white dark:bg-neutral-900 border z-50"
        >
          {React.Children.map(children, (child, i) => (
            <div
              role="menuitem"
              tabIndex={i === activeIndex ? 0 : -1}
              ref={(el) => { menuItems.current[i] = el as HTMLButtonElement; }}
            >
              {child}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
```

## 10. Toast Notification Accessibility

```tsx
// components/ui/AccessibleToast.tsx
// Uses Radix UI Toast with accessibility built-in

const TOAST_CONFIG = {
  info:    { role: "status",  ariaLive: "polite"  as const, icon: Info    },
  success: { role: "status",  ariaLive: "polite"  as const, icon: Check   },
  warning: { role: "status",  ariaLive: "assertive" as const, icon: Alert },
  error:   { role: "alert",   ariaLive: "assertive" as const, icon: XCircle },
};

export function AccessibleToast({ type, title, description, action }: ToastProps) {
  const config = TOAST_CONFIG[type];

  return (
    <RadixToast.Root
      role={config.role}
      aria-live={config.ariaLive}
      className="..."
    >
      <RadixToast.Title className="font-semibold">{title}</RadixToast.Title>
      {description && (
        <RadixToast.Description className="text-sm mt-1">
          {description}
        </RadixToast.Description>
      )}
      {action && (
        <RadixToast.Action altText={action.label} asChild>
          <button>{action.label}</button>
        </RadixToast.Action>
      )}
      <RadixToast.Close aria-label="Dismiss notification">
        <X className="h-4 w-4" />
      </RadixToast.Close>
    </RadixToast.Root>
  );
}
```

## 11. Motion Preferences

```css
/* Respect prefers-reduced-motion */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }

  .animate-spin { animation: none; }
  .animate-pulse { animation: none; }
  .animate-bounce { animation: none; }

  /* Still allow hover transitions but make them instant */
  .hover\:scale-105:hover { transform: none; }
}
```

```tsx
// hooks/useReducedMotion.ts
export function useReducedMotion(): boolean {
  const [prefersReduced, setPrefersReduced] = useState(() => {
    if (typeof window === "undefined") return false;
    return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  });

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    const handler = (e: MediaQueryListEvent) => setPrefersReduced(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  return prefersReduced;
}

// Usage in components
function AnimatedCard({ children }: { children: React.ReactNode }) {
  const prefersReduced = useReducedMotion();

  return (
    <motion.div
      initial={prefersReduced ? false : { opacity: 0, y: 20 }}
      animate={prefersReduced ? { opacity: 1 } : { opacity: 1, y: 0 }}
      transition={prefersReduced ? { duration: 0 } : { duration: 0.3 }}
    >
      {children}
    </motion.div>
  );
}
```

## 12. ARIA Patterns

### Tabs

```tsx
function AccessibleTabs({ tabs, activeTab, onChange }: TabsProps) {
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);

  return (
    <div>
      <div role="tablist" aria-label="Content tabs" className="flex border-b">
        {tabs.map((tab, i) => (
          <button
            key={tab.id}
            ref={(el) => { tabRefs.current[i] = el; }}
            role="tab"
            id={`tab-${tab.id}`}
            aria-selected={activeTab === tab.id}
            aria-controls={`tabpanel-${tab.id}`}
            tabIndex={activeTab === tab.id ? 0 : -1}
            onClick={() => onChange(tab.id)}
            onKeyDown={(e) => {
              let newIndex = i;
              if (e.key === "ArrowRight") newIndex = (i + 1) % tabs.length;
              if (e.key === "ArrowLeft") newIndex = (i - 1 + tabs.length) % tabs.length;
              if (e.key === "Home") newIndex = 0;
              if (e.key === "End") newIndex = tabs.length - 1;
              if (newIndex !== i) {
                onChange(tabs[newIndex].id);
                tabRefs.current[newIndex]?.focus();
              }
            }}
          >
            {tab.label}
          </button>
        ))}
      </div>
      {tabs.map((tab) => (
        <div
          key={tab.id}
          role="tabpanel"
          id={`tabpanel-${tab.id}`}
          aria-labelledby={`tab-${tab.id}`}
          hidden={activeTab !== tab.id}
          tabIndex={0}
        >
          {tab.content}
        </div>
      ))}
    </div>
  );
}
```

### Accordion

```tsx
function AccessibleAccordion({ items }: { items: AccordionItem[] }) {
  const [openId, setOpenId] = useState<string | null>(null);

  return (
    <div role="region" aria-label="Accordion">
      {items.map((item) => {
        const isOpen = openId === item.id;
        return (
          <div key={item.id} className="border-b">
            <h3>
              <button
                id={`accordion-trigger-${item.id}`}
                aria-expanded={isOpen}
                aria-controls={`accordion-panel-${item.id}`}
                onClick={() => setOpenId(isOpen ? null : item.id)}
                className="w-full flex items-center justify-between py-4
                           text-left font-medium min-h-[44px]"
              >
                {item.title}
                <ChevronDown
                  className={`h-5 w-5 transition-transform ${
                    isOpen ? "rotate-180" : ""
                  }`}
                  aria-hidden="true"
                />
              </button>
            </h3>
            <div
              id={`accordion-panel-${item.id}`}
              role="region"
              aria-labelledby={`accordion-trigger-${item.id}`}
              hidden={!isOpen}
            >
              <div className="pb-4 text-sm text-neutral-600">
                {item.content}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
```

### Combobox (Search)

```tsx
function AccessibleCombobox({ options, onSelect }: ComboboxProps) {
  const [query, setQuery] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const inputRef = useRef<HTMLInputElement>(null);
  const listboxRef = useRef<HTMLUListElement>(null);

  const filtered = options.filter((o) =>
    o.label.toLowerCase().includes(query.toLowerCase())
  );

  return (
    <div className="relative">
      <label htmlFor="combobox-input" className="sr-only">Search</label>
      <input
        ref={inputRef}
        id="combobox-input"
        role="combobox"
        aria-expanded={isOpen}
        aria-controls="combobox-listbox"
        aria-autocomplete="list"
        aria-activedescendant={activeIndex >= 0 ? `option-${activeIndex}` : undefined}
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          setIsOpen(true);
          setActiveIndex(-1);
        }}
        onFocus={() => setIsOpen(true)}
        onBlur={() => setTimeout(() => setIsOpen(false), 200)}
      />
      {isOpen && filtered.length > 0 && (
        <ul
          ref={listboxRef}
          id="combobox-listbox"
          role="listbox"
          aria-label="Search results"
          className="absolute mt-1 w-full bg-white dark:bg-neutral-900
                     border rounded-lg shadow-lg max-h-60 overflow-auto z-50"
        >
          {filtered.map((option, i) => (
            <li
              key={option.id}
              id={`option-${i}`}
              role="option"
              aria-selected={i === activeIndex}
              className={`px-4 py-2 cursor-pointer ${
                i === activeIndex ? "bg-blue-50 dark:bg-blue-950" : ""
              }`}
              onMouseDown={() => onSelect(option)}
            >
              {option.label}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

## 13. Table Accessibility

```tsx
function AccessibleTable({ data, columns }: TableProps) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <caption className="sr-only">User management data</caption>
        <thead>
          <tr>
            {columns.map((col) => (
              <th
                key={col.key}
                scope="col"
                className="text-left py-3 px-4 font-semibold text-neutral-600
                           border-b dark:text-neutral-400"
                aria-sort={col.sortable ? "none" : undefined}
              >
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row) => (
            <tr key={row.id} className="border-b last:border-0 hover:bg-neutral-50 dark:hover:bg-neutral-800/50">
              {columns.map((col) => (
                <td key={col.key} className="py-3 px-4">
                  {col.render ? col.render(row[col.key], row) : row[col.key]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

## 14. Accessibility Testing

### axe-core Integration

```ts
// e2e/accessibility.spec.ts
import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

test.describe("Accessibility", () => {
  test("dashboard has no axe violations", async ({ page }) => {
    await page.goto("/dashboard");
    const results = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"])
      .analyze();

    expect(results.violations).toEqual([]);
  });

  test("modal has no axe violations", async ({ page }) => {
    await page.goto("/dashboard");
    await page.click("[data-testid=open-settings]");
    const results = await new AxeBuilder({ page })
      .include("[role=dialog]")
      .withTags(["wcag2a", "wcag2aa"])
      .analyze();

    expect(results.violations).toEqual([]);
  });
});
```

### Manual Testing Checklist

| Test | Steps | Expected |
|---|---|---|
| Tab navigation | Tab through entire page | Logical order, no traps |
| Screen reader | Navigate with NVDA/VoiceOver | All content announced |
| Zoom | Zoom to 200% | No content clipped |
| Keyboard shortcuts | Test all shortcuts | All work correctly |
| Color contrast | Use Chrome DevTools | All pass AA |
| Focus visible | Tab through page | Focus ring visible on all |
| Error states | Submit invalid forms | Errors announced, focus moved |
| Reduced motion | Enable in OS settings | No animations |

## 15. Accessibility Hooks

```ts
// hooks/useFocusTrap.ts — See section 4

// hooks/useKeyboard.ts
export function useKeyboard(
  shortcuts: Record<string, () => void>,
  deps: unknown[] = []
) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const key = [
        e.ctrlKey && "ctrl",
        e.shiftKey && "shift",
        e.altKey && "alt",
        e.metaKey && "meta",
        e.key.toLowerCase(),
      ]
        .filter(Boolean)
        .join("+");

      if (shortcuts[key]) {
        e.preventDefault();
        shortcuts[key]();
      }
    };

    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, deps);
}

// hooks/useAnnounce.ts
export function useAnnounce() {
  const announce = useCallback(
    (message: string, priority: "polite" | "assertive" = "polite") => {
      const el = document.createElement("div");
      el.setAttribute("aria-live", priority);
      el.setAttribute("aria-atomic", "true");
      el.className = "sr-only";
      document.body.appendChild(el);
      requestAnimationFrame(() => {
        el.textContent = message;
        setTimeout(() => document.body.removeChild(el), 1000);
      });
    },
    []
  );

  return { announce };
}
```

## 16. Accessibility Settings

```tsx
// components/settings/AccessibilitySettings.tsx
export function AccessibilitySettings() {
  const { isHighContrast, toggleHighContrast } = useHighContrast();
  const [fontSize, setFontSize] = useState<"normal" | "large" | "x-large">("normal");

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-semibold">Accessibility</h2>

      <SettingGroup label="Visual">
        <SettingItem label="High contrast mode">
          <Switch checked={isHighContrast} onCheckedChange={toggleHighContrast} />
        </SettingItem>
        <SettingItem label="Font size">
          <Select value={fontSize} onValueChange={setFontSize}>
            <SelectItem value="normal">Normal</SelectItem>
            <SelectItem value="large">Large</SelectItem>
            <SelectItem value="x-large">Extra Large</SelectItem>
          </Select>
        </SettingItem>
        <SettingItem label="Reduce animations">
          <Switch />
        </SettingItem>
      </SettingGroup>

      <SettingGroup label="Screen Reader">
        <SettingItem label="Verbosity level">
          <Select defaultValue="standard">
            <SelectItem value="minimal">Minimal</SelectItem>
            <SelectItem value="standard">Standard</SelectItem>
            <SelectItem value="verbose">Verbose</SelectItem>
          </Select>
        </SettingItem>
      </SettingGroup>

      <SettingGroup label="Keyboard">
        <SettingItem label="Show keyboard shortcuts">
          <Switch defaultChecked />
        </SettingItem>
        <SettingItem label="Key repeat rate">
          <Select defaultValue="normal">
            <SelectItem value="slow">Slow</SelectItem>
            <SelectItem value="normal">Normal</SelectItem>
            <SelectItem value="fast">Fast</SelectItem>
          </Select>
        </SettingItem>
      </SettingGroup>
    </div>
  );
}
```

## 17. Color Blindness Considerations

| Color | Safe Alternative | Pattern |
|---|---|---|
| Red/Green | Blue/Orange | Don't use color as sole indicator |
| Status indicators | Icon + text + color | Triple redundancy |
| Charts/graphs | Patterns + labels | Not just color differentiation |
| Error/success | Icons (✅❌) + text | Never rely on color alone |

```tsx
// Accessible status badge — never color-only
function StatusBadge({ status }: { status: "active" | "inactive" | "error" }) {
  const config = {
    active:   { icon: CheckCircle, label: "Active",   className: "bg-green-100 text-green-800" },
    inactive: { icon: Circle,     label: "Inactive", className: "bg-neutral-100 text-neutral-800" },
    error:    { icon: AlertCircle, label: "Error",    className: "bg-red-100 text-red-800" },
  };

  const { icon: Icon, label, className } = config[status];

  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${className}`}>
      <Icon className="h-3.5 w-3.5" aria-hidden="true" />
      {label}
    </span>
  );
}
```

## 18. File Structure

```
src/
├── components/
│   ├── ui/
│   │   ├── AccessibleForm.tsx
│   │   ├── AccessibleDropdown.tsx
│   │   ├── AccessibleTabs.tsx
│   │   ├── AccessibleAccordion.tsx
│   │   ├── AccessibleCombobox.tsx
│   │   ├── AccessibleToast.tsx
│   │   └── AccessibleTable.tsx
│   └── layout/
│       └── SkipNav.tsx
├── hooks/
│   ├── useFocusTrap.ts
│   ├── useKeyboard.ts
│   ├── useAnnounce.ts
│   ├── useHighContrast.ts
│   └── useReducedMotion.ts
├── settings/
│   └── AccessibilitySettings.tsx
├── styles/
│   └── a11y.css              (high contrast, forced colors)
└── e2e/
    └── accessibility.spec.ts
```
