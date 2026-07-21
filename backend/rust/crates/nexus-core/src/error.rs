use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use serde_json::json;

#[derive(Debug, thiserror::Error)]
pub enum NexusError {
    #[error("Unauthorized: {0}")]
    Unauthorized(String),

    #[error("Forbidden: {0}")]
    Forbidden(String),

    #[error("Not found: {0}")]
    NotFound(String),

    #[error("Bad request: {0}")]
    BadRequest(String),

    #[error("Conflict: {0}")]
    Conflict(String),

    #[error("Internal error: {0}")]
    Internal(String),

    #[error("Rate limit exceeded, retry after {retry_after}s")]
    RateLimit { retry_after: u64 },

    #[error("Validation error: {0}")]
    Validation(String),

    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),

    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),

    #[error("JWT error: {0}")]
    Jwt(#[from] jsonwebtoken::errors::Error),
}

impl IntoResponse for NexusError {
    fn into_response(self) -> Response {
        let (status, code, message) = match &self {
            NexusError::Unauthorized(msg) => (StatusCode::UNAUTHORIZED, "UNAUTHORIZED", msg.clone()),
            NexusError::Forbidden(msg) => (StatusCode::FORBIDDEN, "FORBIDDEN", msg.clone()),
            NexusError::NotFound(msg) => (StatusCode::NOT_FOUND, "NOT_FOUND", msg.clone()),
            NexusError::BadRequest(msg) => (StatusCode::BAD_REQUEST, "BAD_REQUEST", msg.clone()),
            NexusError::Conflict(msg) => (StatusCode::CONFLICT, "CONFLICT", msg.clone()),
            NexusError::Internal(msg) => (StatusCode::INTERNAL_SERVER_ERROR, "INTERNAL_ERROR", msg.clone()),
            NexusError::RateLimit { retry_after } => (
                StatusCode::TOO_MANY_REQUESTS,
                "RATE_LIMIT_EXCEEDED",
                format!("Retry after {}s", retry_after),
            ),
            NexusError::Validation(msg) => (StatusCode::BAD_REQUEST, "VALIDATION_ERROR", msg.clone()),
            NexusError::Http(e) => (StatusCode::BAD_GATEWAY, "BAD_GATEWAY", e.to_string()),
            NexusError::Json(e) => (StatusCode::BAD_REQUEST, "INVALID_JSON", e.to_string()),
            NexusError::Jwt(e) => (StatusCode::UNAUTHORIZED, "INVALID_TOKEN", e.to_string()),
        };

        let body = json!({
            "error": {
                "code": code,
                "message": message,
                "timestamp": chrono::Utc::now().to_rfc3339(),
            }
        });

        (status, axum::Json(body)).into_response()
    }
}

pub type NexusResult<T> = Result<T, NexusError>;
