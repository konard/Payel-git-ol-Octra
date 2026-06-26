package cli

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeProcess is an in-memory Process.
type fakeProcess struct {
	id      int
	alive   bool
	killed  int
	sendErr error
	calls   int
}

func (p *fakeProcess) Send(_ context.Context, prompt string) (string, error) {
	p.calls++
	if p.sendErr != nil {
		return "", p.sendErr
	}
	return "reply:" + prompt, nil
}
func (p *fakeProcess) Alive() bool { return p.alive }
func (p *fakeProcess) Kill() error { p.killed++; p.alive = false; return nil }
func (p *fakeProcess) PID() int    { return p.id }

// fakeLauncher hands out preconfigured processes and counts launches.
type fakeLauncher struct {
	launches int
	next     func(n int) Process
}

func (l *fakeLauncher) Launch(_ context.Context, _ LaunchSpec) (Process, error) {
	l.launches++
	return l.next(l.launches), nil
}

// memStore is an in-memory StateStore honouring TTL via a manual clock flag.
type memStore struct {
	mu    sync.Mutex
	alive map[string]bool
}

func newMemStore() *memStore { return &memStore{alive: map[string]bool{}} }
func (s *memStore) Alive(_ context.Context, id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alive[id], nil
}
func (s *memStore) Save(_ context.Context, id string, _ State, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alive[id] = true
	return nil
}
func (s *memStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.alive, id)
	return nil
}

func TestManagerReusesLiveProcess(t *testing.T) {
	proc := &fakeProcess{id: 1, alive: true}
	launcher := &fakeLauncher{next: func(int) Process { return proc }}
	m := NewManager(launcher, newMemStore(), time.Minute)
	spec := LaunchSpec{UserID: "u1", CLI: "claude-code"}

	for i := 0; i < 3; i++ {
		if _, err := m.Send(context.Background(), spec, "hi"); err != nil {
			t.Fatalf("Send: %v", err)
		}
	}
	if launcher.launches != 1 {
		t.Fatalf("expected 1 launch, got %d", launcher.launches)
	}
	if proc.calls != 3 {
		t.Fatalf("expected 3 sends to same process, got %d", proc.calls)
	}
}

func TestManagerRelaunchesWhenTTLExpired(t *testing.T) {
	store := newMemStore()
	procs := []*fakeProcess{{id: 1, alive: true}, {id: 2, alive: true}}
	launcher := &fakeLauncher{next: func(n int) Process { return procs[n-1] }}
	m := NewManager(launcher, store, time.Minute)
	spec := LaunchSpec{UserID: "u1"}

	if _, err := m.Send(context.Background(), spec, "a"); err != nil {
		t.Fatal(err)
	}
	// Simulate TTL expiry in the store.
	store.Delete(context.Background(), "u1")

	if _, err := m.Send(context.Background(), spec, "b"); err != nil {
		t.Fatal(err)
	}
	if launcher.launches != 2 {
		t.Fatalf("expected relaunch, got %d launches", launcher.launches)
	}
	if procs[0].killed == 0 {
		t.Fatalf("expected stale process to be killed")
	}
}

func TestManagerRelaunchesWhenProcessDead(t *testing.T) {
	store := newMemStore()
	procs := []*fakeProcess{{id: 1, alive: false}, {id: 2, alive: true}}
	launcher := &fakeLauncher{next: func(n int) Process { return procs[n-1] }}
	m := NewManager(launcher, store, time.Minute)
	spec := LaunchSpec{UserID: "u1"}

	// First call launches proc[0] (alive=false). Manager stores it; next call
	// sees it dead and relaunches.
	if _, err := m.Send(context.Background(), spec, "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Send(context.Background(), spec, "b"); err != nil {
		t.Fatal(err)
	}
	if launcher.launches != 2 {
		t.Fatalf("expected 2 launches, got %d", launcher.launches)
	}
}

func TestManagerDropsProcessOnSendError(t *testing.T) {
	store := newMemStore()
	procs := []*fakeProcess{{id: 1, alive: true, sendErr: errors.New("boom")}, {id: 2, alive: true}}
	launcher := &fakeLauncher{next: func(n int) Process { return procs[n-1] }}
	m := NewManager(launcher, store, time.Minute)
	spec := LaunchSpec{UserID: "u1"}

	if _, err := m.Send(context.Background(), spec, "a"); err == nil {
		t.Fatal("expected error from failing process")
	}
	if procs[0].killed == 0 {
		t.Fatal("expected failed process to be killed")
	}
	// Next call should relaunch a fresh process and succeed.
	if _, err := m.Send(context.Background(), spec, "b"); err != nil {
		t.Fatalf("expected recovery, got %v", err)
	}
	if launcher.launches != 2 {
		t.Fatalf("expected 2 launches, got %d", launcher.launches)
	}
}
