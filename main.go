package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type Process struct {
	Pid int
	Cmd string
}

type Server struct {
	Server  string
	Running bool
	Cmd     string
}

const (
	Zomboid   string = "Zomboid"
	Minecraft        = "Minecraft"
	Valheim          = "Valheim"
)

func systemDirectory(w http.ResponseWriter, _ *http.Request) {
	servers := []string{Zomboid, Minecraft, Valheim}
	processes := make(map[string]Process, 0)

	for _, server := range servers {
		psCommand := "-c ps -ao pid,cmd | tr -s ' ' | grep \"" + server + "\" | grep -v \"grep\""
		cmd := exec.Command("bash", strings.Fields(psCommand)...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		process, err := cmd.Output()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return
		}

		// TODO split the string to get pid + cmd
		processes[server] = Process{
			000, string(process),
		}
	}

	response := make([]Server, 0)
	for server, process := range processes {
		if strings.Contains(process.Cmd, "-servername") {
			response = append(response, Server{
				Server:  server,
				Running: true,
				Cmd:     process.Cmd,
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
