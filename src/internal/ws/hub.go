package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a single WebSocket connection in a ticket room.
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	TicketID string
	UserID   string
	Role     string // "user" | "admin"
}

// Hub manages all active ticket rooms and their connected clients.
type Hub struct {
	mu      sync.RWMutex
	tickets map[string]map[*Client]bool
}

var Global = &Hub{
	tickets: make(map[string]map[*Client]bool),
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.tickets[c.TicketID] == nil {
		h.tickets[c.TicketID] = make(map[*Client]bool)
	}
	h.tickets[c.TicketID][c] = true
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.tickets[c.TicketID]
	if room == nil {
		return
	}
	delete(room, c)
	if len(room) == 0 {
		delete(h.tickets, c.TicketID)
	}
	close(c.send)
}

// Broadcast sends a raw JSON payload to every client in the ticket room.
func (h *Hub) Broadcast(ticketID string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.tickets[ticketID] {
		select {
		case c.send <- msg:
		default:
			// Slow client — drop message to avoid blocking.
		}
	}
}

// writePump pumps messages from the send channel to the WebSocket.
func (c *Client) WritePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
