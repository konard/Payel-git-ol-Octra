package boss

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/search"
)

// searchNodeRole — роль ноды поиска на канве (issue #97). Менеджер/воркер с этой
// ролью трактуется не как обычный исполнитель, а как нода поиска: она не
// генерирует файлы, а выполняет веб-поиск для других нод.
const searchNodeRole = "search"

// isSearchNodeRole — роль соответствует ноде поиска (учитываем синонимы).
func isSearchNodeRole(role string) bool {
	r := strings.ToLower(strings.TrimSpace(role))
	return r == searchNodeRole || r == "web_search" || r == "websearch" || r == "web-search"
}

// explicitSearchNode сканирует канву (predefined managers/workers) и возвращает,
// размещена ли пользователем нода поиска, и какую модель/провайдер она задаёт.
// Модель берётся из meta (search_model/search_provider) либо из описания ноды
// в формате "model: <name>".
func explicitSearchNode(req *CreateTaskRequest) (explicit bool, model, provider string) {
	if req == nil {
		return false, "", ""
	}
	if req.Meta != nil {
		model = strings.TrimSpace(req.Meta["search_model"])
		provider = strings.TrimSpace(req.Meta["search_provider"])
	}
	for _, m := range req.PredefinedManagers {
		if isSearchNodeRole(m.Role) {
			explicit = true
			if mm := parseModelHint(m.Description); mm != "" && model == "" {
				model = mm
			}
		}
		for _, w := range m.Workers {
			if isSearchNodeRole(w.Role) {
				explicit = true
				if mm := parseModelHint(w.Description); mm != "" && model == "" {
					model = mm
				}
			}
		}
	}
	return explicit, model, provider
}

// parseModelHint вытаскивает имя модели из описания ноды вида "model: gpt-4o".
func parseModelHint(desc string) string {
	d := strings.ToLower(desc)
	idx := strings.Index(d, "model:")
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(desc[idx+len("model:"):])
	if i := strings.IndexAny(rest, " \n\t,;"); i >= 0 {
		rest = rest[:i]
	}
	return strings.TrimSpace(rest)
}

// runSearchNode планирует и (при необходимости) выполняет веб-поиск для задачи.
// Возвращает отформатированный блок результатов для промпта, Markdown-список
// источников и саму ноду (для логов/прогресса). Если поиск не нужен или недоступен,
// блок и источники пустые.
func (s *Service) runSearchNode(
	ctx context.Context,
	req *CreateTaskRequest,
	attachTo string,
	progress rules.ProgressFunc,
) (block, sources string, node search.Node) {
	taskModel, taskProvider := pickProviderModel(req.Meta)
	explicit, nodeModel, nodeProvider := explicitSearchNode(req)

	text := strings.TrimSpace(req.Title + "\n" + req.Description)
	node = search.PlanNode(text, explicit, nodeModel, nodeProvider, taskModel, taskProvider, attachTo)

	if !node.Enabled {
		return "", "", node
	}
	if s.searchClient == nil || !search.Enabled() {
		log.Printf("[Search] node enabled but web search is unavailable (client/env) — answering without sources")
		return "", "", node
	}

	// Сообщаем фронтенду о ноде поиска, чтобы её можно было отрисовать на канве
	// и прикрепить к узлу, который ею пользуется.
	emitSearchNode(progress, node, "searching", "", 0)

	queries := buildSearchNodeQueries(req.Title, req.Description)
	for _, q := range queries {
		emitSearchNode(progress, node, "query", q, 0)
	}

	searchConfig := parseSearchNodeConfig(req.Meta["search"])
	seen := make(map[string]struct{})
	var merged []search.Result
	var firstErr error

	addResults := func(results []search.Result) {
		for _, r := range results {
			key := strings.ToLower(strings.TrimRight(strings.TrimSpace(r.URL), "/"))
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, r)
		}
	}

	if searchConfig != nil {
		mp := search.NewModelProvider(*searchConfig)
		mp.SetProgressReporter(func(pct int32, msg string, data map[string]string) {
			if progress != nil {
				progress(pct, msg, data)
			}
		})
		modelResults, err := search.NewClientWithProvider(mp).Research(ctx, text, queries, 2, 4)
		if err != nil {
			firstErr = err
			log.Printf("[Search] AI search provider failed: %v", err)
		} else {
			addResults(modelResults)
		}
	}

	if s.searchClient != nil && search.Enabled() {
		webResults, err := s.searchClient.Research(ctx, text, queries, 5, 8)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			log.Printf("[Search] DuckDuckGo search failed: %v", err)
		} else {
			addResults(webResults)
		}
	}

	if len(merged) == 0 {
		if firstErr != nil {
			log.Printf("[Search] node research failed: %v", firstErr)
			emitSearchNode(progress, node, "error", "", 0)
			return "", "", node
		}
		log.Printf("[Search] node found no results for %d queries", len(queries))
		emitSearchNode(progress, node, "no_results", "", 0)
		return "", "", node
	}

	ranked := search.RankBM25(text, merged)
	if len(ranked) > 8 {
		ranked = ranked[:8]
	}

	log.Printf("[Search] node attached to %q: %d queries → %d sources (model=%s)", node.AttachTo, len(queries), len(ranked), node.Model)
	emitSearchNode(progress, node, "done", strings.Join(queries, " | "), len(ranked))
	return search.FormatForPrompt(ranked), search.FormatSourcesMarkdown(ranked), node
}

