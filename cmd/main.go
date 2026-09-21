package main

import (
	"github.com/gin-gonic/gin"
	"lgsm-info-api/pkg/gameServers"
	"lgsm-info-api/pkg/gameServers/client"
	"log"
	"os"
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

const (
	windroseStatusPath = "/home/windrose/windrose/server-files/windrose_plus_data/server_status.json"
	windroseMaxAge     = 90 * time.Second

	// Palworld's REST API binds localhost only; it is never exposed publicly.
	palworldBaseURL = "http://127.0.0.1:8212"
	palworldTimeout = 3 * time.Second

	// The REST API fixes the basic-auth user as "admin"; only the password is
	// configurable, and it comes from the systemd EnvironmentFile, not the repo.
	palworldUser        = "admin"
	palworldPasswordEnv = "PALWORLD_ADMIN_PASSWORD"
)

// newPalworldClientFromEnv builds the client from the environment, warning
// once at startup if the password is missing rather than on every refresh.
func newPalworldClientFromEnv() client.PalworldClient {
	password := os.Getenv(palworldPasswordEnv)
	if password == "" {
		log.Printf("%s is not set: Palworld will always report offline", palworldPasswordEnv)
	}

	return client.NewPalworldClient(palworldBaseURL, palworldUser, password, palworldTimeout)
}

func main() {
	gameDigClient := client.NewGameDigClient()
	windroseClient := client.NewWindroseClient(windroseStatusPath, windroseMaxAge)
	palworldClient := newPalworldClientFromEnv()
	cache := gameServers.NewServerCache(gameDigClient, windroseClient, palworldClient, 30*time.Second)
	cache.Start()

	router := setupRouter(cache)
	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Error running gin server: %s", err)
	}
}
