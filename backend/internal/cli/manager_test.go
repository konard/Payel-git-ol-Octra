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
	id     int
	alive  bool
	killed int
}

func (p *fakeProcess) Alive() bool { return p.alive }
func (p *fakeProcess) Kill() error { p.killed++; p.alive = false; return nil }
func (p *fakeProcess) PID() int    { return p.id }
func (p *fakeProcess) Port() int   { return 0 }

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
func (s *memStore) Get(_ context.Context, id string) (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.alive[id]
	if !ok {
		return State{}, errors.New("not found")
	}
	return State{PID: 1, StartedAt: time.Now()}, nil
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
	spec := LaunchSpec{UserID: "u1"}

	for i := 0; i < 3; i++ {
		if _, err := m.EnsureOcawe(context.Background(), spec); err != nil {
			t.Fatalf("EnsureOcawe: %v", err)
		}
	}
	if launcher.launches != 1 {
		t.Fatalf("expected 1 launch, got %d", launcher.launches)
	}
}

func TestManagerRelaunchesWhenTTLExpired(t *testing.T) {
	store := newMemStore()
	procs := []*fakeProcess{{id: 1, alive: true}, {id: 2, alive: true}}
	launcher := &fakeLauncher{next: func(n int) Process { return procs[n-1] }}
	m := NewManager(launcher, store, time.Minute)
	spec := LaunchSpec{UserID: "u1"}

	if _, err := m.EnsureOcawe(context.Background(), spec); err != nil {
		t.Fatal(err)
	}
	store.Delete(context.Background(), "u1")

	if _, err := m.EnsureOcawe(context.Background(), spec); err != nil {
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

	if _, err := m.EnsureOcawe(context.Background(), spec); err != nil {
		t.Fatal(err)
	}
	if _, err := m.EnsureOcawe(context.Background(), spec); err != nil {
		t.Fatal(err)
	}
	if launcher.launches != 2 {
		t.Fatalf("expected 2 launches, got %d", launcher.launches)
	}
}
