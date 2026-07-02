package main

import (
	"log"
	"net/http"
	"part7/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// ============================================================
// WalletHandler — 钱包处理层
// ============================================================
type WalletHandler struct {
	DB *gorm.DB
}

// ============================================================
// 创建钱包
// ============================================================
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req struct {
		UserID  uint   `json:"user_id" binding:"required"`
		Address string `json:"address" binding:"required"`
		Network string `json:"network" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

// ============================================================
// 获取钱包列表（支持按 user_id 筛选）
// ============================================================
func (h *WalletHandler) GetWallets(c *gin.Context) {
	userID := c.Query("user_id")

	var wallets []models.Wallet
	query := h.DB.Preload("Transactions")

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&wallets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallets": wallets,
	})
}

// ============================================================
// 更新余额
// ============================================================
func (h *WalletHandler) UpdateBalance(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Balance float64 `json:"balance" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Model(&models.Wallet{}).Where("id = ?", id).Update("balance", req.Balance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "余额更新成功"})
}

// ============================================================
// 删除钱包
// ============================================================
func (h *WalletHandler) DeleteWallet(c *gin.Context) {
	id := c.Param("id")

	if err := h.DB.Delete(&models.Wallet{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "钱包删除成功"})
}

// ============================================================
// main — 初始化数据库 + 注册路由
// ============================================================
func main() {
	// ① 连接数据库（SQLite）
	db, err := gorm.Open(sqlite.Open("wallet.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	log.Println("数据库连接成功")

	// ② 自动建表
	db.AutoMigrate(&models.User{}, &models.Wallet{}, &models.Transaction{})
	log.Println("数据库迁移完成")

	// ③ 创建 handler
	handler := &WalletHandler{DB: db}

	// ④ 注册路由
	r := gin.Default()

	r.POST("/wallets", handler.CreateWallet)       // 创建钱包
	r.GET("/wallets", handler.GetWallets)          // 查钱包列表
	r.PUT("/wallets/:id", handler.UpdateBalance)   // 更新余额
	r.DELETE("/wallets/:id", handler.DeleteWallet) // 删除钱包

	// ⑤ 启动服务
	r.Run(":8082")
}
