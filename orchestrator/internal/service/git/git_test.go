package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// run executes a git command in dir and fails the test on error so the helper
// setup reads top-to-bottom without repetitive error handling.
func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// TestDiffStats verifies that DiffStats reports the commits, line additions and
// deletions, and changed file list of a feature branch relative to its base —
// the data the Solution pane uses to show a pull request overview (issue #44).
func TestDiffStats(t *testing.T) {
	dir := t.TempDir()

	if err := InitRepo(dir); err != nil {
		t.Fatalf("InitRepo: %v", err)
	}
	if err := SetUser(dir, "Tester", "tester@example.com"); err != nil {
		t.Fatalf("SetUser: %v", err)
	}
	// Some environments default to "master"; pin the base branch name so the
	// rev range is deterministic regardless of git's init.defaultBranch.
	run(t, dir, "checkout", "-b", "main")

	writeFile(t, dir, "keep.txt", "line1\nline2\nline3\n")
	if err := Add(dir); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := Commit(dir, "base commit"); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if err := CreateBranch(dir, "feature"); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	// First change: add a brand new file (3 additions).
	writeFile(t, dir, "added.txt", "a\nb\nc\n")
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-m", "add file")
	// Second change: drop a line from the existing file (1 addition, 1 deletion).
	writeFile(t, dir, "keep.txt", "line1\nline3\nline4\n")
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-m", "edit file")

	stats, err := DiffStats(dir, "main", "feature")
	if err != nil {
		t.Fatalf("DiffStats: %v", err)
	}

	if stats.Commits != 2 {
		t.Errorf("Commits = %d, want 2", stats.Commits)
	}
	// added.txt: +3, keep.txt: +1 -1 → additions 4, deletions 1.
	if stats.Additions != 4 {
		t.Errorf("Additions = %d, want 4", stats.Additions)
	}
	if stats.Deletions != 1 {
		t.Errorf("Deletions = %d, want 1", stats.Deletions)
	}
	wantFiles := map[string]bool{"added.txt": true, "keep.txt": true}
	if len(stats.ChangedFiles) != len(wantFiles) {
		t.Fatalf("ChangedFiles = %v, want %v", stats.ChangedFiles, wantFiles)
	}
	for _, f := range stats.ChangedFiles {
		if !wantFiles[f] {
			t.Errorf("unexpected changed file %q", f)
		}
	}
}

// TestDiffStatsRequiresBaseAndHead guards the input validation so a missing
// branch name degrades to an error instead of an empty/false overview.
func TestDiffStatsRequiresBaseAndHead(t *testing.T) {
	if _, err := DiffStats(t.TempDir(), "", "feature"); err == nil {
		t.Error("expected error when base is empty")
	}
	if _, err := DiffStats(t.TempDir(), "main", ""); err == nil {
		t.Error("expected error when head is empty")
	}
}

