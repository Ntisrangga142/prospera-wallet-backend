package pkg

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Clients map[int]*websocket.Conn
	Mutex   sync.Mutex
}

var WebSocketHub = &Hub{
	Clients: make(map[int]*websocket.Conn),
}

// Register user ke hub
func (h *Hub) Register(userID int, conn *websocket.Conn) {
	h.Mutex.Lock()
	h.Clients[userID] = conn
	h.Mutex.Unlock()
}

// Kirim notifikasi ke user tertentu
func (h *Hub) SendToUser(userID int, message string) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	if conn, ok := h.Clients[userID]; ok {
		conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
}
