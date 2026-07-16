# Knowledge Center

The Knowledge Center manages documents, document sets, semantic search, and the processing pipeline. It handles file uploads, chunk viewing, and integration with agents.

## Architecture Overview

```
+------------------------------------------------------------------+
|                    Knowledge Center                                |
+------------------------------------------------------------------+
|                                                                    |
|  +-------------+    +--------------+    +-----------------+       |
|  | Document    |--->| Document     |--->| Processing      |       |
|  | List        |    | Upload       |    | Pipeline        |       |
|  +-------------+    +--------------+    +-----------------+       |
|        |                                        |                  |
|        v                                        v                  |
|  +-------------+    +--------------+    +-----------------+       |
|  | Document    |    | Document     |    | Chunk           |       |
|  | Search      |    | Sets         |    | Viewer          |       |
|  +-------------+    +--------------+    +-----------------+       |
|                                                                    |
+------------------------------------------------------------------+
```

## Processing Pipeline

```
+----------+    +----------+    +----------+    +----------+    +----------+
| Uploaded |--->| Extracted|--->| Chunked  |--->| Embedded |--->|  Ready   |
|          |    |          |    |          |    |          |    |          |
| File on  |    | Text     |    | Split    |    | Vectors  |    | Search-  |
| server   |    | content  |    | into     |    | created  |    | able     |
|          |    | parsed   |    | chunks   |    |          |    |          |
+----------+    +----------+    +----------+    +----------+    +----------+
     ^                                                                |
     |  Error at any step -> status = "failed" with error message    |
     +----------------------------------------------------------------+
```

## Data Models

```swift
struct Document: Codable, Identifiable, Equatable {
    let id: String
    var name: String
    var description: String?
    var category: String?
    var tags: [String]
    let fileType: FileType
    let fileSize: Int64
    var status: ProcessingStatus
    var chunks: [DocumentChunk]?
    var chunkCount: Int
    let metadata: DocumentMetadata
    let processingLog: [ProcessingLogEntry]?
    let createdAt: Date
    let updatedAt: Date
    let uploadedBy: String?

    enum CodingKeys: String, CodingKey {
        case id, name, description, category, tags
        case fileType = "file_type"
        case fileSize = "file_size"
        case status, chunks
        case chunkCount = "chunk_count"
        case metadata
        case processingLog = "processing_log"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case uploadedBy = "uploaded_by"
    }
}

enum FileType: String, Codable, CaseIterable {
    case pdf, docx, html, markdown, code, txt, csv, json

    var iconName: String {
        switch self {
        case .pdf: return "doc.fill"
        case .docx: return "doc.text.fill"
        case .html: return "globe"
        case .markdown: return "doc.plaintext"
        case .code: return "chevron.left.forwardslash.chevron.right"
        case .txt: return "doc.plaintext"
        case .csv: return "tablecells"
        case .json: return "curlybraces"
        }
    }

    var allowedExtensions: [String] {
        switch self {
        case .pdf: return ["pdf"]
        case .docx: return ["docx", "doc"]
        case .html: return ["html", "htm"]
        case .markdown: return ["md", "markdown"]
        case .code: return ["swift", "py", "js", "ts", "go", "rs", "java", "kt", "c", "cpp", "h"]
        case .txt: return ["txt"]
        case .csv: return ["csv"]
        case .json: return ["json"]
        }
    }

    var maxFileSize: Int64 {
        switch self {
        case .pdf: return 50_000_000
        case .docx: return 25_000_000
        case .html: return 10_000_000
        case .markdown: return 5_000_000
        case .code: return 5_000_000
        case .txt: return 5_000_000
        case .csv: return 10_000_000
        case .json: return 10_000_000
        }
    }
}

enum ProcessingStatus: String, Codable {
    case uploaded, extracting, extracted, chunking, chunked
    case embedding, embedded, ready, failed

    var displayName: String {
        switch self {
        case .uploaded: return "Uploaded"
        case .extracting: return "Extracting..."
        case .extracted: return "Extracted"
        case .chunking: return "Chunking..."
        case .chunked: return "Chunked"
        case .embedding: return "Embedding..."
        case .embedded: return "Embedded"
        case .ready: return "Ready"
        case .failed: return "Failed"
        }
    }

    var progress: Double {
        switch self {
        case .uploaded: return 0.1
        case .extracting: return 0.2
        case .extracted: return 0.3
        case .chunking: return 0.4
        case .chunked: return 0.6
        case .embedding: return 0.7
        case .embedded: return 0.9
        case .ready: return 1.0
        case .failed: return 0
        }
    }

    var isActive: Bool {
        [.extracting, .chunking, .embedding].contains(self)
    }
}

struct DocumentMetadata: Codable, Equatable {
    let wordCount: Int?
    let pageCount: Int?
    let language: String?
    let author: String?
    let createdDate: Date?
    let modifiedDate: Date?
    let summary: String?

    enum CodingKeys: String, CodingKey {
        case wordCount = "word_count"
        case pageCount = "page_count"
        case language, author
        case createdDate = "created_date"
        case modifiedDate = "modified_date"
        case summary
    }
}

struct ProcessingLogEntry: Codable, Identifiable, Equatable {
    let id = UUID()
    let step: String
    let status: String
    let message: String?
    let timestamp: Date
    let duration: TimeInterval?
}

struct DocumentChunk: Codable, Identifiable, Equatable {
    let id: String
    let documentId: String
    let index: Int
    let content: String
    let tokenCount: Int
    let metadata: ChunkMetadata?

    enum CodingKeys: String, CodingKey {
        case id, index, content, metadata
        case documentId = "document_id"
        case tokenCount = "token_count"
    }
}

struct ChunkMetadata: Codable, Equatable {
    let page: Int?
    let section: String?
    let startOffset: Int?
    let endOffset: Int?

    enum CodingKeys: String, CodingKey {
        case page, section
        case startOffset = "start_offset"
        case endOffset = "end_offset"
    }
}

struct DocumentSet: Codable, Identifiable, Equatable {
    let id: String
    var name: String
    var description: String?
    var documentCount: Int
    var totalChunks: Int
    var documentIds: [String]
    let createdAt: Date
    var updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id, name, description
        case documentCount = "document_count"
        case totalChunks = "total_chunks"
        case documentIds = "document_ids"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct DocumentSearchQuery: Codable {
    let query: String
    let setSearch: Bool
    let setIds: [String]?
    let fileTypes: [FileType]?
    let limit: Int

    enum CodingKeys: String, CodingKey {
        case query
        case setSearch = "set_search"
        case setIds = "set_ids"
        case fileTypes = "file_types"
        case limit
    }
}

struct DocumentSearchResult: Codable, Identifiable {
    let id = UUID()
    let chunk: DocumentChunk
    let document: Document?
    let score: Double
    let highlights: [SearchHighlight]?

    enum CodingKeys: String, CodingKey {
        case chunk, document, score, highlights
    }
}

struct SearchHighlight: Codable {
    let field: String
    let snippet: String
}

struct UploadProgress: Identifiable {
    let id = UUID()
    let fileName: String
    let fileSize: Int64
    var progress: Double
    var status: UploadStatus
    var error: String?

    enum UploadStatus {
        case waiting, uploading, processing, complete, failed
    }
}

struct DocumentPermission: Codable, Identifiable {
    let id: String
    let documentId: String
    let userId: String
    let permissionLevel: PermissionLevel
    let grantedAt: Date
    let grantedBy: String

    enum CodingKeys: String, CodingKey {
        case id, userId, permissionLevel, grantedAt, grantedBy
        case documentId = "document_id"
    }
}

enum PermissionLevel: String, Codable, CaseIterable {
    case view, edit, admin
    var displayName: String { rawValue.capitalized }
}
```

