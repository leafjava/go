package main

import (
	"net/http"
	"your-project/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WalletHandler struct {
	DB *gorm.DB
}

// 创建钱包
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req struct {
		UserID  uint   `json:"user_id" binding:"required"`
		Address string `json:"address" binding:"required"`
		Network string `json:"network" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest), gin.H{"error": err.Error()}
		return
	}

	wallet := models.Wallet{
		UserID:  req.UserID,
		Address: req.Address,
		Network: req.Network,
		Balance: 0,
	}

	if err := h.DB.Create(&wallet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建钱包失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "钱包创建成功",
		"wallet":  wallet,
	})

}
