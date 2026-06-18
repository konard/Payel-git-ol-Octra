// Package config centralizes the deterministic runtime parameters of the Octra
// orchestrator.
//
// Octra's philosophy treats the AI as a capable but unpredictable executor that
// must be tightly constrained: every role (Boss, Manager, Worker) has to run in a
// single, predictable mode so behavior cannot drift between steps. Spreading the
// same knobs (sampling temperature, the provider fallback chain, the code
// generation mode) across many files made the pipeline inconsistent — one part
// conservative, another creative — which is exactly the unpredictability the
// system is meant to eliminate.
//
// This package is the single source of truth for those knobs.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Temperature is the single sampling temperature shared by every role.
//
// Previously the roles used different temperatures (Boss 0.2, Manager 0.7,
// Worker 0.3). A creative Manager combined with a conservative Worker produced
// inconsistent, hard-to-reproduce runs. A single low value keeps the whole
// Boss → Manager → Worker pipeline maximally deterministic.
const Temperature float32 = 0.2

// ProviderModel identifies a concrete provider/model pair used when the primary
// AI provider is unavailable.
type ProviderModel struct {
	Provider string
	Model    string
}

// fallbackProviders is the deterministic, ordered chain tried after the
// caller-supplied provider fails. Keeping it here (instead of inline in the Boss)
// makes the chain a single, auditable source of truth that every node can share.
var fallbackProviders = []ProviderModel{
	{Provider: "openai", Model: "gpt-4o-mini"},
	{Provider: "anthropic", Model: "claude-3-haiku-20240307"},
	{Provider: "gemini", Model: "gemini-pro"},
}

// FallbackChain returns the ordered list of provider/model pairs to try for a
// task, starting with the caller's primary choice and followed by the shared
// deterministic fallbacks. The primary pair is never duplicated in the chain.
//
// Passing an empty provider or model skips the primary entry, so callers can rely
// on the fallback chain alone.
func FallbackChain(primaryProvider, primaryModel string) []ProviderModel {
	chain := make([]ProviderModel, 0, len(fallbackProviders)+1)
	seen := make(map[ProviderModel]bool, len(fallbackProviders)+1)

	add := func(p ProviderModel) {
		if p.Provider == "" || p.Model == "" || seen[p] {
			return
		}
		seen[p] = true
		chain = append(chain, p)
	}

	add(ProviderModel{Provider: primaryProvider, Model: primaryModel})
	for _, p := range fallbackProviders {
		add(p)
	}
	return chain
}

// UniversalNodeMaxGrade is the highest AI-graded complexity (1-10) that is still
// handled by the single "universal node" fast path before the full Boss →
// Manager → Worker fan-out starts.
//
// Octra's pipeline is built for divide-and-conquer on complex tasks, but the same
// machinery massively over-engineers trivial requests: "write hello world in
// Python" would spawn managers and workers and emit a tree of files with unclear
// logic instead of the one obvious line of code (issue #91). For tasks the model
// itself grades as trivial, a single universal node produces the minimal, correct
// answer immediately — no Boss planning, no extra nodes, no big workflow.
//
// The threshold is derived purely from the model's complexity analysis (not from
// trigger words), so a 1-2/10 task takes the fast path while anything heavier
// keeps the full pipeline.
const UniversalNodeMaxGrade = 2

// envUniversalNodeDisabled is an explicit escape hatch to force every task, even
// trivial ones, through the full Boss → Manager → Worker pipeline. It does not
// add a second behavior; it merely opts out of the fast path for environments
// that want the classic flow.
func envUniversalNodeDisabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("OCTRA_DISABLE_UNIVERSAL_NODE")), "true")
}

// UniversalNodeMaxGradeResolved returns the effective universal-node threshold.
// It honors the OCTRA_DISABLE_UNIVERSAL_NODE escape hatch (returns 0 → fast path
// never triggers) and the OCTRA_UNIVERSAL_MAX_GRADE override (1-10) for operators
// who want to widen or narrow the trivial band.
func UniversalNodeMaxGradeResolved() int {
	if envUniversalNodeDisabled() {
		return 0
	}
	if v := strings.TrimSpace(os.Getenv("OCTRA_UNIVERSAL_MAX_GRADE")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 10 {
			return n
		}
	}
	return UniversalNodeMaxGrade
}

// Generation modes for the Worker. There is exactly one path per task — chosen
// deterministically by the runtime, never by an environment flag.
const (
	// ModeTool runs the native toolchain scaffolder (npm init, cargo init, …)
	// inside the Nix dev shell for tech stacks that support it.
	ModeTool = "tool"
	// ModeMultiPass generates every file in a single LLM request. It is the
	// default for stacks without a native scaffolder and the fallback when tool
	// scaffolding produces nothing.
	ModeMultiPass = "multipass"
)

// GenerationMode returns the single deterministic code-generation mode for a
// tech stack. Tool scaffolding is preferred for stacks with a native toolchain;
// everything else uses multi-pass generation.
//
// The decision is derived purely from the task's tech stack — there is no
// WORKER_MODE override — so the same task always takes the same path.
func GenerationMode(techStack string, toolSupported func(string) bool) string {
	if toolSupported != nil && toolSupported(strings.TrimSpace(techStack)) {
		return ModeTool
	}
	return ModeMultiPass
}

// envToolDisabled is an explicit escape hatch kept only for environments where
// the Nix toolchain is unavailable (e.g. CI sandboxes). It does NOT introduce a
// second creative path — it merely forces the deterministic multi-pass fallback
// that tool mode would have used on failure anyway.
func envToolDisabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("OCTRA_DISABLE_TOOLS")), "true")
}

// ResolveGenerationMode is the runtime entry point used by workers. It applies
// the deterministic GenerationMode decision and honors the OCTRA_DISABLE_TOOLS
// escape hatch for tool-less environments.
func ResolveGenerationMode(techStack string, toolSupported func(string) bool) string {
	if envToolDisabled() {
		return ModeMultiPass
	}
	return GenerationMode(techStack, toolSupported)
}