## Document List ViewModel

```swift
import SwiftUI
import Combine

@MainActor
final class DocumentListViewModel: ObservableObject {
    @Published var documents: [Document] = []
    @Published var documentSets: [DocumentSet] = []
    @Published var isLoading = false
    @Published var error: KnowledgeError?
    @Published var searchText = ""
    @Published var filterType: FileType?
    @Published var filterStatus: ProcessingStatus?
    @Published var sortBy: SortOption = .newest

    enum SortOption: String, CaseIterable {
        case newest = "Newest", oldest = "Oldest", nameAZ = "Name A-Z"
        case largest = "Largest", chunks = "Most Chunks"
    }

    private let apiService: DocumentAPIService
    private let cacheService: DocumentCacheService

    var filteredDocuments: [Document] {
        var result = documents
        if !searchText.isEmpty {
            result = result.filter {
                $0.name.localizedCaseInsensitiveContains(searchText) ||
                ($0.description?.localizedCaseInsensitiveContains(searchText) ?? false) ||
                $0.tags.contains { $0.localizedCaseInsensitiveContains(searchText) }
            }
        }
        if let type = filterType { result = result.filter { $0.fileType == type } }
        if let status = filterStatus { result = result.filter { $0.status == status } }
        switch sortBy {
        case .newest: result.sort { $0.createdAt > $1.createdAt }
        case .oldest: result.sort { $0.createdAt < $1.createdAt }
        case .nameAZ: result.sort { $0.name < $1.name }
        case .largest: result.sort { $0.fileSize > $1.fileSize }
        case .chunks: result.sort { $0.chunkCount > $1.chunkCount }
        }
        return result
    }

    init(apiService: DocumentAPIService = .shared, cacheService: DocumentCacheService = .shared) {
        self.apiService = apiService
        self.cacheService = cacheService
    }

    func loadDocuments() async {
        isLoading = true
        error = nil
        if let cached = await cacheService.getCachedDocuments() {
            documents = cached
            isLoading = false
            await refreshDocuments()
            return
        }
        do {
            documents = try await apiService.fetchDocuments()
            documentSets = try await apiService.fetchDocumentSets()
            await cacheService.cacheDocuments(documents)
        } catch {
            self.error = KnowledgeError.from(error)
        }
        isLoading = false
    }

    func refreshDocuments() async {
        do {
            documents = try await apiService.fetchDocuments()
            documentSets = try await apiService.fetchDocumentSets()
            await cacheService.cacheDocuments(documents)
        } catch {
            self.error = KnowledgeError.from(error)
        }
    }

    func deleteDocument(_ document: Document) async {
        do {
            try await apiService.deleteDocument(id: document.id)
            documents.removeAll { $0.id == document.id }
            await cacheService.cacheDocuments(documents)
        } catch {
            self.error = KnowledgeError.from(error)
        }
    }

    func searchDocuments(query: String) async -> [DocumentSearchResult] {
        do {
            let searchQuery = DocumentSearchQuery(
                query: query, setSearch: false, setIds: nil,
                fileTypes: nil, limit: 20
            )
            return try await apiService.searchDocuments(query: searchQuery)
        } catch {
            self.error = KnowledgeError.from(error)
            return []
        }
    }
}

enum KnowledgeError: LocalizedError {
    case network(String), server(Int, String?), decoding, unauthorized
    case uploadFailed(String), invalidFileType(String)
    case fileSizeExceeded(String, Int64), processingFailed(String)
    case unknown(String)

    var errorDescription: String? {
        switch self {
        case .network(let m): return "Network error: \(m)"
        case .server(let c, let m): return "Server error (\(c)): \(m ?? "Unknown")"
        case .decoding: return "Failed to parse response"
        case .unauthorized: return "Session expired"
        case .uploadFailed(let m): return "Upload failed: \(m)"
        case .invalidFileType(let t): return "File type '\(t)' is not supported"
        case .fileSizeExceeded(let t, _): return "File type '\(t)' exceeds maximum size"
        case .processingFailed(let m): return "Processing failed: \(m)"
        case .unknown(let m): return m
        }
    }

    static func from(_ error: Error) -> KnowledgeError {
        if let apiError = error as? APIError {
            switch apiError {
            case .network(let e): return .network(e.localizedDescription)
            case .http(let c, let m): return .server(c, m)
            case .decoding: return .decoding
            case .unauthorized: return .unauthorized
            }
        }
        return .unknown(error.localizedDescription)
    }
}
```

## Document Upload ViewModel

