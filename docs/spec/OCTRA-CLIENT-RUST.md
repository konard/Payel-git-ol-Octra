# OCTRA Rust Client

A minimal client for the Octra HTTP/JSON API using
[`reqwest`](https://crates.io/crates/reqwest) with the `tokio` async runtime.

```toml
# Cargo.toml
[dependencies]
reqwest = { version = "0.12", features = ["json"] }
serde = { version = "1", features = ["derive"] }
serde_json = "1"
tokio = { version = "1", features = ["full"] }
```

The full flow is: **register** to get an `api_key`, **create an environment**
(an AI CLI plus skills), then **chat**.

---

## 1. Reusable Client

```rust
use serde::{Deserialize, Serialize};
use serde_json::json;

#[derive(Serialize, Default)]
pub struct LlmConfig {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub provider: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub api_key: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub base_url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub model: Option<String>,
}

#[derive(Deserialize)]
pub struct Account {
    pub user_id: String,
    pub api_key: String,
    pub balance: i32,
}

pub struct OctraClient {
    base_url: String,
    api_key: Option<String>,
    http: reqwest::Client,
}

impl OctraClient {
    pub fn new(base_url: &str) -> Self {
        Self {
            base_url: base_url.trim_end_matches('/').to_string(),
            api_key: None,
            http: reqwest::Client::new(),
        }
    }

    async fn post(&self, path: &str, body: serde_json::Value) -> reqwest::Result<reqwest::Response> {
        let mut req = self
            .http
            .post(format!("{}{}", self.base_url, path))
            .json(&body);
        if let Some(key) = &self.api_key {
            req = req.header("octra-api-token", key);
        }
        req.send().await?.error_for_status()
    }

    pub async fn register(&mut self, username: &str, email: &str, password: &str) -> reqwest::Result<Account> {
        let account: Account = self
            .post("/register", json!({ "username": username, "email": email, "password": password }))
            .await?
            .json()
            .await?;
        self.api_key = Some(account.api_key.clone()); // remember it for later calls
        Ok(account)
    }

    pub async fn create_environment(
        &self,
        llm: LlmConfig,
        cli: &str,
        skills: &[&str],
    ) -> reqwest::Result<()> {
        self.post(
            "/environment",
            json!({ "llm": llm, "agent": { "cli": cli }, "skills": skills }),
        )
        .await?;
        Ok(())
    }

    pub async fn chat(&self, prompt: &str, skills: &[&str]) -> reqwest::Result<String> {
        #[derive(Deserialize)]
        struct ChatResponse {
            response: String,
        }
        let body: ChatResponse = self
            .post("/api/chat", json!({ "prompt": prompt, "skills": skills }))
            .await?
            .json()
            .await?;
        Ok(body.response)
    }
}
```

---

## 2. Full Flow Example

```rust
#[tokio::main]
async fn main() -> reqwest::Result<()> {
    let mut client = OctraClient::new("http://localhost:8080");

    // 1. Register and capture the API key.
    let account = client.register("me", "me@example.com", "secret").await?;
    println!("user_id: {}", account.user_id);

    // 2. Create an environment: an AI CLI plus some skills.
    client
        .create_environment(
            LlmConfig {
                provider: Some("claude".into()),
                api_key: Some("sk-...".into()),
                base_url: Some("https://api.anthropic.com".into()),
                model: Some("claude-sonnet-4-6".into()),
            },
            "claude-code",
            &["filesystem", "github", "brave-search"],
        )
        .await?;

    // 3. Send a prompt.
    let answer = client.chat("write a csv parser", &["filesystem"]).await?;
    println!("{}", answer);

    Ok(())
}
```
