# OCTRA Go Client

A minimal client for the Octra HTTP/JSON API using only the standard library
(`net/http` and `encoding/json`).

The full flow is: **register** to get an `api_key`, **create an environment**
(an AI CLI plus skills), then **chat**.

---

## 1. Reusable Client

```go
package octra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL string
	APIKey  string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL, http: &http.Client{}}
}

type LLMConfig struct {
	Provider string `json:"provider,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
	BaseURL  string `json:"base_url,omitempty"`
	Model    string `json:"model,omitempty"`
}

func (c *Client) post(path string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("octra-api-token", c.APIKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("octra %s failed: %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) Register(username, email, password string) (string, error) {
	var out struct {
		UserID string `json:"user_id"`
		APIKey string `json:"api_key"`
	}
	err := c.post("/register", map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}, &out)
	if err == nil {
		c.APIKey = out.APIKey // remember it for later calls
	}
	return out.UserID, err
}

func (c *Client) CreateEnvironment(llm LLMConfig, cli string, skills []string) error {
	body := map[string]any{
		"llm":    llm,
		"agent":  map[string]string{"cli": cli},
		"skills": skills,
	}
	var out map[string]any
	return c.post("/environment", body, &out)
}

func (c *Client) Chat(prompt string, skills []string) (string, error) {
	body := map[string]any{"prompt": prompt, "skills": skills}
	var out struct {
		Response string `json:"response"`
	}
	err := c.post("/api/chat", body, &out)
	return out.Response, err
}
```

---

## 2. Full Flow Example

```go
package main

import (
	"fmt"
	"log"

	"example.com/octra"
)

func main() {
	client := octra.NewClient("http://localhost:8080")

	// 1. Register and capture the API key.
	userID, err := client.Register("me", "me@example.com", "secret")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("user_id:", userID)

	// 2. Create an environment: an AI CLI plus some skills.
	err = client.CreateEnvironment(octra.LLMConfig{
		Provider: "claude",
		APIKey:   "sk-...",
		BaseURL:  "https://api.anthropic.com",
		Model:    "claude-sonnet-4-6",
	}, "claude-code", []string{"filesystem", "github", "brave-search"})
	if err != nil {
		log.Fatal(err)
	}

	// 3. Send a prompt.
	answer, err := client.Chat("write a csv parser", []string{"filesystem"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(answer)
}
```
