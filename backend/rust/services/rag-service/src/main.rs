use axum::{
    extract::{Path, State},
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;
use tower_http::cors::{Any, CorsLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use nexus_core::config::AppConfig;
use nexus_core::error::{NexusError, NexusResult};

#[derive(Clone)]
struct AppState {
    config: Arc<AppConfig>,
    http_client: reqwest::Client,
    documents: Arc<RwLock<Vec<InMemoryDocument>>>,
    document_sets: Arc<RwLock<Vec<DocumentSet>>>,
    stats: Arc<RwLock<RAGStats>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct InMemoryDocument {
    id: String,
    title: String,
    content: String,
    doc_type: String,
    tenant_id: i64,
    chunks: Vec<Chunk>,
    created_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Chunk {
    id: String,
    content: String,
    embedding: Vec<f32>,
    chunk_index: usize,
    metadata: serde_json::Value,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct DocumentSet {
    id: String,
    name: String,
    description: Option<String>,
    tenant_id: i64,
    document_ids: Vec<String>,
    created_at: String,
}

#[derive(Debug, Deserialize)]
struct SearchQuery {
    query: String,
    limit: Option<usize>,
    document_set_id: Option<String>,
    min_score: Option<f64>,
}

#[derive(Debug, Serialize)]
struct SearchHit {
    chunk_id: String,
    document_id: String,
    content: String,
    score: f64,
    metadata: serde_json::Value,
}

#[derive(Debug, Deserialize)]
struct RAGQuery {
    query: String,
    model: Option<String>,
    context_limit: Option<usize>,
    document_set_id: Option<String>,
}

#[derive(Debug, Serialize)]
struct RAGResponse {
    answer: String,
    sources: Vec<SearchHit>,
    model_used: String,
    tokens_context: usize,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct IngestRequest {
    title: String,
    content: String,
    doc_type: Option<String>,
    tenant_id: Option<i64>,
    metadata: Option<serde_json::Value>,
}

#[derive(Debug, Deserialize)]
struct CreateDocSetRequest {
    name: String,
    description: Option<String>,
    tenant_id: i64,
    document_ids: Vec<String>,
}

#[derive(Debug, Default, Clone, Serialize)]
struct RAGStats {
    total_documents: usize,
    total_chunks: usize,
    total_embeddings: usize,
    queries_processed: u64,
}

fn chunk_text(text: &str, chunk_size: usize, overlap: usize) -> Vec<String> {
    let chars: Vec<char> = text.chars().collect();
    let mut chunks = Vec::new();
    let mut start = 0;

    while start < chars.len() {
        let end = std::cmp::min(start + chunk_size, chars.len());
        let chunk: String = chars[start..end].iter().collect();
        chunks.push(chunk);
        if end >= chars.len() {
            break;
        }
        start = end - overlap;
    }
    chunks
}

async fn get_embeddings(
    http_client: &reqwest::Client,
    base_url: &str,
    texts: &[String],
    model: &str,
) -> anyhow::Result<Vec<Vec<f32>>> {
    let mut all_embeddings = Vec::new();

    for text in texts {
        let resp = http_client
            .post(format!("{}/api/embeddings", base_url))
            .json(&serde_json::json!({
                "model": model,
                "prompt": text,
            }))
            .send()
            .await?;

        let body: serde_json::Value = resp.json().await?;
        if let Some(embedding) = body.get("embedding") {
            let vec: Vec<f32> = embedding
                .as_array()
                .unwrap_or(&vec![])
                .iter()
                .filter_map(|v| v.as_f64().map(|f| f as f32))
                .collect();
            all_embeddings.push(vec);
        } else {
            all_embeddings.push(vec![0.0; 384]);
        }
    }

    Ok(all_embeddings)
}

fn cosine_similarity(a: &[f32], b: &[f32]) -> f64 {
    if a.len() != b.len() || a.is_empty() {
        return 0.0;
    }
    let dot: f64 = a.iter().zip(b.iter()).map(|(x, y)| (*x as f64) * (*y as f64)).sum();
    let norm_a: f64 = a.iter().map(|x| (*x as f64).powi(2)).sum::<f64>().sqrt();
    let norm_b: f64 = b.iter().map(|x| (*x as f64).powi(2)).sum::<f64>().sqrt();
    if norm_a == 0.0 || norm_b == 0.0 {
        0.0
    } else {
        dot / (norm_a * norm_b)
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::try_from_default_env()
            .unwrap_or_else(|_| "rag-service=debug,tower_http=debug".into()))
        .with(tracing_subscriber::fmt::layer())
        .init();

    let config = AppConfig::from_file(
        &std::env::var("CONFIG_PATH").unwrap_or_else(|_| "configs/rag-service.json".to_string()),
    )?;

    let http_client = reqwest::Client::builder()
        .timeout(config.ollama.timeout)
        .build()?;

    let state = AppState {
        config: Arc::new(config.clone()),
        http_client,
        documents: Arc::new(RwLock::new(Vec::new())),
        document_sets: Arc::new(RwLock::new(Vec::new())),
        stats: Arc::new(RwLock::new(RAGStats::default())),
    };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/api/v1/rag/documents/ingest", post(ingest_document))
        .route("/api/v1/rag/search", post(search_documents))
        .route("/api/v1/rag/rag-query", post(rag_query))
        .route("/api/v1/rag/documents", get(list_documents))
        .route("/api/v1/rag/documents/{id}", get(get_document).delete(delete_document))
        .route("/api/v1/rag/documents/{id}/chunks", post(rechunk_document))
        .route("/api/v1/rag/document-sets", get(list_document_sets).post(create_document_set))
        .route("/api/v1/rag/stats", get(get_stats))
        .layer(cors)
        .with_state(state);

    let addr = format!("0.0.0.0:{}", config.server.http_port);
    tracing::info!("RAG Service starting on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({
        "status": "healthy",
        "service": "rag-service",
        "version": "1.0.0"
    }))
}

async fn ingest_document(
    State(state): State<AppState>,
    Json(req): Json<IngestRequest>,
) -> NexusResult<impl IntoResponse> {
    let doc_id = uuid::Uuid::new_v4().to_string();
    let chunk_size = 512;
    let overlap = 50;

    let text_chunks = chunk_text(&req.content, chunk_size, overlap);

    let embeddings = get_embeddings(
        &state.http_client,
        &state.config.ollama.base_url,
        &text_chunks,
        "nomic-embed-text",
    )
    .await
    .unwrap_or_else(|_| text_chunks.iter().map(|_| vec![0.0; 384]).collect());

    let chunks: Vec<Chunk> = text_chunks
        .into_iter()
        .enumerate()
        .zip(embeddings.into_iter())
        .map(|((i, content), embedding)| Chunk {
            id: uuid::Uuid::new_v4().to_string(),
            content,
            embedding,
            chunk_index: i,
            metadata: serde_json::json!({"document_id": doc_id}),
        })
        .collect();

    let doc = InMemoryDocument {
        id: doc_id.clone(),
        title: req.title,
        content: req.content,
        doc_type: req.doc_type.unwrap_or_else(|| "text".to_string()),
        tenant_id: req.tenant_id.unwrap_or(1),
        chunks: chunks.clone(),
        created_at: chrono::Utc::now().to_rfc3339(),
    };

    let chunk_count = chunks.len();
    state.documents.write().await.push(doc);

    let mut stats = state.stats.write().await;
    stats.total_documents += 1;
    stats.total_chunks += chunk_count;
    stats.total_embeddings += chunk_count;

    Ok(Json(serde_json::json!({
        "data": {
            "document_id": doc_id,
            "chunks_created": chunk_count,
            "status": "ingested"
        }
    })))
}

async fn search_documents(
    State(state): State<AppState>,
    Json(query): Json<SearchQuery>,
) -> NexusResult<impl IntoResponse> {
    let limit = query.limit.unwrap_or(5);
    let min_score = query.min_score.unwrap_or(0.0);

    let query_embedding = get_embeddings(
        &state.http_client,
        &state.config.ollama.base_url,
        &[query.query.clone()],
        "nomic-embed-text",
    )
    .await
    .unwrap_or_else(|_| vec![vec![0.0; 384]])
    .into_iter()
    .next()
    .unwrap_or_else(|| vec![0.0; 384]);

    let docs = state.documents.read().await;
    let mut hits: Vec<SearchHit> = Vec::new();

    for doc in docs.iter() {
        if let Some(ref set_id) = query.document_set_id {
            let sets = state.document_sets.read().await;
            if !sets.iter().any(|s| &s.id == set_id && s.document_ids.contains(&doc.id)) {
                continue;
            }
        }

        for chunk in &doc.chunks {
            let score = cosine_similarity(&query_embedding, &chunk.embedding);
            if score >= min_score {
                hits.push(SearchHit {
                    chunk_id: chunk.id.clone(),
                    document_id: doc.id.clone(),
                    content: chunk.content.clone(),
                    score,
                    metadata: serde_json::json!({
                        "document_title": doc.title,
                        "chunk_index": chunk.chunk_index,
                    }),
                });
            }
        }
    }

    hits.sort_by(|a, b| b.score.partial_cmp(&a.score).unwrap());
    hits.truncate(limit);

    let mut stats = state.stats.write().await;
    stats.queries_processed += 1;

    Ok(Json(serde_json::json!({
        "data": {
            "hits": hits,
            "total_hits": hits.len(),
        }
    })))
}

async fn rag_query(
    State(state): State<AppState>,
    Json(query): Json<RAGQuery>,
) -> NexusResult<impl IntoResponse> {
    let context_limit = query.context_limit.unwrap_or(5);
    let model = query.model.unwrap_or_else(|| "phi4-mini:3.8b".to_string());

    let search_results = {
        let query_embedding = get_embeddings(
            &state.http_client,
            &state.config.ollama.base_url,
            &[query.query.clone()],
            "nomic-embed-text",
        )
        .await
        .unwrap_or_else(|_| vec![vec![0.0; 384]])
        .into_iter()
        .next()
        .unwrap_or_else(|| vec![0.0; 384]);

        let docs = state.documents.read().await;
        let mut hits: Vec<SearchHit> = Vec::new();

        for doc in docs.iter() {
            if let Some(ref set_id) = query.document_set_id {
                let sets = state.document_sets.read().await;
                if !sets.iter().any(|s| &s.id == set_id && s.document_ids.contains(&doc.id)) {
                    continue;
                }
            }
            for chunk in &doc.chunks {
                let score = cosine_similarity(&query_embedding, &chunk.embedding);
                if score > 0.0 {
                    hits.push(SearchHit {
                        chunk_id: chunk.id.clone(),
                        document_id: doc.id.clone(),
                        content: chunk.content.clone(),
                        score,
                        metadata: serde_json::json!({"document_title": doc.title}),
                    });
                }
            }
        }

        hits.sort_by(|a, b| b.score.partial_cmp(&a.score).unwrap());
        hits.truncate(context_limit);
        hits
    };

    let context = search_results
        .iter()
        .map(|h| h.content.as_str())
        .collect::<Vec<_>>()
        .join("\n\n---\n\n");

    let prompt = format!(
        "Based on the following context, answer the user's question. If the context doesn't contain enough information, say so.\n\nContext:\n{}\n\nQuestion: {}\n\nAnswer:",
        context, query.query
    );

    let ollama_resp = state
        .http_client
        .post(format!("{}/api/generate", state.config.ollama.base_url))
        .json(&serde_json::json!({
            "model": model,
            "prompt": prompt,
            "stream": false,
        }))
        .send()
        .await;

    let answer = match ollama_resp {
        Ok(resp) => {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            body.get("response")
                .and_then(|v| v.as_str())
                .unwrap_or("Unable to generate answer")
                .to_string()
        }
        Err(e) => format!("Error generating response: {}", e),
    };

    let token_estimate = context.len() / 4 + query.query.len() / 4;

    let mut stats = state.stats.write().await;
    stats.queries_processed += 1;

    Ok(Json(serde_json::json!({
        "data": RAGResponse {
            answer,
            sources: search_results,
            model_used: model,
            tokens_context: token_estimate,
        }
    })))
}

async fn list_documents(
    State(state): State<AppState>,
) -> NexusResult<impl IntoResponse> {
    let docs = state.documents.read().await;
    let summaries: Vec<serde_json::Value> = docs
        .iter()
        .map(|d| {
            serde_json::json!({
                "id": d.id,
                "title": d.title,
                "doc_type": d.doc_type,
                "tenant_id": d.tenant_id,
                "chunk_count": d.chunks.len(),
                "created_at": d.created_at,
            })
        })
        .collect();

    Ok(Json(serde_json::json!({"data": summaries})))
}

async fn get_document(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> NexusResult<impl IntoResponse> {
    let docs = state.documents.read().await;
    let doc = docs.iter().find(|d| d.id == id);

    match doc {
        Some(d) => Ok(Json(serde_json::json!({"data": d}))),
        None => Err(NexusError::NotFound("Document not found".to_string())),
    }
}

async fn delete_document(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> NexusResult<impl IntoResponse> {
    let mut docs = state.documents.write().await;
    let before = docs.len();
    docs.retain(|d| d.id != id);

    if docs.len() == before {
        return Err(NexusError::NotFound("Document not found".to_string()));
    }

    Ok(Json(serde_json::json!({"data": {"deleted": true}})))
}

async fn rechunk_document(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<serde_json::Value>,
) -> NexusResult<impl IntoResponse> {
    let chunk_size = req.get("chunk_size").and_then(|v| v.as_u64()).unwrap_or(512) as usize;
    let overlap = req.get("overlap").and_then(|v| v.as_u64()).unwrap_or(50) as usize;

    let mut docs = state.documents.write().await;
    let doc = docs.iter_mut().find(|d| d.id == id);

    match doc {
        Some(d) => {
            let text_chunks = chunk_text(&d.content, chunk_size, overlap);
            let embeddings = get_embeddings(
                &state.http_client,
                &state.config.ollama.base_url,
                &text_chunks,
                "nomic-embed-text",
            )
            .await
            .unwrap_or_else(|_| text_chunks.iter().map(|_| vec![0.0; 384]).collect());

            let old_chunks = d.chunks.len();
            d.chunks = text_chunks
                .into_iter()
                .enumerate()
                .zip(embeddings.into_iter())
                .map(|((i, content), embedding)| Chunk {
                    id: uuid::Uuid::new_v4().to_string(),
                    content,
                    embedding,
                    chunk_index: i,
                    metadata: serde_json::json!({"document_id": d.id}),
                })
                .collect();
            let new_chunks = d.chunks.len();

            Ok(Json(serde_json::json!({
                "data": {
                    "document_id": id,
                    "old_chunk_count": old_chunks,
                    "new_chunk_count": new_chunks,
                }
            })))
        }
        None => Err(NexusError::NotFound("Document not found".to_string())),
    }
}

async fn list_document_sets(
    State(state): State<AppState>,
) -> NexusResult<impl IntoResponse> {
    let sets = state.document_sets.read().await;
    Ok(Json(serde_json::json!({"data": *sets})))
}

async fn create_document_set(
    State(state): State<AppState>,
    Json(req): Json<CreateDocSetRequest>,
) -> NexusResult<impl IntoResponse> {
    let set = DocumentSet {
        id: uuid::Uuid::new_v4().to_string(),
        name: req.name,
        description: req.description,
        tenant_id: req.tenant_id,
        document_ids: req.document_ids,
        created_at: chrono::Utc::now().to_rfc3339(),
    };

    state.document_sets.write().await.push(set.clone());
    Ok(Json(serde_json::json!({"data": set})))
}

async fn get_stats(
    State(state): State<AppState>,
) -> NexusResult<impl IntoResponse> {
    let stats = state.stats.read().await.clone();
    Ok(Json(serde_json::json!({"data": stats})))
}
