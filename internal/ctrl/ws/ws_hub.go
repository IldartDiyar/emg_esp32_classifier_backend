package ws

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Hub struct {
	esp            map[int]*websocket.Conn   // deviceID → ESP conn
	frontend       map[int][]*websocket.Conn // deviceID → all clients
	masterFrontend map[int]*websocket.Conn   // deviceID → MASTER client
	mu             sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		esp:            make(map[int]*websocket.Conn),
		frontend:       make(map[int][]*websocket.Conn),
		masterFrontend: make(map[int]*websocket.Conn),
	}
}

func (h *Hub) RegisterESP(deviceID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.esp[deviceID] = conn
}

func (h *Hub) RegisterFrontend(deviceID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.frontend[deviceID] = append(h.frontend[deviceID], conn)
}

func (h *Hub) RegisterMasterFrontend(deviceID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.masterFrontend[deviceID]; !exists {
		h.masterFrontend[deviceID] = conn
	}
}

func (h *Hub) GetMasterFrontend(deviceID int) *websocket.Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.masterFrontend[deviceID]
}

func (h *Hub) SendToESP(deviceID int, data []byte) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conn := h.esp[deviceID]
	if conn == nil {
		return nil
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

func (h *Hub) SendToFrontend(deviceID int, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns := h.frontend[deviceID]
	for _, c := range conns {
		c.WriteMessage(websocket.TextMessage, data)
	}
}

func (h *Hub) RemoveESP(deviceID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.esp, deviceID)
}
