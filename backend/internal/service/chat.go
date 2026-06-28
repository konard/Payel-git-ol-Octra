package service

import (
	"context"
	"errors"

	"backend/internal/cli"
	"backend/internal/llm"
	"backend/internal/model"
	"backend/internal/repository"
)

// ErrNoEnvironment is returned when a user chats before creating an environment.
var ErrNoEnvironment = errors.New("no environment configured; create one via POST /environment")

// ErrEnvironmentInactive is returned when the agent is disabled.
var ErrEnvironmentInactive = errors.New("environment is inactive")

// CLIRouter sends a prompt to a user's persistent CLI process.
type CLIRouter interface {
	Send(ctx context.Context, spec cli.LaunchSpec, prompt string) (string, error)
}

// LLMClient performs a direct LLM completion (proxy mode).
type LLMClient interface {
	Complete(ctx context.Context, req llm.Request) (string, error)
}

// EnvPathResolver resolves a user's environment directory.
type EnvPathResolver interface {
	EnvPath(userID string) string
}

// ChatService routes chat requests either to a CLI subprocess or, when no CLI
// is configured, straight to the LLM.
type ChatService struct {
	agents   repository.AgentRepository
	cli      CLIRouter
	llm      LLMClient
	envPaths EnvPathResolver
}

// NewChatService builds a ChatService.
func NewChatService(agents repository.AgentRepository, cliRouter CLIRouter, llmClient LLMClient, envPaths EnvPathResolver) *ChatService {
	return &ChatService{agents: agents, cli: cliRouter, llm: llmClient, envPaths: envPaths}
}

// Chat handles a single prompt for the given user. The skills slice lists which
// skills are enabled for this request (a skill installed in the environment can
// be omitted to disable it).
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

	// CLI mode: route to the user's persistent subprocess.
	if agent.CLI != "" {
		spec := cli.LaunchSpec{
			UserID:  user.ID.String(),
			EnvPath: s.envPaths.EnvPath(user.ID.String()),
			CLI:     agent.CLI,
			LLM: cli.LLMConfig{
				Provider: agent.LLMProvider,
				APIKey:   agent.LLMAPIKey,
				BaseURL:  agent.LLMBaseURL,
				Model:    agent.LLMModel,
			},
		}
		return s.cli.Send(ctx, spec, prompt)
	}

	// Proxy mode: forward straight to the LLM.
	return s.llm.Complete(ctx, llm.Request{
		Provider: agent.LLMProvider,
		APIKey:   agent.LLMAPIKey,
		BaseURL:  agent.LLMBaseURL,
		Model:    agent.LLMModel,
		Prompt:   prompt,
	})
}
