package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"emg_esp32_classifier_backend/internal/svc"
	"emg_esp32_classifier_backend/pkg/models"

	"github.com/gorilla/websocket"
)

type EspWSHandler struct {
	svc *svc.Service
	hub *Hub
}

func NewEspWSHandler(s *svc.Service, hub *Hub) *EspWSHandler {
	return &EspWSHandler{svc: s, hub: hub}
}

var espUpgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (h *EspWSHandler) HandleEspWS(w http.ResponseWriter, r *http.Request) {
	conn, err := espUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS ESP] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Println("[WS ESP] connected")

	var deviceID int

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WS ESP] read error: %v", err)

			if deviceID != 0 {
				h.hub.RemoveESP(deviceID)
			}

			conn.Close() // VERY IMPORTANT
			return
		}

		var msg models.WsEspToBackend
		if err := json.Unmarshal(data, &msg); err != nil {
			h.writeError(conn, "invalid json: "+err.Error())
			continue
		}

		ctx := context.Background()

		if msg.Event == models.HandShake {
			deviceID, err = h.svc.RegisterDevice(ctx, msg.DeviceName)
			if err != nil {
				h.writeError(conn, "device registration failed: "+err.Error())
				continue
			}

			h.hub.RegisterESP(deviceID, conn)

			log.Printf("[WS ESP] registered device %s → ID=%d", msg.DeviceName, deviceID)

			resp := map[string]any{"event": "handshake_ok", "device_id": deviceID}
			b, _ := json.Marshal(resp)
			conn.WriteMessage(websocket.TextMessage, b)
			continue
		}

		if deviceID == 0 {
			continue
		}

		resp, err := h.svc.WSRawStream(ctx, msg, deviceID)
		if err != nil {
			h.writeError(conn, err.Error())
			continue
		}

		b, _ := json.Marshal(resp)
		h.hub.SendToFrontend(deviceID, b)
	}
}

func (h *EspWSHandler) writeError(conn *websocket.Conn, msg string) {
	resp := map[string]any{"event": "error", "error": msg}
	b, _ := json.Marshal(resp)
	conn.WriteMessage(websocket.TextMessage, b)
}