```swift
@MainActor
final class DocumentUploadViewModel: ObservableObject {
    @Published var selectedFiles: [URL] = []
    @Published var uploadProgresses: [UploadProgress] = []
    @Published var isUploading = false
    @Published var uploadComplete = false
    @Published var error: KnowledgeError?

    var overallProgress: Double {
        guard !uploadProgresses.isEmpty else { return 0 }
        return uploadProgresses.map(\.progress).reduce(0, +) / Double(uploadProgresses.count)
    }

    var allComplete: Bool {
        uploadProgresses.allSatisfy { $0.status == .complete }
    }

    private let apiService: DocumentAPIService

    init(apiService: DocumentAPIService = .shared) {
        self.apiService = apiService
    }

    func validateFile(_ url: URL) -> KnowledgeError? {
        guard let ext = url.pathExtension.lowercased().flatMap(FileType.init) else {
            return .invalidFileType(url.pathExtension)
        }
        let resources = try? url.resourceValues(forKeys: [.fileSizeKey])
        if let fileSize = resources?.fileSize, Int64(fileSize) > ext.maxFileSize {
            return .fileSizeExceeded(ext.rawValue, ext.maxFileSize)
        }
        return nil
    }

    func addFile(_ url: URL) {
        if let validationError = validateFile(url) {
            error = validationError
            return
        }
        guard !selectedFiles.contains(url) else { return }
        selectedFiles.append(url)
    }

    func removeFile(_ url: URL) {
        selectedFiles.removeAll { $0 == url }
    }

    func uploadAll() async {
        guard !selectedFiles.isEmpty else { return }
        isUploading = true
        uploadComplete = false

        uploadProgresses = selectedFiles.map { url in
            UploadProgress(
                fileName: url.lastPathComponent,
                fileSize: (try? url.resourceValues(forKeys: [.fileSizeKey])?.fileSize).map(Int64.init) ?? 0,
                progress: 0,
                status: .waiting
            )
        }

        for (index, url) in selectedFiles.enumerated() {
            uploadProgresses[index].status = .uploading
            do {
                let document = try await apiService.uploadDocument(
                    fileURL: url,
                    progressHandler: { [weak self] progress in
                        Task { @MainActor in
                            self?.uploadProgresses[index].progress = progress
                        }
                    }
                )
                uploadProgresses[index].status = .processing
                uploadProgresses[index].progress = 1.0
                _ = document
            } catch {
                uploadProgresses[index].status = .failed
                uploadProgresses[index].error = error.localizedDescription
            }
        }

        isUploading = false
        uploadComplete = allComplete
    }
}

extension String {
    init?(fileExtension ext: String) {
        guard FileType.allCases.contains(where: { $0.allowedExtensions.contains(ext) }) else {
            return nil
        }
        self = ext
    }
}
```

## Document List Screen

```swift
struct DocumentListView: View {
    @StateObject private var viewModel = DocumentListViewModel()
    @State private var showUpload = false
    @State private var documentToDelete: Document?

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.documents.isEmpty {
                    DocumentListLoadingView()
                } else if let error = viewModel.error, viewModel.documents.isEmpty {
                    DocumentListErrorView(error: error) {
                        Task { await viewModel.loadDocuments() }
                    }
                } else if viewModel.filteredDocuments.isEmpty && viewModel.searchText.isEmpty {
                    DocumentListEmptyView { showUpload = true }
                } else {
                    documentList
                }
            }
            .navigationTitle("Knowledge Center")
            .searchable(text: $viewModel.searchText, prompt: "Search documents...")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Menu {
                        sortMenu
                        filterMenu
                    } label: {
                        Image(systemName: "line.3.horizontal.decrease.circle")
                    }
                }
                ToolbarItem(placement: .topBarTrailing) {
                    Button(action: { showUpload = true }) {
                        Image(systemName: "arrow.up.doc")
                    }
                }
            }
            .sheet(isPresented: $showUpload) {
                DocumentUploadSheet(viewModel: viewModel)
            }
            .refreshable { await viewModel.refreshDocuments() }
            .task { await viewModel.loadDocuments() }
        }
    }

    private var documentList: some View {
        List {
            Section {
                ForEach(viewModel.filteredDocuments) { document in
                    NavigationLink(destination: DocumentDetailView(document: document)) {
                        DocumentRowView(document: document)
                    }
                    .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                        Button(role: .destructive) { documentToDelete = document } label: {
                            Label("Delete", systemImage: "trash")
                        }
                    }
                }
            } header: {
                HStack {
                    Text("\(viewModel.filteredDocuments.count) documents")
                    Spacer()
                    Text("Sorted by \(viewModel.sortBy.rawValue)")
                        .font(.caption)
                }
            }
        }
        .listStyle(.insetGrouped)
    }

    private var sortMenu: some View {
        Menu("Sort") {
            ForEach(DocumentListViewModel.SortOption.allCases, id: \.self) { option in
                Button {
                    viewModel.sortBy = option
                } label: {
                    HStack {
                        Text(option.rawValue)
                        if viewModel.sortBy == option { Image(systemName: "checkmark") }
                    }
                }
            }
        }
    }

    private var filterMenu: some View {
        Menu("Filter") {
            Button("All Types") { viewModel.filterType = nil }
            ForEach(FileType.allCases, id: \.self) { type in
                Button(type.rawValue.uppercased()) { viewModel.filterType = type }
            }
            Divider()
            Button("All Statuses") { viewModel.filterStatus = nil }
            ForEach([ProcessingStatus.ready, .failed], id: \.self) { status in
                Button(status.displayName) { viewModel.filterStatus = status }
            }
        }
    }
}

struct DocumentRowView: View {
    let document: Document

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: document.fileType.iconName)
                .font(.title2)
                .foregroundStyle(colorForType(document.fileType))
                .frame(width: 32)

            VStack(alignment: .leading, spacing: 4) {
                Text(document.name)
                    .font(.subheadline.bold())
                HStack(spacing: 8) {
                    Text(document.fileType.rawValue.uppercased())
                    Text(formatBytes(document.fileSize))
                    Text("\(document.chunkCount) chunks")
                }
                .font(.caption)
                .foregroundStyle(.secondary)
            }

            Spacer()

            statusBadge
        }
    }

    private var statusBadge: some View {
        Text(document.status.displayName)
            .font(.caption2.bold())
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(statusColor.opacity(0.15))
            .foregroundStyle(statusColor)
            .clipShape(Capsule())
    }

    private var statusColor: Color {
        switch document.status {
        case .ready: return .green
        case .failed: return .red
        case .uploaded, .extracting, .extracted, .chunking, .chunked, .embedding, .embedded:
            return .orange
        }
    }

    private func colorForType(_ type: FileType) -> Color {
        switch type {
        case .pdf: return .red
        case .docx: return .blue
        case .html: return .cyan
        case .markdown: return .purple
        case .code: return .green
        case .txt: return .gray
        case .csv: return .orange
        case .json: return .yellow
        }
    }

    private func formatBytes(_ bytes: Int64) -> String {
        let formatter = ByteCountFormatter()
        return formatter.string(fromByteCount: bytes)
    }
}
```

