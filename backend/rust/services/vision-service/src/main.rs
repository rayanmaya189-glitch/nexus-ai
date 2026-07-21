use axum::{
    extract::State,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use base64::Engine;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::sync::atomic::{AtomicU64, Ordering};
use tower_http::cors::{Any, CorsLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use nexus_core::config::AppConfig;
use nexus_core::error::{NexusError, NexusResult};

#[derive(Clone)]
struct AppState {
    config: Arc<AppConfig>,
    http_client: reqwest::Client,
    request_count: Arc<AtomicU64>,
    total_latency_ms: Arc<AtomicU64>,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct ImageInput {
    image_url: Option<String>,
    base64_data: Option<String>,
    content_type: Option<String>,
}

#[derive(Debug, Deserialize)]
struct VisionRequest {
    image: ImageInput,
    prompt: Option<String>,
    model: Option<String>,
}

#[derive(Debug, Deserialize)]
struct MultimodalRequest {
    image: ImageInput,
    prompt: String,
    model: Option<String>,
}

#[derive(Debug, Serialize)]
struct VisionAnalysis {
    model_used: String,
    result: String,
    latency_ms: f64,
}

#[derive(Debug, Serialize)]
struct OCRResult {
    extracted_text: String,
    language: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
#[allow(dead_code)]
struct Category {
    name: String,
    confidence: f64,
}

#[derive(Debug, Serialize)]
struct ClassificationResult {
    categories: Vec<Category>,
}

#[derive(Debug, Serialize)]
struct VisionModel {
    name: String,
    model_id: String,
    capabilities: Vec<String>,
    status: String,
}

fn vision_models() -> Vec<VisionModel> {
    vec![
        VisionModel { name: "Qwen3 VL".into(), model_id: "qwen3-vl:4b".into(), capabilities: vec!["ocr".into(), "describe".into(), "classify".into(), "multimodal".into()], status: "active".into() },
        VisionModel { name: "LLaVA".into(), model_id: "llava:7b".into(), capabilities: vec!["ocr".into(), "describe".into(), "classify".into()], status: "active".into() },
        VisionModel { name: "LLaVA-Phi3".into(), model_id: "llava-phi3:3.6b".into(), capabilities: vec!["describe".into(), "multimodal".into()], status: "active".into() },
        VisionModel { name: "Moondream2".into(), model_id: "moondream2:1.8b".into(), capabilities: vec!["describe".into(), "ocr".into()], status: "active".into() },
    ]
}

async fn resolve_image_base64(image: &ImageInput, http_client: &reqwest::Client) -> NexusResult<String> {
    if let Some(ref b64) = image.base64_data {
        return Ok(b64.clone());
    }

    if let Some(ref url) = image.image_url {
        let resp = http_client.get(url).send().await?;
        let bytes = resp.bytes().await?;
        return Ok(base64::engine::general_purpose::STANDARD.encode(&bytes));
    }

    Err(NexusError::BadRequest("No image data provided (use base64_data or image_url)".to_string()))
}

async fn call_ollama_vision(
    http_client: &reqwest::Client,
    base_url: &str,
    model: &str,
    prompt: &str,
    image_base64: &str,
) -> NexusResult<String> {
    let resp = http_client
        .post(format!("{}/api/generate", base_url))
        .json(&serde_json::json!({
            "model": model,
            "prompt": prompt,
            "images": [image_base64],
            "stream": false,
        }))
        .send()
        .await?;

    let body: serde_json::Value = resp.json().await?;
    Ok(body
        .get("response")
        .and_then(|v| v.as_str())
        .unwrap_or("No response")
        .to_string())
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::try_from_default_env()
            .unwrap_or_else(|_| "vision-service=debug,tower_http=debug".into()))
        .with(tracing_subscriber::fmt::layer())
        .init();

    let config = AppConfig::from_file(
        &std::env::var("CONFIG_PATH").unwrap_or_else(|_| "configs/vision-service.json".to_string()),
    )?;

    let http_client = reqwest::Client::builder()
        .timeout(config.ollama.timeout)
        .build()?;

    let state = AppState {
        config: Arc::new(config.clone()),
        http_client,
        request_count: Arc::new(AtomicU64::new(0)),
        total_latency_ms: Arc::new(AtomicU64::new(0)),
    };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/api/v1/vision/analyze", post(analyze_image))
        .route("/api/v1/vision/ocr", post(ocr_extract))
        .route("/api/v1/vision/classify", post(classify_image))
        .route("/api/v1/vision/describe", post(describe_image))
        .route("/api/v1/vision/multimodal", post(multimodal_query))
        .route("/api/v1/vision/models", get(list_models))
        .route("/api/v1/vision/stats", get(get_stats))
        .layer(cors)
        .with_state(state);

    let addr = format!("0.0.0.0:{}", config.server.http_port);
    tracing::info!("Vision Service starting on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({
        "status": "healthy",
        "service": "vision-service",
        "version": "1.0.0"
    }))
}

async fn analyze_image(
    State(state): State<AppState>,
    Json(req): Json<VisionRequest>,
) -> NexusResult<impl IntoResponse> {
    let start = std::time::Instant::now();
    let model = req.model.unwrap_or_else(|| "qwen3-vl:4b".to_string());
    let prompt = req.prompt.unwrap_or_else(|| "Analyze this image in detail. Describe what you see.".to_string());

    let image_b64 = resolve_image_base64(&req.image, &state.http_client).await?;
    let result = call_ollama_vision(&state.http_client, &state.config.ollama.base_url, &model, &prompt, &image_b64).await?;

    let latency = start.elapsed().as_millis() as f64;
    state.request_count.fetch_add(1, Ordering::Relaxed);
    state.total_latency_ms.fetch_add(latency as u64, Ordering::Relaxed);

    Ok(Json(serde_json::json!({
        "data": VisionAnalysis {
            model_used: model,
            result,
            latency_ms: latency,
        }
    })))
}

async fn ocr_extract(
    State(state): State<AppState>,
    Json(req): Json<VisionRequest>,
) -> NexusResult<impl IntoResponse> {
    let start = std::time::Instant::now();
    let model = req.model.unwrap_or_else(|| "qwen3-vl:4b".to_string());
    let image_b64 = resolve_image_base64(&req.image, &state.http_client).await?;

    let result = call_ollama_vision(
        &state.http_client,
        &state.config.ollama.base_url,
        &model,
        "Extract all text from this image. Return only the extracted text, nothing else.",
        &image_b64,
    )
    .await?;

    let latency = start.elapsed().as_millis() as f64;
    state.request_count.fetch_add(1, Ordering::Relaxed);
    state.total_latency_ms.fetch_add(latency as u64, Ordering::Relaxed);

    Ok(Json(serde_json::json!({
        "data": OCRResult {
            extracted_text: result,
            language: Some("auto".to_string()),
        }
    })))
}

async fn classify_image(
    State(state): State<AppState>,
    Json(req): Json<VisionRequest>,
) -> NexusResult<impl IntoResponse> {
    let start = std::time::Instant::now();
    let model = req.model.unwrap_or_else(|| "qwen3-vl:4b".to_string());
    let image_b64 = resolve_image_base64(&req.image, &state.http_client).await?;

    let result = call_ollama_vision(
        &state.http_client,
        &state.config.ollama.base_url,
        &model,
        "Classify this image into categories. Return a JSON array with objects having 'name' and 'confidence' fields. Return only valid JSON, no other text.",
        &image_b64,
    )
    .await?;

    let categories: Vec<Category> = serde_json::from_str(&result).unwrap_or_else(|_| {
        vec![Category {
            name: "unknown".to_string(),
            confidence: 0.5,
        }]
    });

    let latency = start.elapsed().as_millis() as f64;
    state.request_count.fetch_add(1, Ordering::Relaxed);
    state.total_latency_ms.fetch_add(latency as u64, Ordering::Relaxed);

    Ok(Json(serde_json::json!({
        "data": ClassificationResult { categories }
    })))
}

async fn describe_image(
    State(state): State<AppState>,
    Json(req): Json<VisionRequest>,
) -> NexusResult<impl IntoResponse> {
    let start = std::time::Instant::now();
    let model = req.model.unwrap_or_else(|| "llava:7b".to_string());
    let image_b64 = resolve_image_base64(&req.image, &state.http_client).await?;

    let result = call_ollama_vision(
        &state.http_client,
        &state.config.ollama.base_url,
        &model,
        "Describe this image in detail. Cover the main subjects, setting, colors, and any notable features.",
        &image_b64,
    )
    .await?;

    let latency = start.elapsed().as_millis() as f64;
    state.request_count.fetch_add(1, Ordering::Relaxed);
    state.total_latency_ms.fetch_add(latency as u64, Ordering::Relaxed);

    Ok(Json(serde_json::json!({
        "data": VisionAnalysis {
            model_used: model,
            result,
            latency_ms: latency,
        }
    })))
}

async fn multimodal_query(
    State(state): State<AppState>,
    Json(req): Json<MultimodalRequest>,
) -> NexusResult<impl IntoResponse> {
    let start = std::time::Instant::now();
    let model = req.model.unwrap_or_else(|| "qwen3-vl:4b".to_string());
    let image_b64 = resolve_image_base64(&req.image, &state.http_client).await?;

    let result = call_ollama_vision(
        &state.http_client,
        &state.config.ollama.base_url,
        &model,
        &req.prompt,
        &image_b64,
    )
    .await?;

    let latency = start.elapsed().as_millis() as f64;
    state.request_count.fetch_add(1, Ordering::Relaxed);
    state.total_latency_ms.fetch_add(latency as u64, Ordering::Relaxed);

    Ok(Json(serde_json::json!({
        "data": VisionAnalysis {
            model_used: model,
            result,
            latency_ms: latency,
        }
    })))
}

async fn list_models() -> impl IntoResponse {
    Json(serde_json::json!({"data": vision_models()}))
}

async fn get_stats(
    State(state): State<AppState>,
) -> NexusResult<impl IntoResponse> {
    let count = state.request_count.load(Ordering::Relaxed);
    let total_ms = state.total_latency_ms.load(Ordering::Relaxed);
    let avg_latency = if count > 0 { total_ms as f64 / count as f64 } else { 0.0 };

    Ok(Json(serde_json::json!({
        "data": {
            "total_requests": count,
            "total_latency_ms": total_ms,
            "avg_latency_ms": avg_latency,
        }
    })))
}
