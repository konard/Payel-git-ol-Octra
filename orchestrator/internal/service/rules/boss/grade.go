package boss

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// gradeTaskViaAI оценивает сложность задачи через AI (1-10).
// Использует того же провайдера/модель, что и пользователь.
// При ошибке возвращает grade=3 (базовый уровень).
func (s *Service) gradeTaskViaAI(ctx context.Context, provider, model string, tokens map[string]string, title, description string) int {
	taskText := title + "\n" + description
	if len(taskText) > 500 {
		taskText = taskText[:500]
	}

	prompt := fmt.Sprintf(`Rate the complexity of this task from 1 to 10.
1 = absolutely trivial (hello world, one function, minimal script)
10 = extremely complex (distributed system, ML pipeline, 10+ microservices)

Task: %s

Return ONLY a single number from 1 to 10. No explanation, no formatting.`, taskText)

	resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 64, 0.1)
	if err != nil {
		log.Printf("AI grading failed: %v, using default grade=3", err)
		return 3
	}

	resp = strings.TrimSpace(resp)
	grade, err := strconv.Atoi(resp)
	if err != nil || grade < 1 || grade > 10 {
		log.Printf("AI grading returned invalid grade: %q, using default grade=3", resp)
		return 3
	}

	log.Printf("AI graded task as %d/10", grade)
	return grade
}
