package cli

import (
	"context"
	"sync"
	"time"
)

// Process is a running Ocawe HTTP server process.
type Process interface {
	// Alive reports whether the process is still running.
	Alive() bool
	// Kill terminates the process.
	Kill() error
	// PID returns the OS process id.
	PID() int
	// Port returns the HTTP port of the process.
	Port() int
}

// LaunchSpec carries everything needed to start Ocawe for a user.
type LaunchSpec struct {
	UserID  string
	EnvPath string
}

// Launcher starts a new Ocawe process for a user environment.
type Launcher interface {
	Launch(ctx context.Context, spec LaunchSpec) (Process, error)
}

// Manager owns the per-user Ocawe process registry.
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
		_ = proc.Kill()
		delete(m.procs, spec.UserID)
	}

	proc, err := m.launcher.Launch(ctx, spec)
	if err != nil {
		return nil, err
	}
	m.procs[spec.UserID] = proc
	if err := m.store.Save(ctx, spec.UserID, State{PID: proc.PID(), StartedAt: time.Now()}, m.ttl); err != nil {
		_ = err
	}
	return proc, nil
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
