package search

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

// Базовые поисковые алгоритмы из issue (PageRank, BM25, нейросетевые модели
// Яндекса) применяются разными поисковиками. Внутри Octra мы используем BM25 —
// классический алгоритм оценки релевантности текста: насколько хорошо документ
// соответствует запросу с учётом частоты слов и длины документа. Он не требует
// обучения и сети, поэтому идеально подходит для локального переранжирования
// результатов, полученных от веб-провайдера.

// Параметры BM25 со стандартными значениями из литературы.
const (
	bm25K1 = 1.5  // насыщение частотой терма
	bm25B  = 0.75 // влияние длины документа
)

var tokenRe = regexp.MustCompile(`[\p{L}\p{N}]+`)

// tokenize — простая токенизация: слова/числа в нижнем регистре.
// Поддерживает любые алфавиты (Unicode), так как пользователи Octra пишут
// запросы и на русском, и на английском.
func tokenize(s string) []string {
	matches := tokenRe.FindAllString(strings.ToLower(s), -1)
	return matches
}

// RankBM25 переранжирует результаты по релевантности запросу query и возвращает
// новый отсортированный по убыванию срез. Поле Score каждого результата
// заполняется вычисленным значением BM25. Исходный срез не мутируется по порядку,
// но элементы копируются по значению (Result — простая структура).
func RankBM25(query string, results []Result) []Result {
	if len(results) == 0 {
		return nil
	}

	queryTerms := tokenize(query)

	// Токенизируем каждый документ и считаем длины.
	docTokens := make([][]string, len(results))
	totalLen := 0
	for i, r := range results {
		docTokens[i] = tokenize(r.text())
		totalLen += len(docTokens[i])
	}
	avgLen := float64(totalLen) / float64(len(results))
	if avgLen == 0 {
		avgLen = 1
	}

	// Документная частота (df) каждого терма запроса.
	df := make(map[string]int, len(queryTerms))
	for _, term := range uniqueStrings(queryTerms) {
		for _, tokens := range docTokens {
			if containsToken(tokens, term) {
				df[term]++
			}
		}
	}

	n := float64(len(results))
	ranked := make([]Result, len(results))
	copy(ranked, results)

	for i := range ranked {
		tf := termFrequencies(docTokens[i])
		docLen := float64(len(docTokens[i]))
		var score float64
		for _, term := range uniqueStrings(queryTerms) {
			f := float64(tf[term])
			if f == 0 {
				continue
			}
			// idf по формуле BM25 (со сглаживанием, всегда > 0).
			idf := math.Log(1 + (n-float64(df[term])+0.5)/(float64(df[term])+0.5))
			numerator := f * (bm25K1 + 1)
			denominator := f + bm25K1*(1-bm25B+bm25B*docLen/avgLen)
			score += idf * numerator / denominator
		}
		ranked[i].Score = score
	}

	// Стабильная сортировка по убыванию score; при равенстве сохраняем исходный
	// порядок провайдера (он обычно уже отражает авторитетность, ср. PageRank).
	sort.SliceStable(ranked, func(a, b int) bool {
		return ranked[a].Score > ranked[b].Score
	})
	return ranked
}

func termFrequencies(tokens []string) map[string]int {
	tf := make(map[string]int, len(tokens))
	for _, t := range tokens {
		tf[t]++
	}
	return tf
}

func containsToken(tokens []string, term string) bool {
	for _, t := range tokens {
		if t == term {
			return true
		}
	}
	return false
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
