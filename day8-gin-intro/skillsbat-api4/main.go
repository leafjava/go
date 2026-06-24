package main

import (
	"skillsbay-api4/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		wallets := api.Group("/wallets")
		{
			wallets.GET("/:address", handlers.GetWallet)
			wallets.GET("/:address/balance", handlers.GetBalance)
			wallets.POST("/:address/transfer", handlers.Transfer)
		}
	}
}
