package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/users", getUsers)

	r.POST("/users", createUser)

	r.PUT("/user/:id", updateUser)

	r.DELETE("/user/:id", deleteUser)

	r.Run(":8080")
}

func getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"users": []string{"Alice", "Bob", "Charlie"},
	})
}

func createUser(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
	})
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "用户更新成功",
		"id":      id,
	})
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "用户删除成功",
		"id":      id,
	})
}
