use axum::{
    extract::State,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;
use tower_http::cors::{Any, CorsLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use nexus_core::config::AppConfig;
use nexus_core::error::NexusResult;

#[derive(Clone)]
#[allow(dead_code)]
struct AppState {
    config: Arc<AppConfig>,
    filters: Arc<SensitiveWordFilters>,
    stats: Arc<RwLock<SecurityStats>>,
}

#[allow(dead_code)]
struct SensitiveWordFilters {
    profanity: Vec<String>,
    hate_speech: Vec<String>,
    violence: Vec<String>,
    self_harm: Vec<String>,
    sexual_content: Vec<String>,
    illegal_activity: Vec<String>,
    pii_email: Regex,
    pii_phone_us: Regex,
    pii_phone_intl: Regex,
    pii_ssn: Regex,
    pii_credit_card: Regex,
    credential_api_key: Regex,
    credential_aws_key: Regex,
    credential_private_key: Regex,
    injection_patterns: Vec<String>,
}

#[derive(Debug, Default, Clone, Serialize)]
#[allow(dead_code)]
struct SecurityStats {
    total_scans: u64,
    violations_found: u64,
    injections_detected: u64,
    pii_detected: u64,
    credentials_detected: u64,
}

#[derive(Debug, Deserialize)]
struct FilterRequest {
    text: String,
    mode: Option<String>,
}

#[derive(Debug, Serialize, Clone)]
#[allow(dead_code)]
struct Violation {
    category: String,
    matched_text: String,
    position_start: usize,
    position_end: usize,
    severity: String,
}

#[derive(Debug, Serialize)]
struct FilterResult {
    filtered_text: String,
    violations: Vec<Violation>,
    safe: bool,
    score: f64,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct ValidationRequest {
    text: String,
    content_type: String,
}

#[derive(Debug, Serialize)]
struct ValidationResult {
    safe: bool,
    score: f64,
    violations: Vec<Violation>,
    recommendation: String,
}

#[derive(Debug, Deserialize)]
struct InjectionCheck {
    prompt: String,
}

#[derive(Debug, Serialize)]
struct InjectionResult {
    safe: bool,
    attack_type: Option<String>,
    confidence: f64,
    details: String,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct ScanRequest {
    text: String,
    check_types: Option<Vec<String>>,
}

#[derive(Debug, Serialize)]
struct ScanResult {
    safe: bool,
    score: f64,
    violations: Vec<Violation>,
    injection_detected: bool,
    pii_detected: bool,
    credentials_detected: bool,
}

#[derive(Debug, Serialize)]
struct CategoryInfo {
    id: String,
    name: String,
    description: String,
    severity: String,
}

fn build_filters() -> SensitiveWordFilters {
    SensitiveWordFilters {
        profanity: vec![
            "damn", "hell", "crap", "ass", "asshole", "bastard", "bitch", "shit",
            "fuck", "fucking", "fucked", "dick", "cock", "piss", "bollocks",
            "bugger", "bloody", "slut", "whore", "dumbass", " jackass",
        ].iter().map(|s| s.to_string()).collect(),

        hate_speech: vec![
            "nigger", "chink", "spic", "wetback", "kike", "gook", "towelhead",
            "raghead", "beaner", "cracker", "redneck", "white trash", "negro",
            "retard", "retarded", "cripple", "fag", "faggot", "dyke", "tranny",
        ].iter().map(|s| s.to_string()).collect(),

        violence: vec![
            "kill", "murder", "assassinate", "execute", "slaughter", "massacre",
            "bomb", "detonate", "explode", "shoot", "stab", "strangle", "hang",
            "dismember", "behead", "torture", "maim", "assault", "attack",
            "terrorize", "threaten", "lynch",
        ].iter().map(|s| s.to_string()).collect(),

        self_harm: vec![
            "suicide", "kill myself", "end my life", "self harm", "cut myself",
            "overdose", "hang myself", "jump off", "slit wrists", "die",
            "want to die", "better off dead", "no reason to live",
            "self-injury", "self-mutilation",
        ].iter().map(|s| s.to_string()).collect(),

        sexual_content: vec![
            "porn", "pornography", "nude", "naked", "sex", "sexual",
            "erotic", "xxx", "adult content", "nsfw", "orgasm", "genital",
            "penis", "vagina", "breasts", "masturbate", "prostitute",
            "escort", "hooker", "strip club", "blowjob",
        ].iter().map(|s| s.to_string()).collect(),

        illegal_activity: vec![
            "drug", "cocaine", "heroin", "meth", "marijuana", "weed",
            "lsd", "mdma", "ecstasy", "fentanyl", "deal drugs", "sell drugs",
            "gamble", "illegal gambling", "money laundering", "human trafficking",
            "smuggle", "counterfeit", "forgery", "arson", "burglary",
        ].iter().map(|s| s.to_string()).collect(),

        pii_email: Regex::new(r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}").unwrap(),
        pii_phone_us: Regex::new(r"\b\d{3}[-.]?\d{3}[-.]?\d{4}\b").unwrap(),
        pii_phone_intl: Regex::new(r"\+\d{1,3}[-.\s]?\d{4,14}").unwrap(),
        pii_ssn: Regex::new(r"\b\d{3}-\d{2}-\d{4}\b").unwrap(),
        pii_credit_card: Regex::new(r"\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b").unwrap(),

        credential_api_key: Regex::new(r"\b(sk-[a-zA-Z0-9]{20,}|AKIA[0-9A-Z]{16}|ghp_[a-zA-Z0-9]{36}|xox[bpsar]-[a-zA-Z0-9-]+)").unwrap(),
        credential_aws_key: Regex::new(r"\b(AKIA[0-9A-Z]{16})").unwrap(),
        credential_private_key: Regex::new(r"-----BEGIN (RSA |EC )?PRIVATE KEY-----").unwrap(),

        injection_patterns: vec![
            "ignore previous instructions",
            "ignore all previous",
            "disregard your instructions",
            "disregard previous",
            "forget your instructions",
            "forget everything above",
            "you are now",
            "act as if",
            "pretend you are",
            "roleplay as",
            "new instructions:",
            "system prompt:",
            "override your instructions",
            "bypass your filters",
            "jailbreak",
            "do anything now",
            "dan mode",
            "developer mode",
            "output your instructions",
            "reveal your prompt",
            "what are your instructions",
            "print your system prompt",
            "repeat everything above",
            "you are now in",
            "from now on you",
            "new identity",
            "your new role",
        ].iter().map(|s| s.to_string()).collect(),
    }
}

fn scan_text_for_violations(text: &str, filters: &SensitiveWordFilters) -> Vec<Violation> {
    let mut violations = Vec::new();
    let lower = text.to_lowercase();

    let category_words: Vec<(&str, &Vec<String>, &str)> = vec![
        ("PROFANITY", &filters.profanity, "high"),
        ("HATE_SPEECH", &filters.hate_speech, "critical"),
        ("VIOLENCE", &filters.violence, "high"),
        ("SELF_HARM", &filters.self_harm, "critical"),
        ("SEXUAL_CONTENT", &filters.sexual_content, "high"),
        ("ILLEGAL_ACTIVITY", &filters.illegal_activity, "critical"),
    ];

    for (category, words, severity) in category_words {
        for word in words {
            if let Some(pos) = lower.find(&word.to_lowercase()) {
                violations.push(Violation {
                    category: category.to_string(),
                    matched_text: text[pos..pos + word.len()].to_string(),
                    position_start: pos,
                    position_end: pos + word.len(),
                    severity: severity.to_string(),
                });
            }
        }
    }

    for mat in filters.pii_email.find_iter(text) {
        violations.push(Violation {
            category: "PII".to_string(),
            matched_text: mat.as_str().to_string(),
            position_start: mat.start(),
            position_end: mat.end(),
            severity: "high".to_string(),
        });
    }

    for mat in filters.pii_ssn.find_iter(text) {
        violations.push(Violation {
            category: "PII".to_string(),
            matched_text: "***-**-****".to_string(),
            position_start: mat.start(),
            position_end: mat.end(),
            severity: "critical".to_string(),
        });
    }

    for mat in filters.pii_credit_card.find_iter(text) {
        violations.push(Violation {
            category: "PII".to_string(),
            matched_text: "****-****-****-****".to_string(),
            position_start: mat.start(),
            position_end: mat.end(),
            severity: "critical".to_string(),
        });
    }

    for mat in filters.credential_api_key.find_iter(text) {
        violations.push(Violation {
            category: "CREDENTIALS".to_string(),
            matched_text: "***REDACTED***".to_string(),
            position_start: mat.start(),
            position_end: mat.end(),
            severity: "critical".to_string(),
        });
    }

    if filters.credential_private_key.is_match(text) {
        violations.push(Violation {
            category: "CREDENTIALS".to_string(),
            matched_text: "PRIVATE_KEY".to_string(),
            position_start: 0,
            position_end: text.len(),
            severity: "critical".to_string(),
        });
    }

    violations
}

fn detect_injection(prompt: &str, filters: &SensitiveWordFilters) -> InjectionResult {
    let lower = prompt.to_lowercase();

    for pattern in &filters.injection_patterns {
        if lower.contains(&pattern.to_lowercase()) {
            return InjectionResult {
                safe: false,
                attack_type: Some("prompt_injection".to_string()),
                confidence: 0.9,
                details: format!("Detected injection pattern: '{}'", pattern),
            };
        }
    }

    if lower.contains("act as") && (lower.contains("different") || lower.contains("new") || lower.contains("evil")) {
        return InjectionResult {
            safe: false,
            attack_type: Some("role_hijacking".to_string()),
            confidence: 0.85,
            details: "Possible role-hijacking attempt detected".to_string(),
        };
    }

    InjectionResult {
        safe: true,
        attack_type: None,
        confidence: 0.95,
        details: "No injection patterns detected".to_string(),
    }
}

fn sanitize_text(text: &str, violations: &[Violation]) -> String {
    let mut result = text.to_string();
    let mut sorted_violations = violations.to_vec();
    sorted_violations.sort_by(|a, b| b.position_start.cmp(&a.position_start));

    for v in &sorted_violations {
        let replacement = match v.category.as_str() {
            "PII" => "[REDACTED_PII]",
            "CREDENTIALS" => "[REDACTED_CREDENTIALS]",
            _ => "[REDACTED]",
        };
        result.replace_range(v.position_start..v.position_end, replacement);
    }
    result
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::try_from_default_env()
            .unwrap_or_else(|_| "security-ai=debug,tower_http=debug".into()))
        .with(tracing_subscriber::fmt::layer())
        .init();

    let config = AppConfig::from_file(
        &std::env::var("CONFIG_PATH").unwrap_or_else(|_| "configs/security-ai.json".to_string()),
    )?;

    let state = AppState {
        config: Arc::new(config.clone()),
        filters: Arc::new(build_filters()),
        stats: Arc::new(RwLock::new(SecurityStats::default())),
    };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/api/v1/security/filter", post(filter_text))
        .route("/api/v1/security/validate-prompt", post(validate_prompt))
        .route("/api/v1/security/validate-response", post(validate_response))
        .route("/api/v1/security/injection-check", post(check_injection))
        .route("/api/v1/security/scan", post(comprehensive_scan))
        .route("/api/v1/security/categories", get(list_categories))
        .route("/api/v1/security/stats", get(get_stats))
        .layer(cors)
        .with_state(state);

    let addr = format!("0.0.0.0:{}", config.server.http_port);
    tracing::info!("Security AI Service starting on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({
        "status": "healthy",
        "service": "security-ai",
        "version": "1.0.0"
    }))
}

async fn filter_text(
    State(state): State<AppState>,
    Json(req): Json<FilterRequest>,
) -> NexusResult<impl IntoResponse> {
    let violations = scan_text_for_violations(&req.text, &state.filters);
    let mode = req.mode.unwrap_or_else(|| "block".to_string());

    let filtered_text = if mode == "sanitize" && !violations.is_empty() {
        sanitize_text(&req.text, &violations)
    } else {
        req.text.clone()
    };

    let safe = violations.is_empty();
    let score = if violations.is_empty() {
        1.0
    } else {
        (1.0 - violations.len() as f64 * 0.1).max(0.0)
    };

    let mut stats = state.stats.write().await;
    stats.total_scans += 1;
    if !safe {
        stats.violations_found += 1;
    }

    Ok(Json(serde_json::json!({
        "data": FilterResult {
            filtered_text,
            violations,
            safe,
            score,
        }
    })))
}

async fn validate_prompt(
    State(state): State<AppState>,
    Json(req): Json<ValidationRequest>,
) -> NexusResult<impl IntoResponse> {
    let violations = scan_text_for_violations(&req.text, &state.filters);
    let injection = detect_injection(&req.text, &state.filters);

    let safe = violations.is_empty() && injection.safe;
    let score = if safe {
        1.0
    } else {
        let v_score = if violations.is_empty() { 1.0 } else { (1.0 - violations.len() as f64 * 0.1).max(0.0) };
        let i_score = injection.confidence;
        v_score.min(i_score)
    };

    let recommendation = if safe {
        "Prompt is safe to process".to_string()
    } else if !injection.safe {
        format!("BLOCK: Potential prompt injection detected (type: {:?})", injection.attack_type)
    } else {
        format!("BLOCK: {} violation(s) found in prompt", violations.len())
    };

    let mut stats = state.stats.write().await;
    stats.total_scans += 1;
    if !injection.safe {
        stats.injections_detected += 1;
    }
    if violations.iter().any(|v| v.category == "PII") {
        stats.pii_detected += 1;
    }

    Ok(Json(serde_json::json!({
        "data": ValidationResult {
            safe,
            score,
            violations,
            recommendation,
        }
    })))
}

async fn validate_response(
    State(state): State<AppState>,
    Json(req): Json<ValidationRequest>,
) -> NexusResult<impl IntoResponse> {
    let violations = scan_text_for_violations(&req.text, &state.filters);
    let safe = violations.is_empty();
    let score = if safe { 1.0 } else { (1.0 - violations.len() as f64 * 0.1).max(0.0) };

    let recommendation = if safe {
        "Response is safe to return".to_string()
    } else {
        format!("MODIFY: {} violation(s) found in response", violations.len())
    };

    let mut stats = state.stats.write().await;
    stats.total_scans += 1;
    if !safe {
        stats.violations_found += 1;
    }

    Ok(Json(serde_json::json!({
        "data": ValidationResult {
            safe,
            score,
            violations,
            recommendation,
        }
    })))
}

async fn check_injection(
    State(state): State<AppState>,
    Json(req): Json<InjectionCheck>,
) -> NexusResult<impl IntoResponse> {
    let result = detect_injection(&req.prompt, &state.filters);

    let mut stats = state.stats.write().await;
    stats.total_scans += 1;
    if !result.safe {
        stats.injections_detected += 1;
    }

    Ok(Json(serde_json::json!({"data": result})))
}

async fn comprehensive_scan(
    State(state): State<AppState>,
    Json(req): Json<ScanRequest>,
) -> NexusResult<impl IntoResponse> {
    let violations = scan_text_for_violations(&req.text, &state.filters);
    let injection = detect_injection(&req.text, &state.filters);
    let pii_detected = violations.iter().any(|v| v.category == "PII");
    let credentials_detected = violations.iter().any(|v| v.category == "CREDENTIALS");

    let safe = violations.is_empty() && injection.safe;
    let score = if safe { 1.0 } else { (1.0 - violations.len() as f64 * 0.1).max(0.0) };

    let mut stats = state.stats.write().await;
    stats.total_scans += 1;
    if !safe { stats.violations_found += 1; }
    if !injection.safe { stats.injections_detected += 1; }
    if pii_detected { stats.pii_detected += 1; }
    if credentials_detected { stats.credentials_detected += 1; }

    Ok(Json(serde_json::json!({
        "data": ScanResult {
            safe,
            score,
            violations,
            injection_detected: !injection.safe,
            pii_detected,
            credentials_detected,
        }
    })))
}

async fn list_categories() -> impl IntoResponse {
    let categories = vec![
        CategoryInfo { id: "PROFANITY".into(), name: "Profanity".into(), description: "Profanity and obscenity".into(), severity: "high".into() },
        CategoryInfo { id: "HATE_SPEECH".into(), name: "Hate Speech".into(), description: "Hate speech and discrimination".into(), severity: "critical".into() },
        CategoryInfo { id: "VIOLENCE".into(), name: "Violence".into(), description: "Violent content".into(), severity: "high".into() },
        CategoryInfo { id: "SELF_HARM".into(), name: "Self Harm".into(), description: "Self-harm and suicide".into(), severity: "critical".into() },
        CategoryInfo { id: "SEXUAL_CONTENT".into(), name: "Sexual Content".into(), description: "Sexual or explicit content".into(), severity: "high".into() },
        CategoryInfo { id: "ILLEGAL_ACTIVITY".into(), name: "Illegal Activity".into(), description: "Illegal activity promotion".into(), severity: "critical".into() },
        CategoryInfo { id: "PII".into(), name: "PII".into(), description: "Personally identifiable information".into(), severity: "high".into() },
        CategoryInfo { id: "CREDENTIALS".into(), name: "Credentials".into(), description: "Exposed credentials or secrets".into(), severity: "critical".into() },
        CategoryInfo { id: "INTERNAL_INFO".into(), name: "Internal Info".into(), description: "Internal sensitive information".into(), severity: "high".into() },
        CategoryInfo { id: "COMPETITOR".into(), name: "Competitor".into(), description: "Competitor information".into(), severity: "medium".into() },
    ];

    Json(serde_json::json!({"data": categories}))
}

async fn get_stats(
    State(state): State<AppState>,
) -> NexusResult<impl IntoResponse> {
    let stats = state.stats.read().await.clone();
    Ok(Json(serde_json::json!({"data": stats})))
}
