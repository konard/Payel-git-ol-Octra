package worker

import (
	"context"
	"log"
	"strconv"
	"strings"

	"nodes/internal/service/rules"
	"nodes/internal/service/search"
)

// searchEmitter передаёт шаги веб-поиска вверх по цепочке (worker → manager →
// boss → apigateway → чат). Каждый вызов соответствует одному пункту в блоке
// «Searching the web» в интерфейсе. phase — "searching" пока идёт поиск и
// "done", когда воркер закончил; count — итоговое число выполненных шагов.
type searchEmitter func(step, phase string, count int)

func (e searchEmitter) emit(step, phase string, count int) {
	if e != nil {
		e(step, phase, count)
	}
}

// searchEmitterFor строит searchEmitter, который транслирует шаги веб-поиска вверх
// по цепочке через ProgressFunc. Каждый шаг отправляется с тем же basePct (чтобы
// шкала прогресса не «прыгала»), а сами данные шага едут в data-карте:
//   - search_step  — текст пункта («Searching the web for …»),
//   - search_phase — "searching" пока идёт поиск, "done" по завершении,
//   - search_steps_count — итоговое число шагов (для строки «Completed N steps»).
// Фронтенд читает эти ключи и рисует сворачиваемый блок «Searching the web».
// Если progress == nil (например, в тестах), возвращается nil-эмиттер, безопасный
// к вызовам через метод emit.
func (s *Service) searchEmitterFor(progress rules.ProgressFunc, basePct int32, req *rules.AssignWorkersRequest, role string) searchEmitter {
	if progress == nil {
		return nil
	}
	return func(step, phase string, count int) {
		data := map[string]string{
			"manager_id":   req.ManagerId,
			"manager_role": req.ManagerRole,
			"worker_role":  role,
			"search_phase": phase,
		}
		if step != "" {
			data["search_step"] = step
		}
		if count > 0 {
			data["search_steps_count"] = strconv.Itoa(count)
		}
		message := "Searching the web"
		if step != "" {
			message = step
		}
		progress(basePct, message, data)
	}
}

// gatherSearch выполняет реальный веб-поиск для ресёрч-воркера в рамках
// диапазона (angle), заданного менеджером, и возвращает:
//   - block: отформатированные результаты для вставки в LLM-промпт,
//   - sources: Markdown-список источников для файла solution/sources-*.md,
//   - n: число найденных источников.
//
// Поиск устойчив к сбоям: при ошибке сети или выключенном поиске воркер просто
// получает пустой block и продолжает работу на собственных знаниях LLM.
func (s *Service) gatherSearch(ctx context.Context, emit searchEmitter, role, topic, angle string) (block, sources string, n int) {
	if s.searchClient == nil || !search.Enabled() {
		return "", "", 0
	}

	queries := buildSearchQueries(topic, angle)
	if len(queries) == 0 {
		return "", "", 0
	}

	for _, q := range queries {
		emit.emit("Searching the web for «"+q+"»", "searching", 0)
	}

	results, err := s.searchClient.Research(ctx, topic, queries, 5, 8)
	if err != nil {
		log.Printf("[Worker] web search failed (%s): %v", role, err)
		emit.emit("", "done", len(queries))
		return "", "", 0
	}
	if len(results) == 0 {
		emit.emit("", "done", len(queries))
		return "", "", 0
	}

	emit.emit("", "done", len(queries))
	log.Printf("[Worker] web search (%s): %d queries → %d sources", role, len(queries), len(results))
	return search.FormatForPrompt(results), search.FormatSourcesMarkdown(results), len(results)
}

// buildSearchQueries формирует набор поисковых запросов из темы и диапазона
// (angle) воркера. Тема даёт базовый запрос, а ключевые слова диапазона —
// уточняющий, чтобы аналитики искали в разных «диапазонах», как требует issue.
func buildSearchQueries(topic, angle string) []string {
	base := cleanQuery(topic)
	if base == "" {
		return nil
	}

	queries := []string{base}

	if kw := angleKeywords(angle); kw != "" {
		combined := cleanQuery(base + " " + kw)
		if combined != "" && combined != base {
			queries = append(queries, combined)
		}
	}
	return dedupeQueries(queries)
}

// cleanQuery нормализует строку в компактный поисковый запрос: первая строка,
// без переводов строки, ограниченная по длине.
func cleanQuery(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 120 {
		// Обрезаем по границе слова, чтобы не рвать слова посередине.
		s = s[:120]
		if i := strings.LastIndexByte(s, ' '); i > 0 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

// angleKeywords извлекает короткие ключевые слова из описания диапазона,
// отбрасывая служебные слова, чтобы получить полезный уточняющий запрос.
func angleKeywords(angle string) string {
	angle = strings.ToLower(angle)
	stop := map[string]bool{
		"search": true, "via": true, "the": true, "and": true, "for": true, "of": true,
		"a": true, "an": true, "to": true, "with": true, "investigate": true, "find": true,
		"using": true, "from": true, "on": true, "in": true, "by": true, "sources": true,
		"information": true, "range": true, "look": true, "links": true, "pages": true,
	}
	words := strings.FieldsFunc(angle, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	kept := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) < 3 || stop[w] {
			continue
		}
		kept = append(kept, w)
		if len(kept) >= 4 {
			break
		}
	}
	return strings.Join(kept, " ")
}

func dedupeQueries(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, q := range in {
		if q == "" {
			continue
		}
		key := strings.ToLower(q)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, q)
	}
	return out
}
