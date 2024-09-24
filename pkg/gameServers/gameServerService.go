package gameServers

import (
	"encoding/json"
	"fmt"
	"lgsm-info-api/pkg/gameServers/client"
	"lgsm-info-api/pkg/gameServers/model"
	"log"
)

var serverLookups = map[string]string{
	"minecraft": "disqt.com",
	"valheim":   "disqt.com",
}

// GetGameServers Run command, if error then add an OfflineServer to response
// If successful, add an OnlineServer
// Also append extras if present
func GetGameServers(gameDigClient client.GameDigClient) ([]model.GameServer, error) {
	var servers []model.GameServer

	for game, host := range serverLookups {
		output, err := gameDigClient.GetServerInfo(game, host)
		if err != nil {
			log.Fatalf("Error executing command: %s", err)
			return nil, err
		}

		fmt.Println("Raw JSON Output:", string(output))

		if isError(output) {
			// if err exists, then the server is offline
			servers = append(servers, model.NewOfflineGameServer(game))
		} else {
			// if err does not exist, then the server is online
			var response model.GameDigResponse
			err = json.Unmarshal(output, &response)

			if err != nil {
				log.Fatalf("Error unmarshalling JSON: %s", err)
				return nil, err
			}

			servers = append(servers, model.NewOnlineGameServer(game, host, string(response.Port), response.Players, response.MaxPlayers))
		}
	}

	return servers, nil
}

func isError(output []byte) bool {
	// Check if server is online or if gamedig returned an error
	var result map[string]interface{}
	err := json.Unmarshal(output, &result)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	if _, ok := result["error"]; ok {
		return true
	} else {
		return false
	}
}
