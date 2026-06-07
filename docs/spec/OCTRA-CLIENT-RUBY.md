# OCTRA Ruby Client

## Reconnection Example

```ruby
require 'websocket-client-simple'
require 'json'

task_id = nil
backoff = 1
uri = "wss://your-domain.com/ws"

loop do
  begin
    ws = WebSocket::Client::Simple.connect uri

    if task_id
      ws.send({ type: 'resume', task_id: task_id }.to_json)
    else
      ws.send({
        username: "RubyClient",
        user_id: "00000000-0000-0000-0000-000000000006",
        title: "Ruby Task",
        description: "Test from Ruby",
        meta: { model: "your-model", provider: "provider" },
        tokens: { provider: "your-api-key" }
      }.to_json)
    end

    ws.on :message do |msg|
      data = JSON.parse(msg.data)
      puts "Update: #{data}"
      task_id = data['taskId'] if data['taskId']
    end

    ws.on :close do
      sleep backoff
      backoff = [backoff * 2, 30].min
    end
  rescue => e
    puts e
    sleep backoff
  end
end
```