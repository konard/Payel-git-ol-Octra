package boss

import (
	"context"
	"sort"
	"strings"

	"nodes/internal/prompts"
	"nodes/internal/service/rules"
)

// collectSolutionMarkdown — собирает Markdown-документы из папки solution/,
// произведённые воркерами не-кодовых задач, в один текст. Возвращает также
// количество исходных документов (нужно, чтобы решить, требуется ли синтез).
func collectSolutionMarkdown(results []*rules.ManagerResult) (string, int) {
	type doc struct {
		path    string
		content string
	}
	var docs []doc
	seen := map[string]bool{}
	for _, mr := range results {
		for _, wr := range mr.WorkerResults {
			for path, content := range wr.Files {
				lower := strings.ToLower(path)
				if !strings.HasSuffix(lower, ".md") && !strings.HasSuffix(lower, ".markdown") {
					continue
				}
				if seen[path] {
					continue
				}
				seen[path] = true
				docs = append(docs, doc{path: path, content: content})
			}
		}
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].path < docs[j].path })

	var b strings.Builder
	for _, d := range docs {
		if b.Len() > 0 {
			b.WriteString("\n\n---\n\n")
		}
		b.WriteString(d.content)
	}
	return b.String(), len(docs)
}

// synthesizeSolution — менеджерский синтез: сливает Markdown-результаты
// нескольких воркеров в один связный документ, проверяя факты и убирая
// повторы. Если AI недоступен — возвращает "" (вызывающий код оставит исходную
// механическую склейку).
func (s *Service) synthesizeSolution(
	ctx context.Context, provider, model string, tokens map[string]string,
	decision *DecisionResult, topic, workerOutputs string,
) string {
	managerRole := "lead"
	if len(decision.ManagerRoles) > 0 && decision.ManagerRoles[0].Role != "" {
		managerRole = decision.ManagerRoles[0].Role
	}
	outputs := workerOutputs
	if len(outputs) > 12000 {
		outputs = outputs[:12000]
	}
	prompt := prompts.ManagerSynthesis(managerRole, decision.TaskType, topic, outputs)
	resp, err := s.agentsClient.GenerateFromTask(ctx, provider, model, prompt, tokens)
	if err != nil || strings.TrimSpace(resp) == "" {
		return ""
	}
	return strings.TrimSpace(resp)
}

// buildChatSummary — босс готовит короткий текстовый ответ для чата по итогам
// research/document/presentation задачи (полная версия уже лежит во вкладке Solution).
func (s *Service) buildChatSummary(
	ctx context.Context, provider, model string, tokens map[string]string,
	taskType, topic, fullDocument string,
) string {
	full := fullDocument
	if len(full) > 6000 {
		full = full[:6000]
	}
	prompt := prompts.BossChatSummary(taskType, topic, full)
	resp, err := s.agentsClient.GenerateFromTask(ctx, provider, model, prompt, tokens)
	if err != nil || strings.TrimSpace(resp) == "" {
		// Фолбэк: краткое сообщение без AI.
		return "Your " + taskType + " is ready — open the Solution tab to see the full result."
	}
	return strings.TrimSpace(resp)
}
