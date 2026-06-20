package search

import (
	"log"
	"strings"
)

// Node описывает «ноду поиска» — отдельный узел воркфлоу, отвечающий за веб-поиск
// (issue #97). Нода может быть размещена пользователем на канве (Explicit) или
// создана системой автоматически, когда задача требует поиска информации, а явной
// ноды на канве нет (Created). Если задача поиск не требует, нода не активируется
// (Enabled == false), даже если пользователь разместил её на канве.
type Node struct {
	Enabled  bool   // нужно ли выполнять веб-поиск для этой задачи
	Explicit bool   // пользователь явно разместил ноду поиска на канве
	Created  bool   // ноду создали автоматически (Enabled && !Explicit)
	Provider string // AI-провайдер ноды поиска (для синтеза/уточнения)
	Model    string // AI-модель, указанная в ноде поиска
	Reason   string // почему поиск включён/выключен — для логов и UI
	AttachTo string // роль ноды, к которой прикреплён поиск (boss/manager/worker/universal)
}

// PlanNode решает, нужна ли ноде поиска работа для конкретной задачи, и с какой
// моделью её выполнять.
//
//   - text          — текст задачи (title + description), по которому оценивается
//     необходимость поиска;
//   - explicit      — на канве размещена нода поиска;
//   - nodeModel/nodeProvider — модель/провайдер, указанные в ноде поиска (могут быть пустыми);
//   - taskModel/taskProvider — модель/провайдер задачи по умолчанию (fallback).
//
// Правило (из issue): «обязательно производить поиск конкретной информации, если
// нужно; если нода размещена, но поиск не нужен — не обращаться к ней; если поиск
// нужен, а ноды нет — создаём её и прикрепляем».
func PlanNode(text string, explicit bool, nodeModel, nodeProvider, taskModel, taskProvider, attachTo string) Node {
	needs, reason := NeedsWebSearch(text)

	model := strings.TrimSpace(nodeModel)
	if model == "" {
		model = strings.TrimSpace(taskModel)
	}
	provider := strings.TrimSpace(nodeProvider)
	if provider == "" {
		provider = strings.TrimSpace(taskProvider)
	}

	node := Node{
		Enabled:  needs,
		Explicit: explicit,
		Created:  needs && !explicit,
		Provider: provider,
		Model:    model,
		Reason:   reason,
		AttachTo: strings.TrimSpace(attachTo),
	}

	switch {
	case !needs && explicit:
		log.Printf("[Search] node present on canvas but task needs no web search — skipping (attach=%s)", node.AttachTo)
	case needs && explicit:
		log.Printf("[Search] node on canvas activated: model=%s reason=%q attach=%s", node.Model, reason, node.AttachTo)
	case needs && !explicit:
		log.Printf("[Search] task needs web search — auto-creating a search node: model=%s reason=%q attach=%s", node.Model, reason, node.AttachTo)
	default:
		log.Printf("[Search] task needs no web search — no search node")
	}

	return node
}

// recencySignals — слова, указывающие, что нужен свежий/актуальный факт, который
// модель не может знать из обучающих данных.
var recencySignals = []string{
	// English
	"latest", "current", "currently", "today", "tonight", "right now", "as of",
	"recent", "recently", "this year", "this week", "this month", "nowadays",
	"up to date", "up-to-date", "newest", "breaking", "news",
	// Russian
	"последн", "текущ", "сегодня", "сейчас", "недавн", "актуальн", "в этом году",
	"на данный момент", "новост", "свеж",
}

// lookupSignals — слова/паттерны, указывающие на запрос внешнего факта (а не на
// математику/логику, которые универсальная нода решает сама).
var lookupSignals = []string{
	// English
	"weather", "forecast", "price", "cost of", "exchange rate", "stock",
	"who is", "who won", "release date", "released", "population of",
	"how much does", "how much is", "schedule", "score", "results of",
	"capital of", "ceo of", "version of", "documentation for",
	// Russian
	"погод", "прогноз", "цена", "стоимость", "курс ", "котировк", "акци",
	"кто так", "кто выиграл", "дата выход", "вышел", "вышла", "населен",
	"сколько стоит", "расписан", "счёт ", "счет ", "столица", "версия ",
	"какой", "какая", "какое", "какие", "какого", "который", "которого",
	"tiobe", "рейтинг", "топ", "место в",
}

