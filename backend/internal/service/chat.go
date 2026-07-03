package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"backend/internal/cli"
	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

var ErrNoEnvironment = errors.New("no environment configured; create one via POST /environment")

var ErrEnvironmentInactive = errors.New("environment is inactive")

var ErrNoProviderNode = errors.New("no provider node found in environment canvas")

type OcawePortProvider interface {
	EnsureOcawe(ctx context.Context, spec cli.LaunchSpec) (int, error)
}

type EnvPathResolver interface {
	EnvPath(userID string) string
}

type DashboardEnvRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.DashboardEnvironment, error)
	ListByBuilding(ctx context.Context, building bool) ([]model.DashboardEnvironment, error)
	SetBuilding(ctx context.Context, id uuid.UUID, building bool) error
}

type CanvasNodeRepository interface {
	ListByEnvironment(ctx context.Context, envID uuid.UUID) ([]model.CanvasNode, error)
}

type ChatService struct {
	agents    repository.AgentRepository
	ocaweProv OcawePortProvider
	envPaths  EnvPathResolver
	httpCli   *http.Client

	dashboardEnvRepo DashboardEnvRepository
	canvasNodeRepo   CanvasNodeRepository

	ocaweHost    string
	ocaweBaseURL string
}

func NewChatService(agents repository.AgentRepository, ocaweProv OcawePortProvider, envPaths EnvPathResolver) *ChatService {
	return &ChatService{
		agents:    agents,
		ocaweProv: ocaweProv,
		envPaths:  envPaths,
		httpCli:   http.DefaultClient,
		ocaweHost: "127.0.0.1",
	}
}

func (s *ChatService) WithEnvironmentRepos(dashboardEnvRepo DashboardEnvRepository, canvasNodeRepo CanvasNodeRepository) *ChatService {
	s.dashboardEnvRepo = dashboardEnvRepo
	s.canvasNodeRepo = canvasNodeRepo
	return s
}

func (s *ChatService) WithBaseURL(baseURL string) *ChatService {
	if baseURL != "" {
		s.ocaweBaseURL = baseURL
	}
	return s
}

func (s *ChatService) Chat(ctx context.Context, user *model.User, prompt string, skills []string) (string, error) {
	agent, err := s.agents.GetByUserID(ctx, user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		return "", ErrNoEnvironment
	}
	if err != nil {
		return "", err
	}
	if !agent.Active {
		return "", ErrEnvironmentInactive
	}

	spec := cli.LaunchSpec{
		UserID:  user.ID.String(),
		EnvPath: s.envPaths.EnvPath(user.ID.String()),
	}

	port, err := s.ocaweProv.EnsureOcawe(ctx, spec)
	if err != nil {
		return "", fmt.Errorf("ocawe ensure: %w", err)
	}

	modelStr := resolveModel(agent)

	return s.sendChat(ctx, port, modelStr, prompt, agent.LLMAPIKey, agent.LLMBaseURL)
}

func (s *ChatService) ChatWithEnvironment(ctx context.Context, user *model.User, envID uuid.UUID, prompt string) (string, error) {
	env, err := s.dashboardEnvRepo.GetByID(ctx, envID)
	if err != nil {
		return "", fmt.Errorf("environment: %w", err)
	}
	if env.UserID != user.ID {
		return "", fmt.Errorf("not your environment")
	}

	nodes, err := s.canvasNodeRepo.ListByEnvironment(ctx, envID)
	if err != nil {
		return "", fmt.Errorf("canvas nodes: %w", err)
	}

	llmCfg, cliType, err := extractConfigFromNodes(nodes)
	if err != nil {
		return "", err
	}

	spec := cli.LaunchSpec{
		UserID:  envID.String(),
		EnvPath: s.envPaths.EnvPath(user.ID.String()),
	}

	port, err := s.ocaweProv.EnsureOcawe(ctx, spec)
	if err != nil {
		return "", fmt.Errorf("ocawe ensure: %w", err)
	}

	modelStr := resolveModelFromConfig(llmCfg, cliType)
	return s.sendChat(ctx, port, modelStr, prompt, llmCfg.APIKey, llmCfg.BaseURL)
}

