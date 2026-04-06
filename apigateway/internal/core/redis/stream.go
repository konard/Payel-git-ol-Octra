package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// StreamState represents the current state of a task stream
type StreamState struct {
	TaskID    string `json:"task_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`     // processing, done, error
	Progress  int32  `json:"progress"`   // 0-100
	Message   string `json:"message"`    // Human-readable message
	CreatedAt string `json:"created_at"` // ISO 8601 timestamp
	UpdatedAt string `json:"updated_at"` // ISO 8601 timestamp
}

// StreamUpdate represents a single update in the stream
type StreamUpdate struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	Progress  int32  `json:"progress"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

const (
	streamStatePrefix   = "STREAM:"
	streamUpdatesSuffix = ":UPDATES"
	wsConnectionsPrefix = "WS:"
	maxUpdatesHistory   = 50
)

// UpdateStreamState updates the current state of a task stream
func (c *Client) UpdateStreamState(ctx context.Context, taskID string, state StreamState) error {
	if !c.enabled {
		return nil
	}

	key := streamStatePrefix + taskID
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if state.CreatedAt == "" {
		state.CreatedAt = state.UpdatedAt
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal stream state: %w", err)
	}

	if err := c.rdb.HSet(ctx, key, "data", string(data)).Err(); err != nil {
		return fmt.Errorf("failed to set stream state: %w", err)
	}

	// Set TTL to 24 hours to prevent memory leaks
	c.rdb.Expire(ctx, key, 24*time.Hour)

	return nil
}

// GetStreamState retrieves the current state of a task stream
func (c *Client) GetStreamState(ctx context.Context, taskID string) (*StreamState, error) {
	if !c.enabled {
		return nil, nil // Not an error, just no data
	}

	key := streamStatePrefix + taskID
	data, err := c.rdb.HGet(ctx, key, "data").Result()
	if err == redis.Nil {
		return nil, nil // Not found
	} else if err != nil {
		return nil, fmt.Errorf("failed to get stream state: %w", err)
	}

	var state StreamState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stream state: %w", err)
	}

	return &state, nil
}

// AddStreamUpdate adds an update to the stream updates list
func (c *Client) AddStreamUpdate(ctx context.Context, taskID string, update StreamUpdate) error {
	if !c.enabled {
		return nil
	}

	key := streamStatePrefix + taskID + streamUpdatesSuffix
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal stream update: %w", err)
	}

	// Push to the end of the list
	if err := c.rdb.RPush(ctx, key, string(data)).Err(); err != nil {
		return fmt.Errorf("failed to add stream update: %w", err)
	}

	// Keep only last N updates
	if err := c.rdb.LTrim(ctx, key, int64(-maxUpdatesHistory), -1).Err(); err != nil {
		log.Printf("[Redis] Failed to trim updates list: %v", err)
	}

	// Set TTL to 24 hours
	c.rdb.Expire(ctx, key, 24*time.Hour)

	return nil
}

// GetStreamUpdates retrieves recent updates from the stream
func (c *Client) GetStreamUpdates(ctx context.Context, taskID string) ([]StreamUpdate, error) {
	if !c.enabled {
		return nil, nil
	}

	key := streamStatePrefix + taskID + streamUpdatesSuffix
	updates, err := c.rdb.LRange(ctx, key, 0, -1).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get stream updates: %w", err)
	}

	result := make([]StreamUpdate, 0, len(updates))
	for _, u := range updates {
		var update StreamUpdate
		if err := json.Unmarshal([]byte(u), &update); err != nil {
			log.Printf("[Redis] Failed to unmarshal update: %v", err)
			continue
		}
		result = append(result, update)
	}

	return result, nil
}

// RegisterWSConnection registers a WebSocket connection for a task
func (c *Client) RegisterWSConnection(ctx context.Context, taskID, connID string) error {
	if !c.enabled {
		return nil
	}

	key := wsConnectionsPrefix + taskID
	if err := c.rdb.SAdd(ctx, key, connID).Err(); err != nil {
		return fmt.Errorf("failed to register WS connection: %w", err)
	}

	// Set TTL to 1 hour
	c.rdb.Expire(ctx, key, time.Hour)

	return nil
}

// UnregisterWSConnection removes a WebSocket connection for a task
func (c *Client) UnregisterWSConnection(ctx context.Context, taskID, connID string) error {
	if !c.enabled {
		return nil
	}

	key := wsConnectionsPrefix + taskID
	if err := c.rdb.SRem(ctx, key, connID).Err(); err != nil {
		return fmt.Errorf("failed to unregister WS connection: %w", err)
	}

	return nil
}

// GetWSConnections returns all active WebSocket connections for a task
func (c *Client) GetWSConnections(ctx context.Context, taskID string) ([]string, error) {
	if !c.enabled {
		return nil, nil
	}

	key := wsConnectionsPrefix + taskID
	conns, err := c.rdb.SMembers(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get WS connections: %w", err)
	}

	return conns, nil
}
