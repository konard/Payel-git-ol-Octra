package skillapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type SearchResponse struct {
	Query      string  `json:"query"`
	SearchType string  `json:"searchType"`
	Skills     []Skill `json:"skills"`
	Count      int     `json:"count"`
	DurationMs int     `json:"duration_ms"`
}

type Skill struct {
	ID       string `json:"id"`
	SkillID  string `json:"skillId"`
	Name     string `json:"name"`
	Source   string `json:"source"`
}

type Client struct {
	BaseURL    string
	httpClient *http.Client
}

func New() *Client {
	return NewWithBaseURL("https://skills.sh")
}

func NewWithBaseURL(baseURL string) *Client {
	return NewWithClientAndBaseURL(&http.Client{Timeout: 30 * time.Second}, baseURL)
}

func NewWithClient(httpClient *http.Client) *Client {
	return NewWithClientAndBaseURL(httpClient, "https://skills.sh")
}

func NewWithClientAndBaseURL(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *Client) Search(query string, limit int) (*SearchResponse, error) {
	u, _ := url.Parse(c.BaseURL + "/api/search")
	u.RawQuery = url.Values{
		"q":     {query},
		"limit": {fmt.Sprintf("%d", limit)},
	}.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("skillapi: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("skillapi: read: %w", err)
	}

	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("skillapi: decode: %w", err)
	}
	return &result, nil
}

func InstallCmd(skill Skill) string {
	return fmt.Sprintf("npx skills add https://github.com/%s --skill %s", skill.Source, skill.SkillID)
}

var DefaultQueries = []string{
	"claude", "python", "aws", "react", "go", "docker", "typescript",
	"node", "javascript", "java", "rust", "kubernetes", "devops", "security",
	"ai", "machine-learning", "data", "frontend", "backend", "fullstack",
	"database", "api", "testing", "git", "linux", "mobile", "swift",
	"angular", "vue", "nextjs", "nest", "express", "fastapi", "django",
	"spring", "terraform", "ansible", "monitoring", "observability",
	"analytics", "blockchain", "web3", "solidity", "rust", "cplusplus",
	"csharp", "dotnet", "php", "ruby", "scala", "kotlin", "flutter",
	"react-native", "graphql", "grpc", "rest", "microservices",
	"serverless", "sass", "tailwind", "bootstrap", "material-ui",
	"redux", "mobx", "webpack", "vite", "rollup", "babel", "eslint",
	"prettier", "jest", "cypress", "playwright", "pytest", "junit",
	"selenium", "postgres", "mysql", "mongodb", "redis", "elasticsearch",
	"typesense", "kafka", "rabbitmq", "nginx", "traefik", "istio",
	"prometheus", "grafana", "datadog", "newrelic", "sentry",
	"openai", "anthropic", "gemini", "llama", "langchain",
	"mcp", "vector", "embedding", "rag", "agent", "copilot",
}
