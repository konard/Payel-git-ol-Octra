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
      ws.send({type: "create", title: "Crystal Task"}.to_json)
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