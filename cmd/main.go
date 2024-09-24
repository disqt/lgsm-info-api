package main

import (
	"github.com/gin-gonic/gin"
	"lgsm-info-api/pkg/gameServers"
)

func getGameServers(c *gin.Context) {
	c.IndentedJSON(200, gameServers.GetGameServers())
}

func main() {
	router := gin.Default()
	router.GET("/servers", getGameServers)

	err := router.Run("localhost:8080")
	if err != nil {
		// Print error
		println(err)
	}
}
