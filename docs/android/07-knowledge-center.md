# 07 - Knowledge Center

The knowledge center manages the entire document lifecycle: upload, processing,
search, organization into sets, and access control. Documents are the grounding
data for RAG-based agent queries.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Knowledge Center Screen](#knowledge-center-screen)
3. [Document List ViewModel](#document-list-viewmodel)
4. [Document Upload Screen](#document-upload-screen)
5. [Document Upload ViewModel](#document-upload-viewmodel)
6. [Document Processing Status](#document-processing-status)
7. [Document List View](#document-list-view)
8. [Document Detail Screen](#document-detail-screen)
9. [Document Search](#document-search)
10. [Document Sets Management](#document-sets-management)
11. [Document Set Detail](#document-set-detail)
12. [Document Permissions](#document-permissions)
13. [Document Deletion](#document-deletion)
14. [Supported File Types](#supported-file-types)
15. [File Size Limits and Validation](#file-size-limits-and-validation)
16. [Upload Progress Tracking](#upload-progress-tracking)
17. [Processing Pipeline Visualization](#processing-pipeline-visualization)
18. [Document Chunk Viewer](#document-chunk-viewer)
19. [Document Metadata Editor](#document-metadata-editor)
20. [Camera Capture](#camera-capture)
21. [Gallery Pick](#gallery-pick)
22. [Document API Integration](#document-api-integration)
23. [Document Data Models](#document-data-models)
24. [Document Caching](#document-caching)
25. [Pull-to-Refresh](#pull-to-refresh)
26. [Empty States](#empty-states)
27. [Loading States](#loading-states)
28. [Accessibility](#accessibility)
29. [Responsive Design](#responsive-design)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                         UI Layer                                 │
│  KnowledgeCenterScreen ──► DocumentListScreen                    │
│       │                    DocumentUploadScreen                  │
│       │                    DocumentDetailScreen                  │
│       │                    DocumentSetScreen                     │
│       ▼                                                          │
│  DocumentListViewModel    DocumentUploadViewModel                │
│  DocumentDetailViewModel  DocumentSetViewModel                   │
│       │                                                          │
│  ┌────┴──────────────────────────────────────────┐              │
│  │            DocumentRepository                  │              │
│  │  ├── getDocuments()    ├── uploadDocument()    │              │
│  │  ├── getDocument(id)   ├── deleteDocument()    │              │
│  │  ├── searchDocuments() ├── updateMetadata()    │              │
│  │  ├── getSets()         ├── createSet()         │              │
│  │  ├── addToSet()        ├── removeFromSet()     │              │
│  │  └── getProcessingStatus()                     │              │
│  └───────────────────────────────────────────────┘              │
│                         │                                        │
│  Data Layer             ▼                                        │
│  ├── DocumentApiService (Retrofit + Multipart)                   │
│  ├── DocumentDao (Room)                                          │
│  └── ProcessingPollingManager (status polling)                   │
└──────────────────────────────────────────────────────────────────┘
```

### Processing Pipeline

```
Upload ──► Extract ──► Chunk ──► Embed ──► Ready
  │          │          │         │         │
  ▼          ▼          ▼         ▼         ▼
Sent to    Text/PDF   Split     Vector    Available
server     parsed     into      embeds    for RAG
           content    chunks    generated  queries
```

---

## Knowledge Center Screen

Top-level screen with tabs for documents and document sets.

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun KnowledgeCenterScreen(
    documentListViewModel: DocumentListViewModel = hiltViewModel(),
    documentSetViewModel: DocumentSetViewModel = hltViewModel(),
    onDocumentClick: (String) -> Unit,
    onUploadClick: () -> Unit
) {
    var selectedTab by remember { mutableIntStateOf(0) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Knowledge Center") }
            )
        },
        floatingActionButton = {
            if (selectedTab == 0) {
                ExtendedFloatingActionButton(
                    onClick = onUploadClick,
                    icon = { Icon(Icons.Default.Upload, "Upload") },
                    text = { Text("Upload") }
                )
            }
        }
    ) { padding ->
        Column(modifier = Modifier.padding(padding)) {
            TabRow(selectedTabIndex = selectedTab) {
                Tab(
                    selected = selectedTab == 0,
                    onClick = { selectedTab = 0 },
                    text = { Text("Documents") },
                    icon = { Icon(Icons.Default.Description, contentDescription = null) }
                )
                Tab(
                    selected = selectedTab == 1,
                    onClick = { selectedTab = 1 },
                    text = { Text("Sets") },
                    icon = { Icon(Icons.Default.Folder, contentDescription = null) }
                )
            }

            when (selectedTab) {
                0 -> DocumentListContent(
                    viewModel = documentListViewModel,
                    onDocumentClick = onDocumentClick
                )
                1 -> DocumentSetListContent(
                    viewModel = documentSetViewModel,
                    onSetClick = { /* navigate */ }
                )
            }
        }
    }
}
```

---

## Document List ViewModel

```kotlin
@HiltViewModel
class DocumentListViewModel @Inject constructor(
    private val documentRepository: DocumentRepository,
    @IoDispatcher private val ioDispatcher: CoroutineDispatcher
) : ViewModel() {

    private val _state = MutableStateFlow(DocumentListState())
    val state: StateFlow<DocumentListState> = _state.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing: StateFlow<Boolean> = _isRefreshing.asStateFlow()

    private var allDocuments: List<DocumentSummary> = emptyList()

    init {
        loadDocuments()
    }

    fun onAction(action: DocumentListAction) {
        when (action) {
            is DocumentListAction.Refresh -> refresh()
            is DocumentListAction.UpdateSearch -> {
                _state.update { it.copy(searchQuery = action.query) }
                filterDocuments()
            }
            is DocumentListAction.FilterByType -> {
                _state.update {
                    it.copy(typeFilter = if (it.typeFilter == action.type) null else action.type)
                }
                filterDocuments()
            }
            is DocumentListAction.FilterByStatus -> {
                _state.update {
                    it.copy(statusFilter = if (it.statusFilter == action.status) null else action.status)
                }
                filterDocuments()
            }
            is DocumentListAction.SortBy -> {
                _state.update { it.copy(sortBy = action.sort) }
                filterDocuments()
            }
            is DocumentListAction.DeleteDocument -> deleteDocument(action.documentId)
        }
    }

    private fun loadDocuments() {
        viewModelScope.launch(ioDispatcher) {
            _state.update { it.copy(isLoading = true, error = null) }
            try {
                allDocuments = documentRepository.getDocuments()
                filterDocuments()
                _state.update { it.copy(isLoading = false) }
            } catch (e: Exception) {
                _state.update {
                    it.copy(isLoading = false, error = e.message ?: "Failed to load documents")
                }
            }
        }
    }

    private fun filterDocuments() {
        viewModelScope.launch(ioDispatcher) {
            val s = _state.value
            var filtered = allDocuments

            if (s.searchQuery.isNotBlank()) {
                filtered = filtered.filter {
                    it.name.contains(s.searchQuery, ignoreCase = true) ||
                    it.tags.any { tag -> tag.contains(s.searchQuery, ignoreCase = true) }
                }
            }

            if (s.typeFilter != null) {
                filtered = filtered.filter { it.fileType == s.typeFilter }
            }

            if (s.statusFilter != null) {
                filtered = filtered.filter { it.processingStatus == s.statusFilter }
            }

            filtered = when (s.sortBy) {
                DocumentSort.NAME -> filtered.sortedBy { it.name.lowercase() }
                DocumentSort.DATE_DESC -> filtered.sortedByDescending { it.uploadedAt }
                DocumentSort.DATE_ASC -> filtered.sortedBy { it.uploadedAt }
                DocumentSort.SIZE -> filtered.sortedByDescending { it.fileSizeBytes }
                DocumentSort.CHUNKS -> filtered.sortedByDescending { it.chunkCount }
            }

            _state.update { it.copy(documents = filtered) }
        }
    }

    private fun deleteDocument(documentId: String) {
        viewModelScope.launch(ioDispatcher) {
            try {
                documentRepository.deleteDocument(documentId)
                allDocuments = allDocuments.filter { it.documentId != documentId }
                filterDocuments()
                _state.update { it.copy(deletedDocumentId = documentId) }
            } catch (e: Exception) {
                _state.update { it.copy(error = "Failed to delete: ${e.message}") }
            }
        }
    }

    private fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            loadDocuments()
            _isRefreshing.value = false
        }
    }
}

data class DocumentListState(
    val isLoading: Boolean = false,
    val documents: List<DocumentSummary> = emptyList(),
    val searchQuery: String = "",
    val typeFilter: FileType? = null,
    val statusFilter: ProcessingStatus? = null,
    val sortBy: DocumentSort = DocumentSort.DATE_DESC,
    val error: String? = null,
    val deletedDocumentId: String? = null
)

sealed interface DocumentListAction {
    data object Refresh : DocumentListAction
    data class UpdateSearch(val query: String) : DocumentListAction
    data class FilterByType(val type: FileType) : DocumentListAction
    data class FilterByStatus(val status: ProcessingStatus) : DocumentListAction
    data class SortBy(val sort: DocumentSort) : DocumentListAction
    data class DeleteDocument(val documentId: String) : DocumentListAction
}

enum class DocumentSort(val label: String) {
    NAME("Name"), DATE_DESC("Newest"), DATE_ASC("Oldest"),
    SIZE("Largest"), CHUNKS("Most Chunks")
}
```

---

## Document Upload Screen

Supports file picker, drag-and-drop, camera capture, and gallery pick.

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DocumentUploadScreen(
    viewModel: DocumentUploadViewModel = hiltViewModel(),
    onBack: () -> Unit,
    onUploadComplete: () -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val context = LocalContext.current

    val filePickerLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.OpenMultipleDocuments()
    ) { uris ->
        viewModel.onAction(UploadAction.FilesSelected(uris))
    }

    val cameraLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.TakePicture()
    ) { success ->
        if (success) viewModel.onAction(UploadAction.CameraImageCaptured)
    }

    val galleryLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetMultipleContents()
    ) { uris ->
        viewModel.onAction(UploadAction.FilesSelected(uris))
    }

    LaunchedEffect(state.uploadComplete) {
        if (state.uploadComplete) onUploadComplete()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Upload Documents") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back")
                    }
                },
                actions = {
                    if (state.selectedFiles.isNotEmpty() && !state.isUploading) {
                        TextButton(onClick = { viewModel.onAction(UploadAction.StartUpload) }) {
                            Text("Upload (${state.selectedFiles.size})")
                        }
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize()
        ) {
            // Upload area / drop zone
            if (state.selectedFiles.isEmpty()) {
                UploadDropZone(
                    onPickFiles = { filePickerLauncher.launch(arrayOf("*/*")) },
                    onTakePhoto = { cameraLauncher.launch(createTempImageUri(context)) },
                    onPickImages = { galleryLauncher.launch("image/*") },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp)
                        .weight(1f)
                )
            } else {
                // Selected files list
                LazyColumn(
                    modifier = Modifier.weight(1f),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    items(
                        state.selectedFiles,
                        key = { it.uri.toString() }
                    ) { file ->
                        SelectedFileCard(
                            file = file,
                            progress = state.uploadProgress[file.uri.toString()],
                            onRemove = { viewModel.onAction(UploadAction.RemoveFile(file.uri)) }
                        )
                    }
                }

                // Upload button
                if (!state.isUploading) {
                    Row(
                        modifier = Modifier.padding(16.dp),
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        OutlinedButton(
                            onClick = { filePickerLauncher.launch(arrayOf("*/*")) },
                            modifier = Modifier.weight(1f)
                        ) {
                            Icon(Icons.Default.Add, null, modifier = Modifier.size(16.dp))
                            Spacer(Modifier.width(4.dp))
                            Text("Add Files")
                        }
                        Button(
                            onClick = { viewModel.onAction(UploadAction.StartUpload) },
                            modifier = Modifier.weight(1f)
                        ) {
                            Icon(Icons.Default.CloudUpload, null, modifier = Modifier.size(16.dp))
                            Spacer(Modifier.width(4.dp))
                            Text("Upload All")
                        }
                    }
                }

                // Upload progress
                if (state.isUploading) {
                    UploadProgressBar(
                        totalFiles = state.selectedFiles.size,
                        completedFiles = state.completedUploads,
                        currentProgress = state.currentUploadProgress,
                        modifier = Modifier.padding(16.dp)
                    )
                }
            }
        }
    }
}

@Composable
fun UploadDropZone(
    onPickFiles: () -> Unit,
    onTakePhoto: () -> Unit,
    onPickImages: () -> Unit,
    modifier: Modifier = Modifier
) {
    var isDragOver by remember { mutableStateOf(false) }

    Card(
        modifier = modifier,
        colors = CardDefaults.cardColors(
            containerColor = if (isDragOver)
                MaterialTheme.colorScheme.primaryContainer
            else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f)
        ),
        shape = RoundedCornerShape(16.dp),
        border = BorderStroke(
            width = 2.dp,
            color = if (isDragOver) MaterialTheme.colorScheme.primary
            else MaterialTheme.colorScheme.outline.copy(alpha = 0.3f)
        )
    ) {
        Column(
            modifier = Modifier.fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Icon(
                Icons.Default.CloudUpload,
                contentDescription = null,
                modifier = Modifier.size(64.dp),
                tint = MaterialTheme.colorScheme.primary.copy(alpha = 0.6f)
            )
            Spacer(Modifier.height(16.dp))
            Text("Drop files here", style = MaterialTheme.typography.titleMedium)
            Spacer(Modifier.height(4.dp))
            Text(
                "or choose an option below",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Spacer(Modifier.height(24.dp))

            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                FilledTonalButton(onClick = onPickFiles) {
                    Icon(Icons.Default.FolderOpen, null, modifier = Modifier.size(16.dp))
                    Spacer(Modifier.width(4.dp))
                    Text("Files")
                }
                FilledTonalButton(onClick = onTakePhoto) {
                    Icon(Icons.Default.CameraAlt, null, modifier = Modifier.size(16.dp))
                    Spacer(Modifier.width(4.dp))
                    Text("Camera")
                }
                FilledTonalButton(onClick = onPickImages) {
                    Icon(Icons.Default.Photo, null, modifier = Modifier.size(16.dp))
                    Spacer(Modifier.width(4.dp))
                    Text("Gallery")
                }
            }

            Spacer(Modifier.height(24.dp))

            Text(
                text = "Supported: PDF, DOCX, TXT, MD, HTML, JSON, CSV\nMax size: 50MB per file",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                textAlign = TextAlign.Center
            )
        }
    }
}

@Composable
fun SelectedFileCard(
    file: SelectedFile,
    progress: Float?,
    onRemove: () -> Unit
) {
    val (icon, color) = getFileIcon(file.fileType)

    Card {
        Row(
            modifier = Modifier.padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .clip(RoundedCornerShape(8.dp))
                    .background(color.copy(alpha = 0.12f)),
                contentAlignment = Alignment.Center
            ) {
                Icon(icon, contentDescription = null, tint = color, modifier = Modifier.size(20.dp))
            }
            Spacer(Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = file.name,
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.Medium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                Text(
                    text = "${file.formattedSize} · ${file.fileType.name}",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                if (progress != null) {
                    Spacer(Modifier.height(4.dp))
                    LinearProgressIndicator(
                        progress = { progress },
                        modifier = Modifier.fillMaxWidth().height(4.dp).clip(RoundedCornerShape(2.dp))
                    )
                }
            }
            IconButton(onClick = onRemove, enabled = progress == null) {
                Icon(Icons.Default.Close, "Remove", modifier = Modifier.size(18.dp))
            }
        }
    }
}

@Composable
fun UploadProgressBar(
    totalFiles: Int,
    completedFiles: Int,
    currentProgress: Float,
    modifier: Modifier = Modifier
) {
    Card(modifier = modifier) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text("Uploading...", style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.SemiBold)
                Text(
                    "$completedFiles / $totalFiles",
                    style = MaterialTheme.typography.bodySmall
                )
            }
            Spacer(Modifier.height(8.dp))
            LinearProgressIndicator(
                progress = { (completedFiles + currentProgress) / totalFiles },
                modifier = Modifier.fillMaxWidth().height(6.dp).clip(RoundedCornerShape(3.dp))
            )
        }
    }
}

fun getFileIcon(type: FileType): Pair<ImageVector, Color> = when (type) {
    FileType.PDF -> Icons.Default.PictureAsPdf to Color(0xFFF44336)
    FileType.DOCX -> Icons.Default.Article to Color(0xFF2196F3)
    FileType.TXT, FileType.MD -> Icons.Default.TextSnippet to Color(0xFF607D8B)
    FileType.HTML -> Icons.Default.Language to Color(0xFFFF9800)
    FileType.JSON -> Icons.Default.DataObject to Color(0xFF4CAF50)
    FileType.CSV -> Icons.Default.TableChart to Color(0xFF9C27B0)
    FileType.IMAGE -> Icons.Default.Image to Color(0xFFE91E63)
    FileType.CODE -> Icons.Default.Code to Color(0xFF00BCD4)
    FileType.UNKNOWN -> Icons.Default.InsertDriveFile to Color(0xFF9E9E9E)
}
```

---

## Document Upload ViewModel

```kotlin
@HiltViewModel
class DocumentUploadViewModel @Inject constructor(
    private val documentRepository: DocumentRepository,
    private val fileValidator: FileValidator,
    @IoDispatcher private val ioDispatcher: CoroutineDispatcher,
    @ApplicationContext private val context: Context
) : ViewModel() {

    private val _state = MutableStateFlow(UploadState())
    val state: StateFlow<UploadState> = _state.asStateFlow()

    fun onAction(action: UploadAction) {
        when (action) {
            is UploadAction.FilesSelected -> addFiles(action.uris)
            is UploadAction.RemoveFile -> removeFile(action.uri)
            is UploadAction.StartUpload -> startUpload()
            is UploadAction.CameraImageCaptured -> handleCameraCapture()
        }
    }

    private fun addFiles(uris: List<Uri>) {
        viewModelScope.launch(ioDispatcher) {
            val newFiles = uris.mapNotNull { uri ->
                val result = fileValidator.validate(uri)
                when (result) {
                    is ValidationResult.Valid -> SelectedFile(
                        uri = uri,
                        name = result.fileName,
                        fileSizeBytes = result.fileSize,
                        fileType = result.fileType,
                        formattedSize = formatFileSize(result.fileSize)
                    )
                    is ValidationResult.Invalid -> {
                        _state.update { it.copy(error = result.reason) }
                        null
                    }
                }
            }

            _state.update { state ->
                state.copy(
                    selectedFiles = (state.selectedFiles + newFiles).distinctBy { it.uri },
                    error = null
                )
            }
        }
    }

    private fun removeFile(uri: Uri) {
        _state.update { state ->
            state.copy(
                selectedFiles = state.selectedFiles.filter { it.uri != uri },
                uploadProgress = state.uploadProgress - uri.toString()
            )
        }
    }

    private fun startUpload() {
        viewModelScope.launch(ioDispatcher) {
            _state.update { it.copy(isUploading = true, completedUploads = 0) }

            val files = _state.value.selectedFiles
            var completed = 0

            files.forEach { file ->
                try {
                    documentRepository.uploadDocument(
                        uri = file.uri,
                        fileName = file.name,
                        onProgress = { progress ->
                            _state.update { state ->
                                state.copy(
                                    uploadProgress = state.uploadProgress + (file.uri.toString() to progress),
                                    currentUploadProgress = progress
                                )
                            }
                        }
                    )
                    completed++
                    _state.update { it.copy(completedUploads = completed) }
                } catch (e: Exception) {
                    _state.update {
                        it.copy(
                            isUploading = false,
                            error = "Failed to upload ${file.name}: ${e.message}"
                        )
                    }
                    return@launch
                }
            }

            _state.update {
                it.copy(isUploading = false, uploadComplete = true)
            }
        }
    }

    private fun handleCameraCapture() {
        // Camera image URI is stored in state by the camera launcher
    }
}

data class UploadState(
    val selectedFiles: List<SelectedFile> = emptyList(),
    val isUploading: Boolean = false,
    val uploadProgress: Map<String, Float> = emptyMap(),
    val completedUploads: Int = 0,
    val currentUploadProgress: Float = 0f,
    val uploadComplete: Boolean = false,
    val error: String? = null
)

data class SelectedFile(
    val uri: Uri,
    val name: String,
    val fileSizeBytes: Long,
    val fileType: FileType,
    val formattedSize: String
)

sealed interface UploadAction {
    data class FilesSelected(val uris: List<Uri>) : UploadAction
    data class RemoveFile(val uri: Uri) : UploadAction
    data object StartUpload : UploadAction
    data object CameraImageCaptured : UploadAction
}
```

---

## Document Processing Status

### Processing Stages

| Stage      | Description                    | Duration     | Indicators                      |
|------------|--------------------------------|--------------|---------------------------------|
| Uploaded   | File received by server        | Instant      | Checkmark                       |
| Extracting | Text/image content parsed      | 1–30s        | Spinner + "Extracting text..."  |
| Chunking   | Content split into chunks      | 1–5s         | Spinner + "Creating chunks..."  |
| Embedding  | Vector embeddings generated    | 5–60s        | Spinner + "Generating vectors"  |
| Ready      | Document available for queries | —            | Green checkmark                 |
| Failed     | Error at any stage             | —            | Red error icon + retry          |

```kotlin
@Composable
fun ProcessingStatusBadge(status: ProcessingStatus) {
    val (icon, color, label) = when (status) {
        ProcessingStatus.UPLOADED -> Triple(Icons.Default.CloudDone, Color(0xFF4CAF50), "Uploaded")
        ProcessingStatus.EXTRACTING -> Triple(Icons.Default.Sync, MaterialTheme.colorScheme.primary, "Extracting")
        ProcessingStatus.CHUNKING -> Triple(Icons.Default.Sync, MaterialTheme.colorScheme.secondary, "Chunking")
        ProcessingStatus.EMBEDDING -> Triple(Icons.Default.Sync, MaterialTheme.colorScheme.tertiary, "Embedding")
        ProcessingStatus.READY -> Triple(Icons.Default.CheckCircle, Color(0xFF4CAF50), "Ready")
        ProcessingStatus.FAILED -> Triple(Icons.Default.Error, MaterialTheme.colorScheme.error, "Failed")
    }

    Surface(
        shape = RoundedCornerShape(12.dp),
        color = color.copy(alpha = 0.12f)
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            if (status in listOf(ProcessingStatus.EXTRACTING, ProcessingStatus.CHUNKING, ProcessingStatus.EMBEDDING)) {
                CircularProgressIndicator(
                    modifier = Modifier.size(12.dp),
                    strokeWidth = 1.5.dp,
                    color = color
                )
            } else {
                Icon(icon, contentDescription = null, modifier = Modifier.size(12.dp), tint = color)
            }
            Spacer(Modifier.width(4.dp))
            Text(label, style = MaterialTheme.typography.labelSmall, color = color, fontWeight = FontWeight.Medium)
        }
    }
}
```

### Processing Status Polling

```kotlin
class ProcessingPollingManager @Inject constructor(
    private val documentRepository: DocumentRepository
) {
    private val _statusUpdates = MutableSharedFlow<DocumentProcessingUpdate>(
        replay = 0,
        extraBufferCapacity = 32
    )
    val statusUpdates: SharedFlow<DocumentProcessingUpdate> = _statusUpdates.asSharedFlow()

    private var pollingJobs = mutableMapOf<String, Job>()

    fun startPolling(documentId: String, intervalMs: Long = 2000L) {
        pollingJobs[documentId]?.cancel()
        pollingJobs[documentId] = CoroutineScope(Dispatchers.IO).launch {
            while (isActive) {
                try {
                    val status = documentRepository.getProcessingStatus(documentId)
                    _statusUpdates.emit(DocumentProcessingUpdate(documentId, status))
                    if (status == ProcessingStatus.READY || status == ProcessingStatus.FAILED) {
                        break
                    }
                } catch (e: Exception) {
                    _statusUpdates.emit(DocumentProcessingUpdate(documentId, ProcessingStatus.FAILED))
                    break
                }
                delay(intervalMs)
            }
        }
    }

    fun stopPolling(documentId: String) {
        pollingJobs[documentId]?.cancel()
        pollingJobs.remove(documentId)
    }

    fun stopAll() {
        pollingJobs.values.forEach { it.cancel() }
        pollingJobs.clear()
    }
}

data class DocumentProcessingUpdate(
    val documentId: String,
    val status: ProcessingStatus
)
```

---

## Document List View

```kotlin
@Composable
fun DocumentListContent(
    viewModel: DocumentListViewModel,
    onDocumentClick: (String) -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val isRefreshing by viewModel.isRefreshing.collectAsStateWithLifecycle()

    Column {
        // Search bar
        DocumentSearchBar(
            query = state.searchQuery,
            onQueryChange = { viewModel.onAction(DocumentListAction.UpdateSearch(it)) },
            typeFilter = state.typeFilter,
            onTypeFilter = { viewModel.onAction(DocumentListAction.FilterByType(it)) },
            statusFilter = state.statusFilter,
            onStatusFilter = { viewModel.onAction(DocumentListAction.FilterByStatus(it)) },
            sortBy = state.sortBy,
            onSortChange = { viewModel.onAction(DocumentListAction.SortBy(it)) }
        )

        SwipeRefresh(
            state = rememberSwipeRefreshState(isRefreshing),
            onRefresh = { viewModel.onAction(DocumentListAction.Refresh) }
        ) {
            when {
                state.isLoading -> DocumentListSkeleton()
                state.error != null -> DocumentListError(
                    error = state.error!!,
                    onRetry = { viewModel.onAction(DocumentListAction.Refresh) }
                )
                state.documents.isEmpty() -> DocumentListEmpty()
                else -> DocumentTable(
                    documents = state.documents,
                    onDocumentClick = onDocumentClick,
                    onDeleteClick = { viewModel.onAction(DocumentListAction.DeleteDocument(it)) }
                )
            }
        }
    }
}

@Composable
fun DocumentTable(
    documents: List<DocumentSummary>,
    onDocumentClick: (String) -> Unit,
    onDeleteClick: (String) -> Unit
) {
    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        // Header
        item {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("Name", modifier = Modifier.weight(2f), style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                Text("Type", modifier = Modifier.weight(0.8f), style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                Text("Status", modifier = Modifier.weight(1f), style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                Text("Chunks", modifier = Modifier.weight(0.7f), style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold, textAlign = TextAlign.End)
                Text("Date", modifier = Modifier.weight(1f), style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                Spacer(Modifier.width(40.dp))
            }
        }

        items(documents, key = { it.documentId }) { doc ->
            DocumentRow(
                document = doc,
                onClick = { onDocumentClick(doc.documentId) },
                onDelete = { onDeleteClick(doc.documentId) }
            )
        }
    }
}

@Composable
fun DocumentRow(
    document: DocumentSummary,
    onClick: () -> Unit,
    onDelete: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        // Name + tags
        Column(modifier = Modifier.weight(2f)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    getFileIcon(document.fileType).first,
                    contentDescription = null,
                    modifier = Modifier.size(16.dp),
                    tint = getFileIcon(document.fileType).second
                )
                Spacer(Modifier.width(6.dp))
                Text(
                    text = document.name,
                    style = MaterialTheme.typography.bodyMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
            }
            if (document.tags.isNotEmpty()) {
                Spacer(Modifier.height(2.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                    document.tags.take(2).forEach { tag ->
                        Surface(
                            shape = RoundedCornerShape(4.dp),
                            color = MaterialTheme.colorScheme.secondaryContainer
                        ) {
                            Text(
                                text = tag,
                                modifier = Modifier.padding(horizontal = 4.dp, vertical = 1.dp),
                                style = MaterialTheme.typography.labelSmall
                            )
                        }
                    }
                }
            }
        }

        // Type
        Text(
            text = document.fileType.name,
            style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.weight(0.8f)
        )

        // Status
        Box(modifier = Modifier.weight(1f)) {
            ProcessingStatusBadge(status = document.processingStatus)
        }

        // Chunks
        Text(
            text = "${document.chunkCount}",
            style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.weight(0.7f),
            textAlign = TextAlign.End
        )

        // Date
        Text(
            text = document.uploadedAt.toFormattedDate(),
            style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.weight(1f)
        )

        // Delete
        IconButton(onClick = onDelete, modifier = Modifier.size(32.dp)) {
            Icon(
                Icons.Default.Delete,
                contentDescription = "Delete",
                modifier = Modifier.size(16.dp),
                tint = MaterialTheme.colorScheme.error.copy(alpha = 0.6f)
            )
        }
    }
}
```

---

## Document Detail Screen

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DocumentDetailScreen(
    documentId: String,
    viewModel: DocumentDetailViewModel = hiltViewModel(),
    onBack: () -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var selectedTab by remember { mutableIntStateOf(0) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(state.document?.name ?: "Document") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { viewModel.onAction(DetailAction.EditMetadata) }) {
                        Icon(Icons.Default.Edit, "Edit metadata")
                    }
                    IconButton(onClick = { viewModel.onAction(DetailAction.DeleteDocument) }) {
                        Icon(Icons.Default.Delete, "Delete", tint = MaterialTheme.colorScheme.error)
                    }
                }
            )
        }
    ) { padding ->
        Column(modifier = Modifier.padding(padding)) {
            // Processing status banner
            state.document?.let { doc ->
                if (doc.processingStatus != ProcessingStatus.READY) {
                    ProcessingBanner(status = doc.processingStatus)
                }
            }

            TabRow(selectedTabIndex = selectedTab) {
                Tab(selected = selectedTab == 0, onClick = { selectedTab = 0 }, text = { Text("Metadata") })
                Tab(selected = selectedTab == 1, onClick = { selectedTab = 1 }, text = { Text("Chunks") })
                Tab(selected = selectedTab == 2, onClick = { selectedTab = 2 }, text = { Text("Log") })
            }

            when (selectedTab) {
                0 -> DocumentMetadataTab(document = state.document)
                1 -> DocumentChunksTab(
                    chunks = state.chunks,
                    onLoadMore = { viewModel.onAction(DetailAction.LoadMoreChunks) }
                )
                2 -> DocumentProcessingLog(logs = state.processingLogs)
            }
        }
    }
}

@Composable
fun DocumentMetadataTab(document: Document?) {
    if (document == null) return

    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item {
            MetadataSection("General") {
                MetadataRow("Name", document.name)
                MetadataRow("File Type", document.fileType.name)
                MetadataRow("File Size", document.formattedSize)
                MetadataRow("Uploaded", document.uploadedAt.toFormattedDateTime())
                MetadataRow("Status", document.processingStatus.name)
                MetadataRow("Chunks", "${document.chunkCount}")
            }
        }

        item {
            MetadataSection("Content") {
                MetadataRow("Characters", "${document.characterCount}")
                MetadataRow("Words", "${document.wordCount}")
                MetadataRow("Pages", document.pageCount?.toString() ?: "N/A")
                MetadataRow("Language", document.language ?: "N/A")
            }
        }

        item {
            MetadataSection("Tags") {
                FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    document.tags.forEach { tag ->
                        AssistChip(
                            onClick = {},
                            label = { Text(tag) },
                            leadingIcon = { Icon(Icons.Default.Label, null, modifier = Modifier.size(14.dp)) }
                        )
                    }
                }
            }
        }

        item {
            MetadataSection("Description") {
                Text(
                    text = document.description.ifEmpty { "No description" },
                    style = MaterialTheme.typography.bodyMedium
                )
            }
        }

        item {
            MetadataSection("Sets") {
                document.sets.forEach { set ->
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(Icons.Default.Folder, null, modifier = Modifier.size(16.dp))
                        Spacer(Modifier.width(8.dp))
                        Text(set.name, style = MaterialTheme.typography.bodyMedium)
                    }
                }
            }
        }
    }
}

@Composable
fun MetadataSection(title: String, content: @Composable ColumnScope.() -> Unit) {
    Card {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(title, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
            Spacer(Modifier.height(12.dp))
            content()
        }
    }
}

@Composable
fun MetadataRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 3.dp),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(label, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Text(value, style = MaterialTheme.typography.bodySmall, fontWeight = FontWeight.Medium)
    }
}

@Composable
fun ProcessingBanner(status: ProcessingStatus) {
    Surface(
        color = MaterialTheme.colorScheme.primaryContainer,
        modifier = Modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier.padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            CircularProgressIndicator(modifier = Modifier.size(16.dp), strokeWidth = 2.dp)
            Spacer(Modifier.width(8.dp))
            Text(
                text = "Processing: ${status.name.lowercase()}...",
                style = MaterialTheme.typography.bodySmall,
                fontWeight = FontWeight.Medium
            )
        }
    }
}
```

---

## Document Search

```kotlin
@Composable
fun DocumentSearchBar(
    query: String,
    onQueryChange: (String) -> Unit,
    typeFilter: FileType?,
    onTypeFilter: (FileType) -> Unit,
    statusFilter: ProcessingStatus?,
    onStatusFilter: (ProcessingStatus) -> Unit,
    sortBy: DocumentSort,
    onSortChange: (DocumentSort) -> Unit
) {
    Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)) {
        OutlinedTextField(
            value = query,
            onValueChange = onQueryChange,
            placeholder = { Text("Search documents...") },
            leadingIcon = { Icon(Icons.Default.Search, "Search") },
            trailingIcon = {
                if (query.isNotEmpty()) {
                    IconButton(onClick = { onQueryChange("") }) {
                        Icon(Icons.Default.Clear, "Clear")
                    }
                }
            },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )

        Spacer(Modifier.height(8.dp))

        Row(
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            modifier = Modifier.horizontalScroll(rememberScrollState())
        ) {
            // Type filter chips
            FileType.entries.take(6).forEach { type ->
                FilterChip(
                    selected = typeFilter == type,
                    onClick = { onTypeFilter(type) },
                    label = { Text(type.name, style = MaterialTheme.typography.labelSmall) }
                )
            }

            HorizontalDivider(
                modifier = Modifier
                    .width(1.dp)
                    .height(32.dp)
                    .align(Alignment.CenterVertically)
            )

            // Status filter chips
            ProcessingStatus.entries.forEach { status ->
                FilterChip(
                    selected = statusFilter == status,
                    onClick = { onStatusFilter(status) },
                    label = { Text(status.name, style = MaterialTheme.typography.labelSmall) }
                )
            }
        }
    }
}
```

### Semantic Search Results

```kotlin
@Composable
fun SemanticSearchResults(
    results: List<SearchResult>,
    modifier: Modifier = Modifier
) {
    LazyColumn(
        modifier = modifier,
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        items(results, key = { it.chunkId }) { result ->
            SearchResultCard(result = result)
        }
    }
}

@Composable
fun SearchResultCard(result: SearchResult) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = result.documentName,
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = MaterialTheme.colorScheme.primary
                )
                Surface(
                    shape = RoundedCornerShape(12.dp),
                    color = MaterialTheme.colorScheme.tertiaryContainer
                ) {
                    Text(
                        text = "Score: ${String.format("%.2f", result.similarityScore)}",
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp),
                        style = MaterialTheme.typography.labelSmall
                    )
                }
            }
            Spacer(Modifier.height(8.dp))
            Text(
                text = result.chunkText,
                style = MaterialTheme.typography.bodySmall,
                maxLines = 4,
                overflow = TextOverflow.Ellipsis
            )
            Spacer(Modifier.height(4.dp))
            Text(
                text = "Chunk ${result.chunkIndex + 1} of ${result.totalChunks}",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

data class SearchResult(
    val chunkId: String,
    val documentId: String,
    val documentName: String,
    val chunkText: String,
    val chunkIndex: Int,
    val totalChunks: Int,
    val similarityScore: Float
)
```

---

## Document Sets Management

```kotlin
@Composable
fun DocumentSetListContent(
    viewModel: DocumentSetViewModel,
    onSetClick: (String) -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    when {
        state.isLoading -> DocumentSetSkeleton()
        state.sets.isEmpty() -> DocumentSetEmpty(onCreate = { viewModel.onAction(SetAction.CreateSet) })
        else -> LazyColumn(
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(state.sets, key = { it.setId }) { set ->
                DocumentSetCard(set = set, onClick = { onSetClick(set.setId) })
            }
        }
    }
}

@Composable
fun DocumentSetCard(
    set: DocumentSetSummary,
    onClick: () -> Unit
) {
    Card(onClick = onClick) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(RoundedCornerShape(12.dp))
                    .background(MaterialTheme.colorScheme.primaryContainer),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    Icons.Default.Folder,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.primary
                )
            }
            Spacer(Modifier.width(16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(set.name, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
                Text(
                    "${set.documentCount} documents · ${set.totalChunks} chunks",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                if (set.description.isNotEmpty()) {
                    Text(
                        set.description,
                        style = MaterialTheme.typography.bodySmall,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            Icon(Icons.Default.ChevronRight, contentDescription = "View", tint = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

@Composable
fun DocumentSetEmpty(onCreate: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            Icons.Default.FolderOpen,
            contentDescription = null,
            modifier = Modifier.size(64.dp),
            tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.4f)
        )
        Spacer(Modifier.height(16.dp))
        Text("No document sets", style = MaterialTheme.typography.titleMedium)
        Spacer(Modifier.height(8.dp))
        Text(
            "Create a set to organize related documents",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Spacer(Modifier.height(24.dp))
        Button(onClick = onCreate) {
            Icon(Icons.Default.Add, null)
            Spacer(Modifier.width(8.dp))
            Text("Create Set")
        }
    }
}
```

---

## Document Set Detail

```kotlin
@Composable
fun DocumentSetDetailScreen(
    setId: String,
    viewModel: DocumentSetDetailViewModel = hiltViewModel(),
    onBack: () -> Unit
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(state.set?.name ?: "Document Set") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { viewModel.onAction(SetDetailAction.EditSet) }) {
                        Icon(Icons.Default.Edit, "Edit")
                    }
                    IconButton(onClick = { viewModel.onAction(SetDetailAction.DeleteSet) }) {
                        Icon(Icons.Default.Delete, "Delete", tint = MaterialTheme.colorScheme.error)
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = { viewModel.onAction(SetDetailAction.AddDocuments) }) {
                Icon(Icons.Default.Add, "Add documents")
            }
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier.padding(padding),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            state.set?.let { set ->
                item {
                    Card {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(set.description.ifEmpty { "No description" }, style = MaterialTheme.typography.bodyMedium)
                            Spacer(Modifier.height(8.dp))
                            Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                                Text("${set.documents.size} documents", style = MaterialTheme.typography.labelMedium)
                                Text("${set.totalChunks} chunks", style = MaterialTheme.typography.labelMedium)
                            }
                        }
                    }
                }

                items(set.documents, key = { it.documentId }) { doc ->
                    DocumentRow(
                        document = doc,
                        onClick = { /* navigate to document */ },
                        onDelete = { viewModel.onAction(SetDetailAction.RemoveDocument(doc.documentId)) }
                    )
                }
            }
        }
    }
}
```

---

## Document Permissions

```kotlin
data class DocumentPermission(
    val userId: String,
    val userName: String,
    val level: AccessLevel,
    val grantedAt: Instant,
    val grantedBy: String
)

enum class AccessLevel(val label: String) {
    VIEWER("Can view"),
    EDITOR("Can edit"),
    ADMIN("Full access")
}

@Composable
fun DocumentPermissionsSection(
    permissions: List<DocumentPermission>,
    onGrant: (String, AccessLevel) -> Unit,
    onRevoke: (String) -> Unit,
    onUpdateLevel: (String, AccessLevel) -> Unit
) {
    Column {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text("Permissions", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
            TextButton(onClick = { /* open grant dialog */ }) {
                Icon(Icons.Default.PersonAdd, null, modifier = Modifier.size(16.dp))
                Spacer(Modifier.width(4.dp))
                Text("Grant Access")
            }
        }

        permissions.forEach { perm ->
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(vertical = 6.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Avatar
                Box(
                    modifier = Modifier
                        .size(32.dp)
                        .clip(CircleShape)
                        .background(MaterialTheme.colorScheme.primaryContainer),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        perm.userName.first().uppercase(),
                        style = MaterialTheme.typography.labelMedium
                    )
                }
                Spacer(Modifier.width(12.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(perm.userName, style = MaterialTheme.typography.bodyMedium)
                    Text("Granted by ${perm.grantedBy}", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                // Permission level dropdown
                var expanded by remember { mutableStateOf(false) }
                ExposedDropdownMenuBox(expanded = expanded, onExpandedChange = { expanded = it }) {
                    AssistChip(
                        onClick = { expanded = true },
                        label = { Text(perm.level.label, style = MaterialTheme.typography.labelSmall) },
                        modifier = Modifier.menuAnchor()
                    )
                    ExposedDropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
                        AccessLevel.entries.forEach { level ->
                            DropdownMenuItem(
                                text = { Text(level.label) },
                                onClick = { onUpdateLevel(perm.userId, level); expanded = false }
                            )
                        }
                    }
                }
            }
        }
    }
}
```

---

## Document Deletion

```kotlin
@Composable
fun DocumentDeleteDialog(
    documentName: String,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        icon = { Icon(Icons.Default.Warning, contentDescription = null, tint = MaterialTheme.colorScheme.error) },
        title = { Text("Delete Document") },
        text = {
            Text("Are you sure you want to delete \"$documentName\"? This action cannot be undone. The document and all its chunks will be permanently removed.")
        },
        confirmButton = {
            Button(
                onClick = onConfirm,
                colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.error)
            ) { Text("Delete") }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text("Cancel") }
        }
    )
}
```

---

## Supported File Types

| Extension | MIME Type                         | Processing  | Max Size |
|-----------|-----------------------------------|-------------|----------|
| PDF       | application/pdf                   | Text+Image  | 50MB     |
| DOCX      | application/vnd.openxmlformats... | Text+Image  | 50MB     |
| TXT       | text/plain                        | Text        | 10MB     |
| MD        | text/markdown                     | Text        | 10MB     |
| HTML      | text/html                         | Text        | 10MB     |
| JSON      | application/json                  | Text        | 10MB     |
| CSV       | text/csv                          | Text/Table  | 10MB     |
| XLSX      | application/vnd.openxmlformats... | Table       | 20MB     |
| JPG/PNG   | image/*                           | OCR         | 20MB     |
| PY/JS/TS  | text/*                            | Text        | 10MB     |

```kotlin
enum class FileType(
    val extensions: List<String>,
    val mimeTypes: List<String>,
    val maxSizeBytes: Long = 50 * 1024 * 1024
) {
    PDF(listOf("pdf"), listOf("application/pdf")),
    DOCX(listOf("docx"), listOf("application/vnd.openxmlformats-officedocument.wordprocessingml.document")),
    TXT(listOf("txt"), listOf("text/plain"), 10 * 1024 * 1024),
    MD(listOf("md", "markdown"), listOf("text/markdown"), 10 * 1024 * 1024),
    HTML(listOf("html", "htm"), listOf("text/html"), 10 * 1024 * 1024),
    JSON(listOf("json"), listOf("application/json"), 10 * 1024 * 1024),
    CSV(listOf("csv"), listOf("text/csv"), 10 * 1024 * 1024),
    IMAGE(listOf("jpg", "jpeg", "png", "gif", "webp"), listOf("image/*"), 20 * 1024 * 1024),
    CODE(listOf("py", "js", "ts", "kt", "java", "go", "rs", "cpp", "c", "rb"), listOf("text/*"), 10 * 1024 * 1024),
    UNKNOWN(listOf(), listOf(), 50 * 1024 * 1024);

    companion object {
        fun fromFileName(fileName: String): FileType {
            val ext = fileName.substringAfterLast('.').lowercase()
            return entries.find { ext in it.extensions } ?: UNKNOWN
        }

        fun fromMimeType(mime: String): FileType {
            return entries.find { mime in it.mimeTypes || mime.startsWith(it.mimeTypes.firstOrNull()?.substringBefore('/') ?: "") } ?: UNKNOWN
        }
    }
}
```

---

## File Size Limits and Validation

```kotlin
class FileValidator @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun validate(uri: Uri): ValidationResult {
        val cursor = context.contentResolver.query(uri, null, null, null, null) ?: return ValidationResult.Invalid("Cannot read file")

        cursor.use {
            if (!it.moveToFirst()) return ValidationResult.Invalid("Cannot read file")

            val nameIndex = it.getColumnIndex(OpenableColumns.DISPLAY_NAME)
            val sizeIndex = it.getColumnIndex(OpenableColumns.SIZE)

            val fileName = if (nameIndex >= 0) it.getString(nameIndex) else "unknown"
            val fileSize = if (sizeIndex >= 0) it.getLong(sizeIndex) else 0L

            val fileType = FileType.fromFileName(fileName)

            if (fileSize > fileType.maxSizeBytes) {
                return ValidationResult.Invalid(
                    "File \"$fileName\" exceeds ${formatFileSize(fileType.maxSizeBytes)} limit"
                )
            }

            if (fileType == FileType.UNKNOWN) {
                return ValidationResult.Invalid("Unsupported file type: ${fileName.substringAfterLast('.')}")
            }

            return ValidationResult.Valid(
                fileName = fileName,
                fileSize = fileSize,
                fileType = fileType
            )
        }
    }
}

sealed interface ValidationResult {
    data class Valid(
        val fileName: String,
        val fileSize: Long,
        val fileType: FileType
    ) : ValidationResult

    data class Invalid(val reason: String) : ValidationResult
}

fun formatFileSize(bytes: Long): String = when {
    bytes >= 1_073_741_824 -> "${bytes / 1_073_741_824} GB"
    bytes >= 1_048_576 -> "${bytes / 1_048_576} MB"
    bytes >= 1_024 -> "${bytes / 1_024} KB"
    else -> "$bytes B"
}
```

---

## Upload Progress Tracking

```kotlin
@Composable
fun UploadProgressOverlay(
    state: UploadState,
    modifier: Modifier = Modifier
) {
    if (state.isUploading) {
        Surface(
            modifier = modifier.fillMaxWidth(),
            color = MaterialTheme.colorScheme.surface,
            tonalElevation = 2.dp
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text(
                    "Uploading ${state.completedUploads + 1} of ${state.selectedFiles.size}",
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.SemiBold
                )
                Spacer(Modifier.height(8.dp))
                LinearProgressIndicator(
                    progress = {
                        (state.completedUploads + state.currentUploadProgress) /
                        state.selectedFiles.size
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(6.dp)
                        .clip(RoundedCornerShape(3.dp))
                )
                Spacer(Modifier.height(4.dp))
                Text(
                    text = "${state.selectedFiles.getOrNull(state.completedUploads)?.name ?: ""}",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
            }
        }
    }
}
```

---

## Processing Pipeline Visualization

```kotlin
@Composable
fun ProcessingPipeline(
    status: ProcessingStatus,
    modifier: Modifier = Modifier
) {
    val stages = ProcessingStatus.entries.filter { it != ProcessingStatus.FAILED }
    val currentIndex = stages.indexOf(status).coerceAtLeast(0)

    Row(
        modifier = modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        stages.forEachIndexed { index, stage ->
            val isActive = index == currentIndex
            val isCompleted = index < currentIndex || status == ProcessingStatus.READY

            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                modifier = Modifier.weight(1f)
            ) {
                // Stage circle
                Box(
                    modifier = Modifier
                        .size(32.dp)
                        .clip(CircleShape)
                        .background(
                            when {
                                isCompleted -> Color(0xFF4CAF50)
                                isActive -> MaterialTheme.colorScheme.primary
                                else -> MaterialTheme.colorScheme.surfaceVariant
                            }
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    if (isActive && status != ProcessingStatus.READY) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(20.dp),
                            strokeWidth = 2.dp,
                            color = MaterialTheme.colorScheme.onPrimary
                        )
                    } else {
                        Icon(
                            if (isCompleted) Icons.Default.Check else Icons.Default.Circle,
                            contentDescription = null,
                            modifier = Modifier.size(16.dp),
                            tint = if (isCompleted || isActive) MaterialTheme.colorScheme.onPrimary
                                   else MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
                Spacer(Modifier.height(4.dp))
                Text(
                    text = stage.name,
                    style = MaterialTheme.typography.labelSmall,
                    fontWeight = if (isActive) FontWeight.Bold else FontWeight.Normal,
                    color = if (isActive) MaterialTheme.colorScheme.primary
                           else MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            // Connector line
            if (index < stages.lastIndex) {
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .height(2.dp)
                        .background(
                            if (index < currentIndex) Color(0xFF4CAF50)
                            else MaterialTheme.colorScheme.surfaceVariant
                        )
                )
            }
        }
    }
}
```

---

## Document Chunk Viewer

```kotlin
@Composable
fun DocumentChunksTab(
    chunks: List<DocumentChunk>,
    onLoadMore: () -> Unit
) {
    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        items(chunks, key = { it.chunkId }) { chunk ->
            ChunkCard(chunk = chunk)
        }

        item {
            LaunchedEffect(Unit) { onLoadMore() }
            Box(modifier = Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(modifier = Modifier.size(24.dp))
            }
        }
    }
}

@Composable
fun ChunkCard(chunk: DocumentChunk) {
    var expanded by remember { mutableStateOf(false) }

    Card(
        onClick = { expanded = !expanded },
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    "Chunk ${chunk.index + 1}",
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.SemiBold
                )
                Text(
                    "${chunk.charCount} chars",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            Spacer(Modifier.height(8.dp))
            Text(
                text = if (expanded) chunk.text else chunk.text.take(200) + if (chunk.text.length > 200) "..." else "",
                style = MaterialTheme.typography.bodySmall,
                fontFamily = FontFamily.Monospace,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            if (chunk.text.length > 200) {
                Spacer(Modifier.height(4.dp))
                Text(
                    text = if (expanded) "Show less" else "Show more",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.primary
                )
            }
        }
    }
}
```

---

## Document Metadata Editor

```kotlin
@Composable
fun DocumentMetadataEditor(
    document: Document,
    onSave: (MetadataUpdate) -> Unit,
    modifier: Modifier = Modifier
) {
    var name by remember { mutableStateOf(document.name) }
    var description by remember { mutableStateOf(document.description) }
    var tags by remember { mutableStateOf(document.tags.joinToString(", ")) }

    LazyColumn(
        modifier = modifier,
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            OutlinedTextField(
                value = name,
                onValueChange = { name = it },
                label = { Text("Document Name") },
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            OutlinedTextField(
                value = description,
                onValueChange = { description = it },
                label = { Text("Description") },
                modifier = Modifier.fillMaxWidth(),
                minLines = 3
            )
        }
        item {
            OutlinedTextField(
                value = tags,
                onValueChange = { tags = it },
                label = { Text("Tags (comma separated)") },
                modifier = Modifier.fillMaxWidth()
            )
        }
        item {
            Button(
                onClick = {
                    onSave(MetadataUpdate(
                        name = name,
                        description = description,
                        tags = tags.split(",").map { it.trim() }.filter { it.isNotEmpty() }
                    ))
                },
                modifier = Modifier.fillMaxWidth()
            ) { Text("Save Changes") }
        }
    }
}

data class MetadataUpdate(
    val name: String,
    val description: String,
    val tags: List<String>
)
```

---

## Camera Capture

```kotlin
@Composable
fun CameraCaptureButton(
    onImageCaptured: (Uri) -> Unit,
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current

    val cameraLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.TakePicturePreview()
    ) { bitmap ->
        bitmap?.let {
            val uri = saveBitmapToCache(context, it)
            onImageCaptured(uri)
        }
    }

    FilledTonalButton(onClick = { cameraLauncher.launch(null) }, modifier = modifier) {
        Icon(Icons.Default.CameraAlt, contentDescription = null, modifier = Modifier.size(16.dp))
        Spacer(Modifier.width(4.dp))
        Text("Scan with Camera")
    }
}

private fun saveBitmapToCache(context: Context, bitmap: Bitmap): Uri {
    val file = File(context.cacheDir, "camera_${System.currentTimeMillis()}.jpg")
    file.outputStream().use { bitmap.compress(Bitmap.CompressFormat.JPEG, 90, it) }
    return FileProvider.getUriForFile(context, "${context.packageName}.fileprovider", file)
}
```

---

## Gallery Pick

```kotlin
@Composable
fun GalleryPickButton(
    onImagesPicked: (List<Uri>) -> Unit,
    modifier: Modifier = Modifier
) {
    val galleryLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetMultipleContents()
    ) { uris ->
        if (uris.isNotEmpty()) onImagesPicked(uris)
    }

    FilledTonalButton(
        onClick = { galleryLauncher.launch("image/*") },
        modifier = modifier
    ) {
        Icon(Icons.Default.PhotoLibrary, contentDescription = null, modifier = Modifier.size(16.dp))
        Spacer(Modifier.width(4.dp))
        Text("Pick from Gallery")
    }
}
```

---

## Document API Integration

```kotlin
interface DocumentApiService {

    @Multipart
    @POST("api/v1/documents/upload")
    suspend fun uploadDocument(
        @Part file: MultipartBody.Part,
        @Part("name") name: RequestBody,
        @Part("description") description: RequestBody? = null,
        @Part("tags") tags: RequestBody? = null
    ): DocumentDto

    @POST("api/v1/documents")
    suspend fun getDocuments(
        @Body request: GetDocumentsRequest
    ): PaginatedResponse<DocumentDto>

    @POST("api/v1/documents/{id}")
    suspend fun getDocument(@Path("id") documentId: String, @Body request: GetDocumentRequest): DocumentDto

    @DELETE("api/v1/documents/{id}")
    suspend fun deleteDocument(@Path("id") documentId: String)

    @PATCH("api/v1/documents/{id}/metadata")
    suspend fun updateMetadata(
        @Path("id") documentId: String,
        @Body request: UpdateMetadataRequest
    ): DocumentDto

    @POST("api/v1/documents/{id}/chunks")
    suspend fun getChunks(
        @Path("id") documentId: String,
        @Body request: GetChunksRequest
    ): PaginatedResponse<DocumentChunkDto>

    @POST("api/v1/documents/{id}/processing-status")
    suspend fun getProcessingStatus(@Path("id") documentId: String, @Body request: GetProcessingStatusRequest): ProcessingStatusDto

    @POST("api/v1/documents/search")
    suspend fun searchDocuments(@Body request: DocumentSearchRequest): List<SearchResultDto>

    // Document Sets
    @POST("api/v1/document-sets")
    suspend fun getSets(@Body request: GetDocumentSetsRequest): List<DocumentSetDto>

    @POST("api/v1/document-sets/create")
    suspend fun createSet(@Body request: CreateDocumentSetRequest): DocumentSetDto

    @DELETE("api/v1/document-sets/{id}")
    suspend fun deleteSet(@Path("id") setId: String)

    @POST("api/v1/document-sets/{id}/documents")
    suspend fun addToSet(
        @Path("id") setId: String,
        @Body request: AddDocumentsToSetRequest
    ): DocumentSetDto

    @DELETE("api/v1/document-sets/{id}/documents/{docId}")
    suspend fun removeFromSet(
        @Path("id") setId: String,
        @Path("docId") documentId: String
    ): DocumentSetDto
}
```

---

## Document Data Models

```kotlin
data class DocumentSummary(
    val documentId: String,
    val name: String,
    val fileType: FileType,
    val fileSizeBytes: Long,
    val formattedSize: String,
    val processingStatus: ProcessingStatus,
    val chunkCount: Int,
    val tags: List<String>,
    val uploadedAt: Instant
)

data class Document(
    val documentId: String,
    val name: String,
    val description: String,
    val fileType: FileType,
    val fileSizeBytes: Long,
    val formattedSize: String,
    val processingStatus: ProcessingStatus,
    val chunkCount: Int,
    val characterCount: Long,
    val wordCount: Long,
    val pageCount: Int?,
    val language: String?,
    val tags: List<String>,
    val sets: List<DocumentSetRef>,
    val uploadedAt: Instant,
    val processedAt: Instant?
)

enum class ProcessingStatus {
    UPLOADED, EXTRACTING, CHUNKING, EMBEDDING, READY, FAILED
}

data class DocumentChunk(
    val chunkId: String,
    val documentId: String,
    val index: Int,
    val text: String,
    val charCount: Int,
    val tokenCount: Int,
    val embedding: List<Float>? = null
)

data class DocumentSetSummary(
    val setId: String,
    val name: String,
    val description: String,
    val documentCount: Int,
    val totalChunks: Int
)

data class DocumentSetDetail(
    val setId: String,
    val name: String,
    val description: String,
    val documents: List<DocumentSummary>,
    val totalChunks: Int
)

data class DocumentSetRef(
    val id: String,
    val name: String
)

data class ProcessingLog(
    val timestamp: Instant,
    val stage: String,
    val message: String,
    val level: LogLevel
)

enum class LogLevel { INFO, WARNING, ERROR }
```

---

## Document Caching

```kotlin
@Entity(tableName = "documents")
data class DocumentCacheEntity(
    @PrimaryKey val documentId: String,
    val name: String,
    val description: String,
    val fileType: String,
    val fileSizeBytes: Long,
    val processingStatus: String,
    val chunkCount: Int,
    val characterCount: Long,
    val wordCount: Long,
    val tagsJson: String,
    val uploadedAt: Long,
    val processedAt: Long?,
    val lastUpdated: Long
)

@Dao
interface DocumentDao {

    @Query("SELECT * FROM documents ORDER BY uploadedAt DESC")
    fun getAllDocuments(): Flow<List<DocumentCacheEntity>>

    @Query("SELECT * FROM documents WHERE documentId = :id")
    suspend fun getDocument(id: String): DocumentCacheEntity?

    @Upsert
    suspend fun upsertAll(documents: List<DocumentCacheEntity>)

    @Upsert
    suspend fun upsert(document: DocumentCacheEntity)

    @Query("DELETE FROM documents WHERE documentId = :id")
    suspend fun delete(id: String)

    @Query("SELECT * FROM documents WHERE name LIKE '%' || :query || '%'")
    suspend fun search(query: String): List<DocumentCacheEntity>

    @Query("UPDATE documents SET processingStatus = :status WHERE documentId = :id")
    suspend fun updateStatus(id: String, status: String)
}

@Entity(tableName = "document_chunks")
data class ChunkCacheEntity(
    @PrimaryKey val chunkId: String,
    val documentId: String,
    val index: Int,
    val text: String,
    val charCount: Int,
    val tokenCount: Int
)

@Dao
interface ChunkDao {

    @Query("SELECT * FROM document_chunks WHERE documentId = :docId ORDER BY `index` ASC")
    suspend fun getChunks(docId: String): List<ChunkCacheEntity>

    @Upsert
    suspend fun upsertAll(chunks: List<ChunkCacheEntity>)

    @Query("DELETE FROM document_chunks WHERE documentId = :docId")
    suspend fun deleteForDocument(docId: String)
}
```

---

## Pull-to-Refresh

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DocumentListPullToRefresh(
    isRefreshing: Boolean,
    onRefresh: () -> Unit,
    content: @Composable () -> Unit
) {
    PullToRefreshBox(
        isRefreshing = isRefreshing,
        onRefresh = onRefresh
    ) { content() }
}
```

---

## Empty States

```kotlin
@Composable
fun DocumentListEmpty() {
    Column(
        modifier = Modifier.fillMaxSize().padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            Icons.Default.CloudUpload,
            contentDescription = null,
            modifier = Modifier.size(80.dp),
            tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.4f)
        )
        Spacer(Modifier.height(16.dp))
        Text("No documents yet", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        Spacer(Modifier.height(8.dp))
        Text(
            "Upload documents to build your knowledge base",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
    }
}
```

---

## Loading States

```kotlin
@Composable
fun DocumentListSkeleton() {
    val shimmer = shimmerBrush()

    LazyColumn(contentPadding = PaddingValues(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        // Header skeleton
        item {
            Row(modifier = Modifier.fillMaxWidth().padding(12.dp)) {
                Box(modifier = Modifier.weight(2f).height(12.dp).background(shimmer, RoundedCornerShape(4.dp)))
                Spacer(Modifier.width(16.dp))
                Box(modifier = Modifier.weight(1f).height(12.dp).background(shimmer, RoundedCornerShape(4.dp)))
            }
        }
        // Row skeletons
        items(8) {
            Card {
                Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(modifier = Modifier.weight(2f)) {
                        Box(modifier = Modifier.fillMaxWidth(0.6f).height(14.dp).background(shimmer, RoundedCornerShape(4.dp)))
                        Spacer(Modifier.height(6.dp))
                        Box(modifier = Modifier.fillMaxWidth(0.3f).height(10.dp).background(shimmer, RoundedCornerShape(4.dp)))
                    }
                    Spacer(Modifier.width(16.dp))
                    Box(modifier = Modifier.weight(1f).height(14.dp).background(shimmer, RoundedCornerShape(4.dp)))
                }
            }
        }
    }
}
```

---

## Accessibility

```kotlin
// Document row
Row(modifier = Modifier.semantics(mergeDescendants = true) {
    contentDescription = "Document ${doc.name}, type ${doc.fileType.name}, ${doc.processingStatus.name}, ${doc.chunkCount} chunks, uploaded ${doc.uploadedAt.toFormattedDate()}"
})

// Processing status
ProcessingStatusBadge(status = status, modifier = Modifier.semantics {
    stateDescription = "Processing status: ${status.name}"
})

// Chunk card
ChunkCard(chunk = chunk, modifier = Modifier.semantics {
    contentDescription = "Chunk ${chunk.index + 1}, ${chunk.charCount} characters"
})

// Camera capture
FilledTonalButton(
    onClick = { ... },
    modifier = Modifier.semantics { role = Role.Button }
) { Text("Scan document with camera") }

// Upload progress
LinearProgressIndicator(
    progress = { ... },
    modifier = Modifier.semantics {
        stateDescription = "Upload progress ${(progress * 100).toInt()} percent"
    }
)
```

---

## Responsive Design

| Width       | Document List           | Upload Screen           | Detail Screen       |
|-------------|-------------------------|-------------------------|---------------------|
| Compact     | Full-width rows         | Full-screen drop zone   | Full-screen tabs    |
| Medium      | Table with columns      | Two-panel: list + drop  | Split view          |
| Expanded    | Table + sidebar preview | Three-column layout     | Side-by-side tabs   |

```kotlin
@Composable
fun AdaptiveDocumentLayout(
    documents: List<DocumentSummary>,
    onDocumentClick: (String) -> Unit
) {
    val windowSize = currentWindowAdaptiveInfo().windowSizeClass

    when (windowSize.windowWidthSizeClass) {
        WindowWidthSizeClass.COMPACT -> {
            DocumentMobileLayout(documents, onDocumentClick)
        }
        WindowWidthSizeClass.MEDIUM -> {
            DocumentTabletLayout(documents, onDocumentClick)
        }
        WindowWidthSizeClass.EXPANDED -> {
            DocumentDesktopLayout(documents, onDocumentClick)
        }
    }
}
```

---

## Summary

| Feature                | Composable / Screen              | ViewModel                  |
|------------------------|----------------------------------|----------------------------|
| Document List          | `DocumentListContent`            | `DocumentListViewModel`    |
| Document Table         | `DocumentTable` / `DocumentRow`  | —                          |
| Upload                 | `DocumentUploadScreen`           | `DocumentUploadViewModel`  |
| Upload Drop Zone       | `UploadDropZone`                 | —                          |
| File Validation        | `FileValidator`                  | —                          |
| Detail                 | `DocumentDetailScreen`           | `DocumentDetailViewModel`  |
| Metadata Editor        | `DocumentMetadataEditor`         | `DocumentDetailViewModel`  |
| Chunk Viewer           | `DocumentChunksTab`              | `DocumentDetailViewModel`  |
| Processing Pipeline    | `ProcessingPipeline`             | —                          |
| Processing Polling     | `ProcessingPollingManager`       | —                          |
| Search                 | `DocumentSearchBar`              | `DocumentListViewModel`    |
| Semantic Search        | `SemanticSearchResults`          | —                          |
| Document Sets          | `DocumentSetListContent`         | `DocumentSetViewModel`     |
| Set Detail             | `DocumentSetDetailScreen`        | `DocumentSetDetailViewModel`|
| Permissions            | `DocumentPermissionsSection`     | —                          |
| Delete Confirmation    | `DocumentDeleteDialog`           | —                          |
| Camera Capture         | `CameraCaptureButton`            | —                          |
| Gallery Pick           | `GalleryPickButton`              | —                          |
| Skeleton               | `DocumentListSkeleton`           | —                          |
| Empty State            | `DocumentListEmpty`              | —                          |
