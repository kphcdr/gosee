package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gosee/internal/response"
	"gosee/internal/utils"
)

// Recovery 捕获 panic，用 zap 记录后返回 500
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				if utils.Logger != nil {
					utils.Logger.Error("发生 panic",
						zap.Any("error", r),
						zap.String("stack", string(debug.Stack())),
						zap.String("path", c.Request.URL.Path),
					)
				}
				response.FailWithHTTP(c, http.StatusInternalServerError, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