## Document Upload Screen

```swift
struct DocumentUploadSheet: View {
    @Environment(\.dismiss) private var dismiss
    @ObservedObject var viewModel: DocumentListViewModel
    @StateObject private var uploadVM = DocumentUploadViewModel()
    @State private var showFilePicker = false
    @State private var showCamera = false
    @State private var showGallery = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 16) {
                if uploadVM.isUploading {
                    uploadProgressView
                } else {
                    fileSelectionView
                }
            }
            .padding()
            .navigationTitle("Upload Documents")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                if !uploadVM.selectedFiles.isEmpty && !uploadVM.isUploading {
                    ToolbarItem(placement: .confirmationAction) {
                        Button("Upload") { Task { await uploadVM.uploadAll() } }
                    }
                }
            }
            .fileImporter(
                isPresented: $showFilePicker,
                allowedContentTypes: FileUploader.allowedTypes,
                allowsMultipleSelection: true
            ) { result in
                handleFileImport(result)
            }
            .sheet(isPresented: $showCamera) {
                CameraCaptureView { url in uploadVM.addFile(url) }
            }
            .onChange(of: uploadVM.uploadComplete) { complete in
                if complete { dismiss() }
            }
        }
    }

    private var fileSelectionView: some View {
        VStack(spacing: 20) {
            if uploadVM.selectedFiles.isEmpty {
                VStack(spacing: 16) {
                    Image(systemName: "arrow.up.doc.on.clipboard")
                        .font(.system(size: 48))
                        .foregroundStyle(.secondary)
                    Text("Select files to upload")
                        .font(.headline)
                    Text("Supported: PDF, DOCX, HTML, Markdown, Code, TXT, CSV, JSON")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }
            }

            HStack(spacing: 16) {
                uploadActionButton(icon: "folder", title: "Files", color: .blue) {
                    showFilePicker = true
                }
                uploadActionButton(icon: "camera", title: "Camera", color: .green) {
                    showCamera = true
                }
            }
            .padding(.top, uploadVM.selectedFiles.isEmpty ? 20 : 0)

            if !uploadVM.selectedFiles.isEmpty {
                fileList
            }
        }
    }

    private var fileList: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Selected Files (\(uploadVM.selectedFiles.count))")
                .font(.headline)
            ForEach(uploadVM.selectedFiles, id: \.self) { url in
                HStack {
                    Image(systemName: fileIcon(for: url))
                    Text(url.lastPathComponent)
                        .font(.subheadline)
                    Spacer()
                    Button(action: { uploadVM.removeFile(url) }) {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundStyle(.red)
                    }
                }
                .padding(.vertical, 4)
            }
        }
    }

    private var uploadProgressView: some View {
        VStack(spacing: 20) {
            ProgressView(value: uploadVM.overallProgress) {
                Text("Uploading... \(Int(uploadVM.overallProgress * 100))%")
                    .font(.headline)
            }
            .progressViewStyle(.linear)

            List(uploadVM.uploadProgresses) { progress in
                HStack {
                    VStack(alignment: .leading) {
                        Text(progress.fileName)
                            .font(.subheadline)
                        if let error = progress.error {
                            Text(error).font(.caption).foregroundStyle(.red)
                        }
                    }
                    Spacer()
                    uploadStatusIcon(progress.status)
                }
            }
        }
    }

    private func uploadActionButton(icon: String, title: String, color: Color, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            VStack(spacing: 8) {
                Image(systemName: icon)
                    .font(.title2)
                Text(title)
                    .font(.caption.bold())
            }
            .frame(maxWidth: .infinity)
            .padding()
            .background(color.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
        .buttonStyle(.plain)
    }

    private func uploadStatusIcon(_ status: UploadProgress.UploadStatus) -> some View {
        Group {
            switch status {
            case .waiting: Image(systemName: "clock").foregroundStyle(.gray)
            case .uploading: ProgressView().scaleEffect(0.7)
            case .processing: ProgressView().scaleEffect(0.7)
            case .complete: Image(systemName: "checkmark.circle.fill").foregroundStyle(.green)
            case .failed: Image(systemName: "xmark.circle.fill").foregroundStyle(.red)
            }
        }
    }

    private func fileIcon(for url: URL) -> String {
        guard let ext = url.pathExtension.lowercased().flatMap(FileType.init) else {
            return "doc"
        }
        return ext.iconName
    }

    private func handleFileImport(_ result: Result<[URL], Error>) {
        switch result {
        case .success(let urls):
            urls.forEach { uploadVM.addFile($0) }
        case .failure(let error):
            uploadVM.error = KnowledgeError.from(error)
        }
    }
}
```

## Document Detail Screen

