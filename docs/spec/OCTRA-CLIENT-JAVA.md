# OCTRA Java Client

A minimal client for the Octra HTTP/JSON API using the built-in
`java.net.http.HttpClient` (Java 11+). JSON bodies are small, so this example
builds them with a tiny helper instead of pulling in a JSON library.

The full flow is: **register** to get an `api_key`, **create an environment**
(an AI CLI plus skills), then **chat**.

---

## 1. Reusable Client

```java
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class OctraClient {
    private final String baseUrl;
    private String apiKey;
    private final HttpClient http = HttpClient.newHttpClient();

    public OctraClient(String baseUrl) {
        this.baseUrl = baseUrl;
    }

    private String post(String path, String jsonBody) throws Exception {
        HttpRequest.Builder builder = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(jsonBody));
        if (apiKey != null) {
            builder.header("octra-api-token", apiKey);
        }

        HttpResponse<String> resp = http.send(builder.build(),
                HttpResponse.BodyHandlers.ofString());
        if (resp.statusCode() >= 300) {
            throw new RuntimeException("Octra " + path + " failed: "
                    + resp.statusCode() + " " + resp.body());
        }
        return resp.body();
    }

    // Tiny extractor for flat string fields like "api_key" or "response".
    private static String field(String json, String name) {
        String needle = "\"" + name + "\"";
        int i = json.indexOf(needle);
        if (i < 0) return null;
        int start = json.indexOf('"', json.indexOf(':', i) + 1) + 1;
        int end = json.indexOf('"', start);
        return json.substring(start, end);
    }

    public String register(String email, String password) throws Exception {
        String body = "{\"email\":\"" + email + "\",\"password\":\"" + password + "\"}";
        String resp = post("/register", body);
        this.apiKey = field(resp, "api_key"); // remember it for later calls
        return field(resp, "user_id");
    }

    public void createEnvironment(String llmApiKey, String llmBaseUrl,
                                  String model, String cli, String[] skills) throws Exception {
        StringBuilder skillsJson = new StringBuilder("[");
        for (int i = 0; i < skills.length; i++) {
            if (i > 0) skillsJson.append(',');
            skillsJson.append('"').append(skills[i]).append('"');
        }
        skillsJson.append(']');

        String body = "{"
                + "\"llm\":{\"api_key\":\"" + llmApiKey + "\","
                + "\"base_url\":\"" + llmBaseUrl + "\","
                + "\"model\":\"" + model + "\"},"
                + "\"agent\":{\"cli\":\"" + cli + "\"},"
                + "\"skills\":" + skillsJson
                + "}";
        post("/environment", body);
    }

    public String chat(String prompt, String[] skills) throws Exception {
        StringBuilder skillsJson = new StringBuilder("[");
        for (int i = 0; i < skills.length; i++) {
            if (i > 0) skillsJson.append(',');
            skillsJson.append('"').append(skills[i]).append('"');
        }
        skillsJson.append(']');

        String body = "{\"prompt\":\"" + prompt + "\",\"skills\":" + skillsJson + "}";
        return field(post("/api/chat", body), "response");
    }
}
```

> For production code, prefer a real JSON library (Jackson or Gson) over manual
> string building so values are escaped correctly.

---

## 2. Full Flow Example

```java
public class Main {
    public static void main(String[] args) throws Exception {
        OctraClient client = new OctraClient("http://localhost:8080");

        // 1. Register and capture the API key.
        String userId = client.register("me@example.com", "secret");
        System.out.println("user_id: " + userId);

        // 2. Create an environment: an AI CLI plus some skills.
        client.createEnvironment(
                "sk-...",
                "https://api.anthropic.com",
                "claude-sonnet-4-6",
                "claude-code",
                new String[]{"filesystem", "github", "brave-search"});

        // 3. Send a prompt.
        String answer = client.chat("write a csv parser", new String[]{"filesystem"});
        System.out.println(answer);
    }
}
```
