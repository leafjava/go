package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func validateEthAddress(f1 validator.FieldLevel) bool {
	address := f1.Field().String()
	return len(address) == 42 && strings.HasPrefix(address, "0x")
}

type TransferRequest struct {
	From   string  `json:"from" binding:"required,eth_address"`
	To     string  `json:"to" binding:"required,eth_address"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func main() {
	r := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("eth_address", validateEthAddress)
	}

	r.POST("/transfer", func(c *gin.Context) {
		var req TransferRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "转账成功",
			"data":    req,
		})
	})

	r.Run(":8081")
}
