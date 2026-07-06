package middleware

import (
	"net/http"
	"part1/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "请求头中缺少 Authorization",
			})
			c.Abort()
			return
		}

		// 2. 验证 token 格式（Bearer token）
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Authorization 格式错误，应为 Bearer {token}",
			})
			c.Abort()
			return
		}

		// 3. 解析 token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "无效的 token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 4. 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "无权限访问",
			})
			c.Abort()
			return
		}

		// 检查角色
		roleStr := userRole.(string)
		hasPermission := false
		for _, role := range roles {
			if roleStr == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
