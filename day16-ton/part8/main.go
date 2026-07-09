package main

import (
	"fmt"
	"net/http"
	"part8/services"

	"github.com/gin-gonic/gin"
)

var tonService *services.TONService

func main() {
	var err error
	tonService, err = services.NewTONService("https://ton.org/global.config.json")
	if err != nil {
		panic("TON 服务初始化失败: " + err.Error())
	}
	defer tonService.Close()

	r := gin.Default()

	// TON API 路由组
	ton := r.Group("/api/v1/ton")
	{
		ton.GET("/balance/:address", getTONBalance)
		ton.GET("/account/:address", getTONAccount)
		ton.GET("/masterchain/info", getMasterchainInfo)
		ton.POST("/transfer", sendTONTransfer)
	}

	r.Run(":8083")
}

// GET /api/v1/ton/balance/:address
func getTONBalance(c *gin.Context) {
	address := c.Param("address")

	balance, err := tonService.GetBalance(address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"balance": balance,
		"unit":    "TON",
	})
}

func getTONAccount(c *gin.Context) {
	address := c.Param("address")

	balance, err := tonService.GetBalance(address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	status, err := tonService.GetAccountStatus(address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"balance": balance,
		"status":  status,
	})
}

func getMasterchainInfo(c *gin.Context) {
	info, err := tonService.GetMasterchainInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"seqno":     info.SeqNo,
		"root_hash": fmt.Sprintf("%x", info.RootHash),
		"file_hash": fmt.Sprintf("%x", info.FileHash),
	})
}

type TransferRequest struct {
	Mnemonic  string  `json:"mnemonic" binding:"required"`
	ToAddress string  `json:"to_address" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Comment   string  `json:"comment"`
}

func sendTONTransfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := tonService.SendTON(
		req.Mnemonic,
		req.ToAddress,
		fmt.Sprintf("%.9f", req.Amount),
		req.Comment,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": result,
	})

}
