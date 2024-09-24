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
	"valheim":   "disqt.com",
}

// GetGameServers Run command, if error then add an OfflineServer to response
// If successful, add an OnlineServer
// Also append extras if present
func GetGameServers() ([]model.GameServer, error) {
	var servers []model.GameServer

	for game, host := range serverLookups {
		cmd := exec.Command("gamedig", "--type", game, host)

		output, err := cmd.Output()
		if err != nil {
			log.Fatalf("Error executing command: %s", err)
			return nil, err
		}

		fmt.Println("Raw JSON Output:", string(output))

		// Check if server is online or if gamedig returned an error
		var errorResponse struct {
			Error string `json:"error"`
		}

		if err := json.Unmarshal(output, &errorResponse); err == nil {
			servers = append(servers, model.OfflineGameServer{Name: game, IsOnline: false})
		} else {
			var response model.GameDigResponse
			err = json.Unmarshal(output, &response)

			if err != nil {
				log.Fatalf("Error unmarshalling JSON: %s", err)
				return nil, err
			}

			servers = append(servers, model.OnlineGameServer{Name: game, Host: host, Port: response.Port, Players: response.Players, MaxPlayers: response.MaxPlayers, IsOnline: true})
		}
	}

	return servers, nil
}
