use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};
use serde::{Deserialize, Serialize};
use std::time::{Duration, SystemTime, UNIX_EPOCH};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct JWTClaims {
    pub sub: String,
    pub tenant_id: i64,
    pub email: String,
    pub roles: Vec<String>,
    pub permissions: Vec<String>,
    pub exp: u64,
    pub iat: u64,
    pub iss: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct TokenPair {
    pub access_token: String,
    pub refresh_token: String,
    pub expires_in: i64,
}

pub struct JWTManager {
    encoding_key: EncodingKey,
    decoding_key: DecodingKey,
    issuer: String,
    access_ttl: Duration,
    refresh_ttl: Duration,
}

impl JWTManager {
    pub fn new(secret: &str, issuer: &str, access_ttl: Duration, refresh_ttl: Duration) -> Self {
        Self {
            encoding_key: EncodingKey::from_secret(secret.as_bytes()),
            decoding_key: DecodingKey::from_secret(secret.as_bytes()),
            issuer: issuer.to_string(),
            access_ttl,
            refresh_ttl,
        }
    }

    pub fn generate_token_pair(
        &self,
        user_id: i64,
        tenant_id: i64,
        email: &str,
        roles: Vec<String>,
        permissions: Vec<String>,
    ) -> anyhow::Result<TokenPair> {
        let now = SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs();

        let access_claims = JWTClaims {
            sub: user_id.to_string(),
            tenant_id,
            email: email.to_string(),
            roles,
            permissions,
            exp: now + self.access_ttl.as_secs(),
            iat: now,
            iss: self.issuer.clone(),
        };

        let access_token = encode(&Header::default(), &access_claims, &self.encoding_key)?;

        let refresh_claims = JWTClaims {
            sub: email.to_string(),
            tenant_id,
            email: email.to_string(),
            roles: vec![],
            permissions: vec![],
            exp: now + self.refresh_ttl.as_secs(),
            iat: now,
            iss: self.issuer.clone(),
        };

        let refresh_token = encode(&Header::default(), &refresh_claims, &self.encoding_key)?;

        Ok(TokenPair {
            access_token,
            refresh_token,
            expires_in: self.access_ttl.as_secs() as i64,
        })
    }

    pub fn validate_access_token(&self, token: &str) -> anyhow::Result<JWTClaims> {
        let token_data = decode::<JWTClaims>(token, &self.decoding_key, &Validation::default())?;
        let now = SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs();
        if token_data.claims.exp < now {
            anyhow::bail!("Token expired");
        }
        Ok(token_data.claims)
    }
}
