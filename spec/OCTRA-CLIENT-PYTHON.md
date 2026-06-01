# OCTRA Python Client

## 1. Async WebSocket Client (websockets)

```python
import asyncio
import websockets
import json

async def run_client():
    uri = "wss://your-domain.com/ws"
    task_id = None
    backoff = 1

    while True:
        try:
            async with websockets.connect(uri) as ws:
                if task_id:
                    await ws.send(json.dumps({"type": "resume", "taskId": task_id}))
                else:
                    await ws.send(json.dumps({
                        "username": "PythonClient",
                        "user_id": "00000000-0000-0000-0000-000000000001",
                        "title": "Python client task",
                        "description": "Test task from Python",
                        "meta": {"model": "your-model", "provider": "provider"},
                        "tokens": {"provider": "your-api-key"}
                    }))

                async for message in ws:
                    data = json.loads(message)
                    print("Update:", data)
                    if "taskId" in data:
                        task_id = data["taskId"]
        except Exception as e:
            print("Disconnected:", e)
            await asyncio.sleep(backoff)
            backoff = min(backoff * 2, 30)
```

---

## 2. FastAPI + WebSocket Client (Background Task)

```python
from fastapi import FastAPI, BackgroundTasks
import asyncio
import websockets
import json

app = FastAPI()
task_id = None

async def octra_listener():
    global task_id
    uri = "wss://your-domain.com/ws"
    backoff = 1

    while True:
        try:
            async with websockets.connect(uri) as ws:
                if task_id:
                    await ws.send(json.dumps({"type": "resume", "taskId": task_id}))
                else:
                    await ws.send(json.dumps({"type": "create", "title": "FastAPI Task"}))

                async for message in ws:
                    data = json.loads(message)
                    print("Received:", data)
                    if "taskId" in data:
                        task_id = data["taskId"]
        except Exception as e:
            print("Connection error:", e)
            await asyncio.sleep(backoff)
            backoff = min(backoff * 2, 30)

@app.on_event("startup")
async def startup_event():
    asyncio.create_task(octra_listener())
```