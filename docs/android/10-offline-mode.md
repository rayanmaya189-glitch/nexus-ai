# 10. Offline Mode

## Table of Contents

1. [Offline Architecture Overview](#1-offline-architecture-overview)
2. [Room Database Configuration](#2-room-database-configuration)
3. [Room Entities](#3-room-entities)
4. [Room DAOs](#4-room-daos)
5. [Room Database](#5-room-database)
6. [DataStore Preferences](#6-datastore-preferences)
7. [EncryptedSharedPreferences](#7-encryptedsharedpreferences)
8. [Offline Cache Strategy](#8-offline-cache-strategy)
9. [Data Synchronization](#9-data-synchronization)
10. [Sync Conflict Resolution](#10-sync-conflict-resolution)
11. [Sync Queue](#11-sync-queue)
12. [Sync Status Indicators](#12-sync-status-indicators)
13. [WorkManager Configuration](#13-workmanager-configuration)
14. [WorkManager Sync Tasks](#14-workmanager-sync-tasks)
15. [Background Sync](#15-background-sync)
16. [Offline Chat](#16-offline-chat)
17. [Offline Document Upload](#17-offline-document-upload)
18. [Offline Search](#18-offline-search)
19. [Offline Data Freshness](#19-offline-data-freshness)
20. [Offline Mode Toggle](#20-offline-mode-toggle)
21. [Offline Storage Limits](#21-offline-storage-limits)
22. [Offline Data Encryption](#22-offline-data-encryption)
23. [Offline Testing](#23-offline-testing)
24. [Offline UX Patterns](#24-offline-ux-patterns)
25. [Offline Error Handling](#25-offline-error-handling)
26. [Offline Performance](#26-offline-performance)
27. [Offline Accessibility](#27-offline-accessibility)

---

## 1. Offline Architecture Overview

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         UI Layer                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Compose  │  │ Compose  │  │ Compose  │  │ Compose  │   │
│  │ Screen A │  │ Screen B │  │ Screen C │  │ Screen D │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
│       │              │              │              │         │
│  ┌────▼──────────────▼──────────────▼──────────────▼─────┐  │
│  │                   ViewModels                          │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐    │  │
│  │  │  VM A   │ │  VM B   │ │  VM C   │ │  VM D   │    │  │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘    │  │
│  └───────┼───────────┼───────────┼───────────┼──────────┘  │
├──────────┼───────────┼───────────┼───────────┼─────────────┤
│          ▼           ▼           ▼           ▼              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Repository Layer                       │   │
│  │  ┌─────────────────────────────────────────────┐    │   │
│  │  │         UnifiedRepository                   │    │   │
│  │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  │    │   │
│  │  │  │Local Data│  │ Sync Mgr │  │Remote API│  │    │   │
│  │  │  │  (Room)  │  │(WorkMgr) │  │ (Retrofit)│  │    │   │
│  │  │  └────┬─────┘  └────┬─────┘  └────┬─────┘  │    │   │
│  │  └───────┼──────────────┼──────────────┼────────┘    │   │
│  └──────────┼──────────────┼──────────────┼─────────────┘   │
├─────────────┼──────────────┼──────────────┼─────────────────┤
│             ▼              ▼              ▼                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐        │
│  │  Room DB     │ │ WorkManager  │ │   Retrofit   │        │
│  │  ┌────────┐  │ │  ┌────────┐  │ │  ┌────────┐  │        │
│  │  │Entities│  │ │  │ Sync   │  │ │  │ API    │  │        │
│  │  │ DAOs   │  │ │  │ Workers│  │ │  │Service │  │        │
│  │  │Migrate │  │ │  │Constraints│ │  │Models  │  │        │
│  │  └────────┘  │ │  └────────┘  │ │  └────────┘  │        │
│  └──────────────┘ └──────────────┘ └──────────────┘        │
│                                                             │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐        │
│  │ DataStore    │ │ EncryptedSP  │ │ Connectivity │        │
│  │ Preferences  │ │ Keystore     │ │  Monitor     │        │
│  └──────────────┘ └──────────────┘ └──────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow Diagram

```
                    ┌─────────────┐
                    │  User Action │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │   Is Online? │
                    └──────┬──────┘
                     Yes   │   No
                ┌──────────┤──────────┐
                ▼                     ▼
        ┌───────────────┐    ┌───────────────┐
        │  Remote API   │    │  Local Cache  │
        │  (Retrofit)   │    │  (Room DB)    │
        └───────┬───────┘    └───────┬───────┘
                │                     │
                ▼                     ▼
        ┌───────────────┐    ┌───────────────┐
        │  Update Cache  │    │  Show Data   │
        │  (Room DB)     │    │  + Stale     │
        └───────┬───────┘    │  Indicator   │
                │             └───────────────┘
                ▼
        ┌───────────────┐
        │  Show to User  │
        │  (Compose UI)  │
        └───────┬───────┘
                │
        ┌───────▼───────┐
        │  Queue Changes │
        │  (Sync Queue)  │
        └───────────────┘
```

### Core Components

| Component | Responsibility | Technology |
|-----------|---------------|------------|
| Local Database | Persistent offline storage | Room |
| Data Store | Lightweight key-value settings | DataStore |
| Secure Storage | Tokens and sensitive data | EncryptedSharedPreferences |
| Sync Manager | Coordinate sync operations | WorkManager |
| Network Monitor | Detect connectivity changes | ConnectivityManager |
| Cache Strategy | Determine data source priority | Repository pattern |
| Conflict Resolver | Handle merge conflicts | Custom resolver |

---

## 2. Room Database Configuration

### build.gradle.kts Dependencies

```kotlin
// build.gradle.kts (app)
plugins {
    id("com.google.devtools.ksp")
}

dependencies {
    // Room
    val roomVersion = "2.6.1"
    implementation("androidx.room:room-runtime:$roomVersion")
    implementation("androidx.room:room-ktx:$roomVersion")
    ksp("androidx.room:room-compiler:$roomVersion")

    // DataStore
    implementation("androidx.datastore:datastore-preferences:1.0.0")

    // WorkManager
    implementation("androidx.work:work-runtime-ktx:2.9.0")
    implementation("androidx.hilt:hilt-work:1.2.0")
    ksp("androidx.hilt:hilt-compiler:1.2.0")

    // Coroutines
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3")
}
```

### Room Database Entity Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Room Database                     │
│                                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌────────────┐ │
│  │ Conversation │  │   Message   │  │  Document  │ │
│  │   Entity     │  │   Entity    │  │   Entity   │ │
│  │              │  │             │  │            │ │
│  │ id (PK)     │  │ id (PK)    │  │ id (PK)   │ │
│  │ title       │  │ content    │  │ name      │ │
│  │ agentId     │  │ convId(FK) │  │ type      │ │
│  │ createdAt   │  │ role       │  │ size      │ │
│  │ updatedAt   │  │ timestamp  │  │ uri       │ │
│  │ syncStatus  │  │ syncStatus │  │ syncStatus│ │
│  │ isDeleted   │  │ isDeleted  │  │ isDeleted │ │
│  └──────┬──────┘  └──────┬─────┘  └─────┬─────┘ │
│         │                │              │         │
│  ┌──────▼──────┐  ┌──────▼─────┐  ┌─────▼─────┐ │
│  │ Agent Entity │  │ UserPref   │  │ SyncQueue │ │
│  │              │  │  Entity    │  │  Entity   │ │
│  │ id (PK)     │  │            │  │           │ │
│  │ name        │  │ key (PK)   │  │ id (PK)  │ │
│  │ model       │  │ value      │  │ operation│ │
│  │ systemPrompt│  │ updatedAt  │  │ entityId │ │
│  │ syncStatus  │  │ syncStatus │  │ entityType│ │
│  └─────────────┘  └────────────┘  │ payload  │ │
│                                    │ retryCnt │ │
│                                    │ status   │ │
│                                    └──────────┘ │
└─────────────────────────────────────────────────────┘
```

---

## 3. Room Entities

### Conversation Entity

```kotlin
import androidx.room.Entity
import androidx.room.PrimaryKey
import androidx.room.ColumnInfo
import androidx.room.Index

@Entity(
    tableName = "conversations",
    indices = [
        Index(value = ["agent_id"]),
        Index(value = ["sync_status"]),
        Index(value = ["updated_at"]),
        Index(value = ["is_deleted"])
    ]
)
data class ConversationEntity(
    @PrimaryKey
    @ColumnInfo(name = "id")
    val id: String,

    @ColumnInfo(name = "title")
    val title: String,

    @ColumnInfo(name = "agent_id")
    val agentId: String,

    @ColumnInfo(name = "created_at")
    val createdAt: Long = System.currentTimeMillis(),

    @ColumnInfo(name = "updated_at")
    val updatedAt: Long = System.currentTimeMillis(),

    @ColumnInfo(name = "sync_status")
    val syncStatus: SyncStatus = SyncStatus.PENDING,

    @ColumnInfo(name = "is_deleted")
    val isDeleted: Boolean = false,

    @ColumnInfo(name = "is_pinned")
    val isPinned: Boolean = false,

    @ColumnInfo(name = "message_count")
    val messageCount: Int = 0,

    @ColumnInfo(name = "last_message_preview")
    val lastMessagePreview: String? = null
)

enum class SyncStatus {
    PENDING,
    SYNCED,
    CONFLICT,
    FAILED
}
```

### Message Entity

```kotlin
@Entity(
    tableName = "messages",
    indices = [
        Index(value = ["conversation_id"]),
        Index(value = ["sync_status"]),
        Index(value = ["timestamp"]),
        Index(value = ["role"])
    ],
    foreignKeys = [
        ForeignKey(
            entity = ConversationEntity::class,
            parentColumns = ["id"],
            childColumns = ["conversation_id"],
            onDelete = ForeignKey.CASCADE
        )
    ]
)
data class MessageEntity(
    @PrimaryKey
    @ColumnInfo(name = "id")
    val id: String,

    @ColumnInfo(name = "conversation_id")
    val conversationId: String,

    @ColumnInfo(name = "role")
    val role: MessageRole,

    @ColumnInfo(name = "content")
    val content: String,

    @ColumnInfo(name = "timestamp")
    val timestamp: Long = System.currentTimeMillis(),

    @ColumnInfo(name = "sync_status")
    val syncStatus: SyncStatus = SyncStatus.PENDING,

    @ColumnInfo(name = "is_deleted")
    val isDeleted: Boolean = false,

    @ColumnInfo(name = "attachments")
    val attachments: String? = null, // JSON array of attachment URIs

    @ColumnInfo(name = "metadata")
    val metadata: String? = null, // JSON object for extra data

    @ColumnInfo(name = "parent_message_id")
    val parentMessageId: String? = null,

    @ColumnInfo(name = "version")
    val version: Int = 1 // For conflict resolution
)

enum class MessageRole {
    USER,
    ASSISTANT,
    SYSTEM
}
```

### Document Entity

```kotlin
@Entity(
    tableName = "documents",
    indices = [
        Index(value = ["user_id"]),
        Index(value = ["sync_status"]),
        Index(value = ["processing_status"]),
        Index(value = ["mime_type"])
    ]
)
data class DocumentEntity(
    @PrimaryKey
    @ColumnInfo(name = "id")
    val id: String,

    @ColumnInfo(name = "user_id")
    val userId: String,

    @ColumnInfo(name = "name")
    val name: String,

    @ColumnInfo(name = "mime_type")
    val mimeType: String,

    @ColumnInfo(name = "size_bytes")
    val sizeBytes: Long,

    @ColumnInfo(name = "local_uri")
    val localUri: String? = null, // Local file path

    @ColumnInfo(name = "remote_url")
    val remoteUrl: String? = null, // Server URL

    @ColumnInfo(name = "checksum")
    val checksum: String? = null, // SHA-256 for integrity

    @ColumnInfo(name = "processing_status")
    val processingStatus: ProcessingStatus = ProcessingStatus.PENDING,

    @ColumnInfo(name = "chunk_count")
    val chunkCount: Int = 0,

    @ColumnInfo(name = "sync_status")
    val syncStatus: SyncStatus = SyncStatus.PENDING,

    @ColumnInfo(name = "created_at")
    val createdAt: Long = System.currentTimeMillis(),

    @ColumnInfo(name = "updated_at")
    val updatedAt: Long = System.currentTimeMillis(),

    @ColumnInfo(name = "is_deleted")
    val isDeleted: Boolean = false,

    @ColumnInfo(name = "tags")
    val tags: String? = null // JSON array of tags
)

enum class ProcessingStatus {
    PENDING,
    PROCESSING,
    COMPLETED,
    FAILED
}
```

### Agent Entity

```kotlin
@Entity(
    tableName = "agents",
    indices = [
        Index(value = ["user_id"]),
        Index(value = ["sync_status"]),
        Index(value = ["is_active"])
    ]
)
data class AgentEntity(
    @PrimaryKey
    @ColumnInfo(name = "id")
    val id: String,

    @ColumnInfo(name = "user_id")
    val userId: String,

    @ColumnInfo(name = "name")
    val name: String,

    @ColumnInfo(name = "description")
    val description: String? = null,

    @ColumnInfo(name = "model")
    val model: String,

    @ColumnInfo(name = "system_prompt")
    val systemPrompt: String? = null,

    @ColumnInfo(name = "avatar_url")
    val avatarUrl: String? = null,

    @ColumnInfo(name = "is_active")
    val isActive: Boolean = true,

    @ColumnInfo(name = "config_json")
    val configJson: String? = null, // Agent configuration JSON

    @ColumnInfo(name = "sync_status")
    val syncStatus: SyncStatus = SyncStatus.PENDING,

    @ColumnInfo(name = "created_at")
    val createdAt: Long = System.currentTimeMillis(),

    @ColumnInfo(name = "updated_at")
    val updatedAt: Long = System.currentTimeMillis(),

    @ColumnInfo(name = "is_deleted")
    val isDeleted: Boolean = false
)
```

### User Preferences Entity

```kotlin
@Entity(
    tableName = "user_preferences",
    indices = [
        Index(value = ["user_id"]),
        Index(value = ["key"], unique = true)
    ]
)
data class UserPreferenceEntity(
    @PrimaryKey(autoGenerate = true)
    val uid: Long = 0,

    @ColumnInfo(name = "user_id")
    val userId: String,

    @ColumnInfo(name = "key")
    val key: String,

    @ColumnInfo(name = "value")
    val value: String,

    @ColumnInfo(name = "value_type")
    val valueType: String = "string", // string, int, boolean, json

    @ColumnInfo(name = "sync_status")
    val syncStatus: SyncStatus = SyncStatus.PENDING,

    @ColumnInfo(name = "updated_at")
    val updatedAt: Long = System.currentTimeMillis()
)
```

### Sync Queue Entity

```kotlin
@Entity(
    tableName = "sync_queue",
    indices = [
        Index(value = ["status"]),
        Index(value = ["entity_type"]),
        Index(value = ["created_at"]),
        Index(value = ["next_retry_at"])
    ]
)
data class SyncQueueEntity(
    @PrimaryKey(autoGenerate = true)
    val uid: Long = 0,

    @ColumnInfo(name = "operation_id")
    val operationId: String = UUID.randomUUID().toString(),

    @ColumnInfo(name = "entity_type")
    val entityType: EntityType,

    @ColumnInfo(name = "entity_id")
    val entityId: String,

    @ColumnInfo(name = "operation")
    val operation: SyncOperation,

    @ColumnInfo(name = "payload")
    val payload: String, // JSON serialized entity data

    @ColumnInfo(name = "status")
    val status: SyncQueueStatus = SyncQueueStatus.PENDING,

    @ColumnInfo(name = "retry_count")
    val retryCount: Int = 0,

    @ColumnInfo(name = "max_retries")
    val maxRetries: Int = 5,

    @ColumnInfo(name = "created_at")
    val createdAt: Long = System.currentTimeMillis(),

    @ColumnInfo(name = "next_retry_at")
    val nextRetryAt: Long = System.currentTimeMillis(),

    @ColumnInfo(name = "error_message")
    val errorMessage: String? = null,

    @ColumnInfo(name = "checksum")
    val checksum: String? = null
)

enum class EntityType {
    CONVERSATION,
    MESSAGE,
    DOCUMENT,
    AGENT,
    PREFERENCE
}

enum class SyncOperation {
    CREATE,
    UPDATE,
    DELETE
}

enum class SyncQueueStatus {
    PENDING,
    IN_PROGRESS,
    COMPLETED,
    FAILED,
    CANCELLED
}
```

### Entity Relationship Diagram

```
┌──────────────────┐       ┌──────────────────┐
│  Conversation    │       │     Message       │
├──────────────────┤       ├──────────────────┤
│ id (PK)          │──┐    │ id (PK)          │
│ title            │  │    │ conversation_id   │──── (FK → Conversation.id)
│ agent_id         │  └───▶│ role              │
│ created_at       │       │ content           │
│ updated_at       │       │ timestamp         │
│ sync_status      │       │ sync_status       │
│ is_deleted       │       │ version           │
│ is_pinned        │       │ parent_message_id │──┐
│ message_count    │       └──────────────────┘  │
│ last_message_preview│                           │
└──────────────────┘                   (self-ref) │
                                                  │
┌──────────────────┐       ┌──────────────────┐   │
│    Document      │       │      Agent       │   │
├──────────────────┤       ├──────────────────┤   │
│ id (PK)          │       │ id (PK)          │   │
│ user_id          │       │ user_id          │   │
│ name             │       │ name             │   │
│ mime_type        │       │ model            │   │
│ local_uri        │       │ system_prompt    │   │
│ remote_url       │       │ sync_status      │   │
│ processing_status│       └──────────────────┘   │
│ sync_status      │                              │
└──────────────────┘       ┌──────────────────┐   │
                           │   SyncQueue      │   │
┌──────────────────┐       ├──────────────────┤   │
│  UserPreference  │       │ uid (PK)         │   │
├──────────────────┤       │ entity_type      │   │
│ uid (PK)         │       │ entity_id        │◀──┘
│ user_id          │       │ operation        │
│ key              │       │ payload          │
│ value            │       │ status           │
│ sync_status      │       │ retry_count      │
└──────────────────┘       └──────────────────┘
```

---

## 4. Room DAOs

### Conversation DAO

```kotlin
@Dao
interface ConversationDao {

    // ─── Query Operations ───────────────────────────────────

    @Query("SELECT * FROM conversations WHERE is_deleted = 0 ORDER BY updated_at DESC")
    fun getAllConversations(): Flow<List<ConversationEntity>>

    @Query("SELECT * FROM conversations WHERE id = :id AND is_deleted = 0")
    fun getConversationById(id: String): Flow<ConversationEntity?>

    @Query("SELECT * FROM conversations WHERE id = :id AND is_deleted = 0")
    suspend fun getConversationByIdOnce(id: String): ConversationEntity?

    @Query("""
        SELECT * FROM conversations 
        WHERE is_deleted = 0 
        AND (title LIKE '%' || :query || '%' OR last_message_preview LIKE '%' || :query || '%')
        ORDER BY updated_at DESC
    """)
    fun searchConversations(query: String): Flow<List<ConversationEntity>>

    @Query("SELECT * FROM conversations WHERE agent_id = :agentId AND is_deleted = 0 ORDER BY updated_at DESC")
    fun getConversationsByAgent(agentId: String): Flow<List<ConversationEntity>>

    @Query("SELECT * FROM conversations WHERE sync_status != 'SYNCED' AND is_deleted = 0")
    suspend fun getUnsyncedConversations(): List<ConversationEntity>

    @Query("SELECT * FROM conversations WHERE is_pinned = 1 AND is_deleted = 0 ORDER BY updated_at DESC")
    fun getPinnedConversations(): Flow<List<ConversationEntity>>

    @Query("SELECT COUNT(*) FROM conversations WHERE is_deleted = 0")
    fun getConversationCount(): Flow<Int>

    @Query("""
        SELECT * FROM conversations 
        WHERE is_deleted = 0 
        ORDER BY updated_at DESC 
        LIMIT :limit OFFSET :offset
    """)
    fun getConversationsPaged(limit: Int, offset: Int): Flow<List<ConversationEntity>>

    // ─── Insert Operations ──────────────────────────────────

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertConversation(conversation: ConversationEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertConversations(conversations: List<ConversationEntity>)

    // ─── Update Operations ──────────────────────────────────

    @Update
    suspend fun updateConversation(conversation: ConversationEntity)

    @Query("UPDATE conversations SET sync_status = :status WHERE id = :id")
    suspend fun updateSyncStatus(id: String, status: SyncStatus)

    @Query("UPDATE conversations SET is_pinned = :isPinned WHERE id = :id")
    suspend fun updatePinnedStatus(id: String, isPinned: Boolean)

    @Query("UPDATE conversations SET message_count = :count, last_message_preview = :preview, updated_at = :timestamp WHERE id = :id")
    suspend fun updateMessagePreview(id: String, count: Int, preview: String, timestamp: Long)

    @Query("UPDATE conversations SET sync_status = 'SYNCED' WHERE id = :id")
    suspend fun markAsSynced(id: String)

    // ─── Delete Operations ──────────────────────────────────

    @Query("UPDATE conversations SET is_deleted = 1, sync_status = 'PENDING' WHERE id = :id")
    suspend fun softDelete(id: String)

    @Query("DELETE FROM conversations WHERE id = :id")
    suspend fun hardDelete(id: String)

    @Query("DELETE FROM conversations WHERE is_deleted = 1")
    suspend fun purgeDeletedConversations()

    // ─── Upsert (Insert or Update) ─────────────────────────

    @Transaction
    suspend fun upsert(conversation: ConversationEntity) {
        val existing = getConversationByIdOnce(conversation.id)
        if (existing != null) {
            updateConversation(conversation)
        } else {
            insertConversation(conversation)
        }
    }
}
```

### Message DAO

```kotlin
@Dao
interface MessageDao {

    @Query("SELECT * FROM messages WHERE conversation_id = :conversationId AND is_deleted = 0 ORDER BY timestamp ASC")
    fun getMessagesByConversation(conversationId: String): Flow<List<MessageEntity>>

    @Query("SELECT * FROM messages WHERE id = :id AND is_deleted = 0")
    suspend fun getMessageById(id: String): MessageEntity?

    @Query("SELECT * FROM messages WHERE conversation_id = :conversationId AND is_deleted = 0 ORDER BY timestamp DESC LIMIT :limit")
    fun getRecentMessages(conversationId: String, limit: Int = 50): Flow<List<MessageEntity>>

    @Query("""
        SELECT * FROM messages 
        WHERE is_deleted = 0 
        AND content LIKE '%' || :query || '%'
        ORDER BY timestamp DESC
    """)
    fun searchMessages(query: String): Flow<List<MessageEntity>>

    @Query("SELECT * FROM messages WHERE sync_status != 'SYNCED' AND is_deleted = 0")
    suspend fun getUnsyncedMessages(): List<MessageEntity>

    @Query("SELECT * FROM messages WHERE conversation_id = :conversationId AND is_deleted = 0 ORDER BY timestamp ASC LIMIT :limit OFFSET :offset")
    fun getMessagesPaged(conversationId: String, limit: Int, offset: Int): Flow<List<MessageEntity>>

    @Query("""
        SELECT * FROM messages 
        WHERE conversation_id = :conversationId 
        AND role = 'ASSISTANT' 
        AND is_deleted = 0 
        ORDER BY timestamp DESC 
        LIMIT 1
    """)
    suspend fun getLastAssistantMessage(conversationId: String): MessageEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertMessage(message: MessageEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertMessages(messages: List<MessageEntity>)

    @Update
    suspend fun updateMessage(message: MessageEntity)

    @Query("UPDATE messages SET sync_status = :status WHERE id = :id")
    suspend fun updateSyncStatus(id: String, status: SyncStatus)

    @Query("UPDATE messages SET content = :content, version = version + 1, sync_status = 'PENDING' WHERE id = :id")
    suspend fun updateContent(id: String, content: String)

    @Query("UPDATE messages SET sync_status = 'SYNCED' WHERE id = :id")
    suspend fun markAsSynced(id: String)

    @Transaction
    suspend fun upsert(message: MessageEntity) {
        val existing = getMessageById(message.id)
        if (existing != null) {
            updateMessage(message)
        } else {
            insertMessage(message)
        }
    }

    @Query("UPDATE messages SET is_deleted = 1, sync_status = 'PENDING' WHERE id = :id")
    suspend fun softDelete(id: String)

    @Query("DELETE FROM messages WHERE id = :id")
    suspend fun hardDelete(id: String)

    @Query("DELETE FROM messages WHERE conversation_id = :conversationId")
    suspend fun deleteAllByConversation(conversationId: String)

    @Query("SELECT COUNT(*) FROM messages WHERE conversation_id = :conversationId AND is_deleted = 0")
    suspend fun getMessageCount(conversationId: String): Int
}
```

### Document DAO

```kotlin
@Dao
interface DocumentDao {

    @Query("SELECT * FROM documents WHERE user_id = :userId AND is_deleted = 0 ORDER BY created_at DESC")
    fun getDocumentsByUser(userId: String): Flow<List<DocumentEntity>>

    @Query("SELECT * FROM documents WHERE id = :id AND is_deleted = 0")
    suspend fun getDocumentById(id: String): DocumentEntity?

    @Query("SELECT * FROM documents WHERE id = :id AND is_deleted = 0")
    fun getDocumentByIdFlow(id: String): Flow<DocumentEntity?>

    @Query("""
        SELECT * FROM documents 
        WHERE user_id = :userId 
        AND is_deleted = 0 
        AND (name LIKE '%' || :query || '%' OR tags LIKE '%' || :query || '%')
        ORDER BY created_at DESC
    """)
    fun searchDocuments(userId: String, query: String): Flow<List<DocumentEntity>>

    @Query("SELECT * FROM documents WHERE processing_status = :status AND is_deleted = 0")
    suspend fun getDocumentsByProcessingStatus(status: ProcessingStatus): List<DocumentEntity>

    @Query("SELECT * FROM documents WHERE sync_status != 'SYNCED' AND is_deleted = 0")
    suspend fun getUnsyncedDocuments(): List<DocumentEntity>

    @Query("SELECT SUM(size_bytes) FROM documents WHERE user_id = :userId AND is_deleted = 0")
    suspend fun getTotalStorageUsed(userId: String): Long?

    @Query("SELECT * FROM documents WHERE user_id = :userId AND mime_type = :mimeType AND is_deleted = 0")
    fun getDocumentsByType(userId: String, mimeType: String): Flow<List<DocumentEntity>>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertDocument(document: DocumentEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertDocuments(documents: List<DocumentEntity>)

    @Update
    suspend fun updateDocument(document: DocumentEntity)

    @Query("UPDATE documents SET processing_status = :status WHERE id = :id")
    suspend fun updateProcessingStatus(id: String, status: ProcessingStatus)

    @Query("UPDATE documents SET sync_status = :status WHERE id = :id")
    suspend fun updateSyncStatus(id: String, status: SyncStatus)

    @Query("UPDATE documents SET remote_url = :url, sync_status = 'SYNCED' WHERE id = :id")
    suspend fun markAsUploaded(id: String, url: String)

    @Query("UPDATE documents SET is_deleted = 1, sync_status = 'PENDING' WHERE id = :id")
    suspend fun softDelete(id: String)

    @Query("DELETE FROM documents WHERE id = :id")
    suspend fun hardDelete(id: String)

    @Transaction
    suspend fun upsert(document: DocumentEntity) {
        val existing = getDocumentById(document.id)
        if (existing != null) {
            updateDocument(document)
        } else {
            insertDocument(document)
        }
    }
}
```

### SyncQueue DAO

```kotlin
@Dao
interface SyncQueueDao {

    @Query("SELECT * FROM sync_queue WHERE status = 'PENDING' OR status = 'IN_PROGRESS' ORDER BY created_at ASC")
    fun getPendingOperations(): Flow<List<SyncQueueEntity>>

    @Query("SELECT * FROM sync_queue WHERE status = 'PENDING' AND next_retry_at <= :currentTime ORDER BY created_at ASC LIMIT :limit")
    suspend fun getReadyOperations(currentTime: Long = System.currentTimeMillis(), limit: Int = 10): List<SyncQueueEntity>

    @Query("SELECT * FROM sync_queue WHERE status = 'FAILED' AND retry_count < max_retries AND next_retry_at <= :currentTime")
    suspend fun getRetryableOperations(currentTime: Long = System.currentTimeMillis()): List<SyncQueueEntity>

    @Query("SELECT COUNT(*) FROM sync_queue WHERE status IN ('PENDING', 'IN_PROGRESS')")
    fun getPendingCount(): Flow<Int>

    @Query("SELECT * FROM sync_queue WHERE entity_type = :entityType AND entity_id = :entityId AND status IN ('PENDING', 'IN_PROGRESS')")
    suspend fun getExistingOperation(entityType: EntityType, entityId: String): SyncQueueEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insert(operation: SyncQueueEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAll(operations: List<SyncQueueEntity>)

    @Update
    suspend fun update(operation: SyncQueueEntity)

    @Query("UPDATE sync_queue SET status = :status WHERE uid = :uid")
    suspend fun updateStatus(uid: Long, status: SyncQueueStatus)

    @Query("UPDATE sync_queue SET status = 'IN_PROGRESS' WHERE uid = :uid")
    suspend fun markInProgress(uid: Long)

    @Query("UPDATE sync_queue SET status = 'COMPLETED' WHERE uid = :uid")
    suspend fun markCompleted(uid: Long)

    @Query("UPDATE sync_queue SET status = 'FAILED', retry_count = retry_count + 1, error_message = :error, next_retry_at = :nextRetryAt WHERE uid = :uid")
    suspend fun markFailed(uid: Long, error: String, nextRetryAt: Long)

    @Query("DELETE FROM sync_queue WHERE status = 'COMPLETED'")
    suspend fun purgeCompleted()

    @Query("DELETE FROM sync_queue WHERE status = 'FAILED' AND retry_count >= max_retries")
    suspend fun purgeExhaustedRetries()

    @Query("DELETE FROM sync_queue")
    suspend fun purgeAll()

    @Transaction
    suspend fun enqueue(
        entityType: EntityType,
        entityId: String,
        operation: SyncOperation,
        payload: String
    ) {
        val existing = getExistingOperation(entityType, entityId)
        if (existing != null) {
            // Coalesce: merge operations
            val mergedOperation = coalesceOperations(existing.operation, operation)
            update(existing.copy(
                operation = mergedOperation,
                payload = payload,
                status = SyncQueueStatus.PENDING,
                retryCount = 0,
                errorMessage = null,
                nextRetryAt = System.currentTimeMillis()
            ))
        } else {
            insert(SyncQueueEntity(
                entityType = entityType,
                entityId = entityId,
                operation = operation,
                payload = payload
            ))
        }
    }

    private fun coalesceOperations(existing: SyncOperation, new: SyncOperation): SyncOperation {
        return when {
            existing == SyncOperation.CREATE && new == SyncOperation.UPDATE -> SyncOperation.CREATE
            existing == SyncOperation.CREATE && new == SyncOperation.DELETE -> null // Remove entirely
            existing == SyncOperation.UPDATE && new == SyncOperation.DELETE -> SyncOperation.DELETE
            else -> new
        } ?: existing
    }
}
```

### Agent DAO

```kotlin
@Dao
interface AgentDao {

    @Query("SELECT * FROM agents WHERE user_id = :userId AND is_deleted = 0 ORDER BY name ASC")
    fun getAgentsByUser(userId: String): Flow<List<AgentEntity>>

    @Query("SELECT * FROM agents WHERE id = :id AND is_deleted = 0")
    suspend fun getAgentById(id: String): AgentEntity?

    @Query("SELECT * FROM agents WHERE id = :id AND is_deleted = 0")
    fun getAgentByIdFlow(id: String): Flow<AgentEntity?>

    @Query("SELECT * FROM agents WHERE is_active = 1 AND is_deleted = 0")
    fun getActiveAgents(): Flow<List<AgentEntity>>

    @Query("SELECT * FROM agents WHERE sync_status != 'SYNCED' AND is_deleted = 0")
    suspend fun getUnsyncedAgents(): List<AgentEntity>

    @Query("""
        SELECT * FROM agents 
        WHERE user_id = :userId 
        AND is_deleted = 0 
        AND (name LIKE '%' || :query || '%' OR description LIKE '%' || :query || '%')
    """)
    fun searchAgents(userId: String, query: String): Flow<List<AgentEntity>>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAgent(agent: AgentEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAgents(agents: List<AgentEntity>)

    @Update
    suspend fun updateAgent(agent: AgentEntity)

    @Query("UPDATE agents SET is_active = :isActive WHERE id = :id")
    suspend fun updateActiveStatus(id: String, isActive: Boolean)

    @Query("UPDATE agents SET sync_status = :status WHERE id = :id")
    suspend fun updateSyncStatus(id: String, status: SyncStatus)

    @Query("UPDATE agents SET is_deleted = 1, sync_status = 'PENDING' WHERE id = :id")
    suspend fun softDelete(id: String)

    @Query("DELETE FROM agents WHERE id = :id")
    suspend fun hardDelete(id: String)

    @Transaction
    suspend fun upsert(agent: AgentEntity) {
        val existing = getAgentById(agent.id)
        if (existing != null) {
            updateAgent(agent)
        } else {
            insertAgent(agent)
        }
    }
}
```

---

## 5. Room Database

### Database Configuration

```kotlin
@Database(
    entities = [
        ConversationEntity::class,
        MessageEntity::class,
        DocumentEntity::class,
        AgentEntity::class,
        UserPreferenceEntity::class,
        SyncQueueEntity::class
    ],
    version = 1,
    exportSchema = true
)
@TypeConverters(Converters::class)
abstract class NexusDatabase : RoomDatabase() {

    abstract fun conversationDao(): ConversationDao
    abstract fun messageDao(): MessageDao
    abstract fun documentDao(): DocumentDao
    abstract fun agentDao(): AgentDao
    abstract fun userPreferenceDao(): UserPreferenceDao
    abstract fun syncQueueDao(): SyncQueueDao

    companion object {
        const val DATABASE_NAME = "nexus_ai_database"
    }
}
```

### Type Converters

```kotlin
class Converters {

    @TypeConverter
    fun fromSyncStatus(status: SyncStatus): String = status.name

    @TypeConverter
    fun toSyncStatus(value: String): SyncStatus = SyncStatus.valueOf(value)

    @TypeConverter
    fun fromMessageRole(role: MessageRole): String = role.name

    @TypeConverter
    fun toMessageRole(value: String): MessageRole = MessageRole.valueOf(value)

    @TypeConverter
    fun fromProcessingStatus(status: ProcessingStatus): String = status.name

    @TypeConverter
    fun toProcessingStatus(value: String): ProcessingStatus = ProcessingStatus.valueOf(value)

    @TypeConverter
    fun fromEntityType(type: EntityType): String = type.name

    @TypeConverter
    fun toEntityType(value: String): EntityType = EntityType.valueOf(value)

    @TypeConverter
    fun fromSyncOperation(op: SyncOperation): String = op.name

    @TypeConverter
    fun toSyncOperation(value: String): SyncOperation = SyncOperation.valueOf(value)

    @TypeConverter
    fun fromSyncQueueStatus(status: SyncQueueStatus): String = status.name

    @TypeConverter
    fun toSyncQueueStatus(value: String): SyncQueueStatus = SyncQueueStatus.valueOf(value)
}
```

### Database Module (Hilt)

```kotlin
@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideNexusDatabase(
        @ApplicationContext context: Context
    ): NexusDatabase {
        return Room.databaseBuilder(
            context,
            NexusDatabase::class.java,
            NexusDatabase.DATABASE_NAME
        )
        .addMigrations(*allMigrations)
        .fallbackToDestructiveMigrationOnDowngrade()
        .build()
    }

    @Provides
    fun provideConversationDao(db: NexusDatabase): ConversationDao =
        db.conversationDao()

    @Provides
    fun provideMessageDao(db: NexusDatabase): MessageDao =
        db.messageDao()

    @Provides
    fun provideDocumentDao(db: NexusDatabase): DocumentDao =
        db.documentDao()

    @Provides
    fun provideAgentDao(db: NexusDatabase): AgentDao =
        db.agentDao()

    @Provides
    fun provideSyncQueueDao(db: NexusDatabase): SyncQueueDao =
        db.syncQueueDao()

    @Provides
    fun provideUserPreferenceDao(db: NexusDatabase): UserPreferenceDao =
        db.userPreferenceDao()
}
```

### Migrations

```kotlin
val MIGRATION_1_2 = object : Migration(1, 2) {
    override fun migrate(db: SupportSQLiteDatabase) {
        db.execSQL("""
            ALTER TABLE conversations 
            ADD COLUMN is_pinned INTEGER NOT NULL DEFAULT 0
        """)
        db.execSQL("""
            ALTER TABLE messages 
            ADD COLUMN parent_message_id TEXT
        """)
    }
}

val MIGRATION_2_3 = object : Migration(2, 3) {
    override fun migrate(db: SupportSQLiteDatabase) {
        db.execSQL("""
            CREATE TABLE IF NOT EXISTS sync_queue (
                uid INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
                operation_id TEXT NOT NULL,
                entity_type TEXT NOT NULL,
                entity_id TEXT NOT NULL,
                operation TEXT NOT NULL,
                payload TEXT NOT NULL,
                status TEXT NOT NULL DEFAULT 'PENDING',
                retry_count INTEGER NOT NULL DEFAULT 0,
                max_retries INTEGER NOT NULL DEFAULT 5,
                created_at INTEGER NOT NULL,
                next_retry_at INTEGER NOT NULL,
                error_message TEXT,
                checksum TEXT
            )
        """)
        db.execSQL("CREATE INDEX IF NOT EXISTS index_sync_queue_status ON sync_queue(status)")
        db.execSQL("CREATE INDEX IF NOT EXISTS index_sync_queue_entity_type ON sync_queue(entity_type)")
        db.execSQL("CREATE INDEX IF NOT EXISTS index_sync_queue_next_retry_at ON sync_queue(next_retry_at)")
    }
}

val MIGRATION_3_4 = object : Migration(3, 4) {
    override fun migrate(db: SupportSQLiteDatabase) {
        db.execSQL("""
            ALTER TABLE documents 
            ADD COLUMN tags TEXT
        """)
        db.execSQL("""
            ALTER TABLE documents 
            ADD COLUMN checksum TEXT
        """)
        db.execSQL("CREATE INDEX IF NOT EXISTS index_documents_mime_type ON documents(mime_type)")
    }
}

val allMigrations = arrayOf(
    MIGRATION_1_2,
    MIGRATION_2_3,
    MIGRATION_3_4
)
```

### Migration Strategy Diagram

```
┌─────────────────────────────────────────────────┐
│              Migration Strategy                  │
│                                                  │
│  App Update Detected                             │
│       │                                          │
│       ▼                                          │
│  ┌──────────────┐                                │
│  │ Old Version  │                                │
│  │ (e.g., v2)   │                                │
│  └──────┬───────┘                                │
│         │                                        │
│         ▼                                        │
│  ┌──────────────┐    Yes    ┌──────────────┐    │
│  │ Migration    │──────────▶│ Apply All    │    │
│  │ Path Exists? │           │ Migrations   │    │
│  └──────┬───────┘           │ 2→3, 3→4     │    │
│         │ No                └──────┬───────┘    │
│         ▼                          │             │
│  ┌──────────────┐                  ▼             │
│  │ Destructive  │         ┌──────────────┐      │
│  │ Migration    │         │  Re-open DB   │      │
│  │ (fallback)   │         │  on new ver   │      │
│  └──────────────┘         └──────────────┘      │
└─────────────────────────────────────────────────┘
```

---

## 6. DataStore Preferences

### UserSettings DataStore

```kotlin
object UserSettingsKeys {
    // Theme
    val THEME_MODE = stringPreferencesKey("theme_mode")
    val DYNAMIC_COLOR = booleanPreferencesKey("dynamic_color")

    // Language
    val LANGUAGE = stringPreferencesKey("language")
    val LOCALE = stringPreferencesKey("locale")

    // Chat
    val DEFAULT_AGENT_ID = stringPreferencesKey("default_agent_id")
    val SEND_ON_ENTER = booleanPreferencesKey("send_on_enter")
    val STREAMING_ENABLED = booleanPreferencesKey("streaming_enabled")
    val MESSAGE_FONT_SIZE = floatPreferencesKey("message_font_size")

    // Notifications
    val NOTIFICATIONS_ENABLED = booleanPreferencesKey("notifications_enabled")
    val SOUND_ENABLED = booleanPreferencesKey("sound_enabled")
    val VIBRATION_ENABLED = booleanPreferencesKey("vibration_enabled")

    // Offline
    val OFFLINE_MODE_ENABLED = booleanPreferencesKey("offline_mode_enabled")
    val AUTO_SYNC = booleanPreferencesKey("auto_sync")
    val SYNC_INTERVAL_MINUTES = intPreferencesKey("sync_interval_minutes")
    val WIFI_ONLY_SYNC = booleanPreferencesKey("wifi_only_sync")

    // Privacy
    val ANALYTICS_ENABLED = booleanPreferencesKey("analytics_enabled")
    val CRASH_REPORTING = booleanPreferencesKey("crash_reporting")

    // Storage
    val MAX_CACHE_SIZE_MB = intPreferencesKey("max_cache_size_mb")
    val AUTO_DELETE_OLD_CONVERSATIONS = booleanPreferencesKey("auto_delete_old_conversations")
    val RETENTION_DAYS = intPreferencesKey("retention_days")

    // Accessibility
    val HIGH_CONTRAST = booleanPreferencesKey("high_contrast")
    val REDUCE_MOTION = booleanPreferencesKey("reduce_motion")
    val SCREEN_READER_OPTIMIZED = booleanPreferencesKey("screen_reader_optimized")
}

data class UserSettings(
    val themeMode: String = "system",
    val dynamicColor: Boolean = true,
    val language: String = "en",
    val defaultAgentId: String? = null,
    val sendOnEnter: Boolean = true,
    val streamingEnabled: Boolean = true,
    val messageFontSize: Float = 16f,
    val notificationsEnabled: Boolean = true,
    val soundEnabled: Boolean = true,
    val vibrationEnabled: Boolean = true,
    val offlineModeEnabled: Boolean = false,
    val autoSync: Boolean = true,
    val syncIntervalMinutes: Int = 15,
    val wifiOnlySync: Boolean = false,
    val analyticsEnabled: Boolean = false,
    val crashReporting: Boolean = true,
    val maxCacheSizeMb: Int = 500,
    val autoDeleteOldConversations: Boolean = false,
    val retentionDays: Int = 90,
    val highContrast: Boolean = false,
    val reduceMotion: Boolean = false,
    val screenReaderOptimized: Boolean = false
)

class UserSettingsRepository @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val Context.dataStore by preferencesDataStore(
        name = "user_settings"
    )

    val settings: Flow<UserSettings> = context.dataStore.data
        .catch { exception ->
            if (exception is IOException) {
                emit(emptyPreferences())
            } else {
                throw exception
            }
        }
        .map { preferences ->
            UserSettings(
                themeMode = preferences[UserSettingsKeys.THEME_MODE] ?: "system",
                dynamicColor = preferences[UserSettingsKeys.DYNAMIC_COLOR] ?: true,
                language = preferences[UserSettingsKeys.LANGUAGE] ?: "en",
                defaultAgentId = preferences[UserSettingsKeys.DEFAULT_AGENT_ID],
                sendOnEnter = preferences[UserSettingsKeys.SEND_ON_ENTER] ?: true,
                streamingEnabled = preferences[UserSettingsKeys.STREAMING_ENABLED] ?: true,
                messageFontSize = preferences[UserSettingsKeys.MESSAGE_FONT_SIZE] ?: 16f,
                notificationsEnabled = preferences[UserSettingsKeys.NOTIFICATIONS_ENABLED] ?: true,
                soundEnabled = preferences[UserSettingsKeys.SOUND_ENABLED] ?: true,
                vibrationEnabled = preferences[UserSettingsKeys.VIBRATION_ENABLED] ?: true,
                offlineModeEnabled = preferences[UserSettingsKeys.OFFLINE_MODE_ENABLED] ?: false,
                autoSync = preferences[UserSettingsKeys.AUTO_SYNC] ?: true,
                syncIntervalMinutes = preferences[UserSettingsKeys.SYNC_INTERVAL_MINUTES] ?: 15,
                wifiOnlySync = preferences[UserSettingsKeys.WIFI_ONLY_SYNC] ?: false,
                analyticsEnabled = preferences[UserSettingsKeys.ANALYTICS_ENABLED] ?: false,
                crashReporting = preferences[UserSettingsKeys.CRASH_REPORTING] ?: true,
                maxCacheSizeMb = preferences[UserSettingsKeys.MAX_CACHE_SIZE_MB] ?: 500,
                autoDeleteOldConversations = preferences[UserSettingsKeys.AUTO_DELETE_OLD_CONVERSATIONS] ?: false,
                retentionDays = preferences[UserSettingsKeys.RETENTION_DAYS] ?: 90,
                highContrast = preferences[UserSettingsKeys.HIGH_CONTRAST] ?: false,
                reduceMotion = preferences[UserSettingsKeys.REDUCE_MOTION] ?: false,
                screenReaderOptimized = preferences[UserSettingsKeys.SCREEN_READER_OPTIMIZED] ?: false
            )
        }

    suspend fun updateThemeMode(mode: String) {
        context.dataStore.edit { it[UserSettingsKeys.THEME_MODE] = mode }
    }

    suspend fun updateLanguage(language: String) {
        context.dataStore.edit { it[UserSettingsKeys.LANGUAGE] = language }
    }

    suspend fun updateOfflineMode(enabled: Boolean) {
        context.dataStore.edit { it[UserSettingsKeys.OFFLINE_MODE_ENABLED] = enabled }
    }

    suspend fun updateSyncInterval(minutes: Int) {
        context.dataStore.edit { it[UserSettingsKeys.SYNC_INTERVAL_MINUTES] = minutes }
    }

    suspend fun <T> update(key: Preferences.Key<T>, value: T) {
        context.dataStore.edit { it[key] = value }
    }

    suspend fun clearAll() {
        context.dataStore.edit { it.clear() }
    }
}
```

---

## 7. EncryptedSharedPreferences

### Secure Token Storage

```kotlin
@Module
@InstallIn(SingletonComponent::class)
object SecureStorageModule {

    @Provides
    @Singleton
    fun provideEncryptedSharedPreferences(
        @ApplicationContext context: Context
    ): EncryptedSharedPreferences {
        val masterKey = MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .setRequestStrongBoxBacked(true)
            .build()

        return EncryptedSharedPreferences.create(
            context,
            "nexus_secure_prefs",
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        ) as EncryptedSharedPreferences
    }

    @Provides
    @Singleton
    fun provideTokenManager(
        encryptedPrefs: EncryptedSharedPreferences
    ): TokenManager = TokenManager(encryptedPrefs)
}

class TokenManager @Inject constructor(
    private val prefs: EncryptedSharedPreferences
) {
    companion object {
        private const val KEY_ACCESS_TOKEN = "access_token"
        private const val KEY_REFRESH_TOKEN = "refresh_token"
        private const val KEY_TOKEN_EXPIRY = "token_expiry"
        private const val KEY_USER_ID = "user_id"
        private const val KEY_DEVICE_ID = "device_id"
    }

    fun saveTokens(accessToken: String, refreshToken: String, expiresIn: Long) {
        val expiryTime = System.currentTimeMillis() + (expiresIn * 1000)
        prefs.edit().apply {
            putString(KEY_ACCESS_TOKEN, accessToken)
            putString(KEY_REFRESH_TOKEN, refreshToken)
            putLong(KEY_TOKEN_EXPIRY, expiryTime)
            apply()
        }
    }

    fun getAccessToken(): String? = prefs.getString(KEY_ACCESS_TOKEN, null)

    fun getRefreshToken(): String? = prefs.getString(KEY_REFRESH_TOKEN, null)

    fun isTokenExpired(): Boolean {
        val expiry = prefs.getLong(KEY_TOKEN_EXPIRY, 0)
        return System.currentTimeMillis() >= expiry
    }

    fun isTokenExpiringSoon(thresholdMs: Long = 300_000): Boolean {
        val expiry = prefs.getLong(KEY_TOKEN_EXPIRY, 0)
        return System.currentTimeMillis() >= (expiry - thresholdMs)
    }

    fun saveUserId(userId: String) {
        prefs.edit().putString(KEY_USER_ID, userId).apply()
    }

    fun getUserId(): String? = prefs.getString(KEY_USER_ID, null)

    fun saveDeviceId(deviceId: String) {
        prefs.edit().putString(KEY_DEVICE_ID, deviceId).apply()
    }

    fun getDeviceId(): String? = prefs.getString(KEY_DEVICE_ID, null)

    fun clearAll() {
        prefs.edit().clear().apply()
    }

    fun clearTokens() {
        prefs.edit().apply {
            remove(KEY_ACCESS_TOKEN)
            remove(KEY_REFRESH_TOKEN)
            remove(KEY_TOKEN_EXPIRY)
            apply()
        }
    }
}
```

### Android Keystore Integration

```kotlin
class KeyStoreManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    companion object {
        private const val KEYSTORE_PROVIDER = "AndroidKeyStore"
        private const val KEY_ALIAS = "nexus_ai_key"
        private const val KEYSTORE_NAME = "NexusAIKeyStore"
    }

    fun generateSecretKey(): SecretKey {
        val keyGenerator = KeyGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_AES,
            KEYSTORE_PROVIDER
        )
        val keyGenSpec = KeyGenParameterSpec.Builder(
            KEY_ALIAS,
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        )
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            .setKeySize(256)
            .setUserAuthenticationRequired(false)
            .setRandomizedEncryptionRequired(true)
            .build()

        keyGenerator.init(keyGenSpec)
        return keyGenerator.generateKey()
    }

    fun getSecretKey(): SecretKey? {
        val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER).apply { load(null) }
        val entry = keyStore.getEntry(KEY_ALIAS, null) as? KeyStore.SecretKeyEntry
        return entry?.secretKey
    }

    fun encrypt(plaintext: String): String {
        val key = getSecretKey() ?: generateSecretKey()
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, key)

        val iv = cipher.iv
        val encrypted = cipher.doFinal(plaintext.toByteArray(Charsets.UTF_8))

        // Prepend IV to ciphertext
        val combined = iv + encrypted
        return Base64.encodeToString(combined, Base64.NO_WRAP)
    }

    fun decrypt(encryptedBase64: String): String {
        val key = getSecretKey() ?: throw IllegalStateException("Key not found")
        val combined = Base64.decode(encryptedBase64, Base64.NO_WRAP)

        val iv = combined.sliceArray(0 until 12) // GCM IV is 12 bytes
        val ciphertext = combined.sliceArray(12 until combined.size)

        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val spec = GCMParameterSpec(128, iv)
        cipher.init(Cipher.DECRYPT_MODE, key, spec)

        val decrypted = cipher.doFinal(ciphertext)
        return String(decrypted, Charsets.UTF_8)
    }

    fun deleteKey() {
        val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER).apply { load(null) }
        keyStore.deleteEntry(KEY_ALIAS)
    }

    fun isKeyAvailable(): Boolean {
        return try {
            getSecretKey() != null
        } catch (e: Exception) {
            false
        }
    }
}
```

---

## 8. Offline Cache Strategy

### Cache Strategy Enum

```kotlin
enum class CacheStrategy {
    CACHE_FIRST,    // Check cache first, fetch from network if stale
    NETWORK_FIRST,  // Try network, fall back to cache
    CACHE_ONLY,     // Only use cache (true offline mode)
    NETWORK_ONLY    // Only use network (no caching)
}

enum class DataFreshness(
    val maxAgeMs: Long
) {
    REAL_TIME(0),                    // Always fetch from network
    SHORT(5 * 60 * 1000),           // 5 minutes
    MEDIUM(30 * 60 * 1000),         // 30 minutes
    LONG(2 * 60 * 60 * 1000),       // 2 hours
    DAY(24 * 60 * 60 * 1000),       // 24 hours
    WEEK(7 * 24 * 60 * 60 * 1000)   // 7 days
}
```

### Repository with Cache Strategy

```kotlin
abstract class CachedRepository<T : Any, E : Any> {

    abstract suspend fun fetchFromNetwork(): T?
    abstract suspend fun fetchFromLocal(): T?
    abstract suspend fun saveToLocal(data: T)
    abstract suspend fun getLastFetchTime(): Long
    abstract suspend fun setLastFetchTime(time: Long)

    open fun getDataFreshness(): DataFreshness = DataFreshness.MEDIUM

    suspend fun get(
        strategy: CacheStrategy = CacheStrategy.CACHE_FIRST
    ): Result<T> {
        return when (strategy) {
            CacheStrategy.CACHE_FIRST -> getCacheFirst()
            CacheStrategy.NETWORK_FIRST -> getNetworkFirst()
            CacheStrategy.CACHE_ONLY -> getCacheOnly()
            CacheStrategy.NETWORK_ONLY -> getNetworkOnly()
        }
    }

    private suspend fun getCacheFirst(): Result<T> {
        // 1. Try cache
        val cached = fetchFromLocal()
        val lastFetch = getLastFetchTime()
        val freshness = getDataFreshness()
        val isStale = System.currentTimeMillis() - lastFetch > freshness.maxAgeMs

        if (cached != null && !isStale) {
            return Result.success(cached)
        }

        // 2. Try network if stale or missing
        return try {
            val networkData = fetchFromNetwork()
            if (networkData != null) {
                saveToLocal(networkData)
                setLastFetchTime(System.currentTimeMillis())
                Result.success(networkData)
            } else if (cached != null) {
                Result.success(cached) // Stale cache is better than nothing
            } else {
                Result.failure(Exception("No data available"))
            }
        } catch (e: Exception) {
            if (cached != null) {
                Result.success(cached) // Return stale cache on network error
            } else {
                Result.failure(e)
            }
        }
    }

    private suspend fun getNetworkFirst(): Result<T> {
        return try {
            val networkData = fetchFromNetwork()
            if (networkData != null) {
                saveToLocal(networkData)
                setLastFetchTime(System.currentTimeMillis())
                Result.success(networkData)
            } else {
                val cached = fetchFromLocal()
                if (cached != null) {
                    Result.success(cached)
                } else {
                    Result.failure(Exception("No data available"))
                }
            }
        } catch (e: Exception) {
            val cached = fetchFromLocal()
            if (cached != null) {
                Result.success(cached)
            } else {
                Result.failure(e)
            }
        }
    }

    private suspend fun getCacheOnly(): Result<T> {
        val cached = fetchFromLocal()
        return if (cached != null) {
            Result.success(cached)
        } else {
            Result.failure(Exception("No cached data available"))
        }
    }

    private suspend fun getNetworkOnly(): Result<T> {
        return try {
            val networkData = fetchFromNetwork()
            if (networkData != null) {
                Result.success(networkData)
            } else {
                Result.failure(Exception("No data from network"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
```

### Cache Strategy Decision Flow

```
┌─────────────────────────────────────────────────────────┐
│                    Cache Strategy Flow                   │
│                                                         │
│  Request Data                                           │
│       │                                                 │
│       ▼                                                 │
│  ┌──────────────┐                                       │
│  │ User Setting  │                                     │
│  │ Cache Strategy│                                     │
│  └──────┬───────┘                                       │
│         │                                               │
│    ┌────┼────────────┬──────────────┐                   │
│    ▼    ▼            ▼              ▼                   │
│  CACHE  NETWORK    CACHE         NETWORK               │
│  FIRST  FIRST      ONLY          ONLY                  │
│    │    │            │              │                   │
│    ▼    ▼            ▼              ▼                   │
│  ┌────────┐   ┌──────────┐   ┌──────────┐              │
│  │ Cache  │   │ Network  │   │ Network  │              │
│  │ Valid? │   │ Available│   │ Only     │              │
│  └───┬────┘   └────┬─────┘   └────┬─────┘              │
│   Y  │  N     Y    │  N      Y    │  N                 │
│   │  │      │      │        │     │                    │
│   ▼  │      ▼      ▼        ▼     ▼                    │
│  Use │  ┌──────┐  Use     Use    Error                 │
│  Cache│  │Net?  │  Cache   Net   (No Data)              │
│      │  └──┬───┘                                             │
│      │  Y  │  N                                             │
│      ▼  │  ▼                                                 │
│    ┌────┐ Use                                                 │
│    │Net│ Cache                                               │
│    │?  │                                                     │
│    └────┘                                                     │
└─────────────────────────────────────────────────────────┘
```

---

## 9. Data Synchronization

### SyncManager

```kotlin
class SyncManager @Inject constructor(
    private val syncQueueDao: SyncQueueDao,
    private val conversationDao: ConversationDao,
    private val messageDao: MessageDao,
    private val documentDao: DocumentDao,
    private val agentDao: AgentDao,
    private val apiService: NexusApiService,
    private val networkMonitor: NetworkMonitor,
    private val keyStoreManager: KeyStoreManager
) {
    private val _syncState = MutableStateFlow<SyncState>(SyncState.Idle)
    val syncState: StateFlow<SyncState> = _syncState.asStateFlow()

    private val _syncProgress = MutableStateFlow(SyncProgress(0, 0))
    val syncProgress: StateFlow<SyncProgress> = _syncProgress.asStateFlow()

    suspend fun performSync(): SyncResult {
        if (!networkMonitor.isOnline()) {
            return SyncResult.Offline
        }

        _syncState.value = SyncState.Syncing

        return try {
            // 1. Push local changes
            pushLocalChanges()

            // 2. Pull remote changes
            pullRemoteChanges()

            // 3. Resolve any conflicts
            resolveConflicts()

            _syncState.value = SyncState.Success
            SyncResult.Success
        } catch (e: Exception) {
            _syncState.value = SyncState.Error(e.message ?: "Sync failed")
            SyncResult.Error(e)
        }
    }

    private suspend fun pushLocalChanges() {
        val pendingOps = syncQueueDao.getReadyOperations()

        _syncProgress.value = SyncProgress(total = pendingOps.size, completed = 0)

        for ((index, operation) in pendingOps.withIndex()) {
            try {
                syncQueueDao.markInProgress(operation.uid)

                when (operation.entityType) {
                    EntityType.CONVERSATION -> pushConversation(operation)
                    EntityType.MESSAGE -> pushMessage(operation)
                    EntityType.DOCUMENT -> pushDocument(operation)
                    EntityType.AGENT -> pushAgent(operation)
                    EntityType.PREFERENCE -> pushPreference(operation)
                }

                syncQueueDao.markCompleted(operation.uid)
                _syncProgress.value = SyncProgress(
                    total = pendingOps.size,
                    completed = index + 1
                )
            } catch (e: Exception) {
                val nextRetry = calculateNextRetry(operation.retryCount)
                syncQueueDao.markFailed(
                    uid = operation.uid,
                    error = e.message ?: "Unknown error",
                    nextRetryAt = nextRetry
                )
            }
        }
    }

    private suspend fun pushConversation(operation: SyncQueueEntity) {
        val data = Json.decodeFromString<ConversationDto>(operation.payload)
        when (operation.operation) {
            SyncOperation.CREATE -> apiService.createConversation(data)
            SyncOperation.UPDATE -> apiService.updateConversation(data.id, data)
            SyncOperation.DELETE -> apiService.deleteConversation(data.id)
        }
        conversationDao.markAsSynced(data.id)
    }

    private suspend fun pushMessage(operation: SyncQueueEntity) {
        val data = Json.decodeFromString<MessageDto>(operation.payload)
        when (operation.operation) {
            SyncOperation.CREATE -> apiService.createMessage(data.conversationId, data)
            SyncOperation.UPDATE -> apiService.updateMessage(data.id, data)
            SyncOperation.DELETE -> apiService.deleteMessage(data.id)
        }
        messageDao.markAsSynced(data.id)
    }

    private suspend fun pushDocument(operation: SyncQueueEntity) {
        val data = Json.decodeFromString<DocumentDto>(operation.payload)
        when (operation.operation) {
            SyncOperation.CREATE -> apiService.uploadDocument(data)
            SyncOperation.UPDATE -> apiService.updateDocument(data.id, data)
            SyncOperation.DELETE -> apiService.deleteDocument(data.id)
        }
        documentDao.markAsUploaded(data.id, data.remoteUrl ?: "")
    }

    private suspend fun pushAgent(operation: SyncQueueEntity) {
        val data = Json.decodeFromString<AgentDto>(operation.payload)
        when (operation.operation) {
            SyncOperation.CREATE -> apiService.createAgent(data)
            SyncOperation.UPDATE -> apiService.updateAgent(data.id, data)
            SyncOperation.DELETE -> apiService.deleteAgent(data.id)
        }
        agentDao.updateSyncStatus(data.id, SyncStatus.SYNCED)
    }

    private suspend fun pushPreference(operation: SyncQueueEntity) {
        val data = Json.decodeFromString<PreferenceDto>(operation.payload)
        apiService.updatePreference(data.key, data)
    }

    private suspend fun pullRemoteChanges() {
        try {
            val lastSyncTime = getLastSyncTimestamp()

            val remoteConversations = apiService.getConversationsSince(lastSyncTime)
            remoteConversations.forEach { conversationDao.upsert(it.toEntity()) }

            val remoteMessages = apiService.getMessagesSince(lastSyncTime)
            remoteMessages.forEach { messageDao.upsert(it.toEntity()) }

            val remoteDocuments = apiService.getDocumentsSince(lastSyncTime)
            remoteDocuments.forEach { documentDao.upsert(it.toEntity()) }

            val remoteAgents = apiService.getAgentsSince(lastSyncTime)
            remoteAgents.forEach { agentDao.upsert(it.toEntity()) }

            setLastSyncTimestamp(System.currentTimeMillis())
        } catch (e: Exception) {
            // Pull failure is non-fatal; local data is still valid
        }
    }

    private suspend fun resolveConflicts() {
        val conflictedConversations = conversationDao.getUnsyncedConversations()
        conflictedConversations.forEach { local ->
            val remote = apiService.getConversation(local.id) ?: return@forEach
            val resolved = ConflictResolver.resolve(local, remote)
            conversationDao.updateConversation(resolved)
        }
    }

    private fun calculateNextRetry(retryCount: Int): Long {
        val baseDelay = 30_000L // 30 seconds
        val maxDelay = 3_600_000L // 1 hour
        val delay = (baseDelay * 2.0.pow(retryCount.coerceAtMost(5))).toLong()
        return System.currentTimeMillis() + delay.coerceAtMost(maxDelay)
    }

    private suspend fun getLastSyncTimestamp(): Long {
        // Retrieve from DataStore
        return 0L // Placeholder
    }

    private suspend fun setLastSyncTimestamp(timestamp: Long) {
        // Store in DataStore
    }
}

sealed class SyncState {
    data object Idle : SyncState()
    data object Syncing : SyncState()
    data object Success : SyncState()
    data class Error(val message: String) : SyncState()
}

data class SyncProgress(
    val total: Int,
    val completed: Int
) {
    val percentage: Float
        get() = if (total > 0) completed.toFloat() / total else 0f
}

sealed class SyncResult {
    data object Success : SyncResult()
    data object Offline : SyncResult()
    data class Error(val exception: Exception) : SyncResult()
}
```

---

## 10. Sync Conflict Resolution

### Conflict Resolution Strategies

```kotlin
enum class ConflictStrategy {
    LAST_WRITE_WINS,
    SERVER_WINS,
    CLIENT_WINS,
    MERGE,
    PROMPT_USER
}

class ConflictResolver @Inject constructor(
    private val conflictStrategy: ConflictStrategy = ConflictStrategy.LAST_WRITE_WINS
) {

    fun <T : SyncableEntity> resolve(
        local: T,
        remote: T,
        strategy: ConflictStrategy = conflictStrategy
    ): T {
        return when (strategy) {
            ConflictStrategy.LAST_WRITE_WINS -> resolveByTimestamp(local, remote)
            ConflictStrategy.SERVER_WINS -> remote
            ConflictStrategy.CLIENT_WINS -> local
            ConflictStrategy.MERGE -> merge(local, remote)
            ConflictStrategy.PROMPT_USER -> throw ConflictRequiresUserInput(local, remote)
        }
    }

    private fun <T : SyncableEntity> resolveByTimestamp(local: T, remote: T): T {
        return if (local.updatedAt >= remote.updatedAt) local else remote
    }

    private fun <T : SyncableEntity> merge(local: T, remote: T): T {
        return when {
            local is ConversationEntity && remote is ConversationEntity -> {
                mergeConversations(local, remote) as T
            }
            local is MessageEntity && remote is MessageEntity -> {
                mergeMessages(local, remote) as T
            }
            else -> resolveByTimestamp(local, remote)
        }
    }

    private fun mergeConversations(
        local: ConversationEntity,
        remote: ConversationEntity
    ): ConversationEntity {
        return local.copy(
            title = mergeField(local.title, remote.title, local.updatedAt, remote.updatedAt),
            lastMessagePreview = mergeField(
                local.lastMessagePreview,
                remote.lastMessagePreview,
                local.updatedAt,
                remote.updatedAt
            ),
            isPinned = if (local.updatedAt >= remote.updatedAt) local.isPinned else remote.isPinned,
            messageCount = maxOf(local.messageCount, remote.messageCount),
            updatedAt = maxOf(local.updatedAt, remote.updatedAt),
            syncStatus = SyncStatus.SYNCED
        )
    }

    private fun mergeMessages(
        local: MessageEntity,
        remote: MessageEntity
    ): MessageEntity {
        return local.copy(
            content = if (local.updatedAt >= remote.updatedAt) local.content else remote.content,
            version = maxOf(local.version, remote.version) + 1,
            updatedAt = maxOf(local.updatedAt, remote.updatedAt),
            syncStatus = SyncStatus.SYNCED
        )
    }

    private fun mergeField(
        local: String?,
        remote: String?,
        localTimestamp: Long,
        remoteTimestamp: Long
    ): String? {
        return if (localTimestamp >= remoteTimestamp) local else remote
    }
}

interface SyncableEntity {
    val updatedAt: Long
    val syncStatus: SyncStatus
}

data class ConflictRequiresUserInput(
    val local: Any,
    val remote: Any
) : Exception("Conflict requires user resolution")
```

### Conflict Resolution Flow

```
┌─────────────────────────────────────────────────────┐
│              Conflict Resolution Flow                │
│                                                     │
│  Sync Operation                                     │
│       │                                             │
│       ▼                                             │
│  ┌──────────────┐                                   │
│  │ Detect       │                                   │
│  │ Conflict     │                                   │
│  │ (version     │                                   │
│  │  mismatch)   │                                   │
│  └──────┬───────┘                                   │
│         │                                           │
│         ▼                                           │
│  ┌──────────────────┐                               │
│  │ Conflict Strategy │                              │
│  └──────┬───────────┘                               │
│         │                                           │
│    ┌────┼────┬────────┬──────────┐                  │
│    ▼    ▼    ▼        ▼          ▼                  │
│  LWW  Server Client  Merge     Prompt               │
│       Wins  Wins              User                  │
│    │    │    │        │          │                   │
│    ▼    ▼    ▼        ▼          ▼                   │
│  ┌────┐┌────┐┌────┐┌──────┐┌────────┐              │
│  │Time││Use ││Use ││Merge ││Show UI │              │
│  │Comp││Serv││Clie││Fields││Dialog  │              │
│  │are ││er  ││nt  ││      ││        │              │
│  └──┬─┘└──┬─┘└──┬─┘└──┬───┘└───┬────┘              │
│     │     │     │     │        │                    │
│     ▼     ▼     ▼     ▼        ▼                    │
│  ┌─────────────────────────────────┐                │
│  │      Update Local Database      │                │
│  │      Mark as SYNCED             │                │
│  └─────────────────────────────────┘                │
└─────────────────────────────────────────────────────┘
```

---

## 11. Sync Queue

### Sync Queue Manager

```kotlin
class SyncQueueManager @Inject constructor(
    private val syncQueueDao: SyncQueueDao,
    private val networkMonitor: NetworkMonitor
) {
    suspend fun enqueue(
        entityType: EntityType,
        entityId: String,
        operation: SyncOperation,
        payload: String
    ) {
        syncQueueDao.enqueue(entityType, entityId, operation, payload)
    }

    suspend fun getPendingCount(): Int {
        return syncQueueDao.getPendingOperations().first().size
    }

    suspend fun getPendingCountFlow(): Flow<Int> {
        return syncQueueDao.getPendingCount()
    }

    suspend fun retryFailed() {
        val failedOps = syncQueueDao.getRetryableOperations()
        failedOps.forEach { op ->
            syncQueueDao.updateStatus(op.uid, SyncQueueStatus.PENDING)
        }
    }

    suspend fun cancelAll() {
        val pending = syncQueueDao.getPendingOperations().first()
        pending.forEach { op ->
            syncQueueDao.updateStatus(op.uid, SyncQueueStatus.CANCELLED)
        }
    }

    suspend fun purgeCompleted() {
        syncQueueDao.purgeCompleted()
    }

    suspend fun purgeAll() {
        syncQueueDao.purgeAll()
    }

    suspend fun getRetryDelays(): List<Pair<Long, Long>> {
        val pending = syncQueueDao.getPendingOperations().first()
        return pending.map { op ->
            op.uid to op.nextRetryAt
        }
    }
}
```

### Retry Logic Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    Retry Logic                           │
│                                                         │
│  Operation Fails                                        │
│       │                                                 │
│       ▼                                                 │
│  ┌──────────────┐                                       │
│  │ retry_count  │                                       │
│  │ < max_retries│                                       │
│  └──────┬───────┘                                       │
│     Yes │  No                                           │
│     │   │                                               │
│     ▼   ▼                                               │
│  ┌──────┐ ┌──────────┐                                  │
│  │Calc  │ │ Purge    │                                  │
│  │Retry │ │ Operation│                                  │
│  │Delay │ └──────────┘                                  │
│  └──┬───┘                                               │
│     │                                                   │
│     ▼                                                   │
│  ┌──────────────────────────────┐                       │
│  │ next_retry_at = now + delay  │                       │
│  │                              │                       │
│  │ delay = min(                 │                       │
│  │   base_delay * 2^retry_count │                       │
│  │   max_delay                  │                       │
│  │ )                            │                       │
│  └──────────────┬───────────────┘                       │
│                 │                                       │
│                 ▼                                       │
│  ┌──────────────────────────────┐                       │
│  │ Wait for next_retry_at       │                       │
│  │ (WorkManager handles this)   │                       │
│  └──────────────────────────────┘                       │
│                                                         │
│  Retry Schedule:                                        │
│  ┌───────┬──────────┬──────────┐                        │
│  │ Retry │ Delay    │ Total    │                        │
│  ├───────┼──────────┼──────────┤                        │
│  │ 1     │ 30s      │ 30s      │                        │
│  │ 2     │ 1m       │ 1m30s    │                        │
│  │ 3     │ 2m       │ 3m30s    │                        │
│  │ 4     │ 4m       │ 7m30s    │                        │
│  │ 5     │ 8m       │ 15m30s   │                        │
│  └───────┴──────────┴──────────┘                        │
└─────────────────────────────────────────────────────────┘
```

---

## 12. Sync Status Indicators

### Connectivity Monitor

```kotlin
class NetworkMonitor @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val connectivityManager = context.getSystemService<ConnectivityManager>()

    private val _isOnline = MutableStateFlow(false)
    val isOnline: StateFlow<Boolean> = _isOnline.asStateFlow()

    private val _connectionType = MutableStateFlow(ConnectionType.NONE)
    val connectionType: StateFlow<ConnectionType> = _connectionType.asStateFlow()

    private val networkCallback = object : ConnectivityManager.NetworkCallback() {
        override fun onAvailable(network: Network) {
            _isOnline.value = true
        }

        override fun onLost(network: Network) {
            _isOnline.value = false
        }

        override fun onCapabilitiesChanged(
            network: Network,
            capabilities: NetworkCapabilities
        ) {
            _connectionType.value = when {
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) -> ConnectionType.WIFI
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) -> ConnectionType.CELLULAR
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET) -> ConnectionType.ETHERNET
                else -> ConnectionType.OTHER
            }
        }
    }

    fun startMonitoring() {
        val request = NetworkRequest.Builder()
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .build()
        connectivityManager?.registerNetworkCallback(request, networkCallback)

        // Check current state
        val activeNetwork = connectivityManager?.activeNetwork
        val capabilities = connectivityManager?.getNetworkCapabilities(activeNetwork)
        _isOnline.value = capabilities?.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET) == true
    }

    fun stopMonitoring() {
        connectivityManager?.unregisterNetworkCallback(networkCallback)
    }

    fun isOnline(): Boolean = _isOnline.value
}

enum class ConnectionType {
    WIFI, CELLULAR, ETHERNET, OTHER, NONE
}
```

### Online/Offline Banner Composable

```kotlin
@Composable
fun SyncStatusBar(
    syncState: SyncState,
    syncProgress: SyncProgress,
    isOnline: Boolean,
    modifier: Modifier = Modifier
) {
    val animatedColor by animateColorAsState(
        targetValue = when {
            !isOnline -> MaterialTheme.colorScheme.error
            syncState is SyncState.Syncing -> MaterialTheme.colorScheme.tertiary
            syncState is SyncState.Error -> MaterialTheme.colorScheme.error
            else -> MaterialTheme.colorScheme.primary
        },
        label = "statusBarColor"
    )

    AnimatedVisibility(
        visible = !isOnline || syncState is SyncState.Syncing || syncState is SyncState.Error,
        enter = slideInVertically(initialOffsetY = { -it }),
        exit = slideOutVertically(targetOffsetY = { -it })
    ) {
        Surface(
            color = animatedColor,
            modifier = modifier.fillMaxWidth()
        ) {
            Row(
                modifier = Modifier
                    .padding(horizontal = 16.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    when {
                        !isOnline -> {
                            Icon(
                                Icons.Filled.CloudOff,
                                contentDescription = "Offline",
                                tint = MaterialTheme.colorScheme.onError
                            )
                            Text(
                                text = "You're offline",
                                color = MaterialTheme.colorScheme.onError,
                                style = MaterialTheme.typography.labelMedium
                            )
                        }
                        syncState is SyncState.Syncing -> {
                            CircularProgressIndicator(
                                modifier = Modifier.size(16.dp),
                                strokeWidth = 2.dp,
                                color = MaterialTheme.colorScheme.onTertiary
                            )
                            Text(
                                text = "Syncing... ${syncProgress.completed}/${syncProgress.total}",
                                color = MaterialTheme.colorScheme.onTertiary,
                                style = MaterialTheme.typography.labelMedium
                            )
                        }
                        syncState is SyncState.Error -> {
                            Icon(
                                Icons.Filled.SyncProblem,
                                contentDescription = "Sync error",
                                tint = MaterialTheme.colorScheme.onError
                            )
                            Text(
                                text = "Sync failed",
                                color = MaterialTheme.colorScheme.onError,
                                style = MaterialTheme.typography.labelMedium
                            )
                        }
                    }
                }

                if (syncState is SyncState.Syncing) {
                    LinearProgressIndicator(
                        progress = { syncProgress.percentage },
                        modifier = Modifier
                            .width(80.dp)
                            .height(4.dp),
                        color = MaterialTheme.colorScheme.onTertiary
                    )
                }
            }
        }
    }
}
```

---

## 13. WorkManager Configuration

### Sync Worker

```kotlin
@HiltWorker
class SyncWorker @AssistedInject constructor(
    @Assisted context: Context,
    @Assisted workerParams: WorkerParameters,
    private val syncManager: SyncManager,
    private val networkMonitor: NetworkMonitor
) : CoroutineWorker(context, workerParams) {

    override suspend fun doWork(): Result {
        if (!networkMonitor.isOnline()) {
            return Result.retry()
        }

        return when (val syncResult = syncManager.performSync()) {
            is SyncResult.Success -> Result.success()
            is SyncResult.Offline -> Result.retry()
            is SyncResult.Error -> {
                if (runAttemptCount < 3) {
                    Result.retry()
                } else {
                    Result.failure()
                }
            }
        }
    }

    override suspend fun getForegroundInfo(): ForegroundInfo {
        return ForegroundInfo(
            SYNC_NOTIFICATION_ID,
            createSyncNotification()
        )
    }

    private fun createSyncNotification(): Notification {
        val channel = NotificationChannel(
            SYNC_CHANNEL_ID,
            "Sync",
            NotificationManager.IMPORTANCE_LOW
        )
        val manager = applicationContext.getSystemService<NotificationManager>()
        manager?.createNotificationChannel(channel)

        return NotificationCompat.Builder(applicationContext, SYNC_CHANNEL_ID)
            .setContentTitle("Syncing data")
            .setContentText("Synchronizing your data...")
            .setSmallIcon(R.drawable.ic_sync)
            .setOngoing(true)
            .build()
    }

    companion object {
        const val SYNC_NOTIFICATION_ID = 1001
        const val SYNC_CHANNEL_ID = "sync_channel"
        const val WORK_NAME_PERIODIC = "periodic_sync"
        const val WORK_NAME_ONE_TIME = "one_time_sync"
    }
}
```

### WorkManager Setup

```kotlin
@Module
@InstallIn(SingletonComponent::class)
object WorkManagerModule {

    @Provides
    @Singleton
    fun provideWorkManager(
        @ApplicationContext context: Context
    ): WorkManager {
        return WorkManager.getInstance(context)
    }
}

class SyncScheduler @Inject constructor(
    private val workManager: WorkManager
) {
    fun schedulePeriodicSync(intervalMinutes: Long = 15) {
        val constraints = Constraints.Builder()
            .setRequiredNetworkType(NetworkType.CONNECTED)
            .setRequiresBatteryNotLow(true)
            .setRequiresCharging(false)
            .build()

        val syncRequest = PeriodicWorkRequestBuilder<SyncWorker>(
            repeatInterval = intervalMinutes,
            repeatIntervalTimeUnit = TimeUnit.MINUTES
        )
            .setConstraints(constraints)
            .setBackoffCriteria(
                BackoffPolicy.EXPONENTIAL,
                WorkRequest.MIN_BACKOFF_MILLIS,
                TimeUnit.MILLISECONDS
            )
            .addTag("sync")
            .build()

        workManager.enqueueUniquePeriodicWork(
            SyncWorker.WORK_NAME_PERIODIC,
            ExistingPeriodicWorkPolicy.UPDATE,
            syncRequest
        )
    }

    fun scheduleOneTimeSync() {
        val constraints = Constraints.Builder()
            .setRequiredNetworkType(NetworkType.CONNECTED)
            .build()

        val syncRequest = OneTimeWorkRequestBuilder<SyncWorker>()
            .setConstraints(constraints)
            .setExpedited(OutOfQuotaPolicy.RUN_AS_NON_EXPEDITED_WORK_REQUEST)
            .addTag("sync")
            .build()

        workManager.enqueueUniqueWork(
            SyncWorker.WORK_NAME_ONE_TIME,
            ExistingWorkPolicy.REPLACE,
            syncRequest
        )
    }

    fun scheduleChainedSync() {
        val syncWork = OneTimeWorkRequestBuilder<SyncWorker>()
            .setConstraints(
                Constraints.Builder()
                    .setRequiredNetworkType(NetworkType.CONNECTED)
                    .build()
            )
            .addTag("sync")
            .build()

        val cleanupWork = OneTimeWorkRequestBuilder<CleanupWorker>()
            .addTag("cleanup")
            .build()

        val uploadWork = OneTimeWorkRequestBuilder<UploadWorker>()
            .setConstraints(
                Constraints.Builder()
                    .setRequiredNetworkType(NetworkType.CONNECTED)
                    .build()
            )
            .addTag("upload")
            .build()

        workManager.beginWith(syncWork)
            .then(uploadWork)
            .then(cleanupWork)
            .enqueue()
    }

    fun cancelAllSync() {
        workManager.cancelAllWorkByTag("sync")
    }

    fun getSyncStatus(): LiveData<List<WorkInfo>> {
        return workManager.getWorkInfosByTagLiveData("sync")
    }
}
```

---

## 14. WorkManager Sync Tasks

### Cleanup Worker

```kotlin
@HiltWorker
class CleanupWorker @AssistedInject constructor(
    @Assisted context: Context,
    @Assisted workerParams: WorkerParameters,
    private val syncQueueDao: SyncQueueDao,
    private val conversationDao: ConversationDao,
    private val documentDao: DocumentDao,
    private val userSettingsRepository: UserSettingsRepository
) : CoroutineWorker(context, workerParams) {

    override suspend fun doWork(): Result {
        return try {
            // 1. Purge completed sync queue entries
            syncQueueDao.purgeCompleted()

            // 2. Purge exhausted retries
            syncQueueDao.purgeExhaustedRetries()

            // 3. Purge soft-deleted records older than 30 days
            val thirtyDaysAgo = System.currentTimeMillis() - (30L * 24 * 60 * 60 * 1000)
            conversationDao.purgeDeletedConversations()

            // 4. Check storage limits
            val settings = userSettingsRepository.settings.first()
            val maxSizeBytes = settings.maxCacheSizeMb.toLong() * 1024 * 1024
            checkStorageLimits(maxSizeBytes)

            Result.success()
        } catch (e: Exception) {
            Result.failure()
        }
    }

    private suspend fun checkStorageLimits(maxSizeBytes: Long) {
        // Implementation to check and enforce storage limits
    }
}
```

### Upload Worker

```kotlin
@HiltWorker
class UploadWorker @AssistedInject constructor(
    @Assisted context: Context,
    @Assisted workerParams: WorkerParameters,
    private val documentDao: DocumentDao,
    private val apiService: NexusApiService,
    private val keyStoreManager: KeyStoreManager
) : CoroutineWorker(context, workerParams) {

    override suspend fun doWork(): Result {
        val pendingDocs = documentDao.getDocumentsByProcessingStatus(ProcessingStatus.PENDING)

        for (doc in pendingDocs) {
            try {
                documentDao.updateProcessingStatus(doc.id, ProcessingStatus.PROCESSING)

                val localUri = doc.localUri ?: continue
                val file = File(Uri.parse(localUri).path ?: continue)

                if (!file.exists()) {
                    documentDao.updateProcessingStatus(doc.id, ProcessingStatus.FAILED)
                    continue
                }

                // Calculate checksum for integrity verification
                val checksum = file.calculateSHA256()
                if (doc.checksum != null && doc.checksum != checksum) {
                    documentDao.updateProcessingStatus(doc.id, ProcessingStatus.FAILED)
                    continue
                }

                val requestBody = file.asRequestBody("application/octet-stream".toMediaTypeOrNull())
                val multipart = MultipartBody.Builder()
                    .setType(MultipartBody.FORM)
                    .addFormDataPart("file", doc.name, requestBody)
                    .addFormDataPart("checksum", checksum)
                    .build()

                val response = apiService.uploadFile(multipart)

                documentDao.updateProcessingStatus(doc.id, ProcessingStatus.COMPLETED)
                documentDao.markAsUploaded(doc.id, response.url)
            } catch (e: Exception) {
                documentDao.updateProcessingStatus(doc.id, ProcessingStatus.FAILED)
                if (runAttemptCount >= 3) {
                    continue
                }
                return Result.retry()
            }
        }

        return Result.success()
    }
}
```

---

## 15. Background Sync

### Background Sync Configuration

```kotlin
class BackgroundSyncManager @Inject constructor(
    private val workManager: WorkManager,
    private val userSettingsRepository: UserSettingsRepository,
    private val networkMonitor: NetworkMonitor
) {
    private var isConfigured = false

    fun configure() {
        if (isConfigured) return

        CoroutineScope(Dispatchers.IO).launch {
            userSettingsRepository.settings.collect { settings ->
                if (settings.autoSync && settings.offlineModeEnabled) {
                    schedulePeriodicSync(settings.syncIntervalMinutes.toLong())
                } else {
                    cancelPeriodicSync()
                }
            }
        }

        isConfigured = true
    }

    private fun schedulePeriodicSync(intervalMinutes: Long) {
        val constraints = Constraints.Builder()
            .setRequiredNetworkType(NetworkType.CONNECTED)
            .setRequiresBatteryNotLow(true)
            .build()

        val request = PeriodicWorkRequestBuilder<SyncWorker>(
            intervalMinutes, TimeUnit.MINUTES
        )
            .setConstraints(constraints)
            .setBackoffCriteria(
                BackoffPolicy.EXPONENTIAL,
                15, TimeUnit.MINUTES
            )
            .addTag("periodic_sync")
            .build()

        workManager.enqueueUniquePeriodicWork(
            "background_sync",
            ExistingPeriodicWorkPolicy.UPDATE,
            request
        )
    }

    private fun cancelPeriodicSync() {
        workManager.cancelUniqueWork("background_sync")
    }

    fun forceSyncNow() {
        val constraints = Constraints.Builder()
            .setRequiredNetworkType(NetworkType.CONNECTED)
            .build()

        val request = OneTimeWorkRequestBuilder<SyncWorker>()
            .setConstraints(constraints)
            .addTag("force_sync")
            .build()

        workManager.enqueue(request)
    }
}
```

### Sync Schedule Diagram

```
┌─────────────────────────────────────────────────────────┐
│              Background Sync Timeline                     │
│                                                         │
│  0min    15min    30min    45min    60min    75min     │
│  │        │        │        │        │        │        │
│  ▼        ▼        ▼        ▼        ▼        ▼        │
│  ┌───┐   ┌───┐   ┌───┐   ┌───┐   ┌───┐   ┌───┐      │
│  │ S │   │ S │   │ S │   │ S │   │ S │   │ S │      │
│  └─┬─┘   └─┬─┘   └─┬─┘   └─┬─┘   └─┬─┘   └─┬─┘      │
│    │       │       │       │       │       │          │
│    ▼       ▼       ▼       ▼       ▼       ▼          │
│  ┌──────────────────────────────────────────────┐     │
│  │ Each sync:                                    │     │
│  │ 1. Push pending local changes                 │     │
│  │ 2. Pull remote changes                        │     │
│  │ 3. Resolve conflicts                          │     │
│  │ 4. Update sync timestamps                     │     │
│  └──────────────────────────────────────────────┘     │
│                                                         │
│  Conditions:                                            │
│  ✓ Network available                                    │
│  ✓ Battery not low                                      │
│  ✓ Auto-sync enabled                                    │
│  ✓ Offline mode enabled                                 │
└─────────────────────────────────────────────────────────┘
```

---

## 16. Offline Chat

### Offline Chat Manager

```kotlin
class OfflineChatManager @Inject constructor(
    private val messageDao: MessageDao,
    private val conversationDao: ConversationDao,
    private val syncQueueManager: SyncQueueManager,
    private val networkMonitor: NetworkMonitor,
    private val apiService: NexusApiService
) {
    data class ChatResult(
        val message: MessageEntity,
        val isQueued: Boolean = false,
        val error: String? = null
    )

    suspend fun sendMessage(
        conversationId: String,
        content: String,
        role: MessageRole = MessageRole.USER,
        attachments: List<String>? = null
    ): ChatResult {
        val messageId = UUID.randomUUID().toString()
        val timestamp = System.currentTimeMillis()

        val message = MessageEntity(
            id = messageId,
            conversationId = conversationId,
            role = role,
            content = content,
            timestamp = timestamp,
            syncStatus = SyncStatus.PENDING,
            attachments = attachments?.let { Json.encodeToString(it) }
        )

        // Always save locally first
        messageDao.insertMessage(message)

        // Update conversation preview
        val conv = conversationDao.getConversationByIdOnce(conversationId)
        if (conv != null) {
            conversationDao.updateMessagePreview(
                id = conversationId,
                count = conv.messageCount + 1,
                preview = content.take(100),
                timestamp = timestamp
            )
        }

        if (networkMonitor.isOnline()) {
            // Try to send immediately
            return try {
                val response = apiService.sendMessage(conversationId, message.toDto())
                messageDao.markAsSynced(messageId)
                ChatResult(message.copy(syncStatus = SyncStatus.SYNCED))
            } catch (e: Exception) {
                // Queue for later sync
                syncQueueManager.enqueue(
                    entityType = EntityType.MESSAGE,
                    entityId = messageId,
                    operation = SyncOperation.CREATE,
                    payload = Json.encodeToString(message.toDto())
                )
                ChatResult(message, isQueued = true)
            }
        } else {
            // Queue for sync when online
            syncQueueManager.enqueue(
                entityType = EntityType.MESSAGE,
                entityId = messageId,
                operation = SyncOperation.CREATE,
                payload = Json.encodeToString(message.toDto())
            )
            return ChatResult(message, isQueued = true)
        }
    }

    suspend fun getQueuedMessages(conversationId: String): List<MessageEntity> {
        return messageDao.getMessagesByConversation(conversationId).first()
            .filter { it.syncStatus == SyncStatus.PENDING }
    }

    fun observeMessages(conversationId: String): Flow<List<MessageEntity>> {
        return messageDao.getMessagesByConversation(conversationId)
    }
}
```

### Offline Chat UI

```kotlin
@Composable
fun ChatScreen(
    conversationId: String,
    viewModel: ChatViewModel = hiltViewModel()
) {
    val messages by viewModel.messages.collectAsStateWithLifecycle()
    val isOnline by viewModel.isOnline.collectAsStateWithLifecycle()
    val inputText by viewModel.inputText.collectAsStateWithLifecycle()
    val isSending by viewModel.isSending.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize()) {
        // Connection status banner
        if (!isOnline) {
            OfflineBanner(
                message = "Messages will be sent when you're back online",
                modifier = Modifier.fillMaxWidth()
            )
        }

        // Message list
        LazyColumn(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth(),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(messages, key = { it.id }) { message ->
                MessageBubble(
                    message = message,
                    isQueued = message.syncStatus == SyncStatus.PENDING,
                    isFailed = message.syncStatus == SyncStatus.FAILED
                )
            }
        }

        // Input bar
        ChatInputBar(
            text = inputText,
            onTextChange = viewModel::updateInputText,
            onSend = {
                viewModel.sendMessage(inputText)
            },
            isOnline = isOnline,
            isSending = isSending,
            modifier = Modifier.fillMaxWidth()
        )
    }
}

@Composable
fun MessageBubble(
    message: MessageEntity,
    isQueued: Boolean,
    isFailed: Boolean,
    modifier: Modifier = Modifier
) {
    val isUser = message.role == MessageRole.USER

    Row(
        modifier = modifier.fillMaxWidth(),
        horizontalArrangement = if (isUser) Arrangement.End else Arrangement.Start
    ) {
        Column(
            modifier = Modifier
                .widthIn(max = 280.dp)
                .background(
                    color = if (isUser) {
                        MaterialTheme.colorScheme.primaryContainer
                    } else {
                        MaterialTheme.colorScheme.surfaceVariant
                    },
                    shape = RoundedCornerShape(16.dp)
                )
                .padding(12.dp)
        ) {
            Text(
                text = message.content,
                style = MaterialTheme.typography.bodyMedium,
                color = if (isUser) {
                    MaterialTheme.colorScheme.onPrimaryContainer
                } else {
                    MaterialTheme.colorScheme.onSurfaceVariant
                }
            )

            if (isQueued || isFailed) {
                Spacer(modifier = Modifier.height(4.dp))
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    Icon(
                        imageVector = when {
                            isFailed -> Icons.Filled.ErrorOutline
                            isQueued -> Icons.Filled.Schedule
                            else -> Icons.Filled.Check
                        },
                        contentDescription = when {
                            isFailed -> "Failed to send"
                            isQueued -> "Queued to send"
                            else -> "Sent"
                        },
                        modifier = Modifier.size(12.dp),
                        tint = when {
                            isFailed -> MaterialTheme.colorScheme.error
                            isQueued -> MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.6f)
                            else -> MaterialTheme.colorScheme.primary
                        }
                    )
                    Text(
                        text = when {
                            isFailed -> "Failed"
                            isQueued -> "Queued"
                            else -> "Sent"
                        },
                        style = MaterialTheme.typography.labelSmall,
                        color = when {
                            isFailed -> MaterialTheme.colorScheme.error
                            isQueued -> MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.6f)
                            else -> MaterialTheme.colorScheme.primary
                        }
                    )
                }
            }
        }
    }
}

@Composable
fun OfflineBanner(
    message: String,
    modifier: Modifier = Modifier
) {
    Surface(
        color = MaterialTheme.colorScheme.errorContainer,
        modifier = modifier
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Icon(
                Icons.Filled.CloudOff,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.onErrorContainer,
                modifier = Modifier.size(16.dp)
            )
            Text(
                text = message,
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onErrorContainer
            )
        }
    }
}
```

---

## 17. Offline Document Upload

### Offline Upload Queue

```kotlin
class OfflineDocumentManager @Inject constructor(
    private val documentDao: DocumentDao,
    private val syncQueueManager: SyncQueueManager,
    private val networkMonitor: NetworkMonitor,
    private val keyStoreManager: KeyStoreManager,
    @ApplicationContext private val context: Context
) {
    data class UploadResult(
        val document: DocumentEntity,
        val isQueued: Boolean,
        val error: String? = null
    )

    suspend fun queueDocument(
        name: String,
        mimeType: String,
        uri: Uri,
        userId: String
    ): UploadResult {
        val documentId = UUID.randomUUID().toString()

        // Copy file to app's internal storage for persistence
        val localPath = copyToInternalStorage(uri, documentId, name)
        val file = File(localPath)
        val checksum = file.calculateSHA256()

        val document = DocumentEntity(
            id = documentId,
            userId = userId,
            name = name,
            mimeType = mimeType,
            sizeBytes = file.length(),
            localUri = localPath,
            checksum = checksum,
            processingStatus = ProcessingStatus.PENDING,
            syncStatus = SyncStatus.PENDING
        )

        documentDao.insertDocument(document)

        if (networkMonitor.isOnline()) {
            return try {
                uploadDocument(document)
                documentDao.markAsUploaded(documentId, document.remoteUrl ?: "")
                UploadResult(document.copy(syncStatus = SyncStatus.SYNCED), isQueued = false)
            } catch (e: Exception) {
                queueForSync(document)
                UploadResult(document, isQueued = true)
            }
        } else {
            queueForSync(document)
            return UploadResult(document, isQueued = true)
        }
    }

    private suspend fun uploadDocument(document: DocumentEntity) {
        val localUri = document.localUri ?: throw IllegalStateException("No local URI")
        val file = File(Uri.parse(localUri).path ?: throw IllegalStateException("Invalid URI"))

        val requestBody = file.asRequestBody(document.mimeType.toMediaTypeOrNull())
        val multipart = MultipartBody.Builder()
            .setType(MultipartBody.FORM)
            .addFormDataPart("file", document.name, requestBody)
            .addFormDataPart("checksum", document.checksum ?: "")
            .build()

        val response = apiService.uploadFile(multipart)
        documentDao.markAsUploaded(document.id, response.url)
    }

    private suspend fun queueForSync(document: DocumentEntity) {
        syncQueueManager.enqueue(
            entityType = EntityType.DOCUMENT,
            entityId = document.id,
            operation = SyncOperation.CREATE,
            payload = Json.encodeToString(document.toDto())
        )
    }

    private suspend fun copyToInternalStorage(uri: Uri, documentId: String, fileName: String): String {
        val dir = File(context.filesDir, "documents/$documentId")
        dir.mkdirs()
        val destFile = File(dir, fileName)

        context.contentResolver.openInputStream(uri)?.use { input ->
            destFile.outputStream().use { output ->
                input.copyTo(output)
            }
        } ?: throw IllegalStateException("Cannot open input stream")

        return destFile.absolutePath
    }

    suspend fun getPendingUploads(): List<DocumentEntity> {
        return documentDao.getDocumentsByProcessingStatus(ProcessingStatus.PENDING)
    }

    suspend fun retryUpload(documentId: String): UploadResult {
        val document = documentDao.getDocumentById(documentId)
            ?: return UploadResult(
                DocumentEntity(
                    id = documentId, userId = "", name = "", mimeType = "",
                    sizeBytes = 0
                ),
                isQueued = false,
                error = "Document not found"
            )

        return if (networkMonitor.isOnline()) {
            try {
                uploadDocument(document)
                UploadResult(document.copy(syncStatus = SyncStatus.SYNCED), isQueued = false)
            } catch (e: Exception) {
                UploadResult(document, isQueued = true, error = e.message)
            }
        } else {
            UploadResult(document, isQueued = true)
        }
    }
}
```

---

## 18. Offline Search

### Local Search Implementation

```kotlin
class OfflineSearchManager @Inject constructor(
    private val conversationDao: ConversationDao,
    private val messageDao: MessageDao,
    private val documentDao: DocumentDao,
    private val agentDao: AgentDao
) {
    data class SearchResult(
        val conversations: List<ConversationEntity>,
        val messages: List<MessageEntity>,
        val documents: List<DocumentEntity>,
        val agents: List<AgentEntity>
    )

    fun search(query: String, userId: String): Flow<SearchResult> {
        if (query.isBlank()) {
            return flowOf(SearchResult(emptyList(), emptyList(), emptyList(), emptyList()))
        }

        val sanitizedQuery = sanitizeQuery(query)

        return combine(
            conversationDao.searchConversations(sanitizedQuery),
            messageDao.searchMessages(sanitizedQuery),
            documentDao.searchDocuments(userId, sanitizedQuery),
            agentDao.searchAgents(userId, sanitizedQuery)
        ) { conversations, messages, documents, agents ->
            SearchResult(conversations, messages, documents, agents)
        }
    }

    fun searchConversations(query: String): Flow<List<ConversationEntity>> {
        return conversationDao.searchConversations(sanitizeQuery(query))
    }

    fun searchMessages(query: String): Flow<List<MessageEntity>> {
        return messageDao.searchMessages(sanitizeQuery(query))
    }

    fun searchDocuments(userId: String, query: String): Flow<List<DocumentEntity>> {
        return documentDao.searchDocuments(userId, sanitizeQuery(query))
    }

    private fun sanitizeQuery(query: String): String {
        // Remove special SQL characters to prevent injection
        return query.replace(Regex("['%_\\\\]"), "")
            .trim()
            .take(100)
    }
}
```

### Search Fallback Strategy

```
┌─────────────────────────────────────────────────┐
│              Search Fallback Flow                │
│                                                 │
│  User Enters Search Query                       │
│       │                                         │
│       ▼                                         │
│  ┌──────────────┐                               │
│  │ Is Online?   │                               │
│  └──────┬───────┘                               │
│     Yes │  No                                   │
│     │   │                                       │
│     ▼   ▼                                       │
│  ┌────────────────┐    ┌──────────────────┐     │
│  │ Try Remote     │    │ Use Local Only   │     │
│  │ Search First   │    │ (Room Full-Text) │     │
│  └───────┬────────┘    └────────┬─────────┘     │
│       Success  Failure          │               │
│       │        │               │               │
│       ▼        ▼               ▼               │
│  ┌─────────┐ ┌───────────┐ ┌──────────────┐   │
│  │ Show    │ │Fall back  │ │ Show Local   │   │
│  │ Remote  │ │to local   │ │ Results Only │   │
│  │ Results │ │search     │ │ + Stale      │   │
│  └─────────┘ └─────┬─────┘ │  Indicator   │   │
│                    │       └──────────────┘   │
│                    ▼                           │
│              ┌──────────────┐                  │
│              │ Show Local   │                  │
│              │ Results +    │                  │
│              │ Stale Badge  │                  │
│              └──────────────┘                  │
└─────────────────────────────────────────────────┘
```

---

## 19. Offline Data Freshness

### Stale Data Indicators

```kotlin
data class DataFreshnessInfo(
    val lastUpdated: Long,
    val freshness: DataFreshness,
    val isStale: Boolean,
    val staleDuration: String
) {
    companion object {
        fun from(lastUpdated: Long, freshness: DataFreshness): DataFreshnessInfo {
            val now = System.currentTimeMillis()
            val age = now - lastUpdated
            val isStale = age > freshness.maxAgeMs

            return DataFreshnessInfo(
                lastUpdated = lastUpdated,
                freshness = freshness,
                isStale = isStale,
                staleDuration = formatDuration(age)
            )
        }

        private fun formatDuration(durationMs: Long): String {
            val minutes = durationMs / (60 * 1000)
            val hours = durationMs / (60 * 60 * 1000)
            val days = durationMs / (24 * 60 * 60 * 1000)

            return when {
                minutes < 1 -> "just now"
                minutes < 60 -> "${minutes}m ago"
                hours < 24 -> "${hours}h ago"
                else -> "${days}d ago"
            }
        }
    }
}

@Composable
fun StaleDataIndicator(
    freshnessInfo: DataFreshnessInfo,
    onRefresh: () -> Unit,
    modifier: Modifier = Modifier
) {
    AnimatedVisibility(
        visible = freshnessInfo.isStale,
        enter = fadeIn(),
        exit = fadeOut()
    ) {
        Surface(
            color = MaterialTheme.colorScheme.secondaryContainer,
            modifier = modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(8.dp))
        ) {
            Row(
                modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    Icon(
                        Icons.Filled.Update,
                        contentDescription = null,
                        modifier = Modifier.size(14.dp),
                        tint = MaterialTheme.colorScheme.onSecondaryContainer
                    )
                    Text(
                        text = "Last updated ${freshnessInfo.staleDuration}",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSecondaryContainer
                    )
                }

                TextButton(
                    onClick = onRefresh,
                    contentPadding = PaddingValues(horizontal = 8.dp, vertical = 0.dp)
                ) {
                    Text(
                        text = "Refresh",
                        style = MaterialTheme.typography.labelSmall
                    )
                }
            }
        }
    }
}
```

---

## 20. Offline Mode Toggle

### Offline Mode Preference Screen

```kotlin
@Composable
fun OfflineSettingsScreen(
    viewModel: OfflineSettingsViewModel = hiltViewModel()
) {
    val settings by viewModel.settings.collectAsStateWithLifecycle()
    val syncState by viewModel.syncState.collectAsStateWithLifecycle()
    val pendingCount by viewModel.pendingSyncCount.collectAsStateWithLifecycle()

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        // Offline Mode Toggle
        item {
            SettingsSwitchItem(
                title = "Offline Mode",
                subtitle = "Cache data for offline use",
                checked = settings.offlineModeEnabled,
                onCheckedChange = viewModel::toggleOfflineMode
            )
        }

        // Auto Sync
        item {
            SettingsSwitchItem(
                title = "Auto Sync",
                subtitle = "Automatically sync when online",
                checked = settings.autoSync,
                onCheckedChange = viewModel::toggleAutoSync,
                enabled = settings.offlineModeEnabled
            )
        }

        // Sync Interval
        item {
            SettingsSliderItem(
                title = "Sync Interval",
                subtitle = "Sync every ${settings.syncIntervalMinutes} minutes",
                value = settings.syncIntervalMinutes.toFloat(),
                valueRange = 5f..60f,
                onValueChange = viewModel::updateSyncInterval,
                enabled = settings.offlineModeEnabled && settings.autoSync
            )
        }

        // WiFi Only
        item {
            SettingsSwitchItem(
                title = "WiFi Only Sync",
                subtitle = "Only sync on WiFi connections",
                checked = settings.wifiOnlySync,
                onCheckedChange = viewModel::toggleWifiOnlySync,
                enabled = settings.offlineModeEnabled
            )
        }

        // Max Cache Size
        item {
            SettingsSliderItem(
                title = "Max Cache Size",
                subtitle = "${settings.maxCacheSizeMb} MB",
                value = settings.maxCacheSizeMb.toFloat(),
                valueRange = 100f..2000f,
                onValueChange = viewModel::updateMaxCacheSize,
                enabled = settings.offlineModeEnabled
            )
        }

        // Retention
        item {
            SettingsSliderItem(
                title = "Data Retention",
                subtitle = "Keep data for ${settings.retentionDays} days",
                value = settings.retentionDays.toFloat(),
                valueRange = 7f..365f,
                onValueChange = viewModel::updateRetentionDays,
                enabled = settings.offlineModeEnabled
            )
        }

        // Sync Status Section
        item {
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = "Sync Status",
                style = MaterialTheme.typography.titleMedium,
                modifier = Modifier.padding(bottom = 8.dp)
            )
        }

        item {
            SyncStatusCard(
                syncState = syncState,
                pendingCount = pendingCount,
                onForceSync = viewModel::forceSyncNow,
                onClearQueue = viewModel::clearSyncQueue
            )
        }
    }
}

@Composable
fun SyncStatusCard(
    syncState: SyncState,
    pendingCount: Int,
    onForceSync: () -> Unit,
    onClearQueue: () -> Unit,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Pending Operations",
                    style = MaterialTheme.typography.bodyMedium
                )
                Badge {
                    Text(text = pendingCount.toString())
                }
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Status",
                    style = MaterialTheme.typography.bodyMedium
                )
                Text(
                    text = when (syncState) {
                        is SyncState.Idle -> "Idle"
                        is SyncState.Syncing -> "Syncing..."
                        is SyncState.Success -> "Synced"
                        is SyncState.Error -> "Error"
                    },
                    style = MaterialTheme.typography.bodyMedium,
                    color = when (syncState) {
                        is SyncState.Error -> MaterialTheme.colorScheme.error
                        is SyncState.Syncing -> MaterialTheme.colorScheme.primary
                        else -> MaterialTheme.colorScheme.onSurface
                    }
                )
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                OutlinedButton(
                    onClick = onForceSync,
                    modifier = Modifier.weight(1f),
                    enabled = syncState !is SyncState.Syncing
                ) {
                    Icon(Icons.Filled.Sync, contentDescription = null, modifier = Modifier.size(16.dp))
                    Spacer(modifier = Modifier.width(4.dp))
                    Text("Sync Now")
                }

                OutlinedButton(
                    onClick = onClearQueue,
                    modifier = Modifier.weight(1f),
                    colors = ButtonDefaults.outlinedButtonColors(
                        contentColor = MaterialTheme.colorScheme.error
                    )
                ) {
                    Text("Clear Queue")
                }
            }
        }
    }
}
```

---

## 21. Offline Storage Limits

### Storage Manager

```kotlin
class OfflineStorageManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val documentDao: DocumentDao,
    private val conversationDao: ConversationDao,
    private val messageDao: MessageDao,
    private val syncQueueDao: SyncQueueDao,
    private val userSettingsRepository: UserSettingsRepository
) {
    data class StorageInfo(
        val totalUsedBytes: Long,
        val documentBytes: Long,
        val databaseBytes: Long,
        val cacheBytes: Long,
        val maxBytes: Long,
        val usagePercentage: Float
    )

    suspend fun getStorageInfo(): StorageInfo {
        val settings = userSettingsRepository.settings.first()
        val maxBytes = settings.maxCacheSizeMb.toLong() * 1024 * 1024

        val dbFile = context.getDatabasePath(NexusDatabase.DATABASE_NAME)
        val databaseBytes = if (dbFile.exists()) dbFile.length() else 0L

        val cacheDir = context.cacheDir
        val cacheBytes = calculateDirSize(cacheDir)

        val documentBytes = calculateDocumentStorage()

        val totalUsed = databaseBytes + cacheBytes + documentBytes

        return StorageInfo(
            totalUsedBytes = totalUsed,
            documentBytes = documentBytes,
            databaseBytes = databaseBytes,
            cacheBytes = cacheBytes,
            maxBytes = maxBytes,
            usagePercentage = (totalUsed.toFloat() / maxBytes).coerceIn(0f, 1f)
        )
    }

    suspend fun enforceStorageLimits() {
        val settings = userSettingsRepository.settings.first()
        val maxBytes = settings.maxCacheSizeMb.toLong() * 1024 * 1024
        val info = getStorageInfo()

        if (info.totalUsedBytes <= maxBytes) return

        val bytesToFree = info.totalUsedBytes - maxBytes

        // 1. Clear cache first
        clearCache(bytesToFree)

        // 2. Purge old conversations if auto-delete enabled
        if (settings.autoDeleteOldConversations) {
            purgeOldConversations(settings.retentionDays)
        }

        // 3. Purge completed sync queue entries
        syncQueueDao.purgeCompleted()
        syncQueueDao.purgeExhaustedRetries()

        // 4. If still over limit, delete oldest un-pinned conversations
        val remaining = getStorageInfo().totalUsedBytes
        if (remaining > maxBytes) {
            deleteOldestConversations(maxBytes - remaining)
        }
    }

    private suspend fun calculateDocumentStorage(): Long {
        val dir = File(context.filesDir, "documents")
        return if (dir.exists()) calculateDirSize(dir) else 0L
    }

    private fun calculateDirSize(dir: File): Long {
        var size = 0L
        dir.listFiles()?.forEach { file ->
            size += if (file.isDirectory) {
                calculateDirSize(file)
            } else {
                file.length()
            }
        }
        return size
    }

    private suspend fun clearCache(targetBytes: Long) {
        var freed = 0L
        val cacheDir = context.cacheDir

        cacheDir.listFiles()
            ?.sortedBy { it.lastModified() }
            ?.forEach { file ->
                if (freed >= targetBytes) return
                val fileSize = file.length()
                if (file.delete()) {
                    freed += fileSize
                }
            }
    }

    private suspend fun purgeOldConversations(retentionDays: Int) {
        val cutoff = System.currentTimeMillis() - (retentionDays.toLong() * 24 * 60 * 60 * 1000)
        conversationDao.getConversationsPaged(Int.MAX_VALUE, 0).first()
            .filter { it.updatedAt < cutoff && !it.isPinned }
            .forEach { conversation ->
                messageDao.deleteAllByConversation(conversation.id)
                conversationDao.hardDelete(conversation.id)
            }
    }

    private suspend fun deleteOlgestConversations(bytesNeeded: Long) {
        var freed = 0L
        conversationDao.getConversationsPaged(Int.MAX_VALUE, 0).first()
            .filter { !it.isPinned }
            .sortedBy { it.updatedAt }
            .forEach { conversation ->
                if (freed >= bytesNeeded) return
                val convSize = estimateConversationSize(conversation.id)
                messageDao.deleteAllByConversation(conversation.id)
                conversationDao.hardDelete(conversation.id)
                freed += convSize
            }
    }

    private suspend fun estimateConversationSize(conversationId: String): Long {
        val messages = messageDao.getMessagesByConversation(conversationId).first()
        return messages.sumOf { it.content.length.toLong() * 2 } // Approximate
    }
}
```

---

## 22. Offline Data Encryption

### Room Encryption Setup

```kotlin
object DatabaseEncryption {

    fun provideEncryptedDatabase(
        context: Context,
        passphrase: ByteArray
    ): SupportFactory {
        val factory = SupportFactory(passphrase)
        return factory
    }

    fun getOrCreatePassphrase(context: Context): ByteArray {
        val keyStore = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }

        if (!keyStore.containsAlias("db_passphrase_key")) {
            val keyGenerator = KeyGenerator.getInstance(
                KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore"
            )
            val spec = KeyGenParameterSpec.Builder(
                "db_passphrase_key",
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
            )
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .build()
            keyGenerator.init(spec)
            keyGenerator.generateKey()
        }

        val keyStoreEntry = keyStore.getEntry("db_passphrase_key", null) as KeyStore.SecretKeyEntry
        val secretKey = keyStoreEntry.secretKey

        // Check if passphrase exists in EncryptedSharedPreferences
        val masterKey = MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()

        val prefs = EncryptedSharedPreferences.create(
            context, "db_crypto", masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        )

        val existing = prefs.getString("db_passphrase", null)
        if (existing != null) {
            return Base64.decode(existing, Base64.NO_WRAP)
        }

        // Generate new passphrase
        val passphrase = ByteArray(32).also { SecureRandom().nextBytes(it) }

        // Encrypt and store
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, secretKey)
        val encrypted = cipher.doFinal(passphrase)
        val iv = cipher.iv

        prefs.edit().putString(
            "db_passphrase",
            Base64.encodeToString(iv + encrypted, Base64.NO_WRAP)
        ).apply()

        return passphrase
    }
}
```

### File Encryption

```kotlin
class FileEncryptionManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val keyStoreManager = KeyStoreManager()

    fun encryptFile(inputUri: Uri, outputUri: Uri) {
        val key = keyStoreManager.getSecretKey() ?: keyStoreManager.generateSecretKey()
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, key)

        val iv = cipher.iv
        val outputStream = context.contentResolver.openOutputStream(outputUri) ?: return

        outputStream.use { os ->
            // Write IV header
            os.write(iv.size)
            os.write(iv)

            val inputStream = context.contentResolver.openInputStream(inputUri) ?: return
            val buffer = ByteArray(8192)
            var bytesRead: Int

            inputStream.use { input ->
                val cipherOutputStream = CipherOutputStream(os, cipher)
                cipherOutputStream.use { cos ->
                    while (input.read(buffer).also { bytesRead = it } != -1) {
                        cos.write(buffer, 0, bytesRead)
                    }
                }
            }
        }
    }

    fun decryptFile(inputUri: Uri, outputUri: Uri) {
        val key = keyStoreManager.getSecretKey() ?: throw IllegalStateException("Key not found")
        val inputStream = context.contentResolver.openInputStream(inputUri) ?: return

        inputStream.use { input ->
            // Read IV header
            val ivSize = input.read()
            val iv = ByteArray(ivSize)
            input.read(iv)

            val spec = GCMParameterSpec(128, iv)
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")
            cipher.init(Cipher.DECRYPT_MODE, key, spec)

            val outputStream = context.contentResolver.openOutputStream(outputUri) ?: return
            val cipherInputStream = CipherInputStream(input, cipher)

            outputStream.use { os ->
                cipherInputStream.use { cis ->
                    cis.copyTo(os)
                }
            }
        }
    }
}
```

---

## 23. Offline Testing

### Testing Strategies

```kotlin
// Flight mode simulation in tests
class OfflineTestHelper {

    companion object {
        fun simulateOffline(context: Context) {
            val connectivityManager = context.getSystemService<ConnectivityManager>()
            // In instrumented tests, use Espresso or mock the connectivity
        }

        fun simulateOnline(context: Context) {
            val connectivityManager = context.getSystemService<ConnectivityManager>()
            // Restore connectivity
        }
    }
}

// Unit test for offline repository
@RunWith(AndroidJUnit4::class)
class OfflineRepositoryTest {

    private lateinit var database: NexusDatabase
    private lateinit var conversationDao: ConversationDao
    private lateinit var repository: ConversationRepository

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(context, NexusDatabase::class.java)
            .allowMainThreadQueries()
            .build()
        conversationDao = database.conversationDao()
        repository = ConversationRepository(conversationDao, mockApiService)
    }

    @After
    fun teardown() {
        database.close()
    }

    @Test
    fun `returns cached data when offline`() = runTest {
        // Given: cached conversation exists
        val conversation = createTestConversation(id = "1", title = "Test")
        conversationDao.insertConversation(conversation)

        // When: fetching with offline strategy
        val result = repository.get(CacheStrategy.CACHE_ONLY)

        // Then: returns cached data
        assertTrue(result.isSuccess)
        assertEquals("Test", result.getOrNull()?.title)
    }

    @Test
    fun `queues operation when offline`() = runTest {
        // Given: a new conversation
        val conversation = createTestConversation(id = "2", title = "New Chat")

        // When: saving while offline
        repository.saveOffline(conversation)

        // Then: saved to local DB with pending sync
        val local = conversationDao.getConversationByIdOnce("2")
        assertNotNull(local)
        assertEquals(SyncStatus.PENDING, local?.syncStatus)
    }

    @Test
    fun `syncs queued operations when back online`() = runTest {
        // Given: queued operations
        val conversation = createTestConversation(id = "3", title = "Queued")
        conversationDao.insertConversation(conversation.copy(syncStatus = SyncStatus.PENDING))

        // When: sync is triggered
        val result = repository.syncPending()

        // Then: operations are synced
        assertTrue(result.isSuccess)
    }

    @Test
    fun `handles conflict with last-write-wins`() = runTest {
        // Given: local and remote versions of same conversation
        val local = createTestConversation(id = "4", title = "Local", updatedAt = 1000L)
        val remote = createTestConversation(id = "4", title = "Remote", updatedAt = 2000L)

        // When: resolving conflict
        val resolver = ConflictResolver()
        val resolved = resolver.resolve(local, remote, ConflictStrategy.LAST_WRITE_WINS)

        // Then: newer version wins
        assertEquals("Remote", resolved.title)
    }

    private fun createTestConversation(
        id: String,
        title: String,
        updatedAt: Long = System.currentTimeMillis()
    ) = ConversationEntity(
        id = id,
        title = title,
        agentId = "agent-1",
        updatedAt = updatedAt
    )
}
```

### Sync Integration Test

```kotlin
@RunWith(AndroidJUnit4::class)
class SyncIntegrationTest {

    private lateinit var syncManager: SyncManager
    private lateinit var database: NexusDatabase
    private lateinit var mockApi: MockApiService

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(context, NexusDatabase::class.java)
            .allowMainThreadQueries()
            .build()
        mockApi = MockApiService()
        syncManager = SyncManager(
            syncQueueDao = database.syncQueueDao(),
            conversationDao = database.conversationDao(),
            messageDao = database.messageDao(),
            documentDao = database.documentDao(),
            agentDao = database.agentDao(),
            apiService = mockApi,
            networkMonitor = FakeNetworkMonitor(isOnline = true),
            keyStoreManager = mockKeyStoreManager
        )
    }

    @Test
    fun `full sync cycle`() = runTest {
        // 1. Create local data
        val conversation = ConversationEntity(
            id = "test-1", title = "Test Chat",
            agentId = "agent-1", syncStatus = SyncStatus.PENDING
        )
        database.conversationDao().insertConversation(conversation)

        // 2. Add to sync queue
        database.syncQueueDao().enqueue(
            EntityType.CONVERSATION, "test-1", SyncOperation.CREATE,
            Json.encodeToString(conversation.toDto())
        )

        // 3. Perform sync
        val result = syncManager.performSync()

        // 4. Verify
        assertTrue(result is SyncResult.Success)
        val synced = database.conversationDao().getConversationByIdOnce("test-1")
        assertEquals(SyncStatus.SYNCED, synced?.syncStatus)
        assertTrue(mockApi.wasCalled("createConversation"))
    }
}
```

---

## 24. Offline UX Patterns

### UX Pattern Summary

| Pattern | When to Use | Component |
|---------|-------------|-----------|
| Offline Banner | Device is offline | `OfflineBanner` |
| Sync Progress | Sync in progress | `SyncStatusBar` |
| Queued Indicator | Message queued | `MessageBubble.isQueued` |
| Stale Data | Data older than threshold | `StaleDataIndicator` |
| Retry Button | Sync failed | `RetryButton` |
| Pending Badge | Queue has items | `Badge` on nav item |
| Error Toast | Operation failed | `Snackbar` with action |
| Skeleton Loading | Initial load | `SkeletonLoader` |
| Empty State | No cached data | `EmptyState` |

### Retry Button Component

```kotlin
@Composable
fun RetryButton(
    onRetry: () -> Unit,
    isLoading: Boolean,
    modifier: Modifier = Modifier
) {
    var retryCount by remember { mutableIntStateOf(0) }

    Button(
        onClick = {
            retryCount++
            onRetry()
        },
        modifier = modifier,
        enabled = !isLoading
    ) {
        if (isLoading) {
            CircularProgressIndicator(
                modifier = Modifier.size(16.dp),
                strokeWidth = 2.dp,
                color = MaterialTheme.colorScheme.onPrimary
            )
        } else {
            Icon(
                Icons.Filled.Refresh,
                contentDescription = "Retry",
                modifier = Modifier.size(16.dp)
            )
            Spacer(modifier = Modifier.width(4.dp))
            Text(
                text = if (retryCount > 0) "Retry ($retryCount)" else "Retry"
            )
        }
    }
}
```

---

## 25. Offline Error Handling

### Error Types and Handling

```kotlin
sealed class OfflineError(
    val message: String,
    val userMessage: String,
    val isRecoverable: Boolean = true
) {
    data object NetworkUnavailable : OfflineError(
        message = "Network connection is unavailable",
        userMessage = "You're offline. Changes will be synced when you reconnect."
    )

    data object SyncFailed : OfflineError(
        message = "Sync operation failed",
        userMessage = "Failed to sync. Will retry automatically."
    )

    data object StorageFull : OfflineError(
        message = "Local storage limit exceeded",
        userMessage = "Storage is full. Please clear some cached data.",
        isRecoverable = false
    )

    data object ConflictDetected : OfflineError(
        message = "Sync conflict detected",
        userMessage = "Some changes conflict with the server version."
    )

    data class UploadFailed(val fileName: String) : OfflineError(
        message = "Failed to upload $fileName",
        userMessage = "Could not upload $fileName. Will retry when online."
    )

    data class DatabaseError(val cause: Throwable) : OfflineError(
        message = "Database error: ${cause.message}",
        userMessage = "Something went wrong. Please try again.",
        isRecoverable = false
    )

    data class EncryptionError(val cause: Throwable) : OfflineError(
        message = "Encryption error: ${cause.message}",
        userMessage = "Security error. Please restart the app.",
        isRecoverable = false
    )
}

class OfflineErrorHandler @Inject constructor(
    private val syncQueueManager: SyncQueueManager,
    private val storageManager: OfflineStorageManager
) {
    suspend fun handleError(error: OfflineError): ErrorAction {
        return when (error) {
            is OfflineError.NetworkUnavailable -> {
                ErrorAction.ShowBanner(error.userMessage)
            }
            is OfflineError.SyncFailed -> {
                ErrorAction.RetryAutomatically
            }
            is OfflineError.StorageFull -> {
                storageManager.enforceStorageLimits()
                ErrorAction.ShowDialog(error.userMessage)
            }
            is OfflineError.ConflictDetected -> {
                ErrorAction.ShowConflictResolution
            }
            is OfflineError.UploadFailed -> {
                ErrorAction.QueueForRetry
            }
            is OfflineError.DatabaseError -> {
                ErrorAction.ShowDialog(error.userMessage)
            }
            is OfflineError.EncryptionError -> {
                ErrorAction.ShowDialog(error.userMessage)
            }
        }
    }
}

sealed class ErrorAction {
    data class ShowBanner(val message: String) : ErrorAction()
    data class ShowDialog(val message: String) : ErrorAction()
    data object RetryAutomatically : ErrorAction()
    data object QueueForRetry : ErrorAction()
    data object ShowConflictResolution : ErrorAction()
}
```

---

## 26. Offline Performance

### Query Optimization

```kotlin
// Index strategy for offline performance
@Entity(
    tableName = "messages",
    indices = [
        // Primary query pattern: messages by conversation, ordered by time
        Index(value = ["conversation_id", "timestamp", "is_deleted"]),

        // Sync status for batch operations
        Index(value = ["sync_status", "is_deleted"]),

        // Search optimization
        Index(value = ["content"]),

        // Role-based filtering
        Index(value = ["role", "conversation_id"])
    ]
)

// Batch operations for bulk sync
@Dao
interface OptimizedMessageDao {

    @Query("""
        UPDATE messages 
        SET sync_status = 'SYNCED' 
        WHERE id IN (:ids)
    """)
    suspend fun markMultipleAsSynced(ids: List<String>)

    @Query("""
        UPDATE messages 
        SET sync_status = 'SYNCED' 
        WHERE conversation_id = :conversationId 
        AND sync_status != 'SYNCED'
    """)
    suspend fun markConversationAsSynced(conversationId: String)

    @Transaction
    suspend fun batchInsert(messages: List<MessageEntity>) {
        messages.chunked(500).forEach { chunk ->
            insertMessages(chunk)
        }
    }

    @Query("DELETE FROM messages WHERE is_deleted = 1 AND updated_at < :before")
    suspend fun purgeOldDeleted(before: Long)
}

// Pagination for large datasets
@Dao
interface PaginatedDao {

    @Query("""
        SELECT * FROM messages 
        WHERE conversation_id = :conversationId 
        AND is_deleted = 0 
        ORDER BY timestamp DESC 
        LIMIT :limit 
        OFFSET :offset
    """)
    suspend fun getMessagesPage(
        conversationId: String,
        limit: Int = 50,
        offset: Int = 0
    ): List<MessageEntity>

    @Query("SELECT COUNT(*) FROM messages WHERE conversation_id = :conversationId AND is_deleted = 0")
    suspend fun getMessageCount(conversationId: String): Int
}
```

### Performance Metrics

| Operation | Target | Optimization |
|-----------|--------|--------------|
| Local query | < 50ms | Proper indexing |
| Batch insert | < 200ms per 500 | Chunked transactions |
| Sync push | < 5s per batch | Parallel uploads |
| Search | < 100ms | Full-text index |
| Cache lookup | < 10ms | In-memory LRU |
| DB open | < 100ms | WAL mode |

---

## 27. Offline Accessibility

### Accessible Offline Components

```kotlin
@Composable
fun AccessibleOfflineBanner(
    isOnline: Boolean,
    modifier: Modifier = Modifier
) {
    AnimatedVisibility(
        visible = !isOnline,
        enter = slideInVertically(),
        exit = slideOutVertically()
    ) {
        Surface(
            color = MaterialTheme.colorScheme.errorContainer,
            modifier = modifier.semantics {
                liveRegion = LiveRegionMode.Polite
                contentDescription = "You are currently offline. " +
                    "Any changes you make will be saved locally and " +
                    "synced when your connection is restored."
            }
        ) {
            Row(
                modifier = Modifier.padding(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    Icons.Filled.CloudOff,
                    contentDescription = null, // Decorative; parent has contentDescription
                    modifier = Modifier.size(20.dp)
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "Offline",
                    style = MaterialTheme.typography.labelLarge,
                    color = MaterialTheme.colorScheme.onErrorContainer
                )
            }
        }
    }
}

@Composable
fun AccessibleSyncProgress(
    progress: SyncProgress,
    modifier: Modifier = Modifier
) {
    val percentage = (progress.percentage * 100).toInt()

    LinearProgressIndicator(
        progress = { progress.percentage },
        modifier = modifier.semantics {
            liveRegion = LiveRegionMode.Polite
            contentDescription = "Syncing data. " +
                "$percentage percent complete. " +
                "${progress.completed} of ${progress.total} items synced."
            progress = progress.percentage
        }
    )
}

@Composable
fun AccessibleQueuedMessage(
    message: MessageEntity,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier.semantics {
            contentDescription = "Message from ${message.role.name.lowercase()}: " +
                "${message.content}. " +
                "Status: ${when (message.syncStatus) {
                    SyncStatus.PENDING -> "Queued to send"
                    SyncStatus.SYNCED -> "Sent"
                    SyncStatus.FAILED -> "Failed to send. Tap to retry."
                    SyncStatus.CONFLICT -> "Has a conflict. Needs resolution."
                }}"
        }
    ) {
        // Message content...
    }
}
```

### Accessibility Checklist

| Requirement | Implementation | Status |
|-------------|---------------|--------|
| TalkBack labels | `contentDescription` on all interactive elements | Done |
| Live regions | `liveRegion` on status changes | Done |
| State announcements | `semantics { liveRegion = LiveRegionMode.Polite }` | Done |
| Touch targets | Minimum 48x48dp | Done |
| Color contrast | WCAG AA (4.5:1) | Done |
| Font scaling | `sp` units for text | Done |
| Reduce motion | Respect `reduceMotion` setting | Done |
| Screen reader order | Logical reading order in Compose | Done |
| Error messages | Announced via `liveRegion` | Done |
| Action confirmation | Snackbar with undo action | Done |

---

## Summary

The offline mode architecture for Nexus AI provides:

| Component | Purpose | Technology |
|-----------|---------|------------|
| **Room Database** | Persistent local storage | Room + KSP |
| **DataStore** | Key-value user settings | DataStore Preferences |
| **EncryptedSP** | Secure token storage | AndroidX Security |
| **WorkManager** | Background sync scheduling | AndroidX Work |
| **NetworkMonitor** | Connectivity detection | ConnectivityManager |
| **SyncManager** | Push/pull synchronization | Custom + Retrofit |
| **ConflictResolver** | Merge conflict handling | Custom logic |
| **StorageManager** | Cache size enforcement | Custom + Room |
| **OfflineSearch** | Local full-text search | Room FTS |

The system ensures users can work seamlessly offline while maintaining data consistency through intelligent caching, conflict resolution, and automatic background synchronization.
