package ws

import (
	"context"
	"emg_esp32_classifier_backend/internal/svc"
	"emg_esp32_classifier_backend/pkg/models"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type FrontendWSHandler struct {
	svc *svc.Service
	hub *Hub
}

func NewFrontendWSHandler(s *svc.Service, hub *Hub) *FrontendWSHandler {
	return &FrontendWSHandler{svc: s, hub: hub}
}

var frontendUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (h *FrontendWSHandler) HandleFrontendWS(w http.ResponseWriter, r *http.Request) {
	conn, err := frontendUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS FRONTEND] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Println("[WS FRONTEND] connected")

	var deviceID int

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WS FRONTEND] read error: %v", err)
			return
		}

		var msg models.WsFrontendToBackend
		if err := json.Unmarshal(data, &msg); err != nil {
			h.writeError(conn, "invalid json: "+err.Error())
			continue
		}

		deviceID = msg.DeviceID

		h.hub.RegisterFrontend(deviceID, conn)

		h.hub.RegisterMasterFrontend(deviceID, conn)

		if h.hub.GetMasterFrontend(deviceID) != conn {
			log.Printf("[WS FRONTEND] client is not MASTER for deviceID=%d", deviceID)
			continue
		}

		ctx := context.Background()

		switch msg.Event {

		case models.EventStartTraining:
			resp, err := h.svc.WSStartTraining(ctx, msg)
			if err != nil {
				h.writeError(conn, err.Error())
				continue
			}

			b, _ := json.Marshal(resp)
			h.hub.SendToESP(deviceID, b)

		case models.EventStartStreaming:
			resp, err := h.svc.WSStartStreaming(ctx, msg)
			if err != nil {
				h.writeError(conn, err.Error())
				continue
			}

			b, _ := json.Marshal(resp)
			h.hub.SendToESP(deviceID, b)

		case models.EventStopTraining:
			resp, err := h.svc.WSStopStreaming(ctx, deviceID)
			if err != nil {
				h.writeError(conn, err.Error())
				continue
			}

			b, _ := json.Marshal(resp)
			h.hub.SendToESP(deviceID, b)
		}
	}
}

func (h *FrontendWSHandler) writeError(conn *websocket.Conn, msg string) {
	resp := map[string]any{"event": "error", "error": msg}
	b, _ := json.Marshal(resp)
	conn.WriteMessage(websocket.TextMessage, b)
}
