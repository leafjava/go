package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

r.GET("/users/:id", func(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"id": id})
})

r.GET("/search",func(c *gin.Context){
	keyword := c.Query("keyword")
	page := c.DefaultQuery("page","1")

	c.JSON(http.StatusOK,gin.H{
		"keyword": keyword,
		"page": page,
	})
})