// explicitSearchSignals — пользователь прямо просит найти/поискать информацию.
var explicitSearchSignals = []string{
	// English
	"search for", "search the web", "look up", "look it up", "google",
	"find information", "find info", "find out", "research ", "find the latest",
	"web search", "browse the web",
	// Russian
	"найди", "найти информаци", "поищи", "поиск ", "загугли", "узнай",
	"посмотри в интернет", "найти в интернет", "найди информаци",
}

// NeedsWebSearch эвристически решает, требует ли задача веб-поиска. Возвращает
// флаг и краткую причину (для логов/UI). Эвристика детерминирована и не ходит в
// сеть — в духе остального пакета search.
//
// Поиск НЕ нужен для самодостаточных задач: чистой математики/логики и явных
// запросов на написание кода — их решает универсальная нода или конвейер сам.
func NeedsWebSearch(text string) (bool, string) {
	t := strings.ToLower(strings.TrimSpace(text))
	if t == "" {
		return false, "empty task"
	}

	// Явная просьба найти информацию — всегда поиск.
	if hit := firstMatch(t, explicitSearchSignals); hit != "" {
		return true, "explicit search request: " + hit
	}

	// Запрос на написание кода/скрипта — самодостаточен, поиск не нужен.
	if isCodeRequest(t) {
		return false, "code/implementation task — self-contained"
	}

	// Чистое арифметическое/логическое выражение — решается без поиска.
	if isSelfContainedMath(t) {
		return false, "math/logic question — self-contained"
	}

	if hit := firstMatch(t, recencySignals); hit != "" {
		return true, "recency signal: " + hit
	}
	if hit := firstMatch(t, lookupSignals); hit != "" {
		return true, "factual-lookup signal: " + hit
	}

	// Год вида 20xx в тексте обычно означает запрос конкретного периода.
	if containsRecentYear(t) {
		return true, "explicit year reference"
	}

	return false, "no external-information signal"
}

func firstMatch(text string, signals []string) string {
	for _, s := range signals {
		if strings.Contains(text, s) {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

// isCodeRequest — повторяет сигналы looksLikeDirectAnswer боса: явная просьба
// написать код самодостаточна и поиска не требует.
func isCodeRequest(t string) bool {
	codeSignals := []string{
		"hello world", "write a script", "write a program", "write a function",
		"implement", "fix the bug", "fix bug", "write code", "code for",
		"function that", "class that", "напиши", "скрипт",
		"напиши программ", "напиши код", "создай программ",
		"функци", "реализ", "создай код", "сделай скрипт",
	}
	return firstMatch(t, codeSignals) != ""
}

// isSelfContainedMath — задача выглядит как чистая арифметика/логика без внешних
// фактов: содержит арифметические операторы и состоит в основном из цифр/символов.
func isSelfContainedMath(t string) bool {
	if !strings.ContainsAny(t, "+-*/=") {
		return false
	}
	letters := 0
	for _, r := range t {
		if (r >= 'a' && r <= 'z') || (r >= 'а' && r <= 'я') {
			letters++
		}
	}
	// Короткая строка с операторами и почти без букв — это выражение, а не вопрос.
	return letters <= 4
}

// containsRecentYear ищет упоминание года в диапазоне 2000–2099.
func containsRecentYear(t string) bool {
	for i := 0; i+4 <= len(t); i++ {
		if t[i] != '2' || t[i+1] != '0' {
			continue
		}
		c2, c3 := t[i+2], t[i+3]
		if c2 < '0' || c2 > '9' || c3 < '0' || c3 > '9' {
			continue
		}
		// Граница слова, чтобы не цеплять, например, "2000000".
		if i > 0 && isDigit(t[i-1]) {
			continue
		}
		if i+4 < len(t) && isDigit(t[i+4]) {
			continue
		}
		return true
	}
	return false
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }
