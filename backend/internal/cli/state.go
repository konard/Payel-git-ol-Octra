package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// State is the per-user CLI process state mirrored into Redis. The in-process
// registry owns the actual subprocess handle; Redis is the source of truth for
// liveness and TTL so the process can be reaped once it goes stale.
type State struct {
	PID       int       `json:"pid"`
	Port      int       `json:"port,omitempty"`
	StartedAt time.Time `json:"started_at"`
}

// StateStore persists CLI liveness with a TTL.
type StateStore interface {
	// Alive reports whether a live state record exists for the user.
	Alive(ctx context.Context, userID string) (bool, error)
	// Save writes the state with the given TTL (refreshing it on every call).
	Save(ctx context.Context, userID string, st State, ttl time.Duration) error
	// Delete removes the state record.
	Delete(ctx context.Context, userID string) error
}

func stateKey(userID string) string { return fmt.Sprintf("user:%s:cli_state", userID) }

// RedisStateStore is a Redis-backed StateStore. The Redis key TTL doubles as
// the `user:{id}:cli_ttl` described in the design.
type RedisStateStore struct {
	rdb *redis.Client
}

// NewRedisStateStore wraps a redis client.
func NewRedisStateStore(rdb *redis.Client) *RedisStateStore {
	return &RedisStateStore{rdb: rdb}
}

// Alive implements StateStore.
func (s *RedisStateStore) Alive(ctx context.Context, userID string) (bool, error) {
	n, err := s.rdb.Exists(ctx, stateKey(userID)).Result()
	return n > 0, err
}

// Save implements StateStore.
func (s *RedisStateStore) Save(ctx context.Context, userID string, st State, ttl time.Duration) error {
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, stateKey(userID), data, ttl).Err()
}

// Delete implements StateStore.
func (s *RedisStateStore) Delete(ctx context.Context, userID string) error {
	return s.rdb.Del(ctx, stateKey(userID)).Err()
}
