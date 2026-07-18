package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub broadcasts balance updates to all connected WebSocket clients.
type Hub struct {
	mu       sync.Mutex // protects conns and serialises Broadcast calls.
	conns    map[*websocket.Conn]struct{}
	upgrader websocket.Upgrader
}

// NewHub returns a Hub ready to accept WebSocket connections.
func NewHub() *Hub {
	return &Hub{
		conns: make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				return strings.HasPrefix(origin, "http://"+r.Host) || strings.HasPrefix(origin, "https://"+r.Host)
			},
		},
	}
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	h.conns[conn] = struct{}{}
	h.mu.Unlock()
}

// Unregister removes a connection from the broadcast list.
func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.conns, conn)
	h.mu.Unlock()
}

// Broadcast sends the new balance as JSON to every connected client.
// Connections that fail to receive the message are closed and removed.
// Safe for concurrent use.
func (h *Hub) Broadcast(balance int) {
	h.mu.Lock()

	payload, err := json.Marshal(struct {
		Balance int `json:"balance"`
	}{Balance: balance})
	if err != nil {
		h.mu.Unlock()
		log.Printf("ws marshal error: %v", err)
		return
	}

	var dead []*websocket.Conn
	for conn := range h.conns {
		if err := conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
			dead = append(dead, conn)
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("ws write error: %v", err)
			dead = append(dead, conn)
		}
	}

	for _, conn := range dead {
		delete(h.conns, conn)
	}
	h.mu.Unlock()

	for _, conn := range dead {
		_ = conn.Close()
	}
}

// ServeHTTP upgrades HTTP requests to WebSocket connections and keeps them
// alive until the client disconnects.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	h.Register(conn)
	defer h.unregisterAndClose(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

func (h *Hub) unregisterAndClose(conn *websocket.Conn) {
	h.Unregister(conn)
	_ = conn.Close()
}
