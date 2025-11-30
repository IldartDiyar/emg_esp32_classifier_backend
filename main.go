package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// =============================
// WebSocket HUB
// =============================

type Client struct {
	Conn *websocket.Conn
	ID   string
}

type Hub struct {
	clients map[string]*Client
	mu      sync.Mutex
}

func (h *Hub) getClients() (res []string) {
	for id, _ := range h.clients {
		res = append(res, id)
	}

	return
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS Upgrade error:", err)
		return
	}

	client := &Client{
		Conn: conn,
		ID:   id,
	}

	h.mu.Lock()
	h.clients[id] = client
	h.mu.Unlock()

	log.Println("[WS] Device connected:", id)

	go h.listen(client)
}

func (h *Hub) listen(c *Client) {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("[WS] Disconnected:", c.ID)
			h.mu.Lock()
			delete(h.clients, c.ID)
			h.mu.Unlock()
			return
		}

		log.Printf("[WS] %s -> %s\n", c.ID, string(msg))
	}
}

func (h *Hub) sendTo(deviceID string, data []byte) error {
	h.mu.Lock()
	c := h.clients[deviceID]
	h.mu.Unlock()

	if c == nil {
		return fmt.Errorf("device %s not connected", deviceID)
	}

	return c.Conn.WriteMessage(websocket.TextMessage, data)
}

// =============================
// COMMANDS
// =============================

func (h *Hub) SendStartTrain(id string) {
	msg := []byte(`{"cmd":"START_TRAIN"}`)
	_ = h.sendTo(id, msg)
	log.Println("[CMD] START_TRAIN ->", id)
}

func (h *Hub) SendSetTime(id string) {
	ts := time.Now().Unix()
	msg := []byte(fmt.Sprintf(`{"cmd":"SET_TIME","timestamp":%d}`, ts))
	_ = h.sendTo(id, msg)
	log.Println("[CMD] SET_TIME ->", id, ts)
}

func main() {
	hub := NewHub()

	// WebSocket endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.HandleWS(w, r)
	})

	http.HandleFunc("/cmd/start_train", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		hub.SendStartTrain(id)
		fmt.Fprintf(w, "OK\n")
	})

	http.HandleFunc("/cmd/set_time", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		hub.SendSetTime(id)
		fmt.Fprintf(w, "OK\n")
	})

	http.HandleFunc("/cmd/get_esps", func(w http.ResponseWriter, r *http.Request) {
		clients := hub.getClients()
		jj, _ := json.Marshal(clients)
		fmt.Fprintf(w, string(jj))
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
