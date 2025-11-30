package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/timestamp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int64{
			"timestamp": time.Now().Unix(),
		})
	})

	http.ListenAndServe(":8080", nil)
}
