# OCTRA Crystal Client

A minimal client for the Octra HTTP/JSON API using the standard library
(`http/client` and `json`).

The full flow is: **register** to get an `api_key`, **create an environment**
(an AI CLI plus skills), then **chat**.

---

## 1. Reusable Client

```crystal
require "http/client"
require "json"

class OctraClient
  def initialize(@base_url : String = "http://localhost:8080", @api_key : String? = nil)
  end

  def register(username : String, email : String, password : String) : JSON::Any
    data = post("/register", {username: username, email: email, password: password})
    @api_key = data["api_key"].as_s # remember it for later calls
    data
  end

  def create_environment(llm, cli : String = "", skills : Array(String) = [] of String) : JSON::Any
    post("/environment", {llm: llm, agent: {cli: cli}, skills: skills})
  end

  def chat(prompt : String, skills : Array(String) = [] of String) : String
    post("/api/chat", {prompt: prompt, skills: skills})["response"].as_s
  end

  private def post(path : String, body) : JSON::Any
    headers = HTTP::Headers{"Content-Type" => "application/json"}
    if key = @api_key
      headers["octra-api-token"] = key
    end

    response = HTTP::Client.post("#{@base_url}#{path}", headers: headers, body: body.to_json)
    unless response.success?
      raise "Octra #{path} failed: #{response.status_code} #{response.body}"
    end
    JSON.parse(response.body)
  end
end
```

---

## 2. Full Flow Example

```crystal
client = OctraClient.new("http://localhost:8080")

# 1. Register and capture the API key.
account = client.register("me", "me@example.com", "secret")
puts "user_id: #{account["user_id"]}"

# 2. Create an environment: an AI CLI plus some skills.
client.create_environment(
  llm: {
    provider: "claude",
    api_key:  "sk-...",
    base_url: "https://api.anthropic.com",
    model:    "claude-sonnet-4-6",
  },
  cli: "claude-code",
  skills: ["filesystem", "github", "brave-search"]
)

# 3. Send a prompt.
answer = client.chat("write a csv parser", skills: ["filesystem"])
puts answer
```
