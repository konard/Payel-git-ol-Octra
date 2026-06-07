# OCTRA Crystal Client

## Reconnection + Resume

```crystal
require "http/web_socket"
require "json"

task_id = nil
backoff = 1
uri = "wss://your-domain.com/ws"

loop do
  begin
    ws = HTTP::WebSocket.new(uri)

    if task_id
      ws.send({type: "resume", taskId: task_id}.to_json)
    else
      ws.send({
        username: "CrystalClient",
        user_id: "00000000-0000-0000-0000-000000000007",
        title: "Crystal Task",
        description: "Test from Crystal",
        meta: { model: "your-model", provider: "provider" },
        tokens: { provider: "your-api-key" }
      }.to_json)
    end

    ws.on_message do |message|
      data = JSON.parse(message)
      puts "Update: #{data}"
      task_id = data["taskId"]? if data["taskId"]?
    end

    ws.run
  rescue
    sleep backoff.seconds
    backoff = Math.min(backoff * 2, 30)
  end
end
```