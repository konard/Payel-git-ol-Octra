package service

import (
	"context"
	"errors"
	"log"

	"backend/internal/model"
	"backend/internal/nix"
	"backend/internal/repository"
)

// EnvProvisioner is the subset of nix.Manager the environment service needs.
// Declaring it as an interface keeps the service unit-testable.
type EnvProvisioner interface {
	CreateEnvironment(ctx context.Context, userID string, cli model.CLIType) error
	InstallSkill(ctx context.Context, userID string, skill model.Skill) error
}

// EnvironmentInput is the request payload for creating/updating an environment.
type EnvironmentInput struct {
	LLMAPIKey  string
	LLMBaseURL string
	LLMModel   string
	CLI        model.CLIType
	Skills     []string
}

// EnvironmentService provisions a user's personal MCP environment.
type EnvironmentService struct {
	agents     repository.AgentRepository
	skills     repository.SkillRepository
	userSkills repository.UserSkillRepository
	nix        EnvProvisioner
}

// NewEnvironmentService builds an EnvironmentService.
func NewEnvironmentService(
	agents repository.AgentRepository,
	skills repository.SkillRepository,
	userSkills repository.UserSkillRepository,
	provisioner EnvProvisioner,
) *EnvironmentService {
	return &EnvironmentService{agents: agents, skills: skills, userSkills: userSkills, nix: provisioner}
}

// Create provisions (or reconfigures) the environment for a user. It persists
// the agent, creates the Nix environment, then installs the requested skills.
func (s *EnvironmentService) Create(ctx context.Context, user *model.User, in EnvironmentInput) (*model.Agent, error) {
	agent := &model.Agent{
		UserID:     user.ID,
		LLMAPIKey:  in.LLMAPIKey,
		LLMBaseURL: in.LLMBaseURL,
		LLMModel:   in.LLMModel,
		CLI:        in.CLI,
		Active:     true,
	}
	if err := s.agents.Upsert(ctx, agent); err != nil {
		return nil, err
	}

	userID := user.ID.String()
	if err := s.nix.CreateEnvironment(ctx, userID, in.CLI); err != nil {
		return nil, err
	}

	for _, name := range in.Skills {
		skill, err := s.skills.GetByName(ctx, name)
		if errors.Is(err, repository.ErrNotFound) {
			log.Printf("environment: unknown skill %q requested by %s, skipping", name, userID)
			continue
		}
		if err != nil {
			return nil, err
		}

		status := "installed"
		if err := s.nix.InstallSkill(ctx, userID, *skill); err != nil {
			log.Printf("environment: failed installing skill %q: %v", name, err)
			status = "failed"
		}
		if err := s.userSkills.Add(ctx, &model.UserSkill{
			UserID:  user.ID,
			AgentID: agent.ID,
			SkillID: skill.ID,
			Status:  status,
		}); err != nil {
			return nil, err
		}
	}

	return agent, nil
}

var _ EnvProvisioner = (*nix.Manager)(nil)
