package worker

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type streamedCodeFile struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	Language    string `json:"language"`
	WorkerRole  string `json:"worker_role"`
	ManagerRole string `json:"manager_role"`
	Status      string `json:"status"`
	UpdatedAt   int64  `json:"updated_at"`
}

func buildCodeFilesPayload(files map[string]string, workerRole, managerRole string) string {
	if len(files) == 0 {
		return ""
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		if strings.TrimSpace(path) != "" {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)

	now := time.Now().Unix()
	payload := make([]streamedCodeFile, 0, len(paths))
	for _, path := range paths {
		payload = append(payload, streamedCodeFile{
			Path:        path,
			Content:     files[path],
			Language:    languageForPath(path),
			WorkerRole:  workerRole,
			ManagerRole: managerRole,
			Status:      "streaming",
			UpdatedAt:   now,
		})
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func languageForPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c":
		return "c"
	case ".cc", ".cpp", ".cxx", ".hpp", ".hxx":
		return "cpp"
	case ".css":
		return "css"
	case ".html", ".htm":
		return "html"
	case ".json":
		return "json"
	case ".md", ".markdown":
		return "markdown"
	case ".yml", ".yaml":
		return "yaml"
	case ".sh", ".bash":
		return "shell"
	case ".sql":
		return "sql"
	default:
		return "plaintext"
	}
}
