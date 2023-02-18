package main

import (
	"encoding/json"
	"lgsm-info-api/pkg/gameServers"
	"log"
	"net/http"
)

func getGameServers(w http.ResponseWriter, _ *http.Request) {
	servers := gameServers.GetRunningGameServers()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(servers)
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/servers", getGameServers)
	log.Default().Println("Starting server on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