```swift
struct DocumentDetailView: View {
    let document: Document
    @State private var selectedTab: DetailTab = .metadata
    @State private var showMetadataEditor = false
    @State private var showDeleteConfirmation = false
    @StateObject private var viewModel: DocumentDetailViewModel

    init(document: Document) {
        self.document = document
        _viewModel = StateObject(wrappedValue: DocumentDetailViewModel(documentId: document.id))
    }

    enum DetailTab: String, CaseIterable {
        case metadata = "Info"
        case chunks = "Chunks"
        case processing = "Log"
        case permissions = "Access"
    }

    var body: some View {
        VStack(spacing: 0) {
            Picker("Tab", selection: $selectedTab) {
                ForEach(DetailTab.allCases, id: \.self) { Text($0.rawValue).tag($0) }
            }
            .pickerStyle(.segmented)
            .padding(.horizontal)

            switch selectedTab {
            case .metadata: metadataTab
            case .chunks: chunksTab
            case .processing: processingLogTab
            case .permissions: permissionsTab
            }
        }
        .navigationTitle(document.name)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Menu {
                    Button("Edit Metadata") { showMetadataEditor = true }
                    Button("Delete", role: .destructive) { showDeleteConfirmation = true }
                } label: {
                    Image(systemName: "ellipsis.circle")
                }
            }
        }
        .sheet(isPresented: $showMetadataEditor) {
            DocumentMetadataEditorSheet(document: document)
        }
        .alert("Delete Document", isPresented: $showDeleteConfirmation) {
            Button("Delete", role: .destructive) { Task { await viewModel.deleteDocument() } }
            Button("Cancel", role: .cancel) { }
        } message: {
            Text("This will permanently delete this document and all its chunks.")
        }
    }

    private var metadataTab: some View {
        ScrollView {
            VStack(spacing: 16) {
                infoCard("Status", content: documentStatusView)
                infoCard("Details", content: detailsView)
                infoCard("Metadata", content: metadataView)
                if let desc = document.description {
                    infoCard("Description", content: Text(desc).font(.subheadline))
                }
            }
            .padding()
        }
    }

    private var documentStatusView: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Circle()
                    .fill(document.status == .ready ? .green : .orange)
                    .frame(width: 8, height: 8)
                Text(document.status.displayName)
                    .font(.subheadline.bold())
            }
            if document.status.isActive {
                ProgressView(value: document.status.progress)
            }
        }
    }

    private var detailsView: some View {
        VStack(spacing: 6) {
            LabeledContent("Type", value: document.fileType.rawValue.uppercased())
            LabeledContent("Size", value: formatBytes(document.fileSize))
            LabeledContent("Chunks", value: "\(document.chunkCount)")
            LabeledContent("Created", value: document.createdAt.formatted())
        }
    }

    private var metadataView: some View {
        VStack(spacing: 6) {
            if let words = document.metadata.wordCount {
                LabeledContent("Words", value: "\(words)")
            }
            if let pages = document.metadata.pageCount {
                LabeledContent("Pages", value: "\(pages)")
            }
            if let lang = document.metadata.language {
                LabeledContent("Language", value: lang)
            }
            if let author = document.metadata.author {
                LabeledContent("Author", value: author)
            }
            if let summary = document.metadata.summary {
                Divider()
                Text(summary).font(.caption).foregroundStyle(.secondary)
            }
        }
    }

    private var chunksTab: some View {
        Group {
            if let chunks = viewModel.chunks, !chunks.isEmpty {
                List(chunks) { chunk in
                    ChunkRowView(chunk: chunk)
                }
                .listStyle(.insetGrouped)
            } else {
                VStack(spacing: 12) {
                    ProgressView()
                    Text("Loading chunks...")
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .task { await viewModel.loadChunks() }
    }

    private var processingLogTab: some View {
        Group {
            if let log = document.processingLog, !log.isEmpty {
                List(log) { entry in
                    ProcessingLogRow(entry: entry)
                }
                .listStyle(.insetGrouped)
            } else {
                Text("No processing log available")
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
    }

    private var permissionsTab: some View {
        Group {
            if let perms = viewModel.permissions, !perms.isEmpty {
                List(perms) { perm in
                    HStack {
                        VStack(alignment: .leading) {
                            Text(perm.userId).font(.subheadline.bold())
                            Text("Granted \(perm.grantedAt.formatted(.relative(presentation: .named)))")
                                .font(.caption).foregroundStyle(.secondary)
                        }
                        Spacer()
                        Text(perm.permissionLevel.displayName)
                            .font(.caption.bold())
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(.blue.opacity(0.1))
                            .clipShape(Capsule())
                    }
                }
                .listStyle(.insetGrouped)
            } else {
                VStack(spacing: 12) {
                    Image(systemName: "lock.open")
                        .font(.largeTitle)
                        .foregroundStyle(.secondary)
                    Text("No permissions set")
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .task { await viewModel.loadPermissions() }
    }

    private func infoCard<Content: View>(_ title: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title).font(.headline)
            content()
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func formatBytes(_ bytes: Int64) -> String {
        ByteCountFormatter().string(fromByteCount: bytes)
    }
}

struct ChunkRowView: View {
    let chunk: DocumentChunk

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Text("Chunk \(chunk.index + 1)")
                    .font(.caption.bold())
                Spacer()
                Text("\(chunk.tokenCount) tokens")
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            }
            Text(chunk.content)
                .font(.caption)
                .lineLimit(4)
            if let meta = chunk.metadata {
                HStack(spacing: 8) {
                    if let page = meta.page { Text("p.\(page)") }
                    if let section = meta.section { Text(section) }
                }
                .font(.caption2)
                .foregroundStyle(.tertiary)
            }
        }
        .padding(.vertical, 4)
    }
}

struct ProcessingLogRow: View {
    let entry: ProcessingLogEntry

    var body: some View {
        HStack(spacing: 12) {
            Circle()
                .fill(entry.status == "success" ? .green : .red)
                .frame(width: 8, height: 8)
            VStack(alignment: .leading, spacing: 2) {
                Text(entry.step).font(.subheadline.bold())
                if let msg = entry.message {
                    Text(msg).font(.caption).foregroundStyle(.secondary)
                }
            }
            Spacer()
            VStack(alignment: .trailing) {
                Text(entry.timestamp.formatted(date: .omitted, time: .shortened))
                    .font(.caption2)
                if let dur = entry.duration {
                    Text(String(format: "%.1fs", dur))
                        .font(.caption2).foregroundStyle(.secondary)
                }
            }
        }
    }
}

@MainActor
final class DocumentDetailViewModel: ObservableObject {
    @Published var chunks: [DocumentChunk]?
    @Published var permissions: [DocumentPermission]?
    @Published var isLoading = false

    let documentId: String
    private let apiService: DocumentAPIService

    init(documentId: String, apiService: DocumentAPIService = .shared) {
        self.documentId = documentId
        self.apiService = apiService
    }

    func loadChunks() async {
        do { chunks = try await apiService.fetchChunks(documentId: documentId) }
        catch { }
    }

    func loadPermissions() async {
        do { permissions = try await apiService.fetchPermissions(documentId: documentId) }
        catch { }
    }

    func deleteDocument() async {
        do { try await apiService.deleteDocument(id: documentId) }
        catch { }
    }
}
```

