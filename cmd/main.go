package main

import (
	"encoding/json"
	"lgsm-info-api/pkg/process"
	"log"
	"net/http"
)

func systemDirectory(w http.ResponseWriter, _ *http.Request) {
	servers := process.GetRunningServers()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(servers)
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/", systemDirectory)
	log.Default().Println("Starting server on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
