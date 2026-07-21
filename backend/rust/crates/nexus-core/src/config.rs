use serde::Deserialize;
use std::time::Duration;

#[derive(Debug, Deserialize, Clone)]
pub struct AppConfig {
    pub server: ServerConfig,
    pub ollama: OllamaConfig,
}

#[derive(Debug, Deserialize, Clone)]
pub struct ServerConfig {
    #[serde(default = "default_port")]
    pub http_port: u16,
    #[serde(default)]
    pub debug: bool,
}

fn default_port() -> u16 {
    8080
}

#[derive(Debug, Deserialize, Clone)]
pub struct OllamaConfig {
    #[serde(default = "default_ollama_url")]
    pub base_url: String,
    #[serde(default = "default_timeout")]
    pub timeout: Duration,
}

fn default_ollama_url() -> String {
    "http://localhost:11434".to_string()
}

fn default_timeout() -> Duration {
    Duration::from_secs(120)
}

impl AppConfig {
    pub fn from_file(path: &str) -> anyhow::Result<Self> {
        let content = std::fs::read_to_string(path)?;
        let config: AppConfig = serde_json::from_str(&content)?;
        Ok(config)
    }
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            server: ServerConfig {
                http_port: 8080,
                debug: false,
            },
            ollama: OllamaConfig {
                base_url: default_ollama_url(),
                timeout: default_timeout(),
            },
        }
    }
}
