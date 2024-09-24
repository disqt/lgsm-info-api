package main

import (
	"github.com/gin-gonic/gin"
	"lgsm-info-api/cmd/model"
	"lgsm-info-api/pkg/gameServers"
	"lgsm-info-api/pkg/gameServers/client"
	"log"
)

func GameServersHandler(gameDigClient client.GameDigClient) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		servers, err := gameServers.GetGameServers(gameDigClient)
		if err != nil {
			c.IndentedJSON(500, gin.H{"error": err.Error()})
		} else {
			response, err := model.NewResponse(servers)
			if err != nil {
				c.IndentedJSON(500, gin.H{"error": err.Error()})
				return
			}
			c.IndentedJSON(200, response)
		}
	}

	return fn
}

func setupRouter(gameDigClient client.GameDigClient) *gin.Engine {
	router := gin.Default()

	router.GET("/servers", GameServersHandler(gameDigClient))
	return router
}

func main() {
	router := setupRouter(client.NewGameDigClient())
	err := router.Run(":8080")
	if err != nil {
		// Print error
		log.Fatalf("Error running gin server: %s", err)
	}
}
