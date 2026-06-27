# OCTRA C# Client

A minimal client for the Octra HTTP/JSON API using `HttpClient` and
`System.Text.Json` (built into modern .NET).

The full flow is: **register** to get an `api_key`, **create an environment**
(an AI CLI plus skills), then **chat**.

---

## 1. Reusable Client

```csharp
using System;
using System.Net.Http;
using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;

public record LLMConfig(
    [property: JsonPropertyName("api_key")] string ApiKey,
    [property: JsonPropertyName("base_url")] string BaseUrl,
    [property: JsonPropertyName("model")] string Model);

public class OctraClient
{
    private readonly HttpClient _http = new();
    private readonly string _baseUrl;
    private string? _apiKey;

    public OctraClient(string baseUrl = "http://localhost:8080")
    {
        _baseUrl = baseUrl.TrimEnd('/');
    }

    private async Task<JsonElement> PostAsync(string path, object body)
    {
        using var req = new HttpRequestMessage(HttpMethod.Post, _baseUrl + path)
        {
            Content = JsonContent.Create(body)
        };
        if (_apiKey is not null)
            req.Headers.Add("octra-api-token", _apiKey);

        var resp = await _http.SendAsync(req);
        var text = await resp.Content.ReadAsStringAsync();
        if (!resp.IsSuccessStatusCode)
            throw new Exception($"Octra {path} failed: {(int)resp.StatusCode} {text}");

        return JsonDocument.Parse(text).RootElement;
    }

    public async Task<string> RegisterAsync(string email, string password)
    {
        var root = await PostAsync("/register", new { email, password });
        _apiKey = root.GetProperty("api_key").GetString(); // remember it for later calls
        return root.GetProperty("user_id").GetString()!;
    }

    public Task CreateEnvironmentAsync(LLMConfig llm, string cli = "", string[]? skills = null)
        => PostAsync("/environment", new
        {
            llm,
            agent = new { cli },
            skills = skills ?? Array.Empty<string>()
        });

    public async Task<string> ChatAsync(string prompt, string[]? skills = null)
    {
        var root = await PostAsync("/api/chat", new
        {
            prompt,
            skills = skills ?? Array.Empty<string>()
        });
        return root.GetProperty("response").GetString()!;
    }
}
```

---

## 2. Full Flow Example

```csharp
var client = new OctraClient("http://localhost:8080");

// 1. Register and capture the API key.
var userId = await client.RegisterAsync("me@example.com", "secret");
Console.WriteLine($"user_id: {userId}");

// 2. Create an environment: an AI CLI plus some skills.
await client.CreateEnvironmentAsync(
    new LLMConfig("sk-...", "https://api.anthropic.com", "claude-sonnet-4-6"),
    cli: "claude-code",
    skills: new[] { "filesystem", "github", "brave-search" });

// 3. Send a prompt.
var answer = await client.ChatAsync("write a csv parser", new[] { "filesystem" });
Console.WriteLine(answer);
```
