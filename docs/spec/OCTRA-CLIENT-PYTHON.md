# OCTRA Python Client

A minimal client for the Octra HTTP/JSON API using the popular
[`requests`](https://pypi.org/project/requests/) library
(`pip install requests`).

The full flow is: **register** to get an `api_key`, **create an environment**
(an AI CLI plus skills), then **chat**.

---

## 1. Reusable Client

```python
import requests


class OctraClient:
    def __init__(self, base_url="http://localhost:8080", api_key=None):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.session = requests.Session()

    def _headers(self):
        headers = {"Content-Type": "application/json"}
        if self.api_key:
            headers["octra-api-token"] = self.api_key
        return headers

    def register(self, email, password):
        resp = self.session.post(
            f"{self.base_url}/register",
            json={"email": email, "password": password},
            headers=self._headers(),
        )
        resp.raise_for_status()
        data = resp.json()
        self.api_key = data["api_key"]  # remember it for later calls
        return data

    def create_environment(self, llm, cli=None, skills=None):
        body = {"llm": llm, "agent": {"cli": cli or ""}, "skills": skills or []}
        resp = self.session.post(
            f"{self.base_url}/environment",
            json=body,
            headers=self._headers(),
        )
        resp.raise_for_status()
        return resp.json()

    def chat(self, prompt, skills=None):
        resp = self.session.post(
            f"{self.base_url}/api/chat",
            json={"prompt": prompt, "skills": skills or []},
            headers=self._headers(),
        )
        resp.raise_for_status()
        return resp.json()["response"]
```

---

## 2. Full Flow Example

```python
client = OctraClient(base_url="http://localhost:8080")

# 1. Register and capture the API key.
account = client.register("me@example.com", "secret")
print("user_id:", account["user_id"])

# 2. Create an environment: an AI CLI plus some skills.
client.create_environment(
    llm={
        "api_key": "sk-...",
        "base_url": "https://api.anthropic.com",
        "model": "claude-sonnet-4-6",
    },
    cli="claude-code",
    skills=["filesystem", "github", "brave-search"],
)

# 3. Send a prompt.
answer = client.chat("write a csv parser", skills=["filesystem"])
print(answer)
```
