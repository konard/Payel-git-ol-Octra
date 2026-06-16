package memory

import (
	"runtime"
	"runtime/debug"
	"testing"
)

// TestReleaseToOS — освобождение памяти не должно паниковать и должно реально
// уменьшать (или хотя бы не увеличивать) объём кучи после отбрасывания крупной
// аллокации.
func TestReleaseToOS(t *testing.T) {
	// Создаём и отбрасываем крупную аллокацию (~64 МБ).
	big := make([]byte, 64*1024*1024)
	for i := range big {
		big[i] = byte(i)
	}
	_ = big[len(big)-1]
	big = nil

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	ReleaseToOS()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	if after.HeapAlloc > before.HeapAlloc {
		t.Fatalf("heap grew after release: %d -> %d", before.HeapAlloc, after.HeapAlloc)
	}
}

// TestConfigureRespectsExistingLimit — Configure не трогает лимит, выставленный
// через GOMEMLIMIT окружения (имитируем уже выставленным лимитом).
func TestConfigureRespectsExistingLimit(t *testing.T) {
	orig := debug.SetMemoryLimit(-1) // прочитать текущий, не меняя
	t.Cleanup(func() { debug.SetMemoryLimit(orig) })

	t.Setenv("GOMEMLIMIT", "512MiB")
	t.Setenv("ORCHESTRATOR_MEMORY_LIMIT_MIB", "128")

	Configure()

	if got := debug.SetMemoryLimit(-1); got != orig {
		t.Fatalf("Configure changed memory limit despite GOMEMLIMIT set: %d -> %d", orig, got)
	}
}

// TestConfigureAppliesCustomLimit — при заданном ORCHESTRATOR_MEMORY_LIMIT_MIB
// (и без GOMEMLIMIT) выставляется мягкий лимит памяти.
func TestConfigureAppliesCustomLimit(t *testing.T) {
	orig := debug.SetMemoryLimit(-1)
	t.Cleanup(func() { debug.SetMemoryLimit(orig) })

	t.Setenv("GOMEMLIMIT", "")
	t.Setenv("ORCHESTRATOR_MEMORY_LIMIT_MIB", "256")

	Configure()

	want := int64(256) * 1024 * 1024
	if got := debug.SetMemoryLimit(-1); got != want {
		t.Fatalf("expected memory limit %d, got %d", want, got)
	}
}
