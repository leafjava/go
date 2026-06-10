package main

import (
	"skillsbay-api3/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		wallets := api.Group("/wallets")
		{
			wallets.GET("/:address", handlers.Getwallet)
			wallets.GET("/:address/balance", handlers.GetBalance)
			wallets.POST("/:address/transfer", handlers.Transfer)
		}
	}

	r.Run(":8081")
}
