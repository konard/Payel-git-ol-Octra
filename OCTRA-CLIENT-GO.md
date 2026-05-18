# OCTRA Go Client

## Reconnection with Resume

```go
package main

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	var taskID string
	backoff := time.Second
	uri := "wss://your-domain.com/ws"

	for {
		conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
		if err != nil {
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * 1.5)
			continue
		}

		if taskID != "" {
			conn.WriteJSON(map[string]string{"type": "resume", "taskId": taskID})
		} else {
			conn.WriteJSON(map[string]string{
				"type":  "create",
				"title": "Go client task",
			})
		}

		for {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				break
			}
			fmt.Println("Update:", msg)
			if id, ok := msg["taskId"].(string); ok {
				taskID = id
			}
		}
		conn.Close()
		time.Sleep(backoff)
	}
}
```