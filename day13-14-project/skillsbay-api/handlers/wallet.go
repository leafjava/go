package handlers

import (
	"net/http"
	"skillsbay-api/database"
	"skillsbay-api/models"
	"skillsbay-api/utils"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct{}

type CreateWalletRequest struct {
	Address string `json:"address" binding:"required"`
	Network string `json:"network" binding:"required"`
}

func (h *WalletHandler) CreateWallet(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 检查钱包是否已存在
	var existingWallet models.Wallet
	if err := database.DB.Where("address = ?", req.Address).First(&existingWallet).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "钱包已存在")
		return
	}

	wallet := models.Wallet{
		UserID:  userID,
		Address: req.Address,
		Network: req.Network,
		Balance: 0,
	}

	if err := database.DB.Create(&wallet).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "钱包创建失败")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"wallet": wallet,
	})

}

func (h *WalletHandler) GetWallets(c *gin.Context) {
	userID := c.GetUint("user_id")

	var wallets []models.Wallet
	if err := database.DB.Where("user_id = ?", userID).Preload("Transactions").Find(&wallets).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"wallets": wallets,
	})
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
	userID := c.GetUint("user_id")
	walletID := c.Param("id")

	var wallet models.Wallet
	if err := database.DB.Where("id = ? AND user_id = ?", walletID, userID).Preload("Transactions").First(&wallet).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "钱包不存在")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"wallet": wallet,
	})
}
