package gameServers

import (
	"lgsm-info-api/cmd/model"
	"lgsm-info-api/pkg/gameServers/client"
	"log"
	"sync"
	"time"
)

type ServerCache struct {
	mu       sync.RWMutex
	response model.OrderedServerMap
	client   client.GameDigClient
	interval time.Duration
}

func NewServerCache(gameDigClient client.GameDigClient, interval time.Duration) *ServerCache {
	return &ServerCache{
		client:   gameDigClient,
		interval: interval,
	}
}

func (c *ServerCache) Get() model.OrderedServerMap {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.response
}

func (c *ServerCache) refresh() {
	servers, err := GetGameServers(c.client)
	if err != nil {
		log.Printf("Cache refresh error: %s", err)
		return
	}

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
