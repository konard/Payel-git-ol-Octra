package context

import (
	"encoding/json"
	"strings"
	"time"

	"orchestrator/pkg/models"

	"github.com/google/uuid"
)

const (
	RedisTTL     = 5 * time.Minute
	MessageLimit = 20
)

type AIResponseContext struct {
	Scope       string `json:"scope"`
	TargetID    string `json:"target_id,omitempty"`
	ContextType string `json:"type"`
	Content     string `json:"content"`
	Important   bool   `json:"important,omitempty"`
}

func entryFromAIResponse(projectID uuid.UUID, sourceAgentID string, aiCtx *AIResponseContext) *models.ContextEntry {
	if aiCtx.Scope == "" {
		aiCtx.Scope = "global"
	}
	if aiCtx.ContextType == "" {
		aiCtx.ContextType = "message"
	}
	return &models.ContextEntry{
		ID:            uuid.New(),
		ProjectID:     projectID,
		Scope:         aiCtx.Scope,
		SourceAgentID: sourceAgentID,
		TargetID:      aiCtx.TargetID,
		ContextType:   aiCtx.ContextType,
		Content:       aiCtx.Content,
		Timestamp:     time.Now().UTC(),
		Important:     aiCtx.Important,
	}
}

func ExtractEntryFromJSON(projectID uuid.UUID, sourceAgentID, jsonStr string) *models.ContextEntry {
	var wrapper struct {
		Context *AIResponseContext `json:"context"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &wrapper); err != nil {
		return nil
	}
	if wrapper.Context == nil || wrapper.Context.Content == "" {
		return nil
	}
	return entryFromAIResponse(projectID, sourceAgentID, wrapper.Context)
}

func IsForgetRequest(content string) bool {
	s := strings.TrimSpace(strings.ToLower(content))
	return s == "forget" || s == "забудь" || s == "forget it" || s == "забудьте" || s == "forget this"
}
