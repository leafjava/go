package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

func main() {
	r := gin.Default()

	//r.GET("/users", func(c *gin.Context) {
	//	users := []string{"Alice", "Bob", "Charlie"}
	//	Success(c, gin.H{
	//		"users": users,
	//	})
	//})

	r.GET("/users", func(c *gin.Context) {
		users := []string{"Alice", "Bob", "Charlie"}

		Success(c, gin.H{"users": users})
	})

	r.GET("/error", func(c *gin.Context) {
		Error(c, http.StatusBadRequest, "参数错误")
	})

	r.Run(":8080")
}
