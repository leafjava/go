package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"required,gte=18"`
}

r.POST("/users",func(c *gin.Context){
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req);err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"error":err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated,gin.H{
		"message":"用户创建成功",
		"user":req,
	})
})