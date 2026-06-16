package worker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBoundedBufferKeepsTail проверяет, что boundedBuffer удерживает не более
// limit байт и сохраняет именно хвост вывода (где обычно лежит причина ошибки).
func TestBoundedBufferKeepsTail(t *testing.T) {
	b := newBoundedBuffer(16)
	for i := 0; i < 1000; i++ {
		b.WriteString("0123456789")
	}
	out := b.String()

	// Маркер усечения + не больше limit байт самих данных.
	if !strings.HasPrefix(out, "[...output truncated...]\n") {
		t.Fatalf("expected truncation marker, got: %q", out)
	}
	data := strings.TrimPrefix(out, "[...output truncated...]\n")
	if len(data) > 16 {
		t.Fatalf("expected at most 16 bytes retained, got %d", len(data))
	}
	// Хвост — это последние записанные символы.
	if !strings.HasSuffix("0123456789", data[len(data)-1:]) {
		t.Fatalf("retained data should be the tail, got %q", data)
	}
}

// TestBoundedBufferNoTruncationWhenSmall — небольшой вывод не усекается и не
// помечается маркером.
func TestBoundedBufferNoTruncationWhenSmall(t *testing.T) {
	b := newBoundedBuffer(1024)
	b.WriteString("hello\n")
	b.WriteString("world\n")
	if got := b.String(); got != "hello\nworld\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

// TestReadProjectFilesSkipsIgnoredDirs — fallback-чтение проекта не должно
// затягивать в память node_modules и прочие тяжёлые артефакты (issue #89).
func TestReadProjectFilesSkipsIgnoredDirs(t *testing.T) {
	dir := t.TempDir()

	write := func(rel, content string) {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	write("main.go", "package main")
	write("node_modules/left-pad/index.js", strings.Repeat("x", 10000))
	write("dist/bundle.js", strings.Repeat("y", 10000))
	write(".git/config", "gitstuff")

	files := readProjectFiles(dir)

	if _, ok := files["main.go"]; !ok {
		t.Fatalf("expected main.go to be read, got keys: %v", keys(files))
	}
	for path := range files {
		if strings.HasPrefix(path, "node_modules/") ||
			strings.HasPrefix(path, "dist/") ||
			strings.HasPrefix(path, ".git/") {
			t.Fatalf("ignored path should not be read into memory: %s", path)
		}
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
