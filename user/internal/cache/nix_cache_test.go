package cache

import (
	"testing"
)

func TestNixCache_SetAndGet(t *testing.T) {
	c := NewNixCache(1024*1024, 10)

	files := []CodeFileEntry{
		{Path: "main.go", Content: "package main", Language: "go"},
	}

	c.Set("task1", files)

	got, ok := c.Get("task1")
	if !ok {
		t.Fatal("expected to find task1")
	}
	if len(got) != 1 || got[0].Path != "main.go" {
		t.Fatalf("unexpected files: %+v", got)
	}
}

func TestNixCache_Eviction(t *testing.T) {
	c := NewNixCache(100, 2) // 100 bytes, 2 items max

	c.Set("task1", []CodeFileEntry{{Path: "a.go", Content: "aaaa"}})
	c.Set("task2", []CodeFileEntry{{Path: "b.go", Content: "bbbb"}})
	c.Set("task3", []CodeFileEntry{{Path: "c.go", Content: "cccc"}})

	if c.Len() != 2 {
		t.Fatalf("expected 2 items, got %d", c.Len())
	}

	// task1 should be evicted
	if _, ok := c.Get("task1"); ok {
		t.Fatal("task1 should have been evicted")
	}

	// task2 and task3 should still be there
	if _, ok := c.Get("task2"); !ok {
		t.Fatal("task2 should still be cached")
	}
	if _, ok := c.Get("task3"); !ok {
		t.Fatal("task3 should still be cached")
	}
}

func TestNixCache_LRUEviction(t *testing.T) {
	c := NewNixCache(1024*1024, 2)

	c.Set("task1", []CodeFileEntry{{Path: "a.go", Content: "aaaa"}})
	c.Set("task2", []CodeFileEntry{{Path: "b.go", Content: "bbbb"}})

	// Access task1 to make it most recent
	c.Get("task1")

	// Add task3 - should evict task2 (oldest)
	c.Set("task3", []CodeFileEntry{{Path: "c.go", Content: "cccc"}})

	if _, ok := c.Get("task1"); !ok {
		t.Fatal("task1 should still be cached (recently accessed)")
	}
	if _, ok := c.Get("task2"); ok {
		t.Fatal("task2 should have been evicted (oldest)")
	}
	if _, ok := c.Get("task3"); !ok {
		t.Fatal("task3 should be cached")
	}
}
