package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrCode int

const (
	CodeSuccess        ErrCode = 0
	CodeBadRequest     ErrCode = 40000
	CodeUnauthorized   ErrCode = 40100
	CodeInternal       ErrCode = 50000
	CodeUpstreamFailed ErrCode = 50200
	CodeRateLimited    ErrCode = 42900
)

type BusinessError struct {
	Code    ErrCode
	Message string
	Err     error
}

func (e *BusinessError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *BusinessError) Unwrap() error { return e.Err }

func NewBizError(code ErrCode, msg string, err error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

func BadRequest(msg string) *BusinessError {
	return NewBizError(CodeBadRequest, msg, nil)
}

func Unauthorized(msg string) *BusinessError {
	return NewBizError(CodeUnauthorized, msg, nil)
}

func Internal(err error) *BusinessError {
	return NewBizError(CodeInternal, "服务暂时不可用", err)
}

type Response struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
	Data    any     `json:data,omitempty`
	TraceID string  `json:"trace_id,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func Fail(c *gin.Context, err error) {
	var biz *BusinessError
	if errors.As(err, &biz) {
		c.JSON(http.StatusOK, Response{
			Code:    biz.Code,
			Message: biz.Message,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code:    CodeInternal,
		Message: "服务暂时不可用",
	})
}
