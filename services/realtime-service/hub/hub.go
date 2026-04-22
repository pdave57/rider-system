package hub

import (
	"encoding/json"
	"log"
	"sync"
)

type MessageType string

const (
	MsgGPSUpdate   MessageType = "gps_update"
	MsgOrderUpdate MessageType = "order_update"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type Client struct {
	ID   uint
	Role string
	Send chan []byte
}

func (c *Client) Emit(msg Message) {
	b, _ := json.Marshal(msg)
	select {
	case c.Send <- b:
	default:
	}
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uint]*Client
}

func NewHub() *Hub { return &Hub{clients: make(map[uint]*Client)} }

func (h *Hub) Register(c *Client) {
	c.Send = make(chan []byte, 128)
	h.mu.Lock()
	h.clients[c.ID] = c
	h.mu.Unlock()
	log.Printf("[hub] user %d connected role=%s", c.ID, c.Role)
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c.ID)
	h.mu.Unlock()
	close(c.Send)
	log.Printf("[hub] user %d disconnected", c.ID)
}

func (h *Hub) SendTo(userID uint, msg Message) {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if ok { c.Emit(msg) }
}

func (h *Hub) Broadcast(role string, msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		if c.Role == role { c.Emit(msg) }
	}
}

func (h *Hub) BroadcastAll(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients { c.Emit(msg) }
}
