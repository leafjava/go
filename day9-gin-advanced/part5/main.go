package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if appErr, ok := err.(*AppError); ok {
				c.JSON(appErr.Code, gin.H{
					"code":    appErr.Code,
					"message": appErr.Message,
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部错误",
				})
			}

		}
	}
}

func main() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(ErrorHandler())

	r.GET("/error", func(c *gin.Context) {
		c.Error(&AppError{
			Code:    400,
			Message: "参数验证失败",
		})
	})

	r.GET("/panic", func(c *gin.Context) {
		panic("something went wrong")
	})

	r.Run(":8081")
}

//2. 测试

//go run .
//接口一：自定义 AppError
//
//curl.exe http://localhost:8080/error
//返回（HTTP 状态码也是 400）：
//
//
//{"code": 400, "message": "参数验证失败"}
//接口二：panic 崩溃兜底
//
//curl.exe http://localhost:8080/panic
//返回：
//
//
//{"code": 500, "message": "服务器内部错误"}
