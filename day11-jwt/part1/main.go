package main

import (
	"net/http"
	"part1/middleware"
	"part1/models"
	"part1/utils"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// 初始化数据库
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移
	db.AutoMigrate(&models.User{})

	// 创建路由
	r := gin.Default()

	// 公开路由
	r.POST("/register", register)
	r.POST("/login", login)

	// 需要认证的路由
	auth := r.Group("/api")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/profile", getProfile)
		auth.PUT("/profile", updateProfile)

		// 需要管理员权限的路由
		admin := auth.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", listUsers)
			admin.DELETE("/users/:id", deleteUser)
		}
	}

	r.Run(":8080")
}

// 注册
func register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//创建用户
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     "user",
	}
	user.SetPassword(req.Password)

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "注册成功",
		"data": user,
	})

}

// 登录
func login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查询用户
	var user models.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成 token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token 生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token": token,
			"user":  user,
		},
	})
}

// 获取个人信息
func getProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": user,
	})
}

// 更新个人信息
func updateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Email string `json:"email" binding:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&models.User{}).Where("id = ?", userID).Update("email", req.
		Email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// 管理员：获取所有用户
func listUsers(c *gin.Context) {
	var users []models.User
	db.Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": users,
	})
}

// 管理员：删除用户
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	db.Delete(&models.User{}, id)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
