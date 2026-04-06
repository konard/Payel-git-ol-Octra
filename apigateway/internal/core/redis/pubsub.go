package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// PubSubManager handles Redis Pub/Sub for real-time stream updates
type PubSubManager struct {
	client      *Client
	subscribers map[string]map[chan StreamUpdate]bool // taskID -> channels
	mu          sync.RWMutex
}

// NewPubSubManager creates a new PubSub manager
func NewPubSubManager(client *Client) *PubSubManager {
	return &PubSubManager{
		client:      client,
		subscribers: make(map[string]map[chan StreamUpdate]bool),
	}
}

// Subscribe subscribes to a task channel and returns a channel for updates
func (pm *PubSubManager) Subscribe(ctx context.Context, taskID string) (<-chan StreamUpdate, error) {
	if !pm.client.enabled {
		return nil, nil
	}

	ch := make(chan StreamUpdate, 64)
	channelName := "CHANNEL:task:" + taskID

	pm.mu.Lock()
	if _, exists := pm.subscribers[taskID]; !exists {
		pm.subscribers[taskID] = make(map[chan StreamUpdate]bool)
	}
	pm.subscribers[taskID][ch] = true
	pm.mu.Unlock()

	// Start subscriber if not already running
	pm.startSubscriber(ctx, channelName, taskID)

	return ch, nil
}

// Publish publishes an update to the task channel
func (pm *PubSubManager) Publish(ctx context.Context, taskID string, update StreamUpdate) error {
	if !pm.client.enabled {
		return nil
	}

	channelName := "CHANNEL:task:" + taskID
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	if err := pm.client.rdb.Publish(ctx, channelName, data).Err(); err != nil {
		return fmt.Errorf("failed to publish update: %w", err)
	}

	return nil
}

// Unsubscribe removes a subscriber for a task
func (pm *PubSubManager) Unsubscribe(taskID string, ch chan StreamUpdate) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if subscribers, exists := pm.subscribers[taskID]; exists {
		delete(subscribers, ch)
		close(ch)

		if len(subscribers) == 0 {
			delete(pm.subscribers, taskID)
		}
	}
}

// startSubscriber starts a goroutine to listen for updates on a channel
func (pm *PubSubManager) startSubscriber(ctx context.Context, channelName, taskID string) {
	// Check if subscriber already running
	pm.mu.RLock()
	// Simple check: if there are subscribers, assume subscriber is running
	if len(pm.subscribers[taskID]) > 1 {
		pm.mu.RUnlock()
		return
	}
	pm.mu.RUnlock()

	go func() {
		pubsub := pm.client.rdb.Subscribe(ctx, channelName)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var update StreamUpdate
				if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
					log.Printf("[Redis PubSub] Failed to unmarshal message: %v", err)
					continue
				}

				// Broadcast to all local subscribers
				pm.broadcast(taskID, update)
			}
		}
	}()
}

// broadcast sends an update to all local subscribers for a task
func (pm *PubSubManager) broadcast(taskID string, update StreamUpdate) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if subscribers, exists := pm.subscribers[taskID]; exists {
		for ch := range subscribers {
			select {
			case ch <- update:
			default:
				// Channel full, skip
				log.Printf("[Redis PubSub] Subscriber channel full for task %s, skipping", taskID)
			}
		}
	}
}
