package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type Server struct {
	Server  string
	Running bool
}

const (
	Zomboid   string = "Zomboid"
	Minecraft        = "Minecraft"
	Valheim          = "Valheim"
)

func systemDirectory(w http.ResponseWriter, r *http.Request) {
	servers := []string{Zomboid, Minecraft, Valheim}
	processes := make(map[string]string)

	for _, server := range servers {
		process, err := exec.Command("zsh", "-c", "ps -ao pid,cmd | tr -s ' ' | grep \""+server+"\" | grep -v \"grep\"").Output()
		if err != nil {
			log.Fatal(err)
		}
		processes[server] = string(process)
	}

	response := make([]Server, 0)
	for server, process := range processes {
		if strings.Contains(process, "-servername") {
			response = append(response, Server{
				Server:  server,
				Running: true,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/", systemDirectory)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
