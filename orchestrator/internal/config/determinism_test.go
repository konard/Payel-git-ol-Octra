package config

import (
	"os"
	"testing"
)

// TestTemperatureIsSingleLowValue documents the philosophy invariant: every role
// must share one low temperature for maximum determinism.
func TestTemperatureIsSingleLowValue(t *testing.T) {
	if Temperature != 0.2 {
		t.Fatalf("Temperature = %v, want 0.2 (single deterministic value for all roles)", Temperature)
	}
}

func TestFallbackChainStartsWithPrimary(t *testing.T) {
	chain := FallbackChain("custom", "model-x")
	if len(chain) == 0 {
		t.Fatal("chain is empty")
	}
	if chain[0] != (ProviderModel{Provider: "custom", Model: "model-x"}) {
		t.Fatalf("chain[0] = %+v, want primary {custom model-x}", chain[0])
	}
	// Followed by the deterministic fallbacks.
	want := []ProviderModel{
		{"custom", "model-x"},
		{"openai", "gpt-4o-mini"},
		{"anthropic", "claude-3-haiku-20240307"},
		{"gemini", "gemini-pro"},
	}
	if len(chain) != len(want) {
		t.Fatalf("chain len = %d, want %d (%+v)", len(chain), len(want), chain)
	}
	for i := range want {
		if chain[i] != want[i] {
			t.Errorf("chain[%d] = %+v, want %+v", i, chain[i], want[i])
		}
	}
}

func TestFallbackChainDeduplicatesPrimary(t *testing.T) {
	// When the primary already equals a fallback entry, it must not appear twice.
	chain := FallbackChain("openai", "gpt-4o-mini")
	seen := 0
	for _, p := range chain {
		if p == (ProviderModel{Provider: "openai", Model: "gpt-4o-mini"}) {
			seen++
		}
	}
	if seen != 1 {
		t.Fatalf("openai/gpt-4o-mini appears %d times, want exactly 1: %+v", seen, chain)
	}
}

func TestFallbackChainSkipsEmptyPrimary(t *testing.T) {
	chain := FallbackChain("", "")
	want := []ProviderModel{
		{"openai", "gpt-4o-mini"},
		{"anthropic", "claude-3-haiku-20240307"},
		{"gemini", "gemini-pro"},
	}
	if len(chain) != len(want) {
		t.Fatalf("chain len = %d, want %d (%+v)", len(chain), len(want), chain)
	}
	for i := range want {
		if chain[i] != want[i] {
			t.Errorf("chain[%d] = %+v, want %+v", i, chain[i], want[i])
		}
	}
}

func TestGenerationModeIsDeterministic(t *testing.T) {
	supported := func(stack string) bool { return stack == "Go" }

	if got := GenerationMode("Go", supported); got != ModeTool {
		t.Errorf("GenerationMode(Go) = %q, want %q", got, ModeTool)
	}
	if got := GenerationMode("Brainfuck", supported); got != ModeMultiPass {
		t.Errorf("GenerationMode(Brainfuck) = %q, want %q", got, ModeMultiPass)
	}
	// Same input → same output, every time.
	for i := 0; i < 5; i++ {
		if GenerationMode("Go", supported) != ModeTool {
			t.Fatal("GenerationMode is not deterministic")
		}
	}
}

func TestGenerationModeNilToolSupport(t *testing.T) {
	if got := GenerationMode("Go", nil); got != ModeMultiPass {
		t.Errorf("GenerationMode with nil support = %q, want %q", got, ModeMultiPass)
	}
}

func TestResolveGenerationModeEscapeHatch(t *testing.T) {
	supported := func(string) bool { return true }

	t.Setenv("OCTRA_DISABLE_TOOLS", "true")
	if got := ResolveGenerationMode("Go", supported); got != ModeMultiPass {
		t.Errorf("with OCTRA_DISABLE_TOOLS=true: got %q, want %q", got, ModeMultiPass)
	}

	os.Unsetenv("OCTRA_DISABLE_TOOLS")
	if got := ResolveGenerationMode("Go", supported); got != ModeTool {
		t.Errorf("without escape hatch: got %q, want %q", got, ModeTool)
	}
}

// TestUniversalNodeMaxGradeResolved documents the issue #91 fast-path threshold:
// trivial tasks (low model grade) are handled by a single universal node instead
// of the full pipeline. The threshold is operator-overridable and can be disabled.
func TestUniversalNodeMaxGradeResolved(t *testing.T) {
	t.Run("default threshold", func(t *testing.T) {
		os.Unsetenv("OCTRA_DISABLE_UNIVERSAL_NODE")
		os.Unsetenv("OCTRA_UNIVERSAL_MAX_GRADE")
		if got := UniversalNodeMaxGradeResolved(); got != UniversalNodeMaxGrade {
			t.Fatalf("UniversalNodeMaxGradeResolved() = %d, want %d", got, UniversalNodeMaxGrade)
		}
	})

	t.Run("disabled via env returns 0", func(t *testing.T) {
		t.Setenv("OCTRA_DISABLE_UNIVERSAL_NODE", "true")
		if got := UniversalNodeMaxGradeResolved(); got != 0 {
			t.Fatalf("disabled threshold = %d, want 0", got)
		}
	})

	t.Run("env override within range", func(t *testing.T) {
		os.Unsetenv("OCTRA_DISABLE_UNIVERSAL_NODE")
		t.Setenv("OCTRA_UNIVERSAL_MAX_GRADE", "4")
		if got := UniversalNodeMaxGradeResolved(); got != 4 {
			t.Fatalf("override threshold = %d, want 4", got)
		}
	})

	t.Run("invalid override falls back to default", func(t *testing.T) {
		os.Unsetenv("OCTRA_DISABLE_UNIVERSAL_NODE")
		t.Setenv("OCTRA_UNIVERSAL_MAX_GRADE", "99")
		if got := UniversalNodeMaxGradeResolved(); got != UniversalNodeMaxGrade {
			t.Fatalf("invalid override threshold = %d, want default %d", got, UniversalNodeMaxGrade)
		}
		t.Setenv("OCTRA_UNIVERSAL_MAX_GRADE", "notanumber")
		if got := UniversalNodeMaxGradeResolved(); got != UniversalNodeMaxGrade {
			t.Fatalf("non-numeric override threshold = %d, want default %d", got, UniversalNodeMaxGrade)
		}
	})
}
