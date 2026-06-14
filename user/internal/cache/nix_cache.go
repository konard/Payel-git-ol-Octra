package cache

import (
	"sync"
	"time"
)

// CodeFileEntry represents a single code file for caching
type CodeFileEntry struct {
	Path     string
	Content  string
	Language string
	Encoding string
}

type cacheEntry struct {
	taskID   string
	files    []CodeFileEntry
	size     int64
	accessed time.Time
}

// NixCache is an LRU cache for project files from Nix store
type NixCache struct {
	mu       sync.RWMutex
	entries  map[string]*cacheEntry
	order    []string // LRU order: most recent at end
	maxSize  int64    // max total size in bytes (500MB)
	maxItems int      // max number of items (10)
}

// NewNixCache creates a new LRU cache
func NewNixCache(maxSizeBytes int64, maxItems int) *NixCache {
	return &NixCache{
		entries:  make(map[string]*cacheEntry),
		order:    make([]string, 0, maxItems),
		maxSize:  maxSizeBytes,
		maxItems: maxItems,
	}
}

// Get retrieves cached files for a task. Returns files and true if found.
func (c *NixCache) Get(taskID string) ([]CodeFileEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[taskID]
	if !ok {
		return nil, false
	}

	entry.accessed = time.Now()
	c.moveToEnd(taskID)
	return entry.files, true
}

// Set stores files for a task, evicting if necessary
func (c *NixCache) Set(taskID string, files []CodeFileEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate size
	var size int64
	for _, f := range files {
		size += int64(len(f.Content))
	}

	// Remove existing entry if re-setting
	if _, ok := c.entries[taskID]; ok {
		c.remove(taskID)
	}

	// Evict until we have space
	for len(c.entries) >= c.maxItems || (len(c.entries) > 0 && c.totalSize()+size > c.maxSize) {
		c.evictOldest()
	}

	c.entries[taskID] = &cacheEntry{
		taskID:   taskID,
		files:    files,
		size:     size,
		accessed: time.Now(),
	}
	c.order = append(c.order, taskID)
}

// totalSize returns the total size of all cached entries
func (c *NixCache) totalSize() int64 {
	var total int64
	for _, e := range c.entries {
		total += e.size
	}
	return total
}

// evictOldest removes the least recently used entry
func (c *NixCache) evictOldest() {
	if len(c.order) == 0 {
		return
	}
	oldest := c.order[0]
	c.remove(oldest)
}

// remove removes an entry from the cache
func (c *NixCache) remove(taskID string) {
	delete(c.entries, taskID)
	for i, id := range c.order {
		if id == taskID {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}

// moveToEnd moves an entry to the end of the LRU order
func (c *NixCache) moveToEnd(taskID string) {
	for i, id := range c.order {
		if id == taskID {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, taskID)
			return
		}
	}
}

// Len returns the number of items in the cache
func (c *NixCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
