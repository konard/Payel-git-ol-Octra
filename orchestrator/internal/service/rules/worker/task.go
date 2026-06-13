package worker

import (
	"context"
	"fmt"
	"log"
)

// createTaskMD — воркер пишет план работы (TASK.md) с помощью LLM.
// Для кодовых задач это разбивка по файлам/функциональности, для
// не-кодовых (research/document/presentation) — план содержания результата.
func (s *Service) createTaskMD(ctx context.Context, provider, model string, tokens map[string]string, role, description, taskMD, externalContext, taskType string) (string, error) {
	contextSection := ""
	if externalContext != "" {
		contextSection = "\n\nCONTEXT FROM OTHER WORKERS (use for coordination):\n" + externalContext
	}

	var prompt string
	switch taskType {
	case "research", "document", "presentation":
		prompt = documentPlanPrompt(taskType, role, description, taskMD, contextSection)
	default:
		prompt = fmt.Sprintf(`You are a %s developer on a project team.

YOUR ROLE: %s
TASK:
%s
%s

SCOPE FIDELITY (critical): plan EXACTLY what the task asks for — nothing more.
Do NOT invent requirements. If the task does not mention authentication, a database,
an ORM, validation, logging, or a frontend framework, then DO NOT add them. A request
like "express hello world server" needs ONE source file plus its manifest — no prisma,
jwt, bcrypt, zod, winston, or React. Only list a dependency if the task literally
requires it. When unsure, prefer fewer files and fewer dependencies.

Create TASK.md file with a task breakdown for your role. Include:
1. Files to create (the MINIMUM set)
2. Functionality to implement (only what was requested)
3. Dependencies (only those strictly required by the task — usually just the runtime/framework named)
4. How your work integrates with other workers (if context provided)

Return ONLY the content of TASK.md file.`, role, description, taskMD, contextSection)
	}

	resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 2048, 0.3)
	if err != nil {
		return "", err
	}
	log.Printf("Worker created plan (%s, type=%s)", role, taskType)
	return resp, nil
}

// documentPlanPrompt — план содержания для не-кодовых задач. Это НЕ финальный
// результат, а краткий план, который менеджер видит при ревью.
func documentPlanPrompt(taskType, role, description, taskMD, contextSection string) string {
	kind := "text document"
	switch taskType {
	case "research":
		kind = "research report"
	case "presentation":
		kind = "slide deck"
	}
	return fmt.Sprintf(`You are a %s producing a %s.

YOUR ROLE / ANGLE: %s
REQUEST:
%s
%s

Write a SHORT content plan (an outline) for your part of the %s. Include:
1. The main sections / slides you will produce
2. The key points each section will cover
3. How your part complements the other workers (if context provided)

Keep it concise (a bulleted outline). Return ONLY the outline in Markdown.`, role, kind, description, taskMD, contextSection, kind)
}