## Document Metadata Editor

```swift
struct DocumentMetadataEditorSheet: View {
    @Environment(\.dismiss) private var dismiss
    let document: Document
    @State private var name: String
    @State private var description: String
    @State private var category: String
    @State private var tags: [String]
    @State private var newTag = ""

    init(document: Document) {
        self.document = document
        _name = State(initialValue: document.name)
        _description = State(initialValue: document.description ?? "")
        _category = State(initialValue: document.category ?? "")
        _tags = State(initialValue: document.tags)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("Name") { TextField("Name", text: $name) }
                Section("Description") {
                    TextEditor(text: $description).frame(minHeight: 80)
                }
                Section("Category") { TextField("Category", text: $category) }
                Section("Tags") {
                    ForEach(tags, id: \.self) { tag in
                        HStack {
                            Text(tag)
                            Spacer()
                            Button(action: { tags.removeAll { $0 == tag } }) {
                                Image(systemName: "minus.circle.fill").foregroundStyle(.red)
                            }
                        }
                    }
                    HStack {
                        TextField("Add tag", text: $newTag)
                        Button(action: {
                            guard !newTag.isEmpty else { return }
                            tags.append(newTag)
                            newTag = ""
                        }) {
                            Image(systemName: "plus.circle.fill")
                        }
                        .disabled(newTag.isEmpty)
                    }
                }
            }
            .navigationTitle("Edit Metadata")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        Task { await save() }
                    }
                    .disabled(name.trimmingCharacters(in: .whitespaces).isEmpty)
                }
            }
        }
    }

    private func save() async {
        dismiss()
    }
}
```

## Document Search

```swift
struct DocumentSearchView: View {
    @StateObject private var viewModel = DocumentSearchViewModel()
    @State private var searchText = ""
    @State private var searchScope: SearchScope = .all
    @State private var selectedSets: [String] = []

    enum SearchScope: String, CaseIterable {
        case all = "All", semantic = "Semantic", keyword = "Keyword"
    }

    var body: some View {
        VStack(spacing: 0) {
            Picker("Scope", selection: $searchScope) {
                ForEach(SearchScope.allCases, id: \.self) { Text($0.rawValue).tag($0) }
            }
            .pickerStyle(.segmented)
            .padding()

            if viewModel.isSearching {
                ProgressView("Searching...")
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if let results = viewModel.results, results.isEmpty {
                Text("No results found")
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if let results = viewModel.results {
                List(results) { result in
                    SearchResultRow(result: result)
                }
                .listStyle(.plain)
            }
        }
        .searchable(text: $searchText, prompt: "Search documents...")
        .onSubmit(of: .search) {
            Task { await viewModel.search(query: searchText, scope: searchScope, setIds: selectedSets) }
        }
    }
}

struct SearchResultRow: View {
    let result: DocumentSearchResult

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                if let doc = result.document {
                    Text(doc.name)
                        .font(.caption.bold())
                        .foregroundStyle(.blue)
                }
                Spacer()
                Text(String(format: "%.0f%%", result.score * 100))
                    .font(.caption2.bold())
                    .foregroundStyle(.green)
            }
            Text(result.chunk.content)
                .font(.subheadline)
                .lineLimit(4)
            if let highlights = result.highlights {
                ForEach(highlights, id: \.field) { highlight in
                    Text(highlight.snippet)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .padding(4)
                        .background(.yellow.opacity(0.2))
                        .clipShape(RoundedRectangle(cornerRadius: 4))
                }
            }
        }
        .padding(.vertical, 8)
    }
}

@MainActor
final class DocumentSearchViewModel: ObservableObject {
    @Published var results: [DocumentSearchResult]?
    @Published var isSearching = false

    private let apiService: DocumentAPIService

    init(apiService: DocumentAPIService = .shared) {
        self.apiService = apiService
    }

    func search(query: String, scope: DocumentSearchView.SearchScope, setIds: [String]) async {
        guard !query.trimmingCharacters(in: .whitespaces).isEmpty else { return }
        isSearching = true
        do {
            let searchQuery = DocumentSearchQuery(
                query: query,
                setSearch: !setIds.isEmpty,
                setIds: setIds.isEmpty ? nil : setIds,
                fileTypes: nil,
                limit: 20
            )
            results = try await apiService.searchDocuments(query: searchQuery)
        } catch { results = [] }
        isSearching = false
    }
}
```

## Document Sets Management

