package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gosee/internal/response"
	"gosee/internal/utils"
)

// 上下文键
const (
	ContextUserIDKey   = "user_id"
	ContextUsernameKey = "username"
)

// JWTAuth 校验 Authorization: Bearer <token>，把用户信息写入上下文
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			response.FailWithHTTP(c, http.StatusUnauthorized, "未提供认证信息")
			c.Abort()
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.FailWithHTTP(c, http.StatusUnauthorized, "认证格式错误，应为 Bearer <token>")
			c.Abort()
			return
		}
		claims, err := utils.ParseToken(parts[1], secret)
		if err != nil {
			response.FailWithHTTP(c, http.StatusUnauthorized, "认证已失效或无效")
			c.Abort()
			return
		}
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Next()
	}
}

// CurrentUserID 从上下文取当前用户 ID
func CurrentUserID(c *gin.Context) int64 {
	if v, ok := c.Get(ContextUserIDKey); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// CurrentUsername 从上下文取当前用户名
func CurrentUsername(c *gin.Context) string {
	if v, ok := c.Get(ContextUsernameKey); ok {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return ""
}
