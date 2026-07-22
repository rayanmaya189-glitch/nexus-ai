# Knowledge Center

## Table of Contents

- [Overview](#overview)
- [Page Layout](#page-layout)
- [Document Upload](#document-upload)
- [Document Processing Pipeline](#document-processing-pipeline)
- [Document List View](#document-list-view)
- [Document Detail View](#document-detail-view)
- [Document Search](#document-search)
- [Document Sets](#document-sets)
- [Document Permissions](#document-permissions)
- [Document Deletion](#document-deletion)
- [Supported File Types](#supported-file-types)
- [File Size Limits and Validation](#file-size-limits-and-validation)
- [Upload Progress Tracking](#upload-progress-tracking)
- [Processing Pipeline Visualization](#processing-pipeline-visualization)
- [Document Chunk Viewer](#document-chunk-viewer)
- [Document Metadata Editor](#document-metadata-editor)
- [Bulk Operations](#bulk-operations)
- [API Integration](#api-integration)
- [Hooks](#hooks)
- [Stores](#stores)
- [Error Handling](#error-handling)
- [Accessibility](#accessibility)

---

## Overview

The Knowledge Center is the central hub for managing documents that power RAG (Retrieval-Augmented Generation) workflows. It provides full lifecycle management from upload through processing to search and retrieval.

```
┌─────────────────────────────────────────────────────────┐
│  Knowledge Center                                       │
├────────┬────────────────────────────────────────────────┤
│ Sidebar│  [Search]  [Upload]  [Sets]  [Bulk Actions]   │
│        ├────────────────────────────────────────────────┤
│ Docs   │  ┌─────┬─────┬─────┬─────┬─────┬─────┬─────┐ │
│ Sets   │  │Name │Type │Status│Chunks│Tags │Date │Act  │ │
│ Tags   │  ├─────┼─────┼─────┼─────┼─────┼─────┼─────┤ │
│ Recent │  │ ... │ ... │ ... │ ... │ ... │ ... │ ... │ │
│        │  └─────┴─────┴─────┴─────┴─────┴─────┴─────┘ │
│        │  ◄ 1 2 3 ... 10 ►          Showing 1-20/200  │
└────────┴────────────────────────────────────────────────┘
```

---

## Page Layout

The Knowledge Center uses a sidebar + main content layout with a top action bar.

### Component Tree

```
KnowledgeCenterPage
├── KnowledgeSidebar
│   ├── NavigationLinks (Docs, Sets, Tags, Recent)
│   ├── QuickStats (total docs, processing, ready)
│   └── StorageIndicator (used / total)
├── MainContent
│   ├── ActionBar
│   │   ├── SearchBar
│   │   ├── FilterDropdown
│   │   ├── UploadButton
│   │   ├── BulkActionsToolbar
│   │   └── ViewToggle (list/grid)
│   ├── DocumentList
│   │   ├── DocumentTableHeader
│   │   ├── DocumentRow[] (or DocumentCard[])
│   │   └── Pagination
│   └── EmptyState (when no documents)
└── UploadModal / UploadDrawer
```

### Layout Code

```tsx
// pages/knowledge/KnowledgeCenterPage.tsx
import { useState } from 'react';
import { KnowledgeSidebar } from '@/components/knowledge/KnowledgeSidebar';
import { DocumentList } from '@/components/knowledge/DocumentList';
import { UploadModal } from '@/components/knowledge/UploadModal';
import { ActionBar } from '@/components/knowledge/ActionBar';
import { useDocuments } from '@/hooks/knowledge/useDocuments';

export function KnowledgeCenterPage() {
  const [showUpload, setShowUpload] = useState(false);
  const [view, setView] = useState<'list' | 'grid'>('list');
  const [filters, setFilters] = useState<DocumentFilters>({});
  const { documents, isLoading } = useDocuments(filters);

  return (
    <div className="flex h-screen bg-background">
      <KnowledgeSidebar />
      <main className="flex-1 overflow-hidden flex flex-col">
        <ActionBar
          onUpload={() => setShowUpload(true)}
          view={view}
          onViewChange={setView}
          filters={filters}
          onFilterChange={setFilters}
        />
        <DocumentList
          documents={documents}
          view={view}
          isLoading={isLoading}
        />
      </main>
      {showUpload && (
        <UploadModal onClose={() => setShowUpload(false)} />
      )}
    </div>
  );
}
```

### Sidebar Stats

| Stat              | Description                   | Color    |
|-------------------|-------------------------------|----------|
| Total Documents   | All uploaded documents        | Default  |
| Processing        | Currently being processed     | Yellow   |
| Ready             | Searchable and available      | Green    |
| Failed            | Processing errors             | Red      |
| Storage Used      | Disk space consumed           | Blue     |

---

## Document Upload

### Upload Modal Layout

```
┌──────────────────────────────────────────────┐
│  Upload Documents                       [X]  │
├──────────────────────────────────────────────┤
│                                              │
│  ┌────────────────────────────────────────┐  │
│  │                                        │  │
│  │     📁 Drop files here or click       │  │
│  │        to browse                       │  │
│  │                                        │  │
│  │   Supports: PDF, DOCX, MD, HTML, Code │  │
│  │   Max size: 50MB per file             │  │
│  │                                        │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  Target Set: [Select or create set     ▼]   │
│  Tags:        [tag1] [tag2] [+ Add tag]     │
│                                              │
│  ┌────────────────────────────────────────┐  │
│  │ Selected Files (3):                    │  │
│  │  ✅ report.pdf        2.3MB  [Remove] │  │
│  │  ✅ docs.md           156KB  [Remove] │  │
│  │  ✅ code.py           89KB   [Remove] │  │
│  │                                        │  │
│  │  Total: 2.5MB                          │  │
│  └────────────────────────────────────────┘  │
│                                              │
│         [Cancel]           [Upload 3 Files]  │
└──────────────────────────────────────────────┘
```

### Drag and Drop Component

```tsx
// components/knowledge/DropZone.tsx
import { useCallback, useState } from 'react';
import { ACCEPTED_TYPES, MAX_FILE_SIZE } from '@/lib/knowledge/constants';
import { validateFile } from '@/lib/knowledge/validation';

interface DropZoneProps {
  onFilesSelected: (files: File[]) => void;
  disabled?: boolean;
}

export function DropZone({ onFilesSelected, disabled }: DropZoneProps) {
  const [isDragOver, setIsDragOver] = useState(false);
  const [errors, setErrors] = useState<string[]>([]);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragOver(false);
      if (disabled) return;

      const files = Array.from(e.dataTransfer.files);
      const validFiles: File[] = [];
      const newErrors: string[] = [];

      files.forEach((file) => {
        const result = validateFile(file);
        if (result.valid) {
          validFiles.push(file);
        } else {
          newErrors.push(`${file.name}: ${result.error}`);
        }
      });

      setErrors(newErrors);
      if (validFiles.length > 0) {
        onFilesSelected(validFiles);
      }
    },
    [disabled, onFilesSelected]
  );

  const handleFileInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = Array.from(e.target.files || []);
      onFilesSelected(files);
      e.target.value = '';
    },
    [onFilesSelected]
  );

  return (
    <div
      onDragOver={(e) => { e.preventDefault(); setIsDragOver(true); }}
      onDragLeave={() => setIsDragOver(false)}
      onDrop={handleDrop}
      className={`
        border-2 border-dashed rounded-xl p-12 text-center
        transition-colors cursor-pointer
        ${isDragOver ? 'border-primary bg-primary/5' : 'border-muted-foreground/25'}
        ${disabled ? 'opacity-50 cursor-not-allowed' : 'hover:border-primary/50'}
      `}
      role="button"
      aria-label="Drop files here or click to browse"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          document.getElementById('file-input')?.click();
        }
      }}
    >
      <input
        id="file-input"
        type="file"
        multiple
        accept={Object.keys(ACCEPTED_TYPES).join(',')}
        onChange={handleFileInput}
        className="hidden"
        aria-hidden="true"
      />
      <p className="text-lg font-medium">Drop files here or click to browse</p>
      <p className="text-sm text-muted-foreground mt-2">
        Supports: PDF, DOCX, Markdown, HTML, Code files — Max 50MB
      </p>
      {errors.length > 0 && (
        <ul className="mt-4 text-sm text-destructive text-left max-h-32 overflow-y-auto">
          {errors.map((err, i) => <li key={i}>{err}</li>)}
        </ul>
      )}
    </div>
  );
}
```

### File Selector (Click to Browse)

The file selector is integrated into the drop zone. It uses a hidden `<input type="file">` triggered by clicking the drop zone area or pressing Enter/Space when focused.

---

## Document Processing Pipeline

### Pipeline Stages

```
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ Uploaded │──▶│ Extracted│──▶│ Chunked  │──▶│ Embedded │──▶│  Ready   │
│          │   │          │   │          │   │          │   │          │
│ File on  │   │ Text     │   │ Split    │   │ Vectors  │   │ Search-  │
│ disk     │   │ content  │   │ into     │   │ created  │   │ able     │
│          │   │ parsed   │   │ chunks   │   │          │   │          │
└──────────┘   └──────────┘   └──────────┘   └──────────┘   └──────────┘
     │              │              │              │              │
     ▼              ▼              ▼              ▼              ▼
  Status:       Status:       Status:       Status:       Status:
  uploaded      extracted     chunked       embedded      ready
```

### Processing Status Codes

| Status      | Description                       | Retryable |
|-------------|-----------------------------------|-----------|
| `uploaded`  | File uploaded, awaiting processing | Yes      |
| `extracting`| Extracting text content            | Yes      |
| `extracted` | Text extraction complete           | —        |
| `chunking`  | Splitting into chunks              | Yes      |
| `chunked`   | Chunking complete                  | —        |
| `embedding` | Generating vector embeddings       | Yes      |
| `embedded`  | Embedding complete                 | —        |
| `ready`     | Document fully processed           | —        |
| `failed`    | Processing failed at some stage    | Yes      |

### Processing API Response

```json
{
  "document_id": "doc_abc123",
  "status": "embedding",
  "progress": {
    "stage": "embedding",
    "percentage": 65,
    "chunks_total": 42,
    "chunks_embedded": 27,
    "started_at": "2026-07-16T10:30:00Z",
    "estimated_remaining_seconds": 12
  },
  "history": [
    { "stage": "uploaded", "at": "2026-07-16T10:29:55Z" },
    { "stage": "extracted", "at": "2026-07-16T10:30:01Z", "duration_ms": 6000 },
    { "stage": "chunked", "at": "2026-07-16T10:30:05Z", "duration_ms": 4000 },
    { "stage": "embedding", "at": "2026-07-16T10:30:06Z", "duration_ms": null }
  ]
}
```

---

## Document List View

### Table Columns

| Column     | Width    | Sortable | Description                              |
|------------|----------|----------|------------------------------------------|
| Checkbox   | 40px     | No       | Bulk selection                           |
| Name       | Flexible | Yes      | Document filename                        |
| Type       | 100px    | Yes      | File type badge (PDF, MD, etc.)          |
| Status     | 120px    | Yes      | Processing status badge                  |
| Chunks     | 80px     | Yes      | Number of chunks (when ready)            |
| Size       | 80px     | Yes      | Original file size                       |
| Tags       | 150px    | No       | Tag pills                                |
| Set        | 120px    | Yes      | Document set name                        |
| Uploaded   | 100px    | Yes      | Upload date                              |
| Actions    | 80px     | No       | View / Edit / Delete                     |

### Document Row Component

```tsx
// components/knowledge/DocumentRow.tsx
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { StatusBadge } from '@/components/knowledge/StatusBadge';
import { TagPills } from '@/components/knowledge/TagPills';
import { DocumentActions } from '@/components/knowledge/DocumentActions';
import { useNavigate } from 'react-router-dom';

interface DocumentRowProps {
  document: Document;
  selected: boolean;
  onSelect: (id: string, selected: boolean) => void;
}

export function DocumentRow({ document, selected, onSelect }: DocumentRowProps) {
  const navigate = useNavigate();

  return (
    <tr
      className="hover:bg-muted/50 cursor-pointer"
      onClick={() => navigate(`/knowledge/${document.id}`)}
    >
      <td onClick={(e) => e.stopPropagation()}>
        <Checkbox
          checked={selected}
          onCheckedChange={(checked) => onSelect(document.id, !!checked)}
          aria-label={`Select ${document.name}`}
        />
      </td>
      <td className="font-medium">{document.name}</td>
      <td><Badge variant="outline">{document.type}</Badge></td>
      <td><StatusBadge status={document.status} /></td>
      <td className="text-muted-foreground">{document.chunk_count ?? '—'}</td>
      <td className="text-muted-foreground">{formatSize(document.size)}</td>
      <td><TagPills tags={document.tags} /></td>
      <td className="text-muted-foreground">{document.set_name ?? '—'}</td>
      <td className="text-muted-foreground">{formatDate(document.created_at)}</td>
      <td onClick={(e) => e.stopPropagation()}>
        <DocumentActions document={document} />
      </td>
    </tr>
  );
}
```

---

## Document Detail View

### Detail Page Layout

```
┌─────────────────────────────────────────────────────────┐
│  ← Back to Documents                                    │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  📄 report.pdf                                          │
│  Status: ● Ready    Type: PDF    Size: 2.3MB           │
│  Uploaded: Jul 16, 2026 by john@example.com            │
│                                                         │
│  [Overview] [Chunks] [Metadata] [Processing Log] [Set] │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Overview Tab:                                          │
│  ┌─────────────────────────────────────────────┐       │
│  │  Description: Q2 financial report           │       │
│  │  Tags: [finance] [quarterly] [2026]         │       │
│  │  Set: Finance Documents                     │       │
│  │  Chunks: 42     Words: 12,450               │       │
│  │  Embedding Model: text-embedding-3-small     │       │
│  │  Last Updated: Jul 16, 2026 10:30 AM        │       │
│  └─────────────────────────────────────────────┘       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Metadata Fields

| Field            | Type     | Editable | Description                       |
|------------------|----------|----------|-----------------------------------|
| Name             | string   | Yes      | Display name                      |
| Description      | text     | Yes      | User-provided description         |
| Tags             | string[] | Yes      | Classification tags               |
| Category         | enum     | Yes      | Document category                 |
| Language         | string   | Auto     | Detected language                 |
| Chunk Count      | number   | Auto     | Number of chunks                  |
| Word Count       | number   | Auto     | Total words                       |
| Embedding Model  | string   | Yes      | Model used for embeddings         |
| Chunk Size       | number   | Yes      | Tokens per chunk                  |
| Chunk Overlap    | number   | Yes      | Overlap between chunks            |

---

## Document Search

### Search Modes

| Mode         | Method           | Best For                        |
|--------------|------------------|----------------------------------|
| Semantic     | Vector similarity| Conceptual queries, natural lang |
| Keyword      | Full-text (BM25) | Exact term matching              |
| Hybrid       | Combined         | Best overall results             |

### Search Component

```tsx
// components/knowledge/DocumentSearch.tsx
import { useState, useCallback } from 'react';
import { useDocumentSearch } from '@/hooks/knowledge/useDocumentSearch';
import { SearchFilters } from '@/components/knowledge/SearchFilters';
import { SearchResultList } from '@/components/knowledge/SearchResultList';

export function DocumentSearch() {
  const [query, setQuery] = useState('');
  const [mode, setMode] = useState<'semantic' | 'keyword' | 'hybrid'>('hybrid');
  const [filters, setFilters] = useState<SearchFilters>({});
  const { results, isSearching, search } = useDocumentSearch();

  const handleSearch = useCallback(() => {
    search({ query, mode, filters });
  }, [query, mode, filters, search]);

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          placeholder="Search documents..."
          aria-label="Search query"
        />
        <Select value={mode} onValueChange={setMode}>
          <SelectTrigger className="w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="semantic">Semantic</SelectItem>
            <SelectItem value="keyword">Keyword</SelectItem>
            <SelectItem value="hybrid">Hybrid</SelectItem>
          </SelectContent>
        </Select>
        <Button onClick={handleSearch} disabled={isSearching}>
          {isSearching ? 'Searching...' : 'Search'}
        </Button>
      </div>
      <SearchFilters filters={filters} onChange={setFilters} />
      <SearchResultList results={results} />
    </div>
  );
}
```

### Search Filters

| Filter     | Type     | Options                                    |
|------------|----------|--------------------------------------------|
| File Type  | multi    | PDF, DOCX, MD, HTML, Code                 |
| Status     | multi    | Ready, Processing, Failed                  |
| Tags       | multi    | User-defined tags                          |
| Set        | single   | Document set name                          |
| Date Range | range    | Start date — End date                      |
| Uploaded By| single   | User email                                 |
| Min Chunks | number   | Minimum chunk count                        |

### Search Results

```json
{
  "results": [
    {
      "document_id": "doc_abc123",
      "document_name": "report.pdf",
      "chunk_id": "chunk_015",
      "chunk_index": 14,
      "content": "Revenue increased by 15% compared to Q1...",
      "score": 0.92,
      "match_type": "semantic",
      "highlights": ["Revenue increased by <mark>15%</mark> compared to Q1"]
    }
  ],
  "total": 3,
  "query_time_ms": 45
}
```

---

## Document Sets

### Create Set Dialog

```
┌──────────────────────────────────────────┐
│  Create Document Set                [X]  │
├──────────────────────────────────────────┤
│                                          │
│  Name:        [Finance Documents    ]    │
│  Description: [Q2 reports and docs  ]    │
│  Icon:        📊 (click to change)       │
│  Color:       #3B82F6 (click to pick)   │
│                                          │
│  Add Documents:                          │
│  ┌────────────────────────────────────┐  │
│  │ 🔍 Search documents...            │  │
│  │                                    │  │
│  │  ☑ report.pdf                      │  │
│  │  ☑ budget.xlsx                     │  │
│  │  ☐ meeting-notes.md                │  │
│  │  ☑ quarterly-review.docx           │  │
│  └────────────────────────────────────┘  │
│                                          │
│         [Cancel]           [Create Set]  │
└──────────────────────────────────────────┘
```

### Set Detail View

```
┌─────────────────────────────────────────────────────────┐
│  📊 Finance Documents                                   │
│  Q2 reports and docs                    [Edit] [Delete] │
├─────────────────────────────────────────────────────────┤
│  Documents: 3    Total Chunks: 124    Size: 8.2MB     │
├─────────────────────────────────────────────────────────┤
│  [Documents] [Settings] [Usage Stats]                   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │ Name              │ Chunks │ Size  │ Added       │  │
│  ├───────────────────┼────────┼───────┼─────────────┤  │
│  │ report.pdf        │ 42     │ 2.3MB │ Jul 16      │  │
│  │ budget.xlsx       │ 18     │ 1.1MB │ Jul 15      │  │
│  │ quarterly-review  │ 64     │ 4.8MB │ Jul 14      │  │
│  └───────────────────────────────────────────────────┘  │
│                                                         │
│  [+ Add Documents]    [Remove Selected]                 │
└─────────────────────────────────────────────────────────┘
```

---

## Document Permissions

### Permission Matrix

| Action        | Owner | Admin | Editor | Viewer |
|---------------|-------|-------|--------|--------|
| View          | ✅    | ✅    | ✅     | ✅     |
| Download      | ✅    | ✅    | ✅     | ❌     |
| Edit Metadata | ✅    | ✅    | ✅     | ❌     |
| Search        | ✅    | ✅    | ✅     | ✅     |
| Delete        | ✅    | ✅    | ❌     | ❌     |
| Share         | ✅    | ✅    | ❌     | ❌     |
| Manage Sets   | ✅    | ✅    | ❌     | ❌     |

### Sharing Dialog

```tsx
// components/knowledge/ShareDialog.tsx
interface ShareDialogProps {
  documentId: string;
  currentShares: DocumentShare[];
  onShare: (email: string, permission: Permission) => Promise<void>;
  onRevoke: (shareId: string) => Promise<void>;
}

export function ShareDialog({ documentId, currentShares, onShare, onRevoke }: ShareDialogProps) {
  const [email, setEmail] = useState('');
  const [permission, setPermission] = useState<Permission>('viewer');

  return (
    <Dialog>
      <DialogHeader>
        <DialogTitle>Share Document</DialogTitle>
      </DialogHeader>
      <div className="space-y-4">
        <div className="flex gap-2">
          <Input
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Email address"
            type="email"
          />
          <Select value={permission} onValueChange={setPermission}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="viewer">Viewer</SelectItem>
              <SelectItem value="editor">Editor</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
          <Button onClick={() => onShare(email, permission)}>Share</Button>
        </div>
        <div className="space-y-2">
          {currentShares.map((share) => (
            <div key={share.id} className="flex justify-between items-center">
              <span>{share.email}</span>
              <Badge>{share.permission}</Badge>
              <Button variant="ghost" size="sm" onClick={() => onRevoke(share.id)}>
                Revoke
              </Button>
            </div>
          ))}
        </div>
      </div>
    </Dialog>
  );
}
```

---

## Document Deletion

### Soft Delete Flow

```
User clicks Delete
       │
       ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ Confirmation │────▶│  Soft Delete │────▶│  Moved to    │
│ Dialog       │     │  (API call)  │     │  Trash Bin   │
│              │     │              │     │              │
│ Are you sure?│     │ status:      │     │ Recoverable  │
│ [Cancel]     │     │   deleted    │     │ for 30 days  │
│ [Delete]     │     │              │     │              │
└──────────────┘     └──────────────┘     └──────────────┘
                            │
                     After 30 days
                            │
                            ▼
                     ┌──────────────┐
                     │ Hard Delete  │
                     │ (permanent)  │
                     └──────────────┘
```

### Confirmation Dialog

```tsx
// components/knowledge/DeleteConfirmDialog.tsx
import { AlertDialog, AlertDialogAction, AlertDialogCancel,
         AlertDialogContent, AlertDialogDescription,
         AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog';

interface DeleteConfirmDialogProps {
  documentName: string;
  onConfirm: () => Promise<void>;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function DeleteConfirmDialog({ documentName, onConfirm, open, onOpenChange }: DeleteConfirmDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Document</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete <strong>{documentName}</strong>?
            It will be moved to trash and permanently deleted after 30 days.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
```

---

## Supported File Types

| Category  | Extensions         | MIME Types                              | Max Size |
|-----------|--------------------|-----------------------------------------|----------|
| PDF       | `.pdf`             | `application/pdf`                       | 50MB     |
| Word      | `.docx`            | `application/vnd.openxmlformats-...`    | 50MB     |
| Markdown  | `.md`, `.markdown` | `text/markdown`                         | 10MB     |
| HTML      | `.html`, `.htm`    | `text/html`                             | 10MB     |
| Code      | `.py`, `.js`, `.ts`, `.go`, `.rs`, `.java`, `.cpp`, `.c` | `text/plain` | 5MB |
| Text      | `.txt`, `.csv`     | `text/plain`, `text/csv`               | 10MB     |
| JSON      | `.json`            | `application/json`                      | 10MB     |
| XML       | `.xml`             | `application/xml`                       | 10MB     |

---

## File Size Limits and Validation

```typescript
// lib/knowledge/validation.ts
export const FILE_LIMITS = {
  maxFileSize: 50 * 1024 * 1024,       // 50MB
  maxTotalUploadSize: 200 * 1024 * 1024, // 200MB total
  maxFilesPerUpload: 20,
  acceptedTypes: {
    'application/pdf': '.pdf',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document': '.docx',
    'text/markdown': '.md',
    'text/html': '.html',
    'text/plain': '.txt',
    'text/csv': '.csv',
    'application/json': '.json',
    'application/xml': '.xml',
  },
};

export interface ValidationResult {
  valid: boolean;
  error?: string;
}

export function validateFile(file: File): ValidationResult {
  if (file.size > FILE_LIMITS.maxFileSize) {
    return { valid: false, error: `File exceeds ${FILE_LIMITS.maxFileSize / 1024 / 1024}MB limit` };
  }
  if (!FILE_LIMITS.acceptedTypes[file.type]) {
    return { valid: false, error: `Unsupported file type: ${file.type}` };
  }
  return { valid: true };
}

export function validateUploadBatch(files: File[]): ValidationResult {
  if (files.length > FILE_LIMITS.maxFilesPerUpload) {
    return { valid: false, error: `Cannot upload more than ${FILE_LIMITS.maxFilesPerUpload} files at once` };
  }
  const totalSize = files.reduce((sum, f) => sum + f.size, 0);
  if (totalSize > FILE_LIMITS.maxTotalUploadSize) {
    return { valid: false, error: `Total upload size exceeds ${FILE_LIMITS.maxTotalUploadSize / 1024 / 1024}MB` };
  }
  for (const file of files) {
    const result = validateFile(file);
    if (!result.valid) return { valid: false, error: `${file.name}: ${result.error}` };
  }
  return { valid: true };
}
```

---

## Upload Progress Tracking

### Progress Component

```
┌────────────────────────────────────────────────────────┐
│  Uploading 3 files...                          67%    │
│  ████████████████████████████░░░░░░░░░░░░░░░░░░░░░░  │
├────────────────────────────────────────────────────────┤
│  ✅ report.pdf         ████████████████████  100%     │
│  ⏳ docs.md            ████████████░░░░░░░░   58%     │
│  ⏳ code.py            ░░░░░░░░░░░░░░░░░░░░    0%     │
│                                                        │
│  Speed: 1.2 MB/s    ETA: 8s    Uploaded: 5.7MB/8.5MB │
└────────────────────────────────────────────────────────┘
```

### Progress Hook

```tsx
// hooks/knowledge/useDocumentUpload.ts
import { useState, useCallback, useRef } from 'react';
import { api } from '@/lib/api';

interface UploadProgress {
  fileId: string;
  fileName: string;
  percentage: number;
  status: 'pending' | 'uploading' | 'processing' | 'complete' | 'error';
  error?: string;
}

export function useDocumentUpload() {
  const [uploads, setUploads] = useState<UploadProgress[]>([]);
  const [overallProgress, setOverallProgress] = useState(0);
  const abortControllerRef = useRef<AbortController | null>(null);

  const uploadFiles = useCallback(async (files: File[], setId?: string, tags?: string[]) => {
    abortControllerRef.current = new AbortController();

    const initial = files.map((f) => ({
      fileId: crypto.randomUUID(),
      fileName: f.name,
      percentage: 0,
      status: 'pending' as const,
    }));
    setUploads(initial);

    for (let i = 0; i < files.length; i++) {
      const formData = new FormData();
      formData.append('file', files[i]);
      if (setId) formData.append('set_id', setId);
      if (tags) tags.forEach((t) => formData.append('tags', t));

      try {
        setUploads((prev) =>
          prev.map((u, idx) => (idx === i ? { ...u, status: 'uploading' } : u))
        );

        await api.post('/documents/upload', formData, {
          signal: abortControllerRef.current.signal,
          onUploadProgress: (event) => {
            const pct = Math.round((event.loaded * 100) / (event.total || 1));
            setUploads((prev) =>
              prev.map((u, idx) => (idx === i ? { ...u, percentage: pct } : u))
            );
            const totalUploaded = uploads
              .filter((u) => u.status === 'complete')
              .reduce((sum, u) => sum + 100, 0) + pct;
            setOverallProgress(Math.round(totalUploaded / files.length));
          },
        });

        setUploads((prev) =>
          prev.map((u, idx) =>
            idx === i ? { ...u, status: 'processing', percentage: 100 } : u
          )
        );
      } catch (err) {
        setUploads((prev) =>
          prev.map((u, idx) =>
            idx === i ? { ...u, status: 'error', error: String(err) } : u
          )
        );
      }
    }

    setOverallProgress(100);
  }, [uploads]);

  const cancel = useCallback(() => {
    abortControllerRef.current?.abort();
  }, []);

  return { uploads, overallProgress, uploadFiles, cancel };
}
```

---

## Processing Pipeline Visualization

```
┌──────────────────────────────────────────────────────────────────┐
│  Document Processing Pipeline — report.pdf                       │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ● Upload    ──── ● Extract   ──── ● Chunk    ──── ◐ Embed     │
│  ✅ Done         ✅ Done         ✅ Done         ⏳ 65%        │
│  0.2s            6.0s            4.0s            8.2s...        │
│                                                                  │
│  ════════════════════════════════════════════░░░░░░░░░░░░░░░░░  │
│  0%        25%        50%        75%        ▓▓▓▓░░░░░  100%    │
│                                                                  │
│  Current: Generating embeddings for chunk 27/42                  │
│  Model: text-embedding-3-small    Estimated: 12s remaining      │
└──────────────────────────────────────────────────────────────────┘
```

### Pipeline Component

```tsx
// components/knowledge/ProcessingPipeline.tsx
const STAGES = ['uploaded', 'extracted', 'chunked', 'embedded', 'ready'] as const;
const STAGE_LABELS: Record<string, string> = {
  uploaded: 'Upload', extracted: 'Extract', chunked: 'Chunk',
  embedded: 'Embed', ready: 'Ready',
};

interface ProcessingPipelineProps {
  documentId: string;
  currentStage: string;
  stageHistory: StageEvent[];
  progress: number;
}

export function ProcessingPipeline({ currentStage, stageHistory, progress }: ProcessingPipelineProps) {
  const currentIdx = STAGES.indexOf(currentStage as any);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        {STAGES.map((stage, idx) => {
          const isComplete = idx < currentIdx;
          const isCurrent = idx === currentIdx;
          const event = stageHistory.find((h) => h.stage === stage);

          return (
            <div key={stage} className="flex items-center">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center
                ${isComplete ? 'bg-green-500 text-white' : ''}
                ${isCurrent ? 'bg-primary text-primary-foreground animate-pulse' : ''}
                ${idx > currentIdx ? 'bg-muted text-muted-foreground' : ''}
              `}>
                {isComplete ? '✓' : idx + 1}
              </div>
              <span className="ml-2 text-sm font-medium">{STAGE_LABELS[stage]}</span>
              {event?.duration_ms && (
                <span className="ml-1 text-xs text-muted-foreground">
                  {(event.duration_ms / 1000).toFixed(1)}s
                </span>
              )}
              {idx < STAGES.length - 1 && (
                <div className={`w-12 h-0.5 mx-2
                  ${isComplete ? 'bg-green-500' : 'bg-muted'}`}
                />
              )}
            </div>
          );
        })}
      </div>
      <Progress value={progress} className="h-2" />
    </div>
  );
}
```

---

## Document Chunk Viewer

```
┌─────────────────────────────────────────────────────────────┐
│  Chunks for report.pdf (42 total)                    ◄ 1 ► │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Chunk 15 of 42                  [◀ Prev] [Next ▶]        │
│  Words: 312  |  Tokens: ~420  |  Embedding: ✅             │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Revenue increased by 15% compared to the previous   │   │
│  │ quarter, driven primarily by the expansion into     │   │
│  │ Asian markets. The enterprise segment showed the    │   │
│  │ strongest growth at 23%, while consumer segment     │   │
│  │ maintained steady 8% growth. Key product launches   │   │
│  │ in March contributed approximately $4.2M in new     │   │
│  │ revenue, with the AI-powered analytics suite being  │   │
│  │ the top-performing product line...                  │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  [Edit Chunk] [Delete Chunk] [Split Here]                  │
└─────────────────────────────────────────────────────────────┘
```

---

## Document Metadata Editor

```tsx
// components/knowledge/MetadataEditor.tsx
export function MetadataEditor({ document }: { document: Document }) {
  const { updateMetadata } = useDocumentMutations();
  const [form, setForm] = useState({
    name: document.name,
    description: document.description || '',
    tags: document.tags || [],
    category: document.category || 'general',
    embedding_model: document.embedding_model,
    chunk_size: document.chunk_size,
    chunk_overlap: document.chunk_overlap,
  });

  return (
    <form onSubmit={() => updateMetadata(document.id, form)} className="space-y-4">
      <div>
        <Label>Name</Label>
        <Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
      </div>
      <div>
        <Label>Description</Label>
        <Textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
      </div>
      <div>
        <Label>Tags</Label>
        <TagInput value={form.tags} onChange={(tags) => setForm({ ...form, tags })} />
      </div>
      <div>
        <Label>Category</Label>
        <Select value={form.category} onValueChange={(v) => setForm({ ...form, category: v })}>
          <SelectContent>
            <SelectItem value="general">General</SelectItem>
            <SelectItem value="technical">Technical</SelectItem>
            <SelectItem value="legal">Legal</SelectItem>
            <SelectItem value="finance">Finance</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div>
        <Label>Embedding Model</Label>
        <Select value={form.embedding_model} onValueChange={(v) => setForm({ ...form, embedding_model: v })}>
          <SelectContent>
            <SelectItem value="text-embedding-3-small">text-embedding-3-small</SelectItem>
            <SelectItem value="text-embedding-3-large">text-embedding-3-large</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label>Chunk Size (tokens)</Label>
          <Input type="number" value={form.chunk_size}
            onChange={(e) => setForm({ ...form, chunk_size: +e.target.value })} />
        </div>
        <div>
          <Label>Chunk Overlap (tokens)</Label>
          <Input type="number" value={form.chunk_overlap}
            onChange={(e) => setForm({ ...form, chunk_overlap: +e.target.value })} />
        </div>
      </div>
      <Button type="submit">Save Changes</Button>
    </form>
  );
}
```

---

## Bulk Operations

### Bulk Actions Toolbar

```
┌──────────────────────────────────────────────────────────────┐
│  5 documents selected    [Delete] [Tag] [Move to Set] [Reprocess]  │
│  [Clear Selection]                                              │
└──────────────────────────────────────────────────────────────┘
```

### Supported Bulk Operations

| Operation         | Confirmation | Description                        |
|-------------------|--------------|------------------------------------|
| Delete            | Yes          | Soft delete selected documents     |
| Add Tags          | No           | Add tags to selected documents     |
| Remove Tags       | Yes          | Remove tags from selected          |
| Move to Set       | No           | Assign to a document set           |
| Remove from Set   | Yes          | Unassign from current set          |
| Reprocess         | Yes          | Re-run processing pipeline         |
| Change Visibility | No           | Update access permissions          |
| Export Metadata    | No           | Download metadata as CSV/JSON      |

---

## API Integration

### Endpoints

| Method   | Endpoint                               | Description              |
|----------|----------------------------------------|--------------------------|
| `POST`   | `/api/v1/rag/documents`                | Upload a document        |
| `POST`    | `/api/v1/rag/documents`                | List documents           |
| `POST`    | `/api/v1/rag/documents/:id`            | Get document details     |
| `PATCH`  | `/api/v1/rag/documents/:id`            | Update metadata          |
| `DELETE` | `/api/v1/rag/documents/:id`            | Delete document          |
| `POST`   | `/api/v1/rag/documents/:id/reprocess`  | Reprocess document       |
| `POST`    | `/api/v1/rag/documents/:id/status`     | Get processing status    |
| `POST`    | `/api/v1/rag/documents/:id/chunks`     | Get document chunks      |
| `PATCH`  | `/api/v1/rag/documents/:id/chunks/:chunkId` | Edit chunk         |
| `POST`   | `/api/v1/rag/search`                   | Search documents         |
| `POST`   | `/api/v1/rag/documents/bulk`           | Bulk operations          |
| `POST`    | `/api/v1/document-sets`                | List document sets       |
| `POST`   | `/api/v1/document-sets`                | Create a document set    |
| `PATCH`  | `/api/v1/document-sets/:id`            | Update a document set    |
| `DELETE` | `/api/v1/document-sets/:id`            | Delete a document set    |
| `POST`   | `/api/v1/document-sets/:id/documents`  | Add docs to set          |
| `DELETE` | `/api/v1/document-sets/:id/documents/:docId` | Remove doc from set |

### Upload Request/Response

```typescript
// POST /api/v1/rag/documents
// Request: multipart/form-data
interface UploadRequest {
  file: File;
}

// Response
interface UploadResponse {
  document_id: string;
  status: 'processing';
}
```

### Search Request/Response

```typescript
// POST /api/v1/rag/search
interface SearchRequest {
  query: string;
  limit?: number;
}

interface SearchResponse {
  results: SearchResult[];
  total_latency_ms: number;
}

interface SearchResult {
  document_id: string;
  title: string;
  content: string;
  score: number;
  source: string;
  metadata: Record<string, unknown>;
}
```

---

## Hooks

### useDocuments

```typescript
// hooks/knowledge/useDocuments.ts
export function useDocuments(filters?: DocumentFilters) {
  return useQuery({
    queryKey: ['documents', filters],
    queryFn: () => api.post('/documents', filters),
    staleTime: 30_000,
  });
}
```

### useDocumentUpload

```typescript
// hooks/knowledge/useDocumentUpload.ts
export function useDocumentUpload() {
  // Returns: { uploads, overallProgress, uploadFiles, cancel }
  // See Upload Progress Tracking section above for implementation
}
```

### useDocumentSearch

```typescript
// hooks/knowledge/useDocumentSearch.ts
export function useDocumentSearch() {
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const search = useCallback(async (params: SearchRequest) => {
    setIsSearching(true);
    try {
      const response = await api.post('/documents/search', params);
      setResults(response.results);
    } finally {
      setIsSearching(false);
    }
  }, []);

  return { results, isSearching, search };
}
```

### useDocumentSets

```typescript
// hooks/knowledge/useDocumentSets.ts
export function useDocumentSets() {
  const queryClient = useQueryClient();

  const sets = useQuery({
    queryKey: ['document-sets'],
    queryFn: () => api.post('/document-sets', {}),
  });

  const createSet = useMutation({
    mutationFn: (data: CreateSetRequest) => api.post('/document-sets', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['document-sets'] }),
  });

  const updateSet = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateSetRequest }) =>
      api.patch(`/document-sets/${id}`, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['document-sets'] }),
  });

  const deleteSet = useMutation({
    mutationFn: (id: string) => api.delete(`/document-sets/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['document-sets'] }),
  });

  return { sets, createSet, updateSet, deleteSet };
}
```

### useDocumentMutations

```typescript
// hooks/knowledge/useDocumentMutations.ts
export function useDocumentMutations() {
  const queryClient = useQueryClient();

  const updateMetadata = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Document> }) =>
      api.patch(`/documents/${id}`, data),
    onSuccess: (_, { id }) =>
      queryClient.invalidateQueries({ queryKey: ['document', id] }),
  });

  const deleteDocument = useMutation({
    mutationFn: (id: string) => api.delete(`/documents/${id}`),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['documents'] }),
  });

  const reprocess = useMutation({
    mutationFn: (id: string) => api.post(`/documents/${id}/reprocess`),
    onSuccess: (_, id) =>
      queryClient.invalidateQueries({ queryKey: ['document', id] }),
  });

  return { updateMetadata, deleteDocument, reprocess };
}
```

---

## Stores

### Documents Store (Zustand)

```typescript
// stores/knowledge/documents.ts
import { create } from 'zustand';

interface DocumentsState {
  documents: Document[];
  selectedIds: Set<string>;
  currentDocument: Document | null;
  processingStatus: Map<string, ProcessingStatus>;
  filters: DocumentFilters;
  view: 'list' | 'grid';
  sortBy: string;
  sortDir: 'asc' | 'desc';

  // Actions
  setDocuments: (docs: Document[]) => void;
  selectDocument: (id: string, selected: boolean) => void;
  selectAll: (selected: boolean) => void;
  clearSelection: () => void;
  setCurrentDocument: (doc: Document | null) => void;
  updateProcessingStatus: (docId: string, status: ProcessingStatus) => void;
  setFilters: (filters: DocumentFilters) => void;
  setView: (view: 'list' | 'grid') => void;
  setSort: (sortBy: string, sortDir: 'asc' | 'desc') => void;
}

export const useDocumentsStore = create<DocumentsState>((set, get) => ({
  documents: [],
  selectedIds: new Set(),
  currentDocument: null,
  processingStatus: new Map(),
  filters: {},
  view: 'list',
  sortBy: 'created_at',
  sortDir: 'desc',

  setDocuments: (documents) => set({ documents }),

  selectDocument: (id, selected) => set((state) => {
    const next = new Set(state.selectedIds);
    if (selected) next.add(id);
    else next.delete(id);
    return { selectedIds: next };
  }),

  selectAll: (selected) => set((state) => ({
    selectedIds: selected
      ? new Set(state.documents.map((d) => d.id))
      : new Set(),
  })),

  clearSelection: () => set({ selectedIds: new Set() }),

  setCurrentDocument: (doc) => set({ currentDocument: doc }),

  updateProcessingStatus: (docId, status) => set((state) => {
    const next = new Map(state.processingStatus);
    next.set(docId, status);
    return { processingStatus: next };
  }),

  setFilters: (filters) => set({ filters }),
  setView: (view) => set({ view }),
  setSort: (sortBy, sortDir) => set({ sortBy, sortDir }),
}));
```

---

## Error Handling

### Error Types

| Error Category     | Message                              | Action                          |
|--------------------|--------------------------------------|---------------------------------|
| Upload failed      | "Failed to upload {name}"           | Retry button                    |
| Invalid file type  | "Unsupported file format"            | Show accepted types             |
| File too large     | "File exceeds 50MB limit"           | Suggest compression             |
| Processing failed  | "Processing failed at {stage}"      | Retry or view details           |
| Search timeout     | "Search timed out, try refining"    | Auto-retry with shorter query   |
| Permission denied  | "You don't have access"             | Show share request option       |
| Network error      | "Connection lost"                    | Auto-retry with backoff         |
| Quota exceeded     | "Storage quota reached"             | Show upgrade plan               |

### Error Boundary

```tsx
// components/knowledge/DocumentErrorBoundary.tsx
import { ErrorBoundary } from '@/components/ErrorBoundary';

export function DocumentErrorBoundary({ children }: { children: React.ReactNode }) {
  return (
    <ErrorBoundary
      fallback={({ error, reset }) => (
        <div className="p-6 text-center space-y-4">
          <AlertCircle className="w-12 h-12 text-destructive mx-auto" />
          <h3 className="text-lg font-semibold">Something went wrong</h3>
          <p className="text-muted-foreground">{error.message}</p>
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

## Accessibility

### Keyboard Navigation

| Key            | Context            | Action                              |
|----------------|--------------------|-------------------------------------|
| `Enter`        | Drop zone          | Open file browser                   |
| `Space`        | Drop zone          | Open file browser                   |
| `Tab`          | Document list      | Move to next row/cell               |
| `Shift+Tab`    | Document list      | Move to previous row/cell           |
| `Enter`        | Document row       | Open document detail                |
| `Delete`       | Document row (sel) | Trigger delete action               |
| `Ctrl+A`       | Document list      | Select all documents                |
| `Escape`       | Any modal          | Close modal / clear selection       |
| `↑` / `↓`     | Chunk viewer       | Navigate between chunks             |
| `←` / `→`     | Chunk viewer       | Previous / Next chunk               |

### Screen Reader Support

- All interactive elements have `aria-label` attributes
- Document status changes announced via `aria-live="polite"` regions
- Drop zone has `role="button"` with descriptive `aria-label`
- Table headers use `scope="col"` for proper association
- Progress indicators use `aria-valuenow`, `aria-valuemin`, `aria-valuemax`
- Modal dialogs trap focus and return focus on close

### Drag-and-Drop Alternatives

- Click-to-browse file input always available alongside drag-and-drop
- Keyboard-triggered file selection via Enter/Space on drop zone
- Upload via paste (Ctrl+V) with clipboard file support
- URL-based upload for remote documents

### Color and Contrast

- Status badges use text + icon (not color alone) to convey meaning
- Processing status includes icon indicators alongside color
- All text meets WCAG 2.1 AA contrast requirements (4.5:1 ratio)
- Focus indicators visible on all interactive elements