```swift
struct DocumentSetsView: View {
    @StateObject private var viewModel = DocumentSetsViewModel()
    @State private var showCreateSet = false
    @State private var setToDelete: DocumentSet?

    var body: some View {
        List {
            ForEach(viewModel.sets) { set in
                NavigationLink(destination: DocumentSetDetailView(documentSet: set)) {
                    VStack(alignment: .leading, spacing: 4) {
                        Text(set.name).font(.headline)
                        Text(set.description ?? "No description")
                            .font(.caption).foregroundStyle(.secondary)
                        HStack(spacing: 12) {
                            Label("\(set.documentCount) docs", systemImage: "doc")
                            Label("\(set.totalChunks) chunks", systemImage: "text.word.spacing")
                        }
                        .font(.caption2).foregroundStyle(.secondary)
                    }
                }
                .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                    Button(role: .destructive) { setToDelete = set } label: {
                        Label("Delete", systemImage: "trash")
                    }
                }
            }
        }
        .listStyle(.insetGrouped)
        .navigationTitle("Document Sets")
        .toolbar {
            Button(action: { showCreateSet = true }) { Image(systemName: "plus") }
        }
        .sheet(isPresented: $showCreateSet) { CreateDocumentSetSheet(viewModel: viewModel) }
        .refreshable { await viewModel.loadSets() }
        .task { await viewModel.loadSets() }
    }
}

struct DocumentSetDetailView: View {
    let documentSet: DocumentSet
    @StateObject private var viewModel: DocumentSetDetailViewModel

    init(documentSet: DocumentSet) {
        self.documentSet = documentSet
        _viewModel = StateObject(wrappedValue: DocumentSetDetailViewModel(setId: documentSet.id))
    }

    var body: some View {
        Group {
            if viewModel.isLoading {
                ProgressView("Loading...")
            } else if let docs = viewModel.documents, !docs.isEmpty {
                List(docs) { doc in
                    NavigationLink(destination: DocumentDetailView(document: doc)) {
                        DocumentRowView(document: doc)
                    }
                }
                .listStyle(.insetGrouped)
            } else {
                VStack(spacing: 12) {
                    Image(systemName: "doc.on.doc").font(.largeTitle).foregroundStyle(.secondary)
                    Text("No documents in this set").foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            }
        }
        .navigationTitle(documentSet.name)
        .task { await viewModel.loadDocuments() }
    }
}

struct CreateDocumentSetSheet: View {
    @Environment(\.dismiss) private var dismiss
    @ObservedObject var viewModel: DocumentSetsViewModel
    @State private var name = ""
    @State private var description = ""

    var body: some View {
        NavigationStack {
            Form {
                Section("Set Details") {
                    TextField("Name", text: $name)
                    TextField("Description", text: $description)
                }
                Section {
                    Button("Create Set") {
                        Task {
                            await viewModel.createSet(name: name, description: description)
                            dismiss()
                        }
                    }
                    .disabled(name.trimmingCharacters(in: .whitespaces).isEmpty)
                    .frame(maxWidth: .infinity)
                }
            }
            .navigationTitle("New Document Set")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
    }
}

@MainActor
final class DocumentSetsViewModel: ObservableObject {
    @Published var sets: [DocumentSet] = []
    @Published var isLoading = false

    private let apiService: DocumentAPIService

    init(apiService: DocumentAPIService = .shared) {
        self.apiService = apiService
    }

    func loadSets() async {
        isLoading = true
        do { sets = try await apiService.fetchDocumentSets() }
        catch { }
        isLoading = false
    }

    func createSet(name: String, description: String) async {
        do {
            let set = try await apiService.createDocumentSet(
                name: name, description: description.isEmpty ? nil : description
            )
            sets.append(set)
        } catch { }
    }

    func deleteSet(_ set: DocumentSet) async {
        do {
            try await apiService.deleteDocumentSet(id: set.id)
            sets.removeAll { $0.id == set.id }
        } catch { }
    }
}

@MainActor
final class DocumentSetDetailViewModel: ObservableObject {
    @Published var documents: [Document]?
    @Published var isLoading = false

    let setId: String
    private let apiService: DocumentAPIService

    init(setId: String, apiService: DocumentAPIService = .shared) {
        self.setId = setId
        self.apiService = apiService
    }

    func loadDocuments() async {
        isLoading = true
        do { documents = try await apiService.fetchDocumentsInSet(setId: setId) }
        catch { documents = [] }
        isLoading = false
    }
}
```

## Camera Capture (AVFoundation)

```swift
struct CameraCaptureView: UIViewControllerRepresentable {
    let onCapture: (URL) -> Void
    @Environment(\.dismiss) private var dismiss

    func makeUIViewController(context: Context) -> UIImagePickerController {
        let picker = UIImagePickerController()
        picker.sourceType = .camera
        picker.delegate = context.coordinator
        return picker
    }

    func updateUIViewController(_ uiViewController: UIImagePickerController, context: Context) {}

    func makeCoordinator() -> Coordinator { Coordinator(self) }

    class Coordinator: NSObject, UIImagePickerControllerDelegate, UINavigationControllerDelegate {
        let parent: CameraCaptureView

        init(_ parent: CameraCaptureView) { self.parent = parent }

        func imagePickerController(_ picker: UIImagePickerController,
                                    didFinishPickingMediaWithInfo info: [UIImagePickerController.InfoKey: Any]) {
            if let image = info[.originalImage] as? UIImage,
               let data = image.jpegData(compressionQuality: 0.8) {
                let tempURL = FileManager.default.temporaryDirectory
                    .appendingPathComponent("\(UUID().uuidString).jpg")
                try? data.write(to: tempURL)
                parent.onCapture(tempURL)
            }
            parent.dismiss()
        }

        func imagePickerControllerDidCancel(_ picker: UIImagePickerController) {
            parent.dismiss()
        }
    }
}
```

## Supported File Types Reference

| Type | Extensions | Max Size | Processing |
|---|---|---|---|
| PDF | `.pdf` | 50 MB | Text extraction via PDFKit |
| DOCX | `.docx`, `.doc` | 25 MB | XML parsing |
| HTML | `.html`, `.htm` | 10 MB | DOM text extraction |
| Markdown | `.md`, `.markdown` | 5 MB | Native parsing |
| Code | `.swift`, `.py`, `.js`, `.ts`, `.go`, `.rs`, `.java`, `.kt` | 5 MB | Plain text + syntax metadata |
| TXT | `.txt` | 5 MB | Direct read |
| CSV | `.csv` | 10 MB | Row-based chunking |
| JSON | `.json` | 10 MB | Structured parsing |

## API Integration

