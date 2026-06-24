package handlers

import (
	"net/http"
	"skillsbay-api4/models"

	"github.com/gin-gonic/gin"
)

var wallets = map[string]*models.Wallet{
	"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb": {
		Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		Balance: 10.5,
		Network: "Ethereum",
	},
}

func GetWallet(c *gin.Context) {
	address := c.Param("address")

	wallet, exists := wallets[address]

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "钱包不存在",
		})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

func GetBalance(c *gin.Context) {
	address := c.Param("address")

	wallet, exists := wallets[address]

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "钱包不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"balance": wallet.Balance,
	})
}

func Transfer(c *gin.Context) {
	from := c.Param("address")

	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fromWallet, exists := wallets[from]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "发送方钱包不存在",
		})
		return
	}

	if fromWallet.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "余额不足",
		})
	}

	fromWallet.Balance -= req.Amount

	c.JSON(http.StatusOK, gin.H{
		"message": "转账成功",
		"from":    from,
		"to":      req.To,
		"amount":  req.Amount,
	})
}
