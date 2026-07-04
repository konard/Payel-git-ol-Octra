# OCTRA Ruby Client

A minimal client for the Octra HTTP/JSON API using only the standard library
(`net/http` and `json`).

The full flow is: **register** to get an `api_key`, **create an environment**
(an AI CLI plus skills), then **chat**.

---

## 1. Reusable Client

```ruby
require 'net/http'
require 'json'
require 'uri'

class OctraClient
  def initialize(base_url = 'http://localhost:8080', api_key: nil)
    @base_url = base_url.chomp('/')
    @api_key = api_key
  end

  def register(username, email, password)
    data = post('/register', username: username, email: email, password: password)
    @api_key = data['api_key'] # remember it for later calls
    data
  end

  def create_environment(llm:, cli: '', skills: [])
    post('/environment', llm: llm, agent: { cli: cli }, skills: skills)
  end

  def chat(prompt, skills: [])
    post('/api/chat', prompt: prompt, skills: skills)['response']
  end

  private

  def post(path, body)
    uri = URI("#{@base_url}#{path}")
    req = Net::HTTP::Post.new(uri)
    req['Content-Type'] = 'application/json'
    req['octra-api-token'] = @api_key if @api_key
    req.body = body.to_json

    res = Net::HTTP.start(uri.hostname, uri.port,
                          use_ssl: uri.scheme == 'https') do |http|
      http.request(req)
    end

    unless res.is_a?(Net::HTTPSuccess)
      raise "Octra #{path} failed: #{res.code} #{res.body}"
    end

    JSON.parse(res.body)
  end
end
```

---

## 2. Full Flow Example

```ruby
client = OctraClient.new('http://localhost:8080')

# 1. Register and capture the API key.
account = client.register('me', 'me@example.com', 'secret')
puts "user_id: #{account['user_id']}"

# 2. Create an environment: an AI CLI plus some skills.
client.create_environment(
  llm: {
    provider: 'claude',
    api_key: 'sk-...',
    base_url: 'https://api.anthropic.com',
    model: 'claude-sonnet-4-6'
  },
  cli: 'claude-code',
  skills: %w[filesystem github brave-search]
)

# 3. Send a prompt.
answer = client.chat('write a csv parser', skills: %w[filesystem])
puts answer
```
