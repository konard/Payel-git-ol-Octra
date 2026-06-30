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
)

var ErrNoEnvironment = errors.New("no environment configured; create one via POST /environment")

var ErrEnvironmentInactive = errors.New("environment is inactive")

type OcawePortProvider interface {
	EnsureOcawe(ctx context.Context, spec cli.LaunchSpec) (int, error)
}

type EnvPathResolver interface {
	EnvPath(userID string) string
}

type ChatService struct {
	agents    repository.AgentRepository
	ocaweProv OcawePortProvider
	envPaths  EnvPathResolver
	httpCli   *http.Client
}

func NewChatService(agents repository.AgentRepository, ocaweProv OcawePortProvider, envPaths EnvPathResolver) *ChatService {
	return &ChatService{
		agents:    agents,
		ocaweProv: ocaweProv,
		envPaths:  envPaths,
		httpCli:   http.DefaultClient,
	}
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

	url := fmt.Sprintf("http://127.0.0.1:%d/v1/chat/completions", port)
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
