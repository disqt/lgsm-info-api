package gameServers

import (
	"encoding/json"
	"fmt"
	"lgsm-info-api/pkg/gameServers/model"
	"log"
	"os/exec"
)

var serverLookups = map[string]string{
	"minecraft": "disqt.com",
}

// GetGameServers Run command, if error then add an OfflineServer to response
// If successful, add an OnlineServer
// Also append extras if present
func GetGameServers() []model.GameServer {
	var servers []model.GameServer

	for game, host := range serverLookups {
		cmd := exec.Command("gamedig", "--type", game, host)

		output, err := cmd.Output()
		if err != nil {
			log.Fatalf("Error executing command: %s", err)
		}

		fmt.Println("Raw JSON Output:", string(output))

		var response model.GameDigResponse
		err = json.Unmarshal(output, &response)

		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %s", err)
		}

		servers = append(servers, model.OnlineGameServer{Name: game, Host: host, Port: response.Port, Players: response.Players, MaxPlayers: response.MaxPlayers, IsOnline: true})
	}

	return servers
}
