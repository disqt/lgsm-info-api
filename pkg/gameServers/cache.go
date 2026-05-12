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
	interval       time.Duration
}

func NewServerCache(gameDigClient client.GameDigClient, windroseClient client.WindroseClient, interval time.Duration) *ServerCache {
	return &ServerCache{
		gameDigClient:  gameDigClient,
		windroseClient: windroseClient,
		interval:       interval,
	}
}

func (c *ServerCache) Get() model.OrderedServerMap {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.response
}

func (c *ServerCache) refresh() {
	servers, err := GetGameServers(c.gameDigClient)
	if err != nil {
		log.Printf("Cache refresh error: %s", err)
		return
	}

	servers = append(servers, GetWindroseServer(c.windroseClient))

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

func (c *ServerCache) Start() {
	c.refresh()
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for range ticker.C {
			c.refresh()
		}
	}()
}