// emitSearchNode отправляет прогресс-апдейт о ноде поиска во фронтенд. Ключи в
// data позволяют UI нарисовать ноду «Search» и прикрепить её к нужному узлу.
func emitSearchNode(progress rules.ProgressFunc, node search.Node, phase, queries string, resultCount int) {
	if progress == nil {
		return
	}
	data := map[string]string{
		"node_type":          searchNodeRole,
		"search_node":        "true",
		"search_phase":       phase,
		"search_model":       node.Model,
		"search_attach_to":   node.AttachTo,
		"search_explicit":    boolStr(node.Explicit),
		"search_created":     boolStr(node.Created),
		"search_reason":      node.Reason,
		"search_node_result": strconv.Itoa(resultCount),
	}
	if queries != "" {
		data["search_queries"] = queries
	}
	msg := "Search node looking up information"
	switch phase {
	case "query":
		msg = fmt.Sprintf("Searching: %s", queries)
	case "error":
		msg = "Search failed — no results from any provider"
	case "no_results":
		msg = "Search found no results"
	case "done":
		if resultCount > 0 {
			msg = fmt.Sprintf("Search node found %d sources", resultCount)
		} else {
			msg = "Search node finished"
		}
	}
	progress(20, msg, data)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// buildSearchNodeQueries формирует поисковые запросы из заголовка и описания задачи.
func buildSearchNodeQueries(title, description string) []string {
	var queries []string
	add := func(s string) {
		s = cleanSearchQuery(s)
		if s != "" {
			queries = append(queries, s)
		}
	}
	add(title)
	// Первая строка описания часто несёт сам вопрос.
	add(firstNonEmptyLine(description))
	// Заголовок + ключевые слова описания как уточняющий запрос.
	combined := cleanSearchQuery(title + " " + firstNonEmptyLine(description))
	add(combined)
	return dedupeSearchQueries(queries)
}

func cleanSearchQuery(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 120 {
		s = s[:120]
		if i := strings.LastIndexByte(s, ' '); i > 0 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			return t
		}
	}
	return ""
}

type searchNodeConfig struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"base-url"`
	APIKey   string `json:"api-key"`
	Striming bool   `json:"striming"`
}

func parseSearchNodeConfig(raw string) *search.ModelConfig {
	if raw == "" {
		return nil
	}
	var meta searchNodeConfig
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return nil
	}
	if meta.Provider == "" || meta.Model == "" || meta.BaseURL == "" || meta.APIKey == "" {
		return nil
	}
	streaming := meta.Striming
	// Default to streaming=true for Responses API (Apodex deep research requires SSE)
	if !streaming {
		streaming = strings.Contains(strings.ToLower(meta.BaseURL), "/responses")
	}
	return &search.ModelConfig{
		Provider:  meta.Provider,
		Model:     meta.Model,
		BaseURL:   meta.BaseURL,
		APIKey:    meta.APIKey,
		Streaming: streaming,
	}
}

func dedupeSearchQueries(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, q := range in {
		key := strings.ToLower(q)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, q)
	}
	return out
}
