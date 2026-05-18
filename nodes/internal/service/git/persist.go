package git

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SaveGitData — собирает содержимое .git в map для сохранения в БД
func SaveGitData(projectPath string) (map[string]string, error) {
	gitDir := filepath.Join(projectPath, ".git")
	result := make(map[string]string)

	err := filepath.Walk(gitDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		result[relPath] = string(data)
		return nil
	})

	return result, err
}

// SaveGitDataToJSON — сериализует .git в JSON-строку для хранения в БД
func SaveGitDataToJSON(projectPath string) (string, error) {
	data, err := SaveGitData(projectPath)
	if err != nil {
		return "", err
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// RestoreGitData — восстанавливает .git из map
func RestoreGitData(projectPath string, gitData map[string]string) error {
	gitDir := filepath.Join(projectPath, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		return err
	}
	for relPath, content := range gitData {
		fullPath := filepath.Join(projectPath, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// RestoreGitDataFromJSON — восстанавливает .git из JSON-строки
func RestoreGitDataFromJSON(projectPath string, jsonData string) error {
	var gitData map[string]string
	if err := json.Unmarshal([]byte(jsonData), &gitData); err != nil {
		return err
	}
	return RestoreGitData(projectPath, gitData)
}
