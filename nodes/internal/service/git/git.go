package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// BranchStats summarizes the diff between a base branch and a head branch so the
// UI can show a pull request overview (commits / additions / deletions / files)
// without sending the user back to GitHub.
type BranchStats struct {
	Commits      int
	Additions    int
	Deletions    int
	ChangedFiles []string
}

// DiffStats computes commit and line statistics for head relative to base.
// Errors are reported to the caller, which is expected to degrade gracefully
// (an empty overview rather than failing the whole task).
func DiffStats(dir, base, head string) (BranchStats, error) {
	stats := BranchStats{}
	if base == "" || head == "" {
		return stats, fmt.Errorf("base and head are required")
	}
	revRange := base + ".." + head

	commitsCmd := exec.Command("git", "rev-list", "--count", revRange)
	commitsCmd.Dir = dir
	if out, err := commitsCmd.Output(); err == nil {
		if n, convErr := strconv.Atoi(strings.TrimSpace(string(out))); convErr == nil {
			stats.Commits = n
		}
	}

	numstatCmd := exec.Command("git", "diff", "--numstat", base+"..."+head)
	numstatCmd.Dir = dir
	out, err := numstatCmd.Output()
	if err != nil {
		return stats, fmt.Errorf("git diff --numstat failed: %w", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.SplitN(line, "\t", 3)
		if len(fields) < 3 {
			continue
		}
		// Binary files report "-" instead of a count; treat those as zero lines.
		if n, convErr := strconv.Atoi(fields[0]); convErr == nil {
			stats.Additions += n
		}
		if n, convErr := strconv.Atoi(fields[1]); convErr == nil {
			stats.Deletions += n
		}
		stats.ChangedFiles = append(stats.ChangedFiles, fields[2])
	}
	return stats, nil
}

// IsGitRepo checks whether the directory is a valid git repository.
func IsGitRepo(dir string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		if _, statErr := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(statErr) {
			return false, nil
		}
		return false, fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(output)) == "true", nil
}

