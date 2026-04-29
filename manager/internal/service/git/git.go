package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InitRepo initializes a new Git repository in the specified directory
func InitRepo(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}
	return nil
}

// SetUser configures Git user name and email for the repository
func SetUser(dir, name, email string) error {
	// Set user name
	cmd := exec.Command("git", "config", "user.name", name)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git config user.name failed: %w", err)
	}

	// Set user email
	cmd = exec.Command("git", "config", "user.email", email)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git config user.email failed: %w", err)
	}

	return nil
}

// Add adds files to the Git staging area
// If no files specified, adds all files (".")
func Add(dir string, files ...string) error {
	if len(files) == 0 {
		files = []string{"."}
	}

	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	return nil
}

// Commit commits staged changes with the given message
func Commit(dir, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	return nil
}

// CreateBranch creates a new branch and switches to it
func CreateBranch(dir, branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout -b failed: %w", err)
	}
	return nil
}

// CheckoutBranch switches to the specified branch
func CheckoutBranch(dir, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}
	return nil
}

// GetCurrentBranch returns the name of the current branch
func GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// MergeBranch merges the specified branch into the current branch
func MergeBranch(dir, branchName, message string) error {
	cmd := exec.Command("git", "merge", branchName, "-m", message)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git merge failed: %w", err)
	}
	return nil
}

// InitialCommit creates an initial empty commit
func InitialCommit(dir, message string) error {
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", message)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("initial git commit failed: %w", err)
	}
	return nil
}

// Status returns the current Git status
func Status(dir string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}
	return string(output), nil
}

// HasChanges checks if there are any uncommitted changes
func HasChanges(dir string) (bool, error) {
	status, err := Status(dir)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(status) != "", nil
}

// SaveGitData saves the .git folder contents to a map (for DB storage)
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
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}
		result[relPath] = string(data)
		return nil
	})

	return result, err
}

// SaveGitDataToJSON converts git data to JSON string for DB storage
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

// RestoreGitData restores the .git folder from saved data
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
		if err := ioutil.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// RestoreGitDataFromJSON restores .git folder from JSON string
func RestoreGitDataFromJSON(projectPath string, jsonData string) error {
	var gitData map[string]string
	if err := json.Unmarshal([]byte(jsonData), &gitData); err != nil {
		return err
	}
	return RestoreGitData(projectPath, gitData)
}