func (s *ChatService) ChatStream(ctx context.Context, user *model.User, prompt string, skills []string, w io.Writer) error {
	agent, err := s.agents.GetByUserID(ctx, user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNoEnvironment
	}
	if err != nil {
		return err
	}
	if !agent.Active {
		return ErrEnvironmentInactive
	}

	spec := cli.LaunchSpec{
		UserID:  user.ID.String(),
		EnvPath: s.envPaths.EnvPath(user.ID.String()),
	}

	port, err := s.ocaweProv.EnsureOcawe(ctx, spec)
	if err != nil {
		return fmt.Errorf("ocawe ensure: %w", err)
	}

	modelStr := resolveModel(agent)
	_ = skills
	return s.sendChatStream(ctx, port, modelStr, prompt, agent.LLMAPIKey, agent.LLMBaseURL, w)
}

func (s *ChatService) ChatWithEnvironmentStream(ctx context.Context, user *model.User, envID uuid.UUID, prompt string, w io.Writer) error {
	env, err := s.dashboardEnvRepo.GetByID(ctx, envID)
	if err != nil {
		return fmt.Errorf("environment: %w", err)
	}
	if env.UserID != user.ID {
		return fmt.Errorf("not your environment")
	}

	nodes, err := s.canvasNodeRepo.ListByEnvironment(ctx, envID)
	if err != nil {
		return fmt.Errorf("canvas nodes: %w", err)
	}

	llmCfg, cliType, err := extractConfigFromNodes(nodes)
	if err != nil {
		return err
	}

	spec := cli.LaunchSpec{
		UserID:  envID.String(),
		EnvPath: s.envPaths.EnvPath(user.ID.String()),
	}

	port, err := s.ocaweProv.EnsureOcawe(ctx, spec)
	if err != nil {
		return fmt.Errorf("ocawe ensure: %w", err)
	}

	modelStr := resolveModelFromConfig(llmCfg, cliType)
	return s.sendChatStream(ctx, port, modelStr, prompt, llmCfg.APIKey, llmCfg.BaseURL, w)
}

func (s *ChatService) SyncEnvironmentBuilds(ctx context.Context) error {
	if s.ocaweBaseURL == "" {
		return nil
	}

	envs, err := s.dashboardEnvRepo.ListByBuilding(ctx, false)
	if err != nil {
		return fmt.Errorf("list unbuilt: %w", err)
	}

	for _, env := range envs {
		if err := s.syncOne(ctx, env.ID); err != nil {
			log.Printf("sync env %s: %v", env.ID, err)
			continue
		}
	}
	return nil
}

func (s *ChatService) SyncEnvironment(ctx context.Context, envID uuid.UUID) error {
	if s.ocaweBaseURL == "" {
		return nil
	}
	return s.syncOne(ctx, envID)
}

