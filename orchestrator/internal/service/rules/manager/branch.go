package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"orchestrator/internal/service/git"
	"orchestrator/internal/service/rules"
	"orchestrator/pkg/models"
)

// switchToManagerBranch — создаёт ветку manager-{role} и переключается на неё
func switchToManagerBranch(repoPath, role string) error {
	branchName := fmt.Sprintf("manager-%s", role)
	if err := git.CreateBranch(repoPath, branchName); err != nil {
		return err
	}
	log.Printf("Switched to branch: %s", branchName)
	return nil
}

// mergeWorkerBranches — мерджит ветки worker-{role} обратно в текущую (manager)
func mergeWorkerBranches(repoPath string, results []*rules.WorkerResult) {
	for _, wr := range results {
		branchName := fmt.Sprintf("worker-%s", wr.Role)
		mergeMessage := fmt.Sprintf("Merge worker %s branch", wr.Role)
		if err := git.MergeBranch(repoPath, branchName, mergeMessage); err != nil {
			log.Printf("Warning: failed to merge worker branch %s: %v", branchName, err)
		} else {
			log.Printf("Merged worker branch: %s", branchName)
		}
	}
}

// updateContextFile — добавляет решение менеджера в .octra/context.json
func updateContextFile(projectPath, taskID, role string, workerRoles []models.WorkerRole) {
	if projectPath == "" {
		return
	}
	octraDir := filepath.Join(projectPath, ".octra")
	contextPath := filepath.Join(octraDir, "context.json")
	var contextData map[string]interface{}
	if data, err := os.ReadFile(contextPath); err == nil {
		json.Unmarshal(data, &contextData)
	} else {
		contextData = map[string]interface{}{
			"task_id": taskID,
			"history": []map[string]interface{}{},
		}
	}
	history, ok := contextData["history"].([]map[string]interface{})
	if !ok {
		history = []map[string]interface{}{}
	}
	history = append(history, map[string]interface{}{
		"type":      "manager_decision",
		"manager":   role,
		"workers":   workerRoles,
		"timestamp": time.Now().Unix(),
	})
	contextData["history"] = history
	contextJSON, _ := json.MarshalIndent(contextData, "", "  ")
	os.MkdirAll(octraDir, 0755)
	os.WriteFile(contextPath, contextJSON, 0644)
}

