# OCTRA TypeScript / Node.js Client

## Basic Connection

```ts
import WebSocket from 'ws';

const ws = new WebSocket('wss://your-domain.com/ws');

ws.on('open', () => {
  ws.send(JSON.stringify({
    type: 'create',
    title: 'TS Client Task'
  }));
});

ws.on('message', (data) => {
  console.log('Update:', JSON.parse(data.toString()));
});
```

## Reconnection + Resume

```ts
let taskId: string | null = null;
let backoff = 1000;

function connect() {
  const ws = new WebSocket('wss://your-domain.com/ws');

  ws.on('open', () => {
    if (taskId) {
      ws.send(JSON.stringify({ type: 'resume', taskId }));
    } else {
      ws.send(JSON.stringify({
        username: "TSClient",
        user_id: "00000000-0000-0000-0000-000000000002",
        title: "TS Task",
        description: "Test from TypeScript",
        meta: { model: "your-model", provider: "provider", publish_repositories: "true", create_pull_requests: "true" },
        tokens: { provider: "your-api-key" }
      }));
    }
  });

  ws.on('message', (data) => {
    const msg = JSON.parse(data.toString());
    if (msg.taskId) taskId = msg.taskId;
    console.log('Update:', msg);
  });

  ws.on('close', () => {
    setTimeout(() => {
      backoff = Math.min(backoff * 2, 30000);
      connect();
    }, backoff);
  });
}

connect();
```