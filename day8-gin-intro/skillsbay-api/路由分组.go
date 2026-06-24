package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/users", getUsers)
		v1.POST("/users", createUser)

		wallets := v1.Group("/wallets")
		{
			wallets.GET("/:address", getWallet)
			wallets.GET("/:address/balance", getBalance)
			wallets.POST("/:address/transfer", transfer)
		}
	}

	v2 := r.Group("/api/v2")
	{
		v2.GET("/users", getUsersV2)
	}

}

func getWallet(c *gin.Context) {
	address := c.Param("address")
	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"balance": 10.5,
	})
}

func getBalance(c *gin.Context) {
	address := c.Param("address")
	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"balance": 10.5,
	})
}

func getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": "v1",
	})
}

func createUser(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"version": "v1"})
}

func getUsersV2(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": "v2"})
}
