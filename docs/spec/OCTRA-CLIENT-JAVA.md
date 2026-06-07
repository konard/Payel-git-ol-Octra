# OCTRA Java Client

## 1. Plain Java (with java.net.http)

```java
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.WebSocket;
import java.util.concurrent.CompletableFuture;

public class OctraClient {
    private String taskId;
    private int backoff = 1000;

    public void connect() {
        String uri = "wss://your-domain.com/ws";

        HttpClient.newHttpClient().newWebSocketBuilder()
            .buildAsync(URI.create(uri), new WebSocket.Listener() {
                @Override
                public void onOpen(WebSocket ws) {
                    if (taskId != null) {
                        ws.sendText("{\"type\":\"resume\",\"taskId\":\"" + taskId + "\"}", true);
                    } else {
                        ws.sendText("{\"username\":\"JavaClient\",\"user_id\":\"00000000-0000-0000-0000-000000000004\",\"title\":\"Java Task\",\"description\":\"Test from Java\",\"meta\":{\"model\":\"your-model\",\"provider\":\"provider\",\"publish_repositories\":\"true\",\"create_pull_requests\":\"true\"},\"tokens\":{\"provider\":\"your-api-key\"}}", true);
                    }
                }

                @Override
                public CompletionStage<?> onText(WebSocket ws, CharSequence data, boolean last) {
                    System.out.println("Update: " + data);
                    return null;
                }

                @Override
                public void onError(WebSocket ws, Throwable error) {
                    try { Thread.sleep(backoff); } catch (InterruptedException ignored) {}
                    backoff = Math.min(backoff * 2, 30000);
                    connect();
                }
            });
    }
}
```

---

## 2. Spring Boot WebSocket Client

```java
package com.octra.client;

import org.springframework.stereotype.Component;
import org.springframework.web.socket.*;
import org.springframework.web.socket.client.WebSocketConnectionManager;
import org.springframework.web.socket.client.standard.StandardWebSocketClient;

@Component
public class OctraSpringClient implements WebSocketHandler {

    private String taskId;
    private WebSocketSession session;
    private int backoff = 1000;

    public void connect() {
        String url = "wss://your-domain.com/ws";
        WebSocketConnectionManager manager = new WebSocketConnectionManager(
                new StandardWebSocketClient(), this, url);
        manager.start();
    }

    @Override
    public void afterConnectionEstablished(WebSocketSession session) throws Exception {
        this.session = session;
        if (taskId != null) {
            session.sendMessage(new TextMessage("{\"type\":\"resume\",\"taskId\":\"" + taskId + "\"}"));
        } else {
            session.sendMessage(new TextMessage("{\"type\":\"create\",\"title\":\"Spring Task\"}"));
        }
    }

    @Override
    public void handleMessage(WebSocketSession session, WebSocketMessage<?> message) throws Exception {
        System.out.println("Update: " + message.getPayload());
    }

    @Override
    public void handleTransportError(WebSocketSession session, Throwable exception) throws Exception {
        reconnect();
    }

    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus closeStatus) throws Exception {
        reconnect();
    }

    private void reconnect() {
        try {
            Thread.sleep(backoff);
            backoff = Math.min(backoff * 2, 30000);
            connect();
        } catch (InterruptedException ignored) {}
    }

    @Override
    public boolean supportsPartialMessages() {
        return false;
    }
}
```