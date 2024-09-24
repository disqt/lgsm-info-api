package main

import (
	"github.com/gin-gonic/gin"
	"lgsm-info-api/pkg/gameServers"
	"log"
)

func getGameServers(c *gin.Context) {
	servers, err := gameServers.GetGameServers()
	if err != nil {
		c.IndentedJSON(500, gin.H{"error": err.Error()})
	} else {
		c.IndentedJSON(200, servers)
	}
}

func main() {
	router := gin.Default()
	router.GET("/servers", getGameServers)

	err := router.Run("localhost:8080")

	if err != nil {
		// Print error
		log.Fatalf("Error running gin server: %s", err)
	}
}
