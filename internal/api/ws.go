package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const wsPingInterval = 30 * time.Second
const wsPongDeadline = 10 * time.Second

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		h := r.Host
		return h == "localhost" || h == "127.0.0.1" ||
			strings.HasPrefix(h, "localhost:") || strings.HasPrefix(h, "127.0.0.1:")
	},
}

// Hub manages all connected WebSocket clients and broadcasts messages.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{clients: map[*websocket.Conn]bool{}}
}

func (h *Hub) Register(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

func (h *Hub) Unregister(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

// Broadcast sends msg to all connected clients. Dead connections are pruned.
func (h *Hub) Broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			c.Close()
			delete(h.clients, c)
		}
	}
}

// ServeWS upgrades an HTTP connection to WebSocket and registers it with the hub.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(wsPingInterval + wsPongDeadline))
	})
	h.Register(conn)
	go func() {
		defer h.Unregister(conn)
		ticker := time.NewTicker(wsPingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.mu.Lock()
				err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(wsPongDeadline))
				h.mu.Unlock()
				if err != nil {
					return
				}
			}
		}
	}()
	// Read loop: drives pong handler and detects disconnects.
	conn.SetReadDeadline(time.Now().Add(wsPingInterval + wsPongDeadline)) //nolint:errcheck
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
