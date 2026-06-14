package boss

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCopyFileSkipsNixStoreSymlink проверяет, что симлинк в /nix/store
// (например, `result` от `nix build`) пропускается, а не копируется по ссылке.
// Раньше это валило снапшот с `copy_file_range: is a directory` (issue #85).
func TestCopyFileSkipsNixStoreSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "result")
	if err := os.Symlink("/nix/store/abc-some-build", link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	dst := filepath.Join(dir, "out", "result")
	skipped, err := copyFile(link, dst)
	if err != nil {
		t.Fatalf("copyFile returned error: %v", err)
	}
	if !skipped {
		t.Fatalf("expected nix-store symlink to be skipped")
	}
	if _, err := os.Lstat(dst); !os.IsNotExist(err) {
		t.Fatalf("expected dst to not exist, got err=%v", err)
	}
}

// TestCopyFileSkipsDirectorySymlink проверяет, что симлинк на директорию
// пропускается (а не приводит к ошибке "is a directory").
func TestCopyFileSkipsDirectorySymlink(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "build-output")
	if err := os.Mkdir(realDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	link := filepath.Join(dir, "result")
	if err := os.Symlink(realDir, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	skipped, err := copyFile(link, filepath.Join(dir, "out", "result"))
	if err != nil {
		t.Fatalf("copyFile returned error: %v", err)
	}
	if !skipped {
		t.Fatalf("expected directory symlink to be skipped")
	}
}

// TestCopyFileCopiesRegularFile проверяет, что обычные файлы копируются как раньше.
func TestCopyFileCopiesRegularFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	if err := os.WriteFile(src, []byte("package main"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	dst := filepath.Join(dir, "out", "main.go")
	skipped, err := copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile error: %v", err)
	}
	if skipped {
		t.Fatalf("expected regular file to be copied, not skipped")
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "package main" {
		t.Fatalf("content mismatch: %q", string(got))
	}
}

// TestPrepareSnapshotDirIgnoresNixResult — интеграционный тест: проект с git и
// `result` симлинком успешно снапшотится (не падает на симлинке в /nix/store).
func TestPrepareSnapshotDirIgnoresNixResult(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v - %s", args, err, string(out))
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Симулируем артефакт `nix build`: симлинк на директорию в /nix/store.
	if err := os.Symlink("/nix/store/xxxx-build", filepath.Join(dir, "result")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	run("add", "main.go")
	run("commit", "-m", "init")

	staging, err := prepareSnapshotDir(dir)
	if err != nil {
		t.Fatalf("prepareSnapshotDir failed: %v", err)
	}
	defer os.RemoveAll(staging)

	if _, err := os.Stat(filepath.Join(staging, "main.go")); err != nil {
		t.Fatalf("expected main.go in staging: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(staging, "result")); !os.IsNotExist(err) {
		t.Fatalf("expected `result` symlink to be excluded from snapshot, err=%v", err)
	}
}
