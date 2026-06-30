// Package cli manages long-lived AI CLI subprocesses, one per user. Processes
// survive across requests: the manager keeps the in-process handle and mirrors
// liveness/TTL into a StateStore (Redis in production). When a process is
// missing or its TTL has expired it is (re)launched.
package cli

import (
	"context"
	"sync"
	"time"

	"backend/internal/model"
)

// Process is a running CLI subprocess that accepts prompts and returns replies
// over its stdin/stdout pipe.
type Process interface {
	// Send writes a prompt and reads the reply.
	Send(ctx context.Context, prompt string) (string, error)
	// Alive reports whether the process is still running.
	Alive() bool
	// Kill terminates the process.
	Kill() error
	// PID returns the OS process id.
	PID() int
	// Port returns the HTTP port of the process (0 for non-HTTP processes).
	Port() int
}

// LaunchSpec carries everything needed to start a CLI for a user.
type LaunchSpec struct {
	UserID  string
	EnvPath string
	CLI     model.CLIType
	LLM     LLMConfig
}

// LLMConfig is the LLM connection injected into the CLI's environment.
type LLMConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

// Launcher starts a new CLI process for a user environment.
type Launcher interface {
	Launch(ctx context.Context, spec LaunchSpec) (Process, error)
}

// Manager owns the per-user process registry.
type Manager struct {
	launcher Launcher
	store    StateStore
	ttl      time.Duration

	mu    sync.Mutex
	procs map[string]Process
}

// NewManager builds a Manager. ttl bounds how long an idle process is kept.
func NewManager(launcher Launcher, store StateStore, ttl time.Duration) *Manager {
	return &Manager{
		launcher: launcher,
		store:    store,
		ttl:      ttl,
		procs:    make(map[string]Process),
	}
}

// Send routes a prompt to the user's CLI, launching or relaunching it when the
// previous process is dead or its TTL has expired. The TTL is refreshed on
// every successful request.
func (m *Manager) Send(ctx context.Context, spec LaunchSpec, prompt string) (string, error) {
	proc, err := m.ensure(ctx, spec)
	if err != nil {
		return "", err
	}

	reply, err := proc.Send(ctx, prompt)
	if err != nil {
		// The process likely died mid-request; drop it so the next call
		// relaunches from a clean slate.
		m.drop(ctx, spec.UserID)
		return "", err
	}

	// Keep the process warm for another TTL window.
	_ = m.store.Save(ctx, spec.UserID, State{PID: proc.PID(), Port: proc.Port(), StartedAt: time.Now()}, m.ttl)
	return reply, nil
}

// EnsureOcawe returns the port of a running Ocawe HTTP server for the user,
// launching one if necessary.
func (m *Manager) EnsureOcawe(ctx context.Context, spec LaunchSpec) (int, error) {
	proc, err := m.ensure(ctx, spec)
	if err != nil {
		return 0, err
	}
	return proc.Port(), nil
}

// OcawePort returns the stored Ocawe port for the user from Redis, or 0 if
// no port is recorded.
func (m *Manager) OcawePort(ctx context.Context, userID string) (int, error) {
	st, err := m.store.Get(ctx, userID)
	if err != nil {
		return 0, err
	}
	return st.Port, nil
}

// ensure returns a live process for the user, launching one if necessary.
func (m *Manager) ensure(ctx context.Context, spec LaunchSpec) (Process, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if proc, ok := m.procs[spec.UserID]; ok && proc.Alive() {
		if alive, err := m.store.Alive(ctx, spec.UserID); err == nil && alive {
			return proc, nil
		}
		// TTL expired (or store says dead): force-kill the stale process.
		_ = proc.Kill()
		delete(m.procs, spec.UserID)
	}

	proc, err := m.launcher.Launch(ctx, spec)
	if err != nil {
		return nil, err
	}
	m.procs[spec.UserID] = proc
	if err := m.store.Save(ctx, spec.UserID, State{PID: proc.PID(), StartedAt: time.Now()}, m.ttl); err != nil {
		// Non-fatal: the process is usable even if we failed to record state.
		_ = err
	}
	return proc, nil
}

// drop kills and forgets the user's process.
func (m *Manager) drop(ctx context.Context, userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if proc, ok := m.procs[userID]; ok {
		_ = proc.Kill()
		delete(m.procs, userID)
	}
	_ = m.store.Delete(ctx, userID)
}

// Shutdown kills every tracked process. Used on graceful shutdown.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, proc := range m.procs {
		_ = proc.Kill()
		delete(m.procs, id)
	}
}
