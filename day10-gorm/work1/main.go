package main

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"work1/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// ============================================================
// UserHandler — 用户管理处理层
// ============================================================
type UserHandler struct {
	DB *gorm.DB
}

// ============================================================
// 创建用户
// POST /users
// ============================================================
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Email    string `json:"email" binding:"required,email"`
		Age      int    `json:"age" binding:"gte=0,lte=150"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认角色
	if req.Role == "" {
		req.Role = "user"
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Age:      req.Age,
		Role:     req.Role,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"user":    user,
	})
}

// ============================================================
// 查询用户（支持分页 + 搜索）
// GET /users?page=1&page_size=10&search=张三
// ============================================================
func (h *UserHandler) GetUsers(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var users []models.User
	var total int64

	query := h.DB.Model(&models.User{})

	// 搜索：按用户名或邮箱模糊匹配
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 先查总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": int(math.Ceil(float64(total) / float64(pageSize))),
		},
	})
}

// ============================================================
// 更新用户信息
// PUT /users/:id
// ============================================================
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	// 检查用户是否存在
	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"omitempty,min=3,max=20"`
		Email    string `json:"email" binding:"omitempty,email"`
		Age      int    `json:"age" binding:"omitempty,gte=0,lte=150"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 只更新传了的字段
	updates := map[string]interface{}{}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Age > 0 {
		updates["age"] = req.Age
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}

	if err := h.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户更新成功",
		"user":    user,
	})
}

// ============================================================
// 删除用户（软删除）
// DELETE /users/:id
// ============================================================
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if err := h.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功（软删除）"})
}

// ============================================================
// 用户统计
// GET /users/stats
// ============================================================
func (h *UserHandler) GetUserStats(c *gin.Context) {
	var totalUsers int64
	var totalAdmins int64
	var avgAge float64

	h.DB.Model(&models.User{}).Count(&totalUsers)
	h.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&totalAdmins)
	h.DB.Model(&models.User{}).Select("AVG(age)").Scan(&avgAge)

	c.JSON(http.StatusOK, gin.H{
		"total_users":  totalUsers,
		"total_admins": totalAdmins,
		"avg_age":      avgAge,
	})
}

// ============================================================
// main — 初始化数据库 + 注册路由
// ============================================================
func main() {
	// ① 连接数据库（纯 Go SQLite，无需 CGO）
	db, err := gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	log.Println("数据库连接成功")

	// ② 自动建表
	db.AutoMigrate(&models.User{})
	log.Println("数据库迁移完成")

	// ③ 创建 handler
	handler := &UserHandler{DB: db}

	// ④ 注册路由
	r := gin.Default()

	r.POST("/users", handler.CreateUser)       // 创建用户
	r.GET("/users", handler.GetUsers)           // 查询用户（分页+搜索）
	r.GET("/users/stats", handler.GetUserStats) // 用户统计
	r.PUT("/users/:id", handler.UpdateUser)     // 更新用户
	r.DELETE("/users/:id", handler.DeleteUser)  // 删除用户（软删除）

	// ⑤ 启动服务
	r.Run(":8080")
}
