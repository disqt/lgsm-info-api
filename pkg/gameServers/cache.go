package gameServers

import (
	"lgsm-info-api/cmd/model"
	"lgsm-info-api/pkg/gameServers/client"
	"log"
	"sync"
	"time"
)

type ServerCache struct {
	mu             sync.RWMutex
	response       model.OrderedServerMap
	gameDigClient  client.GameDigClient
	windroseClient client.WindroseClient
	palworldClient client.PalworldClient
	interval       time.Duration
}

func NewServerCache(gameDigClient client.GameDigClient, windroseClient client.WindroseClient, palworldClient client.PalworldClient, interval time.Duration) *ServerCache {
	return &ServerCache{
		gameDigClient:  gameDigClient,
		windroseClient: windroseClient,
		palworldClient: palworldClient,
		interval:       interval,
	}
}

func (c *ServerCache) Get() model.OrderedServerMap {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.response
}

// Refresh performs one refresh synchronously.
func (c *ServerCache) Refresh() {
	servers, err := GetGameServers(c.gameDigClient)
	if err != nil {
		log.Printf("Cache refresh error: %s", err)
		return
	}

	servers = append(servers, GetWindroseServer(c.windroseClient))
	servers = append(servers, GetPalworldServer(c.palworldClient))

	response, err := model.NewResponse(servers)
	if err != nil {
		log.Printf("Cache response build error: %s", err)
		return
	}

	c.mu.Lock()
	c.response = response
	c.mu.Unlock()

	log.Println("Server cache refreshed")
}

// Start refreshes in the background, including the first pass: the handler
// already answers 503 until data lands, and a synchronous first refresh would
// gate the listener on a Palworld endpoint that can be bound but silent.
func (c *ServerCache) Start() {
	go func() {
		c.Refresh()
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for range ticker.C {
			c.Refresh()
		}
	}()
}
