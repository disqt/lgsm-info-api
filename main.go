package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

type Server struct {
	Server  string
	Running bool
	Pid     string
}

const (
	Zomboid   string = "Zomboid"
	Minecraft        = "Minecraft"
	Valheim          = "Valheim"
)

func systemDirectory(w http.ResponseWriter, _ *http.Request) {
	serverNames := []string{Zomboid, Minecraft, Valheim}
	res := make(map[string]Server, len(serverNames))

	for _, serverName := range serverNames {
		regex := fmt.Sprintf("%s.*servername|servername.*%s", serverName, serverName)
		cmd := exec.Command("bash", "-c", fmt.Sprintf("pgrep -f '%s'", regex))
		fmt.Println(cmd.String())
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(out))
			res[serverName] = Server{
				Server:  serverName,
				Running: false,
				Pid:     "",
			}
		} else {
			res[serverName] = Server{
				Server:  serverName,
				Running: true,
				Pid:     string(out),
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/", systemDirectory)
	fmt.Println("Starting server on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
