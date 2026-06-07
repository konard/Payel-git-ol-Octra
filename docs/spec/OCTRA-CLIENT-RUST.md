# OCTRA Rust Client

## Reconnection with Resume

```rust
use futures_util::{SinkExt, StreamExt};
use tokio_tungstenite::tungstenite::protocol::Message;
use tokio::time::{sleep, Duration};
use serde_json::json;

#[tokio::main]
async fn main() {
    let mut task_id: Option<String> = None;
    let mut backoff = 1;
    let uri = "wss://your-domain.com/ws";

    loop {
        match tokio_tungstenite::connect_async(uri).await {
            Ok((mut ws, _)) => {
                if let Some(id) = &task_id {
                    let msg = json!({"type": "resume", "taskId": id});
                    ws.send(Message::Text(msg.to_string())).await.unwrap();
                } else {
                    let msg = json!({
                        "username": "RustClient",
                        "user_id": "00000000-0000-0000-0000-000000000008",
                        "title": "Rust Client",
                        "description": "Test from Rust",
                        "meta": {"model": "your-model", "provider": "provider", "publish_repositories": "true", "create_pull_requests": "true"},
                        "tokens": {"provider": "your-api-key"}
                    });
                    ws.send(Message::Text(msg.to_string())).await.unwrap();
                }

                while let Some(Ok(msg)) = ws.next().await {
                    println!("Update: {}", msg);
                }
            }
            Err(_) => {
                sleep(Duration::from_secs(backoff)).await;
                backoff = std::cmp::min(backoff * 2, 30);
            }
        }
    }
}
```