```swift
protocol DocumentAPIService {
    func fetchDocuments() async throws -> [Document]
    func fetchDocument(id: String) async throws -> Document
    func uploadDocument(fileURL: URL, progressHandler: @escaping (Double) -> Void) async throws -> Document
    func deleteDocument(id: String) async throws
    func searchDocuments(query: DocumentSearchQuery) async throws -> [DocumentSearchResult]
    func fetchChunks(documentId: String) async throws -> [DocumentChunk]
    func fetchDocumentSets() async throws -> [DocumentSet]
    func createDocumentSet(name: String, description: String?) async throws -> DocumentSet
    func deleteDocumentSet(id: String) async throws
    func fetchDocumentsInSet(setId: String) async throws -> [Document]
    func fetchPermissions(documentId: String) async throws -> [DocumentPermission]
}

final class DefaultDocumentAPIService: DocumentAPIService {
    static let shared = DefaultDocumentAPIService()
    private let client: NetworkClient

    init(client: NetworkClient = .shared) { self.client = client }

    func fetchDocuments() async throws -> [Document] {
        let req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/documents")!, method: .get)
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode([Document].self, from: data)
    }

    func fetchDocument(id: String) async throws -> Document {
        let req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/documents/\(id)")!, method: .get)
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode(Document.self, from: data)
    }

    func uploadDocument(fileURL: URL, progressHandler: @escaping (Double) -> Void) async throws -> Document {
        var req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/documents/upload")!, method: .post)
        let boundary = "Boundary-\(UUID().uuidString)"
        req.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")
        let fileData = try Data(contentsOf: fileURL)
        var body = Data()
        body.append("--\(boundary)\r\n".data(using: .utf8)!)
        body.append("Content-Disposition: form-data; name=\"file\"; filename=\"\(fileURL.lastPathComponent)\"\r\n".data(using: .utf8)!)
        body.append("Content-Type: application/octet-stream\r\n\r\n".data(using: .utf8)!)
        body.append(fileData)
        body.append("\r\n--\(boundary)--\r\n".data(using: .utf8)!)
        req.httpBody = body
        let (data, _) = try await client.executeWithProgress(req, progressHandler: progressHandler)
        return try JSONDecoder().decode(Document.self, from: data)
    }

    func deleteDocument(id: String) async throws {
        let req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/documents/\(id)")!, method: .delete)
        _ = try await client.execute(req)
    }

    func searchDocuments(query: DocumentSearchQuery) async throws -> [DocumentSearchResult] {
        var req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/documents/search")!, method: .post)
        req.httpBody = try JSONEncoder().encode(query)
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode([DocumentSearchResult].self, from: data)
    }

    func fetchChunks(documentId: String) async throws -> [DocumentChunk] {
        let req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/documents/\(documentId)/chunks")!, method: .get)
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode([DocumentChunk].self, from: data)
    }

    func fetchDocumentSets() async throws -> [DocumentSet] {
        let req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/document-sets")!, method: .get)
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode([DocumentSet].self, from: data)
    }

    func createDocumentSet(name: String, description: String?) async throws -> DocumentSet {
        var req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/document-sets")!, method: .post)
        let body: [String: Any] = ["name": name, "description": description ?? ""]
        req.httpBody = try JSONSerialization.data(withJSONObject: body)
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode(DocumentSet.self, from: data)
    }

    func deleteDocumentSet(id: String) async throws {
        let req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/document-sets/\(id)")!, method: .delete)
        _ = try await client.execute(req)
    }

    func fetchDocumentsInSet(setId: String) async throws -> [Document] {
        let req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/document-sets/\(setId)/documents")!, method: .get)
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode([Document].self, from: data)
    }

    func fetchPermissions(documentId: String) async throws -> [DocumentPermission] {
        let req = try URLRequest(url: URL(string: "\(APIConfig.baseURL)/api/v1/documents/\(documentId)/permissions")!, method: .get)
        let (data, _) = try await client.execute(req)
        return try JSONDecoder().decode([DocumentPermission].self, from: data)
    }
}
```

## Loading, Error, Empty States

```swift
struct DocumentListLoadingView: View {
    var body: some View {
        List(0..<5) { _ in
            HStack(spacing: 12) {
                RoundedRectangle(cornerRadius: 8)
                    .fill(.quaternary)
                    .frame(width: 32, height: 32)
                VStack(alignment: .leading, spacing: 6) {
                    RoundedRectangle(cornerRadius: 4).fill(.quaternary).frame(width: 140, height: 14)
                    RoundedRectangle(cornerRadius: 4).fill(.quaternary).frame(width: 200, height: 10)
                }
                Spacer()
            }
            .padding(.vertical, 6)
        }
        .listStyle(.insetGrouped)
        .redacted(reason: .placeholder)
    }
}

struct DocumentListErrorView: View {
    let error: KnowledgeError
    let onRetry: () -> Void
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle").font(.largeTitle).foregroundStyle(.orange)
            Text(error.localizedDescription).multilineTextAlignment(.center)
            Button("Retry", action: onRetry).buttonStyle(.borderedProminent)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct DocumentListEmptyView: View {
    let onUpload: () -> Void
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "doc.on.clipboard").font(.system(size: 48)).foregroundStyle(.secondary)
            Text("No Documents Yet").font(.title2.bold())
            Text("Upload documents to build your knowledge base.").foregroundStyle(.secondary)
            Button("Upload Document", action: onUpload).buttonStyle(.borderedProminent)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
```

## Responsive Design

| Screen | iPhone | iPad |
|---|---|---|
| Document List | Single column list | Grid + detail split |
| Document Upload | Full screen sheet | Centered sheet |
| Document Detail | Tab view | Sidebar tabs |
| Search | Full screen | Split view results |
| Document Sets | List + detail nav | Two-column layout |

## File Structure

```
KnowledgeCenter/
├── Views/
│   ├── DocumentListView.swift
│   ├── DocumentRowView.swift
│   ├── DocumentDetailView.swift
│   ├── DocumentUploadSheet.swift
│   ├── DocumentMetadataEditorSheet.swift
│   ├── DocumentSearchView.swift
│   ├── DocumentSetsView.swift
│   ├── DocumentSetDetailView.swift
│   ├── CreateDocumentSetSheet.swift
│   ├── CameraCaptureView.swift
│   ├── ChunkRowView.swift
│   ├── ProcessingLogRow.swift
│   └── States/
│       ├── DocumentListLoadingView.swift
│       ├── DocumentListErrorView.swift
│       └── DocumentListEmptyView.swift
├── ViewModels/
│   ├── DocumentListViewModel.swift
│   ├── DocumentUploadViewModel.swift
│   ├── DocumentDetailViewModel.swift
│   ├── DocumentSearchViewModel.swift
│   ├── DocumentSetsViewModel.swift
│   └── DocumentSetDetailViewModel.swift
├── Models/
│   ├── Document.swift
│   ├── DocumentChunk.swift
│   ├── DocumentSet.swift
│   ├── UploadProgress.swift
│   └── DocumentPermission.swift
├── Services/
│   ├── DocumentAPIService.swift
│   └── DocumentCacheService.swift
└── Utils/
    ├── FileUploader.swift
    └── FileTypeValidator.swift
```
