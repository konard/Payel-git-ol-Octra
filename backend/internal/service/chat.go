package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	ocaweAddr string
}

func NewChatService(agents repository.AgentRepository, ocaweProv OcawePortProvider, envPaths EnvPathResolver) *ChatService {
	return &ChatService{
		agents:    agents,
		ocaweProv: ocaweProv,
		envPaths:  envPaths,
		httpCli:   http.DefaultClient,
		ocaweAddr: "127.0.0.1",
	}
}

func (s *ChatService) WithEnvironmentRepos(dashboardEnvRepo DashboardEnvRepository, canvasNodeRepo CanvasNodeRepository) *ChatService {
	s.dashboardEnvRepo = dashboardEnvRepo
	s.canvasNodeRepo = canvasNodeRepo
	return s
}

func (s *ChatService) WithOcaweAddr(addr string) *ChatService {
	if addr != "" {
		s.ocaweAddr = addr
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
		LLM: cli.LLMConfig{
			Provider: agent.LLMProvider,
			APIKey:   agent.LLMAPIKey,
			BaseURL:  agent.LLMBaseURL,
			Model:    agent.LLMModel,
		},
	}

	port, err := s.ocaweProv.EnsureOcawe(ctx, spec)
	if err != nil {
		return "", fmt.Errorf("ocawe ensure: %w", err)
	}

	modelStr := resolveModel(agent)

	return s.sendChat(ctx, port, modelStr, prompt)
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
		CLI:     cliType,
		LLM:     llmCfg,
	}

	port, err := s.ocaweProv.EnsureOcawe(ctx, spec)
	if err != nil {
		return "", fmt.Errorf("ocawe ensure: %w", err)
	}

	modelStr := resolveModelFromConfig(llmCfg, cliType)
	return s.sendChat(ctx, port, modelStr, prompt)
}

func (s *ChatService) sendChat(ctx context.Context, port int, modelStr, prompt string) (string, error) {
	body := map[string]any{
		"model": modelStr,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s:%d/v1/chat/completions", s.ocaweAddr, port)
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

func extractConfigFromNodes(nodes []model.CanvasNode) (cli.LLMConfig, model.CLIType, error) {
	var llmCfg cli.LLMConfig
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
					llmCfg.APIKey = v
				}
				if v := strPtrVal(meta["base_url"]); v != "" {
					llmCfg.BaseURL = v
				}
				if v := strPtrVal(meta["model"]); v != "" {
					llmCfg.Model = v
				}
				if v := strPtrVal(meta["provider"]); v != "" {
					llmCfg.Provider = v
				}
			}
			if llmCfg.Provider == "" {
				llmCfg.Provider = "openai"
			}
		case "cli":
			if meta != nil {
				if v := strPtrVal(meta["cli"]); v != "" {
					cliType = model.CLIType(v)
				}
			}
		}
	}

	if llmCfg.Provider == "" && llmCfg.APIKey == "" {
		return llmCfg, cliType, ErrNoProviderNode
	}
	return llmCfg, cliType, nil
}

func resolveModelFromConfig(llmCfg cli.LLMConfig, cliType model.CLIType) string {
	if cliType != "" {
		return "cli/" + string(cliType)
	}
	if llmCfg.Provider != "" {
		model := llmCfg.Model
		if model == "" {
			model = "gpt-4o-mini"
		}
		return llmCfg.Provider + "/" + model
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
