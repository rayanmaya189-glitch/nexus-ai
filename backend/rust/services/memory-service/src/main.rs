use axum::{
    extract::{Path, State},
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tower_http::cors::{Any, CorsLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use nexus_core::config::AppConfig;
use nexus_core::error::{NexusError, NexusResult};

#[derive(Clone)]
#[allow(dead_code)]
struct AppState {
    config: Arc<AppConfig>,
    entries: Arc<RwLock<Vec<MemoryEntry>>>,
    stats: Arc<RwLock<MemoryStats>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct MemoryEntry {
    id: String,
    tenant_id: i64,
    user_id: i64,
    agent_id: i64,
    session_id: Option<String>,
    memory_type: String,
    content: String,
    summary: Option<String>,
    metadata: serde_json::Value,
    importance: f64,
    access_count: i32,
    last_accessed: Option<String>,
    expires_at: Option<String>,
    created_at: String,
}

#[derive(Debug, Deserialize)]
struct CreateMemoryRequest {
    tenant_id: i64,
    user_id: i64,
    agent_id: i64,
    session_id: Option<String>,
    memory_type: String,
    content: String,
    importance: Option<f64>,
    ttl_seconds: Option<i64>,
    metadata: Option<serde_json::Value>,
}

#[derive(Debug, Deserialize)]
struct SearchMemoryRequest {
    tenant_id: i64,
    user_id: Option<i64>,
    agent_id: Option<i64>,
    query: String,
    memory_type: Option<String>,
    limit: Option<usize>,
}

#[derive(Debug, Deserialize)]
struct ContextRequest {
    query: String,
    max_tokens: Option<usize>,
}

#[derive(Debug, Serialize)]
struct MemoryStats {
    total_entries: usize,
    by_type: HashMap<String, usize>,
    avg_importance: f64,
    total_size_bytes: usize,
}

#[derive(Debug, Serialize)]
struct ContextResult {
    memories: Vec<MemoryEntry>,
    total_tokens_estimate: usize,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::try_from_default_env()
            .unwrap_or_else(|_| "memory-service=debug,tower_http=debug".into()))
        .with(tracing_subscriber::fmt::layer())
        .init();

    let config = AppConfig::from_file(
        &std::env::var("CONFIG_PATH").unwrap_or_else(|_| "configs/memory-service.json".to_string()),
    )?;

    let state = AppState {
        config: Arc::new(config.clone()),
        entries: Arc::new(RwLock::new(Vec::new())),
        stats: Arc::new(RwLock::new(MemoryStats {
            total_entries: 0,
            by_type: HashMap::new(),
            avg_importance: 0.0,
            total_size_bytes: 0,
        })),
    };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/api/v1/memory/entries", get(list_entries).post(create_entry))
        .route("/api/v1/memory/entries/{id}", get(get_entry).delete(delete_entry))
        .route("/api/v1/memory/search", post(search_entries))
        .route("/api/v1/memory/consolidate", post(consolidate_memories))
        .route("/api/v1/memory/stats", get(get_stats))
        .route("/api/v1/memory/sessions/{id}", post(get_session_memory))
        .route("/api/v1/memory/context/{agent_id}", post(build_context))
        .layer(cors)
        .with_state(state);

    let addr = format!("0.0.0.0:{}", config.server.http_port);
    tracing::info!("Memory Service starting on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({
        "status": "healthy",
        "service": "memory-service",
        "version": "1.0.0"
    }))
}

async fn create_entry(
    State(state): State<AppState>,
    Json(req): Json<CreateMemoryRequest>,
) -> NexusResult<impl IntoResponse> {
    let entry_id = uuid::Uuid::new_v4().to_string();
    let now = chrono::Utc::now().to_rfc3339();

    let expires_at = req.ttl_seconds.map(|ttl| {
        (chrono::Utc::now() + chrono::Duration::seconds(ttl)).to_rfc3339()
    });

    let entry = MemoryEntry {
        id: entry_id.clone(),
        tenant_id: req.tenant_id,
        user_id: req.user_id,
        agent_id: req.agent_id,
        session_id: req.session_id,
        memory_type: req.memory_type.clone(),
        content: req.content.clone(),
        summary: None,
        metadata: req.metadata.unwrap_or_default(),
        importance: req.importance.unwrap_or(0.5),
        access_count: 0,
        last_accessed: None,
        expires_at,
        created_at: now,
    };

    let size = req.content.len();
    state.entries.write().await.push(entry.clone());

    let mut stats = state.stats.write().await;
    stats.total_entries += 1;
    *stats.by_type.entry(req.memory_type).or_insert(0) += 1;
    stats.total_size_bytes += size;

    Ok(Json(serde_json::json!({"data": entry})))
}

async fn list_entries(
    State(state): State<AppState>,
) -> NexusResult<impl IntoResponse> {
    let entries = state.entries.read().await;
    Ok(Json(serde_json::json!({"data": *entries})))
}

