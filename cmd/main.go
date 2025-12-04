package main

import (
	"log"
	"net/http"
	"strings"

	"emg_esp32_classifier_backend/internal/ctrl/httpH"
	"emg_esp32_classifier_backend/internal/ctrl/ws"
	"emg_esp32_classifier_backend/internal/repo"
	"emg_esp32_classifier_backend/internal/svc"
)

func main() {
	db, err := repo.NewPostgresConnection()
	if err != nil {
		log.Fatal(err)
	}

	repository := repo.NewPostgresRepository(db)

	service := svc.NewService(repository)

	hub := ws.NewHub()

	frontendWS := ws.NewFrontendWSHandler(service, hub)
	espWS := ws.NewEspWSHandler(service, hub)

	httpHandler := httpH.NewHTTPHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/ws/frontend", frontendWS.HandleFrontendWS)
	mux.HandleFunc("/ws/esp", espWS.HandleEspWS)

	mux.HandleFunc("/devices", httpHandler.GetDeviceList)
	mux.HandleFunc("/movements", httpHandler.GetMovements)
	mux.HandleFunc("/training/raw/csv", httpHandler.GetTrainingRawCSV)

	mux.HandleFunc("/device/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/reserve") {
			httpHandler.ReserveDevice(w, r)
			return
		}

		http.NotFound(w, r)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
