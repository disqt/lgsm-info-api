package gameServers

import (
	"encoding/json"
	"fmt"
	"lgsm-info-api/pkg/gameServers/client"
	"lgsm-info-api/pkg/gameServers/model"
	"log"
	"sync"
)

type ServerLookup struct {
	Id   string
	Host string
	Port string
}

var serverLookups = [...]ServerLookup{
	{Id: "minecraft", Host: "disqt.com", Port: ""},
	{Id: "valheim", Host: "disqt.com", Port: ""},
	{Id: "xonotic", Host: "disqt.com", Port: "26420"},
	{Id: "csgo", Host: "disqt.com", Port: "27015"},
}

// GetGameServers queries all game servers concurrently via gamedig.
// If a server query fails, it is reported as offline rather than crashing the API.
func GetGameServers(gameDigClient client.GameDigClient) ([]model.GameServer, error) {
	type result struct {
		server model.GameServer
		err    error
	}

	results := make([]result, len(serverLookups))
	var wg sync.WaitGroup

	for i, lookup := range serverLookups {
		wg.Add(1)
		go func(i int, lookup ServerLookup) {
			defer wg.Done()

			output, err := gameDigClient.GetServerInfo(lookup.Id, lookup.Host, lookup.Port)
			if err != nil {
				log.Printf("Error querying %s: %s", lookup.Id, err)
				results[i] = result{server: model.NewOfflineGameServer(lookup.Id)}
				return
			}

			fmt.Println("Raw JSON Output:", string(output))

			if isError(output) {
				results[i] = result{server: model.NewOfflineGameServer(lookup.Id)}
			} else {
				var response model.GameDigResponse
				err = json.Unmarshal(output, &response)
				if err != nil {
					log.Printf("Error unmarshalling JSON for %s: %s", lookup.Id, err)
					results[i] = result{server: model.NewOfflineGameServer(lookup.Id)}
					return
				}

				currentPlayer, err := response.Players.Int64()
				if err != nil {
					log.Println("Error getting current player")
					currentPlayer = 0
				}
				maxPlayers, err := response.MaxPlayers.Int64()
				if err != nil {
					log.Println("Error getting max players")
					maxPlayers = 0
				}
				results[i] = result{server: model.NewOnlineGameServer(lookup.Id, lookup.Host, string(response.Port), int(currentPlayer), int(maxPlayers))}
			}
		}(i, lookup)
	}

	wg.Wait()

	var servers []model.GameServer
	for _, r := range results {
		servers = append(servers, r.server)
	}

	return servers, nil
}

func isError(output []byte) bool {
	var result map[string]interface{}
	err := json.Unmarshal(output, &result)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return true
	}

	_, ok := result["error"]
	return ok
}
