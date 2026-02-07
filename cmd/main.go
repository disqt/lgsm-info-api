package main

import (
	"github.com/gin-gonic/gin"
	"lgsm-info-api/pkg/gameServers"
	"lgsm-info-api/pkg/gameServers/client"
	"log"
	"time"
)

func GameServersHandler(cache *gameServers.ServerCache) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		response := cache.Get()
		if response == nil {
			c.IndentedJSON(503, gin.H{"error": "server data not yet available"})
			return
		}
		c.IndentedJSON(200, response)
	}
	return fn
}

func setupRouter(cache *gameServers.ServerCache) *gin.Engine {
	router := gin.Default()
	router.GET("/servers", GameServersHandler(cache))
	return router
}

func main() {
	gameDigClient := client.NewGameDigClient()
	cache := gameServers.NewServerCache(gameDigClient, 30*time.Second)
	cache.Start()

	router := setupRouter(cache)
	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Error running gin server: %s", err)
	}
}
