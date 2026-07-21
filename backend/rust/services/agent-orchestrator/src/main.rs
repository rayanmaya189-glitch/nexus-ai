use axum::{
    extract::{Path, State},
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tower_http::cors::{Any, CorsLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use nexus_core::config::AppConfig;
use nexus_core::error::{NexusError, NexusResult};

#[derive(Clone)]
struct AppState {
    config: Arc<AppConfig>,
    http_client: reqwest::Client,
}

#[derive(Debug, Deserialize)]
struct ExecuteAgentRequest {
    agent: String,
    task: String,
}

#[derive(Debug, Serialize)]
struct ExecuteAgentResponse {
    execution_id: String,
    status: String,
    model_used: String,
    result: String,
}

#[derive(Debug, Serialize)]
struct Agent {
    id: i64,
    name: String,
    agent_type: String,
    model: String,
    status: String,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::try_from_default_env()
            .unwrap_or_else(|_| "agent-orchestrator=debug,tower_http=debug".into()))
        .with(tracing_subscriber::fmt::layer())
        .init();

    let config = AppConfig::default();
    let http_client = reqwest::Client::builder()
        .timeout(config.ollama.timeout)
        .build()?;

    let state = AppState {
        config: Arc::new(config),
        http_client,
    };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/api/v1/agents", get(list_agents))
        .route("/api/v1/agents/{id}", get(get_agent))
        .route("/api/v1/agents/execute", post(execute_agent))
        .layer(cors)
        .with_state(state);

    let port = std::env::var("PORT").unwrap_or_else(|_| "8091".to_string());
    let addr = format!("0.0.0.0:{}", port);
    tracing::info!("Agent Orchestrator starting on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({
        "status": "healthy",
        "service": "agent-orchestrator",
        "version": "1.0.0"
    }))
}

async fn list_agents() -> NexusResult<impl IntoResponse> {
    let agents = vec![
        Agent { id: 1, name: "planner".into(), agent_type: "planner".into(), model: "lfm2.5-thinking:1.2b".into(), status: "active".into() },
        Agent { id: 2, name: "customer-agent".into(), agent_type: "customer".into(), model: "command-r7b:7b".into(), status: "active".into() },
        Agent { id: 3, name: "developer-agent".into(), agent_type: "developer".into(), model: "qwen2.5-coder:3b".into(), status: "active".into() },
        Agent { id: 4, name: "vision-agent".into(), agent_type: "vision".into(), model: "qwen3-vl:4b".into(), status: "active".into() },
        Agent { id: 5, name: "security-agent".into(), agent_type: "security".into(), model: "whiterabbitneo:7b".into(), status: "active".into() },
        Agent { id: 6, name: "business-agent".into(), agent_type: "business".into(), model: "llama3.1:7b".into(), status: "active".into() },
    ];
    Ok(Json(serde_json::json!({"data": agents})))
}

async fn get_agent(Path(id): Path<i64>) -> NexusResult<impl IntoResponse> {
    let agent = Agent {
        id,
        name: "customer-agent".into(),
        agent_type: "customer".into(),
        model: "command-r7b:7b".into(),
        status: "active".into(),
    };
    Ok(Json(serde_json::json!({"data": agent})))
}

async fn execute_agent(
    State(state): State<AppState>,
    Json(req): Json<ExecuteAgentRequest>,
) -> NexusResult<impl IntoResponse> {
    let execution_id = uuid::Uuid::new_v4().to_string();

    let model = match req.agent.as_str() {
        "planner" => "lfm2.5-thinking:1.2b",
        "customer-agent" => "command-r7b:7b",
        "developer-agent" => "qwen2.5-coder:3b",
        "vision-agent" => "qwen3-vl:4b",
        "security-agent" => "whiterabbitneo:7b",
        "business-agent" => "llama3.1:7b",
        _ => "phi4-mini:3.8b",
    };

    let ollama_url = format!("{}/api/generate", state.config.ollama.base_url);
    let response = state.http_client
        .post(&ollama_url)
        .json(&serde_json::json!({
            "model": model,
            "prompt": req.task,
            "stream": false,
        }))
        .send()
        .await;

    let result = match response {
        Ok(resp) => {
            let body: serde_json::Value = resp.json().await.unwrap_or_default();
            body.get("response").and_then(|v| v.as_str()).unwrap_or("No response").to_string()
        }
        Err(e) => {
            return Err(NexusError::Internal(format!("AI model error: {}", e)));
        }
    };

    Ok(Json(serde_json::json!({
        "data": ExecuteAgentResponse {
            execution_id,
            status: "completed".into(),
            model_used: model.into(),
            result,
        }
    })))
}
