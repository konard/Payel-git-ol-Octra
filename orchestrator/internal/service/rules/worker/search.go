package worker

import (
	"context"
	"log"
	"strconv"
	"strings"

	"orchestrator/internal/service/document"
	"orchestrator/internal/service/rules"
	"orchestrator/internal/service/search"
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
//
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

// gatherSearch выполняет реальный веб-поиск для документ-воркера в рамках
// диапазона или визуального направления (angle), заданного менеджером, и возвращает:
//   - block: отформатированные результаты для вставки в LLM-промпт,
//   - sources: Markdown-список источников для файла solution/sources-*.md,
//   - n: число найденных источников.
//
// Поиск устойчив к сбоям: при ошибке сети или выключенном поиске воркер просто
// получает пустой block и продолжает работу на собственных знаниях LLM.
func (s *Service) gatherSearch(ctx context.Context, emit searchEmitter, role, topic, angle string, searchConfig *search.ModelConfig) (block, sources string, n int) {
	if searchConfig == nil && (s.searchClient == nil || !search.Enabled()) {
		return "", "", 0
	}

	queries := buildSearchQueries(topic, angle)
	if len(queries) == 0 {
		return "", "", 0
	}

	for _, q := range queries {
		emit.emit("Searching the web for «"+q+"»", "searching", 0)
	}

	results, err := s.research(ctx, topic, queries, searchConfig)
	if err != nil {
		log.Printf("[Worker] search failed (%s): %v", role, err)
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

func (s *Service) research(ctx context.Context, topic string, queries []string, searchConfig *search.ModelConfig) ([]search.Result, error) {
	var merged []search.Result
	seen := make(map[string]struct{})
	var firstErr error

	addResults := func(results []search.Result) {
		for _, result := range results {
			key := strings.ToLower(strings.TrimRight(strings.TrimSpace(result.URL), "/"))
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, result)
		}
	}

	if searchConfig != nil {
		modelResults, err := search.NewClientWithProvider(search.NewModelProvider(*searchConfig)).Research(ctx, topic, queries, 2, 4)
		if err != nil {
			firstErr = err
			log.Printf("[Worker] AI search provider failed: %v", err)
		} else {
			addResults(modelResults)
		}
	}

	if s.searchClient != nil && search.Enabled() {
		webResults, err := s.searchClient.Research(ctx, topic, queries, 5, 8)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		} else {
			addResults(webResults)
		}
	}

	if len(merged) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, nil
	}

	ranked := search.RankBM25(topic, merged)
	if len(ranked) > 8 {
		ranked = ranked[:8]
	}
	return ranked, nil
}

// attachImages ищет в интернете и встраивает реальные картинки в слайды колоды.
// Для каждого слайда формируется запрос по его заголовку, теме и визуальному
// направлению; берётся первая успешно скачанная картинка поддерживаемого формата
// (jpeg/png). Если ИИ уже указал прямой URL картинки в Markdown, он пробуется
// первым. Шаг устойчив к сбоям: при выключенном или недоступном поиске колода
// просто остаётся без картинок, и презентация всё равно собирается.
func (s *Service) attachImages(ctx context.Context, deck *document.Deck, topic string) int {
	if s == nil || s.searchClient == nil || !search.Enabled() || deck == nil || len(deck.Slides) == 0 {
		return 0
	}

	// Ограничиваем число картинок, чтобы файл не раздувался, а поиск не затягивался.
	const maxImages = 6

	used := make(map[string]bool)
	attached := 0
	for i := range deck.Slides {
		if attached >= maxImages {
			break
		}
		sl := &deck.Slides[i]
		if sl.Image.Embeddable() {
			// Уже есть готовые байты (например, после предыдущего шага) — не трогаем.
			attached++
			continue
		}

		queries := buildImageQueries(sl.Title, topic, sl.Visual)
		candidates := s.searchClient.SearchImages(ctx, queries, 6)
		// Прямой URL от ИИ пробуем первым: он мог прийти из результатов веб-поиска.
		if u := strings.TrimSpace(sl.Image.URL); u != "" {
			candidates = append([]search.Image{{URL: u, Title: sl.Image.Alt}}, candidates...)
		}

		for _, cand := range candidates {
			key := normalizeImageKey(cand.URL)
			if key == "" || used[key] {
				continue
			}
			data, err := s.searchClient.FetchImage(ctx, cand)
			if err != nil {
				continue
			}
			used[key] = true
			sl.Image = document.Image{
				URL:         data.URL,
				Alt:         data.Alt,
				Source:      imageAttribution(cand),
				Data:        data.Data,
				ContentType: data.ContentType,
			}
			attached++
			break
		}
	}

	if attached > 0 {
		log.Printf("[Worker] attached %d web images to presentation", attached)
	}
	return attached
}

// buildImageQueries строит набор запросов на поиск картинки для слайда, комбинируя
// заголовок слайда, тему презентации и текстовое визуальное направление.
func buildImageQueries(slideTitle, topic, visual string) []string {
	var qs []string
	add := func(s string) {
		if s = cleanQuery(s); s != "" {
			qs = append(qs, s)
		}
	}
	title := cleanQuery(slideTitle)
	t := cleanQuery(topic)
	if title != "" && t != "" {
		add(title + " " + t)
	}
	add(title)
	add(t)
	add(visual)
	return dedupeQueries(qs)
}

// imageAttribution собирает короткую подпись-атрибуцию из источника и лицензии.
func imageAttribution(img search.Image) string {
	var parts []string
	if img.Source != "" {
		parts = append(parts, img.Source)
	}
	if img.License != "" {
		parts = append(parts, strings.ToUpper(img.License))
	}
	return strings.Join(parts, " · ")
}

func normalizeImageKey(u string) string {
	return strings.ToLower(strings.TrimRight(strings.TrimSpace(u), "/"))
}

// buildSearchQueries формирует набор поисковых запросов из темы и диапазона
// или визуального направления (angle) воркера. Тема даёт базовый запрос,
// а ключевые слова диапазона — уточняющий.
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
