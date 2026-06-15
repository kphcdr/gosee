package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 约定：Code = 0 表示成功，非 0 表示业务错误
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 成功返回数据
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "success", Data: data})
}

// OKMsg 成功返回仅消息
func OKMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: msg})
}

// Fail 业务失败，HTTP 200
func Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Body{Code: -1, Message: msg})
}

// FailWithCode 指定业务码失败
func FailWithCode(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Body{Code: code, Message: msg})
}

// FailWithHTTP 带 HTTP 状态码的失败（参数校验、未授权等）
func FailWithHTTP(c *gin.Context, httpStatus int, msg string) {
	c.JSON(httpStatus, Body{Code: -1, Message: msg})
}