func (s *ChatService) syncOne(ctx context.Context, envID uuid.UUID) error {
	nodes, err := s.canvasNodeRepo.ListByEnvironment(ctx, envID)
	if err != nil {
		return fmt.Errorf("canvas nodes: %w", err)
	}

	llmCfg, cliType, err := extractConfigFromNodes(nodes)
	if errors.Is(err, ErrNoProviderNode) || llmCfg.APIKey == "" {
		return nil
	}
	if err != nil {
		return err
	}

	body := map[string]any{
		"provider": llmCfg.Provider,
		"api_key":  llmCfg.APIKey,
	}
	if llmCfg.BaseURL != "" {
		body["base_url"] = llmCfg.BaseURL
	}
	if llmCfg.Model != "" {
		body["model"] = llmCfg.Model
	}
	if cliType != "" {
		body["cli"] = string(cliType)
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	url := s.ocaweBaseURL + "/v1/environments/config"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("create config request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return fmt.Errorf("config request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ocawe config responded %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return s.dashboardEnvRepo.SetBuilding(ctx, envID, true)
}

func (s *ChatService) ChatOpenAIWithEnvironment(ctx context.Context, user *model.User, envID uuid.UUID, bodyJSON json.RawMessage, w io.Writer) error {
	env, err := s.dashboardEnvRepo.GetByID(ctx, envID)
	if err != nil {
		return fmt.Errorf("environment: %w", err)
	}
	if env.UserID != user.ID {
		return fmt.Errorf("not your environment")
	}

	nodes, err := s.canvasNodeRepo.ListByEnvironment(ctx, envID)
	if err != nil {
		return fmt.Errorf("canvas nodes: %w", err)
	}

	llmCfg, cliType, err := extractConfigFromNodes(nodes)
	if err != nil {
		return err
	}

	spec := cli.LaunchSpec{
		UserID:  envID.String(),
		EnvPath: s.envPaths.EnvPath(user.ID.String()),
	}

	port, err := s.ocaweProv.EnsureOcawe(ctx, spec)
	if err != nil {
		return fmt.Errorf("ocawe ensure: %w", err)
	}

	var req map[string]any
	if err := json.Unmarshal(bodyJSON, &req); err != nil {
		return fmt.Errorf("invalid body: %w", err)
	}

	if _, ok := req["model"]; !ok {
		req["model"] = resolveModelFromConfig(llmCfg, cliType)
	}
	if llmCfg.APIKey != "" {
		req["api_key"] = llmCfg.APIKey
	}
	if llmCfg.BaseURL != "" {
		req["base_url"] = llmCfg.BaseURL
	}
	req["stream"] = true

	newBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	var url string
	if s.ocaweBaseURL != "" {
		url = fmt.Sprintf("%s/v1/chat/completions", s.ocaweBaseURL)
	} else {
		url = fmt.Sprintf("http://%s:%d/v1/chat/completions", s.ocaweHost, port)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(newBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpCli.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ocawe request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ocawe responded %d: %s", resp.StatusCode, string(bodyBytes))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return err
		}
		if line == "" {
			if f, ok := w.(interface{ Flush() error }); ok {
				f.Flush()
			}
		}
	}
	return scanner.Err()
}

func (s *ChatService) ChatOpenAIWithEnvironmentSync(ctx context.Context, user *model.User, envID uuid.UUID, bodyJSON json.RawMessage) (string, error) {
	env, err := s.dashboardEnvRepo.GetByID(ctx, envID)
	if err != nil {
		return "", fmt.Errorf("environment: %w", err)
	}
	if env.UserID != user.ID {
		return "", fmt.Errorf("not your environment")
	}

	nodes, err := s.canvasNodeRepo.ListByEnvironment(ctx, envID)
	if err != nil {
		return "", fmt.Errorf("canvas nodes: %w", err)
	}

	llmCfg, cliType, err := extractConfigFromNodes(nodes)
	if err != nil {
		return "", err
	}

	spec := cli.LaunchSpec{
		UserID:  envID.String(),
		EnvPath: s.envPaths.EnvPath(user.ID.String()),
	}

	port, err := s.ocaweProv.EnsureOcawe(ctx, spec)
	if err != nil {
		return "", fmt.Errorf("ocawe ensure: %w", err)
	}

	var req map[string]any
	if err := json.Unmarshal(bodyJSON, &req); err != nil {
		return "", fmt.Errorf("invalid body: %w", err)
	}

	if _, ok := req["model"]; !ok {
		req["model"] = resolveModelFromConfig(llmCfg, cliType)
	}
	if llmCfg.APIKey != "" {
		req["api_key"] = llmCfg.APIKey
	}
	if llmCfg.BaseURL != "" {
		req["base_url"] = llmCfg.BaseURL
	}

	newBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	var url string
	if s.ocaweBaseURL != "" {
		url = fmt.Sprintf("%s/v1/chat/completions", s.ocaweBaseURL)
	} else {
		url = fmt.Sprintf("http://%s:%d/v1/chat/completions", s.ocaweHost, port)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(newBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpCli.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ocawe request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ocawe responded %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("decode ocawe response: %w", err)
	}

	var sb strings.Builder
	for _, choice := range chatResp.Choices {
		sb.WriteString(choice.Message.Content)
	}
	return sb.String(), nil
}

func (s *ChatService) sendChat(ctx context.Context, port int, modelStr, prompt string, apiKey, baseURL string) (string, error) {
	body := map[string]any{
		"model": modelStr,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	if apiKey != "" {
		body["api_key"] = apiKey
	}
	if baseURL != "" {
		body["base_url"] = baseURL
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	var url string
	if s.ocaweBaseURL != "" {
		url = fmt.Sprintf("%s/v1/chat/completions", s.ocaweBaseURL)
	} else {
		url = fmt.Sprintf("http://%s:%d/v1/chat/completions", s.ocaweHost, port)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return "", fmt.Errorf("ocawe request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ocawe responded %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("decode ocawe response: %w", err)
	}

	var sb strings.Builder
	for _, choice := range chatResp.Choices {
		sb.WriteString(choice.Message.Content)
	}
	return sb.String(), nil
}

func (s *ChatService) sendChatStream(ctx context.Context, port int, modelStr, prompt string, apiKey, baseURL string, w io.Writer) error {
	body := map[string]any{
		"model": modelStr,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": true,
	}
	if apiKey != "" {
		body["api_key"] = apiKey
	}
	if baseURL != "" {
		body["base_url"] = baseURL
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	var url string
	if s.ocaweBaseURL != "" {
		url = fmt.Sprintf("%s/v1/chat/completions", s.ocaweBaseURL)
	} else {
		url = fmt.Sprintf("http://%s:%d/v1/chat/completions", s.ocaweHost, port)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return fmt.Errorf("ocawe request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ocawe responded %d: %s", resp.StatusCode, string(bodyBytes))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return err
		}
		if line == "" {
			if f, ok := w.(interface{ Flush() error }); ok {
				f.Flush()
			}
		}
	}
	return scanner.Err()
}

type nodeConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

func extractConfigFromNodes(nodes []model.CanvasNode) (nodeConfig, model.CLIType, error) {
	var cfg nodeConfig
	var cliType model.CLIType

	for _, n := range nodes {
		var meta map[string]*string
		if n.Meta != "" {
			json.Unmarshal([]byte(n.Meta), &meta)
		}

		switch n.Kind {
		case "provider", "custom_provider":
			if meta != nil {
				if v := strPtrVal(meta["auth"]); v != "" {
					cfg.APIKey = v
				}
				if v := strPtrVal(meta["base_url"]); v != "" {
					cfg.BaseURL = v
				}
				if v := strPtrVal(meta["model"]); v != "" {
					cfg.Model = v
				}
				if v := strPtrVal(meta["provider"]); v != "" {
					cfg.Provider = v
				}
			}
			if cfg.Provider == "" {
				cfg.Provider = "openai"
			}
		case "cli":
			if meta != nil {
				if v := strPtrVal(meta["cli"]); v != "" {
					cliType = model.CLIType(v)
				}
			}
		}
	}

	if cfg.Provider == "" && cfg.APIKey == "" {
		return cfg, cliType, ErrNoProviderNode
	}
	return cfg, cliType, nil
}

func resolveModelFromConfig(cfg nodeConfig, cliType model.CLIType) string {
	if cliType != "" {
		return "cli/" + string(cliType)
	}
	if cfg.Provider != "" {
		model := cfg.Model
		if model == "" {
			model = "gpt-4o-mini"
		}
		return cfg.Provider + "/" + model
	}
	return "openai/gpt-4o-mini"
}

func strPtrVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func resolveModel(agent *model.Agent) string {
	if agent.CLI != "" {
		return "cli/" + string(agent.CLI)
	}
	if agent.LLMProvider != "" {
		model := agent.LLMModel
		if model == "" {
			model = "gpt-4o-mini"
		}
		return agent.LLMProvider + "/" + model
	}
	return "openai/gpt-4o-mini"
}
