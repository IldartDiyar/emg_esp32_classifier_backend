package httpH

import (
	"emg_esp32_classifier_backend/internal/svc"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type HTTPHandler struct {
	svc *svc.Service
}

func NewHTTPHandler(s *svc.Service) *HTTPHandler {
	return &HTTPHandler{svc: s}
}

func (h *HTTPHandler) GetDeviceList(w http.ResponseWriter, r *http.Request) {
	devices, err := h.svc.GetDeviceList(r.Context())
	if err != nil {
		http.Error(w, "failed to get device list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, devices)
}

func (h *HTTPHandler) GetMovements(w http.ResponseWriter, r *http.Request) {
	movs, err := h.svc.GetMovements(r.Context())
	if err != nil {
		http.Error(w, "failed to get movements: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, movs)
}

func (h *HTTPHandler) GetTrainingRawCSV(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetTrainingRawCSV(r.Context())
	if err != nil {
		http.Error(w, "failed to generate CSV: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="training_raw.csv"`)
	w.Write(data)
}

func (h *HTTPHandler) ReserveDevice(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/device/")
	id = strings.TrimSuffix(id, "/reserve")

	deviceID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid device ID", http.StatusBadRequest)
		return
	}

	// вызываем сервис
	if err := h.svc.ReserveDevice(r.Context(), deviceID); err != nil {
		http.Error(w, "failed to reserve device: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"status":    "reserved",
		"device_id": deviceID,
	})
}

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