async fn get_entry(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> NexusResult<impl IntoResponse> {
    let mut entries = state.entries.write().await;
    let entry = entries.iter_mut().find(|e| e.id == id);

    match entry {
        Some(e) => {
            e.access_count += 1;
            e.last_accessed = Some(chrono::Utc::now().to_rfc3339());
            Ok(Json(serde_json::json!({"data": e.clone()})))
        }
        None => Err(NexusError::NotFound("Memory entry not found".to_string())),
    }
}

async fn delete_entry(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> NexusResult<impl IntoResponse> {
    let mut entries = state.entries.write().await;
    let before = entries.len();
    entries.retain(|e| e.id != id);

    if entries.len() == before {
        return Err(NexusError::NotFound("Memory entry not found".to_string()));
    }

    Ok(Json(serde_json::json!({"data": {"deleted": true}})))
}

async fn search_entries(
    State(state): State<AppState>,
    Json(req): Json<SearchMemoryRequest>,
) -> NexusResult<impl IntoResponse> {
    let limit = req.limit.unwrap_or(10);
    let entries = state.entries.read().await;

    let mut results: Vec<&MemoryEntry> = entries
        .iter()
        .filter(|e| e.tenant_id == req.tenant_id)
        .filter(|e| {
            if let Some(uid) = req.user_id {
                e.user_id == uid
            } else {
                true
            }
        })
        .filter(|e| {
            if let Some(aid) = req.agent_id {
                e.agent_id == aid
            } else {
                true
            }
        })
        .filter(|e| {
            if let Some(ref mt) = req.memory_type {
                e.memory_type == *mt
            } else {
                true
            }
        })
        .filter(|e| e.content.to_lowercase().contains(&req.query.to_lowercase()))
        .collect();

    results.sort_by(|a, b| b.importance.partial_cmp(&a.importance).unwrap());
    results.truncate(limit);

    let hits: Vec<&MemoryEntry> = results;
    Ok(Json(serde_json::json!({
        "data": {
            "results": hits,
            "total": hits.len(),
        }
    })))
}

async fn consolidate_memories(
    State(state): State<AppState>,
) -> NexusResult<impl IntoResponse> {
    let entries = state.entries.read().await;
    let mut consolidated = 0;
    let mut merged = 0;

    let mut type_groups: HashMap<String, Vec<&MemoryEntry>> = HashMap::new();
    for entry in entries.iter() {
        let key = format!("{}:{}:{}", entry.tenant_id, entry.agent_id, entry.memory_type);
        type_groups.entry(key).or_default().push(entry);
    }

    for (_key, group) in &type_groups {
        if group.len() > 1 {
            consolidated += 1;
            merged += group.len() - 1;
        }
    }

    Ok(Json(serde_json::json!({
        "data": {
            "consolidated_groups": consolidated,
            "entries_merged": merged,
            "remaining_entries": entries.len() - merged,
        }
    })))
}

async fn get_session_memory(
    State(state): State<AppState>,
    Path(session_id): Path<String>,
) -> NexusResult<impl IntoResponse> {
    let entries = state.entries.read().await;
    let session_entries: Vec<&MemoryEntry> = entries
        .iter()
        .filter(|e| e.session_id.as_ref() == Some(&session_id))
        .collect();

    Ok(Json(serde_json::json!({
        "data": {
            "session_id": session_id,
            "memories": session_entries,
            "count": session_entries.len(),
        }
    })))
}

async fn build_context(
    State(state): State<AppState>,
    Path(agent_id): Path<i64>,
    Json(req): Json<ContextRequest>,
) -> NexusResult<impl IntoResponse> {
    let max_tokens = req.max_tokens.unwrap_or(4000);
    let entries = state.entries.read().await;

    let mut relevant: Vec<&MemoryEntry> = entries
        .iter()
        .filter(|e| e.agent_id == agent_id)
        .filter(|e| e.expires_at.is_none() || {
            if let Some(ref exp) = e.expires_at {
                exp > &chrono::Utc::now().to_rfc3339()
            } else {
                true
            }
        })
        .filter(|e| e.content.to_lowercase().contains(&req.query.to_lowercase()) || e.importance >= 0.7)
        .collect();

    relevant.sort_by(|a, b| b.importance.partial_cmp(&a.importance).unwrap());

    let mut selected = Vec::new();
    let mut token_estimate = 0;

    for entry in relevant {
        let entry_tokens = entry.content.len() / 4;
        if token_estimate + entry_tokens > max_tokens {
            break;
        }
        token_estimate += entry_tokens;
        selected.push(entry.clone());
    }

    Ok(Json(serde_json::json!({
        "data": ContextResult {
            memories: selected,
            total_tokens_estimate: token_estimate,
        }
    })))
}

async fn get_stats(
    State(state): State<AppState>,
) -> NexusResult<impl IntoResponse> {
    let stats_lock = state.stats.read().await;
    let stats = MemoryStats {
        total_entries: stats_lock.total_entries,
        by_type: stats_lock.by_type.clone(),
        avg_importance: stats_lock.avg_importance,
        total_size_bytes: stats_lock.total_size_bytes,
    };
    Ok(Json(serde_json::json!({"data": stats})))
}
