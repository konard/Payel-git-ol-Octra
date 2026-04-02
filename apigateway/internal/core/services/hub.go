package services

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	TaskID string
	Conn   *websocket.Conn
	Hub    *Hub
	Send   chan interface{}
	Done   chan bool
}

// Hub maintains active client connections and broadcasts messages
type Hub struct {
	clients    map[string]map[*Client]bool // taskID -> clients
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Message represents a message to be sent to clients
type Message struct {
	TaskID string
	Data   interface{}
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.TaskID] == nil {
		h.clients[client.TaskID] = make(map[*Client]bool)
	}
	h.clients[client.TaskID][client] = true
	log.Printf("✅ Client registered for task %s", client.TaskID)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.TaskID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)
			if len(clients) == 0 {
				delete(h.clients, client.TaskID)
			}
			log.Printf("❌ Client unregistered from task %s", client.TaskID)
		}
	}
}

// broadcastMessage sends a message to all clients for a specific task
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	clients := h.clients[message.TaskID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.Send <- message.Data:
		default:
			// Client's send channel is full, close it
			h.unregister <- client
		}
	}
}

// Broadcast sends a message to all clients for a specific task
func (h *Hub) Broadcast(taskID string, data interface{}) {
	h.broadcast <- &Message{
		TaskID: taskID,
		Data:   data,
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
