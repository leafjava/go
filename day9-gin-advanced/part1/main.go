package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 1. 中间件基础
func main() {
	r := gin.Default()

	r2 := gin.New()

	r2.Use(gin.Logger())
	r2.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello"})
	})
	
	r.Run(":8081")
}
