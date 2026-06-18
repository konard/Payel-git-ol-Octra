package boss

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"orchestrator/internal/config"
)

// gradeTaskViaAI оценивает сложность задачи через AI (1-10).
// Использует того же провайдера/модель, что и пользователь.
// При ошибке возвращает grade=3 (базовый уровень).
func (s *Service) gradeTaskViaAI(ctx context.Context, provider, model string, tokens map[string]string, title, description string) int {
	taskText := title + "\n" + description
	if len(taskText) > 500 {
		taskText = taskText[:500]
	}

	prompt := fmt.Sprintf(`Rate the REAL complexity of this task from 1 to 10.

Judge by the actual work the task requires — NOT by keywords or how the request is
phrased. A request can sound technical yet be trivial, or sound casual yet be hard.
Estimate how much a competent engineer would actually have to build.

1  = absolutely trivial: a single obvious answer or one tiny file (hello world,
     one short function, a math/logic question, a one-line script).
2  = very simple: one small self-contained file, no real architecture.
3-4 = a small program or focused script with a couple of pieces.
5-7 = a real application or service with multiple components.
8-10 = extremely complex (distributed system, ML pipeline, many microservices).

Task: %s

Return ONLY a single number from 1 to 10. No explanation, no formatting.`, taskText)

	resp, err := s.agentsClient.Generate(ctx, provider, model, prompt, tokens, 64, config.Temperature)
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
