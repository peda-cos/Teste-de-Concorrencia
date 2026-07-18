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
	mu       sync.Mutex // serialises Broadcast so concurrent calls don't race on WS writes.
	conns    sync.Map
	upgrader websocket.Upgrader
}

// NewHub returns a Hub ready to accept WebSocket connections.
func NewHub() *Hub {
	return &Hub{
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

// Register adds a connection to the broadcast list.
func (h *Hub) Register(conn *websocket.Conn) {
	h.conns.Store(conn, true)
}

// Unregister removes a connection from the broadcast list.
func (h *Hub) Unregister(conn *websocket.Conn) {
	h.conns.Delete(conn)
}

// Broadcast sends the new balance as JSON to every connected client.
// Connections that fail to receive the message are closed and removed.
// Safe for concurrent use.
func (h *Hub) Broadcast(balance int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	payload, err := json.Marshal(map[string]int{"balance": balance})
	if err != nil {
		log.Printf("ws marshal error: %v", err)
		return
	}

	var dead []*websocket.Conn
	h.conns.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)
		if err := conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
			dead = append(dead, conn)
			return true
		}
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("ws write error: %v", err)
			dead = append(dead, conn)
		}
		return true
	})

	for _, conn := range dead {
		h.unregisterAndClose(conn)
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
