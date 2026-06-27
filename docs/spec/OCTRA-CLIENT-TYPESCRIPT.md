# OCTRA TypeScript / Node.js Client

A minimal client for the Octra HTTP/JSON API using the built-in `fetch`
(available in Node.js 18+ and all modern browsers).

The full flow is: **register** to get an `api_key`, **create an environment**
(an AI CLI plus skills), then **chat**.

---

## 1. Reusable Client

```ts
interface LLMConfig {
  api_key: string;
  base_url: string;
  model: string;
}

export class OctraClient {
  private baseUrl: string;
  private apiKey?: string;

  constructor(baseUrl = "http://localhost:8080", apiKey?: string) {
    this.baseUrl = baseUrl.replace(/\/$/, "");
    this.apiKey = apiKey;
  }

  private headers(): Record<string, string> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (this.apiKey) headers["octra-api-token"] = this.apiKey;
    return headers;
  }

  private async post<T>(path: string, body: unknown): Promise<T> {
    const resp = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify(body),
    });
    if (!resp.ok) {
      throw new Error(`Octra ${path} failed: ${resp.status} ${await resp.text()}`);
    }
    return resp.json() as Promise<T>;
  }

  async register(email: string, password: string) {
    const data = await this.post<{ user_id: string; api_key: string }>(
      "/register",
      { email, password },
    );
    this.apiKey = data.api_key; // remember it for later calls
    return data;
  }

  async createEnvironment(llm: LLMConfig, cli = "", skills: string[] = []) {
    return this.post<{ user_id: string; agent_id: string; api_key: string }>(
      "/environment",
      { llm, agent: { cli }, skills },
    );
  }

  async chat(prompt: string, skills: string[] = []): Promise<string> {
    const data = await this.post<{ response: string }>("/api/chat", {
      prompt,
      skills,
    });
    return data.response;
  }
}
```

---

## 2. Full Flow Example

```ts
const client = new OctraClient("http://localhost:8080");

// 1. Register and capture the API key.
const account = await client.register("me@example.com", "secret");
console.log("user_id:", account.user_id);

// 2. Create an environment: an AI CLI plus some skills.
await client.createEnvironment(
  {
    api_key: "sk-...",
    base_url: "https://api.anthropic.com",
    model: "claude-sonnet-4-6",
  },
  "claude-code",
  ["filesystem", "github", "brave-search"],
);

// 3. Send a prompt.
const answer = await client.chat("write a csv parser", ["filesystem"]);
console.log(answer);
```
